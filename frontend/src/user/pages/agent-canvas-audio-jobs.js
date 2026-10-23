import { normalizeNodeConfig } from './agent-node-config.js';

const GENERATION_OPERATIONS = new Set(['speech', 'music', 'sound_effect']);
const PROCESSING_OPERATIONS = new Set(['trim', 'speed', 'smart_split', 'custom_split']);

export function normalizeAudioOperation(value) {
  const operation = String(value || '').trim();
  if (operation === 'effect') return 'sound_effect';
  return GENERATION_OPERATIONS.has(operation) || PROCESSING_OPERATIONS.has(operation) ? operation : 'speech';
}

export function audioOperationNeedsModel(value) {
  return GENERATION_OPERATIONS.has(normalizeAudioOperation(value));
}

export function audioSourceAsset(node) {
  const direct = node?.data?.assetRef;
  if (direct?.kind === 'audio' && String(direct.id || direct.assetId || '').trim()) return direct;
  const references = Array.isArray(node?.data?.config?.referenceAssets) ? node.data.config.referenceAssets : [];
  return references.find(asset => asset?.kind === 'audio' && String(asset.id || asset.assetId || '').trim()) || null;
}

export function audioJobRequestForNode(node, { taskId = '', attempt = 0, modelId = '' } = {}) {
  const config = normalizeNodeConfig('audio', node?.data?.config || {});
  const operation = normalizeAudioOperation(config.parameters.operation);
  const source = audioSourceAsset(node);
  return {
    nodeId: String(node?.id || '').trim(),
    nodeVersion: Number(node?.data?.version) || 0,
    operation,
    prompt: String(config.prompt || '').trim(),
    provider: config.provider,
    modelId: String(modelId || config.modelId).trim(),
    sourceAssetId: String(source?.id || source?.assetId || '').trim(),
    settings: { ...config.parameters },
    advanced: { ...config.advanced },
    idempotencyKey: `agent-audio:${String(taskId || '').trim()}:${String(node?.id || '').trim()}:${Number(node?.data?.version) || 0}:${Number(attempt) || 0}`
  };
}

export function canSubmitAudioNode(node, selectedModel = '') {
  const payload = audioJobRequestForNode(node, { modelId: selectedModel });
  if (audioOperationNeedsModel(payload.operation)) return Boolean(payload.prompt && payload.modelId);
  return Boolean(payload.sourceAssetId);
}
