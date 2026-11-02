const VERSION_LABELS = {
  sensitive: '敏感处理后原文',
  original: '原文',
  ai1: 'AI1',
  ai2: 'AI2',
  ai3: 'AI3',
  ai4: 'AI4',
  ai5: 'AI5'
};

const textValue = value => String(value || '').trim();

export function batchFactoryVersionLabel(version) {
  return VERSION_LABELS[String(version || '').toLowerCase()] || String(version || '正文');
}

export function batchFactorySourceText(detail = {}, version = 'sensitive') {
  const selected = String(version || 'sensitive').toLowerCase();
  const versions = detail?.versions && typeof detail.versions === 'object' ? detail.versions : {};
  if (selected === 'sensitive') {
    return textValue(detail.processedText || detail.original || '');
  }
  if (selected === 'original') {
    return textValue(detail.originalRaw || detail.original || detail.processedText || '');
  }
  return textValue(versions[selected]);
}

export function batchFactoryRawSourceText(detail = {}) {
  const source = detail?.originalRaw ?? detail?.rawData ?? detail?.original ?? detail?.processedText ?? '';
  return typeof source === 'string' ? source : '';
}

export function batchFactorySourceMetadata(detail = {}, version = 'sensitive', fallback = {}) {
  const meta = detail?.meta && typeof detail.meta === 'object' ? detail.meta : {};
  const bookInfo = detail?.bookInfo && typeof detail.bookInfo === 'object' ? detail.bookInfo : {};
  const value = (...values) => values.map(textValue).find(Boolean) || '';
  const originalRawChars = Number(detail?.originalRawChars || meta.originalRawChars || 0);
  const processedChars = Number(detail?.processedChars || meta.processedChars || 0);
  const sensitiveReplaced = Number(meta.sensitiveReplaced || 0);
  const genderSource = value(detail?.genderSource, meta.genderSource);
  const styleSource = value(detail?.styleSource, meta.styleSource);
  const savedGenderConfidence = Number(detail?.genderConfidence ?? meta.genderConfidence ?? 0);
  const savedStyleConfidence = Number(detail?.styleConfidence ?? meta.styleConfidence ?? 0);
  const genderConfidence = savedGenderConfidence > 0 ? savedGenderConfidence : (genderSource && genderSource !== 'ai' ? 1 : 0);
  const styleConfidence = savedStyleConfidence > 0 ? savedStyleConfidence : (styleSource && styleSource !== 'ai' ? 1 : 0);
  const stages = Array.isArray(meta.rulesTrace) ? meta.rulesTrace.map(stage => ({
    stage: textValue(stage?.stage),
    title: textValue(stage?.title),
    changed: Boolean(stage?.changed),
    chars: Number(stage?.chars || 0),
    lines: Number(stage?.lines || 0)
  })) : [];
  return {
    sourceMode: 'novel_fetch',
    sourceContentVersion: String(version || 'sensitive').toLowerCase(),
    sourceTaskId: value(detail?.sourceTaskId, detail?.taskId, detail?.bookId, meta.sourceTaskId, meta.taskId),
    processingRuleVersion: value(meta.processingRuleVersion, meta.processingVersion) || 'novel-fetch-rules-v1',
    gender: value(fallback.gender, detail?.gender, meta.gender),
    genderSource,
    genderReason: value(detail?.genderReason, meta.genderReason),
    genderConfidence,
    style: value(fallback.style, detail?.style, meta.style),
    styleSource,
    styleReason: value(detail?.styleReason, meta.styleReason),
    styleConfidence,
    category: value(meta.category, bookInfo.category),
    genre: value(meta.genre, bookInfo.genre),
    sensitiveStatus: value(meta.sensitiveStatus),
    sensitiveReplaced,
    originalRawChars,
    processedChars,
    processingSummary: { originalRawChars, processedChars, sensitiveReplaced, stages }
  };
}

export function batchFactoryAvailableVersions(details = []) {
  const available = new Set(['sensitive']);
  if (details.some(detail => batchFactorySourceText(detail, 'original'))) available.add('original');
  for (const version of ['ai1', 'ai2', 'ai3', 'ai4', 'ai5']) {
    if (details.length && details.every(detail => batchFactorySourceText(detail, version))) available.add(version);
  }
  return [...available];
}
