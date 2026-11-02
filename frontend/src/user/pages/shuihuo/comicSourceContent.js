const boundaryNames = {
  sentence: '完整句子',
  dialogue: '连续对白结束',
  paragraph: '自然段结束',
  scene: '场景转换'
};

export function comicSourceBoundaryOptions(segment = {}) {
  return (segment?.safeBoundaries || []).map((boundary, index) => ({
    value: Number(boundary.offset),
    label: `${boundaryNames[boundary.type] || '安全拆分点'} ${index + 1} · ${String(boundary.preview || '').trim() || '查看原文位置'}`
  }));
}

export function comicSourceSplitPayload(direction, boundaryOffset, expectedRevision) {
  return {
    direction,
    boundaryOffset: Number(boundaryOffset),
    expectedRevision: Number(expectedRevision)
  };
}

export function comicSourceStatusLabel(segment = {}) {
  if (segment.mediaStale) return '原文分段已变化';
  if (segment.promptStale) return '等待生成提示词';
  return '原文已绑定';
}
