import { apiRequest } from './client';

export function getApiIntegrations() {
  return apiRequest('/api/settings/api-integrations');
}

export function saveApiIntegrations(payload) {
  return apiRequest('/api/settings/api-integrations', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

// 只取一个字段的明文；列表接口永远不下发密钥。
export function revealApiIntegration(key) {
  return apiRequest('/api/settings/api-integrations/reveal', {
    method: 'POST',
    body: JSON.stringify({ key }),
    suppressGlobalError: true
  });
}
