import { apiRequest } from './client.js';

const AGENT_ASSET_CONTENT_PATTERN = /^\/api\/agent\/tasks\/[^/?#]+\/assets\/[^/?#]+\/content$/;

export function getAgentAssetBlob(assetUrl) {
  const path = String(assetUrl || '').trim();
  if (!AGENT_ASSET_CONTENT_PATTERN.test(path)) return Promise.reject(new Error('Agent asset URL is not a local protected asset'));
  return apiRequest(path, { responseType: 'blob', silent: true, suppressGlobalError: true });
}

export function listAgentTasks() {
  return apiRequest('/api/agent/tasks');
}

export function createAgentTask() {
  return apiRequest('/api/agent/tasks', { method: 'POST' });
}

export function getAgentTask(id) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(id)}`);
}

export function renameAgentTask(id, title) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(id)}`, { method: 'PATCH', body: JSON.stringify({ title }) });
}

export function clearAgentTask(id) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(id)}/messages`, { method: 'DELETE' });
}

export function deleteAgentTask(id) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export function getAgentCanvas(taskId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/canvas`);
}

export function saveAgentCanvas(taskId, canvas, revision) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/canvas`, {
    method: 'PATCH',
    body: JSON.stringify({ canvas, revision })
  });
}

export function listAgentExecutions(taskId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/executions`);
}

export function listAgentAssets(taskId, kind = '') {
  const query = kind ? `?kind=${encodeURIComponent(kind)}` : '';
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/assets${query}`);
}

export function uploadAgentAsset(taskId, file) {
  const body = new FormData();
  body.append('file', file);
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/assets`, { method: 'POST', body });
}

export function createAgentVideoJob(taskId, payload) {
  return requestAgentVideoJob(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs`, {
    method: 'POST', body: JSON.stringify(payload)
  });
}

export function listAgentVideoJobs(taskId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs`, { suppressGlobalError: true });
}

export function getAgentVideoJob(taskId, jobId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs/${encodeURIComponent(jobId)}`, { suppressGlobalError: true });
}

export function confirmAgentVideoJob(taskId, jobId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs/${encodeURIComponent(jobId)}/confirm`, {
    method: 'POST', suppressGlobalError: true
  });
}

export function cancelAgentVideoJob(taskId, jobId) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs/${encodeURIComponent(jobId)}/cancel`, {
    method: 'POST', suppressGlobalError: true
  });
}

export function retryAgentVideoJob(taskId, jobId, idempotencyKey) {
  return requestAgentVideoJob(`/api/agent/tasks/${encodeURIComponent(taskId)}/video-jobs/${encodeURIComponent(jobId)}/retry`, {
    method: 'POST', body: JSON.stringify({ idempotencyKey })
  });
}

async function requestAgentVideoJob(path, options) {
  const result = await apiRequest(path, { ...options, allowStatuses: [httpBadGateway], suppressGlobalError: true });
  if (result?.job) return result;
  const error = new Error(result?.error || '视频任务提交失败');
  error.status = httpBadGateway;
  throw error;
}

const httpBadGateway = 502;

export function listAgentSkills() {
  return apiRequest('/api/agent/skills');
}

export function getMyAgentSkill(id) {
  return apiRequest(`/api/agent/skills/mine/${encodeURIComponent(id)}`);
}

export function createMyAgentSkill(input) {
  return apiRequest('/api/agent/skills/mine', { method: 'POST', body: JSON.stringify(input) });
}

export function updateMyAgentSkill(id, input) {
  return apiRequest(`/api/agent/skills/mine/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteMyAgentSkill(id) {
  return apiRequest(`/api/agent/skills/mine/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

function text(value, limit = 240) {
  return String(value ?? '').trim().slice(0, limit);
}

function safeSelectionMeta(selection) {
  const meta = selection?.meta;
  if (!meta || typeof meta !== 'object' || Array.isArray(meta)) return null;
  const output = {};
  const textFields = ['fieldId', 'fieldLabel', 'text', 'input', 'name', 'description', 'appearance', 'voice', 'style'];
  const numberFields = ['start', 'end', 'totalLength', 'index', 'speed', 'pitch'];
  for (const key of textFields) {
    if (typeof meta[key] === 'string') output[key] = text(meta[key], key === 'text' || key === 'input' || key === 'description' || key === 'appearance' ? 1600 : 240);
  }
  for (const key of numberFields) {
    if (Number.isFinite(Number(meta[key]))) output[key] = Number(meta[key]);
  }
  // 人物/场景卡的完整可见字段是 CM 撰写候选提示词的事实基线。
  // 这里先做一次有界序列化，后端仍会再按统一上下文规则过滤敏感字段。
  if (meta.data && typeof meta.data === 'object' && !Array.isArray(meta.data)) {
    const data = {};
    for (const [key, value] of Object.entries(meta.data).slice(0, 20)) {
      if (/token|key|secret|password|authorization|credential/i.test(key)) continue;
      if (typeof value === 'string') data[key] = text(value, 1600);
      else if (typeof value === 'number' || typeof value === 'boolean') data[key] = value;
    }
    if (Object.keys(data).length) output.data = data;
  }
  return Object.keys(output).length ? output : null;
}

function safeCmProject(value) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null;
  const output = {
    id: text(value.id, 160),
    name: text(value.name || value.label, 160),
    historyId: text(value.historyId, 160),
    sourceTargetId: text(value.sourceTargetId, 180)
  };
  for (const key of ['sourceLength', 'characterCount', 'sceneCount', 'shotCount', 'segmentCount']) {
    if (Number.isFinite(Number(value[key]))) output[key] = Number(value[key]);
  }
  if (typeof value.sourceDirty === 'boolean') output.sourceDirty = value.sourceDirty;
  return Object.values(output).some(item => item !== '' && item !== undefined) ? output : null;
}

function prepareAgentContext(context) {
  if (!context || typeof context !== 'object') return context;
  const selection = context.cmSelection && typeof context.cmSelection === 'object' ? context.cmSelection : null;
  const project = safeCmProject(context.cmProject);
  const capabilities = Array.isArray(context.cmCapabilities)
    ? context.cmCapabilities.map(value => String(value || '').trim()).filter(Boolean).slice(0, 20)
    : [];
  const contract = context.cmResponseContract && typeof context.cmResponseContract === 'object'
    ? context.cmResponseContract
    : null;

  if (!selection && !project && !capabilities.length && !contract) return context;

  const selectionLabel = text(selection?.label || selection?.name, 120);
  const selectionType = text(selection?.type, 48);
  const selectionId = text(selection?.id, 160);
  const selectionMeta = safeSelectionMeta(selection);
  const selectedText = text(selectionMeta?.text || selectionMeta?.input, 1200);
  const cmSummary = [
    project ? `CM 当前工作区：${project.name || '当前工作区'}${project.id ? `；工作区ID ${project.id}` : ''}${project.sourceTargetId ? `；整段原文 targetId ${project.sourceTargetId}` : ''}` : '',
    selection ? `CM 当前选中对象：${selectionLabel || '未命名对象'}${selectionType ? `；类型 ${selectionType}` : ''}${selectionId ? `；ID ${selectionId}` : ''}` : '',
    selectedText ? `CM 当前选中的实际内容：${selectedText}` : '',
    selectionMeta?.data ? `CM 当前选中对象的完整字段（修改后必须保留未要求变更的有效信息）：${JSON.stringify(selectionMeta.data)}` : '',
    capabilities.length ? `CM 当前页面允许申请的动作：${capabilities.join('、')}` : '',
    contract ? 'CM 交互约定：先正常回答；只有用户明确要求修改且动作在允许范围内时，才在回答末尾追加 ```cm-actions JSON```。动作不是名称占位：必须把可以直接写入的完整候选提示词放进 patch。人物使用 character.update，patch 至少包含完整“外形”或“appearance”；场景使用 scene.update，patch 至少包含完整“场景描述”或“description”；约束使用 constraint.update，payload.category 指定 baseSetup/prefix/quality/restriction/negative，payload.value.body 必须是完整提示词。格式为 {"summary":"修改摘要","actions":[{"type":"动作类型","targetId":"实体ID","label":"给用户看的动作名称","patch":{}}]}。不能声称已执行，必须等待用户点击应用。' : ''
  ].filter(Boolean).join('\n');

  const sourceEntities = context.entities && typeof context.entities === 'object' && !Array.isArray(context.entities)
    ? context.entities
    : {};
  const sourceActions = Array.isArray(context.actions) ? context.actions : [];

  return {
    ...context,
    summary: [String(context.summary || '').trim(), cmSummary].filter(Boolean).join('\n').slice(0, 2400),
    entities: {
      ...sourceEntities,
      ...(project ? { cmProject: project } : {}),
      ...(selection ? { cmSelection: { type: selectionType, id: selectionId, label: selectionLabel, ...(selectionMeta ? { meta: selectionMeta } : {}) } } : {})
    },
    actions: [...sourceActions, ...capabilities.map(action => `CM 可申请：${action}`)].slice(0, 8)
  };
}

export function askAgent({ taskId, prompt, context, skillIds = [], suppressGlobalError = false }) {
  return apiRequest('/api/agent/chat', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, context: prepareAgentContext(context), skillIds }),
    suppressGlobalError
  });
}
