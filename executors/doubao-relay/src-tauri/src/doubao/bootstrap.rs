//! 启动豆包并等待 CDP 就绪(`docs/analysis/02-cdp-bootstrap.md` "Ensure-DoubaoStarted" 的 Rust 版)。
//!
//! 决策树:
//! 1. `/json/version` 通 且 监听者是豆包 → `AlreadyConnected`
//! 2. `/json/version` 通 但 监听者不是豆包 → `doubao_cdp_port_occupied`
//! 3. 豆包在跑但端口不通 → 杀豆包 → 带调试参数启动 → 等 CDP(20s)→ `Restarted`
//! 4. 豆包没跑 → 找可执行 → 启动 → 等 CDP(20s)→ `Started`;超时 → `doubao_launch_timeout`

use std::path::Path;
use std::time::{Duration, Instant};

use serde::Serialize;
use tokio::runtime::{Handle, RuntimeFlavor};

use crate::cdp::{check_version, list_targets, select_pages};
use crate::error::{RelayError, RelayResult};
use crate::platform::{Platform, PlatformError};

use super::CDP_HOST;

/// 启动后等待 CDP 就绪的上限(spec:20s)。
pub const LAUNCH_TIMEOUT: Duration = Duration::from_secs(20);
/// `wait_for_cdp` 轮询间隔。
const POLL_INTERVAL: Duration = Duration::from_millis(250);

/// `ensure_doubao_started` 的结果:豆包是怎么起来的。
#[derive(Clone, Copy, Debug, Serialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum StartupAction {
    /// 豆包已在跑且 CDP 可连。
    AlreadyConnected,
    /// 豆包在跑,只是重新建立了 CDP 会话(由 Task 1.8 的重连流程返回)。
    Reconnected,
    /// 豆包在跑但没开 CDP 端口,被杀掉后带调试参数重启。
    Restarted,
    /// 豆包原本没跑,被带调试参数启动。
    Started,
}

/// 确保豆包以 CDP 调试端口运行。`port` 为配置的 CDP 端口(默认 9707)。
pub async fn ensure_doubao_started(
    platform: &dyn Platform,
    port: u16,
) -> RelayResult<StartupAction> {
    ensure_doubao_started_with(platform, port, LAUNCH_TIMEOUT).await
}

/// 同 [`ensure_doubao_started`],可指定等待 CDP 就绪的上限(测试用)。
pub async fn ensure_doubao_started_with(
    platform: &dyn Platform,
    port: u16,
    launch_timeout: Duration,
) -> RelayResult<StartupAction> {
    let owner = run_blocking(|| platform.listening_process_exe(port));
    let owner_is_doubao = owner
        .as_deref()
        .map(|p| platform.is_doubao_exe_path(p))
        .unwrap_or(false);

    match check_version(CDP_HOST, port).await {
        Ok(version) => {
            if owner_is_doubao {
                tracing::info!(
                    port,
                    browser = version
                        .get("Browser")
                        .and_then(|v| v.as_str())
                        .unwrap_or(""),
                    "豆包已在运行且 CDP 端口可连"
                );
                return Ok(StartupAction::AlreadyConnected);
            }
            match owner {
                Some(exe) => {
                    tracing::warn!(port, owner = %exe.display(), "CDP 端口被非豆包进程占用");
                    Err(port_occupied(port, exe.display().to_string()))
                }
                None => {
                    // 查不到监听者(lsof/netstat 失败或权限不足):用 target 列表兜底判断是不是豆包。
                    let has_chat = list_targets(CDP_HOST, port)
                        .await
                        .map(|t| !select_pages(&t, port).0.is_empty())
                        .unwrap_or(false);
                    if has_chat {
                        tracing::warn!(
                            port,
                            "无法确定 CDP 端口监听进程,但 target 列表含 doubao-chat,按豆包处理"
                        );
                        Ok(StartupAction::AlreadyConnected)
                    } else {
                        tracing::warn!(port, "CDP 端口已被占用,且 target 列表里没有豆包页面");
                        Err(port_occupied(port, "unknown".to_string()))
                    }
                }
            }
        }
        Err(e) => {
            tracing::debug!(port, error = %e, "CDP 端口不通");
            if let (Some(exe), false) = (&owner, owner_is_doubao) {
                // 端口被别的程序占着(不是 CDP 服务),豆包起来也开不了这个端口。
                tracing::warn!(port, owner = %exe.display(), "CDP 端口被非豆包进程占用");
                return Err(port_occupied(port, exe.display().to_string()));
            }
            let running = platform.is_doubao_running().await;
            // 先定位可执行文件:macOS 的 find_doubao_exe 第一来源是"已运行进程",要在杀之前找。
            let exe = platform.find_doubao_exe().await?;
            if running {
                tracing::info!(port, exe = %exe.display(), "豆包在运行但未开 CDP 端口,准备重启");
                platform.kill_doubao().await?;
            } else {
                tracing::info!(port, exe = %exe.display(), "豆包未运行,准备启动");
            }
            launch_and_wait(platform, &exe, port, launch_timeout).await?;
            Ok(if running {
                StartupAction::Restarted
            } else {
                StartupAction::Started
            })
        }
    }
}

async fn launch_and_wait(
    platform: &dyn Platform,
    exe: &Path,
    port: u16,
    launch_timeout: Duration,
) -> RelayResult<()> {
    let started = Instant::now();
    platform.launch_doubao_with_cdp(exe, port).await?;
    wait_for_cdp(CDP_HOST, port, launch_timeout).await?;
    tracing::info!(
        port,
        elapsed_ms = started.elapsed().as_millis() as u64,
        "豆包已启动,CDP 就绪"
    );
    Ok(())
}

/// spec 09"安全校验(每次重连都要重跑)"/ Global Constraint"9707 监听进程的可执行路径必须位于豆包安装目录下"。
///
/// 监听者可确定且不是豆包 → `doubao_cdp_port_occupied{owning_process, port}`;是豆包 → Ok;
/// 查不到监听者(lsof/netstat 失败、权限不足或尚无人监听)→ Ok,由调用方用"target 列表里有
/// doubao-chat"兜底(重连路径在调用前已确认恰好 1 个 chat page,`wait_for_cdp` 也只在出现 chat
/// page 后返回)。
pub async fn verify_cdp_owner(platform: &dyn Platform, port: u16) -> RelayResult<()> {
    let owner = run_blocking(|| platform.listening_process_exe(port));
    match owner {
        Some(exe) if !platform.is_doubao_exe_path(&exe) => {
            tracing::warn!(port, owner = %exe.display(), "CDP 端口监听进程不是豆包,拒绝连接");
            Err(port_occupied(port, exe.display().to_string()))
        }
        _ => Ok(()),
    }
}

fn port_occupied(port: u16, owner: String) -> RelayError {
    PlatformError::PortOccupied(port, owner).into()
}

/// 每 250ms 请求一次 `/json/list`,直到出现合法的 `doubao-chat` page;超时 → `doubao_launch_timeout`。
pub async fn wait_for_cdp(host: &str, port: u16, timeout: Duration) -> RelayResult<()> {
    let deadline = Instant::now() + timeout;
    loop {
        let remaining = deadline.saturating_duration_since(Instant::now());
        if remaining.is_zero() {
            break;
        }
        if let Ok(Ok(targets)) = tokio::time::timeout(remaining, list_targets(host, port)).await {
            let (chat, _) = select_pages(&targets, port);
            if !chat.is_empty() {
                return Ok(());
            }
        }
        let remaining = deadline.saturating_duration_since(Instant::now());
        if remaining.is_zero() {
            break;
        }
        tokio::time::sleep(POLL_INTERVAL.min(remaining)).await;
    }
    tracing::warn!(
        port,
        timeout_ms = timeout.as_millis() as u64,
        "等待豆包 CDP 就绪超时"
    );
    Err(RelayError::doubao_launch_timeout())
}

/// 在 async 上下文里跑阻塞的平台调用(`listening_process_exe` 约 45ms)。
/// `platform` 是借用,不能丢进 `spawn_blocking`;多线程运行时用 `block_in_place`,
/// 单线程运行时(测试)直接调用。
fn run_blocking<T>(f: impl FnOnce() -> T) -> T {
    match Handle::try_current().map(|h| h.runtime_flavor()) {
        Ok(RuntimeFlavor::MultiThread) => tokio::task::block_in_place(f),
        _ => f(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::{Path, PathBuf};
    use std::sync::atomic::{AtomicUsize, Ordering};
    use std::sync::Arc;

    use async_trait::async_trait;
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::net::TcpListener;

    const CHAT_LIST: &str = r#"[{"id":"A","type":"page","url":"doubao://doubao-chat/chat","webSocketDebuggerUrl":"ws://127.0.0.1:PORT/devtools/page/A"},{"id":"D","type":"page","url":"doubao://doubao-background/","webSocketDebuggerUrl":"ws://127.0.0.1:PORT/devtools/page/D"}]"#;
    const NO_CHAT_LIST: &str = r#"[{"id":"D","type":"page","url":"doubao://doubao-background/","webSocketDebuggerUrl":"ws://127.0.0.1:PORT/devtools/page/D"}]"#;

    /// 极简 HTTP mock:`/json/version` 回版本,`/json/list` 回 `list_body`(`PORT` 占位替换为真实端口)。
    fn spawn_mock_cdp_http(listener: TcpListener, list_body: &'static str) {
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            loop {
                let Ok((mut sock, _)) = listener.accept().await else {
                    break;
                };
                let list_body = list_body.replace("PORT", &port.to_string());
                tokio::spawn(async move {
                    let mut buf = vec![0u8; 4096];
                    let n = sock.read(&mut buf).await.unwrap_or(0);
                    let req = String::from_utf8_lossy(&buf[..n]).to_string();
                    let body = if req.starts_with("GET /json/version") {
                        r#"{"Browser":"Chrome/147.0.0.0","Protocol-Version":"1.3"}"#.to_string()
                    } else {
                        list_body
                    };
                    let resp = format!(
                        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                        body.len(),
                        body
                    );
                    let _ = sock.write_all(resp.as_bytes()).await;
                    let _ = sock.shutdown().await;
                });
            }
        });
    }

    fn free_port() -> u16 {
        let l = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
        let port = l.local_addr().unwrap().port();
        drop(l);
        port
    }

    #[tokio::test]
    async fn wait_for_cdp_times_out_on_unused_port() {
        let port = free_port();
        let started = Instant::now();
        let err = wait_for_cdp("127.0.0.1", port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_launch_timeout");
        let elapsed = started.elapsed();
        assert!(
            elapsed >= Duration::from_millis(300) && elapsed < Duration::from_secs(2),
            "elapsed {elapsed:?}"
        );
    }

    #[tokio::test]
    async fn wait_for_cdp_ok_when_chat_page_listed() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, CHAT_LIST);
        wait_for_cdp("127.0.0.1", port, Duration::from_secs(3))
            .await
            .unwrap();
    }

    #[tokio::test]
    async fn wait_for_cdp_times_out_when_only_background_listed() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, NO_CHAT_LIST);
        let err = wait_for_cdp("127.0.0.1", port, Duration::from_millis(600))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_launch_timeout");
    }

    #[test]
    fn startup_action_serializes_snake_case() {
        assert_eq!(
            serde_json::to_value(StartupAction::AlreadyConnected).unwrap(),
            "already_connected"
        );
        assert_eq!(
            serde_json::to_value(StartupAction::Reconnected).unwrap(),
            "reconnected"
        );
        assert_eq!(
            serde_json::to_value(StartupAction::Restarted).unwrap(),
            "restarted"
        );
        assert_eq!(
            serde_json::to_value(StartupAction::Started).unwrap(),
            "started"
        );
    }

    // ---- 假平台:驱动 ensure_doubao_started 的四条分支 ----

    #[derive(Default)]
    struct FakePlatform {
        running: bool,
        listening_exe: Option<PathBuf>,
        exe_found: bool,
        /// launch 时在该端口起 mock CDP HTTP(模拟豆包真的起来了)。
        serve_on_launch: bool,
        kill_calls: AtomicUsize,
        launch_calls: AtomicUsize,
        find_calls: AtomicUsize,
    }

    #[async_trait]
    impl Platform for FakePlatform {
        async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
            self.find_calls.fetch_add(1, Ordering::SeqCst);
            if self.exe_found {
                Ok(PathBuf::from(
                    "/Applications/Doubao.app/Contents/MacOS/Doubao",
                ))
            } else {
                Err(PlatformError::DoubaoNotFound("/a, /b".into()))
            }
        }
        async fn launch_doubao_with_cdp(
            &self,
            _exe: &Path,
            port: u16,
        ) -> Result<(), PlatformError> {
            self.launch_calls.fetch_add(1, Ordering::SeqCst);
            if self.serve_on_launch {
                let listener = TcpListener::bind(("127.0.0.1", port))
                    .await
                    .map_err(|e| PlatformError::LaunchFailed(e.to_string()))?;
                spawn_mock_cdp_http(listener, CHAT_LIST);
            }
            Ok(())
        }
        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            self.kill_calls.fetch_add(1, Ordering::SeqCst);
            Ok(())
        }
        async fn is_doubao_running(&self) -> bool {
            self.running
        }
        fn app_data_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay")
        }
        fn log_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/logs")
        }
        fn config_path(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/config.toml")
        }
        async fn reveal_in_file_manager(&self, _path: &Path) -> Result<(), PlatformError> {
            Ok(())
        }
        fn platform_name(&self) -> &'static str {
            "fake"
        }
        fn listening_process_exe(&self, _port: u16) -> Option<PathBuf> {
            self.listening_exe.clone()
        }
        fn is_doubao_exe_path(&self, p: &Path) -> bool {
            p.components()
                .any(|c| c.as_os_str().to_string_lossy() == "Doubao.app")
        }
        fn doubao_version(&self) -> Option<String> {
            None
        }
    }

    const DOUBAO_BROWSER: &str =
        "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser";

    #[tokio::test]
    async fn already_connected_when_port_served_by_doubao() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, CHAT_LIST);
        let platform = FakePlatform {
            running: true,
            listening_exe: Some(PathBuf::from(DOUBAO_BROWSER)),
            exe_found: true,
            ..Default::default()
        };
        let action = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap();
        assert_eq!(action, StartupAction::AlreadyConnected);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn port_occupied_when_cdp_served_by_other_process() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, CHAT_LIST);
        let platform = FakePlatform {
            running: false,
            listening_exe: Some(PathBuf::from(
                "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
            )),
            exe_found: true,
            ..Default::default()
        };
        let err = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_cdp_port_occupied");
        let details = err.details.unwrap();
        assert_eq!(
            details["owning_process"],
            "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
        );
        assert_eq!(details["port"], port);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn unknown_owner_falls_back_to_target_list() {
        // 查不到监听进程,但 target 列表含 doubao-chat → 按豆包处理
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, CHAT_LIST);
        let platform = FakePlatform {
            running: true,
            exe_found: true,
            ..Default::default()
        };
        let action = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap();
        assert_eq!(action, StartupAction::AlreadyConnected);

        // 查不到监听进程,且列表里没有豆包页面 → 端口占用,owning_process=unknown
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        spawn_mock_cdp_http(listener, NO_CHAT_LIST);
        let err = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_cdp_port_occupied");
        assert_eq!(err.details.unwrap()["owning_process"], "unknown");
    }

    #[tokio::test]
    async fn restarts_when_running_without_cdp() {
        let port = free_port();
        let platform = FakePlatform {
            running: true,
            exe_found: true,
            serve_on_launch: true,
            ..Default::default()
        };
        let action = ensure_doubao_started_with(&platform, port, Duration::from_secs(3))
            .await
            .unwrap();
        assert_eq!(action, StartupAction::Restarted);
        assert_eq!(platform.find_calls.load(Ordering::SeqCst), 1);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 1);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 1);
    }

    #[tokio::test]
    async fn starts_when_not_running() {
        let port = free_port();
        let platform = FakePlatform {
            running: false,
            exe_found: true,
            serve_on_launch: true,
            ..Default::default()
        };
        let action = ensure_doubao_started_with(&platform, port, Duration::from_secs(3))
            .await
            .unwrap();
        assert_eq!(action, StartupAction::Started);
        assert_eq!(platform.kill_calls.load(Ordering::SeqCst), 0);
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 1);
    }

    #[tokio::test]
    async fn launch_timeout_when_cdp_never_appears() {
        let port = free_port();
        let platform = FakePlatform {
            running: false,
            exe_found: true,
            serve_on_launch: false,
            ..Default::default()
        };
        let err = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_launch_timeout");
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 1);
    }

    #[tokio::test]
    async fn not_found_when_exe_missing() {
        let port = free_port();
        let platform = FakePlatform {
            running: false,
            exe_found: false,
            ..Default::default()
        };
        let err = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_not_found");
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test]
    async fn port_occupied_by_non_cdp_process_before_launch() {
        // 端口被占但不是 CDP 服务(/json/version 不通):不要去启动豆包,直接报占用
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let _keep = listener; // 只占着端口,不应答
        let platform = FakePlatform {
            running: false,
            listening_exe: Some(PathBuf::from("/usr/bin/nc")),
            exe_found: true,
            ..Default::default()
        };
        let err = ensure_doubao_started_with(&platform, port, Duration::from_millis(300))
            .await
            .unwrap_err();
        assert_eq!(err.code, "doubao_cdp_port_occupied");
        assert_eq!(err.details.unwrap()["owning_process"], "/usr/bin/nc");
        assert_eq!(platform.launch_calls.load(Ordering::SeqCst), 0);
    }

    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    async fn run_blocking_works_on_multi_thread_runtime() {
        assert_eq!(run_blocking(|| 41 + 1), 42);
    }

    #[tokio::test]
    async fn run_blocking_works_on_current_thread_runtime() {
        assert_eq!(run_blocking(|| "ok"), "ok");
    }

    /// 需要真豆包(端口 9707 已以调试模式运行):`cargo test live_ -- --ignored --nocapture`。
    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    #[ignore]
    async fn live_ensure_started() {
        let platform: Arc<dyn Platform> = crate::platform::create_platform();
        let action = ensure_doubao_started(&*platform, 9707).await.unwrap();
        println!("startup action: {action:?}");
        assert_eq!(action, StartupAction::AlreadyConnected);
    }
}
