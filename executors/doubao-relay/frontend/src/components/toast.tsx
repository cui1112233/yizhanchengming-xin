// 右下角 toast 堆叠。普通 3s 自动消失(state.pushToast 负责计时),错误常驻直到点关。
import { dismissToast, toasts } from '../state';

export function ToastHost() {
  const items = toasts.value;
  if (items.length === 0) return null;
  return (
    <div class="toast-host" aria-live="polite">
      {items.map((t) => (
        <div key={t.id} class={`toast toast-${t.kind}`} role={t.kind === 'error' ? 'alert' : 'status'}>
          <span class="toast-message">{t.message}</span>
          <button
            type="button"
            class="toast-close"
            aria-label="关闭提示"
            onClick={() => dismissToast(t.id)}
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
}
