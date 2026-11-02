// @vitest-environment happy-dom
// 确认框键盘行为:Enter 落在"取消"按钮上不能变成"确定";焦点要被挪进对话框,触发源按钮不会被再次激活。
import { render } from 'preact';
import { act } from 'preact/test-utils';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { ConfirmDialog, confirm, keyAction } from './confirm-dialog';

let root: HTMLDivElement;

/** 在 act 里派发 keydown,返回事件对象以便断言 defaultPrevented。 */
const press = async (target: EventTarget, key: string) => {
  const ev = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true });
  await act(() => {
    target.dispatchEvent(ev);
  });
  return ev;
};

const open = async () => {
  let p!: Promise<boolean>;
  await act(() => {
    p = confirm({ title: '切换账号', message: '确定?', confirmLabel: '切换' });
  });
  const buttons = Array.from(root.querySelectorAll('.dialog-actions button'));
  const cancel = buttons.find((b) => b.textContent === '取消') as HTMLButtonElement;
  const ok = buttons.find((b) => b.textContent === '切换') as HTMLButtonElement;
  expect(cancel).toBeTruthy();
  expect(ok).toBeTruthy();
  return { p, cancel, ok };
};

/** 已 resolve 的 Promise 立刻返回其值;仍 pending 返回 'pending'。 */
const settled = (p: Promise<boolean>) =>
  Promise.race([p.then((v) => v), new Promise<'pending'>((r) => setTimeout(() => r('pending'), 0))]);

beforeEach(() => {
  root = document.createElement('div');
  document.body.appendChild(root);
  render(<ConfirmDialog />, root);
});

afterEach(async () => {
  // 把可能残留的对话框按 Escape 关掉,再卸载
  await press(window, 'Escape');
  render(null, root);
  root.remove();
});

describe('keyAction', () => {
  const dialog = document.createElement('div');
  const inner = document.createElement('button');
  dialog.appendChild(inner);
  const outside = document.createElement('button');

  it('Escape → cancel;其它键不处理', () => {
    expect(keyAction('Escape', inner, dialog)).toBe('cancel');
    expect(keyAction('a', document.body, dialog)).toBeNull();
    expect(keyAction('Tab', inner, dialog)).toBeNull();
  });

  it('Enter 落在对话框内的按钮上 → 不处理(交给按钮 click)', () => {
    expect(keyAction('Enter', inner, dialog)).toBeNull();
  });

  it('Enter 落在对话框外 / 非按钮 → confirm', () => {
    expect(keyAction('Enter', outside, dialog)).toBe('confirm');
    expect(keyAction('Enter', document.body, dialog)).toBe('confirm');
    expect(keyAction('Enter', null, dialog)).toBe('confirm');
    expect(keyAction('Enter', inner, null)).toBe('confirm');
  });
});

describe('ConfirmDialog', () => {
  it('打开后焦点在确认按钮上(不依赖被 Chromium 拒绝的 autofocus)', async () => {
    const { ok } = await open();
    expect(document.activeElement).toBe(ok);
  });

  it('焦点在"取消"上按 Enter 不会当成"确定";随后 click 取消 → false', async () => {
    const { p, cancel } = await open();
    cancel.focus();
    const ev = await press(cancel, 'Enter');
    expect(ev.defaultPrevented).toBe(false);
    expect(await settled(p)).toBe('pending');
    expect(root.querySelector('.dialog')).not.toBeNull();
    await act(() => {
      cancel.click();
    });
    expect(await settled(p)).toBe(false);
    expect(root.querySelector('.dialog')).toBeNull();
  });

  it('焦点在"确定"上按 Enter 同样交给按钮 click', async () => {
    const { p, ok } = await open();
    await press(ok, 'Enter');
    expect(await settled(p)).toBe('pending');
    await act(() => {
      ok.click();
    });
    expect(await settled(p)).toBe(true);
  });

  it('焦点仍在对话框外的触发按钮上时按 Enter → 确定,且 preventDefault 阻止该按钮再被激活', async () => {
    const trigger = document.createElement('button');
    trigger.textContent = '切换';
    document.body.appendChild(trigger);
    try {
      const { p } = await open();
      trigger.focus();
      const ev = await press(trigger, 'Enter');
      expect(ev.defaultPrevented).toBe(true);
      expect(await settled(p)).toBe(true);
    } finally {
      trigger.remove();
    }
  });

  it('Escape → false;点击遮罩 → false', async () => {
    const a = await open();
    await press(window, 'Escape');
    expect(await settled(a.p)).toBe(false);

    const b = await open();
    await act(() => {
      (root.querySelector('.dialog-backdrop') as HTMLElement).click();
    });
    expect(await settled(b.p)).toBe(false);
  });

  it('再次 confirm 会把前一个对话框按"取消"关掉', async () => {
    const a = await open();
    const b = await open();
    expect(await settled(a.p)).toBe(false);
    expect(await settled(b.p)).toBe('pending');
  });
});
