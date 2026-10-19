export const CM_BRIDGE_CONTEXT_EVENT = 'qiantie:cm-bridge-context';
export const CM_BRIDGE_SELECTION_EVENT = 'qiantie:cm-bridge-selection';
export const CM_BRIDGE_ACTION_EVENT = 'qiantie:cm-bridge-action';
export const CM_BRIDGE_ACTION_RESULT_EVENT = 'qiantie:cm-bridge-action-result';
export const CM_BRIDGE_UNDO_EVENT = 'qiantie:cm-bridge-undo';

export const CM_ACTION_TYPES = Object.freeze([
  'character.update',
  'character.create',
  'character.delete',
  'character.setProtagonist',
  'scene.update',
  'scene.create',
  'scene.delete',
  'script.replace',
  'script.insert',
  'shot.update',
  'constraint.bind',
  'constraint.update',
  'asset.update',
  'asset.create',
  'segment.update',
  'segment.bindAsset',
  'tts.update',
  'novel.selection.replace'
]);

const actionTypes = new Set(CM_ACTION_TYPES);
let activeBridge = null;
let actionTail = Promise.resolve();

const defaultCapabilities = {
  '/script': ['character.update', 'character.create', 'character.setProtagonist', 'scene.update', 'scene.create', 'script.replace', 'script.insert', 'shot.update', 'constraint.bind', 'constraint.update'],
  '/batch-factory': ['asset.update', 'asset.create', 'segment.update', 'segment.bindAsset'],
  '/tts': ['tts.update'],
  '/novel-panel': ['novel.selection.replace']
};

function text(value, limit = 240) {
  return String(value ?? '').trim().slice(0, limit);
}

function makeRequestId() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  return `cm-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

function inferredCapabilities(pagePath) {
  return defaultCapabilities[text(pagePath, 240)] || [];
}

export function normalizeCmSelection(value) {
  if (!value || typeof value !== 'object') return null;
  const type = text(value.type, 48);
  const id = text(value.id, 160);
  const label = text(value.label || value.name, 160);
  if (!type && !id && !label) return null;
  return {
    type,
    id,
    label,
    meta: value.meta && typeof value.meta === 'object' ? value.meta : {}
  };
}

export function normalizeCmCapabilities(value) {
  if (!Array.isArray(value)) return [];
  return [...new Set(value.map(item => text(item, 80)).filter(item => actionTypes.has(item)))].slice(0, 40);
}

export function normalizeCmBridgeContext(value = {}) {
  const source = value && typeof value === 'object' ? value : {};
  const pagePath = text(source.pagePath, 240);
  const explicitCapabilities = normalizeCmCapabilities(source.capabilities);
  return {
    page: text(source.page, 120),
    pagePath,
    project: source.project && typeof source.project === 'object' ? source.project : null,
    selection: normalizeCmSelection(source.selection),
    capabilities: explicitCapabilities.length ? explicitCapabilities : inferredCapabilities(pagePath),
    summary: text(source.summary, 1200),
    canApply: source.canApply === true
  };
}

export function dispatchCmBridgeContext(context) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(CM_BRIDGE_CONTEXT_EVENT, {
    detail: normalizeCmBridgeContext(context)
  }));
}

export function dispatchCmSelection(selection) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(CM_BRIDGE_SELECTION_EVENT, {
    detail: normalizeCmSelection(selection)
  }));
}

export function normalizeCmAction(action) {
  if (!action || typeof action !== 'object') return null;
  const type = text(action.type, 80);
  if (!actionTypes.has(type)) return null;
  return {
    type,
    targetId: text(action.targetId, 160),
    label: text(action.label, 160),
    patch: action.patch && typeof action.patch === 'object' ? action.patch : {},
    payload: action.payload && typeof action.payload === 'object' ? action.payload : {}
  };
}

export function dispatchCmActionResult(requestId, result = {}) {
  if (typeof window === 'undefined' || !requestId) return;
  window.dispatchEvent(new CustomEvent(CM_BRIDGE_ACTION_RESULT_EVENT, {
    detail: {
      requestId: text(requestId, 160),
      ok: result.ok === true,
      message: text(result.message, 300),
      undoToken: text(result.undoToken, 240)
    }
  }));
}

export function refreshCmBridgeContext() {
  if (!activeBridge) return;
  let context = {};
  try {
    context = typeof activeBridge.getContext === 'function' ? activeBridge.getContext() || {} : {};
  } catch {
    context = {};
  }
  dispatchCmBridgeContext({
    ...context,
    page: context.page || activeBridge.page,
    pagePath: context.pagePath || activeBridge.pagePath,
    capabilities: activeBridge.capabilities,
    canApply: typeof activeBridge.apply === 'function'
  });
}

export function registerCmBridge(options = {}) {
  const bridge = {
    page: text(options.page, 120),
    pagePath: text(options.pagePath, 240),
    capabilities: normalizeCmCapabilities(options.capabilities),
    getContext: typeof options.getContext === 'function' ? options.getContext : null,
    apply: typeof options.apply === 'function' ? options.apply : null,
    undo: typeof options.undo === 'function' ? options.undo : null
  };
  activeBridge = bridge;
  refreshCmBridgeContext();

  return () => {
    if (activeBridge !== bridge) return;
    activeBridge = null;
    dispatchCmBridgeContext({ page: bridge.page, pagePath: bridge.pagePath, capabilities: inferredCapabilities(bridge.pagePath), canApply: false });
    Promise.resolve().then(() => {
      if (!activeBridge || activeBridge.pagePath !== bridge.pagePath) dispatchCmSelection(null);
    });
  };
}

export function dispatchCmAction(action, meta = {}) {
  if (typeof window === 'undefined') return '';
  const normalized = normalizeCmAction(action);
  if (!normalized) return '';
  const requestId = makeRequestId();
  const detail = {
    requestId,
    action: normalized,
    meta: meta && typeof meta === 'object' ? meta : {}
  };
  window.dispatchEvent(new CustomEvent(CM_BRIDGE_ACTION_EVENT, { detail }));

  const bridge = activeBridge;
  if (!bridge || typeof bridge.apply !== 'function' || !bridge.capabilities.includes(normalized.type)) return requestId;

  actionTail = actionTail
    .catch(() => undefined)
    .then(() => {
      if (activeBridge !== bridge) throw new Error('页面已经切换，本次修改未应用。');
      return bridge.apply(normalized, detail.meta);
    })
    .then(result => {
      dispatchCmActionResult(requestId, {
        ok: result?.ok !== false,
        message: result?.message || '修改已应用。',
        undoToken: result?.undoToken || ''
      });
      refreshCmBridgeContext();
    })
    .catch(error => {
      dispatchCmActionResult(requestId, { ok: false, message: error?.message || '修改应用失败。' });
    });

  return requestId;
}

export function dispatchCmUndo(undoToken) {
  if (typeof window === 'undefined' || !undoToken) return;
  const token = text(undoToken, 240);
  window.dispatchEvent(new CustomEvent(CM_BRIDGE_UNDO_EVENT, { detail: { undoToken: token } }));
  const bridge = activeBridge;
  if (!bridge || typeof bridge.undo !== 'function') return;
  actionTail = actionTail
    .catch(() => undefined)
    .then(() => {
      if (activeBridge !== bridge) return undefined;
      return bridge.undo(token);
    })
    .then(() => refreshCmBridgeContext())
    .catch(() => undefined);
}
