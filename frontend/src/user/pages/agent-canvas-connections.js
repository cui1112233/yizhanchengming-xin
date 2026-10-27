function nodeKind(node) {
  return String(node?.data?.type || node?.type || 'text');
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function relationFor(source, target) {
  const sourceKind = nodeKind(source);
  const targetKind = nodeKind(target);
  if (targetKind === 'video' && sourceKind === 'image') return 'first_frame';
  if (targetKind === 'video' && sourceKind === 'video') return 'sequence';
  if (targetKind === 'video' && sourceKind === 'audio') return 'audio_source';
  if (sourceKind === 'text' && ['image', 'video', 'audio'].includes(targetKind)) return 'script_source';
  return 'reference';
}

export function connectionDropChoices(sourceNode) {
  const sourceKind = nodeKind(sourceNode);
  if (sourceKind === 'image') return [{ kind: 'image', label: '新图片' }, { kind: 'video', label: '新视频' }];
  if (sourceKind === 'video') return [{ kind: 'video', label: '新视频' }, { kind: 'image', label: '提取图片' }];
  if (sourceKind === 'audio') return [{ kind: 'video', label: '新视频' }, { kind: 'audio', label: '新音频' }];
  return [{ kind: 'text', label: '新文本' }, { kind: 'image', label: '新图片' }];
}

export function createTypedCanvasEdge(connection, nodes = []) {
  const sourceID = String(connection?.source || '');
  const targetID = String(connection?.target || '');
  if (!sourceID || !targetID || sourceID === targetID) return null;
  const source = nodes.find(node => node.id === sourceID);
  const target = nodes.find(node => node.id === targetID);
  if (!source || !target) return null;
  const relation = String(connection?.data?.relation || relationFor(source, target));
  return {
    id: `edge-${sourceID}-${targetID}-${relation}`,
    source: sourceID,
    target: targetID,
    data: { relation }
  };
}

function allowedReferenceKinds(targetKind) {
  if (targetKind === 'image') return ['image'];
  if (targetKind === 'director') return ['image'];
  if (targetKind === 'video') return ['image', 'video', 'audio'];
  if (targetKind === 'audio') return ['audio'];
  return [];
}

function uniqueAssets(assets, limit = 10) {
  const seen = new Set();
  return assets.filter(asset => {
    const id = String(asset?.id || '');
    if (!id || seen.has(id)) return false;
    seen.add(id);
    return true;
  }).slice(0, limit);
}

function nodePromptSource(node) {
  const text = String(node?.data?.summary || node?.data?.config?.prompt || '').trim();
  return text ? { nodeId: node.id, text } : null;
}

export function syncDirectorVideoLinks(nodes = [], edges = [], directorNodeID, project = {}) {
  const directorID = String(directorNodeID || '');
  if (!directorID) return nodes;
  const referenceAssetIds = Array.isArray(project.referenceAssetIds)
    ? project.referenceAssetIds
    : [...new Set([project.scene?.assetId, ...(Array.isArray(project.objects) ? project.objects : []).map(item => item?.assetId), ...(Array.isArray(project.actors) ? project.actors : []).map(item => item?.assetId), ...(Array.isArray(project.props) ? project.props : []).map(item => item?.assetId)].filter(Boolean))];
  const linkedVideoIDs = new Set(edges.filter(edge => edge?.source === directorID && nodeKind(nodes.find(node => node.id === edge.target)) === 'video').map(edge => edge.target));
  if (!linkedVideoIDs.size) return nodes;
  return nodes.map(node => {
    if (!linkedVideoIDs.has(node.id)) return node;
    const config = node.data?.config || {};
    const shotPrompt = String(project.shotPrompt || '').trim();
    return {
      ...node,
      data: {
        ...(node.data || {}),
        config: {
          ...config,
          ...(shotPrompt ? { prompt: shotPrompt } : {}),
          directorProject: clone(project),
          directorCamera: clone(project.directorCamera || project.camera || {}),
          referenceAssetIds: [...referenceAssetIds]
        }
      }
    };
  });
}

export function syncEffectiveReferences(node, edges = [], nodes = []) {
  const supportedKinds = allowedReferenceKinds(nodeKind(node));
  if (!supportedKinds.length) return node;
  const config = node?.data?.config || {};
  const previousAutoIDs = new Set([
    ...(Array.isArray(config.autoReferenceAssetIDs) ? config.autoReferenceAssetIDs : []),
    ...(Array.isArray(config.edgeReferences) ? config.edgeReferences.map(item => item?.assetId) : [])
  ]);
  const legacyReferences = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  let manualReferences = uniqueAssets(
    (Array.isArray(config.manualReferences) ? config.manualReferences : legacyReferences.filter(asset => !previousAutoIDs.has(asset?.id)))
      .filter(asset => supportedKinds.includes(asset?.kind))
      .map(clone)
  );
  const nodeByID = new Map(nodes.map(item => [item.id, item]));
  const edgePromptSources = edges.flatMap(edge => {
    if (edge?.target !== node.id) return [];
    const sourceNode = nodeByID.get(edge.source);
    if (nodeKind(sourceNode) !== 'text') return [];
    const source = nodePromptSource(sourceNode);
    return source ? [source] : [];
  });
  const timingAudioEdge = edges.find(edge => {
    if (edge?.target !== node.id || String(edge?.data?.relation || edge?.relation || '') !== 'audio_source') return false;
    return nodeKind(nodeByID.get(edge.source)) === 'audio';
  });
  const timingAudioNode = timingAudioEdge ? nodeByID.get(timingAudioEdge.source) : null;
  const timingAudioAssetID = String(timingAudioNode?.data?.assetRef?.id || timingAudioNode?.assetRef?.id || '');
  if (timingAudioAssetID) manualReferences = manualReferences.filter(asset => String(asset?.id || '') !== timingAudioAssetID);
  const connectedAudioDurationSeconds = Number(
    timingAudioNode?.data?.config?.durationSeconds
      || timingAudioNode?.data?.assetRef?.durationSeconds
      || 0
  );
  const edgeReferences = edges.flatMap(edge => {
    if (edge?.target !== node.id) return [];
    const sourceNode = nodeByID.get(edge.source);
    if (nodeKind(sourceNode) === 'audio' && String(edge?.data?.relation || edge?.relation || '') === 'audio_source') return [];
    const asset = sourceNode?.data?.assetRef || sourceNode?.assetRef;
    if (!asset?.id || !supportedKinds.includes(asset.kind)) return [];
    return [{ assetId: asset.id, sourceNodeId: edge.source, relation: edge.data?.relation || edge.relation || 'reference', asset: clone(asset) }];
  });
  const dedupedEdgeReferences = edgeReferences.filter((item, index, values) => values.findIndex(candidate => candidate.assetId === item.assetId) === index);
  const referenceAssets = uniqueAssets([...manualReferences, ...dedupedEdgeReferences.map(item => item.asset)]);
  const nextConfig = {
    ...config,
    manualReferences,
    edgePromptSources,
    edgeReferences: dedupedEdgeReferences,
    referenceAssets,
    autoReferenceAssetIDs: dedupedEdgeReferences.map(item => item.assetId)
  };
  if (connectedAudioDurationSeconds > 0) nextConfig.connectedAudioDurationSeconds = connectedAudioDurationSeconds;
  else delete nextConfig.connectedAudioDurationSeconds;
  return {
    ...node,
    data: {
      ...(node.data || {}),
      config: nextConfig
    }
  };
}

export function syncCanvasEffectiveReferences(nodes = [], edges = []) {
  return nodes.map(node => syncEffectiveReferences(node, edges, nodes));
}
