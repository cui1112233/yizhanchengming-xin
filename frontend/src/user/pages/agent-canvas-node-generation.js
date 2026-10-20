import { normalizeNodeConfig } from './agent-node-config.js';

const KIND_LABELS = Object.freeze({ text: '文字', image: '图片', video: '视频' });
const LOCAL_EXECUTOR_MODELS = Object.freeze({
  image: { value: 'local-doubao-executor-image', label: '豆包本地执行器（图片）' },
  video: { value: 'local-doubao-executor-video', label: '豆包本地执行器（视频）' }
});

export function agentCanvasModelOptions(options = [], kind = '') {
  const local = LOCAL_EXECUTOR_MODELS[kind];
  const normalized = Array.isArray(options) ? options.filter(option => option?.value) : [];
  if (!local) return normalized;
  return [local, ...normalized.filter(option => option.value !== local.value)];
}

export function composerModelSelection(kind, config = {}, models = []) {
  const options = (Array.isArray(models) ? models : [])
    .filter(model => model?.kind === kind && model?.enabled !== false && String(model?.id || '').trim())
    .map(model => ({
      value: String(model.id).trim(),
      label: String(model.displayName || model.name || model.modelId || model.id).trim()
    }));
  const requested = String(config?.modelId || '').trim();
  const modelId = options.some(option => option.value === requested) ? requested : (options[0]?.value || '');
  return {
    modelId,
    options,
    disabled: !modelId,
    message: modelId ? '' : `尚未配置可用的${KIND_LABELS[kind] || ''}模型`
  };
}

export function nodeGenerationViewState({ kind, busy = false, error = '', execution = null } = {}) {
  if (busy || execution?.status === 'running') {
    return { busy: true, disabled: true, tone: 'running', message: `正在生成${KIND_LABELS[kind] || '内容'}…` };
  }
  const message = String(error?.message || error || '').trim();
  if (message) return { busy: false, disabled: false, tone: 'error', message };
  if (execution?.status === 'completed') return { busy: false, disabled: false, tone: 'success', message: '生成完成' };
  return { busy: false, disabled: false, tone: 'idle', message: '' };
}

export function nodeGenerationIdempotencyKey({ taskId, nodeId, nodeVersion, kind, attempt = 0 } = {}) {
  return ['agent-node', taskId, nodeId, kind, Number(nodeVersion) || 0, Number(attempt) || 0]
    .map(value => encodeURIComponent(String(value || '')))
    .join(':');
}

export function buildNodeGenerationRequest(node, { taskId = '', attempt = 0, modelId = '' } = {}) {
  const kind = String(node?.data?.type || '').trim();
  const config = normalizeNodeConfig(kind, node?.data?.config || {});
  const referenceAssetIds = (config.referenceAssets || [])
    .filter(asset => kind === 'image' ? asset?.kind === 'image' : true)
    .map(asset => String(asset?.id || asset?.assetId || '').trim())
    .filter(Boolean);
  return {
    nodeId: String(node?.id || '').trim(),
    nodeVersion: Number(node?.data?.version) || 0,
    provider: config.provider,
    modelId: String(modelId || config.modelId).trim(),
    mode: config.mode,
    prompt: String(config.prompt || '').trim(),
    parameters: { ...config.parameters },
    advanced: { ...config.advanced },
    referenceAssetIds: [...new Set(referenceAssetIds)],
    idempotencyKey: nodeGenerationIdempotencyKey({
      taskId,
      nodeId: node?.id,
      nodeVersion: node?.data?.version,
      kind,
      attempt
    })
  };
}

export function submitNodeGeneration({ event, onSubmit, nodeID, modelID } = {}) {
  event?.stopPropagation?.();
  onSubmit?.(nodeID, modelID);
}

export function composerShouldClose(eventState = {}) {
  if (eventState.insideComposer || eventState.insidePortal) return false;
  return Boolean(eventState.paneClick || eventState.escape || eventState.switchedNode || eventState.nodeDeleted || eventState.projectChanged);
}
