//! 隐私声明(`docs/spec/12-security.md` "隐私声明":首次启动显示,MUST 明确同意一次才可用)
//! 与退出命令。
//!
//! - 同意状态存 `metadata.privacy_accepted`(值为同意时刻的 Unix epoch 毫秒),启动时
//!   [`load`] 载入到 `AppState.privacy`;
//! - 同意前 `lib.rs::spawn_background` 不启动 HTTP API、不自动连豆包([`wait_accepted`]);
//! - 前端在 `accepted=false` 时显示全屏声明,"同意并继续" → `app_privacy_accept`,
//!   "不同意并退出" → `app_quit`(走关闭流程)。

use std::sync::Arc;

use serde_json::{json, Value};
use tauri::{AppHandle, Runtime, State};

use crate::error::{RelayError, RelayResult};
use crate::logging::now_ms;
use crate::state::AppState;

/// `metadata` 表里的键。
pub const PRIVACY_META_KEY: &str = "privacy_accepted";

/// 从库里读同意时刻(没有 / 不是数字 → None),并同步到 `state.privacy`。
pub async fn load(state: &AppState) -> RelayResult<Option<i64>> {
    let at = state
        .storage
        .get_meta(PRIVACY_META_KEY)
        .await?
        .and_then(|v| v.trim().parse::<i64>().ok());
    state.privacy.send_replace(at.is_some());
    Ok(at)
}

/// `app.privacy_status` → `{accepted, acceptedAt}`。
pub async fn status(state: &AppState) -> RelayResult<Value> {
    let at = load(state).await?;
    Ok(json!({ "accepted": at.is_some(), "acceptedAt": at }))
}

/// `app.privacy_accept`:写入同意时刻(已同意过则保留首次时刻)→ 放行 HTTP API / 自动连接。
pub async fn accept(state: &AppState) -> RelayResult<Value> {
    if let Some(at) = load(state).await? {
        return Ok(json!({ "accepted": true, "acceptedAt": at }));
    }
    let now = now_ms();
    state
        .storage
        .set_meta(PRIVACY_META_KEY, &now.to_string(), now)
        .await?;
    state.privacy.send_replace(true);
    tracing::info!("用户已同意隐私声明");
    Ok(json!({ "accepted": true, "acceptedAt": now }))
}

/// 等到用户同意隐私声明(已同意则立即返回)。
pub async fn wait_accepted(state: &AppState) {
    let mut rx = state.privacy.subscribe();
    // send_replace 不会关闭通道;Sender 与 state 同生命周期
    let _ = rx.wait_for(|accepted| *accepted).await;
}

#[tauri::command]
pub async fn app_privacy_status(state: State<'_, Arc<AppState>>) -> Result<Value, RelayError> {
    status(&state).await
}

#[tauri::command]
pub async fn app_privacy_accept(state: State<'_, Arc<AppState>>) -> Result<Value, RelayError> {
    accept(&state).await
}

/// `app.quit`:走完整关闭流程后退出(隐私声明"不同意并退出"、托盘"退出"同一路径)。
#[tauri::command]
pub async fn app_quit<R: Runtime>(app: AppHandle<R>) -> Result<Value, RelayError> {
    crate::shutdown::request_quit(&app);
    Ok(json!({ "ok": true }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    use crate::commands::testing::test_state;

    #[tokio::test]
    async fn not_accepted_until_accept_then_persisted() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        assert_eq!(
            status(&state).await.unwrap(),
            json!({ "accepted": false, "acceptedAt": null })
        );
        assert!(!*state.privacy.borrow());

        let waiter = {
            let state = Arc::clone(&state);
            tokio::spawn(async move { wait_accepted(&state).await })
        };
        tokio::time::sleep(Duration::from_millis(30)).await;
        assert!(!waiter.is_finished(), "同意前一直等待");

        let v = accept(&state).await.unwrap();
        assert_eq!(v["accepted"], true);
        let at = v["acceptedAt"].as_i64().unwrap();
        assert!(at > 0);
        tokio::time::timeout(Duration::from_secs(1), waiter)
            .await
            .expect("同意后 wait_accepted 立即返回")
            .unwrap();
        assert_eq!(
            state.storage.get_meta(PRIVACY_META_KEY).await.unwrap(),
            Some(at.to_string())
        );

        // 再次同意保留首次时刻
        let again = accept(&state).await.unwrap();
        assert_eq!(again["acceptedAt"], at);
    }

    #[tokio::test]
    async fn load_restores_acceptance_from_metadata() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        state
            .storage
            .set_meta(PRIVACY_META_KEY, "1700000000000", 1)
            .await
            .unwrap();
        assert_eq!(load(&state).await.unwrap(), Some(1_700_000_000_000));
        assert!(*state.privacy.borrow());
        // 已同意:wait_accepted 立即返回
        tokio::time::timeout(Duration::from_millis(200), wait_accepted(&state))
            .await
            .unwrap();
        // 值损坏视为未同意(重新弹声明)
        state
            .storage
            .set_meta(PRIVACY_META_KEY, "", 2)
            .await
            .unwrap();
        assert_eq!(load(&state).await.unwrap(), None);
        assert!(!*state.privacy.borrow());
    }
}
