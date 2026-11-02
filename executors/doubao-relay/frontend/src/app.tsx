import { useEffect } from 'preact/hooks';
import * as api from './api';
import { asRelayError } from './api';
import type {
  AccountList,
  AccountSwitchedEvent,
  DoubaoStatus,
  TaskDetail,
  TaskSummary,
  TaskUpdatedEvent,
} from './api';
import { ConfirmDialog } from './components/confirm-dialog';
import { NewTaskDialog } from './components/new-task-dialog';
import { Sidebar } from './components/sidebar';
import { ToastHost } from './components/toast';
import { AccountsPage } from './pages/accounts';
import { applyCompletedDetail, refreshTaskDetail } from './pages/task-detail';
import { matchesFilter, TasksPage } from './pages/tasks';
import {
  accountList,
  accountsError,
  apiStatus,
  appVersion,
  appConfig,
  appConfigLoading,
  doubaoStatus,
  pushToast,
  route,
  selectedTaskId,
  switching,
  tasks,
} from './state';
import { Shortcuts } from './components/shortcuts'; // Task 5.1
import { LogsPage } from './pages/logs'; // Task 5.1
import { SettingsPage } from './pages/settings'; // Task 5.1
import { loadPrivacy, PrivacyNotice } from './components/privacy-notice'; // Task 5.2
import { applyTaskUpdate, shortId, summaryOf, taskErrorHint, upsertTask } from './view-model';

/** API 监听状态没有事件推送(服务比窗口晚启动、也可能中途停掉),定时读。 */
const API_STATUS_POLL_MS = 5000;

async function refreshApiStatus() {
  try {
    apiStatus.value = await api.appApiStatus();
  } catch {
    apiStatus.value = null;
  }
}

/** 启动:拉一次状态;之后全靠事件推送(API 状态除外,见 refreshApiStatus)。 */
/** 窗口启动时并行读取运行状态与配对配置，导出供回归测试覆盖账号页的首屏状态。 */
export async function bootstrap() {
  void refreshApiStatus();
  const configPromise = api
    .appGetConfig()
    .then((config) => {
      appConfig.value = config;
    })
    .catch((raw) => {
      pushToast('error', `读取执行器配置失败:${asRelayError(raw).message}`);
    })
    .finally(() => {
      appConfigLoading.value = false;
    });
  api
    .appGetVersion()
    .then((v) => (appVersion.value = v.version))
    .catch(() => {});
  try {
    doubaoStatus.value = await api.doubaoStatus();
  } catch (raw) {
    pushToast('error', `读取豆包状态失败:${asRelayError(raw).message}`);
  }
  await configPromise;
}

/** `doubao_status_changed` 的处理(导出便于单测)。 */
export function applyDoubaoStatus(s: DoubaoStatus) {
  const prev = doubaoStatus.value;
  doubaoStatus.value = s;
  if (prev && prev.running && !s.running) {
    pushToast('error', '豆包已退出,等待重启。点击“重连豆包”可由 relay 重新启动');
  }
  // 断线后列表已不可信,清掉让账号页显示"未连接",重新连上时自动重读。
  // 例外:账号切换期间豆包页面 reload 会让心跳短暂看到断线,此时保留列表,让"切换中…"行一直显示到
  // account_switched / 切换失败为止(UC-04);豆包进程真退出了则照常清空。
  if (!s.cdpConnected && (switching.value === null || !s.running)) accountList.value = null;
}

/** `task_created`:符合当前过滤条件的新任务插到列表最前(列表还没读过则不动,进页面时会整页读)。 */
export function applyTaskCreated(t: TaskSummary) {
  if (tasks.value !== null && matchesFilter(t)) tasks.value = upsertTask(tasks.value, t);
}

/** `task_updated`:更新列表行;正在看的详情重读一次(时间轴 live)。 */
export function applyTaskUpdated(ev: TaskUpdatedEvent) {
  if (tasks.value !== null) tasks.value = applyTaskUpdate(tasks.value, ev);
  if (selectedTaskId.value === ev.taskId) void refreshTaskDetail(ev.taskId);
}

/** `task_completed`(含 failed / cancelled):更新列表行与详情;完成 / 失败弹 toast。 */
export function applyTaskCompleted(t: TaskDetail) {
  if (tasks.value !== null) {
    const exists = tasks.value.some((x) => x.taskId === t.taskId);
    if (exists || matchesFilter(t)) tasks.value = upsertTask(tasks.value, summaryOf(t));
  }
  applyCompletedDetail(t);
  if (t.status === 'completed') pushToast('success', `任务 ${shortId(t.taskId)} 已完成`);
  else if (t.status === 'failed')
    pushToast('error', `任务 ${shortId(t.taskId)} 失败:${taskErrorHint(t.error) ?? t.error?.message ?? '未知原因'}`);
}

function subscribeEvents(): () => void {
  const offs = [
    api.onEvent<DoubaoStatus>('doubao_status_changed', applyDoubaoStatus),
    api.onEvent<AccountList>('account_list_changed', (list) => {
      accountList.value = list;
      accountsError.value = null;
    }),
    api.onEvent<AccountSwitchedEvent>('account_switched', (e) => {
      switching.value = null;
      pushToast('success', `已切换到账号“${e.current}”`);
    }),
    api.onEvent<TaskSummary>('task_created', applyTaskCreated),
    api.onEvent<TaskUpdatedEvent>('task_updated', applyTaskUpdated),
    api.onEvent<TaskDetail>('task_completed', applyTaskCompleted),
    api.onEvent<{ at: number }>('doubao_crashed', () => {
      pushToast('error', '豆包已崩溃,进行中的任务已失败;点击“重连豆包”恢复');
    }),
  ];
  return () => offs.forEach((off) => off());
}

export function App() {
  useEffect(() => {
    const off = subscribeEvents();
    void loadPrivacy();
    void bootstrap();
    const timer = setInterval(() => void refreshApiStatus(), API_STATUS_POLL_MS);
    return () => {
      clearInterval(timer);
      off();
    };
  }, []);

  const page = route.value;
  return (
    <div class="layout">
      <Sidebar />
      <main class="main">
        {page === 'accounts' ? (
          <AccountsPage />
        ) : page === 'tasks' ? (
          <TasksPage />
        ) : page === 'logs' ? (
          <LogsPage />
        ) : (
          <SettingsPage />
        )}
      </main>
      <NewTaskDialog />
      <ToastHost />
      <ConfirmDialog />
      <Shortcuts />
      <PrivacyNotice />
    </div>
  );
}
