function nodeTitle(node) {
  return String(node?.title || node?.data?.title || '').trim();
}

function activeMentionQuery(prompt) {
  const match = String(prompt || '').match(/(?:^|\s)@([^\s@]*)$/);
  return match ? match[1] : null;
}

export function findNodeMentionCandidates(nodes, prompt, limit = 8) {
  const query = activeMentionQuery(prompt);
  if (query === null) return [];
  const normalizedQuery = query.toLocaleLowerCase();
  return (Array.isArray(nodes) ? nodes : [])
    .filter(node => node?.id && nodeTitle(node))
    .filter(node => nodeTitle(node).toLocaleLowerCase().includes(normalizedQuery))
    .slice(0, limit);
}

export function insertNodeMention(prompt, node) {
  const title = nodeTitle(node);
  if (!title || activeMentionQuery(prompt) === null) return String(prompt || '');
  return String(prompt || '').replace(/@[^\s@]*$/, `@${title} `);
}

export function findAgentMentionCandidates({ nodes = [], assets = [] } = {}, prompt, limit = 8) {
  const query = activeMentionQuery(prompt);
  if (query === null) return [];
  const normalizedQuery = query.toLocaleLowerCase();
  const nodeCandidates = (Array.isArray(nodes) ? nodes : []).flatMap(node => {
    const id = String(node?.id || '').trim();
    const title = nodeTitle(node);
    if (!id || !title || !title.toLocaleLowerCase().includes(normalizedQuery)) return [];
    return [{ mentionSource: 'node', referenceId: id, kind: String(node?.type || node?.data?.type || 'text'), title }];
  });
  const assetCandidates = (Array.isArray(assets) ? assets : []).flatMap(asset => {
    const id = String(asset?.id || asset?.assetId || '').trim();
    const title = String(asset?.name || asset?.title || '').trim();
    if (!id || !title || !title.toLocaleLowerCase().includes(normalizedQuery)) return [];
    return [{ mentionSource: 'asset', referenceId: id, kind: String(asset?.kind || 'file'), title }];
  });
  return [...nodeCandidates, ...assetCandidates].slice(0, limit);
}

export function insertAgentMention(prompt, candidate) {
  return insertNodeMention(prompt, candidate);
}
