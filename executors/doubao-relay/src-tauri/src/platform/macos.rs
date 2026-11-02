//! macOS 平台适配。
//!
//! 真机事实(`docs/analysis/10-live-probe-macos.md`):豆包主进程是
//! `/Applications/Doubao.app/Contents/MacOS/Doubao`,但监听 9707 的是子进程
//! `Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser`,
//! 所以"是否豆包"按路径位于豆包 .app 包内判定(`Doubao.app` 不区分大小写,或 `*豆包*.app`),
//! 而不是进程名。"是否在运行"另外排除 launchd 托管的 XPC / crashpad 瞬态进程。
//!
//! 纯函数(路径判定 / `lsof` 解析)不带 cfg,所有平台都编译并测试;
//! 实际调用系统命令的 [`MacosPlatform`] 只在 macOS 编译。

use std::path::{Path, PathBuf};

/// 豆包 .app 包名。
pub const DOUBAO_BUNDLE_NAME: &str = "Doubao.app";

/// 路径分量是否是豆包的 .app 包:`Doubao.app`(macOS 文件系统默认不区分大小写)
/// 或含"豆包"且以 `.app` 结尾(与 `find_doubao_exe` 第 3 级 `/Applications/*豆包*.app` 一致)。
pub fn is_doubao_bundle_name(name: &str) -> bool {
    name.eq_ignore_ascii_case(DOUBAO_BUNDLE_NAME)
        || (name.contains("豆包") && name.to_ascii_lowercase().ends_with(".app"))
}

/// 该路径是否属于豆包安装:某个**非末尾**路径分量是豆包 .app 包(即路径位于包之内)。
pub fn is_doubao_exe_path(p: &Path) -> bool {
    let s = p.to_string_lossy();
    let mut parts = s.split('/').filter(|c| !c.is_empty()).peekable();
    while let Some(part) = parts.next() {
        if parts.peek().is_some() && is_doubao_bundle_name(part) {
            return true;
        }
    }
    false
}

/// 豆包的"瞬态辅助进程":`XPCServices/*.xpc` 服务与 `chrome_crashpad_handler`。
/// 它们由 launchd 直接托管(ppid=1),主进程退出后可能晚几秒才消失,
/// 判断"豆包是否还在运行 / 是否已退出"时要排除,只在发信号时一并处理。
pub fn is_doubao_transient_helper_path(p: &Path) -> bool {
    let s = p.to_string_lossy();
    if s.contains("/XPCServices/") {
        return true;
    }
    p.file_name()
        .map(|n| n.to_string_lossy().contains("crashpad"))
        .unwrap_or(false)
}

/// 解析 `lsof -Fp` 输出:每个 `p<pid>` 行一个 pid(去重、保序)。
pub fn parse_lsof_pids(output: &str) -> Vec<u32> {
    let mut pids = Vec::new();
    for line in output.lines() {
        let line = line.trim();
        if let Some(rest) = line.strip_prefix('p') {
            if let Ok(pid) = rest.trim().parse::<u32>() {
                if !pids.contains(&pid) {
                    pids.push(pid);
                }
            }
        }
    }
    pids
}

/// 从任意豆包进程的可执行路径回推 `Doubao.app` 包目录
/// (`.../Doubao.app/Contents/Helpers/Doubao Browser.app/...` → `.../Doubao.app`)。
pub fn app_bundle_from_exe(exe: &Path) -> Option<PathBuf> {
    // 取最外层命中(`Doubao.app/Contents/Helpers/豆包.app/...` 要回推到 `Doubao.app`)。
    exe.ancestors()
        .filter(|a| a != &exe)
        .filter(|a| {
            a.file_name()
                .map(|n| is_doubao_bundle_name(&n.to_string_lossy()))
                .unwrap_or(false)
        })
        .last()
        .map(Path::to_path_buf)
}

/// 包目录 → 主可执行文件 `Contents/MacOS/Doubao`。
pub fn main_exe_in_bundle(bundle: &Path) -> PathBuf {
    bundle.join("Contents").join("MacOS").join("Doubao")
}

#[cfg(target_os = "macos")]
pub use imp::MacosPlatform;

#[cfg(target_os = "macos")]
mod imp {
    use std::path::{Path, PathBuf};
    use std::process::Stdio;
    use std::time::Duration;

    use async_trait::async_trait;
    use sysinfo::{Pid, ProcessRefreshKind, ProcessesToUpdate, Signal, System, UpdateKind};

    use super::{
        app_bundle_from_exe, is_doubao_exe_path, is_doubao_transient_helper_path,
        main_exe_in_bundle, parse_lsof_pids, DOUBAO_BUNDLE_NAME,
    };
    use crate::platform::{join_error, Platform, PlatformError};

    const APP_DIR_NAME: &str = "doubao-relay";
    /// `osascript quit` 的最长等待:豆包主进程若对 AppleEvent 无响应,AppleScript 默认
    /// 要 1–2 分钟才超时,会拖垮"重连 20s 失败升级重启"的预算,所以自己掐。
    pub(super) const GRACEFUL_QUIT_TIMEOUT: Duration = Duration::from_secs(5);

    #[derive(Debug, Default, Clone, Copy)]
    pub struct MacosPlatform;

    impl MacosPlatform {
        pub fn new() -> Self {
            Self
        }
    }

    fn home_dir() -> PathBuf {
        dirs::home_dir().unwrap_or_else(|| PathBuf::from("/tmp"))
    }

    /// 枚举所有可执行路径属于豆包的进程(含 Doubao Browser、XPC、crashpad 等全部子进程)。
    fn doubao_processes() -> Vec<(Pid, PathBuf)> {
        let mut sys = System::new();
        sys.refresh_processes_specifics(
            ProcessesToUpdate::All,
            true,
            ProcessRefreshKind::nothing().with_exe(UpdateKind::Always),
        );
        sys.processes()
            .iter()
            .filter_map(|(pid, p)| p.exe().map(|e| (*pid, e.to_path_buf())))
            .filter(|(_, e)| is_doubao_exe_path(e))
            .collect()
    }

    /// "豆包是否在运行"的判据:排除 launchd 托管的 XPC / crashpad 瞬态进程,
    /// 否则主进程退出后它们晚几秒消失会让 `wait_until_gone` 误判并升级到 SIGTERM/SIGKILL。
    fn doubao_core_processes() -> Vec<(Pid, PathBuf)> {
        doubao_processes()
            .into_iter()
            .filter(|(_, e)| !is_doubao_transient_helper_path(e))
            .collect()
    }

    /// 跑一个外部命令并在 `timeout` 内等它退出;超时 / 无法执行返回 `None`。
    /// 子进程 `kill_on_drop`,超时后不会残留。
    pub(super) async fn status_with_timeout(
        mut cmd: tokio::process::Command,
        timeout: Duration,
    ) -> Option<std::process::ExitStatus> {
        cmd.stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .kill_on_drop(true);
        let mut child = cmd.spawn().ok()?;
        match tokio::time::timeout(timeout, child.wait()).await {
            Ok(Ok(status)) => Some(status),
            Ok(Err(_)) => None,
            Err(_elapsed) => {
                let _ = child.start_kill();
                None
            }
        }
    }

    fn exe_of_pid(pid: u32) -> Option<PathBuf> {
        let pid = Pid::from_u32(pid);
        let mut sys = System::new();
        sys.refresh_processes_specifics(
            ProcessesToUpdate::Some(&[pid]),
            true,
            ProcessRefreshKind::nothing().with_exe(UpdateKind::Always),
        );
        sys.process(pid)
            .and_then(|p| p.exe())
            .map(Path::to_path_buf)
    }

    /// 从运行中的豆包进程推出主可执行文件(存在才返回)。
    fn running_main_exe() -> Option<PathBuf> {
        doubao_processes()
            .into_iter()
            .filter_map(|(_, exe)| app_bundle_from_exe(&exe))
            .map(|b| main_exe_in_bundle(&b))
            .find(|p| p.is_file())
    }

    /// `/Applications/*豆包*.app/Contents/MacOS/*`。
    fn glob_doubao_cn_apps(apps_dir: &Path) -> Vec<PathBuf> {
        let Ok(entries) = std::fs::read_dir(apps_dir) else {
            return Vec::new();
        };
        let mut out = Vec::new();
        for entry in entries.flatten() {
            let name = entry.file_name().to_string_lossy().to_string();
            if !(name.contains("豆包") && name.ends_with(".app")) {
                continue;
            }
            let macos_dir = entry.path().join("Contents").join("MacOS");
            if let Ok(bins) = std::fs::read_dir(&macos_dir) {
                let mut bins: Vec<PathBuf> = bins
                    .flatten()
                    .map(|b| b.path())
                    .filter(|p| p.is_file())
                    .collect();
                bins.sort();
                out.extend(bins);
            }
        }
        out
    }

    fn mdfind_doubao_app() -> Vec<PathBuf> {
        let out = std::process::Command::new("mdfind")
            .arg(format!("kMDItemFSName == '{DOUBAO_BUNDLE_NAME}'"))
            .stderr(Stdio::null())
            .output();
        let Ok(out) = out else {
            return Vec::new();
        };
        String::from_utf8_lossy(&out.stdout)
            .lines()
            .map(str::trim)
            .filter(|l| !l.is_empty())
            .map(|l| main_exe_in_bundle(Path::new(l)))
            .collect()
    }

    /// 同步版发现逻辑,顺序照 spec 10。返回 (命中, 已搜索路径)。
    fn find_doubao_exe_blocking() -> Result<PathBuf, Vec<PathBuf>> {
        let mut searched: Vec<PathBuf> = Vec::new();
        if let Some(p) = running_main_exe() {
            return Ok(p);
        }
        let mut candidates = vec![main_exe_in_bundle(
            &Path::new("/Applications").join(DOUBAO_BUNDLE_NAME),
        )];
        candidates.extend(glob_doubao_cn_apps(Path::new("/Applications")));
        candidates.push(main_exe_in_bundle(
            &home_dir().join("Applications").join(DOUBAO_BUNDLE_NAME),
        ));
        for c in candidates {
            if c.is_file() {
                return Ok(c);
            }
            searched.push(c);
        }
        for c in mdfind_doubao_app() {
            if c.is_file() {
                return Ok(c);
            }
            searched.push(c);
        }
        Err(searched)
    }

    /// 启动时优先用的 `open -a` 目标:有包目录就给绝对路径,否则按名字。
    fn open_target(exe: &Path) -> String {
        app_bundle_from_exe(exe)
            .map(|b| b.to_string_lossy().to_string())
            .unwrap_or_else(|| "Doubao".to_string())
    }

    async fn wait_until_gone(timeout: Duration) -> bool {
        let deadline = tokio::time::Instant::now() + timeout;
        loop {
            let alive = tokio::task::spawn_blocking(doubao_core_processes)
                .await
                .map(|v| !v.is_empty())
                .unwrap_or(true);
            if !alive {
                return true;
            }
            if tokio::time::Instant::now() >= deadline {
                return false;
            }
            tokio::time::sleep(Duration::from_millis(250)).await;
        }
    }

    fn signal_all(signal: Signal) {
        let mut sys = System::new();
        sys.refresh_processes_specifics(
            ProcessesToUpdate::All,
            true,
            ProcessRefreshKind::nothing().with_exe(UpdateKind::Always),
        );
        for p in sys.processes().values() {
            if p.exe().map(is_doubao_exe_path).unwrap_or(false) {
                if signal == Signal::Kill {
                    p.kill();
                } else {
                    p.kill_with(signal);
                }
            }
        }
    }

    #[async_trait]
    impl Platform for MacosPlatform {
        async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
            let found = tokio::task::spawn_blocking(find_doubao_exe_blocking)
                .await
                .map_err(join_error)?;
            found.map_err(|searched| {
                let joined = searched
                    .iter()
                    .map(|p| p.to_string_lossy().to_string())
                    .collect::<Vec<_>>()
                    .join(", ");
                PlatformError::DoubaoNotFound(joined)
            })
        }

        async fn launch_doubao_with_cdp(&self, exe: &Path, port: u16) -> Result<(), PlatformError> {
            let port_arg = format!("--remote-debugging-port={port}");
            let addr_arg = "--remote-debugging-address=127.0.0.1";
            let target = open_target(exe);
            let open_result = tokio::process::Command::new("open")
                .args(["-a", &target, "--args", &port_arg, addr_arg])
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::piped())
                .output()
                .await;
            let open_err = match open_result {
                Ok(out) if out.status.success() => return Ok(()),
                Ok(out) => format!(
                    "open 退出码 {:?}:{}",
                    out.status.code(),
                    String::from_utf8_lossy(&out.stderr).trim()
                ),
                Err(e) => format!("open 无法执行:{e}"),
            };
            tracing::warn!(reason = %open_err, "open -a 启动豆包失败,改为直接执行二进制");
            std::process::Command::new(exe)
                .args([&port_arg, addr_arg])
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .spawn()
                .map(|_| ())
                .map_err(|e| {
                    PlatformError::LaunchFailed(format!(
                        "{open_err};直接执行 {} 也失败:{e}",
                        exe.display()
                    ))
                })
        }

        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            if !self.is_doubao_running().await {
                return Ok(());
            }
            let mut quit = tokio::process::Command::new("osascript");
            quit.args(["-e", "tell application \"Doubao\" to quit"]);
            match status_with_timeout(quit, GRACEFUL_QUIT_TIMEOUT).await {
                Some(status) if status.success() => {}
                Some(status) => {
                    tracing::warn!(code = ?status.code(), "osascript quit 退出码非 0");
                }
                None => {
                    tracing::warn!(
                        timeout_secs = GRACEFUL_QUIT_TIMEOUT.as_secs(),
                        "osascript quit 超时或无法执行,视为优雅退出失败"
                    );
                }
            }
            if wait_until_gone(Duration::from_secs(5)).await {
                return Ok(());
            }
            tracing::warn!("osascript quit 5 秒后豆包仍在运行,发送 SIGTERM");
            let _ = tokio::task::spawn_blocking(|| signal_all(Signal::Term)).await;
            if wait_until_gone(Duration::from_secs(1)).await {
                return Ok(());
            }
            tracing::warn!("SIGTERM 1 秒后豆包仍在运行,发送 SIGKILL");
            let _ = tokio::task::spawn_blocking(|| signal_all(Signal::Kill)).await;
            if wait_until_gone(Duration::from_secs(1)).await {
                Ok(())
            } else {
                Err(PlatformError::KillFailed(
                    "SIGKILL 后豆包进程仍存在".to_string(),
                ))
            }
        }

        async fn is_doubao_running(&self) -> bool {
            tokio::task::spawn_blocking(doubao_core_processes)
                .await
                .map(|v| !v.is_empty())
                .unwrap_or(false)
        }

        fn app_data_dir(&self) -> PathBuf {
            dirs::data_dir()
                .unwrap_or_else(|| home_dir().join("Library").join("Application Support"))
                .join(APP_DIR_NAME)
        }

        fn log_dir(&self) -> PathBuf {
            home_dir().join("Library").join("Logs").join(APP_DIR_NAME)
        }

        fn config_path(&self) -> PathBuf {
            self.app_data_dir().join("config.toml")
        }

        async fn reveal_in_file_manager(&self, path: &Path) -> Result<(), PlatformError> {
            let status = tokio::process::Command::new("open")
                .arg("-R")
                .arg(path)
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .status()
                .await?;
            if status.success() {
                Ok(())
            } else {
                Err(PlatformError::Io(std::io::Error::other(format!(
                    "open -R 退出码 {:?}",
                    status.code()
                ))))
            }
        }

        fn platform_name(&self) -> &'static str {
            "macos"
        }

        fn listening_process_exe(&self, port: u16) -> Option<PathBuf> {
            let out = std::process::Command::new("lsof")
                .args(["-nP", &format!("-iTCP:{port}"), "-sTCP:LISTEN", "-Fp"])
                .stdin(Stdio::null())
                .stderr(Stdio::null())
                .output()
                .ok()?;
            let pids = parse_lsof_pids(&String::from_utf8_lossy(&out.stdout));
            pids.into_iter().find_map(exe_of_pid)
        }

        fn is_doubao_exe_path(&self, p: &Path) -> bool {
            is_doubao_exe_path(p)
        }

        fn doubao_version(&self) -> Option<String> {
            let bundle = running_main_exe()
                .and_then(|e| app_bundle_from_exe(&e))
                .unwrap_or_else(|| Path::new("/Applications").join(DOUBAO_BUNDLE_NAME));
            let plist = bundle.join("Contents").join("Info.plist");
            if !plist.is_file() {
                return None;
            }
            let out = std::process::Command::new("defaults")
                .arg("read")
                .arg(&plist)
                .arg("CFBundleShortVersionString")
                .stdin(Stdio::null())
                .stderr(Stdio::null())
                .output()
                .ok()?;
            if !out.status.success() {
                return None;
            }
            let v = String::from_utf8_lossy(&out.stdout).trim().to_string();
            (!v.is_empty()).then_some(v)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn is_doubao_exe_path_accepts_bundle_children_and_rejects_others() {
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser"
        )));
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/Doubao.app/Contents/MacOS/Doubao"
        )));
        assert!(is_doubao_exe_path(Path::new(
            "/Users/x/Applications/Doubao.app/Contents/MacOS/Doubao"
        )));
        assert!(!is_doubao_exe_path(Path::new(
            "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
        )));
        assert!(!is_doubao_exe_path(Path::new("/usr/bin/nc")));
        // 只是名字里带 Doubao 不算
        assert!(!is_doubao_exe_path(Path::new("/opt/Doubao/bin/Doubao")));
        assert!(!is_doubao_exe_path(Path::new("/Applications/Doubao.app")));
        assert!(!is_doubao_exe_path(Path::new("/Applications/豆包.app")));
    }

    #[test]
    fn is_doubao_exe_path_is_case_insensitive_and_accepts_cn_bundle() {
        // macOS 默认文件系统不区分大小写
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/doubao.app/Contents/MacOS/Doubao"
        )));
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/DOUBAO.APP/Contents/MacOS/Doubao"
        )));
        // 与 find_doubao_exe 第 3 级 `/Applications/*豆包*.app/Contents/MacOS/*` 一致
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/豆包.app/Contents/MacOS/Doubao"
        )));
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/豆包桌面版.app/Contents/MacOS/豆包"
        )));
        // 真机:Doubao.app/Contents/Helpers/ 里还有一个 豆包.app
        assert!(is_doubao_exe_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/豆包.app/Contents/MacOS/豆包"
        )));
        assert!(!is_doubao_exe_path(Path::new("/Applications/豆包.txt/x")));
        assert!(!is_doubao_exe_path(Path::new("/Users/豆包/bin/tool")));
        assert!(is_doubao_bundle_name("Doubao.app"));
        assert!(is_doubao_bundle_name("豆包.app"));
        assert!(!is_doubao_bundle_name("Doubao"));
        assert!(!is_doubao_bundle_name("豆包"));
    }

    #[test]
    fn transient_helpers_are_xpc_and_crashpad_only() {
        assert!(is_doubao_transient_helper_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/XPCServices/AgentInfraService.xpc/Contents/MacOS/AgentInfraService"
        )));
        assert!(is_doubao_transient_helper_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/Frameworks/Doubao Browser Framework.framework/Versions/147.0.7727.149/Helpers/chrome_crashpad_handler"
        )));
        assert!(!is_doubao_transient_helper_path(Path::new(
            "/Applications/Doubao.app/Contents/MacOS/Doubao"
        )));
        assert!(!is_doubao_transient_helper_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser"
        )));
        assert!(!is_doubao_transient_helper_path(Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/Frameworks/Doubao Browser Framework.framework/Versions/147.0.7727.149/Helpers/Doubao Browser Helper (Renderer).app/Contents/MacOS/Doubao Browser Helper (Renderer)"
        )));
    }

    #[test]
    fn parse_lsof_pids_reads_p_lines_only() {
        assert_eq!(parse_lsof_pids("p74576\nf57\n"), vec![74576]);
        assert_eq!(
            parse_lsof_pids("p1\nf3\np2\nf4\np1\n"),
            vec![1, 2],
            "去重保序"
        );
        assert!(parse_lsof_pids("").is_empty());
        assert!(parse_lsof_pids("pabc\nf1\n").is_empty());
    }

    #[test]
    fn app_bundle_from_exe_walks_up_to_doubao_app() {
        let helper = Path::new(
            "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser",
        );
        assert_eq!(
            app_bundle_from_exe(helper),
            Some(PathBuf::from("/Applications/Doubao.app"))
        );
        assert_eq!(
            app_bundle_from_exe(Path::new("/Applications/Doubao.app/Contents/MacOS/Doubao")),
            Some(PathBuf::from("/Applications/Doubao.app"))
        );
        assert_eq!(
            app_bundle_from_exe(Path::new("/Applications/Chrome.app/Contents/MacOS/Chrome")),
            None
        );
        // 嵌套的 Helpers/豆包.app 要回推到最外层 Doubao.app
        assert_eq!(
            app_bundle_from_exe(Path::new(
                "/Applications/Doubao.app/Contents/Helpers/豆包.app/Contents/MacOS/豆包"
            )),
            Some(PathBuf::from("/Applications/Doubao.app"))
        );
        assert_eq!(
            app_bundle_from_exe(Path::new("/Applications/豆包.app/Contents/MacOS/豆包")),
            Some(PathBuf::from("/Applications/豆包.app"))
        );
        // 包目录本身不算"包内"
        assert_eq!(
            app_bundle_from_exe(Path::new("/Applications/Doubao.app")),
            None
        );
        assert_eq!(
            main_exe_in_bundle(Path::new("/Applications/Doubao.app")),
            PathBuf::from("/Applications/Doubao.app/Contents/MacOS/Doubao")
        );
    }

    #[cfg(target_os = "macos")]
    mod macos_only {
        use super::super::MacosPlatform;
        use crate::platform::Platform;
        use std::path::Path;

        #[test]
        fn directories_follow_macos_conventions() {
            let p = MacosPlatform::new();
            let data = p.app_data_dir();
            assert!(data.ends_with("doubao-relay"), "{data:?}");
            assert!(
                data.to_string_lossy()
                    .contains("/Library/Application Support/"),
                "{data:?}"
            );
            let logs = p.log_dir();
            assert!(logs.ends_with("doubao-relay"), "{logs:?}");
            assert!(
                logs.to_string_lossy().contains("/Library/Logs/"),
                "{logs:?}"
            );
            assert_eq!(p.config_path(), data.join("config.toml"));
            assert_eq!(p.platform_name(), "macos");
        }

        #[test]
        fn trait_is_doubao_exe_path_delegates() {
            let p = MacosPlatform::new();
            assert!(p.is_doubao_exe_path(Path::new(
                "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser"
            )));
            assert!(!p.is_doubao_exe_path(Path::new(
                "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
            )));
        }

        #[tokio::test]
        async fn status_with_timeout_kills_hung_command_within_budget() {
            use super::super::imp::{status_with_timeout, GRACEFUL_QUIT_TIMEOUT};
            use std::time::{Duration, Instant};

            // 模拟"豆包对 AppleEvent 无响应 → osascript 挂住":一个 60s 不退出的子进程
            let marker = format!("doubao-relay-test-{}", std::process::id());
            let mut cmd = tokio::process::Command::new("sh");
            cmd.args(["-c", &format!("exec sleep 60 # {marker}")]);
            let t0 = Instant::now();
            let got = status_with_timeout(cmd, Duration::from_millis(300)).await;
            assert_eq!(got, None, "超时应返回 None");
            assert!(
                t0.elapsed() < Duration::from_secs(3),
                "必须在超时预算内返回,实际 {:?}",
                t0.elapsed()
            );
            assert!(GRACEFUL_QUIT_TIMEOUT <= Duration::from_secs(5));

            // 子进程应被 kill,不残留
            tokio::time::sleep(Duration::from_millis(200)).await;
            let out = std::process::Command::new("pgrep")
                .args(["-f", &marker])
                .output()
                .unwrap();
            let leftover = String::from_utf8_lossy(&out.stdout);
            let leftover: Vec<&str> = leftover.lines().filter(|l| !l.trim().is_empty()).collect();
            assert!(leftover.is_empty(), "残留进程 {leftover:?}");
        }

        #[tokio::test]
        async fn status_with_timeout_returns_status_of_quick_command() {
            use super::super::imp::status_with_timeout;
            use std::time::Duration;

            let mut ok = tokio::process::Command::new("true");
            ok.arg("x");
            let s = status_with_timeout(ok, Duration::from_secs(5))
                .await
                .unwrap();
            assert!(s.success());

            let mut fail = tokio::process::Command::new("false");
            fail.arg("x");
            let s = status_with_timeout(fail, Duration::from_secs(5))
                .await
                .unwrap();
            assert!(!s.success());

            let missing = tokio::process::Command::new("/nonexistent/doubao-relay-no-such-binary");
            assert_eq!(
                status_with_timeout(missing, Duration::from_secs(1)).await,
                None
            );
        }

        #[test]
        fn listening_process_exe_on_unused_port_is_none() {
            // 绑定一个临时端口再释放,极大概率没人监听。
            let l = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
            let port = l.local_addr().unwrap().port();
            drop(l);
            assert_eq!(MacosPlatform::new().listening_process_exe(port), None);
        }

        #[test]
        fn listening_process_exe_finds_own_process() {
            let l = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
            let port = l.local_addr().unwrap().port();
            let exe = MacosPlatform::new().listening_process_exe(port);
            let me = std::env::current_exe().unwrap();
            assert_eq!(
                exe.as_deref().and_then(Path::file_name),
                me.file_name(),
                "{exe:?} vs {me:?}"
            );
            drop(l);
        }

        #[tokio::test]
        #[ignore]
        async fn live_find_doubao_exe() {
            let p = MacosPlatform::new();
            let exe = p.find_doubao_exe().await.unwrap();
            println!("find_doubao_exe = {}", exe.display());
            assert_eq!(
                exe,
                Path::new("/Applications/Doubao.app/Contents/MacOS/Doubao")
            );
            assert!(p.is_doubao_running().await);
        }

        #[test]
        #[ignore]
        fn live_listening_process_exe() {
            let p = MacosPlatform::new();
            let exe = p.listening_process_exe(9707).expect("9707 应有监听者");
            println!("listening_process_exe(9707) = {}", exe.display());
            assert!(exe.to_string_lossy().contains("Doubao.app"));
            assert!(p.is_doubao_exe_path(&exe));
        }

        #[test]
        #[ignore]
        fn live_doubao_version() {
            let v = MacosPlatform::new().doubao_version().expect("应读到版本");
            println!("doubao_version = {v}");
            assert!(v.split('.').count() >= 2, "{v}");
        }
    }
}
