//! `POST /uploads`(multipart `file` + 可选 `kind`)/ `GET /uploads/{id}`。
//!
//! - 单文件 ≤ 50MB:`file` 字段边读边计数,超过即 `upload_too_large{bytes, max_bytes}`(400),
//!   `bytes` = 截止时已数到的**文件**字节数(不是整个 multipart body);路由层 `DefaultBodyLimit`
//!   略高于 50MB(容纳 multipart 边界),它触发的 413 也映射成同一错误——此时 `bytes` 取已数到的
//!   文件字节(若已超限),否则退回请求的 `Content-Length`(没有则记 `max_bytes + 1`)。
//! - MIME 按内容判定(不信任客户端声明):PNG / JPEG / WebP 走 `download::detect_image_format`,
//!   MP4 看 offset 4 的 `ftyp`;其它 → `upload_mime_not_supported{mime, allowed}`。
//! - 写 `uploads/{ulid}.{ext}`,`expires_at = now + storage.upload_expiry_hours`。
//! - `GET` 不存在**或已过期**都是 404 `upload_not_found`(spec 04;过期时 details 另带 `expired_at`)。

use std::sync::Arc;

use axum::extract::multipart::{Field, MultipartError, MultipartRejection};
use axum::extract::{DefaultBodyLimit, Multipart, Path, State};
use axum::http::header::CONTENT_LENGTH;
use axum::http::HeaderMap;
use axum::routing::{get, post};
use axum::{Json, Router};
use serde_json::{json, Value};

use crate::doubao::download::detect_image_format;
use crate::error::{RelayError, RelayResult};
use crate::http::error::malformed;
use crate::logging::now_ms;
use crate::state::AppState;
use crate::storage::Upload;
use crate::tasks::scheduler::upload_not_found;
use crate::tasks::validate::{invalid_enum, missing_field};

/// 单文件上限(spec 04:50MB)。
pub const UPLOAD_MAX_BYTES: u64 = 50 * 1024 * 1024;
/// 请求体上限 = 文件上限 + multipart 头部 / 边界的余量。
const BODY_LIMIT_SLACK: u64 = 1024 * 1024;
/// 支持的 MIME(spec 04)。
pub const ALLOWED_MIMES: &[&str] = &["image/png", "image/jpeg", "image/webp", "video/mp4"];
const IMAGE_MIMES: &[&str] = &["image/png", "image/jpeg", "image/webp"];
const VIDEO_MIMES: &[&str] = &["video/mp4"];
const KINDS: &[&str] = &["image", "video_reference"];

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route(
            "/uploads",
            post(upload).layer(DefaultBodyLimit::max(
                (UPLOAD_MAX_BYTES + BODY_LIMIT_SLACK) as usize,
            )),
        )
        .route("/uploads/{id}", get(get_upload))
}

/// `kind` 字段。
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum UploadKind {
    Image,
    VideoReference,
}

impl UploadKind {
    pub fn parse(s: &str) -> RelayResult<Self> {
        match s.trim() {
            "image" => Ok(Self::Image),
            "video_reference" => Ok(Self::VideoReference),
            other => Err(invalid_enum("kind", other, KINDS)),
        }
    }

    fn allowed(self) -> &'static [&'static str] {
        match self {
            Self::Image => IMAGE_MIMES,
            Self::VideoReference => VIDEO_MIMES,
        }
    }
}

/// 按内容识别:`(mime, 扩展名)`;不支持的内容返回 `None`。
pub fn detect_upload_mime(bytes: &[u8]) -> Option<(&'static str, &'static str)> {
    if let Ok((mime, ext, _)) = detect_image_format(bytes) {
        return Some((mime, ext));
    }
    if is_mp4(bytes) {
        return Some(("video/mp4", "mp4"));
    }
    None
}

/// ISO BMFF:offset 4 起为 `ftyp`。
pub fn is_mp4(bytes: &[u8]) -> bool {
    bytes.len() >= 12 && &bytes[4..8] == b"ftyp"
}

/// `upload_too_large{bytes, max_bytes}`(400)。
pub fn upload_too_large(bytes: u64, max_bytes: u64) -> RelayError {
    RelayError::new("upload_too_large", "文件超过 50MB 上限,请压缩后重新上传")
        .with_details(json!({ "bytes": bytes, "max_bytes": max_bytes }))
}

/// `upload_mime_not_supported{mime, allowed}`(400)。
pub fn upload_mime_not_supported(mime: &str, allowed: &[&str]) -> RelayError {
    RelayError::new(
        "upload_mime_not_supported",
        "文件类型不支持,请上传 PNG / JPEG / WebP 图片或 MP4 视频",
    )
    .with_details(json!({ "mime": mime, "allowed": allowed }))
}

/// 校验 `kind` 与实际内容是否相符,返回 `(mime, ext)`。
pub fn classify_upload(
    bytes: &[u8],
    declared_mime: Option<&str>,
    kind: Option<UploadKind>,
) -> RelayResult<(&'static str, &'static str)> {
    let Some((mime, ext)) = detect_upload_mime(bytes) else {
        return Err(upload_mime_not_supported(
            declared_mime.unwrap_or("unknown"),
            kind.map(UploadKind::allowed).unwrap_or(ALLOWED_MIMES),
        ));
    };
    if let Some(kind) = kind {
        if !kind.allowed().contains(&mime) {
            return Err(upload_mime_not_supported(mime, kind.allowed()));
        }
    }
    Ok((mime, ext))
}

/// body 超限(413)时 `upload_too_large.bytes` 的取值:已数到的文件字节若已超限就用它,
/// 否则退回 `Content-Length`,都没有则 `max_bytes + 1`。
pub fn too_large_bytes(file_bytes_so_far: u64, declared_len: Option<u64>) -> u64 {
    if file_bytes_so_far > UPLOAD_MAX_BYTES {
        file_bytes_so_far
    } else {
        declared_len.unwrap_or(UPLOAD_MAX_BYTES + 1)
    }
}

/// multipart 层错误:body 超限(413)→ `upload_too_large`,其它 → `request_malformed`。
pub(crate) fn multipart_error(
    e: MultipartError,
    file_bytes_so_far: u64,
    declared_len: Option<u64>,
) -> RelayError {
    if e.status() == axum::http::StatusCode::PAYLOAD_TOO_LARGE {
        upload_too_large(
            too_large_bytes(file_bytes_so_far, declared_len),
            UPLOAD_MAX_BYTES,
        )
    } else {
        malformed(format!("multipart 请求体解析失败:{}", e.body_text()))
    }
}

/// 边读边计数,超过 `UPLOAD_MAX_BYTES` 立即停止,`bytes` 报已数到的文件字节数。
pub(crate) async fn read_file_field(
    mut field: Field<'_>,
    declared_len: Option<u64>,
) -> RelayResult<Vec<u8>> {
    let mut buf: Vec<u8> = Vec::new();
    while let Some(chunk) = field
        .chunk()
        .await
        .map_err(|e| multipart_error(e, buf.len() as u64, declared_len))?
    {
        let next = buf.len() as u64 + chunk.len() as u64;
        if next > UPLOAD_MAX_BYTES {
            return Err(upload_too_large(next, UPLOAD_MAX_BYTES));
        }
        buf.extend_from_slice(&chunk);
    }
    Ok(buf)
}

fn write_error(e: std::io::Error, path: &std::path::Path) -> RelayError {
    if e.kind() == std::io::ErrorKind::StorageFull {
        RelayError::new("system_disk_full", "磁盘空间不足,请清理磁盘后重试")
            .with_details(json!({ "path": path.display().to_string() }))
    } else {
        RelayError::from(e)
    }
}

fn upload_json(u: &Upload) -> Value {
    json!({
        "upload_id": u.id,
        "mime": u.mime,
        "bytes": u.bytes,
        "expires_at": u.expires_at,
    })
}

async fn upload(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
    multipart: Result<Multipart, MultipartRejection>,
) -> RelayResult<Json<Value>> {
    let declared_len = headers
        .get(CONTENT_LENGTH)
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());
    let mut multipart = multipart
        .map_err(|e| malformed(format!("需要 multipart/form-data 请求体:{}", e.body_text())))?;

    let mut file: Option<(Vec<u8>, Option<String>)> = None;
    let mut kind: Option<UploadKind> = None;
    loop {
        let field = match multipart.next_field().await {
            Ok(Some(f)) => f,
            Ok(None) => break,
            Err(e) => return Err(multipart_error(e, 0, declared_len)),
        };
        match field.name() {
            Some("file") => {
                let declared_mime = field.content_type().map(str::to_string);
                let data = read_file_field(field, declared_len).await?;
                file = Some((data, declared_mime));
            }
            Some("kind") => {
                let text = field
                    .text()
                    .await
                    .map_err(|e| multipart_error(e, 0, declared_len))?;
                kind = Some(UploadKind::parse(&text)?);
            }
            _ => {} // 未知字段:丢弃(下一次 next_field 会跳过它)
        }
    }
    let Some((data, declared_mime)) = file else {
        return Err(missing_field("file"));
    };
    let record = store_upload(&state, &data, declared_mime.as_deref(), kind).await?;
    Ok(Json(upload_json(&record)))
}

/// 校验内容类型 → 写 `uploads/{ulid}.{ext}` → 入库(`/uploads` 与 `/v1/images/upload` 共用)。
pub(crate) async fn store_upload(
    state: &AppState,
    data: &[u8],
    declared_mime: Option<&str>,
    kind: Option<UploadKind>,
) -> RelayResult<Upload> {
    let (mime, ext) = classify_upload(data, declared_mime, kind)?;

    let id = ulid::Ulid::new().to_string();
    let path = state.storage.upload_path(&id, ext);
    tokio::fs::write(&path, data)
        .await
        .map_err(|e| write_error(e, &path))?;
    let now = now_ms();
    let expiry_hours = state.config.read().await.storage.upload_expiry_hours;
    let record = Upload {
        id,
        mime: mime.to_string(),
        bytes: data.len() as u64,
        path: path.clone(),
        created_at: now,
        expires_at: now + i64::from(expiry_hours) * 3_600_000,
    };
    if let Err(e) = state.storage.insert_upload(&record).await {
        let _ = tokio::fs::remove_file(&path).await;
        return Err(e);
    }
    tracing::info!(
        upload_id = %record.id,
        mime,
        bytes = record.bytes,
        kind = ?kind,
        "参考文件已上传"
    );
    Ok(record)
}

async fn get_upload(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> RelayResult<Json<Value>> {
    let Some(u) = state.storage.get_upload(&id).await? else {
        return Err(upload_not_found(&id));
    };
    if now_ms() >= u.expires_at {
        let mut e = upload_not_found(&id);
        if let Some(Value::Object(d)) = e.details.as_mut() {
            d.insert("expired_at".into(), json!(u.expires_at));
        }
        return Err(e);
    }
    Ok(Json(upload_json(&u)))
}

#[cfg(test)]
pub(crate) mod tests {
    use super::*;

    /// 1×1 PNG(与 download.rs 测试同源)。
    pub(crate) const PNG_1X1: &[u8] = &[
        0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44,
        0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F,
        0x15, 0xC4, 0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00,
        0x01, 0x00, 0x00, 0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
        0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
    ];

    pub(crate) fn mp4_bytes() -> Vec<u8> {
        let mut v = vec![0x00, 0x00, 0x00, 0x18];
        v.extend_from_slice(b"ftypisom");
        v.extend_from_slice(&[0u8; 24]);
        v
    }

    #[test]
    fn detect_by_content_not_declaration() {
        assert_eq!(detect_upload_mime(PNG_1X1), Some(("image/png", "png")));
        assert_eq!(detect_upload_mime(&mp4_bytes()), Some(("video/mp4", "mp4")));
        assert_eq!(detect_upload_mime(b"hello world, not a file"), None);
        assert_eq!(detect_upload_mime(b""), None);
        assert!(!is_mp4(b"\0\0\0\x18ftyp"), "不足 12 字节");
    }

    #[test]
    fn classify_checks_kind_against_content() {
        assert_eq!(
            classify_upload(PNG_1X1, Some("image/png"), None).unwrap(),
            ("image/png", "png")
        );
        assert_eq!(
            classify_upload(PNG_1X1, None, Some(UploadKind::Image)).unwrap(),
            ("image/png", "png")
        );
        let e = classify_upload(PNG_1X1, None, Some(UploadKind::VideoReference)).unwrap_err();
        assert_eq!(e.code, "upload_mime_not_supported");
        let d = e.details.unwrap();
        assert_eq!(d["mime"], "image/png");
        assert_eq!(d["allowed"], json!(["video/mp4"]));
        let e = classify_upload(&mp4_bytes(), None, Some(UploadKind::Image)).unwrap_err();
        assert_eq!(e.details.unwrap()["allowed"], json!(IMAGE_MIMES));
        let e = classify_upload(b"text", Some("text/plain"), None).unwrap_err();
        assert_eq!(e.code, "upload_mime_not_supported");
        let d = e.details.unwrap();
        assert_eq!(d["mime"], "text/plain");
        assert_eq!(d["allowed"], json!(ALLOWED_MIMES));
        assert_eq!(UploadKind::parse("image").unwrap(), UploadKind::Image);
        let e = UploadKind::parse("audio").unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "kind");
    }

    #[test]
    fn too_large_bytes_prefers_counted_file_bytes_over_content_length() {
        let max = UPLOAD_MAX_BYTES;
        assert_eq!(too_large_bytes(max + 7, Some(max + 9_000)), max + 7);
        assert_eq!(too_large_bytes(0, Some(max + 9_000)), max + 9_000);
        assert_eq!(too_large_bytes(max - 1, Some(max + 9_000)), max + 9_000);
        assert_eq!(too_large_bytes(0, None), max + 1);
    }

    #[test]
    fn upload_errors_have_spec_codes() {
        let e = upload_too_large(60, 50);
        assert_eq!(e.code, "upload_too_large");
        assert_eq!(e.http_status(), 400);
        let d = e.details.unwrap();
        assert_eq!(d["bytes"], 60);
        assert_eq!(d["max_bytes"], 50);
        assert_eq!(upload_mime_not_supported("x", &[]).http_status(), 400);
    }
}
