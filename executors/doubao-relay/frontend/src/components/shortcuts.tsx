// spec 06 快捷键:Cmd/Ctrl+, 设置 · Cmd/Ctrl+L 日志 · Cmd/Ctrl+R 刷新账号 · Cmd/Ctrl+N 新建任务 · Cmd/Ctrl+K 搜索任务。
// 前三个在这里直接处理;新建 / 搜索属于任务页(Task 2.9),统一以 window 事件 `relay:shortcut` 广播,
// 任务页 `addEventListener('relay:shortcut', e => e.detail === 'new_task' …)` 即可接上,不必改这里。
import { useEffect } from 'preact/hooks';
import { loadAccounts } from '../pages/accounts';
import { doubaoStatus, route } from '../state';
import { shortcutFor } from '../view-model';
import type { ShortcutAction } from '../view-model';

export const SHORTCUT_EVENT = 'relay:shortcut';

export const isMacPlatform = (): boolean =>
  typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform ?? '');

/** 执行一个快捷键动作(导出便于单测)。返回是否已处理。 */
export function runShortcut(action: ShortcutAction): boolean {
  switch (action) {
    case 'settings':
      route.value = 'settings';
      return true;
    case 'logs':
      route.value = 'logs';
      return true;
    case 'refresh_accounts':
      route.value = 'accounts';
      if (doubaoStatus.value?.cdpConnected) void loadAccounts(true);
      return true;
    default:
      return false;
  }
}

export function handleShortcutKeydown(e: KeyboardEvent, isMac = isMacPlatform()): void {
  const action = shortcutFor(e, isMac);
  if (!action) return;
  // 浏览器 / WebView 默认行为(Cmd+R 刷新页面、Cmd+L 地址栏等)一律拦下
  e.preventDefault();
  runShortcut(action);
  window.dispatchEvent(new CustomEvent<ShortcutAction>(SHORTCUT_EVENT, { detail: action }));
}

/** 挂在 App 根上的无渲染组件。 */
export function Shortcuts() {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => handleShortcutKeydown(e);
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);
  return null;
}
