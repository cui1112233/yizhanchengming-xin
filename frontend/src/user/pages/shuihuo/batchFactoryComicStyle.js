export function comicStyleSelection(preset) {
  return {
    presetId: String(preset?.id || ''),
    presetName: String(preset?.name || preset?.id || ''),
    presetSlot: String(preset?.slot || 'batch.comic-style'),
    presetVersion: Number(preset?.version || 1)
  };
}

export function normalizeComicReference(value) {
  const reference = value && typeof value === 'object' ? value : {};
  return {
    assetId: String(reference.assetId || ''),
    url: String(reference.url || ''),
    previewUrl: String(reference.previewUrl || reference.url || ''),
    enabled: reference.enabled !== false,
    applyToAssets: reference.applyToAssets !== false,
    applyToPages: reference.applyToPages !== false
  };
}

export function comicStylePreviewReference(preset) {
  const presetId = String(preset?.id || '');
  return normalizeComicReference({
    assetId: `preset:${presetId}`,
    url: `/api/local-executor/v1/batch-comic-style-previews/${encodeURIComponent(presetId)}`,
    previewUrl: String(preset?.previewUrl || ''),
    enabled: true,
    applyToAssets: true,
    applyToPages: true
  });
}
