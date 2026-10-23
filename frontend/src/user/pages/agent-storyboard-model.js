import { formatNodeParameterSummary, patchNodeConfig } from './agent-node-config.js';
import { scriptSplitReady } from './agent-script-shots.js';

const SECTION_ORDER = ['shot', 'text', 'audio', 'image', 'video'];
const SECTION_LABELS = { shot: '镜头', text: '文本', audio: '音频', image: '图片', video: '视频' };

function nodeField(node, key, fallback = '') {
  return node?.[key] ?? node?.data?.[key] ?? fallback;
}

function assetID(value) {
  return String(value?.id || value?.assetId || '').trim();
}

function nodeItem(node, kind) {
  const media = nodeField(node, 'assetRef', null);
  return {
    key: `node:${node.id}`,
    id: String(node.id),
    kind,
    source: 'node',
    nodeId: String(node.id),
    title: String(nodeField(node, 'title') || `${SECTION_LABELS[kind]}节点`),
    summary: String(nodeField(node, 'summary')),
    status: String(nodeField(node, 'status') || 'draft'),
    config: nodeField(node, 'config', {}) || {},
    asset: media || null,
    url: String(media?.url || '')
  };
}

function assetItem(asset, node, kind) {
  return {
    key: `asset:${asset.id}`,
    id: String(asset.id),
    kind,
    source: 'asset',
    nodeId: node ? String(node.id) : '',
    title: String(nodeField(node, 'title') || asset.name || asset.title || `${SECTION_LABELS[kind]}素材`),
    summary: String(nodeField(node, 'summary') || asset.name || ''),
    status: String(nodeField(node, 'status') || 'done'),
    config: nodeField(node, 'config', {}) || {},
    asset,
    url: String(asset.url || '')
  };
}

function shotItem(shot) {
  const sequence = Math.max(0, Number(shot?.sequence) || 0);
  const content = [shot?.scene, shot?.dialogue, shot?.narration].map(value => String(value || '').trim()).filter(Boolean);
  return {
    key: `shot:${shot.id}`,
    id: String(shot.id),
    kind: 'shot',
    source: 'shot',
    nodeId: String(shot.scriptNodeId || ''),
    title: `镜头 ${String(sequence).padStart(2, '0')}`,
    summary: content.join(' · '),
    status: 'ready',
    config: {},
    asset: null,
    url: '',
    shot
  };
}

export function currentStoryboardShots(nodes = [], shots = []) {
  const scriptNodes = new Map((Array.isArray(nodes) ? nodes : [])
    .filter(node => String(nodeField(node, 'type')).toLowerCase() === 'script')
    .map(node => [String(node?.id || ''), node]));
  return (Array.isArray(shots) ? shots : []).filter(shot => {
    const scriptNode = scriptNodes.get(String(shot?.scriptNodeId || ''));
    return Boolean(scriptNode && scriptSplitReady(nodeField(scriptNode, 'config', {}) || {}));
  });
}

export function buildStoryboardLibrary(nodes = [], assets = [], shots = []) {
  const safeNodes = Array.isArray(nodes) ? nodes : [];
  const safeAssets = Array.isArray(assets) ? assets : [];
  const safeShots = currentStoryboardShots(safeNodes, shots);
  const linkedNodes = new Map();
  for (const node of safeNodes) {
    const id = assetID(nodeField(node, 'assetRef', null));
    if (id) linkedNodes.set(id, node);
  }

  return SECTION_ORDER.map(kind => {
    if (kind === 'shot') {
      return {
        kind,
        label: SECTION_LABELS[kind],
        items: safeShots
          .filter(shot => shot?.id)
          .slice()
          .sort((left, right) => (Number(left?.sequence) || 0) - (Number(right?.sequence) || 0) || String(left.id).localeCompare(String(right.id)))
          .map(shotItem)
      };
    }
    if (kind === 'text') {
      return {
        kind,
        label: SECTION_LABELS[kind],
        items: safeNodes.filter(node => String(nodeField(node, 'type')) === kind).map(node => nodeItem(node, kind))
      };
    }

    const kindAssets = safeAssets.filter(asset => String(asset?.kind || '') === kind);
    const persistedIDs = new Set(kindAssets.map(assetID).filter(Boolean));
    const nodeOnly = safeNodes
      .filter(node => String(nodeField(node, 'type')) === kind)
      .filter(node => !persistedIDs.has(assetID(nodeField(node, 'assetRef', null))))
      .map(node => nodeItem(node, kind));
    const persisted = kindAssets.map(asset => assetItem(asset, linkedNodes.get(assetID(asset)), kind));
    return { kind, label: SECTION_LABELS[kind], items: [...nodeOnly, ...persisted] };
  });
}

export function storyboardSelectionKey(item) {
  if (!item?.id) return '';
  if (item.source === 'shot') return `shot:${item.id}`;
  return `${item.source === 'asset' ? 'asset' : 'node'}:${item.id}`;
}

export function findStoryboardSelection(library = [], key = '') {
  const normalizedKey = String(key || '').trim();
  if (!normalizedKey) return null;
  const items = (Array.isArray(library) ? library : []).flatMap(section => section.items || []);
  return items.find(item => storyboardSelectionKey(item) === normalizedKey) || null;
}

export function storyboardItemMetaLabel(item = {}) {
  if (item?.source === 'shot') {
    return [item?.shot?.scene, item?.shot?.characters].map(value => String(value || '').trim()).filter(Boolean).join(' · ');
  }
  if (item?.nodeId) {
    const model = String(item?.config?.modelId || item?.config?.model || '').trim();
    const summary = formatNodeParameterSummary(item.kind, item.config || {});
    return [model, summary].filter(Boolean).join(' · ');
  }
  const contentType = String(item?.asset?.contentType || '').trim();
  const subtype = contentType.includes('/') ? contentType.split('/').pop().toUpperCase() : '';
  const byteSize = Number(item?.asset?.byteSize);
  const size = Number.isFinite(byteSize) && byteSize > 0 ? `${Math.max(1, Math.round(byteSize / 1024))} KB` : '';
  return [subtype, size].filter(Boolean).join(' · ');
}

export function patchStoryboardNodeConfig(nodes = [], nodeId = '', patch = {}) {
  const safeNodes = Array.isArray(nodes) ? nodes : [];
  let changed = false;
  const next = safeNodes.map(node => {
    if (String(node?.id || '') !== String(nodeId || '')) return node;
    changed = true;
    const kind = String(nodeField(node, 'type') || 'text');
    const config = patchNodeConfig(kind, nodeField(node, 'config', {}) || {}, patch);
    if (node?.data && typeof node.data === 'object') {
      return { ...node, data: { ...node.data, config } };
    }
    return { ...node, config };
  });
  return changed ? next : safeNodes;
}

export function buildStoryboardAssetContext(assets = [], selectedIDs = []) {
  const byID = new Map((Array.isArray(assets) ? assets : []).map(asset => [assetID(asset), asset]));
  const seen = new Set();
  return (Array.isArray(selectedIDs) ? selectedIDs : []).flatMap(value => {
    const id = String(value || '').trim();
    if (!id || seen.has(id) || !byID.has(id)) return [];
    seen.add(id);
    const asset = byID.get(id);
    return [{
      id,
      assetId: id,
      kind: String(asset.kind || ''),
      name: String(asset.name || asset.title || '未命名素材'),
      ...(asset.contentType ? { contentType: String(asset.contentType) } : {}),
      ...(Number.isFinite(Number(asset.byteSize)) ? { byteSize: Number(asset.byteSize) } : {})
    }];
  });
}
