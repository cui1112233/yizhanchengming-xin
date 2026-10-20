const publicAssetFields = ['assetId', 'id', 'kind', 'mediaType', 'name', 'width', 'height', 'duration'];

function publicAssetRef(value) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined;
  const output = {};
  for (const field of publicAssetFields) {
    if (typeof value[field] === 'string' || Number.isFinite(value[field])) output[field] = value[field];
  }
  return Object.keys(output).length ? output : undefined;
}

export function buildSelectedNodeContext(canvas, selectedIds) {
  const selected = new Set(Array.isArray(selectedIds) ? selectedIds : []);
  return (Array.isArray(canvas?.nodes) ? canvas.nodes : [])
    .filter(node => selected.has(node.id))
    .map(node => ({
      id: String(node.id || '').slice(0, 96),
      type: String(node.type || '').slice(0, 64),
      title: String(node.title || '').slice(0, 191),
      summary: String(node.summary || '').slice(0, 1600),
      version: Number.isFinite(Number(node.version)) ? Number(node.version) : 0,
      status: String(node.status || '').slice(0, 32),
      ...(publicAssetRef(node.assetRef) ? { assetRef: publicAssetRef(node.assetRef) } : {})
    }));
}

export const canvasSaveStatusText = {
  idle: '已保存',
  dirty: '未保存',
  saving: '正在保存…',
  saved: '已保存',
  conflict: '保存冲突，已重新加载',
  error: '保存失败'
};

export function canvasSaveAction(status) {
  if (status === 'error') return { method: 'retry', label: '重试保存', disabled: false };
  if (status === 'dirty' || status === 'saving') return { method: 'flush', label: '保存', disabled: false };
  if (status === 'conflict') return { method: null, label: '已重新加载', disabled: true };
  return { method: null, label: '已保存', disabled: true };
}

export function shouldPersistFlowChanges(changes) {
  return Array.isArray(changes) && changes.some(change => change?.type !== 'select' && change?.type !== 'dimensions');
}

export function shouldKeepPendingCreatedTask({ requestedTaskId, activeTaskId, knownTaskIds }) {
  return Boolean(requestedTaskId && requestedTaskId === activeTaskId && !new Set(knownTaskIds || []).has(requestedTaskId));
}

export function shouldAcceptCanvasRefresh({ requestedTaskId, currentTaskId, requestedQueue, currentQueue, mounted = true }) {
  return Boolean(mounted && requestedTaskId && requestedTaskId === currentTaskId && requestedQueue === currentQueue);
}

export function createCanvasSaveQueue({ revision = 0, delay = 220, save, reload, onCanvas, onStatus }) {
  let currentRevision = Number(revision) || 0;
  let pending = null;
  let failed = null;
  let active = null;
  let disposed = false;
  let timer = null;

  function setStatus(status) {
    if (!disposed) onStatus?.(status);
  }

  function schedule(wait = delay) {
    if (disposed || active || !pending || timer) return;
    timer = setTimeout(() => {
      timer = null;
      void savePending();
    }, wait);
  }

  async function savePending() {
    if (disposed) return;
    if (active) {
      await active;
      if (pending) return savePending();
      return;
    }
    if (!pending) return;
    const canvas = pending;
    pending = null;
    setStatus('saving');
    active = (async () => {
      try {
        const result = await save({ ...canvas, revision: currentRevision }, currentRevision);
        if (disposed) return;
        failed = null;
        const savedCanvas = result?.canvas || { ...canvas, revision: currentRevision + 1 };
        currentRevision = Number(savedCanvas.revision) || currentRevision + 1;
        if (!pending) {
          onCanvas?.(savedCanvas);
          setStatus('saved');
        }
      } catch (error) {
        if (disposed) return;
        if (error?.status === 409) {
          pending = null;
          failed = null;
          try {
            const result = await reload();
            if (disposed) return;
            const fresh = result?.canvas;
            if (fresh) {
              currentRevision = Number(fresh.revision) || 0;
              onCanvas?.(fresh);
            }
          } finally {
            setStatus('conflict');
          }
        } else {
          failed = pending || canvas;
          pending = null;
          setStatus('error');
        }
      }
    })();
    await active;
    active = null;
    if (pending) return savePending();
  }

  function clearTimer() {
    if (timer) clearTimeout(timer);
    timer = null;
  }

  function mergePendingCanvas(serverCanvas, localCanvas) {
    const serverNodes = new Map((serverCanvas?.nodes || []).map(node => [node.id, node]));
    return {
      ...serverCanvas,
      ...localCanvas,
      revision: Number(serverCanvas?.revision) || 0,
      nodes: (localCanvas?.nodes || []).map(node => {
        const serverNode = serverNodes.get(node.id);
        if (!serverNode) return node;
        return {
          ...serverNode,
          ...node,
          version: serverNode.version,
          status: serverNode.status,
          ...(serverNode.assetRef ? { assetRef: serverNode.assetRef } : {})
        };
      })
    };
  }

  return {
    enqueue(canvas) {
      if (disposed) return;
      pending = canvas;
      failed = null;
      setStatus('dirty');
      schedule();
    },
    flush() {
      clearTimer();
      return savePending();
    },
    retry() {
      if (disposed || !failed) return Promise.resolve();
      pending = failed;
      failed = null;
      setStatus('dirty');
      clearTimer();
      return savePending();
    },
    async acceptServerCanvas(canvas) {
      if (disposed || !canvas) return false;
      if (active) await active;
      if (disposed) return false;
      const incomingRevision = Number(canvas.revision) || 0;
      if (incomingRevision < currentRevision) return false;
      clearTimer();
      currentRevision = incomingRevision;
      failed = null;
      if (pending) {
        pending = mergePendingCanvas(canvas, pending);
        onCanvas?.(pending);
        setStatus('dirty');
        schedule(0);
        return true;
      }
      onCanvas?.(canvas);
      setStatus('saved');
      return true;
    },
    dispose() {
      disposed = true;
      pending = null;
      failed = null;
      clearTimer();
    }
  };
}
