const BOOK_GROUP_PREFIX = 'book:';

export function bookGroupId(bookId) {
  const value = String(bookId || '').trim();
  return value ? `${BOOK_GROUP_PREFIX}${value}` : '';
}

export function bookIdFromGroupId(groupId) {
  const value = String(groupId || '').trim();
  return value.startsWith(BOOK_GROUP_PREFIX) ? value.slice(BOOK_GROUP_PREFIX.length) : '';
}

export function agentCanvasURL({ projectId, batchId, bookId } = {}) {
  const query = new URLSearchParams();
  query.set('taskId', String(projectId || '').trim());
  query.set('batchId', String(batchId || '').trim());
  const requestedBookId = String(bookId || '').trim();
  if (requestedBookId) query.set('bookId', requestedBookId);
  query.set('view', 'workflow');
  return `/agent/canvas?${query.toString()}`;
}

export async function bindBatchFactoryBookAndNavigate({ batchId, bookId, bind, navigate }) {
  const binding = await bind(batchId);
  const projectId = String(binding?.projectId || '').trim();
  if (!projectId) throw new Error('Agent 项目绑定失败');
  const url = agentCanvasURL({ projectId, batchId, bookId });
  navigate(url);
  return binding;
}

export async function syncBookThenLoadCanvas({ taskId, bookId, sync, getCanvas, createSaveQueue }) {
  const requestedBookId = String(bookId || '').trim();
  if (requestedBookId) await sync(taskId, requestedBookId);
  const result = await getCanvas(taskId);
  return {
    result,
    saveQueue: createSaveQueue?.(result?.canvas || { version: 1, revision: 0, nodes: [], edges: [] })
  };
}

// A mapped book may change in Batch Factory while this tab is in the
// background. Flush only the user's local layout edits, then reload the
// authoritative Batch Factory projection before applying it to the canvas.
export async function refreshBatchFactoryBookCanvas({ taskId, bookId, queue, sync, getCanvas, isCurrent = () => true }) {
  const requestedTaskId = String(taskId || '').trim();
  const requestedBookId = String(bookId || '').trim();
  if (!requestedTaskId || !requestedBookId || typeof sync !== 'function' || typeof getCanvas !== 'function') {
    throw new Error('批量工厂画布同步参数无效');
  }
  await queue?.flushRequired?.();
  if (!isCurrent()) return { ignored: true };
  await sync(requestedTaskId, requestedBookId);
  if (!isCurrent()) return { ignored: true };
  const result = await getCanvas(requestedTaskId);
  return isCurrent() ? result : { ...result, ignored: true };
}

export function focusRequestedBookGroup({ bookId, nodes, focus }) {
  const groupId = bookGroupId(bookId);
  const groupNodes = (Array.isArray(nodes) ? nodes : []).filter(node => String((node?.groupId ?? node?.data?.groupId) || '') === groupId);
  if (groupNodes.length) focus?.(groupNodes);
  return groupNodes;
}

export async function removeBatchFactoryCanvasNodes({ taskId, nodeIds, flush, remove, refresh, isCurrent = () => true }) {
  const ids = [...new Set((nodeIds || []).map(value => String(value || '').trim()).filter(Boolean))];
  if (!taskId || !ids.length) return;
  await flush?.();
  if (!isCurrent()) return;
  for (const nodeId of ids) {
    if (!isCurrent()) return;
    await remove(taskId, nodeId);
  }
  if (!isCurrent()) return;
  await refresh?.();
}

export function isBatchFactoryNode(node) {
  return Boolean(bookIdFromGroupId(node?.groupId ?? node?.data?.groupId));
}

export function projectedNodePresentation(node) {
  const groupId = String(node?.groupId ?? node?.data?.groupId ?? '').trim();
  if (!bookIdFromGroupId(groupId)) return { groupId, groupLabel: '', sourceLabel: '' };
  const groupLabel = {
    script: '书籍正文',
    text: '书籍文本',
    image: '书籍素材',
    video: '书籍视频',
    audio: '书籍音频'
  }[String(node?.type ?? node?.data?.type ?? '').trim()] || '书籍内容';
  return { groupId, groupLabel, sourceLabel: '批量工厂' };
}

export function batchFactoryGroupLabel(groupId, nodes = []) {
  const bookId = bookIdFromGroupId(groupId);
  if (!bookId) return String(groupId || '').trim();
  const scriptNode = (Array.isArray(nodes) ? nodes : []).find(node => {
    const nodeGroupId = String(node?.groupId ?? node?.data?.groupId ?? '').trim();
    const nodeType = String(node?.data?.type ?? node?.type ?? '').trim();
    return nodeGroupId === groupId && nodeType === 'script';
  });
  const title = String(scriptNode?.title ?? scriptNode?.data?.title ?? '').trim();
  return `书籍分组 ${title || bookId}`;
}

export function videoNodeTiming({ audioSeconds, selectedSeconds, allowedSeconds } = {}) {
  const audio = Number(audioSeconds) || 0;
  const selected = Number(selectedSeconds) || 0;
  const allowed = [...new Set((Array.isArray(allowedSeconds) ? allowedSeconds : []).map(Number).filter(value => value > 0))].sort((left, right) => left - right);
  if (selected > 0) {
    const mismatch = audio > 0 && Math.abs(selected - audio) > 0.01;
    return {
      requestedSeconds: selected,
      adjusted: false,
      warning: mismatch ? `所选视频时长 ${selected} 秒与已连接音频时长不一致（${audio} 秒）` : ''
    };
  }
  if (audio <= 0 || !allowed.length) return { requestedSeconds: audio, adjusted: false, warning: '' };
  const requestedSeconds = allowed.reduce((nearest, value) => (
    Math.abs(value - audio) < Math.abs(nearest - audio) ? value : nearest
  ), allowed[0]);
  return { requestedSeconds, adjusted: Math.abs(requestedSeconds - audio) > 0.01, warning: '' };
}

export function mappedNodeStatePresentation(node) {
  if (!isBatchFactoryNode(node)) return null;
  const config = node?.config || node?.data?.config || {};
  if (String(config.resourceType || '').trim() === 'merge' && String(config.mergeState || '').trim()) {
    const ready = String(config.mergeState).trim() === 'ready';
    return {
      state: ready ? 'ready' : 'blocked',
      errorMessage: '',
      canRetry: false,
      busy: false,
      message: ready ? '所有分镜已就绪，可以合成最终视频' : String(config.mergeReason || '等待所有分镜视频生成完成')
    };
  }
  if (String(config.resourceType || '').trim() === 'audio' && config.audioPlanPending === true) {
    return { state: 'draft', errorMessage: '', canRetry: false, busy: false, message: '待生成配音' };
  }
  const state = ['queued', 'running', 'succeeded', 'failed', 'cancelled'].includes(String(config.mappedState || node?.status || '').trim())
    ? String(config.mappedState || node?.status).trim()
    : 'queued';
  const errorMessage = String(config.errorMessage || '').trim();
  const message = errorMessage || ({
    queued: '正在排队', running: '正在执行中…', succeeded: '已完成', failed: '批量工厂任务失败，请重试', cancelled: '已取消'
  }[state] || '正在排队');
  return { state, errorMessage, canRetry: config.canRetry === true, busy: state === 'queued' || state === 'running', message };
}

// Mapped production remains owned by Batch Factory. The canvas only polls a
// book while one of its projected nodes is still in a non-terminal state.
export function mappedBatchFactoryBooksNeedingRefresh(nodes) {
  const books = new Map();
  for (const node of Array.isArray(nodes) ? nodes : []) {
    const config = node?.data?.config || node?.config || {};
    if (String(config.source || '').trim() !== 'batch_factory') continue;
    const state = String(config.mappedState || node?.data?.status || node?.status || '').trim().toLowerCase();
    if (state !== 'queued' && state !== 'running') continue;
    const batchId = String(config.batchId || '').trim();
    const bookId = String(config.bookId || '').trim();
    if (!batchId || !bookId) continue;
    books.set(`${batchId}\u0000${bookId}`, { batchId, bookId });
  }
  return [...books.values()];
}

export function batchFactoryMappedAction(node) {
  if (!isBatchFactoryNode(node)) return null;
  const config = node?.config || node?.data?.config || {};
  const batchId = String(config.batchId || '').trim();
  const bookId = String(config.bookId || '').trim();
  const resourceType = String(config.resourceType || '').trim();
  const resourceId = String(config.resourceId || '').trim();
  if (!batchId || !bookId || !resourceType) return null;
  if (resourceType === 'book_script') return { type: 'script', batchId, bookId, resourceId };
  if (resourceType === 'book_asset' && resourceId) return { type: 'image', batchId, bookId, resourceId };
  if (resourceType === 'video') {
    const timing = videoNodeTiming({
      audioSeconds: config.connectedAudioDurationSeconds,
      selectedSeconds: config.durationOverrideSeconds,
      allowedSeconds: config.allowedDurationSeconds
    });
    const resolvedSeconds = Number(config.resolvedDurationSeconds) || timing.requestedSeconds;
    return {
      type: 'video', batchId, bookId, resourceId,
      ...(String(config.videoProvider || '').trim() ? { provider: String(config.videoProvider).trim() } : {}),
      ...(String(config.videoModelId || '').trim() ? { modelId: String(config.videoModelId).trim() } : {}),
      ...(config.canRetry === true ? { canRetry: true } : {}),
      ...(resolvedSeconds > 0 ? { durationSeconds: resolvedSeconds } : {}),
      ...(Number(config.resourceRevision) > 0 ? { expectedRevision: Number(config.resourceRevision) } : {})
    };
  }
  if (resourceType === 'merge') return {
    type: 'merge', batchId, bookId, resourceId,
    canMerge: config.canMerge === true,
    canPublish: config.canPublish === true,
    mergeState: String(config.mergeState || '').trim(),
    mergeReason: String(config.mergeReason || '').trim()
  };
  if (resourceType === 'audio' && resourceId) return { type: 'audio', batchId, bookId, resourceId };
  return { type: 'sync', batchId, bookId, resourceId };
}

// Settings and metadata save responses differ, but both carry the next
// book revision needed by any following save.
export function latestBookRevision(result, fallback = 0) {
  const direct = Number(result?.revision);
  if (Number.isFinite(direct) && direct > 0) return direct;
  const bookRevision = Number(result?.book?.revision);
  if (Number.isFinite(bookRevision) && bookRevision > 0) return bookRevision;
  return Number(fallback) || 0;
}

export async function runMappedBatchFactoryScriptStage({ action, stage, requestId, options = {}, runStage }) {
  if (action?.type !== 'script' || !action.batchId || !action.bookId || !['assets', 'director'].includes(stage) || typeof runStage !== 'function') {
    throw new Error('映射剧本生产参数无效');
  }
  const contentRangeLines = Math.min(500, Math.max(1, Number(options.contentRangeLines) || 5));
  return runStage(action.batchId, action.bookId, stage, { mode: 'missing', requestId, contentRangeLines });
}

export async function runMappedBatchFactoryVideoProduction({ action, requestId, runStage }) {
  if (action?.type !== 'video' || !action.batchId || !action.bookId || !action.resourceId || typeof runStage !== 'function') {
    throw new Error('映射视频生产参数无效');
  }
  return runStage(action.batchId, action.bookId, 'video', {
    mode: action.canRetry ? 'force' : 'missing',
    requestId,
    videoId: action.resourceId,
    ...(action.provider ? { provider: action.provider } : {}),
    ...(action.modelId ? { videoModelId: action.modelId } : {})
  });
}

export async function runMappedBatchFactoryAudioPlan({ action, requestId, createAudioPlan }) {
  if (action?.type !== 'audio' || !action.batchId || !action.bookId || typeof createAudioPlan !== 'function') {
    throw new Error('映射音频生产参数无效');
  }
  return createAudioPlan(action.batchId, action.bookId, { requestId });
}
