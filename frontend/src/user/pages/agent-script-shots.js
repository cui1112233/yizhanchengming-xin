const shotFields = ['scene', 'characters', 'dialogue', 'narration', 'imagePrompt', 'videoPrompt', 'ttsText'];

function text(value) {
  return String(value ?? '').trim();
}

function normalizeShot(shot, index) {
  const normalized = {
    ...shot,
    id: text(shot?.id),
    scriptNodeId: text(shot?.scriptNodeId),
    sequence: Number.isFinite(Number(shot?.sequence)) && Number(shot.sequence) > 0 ? Number(shot.sequence) : index + 1,
    revision: Number.isFinite(Number(shot?.revision)) ? Number(shot.revision) : 0
  };
  for (const field of shotFields) normalized[field] = text(shot?.[field]);
  return normalized;
}

export function normalizeScriptShots(response) {
  const source = response && typeof response === 'object' ? response : {};
  const generation = source.generation && typeof source.generation === 'object' ? source.generation : {};
  const split = source.split && typeof source.split === 'object' ? source.split : {};
  const rawShots = Array.isArray(split.shots) ? split.shots : Array.isArray(source.shots) ? source.shots : [];
  const shots = rawShots
    .map(normalizeShot)
    .filter(shot => shot.id)
    .sort((left, right) => left.sequence - right.sequence || left.id.localeCompare(right.id));
  const explicitState = text(generation.state || generation.status || split.state || source.state || source.status).toLowerCase();
  const errorMessage = text(generation.errorMessage || generation.error || split.errorMessage || split.error || source.errorMessage || source.error);
  const state = explicitState || (errorMessage ? 'failed' : shots.length ? 'completed' : 'idle');
  return {
    state,
    errorMessage,
    generationId: text(generation.id || source.generationId),
    shots
  };
}

export function scriptSourceMarker(value) {
  const source = String(value ?? '');
  let hash = 2166136261;
  for (let index = 0; index < source.length; index += 1) {
    hash ^= source.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return `${source.length}:${(hash >>> 0).toString(16).padStart(8, '0')}`;
}

export function invalidateScriptSplitForSource(config = {}, sourceText = '') {
  const nextSource = String(sourceText ?? '');
  if (nextSource === String(config.sourceText || '')) return { ...config, sourceText: nextSource };
  return {
    ...config,
    sourceText: nextSource,
    splitState: { state: 'stale', errorMessage: '', generationId: '', sourceMarker: '' }
  };
}

export function scriptSplitReady(config = {}) {
  const state = config.splitState && typeof config.splitState === 'object' ? config.splitState : {};
  const status = text(state.state).toLowerCase();
  return ['completed', 'done'].includes(status) &&
    text(state.sourceMarker) !== '' &&
    text(state.sourceMarker) === scriptSourceMarker(config.sourceText);
}

export function normalizeScriptShotsForConfig(config = {}, response) {
  const normalized = normalizeScriptShots(response);
  const stored = config.splitState && typeof config.splitState === 'object' ? config.splitState : {};
  const storedState = text(stored.state).toLowerCase();
  const storedSourceMarker = text(stored.sourceMarker);
  const sourceIsStale = storedState === 'stale' || (
    storedSourceMarker !== '' && storedSourceMarker !== scriptSourceMarker(config.sourceText)
  );
  if (sourceIsStale) {
    return { ...normalized, state: 'stale', errorMessage: '', shots: [] };
  }
  const responseState = text(normalized.state).toLowerCase();
  if (scriptSplitReady(config) && ['completed', 'done'].includes(responseState)) return normalized;
  const activeResponseState = ['loading', 'pending', 'queued', 'running', 'processing', 'failed', 'canceled', 'cancelled'].includes(responseState)
    ? responseState
    : '';
  return {
    ...normalized,
    state: activeResponseState || storedState || 'stale',
    errorMessage: normalized.errorMessage || text(stored.errorMessage),
    shots: []
  };
}

export async function runAfterDurableCanvasSave(flushCanvas, submit) {
  if (typeof flushCanvas !== 'function') {
    const error = new Error('画布尚未准备完成，未提交生成请求');
    error.code = 'CANVAS_SAVE_REQUIRED';
    throw error;
  }
  try {
    await flushCanvas();
  } catch (cause) {
    const error = new Error(`画布保存失败，未提交生成请求：${cause?.message || '请重试'}`);
    error.code = 'CANVAS_SAVE_REQUIRED';
    error.cause = cause;
    throw error;
  }
  return submit();
}

function selectionHas(selection, id) {
  if (selection instanceof Set) return selection.has(id);
  if (Array.isArray(selection)) return selection.includes(id);
  return Boolean(selection && typeof selection === 'object' && selection[id]);
}

export function selectedShotIDs(shots, selection) {
  return (Array.isArray(shots) ? shots : [])
    .map(shot => text(shot?.id))
    .filter((id, index, values) => id && values.indexOf(id) === index && selectionHas(selection, id));
}

export function scriptShotsByNode(shots = []) {
  const grouped = {};
  for (const shot of Array.isArray(shots) ? shots : []) {
    const nodeID = text(shot?.scriptNodeId);
    if (!nodeID) continue;
    if (!grouped[nodeID]) grouped[nodeID] = [];
    grouped[nodeID].push(shot);
  }
  for (const nodeID of Object.keys(grouped)) {
    grouped[nodeID].sort((left, right) => (
      Number(left?.sequence || 0) - Number(right?.sequence || 0) || text(left?.id).localeCompare(text(right?.id))
    ));
  }
  return grouped;
}

export function buildShotBatchPayload({ scriptNodeId, selectedIDs, kind, provider, modelId }) {
  const shotIds = [...new Set((Array.isArray(selectedIDs) ? selectedIDs : []).map(text).filter(Boolean))];
  const payload = {
    scriptNodeId: text(scriptNodeId),
    shotIds,
    kind: text(kind).toLowerCase(),
    modelId: text(modelId)
  };
  const normalizedProvider = text(provider);
  if (normalizedProvider) payload.provider = normalizedProvider;
  return payload;
}
