//! `GET /doubao/status` / `POST /doubao/reconnect`。
//!
//! `reconnect` 可选 body `{force: bool}`(默认 false):false = 未运行则启动、已运行则重建会话,
//! true = 杀掉并重启(与 Tauri 命令 `doubao_reconnect` 一致;spec 04 只写了重启语义)。

use std::sync::Arc;

use axum::body::Bytes;
use axum::extract::State;
use axum::routing::{get, post};
use axum::{Json, Router};
use serde::Deserialize;
use serde_json::{json, Value};

use crate::doubao::DoubaoStatus;
use crate::error::RelayResult;
use crate::http::error::{from_json_value, parse_json_object};
use crate::logging::now_ms;
use crate::state::AppState;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/doubao/status", get(status))
        .route("/doubao/reconnect", post(reconnect))
}

#[derive(Deserialize, Default)]
struct ReconnectBody {
    #[serde(default)]
    force: bool,
}

async fn status(State(state): State<Arc<AppState>>) -> Json<DoubaoStatus> {
    Json(state.doubao.status().await)
}

async fn reconnect(State(state): State<Arc<AppState>>, body: Bytes) -> RelayResult<Json<Value>> {
    let body: ReconnectBody = from_json_value(Value::Object(parse_json_object(&body, true)?))?;
    let action = state.doubao.connect(body.force).await?;
    Ok(Json(json!({ "action": action, "cdp_ready_at": now_ms() })))
}
