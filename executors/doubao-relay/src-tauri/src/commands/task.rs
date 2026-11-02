//! `task.*` 命令:列表 / 详情 / 提交 / 取消 / 手动确认 / 打开结果目录(`docs/spec/05-tauri-ipc.md`)。
//!
//! - 详情体与 HTTP `GET /tasks/{id}` 同源(`tasks::detail::task_detail_json`)再 `to_camel`,
//!   另给每个结果文件加 `localPath`(绝对路径,前端经 `convertFileSrc` 显示缩略图 / 播放视频)。
//! - 列表项比 HTTP 多 `promptPreview`(prompt 前 60 字,spec 06 "hover 显示 prompt 前 60 字";
//!   `TaskSummary` 不含 prompt,这里按 id 逐条补读,本地 SQLite 单页 ≤100 条)。
//! - `task_submit` 的 `request` 是 camelCase(spec 05 示例),`to_snake` 后与 HTTP `POST /generate`
//!   走同一解析(`request_missing_field` / `request_invalid_enum` / `request_malformed`)。

use std::path::{Component, Path, PathBuf};
use std::sync::Arc;

use serde::Deserialize;
use serde_json::{json, Value};
use tauri::State;

use crate::error::{RelayError, RelayResult};
use crate::http::routes::generate::generate_request_from_object;
use crate::http::routes::tasks::{clamp_limit, result_not_available};
use crate::state::AppState;
use crate::storage::files::is_safe_relative;
use crate::tasks::detail::task_detail_json;
use crate::tasks::scheduler::task_not_found;
use crate::tasks::validate::invalid_enum;
use crate::tasks::{Task, TaskStatus};

use super::{respond, to_camel, to_snake};

/// prompt 预览长度(字符数)。
pub const PROMPT_PREVIEW_CHARS: usize = 60;

/// 把数据库里的 `/` 分隔相对路径转换成当前系统的真实路径。
///
/// 不能直接 `root.join(PathBuf::from("a/b"))`：Windows 会把原始 `/` 保留在
/// `display()` 结果里，前端拿到的本地路径就和 Explorer 的路径不一致。
fn safe_rooted_path(root: &Path, relative: &Path) -> Option<PathBuf> {
    if !is_safe_relative(relative) {
        return None;
    }
    let mut resolved = root.to_path_buf();
    for component in relative.components() {
        match component {
            Component::Normal(part) => resolved.push(part),
            Component::CurDir => {}
            Component::RootDir | Component::Prefix(_) | Component::ParentDir => return None,
        }
    }
    Some(resolved)
}

/// `task.list` 的过滤参数(camelCase)。
#[derive(Debug, Default, Clone, Deserialize)]
#[serde(rename_all = "camelCase", default)]
pub struct ListFilter {
    pub status: Option<Vec<String>>,
    pub account: Option<String>,
    pub limit: Option<u32>,
    pub cursor: Option<String>,
}

const STATUS_VALUES: &[&str] = &[
    "pending",
    "preparing",
    "generating",
    "waiting_user_confirm",
    "completed",
    "failed",
    "cancelled",
];

/// prompt 前 `PROMPT_PREVIEW_CHARS` 个字符(超出加 `…`)。
pub fn prompt_preview(prompt: &str) -> String {
    let mut it = prompt.chars();
    let head: String = it.by_ref().take(PROMPT_PREVIEW_CHARS).collect();
    if it.next().is_some() {
        format!("{head}…")
    } else {
        head
    }
}

/// `task.list` → `{tasks:[TaskSummary + promptPreview], nextCursor}`。
pub async fn list(state: &AppState, filter: ListFilter) -> RelayResult<Value> {
    let statuses = match filter.status {
        None => None,
        Some(items) => {
            let parsed = items
                .iter()
                .map(|s| s.trim())
                .filter(|s| !s.is_empty())
                .map(|s| {
                    TaskStatus::parse(s).ok_or_else(|| invalid_enum("status", s, STATUS_VALUES))
                })
                .collect::<RelayResult<Vec<_>>>()?;
            Some(parsed)
        }
    };
    let account = filter
        .account
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty());
    let cursor = filter
        .cursor
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty());
    let (summaries, next_cursor) = state
        .storage
        .list_tasks(
            statuses.as_deref(),
            account,
            clamp_limit(filter.limit),
            cursor,
        )
        .await?;
    let mut tasks = Vec::with_capacity(summaries.len());
    for s in &summaries {
        let mut v = serde_json::to_value(s)
            .map_err(|e| RelayError::internal(format!("序列化任务列表失败:{e}")))?;
        let preview = state
            .storage
            .get_task(&s.id)
            .await?
            .map(|t| prompt_preview(&t.request.prompt))
            .unwrap_or_default();
        if let Some(obj) = v.as_object_mut() {
            obj.insert("prompt_preview".into(), json!(preview));
        }
        tasks.push(v);
    }
    Ok(to_camel(
        json!({ "tasks": tasks, "next_cursor": next_cursor }),
    ))
}

/// 详情体(HTTP 同源)+ `result.files[].local_path`(绝对路径)。
pub fn task_detail_ipc(task: &Task, root: &std::path::Path) -> Value {
    let mut v = task_detail_json(task);
    if let (Some(files), Some(result)) = (
        &task.result_files,
        v.get_mut("result").and_then(|r| r.get_mut("files")),
    ) {
        if let Some(items) = result.as_array_mut() {
            for item in items.iter_mut() {
                let Some(index) = item.get("index").and_then(Value::as_u64) else {
                    continue;
                };
                let local = files
                    .iter()
                    .find(|f| u64::from(f.index) == index)
                    .and_then(|f| safe_rooted_path(root, Path::new(&f.relative_path)))
                    .map(|path| path.display().to_string());
                if let Some(obj) = item.as_object_mut() {
                    obj.insert("local_path".into(), json!(local));
                }
            }
        }
    }
    to_camel(v)
}

/// `task.get` → 完整任务(camelCase)。
pub async fn get(state: &AppState, task_id: &str) -> RelayResult<Value> {
    let task = state
        .scheduler
        .get_task(task_id)
        .await?
        .ok_or_else(|| task_not_found(task_id))?;
    Ok(task_detail_ipc(&task, &state.storage.root))
}

/// `task.submit`(camelCase 请求体)→ `{taskId, status:"pending", queuePosition, createdAt}`。
pub async fn submit(state: &AppState, request: Value) -> RelayResult<Value> {
    let Value::Object(obj) = to_snake(request) else {
        return Err(RelayError::new(
            "request_malformed",
            "任务请求必须是对象,请按 {account, type, modelName, prompt, ratio} 提交",
        ));
    };
    let req = generate_request_from_object(obj)?;
    let (task_id, queue_position, created_at) = state.submit_task(req).await?;
    respond(&json!({
        "task_id": task_id,
        "status": TaskStatus::Pending.as_str(),
        "queue_position": queue_position,
        "created_at": created_at,
    }))
}

/// `task.cancel` → `{ok, cancelledFromStatus}`。
pub async fn cancel(state: &AppState, task_id: &str) -> RelayResult<Value> {
    let from = state.scheduler.cancel(task_id).await?;
    respond(&json!({ "ok": true, "cancelled_from_status": from.as_str() }))
}

/// `task.confirm` → `{ok, newStage}`。
pub async fn confirm(state: &AppState, task_id: &str) -> RelayResult<Value> {
    let new_stage = state.scheduler.confirm(task_id).await?;
    respond(&json!({ "ok": true, "new_stage": new_stage }))
}

/// `task.open_result_folder` → `{ok, path}`:在 Finder / Explorer 里打开结果目录
/// (优先定位到第一个结果文件);任务不存在 → `task_not_found`,目录不存在 → `result_not_available`。
pub async fn open_result_folder(state: &AppState, task_id: &str) -> RelayResult<Value> {
    let task = state
        .storage
        .get_task(task_id)
        .await?
        .ok_or_else(|| task_not_found(task_id))?;
    let dir = state.storage.result_dir(task_id);
    if !tokio::fs::try_exists(&dir).await.unwrap_or(false) {
        return Err(result_not_available(task_id, 0));
    }
    let first_file = task
        .result_files
        .as_deref()
        .and_then(|files| files.iter().find(|f| f.index == 0).or(files.first()))
        .and_then(|f| safe_rooted_path(&state.storage.root, Path::new(&f.relative_path)));
    let target = match first_file {
        Some(p) if tokio::fs::try_exists(&p).await.unwrap_or(false) => p,
        _ => dir.clone(),
    };
    state.platform.reveal_in_file_manager(&target).await?;
    respond(&json!({ "ok": true, "path": dir.display().to_string() }))
}

/// `task_list` 的 `filter` 既可省略也可是对象。
fn filter_from_value(filter: Option<Value>) -> RelayResult<ListFilter> {
    match filter {
        None | Some(Value::Null) => Ok(ListFilter::default()),
        Some(v @ Value::Object(_)) => serde_json::from_value(v).map_err(|e| {
            RelayError::new("request_malformed", "任务过滤条件不合法,请检查参数类型")
                .with_details(json!({ "reason": e.to_string() }))
        }),
        Some(_) => Err(RelayError::new(
            "request_malformed",
            "任务过滤条件必须是对象",
        )),
    }
}

#[tauri::command]
pub async fn task_list(
    state: State<'_, Arc<AppState>>,
    filter: Option<Value>,
) -> RelayResult<Value> {
    list(&state, filter_from_value(filter)?).await
}

#[tauri::command]
pub async fn task_get(state: State<'_, Arc<AppState>>, task_id: String) -> RelayResult<Value> {
    get(&state, &task_id).await
}

#[tauri::command]
pub async fn task_submit(state: State<'_, Arc<AppState>>, request: Value) -> RelayResult<Value> {
    submit(&state, request).await
}

#[tauri::command]
pub async fn task_cancel(state: State<'_, Arc<AppState>>, task_id: String) -> RelayResult<Value> {
    cancel(&state, &task_id).await
}

#[tauri::command]
pub async fn task_confirm(state: State<'_, Arc<AppState>>, task_id: String) -> RelayResult<Value> {
    confirm(&state, &task_id).await
}

#[tauri::command]
pub async fn task_open_result_folder(
    state: State<'_, Arc<AppState>>,
    task_id: String,
) -> RelayResult<Value> {
    open_result_folder(&state, &task_id).await
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::testing::{test_state_with, FakeBackend};
    use crate::tasks::{
        ConfirmationMode, GenerateRequest, ResultFile, StageEntry, TaskType, WatermarkSource,
    };

    fn camel_request(account: &str, prompt: &str) -> Value {
        json!({
            "account": account,
            "type": "image",
            "modelName": "Seedream 5.0 Pro",
            "prompt": prompt,
            "ratio": "1:1",
            "confirmationMode": "manual",
            "autoUseMembership": false
        })
    }

    #[test]
    fn prompt_preview_cuts_at_60_chars() {
        assert_eq!(prompt_preview("短"), "短");
        let long: String = "字".repeat(61);
        let p = prompt_preview(&long);
        assert_eq!(p.chars().count(), 61);
        assert!(p.ends_with('…'));
        assert_eq!(prompt_preview(&"a".repeat(60)), "a".repeat(60));
    }

    #[tokio::test]
    async fn submit_list_get_cancel_round_trip_in_camel_case() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_with(dir.path(), Some(FakeBackend::new(&["张三"]))).await;

        let r = submit(&state, camel_request("张三", "一只蓝色小鸟"))
            .await
            .unwrap();
        assert_eq!(r["status"], "pending");
        assert_eq!(r["queuePosition"], 1);
        let id = r["taskId"].as_str().unwrap().to_string();
        assert_eq!(id.len(), 26);
        assert!(r["createdAt"].as_i64().unwrap() > 0);
        assert!(r.get("task_id").is_none());

        // 列表:camelCase + promptPreview
        let v = list(&state, ListFilter::default()).await.unwrap();
        let tasks = v["tasks"].as_array().unwrap();
        assert_eq!(tasks.len(), 1);
        assert_eq!(tasks[0]["taskId"], id);
        assert_eq!(tasks[0]["type"], "image");
        assert_eq!(tasks[0]["status"], "pending");
        assert_eq!(tasks[0]["promptPreview"], "一只蓝色小鸟");
        assert!(tasks[0]["createdAt"].is_i64());
        assert!(v["nextCursor"].is_null());

        // 过滤:状态不匹配 → 空;非法状态 → request_invalid_enum
        let v = list(
            &state,
            ListFilter {
                status: Some(vec!["completed".into()]),
                ..Default::default()
            },
        )
        .await
        .unwrap();
        assert!(v["tasks"].as_array().unwrap().is_empty());
        let e = list(
            &state,
            ListFilter {
                status: Some(vec!["done".into()]),
                ..Default::default()
            },
        )
        .await
        .unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "status");

        // 详情:与 HTTP 同源字段 + camelCase;request 回显也是 camelCase
        let v = get(&state, &id).await.unwrap();
        assert_eq!(v["taskId"], id);
        assert_eq!(v["account"], "张三");
        assert_eq!(v["status"], "pending");
        assert_eq!(v["queuePosition"], 1);
        assert_eq!(v["request"]["modelName"], "Seedream 5.0 Pro");
        assert_eq!(v["request"]["confirmationMode"], "manual");
        assert_eq!(v["request"]["autoUseMembership"], false);
        assert_eq!(v["stageHistory"][0]["stage"], "pending");
        assert!(v.get("stage_history").is_none());

        // 取消 pending
        let v = cancel(&state, &id).await.unwrap();
        assert_eq!(v, json!({ "ok": true, "cancelledFromStatus": "pending" }));
        let v = get(&state, &id).await.unwrap();
        assert_eq!(v["status"], "cancelled");
        assert_eq!(v["error"]["code"], "generation_cancelled");
        // 再取消 → task_not_cancellable;确认 → task_not_waiting_confirm;不存在 → task_not_found
        assert_eq!(
            cancel(&state, &id).await.unwrap_err().code,
            "task_not_cancellable"
        );
        assert_eq!(
            confirm(&state, &id).await.unwrap_err().code,
            "task_not_waiting_confirm"
        );
        assert_eq!(
            get(&state, "nope").await.unwrap_err().code,
            "task_not_found"
        );
        assert_eq!(
            cancel(&state, "nope").await.unwrap_err().code,
            "task_not_found"
        );
        assert_eq!(
            confirm(&state, "nope").await.unwrap_err().code,
            "task_not_found"
        );
    }

    #[tokio::test]
    async fn submit_maps_validation_errors_like_http() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_with(dir.path(), Some(FakeBackend::new(&["张三"]))).await;

        let e = submit(&state, json!({ "type": "image" }))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "account");

        let mut r = camel_request("张三", "p");
        r["type"] = json!("audio");
        let e = submit(&state, r).await.unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "type");

        let mut r = camel_request("张三", "p");
        r["ratio"] = json!("21:9");
        let e = submit(&state, r).await.unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "ratio");

        let e = submit(&state, camel_request("张三", "   "))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "prompt");

        let e = submit(&state, camel_request("李四", "p"))
            .await
            .unwrap_err();
        assert_eq!(e.code, "account_not_found");

        let e = submit(&state, json!([1])).await.unwrap_err();
        assert_eq!(e.code, "request_malformed");

        let mut r = camel_request("张三", "p");
        r["prompt"] = json!(5);
        let e = submit(&state, r).await.unwrap_err();
        assert_eq!(e.code, "request_malformed");
    }

    #[tokio::test]
    async fn detail_adds_local_path_for_result_files() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_with(dir.path(), Some(FakeBackend::new(&["张三"]))).await;
        let id = "01HQZ9ABCDEFGHJKMNPQRSTVWX";
        let task = Task {
            id: id.into(),
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
            status: TaskStatus::Completed,
            stage: None,
            stage_message: None,
            progress: None,
            error: None,
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: Some(WatermarkSource::ConversationOriginal),
            result_files: Some(vec![
                ResultFile {
                    index: 0,
                    filename: "image_1.png".into(),
                    relative_path: format!("results/{id}/image_1.png"),
                    mime: "image/png".into(),
                    bytes: 3,
                    width: Some(1),
                    height: Some(1),
                    duration_ms: None,
                    source_url: None,
                },
                ResultFile {
                    index: 1,
                    filename: "evil.png".into(),
                    relative_path: "../evil.png".into(),
                    mime: "image/png".into(),
                    bytes: 3,
                    width: None,
                    height: None,
                    duration_ms: None,
                    source_url: None,
                },
            ]),
            created_at: 1,
            started_at: Some(2),
            completed_at: Some(3),
            expires_at: i64::MAX,
            stage_history: vec![StageEntry {
                stage: "completed".into(),
                message: None,
                ts: 3,
            }],
            queue_position: None,
        };
        state.storage.insert_task(&task).await.unwrap();

        let v = get(&state, id).await.unwrap();
        let files = v["result"]["files"].as_array().unwrap();
        assert_eq!(files.len(), 2);
        let expected = state
            .storage
            .root
            .join("results")
            .join(id)
            .join("image_1.png")
            .display()
            .to_string();
        assert_eq!(files[0]["localPath"], expected);
        assert_eq!(
            files[0]["downloadUrl"],
            format!("/api/v1/tasks/{id}/download/0")
        );
        assert!(
            files[1]["localPath"].is_null(),
            "不安全的相对路径不给出本地路径"
        );
        assert_eq!(v["result"]["watermarkSource"], "conversation_original");

        // 结果目录不存在 → result_not_available;建目录 + 文件后 FakePlatform 的 reveal 直接成功
        let e = open_result_folder(&state, id).await.unwrap_err();
        assert_eq!(e.code, "result_not_available");
        let dir_path = state.storage.result_dir(id);
        tokio::fs::create_dir_all(&dir_path).await.unwrap();
        tokio::fs::write(dir_path.join("image_1.png"), b"png")
            .await
            .unwrap();
        let v = open_result_folder(&state, id).await.unwrap();
        assert_eq!(v["ok"], true);
        assert_eq!(v["path"], dir_path.display().to_string());
        assert_eq!(
            open_result_folder(&state, "nope").await.unwrap_err().code,
            "task_not_found"
        );
    }

    #[test]
    fn filter_from_value_accepts_missing_null_and_object() {
        assert!(filter_from_value(None).unwrap().status.is_none());
        assert!(filter_from_value(Some(Value::Null))
            .unwrap()
            .limit
            .is_none());
        let f = filter_from_value(Some(json!({
            "status": ["pending", "generating"],
            "account": "张三",
            "limit": 5,
            "cursor": "01A"
        })))
        .unwrap();
        assert_eq!(f.status.unwrap(), vec!["pending", "generating"]);
        assert_eq!(f.account.as_deref(), Some("张三"));
        assert_eq!(f.limit, Some(5));
        assert_eq!(f.cursor.as_deref(), Some("01A"));
        assert_eq!(
            filter_from_value(Some(json!("x"))).unwrap_err().code,
            "request_malformed"
        );
        assert_eq!(
            filter_from_value(Some(json!({ "limit": "abc" })))
                .unwrap_err()
                .code,
            "request_malformed"
        );
    }
}
