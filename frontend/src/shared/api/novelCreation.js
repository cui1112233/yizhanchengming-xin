import { apiRequest } from './client.js';

export function listNovelCreationModes() {
  return apiRequest('/api/novel-creation/modes');
}

export function listNovelCreationProjects() {
  return apiRequest('/api/novel-creation/projects');
}

export function createNovelCreationProject(input) {
  return apiRequest('/api/novel-creation/projects', { method: 'POST', body: JSON.stringify(input) });
}

export function listNovelCreationSessions(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/sessions`);
}

export function createNovelCreationSession(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/sessions`, { method: 'POST', body: JSON.stringify(input) });
}

export function sendNovelCreationMessage(sessionId, input) {
  return apiRequest(`/api/novel-creation/sessions/${encodeURIComponent(sessionId)}/messages`, { method: 'POST', body: JSON.stringify(input) });
}

export function getNovelCreationRun(runId, afterEventId = 0, signal) {
  return apiRequest(`/api/novel-creation/runs/${encodeURIComponent(runId)}?afterEventId=${Number(afterEventId) || 0}`, { signal, suppressGlobalError: true });
}

export function retryNovelCreationRun(runId) {
  return apiRequest(`/api/novel-creation/runs/${encodeURIComponent(runId)}/retry`, { method: 'POST', body: '{}' });
}

export function cancelNovelCreationRun(runId) {
  return apiRequest(`/api/novel-creation/runs/${encodeURIComponent(runId)}/cancel`, { method: 'POST', body: '{}' });
}

export function confirmNovelCreationRun(runId, blueprint) {
  const body = blueprint ? JSON.stringify({ blueprint }) : '{}';
  return apiRequest(`/api/novel-creation/runs/${encodeURIComponent(runId)}/confirm`, { method: 'POST', body });
}

export function getNovelCreationProjectSettings(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/settings`);
}

export function updateNovelCreationProjectSettings(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/settings`, {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}

export function exportNovelCreationProject(projectId, format = 'markdown') {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/export?format=${encodeURIComponent(format)}`, {
    responseType: 'blob'
  });
}
