//! feedback 阶段映射(Task 2.4):把 `doubao-generation-page.js` / `doubao-result-reader.js`
//! 返回的 feedback `stage` 字符串映射为 [`TaskStage`] / 错误码,并给出自动确认的真值表。
//!
//! JS 侧 7 类 stage(analysis/05 "feedback 系统"):`unknown` / `checking` / `quota_checking` /
//! `generating` / `result` / `needs_check` / `rendering_unverified`,终态拒绝 `refused` / `failed` /
//! `quota_exhausted` / `not_executed`。
//!
//! 真机(Node 里对混淆脚本实测):`run('terminalFeedback', req)` 只读 `req.type` / `req.prompt`,
//! 返回 `{stage, message?}`;`run('videoFollowup', req)` 返回 `{sent, stage, manualRequired?}`。
//! `run('acceptFeedback', req)` **不发送任何消息**(它把 `req.reply` 当 native-shadow 的反馈消息对象做
//! 分类,字符串直接得 `{stage:"unknown"}`);图像代回复由 `capture::accept_reply_expression`
//! 自己 fill + 点发送,回复文本 [`CONTINUE_REPLY`]。

use std::time::Duration;

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

use crate::error::RelayError;
use crate::tasks::{ConfirmationMode, TaskError, TaskStage, TaskType};

/// `rendering_unverified` 持续超过此时长仍未进入 generating/result → `not_executed`(analysis/05)。
pub const RENDERING_UNVERIFIED_LIMIT: Duration = Duration::from_secs(30);

/// 代回复"继续生成"的文本(reader 的 `collectFeedback` 按 `normalize(text) === '继续生成'` 认领)。
pub const CONTINUE_REPLY: &str = "继续生成";

/// JS 侧 feedback 的 stage 取值。
pub const STAGE_UNKNOWN: &str = "unknown";
pub const STAGE_CHECKING: &str = "checking";
pub const STAGE_QUOTA_CHECKING: &str = "quota_checking";
pub const STAGE_GENERATING: &str = "generating";
pub const STAGE_RESULT: &str = "result";
pub const STAGE_NEEDS_CHECK: &str = "needs_check";
pub const STAGE_RENDERING_UNVERIFIED: &str = "rendering_unverified";
pub const STAGE_REFUSED: &str = "refused";
pub const STAGE_FAILED: &str = "failed";
pub const STAGE_QUOTA_EXHAUSTED: &str = "quota_exhausted";
pub const STAGE_NOT_EXECUTED: &str = "not_executed";

/// 一次 feedback 轮询 / 历史重放得到的状态。
#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct FeedbackState {
    /// JS 侧 stage 原文(`unknown` 表示未识别)。
    pub stage: String,
    pub message: Option<String>,
    /// `videoFollowup` 已代回复"继续生成"。
    pub sent: bool,
    /// `videoFollowup` 在 auto 模式下找不到可自动确认的提示(`manualRequired:true`),需要用户介入。
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub manual_required: bool,
    /// 由 relay 自己读页面识别出的确认提示(沈浪脚本没认出):代回复走 relay 的发送路径。
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub relay_detected: bool,
}

impl FeedbackState {
    pub fn unknown() -> Self {
        Self {
            stage: STAGE_UNKNOWN.to_string(),
            message: None,
            sent: false,
            manual_required: false,
            relay_detected: false,
        }
    }

    pub fn is_unknown(&self) -> bool {
        self.stage.is_empty() || self.stage == STAGE_UNKNOWN
    }
}

/// 解析 JS 返回的 feedback 对象:`null` / 非对象 / 缺 `stage` → `unknown`;空 `message` → None。
pub fn parse_feedback(v: &Value) -> FeedbackState {
    let Some(obj) = v.as_object() else {
        return FeedbackState::unknown();
    };
    let stage = obj
        .get("stage")
        .and_then(Value::as_str)
        .filter(|s| !s.is_empty())
        .unwrap_or(STAGE_UNKNOWN)
        .to_string();
    let message = obj
        .get("message")
        .and_then(Value::as_str)
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(str::to_string);
    let sent = obj.get("sent").and_then(Value::as_bool).unwrap_or(false);
    let manual_required = obj
        .get("manualRequired")
        .and_then(Value::as_bool)
        .unwrap_or(false);
    FeedbackState {
        stage,
        message,
        sent,
        manual_required,
        relay_detected: false,
    }
}

/// JS stage → `TaskStage`:checking→Checking、quota_checking→QuotaChecking、generating→Rendering、
/// rendering_unverified→RenderingUnverified、result→Result、refused|failed|quota_exhausted|not_executed→Refused;
/// `unknown` / `needs_check` / 其它 → None。
pub fn map_feedback_stage(js_stage: &str) -> Option<TaskStage> {
    match js_stage {
        STAGE_CHECKING => Some(TaskStage::Checking),
        STAGE_QUOTA_CHECKING => Some(TaskStage::QuotaChecking),
        STAGE_GENERATING => Some(TaskStage::Rendering),
        STAGE_RENDERING_UNVERIFIED => Some(TaskStage::RenderingUnverified),
        STAGE_RESULT => Some(TaskStage::Result),
        STAGE_REFUSED | STAGE_FAILED | STAGE_QUOTA_EXHAUSTED | STAGE_NOT_EXECUTED => {
            Some(TaskStage::Refused)
        }
        _ => None,
    }
}

/// 终态拒绝 stage。
pub fn is_refusal_stage(js_stage: &str) -> bool {
    matches!(
        js_stage,
        STAGE_REFUSED | STAGE_FAILED | STAGE_QUOTA_EXHAUSTED | STAGE_NOT_EXECUTED
    )
}

/// 需要确认的 stage(checking / quota_checking)。
pub fn is_confirmation_stage(js_stage: &str) -> bool {
    matches!(js_stage, STAGE_CHECKING | STAGE_QUOTA_CHECKING)
}

/// 拒绝类 stage → `RelayError`:quota_exhausted→`account_quota_exhausted`、
/// not_executed→`generation_rendering_not_executed`、其它→`generation_doubao_refused{reason}`。
/// `details.doubao_message` 带豆包原话。
pub fn refusal_relay_error(js_stage: &str, message: &str) -> RelayError {
    let doubao_message = if message.trim().is_empty() {
        Value::Null
    } else {
        Value::String(message.trim().to_string())
    };
    match js_stage {
        STAGE_QUOTA_EXHAUSTED => RelayError::new(
            "account_quota_exhausted",
            "账号配额已用完,请更换账号或等待配额恢复后重试",
        )
        .with_details(json!({ "reason": js_stage, "doubao_message": doubao_message })),
        STAGE_NOT_EXECUTED => not_executed_error(message),
        _ => RelayError::new(
            "generation_doubao_refused",
            "豆包拒绝了本次生成请求,请调整提示词或参考图后重试",
        )
        .with_details(json!({ "reason": js_stage, "doubao_message": doubao_message })),
    }
}

/// `generation_rendering_not_executed`(豆包只回"正在渲染",30 秒未开始)。
pub fn not_executed_error(message: &str) -> RelayError {
    let doubao_message = if message.trim().is_empty() {
        Value::Null
    } else {
        Value::String(message.trim().to_string())
    };
    RelayError::new(
        "generation_rendering_not_executed",
        "豆包仅返回正在渲染但 30 秒内未开始生成,请核对账号权限后重试",
    )
    .with_details(json!({
        "reason": STAGE_NOT_EXECUTED,
        "limit_ms": RENDERING_UNVERIFIED_LIMIT.as_millis() as u64,
        "doubao_message": doubao_message,
    }))
}

/// 同 [`refusal_relay_error`],但返回持久化形态的 [`TaskError`]。
pub fn refusal_error(js_stage: &str, message: &str) -> TaskError {
    let e = refusal_relay_error(js_stage, message);
    TaskError {
        code: e.code.to_string(),
        message: e.message,
        details: e.details,
    }
}

/// 是否自动代回复:`Checking` 看 `confirmation_mode == Auto`,`QuotaChecking` 看
/// `auto_use_membership`,其它 stage 永远 false(F4.5)。
pub fn should_auto_confirm(
    stage: &TaskStage,
    mode: ConfirmationMode,
    auto_membership: bool,
) -> bool {
    match stage {
        TaskStage::Checking => mode == ConfirmationMode::Auto,
        TaskStage::QuotaChecking => auto_membership,
        _ => false,
    }
}

/// 每轮轮询用哪个 `run` 步骤(analysis/05 "自动代回复"):图像 `terminalFeedback`(只识别);
/// 视频在已确认 / auto / 自动用会员时 `videoFollowup`(识别到提示即自行代回复),否则 `videoFeedback`。
pub fn feedback_step(
    task_type: TaskType,
    mode: ConfirmationMode,
    auto_membership: bool,
    user_confirmed: bool,
) -> &'static str {
    match task_type {
        TaskType::Image => "terminalFeedback",
        TaskType::Video => {
            if user_confirmed || mode == ConfirmationMode::Auto || auto_membership {
                "videoFollowup"
            } else {
                "videoFeedback"
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn map_feedback_stage_table() {
        assert_eq!(map_feedback_stage("checking"), Some(TaskStage::Checking));
        assert_eq!(
            map_feedback_stage("quota_checking"),
            Some(TaskStage::QuotaChecking)
        );
        assert_eq!(map_feedback_stage("generating"), Some(TaskStage::Rendering));
        assert_eq!(
            map_feedback_stage("rendering_unverified"),
            Some(TaskStage::RenderingUnverified)
        );
        assert_eq!(map_feedback_stage("result"), Some(TaskStage::Result));
        for s in ["refused", "failed", "quota_exhausted", "not_executed"] {
            assert_eq!(map_feedback_stage(s), Some(TaskStage::Refused), "{s}");
            assert!(is_refusal_stage(s), "{s}");
        }
        assert_eq!(map_feedback_stage("unknown"), None);
        assert_eq!(map_feedback_stage("needs_check"), None);
        assert_eq!(map_feedback_stage(""), None);
        assert_eq!(map_feedback_stage("Checking"), None);
        assert!(!is_refusal_stage("checking"));
        assert!(!is_refusal_stage("unknown"));
        assert!(is_confirmation_stage("checking"));
        assert!(is_confirmation_stage("quota_checking"));
        assert!(!is_confirmation_stage("generating"));
    }

    #[test]
    fn refusal_error_maps_codes_and_keeps_doubao_message() {
        let e = refusal_error("quota_exhausted", "额度不足");
        assert_eq!(e.code, "account_quota_exhausted");
        let d = e.details.unwrap();
        assert_eq!(d["reason"], "quota_exhausted");
        assert_eq!(d["doubao_message"], "额度不足");

        let e = refusal_error("not_executed", "");
        assert_eq!(e.code, "generation_rendering_not_executed");
        assert!(e.details.unwrap()["doubao_message"].is_null());

        for s in ["refused", "failed", "whatever"] {
            let e = refusal_error(s, "不能生成");
            assert_eq!(e.code, "generation_doubao_refused", "{s}");
            let d = e.details.unwrap();
            assert_eq!(d["reason"], s);
            assert_eq!(d["doubao_message"], "不能生成");
        }

        // RelayError 形态:HTTP 状态按前缀表;message 不带句末标点
        for e in [
            refusal_relay_error("quota_exhausted", "x"),
            refusal_relay_error("not_executed", "x"),
            refusal_relay_error("refused", "x"),
            not_executed_error(""),
        ] {
            let last = e.message.chars().last().unwrap();
            assert!(
                !matches!(last, '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'),
                "{}",
                e.message
            );
        }
        assert_eq!(
            refusal_relay_error("quota_exhausted", "x").http_status(),
            409
        );
        assert_eq!(refusal_relay_error("refused", "x").http_status(), 500);
    }

    #[test]
    fn should_auto_confirm_truth_table() {
        use ConfirmationMode::{Auto, Manual};
        let c = TaskStage::Checking;
        let q = TaskStage::QuotaChecking;
        assert!(should_auto_confirm(&c, Auto, true));
        assert!(should_auto_confirm(&c, Auto, false));
        assert!(!should_auto_confirm(&c, Manual, true));
        assert!(!should_auto_confirm(&c, Manual, false));
        assert!(should_auto_confirm(&q, Auto, true));
        assert!(should_auto_confirm(&q, Manual, true));
        assert!(!should_auto_confirm(&q, Auto, false));
        assert!(!should_auto_confirm(&q, Manual, false));
        for s in [
            TaskStage::Preparing,
            TaskStage::Sent,
            TaskStage::Rendering,
            TaskStage::RenderingUnverified,
            TaskStage::Downloading,
            TaskStage::Result,
            TaskStage::Refused,
        ] {
            assert!(!should_auto_confirm(&s, Auto, true), "{s:?}");
        }
    }

    #[test]
    fn feedback_step_follows_capture_ps_rule() {
        use ConfirmationMode::{Auto, Manual};
        assert_eq!(
            feedback_step(TaskType::Image, Manual, false, false),
            "terminalFeedback"
        );
        assert_eq!(
            feedback_step(TaskType::Image, Auto, true, true),
            "terminalFeedback"
        );
        assert_eq!(
            feedback_step(TaskType::Video, Manual, false, false),
            "videoFeedback"
        );
        assert_eq!(
            feedback_step(TaskType::Video, Auto, false, false),
            "videoFollowup"
        );
        assert_eq!(
            feedback_step(TaskType::Video, Manual, true, false),
            "videoFollowup"
        );
        assert_eq!(
            feedback_step(TaskType::Video, Manual, false, true),
            "videoFollowup"
        );
    }

    #[test]
    fn parse_feedback_shapes() {
        assert_eq!(parse_feedback(&Value::Null), FeedbackState::unknown());
        assert_eq!(parse_feedback(&json!("x")), FeedbackState::unknown());
        assert_eq!(parse_feedback(&json!({})).stage, "unknown");
        let f = parse_feedback(&json!({ "stage": "checking", "message": " 是否继续生成? " }));
        assert_eq!(f.stage, "checking");
        assert_eq!(f.message.as_deref(), Some("是否继续生成?"));
        assert!(!f.sent);
        let f = parse_feedback(&json!({ "stage": "generating", "message": "", "sent": true }));
        assert_eq!(f.message, None);
        assert!(f.sent);
        assert!(!f.manual_required);
        assert!(parse_feedback(&json!({ "stage": "" })).is_unknown());
        assert!(!f.is_unknown());
        let f = parse_feedback(
            &json!({ "stage": "checking", "sent": false, "manualRequired": true, "message": "m" }),
        );
        assert!(f.manual_required);
        // manual_required 为 false 时不序列化(IPC / 持久化形状不变)
        let v = serde_json::to_value(parse_feedback(&json!({ "stage": "checking" }))).unwrap();
        assert!(v.get("manual_required").is_none());
        let v = serde_json::to_value(&f).unwrap();
        assert_eq!(v["manual_required"], true);
    }

    #[test]
    fn constants_follow_spec() {
        assert_eq!(RENDERING_UNVERIFIED_LIMIT, Duration::from_secs(30));
        assert_eq!(CONTINUE_REPLY, "继续生成");
    }
}
