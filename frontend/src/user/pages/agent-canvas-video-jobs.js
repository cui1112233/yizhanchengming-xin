const runningStates = new Set(['preparing', 'queued', 'running', 'result_ready']);

export function canSubmitVideoFromConfig(config = {}) {
  return Boolean(String(config?.prompt || '').trim());
}

export function collectVideoReferences({ nodes = [], edges = [], videoNodeId = '' } = {}) {
  const byID = new Map(nodes.map(node => [node.id, node]));
  const seen = new Set();
  return edges.flatMap(edge => {
    if (edge.target !== videoNodeId || seen.has(edge.source)) return [];
    const node = byID.get(edge.source);
    const asset = node?.data?.assetRef;
    if (node?.data?.type !== 'image' || node?.data?.status !== 'confirmed' || !asset?.id || asset.kind !== 'image') return [];
    seen.add(edge.source);
    return [asset];
  });
}

export function videoReferenceAssetIDs(config = {}) {
  const seen = new Set();
  return (Array.isArray(config?.referenceAssets) ? config.referenceAssets : []).flatMap(asset => {
    const id = String(asset?.id || '').trim();
    if (asset?.kind !== 'image' || !id || seen.has(id)) return [];
    seen.add(id);
    return [id];
  });
}

export function videoNodeViewState(job) {
  const state = String(job?.state || 'draft');
  return {
    state,
    polling: runningStates.has(state),
    canGenerate: !job || ['failed', 'canceled', 'abandoned', 'applied'].includes(state),
    canCancel: ['preparing', 'queued', 'running'].includes(state),
    canRetry: ['failed', 'canceled', 'abandoned'].includes(state),
    canConfirm: state === 'awaiting_confirmation',
    waitingForConfirmation: state === 'awaiting_confirmation'
  };
}

export function makeVideoJobIdempotencyKey() {
  const random = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  return `agent-video-${random}`;
}
