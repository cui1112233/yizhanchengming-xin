import { normalizeNodeConfig, patchNodeConfig } from './agent-node-config.js';
import { normalizeCanvasNodeType } from './agent-canvas-node-presentation.js';

const shareableKinds = new Set(['text', 'image', 'video', 'audio']);

export function shareModelSelection(nodes = [], nodeID, patch = {}) {
  const selected = nodes.find(node => node.id === nodeID);
  const kind = normalizeCanvasNodeType(selected?.data?.type);
  if (!shareableKinds.has(kind) || !Object.hasOwn(patch, 'provider')) return nodes;

  return nodes.map(node => normalizeCanvasNodeType(node?.data?.type) === kind
    ? { ...node, data: { ...node.data, config: patchNodeConfig(kind, node.data?.config || {}, patch) } }
    : node);
}

export function inheritedModelSelection(nodes = [], kind) {
  const normalizedKind = normalizeCanvasNodeType(kind);
  if (!shareableKinds.has(normalizedKind)) return {};
  const node = [...nodes].reverse().find(item => normalizeCanvasNodeType(item?.data?.type) === normalizedKind && (item.data?.config?.provider || item.data?.config?.modelId || item.data?.config?.model));
  if (!node) return {};
  const config = normalizeNodeConfig(normalizedKind, node.data?.config || {});
  return { provider: config.provider, modelId: config.modelId, model: config.model };
}
