import { apiRequest } from './client.js';

export function getImageCapacity(modelId, { signal, aspectRatio } = {}) {
  const query = new URLSearchParams({ modelId });
  if (aspectRatio) query.set('aspectRatio', aspectRatio);
  return apiRequest(`/api/config/image-capacity?${query}`, { signal, suppressGlobalError: true });
}

export function saveImageCapacity(modelId, policy, { signal } = {}) {
  return apiRequest('/api/config/image-capacity', {
    method: 'PUT',
    body: JSON.stringify({ ...policy, modelId }),
    signal,
    suppressGlobalError: true
  });
}
