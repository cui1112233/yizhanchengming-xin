import type { ApiStatus } from './api';
// 全局 signals(单例)。页面与组件只读/写这些 signal,不自己持有后端状态。
import { signal } from '@preact/signals';
import type { AccountList, AppConfig, DoubaoStatus, TaskDetail, TaskStatus, TaskSummary } from './api';

export type Route = 'accounts' | 'tasks' | 'logs' | 'settings';

export type ToastKind = 'info' | 'success' | 'error';
export type Toast = { id: number; kind: ToastKind; message: string };

/** 最近一次 `doubao.status` / `doubao_status_changed`;`null` 表示还没拿到。 */
export const doubaoStatus = signal<DoubaoStatus | null>(null);
/** 最近一次账号列表;`null` 表示还没读到(或豆包未连)。 */
export const accountList = signal<AccountList | null>(null);
/** 账号列表读取中(首屏骨架屏 / 刷新按钮 loading)。 */
export const accountsLoading = signal(false);
/** 最近一次读取账号列表的失败原因(已连接却读不到,如豆包改版);`null` 表示没有。 */
export const accountsError = signal<string | null>(null);
/** 正在切换到的账号名;`null` 表示没有切换进行中。 */
export const switching = signal<string | null>(null);
/** 重连豆包进行中。 */
export const reconnecting = signal(false);
export const toasts = signal<Toast[]>([]);
export const route = signal<Route>('accounts');
/** HTTP API 监听状态(`app_api_status`,启动时读一次后定时刷新);`null` = 还没读到。 */
export const apiStatus = signal<ApiStatus | null>(null);
/** relay 自身版本(`app_get_version`,启动时读一次)。 */
export const appVersion = signal<string | null>(null);

// ---- 任务页 / 详情页 / 新建弹窗 ----

/** 任务列表(已加载的若干页);`null` 表示还没读过。 */
export const tasks = signal<TaskSummary[] | null>(null);
/** 下一页游标;`null` 表示没有更多。 */
export const tasksNextCursor = signal<string | null>(null);
export const tasksLoading = signal(false);
/** 列表读取失败原因(空状态里显示);`null` 表示没有。 */
export const tasksError = signal<string | null>(null);
/** 过滤:状态(空串 = 全部)+ 账号(空串 = 全部)。 */
export const tasksFilter = signal<{ status: TaskStatus | ''; account: string }>({
  status: '',
  account: '',
});
/** 任务页里打开的任务详情 id(spec 06 `/tasks/:id`);`null` 表示显示列表。 */
export const selectedTaskId = signal<string | null>(null);
/** 当前详情(与 selectedTaskId 对应);`null` 表示加载中或没有。 */
export const taskDetail = signal<TaskDetail | null>(null);
export const taskDetailError = signal<string | null>(null);
/** 详情页上正在进行的操作(取消 / 确认 / 打开目录),按钮转圈防重入。 */
export const taskAction = signal<'cancel' | 'confirm' | 'open' | null>(null);
/** 新建任务弹窗是否打开。 */
export const newTaskOpen = signal(false);

/** 普通提示 3s 自动消失;错误常驻直到点关。 */
export const TOAST_AUTO_DISMISS_MS = 3000;

let nextToastId = 1;

export function pushToast(kind: ToastKind, message: string): number {
  const id = nextToastId++;
  toasts.value = [...toasts.value, { id, kind, message }];
  if (kind !== 'error') {
    setTimeout(() => dismissToast(id), TOAST_AUTO_DISMISS_MS);
  }
  return id;
}

export function dismissToast(id: number) {
  if (toasts.value.some((t) => t.id === id)) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }
}

// ---- Task 5.1 ----
/** 最近一次 `app.get_config` / 保存后的配置;`null` 表示还没读到。 */
export const appConfig = signal<AppConfig | null>(null);
/** 启动时正在读取执行器配置;未完成前状态栏不能把未知状态误报成未配对。 */
export const appConfigLoading = signal(true);
/** CDP 调试日志开关(跟随 `appConfig.logging.cdpDebugLog`,日志页据此显示"CDP 调试"面板)。 */
export const cdpDebugEnabled = signal(false);

// ---- Task 5.2 ----
/** 隐私声明是否已同意;`null` = 还没读到(此时先显示声明遮罩,读到已同意再撤掉)。 */
export const privacyAccepted = signal<boolean | null>(null);
