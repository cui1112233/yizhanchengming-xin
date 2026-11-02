//! HTTP 错误响应:`RelayError` → `{ error: { code, message, details } }`,状态码按
//! `docs/spec/11-error-codes.md` 的对照表(`RelayError::http_status`);以及请求体解析的
//! `request_*` 错误构造(JSON 坏 → `request_malformed`,serde 缺字段 → `request_missing_field`)。

use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde::de::DeserializeOwned;
use serde_json::{json, Map, Value};

use crate::error::{RelayError, RelayResult};
use crate::tasks::validate::missing_field;

impl IntoResponse for RelayError {
    fn into_response(self) -> Response {
        let status =
            StatusCode::from_u16(self.http_status()).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
        (status, Json(json!({ "error": self }))).into_response()
    }
}

/// `request_malformed{reason}`(400)。
pub fn malformed(reason: impl Into<String>) -> RelayError {
    RelayError::new(
        "request_malformed",
        "请求格式错误,请按接口文档提交 JSON 请求体",
    )
    .with_details(json!({ "reason": reason.into() }))
}

/// 未知路径(404,JSON 体;11-error-codes 没有专门的"接口不存在"码,沿用 `request_malformed`)。
pub fn route_not_found() -> Response {
    (
        StatusCode::NOT_FOUND,
        Json(json!({ "error": RelayError::new(
            "request_malformed",
            "接口不存在,请检查请求路径与方法"
        ) })),
    )
        .into_response()
}

/// 已知路径但方法不对(405,JSON 体;同样沿用 `request_malformed`)。
pub fn method_not_allowed() -> Response {
    (
        StatusCode::METHOD_NOT_ALLOWED,
        Json(json!({ "error": RelayError::new(
            "request_malformed",
            "该接口不支持此请求方法,请检查接口文档中的方法"
        ) })),
    )
        .into_response()
}

/// 把请求体解析成 JSON 对象。`allow_empty` 时空 body 视为 `{}`(可选 body 的端点)。
pub fn parse_json_object(bytes: &[u8], allow_empty: bool) -> RelayResult<Map<String, Value>> {
    if bytes.iter().all(|b| b.is_ascii_whitespace()) {
        if allow_empty {
            return Ok(Map::new());
        }
        return Err(malformed("请求体为空"));
    }
    match serde_json::from_slice::<Value>(bytes) {
        Ok(Value::Object(map)) => Ok(map),
        Ok(other) => Err(malformed(format!(
            "请求体必须是 JSON 对象,实际为 {}",
            value_kind(&other)
        ))),
        Err(e) => Err(malformed(format!("JSON 解析失败:{e}"))),
    }
}

/// `serde_json::from_value` 的错误映射:`missing field \`x\`` → `request_missing_field{field:x}`,
/// 其它(类型不对、未知枚举值等)→ `request_malformed{reason}`。
pub fn from_json_value<T: DeserializeOwned>(value: Value) -> RelayResult<T> {
    serde_json::from_value(value).map_err(|e| {
        let text = e.to_string();
        match missing_field_name(&text) {
            Some(field) => missing_field(field),
            None => malformed(text),
        }
    })
}

/// 从 serde 错误文本里提取 `missing field \`name\`` 的字段名。
pub fn missing_field_name(text: &str) -> Option<&str> {
    let rest = text.strip_prefix("missing field `")?;
    let end = rest.find('`')?;
    Some(&rest[..end])
}

fn value_kind(v: &Value) -> &'static str {
    match v {
        Value::Null => "null",
        Value::Bool(_) => "boolean",
        Value::Number(_) => "number",
        Value::String(_) => "string",
        Value::Array(_) => "array",
        Value::Object(_) => "object",
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use http_body_util::BodyExt;
    use serde::Deserialize;

    #[tokio::test]
    async fn relay_error_renders_error_envelope_with_mapped_status() {
        let resp = RelayError::doubao_not_ready().into_response();
        assert_eq!(resp.status(), StatusCode::SERVICE_UNAVAILABLE);
        assert_eq!(
            resp.headers()["content-type"].to_str().unwrap(),
            "application/json"
        );
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let v: Value = serde_json::from_slice(&body).unwrap();
        assert_eq!(v["error"]["code"], "doubao_not_ready");
        assert!(v["error"]["message"].is_string());
        assert!(v["error"]["details"].is_null());
        assert_eq!(v.as_object().unwrap().len(), 1);

        let resp = RelayError::account_not_found("a", vec!["b".into()]).into_response();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
        let resp = malformed("x").into_response();
        assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
    }

    #[test]
    fn parse_json_object_handles_empty_bad_and_non_object() {
        assert!(parse_json_object(b"", true).unwrap().is_empty());
        assert!(parse_json_object(b"  \n", true).unwrap().is_empty());
        let e = parse_json_object(b"", false).unwrap_err();
        assert_eq!(e.code, "request_malformed");
        let e = parse_json_object(b"{bad", false).unwrap_err();
        assert_eq!(e.code, "request_malformed");
        assert!(e.details.unwrap()["reason"]
            .as_str()
            .unwrap()
            .contains("JSON"));
        let e = parse_json_object(b"[1,2]", false).unwrap_err();
        assert_eq!(e.code, "request_malformed");
        assert!(e.details.unwrap()["reason"]
            .as_str()
            .unwrap()
            .contains("array"));
        let m = parse_json_object(br#"{"a":1}"#, false).unwrap();
        assert_eq!(m["a"], 1);
    }

    #[test]
    fn from_json_value_maps_missing_field_and_type_errors() {
        #[derive(Deserialize, Debug)]
        #[allow(dead_code)]
        struct S {
            account: String,
            n: u32,
        }
        let e = from_json_value::<S>(json!({ "n": 1 })).unwrap_err();
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "account");
        let e = from_json_value::<S>(json!({ "account": "a", "n": "x" })).unwrap_err();
        assert_eq!(e.code, "request_malformed");
        let s: S = from_json_value(json!({ "account": "a", "n": 2 })).unwrap();
        assert_eq!(s.n, 2);
        assert_eq!(
            missing_field_name("missing field `prompt` at line 1"),
            Some("prompt")
        );
        assert_eq!(missing_field_name("invalid type"), None);
    }
}
