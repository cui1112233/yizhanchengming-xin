//! Windows 平台适配。开发机是 macOS,本实现只能 `cargo check --target x86_64-pc-windows-msvc`;
//! 纯函数(路径判定 / `netstat -ano` 解析)不带 cfg,所有平台都编译并测试。
//!
//! 与 spec 10 的差异(计划 Task 1.5 规定):`kill_doubao` 用 `taskkill /IM Doubao.exe` →
//! `taskkill /F /IM Doubao.exe`,不用 `PostMessage(WM_CLOSE)` + `TerminateProcess`,
//! 因此不需要 `windows-sys`。

use std::path::Path;

/// 豆包主程序文件名。
pub const DOUBAO_EXE_NAME: &str = "Doubao.exe";

fn doubao_launch_args(port: u16) -> Vec<String> {
    vec![
        format!("--remote-debugging-port={port}"),
        "--remote-debugging-address=127.0.0.1".into(),
        "--force-renderer-accessibility=complete".into(),
        "--disable-backgrounding-occluded-windows".into(),
        "--disable-renderer-backgrounding".into(),
        "--disable-background-timer-throttling".into(),
    ]
}

/// 该路径是否属于豆包安装:文件名 `Doubao.exe`(忽略大小写)或路径含 `\Doubao\`。
///
/// 按字符串处理而不是 `Path::file_name`,这样在 macOS 上跑测试时反斜杠路径也能正确判定。
pub fn is_doubao_exe_path(p: &Path) -> bool {
    let s = p.to_string_lossy();
    let last = s.rsplit(['\\', '/']).next().unwrap_or("");
    if last.eq_ignore_ascii_case(DOUBAO_EXE_NAME) {
        return true;
    }
    s.to_ascii_lowercase().contains("\\doubao\\")
}

/// 解析 `netstat -ano` 输出,返回监听 `port` 的 TCP 进程 pid(IPv4 / IPv6 都认)。
///
/// 行形如 `  TCP    127.0.0.1:9707    0.0.0.0:0    LISTENING    1234`。
pub fn parse_netstat_listening_pid(output: &str, port: u16) -> Option<u32> {
    let suffix = format!(":{port}");
    for line in output.lines() {
        let fields: Vec<&str> = line.split_whitespace().collect();
        if fields.len() < 5 {
            continue;
        }
        if !fields[0].eq_ignore_ascii_case("TCP") {
            continue;
        }
        if !fields[1].ends_with(&suffix) {
            continue;
        }
        if !fields[3].eq_ignore_ascii_case("LISTENING") {
            continue;
        }
        if let Ok(pid) = fields[4].parse::<u32>() {
            return Some(pid);
        }
    }
    None
}

#[cfg(windows)]
pub use imp::WindowsPlatform;

#[cfg(windows)]
mod imp {
    use std::os::windows::process::CommandExt;
    use std::path::{Path, PathBuf};
    use std::process::Stdio;
    use std::time::Duration;

    use async_trait::async_trait;
    use sysinfo::{Pid, ProcessRefreshKind, ProcessesToUpdate, System, UpdateKind};
    use winreg::enums::{HKEY_CURRENT_USER, HKEY_LOCAL_MACHINE};
    use winreg::RegKey;

    use super::{
        doubao_launch_args, is_doubao_exe_path, parse_netstat_listening_pid, DOUBAO_EXE_NAME,
    };
    use crate::platform::{join_error, Platform, PlatformError};

    const APP_DIR_NAME: &str = "doubao-relay";
    const DETACHED_PROCESS: u32 = 0x0000_0008;
    const CREATE_NEW_PROCESS_GROUP: u32 = 0x0000_0200;
    const APP_PATHS_KEY: &str = r"Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe";
    const UNINSTALL_KEYS: &[(winreg::HKEY, &str)] = &[
        (
            HKEY_CURRENT_USER,
            r"Software\Microsoft\Windows\CurrentVersion\Uninstall",
        ),
        (
            HKEY_LOCAL_MACHINE,
            r"Software\Microsoft\Windows\CurrentVersion\Uninstall",
        ),
        (
            HKEY_LOCAL_MACHINE,
            r"Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall",
        ),
    ];

    #[derive(Debug, Default, Clone, Copy)]
    pub struct WindowsPlatform;

    impl WindowsPlatform {
        pub fn new() -> Self {
            Self
        }
    }

    fn is_doubao_process_name(name: &std::ffi::OsStr) -> bool {
        name.to_string_lossy().eq_ignore_ascii_case(DOUBAO_EXE_NAME)
    }

    /// 枚举豆包主进程(按进程名 `Doubao.exe`)。
    fn doubao_processes() -> Vec<(Pid, Option<PathBuf>)> {
        let mut sys = System::new();
        sys.refresh_processes_specifics(
            ProcessesToUpdate::All,
            true,
            ProcessRefreshKind::nothing().with_exe(UpdateKind::Always),
        );
        sys.processes()
            .iter()
            .filter(|(_, p)| is_doubao_process_name(p.name()))
            .map(|(pid, p)| (*pid, p.exe().map(Path::to_path_buf)))
            .collect()
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

    fn app_paths_value(root: winreg::HKEY) -> Option<PathBuf> {
        let key = RegKey::predef(root).open_subkey(APP_PATHS_KEY).ok()?;
        let v: String = key.get_value("").ok()?;
        let v = v.trim().trim_matches('"').to_string();
        (!v.is_empty()).then(|| PathBuf::from(v))
    }

    /// Uninstall 子键里 DisplayName 匹配 `豆包|Doubao` 的 InstallLocation。
    fn uninstall_locations(root: winreg::HKEY, path: &str) -> Vec<PathBuf> {
        let Ok(key) = RegKey::predef(root).open_subkey(path) else {
            return Vec::new();
        };
        let mut out = Vec::new();
        for sub in key.enum_keys().flatten() {
            let Ok(k) = key.open_subkey(&sub) else {
                continue;
            };
            let name: String = k.get_value("DisplayName").unwrap_or_default();
            if !(name.contains("豆包") || name.to_ascii_lowercase().contains("doubao")) {
                continue;
            }
            let loc: String = k.get_value("InstallLocation").unwrap_or_default();
            let loc = loc.trim().trim_matches('"');
            if loc.is_empty() {
                continue;
            }
            let base = PathBuf::from(loc);
            out.push(base.join(DOUBAO_EXE_NAME));
            out.push(base.join("app").join(DOUBAO_EXE_NAME));
        }
        out
    }

    fn env_path(var: &str) -> Option<PathBuf> {
        std::env::var_os(var).map(PathBuf::from)
    }

    fn find_doubao_exe_blocking() -> Result<PathBuf, Vec<PathBuf>> {
        let mut searched: Vec<PathBuf> = Vec::new();
        // 1) 已运行进程
        for (_, exe) in doubao_processes() {
            if let Some(exe) = exe {
                if exe.is_file() {
                    return Ok(exe);
                }
            }
        }
        let mut candidates: Vec<PathBuf> = Vec::new();
        // 2–3) App Paths
        candidates.extend(app_paths_value(HKEY_CURRENT_USER));
        candidates.extend(app_paths_value(HKEY_LOCAL_MACHINE));
        // 4–6) Uninstall
        for (root, path) in UNINSTALL_KEYS {
            candidates.extend(uninstall_locations(*root, path));
        }
        // 7–9) 固定目录
        if let Some(local) = env_path("LOCALAPPDATA") {
            candidates.push(local.join("Doubao").join("app").join(DOUBAO_EXE_NAME));
            candidates.push(local.join("Programs").join("Doubao").join(DOUBAO_EXE_NAME));
        }
        if let Some(pf) = env_path("ProgramFiles") {
            candidates.push(pf.join("Doubao").join(DOUBAO_EXE_NAME));
        }
        // 10) 所有盘根
        for d in b'C'..=b'Z' {
            candidates.push(PathBuf::from(format!(
                "{}:\\Doubao\\app\\{DOUBAO_EXE_NAME}",
                d as char
            )));
        }
        for c in candidates {
            if c.is_file() {
                return Ok(c);
            }
            searched.push(c);
        }
        Err(searched)
    }

    async fn wait_until_gone(timeout: Duration) -> bool {
        let deadline = tokio::time::Instant::now() + timeout;
        loop {
            let alive = tokio::task::spawn_blocking(doubao_processes)
                .await
                .map(|v| !v.is_empty())
                .unwrap_or(true);
            if !alive {
                return true;
            }
            if tokio::time::Instant::now() >= deadline {
                return false;
            }
            tokio::time::sleep(Duration::from_millis(200)).await;
        }
    }

    async fn taskkill(force: bool) {
        let mut cmd = tokio::process::Command::new("taskkill");
        if force {
            cmd.arg("/F");
        }
        cmd.args(["/IM", DOUBAO_EXE_NAME]);
        if force {
            cmd.arg("/T");
        }
        let _ = cmd
            .stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .await;
    }

    #[async_trait]
    impl Platform for WindowsPlatform {
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
            std::process::Command::new(exe)
                .args(doubao_launch_args(port))
                .creation_flags(DETACHED_PROCESS | CREATE_NEW_PROCESS_GROUP)
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .spawn()
                .map(|_| ())
                .map_err(|e| PlatformError::LaunchFailed(format!("{}:{e}", exe.display())))
        }

        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            if !self.is_doubao_running().await {
                return Ok(());
            }
            taskkill(false).await;
            if wait_until_gone(Duration::from_millis(800)).await {
                return Ok(());
            }
            tracing::warn!("taskkill 后豆包仍在运行,改为强制结束");
            taskkill(true).await;
            if wait_until_gone(Duration::from_secs(2)).await {
                Ok(())
            } else {
                Err(PlatformError::KillFailed(
                    "taskkill /F 后豆包进程仍存在".to_string(),
                ))
            }
        }

        async fn is_doubao_running(&self) -> bool {
            tokio::task::spawn_blocking(doubao_processes)
                .await
                .map(|v| !v.is_empty())
                .unwrap_or(false)
        }

        fn app_data_dir(&self) -> PathBuf {
            dirs::config_dir()
                .or_else(|| env_path("APPDATA"))
                .unwrap_or_else(|| PathBuf::from("."))
                .join(APP_DIR_NAME)
        }

        fn log_dir(&self) -> PathBuf {
            self.app_data_dir().join("logs")
        }

        fn config_path(&self) -> PathBuf {
            self.app_data_dir().join("config.toml")
        }

        async fn reveal_in_file_manager(&self, path: &Path) -> Result<(), PlatformError> {
            // explorer 的退出码不可靠(常返回 1),只要能 spawn 就算成功。
            tokio::process::Command::new("explorer")
                .arg(format!("/select,{}", path.display()))
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .spawn()
                .map(|_| ())
                .map_err(PlatformError::Io)
        }

        fn platform_name(&self) -> &'static str {
            "windows"
        }

        fn listening_process_exe(&self, port: u16) -> Option<PathBuf> {
            // 不加 `-p TCP`:那样只列 IPv4,`[::1]:<port>` 的 TCPv6 监听会漏掉。
            let out = std::process::Command::new("netstat")
                .arg("-ano")
                .stdin(Stdio::null())
                .stderr(Stdio::null())
                .creation_flags(0x0800_0000) // CREATE_NO_WINDOW
                .output()
                .ok()?;
            let pid = parse_netstat_listening_pid(&String::from_utf8_lossy(&out.stdout), port)?;
            exe_of_pid(pid)
        }

        fn is_doubao_exe_path(&self, p: &Path) -> bool {
            is_doubao_exe_path(p)
        }

        fn doubao_version(&self) -> Option<String> {
            None
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn doubao_launch_keeps_occluded_chat_page_rendering() {
        let args = doubao_launch_args(9707);
        assert!(args
            .iter()
            .any(|arg| arg == "--disable-backgrounding-occluded-windows"));
        assert!(args
            .iter()
            .any(|arg| arg == "--disable-renderer-backgrounding"));
        assert!(args
            .iter()
            .any(|arg| arg == "--disable-background-timer-throttling"));
    }

    #[test]
    fn is_doubao_exe_path_windows_rules() {
        assert!(is_doubao_exe_path(Path::new(
            r"C:\Users\me\AppData\Local\Doubao\app\Doubao.exe"
        )));
        assert!(is_doubao_exe_path(Path::new(r"D:\Doubao\app\doubao.exe")));
        assert!(is_doubao_exe_path(Path::new(
            r"C:\Program Files\Doubao\resources\helper.exe"
        )));
        assert!(!is_doubao_exe_path(Path::new(
            r"C:\Program Files\Google\Chrome\Application\chrome.exe"
        )));
        assert!(!is_doubao_exe_path(Path::new(
            r"C:\Windows\System32\svchost.exe"
        )));
        assert!(!is_doubao_exe_path(Path::new(
            r"C:\Tools\DoubaoRelay\relay.exe"
        )));
    }

    #[test]
    fn parse_netstat_finds_listening_pid_for_port() {
        let out = "\n\
Active Connections\n\
\n\
  Proto  Local Address          Foreign Address        State           PID\n\
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1092\n\
  TCP    127.0.0.1:9707         0.0.0.0:0              LISTENING       4321\n\
  TCP    127.0.0.1:9707         127.0.0.1:52001        ESTABLISHED     4321\n\
  TCP    127.0.0.1:52001        127.0.0.1:9707         ESTABLISHED     9999\n\
  TCP    [::1]:8787             [::]:0                 LISTENING       777\n\
  UDP    0.0.0.0:9707           *:*                                    5555\n";
        assert_eq!(parse_netstat_listening_pid(out, 9707), Some(4321));
        assert_eq!(parse_netstat_listening_pid(out, 8787), Some(777));
        assert_eq!(parse_netstat_listening_pid(out, 135), Some(1092));
        assert_eq!(
            parse_netstat_listening_pid(out, 52001),
            None,
            "非 LISTENING 不算"
        );
        assert_eq!(
            parse_netstat_listening_pid(out, 970),
            None,
            "端口必须整体匹配"
        );
        assert_eq!(parse_netstat_listening_pid("", 9707), None);
    }

    #[cfg(windows)]
    mod windows_only {
        use super::super::WindowsPlatform;
        use crate::platform::Platform;

        #[test]
        fn directories_follow_windows_conventions() {
            let p = WindowsPlatform::new();
            let data = p.app_data_dir();
            assert!(data.ends_with("doubao-relay"), "{data:?}");
            assert_eq!(p.log_dir(), data.join("logs"));
            assert_eq!(p.config_path(), data.join("config.toml"));
            assert_eq!(p.platform_name(), "windows");
            assert_eq!(p.doubao_version(), None);
        }
    }
}
