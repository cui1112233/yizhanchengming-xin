// 纯函数视图模型:把后端状态映射为 UI 展示值,无副作用,便于单测。
import type {
  AccountList,
  AppConfig,
  DoubaoStatus,
  RelayError,
  StageEntry,
  TaskDetail,
  TaskErrorInfo,
  TaskStatus,
  TaskSummary,
  TaskType,
  TaskUpdatedEvent,
} from './api';

export type DoubaoDot = 'green' | 'red' | 'yellow';

export function doubaoDot(s: {
  running: boolean;
  cdpConnected: boolean;
  reconnecting?: boolean;
}): DoubaoDot {
  if (s.reconnecting) return 'yellow';
  return s.running && s.cdpConnected ? 'green' : 'red';
}

/** 状态栏"● 豆包"旁的短文案。 */
export function doubaoStatusText(s: DoubaoStatus | null): string {
  if (!s) return '状态未知';
  if (s.reconnecting) return '重连中';
  if (!s.running) return '豆包已退出,等待重启';
  return s.cdpConnected ? '已连接' : '未连接';
}

/** V88 配对状态:配置尚在读取或读取失败时，明确显示未知，不能误报为未配对。 */
export function v88PairingStatus(
  config: Pick<AppConfig, 'v88Executor'> | null,
  loading: boolean,
): { color: DoubaoDot; detail: string } {
  if (loading) return { color: 'yellow', detail: '正在检查配对状态' };
  if (!config) return { color: 'yellow', detail: '配对状态未读取' };
  return config.v88Executor.executorToken
    ? { color: 'green', detail: '已配对，等待任务' }
    : { color: 'red', detail: '未配对' };
}

/** 任务失败时优先返回执行器已持久化的中文页面原因，避免只显示笼统阶段名。 */
export function taskErrorHint(error: TaskErrorInfo | null | undefined): string | null {
  if (!error || !error.details || typeof error.details !== 'object') return null;
  const details = error.details as Record<string, unknown>;
  return typeof details.js_exception === 'string' && details.js_exception.trim()
    ? details.js_exception.trim()
    : null;
}

export type AccountRow = {
  name: string;
  subscription: string | null;
  isCurrent: boolean;
  isAmbiguous: boolean;
  /** 同名账号与当前账号都不可切换。 */
  canSwitch: boolean;
};

/** 账号表格的行,顺序与后端一致。 */
export function accountRows(list: AccountList | null): AccountRow[] {
  if (!list) return [];
  return list.accounts.map((a) => {
    const isAmbiguous = a.isAmbiguous;
    const isCurrent =
      a.isCurrent || (!isAmbiguous && list.current !== null && a.name === list.current);
    return {
      name: a.name,
      subscription: a.subscription ?? null,
      isCurrent,
      isAmbiguous,
      canSwitch: !isCurrent && !isAmbiguous,
    };
  });
}

export type EmptyState = 'not_connected' | 'not_logged_in' | 'empty' | null;

/**
 * 账号页空状态:
 * - `not_connected`:没有状态 / 豆包未运行 / CDP 未连(含重连中);
 * - `not_logged_in`:已连,列表为空且没有当前账号;
 * - `empty`:已连,但还没读到列表,或列表为空却有当前账号(异常,提示刷新);
 * - `null`:正常显示表格。
 */
export function emptyStateFor(status: DoubaoStatus | null, list: AccountList | null): EmptyState {
  if (!status || !status.running || !status.cdpConnected) return 'not_connected';
  if (!list) return 'empty';
  if (list.accounts.length === 0) return list.current ? 'empty' : 'not_logged_in';
  return null;
}

// ---------------- 任务(spec 06 页面 2/3 + 新建弹窗) ----------------

export type StatusTone = 'green' | 'blue' | 'red' | 'gray';

const STATUS_ICON: Record<TaskStatus, { icon: string; tone: StatusTone; label: string }> = {
  completed: { icon: '✓', tone: 'green', label: '完成' },
  preparing: { icon: '⏳', tone: 'blue', label: '准备中' },
  generating: { icon: '⏳', tone: 'blue', label: '生成中' },
  waiting_user_confirm: { icon: '⏳', tone: 'blue', label: '待确认' },
  failed: { icon: '⚠', tone: 'red', label: '失败' },
  pending: { icon: '⏸', tone: 'gray', label: '排队' },
  cancelled: { icon: '🚫', tone: 'gray', label: '已取消' },
};

/** 任务状态 → 图标 / 颜色 / 中文(spec 06:✓绿 / ⏳蓝 / ⚠红 / ⏸灰 / 🚫灰)。 */
export function statusIcon(status: TaskStatus): { icon: string; tone: StatusTone; label: string } {
  return STATUS_ICON[status] ?? { icon: '•', tone: 'gray', label: String(status) };
}

const pad2 = (n: number) => String(n).padStart(2, '0');

/** 本地时间 `HH:mm:ss`。 */
export function clockTime(ts: number): string {
  const d = new Date(ts);
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`;
}

/** 列表"时间"列:刚刚 / N 分前 / N 小时前 / N 天前 / 超过 7 天显示 `MM-DD HH:mm`。 */
export function relativeTime(ts: number, now: number): string {
  const diff = now - ts;
  if (diff < 60_000) return '刚刚';
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分前`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
  if (diff < 7 * 86_400_000) return `${Math.floor(diff / 86_400_000)} 天前`;
  const d = new Date(ts);
  return `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
}

export type TimelineTone = 'normal' | 'success' | 'error' | 'muted';

export type TimelineRow = {
  ts: number;
  time: string;
  stage: string;
  icon: string;
  label: string;
  message: string | null;
  tone: TimelineTone;
};

/** stage_history 里的名字既有阶段(sent/checking/…)也有状态(pending/generating/终态)。 */
const TIMELINE: Record<string, { icon: string; label: string; tone?: TimelineTone }> = {
  pending: { icon: '📥', label: '已入队' },
  preparing: { icon: '🔧', label: '准备中' },
  sent: { icon: '📤', label: '已发送到豆包' },
  checking: { icon: '❓', label: '豆包确认参数' },
  quota_checking: { icon: '💳', label: '豆包确认会员配额' },
  waiting_user_confirm: { icon: '✋', label: '等待手动确认' },
  generating: { icon: '🎬', label: '生成中' },
  rendering: { icon: '🎬', label: '豆包开始生成' },
  rendering_unverified: { icon: '🎬', label: '豆包生成中(未确认渲染)' },
  downloading: { icon: '✨', label: '生成完成,下载中' },
  result: { icon: '📦', label: '已拿到结果' },
  completed: { icon: '📦', label: '下载完成', tone: 'success' },
  refused: { icon: '⛔', label: '豆包拒绝生成', tone: 'error' },
  failed: { icon: '⚠', label: '失败', tone: 'error' },
  cancelled: { icon: '🚫', label: '已取消', tone: 'muted' },
};

/** 状态时间轴(按 ts 稳定排序)。 */
export function timelineRows(history: StageEntry[] | null | undefined): TimelineRow[] {
  if (!history) return [];
  return history
    .map((e, i) => ({ e, i }))
    .sort((a, b) => a.e.ts - b.e.ts || a.i - b.i)
    .map(({ e }) => {
      const def = TIMELINE[e.stage] ?? { icon: '•', label: e.stage };
      return {
        ts: e.ts,
        time: clockTime(e.ts),
        stage: e.stage,
        icon: def.icon,
        label: def.label,
        message: e.message ?? null,
        tone: def.tone ?? 'normal',
      };
    });
}

/** `task_created` / `task_completed`:新任务插最前,已有的原地替换(保留列表独有的 promptPreview)。 */
export function upsertTask(list: TaskSummary[], t: TaskSummary): TaskSummary[] {
  const i = list.findIndex((x) => x.taskId === t.taskId);
  if (i < 0) return [t, ...list];
  const next = list.slice();
  next[i] = { ...list[i], ...t, promptPreview: t.promptPreview ?? list[i].promptPreview };
  return next;
}

/** `task_updated`:只改对应行的 status / stage;不在列表里则原样返回(同一引用)。 */
export function applyTaskUpdate(list: TaskSummary[], ev: TaskUpdatedEvent): TaskSummary[] {
  const i = list.findIndex((x) => x.taskId === ev.taskId);
  if (i < 0) return list;
  const next = list.slice();
  next[i] = { ...list[i], status: ev.status, stage: ev.stage };
  return next;
}

/** `TaskDetail` → `TaskSummary`(task_completed 事件更新列表行用)。 */
export function summaryOf(t: TaskDetail): TaskSummary {
  return {
    taskId: t.taskId,
    account: t.account,
    type: t.type,
    status: t.status,
    stage: t.stage,
    createdAt: t.createdAt,
    completedAt: t.completedAt,
  };
}

const TERMINAL: TaskStatus[] = ['completed', 'failed', 'cancelled'];

export const isTerminal = (s: TaskStatus) => TERMINAL.includes(s);
/** 非终态都可取消(spec 08)。 */
export const canCancel = (s: TaskStatus) => !isTerminal(s);
/** 只有等待手动确认时显示"手动确认"。 */
export const canConfirm = (s: TaskStatus) => s === 'waiting_user_confirm';

/** 详情页"类型"行:`图像 (模型, 比例)` / `视频 (模型, N 秒, 比例)`。 */
export function taskTypeText(t: Pick<TaskDetail, 'type' | 'request'>): string {
  const r = t.request;
  if (t.type === 'video') {
    const dur = r.duration != null ? `${r.duration} 秒, ` : '';
    return `视频 (${r.modelName}, ${dur}${r.ratio})`;
  }
  return `图像 (${r.modelName}, ${r.ratio})`;
}

/** 列表 ID 列:ULID 前 10 位 + …。 */
export function shortId(id: string): string {
  return id.length > 10 ? `${id.slice(0, 10)}…` : id;
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

// 与后端 tasks/validate.rs 白名单一致(比例按真机菜单分类型,见 docs/analysis/10 Task 2.1)。
export const IMAGE_MODELS = ['Seedream 5.0 Flash', 'Seedream 5.0 Pro', 'Seedream 4.5', 'Seedream 4.0'];
export const VIDEO_MODELS = ['Seedance 2.5', 'Seedance 2.0', 'Seedance 2.0 Fast', 'Seedance 2.0 Mini'];
export const IMAGE_RATIOS = ['自动', '1:1', '16:9', '9:16', '4:3', '3:4', '3:2', '2:3'];
export const VIDEO_RATIOS = ['自动', '1:1', '16:9', '9:16', '4:3', '3:4', '21:9'];

export const modelsFor = (t: TaskType) => (t === 'image' ? IMAGE_MODELS : VIDEO_MODELS);
export const ratiosFor = (t: TaskType) => (t === 'image' ? IMAGE_RATIOS : VIDEO_RATIOS);

/** 视频时长范围(秒,含端点):Seedance 2.5 为 4–30,其余 4–15。 */
export function durationRange(model: string): [number, number] {
  return model === 'Seedance 2.5' ? [4, 30] : [4, 15];
}

/** 新建弹窗的字段名(camelCase)。 */
export type FormField =
  | 'account'
  | 'type'
  | 'modelName'
  | 'prompt'
  | 'ratio'
  | 'duration'
  | 'referenceFiles'
  | 'confirmationMode'
  | 'callbackUrl';

const SNAKE_TO_FIELD: Record<string, FormField> = {
  account: 'account',
  type: 'type',
  model_name: 'modelName',
  prompt: 'prompt',
  ratio: 'ratio',
  duration: 'duration',
  reference_files: 'referenceFiles',
  confirmation_mode: 'confirmationMode',
  callback_url: 'callbackUrl',
};

/**
 * 提交失败 → 表单内联错误(spec 06"字段错用 inline 红字");不属于某个字段的错误返回 null,由调用方弹 toast。
 */
export function fieldErrorOf(e: RelayError): { field: FormField; message: string } | null {
  const d = (e.details ?? {}) as { field?: unknown };
  const raw = typeof d.field === 'string' ? d.field.replace(/\[\d+\]$/, '') : null;
  let field: FormField | null = null;
  if (e.code === 'request_prompt_too_long') field = 'prompt';
  else if (e.code === 'request_too_many_references') field = 'referenceFiles';
  else if (e.code === 'request_invalid_callback_url') field = 'callbackUrl';
  else if (e.code.startsWith('request_') && raw) field = SNAKE_TO_FIELD[raw] ?? null;
  else if (e.code.startsWith('upload_')) field = 'referenceFiles';
  else if (e.code === 'account_not_found' || e.code === 'account_ambiguous') field = 'account';
  return field ? { field, message: e.message } : null;
}

// ---- Task 5.1:设置页 / 日志页 / 快捷键 ----

/** 设置页展示用:长 token 只露首尾各 4 位,短的全打码。 */
export function maskToken(token: string): string {
  if (!token) return '(未生成)';
  if (token.length <= 8) return '•'.repeat(token.length);
  return `${token.slice(0, 4)}…${token.slice(-4)}`;
}

function shellSingleQuote(s: string): string {
  return `'${s.replace(/'/g, "'\\''")}'`;
}

/** 监听地址 → URL 里的主机部分:IPv6 加方括号,通配地址换成对应的回环地址。 */
export function urlHost(host: string | undefined): string {
  const h = (host ?? '').trim();
  if (!h || h === '0.0.0.0') return '127.0.0.1';
  if (h === '::' || h === '[::]') return '[::1]';
  if (h.startsWith('[')) return h;
  return h.includes(':') ? `[${h}]` : h;
}

export type ApiExample = { title: string; note?: string; command: string };

/** OpenAI 兼容接口的全部调用示例(设置页"调用示例")。 */
export function apiExamples(port: number, token: string, host = '127.0.0.1'): ApiExample[] {
  const base = `http://${urlHost(host)}:${port}`;
  const auth = `-H ${shellSingleQuote(`Authorization: Bearer ${token}`)}`;
  const get = (path: string) => `curl ${auth} \\\n  ${base}${path}`;
  const post = (path: string, body: object) =>
    [
      `curl ${auth} \\`,
      `  -H 'Content-Type: application/json' \\`,
      `  -d ${shellSingleQuote(JSON.stringify(body))} \\`,
      `  ${base}${path}`,
    ].join('\n');
  return [
    {
      title: '账号列表',
      note: 'current 为豆包当前账号;加 ?refresh=1 强制重读',
      command: get('/v1/accounts'),
    },
    {
      title: '切换账号',
      note: '可选:生成请求带 account 时会自动切换;有任务在跑时返回 409',
      command: post('/v1/accounts/switch', { account: '张三' }),
    },
    { title: '健康检查', note: '免鉴权', command: `curl ${base}/health` },
    { title: '模型列表', command: get('/v1/models') },
    {
      title: '文生图(同步,出图后返回)',
      note: 'data[].url 为豆包 CDN 原图地址;不传 account 时用豆包当前账号',
      command: post('/v1/images/generations', {
        model: 'doubao-image-4.5',
        prompt: '一只蓝色小鸟站在樱花枝上',
        ratio: '1:1',
      }),
    },
    {
      title: '图生图(参考图)',
      note: 'ref_image 可传图片网址或 data:image/png;base64,...,多张用数组',
      command: post('/v1/images/generations', {
        model: 'doubao-image-5.0-flash',
        prompt: '改成水彩画风格',
        ref_image: 'https://example.com/ref.png',
        account: '张三',
      }),
    },
    {
      title: '生视频(异步,立即返回 id)',
      note: 'model:doubao-video-fast / doubao-video-mini;duration 默认 5 秒',
      command: post('/v1/video/generations', {
        model: 'doubao-video-fast',
        prompt: '一只柴犬在雪地里奔跑',
        ratio: '16:9',
        duration: 5,
      }),
    },
    {
      title: '查询视频结果',
      note: '把 <id> 换成上一步返回的 id;完成后 data[].video_url 为豆包 CDN 地址',
      command: get('/v1/video/generations/<id>'),
    },
  ];
}

/** 设置页展示的调用示例:token 未"显示"时示例里也只放打码值,避免绕过打码。 */
export function settingsApiExamples(
  server: { host: string; port: number; bearerToken: string },
  tokenVisible: boolean,
): ApiExample[] {
  const t = tokenVisible ? server.bearerToken : maskToken(server.bearerToken);
  return apiExamples(server.port, t, server.host);
}

export type SettingsNumbers = {
  port?: number;
  globalMax?: number;
  perAccountMax?: number;
  cdpPort?: number;
  resultRetentionHours?: number;
  metadataRetentionDays?: number;
};

/** 每账号并发的唯一合法值。 */
export const PER_ACCOUNT_MAX_FIXED = 1;

function isIntIn(v: number, min: number, max: number): boolean {
  return Number.isInteger(v) && v >= min && v <= max;
}

/** 设置页 inline 校验(与后端 `Config::validate` 同口径,前端先挡住明显错误);返回 字段 → 中文提示。 */
export function validateSettings(v: SettingsNumbers): Partial<Record<keyof SettingsNumbers, string>> {
  const errs: Partial<Record<keyof SettingsNumbers, string>> = {};
  if (v.port !== undefined && !isIntIn(v.port, 1, 65535)) errs.port = '端口必须是 1 到 65535 之间的整数';
  if (v.globalMax !== undefined && !isIntIn(v.globalMax, 1, 100))
    errs.globalMax = '全局最大并发必须是 1 到 100 之间的整数';
  // Global Constraints:每账号并发固定为 1(调度器按账号互斥,不读这个值),改了也不会生效
  if (v.perAccountMax !== undefined && v.perAccountMax !== PER_ACCOUNT_MAX_FIXED)
    errs.perAccountMax = '每账号并发固定为 1(同一账号同时只能有一个生成在豆包里进行)';
  if (v.cdpPort !== undefined && !isIntIn(v.cdpPort, 1, 65535))
    errs.cdpPort = 'CDP 端口必须是 1 到 65535 之间的整数';
  if (v.resultRetentionHours !== undefined && !isIntIn(v.resultRetentionHours, 1, 24 * 365))
    errs.resultRetentionHours = '结果文件保留时长至少 1 小时';
  if (v.metadataRetentionDays !== undefined && !isIntIn(v.metadataRetentionDays, 1, 3650))
    errs.metadataRetentionDays = '元数据保留天数至少 1 天';
  return errs;
}

/** `app.save_config` 成功后的 toast 文案。 */
export function saveResultMessage(needsRestart: boolean): string {
  return needsRestart ? '已保存，相关改动需要重启一战晟铭执行器才生效' : '已保存';
}

/** 复制 token 60s 后:剪贴板里仍是刚复制的 token 才清空(spec 12),不覆盖用户之后复制的别的东西。 */
export function shouldClearClipboard(current: string | null, copied: string): boolean {
  return current !== null && current !== '' && current === copied;
}

export type GcReportLike = { resultFilesRemoved: number; tasksRemoved: number; uploadsRemoved: number };

export function gcSummary(r: GcReportLike): string {
  return `清理完成:删除 ${r.resultFilesRemoved} 个结果文件、${r.tasksRemoved} 条任务记录、${r.uploadsRemoved} 条上传`;
}

/** 级别严重度:error=0 … trace=4;未知级别排最后(与后端 `level_rank` 一致)。 */
export function levelRank(level: string): number {
  switch (level.toLowerCase()) {
    case 'error':
      return 0;
    case 'warn':
    case 'warning':
      return 1;
    case 'info':
      return 2;
    case 'debug':
      return 3;
    case 'trace':
      return 4;
    default:
      return 99;
  }
}

/** 日志页级别过滤:`minLevel` 为 `null` 时原样返回(同一引用)。 */
export function filterByMinLevel<T extends { level: string }>(entries: T[], minLevel: string | null): T[] {
  if (minLevel === null) return entries;
  const max = levelRank(minLevel);
  return entries.filter((e) => levelRank(e.level) <= max);
}

/** 实时追加并限制条数(与后端内存缓冲 2000 对齐)。 */
const rowIds = new WeakMap<object, number>();
let nextRowId = 0;

/** 列表行的稳定 key:按对象身份分配自增 id(封顶裁掉旧行后,留下的行 key 不变,Preact 不用整表重建)。 */
export function rowKey(row: object): number {
  let id = rowIds.get(row);
  if (id === undefined) {
    id = ++nextRowId;
    rowIds.set(row, id);
  }
  return id;
}

export function appendCapped<T>(list: T[], entry: T, cap: number): T[] {
  const next = list.length >= cap ? list.slice(list.length - cap + 1) : list.slice();
  next.push(entry);
  return next;
}

const pad = (n: number, w: number) => String(n).padStart(w, '0');

/** 日志行时间:本地 `HH:mm:ss.SSS`。 */
export function formatLogTime(ts: number): string {
  const d = new Date(ts);
  return `${pad(d.getHours(), 2)}:${pad(d.getMinutes(), 2)}:${pad(d.getSeconds(), 2)}.${pad(d.getMilliseconds(), 3)}`;
}

export type ShortcutAction = 'new_task' | 'refresh_accounts' | 'settings' | 'logs' | 'search_tasks';

/** spec 06 快捷键表:macOS 用 Cmd,其它平台用 Ctrl;带 Alt/Shift 的组合不算。 */
export function shortcutFor(
  e: Pick<KeyboardEvent, 'key' | 'metaKey' | 'ctrlKey' | 'altKey' | 'shiftKey'>,
  isMac: boolean,
): ShortcutAction | null {
  const mod = isMac ? e.metaKey && !e.ctrlKey : e.ctrlKey && !e.metaKey;
  if (!mod || e.altKey || e.shiftKey) return null;
  switch (e.key.toLowerCase()) {
    case 'n':
      return 'new_task';
    case 'r':
      return 'refresh_accounts';
    case ',':
      return 'settings';
    case 'l':
      return 'logs';
    case 'k':
      return 'search_tasks';
    default:
      return null;
  }
}

/** 状态栏"● API"的文字:监听中带地址,没读到 / 未监听为"未启动"。 */
export function apiStatusText(s: { enabled?: boolean; listening: boolean; host: string; port: number; error?: string | null } | null): string {
  if (!s) return '未启动';
  if (s.enabled === false) return '已关闭';
  if (s.error) return `启动失败：${s.error}`;
  if (!s.listening) return '未启动';
  const host = s.host.includes(':') ? `[${s.host}]` : s.host;
  return `监听中 ${host}:${s.port}`;
}

/** 状态栏底部的版本信息:relay 版本 + 检测到的豆包版本(客户排查时区分新旧版)。 */
export function versionLines(relayVersion: string | null, doubaoVersion: string | null): string[] {
  const lines: string[] = [];
  if (relayVersion) lines.push(`一战晟铭执行器 v${relayVersion}`);
  if (doubaoVersion) lines.push(`豆包 ${doubaoVersion}`);
  return lines;
}
