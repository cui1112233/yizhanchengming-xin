const PRESENTATIONS = Object.freeze({
  draft: { label: '未生成', primaryAction: '开始生成', action: 'generate', tone: 'idle' },
  queued: { label: '正在排队', primaryAction: '取消', action: 'cancel', tone: 'running' },
  running: { label: '正在生成…', primaryAction: '取消', action: 'cancel', tone: 'running' },
  awaiting_confirmation: { label: '等待确认', primaryAction: '确认使用', action: 'confirm', tone: 'review' },
  completed: { label: '已完成', primaryAction: '', action: 'none', tone: 'success' },
  failed: { label: '生成失败', primaryAction: '重新生成', action: 'retry', tone: 'error' },
  cancelled: { label: '已取消', primaryAction: '重新生成', action: 'retry', tone: 'idle' }
});

function taskUpdatedAt(task) {
  const timestamp = Date.parse(String(task?.updatedAt || ''));
  return Number.isFinite(timestamp) ? timestamp : 0;
}

function taskAttemptAt(task) {
  const timestamp = Date.parse(String(task?.createdAt || ''));
  return Number.isFinite(timestamp) ? timestamp : taskUpdatedAt(task);
}

export function activeGenerationTaskForNode(tasks = [], nodeId = '', kind = '') {
  return latestGenerationTaskForNode(tasks, nodeId, kind, shouldPollGeneration);
}

export function latestGenerationTaskForNode(tasks = [], nodeId = '', kind = '', predicate = () => true) {
  const targetNode = String(nodeId || '').trim();
  const targetKind = String(kind || '').trim();
  return (Array.isArray(tasks) ? tasks : [])
    .filter(task => String(task?.nodeId || '').trim() === targetNode
      && (!targetKind || String(task?.kind || '').trim() === targetKind)
      && predicate(task))
    .sort((left, right) => taskUpdatedAt(right) - taskUpdatedAt(left))[0] || null;
}

function resultOption(result, fallbackKind = '') {
  const asset = result?.asset && typeof result.asset === 'object' ? { ...result.asset } : {};
  const id = String(result?.id || '').trim();
  const assetId = String(result?.assetId || asset.id || '').trim();
  const kind = String(result?.kind || asset.kind || fallbackKind).trim();
  return {
    ...asset,
    ...result,
    id,
    assetId,
    kind,
    asset: { ...asset, ...(assetId ? { id: assetId, assetId } : {}), ...(kind ? { kind } : {}) }
  };
}

export function generationPresentation(task = {}) {
  const state = String(task?.state || 'draft');
  const presentation = PRESENTATIONS[state] || PRESENTATIONS.draft;
  const message = state === 'failed' && String(task?.errorMessage || '').trim()
    ? String(task.errorMessage).trim()
    : presentation.label;
  return {
    state,
    ...presentation,
    message,
    busy: state === 'queued' || state === 'running',
    polling: shouldPollGeneration(task),
    canCancel: state === 'queued' || state === 'running',
    canConfirm: state === 'awaiting_confirmation',
    canRetry: state === 'failed' || state === 'cancelled'
  };
}

export function generationTaskActions(task = {}) {
  const presentation = generationPresentation(task);
  const legacyVideo = task?.kind === 'video' && !String(task?.mode || '').startsWith('video_');
  const needsCanvas = legacyVideo || presentation.canConfirm;
  return {
    canCancel: presentation.canCancel && !legacyVideo,
    canRetry: presentation.canRetry && !legacyVideo,
    openCanvasOnly: needsCanvas
  };
}

export function shouldPollGeneration(task) {
  return task?.state === 'queued' || task?.state === 'running';
}

export function generationSummaryText(summary = {}) {
  const count = key => Math.max(0, Number(summary?.[key]) || 0);
  const parts = [
    [count('queued'), '排队'],
    [count('running'), '生成中'],
    [count('awaitingConfirmation'), '待确认']
  ].filter(([value]) => value > 0)
    .map(([value, label]) => `${label} ${value}`);
  return parts.length ? parts.join(' · ') : '暂无进行中任务';
}

export function mergeGenerationTask(node, task, preferredResultId = '') {
  const nodeId = String(node?.id || '').trim();
  const nodeKind = String(node?.data?.type || node?.type || '').trim();
  const nodeVersion = Number(node?.data?.version ?? node?.version) || 0;
  const matches = task
    && String(task.nodeId || '').trim() === nodeId
    && String(task.kind || '').trim() === nodeKind
    && (!nodeVersion || Number(task.nodeVersion) === nodeVersion);
  if (!matches) {
    return { task: null, resultOptions: [], selectedResultId: '', previewAssetId: '', previewAsset: null };
  }

  const resultOptions = (Array.isArray(task.results) ? task.results : [])
    .map(result => resultOption(result, task.kind))
    .filter(result => result.id && result.assetId);
  const requested = String(preferredResultId || task.selectedResultId || '').trim();
  const selected = resultOptions.find(result => result.id === requested || result.assetId === requested) || resultOptions[0] || null;
  return {
    task,
    resultOptions,
    selectedResultId: selected?.id || '',
    previewAssetId: selected?.assetId || '',
    previewAsset: selected?.asset || null
  };
}

export function mergeLatestGenerationTasksByNode(current = {}, tasks = [], projectId = '') {
  const activeProjectId = String(projectId || '').trim();
  const next = {};
  const candidates = [
    ...Object.values(current && typeof current === 'object' ? current : {}),
    ...(Array.isArray(tasks) ? tasks : [])
  ];
  for (const task of candidates) {
    const nodeId = String(task?.nodeId || '').trim();
    if (!nodeId || String(task?.projectId || '').trim() !== activeProjectId) continue;
    const existing = next[nodeId];
    // A late error from an older request must not replace the user's newer
    // running attempt for the same node.
    const incomingTime = taskAttemptAt(task);
    const existingTime = taskAttemptAt(existing);
    const sameTaskUpdated = existing?.id === task?.id && taskUpdatedAt(task) > taskUpdatedAt(existing);
    if (!existing || incomingTime > existingTime || sameTaskUpdated || (incomingTime === existingTime && String(task?.id || '') > String(existing?.id || ''))) {
      next[nodeId] = task;
    }
  }
  return next;
}

export function shouldAcceptGenerationResponse({ requestedTaskId, currentTaskId, active = true } = {}) {
  const requested = String(requestedTaskId || '').trim();
  return Boolean(active && requested && requested === String(currentTaskId || '').trim());
}
