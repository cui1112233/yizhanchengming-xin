//! `GET /health`(免鉴权):`{status, version, uptime_seconds}`。

use std::sync::Arc;

use axum::extract::State;
use axum::routing::get;
use axum::{Json, Router};
use serde_json::{json, Value};

use crate::logging::now_ms;
use crate::state::AppState;

pub fn router() -> Router<Arc<AppState>> {
    Router::new().route("/health", get(health))
}

/// 启动至今的秒数(向下取整)。
pub fn uptime_seconds(started_at: i64, now: i64) -> u64 {
    ((now - started_at).max(0) / 1000) as u64
}

async fn health(State(state): State<Arc<AppState>>) -> Json<Value> {
    Json(json!({
        "status": "ok",
        "version": env!("CARGO_PKG_VERSION"),
        "uptime_seconds": uptime_seconds(state.started_at, now_ms()),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn uptime_rounds_down_and_never_negative() {
        assert_eq!(uptime_seconds(1_000, 4_999), 3);
        assert_eq!(uptime_seconds(1_000, 1_000), 0);
        assert_eq!(uptime_seconds(5_000, 1_000), 0);
    }
}
