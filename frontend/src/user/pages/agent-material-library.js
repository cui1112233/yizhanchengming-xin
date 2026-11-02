export function materialLibraryRows(items = []) {
  return (Array.isArray(items) ? items : []).map(item => ({
    id: String(item?.id || '').trim(),
    folderId: String(item?.folderId || '').trim(),
    sourceProjectId: String(item?.sourceProjectId || '').trim(),
    sourceAssetId: String(item?.sourceAssetId || '').trim(),
    kind: String(item?.kind || '').trim(),
    name: String(item?.name || '').trim(),
    tags: Array.isArray(item?.tags) ? item.tags.map(tag => String(tag || '').trim()).filter(Boolean) : [],
    sourceLabel: `项目 ${String(item?.sourceProjectId || '').trim()}`
  })).filter(item => item.id && item.sourceProjectId && item.sourceAssetId && ['image', 'video', 'audio'].includes(item.kind));
}
