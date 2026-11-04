const PROVIDER_LABELS = {
  personal_api: '个人中心 API',
  doubao_local_executor: '豆包本地执行器',
  autodl_comfyui: 'AutoDL ComfyUI',
  yfai: 'YFAI',
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

// 梵客视频 API 在后台是一条目录记录 + 一个 modelId 单选，但实际可用型号有 6 个。
// 业务下拉需要把它们展开成独立选项；值用 `fanke-video-api::<modelId>` 限定归属，
// 避免与其它引擎的同名型号串味，后端按前缀解析到 yfai_seedance 执行。
export const FANKE_VIDEO_SELECTION_PREFIX = 'fanke-video-api::';

export const FANKE_VIDEO_MODELS = [
  { value: 'ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746', label: 'MiniMax H3-768p' },
  { value: 'ft-video-v1-dd97afc97a7c3cba8e78a6e429933597', label: 'Seedance 2.0 Mini 720p' },
  { value: 'ft-video-v1-20205b9fbcec3a1094671e1c08ab7f93', label: 'Seedance 2.0 Fast 720p' },
  { value: 'ft-video-v1-d40a056d0fc20901a8b7e33d737fef51', label: 'Seedance 2.0 满血 720p' },
  { value: 'ft-video-v1-451adae35b0c4a3d275c2c46394abc98', label: 'Seedance 2.5 满血 720p' },
  { value: 'ft-video-v1-7ac9cee104fdf0f39dcf2402e3fb1a10', label: 'Wan 3.0 Prime 720p' }
];

// 从 fanke-video-api::<modelId> 中取出具体型号；旧值（裸 fanke-video-api）返回空，
// 由调用方回退到目录里的默认型号。
export function fanKeModelFromSelection(modelId) {
  const raw = String(modelId || '').trim();
  if (!raw.toLowerCase().startsWith(FANKE_VIDEO_SELECTION_PREFIX)) return '';
  return raw.slice(FANKE_VIDEO_SELECTION_PREFIX.length);
}

function fanKeCatalogRow(models) {
  return (Array.isArray(models) ? models : []).find(model => String(model?.id || '').trim().toLowerCase() === 'fanke-video-api');
}

function fanKeDefaultModelId(models) {
  const rowModelId = String(fanKeCatalogRow(models)?.modelId || '').trim();
  if (rowModelId && FANKE_VIDEO_MODELS.some(item => item.value === rowModelId)) return rowModelId;
  return FANKE_VIDEO_MODELS[0].value;
}

function fanKeVideoOptions(models) {
  const row = fanKeCatalogRow(models);
  const options = FANKE_VIDEO_MODELS.map(item => ({
    value: `${FANKE_VIDEO_SELECTION_PREFIX}${item.value}`,
    label: item.label
  }));
  // 目录里若配了 6 个之外的型号，补一条，避免丢失已配置的默认值。
  const rowModelId = String(row?.modelId || '').trim();
  if (rowModelId && !FANKE_VIDEO_MODELS.some(item => item.value === rowModelId)) {
    options.push({
      value: `${FANKE_VIDEO_SELECTION_PREFIX}${rowModelId}`,
      label: row?.displayName || row?.name || rowModelId
    });
  }
  return options;
}

function normalizeFanKeSelection(value, models) {
  const raw = String(value || '').trim();
  if (raw.toLowerCase().startsWith(FANKE_VIDEO_SELECTION_PREFIX)) return raw;
  const defaultModelId = fanKeDefaultModelId(models);
  if (!raw || raw.toLowerCase() === 'fanke-video-api') {
    return `${FANKE_VIDEO_SELECTION_PREFIX}${defaultModelId}`;
  }
  if (FANKE_VIDEO_MODELS.some(item => item.value === raw)) {
    return `${FANKE_VIDEO_SELECTION_PREFIX}${raw}`;
  }
  if (String(fanKeCatalogRow(models)?.modelId || '').trim() === raw) {
    return `${FANKE_VIDEO_SELECTION_PREFIX}${raw}`;
  }
  return raw;
}

export function normalizeModelProvider(provider, fallback = 'personal_api', modelId = '') {
  const value = String(provider || '').trim().toLowerCase();
  const legacyModel = String(modelId || '').trim().toLowerCase();
  if (!value) return fallback;
  if (value === 'personal' || value === 'yd_video' || value === 'yadi' || value === 'openai_compatible' || value === 'custom' || value === 'custom_api') return 'personal_api';
  if (value === 'doubao' || value === 'doubao_local' || value === 'local-doubao-executor-video') return 'doubao_local_executor';
  if (value === 'h3' || value === 'autodl' || value === 'minimax_h3' || value === 'minimax-h3-video') return 'autodl_comfyui';
  if (value === 'yfai' || value === 'seedance-2-0-official') return 'yfai';
  // 旧配置把 YFAI 官方直连与梵客聚合站共用 yfai_seedance；只有仍选着官方
  // Seedance 的存量记录才映射到新分组，梵客那几个变体保持原组不动。
  if (value === 'yfai_seedance' && legacyModel === 'seedance-2-0-official') return 'yfai';
  return value;
}

export function providerForCatalogModel(model, fallback = 'personal_api') {
  const id = String(model?.id || model?.modelId || '').trim().toLowerCase();
  // 梵客视频的具体型号以 fanke-video-api::<modelId> 形式保存，归属仍是梵客引擎。
  if (id.startsWith(FANKE_VIDEO_SELECTION_PREFIX)) return 'yfai_seedance';
  // 平台内置模型按 id 归属分组：官方直连 Seedance 归 YFAI，梵客聚合站
  // (fanke-video-api) 仍归梵客分组，互不搬迁。
  if (id === 'seedance-2-0-official') return 'yfai';
  if (id === 'fanke-video-api') return 'yfai_seedance';
  const provider = String(model?.provider || '').trim();
  if (provider) return normalizeModelProvider(provider, fallback);
  if (id === 'minimax-h3-video') return 'autodl_comfyui';
  if (id === 'local-doubao-executor-image' || id === 'local-doubao-executor-video' || id === 'doubao-seedance') return 'doubao_local_executor';
  return normalizeModelProvider('', fallback);
}

export function providerOptionsForModels(models, kind = '') {
  let providers = ['personal_api'];
  if (kind === 'video') providers = ['personal_api', 'doubao_local_executor', 'autodl_comfyui', 'yfai', 'yfai_seedance'];
  else if (kind === 'image') providers = ['personal_api', 'doubao_local_executor', 'autodl_comfyui'];
  for (const model of Array.isArray(models) ? models : []) {
    if (kind && model?.kind !== kind) continue;
    const provider = providerForCatalogModel(model, 'personal_api');
    if (!providers.includes(provider)) providers.push(provider);
  }
  return providers.map(value => ({ value, label: PROVIDER_LABELS[value] || value }));
}

function localOptions(kind) {
  if (kind === 'video') return LOCAL_DOUBAO_VIDEO_MODEL_OPTIONS;
  if (kind === 'image') return LOCAL_DOUBAO_IMAGE_MODEL_OPTIONS;
  return [];
}

function defaultLocalModel(kind) {
  if (kind === 'video') return 'Seedance 2.0 Fast';
  if (kind === 'image') return 'Seedream 4.5';
  return '';
}

export function modelOptionsForProvider(models, kind, provider, modelId = '', options = {}) {
  const normalizedProvider = normalizeModelProvider(provider, 'personal_api', modelId);
  if (normalizedProvider === 'doubao_local_executor') return localOptions(kind);
  const catalogOptions = (Array.isArray(models) ? models : [])
    .filter(model => model?.kind === kind && String(model?.id || '').trim())
    .filter(model => providerForCatalogModel(model, 'personal_api') === normalizedProvider)
    .map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id }));
  if (kind === 'video' && normalizedProvider === 'yfai_seedance' && options.expandFanKeModels) {
    const fanKeOptions = fanKeVideoOptions(models);
    const fanKeValues = new Set(fanKeOptions.map(option => option.value));
    // 目录里的 fanke-video-api 记录已被展开成具体型号，不能再以裸值重复出现。
    const extras = catalogOptions.filter(option => {
      if (String(option.value || '').trim().toLowerCase() === 'fanke-video-api') return false;
      return !fanKeValues.has(option.value);
    });
    return [...fanKeOptions, ...extras];
  }
  return catalogOptions;
}

export function reconcileProviderModel({ models, kind, provider, modelId, expandFanKeModels = false }) {
  const normalizedProvider = normalizeModelProvider(provider, 'personal_api', modelId);
  const options = modelOptionsForProvider(models, kind, normalizedProvider, modelId, { expandFanKeModels });
  let selected = String(modelId || '').trim();
  // 旧配置里的裸 fanke-video-api、裸型号 id 都归一成具体型号，避免下拉显示空值。
  if (kind === 'video' && normalizedProvider === 'yfai_seedance' && expandFanKeModels) {
    selected = normalizeFanKeSelection(selected, models);
  }
  return {
    provider: normalizedProvider,
    modelId: options.some(option => option.value === selected) ? selected : (normalizedProvider === 'doubao_local_executor' ? defaultLocalModel(kind) : (options[0]?.value || ''))
  };
}

export function providerModelSettingsPatch(input = {}) {
  const kind = input.kind === 'video' ? 'video' : 'image';
  const reconciled = reconcileProviderModel({ ...input, kind });
  return kind === 'video'
    ? { videoProvider: reconciled.provider, videoModelId: reconciled.modelId }
    : { imageProvider: reconciled.provider, imageModelId: reconciled.modelId };
}
