//! `POST /generate`:请求体先按 `serde_json::Value` 读(坏 JSON → `request_malformed`),
//! 枚举字段(`type` / `confirmation_mode`)先行检查给出 `request_invalid_enum`,再反序列化成
//! `GenerateRequest`(缺必填字段 → `request_missing_field`),最后交给 `Scheduler::submit`
//! (白名单 / 范围 / 账号 / 上传 / 队列深度校验都在那里)。

use std::sync::Arc;

use axum::body::Bytes;
use axum::extract::State;
use axum::routing::post;
use axum::{Json, Router};
use serde_json::{json, Map, Value};

use crate::error::RelayResult;
use crate::http::error::{from_json_value, malformed, parse_json_object};
use crate::state::AppState;
use crate::tasks::validate::invalid_enum;
use crate::tasks::{GenerateRequest, TaskStatus};

pub fn router() -> Router<Arc<AppState>> {
    Router::new().route("/generate", post(generate))
}

const TYPES: &[&str] = &["image", "video"];
const CONFIRMATION_MODES: &[&str] = &["auto", "manual"];

/// 枚举字段若存在必须是白名单内的字符串;其它类型 → `request_malformed`。
fn check_enum(obj: &Map<String, Value>, field: &str, allowed: &[&str]) -> RelayResult<()> {
    match obj.get(field) {
        None | Some(Value::Null) => Ok(()),
        Some(Value::String(s)) if allowed.contains(&s.as_str()) => Ok(()),
        Some(Value::String(s)) => Err(invalid_enum(field, s, allowed)),
        Some(other) => Err(malformed(format!(
            "字段 {field} 必须是字符串,实际为 {other}"
        ))),
    }
}

/// 请求体 → `GenerateRequest`(纯函数,便于测试)。
pub fn parse_generate_request(bytes: &[u8]) -> RelayResult<GenerateRequest> {
    let obj = parse_json_object(bytes, false)?;
    generate_request_from_object(obj)
}

/// 已解析的 JSON 对象(snake_case 字段)→ `GenerateRequest`:枚举先行检查 → serde
/// (IPC `task_submit` 把 camelCase 还原成 snake_case 后也走这里,两端错误码一致)。
pub fn generate_request_from_object(obj: Map<String, Value>) -> RelayResult<GenerateRequest> {
    check_enum(&obj, "type", TYPES)?;
    check_enum(&obj, "confirmation_mode", CONFIRMATION_MODES)?;
    from_json_value(Value::Object(obj))
}

async fn generate(State(state): State<Arc<AppState>>, body: Bytes) -> RelayResult<Json<Value>> {
    let req = parse_generate_request(&body)?;
    let (task_id, queue_position, created_at) = state.submit_task(req).await?;
    Ok(Json(json!({
        "task_id": task_id,
        "status": TaskStatus::Pending.as_str(),
        "queue_position": queue_position,
        "created_at": created_at,
    })))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::tasks::{ConfirmationMode, TaskType};

    #[test]
    fn parse_maps_missing_field_bad_enum_and_malformed() {
        let e = parse_generate_request(br#"{"type":"image"}"#).unwrap_err();
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "account");

        let e = parse_generate_request(
            br#"{"account":"a","type":"audio","model_name":"m","prompt":"p","ratio":"1:1"}"#,
        )
        .unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "type");
        assert_eq!(d["value"], "audio");
        assert_eq!(d["allowed"], json!(["image", "video"]));

        let e = parse_generate_request(
            br#"{"account":"a","type":"image","model_name":"m","prompt":"p","ratio":"1:1","confirmation_mode":"sometimes"}"#,
        )
        .unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "confirmation_mode");

        let e = parse_generate_request(br#"{"account":"a","type":5}"#).unwrap_err();
        assert_eq!(e.code, "request_malformed");
        let e = parse_generate_request(b"not json").unwrap_err();
        assert_eq!(e.code, "request_malformed");
        let e = parse_generate_request(
            br#"{"account":"a","type":"image","model_name":"m","prompt":5,"ratio":"1:1"}"#,
        )
        .unwrap_err();
        assert_eq!(e.code, "request_malformed");
    }

    #[test]
    fn parse_accepts_full_request_with_defaults() {
        let r = parse_generate_request(
            r#"{"account":"张三","type":"video","model_name":"Seedance 2.5","prompt":"p","ratio":"16:9","duration":10}"#.as_bytes(),
        )
        .unwrap();
        assert_eq!(r.task_type, TaskType::Video);
        assert_eq!(r.duration, Some(10));
        assert_eq!(r.confirmation_mode, ConfirmationMode::Auto);
        assert!(r.auto_use_membership);
        assert!(r.reference_files.is_empty());
    }
}
