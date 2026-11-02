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
      ? {
          ratios: ['16:9', '9:16', '1:1'],
          qualityLevels: ['low', 'standard', 'high'],
          maxOutputCount: 4,
          referenceKinds: ['image'],
          resolutionManagedByModel: true,
          acceptsTextPrompt: true,
          acceptsImageReferences: true,
          acceptsVideoReferences: false,
          acceptsAudioReferences: false,
          maxReferenceAssets: 10
        }
      : {
          ratios: ['16:9', '9:16'],
          resolutions: ['720P', '1080P'],
          duration: { min: 5, max: 15, default: 5 },
          maxOutputCount: 1,
          referenceKinds: ['image'],
          acceptsTextPrompt: true,
          acceptsImageReferences: true,
          acceptsVideoReferences: false,
          acceptsAudioReferences: false,
          maxReferenceAssets: 10
        }
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
  const selected = String(config.modelId || config.model || config[`${kind}ModelId`] || '').trim();
  const catalogModel = (Array.isArray(models) ? models : []).find(model => model?.kind === kind && model?.id === selected);
  const provider = normalizeModelProvider(config.provider || config[`${kind}Provider`] || providerForCatalogModel(catalogModel, 'personal_api'));
  const available = compatibleModels(models, { kind, provider }).find(model => model.id === selected) || null;
  return {
    provider,
    modelId: selected,
    model: available,
    available: Boolean(available),
    reason: selected && !available ? '当前模型不可用' : ''
  };
}

export function modelInputCapabilityError(model = {}, { prompt = '', imageReferences = 0, videoReferences = 0, audioReferences = 0 } = {}) {
  const capabilities = model?.capabilities;
  if (!capabilities || typeof capabilities !== 'object') return '';
  if (String(prompt || '').trim() && capabilities.acceptsTextPrompt === false) return '当前模型不支持文字提示词';
  if (Number(imageReferences) > 0 && capabilities.acceptsImageReferences === false) return '当前模型不支持图片参考';
  if (Number(videoReferences) > 0 && capabilities.acceptsVideoReferences === false) return '当前模型不支持视频参考';
  if (Number(audioReferences) > 0 && capabilities.acceptsAudioReferences === false) return '当前模型不支持音频参考';
  const max = Number(capabilities.maxReferenceAssets);
  const total = Math.max(0, Number(imageReferences) || 0) + Math.max(0, Number(videoReferences) || 0) + Math.max(0, Number(audioReferences) || 0);
  if (Number.isFinite(max) && max >= 0 && total > max) return `当前模型最多支持 ${max} 个参考素材`;
  return '';
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
