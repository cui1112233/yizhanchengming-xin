//! 统一错误类型 `RelayError`。
//!
//! 错误码取自 `docs/spec/11-error-codes.md`,对外序列化为 `{ code, message, details }`;
//! HTTP 状态码按该文档的"HTTP 状态码对照"前缀表推导。
//! message 为中文、不带句末标点、尽量附一句下一步建议。

use serde::ser::SerializeStruct;
use serde::{Serialize, Serializer};
use serde_json::{json, Value};

/// `doubao_crashed` 的唯一文案(崩溃监听 `spawn_crash_watcher` 与 runner `crash_aware_error` 共用)。
pub const DOUBAO_CRASHED_MESSAGE: &str = "豆包已退出,进行中的任务已中止,请重新启动豆包后再提交任务";

/// 后端统一错误。Tauri 命令与 HTTP 层都直接返回它。
#[derive(Debug, Clone, thiserror::Error)]
#[error("{message}")]
pub struct RelayError {
    /// 稳定的错误码(契约),见 11-error-codes。
    pub code: &'static str,
    /// 面向用户的中文消息。
    pub message: String,
    /// 可选的结构化排查信息。
    pub details: Option<Value>,
}

/// `Result<T, RelayError>` 的简写。
pub type RelayResult<T> = Result<T, RelayError>;

/// IPC(Tauri 命令)错误与 `RelayError` 同结构。
pub type IpcError = RelayError;

impl RelayError {
    pub fn new(code: &'static str, message: impl Into<String>) -> Self {
        Self {
            code,
            message: message.into(),
            details: None,
        }
    }

    pub fn with_details(mut self, details: Value) -> Self {
        self.details = Some(details);
        self
    }

    /// 按 11-error-codes "HTTP 状态码对照"推导 HTTP 状态码。
    pub fn http_status(&self) -> u16 {
        http_status_for_code(self.code)
    }

    // ---- 豆包类(503) ----

    pub fn doubao_not_running() -> Self {
        Self::new(
            "doubao_not_running",
            "豆包未运行,请先启动豆包或点击重连豆包",
        )
    }

    pub fn doubao_not_ready() -> Self {
        Self::new("doubao_not_ready", "豆包尚未就绪,请等待豆包连接完成后重试")
    }

    pub fn doubao_not_logged_in() -> Self {
        Self::new(
            "doubao_not_logged_in",
            "豆包尚未登录,请在豆包客户端完成登录后重试",
        )
    }

    pub fn doubao_cdp_disconnected() -> Self {
        Self::new(
            "doubao_cdp_disconnected",
            "与豆包的连接已断开,正在尝试重连,请稍后重试",
        )
    }

    pub fn doubao_ui_changed(failed: Vec<&str>) -> Self {
        Self::new(
            "doubao_ui_changed",
            "豆包界面结构发生变化,请升级 doubao-relay 以适配新版",
        )
        .with_details(json!({ "failed_selectors": failed }))
    }

    /// `doubao_crashed{last_seen_at}`(文案 [`DOUBAO_CRASHED_MESSAGE`]):豆包进程退出(UC-32)。`last_seen_at` 为 0(从未连通)时写 null。
    pub fn doubao_crashed(last_seen_at: i64) -> Self {
        Self::new("doubao_crashed", DOUBAO_CRASHED_MESSAGE)
            .with_details(json!({ "last_seen_at": (last_seen_at > 0).then_some(last_seen_at) }))
    }

    pub fn doubao_launch_timeout() -> Self {
        Self::new(
            "doubao_launch_timeout",
            "启动豆包后 20 秒内仍未就绪,请检查豆包是否正常打开后重试",
        )
    }

    // ---- 账号类(404/409) ----

    pub fn account_not_found(requested: &str, available: Vec<String>) -> Self {
        Self::new(
            "account_not_found",
            format!("指定的账号 {requested} 未在当前豆包中找到,请确认账号名或先在豆包中登录该账号"),
        )
        .with_details(json!({ "requested": requested, "available": available }))
    }

    pub fn account_ambiguous(requested: &str, matches: Vec<String>) -> Self {
        Self::new(
            "account_ambiguous",
            format!("账号名 {requested} 在豆包中存在多个,请使用唯一的账号名"),
        )
        .with_details(json!({ "requested": requested, "matches": matches }))
    }

    pub fn account_switch_failed(reason: &str) -> Self {
        Self::new(
            "account_switch_failed",
            "账号切换失败,请在豆包中确认目标账号登录状态后重试",
        )
        .with_details(json!({ "reason": reason }))
    }

    pub fn account_switch_timeout(elapsed_ms: u64) -> Self {
        Self::new(
            "account_switch_timeout",
            "账号切换超时,请检查豆包是否卡住后重试",
        )
        .with_details(json!({ "elapsed_ms": elapsed_ms }))
    }

    // ---- 系统类(500) ----

    /// `system_internal_error`。
    pub fn internal(msg: impl Into<String>) -> Self {
        Self::new("system_internal_error", msg)
    }
}

/// spec 08 超时表的短名(`queue_timeout` 等)→ 11-error-codes 体系下实际写入的 `generation_*` 码
/// (11 没列这几个短名;调度器 / runner 只写正式码,别名供按 spec 08 名字查询或映射时用)。其它码原样返回。
pub fn canonical_code(code: &str) -> &str {
    match code {
        "queue_timeout" => "generation_queue_timeout",
        "preparing_timeout" => "generation_preparing_timeout",
        "task_timeout" => "generation_task_timeout",
        "confirm_timeout" => "generation_confirm_timeout",
        other => other,
    }
}

/// 11-error-codes "HTTP 状态码对照":精确码优先,再按前缀(spec 08 短名先经 [`canonical_code`])。
pub fn http_status_for_code(code: &str) -> u16 {
    let code = canonical_code(code);
    match code {
        "account_not_found" | "upload_not_found" | "task_not_found" | "result_not_available" => {
            return 404
        }
        "account_busy" | "task_not_cancellable" | "task_not_waiting_confirm" => return 409,
        "queue_full" => return 429,
        _ => {}
    }
    if code.starts_with("request_") {
        400
    } else if code.starts_with("auth_") {
        401
    } else if code.starts_with("doubao_") {
        503
    } else if code.starts_with("system_") || code.starts_with("platform_") {
        500
    } else if code.starts_with("upload_") {
        400
    } else if code.starts_with("result_") {
        404
    } else if code.starts_with("account_") || code.starts_with("task_") {
        409
    } else {
        500
    }
}

impl Serialize for RelayError {
    fn serialize<S: Serializer>(&self, serializer: S) -> Result<S::Ok, S::Error> {
        let mut s = serializer.serialize_struct("RelayError", 3)?;
        s.serialize_field("code", self.code)?;
        s.serialize_field("message", &self.message)?;
        s.serialize_field("details", &self.details)?;
        s.end()
    }
}

impl From<anyhow::Error> for RelayError {
    fn from(e: anyhow::Error) -> Self {
        RelayError::internal(format!("内部错误:{e:#}"))
    }
}

impl From<std::io::Error> for RelayError {
    fn from(e: std::io::Error) -> Self {
        RelayError::new("system_io_error", "文件读写失败,请检查磁盘与目录权限")
            .with_details(json!({ "io": e.to_string() }))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn http_status_follows_prefix_table() {
        assert_eq!(RelayError::new("request_malformed", "x").http_status(), 400);
        assert_eq!(
            RelayError::new("request_invalid_enum", "x").http_status(),
            400
        );
        assert_eq!(
            RelayError::new("auth_missing_token", "x").http_status(),
            401
        );
        assert_eq!(
            RelayError::new("auth_invalid_token", "x").http_status(),
            401
        );
        assert_eq!(
            RelayError::account_not_found("a", vec![]).http_status(),
            404
        );
        assert_eq!(RelayError::new("upload_not_found", "x").http_status(), 404);
        assert_eq!(RelayError::new("task_not_found", "x").http_status(), 404);
        assert_eq!(
            RelayError::new("result_not_available", "x").http_status(),
            404
        );
        assert_eq!(RelayError::new("account_busy", "x").http_status(), 409);
        assert_eq!(
            RelayError::new("task_not_cancellable", "x").http_status(),
            409
        );
        assert_eq!(
            RelayError::new("task_not_waiting_confirm", "x").http_status(),
            409
        );
        assert_eq!(RelayError::new("queue_full", "x").http_status(), 429);
        assert_eq!(RelayError::doubao_not_running().http_status(), 503);
        assert_eq!(RelayError::doubao_not_ready().http_status(), 503);
        assert_eq!(RelayError::doubao_cdp_disconnected().http_status(), 503);
        assert_eq!(RelayError::doubao_ui_changed(vec!["a"]).http_status(), 503);
        assert_eq!(RelayError::internal("x").http_status(), 500);
        assert_eq!(
            RelayError::new("system_database_error", "x").http_status(),
            500
        );
        assert_eq!(
            RelayError::new("platform_registry_error", "x").http_status(),
            500
        );
    }

    #[test]
    fn serializes_three_keys_without_details() {
        let v = serde_json::to_value(RelayError::doubao_not_running()).unwrap();
        let obj = v.as_object().unwrap();
        assert_eq!(obj.len(), 3);
        assert_eq!(obj["code"], "doubao_not_running");
        assert!(obj["message"].as_str().unwrap().contains("豆包未运行"));
        assert!(obj["details"].is_null());
    }

    #[test]
    fn serializes_three_keys_with_details() {
        let e = RelayError::account_not_found("张三", vec!["李四".into(), "王五".into()]);
        let v = serde_json::to_value(&e).unwrap();
        let obj = v.as_object().unwrap();
        assert_eq!(obj.len(), 3);
        assert_eq!(obj["code"], "account_not_found");
        assert_eq!(obj["details"]["requested"], "张三");
        assert_eq!(obj["details"]["available"], json!(["李四", "王五"]));
    }

    #[test]
    fn constructors_carry_expected_codes_and_details() {
        assert_eq!(
            RelayError::doubao_not_logged_in().code,
            "doubao_not_logged_in"
        );
        assert_eq!(
            RelayError::doubao_launch_timeout().code,
            "doubao_launch_timeout"
        );
        let e = RelayError::doubao_ui_changed(vec!["[role=menu]", "button"]);
        assert_eq!(e.code, "doubao_ui_changed");
        assert_eq!(
            e.details.unwrap()["failed_selectors"],
            json!(["[role=menu]", "button"])
        );
        let e = RelayError::account_ambiguous("张三", vec!["张三".into(), "张三".into()]);
        assert_eq!(e.code, "account_ambiguous");
        assert_eq!(e.details.unwrap()["matches"].as_array().unwrap().len(), 2);
        let e = RelayError::account_switch_failed("login_modal");
        assert_eq!(e.code, "account_switch_failed");
        assert_eq!(e.details.unwrap()["reason"], "login_modal");
        let e = RelayError::account_switch_timeout(20_000);
        assert_eq!(e.code, "account_switch_timeout");
        assert_eq!(e.details.unwrap()["elapsed_ms"], 20_000);
        let e = RelayError::internal("boom");
        assert_eq!(e.code, "system_internal_error");
        assert_eq!(e.message, "boom");
    }

    #[test]
    fn display_is_message_and_messages_have_no_trailing_punctuation() {
        let e = RelayError::doubao_not_ready();
        assert_eq!(e.to_string(), e.message);
        for e in [
            RelayError::doubao_not_running(),
            RelayError::doubao_not_ready(),
            RelayError::doubao_not_logged_in(),
            RelayError::doubao_cdp_disconnected(),
            RelayError::doubao_ui_changed(vec![]),
            RelayError::doubao_launch_timeout(),
            RelayError::account_not_found("a", vec![]),
            RelayError::account_ambiguous("a", vec![]),
            RelayError::account_switch_failed("r"),
            RelayError::account_switch_timeout(1),
        ] {
            let last = e.message.chars().last().unwrap();
            assert!(
                !matches!(last, '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'),
                "message ends with punctuation: {}",
                e.message
            );
        }
    }

    #[test]
    fn spec08_short_timeout_codes_alias_to_generation_codes() {
        assert_eq!(canonical_code("queue_timeout"), "generation_queue_timeout");
        assert_eq!(
            canonical_code("preparing_timeout"),
            "generation_preparing_timeout"
        );
        assert_eq!(canonical_code("task_timeout"), "generation_task_timeout");
        assert_eq!(
            canonical_code("confirm_timeout"),
            "generation_confirm_timeout"
        );
        // 其它码原样返回
        assert_eq!(canonical_code("doubao_crashed"), "doubao_crashed");
        assert_eq!(
            canonical_code("generation_task_timeout"),
            "generation_task_timeout"
        );
        // 别名与正式码的 HTTP 状态一致
        assert_eq!(
            http_status_for_code("task_timeout"),
            http_status_for_code("generation_task_timeout")
        );
    }

    #[test]
    fn doubao_crashed_carries_last_seen_at_and_is_503() {
        let e = RelayError::doubao_crashed(1_700_000_000_000);
        assert_eq!(e.code, "doubao_crashed");
        assert_eq!(e.http_status(), 503);
        assert_eq!(e.details.unwrap()["last_seen_at"], 1_700_000_000_000i64);
        let e = RelayError::doubao_crashed(0);
        assert!(e.details.unwrap()["last_seen_at"].is_null());
        assert!(!e.message.ends_with('。'));
        assert_eq!(e.message, DOUBAO_CRASHED_MESSAGE);
    }

    #[test]
    fn from_io_error_maps_to_system_io_error() {
        let io = std::io::Error::new(std::io::ErrorKind::NotFound, "nope");
        let e: RelayError = io.into();
        assert_eq!(e.code, "system_io_error");
        assert_eq!(e.http_status(), 500);
    }
}
