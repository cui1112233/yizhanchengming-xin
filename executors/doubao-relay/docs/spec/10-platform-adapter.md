# 10 · 平台适配层

## 职责

**封装 Windows 和 macOS 的差异 API**,让核心服务只依赖 trait,不用 `#[cfg(target_os=...)]`。

## trait 定义

```rust
#[async_trait]
pub trait Platform: Send + Sync {
    /// 发现 Doubao 可执行文件路径
    async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError>;

    /// 启动豆包并带 CDP 调试端口
    async fn launch_doubao_with_cdp(
        &self,
        exe: &Path,
        port: u16,
    ) -> Result<(), PlatformError>;

    /// 杀掉所有已运行的 Doubao 进程(优雅 + 强制)
    async fn kill_doubao(&self) -> Result<(), PlatformError>;

    /// 检查 Doubao 是否在运行
    async fn is_doubao_running(&self) -> bool;

    /// 应用数据目录
    fn app_data_dir(&self) -> PathBuf;

    /// 日志目录
    fn log_dir(&self) -> PathBuf;

    /// 配置文件路径
    fn config_path(&self) -> PathBuf;

    /// 在文件管理器打开指定目录(Finder / Explorer)
    async fn reveal_in_file_manager(&self, path: &Path) -> Result<(), PlatformError>;

    /// 平台标识(日志/诊断用)
    fn platform_name(&self) -> &'static str;
}
```

## 错误类型

```rust
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
```

## Windows 实现

### `find_doubao_exe` 的搜索来源

按优先级依次尝试,返回第一个真实存在的:

1. 已运行的 Doubao 进程的 `.Path`(用 `sysinfo` 枚举)
2. 注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe`
3. 注册表 `HKLM\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe`
4. 注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\*` + DisplayName 匹配 `豆包|Doubao` → 读 InstallLocation
5. 同上但 `HKLM` 版本
6. 同上但 `HKLM\Software\WOW6432Node\...`
7. `%LOCALAPPDATA%\Doubao\app\Doubao.exe`
8. `%LOCALAPPDATA%\Programs\Doubao\Doubao.exe`
9. `%ProgramFiles%\Doubao\Doubao.exe`
10. 扫所有盘根 `<drive>\Doubao\app\Doubao.exe`

### `launch_doubao_with_cdp`

```rust
std::process::Command::new(exe)
    .args(&[
        "--remote-debugging-port=9707",
        "--remote-debugging-address=127.0.0.1",
        "--force-renderer-accessibility=complete",
    ])
    .creation_flags(DETACHED_PROCESS | CREATE_NEW_PROCESS_GROUP)
    .spawn()?;
```

**注意**:用 `DETACHED_PROCESS` 让豆包脱离 doubao-relay 进程树,避免 relay 退出时豆包也被带走。

### `kill_doubao`

```rust
// 第 1 轮:优雅
for process in enum_doubao_processes() {
    windows_sys::Win32::UI::WindowsAndMessaging::PostMessageW(hwnd, WM_CLOSE, 0, 0);
}
tokio::time::sleep(Duration::from_millis(800)).await;

// 第 2 轮:强制
for process in enum_doubao_processes() {
    windows_sys::Win32::System::Threading::TerminateProcess(handle, 1);
}
```

### `app_data_dir`

`%APPDATA%\doubao-relay\`(即 `C:\Users\<name>\AppData\Roaming\doubao-relay\`)

**优先用** `tauri::api::path::app_data_dir()`,保证和 Tauri 自动管理一致。

## macOS 实现

### `find_doubao_exe` 的搜索来源

1. 已运行的 Doubao 进程(用 `sysinfo`)
2. `/Applications/Doubao.app/Contents/MacOS/Doubao`
3. `/Applications/*豆包*.app/Contents/MacOS/*`(通配)
4. `~/Applications/Doubao.app/Contents/MacOS/Doubao`
5. `mdfind "kMDItemFSName == 'Doubao.app'"`(Spotlight 全盘)

### `launch_doubao_with_cdp`

**两种方式**,优先用 `open -a`,原因是它能走 macOS 的 LSUIElement 和 Launch Services:

```rust
std::process::Command::new("open")
    .args(&[
        "-a", "Doubao",
        "--args",
        "--remote-debugging-port=9707",
        "--remote-debugging-address=127.0.0.1",
    ])
    .spawn()?;
```

**fallback**:如果 open 失败,直接执行:

```rust
std::process::Command::new(exe_path)  // "/Applications/Doubao.app/Contents/MacOS/Doubao"
    .args(&[
        "--remote-debugging-port=9707",
        "--remote-debugging-address=127.0.0.1",
    ])
    .spawn()?;
```

### `kill_doubao`

```bash
# 优雅
osascript -e 'tell application "Doubao" to quit'
# 等 1 秒
# 强制
pkill -x Doubao
```

或者用 `sysinfo` 枚举 + libc::kill(pid, SIGTERM)。

### `app_data_dir`

`~/Library/Application Support/doubao-relay/`

### `log_dir`

`~/Library/Logs/doubao-relay/`(macOS 日志惯例目录)

### `reveal_in_file_manager`

```bash
open -R <path>
```

Windows 等价:
```powershell
explorer /select,<path>
```

## 实例化

```rust
pub fn create_platform() -> Arc<dyn Platform> {
    #[cfg(target_os = "windows")]
    {
        Arc::new(WindowsPlatform::new())
    }
    #[cfg(target_os = "macos")]
    {
        Arc::new(MacosPlatform::new())
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos")))]
    {
        panic!("不支持的平台")
    }
}
```

## 差异速查表

| 功能 | Windows | macOS |
|---|---|---|
| 发现 exe | 注册表 + 文件扫描 | /Applications 扫描 + Spotlight |
| 启动命令 | `.exe` 直接执行 | `open -a` 或 `.app/Contents/MacOS/*` |
| 杀进程 | PostMessage WM_CLOSE + TerminateProcess | osascript quit + pkill |
| 用户数据目录 | `%APPDATA%\doubao-relay\` | `~/Library/Application Support/doubao-relay/` |
| 日志目录 | 同用户数据下的 `logs/` | `~/Library/Logs/doubao-relay/` |
| 配置目录 | 同用户数据 | 同用户数据 |
| 文件管理器 reveal | `explorer /select,<path>` | `open -R <path>` |
| 开机自启 | 注册表 Run key 或 Startup 文件夹 | LaunchAgent plist |
| 原生对话框 | #32770 Win32(UIA) | Cocoa NSOpenPanel(通常 CDP 能拦截,不需要 fallback) |

## 平台特定依赖

### Windows 依赖

```toml
[target.'cfg(windows)'.dependencies]
winreg = "0.52"
windows-sys = { version = "0.52", features = [
    "Win32_System_Threading",
    "Win32_UI_WindowsAndMessaging",
    "Win32_Foundation",
] }
```

### macOS 依赖

通常只需要 `std::process::Command` + `libc` + `tokio`,**无需额外 crate**。
特殊情况(将来)可能用:

```toml
[target.'cfg(target_os = "macos")'.dependencies]
objc = "0.2"       # 如果需要 NSWorkspace API
core-foundation = "0.9"
```

## UI 兼容性自检

每次启动 doubao-relay 时,启动完成后 3 秒跑一次自检:

```rust
pub async fn ui_compat_check(session: &CdpSession) -> UiCompatResult {
    let probe = r#"({
        hasInput: !!document.querySelector('[contenteditable="true"]'),
        hasAccountButton: !!document.querySelector('button[aria-haspopup="menu"]'),
        hasAnyImg: document.querySelectorAll('img').length > 0,
        visibility: document.visibilityState,
        bodyLen: document.body.innerText.length
    })"#;
    let result: UiProbeResult = evaluate_js(session, probe, Duration::from_secs(10)).await?;
    UiCompatResult {
        ok: result.has_input && result.has_account_button && result.body_len > 100,
        details: result,
    }
}
```

结果写入 metadata 表并通过 `doubao_status_changed` 事件推 UI。
