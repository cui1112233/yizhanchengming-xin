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

// 执行器状态的统一只读入口：一次返回本地与巨量两类执行器的
// 在线/离线/未安装状态和设备列表，页面不再分别猜测两套口径。
export function getExecutorStatus(options = {}) {
  return apiRequest('/api/shuihuo-production/executors/status', {
    suppressGlobalError: true,
    ...options
  });
}

// 执行器状态只读合并：优先采用后端 state；没有设备时，后端返回的
// 未过期待配对记录（pairing）会恢复 "needs_pairing"，刷新后不再丢失。
// 待配对信息由状态接口持久化返回，前端不再叠加本地配对码状态。
export function mergeExecutorState(group) {
  const devices = Array.isArray(group?.devices) ? group.devices : [];
  const pairing = group?.pairing && group.pairing.expired !== true ? group.pairing : null;
  if (devices.length === 0 && pairing) return 'needs_pairing';
  const state = String(group?.state || '').trim();
  if (state) return state;
  if (devices.length === 0) return 'not_installed';
  return devices.some(item => item?.online) ? 'online' : 'offline';
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
