//! `GET /accounts` / `POST /accounts/switch`。
//!
//! 切换账号会整页 reload(Task 1.2 实测),正在运行的任务会丢失页面内提交状态,所以只要有
//! **任何**非终态任务在运行就拒绝切换(409 `account_busy{account, running_task_ids}`),
//! 不只是 spec 字面的"目标账号有进行中任务"。守卫在 `AppState::switch_account`
//! (`Scheduler::begin_switch` 返回的 `SwitchGuard`:检查与标记切换在同一把锁下,守卫存活期间暂停派发),
//! Tauri `account_switch` 走同一条路。

use std::sync::Arc;

use axum::body::Bytes;
use axum::extract::State;
use axum::routing::{get, post};
use axum::{Json, Router};
use serde::Deserialize;
use serde_json::Value;

use crate::doubao::{AccountList, SwitchResult};
use crate::error::RelayResult;
use crate::http::error::{from_json_value, parse_json_object};
use crate::state::AppState;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/accounts", get(list))
        .route("/accounts/switch", post(switch))
}

#[derive(Deserialize)]
struct SwitchBody {
    target: String,
}

async fn list(State(state): State<Arc<AppState>>) -> RelayResult<Json<AccountList>> {
    Ok(Json(state.doubao.accounts(false).await?))
}

async fn switch(
    State(state): State<Arc<AppState>>,
    body: Bytes,
) -> RelayResult<Json<SwitchResult>> {
    let body: SwitchBody = from_json_value(Value::Object(parse_json_object(&body, false)?))?;
    if body.target.trim().is_empty() {
        return Err(crate::tasks::validate::missing_field("target"));
    }
    Ok(Json(state.switch_account(&body.target).await?))
}
