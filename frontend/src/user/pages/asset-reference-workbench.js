const VALID_RESOLUTIONS = new Set(['1K', '2K', '4K']);
const VALID_RATIOS = new Set(['16:9', '9:16', '21:9', '1:1', '3:4', '4:3', '2:3', '3:2']);

export function normalizeAssetReferenceSettings(input = {}) {
  const count = Math.min(20, Math.max(1, Number(input.count) || 1));
  const resolution = VALID_RESOLUTIONS.has(String(input.resolution || '').toUpperCase())
    ? String(input.resolution).toUpperCase()
    : '1K';
  const ratio = VALID_RATIOS.has(String(input.ratio || '')) ? String(input.ratio) : '16:9';
  return { count, resolution, ratio };
}

export function validateAssetReferenceInput({ prompt = '', referenceAssets = [], modelId = '' } = {}) {
  if (!String(modelId).trim()) return { ok: false, message: '请选择图片模型' };
  if (!String(prompt).trim() && !(Array.isArray(referenceAssets) && referenceAssets.length)) {
    return { ok: false, message: '请输入提示词或添加参考图片' };
  }
  return { ok: true, message: '' };
}

export function buildAssetReferenceRequest(input = {}) {
  const settings = normalizeAssetReferenceSettings(input);
  const referenceAssetIds = (Array.isArray(input.referenceAssets) ? input.referenceAssets : [])
    .filter(asset => asset?.kind === 'image' && String(asset?.id || '').trim())
    .map(asset => String(asset.id).trim());
  return {
    nodeId: String(input.nodeId || '').trim(),
    nodeVersion: Number(input.nodeVersion) || 1,
    modelId: String(input.modelId || '').trim(),
    mode: 'text_to_image',
    prompt: String(input.prompt || '').trim(),
    parameters: { resolution: settings.resolution, ratio: settings.ratio, quality: 'standard', count: settings.count },
    advanced: {},
    referenceAssetIds: [...new Set(referenceAssetIds)],
    idempotencyKey: `asset-reference:${String(input.taskId || '').trim()}:${String(input.operationId || '').trim()}`
  };
}

// The recovered client submits one single-image job per requested result. This
// also keeps each local request inside the Agent image API's 1..4 limit while
// preserving independent retries, cancellation and partial success.
export function buildAssetReferenceRequests(input = {}) {
  const { count } = normalizeAssetReferenceSettings(input);
  return Array.from({ length: count }, (_, index) => buildAssetReferenceRequest({
    ...input,
    count: 1,
    operationId: `${String(input.operationId || '').trim()}:${index + 1}`
  }));
}

export function assetReferenceRunState({ running = false, cancelled = false } = {}) {
  if (running) return { canSubmit: false, canCancel: true, message: '正在生成，请勿重复提交' };
  if (cancelled) return { canSubmit: true, canCancel: false, message: '已取消本次操作' };
  return { canSubmit: true, canCancel: false, message: '' };
}

export function assetReferenceErrorMessage(error) {
  return String(error?.message || '').trim() || '图片生成失败，请稍后重试';
}

export function normalizeAssetReferenceResults(results = [], metadata = {}) {
  return (Array.isArray(results) ? results : []).flatMap(asset => {
    if (asset?.kind !== 'image' || !String(asset?.id || '').trim()) return [];
    return [{
      ...asset,
      prompt: String(metadata.prompt || '').trim(),
      modelId: String(metadata.modelId || '').trim(),
      ratio: String(metadata.ratio || ''),
      resolution: String(metadata.resolution || '')
    }];
  });
}
