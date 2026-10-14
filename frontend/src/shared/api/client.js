import { reportClientError } from '../error-reporting.js';
import { normalizeLocalExecutorsResponse } from './localExecutors.js';

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

function readableErrorMessage(text, status) {
  try {
    const payload = JSON.parse(text);
    if (payload && typeof payload.error === 'string' && payload.error.trim()) {
      return payload.error.trim();
    }
  } catch (_) {
    // Plain-text responses use their original body below.
  }
  return text || `请求失败：${status}`;
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

export async function apiRequest(path, options = {}) {
  const headers = new Headers(options.headers || {});
  const token = getToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  let response;
  try {
    response = await fetch(path, { ...options, headers });
  } catch (error) {
    const failure = { kind: 'api-network', message: error?.message || 'Network request failed', stack: error?.stack, source: path, method: options.method || 'GET' };
    reportClientError(failure);
    if (!options.silent && !options.suppressGlobalError) notifyApiFailure({ ...failure, path });
    throw attachApiError(error, { path, method: options.method });
  }
  if (response.status === 401 && path !== '/api/login' && isSessionAuthFailure(response)) {
    const failure = { kind: 'api-response', message: '登录已失效，请重新登录', source: path, method: options.method || 'GET', status: response.status };
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
    throw attachApiError(new Error('登录已失效，请重新登录'), { path, method: options.method, status: response.status });
  }
  const allowedStatuses = Array.isArray(options.allowStatuses) ? options.allowStatuses : [];
  if (!response.ok && !allowedStatuses.includes(response.status)) {
    const text = await response.text();
    const error = new Error(readableErrorMessage(text, response.status));
    error.status = response.status;
    error.responseText = text;
    const failure = { kind: 'api-response', message: error.message, stack: error.stack, source: path, method: options.method || 'GET', status: response.status };
    reportClientError(failure);
    if (!options.silent && !options.suppressGlobalError) notifyApiFailure({ ...failure, path });
    throw attachApiError(error, { path, method: options.method, status: response.status });
  }
  if (options.responseType === 'blob') return response.blob();
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    const payload = await response.json();
    if (path === '/api/shuihuo-production/local-executors') {
      return normalizeLocalExecutorsResponse(payload);
    }
    return payload;
  }
  return response.text();
}
export function listMyErrorLogs(limit = 100) {
  return apiRequest(`/api/client-errors/mine?limit=${encodeURIComponent(limit)}`);
}

export function listNovelPanelAiDiagnostics(limit = 100) {
  return apiRequest(`/api/novel-panel/diagnostics?limit=${encodeURIComponent(limit)}`);
}
