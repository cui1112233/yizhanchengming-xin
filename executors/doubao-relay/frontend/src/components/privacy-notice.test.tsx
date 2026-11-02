// @vitest-environment happy-dom
// 隐私声明:未同意时全屏显示;必须勾选后才能"同意并继续";同意写后端后撤掉;"不同意并退出"走 app_quit。
import { render } from 'preact';
import { act } from 'preact/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const invoke = vi.fn();
vi.mock('@tauri-apps/api/core', () => ({ invoke: (...a: unknown[]) => invoke(...a) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));

const { PrivacyNotice, canAccept, loadPrivacy, resetPrivacyNotice, PROMISES, WARNINGS } =
  await import('./privacy-notice');
const { privacyAccepted } = await import('../state');

let root: HTMLDivElement;

const flush = () => act(async () => {});
const buttons = () => Array.from(root.querySelectorAll('button')) as HTMLButtonElement[];
const byText = (t: string) => buttons().find((b) => b.textContent === t)!;
const checkbox = () => root.querySelector('input[type=checkbox]') as HTMLInputElement;

beforeEach(() => {
  invoke.mockReset();
  resetPrivacyNotice();
  privacyAccepted.value = null;
  root = document.createElement('div');
  document.body.appendChild(root);
});

afterEach(() => {
  render(null, root);
  root.remove();
});

describe('canAccept', () => {
  it('只有勾选且不忙时可点', () => {
    expect(canAccept(false, false)).toBe(false);
    expect(canAccept(true, true)).toBe(false);
    expect(canAccept(true, false)).toBe(true);
  });
});

describe('loadPrivacy', () => {
  it('读后端状态', async () => {
    invoke.mockResolvedValueOnce({ accepted: true, acceptedAt: 1 });
    await loadPrivacy();
    expect(invoke).toHaveBeenCalledWith('app_privacy_status');
    expect(privacyAccepted.value).toBe(true);
  });

  it('读取失败按未同意处理', async () => {
    invoke.mockRejectedValueOnce({ code: 'system_database_error', message: 'x' });
    await loadPrivacy();
    expect(privacyAccepted.value).toBe(false);
  });
});

describe('PrivacyNotice', () => {
  it('已同意时不渲染;未读到时只渲染遮罩', async () => {
    privacyAccepted.value = true;
    await act(() => render(<PrivacyNotice />, root));
    expect(root.innerHTML).toBe('');
    privacyAccepted.value = null;
    await flush();
    expect(root.querySelector('.privacy-backdrop')).toBeTruthy();
    expect(root.querySelector('[role=dialog]')).toBeNull();
  });

  it('显示 spec 12 全部承诺与警告,勾选前不能同意', async () => {
    privacyAccepted.value = false;
    await act(() => render(<PrivacyNotice />, root));
    const text = root.textContent ?? '';
    for (const t of [...PROMISES, ...WARNINGS]) expect(text).toContain(t);
    expect(byText('同意并继续').disabled).toBe(true);
    expect(document.activeElement).toBe(checkbox());
  });

  it('勾选 → 同意并继续 → 写后端并撤掉声明', async () => {
    privacyAccepted.value = false;
    invoke.mockResolvedValueOnce({ accepted: true, acceptedAt: 5 });
    await act(() => render(<PrivacyNotice />, root));
    await act(() => {
      checkbox().click();
    });
    expect(byText('同意并继续').disabled).toBe(false);
    await act(() => {
      byText('同意并继续').click();
    });
    await flush();
    expect(invoke).toHaveBeenCalledWith('app_privacy_accept');
    expect(privacyAccepted.value).toBe(true);
    expect(root.innerHTML).toBe('');
  });

  it('同意失败时显示原因并保留声明', async () => {
    privacyAccepted.value = false;
    invoke.mockRejectedValueOnce({ code: 'system_database_error', message: '数据库写入失败' });
    await act(() => render(<PrivacyNotice />, root));
    await act(() => {
      checkbox().click();
    });
    await act(() => {
      byText('同意并继续').click();
    });
    await flush();
    expect(privacyAccepted.value).toBe(false);
    expect(root.querySelector('[role=alert]')?.textContent).toContain('数据库写入失败');
  });

  it('不同意并退出 → app_quit', async () => {
    privacyAccepted.value = false;
    invoke.mockResolvedValueOnce({ ok: true });
    await act(() => render(<PrivacyNotice />, root));
    await act(() => {
      byText('不同意并退出').click();
    });
    await flush();
    expect(invoke).toHaveBeenCalledWith('app_quit');
    expect(privacyAccepted.value).toBe(false);
  });
});
