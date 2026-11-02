//! `account.*` 命令:列表(2s 缓存)/ 强制刷新 / 切换。

use std::sync::Arc;

use serde_json::Value;
use tauri::State;

use crate::error::RelayResult;
use crate::state::AppState;

use super::respond;

/// `account.list`(走缓存)。
pub async fn list(state: &AppState) -> RelayResult<Value> {
    respond(&state.doubao.accounts(false).await?)
}

/// `account.refresh`(跳过缓存)。
pub async fn refresh(state: &AppState) -> RelayResult<Value> {
    respond(&state.doubao.accounts(true).await?)
}

/// `account.switch` → `{ok, previous, current, switchedAt}`;有任务运行中 → `account_busy`
/// (spec 05 `AccountBusy`,与 HTTP `POST /accounts/switch` 共用 `AppState::switch_account`)。
pub async fn switch(state: &AppState, target: &str) -> RelayResult<Value> {
    respond(&state.switch_account(target).await?)
}

#[tauri::command]
pub async fn account_list(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    list(&state).await
}

#[tauri::command]
pub async fn account_refresh(state: State<'_, Arc<AppState>>) -> RelayResult<Value> {
    refresh(&state).await
}

#[tauri::command]
pub async fn account_switch(state: State<'_, Arc<AppState>>, target: String) -> RelayResult<Value> {
    switch(&state, &target).await
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::testing::test_state;

    #[tokio::test]
    async fn list_without_doubao_returns_doubao_error_code() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let e = list(&state).await.unwrap_err();
        assert!(e.code.starts_with("doubao_"), "{e:?}");
        let e = refresh(&state).await.unwrap_err();
        assert!(e.code.starts_with("doubao_"), "{e:?}");
        let e = switch(&state, "张三").await.unwrap_err();
        assert!(e.code.starts_with("doubao_"), "{e:?}");
    }

    /// 与 `http::tests::switch_is_refused_while_a_task_is_running` 对应:IPC 路径同样受守卫。
    #[tokio::test]
    async fn switch_is_refused_while_a_task_is_running() {
        use std::time::Duration;

        use crate::commands::testing::{test_state_with, FakeBackend};
        use crate::tasks::{ConfirmationMode, GenerateRequest, TaskType};

        let dir = tempfile::tempdir().unwrap();
        let state = test_state_with(dir.path(), Some(FakeBackend::new(&["张三", "李四"]))).await;
        // 真调度循环 + FakeBackend 的 submit 挂起到取消 → 任务停在 preparing
        state.scheduler.spawn_loop();
        let (id, _, _) = state
            .scheduler
            .submit(GenerateRequest {
                account: "张三".into(),
                task_type: TaskType::Image,
                model_name: "Seedream 5.0 Pro".into(),
                prompt: "一只蓝色小鸟".into(),
                ratio: "1:1".into(),
                duration: None,
                reference_files: vec![],
                confirmation_mode: ConfirmationMode::Auto,
                auto_use_membership: true,
                callback_url: None,
            })
            .await
            .unwrap();
        let mut running = Vec::new();
        for _ in 0..200 {
            running = state.scheduler.running_task_ids(None);
            if !running.is_empty() {
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        assert_eq!(running, vec![id.clone()], "任务应已被派发");

        let e = switch(&state, "李四").await.unwrap_err();
        assert_eq!(e.code, "account_busy");
        let d = e.details.unwrap();
        assert_eq!(d["account"], "李四");
        assert_eq!(d["running_task_ids"], serde_json::json!([id]));

        // 取消后守卫放行,切换走到豆包连接(测试里连不上 → doubao_*)
        state.scheduler.cancel(&id).await.unwrap();
        for _ in 0..200 {
            if state.scheduler.running_task_ids(None).is_empty() {
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let e = switch(&state, "李四").await.unwrap_err();
        assert!(e.code.starts_with("doubao_"), "{e:?}");
    }
}
