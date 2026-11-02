//! HTTP API 服务器(Task 2.8,`docs/spec/04-http-api.md`):axum 0.8。
//!
//! - 路由统一挂在 [`API_PREFIX`](`/api/v1`)下;`/health` 免鉴权,其余经 [`auth::auth_middleware`]
//!   (Bearer,常量时间比较);错误一律 `RelayError` → `{error:{code,message,details}}`
//!   + 11-error-codes 的状态码([`error`])。
//! - 只允许绑定 loopback(`127.0.0.1` / `::1` / `localhost`),[`serve`] 绑定失败返回 `Err`
//!   (用户可在设置页改端口),成功时 info 日志 `HTTP API 已监听 {host}:{port}`。
//! - 每个请求一条 info 访问日志 `{method, path, status, duration_ms}`,不含 header / body
//!   (bearer token 与 prompt 不得出现在日志里)。

pub mod auth;
pub mod error;
pub mod routes;

use std::sync::atomic::Ordering;
use std::sync::Arc;
use std::time::Instant;

use anyhow::Context;
use axum::extract::Request;
use axum::middleware::{self, Next};
use axum::response::Response;
use axum::Router;
use tokio::net::TcpListener;
use tokio_util::sync::CancellationToken;
use tower_http::cors::{Any, CorsLayer};

use crate::state::AppState;

/// 所有端点的前缀(定义在任务域,详情体的 `download_url` 也用它)。
pub use crate::tasks::detail::API_PREFIX;

const LOOPBACK_HOSTS: &[&str] = &["127.0.0.1", "::1", "localhost"];

/// 组装完整路由(带状态)。
///
/// 已知路径 + 不支持的方法 → 405 JSON(`method_not_allowed_fallback` 在鉴权层**之前**装到受保护
/// 路由上,所以没带 token 的错误方法仍是 401,不泄露路由表);`/health` 单独装一次。
pub fn router(state: Arc<AppState>) -> Router {
    let protected = Router::new()
        .merge(routes::accounts::router())
        .merge(routes::uploads::router())
        .merge(routes::generate::router())
        .merge(routes::tasks::router())
        .merge(routes::doubao::router())
        .method_not_allowed_fallback(|| async { error::method_not_allowed() })
        .layer(middleware::from_fn_with_state(
            Arc::clone(&state),
            auth::auth_middleware,
        ));
    let api = Router::new()
        .merge(
            routes::health::router()
                .method_not_allowed_fallback(|| async { error::method_not_allowed() }),
        )
        .merge(protected);
    // OpenAI 兼容接口挂在根路径(`/v1/...`、`/health`),鉴权额外接受 `?key=`。
    let openai = routes::openai::router()
        .method_not_allowed_fallback(|| async { error::method_not_allowed() })
        .layer(middleware::from_fn_with_state(
            Arc::clone(&state),
            auth::key_or_bearer_middleware,
        ));
    Router::new()
        .nest(API_PREFIX, api)
        .merge(openai)
        .merge(routes::health::router())
        .fallback(|| async { error::route_not_found() })
        // relay 只绑定本机回环地址，允许本机前端跨端口调用；真正的访问控制仍由
        // Bearer Token 中间件负责。这里不启用 credentials，避免把本地 API 变成
        // 可被浏览器自动带凭据调用的服务。
        .layer(
            CorsLayer::new()
                .allow_origin(Any)
                .allow_methods(Any)
                .allow_headers(Any),
        )
        .layer(middleware::from_fn(access_log))
        .with_state(state)
}

/// `host:port` / `[v6]:port`。
pub fn bind_address(host: &str, port: u16) -> String {
    if host.contains(':') {
        format!("[{host}]:{port}")
    } else {
        format!("{host}:{port}")
    }
}

/// 绑定并运行到 `shutdown` 取消为止。bind 失败(端口被占等)→ `Err`;非 loopback host → `Err`。
pub async fn serve(
    state: Arc<AppState>,
    host: &str,
    port: u16,
    shutdown: CancellationToken,
) -> anyhow::Result<()> {
    if !LOOPBACK_HOSTS.contains(&host) {
        anyhow::bail!("出于安全考虑 HTTP API 只允许监听 loopback,当前 server.host 为 {host}");
    }
    let addr = bind_address(host, port);
    let listener = TcpListener::bind(&addr)
        .await
        .with_context(|| format!("HTTP API 绑定 {addr} 失败,请在设置页更换端口后重启"))?;
    tracing::info!(host, port, "HTTP API 已监听 {host}:{port}");
    state.api_listening.store(true, Ordering::Relaxed);
    let result = axum::serve(listener, router(Arc::clone(&state)))
        .with_graceful_shutdown(async move { shutdown.cancelled().await })
        .await
        .context("HTTP API 运行出错");
    state.api_listening.store(false, Ordering::Relaxed);
    tracing::info!(host, port, "HTTP API 已停止");
    result
}

/// 访问日志中间件(info;5xx 用 warn)。
async fn access_log(req: Request, next: Next) -> Response {
    let method = req.method().to_string();
    let path = req.uri().path().to_string();
    let start = Instant::now();
    let resp = next.run(req).await;
    let status = resp.status().as_u16();
    let duration_ms = start.elapsed().as_millis() as u64;
    if resp.status().is_server_error() {
        tracing::warn!(method, path, status, duration_ms, "HTTP 请求返回服务端错误");
    } else {
        tracing::info!(method, path, status, duration_ms, "HTTP 请求");
    }
    resp
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::Path;
    use std::time::Duration;

    use axum::body::{Body, Bytes};
    use axum::http::{header, Method, Request as HttpRequest, StatusCode};
    use http_body_util::BodyExt;
    use serde_json::{json, Value};
    use tower::ServiceExt;

    use crate::commands::testing::{test_state, test_state_with, FakeBackend};
    use crate::logging::now_ms;
    use crate::storage::Upload;
    use crate::tasks::{
        ConfirmationMode, GenerateRequest, ResultFile, StageEntry, Task, TaskStatus, TaskType,
        WatermarkSource,
    };

    const PNG_1X1: &[u8] = routes::uploads::tests::PNG_1X1;

    struct Harness {
        state: Arc<AppState>,
        token: String,
        _dir: tempfile::TempDir,
    }

    impl Harness {
        async fn new() -> Self {
            Self::with_backend(None).await
        }

        async fn with_fake_accounts(accounts: &[&str]) -> Self {
            Self::with_backend(Some(FakeBackend::new(accounts))).await
        }

        async fn with_backend(backend: Option<Arc<FakeBackend>>) -> Self {
            let dir = tempfile::tempdir().unwrap();
            let state = match backend {
                Some(b) => test_state_with(dir.path(), Some(b)).await,
                None => test_state(dir.path()).await,
            };
            let token = state.config().await.server.bearer_token.clone();
            Self {
                state,
                token,
                _dir: dir,
            }
        }

        fn router(&self) -> Router {
            router(Arc::clone(&self.state))
        }

        async fn send(&self, req: HttpRequest<Body>) -> (StatusCode, Value, axum::http::HeaderMap) {
            let resp = self.router().oneshot(req).await.unwrap();
            let status = resp.status();
            let headers = resp.headers().clone();
            let body = resp.into_body().collect().await.unwrap().to_bytes();
            let value = if body.is_empty() {
                Value::Null
            } else {
                serde_json::from_slice(&body)
                    .unwrap_or(Value::String(String::from_utf8_lossy(&body).into_owned()))
            };
            (status, value, headers)
        }

        async fn raw(&self, req: HttpRequest<Body>) -> (StatusCode, axum::http::HeaderMap, Bytes) {
            let resp = self.router().oneshot(req).await.unwrap();
            let status = resp.status();
            let headers = resp.headers().clone();
            let body = resp.into_body().collect().await.unwrap().to_bytes();
            (status, headers, body)
        }

        fn request(&self, method: Method, path: &str) -> axum::http::request::Builder {
            HttpRequest::builder()
                .method(method)
                .uri(format!("{API_PREFIX}{path}"))
                .header(header::AUTHORIZATION, format!("Bearer {}", self.token))
        }

        async fn get(&self, path: &str) -> (StatusCode, Value) {
            let (s, v, _) = self
                .send(self.request(Method::GET, path).body(Body::empty()).unwrap())
                .await;
            (s, v)
        }

        async fn post_json(&self, path: &str, body: &str) -> (StatusCode, Value) {
            let (s, v, _) = self
                .send(
                    self.request(Method::POST, path)
                        .header(header::CONTENT_TYPE, "application/json")
                        .body(Body::from(body.to_string()))
                        .unwrap(),
                )
                .await;
            (s, v)
        }

        async fn delete(&self, path: &str) -> (StatusCode, Value) {
            let (s, v, _) = self
                .send(
                    self.request(Method::DELETE, path)
                        .body(Body::empty())
                        .unwrap(),
                )
                .await;
            (s, v)
        }

        async fn upload(&self, body: Vec<u8>) -> (StatusCode, Value) {
            let (s, v, _) = self
                .send(
                    self.request(Method::POST, "/uploads")
                        .header(
                            header::CONTENT_TYPE,
                            format!("multipart/form-data; boundary={BOUNDARY}"),
                        )
                        .header(header::CONTENT_LENGTH, body.len())
                        .body(Body::from(body))
                        .unwrap(),
                )
                .await;
            (s, v)
        }
    }

    const BOUNDARY: &str = "XBOUNDARYxyz";

    /// `(字段名, Some((文件名, Content-Type)) 或 None, 内容)`。
    type Part<'a> = (&'a str, Option<(&'a str, &'a str)>, &'a [u8]);

    fn multipart(parts: &[Part<'_>]) -> Vec<u8> {
        let mut out = Vec::new();
        for (name, file, data) in parts {
            out.extend_from_slice(format!("--{BOUNDARY}\r\n").as_bytes());
            match file {
                Some((filename, ct)) => out.extend_from_slice(
                    format!(
                        "Content-Disposition: form-data; name=\"{name}\"; filename=\"{filename}\"\r\nContent-Type: {ct}\r\n\r\n"
                    )
                    .as_bytes(),
                ),
                None => out.extend_from_slice(
                    format!("Content-Disposition: form-data; name=\"{name}\"\r\n\r\n").as_bytes(),
                ),
            }
            out.extend_from_slice(data);
            out.extend_from_slice(b"\r\n");
        }
        out.extend_from_slice(format!("--{BOUNDARY}--\r\n").as_bytes());
        out
    }

    fn generate_body(account: &str) -> String {
        json!({
            "account": account,
            "type": "image",
            "model_name": "Seedream 5.0 Pro",
            "prompt": "一只蓝色小鸟",
            "ratio": "1:1"
        })
        .to_string()
    }

    fn completed_task(id: &str, expires_at: i64) -> Task {
        let now = now_ms();
        Task {
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
            conversation_id: Some("1".into()),
            submission_id: Some("2".into()),
            local_message_id: None,
            watermark_source: Some(WatermarkSource::ConversationOriginal),
            result_files: Some(vec![ResultFile {
                index: 0,
                filename: "image_1.png".into(),
                relative_path: format!("results/{id}/image_1.png"),
                mime: "image/png".into(),
                bytes: PNG_1X1.len() as u64,
                width: Some(1),
                height: Some(1),
                duration_ms: None,
                source_url: None,
            }]),
            created_at: now - 1000,
            started_at: Some(now - 900),
            completed_at: Some(now - 100),
            expires_at,
            stage_history: vec![StageEntry {
                stage: "pending".into(),
                message: None,
                ts: now - 1000,
            }],
            queue_position: None,
        }
    }

    async fn write_result(root: &Path, id: &str) {
        let dir = root.join("results").join(id);
        tokio::fs::create_dir_all(&dir).await.unwrap();
        tokio::fs::write(dir.join("image_1.png"), PNG_1X1)
            .await
            .unwrap();
    }

    // ---------------- health / auth ----------------

    #[tokio::test]
    async fn health_is_public_and_reports_version_and_uptime() {
        let h = Harness::new().await;
        let (status, v, headers) = h
            .send(
                HttpRequest::builder()
                    .uri("/api/v1/health")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["status"], "ok");
        assert_eq!(v["version"], env!("CARGO_PKG_VERSION"));
        assert!(v["uptime_seconds"].as_u64().is_some());
        assert!(headers[header::CONTENT_TYPE]
            .to_str()
            .unwrap()
            .starts_with("application/json"));
    }

    #[tokio::test]
    async fn protected_routes_require_a_valid_bearer_token() {
        let h = Harness::new().await;
        let (status, v, _) = h
            .send(
                HttpRequest::builder()
                    .uri("/api/v1/accounts")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::UNAUTHORIZED);
        assert_eq!(v["error"]["code"], "auth_missing_token");

        let (status, v, _) = h
            .send(
                HttpRequest::builder()
                    .uri("/api/v1/accounts")
                    .header(header::AUTHORIZATION, "Bearer nope")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::UNAUTHORIZED);
        assert_eq!(v["error"]["code"], "auth_invalid_token");

        let (status, v, _) = h
            .send(
                HttpRequest::builder()
                    .uri("/api/v1/tasks")
                    .header(header::AUTHORIZATION, format!("Basic {}", h.token))
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::UNAUTHORIZED);
        assert_eq!(v["error"]["code"], "auth_invalid_token");

        // 正确 token:豆包没连上 → 503 doubao_*
        let (status, v) = h.get("/accounts").await;
        assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE, "{v}");
        assert!(v["error"]["code"].as_str().unwrap().starts_with("doubao_"));
    }

    #[tokio::test]
    async fn token_change_in_config_takes_effect_immediately() {
        let h = Harness::new().await;
        h.state.config.write().await.server.bearer_token = "new-token".into();
        let (status, v) = h.get("/tasks").await;
        assert_eq!(status, StatusCode::UNAUTHORIZED);
        assert_eq!(v["error"]["code"], "auth_invalid_token");
        let (status, _, _) = h
            .send(
                HttpRequest::builder()
                    .uri("/api/v1/tasks")
                    .header(header::AUTHORIZATION, "Bearer new-token")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::OK);
    }

    #[tokio::test]
    async fn unknown_route_is_json_404() {
        let h = Harness::new().await;
        let (status, v) = h.get("/nope").await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "request_malformed");
        let (status, _, _) = h
            .send(
                HttpRequest::builder()
                    .uri("/other")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn wrong_method_on_known_route_is_json_405() {
        let h = Harness::new().await;
        // 免鉴权的 /health
        let (status, v, headers) = h
            .send(
                HttpRequest::builder()
                    .method(Method::POST)
                    .uri("/api/v1/health")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::METHOD_NOT_ALLOWED);
        assert_eq!(headers["content-type"], "application/json");
        assert_eq!(v["error"]["code"], "request_malformed");
        assert!(v["error"]["message"].is_string());
        assert_eq!(v.as_object().unwrap().len(), 1);

        // 受保护路由:带 token → 405 JSON
        let (status, v) = h.post_json("/tasks", "{}").await;
        assert_eq!(status, StatusCode::METHOD_NOT_ALLOWED, "{v}");
        assert_eq!(v["error"]["code"], "request_malformed");
        let (status, v, _) = h
            .send(
                h.request(Method::PUT, "/accounts")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::METHOD_NOT_ALLOWED, "{v}");
        assert_eq!(v["error"]["code"], "request_malformed");

        // 不带 token 的错误方法仍先吃 401,不泄露路由表
        let (status, v, _) = h
            .send(
                HttpRequest::builder()
                    .method(Method::PUT)
                    .uri("/api/v1/accounts")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::UNAUTHORIZED, "{v}");
        assert_eq!(v["error"]["code"], "auth_missing_token");
    }

    // ---------------- generate / tasks ----------------

    #[tokio::test]
    async fn generate_validation_errors_are_400_with_codes() {
        let h = Harness::new().await;
        let (status, v) = h.post_json("/generate", r#"{"type":"image"}"#).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_missing_field");
        assert_eq!(v["error"]["details"]["field"], "account");

        let (status, v) = h.post_json("/generate", "{oops").await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_malformed");

        let (status, v) = h
            .post_json(
                "/generate",
                r#"{"account":"a","type":"image","model_name":"Nope","prompt":"p","ratio":"1:1"}"#,
            )
            .await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_invalid_enum");
        assert_eq!(v["error"]["details"]["field"], "model_name");

        // 校验通过但豆包没连上 → 503
        let (status, v) = h.post_json("/generate", &generate_body("张三")).await;
        assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE, "{v}");
    }

    #[tokio::test]
    async fn generate_list_detail_cancel_confirm_round_trip() {
        let h = Harness::with_fake_accounts(&["张三"]).await;

        let (status, v) = h.post_json("/generate", &generate_body("李四")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "account_not_found");
        assert_eq!(v["error"]["details"]["available"], json!(["张三"]));

        let (status, v) = h.post_json("/generate", &generate_body("张三")).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        let id = v["task_id"].as_str().unwrap().to_string();
        assert_eq!(id.len(), 26);
        assert_eq!(v["status"], "pending");
        assert_eq!(v["queue_position"], 1);
        assert!(v["created_at"].as_i64().unwrap() > 0);

        let (status, v) = h.get("/tasks").await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["tasks"][0]["task_id"], id);
        assert_eq!(v["tasks"][0]["type"], "image");
        assert_eq!(v["tasks"][0]["account"], "张三");
        assert!(v["next_cursor"].is_null());

        let (status, v) = h
            .get("/tasks?status=pending,preparing&account=张三&limit=5")
            .await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["tasks"].as_array().unwrap().len(), 1);
        let (status, v) = h.get("/tasks?status=completed").await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["tasks"].as_array().unwrap().len(), 0);
        let (status, v) = h.get("/tasks?status=done").await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_invalid_enum");
        let (status, v) = h.get("/tasks?limit=abc").await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_malformed");

        let (status, v) = h.get(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["task_id"], id);
        assert_eq!(v["status"], "pending");
        assert_eq!(v["account"], "张三");
        assert_eq!(v["type"], "image");
        assert_eq!(v["queue_position"], 1);
        assert_eq!(v["request"]["prompt"], "一只蓝色小鸟");
        assert_eq!(v["stage_history"][0]["stage"], "pending");

        let (status, v) = h.post_json(&format!("/tasks/{id}/confirm"), "").await;
        assert_eq!(status, StatusCode::CONFLICT);
        assert_eq!(v["error"]["code"], "task_not_waiting_confirm");
        assert_eq!(v["error"]["details"]["current_status"], "pending");

        let (status, v) = h.delete(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v, json!({ "ok": true, "cancelled_from_status": "pending" }));
        let (status, v) = h.delete(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::CONFLICT);
        assert_eq!(v["error"]["code"], "task_not_cancellable");

        let (status, v) = h.get(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["status"], "cancelled");
        assert_eq!(v["error"]["code"], "generation_cancelled");
    }

    #[tokio::test]
    async fn unknown_task_is_404_everywhere() {
        let h = Harness::new().await;
        for (method, path) in [
            (Method::GET, "/tasks/x"),
            (Method::DELETE, "/tasks/x"),
            (Method::POST, "/tasks/x/confirm"),
            (Method::GET, "/tasks/x/download/0"),
        ] {
            let (status, v, _) = h
                .send(h.request(method.clone(), path).body(Body::empty()).unwrap())
                .await;
            assert_eq!(status, StatusCode::NOT_FOUND, "{method} {path}: {v}");
            assert_eq!(v["error"]["code"], "task_not_found", "{method} {path}");
            assert_eq!(v["error"]["details"]["task_id"], "x");
        }
    }

    // ---------------- uploads ----------------

    #[tokio::test]
    async fn upload_png_then_get_and_reference_in_generate() {
        let h = Harness::with_fake_accounts(&["张三"]).await;
        let body = multipart(&[("file", Some(("a.png", "image/png")), PNG_1X1)]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        let id = v["upload_id"].as_str().unwrap().to_string();
        assert_eq!(id.len(), 26);
        assert_eq!(v["mime"], "image/png");
        assert_eq!(v["bytes"], PNG_1X1.len());
        let expires = v["expires_at"].as_i64().unwrap();
        assert!(expires > now_ms() + 3_500_000 && expires <= now_ms() + 3_600_000);
        assert_eq!(v.as_object().unwrap().len(), 4);
        let on_disk = h.state.storage.upload_path(&id, "png");
        assert_eq!(tokio::fs::read(&on_disk).await.unwrap(), PNG_1X1);

        let (status, got) = h.get(&format!("/uploads/{id}")).await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(got, v);

        let req = json!({
            "account": "张三", "type": "image", "model_name": "Seedream 5.0 Pro",
            "prompt": "p", "ratio": "1:1", "reference_files": [id]
        });
        let (status, v) = h.post_json("/generate", &req.to_string()).await;
        assert_eq!(status, StatusCode::OK, "{v}");

        let (status, v) = h.get("/uploads/01NOPE").await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "upload_not_found");
        let req = json!({
            "account": "张三", "type": "image", "model_name": "Seedream 5.0 Pro",
            "prompt": "p", "ratio": "1:1", "reference_files": ["01NOPE"]
        });
        let (status, v) = h.post_json("/generate", &req.to_string()).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "upload_not_found");
    }

    #[tokio::test]
    async fn upload_mp4_kind_and_mime_rules() {
        let h = Harness::new().await;
        let mp4 = routes::uploads::tests::mp4_bytes();
        let body = multipart(&[
            ("kind", None, b"video_reference"),
            ("file", Some(("a.mp4", "video/mp4")), &mp4),
        ]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        assert_eq!(v["mime"], "video/mp4");

        // 内容是 PNG 但声明 kind=video_reference
        let body = multipart(&[
            ("file", Some(("a.png", "image/png")), PNG_1X1),
            ("kind", None, b"video_reference"),
        ]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "upload_mime_not_supported");
        assert_eq!(v["error"]["details"]["mime"], "image/png");

        // 文本文件冒充 PNG
        let body = multipart(&[("file", Some(("a.png", "image/png")), b"not an image")]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "upload_mime_not_supported");
        assert_eq!(v["error"]["details"]["mime"], "image/png");
        assert_eq!(
            v["error"]["details"]["allowed"],
            json!(["image/png", "image/jpeg", "image/webp", "video/mp4"])
        );

        let body = multipart(&[
            ("kind", None, b"audio"),
            ("file", Some(("a.png", "image/png")), PNG_1X1),
        ]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_invalid_enum");
        assert_eq!(v["error"]["details"]["field"], "kind");

        let body = multipart(&[("other", None, b"x")]);
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_missing_field");
        assert_eq!(v["error"]["details"]["field"], "file");

        let (status, v) = h.post_json("/uploads", "{}").await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_malformed");
    }

    #[tokio::test]
    async fn upload_over_50mb_is_400_upload_too_large() {
        let h = Harness::new().await;
        let max = routes::uploads::UPLOAD_MAX_BYTES;
        let big = vec![b'a'; max as usize + 1];
        let body = multipart(&[("file", Some(("a.png", "image/png")), &big)]);
        let body_len = body.len() as u64;
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST, "{v}");
        assert_eq!(v["error"]["code"], "upload_too_large");
        assert_eq!(v["error"]["details"]["max_bytes"], max);
        // bytes 是文件字节数(刚过线处停止计数),不是整个 multipart body 的 Content-Length
        let bytes = v["error"]["details"]["bytes"].as_u64().unwrap();
        assert!(bytes > max && bytes <= max + 1, "bytes={bytes}");
        assert!(bytes < body_len);
        let mut entries = tokio::fs::read_dir(h.state.storage.root.join("uploads"))
            .await
            .unwrap();
        assert!(entries.next_entry().await.unwrap().is_none(), "没有落盘");

        // body 超过 DefaultBodyLimit(50MB + 1MB 余量)→ multipart 413 → 同一错误;
        // file 字段还没读到任何字节,bytes 退回 Content-Length
        let huge = vec![b'a'; (max + 2 * 1024 * 1024) as usize];
        let body = multipart(&[("file", Some(("a.png", "image/png")), &huge)]);
        let body_len = body.len() as u64;
        let (status, v) = h.upload(body).await;
        assert_eq!(status, StatusCode::BAD_REQUEST, "{v}");
        assert_eq!(v["error"]["code"], "upload_too_large");
        assert_eq!(v["error"]["details"]["bytes"], body_len);
        assert_eq!(v["error"]["details"]["max_bytes"], max);
    }

    #[tokio::test]
    async fn expired_upload_is_404() {
        let h = Harness::new().await;
        let now = now_ms();
        h.state
            .storage
            .insert_upload(&Upload {
                id: "01OLD".into(),
                mime: "image/png".into(),
                bytes: 1,
                path: h.state.storage.upload_path("01OLD", "png"),
                created_at: now - 10_000,
                expires_at: now - 1,
            })
            .await
            .unwrap();
        let (status, v) = h.get("/uploads/01OLD").await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "upload_not_found");
        assert_eq!(v["error"]["details"]["expired_at"], now - 1);
    }

    // ---------------- download ----------------

    #[tokio::test]
    async fn download_serves_file_with_headers_and_supports_range() {
        let h = Harness::new().await;
        let id = "01HQZ9ABCDEFGHJKMNPQRSTVWX";
        write_result(&h.state.storage.root, id).await;
        h.state
            .storage
            .insert_task(&completed_task(id, now_ms() + 60_000))
            .await
            .unwrap();

        let (status, v) = h.get(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(
            v["result"]["files"][0]["download_url"],
            format!("/api/v1/tasks/{id}/download/0")
        );

        let (status, headers, body) = h
            .raw(
                h.request(Method::GET, &format!("/tasks/{id}/download/0"))
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(headers[header::CONTENT_TYPE], "image/png");
        assert_eq!(
            headers[header::CONTENT_DISPOSITION],
            format!("attachment; filename=\"task-{id}-0.png\"")
        );
        assert_eq!(
            headers[header::CONTENT_LENGTH],
            PNG_1X1.len().to_string().as_str()
        );
        assert_eq!(&body[..], PNG_1X1);

        let (status, headers, body) = h
            .raw(
                h.request(Method::GET, &format!("/tasks/{id}/download/0"))
                    .header(header::RANGE, "bytes=0-3")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await;
        assert_eq!(status, StatusCode::PARTIAL_CONTENT);
        assert_eq!(&body[..], &PNG_1X1[..4]);
        assert_eq!(headers[header::CONTENT_LENGTH], "4");
        assert!(headers[header::CONTENT_RANGE]
            .to_str()
            .unwrap()
            .starts_with("bytes 0-3/"));

        let (status, v) = h.get(&format!("/tasks/{id}/download/5")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_index_out_of_range");
        assert_eq!(v["error"]["details"]["requested"], 5);
        assert_eq!(v["error"]["details"]["available"], 1);

        let (status, v) = h.get(&format!("/tasks/{id}/download/abc")).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_malformed");

        // 文件被清掉 → result_expired
        tokio::fs::remove_file(h.state.storage.result_dir(id).join("image_1.png"))
            .await
            .unwrap();
        let (status, v) = h.get(&format!("/tasks/{id}/download/0")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_expired");
    }

    #[tokio::test]
    async fn download_rejects_expired_and_unfinished_tasks() {
        let h = Harness::new().await;
        let expired = "01EXPIREDAAAAAAAAAAAAAAAAA";
        write_result(&h.state.storage.root, expired).await;
        h.state
            .storage
            .insert_task(&completed_task(expired, now_ms() - 1))
            .await
            .unwrap();
        let (status, v) = h.get(&format!("/tasks/{expired}/download/0")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_expired");
        assert_eq!(v["error"]["details"]["task_id"], expired);

        let pending = "01PENDINGAAAAAAAAAAAAAAAAA";
        let mut t = completed_task(pending, now_ms() + 60_000);
        t.status = TaskStatus::Generating;
        t.result_files = None;
        h.state.storage.insert_task(&t).await.unwrap();
        let (status, v) = h.get(&format!("/tasks/{pending}/download/0")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_not_available");
        assert_eq!(v["error"]["details"]["index"], 0);
        assert!(v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("请等任务完成后再下载"));

        // 失败 / 取消的任务永远不会完成:同一个码,但提示不是"等完成"
        for (id, status) in [
            ("01FAILEDAAAAAAAAAAAAAAAAAA", TaskStatus::Failed),
            ("01CANCELLEDAAAAAAAAAAAAAAA", TaskStatus::Cancelled),
        ] {
            let mut t = completed_task(id, now_ms() + 60_000);
            t.status = status;
            t.result_files = None;
            h.state.storage.insert_task(&t).await.unwrap();
            let (status, v) = h.get(&format!("/tasks/{id}/download/0")).await;
            assert_eq!(status, StatusCode::NOT_FOUND);
            assert_eq!(v["error"]["code"], "result_not_available");
            assert_eq!(v["error"]["details"]["task_id"], id);
            let msg = v["error"]["message"].as_str().unwrap();
            assert!(msg.contains("没有可下载的结果"), "{msg}");
            assert!(!msg.contains("请等任务完成"), "{msg}");
        }
    }

    /// GC 步骤 1 把过期任务的 `result_files` 置 NULL 并删目录:之后下载必须是
    /// `result_expired{task_id, expired_at}`,而不是 `result_not_available`。
    #[tokio::test]
    async fn download_after_gc_reports_result_expired() {
        let h = Harness::new().await;
        let id = "01GCEDAAAAAAAAAAAAAAAAAAAA";
        let expires_at = now_ms() - 1;
        write_result(&h.state.storage.root, id).await;
        h.state
            .storage
            .insert_task(&completed_task(id, expires_at))
            .await
            .unwrap();
        h.state
            .storage
            .gc(now_ms(), &crate::config::StorageConfig::default())
            .await
            .unwrap();
        let after = h.state.storage.get_task(id).await.unwrap().unwrap();
        assert_eq!(after.status, TaskStatus::Completed);
        assert!(after.result_files.is_none(), "GC 应清空 result_files");

        let (status, v) = h.get(&format!("/tasks/{id}/download/0")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_expired");
        assert_eq!(v["error"]["details"]["task_id"], id);
        assert_eq!(v["error"]["details"]["expired_at"], expires_at);

        // 完成但文件列表为空(未到期)同样按已清理处理
        let empty = "01EMPTYFILESAAAAAAAAAAAAAA";
        let mut t = completed_task(empty, now_ms() + 60_000);
        t.result_files = Some(vec![]);
        h.state.storage.insert_task(&t).await.unwrap();
        let (status, v) = h.get(&format!("/tasks/{empty}/download/0")).await;
        assert_eq!(status, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "result_expired");
    }

    // ---------------- accounts / doubao ----------------

    #[tokio::test]
    async fn accounts_switch_validates_body_and_doubao_endpoints_respond() {
        let h = Harness::new().await;
        let (status, v) = h.post_json("/accounts/switch", "{}").await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_missing_field");
        assert_eq!(v["error"]["details"]["field"], "target");
        let (status, v) = h
            .post_json("/accounts/switch", r#"{"target":"李四"}"#)
            .await;
        assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE, "{v}");

        let (status, v) = h.get("/doubao/status").await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["running"], false);
        assert_eq!(v["cdp_connected"], false);
        assert_eq!(v["cdp_port"], 1);
        assert_eq!(v["chat_targets"], 0);
        assert!(v["last_check_at"].is_i64());

        let (status, v) = h.post_json("/doubao/reconnect", "").await;
        assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE, "{v}");
        assert!(v["error"]["code"].as_str().unwrap().starts_with("doubao_"));
        let (status, v) = h.post_json("/doubao/reconnect", r#"{"force":"yes"}"#).await;
        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_malformed");
    }

    #[tokio::test]
    async fn switch_is_refused_while_a_task_is_running() {
        let h = Harness::with_fake_accounts(&["张三", "李四"]).await;
        // 真调度循环 + FakeBackend 的 submit 挂起到取消 → 任务停在 preparing
        h.state.scheduler.spawn_loop();
        let (status, v) = h.post_json("/generate", &generate_body("张三")).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        let id = v["task_id"].as_str().unwrap().to_string();
        let mut running = Vec::new();
        for _ in 0..200 {
            running = h.state.scheduler.running_task_ids(None);
            if !running.is_empty() {
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        assert_eq!(running, vec![id.clone()], "任务应已被派发");
        assert_eq!(
            h.state.scheduler.running_task_ids(Some("张三")),
            vec![id.clone()]
        );
        assert!(h.state.scheduler.running_task_ids(Some("李四")).is_empty());

        let (status, v) = h
            .post_json("/accounts/switch", r#"{"target":"李四"}"#)
            .await;
        assert_eq!(status, StatusCode::CONFLICT, "{v}");
        assert_eq!(v["error"]["code"], "account_busy");
        assert_eq!(v["error"]["details"]["account"], "李四");
        assert_eq!(v["error"]["details"]["running_task_ids"], json!([id]));

        let (status, v) = h.delete(&format!("/tasks/{id}")).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        assert_eq!(v["cancelled_from_status"], "preparing");
    }

    /// 真豆包(9707)只读冒烟:`/doubao/status` 与 `/accounts` 经路由返回真实数据,不提交任务。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_http_status_and_accounts() {
        use crate::config::Config;
        use crate::doubao::DoubaoConnection;
        use crate::logging::LogBuffer;
        use crate::state::CdpLogBuffer;
        use crate::storage::Storage;
        use crate::tasks::DoubaoBackend;

        let _menu = crate::doubao::LIVE_MENU_LOCK.lock().await;
        let dir = tempfile::tempdir().unwrap();
        let platform = crate::platform::create_platform();
        let config_path = dir.path().join("config.toml");
        let config = Config::load_or_create(&config_path).unwrap();
        let token = config.server.bearer_token.clone();
        let doubao = DoubaoConnection::new(Arc::clone(&platform), "127.0.0.1".into(), 9707);
        let storage = Storage::open(dir.path()).await.unwrap();
        let backend = DoubaoBackend::new(Arc::clone(&doubao));
        let state = AppState::new(
            config,
            config_path,
            platform,
            doubao,
            LogBuffer::new(),
            CdpLogBuffer::new(false),
            storage,
            backend,
        );
        let action = state.doubao.connect(false).await.unwrap();
        println!("connect: {action:?}");
        let h = Harness {
            state,
            token,
            _dir: dir,
        };

        let started = std::time::Instant::now();
        let (status, v) = h.get("/doubao/status").await;
        println!(
            "GET /doubao/status ({}ms): {v}",
            started.elapsed().as_millis()
        );
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["running"], true);
        assert_eq!(v["cdp_connected"], true);
        assert_eq!(v["cdp_port"], 9707);
        assert_eq!(v["chat_targets"], 1);
        assert!(v["doubao_version_detected"].is_string());

        let started = std::time::Instant::now();
        let (status, v) = h.get("/accounts").await;
        println!("GET /accounts ({}ms): {v}", started.elapsed().as_millis());
        assert_eq!(status, StatusCode::OK);
        assert!(v["current"].is_string());
        let accounts = v["accounts"].as_array().unwrap();
        assert!(!accounts.is_empty());
        assert!(accounts[0]["name"].is_string());
        assert!(accounts[0]["is_current"].is_boolean());
        assert!(v["read_at"].is_i64());

        let (status, v) = h.get("/health").await;
        assert_eq!(status, StatusCode::OK);
        assert_eq!(v["status"], "ok");
    }

    // ---------------- serve ----------------

    fn free_port() -> u16 {
        std::net::TcpListener::bind("127.0.0.1:0")
            .unwrap()
            .local_addr()
            .unwrap()
            .port()
    }

    #[tokio::test]
    async fn serve_binds_loopback_answers_health_and_stops_on_shutdown() {
        let h = Harness::new().await;
        let port = free_port();
        let shutdown = CancellationToken::new();
        let task = tokio::spawn(serve(
            Arc::clone(&h.state),
            "127.0.0.1",
            port,
            shutdown.clone(),
        ));
        let client = reqwest::Client::new();
        let url = format!("http://127.0.0.1:{port}/api/v1/health");
        let mut body = None;
        for _ in 0..100 {
            if let Ok(resp) = client.get(&url).send().await {
                body = Some(resp.json::<Value>().await.unwrap());
                break;
            }
            tokio::time::sleep(Duration::from_millis(20)).await;
        }
        assert_eq!(body.unwrap()["status"], "ok");
        assert!(h.state.api_listening.load(Ordering::Relaxed));

        let resp = client
            .get(format!("http://127.0.0.1:{port}/api/v1/tasks"))
            .send()
            .await
            .unwrap();
        assert_eq!(resp.status().as_u16(), 401);

        shutdown.cancel();
        let result = tokio::time::timeout(Duration::from_secs(5), task)
            .await
            .expect("serve 应在 shutdown 后退出")
            .unwrap();
        assert!(result.is_ok(), "{result:?}");
        assert!(!h.state.api_listening.load(Ordering::Relaxed));
    }

    #[tokio::test]
    async fn serve_fails_on_occupied_port_and_non_loopback_host() {
        let h = Harness::new().await;
        let holder = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
        let port = holder.local_addr().unwrap().port();
        let err = serve(
            Arc::clone(&h.state),
            "127.0.0.1",
            port,
            CancellationToken::new(),
        )
        .await
        .unwrap_err();
        assert!(format!("{err:#}").contains(&port.to_string()), "{err:#}");
        assert!(!h.state.api_listening.load(Ordering::Relaxed));

        let err = serve(
            Arc::clone(&h.state),
            "0.0.0.0",
            free_port(),
            CancellationToken::new(),
        )
        .await
        .unwrap_err();
        assert!(err.to_string().contains("loopback"));
        assert_eq!(bind_address("::1", 8787), "[::1]:8787");
        assert_eq!(bind_address("localhost", 1), "localhost:1");
    }

    // ---- OpenAI 兼容接口(根路径) ----

    async fn root(
        h: &Harness,
        method: Method,
        uri: &str,
        body: Body,
        ct: Option<String>,
    ) -> (StatusCode, Value) {
        let mut b = HttpRequest::builder().method(method).uri(uri);
        if let Some(ct) = ct {
            b = b.header(header::CONTENT_TYPE, ct);
        }
        let (s, v, _) = h.send(b.body(body).unwrap()).await;
        (s, v)
    }

    #[tokio::test]
    async fn openai_models_health_and_key_auth() {
        let h = Harness::new().await;
        let (s, v) = root(&h, Method::GET, "/health", Body::empty(), None).await;
        assert_eq!(s, StatusCode::OK, "根路径 /health 免鉴权");
        assert_eq!(v["status"], "ok");

        let (s, v) = root(&h, Method::GET, "/v1/models", Body::empty(), None).await;
        assert_eq!(s, StatusCode::UNAUTHORIZED);
        assert_eq!(v["error"]["code"], "auth_missing_token");
        let (s, _) = root(&h, Method::GET, "/v1/models?key=wrong", Body::empty(), None).await;
        assert_eq!(s, StatusCode::UNAUTHORIZED);

        let uri = format!("/v1/models?key={}", h.token);
        let (s, v) = root(&h, Method::GET, &uri, Body::empty(), None).await;
        assert_eq!(s, StatusCode::OK, "{v}");
        let ids: Vec<&str> = v["data"]
            .as_array()
            .unwrap()
            .iter()
            .map(|m| m["id"].as_str().unwrap())
            .collect();
        assert_eq!(
            ids,
            [
                "doubao-image-4.0",
                "doubao-image-4.5",
                "doubao-image-5.0-flash",
                "doubao-video-mini",
                "doubao-video-fast"
            ]
        );
    }

    #[tokio::test]
    async fn openai_ref_image_accepts_data_uri_and_url() {
        use base64::Engine;
        let h = Harness::with_fake_accounts(&["张三"]).await;
        // 本地图片服务器,模拟调用方给的图片网址
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        let app = Router::new()
            .route("/a.png", axum::routing::get(|| async { PNG_1X1 }))
            .route(
                "/missing.png",
                axum::routing::get(|| async { StatusCode::NOT_FOUND }),
            );
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });

        let data_uri = format!(
            "data:image/png;base64,{}",
            base64::engine::general_purpose::STANDARD.encode(PNG_1X1)
        );
        let body = json!({
            "prompt": "让画面动起来",
            "ref_image": [data_uri, format!("http://{addr}/a.png")],
        })
        .to_string();
        let uri = format!("/v1/video/generations?key={}", h.token);
        let ct = Some("application/json".to_string());
        let (s, v) = root(&h, Method::POST, &uri, Body::from(body), ct.clone()).await;
        assert_eq!(s, StatusCode::OK, "{v}");
        let task = h
            .state
            .scheduler
            .get_task(v["id"].as_str().unwrap())
            .await
            .unwrap()
            .unwrap();
        assert_eq!(task.request.reference_files.len(), 2);
        for id in &task.request.reference_files {
            let u = h.state.storage.get_upload(id).await.unwrap().unwrap();
            assert_eq!(u.mime, "image/png");
        }

        for (bad, reason_part) in [
            ("data:image/png;base64,***".to_string(), "base64"),
            (format!("http://{addr}/missing.png"), "404"),
            ("ftp://x/a.png".to_string(), "http"),
            ("data:text/plain;base64,aGk=".to_string(), ""),
        ] {
            let body = json!({ "prompt": "x", "ref_image": bad }).to_string();
            let (s, v) = root(&h, Method::POST, &uri, Body::from(body), ct.clone()).await;
            assert_eq!(s, StatusCode::BAD_REQUEST, "{bad}: {v}");
            if reason_part.is_empty() {
                assert_eq!(v["error"]["code"], "upload_mime_not_supported", "{v}");
            } else {
                assert_eq!(v["error"]["details"]["field"], "ref_image[0]", "{v}");
                assert!(
                    v["error"]["details"]["reason"]
                        .as_str()
                        .unwrap()
                        .contains(reason_part),
                    "{v}"
                );
            }
        }

        let upload = format!("/v1/images/upload?key={}", h.token);
        let (s, _) = root(&h, Method::POST, &upload, Body::empty(), None).await;
        assert_eq!(s, StatusCode::NOT_FOUND, "上传接口已移除");
    }

    #[tokio::test]
    async fn openai_accounts_list_and_switch() {
        let h = Harness::with_fake_accounts(&["张三", "李四"]).await;
        let key = format!("?key={}", h.token);
        let (s, v) = root(
            &h,
            Method::GET,
            &format!("/v1/accounts{key}"),
            Body::empty(),
            None,
        )
        .await;
        assert_eq!(s, StatusCode::OK, "{v}");
        assert_eq!(v["object"], "list");
        assert_eq!(v["current"], "张三");
        assert_eq!(v["data"].as_array().unwrap().len(), 2);
        assert_eq!(v["data"][1]["name"], "李四");
        assert_eq!(v["data"][0]["is_current"], true);
        assert!(v["data"][0].get("label").is_none());
        let (s, _) = root(&h, Method::GET, "/v1/accounts", Body::empty(), None).await;
        assert_eq!(s, StatusCode::UNAUTHORIZED);

        let ct = Some("application/json".to_string());
        let switch = format!("/v1/accounts/switch{key}");
        let (s, v) = root(&h, Method::POST, &switch, Body::from("{}"), ct.clone()).await;
        assert_eq!(s, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_missing_field");
        assert_eq!(v["error"]["details"]["field"], "account");

        // 有任务在跑时拒绝切换(与 /api/v1/accounts/switch 同一守卫)
        h.state.scheduler.spawn_loop();
        let (status, v) = h.post_json("/generate", &generate_body("张三")).await;
        assert_eq!(status, StatusCode::OK, "{v}");
        let id = v["task_id"].as_str().unwrap().to_string();
        for _ in 0..200 {
            if !h.state.scheduler.running_task_ids(None).is_empty() {
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let (s, v) = root(
            &h,
            Method::POST,
            &switch,
            Body::from(r#"{"account":"李四"}"#),
            ct,
        )
        .await;
        assert_eq!(s, StatusCode::CONFLICT, "{v}");
        assert_eq!(v["error"]["code"], "account_busy");
        h.state.scheduler.cancel(&id).await.unwrap();
    }

    #[tokio::test]
    async fn openai_video_is_async_and_queryable() {
        let h = Harness::with_fake_accounts(&["张三"]).await;
        let auth = format!("Bearer {}", h.token);
        let send = |method: Method, uri: &str, body: &str| {
            HttpRequest::builder()
                .method(method)
                .uri(uri)
                .header(header::AUTHORIZATION, auth.clone())
                .header(header::CONTENT_TYPE, "application/json")
                .body(Body::from(body.to_string()))
                .unwrap()
        };
        let (s, v, _) = h
            .send(send(
                Method::POST,
                "/v1/video/generations",
                r#"{"prompt":"柴犬奔跑"}"#,
            ))
            .await;
        assert_eq!(s, StatusCode::OK, "{v}");
        assert_eq!(v["object"], "video.generation");
        assert_eq!(v["status"], "pending");
        let id = v["id"].as_str().unwrap().to_string();
        let task = h.state.scheduler.get_task(&id).await.unwrap().unwrap();
        assert_eq!(task.request.account, "张三", "未传 account 用当前账号");
        assert_eq!(task.request.model_name, "Seedance 2.0 Fast");
        assert_eq!(task.request.duration, Some(5));

        let (s, v, _) = h
            .send(send(
                Method::GET,
                &format!("/v1/video/generations/{id}"),
                "",
            ))
            .await;
        assert_eq!(s, StatusCode::OK, "{v}");
        assert_eq!(v["id"], id.as_str());
        assert_eq!(v["status"], "pending");
        assert!(v.get("data").is_none());

        let (s, v, _) = h
            .send(send(
                Method::POST,
                "/v1/video/generations",
                r#"{"prompt":"x","model":"doubao-image-4.5"}"#,
            ))
            .await;
        assert_eq!(s, StatusCode::BAD_REQUEST);
        assert_eq!(v["error"]["code"], "request_invalid_enum");

        let (s, v, _) = h
            .send(send(Method::GET, "/v1/video/generations/NOPE", ""))
            .await;
        assert_eq!(s, StatusCode::NOT_FOUND);
        assert_eq!(v["error"]["code"], "task_not_found");
    }

    #[tokio::test]
    async fn openai_image_sync_returns_task_error_when_cancelled() {
        let h = Harness::with_fake_accounts(&["张三"]).await;
        h.state.scheduler.spawn_loop();
        let req = HttpRequest::builder()
            .method(Method::POST)
            .uri("/v1/images/generations")
            .header(header::AUTHORIZATION, format!("Bearer {}", h.token))
            .header(header::CONTENT_TYPE, "application/json")
            .body(Body::from(r#"{"prompt":"一只猫","size":"1792x1024"}"#))
            .unwrap();
        let router = h.router();
        let call = tokio::spawn(async move { router.oneshot(req).await.unwrap() });
        // FakeBackend 的 submit 挂起到取消:等任务被派发后取消,同步请求应带任务错误返回
        let mut id = None;
        for _ in 0..200 {
            if let Some(t) = h.state.scheduler.running_task_ids(None).first() {
                id = Some(t.clone());
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
        let id = id.expect("任务应已被派发");
        let task = h.state.scheduler.get_task(&id).await.unwrap().unwrap();
        assert_eq!(task.request.ratio, "16:9");
        assert_eq!(task.request.model_name, "Seedream 4.5");
        h.state.scheduler.cancel(&id).await.unwrap();
        let resp = tokio::time::timeout(Duration::from_secs(10), call)
            .await
            .expect("同步请求应在任务终态后返回")
            .unwrap();
        assert_eq!(resp.status(), StatusCode::BAD_GATEWAY);
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let v: Value = serde_json::from_slice(&body).unwrap();
        assert_eq!(v["error"]["code"], "generation_cancelled");
        assert_eq!(v["error"]["details"]["task_id"], id.as_str());
    }
}
