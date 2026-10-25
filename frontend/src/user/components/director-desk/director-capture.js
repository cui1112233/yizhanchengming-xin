import { createNodeAt } from '../../pages/agent-canvas-context-menu.js';
import { createTypedCanvasEdge, syncCanvasEffectiveReferences } from '../../pages/agent-canvas-connections.js';

export function insertDirectorCapture({ nodes, edges, directorNodeID, asset } = {}) {
  const currentNodes = Array.isArray(nodes) ? nodes : [];
  const currentEdges = Array.isArray(edges) ? edges : [];
  const director = currentNodes.find(node => node.id === directorNodeID && node.data?.type === 'director');
  if (!director) throw new Error('导演台节点不存在，无法放入构图');
  if (!asset?.id || asset.kind !== 'image') throw new Error('导演台构图必须是图片素材');

  const imageNode = createNodeAt({
    id: `image-director-${asset.id}`,
    kind: 'image',
    asset,
    position: { x: (Number(director.position?.x) || 0) + 374, y: Number(director.position?.y) || 0 }
  });
  const nextNodes = currentNodes.map(node => ({ ...node, selected: false })).concat(imageNode);
  const edge = createTypedCanvasEdge({ source: director.id, target: imageNode.id }, nextNodes);
  if (!edge) throw new Error('导演台构图连线创建失败');
  return { nodes: syncCanvasEffectiveReferences(nextNodes, [...currentEdges, edge]), edges: [...currentEdges, edge], imageNode };
}
