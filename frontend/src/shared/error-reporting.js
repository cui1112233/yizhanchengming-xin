const MAX_REPORT_LENGTH = 4000;
// 同一个面板的同一个错误只上报一次。状态接口 503 会被轮询反复触发，
// 若不去重，/api/client-errors 会变成“自己刷自己的错误上报”。
const DEDUPE_WINDOW_MS = 10 * 60 * 1000;
const MAX_DEDUPE_ENTRIES = 200;
const reportedSignatures = new Map();

function clean(value) {
  return String(value || '').slice(0, MAX_REPORT_LENGTH);
}

export function resetClientErrorDedupe() {
  reportedSignatures.clear();
}

function serializeReason(reason) {
  if (reason instanceof Error) return { message: reason.message, stack: reason.stack || '' };
  if (typeof reason === 'string') return { message: reason, stack: '' };
  try {
    return { message: JSON.stringify(reason), stack: '' };
  } catch (_) {
    return { message: 'Unknown browser error', stack: '' };
  }
}

export function reportClientError({ kind = 'error', message, stack, source, method, status } = {}) {
  if (typeof window === 'undefined' || typeof localStorage === 'undefined') return;
  const record = {
    kind: clean(kind),
    message: clean(message),
    source: clean(source),
    path: window.location.pathname,
    method: clean(method),
    status: Number.isInteger(status) ? status : undefined
  };
  const signature = [record.kind, record.message, record.source, record.method, record.status ?? '', record.path].join('|');
  const now = Date.now();
  const lastReportedAt = reportedSignatures.get(signature);
  if (lastReportedAt && now - lastReportedAt < DEDUPE_WINDOW_MS) return;
  reportedSignatures.set(signature, now);
  if (reportedSignatures.size > MAX_DEDUPE_ENTRIES) {
    for (const [key, reportedAt] of reportedSignatures) {
      if (now - reportedAt >= DEDUPE_WINDOW_MS) reportedSignatures.delete(key);
    }
  }
  const payload = JSON.stringify({ ...record, stack: clean(stack) });

  try {
    const token = localStorage.getItem('auth_token') || '';
    void fetch('/api/client-errors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
      body: payload,
      keepalive: true
    }).catch(() => {});
  } catch (_) {
    // Reporting cannot be allowed to create a second browser failure.
  }
}

export function installClientErrorReporting() {
  window.addEventListener('error', event => {
    reportClientError({
      kind: 'window-error',
      message: event.message,
      stack: event.error?.stack,
      source: event.filename
    });
  });

  window.addEventListener('unhandledrejection', event => {
    const reason = serializeReason(event.reason);
    reportClientError({ kind: 'unhandledrejection', ...reason });
  });
}
