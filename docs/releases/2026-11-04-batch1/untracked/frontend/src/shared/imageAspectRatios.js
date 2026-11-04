// Single source of truth for still-image aspect ratios in the web client.
// Backend mirror: backend/internal/batchfactoryv11/image_aspect_ratios.go
// (SupportedImageAspectRatios). Keep both lists identical.
export const IMAGE_ASPECT_RATIOS = ['16:9', '9:16', '1:1', '4:3', '3:4', '3:2', '2:3'];

export const IMAGE_ASPECT_RATIO_OPTIONS = IMAGE_ASPECT_RATIOS.map((value) => ({ value, label: value }));

export const PORTRAIT_IMAGE_ASPECT_RATIOS = ['9:16', '3:4', '2:3'];
export const LANDSCAPE_IMAGE_ASPECT_RATIOS = ['16:9', '4:3', '3:2'];

export function isSupportedImageAspectRatio(value) {
  return IMAGE_ASPECT_RATIOS.includes(String(value || '').trim());
}

export function isPortraitImageAspectRatio(value) {
  return PORTRAIT_IMAGE_ASPECT_RATIOS.includes(String(value || '').trim());
}

// Unknown/legacy values fall back instead of throwing so stored batches stay readable.
export function normalizeImageAspectRatio(value, fallback = '9:16') {
  const ratio = String(value || '').trim();
  return isSupportedImageAspectRatio(ratio) ? ratio : fallback;
}

export function imageAspectRatioOrientation(value) {
  const ratio = String(value || '').trim();
  if (PORTRAIT_IMAGE_ASPECT_RATIOS.includes(ratio)) return '竖幅构图';
  if (ratio === '1:1') return '方形构图';
  return '横幅构图';
}
