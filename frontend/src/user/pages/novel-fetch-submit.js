export function submissionPreviewMatches(selectedIds, previewIds, plan) {
  if (!plan || !Array.isArray(selectedIds) || !selectedIds.length) return false;
  const selected = selectedIds.map(String).sort().join('|');
  const previewed = (previewIds || []).map(String).sort().join('|');
  return selected === previewed;
}

export function singleVersionRetryPayload(row = {}) {
  const bookId = row.bookId || row.book_id;
  const version = row.version;
  if (!bookId || !version) return null;
  return { taskIds: [bookId], versions: [version] };
}
