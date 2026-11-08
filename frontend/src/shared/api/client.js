import { reportClientError } from '../error-reporting.js';
import { normalizeLocalExecutorsResponse } from './localExecutors.js';
import { getSelectedNovelFetchAccount, isNovelFetchScopedPath } from './novelFetchAccountScope.js';

export function getToken() {
  // The Go application is now the single runtime origin. Keep the session in
  // that origin's localStorage; no legacy Node port or cookie bridge is needed.
  return localStorage.getItem('auth_token') || '';
}

export function setToken(token) {
  if (token) {
    localStorage.setItem('auth_token', token);
  } else {
    localStorage.removeItem('auth_token');
  }
}

export function isSessionAuthFailure(response) {
  return response?.headers?.get?.('X-Qiantie-Auth-Failure') === 'session';
}

function parseApiError(text, status) {
  try {
    const payload = JSON.parse(text);
    if (payload && typeof payload === 'object') {
      const message = typeof payload.error === 'string' && payload.error.trim() ? payload.error.trim() : '';
      const code = typeof payload.code === 'string' ? payload.code.trim() : '';
      return { message: message || text || `请求失败：${status}`, code };
    }
  } catch (_) {
    // Plain-text responses use their original body below.
  }
  return { message: text || `请求失败：${status}`, code: '' };
}

function notifyApiFailure({ path, method, status, message, sessionAuthFailure = false }) {
  if (path === '/api/client-errors' || typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent('qiantie:api-error', {
    detail: {
      source: path,
      method: method || 'GET',
      status: Number.isInteger(status) ? status : 0,
      message: message || '请求失败，请稍后重试。',
      sessionAuthFailure: sessionAuthFailure === true
    }
  }));
}

function attachApiError(error, { path, method, status }) {
  error.status = status;
  error.source = path;
  error.method = method || 'GET';
  return error;
}

// 统一请求超时。控制面接口默认 60s：本机/内网 Go 服务的普通读写通常 <1s，
// 60s 足以覆盖偶发慢查询，同时保证任何挂住的请求都会失败退出，不会让界面
// 永久停在加载态。大文件上传/下载默认放宽到 10 分钟，长任务接口可按调用传
// `timeoutMs` 覆盖；传 0 表示由调用方自己负责取消。
export const DEFAULT_API_TIMEOUT_MS = 60 * 1000;
export const LONG_RUNNING_API_TIMEOUT_MS = 10 * 60 * 1000;
export const API_TIMEOUT_CODE = 'timeout';

export function isApiTimeoutError(error) {
  return error?.code === API_TIMEOUT_CODE || error?.name === 'TimeoutError';
}

function isLongRunningRequest(options) {
  if (options?.responseType === 'blob') return true;
  if (typeof FormData !== 'undefined' && options?.body instanceof FormData) return true;
  return false;
}

function createTimeoutError({ path, method, timeoutMs }) {
  const seconds = Math.max(1, Math.round(timeoutMs / 1000));
  const error = new Error(`请求超时（${seconds} 秒未返回）：${method} ${path}`);
  error.name = 'TimeoutError';
  error.code = API_TIMEOUT_CODE;
  error.timeoutMs = timeoutMs;
  error.status = 0;
  error.source = path;
  error.method = method;
  return error;
}

function reportTimeoutFailure({ path, method, timeoutMs }, options) {
  const error = createTimeoutError({ path, method, timeoutMs });
  reportClientError({ kind: 'api-timeout', message: error.message, source: path, method, timeoutMs });
  if (!options.silent && !options.suppressGlobalError) {
    notifyApiFailure({ path, method, status: 0, message: error.message });
  }
  return error;
}

export async function apiRequest(path, options = {}) {
  const headers = new Headers(options.headers || {});
  const token = getToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (isNovelFetchScopedPath(path)) {
    const selectedAccount = getSelectedNovelFetchAccount();
    if (selectedAccount) headers.set('X-Qiantie-Novel-Fetch-Account', selectedAccount);
  }
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const method = options.method || 'GET';
  const timeoutMs = Number.isFinite(options.timeoutMs)
    ? Math.max(0, Number(options.timeoutMs))
    : (isLongRunningRequest(options) ? LONG_RUNNING_API_TIMEOUT_MS : DEFAULT_API_TIMEOUT_MS);
  const externalSignal = options.signal;
  const controller = new AbortController();
  let timedOut = false;
  let timeoutId = null;
  const abortFromExternal = () => {
    if (controller.signal.aborted) return;
    controller.abort(externalSignal?.reason);
  };
  if (externalSignal) {
    if (externalSignal.aborted) abortFromExternal();
    else externalSignal.addEventListener('abort', abortFromExternal, { once: true });
  }
  if (timeoutMs > 0) {
    timeoutId = setTimeout(() => {
      timedOut = true;
      controller.abort();
    }, timeoutMs);
  }
  const fetchOptions = { ...options, headers, signal: controller.signal };
  delete fetchOptions.timeoutMs;

  try {
    let response;
    try {
      response = await fetch(path, fetchOptions);
    } catch (error) {
      // fetch 挂住被超时中止：给出可识别的 TimeoutError，而不是把 AbortError
      // 当成普通网络错误，方便上层区分「超时」与「业务失败」。
      if (timedOut) throw reportTimeoutFailure({ path, method, timeoutMs }, options);
      const failure = { kind: 'api-network', message: error?.message || 'Network request failed', stack: error?.stack, source: path, method };
      reportClientError(failure);
      if (!options.silent && !options.suppressGlobalError) notifyApiFailure({ ...failure, path });
      throw attachApiError(error, { path, method });
    }
    if (response.status === 401 && path !== '/api/login' && isSessionAuthFailure(response)) {
      const failure = { kind: 'api-response', message: '登录已失效，请重新登录', source: path, method, status: response.status };
      // The request may complete after logout or after a newer account has
      // replaced its token. Preserve the diagnostic, but never surface a stale
      // response as a new login failure.
      const sessionStillCurrent = getToken() === token;
      reportClientError(failure);
      if (sessionStillCurrent && !options.silent && !options.suppressGlobalError) notifyApiFailure({ ...failure, path, sessionAuthFailure: true });
      // A slow request from an older account must never erase a newer session.
      // Notify before clearing so the layout can compare the request token with
      // the session that is currently active.
      if (sessionStillCurrent && typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('qiantie:auth-expired', { detail: { token } }));
        setToken('');
        localStorage.removeItem('auth_username');
      }
      throw attachApiError(new Error('登录已失效，请重新登录'), { path, method, status: response.status });
    }
    const allowedStatuses = Array.isArray(options.allowStatuses) ? options.allowStatuses : [];
    if (!response.ok && !allowedStatuses.includes(response.status)) {
      const text = await response.text();
      const parsedError = parseApiError(text, response.status);
      const error = new Error(parsedError.message);
      // 服务端稳定错误码透传给调用方；没有 code 的旧接口仍按 error 文案处理。
      if (parsedError.code) error.code = parsedError.code;
      error.status = response.status;
      error.responseText = text;
      const failure = { kind: 'api-response', message: error.message, stack: error.stack, source: path, method, status: response.status };
      reportClientError(failure);
      if (!options.silent && !options.suppressGlobalError) notifyApiFailure({ ...failure, path });
      throw attachApiError(error, { path, method, status: response.status });
    }
    if (options.responseType === 'blob') return await response.blob();
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      const payload = await response.json();
      if (path === '/api/shuihuo-production/local-executors') {
        return normalizeLocalExecutorsResponse(payload);
      }
      return payload;
    }
    return await response.text();
  } catch (error) {
    // 响应体读取阶段被超时中止（不是上方 fetch 阶段），同样转成可识别的超时错误。
    if (timedOut && error?.name === 'AbortError') {
      throw reportTimeoutFailure({ path, method, timeoutMs }, options);
    }
    throw error;
  } finally {
    if (timeoutId) clearTimeout(timeoutId);
    if (externalSignal?.removeEventListener) externalSignal.removeEventListener('abort', abortFromExternal);
  }
}
export function listMyErrorLogs(limit = 100) {
  return apiRequest(`/api/client-errors/mine?limit=${encodeURIComponent(limit)}`);
}
