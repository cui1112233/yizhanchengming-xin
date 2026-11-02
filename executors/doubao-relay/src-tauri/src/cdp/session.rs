//! 单个 CDP target 的 WebSocket 会话。
//!
//! - 写:所有命令经 `mpsc` 串行写入 socket(写循环一个 task);
//! - 读:读循环一个 task,按 `id` 匹配响应、无 `id` 的消息作为事件广播;
//! - 断线:读循环结束时置 `closed`、清空 pending(等待中的调用全部得到 `CdpError::Closed`);
//! - 超时:`send` 15s,`evaluate` 30s(`docs/spec/09-cdp-contract.md`);
//! - 调试:`debug_log` 为 Some 时每条收发各推一条 `CdpLogEntry`。

use std::collections::HashMap;
use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::{Arc, Mutex};
use std::time::Duration;

use futures_util::{SinkExt, StreamExt};
use serde::de::DeserializeOwned;
use serde_json::{json, Value};
use tokio::sync::{broadcast, mpsc, oneshot};
use tokio::task::JoinHandle;
use tokio_tungstenite::tungstenite::{protocol::WebSocketConfig, Message};

use super::{now_ms, CdpError, CdpLogEntry};

/// 普通命令默认超时。
pub const DEFAULT_TIMEOUT: Duration = Duration::from_secs(15);
/// `Runtime.evaluate` 默认超时(JS 可能慢)。
pub const EVALUATE_TIMEOUT: Duration = Duration::from_secs(30);
/// WebSocket 握手超时。
const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);
/// 事件广播容量(慢订阅者会 Lagged)。
const EVENT_CAPACITY: usize = 4096;
/// 写队列容量。
const OUTGOING_CAPACITY: usize = 256;
/// 豆包上传参考图后会通过 CDP 回传超过默认 16 MiB 的单帧消息。
// A 50 MiB reference can expand past 64 MiB when Doubao reports the upload over CDP.
const MAX_CDP_MESSAGE_SIZE: usize = 80 * 1024 * 1024;

/// 一条 CDP 事件(无 `id` 的消息)。
#[derive(Clone, Debug)]
pub struct CdpEvent {
    pub method: String,
    pub params: Value,
}

enum Outgoing {
    Text(String),
    Close,
}

struct Pending {
    method: String,
    tx: oneshot::Sender<Value>,
}

type PendingMap = Arc<Mutex<HashMap<u64, Pending>>>;

/// 一个 target 的 CDP 会话。可被多个任务并发使用(`send` / `evaluate` 各自独立 id)。
pub struct CdpSession {
    tx: mpsc::Sender<Outgoing>,
    pending: PendingMap,
    events: broadcast::Sender<CdpEvent>,
    next_id: AtomicU64,
    closed: Arc<AtomicBool>,
    pub target_id: String,
    debug_log: Option<broadcast::Sender<CdpLogEntry>>,
    reader: Mutex<Option<JoinHandle<()>>>,
}

impl std::fmt::Debug for CdpSession {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("CdpSession")
            .field("target_id", &self.target_id)
            .field("closed", &self.is_closed())
            .finish()
    }
}

impl CdpSession {
    /// 连接 `ws_url`(5s 握手超时)。不发任何命令;`Network.enable` 等由上层决定。
    pub async fn connect(ws_url: &str, target_id: &str) -> Result<Arc<Self>, CdpError> {
        Self::connect_with_debug(ws_url, target_id, None).await
    }

    /// 同 `connect`,并把每条收发消息推到 `debug_log`(`cdp_debug_log=true` 时由上层传入)。
    pub async fn connect_with_debug(
        ws_url: &str,
        target_id: &str,
        debug_log: Option<broadcast::Sender<CdpLogEntry>>,
    ) -> Result<Arc<Self>, CdpError> {
        let websocket_config = WebSocketConfig::default()
            .max_message_size(Some(MAX_CDP_MESSAGE_SIZE))
            .max_frame_size(Some(MAX_CDP_MESSAGE_SIZE));
        let connect =
            tokio_tungstenite::connect_async_with_config(ws_url, Some(websocket_config), false);
        let (ws, _resp) = tokio::time::timeout(CONNECT_TIMEOUT, connect)
            .await
            .map_err(|_| CdpError::Connect(format!("{ws_url} 握手超时")))?
            .map_err(|e| CdpError::Connect(format!("{ws_url}: {e}")))?;
        let (mut sink, mut stream) = ws.split();

        let (tx, mut rx) = mpsc::channel::<Outgoing>(OUTGOING_CAPACITY);
        let (events, _) = broadcast::channel(EVENT_CAPACITY);
        let pending: PendingMap = Arc::new(Mutex::new(HashMap::new()));
        let closed = Arc::new(AtomicBool::new(false));

        // 写循环:串行写 socket;发送方全部 drop 或收到 Close 时发 Close 帧并退出。
        tokio::spawn(async move {
            while let Some(out) = rx.recv().await {
                match out {
                    Outgoing::Text(text) => {
                        if sink.send(Message::text(text)).await.is_err() {
                            break;
                        }
                    }
                    Outgoing::Close => {
                        let _ = sink.send(Message::Close(None)).await;
                        break;
                    }
                }
            }
            let _ = sink.close().await;
        });

        // 读循环:分发响应/事件;结束时标记 closed、清空 pending。
        let reader = {
            let pending = Arc::clone(&pending);
            let closed = Arc::clone(&closed);
            let events = events.clone();
            let debug_log = debug_log.clone();
            let target_id = target_id.to_string();
            tokio::spawn(async move {
                while let Some(msg) = stream.next().await {
                    match msg {
                        Ok(Message::Text(text)) => {
                            handle_incoming(text.as_str(), &pending, &events, &debug_log)
                        }
                        Ok(Message::Binary(bytes)) => {
                            if let Ok(text) = std::str::from_utf8(&bytes) {
                                handle_incoming(text, &pending, &events, &debug_log);
                            }
                        }
                        Ok(Message::Close(frame)) => {
                            tracing::debug!(target: "cdp", target_id = %target_id, frame = ?frame, "CDP 连接被对端关闭");
                            break;
                        }
                        Ok(_) => {}
                        Err(e) => {
                            tracing::debug!(target: "cdp", target_id = %target_id, error = %e, "CDP 读取失败");
                            break;
                        }
                    }
                }
                closed.store(true, Ordering::SeqCst);
                drain_pending(&pending);
            })
        };

        Ok(Arc::new(Self {
            tx,
            pending,
            events,
            next_id: AtomicU64::new(0),
            closed,
            target_id: target_id.to_string(),
            debug_log,
            reader: Mutex::new(Some(reader)),
        }))
    }

    /// 发送命令,默认 15s 超时。
    pub async fn send(&self, method: &str, params: Value) -> Result<Value, CdpError> {
        self.send_timeout(method, params, DEFAULT_TIMEOUT).await
    }

    /// 发送命令并等待响应。返回 `result` 对象;`error` → `Protocol`;超时 → `Timeout(method)`;
    /// 连接断开 → `Closed`。
    pub async fn send_timeout(
        &self,
        method: &str,
        params: Value,
        timeout: Duration,
    ) -> Result<Value, CdpError> {
        if self.is_closed() {
            return Err(CdpError::Closed);
        }
        let id = self.next_id.fetch_add(1, Ordering::SeqCst) + 1;
        let (otx, orx) = oneshot::channel();
        self.pending
            .lock()
            .unwrap_or_else(|p| p.into_inner())
            .insert(
                id,
                Pending {
                    method: method.to_string(),
                    tx: otx,
                },
            );
        // 无论正常返回、超时还是 future 被调用方丢弃(外层 timeout / select!),都把 pending 条目删掉
        // (Task 6.1:长期存活的 session 不能因挂起的求值累积条目)。
        let _cleanup = PendingGuard {
            pending: &self.pending,
            id,
        };
        // 与读循环的"置 closed → 清 pending"竞争:插入后再查一次,避免永远等待。
        if self.is_closed() {
            return Err(CdpError::Closed);
        }
        if let Some(log) = debug_sink(&self.debug_log) {
            let _ = log.send(CdpLogEntry {
                ts: now_ms(),
                direction: "send".into(),
                method: Some(method.to_string()),
                params: Some(params.clone()),
                result: None,
                event: None,
            });
        }
        let text = json!({ "id": id, "method": method, "params": params }).to_string();
        if self.tx.send(Outgoing::Text(text)).await.is_err() {
            return Err(CdpError::Closed);
        }
        let msg = match tokio::time::timeout(timeout, orx).await {
            Err(_) => {
                return Err(CdpError::Timeout(method.to_string()));
            }
            Ok(Err(_)) => return Err(CdpError::Closed),
            Ok(Ok(msg)) => msg,
        };
        if let Some(err) = msg.get("error") {
            return Err(CdpError::Protocol {
                code: err.get("code").and_then(Value::as_i64).unwrap_or(0),
                message: err
                    .get("message")
                    .and_then(Value::as_str)
                    .unwrap_or("")
                    .to_string(),
            });
        }
        Ok(msg.get("result").cloned().unwrap_or_else(|| json!({})))
    }

    /// `Runtime.evaluate`(`returnByValue` + `awaitPromise`,30s)。返回 `result.value`;
    /// 页面抛异常 → `JsException(消息)`(从 `description` 第一行去掉 `Error: ` 前缀)。
    pub async fn evaluate(&self, expression: &str) -> Result<Value, CdpError> {
        let result = self
            .send_timeout(
                "Runtime.evaluate",
                json!({
                    "expression": expression,
                    "returnByValue": true,
                    "awaitPromise": true,
                }),
                EVALUATE_TIMEOUT,
            )
            .await?;
        if let Some(details) = result.get("exceptionDetails") {
            return Err(CdpError::JsException(extract_js_exception(details)));
        }
        Ok(result
            .get("result")
            .and_then(|r| r.get("value"))
            .cloned()
            .unwrap_or(Value::Null))
    }

    /// `evaluate` 后反序列化为 `T`。
    pub async fn evaluate_as<T: DeserializeOwned>(&self, expression: &str) -> Result<T, CdpError> {
        let value = self.evaluate(expression).await?;
        serde_json::from_value(value).map_err(|e| CdpError::Deserialize(e.to_string()))
    }

    /// 订阅 CDP 事件(无 `id` 的消息)。
    pub fn subscribe(&self) -> broadcast::Receiver<CdpEvent> {
        self.events.subscribe()
    }

    pub fn is_closed(&self) -> bool {
        self.closed.load(Ordering::SeqCst)
    }

    /// 主动关闭:置 closed、发 Close 帧、等待读循环退出(最多 1s,否则 abort)、清空 pending。
    pub async fn close(&self) {
        let already = self.closed.swap(true, Ordering::SeqCst);
        if !already {
            let _ = tokio::time::timeout(Duration::from_millis(500), self.tx.send(Outgoing::Close))
                .await;
        }
        let reader = self.reader.lock().unwrap().take();
        if let Some(mut handle) = reader {
            if tokio::time::timeout(Duration::from_secs(1), &mut handle)
                .await
                .is_err()
            {
                handle.abort();
            }
        }
        drain_pending(&self.pending);
    }
}

impl Drop for CdpSession {
    fn drop(&mut self) {
        self.closed.store(true, Ordering::SeqCst);
        if let Ok(mut guard) = self.reader.lock() {
            if let Some(handle) = guard.take() {
                handle.abort();
            }
        }
        drain_pending(&self.pending);
    }
}

/// `send_timeout` 的 pending 条目守卫:drop 时删掉自己的 id(已被读循环取走则是空操作)。
struct PendingGuard<'a> {
    pending: &'a PendingMap,
    id: u64,
}

impl Drop for PendingGuard<'_> {
    fn drop(&mut self) {
        if let Ok(mut map) = self.pending.lock() {
            map.remove(&self.id);
        }
    }
}

fn drain_pending(pending: &PendingMap) {
    if let Ok(mut map) = pending.lock() {
        // drop 掉所有 oneshot Sender,等待方收到 RecvError → Closed。
        map.clear();
    }
}

/// 调试日志只在有人订阅时才构造(`CdpLogBuffer` 关闭状态下没有订阅者,`params`/`result` 不克隆)。
fn debug_sink(
    debug_log: &Option<broadcast::Sender<CdpLogEntry>>,
) -> Option<&broadcast::Sender<CdpLogEntry>> {
    debug_log.as_ref().filter(|log| log.receiver_count() > 0)
}

fn handle_incoming(
    text: &str,
    pending: &PendingMap,
    events: &broadcast::Sender<CdpEvent>,
    debug_log: &Option<broadcast::Sender<CdpLogEntry>>,
) {
    let value: Value = match serde_json::from_str(text) {
        Ok(v) => v,
        Err(e) => {
            tracing::warn!(target: "cdp", error = %e, len = text.len(), "CDP 消息不是合法 JSON,已忽略");
            return;
        }
    };
    if let Some(id) = value.get("id").and_then(Value::as_u64) {
        let entry = pending
            .lock()
            .unwrap_or_else(|p| p.into_inner())
            .remove(&id);
        if let Some(log) = debug_sink(debug_log) {
            let _ = log.send(CdpLogEntry {
                ts: now_ms(),
                direction: "recv".into(),
                method: entry.as_ref().map(|p| p.method.clone()),
                params: None,
                result: Some(
                    value
                        .get("result")
                        .or_else(|| value.get("error"))
                        .cloned()
                        .unwrap_or(Value::Null),
                ),
                event: None,
            });
        }
        match entry {
            Some(p) => {
                let _ = p.tx.send(value);
            }
            None => {
                tracing::debug!(target: "cdp", id, "收到未知 id 的 CDP 响应(可能已超时),已忽略");
            }
        }
    } else if let Some(method) = value.get("method").and_then(Value::as_str) {
        let params = value.get("params").cloned().unwrap_or(Value::Null);
        if let Some(log) = debug_sink(debug_log) {
            let _ = log.send(CdpLogEntry {
                ts: now_ms(),
                direction: "recv".into(),
                method: Some(method.to_string()),
                params: None,
                result: None,
                event: Some(params.clone()),
            });
        }
        let _ = events.send(CdpEvent {
            method: method.to_string(),
            params,
        });
    }
}

/// 从 `exceptionDetails` 提取人类可读的异常消息:
/// 取 `exception.description`(否则 `exception.value` / `text`)第一行,去掉 `XxxError: ` 前缀。
pub fn extract_js_exception(details: &Value) -> String {
    let exception = details.get("exception");
    let raw = exception
        .and_then(|e| e.get("description"))
        .and_then(Value::as_str)
        .map(str::to_string)
        .or_else(|| {
            exception.and_then(|e| e.get("value")).map(|v| match v {
                Value::String(s) => s.clone(),
                other => other.to_string(),
            })
        })
        .or_else(|| {
            details
                .get("text")
                .and_then(Value::as_str)
                .map(str::to_string)
        })
        .unwrap_or_default();
    let first_line = raw.lines().next().unwrap_or("").trim();
    let message = match first_line.split_once(": ") {
        Some((prefix, rest))
            if prefix.ends_with("Error") && prefix.chars().all(|c| c.is_ascii_alphanumeric()) =>
        {
            rest.trim()
        }
        _ => first_line,
    };
    if message.is_empty() {
        "未知脚本异常".to_string()
    } else {
        message.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde::Deserialize;
    use tokio::net::TcpListener;
    use tokio_tungstenite::accept_async;

    /// 本地 mock CDP server:按 method 分支响应,见计划 Task 1.4 Step 1。
    async fn spawn_mock() -> String {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            loop {
                let Ok((stream, _)) = listener.accept().await else {
                    break;
                };
                tokio::spawn(async move {
                    let Ok(mut ws) = accept_async(stream).await else {
                        return;
                    };
                    while let Some(Ok(msg)) = ws.next().await {
                        let Message::Text(text) = msg else {
                            continue;
                        };
                        let req: Value = serde_json::from_str(text.as_str()).unwrap();
                        let id = req["id"].clone();
                        let method = req["method"].as_str().unwrap_or("");
                        let params = req["params"].clone();
                        let reply = match method {
                            "Echo.ping" => json!({ "id": id, "result": { "pong": params } }),
                            "Echo.event" => {
                                let ev = json!({ "method": "Test.fired", "params": { "n": 1 } });
                                ws.send(Message::text(ev.to_string())).await.unwrap();
                                json!({ "id": id, "result": {} })
                            }
                            "Echo.error" => json!({
                                "id": id,
                                "error": { "code": -32601, "message": "no such method" }
                            }),
                            "Echo.hang" => continue,
                            "Echo.close" => {
                                let _ = ws.close(None).await;
                                return;
                            }
                            "Runtime.evaluate" => {
                                let expr = params["expression"].as_str().unwrap_or("");
                                if expr.contains("throw") {
                                    json!({ "id": id, "result": { "exceptionDetails": {
                                        "text": "Uncaught",
                                        "exception": { "description": "Error: 中文错误\n    at <anonymous>:1:7" }
                                    }}})
                                } else {
                                    json!({ "id": id, "result": { "result": {
                                        "type": "object", "value": { "ok": true }
                                    }}})
                                }
                            }
                            _ => json!({
                                "id": id,
                                "error": { "code": -32601, "message": format!("unknown {method}") }
                            }),
                        };
                        if ws.send(Message::text(reply.to_string())).await.is_err() {
                            return;
                        }
                    }
                });
            }
        });
        format!("ws://127.0.0.1:{port}/devtools/page/MOCK")
    }

    async fn connect_mock() -> Arc<CdpSession> {
        let url = spawn_mock().await;
        CdpSession::connect(&url, "MOCK").await.unwrap()
    }

    #[tokio::test]
    async fn accepts_cdp_response_large_enough_for_maximum_reference_upload() {
        const ENCODED_UPLOAD_BYTES: usize = 68 * 1024 * 1024;
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (stream, _) = listener.accept().await.unwrap();
            let mut ws = accept_async(stream).await.unwrap();
            let request = ws.next().await.unwrap().unwrap();
            let Message::Text(text) = request else {
                panic!("expected text request");
            };
            let request: Value = serde_json::from_str(text.as_str()).unwrap();
            let response = json!({
                "id": request["id"],
                "result": { "payload": "x".repeat(ENCODED_UPLOAD_BYTES) }
            });
            ws.send(Message::text(response.to_string())).await.unwrap();
        });

        let session = CdpSession::connect(
            &format!("ws://127.0.0.1:{port}/devtools/page/LARGE"),
            "LARGE",
        )
        .await
        .unwrap();
        let response = session.send("Test.large", json!({})).await.unwrap();

        assert_eq!(
            response["payload"].as_str().unwrap().len(),
            ENCODED_UPLOAD_BYTES
        );
    }

    #[tokio::test]
    async fn roundtrip_returns_result() {
        let s = connect_mock().await;
        assert_eq!(s.target_id, "MOCK");
        let r = s.send("Echo.ping", json!({ "a": 1 })).await.unwrap();
        assert_eq!(r["pong"]["a"], 1);
        assert!(!s.is_closed());
    }

    #[tokio::test]
    async fn concurrent_sends_match_by_id() {
        let s = connect_mock().await;
        let mut handles = Vec::new();
        for i in 0..20u64 {
            let s = Arc::clone(&s);
            handles.push(tokio::spawn(async move {
                let r = s.send("Echo.ping", json!({ "i": i })).await.unwrap();
                assert_eq!(r["pong"]["i"], i);
            }));
        }
        for h in handles {
            h.await.unwrap();
        }
    }

    #[tokio::test]
    async fn protocol_error_is_mapped() {
        let s = connect_mock().await;
        let err = s.send("Echo.error", json!({})).await.unwrap_err();
        match err {
            CdpError::Protocol { code, message } => {
                assert_eq!(code, -32601);
                assert_eq!(message, "no such method");
            }
            other => panic!("unexpected {other:?}"),
        }
    }

    #[tokio::test]
    async fn events_are_broadcast() {
        let s = connect_mock().await;
        let mut rx = s.subscribe();
        s.send("Echo.event", json!({})).await.unwrap();
        let ev = tokio::time::timeout(Duration::from_secs(2), rx.recv())
            .await
            .unwrap()
            .unwrap();
        assert_eq!(ev.method, "Test.fired");
        assert_eq!(ev.params["n"], 1);
    }

    #[tokio::test]
    async fn send_timeout_yields_timeout_error() {
        let s = connect_mock().await;
        let err = s
            .send_timeout("Echo.hang", json!({}), Duration::from_millis(100))
            .await
            .unwrap_err();
        assert!(
            matches!(err, CdpError::Timeout(ref m) if m == "Echo.hang"),
            "{err:?}"
        );
        // 超时后 pending 已清理,连接仍可用。
        assert!(s.pending.lock().unwrap().is_empty());
        assert!(!s.is_closed());
        s.send("Echo.ping", json!({})).await.unwrap();
    }

    /// Task 6.1:调用方自己套 `tokio::time::timeout` / `select!` 把 `send` 的 future 丢掉时,
    /// pending 表里的条目也要清掉;否则长期存活的 chat session 每遇到一次挂起的求值就泄漏一条(24h 稳定性)。
    #[tokio::test]
    async fn dropped_send_future_removes_pending_entry() {
        let s = connect_mock().await;
        for _ in 0..5 {
            let r = tokio::time::timeout(
                Duration::from_millis(30),
                s.send_timeout("Echo.hang", json!({}), Duration::from_secs(60)),
            )
            .await;
            assert!(r.is_err(), "外层超时应先到");
        }
        assert!(
            s.pending.lock().unwrap().is_empty(),
            "被丢弃的调用不应留在 pending 里"
        );
        s.send("Echo.ping", json!({})).await.unwrap();
        assert!(s.pending.lock().unwrap().is_empty());
    }

    /// Task 6.1:事件广播容量 4096;不读的订阅者只会 `Lagged`,不阻塞读循环与命令响应。
    #[tokio::test]
    async fn slow_subscriber_lags_without_blocking_commands() {
        let s = connect_mock().await;
        let mut slow = s.subscribe();
        for _ in 0..(EVENT_CAPACITY + 10) {
            s.send("Echo.event", json!({})).await.unwrap();
        }
        match slow.recv().await {
            Err(broadcast::error::RecvError::Lagged(n)) => assert!(n >= 10, "skipped {n}"),
            other => panic!("应为 Lagged,实际 {other:?}"),
        }
        // 跳过积压后还能继续收新事件
        assert!(slow.recv().await.is_ok());
        s.send("Echo.ping", json!({})).await.unwrap();
    }

    #[tokio::test]
    async fn server_close_fails_pending_with_closed() {
        let s = connect_mock().await;
        let hang = {
            let s = Arc::clone(&s);
            tokio::spawn(async move { s.send("Echo.hang", json!({})).await })
        };
        tokio::time::sleep(Duration::from_millis(50)).await;
        let closing = s.send("Echo.close", json!({})).await;
        assert!(matches!(closing, Err(CdpError::Closed)), "{closing:?}");
        let hung = hang.await.unwrap();
        assert!(matches!(hung, Err(CdpError::Closed)), "{hung:?}");
        assert!(s.is_closed());
        assert!(matches!(
            s.send("Echo.ping", json!({})).await,
            Err(CdpError::Closed)
        ));
    }

    #[tokio::test]
    async fn client_close_fails_pending_with_closed() {
        let s = connect_mock().await;
        let hang = {
            let s = Arc::clone(&s);
            tokio::spawn(async move { s.send("Echo.hang", json!({})).await })
        };
        tokio::time::sleep(Duration::from_millis(50)).await;
        s.close().await;
        assert!(s.is_closed());
        let hung = hang.await.unwrap();
        assert!(matches!(hung, Err(CdpError::Closed)), "{hung:?}");
        assert!(matches!(
            s.send("Echo.ping", json!({})).await,
            Err(CdpError::Closed)
        ));
        // 重复 close 不 panic。
        s.close().await;
    }

    #[tokio::test]
    async fn evaluate_exception_is_js_exception_with_chinese_message() {
        let s = connect_mock().await;
        let err = s.evaluate("throw new Error('中文错误')").await.unwrap_err();
        assert!(
            matches!(err, CdpError::JsException(ref m) if m == "中文错误"),
            "{err:?}"
        );
    }

    #[derive(Debug, Deserialize)]
    struct Probe {
        ok: bool,
    }

    #[tokio::test]
    async fn evaluate_as_deserializes_value() {
        let s = connect_mock().await;
        let v = s.evaluate("1+1").await.unwrap();
        assert_eq!(v["ok"], true);
        let p: Probe = s.evaluate_as("1+1").await.unwrap();
        assert!(p.ok);
        #[derive(Debug, Deserialize)]
        struct Wrong {
            #[allow(dead_code)]
            missing: String,
        }
        let err = s.evaluate_as::<Wrong>("1+1").await.unwrap_err();
        assert!(matches!(err, CdpError::Deserialize(_)), "{err:?}");
    }

    #[tokio::test]
    async fn debug_log_records_send_recv_and_events() {
        let url = spawn_mock().await;
        let (log_tx, mut log_rx) = broadcast::channel(64);
        let s = CdpSession::connect_with_debug(&url, "MOCK", Some(log_tx))
            .await
            .unwrap();
        s.send("Echo.ping", json!({ "x": 1 })).await.unwrap();
        let sent = log_rx.recv().await.unwrap();
        assert_eq!(sent.direction, "send");
        assert_eq!(sent.method.as_deref(), Some("Echo.ping"));
        assert_eq!(sent.params.unwrap()["x"], 1);
        let recv = log_rx.recv().await.unwrap();
        assert_eq!(recv.direction, "recv");
        assert_eq!(recv.method.as_deref(), Some("Echo.ping"));
        assert_eq!(recv.result.unwrap()["pong"]["x"], 1);
        assert!(recv.event.is_none());

        s.send("Echo.event", json!({})).await.unwrap();
        let _sent = log_rx.recv().await.unwrap();
        let ev = log_rx.recv().await.unwrap();
        assert_eq!(ev.direction, "recv");
        assert_eq!(ev.method.as_deref(), Some("Test.fired"));
        assert_eq!(ev.event.unwrap()["n"], 1);
        let _resp = log_rx.recv().await.unwrap();
    }

    #[tokio::test]
    async fn connect_to_unreachable_port_is_connect_error() {
        let listener = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
        let port = listener.local_addr().unwrap().port();
        drop(listener);
        let err = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .expect_err("应连接失败");
        assert!(matches!(err, CdpError::Connect(_)), "{err:?}");
    }

    #[test]
    fn extract_js_exception_strips_error_prefix() {
        let d = json!({ "exception": { "description": "Error: 中文错误\n    at foo (<anonymous>:1:7)" } });
        assert_eq!(extract_js_exception(&d), "中文错误");
        let d =
            json!({ "exception": { "description": "TypeError: Cannot read properties of null" } });
        assert_eq!(extract_js_exception(&d), "Cannot read properties of null");
        let d = json!({ "exception": { "description": "账号按钮未找到: 请展开侧栏" } });
        assert_eq!(extract_js_exception(&d), "账号按钮未找到: 请展开侧栏");
        let d = json!({ "text": "Uncaught", "exception": { "type": "string", "value": "plain" } });
        assert_eq!(extract_js_exception(&d), "plain");
        let d = json!({ "text": "Uncaught" });
        assert_eq!(extract_js_exception(&d), "Uncaught");
        assert_eq!(extract_js_exception(&json!({})), "未知脚本异常");
    }

    /// 需要真豆包(端口 9707):`cargo test live_ -- --ignored --nocapture`。只读,不触发任何发送。
    #[tokio::test]
    #[ignore]
    async fn live_connect_chat_and_evaluate() {
        let targets = super::super::discovery::list_targets("127.0.0.1", 9707)
            .await
            .unwrap();
        let (chat, _) = super::super::discovery::select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let s = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        let targets = s.send("Target.getTargets", json!({})).await.unwrap();
        println!(
            "Target.getTargets: {}",
            targets["targetInfos"]
                .as_array()
                .map(|a| a.len())
                .unwrap_or(0)
        );
        let v = s
            .evaluate("Promise.resolve(42).then((p) => ({ href: location.href, title: document.title, ready: document.readyState, p }))")
            .await
            .unwrap();
        println!("evaluate: {v}");
        assert_eq!(v["p"], 42);
        assert!(v["href"].as_str().unwrap().contains("doubao-chat"));
        let err = s.evaluate("throw new Error('探针异常')").await.unwrap_err();
        println!("exception: {err:?}");
        assert!(matches!(err, CdpError::JsException(ref m) if m == "探针异常"));
        s.close().await;
        assert!(s.is_closed());
    }
}
