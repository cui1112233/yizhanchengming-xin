//! `upload.add` 命令:把本地文件(前端用 `tauri-plugin-dialog` 选出的路径)拷进 `uploads/`,
//! 返回 `{uploadId, mime, bytes, expiresAt}`。规则与 HTTP `POST /uploads` 一致:
//! ≤50MB(`upload_too_large`)、MIME 按内容判定且与 `kind` 相符(`upload_mime_not_supported`)、
//! 有效期 `storage.upload_expiry_hours`。文件读不到 → `system_io_error{path}`。

use std::path::Path;
use std::sync::Arc;

use serde_json::{json, Value};
use tauri::State;

use crate::error::{RelayError, RelayResult};
use crate::http::routes::uploads::{
    classify_upload, upload_too_large, UploadKind, UPLOAD_MAX_BYTES,
};
use crate::logging::now_ms;
use crate::state::AppState;
use crate::storage::Upload;

use super::respond;

/// `upload.add`。
pub async fn add(state: &AppState, file_path: &str, kind: Option<&str>) -> RelayResult<Value> {
    let kind = match kind.map(str::trim).filter(|k| !k.is_empty()) {
        Some(k) => Some(UploadKind::parse(k)?),
        None => None,
    };
    let path = Path::new(file_path.trim());
    if path.as_os_str().is_empty() {
        return Err(crate::tasks::validate::missing_field("filePath"));
    }
    let meta = tokio::fs::metadata(path)
        .await
        .map_err(|e| io_error(e, path))?;
    if !meta.is_file() {
        return Err(
            RelayError::new("system_io_error", "所选路径不是文件,请重新选择")
                .with_details(json!({ "path": path.display().to_string() })),
        );
    }
    if meta.len() > UPLOAD_MAX_BYTES {
        return Err(upload_too_large(meta.len(), UPLOAD_MAX_BYTES));
    }
    let data = tokio::fs::read(path).await.map_err(|e| io_error(e, path))?;
    if data.len() as u64 > UPLOAD_MAX_BYTES {
        return Err(upload_too_large(data.len() as u64, UPLOAD_MAX_BYTES));
    }
    let (mime, ext) = classify_upload(&data, None, kind)?;

    let id = ulid::Ulid::new().to_string();
    let dest = state.storage.upload_path(&id, ext);
    tokio::fs::write(&dest, &data)
        .await
        .map_err(|e| io_error(e, &dest))?;
    let now = now_ms();
    let expiry_hours = state.config.read().await.storage.upload_expiry_hours;
    let record = Upload {
        id,
        mime: mime.to_string(),
        bytes: data.len() as u64,
        path: dest.clone(),
        created_at: now,
        expires_at: now + i64::from(expiry_hours) * 3_600_000,
    };
    if let Err(e) = state.storage.insert_upload(&record).await {
        let _ = tokio::fs::remove_file(&dest).await;
        return Err(e);
    }
    tracing::info!(
        upload_id = %record.id,
        mime,
        bytes = record.bytes,
        kind = ?kind,
        "参考文件已从本地加入"
    );
    respond(&json!({
        "upload_id": record.id,
        "mime": record.mime,
        "bytes": record.bytes,
        "expires_at": record.expires_at,
    }))
}

fn io_error(e: std::io::Error, path: &Path) -> RelayError {
    let code = if e.kind() == std::io::ErrorKind::StorageFull {
        "system_disk_full"
    } else {
        "system_io_error"
    };
    let message = match e.kind() {
        std::io::ErrorKind::NotFound => "文件不存在,请重新选择后重试",
        std::io::ErrorKind::PermissionDenied => "没有读取该文件的权限,请检查文件权限后重试",
        std::io::ErrorKind::StorageFull => "磁盘空间不足,请清理磁盘后重试",
        _ => "读取或写入文件失败,请检查文件后重试",
    };
    RelayError::new(code, message)
        .with_details(json!({ "path": path.display().to_string(), "io": e.to_string() }))
}

#[tauri::command]
pub async fn upload_add(
    state: State<'_, Arc<AppState>>,
    file_path: String,
    kind: Option<String>,
) -> RelayResult<Value> {
    add(&state, &file_path, kind.as_deref()).await
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::testing::test_state;
    use crate::http::routes::uploads::tests::{mp4_bytes, PNG_1X1};

    #[tokio::test]
    async fn add_copies_png_and_records_upload() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;
        let src = dir.path().join("pick.png");
        tokio::fs::write(&src, PNG_1X1).await.unwrap();

        let v = add(&state, src.to_str().unwrap(), Some("image"))
            .await
            .unwrap();
        let id = v["uploadId"].as_str().unwrap();
        assert_eq!(id.len(), 26);
        assert_eq!(v["mime"], "image/png");
        assert_eq!(v["bytes"], PNG_1X1.len());
        let expires = v["expiresAt"].as_i64().unwrap();
        assert!(expires > now_ms() + 3_500_000, "默认 1h 过期");
        assert!(v.get("upload_id").is_none());

        let stored = state.storage.get_upload(id).await.unwrap().unwrap();
        assert_eq!(stored.mime, "image/png");
        assert_eq!(stored.path, state.storage.upload_path(id, "png"));
        assert_eq!(tokio::fs::read(&stored.path).await.unwrap(), PNG_1X1);

        // kind 缺省也可(MIME 仍按内容)
        let v = add(&state, src.to_str().unwrap(), None).await.unwrap();
        assert_eq!(v["mime"], "image/png");
        let v = add(&state, src.to_str().unwrap(), Some("")).await.unwrap();
        assert_eq!(v["mime"], "image/png");
    }

    #[tokio::test]
    async fn add_rejects_missing_file_bad_kind_and_mismatched_content() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state(dir.path()).await;

        let e = add(&state, dir.path().join("nope.png").to_str().unwrap(), None)
            .await
            .unwrap_err();
        assert_eq!(e.code, "system_io_error");
        assert!(e.details.unwrap()["path"]
            .as_str()
            .unwrap()
            .ends_with("nope.png"));

        let e = add(&state, dir.path().to_str().unwrap(), None)
            .await
            .unwrap_err();
        assert_eq!(e.code, "system_io_error");

        let e = add(&state, "  ", None).await.unwrap_err();
        assert_eq!(e.code, "request_missing_field");

        let png = dir.path().join("a.png");
        tokio::fs::write(&png, PNG_1X1).await.unwrap();
        let e = add(&state, png.to_str().unwrap(), Some("audio"))
            .await
            .unwrap_err();
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["field"], "kind");

        let e = add(&state, png.to_str().unwrap(), Some("video_reference"))
            .await
            .unwrap_err();
        assert_eq!(e.code, "upload_mime_not_supported");
        assert_eq!(e.details.unwrap()["mime"], "image/png");

        let mp4 = dir.path().join("a.mp4");
        tokio::fs::write(&mp4, mp4_bytes()).await.unwrap();
        let v = add(&state, mp4.to_str().unwrap(), Some("video_reference"))
            .await
            .unwrap();
        assert_eq!(v["mime"], "video/mp4");

        let txt = dir.path().join("a.txt");
        tokio::fs::write(&txt, b"hello world, not a file")
            .await
            .unwrap();
        let e = add(&state, txt.to_str().unwrap(), None).await.unwrap_err();
        assert_eq!(e.code, "upload_mime_not_supported");
        // 失败时 uploads/ 里不留文件
        let mut entries = tokio::fs::read_dir(dir.path().join("uploads"))
            .await
            .unwrap();
        let mut n = 0;
        while entries.next_entry().await.unwrap().is_some() {
            n += 1;
        }
        assert_eq!(n, 1, "只有 mp4 那次成功落盘");
    }
}
