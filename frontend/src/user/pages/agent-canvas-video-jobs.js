const runningStates = new Set(['preparing', 'submitting', 'queued', 'running', 'result_ready', 'ingesting', 'canceling']);

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
  const references = Array.isArray(config?.referenceAssets) ? config.referenceAssets : [];
  const unsupportedVideo = references
    .find(asset => asset?.kind === 'video' && String(asset?.id || '').trim());
  if (unsupportedVideo) return '当前视频生成暂不支持视频参考素材，请移除视频参考后再生成。';
  const unsupportedAudio = references
    .find(asset => asset?.kind === 'audio' && String(asset?.id || '').trim());
  return unsupportedAudio ? '当前视频生成暂不支持音频参考素材，请移除音频参考后再生成。' : '';
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
  const defaultLabel = {
    draft: '准备生成',
    preparing: '正在准备',
    submitting: '正在提交',
    queued: '正在排队',
    running: '正在生成视频…',
    result_ready: '正在保存结果',
    ingesting: '正在保存结果',
    canceling: '正在取消',
    awaiting_confirmation: '等待确认',
    failed: '生成失败',
    canceled: '已取消',
    abandoned: '需处理',
    applied: '已完成'
  }[state] || '准备生成';
  const failureReason = state === 'failed' ? String(job?.errorMessage || '').trim() : '';
  const label = failureReason || defaultLabel;
  return {
    state, label, message: label,
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
    submitting: 1,
    queued: 2,
    running: 3,
    result_ready: 4,
    ingesting: 5,
    canceling: 5,
    awaiting_confirmation: 6,
    failed: 7,
    canceled: 7,
    abandoned: 7,
    applied: 8
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
