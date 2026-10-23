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
