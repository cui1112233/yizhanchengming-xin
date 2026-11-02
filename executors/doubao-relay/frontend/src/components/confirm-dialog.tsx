// 居中确认对话框。`confirm(opts)` 返回 Promise<boolean>,同一时间只有一个对话框。
import { signal } from '@preact/signals';
import { useLayoutEffect, useRef } from 'preact/hooks';

export type ConfirmOptions = {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** 危险操作用红色确认按钮。 */
  danger?: boolean;
};

type Pending = ConfirmOptions & { resolve: (ok: boolean) => void };

const pending = signal<Pending | null>(null);

export function confirm(opts: ConfirmOptions): Promise<boolean> {
  // 已有对话框时先把它按"取消"关掉,避免 Promise 悬挂
  pending.value?.resolve(false);
  return new Promise<boolean>((resolve) => {
    pending.value = { ...opts, resolve };
  });
}

function settle(ok: boolean) {
  const p = pending.value;
  if (!p) return;
  pending.value = null;
  p.resolve(ok);
}

/**
 * 快捷键策略(纯函数,便于单测):
 * - Escape → 取消;
 * - Enter 落在对话框内的按钮上(取消 / 确定)→ 不处理,交给按钮自己的 click,否则"取消"会被当成"确定";
 * - 其它位置的 Enter → 确定,并 preventDefault,防止对话框外仍持有焦点的按钮(如触发它的"切换")被再次激活。
 */
export function keyAction(
  key: string,
  target: EventTarget | null,
  dialog: Element | null,
): 'confirm' | 'cancel' | null {
  if (key === 'Escape') return 'cancel';
  if (key !== 'Enter') return null;
  const el = target instanceof Element ? target : null;
  if (el && dialog?.contains(el) && el.closest('button')) return null;
  return 'confirm';
}

export function ConfirmDialog() {
  const p = pending.value;
  const dialogRef = useRef<HTMLDivElement>(null);
  const confirmRef = useRef<HTMLButtonElement>(null);
  // useLayoutEffect:对话框一出现就同步装上快捷键监听并移焦,不等下一帧(useEffect 在 rAF 后才跑,
  // 真机 Chromium 下键盘操作足够快时会漏掉第一个按键)
  useLayoutEffect(() => {
    if (!p) return;
    // autofocus 在页面已有焦点元素(刚点过的"切换"按钮)时会被 Chromium 拒绝,这里显式把焦点挪进对话框
    confirmRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      const action = keyAction(e.key, e.target, dialogRef.current);
      if (!action) return;
      e.preventDefault();
      settle(action === 'confirm');
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [p]);
  if (!p) return null;
  return (
    <div class="dialog-backdrop" onClick={() => settle(false)}>
      <div
        ref={dialogRef}
        class="dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="confirm-dialog-title" class="dialog-title">
          {p.title}
        </h2>
        <p class="dialog-message">{p.message}</p>
        <div class="dialog-actions">
          <button type="button" class="btn" onClick={() => settle(false)}>
            {p.cancelLabel ?? '取消'}
          </button>
          <button
            ref={confirmRef}
            type="button"
            class={`btn ${p.danger ? 'btn-danger' : 'btn-primary'}`}
            onClick={() => settle(true)}
          >
            {p.confirmLabel ?? '确定'}
          </button>
        </div>
      </div>
    </div>
  );
}
