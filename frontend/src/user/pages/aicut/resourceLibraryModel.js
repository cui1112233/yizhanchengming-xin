export function assetDurationLabel(asset) {
  if (asset.kind === 'image') return '图片';
  const seconds = Math.floor(Number(asset.durationMs) / 1000);
  if (!Number.isFinite(seconds) || seconds <= 0) return asset.kind === 'audio' ? '音频' : '视频';
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor(seconds % 3600 / 60);
  const remainder = String(seconds % 60).padStart(2, '0');
  return hours ? `${hours}:${String(minutes).padStart(2, '0')}:${remainder}` : `${String(minutes).padStart(2, '0')}:${remainder}`;
}

export function assetThumbnailUrl(asset) {
  return asset.posterUrl || asset.thumbnailUrl || asset.poster || (asset.kind === 'image' ? asset.url : '') || '';
}

export function assetToolActions(asset, { proxyBusyId = '', backgroundBusyId = '' } = {}) {
  const actions = [];
  if (['audio', 'video'].includes(asset.kind)) actions.push({ key: 'asr', label: '识别字幕' });
  if (asset.kind === 'video') actions.push({
    key: 'proxy',
    label: proxyBusyId === asset.id ? '生成中…' : asset.proxyUrl ? '重新生成代理' : '生成代理',
    disabled: Boolean(asset.id && proxyBusyId === asset.id),
  });
  if (asset.kind === 'image') actions.push({ key: 'background', label: backgroundBusyId === asset.id ? '取消去背' : '背景移除' });
  if (asset.kind === 'audio') actions.push({ key: 'analyze', label: '分析节拍' }, { key: 'clone', label: '克隆音色' });
  return actions;
}
