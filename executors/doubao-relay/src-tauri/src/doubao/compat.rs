//! UI 兼容性自检(spec 09 "契约版本化" / 05 `doubao.status.uiCompatDetails`)。
//!
//! 注入 `accounts.js` 后调 `probe()` 看账号按钮与输入框,再找发送按钮。
//! macOS 2.32.5 实测(`docs/analysis/10-live-probe-macos.md` Task 1.6):发送按钮是
//! `button[data-testid="chat_input_send_button"]`(在 `.send-btn-wrapper` 内,`aria-label=""`,
//! 无 `type=submit`),**只在输入框有文字时才渲染**,空输入时同一槽位是麦克风 `div`。
//! 所以直接 selector 没命中且输入框为空时,临时插入一个字符再清空,以此确认发送按钮存在。

use serde::Serialize;

use crate::cdp::{now_ms, CdpSession};

use super::ACCOUNTS_SCRIPT;

/// 发送按钮 selector(按可靠度排序;后两个是计划里的通用兜底)。
pub const SEND_BUTTON_SELECTOR: &str = r#"[data-testid="chat_input_send_button"],.send-btn-wrapper button,button[aria-label*="发送"],button[type="submit"]"#;

/// `doubao.status.uiCompatDetails` + `uiCompatCheck`。
#[derive(Clone, Debug, Serialize, Default, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct UiCompat {
    /// `ok` | `warn` | `error` | `unknown`(自检本身没跑成)。
    pub status: String,
    pub account_button_found: bool,
    pub input_found: bool,
    pub send_button_found: bool,
    /// Unix epoch 毫秒。
    pub checked_at: i64,
}

/// 三者全有 → `ok`;账号按钮 + 输入框有 → `warn`;否则 `error`。
pub fn compat_status(account_button: bool, input: bool, send_button: bool) -> &'static str {
    if account_button && input && send_button {
        "ok"
    } else if account_button && input {
        "warn"
    } else {
        "error"
    }
}

#[derive(Debug, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
struct CompatProbe {
    account_button_found: bool,
    input_found: bool,
    send_button_found: bool,
    #[serde(default)]
    send_via: String,
}

/// 自检表达式(async IIFE,返回 Promise;`CdpSession::evaluate` 带 `awaitPromise`)。
fn compat_expression() -> String {
    format!(
        r#"(async () => {{
  const SEND = {send_sel:?};
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const visible = el => {{ const r = el.getBoundingClientRect(); return r.width > 0 && r.height > 0; }};
  const findSend = () => [...document.querySelectorAll(SEND)].some(visible);
  const p = window.__doubaoRelayAccounts.probe();
  let send = findSend();
  let via = send ? 'direct' : 'none';
  if (!send) {{
    const ed = document.querySelector('[contenteditable="true"]');
    const empty = !!ed && !(ed.innerText || ed.textContent || '').trim();
    if (empty) {{
      const hadFocus = document.activeElement === ed;
      try {{
        ed.focus();
        document.execCommand('insertText', false, 'x');
        await sleep(250);
        send = findSend();
        via = send ? 'transient_text' : 'none';
      }} finally {{
        document.execCommand('selectAll');
        document.execCommand('delete');
        if (!hadFocus) ed.blur();
      }}
    }}
  }}
  return {{ accountButtonFound: !!p.buttonFound, inputFound: !!p.hasInput, sendButtonFound: send, sendVia: via }};
}})()"#,
        send_sel = SEND_BUTTON_SELECTOR
    )
}

/// 跑一次 UI 兼容性自检。永不返回 Err:CDP/脚本失败时 `status="unknown"`、三项全 false。
pub async fn ui_compat_check(session: &CdpSession) -> UiCompat {
    let checked_at = now_ms();
    let probe = async {
        session.evaluate(ACCOUNTS_SCRIPT).await?;
        session
            .evaluate_as::<CompatProbe>(&compat_expression())
            .await
    }
    .await;
    match probe {
        Ok(p) => {
            let status = compat_status(p.account_button_found, p.input_found, p.send_button_found);
            tracing::info!(
                status,
                account_button_found = p.account_button_found,
                input_found = p.input_found,
                send_button_found = p.send_button_found,
                send_via = p.send_via.as_str(),
                "UI 兼容性自检完成"
            );
            UiCompat {
                status: status.to_string(),
                account_button_found: p.account_button_found,
                input_found: p.input_found,
                send_button_found: p.send_button_found,
                checked_at,
            }
        }
        Err(e) => {
            tracing::warn!(error = %e, "UI 兼容性自检未能执行");
            UiCompat {
                status: "unknown".to_string(),
                checked_at,
                ..Default::default()
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn status_table() {
        assert_eq!(compat_status(true, true, true), "ok");
        assert_eq!(compat_status(true, true, false), "warn");
        assert_eq!(compat_status(true, false, true), "error");
        assert_eq!(compat_status(false, true, true), "error");
        assert_eq!(compat_status(false, false, false), "error");
    }

    #[test]
    fn serializes_snake_case_fields() {
        let c = UiCompat {
            status: "ok".into(),
            account_button_found: true,
            input_found: true,
            send_button_found: true,
            checked_at: 1,
        };
        let v = serde_json::to_value(c).unwrap();
        let obj = v.as_object().unwrap();
        assert_eq!(obj.len(), 5);
        assert_eq!(obj["status"], "ok");
        assert_eq!(obj["account_button_found"], true);
        assert_eq!(obj["send_button_found"], true);
        assert_eq!(obj["checked_at"], 1);
    }

    #[test]
    fn expression_embeds_selector_and_is_async_iife() {
        let e = compat_expression();
        assert!(e.starts_with("(async () => {"));
        assert!(e.ends_with("})()"));
        assert!(e.contains("chat_input_send_button"));
        assert!(e.contains("__doubaoRelayAccounts.probe()"));
        // Runtime.evaluate 无 replMode,不能有顶层 await:所有 await 都在 async 函数体内
        assert!(!e.trim_start_matches("(async () => {").starts_with("await"));
    }

    #[tokio::test]
    async fn unknown_when_session_closed() {
        // mock WS:握手后立刻关闭,evaluate 得 Closed → status unknown
        use futures_util::StreamExt;
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            let _ = ws.next().await;
            let _ = ws.close(None).await;
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        let c = ui_compat_check(&session).await;
        assert_eq!(c.status, "unknown");
        assert!(!c.account_button_found && !c.input_found && !c.send_button_found);
        assert!(c.checked_at > 0);
    }

    #[tokio::test]
    async fn ok_when_mock_page_answers_probe() {
        // mock WS:对 Runtime.evaluate 一律回 {ok:true} 或探针对象
        use futures_util::{SinkExt, StreamExt};
        use serde_json::{json, Value};
        use tokio_tungstenite::tungstenite::Message;
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let expr = req["params"]["expression"].as_str().unwrap_or("");
                let value = if expr.contains("probe()") {
                    json!({"accountButtonFound": true, "inputFound": true, "sendButtonFound": true, "sendVia": "direct"})
                } else {
                    json!({"ok": true})
                };
                let resp = json!({"id": req["id"], "result": {"result": {"type": "object", "value": value}}});
                ws.send(Message::Text(resp.to_string().into()))
                    .await
                    .unwrap();
            }
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        let c = ui_compat_check(&session).await;
        assert_eq!(c.status, "ok");
        assert!(c.account_button_found && c.input_found && c.send_button_found);
        session.close().await;
    }

    /// 需要真豆包:`cargo test live_ -- --ignored --nocapture`。
    #[tokio::test]
    #[ignore]
    async fn live_compat() {
        use crate::cdp::{list_targets, select_pages};
        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        let c = ui_compat_check(&session).await;
        println!("compat: {c:?}");
        session.close().await;
        assert_eq!(c.status, "ok");
    }
}
