//! `doubao.*` 命令:连接状态 / 重连。

use std::sync::Arc;

use serde_json::{json, Value};
use tauri::State;

use crate::error::RelayResult;
use crate::logging::now_ms;
use crate::state::AppState;

use super::respond;

/// `doubao.status`。
pub async fn status(state: &AppState) -> RelayResult<Value> {
    respond(&state.doubao.status().await)
}

/// `doubao.reconnect` → `{action, cdpReadyAt}`;`force=true` 杀掉并重启豆包。
pub async fn reconnect(state: &AppState, force: bool) -> RelayResult<Value> {
    let action = state.doubao.connect(force).await?;
    respond(&json!({ "action": action, "cdp_ready_at": now_ms() }))
}

#[tauri::command]
pub async fn doubao_status(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    status(&state).await
}

#[tauri::command]
pub async fn doubao_reconnect(
    state: State<'_, Arc<AppState>>,
    force: Option<bool>,
) -> RelayResult<Value> {
    reconnect(&state, force.unwrap_or(false)).await
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::testing::test_state;

    #[tokio::test]
    async fn status_is_camel_and_initially_disconnected() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let v = status(&state).await.unwrap();
        assert_eq!(v["running"], false);
        assert_eq!(v["cdpConnected"], false);
        assert_eq!(v["cdpPort"], 1);
        assert_eq!(v["uiCompatCheck"], "unknown");
        assert_eq!(v["reconnecting"], false);
        assert_eq!(v["chatTargets"], 0);
        assert!(v["doubaoVersionDetected"].is_null());
        assert!(v["uiCompatDetails"].get("accountButtonFound").is_some());
        assert!(v.get("cdp_connected").is_none());
    }

    #[tokio::test]
    async fn reconnect_without_doubao_fails_with_doubao_code() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let e = reconnect(&state, false).await.unwrap_err();
        assert!(e.code.starts_with("doubao_"), "{e:?}");
        // force=true 先找豆包可执行文件,假平台找不到 → doubao_not_found
        let e = reconnect(&state, true).await.unwrap_err();
        assert_eq!(e.code, "doubao_not_found");
    }
}
