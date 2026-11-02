# 开发决策记录

> 2026-10-16 对齐。本文档记录中转程序开发前已对齐的决策,与"沈浪本身是 Windows-only"的事实区分开。

## D1 · 账号调度:调用方指定

**决策**:API 请求必须带 `account` 字段,中转程序只负责"切到这个账号,如果当前不是"。

**不做**:自动轮询、自动负载均衡、后台静默切换。

**理由**:
- 频繁账号切换触发豆包风控
- 调用方对自己账号的配额、权限了解比中转程序清楚
- 简化状态管理:中转程序不维护账号池

**API 行为**:
```http
POST /api/generate
{
  "account": "张三",       // 必填或可选?见下
  "type": "video",
  ...
}
```

**待定子项**:
- `account` 字段是必填还是可选?可选时默认用当前账号?**建议必填**,避免歧义。
- 账号标识用 `name` 还是别的 ID?豆包 UI 上同名不同号的处理?**建议**用`name`,同名情况开发时如果真遇到再加 `label` 消歧。
- 账号不存在/未登录怎么返回?**建议**:400 + error_code `account_not_found` / `account_not_logged_in`。

## D2 · API 模式:异步 task_id + 轮询

**决策**:提交生成请求立即返回 task_id,调用方轮询状态。

**接口形态**:

```http
POST /api/generate
{
  "account": "张三",
  "type": "image",
  "model_name": "Seedream 5.0 Pro",
  "prompt": "...",
  "ratio": "16:9",
  "reference_files": ["<pre-uploaded id>"]    // 或 base64
}
→ 200 OK
{
  "task_id": "01HQZ...",
  "status": "pending",
  "created_at": 1697433600
}

GET /api/tasks/{task_id}
→ 200 OK
{
  "task_id": "01HQZ...",
  "status": "generating",              // pending|preparing|generating|completed|failed
  "stage": "checking",                 // 来自豆包 feedback 阶段
  "progress": 0.4,                     // 可选,如果能估算
  "message": "豆包正在确认参数",
  "result": null                       // completed 时才有
}

GET /api/tasks/{task_id}
→ 200 OK (完成态)
{
  ...
  "status": "completed",
  "result": {
    "files": [
      { "url": "/api/tasks/.../download/0", "mime": "image/png", "width": 1920, "height": 1080, "bytes": 2456321 }
    ]
  }
}

GET /api/tasks/{task_id}/download/{index}
→ 200 OK (binary)
Content-Type: image/png / video/mp4
```

**理由**:
- 视频 5-15 分钟,HTTP 长连接撑不住
- 轮询简单,无状态
- 调用方可以自己决定轮询频率

**可选扩展**(未来):
- Webhook 回调(请求里带 `callback_url`,生成完 POST 回去)
- Server-Sent Events(调用方实时拿进度)

**待定子项**:
- 轮询建议频率:**建议**每 3 秒,文档明确告知。
- 任务过期策略:生成后保留 24 小时,之后清理文件 + 保留元数据 7 天。
- `reference_files` 的传递方式:**待定**。候选:
  - 先 `POST /api/uploads` 上传参考图拿 id → generate 带 id
  - 直接 base64 内联(简单但请求体大)
  - 文件路径(安全风险,只在内网局域网时可用)

## D3 · 平台:Windows + macOS 双平台

**决策**:中转程序需同时支持 Windows 和 macOS 的豆包桌面端。

**沈浪现状**:**仅支持 Windows**。所有 PS1、注册表读取、Win32 UIA 都是 Windows-only。

**macOS 实现**:**从零开始**,但基于与 Windows 相同的 CDP 协议。

### 跨平台的部分(两边一样)

| 模块 | 跨平台实现 |
|---|---|
| CDP WebSocket 协议 | `tokio-tungstenite`,行为完全一致 |
| 注入 JS(3 个文件) | 完全复用,豆包页面渲染一致 |
| Network 事件监听 | 完全一致 |
| 视频 AES-CBC 解密 | 纯算法,完全一致 |
| 任务队列、SQLite 持久化、HTTP server | 纯 Rust,跨平台 |

### 平台差异的部分

| 模块 | Windows | macOS |
|---|---|---|
| 发现豆包安装路径 | 注册表 App Paths + Uninstall + `%LOCALAPPDATA%` + 盘根扫描 | `/Applications/*Doubao*.app` + `~/Applications/` + `mdfind "kMDItemFSName == 'Doubao.app'"` |
| 启动带调试端口 | `Start-Process Doubao.exe --remote-debugging-port=9707` | `/Applications/Doubao.app/Contents/MacOS/Doubao --remote-debugging-port=9707` 或 `open -a Doubao --args --remote-debugging-port=9707` |
| 杀掉现有豆包进程 | `Stop-Process -Name Doubao` | `pkill -x Doubao` 或 `kill -TERM <pid>` |
| 枚举进程 | sysinfo crate(跨平台可用) | sysinfo crate |
| 原生文件对话框 fallback | Win32 `#32770` + UIA | **多数情况不需要** —— macOS 的 Electron 调起 Cocoa NSOpenPanel 时,CDP 的 `Page.setInterceptFileChooserDialog` 仍然能拦住(因为它拦截在 Chromium 侧,不是系统侧) |
| 账号读取备用通道 | Windows UIAutomation | macOS Accessibility API(需用户在系统偏好-安全与隐私授权) |
| HTTP server 绑定 | `127.0.0.1:8080`(可改) | 一样 |
| 日志目录 | `%APPDATA%\doubao-relay\logs\` | `~/Library/Application Support/doubao-relay/logs/` |
| SQLite 文件 | `%APPDATA%\doubao-relay\tasks.db` | `~/Library/Application Support/doubao-relay/tasks.db` |

### Phase 0 验证结果(2026-10-16 已完成)

**豆包 macOS 版完全支持 CDP**,核心机制与 Windows 一致。实测结论:

#### ✅ 一致的部分

| 项 | 实测结果 |
|---|---|
| CDP 启动参数 | `open -a Doubao --args --remote-debugging-port=9707` 可用 |
| URL scheme | `doubao://` 一致 |
| 页面 host | `doubao-chat` / `doubao-background` 一致 |
| WebSocket 端点 | `ws://127.0.0.1:9707/devtools/page/<id>` 一致 |
| 输入框 | `[contenteditable="true"]` 可定位(=1) |
| 账号按钮位置 | 左下角(与 Windows 一致,x<80, y 近底部) |
| 账号按钮 `aria-haspopup` | `"menu"` 一致 |
| 侧边栏 DOM 可访问 | ✅ |

#### 🔴 macOS 不同的部分(需要适配)

| 属性 | Windows | macOS | 适配方案 |
|---|---|---|---|
| 账号按钮 `aria-label` | `"avatar 张三 体验套餐"` | `null` | **不用 aria-label**,改用 `aria-haspopup="menu"` + 位置 + `<img>` 三特征组合 |
| 账号头像 `<img alt>` | `"avatar"` | `""`(空) | **不用 alt 匹配**,改用"有 `<img>` 即可" |
| 账号按钮文本结构 | `"张三 体验套餐"`(单行) | `"张三\n体验套餐"`(两行) | 解析时同时支持换行分隔和空格分隔两种 |

#### 👁 macOS 实测的额外发现

- 豆包 macOS 版有一个 `drive-iframe` 的 iframe(网盘),但 `type: "iframe"`,工具按 `type === "page"` 过滤即可忽略。
- 还有一个 `cross-site-support` 页,`type: "other"`,也会被过滤规则排除。
- `document.visibilityState === "visible"`,焦点伪装(`Emulation.setFocusEmulationEnabled`)在 macOS 上可能可选 —— 待开发时确认。

#### 跨平台 selector 建议(写进 Rust 常量字符串)

```js
// 跨平台找账号按钮
function findAccountButton() {
  return [...document.querySelectorAll('button')]
    .find(el => {
      const r = el.getBoundingClientRect();
      return el.getAttribute('aria-haspopup') === 'menu' &&
             el.querySelector('img') &&
             r.left < 80 && r.top > innerHeight - 120 &&
             r.width >= 100 && r.width <= 360;
    });
}

// 跨平台解析账号名
function parseAccountLabel(button) {
  const text = (button.innerText || '').trim();
  const lines = text.split('\n').map(s => s.trim()).filter(Boolean);
  const subRe = /^(体验|标准|加强|高级|专业|基础)套餐$|套餐$/;
  // macOS: 两行
  if (lines.length >= 2 && subRe.test(lines[1])) {
    return { name: lines[0], subscription: lines[1] };
  }
  // Windows: 单行
  const m = lines[0]?.match(/^(.+?)\s+((?:体验|标准|加强|高级|专业|基础)套餐)$/);
  if (m) return { name: m[1].trim(), subscription: m[2] };
  return { name: lines[0] || '', subscription: '' };
}
```

**结论**:Phase 1 可直接开始,Windows 和 macOS **共用一套 Rust 代码 + 一套注入 JS**,仅平台相关的"启动豆包 / 发现路径 / 杀进程"3 处需要分支。

---

### 原规划的实测清单(仅作参考,已完成)

**Phase 1 开发前必须验证**:

1. **豆包 macOS 版是否支持 CDP**
   ```bash
   # 终端
   /Applications/Doubao.app/Contents/MacOS/Doubao --remote-debugging-port=9707 &
   sleep 3
   curl http://127.0.0.1:9707/json/list
   ```
   预期:返回 JSON 包含 `type:"page"` 的目标,host 是 `doubao-chat` 或 `doubao-background`。
   如果返回空数组或连不上 → 豆包 macOS 版可能禁了 CDP,整条路径不可行。

2. **豆包页面 DOM 结构在 macOS 版是否一致**
   启动后在 DevTools 里跑一遍 `document.querySelectorAll('[aria-label^="avatar"]')`,看是否能找到账号按钮。
   如果 DOM 不同 → 需要为 macOS 单独写 JS 适配(但大概率一样,React 代码不区分 OS)。

3. **豆包 macOS 版的生成流程是否和 Windows 一致**
   手动在 Mac 上点一次视频生成,观察:
   - 模型菜单是不是 Electron 原生菜单(macOS 下 Electron 更容易用原生 menu bar)
   - 文件上传对话框是不是 NSOpenPanel
   - 时长滑块的拖拽逻辑

### 建议开发顺序

**Phase 1-2 先做 Windows 版**,因为沈浪代码可以逐条对照验证。
**Phase 3 在 Mac 上做跨平台适配** —— 到时候 Windows 版已经稳,移植难度低。

### 待定子项

- Tauri 打包:**建议**用 universal binary(Intel + Apple Silicon)。
- 代码签名:macOS 必须签名才能分发;Windows 不签名会被 SmartScreen 拦。两边都需要开发者证书。**待定**:要不要走企业签名 / 开发者 ID。

## 开发路线(修订版)

### Phase 0:macOS 可行性验证(1 天)

0.1 ✅ 在 Mac 上实测豆包 CDP 可用性(上述 3 件事)
0.2 ✅ 若不可用,先只做 Windows,macOS 支持延后或放弃

### Phase 1:Windows 最小闭环(2 周)

1.1 ✅ Tauri 骨架
1.2 ✅ CDP 客户端
1.3 ✅ 启动豆包 + 账号列表
1.4 ✅ WebView 显示账号

### Phase 2:Windows 生图(2 周)

2.1 ✅ 生图核心流程
2.2 ✅ 图片原图捕获与下载
2.3 ✅ HTTP API 骨架 + task_id 轮询

### Phase 3:Windows 生视频(2 周)

3.1 ✅ 视频时长/比例/参考图上传
3.2 ✅ 视频三条路径捕获
3.3 ✅ Rust AES-CBC 解密 + 单测

### Phase 4:并发 + 多轮确认(1-2 周)

4.1 ✅ parallelTracking 身份锚
4.2 ✅ 自动点"继续生成"
4.3 ✅ 任务队列(按账号串行)
4.4 ✅ 用户干扰检测

### Phase 5:macOS 适配(2 周)

5.1 ✅ 平台抽象层(trait Platform)
5.2 ✅ macOS 安装路径发现
5.3 ✅ macOS 启动 + 杀进程
5.4 ✅ Mac 上端到端联调

### Phase 6:健壮性 + 打包(1 周)

6.1 ✅ CDP 断线重连
6.2 ✅ SQLite 持久化 + 任务恢复
6.3 ✅ Tauri 打包 Windows installer + macOS DMG
6.4 ✅ 代码签名(如果已有证书)

**总计:约 10-12 周**

## 关键风险

| 风险 | 影响 | 应对 |
|---|---|---|
| 豆包 macOS 版禁了 CDP | 🔴 整个 macOS 支持砍掉 | Phase 0 优先验证 |
| 字节升级豆包改了 image_ori_raw 字段名 | 🔴 图像捕获失败 | 多字段兼容 + 快速发版 |
| 字节改了视频解密算法 | 🔴 无水印视频失败 | 保留"下载按钮触发"路径作为 fallback |
| 字节加强风控识别 CDP 连接 | 🟡 工具被封 | 低频使用 + 账号专号专用 |
| 字节改了对话 UI(React 组件变更) | 🟡 fill/send 失败 | 多选择器兼容 + 版本自检 |

## 下一步

你定了:
- D1 账号调度:调用方指定 ✅
- D2 API 模式:异步 task_id + 轮询 ✅
- D3 平台:Windows + macOS ✅

**Phase 0 该做的事**:
1. 在你 Mac 上实测豆包 CDP 可行性(3 件事,20 分钟)
2. 结果出来后再启动 Phase 1

需要我帮你起 Phase 0 的实测清单吗?
