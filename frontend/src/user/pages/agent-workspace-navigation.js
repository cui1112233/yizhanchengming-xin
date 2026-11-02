function nodeField(node, key, fallback = '') {
  return node?.[key] ?? node?.data?.[key] ?? fallback;
}

function selected(values) {
  return Array.isArray(values) ? values.map(value => String(value)) : [];
}

function includesText(values, query) {
  const needle = String(query || '').trim().toLocaleLowerCase();
  if (!needle) return true;
  return values.some(value => String(value || '').toLocaleLowerCase().includes(needle));
}

export function filterWorkspaceNodes(nodes = [], filters = {}) {
  const kinds = selected(filters.kinds);
  const statuses = selected(filters.statuses);
  const ratings = selected(filters.ratings);
  return (Array.isArray(nodes) ? nodes : []).filter(node => {
    const kind = String(nodeField(node, 'type'));
    const status = String(nodeField(node, 'status'));
    const config = nodeField(node, 'config', {}) || {};
    const rating = String(config.rating ?? nodeField(node, 'rating'));
    return includesText([nodeField(node, 'title'), nodeField(node, 'summary')], filters.query)
      && (!kinds.length || kinds.includes(kind))
      && (!statuses.length || statuses.includes(status))
      && (!ratings.length || ratings.includes(rating));
  });
}

export function buildWorkspaceGroupTree(nodes = [], groups = []) {
  const groupItems = (Array.isArray(groups) ? groups : []).map(group => {
    if (typeof group === 'string') return { id: group, title: group };
    return { id: String(group?.id || group?.groupId || ''), title: String(group?.title || group?.name || group?.id || group?.groupId || '未命名分组') };
  }).filter(group => group.id);
  const knownGroupIDs = new Set(groupItems.map(group => group.id));
  const result = groupItems.map(group => ({
    ...group,
    children: nodes.filter(node => String(nodeField(node, 'groupId')) === group.id)
  })).filter(group => group.children.length);
  const ungrouped = nodes.filter(node => {
    const groupID = String(nodeField(node, 'groupId'));
    return !groupID || !knownGroupIDs.has(groupID);
  });
  if (ungrouped.length) result.push({ id: 'ungrouped', title: '未分组', children: ungrouped });
  return result;
}

export function filterWorkspaceAssets(assets = [], filters = {}) {
  const kinds = selected(filters.kinds);
  return (Array.isArray(assets) ? assets : []).filter(asset => (
    includesText([asset?.name, asset?.title, asset?.id], filters.query)
    && (!kinds.length || kinds.includes(String(asset?.kind || '')))
  ));
}

function assetField(asset, key) {
  return asset?.[key] ?? asset?.config?.[key] ?? asset?.metadata?.[key] ?? asset?.source?.[key] ?? '';
}

export function buildWorkspaceAssetGroupTree(assets = [], groups = [], nodes = []) {
  const groupTitles = new Map((Array.isArray(groups) ? groups : []).map(group => [
    String(group?.id || group?.groupId || ''),
    String(group?.title || group?.name || group?.id || group?.groupId || '未命名分组')
  ]));
  const nodeGroups = new Map();
  (Array.isArray(nodes) ? nodes : []).forEach(node => {
    const config = node?.config ?? node?.data?.config ?? {};
    const source = config.source ?? {};
    const assetId = String(config.resourceId || config.assetId || source.resourceId || source.assetId || '').trim();
    const groupId = String(node?.groupId ?? node?.data?.groupId ?? '').trim();
    if (assetId && groupId) nodeGroups.set(assetId, groupId);
  });
  const grouped = new Map();
  const fallbackGroupID = groupTitles.size === 1 ? [...groupTitles.keys()][0] : '';
  (Array.isArray(assets) ? assets : []).forEach(asset => {
    const assetId = String(asset?.id || asset?.assetId || '').trim();
    const rawGroup = String(assetField(asset, 'groupId') || '').trim();
    const bookId = String(assetField(asset, 'bookId') || '').trim();
    const key = rawGroup || nodeGroups.get(assetId) || (bookId ? `book:${bookId}` : fallbackGroupID || 'ungrouped');
    const title = String(
      assetField(asset, 'bookTitle') || assetField(asset, 'bookName') || groupTitles.get(key) || groupTitles.get(rawGroup) || (bookId ? `书籍 ${bookId}` : '未分组')
    );
    if (!grouped.has(key)) grouped.set(key, { id: key, title, children: [] });
    grouped.get(key).children.push(asset);
  });
  return [...grouped.values()];
}

export function workspaceExecutionKind(execution = {}) {
  const value = String(execution?.kind || execution?.mediaKind || execution?.toolName || execution?.tool || '').toLocaleLowerCase();
  if (value.includes('image')) return 'image';
  if (value.includes('video')) return 'video';
  if (value.includes('audio')) return 'audio';
  if (value.includes('text')) return 'text';
  return 'other';
}

export function filterWorkspaceExecutions(executions = [], filters = {}) {
  const kinds = selected(filters.kinds);
  return (Array.isArray(executions) ? executions : []).filter(execution => (
    includesText([execution?.name, execution?.toolName, execution?.tool, execution?.status, execution?.state], filters.query)
    && (!kinds.length || kinds.includes(workspaceExecutionKind(execution)))
  ));
}

export function workspaceDrawerWidth(value) {
  const width = Number(value);
  if (!Number.isFinite(width)) return 288;
  return Math.min(460, Math.max(240, Math.round(width)));
}
