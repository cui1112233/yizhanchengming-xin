# 沈浪超级 Ai 画布 (独家版) 逆向分析

> 为 Tauri 中转程序的开发提供参考基础。文档只覆盖"豆包操控"核心能力,不涉及画布业务。

## 样本信息

| 项 | 值 |
|---|---|
| 文件名 | 沈浪超级Ai画布（独家版）.exe |
| 大小 | 114,363,463 字节 (109 MiB) |
| 类型 | NSIS-3 自解压安装器 → Electron 应用 |
| Electron 版本 | 44.5.1 |
| 内部名 | `shenlang-ai-canvas` v1.1.8 |
| 产品名 | 沈浪超级Ai画布 |
| 分析日期 | 2026-10-16 |
| 分析来源 | `/Users/james/Downloads/测试视频/沈浪超级Ai画布（独家版）.exe` |

## 一句话定性

**用 CDP 调试协议劫持用户本机已登录的豆包桌面端,用程序模拟 UI 操作在豆包聊天框里发 prompt 代用户生成图/视频,并从网络层旁录原图/原视频下载。整个过程完全模拟用户操作,不自己向字节服务器发任何业务请求。**

## 核心技术栈

| 层 | 用什么 |
|---|---|
| 启动豆包 | `--remote-debugging-port=9707 --remote-debugging-address=127.0.0.1` |
| 控制通道 | CDP (Chrome DevTools Protocol) over WebSocket |
| 操控驱动 | PowerShell 脚本 + 自定义 C# WebSocket Socket |
| 页面侧逻辑 | 3 个混淆 JS 文件 inject 到豆包页面(`Runtime.evaluate`) |
| 结果捕获 | `Network.enable` + 监听 `requestWillBeSent/responseReceived/loadingFinished/eventSourceMessageReceived` |
| 水印绕过 | ① 从响应 JSON 挖 `image_ori_raw.url` ② `lr=unwatermarked` 直链 ③ fallback API + 本地 AES-128-CBC 解密 token |
| 持久化 | `doubao-tracker.exe`(本地旁路进程,JSON over stdio) |

## 开发决策(2026-10-16 已对齐)

见 **[DECISIONS.md](./DECISIONS.md)**:
- **D1** 账号调度:调用方指定(API 带 `account` 字段)
- **D2** API 模式:异步 `task_id` + 轮询
- **D3** 平台:Windows + macOS 双平台(**Phase 0 先验证豆包 macOS 版 CDP 可行性**)

## 文档导航

0. **[DECISIONS.md](./DECISIONS.md)** — 开发决策记录 + 修订版开发路线
1. **[01-architecture.md](./01-architecture.md)** — 整体架构 + 文件职责表 + 数据流图
2. **[02-cdp-bootstrap.md](./02-cdp-bootstrap.md)** — 启动豆包 + 发现调试端口 + 握手流程
3. **[03-account-module.md](./03-account-module.md)** — 账号读取 / 切换(CDP DOM + Windows UIA 双通道)
4. **[04-generation-pipeline.md](./04-generation-pipeline.md)** — 发任务:选模型、填 prompt、点发送的完整序列
5. **[05-capture-pipeline.md](./05-capture-pipeline.md)** — 结果捕获:网络监听 + 多轮确认 + 水印绕过
6. **[06-video-decryption.md](./06-video-decryption.md)** — 视频无水印 token 本地 AES-CBC 解密全算法
7. **[07-native-shadow.md](./07-native-shadow.md)** — 影子追踪进程 doubao-tracker.exe
8. **[08-anti-conflict.md](./08-anti-conflict.md)** — 工具跑任务时用户同时操作豆包的冲突面
9. **[09-reuse-for-tauri.md](./09-reuse-for-tauri.md)** — Tauri 中转程序开发的可复用清单 + 决策建议

## 本地样本参考路径

**原始解压路径**(临时):`/private/tmp/claude-501/.../scratchpad/shenlang_extract/asar_out/desktop/`

**项目内已拷贝的参考副本**:`./reference/`

```
docs/analysis/reference/
├── native/                              # 7 个明文 PS1 / JS 文件(沈浪 Windows 原始实现)
│   ├── doubao-accounts.ps1
│   ├── doubao-accounts-background.ps1
│   ├── doubao-capture.ps1
│   ├── doubao-generate.ps1
│   ├── doubao-generation-page.js
│   ├── doubao-result-reader.js
│   └── doubao-video-decoder.js
└── native-shadow-bin/
    ├── doubao-tracker.exe               # SHA256: 90fe9cd4...37e2be
    ├── enabled                          # 开关 flag,内容 "enabled\r\n"
    └── delivery-gate-enabled            # 内容 "enabled\r\n"
```

**另外**:`./scripts/` 下有 3 个 JS(与 native/ 下的同名文件内容一致),
这是 Rust 实现要通过 `include_str!` 嵌入到二进制中的工作副本。

分析本 README 时未搬到项目内的沈浪文件(保留供回查原始解压目录):
- `main.cjs` / `preload.cjs` / `renderer.js` — javascript-obfuscator 混淆的 Electron 代码
- `release-integrity.json`
- `third-party/doubao-assistant-LICENSE.txt`

## 后续开发前要知道的关键结论

- **豆包账号不需要密码**。工具靠你本机已登录的豆包桌面端跑,它只是重启豆包时加上 `--remote-debugging-port` 参数来打开 CDP 后门。
- **豆包桌面端只有 Windows / macOS 版本**,其中 Windows 版才是本工具的主要支持平台(PS1 全是 Windows API)。
- **所有真实网络请求都是豆包自己发的**,中转程序只做"操控 + 旁听"。
- **多轮确认已经处理过**(2 次、3 次确认都能自动继续),但遇到弹窗/验证码会挂。
- **视频解密算法可能跟豆包版本绑定**(密钥种子和盐都是从豆包页面响应里提取的),豆包升级后要重新对盐。

## 适用边界

该工具属于典型"**用户账号代操作**"产品:违反豆包用户协议,被封号责任在账号持有人。文档仅做技术记录,不建议作为商业产品投产。
