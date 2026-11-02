import { apiRequest } from './client.js';

export function parseLocalExecutorsResponse(payload) {
  if (Array.isArray(payload?.executors)) return payload.executors;
  if (Array.isArray(payload?.items)) return payload.items;
  return [];
}

export function hasReadyLocalExecutor(executors) {
  return Array.isArray(executors) && executors.some(item => item?.online === true);
}

export function localExecutorStatus(executors) {
  const items = Array.isArray(executors) ? executors : [];
  const onlineCount = items.filter(item => item?.online === true).length;
  return { online: onlineCount > 0, paired: items.length, onlineCount };
}

export async function listLocalExecutors(options = {}) {
  const result = await apiRequest('/api/shuihuo-production/local-executors', {
    suppressGlobalError: true,
    ...options
  });
  return parseLocalExecutorsResponse(result);
}

export function createLocalExecutorPairing() {
  return apiRequest('/api/shuihuo-production/local-executors/pairings', {
    method: 'POST',
    body: JSON.stringify({ platform: 'doubao' }),
    suppressGlobalError: true
  });
}

export async function getLocalExecutorReleaseManifest() {
  const response = await fetch('/downloads/local-executor/manifest.json', { cache: 'no-store' });
  if (!response.ok) throw new Error('读取执行器版本失败');
  const payload = await response.json();
  const versions = payload?.versions && typeof payload.versions === 'object' ? payload.versions : {};
  const downloads = payload?.downloads && typeof payload.downloads === 'object' ? payload.downloads : {};
  return {
    windows: String(versions.windows || payload?.version || '').trim(),
    macos: String(versions.macos || '').trim(),
    windowsUrl: String(downloads.windows || '').trim(),
    macosUrl: String(downloads.mac || '').trim(),
  };
}

export function normalizeLocalExecutorsResponse(payload) {
  const source = payload && typeof payload === 'object' && !Array.isArray(payload) ? payload : {};
  return { ...source, executors: parseLocalExecutorsResponse(payload) };
}
