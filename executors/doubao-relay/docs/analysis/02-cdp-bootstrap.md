# 02 · CDP 握手与启动豆包

## 为什么能操控豆包

**豆包桌面端是一个 Electron 应用**(字节自己也用 Electron 做 Doubao 桌面)。所有 Electron 应用在启动时如果加 `--remote-debugging-port=XXXX`,Chromium 内核会在那个端口开 **CDP (Chrome DevTools Protocol)** 服务,外部任何进程只要能连上,就能:

- 读/改 DOM(`DOM.*`)
- 执行任意 JavaScript(`Runtime.evaluate`)
- 监听所有网络请求、响应体、SSE 事件流(`Network.enable` + 事件)
- 模拟鼠标键盘(`Input.dispatchMouseEvent`, `Input.dispatchKeyEvent`)
- 拦截文件选择对话框(`Page.setInterceptFileChooserDialog`)
- 伪装焦点(`Emulation.setFocusEmulationEnabled`)
- 控制页面生命周期(`Page.setWebLifecycleState`)

**整个沈浪的操控能力,全部来自于让豆包在它不知情的情况下用这个参数启动起来**。用户以为"豆包好像自己重启了一下",其实是被套壳 App 杀掉重拉。

## 启动豆包的完整逻辑(doubao-accounts-background.ps1:134-177)

```powershell
function Ensure-DoubaoStarted {
  $running = @(Get-Process -Name Doubao -ErrorAction SilentlyContinue)

  # 情况 1: 已经跑着并且 9707 端口可连 → 直接返回
  if ($running.Count -and (Test-DoubaoConnection)) {
    $report.startup.action = 'already_connected'
    return
  }

  # 情况 2: 跑着但 CDP 端口没开 → 杀掉重启
  # 情况 3: 没跑 → 直接启动

  # 第 1 步: 找 Doubao.exe 的路径(三条来源)
  $paths = @()
  # 来源 A: 从已运行进程读 .Path
  foreach ($process in $running) { if ($process.Path) { $paths += $process.Path } }
  # 来源 B: 注册表 App Paths
  foreach ($key in @(
    'HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe',
    'HKLM:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe'
  )) {
    if (Test-Path $key) { $paths += (Get-Item $key).GetValue('') }
  }
  # 来源 C: 注册表 Uninstall 里找 DisplayName 匹配 '豆包|Doubao' 的
  foreach ($key in @(
    'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
    'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
    'HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'
  )) {
    foreach ($item in @(Get-ItemProperty $key -ErrorAction SilentlyContinue |
                        Where-Object { $_.DisplayName -match '豆包|Doubao' })) {
      if ($item.InstallLocation) {
        $paths += (Join-Path $item.InstallLocation 'Doubao.exe')
        $paths += (Join-Path $item.InstallLocation 'app\Doubao.exe')
      }
      if ($item.DisplayIcon) {
        $paths += ($item.DisplayIcon -replace ',\d+$','').Trim('"')
      }
    }
  }
  # 来源 D: 常见默认安装目录
  $paths += @(
    "$env:LOCALAPPDATA\Doubao\app\Doubao.exe",
    "$env:LOCALAPPDATA\Programs\Doubao\Doubao.exe",
    "$env:ProgramFiles\Doubao\Doubao.exe"
  )
  # 来源 E: 扫所有盘的根目录(有的用户装到 D:\Doubao\)
  foreach ($drive in @(Get-PSDrive -PSProvider FileSystem)) {
    $paths += (Join-Path $drive.Root 'Doubao\app\Doubao.exe')
  }

  $exe = @($paths | Where-Object {
    $_ -and (Split-Path $_ -Leaf) -ieq 'Doubao.exe' -and (Test-Path -LiteralPath $_ -PathType Leaf)
  } | Select-Object -Unique)

  if (!$exe.Count) { throw '未找到豆包安装程序' }

  # 第 2 步: 如果 Doubao 已在跑,先优雅关闭,再强制 kill
  if ($running.Count) {
    foreach ($process in $running) { if ($process.MainWindowHandle -ne 0) {
      [void]$process.CloseMainWindow()
    }}
    Start-Sleep -Milliseconds 800
    foreach ($process in @(Get-Process -Name Doubao -ErrorAction SilentlyContinue)) {
      Stop-Process -Id $process.Id -ErrorAction Stop
    }
    # 等退出
    $exitDeadline = [DateTime]::UtcNow.AddSeconds(5)
    while (@(Get-Process -Name Doubao).Count -and [DateTime]::UtcNow -lt $exitDeadline) {
      Start-Sleep -Milliseconds 200
    }
  }

  # 第 3 步: 用带调试端口的参数重启
  Start-Process -FilePath $exe[0] -ArgumentList @(
    '--remote-debugging-port=9707',
    '--remote-debugging-address=127.0.0.1',
    '--force-renderer-accessibility=complete'
  ) -ErrorAction Stop | Out-Null

  # 第 4 步: 轮询 9707 直到连通(最多 20 秒)
  $deadline = [DateTime]::UtcNow.AddSeconds(20)
  do {
    try {
      $targets = Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 1
      $pages = @($targets | Where-Object {
        try {
          $u = [uri]$_.url
          $_.type -eq 'page' -and $u.Scheme -in @('doubao','chrome') -and $u.Host -eq 'doubao-chat'
        } catch { $false }
      })
      if ($pages.Count -and (Test-DoubaoConnection)) {
        $report.startup.action = 'started'; return
      }
    } catch {}
    Start-Sleep -Milliseconds 250
  } while ([DateTime]::UtcNow -lt $deadline)

  throw '等待豆包启动超时'
}
```

### 启动参数解读

- `--remote-debugging-port=9707` → 开 CDP 服务。固定端口 9707 是沈浪约定的。
- `--remote-debugging-address=127.0.0.1` → 只监听本机 loopback,外部无法访问。
- `--force-renderer-accessibility=complete` → 强制开启完整的 Accessibility Tree,让后面的 UIA(UI Automation)也能读到元素。这是**双通道容错**的关键 —— 一条 DOM 路径,一条 UIA 路径,互为备份。

## 发现豆包页面(doubao-capture.ps1:100-117)

CDP 的 `GET /json/list` 返回所有可调试的"目标"(target):

```powershell
try {
  # 第 1 步: 枚举 9707 端口的监听进程,只认 Doubao.exe 自己开的
  $listeners = @(Get-NetTCPConnection -LocalPort 9707 -State Listen -ErrorAction SilentlyContinue)
  if (!$listeners.Count) { throw '本地接口未开启' }

  foreach ($l in $listeners) {
    if ($l.LocalAddress -notin @('127.0.0.1','::1') -or
        (Get-Process -Id $l.OwningProcess).ProcessName -ine 'Doubao') {
      throw '接口不属于本机豆包,已停止测试。'  # ← 防止被别的程序钓鱼
    }
  }

  # 第 2 步: HTTP GET 列表
  $targets = Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3

  # 第 3 步: 筛选关心的页面
  foreach ($target in $targets) {
    $u = [uri]$target.url
    if ($target.type -ne 'page' -or
        $u.Host -notin @('doubao-chat','doubao-background')) { continue }

    # 第 4 步: 校验 WebSocket 调试 URL 合法
    $ws = [uri]$target.webSocketDebuggerUrl
    if ($ws.Scheme -ne 'ws' -or
        $ws.Host -notin @('127.0.0.1','localhost','::1') -or
        $ws.Port -ne 9707) {
      throw '页面地址不符合本机连接范围。'
    }

    # 第 5 步: 建立 WebSocket 连接
    $socket = New-Object DoubaoCaptureSocket
    $socket.Connect($ws.AbsoluteUri)

    # 第 6 步: 第一条命令 - 开启网络事件推送
    $session.enableId = Command $session 'Network.enable' @{
      maxTotalBufferSize = 16777216      # 16 MB
      maxResourceBufferSize = 8388608    # 8 MB
    }
  }
}
```

### 豆包的两个页面类型

豆包桌面端在 Chromium 内部开了两个 BrowserWindow,对应两种 URL scheme:

| URL Host | 用途 |
|---|---|
| `doubao-chat` | 主聊天界面。用户看到的、生成任务点击的就是这个。|
| `doubao-background` | 后台任务页。可能是消息同步、通知推送、WebSocket 长连接的壳。|

工具同时连两个 → chat 页做操作,background 页做静默侦听(有些豆包的网络请求是从 background 发起的)。

### 安全校验链(很严,给 Rust 重写时参考)

1. `Get-NetTCPConnection -LocalPort 9707` 的监听进程必须是 `Doubao.exe`。
2. 监听地址必须是 `127.0.0.1` 或 `::1`,不能是 `0.0.0.0`。
3. `webSocketDebuggerUrl` 的 scheme 必须是 `ws`、host 必须是 loopback、port 必须是 9707。
4. 目标 `url` 的 scheme 必须是 `doubao` 或 `chrome`、host 必须是 `doubao-chat` 或 `doubao-background`。

**所有这些校验的目的**:防止一个恶意程序在 9707 端口上伪造服务钓鱼,或者防止用户同时装了另一个带 debug 端口的 Electron 应用(比如 VS Code)占了 9707 导致误连。

## 自定义 WebSocket 客户端(doubao-capture.ps1:7-47)

PowerShell 没有现成的 WebSocket 客户端,它内嵌了一段 C# 代码用 `Add-Type` 编译:

```csharp
public sealed class DoubaoCaptureSocket : IDisposable {
    public readonly ConcurrentQueue<string> Messages = new ConcurrentQueue<string>();
    private readonly ClientWebSocket socket = new ClientWebSocket();
    private readonly CancellationTokenSource cancel = new CancellationTokenSource();

    public void Connect(string url) {
        using (var timeout = new CancellationTokenSource(5000))
            socket.ConnectAsync(new Uri(url), timeout.Token).GetAwaiter().GetResult();
        Task.Run((Func<Task>)Read);         // 后台线程读消息
    }

    public void Send(string json) {
        var data = Encoding.UTF8.GetBytes(json);
        using (var timeout = new CancellationTokenSource(5000))
            socket.SendAsync(new ArraySegment<byte>(data),
                             WebSocketMessageType.Text, true, timeout.Token)
                  .GetAwaiter().GetResult();
    }

    private async Task Read() {
        try {
            byte[] buffer = new byte[16384];
            while (!cancel.IsCancellationRequested) {
                using (var stream = new MemoryStream()) {
                    WebSocketReceiveResult result;
                    do {
                        result = await socket.ReceiveAsync(
                            new ArraySegment<byte>(buffer), cancel.Token);
                        if (result.MessageType == WebSocketMessageType.Close) return;
                        stream.Write(buffer, 0, result.Count);
                        if (stream.Length > 16*1024*1024) throw new Exception("响应过大");
                    } while (!result.EndOfMessage);
                    if (Messages.Count > 5000) throw new Exception("事件队列过大");
                    Messages.Enqueue(Encoding.UTF8.GetString(stream.ToArray()));
                }
            }
        } catch (Exception) {
            Messages.Enqueue("{\"captureSocketClosed\":true}");  // 哨兵消息
        }
    }

    public void Dispose() {
        cancel.Cancel();
        socket.Dispose();
        cancel.Dispose();
    }
}
```

### 设计要点

- **异步读 + 同步 API**:`Connect`/`Send` 都用 `.GetAwaiter().GetResult()` 同步阻塞 5 秒超时,好让 PowerShell 代码可以线性写。
- **线程安全队列**:`ConcurrentQueue<string>` 从后台读线程生产,PS 主循环消费。
- **背压保护**:单条消息超 16 MB 或队列超 5000 条就扔异常,防止长时间不读导致内存爆。
- **哨兵消息**:断线时塞一条 `{"captureSocketClosed":true}` 到队列,主循环 `TryDequeue` 时就能感知到。

Rust 侧用 `tokio-tungstenite` 可以更简洁地做同样的事,甚至不需要队列(直接 `tokio::sync::mpsc` channel)。

## CDP 命令封装(doubao-capture.ps1:94-99)

```powershell
function Command($session, [string]$method, $params) {
    # 如果有并行追踪 key,在注入 JS 的表达式里替换标识符
    if ($trackingKey -and $method -eq 'Runtime.evaluate' -and $params.expression) {
        $params.expression = $params.expression.Replace('shenlangDoubaoGeneration', $trackingKey)
    }
    $session.nextId++
    $session.socket.Send((@{
        id = $session.nextId
        method = $method
        params = $params
    } | ConvertTo-Json -Depth 8 -Compress))
    return $session.nextId
}
```

CDP 的每个请求都是 `{id, method, params}`,响应用相同的 id 回复 `{id, result}` 或 `{id, error}`。

**trackingKey 技巧**:当多个画布任务并行时,每个任务会分配一个唯一的全局变量名 `shenlangDoubaoTask_abc123`,把注入 JS 里原本写死的 `shenlangDoubaoGeneration` 替换成这个,这样同一个豆包页面里能挂多个独立的生成 state 对象,互不干扰。

## 焦点伪装(关键于不抢用户焦点)

启动阶段还会设置两件事(`doubao-generate.ps1:127` 和 `doubao-accounts-background.ps1`):

```powershell
[void](Command 'Page.setWebLifecycleState' @{ state = 'active' })
[void](Command 'Emulation.setFocusEmulationEnabled' @{ enabled = $true })
```

### `Page.setWebLifecycleState state='active'`

Chromium 对后台 tab 会有节流(定时器减速、动画暂停、`requestAnimationFrame` 停)。把 lifecycle state 强制设为 `active` 可以绕过这个节流,让豆包页面即使被最小化也全速运行。

### `Emulation.setFocusEmulationEnabled enabled=$true`

更关键。这个 API 让豆包页面"误以为自己一直有焦点"。效果:
- `document.hasFocus()` 返回 true
- `document.visibilityState` 返回 'visible'
- `onfocus`/`onblur` 事件行为正常
- 一些检查焦点的 React 组件(比如输入框 autofocus)能正常工作

**对用户的意义**:即使你最小化豆包窗口、把焦点切到别的程序,沈浪的操作对豆包来说还是"用户在正常用"。这就是为什么它跑任务时不抢你的焦点 —— 它不需要抢。

## 启动阶段的 report 字段

`doubao-accounts-background.ps1` 的 `$report.startup` 会告诉你豆包是怎么起来的:

| action 值 | 含义 |
|---|---|
| `already_connected` | 豆包已在跑且 CDP 可连 |
| `locating_installation` | 正在找 Doubao.exe 路径 |
| `restarting` | 发现在跑但 CDP 没开,正在杀掉 |
| `starting` | 正在用调试参数启动 |
| `started` | 启动成功 |

error 字段可能的值:
- `installation_not_found` — 五条路径都没找到
- `port_occupied` — 9707 已被别的程序占
- `restart_failed` — 杀 Doubao.exe 失败(可能被 UAC 保护)
- `launch_failed` — `Start-Process` 异常
- `startup_timeout` — 20 秒内 CDP 还没就绪
- `connection_required` — 需要用户先手动开豆包并登录

## 给 Rust 重写的接口清单

Tauri 后端需要实现:

```rust
async fn ensure_doubao_started() -> Result<StartupReport>
async fn find_doubao_exe() -> Result<PathBuf>
async fn kill_doubao() -> Result<()>
async fn launch_doubao_with_debug(exe: &Path) -> Result<()>
async fn wait_for_cdp_ready(port: u16, timeout: Duration) -> Result<()>

async fn list_cdp_targets() -> Result<Vec<CdpTarget>>
async fn connect_target(target: &CdpTarget) -> Result<CdpSession>

// 每个 CdpSession 内部:
async fn send_cmd(&self, method: &str, params: Value) -> Result<Value>
fn on_event(&self, handler: impl Fn(CdpEvent))
```

推荐库:
- `tokio-tungstenite` — WebSocket
- `winreg` — 读 Windows 注册表
- `sysinfo` — 枚举进程
- 自己写个薄 CDP 协议封装(别用现成库,CDP 的 generated typings 太大)
