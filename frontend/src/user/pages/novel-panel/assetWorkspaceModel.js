import { normalizeNovelAsset } from './novelPanelContracts.js';

const TYPE_LABELS = { character: '人物', scene: '场景', prop: '道具', material: '素材' };

export function buildAssetCards(values = []) {
  return (Array.isArray(values) ? values : []).map(value => {
    const asset = normalizeNovelAsset(value);
    const preview = asset.imageUrl
      ? { kind: 'image', src: asset.imageUrl }
      : { kind: 'placeholder', label: TYPE_LABELS[asset.assetType] || '资产' };
    const error = typeof value?.error === 'string' ? value.error : '';
    return { ...asset, preview, error, retryable: Boolean(error) };
  });
}
