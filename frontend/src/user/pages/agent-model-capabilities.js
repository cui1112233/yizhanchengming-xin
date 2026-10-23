import {
  LOCAL_DOUBAO_IMAGE_MODEL_OPTIONS,
  LOCAL_DOUBAO_VIDEO_MODEL_OPTIONS,
  normalizeModelProvider,
  providerForCatalogModel
} from './shuihuo/providerModelBinding.js';

function localDoubaoModels(kind) {
  const options = kind === 'video'
    ? LOCAL_DOUBAO_VIDEO_MODEL_OPTIONS
    : kind === 'image'
      ? LOCAL_DOUBAO_IMAGE_MODEL_OPTIONS
      : [];
  return options.map(option => ({
    id: option.value,
    modelId: option.value,
    displayName: option.label,
    kind,
    provider: 'doubao_local_executor',
    enabled: true,
    capabilities: kind === 'image'
      ? { ratios: ['16:9', '9:16', '1:1'], qualityLevels: ['low', 'standard', 'high'], maxOutputCount: 4, referenceKinds: ['image'], resolutionManagedByModel: true }
      : { ratios: ['16:9', '9:16'], resolutions: ['720P', '1080P'], duration: { min: 5, max: 15, default: 5 }, maxOutputCount: 1, referenceKinds: ['image'] }
  }));
}

export function compatibleModels(models = [], { kind, provider } = {}) {
  const normalizedProvider = normalizeModelProvider(provider);
  if (normalizedProvider === 'doubao_local_executor') return localDoubaoModels(kind);
  return (Array.isArray(models) ? models : [])
    .filter(model => model?.kind === kind && String(model?.id || '').trim())
    .filter(model => providerForCatalogModel(model, 'personal_api') === normalizedProvider)
    .map(model => ({ ...model, provider: normalizedProvider }));
}

export function reconcileNodeModel(config = {}, models = [], kind = '') {
  const selected = String(config.modelId || config.model || '').trim();
  const catalogModel = (Array.isArray(models) ? models : []).find(model => model?.kind === kind && model?.id === selected);
  const provider = normalizeModelProvider(config.provider || providerForCatalogModel(catalogModel, 'personal_api'));
  const available = compatibleModels(models, { kind, provider }).find(model => model.id === selected) || null;
  return {
    provider,
    modelId: selected,
    model: available,
    available: Boolean(available),
    reason: selected && !available ? '当前模型不可用' : ''
  };
}

export function capabilityDescription(model = {}) {
  const capabilities = model.capabilities || {};
  const parts = [];
  if (capabilities.resolutions?.length) parts.push(capabilities.resolutions.join('/'));
  if (capabilities.referenceKinds?.includes('image')) parts.push('支持参考图');
  if (capabilities.supportsAudio) parts.push('可生成声音');
  return parts.join(' · ') || '使用当前渠道配置';
}

export function expectedDuration(model = {}) {
  const seconds = Number(model.capabilities?.duration?.default);
  return seconds > 0 ? `默认 ${seconds} 秒` : '';
}
