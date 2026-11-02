//! 任务详情体(`docs/spec/04-http-api.md` `GET /tasks/{id}`)。
//!
//! 详情体是 `Task` 的序列化(`task_id` / `request` / `stage_message` / `result_files` …)再叠加
//! spec 04 命名的字段:`account` / `type`(来自 request)、`message`(= `stage_message`)、
//! 终态 `completed_at`、失败另给 `failed_at`、完成另给 `result{files[{…, download_url}], watermark_source}`。
//! 两套字段同时存在。
//!
//! 放在任务域而不是 http 层:HTTP `GET /tasks/{id}`、IPC `task_get`(Task 2.9 经 `to_camel`)
//! 与 Webhook 回调 body 的 `task` 字段(spec 04 "同 GET /tasks/{id} 的响应体")共用同一函数。

use serde_json::{json, Value};

use crate::tasks::{Task, TaskStatus};

/// 所有 HTTP 端点的前缀(`http::API_PREFIX` 即此常量)。
pub const API_PREFIX: &str = "/api/v1";

/// 结果文件的下载路径(`/api/v1/tasks/{id}/download/{index}`)。
pub fn download_url(task_id: &str, index: u32) -> String {
    format!("{API_PREFIX}/tasks/{task_id}/download/{index}")
}

/// 任务详情体(见模块文档)。
pub fn task_detail_json(task: &Task) -> Value {
    let mut v = serde_json::to_value(task).unwrap_or(Value::Null);
    let Some(obj) = v.as_object_mut() else {
        return v;
    };
    obj.insert("account".into(), json!(task.request.account));
    obj.insert("type".into(), json!(task.request.task_type));
    obj.insert("message".into(), json!(task.stage_message));
    if task.status == TaskStatus::Failed {
        obj.insert("failed_at".into(), json!(task.completed_at));
    }
    if let Some(files) = &task.result_files {
        let files: Vec<Value> = files
            .iter()
            .map(|f| {
                json!({
                    "index": f.index,
                    "filename": f.filename,
                    "mime": f.mime,
                    "bytes": f.bytes,
                    "width": f.width,
                    "height": f.height,
                    "duration_ms": f.duration_ms,
                    "download_url": download_url(&task.id, f.index),
                    "source_url": f.source_url,
                })
            })
            .collect();
        obj.insert(
            "result".into(),
            json!({ "files": files, "watermark_source": task.watermark_source }),
        );
    }
    v
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::tasks::{
        ConfirmationMode, GenerateRequest, ResultFile, StageEntry, TaskError, TaskType,
        WatermarkSource,
    };

    fn task(status: TaskStatus) -> Task {
        Task {
            id: "01HQZ9ABCDEFGHJKMNPQRSTVWX".into(),
            request: GenerateRequest {
                account: "张三".into(),
                task_type: TaskType::Image,
                model_name: "Seedream 5.0 Pro".into(),
                prompt: "p".into(),
                ratio: "1:1".into(),
                duration: None,
                reference_files: vec![],
                confirmation_mode: ConfirmationMode::Auto,
                auto_use_membership: true,
                callback_url: None,
            },
            status,
            stage: None,
            stage_message: Some("豆包确认参数".into()),
            progress: None,
            error: None,
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: None,
            result_files: None,
            created_at: 1,
            started_at: Some(2),
            completed_at: Some(3),
            expires_at: 4,
            stage_history: vec![StageEntry {
                stage: "pending".into(),
                message: None,
                ts: 1,
            }],
            queue_position: None,
        }
    }

    #[test]
    fn detail_json_adds_spec_fields_for_running_task() {
        let v = task_detail_json(&task(TaskStatus::Generating));
        assert_eq!(v["task_id"], "01HQZ9ABCDEFGHJKMNPQRSTVWX");
        assert_eq!(v["account"], "张三");
        assert_eq!(v["type"], "image");
        assert_eq!(v["status"], "generating");
        assert_eq!(v["message"], "豆包确认参数");
        assert_eq!(v["request"]["model_name"], "Seedream 5.0 Pro");
        assert_eq!(v["stage_history"][0]["stage"], "pending");
        assert!(v.get("result").is_none());
        assert!(v.get("failed_at").is_none());
    }

    #[test]
    fn detail_json_completed_has_result_with_download_urls() {
        let mut t = task(TaskStatus::Completed);
        t.watermark_source = Some(WatermarkSource::ConversationOriginal);
        t.result_files = Some(vec![ResultFile {
            index: 0,
            filename: "image_1.png".into(),
            relative_path: "results/01HQZ9ABCDEFGHJKMNPQRSTVWX/image_1.png".into(),
            mime: "image/png".into(),
            bytes: 10,
            width: Some(1),
            height: Some(2),
            duration_ms: None,
            source_url: Some("https://p3.byteimg.com/a.png".into()),
        }]);
        let v = task_detail_json(&t);
        assert_eq!(v["completed_at"], 3);
        assert_eq!(
            v["result"]["files"][0]["source_url"],
            "https://p3.byteimg.com/a.png"
        );
        let f = &v["result"]["files"][0];
        assert_eq!(f["index"], 0);
        assert_eq!(f["mime"], "image/png");
        assert_eq!(f["width"], 1);
        assert_eq!(f["bytes"], 10);
        assert!(f["duration_ms"].is_null());
        assert_eq!(
            f["download_url"],
            "/api/v1/tasks/01HQZ9ABCDEFGHJKMNPQRSTVWX/download/0"
        );
        assert_eq!(v["result"]["watermark_source"], "conversation_original");
    }

    #[test]
    fn detail_json_failed_has_failed_at_and_error() {
        let mut t = task(TaskStatus::Failed);
        t.error = Some(TaskError {
            code: "generation_doubao_refused".into(),
            message: "豆包拒绝生成".into(),
            details: None,
        });
        let v = task_detail_json(&t);
        assert_eq!(v["failed_at"], 3);
        assert_eq!(v["error"]["code"], "generation_doubao_refused");
    }
}
