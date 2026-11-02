//! Webhook 回调(Task 4.1,`docs/spec/04-http-api.md` "Webhook 回调规格" + `12-security.md` "Webhook 安全")。
//!
//! 任务进入终态(completed / failed / cancelled)且请求带了 `callback_url` 时,由运行器调用 [`deliver`]:
//!
//! - body:`{ "task_id", "event", "status", "task": <同 GET /tasks/{id} 的响应体> }`(字段顺序固定;
//!   `task` 由 [`crate::tasks::detail::task_detail_json`] 生成,含 `account` / `type` / `message` /
//!   `failed_at` / `result.files[].download_url` / `result.watermark_source`);
//! - 头:`Content-Type: application/json`、`X-Doubao-Relay-Timestamp: <Unix ms>`、
//!   `X-Doubao-Relay-Signature: sha256=<hex>`,签名 = HMAC-SHA256(key=bearer_token, data=timestamp + "\n" + body)
//!   (以 12-security 为准;04 写的 `hmac-sha256=` 前缀与无换行拼接是旧写法,计划已选定 12);
//! - 单次 30s 超时;失败按 1s / 10s / 60s 重试 3 次(共最多 4 次),每次尝试都写一行 `webhook_log`
//!   (`error_message` 以 11-error-codes 的 webhook 错误码开头:`webhook_timeout` / `webhook_http_error` /
//!   `webhook_connection_failed`);
//! - 全部失败 → 在任务上记 `callback_failed`:追加一条 `stage_history`(stage=`callback_failed`,
//!   `GET /tasks/{id}` 可见),失败/取消任务另把 `{ code: "webhook_max_retries_exceeded", ... }`
//!   合并进 `error.details.callback_failed`。
//!
//! 每次尝试重新取时间戳并重新签名(调用方按 ±5 分钟校验 timestamp,60s 退避后仍新鲜)。
//! 日志不含 bearer_token 与 prompt,URL 只记 host。

use std::time::Duration;

use hmac::{Hmac, Mac};
use serde::Serialize;
use serde_json::{json, Value};
use sha2::Sha256;

use crate::error::{RelayError, RelayResult};
use crate::storage::{Storage, WebhookLog};
use crate::tasks::detail::task_detail_json;
use crate::tasks::{Task, TaskStatus};

/// 签名头。
pub const HEADER_SIGNATURE: &str = "X-Doubao-Relay-Signature";
/// 时间戳头(Unix 毫秒)。
pub const HEADER_TIMESTAMP: &str = "X-Doubao-Relay-Timestamp";
/// 签名值前缀(12-security)。
pub const SIGNATURE_PREFIX: &str = "sha256=";
/// 单次回调超时(04 "超时:30 秒")。
pub const TIMEOUT: Duration = Duration::from_secs(30);
/// 重试间隔(04 "1s / 10s / 60s 三次重试")。
pub const RETRY_DELAYS: [Duration; 3] = [
    Duration::from_secs(1),
    Duration::from_secs(10),
    Duration::from_secs(60),
];
/// `webhook_log.response_body` 最多保留的字节数。
pub const RESPONSE_BODY_LIMIT: usize = 4096;
/// 最终失败时写进 stage_history 的阶段名。
pub const CALLBACK_FAILED_STAGE: &str = "callback_failed";

/// 重试耗尽后的汇总错误码(spec 11),写入 `callback_failed` 阶段文案与 `error.details`。
pub const MAX_RETRIES_EXCEEDED: &str = "webhook_max_retries_exceeded";

/// 回调时序参数(测试用小值;生产用 [`WebhookPolicy::default`])。
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct WebhookPolicy {
    /// 单次请求超时(连接 + 发送 + 读响应)。
    pub timeout: Duration,
    /// 首次失败后的各次重试间隔;长度即重试次数。
    pub retry_delays: Vec<Duration>,
}

impl Default for WebhookPolicy {
    fn default() -> Self {
        Self {
            timeout: TIMEOUT,
            retry_delays: RETRY_DELAYS.to_vec(),
        }
    }
}

/// 一次 [`deliver`] 的结果(失败已落库,调用方一般只用来打日志 / 测试断言)。
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DeliveryOutcome {
    /// 是否有一次拿到 2xx。
    pub delivered: bool,
    /// 实际尝试次数(含首次)。
    pub attempts: u32,
    /// 最后一次尝试的 HTTP 状态码(没拿到响应为 `None`)。
    pub last_http_status: Option<u16>,
    /// 最后一次失败的错误码(成功为 `None`)。
    pub last_error_code: Option<&'static str>,
}

/// 回调 body(字段顺序即序列化顺序)。
#[derive(Serialize)]
struct WebhookBody<'a> {
    task_id: &'a str,
    event: &'a str,
    status: TaskStatus,
    task: Value,
}

/// 单次尝试的结果。
enum Attempt {
    Ok { status: u16, body: String },
    Http { status: u16, body: String },
    Failed { code: &'static str, message: String },
}

/// 终态 → 回调事件名;非终态返回 `None`(不回调)。
pub fn event_for(status: TaskStatus) -> Option<&'static str> {
    match status {
        TaskStatus::Completed => Some("task.completed"),
        TaskStatus::Failed => Some("task.failed"),
        TaskStatus::Cancelled => Some("task.cancelled"),
        _ => None,
    }
}

fn hmac_sha256_hex(key: &[u8], data: &[u8]) -> String {
    // HMAC 接受任意长度 key,`new_from_slice` 不会失败。
    let mut mac = <Hmac<Sha256> as Mac>::new_from_slice(key).expect("HMAC 接受任意长度 key");
    mac.update(data);
    hex::encode(mac.finalize().into_bytes())
}

/// HMAC-SHA256(key=token, data=timestamp + "\n" + body) 的小写 hex。
pub fn sign(token: &str, timestamp: i64, body: &[u8]) -> String {
    let mut data = Vec::with_capacity(body.len() + 24);
    data.extend_from_slice(timestamp.to_string().as_bytes());
    data.push(b'\n');
    data.extend_from_slice(body);
    hmac_sha256_hex(token.as_bytes(), &data)
}

/// `X-Doubao-Relay-Signature` 头的值:`sha256=<hex>`。
pub fn signature_header(token: &str, timestamp: i64, body: &[u8]) -> String {
    format!("{SIGNATURE_PREFIX}{}", sign(token, timestamp, body))
}

/// 序列化回调 body。
pub fn build_body(task: &Task, event: &str) -> RelayResult<Vec<u8>> {
    serde_json::to_vec(&WebhookBody {
        task_id: &task.id,
        event,
        status: task.status,
        task: task_detail_json(task),
    })
    .map_err(|e| RelayError::internal(format!("序列化回调内容失败:{e}")))
}

/// 回调专用 HTTP 客户端:不跟随重定向(防止被 302 引到内网),UA `doubao-relay/<版本>`。
pub fn new_client() -> reqwest::Client {
    reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .user_agent(concat!("doubao-relay/", env!("CARGO_PKG_VERSION")))
        .build()
        .expect("reqwest client 构造失败")
}

fn now_ms() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as i64)
        .unwrap_or(0)
}

fn url_host(url: &str) -> String {
    url::Url::parse(url)
        .ok()
        .and_then(|u| u.host_str().map(str::to_string))
        .unwrap_or_default()
}

/// 按 UTF-8 字符边界截断到 `limit` 字节以内。
fn truncate_utf8(mut s: String, limit: usize) -> String {
    if s.len() > limit {
        let mut end = limit;
        while !s.is_char_boundary(end) {
            end -= 1;
        }
        s.truncate(end);
    }
    s
}

/// 读响应 body,最多 [`RESPONSE_BODY_LIMIT`] 字节(超出部分不再读)。
async fn read_body_limited(resp: &mut reqwest::Response) -> Result<String, reqwest::Error> {
    let mut buf: Vec<u8> = Vec::new();
    while buf.len() <= RESPONSE_BODY_LIMIT {
        match resp.chunk().await? {
            Some(chunk) => buf.extend_from_slice(&chunk),
            None => break,
        }
    }
    Ok(truncate_utf8(
        String::from_utf8_lossy(&buf).into_owned(),
        RESPONSE_BODY_LIMIT,
    ))
}

fn classify(e: &reqwest::Error) -> (&'static str, String) {
    if e.is_timeout() {
        ("webhook_timeout", "回调请求超时".to_string())
    } else {
        ("webhook_connection_failed", format!("回调连接失败:{e}"))
    }
}

async fn attempt_once(
    client: &reqwest::Client,
    url: &str,
    token: &str,
    body: &[u8],
    timestamp: i64,
    timeout: Duration,
) -> Attempt {
    let send = async {
        let mut resp = client
            .post(url)
            .header(reqwest::header::CONTENT_TYPE, "application/json")
            .header(HEADER_TIMESTAMP, timestamp.to_string())
            .header(HEADER_SIGNATURE, signature_header(token, timestamp, body))
            .body(body.to_vec())
            .send()
            .await?;
        let status = resp.status();
        let text = read_body_limited(&mut resp).await?;
        Ok::<_, reqwest::Error>((status, text))
    };
    match tokio::time::timeout(timeout, send).await {
        Err(_) => Attempt::Failed {
            code: "webhook_timeout",
            message: format!("回调请求超时(> {}s)", timeout.as_secs_f64()),
        },
        Ok(Err(e)) => {
            let (code, message) = classify(&e);
            Attempt::Failed { code, message }
        }
        Ok(Ok((status, text))) if status.is_success() => Attempt::Ok {
            status: status.as_u16(),
            body: text,
        },
        Ok(Ok((status, text))) => Attempt::Http {
            status: status.as_u16(),
            body: text,
        },
    }
}

/// 生产入口:默认 30s 超时、1s/10s/60s 重试。见模块文档。
pub async fn deliver(
    client: &reqwest::Client,
    task: &Task,
    event: &str,
    url: &str,
    token: &str,
    storage: &Storage,
) -> DeliveryOutcome {
    deliver_with(
        client,
        task,
        event,
        url,
        token,
        storage,
        &WebhookPolicy::default(),
    )
    .await
}

/// 同 [`deliver`],时序参数可配(测试缩短重试间隔)。
pub async fn deliver_with(
    client: &reqwest::Client,
    task: &Task,
    event: &str,
    url: &str,
    token: &str,
    storage: &Storage,
    policy: &WebhookPolicy,
) -> DeliveryOutcome {
    let host = url_host(url);
    let body = match build_body(task, event) {
        Ok(b) => b,
        Err(e) => {
            tracing::error!(task_id = %task.id, event, error = %e.message, "回调内容序列化失败,放弃回调");
            return DeliveryOutcome {
                delivered: false,
                attempts: 0,
                last_http_status: None,
                last_error_code: Some("webhook_connection_failed"),
            };
        }
    };

    let max_attempts = policy.retry_delays.len() as u32 + 1;
    let mut last_http_status = None;
    let mut last_error_code = None;
    let mut last_error_message = String::new();

    for attempt in 1..=max_attempts {
        if attempt > 1 {
            tokio::time::sleep(policy.retry_delays[(attempt - 2) as usize]).await;
        }
        let timestamp = now_ms();
        let result = attempt_once(client, url, token, &body, timestamp, policy.timeout).await;
        let (http_status, response_body, error_message, ok) = match result {
            Attempt::Ok { status, body } => (Some(status), Some(body), None, true),
            Attempt::Http { status, body } => {
                last_error_code = Some("webhook_http_error");
                let msg = format!("webhook_http_error:回调返回非 2xx(HTTP {status})");
                (Some(status), Some(body), Some(msg), false)
            }
            Attempt::Failed { code, message } => {
                last_error_code = Some(code);
                (None, None, Some(format!("{code}:{message}")), false)
            }
        };
        last_http_status = http_status;
        if let Some(m) = &error_message {
            last_error_message = m.clone();
        }

        let log = WebhookLog {
            task_id: task.id.clone(),
            event: event.to_string(),
            url: url.to_string(),
            attempt,
            http_status,
            response_body,
            error_message,
            sent_at: timestamp,
        };
        if let Err(e) = storage.insert_webhook_log(&log).await {
            tracing::warn!(task_id = %task.id, attempt, error = %e.message, "写入 webhook_log 失败");
        }

        if ok {
            tracing::info!(
                task_id = %task.id,
                event,
                host = %host,
                attempt,
                http_status = ?http_status,
                "回调成功"
            );
            return DeliveryOutcome {
                delivered: true,
                attempts: attempt,
                last_http_status,
                last_error_code: None,
            };
        }
        tracing::warn!(
            task_id = %task.id,
            event,
            host = %host,
            attempt,
            max_attempts,
            http_status = ?http_status,
            error = %last_error_message,
            "回调失败"
        );
    }

    let details = json!({
        "code": MAX_RETRIES_EXCEEDED,
        "event": event,
        "url": url,
        "attempts": max_attempts,
        "last_http_status": last_http_status,
        "last_error": last_error_message,
    });
    record_callback_failed(storage, &task.id, max_attempts, details).await;
    tracing::warn!(task_id = %task.id, event, host = %host, attempts = max_attempts, "回调重试用尽,已在任务上记录 callback_failed");

    DeliveryOutcome {
        delivered: false,
        attempts: max_attempts,
        last_http_status,
        last_error_code,
    }
}

/// `callback_failed` 阶段的 message:固定带汇总码 `webhook_max_retries_exceeded`
/// (completed 任务没有 `error.details`,错误码只能从这里看到)。
pub fn callback_failed_message(attempts: u32) -> String {
    format!(
        "回调 {attempts} 次均失败,已放弃({MAX_RETRIES_EXCEEDED}),请检查 callback_url 服务是否可用"
    )
}

/// 在任务上记 `callback_failed`(见模块文档);存储失败只 warn。
async fn record_callback_failed(storage: &Storage, task_id: &str, attempts: u32, details: Value) {
    let now = now_ms();
    let message = callback_failed_message(attempts);
    if let Err(e) = storage
        .append_stage(task_id, CALLBACK_FAILED_STAGE, Some(&message), now)
        .await
    {
        tracing::warn!(task_id, error = %e.message, "记录 callback_failed 阶段失败");
    }
    let mut task = match storage.get_task(task_id).await {
        Ok(Some(t)) => t,
        Ok(None) => return,
        Err(e) => {
            tracing::warn!(task_id, error = %e.message, "读取任务失败,callback_failed 未写入 error.details");
            return;
        }
    };
    let Some(err) = task.error.as_mut() else {
        return;
    };
    let mut merged = match err.details.take() {
        Some(Value::Object(m)) => m,
        Some(other) => {
            let mut m = serde_json::Map::new();
            m.insert("original".into(), other);
            m
        }
        None => serde_json::Map::new(),
    };
    merged.insert(CALLBACK_FAILED_STAGE.into(), details);
    err.details = Some(Value::Object(merged));
    if let Err(e) = storage.update_task(&task).await {
        tracing::warn!(task_id, error = %e.message, "callback_failed 写入 error.details 失败");
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::storage::db::tests::{open_tmp, pending_task};
    use crate::tasks::TaskError;
    use axum::extract::State;
    use axum::http::{HeaderMap, StatusCode};
    use axum::routing::post;
    use axum::Router;
    use sqlx::Row;
    use std::sync::atomic::{AtomicU32, Ordering};
    use std::sync::{Arc, Mutex};

    const TOKEN: &str = "drk_test_token_0123456789abcdef";

    // ---- 签名 ----

    #[test]
    fn hmac_matches_known_vector() {
        // 经典向量:HMAC_SHA256("key", "The quick brown fox jumps over the lazy dog")
        assert_eq!(
            hmac_sha256_hex(b"key", b"The quick brown fox jumps over the lazy dog"),
            "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
        );
    }

    #[test]
    fn sign_matches_python_hmac_vector() {
        // python3: hmac.new(TOKEN, (ts + "\n" + body).encode(), hashlib.sha256).hexdigest()
        let body = br#"{"task_id":"01HQZ9ABCDEFGHJKMNPQRSTVWX","event":"task.completed","status":"completed"}"#;
        let expected = "bba590dd38852da26c979f389b3bbc698d64ac0e9f60dbfb532ddd935b3f0570";
        assert_eq!(sign(TOKEN, 1_697_433_700_000, body), expected);
        assert_eq!(
            signature_header(TOKEN, 1_697_433_700_000, body),
            format!("sha256={expected}")
        );
        // 时间戳或 body 变了签名就变
        assert_ne!(sign(TOKEN, 1_697_433_700_001, body), expected);
        assert_ne!(sign("other", 1_697_433_700_000, body), expected);
    }

    #[test]
    fn event_for_terminal_statuses_only() {
        assert_eq!(event_for(TaskStatus::Completed), Some("task.completed"));
        assert_eq!(event_for(TaskStatus::Failed), Some("task.failed"));
        assert_eq!(event_for(TaskStatus::Cancelled), Some("task.cancelled"));
        for s in [
            TaskStatus::Pending,
            TaskStatus::Preparing,
            TaskStatus::Generating,
            TaskStatus::WaitingUserConfirm,
        ] {
            assert_eq!(event_for(s), None);
        }
    }

    #[test]
    fn default_policy_is_30s_and_1_10_60() {
        let p = WebhookPolicy::default();
        assert_eq!(p.timeout, Duration::from_secs(30));
        assert_eq!(
            p.retry_delays,
            vec![
                Duration::from_secs(1),
                Duration::from_secs(10),
                Duration::from_secs(60)
            ]
        );
    }

    #[test]
    fn build_body_has_fixed_field_order_and_full_task() {
        let mut t = completed_task("01HQZ9ABCDEFGHJKMNPQRSTVWX");
        t.request.callback_url = Some("https://example.com/hook".into());
        t.watermark_source = Some(crate::tasks::WatermarkSource::ConversationOriginal);
        t.result_files = Some(vec![crate::tasks::ResultFile {
            index: 0,
            filename: "image_1.png".into(),
            relative_path: "results/01HQZ9ABCDEFGHJKMNPQRSTVWX/image_1.png".into(),
            mime: "image/png".into(),
            bytes: 10,
            width: Some(1),
            height: Some(2),
            duration_ms: None,
            source_url: None,
        }]);
        let body = build_body(&t, "task.completed").unwrap();
        let text = String::from_utf8(body.clone()).unwrap();
        assert!(text.starts_with(
            r#"{"task_id":"01HQZ9ABCDEFGHJKMNPQRSTVWX","event":"task.completed","status":"completed","task":{"#
        ));
        let v: Value = serde_json::from_slice(&body).unwrap();
        assert_eq!(v["task"]["task_id"], "01HQZ9ABCDEFGHJKMNPQRSTVWX");
        assert_eq!(v["task"]["status"], "completed");
        assert_eq!(v["task"]["request"]["type"], "image");
        assert_eq!(v.as_object().unwrap().len(), 4);
        // `task` 与 GET /tasks/{id} 的响应体完全一致(spec 04)
        assert_eq!(v["task"], task_detail_json(&t));
        assert_eq!(v["task"]["account"], "张三");
        assert_eq!(v["task"]["type"], "image");
        assert_eq!(
            v["task"]["result"]["files"][0]["download_url"],
            "/api/v1/tasks/01HQZ9ABCDEFGHJKMNPQRSTVWX/download/0"
        );
        assert_eq!(
            v["task"]["result"]["watermark_source"],
            "conversation_original"
        );
    }

    #[test]
    fn build_body_failed_task_has_failed_at_and_error() {
        let t = failed_task("01HQZ9ABCDEFGHJKMNPQRSTVWX");
        let v: Value = serde_json::from_slice(&build_body(&t, "task.failed").unwrap()).unwrap();
        assert_eq!(v["task"], task_detail_json(&t));
        assert_eq!(v["task"]["failed_at"], 1_697_433_700_000_i64);
        assert_eq!(v["task"]["error"]["code"], "generation_refused");
        assert!(v["task"].get("result").is_none());
    }

    #[test]
    fn truncate_utf8_respects_char_boundary() {
        assert_eq!(truncate_utf8("abc".into(), 10), "abc");
        assert_eq!(truncate_utf8("豆包".into(), 4), "豆"); // 每字 3 字节
        assert_eq!(truncate_utf8("豆包".into(), 6), "豆包");
    }

    // ---- 投递(axum mock) ----

    fn completed_task(id: &str) -> Task {
        let mut t = pending_task(id, "张三", 1_697_433_600_000);
        t.status = TaskStatus::Completed;
        t.completed_at = Some(1_697_433_700_000);
        t
    }

    fn failed_task(id: &str) -> Task {
        let mut t = pending_task(id, "张三", 1_697_433_600_000);
        t.status = TaskStatus::Failed;
        t.completed_at = Some(1_697_433_700_000);
        t.error = Some(TaskError {
            code: "generation_refused".into(),
            message: "豆包拒绝生成,请调整提示词后重试".into(),
            details: Some(json!({ "reason": "policy" })),
        });
        t
    }

    fn fast_policy() -> WebhookPolicy {
        WebhookPolicy {
            timeout: Duration::from_secs(5),
            retry_delays: vec![
                Duration::from_millis(10),
                Duration::from_millis(20),
                Duration::from_millis(30),
            ],
        }
    }

    /// mock 收到的一次请求:(头, body)。
    type Received = (HeaderMap, Vec<u8>);

    #[derive(Clone, Default)]
    struct Mock {
        hits: Arc<AtomicU32>,
        /// 第 n 次(1-based)请求返回 `statuses[n-1]`,超出部分用最后一个。
        statuses: Arc<Vec<u16>>,
        /// 每次请求先睡多久再回。
        delay: Duration,
        /// 响应 body。
        reply: Arc<String>,
        received: Arc<Mutex<Vec<Received>>>,
    }

    async fn hook(
        State(m): State<Mock>,
        headers: HeaderMap,
        body: axum::body::Bytes,
    ) -> (StatusCode, String) {
        let n = m.hits.fetch_add(1, Ordering::SeqCst) as usize;
        m.received.lock().unwrap().push((headers, body.to_vec()));
        if !m.delay.is_zero() {
            tokio::time::sleep(m.delay).await;
        }
        let code = m
            .statuses
            .get(n)
            .or(m.statuses.last())
            .copied()
            .unwrap_or(200);
        (StatusCode::from_u16(code).unwrap(), (*m.reply).clone())
    }

    async fn spawn_mock(m: Mock) -> String {
        let app = Router::new().route("/hook", post(hook)).with_state(m);
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        format!("http://{addr}/hook")
    }

    fn mock_with(statuses: &[u16]) -> Mock {
        Mock {
            statuses: Arc::new(statuses.to_vec()),
            reply: Arc::new("ok".into()),
            ..Default::default()
        }
    }

    #[derive(Debug)]
    struct LogRow {
        attempt: i64,
        event: String,
        url: String,
        http_status: Option<i64>,
        response_body: Option<String>,
        error_message: Option<String>,
    }

    async fn logs(storage: &Storage, task_id: &str) -> Vec<LogRow> {
        sqlx::query(
            "SELECT attempt, event, url, http_status, response_body, error_message FROM webhook_log WHERE task_id = ? ORDER BY id",
        )
        .bind(task_id)
        .fetch_all(&storage.pool)
        .await
        .unwrap()
        .into_iter()
        .map(|r| LogRow {
            attempt: r.get("attempt"),
            event: r.get("event"),
            url: r.get("url"),
            http_status: r.get("http_status"),
            response_body: r.get("response_body"),
            error_message: r.get("error_message"),
        })
        .collect()
    }

    #[tokio::test]
    async fn retries_until_200_and_logs_each_attempt_with_valid_signature() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA1");
        storage.insert_task(&task).await.unwrap();
        let mock = mock_with(&[500, 500, 200]);
        let url = spawn_mock(mock.clone()).await;

        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &fast_policy(),
        )
        .await;
        assert_eq!(
            out,
            DeliveryOutcome {
                delivered: true,
                attempts: 3,
                last_http_status: Some(200),
                last_error_code: None
            }
        );

        let rows = logs(&storage, &task.id).await;
        assert_eq!(rows.len(), 3, "{rows:?}");
        assert_eq!(
            rows.iter().map(|r| r.attempt).collect::<Vec<_>>(),
            vec![1, 2, 3]
        );
        assert_eq!(
            rows.iter().map(|r| r.http_status).collect::<Vec<_>>(),
            vec![Some(500), Some(500), Some(200)]
        );
        assert!(rows
            .iter()
            .all(|r| r.event == "task.completed" && r.url == url));
        assert!(rows[0]
            .error_message
            .as_deref()
            .unwrap()
            .starts_with("webhook_http_error"));
        assert_eq!(rows[2].error_message, None);
        assert_eq!(rows[2].response_body.as_deref(), Some("ok"));

        // 服务端收到的每次请求:签名可用 token 校验、body 一致
        let received = mock.received.lock().unwrap().clone();
        assert_eq!(received.len(), 3);
        for (headers, body) in &received {
            assert_eq!(headers["content-type"], "application/json");
            let ts: i64 = headers[HEADER_TIMESTAMP].to_str().unwrap().parse().unwrap();
            let sig = headers[HEADER_SIGNATURE].to_str().unwrap();
            assert_eq!(sig, format!("sha256={}", sign(TOKEN, ts, body)));
            let v: Value = serde_json::from_slice(body).unwrap();
            assert_eq!(v["task_id"], task.id);
            assert_eq!(v["event"], "task.completed");
            assert_eq!(v["status"], "completed");
            assert_eq!(v["task"]["task_id"], task.id);
            assert!(!headers.contains_key("authorization"), "不带 bearer token");
        }

        // 成功不记 callback_failed
        let got = storage.get_task(&task.id).await.unwrap().unwrap();
        assert!(got
            .stage_history
            .iter()
            .all(|s| s.stage != CALLBACK_FAILED_STAGE));
    }

    #[tokio::test]
    async fn all_attempts_fail_logs_four_rows_and_records_callback_failed() {
        let (_dir, storage) = open_tmp().await;
        let task = failed_task("01HQZ9AAAAAAAAAAAAAAAAAAA2");
        storage.insert_task(&task).await.unwrap();
        let mock = mock_with(&[500]);
        let url = spawn_mock(mock.clone()).await;

        let out = deliver_with(
            &new_client(),
            &task,
            "task.failed",
            &url,
            TOKEN,
            &storage,
            &fast_policy(),
        )
        .await;
        assert!(!out.delivered);
        assert_eq!(out.attempts, 4);
        assert_eq!(out.last_http_status, Some(500));
        assert_eq!(out.last_error_code, Some("webhook_http_error"));
        assert_eq!(mock.hits.load(Ordering::SeqCst), 4);

        let rows = logs(&storage, &task.id).await;
        assert_eq!(
            rows.iter().map(|r| r.attempt).collect::<Vec<_>>(),
            vec![1, 2, 3, 4]
        );
        assert!(rows.iter().all(|r| r.http_status == Some(500)));

        let got = storage.get_task(&task.id).await.unwrap().unwrap();
        let entry = got
            .stage_history
            .iter()
            .find(|s| s.stage == CALLBACK_FAILED_STAGE)
            .expect("stage_history 有 callback_failed");
        assert!(entry.message.as_deref().unwrap().contains("4 次"));
        assert!(entry
            .message
            .as_deref()
            .unwrap()
            .contains(MAX_RETRIES_EXCEEDED));
        let details = got.error.unwrap().details.unwrap();
        assert_eq!(details["reason"], "policy", "原 details 保留");
        assert_eq!(
            details["callback_failed"]["code"],
            "webhook_max_retries_exceeded"
        );
        assert_eq!(details["callback_failed"]["attempts"], 4);
        assert_eq!(details["callback_failed"]["last_http_status"], 500);
        assert_eq!(details["callback_failed"]["event"], "task.failed");
        // 任务状态不受影响
        assert_eq!(got.status, TaskStatus::Failed);
    }

    #[tokio::test]
    async fn completed_task_failure_records_stage_only() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA3");
        storage.insert_task(&task).await.unwrap();
        let url = spawn_mock(mock_with(&[503])).await;
        let policy = WebhookPolicy {
            retry_delays: vec![Duration::from_millis(5)],
            ..fast_policy()
        };
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &policy,
        )
        .await;
        assert_eq!(out.attempts, 2, "重试次数 = retry_delays 长度");
        let got = storage.get_task(&task.id).await.unwrap().unwrap();
        assert_eq!(got.status, TaskStatus::Completed);
        assert_eq!(got.error, None);
        let entry = got
            .stage_history
            .iter()
            .find(|s| s.stage == CALLBACK_FAILED_STAGE)
            .expect("stage_history 有 callback_failed");
        let msg = entry.message.as_deref().unwrap();
        assert!(msg.contains(MAX_RETRIES_EXCEEDED), "{msg}");
        assert!(msg.contains("2 次"), "{msg}");
    }

    #[test]
    fn callback_failed_message_carries_max_retries_code() {
        let msg = callback_failed_message(4);
        assert!(msg.contains("webhook_max_retries_exceeded"), "{msg}");
        assert!(msg.contains("4 次"), "{msg}");
    }

    #[tokio::test]
    async fn redirect_with_location_is_not_followed() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA8");
        storage.insert_task(&task).await.unwrap();

        let target_hits = Arc::new(AtomicU32::new(0));
        let hook_hits = Arc::new(AtomicU32::new(0));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        let (th, hh) = (Arc::clone(&target_hits), Arc::clone(&hook_hits));
        let app = Router::new()
            .route(
                "/hook",
                post(move || {
                    let hh = Arc::clone(&hh);
                    async move {
                        hh.fetch_add(1, Ordering::SeqCst);
                        (
                            StatusCode::FOUND,
                            [(axum::http::header::LOCATION, "/elsewhere")],
                            "",
                        )
                    }
                }),
            )
            .route(
                "/elsewhere",
                axum::routing::any(move || {
                    let th = Arc::clone(&th);
                    async move {
                        th.fetch_add(1, Ordering::SeqCst);
                        "ok"
                    }
                }),
            );
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        let url = format!("http://{addr}/hook");

        let policy = WebhookPolicy {
            retry_delays: vec![Duration::from_millis(5)],
            ..fast_policy()
        };
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &policy,
        )
        .await;
        assert!(!out.delivered);
        assert_eq!(out.attempts, 2);
        assert_eq!(out.last_http_status, Some(302));
        assert_eq!(out.last_error_code, Some("webhook_http_error"));
        assert_eq!(hook_hits.load(Ordering::SeqCst), 2);
        assert_eq!(target_hits.load(Ordering::SeqCst), 0, "不得跟随 Location");

        // 对照:默认客户端会跟随 302,证明上面的 0 来自 new_client() 而不是 mock
        let resp = reqwest::Client::new().post(&url).send().await.unwrap();
        assert_eq!(resp.status(), 200);
        assert_eq!(target_hits.load(Ordering::SeqCst), 1);
    }

    #[tokio::test]
    async fn connection_refused_is_webhook_connection_failed() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA4");
        storage.insert_task(&task).await.unwrap();
        // 绑一个端口再释放,得到一个(几乎肯定)没人监听的端口
        let port = {
            let l = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
            l.local_addr().unwrap().port()
        };
        let url = format!("http://127.0.0.1:{port}/hook");
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &fast_policy(),
        )
        .await;
        assert!(!out.delivered);
        assert_eq!(out.last_http_status, None);
        assert_eq!(out.last_error_code, Some("webhook_connection_failed"));
        let rows = logs(&storage, &task.id).await;
        assert_eq!(rows.len(), 4);
        for r in &rows {
            assert_eq!(r.http_status, None);
            assert_eq!(r.response_body, None);
            assert!(r
                .error_message
                .as_deref()
                .unwrap()
                .starts_with("webhook_connection_failed"));
        }
    }

    #[tokio::test]
    async fn slow_server_is_webhook_timeout() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA5");
        storage.insert_task(&task).await.unwrap();
        let mock = Mock {
            delay: Duration::from_secs(3),
            ..mock_with(&[200])
        };
        let url = spawn_mock(mock).await;
        let policy = WebhookPolicy {
            timeout: Duration::from_millis(200),
            retry_delays: vec![Duration::from_millis(5)],
        };
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &policy,
        )
        .await;
        assert!(!out.delivered);
        assert_eq!(out.attempts, 2);
        assert_eq!(out.last_error_code, Some("webhook_timeout"));
        let rows = logs(&storage, &task.id).await;
        assert_eq!(rows.len(), 2);
        assert!(rows.iter().all(|r| r
            .error_message
            .as_deref()
            .unwrap()
            .starts_with("webhook_timeout")));
    }

    #[tokio::test]
    async fn redirect_is_not_followed_and_counts_as_http_error() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA6");
        storage.insert_task(&task).await.unwrap();
        let url = spawn_mock(mock_with(&[302])).await;
        let policy = WebhookPolicy {
            retry_delays: vec![],
            ..fast_policy()
        };
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &policy,
        )
        .await;
        assert_eq!(out.attempts, 1);
        assert_eq!(out.last_http_status, Some(302));
        assert_eq!(out.last_error_code, Some("webhook_http_error"));
    }

    #[tokio::test]
    async fn long_response_body_is_truncated_in_log() {
        let (_dir, storage) = open_tmp().await;
        let task = completed_task("01HQZ9AAAAAAAAAAAAAAAAAAA7");
        storage.insert_task(&task).await.unwrap();
        let mock = Mock {
            reply: Arc::new("x".repeat(100_000)),
            ..mock_with(&[200])
        };
        let url = spawn_mock(mock).await;
        let out = deliver_with(
            &new_client(),
            &task,
            "task.completed",
            &url,
            TOKEN,
            &storage,
            &fast_policy(),
        )
        .await;
        assert!(out.delivered);
        let rows = logs(&storage, &task.id).await;
        assert_eq!(
            rows[0].response_body.as_ref().unwrap().len(),
            RESPONSE_BODY_LIMIT
        );
    }
}
