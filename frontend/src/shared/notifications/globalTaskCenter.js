export const GLOBAL_TASK_NOTIFICATION_EVENT = 'qiantie:task-notification';
export const GLOBAL_STATUS_EVENT = 'qiantie:global-status';

export function dispatchGlobalStatus({ text = '', tone = 'info', action = null, durationMs } = {}) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(GLOBAL_STATUS_EVENT, { detail: { text: String(text || ''), tone, action, durationMs } }));
}

const MAX_TEXT_LENGTH = 280;

function compactText(value, maxLength = MAX_TEXT_LENGTH) {
  return String(value || '').replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

export function normalizeGlobalTaskNotification(input = {}) {
  const status = input.status === 'error' ? 'error' : input.status === 'working' ? 'working' : 'success';
  return {
    id: compactText(input.id, 120) || `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    status,
    title: compactText(input.title, 120) || (status === 'error' ? '任务执行失败' : status === 'working' ? '任务正在执行' : '任务已完成'),
    detail: compactText(input.detail),
    page: compactText(input.page, 80),
    pagePath: compactText(input.pagePath, 240),
    createdAt: Number.isFinite(input.createdAt) ? input.createdAt : Date.now()
  };
}

export function dispatchGlobalTaskNotification(notification) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(GLOBAL_TASK_NOTIFICATION_EVENT, {
    detail: normalizeGlobalTaskNotification(notification)
  }));
}
