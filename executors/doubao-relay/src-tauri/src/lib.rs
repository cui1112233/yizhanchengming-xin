//! doubao-relay 后端入口。
//!
//! `run()`:platform → config → logging → `DoubaoConnection` → `Storage` → `Scheduler` → `AppState`
//! → Tauri Builder(`manage(state)`,`setup` 里起事件转发、心跳、按配置自动连豆包,
//! 以及 Phase 2 的启动恢复(`relay_restart`)、调度循环、豆包崩溃处理、每小时 GC、HTTP API server)。

pub mod cdp;
pub mod commands;
pub mod config;
pub mod doubao;
pub mod error;
pub mod http;
pub mod logging;
pub mod platform;
pub mod shutdown; // Task 5.2
pub mod state;
pub mod storage;
pub mod tasks;
pub mod tray; // Task 5.2
pub mod v88_executor;

use std::sync::atomic::Ordering;
use std::sync::Arc;
use std::time::Duration;

use serde_json::{json, Value};
use tauri::{AppHandle, Emitter, Runtime};
use tokio::sync::broadcast::error::RecvError;

use crate::commands::task::task_detail_ipc;
use crate::commands::to_camel;
use crate::doubao::ConnEvent;
use crate::logging::now_ms;
use crate::state::{AppState, HEARTBEAT_INTERVAL};
use crate::tasks::TaskEvent;

/// GC 周期(spec 07:每 1 小时)。
pub const GC_INTERVAL: Duration = Duration::from_secs(3600);
/// 启动恢复时写进非终态任务的失败信息(11-error-codes `relay_restart`)。
pub const RELAY_RESTART_MESSAGE: &str = "中转程序重启,任务已作废,请重新提交";
/// 豆包崩溃时写进非终态任务的失败信息(`doubao_crashed`)。
/// 与 [`RelayError::doubao_crashed`](crate::error::RelayError::doubao_crashed) 共用同一文案,
/// 崩溃监听与 runner 的 `crash_aware_error` 两条路径给出的 `task.error.message` 一致。
pub const DOUBAO_CRASHED_MESSAGE: &str = crate::error::DOUBAO_CRASHED_MESSAGE;

/// 事件名(Rust → WebView,`docs/spec/05-tauri-ipc.md`)。
pub mod events {
    pub const DOUBAO_STATUS_CHANGED: &str = "doubao_status_changed";
    pub const ACCOUNT_LIST_CHANGED: &str = "account_list_changed";
    pub const ACCOUNT_SWITCHED: &str = "account_switched";
    /// spec 05 没有列;豆包崩溃(断线 ≥20s 且进程不在)时额外发一条,payload `{at}`。
    pub const DOUBAO_CRASHED: &str = "doubao_crashed";
    pub const LOG_ENTRY: &str = "log_entry";
    pub const CDP_LOG_ENTRY: &str = "cdp_log_entry";
    pub const TASK_CREATED: &str = "task_created";
    pub const TASK_UPDATED: &str = "task_updated";
    pub const TASK_COMPLETED: &str = "task_completed";
}

/// `ConnEvent` → `(事件名, camelCase payload)`。
pub fn conn_event_payload(event: ConnEvent) -> (&'static str, Value) {
    match event {
        ConnEvent::StatusChanged(status) => (
            events::DOUBAO_STATUS_CHANGED,
            to_camel(serde_json::to_value(status).unwrap_or(Value::Null)),
        ),
        ConnEvent::AccountListChanged(list) => (
            events::ACCOUNT_LIST_CHANGED,
            to_camel(serde_json::to_value(list).unwrap_or(Value::Null)),
        ),
        ConnEvent::AccountSwitched {
            previous,
            current,
            switched_at,
        } => (
            events::ACCOUNT_SWITCHED,
            json!({ "previous": previous, "current": current, "switchedAt": switched_at }),
        ),
        ConnEvent::Crashed => (events::DOUBAO_CRASHED, json!({ "at": logging::now_ms() })),
    }
}

/// `TaskEvent` → `(事件名, camelCase payload)`(spec 05:`task_created` = TaskSummary,
/// `task_updated` = `{taskId,status,stage,message,progress,updatedAt}`,`task_completed` = 完整任务,
/// 与 `task_get` 同形,含 `result.files[].localPath`)。
pub fn task_event_payload(
    event: TaskEvent,
    results_root: &std::path::Path,
) -> (&'static str, Value) {
    match event {
        TaskEvent::Created(summary) => (
            events::TASK_CREATED,
            to_camel(serde_json::to_value(summary).unwrap_or(Value::Null)),
        ),
        TaskEvent::Updated {
            task_id,
            status,
            stage,
            message,
            progress,
            updated_at,
        } => (
            events::TASK_UPDATED,
            json!({
                "taskId": task_id,
                "status": status,
                "stage": stage,
                "message": message,
                "progress": progress,
                "updatedAt": updated_at,
            }),
        ),
        TaskEvent::Completed(task) => {
            (events::TASK_COMPLETED, task_detail_ipc(&task, results_root))
        }
    }
}

/// 把 `ConnEvent` / `TaskEvent` / `LogEntry` / `CdpLogEntry` 四路广播转发为 Tauri 事件。需在 tokio 运行时内调用。
pub fn spawn_event_forwarders<R: Runtime>(app: AppHandle<R>, state: &Arc<AppState>) {
    // 任务事件
    {
        let app = app.clone();
        let root = state.storage.root.clone();
        let mut rx = state.scheduler.subscribe();
        tokio::spawn(async move {
            loop {
                match rx.recv().await {
                    Ok(ev) => {
                        let (name, payload) = task_event_payload(ev, &root);
                        if let Err(e) = app.emit(name, payload) {
                            tracing::debug!(event = name, error = %e, "事件发送失败");
                        }
                    }
                    Err(RecvError::Lagged(n)) => {
                        tracing::warn!(dropped = n, "任务事件转发落后,丢弃若干条");
                    }
                    Err(RecvError::Closed) => break,
                }
            }
        });
    }
    // 豆包连接事件
    {
        let app = app.clone();
        let mut rx = state.doubao.subscribe();
        tokio::spawn(async move {
            loop {
                match rx.recv().await {
                    Ok(ev) => {
                        let (name, payload) = conn_event_payload(ev);
                        if let Err(e) = app.emit(name, payload) {
                            tracing::debug!(event = name, error = %e, "事件发送失败");
                        }
                    }
                    Err(RecvError::Lagged(n)) => {
                        tracing::warn!(dropped = n, "连接事件转发落后,丢弃若干条");
                    }
                    Err(RecvError::Closed) => break,
                }
            }
        });
    }
    // 日志
    {
        let app = app.clone();
        let mut rx = state.logs.subscribe();
        tokio::spawn(async move {
            loop {
                match rx.recv().await {
                    Ok(entry) => {
                        let payload = to_camel(serde_json::to_value(entry).unwrap_or(Value::Null));
                        let _ = app.emit(events::LOG_ENTRY, payload);
                    }
                    Err(RecvError::Lagged(_)) => continue,
                    Err(RecvError::Closed) => break,
                }
            }
        });
    }
    // CDP 调试日志:常驻订阅 UI 通道(关闭时收集任务不转发,通道空转;开关可在设置页热切换)
    {
        let mut rx = state.cdp_log.subscribe();
        tokio::spawn(async move {
            loop {
                match rx.recv().await {
                    Ok(entry) => {
                        let payload = to_camel(serde_json::to_value(entry).unwrap_or(Value::Null));
                        let _ = app.emit(events::CDP_LOG_ENTRY, payload);
                    }
                    Err(RecvError::Lagged(_)) => continue,
                    Err(RecvError::Closed) => break,
                }
            }
        });
    }
}

/// 启动恢复:上次运行中断的任务标记 `failed/relay_restart`(spec 08 "恢复策略"),
/// pending 任务保留并重新装入调度队列,
/// 带 `callback_url` 的任务补发 `task.failed` 回调(spec 04 "任务完成/失败/取消时 MUST POST";
/// 后台投递,不阻塞启动)。上次进程退出时仍在重试中的回调不会恢复。
pub async fn recover_after_restart(state: &AppState) {
    let ids = match state
        .storage
        .fail_running_non_terminal("relay_restart", RELAY_RESTART_MESSAGE, now_ms())
        .await
    {
        Ok(ids) => ids,
        Err(e) => {
            tracing::error!(code = e.code, error = %e, "启动恢复失败");
            return;
        }
    };
    if ids.is_empty() {
        return;
    }
    tracing::info!(count = ids.len(), "启动恢复:上次运行中的任务已标记失败");
    for id in &ids {
        match state.storage.get_task(id).await {
            Ok(Some(task)) => {
                state.scheduler.notify_terminal(&task);
            }
            Ok(None) => {}
            Err(e) => {
                tracing::warn!(task_id = %id, code = e.code, error = %e, "启动恢复:读取任务失败,未回调")
            }
        }
    }
}

/// 订阅连接层事件:`ConnEvent::Crashed` → `fail_all_running_with("doubao_crashed", …, {last_seen_at})`
/// (spec 08:所有非终态任务置 failed;新请求的 503 由 `Scheduler::submit` 的就绪检查负责)。
pub fn spawn_crash_watcher(
    scheduler: Arc<crate::tasks::Scheduler>,
    mut rx: tokio::sync::broadcast::Receiver<ConnEvent>,
    last_seen_at: impl Fn() -> i64 + Send + 'static,
) -> tokio::task::JoinHandle<()> {
    tokio::spawn(async move {
        loop {
            match rx.recv().await {
                Ok(ConnEvent::Crashed) => {
                    tracing::warn!("豆包崩溃,中止所有进行中的任务");
                    let e = crate::error::RelayError::doubao_crashed(last_seen_at());
                    scheduler
                        .fail_all_running_with(e.code, &e.message, e.details)
                        .await;
                }
                Ok(_) | Err(RecvError::Lagged(_)) => continue,
                Err(RecvError::Closed) => break,
            }
        }
    })
}

/// Phase 2 的后台任务:调度循环、豆包崩溃 → 全部运行中任务失败、每小时 GC。需在 tokio 运行时内调用。
pub fn spawn_task_services(state: &Arc<AppState>) {
    state.scheduler.spawn_loop();
    {
        let doubao = Arc::clone(&state.doubao);
        spawn_crash_watcher(
            Arc::clone(&state.scheduler),
            state.doubao.subscribe(),
            move || doubao.last_seen_at(),
        );
    }
    {
        let state = Arc::clone(state);
        tokio::spawn(async move {
            let mut tick = tokio::time::interval(GC_INTERVAL);
            tick.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
            loop {
                tick.tick().await;
                let cfg = state.config().await.storage;
                match state.storage.gc(now_ms(), &cfg).await {
                    Ok(report) => tracing::info!(
                        result_files_removed = report.result_files_removed,
                        tasks_removed = report.tasks_removed,
                        uploads_removed = report.uploads_removed,
                        "定时清理完成"
                    ),
                    Err(e) => tracing::warn!(code = e.code, error = %e, "定时清理失败"),
                }
            }
        });
    }
}

/// 以独立任务启动 HTTP API server;绑定失败只记错误日志(用户可在设置页改端口后重启)。
pub fn spawn_http_server(state: &Arc<AppState>) {
    let state = Arc::clone(state);
    tokio::spawn(async move {
        let (enabled, host, port) = {
            let cfg = state.config().await;
            (cfg.server.enabled, cfg.server.host, cfg.server.port)
        };
        state.api_listening.store(false, Ordering::Relaxed);
        *state.api_error.write().await = None;
        if !enabled {
            tracing::info!("HTTP API 已关闭");
            return;
        }
        let shutdown = state.shutdown.clone();
        if let Err(e) = http::serve(Arc::clone(&state), &host, port, shutdown).await {
            let message = format!("{e:#}");
            *state.api_error.write().await = Some(message.clone());
            tracing::error!(host, port, error = message, "HTTP API 启动失败");
        }
    });
}

/// Task 5.2:载入隐私声明同意状态;已同意且 `ui.start_minimized` 时把主窗口收到托盘;
/// 未同意时保证主窗口可见(前端显示全屏声明)并等到用户同意。
pub async fn wait_for_privacy<R: Runtime>(app: &AppHandle<R>, state: &Arc<AppState>) {
    let accepted = match commands::privacy::load(state).await {
        Ok(at) => at.is_some(),
        Err(e) => {
            tracing::error!(code = e.code, error = %e, "读取隐私声明同意状态失败,按未同意处理");
            false
        }
    };
    if accepted {
        if state.config().await.ui.start_minimized {
            if let Some(w) = tauri::Manager::get_webview_window(app, tray::MAIN_WINDOW) {
                let _ = w.hide();
            }
        }
        return;
    }
    tray::show_main_window(app);
    tracing::info!("等待用户同意隐私声明,同意前不启动 HTTP API、不自动连接豆包");
    commands::privacy::wait_accepted(state).await;
}

/// `setup` 里的后台任务:事件转发、CDP 日志收集、心跳、按配置自动连豆包、
/// 启动恢复、调度循环、崩溃处理、GC、HTTP API。
pub fn spawn_background<R: Runtime>(app: AppHandle<R>, state: Arc<AppState>) {
    tauri::async_runtime::spawn(async move {
        let app_for_privacy = app.clone();
        spawn_event_forwarders(app, &state);
        state
            .cdp_log
            .ensure_collector(Some(&state.platform.log_dir()));
        state.doubao.spawn_heartbeat(HEARTBEAT_INTERVAL);
        recover_after_restart(&state).await;
        match state.scheduler.restore_pending().await {
            Ok(count) if count > 0 => tracing::info!(count, "启动恢复:排队任务已重新装入调度队列"),
            Ok(_) => {}
            Err(e) => tracing::error!(code = e.code, error = %e, "启动恢复:恢复排队任务失败"),
        }
        spawn_task_services(&state);
        // Task 5.2:隐私声明同意前不启动 HTTP API、不自动连豆包(spec 12"MUST 明确同意一次才可用")
        wait_for_privacy(&app_for_privacy, &state).await;
        spawn_http_server(&state);
        v88_executor::spawn_v88_service(Arc::clone(&state));
        if state.config().await.doubao.auto_connect_on_startup {
            let doubao = Arc::clone(&state.doubao);
            tokio::spawn(async move {
                match doubao.connect(false).await {
                    Ok(action) => tracing::info!(action = ?action, "启动时自动连接豆包完成"),
                    Err(e) => tracing::warn!(code = e.code, error = %e, "启动时自动连接豆包失败"),
                }
            });
        }
    });
}

/// 注册全部命令的 Builder(测试用 `mock_builder()`,生产用 `Builder::default()`)。
pub fn configure<R: Runtime>(
    builder: tauri::Builder<R>,
    state: Arc<AppState>,
) -> tauri::Builder<R> {
    builder
        .manage(state)
        .invoke_handler(tauri::generate_handler![
            commands::app::app_get_version,
            commands::app::app_api_status,
            commands::app::app_find_available_port,
            commands::app::app_get_config,
            commands::app::app_save_config,
            commands::app::app_pair_v88,
            commands::account::account_list,
            commands::account::account_refresh,
            commands::account::account_switch,
            commands::doubao::doubao_status,
            commands::doubao::doubao_reconnect,
            commands::log::log_recent,
            commands::log::debug_cdp_log,
            commands::task::task_list,
            commands::task::task_get,
            commands::task::task_submit,
            commands::task::task_cancel,
            commands::task::task_confirm,
            commands::task::task_open_result_folder,
            commands::upload::upload_add,
            commands::app::app_regenerate_token,   // Task 5.1
            commands::app::maintenance_gc,         // Task 5.1
            commands::log::log_export,             // Task 5.1
            commands::privacy::app_privacy_status, // Task 5.2
            commands::privacy::app_privacy_accept, // Task 5.2
            commands::privacy::app_quit,           // Task 5.2
        ])
        .on_window_event(tray::handle_window_event) // Task 5.2:关闭按钮 → 隐藏到托盘
}

pub fn run() {
    let platform = platform::create_platform();
    let state = match AppState::bootstrap(platform) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("doubao-relay 启动失败:{e:#}");
            std::process::exit(1);
        }
    };
    let setup_state = Arc::clone(&state);
    let exit_state = Arc::clone(&state);
    // 插件只在生产 Builder 上注册(mock runtime 的测试不碰系统剪贴板 / 对话框)
    let builder = tauri::Builder::default()
        .plugin(tauri_plugin_clipboard_manager::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_single_instance::init(|app, _argv, _cwd| {
            // 第二次启动只唤回已有窗口，不创建第二套 HTTP/心跳服务。
            if let Some(window) = tauri::Manager::get_webview_window(app, tray::MAIN_WINDOW) {
                let _ = window.unminimize();
                let _ = window.show();
                let _ = window.set_focus();
            }
        }));
    // Task 5.2:macOS 应用菜单的 Cmd+Q 换成走关闭流程的自定义项
    #[cfg(target_os = "macos")]
    let builder = builder.menu(tray::app_menu);
    configure(builder, state)
        .on_menu_event(tray::handle_menu_event)
        .setup(move |app| {
            if let Err(e) = tray::setup_tray(app.handle(), &setup_state) {
                tracing::error!(error = %e, "创建系统托盘失败,关闭窗口将直接退出");
            }
            spawn_background(app.handle().clone(), setup_state);
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("doubao-relay 启动失败")
        .run(move |app, event| match event {
            // 关闭流程(spec 03):先拦下退出,跑完关闭流程后由 request_quit 再 exit
            tauri::RunEvent::ExitRequested { api, .. } if !exit_state.quit.is_done() => {
                api.prevent_exit();
                shutdown::request_quit(app);
            }
            tauri::RunEvent::Exit => shutdown::finish_on_exit(&exit_state),
            #[cfg(target_os = "macos")]
            tauri::RunEvent::Reopen { .. } => tray::show_main_window(app),
            _ => {}
        });
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::mpsc;
    use std::time::Duration;

    use tauri::ipc::{CallbackFn, InvokeBody};
    use tauri::test::{mock_builder, mock_context, noop_assets, INVOKE_KEY};
    use tauri::webview::InvokeRequest;
    use tauri::Listener;

    use crate::commands::testing::test_state_blocking;
    use crate::doubao::DoubaoStatus;
    use crate::logging::{now_ms, LogEntry};

    fn mock_app(state: Arc<AppState>) -> tauri::App<tauri::test::MockRuntime> {
        let mut context = mock_context(noop_assets());
        for command in [
            "app_get_version",
            "doubao_status",
            "app_save_config",
            "app_get_config",
            "log_recent",
            "debug_cdp_log",
            "app_regenerate_token",
            "maintenance_gc",
            "log_export",
            "account_switch",
            "doubao_reconnect",
            "app_privacy_status",
            "app_privacy_accept",
        ] {
            context.runtime_authority_mut().__allow_command(
                command.into(),
                tauri::utils::acl::ExecutionContext::Remote {
                    url: "tauri://localhost".parse().unwrap(),
                },
            );
        }
        configure(mock_builder(), state)
            .build(context)
            .expect("mock app")
    }

    fn invoke(
        webview: &tauri::WebviewWindow<tauri::test::MockRuntime>,
        cmd: &str,
        body: Value,
    ) -> Result<Value, Value> {
        tauri::test::get_ipc_response(
            webview,
            InvokeRequest {
                cmd: cmd.into(),
                callback: CallbackFn(0),
                error: CallbackFn(1),
                url: "tauri://localhost".parse().unwrap(),
                body: InvokeBody::Json(body),
                headers: Default::default(),
                invoke_key: INVOKE_KEY.to_string(),
            },
        )
        .map(|b| b.deserialize::<Value>().unwrap())
    }

    #[test]
    fn commands_are_wired_and_return_camel_case() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        let app = mock_app(state);
        let webview = tauri::WebviewWindowBuilder::new(&app, "main", Default::default())
            .build()
            .unwrap();

        let v = invoke(&webview, "app_get_version", json!({})).unwrap();
        assert_eq!(v["tauriVersion"], tauri::VERSION);

        let v = invoke(&webview, "doubao_status", json!({})).unwrap();
        assert_eq!(v["cdpConnected"], false);

        let v = invoke(
            &webview,
            "app_save_config",
            json!({ "patch": { "logging": { "level": "debug" } } }),
        )
        .unwrap();
        assert_eq!(
            v,
            json!({ "ok": true, "needsRestart": false }),
            "日志级别热生效,不需要重启"
        );
        let v = invoke(&webview, "app_get_config", json!({})).unwrap();
        assert_eq!(v["logging"]["level"], "debug");
        let v = invoke(
            &webview,
            "app_save_config",
            json!({ "patch": { "server": { "port": 8790 } } }),
        )
        .unwrap();
        assert_eq!(v["needsRestart"], true, "端口变更需要重启");

        let v = invoke(
            &webview,
            "log_recent",
            json!({ "limit": 5, "level": "info" }),
        )
        .unwrap();
        assert!(v["entries"].is_array());
        let v = invoke(&webview, "debug_cdp_log", json!({})).unwrap();
        assert_eq!(v["enabled"], true);
        let v = invoke(&webview, "app_regenerate_token", json!({})).unwrap();
        assert_eq!(v["bearerToken"].as_str().unwrap().len(), 64);
        let v = invoke(&webview, "maintenance_gc", json!({})).unwrap();
        assert_eq!(v["tasksRemoved"], 0);
        let export_path = dir.path().join("logs.ndjson");
        let v = invoke(
            &webview,
            "log_export",
            json!({ "path": export_path.to_str().unwrap(), "level": "info" }),
        )
        .unwrap();
        assert_eq!(v["ok"], true);
        assert!(export_path.exists());

        // 错误按 {code, message, details} 结构返回
        let e = invoke(&webview, "account_switch", json!({ "target": "张三" })).unwrap_err();
        assert!(e["code"].as_str().unwrap().starts_with("doubao_"), "{e}");
        assert!(e["message"].is_string());
        let e = invoke(&webview, "doubao_reconnect", json!({ "force": false })).unwrap_err();
        assert!(e["code"].as_str().unwrap().starts_with("doubao_"), "{e}");
        let e = invoke(&webview, "doubao_reconnect", json!({})).unwrap_err();
        assert!(e["code"].as_str().unwrap().starts_with("doubao_"), "{e}");
    }

    /// Task 5.2:隐私声明命令已注册,首次为未同意,同意后持久化。
    #[test]
    fn privacy_commands_are_wired() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        let app = mock_app(Arc::clone(&state));
        let webview = tauri::WebviewWindowBuilder::new(&app, "main", Default::default())
            .build()
            .unwrap();
        let v = invoke(&webview, "app_privacy_status", json!({})).unwrap();
        assert_eq!(v, json!({ "accepted": false, "acceptedAt": null }));
        let v = invoke(&webview, "app_privacy_accept", json!({})).unwrap();
        assert_eq!(v["accepted"], true);
        let v = invoke(&webview, "app_privacy_status", json!({})).unwrap();
        assert_eq!(v["accepted"], true);
        assert!(v["acceptedAt"].as_i64().unwrap() > 0);
        assert!(*state.privacy.borrow());
    }

    #[tokio::test]
    async fn recover_after_restart_fails_only_non_terminal_tasks() {
        use crate::tasks::{ConfirmationMode, GenerateRequest, Task, TaskStatus, TaskType};

        let dir = tempfile::tempdir().unwrap();
        let state = crate::commands::testing::test_state(dir.path()).await;
        let make = |id: &str, status: TaskStatus| Task {
            id: id.into(),
            request: GenerateRequest {
                account: "张三".into(),
                task_type: TaskType::Image,
                model_name: "Seedream 5.0 Pro".into(),
                prompt: "p".into(),
                ratio: "1:1".into(),
                duration: None,
                reference_files: vec![],
                confirmation_mode: ConfirmationMode::Auto,
                auto_use_membership: true,
                callback_url: None,
            },
            status,
            stage: None,
            stage_message: None,
            progress: None,
            error: None,
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: None,
            result_files: None,
            created_at: 1,
            started_at: None,
            completed_at: None,
            expires_at: now_ms() + 60_000,
            stage_history: vec![],
            queue_position: None,
        };
        state
            .storage
            .insert_task(&make("01AAAAAAAAAAAAAAAAAAAAAAAA", TaskStatus::Generating))
            .await
            .unwrap();
        state
            .storage
            .insert_task(&make("01BBBBBBBBBBBBBBBBBBBBBBBB", TaskStatus::Completed))
            .await
            .unwrap();

        // 带 callback_url 的非终态任务:恢复后补发 task.failed;已终态的任务不回调
        type Hits = Arc<std::sync::Mutex<Vec<(axum::http::HeaderMap, Vec<u8>)>>>;
        let hits: Hits = Arc::default();
        let app = axum::Router::new()
            .route(
                "/hook",
                axum::routing::post(
                    |axum::extract::State(h): axum::extract::State<Hits>,
                     headers: axum::http::HeaderMap,
                     body: axum::body::Bytes| async move {
                        h.lock().unwrap().push((headers, body.to_vec()));
                        "ok"
                    },
                ),
            )
            .with_state(Arc::clone(&hits));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let hook_url = format!("http://{}/hook", listener.local_addr().unwrap());
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        let mut c = make("01CCCCCCCCCCCCCCCCCCCCCCCC", TaskStatus::WaitingUserConfirm);
        c.request.callback_url = Some(hook_url.clone());
        state.storage.insert_task(&c).await.unwrap();
        let mut d = make("01DDDDDDDDDDDDDDDDDDDDDDDD", TaskStatus::Completed);
        d.request.callback_url = Some(hook_url);
        state.storage.insert_task(&d).await.unwrap();

        recover_after_restart(&state).await;

        let deadline = std::time::Instant::now() + std::time::Duration::from_secs(5);
        while hits.lock().unwrap().is_empty() {
            assert!(std::time::Instant::now() < deadline, "启动恢复未补发回调");
            tokio::time::sleep(std::time::Duration::from_millis(10)).await;
        }
        tokio::time::sleep(std::time::Duration::from_millis(100)).await;
        let got = hits.lock().unwrap().clone();
        assert_eq!(got.len(), 1, "只回调被恢复置为失败的任务");
        let (headers, body) = &got[0];
        let token = state.config().await.server.bearer_token;
        let ts: i64 = headers["x-doubao-relay-timestamp"]
            .to_str()
            .unwrap()
            .parse()
            .unwrap();
        assert_eq!(
            headers["x-doubao-relay-signature"].to_str().unwrap(),
            crate::tasks::webhook::signature_header(&token, ts, body)
        );
        let v: Value = serde_json::from_slice(body).unwrap();
        assert_eq!(v["task_id"], "01CCCCCCCCCCCCCCCCCCCCCCCC");
        assert_eq!(v["event"], "task.failed");
        assert_eq!(v["status"], "failed");
        assert_eq!(v["task"]["error"]["code"], "relay_restart");
        assert!(v["task"]["failed_at"].is_i64());
        let c = state
            .storage
            .get_task("01CCCCCCCCCCCCCCCCCCCCCCCC")
            .await
            .unwrap()
            .unwrap();
        assert_eq!(c.status, TaskStatus::Failed);
        let a = state
            .storage
            .get_task("01AAAAAAAAAAAAAAAAAAAAAAAA")
            .await
            .unwrap()
            .unwrap();
        assert_eq!(a.status, TaskStatus::Failed);
        assert_eq!(a.error.as_ref().unwrap().code, "relay_restart");
        assert_eq!(a.error.unwrap().message, RELAY_RESTART_MESSAGE);
        let b = state
            .storage
            .get_task("01BBBBBBBBBBBBBBBBBBBBBBBB")
            .await
            .unwrap()
            .unwrap();
        assert_eq!(b.status, TaskStatus::Completed);
        // 再跑一次没有可恢复的任务也不报错
        recover_after_restart(&state).await;
    }

    #[test]
    fn conn_events_map_to_spec_names_and_camel_payloads() {
        let status = DoubaoStatus {
            running: true,
            cdp_connected: true,
            cdp_port: 9707,
            doubao_version_detected: Some("2.32.5".into()),
            ui_compat_check: "ok".into(),
            ui_compat_details: Default::default(),
            chat_targets: 1,
            reconnecting: false,
            last_check_at: 1,
        };
        let (name, p) = conn_event_payload(ConnEvent::StatusChanged(status));
        assert_eq!(name, "doubao_status_changed");
        assert_eq!(p["cdpConnected"], true);
        assert_eq!(p["doubaoVersionDetected"], "2.32.5");
        assert!(p["uiCompatDetails"].get("accountButtonFound").is_some());

        let (name, p) = conn_event_payload(ConnEvent::AccountSwitched {
            previous: None,
            current: "李四".into(),
            switched_at: 7,
        });
        assert_eq!(name, "account_switched");
        assert_eq!(
            p,
            json!({ "previous": null, "current": "李四", "switchedAt": 7 })
        );

        let list = crate::doubao::AccountList {
            current: Some("a".into()),
            accounts: vec![],
            read_at: 3,
        };
        let (name, p) = conn_event_payload(ConnEvent::AccountListChanged(list));
        assert_eq!(name, "account_list_changed");
        assert_eq!(p["readAt"], 3);

        let (name, p) = conn_event_payload(ConnEvent::Crashed);
        assert_eq!(name, "doubao_crashed");
        assert!(p["at"].as_i64().unwrap() > 0);
    }

    #[test]
    fn task_events_map_to_spec_names_and_camel_payloads() {
        use crate::tasks::{
            ConfirmationMode, GenerateRequest, Task, TaskStage, TaskStatus, TaskSummary, TaskType,
        };
        let root = std::path::Path::new("/data");
        let summary = TaskSummary {
            id: "01A".into(),
            account: "张三".into(),
            task_type: TaskType::Image,
            status: TaskStatus::Pending,
            stage: None,
            created_at: 1,
            completed_at: None,
        };
        let (name, p) = task_event_payload(TaskEvent::Created(summary), root);
        assert_eq!(name, "task_created");
        assert_eq!(p["taskId"], "01A");
        assert_eq!(p["type"], "image");
        assert_eq!(p["createdAt"], 1);
        assert!(p.get("task_id").is_none());

        let (name, p) = task_event_payload(
            TaskEvent::Updated {
                task_id: "01A".into(),
                status: TaskStatus::Generating,
                stage: Some(TaskStage::QuotaChecking),
                message: Some("豆包确认是否使用会员配额".into()),
                progress: None,
                updated_at: 7,
            },
            root,
        );
        assert_eq!(name, "task_updated");
        assert_eq!(
            p,
            json!({
                "taskId": "01A", "status": "generating", "stage": "quota_checking",
                "message": "豆包确认是否使用会员配额", "progress": null, "updatedAt": 7
            })
        );

        let task = Task {
            id: "01A".into(),
            request: GenerateRequest {
                account: "张三".into(),
                task_type: TaskType::Image,
                model_name: "Seedream 5.0 Pro".into(),
                prompt: "p".into(),
                ratio: "1:1".into(),
                duration: None,
                reference_files: vec![],
                confirmation_mode: ConfirmationMode::Auto,
                auto_use_membership: true,
                callback_url: None,
            },
            status: TaskStatus::Failed,
            stage: Some(TaskStage::Refused),
            stage_message: None,
            progress: None,
            error: Some(crate::tasks::TaskError {
                code: "generation_doubao_refused".into(),
                message: "豆包拒绝生成".into(),
                details: None,
            }),
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: None,
            result_files: None,
            created_at: 1,
            started_at: Some(2),
            completed_at: Some(3),
            expires_at: 4,
            stage_history: vec![],
            queue_position: None,
        };
        let (name, p) = task_event_payload(TaskEvent::Completed(task), root);
        assert_eq!(name, "task_completed");
        assert_eq!(p["taskId"], "01A");
        assert_eq!(p["status"], "failed");
        assert_eq!(p["failedAt"], 3);
        assert_eq!(p["error"]["code"], "generation_doubao_refused");
        assert_eq!(p["request"]["modelName"], "Seedream 5.0 Pro");
    }

    /// UC-32(mock):`ConnEvent::Crashed` → 运行中与排队中的任务都 `failed/doubao_crashed{last_seen_at}`。
    #[tokio::test]
    async fn crashed_event_fails_running_and_pending_tasks() {
        use crate::commands::testing::{test_state_with, FakeBackend};
        use crate::tasks::{ConfirmationMode, GenerateRequest, TaskStatus, TaskType};

        let dir = tempfile::tempdir().unwrap();
        let state = test_state_with(dir.path(), Some(FakeBackend::new(&["张三"]))).await;
        state.scheduler.spawn_loop();
        let (tx, rx) = tokio::sync::broadcast::channel(8);
        spawn_crash_watcher(Arc::clone(&state.scheduler), rx, || 4242);
        let req = || GenerateRequest {
            account: "张三".into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "p".into(),
            ratio: "1:1".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        };
        let (running, _, _) = state.scheduler.submit(req()).await.unwrap();
        let (queued, _, _) = state.scheduler.submit(req()).await.unwrap();
        let deadline = std::time::Instant::now() + Duration::from_secs(5);
        while state.scheduler.running_task_ids(None) != vec![running.clone()] {
            assert!(std::time::Instant::now() < deadline, "第一个任务未开始运行");
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        tx.send(ConnEvent::Crashed).unwrap();
        for id in [&running, &queued] {
            let deadline = std::time::Instant::now() + Duration::from_secs(5);
            let t = loop {
                let t = state.storage.get_task(id).await.unwrap().unwrap();
                if t.status.is_terminal() {
                    break t;
                }
                assert!(std::time::Instant::now() < deadline, "{id} 未进入终态");
                tokio::time::sleep(Duration::from_millis(10)).await;
            };
            assert_eq!(t.status, TaskStatus::Failed);
            let e = t.error.unwrap();
            assert_eq!(e.code, "doubao_crashed");
            assert_eq!(e.message, DOUBAO_CRASHED_MESSAGE);
            assert_eq!(e.details.unwrap()["last_seen_at"], 4242);
        }
    }

    /// 调度器广播的任务事件经 `spawn_event_forwarders` 到达 WebView 监听器。
    #[test]
    fn scheduler_events_are_forwarded_to_webview() {
        use crate::commands::testing::{test_state_with, FakeBackend};
        use crate::tasks::{ConfirmationMode, GenerateRequest, TaskType};

        let dir = tempfile::tempdir().unwrap();
        let state = tauri::async_runtime::block_on(test_state_with(
            dir.path(),
            Some(FakeBackend::new(&["张三"])),
        ));
        let app = mock_app(Arc::clone(&state));
        let handle = app.handle().clone();
        let (tx, rx) = mpsc::channel::<(String, Value)>();
        {
            let tx = tx.clone();
            handle.listen_any(events::TASK_CREATED, move |e| {
                let _ = tx.send(("created".into(), serde_json::from_str(e.payload()).unwrap()));
            });
        }
        handle.listen_any(events::TASK_COMPLETED, move |e| {
            let _ = tx.send((
                "completed".into(),
                serde_json::from_str(e.payload()).unwrap(),
            ));
        });
        let forward_state = Arc::clone(&state);
        tauri::async_runtime::block_on(async move {
            spawn_event_forwarders(handle, &forward_state);
        });
        let id = tauri::async_runtime::block_on(async {
            let (id, _, _) = state
                .scheduler
                .submit(GenerateRequest {
                    account: "张三".into(),
                    task_type: TaskType::Image,
                    model_name: "Seedream 5.0 Pro".into(),
                    prompt: "p".into(),
                    ratio: "1:1".into(),
                    duration: None,
                    reference_files: vec![],
                    confirmation_mode: ConfirmationMode::Auto,
                    auto_use_membership: true,
                    callback_url: None,
                })
                .await
                .unwrap();
            // 调度循环没起,任务停在 pending,取消它产生 Completed(cancelled)
            state.scheduler.cancel(&id).await.unwrap();
            id
        });
        let mut got = Vec::new();
        while got.len() < 2 {
            got.push(
                rx.recv_timeout(Duration::from_secs(5))
                    .expect("任务事件应在 5s 内转发到 WebView 监听器"),
            );
        }
        let created = &got.iter().find(|(k, _)| k == "created").unwrap().1;
        assert_eq!(created["taskId"], id);
        assert_eq!(created["status"], "pending");
        let completed = &got.iter().find(|(k, _)| k == "completed").unwrap().1;
        assert_eq!(completed["taskId"], id);
        assert_eq!(completed["status"], "cancelled");
        assert_eq!(completed["error"]["code"], "generation_cancelled");
    }

    #[test]
    fn log_and_cdp_entries_are_forwarded_as_events() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        let app = mock_app(Arc::clone(&state));
        let handle = app.handle().clone();

        let (tx, rx) = mpsc::channel::<(String, Value)>();
        {
            let tx = tx.clone();
            handle.listen_any(events::LOG_ENTRY, move |e| {
                let _ = tx.send(("log".into(), serde_json::from_str(e.payload()).unwrap()));
            });
        }
        handle.listen_any(events::CDP_LOG_ENTRY, move |e| {
            let _ = tx.send(("cdp".into(), serde_json::from_str(e.payload()).unwrap()));
        });

        let forward_state = Arc::clone(&state);
        tauri::async_runtime::block_on(async move {
            spawn_event_forwarders(handle, &forward_state);
            forward_state.cdp_log.ensure_collector(None);
        });

        state.logs.push(LogEntry {
            ts: now_ms(),
            level: "info".into(),
            target: "t".into(),
            message: "你好".into(),
            fields: json!({ "task_id": "x" }),
        });
        state
            .cdp_log
            .sender()
            .unwrap()
            .send(crate::cdp::CdpLogEntry {
                ts: 2,
                direction: "recv".into(),
                method: None,
                params: None,
                result: Some(json!({ "target_infos": [] })),
                event: None,
            })
            .unwrap();

        let mut got = Vec::new();
        while got.len() < 2 {
            let item = rx
                .recv_timeout(Duration::from_secs(5))
                .expect("事件应在 5s 内转发到 WebView 监听器");
            got.push(item);
        }
        let log = got
            .iter()
            .find(|(k, _)| k == "log")
            .map(|(_, v)| v)
            .unwrap();
        assert_eq!(log["message"], "你好");
        assert_eq!(log["fields"]["taskId"], "x");
        let cdp = got
            .iter()
            .find(|(k, _)| k == "cdp")
            .map(|(_, v)| v)
            .unwrap();
        assert_eq!(cdp["direction"], "recv");
        assert_eq!(cdp["result"]["targetInfos"], json!([]));
    }
}
