import { normalizeNodeConfig } from './agent-node-config.js';

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function copiedTitle(title) {
  const value = String(title || '新节点').trim() || '新节点';
  return value.endsWith('副本') ? value : `${value}副本`;
}

export function readableAssetRequestError(error, action = 'load') {
  const message = String(error?.message || '').trim();
  if (/^(failed to fetch|network request failed|networkerror)/i.test(message)) {
    return action === 'upload' ? '图片上传连接中断，请重试' : '素材读取连接中断，请重试';
  }
  return message || (action === 'upload' ? '素材上传失败' : '素材库读取失败');
}

export function shouldOpenConnectionDrop({ sourceID, connectionCompleted, droppedOnHandle }) {
  return Boolean(String(sourceID || '').trim()) && !connectionCompleted && !droppedOnHandle;
}

export function createNodeAt({ id, kind, position, asset = null, config = {} }) {
  const type = ['script', 'text', 'image', 'video', 'audio'].includes(kind) ? kind : 'text';
  const media = type === 'image' || type === 'video' || type === 'audio';
  const defaults = type === 'script'
    ? { title: '新剧本', config: { sourceText: '', splitModelId: '', splitState: { state: 'idle', errorMessage: '' } } }
    : type === 'audio'
    ? { title: '新音频', config: { prompt: '', voice: '', speed: 1 } }
    : type === 'image'
      ? { title: '新图片', config: { prompt: '', model: '', ratio: '16:9', quality: 'standard', resolution: '2K', count: 1, referenceAssets: [] } }
      : type === 'video'
        ? { title: '新视频', config: { prompt: '', model: '', ratio: '', durationSeconds: 0, referenceAssets: [] } }
        : { title: '新文本', config: {} };
  return {
    id,
    type: 'agent',
    position: { x: Number(position?.x) || 0, y: Number(position?.y) || 0 },
    selected: true,
    data: {
      type,
      title: media ? String(asset?.name || defaults.title) : defaults.title,
      summary: '',
      version: 1,
      sourceNodeIds: [],
      status: 'draft',
      assetRef: media && asset ? clone(asset) : null,
      config: type === 'script' ? { ...defaults.config } : normalizeNodeConfig(type, { ...defaults.config, ...config })
    }
  };
}

export function appendReferenceAsset({ node, asset, limit = 10 }) {
  const allowedKinds = node?.data?.type === 'image'
    ? ['image']
    : node?.data?.type === 'video'
      ? ['image', 'video']
      : node?.data?.type === 'audio'
        ? ['audio']
        : [];
  if (!allowedKinds.includes(asset?.kind) || !String(asset?.id || '').trim()) return node;
  const config = node.data?.config || {};
  const existing = Array.isArray(config.referenceAssets)
    ? config.referenceAssets.filter(item => allowedKinds.includes(item?.kind) && String(item?.id || '').trim())
    : [];
  const automaticIDs = new Set([
    ...(Array.isArray(config.autoReferenceAssetIDs) ? config.autoReferenceAssetIDs : []),
    ...(Array.isArray(config.edgeReferences) ? config.edgeReferences.map(item => item?.assetId) : [])
  ]);
  const manualReferences = (Array.isArray(config.manualReferences)
    ? config.manualReferences
    : existing.filter(item => !automaticIDs.has(item?.id)))
    .filter(item => allowedKinds.includes(item?.kind) && String(item?.id || '').trim());
  if (existing.some(item => item.id === asset.id) || existing.length >= limit) return node;
  return {
    ...node,
    data: {
      ...node.data,
      config: {
        ...config,
        manualReferences: [...manualReferences, clone(asset)],
        referenceAssets: [...existing, clone(asset)]
      }
    }
  };
}

export function removeReferenceAsset({ node, assetID }) {
  const id = String(assetID || '').trim();
  if (!node?.data || !id) return node;
  const config = node.data.config || {};
  const withoutAsset = assets => Array.isArray(assets) ? assets.filter(asset => asset?.id !== id) : [];
  return {
    ...node,
    data: {
      ...node.data,
      config: {
        ...config,
        manualReferences: withoutAsset(config.manualReferences),
        referenceAssets: withoutAsset(config.referenceAssets),
        autoReferenceAssetIDs: (config.autoReferenceAssetIDs || []).filter(currentID => currentID !== id)
      }
    }
  };
}

function referenceKindsForNode(type) {
  if (type === 'image') return ['image'];
  if (type === 'video') return ['image', 'video'];
  if (type === 'audio') return ['audio'];
  return [];
}

function uniqueAssets(assets, limit = 10) {
  const seen = new Set();
  return assets.filter(asset => {
    const id = String(asset?.id || '').trim();
    if (!id || seen.has(id)) return false;
    seen.add(id);
    return true;
  }).slice(0, limit);
}

export function syncNodeReferencesFromEdges({ nodes = [], edges = [] }) {
  const nodeByID = new Map(nodes.map(node => [node.id, node]));
  return nodes.map(node => {
    const supportedKinds = referenceKindsForNode(node?.data?.type);
    if (!supportedKinds.length) return node;

    const config = node.data?.config || {};
    const existing = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
    const previousAutoIDs = new Set(Array.isArray(config.autoReferenceAssetIDs) ? config.autoReferenceAssetIDs : []);
    const manualAssets = existing.filter(asset => !previousAutoIDs.has(asset?.id));
    const automaticAssets = edges
      .filter(edge => edge?.target === node.id)
      .map(edge => nodeByID.get(edge.source)?.data?.assetRef)
      .filter(asset => supportedKinds.includes(asset?.kind));
    const referenceAssets = uniqueAssets([...manualAssets, ...automaticAssets]);
    const autoReferenceAssetIDs = referenceAssets
      .filter(asset => automaticAssets.some(candidate => candidate?.id === asset.id))
      .map(asset => asset.id);
    const existingIDs = existing.map(asset => asset?.id || '');
    const changed = existingIDs.join('|') !== referenceAssets.map(asset => asset.id).join('|') ||
      (Array.isArray(config.autoReferenceAssetIDs) ? config.autoReferenceAssetIDs : []).join('|') !== autoReferenceAssetIDs.join('|');
    if (!changed) return node;
    return {
      ...node,
      data: {
        ...node.data,
        config: { ...config, referenceAssets: referenceAssets.map(clone), autoReferenceAssetIDs }
      }
    };
  });
}

export function attachImageAsset({ node, asset }) {
  if (node?.data?.type !== 'image' || asset?.kind !== 'image' || !String(asset?.id || '').trim()) return node;
  return {
    ...node,
    data: {
      ...node.data,
      title: String(asset.name || node.data.title || '图片'),
      assetRef: clone(asset)
    }
  };
}

export function createLinkedNode({ source, id, kind }) {
  if (!source?.id) throw new Error('source node is required');
  const position = {
    x: (Number(source.position?.x) || 0) + 256,
    y: Number(source.position?.y) || 0
  };
  const node = createNodeAt({ id, kind, position });
  node.data.sourceNodeIds = [source.id];
  return {
    node,
    edge: {
      id: `edge-${source.id}-${node.id}-reference`,
      source: source.id,
      target: node.id,
      data: { relation: 'reference' }
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
  const remainingNodes = nodes.filter(node => !removed.has(node.id));
  const remainingEdges = edges.filter(edge => !removed.has(edge.source) && !removed.has(edge.target));
  return {
    nodes: syncNodeReferencesFromEdges({ nodes: remainingNodes, edges: remainingEdges }),
    edges: remainingEdges
  };
}
