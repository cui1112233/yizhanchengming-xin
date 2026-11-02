//! 结果下载 + 格式识别(Task 2.5)。
//!
//! 捕获管线拿到原图 / 无水印视频 URL 后,由本模块落盘到任务结果目录:
//!
//! - 图像:`download_image` 40s 超时、带 `Referer: https://www.doubao.com/`、超过 100MB 拒绝;
//!   URL 往往**没有扩展名**、MIME 也不可信(`image_ori_raw.url` 可能返回 JPEG),
//!   所以写盘前用 `detect_image_format`(`image::guess_format` 魔数 + 头部尺寸)决定
//!   `image_{index+1}.{png|jpg|webp}` 的扩展名,非 PNG/JPEG/WebP 一律 `generation_download_failed`。
//! - 视频:`download_video` 先要求 URL 带 `lr=unwatermarked`(否则 `generation_watermark_not_removed`),
//!   180s 超时、流式写 `video_1.mp4`、校验 offset 4 处 `ftyp`、≤ 1GB;时长由 `mp4_duration_ms` /
//!   `mp4_file_duration_ms` 解析 `moov/mvhd`(本机没有 ffprobe)。
//! - `new_client`:**不跟随重定向**(`redirect::Policy::none()`,沈浪 `-MaximumRedirection 0`)、固定 UA。
//!
//! 错误码:HTTP 非 200 / 重定向 / 超限 / 格式不符 → `generation_download_failed`(details `{ url, status, reason }`);
//! 写盘 `ENOSPC` → `system_disk_full`,其它 IO 错误 → `system_io_error`。

use std::io::Cursor;
use std::path::{Path, PathBuf};
use std::time::Duration;

use image::ImageFormat;
use serde_json::json;
use tokio::io::AsyncWriteExt;

use crate::error::{RelayError, RelayResult};

/// 下载时伪装的 Referer(视频 URL 可能校验 Referer,见 analysis/05)。
pub const REFERER: &str = "https://www.doubao.com/";

/// 固定 UA(与豆包 2.32.5 内核 Chrome/147 同代)。
pub const USER_AGENT: &str = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36";

/// 单张图像下载总超时。
pub const IMAGE_TIMEOUT: Duration = Duration::from_secs(40);
/// 视频下载总超时。
pub const VIDEO_TIMEOUT: Duration = Duration::from_secs(180);
/// 图像文件上限 100MB(沈浪 `104857600`)。
pub const IMAGE_MAX_BYTES: u64 = 100 * 1024 * 1024;
/// 视频文件上限 1GB(沈浪 `1073741824`)。
pub const VIDEO_MAX_BYTES: u64 = 1024 * 1024 * 1024;
/// 视频文件下限:沈浪判定 `Length <= 12` 为无效(至少要能放下 `ftyp` box 头)。
pub const VIDEO_MIN_BYTES: u64 = 13;

/// 落盘结果。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Downloaded {
    /// 写入的绝对路径(`dest_dir/image_{n}.{ext}` 或 `dest_dir/video_1.mp4`)。
    pub path: PathBuf,
    /// 按内容判定的 MIME(`image/png` / `image/jpeg` / `image/webp` / `video/mp4`)。
    pub mime: String,
    /// 文件字节数。
    pub bytes: u64,
    /// 图像宽(视频为 `None`)。
    pub width: Option<u32>,
    /// 图像高(视频为 `None`)。
    pub height: Option<u32>,
}

/// 按字节内容识别图像格式:返回 `(mime, 扩展名, (宽, 高))`。
///
/// 只接受 PNG / JPEG / WebP;其它格式、无法识别的魔数、头部损坏(读不出尺寸)
/// 都返回 `generation_download_failed`(details.reason 说明原因)。
pub fn detect_image_format(bytes: &[u8]) -> RelayResult<(&'static str, &'static str, (u32, u32))> {
    let format = image::guess_format(bytes)
        .map_err(|_| download_failed(None, None, "unknown_image_format"))?;
    let (mime, ext) = match format {
        ImageFormat::Png => ("image/png", "png"),
        ImageFormat::Jpeg => ("image/jpeg", "jpg"),
        ImageFormat::WebP => ("image/webp", "webp"),
        _ => return Err(download_failed(None, None, "unsupported_image_format")),
    };
    let mut reader = image::ImageReader::new(Cursor::new(bytes));
    reader.set_format(format);
    let (width, height) = reader
        .into_dimensions()
        .map_err(|_| download_failed(None, None, "corrupt_image_header"))?;
    Ok((mime, ext, (width, height)))
}

/// `generation_download_failed`:details 为 11-error-codes 规定的 `{ url, status }` 外加 `reason`。
fn download_failed(url: Option<&str>, status: Option<u16>, reason: &str) -> RelayError {
    RelayError::new(
        "generation_download_failed",
        "结果下载失败,请稍后重试或重新提交任务",
    )
    .with_details(json!({ "url": url, "status": status, "reason": reason }))
}

/// 网络层错误(连接失败 / 超时 / 读 body 中断)→ `generation_download_failed`,status 取响应码(有的话)。
fn transport_error(url: &str, status: Option<u16>, e: &reqwest::Error) -> RelayError {
    let reason = if e.is_timeout() {
        "timeout"
    } else if e.is_connect() {
        "connect_failed"
    } else {
        "transport_error"
    };
    let mut err = download_failed(
        Some(url),
        status.or_else(|| e.status().map(|s| s.as_u16())),
        reason,
    );
    if let Some(serde_json::Value::Object(map)) = err.details.as_mut() {
        map.insert("error".into(), json!(e.to_string()));
    }
    err
}

/// 写盘错误:`ENOSPC` → `system_disk_full`,其它走 `From<io::Error>`(`system_io_error`)。
fn write_error(e: std::io::Error, path: &Path) -> RelayError {
    if e.kind() == std::io::ErrorKind::StorageFull {
        RelayError::new("system_disk_full", "磁盘空间不足,请清理磁盘后重试")
            .with_details(json!({ "path": path.display().to_string() }))
    } else {
        RelayError::from(e)
    }
}

/// 发起 GET(带 Referer、按请求超时),非 200 一律 `generation_download_failed { url, status }`。
async fn get_ok(
    client: &reqwest::Client,
    url: &str,
    timeout: Duration,
) -> RelayResult<reqwest::Response> {
    let resp = client
        .get(url)
        .header(reqwest::header::REFERER, REFERER)
        .timeout(timeout)
        .send()
        .await
        .map_err(|e| transport_error(url, None, &e))?;
    let status = resp.status().as_u16();
    if status != 200 {
        let reason = if resp.status().is_redirection() {
            "redirect"
        } else {
            "http_status"
        };
        return Err(download_failed(Some(url), Some(status), reason));
    }
    Ok(resp)
}

/// `Content-Length` 已超限则不读 body 直接拒绝。
fn check_declared_length(url: &str, resp: &reqwest::Response, max_bytes: u64) -> RelayResult<()> {
    match resp.content_length() {
        Some(len) if len > max_bytes => Err(download_failed(Some(url), Some(200), "too_large")),
        _ => Ok(()),
    }
}

/// 从响应里取下一块;超限 / 传输错误都转成 `RelayError`。
async fn next_chunk(
    url: &str,
    resp: &mut reqwest::Response,
    received: u64,
    max_bytes: u64,
) -> RelayResult<Option<Vec<u8>>> {
    let chunk = resp
        .chunk()
        .await
        .map_err(|e| transport_error(url, Some(200), &e))?;
    if let Some(c) = &chunk {
        if received + c.len() as u64 > max_bytes {
            return Err(download_failed(Some(url), Some(200), "too_large"));
        }
    }
    Ok(chunk.map(|c| c.to_vec()))
}

/// 下载一张原图到 `dest_dir/image_{index+1}.{ext}`(`index` 从 0 起)。
///
/// 40s 总超时、带 Referer、`Content-Length` 或实际字节超过 100MB 都拒绝;
/// 非 200(含 3xx,因为客户端不跟随重定向)→ `generation_download_failed { url, status }`。
pub async fn download_image(
    client: &reqwest::Client,
    url: &str,
    dest_dir: &Path,
    index: u32,
) -> RelayResult<Downloaded> {
    download_image_with_limit(client, url, dest_dir, index, IMAGE_MAX_BYTES).await
}

/// `download_image` 的可配上限版本(测试用小上限验证流式截断)。
pub async fn download_image_with_limit(
    client: &reqwest::Client,
    url: &str,
    dest_dir: &Path,
    index: u32,
    max_bytes: u64,
) -> RelayResult<Downloaded> {
    let mut resp = get_ok(client, url, IMAGE_TIMEOUT).await?;
    check_declared_length(url, &resp, max_bytes)?;

    // 图像 ≤ 100MB,先收进内存再按内容定扩展名,避免写盘后改名
    let mut body: Vec<u8> = Vec::with_capacity(resp.content_length().unwrap_or(0) as usize);
    while let Some(chunk) = next_chunk(url, &mut resp, body.len() as u64, max_bytes).await? {
        body.extend_from_slice(&chunk);
    }

    let (mime, ext, (width, height)) = detect_image_format(&body).map_err(|e| {
        let reason = e
            .details
            .as_ref()
            .and_then(|d| d["reason"].as_str())
            .unwrap_or("not_image")
            .to_string();
        download_failed(Some(url), Some(200), &reason)
    })?;

    tokio::fs::create_dir_all(dest_dir)
        .await
        .map_err(|e| write_error(e, dest_dir))?;
    let path = dest_dir.join(format!("image_{}.{ext}", index + 1));
    tokio::fs::write(&path, &body)
        .await
        .map_err(|e| write_error(e, &path))?;

    tracing::debug!(
        url_host = url::Url::parse(url).ok().and_then(|u| u.host_str().map(String::from)),
        path = %path.display(),
        mime,
        bytes = body.len(),
        width,
        height,
        "原图已下载"
    );
    Ok(Downloaded {
        path,
        mime: mime.to_string(),
        bytes: body.len() as u64,
        width: Some(width),
        height: Some(height),
    })
}

/// 下载无水印视频到 `dest_dir/video_1.mp4`:180s 超时、流式写盘、校验 `ftyp`、13B ≤ 大小 ≤ 1GB。
///
/// URL 不含 `lr=unwatermarked` → `generation_watermark_not_removed`(不发请求)。
/// 校验失败时删除已写的文件。
pub async fn download_video(
    client: &reqwest::Client,
    url: &str,
    dest_dir: &Path,
) -> RelayResult<Downloaded> {
    download_video_with_limit(client, url, dest_dir, VIDEO_MAX_BYTES).await
}

/// `download_video` 的可配上限版本(测试用)。
pub async fn download_video_with_limit(
    client: &reqwest::Client,
    url: &str,
    dest_dir: &Path,
    max_bytes: u64,
) -> RelayResult<Downloaded> {
    // 沈浪:任何候选不含 lr=unwatermarked 都拒绝下载(`未收到无水印视频地址`)
    if !has_lr_unwatermarked(url) {
        return Err(watermark_not_removed(url));
    }
    let mut resp = get_ok(client, url, VIDEO_TIMEOUT).await?;
    check_declared_length(url, &resp, max_bytes)?;

    tokio::fs::create_dir_all(dest_dir)
        .await
        .map_err(|e| write_error(e, dest_dir))?;
    let path = dest_dir.join("video_1.mp4");

    match stream_video_to(url, &mut resp, &path, max_bytes).await {
        Ok(bytes) => {
            tracing::debug!(
                url_host = url::Url::parse(url).ok().and_then(|u| u.host_str().map(String::from)),
                path = %path.display(),
                bytes,
                "无水印视频已下载"
            );
            Ok(Downloaded {
                path,
                mime: "video/mp4".to_string(),
                bytes,
                width: None,
                height: None,
            })
        }
        Err(e) => {
            // 任何失败都不留半成品
            let _ = tokio::fs::remove_file(&path).await;
            Err(e)
        }
    }
}

/// 流式写视频文件并校验:前 12 字节中 offset 4 处必须是 `ftyp`;总大小在 `[VIDEO_MIN_BYTES, max_bytes]`。
async fn stream_video_to(
    url: &str,
    resp: &mut reqwest::Response,
    path: &Path,
    max_bytes: u64,
) -> RelayResult<u64> {
    let mut file = tokio::fs::File::create(path)
        .await
        .map_err(|e| write_error(e, path))?;
    let mut received: u64 = 0;
    let mut header: Vec<u8> = Vec::with_capacity(12);
    let mut header_checked = false;

    while let Some(chunk) = next_chunk(url, resp, received, max_bytes).await? {
        if !header_checked {
            let need = 12 - header.len();
            header.extend_from_slice(&chunk[..chunk.len().min(need)]);
            if header.len() >= 12 {
                if &header[4..8] != b"ftyp" {
                    return Err(download_failed(Some(url), Some(200), "not_mp4"));
                }
                header_checked = true;
            }
        }
        file.write_all(&chunk)
            .await
            .map_err(|e| write_error(e, path))?;
        received += chunk.len() as u64;
    }
    file.flush().await.map_err(|e| write_error(e, path))?;
    drop(file);

    if received < VIDEO_MIN_BYTES {
        return Err(download_failed(Some(url), Some(200), "too_small"));
    }
    if !header_checked {
        return Err(download_failed(Some(url), Some(200), "not_mp4"));
    }
    Ok(received)
}

/// URL 的 query 里有 `lr=unwatermarked`(沈浪 `[?&]lr=unwatermarked(?:&|$)`)。
pub fn has_lr_unwatermarked(url: &str) -> bool {
    url::Url::parse(url)
        .map(|u| {
            u.query_pairs()
                .any(|(k, v)| k == "lr" && v == "unwatermarked")
        })
        .unwrap_or(false)
}

/// `generation_watermark_not_removed`:拿到的视频地址不是无水印版本(沈浪 `未收到无水印视频地址`)。
pub fn watermark_not_removed(url: &str) -> RelayError {
    RelayError::new(
        "generation_watermark_not_removed",
        "未获取到无水印视频地址,请稍后重试或重新提交任务",
    )
    .with_details(json!({ "url": url }))
}

/// moov 超过这个大小不读(正常几十 KB;防坏文件把整个视频读进内存)。
const MAX_MOOV_BYTES: u64 = 64 * 1024 * 1024;

/// 读一个 box 头:返回 `(type, header_len, total_len)`;`total_len=None` 表示延伸到末尾(size=0)。
fn box_header(buf: &[u8]) -> Option<([u8; 4], u64, Option<u64>)> {
    if buf.len() < 8 {
        return None;
    }
    let size = u32::from_be_bytes(buf[0..4].try_into().ok()?) as u64;
    let kind: [u8; 4] = buf[4..8].try_into().ok()?;
    match size {
        0 => Some((kind, 8, None)),
        1 => {
            if buf.len() < 16 {
                return None;
            }
            let large = u64::from_be_bytes(buf[8..16].try_into().ok()?);
            (large >= 16).then_some((kind, 16, Some(large)))
        }
        s if s >= 8 => Some((kind, 8, Some(s))),
        _ => None,
    }
}

/// 在一串 box 里找第一个 `kind`,返回它的内容(不含头)。
fn find_box<'a>(mut buf: &'a [u8], kind: &[u8; 4]) -> Option<&'a [u8]> {
    while let Some((k, header, total)) = box_header(buf) {
        let total = match total {
            Some(t) => usize::try_from(t).ok()?,
            None => buf.len(),
        };
        if total > buf.len() {
            return None;
        }
        if &k == kind {
            return Some(&buf[header as usize..total]);
        }
        buf = &buf[total..];
    }
    None
}

/// moov 内容 → mvhd → `duration * 1000 / timescale`。
fn mvhd_duration_ms(moov: &[u8]) -> Option<u64> {
    let mvhd = find_box(moov, b"mvhd")?;
    let version = *mvhd.first()?;
    let (timescale, duration) = if version == 1 {
        let ts = u32::from_be_bytes(mvhd.get(20..24)?.try_into().ok()?) as u64;
        let d = u64::from_be_bytes(mvhd.get(24..32)?.try_into().ok()?);
        (ts, (d != u64::MAX).then_some(d)?)
    } else {
        let ts = u32::from_be_bytes(mvhd.get(12..16)?.try_into().ok()?) as u64;
        let d = u32::from_be_bytes(mvhd.get(16..20)?.try_into().ok()?);
        (ts, (d != u32::MAX).then_some(d as u64)?)
    };
    if timescale == 0 {
        return None;
    }
    Some((duration as u128 * 1000 / timescale as u128) as u64)
}

/// 遍历一串 box,依次回调每个 `(type, 内容)`;遇到坏头 / 越界即停。
fn for_each_box<'a>(mut buf: &'a [u8], mut f: impl FnMut(&[u8; 4], &'a [u8])) {
    while let Some((k, header, total)) = box_header(buf) {
        let total = match total {
            Some(t) => match usize::try_from(t) {
                Ok(t) => t,
                Err(_) => return,
            },
            None => buf.len(),
        };
        if total > buf.len() {
            return;
        }
        f(&k, &buf[header as usize..total]);
        buf = &buf[total..];
    }
}

/// trak 内容 → tkhd 末尾 8 字节的 16.16 定点宽高;任一为 0(音轨)返回 None。
fn tkhd_dimensions(trak: &[u8]) -> Option<(u32, u32)> {
    let tkhd = find_box(trak, b"tkhd")?;
    if tkhd.len() < 84 {
        return None;
    }
    let tail = &tkhd[tkhd.len() - 8..];
    let w = u32::from_be_bytes(tail[0..4].try_into().ok()?) >> 16;
    let h = u32::from_be_bytes(tail[4..8].try_into().ok()?) >> 16;
    (w > 0 && h > 0).then_some((w, h))
}

/// trak 是否视频轨:`mdia/hdlr` 的 handler_type == `vide`。
fn trak_is_video(trak: &[u8]) -> bool {
    find_box(trak, b"mdia")
        .and_then(|mdia| find_box(mdia, b"hdlr"))
        .and_then(|hdlr| hdlr.get(8..12))
        .is_some_and(|t| t == b"vide")
}

/// moov 内容 → 视频宽高:优先 `hdlr=vide` 的 trak,否则第一个宽高非 0 的 trak。
fn moov_dimensions(moov: &[u8]) -> Option<(u32, u32)> {
    let mut video = None;
    let mut any = None;
    for_each_box(moov, |kind, body| {
        if kind != b"trak" {
            return;
        }
        if let Some(dim) = tkhd_dimensions(body) {
            if video.is_none() && trak_is_video(body) {
                video = Some(dim);
            }
            if any.is_none() {
                any = Some(dim);
            }
        }
    });
    video.or(any)
}

/// mp4 元信息:`mvhd` 时长 + 视频轨 `tkhd` 宽高;各字段解析不出为 None。
#[derive(Debug, Clone, Copy, Default, PartialEq, Eq)]
pub struct Mp4Info {
    pub duration_ms: Option<u64>,
    pub width: Option<u32>,
    pub height: Option<u32>,
}

fn moov_info(moov: &[u8]) -> Mp4Info {
    let dims = moov_dimensions(moov);
    Mp4Info {
        duration_ms: mvhd_duration_ms(moov),
        width: dims.map(|d| d.0),
        height: dims.map(|d| d.1),
    }
}

/// 从整个 mp4 字节里解析时长(毫秒):顶层 `moov` → `mvhd`(v0 32 位 / v1 64 位)。
/// 本机没有 ffprobe,结果文件的 `duration_ms` 靠它;解析不出返回 None。
pub fn mp4_duration_ms(bytes: &[u8]) -> Option<u64> {
    mvhd_duration_ms(find_box(bytes, b"moov")?)
}

/// 从整个 mp4 字节里解析 [`Mp4Info`];没有 moov 返回 None。
pub fn mp4_info(bytes: &[u8]) -> Option<Mp4Info> {
    Some(moov_info(find_box(bytes, b"moov")?))
}

/// 对落盘的 mp4 只读顶层 box 头、跳过 mdat,读出 moov 后解析时长与宽高;任何 IO / 格式问题返回 None。
/// 偏移累加全部 `checked_add`,坏 box(超大 largesize)不会回绕成死循环。
pub async fn mp4_file_info(path: &Path) -> Option<Mp4Info> {
    let path = path.to_path_buf();
    tokio::task::spawn_blocking(move || -> Option<Mp4Info> {
        use std::io::{Read, Seek, SeekFrom};
        let mut f = std::fs::File::open(&path).ok()?;
        let len = f.metadata().ok()?.len();
        let mut offset = 0u64;
        while offset.checked_add(8)? <= len {
            f.seek(SeekFrom::Start(offset)).ok()?;
            let mut head = [0u8; 16];
            let n = f.read(&mut head).ok()?;
            let (kind, header, total) = box_header(&head[..n])?;
            let total = total.unwrap_or(len - offset);
            if total < header {
                return None;
            }
            let end = offset.checked_add(total)?;
            if end > len {
                return None;
            }
            if &kind == b"moov" {
                let body_len = total - header;
                if body_len > MAX_MOOV_BYTES {
                    return None;
                }
                f.seek(SeekFrom::Start(offset + header)).ok()?;
                let mut body = vec![0u8; body_len as usize];
                f.read_exact(&mut body).ok()?;
                return Some(moov_info(&body));
            }
            offset = end;
        }
        None
    })
    .await
    .ok()
    .flatten()
}

/// [`mp4_file_info`] 的时长部分。
pub async fn mp4_file_duration_ms(path: &Path) -> Option<u64> {
    mp4_file_info(path).await.and_then(|i| i.duration_ms)
}

/// 测试用合成 trak:`tkhd`(v0,84 字节,末尾 16.16 宽高)+ `mdia/hdlr`(handler_type)。
#[cfg(test)]
fn synthetic_trak(handler: &[u8; 4], width: u32, height: u32) -> Vec<u8> {
    fn boxed(kind: &[u8; 4], body: &[u8]) -> Vec<u8> {
        let mut v = ((body.len() + 8) as u32).to_be_bytes().to_vec();
        v.extend_from_slice(kind);
        v.extend_from_slice(body);
        v
    }
    let mut tkhd = vec![0u8; 76];
    tkhd.extend_from_slice(&(width << 16).to_be_bytes());
    tkhd.extend_from_slice(&(height << 16).to_be_bytes());
    let mut hdlr = vec![0u8; 8];
    hdlr.extend_from_slice(handler);
    hdlr.extend_from_slice(&[0u8; 13]);
    let mdia = boxed(b"mdia", &boxed(b"hdlr", &hdlr));
    let mut trak = boxed(b"tkhd", &tkhd);
    trak.extend(mdia);
    boxed(b"trak", &trak)
}

/// 测试用合成 mp4:`ftyp` + (`mdat`) + `moov{mvhd, trak(soun 0x0), trak(vide 1920x1080)}`;`v1` 用 64 位 mvhd,`moov_last` 把 moov 放在 mdat 之后。
#[cfg(test)]
pub(crate) fn synthetic_mp4(timescale: u32, duration: u64, v1: bool, moov_last: bool) -> Vec<u8> {
    fn boxed(kind: &[u8; 4], body: &[u8]) -> Vec<u8> {
        let mut v = ((body.len() + 8) as u32).to_be_bytes().to_vec();
        v.extend_from_slice(kind);
        v.extend_from_slice(body);
        v
    }
    let mut mvhd = Vec::new();
    if v1 {
        mvhd.extend_from_slice(&[1, 0, 0, 0]);
        mvhd.extend_from_slice(&[0u8; 16]); // creation + modification
        mvhd.extend_from_slice(&timescale.to_be_bytes());
        mvhd.extend_from_slice(&duration.to_be_bytes());
    } else {
        mvhd.extend_from_slice(&[0, 0, 0, 0]);
        mvhd.extend_from_slice(&[0u8; 8]);
        mvhd.extend_from_slice(&timescale.to_be_bytes());
        mvhd.extend_from_slice(&(duration as u32).to_be_bytes());
    }
    mvhd.extend_from_slice(&[0u8; 80]); // rate/volume/reserved/matrix/pre_defined/next_track_ID
    let mut moov_body = boxed(b"mvhd", &mvhd);
    moov_body.extend(synthetic_trak(b"soun", 0, 0));
    moov_body.extend(synthetic_trak(b"vide", 1920, 1080));
    let moov = boxed(b"moov", &moov_body);
    let ftyp = boxed(b"ftyp", b"isom\0\0\x02\0isomiso2mp41");
    let mdat = boxed(b"mdat", &[0x5A; 16]);
    let mut out = ftyp;
    if moov_last {
        out.extend(mdat);
        out.extend(moov);
    } else {
        out.extend(moov);
        out.extend(mdat);
    }
    out
}

/// 下载专用 HTTP 客户端:不跟随重定向、固定 UA。超时按请求单独设置。
pub fn new_client() -> reqwest::Client {
    reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .user_agent(USER_AGENT)
        .build()
        .expect("构建下载 HTTP 客户端失败")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::{Arc, Mutex};
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::net::TcpListener;

    // ---------- 图像字节工厂 ----------

    fn encode(width: u32, height: u32, format: ImageFormat) -> Vec<u8> {
        let img = image::RgbaImage::from_pixel(width, height, image::Rgba([10, 20, 30, 255]));
        let mut buf = Cursor::new(Vec::new());
        image::DynamicImage::ImageRgba8(img)
            .write_to(&mut buf, format)
            .expect("encode");
        buf.into_inner()
    }

    fn png_bytes(w: u32, h: u32) -> Vec<u8> {
        encode(w, h, ImageFormat::Png)
    }

    fn jpeg_bytes(w: u32, h: u32) -> Vec<u8> {
        encode(w, h, ImageFormat::Jpeg)
    }

    fn webp_bytes(w: u32, h: u32) -> Vec<u8> {
        encode(w, h, ImageFormat::WebP)
    }

    // ---------- detect_image_format ----------

    #[test]
    fn detect_png_jpeg_webp_with_dimensions() {
        assert_eq!(
            detect_image_format(&png_bytes(3, 2)).unwrap(),
            ("image/png", "png", (3, 2))
        );
        assert_eq!(
            detect_image_format(&jpeg_bytes(5, 7)).unwrap(),
            ("image/jpeg", "jpg", (5, 7))
        );
        assert_eq!(
            detect_image_format(&webp_bytes(4, 6)).unwrap(),
            ("image/webp", "webp", (4, 6))
        );
    }

    #[test]
    fn detect_relies_on_magic_not_extension_hint() {
        // JPEG 魔数开头的最小合法字节:识别为 jpeg,即便调用方以为是 png
        let (mime, ext, _) = detect_image_format(&jpeg_bytes(2, 2)).unwrap();
        assert_eq!((mime, ext), ("image/jpeg", "jpg"));
    }

    #[test]
    fn detect_rejects_random_bytes() {
        let junk: Vec<u8> = (0..256u32).map(|i| (i * 7 % 251) as u8).collect();
        let e = detect_image_format(&junk).unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.http_status(), 500);
        assert!(e.details.unwrap()["reason"].is_string());
    }

    #[test]
    fn detect_rejects_empty_and_html() {
        assert_eq!(
            detect_image_format(b"").unwrap_err().code,
            "generation_download_failed"
        );
        assert_eq!(
            detect_image_format(b"<html><body>403 Forbidden</body></html>")
                .unwrap_err()
                .code,
            "generation_download_failed"
        );
    }

    #[test]
    fn detect_rejects_unsupported_image_formats() {
        // GIF89a 魔数:image 能认出格式,但不在 PNG/JPEG/WebP 白名单
        let gif = b"GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;";
        let e = detect_image_format(gif).unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["reason"], "unsupported_image_format");
    }

    #[test]
    fn detect_rejects_truncated_png_header() {
        // 只有 PNG 魔数、没有 IHDR:guess_format 通过但读不出尺寸
        let mut bytes = png_bytes(2, 2);
        bytes.truncate(12);
        let e = detect_image_format(&bytes).unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
    }

    // ---------- mock HTTP server ----------

    enum Body {
        /// 一次性写完,带 Content-Length。
        Full(Vec<u8>),
        /// chunked 编码,重复写 `chunk` 共 `count` 次(不带 Content-Length)。
        Chunked { chunk: Vec<u8>, count: usize },
        /// 声明 `declared` 的 Content-Length,但只发 `bytes` 就关连接(模拟超大文件,不真发)。
        Declared { declared: u64, bytes: Vec<u8> },
    }

    struct Reply {
        status: &'static str,
        extra_headers: Vec<String>,
        body: Body,
    }

    fn ok(body: Vec<u8>) -> Reply {
        Reply {
            status: "200 OK",
            extra_headers: vec![],
            body: Body::Full(body),
        }
    }

    type Seen = Arc<Mutex<Vec<String>>>;

    /// 起一个 HTTP/1.1 mock:记录每次请求头(供断言 Referer/UA),按 `reply` 应答。
    async fn spawn_server(reply: impl Fn(&str) -> Reply + Send + Sync + 'static) -> (String, Seen) {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        let seen: Seen = Arc::new(Mutex::new(vec![]));
        let seen_in = seen.clone();
        let reply = Arc::new(reply);
        tokio::spawn(async move {
            loop {
                let Ok((mut sock, _)) = listener.accept().await else {
                    break;
                };
                let reply = reply.clone();
                let seen = seen_in.clone();
                tokio::spawn(async move {
                    let mut head = Vec::new();
                    let mut b = [0u8; 1024];
                    loop {
                        let n = sock.read(&mut b).await.unwrap_or(0);
                        if n == 0 {
                            break;
                        }
                        head.extend_from_slice(&b[..n]);
                        if head.windows(4).any(|w| w == b"\r\n\r\n") {
                            break;
                        }
                    }
                    let head = String::from_utf8_lossy(&head).to_string();
                    seen.lock().unwrap().push(head.clone());
                    let r = reply(&head);
                    let mut resp = format!("HTTP/1.1 {}\r\nConnection: close\r\n", r.status);
                    for h in &r.extra_headers {
                        resp.push_str(h);
                        resp.push_str("\r\n");
                    }
                    match r.body {
                        Body::Full(body) => {
                            resp.push_str(&format!("Content-Length: {}\r\n\r\n", body.len()));
                            let _ = sock.write_all(resp.as_bytes()).await;
                            let _ = sock.write_all(&body).await;
                        }
                        Body::Declared { declared, bytes } => {
                            resp.push_str(&format!("Content-Length: {declared}\r\n\r\n"));
                            let _ = sock.write_all(resp.as_bytes()).await;
                            let _ = sock.write_all(&bytes).await;
                        }
                        Body::Chunked { chunk, count } => {
                            resp.push_str("Transfer-Encoding: chunked\r\n\r\n");
                            let _ = sock.write_all(resp.as_bytes()).await;
                            let frame = format!("{:x}\r\n", chunk.len());
                            for _ in 0..count {
                                if sock.write_all(frame.as_bytes()).await.is_err()
                                    || sock.write_all(&chunk).await.is_err()
                                    || sock.write_all(b"\r\n").await.is_err()
                                {
                                    return;
                                }
                            }
                            let _ = sock.write_all(b"0\r\n\r\n").await;
                        }
                    }
                    let _ = sock.shutdown().await;
                });
            }
        });
        (format!("http://{addr}"), seen)
    }

    fn header_line(head: &str, name: &str) -> Option<String> {
        head.lines()
            .find(|l| {
                l.to_ascii_lowercase()
                    .starts_with(&format!("{}:", name.to_ascii_lowercase()))
            })
            .and_then(|l| l.split_once(':'))
            .map(|(_, v)| v.trim().to_string())
    }

    // ---------- new_client ----------

    #[tokio::test]
    async fn client_sends_fixed_ua_and_never_follows_redirects() {
        let (base, seen) = spawn_server(|head| {
            if head.starts_with("GET /go ") {
                Reply {
                    status: "302 Found",
                    extra_headers: vec!["Location: /final".into()],
                    body: Body::Full(vec![]),
                }
            } else {
                ok(png_bytes(1, 1))
            }
        })
        .await;
        let client = new_client();
        let resp = client.get(format!("{base}/go")).send().await.unwrap();
        assert_eq!(resp.status().as_u16(), 302, "客户端不得跟随重定向");
        let heads = seen.lock().unwrap().clone();
        assert_eq!(heads.len(), 1, "只应有一次请求(没有跟到 /final)");
        assert_eq!(
            header_line(&heads[0], "user-agent").as_deref(),
            Some(USER_AGENT)
        );
    }

    // ---------- download_image ----------

    #[tokio::test]
    async fn download_image_writes_png_with_mime_and_size() {
        let (base, seen) = spawn_server(|_| ok(png_bytes(6, 4))).await;
        let dir = tempfile::tempdir().unwrap();
        let dest = dir.path().join("results").join("t1");
        let d = download_image(&new_client(), &format!("{base}/img/abc"), &dest, 0)
            .await
            .unwrap();
        assert_eq!(d.path, dest.join("image_1.png"));
        assert_eq!(d.mime, "image/png");
        assert_eq!((d.width, d.height), (Some(6), Some(4)));
        let on_disk = std::fs::read(&d.path).unwrap();
        assert_eq!(d.bytes, on_disk.len() as u64);
        assert_eq!(on_disk, png_bytes(6, 4));
        let head = seen.lock().unwrap()[0].clone();
        assert!(head.starts_with("GET /img/abc HTTP/1.1"));
        assert_eq!(header_line(&head, "referer").as_deref(), Some(REFERER));
        assert_eq!(
            header_line(&head, "user-agent").as_deref(),
            Some(USER_AGENT)
        );
    }

    #[tokio::test]
    async fn download_image_names_file_by_detected_format_not_url() {
        // URL 以 .png 结尾但内容是 JPEG → 文件名 image_3.jpg
        let (base, _) = spawn_server(|_| ok(jpeg_bytes(2, 3))).await;
        let dir = tempfile::tempdir().unwrap();
        let d = download_image(&new_client(), &format!("{base}/x.png"), dir.path(), 2)
            .await
            .unwrap();
        assert_eq!(d.path, dir.path().join("image_3.jpg"));
        assert_eq!(d.mime, "image/jpeg");
        assert_eq!((d.width, d.height), (Some(2), Some(3)));

        let (base, _) = spawn_server(|_| ok(webp_bytes(2, 2))).await;
        let d = download_image(&new_client(), &format!("{base}/y"), dir.path(), 0)
            .await
            .unwrap();
        assert_eq!(d.path, dir.path().join("image_1.webp"));
        assert_eq!(d.mime, "image/webp");
    }

    #[tokio::test]
    async fn download_image_rejects_redirect() {
        let (base, seen) = spawn_server(|_| Reply {
            status: "302 Found",
            extra_headers: vec!["Location: http://127.0.0.1:1/elsewhere".into()],
            body: Body::Full(vec![]),
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let url = format!("{base}/img");
        let e = download_image(&new_client(), &url, dir.path(), 0)
            .await
            .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        let details = e.details.unwrap();
        assert_eq!(details["url"], url);
        assert_eq!(details["status"], 302);
        assert_eq!(seen.lock().unwrap().len(), 1);
        assert!(
            std::fs::read_dir(dir.path()).unwrap().next().is_none(),
            "不应留下文件"
        );
    }

    #[tokio::test]
    async fn download_image_rejects_non_200() {
        let (base, _) = spawn_server(|_| Reply {
            status: "403 Forbidden",
            extra_headers: vec![],
            body: Body::Full(b"denied".to_vec()),
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let e = download_image(&new_client(), &format!("{base}/img"), dir.path(), 0)
            .await
            .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["status"], 403);
    }

    #[tokio::test]
    async fn download_image_rejects_oversized_content_length_before_reading_body() {
        // 声明 101MB 的 Content-Length,实际只发很少字节:应凭头部直接拒绝
        let (base, _) = spawn_server(|_| Reply {
            status: "200 OK",
            extra_headers: vec![],
            body: Body::Declared {
                declared: 101 * 1024 * 1024,
                bytes: vec![0u8; 16],
            },
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let url = format!("{base}/big");
        let e = tokio::time::timeout(
            Duration::from_secs(5),
            download_image(&new_client(), &url, dir.path(), 0),
        )
        .await
        .expect("应立即拒绝,不等超时")
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        let details = e.details.unwrap();
        assert_eq!(details["url"], url);
        assert_eq!(details["reason"], "too_large");
        assert!(std::fs::read_dir(dir.path()).unwrap().next().is_none());
    }

    #[tokio::test]
    async fn download_image_stops_streaming_once_limit_exceeded() {
        // 没有 Content-Length(chunked),服务端会一直发;客户端超过上限必须中止而不是读完
        let (base, _) = spawn_server(|_| Reply {
            status: "200 OK",
            extra_headers: vec![],
            body: Body::Chunked {
                chunk: vec![0xAAu8; 4096],
                count: 1_000_000, // ~4GB,读完是不可能的
            },
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let e = tokio::time::timeout(
            Duration::from_secs(10),
            download_image_with_limit(
                &new_client(),
                &format!("{base}/stream"),
                dir.path(),
                0,
                64 * 1024,
            ),
        )
        .await
        .expect("超过上限应立刻中止")
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["reason"], "too_large");
        assert!(std::fs::read_dir(dir.path()).unwrap().next().is_none());
    }

    #[tokio::test]
    async fn download_image_rejects_non_image_body() {
        let (base, _) = spawn_server(|_| ok(b"<html>not an image</html>".to_vec())).await;
        let dir = tempfile::tempdir().unwrap();
        let url = format!("{base}/img");
        let e = download_image(&new_client(), &url, dir.path(), 0)
            .await
            .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        let details = e.details.unwrap();
        assert_eq!(details["url"], url);
        assert_eq!(details["status"], 200);
        assert!(
            std::fs::read_dir(dir.path()).unwrap().next().is_none(),
            "非图片不落盘"
        );
    }

    #[tokio::test]
    async fn download_image_connection_refused_is_download_failed() {
        // 绑定后立刻释放端口,确保无人监听
        let port = {
            let l = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
            l.local_addr().unwrap().port()
        };
        let dir = tempfile::tempdir().unwrap();
        let url = format!("http://127.0.0.1:{port}/img");
        let e = download_image(&new_client(), &url, dir.path(), 0)
            .await
            .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        let details = e.details.unwrap();
        assert_eq!(details["url"], url);
        assert!(details["status"].is_null());
    }

    // ---------- download_video ----------

    fn mp4_bytes(payload_len: usize) -> Vec<u8> {
        // ftyp box: size(4) 'ftyp' major_brand minor_version compatible_brands + 任意负载
        let mut v = vec![0, 0, 0, 0x18];
        v.extend_from_slice(b"ftypisom");
        v.extend_from_slice(&[0, 0, 2, 0]);
        v.extend_from_slice(b"isomiso2mp41");
        v.extend(std::iter::repeat_n(0x5Au8, payload_len));
        v
    }

    #[tokio::test]
    async fn download_video_writes_mp4() {
        let body = mp4_bytes(100_000);
        let expected = body.clone();
        let (base, seen) = spawn_server(move |_| ok(body.clone())).await;
        let dir = tempfile::tempdir().unwrap();
        let dest = dir.path().join("v");
        let d = download_video(
            &new_client(),
            &format!("{base}/v.mp4?lr=unwatermarked"),
            &dest,
        )
        .await
        .unwrap();
        assert_eq!(d.path, dest.join("video_1.mp4"));
        assert_eq!(d.mime, "video/mp4");
        assert_eq!(d.bytes, expected.len() as u64);
        assert_eq!((d.width, d.height), (None, None));
        assert_eq!(std::fs::read(&d.path).unwrap(), expected);
        let head = seen.lock().unwrap()[0].clone();
        assert!(head.starts_with("GET /v.mp4?lr=unwatermarked HTTP/1.1"));
        assert_eq!(header_line(&head, "referer").as_deref(), Some(REFERER));
    }

    #[tokio::test]
    async fn download_video_rejects_non_mp4_and_leaves_no_file() {
        let (base, _) = spawn_server(|_| ok(png_bytes(8, 8))).await;
        let dir = tempfile::tempdir().unwrap();
        let e = download_video(
            &new_client(),
            &format!("{base}/v?lr=unwatermarked"),
            dir.path(),
        )
        .await
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["reason"], "not_mp4");
        assert!(std::fs::read_dir(dir.path()).unwrap().next().is_none());
    }

    #[tokio::test]
    async fn download_video_rejects_tiny_body() {
        let (base, _) = spawn_server(|_| ok(b"\0\0\0\x08ftyp".to_vec())).await; // 8 字节
        let dir = tempfile::tempdir().unwrap();
        let e = download_video(
            &new_client(),
            &format!("{base}/v?lr=unwatermarked"),
            dir.path(),
        )
        .await
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["reason"], "too_small");
        assert!(std::fs::read_dir(dir.path()).unwrap().next().is_none());
    }

    #[tokio::test]
    async fn download_video_rejects_oversized_stream() {
        let head = mp4_bytes(0);
        let (base, _) = spawn_server(move |_| {
            let mut chunk = head.clone();
            chunk.resize(4096, 0);
            Reply {
                status: "200 OK",
                extra_headers: vec![],
                body: Body::Chunked {
                    chunk,
                    count: 1_000_000,
                },
            }
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let e = tokio::time::timeout(
            Duration::from_secs(10),
            download_video_with_limit(
                &new_client(),
                &format!("{base}/v?lr=unwatermarked"),
                dir.path(),
                64 * 1024,
            ),
        )
        .await
        .expect("超过上限应立刻中止")
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["reason"], "too_large");
        assert!(
            std::fs::read_dir(dir.path()).unwrap().next().is_none(),
            "超限后应删除半成品"
        );
    }

    #[tokio::test]
    async fn download_video_rejects_redirect() {
        let (base, _) = spawn_server(|_| Reply {
            status: "301 Moved Permanently",
            extra_headers: vec!["Location: /v2".into()],
            body: Body::Full(vec![]),
        })
        .await;
        let dir = tempfile::tempdir().unwrap();
        let e = download_video(
            &new_client(),
            &format!("{base}/v?lr=unwatermarked"),
            dir.path(),
        )
        .await
        .unwrap_err();
        assert_eq!(e.code, "generation_download_failed");
        assert_eq!(e.details.unwrap()["status"], 301);
    }

    #[tokio::test]
    async fn download_video_refuses_url_without_lr_unwatermarked() {
        let (base, seen) = spawn_server(|_| ok(mp4_bytes(100))).await;
        let dir = tempfile::tempdir().unwrap();
        for url in [
            format!("{base}/v.mp4"),
            format!("{base}/v.mp4?lr=watermarked"),
            format!("{base}/v.mp4?xlr=unwatermarked"),
        ] {
            let e = download_video(&new_client(), &url, dir.path())
                .await
                .unwrap_err();
            assert_eq!(e.code, "generation_watermark_not_removed", "{url}");
            assert_eq!(e.details.unwrap()["url"], url.as_str());
        }
        assert!(
            seen.lock().unwrap().is_empty(),
            "未去水印的地址不应发起请求"
        );
        assert!(std::fs::read_dir(dir.path()).unwrap().next().is_none());
    }

    #[test]
    fn lr_unwatermarked_query_detection() {
        assert!(has_lr_unwatermarked("https://a.com/v.mp4?lr=unwatermarked"));
        assert!(has_lr_unwatermarked(
            "https://a.com/v.mp4?x=1&lr=unwatermarked&y=2"
        ));
        assert!(!has_lr_unwatermarked("https://a.com/v.mp4"));
        assert!(!has_lr_unwatermarked("https://a.com/v.mp4?lr=watermarked"));
        assert!(!has_lr_unwatermarked(
            "https://a.com/v.mp4?flr=unwatermarked"
        ));
        assert!(!has_lr_unwatermarked("https://a.com/lr=unwatermarked"));
        assert!(!has_lr_unwatermarked("not a url"));
    }

    // ---------- mp4 时长 ----------

    #[test]
    fn mp4_duration_from_mvhd_v0_and_v1() {
        let v0 = synthetic_mp4(1000, 5000, false, false);
        assert!(v0.len() <= 1024);
        assert_eq!(mp4_duration_ms(&v0), Some(5000));
        // 90kHz,64 位 duration
        assert_eq!(
            mp4_duration_ms(&synthetic_mp4(90_000, 450_000, true, false)),
            Some(5000)
        );
        // moov 在 mdat 之后(非 faststart)
        assert_eq!(
            mp4_duration_ms(&synthetic_mp4(600, 7_300, false, true)),
            Some(12_166)
        );
    }

    #[test]
    fn mp4_duration_rejects_garbage() {
        assert_eq!(mp4_duration_ms(b""), None);
        assert_eq!(mp4_duration_ms(&mp4_bytes(64)), None, "没有 moov");
        let full = synthetic_mp4(1000, 5000, false, true);
        assert_eq!(
            mp4_duration_ms(&full[..full.len() - 6]),
            None,
            "moov 被截断"
        );
        assert_eq!(mp4_duration_ms(&synthetic_mp4(0, 5000, false, false)), None);
        assert_eq!(
            mp4_duration_ms(&synthetic_mp4(1000, u32::MAX as u64, false, false)),
            None,
            "未知时长"
        );
        // box size 小于 8 的坏数据不死循环
        let mut bad = vec![0, 0, 0, 4];
        bad.extend_from_slice(b"moov");
        assert_eq!(mp4_duration_ms(&bad), None);
    }

    #[test]
    fn mp4_info_reads_video_track_dimensions() {
        let info = mp4_info(&synthetic_mp4(1000, 5000, false, true)).unwrap();
        assert_eq!(
            info,
            Mp4Info {
                duration_ms: Some(5000),
                width: Some(1920),
                height: Some(1080)
            },
            "跳过前面宽高为 0 的音轨,取 vide 轨"
        );
        // 只有 mvhd、没有 trak:时长有、宽高 None
        let mut mvhd = vec![0u8; 12];
        mvhd.extend_from_slice(&1000u32.to_be_bytes());
        mvhd.extend_from_slice(&3000u32.to_be_bytes());
        mvhd.extend_from_slice(&[0u8; 80]);
        let mut moov = ((mvhd.len() + 16) as u32).to_be_bytes().to_vec();
        moov.extend_from_slice(b"moov");
        moov.extend_from_slice(&((mvhd.len() + 8) as u32).to_be_bytes());
        moov.extend_from_slice(b"mvhd");
        moov.extend_from_slice(&mvhd);
        assert_eq!(
            mp4_info(&moov),
            Some(Mp4Info {
                duration_ms: Some(3000),
                width: None,
                height: None
            })
        );
    }

    #[tokio::test]
    async fn mp4_file_info_rejects_overflowing_largesize() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path().join("video_1.mp4");
        // box1: 正常 ftyp;box2: size=1 + largesize = u64::MAX - 7,offset + total 会回绕
        let mut bytes = synthetic_mp4(1000, 4_000, false, false)[..32].to_vec();
        let ftyp_len = u32::from_be_bytes(bytes[0..4].try_into().unwrap()) as usize;
        bytes.truncate(ftyp_len);
        bytes.extend_from_slice(&1u32.to_be_bytes());
        bytes.extend_from_slice(b"mdat");
        bytes.extend_from_slice(&(u64::MAX - 7).to_be_bytes());
        bytes.extend_from_slice(&[0u8; 64]);
        std::fs::write(&p, &bytes).unwrap();
        let r = tokio::time::timeout(std::time::Duration::from_secs(5), mp4_file_info(&p))
            .await
            .expect("坏 largesize 不得死循环");
        assert_eq!(r, None);
    }

    #[tokio::test]
    async fn mp4_file_duration_reads_only_box_headers() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path().join("video_1.mp4");
        std::fs::write(&p, synthetic_mp4(1000, 4_000, false, true)).unwrap();
        assert_eq!(mp4_file_duration_ms(&p).await, Some(4_000));
        let info = mp4_file_info(&p).await.unwrap();
        assert_eq!((info.width, info.height), (Some(1920), Some(1080)));
        std::fs::write(&p, mp4_bytes(32)).unwrap();
        assert_eq!(mp4_file_duration_ms(&p).await, None);
        assert_eq!(
            mp4_file_duration_ms(&dir.path().join("missing.mp4")).await,
            None
        );
    }

    #[test]
    fn error_message_has_no_trailing_punctuation() {
        let e = detect_image_format(b"junk").unwrap_err();
        let last = e.message.chars().last().unwrap();
        assert!(!matches!(
            last,
            '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'
        ));
    }
}
