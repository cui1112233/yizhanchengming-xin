//! CDP 发现接口:`GET /json/version`、`GET /json/list` 与豆包 target 白名单。
//!
//! 白名单规则(Global Constraints):`type=="page"`,url scheme ∈ {doubao, chrome},
//! host ∈ {doubao-chat, doubao-launcher, doubao-background};`webSocketDebuggerUrl` 必须是 `ws://` + loopback + 配置端口。
//! 真机(macOS 2.32.5)`/json/list` 还会返回 `iframe`(`https://www.doubao.com/...`)与
//! `other`(`doubao://doubao-chat/cross-site-support/`)类型的 target,均被过滤。

use std::time::Duration;

use serde::Deserialize;
use serde_json::Value;
use url::{Host, Url};

use super::CdpError;

/// `/json/version` 超时。
const VERSION_TIMEOUT: Duration = Duration::from_secs(2);
/// `/json/list` 超时。
const LIST_TIMEOUT: Duration = Duration::from_secs(3);

const ALLOWED_SCHEMES: &[&str] = &["doubao", "chrome"];
const LOOPBACK_HOSTS: &[&str] = &["127.0.0.1", "localhost", "::1"];

/// `/json/list` 返回的单个 target(只保留关心的字段,其余忽略)。
#[derive(Clone, Debug, Deserialize)]
pub struct CdpTarget {
    pub id: String,
    #[serde(rename = "type")]
    pub kind: String,
    pub url: String,
    #[serde(default)]
    pub title: String,
    #[serde(rename = "webSocketDebuggerUrl", default)]
    pub ws_url: String,
}

/// 豆包的两个页面。
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum DoubaoPage {
    Chat,
    Background,
}

fn http_base(host: &str, port: u16) -> String {
    if host.contains(':') && !host.starts_with('[') {
        format!("http://[{host}]:{port}")
    } else {
        format!("http://{host}:{port}")
    }
}

async fn get_json(url: &str, timeout: Duration) -> Result<Value, CdpError> {
    let client = reqwest::Client::builder()
        .timeout(timeout)
        .no_proxy()
        .build()
        .map_err(|e| CdpError::Http(e.to_string()))?;
    let resp = client
        .get(url)
        .send()
        .await
        .map_err(|e| CdpError::Http(format!("{url}: {e}")))?;
    let status = resp.status();
    if !status.is_success() {
        return Err(CdpError::Http(format!("{url}: HTTP {status}")));
    }
    resp.json::<Value>()
        .await
        .map_err(|e| CdpError::Http(format!("{url}: {e}")))
}

/// `GET /json/version`(2s 超时),返回原始 JSON(含 `Browser`、`webSocketDebuggerUrl` 等)。
pub async fn check_version(host: &str, port: u16) -> Result<Value, CdpError> {
    let url = format!("{}/json/version", http_base(host, port));
    get_json(&url, VERSION_TIMEOUT).await
}

/// `GET /json/list`(3s 超时),返回全部 target(未过滤)。
pub async fn list_targets(host: &str, port: u16) -> Result<Vec<CdpTarget>, CdpError> {
    let url = format!("{}/json/list", http_base(host, port));
    let value = get_json(&url, LIST_TIMEOUT).await?;
    serde_json::from_value::<Vec<CdpTarget>>(value)
        .map_err(|e| CdpError::Http(format!("{url}: 响应格式不正确:{e}")))
}

fn is_loopback(url: &Url) -> bool {
    match url.host() {
        Some(Host::Ipv4(ip)) => ip.is_loopback(),
        Some(Host::Ipv6(ip)) => ip.is_loopback(),
        Some(Host::Domain(d)) => LOOPBACK_HOSTS.contains(&d),
        None => false,
    }
}

/// 按白名单校验一个 target,合法则返回它是哪个豆包页面。
pub fn validate_target(t: &CdpTarget, port: u16) -> Result<DoubaoPage, CdpError> {
    if t.kind != "page" {
        return Err(CdpError::InvalidTarget(format!(
            "type 为 {},只接受 page",
            t.kind
        )));
    }
    let url = Url::parse(&t.url)
        .map_err(|e| CdpError::InvalidTarget(format!("url 无法解析 {}:{e}", t.url)))?;
    if !ALLOWED_SCHEMES.contains(&url.scheme()) {
        return Err(CdpError::InvalidTarget(format!(
            "url scheme 为 {},只接受 doubao/chrome",
            url.scheme()
        )));
    }
    let page = match url.host_str() {
        Some("doubao-chat" | "doubao-launcher") => DoubaoPage::Chat,
        Some("doubao-background") => DoubaoPage::Background,
        other => {
            return Err(CdpError::InvalidTarget(format!(
                "url host 为 {},不在白名单内",
                other.unwrap_or("")
            )))
        }
    };
    let ws = Url::parse(&t.ws_url).map_err(|e| {
        CdpError::InvalidTarget(format!("webSocketDebuggerUrl 无法解析 {}:{e}", t.ws_url))
    })?;
    if ws.scheme() != "ws" {
        return Err(CdpError::InvalidTarget(format!(
            "webSocketDebuggerUrl scheme 为 {},只接受 ws",
            ws.scheme()
        )));
    }
    if !is_loopback(&ws) {
        return Err(CdpError::InvalidTarget(format!(
            "webSocketDebuggerUrl host 为 {},只接受本机回环地址",
            ws.host_str().unwrap_or("")
        )));
    }
    match ws.port_or_known_default() {
        Some(p) if p == port => {}
        p => {
            return Err(CdpError::InvalidTarget(format!(
                "webSocketDebuggerUrl 端口为 {},期望 {port}",
                p.map(|p| p.to_string()).unwrap_or_default()
            )))
        }
    }
    Ok(page)
}

/// 从 target 列表里挑出合法的 chat / background 页面(不合法的静默丢弃)。
pub fn select_pages(targets: &[CdpTarget], port: u16) -> (Vec<CdpTarget>, Vec<CdpTarget>) {
    let mut chat = Vec::new();
    let mut launcher_chat = Vec::new();
    let mut background = Vec::new();
    for t in targets {
        match validate_target(t, port) {
            Ok(DoubaoPage::Chat) => {
                if Url::parse(&t.url)
                    .ok()
                    .and_then(|url| url.host_str().map(str::to_owned))
                    == Some("doubao-launcher".into())
                {
                    launcher_chat.push(t.clone());
                } else {
                    chat.push(t.clone());
                }
            }
            Ok(DoubaoPage::Background) => background.push(t.clone()),
            Err(_) => {}
        }
    }
    // 生成与账号操作必须落在真正的 doubao-chat 页面。新版桌面端同时暴露
    // launcher 页面时，launcher 只有导航壳，账号按钮和生成输入区不完整。
    // 只有没有普通 chat 页面时，才退回 launcher。
    if chat.is_empty() && !launcher_chat.is_empty() {
        chat = launcher_chat;
    }
    (chat, background)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn target(kind: &str, url: &str, ws: &str) -> CdpTarget {
        CdpTarget {
            id: "X".into(),
            kind: kind.into(),
            url: url.into(),
            title: String::new(),
            ws_url: ws.into(),
        }
    }

    const WS: &str = "ws://127.0.0.1:9707/devtools/page/X";

    #[test]
    fn chat_page_on_loopback_is_valid() {
        let t = target("page", "doubao://doubao-chat/chat", WS);
        assert_eq!(validate_target(&t, 9707).unwrap(), DoubaoPage::Chat);
    }

    #[test]
    fn chrome_scheme_background_is_valid() {
        let t = target("page", "chrome://doubao-background/", WS);
        assert_eq!(validate_target(&t, 9707).unwrap(), DoubaoPage::Background);
    }

    #[test]
    fn localhost_and_ipv6_loopback_ws_are_valid() {
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "ws://localhost:9707/devtools/page/X",
        );
        assert!(validate_target(&t, 9707).is_ok());
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "ws://[::1]:9707/devtools/page/X",
        );
        assert!(validate_target(&t, 9707).is_ok());
    }

    #[test]
    fn iframe_is_rejected() {
        let t = target("iframe", "doubao://doubao-chat/chat", WS);
        assert!(matches!(
            validate_target(&t, 9707),
            Err(CdpError::InvalidTarget(_))
        ));
        let t = target("other", "doubao://doubao-chat/cross-site-support/", WS);
        assert!(validate_target(&t, 9707).is_err());
    }

    #[test]
    fn foreign_host_is_rejected() {
        let t = target("page", "https://www.doubao.com/client-login-modal", WS);
        assert!(validate_target(&t, 9707).is_err());
        let t = target("page", "doubao://www.doubao.com/chat", WS);
        assert!(validate_target(&t, 9707).is_err());
    }

    #[test]
    fn wrong_ws_port_is_rejected() {
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "ws://127.0.0.1:9708/devtools/page/X",
        );
        let err = validate_target(&t, 9707).unwrap_err();
        assert!(err.to_string().contains("9708"), "{err}");
    }

    #[test]
    fn non_loopback_ws_host_is_rejected() {
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "ws://0.0.0.0:9707/devtools/page/X",
        );
        assert!(validate_target(&t, 9707).is_err());
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "ws://10.0.0.5:9707/devtools/page/X",
        );
        assert!(validate_target(&t, 9707).is_err());
        let t = target(
            "page",
            "doubao://doubao-chat/chat",
            "wss://127.0.0.1:9707/devtools/page/X",
        );
        assert!(validate_target(&t, 9707).is_err());
    }

    #[test]
    fn empty_ws_url_is_rejected() {
        let t = target("page", "doubao://doubao-chat/chat", "");
        assert!(validate_target(&t, 9707).is_err());
    }

    #[test]
    fn select_pages_partitions_live_shaped_list() {
        let json = serde_json::json!([
            {"id":"A","type":"page","url":"doubao://doubao-chat/chat","title":"豆包",
             "webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/A","description":""},
            {"id":"B","type":"iframe","url":"https://www.doubao.com/drive-iframe/drive/home/","parentId":"A",
             "webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/B"},
            {"id":"C","type":"other","url":"doubao://doubao-chat/cross-site-support/",
             "webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/C"},
            {"id":"D","type":"page","url":"doubao://doubao-background/",
             "webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/D"},
            {"id":"E","type":"page","url":"https://www.doubao.com/flow-account/client-login/client-login-modal",
             "webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/E"}
        ]);
        let targets: Vec<CdpTarget> = serde_json::from_value(json).unwrap();
        assert_eq!(targets.len(), 5);
        assert_eq!(targets[0].title, "豆包");
        let (chat, background) = select_pages(&targets, 9707);
        assert_eq!(chat.len(), 1);
        assert_eq!(chat[0].id, "A");
        assert_eq!(background.len(), 1);
        assert_eq!(background[0].id, "D");
        let (chat, background) = select_pages(&targets, 9708);
        assert!(chat.is_empty() && background.is_empty());
    }

    #[test]
    fn select_pages_prefers_real_chat_when_launcher_and_chat_coexist() {
        let targets = vec![
            CdpTarget {
                id: "LEGACY".into(),
                kind: "page".into(),
                url: "doubao://doubao-chat/chat".into(),
                title: "豆包".into(),
                ws_url: "ws://127.0.0.1:9707/devtools/page/LEGACY".into(),
            },
            CdpTarget {
                id: "LAUNCHER".into(),
                kind: "page".into(),
                url: "doubao://doubao-launcher/chat?viewId=101".into(),
                title: "豆包".into(),
                ws_url: "ws://127.0.0.1:9707/devtools/page/LAUNCHER".into(),
            },
        ];

        let (chat, background) = select_pages(&targets, 9707);

        assert_eq!(chat.len(), 1);
        assert_eq!(chat[0].id, "LEGACY");
        assert!(background.is_empty());
    }

    #[test]
    fn http_base_brackets_ipv6() {
        assert_eq!(http_base("127.0.0.1", 9707), "http://127.0.0.1:9707");
        assert_eq!(http_base("::1", 9707), "http://[::1]:9707");
        assert_eq!(http_base("localhost", 1), "http://localhost:1");
    }

    #[tokio::test]
    async fn unreachable_port_yields_http_error() {
        // 绑定后立刻释放的端口几乎不可能被占用。
        let listener = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
        let port = listener.local_addr().unwrap().port();
        drop(listener);
        assert!(matches!(
            check_version("127.0.0.1", port).await,
            Err(CdpError::Http(_))
        ));
        assert!(matches!(
            list_targets("127.0.0.1", port).await,
            Err(CdpError::Http(_))
        ));
    }

    #[tokio::test]
    async fn list_targets_parses_mock_http_server() {
        use tokio::io::{AsyncReadExt, AsyncWriteExt};
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            loop {
                let (mut sock, _) = listener.accept().await.unwrap();
                let mut buf = vec![0u8; 4096];
                let n = sock.read(&mut buf).await.unwrap_or(0);
                let req = String::from_utf8_lossy(&buf[..n]).to_string();
                let body = if req.starts_with("GET /json/version") {
                    r#"{"Browser":"Chrome/147.0.0.0","Protocol-Version":"1.3"}"#.to_string()
                } else {
                    r#"[{"id":"A","type":"page","url":"doubao://doubao-chat/chat","webSocketDebuggerUrl":"ws://127.0.0.1:9707/devtools/page/A"}]"#.to_string()
                };
                let resp = format!(
                    "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                    body.len(),
                    body
                );
                let _ = sock.write_all(resp.as_bytes()).await;
                let _ = sock.shutdown().await;
            }
        });
        let v = check_version("127.0.0.1", port).await.unwrap();
        assert_eq!(v["Protocol-Version"], "1.3");
        let targets = list_targets("127.0.0.1", port).await.unwrap();
        assert_eq!(targets.len(), 1);
        assert_eq!(targets[0].kind, "page");
        assert_eq!(targets[0].ws_url, "ws://127.0.0.1:9707/devtools/page/A");
    }

    /// 需要真豆包:`cargo test live_ -- --ignored --nocapture`。
    #[tokio::test]
    #[ignore]
    async fn live_list_targets_has_chat_and_background() {
        let v = check_version("127.0.0.1", 9707).await.unwrap();
        println!("version: {v}");
        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        for t in &targets {
            println!("{} {} {}", t.kind, t.url, t.ws_url);
        }
        let (chat, background) = select_pages(&targets, 9707);
        assert_eq!(chat.len(), 1, "应恰好一个 doubao-chat page");
        assert_eq!(background.len(), 1, "应恰好一个 doubao-background page");
    }
}
