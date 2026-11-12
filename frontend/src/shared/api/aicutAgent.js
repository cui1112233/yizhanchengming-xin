import { apiRequest as baseApiRequest } from './client.js';
import { scopedAICutOptions } from './aicutTimelineScope.js';
const apiRequest = (path, options = {}) => baseApiRequest(path, scopedAICutOptions(path, options));

function safeID(value, label) {
  const id = String(value || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(id)) throw new Error(`${label}无效`);
  return id;
}

function agentPath(projectId, suffix = '') {
  return `/api/ai-cut/projects/${safeID(projectId, 'AI-CUT 项目')}${suffix}`;
}

export async function createAICutAgentRun(projectId, input, request = apiRequest) {
  const instruction = String(input?.instruction || '').trim();
  if (!instruction) throw new Error('请输入剪辑要求');
  const response = await request(agentPath(projectId, '/agent-runs'), {
    method: 'POST',
    body: JSON.stringify({
      instruction,
      mode: input?.mode === 'ask' ? 'ask' : 'agent',
      modelId: String(input?.modelId || '').trim(),
      activeSkillId: String(input?.activeSkillId || '').trim(),
      expectedRevision: Number(input?.expectedRevision || 0),
      clientRequestId: String(input?.clientRequestId || '').trim()
    })
  });
  return response?.run;
}

export async function getAICutAgentRun(projectId, runId, request = apiRequest) {
  const response = await request(agentPath(projectId, `/agent-runs/${safeID(runId, 'Agent 任务')}`));
  return response?.run;
}

export async function listAICutAgentEvents(projectId, runId, after = 0, request = apiRequest) {
  const sequence = Math.max(0, Number(after) || 0);
  const response = await request(agentPath(projectId, `/agent-runs/${safeID(runId, 'Agent 任务')}/events?after=${sequence}`));
  return Array.isArray(response?.events) ? response.events : [];
}

export async function listAICutAgentMessages(projectId, limit = 100, request = apiRequest) {
  const count = Math.min(200, Math.max(1, Number(limit) || 100));
  const response = await request(agentPath(projectId, `/agent-messages?limit=${count}`));
  return Array.isArray(response?.messages) ? response.messages : [];
}

export async function cancelAICutAgentRun(projectId, runId, request = apiRequest) {
  const response = await request(agentPath(projectId, `/agent-runs/${safeID(runId, 'Agent 任务')}/cancel`), { method: 'POST', body: '{}' });
  return response?.run;
}
