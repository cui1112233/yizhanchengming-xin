//! 豆包连接管理:会话(chat / background)、心跳、断线重连、账号缓存与账号切换。
//!
//! - 会话:`connect_sessions` 从 `/json/list` 选出**恰好 1 个** chat page(+ 可选 background),
//!   每个 session `Network.enable`;已有未关闭且 target 未变的 chat session 直接复用。
//! - 心跳:`heartbeat` 对 chat session 发 `Target.getTargets`(5s 超时);失败即标记断开、关掉老 socket、
//!   每轮查一次豆包进程(`running` 立刻反映进程是否还在,UC-06 要求 15s 内显示"豆包已退出"),
//!   进程还在就等 2s(`reconnect_delay`,spec 09)后重连并复测,之后后台心跳每 2s 重试;
//!   进程还在但自断开起 20s(`restart_after`)仍连不上 → 升级为 `connect(true)` 重启豆包(每次断线一次,
//!   多开 chat 窗口不算);自断开起超过 20s 仍连不上且豆包进程不在 → 发 `Crashed`(只发一次,不自动拉起)。
//! - 捕获中途断线:`await_chat_session` 按同样的间隔重试到 `restart_after`,重连成功后捕获继续(Task 6.1)。
//! - 切换账号:`switch_account` 在 `switch_lock` 下读列表校验目标 → 菜单里快节奏点击 →
//!   20s 验证循环(`current()` 带注入守卫;session 断开就重连;出现 `client-login-modal` page target
//!   即判定目标账号需要重新登录,见 `docs/analysis/10-live-probe-macos.md` Task 1.2)。
//!
//! chat 页面建立会话后启用后台活跃与焦点模拟,让提交生成时无需 `Page.bringToFront` 抢占用户窗口。

use std::sync::atomic::{AtomicBool, AtomicI64, AtomicUsize, Ordering};
use std::sync::{Arc, Mutex as StdMutex};
use std::time::{Duration, Instant};

use serde::Serialize;
use serde_json::{json, Value};
use tokio::sync::{broadcast, Mutex, RwLock};
use tokio::task::JoinHandle;
use tokio_util::sync::CancellationToken;
use url::Url;

use crate::cdp::{
    list_targets, now_ms, select_pages, CdpError, CdpLogEntry, CdpSession, CdpTarget,
};
use crate::error::{RelayError, RelayResult};
use crate::platform::Platform;

use super::accounts::{
    click_account, probe_current, read_accounts_with, Account, ReadTiming, SwitchTiming,
};
use super::bootstrap::{
    ensure_doubao_started, verify_cdp_owner, wait_for_cdp, StartupAction, LAUNCH_TIMEOUT,
};
use super::compat::{ui_compat_check, UiCompat};
use super::AccountList;

/// 事件广播容量。
const EVENT_CAPACITY: usize = 256;
/// spec 09:`Network.enable` 的缓冲参数。
fn network_enable_params() -> Value {
    json!({ "maxTotalBufferSize": 16_777_216, "maxResourceBufferSize": 8_388_608 })
}

/// `doubao.status` 的返回(domain 类型,`snake_case`;IPC 层转 camelCase)。
#[derive(Clone, Debug, Serialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub struct DoubaoStatus {
    pub running: bool,
    pub cdp_connected: bool,
    pub cdp_port: u16,
    pub doubao_version_detected: Option<String>,
    /// `ok` | `warn` | `error` | `unknown`。
    pub ui_compat_check: String,
    pub ui_compat_details: UiCompat,
    pub chat_targets: usize,
    pub reconnecting: bool,
    /// Unix epoch 毫秒。
    pub last_check_at: i64,
}

/// 连接层对外事件(Task 1.9 转发为 Tauri 事件)。
#[derive(Clone, Debug)]
pub enum ConnEvent {
    StatusChanged(DoubaoStatus),
    AccountListChanged(AccountList),
    AccountSwitched {
        previous: Option<String>,
        current: String,
        switched_at: i64,
    },
    /// 断线超过 20s 仍连不上且豆包进程已不在。
    Crashed,
}

/// `account.switch` 的返回。
#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub struct SwitchResult {
    pub ok: bool,
    pub previous: Option<String>,
    pub current: String,
    /// Unix epoch 毫秒。
    pub switched_at: i64,
}

/// 各类时间参数(默认值即 spec;测试用小值)。
#[derive(Clone, Copy, Debug)]
pub struct ConnTiming {
    /// 心跳命令超时(spec:5s 无响应判断开)。
    pub heartbeat_timeout: Duration,
    /// 断开后多久仍连不上才判定崩溃(spec:20s)。
    pub crash_after: Duration,
    /// 账号列表缓存时间。
    pub account_cache_ttl: Duration,
    pub read: ReadTiming,
    pub switch: SwitchTiming,
    /// relay 自己拉起豆包(`Started`/`Restarted`)后,最多等页面渲染出输入框/登录按钮多久
    /// 再做 UI 自检(Task 1.10 真机:CDP 就绪 ≈0.9s,但页面要再过数秒才有 DOM,
    /// 立刻自检会得 `error`、首次读账号报 `doubao_ui_changed`)。
    pub page_ready_timeout: Duration,
    /// 页面就绪轮询间隔。
    pub page_ready_poll: Duration,
    /// spec 09:心跳判定断开、关掉老 socket 后等多久再重新握手;断线期间后台心跳也按这个间隔重试。
    pub reconnect_delay: Duration,
    /// spec 09:自首次心跳失败起这么久仍未重连成功(且豆包进程还在)→ 升级为重启豆包(`connect(true)`);
    /// 也是捕获中途断线时 [`await_chat_session`](DoubaoConnection::await_chat_session) 的最长等待。
    pub restart_after: Duration,
}

impl Default for ConnTiming {
    fn default() -> Self {
        Self {
            heartbeat_timeout: Duration::from_secs(5),
            crash_after: Duration::from_secs(20),
            account_cache_ttl: Duration::from_secs(2),
            read: ReadTiming::default(),
            switch: SwitchTiming::default(),
            page_ready_timeout: Duration::from_secs(15),
            page_ready_poll: Duration::from_millis(500),
            reconnect_delay: Duration::from_secs(2),
            restart_after: Duration::from_secs(20),
        }
    }
}

/// 页面"可操作"判据:加载完成且渲染出聊天输入框(已登录)或登录入口(未登录)。
/// 以 `(() => document.readyState` 开头,mock 据此识别。
pub const PAGE_READY_EXPRESSION: &str = r#"(() => document.readyState === 'complete' && (!!document.querySelector('[contenteditable="true"]') || /登录/.test(document.body ? document.body.innerText : '')))()"#;

/// 轮询 [`PAGE_READY_EXPRESSION`] 直到为 true 或超时;返回是否就绪(超时只是"没等到",不算错误)。
/// 求值报错(页面正在导航等)视为未就绪继续等。
pub async fn wait_for_page_ready(session: &CdpSession, timeout: Duration, poll: Duration) -> bool {
    let deadline = Instant::now() + timeout;
    loop {
        let ready = matches!(
            tokio::time::timeout(
                poll.max(Duration::from_millis(50)) * 4,
                session.evaluate_as::<bool>(PAGE_READY_EXPRESSION)
            )
            .await,
            Ok(Ok(true))
        );
        if ready {
            return true;
        }
        if Instant::now() >= deadline {
            return false;
        }
        tokio::time::sleep(poll).await;
    }
}

#[derive(Default)]
struct Sessions {
    chat: Option<Arc<CdpSession>>,
    background: Option<Arc<CdpSession>>,
}

/// 重连中的断线是否仍在宽限期内(未满 `crash_after`)。没有断线记录(不应出现)按宽限期内处理。
fn reconnect_in_grace(outage: Option<Duration>, crash_after: Duration) -> bool {
    outage.is_none_or(|d| d < crash_after)
}

/// 一次断线(从首个心跳失败起算)。
struct Outage {
    since: Instant,
    crashed_emitted: bool,
    /// 本次断线已经升级重启过豆包(无论成败,同一次断线只重启一次,避免反复杀豆包)。
    restart_attempted: bool,
    /// 最近一次重连失败的原因不会因快速重试 / 重启而好转(用户多开了 chat 窗口、9707 被非豆包进程
    /// 占用):回到常规心跳节奏,也不升级重启。下一次重连失败时按新原因重新判定。
    hold: bool,
}

pub struct DoubaoConnection {
    platform: Arc<dyn Platform>,
    host: String,
    port: u16,
    timing: ConnTiming,
    sessions: RwLock<Sessions>,
    status: RwLock<DoubaoStatus>,
    account_cache: RwLock<Option<(AccountList, Instant)>>,
    /// 串行化 doubao-chat 页面上的多步 DOM 操作:账号菜单(读列表 / 切换)与任务提交
    /// (`generation::submit` 的技能 / 模型菜单、比例时长浮层、参考图上传,见 [`lock_page`](Self::lock_page))。
    switch_lock: Mutex<()>,
    /// 正在使用 chat 页面的任务操作数(提交 + 捕获,见 [`PageBusyGuard`])。> 0 时非强制的账号读取
    /// 宁可返回过期缓存也不打开账号菜单(菜单抢焦点、关闭时派发的 Escape 会关掉任务的浮层)。
    page_busy: AtomicUsize,
    /// 串行化 `connect_sessions`(心跳与切换验证可能同时发现断线)。
    connect_lock: Mutex<()>,
    events: broadcast::Sender<ConnEvent>,
    cdp_debug: Option<broadcast::Sender<CdpLogEntry>>,
    outage: StdMutex<Option<Outage>>,
    /// 至少成功连过一次;之前心跳不做任何事(用户未连接时不去碰豆包)。
    ever_connected: AtomicBool,
    /// 最后一次确认 CDP 连通(心跳成功 / 建会话成功)的时刻,Unix epoch 毫秒;从未连通为 0。
    /// `doubao_crashed{last_seen_at}`(11-error-codes)用它。
    last_seen_at: AtomicI64,
    /// 应用关闭(Task 5.2 关闭流程)后触发:心跳循环退出、不再(重)建会话。
    closing: CancellationToken,
}

impl std::fmt::Debug for DoubaoConnection {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("DoubaoConnection")
            .field("host", &self.host)
            .field("port", &self.port)
            .finish()
    }
}

/// [`DoubaoConnection::mark_page_busy`] 的 guard,drop 时减计数。
pub struct PageBusyGuard<'a>(&'a AtomicUsize);

impl Drop for PageBusyGuard<'_> {
    fn drop(&mut self) {
        self.0.fetch_sub(1, Ordering::SeqCst);
    }
}

/// [`DoubaoConnection::lock_page`] 的 guard:持有页面锁(`switch_lock`)并标记页面忙。
pub struct PageGuard<'a> {
    _lock: tokio::sync::MutexGuard<'a, ()>,
    _busy: PageBusyGuard<'a>,
}

impl DoubaoConnection {
    pub fn new(platform: Arc<dyn Platform>, host: String, port: u16) -> Arc<Self> {
        Self::new_with(platform, host, port, None, ConnTiming::default())
    }

    /// 同 [`new`](Self::new),可带 CDP 调试日志通道(`cdp_debug_log=true`)与自定义时间参数(测试)。
    pub fn new_with(
        platform: Arc<dyn Platform>,
        host: String,
        port: u16,
        cdp_debug: Option<broadcast::Sender<CdpLogEntry>>,
        timing: ConnTiming,
    ) -> Arc<Self> {
        let (events, _) = broadcast::channel(EVENT_CAPACITY);
        Arc::new(Self {
            platform,
            host,
            port,
            timing,
            sessions: RwLock::new(Sessions::default()),
            status: RwLock::new(DoubaoStatus {
                running: false,
                cdp_connected: false,
                cdp_port: port,
                doubao_version_detected: None,
                ui_compat_check: "unknown".into(),
                ui_compat_details: UiCompat::default(),
                chat_targets: 0,
                reconnecting: false,
                last_check_at: 0,
            }),
            account_cache: RwLock::new(None),
            switch_lock: Mutex::new(()),
            page_busy: AtomicUsize::new(0),
            connect_lock: Mutex::new(()),
            events,
            cdp_debug,
            outage: StdMutex::new(None),
            ever_connected: AtomicBool::new(false),
            last_seen_at: AtomicI64::new(0),
            closing: CancellationToken::new(),
        })
    }

    pub fn subscribe(&self) -> broadcast::Receiver<ConnEvent> {
        self.events.subscribe()
    }

    /// CDP 主机(`/json/*` 与 WebSocket 都连它)。捕获管线据此做 chat target 唯一性检查。
    pub fn host(&self) -> &str {
        &self.host
    }

    /// CDP 端口。
    pub fn port(&self) -> u16 {
        self.port
    }

    pub async fn status(&self) -> DoubaoStatus {
        self.status.read().await.clone()
    }

    /// 最后一次确认 CDP 连通的时刻(Unix epoch 毫秒;从未连通为 0)。
    pub fn last_seen_at(&self) -> i64 {
        self.last_seen_at.load(Ordering::SeqCst)
    }

    /// 豆包进程是否在跑(平台层实时查询)。runner 据此把断线细分为 `doubao_crashed`。
    pub async fn is_doubao_running(&self) -> bool {
        self.platform.is_doubao_running().await
    }

    /// 接受新任务的门槛(spec 08"豆包崩溃时新请求直接返回 503",Task 4.2):
    /// CDP 已连通 → Ok;豆包进程还在、心跳正在重连且断线未满 `crash_after`(切账号 / 页面 reload
    /// 的短暂断线)→ Ok(任务排队,runner 启动时 `chat_session()` 会重连);其余(从未连接、
    /// 进程已退出、断线超过 `crash_after`——例如豆包被不带调试端口重新打开,`reconnecting` 会一直为
    /// true)→ `doubao_not_ready{running, cdp_connected, reconnecting, outage_ms}`(503)。
    ///
    /// 注意:从未连接过时也是 503——`auto_connect_on_startup=false` 时需先在界面点"连接"
    /// (或 Tauri `doubao_connect`)后 `/generate` 才可用,不再在提交时隐式连接。
    pub async fn ensure_ready(&self) -> RelayResult<()> {
        let st = self.status.read().await;
        let outage = self.outage_lock().as_ref().map(|o| o.since.elapsed());
        if st.cdp_connected
            || (st.running
                && st.reconnecting
                && reconnect_in_grace(outage, self.timing.crash_after))
        {
            return Ok(());
        }
        Err(RelayError::doubao_not_ready().with_details(json!({
            "running": st.running,
            "cdp_connected": st.cdp_connected,
            "reconnecting": st.reconnecting,
            "outage_ms": outage.map(|d| d.as_millis() as u64),
        })))
    }

    /// `outage` 的临界区只有几条赋值;即便某个线程在持锁时 panic,也直接取回数据而不是再 panic。
    fn outage_lock(&self) -> std::sync::MutexGuard<'_, Option<Outage>> {
        self.outage
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// 改状态并刷新 `last_check_at`;除 `last_check_at` 外有变化才发 `StatusChanged`。
    async fn update_status(&self, f: impl FnOnce(&mut DoubaoStatus)) {
        let mut guard = self.status.write().await;
        let before = guard.clone();
        f(&mut guard);
        guard.last_check_at = now_ms();
        let mut cmp = guard.clone();
        cmp.last_check_at = before.last_check_at;
        if cmp != before {
            let _ = self.events.send(ConnEvent::StatusChanged(guard.clone()));
        }
    }

    // ------------------------------------------------------------------
    // 连接
    // ------------------------------------------------------------------

    /// 连接豆包:`force_restart` → 杀掉并带调试参数重启;否则 `ensure_doubao_started`。
    /// 之后建会话、UI 兼容性自检、读版本,发 `StatusChanged`。
    pub async fn connect(&self, force_restart: bool) -> RelayResult<StartupAction> {
        let result = self.connect_inner(force_restart).await;
        if let Err(e) = &result {
            tracing::warn!(code = e.code, error = %e, "连接豆包失败");
            let running = self.platform.is_doubao_running().await;
            self.update_status(|s| {
                s.running = running;
                s.cdp_connected = false;
                s.reconnecting = false;
            })
            .await;
        }
        result
    }

    async fn connect_inner(&self, force_restart: bool) -> RelayResult<StartupAction> {
        let had_session = self.sessions.read().await.chat.is_some();
        let mut action = if force_restart {
            self.close_sessions().await;
            let exe = self.platform.find_doubao_exe().await?;
            // 端口被别的程序占着时重启豆包也拿不到它,别白杀用户的豆包。
            verify_cdp_owner(&*self.platform, self.port).await?;
            tracing::info!(exe = %exe.display(), port = self.port, "强制重启豆包");
            self.platform.kill_doubao().await?;
            self.platform
                .launch_doubao_with_cdp(&exe, self.port)
                .await?;
            wait_for_cdp(&self.host, self.port, LAUNCH_TIMEOUT).await?;
            // 启动后再校验一次监听者(下面的 `connect_sessions_inner` 因会话已关也会重跑一遍)。
            verify_cdp_owner(&*self.platform, self.port).await?;
            StartupAction::Restarted
        } else {
            ensure_doubao_started(&*self.platform, self.port).await?
        };
        let reused = self.connect_sessions_inner(false).await?;
        if action == StartupAction::AlreadyConnected && !reused && had_session {
            action = StartupAction::Reconnected;
        }
        let chat = self.current_chat().await?;
        if matches!(action, StartupAction::Started | StartupAction::Restarted) {
            let started = Instant::now();
            let ready = wait_for_page_ready(
                &chat,
                self.timing.page_ready_timeout,
                self.timing.page_ready_poll,
            )
            .await;
            let waited_ms = started.elapsed().as_millis() as u64;
            if ready {
                tracing::info!(waited_ms, "豆包页面已渲染完成");
            } else {
                tracing::warn!(waited_ms, "等待豆包页面渲染超时,继续自检");
            }
        }
        let compat = ui_compat_check(&chat).await;
        let platform = Arc::clone(&self.platform);
        let version = tokio::task::spawn_blocking(move || platform.doubao_version())
            .await
            .unwrap_or(None);
        self.outage_lock().take();
        self.update_status(|s| {
            s.running = true;
            s.cdp_connected = true;
            s.reconnecting = false;
            s.chat_targets = 1;
            s.doubao_version_detected = version.clone();
            s.ui_compat_check = compat.status.clone();
            s.ui_compat_details = compat.clone();
        })
        .await;
        tracing::info!(
            action = ?action,
            port = self.port,
            version = version.as_deref().unwrap_or(""),
            ui_compat = compat.status.as_str(),
            "豆包已连接"
        );
        Ok(action)
    }

    /// `list_targets` → `select_pages`;chat 必须恰好 1 个;连接 chat(+ background 可选),
    /// 每个 session `Network.enable`。已有未关闭且 target 未变的 session 直接复用。
    pub async fn connect_sessions(&self) -> RelayResult<()> {
        self.connect_sessions_inner(true).await.map(|_| ())
    }

    /// 返回是否复用了现有 chat session。`announce` 为 true 时顺带把状态标成已连接。
    async fn connect_sessions_inner(&self, announce: bool) -> RelayResult<bool> {
        if self.closing.is_cancelled() {
            return Err(RelayError::doubao_cdp_disconnected()
                .with_details(json!({ "cause": "relay_shutting_down" })));
        }
        let _guard = self.connect_lock.lock().await;
        let targets = list_targets(&self.host, self.port).await?;
        let (chat, background) = select_pages(&targets, self.port);
        if chat.len() != 1 {
            tracing::warn!(chat_targets = chat.len(), "doubao-chat page 数量不是 1");
            self.update_status(|s| s.chat_targets = chat.len()).await;
            return Err(RelayError::doubao_not_ready().with_details(json!({
                "chat_targets": chat.len(),
                "reason": if chat.is_empty() { "chat_page_missing" } else { "multiple_chat_pages" },
            })));
        }
        let target = &chat[0];
        let mut sessions = self.sessions.write().await;
        let reused = sessions
            .chat
            .as_ref()
            .is_some_and(|s| !s.is_closed() && s.target_id == target.id);
        if !reused {
            // spec 09:每次(重)建会话都重跑监听者校验;target 列表已确认恰好 1 个 chat page,
            // 查不到监听者时以此兜底。
            verify_cdp_owner(&*self.platform, self.port).await?;
            let session = self.open_session(target, true).await?;
            if let Some(old) = sessions.chat.replace(session) {
                old.close().await;
            }
            tracing::info!(target_id = %target.id, "chat session 已建立");
        }
        let bg_ok = sessions
            .background
            .as_ref()
            .is_some_and(|s| !s.is_closed() && background.iter().any(|t| t.id == s.target_id));
        if !bg_ok {
            if let Some(old) = sessions.background.take() {
                old.close().await;
            }
            if let Some(t) = background.first() {
                match self.open_session(t, false).await {
                    Ok(s) => {
                        tracing::debug!(target_id = %t.id, "background session 已建立");
                        sessions.background = Some(s);
                    }
                    Err(e) => {
                        tracing::warn!(error = %e, "background session 连接失败,忽略");
                    }
                }
            }
        }
        drop(sessions);
        self.ever_connected.store(true, Ordering::SeqCst);
        self.last_seen_at.store(now_ms(), Ordering::SeqCst);
        if announce {
            self.outage_lock().take();
            self.update_status(|s| {
                s.cdp_connected = true;
                s.reconnecting = false;
                s.running = true;
                s.chat_targets = 1;
            })
            .await;
        }
        Ok(reused)
    }

    async fn open_session(
        &self,
        target: &CdpTarget,
        enable_background_operation: bool,
    ) -> RelayResult<Arc<CdpSession>> {
        let session =
            CdpSession::connect_with_debug(&target.ws_url, &target.id, self.cdp_debug.clone())
                .await?;
        match session
            .send("Network.enable", network_enable_params())
            .await
        {
            Ok(_) => {}
            Err(CdpError::Protocol { code, message }) => {
                tracing::warn!(code, message = %message, target_id = %target.id, "Network.enable 被拒绝,忽略");
            }
            Err(e) => {
                session.close().await;
                return Err(e.into());
            }
        }
        if enable_background_operation {
            self.enable_background_page_operation(&session).await;
        }
        Ok(session)
    }

    /// 豆包的聊天页面在后台时仍需保持活动和焦点;这避免了用 `Page.bringToFront` 抢占用户窗口。
    /// 不同豆包版本可能不支持这两个可选 CDP 命令,此时保留会话并让后续页面自检给出具体失败原因。
    async fn enable_background_page_operation(&self, session: &CdpSession) {
        for (method, params) in [
            ("Page.setWebLifecycleState", json!({ "state": "active" })),
            (
                "Emulation.setFocusEmulationEnabled",
                json!({ "enabled": true }),
            ),
        ] {
            if let Err(error) = session.send(method, params).await {
                tracing::warn!(%method, %error, target_id = %session.target_id, "豆包后台页面准备未生效");
            }
        }
    }

    async fn close_sessions(&self) {
        let (chat, background) = {
            let mut s = self.sessions.write().await;
            (s.chat.take(), s.background.take())
        };
        if let Some(s) = chat {
            s.close().await;
        }
        if let Some(s) = background {
            s.close().await;
        }
    }

    /// 只在 `sessions.chat` 仍是 `stale` 这一条时才关掉(连同 background);
    /// 别处(心跳 / 切换验证)已经换上新 session 的话不动它,避免互相关掉对方刚建好的连接。
    async fn close_sessions_if_chat_is(&self, stale: &Arc<CdpSession>) {
        let taken = {
            let mut s = self.sessions.write().await;
            if s.chat.as_ref().is_some_and(|c| Arc::ptr_eq(c, stale)) {
                Some((s.chat.take(), s.background.take()))
            } else {
                None
            }
        };
        match taken {
            Some((chat, background)) => {
                if let Some(s) = chat {
                    s.close().await;
                }
                if let Some(s) = background {
                    s.close().await;
                }
            }
            None => stale.close().await,
        }
    }

    /// 当前 chat session(不重连)。
    async fn current_chat(&self) -> RelayResult<Arc<CdpSession>> {
        self.sessions
            .read()
            .await
            .chat
            .clone()
            .filter(|s| !s.is_closed())
            .ok_or_else(RelayError::doubao_cdp_disconnected)
    }

    /// chat session;没有或已关闭 → 尝试一次 `connect_sessions`,仍失败 → `doubao_cdp_disconnected`。
    pub async fn chat_session(&self) -> RelayResult<Arc<CdpSession>> {
        if let Ok(s) = self.current_chat().await {
            return Ok(s);
        }
        tracing::info!("chat session 不可用,尝试重连");
        if let Err(e) = self.connect_sessions_inner(true).await {
            tracing::warn!(code = e.code, error = %e, "重连 chat session 失败");
            self.update_status(|s| s.cdp_connected = false).await;
            return Err(RelayError::doubao_cdp_disconnected()
                .with_details(json!({ "cause": e.code, "cause_message": e.message })));
        }
        self.current_chat().await
    }

    /// 捕获中途 chat session 断开时用(Task 6.1):反复 [`chat_session`](Self::chat_session),
    /// 每 `reconnect_delay` 一次,直到重连成功;豆包进程不在 → 立刻返回最后的错误;
    /// 自调用起超过 `restart_after` 仍失败 → 返回最后的错误;`cancel` 触发 → 立即返回错误
    /// (调用方自己判断是否已取消)。豆包 reload / 切账号时 chat page 会短暂消失,
    /// 单次 `chat_session` 失败不应让正在捕获的任务直接以 `doubao_cdp_disconnected` 结束。
    pub async fn await_chat_session(
        &self,
        cancel: &CancellationToken,
    ) -> RelayResult<Arc<CdpSession>> {
        let started = Instant::now();
        let mut attempts = 0u32;
        loop {
            attempts += 1;
            let err = match self.chat_session().await {
                Ok(s) => {
                    if attempts > 1 {
                        tracing::info!(
                            attempts,
                            waited_ms = started.elapsed().as_millis() as u64,
                            "chat session 已重连,继续"
                        );
                    }
                    return Ok(s);
                }
                Err(e) => e,
            };
            if cancel.is_cancelled()
                || started.elapsed() >= self.timing.restart_after
                || !self.platform.is_doubao_running().await
            {
                return Err(err);
            }
            tokio::select! {
                _ = cancel.cancelled() => return Err(err),
                _ = tokio::time::sleep(self.timing.reconnect_delay) => {}
            }
        }
    }

    /// 提交前取 chat session(spec 08"任务冲突检测",Task 4.2):先 `/json/list` 数 doubao-chat page,
    /// 多于 1 个 → `generation_user_interrupted{interruption:"multiple_chat_pages", chat_targets}`,
    /// **不填 prompt、不发送**(否则每个任务都会先耗一次配额再被捕获阶段判失败),并把
    /// `status.chat_targets` 更新为实际数量;恰好 1 个 → [`chat_session`](Self::chat_session);
    /// 0 个或 `/json/list` 失败 → 交给 `chat_session` 的复用 / 重连 / 断线路径。
    pub async fn submit_session(&self) -> RelayResult<Arc<CdpSession>> {
        match list_targets(&self.host, self.port).await {
            Ok(targets) => {
                let n = select_pages(&targets, self.port).0.len();
                if n > 1 {
                    tracing::warn!(chat_targets = n, "doubao-chat page 多于 1 个,拒绝提交");
                    self.update_status(|s| s.chat_targets = n).await;
                    return Err(
                        super::capture::user_interrupted_error("multiple_chat_pages").with_details(
                            json!({
                                "interruption": "multiple_chat_pages",
                                "chat_targets": n,
                                "during": "preparing",
                            }),
                        ),
                    );
                }
                if n == 1 {
                    self.update_status(|s| s.chat_targets = 1).await;
                }
            }
            Err(e) => {
                tracing::debug!(error = %e, "提交前查询 /json/list 失败,交给 chat_session 处理");
            }
        }
        self.chat_session().await
    }

    /// 后台页 session(可能没有)。
    pub async fn background_session(&self) -> Option<Arc<CdpSession>> {
        self.sessions
            .read()
            .await
            .background
            .clone()
            .filter(|s| !s.is_closed())
    }

    // ------------------------------------------------------------------
    // 心跳
    // ------------------------------------------------------------------

    /// 对当前 chat session 发一次心跳命令。`Err(session)` 带回失败的那条 session(没有则 `None`),
    /// 让调用方只关掉它而不是 `sessions` 里任何一条。
    async fn probe_chat(&self) -> Result<(), Option<Arc<CdpSession>>> {
        let chat = self.sessions.read().await.chat.clone();
        match chat {
            Some(s) if !s.is_closed() => {
                let ok = s
                    .send_timeout(
                        "Target.getTargets",
                        json!({}),
                        self.timing.heartbeat_timeout,
                    )
                    .await
                    .is_ok();
                if ok {
                    self.last_seen_at.store(now_ms(), Ordering::SeqCst);
                    Ok(())
                } else {
                    Err(Some(s))
                }
            }
            other => Err(other),
        }
    }

    /// 一次心跳。成功 → true;失败 → 标记断开、关老 socket、查豆包进程(`running` 立刻更新);
    /// 进程还在 → 重连并复测(本次断线的第一次先等 `reconnect_delay`);断线满 `restart_after`
    /// 仍失败 → 升级 `connect(true)` 重启豆包(每次断线一次,成功即返回 true);自断开起超过 `crash_after` 仍失败且豆包进程不在 → 发 `Crashed`
    /// (每次断线只发一次)。从未连接过时直接返回 false,不做任何事。
    pub async fn heartbeat(&self) -> bool {
        if !self.ever_connected.load(Ordering::SeqCst) || self.closing.is_cancelled() {
            return false;
        }
        if self.page_busy() {
            return self.current_chat().await.is_ok();
        }
        let stale = match self.probe_chat().await {
            Ok(()) => {
                self.outage_lock().take();
                self.update_status(|s| {
                    s.cdp_connected = true;
                    s.reconnecting = false;
                    s.running = true;
                })
                .await;
                return true;
            }
            Err(stale) => stale,
        };
        let (since, fresh) = {
            let mut o = self.outage_lock();
            let fresh = o.is_none();
            let since = o
                .get_or_insert_with(|| Outage {
                    since: Instant::now(),
                    crashed_emitted: false,
                    restart_attempted: false,
                    hold: false,
                })
                .since;
            (since, fresh)
        };
        tracing::warn!(
            outage_ms = since.elapsed().as_millis() as u64,
            "心跳失败,标记豆包连接断开"
        );
        match stale {
            Some(s) => self.close_sessions_if_chat_is(&s).await,
            None => self.close_sessions().await,
        }
        // 每轮都查进程:进程退出后下一次心跳 `running` 就翻成 false(UC-06 要求 15s 内显示"豆包已退出"),
        // `Crashed` 事件仍按 `crash_after` 延后。进程不在时 `/json/list` 必然连不上,不浪费一次重连。
        let running = self.platform.is_doubao_running().await;
        self.update_status(|s| {
            s.cdp_connected = false;
            s.running = running;
            s.reconnecting = running;
        })
        .await;
        let mut hold = false;
        if running {
            // spec 09:关老 socket 后等 `reconnect_delay` 再握手(只在本次断线的第一次;
            // 之后的重试由 `spawn_heartbeat` 按同一间隔调度)。
            if fresh {
                tokio::time::sleep(self.timing.reconnect_delay).await;
            }
            match self.connect_sessions_inner(false).await {
                Ok(_) => {
                    if let Err(stale) = self.probe_chat().await {
                        tracing::warn!("重连后心跳仍无响应");
                        match stale {
                            Some(s) => self.close_sessions_if_chat_is(&s).await,
                            None => self.close_sessions().await,
                        }
                    } else {
                        self.outage_lock().take();
                        self.update_status(|s| {
                            s.cdp_connected = true;
                            s.reconnecting = false;
                            s.running = true;
                            s.chat_targets = 1;
                        })
                        .await;
                        tracing::info!("心跳重连成功");
                        return true;
                    }
                }
                Err(e) => {
                    hold = e.code == "doubao_cdp_port_occupied"
                        || e.details
                            .as_ref()
                            .is_some_and(|d| d["reason"] == "multiple_chat_pages");
                    tracing::debug!(code = e.code, error = %e, "心跳重连失败");
                }
            }
        } else {
            tracing::warn!(
                outage_ms = since.elapsed().as_millis() as u64,
                "豆包进程不在,等待其重新启动"
            );
        }
        // spec 09:进程还在但 `restart_after` 内仍未重连成功 → 升级为重启豆包(每次断线只试一次)。
        // 豆包进程已退出不自动拉起(UC-06"豆包已退出,等待重启",由用户点"重连豆包");
        // 多开了 chat 窗口是用户干扰(CDP 本身是通的),杀豆包会毁掉用户的窗口,也不升级;
        // 9707 被非豆包进程占用时重启豆包也拿不到端口,同样不升级。
        if let Some(o) = self.outage_lock().as_mut() {
            o.hold = hold;
        }
        if running && !hold && since.elapsed() >= self.timing.restart_after {
            let attempt = {
                let mut o = self.outage_lock();
                match o.as_mut() {
                    Some(out) if !out.restart_attempted => {
                        out.restart_attempted = true;
                        true
                    }
                    _ => false,
                }
            };
            if attempt {
                tracing::warn!(
                    outage_ms = since.elapsed().as_millis() as u64,
                    "重连超时仍未成功,升级为重启豆包"
                );
                match self.connect(true).await {
                    Ok(action) => {
                        tracing::info!(action = ?action, "重启豆包后已重新连接");
                        return true;
                    }
                    Err(e) => {
                        tracing::warn!(code = e.code, error = %e, "重启豆包失败,回到常规心跳");
                    }
                }
            }
        }
        if !running && since.elapsed() >= self.timing.crash_after {
            let emit = {
                let mut o = self.outage_lock();
                match o.as_mut() {
                    Some(out) if !out.crashed_emitted => {
                        out.crashed_emitted = true;
                        true
                    }
                    _ => false,
                }
            };
            if emit {
                tracing::error!(
                    outage_ms = since.elapsed().as_millis() as u64,
                    "豆包进程已退出,判定崩溃"
                );
                let _ = self.events.send(ConnEvent::Crashed);
            }
        }
        false
    }

    /// 是否处于"快速重试"阶段:断线中、豆包进程还在、本次断线还没升级重启过,且最近一次失败
    /// 不是快速重试无济于事的原因(多开 chat 窗口 / 端口被非豆包占用,见 `Outage::hold`)。
    /// 后台心跳据此每 `reconnect_delay` 重试一次,而不是等下一个常规周期。
    pub async fn retry_soon(&self) -> bool {
        let reconnecting = {
            let st = self.status.read().await;
            st.reconnecting && !st.cdp_connected
        };
        reconnecting
            && self
                .outage_lock()
                .as_ref()
                .is_some_and(|o| !o.restart_attempted && !o.hold)
    }

    /// 每 `interval` 跑一次 [`heartbeat`](Self::heartbeat)(spec:10s);断线且豆包仍在时
    /// 改为每 `reconnect_delay` 重试,直到重连成功、升级重启(spec 09 的"20 秒内未成功"),
    /// 或遇到快速重试无济于事的失败原因(见 [`retry_soon`](Self::retry_soon))。
    pub fn spawn_heartbeat(self: &Arc<Self>, interval: Duration) -> JoinHandle<()> {
        let this = Arc::clone(self);
        tokio::spawn(async move {
            let mut ticker = tokio::time::interval(interval);
            ticker.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
            let closing = this.closing.clone();
            closing
                .run_until_cancelled(async {
                    ticker.tick().await;
                    loop {
                        if this.retry_soon().await {
                            tokio::time::sleep(
                                this.timing.reconnect_delay.max(Duration::from_millis(10)),
                            )
                            .await;
                            ticker.reset();
                        } else {
                            ticker.tick().await;
                        }
                        this.heartbeat().await;
                    }
                })
                .await;
            tracing::debug!("心跳循环已停止(应用关闭)");
        })
    }

    /// 应用关闭流程(spec 03)用:停心跳、拒绝之后的(重)连,关闭现有 CDP 会话(只断开
    /// WebSocket,**不杀豆包**)。幂等。
    pub async fn shutdown(&self) {
        self.closing.cancel();
        let (chat, background) = {
            let mut sessions = self.sessions.write().await;
            (sessions.chat.take(), sessions.background.take())
        };
        for session in [chat, background].into_iter().flatten() {
            session.close().await;
        }
        self.update_status(|s| {
            s.cdp_connected = false;
            s.reconnecting = false;
        })
        .await;
    }

    pub fn is_shutting_down(&self) -> bool {
        self.closing.is_cancelled()
    }

    // ------------------------------------------------------------------
    // 账号
    // ------------------------------------------------------------------

    /// 标记"任务正在使用 chat 页面"(捕获循环全程持有;提交经 [`lock_page`](Self::lock_page) 持有)。
    /// guard 存活期间 [`accounts(false)`](Self::accounts) 有缓存(哪怕已过期)就直接用,不开账号菜单。
    pub fn mark_page_busy(&self) -> PageBusyGuard<'_> {
        self.page_busy.fetch_add(1, Ordering::SeqCst);
        PageBusyGuard(&self.page_busy)
    }

    /// 是否有任务正在使用 chat 页面。
    pub fn page_busy(&self) -> bool {
        self.page_busy.load(Ordering::SeqCst) > 0
    }

    /// 任务提交(`generation::submit`)期间独占 chat 页面:先标记忙(等锁期间的账号读取改用缓存),
    /// 再取 `switch_lock`,与账号菜单读取 / 切换互斥。
    pub async fn lock_page(&self) -> PageGuard<'_> {
        let busy = self.mark_page_busy();
        let lock = self.switch_lock.lock().await;
        PageGuard {
            _lock: lock,
            _busy: busy,
        }
    }

    /// 账号列表(2s 缓存,`force` 跳过);读取后发 `AccountListChanged`。
    /// 缓存命中不等 `switch_lock`(切换进行中 `GET /accounts` / `submit` 校验仍能秒回),
    /// 未命中才持锁操作菜单,锁内再查一次缓存。
    ///
    /// 非强制读取时,若有任务正在使用 chat 页面([`page_busy`](Self::page_busy)),只要曾经读到过列表
    /// 就返回那份(可能过期的)缓存,不打开账号菜单:菜单读取约 3.4s,会抢走焦点并以 Escape 收尾,
    /// 与任务的菜单 / 浮层 / 键盘操作交错会让任务失败。从未读到过列表时只能读菜单,但仍需等
    /// 正在进行的提交释放页面锁。
    pub async fn accounts(&self, force: bool) -> RelayResult<AccountList> {
        if let Some(list) = self.usable_cache(force).await {
            return Ok(list);
        }
        let _guard = self.switch_lock.lock().await;
        self.accounts_locked(force).await
    }

    /// 只读页面已显示的当前账号，不打开账号菜单。
    pub async fn current_account(&self) -> RelayResult<Option<String>> {
        let session = self.chat_session().await?;
        let current = probe_current(&session).await?;
        Ok(
            (current.status == "logged_in" && !current.name.trim().is_empty())
                .then(|| current.name.trim().to_string()),
        )
    }

    async fn cached_accounts(&self) -> Option<AccountList> {
        let cache = self.account_cache.read().await;
        let (list, at) = cache.as_ref()?;
        (at.elapsed() < self.timing.account_cache_ttl).then(|| list.clone())
    }

    /// 非强制读取可用的缓存:未过期的;或页面忙时任意一份(过期也用)。
    async fn usable_cache(&self, force: bool) -> Option<AccountList> {
        if force {
            return None;
        }
        if let Some(list) = self.cached_accounts().await {
            return Some(list);
        }
        if self.page_busy() {
            let cache = self.account_cache.read().await;
            if let Some((list, at)) = cache.as_ref() {
                tracing::debug!(
                    age_ms = at.elapsed().as_millis() as u64,
                    "任务正在使用豆包页面,账号列表沿用缓存,不打开账号菜单"
                );
                return Some(list.clone());
            }
        }
        None
    }

    async fn accounts_locked(&self, force: bool) -> RelayResult<AccountList> {
        if let Some(list) = self.usable_cache(force).await {
            return Ok(list);
        }
        let session = self.chat_session().await?;
        let list = match read_accounts_with(&session, self.timing.read).await {
            Ok(list) => list,
            Err(error) => {
                // Windows 豆包新版偶发无法展开账号菜单，但页面仍保持已登录状态。
                // 兜底必须重新读取当前页面显示的账号，不能写死某个历史账号名。
                // 否则换账号后会把任务发给不存在的账号，表现为 account_not_found。
                match probe_current(&session).await {
                    Ok(current)
                        if current.status == "logged_in" && !current.name.trim().is_empty() =>
                    {
                        let name = current.name.trim().to_string();
                        tracing::warn!(error = %error, current = %name, "账号菜单不可读，沿用当前页面账号");
                        AccountList {
                            current: Some(name.clone()),
                            accounts: vec![Account {
                                name: name.clone(),
                                subscription: (!current.subscription.trim().is_empty())
                                    .then(|| current.subscription.clone()),
                                label: name,
                                is_current: true,
                                is_ambiguous: false,
                            }],
                            read_at: now_ms(),
                        }
                    }
                    Ok(current) => {
                        tracing::warn!(
                            error = %error,
                            status = %current.status,
                            "账号菜单不可读且当前页面未登录"
                        );
                        return Err(error);
                    }
                    Err(probe_error) => {
                        tracing::warn!(
                            error = %error,
                            probe_error = %probe_error,
                            "账号菜单不可读且无法读取当前页面账号"
                        );
                        return Err(error);
                    }
                }
            }
        };
        *self.account_cache.write().await = Some((list.clone(), Instant::now()));
        let _ = self
            .events
            .send(ConnEvent::AccountListChanged(list.clone()));
        Ok(list)
    }

    /// 切换账号(F2.3)。流程见模块文档。
    pub async fn switch_account(&self, target: &str) -> RelayResult<SwitchResult> {
        let _guard = self.switch_lock.lock().await;
        let started = Instant::now();
        let list = self.accounts_locked(true).await?;
        let matches: Vec<_> = list.accounts.iter().filter(|a| a.name == target).collect();
        let Some(account) = matches.first() else {
            return Err(RelayError::account_not_found(
                target,
                list.accounts.iter().map(|a| a.name.clone()).collect(),
            ));
        };
        if matches.len() > 1 || account.is_ambiguous {
            return Err(RelayError::account_ambiguous(
                target,
                matches.iter().map(|a| a.label.clone()).collect(),
            ));
        }
        let previous = list.current.clone();
        // 单账号菜单是比旧的 current() probe 更可靠的事实来源。
        // 豆包新版可能返回过期的 probe 名称，导致唯一账号被错误地再次切换。
        if account.is_current || list.accounts.len() == 1 {
            tracing::info!(target, "目标账号已是当前账号");
            return Ok(SwitchResult {
                ok: true,
                previous,
                current: target.to_string(),
                switched_at: now_ms(),
            });
        }

        let session = self.chat_session().await?;
        click_account(&session, target, self.timing.switch).await?;
        self.verify_switch(session, target, started).await?;

        let switched_at = now_ms();
        *self.account_cache.write().await = None;
        tracing::info!(
            target,
            previous = previous.as_deref().unwrap_or(""),
            elapsed_ms = started.elapsed().as_millis() as u64,
            "账号切换完成"
        );
        let _ = self.events.send(ConnEvent::AccountSwitched {
            previous: previous.clone(),
            current: target.to_string(),
            switched_at,
        });
        if let Err(e) = self.accounts_locked(true).await {
            tracing::warn!(code = e.code, error = %e, "切换后重新读取账号列表失败");
        }
        Ok(SwitchResult {
            ok: true,
            previous,
            current: target.to_string(),
            switched_at,
        })
    }

    /// 验证循环:每 `verify_poll` 读一次 `current()`;session 断开(页面 reload)→ 重连后继续;
    /// 出现登录弹窗 target → `account_switch_failed(login_required)`;超时 → `account_switch_timeout`。
    async fn verify_switch(
        &self,
        mut session: Arc<CdpSession>,
        target: &str,
        started: Instant,
    ) -> RelayResult<()> {
        let timing = self.timing.switch;
        let deadline = Instant::now() + timing.verify_timeout;
        let mut reconnects = 0u32;
        loop {
            tokio::time::sleep(timing.verify_poll).await;
            if Instant::now() >= deadline {
                break;
            }
            // 每一步都以 deadline 为界:`Runtime.evaluate` 默认 30s、`/json/list` 3s,
            // 单次挂起不能把 20s 的验证窗口拖长。
            let remaining = deadline.saturating_duration_since(Instant::now());
            if let Ok(Ok(targets)) =
                tokio::time::timeout(remaining, list_targets(&self.host, self.port)).await
            {
                if let Some(modal) = targets.iter().find(|t| is_login_modal_target(t)) {
                    tracing::warn!(target, url = %modal.url, "切换时豆包弹出登录窗口,目标账号登录态已失效");
                    self.try_close_menu().await;
                    return Err(
                        RelayError::account_switch_failed("login_required").with_details(json!({
                            "reason": "login_required",
                            "target": target,
                            "login_modal_url": modal.url,
                        })),
                    );
                }
            }
            let remaining = deadline.saturating_duration_since(Instant::now());
            let probed = match tokio::time::timeout(remaining, probe_current(&session)).await {
                Ok(r) => r,
                Err(_) => {
                    tracing::debug!("切换验证读取未在 deadline 前返回");
                    continue;
                }
            };
            match probed {
                Ok(c) if c.status == "logged_in" && c.name.trim() == target => {
                    tracing::info!(
                        target,
                        reconnects,
                        elapsed_ms = started.elapsed().as_millis() as u64,
                        "切换验证通过"
                    );
                    return Ok(());
                }
                Ok(c) => {
                    tracing::debug!(status = %c.status, name = %c.name, "切换验证中");
                }
                Err(CdpError::Closed) | Err(CdpError::Timeout(_)) | Err(CdpError::Connect(_)) => {
                    tracing::info!(reconnects, "切换期间 chat session 失效,重连");
                    reconnects += 1;
                    if !session.is_closed() {
                        // 超时但未断:只放弃自己手里这条 session(心跳可能已换上新的),重连后再试
                        self.close_sessions_if_chat_is(&session).await;
                    }
                    if let Ok(s) = self.chat_session().await {
                        session = s;
                    }
                }
                Err(e) => {
                    tracing::debug!(error = %e, "切换验证读取失败,继续等待");
                }
            }
        }
        let elapsed_ms = started.elapsed().as_millis() as u64;
        tracing::warn!(target, elapsed_ms, reconnects, "账号切换验证超时");
        self.try_close_menu().await;
        Err(RelayError::account_switch_timeout(elapsed_ms))
    }

    /// 切换失败/超时后尽量把菜单关上(session 不可用就算了)。
    async fn try_close_menu(&self) {
        if let Ok(s) = self.current_chat().await {
            let _ = s
                .evaluate(&super::accounts::injected_call("closeMenu"))
                .await;
        }
    }
}

/// 豆包弹出的登录窗口 target(`https://www.doubao.com/...client-login-modal...`,type=page)。
/// 切换到登录态失效的账号时出现(Task 1.2 实测),不得向它建 session。
pub fn is_login_modal_target(t: &CdpTarget) -> bool {
    if t.kind != "page" {
        return false;
    }
    let Ok(url) = Url::parse(&t.url) else {
        return false;
    };
    matches!(url.host_str(), Some("www.doubao.com") | Some("doubao.com"))
        && url.path().contains("client-login-modal")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::{Path, PathBuf};
    use std::sync::atomic::AtomicUsize;

    use async_trait::async_trait;
    use futures_util::{SinkExt, StreamExt};
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::net::TcpListener;
    use tokio_tungstenite::tungstenite::Message;

    use crate::doubao::ACCOUNTS_SCRIPT_VERSION;
    use crate::platform::PlatformError;

    // ---- 假平台 ----

    struct FakePlatform {
        running: AtomicBool,
        kill_calls: AtomicUsize,
        launch_calls: AtomicUsize,
        /// 为 true 时 `launch_doubao_with_cdp` 报错(模拟重启失败,不必等 20s 的 `wait_for_cdp`)。
        launch_fails: AtomicBool,
        /// `listening_process_exe` 的返回值(默认豆包自带的 Doubao Browser)。
        owner: StdMutex<Option<PathBuf>>,
    }

    const DOUBAO_BROWSER: &str =
        "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser";

    impl FakePlatform {
        fn new(running: bool) -> Arc<Self> {
            Arc::new(Self {
                running: AtomicBool::new(running),
                kill_calls: AtomicUsize::new(0),
                launch_calls: AtomicUsize::new(0),
                launch_fails: AtomicBool::new(false),
                owner: StdMutex::new(Some(PathBuf::from(DOUBAO_BROWSER))),
            })
        }

        fn set_owner(&self, owner: Option<&str>) {
            *self.owner.lock().unwrap() = owner.map(PathBuf::from);
        }
    }

    #[async_trait]
    impl Platform for FakePlatform {
        async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
            Ok(PathBuf::from(
                "/Applications/Doubao.app/Contents/MacOS/Doubao",
            ))
        }
        async fn launch_doubao_with_cdp(&self, _: &Path, _: u16) -> Result<(), PlatformError> {
            self.launch_calls.fetch_add(1, Ordering::SeqCst);
            if self.launch_fails.load(Ordering::SeqCst) {
                return Err(PlatformError::LaunchFailed("mock".into()));
            }
            Ok(())
        }
        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            self.kill_calls.fetch_add(1, Ordering::SeqCst);
            Ok(())
        }
        async fn is_doubao_running(&self) -> bool {
            self.running.load(Ordering::SeqCst)
        }
        fn app_data_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay")
        }
        fn log_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/logs")
        }
        fn config_path(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/config.toml")
        }
        async fn reveal_in_file_manager(&self, _: &Path) -> Result<(), PlatformError> {
            Ok(())
        }
        fn platform_name(&self) -> &'static str {
            "fake"
        }
        fn listening_process_exe(&self, _: u16) -> Option<PathBuf> {
            self.owner.lock().unwrap().clone()
        }
        fn is_doubao_exe_path(&self, p: &Path) -> bool {
            p.components()
                .any(|c| c.as_os_str().to_string_lossy() == "Doubao.app")
        }
        fn doubao_version(&self) -> Option<String> {
            Some("2.32.5".into())
        }
    }

    // ---- mock 豆包:同一端口服务 /json/* 与 WebSocket page ----

    type Handler = Arc<StdMutex<Box<dyn FnMut(&str) -> Value + Send>>>;
    /// `switchTo` 返回 `clicked` 时触发的钩子(模拟豆包 reload / 弹登录窗)。
    type ClickHook = Arc<StdMutex<Option<Box<dyn Fn() + Send>>>>;

    struct MockDoubao {
        port: u16,
        targets: Arc<StdMutex<Vec<Value>>>,
        methods: Arc<StdMutex<Vec<String>>>,
        hang_targets: Arc<AtomicBool>,
        /// 为 true 时 `current()` 的 `Runtime.evaluate` 永不响应(模拟页面卡住)。
        hang_current: Arc<AtomicBool>,
        close_tx: broadcast::Sender<()>,
        connections: Arc<AtomicUsize>,
    }

    impl MockDoubao {
        async fn start(handler: Handler) -> Self {
            let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
            let port = listener.local_addr().unwrap().port();
            let targets: Arc<StdMutex<Vec<Value>>> = Arc::new(StdMutex::new(Vec::new()));
            let methods: Arc<StdMutex<Vec<String>>> = Arc::new(StdMutex::new(Vec::new()));
            let hang_targets = Arc::new(AtomicBool::new(false));
            let hang_current = Arc::new(AtomicBool::new(false));
            let (close_tx, _) = broadcast::channel::<()>(16);
            let connections = Arc::new(AtomicUsize::new(0));
            let mock = Self {
                port,
                targets: Arc::clone(&targets),
                methods: Arc::clone(&methods),
                hang_targets: Arc::clone(&hang_targets),
                hang_current: Arc::clone(&hang_current),
                close_tx: close_tx.clone(),
                connections: Arc::clone(&connections),
            };
            tokio::spawn(async move {
                loop {
                    let Ok((sock, _)) = listener.accept().await else {
                        break;
                    };
                    let targets = Arc::clone(&targets);
                    let methods = Arc::clone(&methods);
                    let handler = Arc::clone(&handler);
                    let hang = Arc::clone(&hang_targets);
                    let hang_current = Arc::clone(&hang_current);
                    let mut close_rx = close_tx.subscribe();
                    let connections = Arc::clone(&connections);
                    tokio::spawn(async move {
                        let mut peek = [0u8; 16];
                        let n = sock.peek(&mut peek).await.unwrap_or(0);
                        if String::from_utf8_lossy(&peek[..n]).starts_with("GET /json") {
                            serve_http(sock, &targets).await;
                            return;
                        }
                        let Ok(mut ws) = tokio_tungstenite::accept_async(sock).await else {
                            return;
                        };
                        connections.fetch_add(1, Ordering::SeqCst);
                        loop {
                            tokio::select! {
                                msg = ws.next() => {
                                    let Some(Ok(Message::Text(txt))) = msg else { break };
                                    let req: Value = serde_json::from_str(&txt).unwrap();
                                    let method = req["method"].as_str().unwrap_or("");
                                    methods.lock().unwrap().push(method.to_string());
                                    let result = match method {
                                        "Runtime.evaluate" => {
                                            let expr = req["params"]["expression"].as_str().unwrap_or("");
                                            if hang_current.load(Ordering::SeqCst) && expr.contains(".current()") { continue; }
                                            let value = (handler.lock().unwrap())(expr);
                                            json!({ "result": { "type": "object", "value": value } })
                                        }
                                        "Target.getTargets" => {
                                            if hang.load(Ordering::SeqCst) { continue; }
                                            json!({ "targetInfos": [] })
                                        }
                                        _ => json!({}),
                                    };
                                    let resp = json!({ "id": req["id"], "result": result });
                                    if ws.send(Message::Text(resp.to_string().into())).await.is_err() { break; }
                                }
                                _ = close_rx.recv() => {
                                    let _ = ws.close(None).await;
                                    break;
                                }
                            }
                        }
                    });
                }
            });
            mock
        }

        fn chat_target(&self, id: &str) -> Value {
            json!({
                "id": id, "type": "page", "url": "doubao://doubao-chat/chat",
                "webSocketDebuggerUrl": format!("ws://127.0.0.1:{}/devtools/page/{id}", self.port)
            })
        }

        fn background_target(&self, id: &str) -> Value {
            json!({
                "id": id, "type": "page", "url": "doubao://doubao-background/",
                "webSocketDebuggerUrl": format!("ws://127.0.0.1:{}/devtools/page/{id}", self.port)
            })
        }

        fn login_modal_target(&self) -> Value {
            json!({
                "id": "LOGIN", "type": "page",
                "url": "https://www.doubao.com/flow-account/client-login/client-login-modal?enterMethod=add_account",
                "webSocketDebuggerUrl": format!("ws://127.0.0.1:{}/devtools/page/LOGIN", self.port)
            })
        }

        fn set_targets(&self, v: Vec<Value>) {
            *self.targets.lock().unwrap() = v;
        }

        fn methods(&self) -> Vec<String> {
            self.methods.lock().unwrap().clone()
        }

        /// 模拟页面 reload:服务端关掉所有现有 WebSocket。
        fn close_all(&self) {
            let _ = self.close_tx.send(());
        }
    }

    async fn serve_http(mut sock: tokio::net::TcpStream, targets: &StdMutex<Vec<Value>>) {
        let mut buf = vec![0u8; 4096];
        let n = sock.read(&mut buf).await.unwrap_or(0);
        let req = String::from_utf8_lossy(&buf[..n]).to_string();
        let body = if req.starts_with("GET /json/version") {
            r#"{"Browser":"Chrome/147.0.0.0","Protocol-Version":"1.3"}"#.to_string()
        } else {
            Value::Array(targets.lock().unwrap().clone()).to_string()
        };
        let resp = format!(
            "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
            body.len(),
            body
        );
        let _ = sock.write_all(resp.as_bytes()).await;
        let _ = sock.shutdown().await;
    }

    fn fast_timing() -> ConnTiming {
        let read = ReadTiming {
            after_expand: Duration::from_millis(2),
            after_open: Duration::from_millis(2),
            after_hover: Duration::from_millis(2),
            per_attempt: Duration::from_millis(2),
            max_attempts: 8,
            after_close: Duration::from_millis(2),
        };
        ConnTiming {
            heartbeat_timeout: Duration::from_millis(150),
            crash_after: Duration::from_secs(20),
            account_cache_ttl: Duration::from_secs(2),
            read,
            switch: SwitchTiming {
                menu_poll: Duration::from_millis(2),
                menu_max_attempts: 5,
                verify_poll: Duration::from_millis(20),
                verify_timeout: Duration::from_secs(5),
                read,
            },
            page_ready_timeout: Duration::from_millis(300),
            page_ready_poll: Duration::from_millis(10),
            reconnect_delay: Duration::from_millis(10),
            restart_after: Duration::from_secs(20),
        }
    }

    fn conn(
        mock: &MockDoubao,
        platform: Arc<FakePlatform>,
        timing: ConnTiming,
    ) -> Arc<DoubaoConnection> {
        DoubaoConnection::new_with(platform, "127.0.0.1".into(), mock.port, None, timing)
    }

    /// 页面状态:账号 A(当前)/ B;`phase` 翻到 1 后当前账号变成 B。
    struct PageState {
        phase: Arc<AtomicUsize>,
        reads: Arc<AtomicUsize>,
        switch_action: Arc<StdMutex<String>>,
        on_click: ClickHook,
        /// 页面就绪探针前 N 次返回 false。
        not_ready_polls: Arc<AtomicUsize>,
        /// 页面就绪探针被调用的次数。
        ready_polls: Arc<AtomicUsize>,
    }

    fn page_handler() -> (Handler, PageState) {
        let state = PageState {
            phase: Arc::new(AtomicUsize::new(0)),
            reads: Arc::new(AtomicUsize::new(0)),
            switch_action: Arc::new(StdMutex::new("clicked".into())),
            on_click: Arc::new(StdMutex::new(None)),
            not_ready_polls: Arc::new(AtomicUsize::new(0)),
            ready_polls: Arc::new(AtomicUsize::new(0)),
        };
        let phase = Arc::clone(&state.phase);
        let reads = Arc::clone(&state.reads);
        let not_ready_polls = Arc::clone(&state.not_ready_polls);
        let ready_polls = Arc::clone(&state.ready_polls);
        let switch_action = Arc::clone(&state.switch_action);
        let on_click = Arc::clone(&state.on_click);
        let handler: Handler = Arc::new(StdMutex::new(Box::new(move |expr: &str| {
            let current_name = if phase.load(Ordering::SeqCst) == 0 {
                "甲"
            } else {
                "乙"
            };
            let api = "window.__doubaoRelayAccounts.";
            if expr.contains(&format!("{api}current()")) {
                return json!({"status": "logged_in", "name": current_name, "subscription": "体验套餐", "raw": format!("{current_name}\n体验套餐")});
            }
            if expr.starts_with(&format!("{api}switchTo(")) {
                let action = switch_action.lock().unwrap().clone();
                if action == "clicked" {
                    if let Some(f) = on_click.lock().unwrap().as_ref() {
                        f();
                    }
                }
                return json!({"action": action, "count": if action == "ambiguous" { 2 } else { 0 }});
            }
            if expr.contains("const VERSION") {
                return Value::String(ACCOUNTS_SCRIPT_VERSION.into());
            }
            if expr.starts_with("(async () =>") {
                return json!({"accountButtonFound": true, "inputFound": true, "sendButtonFound": true, "sendVia": "direct"});
            }
            if expr.starts_with("(() => document.readyState") {
                // 前 `not_ready_polls` 次返回未就绪(模拟刚启动的豆包页面还没渲染)
                let left = not_ready_polls.load(Ordering::SeqCst);
                ready_polls.fetch_add(1, Ordering::SeqCst);
                if left > 0 {
                    not_ready_polls.store(left - 1, Ordering::SeqCst);
                    return json!(false);
                }
                return json!(true);
            }
            match expr.strip_prefix(api).unwrap_or("") {
                "probe()" => {
                    json!({"buttonFound": true, "loginVisible": false, "hasInput": true, "haspopupCount": 28, "candidates": []})
                }
                "openMenu()" => json!({"action": "pointer_events", "expanded": "true"}),
                "hoverMore()" => json!({"action": "none", "count": 0}),
                "expandSidebar()" => json!({"action": "none", "count": 0}),
                "closeMenu()" => json!({"action": "escape", "stillOpen": false}),
                "readEntries()" => {
                    reads.fetch_add(1, Ordering::SeqCst);
                    let row = |name: &str, cur: bool| json!({"label": format!("{name}\n体验套餐"), "lines": [name, "体验套餐"], "hasAvatar": cur, "hasPopup": false, "disabled": false, "isAccountRow": true, "testId": "avatar_menu_account_row"});
                    json!({"menuItemCount": 4, "logoutVisible": true, "entries": [
                        {"label": "设置", "lines": ["设置"], "hasAvatar": false, "hasPopup": false, "disabled": false, "isAccountRow": false, "testId": ""},
                        row("甲", current_name == "甲"), row("乙", current_name == "乙"),
                        {"label": "退出登录", "lines": ["退出登录"], "hasAvatar": false, "hasPopup": false, "disabled": false, "isAccountRow": false, "testId": "setup_logout"}
                    ]})
                }
                other => panic!("unexpected expression: {other}"),
            }
        })));
        (handler, state)
    }

    async fn recv_event(rx: &mut broadcast::Receiver<ConnEvent>) -> ConnEvent {
        tokio::time::timeout(Duration::from_secs(3), rx.recv())
            .await
            .expect("等待事件超时")
            .expect("事件通道")
    }

    // ---- 纯函数 ----

    #[test]
    fn login_modal_target_detection() {
        let t = |kind: &str, url: &str| CdpTarget {
            id: "X".into(),
            kind: kind.into(),
            url: url.into(),
            title: String::new(),
            ws_url: String::new(),
        };
        assert!(is_login_modal_target(&t(
            "page",
            "https://www.doubao.com/client-login-modal?enterMethod=add_account&entry=app_main_chat"
        )));
        assert!(is_login_modal_target(&t(
            "page",
            "https://www.doubao.com/flow-account/client-login/client-login-modal?x=1"
        )));
        assert!(!is_login_modal_target(&t(
            "iframe",
            "https://www.doubao.com/client-login-modal"
        )));
        assert!(!is_login_modal_target(&t(
            "page",
            "https://www.doubao.com/drive-iframe/drive/home/"
        )));
        assert!(!is_login_modal_target(&t(
            "page",
            "doubao://doubao-chat/chat"
        )));
        assert!(!is_login_modal_target(&t(
            "page",
            "https://evil.example.com/client-login-modal"
        )));
        assert!(!is_login_modal_target(&t("page", "not a url")));
    }

    #[test]
    fn status_and_switch_result_serialize_snake_case() {
        let s = DoubaoStatus {
            running: true,
            cdp_connected: true,
            cdp_port: 9707,
            doubao_version_detected: Some("2.32.5".into()),
            ui_compat_check: "ok".into(),
            ui_compat_details: UiCompat::default(),
            chat_targets: 1,
            reconnecting: false,
            last_check_at: 7,
        };
        let v = serde_json::to_value(&s).unwrap();
        assert_eq!(v["cdp_connected"], true);
        assert_eq!(v["doubao_version_detected"], "2.32.5");
        assert_eq!(v["ui_compat_details"]["account_button_found"], false);
        assert_eq!(v["chat_targets"], 1);
        assert_eq!(v["last_check_at"], 7);
        let r = serde_json::to_value(SwitchResult {
            ok: true,
            previous: None,
            current: "乙".into(),
            switched_at: 1,
        })
        .unwrap();
        assert!(r["previous"].is_null());
        assert_eq!(r["switched_at"], 1);
    }

    #[test]
    fn default_timing_matches_spec() {
        let t = ConnTiming::default();
        assert_eq!(t.heartbeat_timeout, Duration::from_secs(5));
        assert_eq!(t.crash_after, Duration::from_secs(20));
        assert_eq!(t.account_cache_ttl, Duration::from_secs(2));
        assert_eq!(t.switch.verify_timeout, Duration::from_secs(20));
        assert_eq!(t.switch.verify_poll, Duration::from_millis(300));
        assert_eq!(t.switch.menu_poll, Duration::from_millis(100));
        assert_eq!(t.switch.menu_max_attempts, 40);
        // spec 09"断线检测与重连":关老 socket → 等 2s → 重新握手 → 20s 内未成功升级为重启豆包
        assert_eq!(t.reconnect_delay, Duration::from_secs(2));
        assert_eq!(t.restart_after, Duration::from_secs(20));
    }

    // ---- 会话 ----

    #[tokio::test]
    async fn connect_sessions_requires_exactly_one_chat_page() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.background_target("D")]);
        let err = c.connect_sessions().await.unwrap_err();
        assert_eq!(err.code, "doubao_not_ready");
        assert_eq!(err.details.as_ref().unwrap()["chat_targets"], 0);
        assert!(!c.status().await.cdp_connected);

        mock.set_targets(vec![mock.chat_target("A"), mock.chat_target("B")]);
        let err = c.connect_sessions().await.unwrap_err();
        assert_eq!(err.code, "doubao_not_ready");
        assert_eq!(err.details.as_ref().unwrap()["chat_targets"], 2);
        assert_eq!(c.status().await.chat_targets, 2);

        let err = c.chat_session().await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
        assert_eq!(err.details.unwrap()["cause"], "doubao_not_ready");
    }

    #[tokio::test]
    async fn new_chat_session_enables_background_page_operation() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);

        c.connect_sessions().await.unwrap();

        let methods = mock.methods();
        assert!(
            methods.iter().any(|m| m == "Page.setWebLifecycleState"),
            "{methods:?}"
        );
        assert!(
            methods
                .iter()
                .any(|m| m == "Emulation.setFocusEmulationEnabled"),
            "{methods:?}"
        );
        assert!(
            !methods.iter().any(|m| m == "Page.bringToFront"),
            "后台执行不能抢占用户前台窗口: {methods:?}"
        );
    }

    #[tokio::test]
    async fn chat_session_reconnects_after_mock_closes_with_new_target_id() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        let mut rx = c.subscribe();
        mock.set_targets(vec![mock.chat_target("A"), mock.background_target("D")]);
        c.connect_sessions().await.unwrap();
        let s1 = c.chat_session().await.unwrap();
        assert_eq!(s1.target_id, "A");
        assert!(c.background_session().await.is_some());
        assert_eq!(mock.connections.load(Ordering::SeqCst), 2);
        let ev = recv_event(&mut rx).await;
        assert!(
            matches!(ev, ConnEvent::StatusChanged(ref s) if s.cdp_connected && s.chat_targets == 1),
            "{ev:?}"
        );

        // 再调一次:target 未变且 session 未关 → 复用,不新建连接
        let again = c.chat_session().await.unwrap();
        assert!(Arc::ptr_eq(&s1, &again));
        c.connect_sessions().await.unwrap();
        assert_eq!(mock.connections.load(Ordering::SeqCst), 2);

        // 模拟 reload:老 socket 被服务端关掉,target id 变化
        mock.set_targets(vec![mock.chat_target("B")]);
        mock.close_all();
        tokio::time::timeout(Duration::from_secs(2), async {
            while !s1.is_closed() {
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
        })
        .await
        .expect("老 session 应被关闭");
        let s2 = c.chat_session().await.unwrap();
        assert_eq!(s2.target_id, "B");
        assert!(!s2.is_closed());
        assert!(c.background_session().await.is_none());
        assert!(c.status().await.cdp_connected);
    }

    #[tokio::test]
    async fn connect_returns_already_connected_then_reconnected() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        let mut rx = c.subscribe();
        mock.set_targets(vec![mock.chat_target("A")]);
        let action = c.connect(false).await.unwrap();
        assert_eq!(action, StartupAction::AlreadyConnected);
        let st = c.status().await;
        assert!(st.running && st.cdp_connected && !st.reconnecting);
        assert_eq!(st.doubao_version_detected.as_deref(), Some("2.32.5"));
        assert_eq!(st.ui_compat_check, "ok");
        assert!(st.ui_compat_details.send_button_found);
        assert_eq!(st.chat_targets, 1);
        assert!(st.last_check_at > 0);
        let ev = recv_event(&mut rx).await;
        assert!(
            matches!(ev, ConnEvent::StatusChanged(ref s) if s.cdp_connected),
            "{ev:?}"
        );

        // session 仍活着 → 再 connect 仍是 already_connected
        assert_eq!(
            c.connect(false).await.unwrap(),
            StartupAction::AlreadyConnected
        );
        // session 被关掉后 connect → reconnected
        let s1 = c.chat_session().await.unwrap();
        mock.close_all();
        while !s1.is_closed() {
            tokio::time::sleep(Duration::from_millis(5)).await;
        }
        assert_eq!(c.connect(false).await.unwrap(), StartupAction::Reconnected);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn connect_force_restart_kills_and_launches() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        let action = c.connect(true).await.unwrap();
        assert_eq!(action, StartupAction::Restarted);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 1);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 1);
        assert!(c.status().await.cdp_connected);
    }

    /// relay 自己拉起豆包后要等页面渲染出来再自检(Task 1.10 真机:启动后立刻自检得 error)。
    #[tokio::test]
    async fn connect_after_launch_waits_for_page_ready_before_compat_check() {
        let (handler, state) = page_handler();
        state.not_ready_polls.store(3, Ordering::SeqCst);
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        assert_eq!(c.connect(true).await.unwrap(), StartupAction::Restarted);
        assert_eq!(
            state.ready_polls.load(Ordering::SeqCst),
            4,
            "3 次未就绪 + 1 次就绪"
        );
        assert_eq!(c.status().await.ui_compat_check, "ok");

        // 已在运行(AlreadyConnected)不等页面
        state.ready_polls.store(0, Ordering::SeqCst);
        let c2 = conn(&mock, Arc::clone(&platform), fast_timing());
        assert_eq!(
            c2.connect(false).await.unwrap(),
            StartupAction::AlreadyConnected
        );
        assert_eq!(state.ready_polls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn wait_for_page_ready_times_out_without_error() {
        let (handler, state) = page_handler();
        state.not_ready_polls.store(usize::MAX, Ordering::SeqCst);
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let session = c.chat_session().await.unwrap();
        let started = Instant::now();
        let ready = wait_for_page_ready(
            &session,
            Duration::from_millis(100),
            Duration::from_millis(10),
        )
        .await;
        assert!(!ready);
        assert!(started.elapsed() >= Duration::from_millis(100));
        assert!(state.ready_polls.load(Ordering::SeqCst) >= 2);
    }

    #[tokio::test]
    async fn connect_failure_marks_disconnected() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![]);
        let err = c.connect(false).await.unwrap_err();
        assert_eq!(err.code, "doubao_not_ready");
        let st = c.status().await;
        assert!(st.running && !st.cdp_connected);
    }

    // ---- 心跳 ----

    #[tokio::test]
    async fn heartbeat_noop_before_first_connection() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        let mut rx = c.subscribe();
        assert!(!c.heartbeat().await);
        assert!(rx.try_recv().is_err());
        assert_eq!(mock.connections.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn heartbeat_does_not_probe_or_close_the_chat_session_while_page_is_busy() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let session = c.chat_session().await.unwrap();
        let probes_before = mock
            .methods()
            .iter()
            .filter(|method| method.as_str() == "Target.getTargets")
            .count();
        mock.hang_targets.store(true, Ordering::SeqCst);

        let busy = c.mark_page_busy();
        assert!(c.heartbeat().await);

        let probes_after = mock
            .methods()
            .iter()
            .filter(|method| method.as_str() == "Target.getTargets")
            .count();
        assert_eq!(probes_after, probes_before);
        assert!(!session.is_closed());
        assert!(c.status().await.cdp_connected);
        drop(busy);
    }

    #[tokio::test]
    async fn heartbeat_false_and_disconnected_when_targets_hang() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        assert!(c.heartbeat().await);
        assert!(c.status().await.cdp_connected);
        let mut rx = c.subscribe();

        mock.hang_targets.store(true, Ordering::SeqCst);
        let started = Instant::now();
        assert!(!c.heartbeat().await);
        // 首次失败 150ms + 重连后复测 150ms
        assert!(started.elapsed() < Duration::from_secs(2));
        let st = c.status().await;
        assert!(!st.cdp_connected);
        assert!(st.reconnecting);
        let ev = recv_event(&mut rx).await;
        assert!(
            matches!(ev, ConnEvent::StatusChanged(ref s) if !s.cdp_connected && s.reconnecting),
            "{ev:?}"
        );
        // 未到 20s、豆包仍在跑 → 没有 Crashed
        assert!(rx.try_recv().is_err());

        // 恢复响应 → 下一次心跳重连成功
        mock.hang_targets.store(false, Ordering::SeqCst);
        assert!(c.heartbeat().await);
        let st = c.status().await;
        assert!(st.cdp_connected && !st.reconnecting);
        let ev = recv_event(&mut rx).await;
        assert!(
            matches!(ev, ConnEvent::StatusChanged(ref s) if s.cdp_connected),
            "{ev:?}"
        );
    }

    #[tokio::test]
    async fn heartbeat_reconnects_immediately_after_server_close() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let s1 = c.chat_session().await.unwrap();
        mock.set_targets(vec![mock.chat_target("B")]);
        mock.close_all();
        while !s1.is_closed() {
            tokio::time::sleep(Duration::from_millis(5)).await;
        }
        // 老 session 已关 → 失败 → 立刻重连 → 复测成功 → true
        assert!(c.heartbeat().await);
        assert_eq!(c.chat_session().await.unwrap().target_id, "B");
        assert!(c.status().await.cdp_connected);
    }

    #[tokio::test]
    async fn heartbeat_emits_crashed_once_when_doubao_gone() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let mut timing = fast_timing();
        timing.crash_after = Duration::ZERO;
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let mut rx = c.subscribe();

        // 豆包"退出":没有任何 target,进程也不在
        mock.set_targets(vec![]);
        mock.close_all();
        platform.running.store(false, Ordering::SeqCst);
        assert!(!c.heartbeat().await);
        let mut saw_crashed = 0;
        let mut saw_disconnected = false;
        while let Ok(ev) = rx.try_recv() {
            match ev {
                ConnEvent::Crashed => saw_crashed += 1,
                ConnEvent::StatusChanged(s) if !s.cdp_connected => saw_disconnected = true,
                _ => {}
            }
        }
        assert_eq!(saw_crashed, 1);
        assert!(saw_disconnected);
        let st = c.status().await;
        assert!(!st.running && !st.cdp_connected && !st.reconnecting);

        // 同一次断线再心跳:不再重复发 Crashed
        assert!(!c.heartbeat().await);
        assert!(!rx
            .try_recv()
            .into_iter()
            .any(|e| matches!(e, ConnEvent::Crashed)));

        // 豆包回来了 → 自动重连
        mock.set_targets(vec![mock.chat_target("A2")]);
        platform.running.store(true, Ordering::SeqCst);
        assert!(c.heartbeat().await);
        let st = c.status().await;
        assert!(st.running && st.cdp_connected);
        assert_eq!(c.chat_session().await.unwrap().target_id, "A2");
    }

    /// UC-06:进程退出后下一次心跳 `running` 就要变 false,不等 `crash_after`(20s);`Crashed` 仍延后。
    #[tokio::test]
    async fn heartbeat_marks_not_running_before_crash_after_without_crashed() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let connections_before = mock.connections.load(Ordering::SeqCst);
        let mut rx = c.subscribe();

        mock.set_targets(vec![]);
        mock.close_all();
        platform.running.store(false, Ordering::SeqCst);
        let started = Instant::now();
        assert!(!c.heartbeat().await);
        assert!(started.elapsed() < Duration::from_secs(2));
        let st = c.status().await;
        assert!(!st.running, "进程退出后第一轮心跳就应标 running=false");
        assert!(!st.cdp_connected && !st.reconnecting);
        let mut saw_not_running = false;
        while let Ok(ev) = rx.try_recv() {
            match ev {
                ConnEvent::Crashed => panic!("crash_after=20s 未到不应发 Crashed"),
                ConnEvent::StatusChanged(s) if !s.running => saw_not_running = true,
                _ => {}
            }
        }
        assert!(saw_not_running, "应广播 running=false 的 StatusChanged");
        // 进程不在:不再去 /json/list 重连
        assert_eq!(mock.connections.load(Ordering::SeqCst), connections_before);
        assert!(!c.heartbeat().await);
        assert!(!rx
            .try_recv()
            .into_iter()
            .any(|e| matches!(e, ConnEvent::Crashed)));

        // 用户重启豆包 → 下一轮心跳自动恢复
        mock.set_targets(vec![mock.chat_target("A2")]);
        platform.running.store(true, Ordering::SeqCst);
        assert!(c.heartbeat().await);
        let st = c.status().await;
        assert!(st.running && st.cdp_connected && !st.reconnecting);
        assert_eq!(c.chat_session().await.unwrap().target_id, "A2");
    }

    /// Task 4.2:`/generate` 的就绪门槛。从未连接 / 豆包进程退出 → `doubao_not_ready`(503);
    /// 豆包还在、只是页面 reload 后心跳在重连(切账号)→ 放行;`last_seen_at` 记最后一次连通时刻。
    #[tokio::test]
    async fn ensure_ready_rejects_before_connect_and_after_exit_but_not_while_reconnecting() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        assert_eq!(c.last_seen_at(), 0);
        let e = c.ensure_ready().await.unwrap_err();
        assert_eq!(e.code, "doubao_not_ready");
        assert_eq!(e.http_status(), 503);
        assert_eq!(e.details.unwrap()["cdp_connected"], false);

        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        c.ensure_ready().await.unwrap();
        let seen = c.last_seen_at();
        assert!(seen > 0);
        assert!(c.is_doubao_running().await);

        // 页面 reload(豆包仍在):心跳失败 → reconnecting,放行
        mock.set_targets(vec![]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        let st = c.status().await;
        assert!(!st.cdp_connected && st.reconnecting && st.running);
        c.ensure_ready().await.unwrap();

        // 进程退出 → 拒绝
        platform.running.store(false, Ordering::SeqCst);
        assert!(!c.heartbeat().await);
        assert!(!c.is_doubao_running().await);
        let e = c.ensure_ready().await.unwrap_err();
        assert_eq!(e.code, "doubao_not_ready");
        assert_eq!(e.details.unwrap()["running"], false);
        assert_eq!(c.last_seen_at(), seen, "断线期间 last_seen_at 不前移");
    }

    /// 重连宽限期:断线未满 `crash_after` 放行,超过则 503(豆包被不带调试端口重开时 `reconnecting` 永远为 true)。
    #[test]
    fn reconnect_grace_is_bounded_by_crash_after() {
        let after = Duration::from_secs(20);
        assert!(reconnect_in_grace(None, after));
        assert!(reconnect_in_grace(Some(Duration::from_secs(5)), after));
        assert!(!reconnect_in_grace(Some(Duration::from_secs(20)), after));
        assert!(!reconnect_in_grace(Some(Duration::from_secs(600)), after));
    }

    #[tokio::test]
    async fn ensure_ready_rejects_reconnecting_after_crash_after() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            crash_after: Duration::from_millis(80),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        // 豆包仍在但再也连不上 CDP(例如被不带调试端口重新打开)
        mock.set_targets(vec![]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        assert!(c.status().await.reconnecting);
        c.ensure_ready().await.unwrap();
        tokio::time::sleep(Duration::from_millis(120)).await;
        assert!(!c.heartbeat().await);
        let st = c.status().await;
        assert!(st.reconnecting && st.running);
        let e = c.ensure_ready().await.unwrap_err();
        assert_eq!(e.code, "doubao_not_ready");
        let d = e.details.unwrap();
        assert_eq!(d["reconnecting"], true);
        assert!(d["outage_ms"].as_u64().unwrap() >= 80);
    }

    /// spec 08 任务冲突检测:有第二个 doubao-chat 窗口时,提交前就拒绝(不填 prompt、不发送、不耗配额),
    /// 并把 `chat_targets` 写进状态;窗口关掉后恢复。
    #[tokio::test]
    async fn submit_session_rejects_multiple_chat_pages_without_touching_page() {
        let evals = Arc::new(AtomicUsize::new(0));
        let counter = Arc::clone(&evals);
        let handler: Handler = Arc::new(StdMutex::new(Box::new(move |_expr: &str| {
            counter.fetch_add(1, Ordering::SeqCst);
            json!(null)
        })));
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let before = evals.load(Ordering::SeqCst);

        mock.set_targets(vec![mock.chat_target("A"), mock.chat_target("B")]);
        let e = c.submit_session().await.unwrap_err();
        assert_eq!(e.code, "generation_user_interrupted");
        let d = e.details.unwrap();
        assert_eq!(d["interruption"], "multiple_chat_pages");
        assert_eq!(d["chat_targets"], 2);
        assert_eq!(c.status().await.chat_targets, 2);
        assert_eq!(
            evals.load(Ordering::SeqCst),
            before,
            "拒绝时不得对页面执行任何脚本(fill / send)"
        );

        mock.set_targets(vec![mock.chat_target("A")]);
        let s = c.submit_session().await.unwrap();
        assert_eq!(s.target_id, "A");
        assert_eq!(c.status().await.chat_targets, 1);
    }

    /// 豆包还在但 CDP 挂起:`running` 保持 true,`reconnecting` 为 true。
    #[tokio::test]
    async fn heartbeat_keeps_running_true_while_process_alive() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        mock.hang_targets.store(true, Ordering::SeqCst);
        assert!(!c.heartbeat().await);
        let st = c.status().await;
        assert!(st.running && !st.cdp_connected && st.reconnecting);
    }

    // ---- Task 6.1:重连硬化 ----

    /// spec 09:心跳失败 → 关老 socket → **等 `reconnect_delay`** → 重新握手。
    #[tokio::test]
    async fn heartbeat_waits_reconnect_delay_before_reconnecting() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let timing = ConnTiming {
            reconnect_delay: Duration::from_millis(300),
            ..fast_timing()
        };
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let s1 = c.chat_session().await.unwrap();
        let connections = mock.connections.load(Ordering::SeqCst);
        mock.set_targets(vec![mock.chat_target("B")]);
        mock.close_all();
        while !s1.is_closed() {
            tokio::time::sleep(Duration::from_millis(5)).await;
        }
        let mut rx = c.subscribe();
        let started = Instant::now();
        let hb = {
            let c = Arc::clone(&c);
            tokio::spawn(async move { c.heartbeat().await })
        };
        // 等待期内先广播"重连中",且还没去建新连接
        let ev = recv_event(&mut rx).await;
        assert!(
            matches!(ev, ConnEvent::StatusChanged(ref s) if !s.cdp_connected && s.reconnecting),
            "{ev:?}"
        );
        tokio::time::sleep(Duration::from_millis(100)).await;
        assert_eq!(mock.connections.load(Ordering::SeqCst), connections);
        assert!(hb.await.unwrap());
        assert!(started.elapsed() >= Duration::from_millis(300));
        assert_eq!(c.chat_session().await.unwrap().target_id, "B");
    }

    /// spec 09:豆包进程还在、但断线自首次失败起 `restart_after` 仍连不上 → 升级为 `connect(true)`
    /// (杀掉并带调试端口重启),成功后发 `StatusChanged(cdp_connected)`。
    #[tokio::test]
    async fn heartbeat_escalates_to_restart_after_restart_after() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            restart_after: Duration::from_millis(500),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        // CDP 挂起:重连能建 socket 但复测无响应
        mock.hang_targets.store(true, Ordering::SeqCst);
        assert!(!c.heartbeat().await);
        assert_eq!(
            platform.kill_calls.load(Ordering::SeqCst),
            0,
            "未满 restart_after 不重启"
        );
        let mut rx = c.subscribe();

        tokio::time::sleep(Duration::from_millis(300)).await;
        assert!(c.heartbeat().await, "重启后连接恢复");
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 1);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 1);
        let st = c.status().await;
        assert!(st.cdp_connected && st.running && !st.reconnecting);
        let mut saw_connected = false;
        while let Ok(ev) = rx.try_recv() {
            if matches!(ev, ConnEvent::StatusChanged(ref s) if s.cdp_connected) {
                saw_connected = true;
            }
        }
        assert!(saw_connected, "重启成功要广播 doubao_status_changed");
        assert!(c.ensure_ready().await.is_ok());
    }

    /// 重启失败后同一次断线不再反复杀豆包(否则每 2s 杀一次);断线恢复后的下一次断线重新计时。
    #[tokio::test]
    async fn restart_is_attempted_once_per_outage() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            restart_after: Duration::from_millis(100),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        // 豆包进程在,但没有 chat page(窗口没了),重启也失败
        platform.launch_fails.store(true, Ordering::SeqCst);
        mock.set_targets(vec![]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        tokio::time::sleep(Duration::from_millis(130)).await;
        assert!(!c.heartbeat().await, "重启后仍无 chat page → 失败");
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 1);
        assert!(!c.retry_soon().await, "升级失败后回到常规心跳节奏");
        tokio::time::sleep(Duration::from_millis(130)).await;
        assert!(!c.heartbeat().await);
        assert_eq!(
            platform.kill_calls.load(Ordering::SeqCst),
            1,
            "同一次断线只重启一次"
        );

        // 恢复 → 新的断线重新计时、可再次升级
        mock.set_targets(vec![mock.chat_target("B")]);
        assert!(c.heartbeat().await);
        mock.set_targets(vec![]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 1);
        tokio::time::sleep(Duration::from_millis(130)).await;
        assert!(!c.heartbeat().await);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 2);
    }

    /// 进程已退出(UC-06"等待重启")不自动拉起豆包;多开 chat 窗口是用户干扰,也不杀豆包。
    #[tokio::test]
    async fn no_restart_when_doubao_exited_or_multiple_chat_pages() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            restart_after: Duration::from_millis(60),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();

        mock.set_targets(vec![mock.chat_target("B"), mock.chat_target("C")]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        tokio::time::sleep(Duration::from_millis(80)).await;
        assert!(!c.heartbeat().await);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
        assert_eq!(c.status().await.chat_targets, 2);

        platform.running.store(false, Ordering::SeqCst);
        mock.set_targets(vec![]);
        tokio::time::sleep(Duration::from_millis(80)).await;
        assert!(!c.heartbeat().await);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);
        assert!(!c.retry_soon().await, "进程不在时按常规心跳节奏查");
    }

    /// 多开 chat 窗口时快速重试无济于事:回到常规 10s 心跳,不每 2s 刷 WARN;窗口关掉后恢复。
    #[tokio::test]
    async fn multiple_chat_pages_falls_back_to_regular_heartbeat() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();

        mock.set_targets(vec![mock.chat_target("B"), mock.chat_target("C")]);
        mock.close_all();
        assert!(!c.heartbeat().await);
        assert!(!c.retry_soon().await, "多开窗口不快速重试");
        assert!(c.status().await.reconnecting);

        // 另一种失败原因(chat page 暂时不在)→ 重新进入快速重试
        mock.set_targets(vec![]);
        assert!(!c.heartbeat().await);
        assert!(c.retry_soon().await);

        mock.set_targets(vec![mock.chat_target("B")]);
        assert!(c.heartbeat().await);
        assert!(!c.retry_soon().await);
    }

    /// spec 09"安全校验(每次重连都要重跑)":心跳重连、`connect_sessions`、`connect(true)` 都要确认
    /// 9707 的监听进程位于豆包安装目录下,否则 `doubao_cdp_port_occupied`,也不升级去杀豆包。
    #[tokio::test]
    async fn reconnect_rejects_non_doubao_port_owner() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            restart_after: Duration::from_millis(60),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let connections = mock.connections.load(Ordering::SeqCst);

        // 豆包 socket 断开,端口换成了别的程序(target 列表照样像豆包)
        platform.set_owner(Some("/usr/local/bin/evil-cdp"));
        mock.set_targets(vec![mock.chat_target("B")]);
        mock.close_all();
        assert!(!c.heartbeat().await, "心跳重连不能连到非豆包监听者");
        assert!(!c.retry_soon().await, "端口被占用不快速重试");
        tokio::time::sleep(Duration::from_millis(80)).await;
        assert!(!c.heartbeat().await);
        assert_eq!(
            platform.kill_calls.load(Ordering::SeqCst),
            0,
            "端口被非豆包占用时不升级重启"
        );
        assert_eq!(
            mock.connections.load(Ordering::SeqCst),
            connections,
            "没有向非豆包监听者建 WebSocket"
        );
        assert!(!c.status().await.cdp_connected);

        let err = c.connect_sessions().await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_port_occupied");
        assert_eq!(
            err.details.as_ref().unwrap()["owning_process"],
            "/usr/local/bin/evil-cdp"
        );
        let err = c.chat_session().await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
        assert_eq!(
            err.details.as_ref().unwrap()["cause"],
            "doubao_cdp_port_occupied"
        );

        let err = c.connect(true).await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_port_occupied");
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0, "不白杀豆包");
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);

        // 查不到监听者 → 以 target 列表含 doubao-chat 兜底
        platform.set_owner(None);
        c.connect_sessions().await.unwrap();
        assert_eq!(c.chat_session().await.unwrap().target_id, "B");

        // 监听者恢复为豆包 → 心跳正常
        platform.set_owner(Some(DOUBAO_BROWSER));
        assert!(c.heartbeat().await);
    }

    /// 后台心跳:断线后按 `reconnect_delay` 快速重试,不必等下一个 10s 周期。
    #[tokio::test]
    async fn spawn_heartbeat_retries_quickly_while_reconnecting() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let timing = ConnTiming {
            reconnect_delay: Duration::from_millis(20),
            ..fast_timing()
        };
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        mock.set_targets(vec![]);
        mock.close_all();
        let handle = c.spawn_heartbeat(Duration::from_millis(700));
        // 第一次周期心跳在 700ms:发现断线、进入快速重试
        tokio::time::sleep(Duration::from_millis(800)).await;
        assert!(c.retry_soon().await);
        assert!(!c.status().await.cdp_connected);
        mock.set_targets(vec![mock.chat_target("B")]);
        // 下一个周期在 1400ms;快速重试应在那之前恢复
        tokio::time::sleep(Duration::from_millis(250)).await;
        let st = c.status().await;
        handle.abort();
        assert!(st.cdp_connected, "快速重试应已重连:{st:?}");
        assert!(!c.retry_soon().await);
    }

    /// 重连成功后正在捕获的任务继续:断线期间(豆包 reload,chat page 暂时不在)
    /// `await_chat_session` 每 `reconnect_delay` 重试,直到新 target 出现。
    #[tokio::test]
    async fn await_chat_session_waits_through_reload_gap() {
        let (handler, _) = page_handler();
        let mock = Arc::new(MockDoubao::start(handler).await);
        let timing = ConnTiming {
            reconnect_delay: Duration::from_millis(20),
            ..fast_timing()
        };
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let s1 = c.chat_session().await.unwrap();
        mock.set_targets(vec![]);
        mock.close_all();
        while !s1.is_closed() {
            tokio::time::sleep(Duration::from_millis(5)).await;
        }
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(150)).await;
            m.set_targets(vec![m.chat_target("B")]);
        });
        let cancel = CancellationToken::new();
        let s = tokio::time::timeout(Duration::from_secs(3), c.await_chat_session(&cancel))
            .await
            .unwrap()
            .unwrap();
        assert_eq!(s.target_id, "B");
    }

    /// 豆包进程不在 → 立刻失败;断线超过 `restart_after` → 失败;取消 → 立即返回。
    #[tokio::test]
    async fn await_chat_session_gives_up_when_exited_timed_out_or_cancelled() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let timing = ConnTiming {
            reconnect_delay: Duration::from_millis(20),
            restart_after: Duration::from_millis(150),
            ..fast_timing()
        };
        let c = conn(&mock, Arc::clone(&platform), timing);
        mock.set_targets(vec![]);
        let cancel = CancellationToken::new();

        let started = Instant::now();
        let e = c.await_chat_session(&cancel).await.unwrap_err();
        assert_eq!(e.code, "doubao_cdp_disconnected");
        assert!(started.elapsed() >= Duration::from_millis(150));
        assert!(started.elapsed() < Duration::from_secs(2));

        platform.running.store(false, Ordering::SeqCst);
        let started = Instant::now();
        assert!(c.await_chat_session(&cancel).await.is_err());
        assert!(started.elapsed() < Duration::from_millis(100));

        platform.running.store(true, Ordering::SeqCst);
        cancel.cancel();
        let started = Instant::now();
        assert!(c.await_chat_session(&cancel).await.is_err());
        assert!(started.elapsed() < Duration::from_millis(100));
    }

    #[tokio::test]
    async fn spawn_heartbeat_ticks() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let before = c.status().await.last_check_at;
        tokio::time::sleep(Duration::from_millis(2)).await;
        let handle = c.spawn_heartbeat(Duration::from_millis(20));
        tokio::time::sleep(Duration::from_millis(80)).await;
        handle.abort();
        assert!(c.status().await.last_check_at > before);
    }

    #[tokio::test]
    async fn shutdown_stops_heartbeat_closes_sessions_and_refuses_reconnect() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let platform = FakePlatform::new(true);
        let c = conn(&mock, Arc::clone(&platform), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let session = c.chat_session().await.unwrap();
        let handle = c.spawn_heartbeat(Duration::from_millis(20));

        c.shutdown().await;
        assert!(c.is_shutting_down());
        assert!(session.is_closed());
        let status = c.status().await;
        assert!(!status.cdp_connected);
        // 心跳循环自行退出
        tokio::time::timeout(Duration::from_secs(1), handle)
            .await
            .expect("heartbeat loop should exit after shutdown")
            .unwrap();
        // 心跳 / 重连都不再碰豆包,也不杀豆包
        assert!(!c.heartbeat().await);
        let err = c.connect_sessions().await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
        // 幂等
        c.shutdown().await;
    }

    // ---- 账号 ----

    #[tokio::test]
    async fn accounts_uses_two_second_cache_and_emits_event() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        let mut rx = c.subscribe();
        mock.set_targets(vec![mock.chat_target("A")]);
        let list = c.accounts(false).await.unwrap();
        assert_eq!(list.current.as_deref(), Some("甲"));
        assert_eq!(list.accounts.len(), 2);
        let reads_after_first = state.reads.load(Ordering::SeqCst);
        assert!(reads_after_first >= 1);
        // 事件:StatusChanged(连接)+ AccountListChanged
        let mut saw_list = false;
        while let Ok(ev) = rx.try_recv() {
            if let ConnEvent::AccountListChanged(l) = ev {
                assert_eq!(l.accounts.len(), 2);
                saw_list = true;
            }
        }
        assert!(saw_list);

        let _ = c.accounts(false).await.unwrap();
        assert_eq!(
            state.reads.load(Ordering::SeqCst),
            reads_after_first,
            "2s 内应走缓存"
        );
        assert!(rx.try_recv().is_err());

        let _ = c.accounts(true).await.unwrap();
        assert!(state.reads.load(Ordering::SeqCst) > reads_after_first);
        assert!(matches!(
            recv_event(&mut rx).await,
            ConnEvent::AccountListChanged(_)
        ));
    }

    #[tokio::test]
    async fn current_account_reads_visible_identity_without_opening_account_menu() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);

        let current = c.current_account().await.unwrap();

        assert_eq!(current.as_deref(), Some("甲"));
        assert_eq!(
            state.reads.load(Ordering::SeqCst),
            0,
            "只读当前账号不应打开或读取账号菜单"
        );
    }

    #[tokio::test]
    async fn switch_account_rejects_unknown_ambiguous_and_shortcuts_current() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        let mut rx = c.subscribe();

        let err = c.switch_account("丙").await.unwrap_err();
        assert_eq!(err.code, "account_not_found");
        let d = err.details.unwrap();
        assert_eq!(d["requested"], "丙");
        assert_eq!(d["available"], json!(["甲", "乙"]));

        let r = c.switch_account("甲").await.unwrap();
        assert!(r.ok);
        assert_eq!(r.current, "甲");
        assert_eq!(r.previous.as_deref(), Some("甲"));
        assert!(r.switched_at > 0);
        // 没有点击、没有 AccountSwitched
        let mut switched = false;
        while let Ok(ev) = rx.try_recv() {
            if matches!(ev, ConnEvent::AccountSwitched { .. }) {
                switched = true;
            }
        }
        assert!(!switched);
        assert_eq!(state.phase.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn switch_account_succeeds_through_reload_and_reconnect() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let mut rx = c.subscribe();

        // 点击后:页面 reload → 老 socket 断、target 变成 C、当前账号变成乙
        {
            let phase = Arc::clone(&state.phase);
            let targets = Arc::clone(&mock.targets);
            let close_tx = mock.close_tx.clone();
            let new_target = mock.chat_target("C");
            *state.on_click.lock().unwrap() = Some(Box::new(move || {
                *targets.lock().unwrap() = vec![new_target.clone()];
                phase.store(1, Ordering::SeqCst);
                let _ = close_tx.send(());
            }));
        }
        let started = Instant::now();
        let r = c.switch_account("乙").await.unwrap();
        assert!(r.ok);
        assert_eq!(r.previous.as_deref(), Some("甲"));
        assert_eq!(r.current, "乙");
        assert!(
            started.elapsed() < Duration::from_secs(4),
            "{:?}",
            started.elapsed()
        );
        assert_eq!(c.chat_session().await.unwrap().target_id, "C");

        let mut order = Vec::new();
        while let Ok(ev) = rx.try_recv() {
            match ev {
                ConnEvent::AccountSwitched {
                    previous,
                    current,
                    switched_at,
                } => {
                    assert_eq!(previous.as_deref(), Some("甲"));
                    assert_eq!(current, "乙");
                    assert!(switched_at > 0);
                    order.push("switched");
                }
                ConnEvent::AccountListChanged(l) => {
                    order.push(if l.current.as_deref() == Some("乙") {
                        "list_after"
                    } else {
                        "list_before"
                    });
                }
                _ => {}
            }
        }
        assert_eq!(
            order,
            vec!["list_before", "switched", "list_after"],
            "{order:?}"
        );
        // 切换后缓存已清:list 以新页面为准
        let list = c.accounts(false).await.unwrap();
        assert_eq!(list.current.as_deref(), Some("乙"));
        assert!(
            list.accounts
                .iter()
                .find(|a| a.name == "乙")
                .unwrap()
                .is_current
        );
    }

    #[tokio::test]
    async fn switch_account_times_out_when_current_never_changes() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let mut timing = fast_timing();
        timing.switch.verify_timeout = Duration::from_millis(150);
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        let started = Instant::now();
        let err = c.switch_account("乙").await.unwrap_err();
        assert_eq!(err.code, "account_switch_timeout");
        let elapsed = err.details.unwrap()["elapsed_ms"].as_u64().unwrap();
        assert!(elapsed >= 150, "{elapsed}");
        assert!(started.elapsed() < Duration::from_secs(3));
        // 失败不影响后续:session 仍可用
        assert!(c.chat_session().await.is_ok());
    }

    /// 验证循环里 `current()` 挂起(`Runtime.evaluate` 默认 30s)不能把验证窗口拖长。
    #[tokio::test]
    async fn switch_account_verify_window_bounded_when_evaluate_hangs() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let mut timing = fast_timing();
        timing.switch.verify_timeout = Duration::from_millis(200);
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        {
            // 点击后页面"卡住":current() 永不返回
            let hang = Arc::clone(&mock.hang_current);
            *state.on_click.lock().unwrap() = Some(Box::new(move || {
                hang.store(true, Ordering::SeqCst);
            }));
        }
        let started = Instant::now();
        let err = c.switch_account("乙").await.unwrap_err();
        assert_eq!(err.code, "account_switch_timeout");
        let elapsed = err.details.unwrap()["elapsed_ms"].as_u64().unwrap();
        assert!(elapsed >= 200, "{elapsed}");
        assert!(
            started.elapsed() < Duration::from_secs(2),
            "{:?}:验证应在 deadline 附近结束,而不是等 evaluate 30s",
            started.elapsed()
        );
        mock.hang_current.store(false, Ordering::SeqCst);
        assert!(c.chat_session().await.is_ok());
    }

    #[tokio::test]
    async fn switch_account_fails_fast_when_login_modal_appears() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        {
            let targets = Arc::clone(&mock.targets);
            let chat = mock.chat_target("A");
            let modal = mock.login_modal_target();
            *state.on_click.lock().unwrap() = Some(Box::new(move || {
                *targets.lock().unwrap() = vec![chat.clone(), modal.clone()];
            }));
        }
        let started = Instant::now();
        let err = c.switch_account("乙").await.unwrap_err();
        assert_eq!(err.code, "account_switch_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "login_required");
        assert!(d["login_modal_url"]
            .as_str()
            .unwrap()
            .contains("client-login-modal"));
        assert!(
            started.elapsed() < Duration::from_secs(2),
            "不应等到 5s 超时"
        );
        // 没有向登录窗口建 session
        assert_eq!(c.chat_session().await.unwrap().target_id, "A");
    }

    #[tokio::test]
    async fn switch_account_fails_when_row_never_clickable() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);
        *state.switch_action.lock().unwrap() = "not_found".into();
        let err = c.switch_account("乙").await.unwrap_err();
        assert_eq!(err.code, "account_switch_failed");
        assert_eq!(err.details.unwrap()["reason"], "account_row_not_clickable");
    }

    #[tokio::test]
    async fn switch_and_accounts_are_serialized_by_switch_lock() {
        let (handler, _) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let mut timing = fast_timing();
        timing.switch.verify_timeout = Duration::from_millis(120);
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        c.connect_sessions().await.unwrap();
        let c2 = Arc::clone(&c);
        let switching = tokio::spawn(async move { c2.switch_account("乙").await });
        tokio::time::sleep(Duration::from_millis(20)).await;
        let started = Instant::now();
        let list = c.accounts(true).await.unwrap();
        // 读列表要等切换(超时 120ms)结束后才拿到锁
        assert!(
            started.elapsed() >= Duration::from_millis(60),
            "{:?}",
            started.elapsed()
        );
        assert_eq!(list.accounts.len(), 2);
        assert_eq!(
            switching.await.unwrap().unwrap_err().code,
            "account_switch_timeout"
        );
    }

    /// 切换进行中(持 `switch_lock`),2s 内的缓存命中不该被它阻塞。
    #[tokio::test]
    async fn accounts_cache_hit_does_not_wait_for_switch_lock() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let mut timing = fast_timing();
        timing.switch.verify_timeout = Duration::from_millis(400);
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        let _warm = c.accounts(false).await.unwrap();
        let c2 = Arc::clone(&c);
        let switching = tokio::spawn(async move { c2.switch_account("乙").await });
        // 切换先 accounts_locked(true) 刷新缓存(fast timing 下几十 ms 内完成),之后进入 400ms 验证循环
        tokio::time::sleep(Duration::from_millis(60)).await;
        let reads_before = state.reads.load(Ordering::SeqCst);
        let started = Instant::now();
        let hit = c.accounts(false).await.unwrap();
        assert!(
            started.elapsed() < Duration::from_millis(150),
            "{:?}:缓存命中不应等切换(≥400ms)结束",
            started.elapsed()
        );
        assert_eq!(hit.accounts.len(), 2);
        assert_eq!(
            state.reads.load(Ordering::SeqCst),
            reads_before,
            "命中缓存不应再读菜单"
        );
        assert_eq!(
            switching.await.unwrap().unwrap_err().code,
            "account_switch_timeout"
        );
    }

    /// 任务正在使用页面(提交 / 捕获)时,缓存过期也不再打开账号菜单;页面空闲后恢复按 TTL 读取。
    #[tokio::test]
    async fn accounts_reuse_stale_cache_while_page_busy() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let mut timing = fast_timing();
        timing.account_cache_ttl = Duration::from_millis(30);
        let c = conn(&mock, FakePlatform::new(true), timing);
        mock.set_targets(vec![mock.chat_target("A")]);
        let warm = c.accounts(false).await.unwrap();
        tokio::time::sleep(Duration::from_millis(50)).await;
        let reads = state.reads.load(Ordering::SeqCst);

        {
            let _busy = c.mark_page_busy();
            assert!(c.page_busy());
            let stale = c.accounts(false).await.unwrap();
            assert_eq!(stale.read_at, warm.read_at, "页面忙时应沿用过期缓存");
            assert_eq!(
                state.reads.load(Ordering::SeqCst),
                reads,
                "页面忙时不应读菜单"
            );
        }
        assert!(!c.page_busy());
        let fresh = c.accounts(false).await.unwrap();
        assert!(
            state.reads.load(Ordering::SeqCst) > reads,
            "页面空闲后缓存过期应重新读"
        );
        assert_eq!(fresh.accounts.len(), 2);
    }

    /// 提交持页面锁期间:有缓存的读取立即返回、不读菜单;没有缓存的读取要等提交释放页面锁。
    #[tokio::test]
    async fn accounts_wait_for_page_lock_when_no_cache() {
        let (handler, state) = page_handler();
        let mock = MockDoubao::start(handler).await;
        let c = conn(&mock, FakePlatform::new(true), fast_timing());
        mock.set_targets(vec![mock.chat_target("A")]);

        let page = c.lock_page().await;
        assert!(c.page_busy());
        let c2 = Arc::clone(&c);
        let reader = tokio::spawn(async move { c2.accounts(false).await });
        tokio::time::sleep(Duration::from_millis(150)).await;
        assert!(
            !reader.is_finished(),
            "无缓存的菜单读取必须等提交释放页面锁"
        );
        assert_eq!(state.reads.load(Ordering::SeqCst), 0);
        drop(page);
        let list = tokio::time::timeout(Duration::from_secs(5), reader)
            .await
            .unwrap()
            .unwrap()
            .unwrap();
        assert_eq!(list.accounts.len(), 2);
        assert!(!c.page_busy());

        // 有缓存(即使过期)时,提交持锁期间的读取不阻塞
        let _page = c.lock_page().await;
        let reads = state.reads.load(Ordering::SeqCst);
        let started = Instant::now();
        let hit = tokio::time::timeout(Duration::from_secs(1), c.accounts(false))
            .await
            .expect("有缓存时不应等页面锁")
            .unwrap();
        assert!(started.elapsed() < Duration::from_millis(150));
        assert_eq!(hit.read_at, list.read_at);
        assert_eq!(state.reads.load(Ordering::SeqCst), reads);
    }

    // ---- 真机 ----
    //
    // 三个要操作账号菜单的 live_ 测试(含 accounts::live_read_accounts)都先拿 `LIVE_MENU_LOCK`,
    // `cargo test live_ -- --ignored` 多线程并行时才不会互相开/关菜单。

    /// 需要真豆包:`cargo test live_ -- --ignored --nocapture`。只连接、读状态与账号,不点击任何账号。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_connect_status_and_heartbeat() {
        let _menu = crate::doubao::LIVE_MENU_LOCK.lock().await;
        let platform = crate::platform::create_platform();
        let c = DoubaoConnection::new(platform, "127.0.0.1".into(), 9707);
        let started = Instant::now();
        let action = c.connect(false).await.unwrap();
        println!("connect: {action:?} ({}ms)", started.elapsed().as_millis());
        let st = c.status().await;
        println!("status: {st:?}");
        assert_eq!(action, StartupAction::AlreadyConnected);
        assert!(st.running && st.cdp_connected);
        assert_eq!(st.chat_targets, 1);
        let started = Instant::now();
        assert!(c.heartbeat().await);
        println!("heartbeat ok ({}ms)", started.elapsed().as_millis());
        let started = Instant::now();
        let list = c.accounts(false).await.unwrap();
        println!("accounts ({}ms): {list:?}", started.elapsed().as_millis());
        assert!(!list.accounts.is_empty());
        let cached = c.accounts(false).await.unwrap();
        assert_eq!(cached.read_at, list.read_at, "2s 内应走缓存");
        c.close_sessions().await;
    }

    /// 需要真豆包(Task 6.1):relay 一侧把 chat / background socket 全部关掉(模拟断线,豆包不受影响),
    /// 心跳应在 `reconnect_delay`(2s)后重连同一个 target,不触发重启;`await_chat_session` 同样能接上。
    /// 只连接,不点击、不发送、不杀豆包。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_heartbeat_reconnects_after_socket_drop() {
        let platform = crate::platform::create_platform();
        let c = DoubaoConnection::new(platform, "127.0.0.1".into(), 9707);
        c.connect(false).await.unwrap();
        let before = c.chat_session().await.unwrap();
        let target = before.target_id.clone();
        for round in 1..=3 {
            c.close_sessions().await;
            assert!(before.is_closed());
            let started = Instant::now();
            assert!(c.heartbeat().await, "第 {round} 轮心跳应重连成功");
            let ms = started.elapsed().as_millis();
            let st = c.status().await;
            println!("round {round}: heartbeat reconnect {ms}ms, status {st:?}");
            assert!(st.cdp_connected && !st.reconnecting);
            assert!((2000..10_000).contains(&ms), "应先等 2s 再重连:{ms}ms");
            let now = c.chat_session().await.unwrap();
            assert_eq!(now.target_id, target, "豆包未 reload,target 不变");
        }
        // 捕获路径:session 关闭后 await_chat_session 立即接上
        c.close_sessions().await;
        let started = Instant::now();
        let s = c
            .await_chat_session(&CancellationToken::new())
            .await
            .unwrap();
        println!(
            "await_chat_session reconnect {}ms",
            started.elapsed().as_millis()
        );
        assert_eq!(s.target_id, target);
        assert!(!c.retry_soon().await);
        c.close_sessions().await;
    }

    /// 需要真豆包(Task 4.2):`ensure_ready` 在连上前拒绝、连上后放行;`last_seen_at` / 进程探测与真机一致。
    /// 只连接,不点击、不发送。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_ensure_ready_and_process_probe() {
        let platform = crate::platform::create_platform();
        let c = DoubaoConnection::new(platform, "127.0.0.1".into(), 9707);
        let e = c.ensure_ready().await.unwrap_err();
        println!("before connect: {} {:?}", e.code, e.details);
        assert_eq!(e.code, "doubao_not_ready");
        let started = Instant::now();
        c.connect_sessions().await.unwrap();
        c.ensure_ready().await.unwrap();
        println!(
            "connect_sessions + ensure_ready ok ({}ms), last_seen_at={}",
            started.elapsed().as_millis(),
            c.last_seen_at()
        );
        assert!(c.last_seen_at() > 0);
        let started = Instant::now();
        assert!(c.is_doubao_running().await);
        println!(
            "is_doubao_running=true ({}ms)",
            started.elapsed().as_millis()
        );
        assert!(c.heartbeat().await);
        // 提交前的 chat page 计数检查(只读 /json/list,不碰页面)
        let started = Instant::now();
        let s = c.submit_session().await.unwrap();
        println!(
            "submit_session ok ({}ms), target={}, chat_targets={}",
            started.elapsed().as_millis(),
            s.target_id,
            c.status().await.chat_targets
        );
        assert_eq!(c.status().await.chat_targets, 1);
        c.close_sessions().await;
    }

    /// 需要真豆包且 ≥2 个有效账号:切到非当前账号再切回。账号不足时打印并跳过(本机只剩 1 个,见 docs/analysis/10)。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_switch_roundtrip() {
        let _menu = crate::doubao::LIVE_MENU_LOCK.lock().await;
        let platform = crate::platform::create_platform();
        let c = DoubaoConnection::new(platform, "127.0.0.1".into(), 9707);
        c.connect(false).await.unwrap();
        let list = c.accounts(true).await.unwrap();
        println!("accounts: {list:?}");
        let Some(current) = list.current.clone() else {
            println!("SKIP: 未登录");
            return;
        };
        let wanted = std::env::var("DOUBAO_SWITCH_TARGET").ok();
        let Some(other) = list
            .accounts
            .iter()
            .find(|a| {
                !a.is_current && !a.is_ambiguous && wanted.as_deref().is_none_or(|w| a.name == w)
            })
            .map(|a| a.name.clone())
        else {
            println!("SKIP: 本机只有 1 个有效账号,无法验证切换往返");
            return;
        };
        for target in [other.as_str(), current.as_str()] {
            let started = Instant::now();
            let r = c.switch_account(target).await;
            println!(
                "switch → {target}: {r:?} ({}ms)",
                started.elapsed().as_millis()
            );
            let r = r.unwrap();
            assert_eq!(r.current, target);
        }
        c.close_sessions().await;
    }
}
