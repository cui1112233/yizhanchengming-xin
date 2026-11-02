import { apiRequest } from './client';

export function generateReferenceAssetImage(payload) {
  return apiRequest('/api/reference-assets/generate', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function getReferenceAssetImageTask(taskId) {
  return apiRequest(`/api/reference-assets/generate/${encodeURIComponent(taskId)}`, { suppressGlobalError: true });
}

export function uploadReferenceAssetImage(payload) {
  return apiRequest('/api/reference-assets/upload', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function loadReferenceAssetImage(url) {
  const value = String(url || '').trim();
  const target = /^https:\/\/[^/]*tos[^/]*\.volces\.com\//i.test(value)
    ? '/api/reference-assets/legacy?url=' + encodeURIComponent(value)
    : value.replace(/\/api\/[^/]+\/reference-assets\//, '/api/reference-assets/');
  return apiRequest(target, { responseType: 'blob' });
}
