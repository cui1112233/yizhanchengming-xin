//! 平台适配层:封装 Windows / macOS 差异,核心服务只依赖 [`Platform`] trait。
//!
//! 契约见 `docs/spec/10-platform-adapter.md`;计划 Task 1.5 额外加了三个方法
//! (`listening_process_exe` / `is_doubao_exe_path` / `doubao_version`),
//! 供 Task 1.6 `bootstrap` 判断 9707 端口监听者是否是豆包自己。

pub mod macos;
pub mod windows;

use std::path::{Path, PathBuf};
use std::sync::Arc;

use async_trait::async_trait;
use serde_json::json;

use crate::error::RelayError;

#[cfg(target_os = "macos")]
pub use macos::MacosPlatform;
#[cfg(windows)]
pub use windows::WindowsPlatform;

/// 平台层错误,经 `From<PlatformError> for RelayError` 映射到 11-error-codes。
#[derive(Debug, thiserror::Error)]
pub enum PlatformError {
    #[error("未找到豆包安装:{0}")]
    DoubaoNotFound(String),

    #[error("启动豆包失败:{0}")]
    LaunchFailed(String),

    #[error("杀掉豆包失败:{0}")]
    KillFailed(String),

    #[error("端口 {0} 被别的程序占用:{1}")]
    PortOccupied(u16, String),

    #[error("注册表读取失败:{0}")]
    RegistryError(String),

    #[error("IO 错误:{0}")]
    Io(#[from] std::io::Error),
}

#[async_trait]
pub trait Platform: Send + Sync {
    /// 发现 Doubao 可执行文件路径。
    async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError>;

    /// 启动豆包并带 CDP 调试端口。
    async fn launch_doubao_with_cdp(&self, exe: &Path, port: u16) -> Result<(), PlatformError>;

    /// 杀掉所有已运行的 Doubao 进程(优雅 + 强制)。
    async fn kill_doubao(&self) -> Result<(), PlatformError>;

    /// 检查 Doubao 是否在运行。
    async fn is_doubao_running(&self) -> bool;

    /// 应用数据目录。
    fn app_data_dir(&self) -> PathBuf;

    /// 日志目录。
    fn log_dir(&self) -> PathBuf;

    /// 配置文件路径。
    fn config_path(&self) -> PathBuf;

    /// 在文件管理器打开指定目录(Finder / Explorer)。
    async fn reveal_in_file_manager(&self, path: &Path) -> Result<(), PlatformError>;

    /// 平台标识(日志/诊断用)。
    fn platform_name(&self) -> &'static str;

    /// 返回监听 `port` 的进程可执行路径(无则 None)。
    /// macOS:`lsof -nP -iTCP:<port> -sTCP:LISTEN -Fp` 取 pid → sysinfo exe();
    /// Windows:`netstat -ano` 解析 → sysinfo。
    ///
    /// **阻塞调用**(同步 spawn 外部命令并等待,实测 `lsof` 约 45ms,冷机可能更久),
    /// 在 async 上下文里请用 `tokio::task::spawn_blocking` 包住。
    fn listening_process_exe(&self, port: u16) -> Option<PathBuf>;

    /// 该路径是否属于豆包安装(macOS:路径分量为 `Doubao.app`(不区分大小写)
    /// 或含"豆包"且以 `.app` 结尾的包,且路径位于该包之内;
    /// Windows:文件名 `Doubao.exe` 或路径含 `\Doubao\`)。
    fn is_doubao_exe_path(&self, p: &Path) -> bool;

    /// 豆包版本号。macOS 读 Info.plist `CFBundleShortVersionString`;Windows 返回 None。
    ///
    /// **阻塞调用**(macOS 同步执行 `defaults read`),async 上下文请用 `spawn_blocking`。
    fn doubao_version(&self) -> Option<String>;
}

/// 按编译目标实例化平台适配。
pub fn create_platform() -> Arc<dyn Platform> {
    #[cfg(windows)]
    {
        Arc::new(WindowsPlatform::new())
    }
    #[cfg(target_os = "macos")]
    {
        Arc::new(MacosPlatform::new())
    }
    #[cfg(not(any(windows, target_os = "macos")))]
    {
        panic!("不支持的平台")
    }
}

/// `spawn_blocking` 的 `JoinError`(阻塞任务 panic / 被取消)映射为 IO 错误,
/// 而不是塞进 `DoubaoNotFound` 的"已搜索路径"里。
pub(crate) fn join_error(e: tokio::task::JoinError) -> PlatformError {
    PlatformError::Io(std::io::Error::other(format!("进程枚举失败:{e}")))
}

/// 映射到 11-error-codes:
/// DoubaoNotFound→`doubao_not_found`,LaunchFailed→`doubao_launch_failed`,
/// PortOccupied→`doubao_cdp_port_occupied`,RegistryError→`platform_registry_error`,
/// KillFailed→`platform_process_enum_failed`,Io→`system_io_error`。
impl From<PlatformError> for RelayError {
    fn from(e: PlatformError) -> Self {
        match e {
            PlatformError::DoubaoNotFound(searched) => RelayError::new(
                "doubao_not_found",
                "未找到豆包安装,请先安装豆包桌面端后点击重连豆包",
            )
            .with_details(json!({ "searched_paths": split_searched(&searched) })),
            PlatformError::LaunchFailed(reason) => RelayError::new(
                "doubao_launch_failed",
                "启动豆包失败,请手动打开豆包后点击重连豆包",
            )
            .with_details(json!({ "reason": reason })),
            PlatformError::KillFailed(reason) => RelayError::new(
                "platform_process_enum_failed",
                "无法结束豆包进程,请手动退出豆包后重试",
            )
            .with_details(json!({ "reason": reason })),
            PlatformError::PortOccupied(port, owner) => RelayError::new(
                "doubao_cdp_port_occupied",
                format!("端口 {port} 被别的程序占用,请关闭该程序或在设置里更换 CDP 端口"),
            )
            .with_details(json!({ "owning_process": owner, "port": port })),
            PlatformError::RegistryError(reason) => RelayError::new(
                "platform_registry_error",
                "读取注册表失败,请以当前用户重新安装豆包后重试",
            )
            .with_details(json!({ "reason": reason })),
            PlatformError::Io(io) => RelayError::from(io),
        }
    }
}

/// `DoubaoNotFound` 的消息是 `", "` 连接的候选路径列表,拆回数组放 details。
fn split_searched(s: &str) -> Vec<String> {
    s.split(", ")
        .map(str::trim)
        .filter(|p| !p.is_empty())
        .map(str::to_string)
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn platform_error_maps_to_relay_error_codes() {
        let e: RelayError = PlatformError::DoubaoNotFound("/a, /b".into()).into();
        assert_eq!(e.code, "doubao_not_found");
        assert_eq!(e.http_status(), 503);
        // 建议必须指向真实存在的入口(没有"指定豆包路径"的设置项)
        assert_eq!(e.message, "未找到豆包安装,请先安装豆包桌面端后点击重连豆包");
        assert_eq!(e.details.unwrap()["searched_paths"], json!(["/a", "/b"]));

        let e: RelayError = PlatformError::LaunchFailed("open: exit 1".into()).into();
        assert_eq!(e.code, "doubao_launch_failed");
        assert_eq!(e.details.unwrap()["reason"], "open: exit 1");

        let e: RelayError = PlatformError::KillFailed("still alive".into()).into();
        assert_eq!(e.code, "platform_process_enum_failed");
        assert_eq!(e.http_status(), 500);

        let e: RelayError = PlatformError::PortOccupied(9707, "/usr/bin/nc".into()).into();
        assert_eq!(e.code, "doubao_cdp_port_occupied");
        let d = e.details.unwrap();
        assert_eq!(d["owning_process"], "/usr/bin/nc");
        assert_eq!(d["port"], 9707);

        let e: RelayError = PlatformError::RegistryError("denied".into()).into();
        assert_eq!(e.code, "platform_registry_error");

        let io = std::io::Error::new(std::io::ErrorKind::PermissionDenied, "x");
        let e: RelayError = PlatformError::Io(io).into();
        assert_eq!(e.code, "system_io_error");
    }

    #[tokio::test]
    async fn join_error_maps_to_system_io_error_not_searched_paths() {
        let err = tokio::task::spawn_blocking(|| -> () { panic!("boom") })
            .await
            .unwrap_err();
        let pe = join_error(err);
        assert!(matches!(pe, PlatformError::Io(_)), "{pe:?}");
        let e: RelayError = pe.into();
        assert_eq!(e.code, "system_io_error");
        assert!(e
            .details
            .as_ref()
            .map(|d| d.get("searched_paths").is_none())
            .unwrap_or(true));
    }

    #[test]
    fn platform_error_display_is_chinese() {
        assert_eq!(
            PlatformError::PortOccupied(9707, "nc".into()).to_string(),
            "端口 9707 被别的程序占用:nc"
        );
        assert_eq!(
            PlatformError::DoubaoNotFound("x".into()).to_string(),
            "未找到豆包安装:x"
        );
    }

    #[test]
    fn create_platform_matches_current_os() {
        let p = create_platform();
        assert_eq!(p.platform_name(), std::env::consts::OS);
        assert!(p.app_data_dir().ends_with("doubao-relay"));
        assert!(p.config_path().ends_with("config.toml"));
        assert!(p.config_path().starts_with(p.app_data_dir()));
        assert!(p.log_dir().ends_with("doubao-relay") || p.log_dir().ends_with("logs"));
    }
}
