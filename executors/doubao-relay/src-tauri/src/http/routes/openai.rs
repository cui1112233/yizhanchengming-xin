//! OpenAI 兼容对外接口(参照 doubao2api `unified_server.py`):挂在根路径,与 `/api/v1` 共用
//! token、上传存储与任务队列。
//!
//! - `GET  /v1/models`
//! - `POST /v1/images/generations`:**同步**,等任务终态后返回
//!
//! 参考图用请求体字段 `ref_image`(字符串或数组),每项是 `data:image/...;base64,...`
//! 或 `http(s)://` 图片地址;relay 解码/下载后存成本机文件再交给豆包。
//! - `POST /v1/video/generations`:**异步**,立即返回 `id`
//! - `GET  /v1/video/generations/{id}`:查询视频任务
//! - `GET  /v1/accounts`(`?refresh=1` 强制重读)/ `POST /v1/accounts/switch`:下游据此指定账号;
//!   生成请求里带 `account` 时会自动切换,切换接口只在想提前切好时用
//!
//! 结果里的 `url` / `video_url` 是捕获到的豆包 CDN 无水印原始地址(带签名,会过期),
//! `local_url` 是 relay 本地下载地址(需同一个 token)。

use std::sync::Arc;
use std::time::Duration;

use axum::extract::{DefaultBodyLimit, Path, State};
use axum::http::header::HOST;
use axum::http::{HeaderMap, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::routing::{get, post};
use axum::{Json, Router};
use serde_json::{json, Map, Value};
use tokio::sync::broadcast::error::RecvError;

use crate::error::{http_status_for_code, RelayError, RelayResult};
use crate::http::error::{malformed, parse_json_object};
use crate::http::routes::uploads::{store_upload, UploadKind, UPLOAD_MAX_BYTES};
use crate::state::AppState;
use crate::tasks::detail::download_url;
use crate::tasks::scheduler::{task_not_found, TaskEvent};
use crate::tasks::validate::{invalid_enum, invalid_range, missing_field};
use crate::tasks::{ConfirmationMode, GenerateRequest, Task, TaskStatus, TaskType};

/// 对外模型 ID → 豆包菜单里的模型名。
pub const MODELS: &[(&str, &str, TaskType)] = &[
    ("doubao-image-4.0", "Seedream 4.0", TaskType::Image),
    ("doubao-image-4.5", "Seedream 4.5", TaskType::Image),
    (
        "doubao-image-5.0-flash",
        "Seedream 5.0 Flash",
        TaskType::Image,
    ),
    ("doubao-video-mini", "Seedance 2.0 Mini", TaskType::Video),
    ("doubao-video-fast", "Seedance 2.0 Fast", TaskType::Video),
];
pub const DEFAULT_IMAGE_MODEL: &str = "doubao-image-4.5";
pub const DEFAULT_VIDEO_MODEL: &str = "doubao-video-fast";
pub const DEFAULT_VIDEO_DURATION: u32 = 5;
pub const DEFAULT_VIDEO_RATIO: &str = "16:9";
/// 同步生图最长等待(含排队);超时返回 504,任务继续在后台跑。
pub const IMAGE_WAIT_LIMIT: Duration = Duration::from_secs(600);
/// 生成请求体上限(内联 base64 参考图)。
const BODY_LIMIT: usize = 100 * 1024 * 1024;
/// 下载参考图网址的超时。
const REF_DOWNLOAD_TIMEOUT: Duration = Duration::from_secs(30);
/// 广播丢消息时的兜底轮询间隔。
const WAIT_POLL: Duration = Duration::from_secs(5);

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/v1/models", get(models))
        .route(
            "/v1/images/generations",
            post(image_generations).layer(DefaultBodyLimit::max(BODY_LIMIT)),
        )
        .route(
            "/v1/video/generations",
            post(video_generations).layer(DefaultBodyLimit::max(BODY_LIMIT)),
        )
        .route("/v1/video/generations/{id}", get(video_generation_status))
        .route("/v1/accounts", get(accounts))
        .route("/v1/accounts/switch", post(switch_account))
}

fn model_ids(kind: TaskType) -> Vec<&'static str> {
    MODELS
        .iter()
        .filter(|(_, _, t)| *t == kind)
        .map(|(id, _, _)| *id)
        .collect()
}

/// `model` 字段 → 豆包模型名;缺省用该类型默认模型,类型不符或未知 → `request_invalid_enum`。
pub fn resolve_model(model: Option<&str>, kind: TaskType) -> RelayResult<&'static str> {
    let id = model.unwrap_or(match kind {
        TaskType::Image => DEFAULT_IMAGE_MODEL,
        TaskType::Video => DEFAULT_VIDEO_MODEL,
    });
    MODELS
        .iter()
        .find(|(m, _, t)| *m == id && *t == kind)
        .map(|(_, name, _)| *name)
        .ok_or_else(|| invalid_enum("model", id, &model_ids(kind)))
}

/// OpenAI 风格 `size` → 豆包比例(同参考项目 `_size_to_ratio`)。
pub fn size_to_ratio(size: Option<&str>) -> String {
    match size {
        None | Some("") => "1:1".into(),
        Some("1024x1024") => "1:1".into(),
        Some("1792x1024") => "16:9".into(),
        Some("1024x1792") => "9:16".into(),
        Some("1024x768") => "4:3".into(),
        Some("768x1024") => "3:4".into(),
        Some(s) if s.contains(':') => s.into(),
        Some(_) => "1:1".into(),
    }
}

fn opt_str<'a>(obj: &'a Map<String, Value>, field: &str) -> RelayResult<Option<&'a str>> {
    match obj.get(field) {
        None | Some(Value::Null) => Ok(None),
        Some(Value::String(s)) if s.trim().is_empty() => Ok(None),
        Some(Value::String(s)) => Ok(Some(s.as_str())),
        Some(other) => Err(malformed(format!(
            "字段 {field} 必须是字符串,实际为 {other}"
        ))),
    }
}

/// `ref_image`:字符串或字符串数组(data URI 或图片网址)。
fn ref_images(obj: &Map<String, Value>) -> RelayResult<Vec<String>> {
    match obj.get("ref_image") {
        None | Some(Value::Null) => Ok(Vec::new()),
        Some(Value::String(s)) if s.trim().is_empty() => Ok(Vec::new()),
        Some(Value::String(s)) => Ok(vec![s.clone()]),
        Some(Value::Array(items)) => items
            .iter()
            .map(|v| {
                v.as_str()
                    .map(str::to_string)
                    .ok_or_else(|| malformed("ref_image 数组里只能是字符串"))
            })
            .collect(),
        Some(other) => Err(malformed(format!(
            "字段 ref_image 必须是字符串或字符串数组,实际为 {other}"
        ))),
    }
}

/// 解析 OpenAI 风格请求体为内部 `GenerateRequest`;`account` 为 `None` 时由调用方补当前账号,
/// 第三项是未解析的参考图(data URI / 网址),由 [`resolve_ref_images`] 落成本机文件。
pub fn parse_openai_request(
    obj: &Map<String, Value>,
    kind: TaskType,
) -> RelayResult<(Option<String>, GenerateRequest, Vec<String>)> {
    let prompt = opt_str(obj, "prompt")?.ok_or_else(|| missing_field("prompt"))?;
    let model_name = resolve_model(opt_str(obj, "model")?, kind)?;
    let account = opt_str(obj, "account")?.map(str::to_string);
    let ratio = match opt_str(obj, "ratio")? {
        Some(r) => r.to_string(),
        None => match kind {
            TaskType::Image => size_to_ratio(opt_str(obj, "size")?),
            TaskType::Video => match opt_str(obj, "size")? {
                Some(s) => size_to_ratio(Some(s)),
                None => DEFAULT_VIDEO_RATIO.into(),
            },
        },
    };
    if let Some(n) = obj.get("n").and_then(Value::as_u64) {
        if n != 1 {
            return Err(invalid_range("n", n.min(u32::MAX as u64) as u32, 1, 1));
        }
    }
    if let Some(f) = opt_str(obj, "response_format")? {
        if f != "url" {
            return Err(invalid_enum("response_format", f, &["url"]));
        }
    }
    let duration = match kind {
        TaskType::Image => None,
        TaskType::Video => match obj.get("duration") {
            None | Some(Value::Null) => Some(DEFAULT_VIDEO_DURATION),
            Some(v) => Some(
                v.as_u64()
                    .and_then(|d| u32::try_from(d).ok())
                    .ok_or_else(|| malformed("字段 duration 必须是正整数(秒)"))?,
            ),
        },
    };
    let req = GenerateRequest {
        account: account.clone().unwrap_or_default(),
        task_type: kind,
        model_name: model_name.to_string(),
        prompt: prompt.to_string(),
        ratio,
        duration,
        reference_files: Vec::new(),
        confirmation_mode: ConfirmationMode::Auto,
        auto_use_membership: true,
        callback_url: None,
    };
    let refs = ref_images(obj)?;
    let max = crate::tasks::validate::max_references(&req);
    if refs.len() > max {
        return Err(crate::tasks::validate::too_many_references(refs.len(), max));
    }
    Ok((account, req, refs))
}

fn bad_ref(index: usize, reason: impl Into<String>) -> RelayError {
    RelayError::new(
        "request_malformed",
        "参考图无法读取,请传 data:image/...;base64,... 或可直接下载的图片网址",
    )
    .with_details(json!({ "field": format!("ref_image[{index}]"), "reason": reason.into() }))
}

/// `data:image/png;base64,....` → 字节。
pub fn decode_data_uri(s: &str) -> Result<Vec<u8>, String> {
    use base64::Engine;
    let rest = s.strip_prefix("data:").ok_or("不是 data URI")?;
    let (meta, data) = rest.split_once(',').ok_or("data URI 缺少逗号")?;
    if !meta.ends_with(";base64") {
        return Err("只支持 base64 编码的 data URI".into());
    }
    let clean: String = data.chars().filter(|c| !c.is_ascii_whitespace()).collect();
    base64::engine::general_purpose::STANDARD
        .decode(clean.as_bytes())
        .map_err(|e| format!("base64 解码失败:{e}"))
}

async fn download_ref(url: &str) -> Result<Vec<u8>, String> {
    let client = reqwest::Client::builder()
        .timeout(REF_DOWNLOAD_TIMEOUT)
        .redirect(reqwest::redirect::Policy::limited(5))
        .build()
        .map_err(|e| e.to_string())?;
    let mut resp = client
        .get(url)
        .send()
        .await
        .map_err(|e| format!("下载失败:{e}"))?;
    if !resp.status().is_success() {
        return Err(format!("下载失败:HTTP {}", resp.status().as_u16()));
    }
    let mut buf = Vec::new();
    while let Some(chunk) = resp.chunk().await.map_err(|e| format!("下载失败:{e}"))? {
        if (buf.len() + chunk.len()) as u64 > UPLOAD_MAX_BYTES {
            return Err("图片超过 50MB 上限".into());
        }
        buf.extend_from_slice(&chunk);
    }
    Ok(buf)
}

/// 参考图(data URI / 网址)→ 存成本机上传文件,返回上传 id(供 `reference_files`)。
pub async fn resolve_ref_images(state: &AppState, refs: &[String]) -> RelayResult<Vec<String>> {
    let mut ids = Vec::with_capacity(refs.len());
    for (i, r) in refs.iter().enumerate() {
        let r = r.trim();
        let bytes = if r.starts_with("data:") {
            decode_data_uri(r).map_err(|e| bad_ref(i, e))?
        } else if r.starts_with("http://") || r.starts_with("https://") {
            download_ref(r).await.map_err(|e| bad_ref(i, e))?
        } else {
            return Err(bad_ref(i, "既不是 data URI 也不是 http(s) 网址"));
        };
        if bytes.len() as u64 > UPLOAD_MAX_BYTES {
            return Err(bad_ref(i, "图片超过 50MB 上限"));
        }
        let record = store_upload(state, &bytes, None, Some(UploadKind::Image)).await?;
        ids.push(record.id);
    }
    Ok(ids)
}

/// 未指定账号时按轮询选择已登录账号；显式传入 account 时仍然固定到该账号。
async fn fill_account(
    state: &AppState,
    account: Option<String>,
    req: &mut GenerateRequest,
) -> RelayResult<()> {
    req.account = match account {
        Some(a) => a,
        None => state.scheduler.next_auto_account().await?,
    };
    Ok(())
}

fn local_url(headers: &HeaderMap, task_id: &str, index: u32) -> String {
    let path = download_url(task_id, index);
    match headers.get(HOST).and_then(|h| h.to_str().ok()) {
        Some(host) => format!("http://{host}{path}"),
        None => path,
    }
}

/// 任务失败 → 错误响应:生成过程失败(`generation_*`)按上游失败给 502,其余按错误码表。
fn task_error_response(task: &Task) -> Response {
    let (code, message, details) = match &task.error {
        Some(e) => (e.code.clone(), e.message.clone(), e.details.clone()),
        None => (
            "generation_cancelled".to_string(),
            "任务已取消".to_string(),
            None,
        ),
    };
    let status = if code.starts_with("generation_") {
        502
    } else {
        http_status_for_code(&code)
    };
    let mut details = details.unwrap_or_else(|| json!({}));
    if let Value::Object(d) = &mut details {
        d.insert("task_id".into(), json!(task.id));
    }
    (
        StatusCode::from_u16(status).unwrap_or(StatusCode::BAD_GATEWAY),
        Json(json!({ "error": { "code": code, "message": message, "details": details } })),
    )
        .into_response()
}

fn image_data(task: &Task, headers: &HeaderMap) -> Vec<Value> {
    task.result_files
        .iter()
        .flatten()
        .map(|f| {
            let local = local_url(headers, &task.id, f.index);
            json!({
                "url": f.source_url.clone().unwrap_or_else(|| local.clone()),
                "local_url": local,
                "revised_prompt": task.request.prompt,
                "mime": f.mime,
                "width": f.width,
                "height": f.height,
                "bytes": f.bytes,
            })
        })
        .collect()
}

fn video_data(task: &Task, headers: &HeaderMap) -> Vec<Value> {
    task.result_files
        .iter()
        .flatten()
        .map(|f| {
            let local = local_url(headers, &task.id, f.index);
            json!({
                "video_url": f.source_url.clone().unwrap_or_else(|| local.clone()),
                "local_url": local,
                "cover_url": Value::Null,
                "duration": f.duration_ms.map(|ms| ms as f64 / 1000.0),
                "width": f.width,
                "height": f.height,
                "bytes": f.bytes,
            })
        })
        .collect()
}

/// 订阅后再提交,等到该任务终态;广播丢消息时退回按 storage 轮询。超时返回 `Ok(None)`。
async fn submit_and_wait(
    state: &AppState,
    req: GenerateRequest,
    limit: Duration,
) -> RelayResult<(String, Option<Task>)> {
    let mut rx = state.scheduler.subscribe();
    let (task_id, _, _) = state.submit_task(req).await?;
    let deadline = tokio::time::Instant::now() + limit;
    let mut poll = tokio::time::interval(WAIT_POLL);
    poll.tick().await;
    loop {
        tokio::select! {
            _ = tokio::time::sleep_until(deadline) => return Ok((task_id, None)),
            ev = rx.recv() => match ev {
                Ok(TaskEvent::Completed(task)) if task.id == task_id => return Ok((task_id, Some(task))),
                Ok(_) => {}
                Err(RecvError::Lagged(_)) => {}
                Err(RecvError::Closed) => return Err(RelayError::internal("任务事件通道已关闭")),
            },
            _ = poll.tick() => {}
        }
        if let Some(task) = state.scheduler.get_task(&task_id).await? {
            if task.status.is_terminal() {
                return Ok((task_id, Some(task)));
            }
        }
    }
}

async fn models() -> Json<Value> {
    let data: Vec<Value> = MODELS
        .iter()
        .map(|(id, name, kind)| {
            json!({
                "id": id,
                "object": "model",
                "owned_by": "doubao",
                "created": 0,
                "type": kind,
                "doubao_model": name,
            })
        })
        .collect();
    Json(json!({ "object": "list", "data": data }))
}

async fn image_generations(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
    body: axum::body::Bytes,
) -> Result<Response, RelayError> {
    let obj = parse_json_object(&body, false)?;
    let (account, mut req, refs) = parse_openai_request(&obj, TaskType::Image)?;
    fill_account(&state, account, &mut req).await?;
    req.reference_files = resolve_ref_images(&state, &refs).await?;
    let (task_id, task) = submit_and_wait(&state, req, IMAGE_WAIT_LIMIT).await?;
    let Some(task) = task else {
        let e = RelayError::new(
            "generation_task_timeout",
            "等待生图结果超时,任务仍在后台运行,可用 task_id 查询",
        )
        .with_details(json!({ "task_id": task_id }));
        return Ok((StatusCode::GATEWAY_TIMEOUT, Json(json!({ "error": e }))).into_response());
    };
    if task.status != TaskStatus::Completed {
        return Ok(task_error_response(&task));
    }
    Ok(Json(json!({
        "created": task.created_at / 1000,
        "task_id": task.id,
        "model": obj.get("model").cloned().unwrap_or(json!(DEFAULT_IMAGE_MODEL)),
        "data": image_data(&task, &headers),
    }))
    .into_response())
}

async fn video_generations(
    State(state): State<Arc<AppState>>,
    body: axum::body::Bytes,
) -> RelayResult<Json<Value>> {
    let obj = parse_json_object(&body, false)?;
    let (account, mut req, refs) = parse_openai_request(&obj, TaskType::Video)?;
    fill_account(&state, account, &mut req).await?;
    req.reference_files = resolve_ref_images(&state, &refs).await?;
    let (task_id, queue_position, created_at) = state.submit_task(req).await?;
    Ok(Json(json!({
        "id": task_id,
        "object": "video.generation",
        "model": obj.get("model").cloned().unwrap_or(json!(DEFAULT_VIDEO_MODEL)),
        "status": TaskStatus::Pending.as_str(),
        "queue_position": queue_position,
        "created": created_at / 1000,
    })))
}

async fn video_generation_status(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> RelayResult<Json<Value>> {
    let task = state
        .scheduler
        .get_task(&id)
        .await?
        .filter(|t| t.request.task_type == TaskType::Video)
        .ok_or_else(|| task_not_found(&id))?;
    let mut v = json!({
        "id": task.id,
        "object": "video.generation",
        "status": task.status.as_str(),
        "stage": task.stage.map(|s| s.as_str()),
        "message": task.stage_message,
        "progress": task.progress,
        "queue_position": task.queue_position,
        "created": task.created_at / 1000,
    });
    if task.status == TaskStatus::Completed {
        v["data"] = json!(video_data(&task, &headers));
        v["completed_at"] = json!(task.completed_at.map(|ms| ms / 1000));
    }
    if matches!(task.status, TaskStatus::Failed | TaskStatus::Cancelled) {
        v["error"] = match &task.error {
            Some(e) => json!({ "code": e.code, "message": e.message, "details": e.details }),
            None => json!({ "code": "generation_cancelled", "message": "任务已取消" }),
        };
    }
    Ok(Json(v))
}

/// 账号列表(OpenAI 风格 `object: list`)。
pub fn accounts_json(list: &crate::doubao::AccountList) -> Value {
    let data: Vec<Value> = list
        .accounts
        .iter()
        .map(|a| {
            json!({
                "name": a.name,
                "subscription": a.subscription,
                "is_current": a.is_current,
                "is_ambiguous": a.is_ambiguous,
            })
        })
        .collect();
    json!({
        "object": "list",
        "current": list.current,
        "data": data,
        "read_at": list.read_at,
    })
}

async fn accounts(
    State(state): State<Arc<AppState>>,
    uri: axum::http::Uri,
) -> RelayResult<Json<Value>> {
    let force = uri.query().is_some_and(|q| {
        url::form_urlencoded::parse(q.as_bytes())
            .any(|(k, v)| k == "refresh" && matches!(v.as_ref(), "1" | "true"))
    });
    let list = state.scheduler.backend().accounts(force).await?;
    Ok(Json(accounts_json(&list)))
}

/// 请求体 `{"account": "张三"}`(也接受 `/api/v1` 的 `target`)。
async fn switch_account(
    State(state): State<Arc<AppState>>,
    body: axum::body::Bytes,
) -> RelayResult<Json<Value>> {
    let obj = parse_json_object(&body, false)?;
    let target = match opt_str(&obj, "account")? {
        Some(a) => a,
        None => opt_str(&obj, "target")?.ok_or_else(|| missing_field("account"))?,
    };
    let r = state.switch_account(target).await?;
    Ok(Json(json!({
        "ok": r.ok,
        "previous": r.previous,
        "current": r.current,
        "switched_at": r.switched_at,
    })))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn obj(v: Value) -> Map<String, Value> {
        v.as_object().unwrap().clone()
    }

    #[test]
    fn resolve_model_maps_ids_defaults_and_rejects_wrong_kind() {
        assert_eq!(
            resolve_model(None, TaskType::Image).unwrap(),
            "Seedream 4.5"
        );
        assert_eq!(
            resolve_model(None, TaskType::Video).unwrap(),
            "Seedance 2.0 Fast"
        );
        assert_eq!(
            resolve_model(Some("doubao-image-4.0"), TaskType::Image).unwrap(),
            "Seedream 4.0"
        );
        assert_eq!(
            resolve_model(Some("doubao-video-mini"), TaskType::Video).unwrap(),
            "Seedance 2.0 Mini"
        );
        let e = resolve_model(Some("doubao-video-fast"), TaskType::Image).unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(
            e.details.unwrap()["allowed"],
            json!([
                "doubao-image-4.0",
                "doubao-image-4.5",
                "doubao-image-5.0-flash"
            ])
        );
        assert!(resolve_model(Some("Seedream 4.5"), TaskType::Image).is_err());
    }

    #[test]
    fn size_maps_like_reference_project() {
        assert_eq!(size_to_ratio(None), "1:1");
        assert_eq!(size_to_ratio(Some("1792x1024")), "16:9");
        assert_eq!(size_to_ratio(Some("768x1024")), "3:4");
        assert_eq!(size_to_ratio(Some("4:3")), "4:3");
        assert_eq!(size_to_ratio(Some("999x999")), "1:1");
    }

    #[test]
    fn image_request_parsing() {
        let (acc, r, refs) = parse_openai_request(
            &obj(json!({"prompt":"猫","size":"1792x1024","ref_image":"https://x/a.png"})),
            TaskType::Image,
        )
        .unwrap();
        assert!(acc.is_none());
        assert_eq!(r.model_name, "Seedream 4.5");
        assert_eq!(r.ratio, "16:9");
        assert_eq!(refs, vec!["https://x/a.png"]);
        assert!(r.reference_files.is_empty());
        assert_eq!(r.duration, None);

        let (acc, r, refs) = parse_openai_request(
            &obj(json!({"prompt":"猫","ratio":"9:16","size":"1792x1024","account":"烤鱼","ref_image":["a","b"]})),
            TaskType::Image,
        )
        .unwrap();
        assert_eq!(acc.as_deref(), Some("烤鱼"));
        assert_eq!(r.ratio, "9:16", "ratio 优先于 size");
        assert_eq!(refs, vec!["a", "b"]);
        let many: Vec<String> = (0..11).map(|i| format!("data:{i}")).collect();
        let e = parse_openai_request(
            &obj(json!({"prompt":"x","ref_image":many})),
            TaskType::Image,
        )
        .unwrap_err();
        assert_eq!(e.code, "request_too_many_references");

        let e = parse_openai_request(&obj(json!({"model":"doubao-image-4.5"})), TaskType::Image)
            .unwrap_err();
        assert_eq!(e.code, "request_missing_field");
        let e =
            parse_openai_request(&obj(json!({"prompt":"x","n":2})), TaskType::Image).unwrap_err();
        assert_eq!(e.code, "request_invalid_range");
        let e = parse_openai_request(
            &obj(json!({"prompt":"x","response_format":"b64_json"})),
            TaskType::Image,
        )
        .unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
    }

    #[test]
    fn data_uri_decoding() {
        assert_eq!(
            decode_data_uri("data:image/png;base64,aGk=").unwrap(),
            b"hi"
        );
        assert_eq!(
            decode_data_uri("data:image/png;base64,aG\nk=").unwrap(),
            b"hi"
        );
        assert!(decode_data_uri("data:image/png,hi").is_err());
        assert!(decode_data_uri("data:image/png;base64,***").is_err());
        assert!(decode_data_uri("https://x").is_err());
    }

    #[test]
    fn video_request_defaults() {
        let (_, r, _) =
            parse_openai_request(&obj(json!({"prompt":"柴犬"})), TaskType::Video).unwrap();
        assert_eq!(r.model_name, "Seedance 2.0 Fast");
        assert_eq!(r.duration, Some(DEFAULT_VIDEO_DURATION));
        assert_eq!(r.ratio, "16:9");
        let (_, r, _) = parse_openai_request(
            &obj(json!({"prompt":"柴犬","model":"doubao-video-mini","duration":12,"ratio":"9:16"})),
            TaskType::Video,
        )
        .unwrap();
        assert_eq!(r.model_name, "Seedance 2.0 Mini");
        assert_eq!(r.duration, Some(12));
        let e = parse_openai_request(&obj(json!({"prompt":"x","duration":"5s"})), TaskType::Video)
            .unwrap_err();
        assert_eq!(e.code, "request_malformed");
    }
}
