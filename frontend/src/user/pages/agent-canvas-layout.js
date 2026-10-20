function nodeID(value) {
  return String(value || '');
}

export function layoutCanvas(nodes = [], edges = [], direction = 'LR') {
  const sourceNodes = Array.isArray(nodes) ? nodes : [];
  if (!sourceNodes.length) return [];
  const ids = new Set(sourceNodes.map(node => nodeID(node.id)));
  const incoming = new Map(sourceNodes.map(node => [nodeID(node.id), 0]));
  const outgoing = new Map(sourceNodes.map(node => [nodeID(node.id), []]));
  for (const edge of Array.isArray(edges) ? edges : []) {
    const source = nodeID(edge?.source);
    const target = nodeID(edge?.target);
    if (!ids.has(source) || !ids.has(target) || source === target) continue;
    outgoing.get(source).push(target);
    incoming.set(target, incoming.get(target) + 1);
  }

  const rank = new Map(sourceNodes.map(node => [nodeID(node.id), 0]));
  const queue = sourceNodes.map(node => nodeID(node.id)).filter(id => incoming.get(id) === 0);
  for (let index = 0; index < queue.length; index += 1) {
    const source = queue[index];
    for (const target of outgoing.get(source)) {
      rank.set(target, Math.max(rank.get(target), rank.get(source) + 1));
      incoming.set(target, incoming.get(target) - 1);
      if (incoming.get(target) === 0) queue.push(target);
    }
  }

  const layerIndex = new Map();
  return sourceNodes.map(node => {
    if (node.locked || node.data?.locked) return node;
    const layer = rank.get(nodeID(node.id)) || 0;
    const itemIndex = layerIndex.get(layer) || 0;
    layerIndex.set(layer, itemIndex + 1);
    const position = direction === 'TB'
      ? { x: itemIndex * 280, y: layer * 260 }
      : { x: layer * 380, y: itemIndex * 220 };
    return { ...node, position };
  });
}
