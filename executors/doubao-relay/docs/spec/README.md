# doubao-relay · 开发规格说明(Spec)

> 本文档是 **doubao-relay 中转程序** 的产品 + 技术规格说明。写在编码之前,作为开发、测试、验收的参照基线。

## 快照

| 项 | 值 |
|---|---|
| 产品代号 | `doubao-relay` |
| 产品形态 | Tauri 桌面客户端(内置 HTTP API server) |
| 支持平台 | Windows 10/11 x64,macOS 12+ (Intel + Apple Silicon) |
| 核心能力 | 把本机豆包桌面端的生图/生视频能力通过 HTTP API 对外暴露 |
| 技术栈 | Rust 2021 + Tauri 2 + WebView(HTML/CSS/JS)+ SQLite |
| 对外接口 | HTTP (`http://127.0.0.1:<port>`) + Bearer Token 鉴权 |
| 调用方式 | 异步 `task_id` + 轮询 |
| 账号调度 | 调用方指定(API 必填 `account` 字段) |

## 范围声明

**本规格覆盖**:
- 产品定义、功能/非功能需求
- 对外 HTTP API 的完整契约
- WebView 与 Rust 之间的 Tauri IPC 命令
- 系统架构、模块划分、数据模型
- 任务生命周期、并发模型、错误码表
- Windows 和 macOS 平台适配层
- 安全模型、构建发布流程、验收标准

**本规格不覆盖**(放到后续设计文档或代码里):
- Rust 具体实现的内部数据结构
- CDP 协议报文的完整定义(引用 `shenlang-analysis`)
- 豆包页面 JS 注入代码的内部实现(直接复用 `shenlang-analysis` 中的 3 个 JS 文件)
- 单元测试用例、集成测试脚本

## 依赖的分析文档

本 spec 基于对沈浪 Ai 画布 v1.1.8 的逆向分析,核心发现见本项目内的 **`../analysis/`** 目录:

```
docs/analysis/
├── README.md
├── DECISIONS.md
├── 01-architecture.md
├── 02-cdp-bootstrap.md
├── 03-account-module.md
├── 04-generation-pipeline.md
├── 05-capture-pipeline.md
├── 06-video-decryption.md
├── 07-native-shadow.md
├── 08-anti-conflict.md
└── 09-reuse-for-tauri.md
```

**开发时参照**:spec 定"做什么",analysis 定"怎么做"。

## 其他项目内资源

- **`../analysis/scripts/`** — 3 个核心 JS,开发时 `cp` 到 `src-tauri/src/doubao/scripts/` 让 Rust 用 `include_str!` 嵌入
  - `doubao-generation-page.js`
  - `doubao-result-reader.js`
  - `doubao-video-decoder.js`
- **`../analysis/reference/`** — 沈浪原始 PS1 + tracker.exe,**仅作参考阅读**,不要在 Rust 项目里调用

## 文档导航

| # | 文件 | 内容 |
|---|---|---|
| 0 | **[00-product.md](./00-product.md)** | 产品定义、目标用户、使用场景、非目标 |
| 1 | **[01-functional.md](./01-functional.md)** | 功能需求清单(F1-F8) |
| 2 | **[02-non-functional.md](./02-non-functional.md)** | 性能预算、兼容性、可用性要求 |
| 3 | **[03-architecture.md](./03-architecture.md)** | 系统架构、模块划分、依赖关系图 |
| 4 | **[04-http-api.md](./04-http-api.md)** | 对外 HTTP API 完整规格(端点、请求、响应) |
| 5 | **[05-tauri-ipc.md](./05-tauri-ipc.md)** | Rust ↔ WebView 的 Tauri 命令契约 |
| 6 | **[06-webview-ui.md](./06-webview-ui.md)** | WebView UI 页面、组件、交互 |
| 7 | **[07-data-model.md](./07-data-model.md)** | SQLite schema、文件布局、运行时状态 |
| 8 | **[08-task-lifecycle.md](./08-task-lifecycle.md)** | 任务状态机 + 并发/排队模型 |
| 9 | **[09-cdp-contract.md](./09-cdp-contract.md)** | CDP 协议使用契约(基于 shenlang-analysis) |
| 10 | **[10-platform-adapter.md](./10-platform-adapter.md)** | Windows/macOS 平台适配层接口 |
| 11 | **[11-error-codes.md](./11-error-codes.md)** | 全局错误码表 |
| 12 | **[12-security.md](./12-security.md)** | 认证授权、敏感数据处理、威胁模型 |
| 13 | **[13-build-deploy.md](./13-build-deploy.md)** | 构建、打包、代码签名、发布流程 |
| 14 | **[14-acceptance.md](./14-acceptance.md)** | 验收标准 + 测试矩阵 |

## 版本

| 版本 | 日期 | 状态 | 说明 |
|---|---|---|---|
| 0.1 | 2026-10-16 | **草稿** | 初版,基于 Phase 0 验证结果 |

## 核心约定(全局生效)

1. **所有时间戳用 Unix epoch 毫秒(i64)**,API 和存储都统一。
2. **所有 ID 用 ULID(26 字符,时序可排)**,不用 UUID v4。
3. **错误用结构化 `{ error: { code, message, details? } }`**,见 `11-error-codes.md`。
4. **HTTP API 版本在 URL 里**:`/api/v1/...`,未来破坏性变更走 `/api/v2/`。
5. **所有 PII**(账号名、prompt 内容、生成结果)**只存本地**,不上报任何外部服务。
6. **豆包账号名作为唯一标识**,同名不同人属已知边界,v0.1 不处理。
7. **日志语言为中文**,方便用户和开发者快速定位。
