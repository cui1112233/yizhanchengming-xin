export function normalizeNovelHistory(value) {
  const records = Array.isArray(value?.history) ? value.history : Array.isArray(value) ? value : [];
  return records.map(item => ({
    id: String(item?.id || ''),
    status: String(item?.status || 'unknown'),
    updatedAt: String(item?.updatedAt || item?.updated_at || '')
  })).sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}
