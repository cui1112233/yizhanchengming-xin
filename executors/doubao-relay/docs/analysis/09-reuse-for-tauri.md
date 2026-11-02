# 09 · Tauri 中转程序可复用清单

## 目标回顾

**用 Tauri 做一个 Windows 桌面客户端**,内置:
1. 一个 WebView 小 UI:通过 CDP 连豆包,显示账号列表 + 当前任务状态
2. Rust 后端起 HTTP server 对外暴露"生图"/"生视频"API
3. API 接口协议后续补充

## 整体架构

```
┌──────────────────────────────────────────────────────────┐
│  Tauri Desktop App (Windows)                              │
│                                                           │
│  ┌─────────────────────┐    ┌──────────────────────────┐  │
│  │  WebView (HTML/JS)  │    │  Rust Backend            │  │
│  │                     │    │                          │  │
│  │  • 账号列表 UI       │◄──►│  • Tauri 命令处理器       │  │
│  │  • 任务进度 UI       │invoke │                         │  │
│  │  • 配置 UI          │    │  • CDP 客户端             │  │
│  │                     │    │  • doubao-* 操控逻辑      │  │
│  │                     │    │  • JS 代码片段 (静态嵌入) │  │
│  │                     │    │  • 任务调度器             │  │
│  │                     │    │  • SQLite 任务持久化      │  │
│  │                     │    │  • HTTP API Server       │  │
│  └─────────────────────┘    └──────────────────────────┘  │
│                                      ▲                    │
└──────────────────────────────────────┼────────────────────┘
                                       │
          ┌────────────────────────────┼────────────────┐
          │                            │                │
          ▼                            ▼                ▼
   豆包桌面端                   外部 API 调用方      本地任务存储
   (CDP :9707)                 (HTTP :8080)          (SQLite)
```

## 可以原封不动复用的东西(3 个 JS 文件)

**这 3 个文件都是注入豆包页面的,与 PowerShell 完全独立,直接复制到 Tauri 项目**:

```
tauri-app/
├── src-tauri/
│   ├── Cargo.toml
│   ├── build.rs
│   └── src/
│       └── doubao_scripts/
│           ├── doubao-generation-page.js   ← 直接复制
│           ├── doubao-result-reader.js     ← 直接复制
│           └── doubao-video-decoder.js     ← 直接复制 (可选,解密也能用 Rust 实现)
```

### 在 Rust 里嵌入

```rust
pub static GENERATION_PAGE_JS: &str =
    include_str!("doubao_scripts/doubao-generation-page.js");
pub static RESULT_READER_JS: &str =
    include_str!("doubao_scripts/doubao-result-reader.js");
pub static VIDEO_DECODER_JS: &str =
    include_str!("doubao_scripts/doubao-video-decoder.js");
```

### 调用模式

```rust
pub async fn inject_and_run(session: &CdpSession, step: &str, request: &Value) -> Result<Value> {
    let request_json = serde_json::to_string(request)?;
    let step_json = serde_json::to_string(step)?;
    let bootstrap = format!(
        "if (typeof shenlangDoubaoGeneration === 'undefined') {{ {} }};",
        GENERATION_PAGE_JS
    );
    let expression = format!(
        "{} typeof shenlangDoubaoGeneration === 'undefined' ? ({{stage:'unknown'}}) : \
         Promise.resolve(shenlangDoubaoGeneration.run({}, {}))",
        bootstrap, step_json, request_json
    );
    let result = session.evaluate(&expression, true).await?;
    Ok(result)
}
```

**关键点**:
- 每次调用都带 `typeof shenlangDoubaoGeneration === 'undefined'` 检查,防止重复初始化。
- 用 `Promise.resolve(...)` 包裹,统一异步/同步分支。
- `returnByValue=true` + `awaitPromise=true` 一起用。

### JS 文件的"反混淆"要不要做

**不要做**。原因:
- 已经被 javascript-obfuscator 处理过,运行时正确就行
- 反混淆后可读性提升不多(标识符仍然是 _0x1a2b 这类)
- 一旦改动,下次沈浪发新版你很难对 diff

**只在真的要修 bug 时局部还原**。

## 需要用 Rust 重写的东西

### 1. CDP 客户端

**选型**:自己基于 `tokio-tungstenite` 写一个薄的 CDP 封装。

**不要**用 `chromiumoxide` / `headless_chrome`,它们是面向启动 Chromium 的,不适合"连接到已存在的 Electron 应用"。

```rust
use tokio_tungstenite::{connect_async, tungstenite::Message};
use futures_util::{SinkExt, StreamExt};
use serde_json::{json, Value};

pub struct CdpSession {
    tx: mpsc::Sender<Value>,
    pending: Arc<Mutex<HashMap<u64, oneshot::Sender<Value>>>>,
    event_tx: broadcast::Sender<Value>,
    next_id: AtomicU64,
}

impl CdpSession {
    pub async fn connect(ws_url: &str) -> Result<Self> {
        let (ws_stream, _) = connect_async(ws_url).await?;
        let (mut write, mut read) = ws_stream.split();
        let (tx, mut rx) = mpsc::channel::<Value>(100);
        let (event_tx, _) = broadcast::channel::<Value>(1024);
        let pending: Arc<Mutex<HashMap<u64, _>>> = Default::default();

        // 发送线程
        tokio::spawn(async move {
            while let Some(msg) = rx.recv().await {
                let _ = write.send(Message::Text(msg.to_string())).await;
            }
        });

        // 接收线程
        let pending_clone = pending.clone();
        let event_tx_clone = event_tx.clone();
        tokio::spawn(async move {
            while let Some(Ok(Message::Text(txt))) = read.next().await {
                let msg: Value = match serde_json::from_str(&txt) { Ok(m) => m, _ => continue };
                if let Some(id) = msg.get("id").and_then(|v| v.as_u64()) {
                    // 这是对一个 command 的响应
                    if let Some(sender) = pending_clone.lock().await.remove(&id) {
                        let _ = sender.send(msg);
                    }
                } else {
                    // 这是一个事件
                    let _ = event_tx_clone.send(msg);
                }
            }
        });

        Ok(Self {
            tx, pending, event_tx,
            next_id: AtomicU64::new(1),
        })
    }

    pub async fn send(&self, method: &str, params: Value) -> Result<Value> {
        let id = self.next_id.fetch_add(1, Ordering::SeqCst);
        let (resp_tx, resp_rx) = oneshot::channel();
        self.pending.lock().await.insert(id, resp_tx);

        let msg = json!({ "id": id, "method": method, "params": params });
        self.tx.send(msg).await?;

        let response = resp_rx.await?;
        if let Some(err) = response.get("error") { return Err(/* ... */); }
        Ok(response["result"].clone())
    }

    pub async fn evaluate(&self, expression: &str, await_promise: bool) -> Result<Value> {
        let result = self.send("Runtime.evaluate", json!({
            "expression": expression,
            "returnByValue": true,
            "awaitPromise": await_promise
        })).await?;
        if let Some(exc) = result.get("exceptionDetails") { return Err(/* ... */); }
        Ok(result["result"]["value"].clone())
    }

    pub fn subscribe_events(&self) -> broadcast::Receiver<Value> {
        self.event_tx.subscribe()
    }
}
```

### 2. 大整数 JSON 预处理

**豆包的所有 ID 都是 >16 位整数**,`serde_json::Value` 的 `Number` 会丢精度。

```rust
use regex::Regex;
use once_cell::sync::Lazy;

static BIG_INT_FIELDS: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r#""(conversation_id|local_message_id|message_id|bot_reply_message_id)"\s*:\s*(-?\d{16,})(?=\s*[,}])"#).unwrap()
});

pub fn fix_big_integers(json_text: &str) -> String {
    BIG_INT_FIELDS.replace_all(json_text, r#""$1":"$2""#).to_string()
}
```

### 3. 启动豆包 (Windows 专属)

```rust
use std::process::Command;
use winreg::enums::*;
use winreg::RegKey;
use sysinfo::{System, SystemExt, ProcessExt};

pub async fn ensure_doubao_started() -> Result<StartupReport> {
    // 1. 已启动且 CDP 通 → 直接返回
    if is_doubao_running() && test_cdp_connection(9707).await.is_ok() {
        return Ok(StartupReport::already_connected());
    }

    // 2. 找 exe 路径
    let exe_path = find_doubao_exe()?;

    // 3. 杀掉现有进程
    if is_doubao_running() {
        kill_doubao_processes();
        tokio::time::sleep(Duration::from_millis(800)).await;
    }

    // 4. 用调试参数启动
    Command::new(&exe_path)
        .args(&[
            "--remote-debugging-port=9707",
            "--remote-debugging-address=127.0.0.1",
            "--force-renderer-accessibility=complete",
        ])
        .spawn()?;

    // 5. 等 CDP 就绪
    wait_for_cdp_ready(9707, Duration::from_secs(20)).await?;
    Ok(StartupReport::started())
}

fn find_doubao_exe() -> Result<PathBuf> {
    let mut candidates = Vec::new();

    // 来源 A: 运行中的进程
    let mut sys = System::new_all();
    sys.refresh_all();
    for proc in sys.processes_by_name("Doubao") {
        if let Some(path) = proc.exe().parent() {
            candidates.push(proc.exe().to_path_buf());
        }
    }

    // 来源 B: 注册表 App Paths
    for root in [HKEY_CURRENT_USER, HKEY_LOCAL_MACHINE] {
        let hkey = RegKey::predef(root);
        if let Ok(key) = hkey.open_subkey(r"Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe") {
            if let Ok(val) = key.get_value::<String, _>("") {
                candidates.push(PathBuf::from(val));
            }
        }
    }

    // 来源 C: 注册表 Uninstall
    for root in [HKEY_CURRENT_USER, HKEY_LOCAL_MACHINE] {
        let hkey = RegKey::predef(root);
        if let Ok(uninstall) = hkey.open_subkey(r"Software\Microsoft\Windows\CurrentVersion\Uninstall") {
            for sub_name in uninstall.enum_keys().flatten() {
                if let Ok(sub) = uninstall.open_subkey(&sub_name) {
                    let display_name: String = sub.get_value("DisplayName").unwrap_or_default();
                    if display_name.contains("豆包") || display_name.contains("Doubao") {
                        if let Ok(install_loc) = sub.get_value::<String, _>("InstallLocation") {
                            candidates.push(PathBuf::from(&install_loc).join("Doubao.exe"));
                            candidates.push(PathBuf::from(&install_loc).join("app").join("Doubao.exe"));
                        }
                    }
                }
            }
        }
    }

    // 来源 D: 常见默认目录
    if let Ok(local) = std::env::var("LOCALAPPDATA") {
        candidates.push(PathBuf::from(&local).join("Doubao").join("app").join("Doubao.exe"));
        candidates.push(PathBuf::from(&local).join("Programs").join("Doubao").join("Doubao.exe"));
    }

    // 筛选存在的文件,返回第一个
    candidates.into_iter()
        .filter(|p| p.file_name().and_then(|f| f.to_str()) == Some("Doubao.exe"))
        .find(|p| p.exists())
        .ok_or_else(|| anyhow!("未找到豆包安装程序"))
}
```

**crates**:
- `winreg` — Windows 注册表
- `sysinfo` — 进程枚举
- `tokio` — 异步

### 4. 账号模块(Rust 简化版)

**决策**:只做 DOM 通道,放弃 UIA 通道(UIAutomation 的 Rust 绑定太麻烦)。

```rust
pub async fn list_accounts(session: &CdpSession) -> Result<Vec<DoubaoAccount>> {
    // 1. 定位账号按钮
    let locate_result = session.evaluate(LOCATE_ACCOUNT_BUTTON_JS, false).await?;

    // 2. 如果找不到,展开侧边栏
    if locate_result["candidates"].as_array().map_or(0, |a| a.len()) == 0 {
        session.evaluate(EXPAND_SIDEBAR_JS, false).await?;
        tokio::time::sleep(Duration::from_millis(400)).await;
    }

    // 3. 点开账号菜单
    session.evaluate(OPEN_ACCOUNT_MENU_JS, false).await?;
    tokio::time::sleep(Duration::from_millis(650)).await;

    // 4. 悬停"更多账号"
    session.evaluate(HOVER_MORE_ACCOUNTS_JS, false).await?;
    tokio::time::sleep(Duration::from_millis(650)).await;

    // 5. 读账号列表,循环直到稳定
    let mut previous_count = usize::MAX;
    let mut stable_rounds = 0;
    let mut entries = vec![];
    for attempt in 0..8 {
        session.evaluate(HOVER_MORE_ACCOUNTS_JS, false).await?;  // 防菜单关
        tokio::time::sleep(Duration::from_millis(400)).await;
        let result = session.evaluate(READ_ACCOUNT_ENTRIES_JS, false).await?;
        entries = result["entries"].as_array().cloned().unwrap_or_default();
        if entries.len() == previous_count {
            stable_rounds += 1;
            if attempt >= 3 && stable_rounds >= 2 { break; }
        } else {
            stable_rounds = 0;
            previous_count = entries.len();
        }
    }

    // 6. 过滤 + 解析套餐
    let excluded = ["设置","外观设置","收藏夹","豆包官网","检查更新",
                    "帮助与反馈","更多账号","添加账号","退出登录",
                    "豆包订阅","豆包订阅升级","免费领额度","免费领取额度"];
    let subscription_re = Regex::new(r"(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|.*套餐)$").unwrap();

    let accounts: Vec<DoubaoAccount> = entries.iter().filter_map(|entry| {
        let label = entry["label"].as_str()?;
        let has_avatar = entry["avatar"].as_bool().unwrap_or(false);
        let has_popup = entry["hasPopup"].is_string();
        if !has_avatar || has_popup || excluded.contains(&label) { return None; }

        let cleaned = label.strip_prefix("avatar ").unwrap_or(label).trim();
        let (name, subscription) = if let Some(m) = subscription_re.find(cleaned) {
            (cleaned[..m.start()].trim(), Some(m.as_str().to_string()))
        } else {
            (cleaned, None)
        };
        if name.is_empty() || name.len() > 120 || excluded.contains(&name) { return None; }

        Some(DoubaoAccount {
            name: name.to_string(),
            label: label.to_string(),
            subscription,
        })
    }).collect();

    Ok(accounts)
}
```

5 段 JS 常量(`LOCATE_ACCOUNT_BUTTON_JS` 等)从 `03-account-module.md` 整理出来放在 Rust 源文件头部。

### 5. 生成管线

**整体复用 Step 模式**,但不要再 spawn PowerShell:

```rust
pub async fn submit_generation(
    session: &CdpSession,
    request: &GenerationRequest,
    event_tx: mpsc::Sender<GenerationEvent>
) -> Result<SubmissionReport> {
    // 校验
    validate_model(&request.model_name)?;
    validate_ratio(&request.ratio)?;
    if request.task_type == TaskType::Video {
        validate_duration(request.duration, &request.model_name)?;
    }

    // 焦点伪装
    session.send("Page.setWebLifecycleState", json!({"state": "active"})).await?;
    session.send("Emulation.setFocusEmulationEnabled", json!({"enabled": true})).await?;

    // 注入 JS 一次性注入
    session.evaluate(GENERATION_PAGE_JS, false).await?;

    // Step 流水线
    event_tx.send(GenerationEvent::stage("mode")).await.ok();
    let prepared = step(session, "prepare", request).await?;

    if prepared["needsDurationInput"].as_bool().unwrap_or(false) {
        handle_duration_drag(session, request).await?;
    }

    if !request.reference_files.is_empty() {
        event_tx.send(GenerationEvent::stage("upload")).await.ok();
        upload_files(session, request, &event_tx).await?;
    }

    event_tx.send(GenerationEvent::stage("prompt")).await.ok();
    step(session, "fill", request).await?;
    step(session, "checkSend", request).await?;

    // 并行追踪:订阅网络事件抓 submissionIdentity
    let submission_requests = Arc::new(Mutex::new(Vec::new()));
    let subscriber = subscribe_submission_requests(session, submission_requests.clone()).await?;

    // 发送
    event_tx.send(GenerationEvent::stage("sending")).await.ok();
    let send_started_at = chrono::Utc::now().timestamp_millis();
    step(session, "send", request).await?;
    event_tx.send(GenerationEvent::stage("sent")).await.ok();

    // 2 秒内识别 submissionIdentity
    let identity = find_submission_identity(&submission_requests, &request.prompt).await?;

    subscriber.abort();
    Ok(SubmissionReport {
        submitted: true,
        submission_identity: identity,
        send_started_at,
    })
}

async fn step(session: &CdpSession, name: &str, request: &GenerationRequest) -> Result<Value> {
    let payload = serde_json::to_string(request)?;
    let name_json = serde_json::to_string(name)?;
    let expression = format!(
        "shenlangDoubaoGeneration.run({}, {})",
        name_json, payload
    );
    session.evaluate(&expression, true).await
}
```

### 6. 捕获管线

**最有工程含量**。要并发处理:
- Network 事件流
- 定时 feedback 轮询
- 定时 history 重放
- 文件下载

```rust
pub async fn capture_result(
    session: &CdpSession,
    request: &GenerationRequest,
    identity: &SubmissionIdentity,
    event_tx: mpsc::Sender<GenerationEvent>,
    output_dir: &Path,
) -> Result<CaptureReport> {
    let mut rx = session.subscribe_events();
    session.send("Network.enable", json!({
        "maxTotalBufferSize": 16777216,
        "maxResourceBufferSize": 8388608
    })).await?;

    let mut state = CaptureState::new(request, identity);
    let deadline = Instant::now() + Duration::from_secs(15 * 60);
    let mut feedback_timer = tokio::time::interval(Duration::from_millis(500));
    let mut history_timer = tokio::time::interval(Duration::from_secs(1));

    loop {
        tokio::select! {
            _ = tokio::time::sleep_until(deadline.into()) => break,
            Ok(event) = rx.recv() => handle_network_event(&event, &mut state, session, &event_tx).await?,
            _ = feedback_timer.tick() => feedback_poll(session, &mut state, &event_tx).await?,
            _ = history_timer.tick() if state.send_since.is_some() && state.originals.is_empty() => {
                history_poll(session, &mut state, &event_tx).await?;
            }
        }

        if !state.originals.is_empty() || state.refused { break; }
    }

    if state.originals.is_empty() {
        return Err(anyhow!("未捕获到生成结果"));
    }

    download_originals(&state.originals, request.task_type, output_dir).await
}
```

### 7. 视频解密

见 `06-video-decryption.md`,用 `aes` + `cbc` + `sha2` 三个 crate 完整重写。**推荐做,不要依赖注入 JS 来解密**。原因:
- JS 版本依赖 WebCrypto 的 `crypto.subtle`,是异步的,时序复杂。
- Rust 直接解更快,错误处理更可控。
- 测试更容易。

### 8. HTTP API Server

```rust
use axum::{Router, routing::{get, post}, extract::State, Json};

#[derive(Clone)]
struct AppState {
    cdp_sessions: Arc<CdpSessionPool>,
    tasks: Arc<DashMap<Uuid, TaskHandle>>,
    db: SqlitePool,
}

async fn api_generate(
    State(state): State<AppState>,
    Json(req): Json<GenerationRequest>,
) -> Result<Json<TaskResponse>, ApiError> {
    let task_id = Uuid::new_v4();

    // 保存任务到数据库
    sqlx::query!("INSERT INTO tasks (id, status) VALUES (?, 'pending')", task_id).execute(&state.db).await?;

    // 异步执行
    let state_clone = state.clone();
    tokio::spawn(async move {
        let _ = run_task(state_clone, task_id, req).await;
    });

    Ok(Json(TaskResponse { task_id, status: "pending" }))
}

async fn api_task_status(
    State(state): State<AppState>,
    Path(task_id): Path<Uuid>,
) -> Result<Json<TaskStatus>, ApiError> {
    let row = sqlx::query!("SELECT status, result_json FROM tasks WHERE id = ?", task_id)
        .fetch_optional(&state.db).await?
        .ok_or(ApiError::NotFound)?;
    Ok(Json(TaskStatus {
        status: row.status,
        result: row.result_json.and_then(|s| serde_json::from_str(&s).ok()),
    }))
}
```

**端口建议**:8080 对外暴露。Tauri app 内嵌。

## 项目骨架

```
tauri-doubao-relay/
├── Cargo.toml
├── tauri.conf.json
├── build.rs
│
├── src-tauri/
│   ├── Cargo.toml
│   ├── build.rs
│   └── src/
│       ├── main.rs                     # Tauri 入口 + HTTP server 启动
│       ├── lib.rs
│       │
│       ├── cdp/
│       │   ├── mod.rs
│       │   ├── session.rs              # CdpSession 实现
│       │   ├── events.rs               # 事件类型定义
│       │   └── json_fix.rs             # 大整数预处理
│       │
│       ├── doubao/
│       │   ├── mod.rs
│       │   ├── bootstrap.rs            # ensure_doubao_started, find_doubao_exe
│       │   ├── accounts.rs             # list_accounts, switch_account
│       │   ├── generation.rs           # submit_generation
│       │   ├── capture.rs              # capture_result 主循环
│       │   ├── feedback.rs             # 多轮确认处理
│       │   ├── download.rs             # 结果下载
│       │   ├── video_decoder.rs        # AES-CBC 本地解密
│       │   └── scripts/
│       │       ├── doubao-generation-page.js
│       │       ├── doubao-result-reader.js
│       │       └── doubao-video-decoder.js
│       │
│       ├── commands.rs                 # Tauri invoke 命令
│       ├── http_api.rs                 # 对外 HTTP server
│       ├── task_queue.rs               # 任务调度
│       └── db.rs                       # SQLite 持久化
│
├── frontend/
│   ├── index.html
│   ├── main.ts
│   └── accounts.vue 或 react 组件
│
└── docs/                                ← 本文档目录
    └── (shenlang-analysis 的内容)
```

### Cargo.toml 依赖

```toml
[dependencies]
tauri = { version = "2", features = ["protocol-asset"] }
tokio = { version = "1", features = ["full"] }
tokio-tungstenite = { version = "0.21", features = ["native-tls"] }

# HTTP
axum = "0.7"
tower = "0.4"
tower-http = { version = "0.5", features = ["cors", "trace"] }

# Serde
serde = { version = "1", features = ["derive"] }
serde_json = "1"

# 加密 (视频解密)
aes = "0.8"
cbc = "0.1"
sha2 = "0.10"
base64 = "0.21"

# Windows
winreg = "0.52"
sysinfo = "0.30"

# 工具
regex = "1"
once_cell = "1"
dashmap = "5"
uuid = { version = "1", features = ["v4", "serde"] }

# 数据库
sqlx = { version = "0.7", features = ["runtime-tokio", "sqlite", "chrono"] }

# 日志
tracing = "0.1"
tracing-subscriber = "0.3"

# 错误处理
anyhow = "1"
thiserror = "1"

# 时间
chrono = "0.4"

# HTTP 客户端 (下载 + 调用 fallback API)
reqwest = { version = "0.12", features = ["json", "stream"] }

# 图片格式识别
image = { version = "0.25", default-features = false, features = ["png", "jpeg"] }
```

## 开发路线建议(分阶段)

### Phase 1:最小闭环(1-2 周)

1. ✅ Tauri 骨架 + WebView 空壳
2. ✅ CDP 客户端 + 启动豆包 (`ensure_doubao_started`)
3. ✅ 账号列表读取(只做 DOM 通道)
4. ✅ WebView 显示账号列表

**里程碑**:启动 app 能看到当前豆包账号列表。

### Phase 2:生图核心(1-2 周)

5. ✅ 生图请求 submit(fill prompt + click send)
6. ✅ Network 监听 + image_ori_raw 挖图
7. ✅ 结果下载到本地
8. ✅ HTTP API `POST /api/generate` + `GET /api/tasks/{id}`

**里程碑**:外部 curl 能调你的 API 生一张图。

### Phase 3:生视频(2 周)

9. ✅ 视频时长拖拽滑块
10. ✅ 参考图上传(CDP 拦截文件对话框)
11. ✅ 视频结果三条路径(直链 / 下载按钮 / fallback 解密)
12. ✅ Rust AES-CBC 解密实现 + 单元测试

**里程碑**:外部 API 能生无水印视频。

### Phase 4:并发 + 多轮确认(1-2 周)

13. ✅ parallelTracking 身份锚
14. ✅ 多轮确认自动点"继续生成"
15. ✅ 任务队列排队
16. ✅ WebView 显示任务进度

**里程碑**:支持多个任务排队,自动处理确认流程。

### Phase 5:健壮性(1 周)

17. ✅ 用户干扰检测(切会话/切账号/新建对话)
18. ✅ CDP 断线自动重连
19. ✅ 任务状态 SQLite 持久化,app 重启后能恢复
20. ✅ 原生文件对话框 fallback(调外部 PS 辅助)

**里程碑**:可以放心跑一天。

## 几个关键决策待办

### 决策 1:账号调度

对外 API 要不要暴露 `account` 参数?**推荐暴露**。

```http
POST /api/generate
{
  "account": "张三",       ← 可选,默认用当前账号
  "type": "video",
  "prompt": "...",
  "model_name": "Seedance 2.5",
  ...
}
```

如果指定了 `account`:
- 中转程序先检查当前账号,不是的话切换过去
- 切换中其他任务排队
- 切换失败返回 503

### 决策 2:同步 vs 异步

**强烈推荐异步 task_id + 轮询**。原因:视频 5-15 分钟,HTTP 长连接撑不住。

```http
POST /api/generate → 立即返回 { task_id, status: 'pending' }
GET  /api/tasks/{id} → { status: 'generating', stage: 'checking', progress: 0.4, result: null }
GET  /api/tasks/{id} → { status: 'completed', result: { url: '...', bytes: ... } }
GET  /api/tasks/{id}/download → 流式返回文件
```

同时支持 Webhook 回调:

```http
POST /api/generate
{
  ...,
  "callback_url": "https://my-service/doubao-webhook"
}
```

### 决策 3:HTTP 认证

**对外 API 必须有鉴权**。至少:
- Bearer Token(简单)
- 或者 HMAC 签名(更安全)

中转程序只在本机 loopback 暴露 → 用 Token 就够:
```http
Authorization: Bearer <long-random-string>
```

Token 在 app 首次启动时生成并显示在 UI,用户配置到调用方。

### 决策 4:并发限制

```rust
// 每账号最多 1 个任务同时跑
let per_account_lock = account_locks.get(&account).lock().await;

// 全局最多 3 个任务同时跑(避免豆包风控)
let global_sem = global_semaphore.acquire().await;
```

### 决策 5:超时

- 单任务:视频 20 分钟,图片 5 分钟
- 任务队列:最多 50 个等待
- API 调用等待(同步模式):5 秒后转异步

## 性能和稳定性注意

### CDP 事件数据量

豆包页面在生成时会推大量 SSE 事件,16MB/消息 + 5000 消息队列。Rust 侧 broadcast channel 要设大 buffer(1024+)。

### 内存泄漏点

- 大整数 JSON 修正后要全量 clone,长时间跑可能占用高内存
- `tokio::spawn` 的任务要有退出条件,不要无限 loop

### 测试策略

1. **单元测试**:视频解密、JSON 修正、注册表读取
2. **集成测试**(难做):mock 一个假 CDP server,校验工具操作序列正确
3. **端到端手动测试**:每个 phase 都在真豆包上跑一遍

### 日志

```rust
tracing_subscriber::fmt()
    .with_env_filter("info,tauri_doubao=debug")
    .init();
```

关键日志点:
- CDP 命令发送/响应
- 每个 Step 的开始/结束
- Network 事件分类
- 任务状态变化
- 错误(带完整栈)

## 下一步

**本文档是开发输入,不是开发指南**。真正开始写代码时:

1. 把 3 个 JS 文件复制到 `src-tauri/src/doubao/scripts/`
2. 按 Phase 1 的 4 个步骤跑最小闭环
3. 每个 Phase 完成都做端到端测试
4. 遇到和沈浪不一致的行为 → 查本文档对应章节的代码片段
5. 遇到本文档没覆盖的问题 → 回到解压目录读相关 PS/JS 源码

**项目内参考库**:
```
docs/analysis/reference/
├── native/                   # 沈浪 Windows 原始 PS1 + JS
└── native-shadow-bin/        # 原生 tracker.exe + flag
```

**项目内 Rust 要嵌入的 JS**:
```
docs/analysis/scripts/
├── doubao-generation-page.js
├── doubao-result-reader.js
└── doubao-video-decoder.js
```

这是你的"原生代码参考库"。
