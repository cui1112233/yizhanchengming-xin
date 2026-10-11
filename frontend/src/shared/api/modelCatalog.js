import { apiRequest } from './client.js';

const MODEL_KIND_LABELS = Object.freeze({
  text: '文本',
  video: '视频',
  image: '图片',
  audio: '音频'
});

export function isModelKind(kind) {
  return Object.hasOwn(MODEL_KIND_LABELS, kind);
}

export function assertModelKind(kind) {
  if (!isModelKind(kind)) throw new TypeError('无效的模型类型');
  return kind;
}

export function modelKindLabel(kind) {
  return isModelKind(kind) ? MODEL_KIND_LABELS[kind] : '';
}

export function modelSelectOptions(models, kind) {
  return (Array.isArray(models) ? models : [])
    .filter(model => model?.kind === kind && String(model?.id || '').trim())
    .map(model => ({
      value: model.id,
      label: model.displayName || model.name || model.modelId || model.id
    }));
}

export function getTypedModelSelectState(kind, { models = [], error = '' } = {}) {
  const validKind = isModelKind(kind);
  const safeModels = Array.isArray(models) ? models : [];
  if (!validKind) {
    return { validKind, models: [], placeholder: '模型类型无效' };
  }
  const label = modelKindLabel(kind);
  return {
    validKind,
    models: safeModels,
    placeholder: error || (safeModels.length ? '选择模型' : `尚未添加可用的${label}模型`)
  };
}

function kindPath(path, kind) {
  return kind ? `${path}?kind=${encodeURIComponent(assertModelKind(kind))}` : path;
}

export async function listAvailableModels(kind, request = apiRequest) {
  const validKind = assertModelKind(kind);
  const result = await request(`/api/models?kind=${encodeURIComponent(validKind)}`);
  return Array.isArray(result?.models) ? result.models : [];
}

export async function refreshAvailableModels(kind, request = apiRequest) {
  try {
    return { models: await listAvailableModels(kind, request), error: '' };
  } catch (error) {
    const label = modelKindLabel(kind) || '指定';
    const message = error instanceof Error && error.message ? error.message : '请求失败，请稍后重试';
    return { models: [], error: `加载${label}模型失败：${message}` };
  }
}

export async function listManagedModels() {
  const groups = await Promise.all(['text', 'video', 'image', 'audio'].map(kind => listManagedModelsByKind(kind)));
  return groups.flat();
}

export async function listManagedModelsByKind(kind) {
  const result = await apiRequest(kindPath('/api/config/models', kind));
  return Array.isArray(result?.models) ? result.models : [];
}

export async function getManagedModelQuotas() {
  const result = await apiRequest('/api/config/models/quotas');
  return Array.isArray(result?.quotas) ? result.quotas : [];
}

export function createManagedModel(model) {
  return apiRequest('/api/config/models', { method: 'POST', body: JSON.stringify(model) });
}

export function testManagedTextModel(model, existingModelId = '') {
  return apiRequest('/api/config/models/test', {
    method: 'POST',
    body: JSON.stringify({ ...model, ...(existingModelId ? { existingModelId } : {}) }),
    suppressGlobalError: true
  });
}

export function testManagedAudioModel(model, existingModelId = '') {
  return apiRequest('/api/config/models/test-audio', {
    method: 'POST',
    body: JSON.stringify({ ...model, ...(existingModelId ? { existingModelId } : {}) }),
    suppressGlobalError: true
  });
}

export function updateManagedModel(modelId, patch) {
  return apiRequest(`/api/config/models/${encodeURIComponent(modelId)}`, { method: 'PATCH', body: JSON.stringify(patch) });
}

export function deleteManagedModel(modelId) {
  return apiRequest(`/api/config/models/${encodeURIComponent(modelId)}`, { method: 'DELETE' });
}

export function refreshLocalDoubaoPairingStatus() {
  return apiRequest('/api/config/models/local-doubao-executor-video/pairing-status');
}
