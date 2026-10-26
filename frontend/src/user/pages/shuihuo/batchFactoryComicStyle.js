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

export function wrapComicStyleIndex(index, count) {
  if (!Number.isFinite(count) || count <= 0) return 0;
  return ((Number(index) % count) + count) % count;
}

export function comicStyleRelativeOffset(index, activeIndex, count) {
  if (!Number.isFinite(count) || count <= 0) return 0;
  let offset = wrapComicStyleIndex(index, count) - wrapComicStyleIndex(activeIndex, count);
  if (offset > count / 2) offset -= count;
  if (offset < -count / 2) offset += count;
  return offset;
}

export function comicStyleIndexFromDrag(activeIndex, deltaX, count, threshold = 36) {
  if (Math.abs(deltaX) < threshold) return wrapComicStyleIndex(activeIndex, count);
  return wrapComicStyleIndex(activeIndex + (deltaX < 0 ? 1 : -1), count);
}

export function comicStyleWheelDirection(deltaX, deltaY, shiftKey = false) {
  const horizontal = Number(deltaX) || 0;
  const vertical = Number(deltaY) || 0;
  const value = shiftKey ? (vertical || horizontal) : horizontal;
  if ((!shiftKey && Math.abs(horizontal) <= Math.max(12, Math.abs(vertical))) || Math.abs(value) < 12) return 0;
  return value > 0 ? 1 : -1;
}

export function comicStyleCoverFlowStyle(offset) {
  const normalizedOffset = Number(offset) || 0;
  const distance = Math.abs(normalizedOffset);
  return {
    '--comic-style-offset': normalizedOffset,
    '--comic-style-distance': Math.min(distance, 4),
    '--comic-style-brightness': Number(Math.max(0.52, 1 - distance * 0.11).toFixed(2)),
    '--comic-style-rotation': normalizedOffset === 0 ? '0deg' : (normalizedOffset < 0 ? '42deg' : '-42deg'),
    '--comic-style-scale': Number(Math.max(0.72, 1.08 - distance * 0.09).toFixed(2)),
    '--comic-style-translate-x': `${normalizedOffset * 132}px`,
    '--comic-style-translate-z': `${90 - distance * 52}px`,
    opacity: distance > 4 ? 0 : Math.max(0.3, 1 - distance * 0.17),
    pointerEvents: distance > 4 ? 'none' : 'auto',
    zIndex: 20 - distance
  };
}
