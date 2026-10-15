import { apiRequest } from './client';

export function getNovelPanelProject(projectId) {
  return apiRequest(`/api/novel-panel/projects/${encodeURIComponent(String(projectId || ''))}`);
}

export function listNovelPanelProjects() {
  return apiRequest('/api/novel-panel/projects');
}

export function saveNovelPanelProject(payload) {
  return apiRequest('/api/novel-panel/projects', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function analyzeNovelPanel(payload) {
  return apiRequest('/api/novel-panel/analyze', { method: 'POST', body: JSON.stringify(payload) });
}

export function outlineNovelPanelScenes(payload) {
  return apiRequest('/api/novel-panel/outline-scenes', { method: 'POST', body: JSON.stringify(payload) });
}

export function regenerateNovelPanelSceneOutline(payload) {
  return apiRequest('/api/novel-panel/regenerate-scene-outline', { method: 'POST', body: JSON.stringify(payload) });
}

export function generateNovelPanelScenePrompts(payload) {
  return apiRequest('/api/novel-panel/generate-scene-prompts', { method: 'POST', body: JSON.stringify(payload) });
}

export function listNovelPanelHistory() {
  return apiRequest('/api/novel-panel/history');
}

export function generateReferenceAssetImage(payload) {
  return apiRequest('/api/novel-panel/reference-assets/generate', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function uploadReferenceAssetImage(payload) {
  return apiRequest('/api/novel-panel/reference-assets/upload', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function loadReferenceAssetImage(url) {
  const value = String(url || '').trim();
  const target = /^https:\/\/[^/]*tos[^/]*\.volces\.com\//i.test(value)
    ? '/api/novel-panel/reference-assets/legacy?url=' + encodeURIComponent(value)
    : value;
  return apiRequest(target, { responseType: 'blob' });
}
