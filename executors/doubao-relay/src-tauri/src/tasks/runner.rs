//! 任务运行器(Task 2.7 图像路径 + Task 3.3 视频分支):一个任务从 `pending` 被调度器挑中之后的全部生命周期。
//! 视频:generating 预算 20min、prompt 原样(视频走"视频生成"技能,Task 2.1 结论不加前缀)、
//! 只下载第一条无水印原片(`download_video`,非 `lr=unwatermarked` → `generation_watermark_not_removed`)、
//! `duration_ms` / `width` / `height` 由 `download::mp4_file_info` 解析 `mvhd` 与视频轨 `tkhd` 得出;
//! 下载与取消竞争,取消时删除半截文件并以 `cancelled`(step=download)结束。
//!
//! 流程(`docs/spec/08-task-lifecycle.md`):
//! `preparing`(若当前账号 ≠ 任务账号 → 切换;`generation::submit`,整段 60s 超时 →
//! `generation_preparing_timeout`)→ `generating/sent` → `capture`(`CaptureEvent` 映射为
//! 状态 / 阶段 / DB / 事件;`NeedsUserConfirm` → `waiting_user_confirm`,用户 `confirm` 后回到
//! `generating`,1h 无确认 → `generation_confirm_timeout`)→ `downloading`(每张原图 ≤20 张)
//! → `completed`;任何 `Err` → `failed`;取消 → `cancelled`。
//!
//! 超时归属:`capture` 自己的 deadline 在等待用户确认期间照样走(Task 2.4 的实现),
//! 所以 spec 的"generating 图 5min / 视频 20min(不含等待确认)"与"waiting_user_confirm 1h"
//! 由 runner 自己计时([`RunCtx::next_deadline`]),传给 `capture` 的 deadline 只是一个宽松上限
//! (生成预算 + 两次确认等待)。
//!
//! 所有状态变化:**DB 先写(`update_task` + `append_stage`)再广播** [`TaskEvent`];
//! 日志 info 字段 `{task_id, from_status, to_status, stage, duration_ms, prompt_length, prompt_hash}`,
//! 不出现 prompt 原文。
//!
//! 豆包侧依赖抽象为 [`GenerationBackend`](submit / capture / download 可 mock),
//! 生产实现 [`DoubaoBackend`] 包一个 `DoubaoConnection` + `reqwest::Client`。

use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use async_trait::async_trait;
use serde_json::{json, Value};
use tokio::sync::{mpsc, watch};
use tokio_util::sync::CancellationToken;

use crate::cdp::CdpSession;
use crate::doubao::capture::{self, CaptureEvent, CaptureOptions, CaptureOutcome};
use crate::doubao::download::{self, Downloaded};
use crate::doubao::generation::{self, StageFn, SubmissionPlan, SubmissionReport};
use crate::doubao::{AccountList, DoubaoConnection};
use crate::error::{RelayError, RelayResult};
use crate::logging::now_ms;
use crate::storage::Storage;
use crate::tasks::scheduler::{Scheduler, TaskEvent, TaskHandle};
use crate::tasks::{
    prompt_hash, GenerateRequest, ResultFile, StageEntry, Task, TaskError, TaskStage, TaskStatus,
    TaskType,
};

/// 一个任务最多下载的结果文件数。
pub const MAX_RESULT_FILES: usize = 20;

/// generating 期间发现豆包输入框被用户改动时的 `task_updated.message`(spec 08 `user_modified_input`,
/// 只提示不终止)。
pub const INPUT_MODIFIED_MESSAGE: &str = "检测到输入框被修改";

/// 豆包侧操作的抽象:调度器校验账号、运行器提交 / 捕获 / 下载都经它,测试用 mock 注入。
#[async_trait]
pub trait GenerationBackend: Send + Sync {
    /// 能否接受新任务(`POST /generate` 入队前):豆包崩溃 / 未连接 → `doubao_not_ready`(503)。
    /// 默认就绪(测试替身不必实现)。
    async fn ready(&self) -> RelayResult<()> {
        Ok(())
    }

    /// 账号列表(`force` 跳过缓存)。
    async fn accounts(&self, force: bool) -> RelayResult<AccountList>;

    /// 确保豆包当前账号是 `account`(不是则切换;切换失败按 `account_switch_*` 返回)。
    async fn ensure_account(&self, account: &str) -> RelayResult<()>;

    /// 退出残留技能 → prepare → fill → checkSend → send → identity(`generation::submit`)。
    async fn submit(
        &self,
        request: &GenerateRequest,
        plan: &SubmissionPlan,
        reference_paths: &[PathBuf],
        cancel: &CancellationToken,
        on_stage: StageFn<'_>,
    ) -> RelayResult<SubmissionReport>;

    /// 捕获循环(`capture::capture`)。
    async fn capture(
        &self,
        opts: CaptureOptions,
        cancel: &CancellationToken,
        confirm_rx: &mut watch::Receiver<bool>,
        events: mpsc::Sender<CaptureEvent>,
    ) -> RelayResult<CaptureOutcome>;

    /// 下载一张原图到 `dest_dir`(`download::download_image`)。
    async fn download_image(
        &self,
        url: &str,
        dest_dir: &Path,
        index: u32,
    ) -> RelayResult<Downloaded>;

    /// 下载无水印视频到 `dest_dir/video_1.mp4`(`download::download_video`;
    /// URL 不含 `lr=unwatermarked` → `generation_watermark_not_removed`)。
    async fn download_video(&self, url: &str, dest_dir: &Path) -> RelayResult<Downloaded>;
}

/// 生产实现:真豆包连接 + 下载客户端。
pub struct DoubaoBackend {
    conn: Arc<DoubaoConnection>,
    client: reqwest::Client,
}

impl DoubaoBackend {
    pub fn new(conn: Arc<DoubaoConnection>) -> Arc<Self> {
        Arc::new(Self {
            conn,
            client: download::new_client(),
        })
    }

    pub fn connection(&self) -> &Arc<DoubaoConnection> {
        &self.conn
    }
}

impl DoubaoBackend {
    /// 断线类错误时查一次豆包进程:进程已不在 → `doubao_crashed{last_seen_at}`(UC-32;
    /// capture 发现 chat page 消失约 6s 就返回 `doubao_cdp_disconnected`,早于连接层 20s 的 `Crashed` 事件)。
    async fn crash_aware(&self, e: RelayError) -> RelayError {
        if !is_disconnect(&e) {
            return e;
        }
        let running = self.conn.is_doubao_running().await;
        crash_aware_error(e, running, self.conn.last_seen_at())
    }
}

/// `doubao_cdp_disconnected` / `doubao_not_ready` / `doubao_not_running` 视为可能的崩溃。
fn is_disconnect(e: &RelayError) -> bool {
    matches!(
        e.code,
        "doubao_cdp_disconnected" | "doubao_not_ready" | "doubao_not_running"
    )
}

fn should_retry_pre_submission_disconnect(error: &RelayError, sent: bool, retry_count: u8) -> bool {
    error.code == "doubao_cdp_disconnected" && !sent && retry_count == 0
}

const PREPARING_TIMEOUT_RETRY_LIMIT: u8 = 2;

fn should_retry_preparing_timeout(error: &RelayError, retry_count: u8) -> bool {
    if error.code != "generation_preparing_timeout" || retry_count >= PREPARING_TIMEOUT_RETRY_LIMIT
    {
        return false;
    }
    error
        .details
        .as_ref()
        .and_then(|details| details.get("sent"))
        .and_then(Value::as_bool)
        == Some(false)
}

async fn invalidate_submission_session_on_disconnect(session: &CdpSession, error: &RelayError) {
    if error.code == "doubao_cdp_disconnected" {
        session.close().await;
    }
}

/// 纯函数:断线类错误 + 豆包进程不在 → `doubao_crashed{last_seen_at, cause}`;其它原样返回。
pub fn crash_aware_error(e: RelayError, doubao_running: bool, last_seen_at: i64) -> RelayError {
    if doubao_running || !is_disconnect(&e) {
        return e;
    }
    let mut crashed = RelayError::doubao_crashed(last_seen_at);
    if let Some(Value::Object(d)) = crashed.details.as_mut() {
        d.insert("cause".into(), Value::String(e.code.to_string()));
    }
    crashed
}

#[async_trait]
impl GenerationBackend for DoubaoBackend {
    async fn ready(&self) -> RelayResult<()> {
        self.conn.ensure_ready().await
    }

    async fn accounts(&self, force: bool) -> RelayResult<AccountList> {
        self.conn.accounts(force).await
    }

    /// 先读取页面当前账号;页面未显示账号时强制刷新账号列表,不允许用旧缓存继续提交。
    /// 不一致才 `switch_account`(它自己会再读一次列表并处理"目标已是当前账号")。
    async fn ensure_account(&self, account: &str) -> RelayResult<()> {
        // 任务刚启动时账号列表缓存可能还没有建立。此时不能先读账号菜单：
        // Windows 豆包有时会拦截账号菜单，导致任务在提交前白等 120 秒。
        // 页面上的当前账号是无菜单的轻量探测；已确认一致就直接继续提交。
        if let Ok(Some(current)) = self.conn.current_account().await {
            if current == account {
                tracing::debug!(target = account, "任务账号已是页面当前账号,跳过账号菜单");
                return Ok(());
            }
        }
        // 任务真正开始前必须再次确认账号。提交队列阶段的账号缓存可能已经过期,
        // 特别是用户退出豆包或切换账号后,继续使用旧缓存会把任务伪装成 generating。
        let list = self.conn.accounts(true).await?;
        if list.current.as_deref() == Some(account) {
            return Ok(());
        }
        tracing::info!(
            target = account,
            current = list.current.as_deref().unwrap_or(""),
            "任务账号与当前账号不一致,切换账号"
        );
        self.conn.switch_account(account).await.map(|_| ())
    }

    async fn submit(
        &self,
        request: &GenerateRequest,
        plan: &SubmissionPlan,
        reference_paths: &[PathBuf],
        cancel: &CancellationToken,
        on_stage: StageFn<'_>,
    ) -> RelayResult<SubmissionReport> {
        // 提交全程独占 chat 页面:账号菜单读取 / 切换与技能 / 模型菜单、浮层、上传互斥
        let _page = self.conn.lock_page().await;
        // spec 08 任务冲突检测:多个 chat 窗口时在填写 / 发送之前就拒绝(不耗配额)
        let session = match self.conn.submit_session().await {
            Ok(s) => s,
            Err(e) => return Err(self.crash_aware(e).await),
        };
        let background = self.conn.background_session().await;
        match generation::submit(
            &session,
            background.as_deref(),
            request,
            plan,
            reference_paths,
            cancel,
            on_stage,
        )
        .await
        {
            Ok(r) => Ok(r),
            Err(e) => {
                invalidate_submission_session_on_disconnect(&session, &e).await;
                Err(self.crash_aware(e).await)
            }
        }
    }

    async fn capture(
        &self,
        opts: CaptureOptions,
        cancel: &CancellationToken,
        confirm_rx: &mut watch::Receiver<bool>,
        events: mpsc::Sender<CaptureEvent>,
    ) -> RelayResult<CaptureOutcome> {
        // 捕获期间(含代回复的 fill + 发送)页面忙:非强制账号读取沿用缓存,不打开账号菜单
        let _busy = self.conn.mark_page_busy();
        match capture::capture(&self.conn, opts, cancel, confirm_rx, events).await {
            Ok(o) => Ok(o),
            Err(e) => Err(self.crash_aware(e).await),
        }
    }

    async fn download_image(
        &self,
        url: &str,
        dest_dir: &Path,
        index: u32,
    ) -> RelayResult<Downloaded> {
        download::download_image(&self.client, url, dest_dir, index).await
    }

    async fn download_video(&self, url: &str, dest_dir: &Path) -> RelayResult<Downloaded> {
        download::download_video(&self.client, url, dest_dir).await
    }
}

// ------------------------------------------------------------------
// 错误(11-error-codes `generation_*` / `system_*`)
// ------------------------------------------------------------------

/// `generation_confirm_timeout`(manual 模式等待用户确认超时)。
pub fn confirm_timeout_error(waited: Duration, stage: Option<&str>) -> RelayError {
    RelayError::new(
        "generation_confirm_timeout",
        "等待用户确认超时,请重新提交任务并及时确认",
    )
    .with_details(json!({ "waited_ms": waited.as_millis() as u64, "stage": stage }))
}

/// `system_task_panicked`。
pub fn task_panicked_error() -> RelayError {
    RelayError::new(
        "system_task_panicked",
        "任务执行线程异常退出,请重新提交任务并查看日志",
    )
}

/// `RelayError` → 持久化的 `TaskError`;`account_quota_exhausted` 补 `details.account`
/// (capture 不知道账号,spec 11 要求 `{account}`)。
pub fn task_error_for(e: &RelayError, account: &str) -> TaskError {
    let mut details = e.details.clone();
    if e.code == "account_quota_exhausted" {
        let mut obj = match details.take() {
            Some(Value::Object(m)) => m,
            _ => serde_json::Map::new(),
        };
        obj.insert("account".into(), Value::String(account.to_string()));
        details = Some(Value::Object(obj));
    }
    TaskError {
        code: e.code.to_string(),
        message: e.message.clone(),
        details,
    }
}

/// 取消时写入 `task.error` 的 `generation_cancelled{from_status, step?}`。
pub fn cancelled_error(from: TaskStatus, step: Option<&str>) -> TaskError {
    TaskError {
        code: "generation_cancelled".into(),
        message: "任务已取消".into(),
        details: Some(json!({ "from_status": from.as_str(), "step": step })),
    }
}

/// `results/{task_id}/{filename}`(07-data-model `relative_path`,`/` 分隔)。
pub fn result_relative_path(task_id: &str, filename: &str) -> String {
    format!("results/{task_id}/{filename}")
}

/// `Downloaded` → `ResultFile`。
pub fn result_file_from(task_id: &str, index: u32, d: &Downloaded, source_url: &str) -> ResultFile {
    let filename = d
        .path
        .file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_else(|| format!("image_{}", index + 1));
    ResultFile {
        index,
        relative_path: result_relative_path(task_id, &filename),
        filename,
        mime: d.mime.clone(),
        bytes: d.bytes,
        width: d.width,
        height: d.height,
        duration_ms: None,
        source_url: Some(source_url.to_string()),
    }
}

// ------------------------------------------------------------------
// 运行
// ------------------------------------------------------------------

/// 运行一个已被调度器挑中的任务,直到终态。永不 panic 传播到调度器(调度器另有 JoinError 兜底)。
pub async fn run(sched: Arc<Scheduler>, task_id: String, handle: TaskHandle) {
    let task = match sched.storage().get_task(&task_id).await {
        Ok(Some(t)) => t,
        Ok(None) => {
            tracing::error!(task_id = %task_id, "调度到的任务在数据库中不存在");
            return;
        }
        Err(e) => {
            tracing::error!(task_id = %task_id, code = e.code, error = %e, "读取任务失败,放弃运行");
            return;
        }
    };
    if task.status != TaskStatus::Pending {
        tracing::warn!(task_id = %task_id, status = task.status.as_str(), "任务不在 pending 状态,跳过");
        return;
    }
    let mut ctx = RunCtx::new(sched, task, handle);
    let outcome = ctx.execute().await;
    ctx.finish(outcome).await;
}

/// 一次运行的可变上下文。
struct RunCtx {
    sched: Arc<Scheduler>,
    task: Task,
    handle: TaskHandle,
    last_change: Instant,
    /// 本次中断的原因(超时等),capture 返回 `generation_cancelled` 时用它替换。
    abort: Option<RelayError>,
}

/// 成功结果:原图 URL 已全部下载。
struct Finished {
    files: Vec<ResultFile>,
    outcome: CaptureOutcome,
}

impl RunCtx {
    fn new(sched: Arc<Scheduler>, task: Task, handle: TaskHandle) -> Self {
        Self {
            sched,
            task,
            handle,
            last_change: Instant::now(),
            abort: None,
        }
    }

    fn storage(&self) -> &Storage {
        self.sched.storage()
    }

    fn account(&self) -> &str {
        &self.task.request.account
    }

    // ---- 状态变更(DB 先写再广播) ----

    /// 改状态 / 阶段 / 阶段消息:`update_task` → `append_stage`(阶段变了记阶段名,否则记状态名)
    /// → `TaskEvent::Updated`。状态、阶段都没变只更新消息时不追加历史。
    async fn transition(
        &mut self,
        status: TaskStatus,
        stage: Option<TaskStage>,
        message: Option<String>,
    ) -> RelayResult<()> {
        let from_status = self.task.status;
        let from_stage = self.task.stage;
        let now = now_ms();
        let status_changed = from_status != status;
        let stage_changed = stage.is_some() && from_stage != stage;

        self.task.status = status;
        if stage.is_some() {
            self.task.stage = stage;
        }
        if status_changed && status == TaskStatus::Preparing {
            self.task.started_at = Some(now);
        }
        if status == TaskStatus::Completed {
            self.task.stage = None;
        }
        if status.is_terminal() {
            self.task.completed_at = Some(now);
        }
        if message.is_some() || stage_changed {
            self.task.stage_message = message.clone();
        }
        *self
            .handle
            .status
            .write()
            .unwrap_or_else(|p| p.into_inner()) = status;

        self.storage().update_task(&self.task).await?;
        let entry_name = if stage_changed {
            self.task.stage.map(|s| s.as_str())
        } else if status_changed {
            Some(status.as_str())
        } else {
            None
        };
        if let Some(name) = entry_name {
            let entry_message = if status.is_terminal() && status != TaskStatus::Completed {
                self.task.error.as_ref().map(|e| e.message.clone())
            } else {
                self.task.stage_message.clone()
            };
            self.storage()
                .append_stage(&self.task.id, name, entry_message.as_deref(), now)
                .await?;
            self.task.stage_history.push(StageEntry {
                stage: name.to_string(),
                message: entry_message,
                ts: now,
            });
        }

        if status_changed || stage_changed {
            tracing::info!(
                task_id = %self.task.id,
                from_status = from_status.as_str(),
                to_status = status.as_str(),
                stage = self.task.stage.map(|s| s.as_str()).unwrap_or(""),
                duration_ms = self.last_change.elapsed().as_millis() as u64,
                prompt_length = self.task.request.prompt.chars().count(),
                prompt_hash = %prompt_hash(&self.task.request.prompt),
                "任务状态变化"
            );
            self.last_change = Instant::now();
        }
        self.sched.emit(TaskEvent::Updated {
            task_id: self.task.id.clone(),
            status,
            stage: self.task.stage,
            message: self.task.stage_message.clone(),
            progress: self.task.progress,
            updated_at: now,
        });
        if status.is_terminal() {
            // DB 已写(含终态 stage_history)→ 回调(Task 4.1)→ 广播
            self.sched.notify_terminal(&self.task);
            self.sched.emit(TaskEvent::Completed(self.task.clone()));
        }
        Ok(())
    }

    async fn set_stage(&mut self, stage: TaskStage, message: Option<String>) -> RelayResult<()> {
        let status = self.task.status;
        self.transition(status, Some(stage), message).await
    }

    // ---- 主流程 ----

    async fn execute(&mut self) -> RelayResult<Finished> {
        self.check_cancelled("start")?;
        let mut preparing_timeout_retries = 0;
        let report = loop {
            match self.preparing().await {
                Ok(report) => break report,
                Err(error)
                    if !self.handle.cancel.is_cancelled()
                        && should_retry_preparing_timeout(&error, preparing_timeout_retries) =>
                {
                    preparing_timeout_retries += 1;
                    self.set_stage(
                        TaskStage::Preparing,
                        Some(format!(
                            "生成准备阶段超时，未发送请求，正在自动重试 {}/{}",
                            preparing_timeout_retries, PREPARING_TIMEOUT_RETRY_LIMIT
                        )),
                    )
                    .await?;
                    tokio::time::sleep(Duration::from_secs(1)).await;
                }
                Err(error) => return Err(error),
            }
        };
        let identity = report
            .identity
            .clone()
            .ok_or_else(|| generation::identity_missing(report.send_started_at, 0))?;
        self.task.local_message_id = identity.local_message_id.clone();
        if self.task.conversation_id.is_none() {
            self.task.conversation_id = identity.conversation_id.clone();
        }
        self.transition(TaskStatus::Generating, Some(TaskStage::Sent), None)
            .await?;
        let outcome = self.generating(identity, &report).await?;
        self.check_cancelled("download")?;
        let files = self.downloading(&outcome).await?;
        Ok(Finished { files, outcome })
    }

    fn check_cancelled(&self, step: &str) -> RelayResult<()> {
        if self.handle.cancel.is_cancelled() {
            Err(generation::cancelled(step))
        } else {
            Ok(())
        }
    }

    /// `preparing`:切账号 + `submit`,整段受 `preparing_timeout` 约束;阶段回调即时落库。
    async fn preparing(&mut self) -> RelayResult<SubmissionReport> {
        self.transition(TaskStatus::Preparing, Some(TaskStage::Preparing), None)
            .await?;
        let request = self.task.request.clone();
        let plan = SubmissionPlan::for_request(&request, &self.task.id);
        let reference_paths = self.resolve_reference_paths().await?;
        let backend = self.sched.backend();
        let prep_cancel = self.handle.cancel.child_token();
        let timeout = self.sched.timing().preparing_timeout;
        let (stage_tx, mut stage_rx) = mpsc::unbounded_channel::<(TaskStage, Option<String>)>();
        let sent = Arc::new(AtomicBool::new(false));
        let sent_for_stage = Arc::clone(&sent);
        let on_stage = move |stage: TaskStage, message: Option<String>| {
            if stage == TaskStage::Sent {
                sent_for_stage.store(true, Ordering::Release);
            }
            let _ = stage_tx.send((stage, message));
        };

        let account = request.account.clone();
        let submit_fut = async {
            on_stage(
                TaskStage::Preparing,
                Some(format!("正在确认豆包账号：{account}")),
            );
            backend.ensure_account(&account).await?;
            let mut retry_count = 0u8;
            loop {
                sent.store(false, Ordering::Release);
                let result = backend
                    .submit(&request, &plan, &reference_paths, &prep_cancel, &on_stage)
                    .await;
                match result {
                    Err(error)
                        if !prep_cancel.is_cancelled()
                            && should_retry_pre_submission_disconnect(
                                &error,
                                sent.load(Ordering::Acquire),
                                retry_count,
                            ) =>
                    {
                        retry_count += 1;
                        tracing::warn!(
                            retry_count,
                            "提交前豆包连接关闭,重新取得连接并重试准备阶段"
                        );
                        on_stage(
                            TaskStage::Preparing,
                            Some("豆包连接已恢复，正在重新准备".into()),
                        );
                    }
                    other => break other,
                }
            }
        };
        tokio::pin!(submit_fut);
        let deadline = tokio::time::sleep(timeout);
        tokio::pin!(deadline);
        let mut timed_out = false;
        let result = loop {
            tokio::select! {
                res = &mut submit_fut => break res,
                Some((stage, message)) = stage_rx.recv() => {
                    // 提交阶段的细分进度只更新 stage_message(stage 仍是 preparing)
                    let _ = stage;
                    self.set_stage(TaskStage::Preparing, message).await?;
                }
                _ = &mut deadline, if !timed_out => {
                    timed_out = true;
                    tracing::warn!(task_id = %self.task.id, timeout_ms = timeout.as_millis() as u64, "preparing 超时,中断提交");
                    prep_cancel.cancel();
                    // 让 submit 做完清空输入框等收尾,但不再无限等
                    match tokio::time::timeout(Duration::from_secs(5), &mut submit_fut).await {
                        Ok(res) => break res,
                        Err(_) => break Err(generation::cancelled("preparing_timeout")),
                    }
                }
            }
        };
        // 回调里可能还有没处理完的阶段消息
        while let Ok((_, message)) = stage_rx.try_recv() {
            if message.is_some() && !timed_out {
                self.set_stage(TaskStage::Preparing, message).await?;
            }
        }
        match result {
            Ok(report) => Ok(report),
            Err(e) if timed_out && e.code == "generation_cancelled" => {
                Err(generation::preparing_timeout("submit").with_details(json!({
                    "sent": sent.load(Ordering::Acquire)
                })))
            }
            Err(e) if timed_out && !self.handle.cancel.is_cancelled() => {
                // 超时后 submit 自己报了别的错(通常也是被打断),以超时为准
                tracing::debug!(code = e.code, "preparing 超时后 submit 返回错误");
                Err(generation::preparing_timeout("submit").with_details(json!({
                    "sent": sent.load(Ordering::Acquire)
                })))
            }
            Err(e) => Err(e),
        }
    }

    /// `reference_files`(upload_id)→ 本地路径;不存在 / 已过期按 `upload_*` 失败。
    async fn resolve_reference_paths(&self) -> RelayResult<Vec<PathBuf>> {
        let now = now_ms();
        let mut paths = Vec::with_capacity(self.task.request.reference_files.len());
        for id in &self.task.request.reference_files {
            let upload = self
                .storage()
                .get_upload(id)
                .await?
                .ok_or_else(|| crate::tasks::scheduler::upload_not_found(id))?;
            if upload.expires_at < now {
                return Err(crate::tasks::scheduler::upload_expired(
                    id,
                    upload.expires_at,
                ));
            }
            paths.push(upload.path);
        }
        Ok(paths)
    }

    /// `generating`:驱动 `capture`,把事件映射为状态;自己计"生成预算"(不含等待确认)与"确认等待"。
    async fn generating(
        &mut self,
        identity: generation::SubmissionIdentity,
        report: &SubmissionReport,
    ) -> RelayResult<CaptureOutcome> {
        let timing = self.sched.timing().clone();
        let budget = match self.task.request.task_type {
            TaskType::Image => timing.image_timeout,
            TaskType::Video => timing.video_timeout,
        };
        let opts = CaptureOptions {
            task_type: self.task.request.task_type,
            prompt: self.task.request.prompt.clone(),
            model_name: Some(self.task.request.model_name.clone()),
            ratio: Some(self.task.request.ratio.clone()),
            duration: self.task.request.duration,
            reference_count: self.task.request.reference_files.len(),
            identity,
            send_since: report.send_started_at,
            generation_requests: report.generation_requests.clone(),
            confirmation_mode: self.task.request.confirmation_mode,
            auto_use_membership: self.task.request.auto_use_membership,
            tracking_key: generation::tracking_key_for(&self.task.id),
            deadline: budget + timing.confirm_timeout * 2,
        };
        let backend = self.sched.backend();
        let cap_cancel = self.handle.cancel.child_token();
        let mut confirm_rx = self.handle.confirm_tx.subscribe();
        let mut my_confirm = self.handle.confirm_tx.subscribe();
        let (ev_tx, mut ev_rx) = mpsc::channel::<CaptureEvent>(64);

        let cap_fut = backend.capture(opts, &cap_cancel, &mut confirm_rx, ev_tx);
        tokio::pin!(cap_fut);

        let mut clock = PhaseClock::new(budget, timing.confirm_timeout);
        let result = loop {
            let deadline = clock.next_deadline();
            // biased:用户确认与 capture 紧接着发出的 Stage 事件常同时就绪,
            // 先处理确认边沿,保证 waiting_user_confirm → generating 作为独立一步落库。
            tokio::select! {
                biased;
                changed = my_confirm.changed() => {
                    if changed.is_ok()
                        && *my_confirm.borrow_and_update()
                        && self.task.status == TaskStatus::WaitingUserConfirm
                    {
                        tracing::info!(task_id = %self.task.id, "用户已确认,回到 generating");
                        clock.resume();
                        self.transition(TaskStatus::Generating, None, None).await?;
                    }
                }
                Some(ev) = ev_rx.recv() => {
                    if let Err(e) = self.on_capture_event(ev, &mut clock).await {
                        cap_cancel.cancel();
                        break Err(e);
                    }
                }
                res = &mut cap_fut => break res,
                _ = tokio::time::sleep_until(deadline.into()), if self.abort.is_none() => {
                    let err = if clock.waiting() {
                        confirm_timeout_error(clock.waited(), self.task.stage.map(|s| s.as_str()))
                    } else {
                        capture::task_timeout_error(budget, self.task.stage.map(|s| s.as_str()))
                    };
                    tracing::warn!(task_id = %self.task.id, code = err.code, "generating 阶段超时,中断捕获");
                    self.abort = Some(err);
                    cap_cancel.cancel();
                    // 捕获实现可能不响应取消信号。不能继续等待 cap_fut，
                    // 否则 runner 永远不结束，当前账号锁也不会释放，后续任务会全部排队。
                    break Err(self.abort.clone().expect("timeout error must be recorded"));
                }
            }
        };
        // capture 返回前发出、还没处理的事件(Stage(result) / Bound 等)
        while let Ok(ev) = ev_rx.try_recv() {
            if let Err(e) = self.on_capture_event(ev, &mut clock).await {
                tracing::warn!(task_id = %self.task.id, code = e.code, "收尾处理捕获事件失败");
            }
        }
        match result {
            Err(e) if e.code == "generation_cancelled" => {
                if let Some(abort) = self.abort.take() {
                    Err(abort)
                } else {
                    Err(e)
                }
            }
            other => other,
        }
    }

    async fn on_capture_event(
        &mut self,
        ev: CaptureEvent,
        clock: &mut PhaseClock,
    ) -> RelayResult<()> {
        match ev {
            CaptureEvent::Stage(stage, message) => {
                if self.task.status == TaskStatus::WaitingUserConfirm {
                    // 确认后豆包已经往下走了(或 auto 降级又恢复):回到 generating
                    clock.resume();
                    self.transition(TaskStatus::Generating, Some(stage), message)
                        .await
                } else {
                    self.set_stage(stage, message).await
                }
            }
            CaptureEvent::Bound {
                conversation_id,
                submission_id,
            } => {
                self.task.conversation_id = Some(conversation_id);
                self.task.submission_id = Some(submission_id);
                self.storage().update_task(&self.task).await
            }
            CaptureEvent::NeedsUserConfirm => {
                if self.task.status != TaskStatus::WaitingUserConfirm {
                    clock.pause();
                    self.transition(TaskStatus::WaitingUserConfirm, None, None)
                        .await
                } else {
                    Ok(())
                }
            }
            CaptureEvent::Refused { kind, message } => self
                .set_stage(TaskStage::Refused, Some(message))
                .await
                .map(|_| {
                    tracing::warn!(task_id = %self.task.id, kind = %kind, "豆包拒绝生成");
                }),
            CaptureEvent::Feedback(state) => {
                tracing::debug!(task_id = %self.task.id, stage = %state.stage, "feedback");
                Ok(())
            }
            CaptureEvent::Originals(urls) => {
                tracing::debug!(task_id = %self.task.id, count = urls.len(), "识别到原图");
                Ok(())
            }
            CaptureEvent::InputModified { text_length } => {
                tracing::warn!(
                    task_id = %self.task.id,
                    text_length,
                    "检测到豆包输入框被修改,继续等待结果"
                );
                let status = self.task.status;
                self.transition(status, None, Some(INPUT_MODIFIED_MESSAGE.to_string()))
                    .await
            }
            CaptureEvent::VideoCandidates(_) | CaptureEvent::Fallbacks(_) => Ok(()),
        }
    }

    /// `downloading`:图像逐张下载(≤ [`MAX_RESULT_FILES`]),每张之间检查取消;
    /// 视频只下第一条无水印原片(沈浪 `videoPrimaryURL`)并解析 mp4 时长。
    async fn downloading(&mut self, outcome: &CaptureOutcome) -> RelayResult<Vec<ResultFile>> {
        self.set_stage(TaskStage::Downloading, None).await?;
        if outcome.originals.is_empty() {
            return Err(RelayError::new(
                "generation_result_not_captured",
                "任务结束但未捕获到结果,请稍后重试",
            ));
        }
        let dest_dir = self.storage().result_dir(&self.task.id);
        tokio::fs::create_dir_all(&dest_dir).await?;
        let backend = self.sched.backend();
        if self.task.request.task_type == TaskType::Video {
            self.check_cancelled("download")?;
            // 视频是单个长下载(≤180s),必须和取消竞争;drop 掉下载 future 会跳过
            // `download_video` 自己的清理,所以取消时这里删掉半截 `video_1.mp4`。
            let cancel = self.handle.cancel.clone();
            let d = tokio::select! {
                biased;
                _ = cancel.cancelled() => {
                    let _ = tokio::fs::remove_file(dest_dir.join("video_1.mp4")).await;
                    return Err(generation::cancelled("download"));
                }
                r = backend.download_video(&outcome.originals[0], &dest_dir) => r?,
            };
            if let Err(e) = self.check_cancelled("download") {
                let _ = tokio::fs::remove_file(&d.path).await;
                return Err(e);
            }
            let mut file = result_file_from(&self.task.id, 0, &d, &outcome.originals[0]);
            let info = download::mp4_file_info(&d.path).await.unwrap_or_default();
            file.duration_ms = info.duration_ms;
            file.width = info.width;
            file.height = info.height;
            tracing::info!(
                task_id = %self.task.id,
                bytes = d.bytes,
                duration_ms = file.duration_ms,
                watermark_source = outcome.watermark_source.as_str(),
                "无水印视频已下载"
            );
            return Ok(vec![file]);
        }
        let mut files = Vec::new();
        for (i, url) in outcome.originals.iter().take(MAX_RESULT_FILES).enumerate() {
            self.check_cancelled("download")?;
            let index = i as u32;
            let d = backend.download_image(url, &dest_dir, index).await?;
            files.push(result_file_from(&self.task.id, index, &d, url));
            tracing::info!(
                task_id = %self.task.id,
                index,
                bytes = d.bytes,
                mime = %d.mime,
                "结果文件已下载"
            );
        }
        Ok(files)
    }

    // ---- 终态 ----

    async fn finish(&mut self, outcome: RelayResult<Finished>) {
        let result = match outcome {
            Ok(done) => {
                self.task.result_files = Some(done.files);
                self.task.watermark_source = Some(done.outcome.watermark_source);
                if let Some(c) = done.outcome.conversation_id {
                    self.task.conversation_id = Some(c);
                }
                if let Some(s) = done.outcome.submission_id {
                    self.task.submission_id = Some(s);
                }
                self.task.error = None;
                self.transition(TaskStatus::Completed, None, None).await
            }
            Err(e) => {
                let from = self.task.status;
                let forced = self
                    .handle
                    .abort
                    .lock()
                    .unwrap_or_else(|p| p.into_inner())
                    .clone();
                if let Some(forced) = forced {
                    // 调度器 fail_all_running(豆包崩溃等)
                    self.task.error = Some(forced);
                    self.transition(TaskStatus::Failed, None, None).await
                } else if e.code == "generation_cancelled" {
                    let step = e
                        .details
                        .as_ref()
                        .and_then(|d| d.get("step"))
                        .and_then(Value::as_str)
                        .map(str::to_string);
                    self.task.error = Some(cancelled_error(from, step.as_deref()));
                    self.transition(TaskStatus::Cancelled, None, None).await
                } else {
                    tracing::warn!(
                        task_id = %self.task.id,
                        code = e.code,
                        error = %e,
                        details = ?e.details,
                        "任务失败"
                    );
                    self.task.error = Some(task_error_for(&e, self.account()));
                    self.transition(TaskStatus::Failed, None, None).await
                }
            }
        };
        if let Err(e) = result {
            tracing::error!(task_id = %self.task.id, code = e.code, error = %e, "写入任务终态失败");
        }
    }
}

/// generating 阶段的双计时:生成预算只在非等待期间消耗,等待确认单独计。
struct PhaseClock {
    budget: Duration,
    confirm_timeout: Duration,
    active_since: Instant,
    used: Duration,
    waiting_since: Option<Instant>,
}

impl PhaseClock {
    fn new(budget: Duration, confirm_timeout: Duration) -> Self {
        Self {
            budget,
            confirm_timeout,
            active_since: Instant::now(),
            used: Duration::ZERO,
            waiting_since: None,
        }
    }

    fn waiting(&self) -> bool {
        self.waiting_since.is_some()
    }

    fn waited(&self) -> Duration {
        self.waiting_since.map(|w| w.elapsed()).unwrap_or_default()
    }

    fn pause(&mut self) {
        if self.waiting_since.is_none() {
            self.used += self.active_since.elapsed();
            self.waiting_since = Some(Instant::now());
        }
    }

    fn resume(&mut self) {
        if self.waiting_since.take().is_some() {
            self.active_since = Instant::now();
        }
    }

    fn next_deadline(&self) -> Instant {
        match self.waiting_since {
            Some(w) => w + self.confirm_timeout,
            None => self.active_since + self.budget.saturating_sub(self.used),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn disconnect_with_doubao_gone_becomes_doubao_crashed() {
        let e = crash_aware_error(RelayError::doubao_cdp_disconnected(), false, 42);
        assert_eq!(e.code, "doubao_crashed");
        // 与崩溃监听(lib::spawn_crash_watcher)写入的文案一致
        assert_eq!(e.message, crate::DOUBAO_CRASHED_MESSAGE);
        let d = e.details.unwrap();
        assert_eq!(d["last_seen_at"], 42);
        assert_eq!(d["cause"], "doubao_cdp_disconnected");
        // 豆包还在:原样(页面 reload 等)
        let e = crash_aware_error(RelayError::doubao_cdp_disconnected(), true, 42);
        assert_eq!(e.code, "doubao_cdp_disconnected");
        // 非断线类错误不改
        let e = crash_aware_error(RelayError::new("generation_task_timeout", "x"), false, 42);
        assert_eq!(e.code, "generation_task_timeout");
        assert_eq!(
            crash_aware_error(RelayError::doubao_not_ready(), false, 0).code,
            "doubao_crashed"
        );
    }

    #[test]
    fn preparation_disconnect_retries_only_before_send_and_only_once() {
        let disconnected = RelayError::doubao_cdp_disconnected();
        assert!(should_retry_pre_submission_disconnect(
            &disconnected,
            false,
            0
        ));
        assert!(!should_retry_pre_submission_disconnect(
            &disconnected,
            true,
            0
        ));
        assert!(!should_retry_pre_submission_disconnect(
            &disconnected,
            false,
            1
        ));
        assert!(!should_retry_pre_submission_disconnect(
            &RelayError::new("generation_preparing_failed", "x"),
            false,
            0
        ));
    }

    #[test]
    fn preparation_timeout_retries_only_when_send_never_started() {
        let before_send =
            generation::preparing_timeout("submit").with_details(json!({ "sent": false }));
        let after_send =
            generation::preparing_timeout("submit").with_details(json!({ "sent": true }));

        assert!(should_retry_preparing_timeout(&before_send, 0));
        assert!(should_retry_preparing_timeout(&before_send, 1));
        assert!(!should_retry_preparing_timeout(&before_send, 2));
        assert!(!should_retry_preparing_timeout(&after_send, 0));
        assert!(!should_retry_preparing_timeout(
            &RelayError::new("generation_preparing_failed", "x"),
            0
        ));
        assert!(!should_retry_preparing_timeout(
            &generation::preparing_timeout("submit"),
            0
        ));
    }

    #[tokio::test]
    async fn preparation_disconnect_immediately_invalidates_the_failed_session() {
        use futures_util::StreamExt;

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (socket, _) = listener.accept().await.unwrap();
            let mut websocket = tokio_tungstenite::accept_async(socket).await.unwrap();
            while websocket.next().await.is_some() {}
        });
        let session =
            crate::cdp::CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
                .await
                .unwrap();

        invalidate_submission_session_on_disconnect(
            &session,
            &RelayError::doubao_cdp_disconnected(),
        )
        .await;

        assert!(session.is_closed());
    }

    #[test]
    fn task_error_for_adds_account_only_for_quota_exhausted() {
        let e = RelayError::new("account_quota_exhausted", "x")
            .with_details(json!({ "reason": "quota_exhausted", "doubao_message": "额度不足" }));
        let t = task_error_for(&e, "张三");
        assert_eq!(t.code, "account_quota_exhausted");
        let d = t.details.unwrap();
        assert_eq!(d["account"], "张三");
        assert_eq!(d["reason"], "quota_exhausted");

        let e = RelayError::new("account_quota_exhausted", "x");
        assert_eq!(task_error_for(&e, "a").details.unwrap()["account"], "a");

        let e = RelayError::new("generation_doubao_refused", "x")
            .with_details(json!({ "reason": "refused" }));
        let t = task_error_for(&e, "张三");
        assert!(t.details.unwrap().get("account").is_none());
    }

    #[test]
    fn result_file_from_uses_downloaded_filename_and_relative_path() {
        let d = Downloaded {
            path: PathBuf::from("/data/results/01HQZ9/image_2.jpg"),
            mime: "image/jpeg".into(),
            bytes: 10,
            width: Some(4),
            height: Some(3),
        };
        let f = result_file_from("01HQZ9", 1, &d, "https://p3.byteimg.com/a.png");
        assert_eq!(f.index, 1);
        assert_eq!(f.filename, "image_2.jpg");
        assert_eq!(f.relative_path, "results/01HQZ9/image_2.jpg");
        assert_eq!(f.mime, "image/jpeg");
        assert_eq!((f.width, f.height, f.duration_ms), (Some(4), Some(3), None));
        assert_eq!(
            f.source_url.as_deref(),
            Some("https://p3.byteimg.com/a.png")
        );
    }

    #[test]
    fn cancelled_error_carries_from_status_and_step() {
        let e = cancelled_error(TaskStatus::Generating, Some("capture"));
        assert_eq!(e.code, "generation_cancelled");
        let d = e.details.unwrap();
        assert_eq!(d["from_status"], "generating");
        assert_eq!(d["step"], "capture");
        let e = cancelled_error(TaskStatus::Pending, None);
        assert!(e.details.unwrap()["step"].is_null());
    }

    #[test]
    fn phase_clock_excludes_waiting_from_budget() {
        let mut c = PhaseClock::new(Duration::from_secs(300), Duration::from_secs(3600));
        let d0 = c.next_deadline();
        assert!(d0 <= c.active_since + Duration::from_secs(300));
        c.pause();
        assert!(c.waiting());
        let dw = c.next_deadline();
        assert!(
            dw >= c.waiting_since.unwrap() + Duration::from_secs(3600) - Duration::from_millis(1)
        );
        // 重复 pause 不重置等待起点
        let w1 = c.waiting_since.unwrap();
        c.pause();
        assert_eq!(c.waiting_since.unwrap(), w1);
        c.resume();
        assert!(!c.waiting());
        let d1 = c.next_deadline();
        // 恢复后剩余预算 = 300s - 已用(≈0),deadline 不早于恢复时刻 + 299s
        assert!(d1 >= c.active_since + Duration::from_secs(299));
        // resume 未处于等待时是 no-op
        let a = c.active_since;
        c.resume();
        assert_eq!(c.active_since, a);
    }

    #[test]
    fn confirm_timeout_error_has_code_and_details() {
        let e = confirm_timeout_error(Duration::from_millis(1500), Some("checking"));
        assert_eq!(e.code, "generation_confirm_timeout");
        let d = e.details.unwrap();
        assert_eq!(d["waited_ms"], 1500);
        assert_eq!(d["stage"], "checking");
        assert_eq!(task_panicked_error().code, "system_task_panicked");
    }
}
