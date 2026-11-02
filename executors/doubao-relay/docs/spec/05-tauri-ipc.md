# 05 · Tauri IPC 命令规格

## 用途

**WebView ↔ Rust 后端** 的双向通信契约。WebView 通过 `invoke('<command>', args)` 调 Rust;Rust 通过 `app.emit_all('<event>', payload)` 推送事件给 WebView。

**不被 HTTP 调用方使用**。HTTP API 调用方见 `04-http-api.md`。

## 命名约定

- 命令名:`snake_case`,分组用 `.`,如 `account.list`、`task.submit`
- 事件名:`snake_case`,过去时,如 `task_updated`、`account_switched`
- 参数:`camelCase`(JS 习惯),Rust 侧 `#[serde(rename_all = "camelCase")]`

## 命令目录

| 命令 | 说明 |
|---|---|
| `app.get_config` | 读配置 |
| `app.save_config` | 写配置 |
| `app.get_version` | 版本信息 |
| `account.list` | 账号列表(等效 HTTP GET /accounts) |
| `account.switch` | 切换账号 |
| `account.refresh` | 强制重读 |
| `doubao.status` | 豆包连接状态 |
| `doubao.reconnect` | 重连豆包 |
| `task.list` | 任务列表 |
| `task.get` | 任务详情 |
| `task.submit` | 提交任务(等效 HTTP POST /generate,但便于 UI 内部触发) |
| `task.cancel` | 取消任务 |
| `task.confirm` | 手动确认 |
| `task.open_result_folder` | 在文件管理器打开结果文件所在目录 |
| `upload.add` | 预上传 |
| `log.recent` | 读最近日志 |
| `debug.cdp_log` | 读 CDP 命令日志 |

---

## 命令详细

### `app.get_config`

**请求**:无
**返回**:
```typescript
{
  server: { port: number; bearerToken: string };
  concurrency: { globalMax: number; perAccountMax: number };
  storage: { resultRetentionHours: number; metadataRetentionDays: number };
  logging: { level: 'error' | 'warn' | 'info' | 'debug' };
}
```

### `app.save_config`

**请求**:
```typescript
{
  server?: Partial<ServerConfig>;
  concurrency?: Partial<ConcurrencyConfig>;
  storage?: Partial<StorageConfig>;
  logging?: Partial<LoggingConfig>;
}
```
**返回**:`{ ok: true; needsRestart: boolean }`

**说明**:端口/token 变更需要重启 HTTP server。

### `app.get_version`

**返回**:
```typescript
{
  version: string;
  buildTarget: 'windows' | 'macos';
  tauriVersion: string;
  rustcVersion: string;
}
```

### `account.list`

**返回**:
```typescript
{
  current: string | null;
  accounts: Array<{
    name: string;
    subscription: string | null;
    label: string;
    isCurrent: boolean;
    isAmbiguous: boolean;
  }>;
  readAt: number;
}
```

### `account.switch`

**请求**:`{ target: string }`
**返回**:`{ ok: boolean; previous: string | null; current: string; switchedAt: number }`
**错误**:`AccountNotFound` / `AccountBusy` / `SwitchTimeout`

### `account.refresh`

强制重新从豆包读账号(不走缓存)。

**返回**:同 `account.list`。

### `doubao.status`

**返回**:
```typescript
{
  running: boolean;
  cdpConnected: boolean;
  cdpPort: number;
  doubaoVersionDetected: string | null;
  uiCompatCheck: 'ok' | 'warn' | 'error';
  uiCompatDetails: {
    accountButtonFound: boolean;
    inputFound: boolean;
    sendButtonFound: boolean;
  };
  chatTargets: number;
  lastCheckAt: number;
}
```

### `doubao.reconnect`

**请求**:`{ force?: boolean }` (force=true 强杀重启,否则先尝试优雅重连)
**返回**:`{ action: 'already_connected' | 'reconnected' | 'restarted'; cdpReadyAt: number }`

### `task.list`

**请求**:
```typescript
{
  status?: Array<TaskStatus>;
  account?: string;
  limit?: number;      // 默 20,最大 100
  cursor?: string;
}
```
**返回**:
```typescript
{
  tasks: Array<TaskSummary>;
  nextCursor: string | null;
}
```

### `task.get`

**请求**:`{ taskId: string }`
**返回**:完整 Task 对象,结构同 HTTP GET /tasks/{id}。

### `task.submit`

**请求**:同 HTTP POST /generate 的 body。
**返回**:`{ taskId: string; status: 'pending'; queuePosition: number; createdAt: number }`

**用途**:让 WebView 内置 UI(如"快速生成"面板)能直接提交任务,不走 HTTP。

### `task.cancel`

**请求**:`{ taskId: string }`
**返回**:`{ ok: boolean; cancelledFromStatus: TaskStatus }`

### `task.confirm`

**请求**:`{ taskId: string }`
**返回**:`{ ok: boolean; newStage: string }`

### `task.open_result_folder`

**请求**:`{ taskId: string }`
**返回**:`{ ok: boolean; path: string }`

**行为**:调系统 API 在 Finder (macOS) / Explorer (Windows) 打开结果文件所在目录。

### `upload.add`

**请求**:
```typescript
{
  filePath: string;    // Tauri 的 open dialog 返回的本地路径
  kind?: 'image' | 'video_reference';
}
```
**返回**:`{ uploadId: string; mime: string; bytes: number; expiresAt: number }`

**行为**:Rust 读本地文件,拷贝到 app data 目录,返回 upload_id。

### `log.recent`

**请求**:
```typescript
{
  limit?: number;      // 默 200
  level?: 'error' | 'warn' | 'info' | 'debug';   // 过滤
}
```
**返回**:
```typescript
{
  entries: Array<{
    ts: number;
    level: string;
    target: string;      // tracing target
    message: string;
    fields: Record<string, unknown>;
  }>;
}
```

### `debug.cdp_log`

**请求**:`{ limit?: number }`
**返回**:
```typescript
{
  entries: Array<{
    ts: number;
    direction: 'send' | 'recv';
    method?: string;
    params?: unknown;
    result?: unknown;
    event?: unknown;
  }>;
}
```

**用途**:豆包改版诊断用。默认关闭,设置页勾"开启 CDP 日志"才生效。

---

## 事件(Rust → WebView)

### `account_list_changed`

账号列表变化(账号增/删/切换)。

**Payload**:同 `account.list` 返回。

### `account_switched`

账号切换完成。

**Payload**:
```typescript
{
  previous: string | null;
  current: string;
  switchedAt: number;
}
```

### `task_created`

新任务入队。

**Payload**:`TaskSummary`

### `task_updated`

任务状态/阶段变化。

**Payload**:
```typescript
{
  taskId: string;
  status: TaskStatus;
  stage: string | null;
  message: string | null;
  progress: number | null;
  updatedAt: number;
}
```

### `task_completed`

任务完成(含 completed/failed/cancelled)。

**Payload**:完整 `Task` 对象。

### `doubao_status_changed`

豆包连接状态变化(连上/断开)。

**Payload**:同 `doubao.status` 返回。

### `log_entry`

有新日志条目(level >= 当前 UI 订阅级别)。

**Payload**:单条日志 entry。

### `cdp_log_entry`

有新 CDP 日志(仅在 debug.cdpLog 开启时)。

**Payload**:单条 CDP 日志 entry。

---

## 错误处理

Tauri 命令的 Rust 签名统一返回 `Result<T, IpcError>`:

```rust
#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct IpcError {
    pub code: String,          // 对标 11-error-codes.md
    pub message: String,       // 中文
    pub details: Option<Value>,
}
```

JS 侧:
```typescript
try {
  const result = await invoke('task.submit', { ... });
} catch (e: any) {
  // e 结构同 IpcError
  if (e.code === 'account_not_found') { ... }
}
```

---

## 调用示例

```typescript
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

// 1. 读账号
const { accounts, current } = await invoke<AccountListResult>('account.list');

// 2. 监听任务更新
const unlisten = await listen<TaskUpdate>('task_updated', (event) => {
  console.log('task updated:', event.payload);
});

// 3. 提交任务
const { taskId } = await invoke<SubmitResult>('task.submit', {
  account: '张三',
  type: 'image',
  modelName: 'Seedream 5.0 Pro',
  prompt: '蓝色小鸟',
  ratio: '16:9',
  referenceFiles: [],
  confirmationMode: 'auto',
  autoUseMembership: true,
});

// 4. 取消监听
unlisten();
```

## 并发语义

- 同一命令可被多个 WebView 页面并发调用,Rust 侧用 `Mutex`/`RwLock` 保护共享状态
- `task.submit` / `account.switch` 等变更性命令按先后到达顺序处理
- `account.list` 可能有短暂缓存(2 秒),强制刷新用 `account.refresh`
