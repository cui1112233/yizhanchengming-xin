//! Bearer 鉴权(`docs/spec/12-security.md`):`Authorization: Bearer <token>`,
//! token 取当前配置(设置页重新生成后立即生效),常量时间比较(`subtle::ConstantTimeEq`)。
//! 缺 header → 401 `auth_missing_token`;格式不对或不匹配 → 401 `auth_invalid_token`。

use std::sync::Arc;

use axum::extract::{Request, State};
use axum::http::header::AUTHORIZATION;
use axum::http::HeaderMap;
use axum::middleware::Next;
use axum::response::{IntoResponse, Response};
use subtle::ConstantTimeEq;

use crate::error::{RelayError, RelayResult};
use crate::state::AppState;

/// 常量时间比较;长度不同直接为 false(长度本身不是秘密)。
pub fn verify_token(provided: &str, expected: &str) -> bool {
    if expected.is_empty() || provided.len() != expected.len() {
        return false;
    }
    provided.as_bytes().ct_eq(expected.as_bytes()).into()
}

/// `auth_missing_token`(401)。
pub fn missing_token() -> RelayError {
    RelayError::new(
        "auth_missing_token",
        "未提供 Bearer token,请在 Authorization 头携带设置页显示的 token",
    )
}

/// `auth_invalid_token`(401)。
pub fn invalid_token() -> RelayError {
    RelayError::new(
        "auth_invalid_token",
        "token 无效,请到设置页复制当前 token 后重试",
    )
}

/// 从请求头取 Bearer token:没有 `Authorization` → `auth_missing_token`;
/// 不是 `Bearer <token>` 形式(spec 12:严格剥掉 `"Bearer "`,不做 trim)、token 为空或含空白
/// → `auth_invalid_token`。
pub fn extract_bearer(headers: &HeaderMap) -> RelayResult<&str> {
    let raw = headers.get(AUTHORIZATION).ok_or_else(missing_token)?;
    let text = raw.to_str().map_err(|_| invalid_token())?;
    let token = text.strip_prefix("Bearer ").ok_or_else(invalid_token)?;
    if token.is_empty() || token.bytes().any(|b| b.is_ascii_whitespace()) {
        return Err(invalid_token());
    }
    Ok(token)
}

/// axum 中间件:校验通过才放行。
pub async fn auth_middleware(
    State(state): State<Arc<AppState>>,
    req: Request,
    next: Next,
) -> Response {
    let token = match extract_bearer(req.headers()) {
        Ok(t) => t.to_string(),
        Err(e) => return e.into_response(),
    };
    let expected = state.config.read().await.server.bearer_token.clone();
    if !verify_token(&token, &expected) {
        return invalid_token().into_response();
    }
    next.run(req).await
}

/// 兼容接口用:有 `Authorization` 头就按 Bearer 校验;没有则接受 query `?key=<token>`
/// (参照项目 doubao2api 的行为)。
pub fn extract_bearer_or_key(headers: &HeaderMap, query: Option<&str>) -> RelayResult<String> {
    if headers.contains_key(AUTHORIZATION) {
        return extract_bearer(headers).map(str::to_string);
    }
    let key = query.and_then(|q| {
        url::form_urlencoded::parse(q.as_bytes())
            .find(|(k, _)| k == "key")
            .map(|(_, v)| v.into_owned())
    });
    match key {
        Some(k) if !k.is_empty() => Ok(k),
        Some(_) => Err(invalid_token()),
        None => Err(missing_token()),
    }
}

/// 兼容接口的鉴权中间件(Bearer 或 `?key=`)。
pub async fn key_or_bearer_middleware(
    State(state): State<Arc<AppState>>,
    req: Request,
    next: Next,
) -> Response {
    let token = match extract_bearer_or_key(req.headers(), req.uri().query()) {
        Ok(t) => t,
        Err(e) => return e.into_response(),
    };
    let expected = state.config.read().await.server.bearer_token.clone();
    if !verify_token(&token, &expected) {
        return invalid_token().into_response();
    }
    next.run(req).await
}

#[cfg(test)]
mod tests {
    #[test]
    fn bearer_or_key_accepts_query_key_only_without_header() {
        let mut h = HeaderMap::new();
        assert_eq!(extract_bearer_or_key(&h, Some("key=abc")).unwrap(), "abc");
        assert_eq!(
            extract_bearer_or_key(&h, Some("x=1&key=a%2Bb")).unwrap(),
            "a+b"
        );
        assert_eq!(
            extract_bearer_or_key(&h, None).unwrap_err().code,
            "auth_missing_token"
        );
        assert_eq!(
            extract_bearer_or_key(&h, Some("key=")).unwrap_err().code,
            "auth_invalid_token"
        );
        h.insert(AUTHORIZATION, HeaderValue::from_static("Bearer hdr"));
        assert_eq!(extract_bearer_or_key(&h, Some("key=abc")).unwrap(), "hdr");
    }

    use super::*;
    use axum::http::HeaderValue;

    #[test]
    fn verify_token_is_exact_match_only() {
        assert!(verify_token("abc", "abc"));
        assert!(!verify_token("abd", "abc"));
        assert!(!verify_token("ab", "abc"));
        assert!(!verify_token("abcd", "abc"));
        assert!(!verify_token("", ""), "空 token 永远不通过");
        assert!(!verify_token("x", ""));
    }

    #[test]
    fn extract_bearer_distinguishes_missing_and_invalid() {
        let mut h = HeaderMap::new();
        assert_eq!(extract_bearer(&h).unwrap_err().code, "auth_missing_token");
        h.insert(AUTHORIZATION, HeaderValue::from_static("Basic abc"));
        assert_eq!(extract_bearer(&h).unwrap_err().code, "auth_invalid_token");
        h.insert(AUTHORIZATION, HeaderValue::from_static("Bearer "));
        assert_eq!(extract_bearer(&h).unwrap_err().code, "auth_invalid_token");
        h.insert(AUTHORIZATION, HeaderValue::from_static("bearer abc"));
        assert_eq!(
            extract_bearer(&h).unwrap_err().code,
            "auth_invalid_token",
            "前缀区分大小写"
        );
        h.insert(AUTHORIZATION, HeaderValue::from_static("Bearer abc"));
        assert_eq!(extract_bearer(&h).unwrap(), "abc");
    }

    #[test]
    fn extract_bearer_rejects_surrounding_or_embedded_whitespace() {
        for raw in ["Bearer  abc", "Bearer abc ", "Bearer \tabc", "Bearer a bc"] {
            let mut h = HeaderMap::new();
            h.insert(AUTHORIZATION, HeaderValue::from_static(raw));
            assert_eq!(
                extract_bearer(&h).unwrap_err().code,
                "auth_invalid_token",
                "{raw:?} 不应被接受"
            );
        }
    }
}
