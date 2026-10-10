import { Alert } from 'antd';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { getToken } from '../../shared/api/client';
import { dispatchPetContext } from '../../shared/pet/stacky';

const ALLOWED_METHODS = new Set(['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD']);
const ALLOWED_HEADERS = new Set(['content-type', 'cache-control', 'pragma', 'x-videoprompttool-session']);
const MAX_REQUEST_BODY_BYTES = 4 * 1024 * 1024;
const MAX_REQUEST_ID_LENGTH = 120;
const MAX_HEADER_VALUE_LENGTH = 512;
const MAX_ACTIVE_CONTROLLERS = 8;
const MAX_RESPONSE_BODY_BYTES = 12 * 1024 * 1024;
const MAX_RESPONSE_HEADER_BYTES = 8 * 1024;

function isAllowedNovelPanelApiPath(pathname) {
  return pathname === '/api/tts' || pathname === '/api/models' || pathname.startsWith('/api/novel-panel/');
}

function normalizeTheme(theme) {
  return theme === 'light' ? 'light' : 'dark';
}

function syncTheme(frame, theme) {
  frame?.contentWindow?.postMessage(
    { type: 'qiantie-theme-sync', theme: normalizeTheme(theme) },
    '*'
  );
}

function utf8ByteLength(value) {
  return new TextEncoder().encode(value).length;
}

function normalizeBridgeRequest(data) {
  if (!data || data.type !== 'novel-panel-api-request' || typeof data.id !== 'string' || typeof data.path !== 'string') return { error: 400, id: typeof data?.id === 'string' ? data.id.slice(0, MAX_REQUEST_ID_LENGTH) : '' };
  if (!data.id || data.id.length > MAX_REQUEST_ID_LENGTH) return { error: 413, id: data.id.slice(0, MAX_REQUEST_ID_LENGTH) };
  if (data.body !== undefined && typeof data.body !== 'string') return { error: 400, id: data.id };
  if (typeof data.body === 'string' && utf8ByteLength(data.body) > MAX_REQUEST_BODY_BYTES) return { error: 413, id: data.id };
  const method = String(data.method || 'GET').toUpperCase();
  if (!ALLOWED_METHODS.has(method)) return { error: 400, id: data.id };
  let url;
  try {
    url = new URL(data.path, window.location.origin);
  } catch (_) {
    return { error: 400, id: data.id };
  }
  if (url.origin !== window.location.origin || !isAllowedNovelPanelApiPath(url.pathname)) return { error: 400, id: data.id };

  const headers = {};
  if (data.headers !== undefined && (!data.headers || typeof data.headers !== 'object' || Array.isArray(data.headers))) return { error: 400, id: data.id };
  for (const [name, value] of Object.entries(data.headers || {})) {
    if (typeof value === 'string' && value.length > MAX_HEADER_VALUE_LENGTH) return { error: 413, id: data.id };
    if (ALLOWED_HEADERS.has(name.toLowerCase()) && typeof value === 'string') headers[name] = value;
  }
  return { id: data.id, path: url.pathname + url.search, method, headers, body: data.body };
}

function postBridgeResponse(target, id, response) {
  target.postMessage({ type: 'novel-panel-api-response', id, ...response });
}

function boundedResponseHeaders(headers) {
  const output = {};
  let size = 0;
  for (const [name, value] of headers.entries()) {
    const entrySize = utf8ByteLength(name) + utf8ByteLength(value);
    if (value.length > MAX_HEADER_VALUE_LENGTH || size + entrySize > MAX_RESPONSE_HEADER_BYTES) continue;
    output[name] = value;
    size += entrySize;
  }
  return output;
}

async function readResponseBytes(response) {
  const reader = response.body?.getReader?.();
  if (!reader) return new Uint8Array(await response.arrayBuffer());
  const chunks = [];
  let size = 0;
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      size += value.byteLength;
      if (size > MAX_RESPONSE_BODY_BYTES) throw new Error('Novel panel API response exceeds bridge limits');
      chunks.push(value);
    }
  } finally {
    reader.releaseLock();
  }
  const bytes = new Uint8Array(size);
  let offset = 0;
  for (const chunk of chunks) {
    bytes.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return bytes;
}

function bytesToBase64(bytes) {
  let binary = '';
  for (let index = 0; index < bytes.length; index += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(index, index + 0x8000));
  }
  return btoa(binary);
}

function createSessionNonce() {
  const values = new Uint8Array(24);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(values);
    return Array.from(values, value => value.toString(16).padStart(2, '0')).join('');
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function createInstanceId() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  return createSessionNonce();
}

export function NovelPanelPage({ theme }) {
  const frameRef = useRef(null);
  const sessionNonceRef = useRef(createSessionNonce());
  const instanceIdRef = useRef(createInstanceId());
  const handshakeConsumedRef = useRef(false);
  const portRef = useRef(null);
  const controllersRef = useRef(new Map());
  const channelLoadAcknowledgedRef = useRef(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const searchParams = new URLSearchParams(window.location.search);
  const projectId = searchParams.get('project');
  // Use the concrete document instead of the directory alias. The Go static
  // server supports both, but an older SPA fallback can mistake the directory
  // URL for the main application and embed the home page inside the panel.
  const workbenchSrc = `/novel-panel/workbench/index.html?nonce=${encodeURIComponent(sessionNonceRef.current)}&instance=${encodeURIComponent(instanceIdRef.current)}${projectId ? `&project=${encodeURIComponent(projectId)}` : ''}`;

  function abortPendingRequests() {
    for (const controller of controllersRef.current.values()) controller.abort();
    controllersRef.current.clear();
  }

  useEffect(() => {
    syncTheme(frameRef.current, theme);
  }, [theme]);

  useEffect(() => {
    const stage = error ? '加载失败' : loading ? '加载中' : '可用';
    const opaqueProjectId = String(projectId || '').slice(0, 160);
    dispatchPetContext({
      page: '小说面板',
      pagePath: '/novel-panel',
      summary: `当前项目：${opaqueProjectId || '未选择'}；工作台：${stage}`,
      entities: { projectId: opaqueProjectId, stage },
      actions: ['检查当前工作台', '继续编辑项目']
    });
  }, [projectId, loading, error]);

  useEffect(() => {
    function receiveWorkbenchPetContext(event) {
      if (event.source !== frameRef.current?.contentWindow) return;
      const data = event.data;
      if (!data || data.type !== 'qiantie-novel-panel-pet-context' || !data.context || typeof data.context !== 'object') return;
      const context = data.context;
      dispatchPetContext({
        page: '小说面板',
        pagePath: '/novel-panel',
        summary: String(context.summary || '小说面板工作台已就绪').slice(0, 1600),
        entities: context.entities && typeof context.entities === 'object' ? context.entities : {},
        actions: Array.isArray(context.actions) ? context.actions.slice(0, 6) : [],
        novelText: String(context.novelText || '').slice(0, 4500),
        extracted: context.extracted && typeof context.extracted === 'object' ? context.extracted : undefined
      });
    }
    window.addEventListener('message', receiveWorkbenchPetContext);
    return () => window.removeEventListener('message', receiveWorkbenchPetContext);
  }, []);

  function requestWorkbenchPetContext() {
    frameRef.current?.contentWindow?.postMessage({ type: 'qiantie-novel-panel-pet-context-request' }, '*');
  }

  useLayoutEffect(() => {
    function closePort() {
      abortPendingRequests();
      if (!portRef.current) return;
      portRef.current.onmessage = null;
      portRef.current.close();
      portRef.current = null;
      handshakeConsumedRef.current = false;
      channelLoadAcknowledgedRef.current = false;
    }

    async function handlePortRequest(port, event) {
      if (port !== portRef.current) return;
      const data = event.data;
      if (data?.type === 'novel-panel-api-cancel' && typeof data.id === 'string') {
        const controller = controllersRef.current.get(data.id);
        controller?.abort();
        controllersRef.current.delete(data.id);
        return;
      }
      const request = normalizeBridgeRequest(data);
      if (request?.error) {
        postBridgeResponse(port, request.id, {
          status: request.error,
          headers: { 'content-type': 'application/json' },
          text: JSON.stringify({ error: request.error === 413 ? 'Novel panel request exceeds bridge limits' : 'Invalid novel panel API request' })
        });
        return;
      }
      if (!request) return;
      if (controllersRef.current.size >= MAX_ACTIVE_CONTROLLERS) {
        postBridgeResponse(port, request.id, {
          status: 429,
          headers: { 'content-type': 'application/json' },
          text: JSON.stringify({ error: 'Novel panel bridge is busy' })
        });
        return;
      }
      if (controllersRef.current.has(request.id)) {
        postBridgeResponse(port, request.id, {
          status: 409,
          headers: { 'content-type': 'application/json' },
          text: JSON.stringify({ error: 'Novel panel request id is already active' })
        });
        return;
      }

      const headers = new Headers(request.headers);
      headers.delete('Authorization');
      const token = getToken();
      if (token) headers.set('Authorization', `Bearer ${token}`);
      const controller = new AbortController();
      controllersRef.current.set(request.id, controller);

      try {
        const response = await fetch(request.path, {
          method: request.method,
          headers,
          body: request.method === 'GET' || request.method === 'HEAD' ? undefined : request.body,
          signal: controller.signal
        });
        if (controller.signal.aborted || port !== portRef.current) return;
        const responseHeaders = boundedResponseHeaders(response.headers);
        const contentType = response.headers.get('content-type') || '';
        const isBinary = /^(audio|image)\//i.test(contentType);
        const responseBytes = await readResponseBytes(response);
        postBridgeResponse(port, request.id, {
          status: response.status,
          headers: responseHeaders,
          ...(isBinary ? { bodyBase64: bytesToBase64(responseBytes) } : { text: new TextDecoder().decode(responseBytes) })
        });
      } catch (_) {
        if (controller.signal.aborted || port !== portRef.current) return;
        postBridgeResponse(port, request.id, {
          status: 502,
          headers: { 'content-type': 'application/json' },
          text: JSON.stringify({ error: 'Novel panel API bridge failed' })
        });
      } finally {
        if (controllersRef.current.get(request.id) === controller) controllersRef.current.delete(request.id);
      }
    }

    function handleHandshake(event) {
      if (event.source !== frameRef.current?.contentWindow) return;
      const data = event.data;
      if (handshakeConsumedRef.current || !data || data.type !== 'qiantie-v77-handshake' || data.nonce !== sessionNonceRef.current) return;

      handshakeConsumedRef.current = true;
      const channel = new MessageChannel();
      const port1 = channel.port1;
      portRef.current = port1;
      channelLoadAcknowledgedRef.current = false;
      port1.onmessage = event => { void handlePortRequest(port1, event); };
      port1.start?.();
      event.source.postMessage({ type: 'qiantie-v77-port', nonce: sessionNonceRef.current }, '*', [channel.port2]);
    }

    function handlePageShow(event) {
      if (!event.persisted) return;
      closePort();
      handshakeConsumedRef.current = false;
      channelLoadAcknowledgedRef.current = false;
    }

    window.addEventListener('message', handleHandshake);
    window.addEventListener('pageshow', handlePageShow);
    return () => {
      window.removeEventListener('message', handleHandshake);
      window.removeEventListener('pageshow', handlePageShow);
      closePort();
    };
  }, []);

  function handleFrameLoad() {
    setLoading(false);
    syncTheme(frameRef.current, theme);
    requestWorkbenchPetContext();
    if (!portRef.current) return;
    if (!channelLoadAcknowledgedRef.current) {
      channelLoadAcknowledgedRef.current = true;
      return;
    }
    abortPendingRequests();
    portRef.current.onmessage = null;
    portRef.current.close();
    portRef.current = null;
    handshakeConsumedRef.current = false;
    channelLoadAcknowledgedRef.current = false;
  }

  function PanelLoadingVisual() {
    return <div className="novel-panel-loading-visual" aria-label="正在连接小说面板" role="status">
      <svg viewBox="0 0 520 260" className="novel-panel-loader-svg" aria-hidden="true">
        <rect x="170" y="92" width="180" height="76" rx="20" className="novel-panel-chip-body" />
        <text x="260" y="135" textAnchor="middle" className="novel-panel-chip-text">小说面板</text>
        <path d="M20 55 H155 Q170 55 185 92 M500 55 H365 Q350 55 335 92 M20 205 H155 Q170 205 185 168 M500 205 H365 Q350 205 335 168" className="novel-panel-trace-bg" />
        <path d="M20 55 H155 Q170 55 185 92 M500 55 H365 Q350 55 335 92 M20 205 H155 Q170 205 185 168 M500 205 H365 Q350 205 335 168" className="novel-panel-trace-flow" />
      </svg>
      <strong>正在连接小说面板</strong><span>工作台加载完成后会显示在这里</span>
    </div>;
  }

  return (
    <div className="novel-panel-page">
      {loading && !error ? <div className="novel-panel-status"><PanelLoadingVisual /></div> : null}
      {error ? <Alert className="novel-panel-error" type="error" showIcon message="小说面板加载失败" /> : null}
      <iframe
        ref={frameRef}
        className="novel-panel-frame"
        title="小说面板"
        src={workbenchSrc}
        sandbox="allow-scripts allow-forms allow-downloads allow-modals allow-same-origin"
        onLoad={handleFrameLoad}
        onError={() => {
          setLoading(false);
          setError(true);
        }}
      />
    </div>
  );
}

export default NovelPanelPage;
