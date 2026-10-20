const PROVIDER_LABELS = {
  personal_api: '个人中心 API',
  doubao_local_executor: '豆包本地执行器',
  autodl_comfyui: 'AutoDL ComfyUI',
  yfai_seedance: '梵客视频'
};

export function providerLabel(provider) {
  const normalized = normalizeModelProvider(provider);
  return PROVIDER_LABELS[normalized] || normalized;
}

export const LOCAL_DOUBAO_IMAGE_MODEL_OPTIONS = [
  { value: 'Seedream 5.0 Flash', label: 'Seedream 5.0 Flash（专业快速）' },
  { value: 'Seedream 5.0 Pro', label: 'Seedream 5.0 Pro（专业出图 · 4 倍消耗）' },
  { value: 'Seedream 4.5', label: 'Seedream 4.5（日常生成）' },
  { value: 'Seedream 4.0', label: 'Seedream 4.0（基础生图）' }
];

export const LOCAL_DOUBAO_VIDEO_MODEL_OPTIONS = [
  { value: 'Seedance 2.5', label: 'Seedance 2.5（旗舰视频标杆 · 5 倍消耗）' },
  { value: 'Seedance 2.0', label: 'Seedance 2.0（进阶画面表现 · 2 倍消耗）' },
  { value: 'Seedance 2.0 Fast', label: 'Seedance 2.0 Fast（快速出片选择）' },
  { value: 'Seedance 2.0 Mini', label: 'Seedance 2.0 Mini（日常生成使用）' }
];

export function normalizeModelProvider(provider, fallback = 'personal_api') {
  const value = String(provider || '').trim().toLowerCase();
  if (!value) return fallback;
  if (value === 'personal' || value === 'yd_video' || value === 'yadi') return 'personal_api';
  if (value === 'doubao' || value === 'doubao_local' || value === 'local-doubao-executor-video') return 'doubao_local_executor';
  if (value === 'h3' || value === 'autodl' || value === 'minimax_h3' || value === 'minimax-h3-video') return 'autodl_comfyui';
  if (value === 'yfai' || value === 'seedance-2-0-official') return 'yfai_seedance';
  return value;
}

export function providerForCatalogModel(model, fallback = 'personal_api') {
  const provider = String(model?.provider || '').trim();
  if (provider) return normalizeModelProvider(provider, fallback);
  const id = String(model?.id || model?.modelId || '').trim().toLowerCase();
  if (id === 'minimax-h3-video') return 'autodl_comfyui';
  if (id === 'seedance-2-0-official') return 'yfai_seedance';
  if (id === 'local-doubao-executor-image' || id === 'local-doubao-executor-video' || id === 'doubao-seedance') return 'doubao_local_executor';
  return normalizeModelProvider('', fallback);
}

export function providerOptionsForModels(models, kind = '') {
  const providers = ['personal_api', 'doubao_local_executor', 'autodl_comfyui'];
  for (const model of Array.isArray(models) ? models : []) {
    if (kind && model?.kind !== kind) continue;
    const provider = providerForCatalogModel(model, 'personal_api');
    if (!providers.includes(provider)) providers.push(provider);
  }
  return providers.map(value => ({ value, label: PROVIDER_LABELS[value] || value }));
}

function localOptions(kind) {
  return kind === 'video' ? LOCAL_DOUBAO_VIDEO_MODEL_OPTIONS : LOCAL_DOUBAO_IMAGE_MODEL_OPTIONS;
}

function defaultLocalModel(kind) {
  return kind === 'video' ? 'Seedance 2.0 Fast' : 'Seedream 4.5';
}

export function modelOptionsForProvider(models, kind, provider) {
  const normalizedProvider = normalizeModelProvider(provider);
  if (normalizedProvider === 'doubao_local_executor') return localOptions(kind);
  return (Array.isArray(models) ? models : [])
    .filter(model => model?.kind === kind && String(model?.id || '').trim())
    .filter(model => providerForCatalogModel(model, 'personal_api') === normalizedProvider)
    .map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id }));
}

export function reconcileProviderModel({ models, kind, provider, modelId }) {
  const normalizedProvider = normalizeModelProvider(provider);
  const options = modelOptionsForProvider(models, kind, normalizedProvider);
  const selected = String(modelId || '').trim();
  return {
    provider: normalizedProvider,
    modelId: options.some(option => option.value === selected) ? selected : (normalizedProvider === 'doubao_local_executor' ? defaultLocalModel(kind) : (options[0]?.value || ''))
  };
}
