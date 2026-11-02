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

export function viewportHasChanged(current, next) {
  return Number(current?.x) !== Number(next?.x)
    || Number(current?.y) !== Number(next?.y)
    || Number(current?.zoom) !== Number(next?.zoom);
}

export function shouldKeepPendingCreatedTask({ requestedTaskId, activeTaskId, knownTaskIds }) {
  return Boolean(requestedTaskId && requestedTaskId === activeTaskId && !new Set(knownTaskIds || []).has(requestedTaskId));
}

export function shouldAcceptCanvasRefresh({ requestedTaskId, currentTaskId, requestedQueue, currentQueue, mounted = true }) {
  return Boolean(mounted && requestedTaskId && requestedTaskId === currentTaskId && requestedQueue === currentQueue);
}

export function retainCanvasEditorNodeID(nodeID, nodes = []) {
  const current = String(nodeID || '');
  return current && nodes.some(node => node?.id === current) ? current : '';
}

export function retainCanvasAssetPanel(panel, nodes = []) {
  if (!panel) return null;
  const nodeID = String(panel.nodeID || panel.replaceNodeId || '');
  return nodeID && nodes.some(node => node?.id === nodeID) ? panel : null;
}

export function mergePendingCanvas(serverCanvas, localCanvas) {
  const serverNodes = new Map((serverCanvas?.nodes || []).map(node => [node.id, node]));
  const localNodes = Array.isArray(localCanvas?.nodes) ? localCanvas.nodes : [];
  const localEdges = Array.isArray(localCanvas?.edges) ? localCanvas.edges : [];
  const edgeKey = edge => String(edge?.id || `${edge?.source || ''}:${edge?.target || ''}:${edge?.data?.relation || ''}`);
  const localEdgesByKey = new Map(localEdges.map(edge => [edgeKey(edge), edge]));
  const isBatchFactoryMappedNode = node => String(node?.config?.source || '') === 'batch_factory';
  const isBatchFactoryMappedEdge = edge => isBatchFactoryMappedNode(serverNodes.get(edge?.source)) || isBatchFactoryMappedNode(serverNodes.get(edge?.target));
  const retainedLocalNodes = localNodes.filter(node => node?.id && (serverNodes.has(node.id) || !isBatchFactoryMappedNode(node)));
  const mergedNodes = [
    ...(serverCanvas?.nodes || []).map(serverNode => {
      const localNode = localNodes.find(node => node?.id === serverNode.id);
      if (!localNode) return serverNode;
      if (isBatchFactoryMappedNode(serverNode) || isBatchFactoryMappedNode(localNode)) {
        // The production record lives in Batch Factory. A pending browser
        // draft may keep a just-dragged position, but can never replace the
        // source record's state, prompt, version, or media reference.
        return { ...serverNode, ...(localNode.position ? { position: localNode.position } : {}) };
      }
      return {
        ...serverNode,
        ...localNode,
        version: serverNode.version,
        status: serverNode.status,
        config: { ...(serverNode.config || {}), ...(localNode.config || {}) },
        ...(serverNode.assetRef ? { assetRef: serverNode.assetRef } : {})
      };
    }),
    ...retainedLocalNodes.filter(node => !serverNodes.has(node.id))
  ];
  const mergedNodeIDs = new Set(mergedNodes.map(node => node.id));
  const serverEdgeKeys = new Set((serverCanvas?.edges || []).map(edgeKey));
  const mergedEdges = [
    ...(serverCanvas?.edges || []).map(serverEdge => {
      const localEdge = localEdgesByKey.get(edgeKey(serverEdge));
      if (isBatchFactoryMappedEdge(serverEdge)) return serverEdge;
      return localEdge ? { ...serverEdge, ...localEdge, data: { ...(serverEdge.data || {}), ...(localEdge.data || {}) } } : serverEdge;
    }),
    ...localEdges.filter(edge => !serverEdgeKeys.has(edgeKey(edge)) && mergedNodeIDs.has(edge?.source) && mergedNodeIDs.has(edge?.target))
  ];
  return {
    ...serverCanvas,
    ...localCanvas,
    revision: Number(serverCanvas?.revision) || 0,
    nodes: mergedNodes,
    edges: mergedEdges
  };
}

export function createCanvasSaveQueue({ revision = 0, delay = 1200, save, reload, onCanvas, onStatus, onSynced }) {
  let currentRevision = Number(revision) || 0;
  let pending = null;
  let failed = null;
  let active = null;
  let disposed = false;
  let timer = null;
  let conflictRecoveryAttempted = false;
  let lastSaveError = null;

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
        lastSaveError = null;
        conflictRecoveryAttempted = false;
        const savedCanvas = result?.canvas || { ...canvas, revision: currentRevision + 1 };
        currentRevision = Number(savedCanvas.revision) || currentRevision + 1;
        if (!pending) {
          onCanvas?.(savedCanvas);
          onSynced?.(savedCanvas);
          setStatus('saved');
        }
      } catch (error) {
        if (disposed) return;
        lastSaveError = error;
        if (error?.status === 409) {
          failed = null;
          if (conflictRecoveryAttempted) {
            pending = null;
            try {
              const result = await reload();
              if (disposed) return;
              const fresh = result?.canvas;
              if (fresh) {
                currentRevision = Number(fresh.revision) || currentRevision;
                onCanvas?.(fresh);
              }
              setStatus('conflict');
            } catch (reloadError) {
              lastSaveError = reloadError;
              failed = canvas;
              setStatus('error');
            }
            return;
          }
          conflictRecoveryAttempted = true;
          try {
            const result = await reload();
            if (disposed) return;
            const fresh = result?.canvas;
            if (fresh) {
              currentRevision = Number(fresh.revision) || 0;
              pending = mergePendingCanvas(fresh, canvas);
              onCanvas?.(pending);
              setStatus('dirty');
              return;
            }
          } catch (reloadError) {
            lastSaveError = reloadError;
            failed = canvas;
            setStatus('error');
          } finally {
            if (!pending && !failed) setStatus('conflict');
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
    if (pending) schedule();
  }

  function clearTimer() {
    if (timer) clearTimeout(timer);
    timer = null;
  }

  return {
    enqueue(canvas) {
      if (disposed) return;
      pending = canvas;
      failed = null;
      lastSaveError = null;
      conflictRecoveryAttempted = false;
      setStatus('dirty');
      clearTimer();
      schedule();
    },
    async flush() {
      clearTimer();
      if (active) await active;
      clearTimer();
      return savePending();
    },
    async flushRequired() {
      if (disposed) throw new Error('画布保存队列已关闭');
      clearTimer();
      while (active || pending) {
        await savePending();
        clearTimer();
      }
      if (failed || lastSaveError) throw lastSaveError || new Error('画布保存失败');
      return { revision: currentRevision };
    },
    retry() {
      if (disposed || !failed) return Promise.resolve();
      pending = failed;
      failed = null;
      lastSaveError = null;
      conflictRecoveryAttempted = false;
      setStatus('dirty');
      clearTimer();
      return savePending();
    },
    currentRevision() {
      return currentRevision;
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
      lastSaveError = null;
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
      lastSaveError = null;
      clearTimer();
    }
  };
}
