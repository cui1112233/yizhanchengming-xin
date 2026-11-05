export const AUTOMATION_STAGE_LABELS = Object.freeze({
  audio_planning: '配音与逐行时长',
  assets: '资产处理',
  director: '文本提示词',
  video: '视频生成',
  merge: '视频合成',
  comic: '漫画分页与提示词',
  fetch_original: '获取原文'
});

/**
 * Collect newly failed books in one pass. The caller passes the persisted
 * `notified` map so the same failure signature never produces a second toast.
 * Returning all entries at once is what lets the UI show one aggregated
 * message instead of one toast per book.
 */
export function collectNewAutomationFailures({ books, automationStatus, batchId, notified, firstPass = false, resolveEntry }) {
  const entries = [];
  (books || []).forEach(book => {
    const entry = resolveEntry(automationStatus, batchId, book.id);
    const failed = entry && String(entry.status || '').toLowerCase() === 'failed';
    const errorText = String(entry?.error || '').trim();
    if (!failed || !errorText) {
      if (notified) delete notified[book.id];
      return;
    }
    const stage = String(entry.stage || '').toLowerCase();
    const signature = `${stage}:${errorText}`;
    if (notified?.[book.id] === signature) return;
    if (notified) notified[book.id] = signature;
    if (firstPass) return;
    entries.push({ book, signature, stage, errorText });
  });
  return entries;
}

export function automationFailureSummaryContent(entries) {
  if (!entries?.length) return '';
  const first = entries[0];
  const stageLabel = AUTOMATION_STAGE_LABELS[first.stage] || first.stage || '自动生产';
  if (entries.length > 1) return `${entries.length} 本书自动生产失败（${stageLabel}等）`;
  if (first.stage === 'video' && /ref_image_0/i.test(first.errorText)) return '视频生成失败 · 参考图缺失';
  return `${stageLabel}失败`;
}
