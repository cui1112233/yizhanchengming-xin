//! Task 6.1 稳定性压测用的无界面 relay:真实的 `Storage` / `Scheduler` / runner / HTTP API / 日志 / GC,
//! 只把豆包后端换成本地假实现(不连豆包、不耗配额),供 `tests/soak.sh` 提交大量任务并采样 RSS 与延迟。
//!
//! ```text
//! SOAK_DIR=/tmp/x SOAK_PORT=18787 SOAK_TOKEN=t cargo run --release --example soak_server
//! ```
//!
//! 环境变量:`SOAK_DIR`(数据目录,必填)、`SOAK_PORT`(默认 18787)、`SOAK_TOKEN`(默认 `soak-token`)、
//! `SOAK_CAPTURE_MS`(假捕获耗时,默认 200)、`SOAK_GC_SECS`(GC 周期,默认 3600)、
//! `SOAK_CDP_PORT`(设置后连接本机真豆包并跑 10s 心跳,用于"连着真豆包空闲挂机"的 RSS / CPU 观察;
//! 生成后端仍是假的,提交任务也不会碰豆包、不耗配额)。
//! 收到 SIGINT / SIGTERM 或 stdin 关闭以外的情况都一直运行;脚本用完 kill 掉。

use std::path::{Path, PathBuf};
use std::sync::Arc;
use std::time::Duration;

use async_trait::async_trait;
use doubao_relay_lib::config::Config;
use doubao_relay_lib::doubao::download::Downloaded;
use doubao_relay_lib::doubao::generation::{cancelled, StageFn};
use doubao_relay_lib::doubao::{
    Account, AccountList, CaptureEvent, CaptureOptions, CaptureOutcome, ConnTiming,
    DoubaoConnection, SubmissionIdentity, SubmissionPlan, SubmissionReport,
};
use doubao_relay_lib::error::{RelayError, RelayResult};
use doubao_relay_lib::logging::{self, now_ms};
use doubao_relay_lib::platform::{Platform, PlatformError};
use doubao_relay_lib::state::{AppState, CdpLogBuffer};
use doubao_relay_lib::storage::Storage;
use doubao_relay_lib::tasks::{GenerateRequest, GenerationBackend, TaskStage, WatermarkSource};
use tokio::sync::{mpsc, watch};
use tokio_util::sync::CancellationToken;

const ACCOUNT: &str = "压测账号";

struct SoakPlatform {
    dir: PathBuf,
}

#[async_trait]
impl Platform for SoakPlatform {
    async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
        Err(PlatformError::DoubaoNotFound("soak".into()))
    }
    async fn launch_doubao_with_cdp(&self, _: &Path, _: u16) -> Result<(), PlatformError> {
        Err(PlatformError::LaunchFailed("soak".into()))
    }
    async fn kill_doubao(&self) -> Result<(), PlatformError> {
        Ok(())
    }
    async fn is_doubao_running(&self) -> bool {
        false
    }
    fn app_data_dir(&self) -> PathBuf {
        self.dir.clone()
    }
    fn log_dir(&self) -> PathBuf {
        self.dir.join("logs")
    }
    fn config_path(&self) -> PathBuf {
        self.dir.join("config.toml")
    }
    async fn reveal_in_file_manager(&self, _: &Path) -> Result<(), PlatformError> {
        Ok(())
    }
    fn platform_name(&self) -> &'static str {
        "soak"
    }
    fn listening_process_exe(&self, _: u16) -> Option<PathBuf> {
        None
    }
    fn is_doubao_exe_path(&self, _: &Path) -> bool {
        false
    }
    fn doubao_version(&self) -> Option<String> {
        None
    }
}

/// 假豆包:单账号;submit 立即成功;capture 等 `capture_ms` 后给 1 张"原图";下载写一张 64×64 PNG。
struct SoakBackend {
    capture_ms: u64,
    png: Vec<u8>,
}

#[async_trait]
impl GenerationBackend for SoakBackend {
    async fn accounts(&self, _force: bool) -> RelayResult<AccountList> {
        Ok(AccountList {
            current: Some(ACCOUNT.into()),
            accounts: vec![Account {
                name: ACCOUNT.into(),
                subscription: Some("体验套餐".into()),
                label: format!("{ACCOUNT} 体验套餐"),
                is_current: true,
                is_ambiguous: false,
            }],
            read_at: now_ms(),
        })
    }

    async fn ensure_account(&self, account: &str) -> RelayResult<()> {
        if account == ACCOUNT {
            Ok(())
        } else {
            Err(RelayError::account_not_found(account, vec![ACCOUNT.into()]))
        }
    }

    async fn submit(
        &self,
        _request: &GenerateRequest,
        _plan: &SubmissionPlan,
        _reference_paths: &[PathBuf],
        cancel: &CancellationToken,
        on_stage: StageFn<'_>,
    ) -> RelayResult<SubmissionReport> {
        if cancel.is_cancelled() {
            return Err(cancelled("prepare"));
        }
        on_stage(TaskStage::Preparing, Some("模型与参数已就绪".into()));
        Ok(SubmissionReport {
            identity: Some(SubmissionIdentity {
                local_message_id: Some(format!("lm{}", now_ms())),
                message_id: None,
                conversation_id: Some("7001".into()),
            }),
            send_started_at: now_ms(),
            generation_requests: Vec::new(),
        })
    }

    async fn capture(
        &self,
        _opts: CaptureOptions,
        cancel: &CancellationToken,
        _confirm_rx: &mut watch::Receiver<bool>,
        events: mpsc::Sender<CaptureEvent>,
    ) -> RelayResult<CaptureOutcome> {
        let _ = events
            .send(CaptureEvent::Stage(TaskStage::Rendering, None))
            .await;
        tokio::select! {
            _ = cancel.cancelled() => return Err(cancelled("capture")),
            _ = tokio::time::sleep(Duration::from_millis(self.capture_ms)) => {}
        }
        Ok(CaptureOutcome {
            originals: vec![format!("https://soak.invalid/{}.png", now_ms())],
            watermark_source: WatermarkSource::ConversationOriginal,
            conversation_id: Some("7001".into()),
            submission_id: Some(format!("{}", now_ms())),
        })
    }

    async fn download_image(
        &self,
        _url: &str,
        dest_dir: &Path,
        index: u32,
    ) -> RelayResult<Downloaded> {
        tokio::fs::create_dir_all(dest_dir)
            .await
            .map_err(|e| RelayError::internal(format!("建目录失败:{e}")))?;
        let path = dest_dir.join(format!("image_{}.png", index + 1));
        tokio::fs::write(&path, &self.png)
            .await
            .map_err(|e| RelayError::internal(format!("写文件失败:{e}")))?;
        Ok(Downloaded {
            path,
            mime: "image/png".into(),
            bytes: self.png.len() as u64,
            width: Some(64),
            height: Some(64),
        })
    }

    async fn download_video(&self, _url: &str, _dest_dir: &Path) -> RelayResult<Downloaded> {
        Err(RelayError::internal("压测后端不生成视频"))
    }
}

fn png_bytes() -> anyhow::Result<Vec<u8>> {
    let img = image::RgbImage::from_fn(64, 64, |x, y| image::Rgb([x as u8 * 4, y as u8 * 4, 128]));
    let mut out = std::io::Cursor::new(Vec::new());
    image::DynamicImage::ImageRgb8(img).write_to(&mut out, image::ImageFormat::Png)?;
    Ok(out.into_inner())
}

fn env_or<T: std::str::FromStr>(key: &str, default: T) -> T {
    std::env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let dir =
        PathBuf::from(std::env::var("SOAK_DIR").map_err(|_| anyhow::anyhow!("缺少 SOAK_DIR"))?);
    std::fs::create_dir_all(&dir)?;
    let port: u16 = env_or("SOAK_PORT", 18787);
    let token = std::env::var("SOAK_TOKEN").unwrap_or_else(|_| "soak-token".into());
    let capture_ms: u64 = env_or("SOAK_CAPTURE_MS", 200);
    let gc_secs: u64 = env_or("SOAK_GC_SECS", 3600);

    let platform: Arc<dyn Platform> = Arc::new(SoakPlatform { dir: dir.clone() });
    let config_path = platform.config_path();
    let mut config = Config::load_or_create(&config_path)?;
    config.server.host = "127.0.0.1".into();
    config.server.port = port;
    config.server.bearer_token = token;
    config.doubao.auto_connect_on_startup = false;
    config.doubao.cdp_port = 1;
    config.save(&config_path)?;

    let logs = logging::init(&config.logging.level, &platform.log_dir())?;
    let cdp_log = CdpLogBuffer::new(false);
    let real_cdp_port: Option<u16> = std::env::var("SOAK_CDP_PORT")
        .ok()
        .and_then(|v| v.parse().ok());
    let doubao = match real_cdp_port {
        Some(cdp_port) => DoubaoConnection::new_with(
            doubao_relay_lib::platform::create_platform(),
            "127.0.0.1".into(),
            cdp_port,
            cdp_log.sender(),
            ConnTiming::default(),
        ),
        None => DoubaoConnection::new_with(
            Arc::clone(&platform),
            "127.0.0.1".into(),
            1,
            cdp_log.sender(),
            ConnTiming::default(),
        ),
    };
    if real_cdp_port.is_some() {
        match doubao.connect(false).await {
            Ok(action) => tracing::info!(action = ?action, "压测:已连接真豆包(只跑心跳)"),
            Err(e) => tracing::warn!(code = e.code, error = %e, "压测:连接真豆包失败"),
        }
        doubao.spawn_heartbeat(doubao_relay_lib::state::HEARTBEAT_INTERVAL);
    }
    let storage = Storage::open(&dir).await?;
    let backend: Arc<dyn GenerationBackend> = Arc::new(SoakBackend {
        capture_ms,
        png: png_bytes()?,
    });
    let state = AppState::new(
        config,
        config_path,
        platform,
        doubao,
        logs,
        cdp_log,
        storage,
        backend,
    );
    doubao_relay_lib::recover_after_restart(&state).await;
    state.scheduler.spawn_loop();
    {
        let state = Arc::clone(&state);
        tokio::spawn(async move {
            let mut tick = tokio::time::interval(Duration::from_secs(gc_secs.max(1)));
            tick.tick().await;
            loop {
                tick.tick().await;
                let cfg = state.config().await.storage;
                match state.storage.gc(now_ms(), &cfg).await {
                    Ok(r) => tracing::info!(
                        result_files_removed = r.result_files_removed,
                        tasks_removed = r.tasks_removed,
                        "压测 GC 完成"
                    ),
                    Err(e) => tracing::warn!(code = e.code, error = %e, "压测 GC 失败"),
                }
            }
        });
    }
    println!(
        "soak_server listening on 127.0.0.1:{port} pid={}",
        std::process::id()
    );
    let shutdown = state.shutdown.clone();
    doubao_relay_lib::http::serve(Arc::clone(&state), "127.0.0.1", port, shutdown).await
}
