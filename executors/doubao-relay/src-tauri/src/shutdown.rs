//! 关闭流程(`docs/spec/03-architecture.md` "关闭流程")与"关闭窗口最小化到托盘"的判定。
//!
//! 1. 停止接受新请求:`state.shutdown.cancel()`(HTTP server 优雅关闭;`AppState::submit_task`
//!    此后一律 `system_shutdown`,Tauri `task_submit` 同样拒绝);
//! 2. pending / preparing → 立即取消(`cancelled`);
//! 3. generating → 最多等 10s 自然结束,之后 `fail_all_running("system_shutdown")` 强制 `failed`;
//!    `waiting_user_confirm` 没有用户确认不可能自行结束,不单独等待,随 generating 一起处理;
//! 4. 收尾:仍非终态的(runner 卡住)直接在库里标 `failed/system_shutdown`;
//! 5. [`close_resources`]:关 SQLite 连接池、flush 日志。CDP 连接随进程退出关闭,**不杀豆包**。
//!
//! [`graceful_shutdown`] 幂等:并发 / 重复调用只有第一次真正执行,其余等它完成后返回同一份报告。

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use serde::Serialize;
use tauri::{AppHandle, Manager, Runtime};
use tokio::sync::OnceCell;

use crate::config::UiConfig;
use crate::error::RelayError;
use crate::logging::{self, now_ms};
use crate::state::AppState;
use crate::tasks::TaskStatus;

/// 关闭时强制失败的错误码(11-error-codes `system_shutdown`)。
pub const SHUTDOWN_CODE: &str = "system_shutdown";
/// 关闭时被强制中止的任务写入的失败信息。
pub const SHUTDOWN_TASK_MESSAGE: &str = "中转程序正在关闭,任务已中止,请重新打开后重新提交";

/// `system_shutdown`:关闭流程开始后拒绝新任务(HTTP 500 按 `system_*` 前缀映射)。
pub fn shutdown_error() -> RelayError {
    RelayError::new(
        SHUTDOWN_CODE,
        "中转程序正在关闭,不再接受新任务,请重新打开后再提交",
    )
}

/// 各阶段时长(默认即 spec;测试用小值)。
#[derive(Clone, Debug)]
pub struct ShutdownTiming {
    /// generating 任务自然结束的等待上限(spec:10s)。
    pub generating_grace: Duration,
    /// 取消 / 强制失败后等 runner 落库的上限。
    pub settle: Duration,
    /// 轮询间隔。
    pub poll: Duration,
}

impl Default for ShutdownTiming {
    fn default() -> Self {
        Self {
            generating_grace: Duration::from_secs(10),
            settle: Duration::from_secs(5),
            poll: Duration::from_millis(100),
        }
    }
}

impl ShutdownTiming {
    /// `RunEvent::Exit` 兜底路径(系统注销等不可阻止的退出)用的短时限。
    pub fn quick() -> Self {
        Self {
            generating_grace: Duration::from_secs(2),
            settle: Duration::from_secs(1),
            poll: Duration::from_millis(50),
        }
    }
}

/// 关闭流程的结果(日志与测试用)。
#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize)]
pub struct ShutdownReport {
    /// pending / preparing 被直接取消的任务。
    pub cancelled: Vec<String>,
    /// 宽限期结束时仍在运行、被强制 `failed/system_shutdown` 的任务。
    pub forced: Vec<String>,
    /// runner 没能在 settle 时限内落库、最后直接在库里标失败的条数。
    pub marked_in_db: u64,
    pub duration_ms: u64,
}

/// 退出协调:保证关闭流程只跑一次,并记录是否已完成(`RunEvent::ExitRequested` 据此放行)。
#[derive(Default)]
pub struct QuitGate {
    started: AtomicBool,
    done: AtomicBool,
    report: OnceCell<ShutdownReport>,
}

impl QuitGate {
    /// 第一次调用返回 true(由调用方负责跑关闭流程),之后都返回 false。
    pub fn begin(&self) -> bool {
        !self.started.swap(true, Ordering::SeqCst)
    }

    pub fn is_started(&self) -> bool {
        self.started.load(Ordering::SeqCst)
    }

    pub fn is_done(&self) -> bool {
        self.done.load(Ordering::SeqCst)
    }

    pub fn mark_done(&self) {
        self.done.store(true, Ordering::SeqCst);
    }
}

/// 点窗口关闭按钮时的动作。
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum CloseAction {
    /// 拦截关闭,隐藏窗口到托盘(HTTP 服务继续运行)。
    HideToTray,
    /// 照常关闭 → 最后一个窗口关闭 → 走关闭流程退出。
    Quit,
}

/// `ui.minimize_to_tray_on_close` → 关闭按钮动作(设置页改了立即生效,每次读当前配置)。
pub fn close_action(ui: &UiConfig) -> CloseAction {
    if ui.minimize_to_tray_on_close {
        CloseAction::HideToTray
    } else {
        CloseAction::Quit
    }
}

/// 执行关闭流程(见模块文档)。幂等:只有第一次调用真正执行。
pub async fn graceful_shutdown(state: &AppState, timing: &ShutdownTiming) -> ShutdownReport {
    state
        .quit
        .report
        .get_or_init(|| run_shutdown(state, timing))
        .await
        .clone()
}

async fn run_shutdown(state: &AppState, timing: &ShutdownTiming) -> ShutdownReport {
    let started = Instant::now();
    let mut report = ShutdownReport::default();
    tracing::info!("开始关闭:停止接受新请求");
    state.shutdown.cancel();

    // 2. pending / preparing 立即取消(调度器已派发但 runner 还没写 preparing 的也算)
    cancel_queued_and_preparing(state, &mut report).await;

    // 3. generating 最多等 grace 自然结束;只剩 waiting_user_confirm 时不再等
    let deadline = Instant::now() + timing.generating_grace;
    loop {
        // 宽限期内又有 pending 被派发成 preparing(调度循环与第 2 步竞争)也一并取消
        cancel_queued_and_preparing(state, &mut report).await;
        let running = state.scheduler.running_task_ids(None);
        if running.is_empty() {
            break;
        }
        // 只剩 waiting_user_confirm(没人确认不会自行结束):不必等满宽限期
        if all_waiting_confirm(state, &running).await || Instant::now() >= deadline {
            break;
        }
        tokio::time::sleep(timing.poll).await;
    }

    let remaining = state.scheduler.running_task_ids(None);
    if !remaining.is_empty() {
        tracing::warn!(count = remaining.len(), "宽限期已到,强制中止仍在进行的任务");
        state
            .scheduler
            .fail_all_running(SHUTDOWN_CODE, SHUTDOWN_TASK_MESSAGE)
            .await;
        report.forced = remaining;
    }

    // 等 runner 把终态写进库(handle 置终态 / 移出 running 与写库之间有先后,两者都要等)
    let settle_deadline = Instant::now() + timing.settle;
    loop {
        let mut tracked: Vec<String> = report.cancelled.clone();
        tracked.extend(report.forced.iter().cloned());
        tracked.extend(state.scheduler.running_task_ids(None));
        let mut settled = true;
        for id in &tracked {
            if status_of(state, id).await.is_some_and(|s| !s.is_terminal()) {
                settled = false;
                break;
            }
        }
        if settled || Instant::now() >= settle_deadline {
            break;
        }
        tokio::time::sleep(timing.poll).await;
    }

    // 4. 兜底:库里仍非终态的(runner 卡住 / 刚好漏网的 pending)直接标失败
    match state
        .storage
        .fail_non_terminal(SHUTDOWN_CODE, SHUTDOWN_TASK_MESSAGE, now_ms())
        .await
    {
        Ok(ids) => {
            report.marked_in_db = ids.len() as u64;
            // 与启动恢复一致:兜底标失败的任务也走终态通知(事件 + webhook,尽力而为)
            for id in &ids {
                if let Ok(Some(task)) = state.storage.get_task(id).await {
                    state.scheduler.notify_terminal(&task);
                }
            }
        }
        Err(e) => tracing::error!(code = e.code, error = %e, "关闭时标记未完成任务失败"),
    }

    report.duration_ms = started.elapsed().as_millis() as u64;
    tracing::info!(
        cancelled = report.cancelled.len(),
        forced = report.forced.len(),
        marked_in_db = report.marked_in_db,
        duration_ms = report.duration_ms,
        "关闭流程完成(豆包保持运行)"
    );
    report
}

/// 取消全部 pending,以及运行中但仍处于 pending / preparing 的任务。
async fn cancel_queued_and_preparing(state: &AppState, report: &mut ShutdownReport) {
    let mut ids: Vec<String> = Vec::new();
    match state
        .storage
        .list_tasks(Some(&[TaskStatus::Pending]), None, 1000, None)
        .await
    {
        Ok((rows, _)) => ids.extend(rows.into_iter().map(|t| t.id)),
        Err(e) => tracing::warn!(code = e.code, error = %e, "关闭时读取排队任务失败"),
    }
    for id in state.scheduler.running_task_ids(None) {
        if matches!(
            status_of(state, &id).await,
            Some(TaskStatus::Pending | TaskStatus::Preparing)
        ) && !ids.contains(&id)
        {
            ids.push(id);
        }
    }
    for id in ids {
        if report.cancelled.contains(&id) {
            continue;
        }
        match state.scheduler.cancel(&id).await {
            Ok(from) => {
                tracing::info!(task_id = %id, from_status = from.as_str(), "关闭:取消未开始生成的任务");
                report.cancelled.push(id);
            }
            // 已经结束 / 刚被别处取消:忽略
            Err(e) => tracing::debug!(task_id = %id, code = e.code, "关闭:取消任务跳过"),
        }
    }
}

async fn status_of(state: &AppState, id: &str) -> Option<TaskStatus> {
    state
        .storage
        .get_task(id)
        .await
        .ok()
        .flatten()
        .map(|t| t.status)
}

async fn all_waiting_confirm(state: &AppState, running: &[String]) -> bool {
    for id in running {
        if !matches!(
            status_of(state, id).await,
            Some(TaskStatus::WaitingUserConfirm)
        ) {
            return false;
        }
    }
    true
}

/// 发起退出(托盘"退出"、macOS Cmd+Q、隐私声明"不同意并退出"、最后一个窗口关闭):
/// 隐藏窗口 → 关闭流程 → 关资源 → `app.exit(0)`。重复调用无效(只跑一次)。
pub fn request_quit<R: Runtime>(app: &AppHandle<R>) {
    let Some(state) = app.try_state::<Arc<AppState>>() else {
        app.exit(0);
        return;
    };
    let state = Arc::clone(&state);
    if !state.quit.begin() {
        return;
    }
    for w in app.webview_windows().values() {
        let _ = w.hide();
    }
    let app = app.clone();
    tauri::async_runtime::spawn(async move {
        graceful_shutdown(&state, &ShutdownTiming::default()).await;
        close_resources(&state).await;
        state.quit.mark_done();
        app.exit(0);
    });
}

/// `RunEvent::Exit` 兜底(系统注销等绕过 `ExitRequested` 的退出):关闭流程没跑过就用短时限跑一遍。
/// 在主线程同步调用。
pub fn finish_on_exit(state: &Arc<AppState>) {
    if state.quit.is_done() {
        return;
    }
    let state = Arc::clone(state);
    tauri::async_runtime::block_on(async move {
        graceful_shutdown(&state, &ShutdownTiming::quick()).await;
        close_resources(&state).await;
        state.quit.mark_done();
    });
}

/// 关闭流程第 5 步:停心跳并断开 CDP 会话(不杀豆包)、关 SQLite 连接池、flush 日志文件。
/// 之后不能再用 `state.storage`。
pub async fn close_resources(state: &AppState) {
    state.doubao.shutdown().await;
    tracing::info!("CDP 会话已断开(豆包保持运行)");
    state.storage.pool.close().await;
    tracing::info!("数据库已关闭,日志 flush 后退出");
    logging::flush();
}

#[cfg(test)]
mod tests {
    use super::*;

    use std::path::{Path, PathBuf};
    use std::sync::Arc;

    use async_trait::async_trait;
    use tokio::sync::{mpsc, watch};
    use tokio_util::sync::CancellationToken;

    use crate::commands::testing::test_state_with;
    use crate::doubao::capture::{CaptureEvent, CaptureOptions, CaptureOutcome};
    use crate::doubao::download::Downloaded;
    use crate::doubao::generation::{
        self, StageFn, SubmissionIdentity, SubmissionPlan, SubmissionReport,
    };
    use crate::doubao::{Account, AccountList};
    use crate::error::RelayResult;
    use crate::tasks::{ConfirmationMode, GenerateRequest, GenerationBackend, TaskType};

    /// 按 prompt 决定行为:`hang` → submit 挂起(停在 preparing);`gen` → 进入 generating 后
    /// capture 挂起到取消;`fail:<ms>` → generating 后 `<ms>` 毫秒以 `generation_doubao_refused` 结束。
    struct ScriptBackend;

    #[async_trait]
    impl GenerationBackend for ScriptBackend {
        async fn accounts(&self, _force: bool) -> RelayResult<AccountList> {
            Ok(AccountList {
                current: Some("张三".into()),
                accounts: vec![Account {
                    name: "张三".into(),
                    subscription: Some("体验套餐".into()),
                    label: "张三 体验套餐".into(),
                    is_current: true,
                    is_ambiguous: false,
                }],
                read_at: now_ms(),
            })
        }

        async fn ensure_account(&self, _account: &str) -> RelayResult<()> {
            Ok(())
        }

        async fn submit(
            &self,
            request: &GenerateRequest,
            _plan: &SubmissionPlan,
            _reference_paths: &[PathBuf],
            cancel: &CancellationToken,
            _on_stage: StageFn<'_>,
        ) -> RelayResult<SubmissionReport> {
            if request.prompt == "hang" {
                cancel.cancelled().await;
                return Err(generation::cancelled("prepare"));
            }
            Ok(SubmissionReport {
                identity: Some(SubmissionIdentity {
                    local_message_id: Some("lm1".into()),
                    message_id: None,
                    conversation_id: Some("7001".into()),
                }),
                send_started_at: now_ms(),
                generation_requests: Vec::new(),
            })
        }

        async fn capture(
            &self,
            opts: CaptureOptions,
            cancel: &CancellationToken,
            _confirm_rx: &mut watch::Receiver<bool>,
            _events: mpsc::Sender<CaptureEvent>,
        ) -> RelayResult<CaptureOutcome> {
            let delay = opts
                .prompt
                .strip_prefix("fail:")
                .and_then(|ms| ms.parse::<u64>().ok());
            match delay {
                Some(ms) => {
                    tokio::select! {
                        _ = cancel.cancelled() => Err(generation::cancelled("capture")),
                        _ = tokio::time::sleep(Duration::from_millis(ms)) => Err(RelayError::new(
                            "generation_doubao_refused",
                            "豆包拒绝生成",
                        )),
                    }
                }
                None => {
                    cancel.cancelled().await;
                    Err(generation::cancelled("capture"))
                }
            }
        }

        async fn download_image(
            &self,
            _url: &str,
            _dest_dir: &Path,
            _index: u32,
        ) -> RelayResult<Downloaded> {
            Err(RelayError::internal("不下载"))
        }

        async fn download_video(&self, _url: &str, _dest_dir: &Path) -> RelayResult<Downloaded> {
            Err(RelayError::internal("不下载"))
        }
    }

    fn request(prompt: &str) -> GenerateRequest {
        GenerateRequest {
            account: "张三".into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: prompt.into(),
            ratio: "1:1".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    async fn state_with_loop(dir: &Path) -> Arc<AppState> {
        let state = test_state_with(dir, Some(Arc::new(ScriptBackend))).await;
        state.scheduler.spawn_loop();
        state
    }

    async fn wait_status(state: &AppState, id: &str, want: TaskStatus) {
        for _ in 0..300 {
            if status_of(state, id).await == Some(want) {
                return;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        panic!(
            "任务 {id} 3s 内没有进入 {want:?},当前 {:?}",
            status_of(state, id).await
        );
    }

    async fn task(state: &AppState, id: &str) -> crate::tasks::Task {
        state.storage.get_task(id).await.unwrap().unwrap()
    }

    fn timing(grace_ms: u64) -> ShutdownTiming {
        ShutdownTiming {
            generating_grace: Duration::from_millis(grace_ms),
            settle: Duration::from_secs(3),
            poll: Duration::from_millis(20),
        }
    }

    #[test]
    fn close_action_follows_minimize_to_tray_setting() {
        let mut ui = UiConfig::default();
        assert!(ui.minimize_to_tray_on_close, "spec 07 默认 true");
        assert_eq!(close_action(&ui), CloseAction::HideToTray);
        ui.minimize_to_tray_on_close = false;
        assert_eq!(close_action(&ui), CloseAction::Quit);
    }

    #[test]
    fn quit_gate_begins_once() {
        let g = QuitGate::default();
        assert!(!g.is_started());
        assert!(g.begin());
        assert!(!g.begin());
        assert!(g.is_started());
        assert!(!g.is_done());
        g.mark_done();
        assert!(g.is_done());
    }

    #[test]
    fn shutdown_error_is_system_shutdown() {
        let e = shutdown_error();
        assert_eq!(e.code, "system_shutdown");
        assert_eq!(crate::error::http_status_for_code(e.code), 500);
    }

    #[tokio::test]
    async fn pending_and_preparing_are_cancelled_immediately() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        let (preparing, _, _) = state.scheduler.submit(request("hang")).await.unwrap();
        wait_status(&state, &preparing, TaskStatus::Preparing).await;
        // 同账号 per_account_max=1:第二个停在 pending
        let (pending, _, _) = state.scheduler.submit(request("gen")).await.unwrap();
        assert_eq!(status_of(&state, &pending).await, Some(TaskStatus::Pending));

        let t0 = Instant::now();
        let report = graceful_shutdown(&state, &timing(5_000)).await;
        assert!(
            t0.elapsed() < Duration::from_secs(2),
            "没有 generating 任务时不等宽限期:{:?}",
            t0.elapsed()
        );
        assert!(state.shutdown.is_cancelled(), "HTTP 停止接受新请求");
        let mut cancelled = report.cancelled.clone();
        cancelled.sort();
        let mut want = vec![preparing.clone(), pending.clone()];
        want.sort();
        assert_eq!(cancelled, want);
        assert!(report.forced.is_empty());
        assert_eq!(report.marked_in_db, 0);
        let p = task(&state, &preparing).await;
        assert_eq!(p.status, TaskStatus::Cancelled);
        assert_eq!(p.error.unwrap().code, "generation_cancelled");
        assert_eq!(task(&state, &pending).await.status, TaskStatus::Cancelled);
        assert_eq!(state.scheduler.running_count(), 0);
    }

    #[tokio::test]
    async fn generating_is_forced_failed_after_grace() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        let (id, _, _) = state.scheduler.submit(request("gen")).await.unwrap();
        wait_status(&state, &id, TaskStatus::Generating).await;

        let t0 = Instant::now();
        let report = graceful_shutdown(&state, &timing(300)).await;
        let elapsed = t0.elapsed();
        assert!(elapsed >= Duration::from_millis(300), "{elapsed:?}");
        assert!(elapsed < Duration::from_secs(3), "{elapsed:?}");
        assert_eq!(report.forced, vec![id.clone()]);
        assert!(report.cancelled.is_empty());
        let t = task(&state, &id).await;
        assert_eq!(t.status, TaskStatus::Failed);
        let err = t.error.unwrap();
        assert_eq!(err.code, "system_shutdown");
        assert_eq!(err.message, SHUTDOWN_TASK_MESSAGE);
    }

    #[tokio::test]
    async fn generating_that_finishes_within_grace_keeps_its_own_outcome() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        let (id, _, _) = state.scheduler.submit(request("fail:200")).await.unwrap();
        wait_status(&state, &id, TaskStatus::Generating).await;

        let t0 = Instant::now();
        let report = graceful_shutdown(&state, &timing(5_000)).await;
        assert!(
            t0.elapsed() < Duration::from_secs(3),
            "任务结束后立即返回,不等满宽限期:{:?}",
            t0.elapsed()
        );
        assert!(report.forced.is_empty());
        let t = task(&state, &id).await;
        assert_eq!(t.status, TaskStatus::Failed);
        assert_eq!(t.error.unwrap().code, "generation_doubao_refused");
    }

    #[tokio::test]
    async fn submit_is_refused_after_shutdown_and_shutdown_is_idempotent() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        let first = graceful_shutdown(&state, &timing(100)).await;
        let e = state.submit_task(request("gen")).await.unwrap_err();
        assert_eq!(e.code, "system_shutdown");
        let e = crate::commands::task::submit(
            &state,
            serde_json::json!({
                "account": "张三", "type": "image", "modelName": "Seedream 5.0 Pro",
                "prompt": "p", "ratio": "1:1"
            }),
        )
        .await
        .unwrap_err();
        assert_eq!(e.code, "system_shutdown");
        // 再调一次返回同一份报告,不重复执行
        let second = graceful_shutdown(&state, &timing(100)).await;
        assert_eq!(first, second);
    }

    #[tokio::test]
    async fn close_resources_disconnects_cdp_and_closes_db() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        graceful_shutdown(&state, &timing(100)).await;
        close_resources(&state).await;
        assert!(state.doubao.is_shutting_down());
        assert!(!state.doubao.heartbeat().await);
        assert!(state.storage.pool.is_closed());
    }

    #[tokio::test]
    async fn leftover_non_terminal_rows_are_marked_failed() {
        let dir = tempfile::tempdir().unwrap();
        let state = state_with_loop(dir.path()).await;
        // 库里有一条调度器不知道的非终态任务(例如 runner 卡死):兜底直接标失败
        let (id, _, _) = state.scheduler.submit(request("hang")).await.unwrap();
        wait_status(&state, &id, TaskStatus::Preparing).await;
        let mut ghost = task(&state, &id).await;
        ghost.id = "01GHOSTGHOSTGHOSTGHOSTGHOS".into();
        ghost.status = TaskStatus::Generating;
        state.storage.insert_task(&ghost).await.unwrap();

        let report = graceful_shutdown(&state, &timing(100)).await;
        assert_eq!(report.marked_in_db, 1);
        let g = task(&state, "01GHOSTGHOSTGHOSTGHOSTGHOS").await;
        assert_eq!(g.status, TaskStatus::Failed);
        assert_eq!(g.error.unwrap().code, "system_shutdown");
    }
}
