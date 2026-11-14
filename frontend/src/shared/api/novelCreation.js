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

export function listNovelCreationMaterials(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/materials`);
}

export function createNovelCreationMaterial(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/materials`, { method: 'POST', body: JSON.stringify(input) });
}

export function deleteNovelCreationMaterial(projectId, materialId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/materials/${encodeURIComponent(materialId)}`, { method: 'DELETE' });
}

export function listNovelCreationNotes(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/notes`);
}

export function createNovelCreationNote(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/notes`, { method: 'POST', body: JSON.stringify(input) });
}

export function updateNovelCreationNote(projectId, noteId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/notes/${encodeURIComponent(noteId)}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteNovelCreationNote(projectId, noteId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/notes/${encodeURIComponent(noteId)}`, { method: 'DELETE' });
}

export function getNovelCreationDashboard(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/dashboard`);
}

export function listNovelCreationSchedules(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/schedules`);
}

export function createNovelCreationSchedule(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/schedules`, { method: 'POST', body: JSON.stringify(input) });
}

export function updateNovelCreationSchedule(projectId, scheduleId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/schedules/${encodeURIComponent(scheduleId)}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteNovelCreationSchedule(projectId, scheduleId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/schedules/${encodeURIComponent(scheduleId)}`, { method: 'DELETE' });
}

export function runNovelCreationSchedule(projectId, scheduleId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/schedules/${encodeURIComponent(scheduleId)}/run`, { method: 'POST', body: '{}' });
}

export function listNovelCreationBatches(projectId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/batches`);
}

export function createNovelCreationBatch(projectId, input) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/batches`, { method: 'POST', body: JSON.stringify(input) });
}

export function pauseNovelCreationBatch(projectId, batchId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/batches/${encodeURIComponent(batchId)}/pause`, { method: 'POST', body: '{}' });
}

export function resumeNovelCreationBatch(projectId, batchId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/batches/${encodeURIComponent(batchId)}/resume`, { method: 'POST', body: '{}' });
}

export function stopNovelCreationBatch(projectId, batchId) {
  return apiRequest(`/api/novel-creation/projects/${encodeURIComponent(projectId)}/batches/${encodeURIComponent(batchId)}/stop`, { method: 'POST', body: '{}' });
}
