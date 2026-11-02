# 09 · CDP 协议使用契约

> 本文档只记录 doubao-relay 对 CDP 协议的**使用方式契约**,不是 CDP 协议本身的说明书。CDP 协议细节见 [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)。
>
> 更底层的实现细节(具体响应字段、JS 代码片段)见 `shenlang-analysis/02-cdp-bootstrap.md`、`05-capture-pipeline.md`。

## 连接层

### 握手流程

```
1. HTTP GET http://127.0.0.1:9707/json/list
   → 返回 target 列表 (type/url/webSocketDebuggerUrl)

2. 筛选:
   type == "page" && url.host in {"doubao-chat", "doubao-background"}

3. 对每个关心的 target:
   WebSocket Connect ws://127.0.0.1:9707/devtools/page/<id>

4. 发送 Network.enable:
   { "id": 1, "method": "Network.enable",
     "params": { "maxTotalBufferSize": 16777216, "maxResourceBufferSize": 8388608 } }

5. 等待 { "id": 1, "result": {} } → 握手完成
```

### 安全校验(每次重连都要重跑)

- `Get-NetTCPConnection -LocalPort 9707` 监听进程 MUST 是 Doubao
- 监听地址 MUST 是 127.0.0.1 或 ::1
- `webSocketDebuggerUrl` scheme/host/port MUST 匹配
- target.url scheme MUST ∈ {`doubao`, `chrome`},host MUST ∈ 白名单

对应 Rust 代码应放在 `cdp::connection::validate_target()`。

### 多 session 管理

| session | 负责 | 持续时间 |
|---|---|---|
| chat | 账号读取、生成操作、结果捕获 | 持续,断开自动重连 |
| background | 后台观察(账号信息、submissionIdentity 补充) | 持续,断开自动重连 |

两个 session 独立 WebSocket,独立 broadcast channel。

### 断线检测与重连

- 每 30 秒发 CDP 心跳 (`{"id":...,"method":"Target.getTargets"}`)
- 心跳 5 秒内无响应 → 判定断开
- 断开后:
  1. 关掉老 socket
  2. 等 2 秒
  3. 重新走握手流程
  4. 若 20 秒内未成功 → 升级到"重启豆包"(F1.3+F1.2+F1.4)

## 命令规范

### 命令格式

```json
{
  "id": <u64, 单调递增>,
  "method": "<Domain>.<Command>",
  "params": { ... }
}
```

### 响应格式

```json
{
  "id": <u64, 匹配 request>,
  "result": { ... }
}
// 或
{
  "id": <u64>,
  "error": { "code": -32xxx, "message": "..." }
}
```

### 事件格式(没有 id)

```json
{
  "method": "<Domain>.<Event>",
  "params": { ... }
}
```

### 使用到的 CDP 命令

| 命令 | 用途 |
|---|---|
| `Network.enable` | 开启网络事件推送 |
| `Network.getResponseBody` | 读取响应 body |
| `Page.setWebLifecycleState` | 保持豆包页面活跃 |
| `Emulation.setFocusEmulationEnabled` | 焦点伪装 |
| `Page.setInterceptFileChooserDialog` | 拦截文件对话框 |
| `Input.dispatchMouseEvent` | 模拟鼠标(滑块拖拽) |
| `Input.dispatchKeyEvent` | 模拟键盘(方向键微调) |
| `DOM.getDocument` | 获取 DOM 根 |
| `DOM.querySelector` | 查元素 |
| `DOM.setFileInputFiles` | 塞文件到 input |
| `Runtime.evaluate` | 执行 JS(主力) |
| `Target.getTargets` | 心跳 |

### 使用到的 CDP 事件

| 事件 | 用途 |
|---|---|
| `Network.requestWillBeSent` | 识别"这是一次生成请求"或"这是一次历史查询" |
| `Network.responseReceived` | 视频任务:记下 video/* MIME 的 URL |
| `Network.loadingFinished` | 触发 getResponseBody 读 body |
| `Network.eventSourceMessageReceived` | SSE 流式数据 |
| `Page.fileChooserOpened` | 文件对话框拦截通知 |

### 命令 ID 分配

- 每个 session 独立 counter,从 1 开始
- u64 原子递增

## JS 注入契约

### 注入时机

所有 JS 都通过 `Runtime.evaluate` 注入到豆包页面,**不用 Page.addScriptToEvaluateOnNewDocument**(页面已加载,不需要)。

### 注入文件(直接复用 shenlang-analysis 的 3 个 JS)

| 文件 | 页面对象 |
|---|---|
| `doubao-generation-page.js` | `shenlangDoubaoGeneration` |
| `doubao-result-reader.js` | `shenlangDoubaoResultReader` |
| `doubao-video-decoder.js` | `shenlangDoubaoVideoDecoder` |

**Rust 侧**用 `include_str!("scripts/doubao-generation-page.js")` 常量嵌入,运行时拼接发给 CDP。

### 幂等检查

每次调用前检查对象是否已注入:

```js
if (typeof shenlangDoubaoGeneration === 'undefined') {
    <生成页脚本源码>
}
```

页面可能 reload,所以每次都要检查。

### 命名空间防撞

多个 doubao-relay 实例或并发任务可能同时注入。按 shenlang-analysis 的做法,用 trackingKey 替换标识符:

```rust
pub fn with_tracking_key(script: &str, tracking_key: &str) -> String {
    script.replace("shenlangDoubaoGeneration", tracking_key)
}
```

**v0.1 简化**:单实例,不并发注入同一页面,trackingKey 可以固定。但代码预留出口。

### 返回值契约

所有 JS 入口 MUST 返回可 JSON 序列化的值。约定:
- 成功:`{ok: true, ...}`
- 失败:`throw new Error('中文消息')`
- 异步:返回 `Promise`,Rust 侧用 `awaitPromise: true`

Rust 侧统一处理:

```rust
pub async fn evaluate_js<T: DeserializeOwned>(
    session: &CdpSession,
    expression: &str,
    timeout: Duration,
) -> Result<T, EvaluateError> {
    let result = session.send("Runtime.evaluate", json!({
        "expression": expression,
        "returnByValue": true,
        "awaitPromise": true,
    })).await?;
    if let Some(exc) = result.get("exceptionDetails") {
        return Err(EvaluateError::JsException(parse_message(exc)));
    }
    serde_json::from_value(result["result"]["value"].clone())
        .map_err(EvaluateError::DeserializationFailed)
}
```

## 大整数 JSON 处理

豆包的 ID 字段(`conversation_id`、`message_id`、`local_message_id`、`bot_reply_message_id`)都是 >16 位的整数,超过 JS Number 精度,也超过 `serde_json::Number` 的 `as_i64` 安全范围。

**处理规约**:
- Rust 侧从 CDP 返回的 string 文本中读 JSON 前,先跑正则替换:

```rust
static BIG_INT_FIELDS: Lazy<Regex> = Lazy::new(|| Regex::new(
    r#""(conversation_id|local_message_id|message_id|bot_reply_message_id)"\s*:\s*(-?\d{16,})(?=\s*[,}])"#
).unwrap());

fn fix_big_integers(json_text: &str) -> String {
    BIG_INT_FIELDS.replace_all(json_text, r#""$1":"$2""#).to_string()
}
```

- 这些 ID 在 Rust 侧始终用 `String` 表示,不转数字

## 并发安全

### Session 的并发使用

一个 `CdpSession` 可以被多个任务并发调用 `send` / `evaluate`:
- 内部每条命令独立 id,独立 oneshot channel 等响应
- WebSocket 的写用 mpsc channel 串行化
- 读循环 + broadcast channel 把事件分发给所有订阅者

### 消息总顺序

CDP 消息按发送顺序到达,但不保证按发送顺序响应(因为命令可能是异步执行)。实现:**不依赖响应顺序,只依赖 id 匹配**。

## 错误处理

### 分类

| 类型 | 原因 | 恢复动作 |
|---|---|---|
| 连接错 | WebSocket 断 | 重连 session |
| 协议错 | 响应格式不对 | 重启豆包 |
| 命令错 | CDP error(method 不存在等) | 日志 + 返回 error |
| 页面错 | JS 异常 | 日志 + 返回 EvaluateError |
| 超时错 | 等响应超过 5 秒 | 判定命令失败 |

### 超时

- 每条 CDP 命令默认 15 秒超时
- `Runtime.evaluate` 默认 30 秒超时(JS 可能慢)
- 超时后 oneshot sender 被 drop,caller 收到 timeout error

## CDP 日志调试面板

当 `cdp_debug_log=true`,每条发送和接收的 CDP 消息都 log 到:
- `~/.../logs/cdp-debug.log`
- 以及 Tauri 事件 `cdp_log_entry`(供 UI 调试面板实时展示)

**格式**:JSON lines,一行一事件。字段:`{ts, direction, method?, params?, result?, event?}`。

**用途**:豆包改版时诊断"哪一步失效了"。

## 契约版本化

本 CDP 契约绑定豆包版本。豆包升级可能导致:
- URL 规则变化(新端点)
- JS 对象结构变化(image_ori_raw 字段消失)
- DOM 选择器失效

**监控策略**:
- 启动时跑"UI 兼容性自检"(调关键 JS 入口,检查返回合法)
- 失败则 UI 弹红色警告 + API 返回 `doubao_ui_changed`
- 不自动试错,等用户升级 doubao-relay
