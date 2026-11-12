import { apiRequest as baseApiRequest } from './client.js';
import { scopedAICutOptions } from './aicutTimelineScope.js';
const apiRequest = (path, options = {}) => baseApiRequest(path, scopedAICutOptions(path, options));

const safeIdPattern = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;
const supportedRatios = new Set(['9:16', '16:9', '1:1']);

function safeId(value, label) {
  const normalized = String(value || '').trim();
  if (!safeIdPattern.test(normalized)) throw new Error(`${label}无效`);
  return normalized;
}

function planPath(projectId, planId = '', suffix = '') {
  const project = safeId(projectId, 'AI-CUT 项目');
  if (!planId) return `/api/ai-cut/projects/${project}/edit-plans`;
  return `/api/ai-cut/projects/${project}/edit-plans/${safeId(planId, '剪辑草案')}${suffix}`;
}

export async function createAICutEditPlan(projectId, input = {}, request = apiRequest) {
  const modelId = String(input.modelId || '').trim();
  const goal = String(input.goal || '').trim();
  const assetIds = Array.from(new Set((Array.isArray(input.assetIds) ? input.assetIds : []).map(id => safeId(id, '素材'))));
  const targetDurationMs = Number(input.targetDurationMs ?? 0);
  const aspectRatio = String(input.aspectRatio || '').trim();
  if (!modelId) throw new Error('请选择文本模型');
  if (!assetIds.length) throw new Error('请至少选择一个素材');
  if (!goal) throw new Error('剪辑目标不能为空');
  if (!Number.isFinite(targetDurationMs) || targetDurationMs < 0 || targetDurationMs > 7200000) throw new Error('目标时长必须在 0 到 120 分钟之间');
  if (!supportedRatios.has(aspectRatio)) throw new Error('画面比例仅支持 9:16、16:9 或 1:1');
  const clientRequestId = safeId(input.clientRequestId, '请求编号');
  const response = await request(planPath(projectId), {
    method: 'POST', signal: input.signal, timeoutMs: 0,
    body: JSON.stringify({
      clientRequestId, modelId, expectedRevision: Number(input.expectedRevision || 0), assetIds, goal,
      targetDurationMs: Math.round(targetDurationMs), aspectRatio,
      includeCaptions: Boolean(input.includeCaptions), includeTransitions: Boolean(input.includeTransitions),
      includeFilters: Boolean(input.includeFilters), includeBgm: Boolean(input.includeBgm),
      extraRequirement: String(input.extraRequirement || '').trim()
    })
  });
  return response?.plan;
}

export async function getAICutEditPlan(projectId, planId, request = apiRequest) {
  const response = await request(planPath(projectId, planId), { suppressGlobalError: true });
  return response?.plan;
}

export async function cancelAICutEditPlan(projectId, planId, request = apiRequest) {
  const response = await request(planPath(projectId, planId, '/cancel'), { method: 'POST', body: '{}' });
  return response?.plan;
}

export async function applyAICutEditPlan(projectId, planId, enabledOperationIds, request = apiRequest) {
  const ids = (Array.isArray(enabledOperationIds) ? enabledOperationIds : []).map(id => safeId(id, '剪辑操作'));
  return request(planPath(projectId, planId, '/apply'), { method: 'POST', body: JSON.stringify({ enabledOperationIds: ids }) });
}

export async function previewAICutEditPlan(projectId, planId, enabledOperationIds, request = apiRequest) {
 const ids = (Array.isArray(enabledOperationIds) ? enabledOperationIds : []).map(id => safeId(id, '剪辑操作'));
 const result = await request(planPath(projectId,planId,'/preview'), { method: 'POST', suppressGlobalError: true, body: JSON.stringify({enabledOperationIds:ids}) });
 return result?.summary;
}
