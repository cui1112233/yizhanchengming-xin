(() => {
  const apiPrefix = '/api/';
  const mountedApiPrefix = '/api/novel-panel/';
  const allowedHeaders = new Set(['content-type', 'cache-control', 'pragma', 'x-videoprompttool-session']);
  const nonce = new URLSearchParams(window.location.search).get('nonce');
  // The workbench permits a configured AI request to run for up to 400 seconds.
  // Keep the iframe bridge alive slightly longer so it cannot turn a valid
  // long-running model request into a browser-level "Failed to fetch" error.
  const API_BRIDGE_TIMEOUT_MS = 410000;
  const pendingRequests = new Map();
  const nativeFetch = window.fetch.bind(window);
  const nativeSendBeacon = typeof navigator.sendBeacon === 'function'
    ? navigator.sendBeacon.bind(navigator)
    : null;
  let channelPort = null;
  let handshakeRetryTimer = null;

  function notifyParentTask(status, title, detail = '') {
    if (window.parent === window) return;
    window.parent.postMessage({
      type: 'qiantie:task-notification',
      source: '小说面板',
      status: status === 'error' ? 'error' : 'success',
      title: String(title || '小说面板任务已完成').slice(0, 120),
      detail: String(detail || '').replace(/\s+/g, ' ').trim().slice(0, 280),
      page: '小说面板',
      pagePath: '/novel-panel'
    }, '*');
  }

  // The workbench is a classic static application. Expose a deliberately
  // narrow notifier instead of exposing the parent window or session bridge.
  globalThis.__QIANTE_NOTIFY_TASK__ = notifyParentTask;

  function apiPath(input) {
    try {
      const value = input instanceof Request ? input.url : input instanceof URL ? input.href : input;
      const currentUrl = new URL(window.location.href);
      const url = new URL(value, currentUrl);
      if (url.origin !== currentUrl.origin || !url.pathname.startsWith(apiPrefix)) return null;
      // These platform endpoints are intentionally kept intact so the parent
      // page can attach the authenticated user token. Other API calls are
      // mounted under the novel-panel router.
      if (url.pathname === '/api/tts' || url.pathname === '/api/models') return url.pathname + url.search;
      if (!url.pathname.startsWith(mountedApiPrefix)) {
        url.pathname = mountedApiPrefix + url.pathname.slice(apiPrefix.length);
      }
      return url.pathname + url.search;
    } catch (_) {
      return null;
    }
  }

  function allowedRequestHeaders(headers) {
    const result = {};
    for (const [name, value] of new Headers(headers || {})) {
      if (allowedHeaders.has(name.toLowerCase())) result[name] = value;
    }
    return result;
  }

  function nextRequestId() {
    if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
    return `v78-${Date.now()}-${Math.random().toString(36).slice(2)}`;
  }

  function postPendingRequest(pending) {
    channelPort.postMessage({ type: 'novel-panel-api-request', id: pending.id, ...pending.request });
  }

  function base64ToBytes(value) {
    const binary = atob(String(value || ''));
    const bytes = new Uint8Array(binary.length);
    for (let index = 0; index < binary.length; index += 1) bytes[index] = binary.charCodeAt(index);
    return bytes;
  }

  function createAbortError(reason) {
    if (reason) return reason;
    return new DOMException('The operation was aborted.', 'AbortError');
  }

  function clearPending(pending) {
    clearTimeout(pending.timer);
    if (pending.abortSignal && pending.abortListener) {
      pending.abortSignal.removeEventListener('abort', pending.abortListener);
    }
  }

  function requestParentApi(request, signal, { cancelOnClose = true } = {}) {
    const id = nextRequestId();
    return new Promise((resolve, reject) => {
      if (signal?.aborted) {
        reject(createAbortError(signal.reason));
        return;
      }
      const timer = setTimeout(() => {
        cancelPendingRequest(id, new Error('Novel panel API bridge timed out.'));
      }, API_BRIDGE_TIMEOUT_MS);
      const pending = {
        id,
        request,
        resolve,
        reject,
        timer,
        abortSignal: signal,
        abortListener: null,
        cancelOnClose
      };
      pending.abortListener = () => cancelPendingRequest(id, createAbortError(signal.reason));
      signal?.addEventListener?.('abort', pending.abortListener, { once: true });
      pendingRequests.set(id, pending);
      if (channelPort) postPendingRequest(pending);
    });
  }

  function cancelPendingRequest(id, error) {
    const pending = pendingRequests.get(id);
    if (!pending) return;
    pendingRequests.delete(id);
    clearPending(pending);
    if (channelPort) channelPort.postMessage({ type: 'novel-panel-api-cancel', id });
    pending.reject(error || new Error('Novel panel API bridge closed.'));
  }

  function rejectPendingRequests() {
    for (const [id, pending] of pendingRequests) {
      if (pending.cancelOnClose !== false) cancelPendingRequest(id, new Error('Novel panel API bridge closed.'));
    }
  }

  function beaconPayload(data) {
    if (typeof Blob !== 'undefined' && data instanceof Blob) {
      return { body: data, headers: data.type ? { 'Content-Type': data.type } : {} };
    }
    if (typeof data === 'string') {
      const trimmed = data.trim();
      return {
        body: data,
        headers: { 'Content-Type': trimmed.startsWith('{') || trimmed.startsWith('[') ? 'application/json' : 'text/plain;charset=UTF-8' }
      };
    }
    if (data == null) return { body: '', headers: { 'Content-Type': 'text/plain;charset=UTF-8' } };
    return { body: JSON.stringify(data), headers: { 'Content-Type': 'application/json' } };
  }

  function sendHandshake() {
    if (!nonce || channelPort || window.parent === window) return;
    window.parent.postMessage({ type: 'qiantie-v77-handshake', nonce }, '*');
  }

  function stopHandshakeRetries() {
    if (handshakeRetryTimer === null) return;
    clearInterval(handshakeRetryTimer);
    handshakeRetryTimer = null;
  }

  function startHandshakeRetries() {
    if (!nonce || window.parent === window || handshakeRetryTimer !== null) return;
    sendHandshake();
    handshakeRetryTimer = setInterval(sendHandshake, 150);
  }

  window.addEventListener('message', event => {
    if (event.source !== window.parent) return;
    const data = event.data;
    if (!data || data.type !== 'qiantie-theme-sync') return;
    if (data.theme !== 'dark' && data.theme !== 'light') return;
    document.documentElement.dataset.theme = data.theme;
  });

  window.addEventListener('message', event => {
    if (event.source !== window.parent || channelPort) return;
    const data = event.data;
    const port = event.ports?.[0];
    if (!data || data.type !== 'qiantie-v77-port' || data.nonce !== nonce || !port) return;
    channelPort = port;
    stopHandshakeRetries();
    channelPort.onmessage = message => {
      const response = message.data;
      if (!response || response.type !== 'novel-panel-api-response' || typeof response.id !== 'string') return;
      const pending = pendingRequests.get(response.id);
      if (!pending) return;
      pendingRequests.delete(response.id);
      clearPending(pending);
      const status = Number.isInteger(response.status) && response.status >= 200 && response.status <= 599 ? response.status : 500;
      const rawBody = typeof response.bodyBase64 === 'string'
        ? base64ToBytes(response.bodyBase64)
        : (typeof response.text === 'string' ? response.text : '');
      const body = [204, 205, 304].includes(status) ? null : rawBody;
      pending.resolve(new Response(body, { status, headers: response.headers || {} }));
    };
    channelPort.start?.();
    for (const pending of pendingRequests.values()) postPendingRequest(pending);
  });

  function closeBridge() {
    stopHandshakeRetries();
    // CharacterCore listens before this bridge closes so its lease release
    // still travels through the authenticated parent-page channel.
    const eventTarget = typeof globalThis.dispatchEvent === 'function' ? globalThis : window;
    eventTarget.dispatchEvent?.(new Event('qiantie-v77-bridge-closing'));
    rejectPendingRequests();
    channelPort?.close();
    channelPort = null;
  }

  window.addEventListener('pagehide', closeBridge);
  window.addEventListener('pageshow', event => {
    if (!event.persisted) return;
    closeBridge();
    startHandshakeRetries();
  });
  window.addEventListener('beforeunload', closeBridge);

  startHandshakeRetries();

  window.fetch = function novelPanelFetch(input, init = {}) {
    const path = apiPath(input);
    if (!path) return nativeFetch(input, init);
    const method = String(init.method || (input instanceof Request ? input.method : 'GET')).toUpperCase();
    const headers = allowedRequestHeaders(init.headers || (input instanceof Request ? input.headers : undefined));
    const signal = init.signal || (input instanceof Request ? input.signal : undefined);
    const hasBody = Object.prototype.hasOwnProperty.call(init, 'body');
    if (input instanceof Request && !hasBody && method !== 'GET' && method !== 'HEAD') {
      return input.clone().text().then(body => requestParentApi({ path, method, headers, body }, signal));
    }
    return requestParentApi({ path, method, headers, body: hasBody ? init.body : undefined }, signal);
  };

  navigator.sendBeacon = function novelPanelSendBeacon(endpoint, data) {
    const path = apiPath(endpoint);
    if (!path) return nativeSendBeacon ? nativeSendBeacon(endpoint, data) : false;
    const payload = beaconPayload(data);
    requestParentApi(
      { path, method: 'POST', headers: payload.headers, body: payload.body },
      undefined,
      { cancelOnClose: false }
    ).catch(() => {});
    return true;
  };
})();
