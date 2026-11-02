//! 任务调度器(Task 2.7):入队校验、FIFO 派发、并发约束、取消 / 确认、超时扫描、事件广播。
//!
//! `docs/spec/08-task-lifecycle.md` "并发模型":
//! - 全局最大并发 `concurrency.global_max`(preparing + generating + waiting);
//! - 单账号并发 `per_account_max`(硬 1:每个账号一把 `tokio::sync::Mutex`,runner 持有到结束);
//! - 队列深度 `queue_depth_max`(pending + waiting_user_confirm),满 → `queue_full`(429);
//! - **严格 FIFO**:只看 pending 头任务,全局有名额且该账号锁空闲才派发,否则等待。
//! - 超时扫描每 10s:pending > 1h → `failed/generation_queue_timeout`(spec 08 写作 `queue_timeout`,
//!   11-error-codes 未列,按 `generation_*` 前缀补);preparing / generating / 确认等待的超时由 runner 自己计。
//!
//! 所有状态变化 **DB 先写再广播** [`TaskEvent`];`submit` 时若豆包未连接,`accounts()` 直接
//! 返回 `doubao_*`(503),即 spec 的"豆包崩溃时新请求 503"。

use std::collections::{HashMap, VecDeque};
use std::sync::atomic::{AtomicUsize, Ordering as AtomicOrdering};
use std::sync::{Arc, Mutex as StdMutex, RwLock as StdRwLock};
use std::time::Duration;

use serde_json::json;
use tokio::sync::{broadcast, watch, Mutex, Notify, OwnedMutexGuard, RwLock};
use tokio::task::JoinHandle;
use tokio_util::sync::CancellationToken;

use crate::config::Config;
use crate::error::{RelayError, RelayResult};
use crate::logging::now_ms;
use crate::storage::Storage;
use crate::tasks::runner::{self, GenerationBackend};
use crate::tasks::webhook::{self, WebhookPolicy};
use crate::tasks::{
    new_task_id, prompt_hash, validate_request, GenerateRequest, StageEntry, Task, TaskError,
    TaskStage, TaskStatus, TaskSummary,
};

const EVENT_CAPACITY: usize = 256;

/// 任务事件(Task 2.9 转发为 `task_created` / `task_updated` / `task_completed`)。
/// `Completed` 按计划携带完整 `Task`(spec 05 `task_completed` payload),体积差异可接受。
#[derive(Clone, Debug, PartialEq)]
#[allow(clippy::large_enum_variant)]
pub enum TaskEvent {
    Created(TaskSummary),
    Updated {
        task_id: String,
        status: TaskStatus,
        stage: Option<TaskStage>,
        message: Option<String>,
        progress: Option<f32>,
        updated_at: i64,
    },
    /// 终态(completed / failed / cancelled),完整任务。
    Completed(Task),
}

/// 运行中任务的控制句柄(调度器与 runner 各持一份)。
#[derive(Clone)]
pub struct TaskHandle {
    pub cancel: CancellationToken,
    /// 每次用户 confirm 都 `send(true)`(capture 边沿触发,一次确认只消费一次)。
    pub confirm_tx: watch::Sender<bool>,
    pub status: Arc<StdRwLock<TaskStatus>>,
    pub account: String,
    /// `fail_all_running` 的失败原因:先写它再 `cancel`,runner 收到取消时按它落 `failed`。
    pub abort: Arc<StdMutex<Option<TaskError>>>,
}

impl TaskHandle {
    fn new(account: &str) -> Self {
        let (confirm_tx, _) = watch::channel(false);
        Self {
            cancel: CancellationToken::new(),
            confirm_tx,
            status: Arc::new(StdRwLock::new(TaskStatus::Pending)),
            account: account.to_string(),
            abort: Arc::new(StdMutex::new(None)),
        }
    }

    pub fn current_status(&self) -> TaskStatus {
        *self.status.read().unwrap_or_else(|p| p.into_inner())
    }
}

/// 各类时间参数(默认即 spec;测试用小值)。
#[derive(Clone, Debug)]
pub struct SchedulerTiming {
    /// 超时扫描间隔(spec:10s)。
    pub scan_interval: Duration,
    /// pending 未被挑中的上限(spec:1h)。
    pub pending_timeout: Duration,
    /// preparing 整段(spec:60s)。
    pub preparing_timeout: Duration,
    /// generating(图像,不含等待确认;spec:5min)。
    pub image_timeout: Duration,
    /// generating(视频;spec:20min)。
    pub video_timeout: Duration,
    /// waiting_user_confirm(spec:1h)。
    pub confirm_timeout: Duration,
    /// 终态回调的超时与重试(spec 04:30s,1s / 10s / 60s)。
    pub webhook: WebhookPolicy,
}

impl Default for SchedulerTiming {
    fn default() -> Self {
        Self {
            scan_interval: Duration::from_secs(10),
            pending_timeout: Duration::from_secs(3600),
            // 上传画风库参考图后，豆包页面还需要完成素材入场和提交按钮状态同步。
            // 60 秒在本地 relay 上会把仍在正常准备中的请求误判为失败，尤其是首次
            // 上传较大的 data URL 时。给准备阶段留出两分钟，生成阶段仍由后续状态机控制。
            preparing_timeout: Duration::from_secs(120),
            // 带参考图和长提示词的图片任务在豆包端可能超过 5 分钟。
            // 延长单图预算，避免执行器在豆包仍处于 generating 时提前终止。
            image_timeout: Duration::from_secs(900),
            video_timeout: Duration::from_secs(1200),
            confirm_timeout: Duration::from_secs(3600),
            webhook: WebhookPolicy::default(),
        }
    }
}

#[derive(Clone, Debug)]
struct PendingEntry {
    id: String,
    account: String,
    created_at: i64,
}

pub struct Scheduler {
    storage: Storage,
    backend: Arc<dyn GenerationBackend>,
    config: Arc<RwLock<Config>>,
    timing: SchedulerTiming,
    pending: StdMutex<VecDeque<PendingEntry>>,
    running: StdMutex<HashMap<String, TaskHandle>>,
    account_locks: StdMutex<HashMap<String, Arc<Mutex<()>>>>,
    /// 串行化 `submit` 的"查深度 → 入库 → 入队"。
    submit_lock: Mutex<()>,
    events: broadcast::Sender<TaskEvent>,
    wake: Notify,
    /// 终态回调共用的 HTTP client(`webhook::new_client()`:不跟随重定向)。
    webhook_client: reqwest::Client,
    /// 进行中的账号切换数(spec 08 "账号互斥"的 switch_lock):> 0 时派发循环不启动新 runner。
    /// 只在持有 `running` 锁时读写,与"运行中为空"的检查构成原子的 check-then-act。
    switching: AtomicUsize,
    /// 未指定账号的直连请求使用的轮询游标。任务本身仍由账号锁和单页调度器串行保护。
    auto_account_cursor: AtomicUsize,
}

/// [`Scheduler::begin_switch`] 返回的守卫:持有期间不派发新任务,drop 时恢复并唤醒派发。
pub struct SwitchGuard {
    sched: Arc<Scheduler>,
}

impl Drop for SwitchGuard {
    fn drop(&mut self) {
        {
            let _running = self.sched.lock_running();
            self.sched.switching.fetch_sub(1, AtomicOrdering::SeqCst);
        }
        self.sched.wake.notify_one();
    }
}

impl Scheduler {
    pub fn new(
        storage: Storage,
        backend: Arc<dyn GenerationBackend>,
        config: Arc<RwLock<Config>>,
    ) -> Arc<Self> {
        Self::new_with(storage, backend, config, SchedulerTiming::default())
    }

    pub fn new_with(
        storage: Storage,
        backend: Arc<dyn GenerationBackend>,
        config: Arc<RwLock<Config>>,
        timing: SchedulerTiming,
    ) -> Arc<Self> {
        Self::new_with_webhook_client(storage, backend, config, timing, webhook::new_client())
    }

    /// 同 [`Scheduler::new_with`],回调 client 可替换(测试把回调域名解析到本地 mock)。
    pub fn new_with_webhook_client(
        storage: Storage,
        backend: Arc<dyn GenerationBackend>,
        config: Arc<RwLock<Config>>,
        timing: SchedulerTiming,
        webhook_client: reqwest::Client,
    ) -> Arc<Self> {
        let (events, _) = broadcast::channel(EVENT_CAPACITY);
        Arc::new(Self {
            storage,
            backend,
            config,
            timing,
            pending: StdMutex::new(VecDeque::new()),
            running: StdMutex::new(HashMap::new()),
            account_locks: StdMutex::new(HashMap::new()),
            submit_lock: Mutex::new(()),
            events,
            wake: Notify::new(),
            webhook_client,
            switching: AtomicUsize::new(0),
            auto_account_cursor: AtomicUsize::new(0),
        })
    }

    pub fn subscribe(&self) -> broadcast::Receiver<TaskEvent> {
        self.events.subscribe()
    }

    pub fn storage(&self) -> &Storage {
        &self.storage
    }

    pub fn backend(&self) -> Arc<dyn GenerationBackend> {
        Arc::clone(&self.backend)
    }

    /// 从数据库恢复 relay 重启前仍在排队的任务,避免 pending 任务只存在内存中而永久不再执行。
    pub async fn restore_pending(&self) -> RelayResult<u32> {
        let statuses = [TaskStatus::Pending];
        let mut all = Vec::new();
        let mut cursor = None;
        loop {
            let (mut page, next) = self
                .storage
                .list_tasks(Some(&statuses), None, 1000, cursor.as_deref())
                .await?;
            all.append(&mut page);
            match next {
                Some(value) => cursor = Some(value),
                None => break,
            }
        }
        all.reverse();

        let mut restored = 0;
        let mut q = self.lock_pending();
        for task in all {
            if q.iter().any(|entry| entry.id == task.id) {
                continue;
            }
            q.push_back(PendingEntry {
                id: task.id,
                account: task.account,
                created_at: task.created_at,
            });
            restored += 1;
        }
        drop(q);
        if restored > 0 {
            self.wake.notify_one();
        }
        Ok(restored)
    }

    /// 为未指定账号的直连请求选择下一个已登录账号。
    ///
    /// 账号只在入队前选择；真正执行仍要经过 pending FIFO、全局并发和
    /// per-account 锁，因此不会因为轮换而让同一账号同时跑两个任务。
    pub async fn next_auto_account(&self) -> RelayResult<String> {
        let list = self.backend.accounts(false).await?;
        let mut accounts = Vec::new();
        for account in list.accounts {
            let name = account.name.trim();
            if name.is_empty() || account.is_ambiguous || accounts.iter().any(|item| item == name) {
                continue;
            }
            accounts.push(name.to_string());
        }
        let cursor = self
            .auto_account_cursor
            .fetch_add(1, AtomicOrdering::SeqCst);
        choose_round_robin_account(&accounts, cursor)
            .map(|(account, _)| account)
            .ok_or_else(RelayError::doubao_not_logged_in)
    }

    pub fn timing(&self) -> &SchedulerTiming {
        &self.timing
    }

    pub(crate) fn emit(&self, ev: TaskEvent) {
        let _ = self.events.send(ev);
    }

    /// 任务进入终态(DB 已写、事件已发)后调用:请求带了 `callback_url` 就在后台投递 webhook
    /// (`task.completed` / `task.failed` / `task.cancelled`,spec 04)。token 取**投递时**的
    /// `config.server.bearer_token`(设置页重新生成后立即生效);投递失败只落 `webhook_log` 与
    /// `callback_failed`,不影响任务状态。返回后台任务句柄(没有回调时 `None`)。
    pub(crate) fn notify_terminal(&self, task: &Task) -> Option<JoinHandle<()>> {
        let url = task.request.callback_url.clone()?;
        let event = webhook::event_for(task.status)?;
        let task = task.clone();
        let client = self.webhook_client.clone();
        let storage = self.storage.clone();
        let config = Arc::clone(&self.config);
        let policy = self.timing.webhook.clone();
        Some(tokio::spawn(async move {
            let token = config.read().await.server.bearer_token.clone();
            let outcome =
                webhook::deliver_with(&client, &task, event, &url, &token, &storage, &policy).await;
            tracing::debug!(
                task_id = %task.id,
                event,
                delivered = outcome.delivered,
                attempts = outcome.attempts,
                "回调投递结束"
            );
        }))
    }

    // ------------------------------------------------------------------
    // 提交
    // ------------------------------------------------------------------

    /// 校验 → 实时读取账号存在性(`backend.accounts(true)`)→ uploads 存在且未过期 → 队列深度
    /// → `insert_task(pending)` → `Created` → 唤醒派发。返回 `(task_id, queue_position, created_at)`。
    pub async fn submit(&self, req: GenerateRequest) -> RelayResult<(String, u32, i64)> {
        validate_request(&req)?;
        // 豆包崩溃 / 未连接 → 503(spec 08;`accounts(false)` 可能命中 2s 缓存,不能只靠它)
        self.backend.ready().await?;

        // 这里不能复用缓存:账号可能刚退出、切换或页面已经变成未登录。
        // 如果把旧缓存当成事实,任务会先进入队列,之后在豆包页面里长时间占用一个执行位。
        let list = self.backend.accounts(true).await?;
        let matches: Vec<_> = list
            .accounts
            .iter()
            .filter(|a| a.name == req.account)
            .collect();
        let Some(first) = matches.first() else {
            return Err(RelayError::account_not_found(
                &req.account,
                list.accounts.iter().map(|a| a.name.clone()).collect(),
            ));
        };
        if matches.len() > 1 || first.is_ambiguous {
            return Err(RelayError::account_ambiguous(
                &req.account,
                matches.iter().map(|a| a.label.clone()).collect(),
            ));
        }

        let now = now_ms();
        for id in &req.reference_files {
            match self.storage.get_upload(id).await? {
                None => return Err(upload_not_found(id)),
                Some(u) if u.expires_at < now => return Err(upload_expired(id, u.expires_at)),
                Some(_) => {}
            }
        }

        let _serial = self.submit_lock.lock().await;
        let cfg = self.config.read().await;
        let max_depth = cfg.concurrency.queue_depth_max;
        let retention_hours = cfg.storage.result_retention_hours;
        drop(cfg);
        let depth = self.queue_depth();
        if depth >= max_depth {
            return Err(queue_full(depth, max_depth));
        }

        let now = now_ms();
        let task = Task {
            id: new_task_id(),
            request: req,
            status: TaskStatus::Pending,
            stage: None,
            stage_message: None,
            progress: None,
            error: None,
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: None,
            result_files: None,
            created_at: now,
            started_at: None,
            completed_at: None,
            expires_at: now + i64::from(retention_hours) * 3_600_000,
            stage_history: vec![StageEntry {
                stage: TaskStatus::Pending.as_str().to_string(),
                message: None,
                ts: now,
            }],
            queue_position: None,
        };
        self.storage.insert_task(&task).await?;
        let position = {
            let mut q = self.lock_pending();
            q.push_back(PendingEntry {
                id: task.id.clone(),
                account: task.request.account.clone(),
                created_at: now,
            });
            q.len() as u32
        };
        tracing::info!(
            task_id = %task.id,
            account = %task.request.account,
            task_type = task.request.task_type.as_str(),
            model_name = %task.request.model_name,
            queue_position = position,
            prompt_length = task.request.prompt.chars().count(),
            prompt_hash = %prompt_hash(&task.request.prompt),
            "任务已入队"
        );
        self.emit(TaskEvent::Created(TaskSummary::from(&task)));
        self.wake.notify_one();
        Ok((task.id, position, now))
    }

    /// pending + waiting_user_confirm 的数量(spec 08 队列深度口径)。
    pub fn queue_depth(&self) -> u32 {
        let pending = self.lock_pending().len();
        let waiting = self
            .lock_running()
            .values()
            .filter(|h| h.current_status() == TaskStatus::WaitingUserConfirm)
            .count();
        (pending + waiting) as u32
    }

    /// 在 pending 队列中的 1-based 位置;不在队列里返回 `None`。
    pub async fn queue_position(&self, id: &str) -> Option<u32> {
        self.lock_pending()
            .iter()
            .position(|e| e.id == id)
            .map(|i| i as u32 + 1)
    }

    /// 读任务并补上运行时态 `queue_position`(供 HTTP / IPC 层)。
    pub async fn get_task(&self, id: &str) -> RelayResult<Option<Task>> {
        let Some(mut task) = self.storage.get_task(id).await? else {
            return Ok(None);
        };
        if task.status == TaskStatus::Pending {
            task.queue_position = self.queue_position(id).await;
        }
        Ok(Some(task))
    }

    // ------------------------------------------------------------------
    // 取消 / 确认
    // ------------------------------------------------------------------

    /// pending → 立即 `cancelled`;preparing / generating / waiting → 触发 cancel token(runner 落终态);
    /// 终态 → `task_not_cancellable`;不存在 → `task_not_found`。返回取消前的状态。
    pub async fn cancel(&self, id: &str) -> RelayResult<TaskStatus> {
        let removed = {
            let mut q = self.lock_pending();
            q.iter().position(|e| e.id == id).and_then(|i| q.remove(i))
        };
        if removed.is_some() {
            let error = runner::cancelled_error(TaskStatus::Pending, None);
            self.finish_queued(id, TaskStatus::Cancelled, error).await?;
            return Ok(TaskStatus::Pending);
        }
        let handle = self.lock_running().get(id).cloned();
        if let Some(h) = handle {
            let from = h.current_status();
            if from.is_terminal() {
                return Err(task_not_cancellable(from));
            }
            tracing::info!(task_id = %id, from_status = from.as_str(), "请求取消运行中的任务");
            h.cancel.cancel();
            return Ok(from);
        }
        match self.storage.get_task(id).await? {
            None => Err(task_not_found(id)),
            Some(t) => Err(task_not_cancellable(t.status)),
        }
    }

    /// `waiting_user_confirm` 下 `confirm_tx.send(true)`;否则 `task_not_waiting_confirm`。
    /// 返回新阶段名 `generating`(spec 04 `new_stage`)。
    pub async fn confirm(&self, id: &str) -> RelayResult<String> {
        let handle = self.lock_running().get(id).cloned();
        match handle {
            Some(h) => {
                let status = h.current_status();
                if status != TaskStatus::WaitingUserConfirm {
                    return Err(task_not_waiting_confirm(status));
                }
                let _ = h.confirm_tx.send(true);
                tracing::info!(task_id = %id, "用户已确认,继续生成");
                Ok(TaskStatus::Generating.as_str().to_string())
            }
            None => match self.storage.get_task(id).await? {
                None => Err(task_not_found(id)),
                Some(t) => Err(task_not_waiting_confirm(t.status)),
            },
        }
    }

    /// 豆包崩溃等:pending 全部 `failed/{code}`,运行中的任务写入失败原因后触发取消。
    pub async fn fail_all_running(&self, code: &str, message: &str) {
        self.fail_all_running_with(code, message, None).await
    }

    /// 同 [`fail_all_running`](Self::fail_all_running),失败原因带 `details`
    /// (`doubao_crashed{last_seen_at}`)。
    pub async fn fail_all_running_with(
        &self,
        code: &str,
        message: &str,
        details: Option<serde_json::Value>,
    ) {
        let make = || TaskError {
            code: code.to_string(),
            message: message.to_string(),
            details: details.clone(),
        };
        let queued: Vec<PendingEntry> = self.lock_pending().drain(..).collect();
        for entry in queued {
            let error = make();
            if let Err(e) = self
                .finish_queued(&entry.id, TaskStatus::Failed, error)
                .await
            {
                tracing::error!(task_id = %entry.id, code = e.code, error = %e, "标记排队任务失败时出错");
            }
        }
        let running: Vec<(String, TaskHandle)> = self
            .lock_running()
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect();
        for (id, h) in running {
            if h.current_status().is_terminal() {
                continue;
            }
            *h.abort.lock().unwrap_or_else(|p| p.into_inner()) = Some(make());
            tracing::warn!(task_id = %id, code, "中止运行中的任务");
            h.cancel.cancel();
        }
    }

    // ------------------------------------------------------------------
    // 派发循环
    // ------------------------------------------------------------------

    /// 后台循环:被 `submit` / runner 结束唤醒时派发;每 `scan_interval` 扫一次超时。
    pub fn spawn_loop(self: &Arc<Self>) -> JoinHandle<()> {
        let this = Arc::clone(self);
        tokio::spawn(async move {
            let mut scan = tokio::time::interval(this.timing.scan_interval);
            scan.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
            scan.tick().await; // 第一次立即返回
            loop {
                this.dispatch().await;
                tokio::select! {
                    _ = this.wake.notified() => {}
                    _ = scan.tick() => {
                        this.scan_timeouts().await;
                    }
                }
            }
        })
    }

    /// 严格 FIFO:只看队头;以下全部满足才派发,否则停下等待(被 runner 结束 / 切换结束 / submit 唤醒):
    /// - 没有进行中的账号切换([`begin_switch`](Self::begin_switch),spec 08 switch_lock);
    /// - 全局名额(`global_max`);
    /// - **运行中的任务都属于队头任务的账号**(Task 4.2:豆包只有一个当前账号,切账号整页 reload,
    ///   其它账号的任务在跑时不能切走——跨账号任务因此串行;同账号再由账号锁限 1);
    /// - 该账号锁空闲。
    async fn dispatch(self: &Arc<Self>) {
        loop {
            let global_max = self.config.read().await.concurrency.global_max as usize;
            let picked = {
                let mut q = self.lock_pending();
                let Some(head) = q.front().cloned() else {
                    break;
                };
                let mut running = self.lock_running();
                if self.switching.load(AtomicOrdering::SeqCst) > 0 {
                    break;
                }
                if running.len() >= global_max {
                    break;
                }
                if running.values().any(|h| h.account != head.account) {
                    break;
                }
                let lock = self.account_lock(&head.account);
                let Ok(guard) = lock.try_lock_owned() else {
                    break;
                };
                q.pop_front();
                let handle = TaskHandle::new(&head.account);
                running.insert(head.id.clone(), handle.clone());
                (head, handle, guard)
            };
            let (entry, handle, guard) = picked;
            tracing::info!(
                task_id = %entry.id,
                account = %entry.account,
                waited_ms = now_ms().saturating_sub(entry.created_at),
                "调度器挑中任务"
            );
            self.spawn_runner(entry.id, handle, guard);
        }
    }

    fn spawn_runner(self: &Arc<Self>, id: String, handle: TaskHandle, guard: OwnedMutexGuard<()>) {
        let this = Arc::clone(self);
        tokio::spawn(async move {
            let inner = tokio::spawn(runner::run(Arc::clone(&this), id.clone(), handle.clone()));
            if let Err(e) = inner.await {
                if e.is_panic() {
                    tracing::error!(task_id = %id, "任务运行线程 panic");
                    this.fail_panicked(&id).await;
                }
            }
            this.lock_running().remove(&id);
            drop(guard);
            this.wake.notify_one();
        });
    }

    async fn fail_panicked(&self, id: &str) {
        match self.storage.get_task(id).await {
            Ok(Some(t)) if !t.status.is_terminal() => {
                let e = runner::task_panicked_error();
                let error = TaskError {
                    code: e.code.to_string(),
                    message: e.message,
                    details: None,
                };
                if let Err(e) = self.finish_queued(id, TaskStatus::Failed, error).await {
                    tracing::error!(task_id = %id, code = e.code, "写入 panic 失败态出错");
                }
            }
            Ok(_) => {}
            Err(e) => tracing::error!(task_id = %id, code = e.code, "读取 panic 任务失败"),
        }
    }

    /// 超时扫描:pending 超过 `pending_timeout` → `failed/generation_queue_timeout`。
    async fn scan_timeouts(&self) {
        let now = now_ms();
        let limit = self.timing.pending_timeout.as_millis() as i64;
        let expired: Vec<PendingEntry> = {
            let mut q = self.lock_pending();
            let (expired, keep): (Vec<_>, Vec<_>) = q
                .drain(..)
                .partition(|e| now.saturating_sub(e.created_at) > limit);
            q.extend(keep);
            expired
        };
        for entry in expired {
            let waited = now.saturating_sub(entry.created_at);
            tracing::warn!(task_id = %entry.id, waited_ms = waited, "排队超时");
            let e = queue_timeout(waited);
            let error = TaskError {
                code: e.code.to_string(),
                message: e.message,
                details: e.details,
            };
            if let Err(e) = self
                .finish_queued(&entry.id, TaskStatus::Failed, error)
                .await
            {
                tracing::error!(task_id = %entry.id, code = e.code, "写入排队超时失败态出错");
            }
        }
    }

    // ------------------------------------------------------------------
    // 内部
    // ------------------------------------------------------------------

    /// 不经 runner 的终态写入(pending 取消 / 排队超时 / 崩溃 / panic):DB 先写再广播。
    async fn finish_queued(
        &self,
        id: &str,
        status: TaskStatus,
        error: TaskError,
    ) -> RelayResult<()> {
        let Some(mut task) = self.storage.get_task(id).await? else {
            return Err(task_not_found(id));
        };
        let from = task.status;
        let now = now_ms();
        task.status = status;
        task.completed_at = Some(now);
        task.error = Some(error.clone());
        self.storage.update_task(&task).await?;
        self.storage
            .append_stage(id, status.as_str(), Some(&error.message), now)
            .await?;
        task.stage_history.push(StageEntry {
            stage: status.as_str().to_string(),
            message: Some(error.message.clone()),
            ts: now,
        });
        tracing::info!(
            task_id = %id,
            from_status = from.as_str(),
            to_status = status.as_str(),
            stage = task.stage.map(|s| s.as_str()).unwrap_or(""),
            duration_ms = now.saturating_sub(task.created_at),
            prompt_length = task.request.prompt.chars().count(),
            prompt_hash = %prompt_hash(&task.request.prompt),
            error_code = %error.code,
            "任务状态变化"
        );
        self.emit(TaskEvent::Updated {
            task_id: id.to_string(),
            status,
            stage: task.stage,
            message: Some(error.message),
            progress: None,
            updated_at: now,
        });
        self.notify_terminal(&task);
        self.emit(TaskEvent::Completed(task));
        Ok(())
    }

    fn account_lock(&self, account: &str) -> Arc<Mutex<()>> {
        let mut locks = self.account_locks.lock().unwrap_or_else(|p| p.into_inner());
        Arc::clone(locks.entry(account.to_string()).or_default())
    }

    fn lock_pending(&self) -> std::sync::MutexGuard<'_, VecDeque<PendingEntry>> {
        self.pending.lock().unwrap_or_else(|p| p.into_inner())
    }

    fn lock_running(&self) -> std::sync::MutexGuard<'_, HashMap<String, TaskHandle>> {
        self.running.lock().unwrap_or_else(|p| p.into_inner())
    }

    /// 运行中任务数(测试 / 状态面板)。
    pub fn running_count(&self) -> usize {
        self.lock_running().len()
    }

    /// 仍在运行(非终态)的任务 ID,`account` 为 `Some` 时只看该账号(`account_busy` 用)。
    pub fn running_task_ids(&self, account: Option<&str>) -> Vec<String> {
        let mut ids: Vec<String> = self
            .lock_running()
            .iter()
            .filter(|(_, h)| !h.current_status().is_terminal())
            .filter(|(_, h)| account.is_none_or(|a| h.account == a))
            .map(|(id, _)| id.clone())
            .collect();
        ids.sort();
        ids
    }

    /// 是否有账号切换在进行(派发暂停中)。
    pub fn switching(&self) -> bool {
        self.switching.load(AtomicOrdering::SeqCst) > 0
    }

    /// 开始一次手动账号切换(spec 08 "账号互斥"):没有任何非终态任务在运行才放行
    /// (否则 `account_busy{account, running_task_ids}`,409;切换会整页 reload,运行中的任务必丢,
    /// 所以比 spec 字面的"目标账号有进行中任务"更严),并在返回的
    /// [`SwitchGuard`] 存活期间暂停派发——检查与置位在同一把 `running` 锁里完成,
    /// 不会出现"检查通过后调度循环又派出一个任务、随即被切换 reload 打断"的窗口。
    /// 切换期间提交的任务照常入队,guard drop 后才进入 preparing(UC-31)。
    pub fn begin_switch(self: &Arc<Self>, target: &str) -> RelayResult<SwitchGuard> {
        {
            let running = self.lock_running();
            let mut ids: Vec<String> = running
                .iter()
                .filter(|(_, h)| !h.current_status().is_terminal())
                .map(|(id, _)| id.clone())
                .collect();
            if !ids.is_empty() {
                ids.sort();
                return Err(account_busy(target, ids));
            }
            self.switching.fetch_add(1, AtomicOrdering::SeqCst);
        }
        tracing::info!(target, "开始切换账号,暂停派发新任务");
        Ok(SwitchGuard {
            sched: Arc::clone(self),
        })
    }
}

pub(crate) fn choose_round_robin_account(
    accounts: &[String],
    cursor: usize,
) -> Option<(String, usize)> {
    if accounts.is_empty() {
        return None;
    }
    let index = cursor % accounts.len();
    Some((accounts[index].clone(), (index + 1) % accounts.len()))
}

// ------------------------------------------------------------------
// 错误(11-error-codes 任务类 / 上传类 / 账号类)
// ------------------------------------------------------------------

/// `account_busy{account, running_task_ids}`(409)。
pub fn account_busy(account: &str, running_task_ids: Vec<String>) -> RelayError {
    RelayError::new(
        "account_busy",
        "有任务正在运行,切换账号会中断它,请等任务结束或先取消任务后再切换",
    )
    .with_details(json!({ "account": account, "running_task_ids": running_task_ids }))
}

/// `queue_full{depth, max}`(429)。
pub fn queue_full(depth: u32, max: u32) -> RelayError {
    RelayError::new("queue_full", "任务队列已满,请等待队列中的任务完成后再提交")
        .with_details(json!({ "depth": depth, "max": max }))
}

/// `task_not_found{task_id}`(404)。
pub fn task_not_found(id: &str) -> RelayError {
    RelayError::new("task_not_found", "任务不存在,请确认任务 ID 后重试")
        .with_details(json!({ "task_id": id }))
}

/// `task_not_cancellable{current_status}`(409)。
pub fn task_not_cancellable(status: TaskStatus) -> RelayError {
    RelayError::new("task_not_cancellable", "任务已结束,无法取消")
        .with_details(json!({ "current_status": status.as_str() }))
}

/// `task_not_waiting_confirm{current_status}`(409)。
pub fn task_not_waiting_confirm(status: TaskStatus) -> RelayError {
    RelayError::new(
        "task_not_waiting_confirm",
        "任务当前不在等待确认状态,无需确认",
    )
    .with_details(json!({ "current_status": status.as_str() }))
}

/// `upload_not_found{upload_id}`(404)。
pub fn upload_not_found(id: &str) -> RelayError {
    RelayError::new("upload_not_found", "参考文件不存在,请先上传后再提交")
        .with_details(json!({ "upload_id": id }))
}

/// `upload_expired{upload_id, expired_at}`(400)。
pub fn upload_expired(id: &str, expired_at: i64) -> RelayError {
    RelayError::new("upload_expired", "参考文件已过期,请重新上传后再提交")
        .with_details(json!({ "upload_id": id, "expired_at": expired_at }))
}

/// `generation_queue_timeout{waited_ms}`(写入 task.error;spec 08 的 `queue_timeout`)。
pub fn queue_timeout(waited_ms: i64) -> RelayError {
    RelayError::new(
        "generation_queue_timeout",
        "任务排队超过 1 小时未开始,已自动放弃,请稍后重新提交",
    )
    .with_details(json!({ "waited_ms": waited_ms }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::{Path, PathBuf};
    use std::sync::atomic::{AtomicUsize, Ordering};

    use async_trait::async_trait;
    use serde_json::Value;
    use tokio::sync::{mpsc, Semaphore};

    use crate::doubao::capture::{CaptureEvent, CaptureOptions, CaptureOutcome};
    use crate::doubao::download::Downloaded;
    use crate::doubao::generation::{
        self, StageFn, SubmissionIdentity, SubmissionPlan, SubmissionReport,
    };
    use crate::doubao::{Account, AccountList};
    use crate::storage::Upload;
    use crate::tasks::runner::INPUT_MODIFIED_MESSAGE;
    use crate::tasks::{ConfirmationMode, TaskType, WatermarkSource};

    #[test]
    fn automatic_account_selection_round_robins_without_reusing_first_account() {
        let accounts = vec!["甲".to_string(), "乙".to_string(), "丙".to_string()];
        let mut cursor = 0;
        let mut selected = Vec::new();
        for _ in 0..6 {
            let (account, next_cursor) = choose_round_robin_account(&accounts, cursor).unwrap();
            selected.push(account);
            cursor = next_cursor;
        }
        assert_eq!(selected, ["甲", "乙", "丙", "甲", "乙", "丙"]);
    }

    // ---- mock backend ----

    #[derive(Clone, Debug)]
    enum Step {
        Emit(CaptureEvent),
        /// 等用户确认(`confirm_rx` 变 true)。
        WaitConfirm,
        /// 等测试放行(`MockBackend::release`)。
        Gate,
        /// 永不响应取消，用来验证调度器不会被失控的捕获任务卡住。
        IgnoreCancel,
        Sleep(u64),
        Ok(Vec<String>),
        /// 视频结果:原片地址 + 无水印来源。
        OkVideo(String, WatermarkSource),
        Fail(RelayError),
    }

    #[derive(Clone, Debug)]
    enum SubmitMode {
        Ok,
        Fail(RelayError),
        /// 挂起直到被取消。
        Hang,
    }

    struct MockBackend {
        accounts: Vec<String>,
        current: StdMutex<Option<String>>,
        switches: StdMutex<Vec<String>>,
        submit_mode: SubmitMode,
        script: Vec<Step>,
        gate: Semaphore,
        running: AtomicUsize,
        max_running: AtomicUsize,
        captures_started: AtomicUsize,
        downloads: AtomicUsize,
        video_downloads: StdMutex<Vec<String>>,
        /// 为 true 时 `download_video` 写半截文件后永不返回(模拟 180s 长下载中途被取消)。
        video_hang: std::sync::atomic::AtomicBool,
        last_capture: StdMutex<Option<CaptureOptions>>,
        /// `ready()` 返回的错误(None = 就绪)。
        ready_err: StdMutex<Option<RelayError>>,
    }

    impl MockBackend {
        fn new(accounts: &[&str], script: Vec<Step>) -> Arc<Self> {
            Arc::new(Self {
                accounts: accounts.iter().map(|s| s.to_string()).collect(),
                current: StdMutex::new(accounts.first().map(|s| s.to_string())),
                switches: StdMutex::new(Vec::new()),
                submit_mode: SubmitMode::Ok,
                script,
                gate: Semaphore::new(0),
                running: AtomicUsize::new(0),
                max_running: AtomicUsize::new(0),
                captures_started: AtomicUsize::new(0),
                downloads: AtomicUsize::new(0),
                video_downloads: StdMutex::new(Vec::new()),
                video_hang: std::sync::atomic::AtomicBool::new(false),
                last_capture: StdMutex::new(None),
                ready_err: StdMutex::new(None),
            })
        }

        fn with_submit(accounts: &[&str], mode: SubmitMode) -> Arc<Self> {
            let mut m = Self::new(accounts, vec![]);
            Arc::get_mut(&mut m).unwrap().submit_mode = mode;
            m
        }

        fn release(&self, n: usize) {
            self.gate.add_permits(n);
        }

        fn switches(&self) -> Vec<String> {
            self.switches.lock().unwrap().clone()
        }
    }

    struct RunningGuard<'a>(&'a MockBackend);
    impl Drop for RunningGuard<'_> {
        fn drop(&mut self) {
            self.0.running.fetch_sub(1, Ordering::SeqCst);
        }
    }

    fn identity() -> SubmissionIdentity {
        SubmissionIdentity {
            local_message_id: Some("lm1".into()),
            message_id: None,
            conversation_id: Some("7001".into()),
        }
    }

    #[async_trait]
    impl GenerationBackend for MockBackend {
        async fn ready(&self) -> RelayResult<()> {
            match self.ready_err.lock().unwrap().clone() {
                Some(e) => Err(e),
                None => Ok(()),
            }
        }

        async fn accounts(&self, _force: bool) -> RelayResult<AccountList> {
            let current = self.current.lock().unwrap().clone();
            Ok(AccountList {
                current: current.clone(),
                accounts: self
                    .accounts
                    .iter()
                    .map(|n| Account {
                        name: n.clone(),
                        subscription: Some("体验套餐".into()),
                        label: format!("{n} 体验套餐"),
                        is_current: current.as_deref() == Some(n.as_str()),
                        is_ambiguous: false,
                    })
                    .collect(),
                read_at: now_ms(),
            })
        }

        async fn ensure_account(&self, account: &str) -> RelayResult<()> {
            if !self.accounts.iter().any(|a| a == account) {
                return Err(RelayError::account_not_found(
                    account,
                    self.accounts.clone(),
                ));
            }
            let mut cur = self.current.lock().unwrap();
            if cur.as_deref() != Some(account) {
                self.switches.lock().unwrap().push(account.to_string());
                *cur = Some(account.to_string());
            }
            Ok(())
        }

        async fn submit(
            &self,
            _request: &GenerateRequest,
            _plan: &SubmissionPlan,
            _reference_paths: &[PathBuf],
            cancel: &CancellationToken,
            on_stage: StageFn<'_>,
        ) -> RelayResult<SubmissionReport> {
            on_stage(TaskStage::Preparing, Some("模型与参数已就绪".into()));
            match &self.submit_mode {
                SubmitMode::Ok => Ok(SubmissionReport {
                    identity: Some(identity()),
                    send_started_at: now_ms(),
                    generation_requests: Vec::new(),
                }),
                SubmitMode::Fail(e) => Err(e.clone()),
                SubmitMode::Hang => {
                    cancel.cancelled().await;
                    Err(generation::cancelled("prepare"))
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
            *self.last_capture.lock().unwrap() = Some(opts);
            self.captures_started.fetch_add(1, Ordering::SeqCst);
            let n = self.running.fetch_add(1, Ordering::SeqCst) + 1;
            self.max_running.fetch_max(n, Ordering::SeqCst);
            let _guard = RunningGuard(self);
            let finish = |originals: Vec<String>| CaptureOutcome {
                originals,
                watermark_source: WatermarkSource::ConversationOriginal,
                conversation_id: Some("7001".into()),
                submission_id: Some("9002".into()),
            };
            for step in &self.script {
                if cancel.is_cancelled() {
                    return Err(generation::cancelled("capture"));
                }
                match step {
                    Step::Emit(ev) => {
                        let _ = events.send(ev.clone()).await;
                    }
                    Step::WaitConfirm => loop {
                        tokio::select! {
                            _ = cancel.cancelled() => return Err(generation::cancelled("capture")),
                            r = confirm_rx.changed() => {
                                if r.is_err() { return Err(generation::cancelled("capture")); }
                                if *confirm_rx.borrow_and_update() { break; }
                            }
                        }
                    },
                    Step::Gate => {
                        tokio::select! {
                            _ = cancel.cancelled() => return Err(generation::cancelled("capture")),
                            p = self.gate.acquire() => { p.unwrap().forget(); }
                        }
                    }
                    Step::IgnoreCancel => std::future::pending::<()>().await,
                    Step::Sleep(ms) => {
                        tokio::select! {
                            _ = cancel.cancelled() => return Err(generation::cancelled("capture")),
                            _ = tokio::time::sleep(Duration::from_millis(*ms)) => {}
                        }
                    }
                    Step::Ok(urls) => return Ok(finish(urls.clone())),
                    Step::OkVideo(url, source) => {
                        return Ok(CaptureOutcome {
                            watermark_source: *source,
                            ..finish(vec![url.clone()])
                        })
                    }
                    Step::Fail(e) => return Err(e.clone()),
                }
            }
            Ok(finish(vec!["https://x.byteimg.com/a".into()]))
        }

        async fn download_image(
            &self,
            _url: &str,
            dest_dir: &Path,
            index: u32,
        ) -> RelayResult<Downloaded> {
            self.downloads.fetch_add(1, Ordering::SeqCst);
            let path = dest_dir.join(format!("image_{}.png", index + 1));
            let bytes = b"\x89PNG\r\n\x1a\nfake";
            tokio::fs::write(&path, bytes).await?;
            Ok(Downloaded {
                path,
                mime: "image/png".into(),
                bytes: bytes.len() as u64,
                width: Some(1),
                height: Some(1),
            })
        }

        /// 与真实 `download::download_video` 同样先校验 `lr=unwatermarked`,再写一个 5s 的合成 mp4。
        async fn download_video(&self, url: &str, dest_dir: &Path) -> RelayResult<Downloaded> {
            self.video_downloads.lock().unwrap().push(url.to_string());
            if !crate::doubao::download::has_lr_unwatermarked(url) {
                return Err(crate::doubao::download::watermark_not_removed(url));
            }
            let path = dest_dir.join("video_1.mp4");
            let bytes = crate::doubao::download::synthetic_mp4(1000, 5000, false, true);
            if self.video_hang.load(Ordering::SeqCst) {
                tokio::fs::write(&path, &bytes[..16]).await?;
                std::future::pending::<()>().await;
            }
            tokio::fs::write(&path, &bytes).await?;
            Ok(Downloaded {
                path,
                mime: "video/mp4".into(),
                bytes: bytes.len() as u64,
                width: None,
                height: None,
            })
        }
    }

    // ---- helpers ----

    fn fast_timing() -> SchedulerTiming {
        SchedulerTiming {
            scan_interval: Duration::from_millis(20),
            pending_timeout: Duration::from_secs(3600),
            preparing_timeout: Duration::from_secs(5),
            image_timeout: Duration::from_secs(5),
            video_timeout: Duration::from_secs(5),
            confirm_timeout: Duration::from_secs(5),
            webhook: WebhookPolicy::default(),
        }
    }

    async fn setup(
        backend: Arc<MockBackend>,
        timing: SchedulerTiming,
        tweak: impl FnOnce(&mut Config),
    ) -> (Arc<Scheduler>, tempfile::TempDir) {
        let dir = tempfile::tempdir().unwrap();
        let storage = Storage::open(dir.path()).await.unwrap();
        let mut cfg = Config::default();
        tweak(&mut cfg);
        let sched = Scheduler::new_with(
            storage,
            backend as Arc<dyn GenerationBackend>,
            Arc::new(RwLock::new(cfg)),
            timing,
        );
        (sched, dir)
    }

    fn image_req(account: &str) -> GenerateRequest {
        GenerateRequest {
            account: account.into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "一只蓝色小鸟".into(),
            ratio: "1:1".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    fn video_req(account: &str) -> GenerateRequest {
        GenerateRequest {
            task_type: TaskType::Video,
            model_name: "Seedance 2.0 Fast".into(),
            ratio: "16:9".into(),
            duration: Some(5),
            ..image_req(account)
        }
    }

    async fn wait_status(sched: &Scheduler, id: &str, status: TaskStatus) -> Task {
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        loop {
            let t = sched.storage().get_task(id).await.unwrap().unwrap();
            if t.status == status {
                return t;
            }
            assert!(
                std::time::Instant::now() < deadline,
                "等待 {id} 进入 {} 超时,当前 {}",
                status.as_str(),
                t.status.as_str()
            );
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
    }

    /// runner 写终态(DB + 广播)后,`spawn_runner` 外层才把它移出 `running`:终态可见时
    /// `running_count` 可能还没归零,断言前要轮询。
    async fn wait_running_empty(sched: &Scheduler) {
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        while sched.running_count() != 0 {
            assert!(
                std::time::Instant::now() < deadline,
                "running_count 仍为 {}",
                sched.running_count()
            );
            tokio::time::sleep(Duration::from_millis(5)).await;
        }
    }

    async fn wait_terminal(sched: &Scheduler, id: &str) -> Task {
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        loop {
            let t = sched.storage().get_task(id).await.unwrap().unwrap();
            if t.status.is_terminal() {
                return t;
            }
            assert!(std::time::Instant::now() < deadline, "等待 {id} 终态超时");
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
    }

    /// 收事件直到 `done` 为 true(含那一条);5s 内没等到就 panic。
    async fn collect_events_until(
        rx: &mut broadcast::Receiver<TaskEvent>,
        done: impl Fn(&TaskEvent) -> bool,
    ) -> Vec<TaskEvent> {
        let mut v = Vec::new();
        loop {
            let ev = tokio::time::timeout(Duration::from_secs(5), rx.recv())
                .await
                .expect("等待任务事件超时")
                .expect("事件通道已关闭");
            let stop = done(&ev);
            v.push(ev);
            if stop {
                return v;
            }
        }
    }

    fn stages(t: &Task) -> Vec<&str> {
        t.stage_history.iter().map(|e| e.stage.as_str()).collect()
    }

    // ---- 成功路径 ----

    #[tokio::test]
    async fn success_path_pending_preparing_generating_completed() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(
                    TaskStage::Checking,
                    Some("是否继续生成?".into()),
                )),
                Step::Emit(CaptureEvent::Bound {
                    conversation_id: "7001".into(),
                    submission_id: "9002".into(),
                }),
                Step::Emit(CaptureEvent::Stage(TaskStage::Rendering, None)),
                Step::Emit(CaptureEvent::Stage(TaskStage::Result, None)),
                Step::Emit(CaptureEvent::Originals(vec![
                    "https://x.byteimg.com/a".into(),
                    "https://x.byteimg.com/b".into(),
                ])),
                Step::Ok(vec![
                    "https://x.byteimg.com/a".into(),
                    "https://x.byteimg.com/b".into(),
                ]),
            ],
        );
        let (sched, dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let mut rx = sched.subscribe();
        let _loop = sched.spawn_loop();

        let (id, pos, created_at) = sched.submit(image_req("张三")).await.unwrap();
        assert_eq!(pos, 1);
        assert!(created_at > 0);
        let t = wait_status(&sched, &id, TaskStatus::Completed).await;

        assert_eq!(t.result_files.as_ref().unwrap().len(), 2);
        let f0 = &t.result_files.as_ref().unwrap()[0];
        assert_eq!(f0.index, 0);
        assert_eq!(f0.filename, "image_1.png");
        assert_eq!(f0.relative_path, format!("results/{id}/image_1.png"));
        assert_eq!(f0.mime, "image/png");
        assert!(dir
            .path()
            .join("results")
            .join(&id)
            .join("image_2.png")
            .exists());
        assert_eq!(
            t.watermark_source,
            Some(WatermarkSource::ConversationOriginal)
        );
        assert_eq!(t.conversation_id.as_deref(), Some("7001"));
        assert_eq!(t.submission_id.as_deref(), Some("9002"));
        assert_eq!(t.local_message_id.as_deref(), Some("lm1"));
        assert!(t.started_at.is_some());
        assert!(t.completed_at.is_some());
        assert!(t.stage.is_none());
        assert!(t.error.is_none());
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "checking",
                "rendering",
                "result",
                "downloading",
                "completed"
            ]
        );
        assert_eq!(t.stage_history[3].message.as_deref(), Some("是否继续生成?"));
        assert!(backend.switches().is_empty(), "已是当前账号不应切换");
        assert_eq!(backend.downloads.load(Ordering::SeqCst), 2);

        // 事件:Created → Updated(preparing) … → Updated(completed) + Completed(task)
        let events = collect_events_until(&mut rx, |e| matches!(e, TaskEvent::Completed(_))).await;
        assert!(
            matches!(&events[0], TaskEvent::Created(s) if s.id == id && s.status == TaskStatus::Pending)
        );
        assert!(matches!(
            &events[1],
            TaskEvent::Updated {
                status: TaskStatus::Preparing,
                stage: Some(TaskStage::Preparing),
                ..
            }
        ));
        let statuses: Vec<TaskStatus> = events
            .iter()
            .filter_map(|e| match e {
                TaskEvent::Updated { status, .. } => Some(*status),
                _ => None,
            })
            .collect();
        assert_eq!(statuses.first(), Some(&TaskStatus::Preparing));
        assert!(statuses.contains(&TaskStatus::Generating));
        assert_eq!(statuses.last(), Some(&TaskStatus::Completed));
        assert!(
            matches!(events.last(), Some(TaskEvent::Completed(t)) if t.status == TaskStatus::Completed && t.id == id)
        );
        assert!(sched.queue_position(&id).await.is_none());
        wait_running_empty(&sched).await;
    }

    #[tokio::test]
    async fn runner_switches_account_when_current_differs() {
        let backend = MockBackend::new(&["张三", "李四"], vec![]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("李四")).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Completed).await;
        assert_eq!(backend.switches(), vec!["李四".to_string()]);
    }

    // ---- 失败 ----

    #[tokio::test]
    async fn capture_error_becomes_failed_with_code_and_account() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(TaskStage::QuotaChecking, None)),
                Step::Emit(CaptureEvent::Refused {
                    kind: "quota_exhausted".into(),
                    message: "额度不足".into(),
                }),
                Step::Fail(
                    RelayError::new("account_quota_exhausted", "账号配额已用完").with_details(
                        json!({ "reason": "quota_exhausted", "doubao_message": "额度不足" }),
                    ),
                ),
            ],
        );
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        let e = t.error.clone().unwrap();
        assert_eq!(e.code, "account_quota_exhausted");
        assert_eq!(e.details.as_ref().unwrap()["account"], "张三");
        assert_eq!(e.details.as_ref().unwrap()["doubao_message"], "额度不足");
        assert_eq!(t.stage, Some(TaskStage::Refused));
        assert_eq!(t.stage_message.as_deref(), Some("额度不足"));
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "quota_checking",
                "refused",
                "failed"
            ]
        );
        assert_eq!(
            t.stage_history.last().unwrap().message.as_deref(),
            Some("账号配额已用完")
        );
        assert!(t.completed_at.is_some());
    }

    #[tokio::test]
    async fn submit_error_fails_task_in_preparing() {
        let backend = MockBackend::with_submit(
            &["张三"],
            SubmitMode::Fail(generation::preparing_failed("skill_exit_failed", None)),
        );
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        assert_eq!(
            t.error.as_ref().unwrap().code,
            "generation_preparing_failed"
        );
        assert_eq!(stages(&t), vec!["pending", "preparing", "failed"]);
        assert_eq!(t.stage_message.as_deref(), Some("模型与参数已就绪"));
    }

    #[tokio::test]
    async fn preparing_timeout_fails_task() {
        let backend = MockBackend::with_submit(&["张三"], SubmitMode::Hang);
        let mut timing = fast_timing();
        timing.preparing_timeout = Duration::from_millis(100);
        let (sched, _dir) = setup(backend, timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        assert_eq!(t.error.unwrap().code, "generation_preparing_timeout");
    }

    #[tokio::test]
    async fn generating_timeout_fails_with_task_timeout() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let mut timing = fast_timing();
        timing.image_timeout = Duration::from_millis(150);
        let (sched, _dir) = setup(backend, timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        let e = t.error.unwrap();
        assert_eq!(e.code, "generation_task_timeout");
        assert_eq!(e.details.unwrap()["deadline_ms"], 150);
    }

    #[tokio::test]
    async fn generating_timeout_releases_account_when_capture_ignores_cancel() {
        let backend = MockBackend::new(&["张三"], vec![Step::IgnoreCancel]);
        let mut timing = fast_timing();
        // 留出 Windows CI/全量并行测试的调度抖动；验证的仍是“超时后立刻释放账号锁”。
        timing.image_timeout = Duration::from_millis(300);
        let (sched, _dir) = setup(backend, timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let (second, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Failed).await;
        wait_status(&sched, &second, TaskStatus::Generating).await;
    }

    #[tokio::test]
    async fn empty_originals_fails_with_result_not_captured() {
        let backend = MockBackend::new(&["张三"], vec![Step::Ok(vec![])]);
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        assert_eq!(t.error.unwrap().code, "generation_result_not_captured");
    }

    // ---- 取消 ----

    #[tokio::test]
    async fn pending_task_can_be_cancelled_immediately() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let mut rx = sched.subscribe();
        let _loop = sched.spawn_loop();
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let (second, pos, _) = sched.submit(image_req("张三")).await.unwrap();
        assert_eq!(pos, 1);
        assert_eq!(sched.queue_position(&second).await, Some(1));
        assert_eq!(
            sched
                .get_task(&second)
                .await
                .unwrap()
                .unwrap()
                .queue_position,
            Some(1)
        );

        let from = sched.cancel(&second).await.unwrap();
        assert_eq!(from, TaskStatus::Pending);
        let t = sched.storage().get_task(&second).await.unwrap().unwrap();
        assert_eq!(t.status, TaskStatus::Cancelled);
        let e = t.error.clone().unwrap();
        assert_eq!(e.code, "generation_cancelled");
        assert_eq!(e.details.unwrap()["from_status"], "pending");
        assert_eq!(stages(&t), vec!["pending", "cancelled"]);
        assert!(sched.queue_position(&second).await.is_none());

        // 再取消 → 不可取消
        let err = sched.cancel(&second).await.unwrap_err();
        assert_eq!(err.code, "task_not_cancellable");
        assert_eq!(err.details.unwrap()["current_status"], "cancelled");
        assert_eq!(
            sched.cancel("nope").await.unwrap_err().code,
            "task_not_found"
        );

        let events = collect_events_until(
            &mut rx,
            |e| matches!(e, TaskEvent::Completed(t) if t.id == second),
        )
        .await;
        assert!(
            matches!(events.last(), Some(TaskEvent::Completed(t)) if t.status == TaskStatus::Cancelled)
        );
        assert!(events
            .iter()
            .any(|e| matches!(e, TaskEvent::Created(s) if s.id == second)));
        assert!(events.iter().any(|e| matches!(e, TaskEvent::Updated { task_id, status: TaskStatus::Cancelled, .. } if *task_id == second)));

        backend.release(1);
        wait_status(&sched, &first, TaskStatus::Completed).await;
    }

    #[tokio::test]
    async fn running_task_cancel_becomes_cancelled() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Generating).await;
        let from = sched.cancel(&id).await.unwrap();
        assert_eq!(from, TaskStatus::Generating);
        let t = wait_status(&sched, &id, TaskStatus::Cancelled).await;
        let e = t.error.unwrap();
        assert_eq!(e.code, "generation_cancelled");
        assert_eq!(e.details.unwrap()["from_status"], "generating");
        assert_eq!(t.stage, Some(TaskStage::Sent));
        wait_running_empty(&sched).await;
    }

    // ---- 并发 ----

    #[tokio::test]
    async fn same_account_tasks_run_serially() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (a, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let (b, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &a, TaskStatus::Generating).await;
        tokio::time::sleep(Duration::from_millis(100)).await;
        let tb = sched.storage().get_task(&b).await.unwrap().unwrap();
        assert_eq!(
            tb.status,
            TaskStatus::Pending,
            "同账号第二个任务应保持 pending"
        );
        assert_eq!(backend.captures_started.load(Ordering::SeqCst), 1);
        assert_eq!(sched.queue_position(&b).await, Some(1));

        backend.release(1);
        wait_status(&sched, &a, TaskStatus::Completed).await;
        wait_status(&sched, &b, TaskStatus::Generating).await;
        backend.release(1);
        wait_status(&sched, &b, TaskStatus::Completed).await;
        assert_eq!(backend.max_running.load(Ordering::SeqCst), 1);
    }

    /// Task 4.2:5 个同账号任务排队位置 1..5,严格 FIFO 串行(任意时刻至多 1 个在跑)。
    #[tokio::test]
    async fn five_same_account_tasks_queue_1_to_5_and_run_serially_in_order() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let mut ids = Vec::new();
        for want in 1..=5u32 {
            let (id, pos, _) = sched.submit(image_req("张三")).await.unwrap();
            assert_eq!(pos, want);
            ids.push(id);
        }
        for (i, id) in ids.iter().enumerate() {
            assert_eq!(sched.queue_position(id).await, Some(i as u32 + 1));
        }
        let _loop = sched.spawn_loop();
        for (i, id) in ids.iter().enumerate() {
            wait_status(&sched, id, TaskStatus::Generating).await;
            tokio::time::sleep(Duration::from_millis(30)).await;
            assert_eq!(sched.running_count(), 1);
            for (j, later) in ids.iter().enumerate().skip(i + 1) {
                assert_eq!(
                    sched.queue_position(later).await,
                    Some((j - i) as u32),
                    "第 {} 个运行时第 {} 个的排队位置",
                    i + 1,
                    j + 1
                );
            }
            backend.release(1);
            wait_status(&sched, id, TaskStatus::Completed).await;
        }
        assert_eq!(backend.max_running.load(Ordering::SeqCst), 1);
        let started: Vec<i64> = {
            let mut v = Vec::new();
            for id in &ids {
                v.push(
                    sched
                        .storage()
                        .get_task(id)
                        .await
                        .unwrap()
                        .unwrap()
                        .started_at
                        .unwrap(),
                );
            }
            v
        };
        assert!(
            started.windows(2).all(|w| w[0] <= w[1]),
            "按入队顺序开始:{started:?}"
        );
    }

    /// Task 4.2(Task 2.7 已知问题 ⑦):豆包只有一个"当前账号",切账号整页 reload,
    /// 所以不同账号的任务也不能同时跑——队头任务的账号与运行中任务不同就等它们全部结束
    /// (spec 08 的 global_max=3 / UC-24 在单个豆包实例上做不到,见计划"已知问题")。
    #[tokio::test]
    async fn different_accounts_are_serialized_on_a_single_doubao() {
        let backend = MockBackend::new(&["a", "b", "c", "d"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |c| {
            c.concurrency.global_max = 3;
        })
        .await;
        let _loop = sched.spawn_loop();
        let mut ids = Vec::new();
        for acc in ["a", "b", "c", "d"] {
            ids.push(sched.submit(image_req(acc)).await.unwrap().0);
        }
        for (i, id) in ids.iter().enumerate() {
            wait_status(&sched, id, TaskStatus::Generating).await;
            tokio::time::sleep(Duration::from_millis(50)).await;
            assert_eq!(sched.running_count(), 1, "第 {} 个运行时", i + 1);
            assert_eq!(backend.captures_started.load(Ordering::SeqCst), i + 1);
            for later in &ids[i + 1..] {
                assert_eq!(
                    sched
                        .storage()
                        .get_task(later)
                        .await
                        .unwrap()
                        .unwrap()
                        .status,
                    TaskStatus::Pending
                );
            }
            backend.release(1);
            wait_status(&sched, id, TaskStatus::Completed).await;
        }
        assert_eq!(backend.max_running.load(Ordering::SeqCst), 1);
        // 每次切换都发生在前一个任务结束之后,且不会被切回
        assert_eq!(backend.switches(), vec!["b", "c", "d"]);
    }

    /// 跨账号串行之外 `global_max` 仍然生效(热改配置为 0 → 不派发)。
    #[tokio::test]
    async fn global_max_still_caps_dispatch() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |c| {
            c.concurrency.global_max = 0;
        })
        .await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        tokio::time::sleep(Duration::from_millis(100)).await;
        assert_eq!(
            sched.storage().get_task(&id).await.unwrap().unwrap().status,
            TaskStatus::Pending,
            "global_max=0 不派发"
        );
        sched.cancel(&id).await.unwrap();
    }

    // ---- 切账号互斥(spec 08 "账号互斥" switch_lock)----

    /// 切换进行中提交的任务保持 pending,切换结束(guard 释放)后才进入 preparing(UC-31)。
    #[tokio::test]
    async fn switch_in_progress_blocks_dispatch_until_finished() {
        let backend = MockBackend::new(&["张三", "李四"], vec![]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let guard = sched.begin_switch("李四").unwrap();
        assert!(sched.switching());
        let (id, pos, _) = sched.submit(image_req("李四")).await.unwrap();
        assert_eq!(pos, 1);
        tokio::time::sleep(Duration::from_millis(150)).await;
        let t = sched.storage().get_task(&id).await.unwrap().unwrap();
        assert_eq!(t.status, TaskStatus::Pending, "切换中不应派发");
        assert_eq!(backend.captures_started.load(Ordering::SeqCst), 0);
        // 模拟切换完成(手动切换把当前账号改成了李四)
        *backend.current.lock().unwrap() = Some("李四".into());
        let released_at = now_ms();
        drop(guard);
        assert!(!sched.switching());
        let t = wait_status(&sched, &id, TaskStatus::Completed).await;
        assert!(t.started_at.unwrap() >= released_at);
        assert!(
            backend.switches().is_empty(),
            "已是当前账号,runner 不再切换"
        );
    }

    /// 有任务在跑时拒绝开始切换(`account_busy`,409);切换进行中再开始一个切换也允许排队(计数)。
    #[tokio::test]
    async fn begin_switch_is_refused_while_a_task_runs() {
        let backend = MockBackend::new(&["张三", "李四"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Generating).await;
        let err = sched.begin_switch("李四").err().unwrap();
        assert_eq!(err.code, "account_busy");
        assert_eq!(err.http_status(), 409);
        assert_eq!(
            err.details.as_ref().unwrap()["running_task_ids"],
            json!([id])
        );
        assert!(!sched.switching(), "被拒绝的切换不留下标记");
        backend.release(1);
        wait_status(&sched, &id, TaskStatus::Completed).await;
        let deadline = std::time::Instant::now() + Duration::from_secs(2);
        while sched.running_count() > 0 {
            assert!(std::time::Instant::now() < deadline);
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let g1 = sched.begin_switch("李四").unwrap();
        let g2 = sched.begin_switch("张三").unwrap();
        drop(g1);
        assert!(sched.switching(), "还有一个切换没结束");
        drop(g2);
        assert!(!sched.switching());
    }

    // ---- 就绪门槛 / 干扰 ----

    /// 豆包未就绪(`DoubaoStatus.cdp_connected=false`,崩溃后)→ `/generate` 503,不入队。
    #[tokio::test]
    async fn submit_is_503_when_doubao_not_ready() {
        let backend = MockBackend::new(&["张三"], vec![]);
        *backend.ready_err.lock().unwrap() = Some(RelayError::doubao_not_ready());
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let err = sched.submit(image_req("张三")).await.unwrap_err();
        assert_eq!(err.code, "doubao_not_ready");
        assert_eq!(err.http_status(), 503);
        assert_eq!(sched.queue_depth(), 0);
        // 请求本身不合法时仍先报 400
        let mut bad = image_req("张三");
        bad.prompt.clear();
        assert_eq!(
            sched.submit(bad).await.unwrap_err().code,
            "request_missing_field"
        );
        *backend.ready_err.lock().unwrap() = None;
        sched.submit(image_req("张三")).await.unwrap();
    }

    /// 输入框被用户改动:只更新 `message="检测到输入框被修改"`(task_updated),不终止任务。
    #[tokio::test]
    async fn input_modified_updates_message_without_failing() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(TaskStage::Rendering, None)),
                Step::Emit(CaptureEvent::InputModified { text_length: 4 }),
                Step::Ok(vec!["https://x.byteimg.com/a".into()]),
            ],
        );
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let mut rx = sched.subscribe();
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let events = collect_events_until(&mut rx, |e| matches!(e, TaskEvent::Completed(_))).await;
        assert!(events.iter().any(|e| matches!(
            e,
            TaskEvent::Updated { status: TaskStatus::Generating, stage: Some(TaskStage::Rendering), message: Some(m), .. }
                if m == INPUT_MODIFIED_MESSAGE
        )));
        let t = wait_status(&sched, &id, TaskStatus::Completed).await;
        // 只改消息不追加阶段历史
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "rendering",
                "downloading",
                "completed"
            ]
        );
    }

    #[tokio::test]
    async fn queue_full_at_depth_max() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(backend, fast_timing(), |c| {
            c.concurrency.queue_depth_max = 50;
        })
        .await;
        // 不启动派发循环:50 个都停在 pending
        for i in 0..50 {
            sched
                .submit(image_req("张三"))
                .await
                .unwrap_or_else(|e| panic!("第 {} 个提交失败:{e}", i + 1));
        }
        assert_eq!(sched.queue_depth(), 50);
        let err = sched.submit(image_req("张三")).await.unwrap_err();
        assert_eq!(err.code, "queue_full");
        assert_eq!(err.http_status(), 429);
        let d = err.details.unwrap();
        assert_eq!(d["depth"], 50);
        assert_eq!(d["max"], 50);
    }

    // ---- manual 确认 ----

    #[tokio::test]
    async fn manual_mode_waits_for_confirm_then_continues() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(
                    TaskStage::Checking,
                    Some("是否继续生成?".into()),
                )),
                Step::Emit(CaptureEvent::NeedsUserConfirm),
                Step::WaitConfirm,
                Step::Emit(CaptureEvent::Stage(TaskStage::Rendering, None)),
                Step::Emit(CaptureEvent::Stage(TaskStage::QuotaChecking, None)),
                Step::Emit(CaptureEvent::NeedsUserConfirm),
                Step::WaitConfirm,
                Step::Emit(CaptureEvent::Stage(TaskStage::Rendering, None)),
                Step::Emit(CaptureEvent::Stage(TaskStage::Result, None)),
            ],
        );
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.confirmation_mode = ConfirmationMode::Manual;
        req.auto_use_membership = false;
        let (id, _, _) = sched.submit(req).await.unwrap();

        let t = wait_status(&sched, &id, TaskStatus::WaitingUserConfirm).await;
        assert_eq!(t.stage, Some(TaskStage::Checking));
        assert_eq!(sched.queue_depth(), 1, "waiting 计入队列深度");
        assert_eq!(sched.confirm(&id).await.unwrap(), "generating");
        // 确认后回到 generating(mock 随即发 rendering → quota_checking → 再次等待确认,
        // generating 窗口极短,不单独轮询它,由 stage_history 断言)

        // 第二次确认(quota_checking)
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        loop {
            let t = sched.storage().get_task(&id).await.unwrap().unwrap();
            if t.status == TaskStatus::WaitingUserConfirm
                && t.stage == Some(TaskStage::QuotaChecking)
            {
                break;
            }
            assert!(std::time::Instant::now() < deadline, "等待第二次确认超时");
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        assert_eq!(sched.confirm(&id).await.unwrap(), "generating");
        let t = wait_status(&sched, &id, TaskStatus::Completed).await;
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "checking",
                "waiting_user_confirm",
                "generating",
                "rendering",
                "quota_checking",
                "waiting_user_confirm",
                "generating",
                "rendering",
                "result",
                "downloading",
                "completed"
            ]
        );
        // 终态后 confirm → task_not_waiting_confirm
        let err = sched.confirm(&id).await.unwrap_err();
        assert_eq!(err.code, "task_not_waiting_confirm");
        assert_eq!(err.details.unwrap()["current_status"], "completed");
        assert_eq!(
            sched.confirm("nope").await.unwrap_err().code,
            "task_not_found"
        );
    }

    #[tokio::test]
    async fn confirm_while_generating_is_rejected() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Generating).await;
        let err = sched.confirm(&id).await.unwrap_err();
        assert_eq!(err.code, "task_not_waiting_confirm");
        assert_eq!(err.details.unwrap()["current_status"], "generating");
        backend.release(1);
        wait_status(&sched, &id, TaskStatus::Completed).await;
    }

    #[tokio::test]
    async fn confirm_timeout_fails_task() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(TaskStage::Checking, None)),
                Step::Emit(CaptureEvent::NeedsUserConfirm),
                Step::WaitConfirm,
            ],
        );
        let mut timing = fast_timing();
        timing.confirm_timeout = Duration::from_millis(150);
        let (sched, _dir) = setup(backend, timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.confirmation_mode = ConfirmationMode::Manual;
        let (id, _, _) = sched.submit(req).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Failed).await;
        let e = t.error.unwrap();
        assert_eq!(e.code, "generation_confirm_timeout");
        assert_eq!(e.details.unwrap()["stage"], "checking");
    }

    #[tokio::test]
    async fn waiting_for_confirm_does_not_consume_generating_budget() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::NeedsUserConfirm),
                Step::WaitConfirm,
                Step::Sleep(120),
            ],
        );
        let mut timing = fast_timing();
        timing.image_timeout = Duration::from_millis(200);
        let (sched, _dir) = setup(backend, timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.confirmation_mode = ConfirmationMode::Manual;
        let (id, _, _) = sched.submit(req).await.unwrap();
        wait_status(&sched, &id, TaskStatus::WaitingUserConfirm).await;
        // 等待期超过生成预算,之后确认并再跑 120ms,总 generating 时间 < 200ms → 不应超时
        tokio::time::sleep(Duration::from_millis(400)).await;
        assert_eq!(
            sched.storage().get_task(&id).await.unwrap().unwrap().status,
            TaskStatus::WaitingUserConfirm
        );
        sched.confirm(&id).await.unwrap();
        let t = wait_terminal(&sched, &id).await;
        assert_eq!(t.status, TaskStatus::Completed, "{:?}", t.error);
    }

    // ---- 超时扫描 / 崩溃 ----

    #[tokio::test]
    async fn pending_timeout_scan_fails_queued_task() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let mut timing = fast_timing();
        timing.pending_timeout = Duration::from_millis(80);
        let (sched, _dir) = setup(Arc::clone(&backend), timing, |_| {}).await;
        let _loop = sched.spawn_loop();
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let (second, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &second, TaskStatus::Failed).await;
        let e = t.error.unwrap();
        assert_eq!(e.code, "generation_queue_timeout");
        assert!(e.details.unwrap()["waited_ms"].as_i64().unwrap() >= 80);
        assert!(sched.queue_position(&second).await.is_none());
        backend.release(1);
        wait_status(&sched, &first, TaskStatus::Completed).await;
    }

    #[tokio::test]
    async fn fail_all_running_marks_running_and_pending_failed() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let (second, _, _) = sched.submit(image_req("张三")).await.unwrap();

        sched
            .fail_all_running("doubao_crashed", "豆包已退出,任务作废")
            .await;
        let t1 = wait_status(&sched, &first, TaskStatus::Failed).await;
        assert_eq!(t1.error.unwrap().code, "doubao_crashed");
        let t2 = wait_status(&sched, &second, TaskStatus::Failed).await;
        assert_eq!(t2.error.unwrap().code, "doubao_crashed");
        assert_eq!(sched.queue_depth(), 0);
    }

    /// 崩溃中止带 `details`(`doubao_crashed{last_seen_at}`),运行中 / 排队中都写入。
    #[tokio::test]
    async fn fail_all_running_with_details_records_them() {
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let (second, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let e = RelayError::doubao_crashed(1234);
        sched
            .fail_all_running_with(e.code, &e.message, e.details.clone())
            .await;
        for id in [&first, &second] {
            let t = wait_status(&sched, id, TaskStatus::Failed).await;
            let err = t.error.unwrap();
            assert_eq!(err.code, "doubao_crashed");
            assert_eq!(err.details.unwrap()["last_seen_at"], 1234);
        }
        wait_running_empty(&sched).await;
    }

    // ---- 提交校验 ----

    #[tokio::test]
    async fn submit_rejects_invalid_request_unknown_account_and_bad_uploads() {
        let backend = MockBackend::new(&["张三"], vec![]);
        let (sched, dir) = setup(backend, fast_timing(), |_| {}).await;

        let mut bad = image_req("张三");
        bad.prompt = String::new();
        let err = sched.submit(bad).await.unwrap_err();
        assert_eq!(err.code, "request_missing_field");

        let err = sched.submit(image_req("王五")).await.unwrap_err();
        assert_eq!(err.code, "account_not_found");
        let d = err.details.unwrap();
        assert_eq!(d["requested"], "王五");
        assert_eq!(d["available"], json!(["张三"]));

        let mut refs = image_req("张三");
        refs.reference_files = vec!["01HQZ9MISSING".into()];
        let err = sched.submit(refs).await.unwrap_err();
        assert_eq!(err.code, "upload_not_found");
        assert_eq!(err.details.unwrap()["upload_id"], "01HQZ9MISSING");

        let now = now_ms();
        sched
            .storage()
            .insert_upload(&Upload {
                id: "01HQZ9OLD".into(),
                mime: "image/png".into(),
                bytes: 1,
                path: dir.path().join("uploads/01HQZ9OLD.png"),
                created_at: now - 10_000,
                expires_at: now - 1,
            })
            .await
            .unwrap();
        let mut refs = image_req("张三");
        refs.reference_files = vec!["01HQZ9OLD".into()];
        let err = sched.submit(refs).await.unwrap_err();
        assert_eq!(err.code, "upload_expired");
        assert_eq!(err.details.unwrap()["upload_id"], "01HQZ9OLD");

        // 没有任何任务入库
        let (list, _) = sched
            .storage()
            .list_tasks(None, None, 10, None)
            .await
            .unwrap();
        assert!(list.is_empty());
    }

    // ---- 视频 ----

    #[tokio::test]
    async fn video_task_downloads_one_unwatermarked_mp4_with_duration() {
        let url = "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked";
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(TaskStage::Result, None)),
                Step::OkVideo(url.into(), WatermarkSource::FallbackDecrypted),
            ],
        );
        let timing = fast_timing();
        let (sched, dir) = setup(Arc::clone(&backend), timing.clone(), |_| {}).await;
        sched.spawn_loop();
        let (id, _, _) = sched.submit(video_req("张三")).await.unwrap();
        let t = wait_terminal(&sched, &id).await;
        assert_eq!(t.status, TaskStatus::Completed, "{:?}", t.error);
        assert_eq!(t.watermark_source, Some(WatermarkSource::FallbackDecrypted));
        let files = t.result_files.clone().unwrap();
        assert_eq!(files.len(), 1);
        let f = &files[0];
        assert_eq!(f.filename, "video_1.mp4");
        assert_eq!(f.mime, "video/mp4");
        assert_eq!(f.relative_path, format!("results/{id}/video_1.mp4"));
        assert_eq!(f.duration_ms, Some(5000));
        assert_eq!(
            (f.width, f.height),
            (Some(1920), Some(1080)),
            "取视频轨 tkhd 宽高"
        );
        assert!(dir.path().join(&f.relative_path).exists());
        assert_eq!(
            *backend.video_downloads.lock().unwrap(),
            vec![url.to_string()]
        );
        assert_eq!(
            backend.downloads.load(Ordering::SeqCst),
            0,
            "视频不走图像下载"
        );
        // capture 用视频预算(20min 由 timing 决定)+ 两次确认等待
        let opts = backend.last_capture.lock().unwrap().clone().unwrap();
        assert_eq!(opts.task_type, TaskType::Video);
        assert_eq!(
            opts.deadline,
            timing.video_timeout + timing.confirm_timeout * 2
        );
        assert_eq!(opts.model_name.as_deref(), Some("Seedance 2.0 Fast"));
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "result",
                "downloading",
                "completed"
            ]
        );
    }

    #[tokio::test]
    async fn video_cancel_during_download_ends_cancelled_without_files() {
        let url = "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked";
        let backend = MockBackend::new(
            &["张三"],
            vec![Step::OkVideo(url.into(), WatermarkSource::LrUnwatermarked)],
        );
        backend.video_hang.store(true, Ordering::SeqCst);
        let (sched, dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        sched.spawn_loop();
        let (id, _, _) = sched.submit(video_req("张三")).await.unwrap();
        // 等到 download_video 已被调用(处于 downloading 阶段、半截文件已写)
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        while backend.video_downloads.lock().unwrap().is_empty() {
            assert!(std::time::Instant::now() < deadline, "未进入视频下载");
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let partial = dir.path().join(format!("results/{id}/video_1.mp4"));
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        while !partial.exists() {
            assert!(std::time::Instant::now() < deadline, "半截文件未写出");
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let running = sched.storage().get_task(&id).await.unwrap().unwrap();
        assert_eq!(running.stage, Some(TaskStage::Downloading));
        sched.cancel(&id).await.unwrap();
        let t = wait_terminal(&sched, &id).await;
        assert_eq!(t.status, TaskStatus::Cancelled, "{:?}", t.error);
        let e = t.error.unwrap();
        assert_eq!(e.code, "generation_cancelled");
        assert_eq!(e.details.unwrap()["step"], "download");
        assert!(t.result_files.is_none());
        assert!(!partial.exists(), "取消后删除半截 video_1.mp4");
    }

    #[tokio::test]
    async fn video_without_lr_unwatermarked_fails_with_watermark_not_removed() {
        let backend = MockBackend::new(
            &["张三"],
            vec![Step::OkVideo(
                "https://v3-vdl.doubao.com/wm.mp4".into(),
                WatermarkSource::LrUnwatermarked,
            )],
        );
        let (sched, _dir) = setup(Arc::clone(&backend), fast_timing(), |_| {}).await;
        sched.spawn_loop();
        let (id, _, _) = sched.submit(video_req("张三")).await.unwrap();
        let t = wait_terminal(&sched, &id).await;
        assert_eq!(t.status, TaskStatus::Failed);
        assert_eq!(t.error.unwrap().code, "generation_watermark_not_removed");
        assert!(t.result_files.is_none());
    }

    #[test]
    fn default_timing_gives_video_20_minutes() {
        let t = SchedulerTiming::default();
        assert_eq!(t.video_timeout, Duration::from_secs(20 * 60));
        assert_eq!(t.image_timeout, Duration::from_secs(15 * 60));
    }

    #[tokio::test]
    async fn submit_propagates_doubao_errors_from_accounts() {
        struct Down;
        #[async_trait]
        impl GenerationBackend for Down {
            async fn accounts(&self, _: bool) -> RelayResult<AccountList> {
                Err(RelayError::doubao_cdp_disconnected())
            }
            async fn ensure_account(&self, _: &str) -> RelayResult<()> {
                unreachable!()
            }
            async fn submit(
                &self,
                _: &GenerateRequest,
                _: &SubmissionPlan,
                _: &[PathBuf],
                _: &CancellationToken,
                _: StageFn<'_>,
            ) -> RelayResult<SubmissionReport> {
                unreachable!()
            }
            async fn capture(
                &self,
                _: CaptureOptions,
                _: &CancellationToken,
                _: &mut watch::Receiver<bool>,
                _: mpsc::Sender<CaptureEvent>,
            ) -> RelayResult<CaptureOutcome> {
                unreachable!()
            }
            async fn download_image(&self, _: &str, _: &Path, _: u32) -> RelayResult<Downloaded> {
                unreachable!()
            }
            async fn download_video(&self, _: &str, _: &Path) -> RelayResult<Downloaded> {
                unreachable!()
            }
        }
        let dir = tempfile::tempdir().unwrap();
        let storage = Storage::open(dir.path()).await.unwrap();
        let sched = Scheduler::new(
            storage,
            Arc::new(Down),
            Arc::new(RwLock::new(Config::default())),
        );
        let err = sched.submit(image_req("张三")).await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
        assert_eq!(err.http_status(), 503);
    }

    // ---- Task 4.1:auto 降级 manual 的状态机 + 终态 webhook ----

    /// capture 在 auto 模式下代回复失败 → 降级 manual(NeedsUserConfirm)→ 用户确认 → capture 代回复发出
    /// (`Stage(Sent,"已按用户确认继续")`)→ 完成;阶段历史与消息逐条可查。
    #[tokio::test]
    async fn auto_degraded_to_manual_waits_confirm_then_records_sent() {
        let backend = MockBackend::new(
            &["张三"],
            vec![
                Step::Emit(CaptureEvent::Stage(
                    TaskStage::Checking,
                    Some("是否继续生成?".into()),
                )),
                // 自动代回复失败,capture 降级
                Step::Emit(CaptureEvent::NeedsUserConfirm),
                Step::WaitConfirm,
                Step::Emit(CaptureEvent::Stage(
                    TaskStage::Sent,
                    Some(crate::doubao::capture::USER_CONTINUE_MESSAGE.into()),
                )),
                Step::Emit(CaptureEvent::Stage(TaskStage::Rendering, None)),
                Step::Emit(CaptureEvent::Stage(TaskStage::Result, None)),
            ],
        );
        let (sched, _dir) = setup(backend, fast_timing(), |_| {}).await;
        let _loop = sched.spawn_loop();
        let (id, _, _) = sched.submit(image_req("张三")).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::WaitingUserConfirm).await;
        assert_eq!(t.stage, Some(TaskStage::Checking));
        assert_eq!(t.request.confirmation_mode, ConfirmationMode::Auto);
        sched.confirm(&id).await.unwrap();
        let t = wait_status(&sched, &id, TaskStatus::Completed).await;
        assert_eq!(
            stages(&t),
            vec![
                "pending",
                "preparing",
                "sent",
                "checking",
                "waiting_user_confirm",
                "generating",
                "sent",
                "rendering",
                "result",
                "downloading",
                "completed"
            ]
        );
        let resent = &t.stage_history[6];
        assert_eq!(
            resent.message.as_deref(),
            Some(crate::doubao::capture::USER_CONTINUE_MESSAGE)
        );
    }

    /// mock 回调服务器收到的一次请求:(头, body)。
    type Hit = (axum::http::HeaderMap, Vec<u8>);

    /// 起一个 axum 回调服务器,返回 (callback_url, 解析 `hook.test` 到它的 client, 收到的请求)。
    /// `callback_url` 校验拒绝 loopback,所以用域名 + `ClientBuilder::resolve`。
    async fn spawn_hook(status: u16) -> (String, reqwest::Client, Arc<StdMutex<Vec<Hit>>>) {
        use axum::{extract::State, http::StatusCode, routing::post, Router};
        let hits: Arc<StdMutex<Vec<Hit>>> = Arc::default();
        async fn hook(
            State((hits, status)): State<(Arc<StdMutex<Vec<Hit>>>, u16)>,
            headers: axum::http::HeaderMap,
            body: axum::body::Bytes,
        ) -> StatusCode {
            hits.lock().unwrap().push((headers, body.to_vec()));
            StatusCode::from_u16(status).unwrap()
        }
        let app = Router::new()
            .route("/hook", post(hook))
            .with_state((Arc::clone(&hits), status));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        let client = reqwest::Client::builder()
            .redirect(reqwest::redirect::Policy::none())
            .resolve("hook.test", addr)
            .build()
            .unwrap();
        (
            format!("http://hook.test:{}/hook", addr.port()),
            client,
            hits,
        )
    }

    fn webhook_timing() -> SchedulerTiming {
        SchedulerTiming {
            webhook: crate::tasks::webhook::WebhookPolicy {
                timeout: Duration::from_secs(2),
                retry_delays: vec![Duration::from_millis(10); 3],
            },
            ..fast_timing()
        }
    }

    async fn setup_hook(
        backend: Arc<MockBackend>,
        client: reqwest::Client,
    ) -> (Arc<Scheduler>, tempfile::TempDir) {
        let dir = tempfile::tempdir().unwrap();
        let storage = Storage::open(dir.path()).await.unwrap();
        let mut cfg = Config::default();
        cfg.server.bearer_token = "tok-4.1".into();
        let sched = Scheduler::new_with_webhook_client(
            storage,
            backend as Arc<dyn GenerationBackend>,
            Arc::new(RwLock::new(cfg)),
            webhook_timing(),
            client,
        );
        (sched, dir)
    }

    async fn wait_hits(hits: &Arc<StdMutex<Vec<Hit>>>, n: usize) -> Vec<Hit> {
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        loop {
            let v = hits.lock().unwrap().clone();
            if v.len() >= n {
                return v;
            }
            assert!(
                std::time::Instant::now() < deadline,
                "等待 {n} 次回调超时,已收到 {}",
                v.len()
            );
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
    }

    async fn webhook_log_count(sched: &Scheduler, task_id: &str) -> i64 {
        sqlx::query_scalar("SELECT COUNT(*) FROM webhook_log WHERE task_id = ?")
            .bind(task_id)
            .fetch_one(&sched.storage().pool)
            .await
            .unwrap()
    }

    fn check_hit(hit: &Hit, event: &str, status: &str, task_id: &str) -> Value {
        let (headers, body) = hit;
        let ts: i64 = headers["x-doubao-relay-timestamp"]
            .to_str()
            .unwrap()
            .parse()
            .unwrap();
        assert_eq!(
            headers["x-doubao-relay-signature"].to_str().unwrap(),
            crate::tasks::webhook::signature_header("tok-4.1", ts, body)
        );
        let v: Value = serde_json::from_slice(body).unwrap();
        assert_eq!(v["task_id"], task_id);
        assert_eq!(v["event"], event);
        assert_eq!(v["status"], status);
        assert_eq!(v["task"]["status"], status);
        v
    }

    #[tokio::test]
    async fn completed_task_with_callback_url_posts_signed_webhook() {
        let (url, client, hits) = spawn_hook(200).await;
        let backend = MockBackend::new(&["张三"], vec![]);
        let (sched, _dir) = setup_hook(backend, client).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.callback_url = Some(url.clone());
        let (id, _, _) = sched.submit(req).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Completed).await;
        let got = wait_hits(&hits, 1).await;
        let v = check_hit(&got[0], "task.completed", "completed", &id);
        // body 里的任务已是终态(含结果文件与完整阶段历史)
        // `task` 字段即 GET /tasks/{id} 的响应体(spec 04)
        let files = v["task"]["result"]["files"].as_array().unwrap();
        assert_eq!(files.len(), 1);
        assert_eq!(
            files[0]["download_url"],
            format!("/api/v1/tasks/{id}/download/0")
        );
        assert_eq!(v["task"]["account"], "张三");
        assert_eq!(v["task"]["type"], "image");
        let stored = sched.storage().get_task(&id).await.unwrap().unwrap();
        assert_eq!(
            v["task"]["result"]["watermark_source"],
            serde_json::to_value(stored.watermark_source).unwrap()
        );
        assert_eq!(
            v["task"]["stage_history"]
                .as_array()
                .unwrap()
                .last()
                .unwrap()["stage"],
            "completed"
        );
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        while webhook_log_count(&sched, &id).await < 1 {
            assert!(std::time::Instant::now() < deadline);
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        tokio::time::sleep(Duration::from_millis(100)).await;
        assert_eq!(hits.lock().unwrap().len(), 1, "成功后不重试");
        assert_eq!(webhook_log_count(&sched, &id).await, 1);
    }

    #[tokio::test]
    async fn failed_task_posts_task_failed_with_error() {
        let (url, client, hits) = spawn_hook(200).await;
        let backend = MockBackend::new(
            &["张三"],
            vec![Step::Fail(RelayError::new(
                "generation_doubao_refused",
                "豆包拒绝了本次生成",
            ))],
        );
        let (sched, _dir) = setup_hook(backend, client).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.callback_url = Some(url);
        let (id, _, _) = sched.submit(req).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Failed).await;
        let got = wait_hits(&hits, 1).await;
        let v = check_hit(&got[0], "task.failed", "failed", &id);
        assert_eq!(v["task"]["error"]["code"], "generation_doubao_refused");
        assert!(v["task"]["failed_at"].is_i64(), "{v}");
        assert_eq!(v["task"]["account"], "张三");
    }

    /// 不经 runner 的终态(pending 取消)同样回调;没带 callback_url 的任务不回调。
    #[tokio::test]
    async fn pending_cancel_posts_task_cancelled_and_no_url_means_no_callback() {
        let (url, client, hits) = spawn_hook(200).await;
        let backend = MockBackend::new(&["张三"], vec![Step::Gate]);
        let (sched, _dir) = setup_hook(Arc::clone(&backend), client).await;
        let _loop = sched.spawn_loop();
        // 第一个任务占住账号锁(无 callback_url),第二个排队后取消
        let (first, _, _) = sched.submit(image_req("张三")).await.unwrap();
        wait_status(&sched, &first, TaskStatus::Generating).await;
        let mut req = image_req("张三");
        req.callback_url = Some(url);
        let (second, _, _) = sched.submit(req).await.unwrap();
        assert_eq!(sched.cancel(&second).await.unwrap(), TaskStatus::Pending);
        let got = wait_hits(&hits, 1).await;
        let v = check_hit(&got[0], "task.cancelled", "cancelled", &second);
        assert_eq!(v["task"]["error"]["code"], "generation_cancelled");

        backend.release(1);
        wait_status(&sched, &first, TaskStatus::Completed).await;
        tokio::time::sleep(Duration::from_millis(150)).await;
        assert_eq!(
            hits.lock().unwrap().len(),
            1,
            "无 callback_url 的任务不回调"
        );
        assert_eq!(webhook_log_count(&sched, &first).await, 0);
    }

    /// 回调全部失败:按策略重试(共 4 次),失败不影响任务终态,任务上记 `callback_failed`。
    #[tokio::test]
    async fn webhook_failures_are_retried_and_recorded_without_changing_status() {
        let (url, client, hits) = spawn_hook(500).await;
        let backend = MockBackend::new(&["张三"], vec![]);
        let (sched, _dir) = setup_hook(backend, client).await;
        let _loop = sched.spawn_loop();
        let mut req = image_req("张三");
        req.callback_url = Some(url);
        let (id, _, _) = sched.submit(req).await.unwrap();
        wait_status(&sched, &id, TaskStatus::Completed).await;
        wait_hits(&hits, 4).await;
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        loop {
            let t = sched.storage().get_task(&id).await.unwrap().unwrap();
            if stages(&t).contains(&crate::tasks::webhook::CALLBACK_FAILED_STAGE) {
                assert_eq!(t.status, TaskStatus::Completed);
                break;
            }
            assert!(
                std::time::Instant::now() < deadline,
                "未记录 callback_failed"
            );
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        assert_eq!(webhook_log_count(&sched, &id).await, 4);
    }

    #[test]
    fn error_constructors_have_expected_codes_and_statuses() {
        assert_eq!(queue_full(50, 50).http_status(), 429);
        assert_eq!(task_not_found("x").http_status(), 404);
        assert_eq!(
            task_not_cancellable(TaskStatus::Completed).http_status(),
            409
        );
        assert_eq!(
            task_not_waiting_confirm(TaskStatus::Generating).http_status(),
            409
        );
        assert_eq!(upload_not_found("u").http_status(), 404);
        assert_eq!(upload_expired("u", 1).http_status(), 400);
        assert_eq!(queue_timeout(1).code, "generation_queue_timeout");
        let busy = account_busy("李四", vec!["t1".into()]);
        assert_eq!(busy.http_status(), 409);
        let d = busy.details.clone().unwrap();
        assert_eq!(d["account"], "李四");
        assert_eq!(d["running_task_ids"], json!(["t1"]));
        for e in [
            queue_full(1, 1),
            task_not_found("x"),
            task_not_cancellable(TaskStatus::Failed),
            task_not_waiting_confirm(TaskStatus::Pending),
            upload_not_found("u"),
            upload_expired("u", 1),
            queue_timeout(1),
            busy,
        ] {
            let last = e.message.chars().last().unwrap();
            assert!(
                !matches!(last, '。' | '.' | '!' | ',' | '\u{FF0C}'),
                "{}",
                e.message
            );
        }
    }
}
