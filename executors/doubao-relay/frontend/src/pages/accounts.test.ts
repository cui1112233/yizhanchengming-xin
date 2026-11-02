// 账号页数据逻辑:读取失败的分类(静默 vs 必须可见)、并发 force 刷新不丢。
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { DoubaoStatus, RelayError } from '../api';
import {
  accountList,
  accountsError,
  accountsLoading,
  doubaoStatus,
  toasts,
} from '../state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
vi.mock('@tauri-apps/api/core', () => ({ invoke: (cmd: string, args?: unknown) => invoke(cmd, args) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));

const { describeAccountsError, isSilentAccountsError, loadAccounts } = await import('./accounts');

function status(patch: Partial<DoubaoStatus> = {}): DoubaoStatus {
  return {
    running: true,
    cdpConnected: true,
    cdpPort: 9707,
    doubaoVersionDetected: '2.32.5',
    uiCompatCheck: 'ok',
    uiCompatDetails: {
      status: 'ok',
      accountButtonFound: true,
      inputFound: true,
      sendButtonFound: true,
      checkedAt: 1,
    },
    chatTargets: 1,
    reconnecting: false,
    lastCheckAt: 1,
    ...patch,
  };
}

const err = (code: string, message = `${code} 消息`, details?: unknown): RelayError =>
  details === undefined ? { code, message } : { code, message, details };

const LIST = { current: '她是翻哥', accounts: [], readAt: 1 };

beforeEach(() => {
  invoke.mockReset();
  accountList.value = null;
  accountsError.value = null;
  accountsLoading.value = false;
  toasts.value = [];
  doubaoStatus.value = status();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('isSilentAccountsError', () => {
  it('非 doubao_* 错误永远不静默', () => {
    expect(isSilentAccountsError(err('account_not_found'))).toBe(false);
    expect(isSilentAccountsError(err('system_internal_error'), null)).toBe(false);
  });

  it('断线类错误码静默(状态栏已表达)', () => {
    expect(isSilentAccountsError(err('doubao_cdp_disconnected'))).toBe(true);
    expect(isSilentAccountsError(err('doubao_not_running'))).toBe(true);
  });

  it('状态栏已显示未连接时其它 doubao_* 也静默', () => {
    expect(isSilentAccountsError(err('doubao_not_ready'), status({ cdpConnected: false }))).toBe(true);
    expect(isSilentAccountsError(err('doubao_ui_changed'), null)).toBe(true);
  });

  it('已连接时 doubao_ui_changed / doubao_not_ready 不静默', () => {
    expect(isSilentAccountsError(err('doubao_ui_changed'))).toBe(false);
    expect(isSilentAccountsError(err('doubao_not_ready'))).toBe(false);
  });
});

describe('describeAccountsError', () => {
  it('带 failed_selectors 时附在 message 后', () => {
    const e = err('doubao_ui_changed', '豆包界面结构发生变化', {
      failed_selectors: ['button[aria-haspopup=menu]', '[role=menu]'],
    });
    expect(describeAccountsError(e)).toBe(
      '豆包界面结构发生变化(未命中:button[aria-haspopup=menu]、[role=menu])',
    );
  });

  it('没有 details / 空数组 / 非字符串 时只返回 message', () => {
    expect(describeAccountsError(err('doubao_not_ready', '尚未就绪'))).toBe('尚未就绪');
    expect(describeAccountsError(err('x', 'm', { failed_selectors: [] }))).toBe('m');
    expect(describeAccountsError(err('x', 'm', { failed_selectors: 'nope' }))).toBe('m');
  });
});

describe('loadAccounts', () => {
  it('成功:写入列表、清掉上次错误', async () => {
    accountsError.value = '旧错误';
    invoke.mockResolvedValueOnce(LIST);
    await loadAccounts(false);
    expect(invoke).toHaveBeenCalledWith('account_list', undefined);
    expect(accountList.value).toEqual(LIST);
    expect(accountsError.value).toBeNull();
    expect(toasts.value).toEqual([]);
  });

  it('force 走 account_refresh', async () => {
    invoke.mockResolvedValueOnce(LIST);
    await loadAccounts(true);
    expect(invoke).toHaveBeenCalledWith('account_refresh', undefined);
  });

  it('doubao_not_logged_in → 空列表(未登录空状态),不弹 toast', async () => {
    invoke.mockRejectedValueOnce(err('doubao_not_logged_in'));
    await loadAccounts(false);
    expect(accountList.value).toMatchObject({ current: null, accounts: [] });
    expect(accountsError.value).toBeNull();
    expect(toasts.value).toEqual([]);
  });

  it('doubao_cdp_disconnected → 列表置空、不弹 toast(状态栏已表达)', async () => {
    accountList.value = LIST;
    invoke.mockRejectedValueOnce(err('doubao_cdp_disconnected'));
    await loadAccounts(false);
    expect(accountList.value).toBeNull();
    expect(accountsError.value).toBeNull();
    expect(toasts.value).toEqual([]);
  });

  it('已连接时 doubao_ui_changed → 错误 toast(含 message 与未命中 selector)+ 空状态原因', async () => {
    invoke.mockRejectedValueOnce(
      err('doubao_ui_changed', '豆包界面结构发生变化,请升级 doubao-relay 以适配新版', {
        failed_selectors: ['button[aria-haspopup="menu"]'],
      }),
    );
    await loadAccounts(false);
    expect(accountList.value).toBeNull();
    expect(accountsError.value).toBe(
      '豆包界面结构发生变化,请升级 doubao-relay 以适配新版(未命中:button[aria-haspopup="menu"])',
    );
    expect(toasts.value).toHaveLength(1);
    expect(toasts.value[0].kind).toBe('error');
    expect(toasts.value[0].message).toBe(`读取账号失败:${accountsError.value}`);
  });

  it('已连接时 doubao_not_ready → 错误 toast', async () => {
    invoke.mockRejectedValueOnce(err('doubao_not_ready', '豆包尚未就绪,请等待豆包连接完成后重试'));
    await loadAccounts(true);
    expect(toasts.value.map((t) => t.message)).toEqual([
      '读取账号失败:豆包尚未就绪,请等待豆包连接完成后重试',
    ]);
  });

  it('未连接时 doubao_not_ready → 静默(空状态已是"未连接")', async () => {
    doubaoStatus.value = status({ cdpConnected: false });
    invoke.mockRejectedValueOnce(err('doubao_not_ready'));
    await loadAccounts(false);
    expect(accountList.value).toBeNull();
    expect(toasts.value).toEqual([]);
  });

  it('非 doubao_* 错误 → toast,保留旧列表', async () => {
    accountList.value = LIST;
    invoke.mockRejectedValueOnce(err('system_internal_error', '内部错误'));
    await loadAccounts(false);
    expect(accountList.value).toEqual(LIST);
    expect(toasts.value.map((t) => t.message)).toEqual(['读取账号失败:内部错误']);
  });

  it('读取进行中收到 force 请求:当前读完后再强刷一次,不丢', async () => {
    let resolveFirst!: (v: unknown) => void;
    invoke.mockImplementationOnce(() => new Promise((r) => (resolveFirst = r)));
    const refreshed = { ...LIST, readAt: 2 };
    invoke.mockResolvedValueOnce(refreshed);

    const first = loadAccounts(false);
    await loadAccounts(true); // 读取中,立即返回
    expect(invoke).toHaveBeenCalledTimes(1);

    resolveFirst({ ...LIST, readAt: 1 });
    await first;
    // 待办 force 在 finally 里异步发出
    await vi.waitFor(() => expect(invoke).toHaveBeenCalledTimes(2));
    expect(invoke.mock.calls[1][0]).toBe('account_refresh');
    await vi.waitFor(() => expect(accountList.value).toEqual(refreshed));
    expect(accountsLoading.value).toBe(false);
  });

  it('读取进行中的非 force 请求直接丢弃,不触发第二次读取', async () => {
    let resolveFirst!: (v: unknown) => void;
    invoke.mockImplementationOnce(() => new Promise((r) => (resolveFirst = r)));
    const first = loadAccounts(false);
    await loadAccounts(false);
    resolveFirst(LIST);
    await first;
    await new Promise((r) => setTimeout(r, 0));
    expect(invoke).toHaveBeenCalledTimes(1);
  });
});
