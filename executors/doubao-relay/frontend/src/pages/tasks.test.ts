// 任务页 / 详情页 / 新建弹窗的数据逻辑 + app.tsx 任务事件处理。
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { TaskDetail, TaskSummary } from '../api';
import {
  newTaskOpen,
  selectedTaskId,
  taskAction,
  taskDetail,
  taskDetailError,
  tasks,
  tasksError,
  tasksFilter,
  tasksLoading,
  tasksNextCursor,
  toasts,
} from '../state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
vi.mock('@tauri-apps/api/core', () => ({
  invoke: (cmd: string, args?: unknown) => invoke(cmd, args),
  convertFileSrc: (p: string) => `asset://localhost/${encodeURIComponent(p)}`,
}));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));
vi.mock('@tauri-apps/plugin-dialog', () => ({ open: vi.fn(() => Promise.resolve(null)) }));

const { currentFilter, loadTasks, matchesFilter, openTask } = await import('./tasks');
const { refreshTaskDetail, confirmTask, closeTaskDetail } = await import('./task-detail');
const { applyTaskCompleted, applyTaskCreated, applyTaskUpdated } = await import('../app');
const { buildRequest, defaultForm, withType } = await import('../components/new-task-dialog');

const summary = (id: string, patch: Partial<TaskSummary> = {}): TaskSummary => ({
  taskId: id,
  account: '张三',
  type: 'image',
  status: 'pending',
  stage: null,
  createdAt: 1,
  completedAt: null,
  ...patch,
});

function detail(id: string, patch: Partial<TaskDetail> = {}): TaskDetail {
  return {
    taskId: id,
    account: '张三',
    type: 'image',
    request: { account: '张三', type: 'image', modelName: 'Seedream 5.0 Pro', prompt: 'p', ratio: '1:1' },
    status: 'completed',
    stage: null,
    stageMessage: null,
    message: null,
    progress: null,
    error: null,
    conversationId: null,
    submissionId: null,
    watermarkSource: 'conversation_original',
    createdAt: 1,
    startedAt: 2,
    completedAt: 3,
    expiresAt: 4,
    stageHistory: [],
    queuePosition: null,
    ...patch,
  };
}

beforeEach(() => {
  invoke.mockReset();
  tasks.value = null;
  tasksNextCursor.value = null;
  tasksLoading.value = false;
  tasksError.value = null;
  tasksFilter.value = { status: '', account: '' };
  selectedTaskId.value = null;
  taskDetail.value = null;
  taskDetailError.value = null;
  taskAction.value = null;
  newTaskOpen.value = false;
  toasts.value = [];
});

describe('任务列表', () => {
  it('过滤条件 → task.list 参数', () => {
    expect(currentFilter()).toEqual({ limit: 20 });
    tasksFilter.value = { status: 'failed', account: ' 张三 ' };
    expect(currentFilter('01C')).toEqual({ limit: 20, status: ['failed'], account: '张三', cursor: '01C' });
    expect(matchesFilter({ status: 'failed', account: '张三' })).toBe(true);
    expect(matchesFilter({ status: 'pending', account: '张三' })).toBe(false);
    expect(matchesFilter({ status: 'failed', account: '李四' })).toBe(false);
  });

  it('首页 + 加载更多(按游标追加、去重),没有游标时不再请求', async () => {
    invoke.mockResolvedValueOnce({ tasks: [summary('01B'), summary('01A')], nextCursor: '01A' });
    await loadTasks(false);
    expect(invoke).toHaveBeenLastCalledWith('task_list', { filter: { limit: 20 } });
    expect(tasks.value?.map((t) => t.taskId)).toEqual(['01B', '01A']);
    expect(tasksNextCursor.value).toBe('01A');

    invoke.mockResolvedValueOnce({ tasks: [summary('01A'), summary('009')], nextCursor: null });
    await loadTasks(true);
    expect(invoke).toHaveBeenLastCalledWith('task_list', { filter: { limit: 20, cursor: '01A' } });
    expect(tasks.value?.map((t) => t.taskId)).toEqual(['01B', '01A', '009']);
    expect(tasksNextCursor.value).toBeNull();

    invoke.mockClear();
    await loadTasks(true);
    expect(invoke).not.toHaveBeenCalled();
  });

  it('读取失败:列表置空、空状态显示原因、弹错误 toast', async () => {
    invoke.mockRejectedValueOnce({ code: 'system_database_error', message: '数据库读写失败' });
    await loadTasks(false);
    expect(tasks.value).toBeNull();
    expect(tasksError.value).toBe('数据库读写失败');
    expect(toasts.value[0]?.kind).toBe('error');
    expect(tasksLoading.value).toBe(false);
  });

  it('过滤切换时较早请求的结果被丢弃', async () => {
    let resolveFirst: (v: unknown) => void = () => {};
    invoke.mockImplementationOnce(() => new Promise((r) => (resolveFirst = r)));
    const first = loadTasks(false);
    invoke.mockResolvedValueOnce({ tasks: [summary('NEW')], nextCursor: null });
    await loadTasks(false);
    resolveFirst({ tasks: [summary('OLD')], nextCursor: 'x' });
    await first;
    expect(tasks.value?.map((t) => t.taskId)).toEqual(['NEW']);
    expect(tasksNextCursor.value).toBeNull();
  });
});

describe('任务事件(app.tsx)', () => {
  it('task_created 插入符合过滤条件的新任务;列表未读过时不动', () => {
    applyTaskCreated(summary('01A'));
    expect(tasks.value).toBeNull();
    tasks.value = [];
    applyTaskCreated(summary('01A'));
    expect(tasks.value?.map((t) => t.taskId)).toEqual(['01A']);
    tasksFilter.value = { status: 'completed', account: '' };
    applyTaskCreated(summary('01B'));
    expect(tasks.value?.map((t) => t.taskId)).toEqual(['01A']);
  });

  it('task_updated 改列表行,并重读正在看的详情', async () => {
    tasks.value = [summary('01A')];
    selectedTaskId.value = '01A';
    invoke.mockResolvedValueOnce(detail('01A', { status: 'generating' }));
    applyTaskUpdated({
      taskId: '01A',
      status: 'generating',
      stage: 'sent',
      message: null,
      progress: null,
      updatedAt: 5,
    });
    expect(tasks.value[0]).toMatchObject({ status: 'generating', stage: 'sent' });
    expect(invoke).toHaveBeenCalledWith('task_get', { taskId: '01A' });
    await vi.waitFor(() => expect(taskDetail.value?.status).toBe('generating'));
  });

  it('task_completed 更新行与详情,完成 / 失败弹 toast', () => {
    tasks.value = [summary('01A', { promptPreview: '小鸟' })];
    selectedTaskId.value = '01A';
    applyTaskCompleted(detail('01A'));
    expect(tasks.value[0]).toMatchObject({ status: 'completed', promptPreview: '小鸟' });
    expect(taskDetail.value?.status).toBe('completed');
    expect(toasts.value.at(-1)?.kind).toBe('success');

    applyTaskCompleted(
      detail('01B', {
        status: 'failed',
        error: { code: 'generation_doubao_refused', message: '豆包拒绝生成' },
      }),
    );
    expect(tasks.value.map((t) => t.taskId)).toEqual(['01B', '01A']);
    expect(taskDetail.value?.taskId).toBe('01A');
    expect(toasts.value.at(-1)).toMatchObject({ kind: 'error' });
    expect(toasts.value.at(-1)?.message).toContain('豆包拒绝生成');
  });
});

describe('任务详情', () => {
  it('openTask 切换任务时清掉旧详情;切走后迟到的结果被丢弃', async () => {
    taskDetail.value = detail('01A');
    openTask('01B');
    expect(taskDetail.value).toBeNull();
    let resolve: (v: unknown) => void = () => {};
    invoke.mockImplementationOnce(() => new Promise((r) => (resolve = r)));
    const p = refreshTaskDetail('01B');
    closeTaskDetail();
    resolve(detail('01B'));
    await p;
    expect(taskDetail.value).toBeNull();
    expect(selectedTaskId.value).toBeNull();
  });

  it('task_not_found 只在页内显示,不弹 toast', async () => {
    selectedTaskId.value = 'nope';
    invoke.mockRejectedValueOnce({ code: 'task_not_found', message: '任务不存在' });
    await refreshTaskDetail('nope');
    expect(taskDetailError.value).toBe('任务不存在');
    expect(toasts.value).toEqual([]);
  });

  it('手动确认调用 task_confirm 后重读详情', async () => {
    selectedTaskId.value = '01A';
    invoke.mockImplementation(async (cmd) =>
      cmd === 'task_confirm' ? { ok: true, newStage: 'generating' } : detail('01A', { status: 'generating' }),
    );
    await confirmTask('01A');
    expect(invoke).toHaveBeenCalledWith('task_confirm', { taskId: '01A' });
    expect(taskAction.value).toBeNull();
    await vi.waitFor(() => expect(taskDetail.value?.status).toBe('generating'));
  });
});

describe('新建任务表单', () => {
  it('默认图像 + 当前账号;切到视频时比例 / 模型回到合法值', () => {
    const f = defaultForm('张三');
    expect(f).toMatchObject({ account: '张三', type: 'image', modelName: 'Seedream 5.0 Pro', ratio: '1:1' });
    const v = withType({ ...f, ratio: '3:2' }, 'video');
    expect(v.modelName).toBe('Seedance 2.5');
    expect(v.ratio).toBe('1:1');
    expect(withType({ ...v, ratio: '21:9' }, 'image').ratio).toBe('1:1');
    expect(withType(f, 'image')).toBe(f);
  });

  it('buildRequest:视频带 duration,图像不带;参考图转 uploadId 列表', () => {
    const f = { ...defaultForm(' 张三 '), prompt: ' 一只小鸟 ' };
    expect(buildRequest(f)).toEqual({
      account: '张三',
      type: 'image',
      modelName: 'Seedream 5.0 Pro',
      prompt: ' 一只小鸟 ',
      ratio: '1:1',
      confirmationMode: 'auto',
      autoUseMembership: true,
    });
    const v = withType({ ...f, duration: 10 }, 'video');
    expect(buildRequest(v)).toMatchObject({ type: 'video', duration: 10, modelName: 'Seedance 2.5' });
    const r = buildRequest({
      ...f,
      references: [{ uploadId: 'U1', mime: 'image/png', bytes: 1, expiresAt: 1, name: 'a.png' }],
    });
    expect(r.referenceFiles).toEqual(['U1']);
  });
});
