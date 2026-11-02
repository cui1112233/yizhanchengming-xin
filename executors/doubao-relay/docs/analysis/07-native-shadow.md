# 07 · 影子追踪进程 doubao-tracker.exe

## 位置

```
desktop/native-shadow-bin/
├── doubao-tracker.exe       (239,616 字节,PE32+ x86-64)
├── enabled                   (9 字节: "enabled\r\n")
└── delivery-gate-enabled    (9 字节: "enabled\r\n")
```

## 二进制特征

```
Format        : PE32+ (x86-64, Windows)
Architecture  : x86_64
Imports       : 仅 KERNEL32.dll (!!!)
Procedures    : 870
Strings       : 529
Entry sections: .text (148K) + .rdata (72K) + .data (9K)

SHA256: 90fe9cd4c627da325e91473b4f31aa7c3b776bd2a8399aef9f94a5139a37e2be
```

**关键观察**:**只导入 KERNEL32.dll**。意味着:
- ❌ 没有 WININET / WINHTTP → 不发 HTTP
- ❌ 没有 WS2_32 → 不开 socket,不发 UDP/TCP
- ❌ 没有 USER32 → 不创建窗口,不读键鼠
- ❌ 没有 ADVAPI32 → 不读注册表,不改权限
- ✅ 只有文件 IO、进程、同步等基础功能

**结论:它不联网,也不碰 Windows UI,只做本地的"文件 + 控制台 IO"**。

## 字符串线索

从 PE 的 .rdata 里挖出所有字符串,过滤后关键的只有:

```
boundConversationId
boundSubmissionId
conversationId
{"status":"
{"status":"invalid_request","schemaVersion":1}
IsDebuggerPresent
```

加上大量的 C++ 标准库符号(`.?AV_System_error@std@@`、`basic_filebuf` 等)。

### JSON 协议特征

`{"status":"invalid_request","schemaVersion":1}` 是**明确的 stdout 消息格式**。工具通过 stdin 接收任务,通过 stdout 返回状态。

## PE 节区布局

```
.text    .rdata   .data    .pdata   .fptable  .reloc   tdb
148K     72K      9K       9K       0.5K      2.5K     6K
```

- `.text`:代码段,870 个过程
- `.rdata`:字符串、常量、异常处理表
- `.pdata`:x64 异常处理 unwind info(Windows x64 ABI 要求)
- `.fptable`:**Control Flow Guard 相关**(编译时启用了 CFG 保护)
- `.reloc`:ASLR 重定位表
- `tdb`:TLS 回调表(可能用于反调试,或仅 CRT 初始化)

**没加壳**,是正常的 MSVC 编译产物。所以能被 Ghidra/IDA 直接反编译。

## 调用链

从 PS 和 Electron 主进程的调用关系推测:

```
Electron 主进程 (main.cjs → doubao-native-shadow.cjs,混淆)
       │
       │ child_process.spawn('doubao-tracker.exe')
       │ 保留 stdin/stdout 管道
       ▼
┌─────────────────────────────────┐
│  doubao-tracker.exe             │
│                                 │
│  while (!eof(stdin)) {          │
│      line = readLine(stdin)     │
│      request = parse_json(line) │
│      // 处理请求:               │
│      //  - 落盘记录任务状态     │
│      //  - 从磁盘读历史记录     │
│      response = handle(request) │
│      write(stdout, response)    │
│  }                              │
└─────────────────────────────────┘
       ▲                 │
       │                 ▼
    stdin (JSON)      stdout ({"status":"..."})
       │                 │
       │          Electron 主进程 读取结果
       │
       └── capture.ps1 产出 nativeShadow 对象 后
           Electron 主进程转发给 tracker 落盘
```

## 工作触发条件(doubao-capture.ps1)

### 开关检查

```powershell
$nativeShadowEnabled = Test-Path -LiteralPath (Join-Path $PSScriptRoot '../native-shadow-bin/enabled')
```

**Only if `enabled` 文件存在就激活**。沈浪可以通过 OTA 删除这个文件来禁用整个 shadow 功能。

### 推送 nativeShadow 数据

```powershell
if ($nativeShadowEnabled -and !$event.error -and
    !$event.result.exceptionDetails -and $value.nativeShadow) {
    try {
        [Console]::Out.WriteLine('DOUBAO_NATIVE_SHADOW:' +
            ($value.nativeShadow | ConvertTo-Json -Depth 8 -Compress))
    } catch {}
}
```

父进程(Electron)读到 `DOUBAO_NATIVE_SHADOW:` 开头的行就把 JSON payload 转发给 `doubao-tracker.exe` 的 stdin。

### nativeShadow 的数据结构(来自 result-reader.js 的 `_0x28971b`)

```js
function nativeShadowSnapshot(body, options, binding, feedback, followupSubmissionIds) {
    if (!options?.nativeShadow) return null;
    const conversationId = binding?.conversationId || options.boundConversationId;
    const submissionId = binding?.submissionId || options.boundSubmissionId;
    if (!conversationId || !submissionId) return null;

    // ... 找最新的 bot 回复
    const replyMessage = findMessage(body, latestSubmissionId);
    if (!replyMessage || replyMessage.user_type !== 2) return null;

    const parentId = String(replyMessage.bot_reply_message_id || '0');
    const isFollowup = followupSubmissionIds.includes(parentId);

    return {
        followup: isFollowup,
        originSubmissionId: String(submissionId),
        input: {
            boundConversationId: String(conversationId),
            conversationId: String(replyMessage.conversation_id),
            boundSubmissionId: isFollowup ? parentId : String(submissionId),
            submissionId: parentId === '0' ? String(submissionId) : parentId,
            submittedAt: binding?.createdAt || options.since,
            messageCreatedAt: msToMillis(replyMessage.create_time),
            replyRole: 'assistant',
            replyFinished: true,
            hasResult: Boolean(latestVideoOrImage),
            isGenerating: false,
            type: options.type,
            message: extractMessageText(replyMessage)
        }
    };
}
```

**它存了什么**:
- `boundConversationId`:豆包会话 ID
- `boundSubmissionId`:原始任务消息 ID
- `submissionId`:当前消息 ID(可能是 followup)
- `type`:'image' / 'video'
- `hasResult`:这条消息是否已有结果
- `message`:bot 回复的文字内容

**为什么要持久化**:
- Electron 的 session 存储(IndexedDB 等)在 Chromium 缓存清理时会丢。
- shadow 这份数据写在 Chromium 不管的目录(可能是 `%APPDATA%\沈浪超级Ai画布\shadow-store\`),清缓存不丢。
- 用户下次打开沈浪画布 → 从 shadow 读历史任务 → 继续轮询豆包是否已有结果 → 恢复断线任务。

## 推测的 stdin/stdout 协议

**无直接证据,基于字符串和架构推测**:

### Request (stdin, JSON per line)

```json
{"action":"put","key":{"boundConversationId":"7501234","boundSubmissionId":"7501235"},"value":{...}}
{"action":"get","key":{"conversationId":"7501234","submissionId":"7501235"}}
{"action":"list","filter":{"since":1697433600000}}
{"action":"delete","key":{"submissionId":"7501235"}}
```

### Response (stdout, JSON per line)

```json
{"status":"ok","schemaVersion":1,"value":{...}}
{"status":"not_found","schemaVersion":1}
{"status":"invalid_request","schemaVersion":1}
```

那段固定字符串 `{"status":"invalid_request","schemaVersion":1}` 就是**所有校验失败情况下的统一返回**。

## 持久化到哪

**未直接确认**,但 Windows Electron 应用的习惯位置:

```
%APPDATA%\沈浪超级Ai画布\                     ← Electron userData 默认
%APPDATA%\沈浪超级Ai画布\shadow-store\         ← 可能的路径
%LOCALAPPDATA%\shenlang-ai-canvas\shadow\     ← 另一种可能
```

tracker 通过 KERNEL32 的 `CreateFileW` + `FindFirstFileExW` + `FindNextFileW` 操作文件系统。可能用 LevelDB-like 的键值存储或者简单的 JSON 文件。

**用静态分析可以确认**:在 Ghidra 里跟 CreateFileW 的调用参数就能看到具体路径。

## 不是什么

**明确排除**:
- ❌ 不是网络后门(没网络导入)
- ❌ 不是 keylogger(没 USER32,没 SetWindowsHookEx)
- ❌ 不是进程注入器(没 OpenProcess + VirtualAllocEx + WriteProcessMemory 组合)
- ❌ 不是文件窃取者(KERNEL32 的 CreateFileW 本身是正常文件 IO)
- ❌ 不是提权工具(没 ADVAPI32 的 LookupPrivilegeValue)

**它就是个本地 KV 存储代理**,设计上保守得不像恶意软件。

## 为什么用独立二进制

好问题 —— 为什么不直接在 Electron 主进程里存?猜测:

1. **解耦 Electron 生命周期**:Electron 崩溃时主进程挂,但 tracker 可以作为孤儿进程继续写入最后的状态。
2. **跳过 Chromium 的沙箱限制**:Chromium 对磁盘访问有一些限制(主要在 renderer),独立进程可以更自由。
3. **可以单独更新**:tracker 版本和沈浪主程序版本解耦,出 bug 时可以单独补丁。
4. **二进制形式难以篡改**:相比 JavaScript,逆向难度更高。字节的核心校验逻辑(如果真的在 tracker 里)可以更稳固。

## 验证我的推测(如果以后要深入)

### 用 Ghidra 跟 CreateFileW

```
# 在 Ghidra:
1. 打开 doubao-tracker.exe
2. Symbol Tree → Imports → KERNEL32.DLL → CreateFileW
3. Find References To → 跳到调用点
4. 分析参数:lpFileName 来自哪(TLS? Global 变量? 栈上的字符串?)
```

### 用 Process Monitor 跑一次

```
1. Procmon.exe (sysinternals) 打开
2. Filter: Process Name is "doubao-tracker.exe"
3. 用沈浪跑一次视频生成
4. 观察 tracker 读写了哪些文件
```

### 读 Electron 主进程的 doubao-native-shadow.cjs

```
1. 反混淆 main.cjs / doubao-native-shadow.cjs
2. 找 'doubao-tracker.exe' 字符串的 spawn 调用
3. 跟踪它写入 stdin 的格式
```

## 中转程序要不要移植这个?

**三种选择**:

### 选择 A:整个砍掉

**最简单**。中转程序本身就是服务,直接用 SQLite / Postgres 做任务状态存储,不需要 shadow 二进制。

Tauri 的 Rust 侧直接:
```rust
use sqlx::SqlitePool;

pub async fn save_task(pool: &SqlitePool, task: &Task) -> Result<()> {
    sqlx::query!(
        "INSERT OR REPLACE INTO tasks (submission_id, conversation_id, status, result_json, updated_at)
         VALUES (?, ?, ?, ?, strftime('%s', 'now'))",
        task.submission_id, task.conversation_id, task.status, task.result_json
    ).execute(pool).await?;
    Ok(())
}
```

### 选择 B:用 Rust 重写 tracker

如果中转程序需要支持"主进程崩溃后恢复"场景,可以在 Rust 内再 fork 一个 watcher 线程,把关键状态写进独立的 sled/jammdb 数据库。

### 选择 C:复用原版 tracker.exe

**不推荐**。因为:
- 二进制黑盒,出 bug 没法 debug
- 协议未完全还原,只能通过"喂进去然后看它写出什么"的方式使用
- 协议未来可能变

**推荐 A**。中转程序作为服务,持久化应由中转程序本身负责,不需要 shadow。

## 对调试的建议

如果开发中真要动它:

```powershell
# 替换 enabled flag 为空(禁用 shadow 功能)
Remove-Item -LiteralPath 'native-shadow-bin/enabled'

# 或者修改成空内容
Set-Content -Path 'native-shadow-bin/enabled' -Value ''
```

然后沈浪的捕获管线会继续跑,只是不写 shadow,方便你观察主流程。

## 下一步

[08-anti-conflict.md](./08-anti-conflict.md) — 讨论工具跑任务时用户正在用豆包的冲突面。
