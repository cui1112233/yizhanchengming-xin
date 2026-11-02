# 01 · 整体架构

## 安装包三层拆解

```
沈浪超级Ai画布（独家版）.exe   (NSIS-3 自解压 110 MiB)
   │
   └─ $PLUGINSDIR/app-64.7z                       (113 MiB,7z 压缩)
        │
        └─ 解压后 ≈ 380 MiB,就是一个 Electron 应用:
             ├── 沈浪超级Ai画布.exe                 (234 MiB,Electron shell + Chromium)
             ├── chrome_100_percent.pak            (Chromium 资源)
             ├── ffmpeg.dll / icudtl.dat / ...     (Chromium 运行时)
             ├── locales/                          (多语言包)
             └── resources/app.asar                (11.6 MiB,业务代码)
                   │
                   └─ asar 解包后 → desktop/*.cjs/*.js  (本工具真正的核心)
```

关键观察:外层是普通 Electron 打包流程(electron-builder + NSIS),没有做任何加壳或反调试。所以用标准工具链(7-Zip + `@electron/asar` extract)就能完整拆到源代码。

## package.json

```json
{
  "name": "shenlang-ai-canvas",
  "version": "1.1.8",
  "private": true,
  "productName": "沈浪超级Ai画布",
  "main": "desktop/main.cjs",
  "dependencies": {
    "electron-updater": "6.6.2",
    "mammoth": "1.11.0",
    "word-extractor": "1.0.4"
  }
}
```

- 主入口 `desktop/main.cjs`(99 KB,javascript-obfuscator 混淆)。
- 只有 3 个运行时依赖,其中 `mammoth` 和 `word-extractor` 是读 Word 文档用的,与"操控豆包"无关 —— 画布功能的一部分,可以忽略。
- `electron-updater` 自带升级能力(走沈浪自己的 feed URL,未分析)。

## desktop/ 目录按功能分组

### 画布业务层(与豆包操控无关,可以忽略)

```
canvas-*.js                 画布节点、连线、分组
asset-*.cjs                 素材管理
share-*.{js,cjs}            分享功能
cloud-canvas-client.cjs     云同步
project-store.cjs           项目持久化
voice-*.js                  语音处理
```

### 豆包操控层(本次关心的全部)

```
main.cjs                                  (混淆) Electron 主进程,所有 IPC 入口
preload.cjs                               (混淆) 暴露 window.canvas* 系列 IPC
model-connection.cjs / model-network.cjs  (混淆) 模型适配器
doubao-connection.cjs                     (混淆) 豆包连接管理
doubao-generation-input.cjs               (混淆) 组装 PS 子进程的 request.json
doubao-local.cjs                          (混淆) 豆包本地任务调度
doubao-model-access.cjs                   (混淆) 模型权限检查
doubao-accounts.cjs                       (混淆) 账号 IPC 服务端
doubao-native-adapter.cjs                 (混淆) 调起 native/*.ps1 的适配层
doubao-native-shadow.cjs                  (混淆) 启动/喂养 doubao-tracker.exe
doubao-subscription.js                    (混淆) 豆包会员订阅检查

native/                                   明文,PS1 + JS,真实"干活"的脚本
├── doubao-accounts.ps1                   账号模块入口分发器
├── doubao-accounts-background.ps1        账号模块真正干活(新路径)
├── doubao-generate.ps1                   发任务脚本
├── doubao-capture.ps1                    捕获结果脚本(整个工具的大脑)
├── doubao-generation-page.js             注入豆包页面:选模型/填 prompt/点发送
├── doubao-result-reader.js               注入豆包页面:从消息 JSON 挖图/视频/反馈
└── doubao-video-decoder.js               注入豆包页面:AES-CBC 解视频 token

native-shadow-bin/
├── doubao-tracker.exe                    PE32+ x86-64,仅导入 KERNEL32,870 函数
├── enabled                               开关文件,内容 "enabled\r\n"
└── delivery-gate-enabled                 投递门控,内容 "enabled\r\n"
```

## doubao-* 文件职责一览

### PowerShell 脚本

| 文件 | 执行环境 | 职责 |
|---|---|---|
| `doubao-accounts.ps1` | Windows PowerShell 子进程 | 入口分发器。`operation=read/connect/switch` 时直接转交 `doubao-accounts-background.ps1`。预留了"原生鼠标点击"的老路径(新代码不走到) |
| `doubao-accounts-background.ps1` | Windows PowerShell 子进程 | 账号读取与切换的实际实现。CDP+UIA 双通道。1300 字节脚本,是账号模块的真正本体 |
| `doubao-generate.ps1` | Windows PowerShell 子进程 | 把一次生成请求映射到豆包页面上。流程:prepare → durationDragPlan → fileInput → fill → checkSend → send |
| `doubao-capture.ps1` | Windows PowerShell 子进程 | 整个工具的大脑。调起 `doubao-generate.ps1`,然后开 CDP Network 监听,等结果,下载图/视频 |

### 页面注入 JS(都在豆包页面上下文执行)

| 文件 | 导出的对象 | 职责 |
|---|---|---|
| `doubao-generation-page.js` | `shenlangDoubaoGeneration` | 认识豆包页面 DOM,知道怎么点下拉菜单选模型、拖滑动条设时长、填 prompt、点发送 |
| `doubao-result-reader.js` | `shenlangDoubaoResultReader` | 从豆包的 `/im/chain/single` 响应 JSON 里挖:无水印原图 URL、视频 URL、反馈消息、submissionId 绑定 |
| `doubao-video-decoder.js` | `shenlangDoubaoVideoDecoder` | 对无水印视频下载 token 做本地 AES-128-CBC 解密 |

### 原生二进制

| 文件 | 架构 | 职责 |
|---|---|---|
| `doubao-tracker.exe` | PE32+ x86-64,仅 KERNEL32 导入 | 本地旁路守护进程,通过 stdin/stdout JSON 协议和主进程通信,负责把"任务 ID → 结果"的映射持久化到磁盘。开关是 `native-shadow-bin/enabled` 文件 |

## 父子进程调用链

```
用户在画布上点"生成图/视频"
    │
    ▼
Electron 主进程 (main.cjs,混淆)
    │
    │  ① spawn child_process: 账号就位
    ▼
PowerShell: doubao-accounts.ps1  → doubao-accounts-background.ps1
    │
    │  CDP + UIA
    ▼
豆包桌面端(此时已启动,CDP 端口 9707 开着)
    │  读取账号列表 / 切换到目标账号
    │
    ◄── 返回: 账号列表 JSON (写文件 OutputFile)
    │
Electron 主进程 读 OutputFile
    │
    │  ② spawn child_process: 捕获大脑
    ▼
PowerShell: doubao-capture.ps1
    │
    │  调 Invoke-RestMethod http://127.0.0.1:9707/json/list
    │  WebSocket 连到 doubao-chat / doubao-background 两个页面
    │  Network.enable 开启网络监听
    │
    │  ③ 从 capture.ps1 内部 spawn 发任务脚本
    ▼
    PowerShell: doubao-generate.ps1
        │
        │  在 doubao-chat 页面 Runtime.evaluate
        │  注入 doubao-generation-page.js
        │  调 shenlangDoubaoGeneration.run('prepare')
        │        shenlangDoubaoGeneration.run('fill')
        │        shenlangDoubaoGeneration.run('send')
        │
        │  用 execCommand('insertText') + PointerEvent 链操作豆包
        │
        │  中间可能用 Windows UIAutomation 操作原生对话框
        │  (选模型菜单 / 文件选择对话框)
        ▼
    豆包页面  → POST www.doubao.com/im/message/send
             → 字节服务器开始生成(用你本机登录的账号配额)
             → SSE 流式返回 → 页面显示
    │
    │  (并行) capture.ps1 的 while 循环持续监听
    │   - Network.eventSourceMessageReceived (SSE 流)
    │   - Network.responseReceived (video/*)
    │   - 每秒重放 /im/chain/single 请求读历史
    │   - 调 shenlangDoubaoResultReader.read() 挖结果字段
    │   - 多轮确认: 识别"是否继续生成"→ 自动发"继续生成"
    ▼
    捕获到无水印 URL → Invoke-WebRequest 下载到 OutputDirectory/images/ 或 /videos/
    │
    │  (可选) 每条结果通知 nativeShadow
    │     [Console]::Out.WriteLine('DOUBAO_NATIVE_SHADOW: {...}')
    │                 │
    │                 ▼
    │             Electron 主进程 转发给 doubao-tracker.exe (stdio)
    │                 │
    │                 ▼
    │             落盘持久化
    ▼
Electron 主进程读 doubao-result-capture-report.json
    │
    ▼
画布 UI 显示生成结果
```

## 控制台协议(stdout 关键字)

`doubao-capture.ps1` 用 `[Console]::Out.WriteLine()` 向父进程推进度,格式是 `<KEY>:<JSON>` 单行:

| 关键字 | 含义 |
|---|---|
| `DOUBAO_READY` | CDP 已就绪 |
| `DOUBAO_STAGE:mode` | 正在选模型 |
| `DOUBAO_STAGE:upload` | 正在上传参考图 |
| `DOUBAO_UPLOAD:n/total` | 第 n 张参考图上传进度 |
| `DOUBAO_STAGE:prompt` | 正在填 prompt |
| `DOUBAO_STAGE:sending` | 正在点发送 |
| `DOUBAO_STAGE:sent` | 已发送 |
| `DOUBAO_STAGE:checking` | 豆包进入参数确认阶段 |
| `DOUBAO_STAGE:quota_checking` | 豆包进入会员配额确认 |
| `DOUBAO_STAGE:generating` | 豆包真正开始生成 |
| `DOUBAO_STAGE:result` | 结果已返回 |
| `DOUBAO_STAGE:downloading` | 正在下载原图/原视频 |
| `DOUBAO_STAGE:resent` | 已自动点"继续生成" |
| `DOUBAO_FEEDBACK:{...}` | 豆包回复的文字 |
| `DOUBAO_BOUND:{conversationId,submissionId}` | 任务已绑定会话 ID |
| `DOUBAO_BACKGROUND:{...}` | 切换到后台追踪模式 |
| `DOUBAO_IMAGES:n` | 共 n 张原图待下载 |
| `DOUBAO_WAIT:{reason,message,...}` | 需要用户介入(超时/不确定答复) |
| `DOUBAO_AUTO_MANUAL` | 自动确认失败,降级为手动 |
| `DOUBAO_MEMBERSHIP_MANUAL` | 会员配额自动失败,降级为手动 |
| `DOUBAO_REFUSED:{message,kind}` | 豆包拒绝生成(配额用完/违规等) |
| `DOUBAO_NEEDS_PAGE` | 需要用户切回到绑定的 chat 页 |
| `DOUBAO_NATIVE_SHADOW:{...}` | 向 shadow 进程推送任务状态 |

这个协议对设计 Tauri 的 Rust ↔ WebView 事件通道有参考价值 —— **按阶段推送**而不是请求-响应。

## release-integrity.json

```json
{
  "version": 1,
  "files": {
    "native/doubao-accounts-background.ps1": "58acbc303b3e18fb...4cd4c42",
    "native/doubao-accounts.ps1":            "e5cff41f3165d394...bc97f1d",
    "native/doubao-capture.ps1":             "5053abc1b4af4f55...6b7a4cc",
    "native/doubao-generate.ps1":            "183fea7bac75c818...bc807e80",
    "native/doubao-generation-page.js":      "a7f7226ee6bd729e...4312799f",
    "native/doubao-result-reader.js":        "7bb4f5e446574569...65c5e12",
    "native/doubao-video-decoder.js":        "d2d6aaf1562c0962...1a55ad0813a",
    "native-shadow-bin/doubao-tracker.exe":  "90fe9cd4c627da32...39a37e2be",
    "native-shadow-bin/enabled":             "fe096a06d67909...a80852c8"
  }
}
```

沈浪自己在启动时校验这些文件的 SHA256,防止被第三方替换。你如果要修改 native 下任何文件做调试,得同步改这个清单。
