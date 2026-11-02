# AGENTS.md — doubao-relay

给 AI 编程助手(Claude Code、Codex 等)看的项目说明。用户文档见 `README.md`。

## 项目是什么

Tauri 2 桌面应用:通过 CDP(Chrome DevTools Protocol)操控本机**已登录的豆包桌面端**,把生图 / 生视频能力
暴露成本地 HTTP API(只监听 `127.0.0.1:8787`)。对外有两套接口:

- `/api/v1/...`:原始任务接口(异步 task_id + 轮询),规格见 `docs/spec/04-http-api.md`
- `/v1/...`:OpenAI 兼容接口(`src-tauri/src/http/routes/openai.rs`),下游主要用这套

所有沟通、日志 message、错误文案用**中文**;代码标识符、日志 key 用英文。

## 目录

```
src-tauri/src/
  cdp/            CDP WebSocket 会话、target 发现、大整数 JSON 修复
  platform/       Windows / macOS 适配(找豆包、带调试端口启动、杀进程、数据目录)
  doubao/         操控豆包:connection(会话/心跳/重连/切账号)、accounts、generation(提交)、
                  upload(参考图)、duration(视频时长滑块)、capture(捕获结果)、feedback、download、video_decoder
  doubao/scripts/ 注入页面的 JS:accounts.js(自写)+ 沈浪三段混淆脚本(generation-page / result-reader / video-decoder)
  tasks/          任务模型、校验、调度器(按账号串行)、runner、webhook
  storage/        SQLite(sqlx)+ 结果文件 + GC
  http/           axum 服务器、鉴权、路由(routes/openai.rs 为兼容接口)
  commands/       Tauri IPC 命令(snake_case 命令名,返回 camelCase)
frontend/src/     Preact + Signals 界面(账号 / 任务 / 日志 / 设置),纯函数在 view-model.ts
docs/spec/        产品规格(做什么)    docs/analysis/  逆向分析与真机探测记录(怎么做)
docs/analysis/10-live-probe-macos.md   每次真机探测的事实记录,改操控逻辑前先读
scripts/          cdp-probe.mjs(在真豆包里 eval JS)、bump-version.sh、package.sh、签名脚本
```

## 常用命令

```bash
cargo tauri dev --no-watch          # 开发运行(前端 vite :1477)。务必带 --no-watch,见下方"坑"
cargo fmt --all --check && cargo clippy --workspace --all-targets -- -D warnings && cargo test --workspace
(cd frontend && pnpm typecheck && pnpm test -- --run)
cargo test live_ -- --ignored --nocapture      # 需要真豆包的测试(只读或干跑,不发送)
node scripts/cdp-probe.mjs '<js 表达式>'        # 在豆包聊天页里求值,调试 selector 用
bash scripts/bump-version.sh X.Y.Z             # 同步版本号(6 处)
bash scripts/package.sh [mac|win|all]          # 本机打包:dist/ 下 dmg + Windows zip(cargo-xwin 交叉编译)
```

提交前必须通过 fmt / clippy(`-D warnings`)/ cargo test / pnpm typecheck / pnpm test。

## 约定

- 时间戳 Unix 毫秒 `i64`;ID 用 ULID;豆包的大整数 ID(conversation_id 等)在 Rust 里一律 `String`。
- 错误统一 `RelayError { code, message, details }`,code 取自 `docs/spec/11-error-codes.md`,HTTP 状态按前缀映射。
- `info` 日志不得出现 bearer token 和 prompt 原文(只记 `prompt_length` / `prompt_hash`)。
- 先写测试再实现;CDP 交互用 mock WS server 测(见 `capture.rs` / `generation.rs` / `upload.rs` 的测试工具)。
- 沈浪的三段 JS 是混淆代码,原则上不改;确需修补时做最小改动并在 `docs/analysis/10-live-probe-macos.md` 记录。
- 豆包 UI 改版导致的问题,先用 `cdp-probe.mjs` 只读探测拿到证据,再改代码,结论写进 10-live-probe 文档。

## 踩过的坑(改相关代码前必读)

- **`cargo tauri dev` 默认监视源码,一改代码就重编译并重启 relay,会打断用户正在跑的任务**。用 `--no-watch`。
- 开发端口是 **1477**(不是 Tauri 默认 1420),避免和本机其它 Tauri 项目抢端口导致窗口加载别人的前端。
- macOS 上监听 9707 的是 `Doubao.app/.../Doubao Browser` 子进程,不是 `Doubao` 主进程。
- 豆包会自动升级。2.32.6 的变化:参考图上传框要先点"+"(`upload_file_button`)→"添加图片 / 上传素材"才生成,
  且页面刷新后必须用 CDP `Input.dispatchMouseEvent` 可信点击;发送后可能弹"安全确认"(`safety_authorization_dialog`,
  自动点确认);视频技能会先回"请先确认参数",沈浪脚本识别不了,由 relay 自己读 `receive_message` 判断并代回复。
- 生成结果优先从生成请求的 SSE 响应里拿;拿不到时重放**页面自己发出的** `/im/chain/single`。
  绑定前记下的模板可能属于别的对话,绑定后要丢弃。
- 豆包大部分接口带 `a_bogus` / `msToken` 反爬签名。**不要自己合成或伪造签名请求**(试过,会被 403,且有风控风险)。
- 切换账号会整页 reload,所有 CDP 会话失效;单个豆包实例下任务实际只能串行(并发 1)。
- 新账号可能被豆包直接文字拒绝生图("我目前不能直接生成图片"),属账号额度问题,不是程序 bug。

## 不能做的事

- 不要在没有用户明确同意时发起真实生成(会消耗用户的豆包配额);测试用干跑(到 checkSend 为止)。
- 不要读取、导出豆包账号的 cookie / 登录凭证;不要 `pkill` 豆包;不要修改 `docs/spec/`。
- 不要把仓库改成公开:`docs/analysis/reference/` 和注入脚本是第三方(沈浪 Ai 画布)代码。
