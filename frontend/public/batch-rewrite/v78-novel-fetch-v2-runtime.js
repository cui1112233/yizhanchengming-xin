(() => {
  if (window.QiantieNovelFetchRuntime) return;

  const API_ROOT = '/api/novel-fetch-workshop';
  const WEB_SUBMIT_API = '/api/novel-fetch-workshop/web-submit';
  const inFlight = new Map();
  const controllers = new Map();
  const cache = new Map();
  const CACHE_MS = 5000;
  const POLL_MS = 5000;
  let pollTimer = null;
  let pollCallback = null;
  function authTokenFromSharedSession() {
    let localToken = localStorage.getItem('auth_token') || '';
    if (!localToken && window.parent && window.parent !== window) {
      try { localToken = window.parent.localStorage.getItem('auth_token') || ''; } catch (_) {}
    }
    if (localToken) return localToken;
    let cookieText = document.cookie;
    if (!cookieText && window.parent && window.parent !== window) {
      try { cookieText = window.parent.document.cookie; } catch (_) {}
    }
    const cookie = cookieText.split(';').map(item => item.trim()).find(item => item.startsWith('qiantie_auth_token='));
    if (!cookie) return '';
    try {
      const value = decodeURIComponent(cookie.slice('qiantie_auth_token='.length));
      if (value) localStorage.setItem('auth_token', value);
      return value;
    } catch (_) { return ''; }
  }

  let authToken = authTokenFromSharedSession();
  let authScope = 0;

  window.addEventListener('message', event => {
    if (event.origin !== window.location.origin || event.data?.type !== 'qiantie-auth-token') return;
    const next = String(event.data?.token || '').trim();
    if (!next || next === authToken) return;
    authToken = next;
    try { localStorage.setItem('auth_token', next); } catch (_) {}
    authScope += 1;
    cache.clear();
  });

  function syncAuthScope() {
    const current = authTokenFromSharedSession();
    if (current !== authToken) {
      authToken = current;
      authScope += 1;
      cache.clear();
    }
  }

  function tokenHeaders(extra = {}) {
    const token = authTokenFromSharedSession();
    return { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...extra };
  }

  function normalizePath(path) {
    const value = String(path || '');
    return value.startsWith('/api/') ? value.slice(4) : value;
  }

  function requestUrl(normalized) {
    // 文本模型目录属于统一 API 配置路由，不属于旧工作台资源路由。
    if (normalized.startsWith(WEB_SUBMIT_API.slice(4))) return `/api${normalized}`;
    // 调用方可能传入完整的 /api/novel-fetch-workshop/... 路径。
    // normalizePath 已去掉 /api 前缀，此时直接还原即可，不能再次拼接
    // API_ROOT，否则会变成 /api/novel-fetch-workshop/novel-fetch-workshop/...
    if (normalized.startsWith(API_ROOT.slice(4))) return `/api${normalized}`;
    return /^\/models(?:\?|$)/.test(normalized) ? `/api${normalized}` : `${API_ROOT}${normalized}`;
  }

  function requestKey(path, options = {}) {
    syncAuthScope();
    const method = String(options.method || 'GET').toUpperCase();
    const body = typeof options.body === 'string' ? options.body : '';
    return `${method} ${normalizePath(path)} ${body} scope:${authScope}`;
  }

  async function request(path, options = {}) {
    const normalized = normalizePath(path);
    const method = String(options.method || 'GET').toUpperCase();
    const key = requestKey(normalized, options);
    const now = Date.now();
    const cached = cache.get(key);
    if (method === 'GET' && !options.noCache) {
      if (cached && cached.expiresAt > now) return cached.value;
    }
    if (inFlight.has(key)) return inFlight.get(key);
    const controller = options.signal ? null : new AbortController();
    const promise = (async () => {
      const headers = tokenHeaders(options.headers || {});
      if (method === 'GET' && cached?.etag && !headers['If-None-Match']) headers['If-None-Match'] = cached.etag;
      const response = await fetch(requestUrl(normalized), { ...options, headers, ...(controller ? { signal: controller.signal } : {}) });
      if (response.status === 304 && cached) {
        cached.expiresAt = Date.now() + CACHE_MS;
        return cached.value;
      }
      const text = await response.text();
      let data = {};
      try { data = text ? JSON.parse(text) : {}; } catch (_) { data = { raw: text }; }
      if (!response.ok) {
        const error = new Error(data.error || data.detail || data.message || data.raw || `HTTP ${response.status}`);
        error.status = response.status;
        throw error;
      }
      if (method === 'GET' && !options.noCache) cache.set(key, { value: data, etag: response.headers.get('etag') || '', expiresAt: Date.now() + CACHE_MS });
      if (method !== 'GET') {
        const resource = normalized.split('?')[0];
        invalidate(resource);
        if (resource === '/config') invalidate('/bootstrap');
      }
      return data;
    })();
    inFlight.set(key, promise);
    if (controller) controllers.set(key, controller);
    try { return await promise; } finally { inFlight.delete(key); controllers.delete(key); }
  }

  function cancel(key) {
    const controller = controllers.get(String(key || ''));
    if (controller) controller.abort();
  }

  function invalidate(prefix = '') {
    const value = normalizePath(prefix);
    for (const key of cache.keys()) if (!value || key.includes(` ${value}`)) cache.delete(key);
  }

  function isActiveTask(task = {}) {
    const value = [task.status, task.original_status, task.ai_status, task.classify_status, task.site_submit_status]
      .filter(Boolean).join(' ').toLowerCase();
    return /(queued|pending|running|processing|classifying|generating|waiting|retry|排队|处理中|等待|执行中)/i.test(value)
      && !/(done|completed|failed|error|cancelled|stopped|完成|失败|取消|停止)/i.test(value);
  }

  function stopPolling() {
    if (pollTimer) window.clearInterval(pollTimer);
    pollTimer = null;
    pollCallback = null;
  }

  function startPolling(callback) {
    if (typeof callback !== 'function') return;
    pollCallback = callback;
    if (!pollTimer) pollTimer = window.setInterval(() => { void pollCallback?.(); }, POLL_MS);
  }

  window.QiantieNovelFetchRuntime = {
    api: request,
    request,
    singleFlight: request,
    requestKey,
    inFlight,
    controllers,
    cancel,
    invalidate,
    isActiveTask,
    startPolling,
    stopPolling,
    pollMs: POLL_MS,
    bootstrap: () => request('/bootstrap'),
    config: (options = {}) => request('/config', options),
    tasks: (query = '') => request(`/tasks${query}`),
    currentBatch: () => request('/batches/current')
  };
  window.qiantieNovelFetchAuthToken = authTokenFromSharedSession;
})();
