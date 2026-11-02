# 03 · 系统架构

## 高层架构图

```
┌────────────────────────────────────────────────────────────────────┐
│  doubao-relay Tauri App                                            │
│                                                                    │
│  ┌──────────────────────────┐  invoke  ┌─────────────────────────┐ │
│  │  WebView (TypeScript)    │ ◄──────► │  Rust Backend           │ │
│  │                          │          │                         │ │
│  │  ┌────────────────────┐  │          │  ┌─────────────────┐    │ │
│  │  │ 账号列表页          │  │          │  │ Tauri 命令层    │    │ │
│  │  │ 任务列表页          │  │          │  └────────┬────────┘    │ │
│  │  │ 任务详情页          │  │          │           │             │ │
│  │  │ 设置页              │  │          │  ┌────────▼────────┐    │ │
│  │  └────────────────────┘  │          │  │ 核心服务层       │    │ │
│  └──────────────────────────┘          │  │ ─────────────── │    │ │
│                                        │  │ 账号管理器       │    │ │
│                                        │  │ 任务调度器       │    │ │
│                                        │  │ 生成管线         │    │ │
│                                        │  │ 捕获管线         │    │ │
│                                        │  │ 视频解密         │    │ │
│                                        │  └────────┬────────┘    │ │
│                                        │           │             │ │
│                                        │  ┌────────▼────────┐    │ │
│                                        │  │ CDP 客户端层     │    │ │
│                                        │  └────────┬────────┘    │ │
│                                        │           │             │ │
│                                        │  ┌────────▼────────┐    │ │
│                                        │  │ 平台适配层       │    │ │
│                                        │  │ (Win / macOS)   │    │ │
│                                        │  └────────┬────────┘    │ │
│                                        │           │             │ │
│                                        │  ┌────────▼────────┐    │ │
│                                        │  │ 持久化层         │    │ │
│                                        │  │ (SQLite + FS)   │    │ │
│                                        │  └─────────────────┘    │ │
│                                        │                         │ │
│                                        │  ┌─────────────────┐    │ │
│                                        │  │ HTTP API 服务器  │    │ │
│                                        │  │ (axum,独立线程) │    │ │
│                                        │  └────────▲────────┘    │ │
│                                        └───────────┼─────────────┘ │
└────────────────────────────────────────────────────┼───────────────┘
                                                    │
          ┌─────────────────────────────────────────┼──────────────┐
          │                                         │              │
          ▼                                         ▼              ▼
   豆包桌面端                                 外部 API 调用方      本地日志
   (CDP :9707)                              (HTTP :8787)         + SQLite
```

## 模块划分

### M1 · WebView(TypeScript)

**职责**:账号列表、任务进度、设置界面。

**依赖**:仅通过 Tauri invoke 调 Rust 命令,不直连 CDP、不直访豆包。

**文件**:`frontend/src/*.ts`, `frontend/src/pages/*.vue`(或 React,待定)。

### M2 · Tauri 命令层(Rust)

**职责**:实现 `#[tauri::command]` 注解的命令,是 WebView ↔ 核心服务的唯一桥梁。

**依赖**:核心服务层。

**不做**:业务逻辑,只做参数校验 + 转发 + 格式化返回。

### M3 · 核心服务层

子模块划分:

#### M3.1 · 账号管理器 (`account_manager`)
- `list_accounts()` → `Vec<Account>`
- `switch_account(target: &str)` → `Result<()>`
- `current_account()` → `Option<String>`
- 内部管理账号切换的互斥锁(切换时阻塞所有生成)

#### M3.2 · 任务调度器 (`task_scheduler`)
- 入队、按账号串行、全局并发限流
- 维护任务状态机
- 冲突检测(用户同时在豆包里打字 / 切换账号 → 任务警告)

#### M3.3 · 生成管线 (`generation_pipeline`)
- 对应 shenlang-analysis/04-generation-pipeline.md 的 prepare→fill→send 流程
- 调用 CDP 客户端层执行

#### M3.4 · 捕获管线 (`capture_pipeline`)
- Network 事件监听
- 多轮确认 feedback 处理
- 结果识别与下载

#### M3.5 · 视频解密 (`video_decoder`)
- AES-128-CBC 本地实现
- 对应 shenlang-analysis/06-video-decryption.md

### M4 · CDP 客户端层 (`cdp`)

**职责**:封装 WebSocket 连接、命令发送/响应匹配、事件分发。

**不做**:业务判断。

**关键 API**:
```rust
struct CdpSession
impl CdpSession {
    async fn connect(ws_url: &str) -> Result<Self>
    async fn send(&self, method: &str, params: Value) -> Result<Value>
    async fn evaluate(&self, expression: &str, await_promise: bool) -> Result<Value>
    fn subscribe_events(&self) -> broadcast::Receiver<Value>
    async fn close(self)
}
```

### M5 · 平台适配层 (`platform`)

**职责**:封装 Windows 和 macOS 的差异 API,核心服务只依赖 trait,不写 cfg。

```rust
#[async_trait]
pub trait Platform {
    async fn find_doubao_exe(&self) -> Result<PathBuf>;
    async fn launch_doubao_with_cdp(&self, exe: &Path, port: u16) -> Result<()>;
    async fn kill_doubao(&self) -> Result<()>;
    async fn is_doubao_running(&self) -> bool;
    fn app_data_dir(&self) -> PathBuf;
    fn log_dir(&self) -> PathBuf;
}
```

实现:`WindowsPlatform` + `MacosPlatform`,启动时根据 `cfg!(target_os)` 实例化。

### M6 · 持久化层 (`storage`)

**职责**:SQLite 读写 + 结果文件管理。

```rust
pub struct Storage {
    pool: SqlitePool,
    files_dir: PathBuf,
}
impl Storage {
    async fn insert_task(&self, task: &NewTask) -> Result<TaskId>;
    async fn get_task(&self, id: &TaskId) -> Result<Option<Task>>;
    async fn update_task_status(&self, id: &TaskId, s: TaskStatus) -> Result<()>;
    async fn list_recent_tasks(&self, limit: u32) -> Result<Vec<Task>>;
    async fn save_result_file(&self, task_id: &TaskId, index: u32, bytes: &[u8], mime: &str) -> Result<ResultFile>;
    async fn read_result_file(&self, task_id: &TaskId, index: u32) -> Result<impl AsyncRead>;
    async fn gc_expired(&self, now: i64) -> Result<GcReport>;
}
```

### M7 · HTTP API 服务器 (`http_server`)

**职责**:axum 路由 + 鉴权中间件 + 请求转发。

**线程模型**:在独立 tokio runtime 任务里跑,与 Tauri UI 线程隔离。

### M8 · 配置 (`config`)

**职责**:读写 TOML 配置文件。

```toml
[server]
port = 8787
bearer_token = "..."

[concurrency]
global_max = 3
per_account_max = 1

[storage]
result_retention_hours = 24
metadata_retention_days = 7

[logging]
level = "info"
```

## 线程/异步模型

```
Tauri 主线程
├── UI 渲染(WebView)
└── #[tauri::command] 调用

tokio runtime(主)
├── HTTP server(axum)
├── 任务调度器主循环
├── CDP 事件分发器
└── 日志/指标后台 flush

tokio runtime(每个 CDP 连接)
├── WebSocket 读循环
├── WebSocket 写队列
└── 事件 broadcast
```

**关键原则**:
- UI 线程绝不阻塞;所有耗时操作 spawn 到 tokio
- Tauri 命令返回类型 `Result<T, String>`,错误转为中文字符串给前端

## 依赖关系(模块依赖图)

```
HTTP Server (M7)
    │
    └──► Tauri 命令层 (M2)
            │
            └──► 核心服务 (M3)
                    │
                    ├──► CDP 客户端 (M4)
                    ├──► 平台适配 (M5)
                    ├──► 持久化 (M6)
                    └──► 配置 (M8)

WebView (M1)
    │
    └──► Tauri 命令层 (M2)
```

**不允许反向依赖**。核心服务不知道 HTTP server 存在。CDP 客户端不知道生成管线存在。

## 关键技术选型

| 项 | 选型 | 理由 |
|---|---|---|
| 应用框架 | Tauri 2 | 跨平台桌面 + 比 Electron 轻 10 倍 |
| 后端语言 | Rust 2021 | 并发安全 + 单二进制部署 + 跨平台编译链成熟 |
| 异步 runtime | tokio | 生态最成熟 |
| WebSocket | tokio-tungstenite | 轻、不绑定 http client |
| HTTP 服务器 | axum | 路由简洁 + tower 生态 |
| JSON | serde_json + 大整数预处理 | 必须保留 16 位整数精度 |
| SQL | sqlx + SQLite | 编译时查询校验 |
| 加密 | aes + cbc + sha2 (RustCrypto) | 纯 Rust,无 openssl 依赖 |
| 图像识别 | image | 识别 PNG/JPEG 格式 |
| HTTP 客户端 | reqwest(内部用) | 下载视频 + 调 fallback API |
| 日志 | tracing + tracing-subscriber | 结构化日志标准 |
| 配置 | toml + serde | 简单 |
| 前端 | TypeScript + 原生 DOM 或 Preact | 简单 UI 不上重框架 |

**不选的**:
- ❌ chromiumoxide / headless_chrome — 面向启动 Chromium 的,不适合连 Electron
- ❌ Electron — 太重,Tauri 够用
- ❌ openssl — Rust 原生 crypto 够用,减少依赖

## 启动流程

```
1. Tauri app 启动
2. 读 config.toml
3. 初始化 Storage(SQLite migrate + 建目录)
4. 启动 HTTP server(独立 tokio 任务)
5. 探测豆包状态:
   a. 是否已跑?
   b. 9707 是否通?
   c. 不通 → 让用户在 UI 点"连接豆包"手动启动
6. 启动 UI,展示账号列表
7. 后台心跳(每 30s)检查 CDP 可用性
```

**关键**:不在启动时就强行连豆包。让用户看到 UI 后主动点"连接",避免启动失败卡住。

## 关闭流程

```
1. 用户点关闭 / Cmd+Q
2. 停止接受新 HTTP 请求
3. 当前任务:
   a. pending → 立即取消
   b. preparing → 立即取消
   c. generating → 等 10 秒,之后强制标记 failed
4. 关闭 CDP 连接(不杀豆包)
5. flush 日志 + SQLite 关闭
6. 退出
```

**不杀豆包**,因为是用户自己的工具。
