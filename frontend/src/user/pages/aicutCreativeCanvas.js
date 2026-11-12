import { createNodeAt } from './agent-canvas-context-menu.js';
import { saveCanvasDraft } from './agent-canvas-draft-storage.js';

export function persistCreativeCanvasDraft(storage, session, revision = session?.queue?.currentRevision()) {
  if (!session?.queue || !session.canvas || !session.taskId) return false;
  return saveCanvasDraft(storage, session.username, session.taskId, session.canvas, revision);
}

export const creativeCardTypes = [
  { key: 'story', label: '故事', kind: 'text' },
  { key: 'role', label: '角色', kind: 'text' },
  { key: 'scene', label: '场景', kind: 'text' },
  { key: 'prop', label: '道具', kind: 'text' },
  { key: 'storyboard', label: '分镜', kind: 'script' }
];

export function createCreativeCard(key, { id, position = { x: 0, y: 0 } } = {}) {
  const template = creativeCardTypes.find(item => item.key === key);
  if (!template || !String(id || '').trim()) throw new Error('请选择有效卡片');
  const node = createNodeAt({ id, kind: template.kind, position });
  return { id: node.id, type: template.kind, position: node.position, ...node.data, title: template.label };
}

export function transferableCreativeAssets(assets) {
  const seen = new Set();
  return (Array.isArray(assets) ? assets : []).filter(asset => {
    const id = String(asset?.id || '').trim();
    if (!id || seen.has(id) || !['image', 'video', 'audio'].includes(asset?.kind) || !String(asset?.url || '').trim()) return false;
    seen.add(id);
    return true;
  });
}

export function creativeAssetSelection(ids, assets) {
  const available = new Set(transferableCreativeAssets(assets).map(asset => asset.id));
  return [...new Set(Array.isArray(ids) ? ids : [])].filter(id => available.has(id)).slice(0, 30);
}

export function defaultCreativeAssetSelection(assets, nodes = [], selectedNodeIds = []) {
  const selected = new Set(selectedNodeIds);
  const nodeAssets = nodes.filter(node => selected.has(node.id)).map(node => node.assetRef?.id || node.assetRef?.assetId);
  const preferred = creativeAssetSelection(nodeAssets, assets);
  return preferred.length ? preferred : transferableCreativeAssets(assets).slice(0, 30).map(asset => asset.id);
}

export function creativeScriptText(node) {
  if (!['text', 'script'].includes(node?.type)) return '';
  const candidates = [node.type === 'script' ? node.config?.sourceText : '', node.summary, node.config?.prompt];
  return candidates.map(value => String(value || '').trim()).find(Boolean) || '';
}

export function transferableCreativeScriptNodes(nodes) {
  return (Array.isArray(nodes) ? nodes : []).filter(node => node?.id && creativeScriptText(node));
}
