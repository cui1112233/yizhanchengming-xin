# 07 · 数据模型

## SQLite 数据库

**文件位置**:
- Windows: `%APPDATA%\doubao-relay\relay.db`
- macOS: `~/Library/Application Support/doubao-relay/relay.db`

**WAL 模式**:启用,减少写锁。

**版本迁移**:用 `sqlx-cli` migrate,迁移文件 `migrations/*.sql`。

## 表设计

### `tasks` — 任务主表

```sql
CREATE TABLE tasks (
    id              TEXT PRIMARY KEY,              -- ULID
    account         TEXT NOT NULL,                 -- 账号 name
    type            TEXT NOT NULL,                 -- 'image' | 'video'
    model_name      TEXT NOT NULL,
    prompt          TEXT NOT NULL,
    ratio           TEXT NOT NULL,
    duration        INTEGER,                       -- 视频独有
    reference_files TEXT NOT NULL DEFAULT '[]',    -- JSON array of upload_ids
    confirmation_mode TEXT NOT NULL DEFAULT 'auto',
    auto_use_membership INTEGER NOT NULL DEFAULT 1,
    callback_url    TEXT,

    status          TEXT NOT NULL,                 -- 见下
    stage           TEXT,                          -- 见下
    stage_message   TEXT,                          -- 当前阶段的用户可读消息
    error_code      TEXT,                          -- 失败时的错误码
    error_message   TEXT,
    error_details   TEXT,                         -- JSON

    conversation_id TEXT,                          -- 绑定的豆包会话 ID
    submission_id   TEXT,                          -- 绑定的豆包消息 ID
    local_message_id TEXT,                         -- 本地消息 ID(submissionIdentity)

    watermark_source TEXT,                         -- 'conversation_original' / 'lr_unwatermarked' / 'fallback_decrypted'
    result_files    TEXT,                          -- JSON array,见下

    created_at      INTEGER NOT NULL,              -- Unix ms
    started_at      INTEGER,                       -- 开始 preparing 的时间
    completed_at    INTEGER,                       -- 终态时间
    expires_at      INTEGER NOT NULL               -- 结果文件过期时间
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_account ON tasks(account);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
```

### `task_stage_history` — 阶段变化日志

```sql
CREATE TABLE task_stage_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id         TEXT NOT NULL,
    stage           TEXT NOT NULL,                 -- pending/preparing/checking/...
    message         TEXT,
    ts              INTEGER NOT NULL,              -- Unix ms
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_stage_history_task ON task_stage_history(task_id, ts);
```

### `uploads` — 预上传文件

```sql
CREATE TABLE uploads (
    id              TEXT PRIMARY KEY,              -- ULID
    mime            TEXT NOT NULL,
    bytes           INTEGER NOT NULL,
    path            TEXT NOT NULL,                 -- 文件实际路径
    created_at      INTEGER NOT NULL,
    expires_at      INTEGER NOT NULL
);

CREATE INDEX idx_uploads_expires ON uploads(expires_at);
```

### `webhook_log` — Webhook 回调记录

```sql
CREATE TABLE webhook_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id         TEXT NOT NULL,
    event           TEXT NOT NULL,                 -- 'task.completed' / 'task.failed' / ...
    url             TEXT NOT NULL,
    attempt         INTEGER NOT NULL,              -- 1-based
    http_status     INTEGER,
    response_body   TEXT,
    error_message   TEXT,
    sent_at         INTEGER NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_webhook_task ON webhook_log(task_id);
```

### `metadata` — 杂项键值对

```sql
CREATE TABLE metadata (
    key             TEXT PRIMARY KEY,
    value           TEXT NOT NULL,
    updated_at      INTEGER NOT NULL
);
```

存储如:schema_version、last_startup_at、ui_compat_last_check 等。

## 枚举值

### `TaskStatus`

| 值 | 含义 | 终态 |
|---|---|---|
| `pending` | 入队,等待调度 | 否 |
| `preparing` | 调度到,正在准备豆包页面 | 否 |
| `generating` | 已提交豆包,正在生成 | 否 |
| `waiting_user_confirm` | manual 模式等待用户确认 | 否 |
| `completed` | 成功 | ✅ |
| `failed` | 失败 | ✅ |
| `cancelled` | 被取消 | ✅ |

### `TaskStage`(generating 下的细分)

| 值 | 含义 |
|---|---|
| `preparing` | 选模型/比例/时长/上传参考图/填 prompt |
| `sent` | 已点发送 |
| `checking` | 豆包确认参数(未开始生成) |
| `quota_checking` | 豆包确认会员配额 |
| `rendering` | 豆包真正在生成 |
| `rendering_unverified` | 豆包只回"正在渲染",未开始 |
| `downloading` | 已返回,下载中 |
| `result` | 识别到结果 |
| `refused` | 豆包拒绝生成 |

### `WatermarkSource`

| 值 | 含义 |
|---|---|
| `conversation_original` | 从消息 JSON 挖的 image_ori_raw |
| `lr_unwatermarked` | 带 lr=unwatermarked 的直链 |
| `fallback_decrypted` | fallback API + 本地 AES 解密 |

## result_files JSON 格式

```json
[
  {
    "index": 0,
    "filename": "image_1.png",
    "relative_path": "results/01HQZ9.../image_1.png",
    "mime": "image/png",
    "bytes": 2456321,
    "width": 1920,
    "height": 1080,
    "duration_ms": null
  }
]
```

视频项目额外有 `duration_ms`。

## 文件布局

```
Windows: %APPDATA%\doubao-relay\
macOS:   ~/Library/Application Support/doubao-relay/

<app data root>/
├── config.toml                    用户配置
├── relay.db                       SQLite 主库
├── relay.db-wal                   WAL
├── relay.db-shm                   SHM
│
├── uploads/
│   └── {upload_id}.{ext}          预上传的参考图/视频
│
├── results/
│   └── {task_id}/
│       ├── image_1.png
│       ├── image_2.png
│       └── video_1.mp4
│
├── logs/
│   ├── doubao-relay-2026-10-16.log
│   ├── doubao-relay-2026-10-15.log
│   └── ...
│
└── cdp-debug.log                  可选,开启 CDP 日志时
```

## 配置文件 `config.toml`

```toml
# ~/.../doubao-relay/config.toml

[server]
host = "127.0.0.1"
port = 8787
bearer_token = "<32 bytes hex>"

[concurrency]
global_max = 3
per_account_max = 1
queue_depth_max = 50

[storage]
result_retention_hours = 24
metadata_retention_days = 7
upload_expiry_hours = 1

[logging]
level = "info"
file_max_size_mb = 10
file_retention_days = 7
cdp_debug_log = false

[doubao]
cdp_port = 9707
cdp_address = "127.0.0.1"
auto_connect_on_startup = true
ui_compat_check_on_startup = true

[ui]
minimize_to_tray_on_close = true
start_minimized = false
```

**约定**:
- 启动时若 `bearer_token` 不存在,自动生成并写回
- 配置热重载:logging / ui 字段支持热重载,其他字段需重启

## 运行时状态(内存)

Rust 侧用这些类型表达运行时状态:

### `AccountState`

```rust
pub struct AccountState {
    pub list: Vec<Account>,
    pub current: Option<String>,
    pub last_read_at: Instant,
    pub switch_mutex: Mutex<()>,       // 切换时互斥
}

pub struct Account {
    pub name: String,
    pub subscription: Option<String>,
    pub label: String,
    pub is_current: bool,
    pub is_ambiguous: bool,
}
```

### `TaskHandle`

```rust
pub struct TaskHandle {
    pub task_id: TaskId,
    pub status: Arc<RwLock<TaskStatus>>,
    pub cancel_token: CancellationToken,
    pub completion_notify: Arc<Notify>,
}
```

### `CdpSessions`

```rust
pub struct CdpSessions {
    pub chat: Option<Arc<CdpSession>>,          // doubao-chat 的 session
    pub background: Option<Arc<CdpSession>>,    // doubao-background 的 session
    pub last_health_check: Instant,
}
```

### `TaskQueue`

```rust
pub struct TaskQueue {
    pub pending: VecDeque<TaskId>,
    pub running: HashMap<TaskId, TaskHandle>,
    pub account_locks: HashMap<String, Arc<Mutex<()>>>,
    pub global_semaphore: Arc<Semaphore>,       // 全局并发限流
}
```

## 迁移脚本约定

```
migrations/
├── 20261016000001_initial.sql
├── 20261101000001_add_webhook_log.sql
└── ...
```

每个脚本幂等。回滚脚本可选(用 `*.down.sql`)。

## 清理策略

### 定时 GC

每 1 小时跑一次:
1. 删除 `expires_at < now` 的 `tasks.result_files` 对应的文件(并把 `result_files` 置为 `null`)
2. 删除 `created_at < now - metadata_retention_days` 的 tasks(级联删 `task_stage_history` 和 `webhook_log`)
3. 删除 `expires_at < now` 的 uploads(含物理文件)
4. 日志文件按天滚动,保留 N 天

### 手动触发 GC

Tauri 命令 `maintenance.gc()`,方便设置页"立即清理"按钮使用。

## 大小估算

- 一个任务:tasks 行 ~500 字节 + stage history 10 行 × 200 字节 = ~2.5 KB
- 一张结果图:1-5 MB
- 一段视频:5-50 MB

**典型用户**(每天 100 任务 + 结果保留 1 天):
- SQLite: 2.5 KB × 100 × 7 天 = 1.75 MB
- 结果文件:50 × 2 MB = 100 MB(图) 或 10 × 20 MB = 200 MB(视频)

**磁盘预算**:预留 1-10 GB,正常用户够。
