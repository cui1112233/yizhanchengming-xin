const OPERATION_LABELS = {
  image_generation: '图片生成',
  text_to_image: '图片生成',
  video_generation: '视频生成',
  image_to_video: '视频生成',
  text_to_video: '视频生成',
  audio_generation: '音频生成',
  text_to_speech: '音频生成',
  tts: '音频生成',
  text_generation: '文本生成'
};

const KIND_LABELS = { image: '图片生成', video: '视频生成', audio: '音频生成', text: '文本生成' };

function text(value) {
  return String(value ?? '').trim();
}

function nodeField(node, key, fallback = undefined) {
  return node?.[key] ?? node?.data?.[key] ?? fallback;
}

function assetID(value) {
  return text(value?.id || value?.assetId);
}

function nodeSource(node) {
  const config = nodeField(node, 'config', {}) || {};
  return config.source && typeof config.source === 'object' ? config.source : {};
}

function generationSource(generation) {
  const source = generation?.settings?.source;
  return source && typeof source === 'object' ? source : {};
}

function sourceFor(node, generation) {
  return { ...nodeSource(node), ...generationSource(generation) };
}

function sourceSequence(source = {}) {
  const value = Number(source.sequence ?? source.shotSequence);
  if (Number.isFinite(value) && value > 0) return Math.round(value);
  const match = text(source.shotId).match(/(\d+)(?!.*\d)/);
  return match ? Number(match[1]) : 0;
}

function operationLabel(source = {}, generation = {}, node = {}) {
  const operation = text(source.operation || generation?.mode).toLowerCase();
  if (OPERATION_LABELS[operation]) return OPERATION_LABELS[operation];
  const kind = text(generation?.kind || nodeField(node, 'type')).toLowerCase();
  return KIND_LABELS[kind] || '生成结果';
}

export function lineageLabel(node = {}, generation = {}) {
  const source = sourceFor(node, generation);
  const sequence = sourceSequence(source);
  const shotLabel = sequence ? `镜头 ${String(sequence).padStart(2, '0')}` : `镜头 ${text(source.shotId) || '--'}`;
  const attempt = Math.max(1, Number(source.attempt) || 1);
  return `${shotLabel} · ${operationLabel(source, generation, node)} · 第 ${attempt} 次`;
}

export function locateShotNode(nodes = [], shotId = '', shot = {}) {
  const safeNodes = Array.isArray(nodes) ? nodes : [];
  const normalizedShotID = text(shotId);
  if (!normalizedShotID) return null;
  const derived = safeNodes.filter(node => text(nodeSource(node).shotId) === normalizedShotID);
  for (const node of derived) {
    const scriptNodeID = text(nodeSource(node).scriptNodeId);
    const scriptNode = safeNodes.find(candidate => (
      text(candidate?.id) === scriptNodeID && text(nodeField(candidate, 'type')).toLowerCase() === 'script'
    ));
    if (scriptNode) return scriptNode;
  }
  const fallbackNodeID = text(shot?.scriptNodeId || shot?.nodeId);
  if (fallbackNodeID) {
    const fallbackNode = safeNodes.find(candidate => text(candidate?.id) === fallbackNodeID);
    if (fallbackNode) return fallbackNode;
  }
  return derived[0] || null;
}

function shotGroup(shot) {
  return {
    id: text(shot?.id),
    shot,
    scriptNodeId: text(shot?.scriptNodeId),
    sequence: Number(shot?.sequence) || 0,
    title: `镜头 ${String(Math.max(0, Number(shot?.sequence) || 0)).padStart(2, '0')}`,
    items: []
  };
}

function resultAsset(result, node, assetsByID) {
  const id = text(result?.assetId || result?.asset?.id || result?.id);
  const current = nodeField(node, 'assetRef', null);
  return assetsByID.get(id) || (assetID(current) === id ? current : null) || result?.asset || (id ? { id, kind: result?.kind } : null);
}

function resultItem({ shot, node, generation, result, assetsByID }) {
  const nodeID = text(generation?.nodeId || node?.id);
  const source = { ...sourceFor(node, generation), sequence: Number(shot?.sequence) || sourceSequence(sourceFor(node, generation)) };
  const asset = resultAsset(result, node, assetsByID);
  const resultAssetID = assetID(asset) || text(result?.assetId || result?.id);
  const activeAssetID = assetID(nodeField(node, 'assetRef', null));
  const labelNode = {
    ...node,
    config: { ...(nodeField(node, 'config', {}) || {}), source }
  };
  return {
    id: `${text(generation?.id) || nodeID}:${text(result?.id) || resultAssetID || 'result'}`,
    generationId: text(generation?.id),
    resultId: text(result?.id),
    nodeId: nodeID,
    kind: text(result?.kind || generation?.kind || nodeField(node, 'type')),
    status: text(generation?.state || generation?.status || nodeField(node, 'status') || 'unknown'),
    attempt: Math.max(1, Number(source.attempt) || 1),
    source,
    asset,
    active: Boolean(activeAssetID && resultAssetID && activeAssetID === resultAssetID),
    label: lineageLabel(labelNode, generation),
    createdAt: result?.createdAt || generation?.updatedAt || generation?.createdAt || '',
    title: text(asset?.name || asset?.title || nodeField(node, 'title') || '生成结果')
  };
}

function compareItems(left, right) {
  if (left.active !== right.active) return left.active ? -1 : 1;
  if (left.attempt !== right.attempt) return right.attempt - left.attempt;
  return text(right.createdAt).localeCompare(text(left.createdAt));
}

export function groupResultLibrary({ shots = [], nodes = [], generations = [], assets = [] } = {}) {
  const safeShots = (Array.isArray(shots) ? shots : []).filter(shot => text(shot?.id));
  const safeNodes = Array.isArray(nodes) ? nodes : [];
  const safeGenerations = Array.isArray(generations) ? generations : [];
  const groups = safeShots
    .slice()
    .sort((left, right) => (Number(left?.sequence) || 0) - (Number(right?.sequence) || 0) || text(left?.id).localeCompare(text(right?.id)))
    .map(shotGroup);
  const groupsByID = new Map(groups.map(group => [group.id, group]));
  const nodesByID = new Map(safeNodes.map(node => [text(node?.id), node]));
  const assetsByID = new Map((Array.isArray(assets) ? assets : []).map(asset => [assetID(asset), asset]).filter(([id]) => id));
  const representedActive = new Set();

  for (const generation of safeGenerations) {
    const node = nodesByID.get(text(generation?.nodeId));
    const source = sourceFor(node, generation);
    const group = groupsByID.get(text(source.shotId));
    if (!group) continue;
    for (const result of Array.isArray(generation?.results) ? generation.results : []) {
      const item = resultItem({ shot: group.shot, node, generation, result, assetsByID });
      group.items.push(item);
      if (item.active) representedActive.add(`${item.nodeId}:${assetID(item.asset)}`);
    }
  }

  for (const node of safeNodes) {
    const source = nodeSource(node);
    const group = groupsByID.get(text(source.shotId));
    const activeAsset = nodeField(node, 'assetRef', null);
    const activeAssetID = assetID(activeAsset);
    if (!group || !activeAssetID || representedActive.has(`${text(node?.id)}:${activeAssetID}`)) continue;
    group.items.push(resultItem({
      shot: group.shot,
      node,
      generation: { nodeId: node?.id, kind: nodeField(node, 'type'), state: nodeField(node, 'status'), settings: { source } },
      result: { id: activeAssetID, assetId: activeAssetID, kind: nodeField(node, 'type'), asset: activeAsset },
      assetsByID
    }));
  }

  for (const group of groups) group.items.sort(compareItems);
  return groups;
}
