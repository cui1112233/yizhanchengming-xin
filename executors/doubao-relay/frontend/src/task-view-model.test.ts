// 任务页 / 详情页 / 新建弹窗的纯函数视图模型(view-model.ts 中任务相关部分)。
import { describe, expect, it } from 'vitest';
import type { RelayError, TaskDetail, TaskSummary, TaskUpdatedEvent } from './api';
import {
  applyTaskUpdate,
  canCancel,
  canConfirm,
  clockTime,
  durationRange,
  fieldErrorOf,
  formatBytes,
  modelsFor,
  ratiosFor,
  relativeTime,
  shortId,
  statusIcon,
  taskErrorHint,
  taskTypeText,
  timelineRows,
  upsertTask,
  v88PairingStatus,
} from './view-model';

describe('statusIcon', () => {
  it('按 spec 06:✓绿 / ⏳蓝 / ⚠红 / ⏸灰 / 🚫灰', () => {
    expect(statusIcon('completed')).toEqual({ icon: '✓', tone: 'green', label: '完成' });
    expect(statusIcon('generating')).toEqual({ icon: '⏳', tone: 'blue', label: '生成中' });
    expect(statusIcon('preparing')).toEqual({ icon: '⏳', tone: 'blue', label: '准备中' });
    expect(statusIcon('waiting_user_confirm')).toEqual({ icon: '⏳', tone: 'blue', label: '待确认' });
    expect(statusIcon('failed')).toEqual({ icon: '⚠', tone: 'red', label: '失败' });
    expect(statusIcon('pending')).toEqual({ icon: '⏸', tone: 'gray', label: '排队' });
    expect(statusIcon('cancelled')).toEqual({ icon: '🚫', tone: 'gray', label: '已取消' });
  });

  it('未知状态不崩,原样显示', () => {
    expect(statusIcon('weird' as never)).toEqual({ icon: '•', tone: 'gray', label: 'weird' });
  });
});

describe('relativeTime', () => {
  const now = 1_700_000_000_000;
  it('一分钟内 → 刚刚(含时钟偏差导致的未来时间)', () => {
    expect(relativeTime(now, now)).toBe('刚刚');
    expect(relativeTime(now - 59_000, now)).toBe('刚刚');
    expect(relativeTime(now + 5_000, now)).toBe('刚刚');
  });
  it('分钟 / 小时 / 天', () => {
    expect(relativeTime(now - 60_000, now)).toBe('1 分前');
    expect(relativeTime(now - 2 * 60_000 - 30_000, now)).toBe('2 分前');
    expect(relativeTime(now - 59 * 60_000, now)).toBe('59 分前');
    expect(relativeTime(now - 3_600_000, now)).toBe('1 小时前');
    expect(relativeTime(now - 23 * 3_600_000, now)).toBe('23 小时前');
    expect(relativeTime(now - 86_400_000, now)).toBe('1 天前');
    expect(relativeTime(now - 6 * 86_400_000, now)).toBe('6 天前');
  });
  it('超过 7 天 → 月-日 时:分', () => {
    const ts = new Date(2026, 0, 5, 9, 7).getTime();
    expect(relativeTime(ts, ts + 8 * 86_400_000)).toBe('01-05 09:07');
  });
});

describe('clockTime', () => {
  it('本地时间 HH:mm:ss 补零', () => {
    expect(clockTime(new Date(2026, 9, 16, 10, 3, 5).getTime())).toBe('10:03:05');
  });
});

describe('timelineRows', () => {
  const t = (h: number, m: number, s: number) => new Date(2026, 9, 16, h, m, s).getTime();
  it('按时间排序,阶段名 → 图标 + 中文,附带消息', () => {
    const rows = timelineRows([
      { stage: 'sent', ts: t(10, 30, 15) },
      { stage: 'pending', ts: t(10, 30, 0) },
      { stage: 'preparing', ts: t(10, 30, 5), message: null },
      { stage: 'quota_checking', ts: t(10, 30, 25), message: '豆包确认是否使用会员配额' },
      { stage: 'downloading', ts: t(10, 32, 40) },
      { stage: 'completed', ts: t(10, 32, 50) },
    ]);
    expect(rows.map((r) => r.time)).toEqual([
      '10:30:00',
      '10:30:05',
      '10:30:15',
      '10:30:25',
      '10:32:40',
      '10:32:50',
    ]);
    expect(rows[0]).toMatchObject({ icon: '📥', label: '已入队', message: null, tone: 'normal' });
    expect(rows[1]).toMatchObject({ icon: '🔧', label: '准备中' });
    expect(rows[2]).toMatchObject({ icon: '📤', label: '已发送到豆包' });
    expect(rows[3]).toMatchObject({
      icon: '💳',
      label: '豆包确认会员配额',
      message: '豆包确认是否使用会员配额',
    });
    expect(rows[4]).toMatchObject({ icon: '✨', label: '生成完成,下载中' });
    expect(rows[5]).toMatchObject({ icon: '📦', label: '下载完成', tone: 'success' });
  });

  it('失败 / 取消 / 拒绝标红或置灰;未知阶段原样显示;空历史 → 空数组', () => {
    const rows = timelineRows([
      { stage: 'refused', ts: 1, message: '豆包拒绝生成' },
      { stage: 'failed', ts: 2, message: '豆包拒绝生成' },
      { stage: 'cancelled', ts: 3 },
      { stage: 'brand_new_stage', ts: 4 },
    ]);
    expect(rows.map((r) => r.tone)).toEqual(['error', 'error', 'muted', 'normal']);
    expect(rows[3]).toMatchObject({ icon: '•', label: 'brand_new_stage' });
    expect(timelineRows([])).toEqual([]);
    expect(timelineRows(undefined)).toEqual([]);
  });

  it('同一时间戳保持原顺序(稳定排序)', () => {
    const rows = timelineRows([
      { stage: 'checking', ts: 5 },
      { stage: 'waiting_user_confirm', ts: 5 },
      { stage: 'generating', ts: 5 },
    ]);
    expect(rows.map((r) => r.stage)).toEqual(['checking', 'waiting_user_confirm', 'generating']);
  });
});

const summary = (patch: Partial<TaskSummary> = {}): TaskSummary => ({
  taskId: '01A',
  account: '张三',
  type: 'image',
  status: 'pending',
  stage: null,
  createdAt: 1,
  completedAt: null,
  ...patch,
});

describe('upsertTask / applyTaskUpdate', () => {
  it('新任务插到最前,已有任务原地替换且保留 promptPreview', () => {
    const list = [summary({ taskId: '01A', promptPreview: '小鸟' })];
    const added = upsertTask(list, summary({ taskId: '01B', createdAt: 2 }));
    expect(added.map((t) => t.taskId)).toEqual(['01B', '01A']);
    const replaced = upsertTask(added, summary({ taskId: '01A', status: 'completed' }));
    expect(replaced.map((t) => t.taskId)).toEqual(['01B', '01A']);
    expect(replaced[1].status).toBe('completed');
    expect(replaced[1].promptPreview).toBe('小鸟');
  });

  it('task_updated 只改对应行的 status / stage;不认识的 id 原样返回', () => {
    const list = [summary({ taskId: '01A' }), summary({ taskId: '01B' })];
    const ev: TaskUpdatedEvent = {
      taskId: '01B',
      status: 'generating',
      stage: 'checking',
      message: 'm',
      progress: null,
      updatedAt: 9,
    };
    const next = applyTaskUpdate(list, ev);
    expect(next[1]).toMatchObject({ status: 'generating', stage: 'checking' });
    expect(next[0]).toBe(list[0]);
    expect(applyTaskUpdate(list, { ...ev, taskId: 'nope' })).toBe(list);
  });
});

describe('canCancel / canConfirm', () => {
  it('非终态可取消;仅 waiting_user_confirm 可手动确认', () => {
    for (const s of ['pending', 'preparing', 'generating', 'waiting_user_confirm'] as const) {
      expect(canCancel(s)).toBe(true);
    }
    for (const s of ['completed', 'failed', 'cancelled'] as const) expect(canCancel(s)).toBe(false);
    expect(canConfirm('waiting_user_confirm')).toBe(true);
    expect(canConfirm('generating')).toBe(false);
  });
});

describe('taskTypeText / shortId / formatBytes', () => {
  const req = (patch: Partial<TaskDetail['request']>): Pick<TaskDetail, 'type' | 'request'> => ({
    type: patch.type ?? 'image',
    request: {
      account: '张三',
      type: 'image',
      modelName: 'Seedream 5.0 Pro',
      prompt: 'p',
      ratio: '1:1',
      ...patch,
    },
  });
  it('图像 / 视频描述', () => {
    expect(taskTypeText(req({}))).toBe('图像 (Seedream 5.0 Pro, 1:1)');
    expect(
      taskTypeText(req({ type: 'video', modelName: 'Seedance 2.5', duration: 10, ratio: '16:9' })),
    ).toBe('视频 (Seedance 2.5, 10 秒, 16:9)');
  });
  it('shortId 取前 10 位加省略号', () => {
    expect(shortId('01HQZ9ABCDEFGHJKMNPQRSTVWX')).toBe('01HQZ9ABCD…');
    expect(shortId('01A')).toBe('01A');
  });
  it('formatBytes', () => {
    expect(formatBytes(512)).toBe('512 B');
    expect(formatBytes(2048)).toBe('2.0 KB');
    expect(formatBytes(24.5 * 1024 * 1024)).toBe('24.5 MB');
  });
});

describe('新建弹窗选项', () => {
  it('模型 / 比例按类型(真机菜单:图像无 21:9,视频无 3:2 / 2:3)', () => {
    expect(modelsFor('image')).toEqual(['Seedream 5.0 Flash', 'Seedream 5.0 Pro', 'Seedream 4.5', 'Seedream 4.0']);
    expect(modelsFor('video')).toContain('Seedance 2.5');
    expect(ratiosFor('image')).not.toContain('21:9');
    expect(ratiosFor('video')).toContain('21:9');
    expect(ratiosFor('video')).not.toContain('3:2');
  });
  it('时长范围:Seedance 2.5 为 4–30,其余 4–15', () => {
    expect(durationRange('Seedance 2.5')).toEqual([4, 30]);
    expect(durationRange('Seedance 2.0 Mini')).toEqual([4, 15]);
  });
});

describe('fieldErrorOf', () => {
  const e = (code: string, details?: unknown): RelayError => ({ code, message: `${code} 消息`, details });
  it('request_* 带 field → 映射到表单字段名', () => {
    expect(fieldErrorOf(e('request_missing_field', { field: 'prompt' }))).toEqual({
      field: 'prompt',
      message: 'request_missing_field 消息',
    });
    expect(fieldErrorOf(e('request_invalid_enum', { field: 'model_name' }))?.field).toBe('modelName');
    expect(fieldErrorOf(e('request_too_many_references', {}))?.field).toBe('referenceFiles');
    expect(fieldErrorOf(e('request_invalid_range', { field: 'duration' }))?.field).toBe('duration');
  });
  it('账号错误归到 account 字段;其它错误 → null(弹 toast)', () => {
    expect(fieldErrorOf(e('account_not_found', { account: 'x' }))?.field).toBe('account');
    expect(fieldErrorOf(e('queue_full'))).toBeNull();
    expect(fieldErrorOf(e('request_malformed', { reason: 'x' }))).toBeNull();
  });
});

describe('执行器状态与任务失败说明', () => {
  it('配对配置仍在读取时不能误报为未配对', () => {
    expect(v88PairingStatus(null, true)).toEqual({ color: 'yellow', detail: '正在检查配对状态' });
    expect(v88PairingStatus(null, false)).toEqual({ color: 'yellow', detail: '配对状态未读取' });
    expect(
      v88PairingStatus(
        {
          v88Executor: { baseUrl: 'http://localhost:4001', executorId: 'lex_1', executorToken: 'token' },
        },
        false,
      ),
    ).toEqual({ color: 'green', detail: '已配对，等待任务' });
  });

  it('准备阶段失败时展示执行器记录的页面原因，而不是只给笼统失败文案', () => {
    expect(
      taskErrorHint({
        code: 'generation_preparing_failed',
        message: '生成准备阶段失败,请检查豆包当前页面状态后重试',
        details: { js_exception: '未找到所选生成比例。', reason: 'prepare' },
      }),
    ).toBe('未找到所选生成比例。');
    expect(
      taskErrorHint({
        code: 'generation_preparing_failed',
        message: '生成准备阶段失败',
        details: { js_exception: '豆包控件未唯一匹配，已停止发送。' },
      }),
    ).toBe('豆包控件未唯一匹配，已停止发送。');
  });
});
