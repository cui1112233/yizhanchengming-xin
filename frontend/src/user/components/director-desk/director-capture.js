import { createNodeAt } from '../../pages/agent-canvas-context-menu.js';
import { createTypedCanvasEdge, syncCanvasEffectiveReferences } from '../../pages/agent-canvas-connections.js';

export function insertDirectorCapture({ nodes, edges, directorNodeID, asset } = {}) {
  const currentNodes = Array.isArray(nodes) ? nodes : [];
  const currentEdges = Array.isArray(edges) ? edges : [];
  const director = currentNodes.find(node => node.id === directorNodeID && node.data?.type === 'director');
  if (!director) throw new Error('导演台节点不存在，无法放入构图');
  if (!asset?.id || asset.kind !== 'image') throw new Error('导演台构图必须是图片素材');

  const imageNodeID = `image-director-${asset.id}`;
  const existingImageNode = currentNodes.find(node => node.id === imageNodeID);
  const imageNode = existingImageNode || createNodeAt({
    id: imageNodeID,
    kind: 'image',
    asset,
    position: { x: (Number(director.position?.x) || 0) + 374, y: Number(director.position?.y) || 0 }
  });
  let keptImageNode = false;
  const nextNodes = currentNodes.flatMap(node => {
    if (node.id !== imageNodeID) return [{ ...node, selected: false }];
    if (keptImageNode) return [];
    keptImageNode = true;
    return [{ ...node, selected: true }];
  });
  if (!keptImageNode) nextNodes.push(imageNode);
  const edge = createTypedCanvasEdge({ source: director.id, target: imageNode.id }, nextNodes);
  if (!edge) throw new Error('导演台构图连线创建失败');
  let keptEdge = false;
  const nextEdges = currentEdges.filter(current => {
    if (current.source !== director.id || current.target !== imageNode.id) return true;
    if (keptEdge) return false;
    keptEdge = true;
    return true;
  });
  if (!keptEdge) nextEdges.push(edge);
  return { nodes: syncCanvasEffectiveReferences(nextNodes, nextEdges), edges: nextEdges, imageNode };
}
