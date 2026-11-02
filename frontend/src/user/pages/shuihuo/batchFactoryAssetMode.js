const legacyComicPresetIds = new Set(['batch-comic-assets', 'vip-smart-character-derive']);

function isComicAsset(asset) {
  const source = String(asset?.source || '').toLowerCase();
  return source === 'comic' || (source === 'manual' && legacyComicPresetIds.has(String(asset?.extractionPresetId || '').trim()));
}

export function assetsForProductionMode(assets, productionMode) {
  const list = Array.isArray(assets) ? assets : [];
  return productionMode === 'comic'
    ? list.filter(isComicAsset)
    : list.filter(asset => !isComicAsset(asset));
}
