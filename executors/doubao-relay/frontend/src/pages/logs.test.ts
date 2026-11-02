// 日志页数据逻辑:拉取 / 实时追加封顶 / 字段紧凑展示 / 导出(取消对话框不提示,成功 toast 带条数)。
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { LogEntry } from '../api';
import { cdpDebugEnabled, toasts } from '../state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
const save = vi.fn<() => Promise<string | null>>();
vi.mock('@tauri-apps/api/core', () => ({ invoke: (cmd: string, args?: unknown) => invoke(cmd, args) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));
vi.mock('@tauri-apps/plugin-clipboard-manager', () => ({ writeText: vi.fn(), readText: vi.fn() }));
vi.mock('@tauri-apps/plugin-dialog', () => ({ save: () => save() }));

const { LOG_CAP, appendLogEntry, exportLogs, formatFields, loadLogs, logEntries, logMinLevel, syncCdpEnabled } =
  await import('./logs');

const entry = (level: string, ts: number, fields: Record<string, unknown> = {}): LogEntry => ({
  ts,
  level,
  target: 't',
  message: `m${ts}`,
  fields,
});

beforeEach(() => {
  invoke.mockReset();
  save.mockReset();
  logEntries.value = [];
  logMinLevel.value = null;
  toasts.value = [];
});

describe('loadLogs / appendLogEntry', () => {
  it('log_recent 拉满 2000 条上限', async () => {
    invoke.mockResolvedValueOnce({ entries: [entry('info', 1)] });
    await loadLogs();
    expect(invoke).toHaveBeenCalledWith('log_recent', { limit: LOG_CAP, level: undefined });
    expect(logEntries.value).toHaveLength(1);
  });

  it('实时追加超过上限丢最旧', () => {
    logEntries.value = Array.from({ length: LOG_CAP }, (_, i) => entry('info', i));
    appendLogEntry(entry('warn', 9999));
    expect(logEntries.value).toHaveLength(LOG_CAP);
    expect(logEntries.value[0].ts).toBe(1);
    expect(logEntries.value.at(-1)?.ts).toBe(9999);
  });

  it('读取失败 → 错误 toast', async () => {
    invoke.mockRejectedValueOnce({ code: 'system_internal_error', message: '坏了' });
    await loadLogs();
    expect(toasts.value[0]).toMatchObject({ kind: 'error' });
    expect(toasts.value[0].message).toContain('坏了');
  });
});

describe('syncCdpEnabled', () => {
  it('从 debug_cdp_log 的 enabled 同步信号;失败保持原值', async () => {
    cdpDebugEnabled.value = false;
    invoke.mockResolvedValueOnce({ entries: [], enabled: true });
    await syncCdpEnabled();
    expect(invoke).toHaveBeenCalledWith('debug_cdp_log', { limit: 1 });
    expect(cdpDebugEnabled.value).toBe(true);
    invoke.mockRejectedValueOnce(new Error('x'));
    await syncCdpEnabled();
    expect(cdpDebugEnabled.value).toBe(true);
  });
});

describe('formatFields', () => {
  it('字符串原样,其它 JSON 化,空对象为空串', () => {
    expect(formatFields({})).toBe('');
    expect(formatFields({ task_id: '01A', count: 3, ok: true, d: { a: 1 } })).toBe(
      'task_id=01A count=3 ok=true d={"a":1}',
    );
  });
});

describe('exportLogs', () => {
  it('取消保存对话框 → 不调后端、不提示', async () => {
    save.mockResolvedValueOnce(null);
    await exportLogs();
    expect(invoke).not.toHaveBeenCalled();
    expect(toasts.value).toEqual([]);
  });

  it('选了路径 → log_export(path, 当前级别) → success toast 带条数', async () => {
    save.mockResolvedValueOnce('/tmp/x.ndjson');
    invoke.mockResolvedValueOnce({ ok: true, path: '/tmp/x.ndjson', count: 7 });
    logMinLevel.value = 'warn';
    await exportLogs();
    expect(invoke).toHaveBeenCalledWith('log_export', { path: '/tmp/x.ndjson', level: 'warn' });
    expect(toasts.value[0]).toMatchObject({ kind: 'success' });
    expect(toasts.value[0].message).toContain('7 条');
  });

  it('后端写文件失败 → 错误 toast', async () => {
    save.mockResolvedValueOnce('/nope/x.ndjson');
    invoke.mockRejectedValueOnce({ code: 'system_io_error', message: '写入日志导出文件失败' });
    await exportLogs();
    expect(toasts.value[0].kind).toBe('error');
  });
});

describe('CDP 面板自动滚动', () => {
  it('只在接近底部时跟随新条目', async () => {
    const { isNearBottom } = await import('../components/cdp-debug-panel');
    expect(isNearBottom({ scrollHeight: 1000, scrollTop: 600, clientHeight: 400 })).toBe(true);
    expect(isNearBottom({ scrollHeight: 1000, scrollTop: 580, clientHeight: 400 })).toBe(true);
    expect(isNearBottom({ scrollHeight: 1000, scrollTop: 200, clientHeight: 400 })).toBe(false);
  });
});
