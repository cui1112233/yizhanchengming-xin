const runningStates = new Set(['preparing', 'queued', 'running', 'result_ready', 'ingesting', 'canceling']);

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
    if (!['image', 'video'].includes(asset?.kind) || !id || seen.has(id)) return [];
    seen.add(id);
    return [id];
  });
}

export function videoReferenceCapabilityError(config = {}) {
  const unsupported = (Array.isArray(config?.referenceAssets) ? config.referenceAssets : [])
    .find(asset => asset?.kind === 'video' && String(asset?.id || '').trim());
  return unsupported ? '当前视频生成暂不支持视频参考素材，请移除视频参考后再生成。' : '';
}

export function videoResultPreview(job) {
  const asset = job?.resultAsset;
  const id = String(asset?.id || '').trim();
  const url = String(asset?.url || '').trim();
  if (asset?.kind !== 'video' || !id || !url) return null;
  const name = String(asset?.name || '').trim();
  return { id, kind: 'video', ...(name ? { name } : {}), url };
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

function videoJobUpdatedAt(job) {
  const timestamp = Date.parse(String(job?.updatedAt || ''));
  return Number.isFinite(timestamp) ? timestamp : 0;
}

function videoJobStateRank(job) {
  return {
    preparing: 0,
    queued: 1,
    running: 2,
    result_ready: 3,
    ingesting: 4,
    canceling: 4,
    awaiting_confirmation: 5,
    failed: 6,
    canceled: 6,
    abandoned: 6,
    applied: 7
  }[String(job?.state || '')] ?? -1;
}

export function mergeLatestVideoJobsByNode(current = {}, jobs = [], projectId = '') {
  const activeProjectId = String(projectId || '').trim();
  const next = {};
  const candidates = [
    ...Object.values(current && typeof current === 'object' ? current : {}),
    ...(Array.isArray(jobs) ? jobs : [])
  ];
  for (const job of candidates) {
    const nodeId = String(job?.nodeId || '').trim();
    if (!nodeId || String(job?.projectId || '').trim() !== activeProjectId) continue;
    const existing = next[nodeId];
    const incomingTime = videoJobUpdatedAt(job);
    const existingTime = videoJobUpdatedAt(existing);
    const sameJobProgressed = existing?.id === job?.id
      && incomingTime === existingTime
      && videoJobStateRank(job) >= videoJobStateRank(existing);
    if (!existing || incomingTime > existingTime || sameJobProgressed) next[nodeId] = job;
  }
  return next;
}

export function shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId, active = true } = {}) {
  const requested = String(requestedTaskId || '').trim();
  return Boolean(active && requested && requested === String(currentTaskId || '').trim());
}

export function makeVideoJobIdempotencyKey() {
  const random = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  return `agent-video-${random}`;
}
