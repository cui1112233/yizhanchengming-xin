//! `app.*` 命令:版本信息、读/写配置(logging / cdp_debug_log 热生效)、重新生成 token;
//! `maintenance.gc`(设置页"立即清理")。

use std::sync::Arc;

use serde_json::{json, Value};
use tauri::State;

use crate::config::Config;
use crate::error::{RelayError, RelayResult};
use crate::logging;
use crate::state::AppState;
use crate::v88_executor::{PairRequest, V88ExecutorClient};

use super::{respond, to_snake};

/// `rustc -V` 的版本号,由 `build.rs` 写入(没拿到时为空串)。
pub const RUSTC_VERSION: &str = env!("DOUBAO_RELAY_RUSTC_VERSION");

/// `app.get_version` → `{version, buildTarget, tauriVersion, rustcVersion}`。
pub fn get_version() -> Value {
    json!({
        "version": env!("CARGO_PKG_VERSION"),
        "build_target": std::env::consts::OS,
        "tauri_version": tauri::VERSION,
        "rustc_version": RUSTC_VERSION,
    })
}

/// `app.get_config` → 整个 `Config`(含 bearerToken,设置页要显示/复制)。
pub async fn get_config(state: &AppState) -> RelayResult<Value> {
    respond(&state.config().await)
}

/// `app.save_config`:把 camelCase 的 `patch` 深合并进当前配置 → 校验 → 写盘 → 更新内存 → 热生效。
/// 返回 `{ok, needsRestart}`;`needsRestart` 见 [`needs_restart`]。
///
/// 热生效:`logging.level` 经 [`logging::set_level`];`logging.cdp_debug_log` 经
/// `CdpLogBuffer::set_enabled`;`ui.*` / `storage.*` / `concurrency.*` 由使用方每次读 `state.config`。
pub async fn save_config(state: &AppState, patch: Value) -> RelayResult<Value> {
    let Value::Object(patch) = to_snake(patch) else {
        return Err(RelayError::new(
            "request_malformed",
            "配置补丁必须是对象,请按 {server:{...}} 的形式提交",
        ));
    };
    let mut guard = state.config.write().await;
    let current = guard.clone();
    let mut merged = serde_json::to_value(&current)
        .map_err(|e| RelayError::internal(format!("序列化当前配置失败:{e}")))?;
    deep_merge(&mut merged, Value::Object(patch));
    let next: Config = serde_json::from_value(merged).map_err(|e| {
        RelayError::new("request_malformed", "配置字段类型不正确,请检查提交的值")
            .with_details(json!({ "reason": e.to_string() }))
    })?;
    if let Err(reason) = next.validate() {
        return Err(
            RelayError::new("request_malformed", "配置取值不合法,请按提示修改后重试")
                .with_details(json!({ "reason": reason })),
        );
    }
    next.save(&state.config_path).map_err(|e| {
        RelayError::new("system_io_error", "写入配置文件失败,请检查目录权限")
            .with_details(json!({ "io": format!("{e:#}") }))
    })?;
    let restart = needs_restart(&current, &next);
    *guard = next.clone();
    drop(guard);
    apply_hot_reload(state, &current, &next);
    tracing::info!(needs_restart = restart, "配置已保存");
    Ok(json!({ "ok": true, "needsRestart": restart }))
}

/// 配置写入后立刻生效的部分:日志级别、CDP 调试开关。失败只 warn(配置已经落盘,重启后也会生效)。
fn apply_hot_reload(state: &AppState, before: &Config, after: &Config) {
    if before.logging.level != after.logging.level {
        match logging::set_level(&after.logging.level) {
            Ok(()) => tracing::info!(level = %after.logging.level, "日志级别已热更新"),
            Err(e) => tracing::warn!(error = %e, "日志级别热更新失败,重启后生效"),
        }
    }
    if before.logging.cdp_debug_log != after.logging.cdp_debug_log {
        state
            .cdp_log
            .set_enabled(after.logging.cdp_debug_log, Some(&state.platform.log_dir()));
    }
}

/// 需要重启才生效的项:HTTP 监听地址 / 端口、豆包 CDP 地址 / 端口。
/// token(HTTP 鉴权每次读 `state.config`)、日志级别、CDP 调试开关均热生效,不算。
pub fn needs_restart(before: &Config, after: &Config) -> bool {
    before.server.enabled != after.server.enabled
        || before.server.host != after.server.host
        || before.server.port != after.server.port
        || before.doubao.cdp_port != after.doubao.cdp_port
        || before.doubao.cdp_address != after.doubao.cdp_address
}

/// `app.regenerate_token`:生成新的 bearer token → 写盘 → 更新内存(HTTP 鉴权随即只认新 token)。
/// 返回 `{bearerToken}`。日志里不出现 token 值。
pub async fn regenerate_token(state: &AppState) -> RelayResult<Value> {
    let mut guard = state.config.write().await;
    let mut next = guard.clone();
    next.server.bearer_token = Config::generate_token();
    next.save(&state.config_path).map_err(|e| {
        RelayError::new("system_io_error", "写入配置文件失败,请检查目录权限")
            .with_details(json!({ "io": format!("{e:#}") }))
    })?;
    let token = next.server.bearer_token.clone();
    *guard = next;
    drop(guard);
    tracing::info!("Bearer token 已重新生成,旧 token 立即失效");
    Ok(json!({ "bearerToken": token }))
}

/// `maintenance.gc`:立刻跑一遍 07-data-model 的四步清理,返回 `GcReport`(camelCase)。
pub async fn run_gc(state: &AppState) -> RelayResult<Value> {
    let storage = &state.storage;
    let cfg = state.config().await;
    let report = storage.gc(logging::now_ms(), &cfg.storage).await?;
    tracing::info!(
        result_files_removed = report.result_files_removed,
        tasks_removed = report.tasks_removed,
        uploads_removed = report.uploads_removed,
        "手动清理完成"
    );
    respond(&report)
}

/// 对象递归合并,其余类型整体替换;`null` 也视为替换(调用方想清空就传 null)。
pub fn deep_merge(base: &mut Value, patch: Value) {
    match (base, patch) {
        (Value::Object(b), Value::Object(p)) => {
            for (k, v) in p {
                match b.get_mut(&k) {
                    Some(slot) if slot.is_object() && v.is_object() => deep_merge(slot, v),
                    _ => {
                        b.insert(k, v);
                    }
                }
            }
        }
        (b, p) => *b = p,
    }
}

/// HTTP API 是否在监听 + 配置里的地址(状态栏"● API")。
pub async fn api_status(state: &AppState) -> RelayResult<Value> {
    let cfg = state.config.read().await;
    let error = state.api_error.read().await.clone();
    respond(&serde_json::json!({
        "enabled": cfg.server.enabled,
        "listening": state.api_listening.load(std::sync::atomic::Ordering::Relaxed),
        "host": cfg.server.host,
        "port": cfg.server.port,
        "error": error,
    }))
}

pub fn find_available_port() -> RelayResult<Value> {
    let listener = std::net::TcpListener::bind(("127.0.0.1", 0)).map_err(|e| {
        RelayError::new("system_io_error", "无法查找可用端口，请稍后重试")
            .with_details(json!({ "io": e.to_string() }))
    })?;
    let port = listener
        .local_addr()
        .map_err(|e| {
            RelayError::new("system_io_error", "无法读取可用端口，请稍后重试")
                .with_details(json!({ "io": e.to_string() }))
        })?
        .port();
    Ok(json!({ "port": port }))
}

/// 使用 V88 配对码换取独立的执行器凭证。本机 HTTP API 的 token 不参与该请求。
pub async fn pair_v88(state: &AppState, base_url: String, code: String) -> RelayResult<Value> {
    let device_name = std::env::var("COMPUTERNAME").unwrap_or_else(|_| "一战晟铭执行器".into());
    let client = V88ExecutorClient::new(&base_url)
        .map_err(|e| RelayError::new("request_malformed", e.to_string()))?;
    let request = PairRequest::new(
        &code,
        &device_name,
        std::env::consts::OS,
        env!("CARGO_PKG_VERSION"),
    );
    let result = client
        .pair(&request)
        .await
        .map_err(|e| RelayError::new("v88_pair_failed", e.to_string()))?;
    let mut config = state.config.write().await;
    let mut next = config.clone();
    next.v88_executor.base_url = base_url.trim().trim_end_matches('/').into();
    next.v88_executor.executor_id = result.executor_id.clone();
    next.v88_executor.executor_token = result.token;
    next.save(&state.config_path)
        .map_err(|e| RelayError::new("system_io_error", format!("写入 V88 配对配置失败:{e}")))?;
    *config = next;
    Ok(json!({ "ok": true, "executorId": result.executor_id }))
}

#[tauri::command]
pub async fn app_api_status(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    api_status(&state).await
}

#[tauri::command]
pub fn app_find_available_port() -> RelayResult<Value> {
    find_available_port()
}

#[tauri::command]
pub async fn app_get_version(_state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    respond(&get_version())
}

#[tauri::command]
pub async fn app_get_config(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    get_config(&state).await
}

#[tauri::command]
pub async fn app_save_config(state: State<'_, Arc<AppState>>, patch: Value) -> RelayResult<Value> {
    save_config(&state, patch).await
}

#[tauri::command]
pub async fn app_pair_v88(
    state: State<'_, Arc<AppState>>,
    base_url: String,
    code: String,
) -> RelayResult<Value> {
    pair_v88(&state, base_url, code).await
}

#[tauri::command]
pub async fn app_regenerate_token(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    regenerate_token(&state).await
}

#[tauri::command]
pub async fn maintenance_gc(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    run_gc(&state).await
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::testing::test_state;
    use crate::commands::to_camel;

    #[test]
    fn version_has_spec_keys_after_camel() {
        let v = to_camel(get_version());
        assert_eq!(v["version"], env!("CARGO_PKG_VERSION"));
        assert!(matches!(
            v["buildTarget"].as_str().unwrap(),
            "macos" | "windows" | "linux"
        ));
        assert_eq!(v["tauriVersion"], tauri::VERSION);
        assert!(v["rustcVersion"].is_string());
        assert!(v.get("build_target").is_none());
    }

    #[tokio::test]
    async fn api_status_reflects_listening_flag() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let v = api_status(&state).await.unwrap();
        assert_eq!(v["enabled"], false);
        assert_eq!(v["listening"], false);
        assert_eq!(v["port"], 8787);
        assert_eq!(v["error"], Value::Null);
        state
            .api_listening
            .store(true, std::sync::atomic::Ordering::Relaxed);
        assert_eq!(api_status(&state).await.unwrap()["listening"], true);
    }

    #[test]
    fn available_port_can_be_bound_after_probe() {
        let value = find_available_port().unwrap();
        let port = value["port"].as_u64().unwrap() as u16;
        let listener = std::net::TcpListener::bind(("127.0.0.1", port)).unwrap();
        assert_eq!(listener.local_addr().unwrap().port(), port);
    }

    #[tokio::test]
    async fn get_config_returns_camel_config_with_token() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let v = get_config(&state).await.unwrap();
        assert_eq!(v["server"]["port"], 8787);
        assert_eq!(v["server"]["bearerToken"].as_str().unwrap().len(), 64);
        assert_eq!(v["doubao"]["cdpPort"], 1);
        assert_eq!(v["logging"]["cdpDebugLog"], true);
        assert!(v["server"].get("bearer_token").is_none());
    }

    #[tokio::test]
    async fn save_config_merges_patch_and_reports_restart() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let token_before = state.config().await.server.bearer_token.clone();

        // 不需要重启的项
        let r = save_config(&state, json!({ "ui": { "startMinimized": true } }))
            .await
            .unwrap();
        assert_eq!(r, json!({ "ok": true, "needsRestart": false }));
        let cfg = state.config().await;
        assert!(cfg.ui.start_minimized);
        assert_eq!(cfg.server.bearer_token, token_before, "未提交的字段不变");

        // 需要重启的项
        let r = save_config(&state, json!({ "server": { "port": 9000 } }))
            .await
            .unwrap();
        assert_eq!(r["needsRestart"], true);
        let cfg = state.config().await;
        assert_eq!(cfg.server.port, 9000);
        assert!(cfg.ui.start_minimized, "前一次补丁仍在");

        // 已写盘
        let on_disk = Config::load_or_create(&state.config_path).unwrap();
        assert_eq!(on_disk, cfg);
    }

    #[tokio::test]
    async fn save_config_rejects_invalid_values_without_touching_state() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let before = state.config().await;

        let e = save_config(&state, json!({ "server": { "host": "0.0.0.0" } }))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_malformed");
        assert!(e.details.unwrap()["reason"]
            .as_str()
            .unwrap()
            .contains("server.host"));

        let e = save_config(&state, json!({ "server": { "port": "abc" } }))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_malformed");

        let e = save_config(&state, json!([1, 2])).await.unwrap_err();
        assert_eq!(e.code, "request_malformed");

        assert_eq!(state.config().await, before);
        assert_eq!(
            Config::load_or_create(&state.config_path).unwrap(),
            before,
            "磁盘上的配置也不变"
        );
    }

    /// needsRestart 真值表:只有 HTTP 监听地址/端口与豆包 CDP 地址/端口需要重启。
    #[test]
    fn needs_restart_truth_table() {
        type Mutate = Box<dyn Fn(&mut Config)>;
        let base = Config::default();
        let cases: Vec<(&str, Mutate, bool)> = vec![
            (
                "server.enabled",
                Box::new(|c| c.server.enabled = true),
                true,
            ),
            ("server.port", Box::new(|c| c.server.port = 9000), true),
            (
                "server.host",
                Box::new(|c| c.server.host = "localhost".into()),
                true,
            ),
            (
                "doubao.cdp_port",
                Box::new(|c| c.doubao.cdp_port = 9708),
                true,
            ),
            (
                "doubao.cdp_address",
                Box::new(|c| c.doubao.cdp_address = "::1".into()),
                true,
            ),
            (
                "server.bearer_token(热生效)",
                Box::new(|c| c.server.bearer_token = "x".into()),
                false,
            ),
            (
                "logging.level(热生效)",
                Box::new(|c| c.logging.level = "debug".into()),
                false,
            ),
            (
                "logging.cdp_debug_log(热生效)",
                Box::new(|c| c.logging.cdp_debug_log = true),
                false,
            ),
            (
                "logging.file_retention_days",
                Box::new(|c| c.logging.file_retention_days = 3),
                false,
            ),
            (
                "concurrency.global_max",
                Box::new(|c| c.concurrency.global_max = 2),
                false,
            ),
            (
                "storage.result_retention_hours",
                Box::new(|c| c.storage.result_retention_hours = 48),
                false,
            ),
            (
                "doubao.auto_connect_on_startup",
                Box::new(|c| c.doubao.auto_connect_on_startup = false),
                false,
            ),
            (
                "ui.minimize_to_tray_on_close",
                Box::new(|c| c.ui.minimize_to_tray_on_close = false),
                false,
            ),
            ("无变化", Box::new(|_| {}), false),
        ];
        for (name, mutate, expected) in cases {
            let mut after = base.clone();
            mutate(&mut after);
            assert_eq!(needs_restart(&base, &after), expected, "{name}");
        }
    }

    #[tokio::test]
    async fn save_config_hot_applies_cdp_debug_toggle() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        assert!(state.cdp_log.enabled());
        let r = save_config(&state, json!({ "logging": { "cdpDebugLog": false } }))
            .await
            .unwrap();
        assert_eq!(r, json!({ "ok": true, "needsRestart": false }));
        assert!(!state.cdp_log.enabled());
        let tx = state.cdp_log.sender().unwrap();
        assert_eq!(tx.receiver_count(), 0);

        let r = save_config(&state, json!({ "logging": { "cdpDebugLog": true } }))
            .await
            .unwrap();
        assert_eq!(r["needsRestart"], false);
        assert!(state.cdp_log.enabled());
        for _ in 0..100 {
            if tx.receiver_count() == 1 {
                break;
            }
            tokio::time::sleep(std::time::Duration::from_millis(10)).await;
        }
        assert_eq!(tx.receiver_count(), 1, "收集任务已订阅");
        // 文件落在 log_dir/cdp-debug.log
        tx.send(crate::cdp::CdpLogEntry {
            ts: 1,
            direction: "send".into(),
            method: Some("Echo".into()),
            params: None,
            result: None,
            event: None,
        })
        .unwrap();
        let file = state
            .platform
            .log_dir()
            .join(crate::state::CDP_DEBUG_LOG_FILE);
        for _ in 0..100 {
            if file.exists() && !std::fs::read_to_string(&file).unwrap().is_empty() {
                break;
            }
            tokio::time::sleep(std::time::Duration::from_millis(10)).await;
        }
        assert!(std::fs::read_to_string(&file).unwrap().contains("\"Echo\""));
    }

    #[tokio::test]
    async fn save_config_level_change_is_hot_and_rejects_bad_level() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let r = save_config(&state, json!({ "logging": { "level": "debug" } }))
            .await
            .unwrap();
        assert_eq!(r, json!({ "ok": true, "needsRestart": false }));
        assert_eq!(state.config().await.logging.level, "debug");
        let e = save_config(&state, json!({ "logging": { "level": "verbose" } }))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_malformed");
        assert_eq!(state.config().await.logging.level, "debug");
    }

    #[tokio::test]
    async fn regenerate_token_replaces_token_everywhere() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let before = state.config().await.server.bearer_token.clone();
        let v = regenerate_token(&state).await.unwrap();
        let token = v["bearerToken"].as_str().unwrap().to_string();
        assert_eq!(token.len(), 64);
        assert!(token.chars().all(|c| c.is_ascii_hexdigit()));
        assert_ne!(token, before);
        assert_eq!(
            state.config().await.server.bearer_token,
            token,
            "内存已更新"
        );
        assert_eq!(
            Config::load_or_create(&state.config_path)
                .unwrap()
                .server
                .bearer_token,
            token,
            "已写盘"
        );
        let cfg = get_config(&state).await.unwrap();
        assert_eq!(cfg["server"]["bearerToken"], token);
    }

    #[tokio::test]
    async fn maintenance_gc_runs_and_returns_camel_report() {
        use crate::storage::Upload;
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let v = run_gc(&state).await.unwrap();
        assert_eq!(
            v,
            json!({ "resultFilesRemoved": 0, "tasksRemoved": 0, "uploadsRemoved": 0 })
        );
        // 放一条已过期的上传,再清理 → uploadsRemoved=1 且物理文件被删
        let storage = &state.storage;
        let path = dir.path().join("uploads").join("01EXPIRED.png");
        std::fs::write(&path, b"x").unwrap();
        storage
            .insert_upload(&Upload {
                id: "01EXPIRED".into(),
                mime: "image/png".into(),
                bytes: 1,
                path: path.clone(),
                created_at: 1,
                expires_at: 2,
            })
            .await
            .unwrap();
        let v = run_gc(&state).await.unwrap();
        assert_eq!(v["uploadsRemoved"], 1);
        assert!(!path.exists());
    }

    #[test]
    fn deep_merge_replaces_scalars_and_merges_objects() {
        let mut base = json!({ "a": { "x": 1, "y": 2 }, "b": 3, "c": [1] });
        deep_merge(
            &mut base,
            json!({ "a": { "y": 20, "z": 30 }, "b": null, "c": [2, 3], "d": "new" }),
        );
        assert_eq!(
            base,
            json!({ "a": { "x": 1, "y": 20, "z": 30 }, "b": null, "c": [2, 3], "d": "new" })
        );
    }
}
