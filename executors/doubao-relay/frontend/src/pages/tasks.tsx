// 任务页(docs/spec/06-webview-ui.md 页面 2):列表 + 过滤 + 加载更多;点行进详情(页面 3)。
// 实时更新由 app.tsx 订阅 task_created / task_updated / task_completed 改 `tasks` signal。
import { useEffect, useState } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import type { TaskListFilter, TaskStatus } from '../api';
import {
  accountList,
  newTaskOpen,
  pushToast,
  selectedTaskId,
  taskDetail,
  tasks,
  tasksError,
  tasksFilter,
  tasksLoading,
  tasksNextCursor,
} from '../state';
import { relativeTime, shortId, statusIcon } from '../view-model';
import { TaskDetailPage } from './task-detail';

/** 每页条数(spec 05 默认 20)。 */
export const PAGE_SIZE = 20;

const STATUS_OPTIONS: { value: TaskStatus | ''; label: string }[] = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '排队' },
  { value: 'preparing', label: '准备中' },
  { value: 'generating', label: '生成中' },
  { value: 'waiting_user_confirm', label: '待确认' },
  { value: 'completed', label: '完成' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' },
];

/** 当前过滤条件 → `task.list` 参数。 */
export function currentFilter(cursor?: string): TaskListFilter {
  const f = tasksFilter.value;
  const out: TaskListFilter = { limit: PAGE_SIZE };
  if (f.status) out.status = [f.status];
  if (f.account.trim()) out.account = f.account.trim();
  if (cursor) out.cursor = cursor;
  return out;
}

/** 新任务是否落在当前过滤条件里(task_created 时决定要不要插进列表)。 */
export function matchesFilter(t: { status: TaskStatus; account: string }): boolean {
  const f = tasksFilter.value;
  if (f.status && f.status !== t.status) return false;
  if (f.account.trim() && f.account.trim() !== t.account) return false;
  return true;
}

let loadSeq = 0;

/** 读第一页(`append=false`,替换列表)或下一页(`append=true`)。过滤改变时旧请求的结果丢弃。 */
export async function loadTasks(append = false) {
  if (append && (tasksLoading.value || !tasksNextCursor.value)) return;
  const seq = ++loadSeq;
  tasksLoading.value = true;
  try {
    const r = await api.taskList(currentFilter(append ? tasksNextCursor.value! : undefined));
    if (seq !== loadSeq) return;
    if (append && tasks.value) {
      const seen = new Set(tasks.value.map((t) => t.taskId));
      tasks.value = [...tasks.value, ...r.tasks.filter((t) => !seen.has(t.taskId))];
    } else {
      tasks.value = r.tasks;
    }
    tasksNextCursor.value = r.nextCursor;
    tasksError.value = null;
  } catch (raw) {
    if (seq !== loadSeq) return;
    const e = asRelayError(raw);
    if (!append) {
      tasks.value = null;
      tasksError.value = e.message;
    }
    pushToast('error', `读取任务列表失败:${e.message}`);
  } finally {
    if (seq === loadSeq) tasksLoading.value = false;
  }
}

/** 进入详情。 */
export function openTask(id: string) {
  if (selectedTaskId.value !== id) taskDetail.value = null;
  selectedTaskId.value = id;
}

function setFilter(patch: Partial<{ status: TaskStatus | ''; account: string }>) {
  tasksFilter.value = { ...tasksFilter.value, ...patch };
  tasksNextCursor.value = null;
  void loadTasks(false);
}

/** 每 30s 刷新一次"N 分前"。 */
function useNow(intervalMs = 30_000): number {
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return now;
}

function Skeleton() {
  return (
    <div class="skeleton" aria-busy="true" aria-label="加载中">
      <div class="skeleton-line w-40" />
      <div class="skeleton-line" />
      <div class="skeleton-line" />
      <div class="skeleton-line w-70" />
    </div>
  );
}

function EmptyState() {
  const err = tasksError.value;
  const filtered = tasksFilter.value.status !== '' || tasksFilter.value.account.trim() !== '';
  if (err) {
    return (
      <div class="empty-state" data-kind="error" data-error>
        <p class="empty-title">读取任务列表失败</p>
        <p class="empty-hint">{err}</p>
        <button type="button" class="btn" onClick={() => loadTasks(false)}>
          重试
        </button>
      </div>
    );
  }
  if (filtered) {
    return (
      <div class="empty-state" data-kind="filtered">
        <p class="empty-title">没有符合条件的任务</p>
        <p class="empty-hint">换个过滤条件,或清除过滤查看全部任务</p>
        <button type="button" class="btn" onClick={() => setFilter({ status: '', account: '' })}>
          清除过滤
        </button>
      </div>
    );
  }
  return (
    <div class="empty-state" data-kind="empty">
      <p class="empty-title">还没有任务</p>
      <p class="empty-hint">通过 HTTP API 提交,或点击下方按钮新建一个生图 / 生视频任务</p>
      <button type="button" class="btn btn-primary" onClick={() => (newTaskOpen.value = true)}>
        新建任务
      </button>
    </div>
  );
}

function TaskList() {
  const list = tasks.value;
  const loading = tasksLoading.value;
  const filter = tasksFilter.value;
  const now = useNow();
  const accounts = Array.from(
    new Set([
      ...(accountList.value?.accounts.map((a) => a.name) ?? []),
      ...(list?.map((t) => t.account) ?? []),
      ...(filter.account ? [filter.account] : []),
    ]),
  );

  useEffect(() => {
    if (tasks.value === null && !tasksLoading.value) void loadTasks(false);
  }, []);

  return (
    <section class="page page-tasks">
      <header class="page-header">
        <h1 class="page-title">任务</h1>
        <div class="page-actions">
          <select
            class="select"
            aria-label="按状态过滤"
            value={filter.status}
            onChange={(e) => setFilter({ status: (e.currentTarget as HTMLSelectElement).value as TaskStatus | '' })}
          >
            {STATUS_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <select
            class="select"
            aria-label="按账号过滤"
            value={filter.account}
            onChange={(e) => setFilter({ account: (e.currentTarget as HTMLSelectElement).value })}
          >
            <option value="">全部账号</option>
            {accounts.map((a) => (
              <option key={a} value={a}>
                {a}
              </option>
            ))}
          </select>
          <button type="button" class="btn" disabled={loading} onClick={() => loadTasks(false)}>
            {loading && list !== null ? '读取中…' : '刷新'}
          </button>
          <button type="button" class="btn btn-primary" onClick={() => (newTaskOpen.value = true)}>
            新建
          </button>
        </div>
      </header>

      {list === null && loading ? (
        <Skeleton />
      ) : list === null || list.length === 0 ? (
        <EmptyState />
      ) : (
        <>
          <table class="table tasks-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>账号</th>
                <th>类型</th>
                <th>状态</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              {list.map((t) => {
                const s = statusIcon(t.status);
                return (
                  <tr
                    key={t.taskId}
                    class="row-link"
                    tabIndex={0}
                    title={t.promptPreview ?? ''}
                    data-task-id={t.taskId}
                    onClick={() => openTask(t.taskId)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') openTask(t.taskId);
                    }}
                  >
                    <td class="mono">
                      {shortId(t.taskId)} {t.type === 'video' ? '🎬' : '📷'}
                    </td>
                    <td>{t.account}</td>
                    <td>{t.type === 'video' ? '视频' : '图像'}</td>
                    <td>
                      <span class={`task-status tone-${s.tone}`}>
                        {s.icon}
                        {s.label}
                      </span>
                    </td>
                    <td class="muted">{relativeTime(t.createdAt, now)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <p class="muted table-footer">
            已显示 {list.length} 个任务。
            {tasksNextCursor.value && (
              <button type="button" class="btn btn-small" disabled={loading} onClick={() => loadTasks(true)}>
                {loading ? '加载中…' : '加载更多'}
              </button>
            )}
          </p>
        </>
      )}
    </section>
  );
}

export function TasksPage() {
  return selectedTaskId.value ? <TaskDetailPage id={selectedTaskId.value} /> : <TaskList />;
}
