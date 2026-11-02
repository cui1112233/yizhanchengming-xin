// Tauri IPC 封装(docs/spec/05-tauri-ipc.md):命令名把 spec 的 `account.list` 写成 `account_list`,
// 返回值与事件 payload 均为后端 `to_camel` 后的 camelCase。
// 本文件之外的代码不得直接 import @tauri-apps/api。
import { convertFileSrc, invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';
import { readText, writeText } from '@tauri-apps/plugin-clipboard-manager';
import { open as openDialog, save } from '@tauri-apps/plugin-dialog';

/** 后端 `RelayError` 的 IPC 形态(`IpcError`)。 */
export type RelayError = { code: string; message: string; details?: unknown };

/** `doubao.status` 返回 / `doubao_status_changed` payload。 */
export type DoubaoStatus = {
  running: boolean;
  cdpConnected: boolean;
  cdpPort: number;
  doubaoVersionDetected: string | null;
  /** spec 05 只有 ok/warn/error;后端自检本身没跑成时为 unknown(计划"已知问题" Task 1.8 ⑩)。 */
  uiCompatCheck: 'ok' | 'warn' | 'error' | 'unknown';
  uiCompatDetails: {
    status: string;
    accountButtonFound: boolean;
    inputFound: boolean;
    sendButtonFound: boolean;
    checkedAt: number;
  };
  chatTargets: number;
  reconnecting: boolean;
  lastCheckAt: number;
};

export type Account = {
  name: string;
  subscription: string | null;
  label: string;
  isCurrent: boolean;
  isAmbiguous: boolean;
};

/** `account.list` / `account.refresh` 返回,`account_list_changed` payload。 */
export type AccountList = {
  current: string | null;
  accounts: Account[];
  readAt: number;
};

/** `account.switch` 返回。 */
export type SwitchResult = {
  ok: boolean;
  previous: string | null;
  current: string;
  switchedAt: number;
};

/** `account_switched` 事件 payload。 */
export type AccountSwitchedEvent = {
  previous: string | null;
  current: string;
  switchedAt: number;
};

export type ReconnectAction = 'already_connected' | 'reconnected' | 'restarted' | 'started';

/** `doubao.reconnect` 返回。 */
export type ReconnectResult = { action: ReconnectAction; cdpReadyAt: number };

/** `app.get_version` 返回。 */
export type VersionInfo = {
  version: string;
  buildTarget: string;
  tauriVersion: string;
  rustcVersion: string;
};

/** `log.recent` 的一条。 */
export type LogEntry = {
  ts: number;
  level: string;
  target: string;
  message: string;
  fields: Record<string, unknown>;
};

// ---- 任务(spec 05 task.* / 04 GET /tasks/{id},后端 to_camel 后的形状) ----

export type TaskStatus =
  | 'pending'
  | 'preparing'
  | 'generating'
  | 'waiting_user_confirm'
  | 'completed'
  | 'failed'
  | 'cancelled';

export type TaskStage =
  | 'preparing'
  | 'sent'
  | 'checking'
  | 'quota_checking'
  | 'rendering'
  | 'rendering_unverified'
  | 'downloading'
  | 'result'
  | 'refused';

export type TaskType = 'image' | 'video';
export type ConfirmationMode = 'auto' | 'manual';

/** `task.list` 的一行 / `task_created` payload(列表行比 HTTP 多 `promptPreview`)。 */
export type TaskSummary = {
  taskId: string;
  account: string;
  type: TaskType;
  status: TaskStatus;
  stage: TaskStage | null;
  createdAt: number;
  completedAt: number | null;
  /** prompt 前 60 字(仅 `task.list` 有;`task_created` 事件没有)。 */
  promptPreview?: string;
};

export type TaskListFilter = {
  status?: TaskStatus[];
  account?: string;
  limit?: number;
  cursor?: string;
};

export type TaskListResult = { tasks: TaskSummary[]; nextCursor: string | null };

/** 提交请求(camelCase,后端 to_snake 后与 HTTP `POST /generate` 同一解析)。 */
export type GenerateRequest = {
  account: string;
  type: TaskType;
  modelName: string;
  prompt: string;
  ratio: string;
  duration?: number;
  referenceFiles?: string[];
  confirmationMode?: ConfirmationMode;
  autoUseMembership?: boolean;
  callbackUrl?: string;
};

export type StageEntry = { stage: string; message?: string | null; ts: number };

export type TaskErrorInfo = { code: string; message: string; details?: unknown };

export type ResultFileView = {
  index: number;
  filename: string;
  mime: string;
  bytes: number;
  width: number | null;
  height: number | null;
  durationMs: number | null;
  downloadUrl: string;
  /** 本机绝对路径(IPC 独有),经 `fileSrc` 显示缩略图;路径不安全时为 null。 */
  localPath: string | null;
};

/** `task.get` 返回 / `task_completed` payload。 */
export type TaskDetail = {
  taskId: string;
  account: string;
  type: TaskType;
  request: GenerateRequest;
  status: TaskStatus;
  stage: TaskStage | null;
  stageMessage: string | null;
  message: string | null;
  progress: number | null;
  error: TaskErrorInfo | null;
  conversationId: string | null;
  submissionId: string | null;
  watermarkSource: string | null;
  createdAt: number;
  startedAt: number | null;
  completedAt: number | null;
  failedAt?: number | null;
  expiresAt: number;
  stageHistory: StageEntry[];
  queuePosition: number | null;
  result?: { files: ResultFileView[]; watermarkSource: string | null };
};

/** `task_updated` payload。 */
export type TaskUpdatedEvent = {
  taskId: string;
  status: TaskStatus;
  stage: TaskStage | null;
  message: string | null;
  progress: number | null;
  updatedAt: number;
};

export type SubmitResult = {
  taskId: string;
  status: 'pending';
  queuePosition: number;
  createdAt: number;
};

export type UploadKind = 'image' | 'video_reference';
export type UploadResult = { uploadId: string; mime: string; bytes: number; expiresAt: number };

export const taskList = (filter?: TaskListFilter) =>
  invoke<TaskListResult>('task_list', { filter: filter ?? null });
export const taskGet = (taskId: string) => invoke<TaskDetail>('task_get', { taskId });
export const taskSubmit = (request: GenerateRequest) =>
  invoke<SubmitResult>('task_submit', { request });
export const taskCancel = (taskId: string) =>
  invoke<{ ok: boolean; cancelledFromStatus: TaskStatus }>('task_cancel', { taskId });
export const taskConfirm = (taskId: string) =>
  invoke<{ ok: boolean; newStage: string }>('task_confirm', { taskId });
export const taskOpenResultFolder = (taskId: string) =>
  invoke<{ ok: boolean; path: string }>('task_open_result_folder', { taskId });
export const uploadAdd = (filePath: string, kind?: UploadKind) =>
  invoke<UploadResult>('upload_add', { filePath, kind: kind ?? null });

/** 系统文件选择框选参考图(tauri-plugin-dialog);取消返回空数组。 */
export async function pickImageFiles(): Promise<string[]> {
  const picked = await openDialog({
    title: '选择参考图',
    multiple: true,
    directory: false,
    filters: [{ name: '图片', extensions: ['png', 'jpg', 'jpeg', 'webp'] }],
  });
  if (picked === null) return [];
  return Array.isArray(picked) ? picked : [picked];
}

/** 本地文件 → WebView 可加载的 asset URL(需 tauri.conf assetProtocol)。 */
export const fileSrc = (path: string) => convertFileSrc(path);

export const accountList = () => invoke<AccountList>('account_list');
export const accountRefresh = () => invoke<AccountList>('account_refresh');
export const accountSwitch = (target: string) => invoke<SwitchResult>('account_switch', { target });
export const doubaoStatus = () => invoke<DoubaoStatus>('doubao_status');
export const doubaoReconnect = (force = false) =>
  invoke<ReconnectResult>('doubao_reconnect', { force });
export const appGetVersion = () => invoke<VersionInfo>('app_get_version');
/** `app_api_status`:HTTP API 是否在监听。 */
export interface ApiStatus {
  enabled: boolean;
  listening: boolean;
  host: string;
  port: number;
  error: string | null;
}
export const appApiStatus = () => invoke<ApiStatus>('app_api_status');
export const appFindAvailablePort = () => invoke<{ port: number }>('app_find_available_port');
export const appPairV88 = (baseUrl: string, code: string) =>
  invoke<{ ok: boolean; executorId: string }>('app_pair_v88', { baseUrl, code });
export const logRecent = (limit?: number, level?: string) =>
  invoke<{ entries: LogEntry[] }>('log_recent', { limit, level });

/**
 * 订阅 Rust → WebView 事件。返回同步的取消函数(内部 listen 是异步的,
 * 取消时若尚未注册完成则在注册完成后立刻取消),方便在 useEffect 里直接 return。
 */
export function onEvent<T>(name: string, cb: (payload: T) => void): () => void {
  let cancelled = false;
  let unlisten: (() => void) | null = null;
  listen<T>(name, (event) => {
    if (!cancelled) cb(event.payload);
  })
    .then((fn) => {
      if (cancelled) fn();
      else unlisten = fn;
    })
    .catch((e) => {
      console.error(`订阅事件 ${name} 失败`, e);
    });
  return () => {
    cancelled = true;
    unlisten?.();
    unlisten = null;
  };
}

/** 任何 catch 到的值 → `RelayError`;非结构化错误归为 `system_internal_error`。 */
export function asRelayError(e: unknown): RelayError {
  if (e && typeof e === 'object' && 'code' in e && typeof (e as RelayError).code === 'string') {
    const err = e as RelayError;
    const out: RelayError = {
      code: err.code,
      message: typeof err.message === 'string' ? err.message : '未知错误',
    };
    if (err.details !== undefined) out.details = err.details;
    return out;
  }
  let message: string;
  if (e instanceof Error) message = e.message;
  else if (typeof e === 'string') message = e;
  else if (e === undefined || e === null) message = '发生未知错误,请查看日志';
  else {
    try {
      message = JSON.stringify(e);
    } catch {
      message = String(e);
    }
  }
  return { code: 'system_internal_error', message };
}

/** `doubao_*` 错误:豆包没连上 / 没登录,账号页用空状态表达,不弹 toast。 */
export function isDoubaoUnavailable(e: RelayError): boolean {
  return e.code.startsWith('doubao_');
}

// ---- Task 5.1:设置 / 日志 / 调试面板 ----

/** `app.get_config` 返回:`Config`(docs/spec/07-data-model.md)的 camelCase 形态,含 bearerToken。 */
export type AppConfig = {
  server: { enabled: boolean; host: string; port: number; bearerToken: string };
  concurrency: { globalMax: number; perAccountMax: number; queueDepthMax: number };
  storage: { resultRetentionHours: number; metadataRetentionDays: number; uploadExpiryHours: number };
  logging: { level: string; fileMaxSizeMb: number; fileRetentionDays: number; cdpDebugLog: boolean };
  doubao: {
    cdpPort: number;
    cdpAddress: string;
    autoConnectOnStartup: boolean;
    uiCompatCheckOnStartup: boolean;
  };
  v88Executor: { baseUrl: string; executorId: string; executorToken: string };
  ui: { minimizeToTrayOnClose: boolean; startMinimized: boolean };
};

/** `app.save_config` 的 patch:各分组可选、字段可选(后端深合并)。 */
export type ConfigPatch = { [G in keyof AppConfig]?: Partial<AppConfig[G]> };

export type SaveConfigResult = { ok: boolean; needsRestart: boolean };

/** `maintenance.gc` 返回(`GcReport`)。 */
export type GcReport = { resultFilesRemoved: number; tasksRemoved: number; uploadsRemoved: number };

/** `debug.cdp_log` 的一条 / `cdp_log_entry` payload。 */
export type CdpLogEntry = {
  ts: number;
  direction: 'send' | 'recv';
  method?: string;
  params?: unknown;
  result?: unknown;
  event?: unknown;
};

export const appGetConfig = () => invoke<AppConfig>('app_get_config');
export const appSaveConfig = (patch: ConfigPatch) =>
  invoke<SaveConfigResult>('app_save_config', { patch });
export const appRegenerateToken = () => invoke<{ bearerToken: string }>('app_regenerate_token');
export const maintenanceGc = () => invoke<GcReport>('maintenance_gc');
export const logExport = (path: string, level?: string) =>
  invoke<{ ok: boolean; path: string; count: number }>('log_export', { path, level });
export const debugCdpLog = (limit?: number) =>
  invoke<{ entries: CdpLogEntry[]; enabled: boolean }>('debug_cdp_log', { limit });

/** 剪贴板(tauri-plugin-clipboard-manager,capability `clipboard-manager:allow-write-text/read-text`)。 */
export const clipboardWrite = (text: string) => writeText(text);
export const clipboardRead = () => readText();

/** 保存对话框(tauri-plugin-dialog,capability `dialog:allow-save`);取消时返回 null。 */
export const saveFileDialog = (defaultPath: string, extension: string) =>
  save({ defaultPath, filters: [{ name: extension.toUpperCase(), extensions: [extension] }] });

// ---- Task 5.2:隐私声明 / 退出 ----

/** `app.privacy_status` / `app.privacy_accept` 返回。 */
export type PrivacyStatus = { accepted: boolean; acceptedAt: number | null };

export const privacyStatus = () => invoke<PrivacyStatus>('app_privacy_status');
export const privacyAccept = () => invoke<PrivacyStatus>('app_privacy_accept');
/** 走关闭流程退出(停 HTTP、取消排队任务、生成中任务最多等 10s、不杀豆包)。 */
export const appQuit = () => invoke<{ ok: boolean }>('app_quit');
