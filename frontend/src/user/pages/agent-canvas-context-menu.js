function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function copiedTitle(title) {
  const value = String(title || '新节点').trim() || '新节点';
  return value.endsWith('副本') ? value : `${value}副本`;
}

export function createNodeAt({ id, kind, position, asset = null }) {
  const type = ['text', 'image', 'video'].includes(kind) ? kind : 'text';
  const media = type === 'image' || type === 'video';
  return {
    id,
    type: 'agent',
    position: { x: Number(position?.x) || 0, y: Number(position?.y) || 0 },
    selected: true,
    data: {
      type,
      title: media ? String(asset?.name || `新${type === 'image' ? '图片' : '视频'}`) : '新文本',
      summary: '',
      version: 1,
      sourceNodeIds: [],
      status: 'draft',
      assetRef: media && asset ? clone(asset) : null
    }
  };
}

export function copyCanvasSelection({ nodes = [], edges = [], selectedIds = [] }) {
  const selected = new Set(selectedIds);
  const copiedNodes = nodes.filter(node => selected.has(node.id)).map(clone);
  if (!copiedNodes.length) return null;
  return {
    nodes: copiedNodes,
    edges: edges.filter(edge => selected.has(edge.source) && selected.has(edge.target)).map(clone)
  };
}

export function pasteCanvasSelection({ clipboard, position, nextID }) {
  if (!clipboard?.nodes?.length || typeof nextID !== 'function') return { nodes: [], edges: [] };
  const originX = Math.min(...clipboard.nodes.map(node => Number(node.position?.x) || 0));
  const originY = Math.min(...clipboard.nodes.map(node => Number(node.position?.y) || 0));
  const targetX = Number(position?.x) || 0;
  const targetY = Number(position?.y) || 0;
  const idMap = new Map(clipboard.nodes.map(node => [node.id, nextID(node.id)]));
  const nodes = clipboard.nodes.map(node => {
    const id = idMap.get(node.id);
    const data = clone(node.data || {});
    data.title = copiedTitle(data.title);
    data.sourceNodeIds = (data.sourceNodeIds || []).filter(sourceID => idMap.has(sourceID)).map(sourceID => idMap.get(sourceID));
    return {
      ...clone(node),
      id,
      position: {
        x: targetX + ((Number(node.position?.x) || 0) - originX),
        y: targetY + ((Number(node.position?.y) || 0) - originY)
      },
      selected: true,
      data
    };
  });
  const edges = (clipboard.edges || []).flatMap(edge => {
    const source = idMap.get(edge.source);
    const target = idMap.get(edge.target);
    if (!source || !target) return [];
    const relation = edge.data?.relation || 'reference';
    return [{ ...clone(edge), id: `edge-${source}-${target}-${relation}`, source, target, selected: false }];
  });
  return { nodes, edges };
}

export function removeCanvasNodes({ nodes = [], edges = [], ids = [] }) {
  const removed = new Set(ids);
  return {
    nodes: nodes.filter(node => !removed.has(node.id)),
    edges: edges.filter(edge => !removed.has(edge.source) && !removed.has(edge.target))
  };
}
