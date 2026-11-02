# doubao-relay 全量实现计划(Phase 1–6)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 从零实现 doubao-relay——一个 Tauri 2 桌面应用,通过 CDP 操控本机豆包桌面端,对外暴露 `http://127.0.0.1:8787/api/v1` 生图/生视频 HTTP API,覆盖 spec 定义的全部功能(账号管理、生图、生视频无水印、多轮确认、并发队列、持久化、打包)。

**Architecture:** Rust 后端分层:`http`(axum) → `commands`(Tauri IPC) → 核心服务(`doubao::*`、`tasks::*`、`storage`) → `cdp`(tokio-tungstenite 薄封装) / `platform`(Win/Mac trait)。三段沈浪注入 JS 以 `include_str!` 嵌入,经 `Runtime.evaluate` 注入豆包 `doubao-chat` 页面。前端 Preact + Signals,只经 Tauri `invoke` 访问后端。

**Tech Stack:** Rust 2021 (1.94) · Tauri 2.12 · tokio · tokio-tungstenite 0.30 · axum 0.8 · sqlx 0.8 (sqlite, runtime-tokio) · reqwest 0.12 (rustls) · aes/cbc/sha2/hmac (RustCrypto) · ulid · tracing · Preact 10 + @preact/signals · Vite 6 · TypeScript 5 · vitest · pnpm 9

**Spec:** `docs/spec/README.md`(00–14 全部)、`docs/analysis/DECISIONS.md`、`docs/analysis/01–09`。本计划"做什么"以 spec 为准,"怎么做"以 analysis 为准;冲突时 spec 优先,并在计划里标注。

## Global Constraints(来自 spec,逐条照抄)

- 平台:Windows 10/11 x64,macOS 12+(Intel + Apple Silicon)。开发机是 macOS,Windows 代码只能 `cargo check --target x86_64-pc-windows-msvc`。
- 时间戳一律 Unix epoch 毫秒 `i64`;ID 一律 ULID(26 字符);HTTP 路径 `/api/v1/...`;HTTP JSON 字段 `snake_case`;Tauri IPC 参数/返回 `camelCase`。
- 错误统一 `{ error: { code, message, details? } }`,code 取自 `docs/spec/11-error-codes.md`;message 中文、不带标点、附一句下一步建议。
- 日志语言中文(message),日志 key 英文;`info` 级别不得出现 bearer_token,不得出现 prompt 内容(只 `prompt_length` + `prompt_hash`)。
- HTTP server 只允许绑定 `127.0.0.1` / `::1` / `localhost`;`/api/v1/health` 免鉴权,其余 Bearer 鉴权,常量时间比较。
- 并发默认:`global_max=3`,`per_account_max=1`(硬),`queue_depth_max=50`。超时:preparing 60s;generating 图 5min / 视频 20min;waiting_user_confirm 1h;pending 1h。
- 结果文件保留 24h,元数据 7d,upload 1h;GC 每 1h。
- CDP:命令默认 15s 超时,`Runtime.evaluate` 30s;心跳 `Target.getTargets` 每 10s(UC-06 要求 15s 内发现断线),5s 无响应判断开;重连 20s 失败升级重启豆包。
- 豆包 target 白名单:`type=="page"` 且 url scheme ∈ {doubao, chrome} 且 host ∈ {doubao-chat, doubao-background};ws url 必须 loopback + 配置端口;**9707 监听进程的可执行路径必须位于豆包安装目录下**(macOS 实测监听者是 `Doubao.app/Contents/Helpers/Doubao Browser.app/.../Doubao Browser`,不是 `Doubao` 主进程)。
- 跨平台账号 selector:**不依赖 `aria-label` 和 `img[alt]`**(macOS 为空),按 `button[aria-haspopup="menu"]` + 含 `<img>` + 左下角位置定位;账号文本同时支持 `"张三 体验套餐"`(单行)与 `"张三\n体验套餐"`(两行)。
- 关闭 relay 不杀豆包。豆包崩溃时 relay 不崩,进行中任务 `failed/doubao_crashed`,新请求 503。
- `cargo clippy -- -D warnings` 零警告;`cargo fmt --check` 通过;`pnpm typecheck` 零错误。
- 所有数据只存本地,工具自身不发任何对外 HTTP(webhook 与 opt-in 检查更新除外)。

## Review Focus(spec 隐含但无测试覆盖的 5 个最可能翻车点)

1. **沈浪 JS 的"对话模式 vs 工作模式"**:`doubao-generation-page.js` 内含字符串 `对话模式仅支持Mini的11～15秒任务`,说明 `submissionMode='conversation'` 下视频只能 Seedance 2.0 Mini 11–15s;其它模型/时长可能必须走 `work` 模式(豆包"AI 创作"页)。→ Task 2.1 必须先在真豆包上探明两种模式的页面入口与 `prepare` 行为,结论写入 `docs/analysis/10-live-probe-macos.md`,之后 Task 2.3 据此决定 `submissionMode` 取值。
2. **账号菜单在 macOS 的 DOM 结构**:Phase 0 只验证了账号按钮,没验证展开后的 `[role=menuitem]` 结构、"更多账号"子菜单、"退出登录"项是否存在。→ Task 1.2 的 live probe 必须逐条验证 `accounts.js` 每个方法在 2.32.5 上的返回,失败的 selector 当场改。
3. **切换账号后页面 reload,所有 CdpSession 作废**:若 `switch_account` 的验证循环不触发 `reconnect_sessions`,会在 20s 后误报 `account_switch_timeout`。→ Task 1.8 测试用 mock WS server 主动断开,断言 `DoubaoConnection::chat_session()` 在重连后返回新 session。
4. **大整数精度**:`conversation_id/message_id/local_message_id` 超过 2^53,`serde_json::Value` 读成 `Number` 后 `as_u64` 溢出或 `f64` 丢精度。→ Task 2.2 `json_fix::fix_big_integers` 单测覆盖 17 位、负数、嵌套、已经是字符串的情况;所有 ID 在 Rust 结构里是 `String`。
5. **结果下载 URL 无扩展名、MIME 由内容判定**:`image_ori_raw.url` 可能返回 JPEG 而文件名写 `.png`。→ Task 2.5 `download::detect_image_format` 单测:PNG 魔数、JPEG 魔数、WebP、非图片字节 → `Err(generation_download_failed)`。

---

## 代码布局(最终形态)

```
doubao-relay/
├── Cargo.toml                       workspace, members = ["src-tauri"]
├── package.json                     根编排脚本(dev/build/typecheck/test)
├── .gitignore
├── scripts/
│   ├── cdp-probe.mjs                开发工具:在真豆包里 eval JS(Node 25 自带 WebSocket)
│   └── gen-icon.py                  生成占位图标 PNG(python3 stdlib)
├── migrations/
│   └── 20261016000001_initial.sql   tasks / task_stage_history / uploads / webhook_log / metadata
├── src-tauri/
│   ├── Cargo.toml  build.rs  tauri.conf.json  capabilities/default.json  icons/
│   └── src/
│       ├── main.rs                  fn main() { doubao_relay_lib::run() }
│       ├── lib.rs                   run(): 读配置→日志→Storage→AppState→HTTP server→后台任务→Tauri Builder
│       ├── state.rs                 AppState { config, platform, doubao, storage, scheduler, started_at }
│       ├── error.rs                 RelayError(枚举,code/message/details/http_status) + IpcError + 转换
│       ├── config.rs                Config(TOML,见 07-data-model)+ load/save/默认 token 生成
│       ├── logging.rs               tracing: stdout + 按天滚动文件 + 内存环形缓冲(供 log.recent)
│       ├── commands/{mod,app,account,doubao,task,upload,log}.rs   #[tauri::command]
│       ├── cdp/{mod,session,discovery,json_fix}.rs
│       ├── platform/{mod,macos,windows}.rs
│       ├── doubao/
│       │   ├── mod.rs  bootstrap.rs  connection.rs  compat.rs  accounts.rs
│       │   ├── generation.rs  upload.rs  duration.rs  capture.rs  feedback.rs  download.rs  video_decoder.rs
│       │   └── scripts/accounts.js + doubao-generation-page.js + doubao-result-reader.js + doubao-video-decoder.js
│       ├── storage/{mod,db,files,gc}.rs
│       ├── tasks/{mod,model,detail,validate,scheduler,runner,webhook}.rs
│       └── http/{mod,auth,error}.rs + http/routes/{health,accounts,uploads,generate,tasks,doubao}.rs
├── frontend/
│   ├── package.json  vite.config.ts  tsconfig.json  index.html
│   └── src/
│       ├── main.tsx  app.tsx  api.ts  state.ts  style.css  view-model.ts
│       ├── components/{sidebar,status-bar,toast,confirm-dialog,new-task-dialog}.tsx
│       └── pages/{accounts,tasks,task-detail,logs,settings}.tsx
└── docs/analysis/10-live-probe-macos.md   各 Phase 真豆包探测结论(agent 产出)
```

## 共享约定(所有 Task 遵守)

- **Tauri 命令名**:spec 写 `account.list`,Rust 标识符不能含 `.`,统一用 `account_list`;前端 `api.ts` 封装名字。
- **错误**:后端所有 `Result<T, RelayError>`;`RelayError` 实现 `Serialize` 为 `{code,message,details}`,`impl From<RelayError> for IpcError`(同结构)和 axum `IntoResponse`(按 11-error-codes 的 HTTP 状态码表)。
- **Rust 结构体**:对外(HTTP)用 `#[serde(rename_all = "snake_case")]`,对 WebView 用 `camelCase`——同一个 domain 结构用两套 DTO 太重,规定:domain 类型用 `snake_case`,HTTP 直接返回;IPC 层用 `serde_json::Value` 转换函数 `to_camel(value)`(递归改 key)返回。`to_camel` 在 `commands/mod.rs`,单测覆盖嵌套数组。
- **测试**:纯函数单测放同文件 `#[cfg(test)]`;需要网络的用本地 mock(`tokio::net::TcpListener` + `tokio_tungstenite::accept_async` / axum test server);需要真豆包的测试标 `#[ignore]`,名字以 `live_` 开头,执行 `cargo test live_ -- --ignored --nocapture`。
- **提交**:每个 Task 一个或多个 commit,message 格式 `feat(scope): ...` / `test(scope): ...` / `chore: ...`,末尾加 `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`。
- **验证命令**(每个 Task 结束必跑):`cargo fmt --check && cargo clippy --workspace --all-targets -- -D warnings && cargo test --workspace && (cd frontend && pnpm typecheck && pnpm test -- --run)`。

---

# Phase 1 · 最小闭环(骨架 + CDP + 账号)

验收:`docs/spec/14-acceptance.md` UC-01 ~ UC-04(P0)、UC-05 ~ UC-06(P1)。

### Task 1.1: 仓库初始化 + Tauri/前端骨架

**Files:**
- Create: `.gitignore`, `Cargo.toml`, `package.json`, `src-tauri/Cargo.toml`, `src-tauri/build.rs`, `src-tauri/tauri.conf.json`, `src-tauri/capabilities/default.json`, `src-tauri/src/main.rs`, `src-tauri/src/lib.rs`, `scripts/gen-icon.py`, `frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/index.html`, `frontend/src/main.tsx`, `frontend/src/app.tsx`, `frontend/src/style.css`, `frontend/src/view-model.ts`, `frontend/src/view-model.test.ts`

**Interfaces:**
- Produces: 可运行的 `cargo tauri dev` 空壳窗口,标题 `doubao-relay`,1000×680,最小 900×600;`pnpm typecheck` / `pnpm test` 可跑。

- [ ] **Step 1: git init + .gitignore**

```gitignore
target/
node_modules/
dist/
src-tauri/gen/
*.exe
.DS_Store
```

`git init && git add -A && git commit -m "chore: import spec and analysis docs"`。

- [ ] **Step 2: 根 Cargo.toml / package.json**

```toml
# Cargo.toml
[workspace]
members = ["src-tauri"]
resolver = "2"
```

```json
{ "name": "doubao-relay-root", "private": true, "version": "0.1.0",
  "scripts": {
    "dev": "cargo tauri dev", "build": "cargo tauri build",
    "typecheck": "pnpm --dir frontend typecheck", "test": "pnpm --dir frontend test -- --run" } }
```

- [ ] **Step 3: src-tauri/Cargo.toml(Phase 1 依赖,后续 Task 再追加)**

```toml
[package]
name = "doubao-relay"
version = "0.1.0"
edition = "2021"

[lib]
name = "doubao_relay_lib"
crate-type = ["staticlib", "cdylib", "rlib"]

[build-dependencies]
tauri-build = { version = "2", features = [] }

[dependencies]
tauri = { version = "2", features = [] }
tokio = { version = "1", features = ["full"] }
tokio-tungstenite = "0.30"
futures-util = "0.3"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
reqwest = { version = "0.12", default-features = false, features = ["json", "rustls-tls"] }
url = "2"
sysinfo = "0.39"
async-trait = "0.1"
thiserror = "2"
anyhow = "1"
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter", "json"] }
tracing-appender = "0.2"
toml = "0.9"
regex = "1"
rand = "0.10"
dirs = "6"

[target.'cfg(windows)'.dependencies]
winreg = "0.56"
```

- [ ] **Step 4: build.rs / main.rs / lib.rs / tauri.conf.json / capabilities**

```rust
// build.rs
fn main() { tauri_build::build() }
```
```rust
// src/main.rs
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]
fn main() { doubao_relay_lib::run() }
```
```rust
// src/lib.rs(Phase 1 起点,后续 Task 扩展)
pub fn run() {
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("doubao-relay 启动失败");
}
```
```json
// tauri.conf.json
{
  "$schema": "https://schema.tauri.app/config/2",
  "productName": "doubao-relay",
  "version": "0.1.0",
  "identifier": "com.bwy.doubaorelay",
  "build": {
    "frontendDist": "../frontend/dist",
    "devUrl": "http://localhost:1420",
    "beforeDevCommand": "pnpm --dir frontend dev",
    "beforeBuildCommand": "pnpm --dir frontend build"
  },
  "app": {
    "windows": [{ "title": "doubao-relay", "width": 1000, "height": 680, "minWidth": 900, "minHeight": 600 }],
    "security": { "csp": null }
  },
  "bundle": { "active": true, "targets": "all", "icon": ["icons/icon.icns", "icons/icon.ico", "icons/icon.png"] }
}
```
```json
// capabilities/default.json
{ "identifier": "default", "windows": ["main"], "permissions": ["core:default"] }
```

- [ ] **Step 5: 图标** — `scripts/gen-icon.py` 用 `zlib`+`struct` 写一张 1024×1024 纯色(#3B82F6)PNG 到 `scripts/app-icon.png`,然后 `cargo tauri icon scripts/app-icon.png -o src-tauri/icons`。

- [ ] **Step 6: 前端骨架(Preact)**

`frontend/package.json` devDependencies: `vite@^6`, `typescript@^5`, `@preact/preset-vite@^2`, `vitest@^3`; dependencies: `preact@^10`, `@preact/signals@^2`, `@tauri-apps/api@^2`。scripts: `dev: vite`, `build: tsc --noEmit && vite build`, `typecheck: tsc --noEmit`, `test: vitest`。

`vite.config.ts`: `plugins:[preact()], clearScreen:false, server:{port:1420,strictPort:true}, envPrefix:['VITE_','TAURI_']`。
`tsconfig.json`: `"jsx":"react-jsx","jsxImportSource":"preact","strict":true,"module":"ESNext","moduleResolution":"Bundler","target":"ES2022","types":["vite/client"]`。

`view-model.ts` 第一个纯函数(先写测试):
```ts
export type DoubaoDot = 'green' | 'red' | 'yellow';
export function doubaoDot(s: { running: boolean; cdpConnected: boolean; reconnecting?: boolean }): DoubaoDot {
  if (s.reconnecting) return 'yellow';
  return s.running && s.cdpConnected ? 'green' : 'red';
}
```
`view-model.test.ts`: 三种输入断言三种颜色。

`app.tsx`: 左 180px 侧栏(账号/任务/日志/设置 4 项 + 底部 ●豆包 ●API 状态)+ 右主区占位文本"账号页(Phase 1)"。`style.css` 用 CSS 变量 + `prefers-color-scheme` 深浅主题。

- [ ] **Step 7: 验证** — `cd frontend && pnpm install && pnpm typecheck && pnpm test -- --run`;`cargo tauri dev` 启动 20s 后 `screencapture -x /tmp/relay-1-1.png`,Read 确认窗口出现并有侧栏。退出。
- [ ] **Step 8: Commit** `feat: tauri + preact skeleton`

### Task 1.2: CDP 探针脚本 + 账号 JS 在真豆包上验证

**Files:**
- Create: `scripts/cdp-probe.mjs`, `src-tauri/src/doubao/scripts/accounts.js`, `docs/analysis/10-live-probe-macos.md`
- Copy: `docs/analysis/scripts/*.js` → `src-tauri/src/doubao/scripts/`(原样复制,不改)

**Interfaces:**
- Produces: `window.__doubaoRelayAccounts` 对象,方法与返回值见下;`node scripts/cdp-probe.mjs [--target chat|background] '<expr>'` 打印 JSON。

- [ ] **Step 1: cdp-probe.mjs**

```js
// 用法: node scripts/cdp-probe.mjs [--port 9707] [--target chat|background] [--file path.js] '<expression>'
// --file 先把文件内容注入(evaluate),再 evaluate 表达式。
const args = process.argv.slice(2);
const opt = (k, d) => { const i = args.indexOf(k); return i >= 0 ? args.splice(i, 2)[1] : d; };
const port = opt('--port', '9707'), target = opt('--target', 'chat'), file = opt('--file', null);
const expr = args[0] ?? 'document.title';
const list = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
const page = list.find(t => t.type === 'page' && new URL(t.url).host === `doubao-${target}`);
if (!page) { console.error('target not found', list.map(t => t.url)); process.exit(1); }
const ws = new WebSocket(page.webSocketDebuggerUrl);
let id = 0; const pending = new Map();
const send = (method, params) => new Promise((res, rej) => { const i = ++id; pending.set(i, { res, rej }); ws.send(JSON.stringify({ id: i, method, params })); });
ws.onmessage = e => { const m = JSON.parse(e.data); if (m.id && pending.has(m.id)) { const p = pending.get(m.id); pending.delete(m.id); m.error ? p.rej(m.error) : p.res(m.result); } };
await new Promise(r => ws.onopen = r);
const evaluate = async expression => { const r = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true }); if (r.exceptionDetails) throw new Error(r.exceptionDetails.exception?.description ?? 'js exception'); return r.result.value; };
if (file) await evaluate(await (await import('node:fs/promises')).readFile(file, 'utf8'));
console.log(JSON.stringify(await evaluate(expr), null, 2));
ws.close();
```

- [ ] **Step 2: accounts.js(跨平台,幂等注入)**

```js
// 注入 doubao-chat 页面的账号操作脚本,暴露 window.__doubaoRelayAccounts。
// 约束见 docs/analysis/DECISIONS.md "Phase 0 验证结果":不依赖 aria-label / img[alt]。
(() => {
  if (window.__doubaoRelayAccounts) return;
  const SUB_RE = /^[一-龥A-Za-z0-9]{1,24}套餐$/;
  const LOGIN_RE = /^(登录|立即登录|登录豆包)$/;
  const visible = el => { const r = el.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
  const text = el => (el.innerText || el.textContent || '').trim();
  const lines = s => String(s || '').split(/\r?\n/).map(x => x.trim()).filter(Boolean);
  function parseLabel(raw) {
    const ls = lines(raw);
    if (!ls.length) return { name: '', subscription: '' };
    const sub = ls.slice(1).find(l => SUB_RE.test(l));
    if (sub) return { name: ls[0], subscription: sub };
    const m = ls[0].match(/^(.+?)\s+([一-龥A-Za-z0-9]{1,24}套餐)$/);
    return m ? { name: m[1].trim(), subscription: m[2] } : { name: ls[0], subscription: '' };
  }
  function allButtons() { return [...document.querySelectorAll('button,[role="button"]')].filter(visible); }
  function findAccountButton() {
    const cands = allButtons().filter(el => el.getAttribute('aria-haspopup') === 'menu' && !el.closest('[role="menu"]') && !!el.querySelector('img'));
    const positioned = cands.filter(el => { const r = el.getBoundingClientRect();
      return r.left >= 0 && r.left < 80 && r.top > innerHeight - 120 && r.bottom <= innerHeight + 2 && r.width >= 35 && r.width <= 360 && r.height >= 20 && r.height <= 85; });
    return positioned.length === 1 ? positioned[0] : (cands.length === 1 ? cands[0] : null);
  }
  function pointer(el, type, buttons) {
    const r = el.getBoundingClientRect();
    el.dispatchEvent(new PointerEvent(type, { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', isPrimary: true, button: 0, buttons, clientX: r.x + r.width / 2, clientY: r.y + r.height / 2 }));
  }
  const loginVisible = () => allButtons().some(el => LOGIN_RE.test(text(el)));
  const menuItems = () => [...document.querySelectorAll('[role="menuitem"]')].filter(visible);
  window.__doubaoRelayAccounts = {
    probe() { const b = findAccountButton(); return { buttonFound: !!b, buttonText: b ? text(b) : '', expanded: b ? b.getAttribute('aria-expanded') : null,
      loginVisible: loginVisible(), hasInput: !!document.querySelector('[contenteditable="true"],textarea,[role="textbox"]'), readyState: document.readyState, visibility: document.visibilityState, innerWidth, innerHeight }; },
    expandSidebar() { const m = allButtons().filter(el => { const r = el.getBoundingClientRect();
        return r.width >= 15 && r.width <= 60 && r.height >= 15 && r.height <= 60 && r.left >= 0 && r.left < 50 && r.top >= 0 && r.top < 80 && !!el.querySelector('svg'); });
      if (m.length !== 1) return { action: 'none', count: m.length }; m[0].click(); return { action: 'temporary_expand' }; },
    openMenu() { const b = findAccountButton(); if (!b) return { action: 'button_not_found' };
      if (b.getAttribute('aria-expanded') === 'true') return { action: 'already_open' };
      pointer(b, 'pointerdown', 1); pointer(b, 'pointerup', 0); return { action: 'pointer_events', expanded: b.getAttribute('aria-expanded') }; },
    hoverMore() { const more = menuItems().filter(el => text(el) === '更多账号'); if (more.length !== 1) return { action: 'none', count: more.length };
      more[0].dispatchEvent(new PointerEvent('pointermove', { bubbles: true, pointerType: 'mouse', pointerId: 1 }));
      more[0].dispatchEvent(new MouseEvent('mouseover', { bubbles: true })); return { action: 'hover', expanded: more[0].getAttribute('aria-expanded') }; },
    readEntries() { const items = menuItems(); return { menuItemCount: items.length, logoutVisible: items.some(el => text(el) === '退出登录'),
      entries: items.slice(0, 100).map(el => ({ label: text(el).slice(0, 300), lines: lines(text(el)).slice(0, 6), hasAvatar: !!el.querySelector('img'), hasPopup: !!el.getAttribute('aria-haspopup'), disabled: el.getAttribute('aria-disabled') === 'true' })) }; },
    current() { const b = findAccountButton(); const login = loginVisible();
      if (!b) return { status: login ? 'not_logged_in' : 'unknown', name: '', subscription: '', raw: '' };
      const p = parseLabel(text(b)); return { status: p.name ? 'logged_in' : 'unknown', ...p, raw: text(b) }; },
    closeMenu() { if (!document.querySelector('[role="menu"]')) return { action: 'none', stillOpen: false };
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', keyCode: 27, bubbles: true }));
      const b = findAccountButton(); if (b && b.getAttribute('aria-expanded') === 'true') { pointer(b, 'pointerdown', 1); pointer(b, 'pointerup', 0); }
      return { action: 'escape', stillOpen: !!document.querySelector('[role="menu"]') }; },
    switchTo(name) { const items = menuItems().filter(el => !el.getAttribute('aria-haspopup') && el.querySelector('img') && parseLabel(text(el)).name === name);
      if (items.length !== 1) return { action: 'ambiguous', count: items.length };
      if (items[0].getAttribute('aria-disabled') === 'true') return { action: 'disabled' };
      items[0].click(); return { action: 'clicked', name }; },
  };
})();
```

- [ ] **Step 3: 真豆包逐方法验证(豆包当前已带 CDP 运行)**

依次执行并把每条输出记入 `docs/analysis/10-live-probe-macos.md`(豆包 2.32.5 macOS):
```bash
F=src-tauri/src/doubao/scripts/accounts.js
node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.probe()'
node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.current()'
node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.openMenu()'
sleep 0.7; node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.hoverMore()'
sleep 0.7; node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.readEntries()'
node scripts/cdp-probe.mjs --file $F '__doubaoRelayAccounts.closeMenu()'
```
预期:`probe.buttonFound=true`;`current.status='logged_in'` 且 name 非空;`readEntries.entries` 中至少一条 `hasAvatar=true && !hasPopup` 且 `lines[0]` 等于当前账号名;`closeMenu.stillOpen=false`。若 `stillOpen=true`,追加用 `Input.dispatchKeyEvent` Escape(在 probe 脚本加 `--key Escape` 选项)验证能关;把能关的方法写进文档,Task 1.7 的 Rust 按此实现。若 selector 不命中,改 `accounts.js` 直到命中,并记录差异。
另外记录:`Page.setWebLifecycleState`/`Emulation.setFocusEmulationEnabled` 在 macOS 是否必要(`document.visibilityState` 最小化窗口后是否仍 `visible`)。

- [ ] **Step 4: Commit** `feat(doubao): account injection script verified on macOS 2.32.5`

### Task 1.3: 错误类型 + 配置 + 日志

**Files:**
- Create: `src-tauri/src/error.rs`, `src-tauri/src/config.rs`, `src-tauri/src/logging.rs`
- Modify: `src-tauri/src/lib.rs`(声明 mod)

**Interfaces(Produces):**
```rust
// error.rs
#[derive(Debug, Clone, thiserror::Error)]
#[error("{message}")]
pub struct RelayError { pub code: &'static str, pub message: String, pub details: Option<serde_json::Value> }
impl RelayError {
    pub fn new(code: &'static str, message: impl Into<String>) -> Self;
    pub fn with_details(self, details: serde_json::Value) -> Self;
    pub fn http_status(&self) -> u16;   // 按 11-error-codes "HTTP 状态码对照" 前缀表
    // 常用构造器(每个对应一个错误码,message 固定中文):
    pub fn doubao_not_running() -> Self;  pub fn doubao_not_ready() -> Self;  pub fn doubao_not_logged_in() -> Self;
    pub fn doubao_ui_changed(failed: Vec<&str>) -> Self;  pub fn doubao_launch_timeout() -> Self;
    pub fn account_not_found(requested: &str, available: Vec<String>) -> Self;
    pub fn account_ambiguous(requested: &str, matches: Vec<String>) -> Self;
    pub fn account_switch_failed(reason: &str) -> Self;  pub fn account_switch_timeout(elapsed_ms: u64) -> Self;
    pub fn internal(msg: impl Into<String>) -> Self;     // system_internal_error
}
impl Serialize for RelayError  // -> {"code","message","details"}
pub type RelayResult<T> = Result<T, RelayError>;
pub type IpcError = RelayError;   // 同结构;Tauri 命令直接返回 Result<T, RelayError>
impl From<crate::cdp::CdpError> for RelayError;   // Closed/Connect→doubao_cdp_disconnected, Timeout→doubao_not_ready, JsException→doubao_ui_changed
impl From<crate::platform::PlatformError> for RelayError;
```
```rust
// config.rs  — 结构与 docs/spec/07-data-model.md "config.toml" 一一对应,全部字段 #[serde(default)]
pub struct Config { pub server: ServerConfig, pub concurrency: ConcurrencyConfig, pub storage: StorageConfig, pub logging: LoggingConfig, pub doubao: DoubaoConfig, pub ui: UiConfig }
impl Config {
    pub fn load_or_create(path: &Path) -> anyhow::Result<Config>;  // 不存在→默认+生成 token+写回;存在→读+补默认
    pub fn save(&self, path: &Path) -> anyhow::Result<()>;
    pub fn validate(&self) -> Result<(), String>;   // host ∈ {127.0.0.1, ::1, localhost}
    pub fn generate_token() -> String;              // 32 字节随机 hex(64 字符)
}
```
```rust
// logging.rs
pub struct LogEntry { pub ts: i64, pub level: String, pub target: String, pub message: String, pub fields: serde_json::Value }
pub struct LogBuffer(/* Arc<Mutex<VecDeque<LogEntry>>>, cap 2000 */);
impl LogBuffer { pub fn recent(&self, limit: usize, min_level: Option<&str>) -> Vec<LogEntry>; pub fn subscribe(&self) -> broadcast::Receiver<LogEntry>; }
pub fn init(level: &str, log_dir: &Path) -> anyhow::Result<LogBuffer>;  // stdout + tracing_appender 每日滚动 + 自定义 Layer 写入 LogBuffer;WorkerGuard 存 static OnceLock
```

- [ ] **Step 1: 测试(先写)** — `error.rs`:`http_status` 对 `request_x`→400, `auth_x`→401, `account_not_found`→404, `account_busy`→409, `queue_full`→429, `doubao_x`→503, `system_x`→500, 序列化含三个 key。`config.rs`:临时目录 `load_or_create` 生成 64 字符 token 且文件存在;再次 load 得同 token;手写只含 `[server]\nport=9000` 的文件 load 后其它字段为默认;`host="0.0.0.0"` validate Err。`logging.rs`:`init` 后 `tracing::info!(target: "t", "x")`,`recent(10, None)` 含 message `x`。
- [ ] **Step 2: 实现** 使 测试通过。
- [ ] **Step 3: 验证命令 + Commit** `feat(core): error codes, config, logging`

### Task 1.4: CDP 客户端(session + discovery)

**Files:**
- Create: `src-tauri/src/cdp/mod.rs`, `src-tauri/src/cdp/session.rs`, `src-tauri/src/cdp/discovery.rs`

**Interfaces(Produces):**
```rust
#[derive(Debug, thiserror::Error)]
pub enum CdpError {
    #[error("连接失败:{0}")] Connect(String),
    #[error("连接已关闭")] Closed,
    #[error("命令超时:{0}")] Timeout(String),
    #[error("CDP 错误 {code}:{message}")] Protocol { code: i64, message: String },
    #[error("页面脚本异常:{0}")] JsException(String),
    #[error("返回值解析失败:{0}")] Deserialize(String),
    #[error("HTTP 请求失败:{0}")] Http(String),
    #[error("目标不合法:{0}")] InvalidTarget(String),
}
#[derive(Clone, Debug)] pub struct CdpEvent { pub method: String, pub params: serde_json::Value }
pub struct CdpSession { /* tx: mpsc::Sender<String>, pending: Mutex<HashMap<u64, oneshot::Sender<Value>>>, events: broadcast::Sender<CdpEvent>(cap 4096), next_id: AtomicU64, closed: AtomicBool, pub target_id: String, debug_log: Option<broadcast::Sender<CdpLogEntry>> */ }
impl CdpSession {
    pub async fn connect(ws_url: &str, target_id: &str) -> Result<Arc<Self>, CdpError>;
    pub async fn send(&self, method: &str, params: Value) -> Result<Value, CdpError>;             // 15s
    pub async fn send_timeout(&self, method: &str, params: Value, timeout: Duration) -> Result<Value, CdpError>;
    pub async fn evaluate(&self, expression: &str) -> Result<Value, CdpError>;                     // returnByValue+awaitPromise, 30s;exceptionDetails→JsException(从 description 提 "Error: (.+)" 后的中文)
    pub async fn evaluate_as<T: DeserializeOwned>(&self, expression: &str) -> Result<T, CdpError>;
    pub fn subscribe(&self) -> broadcast::Receiver<CdpEvent>;
    pub fn is_closed(&self) -> bool;
    pub async fn close(&self);
}
// discovery.rs
#[derive(Clone, Debug, Deserialize)] pub struct CdpTarget { pub id: String, #[serde(rename="type")] pub kind: String, pub url: String, #[serde(default)] pub title: String, #[serde(rename="webSocketDebuggerUrl", default)] pub ws_url: String }
#[derive(Clone, Copy, Debug, PartialEq, Eq)] pub enum DoubaoPage { Chat, Background }
pub async fn check_version(host: &str, port: u16) -> Result<Value, CdpError>;          // GET /json/version, 2s 超时
pub async fn list_targets(host: &str, port: u16) -> Result<Vec<CdpTarget>, CdpError>;  // GET /json/list, 3s 超时
pub fn validate_target(t: &CdpTarget, port: u16) -> Result<DoubaoPage, CdpError>;      // 白名单规则见 Global Constraints
pub fn select_pages(targets: &[CdpTarget], port: u16) -> (Vec<CdpTarget> /*chat*/, Vec<CdpTarget> /*background*/);
```

- [ ] **Step 1: 测试(先写)** — `session.rs` 内 mock server(`TcpListener` + `accept_async`,按 method 分支:`Echo.ping`→回 `{id,result:{pong:params}}`;`Echo.event`→先推 `{method:"Test.fired",params:{n:1}}` 再回 result;`Echo.error`→`{id,error:{code:-32601,message:"no such method"}}`;`Echo.hang`→不回;`Echo.close`→服务端关闭连接;`Runtime.evaluate` 若 expression 含 `throw`→回 `{id,result:{exceptionDetails:{exception:{description:"Error: 中文错误\n at ..."}}}}` 否则回 `{id,result:{result:{type:"object",value:{ok:true}}}}`)。用例:roundtrip、error 映射 `Protocol{-32601}`、事件广播、`send_timeout(…,100ms)` 得 `Timeout`、close 后 pending 全部得 `Closed` 且 `is_closed()`、`evaluate` 异常得 `JsException("中文错误")`、`evaluate_as::<Probe>` 反序列化。`discovery.rs`:`validate_target` 对 `doubao://doubao-chat/chat` + `ws://127.0.0.1:9707/devtools/page/X` → `Chat`;`type:"iframe"`→Err;host `www.doubao.com`→Err;ws 端口 9708→Err;ws host `0.0.0.0`→Err;`chrome://doubao-background/`→`Background`。
- [ ] **Step 2: 实现**(写循环/读循环各一个 `tokio::spawn`;读循环结束时置 closed、清空 pending 发 `Closed`;`debug_log` 为 Some 时每条收发都 send `CdpLogEntry{ts,direction,method,params/result/event}`)。
- [ ] **Step 3: 验证 + Commit** `feat(cdp): session and target discovery`

### Task 1.5: 平台适配层(macOS 实装 + Windows 实装/仅 check)

**Files:**
- Create: `src-tauri/src/platform/mod.rs`, `platform/macos.rs`, `platform/windows.rs`

**Interfaces(Produces):** trait 与 `PlatformError` 照抄 `docs/spec/10-platform-adapter.md`,额外增加:
```rust
/// 返回监听 port 的进程可执行路径(无则 None)。macOS: `lsof -nP -iTCP:<port> -sTCP:LISTEN -Fp` 取 pid → sysinfo exe();Windows: `netstat -ano` 解析 → sysinfo。
fn listening_process_exe(&self, port: u16) -> Option<PathBuf>;
/// 该路径是否属于豆包安装(macOS: 包含 "/Doubao.app/";Windows: 文件名 Doubao.exe 或路径含 "\\Doubao\\")
fn is_doubao_exe_path(&self, p: &Path) -> bool;
fn doubao_version(&self) -> Option<String>;   // macOS 读 Info.plist CFBundleShortVersionString(`defaults read`);Windows None
pub fn create_platform() -> Arc<dyn Platform>;
```
macOS `find_doubao_exe` 顺序、`launch_doubao_with_cdp`(`open -a Doubao --args --remote-debugging-port=<p> --remote-debugging-address=127.0.0.1`,失败 fallback 直接执行二进制)、`kill_doubao`(`osascript quit` → 等 5s → sysinfo `kill_with(Term)` → 1s → `kill()`)、目录(`~/Library/Application Support/doubao-relay`、`~/Library/Logs/doubao-relay`)、`reveal_in_file_manager`(`open -R`)照 spec。Windows 照 spec(winreg App Paths/Uninstall、`creation_flags(DETACHED_PROCESS|CREATE_NEW_PROCESS_GROUP)`、`taskkill /IM Doubao.exe` → `/F`、`%APPDATA%\doubao-relay`、`explorer /select,`)。

- [ ] **Step 1: 测试** — `is_doubao_exe_path` 正反例(macOS:`/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser` → true,`/Applications/Google Chrome.app/...` → false);`app_data_dir()` 以 `doubao-relay` 结尾;`#[ignore] live_find_doubao_exe` 在本机返回 `/Applications/Doubao.app/Contents/MacOS/Doubao`;`#[ignore] live_listening_process_exe(9707)` 返回含 `Doubao.app` 的路径。
- [ ] **Step 2: 实现**;`rustup target add x86_64-pc-windows-msvc` 后 `cargo check -p doubao-relay --target x86_64-pc-windows-msvc`(若 tauri 依赖在交叉 check 下失败,记录原因到计划末尾"已知问题",不阻塞)。
- [ ] **Step 3: 验证 + Commit** `feat(platform): macOS and Windows adapters`

### Task 1.6: 豆包启动(bootstrap)+ 兼容性自检(compat)

**Files:**
- Create: `src-tauri/src/doubao/mod.rs`, `doubao/bootstrap.rs`, `doubao/compat.rs`

**Interfaces(Produces):**
```rust
#[derive(Clone, Copy, Debug, Serialize, PartialEq, Eq)] #[serde(rename_all="snake_case")]
pub enum StartupAction { AlreadyConnected, Reconnected, Restarted, Started }
pub async fn ensure_doubao_started(platform: &dyn Platform, port: u16) -> RelayResult<StartupAction>;
// 1) check_version ok 且 listening_process_exe 是豆包 → AlreadyConnected
// 2) check_version ok 但监听者不是豆包 → Err(doubao_cdp_port_occupied, details.owning_process)
// 3) is_doubao_running 且端口不通 → kill_doubao → launch → wait_for_cdp(20s) → Restarted
// 4) 未运行 → find_doubao_exe → launch → wait_for_cdp(20s) → Started;超时 → Err(doubao_launch_timeout)
pub async fn wait_for_cdp(host: &str, port: u16, timeout: Duration) -> RelayResult<()>;  // 每 250ms /json/list 直到出现 doubao-chat page
// compat.rs
#[derive(Clone, Debug, Serialize, Default)] #[serde(rename_all="snake_case")]
pub struct UiCompat { pub status: String /*ok|warn|error|unknown*/, pub account_button_found: bool, pub input_found: bool, pub send_button_found: bool, pub checked_at: i64 }
pub async fn ui_compat_check(session: &CdpSession) -> UiCompat;  // 注入 accounts.js 后 probe();send_button: `button[aria-label*="发送"],button[type="submit"]` 任一;status: 三者全 true→ok,input+account 有→warn,否则 error
```

- [ ] **Step 1: 测试** — `wait_for_cdp` 对未监听端口 300ms 超时返回 `doubao_launch_timeout`;用 axum/tokio 起一个本地 HTTP mock `/json/list` 返回含 chat page 的 JSON 时 `wait_for_cdp` Ok;`#[ignore] live_ensure_started` 在本机返回 `AlreadyConnected`;`#[ignore] live_compat` 返回 `status=="ok"`。
- [ ] **Step 2: 实现 + 验证 + Commit** `feat(doubao): bootstrap and ui compat check`

### Task 1.7: 账号读取(纯解析 + CDP 流程)

**Files:**
- Create: `src-tauri/src/doubao/accounts.rs`

**Interfaces(Produces):**
```rust
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)] #[serde(rename_all="snake_case")]
pub struct Account { pub name: String, pub subscription: Option<String>, pub label: String, pub is_current: bool, pub is_ambiguous: bool }
#[derive(Clone, Debug, Serialize, Deserialize)] #[serde(rename_all="snake_case")]
pub struct AccountList { pub current: Option<String>, pub accounts: Vec<Account>, pub read_at: i64 }
#[derive(Debug, Deserialize)] #[serde(rename_all="camelCase")]
pub struct RawEntry { pub label: String, pub lines: Vec<String>, pub has_avatar: bool, pub has_popup: bool, #[serde(default)] pub disabled: bool }
#[derive(Debug, Deserialize)] pub struct CurrentProbe { pub status: String, pub name: String, pub subscription: String }
pub const EXCLUDED_LABELS: &[&str] = &["设置","外观设置","收藏夹","豆包官网","检查更新","帮助与反馈","更多账号","添加账号","退出登录","豆包订阅","豆包订阅升级","免费领额度","免费领取额度"];
pub fn parse_label(raw: &str) -> (String /*name*/, Option<String> /*subscription*/);   // 与 accounts.js parseLabel 等价
pub fn parse_entries(entries: &[RawEntry], current: Option<&CurrentProbe>) -> Vec<Account>;  // 过滤 !has_avatar/has_popup/排除词/"当前时段额度用量"前缀/name 长度 1..=120;同 label 去重;同 name 不同 label → 全部 is_ambiguous;is_current = name==current.name;若结果为空且 current.status=="logged_in" → 返回仅含 current 的单元素
pub async fn ensure_injected(session: &CdpSession) -> RelayResult<()>;  // `if(!window.__doubaoRelayAccounts){<accounts.js>}`
pub async fn read_accounts(session: &CdpSession) -> RelayResult<AccountList>;
// 流程:ensure_injected → probe(未找到按钮→expandSidebar→400ms→probe;loginVisible→doubao_not_logged_in;仍无→doubao_ui_changed) → current() → openMenu → 650ms → hoverMore → 650ms → 循环≤8 次 {hoverMore; 400ms; readEntries; 计数稳定(attempt>=3 && stable>=2)则 break} → closeMenu(按 Task 1.2 文档的有效方式,含 Input.dispatchKeyEvent Escape 兜底) → parse_entries
```

- [x] **Step 1: 测试** — `parse_label`:`"张三 体验套餐"`→(张三,Some(体验套餐));`"张三\n高级套餐"`→同;`"李四 字节科技"`→(李四 字节科技,None);`""`→("",None)。`parse_entries`:排除词被过滤;`has_popup` 被过滤;同名不同 label 两条都 `is_ambiguous`;current 匹配 `is_current`;空列表 + logged_in → 单元素。`#[ignore] live_read_accounts`:连本机 chat target,断言 `accounts.len()>=1` 且 `current.is_some()`。
- [x] **Step 2: 实现 + 验证 + Commit** `feat(doubao): account list reading`

### Task 1.8: DoubaoConnection(会话管理 / 心跳 / 重连)+ 账号切换

**Files:**
- Create: `src-tauri/src/doubao/connection.rs`
- Modify: `src-tauri/src/doubao/accounts.rs`(新增 `switch_account`)

**Interfaces(Produces):**
```rust
#[derive(Clone, Debug, Serialize)] #[serde(rename_all="snake_case")]
pub struct DoubaoStatus { pub running: bool, pub cdp_connected: bool, pub cdp_port: u16, pub doubao_version_detected: Option<String>, pub ui_compat_check: String, pub ui_compat_details: UiCompat, pub chat_targets: usize, pub reconnecting: bool, pub last_check_at: i64 }
#[derive(Clone, Debug)] pub enum ConnEvent { StatusChanged(DoubaoStatus), AccountListChanged(AccountList), AccountSwitched { previous: Option<String>, current: String, switched_at: i64 }, Crashed }
pub struct DoubaoConnection { /* platform, host, port, sessions: RwLock<Sessions{chat:Option<Arc<CdpSession>>, background:Option<Arc<CdpSession>>}>, status: RwLock<DoubaoStatus>, account_cache: RwLock<Option<(AccountList, Instant)>>, switch_lock: Mutex<()>, events: broadcast::Sender<ConnEvent>, cdp_debug: Option<broadcast::Sender<CdpLogEntry>> */ }
impl DoubaoConnection {
    pub fn new(platform: Arc<dyn Platform>, host: String, port: u16) -> Arc<Self>;
    pub fn subscribe(&self) -> broadcast::Receiver<ConnEvent>;
    pub async fn status(&self) -> DoubaoStatus;
    pub async fn connect(&self, force_restart: bool) -> RelayResult<StartupAction>;   // force→kill+launch;否则 ensure_doubao_started → connect_sessions → compat check → 发 StatusChanged
    pub async fn connect_sessions(&self) -> RelayResult<()>;   // list_targets→select_pages;chat 必须恰好 1 个(0→doubao_not_ready,>1→doubao_not_ready details.chat_targets);连接 chat(+background 可选);每个 session `Network.enable`;发 Page.setWebLifecycleState active + Emulation.setFocusEmulationEnabled true(按 Task 1.2 结论决定是否在 mac 执行)
    pub async fn chat_session(&self) -> RelayResult<Arc<CdpSession>>;   // None 或 is_closed → 尝试一次 connect_sessions,仍失败→doubao_cdp_disconnected
    pub async fn heartbeat(&self) -> bool;   // Target.getTargets 5s 超时;失败→标记断开→发 StatusChanged;连续失败→尝试 connect_sessions;20s 仍失败且 !is_doubao_running→发 Crashed
    pub fn spawn_heartbeat(self: &Arc<Self>, interval: Duration) -> JoinHandle<()>;   // 10s
    pub async fn accounts(&self, force: bool) -> RelayResult<AccountList>;  // 2s 缓存;读取后发 AccountListChanged
    pub async fn switch_account(&self, target: &str) -> RelayResult<SwitchResult>;
}
#[derive(Clone, Debug, Serialize)] #[serde(rename_all="snake_case")] pub struct SwitchResult { pub ok: bool, pub previous: Option<String>, pub current: String, pub switched_at: i64 }
// switch_account:switch_lock → accounts(true) → 目标不存在→account_not_found(available);is_ambiguous→account_ambiguous;is_current→Ok 直接返回
//   → openMenu/hoverMore 快节奏(100ms × ≤40)直到 switchTo(name).action=="clicked",否则 account_switch_failed
//   → 验证循环 20s:每 300ms evaluate current();CdpError::Closed/Timeout → connect_sessions() 后继续;status==logged_in && name==target → 成功;超时 account_switch_timeout(elapsed)
//   → 清缓存,发 AccountSwitched + AccountListChanged
```

- [x] **Step 1: 测试** — 用 Task 1.4 的 mock WS server + 一个 mock HTTP `/json/list`(可通过 `DoubaoConnection::new_with_host(host, port)` 指向 mock 端口):`chat_session()` 在 session 被 mock 关闭后重新连接并返回新 `target_id`;`heartbeat()` 在 mock 不响应 `Target.getTargets` 时返回 false 且 `status().cdp_connected==false`。`#[ignore] live_switch_roundtrip`:若本机账号 ≥2,切到非当前账号再切回,两次都 Ok(用 `--nocapture` 打印耗时)。
- [x] **Step 2: 实现 + 验证 + Commit** `feat(doubao): connection manager, heartbeat, account switch`

### Task 1.9: AppState + Tauri 命令 + 事件转发

**Files:**
- Create: `src-tauri/src/state.rs`, `src-tauri/src/commands/mod.rs`, `commands/app.rs`, `commands/account.rs`, `commands/doubao.rs`, `commands/log.rs`
- Modify: `src-tauri/src/lib.rs`

**Interfaces(Produces):** Tauri 命令(全部 `async`, `State<'_, Arc<AppState>>`, 返回 `Result<Value, RelayError>`,Value 经 `to_camel`):
`app_get_version` → `{version, buildTarget, tauriVersion, rustcVersion}`;`app_get_config` / `app_save_config(patch: Value)` → `{ok, needsRestart}`;`account_list`;`account_refresh`;`account_switch(target: String)`;`doubao_status`;`doubao_reconnect(force: bool)` → `{action, cdpReadyAt}`;`log_recent(limit: Option<usize>, level: Option<String>)`;`debug_cdp_log(limit)`。
事件(Rust→WebView,`app.emit`):`doubao_status_changed`、`account_list_changed`、`account_switched`、`log_entry`、`cdp_log_entry`(payload 均 camelCase)。
`lib.rs::run()`:platform → config → logging → `DoubaoConnection` → `AppState` → `tauri::Builder.manage(state).setup(|app| { spawn 事件转发任务(ConnEvent/LogEntry→emit); spawn_heartbeat; if config.doubao.auto_connect_on_startup { spawn connect(false) } })`。

- [ ] **Step 1: 测试** — `to_camel` 单测(嵌套对象/数组/已是 camel 的 key 不变)。
- [ ] **Step 2: 实现 + `cargo tauri dev` 冒烟(窗口出现无 panic)+ Commit** `feat(ipc): tauri commands and events`

### Task 1.10: 前端账号页 + 状态栏 + Toast + 联调

**Files:**
- Create: `frontend/src/api.ts`, `state.ts`, `components/sidebar.tsx`, `components/status-bar.tsx`, `components/toast.tsx`, `components/confirm-dialog.tsx`, `pages/accounts.tsx`
- Modify: `app.tsx`, `view-model.ts`(+测试)

**Interfaces:**
- `api.ts`:`invoke` 封装 `accountList/accountRefresh/accountSwitch/doubaoStatus/doubaoReconnect/appGetVersion/logRecent`;`onEvent<T>(name, cb)`;`asRelayError(e): {code,message,details?}`(非结构化错误转 `system_internal_error`)。
- `state.ts`:signals `doubaoStatus`, `accountList`, `accountsLoading`, `switching: string|null`, `toasts`, `route`。
- `view-model.ts`:新增 `accountRows(list): Array<{name, subscription, isCurrent, isAmbiguous, canSwitch}>`(ambiguous 不可切)+ `emptyStateFor(status, list): 'not_connected'|'not_logged_in'|'empty'|null`,各带测试。
- 账号页按 `docs/spec/06-webview-ui.md` "页面 1":标题行 [刷新][重连豆包];"当前账号: X (套餐)";表格 名字/套餐/状态;★ 当前;⚠ 同名提示;切换需 `confirm-dialog`;切换中行内 loading;空状态三种。订阅 `account_list_changed`/`doubao_status_changed`/`account_switched` 事件刷新。

- [x] **Step 1: 测试** view-model 两个函数。
- [x] **Step 2: 实现**。
- [x] **Step 3: 真机验收(UC-01~06)** — `cargo tauri dev`,`screencapture -x` 逐项核对:账号列表正确、当前账号 ★、点"切换"弹确认 → 20s 内 UI 显示新当前账号;手动 `pkill -x Doubao` 后 15s 内底部"●豆包"变红并显示"豆包已退出,等待重启";点"重连豆包"恢复(走 Restarted,等待 ≤20s)。每项的截图路径与结果记入 `docs/analysis/10-live-probe-macos.md` 末尾"Phase 1 验收"。
- [x] **Step 4: Commit** `feat(ui): accounts page with live status`

---

# Phase 2 · 生图核心 + 持久化 + HTTP API

验收:UC-10 ~ UC-18。

### Task 2.1: 生成页真机探测(对话模式 vs 工作模式)

**Files:** Create/Modify `docs/analysis/10-live-probe-macos.md`(章节"生成页探测")

- [x] **Step 1** 用 `scripts/cdp-probe.mjs --file src-tauri/src/doubao/scripts/doubao-generation-page.js` 注入后,分别对图像请求 `{type:'image',modelName:'Seedream 5.0 Pro',prompt:'测试',ratio:'1:1',submissionMode:'conversation',parallelTracking:true,trackingKey:'shenlangDoubaoGeneration',confirmationMode:'auto',autoUseMembership:true,referenceFiles:[]}` 执行 `shenlangDoubaoGeneration.run('prepare', req)`,记录返回或异常文本;对视频 `{type:'video',modelName:'Seedance 2.0 Mini',duration:12,ratio:'16:9',...}` 同样;再对 `modelName:'Seedance 2.5',duration:10` 观察是否抛 `对话模式仅支持Mini的11～15秒任务`;若抛,把 `submissionMode:'work'` 再试,记录需要先进入哪个页面(豆包"AI 创作"/"图像生成"tab 的 URL / 点击路径)。**不要执行 `send`**(会消耗配额),`fill` 后立即用 `run('fill', {...req, prompt:''})` 清空。
- [x] **Step 2** 记录模型菜单是 DOM 还是原生(`prepare` 是否抛 `未识别到所选…模型菜单项`),文件上传入口是否触发 `Page.fileChooserOpened`(开 `Page.setInterceptFileChooserDialog` 后 `run('fileInput', req)`)。
- [x] **Step 3** 结论表:每种 (type, model) 组合 → 可用的 submissionMode、prompt 前缀规则、prepare 需要的前置导航。Task 2.3 据此实现 `SubmissionPlan::for_request()`。Commit `docs: live probe of generation page`。

### Task 2.2: 大整数修复 + 任务模型 + 请求校验

**Files:** Create `src-tauri/src/cdp/json_fix.rs`, `src-tauri/src/tasks/mod.rs`, `tasks/model.rs`, `tasks/validate.rs`;Cargo 增 `ulid = "1"`, `sha2 = "0.10"`, `hex = "0.4"`

**Interfaces(Produces):**
```rust
pub fn fix_big_integers(json_text: &str) -> String;   // 正则见 09-cdp-contract "大整数 JSON 处理",字段 conversation_id|local_message_id|message_id|bot_reply_message_id
// tasks/model.rs(serde snake_case)
pub enum TaskType { Image, Video }   pub enum ConfirmationMode { Auto, Manual }
pub enum TaskStatus { Pending, Preparing, Generating, WaitingUserConfirm, Completed, Failed, Cancelled }   impl TaskStatus { pub fn is_terminal(&self)->bool; pub fn as_str(&self)->&'static str; pub fn parse(&str)->Option<Self>; }
pub enum TaskStage { Preparing, Sent, Checking, QuotaChecking, Rendering, RenderingUnverified, Downloading, Result, Refused }   (as_str/parse)
pub enum WatermarkSource { ConversationOriginal, LrUnwatermarked, FallbackDecrypted }
pub struct GenerateRequest { pub account: String, #[serde(rename="type")] pub task_type: TaskType, pub model_name: String, pub prompt: String, pub ratio: String, pub duration: Option<u32>, #[serde(default)] pub reference_files: Vec<String>, #[serde(default)] pub confirmation_mode: ConfirmationMode, #[serde(default="default_true")] pub auto_use_membership: bool, pub callback_url: Option<String> }
pub struct ResultFile { pub index: u32, pub filename: String, pub relative_path: String, pub mime: String, pub bytes: u64, pub width: Option<u32>, pub height: Option<u32>, pub duration_ms: Option<u64> }
pub struct TaskError { pub code: String, pub message: String, pub details: Option<Value> }
pub struct StageEntry { pub stage: String, pub message: Option<String>, pub ts: i64 }
pub struct Task { pub id: String, pub request: GenerateRequest, pub status: TaskStatus, pub stage: Option<TaskStage>, pub stage_message: Option<String>, pub progress: Option<f32>, pub error: Option<TaskError>, pub conversation_id: Option<String>, pub submission_id: Option<String>, pub local_message_id: Option<String>, pub watermark_source: Option<WatermarkSource>, pub result_files: Option<Vec<ResultFile>>, pub created_at: i64, pub started_at: Option<i64>, pub completed_at: Option<i64>, pub expires_at: i64, pub stage_history: Vec<StageEntry>, pub queue_position: Option<u32> }
pub struct TaskSummary { id, account, task_type, status, stage, created_at, completed_at }   impl From<&Task>
pub fn new_task_id() -> String;  // Ulid
pub fn prompt_hash(prompt: &str) -> String;  // sha256 hex 前 16 位(日志用)
// tasks/validate.rs
pub const IMAGE_MODELS: &[&str]; pub const VIDEO_MODELS: &[&str]; pub const RATIOS: &[&str];
pub fn validate_request(r: &GenerateRequest) -> RelayResult<()>;  // 按 F3.1/F3.2/F3.5 + 11-error-codes request_*;callback_url 必须 http(s) 且非 loopback(request_invalid_callback_url)
```
- [ ] **Step 1: 测试** — json_fix(见 Review Focus 4);validate:缺 prompt→request_missing_field{field:"prompt"};model 不在白名单→request_invalid_enum{field,value,allowed};video 无 duration→request_missing_field;Seedance 2.5 duration 30 Ok、31 Err request_invalid_range;Seedance 2.0 duration 16 Err;refs 11 张图像→request_too_many_references;Fast 视频 11 Ok;prompt 10001 字→request_prompt_too_long;callback `http://127.0.0.1/x`→request_invalid_callback_url;`ftp://`→同。
- [ ] **Step 2: 实现 + Commit** `feat(tasks): task model and request validation`

### Task 2.3: 生成管线(prepare/fill/checkSend/send + submissionIdentity)

**Files:** Create `src-tauri/src/doubao/generation.rs`

**Interfaces(Produces):**
```rust
pub struct SubmissionPlan { pub submission_mode: &'static str /*conversation|work*/, pub prompt: String /*含前缀*/, pub tracking_key: String }
impl SubmissionPlan { pub fn for_request(r: &GenerateRequest, task_id: &str) -> SubmissionPlan; }  // 规则来自 Task 2.1 结论;tracking_key = format!("shenlangDoubaoTask_{}", task_id 小写)
pub struct SubmissionIdentity { pub local_message_id: String, pub message_id: Option<String>, pub conversation_id: Option<String> }
pub struct SubmissionReport { pub identity: Option<SubmissionIdentity>, pub send_started_at: i64 }
pub fn generation_bootstrap(tracking_key: &str) -> String;   // `if (typeof <key> === 'undefined') { <generation-page.js 内 shenlangDoubaoGeneration 替换为 key> }`
pub fn step_expression(tracking_key: &str, step: &str, request_json: &Value) -> String;  // `Promise.resolve(<key>.run("<step>", <json>))`
pub fn build_js_request(r: &GenerateRequest, plan: &SubmissionPlan, reference_paths: &[PathBuf], uploaded_count: usize, user_confirmed: bool) -> Value;  // camelCase 字段:type/modelName/prompt/ratio/duration/referenceFiles/submissionMode/parallelTracking=true/trackingKey/confirmationMode/autoUseMembership/uploadedCount/userConfirmed
pub async fn submit(session: &CdpSession, background: Option<&CdpSession>, r: &GenerateRequest, plan: &SubmissionPlan, reference_paths: &[PathBuf], cancel: &CancellationToken, on_stage: &(dyn Fn(TaskStage, Option<String>) + Send + Sync)) -> RelayResult<SubmissionReport>;
// 流程(每步前检查 cancel):focus 伪装 → bootstrap → prepare(60s)(失败含"未识别到所选…模型菜单项"→generation_preparing_failed)→ 若 needsDurationInput → duration::set_duration(Phase 3,此处返回 Err 若视频)→ 若 refs 非空 → upload::upload_references(Phase 3)→ fill → checkSend → 订阅 Network.requestWillBeSent(chat+background,`Network.enable {maxPostDataSize:16777216}`)→ send → 2s 内对每个 POST(host doubao.com|snssdk.com,path 含 completion|generate|/im/message/send,不含 history|list|recommend|config)evaluate `shenlangDoubaoResultReader.submissionIdentity(req, prompt, {verifiedSend:true})`(reader 需先 bootstrap)→ 得 identity 或 None(None→generation_submit_identity_missing)
```
需先 `grep` `docs/analysis/scripts/doubao-result-reader.js` 确认 `submissionIdentity` 的参数/返回 key(`localMessageId/messageId/conversationId`)并在文件头注释记录。
- [ ] **Step 1: 测试** — `SubmissionPlan::for_request` 按 Task 2.1 结论;`build_js_request` 字段名;`step_expression` 字符串;`is_generation_request(url, method)` 正反例(`https://www.doubao.com/samantha/chat/completion` POST → true;`.../history` → false;`https://api.snssdk.com/x/generate` → true;GET → false);mock WS:send 后推送一个 requestWillBeSent 事件 + `Runtime.evaluate` 回 identity → `submit` 返回 identity。
- [ ] **Step 2: 实现 + Commit** `feat(doubao): generation submit pipeline`

### Task 2.4: 捕获管线(图像)+ feedback 阶段映射

**Files:** Create `src-tauri/src/doubao/capture.rs`, `doubao/feedback.rs`

**Interfaces(Produces):**
```rust
pub struct CaptureOptions { pub task_type: TaskType, pub prompt: String, pub identity: SubmissionIdentity, pub send_since: i64, pub confirmation_mode: ConfirmationMode, pub auto_use_membership: bool, pub tracking_key: String, pub deadline: Duration }
pub enum CaptureEvent { Stage(TaskStage, Option<String>), Bound { conversation_id: String, submission_id: String }, Feedback(FeedbackState), NeedsUserConfirm, Refused { kind: String, message: String }, Originals(Vec<String>) /*图像 URL*/, VideoCandidates(Vec<String>), Fallbacks(Vec<FallbackVideo>) }
pub struct FeedbackState { pub stage: String, pub message: Option<String>, pub sent: bool }
pub struct FallbackVideo { pub vid: String, pub api: String }
pub struct CaptureOutcome { pub originals: Vec<String>, pub watermark_source: WatermarkSource, pub conversation_id: Option<String>, pub submission_id: Option<String> }
pub async fn capture(conn: &DoubaoConnection, opts: CaptureOptions, cancel: &CancellationToken, confirm_rx: &mut watch::Receiver<bool>, events: mpsc::Sender<CaptureEvent>) -> RelayResult<CaptureOutcome>;
// 循环(tokio::select!):Network 事件(requestWillBeSent 分类 历史请求/生成请求,过滤 initiator 含 shenlangReadBoundConversation;loadingFinished→getResponseBody;eventSourceMessageReceived 缓冲)| 每 500ms feedback 轮询(terminalFeedback / videoFeedback / videoFollowup 选择规则见 analysis/05)| 每 1s 历史重放 `shenlangDoubaoResultReader.read(historyRequest, options)`(options 字段见 analysis/05 "analyze")| 每 2s chat target 唯一性检查(≠1→generation_user_interrupted{interruption:"multiple_chat_pages"})| 每 10s 输入框脏检查(warn)| deadline → generation_task_timeout
// feedback.rs
pub fn map_feedback_stage(js_stage: &str) -> Option<TaskStage>;  // checking→Checking, quota_checking→QuotaChecking, generating→Rendering, rendering_unverified→RenderingUnverified, result→Result, refused|failed|quota_exhausted|not_executed→Refused
pub fn refusal_error(js_stage: &str, message: &str) -> TaskError;  // quota_exhausted→account_quota_exhausted;not_executed→generation_rendering_not_executed;其它→generation_doubao_refused{reason}
pub fn should_auto_confirm(stage: &TaskStage, mode: ConfirmationMode, auto_membership: bool) -> bool;
// rendering_unverified 连续 30s → not_executed;auto 模式 followup 异常 → 降级 manual 并发 NeedsUserConfirm
pub fn extract_image_originals(body: &str) -> Vec<String>;  // 递归找 image_ori_raw.url,host byteimg.com|volces.com,https,去重保序;支持 SSE "data: {...}" 行
```
- [x] **Step 1: 测试** — `extract_image_originals`(嵌套 JSON、SSE 多行、非法 host 过滤、HTML 实体 `&amp;` 解码);`map_feedback_stage`/`refusal_error`/`should_auto_confirm` 真值表;mock WS 场景:推送 requestWillBeSent(`/im/chain/single` POST)+ loadingFinished → 断言发出 `Network.getResponseBody`;`Runtime.evaluate`(read)回 `{ok:true,images:["https://x.byteimg.com/a"],feedback:{stage:"result"},binding:{conversationId:"1",submissionId:"2"}}` → `capture` 返回 originals;feedback 回 `{stage:"quota_exhausted",message:"额度不足"}` → Err account_quota_exhausted;`cancel.cancel()` → Err generation_cancelled 且 50ms 内返回。
- [x] **Step 2: 实现 + Commit** `feat(doubao): capture loop for images`

### Task 2.5: 结果下载 + 格式识别

**Files:** Create `src-tauri/src/doubao/download.rs`;Cargo 增 `image = { version="0.25", default-features=false, features=["png","jpeg","webp"] }`

**Interfaces(Produces):**
```rust
pub struct Downloaded { pub path: PathBuf, pub mime: String, pub bytes: u64, pub width: Option<u32>, pub height: Option<u32> }
pub fn detect_image_format(bytes: &[u8]) -> RelayResult<(&'static str /*mime*/, &'static str /*ext*/, (u32,u32))>;  // image::guess_format + 尺寸;非 PNG/JPEG/WebP→generation_download_failed
pub async fn download_image(client: &reqwest::Client, url: &str, dest_dir: &Path, index: u32) -> RelayResult<Downloaded>;  // 40s 超时,Referer https://www.doubao.com/,>100MB 拒绝,写 image_{index+1}.{ext}
pub async fn download_video(client: &reqwest::Client, url: &str, dest_dir: &Path) -> RelayResult<Downloaded>;  // 180s,校验 offset 4 'ftyp',≤1GB,video_1.mp4(Phase 3 调用)
pub fn new_client() -> reqwest::Client;  // 不跟随重定向(redirect::Policy::none()),UA 固定
```
- [ ] **Step 1: 测试** — detect(PNG/JPEG/WebP 最小合法字节、随机字节 Err);axum mock server 返回 PNG → `download_image` 写文件、mime、尺寸正确;返回 302 → Err;返回 101MB → Err。
- [ ] **Step 2: 实现 + Commit** `feat(doubao): result download`

### Task 2.6: 持久化(SQLite + 文件 + GC)

**Files:** Create `migrations/20261016000001_initial.sql`(照抄 07-data-model 五张表 + 索引,`PRAGMA journal_mode=WAL` 在连接时设置),`src-tauri/src/storage/mod.rs`, `storage/db.rs`, `storage/files.rs`, `storage/gc.rs`;Cargo 增 `sqlx = { version="0.8", features=["runtime-tokio","sqlite","migrate"] }`

**Interfaces(Produces):**
```rust
pub struct Storage { pool: SqlitePool, pub root: PathBuf /*app_data_dir*/ }
impl Storage {
    pub async fn open(root: &Path) -> anyhow::Result<Storage>;   // 建目录 uploads/ results/ logs/;sqlx::migrate!("../migrations");WAL
    pub async fn insert_task(&self, t: &Task) -> RelayResult<()>;
    pub async fn get_task(&self, id: &str) -> RelayResult<Option<Task>>;        // 含 stage_history
    pub async fn update_task(&self, t: &Task) -> RelayResult<()>;               // 全字段 UPDATE
    pub async fn append_stage(&self, id: &str, stage: &str, message: Option<&str>, ts: i64) -> RelayResult<()>;
    pub async fn list_tasks(&self, status: Option<&[TaskStatus]>, account: Option<&str>, limit: u32, cursor: Option<&str>) -> RelayResult<(Vec<TaskSummary>, Option<String>)>;  // 按 id DESC,cursor = 上页最后 id
    pub async fn fail_non_terminal(&self, code: &str, message: &str, now: i64) -> RelayResult<Vec<String>>;   // 启动恢复 relay_restart / 崩溃 doubao_crashed;实现返回被改任务 id(Task 4.1 审查修复,原计划 u64 行数)
    pub async fn insert_upload(&self, u: &Upload) -> RelayResult<()>;  pub async fn get_upload(&self, id: &str) -> RelayResult<Option<Upload>>;
    pub async fn insert_webhook_log(&self, l: &WebhookLog) -> RelayResult<()>;
    pub async fn get_meta(&self, key: &str) -> RelayResult<Option<String>>;  pub async fn set_meta(&self, key: &str, value: &str, now: i64) -> RelayResult<()>;
    pub fn result_dir(&self, task_id: &str) -> PathBuf;  pub fn upload_path(&self, id: &str, ext: &str) -> PathBuf;
    pub async fn gc(&self, now: i64, cfg: &StorageConfig) -> RelayResult<GcReport>;  // 四步见 07 "定时 GC"
}
pub struct Upload { pub id: String, pub mime: String, pub bytes: u64, pub path: PathBuf, pub created_at: i64, pub expires_at: i64 }
pub struct GcReport { pub result_files_removed: u64, pub tasks_removed: u64, pub uploads_removed: u64 }
```
用 `sqlx::query()` 运行时绑定,**不用** `query!` 宏(避免 DATABASE_URL)。
- [ ] **Step 1: 测试** — 临时目录 open → migrate 成功;insert/get 往返(含 request JSON、stage_history 顺序);list 分页 cursor;`fail_non_terminal` 只改非终态;gc:过期任务文件被删且 `result_files` 置 null、超 7 天任务级联删除、过期 upload 文件删。
- [ ] **Step 2: 实现 + Commit** `feat(storage): sqlite persistence and gc`

### Task 2.7: 任务调度器 + 运行器(图像路径)

**Files:** Create `src-tauri/src/tasks/scheduler.rs`, `tasks/runner.rs`;Cargo 增 `tokio-util = "0.7"`

**Interfaces(Produces):**
```rust
pub enum TaskEvent { Created(TaskSummary), Updated { task_id: String, status: TaskStatus, stage: Option<TaskStage>, message: Option<String>, progress: Option<f32>, updated_at: i64 }, Completed(Task) }
pub struct Scheduler { /* storage, doubao, config: Arc<RwLock<Config>>, pending: Mutex<VecDeque<String>>, running: Mutex<HashMap<String, TaskHandle>>, account_locks: Mutex<HashMap<String, Arc<Mutex<()>>>>, switch_lock(来自 DoubaoConnection), global: Semaphore, events: broadcast<TaskEvent>, wake: Notify */ }
pub struct TaskHandle { pub cancel: CancellationToken, pub confirm_tx: watch::Sender<bool>, pub status: Arc<RwLock<TaskStatus>> }
impl Scheduler {
    pub fn new(...) -> Arc<Self>;  pub fn subscribe(&self) -> broadcast::Receiver<TaskEvent>;
    pub async fn submit(&self, req: GenerateRequest) -> RelayResult<(String /*id*/, u32 /*queue_position*/, i64)>;  // validate → 账号存在性(doubao.accounts(false),不存在→account_not_found;ambiguous→account_ambiguous)→ uploads 存在且未过期 → 队列深度(pending+waiting ≥ max→queue_full)→ insert_task(pending)→ 发 Created → wake
    pub async fn cancel(&self, id: &str) -> RelayResult<TaskStatus /*from*/>;   // pending:直接 cancelled;preparing/generating/waiting:cancel token;终态→task_not_cancellable
    pub async fn confirm(&self, id: &str) -> RelayResult<String /*new_stage*/>;  // 非 waiting_user_confirm→task_not_waiting_confirm;confirm_tx.send(true)
    pub async fn queue_position(&self, id: &str) -> Option<u32>;
    pub fn spawn_loop(self: &Arc<Self>) -> JoinHandle<()>;   // FIFO:对 pending 头任务,若 global 有名额且账号锁空闲 → spawn runner::run;每 10s 超时扫描(pending>1h→queue_timeout…见 08 表)
    pub async fn fail_all_running(&self, code: &str, message: &str);   // doubao Crashed 时调用
}
// runner.rs
pub async fn run(sched: Arc<Scheduler>, task_id: String, handle: TaskHandle);
// preparing:set status;若 current != account → doubao.switch_account;→ generation::submit(60s 总超时→generation_preparing_timeout)→ generating(stage sent)→ capture(…)(映射 CaptureEvent→状态/阶段/DB/事件;NeedsUserConfirm→status waiting_user_confirm,confirm_rx 变 true 或 1h 超时→generation_confirm_timeout)→ downloading → download 每张(≤20)→ result_files → completed;任何 Err → failed(TaskError);cancel → cancelled;所有状态变化:DB 先写再 emit(08 "事件广播")。日志 info 字段 {task_id, from_status, to_status, stage, duration_ms, prompt_length, prompt_hash}
```
- [x] **Step 1: 测试** — 用 trait `GenerationBackend`(submit/capture/download 可 mock)注入 runner,测试:成功路径状态序列 pending→preparing→generating→completed 且 DB 一致;capture Err→failed 且 error.code 正确;pending 取消;同账号两个任务串行(第二个在第一个完成前保持 pending);不同账号并发 ≤ global_max;队列满 51 → queue_full;manual 模式 NeedsUserConfirm→waiting→confirm→继续。
- [x] **Step 2: 实现 + Commit** `feat(tasks): scheduler and runner`

### Task 2.8: HTTP API 服务器

**Files:** Create `src-tauri/src/http/mod.rs`, `http/auth.rs`, `http/error.rs`, `http/routes/{health,accounts,uploads,generate,tasks,doubao}.rs`;Cargo 增 `axum = { version="0.8", features=["multipart"] }`, `tower-http = { version="0.6", features=["fs","trace","limit"] }`, `subtle = "2"`

**Interfaces(Produces):** 路由与响应体严格按 `docs/spec/04-http-api.md`;`/api/v1/health` 免鉴权;其余 `auth_middleware`(缺 header→401 auth_missing_token;不匹配→401 auth_invalid_token;`subtle::ConstantTimeEq`);错误 → `RelayError` IntoResponse(`http_status()`)。
```rust
pub fn router(state: Arc<AppState>) -> axum::Router;
pub async fn serve(state: Arc<AppState>, host: &str, port: u16, shutdown: CancellationToken) -> anyhow::Result<()>;  // bind 失败 → Err(用户可改端口);启动成功 info 日志 "HTTP API 已监听 {host}:{port}"
```
`POST /uploads`:multipart `file`,`DefaultBodyLimit` 50MB,MIME 由内容判(image 用 Task 2.5 detect;mp4 'ftyp')→ 不支持→upload_mime_not_supported;写 `uploads/{ulid}.{ext}`;expires 1h。
`GET /tasks/{id}/download/{index}`:`tower_http::services::ServeFile`(支持 Range)+ `Content-Disposition: attachment; filename="task-{id}-{index}.{ext}"`;过期→404 result_expired;index 越界→404 result_index_out_of_range。
- [x] **Step 1: 测试(axum `oneshot` / `tower::ServiceExt`)** — health 200 无 token;`/accounts` 无 token 401;错 token 401;`POST /generate` 缺字段 400 code;`GET /tasks/x` 404 task_not_found;`POST /uploads` 超 50MB 413→转 400 upload_too_large;上传 PNG 200 带 upload_id;`GET /uploads/{id}` 200;下载端点 Range 请求 206。
- [x] **Step 2: 实现;在 `lib.rs` 中以独立 `tokio::spawn` 启动;启动时 `storage.fail_non_terminal("relay_restart", …)`;GC 每 1h。Commit** `feat(http): axum api server`

### Task 2.9: 任务相关 Tauri 命令 + 前端任务页/详情页/新建弹窗

**Files:** Create `commands/task.rs`, `commands/upload.rs`, `frontend/src/pages/tasks.tsx`, `pages/task-detail.tsx`, `components/new-task-dialog.tsx`;Modify `api.ts`, `state.ts`, `view-model.ts`(+测试), `app.tsx`

命令:`task_list(filter)`、`task_get(taskId)`、`task_submit(request)`、`task_cancel(taskId)`、`task_confirm(taskId)`、`task_open_result_folder(taskId)`、`upload_add(filePath, kind)`(用 `tauri-plugin-dialog` 选文件:Cargo 增 `tauri-plugin-dialog = "2"`,capability 加 `dialog:default`)。事件 `task_created/task_updated/task_completed`。
UI 按 06-webview-ui "页面 2/3":列表(状态图标、hover 显示 prompt 前 60 字、过滤、加载更多)、详情(时间轴 live、结果缩略图 `<img src>` 用 `convertFileSrc`,需 `asset` protocol: tauri.conf `app.security.assetProtocol.enable=true, scope=["$APPDATA/doubao-relay/results/**"]`、取消/手动确认按钮、原始请求折叠)。`view-model.ts` 增 `statusIcon(status)`、`relativeTime(ts, now)`、`timelineRows(stage_history)`,各带测试。
- [x] **Step 1: 测试(view-model)**;**Step 2: 实现**;**Step 3: 真机验收 UC-10 ~ UC-18**(用 curl 脚本 `scripts/uc-10.sh`;生一张图,确认 `download/0` 是合法 PNG;UI 任务页看到完成缩略图;截图记录到 10-live-probe 文档);**Step 4: Commit** `feat(ui): tasks pages and api smoke`

---

# Phase 3 · 生视频

验收:UC-20 ~ UC-23, UC-27。

### Task 3.1: 视频解密(Rust 实现)

**Files:** Create `src-tauri/src/doubao/video_decoder.rs`;Cargo 增 `aes = "0.8"`, `cbc = { version="0.1", features=["alloc"] }`, `base64 = "0.22"`

**Interfaces(Produces):** 完全按 `docs/analysis/06-video-decryption.md`:
```rust
pub fn base64_decode_flexible(input: &str) -> Option<Vec<u8>>;   // 三种变体
pub fn bytes_to_printable(bytes: &[u8]) -> Option<String>;
pub fn extract_key_seed(value: &Value) -> Option<String>;  pub fn extract_key_seed_from_str(s: &str) -> Option<String>;
pub fn pick_main_url_entry(response: &Value) -> Option<(String /*token*/, Value /*entry*/)>;   // 评分 H.264 + bitrate + 面积
pub fn aes_cbc_decrypt(ciphertext: &[u8], key: &[u8], iv: &[u8]) -> Option<Vec<u8>>;   // NoPadding + 手工 PKCS7 校验
pub fn decrypt_token(token: &str, key_seed: &str) -> Option<String>;   // 候选顺序 (key3,iv3),(key1,iv1),(key2,iv2),(offset 0x24 变体…),无 magic 时 (key1,iv2)
pub fn decode(response: &Value, api_url: &str) -> Result<String, &'static str>;   // 'fallback_main_url_missing' | 'fallback_url_decode_failed'
pub fn unwatermarked_url(url: &str) -> String;   // 追加/替换 query lr=unwatermarked(与 reader.js unwatermarkedURL 等价,先 grep 确认规则)
```
- [ ] **Step 1: 测试** — 用已知 key/iv 自己 AES-CBC 加密一个 URL(测试内用 `cbc::Encryptor`)构造 `qAAB`+密文的 token,`decrypt_token` 还原;三种 base64 变体;`pick_main_url_entry` 评分;`extract_key_seed` 三种位置;**真实样本**:若 `docs/analysis/reference/` 或后续抓包得到真实 `fallback_url`+`key_seed`,加入 `tests/fixtures/video_token_1.json`(`#[ignore]` 直到有样本)。
- [ ] **Step 2: 实现 + Commit** `feat(doubao): local video token decryption`

### Task 3.2: 视频时长设置 + 参考图上传

**Files:** Create `src-tauri/src/doubao/duration.rs`, `doubao/upload.rs`

**Interfaces(Produces):**
```rust
pub async fn set_duration(session: &CdpSession, tracking_key: &str, request_json: &Value, target_secs: u32, cancel: &CancellationToken) -> RelayResult<()>;
// 按 analysis/04 "Step 2":≤40 轮 durationDragPlan;attempt 0 用 endpointKey(Input.dispatchKeyEvent Home/End);fineTuneKey Arrow(检测键盘是否生效,无效则 keyboard_disabled);拖拽 12 步 mouseMoved;dragAttempts ≥8 停;最后 verifyVideoSettings;不匹配→generation_preparing_failed("视频时长未生效:期望X秒,实际Y秒")
pub async fn upload_references(session: &CdpSession, tracking_key: &str, request_json: &mut Value, files: &[PathBuf], cancel: &CancellationToken, on_progress: &(dyn Fn(usize, usize) + Send + Sync)) -> RelayResult<()>;
// Page.enable + Page.setInterceptFileChooserDialog{enabled:true};对每个文件:订阅 Page.fileChooserOpened(拿 backendNodeId)→ run('fileInput') → 优先 DOM.setFileInputFiles{backendNodeId} 否则 DOM.querySelector 'input[data-shenlang-reference-upload="true"]' → 都无→generation_preparing_failed("参考图上传入口已变化");request_json.uploadedCount=i+1;run('uploadReady')(70s);结束后 setInterceptFileChooserDialog false
```
- [x] **Step 1: 测试(mock WS)** — durationDragPlan 依次回 `{matched:false,endpointKey:"End",duration:4}`→`{matched:false,fineTuneKey:"ArrowRight",duration:9,startX:..}`→`{matched:true}`:断言发出 Input.dispatchKeyEvent End、ArrowRight 且最终 verifyVideoSettings 被调;fileChooserOpened 事件 → 断言 `DOM.setFileInputFiles` 带 backendNodeId 和文件路径。
- [x] **Step 2: 实现;接入 Task 2.3 `submit` 的两个占位;Commit** `feat(doubao): video duration and reference upload`

### Task 3.3: 视频捕获三条路径

**Files:** Modify `doubao/capture.rs`, `doubao/download.rs`

- 路径 1 `Network.responseReceived` mimeType `video/*` 且 requestId ∈ send 后集合且 host ∈ 视频 CDN 白名单(analysis/05)→ `VideoCandidates`;含 `lr=unwatermarked` → 直接作为 original(`LrUnwatermarked`)。
- 路径 2 feedback stage==result 且未触发 → 每 5s `run('requestVideoDownload')`,旁听新 video 响应。
- 路径 3 `read()` 结果含 `fallbacks[{vid,api}]` → 每 vid 10s 内不重复 → `api` 必须 `https://vas-[a-z0-9-]+\.snssdk\.com` 无 userinfo → reqwest GET(10s,≤8MB,不跟随重定向)→ `video_decoder::decode` → `unwatermarked_url` → original(`FallbackDecrypted`);Rust 解密失败时 fallback 调注入的 `shenlangDoubaoVideoDecoder.decode()`(记 warn)。
- 下载:`download_video`;非 `lr=unwatermarked` 的 URL → `generation_watermark_not_removed`;`ffprobe` 不可用,duration_ms 通过解析 mp4 `mvhd`(timescale/duration)得出(写 `mp4_duration_ms(bytes)->Option<u64>` 并单测,用一个 ≤1KB 的合成 ftyp+moov/mvhd 字节)。
- [x] **Step 1: 测试** — URL 白名单正反例;`mp4_duration_ms`;mock 场景三条路径各一;`generation_watermark_not_removed`。
- [x] **Step 2: 实现;runner 的 video 分支(20min deadline、prompt 前缀);Commit** `feat(doubao): video capture with three unwatermark paths`
- [~] **Step 3: 真机验收 UC-20/21/27**(部分完成:UC-20 已生成,但确认是手动发出的;UC-21 场景 A/B(路径 1/2)仅 mock 覆盖、UC-27 参考图未跑,移交 Task 4.1/UC-22 真机补验;审查修复:视频下载可被取消、mp4 解析防溢出并补宽高、路径 3 Rust 解出非视频 CDN 地址时改用页面脚本):生成一段 Seedance 2.0 Mini 12s 视频(按 Task 2.1 结论选可用模式),确认 mp4 可播放(`open`)且 `watermark_source` 记录;若得到路径 3 真实样本,立即存 `tests/fixtures/` 并去掉 Task 3.1 的 `#[ignore]`。记入 10-live-probe 文档。

---

# Phase 4 · 并发 + 多轮确认 + 冲突检测

验收:UC-22 ~ UC-26, UC-28, UC-30 ~ UC-33。

### Task 4.1: 多轮确认完整态(auto/manual、降级、超时)+ Webhook

**Files:** Modify `doubao/capture.rs`, `tasks/runner.rs`;Create `tasks/webhook.rs`;Cargo 增 `hmac = "0.12"`

- capture:`videoFollowup` 发出 followup 后 `Stage(Sent,"已自动点继续")`;auto 失败 → 降级 manual + `NeedsUserConfirm`;membership 自动失败 → `auto_use_membership=false` + `NeedsUserConfirm`;manual 下 confirm_rx 为 true 时下一轮以 `userConfirmed=true` 调 `videoFollowup`。
- webhook:`pub async fn deliver(client, task: &Task, event: &str, url: &str, token: &str, storage) `:body 按 04 "Webhook 回调规格";`X-Doubao-Relay-Signature: sha256=<hex>`,`X-Doubao-Relay-Timestamp`,HMAC-SHA256(key=token, data=timestamp+"\n"+body);30s 超时;重试 1s/10s/60s;每次写 `webhook_log`;最终失败 task.details 记 `callback_failed`。
- [x] **Step 1: 测试** — 签名向量(固定 token/timestamp/body → 预期 hex,用 Python `hmac` 预先算);axum mock 第 1、2 次 500 第 3 次 200 → 3 条 webhook_log 且最后 http_status 200;全失败 → 4 条(含首次)且 `callback_failed`。manual 流程状态机测试(扩展 Task 2.7 的 mock backend)。
- [x] **Step 2: 实现 + Commit** `feat(tasks): confirmation modes and webhook delivery`

### Task 4.2: 并发队列硬化 + 用户干扰检测 + 崩溃处理

**Files:** Modify `tasks/scheduler.rs`, `doubao/connection.rs`, `doubao/capture.rs`

- 账号切换期间 `switch_lock` 阻塞所有新 runner 启动(08 "账号互斥");`account_busy`:`POST /accounts/switch` 目标账号有 running 任务 → 409。
- 干扰检测:generating 期间每 10s:chat target 数 ≠1 → `generation_user_interrupted{interruption:"multiple_chat_pages"}`;输入框内容 ≠ 预期(evaluate `findInput().innerText`,空或等于 prompt 视为正常)→ warn 日志 + `task_updated.message="检测到输入框被修改"`(不终止)。
- `ConnEvent::Crashed` → `scheduler.fail_all_running("doubao_crashed", …)`;`DoubaoStatus.cdp_connected=false` 时 `/generate` 返回 503 `doubao_not_ready`。
- 超时扫描:pending>1h→`queue_timeout`(code 字符串按 08 表,写入 11 错误码表缺失项时在 error.rs 补 `queue_timeout`、`preparing_timeout`、`task_timeout`、`confirm_timeout` 别名映射到 `generation_*`)。
- [x] **Step 1: 测试** — 5 个同账号任务 queue_position 1..5 且串行;3 不同账号并发 3,第 4 个等待;51 个 → 429;模拟 Crashed → 所有 running 变 failed/doubao_crashed;切换中提交 → 等待切换完成后才 preparing。
- [x] **Step 2: 实现 + 真机 UC-24/25/28/31/32 + Commit** `feat(tasks): concurrency hardening and interference detection`

---

# Phase 5 · macOS 适配收尾 + 设置/日志页 + 跨平台一致性

验收:UC-40 ~ UC-42, UC-07。

### Task 5.1: 设置页 + 日志页 + 调试面板

**Files:** Create `frontend/src/pages/settings.tsx`, `pages/logs.tsx`;Modify `commands/app.rs`(save_config 热重载 logging/ui;port/token 变更 → `needsRestart`),`commands/log.rs`

- 设置页按 06 "页面 5"全部分组;"复制 token"后 60s 清空剪贴板(`tauri-plugin-clipboard-manager`,capability `clipboard-manager:allow-write-text`);"重新生成"token 立刻生效(HTTP auth 读 `Arc<RwLock<Config>>`);"立即清理"调 `maintenance_gc` 命令。
- 日志页:订阅 `log_entry` 实时追加;级别过滤;导出 ndjson(`tauri-plugin-dialog` save)。
- 调试面板:设置页勾选"开启 CDP 日志" → `cdp_log_entry` 事件 + `cdp-debug.log` 文件。
- [ ] 测试:view-model `curlExample(port, token)`、`maskToken(token)`;Rust `save_config` 返回 needsRestart 真值表。Commit `feat(ui): settings and logs pages`

### Task 5.2: 隐私声明 + 关闭最小化托盘 + 跨平台路径验收

**Files:** Modify `lib.rs`(首次启动 `metadata.privacy_accepted` 为空 → 前端弹全屏声明,确认后写入),`tauri.conf.json`/Cargo(`tray-icon` feature),`frontend/src/app.tsx`

- 托盘:图标 + 菜单(显示主窗口 / 退出);`minimize_to_tray_on_close=true` 时拦截 close → hide;退出走关闭流程(03 "关闭流程":停 HTTP、pending/preparing 取消、generating 等 10s、关 CDP 不杀豆包、flush)。
- Windows 侧:在 Windows 机器(或 CI)跑 UC-01/02/04/10/13/40/41;本机只做 `cargo check --target x86_64-pc-windows-msvc`。
- [x] Commit `feat(app): privacy notice, tray, graceful shutdown`

---

# Phase 6 · 稳定性 + 打包

验收:UC-50 ~ UC-54, NF-01 ~ NF-05。

### Task 6.1: 断线重连硬化 + 24h 稳定性

- 心跳失败 → 2s 后重连 → 20s 内未成功 → `connect(true)`(重启豆包)并发 `doubao_status_changed`;重连成功后 runner 中正在 capture 的任务继续(capture 内部用 `conn.chat_session()` 每次取最新 session)。
- 内存:`LogBuffer` 上限 2000;CDP broadcast 4096 + lagged 处理;`streamBindingBacklog` ≤64 条/1MB。
- `tests/soak.sh`:提交 100 个任务(用 mock 或真豆包低配额模式——真豆包仅跑 10 个),记录 RSS(`ps -o rss`)与 `GET /tasks/{id}` P99(`hyperfine` 或 curl 循环);写入 `docs/analysis/10-live-probe-macos.md` "Phase 6 稳定性"。
- [x] Commit `fix(doubao): reconnect hardening`

### Task 6.2: CI + 打包 + 签名脚本 + 发布 checklist

**Files:** Create `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `scripts/sign-macos.sh`, `scripts/notarize-macos.sh`, `scripts/sign-windows.ps1`, `scripts/bump-version.sh`, `CHANGELOG.md`, `LICENSE`(MIT)

- ci.yml:matrix macos-latest + windows-latest:fmt/clippy/test/typecheck/frontend test。release.yml:tag `v*` → `cargo tauri build --target universal-apple-darwin` / `x86_64-pc-windows-msvc` → 签名(secrets 存在时)→ GitHub Release。
- 本机 `cargo tauri build --debug` 产出 DMG,`open` 安装验证启动;版本号三处同步脚本。
- README 增 "安装与使用" + API 快速开始(来自 04 示例流程)。
- [x] Commit `chore: ci, release pipeline, packaging`

---

## 已知问题 / 待定(执行中追加)

- (Task 1.2)真豆包 2.32.5 macOS 实测(详见 `docs/analysis/10-live-probe-macos.md`):账号行靠 `[data-testid="avatar_menu_account_row"]` 判定(无头像账号没有 `<img>`);`closeMenu` 改为 async(Escape 后 aria-expanded 约 50ms 才翻转);`expandSidebar` 优先 `siderbar_closed_status_btn`;只有 1 个账号时菜单没有账号行;点击登录态失效的账号会弹独立 `client-login-modal` page target 并把该账号移出列表,主页面不 reload。本机现仅剩 1 个有效账号,"切换成功 → reload → 重连"路径未能在真机验证,Task 1.8 以 mock 覆盖。
- (Task 1.2 审查修复)`accounts.js` 1.1.0:`findAccountButton` 改为位置唯一命中优先、`<img>`/套餐文本只作收窄(真机摘掉 `<img>` 后仍命中);**当前账号为默认头像时按钮是否无 `<img>` 未能真机证实**,列为 Task 1.7 风险,其 `live_` 测试在 `current().status!=='logged_in'` 时打印 `probe()` 全量(含 `candidates`)。另:版本守卫(`version`)、`menuOpen` 无按钮兜底改看菜单根/退出登录、`switchTo` 区分 `not_found`/`ambiguous`、`expandSidebar` 去 aria-label、探针加 `--timeout` 看门狗。DOM 测试在 `frontend/src/accounts-script.test.ts`(happy-dom)。
- (Task 1.3)`impl From<crate::cdp::CdpError>` / `impl From<crate::platform::PlatformError> for RelayError` 依赖 Task 1.4 / 1.5 才定义的类型,本 Task 不能编译它们,**改由 Task 1.4 和 Task 1.5 在各自模块里补上**(映射规则不变:Closed/Connect→`doubao_cdp_disconnected`,Timeout→`doubao_not_ready`,JsException→`doubao_ui_changed`)。`RelayError::doubao_cdp_disconnected()` 构造器已在 1.3 提供。另:日志文件名为 `doubao-relay.YYYY-MM-DD.log`(tracing-appender 的 prefix.date.suffix 格式,spec 写的是 `doubao-relay-YYYY-MM-DD.log`);`logging::init` 的文件保留数固定 7(与 `file_retention_days` 默认一致),按大小滚动(`file_max_size_mb`)tracing-appender 不支持,留待 Task 5.1 决定是否换实现;`init` 重复调用返回同一个 `LogBuffer`(测试进程内多次 init 不报错)。
- (Task 1.4)`CdpSession` 增加 `connect_with_debug(ws_url, target_id, Option<broadcast::Sender<CdpLogEntry>>)`(`connect` = 无调试日志),Task 1.8 的 `cdp_debug` 经此传入;`CdpLogEntry` 定义在 `cdp/mod.rs`,`None` 字段序列化时省略。`impl From<CdpError> for RelayError` 已在 `cdp/mod.rs` 补上:Closed/Connect→`doubao_cdp_disconnected`,Timeout/Http/InvalidTarget→`doubao_not_ready`,JsException→`doubao_ui_changed`(details.js_exception),Protocol/Deserialize→`system_internal_error`。`Runtime.evaluate` 表达式不能用顶层 `await`(无 replMode),后续 Task 注入脚本返回 Promise 即可。真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 1.4"。
- (Task 1.5)Windows 交叉 `cargo check` 结果:`--target x86_64-pc-windows-msvc` 对整个 crate 失败于 `ring`(reqwest rustls-tls 的传递依赖)build.rs 找不到 `assert.h`(本机无 MSVC 头),与 platform 代码无关、不阻塞;改用 `--target x86_64-pc-windows-gnu`(homebrew mingw-w64)整个 crate `cargo check` 与 `cargo clippy --all-targets -- -D warnings` 均通过,另用只含 `error.rs`+`platform/*` 的临时 crate 对 msvc target check+clippy 通过。实现差异:`kill_doubao`(Windows)按计划用 `taskkill /IM Doubao.exe` → 等 800ms → `taskkill /F /T`,不引入 `windows-sys`;`log_dir`(macOS)按 spec 10 用 `~/Library/Logs/doubao-relay`(spec 07 文件布局画的是 `<app data>/logs/`,以 10 为准);`launch_doubao_with_cdp`(macOS)`open -a` 的目标是从 exe 回推的 `Doubao.app` 绝对路径(找不到包目录才用名字 `Doubao`);`PlatformError`→`RelayError` 映射:DoubaoNotFound→`doubao_not_found`(details.searched_paths)、LaunchFailed→`doubao_launch_failed`、PortOccupied→`doubao_cdp_port_occupied`(details.owning_process/port)、RegistryError→`platform_registry_error`、KillFailed→`platform_process_enum_failed`(11-error-codes 无专门的 kill 失败码)、Io→`system_io_error`。`macos.rs`/`windows.rs` 两个模块在所有平台都编译,纯函数(`is_doubao_exe_path`、`parse_lsof_pids`、`parse_netstat_listening_pid`、`app_bundle_from_exe`)在 macOS 上也有单测;只有 `MacosPlatform`/`WindowsPlatform` 实体按 cfg 编译。真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 1.5"。
- (Task 1.5 审查修复)`kill_doubao`(macOS)的 `osascript quit` 加 5s 超时(`GRACEFUL_QUIT_TIMEOUT`,子进程 `kill_on_drop`),超时视为优雅退出失败直接进入 SIGTERM;`is_doubao_running` / `wait_until_gone` 排除 launchd 托管的 `XPCServices/*.xpc` 与 crashpad 瞬态进程(`is_doubao_transient_helper_path`),`signal_all` 仍发给全部豆包进程;`is_doubao_exe_path`(macOS)改为路径分量匹配 `Doubao.app`(不区分大小写)或 `*豆包*.app`,与 `find_doubao_exe` 第 3 级一致,`app_bundle_from_exe` 取最外层包;`find_doubao_exe` 的 `spawn_blocking` JoinError 改映射为 `PlatformError::Io`(`platform::join_error`),不再混进 `DoubaoNotFound` 的 searched_paths(`DoubaoNotFound(String)` 仍照 spec 10 保留 String);`doubao_not_found` 文案改为"未找到豆包安装,请先安装豆包桌面端后点击重连豆包"(没有"指定豆包路径"的设置项);Windows `listening_process_exe` 改为 `netstat -ano`(不加 `-p TCP`,否则漏 TCPv6);trait 文档标明 `listening_process_exe` / `doubao_version` 为阻塞调用,Task 1.6/1.8 在 async 上下文调用需 `spawn_blocking`。
- (Task 1.6)真豆包 2.32.5 macOS:计划的发送按钮 selector(`button[aria-label*="发送"],button[type="submit"]`)命中 0,实际是 `button[data-testid="chat_input_send_button"]`(`.send-btn-wrapper` 内,`aria-label=""`),且**只在输入框有文字时渲染**;`compat::ui_compat_check` 改为先用 `SEND_BUTTON_SELECTOR`(testid 优先,计划两条作兜底)直接找,没命中且输入框为空时临时 `insertText("x")` 再 `selectAll+delete` 还原来确认(输入框非空不动用户草稿,找不到即 `warn`)。`ensure_doubao_started` 签名照计划用 `&dyn Platform`,阻塞的 `listening_process_exe` 用 `block_in_place`(多线程运行时)/直接调用(单线程运行时)而非 `spawn_blocking`(借用不满足 `'static`);另加 `ensure_doubao_started_with(platform, port, launch_timeout)` 供测试缩短 20s;`/json/version` 通但查不到监听进程时(lsof 失败)退回看 `/json/list` 是否含 doubao-chat 再判 `AlreadyConnected`/`doubao_cdp_port_occupied(owning_process="unknown")`;`/json/version` 不通但端口被非豆包进程占着时直接报 `doubao_cdp_port_occupied`,不启动豆包;重启分支先 `find_doubao_exe` 再 `kill_doubao`(macOS 查找的第一来源是已运行进程)。`ui_compat_check` 永不 Err:CDP/脚本失败时 `status="unknown"`。`doubao/mod.rs` 提供 `ACCOUNTS_SCRIPT`(include_str)/`ACCOUNTS_SCRIPT_VERSION`/`CDP_HOST` 供 Task 1.7/1.8 复用。真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 1.6"。
- (Task 1.7)`parse_entries` 的账号行判定改为 `is_account_row || has_avatar`(计划写 `!has_avatar` 过滤;真机无自定义头像的账号行没有 `<img>`,照计划会丢账号),`RawEntry` 增加 `#[serde(default)] is_account_row` / `test_id`,`CurrentProbe` 增加 `#[serde(default)] raw`(单账号退化为 `[current]` 时用按钮原文作 `label`)。另加 `read_accounts_with(session, ReadTiming)` 让测试缩短各阶段等待(默认值与计划一致:400/650/650/400ms、≤8 轮),`read_accounts` = 默认节奏。`ensure_injected` 一次 evaluate 完成"按版本注入 + 返回版本",版本不符 → `system_internal_error`(details.actual)。侧栏折叠时 `openMenu()` 真机仍可打开菜单,只在 `probe().buttonFound=false` 时才 `expandSidebar()`。`closeMenu()` 返回 `stillOpen=true` 才走 `Input.dispatchKeyEvent` Escape 兜底(真机未触发)。真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 1.7"。
- (Task 1.8)`DoubaoConnection` 实现差异:① 构造函数 `new(platform, host, port)` 本身就带 host,没有再加 `new_with_host`;另有 `new_with(platform, host, port, cdp_debug, ConnTiming)` 传 CDP 调试通道与时间参数(`ConnTiming{heartbeat_timeout 5s, crash_after 20s, account_cache_ttl 2s, read: ReadTiming, switch: SwitchTiming}`,测试用小值)。② 账号切换拆成两层:`accounts::click_account(session, target, SwitchTiming)`(注入 → 定位按钮 → openMenu → 100ms × ≤40 轮 hoverMore+switchTo 直到 `clicked`;`ambiguous`→`account_ambiguous`;轮数用尽→`account_switch_failed(reason=account_row_not_clickable)`)与 `accounts::probe_current(session)`(带注入守卫的 `current()`,页面 reload 后脚本丢失也能一次 evaluate 读到),`DoubaoConnection::switch_account` 负责校验 / 验证循环 / 事件;计划写的"accounts.rs 新增 switch_account"即这两个函数。③ 验证循环每轮先查 `/json/list`,出现 `client-login-modal` page target 立即 `account_switch_failed(reason=login_required)`(Task 1.2 结论);`CdpError::Closed/Timeout/Connect` → `chat_session()` 重连后继续。④ `switch_lock` 同时串行化 `accounts()`(读列表也要操作菜单),`switch_account` 内部用不加锁的 `accounts_locked`;另有 `connect_lock` 串行化 `connect_sessions`(心跳与切换验证可能同时重连)。⑤ `connect_sessions` 对未关闭且 target id 未变的 chat session 直接复用(不重复建连);`connect(false)` 返回 `AlreadyConnected`(复用或首次连接)/ `Reconnected`(之前有过 session 且这次新建)/ `Started`/`Restarted`(按 bootstrap);`connect(true)` = find_doubao_exe → kill → launch → wait_for_cdp(20s) → `Restarted`。⑥ 心跳:从未连接过时 `heartbeat()` 直接返回 false 不碰豆包;失败时立刻关老 socket → `connect_sessions` → 复测一次(页面 reload 这种"断开但豆包还在"的情况一个心跳周期内恢复),复测失败才保持断开;`Crashed` 只在"自首次失败起 ≥20s 且 `is_doubao_running()==false`"时发一次,之后每轮心跳仍静默重连(用户手动重启豆包后自动恢复)。`StatusChanged` 只在除 `last_check_at` 外的字段变化时发(避免每 10s 刷一次 UI)。⑦ `Network.enable` 的 Protocol 错误只 warn 不致命(真机实际成功,见 docs/analysis/10 Task 1.8);`Page.setWebLifecycleState`/`Emulation.setFocusEmulationEnabled` 不发。⑧ `DoubaoStatus` 比 spec 05 多 `reconnecting`(计划已列)与 `ui_compat_details.status/checked_at`(沿用 `UiCompat`),IPC 层照原样转 camelCase 即可。本机只剩 1 个有效账号,`live_switch_roundtrip` 自动 SKIP,"切换成功 → reload → 重连"仅 mock 覆盖(mock 豆包在同一端口同时服务 `/json/*` 与 WebSocket,以通过 `validate_target` 的端口校验)。
- (Task 1.8 审查修复)① `live_` 测试:`doubao::LIVE_MENU_LOCK`(`#[cfg(test)]` 进程级 `tokio::sync::Mutex`)串行化所有要操作真豆包账号菜单的 `live_` 测试(`accounts::live_read_accounts`、`connection::live_connect_status_and_heartbeat`、`connection::live_switch_roundtrip`),`cargo test live_ -- --ignored --nocapture` 命令不变。② **真机新发现**:`closeMenu()` 返回后 0ms 内再 `openMenu()` 菜单不会出现(退出动画期间 pointer 事件被丢弃,≥150ms 后才正常;探针 4/4 复现),`switch_account` 的"读列表关菜单 → 立刻点账号"正走这条路径;`ReadTiming` 增 `after_close`(默认 300ms),`read_accounts_with` 关菜单后等它再返回,`read_accounts` 真机 ~3.4s。③ 心跳失败分支每轮查 `is_doubao_running()`,`DoubaoStatus.running` 在进程退出后的下一次心跳即为 false(`reconnecting=false`、跳过重连),`Crashed` 仍 ≥20s 且 `!running` 只发一次;UC-06 的 15s 要求由 `running` 字段满足。④ `accounts(false)` 先锁外查缓存、未命中再拿 `switch_lock`(锁内二次检查),切换进行中 `GET /accounts` / `submit` 校验秒回缓存。⑤ `verify_switch` 每一步(`/json/list`、`probe_current`)都以验证 deadline 为界(`tokio::time::timeout`),`Runtime.evaluate` 挂起不再把 20s 窗口拖到 50s。⑥ 心跳与切换验证只关自己手里失效的那条 session(`close_sessions_if_chat_is`,`Arc::ptr_eq`),不再互相关掉对方刚建好的连接。⑦ `accounts.js` 1.1.1:`switchTo` ambiguous 时返回 `labels`,`account_ambiguous.details.matches` 为真实的各行原文(缺省回退 `count` 份目标名)。⑧ `outage` 锁改 `unwrap_or_else(|p| p.into_inner())`,`Network.enable` 参数改 `json!` 直接构造。⑨ **未做、留给 Task 6.1**:spec 09 "安全校验(每次重连都要重跑)"要求重连时重跑 9707 监听进程可执行路径校验,目前 `connect_sessions`(心跳 / `chat_session` 重连)只做 `select_pages`/`validate_target`(ws url loopback+端口、scheme/host 白名单),`connect(true)` 重启后 `wait_for_cdp` 也只看 `/json/version`;Task 6.1 断线重连硬化时在 `connect_sessions_inner` 的重连路径加 `listening_process_exe` + `is_doubao_exe_path`,失败返回 `doubao_cdp_port_occupied`。⑩ **IPC 类型约定(Task 1.9/1.10 照此定义 `api.ts`)**:`DoubaoStatus.ui_compat_check` 为 `'ok'|'warn'|'error'|'unknown'`(spec 05 没有 `'unknown'`;CDP/脚本失败时 `ui_compat_check` 为 `unknown`),`reconnecting: boolean` 与 `ui_compat_details.{status,checked_at}` 也一并转 camelCase 暴露。
- (Task 1.10)① **验收手段**:macOS 未授 screencapture,但 Accessibility 已授权,真 relay 窗口(WKWebView)用 `osascript` System Events 读 AX 树 / 点按钮验收;截图用 Playwright 打开 vite `localhost:1420` + 注入假 `__TAURI_INTERNALS__`(`scripts/ui-harness.playwright.js`)逐状态截到 `docs/analysis/screenshots/phase1/`。本机仅 1 个账号,UC-03/04/05 只有 mock 覆盖,真机 SKIP;UC-06 真机通过(pkill 后 9.7s 显示"豆包已退出,等待重启",点"重连豆包"4.6s 恢复)。② **真机新发现并修复**:relay 拉起豆包后 CDP 在 0.9s 就绪但页面 DOM 还要几秒才渲染,`connect()` 紧接着自检得 `error`、首次 `account_list` 报 `doubao_ui_changed`;`DoubaoConnection::connect` 在 `Started`/`Restarted` 后先轮询 `PAGE_READY_EXPRESSION`(≤15s,500ms 一次;`ConnTiming.page_ready_timeout/page_ready_poll`)再自检,mock 覆盖,真机 Started 路径修复后未复跑(只允许杀一次豆包),留 Task 6.1 复验。③ 前端约定:`api.ts` 的 `onEvent` 返回**同步**取消函数(内部等 `listen` 完成后再 unlisten),便于 `useEffect` 直接 return;`asRelayError` 非结构化错误归 `system_internal_error`;账号页对 `doubao_not_logged_in` 把列表置为 `{current:null, accounts:[]}` 以显示"尚未登录"空状态,其它 `doubao_*` 错误置 `null` 走"未连接 / 还没读到"空状态,不弹 toast;`doubao_status_changed` 的 `cdpConnected=false` 时清空列表,重新连上后账号页自动重读;`account_switched` 事件负责成功 toast,`account_switch` 返回后再 `account_list` 一次(命中 2s 缓存)。`state.ts` 多了 `reconnecting`(重连按钮 loading)与 `apiListening`(Phase 2 HTTP server 接入前固定 false),`view-model.ts` 多了 `doubaoStatusText`。`Cmd+R` 等快捷键未做(spec 06 快捷键表留 Task 5.1)。④ `spec 05` 没列的 `doubao_crashed` 事件前端也订阅(错误 toast)。
- (Task 1.10 审查修复)① `confirm-dialog`:快捷键策略抽成纯函数 `keyAction(key, target, dialog)`——Enter 落在对话框内的按钮上(取消 / 确定)不处理、交给按钮自己的 click,其它位置的 Enter 才算"确定"并 `preventDefault`(防止对话框外仍持焦点的触发按钮被再次激活);打开时用 `useLayoutEffect` 同步 `focus()` 确认按钮(`autofocus` 被 Chromium 拒绝)并装监听,真机 Chromium(Playwright)验证:Shift+Tab 到"取消"按 Enter 不切换、焦点留在触发按钮上按 Enter 恰好切换一次且不重开确认框。`switchTo` 在 confirm 返回后再查一次 `switching` 防重入。② 账号页读取失败分类:`doubao_cdp_disconnected` / `doubao_not_running`,或状态栏已显示未连接时的任何 `doubao_*`,静默走空状态;**已连接**时的 `doubao_ui_changed` / `doubao_not_ready` 等必须可见——错误 toast 常驻(message + `details.failed_selectors`)且空状态标题改为"读取账号列表失败"并显示原因(`state.accountsError`,成功读取 / `account_list_changed` 清掉)。③ `doubao_status_changed` 的 `cdpConnected=false` 在 `switching!==null && running` 时不再清空列表(切换期间豆包 reload 的短暂断线保留"切换中…"行);进程退出仍清空。④ `loadAccounts` 读取进行中收到 `force` 请求改为记 `pendingForce`,本次结束后再强刷一次,不再静默丢弃。前端测试 30 → 60(`confirm-dialog.test.tsx` 用 preact `render` + `act` 于 happy-dom;`pages/accounts.test.ts` / `app.test.ts` mock `@tauri-apps/api`)。
- (Task 2.5)`download.rs`:mock server 用 tokio `TcpListener` 手写 HTTP/1.1(与 1.4/1.6 测试同风格),**没有**引入 axum(留给 Task 2.8 统一加,避免 Cargo.toml 冲突)。`generation_download_failed` 的 details 在 spec 的 `{url,status}` 之外多一个 `reason`(`redirect|http_status|too_large|too_small|not_mp4|unknown_image_format|unsupported_image_format|corrupt_image_header|timeout|connect_failed|transport_error`);`detect_image_format` 单独调用时 `url/status` 为 null。另加 `download_image_with_limit` / `download_video_with_limit`(可配字节上限,测试用小上限验证流式截断;`download_image`/`download_video` = 100MB / 1GB 默认)。图像 ≤100MB 先收进内存再按内容定扩展名;视频流式写 `video_1.mp4`,前 12 字节到齐即校验 `ftyp`,任何失败删除半成品;`Content-Length` 超限不读 body 直接拒绝。写盘 `ErrorKind::StorageFull` → `system_disk_full`,其它 IO → `system_io_error`。`new_client()` UA 固定为 Chrome/147(与豆包 2.32.5 内核同代),所有请求带 `Referer: https://www.doubao.com/`。
- (Task 2.1)对话/工作模式结论(详见 `docs/analysis/10-live-probe-macos.md` "Task 2.1"):① 还原 `doubao-generation-page.js` 字符串表后 `prepare` 实际有**三条**分支,不是两条:`work`(豆包"工作"= 通用 agent 输入区,参数只是文字提示)/ 视频 `conversation`(仅 Seedance 2.0 Mini 11–15s,不选模型不设参数,prompt 需 Rust 拼前缀 `请使用 Seedance 2.0 Mini 生成{d}秒视频，比例{r}。\n`)/ **其它任何值**(记作 `skill`)= 进入"视频生成"技能用 DOM 菜单选模型 + 比例 + 时长滑块;图像不看 `submissionMode`,一律进"图像生成"技能。Review Focus 1"其它模型必须走 work"**不成立**。决定:图像传 `conversation`、视频传 `skill`,prompt 不加前缀;`SubmissionPlan.submission_mode` 取值 `conversation|skill`(计划写的 `conversation|work` 作废),`work` 与 Mini-conversation 路径均不启用。② 没有独立"AI 创作"页:技能在同一 `chrome://doubao-chat/chat` 页输入栏(`skill_bar_button_3` 图像生成 / `skill_bar_button_17` 视频生成),脚本按文本点击;模型菜单是 DOM `[role=menuitem]`(四个视频模型、三个白名单图像模型体验套餐全可选,无原生菜单 fallback 需求)。③ **真机新发现**:技能态下"新对话"不退出技能,图像↔视频互切、技能→普通对话的 `prepare` 都在 12s 后失败(`豆包控件未唯一匹配` / `新对话尚未准备好`),Task 2.3 的 `submit` 必须在 `prepare` 前先点 `[data-testid="skill_input_exit_button"]`(若可见)并等 `guidance-skill-bar` 出现。④ 计划的 `run('fill', {...req, prompt:''})` 清空不可行(`fill` 要求输入框为空,抛 `豆包输入框已变化`),清空用 `compat.rs` 的 `selectAll+delete`。⑤ **比例白名单与真机不符**:图像菜单 `自动 9:16 2:3 3:4 1:1 4:3 3:2 16:9`(无 21:9),视频菜单 `自动 3:4 4:3 9:16 16:9 1:1 21:9`(无 3:2 / 2:3);spec 01 的统一列表会在 `prepare` 阶段以 `未找到所选生成比例` 失败,Task 2.2 已按统一列表实现,建议后续按类型拆分 `validate_request`(`request_invalid_enum.allowed` 按类型)。⑥ 视频时长滑块是 `[role=slider]` span,`aria-valuemin=0`,`aria-valuemax` 2.5 为 26(4–30s)、2.0/Fast/Mini 为 11(4–15s),`aria-valuenow` 是索引(秒 = 索引 + 4);`durationState.duration` 单位不统一(滑块不可见且摘要已匹配 → 秒 + `fromSummary:true`;否则滑块原始索引),Task 3.2 只看 `durationDragPlan.matched` + `verifyVideoSettings`;`prepare(skill)` **有两种成功返回**:摘要那一瞬间与请求不一致(含首次进技能摘要为 null)→ `{needsDurationInput:true}`(比例已选好,弹层留在打开状态、滑块 focus),摘要已匹配(豆包记住上次设置,常态)→ 直接 `{mode:"video",model,ratio,duration}` 且无弹层;Task 2.3 必须接受两种形状,只在 needsDurationInput 时调 `duration::set_duration`。⑨ 异常映射拆分:`prepare`/`fill` 异常 → `generation_preparing_failed`,`checkSend`/`send` 异常(`豆包发送按钮尚未就绪` 等)→ `generation_submit_failed`(11-error-codes 语义),均带 `details.js_exception`。⑦ `fileInput` 直接找到输入区里已存在的 `input[type=file]`(`accept=".jpg,.png,.jpeg,.webp,.apng"`,multiple)并打 `data-shenlang-reference-upload` 标记,**不触发 `Page.fileChooserOpened`**,Task 3.2 主路径是 `DOM.querySelector` + `DOM.setFileInputFiles`;视频技能没有参考视频入口。⑧ 窗口隐藏(`visibilityState=hidden`)下 `setTimeout` 不节流,prepare→fill→checkSend 869ms 与前台一致,Task 2.3 的"focus 伪装"(`Page.setWebLifecycleState` / `Emulation.setFocusEmulationEnabled`)可省略。
- (Task 2.6)`storage` 实现差异:① 迁移 `migrations/20261016000001_initial.sql` 照抄 07 五张表(加 `IF NOT EXISTS`),表里没有 `progress` / `queue_position` 列,这两个是运行时态,`get_task` 读出恒为 `None`(Task 2.7 的 Updated 事件带 progress 即可,不入库);`PRAGMA journal_mode=WAL` / `foreign_keys=ON` / `busy_timeout=5s` / `synchronous=NORMAL` 在 `SqliteConnectOptions` 上设置,sqlx 用计划原句 `features=["runtime-tokio","sqlite","migrate"]`(默认 features 保留,`migrate!` 宏需要 `macros`)。② `impl From<sqlx::Error> for RelayError` 放在 `storage/db.rs`(同 1.4/1.5 的模块内 From 惯例),映射 `system_database_error`(details.db);库里枚举值无法识别也报 `system_database_error`(details.column/value)。③ `insert_task` 同一事务连 `stage_history` 一起插;`update_task` 只 UPDATE `tasks` 行、不动 `task_stage_history`(阶段走 `append_stage`),行不存在 → `task_not_found`(404);`fail_non_terminal` 置 `status=failed, error_code, error_message, error_details=NULL, completed_at=now`,不追加 stage_history 行。④ `list_tasks` 取 `limit+1` 判断是否有下一页,`status=Some(&[])` 等同不过滤,`limit=0` 按 1 处理。⑤ 新增计划未列的 `WebhookLog { task_id, event, url, attempt, http_status: Option<u16>, response_body, error_message, sent_at }`(与表一一对应,`id` 自增不暴露);`Storage` 派生 `Clone`(pool 内部是 Arc),`pool` 字段 `pub(crate)` 供 `gc.rs` 与测试用。⑥ GC:步骤 1 按 `result_files` 列出的 `relative_path`(只接受 root 下的安全相对路径,含 `..` 的跳过)删文件后再整目录 `remove_dir_all`,`result_files_removed` 统计实际删掉的文件数(步骤 2 级联删掉的也计入);步骤 4(日志按天保留)由 `tracing-appender` 负责,`gc()` 不处理;文件删除失败只 `warn` 不中断,数据库错误才 Err。⑦ `storage/files.rs` 另提供 `ensure_layout` / `db_path` / `is_safe_relative` / `remove_file_if_exists` / `remove_dir_if_exists` 纯函数(Task 2.5/2.8 可复用)。
- (Task 3.1)真实视频 token 样本:暂无(`docs/analysis/reference/` 里只有混淆后的 `doubao-video-decoder.js`,没有抓到过 `fallback_url`+`key_seed`),`fixture_video_token_1_decodes` 以 `#[ignore]` 等待 `src-tauri/tests/fixtures/video_token_1.json`(`{response, api_url, expected_url}`)。替代做法:用 Node 跑原版混淆脚本当 oracle,生成 5 组 token 向量(pepper / salted 直接 / salted 交叉 / offset 0x24 内嵌 iv / offset 0x24 + iv_a)写进单测,Rust 全部解出且与 JS 一致。**密钥派生以原脚本为准,analysis/06 的伪代码有错**:实际是 `A = SHA512(SHA512(seed[..32]) ‖ salt)` → `key_a/iv_a` 对半,`P = SHA512(SHA512(seed) ‖ pepper)` → `key_p/iv_p` 对半;候选顺序(有 magic)`(key_p,iv_p)`→`(key_a,iv_a)`→`(iv_a,key_a)`→`payload[0x24..]` 配 `(key_a, payload[0x14..0x24])`→`(key_a,iv_a)`;无 magic `(key_a,iv_a)`。文档里"sha512(seed[:32]) 直接对半"的派生原脚本解不出(单测有负向量)。差异:Rust 不要求 token 以 `qAAB` 开头才走 key_seed 解密(更宽松,解不出照样 None);`decode()` 的 `pick_main_url_entry` 内部先做 `extract_video_info`(`video_info` / `data.video_info` / 自身,再剥一层 `data`),直接传整个响应即可;`unwatermarked_url(url)->String` 按计划实现为追加/替换 `lr=unwatermarked`(字符串级,不重编码其它 query,签名参数不受影响),但 **reader.js 的 `unwatermarkedURL` 实际是过滤器**(https、无 userinfo、host ∈ 视频 CDN 白名单、`lr=unwatermarked` 才返回 href),Rust 另给 `video_url` / `unwatermarked_video_url` / `is_video_cdn_host` 与之等价(白名单含文档漏掉的 `365yg.com`),Task 3.3 路径 1 的 host 白名单可直接复用。另:Cargo 比计划多加 `sha2 = "0.10"`(SHA-512 必需);`VIDEO_DECODER_SCRIPT`(include_str)放在 `video_decoder.rs`,供 3.3 的 JS 兜底注入;解密结果裁掉首尾 `\t\n\r` 空格(PKCS7 填充值 9/10/13 恰好是放行的空白字符,原脚本会把它们带进 URL)。

- (Task 2.3)`doubao/generation.rs` 实现差异(真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 2.3"):① `SubmissionPlan.submission_mode` 取值 `conversation`(image)/ `skill`(video),不是计划写的 `conversation|work`(Task 2.1 结论),prompt 不加前缀。② `SubmissionIdentity.local_message_id` 是 **`Option<String>`**(计划写 `String`):reader 的 `submissionIdentity` 在只扫到 `message_id` 时返回 `localMessageId:null`,`parse_identity` 要求 `localMessageId`/`messageId` 至少其一;`postData` 交给页面前先跑 `fix_big_integers`(reader 内部 `JSON.parse` 会丢大整数精度)。③ `submit` 在 `prepare` 前多一步 `exit_residual_skill`:`[data-testid="skill_input_exit_button"]` 可见就按五事件链点击并等 `[data-testid="guidance-skill-bar"]` 出现(≤3s,真机 157ms),等不到 → `generation_preparing_failed{reason:"skill_exit_failed", failed_selectors:[…]}`;"focus 伪装"(`Page.setWebLifecycleState`/`Emulation.setFocusEmulationEnabled`)按 Task 2.1 结论省略。④ 另加 `submit_with(…, &SubmitTiming)`(`prepare 60s / 其它步骤 20s / 技能条 3s / identity 窗口 2s`,测试用小值),`submit` = 默认值;错误构造函数 `preparing_failed / preparing_timeout / submit_failed / identity_missing / cancelled` 放在 `generation.rs`(不动 `error.rs`,避免与并行 Task 冲突),`details` 统一带 `reason` + `js_exception`;`prepare` 的 CDP 超时 → `generation_preparing_timeout`,`checkSend`/`send` 超时 → `generation_submit_failed{reason:"timeout"}`。⑤ identity 窗口内没有候选或 reader 全返回 null → `Err(generation_submit_identity_missing)`,`details.sent=true, send_started_at, candidates`(发送已经发生,调用方据此判断);`send` 之后取消 → `generation_cancelled{step:"identity"}`。`checkSend`/`send` 失败或取消后尽力 `selectAll+delete` 清空输入框(`fill` 要求输入框为空)。⑥ 视频 `prepare` 返回 `{needsDurationInput:true}` → `generation_preparing_failed{reason:"duration_input_required"}`,`reference_paths` 非空 → `{reason:"reference_upload_unsupported"}`,均为 Phase 3(Task 3.2)前的占位;直接形状 `{mode:"video",…,duration}` 与请求时长不符 → `{reason:"duration_mismatch"}`。⑦ `postData` 缺失但 `hasPostData=true` 时补发 `Network.getRequestPostData`;`wallTime` 早于 `send_started_at` 1.5s 以上的请求不算候选;background 页 `Network.enable` 失败只 warn、只监听 chat 页。⑧ `reader_bootstrap()` / `RESULT_READER_SCRIPT` / `clear_input_expression()` 放在 `generation.rs` 供 Task 2.4 复用。⑨ Cargo 加 `tokio-util = "0.7"`(`CancellationToken`)。⑩ 真机只干跑到 `checkSend`(`live_submit_dry_run`,不消耗配额):普通对话态 2958ms、残留技能态 194ms 均通过;`send` → identity 真机未验证,留 UC-10。

- (Task 2.4)`doubao/capture.rs` + `doubao/feedback.rs` 实现差异(脚本契约与真机干跑见 `docs/analysis/10-live-probe-macos.md` "Task 2.4"):① `FeedbackState` 定义在 `feedback.rs`(capture 通过 `pub use` 暴露),另加计划未列的 `parse_feedback` / `feedback_step`(图像 `terminalFeedback`,视频按 PS 规则 `videoFollowup`|`videoFeedback`)/ `is_refusal_stage` / `is_confirmation_stage` / `refusal_relay_error`(`RelayError` 形态,`refusal_error` 是它的 `TaskError` 版)/ `not_executed_error` 与 `STAGE_*` 常量;`needs_check` 映射为 None(未识别)。② 另加 `CaptureTiming`(feedback 500ms / 历史 1s / target 2s / 输入框 10s / rendering_unverified 30s / evaluate 30s / 命令 15s)与 `capture_with(…, timing)`,`capture` = 默认值;`DoubaoConnection` 加 `host()` / `port()` 访问器供 `/json/list` 唯一性检查用。③ **图像的"代回复"**(审查修复后):沈浪 `terminalFeedback` 只识别不发送,**`acceptFeedback` 也不发送**(它是 native-shadow 反馈消息的 ingest/分类钩子,`reply` 为字符串直接回 `{stage:"unknown"}`,原实现调它等于空操作);图像在 `checking`/`quota_checking` 且 `should_auto_confirm` 或用户已确认时由 Rust 注入 `capture::accept_reply_expression`——`run('fill', {...payload, prompt:"继续生成"})` + 点唯一可用的发送按钮 + 等输入框清空,返回 `sent:true` 才算发出,`sent!=true` 连续 3 次(`ACCEPT_MAX_ATTEMPTS`)走降级链;每个 stage 只回一次;视频由 `videoFollowup` 自行发送(`sent:true`),`manualRequired:true` → 降级 manual。auto 代回复 / feedback 求值抛 JS 异常**或 `Runtime.evaluate` 超时**且该 stage 为 `checking` → 降级 manual(`quota_checking` → 关自动用会员),清空已确认标记并发一次 `NeedsUserConfirm`。**用户确认一次性消费**:代回复发出后 `user_confirmed` 回落,`checking` → `quota_checking` 各自需要一次确认(`auto_use_membership=false` 时会员配额单独确认),`confirmed()` 边沿触发(runner 每次 confirm 都 `send(true)`)。④ **严格性由 Rust 兜底**:reader 的 `analyze` 在 `requireBinding` 且 `boundConversationId` 未知时一律返回空(新对话永远绑不上),所以 `read_options.requireBinding = bound_conversation_id.is_some()`(`requireIdentity` 恒 true);但 reader 的 `collect` 在 `boundSubmissionId` 为空时会收下整个对话的图,所以 `images` **只在已绑定 submission 后采信**(否则 warn 丢弃)。生成响应 / SSE 直接 `extract_image_originals` 的旁路只在文本含 `local_message_id` / `message_id` / 已绑定 `submission_id` 时采信;SSE 缓冲(≤64 条 / 1MB)在历史 tick 时交给 `streamSubmissionBinding` 绑定。⑤ 历史模板:postData 里 `conversation_id` 必须恰好一个(reader `validateRequest` 同样要求),已绑定对话时只接受同一对话的请求,`/im/chain/single` 优先;重放 `read` 的 `postData` **原样**传(不做 `fix_big_integers`,那会改掉发给服务器的 body),`analyze` 的 body 也原样传(reader 内部自己修大整数)。⑥ 页面 reload:循环顶部发现 chat session 已关闭就 `conn.chat_session()` 重连、重订阅、重发 `Network.enable`、重置脚本注入标记(feedback 的提交状态随页面丢失,此后靠历史重放);`/json/list` 连续 3 次失败 → `doubao_cdp_disconnected`。feedback 表达式带 `typeof <key>==='undefined'` 守卫,返回 `missing:true` 时下次轮询重新注入,**绝不覆盖**已存在的 trackingKey 全局对象(`send` 的提交状态在里面)。⑦ 错误 details:`generation_task_timeout{deadline_ms,last_stage}`、`generation_user_interrupted{interruption,chat_targets}`、`account_quota_exhausted{reason,doubao_message}`(spec 写 `{account}`,capture 不知道账号,Task 2.7 runner 写入 task.error 时补 `account`)、`generation_doubao_refused{reason,doubao_message}`、`generation_rendering_not_executed{reason,limit_ms,doubao_message}`;Err 之前都先发 `CaptureEvent::Refused{kind,message}`。⑧ 视频只发 `VideoCandidates`(`responseReceived` 的 `video/*` + 视频 CDN 白名单——按 PS `videoRequests` 语义匹配 `send_since` 之后发出的**所有**请求,CDN 媒体是独立 GET、不在生成请求集合里;或 `read` 返回的 `videos`)与 `Fallbacks`(`read` 返回的 `fallbacks[{vid,api}]`)事件,不落盘,留 Task 3.3。⑨ 真机只做了只读干跑(`live_capture_expressions_dry_run`:注入 reader + 一次性 trackingKey,`read`/`analyze`/`terminalFeedback`/`videoFollowup`/`streamSubmissionBinding`/输入框检查各求值一次,形状全部符合预期),`send` → 捕获 → 原图的完整链路留 UC-10。
- (Task 2.4 审查修复)三个 major + 四个 minor(详见 `docs/analysis/10-live-probe-macos.md` "Task 2.4 · 审查后修正"):① 图像代回复改为 `accept_reply_expression`(fill + 点发送,`sent:true` 为准,`Job::Accept` 携带 stage,`sent!=true` 累计 3 次降级 / 再问用户;`acceptFeedback` 契约更正为分类钩子,mock 用例 `accept_reply_without_sent_true_is_not_treated_as_sent` 固化);真机只读验证了表达式语法(不存在的 key → `{sent:false,missing:true}`)与 Node vm 8 条分支,真实发送留 UC-10。② 确认一次性消费 + `confirmed()` 边沿触发(`image_/video_confirmation_is_consumed_per_stage`);**Task 2.7 runner 每次 confirm 都要 `send(true)`**。③ 视频 `video_requests`(`video_cdn_response_after_send_is_a_candidate`)。④ `FeedbackState` 增 `manual_required`(`#[serde(default, skip_serializing_if)]`,IPC 形状不变);`CdpError::Timeout` 也降级(`evaluate_timeout_during_auto_accept_degrades_to_manual`,mock 支持延迟回应);0 个 chat page → 计数 3 次后 `doubao_cdp_disconnected{reason:"chat_page_missing"}`(`zero_chat_pages_is_disconnect_not_interruption`);重连时重取 background session(`background_session_is_reattached_after_reload`,mock 支持 chat+background target);`cancel_returns_within_50ms` 从取消时刻起算。

- (Task 2.4 第二轮审查修复)详见 `docs/analysis/10-live-probe-macos.md` "Task 2.4 · 第二轮审查后修正":① **更正第一轮的结论**:reader `read`/`analyze` 的 `feedback` 是消息对象 `{message, messageId, submissionId, createdAt, scope}`(无 `stage`),`run('acceptFeedback', {boundSubmissionId, reply:<该对象>, prompt, modelName, type})` 正是 PS 用来**分类**它的钩子(不发送);capture 恢复调用它(`accept_feedback_payload` / `accept_feedback_expression`、`Job::Classify`,每个 messageId 一次,已绑定且无图时),分类出的确认 stage 记 `history_confirm` 并在 feedback 节拍上按 auto / 用户确认触发 `accept_reply_expression`。docs/analysis/10 的 `analyze` 契约表已更正。② **接口偏差**:`CaptureOptions` 增 `model_name: Option<String>` 与 `generation_requests: Vec<ObservedRequest>`;`generation::ObservedRequest {target_id, request_id}` 新增,`SubmissionReport` 增 `generation_requests`(identity 窗口内见到的生成请求),runner `generating(identity, &report)` 原样传入;capture 的生成请求集合改按 `(target_id, request_id)`。③ `CaptureTiming` 增 `result_settle`(1.5s)/ `result_settle_max`(6s):拿到原图后静默窗口无新 URL 才返回(多图),窗口内撞 deadline 按已捕获完成;完成前补发 `Stage(Result)`。

- (Task 2.7)`tasks/scheduler.rs` + `tasks/runner.rs` 实现差异(本 Task 没有真机动作,见 `docs/analysis/10-live-probe-macos.md` "Task 2.7"):① 豆包侧依赖抽象为 `runner::GenerationBackend` trait(`accounts / ensure_account / submit / capture / download_image`),生产实现 `runner::DoubaoBackend::new(Arc<DoubaoConnection>)`(内含 `download::new_client()`),`Scheduler::new(storage, backend: Arc<dyn GenerationBackend>, config: Arc<RwLock<Config>>)` / `new_with(…, SchedulerTiming)`;**Task 2.8 接 `AppState` 时需把 `AppState.config` 改为 `Arc<RwLock<Config>>`**(`state.config.read().await` 用法不变),并在 `lib.rs` 建 `DoubaoBackend` → `Scheduler` → `spawn_loop()`,订阅 `ConnEvent::Crashed` 调 `fail_all_running("doubao_crashed", …)`,启动时先 `storage.fail_non_terminal("relay_restart", …)`。② 全局并发不用 `Semaphore`,派发时读 `config.concurrency.global_max` 与 `running.len()` 比较(配置可热改);账号锁 `Arc<tokio::Mutex<()>>` 的 `OwnedMutexGuard` 由 runner 包装任务持有到结束;**严格 FIFO**(队头被账号锁/全局名额挡住时不跳过后面的任务,spec 08 原话)。③ `TaskHandle` 比计划多 `account` 与 `abort: Arc<Mutex<Option<TaskError>>>`(`fail_all_running` 先写原因再 `cancel`,runner 收到取消时按它落 `failed`;`fail_all_running` 同时把 pending 全部置 `failed/{code}`,spec 08"所有非终态任务");`status` 用 `std::sync::RwLock`(调度器同步读)。④ 超时归属:`capture` 的 deadline 在等待用户确认期间照样走(Task 2.4 实现,其 mock 用例依赖此行为),所以 runner 用 `PhaseClock` 自己计"生成预算(图 5min / 视频 20min,**不含**等待确认)"与"确认等待 1h"(`generation_task_timeout{deadline_ms,last_stage}` / `generation_confirm_timeout{waited_ms,stage}`),传给 `capture` 的 deadline = 预算 + 2×确认等待(宽松上限);preparing 60s 由 runner 以子 `CancellationToken` 中断 `submit`(等 ≤5s 收尾)→ `generation_preparing_timeout{step:"submit"}`。超时扫描(10s)只处理 pending > 1h → **`generation_queue_timeout`**(spec 08 写 `queue_timeout`,11-error-codes 未列此码,按 `generation_*` 前缀补,details `{waited_ms}`)。⑤ `cancelled` 任务写 `error = generation_cancelled{from_status, step}`(`Task` 没有 spec 08 的 `cancelled_from_status` 字段,`DELETE /tasks/{id}` 的 `cancelled_from_status` 取自 `Scheduler::cancel` 返回值),`completed_at = now`;`completed` 把 `stage` 置 null(spec 04 示例),`failed`/`cancelled` 保留最后 stage(`refused` 等)。`stage_history` 记法:阶段变了记阶段名(`sent`/`checking`/…),否则记状态名(`preparing`/`waiting_user_confirm`/`generating`/终态);终态条目 message = error.message。`confirm` 返回固定 `"generating"`(spec 04 `new_stage`),runner 自己订阅 `confirm_tx` 边沿把 `waiting_user_confirm → generating` 作为独立一步落库(select 用 `biased` 先处理确认边沿,否则与 capture 紧随的 `Stage` 事件竞争会把两步合并)。⑥ `Scheduler` 另加 `get_task(id)`(补 `queue_position`)、`queue_depth()`、`running_count()`、`storage()/backend()/timing()` 访问器与错误构造 `queue_full / task_not_found / task_not_cancellable / task_not_waiting_confirm / upload_not_found / upload_expired / queue_timeout`;runner 侧 `task_error_for`(`account_quota_exhausted` 补 `details.account`)、`cancelled_error`、`result_file_from`(`relative_path = results/{id}/{filename}`)、`MAX_RESULT_FILES = 20`;原图为空 → `generation_result_not_captured`;runner panic → 调度器兜底 `system_task_panicked`。`TaskEvent::Completed(Task)` 按计划保留(`#[allow(clippy::large_enum_variant)]`)。⑦ **不同账号并发的隐患(推断,未真机验证)**:spec 08 允许 `global_max=3` 跨账号并发,但豆包只有一个"当前账号",切换会整页 reload(Task 1.2 实测),正在 capture 的另一任务会丢失页面内提交状态并对错误账号重放历史;mock 用例 `different_accounts_run_concurrently_up_to_global_max` 里并发 runner 互相把当前账号切来切去(a→b→a…)正是这个现象。v0.1 本机只有 1 个账号不触发;**Task 4.2 并发队列硬化时应改为:需要切账号的任务要等 `running` 清空(或把跨账号任务串行化)**,届时同步调整该用例。

- (Task 2.8)`http/` 实现差异(真机只读冒烟见 `docs/analysis/10-live-probe-macos.md` "Task 2.8"):① **`AppState` 扩展**:`config` 改为 `Arc<RwLock<Config>>`(与 `Scheduler` 共享),新增 `storage: Storage`、`scheduler: Arc<Scheduler>`、`shutdown: CancellationToken`(`run()` 在 `RunEvent::Exit` 时取消,HTTP server 优雅关闭)、`api_listening: AtomicBool`(`http::serve` 维护,供 Task 2.9/5.1 状态栏 "● API");`AppState::new` 多 `storage` + `backend: Arc<dyn GenerationBackend>` 两参,`bootstrap` 用 `tauri::async_runtime::block_on(Storage::open(app_data_dir))`(与之后命令 / HTTP / 调度器同一运行时)。`commands::testing::test_state` 因此改为 **async**(sqlx 必须在运行时里开库),另加 `test_state_with(dir, Option<backend>)`、`test_state_blocking`(同步测试用)与 `FakeBackend`(固定账号列表,submit/capture 挂起到取消),Phase 1 的测试调用点已同步改。② `lib.rs`:`spawn_background` 依次 `recover_after_restart`(`fail_non_terminal("relay_restart")`,0 条时不打日志)→ `spawn_task_services`(`scheduler.spawn_loop()`;订阅 `ConnEvent::Crashed` → `fail_all_running("doubao_crashed", …)`;GC 每 1h,**启动时立刻跑第一轮**)→ `spawn_http_server`(独立 `tokio::spawn`,bind 失败只 `error` 日志、应用照常运行)→ 自动连豆包;真机 HTTP 在启动后 ~550ms 监听,早于豆包连接完成(~1.1s)。③ 错误 / 状态码:未知路径返回 **404 + JSON** `request_malformed`(11-error-codes 没有"接口不存在"码,状态码不按前缀表的 400);`GET /uploads/{id}` 不存在**或已过期**都是 404 `upload_not_found`(过期时 details 多 `expired_at`;`upload_expired`(400)只在 `POST /generate` 引用过期上传时由调度器返回);`POST /uploads` 的 `DefaultBodyLimit` = 50MB + 1MB 余量,`file` 字段逐块计数超 50MB 即停、以及 body 超限触发的 multipart 413,都映射成 400 `upload_too_large{bytes, max_bytes}`(`bytes` 取 `Content-Length`,没有则取已收到的字节数);`kind` 非法 → `request_invalid_enum{field:"kind"}`,`kind` 与内容不符 → `upload_mime_not_supported{mime:<实际>, allowed:<按 kind>}`;JSON 体坏 / 非对象 → `request_malformed{reason}`,serde 缺字段 → `request_missing_field{field}`,`type`/`confirmation_mode` 非法值先行给 `request_invalid_enum`。④ **`POST /accounts/switch` 的 `account_busy` 比 spec 更严**:只要有**任何**非终态任务在运行就 409(切换会整页 reload,运行中的任务必丢;spec 字面是"目标账号有进行中任务"),details `{account:<目标>, running_task_ids:[全部运行中]}`;为此 `Scheduler` 新增 `running_task_ids(Option<&str>)`。⑤ `GET /tasks/{id}` 返回体 = `Task` 序列化(`task_id/request/stage_message/result_files/…`)**叠加** spec 04 字段 `account`、`type`、`message`(= stage_message)、失败时 `failed_at`(= completed_at)、完成时 `result{files[{index,filename,mime,bytes,width,height,duration_ms,download_url}], watermark_source}`;两套并存,Task 2.9 的 IPC 可直接 `to_camel(task_detail_json)`。⑥ `GET /tasks` 的 `limit` 夹到 1..=100 不报错(缺省 20),`status` 非法值 → `request_invalid_enum{field:"status"}`,查询串类型错(`limit=abc`)→ `request_malformed`。⑦ `POST /doubao/reconnect` 接受可选 body `{force:bool}`(默认 false:未运行则启动、已运行则重建会话;true 才杀掉重启,与 Tauri 命令一致);`POST /tasks/{id}/confirm` body 可选,只校验是 JSON 对象。⑧ 下载:`ServeFile::new_with_mime` + `try_call`(Range → 206 由它处理),`Content-Type` 以库里 `result_files[].mime` 为准(覆盖按扩展名推断),`Content-Disposition: attachment; filename="task-{id}-{index}.{ext}"`(ext 取落盘文件扩展名);非 completed / 无结果 → `result_not_available`,`expires_at` 已过或文件已不在 → `result_expired`,index 不在列表 → `result_index_out_of_range{requested, available}`,`relative_path` 不安全(含 `..`)→ `result_not_available`。⑨ 鉴权 token 每次请求从 `state.config` 读(设置页重新生成后立即生效);`Authorization` 非 `Bearer ` 前缀(区分大小写)或空 token → `auth_invalid_token`。访问日志 info 一条 `{method, path, status, duration_ms}`,5xx 用 warn,不记 header / query / body。⑩ Cargo 比计划多 `mime = "0.3"`(`ServeFile::new_with_mime` 的参数类型)与 dev-deps `tower`(`util`,`oneshot`)/ `http-body-util`;`tower-http` 的 `trace`/`limit` feature 按计划声明但未使用(访问日志与 body 上限分别用 `middleware::from_fn` 与 axum 自带 `DefaultBodyLimit`)。⑪ 比例白名单仍是 Task 2.2 的统一列表(Task 2.1 要求"按类型拆分"未做,属 `validate_request` 范围,留 Task 2.9 UC-10 前一并处理)。

- (Task 2.8 审查修复)一个 major + 四个 minor:① **`account_busy` 守卫移到共享层**:`Scheduler::ensure_switchable(target)`(+ 错误构造 `scheduler::account_busy`,从 `http/routes/accounts.rs` 挪过去)与 `AppState::switch_account(target)`(守卫 → `doubao.switch_account`),HTTP `POST /accounts/switch` 与 Tauri `account_switch` 都走它(spec 05 `AccountBusy`),`commands::account::tests::switch_is_refused_while_a_task_is_running` 镜像 HTTP 用例(FakeBackend + `spawn_loop`,取消后放行)。② `extract_bearer` 不再 `trim`:严格剥 `"Bearer "`,token 为空或含任何 ASCII 空白 → `auth_invalid_token`(`Bearer  abc` / `Bearer abc ` 之前被放行)。③ 已知路径 + 不支持的方法 → **405 + JSON** `request_malformed`(`method_not_allowed_fallback`;装在鉴权层之前,无 token 的错误方法仍 401,不泄露路由表;`/health` 单独装)。④ 下载:`ServeFile::try_call` 在文件于存在性检查之后被 GC 删掉时返回的是 404 **响应**而非 `Err`,现统一映射为 `result_expired{task_id, expired_at}` 且 `expired_at` 取任务真实值(原 `ErrorKind::NotFound` 分支报 `0`);`serve_result` 拆成独立函数直接测。⑤ `upload_too_large.bytes` 改为**文件**字节数(`file` 字段逐块计数到刚过线处),不再优先取整个 multipart body 的 `Content-Length`;只有 body 超过 `DefaultBodyLimit` 触发 413 且 `file` 字段尚未读到超限字节时才退回 `Content-Length`(没有则 `max_bytes + 1`),`too_large_bytes` 单测 + 52MB 用例覆盖。**未改(留 Task 4.2)**:`ensure_switchable` 检查与 `switch_account` 之间调度循环仍可能派发一个 pending 任务(check-then-act 窗口),跨账号串行化时一并用调度器锁覆盖。
- (Task 2.8 二次审查修复)① **GC 后下载报错修正**:GC 步骤 1 会把过期任务的 `result_files` 置 NULL,原 `locate_result` 先匹配 `(Completed, Some(files))` 再查过期,导致 24h 后的真实下载几乎都落到 `result_not_available`;现改为完成态先查 `expires_at`,且"完成但 `result_files` 为空 / NULL"一律 `result_expired{task_id, expired_at}`,`result_not_available` 只留给非完成态(用例 `download_after_gc_reports_result_expired` 先跑 `storage.gc` 再下载)。`GET /tasks/{id}` 对这种任务仍是 `status=completed` 且无 `result` 字段,未加额外字段(spec 04 未定义)。② `result_not_available` 的提示按状态区分:进行中 → "请等任务完成后再下载",失败 / 取消 → "任务未成功完成,没有可下载的结果…"(`routes::tasks::result_not_ready`)。③ **`POST /doubao/reconnect` 状态码 spec 冲突**:spec 04 写失败返回 500,11-error-codes 状态码表写 `doubao_*` → 503;**以 11-error-codes 为准**(全局单一映射 `http_status_for_code`,不按端点特判),`doubao_launch_failed` 等在该端点返回 503。

- (Task 2.9)任务命令 + 前端任务页实现差异(真机验收见 `docs/analysis/10-live-probe-macos.md` "Task 2.9"):① **asset 协议 scope 用 `$DATA/doubao-relay/results/**`**,不是计划写的 `$APPDATA/...`:Tauri 的 `$APPDATA` 是 `data_dir()/<identifier>`,而 relay 数据目录是 `platform.app_data_dir()` = `dirs::data_dir()/doubao-relay`(macOS `~/Library/Application Support/doubao-relay`;Windows `dirs::config_dir()` 与 `data_dir()` 同为 Roaming);`tauri` 加 `protocol-asset` feature。② `task_list` 每行比 HTTP 多 `promptPreview`(prompt 前 60 字 + `…`,`TaskSummary` 无 prompt,按 id 补读),`filter` 可省略 / null;`task_get` / `task_completed` = `to_camel(task_detail_json)` + `result.files[].localPath`(不安全相对路径给 null);`task_open_result_folder` 用 `platform.reveal_in_file_manager`(优先选中第一个结果文件),目录不存在 → `result_not_available`。③ `task_submit` 的 camelCase 请求 `to_snake` 后走 HTTP 同一解析 `generate_request_from_object`(从 `parse_generate_request` 拆出),错误码与 `POST /generate` 完全一致。④ `upload_add(filePath, kind?)` 规则同 `POST /uploads`(≤50MB、按内容判 MIME 且与 kind 相符、1h 过期),读不到文件 → `system_io_error{path}`;前端选文件用 `@tauri-apps/plugin-dialog` 的 `open`,capability 加 `dialog:default`。⑤ **比例白名单按类型拆分**(Task 2.1 / 2.8 ⑪ 遗留):`IMAGE_RATIOS`(无 21:9)/ `VIDEO_RATIOS`(无 3:2 / 2:3)/ `ratios_for(type)`,`request_invalid_enum.allowed` 按类型给;`RATIOS` 保留为两者之并。⑥ 前端:详情页不另设路由,`selectedTaskId` signal 非空时任务页显示详情(侧栏再点"任务"回列表);`view-model.ts` 除计划的 `statusIcon / relativeTime / timelineRows` 外另有 `upsertTask / applyTaskUpdate / summaryOf / canCancel / canConfirm / taskTypeText / shortId / formatBytes / modelsFor / ratiosFor / durationRange / fieldErrorOf`(测试在 `task-view-model.test.ts`,页面数据逻辑在 `pages/tasks.test.ts`);`task_updated` 只改列表行 status/stage,正在看的详情整条重读(时间轴 live),`task_completed` 直接替换详情并 toast;列表底部是"已显示 N 个任务"(API 不返回总数,spec 06 示例写"共 47 个")。新建弹窗的参考图只对图像开放(视频技能没有参考图入口,Task 2.1),但 Phase 3(Task 3.2)前带参考图提交会在 prepare 以 `reference_upload_unsupported` 失败。⑦ 状态栏"● API"仍固定"未启动":`api_listening` 已在 `AppState`,但 spec 05 没有读取它的命令,留 Task 5.1。⑧ 真机 UC-10 一次 49s 完成,**没有出现 checking / rendering 阶段**(主线 `sent → result → downloading → completed`),代回复与手动确认真机仍未验证(留 UC-22);UC-16 只验证了同账号第二个任务排队(避免多耗配额);UC-12 只有 1 个账号 SKIP。UI 截图用 Playwright + 真实任务数据(`scripts/ui-harness-tasks.playwright.js`),真 WKWebView 内 asset 缩略图未能截图证实(无屏幕录制,AX 也取不到窗口)。
- (Task 5.1)设置页 / 日志页 / 调试面板实现差异:① **`needsRestart` 真值表收窄**:只有 `server.host/port` 与 `doubao.cdp_port/cdp_address` 为 true;`server.bearer_token`、`logging.level`、`logging.cdp_debug_log` 热生效(计划原文"port/token 变更 → needsRestart"里的 token 改为热生效,因为计划同段又要求"重新生成 token 立刻生效",**Task 2.8 的 HTTP 鉴权必须每次请求读 `state.config` 而不是启动时拷贝一份**)。`logging.level` 经 `logging::set_level`(`tracing_subscriber::reload` 句柄,覆盖 `RUST_LOG`);`cdp_debug_log` 经 `CdpLogBuffer::set_enabled`。② **`CdpLogBuffer` 改为运行时可开关**:`sender()` 总是 `Some`(session 构造时拿一次),`cdp/session.rs` 只在 `receiver_count()>0` 时才构造 `CdpLogEntry`(关闭时零开销,不克隆 params/result);开启 = 启动收集任务(订阅 → 环形缓冲 + 追加 `<log_dir>/cdp-debug.log` ndjson + 转发到 UI 通道),关闭 = abort 收集任务。`subscribe()` 现在返回 UI 通道(`lib.rs` 的 `cdp_log_entry` 转发任务常驻订阅),`spawn_collector` 改名 `ensure_collector(log_dir)`(幂等),`spawn_background` 无条件调用它。文件路径按 spec 09 放 `log_dir`(spec 07 画在数据根目录,以 09 为准)。③ **`AppState` 增加懒打开的 `storage: tokio::sync::OnceCell<Storage>`** + `AppState::storage()`(首次调用 `Storage::open(app_data_dir)`,失败 `system_database_error`),`maintenance_gc` 据此调 `storage.gc(now, &config.storage)`;Task 2.7/2.8 接入调度器 / HTTP 时可把它改成启动即打开(字段名保持 `storage`),**每 1h 的定时 GC 本 Task 未加**(留给接调度器的 Task)。④ 新增命令:`app_regenerate_token` → `{bearerToken}`(Rust 侧 `Config::generate_token`,日志不含 token 值);`maintenance_gc` → `GcReport` camelCase;`log_export(path, level?)` → `{ok, path, count}`(把内存缓冲按最低级别写成 ndjson,字段保持 snake_case;路径来自 `tauri-plugin-dialog` save)。⑤ 插件:`tauri-plugin-clipboard-manager` / `tauri-plugin-dialog` 2.x 只在 `run()` 的生产 Builder 上注册(mock runtime 测试不碰系统剪贴板),capability 加 `clipboard-manager:allow-write-text` / `allow-read-text` / `dialog:allow-save`;前端 `@tauri-apps/plugin-clipboard-manager` / `plugin-dialog` 只在 `api.ts` 里 import。"复制 token" 60s 后先 `readText`,**仍等于刚复制的 token 才清空**(不覆盖用户之后复制的内容)。⑥ 前端:`view-model.ts` 追加 `maskToken` / `curlExample` / `validateSettings` / `saveResultMessage` / `shouldClearClipboard` / `gcSummary` / `levelRank` / `filterByMinLevel` / `appendCapped` / `formatLogTime` / `shortcutFor`;`state.ts` 追加 `appConfig` / `cdpDebugEnabled`;页面私有状态(草稿、日志条目、tab)放在各页模块内。日志页"清空"只清本页视图(文件与内存缓冲不动);级别过滤在前端做(`log_recent` 拉满 2000 条),导出把当前级别传给后端。CDP 调试面板做成日志页第二个 tab(`components/cdp-debug-panel.tsx`,最多留 500 条,未开启时给"去设置"入口)。每组"保存"按钮带 `aria-label="保存<组名>"`(五个同名"保存"按钮在 AX 树里无法区分,真机验收靠它)。⑦ spec 06 快捷键表:`Cmd/Ctrl+,` 设置、`Cmd/Ctrl+L` 日志、`Cmd/Ctrl+R` 刷新账号在 `components/shortcuts.tsx` 直接处理;`Cmd/Ctrl+N` 新建任务、`Cmd/Ctrl+K` 搜索任务只广播 `window` 事件 `relay:shortcut`(detail 为动作名),Task 2.9 的任务页监听即可。"关于"组的 查看开源协议 / 检查更新 / 帮助文档 三个按钮先禁用(LICENSE 与 opt-in 检查更新属 Task 6.x)。⑧ 日志文件按大小滚动(`file_max_size_mb`)仍未实现(Task 1.3 遗留,tracing-appender 不支持),设置页不暴露该项。真机结论见 `docs/analysis/10-live-probe-macos.md` "Task 5.1"。
- (Task 4.1 webhook 部分)`tasks/webhook.rs` 实现差异:① 签名按 12-security:`X-Doubao-Relay-Signature: sha256=<hex>`,data = `timestamp + "\n" + body`(04 写的 `hmac-sha256=` 前缀与无换行拼接不采用);每次尝试重新取时间戳重新签名。② `deliver(client, task, event, url, token, storage) -> DeliveryOutcome{delivered, attempts, last_http_status, last_error_code}`(不返回 Err,失败已落库);另有 `deliver_with(.., &WebhookPolicy{timeout, retry_delays})` 供测试缩短间隔,`WebhookPolicy::default()` = 30s + [1s,10s,60s];`event_for(TaskStatus)` 给出 `task.completed/failed/cancelled`(非终态 None);`webhook::new_client()` 不跟随重定向(3xx 记 `webhook_http_error`)、UA `doubao-relay/<版本>`,runner 应复用同一个 client。③ body 字段顺序固定 `task_id, event, status, task`,`task` **= `GET /tasks/{id}` 响应体**(spec 04):由 `tasks::detail::task_detail_json` 生成,含 `account/type/message/failed_at/result{files[].download_url, watermark_source}`(Task 4.1 审查修复;最初实现时 HTTP 层未落地,曾直接序列化 `Task`)。④ `webhook_log.error_message` 以 11 的码开头:`webhook_http_error:…` / `webhook_timeout:…` / `webhook_connection_failed:…`;`response_body` 最多 4096 字节。⑤ **"task.details 记 callback_failed"**:`Task` 没有 details 字段,改为 a) 追加 `stage_history` 一条 `stage="callback_failed"`(`GET /tasks/{id}` 可见),b) 任务有 `error` 时把 `{code:"webhook_max_retries_exceeded", event, url, attempts, last_http_status, last_error}` 合并进 `error.details.callback_failed`(原 details 保留;completed 任务只有 a)。`status` 不变。⑥ axum 只作 `[dev-dependencies]`(mock 回调服务器),Task 2.8 把 axum 加进 `[dependencies]` 后可删 dev 那行。⑦ capture.rs / runner.rs 的多轮确认部分由 4.1b 负责,runner 在终态 db 写完后 `tokio::spawn` 调 `deliver`(token 读当前 `config.server.bearer_token`)。

- (Task 3.2)`doubao/duration.rs` + `doubao/upload.rs` 实现差异(真机干跑见 `docs/analysis/10-live-probe-macos.md` "Task 3.2"):① `set_duration` / `upload_references` 签名照计划,另有 `set_duration_with(…, &DurationTiming)` / `upload_references_with(…, &UploadTiming)`(测试用小值),`SubmitTiming` 增 `duration` / `upload` 两个字段;`generation.rs` 的 `StepError` / `evaluate_step` / `map_preparing` 改为 `pub(crate)` 复用。② **方向键微调条件比沈浪宽**:PS 只在"至少拖过一次"后才用 `fineTuneKey`;这里离目标 ≤ 2 秒(`FINE_TUNE_MAX_GAP`)时未拖过也直接按方向键(计划 Step 1 的 mock 场景 End → ArrowRight → matched 依赖这一点),键盘检测(按后 `duration` 变化 ≠ ±1 → `keyboard_disabled`)与 8 次拖拽上限照 PS。纯函数 `DurationLoop::next_action` 单测覆盖分支顺序;非 Home/End/Arrow 的按键名不发送。③ 错误:循环结束仍不匹配 → `generation_preparing_failed{reason:"duration_not_applied", expected, actual}`,message "视频时长未生效:期望X秒,实际Y秒,…";`verifyVideoSettings` 抛异常 → `{reason:"verifyVideoSettings", js_exception}`,返回时长 ≠ 目标 → `{reason:"duration_mismatch"}`;`Input.*` / `DOM.*` 协议错误 → `{reason:"input_dispatch_failed"|"cdp_command_failed", method}`;既无 chooser 事件也找不到标记 input → `{reason:"upload_entry_changed", failed_selectors}`;`uploadReady` 抛异常 → `{reason:"uploadReady"}`。拖拽中途取消 / 出错也会发 `mouseReleased`;上传无论成败最后都 `setInterceptFileChooserDialog{enabled:false}`。④ `prepare_and_fill` 返回之后 `send` 用的请求 JSON(带 `uploadedCount`),`fill` / `checkSend` / `send` 都用它(原来 `send` 重新 `build_js_request(…, 0, …)`);阶段消息新增 `设置视频时长 N 秒` / `上传参考图 i/n`。Windows 原生"打开"对话框的 UIAutomation 兜底(沈浪 `Select-NativeUploadFile`)未实现。⑤ **真机新发现**:参考图属于技能草稿,失败残留会让下一次 `prepare` 12s 后 `新图片对话未准备好` 或立即 `豆包输入框已有参考图片…`,且只有在原技能里删除才生效 → 新增 `upload::remove_attachments`(`prepare` 前删可见卡片;`prepare` 抛 `已有参考图` 时删后重试一次;上传后任何失败删本次卡片),是计划外的 Rust 侧步骤。⑥ **视频技能同样接受参考图**(真机 Fast + 1 张通过 `checkSend`),Task 2.9 新建弹窗"参考图只对图像开放"可放开(未改前端)。⑦ **风险(未改)**:runner 的 preparing 60s 预算包住整个 `submit`,而 `uploadReady` 单张上限 70s、最多 10 张(真机每张约 0.9s,正常够用);弱网大图多张时可能以 `generation_preparing_timeout` 结束,需要的话由 Task 4.2 按参考图张数放宽。
- (Task 3.2 复审修复)真机见 `docs/analysis/10-live-probe-macos.md` "复审修复:参考视频与视频草稿残留":① **视频技能接受 mp4 参考视频**(卡片 `attachment-video-card`),上文 ⑥ 与 Task 2.1 ⑦"视频技能没有参考视频入口"作废;spec 允许 `video_reference`,前端新建弹窗可对视频任务开放参考视频(未改前端,留前端相关 Task)。② 视频草稿残留时 `prepare` 抛 `视频输入区已有内容，任务未发送`(不是 `已有参考图`):`LEFTOVER_ATTACHMENT_HINT` 改为 `LEFTOVER_HINTS = ["已有参考图", "输入区已有内容"]` + `is_leftover_exception`,命中后在技能里 `remove_attachments` + 清空输入框再重试 `prepare` 一次。③ `remove_attachments` 覆盖图片与视频卡片(`ATTACHMENT_CARD_SELECTOR` 改为数组 `ATTACHMENT_CARD_SELECTORS`)。④ 失败后清理改为短超时:`upload::cleanup_attachments`(2s)、关文件选择框拦截 `min(step_timeout, 2s)`、`clear_input` 2s(`upload::CLEANUP_TIMEOUT`),放进 runner 取消后的 5s 宽限期;`prepare` 前 / 重试前的删除用 `REMOVE_TIMEOUT` 5s(原为 `session.evaluate` 30s)。⑤ `durationDragPlan` 返回值改为逐字段宽松解析(`DragPlan::from_value`),单个字段类型不对不再丢掉 `matched:true`。⑥ `live_dry_run` 结束时在技能里删掉本次上传的卡片,`DOUBAO_LIVE_KEEP_REFS=1` 保留以复现残留。

- (Task 3.3)视频捕获实现差异(真机 UC-20 见 `docs/analysis/10-live-probe-macos.md` "Task 3.3"):① 路径 1 白名单复用 Task 3.1 的 `is_video_cdn_host`(`capture::video_candidate_url`,另要求无 userinfo),带 `lr=unwatermarked` 的候选直接成原片;reader `read` 的视频真实形状是 `images[{url, unwatermarked:true, vid,…}]`(`get_without_watermark` 换到的原片,记 `lr_unwatermarked`)或 `fallbacks[{vid, api}]`,**没有** Task 2.4 猜的 `videos` 字段(已删);两者都与图像一样**只在已绑定 submission 后采信**。② 路径 3 的 HTTP 抽象为 `capture::FallbackFetcher`(生产 `HttpFallbackFetcher`:不跟随重定向、10s、≤8MB、`Accept: application/json,…`),入口多一个 `capture_with_fetcher(…, timing, fetcher)`(`capture_with` = 默认 fetcher);解密出的地址经 `fallback_original_url`:已带 `lr=unwatermarked` 原样用,视频 CDN 地址无 `lr` 才按计划追加(真机解出的地址本身已带 `lr`);JS 兜底注入 `VIDEO_DECODER_SCRIPT` 求 `shenlangDoubaoVideoDecoder.decode(JSON.parse(body), api)`。`CaptureTiming` 增 `download_click_interval`(5s)/ `fallback_retry`(10s)/ `fallback_timeout`(10s)。`CaptureOutcome.watermark_source` 取第一条原片的来源。③ `download_video` 自己先校验 `lr=unwatermarked`(`download::has_lr_unwatermarked`,否则 `generation_watermark_not_removed{url}` 且不发请求);`mp4_duration_ms(bytes)` + `mp4_file_duration_ms(path)`(只读顶层 box 头、跳过 mdat 读 moov,≤64MB),runner 下载后填 `ResultFile.duration_ms`。④ runner:`GenerationBackend` 增 `download_video(url, dest_dir)`;视频只下第一条原片;**计划 Step 2 的"prompt 前缀"不做**(Task 2.1 结论:视频走 `skill` 分支,前缀只属于已弃用的 Mini `conversation` 路径);20min 预算沿用 Task 2.7 的 `SchedulerTiming.video_timeout`。⑤ **真机发现并修复的 3 个缺陷**:历史反馈按 messageId 去重导致流式半句分类 `unknown` 后永不重分类(改为 messageId+文本,`classify_key`);视频 auto 代回复 `videoFollowup` 需要 `modelName/ratio/duration/referenceFiles`(`CaptureOptions` 增 `ratio / duration / reference_count`,runner 填);reader 自家重放因 slardar 包装 `fetch` 认不出(postData 与模板完全相同即跳过)。⑥ **未修的风险**:沈浪脚本的 DOM 反馈锚点在视频技能下定位不到(基线把技能编辑器算进去),视频阶段识别只靠历史分类;修复后的 auto 确认链路真机未完整跑过(UC-20 是手动调用脚本 `videoFollowup` 补发的),留 Task 4.1 / UC-22 验证。⑦ UC-20 计划写 Seedance 2.0 Mini 12s,按本次指令改用最便宜的 Seedance 2.0 Fast 5s;只出现了路径 3(窗口隐藏下页面不加载视频媒体,路径 1/2 只有 mock);UC-27 未做。

- (Task 4.1b 确认部分 + webhook 接入)实现差异(无真机动作,见 `docs/analysis/10-live-probe-macos.md` "Task 4.1b"):① 计划列的 capture 行为(auto 失败降级 manual、会员自动失败关 `auto_use_membership`、manual 下确认后以 `userConfirmed=true` 调 `videoFollowup`)Task 2.4/3.3 已实现,本 Task 只补 mock 用例(`membership_auto_failure_degrades_then_user_confirm_replies` 等)。② `Stage(Sent, …)` 不只用于 `videoFollowup`:**图像代回复与视频 followup 真正发出(`sent:true`)后都发**,消息按来源区分 `capture::AUTO_CONTINUE_MESSAGE`("已自动点继续")/ `USER_CONTINUE_MESSAGE`("已按用户确认继续"),同一确认 stage 只发一次,`last_emitted_stage` 不改(避免 sent 与 checking 来回抖动)。③ webhook 投递挂在 `Scheduler::notify_terminal(&Task)`(`tokio::spawn`,token 取投递时的 `config.server.bearer_token`),**runner 终态与调度器 `finish_queued`(pending 取消 / 排队超时 / `fail_all_running` 的排队任务 / panic)都调用**——spec 04 要求任何 completed/failed/cancelled 都回调;启动时 `fail_non_terminal("relay_restart")` 批量改的任务**也回调**(Task 4.1 审查修复:`fail_non_terminal` 改为 `UPDATE … RETURNING id` 返回被改任务 id,`recover_after_restart` 逐条重读后调 `notify_terminal` 发 `task.failed`)。④ `SchedulerTiming` 增 `webhook: WebhookPolicy`(默认 30s + 1s/10s/60s),新增 `Scheduler::new_with_webhook_client(…, client)` 供测试替换 client(`new` / `new_with` 签名不变,内部用 `webhook::new_client()`);`callback_url` 校验拒绝 loopback,测试用 `hook.test` + `ClientBuilder::resolve` 指向本地 mock。⑤ 回调进行中(重试最长约 71s + 4×30s)relay 退出会丢弃未完成的重试,未做持久化重投。

- (Task 4.2)并发硬化 / 干扰检测 / 崩溃处理实现差异(真机见 `docs/analysis/10-live-probe-macos.md` "Task 4.2"):① **跨账号任务串行(与 spec 08 `global_max=3`、UC-24 冲突,按 Task 2.7 ⑦ 落实)**:`dispatch` 新增条件"运行中任务都属于队头任务的账号",叠加 `per_account_max=1`,单豆包实例上实际并发恒为 1,`global_max` 仍作上限;用例 `different_accounts_run_concurrently_up_to_global_max` 改为 `different_accounts_are_serialized_on_a_single_doubao`(切换序列恰为 b→c→d、`max_running=1`),另加 `global_max_still_caps_dispatch`。计划 Step 1 的"3 不同账号并发 3,第 4 个等待"因此不成立。② **switch_lock**:`Scheduler::begin_switch(target) -> SwitchGuard`(`running` 锁内检查无非终态任务 + `switching` 计数置位,守卫 drop 时减计数并唤醒派发),`AppState::switch_account` 改用它(审查修复后删除已无调用方的 `ensure_switchable`);派发循环在 `switching>0` 时不启动新 runner。`account_busy` 的判定仍是"任何非终态任务"(Task 2.8 ④)。③ **503**:`GenerationBackend` 增默认方法 `ready()`(测试替身不用实现),`DoubaoBackend` → `DoubaoConnection::ensure_ready()`;判定为 `cdp_connected || (running && reconnecting && 断线未满 crash_after)`——比计划"`cdp_connected=false` 即 503"宽:心跳在重连(切账号 reload、豆包仍在)时放行入队,避免 UC-31"切换后立即提交"误报 503;宽限以 `crash_after`(20s)封顶(审查修复:豆包被不带调试端口重开时 `reconnecting` 会一直为 true);错误码按计划 `doubao_not_ready`(details `{running, cdp_connected, reconnecting, outage_ms}`)。从未连接过也是 503:`auto_connect_on_startup=false` 时须先手动连接,`/generate` 不再隐式连接,`submit` 中位于请求校验之后、账号校验之前。④ **崩溃**:`ConnEvent::Crashed` 的订阅抽成 `lib::spawn_crash_watcher(scheduler, rx, last_seen_at)`,调 `Scheduler::fail_all_running_with(code, message, details)`(`fail_all_running` 保留为无 details 版),task.error = `doubao_crashed{last_seen_at}`(11-error-codes 的 details);`DoubaoConnection` 增 `last_seen_at()`(心跳 / 建会话成功时刻)与 `is_doubao_running()`。**计划外**:`DoubaoBackend` 的 submit / capture 返回断线类错误(`doubao_cdp_disconnected|not_ready|not_running`)且豆包进程已不在 → 改报 `doubao_crashed{last_seen_at, cause}`(`runner::crash_aware_error`),因为 capture 约 6s 就因 chat page 消失结束,早于 20s 的 `Crashed`。`error.rs` 增 `RelayError::doubao_crashed(last_seen_at)`,文案统一为 `error::DOUBAO_CRASHED_MESSAGE`(`lib::DOUBAO_CRASHED_MESSAGE` 重导出;崩溃监听与 `crash_aware_error` 两条路径同文案)。⑤ **超时别名**:`error::canonical_code` 把 spec 08 短名 `queue_timeout / preparing_timeout / task_timeout / confirm_timeout` 映射到 `generation_*`,`http_status_for_code` 先过它;调度器 / runner 写库仍只写正式码(Task 2.7 已实现 pending>1h 扫描)。⑥ **干扰检测**:chat page ≠1 的检查 Task 2.4 已在 capture 里每 2s 做(比计划 10s 更密),本 Task 不改;输入框检查(10s)改为"空 / 等于 prompt / 等于代回复'继续生成'"视为正常、代回复进行中跳过,其它内容 warn(只记长度,同一内容只 warn 一次,后续复查降为 debug)+ 新 `CaptureEvent::InputModified{text_length}`(同一内容只发一次),runner 把 `stage_message` 置为 `runner::INPUT_MODIFIED_MESSAGE`("检测到输入框被修改")并发 `task_updated`,不追加 stage_history、不终止。spec 08 "preparing 前探测 chat 页数量"(审查修复后):`DoubaoBackend::submit` 改用 `DoubaoConnection::submit_session()`——先 `/json/list` + `select_pages` 数 chat page,>1 → `generation_user_interrupted{interruption:"multiple_chat_pages", chat_targets, during:"preparing"}` 且**不填写、不发送**(不耗配额),同时把 `status.chat_targets` 写成实际数量;=1 才取 `chat_session()`;0 / `/json/list` 失败交给 `chat_session` 的重连路径;进入 generating 后仍由 capture 每 2s 检查。⑦ 真机:只读探测 + 一次 `Target.createTarget`/`closeTarget` 空白 chat 窗口(确认第二个 doubao-chat page 的 `/json/list` 形状,UC-28 前置)与 `live_ensure_ready_and_process_probe`;UC-24 不可达(见 ①)、UC-25 / UC-31 / UC-32 只有 mock(真实生成耗配额、只有 1 个账号、不允许杀豆包)。⑧ Task 3.2 ⑦"preparing 60s 按参考图张数放宽"未做(spec 08 固定 60s)。
- (Task 6.1)断线重连硬化 / 稳定性实现差异(真机与压测数字见 `docs/analysis/10-live-probe-macos.md` "Phase 6 稳定性"):① `ConnTiming` 增 `reconnect_delay`(2s)与 `restart_after`(20s);心跳失败后本次断线的第一次重连前等 2s(Task 1.8 的"立刻重连"改为 spec 09 的"等 2 秒"),之后 `spawn_heartbeat` 在 `retry_soon()`(断线中、进程在、未升级过)时每 2s 重试。② **升级重启的范围比计划窄**:只在"豆包进程还在且 20s 仍连不上"时 `connect(true)`,**每次断线只试一次**;进程已退出不自动拉起(与 UC-06"等待重启"一致,`Crashed` 照旧),重连失败原因为 `multiple_chat_pages` 也不升级(用户干扰,杀豆包会毁掉用户窗口)。重启失败(如 `wait_for_cdp` 20s 超时)期间心跳循环被占用约 20s。③ 捕获中途断线改用新 `DoubaoConnection::await_chat_session(cancel)`(每 2s 重试、最长 `restart_after`、进程不在立即放弃);计划说"capture 每次取最新 session",原实现只试一次,reload 时 chat page 短暂缺失就会让任务以 `doubao_cdp_disconnected` 失败。④ 内存:`LogBuffer` 2000 / CDP broadcast 4096 + Lagged / SSE backlog 64 条 1MB 已由前序 Task 实现,本 Task 补测试(`slow_subscriber_lags_without_blocking_commands`)并修 `CdpSession` pending 泄漏(外层丢弃 future 时不删条目 → `PendingGuard`)。⑤ `tests/soak.sh` 用新增的 `src-tauri/examples/soak_server.rs`(假豆包后端,`SOAK_CDP_PORT` 时额外连真豆包只跑心跳);**真豆包 10 个任务未跑**(禁止耗配额),24h 挂机未跑(只做 100 任务 + 10 分钟空闲);20s 升级重启与真杀豆包后的恢复只有 mock(禁止 pkill 豆包)。`tests/soak.sh` 在 git 里是 755(`update-index --chmod=+x`),本机工作区未能 `chmod`(执行环境拒绝),用 `bash tests/soak.sh` 运行即可。
- (Task 6.1 审查修复)① **补上 Task 1.8 ⑨(spec 09"安全校验(每次重连都要重跑)")**:新增 `bootstrap::verify_cdp_owner(platform, port)`(`listening_process_exe` + `is_doubao_exe_path`;监听者可确定且不是豆包 → `doubao_cdp_port_occupied{owning_process, port}`,查不到监听者 → Ok,由调用方已确认的 doubao-chat target 兜底)。`connect_sessions_inner` 在**每次新建 chat session 前**调用(覆盖心跳重连、`chat_session()`、`await_chat_session`、`connect(false)`;复用现有 session 时不重跑,lsof ~45ms);`connect(true)` 在 **kill 之前**先查一次(端口被别的程序占着时重启豆包也拿不到它,不白杀豆包),`wait_for_cdp` 之后再查一次。`chat_session()` 照旧把它包成 `doubao_cdp_disconnected{cause:"doubao_cdp_port_occupied"}`。② 心跳重连失败原因为 `doubao_cdp_port_occupied` 时与 `multiple_chat_pages` 一样**不升级重启**;两者都记入 `Outage.hold`,`retry_soon()` 返回 false → 回到常规 10s 心跳(不再每 2s 一次 sysinfo 扫描 + 2 条 WARN);下一次重连失败按新原因重判。测试 `reconnect_rejects_non_doubao_port_owner`、`multiple_chat_pages_falls_back_to_regular_heartbeat`。③ `CdpSession` 的 pending 插入 / 读循环删除改 `unwrap_or_else(|p| p.into_inner())`;`soak_server` 的 `png_bytes` 返回 `anyhow::Result`;`tests/soak.sh` 的 `rss_kb` 加 `|| true`(relay 退出时不中止脚本)、GET 样本数改 `wc -l`。`scripts/uc-10.sh` 工作区 mode 已 `git checkout` 恢复;`tests/soak.sh` 工作区仍是 644(环境拒绝 `chmod`),提交时用 `git add --chmod=+x` 保持 index 为 755。

- (Task 5.2)隐私声明 / 托盘 / 关闭流程实现差异(真机验收见 `docs/analysis/10-live-probe-macos.md` "Task 5.2"):① **同意前不可用的落法**:同意状态存 `metadata.privacy_accepted`(值 = 同意时刻毫秒),`AppState` 新增 `privacy: watch::Sender<bool>`;`spawn_background` 在启动恢复 / 调度循环 / GC 之后 `wait_for_privacy`(未同意 → 保证主窗口可见并等待),**同意前不启动 HTTP API、不自动连豆包**;已同意且 `ui.start_minimized` 时启动即把主窗口收进托盘(spec 07 字段,之前无人实现)。新 IPC 命令 `app_privacy_status` / `app_privacy_accept` → `{accepted, acceptedAt}`、`app_quit` → `{ok}`(spec 05 未列),放在 `commands/privacy.rs`。前端 `components/privacy-notice.tsx` 全屏声明(spec 12 原文;必须勾选"我已阅读并理解上述说明与风险"才能"同意并继续";"不同意并退出"调 `app_quit`),`state.privacyAccepted`(null 时先盖遮罩),状态读不到按未同意处理。② **关闭流程**在新模块 `shutdown.rs`:`graceful_shutdown(state, &ShutdownTiming)`(幂等,`QuitGate` + `OnceCell` 只跑一次)= `state.shutdown.cancel()`(HTTP 优雅关闭)→ 取消全部 pending 与 pending/preparing 的运行中任务(`scheduler.cancel`,`cancelled`)→ 等 generating 最多 10s(**只剩 `waiting_user_confirm` 时不等**,没人确认不会自行结束)→ 仍在运行的 `fail_all_running("system_shutdown", …)` → 等 runner 落库(≤5s)→ 库里残留非终态 `fail_non_terminal("system_shutdown")`;`close_resources` 关 SQLite 池 + `logging::flush()`(`GUARD` 从 `OnceLock<WorkerGuard>` 改为 `Mutex<Option<_>>`,drop 即写完缓冲;原先进程退出时 static 不 drop,最后几行日志可能丢)。**强制失败的错误码用 `system_shutdown`**(11-error-codes 的"服务正在关闭"),不是 `relay_restart`;关闭开始后的新提交也报 `system_shutdown`(新增 `AppState::submit_task`,HTTP `POST /generate` 与 Tauri `task_submit` 共用,HTTP 状态按 `system_*` 前缀为 500)。③ **CDP 连接没有显式关闭**:`DoubaoConnection` 没有 close 方法且本 Task 不改 `connection.rs`,WebSocket 随进程退出断开(真机确认豆包不受影响);若 Task 6.1 需要"关闭时先停心跳再断开",在 connection.rs 加 `shutdown()` 后接到 `close_resources`。**(合并到 main 时已补齐)**`DoubaoConnection` 新增 `closing: CancellationToken` + `shutdown()`(停 `spawn_heartbeat` 循环、`heartbeat()` 直接返回 false、`connect_sessions_inner` 报 `doubao_cdp_disconnected{cause:"relay_shutting_down"}` 不再重连 / 不升级重启,关闭 chat/background 会话,状态 `cdp_connected=false`;不杀豆包、幂等),`close_resources` 先调它再关 SQLite;测试 `shutdown_stops_heartbeat_closes_sessions_and_refuses_reconnect`、`close_resources_disconnects_cdp_and_closes_db`。另:main 上 `fail_non_terminal` 已改为返回 id 列表(Task 4.1 启动恢复回调),关闭流程第 4 步同样对兜底标失败的任务调 `scheduler.notify_terminal`(webhook 尽力而为,进程随后退出可能来不及送达)。④ **退出入口全部汇到 `shutdown::request_quit`**(隐藏窗口 → 关闭流程 → `app.exit(0)`):托盘"退出"、macOS 应用菜单"退出 doubao-relay"(Cmd+Q)、声明页"不同意并退出"、`RunEvent::ExitRequested`(最后一个窗口关闭 / 其它来源,先 `prevent_exit` 再走流程,`QuitGate.is_done()` 后放行)。**macOS 默认菜单的 Quit 不可拦截**(`NSApp terminate:` → `applicationWillTerminate` → 直接 `RunEvent::Exit`),所以 macOS 换成自定义应用菜单(`tray::app_menu`,含编辑 / 窗口子菜单);`RunEvent::Exit` 兜底用短时限(宽限 2s)同步跑一遍。⑤ 托盘(`tray.rs`,`tauri` 加 `tray-icon` feature,图标用 `default_window_icon`):菜单"显示主窗口 / 退出",Windows 双击托盘显示主窗口,提示文字 `doubao-relay · 空闲` / `运行中 N 个任务`(随任务事件刷新,F7.5 的 MAY);关闭按钮由 `on_window_event` 按当前 `ui.minimize_to_tray_on_close` 决定隐藏还是退出(主线程回调里 `try_read` 配置,拿不到锁按默认 true);macOS `RunEvent::Reopen`(点 Dock)重新显示窗口。⑥ **Windows**:本机只做了 `cargo check` + `cargo clippy --all-targets -D warnings`(`--target x86_64-pc-windows-gnu`,通过;msvc target 仍卡在 `ring` 的 `assert.h`,同 Task 1.5);UC-01/02/04/10/13/40/41 在 Windows 真机 / CI 上**未跑**(本机无 Windows,Task 6.2 的 CI workflow 也还没入库),托盘在 Windows 的表现未验证。⑦ 真机 UI 自动化:直接跑 debug 二进制时 AX 取不到窗口,`open -n` debug bundle 才行;托盘下拉菜单不进 AX 树,两项菜单只能人工点。

## 自检记录

- (Task 6.2)CI / 发布(本机验证见 `docs/analysis/10-live-probe-macos.md` "Task 6.2"):① **`.github/workflows/ci.yml` 与 `release.yml` 未能写入仓库**:执行环境的权限设置拒绝写 `.github/` 目录,两份 workflow 已写好、YAML 解析通过,暂存在执行会话的 scratchpad,需人工拷入 `.github/workflows/` 后单独提交;其余文件(脚本、CHANGELOG、LICENSE、README)已入库。② CI matrix 按计划只有 `macos-latest` + `windows-latest`(spec 13 还列了 ubuntu-latest;本项目不发 Linux,Linux 上 Tauri 还需 webkit2gtk 系统库);工具链钉 `1.94.0`(与开发机一致,避免新 stable 的 clippy lint 打红 CI);spec 13 的 `pnpm lint` 前端没有 lint 脚本,未加。CI 另有 `scripts` job 跑 `scripts/tests/run.sh`(bash + perl,`DRY_RUN` 断言签名命令形状)与 Windows 上 `sign-windows.ps1` 的语法解析 + `-DryRun` 冒烟(本机无 pwsh,ps1 只能在 CI 验证)。**Windows 上的 `cargo test` 从未跑过**(只做过交叉 `cargo check/clippy`),首次 CI 可能暴露平台相关测试问题。③ `bump-version.sh` 同步 6 处而非 spec 的 3 处:另加根 `package.json`、`Cargo.lock` 的 `doubao-relay` 条目(CI 用 `--locked`,不同步必挂)与 README 版本徽章(spec 13 checklist"README 版本号 badge");`--check [vX.Y.Z]` 供 CI 与 release 校验 tag。④ release.yml:`verify`(tag == 版本号且 CHANGELOG 有 `## [X.Y.Z]` 段)→ macOS universal / Windows x64 构建 → secrets 存在时签名(macOS 另公证)→ `softprops/action-gh-release` 发布,说明取 CHANGELOG 段落,tag 含 `-` 标 prerelease;构建步骤剥掉 `APPLE_*` 环境变量,避免 tauri 自己再签一遍(签名统一走 `scripts/sign-macos.sh`)。⑤ `sign-macos.sh` 签 .app 后**重建** DMG(暂存目录 = .app + `/Applications` 链接,丢掉 Tauri DMG 的窗口布局 / 背景)再签 DMG;不带 entitlements,真机 ad-hoc + hardened runtime 启动正常。真 Developer ID 签名 / 公证 / Windows signtool 均无证书未验证。⑥ 发布 checklist 写在 README"发布"节(计划未列独立文件);spec 13 提到的 `uninstall-user-data.sh/ps1` 未做(README 卸载节给了手动命令)。⑦ 本机 `cargo tauri build --debug` 产物名是 `doubao-relay_0.1.0_aarch64.dmg`(按架构命名,spec 13 的 `_x64` 只是示例)。

- Spec 覆盖:F1(1.5/1.6/1.8)F2(1.7/1.8)F3(2.2/2.7/2.8)F4(2.7/4.1/4.2)F5(2.5/2.8)F6(2.8/3.2)F7(1.10/2.9/5.1/5.2)F8(1.3/5.1);非功能 N1–N9(6.1/6.2/2.8/5.2);安全 12(1.3/2.8/4.1/5.1);数据 07(2.6);构建 13(6.2);验收 14 各 UC 已标注到 Task。
- 类型一致性:`RelayError`/`CdpSession`/`DoubaoConnection`/`Scheduler`/`Storage` 的签名在后续 Task 中引用一致;`Task`/`GenerateRequest` 在 2.2 定义后不再改字段名。

## 交付状态(2026-10-17)

代码是事实来源;上面各 Task 的 `- [ ]` 复选框多数没有回填,以本节为准。全部 Task 的提交都在 `main`,验证命令(fmt / clippy `-D warnings` / `cargo test --workspace` 544 passed 17 ignored / 前端 typecheck + vitest 154 passed)在最终提交上通过。

### 最终审查修复(本轮)

- **账号菜单读取与任务页面操作交错**(major,已修,`56eaddc`):`DoubaoBackend::submit` 全程持 `DoubaoConnection::lock_page()`(= `switch_lock` + 页面忙标记),capture 全程持 `mark_page_busy()`。页面忙时非强制的 `accounts(false)`(`POST /generate` 账号校验、`GET /accounts`、UI `account_list`、runner `ensure_account`)只要有过缓存就沿用(过期也用),不再打开账号菜单;从未读到过列表时等提交释放页面锁再读。`account_refresh`(`accounts(true)`)仍会读菜单:提交阶段要等页面锁,捕获阶段不拦(用户主动刷新,代价可接受)。测试:`connection::tests::accounts_reuse_stale_cache_while_page_busy`、`accounts_wait_for_page_lock_when_no_cache`。
- **`fail_all_running_with_details_records_them` 竞态**(major,已修,`1075599`):runner 先写终态,`spawn_runner` 外层再移出 `running`;新增测试辅助 `wait_running_empty` 轮询,同类单次断言(`image_task_success_path…`、`running_task_cancel_becomes_cancelled`)一并改用。连续 3 次全量 `cargo test --workspace` 通过。

### 各 Task 完成情况

| Task | 状态 | 备注 |
|---|---|---|
| 1.1 – 1.10 | 完成 | Phase 1 真机验收见 10-live-probe "Phase 1 验收" |
| 2.1 – 2.9 | 完成 | 2.1 探测结论决定 image=`conversation`、video=`skill` |
| 3.1 | 完成 | `fixture_video_token_1_decodes` 用真机样本通过 |
| 3.2 | 完成 | 时长滑块 / 参考图上传真机探测;残留技能草稿清理 |
| 3.3 | 完成(真机部分) | UC-20 真机生成 1 段视频;UC-21 路径 1/2、UC-27 仅 mock |
| 4.1 | 完成 | auto/manual/降级/超时 + webhook;全部 mock |
| 4.2 | 完成 | 跨账号串行(单实例实际并发 1,见已知问题 Task 4.2 ①);崩溃处理 mock |
| 5.1 | 完成 | 设置 / 日志 / CDP 面板真机通过 |
| 5.2 | 完成 | 隐私声明 / 托盘 / 关闭流程真机通过 |
| 6.1 | 完成(部分验收) | 断 socket 重连真机通过;20s 升级重启、24h 挂机未做 |
| 6.2 | 完成 | `.github/workflows/{ci,release}.yml` 已入库;真 Developer ID 签名 / 公证 / Windows 签名未验证 |

### 已在 macOS 真机通过的验收用例

- UC-01(复用已开 CDP;豆包未开时自动启动)、UC-02、UC-06(Task 1.10 唯一一次 pkill)。
- UC-10(Seedream 5.0 Pro 单图 49s 完成)、UC-11、UC-13、UC-14、UC-15(pending 取消)、UC-17、UC-18(含 Range)。
- UC-20(Seedance 2.0 Fast 5s,路径 3 解密原片,无水印;确认补发是手动调脚本函数发出的)。
- UC-51 / NF-01(mock 后端 + 真实 Storage/Scheduler/HTTP 的 100 任务压测,`tests/soak.sh`)、UC-50 的 10 分钟空闲样本(RSS / CPU / ERROR 达标,非 24h)。
- Task 5.1 / 5.2 / 6.2 的 UI、关闭流程、打包启动冒烟(无 UC 编号的 spec 06 / 12 / 13 条目)。

### 仍需用户在真机 / Windows 上验证

| 用例 | 原因 |
|---|---|
| UC-03、UC-04、UC-12 | 本机只有 1 个有效账号,多账号读取 / 真实切换 / 自动切换只有 mock |
| UC-05 | 未登录态只有 mock(不能登出唯一账号) |
| UC-16、UC-24、UC-25 | 5 任务串行 / 跨账号依次执行需要多次真实生成与多账号;UC-24 的"并发 3"在单豆包实例上按设计不可达 |
| UC-19、UC-53、UC-54 | 过期 / 清理依赖时间推进,只有单测 |
| UC-21(路径 1/2) | 隐藏窗口下本版本只出现路径 3;路径 1/2 仅 mock |
| UC-22、UC-23 | 修复后的 auto 代回复链路(历史分类 → `videoFollowup` 带全参数补发)与 manual `task_confirm` 未在真机完整跑过,需要真实生成 |
| UC-26 | 回调只用 `hook.test` 本地 mock 验证;真回调需要非 loopback 地址 |
| UC-27 | 参考图生成需要额外一次真实生成 |
| UC-28 | 只验证了第二个 chat page 的 `/json/list` 形状;用户在豆包里"新开窗口"是否等价、输入框被改提示需人工操作 |
| UC-30 | 需要真实耗尽配额 |
| UC-31、UC-32、UC-33 | 页面 reload / 真杀豆包 / relay 重启时任务中途状态只有 mock(约束禁止 pkill 豆包与额外生成);20s 升级重启路径同理 |
| UC-34 | UI 自检失败需要改版的豆包 |
| 本轮修复的页面锁 | 需在真机上:视频任务 preparing 期间连续 `POST /generate` / 打开账号页,确认任务不再因菜单被打断(需要真实生成) |
| UC-07、UC-40、UC-41 | 需要 Windows 机器(本机只做 `cargo check --target x86_64-pc-windows-msvc` 与 CI) |
| UC-42 | 需要 Developer ID 证书 + 公证的正式包在另一台 Mac 上首次打开 |
| UC-50(24h)、UC-52 | 24h 挂机与版本升级迁移未做 |
