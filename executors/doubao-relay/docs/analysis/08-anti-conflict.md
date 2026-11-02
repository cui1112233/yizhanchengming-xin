# 08 · 并发冲突 — 工具跑任务时用户同时用豆包

## 为什么需要专门分析

**豆包不是无人豆包**。中转程序跑任务时,用户可能同时:
- 在豆包聊天框打字
- 切换到别的对话查看以前的消息
- 切换账号
- 最小化豆包窗口转去干别的
- 等等

这些场景会不会冲突?答案分三层。

## 第 1 层:物理鼠标键盘(基本不抢)

### 现代路径都是事件注入

所有的"操作"分两类技术:

| 操作 | 技术 | 是否抢物理设备 |
|---|---|---|
| 填 prompt | `document.execCommand('insertText')` | ❌ 纯 DOM API |
| 点"发送"按钮 | `dispatchEvent(new PointerEvent/MouseEvent)` | ❌ 纯 DOM 事件 |
| 选模型菜单 | DOM `.click()` | ❌ |
| 拖视频时长滑块 | CDP `Input.dispatchMouseEvent` | ❌ 事件注入到窗口客户区,不改真实光标 |
| 按方向键微调时长 | CDP `Input.dispatchKeyEvent` | ❌ 事件注入到页面,不占物理键盘 |
| 读账号列表 (DOM 路径) | 纯 JS evaluate | ❌ |
| 读账号列表 (UIA 路径) | Windows UIAutomation Pattern | ❌ 走 Accessibility 层,不触发真鼠标 |

**这些都是把事件直接 dispatch 到豆包窗口的事件总线**,不走系统鼠标/键盘输入队列。

### 焦点伪装(关键于不抢你的焦点)

启动阶段两条命令让豆包"误以为自己有焦点":

```powershell
Command 'Page.setWebLifecycleState' @{ state = 'active' }
Command 'Emulation.setFocusEmulationEnabled' @{ enabled = $true }
```

效果:
- 豆包页面 `document.hasFocus()` 返回 true,即使你最小化它
- `document.visibilityState` 返回 'visible'
- 页面继续全速运行(不被 Chromium 后台节流)
- React 组件认为"用户在正常用"

**所以**:中转程序跑任务时,你可以放心去打微信、刷网页、玩游戏。它不会抢你的焦点,豆包窗口可以缩在后面"自己动"。

### 真会抢物理鼠标的例外路径(极少触发)

两条冷门 fallback:

#### 1. 老版 `doubao-accounts.ps1` 的 `AccountMouse.Click`

```powershell
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class AccountMouse {
    [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
    [DllImport("user32.dll")] public static extern void mouse_event(uint f, uint x, uint y, uint d, UIntPtr e);
    public static void Click(int x, int y) {
        SetCursorPos(x, y);
        mouse_event(2, 0, 0, 0, UIntPtr.Zero);  // LEFTDOWN
        mouse_event(4, 0, 0, 0, UIntPtr.Zero);  // LEFTUP
    }
}
'@
```

**这段代码真的会把物理光标瞬移并左键**。但 operation=read/connect/switch 全部重定向到 `accounts-background.ps1`:

```powershell
# accounts.ps1:4-11
if ($request.operation -in @('read','connect','switch')) {
    & (Join-Path $PSScriptRoot 'doubao-accounts-background.ps1') ...
    exit 0        # ← 新路径直接返回,下面的代码跑不到
}
```

**结论**:老代码死掉了,实际触发不到。

#### 2. 原生文件对话框操作 (Select-NativeUploadFile)

```powershell
# 填"文件名"框 + 按"打开"
$value.SetValue($file)        # UIA ValuePattern
$invoke.Invoke()              # UIA InvokePattern
```

UIA 的 `SetValue` 和 `Invoke` 都走 Accessibility 层,**不是模拟鼠标**。但当豆包弹原生对话框时,**对话框本身就是模态的**,你的鼠标键盘本来就被挡住了。所以这个场景实际上不冲突。

### 第 1 层总结

**几乎没有物理鼠标键盘冲突**。你可以放心做别的事。

## 第 2 层:共享 DOM(真会打架)

这是**真正的冲突面**。中转程序和你操作的是**同一个豆包页面的同一个 DOM**。

### 翻车场景 A:你正在豆包聊天框打字

工具的 fill 流程:

```js
const input = document.querySelector('textarea,[contenteditable="true"],[role="textbox"]');
range.selectNodeContents(input);     // 选中输入框所有内容
selection.removeAllRanges();
selection.addRange(range);
document.execCommand('insertText', false, request.prompt);  // 用 prompt 覆盖
input.dispatchEvent(new Event('input', { bubbles: true }));
```

**如果此刻你正在打字**:
- 你刚打了"今天天气真"
- 工具 execCommand('insertText') **插在你光标处**(或替换 selection)
- 你看到:"今天天气真请使用 Seedance 2.0 Mini 生成5秒视频..."
- 工具紧接着 clickLikeHuman(发送按钮) → **发送**
- 结果:你的"今天天气真"和它的 prompt 被拼成一条消息发给豆包

**这会污染任务输入**。工具后面从消息里挖结果可能挖不到(prompt 匹配失败)。

### 翻车场景 B:你同时按 Enter 发消息

工具 clickLikeHuman + 你按 Enter → 两次发送,豆包可能:
- React 的防抖吞掉一次(常见)
- 偶尔双发(低概率)
- 发了两条,工具 submissionIdentity 只认一条,另一条孤立悬挂

### 翻车场景 C:你中途切换对话

工具靠 `targetId` 定位豆包的 chat 页。你切换对话时豆包不会换 target(同一页面,只换 route),但:

```powershell
function Refresh-ChatConnection {
    $chatTargets = @($liveTargets | Where-Object { ... Host -eq 'doubao-chat' })
    if ($chatTargets.Count -ne 1) {
        $report.lastReadReason = 'chat_page_not_unique'
        return
    }
    ...
}
```

工具检查 doubao-chat 页**必须唯一**。如果你切出了多个 chat 页(新建对话 / 分屏)→ 工具会抛 `chat_page_not_unique` 停摆。

### 翻车场景 D:你中途切换账号

账号切换会让豆包 reload 整个页面,所有 CDP targetId 失效:

```powershell
# accounts-background.ps1:395
if ($reconnectNeeded) {
    $socket.Dispose()
    $freshTargets = Invoke-RestMethod ... /json/list
    $socket = New-Object AccountReadSocket
    $socket.Connect($freshTargets[0].webSocketDebuggerUrl)
    ...
}
```

但 **capture.ps1 没有这个重连逻辑完全体**。account 模块有重连,capture 模块只有 `Refresh-ChatConnection` 简单重连。

**后果**:正在跑的生成任务会断,`boundSubmissionId` 找不到对应消息,任务作废。

### 翻车场景 E:你最小化豆包窗口

好消息:**无影响**。`setFocusEmulationEnabled` 让豆包继续工作。

### 翻车场景 F:你开了多个豆包窗口

豆包桌面端默认不支持多开。但如果某种方式开了两个 Doubao.exe 进程:
- 工具的 `Get-NetTCPConnection -LocalPort 9707` 校验会要求唯一监听进程
- 如果真的有两个进程都在 listen → 直接抛错停摆

## 第 3 层:会话上下文(工具做了防混淆)

**好消息**:工具不会把你的消息错当成生成结果。

### submissionIdentity 身份锚

点发送的瞬间抓到 POST body 里的 `local_message_id`,整个任务链就锁定了。你同时在**别的对话**里发消息,工具一个都不认领。

### expectedConversationId 过滤

每一次历史重放的 analyze() 都做严格检查:

```js
if (options?.requireBinding && options.boundConversationId !== conversationId) {
    return { ok: true, images: [], feedback: null, binding: null };
}
```

**换会话 → 立即过滤掉**。

### followupSubmissionIds 认领链

即使你在**同一对话**里穿插了自己的消息,工具只会认领属于 chain 的消息:

```js
// 认领进 chain 的条件:
// 1. 用户发的消息
// 2. 消息 ID != 原始 submissionId
// 3. 包含 prompt 文本 + "以上参数已由...核对" 关键词 OR
//    上一条是"是否继续生成"提示 + 消息文本是"继续生成"
chain.add(msg.id);
```

你随口打的"你好"不会被当成 followup。

### 过滤自身递归

```powershell
# capture.ps1:237
if ($parallelTracking -and
    @($event.params.initiator.stack.callFrames |
      Where-Object { $_.functionName -eq 'shenlangReadBoundConversation' }).Count) {
    continue
}
```

**工具自己 fetch 的历史请求不会被当成"新生成"**。这避免了无限递归。

## 并发冲突总览

| 场景 | 风险 | 建议 |
|---|---|---|
| 你最小化豆包窗口 | ✅ 无 | 推荐这种工作模式 |
| 你去用别的程序(微信/浏览器/编辑器) | ✅ 无 | 焦点伪装生效 |
| 你在豆包的别的对话里和别人聊天 | ⚠️ 低。identity 过滤生效 | 可以 |
| 你在它正在用的 chat 页上**查看历史消息**(滚动不发消息) | ✅ 无 | 可以 |
| 你在它正在用的 chat 页上**打字准备发消息** | 🔴 高。execCommand 插进你光标 | 禁止 |
| 你同时让豆包和中转程序各提交一个任务 | ⚠️ 中。同一账号配额争抢 | 等一个完 |
| 你中途切换豆包账号 | 🔴 高。整条链断 | 禁止 |
| 你中途新建对话 → 多个 doubao-chat 页 | 🔴 高。工具会停摆 | 禁止 |

## 中转程序的设计建议

### 建议 1:给用户明确告警

```
┌───────────────────────────────────────┐
│ ⚠️ 任务进行中                           │
│                                       │
│ 请不要在豆包主聊天框里打字,            │
│ 不要切换豆包账号,                      │
│ 不要新建对话。                         │
│                                       │
│ 其他操作(最小化窗口、浏览历史、      │
│ 切到别的对话查看)均可。              │
│                                       │
│  [了解]                                │
└───────────────────────────────────────┘
```

### 建议 2:监测用户干扰并抛错

Rust 侧 CDP 循环里:

```rust
// 定期检查 chat 页数量
let targets = cdp_list_targets().await?;
let chat_pages = targets.iter().filter(|t| t.url.host() == "doubao-chat").count();
if chat_pages != 1 {
    task.fail("user_created_multiple_chats", "检测到用户新建了对话,任务已终止");
    return;
}

// 定期检查输入框内容是否被改
let input_value = cdp_evaluate(session, "findInput().value || findInput().innerText").await?;
if input_value != expected_prompt_value {
    task.fail("user_modified_input", "检测到用户修改了输入框,任务已终止");
    return;
}
```

### 建议 3:禁用并发任务到同一账号

```rust
struct TaskQueue {
    accounts: HashMap<AccountName, Mutex<AccountState>>,
}

impl TaskQueue {
    async fn submit(&self, account: &str, task: Task) -> Result<TaskHandle> {
        let state = self.accounts.get(account).ok_or("unknown account")?;
        let guard = state.lock().await;  // ← 同一账号串行
        process_task(task).await
    }
}
```

### 建议 4:避免账号自动轮换

**不要**在中转程序里实现"账号池自动轮询"。原因:
- 频繁切换触发豆包风控
- 切换中失败率高
- 用户看到自己账号被莫名切换会很困惑

**替代做法**:让 API 调用方指定账号,中转程序只负责"切到这个账号,如果不是的话"。

### 建议 5:UI 展示工具"正在用哪个对话"

```
┌───────────────────────────────────────┐
│ 中转状态                               │
│                                       │
│ 当前账号: 张三                         │
│ 工作对话: "生图任务-20261016-01"       │
│ 任务状态: 等待第 2 次确认(自动)      │
│                                       │
│ 请勿在上述对话里手动操作              │
└───────────────────────────────────────┘
```

让用户清楚知道哪个对话"归工具用",哪些对话归自己用。

## 测试清单(验证你的 Tauri 实现没有打架)

| 测试 | 预期 |
|---|---|
| 工具跑生成时,用户在豆包别的对话里打字 | 工具正常完成,用户消息不影响 |
| 工具跑生成时,用户滚动查看历史 | 工具正常完成 |
| 工具跑生成时,用户最小化豆包 | 工具正常完成 |
| 工具跑生成时,用户点击"新建对话" | 工具探测到并弹错,任务作废 |
| 工具跑生成时,用户在工作对话里打字 | 工具探测到输入框脏数据并弹错 |
| 工具跑生成时,用户切换账号 | 工具探测到并终止任务,UI 提示 |
| 工具跑完生成后,用户立即发第二次 | 排队,一个一个处理 |

## 下一步

[09-reuse-for-tauri.md](./09-reuse-for-tauri.md) — Tauri 中转程序开发的完整可复用清单。
