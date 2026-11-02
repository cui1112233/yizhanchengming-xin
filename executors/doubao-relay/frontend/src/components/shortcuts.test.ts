// @vitest-environment happy-dom
// 快捷键:Cmd/Ctrl+, / L / R 直接处理并拦默认行为;N / K 只广播 relay:shortcut 给任务页。
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { accountsLoading, doubaoStatus, route } from '../state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
vi.mock('@tauri-apps/api/core', () => ({ invoke: (cmd: string, args?: unknown) => invoke(cmd, args) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));
vi.mock('@tauri-apps/plugin-clipboard-manager', () => ({ writeText: vi.fn(), readText: vi.fn() }));
vi.mock('@tauri-apps/plugin-dialog', () => ({ save: vi.fn() }));

const { SHORTCUT_EVENT, handleShortcutKeydown } = await import('./shortcuts');

function keydown(key: string, mods: Partial<KeyboardEvent> = {}): KeyboardEvent {
  const e = new KeyboardEvent('keydown', { key, cancelable: true, ...mods });
  return e;
}

beforeEach(() => {
  invoke.mockReset();
  invoke.mockResolvedValue({ current: null, accounts: [], readAt: 1 });
  route.value = 'accounts';
  doubaoStatus.value = null;
  accountsLoading.value = false;
});

describe('handleShortcutKeydown', () => {
  it('Cmd+, → 设置页并 preventDefault', () => {
    const e = keydown(',', { metaKey: true });
    handleShortcutKeydown(e, true);
    expect(route.value).toBe('settings');
    expect(e.defaultPrevented).toBe(true);
  });

  it('Ctrl+L(非 mac)→ 日志页;Cmd+L 在非 mac 不触发', () => {
    handleShortcutKeydown(keydown('l', { ctrlKey: true }), false);
    expect(route.value).toBe('logs');
    route.value = 'accounts';
    handleShortcutKeydown(keydown('l', { metaKey: true }), false);
    expect(route.value).toBe('accounts');
  });

  it('Cmd+R → 回账号页;豆包已连时强制刷新账号', async () => {
    route.value = 'logs';
    handleShortcutKeydown(keydown('r', { metaKey: true }), true);
    expect(route.value).toBe('accounts');
    expect(invoke).not.toHaveBeenCalled();
    doubaoStatus.value = { cdpConnected: true, running: true } as never;
    handleShortcutKeydown(keydown('r', { metaKey: true }), true);
    await Promise.resolve();
    expect(invoke).toHaveBeenCalledWith('account_refresh', undefined);
  });

  it('Cmd+N / Cmd+K 广播 relay:shortcut,供任务页接管', () => {
    const got: string[] = [];
    const on = (ev: Event) => got.push((ev as CustomEvent<string>).detail);
    window.addEventListener(SHORTCUT_EVENT, on);
    handleShortcutKeydown(keydown('n', { metaKey: true }), true);
    handleShortcutKeydown(keydown('k', { metaKey: true }), true);
    handleShortcutKeydown(keydown('x', { metaKey: true }), true);
    window.removeEventListener(SHORTCUT_EVENT, on);
    expect(got).toEqual(['new_task', 'search_tasks']);
    expect(route.value).toBe('accounts');
  });

  it('普通按键不处理也不拦', () => {
    const e = keydown('l');
    handleShortcutKeydown(e, true);
    expect(e.defaultPrevented).toBe(false);
    expect(route.value).toBe('accounts');
  });
});
