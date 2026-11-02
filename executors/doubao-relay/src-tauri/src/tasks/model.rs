//! 任务模型(`docs/spec/07-data-model.md` + `04-http-api.md`)。
//!
//! domain 类型统一 `snake_case` 序列化,HTTP 层直接返回;IPC 层经 `to_camel` 转换。
//! 时间戳一律 Unix epoch 毫秒 `i64`;ID 一律 ULID(26 字符);豆包侧 ID 一律 `String`。

use serde::{Deserialize, Serialize};
use serde_json::Value;

/// 任务类型。
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum TaskType {
    Image,
    Video,
}

/// 多轮确认模式(F4.5)。
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ConfirmationMode {
    #[default]
    Auto,
    Manual,
}

/// 任务状态(07-data-model `TaskStatus`)。
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum TaskStatus {
    Pending,
    Preparing,
    Generating,
    WaitingUserConfirm,
    Completed,
    Failed,
    Cancelled,
}

/// generating 下的细分阶段(07-data-model `TaskStage`)。
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum TaskStage {
    Preparing,
    Sent,
    Checking,
    QuotaChecking,
    Rendering,
    RenderingUnverified,
    Downloading,
    Result,
    Refused,
}

/// 结果文件的无水印来源(07-data-model `WatermarkSource`)。
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum WatermarkSource {
    ConversationOriginal,
    LrUnwatermarked,
    FallbackDecrypted,
}

fn default_true() -> bool {
    true
}

/// `POST /generate` 请求体(04-http-api)。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct GenerateRequest {
    pub account: String,
    #[serde(rename = "type")]
    pub task_type: TaskType,
    pub model_name: String,
    pub prompt: String,
    pub ratio: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub duration: Option<u32>,
    #[serde(default)]
    pub reference_files: Vec<String>,
    #[serde(default)]
    pub confirmation_mode: ConfirmationMode,
    #[serde(default = "default_true")]
    pub auto_use_membership: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub callback_url: Option<String>,
}

/// 结果文件(07-data-model "result_files JSON 格式")。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ResultFile {
    pub index: u32,
    pub filename: String,
    pub relative_path: String,
    pub mime: String,
    pub bytes: u64,
    pub width: Option<u32>,
    pub height: Option<u32>,
    pub duration_ms: Option<u64>,
    /// 捕获到的豆包 CDN 原始地址(带签名,会过期);旧数据没有此字段。
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub source_url: Option<String>,
}

/// 任务失败信息(与 `RelayError` 同形,但 code 是持久化的 `String`)。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TaskError {
    pub code: String,
    pub message: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub details: Option<Value>,
}

/// 阶段历史条目(`task_stage_history` 一行)。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct StageEntry {
    pub stage: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub message: Option<String>,
    pub ts: i64,
}

/// 任务完整态(`tasks` 表一行 + 阶段历史 + 队列位置)。
///
/// 序列化时 `id` 输出为 `task_id`(04-http-api / 05-tauri-ipc 都叫 `task_id` / `taskId`)。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Task {
    #[serde(rename = "task_id")]
    pub id: String,
    pub request: GenerateRequest,
    pub status: TaskStatus,
    pub stage: Option<TaskStage>,
    pub stage_message: Option<String>,
    pub progress: Option<f32>,
    pub error: Option<TaskError>,
    pub conversation_id: Option<String>,
    pub submission_id: Option<String>,
    pub local_message_id: Option<String>,
    pub watermark_source: Option<WatermarkSource>,
    pub result_files: Option<Vec<ResultFile>>,
    pub created_at: i64,
    pub started_at: Option<i64>,
    pub completed_at: Option<i64>,
    pub expires_at: i64,
    #[serde(default)]
    pub stage_history: Vec<StageEntry>,
    pub queue_position: Option<u32>,
}

/// 任务列表项(`GET /tasks` 一行)。
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TaskSummary {
    #[serde(rename = "task_id")]
    pub id: String,
    pub account: String,
    #[serde(rename = "type")]
    pub task_type: TaskType,
    pub status: TaskStatus,
    pub stage: Option<TaskStage>,
    pub created_at: i64,
    pub completed_at: Option<i64>,
}

impl From<&Task> for TaskSummary {
    fn from(t: &Task) -> Self {
        Self {
            id: t.id.clone(),
            account: t.request.account.clone(),
            task_type: t.request.task_type,
            status: t.status,
            stage: t.stage,
            created_at: t.created_at,
            completed_at: t.completed_at,
        }
    }
}

impl TaskType {
    pub fn as_str(&self) -> &'static str {
        match self {
            TaskType::Image => "image",
            TaskType::Video => "video",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "image" => Some(TaskType::Image),
            "video" => Some(TaskType::Video),
            _ => None,
        }
    }
}

impl ConfirmationMode {
    pub fn as_str(&self) -> &'static str {
        match self {
            ConfirmationMode::Auto => "auto",
            ConfirmationMode::Manual => "manual",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "auto" => Some(ConfirmationMode::Auto),
            "manual" => Some(ConfirmationMode::Manual),
            _ => None,
        }
    }
}

impl TaskStatus {
    /// completed / failed / cancelled 为终态。
    pub fn is_terminal(&self) -> bool {
        matches!(
            self,
            TaskStatus::Completed | TaskStatus::Failed | TaskStatus::Cancelled
        )
    }

    pub fn as_str(&self) -> &'static str {
        match self {
            TaskStatus::Pending => "pending",
            TaskStatus::Preparing => "preparing",
            TaskStatus::Generating => "generating",
            TaskStatus::WaitingUserConfirm => "waiting_user_confirm",
            TaskStatus::Completed => "completed",
            TaskStatus::Failed => "failed",
            TaskStatus::Cancelled => "cancelled",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "pending" => Some(TaskStatus::Pending),
            "preparing" => Some(TaskStatus::Preparing),
            "generating" => Some(TaskStatus::Generating),
            "waiting_user_confirm" => Some(TaskStatus::WaitingUserConfirm),
            "completed" => Some(TaskStatus::Completed),
            "failed" => Some(TaskStatus::Failed),
            "cancelled" => Some(TaskStatus::Cancelled),
            _ => None,
        }
    }
}

impl TaskStage {
    pub fn as_str(&self) -> &'static str {
        match self {
            TaskStage::Preparing => "preparing",
            TaskStage::Sent => "sent",
            TaskStage::Checking => "checking",
            TaskStage::QuotaChecking => "quota_checking",
            TaskStage::Rendering => "rendering",
            TaskStage::RenderingUnverified => "rendering_unverified",
            TaskStage::Downloading => "downloading",
            TaskStage::Result => "result",
            TaskStage::Refused => "refused",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "preparing" => Some(TaskStage::Preparing),
            "sent" => Some(TaskStage::Sent),
            "checking" => Some(TaskStage::Checking),
            "quota_checking" => Some(TaskStage::QuotaChecking),
            "rendering" => Some(TaskStage::Rendering),
            "rendering_unverified" => Some(TaskStage::RenderingUnverified),
            "downloading" => Some(TaskStage::Downloading),
            "result" => Some(TaskStage::Result),
            "refused" => Some(TaskStage::Refused),
            _ => None,
        }
    }
}

impl WatermarkSource {
    pub fn as_str(&self) -> &'static str {
        match self {
            WatermarkSource::ConversationOriginal => "conversation_original",
            WatermarkSource::LrUnwatermarked => "lr_unwatermarked",
            WatermarkSource::FallbackDecrypted => "fallback_decrypted",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "conversation_original" => Some(WatermarkSource::ConversationOriginal),
            "lr_unwatermarked" => Some(WatermarkSource::LrUnwatermarked),
            "fallback_decrypted" => Some(WatermarkSource::FallbackDecrypted),
            _ => None,
        }
    }
}

/// 生成新任务 ID(ULID,26 字符,Crockford base32 大写)。
pub fn new_task_id() -> String {
    ulid::Ulid::new().to_string()
}

/// prompt 的 sha256 hex 前 16 位,只用于日志(日志里不得出现 prompt 原文)。
pub fn prompt_hash(prompt: &str) -> String {
    use sha2::{Digest, Sha256};
    let digest = Sha256::digest(prompt.as_bytes());
    hex::encode(digest)[..16].to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn sample_request() -> GenerateRequest {
        GenerateRequest {
            account: "张三".into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "一只蓝色小鸟".into(),
            ratio: "16:9".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    fn sample_task() -> Task {
        Task {
            id: "01HQZ9ABCDEFGHJKMNPQRSTVWX".into(),
            request: sample_request(),
            status: TaskStatus::Generating,
            stage: Some(TaskStage::QuotaChecking),
            stage_message: Some("豆包确认是否使用会员配额".into()),
            progress: None,
            error: None,
            conversation_id: Some("12345678901234567".into()),
            submission_id: None,
            local_message_id: Some("98765432109876543".into()),
            watermark_source: None,
            result_files: None,
            created_at: 1_697_433_600_000,
            started_at: Some(1_697_433_610_000),
            completed_at: None,
            expires_at: 1_697_520_000_000,
            stage_history: vec![
                StageEntry {
                    stage: "pending".into(),
                    message: None,
                    ts: 1_697_433_600_000,
                },
                StageEntry {
                    stage: "quota_checking".into(),
                    message: Some("豆包确认是否使用会员配额".into()),
                    ts: 1_697_433_625_000,
                },
            ],
            queue_position: None,
        }
    }

    #[test]
    fn generate_request_deserializes_with_defaults() {
        let r: GenerateRequest = serde_json::from_value(json!({
            "account": "张三",
            "type": "image",
            "model_name": "Seedream 5.0 Pro",
            "prompt": "一只蓝色小鸟",
            "ratio": "16:9"
        }))
        .unwrap();
        assert_eq!(r, sample_request());
        assert!(r.reference_files.is_empty());
        assert_eq!(r.confirmation_mode, ConfirmationMode::Auto);
        assert!(r.auto_use_membership);
        assert_eq!(r.duration, None);
        assert_eq!(r.callback_url, None);
    }

    #[test]
    fn generate_request_video_round_trips_snake_case() {
        let v = json!({
            "account": "张三",
            "type": "video",
            "model_name": "Seedance 2.5",
            "prompt": "p",
            "ratio": "9:16",
            "duration": 10,
            "reference_files": ["01HQZ9", "01HQZ8"],
            "confirmation_mode": "manual",
            "auto_use_membership": false,
            "callback_url": "https://my-service/doubao-webhook"
        });
        let r: GenerateRequest = serde_json::from_value(v.clone()).unwrap();
        assert_eq!(r.task_type, TaskType::Video);
        assert_eq!(r.duration, Some(10));
        assert_eq!(r.confirmation_mode, ConfirmationMode::Manual);
        assert!(!r.auto_use_membership);
        assert_eq!(serde_json::to_value(&r).unwrap(), v);
    }

    #[test]
    fn generate_request_rejects_unknown_type_and_mode() {
        assert!(serde_json::from_value::<GenerateRequest>(json!({
            "account": "a", "type": "audio", "model_name": "m", "prompt": "p", "ratio": "1:1"
        }))
        .is_err());
        assert!(serde_json::from_value::<GenerateRequest>(json!({
            "account": "a", "type": "image", "model_name": "m", "prompt": "p", "ratio": "1:1",
            "confirmation_mode": "sometimes"
        }))
        .is_err());
        // 缺必填字段由 serde 拒绝(HTTP 层映射为 request_malformed / request_missing_field)。
        assert!(serde_json::from_value::<GenerateRequest>(json!({
            "type": "image", "model_name": "m", "prompt": "p", "ratio": "1:1"
        }))
        .is_err());
    }

    #[test]
    fn task_status_as_str_parse_round_trip_and_terminal() {
        let all = [
            (TaskStatus::Pending, "pending", false),
            (TaskStatus::Preparing, "preparing", false),
            (TaskStatus::Generating, "generating", false),
            (
                TaskStatus::WaitingUserConfirm,
                "waiting_user_confirm",
                false,
            ),
            (TaskStatus::Completed, "completed", true),
            (TaskStatus::Failed, "failed", true),
            (TaskStatus::Cancelled, "cancelled", true),
        ];
        for (s, text, terminal) in all {
            assert_eq!(s.as_str(), text);
            assert_eq!(TaskStatus::parse(text), Some(s));
            assert_eq!(s.is_terminal(), terminal, "{text}");
            assert_eq!(serde_json::to_value(s).unwrap(), json!(text));
        }
        assert_eq!(TaskStatus::parse("done"), None);
        assert_eq!(TaskStatus::parse("Completed"), None);
    }

    #[test]
    fn task_stage_as_str_parse_round_trip() {
        let all = [
            (TaskStage::Preparing, "preparing"),
            (TaskStage::Sent, "sent"),
            (TaskStage::Checking, "checking"),
            (TaskStage::QuotaChecking, "quota_checking"),
            (TaskStage::Rendering, "rendering"),
            (TaskStage::RenderingUnverified, "rendering_unverified"),
            (TaskStage::Downloading, "downloading"),
            (TaskStage::Result, "result"),
            (TaskStage::Refused, "refused"),
        ];
        for (s, text) in all {
            assert_eq!(s.as_str(), text);
            assert_eq!(TaskStage::parse(text), Some(s));
            assert_eq!(serde_json::to_value(s).unwrap(), json!(text));
        }
        assert_eq!(TaskStage::parse("unknown"), None);
    }

    #[test]
    fn task_type_and_watermark_source_as_str_parse() {
        assert_eq!(TaskType::Image.as_str(), "image");
        assert_eq!(TaskType::parse("video"), Some(TaskType::Video));
        assert_eq!(TaskType::parse("gif"), None);
        for (w, text) in [
            (
                WatermarkSource::ConversationOriginal,
                "conversation_original",
            ),
            (WatermarkSource::LrUnwatermarked, "lr_unwatermarked"),
            (WatermarkSource::FallbackDecrypted, "fallback_decrypted"),
        ] {
            assert_eq!(w.as_str(), text);
            assert_eq!(WatermarkSource::parse(text), Some(w));
            assert_eq!(serde_json::to_value(w).unwrap(), json!(text));
        }
        assert_eq!(WatermarkSource::parse("x"), None);
        assert_eq!(ConfirmationMode::default(), ConfirmationMode::Auto);
    }

    #[test]
    fn task_serializes_snake_case_with_task_id() {
        let v = serde_json::to_value(sample_task()).unwrap();
        assert_eq!(v["task_id"], "01HQZ9ABCDEFGHJKMNPQRSTVWX");
        assert!(v.get("id").is_none());
        assert_eq!(v["status"], "generating");
        assert_eq!(v["stage"], "quota_checking");
        assert_eq!(v["stage_message"], "豆包确认是否使用会员配额");
        assert_eq!(v["request"]["type"], "image");
        assert_eq!(v["request"]["model_name"], "Seedream 5.0 Pro");
        assert_eq!(v["conversation_id"], "12345678901234567");
        assert_eq!(v["stage_history"][1]["message"], "豆包确认是否使用会员配额");
        assert!(v["stage_history"][0].get("message").is_none());
        assert!(v["progress"].is_null());
        assert!(v["error"].is_null());
        let back: Task = serde_json::from_value(v).unwrap();
        assert_eq!(back, sample_task());
    }

    #[test]
    fn task_error_and_result_file_serialize() {
        let e = TaskError {
            code: "generation_refused".into(),
            message: "豆包拒绝生成,请调整提示词后重试".into(),
            details: Some(json!({ "reason": "policy" })),
        };
        let v = serde_json::to_value(&e).unwrap();
        assert_eq!(v["code"], "generation_refused");
        assert_eq!(v["details"]["reason"], "policy");
        let f = ResultFile {
            index: 0,
            filename: "image_1.png".into(),
            relative_path: "results/01HQZ9/image_1.png".into(),
            mime: "image/png".into(),
            bytes: 2_456_321,
            width: Some(1920),
            height: Some(1080),
            duration_ms: None,
            source_url: None,
        };
        let v = serde_json::to_value(&f).unwrap();
        assert_eq!(v["relative_path"], "results/01HQZ9/image_1.png");
        assert!(v["duration_ms"].is_null());
        assert_eq!(serde_json::from_value::<ResultFile>(v).unwrap(), f);
    }

    #[test]
    fn task_summary_from_task() {
        let t = sample_task();
        let s = TaskSummary::from(&t);
        assert_eq!(s.id, t.id);
        assert_eq!(s.account, "张三");
        assert_eq!(s.task_type, TaskType::Image);
        assert_eq!(s.status, TaskStatus::Generating);
        assert_eq!(s.stage, Some(TaskStage::QuotaChecking));
        assert_eq!(s.created_at, t.created_at);
        assert_eq!(s.completed_at, None);
        let v = serde_json::to_value(&s).unwrap();
        assert_eq!(v["task_id"], t.id);
        assert_eq!(v["type"], "image");
        assert_eq!(v["stage"], "quota_checking");
        assert_eq!(v.as_object().unwrap().len(), 7);
    }

    #[test]
    fn new_task_id_is_26_char_ulid_and_unique() {
        let a = new_task_id();
        let b = new_task_id();
        assert_eq!(a.len(), 26);
        assert!(a
            .chars()
            .all(|c| c.is_ascii_digit() || c.is_ascii_uppercase()));
        assert_ne!(a, b);
        assert!(ulid::Ulid::from_string(&a).is_ok());
    }

    #[test]
    fn prompt_hash_is_16_hex_and_deterministic() {
        let h = prompt_hash("一只蓝色小鸟");
        assert_eq!(h.len(), 16);
        assert!(h
            .chars()
            .all(|c| c.is_ascii_hexdigit() && !c.is_ascii_uppercase()));
        assert_eq!(h, prompt_hash("一只蓝色小鸟"));
        assert_ne!(h, prompt_hash("一只红色小鸟"));
        // sha256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
        assert_eq!(prompt_hash("abc"), "ba7816bf8f01cfea");
    }
}
