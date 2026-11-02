//! 视频无水印 token 本地解密(`docs/analysis/06-video-decryption.md`,Rust 移植)。
//!
//! 对标 `scripts/doubao-video-decoder.js`(`shenlangDoubaoVideoDecoder`)。
//! **密钥派生以反混淆后的原始 JS 为准**(用 Node 跑原脚本做 oracle 验证过),
//! analysis/06 的伪代码与原脚本有出入:
//!
//! ```text
//! seed  = base64Flexible(key_seed)
//! A     = SHA512( SHA512(seed[..32]) || SALT )      → key_a = A[0..16], iv_a = A[16..32]
//! P     = SHA512( SHA512(seed)       || PEPPER )    → key_p = P[0..16], iv_p = P[16..32]
//! payload = base64Flexible(token)
//! 有 magic A8 00 01 00:候选依次 (payload[4..], key_p, iv_p) → (key_a, iv_a) → (iv_a, key_a)
//!     → payload.len() > 0x24 时再加 (payload[0x24..], key_a, payload[0x14..0x24]) → (key_a, iv_a)
//! 无 magic:(payload, key_a, iv_a)
//! ```
//!
//! 每个候选 AES-128-CBC(NoPadding)解密后,手工 PKCS7 去填充(最多两层)并要求结果是
//! 可打印 ASCII 且为合法 https URL。

use std::sync::LazyLock;

use aes::Aes128;
use base64::alphabet;
use base64::engine::{DecodePaddingMode, GeneralPurpose, GeneralPurposeConfig};
use base64::Engine;
use cbc::cipher::block_padding::NoPadding;
use cbc::cipher::{BlockDecryptMut, KeyIvInit};
use regex::Regex;
use serde_json::Value;
use sha2::{Digest, Sha512};
use url::Url;

/// 原版解密脚本(`window.shenlangDoubaoVideoDecoder`),Rust 解密失败时作兜底注入。
pub const VIDEO_DECODER_SCRIPT: &str = include_str!("scripts/doubao-video-decoder.js");

/// `decode` 失败原因:响应里找不到 `main_url`/`play_url`。
pub const FALLBACK_MAIN_URL_MISSING: &str = "fallback_main_url_missing";
/// `decode` 失败原因:token 解不出合法 https URL。
pub const FALLBACK_URL_DECODE_FAILED: &str = "fallback_url_decode_failed";

/// 16 字节盐(hex)。
const SALT_HEX: &str = "4dd4c2e6b83162090e52b3c7a6733ba4";
/// 64 字节 pepper(base64)。
const EXTRA_PEPPER: &str =
    "TdTC5rgxYgkOUrPHpnM7pByyRiuCmrWKGWs521cXdST0m69/COjWjSanLjfBqVovHwWlGJKu8pSXMrYqOKrdWA==";
/// 加密 payload 的魔术头(base64 以 `qAAB` 开头)。
const MAGIC: [u8; 4] = [0xA8, 0x00, 0x01, 0x00];

/// 对标浏览器 `atob`:不校验尾部多余 bit、填充有无皆可。
const ATOB: GeneralPurpose = GeneralPurpose::new(
    &alphabet::STANDARD,
    GeneralPurposeConfig::new()
        .with_decode_allow_trailing_bits(true)
        .with_decode_padding_mode(DecodePaddingMode::Indifferent),
);

static KEY_SEED_QUERY_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r#"(?i)(?:^|[?&])key_seed=([^&"'<>\\\s]+)"#).unwrap());
static KEY_SEED_JSON_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r#"(?i)["']key_seed["']\s*:\s*["']([^"']+)"#).unwrap());
static VIDEO_CDN_HOST_RE: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(
        r"(?:^|\.)(?:doubaovideo\.com|bytevcloudcn\.com|bytevcloud\.com|volces\.com|volcvideo\.com|bytevideo\.com|ixigua\.com|365yg\.com|v[0-9]+-vdl\.doubao\.com)$",
    )
    .unwrap()
});

// ---------------------------------------------------------------------------
// 基础工具
// ---------------------------------------------------------------------------

/// 容错 base64 解码:依次尝试原文、`$@#`→`_/.`、`$@#`→`+/=` 三种变体,
/// 每种先补 `=`、再把 URL-safe 的 `-_` 换成 `+/`。
pub fn base64_decode_flexible(input: &str) -> Option<Vec<u8>> {
    let raw: String = input.chars().filter(|c| !c.is_ascii_whitespace()).collect();
    if raw.is_empty() {
        return None;
    }
    let variants = [
        raw.clone(),
        raw.chars()
            .map(|c| match c {
                '$' => '_',
                '@' => '/',
                '#' => '.',
                other => other,
            })
            .collect::<String>(),
        raw.chars()
            .map(|c| match c {
                '$' => '+',
                '@' => '/',
                '#' => '=',
                other => other,
            })
            .collect::<String>(),
    ];
    let mut tried: Vec<&str> = Vec::with_capacity(3);
    for candidate in &variants {
        if tried.contains(&candidate.as_str()) {
            continue;
        }
        tried.push(candidate);
        let mut normalized = candidate.replace('-', "+").replace('_', "/");
        let pad = (4 - normalized.len() % 4) % 4;
        normalized.extend(std::iter::repeat_n('=', pad));
        if let Ok(bytes) = ATOB.decode(&normalized) {
            return Some(bytes);
        }
    }
    None
}

/// 全部字节必须是可打印 ASCII(0x20..=0x7E)或 `\t` `\n` `\r`;空输入返回 `None`。
pub fn bytes_to_printable(bytes: &[u8]) -> Option<String> {
    if bytes.is_empty()
        || bytes
            .iter()
            .any(|&b| !matches!(b, 0x09 | 0x0A | 0x0D | 0x20..=0x7E))
    {
        return None;
    }
    String::from_utf8(bytes.to_vec()).ok()
}

/// `https` 且不带用户名/密码的合法 URL。
pub fn is_valid_https_url(s: &str) -> bool {
    match Url::parse(s) {
        Ok(u) => u.scheme() == "https" && u.username().is_empty() && u.password().is_none(),
        Err(_) => false,
    }
}

/// 手工 PKCS7 去填充;填充不合法时原样返回。
pub fn pkcs7_unpad(bytes: &[u8]) -> &[u8] {
    let Some(&pad) = bytes.last() else {
        return bytes;
    };
    let pad = pad as usize;
    if pad == 0 || pad > 16 || pad > bytes.len() {
        return bytes;
    }
    let (head, tail) = bytes.split_at(bytes.len() - pad);
    if tail.iter().all(|&b| b as usize == pad) {
        head
    } else {
        bytes
    }
}

/// AES-128-CBC 解密,不处理填充(调用方自行 `pkcs7_unpad`)。
/// 密文长度必须是 16 的倍数且非空,key/iv 必须 16 字节。
pub fn aes_cbc_decrypt(ciphertext: &[u8], key: &[u8], iv: &[u8]) -> Option<Vec<u8>> {
    if ciphertext.is_empty()
        || !ciphertext.len().is_multiple_of(16)
        || key.len() != 16
        || iv.len() != 16
    {
        return None;
    }
    let decryptor = cbc::Decryptor::<Aes128>::new_from_slices(key, iv).ok()?;
    let mut buf = ciphertext.to_vec();
    let n = decryptor
        .decrypt_padded_mut::<NoPadding>(&mut buf)
        .ok()?
        .len();
    buf.truncate(n);
    Some(buf)
}

// ---------------------------------------------------------------------------
// key_seed / video entry 提取
// ---------------------------------------------------------------------------

/// 从字符串里找 `key_seed`:`?key_seed=...`/`&key_seed=...` 或内嵌 `"key_seed":"..."`,值做 percent 解码。
pub fn extract_key_seed_from_str(s: &str) -> Option<String> {
    let captured = KEY_SEED_QUERY_RE
        .captures(s)
        .or_else(|| KEY_SEED_JSON_RE.captures(s))?;
    percent_decode(&captured[1]).filter(|v| !v.is_empty())
}

/// 递归(≤10 层)找 `key_seed`:对象直接字段优先,否则深入子节点;字符串按 [`extract_key_seed_from_str`]。
pub fn extract_key_seed(value: &Value) -> Option<String> {
    fn walk(value: &Value, depth: u32) -> Option<String> {
        if depth > 10 {
            return None;
        }
        match value {
            Value::String(s) => extract_key_seed_from_str(s),
            Value::Object(map) => {
                if let Some(Value::String(s)) = map.get("key_seed") {
                    let trimmed = s.trim();
                    if !trimmed.is_empty() {
                        return Some(trimmed.to_string());
                    }
                }
                map.values().find_map(|child| walk(child, depth + 1))
            }
            Value::Array(items) => items.iter().find_map(|child| walk(child, depth + 1)),
            _ => None,
        }
    }
    walk(value, 0)
}

/// `response.video_info` / `response.data.video_info` / `response` 三选一,再剥一层 `data`。
pub fn extract_video_info(response: &Value) -> &Value {
    let info = response
        .get("video_info")
        .filter(|v| truthy(v))
        .or_else(|| {
            response
                .get("data")
                .and_then(|d| d.get("video_info"))
                .filter(|v| truthy(v))
        })
        .unwrap_or(response);
    let data = info.get("data").filter(|v| truthy(v)).unwrap_or(info);
    if data.is_object() || data.is_array() {
        data
    } else {
        &Value::Null
    }
}

/// 从 `video_list`(或单个 entry)里按 "H.264 + 比特率 + 面积" 评分挑最佳,返回 `(token, entry)`。
pub fn pick_main_url_entry(response: &Value) -> Option<(String, Value)> {
    let info = extract_video_info(response);
    let entries: Vec<&Value> = match info.get("video_list") {
        Some(Value::Object(map)) if !map.is_empty() => map.values().collect(),
        Some(Value::Array(items)) if !items.is_empty() => items.iter().collect(),
        _ => vec![info],
    };
    let mut best: Option<(String, f64, &Value)> = None;
    for entry in entries {
        if !entry.is_object() {
            continue;
        }
        let token = match first_truthy(entry, "main_url", "play_url") {
            Some(Value::String(s)) if !s.trim().is_empty() => s.trim().to_string(),
            _ => continue,
        };
        let codec = entry.get("codec_type").unwrap_or(&Value::Null);
        let codec_text = js_string(codec).to_ascii_lowercase();
        let is_h264 = codec_text.contains("h264")
            || codec_text.contains("avc")
            || codec.as_f64().is_some_and(|n| n == 0.0);
        let bitrate = js_number(first_truthy(entry, "bitrate", "real_bitrate"));
        let width = js_number(first_truthy(entry, "vwidth", "width"));
        let height = js_number(first_truthy(entry, "vheight", "height"));
        let score = if is_h264 { 1_000_000_000.0 } else { 0.0 } + bitrate + width * height;
        if best
            .as_ref()
            .is_none_or(|(_, best_score, _)| score > *best_score)
        {
            best = Some((token, score, entry));
        }
    }
    best.map(|(token, _, entry)| (token, entry.clone()))
}

// ---------------------------------------------------------------------------
// 解密
// ---------------------------------------------------------------------------

/// 解 token:本身是 https URL → 原样;base64 直译是 https URL → 直译;否则用 `key_seed` 走 AES 候选链。
pub fn decrypt_token(token: &str, key_seed: &str) -> Option<String> {
    if is_valid_https_url(token) {
        return Some(token.to_string());
    }
    if let Some(text) = base64_decode_flexible(token).and_then(|b| bytes_to_printable(&b)) {
        if is_valid_https_url(&text) {
            return Some(text);
        }
    }
    if key_seed.trim().is_empty() {
        return None;
    }
    decrypt_with_key_seed(token, key_seed)
}

/// 入口:挑 entry → 取 key_seed(response 优先,其次 api_url)→ 解密。
pub fn decode(response: &Value, api_url: &str) -> Result<String, &'static str> {
    let (token, _entry) = pick_main_url_entry(response).ok_or(FALLBACK_MAIN_URL_MISSING)?;
    let key_seed = extract_key_seed(response)
        .or_else(|| extract_key_seed_from_str(api_url))
        .unwrap_or_default();
    decrypt_token(&token, &key_seed).ok_or(FALLBACK_URL_DECODE_FAILED)
}

// ---------------------------------------------------------------------------
// 无水印 URL
// ---------------------------------------------------------------------------

/// 追加/替换 query `lr=unwatermarked`,其它 query 原样保留(不重新编码,签名参数不受影响),fragment 保留。
pub fn unwatermarked_url(url: &str) -> String {
    let (without_fragment, fragment) = match url.find('#') {
        Some(i) => (&url[..i], &url[i..]),
        None => (url, ""),
    };
    let (base, query) = match without_fragment.find('?') {
        Some(i) => (&without_fragment[..i], &without_fragment[i + 1..]),
        None => (without_fragment, ""),
    };
    let mut params: Vec<&str> = query
        .split('&')
        .filter(|p| !p.is_empty() && p.split('=').next() != Some("lr"))
        .collect();
    params.push("lr=unwatermarked");
    format!("{base}?{}{fragment}", params.join("&"))
}

/// 豆包视频 CDN 域白名单(与 reader.js `videoURL` 一致)。
pub fn is_video_cdn_host(host: &str) -> bool {
    VIDEO_CDN_HOST_RE.is_match(&host.to_ascii_lowercase())
}

/// 对标 reader.js `videoURL`:非 `https://` 开头先尝试 base64 直译;必须 https、无 userinfo、host 在白名单;返回规范化 href。
pub fn video_url(s: &str) -> Option<String> {
    let mut candidate = s.to_string();
    if !candidate.to_ascii_lowercase().starts_with("https://") {
        let stripped: String = candidate
            .chars()
            .filter(|c| !c.is_ascii_whitespace())
            .collect();
        if let Some(text) = ATOB
            .decode(&stripped)
            .ok()
            .and_then(|b| String::from_utf8(b).ok())
        {
            candidate = text;
        }
    }
    let parsed = Url::parse(&candidate).ok()?;
    if parsed.scheme() != "https" || !parsed.username().is_empty() || parsed.password().is_some() {
        return None;
    }
    if !is_video_cdn_host(parsed.host_str()?) {
        return None;
    }
    Some(parsed.into())
}

/// 对标 reader.js `unwatermarkedURL`:[`video_url`] 且 query `lr` 为 `unwatermarked` 才返回。
pub fn unwatermarked_video_url(s: &str) -> Option<String> {
    let href = video_url(s)?;
    let parsed = Url::parse(&href).ok()?;
    let (_, lr) = parsed.query_pairs().find(|(k, _)| k == "lr")?;
    (lr == "unwatermarked").then_some(href)
}

// ---------------------------------------------------------------------------
// 内部工具
// ---------------------------------------------------------------------------

/// JS 真值:`null`/`false`/`0`/`""` 为假。
fn truthy(v: &Value) -> bool {
    match v {
        Value::Null => false,
        Value::Bool(b) => *b,
        Value::Number(n) => n.as_f64().is_some_and(|f| f != 0.0),
        Value::String(s) => !s.is_empty(),
        Value::Array(_) | Value::Object(_) => true,
    }
}

/// JS `obj.a || obj.b`:第一个为真值的字段,都没有则 `b`(可能为 `None`)。
fn first_truthy<'a>(entry: &'a Value, a: &str, b: &str) -> Option<&'a Value> {
    entry.get(a).filter(|v| truthy(v)).or_else(|| entry.get(b))
}

/// JS `String(x ?? '')`。
fn js_string(v: &Value) -> String {
    match v {
        Value::Null => String::new(),
        Value::String(s) => s.clone(),
        other => other.to_string(),
    }
}

/// JS `Number(x || 0)`;解析不出的当 0(原版会得到 NaN)。
fn js_number(v: Option<&Value>) -> f64 {
    match v {
        Some(Value::Number(n)) => n.as_f64().unwrap_or(0.0),
        Some(Value::String(s)) => s.trim().parse::<f64>().unwrap_or(0.0),
        Some(Value::Bool(true)) => 1.0,
        _ => 0.0,
    }
}

/// `decodeURIComponent`:`%XX` 解码,非法序列或非 UTF-8 → `None`;`+` 不转空格。
fn percent_decode(s: &str) -> Option<String> {
    let bytes = s.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'%' {
            let hex = bytes.get(i + 1..i + 3)?;
            out.push(u8::from_str_radix(std::str::from_utf8(hex).ok()?, 16).ok()?);
            i += 3;
        } else {
            out.push(bytes[i]);
            i += 1;
        }
    }
    String::from_utf8(out).ok()
}

fn sha512(data: &[u8]) -> Vec<u8> {
    Sha512::digest(data).to_vec()
}

fn hex_decode(hex: &str) -> Vec<u8> {
    (0..hex.len())
        .step_by(2)
        .filter_map(|i| u8::from_str_radix(hex.get(i..i + 2)?, 16).ok())
        .collect()
}

/// 对一个 (密文, key, iv) 候选:解密 → [去填充, 去两层填充, 原样] 依次检查可打印 + https URL。
fn try_candidate(ciphertext: &[u8], key: &[u8], iv: &[u8]) -> Option<String> {
    let raw = aes_cbc_decrypt(ciphertext, key, iv)?;
    let once = pkcs7_unpad(&raw);
    let twice = pkcs7_unpad(once);
    // 填充值 0x09/0x0A/0x0D 恰好是 `bytes_to_printable` 放行的空白字符,而 URL 解析又会忽略它们,
    // 所以先看去填充最彻底的候选,并把首尾空白控制字符裁掉,避免把填充带进 URL。
    for bytes in [twice, once, raw.as_slice()] {
        if let Some(text) = bytes_to_printable(bytes) {
            let text = text.trim_matches(|c| matches!(c, '\t' | '\n' | '\r' | ' '));
            if is_valid_https_url(text) {
                return Some(text.to_string());
            }
        }
    }
    None
}

/// 完整解密(派生与候选顺序见模块文档)。
fn decrypt_with_key_seed(token: &str, key_seed: &str) -> Option<String> {
    let payload = base64_decode_flexible(token)?;
    let seed = base64_decode_flexible(key_seed)?;

    let salted = sha512(&[sha512(&seed[..seed.len().min(32)]), hex_decode(SALT_HEX)].concat());
    let (key_a, iv_a) = (&salted[..16], &salted[16..32]);
    let pepper = base64_decode_flexible(EXTRA_PEPPER)?;
    let peppered = sha512(&[sha512(&seed), pepper].concat());
    let (key_p, iv_p) = (&peppered[..16], &peppered[16..32]);

    let mut candidates: Vec<(&[u8], &[u8], &[u8])> = Vec::with_capacity(5);
    if payload.len() >= 4 && payload[..4] == MAGIC {
        let body = &payload[4..];
        candidates.push((body, key_p, iv_p));
        candidates.push((body, key_a, iv_a));
        candidates.push((body, iv_a, key_a));
        if payload.len() > 0x24 {
            candidates.push((&payload[0x24..], key_a, &payload[0x14..0x24]));
            candidates.push((&payload[0x24..], key_a, iv_a));
        }
    } else {
        candidates.push((&payload, key_a, iv_a));
    }
    candidates
        .into_iter()
        .find_map(|(ct, key, iv)| try_candidate(ct, key, iv))
}

#[cfg(test)]
mod tests {
    use super::*;
    use cbc::cipher::block_padding::Pkcs7;
    use cbc::cipher::BlockEncryptMut;
    use serde_json::json;

    // ---- 原版 JS(Node oracle)生成的向量 ----
    const ORACLE_SEED: &str = "e+qZAOEzRnF8Hq8o4bKwzT2I9aXVH+oJqx6nQetrd/SEce9+6qMRdQ==";
    const ORACLE_URL: &str = "https://v3-vdl.doubao.com/abc/video.mp4?lr=unwatermarked&sig=xyz";
    const TOKEN_PEPPER: &str = "qAABAHDiHNTpnQysSbB3cczsvRvOJ2368SAhWvVPkhyWfsgzUxopAj9Wbz0F9AfDiVMgID0/S38BdAwmhhkyN1LVL53W0YMP3IKLGDvKrFb9hhJ5";
    const TOKEN_SALTED_DIRECT: &str = "qAABAK7Lzh/YBEsVT6KQJIgb90QZ5yUOOG60oOmEZCbrv1c53ABDblj3PQMWbDc4tHYnK+r/waSBBLNyLc5cVWSoZ4jKlQqBTcyLxyHAhBgczS10";
    const TOKEN_SALTED_CROSSED: &str = "qAABALcmWG5WbvQM2P8hcQL8jEBIjoGAKaW7dy9+yDKgEnN4K2/5VuTfOIv6ZU+HEBNCMw/pOk72DhDAyboKwrt5R3hC717wUGN9AXMV2zQjFXeI";
    const TOKEN_OFFSET24_EMBEDDED_IV: &str = "qAABAIaHC7IWMXCTVjca40Tv5LPWGlhL7sZauo881rGKFFX3128k7FuxpLDNwyRXKE5X0U9qmEWxX9zBYCr0SBns4VoZiRD5dAmZ8a+e2kjaKj6c8YJe3G0M6m4/1Fhn5uKWi1UE+afG55iSd0f6rncWf+s=";
    const TOKEN_OFFSET24_IV_A: &str = "qAABAEs24iPAQglzbnV4hd0i77eIcKFt/HOSkLxU57FEb20OrsvOH9gESxVPopAkiBv3RBnnJQ44brSg6YRkJuu/VzncAENuWPc9AxZsNzi0dicr6v/BpIEEs3ItzlxVZKhniMqVCoFNzIvHIcCEGBzNLXQ=";
    /// analysis/06 伪代码里 "sha512(seed[:32]) 直接对半" 的派生,原脚本解不出来 → Rust 也不应解出。
    const TOKEN_DOC_DIRECT: &str = "qAABALivD3eSkwFREfylEYTKXs/Kv1EN07cLEPEzkcxMJIwgyrgCYpw3HPfTVXIUXzyp0c87c3frFv+YcOBAIF9kdi/ZArsIuMPXoVu71hLlbw6C";

    fn sha512(b: &[u8]) -> Vec<u8> {
        Sha512::digest(b).to_vec()
    }

    fn derive(seed_b64: &str) -> (Vec<u8>, Vec<u8>, Vec<u8>, Vec<u8>) {
        let seed = base64_decode_flexible(seed_b64).unwrap();
        let salt: Vec<u8> = (0..SALT_HEX.len())
            .step_by(2)
            .map(|i| u8::from_str_radix(&SALT_HEX[i..i + 2], 16).unwrap())
            .collect();
        let pepper = base64_decode_flexible(EXTRA_PEPPER).unwrap();
        let a = sha512(&[sha512(&seed[..seed.len().min(32)]), salt].concat());
        let p = sha512(&[sha512(&seed), pepper].concat());
        (
            a[..16].to_vec(),
            a[16..32].to_vec(),
            p[..16].to_vec(),
            p[16..32].to_vec(),
        )
    }

    fn encrypt_pkcs7(plain: &[u8], key: &[u8], iv: &[u8]) -> Vec<u8> {
        cbc::Encryptor::<Aes128>::new_from_slices(key, iv)
            .unwrap()
            .encrypt_padded_vec_mut::<Pkcs7>(plain)
    }

    fn encrypt_nopad(plain: &[u8], key: &[u8], iv: &[u8]) -> Vec<u8> {
        cbc::Encryptor::<Aes128>::new_from_slices(key, iv)
            .unwrap()
            .encrypt_padded_vec_mut::<NoPadding>(plain)
    }

    fn b64(bytes: &[u8]) -> String {
        base64::engine::general_purpose::STANDARD.encode(bytes)
    }

    // ---- base64_decode_flexible ----

    #[test]
    fn base64_standard_and_urlsafe_variants() {
        assert_eq!(
            base64_decode_flexible("+/+/").unwrap(),
            vec![0xfb, 0xff, 0xbf]
        );
        // URL-safe,无填充,前后空白
        assert_eq!(
            base64_decode_flexible("  -_-_ \n").unwrap(),
            vec![0xfb, 0xff, 0xbf]
        );
        assert_eq!(base64_decode_flexible("+w").unwrap(), vec![0xfb]);
        assert_eq!(base64_decode_flexible("+w==").unwrap(), vec![0xfb]);
    }

    #[test]
    fn base64_dollar_at_hash_variants() {
        // 变体 2:`$`→`_`,`@`→`/`(URL-safe 风格)
        assert_eq!(
            base64_decode_flexible("-$-$").unwrap(),
            vec![0xfb, 0xff, 0xbf]
        );
        // 变体 3:`$`→`+`,`@`→`/`,`#`→`=`(变体 2 会把 `#` 变成 `.` 而失败)
        assert_eq!(base64_decode_flexible("$w##").unwrap(), vec![0xfb]);
        // `$` 二义(`_` 或 `+`):变体 2 先命中,与原版 JS 一致("$@$@" → "_/_/" → "////")
        assert_eq!(
            base64_decode_flexible("$@$@").unwrap(),
            vec![0xff, 0xff, 0xff]
        );
    }

    #[test]
    fn base64_rejects_garbage_and_empty() {
        assert!(base64_decode_flexible("").is_none());
        assert!(base64_decode_flexible("   ").is_none());
        assert!(base64_decode_flexible("!!!!").is_none());
        // 长度 %4 == 1 补不出合法填充
        assert!(base64_decode_flexible("A").is_none());
        assert!(base64_decode_flexible("AAAAA").is_none());
    }

    // ---- bytes_to_printable ----

    #[test]
    fn printable_accepts_ascii_and_whitespace_controls() {
        assert_eq!(
            bytes_to_printable(b"https://a.b/c?d=1\t\n\r").unwrap(),
            "https://a.b/c?d=1\t\n\r"
        );
    }

    #[test]
    fn printable_rejects_empty_control_and_high_bytes() {
        assert!(bytes_to_printable(b"").is_none());
        assert!(bytes_to_printable(b"abc\x00").is_none());
        assert!(bytes_to_printable(b"abc\x7f").is_none());
        assert!(bytes_to_printable("中".as_bytes()).is_none());
        assert!(bytes_to_printable(b"\x08").is_none());
    }

    // ---- is_valid_https_url ----

    #[test]
    fn https_url_validation() {
        assert!(is_valid_https_url("https://v3-vdl.doubao.com/a.mp4?x=1"));
        assert!(!is_valid_https_url("http://v3-vdl.doubao.com/a.mp4"));
        assert!(!is_valid_https_url("https://user:pw@host/a"));
        assert!(!is_valid_https_url("https://user@host/a"));
        assert!(!is_valid_https_url("https://"));
        assert!(!is_valid_https_url("not a url"));
        assert!(!is_valid_https_url(""));
    }

    // ---- pkcs7_unpad / aes_cbc_decrypt ----

    #[test]
    fn pkcs7_unpad_strips_valid_and_keeps_invalid() {
        assert_eq!(pkcs7_unpad(b"abc\x02\x02"), b"abc");
        assert_eq!(pkcs7_unpad(b"abc\x01"), b"abc");
        assert_eq!(pkcs7_unpad(&[0x10; 16]), b"");
        // 最后字节与倒数第二不一致
        assert_eq!(pkcs7_unpad(b"abc\x01\x02"), b"abc\x01\x02");
        assert_eq!(pkcs7_unpad(b"abc\x00"), b"abc\x00");
        assert_eq!(pkcs7_unpad(b"abc\x11"), b"abc\x11");
        assert_eq!(pkcs7_unpad(b"\x03"), b"\x03");
        assert_eq!(pkcs7_unpad(b""), b"");
    }

    #[test]
    fn aes_cbc_roundtrip_without_padding_handling() {
        let key = [0x11u8; 16];
        let iv = [0x22u8; 16];
        let plain = b"https://example.com/video.mp4?lr=unwatermarked";
        let ct = encrypt_pkcs7(plain, &key, &iv);
        let raw = aes_cbc_decrypt(&ct, &key, &iv).unwrap();
        assert_eq!(raw.len(), ct.len());
        assert_eq!(pkcs7_unpad(&raw), plain);
        let plain32 = b"0123456789abcdef0123456789abcdef";
        let ct2 = encrypt_nopad(plain32, &key, &iv);
        assert_eq!(aes_cbc_decrypt(&ct2, &key, &iv).unwrap(), plain32);
    }

    #[test]
    fn aes_cbc_rejects_bad_lengths() {
        let key = [0u8; 16];
        let iv = [0u8; 16];
        assert!(aes_cbc_decrypt(&[], &key, &iv).is_none());
        assert!(aes_cbc_decrypt(&[0u8; 15], &key, &iv).is_none());
        assert!(aes_cbc_decrypt(&[0u8; 17], &key, &iv).is_none());
        assert!(aes_cbc_decrypt(&[0u8; 16], &[0u8; 15], &iv).is_none());
        assert!(aes_cbc_decrypt(&[0u8; 16], &[0u8; 32], &iv).is_none());
        assert!(aes_cbc_decrypt(&[0u8; 16], &key, &[0u8; 8]).is_none());
    }

    // ---- extract_key_seed ----

    #[test]
    fn key_seed_from_str_query_and_json_forms() {
        assert_eq!(
            extract_key_seed_from_str("https://vas-x.snssdk.com/a?b=1&key_seed=ab%2Bcd%3D%3D&c=2")
                .unwrap(),
            "ab+cd=="
        );
        assert_eq!(extract_key_seed_from_str("key_seed=xyz").unwrap(), "xyz");
        assert_eq!(extract_key_seed_from_str("KEY_SEED=Up").unwrap(), "Up");
        assert_eq!(
            extract_key_seed_from_str(r#"{"foo":1,"key_seed" : "s%20d"}"#).unwrap(),
            "s d"
        );
        assert_eq!(extract_key_seed_from_str("{'key_seed':'q'}").unwrap(), "q");
        assert!(extract_key_seed_from_str("nokey_seed=1").is_none());
        assert!(extract_key_seed_from_str("https://a/b").is_none());
        assert!(extract_key_seed_from_str("").is_none());
        // 非法 percent 序列(decodeURIComponent 会抛)→ None
        assert!(extract_key_seed_from_str("key_seed=%zz").is_none());
    }

    #[test]
    fn key_seed_direct_field_is_trimmed_and_preferred() {
        let v = json!({ "data": { "key_seed": "nested" }, "key_seed": "  top  " });
        assert_eq!(extract_key_seed(&v).unwrap(), "top");
        // 空白字段不算
        let v = json!({ "key_seed": "   ", "data": { "key_seed": "nested" } });
        assert_eq!(extract_key_seed(&v).unwrap(), "nested");
        // 非字符串字段不算
        let v = json!({ "key_seed": 123, "x": [1, { "key_seed": "in_array" }] });
        assert_eq!(extract_key_seed(&v).unwrap(), "in_array");
    }

    #[test]
    fn key_seed_three_positions() {
        // 1) JSON 字段
        assert_eq!(
            extract_key_seed(&json!({ "a": { "b": { "key_seed": "f" } } })).unwrap(),
            "f"
        );
        // 2) 字符串里的 query
        assert_eq!(
            extract_key_seed(&json!({ "api": "https://h/p?key_seed=q%2F" })).unwrap(),
            "q/"
        );
        // 3) 字符串里内嵌 JSON
        assert_eq!(
            extract_key_seed(&json!({ "raw": "{\"key_seed\":\"j\"}" })).unwrap(),
            "j"
        );
        assert!(extract_key_seed(&json!({ "a": 1, "b": [true, null] })).is_none());
        assert!(extract_key_seed(&Value::Null).is_none());
        assert!(extract_key_seed(&json!("plain")).is_none());
        assert_eq!(extract_key_seed(&json!("key_seed=s")).unwrap(), "s");
    }

    #[test]
    fn key_seed_depth_limit_is_ten() {
        fn nest(n: usize) -> Value {
            let mut v = json!({ "key_seed": "deep" });
            for _ in 0..n {
                v = json!({ "a": v });
            }
            v
        }
        assert_eq!(extract_key_seed(&nest(10)).unwrap(), "deep");
        assert!(extract_key_seed(&nest(11)).is_none());
    }

    // ---- extract_video_info / pick_main_url_entry ----

    #[test]
    fn video_info_three_shapes() {
        let list = json!({ "video_list": { "v": { "main_url": "T" } } });
        let r1 = json!({ "data": { "video_info": list } });
        assert_eq!(extract_video_info(&r1), &list);
        let r2 = json!({ "video_info": { "data": list } });
        assert_eq!(extract_video_info(&r2), &list);
        assert_eq!(extract_video_info(&list), &list);
        let r3 = json!({ "data": list });
        assert_eq!(extract_video_info(&r3), &list);
        assert!(!extract_video_info(&json!("str")).is_object());
        assert!(!extract_video_info(&json!({ "video_info": "x" })).is_object());
    }

    #[test]
    fn pick_scores_h264_then_bitrate_then_area() {
        let data = json!({ "video_list": {
            "a": { "main_url": "A", "codec_type": "h265", "bitrate": 5000000, "vwidth": 1920, "vheight": 1080 },
            "b": { "main_url": "B", "codec_type": 0, "bitrate": 100 },
            "c": { "play_url": "C", "codec_type": "avc1", "bitrate": "200", "width": 2, "height": 2 },
        } });
        let (token, entry) = pick_main_url_entry(&data).unwrap();
        assert_eq!(token, "C");
        assert_eq!(entry["play_url"], "C");

        let data = json!({ "video_list": {
            "a": { "main_url": "A", "codec_type": "h265", "bitrate": 5000000, "vwidth": 1920, "vheight": 1080 },
            "b": { "main_url": "B", "codec_type": "hevc", "real_bitrate": 1, "vwidth": 4000, "vheight": 4000 },
        } });
        assert_eq!(pick_main_url_entry(&data).unwrap().0, "B");
        // 同分取先出现者(BTreeMap 键序)
        let data = json!({ "video_list": { "x": { "main_url": "X" }, "y": { "main_url": "Y" } } });
        assert_eq!(pick_main_url_entry(&data).unwrap().0, "X");
    }

    #[test]
    fn pick_single_entry_trims_and_falls_back_on_empty_list() {
        let (t, e) = pick_main_url_entry(&json!({ "play_url": " X " })).unwrap();
        assert_eq!(t, "X");
        assert_eq!(e["play_url"], " X ");
        let (t, _) = pick_main_url_entry(&json!({ "video_list": {}, "main_url": "Y" })).unwrap();
        assert_eq!(t, "Y");
        // 数组形式的 video_list
        let (t, _) =
            pick_main_url_entry(&json!({ "video_list": [{ "main_url": "Z", "codec_type": 0 }] }))
                .unwrap();
        assert_eq!(t, "Z");
        // 嵌套响应形态
        let (t, _) = pick_main_url_entry(
            &json!({ "data": { "video_info": { "video_list": { "v": { "main_url": "N" } } } } }),
        )
        .unwrap();
        assert_eq!(t, "N");
    }

    #[test]
    fn pick_skips_entries_without_string_token() {
        assert!(pick_main_url_entry(&json!({ "data": {} })).is_none());
        assert!(
            pick_main_url_entry(&json!({ "video_list": { "a": { "main_url": 5 } } })).is_none()
        );
        assert!(
            pick_main_url_entry(&json!({ "video_list": { "a": { "main_url": "  " } } })).is_none()
        );
        assert!(pick_main_url_entry(&json!({ "video_list": { "a": "str", "b": null } })).is_none());
        assert!(pick_main_url_entry(&json!(null)).is_none());
        assert!(pick_main_url_entry(&json!("x")).is_none());
    }

    // ---- decrypt_token ----

    #[test]
    fn decrypt_token_fast_paths() {
        assert_eq!(decrypt_token(ORACLE_URL, "").unwrap(), ORACLE_URL);
        assert_eq!(
            decrypt_token(&b64(ORACLE_URL.as_bytes()), "").unwrap(),
            ORACLE_URL
        );
        assert!(decrypt_token(&b64(b"http://x.com/a"), "").is_none());
        assert!(decrypt_token("", "").is_none());
        assert!(decrypt_token("not base64 at all!", "").is_none());
    }

    #[test]
    fn decrypt_token_oracle_vectors() {
        for (name, tok) in [
            ("pepper", TOKEN_PEPPER),
            ("salted_direct", TOKEN_SALTED_DIRECT),
            ("salted_crossed", TOKEN_SALTED_CROSSED),
            ("offset24_embedded_iv", TOKEN_OFFSET24_EMBEDDED_IV),
            ("offset24_iv_a", TOKEN_OFFSET24_IV_A),
        ] {
            assert_eq!(
                decrypt_token(tok, ORACLE_SEED).as_deref(),
                Some(ORACLE_URL),
                "{name}"
            );
        }
        assert!(decrypt_token(TOKEN_DOC_DIRECT, ORACLE_SEED).is_none());
        // 没有 key_seed 解不开
        assert!(decrypt_token(TOKEN_PEPPER, "").is_none());
        // 错的 key_seed 解不开
        assert!(decrypt_token(TOKEN_PEPPER, "AAAA").is_none());
    }

    #[test]
    fn decrypt_token_base64_variants_of_token_and_seed() {
        let urlsafe = TOKEN_PEPPER
            .replace('+', "-")
            .replace('/', "_")
            .trim_end_matches('=')
            .to_string();
        assert_eq!(decrypt_token(&urlsafe, ORACLE_SEED).unwrap(), ORACLE_URL);
        let dollar = TOKEN_PEPPER
            .replace('+', "$")
            .replace('/', "@")
            .replace('=', "#");
        assert_eq!(decrypt_token(&dollar, ORACLE_SEED).unwrap(), ORACLE_URL);
        let seed_dollar = ORACLE_SEED
            .replace('+', "$")
            .replace('/', "@")
            .replace('=', "#");
        assert_eq!(
            decrypt_token(TOKEN_PEPPER, &seed_dollar).unwrap(),
            ORACLE_URL
        );
    }

    #[test]
    fn decrypt_token_self_built_candidates() {
        let seed_b64 = b64(&(0u8..40).collect::<Vec<u8>>());
        let (key_a, iv_a, key_p, iv_p) = derive(&seed_b64);
        let url = "https://v9-vdl.doubao.com/x/y.mp4?lr=unwatermarked&a=b";

        let mk = |ct: Vec<u8>| b64(&[MAGIC.to_vec(), ct].concat());
        assert_eq!(
            decrypt_token(&mk(encrypt_pkcs7(url.as_bytes(), &key_p, &iv_p)), &seed_b64).unwrap(),
            url
        );
        assert_eq!(
            decrypt_token(&mk(encrypt_pkcs7(url.as_bytes(), &key_a, &iv_a)), &seed_b64).unwrap(),
            url
        );
        assert_eq!(
            decrypt_token(&mk(encrypt_pkcs7(url.as_bytes(), &iv_a, &key_a)), &seed_b64).unwrap(),
            url
        );
        // offset 0x24 变体:[4..0x14) 任意,[0x14..0x24) 是 iv
        let iv = [0x5au8; 16];
        let payload = [
            MAGIC.to_vec(),
            vec![0x77; 16],
            iv.to_vec(),
            encrypt_pkcs7(url.as_bytes(), &key_a, &iv),
        ]
        .concat();
        assert_eq!(decrypt_token(&b64(&payload), &seed_b64).unwrap(), url);
        // 无 magic:整个 payload 用 (key_a, iv_a)
        assert_eq!(
            decrypt_token(
                &b64(&encrypt_pkcs7(url.as_bytes(), &key_a, &iv_a)),
                &seed_b64
            )
            .unwrap(),
            url
        );
        // 嵌套填充(明文已带一层 PKCS7 再整体加密)也能还原,且不把填充字节带进 URL
        let mut padded = url.as_bytes().to_vec();
        let pad = 16 - padded.len() % 16;
        assert_eq!(pad, 10, "该 URL 的填充值恰好是 0x0A,覆盖空白字符填充的边界");
        padded.extend(std::iter::repeat_n(pad as u8, pad));
        assert_eq!(
            decrypt_token(&mk(encrypt_pkcs7(&padded, &key_p, &iv_p)), &seed_b64).unwrap(),
            url
        );
        // 明文不足一块、用 0x09 手工填满(无标准 PKCS7)的也能还原
        let short = "https://a.b/c";
        let mut tabbed = short.as_bytes().to_vec();
        tabbed.resize(16, 0x09);
        assert_eq!(
            decrypt_token(&mk(encrypt_nopad(&tabbed, &key_p, &iv_p)), &seed_b64).unwrap(),
            short
        );
        // 密文不是 16 倍数 / 明文不是 URL 都解不出
        let mut bad = encrypt_pkcs7(url.as_bytes(), &key_p, &iv_p);
        bad.push(0);
        assert!(decrypt_token(&mk(bad), &seed_b64).is_none());
        assert!(decrypt_token(
            &mk(encrypt_pkcs7(b"hello world, not a url", &key_p, &iv_p)),
            &seed_b64
        )
        .is_none());
    }

    // ---- decode ----

    #[test]
    fn decode_three_response_shapes() {
        let r1 = json!({
            "data": { "video_info": { "video_list": { "video_1": { "main_url": TOKEN_PEPPER, "codec_type": 0, "bitrate": 1 } } } },
            "key_seed": ORACLE_SEED
        });
        assert_eq!(
            decode(&r1, "https://vas-x.snssdk.com/a").unwrap(),
            ORACLE_URL
        );

        let r2 = json!({ "video_info": { "data": { "video_list": { "v": { "main_url": TOKEN_PEPPER, "codec_type": "h265" } } } } });
        let api = format!(
            "https://vas-x.snssdk.com/a?key_seed={}",
            ORACLE_SEED
                .replace('+', "%2B")
                .replace('/', "%2F")
                .replace('=', "%3D")
        );
        assert_eq!(decode(&r2, &api).unwrap(), ORACLE_URL);

        let r3 = json!({ "main_url": TOKEN_PEPPER, "key_seed": ORACLE_SEED });
        assert_eq!(decode(&r3, "").unwrap(), ORACLE_URL);

        // key_seed 在 entry 内
        let r4 = json!({ "data": { "video_info": { "video_list": { "v": { "main_url": TOKEN_SALTED_DIRECT, "key_seed": ORACLE_SEED } } } } });
        assert_eq!(decode(&r4, "").unwrap(), ORACLE_URL);
    }

    #[test]
    fn decode_error_reasons() {
        assert_eq!(
            decode(&json!({ "data": {} }), "").unwrap_err(),
            FALLBACK_MAIN_URL_MISSING
        );
        assert_eq!(
            decode(&json!(null), "").unwrap_err(),
            "fallback_main_url_missing"
        );
        assert_eq!(
            decode(&json!({ "main_url": TOKEN_PEPPER, "key_seed": "AAAA" }), "").unwrap_err(),
            FALLBACK_URL_DECODE_FAILED
        );
        assert_eq!(
            decode(
                &json!({ "main_url": TOKEN_PEPPER }),
                "https://vas-x.snssdk.com/a"
            )
            .unwrap_err(),
            "fallback_url_decode_failed"
        );
    }

    /// 真实样本:`tests/fixtures/video_token_1.json` = `{ "response": {...}, "api_url": "...", "expected_url": "..." }`
    /// (Task 3.3 UC-20 真机抓取,`qAAB` token + `key_seed`,expected_url 由原版 JS 解出)。
    #[test]
    fn fixture_video_token_1_decodes() {
        let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("tests/fixtures/video_token_1.json");
        let raw =
            std::fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
        let fixture: Value = serde_json::from_str(&raw).unwrap();
        let api_url = fixture["api_url"].as_str().unwrap_or("");
        let url = decode(&fixture["response"], api_url).unwrap();
        assert_eq!(url, fixture["expected_url"].as_str().unwrap());
        assert!(unwatermarked_video_url(&url).is_some());
    }

    // ---- unwatermarked_url ----

    #[test]
    fn unwatermarked_url_appends_or_replaces_lr() {
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4"),
            "https://a.b/c.mp4?lr=unwatermarked"
        );
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4?x=1&y=%2F+z"),
            "https://a.b/c.mp4?x=1&y=%2F+z&lr=unwatermarked"
        );
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4?lr=watermarked&x=1"),
            "https://a.b/c.mp4?x=1&lr=unwatermarked"
        );
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4?lr&lr=1"),
            "https://a.b/c.mp4?lr=unwatermarked"
        );
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4?x=1#frag"),
            "https://a.b/c.mp4?x=1&lr=unwatermarked#frag"
        );
        assert_eq!(
            unwatermarked_url("https://a.b/c.mp4?"),
            "https://a.b/c.mp4?lr=unwatermarked"
        );
        let once = unwatermarked_url("https://a.b/c.mp4?sig=abc");
        assert_eq!(unwatermarked_url(&once), once);
        assert!(
            unwatermarked_video_url(&unwatermarked_url("https://v1-vdl.doubao.com/a")).is_some()
        );
    }

    // ---- CDN 白名单 / reader.js 等价 ----

    #[test]
    fn video_cdn_host_whitelist() {
        for h in [
            "v3-vdl.doubao.com",
            "v12-vdl.doubao.com",
            "doubaovideo.com",
            "x.doubaovideo.com",
            "a.b.volces.com",
            "p.bytevcloudcn.com",
            "q.bytevcloud.com",
            "ixigua.com",
            "s.365yg.com",
            "v.volcvideo.com",
            "v.bytevideo.com",
        ] {
            assert!(is_video_cdn_host(h), "{h}");
        }
        for h in [
            "www.doubao.com",
            "vdl.doubao.com",
            "v-vdl.doubao.com",
            "evil-volces.com",
            "volces.com.evil",
            "",
        ] {
            assert!(!is_video_cdn_host(h), "{h}");
        }
    }

    #[test]
    fn video_url_matches_reader_js() {
        assert_eq!(
            video_url("https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked").unwrap(),
            "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked"
        );
        // base64 直译
        assert_eq!(
            video_url(&b64(b"https://v3-vdl.doubao.com/a.mp4")).unwrap(),
            "https://v3-vdl.doubao.com/a.mp4"
        );
        assert!(video_url("http://v3-vdl.doubao.com/a.mp4").is_none());
        assert!(video_url("https://u@v3-vdl.doubao.com/a.mp4").is_none());
        assert!(video_url("https://www.doubao.com/a.mp4").is_none());
        assert!(video_url("nope").is_none());
        assert!(video_url("").is_none());

        assert_eq!(
            unwatermarked_video_url("https://v3-vdl.doubao.com/a.mp4?x=1&lr=unwatermarked")
                .unwrap(),
            "https://v3-vdl.doubao.com/a.mp4?x=1&lr=unwatermarked"
        );
        assert!(
            unwatermarked_video_url("https://v3-vdl.doubao.com/a.mp4?lr=watermarked").is_none()
        );
        assert!(unwatermarked_video_url("https://v3-vdl.doubao.com/a.mp4").is_none());
        assert!(unwatermarked_video_url("https://www.doubao.com/a.mp4?lr=unwatermarked").is_none());
    }

    #[test]
    fn embedded_decoder_script_carries_same_constants() {
        assert!(VIDEO_DECODER_SCRIPT.contains(SALT_HEX));
        assert!(VIDEO_DECODER_SCRIPT.contains(EXTRA_PEPPER));
        assert!(VIDEO_DECODER_SCRIPT.contains("shenlangDoubaoVideoDecoder"));
        assert!(VIDEO_DECODER_SCRIPT.contains(FALLBACK_URL_DECODE_FAILED));
        assert!(VIDEO_DECODER_SCRIPT.contains(FALLBACK_MAIN_URL_MISSING));
    }
}
