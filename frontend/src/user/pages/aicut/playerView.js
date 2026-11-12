export function previewSource(asset, quality = 'original') {
  return quality === 'proxy' && asset?.kind === 'video' ? asset.proxyUrl || asset.url : asset?.url || '';
}
export function previewDimensions(ratio, availableWidth, availableHeight, zoom = 'fit') {
  const [width, height] = ratio === '9:16' ? [1080, 1920] : ratio === '1:1' ? [1080, 1080] : [1920, 1080];
  const fit = Math.max(.01, Math.min(Math.max(1, availableWidth - 32) / width, Math.max(1, availableHeight - 24) / height));
  const scale = zoom === 'fit' ? fit : Math.min(2, Math.max(.1, Number(zoom) / 100 || fit));
  return { width: width * scale, height: height * scale, scale };
}
