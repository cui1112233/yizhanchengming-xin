import { apiRequest } from './client.js';

const BASE = '/api/agent/proxy';

export function listAgentProxyWorkspaces() {
  return apiRequest(`${BASE}/workspaces`);
}

export function createAgentProxyWorkspace({ name, description = '', permissionMode = 'plan' }) {
  return apiRequest(`${BASE}/workspaces`, {
    method: 'POST',
    body: JSON.stringify({ name, description, permissionMode })
  });
}

export function listAgentProxySessions(workspaceId) {
  return apiRequest(`${BASE}/workspaces/${encodeURIComponent(workspaceId)}/sessions`);
}

export function createAgentProxySession(workspaceId, { title, permissionMode = '' } = {}) {
  return apiRequest(`${BASE}/workspaces/${encodeURIComponent(workspaceId)}/sessions`, {
    method: 'POST',
    body: JSON.stringify({ title, permissionMode })
  });
}

export function updateAgentProxySession(sessionId, patch) {
  return apiRequest(`${BASE}/sessions/${encodeURIComponent(sessionId)}`, {
    method: 'PATCH',
    body: JSON.stringify(patch)
  });
}

export function listAgentProxyMessages(sessionId, limit = 100) {
  return apiRequest(`${BASE}/sessions/${encodeURIComponent(sessionId)}/messages?limit=${encodeURIComponent(limit)}`);
}

export function appendAgentProxyMessage(sessionId, { role = 'user', content, metadata } = {}) {
  return apiRequest(`${BASE}/sessions/${encodeURIComponent(sessionId)}/messages`, {
    method: 'POST',
    body: JSON.stringify({ role, content, metadata })
  });
}

export function listAgentProxySkills(workspaceId) {
  return apiRequest(`${BASE}/workspaces/${encodeURIComponent(workspaceId)}/skills`);
}

export function listAgentProxyScheduledJobs(workspaceId) {
  return apiRequest(`${BASE}/workspaces/${encodeURIComponent(workspaceId)}/scheduled-jobs`);
}
