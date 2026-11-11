import { apiRequest } from './client.js';

const AVAILABLE_MODEL_KINDS = Object.freeze(['text', 'image', 'video', 'asr']);
const AVAILABLE_MODEL_TIMEOUT_MS = 8000;
let availableModelCatalogResult = null;
let availableModelCatalogPromise = null;

const MODEL_KIND_LABELS = Object.freeze({
  text: '文本',
  video: '视频',
  image: '图片',
  audio: '音频',
  asr: '语音识别'
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
  const result = await request(`/api/models?kind=${encodeURIComponent(validKind)}`, {
    timeoutMs: AVAILABLE_MODEL_TIMEOUT_MS,
    suppressGlobalError: true
  });
  return Array.isArray(result?.models) ? result.models : [];
}

// Account models are one shared directory. Batch Factory video, comic, create,
// unified and single-book forms must consume this same snapshot instead of
// starting their own three-request chains every time a modal opens.
export function peekAvailableModelCatalog() {
  return availableModelCatalogResult;
}

export function invalidateAvailableModelCatalog() {
  availableModelCatalogResult = null;
  availableModelCatalogPromise = null;
}

export function getAvailableModelCatalog({ force = false, loadKind = listAvailableModels } = {}) {
  if (force) invalidateAvailableModelCatalog();
  if (availableModelCatalogResult) return Promise.resolve(availableModelCatalogResult);
  if (availableModelCatalogPromise) return availableModelCatalogPromise;
  availableModelCatalogPromise = Promise.allSettled(AVAILABLE_MODEL_KINDS.map(kind => loadKind(kind)))
    .then(results => {
      const models = [];
      const failedKinds = [];
      results.forEach((result, index) => {
        if (result.status === 'fulfilled') models.push(...(Array.isArray(result.value) ? result.value : []));
        else failedKinds.push(AVAILABLE_MODEL_KINDS[index]);
      });
      availableModelCatalogResult = Object.freeze({ models, failedKinds });
      return availableModelCatalogResult;
    })
    .finally(() => { availableModelCatalogPromise = null; });
  return availableModelCatalogPromise;
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
  const groups = await Promise.all(['text', 'video', 'image', 'audio', 'asr'].map(kind => listManagedModelsByKind(kind)));
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

export async function createManagedModel(model) {
  const result = await apiRequest('/api/config/models', { method: 'POST', body: JSON.stringify(model) });
  invalidateAvailableModelCatalog();
  return result;
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

export async function updateManagedModel(modelId, patch) {
  const result = await apiRequest(`/api/config/models/${encodeURIComponent(modelId)}`, { method: 'PATCH', body: JSON.stringify(patch) });
  invalidateAvailableModelCatalog();
  return result;
}

export async function deleteManagedModel(modelId) {
  const result = await apiRequest(`/api/config/models/${encodeURIComponent(modelId)}`, { method: 'DELETE' });
  invalidateAvailableModelCatalog();
  return result;
}

export function refreshLocalDoubaoPairingStatus() {
  return apiRequest('/api/config/models/local-doubao-executor-video/pairing-status');
}
