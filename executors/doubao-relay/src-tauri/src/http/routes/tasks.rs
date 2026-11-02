//! `GET /tasks` / `GET /tasks/{id}` / `DELETE /tasks/{id}` / `POST /tasks/{id}/confirm` /
//! `GET /tasks/{id}/download/{index}`。
//!
//! 详情体由 [`crate::tasks::detail::task_detail_json`] 生成(HTTP、IPC、Webhook 回调 `task` 字段同源)。

use std::path::PathBuf;
use std::sync::Arc;

use axum::body::{Body, Bytes};
use axum::extract::rejection::{PathRejection, QueryRejection};
use axum::extract::{Path, Query, Request, State};
use axum::http::header::{CONTENT_DISPOSITION, CONTENT_TYPE};
use axum::http::{HeaderValue, StatusCode};
use axum::response::Response;
use axum::routing::get;
use axum::{Json, Router};
use serde::Deserialize;
use serde_json::{json, Value};
use tower_http::services::ServeFile;

use crate::error::{RelayError, RelayResult};
use crate::http::error::{malformed, parse_json_object};
use crate::logging::now_ms;
use crate::state::AppState;
use crate::storage::files::is_safe_relative;
use crate::tasks::scheduler::task_not_found;
use crate::tasks::validate::invalid_enum;
use crate::tasks::TaskStatus;

pub use crate::tasks::detail::{download_url, task_detail_json};

/// `GET /tasks` 的 `limit` 默认 / 上限。
pub const LIST_DEFAULT_LIMIT: u32 = 20;
pub const LIST_MAX_LIMIT: u32 = 100;

const STATUS_VALUES: &[&str] = &[
    "pending",
    "preparing",
    "generating",
    "waiting_user_confirm",
    "completed",
    "failed",
    "cancelled",
];

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/tasks", get(list))
        .route("/tasks/{id}", get(detail).delete(cancel))
        .route("/tasks/{id}/confirm", axum::routing::post(confirm))
        .route("/tasks/{id}/download/{index}", get(download))
}

#[derive(Deserialize, Default)]
pub struct ListQuery {
    pub status: Option<String>,
    pub account: Option<String>,
    pub limit: Option<u32>,
    pub cursor: Option<String>,
}

/// `status=a,b` → 状态列表(空串 / 缺省 → `None`);非法值 → `request_invalid_enum{field:"status"}`。
pub fn parse_status_filter(raw: Option<&str>) -> RelayResult<Option<Vec<TaskStatus>>> {
    let Some(raw) = raw.map(str::trim).filter(|s| !s.is_empty()) else {
        return Ok(None);
    };
    raw.split(',')
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(|s| TaskStatus::parse(s).ok_or_else(|| invalid_enum("status", s, STATUS_VALUES)))
        .collect::<RelayResult<Vec<_>>>()
        .map(Some)
}

/// `limit` 缺省 20,夹到 1..=100。
pub fn clamp_limit(limit: Option<u32>) -> u32 {
    limit.unwrap_or(LIST_DEFAULT_LIMIT).clamp(1, LIST_MAX_LIMIT)
}

/// `result_not_available{task_id, index}`(404)。
pub fn result_not_available(task_id: &str, index: u32) -> RelayError {
    RelayError::new(
        "result_not_available",
        "结果文件不存在,请等任务完成后再下载",
    )
    .with_details(json!({ "task_id": task_id, "index": index }))
}

/// 非完成态任务下载结果时的 `result_not_available`:进行中提示"等完成",
/// 失败 / 取消(永远不会完成)提示"没有可下载的结果"。
pub fn result_not_ready(task_id: &str, index: u32, status: TaskStatus) -> RelayError {
    if !status.is_terminal() {
        return result_not_available(task_id, index);
    }
    RelayError::new(
        "result_not_available",
        "任务未成功完成,没有可下载的结果,请查看任务错误信息后重新提交",
    )
    .with_details(json!({ "task_id": task_id, "index": index }))
}

/// `result_expired{task_id, expired_at}`(404)。
pub fn result_expired(task_id: &str, expired_at: i64) -> RelayError {
    RelayError::new("result_expired", "结果文件已过期清理,请重新提交任务")
        .with_details(json!({ "task_id": task_id, "expired_at": expired_at }))
}

/// `result_index_out_of_range{requested, available}`(404)。
pub fn result_index_out_of_range(requested: u32, available: usize) -> RelayError {
    RelayError::new(
        "result_index_out_of_range",
        "结果文件序号超出范围,请按 result.files 里的 index 下载",
    )
    .with_details(json!({ "requested": requested, "available": available }))
}

async fn list(
    State(state): State<Arc<AppState>>,
    query: Result<Query<ListQuery>, QueryRejection>,
) -> RelayResult<Json<Value>> {
    let Query(q) = query.map_err(|e| malformed(format!("查询参数不合法:{}", e.body_text())))?;
    let statuses = parse_status_filter(q.status.as_deref())?;
    let account = q
        .account
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty());
    let cursor = q.cursor.as_deref().map(str::trim).filter(|s| !s.is_empty());
    let (tasks, next_cursor) = state
        .storage
        .list_tasks(statuses.as_deref(), account, clamp_limit(q.limit), cursor)
        .await?;
    Ok(Json(json!({ "tasks": tasks, "next_cursor": next_cursor })))
}

async fn detail(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> RelayResult<Json<Value>> {
    let task = state
        .scheduler
        .get_task(&id)
        .await?
        .ok_or_else(|| task_not_found(&id))?;
    Ok(Json(task_detail_json(&task)))
}

async fn cancel(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> RelayResult<Json<Value>> {
    let from = state.scheduler.cancel(&id).await?;
    Ok(Json(
        json!({ "ok": true, "cancelled_from_status": from.as_str() }),
    ))
}

async fn confirm(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    body: Bytes,
) -> RelayResult<Json<Value>> {
    // body 可选(`{payload:"continue"}`),只校验是 JSON 对象,不看内容。
    parse_json_object(&body, true)?;
    let new_stage = state.scheduler.confirm(&id).await?;
    Ok(Json(json!({ "ok": true, "new_stage": new_stage })))
}

/// `locate_result` 的产物:落盘路径、库里的 mime、扩展名、任务过期时间。
struct LocatedResult {
    path: PathBuf,
    mime: String,
    ext: String,
    expires_at: i64,
}

/// 结果文件定位:完成态 + 未过期 + index 在范围内 + 相对路径安全 + 文件存在。
async fn locate_result(state: &AppState, id: &str, index: u32) -> RelayResult<LocatedResult> {
    let task = state
        .scheduler
        .get_task(id)
        .await?
        .ok_or_else(|| task_not_found(id))?;
    if task.status != TaskStatus::Completed {
        return Err(result_not_ready(id, index, task.status));
    }
    // 完成态:先看过期,再看文件列表。GC 步骤 1 会把过期任务的 `result_files` 置 NULL,
    // 所以"完成但没有文件"同样是已过期清理,而不是"结果不可用"。
    let files = match &task.result_files {
        Some(files) if !files.is_empty() && now_ms() < task.expires_at => files,
        _ => return Err(result_expired(id, task.expires_at)),
    };
    let file = files
        .iter()
        .find(|f| f.index == index)
        .ok_or_else(|| result_index_out_of_range(index, files.len()))?;
    let rel = PathBuf::from(&file.relative_path);
    if !is_safe_relative(&rel) {
        return Err(result_not_available(id, index));
    }
    let path = state.storage.root.join(rel);
    if !tokio::fs::try_exists(&path).await.unwrap_or(false) {
        return Err(result_expired(id, task.expires_at));
    }
    let ext = path
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("bin")
        .to_string();
    Ok(LocatedResult {
        path,
        mime: file.mime.clone(),
        ext,
        expires_at: task.expires_at,
    })
}

async fn download(
    State(state): State<Arc<AppState>>,
    path: Result<Path<(String, u32)>, PathRejection>,
    req: Request,
) -> RelayResult<Response> {
    let Path((id, index)) =
        path.map_err(|e| malformed(format!("路径参数不合法:{}", e.body_text())))?;
    let located = locate_result(&state, &id, index).await?;
    serve_result(req, &id, index, located).await
}

/// 用 `ServeFile` 发文件(Range → 206 由它处理)。`locate_result` 的存在性检查与这里的打开之间
/// 文件可能被 GC 删掉:`ServeFile` 对此返回的是 404 **响应**而不是 `Err`,统一映射成
/// `result_expired{task_id, expired_at}`;真正的 `Err` 只有打开 / 读取出错(`system_io_error`)。
async fn serve_result(
    req: Request,
    id: &str,
    index: u32,
    located: LocatedResult,
) -> RelayResult<Response> {
    let LocatedResult {
        path,
        mime,
        ext,
        expires_at,
    } = located;
    let mime_type: mime::Mime = mime.parse().unwrap_or(mime::APPLICATION_OCTET_STREAM);
    let resp = ServeFile::new_with_mime(&path, &mime_type)
        .try_call(req)
        .await
        .map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                result_expired(id, expires_at)
            } else {
                RelayError::from(e)
            }
        })?;
    if resp.status() == StatusCode::NOT_FOUND {
        return Err(result_expired(id, expires_at));
    }
    let mut resp = resp.map(Body::new);
    let disposition = format!("attachment; filename=\"task-{id}-{index}.{ext}\"");
    if let Ok(v) = HeaderValue::from_str(&disposition) {
        resp.headers_mut().insert(CONTENT_DISPOSITION, v);
    }
    if let Ok(v) = HeaderValue::from_str(&mime) {
        resp.headers_mut().insert(CONTENT_TYPE, v);
    }
    Ok(resp)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn status_filter_and_limit_parsing() {
        assert_eq!(parse_status_filter(None).unwrap(), None);
        assert_eq!(parse_status_filter(Some("")).unwrap(), None);
        assert_eq!(
            parse_status_filter(Some("pending, completed")).unwrap(),
            Some(vec![TaskStatus::Pending, TaskStatus::Completed])
        );
        let e = parse_status_filter(Some("pending,done")).unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "status");
        assert_eq!(d["value"], "done");
        assert_eq!(clamp_limit(None), 20);
        assert_eq!(clamp_limit(Some(0)), 1);
        assert_eq!(clamp_limit(Some(500)), 100);
        assert_eq!(clamp_limit(Some(7)), 7);
    }

    /// 文件在 `locate_result` 的存在性检查之后、`ServeFile` 打开之前被 GC 删掉(竞态):
    /// `ServeFile` 给的是 404 响应而非 `Err`,必须变成 JSON `result_expired` 且带真实 `expired_at`。
    #[tokio::test]
    async fn serve_result_maps_vanished_file_to_result_expired_with_real_expiry() {
        let dir = tempfile::tempdir().unwrap();
        let located = LocatedResult {
            path: dir.path().join("gone.png"),
            mime: "image/png".into(),
            ext: "png".into(),
            expires_at: 4_242,
        };
        let req = Request::builder().body(Body::empty()).unwrap();
        let e = serve_result(req, "t1", 0, located).await.unwrap_err();
        assert_eq!(e.code, "result_expired");
        assert_eq!(e.http_status(), 404);
        let d = e.details.unwrap();
        assert_eq!(d["task_id"], "t1");
        assert_eq!(d["expired_at"], 4_242);

        // 文件在 → 正常 200 + 头
        let path = dir.path().join("ok.png");
        tokio::fs::write(&path, b"png-bytes").await.unwrap();
        let located = LocatedResult {
            path,
            mime: "image/png".into(),
            ext: "png".into(),
            expires_at: 4_242,
        };
        let req = Request::builder().body(Body::empty()).unwrap();
        let resp = serve_result(req, "t1", 2, located).await.unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
        assert_eq!(resp.headers()[CONTENT_TYPE], "image/png");
        assert_eq!(
            resp.headers()[CONTENT_DISPOSITION],
            "attachment; filename=\"task-t1-2.png\""
        );
    }

    #[test]
    fn result_errors_carry_codes_and_details() {
        let e = result_not_available("t", 1);
        assert_eq!(e.code, "result_not_available");
        assert_eq!(e.http_status(), 404);
        let e = result_expired("t", 9);
        assert_eq!(e.code, "result_expired");
        assert_eq!(e.http_status(), 404);
        assert_eq!(e.details.unwrap()["expired_at"], 9);
        let e = result_index_out_of_range(3, 1);
        assert_eq!(e.code, "result_index_out_of_range");
        assert_eq!(e.http_status(), 404);
        assert_eq!(e.details.unwrap()["available"], 1);
    }
}
