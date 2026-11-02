const SOURCE_LABELS = {
  'manual-upload': '手动上传',
  'scene-library': '场景图库',
  'book-asset': '漫画资产',
  'global-reference': '全局参考图',
  reference: '参考素材'
};

export function referenceSourceLabel(source) {
  return SOURCE_LABELS[String(source || '')] || '参考素材';
}

export function removablePageReferences(references, removedURL) {
  const target = String(removedURL || '');
  return (Array.isArray(references) ? references : [])
    .filter(reference => reference?.removable && String(reference?.url || '') !== target)
    .map(reference => String(reference.url));
}
