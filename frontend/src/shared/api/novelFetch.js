import { apiRequest } from './client.js';

export function fetchNovelContent({ platform, bookIds, maxTxt, textModel, selectedVersions, rewriteMethods, sensitiveAiProcessEnabled }) {
  return apiRequest('/api/novel-fetch', {
    method: 'POST',
    body: JSON.stringify({ platform, bookIds, maxTxt, textModel, selectedVersions, rewriteMethods, sensitiveAiProcessEnabled })
  });
}
export function listNovelFetchProcessPresets() {
  return apiRequest('/api/presets?module=novel-fetch');
}

export function processNovelContent({ mode, items, platform, platformName, saveToFolder }) {
  return apiRequest('/api/novel-fetch/process', {
    method: 'POST',
    body: JSON.stringify({ mode, items, platform, platformName, saveToFolder })
  });
}

export function saveNovelContent({ bookId, text, meta }) {
  return apiRequest('/api/novel-fetch/save', {
    method: 'POST',
    body: JSON.stringify({ bookId, text, meta })
  });
}

export function listSavedNovels() {
  return apiRequest('/api/novel-fetch/saved');
}

export function uploadLogin({ username, password }) {
  return apiRequest('/api/novel-fetch-workshop/upload/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
    suppressGlobalError: true
  });
}

export function getUploadSession() {
  return apiRequest('/api/novel-fetch-workshop/upload/session');
}

export function uploadBatch(payload) {
  return apiRequest('/api/novel-fetch-workshop/upload/batch', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

const WEB_SUBMIT_API = '/api/novel-fetch-workshop/web-submit';
export function getWebSubmitConfig() { return apiRequest(`${WEB_SUBMIT_API}/config`); }
export function saveWebSubmitConfig(payload) {
  const settings = payload?.settings && typeof payload.settings === 'object' ? payload.settings : payload;
  return apiRequest(`${WEB_SUBMIT_API}/config`, { method: 'POST', body: JSON.stringify({ settings }) });
}
export function checkWebSubmitEnvironment() { return apiRequest(`${WEB_SUBMIT_API}/environment`); }
export function getWebSubmitHistory() { return apiRequest(`${WEB_SUBMIT_API}/history`); }
export function webSubmitOrganizationsPath() { return '/api/batch-factory/v12/publish-121/organizations'; }
export function getWebSubmitOrganizations() { return apiRequest(webSubmitOrganizationsPath()); }
export function syncWebSubmitConfigs(payload = {}) { return apiRequest(`${WEB_SUBMIT_API}/sync-configs`, { method: 'POST', body: JSON.stringify(payload) }); }
export function syncWebSubmitStyles(payload = {}) { return apiRequest(`${WEB_SUBMIT_API}/sync-styles`, { method: 'POST', body: JSON.stringify(payload) }); }
export function testWebSubmitVisible(payload = {}) { return apiRequest(`${WEB_SUBMIT_API}/test-visible`, { method: 'POST', body: JSON.stringify(payload) }); }
export function previewWebSubmit(payload) { return apiRequest(`${WEB_SUBMIT_API}/preview`, { method: 'POST', body: JSON.stringify(payload) }); }
export function startWebSubmit(payload) { return apiRequest(`${WEB_SUBMIT_API}/submit`, { method: 'POST', body: JSON.stringify(payload) }); }
