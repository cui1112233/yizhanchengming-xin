//! 全局应用状态 `AppState`:配置、平台适配、豆包连接、持久化、任务调度器、日志缓冲、CDP 调试日志。
//!
//! `AppState` 由 `tauri::Builder::manage` 托管,命令通过 `State<'_, Arc<AppState>>` 访问,
//! HTTP 层(`crate::http`)通过 axum `State<Arc<AppState>>` 访问。
//! `config` 是 `Arc<RwLock<Config>>`,与 `Scheduler` 共享同一份(配置可热改)。

use std::collections::VecDeque;
use std::fs::{File, OpenOptions};
use std::io::{BufWriter, Write as _};
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::time::Duration;

use anyhow::Context;
use tokio::sync::{broadcast, RwLock};
use tokio::task::JoinHandle;
use tokio_util::sync::CancellationToken;

use crate::cdp::CdpLogEntry;
use crate::config::Config;
use crate::doubao::{ConnTiming, DoubaoConnection, SwitchResult};
use crate::error::RelayResult;
use crate::logging::{self, LogBuffer};
use crate::platform::Platform;
use crate::storage::Storage;
use crate::tasks::{DoubaoBackend, GenerationBackend, Scheduler};

/// 心跳间隔(spec:10s)。
pub const HEARTBEAT_INTERVAL: Duration = Duration::from_secs(10);
/// CDP 调试日志内存缓冲容量。
pub const CDP_LOG_CAPACITY: usize = 2000;
/// CDP 调试日志文件名(spec 09:`<log_dir>/cdp-debug.log`,ndjson)。
pub const CDP_DEBUG_LOG_FILE: &str = "cdp-debug.log";
/// `cdp-debug.log` 单文件上限(与 `logging.file_max_size_mb` 默认值 10MB 一致);超过后轮转为 `.1`(只留一个旧文件)。
pub const CDP_DEBUG_LOG_MAX_BYTES: u64 = 10 * 1024 * 1024;
const CDP_BROADCAST_CAPACITY: usize = 1024;

pub struct AppState {
    pub config: Arc<RwLock<Config>>,
    pub config_path: PathBuf,
    pub platform: Arc<dyn Platform>,
    pub doubao: Arc<DoubaoConnection>,
    pub storage: Storage,
    pub scheduler: Arc<Scheduler>,
    pub logs: LogBuffer,
    pub cdp_log: Arc<CdpLogBuffer>,
    /// 应用退出时取消(HTTP server 优雅关闭等)。
    pub shutdown: CancellationToken,
    /// HTTP API 是否已在监听(`http::serve` 维护,状态栏 "● API" 用)。
    pub api_listening: AtomicBool,
    /// 最近一次 HTTP API 启动错误；成功启动或关闭功能时清空。
    pub api_error: RwLock<Option<String>>,
    /// Unix epoch 毫秒。
    pub started_at: i64,
    /// 关闭流程协调(Task 5.2,`crate::shutdown`)。
    pub quit: crate::shutdown::QuitGate,
    /// 隐私声明是否已同意(Task 5.2,`crate::privacy`):启动时从 `metadata.privacy_accepted` 载入,
    /// 同意前不启动 HTTP API、不自动连豆包。
    pub privacy: tokio::sync::watch::Sender<bool>,
}

impl AppState {
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        config: Config,
        config_path: PathBuf,
        platform: Arc<dyn Platform>,
        doubao: Arc<DoubaoConnection>,
        logs: LogBuffer,
        cdp_log: Arc<CdpLogBuffer>,
        storage: Storage,
        backend: Arc<dyn GenerationBackend>,
    ) -> Arc<Self> {
        let config = Arc::new(RwLock::new(config));
        let scheduler = Scheduler::new(storage.clone(), backend, Arc::clone(&config));
        Arc::new(Self {
            config,
            config_path,
            platform,
            doubao,
            storage,
            scheduler,
            logs,
            cdp_log,
            shutdown: CancellationToken::new(),
            api_listening: AtomicBool::new(false),
            api_error: RwLock::new(None),
            started_at: logging::now_ms(),
            quit: crate::shutdown::QuitGate::default(),
            privacy: tokio::sync::watch::Sender::new(false),
        })
    }

    /// `run()` 的启动序列:读配置 → 初始化日志 → CDP 调试日志通道 → `DoubaoConnection`
    /// → `Storage`(在 Tauri 的 tokio 运行时上打开,之后命令 / HTTP / 调度器都在同一运行时)
    /// → `Scheduler`(`DoubaoBackend`)→ `AppState`。不能在 async 上下文里调用。
    pub fn bootstrap(platform: Arc<dyn Platform>) -> anyhow::Result<Arc<Self>> {
        let config_path = platform.config_path();
        let config = Config::load_or_create(&config_path)?;
        if let Err(reason) = config.validate() {
            anyhow::bail!("配置文件不合法({}):{reason}", config_path.display());
        }
        let logs = logging::init(&config.logging.level, &platform.log_dir())?;
        tracing::info!(
            platform = platform.platform_name(),
            config = %config_path.display(),
            log_dir = %platform.log_dir().display(),
            cdp_port = config.doubao.cdp_port,
            cdp_debug_log = config.logging.cdp_debug_log,
            "doubao-relay 启动"
        );
        let cdp_log = CdpLogBuffer::new(config.logging.cdp_debug_log);
        let doubao = DoubaoConnection::new_with(
            Arc::clone(&platform),
            config.doubao.cdp_address.clone(),
            config.doubao.cdp_port,
            cdp_log.sender(),
            ConnTiming::default(),
        );
        let data_dir = platform.app_data_dir();
        let storage = tauri::async_runtime::block_on(Storage::open(&data_dir))
            .with_context(|| format!("打开数据目录失败:{}", data_dir.display()))?;
        tracing::info!(data_dir = %data_dir.display(), "数据目录已就绪");
        let backend: Arc<dyn GenerationBackend> = DoubaoBackend::new(Arc::clone(&doubao));
        Ok(Self::new(
            config,
            config_path,
            platform,
            doubao,
            logs,
            cdp_log,
            storage,
            backend,
        ))
    }

    pub async fn config(&self) -> Config {
        self.config.read().await.clone()
    }

    /// 带运行中任务守卫的账号切换:HTTP `POST /accounts/switch` 与 Tauri `account_switch` 都走这里,
    /// 有任何非终态任务在运行 → `account_busy`(409),否则在 `Scheduler::begin_switch` 的守卫下
    /// `DoubaoConnection::switch_account`——切换期间(含页面 reload 与验证)调度器不派发新任务,
    /// 切换中提交的任务排队,切换结束(成功或失败)后才进入 preparing(spec 08 "账号互斥",UC-31)。
    pub async fn switch_account(&self, target: &str) -> RelayResult<SwitchResult> {
        let _pause = self.scheduler.begin_switch(target)?;
        self.doubao.switch_account(target).await
    }

    /// 提交任务:HTTP `POST /generate` 与 Tauri `task_submit` 共用;关闭流程开始后一律
    /// `system_shutdown`(spec 03 "关闭流程"第 2 步:停止接受新请求)。
    pub async fn submit_task(
        &self,
        req: crate::tasks::GenerateRequest,
    ) -> RelayResult<(String, u32, i64)> {
        if self.shutdown.is_cancelled() {
            return Err(crate::shutdown::shutdown_error());
        }
        self.scheduler.submit(req).await
    }
}

/// CDP 调试日志。
///
/// 两级广播:`tx` 是 `CdpSession` 往里发的通道,**只在开启时有订阅者**(收集任务);
/// session 侧按 `receiver_count()==0` 跳过构造日志,所以关闭状态下零开销。
/// 收集任务把条目写进环形缓冲(供 `debug.cdp_log`)、追加到 `<log_dir>/cdp-debug.log`,
/// 再转发到 `ui_tx`(`cdp_log_entry` 事件的转发任务常驻订阅它)。
/// 开关可在运行时切换([`set_enabled`](Self::set_enabled),设置页"开启 CDP 日志"热生效)。
pub struct CdpLogBuffer {
    enabled: AtomicBool,
    entries: Mutex<VecDeque<CdpLogEntry>>,
    tx: broadcast::Sender<CdpLogEntry>,
    ui_tx: broadcast::Sender<CdpLogEntry>,
    collector: Mutex<Option<JoinHandle<()>>>,
}

impl CdpLogBuffer {
    /// `enabled` 只是初始开关;收集任务要在 tokio 运行时内由
    /// [`ensure_collector`](Self::ensure_collector) / [`set_enabled`](Self::set_enabled) 启动。
    pub fn new(enabled: bool) -> Arc<Self> {
        let (tx, _) = broadcast::channel(CDP_BROADCAST_CAPACITY);
        let (ui_tx, _) = broadcast::channel(CDP_BROADCAST_CAPACITY);
        Arc::new(Self {
            enabled: AtomicBool::new(enabled),
            entries: Mutex::new(VecDeque::with_capacity(if enabled {
                CDP_LOG_CAPACITY
            } else {
                0
            })),
            tx,
            ui_tx,
            collector: Mutex::new(None),
        })
    }

    pub fn enabled(&self) -> bool {
        self.enabled.load(Ordering::SeqCst)
    }

    /// 交给 `DoubaoConnection::new_with` 的发送端。总是 `Some`:运行时打开开关后 session 才能开始记录;
    /// 关闭状态下没有订阅者,session 侧按 `receiver_count()` 跳过。
    pub fn sender(&self) -> Option<broadcast::Sender<CdpLogEntry>> {
        Some(self.tx.clone())
    }

    /// UI 侧订阅(`cdp_log_entry` 事件转发任务用),只收到收集任务放行的条目。
    pub fn subscribe(&self) -> broadcast::Receiver<CdpLogEntry> {
        self.ui_tx.subscribe()
    }

    /// 写入缓冲(满了丢最旧)。
    pub fn push(&self, entry: CdpLogEntry) {
        let mut q = self.entries.lock().unwrap_or_else(|e| e.into_inner());
        if q.len() >= CDP_LOG_CAPACITY {
            q.pop_front();
        }
        q.push_back(entry);
    }

    /// 最近 `limit` 条(按时间先后)。
    pub fn recent(&self, limit: usize) -> Vec<CdpLogEntry> {
        let q = self.entries.lock().unwrap_or_else(|e| e.into_inner());
        let skip = q.len().saturating_sub(limit);
        q.iter().skip(skip).cloned().collect()
    }

    pub fn len(&self) -> usize {
        self.entries.lock().unwrap_or_else(|e| e.into_inner()).len()
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }

    /// 运行时切换开关(设置页热生效)。开 → 启动收集任务(已在跑则不动);关 → 停掉收集任务,
    /// session 侧随即因没有订阅者而停止记录。缓冲里已有的条目保留。需在 tokio 运行时内调用。
    pub fn set_enabled(self: &Arc<Self>, on: bool, log_dir: Option<&Path>) {
        self.enabled.store(on, Ordering::SeqCst);
        if on {
            self.ensure_collector(log_dir);
        } else {
            let handle = self
                .collector
                .lock()
                .unwrap_or_else(|e| e.into_inner())
                .take();
            if let Some(h) = handle {
                h.abort();
            }
            tracing::info!("CDP 调试日志已关闭");
        }
    }

    /// 开关为开且收集任务未在跑时启动它(幂等)。`log_dir` 为 `Some` 时同时追加写 `cdp-debug.log`。
    /// 需在 tokio 运行时内调用。
    pub fn ensure_collector(self: &Arc<Self>, log_dir: Option<&Path>) {
        if !self.enabled() {
            return;
        }
        let mut slot = self.collector.lock().unwrap_or_else(|e| e.into_inner());
        if slot.as_ref().is_some_and(|h| !h.is_finished()) {
            return;
        }
        let file = log_dir.map(|d| d.join(CDP_DEBUG_LOG_FILE));
        tracing::info!(
            file = file.as_deref().map(|p| p.display().to_string()),
            "CDP 调试日志已开启"
        );
        *slot = Some(self.spawn_collector(file));
    }

    fn spawn_collector(self: &Arc<Self>, file: Option<PathBuf>) -> JoinHandle<()> {
        let this = Arc::clone(self);
        let mut rx = this.tx.subscribe();
        tokio::spawn(async move {
            let mut writer = file
                .as_deref()
                .and_then(|p| CdpFileWriter::open(p, CDP_DEBUG_LOG_MAX_BYTES));
            loop {
                match rx.recv().await {
                    Ok(entry) => {
                        if let Some(w) = writer.as_mut() {
                            if let Err(e) = w.write_entry(&entry) {
                                tracing::warn!(error = %e, "写 cdp-debug.log 失败,停止写文件");
                                writer = None;
                            }
                        }
                        this.push(entry.clone());
                        let _ = this.ui_tx.send(entry);
                    }
                    Err(broadcast::error::RecvError::Lagged(n)) => {
                        tracing::debug!(dropped = n, "CDP 调试日志收集落后,丢弃若干条");
                    }
                    Err(broadcast::error::RecvError::Closed) => break,
                }
            }
        })
    }
}

/// `cdp-debug.log` 写入器:unix 上文件权限 0600(内容可能含豆包 cookie / 请求头 / 完整 prompt,
/// 与 config.toml 同口径);超过 `max_bytes` 时把当前文件改名为 `cdp-debug.log.1`(覆盖旧的)再新开。
struct CdpFileWriter {
    path: PathBuf,
    w: BufWriter<File>,
    written: u64,
    max_bytes: u64,
}

impl CdpFileWriter {
    fn open(path: &Path, max_bytes: u64) -> Option<Self> {
        if let Some(parent) = path.parent() {
            let _ = std::fs::create_dir_all(parent);
        }
        match open_private_append(path) {
            Ok(f) => {
                let written = f.metadata().map(|m| m.len()).unwrap_or(0);
                Some(Self {
                    path: path.to_path_buf(),
                    w: BufWriter::new(f),
                    written,
                    max_bytes,
                })
            }
            Err(e) => {
                tracing::warn!(path = %path.display(), error = %e, "打开 cdp-debug.log 失败,只保留内存缓冲");
                None
            }
        }
    }

    fn write_entry(&mut self, entry: &CdpLogEntry) -> std::io::Result<()> {
        let mut line = serde_json::to_string(entry).map_err(std::io::Error::other)?;
        line.push('\n');
        let len = line.len() as u64;
        if self.written > 0 && self.written + len > self.max_bytes {
            self.rotate()?;
        }
        self.w.write_all(line.as_bytes())?;
        self.w.flush()?;
        self.written += len;
        Ok(())
    }

    fn rotate(&mut self) -> std::io::Result<()> {
        self.w.flush()?;
        let mut old = self.path.clone().into_os_string();
        old.push(".1");
        std::fs::rename(&self.path, PathBuf::from(old))?;
        self.w = BufWriter::new(open_private_append(&self.path)?);
        self.written = 0;
        Ok(())
    }
}

/// 以追加方式打开;unix 上新建为 0600,已存在的文件也收紧到 0600。
fn open_private_append(path: &Path) -> std::io::Result<File> {
    let mut opts = OpenOptions::new();
    opts.create(true).append(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::{OpenOptionsExt as _, PermissionsExt as _};
        opts.mode(0o600);
        let f = opts.open(path)?;
        f.set_permissions(std::fs::Permissions::from_mode(0o600))?;
        Ok(f)
    }
    #[cfg(not(unix))]
    {
        opts.open(path)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn entry(ts: i64) -> CdpLogEntry {
        CdpLogEntry {
            ts,
            direction: "send".into(),
            method: Some("Echo".into()),
            params: None,
            result: None,
            event: None,
        }
    }

    async fn wait_until(mut cond: impl FnMut() -> bool) -> bool {
        for _ in 0..200 {
            if cond() {
                return true;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        cond()
    }

    fn lines(path: &Path) -> Vec<String> {
        std::fs::read_to_string(path)
            .map(|s| s.lines().map(str::to_string).collect())
            .unwrap_or_default()
    }

    #[test]
    fn disabled_buffer_hands_out_sender_without_receivers() {
        let b = CdpLogBuffer::new(false);
        assert!(!b.enabled());
        let tx = b.sender().unwrap();
        assert_eq!(
            tx.receiver_count(),
            0,
            "未开启时没有订阅者,session 侧据此跳过构造日志"
        );
        assert!(b.recent(10).is_empty());
    }

    #[test]
    fn recent_returns_latest_in_order_and_is_bounded() {
        let b = CdpLogBuffer::new(true);
        assert!(b.sender().is_some());
        for i in 0..(CDP_LOG_CAPACITY as i64 + 5) {
            b.push(entry(i));
        }
        assert_eq!(b.len(), CDP_LOG_CAPACITY);
        let last3: Vec<i64> = b.recent(3).iter().map(|e| e.ts).collect();
        let n = CDP_LOG_CAPACITY as i64 + 5;
        assert_eq!(last3, vec![n - 3, n - 2, n - 1]);
        assert_eq!(b.recent(0).len(), 0);
    }

    #[tokio::test]
    async fn ensure_collector_stores_broadcast_entries_and_forwards_to_ui() {
        let b = CdpLogBuffer::new(true);
        let mut ui = b.subscribe();
        b.ensure_collector(None);
        let tx = b.sender().unwrap();
        assert!(wait_until(|| tx.receiver_count() == 1).await);
        tx.send(entry(1)).unwrap();
        tx.send(entry(2)).unwrap();
        assert!(wait_until(|| b.len() == 2).await);
        let ts: Vec<i64> = b.recent(10).iter().map(|e| e.ts).collect();
        assert_eq!(ts, vec![1, 2]);
        assert_eq!(ui.recv().await.unwrap().ts, 1);
        assert_eq!(ui.recv().await.unwrap().ts, 2);
        // 幂等:再调不会多开一个订阅者
        b.ensure_collector(None);
        assert_eq!(tx.receiver_count(), 1);
    }

    #[tokio::test]
    async fn ensure_collector_is_noop_when_disabled() {
        let b = CdpLogBuffer::new(false);
        b.ensure_collector(None);
        assert_eq!(b.sender().unwrap().receiver_count(), 0);
    }

    #[tokio::test]
    async fn set_enabled_toggles_collector_and_appends_debug_file() {
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join(CDP_DEBUG_LOG_FILE);
        let b = CdpLogBuffer::new(false);
        let tx = b.sender().unwrap();

        b.set_enabled(true, Some(dir.path()));
        assert!(b.enabled());
        assert!(wait_until(|| tx.receiver_count() == 1).await);
        tx.send(entry(1)).unwrap();
        assert!(wait_until(|| b.len() == 1).await);
        assert!(
            wait_until(|| lines(&file).len() == 1).await,
            "{:?}",
            lines(&file)
        );
        let first: serde_json::Value = serde_json::from_str(&lines(&file)[0]).unwrap();
        assert_eq!(first["direction"], "send");
        assert_eq!(first["method"], "Echo");
        assert_eq!(first["ts"], 1);
        assert!(first.get("params").is_none(), "None 字段省略");

        // 关:订阅者消失,之后的条目既不进缓冲也不写文件
        b.set_enabled(false, Some(dir.path()));
        assert!(!b.enabled());
        assert!(wait_until(|| tx.receiver_count() == 0).await);
        let _ = tx.send(entry(2));
        tokio::time::sleep(Duration::from_millis(30)).await;
        assert_eq!(b.len(), 1);
        assert_eq!(lines(&file).len(), 1);

        // 再开:追加写,缓冲保留旧条目
        b.set_enabled(true, Some(dir.path()));
        assert!(wait_until(|| tx.receiver_count() == 1).await);
        tx.send(entry(3)).unwrap();
        assert!(wait_until(|| b.len() == 2).await);
        assert!(wait_until(|| lines(&file).len() == 2).await);
        let ts: Vec<i64> = b.recent(10).iter().map(|e| e.ts).collect();
        assert_eq!(ts, vec![1, 3]);
    }

    #[test]
    fn cdp_debug_file_rotates_when_over_cap() {
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join(CDP_DEBUG_LOG_FILE);
        let one = serde_json::to_string(&entry(1)).unwrap().len() as u64 + 1;
        // 上限只容得下两行:第三行触发轮转
        let mut w = CdpFileWriter::open(&file, one * 2).unwrap();
        for i in 1..=3 {
            w.write_entry(&entry(i)).unwrap();
        }
        let rotated = dir.path().join(format!("{CDP_DEBUG_LOG_FILE}.1"));
        assert_eq!(lines(&rotated).len(), 2);
        assert_eq!(lines(&file).len(), 1);
        assert!(std::fs::metadata(&file).unwrap().len() <= one * 2);
        // 再开:接着已有大小计数,不会无限增长
        drop(w);
        let mut w = CdpFileWriter::open(&file, one * 2).unwrap();
        w.write_entry(&entry(4)).unwrap();
        w.write_entry(&entry(5)).unwrap();
        assert_eq!(lines(&file).len(), 1);
        let ts: Vec<i64> = lines(&rotated)
            .iter()
            .map(|l| {
                serde_json::from_str::<serde_json::Value>(l).unwrap()["ts"]
                    .as_i64()
                    .unwrap()
            })
            .collect();
        assert_eq!(ts, vec![3, 4]);
    }

    #[cfg(unix)]
    #[test]
    fn cdp_debug_file_is_private_on_unix() {
        use std::os::unix::fs::PermissionsExt as _;
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join(CDP_DEBUG_LOG_FILE);
        // 旧版本留下的 0644 文件也要被收紧
        std::fs::write(&file, b"").unwrap();
        std::fs::set_permissions(&file, std::fs::Permissions::from_mode(0o644)).unwrap();
        let mut w = CdpFileWriter::open(&file, CDP_DEBUG_LOG_MAX_BYTES).unwrap();
        w.write_entry(&entry(1)).unwrap();
        let mode = std::fs::metadata(&file).unwrap().permissions().mode() & 0o777;
        assert_eq!(mode, 0o600);
        // 轮转后新文件同样是 0600
        let mut w = CdpFileWriter::open(&file, 1).unwrap();
        w.write_entry(&entry(2)).unwrap();
        let mode = std::fs::metadata(&file).unwrap().permissions().mode() & 0o777;
        assert_eq!(mode, 0o600);
        drop(w);
    }

    #[test]
    fn test_state_builds_with_config_on_disk() {
        let dir = tempfile::tempdir().unwrap();
        let state = crate::commands::testing::test_state_blocking(dir.path());
        assert!(state.config_path.exists());
        assert!(state.started_at > 0);
        assert!(state.cdp_log.enabled());
        assert!(dir.path().join("relay.db").exists(), "Storage 已打开");
        assert!(dir.path().join("uploads").is_dir());
        assert_eq!(state.scheduler.running_count(), 0);
        assert!(!state
            .api_listening
            .load(std::sync::atomic::Ordering::Relaxed));
        assert!(!state.shutdown.is_cancelled());
    }
}
