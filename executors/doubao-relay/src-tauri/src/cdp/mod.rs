//! CDP(Chrome DevTools Protocol)薄封装:单 target 的 WebSocket 会话 + `/json` 发现接口。
//!
//! 契约见 `docs/spec/09-cdp-contract.md`。不用现成的 CDP 库,只做:
//! - `session`:`{id, method, params}` 收发、按 id 匹配响应、事件广播、超时、断线感知;
//! - `discovery`:`GET /json/version`、`GET /json/list` 与 target 白名单校验。

pub mod discovery;
pub mod json_fix;
pub mod session;

use serde::{Deserialize, Serialize};
use serde_json::Value;

pub use discovery::{
    check_version, list_targets, select_pages, validate_target, CdpTarget, DoubaoPage,
};
pub use session::{CdpEvent, CdpSession};

use crate::error::RelayError;

/// CDP 层错误。上层通过 `From<CdpError> for RelayError` 映射到对外错误码。
#[derive(Debug, thiserror::Error)]
pub enum CdpError {
    #[error("连接失败:{0}")]
    Connect(String),
    #[error("连接已关闭")]
    Closed,
    #[error("命令超时:{0}")]
    Timeout(String),
    #[error("CDP 错误 {code}:{message}")]
    Protocol { code: i64, message: String },
    #[error("页面脚本异常:{0}")]
    JsException(String),
    #[error("返回值解析失败:{0}")]
    Deserialize(String),
    #[error("HTTP 请求失败:{0}")]
    Http(String),
    #[error("目标不合法:{0}")]
    InvalidTarget(String),
}

/// CDP 调试日志条目(`cdp_debug_log=true` 时每条收发各一条),
/// 结构与 `docs/spec/05-tauri-ipc.md` `debug.cdp_log` 的 entry 同形。
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CdpLogEntry {
    /// Unix epoch 毫秒。
    pub ts: i64,
    /// `send` / `recv`。
    pub direction: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub method: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub params: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub event: Option<Value>,
}

/// 映射规则(计划 Task 1.3 已知问题):Closed/Connect→`doubao_cdp_disconnected`,
/// Timeout→`doubao_not_ready`,JsException→`doubao_ui_changed`;其余归为豆包未就绪/内部错误。
impl From<CdpError> for RelayError {
    fn from(e: CdpError) -> Self {
        match e {
            CdpError::Closed | CdpError::Connect(_) => RelayError::doubao_cdp_disconnected()
                .with_details(serde_json::json!({ "cdp": e.to_string() })),
            CdpError::Timeout(_) | CdpError::Http(_) | CdpError::InvalidTarget(_) => {
                RelayError::doubao_not_ready()
                    .with_details(serde_json::json!({ "cdp": e.to_string() }))
            }
            CdpError::JsException(ref msg) => {
                let msg = msg.clone();
                RelayError::doubao_ui_changed(vec![]).with_details(serde_json::json!({
                    "failed_selectors": [],
                    "js_exception": msg,
                }))
            }
            CdpError::Protocol { .. } | CdpError::Deserialize(_) => {
                RelayError::internal(format!("CDP 返回异常:{e}"))
                    .with_details(serde_json::json!({ "cdp": e.to_string() }))
            }
        }
    }
}

/// 当前 Unix epoch 毫秒。
pub(crate) fn now_ms() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as i64)
        .unwrap_or(0)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn cdp_error_maps_to_relay_error_codes() {
        let e: RelayError = CdpError::Closed.into();
        assert_eq!(e.code, "doubao_cdp_disconnected");
        let e: RelayError = CdpError::Connect("refused".into()).into();
        assert_eq!(e.code, "doubao_cdp_disconnected");
        let e: RelayError = CdpError::Timeout("Runtime.evaluate".into()).into();
        assert_eq!(e.code, "doubao_not_ready");
        let e: RelayError = CdpError::JsException("中文错误".into()).into();
        assert_eq!(e.code, "doubao_ui_changed");
        assert_eq!(e.details.unwrap()["js_exception"], "中文错误");
        let e: RelayError = CdpError::Protocol {
            code: -32601,
            message: "no".into(),
        }
        .into();
        assert_eq!(e.code, "system_internal_error");
        let e: RelayError = CdpError::Http("x".into()).into();
        assert_eq!(e.code, "doubao_not_ready");
    }

    #[test]
    fn log_entry_omits_absent_fields() {
        let entry = CdpLogEntry {
            ts: 1,
            direction: "send".into(),
            method: Some("Echo.ping".into()),
            params: Some(serde_json::json!({ "a": 1 })),
            result: None,
            event: None,
        };
        let v = serde_json::to_value(entry).unwrap();
        let obj = v.as_object().unwrap();
        assert_eq!(obj.len(), 4);
        assert!(!obj.contains_key("result"));
        assert!(!obj.contains_key("event"));
    }
}
