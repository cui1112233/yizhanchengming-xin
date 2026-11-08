import { apiRequest } from './client';

// The endpoint answers 404 when the running binary has no notes for its
// release. Allow that status so the shared client does not raise the global
// "请求失败" banner; the modal treats a missing releaseId as "show nothing".
export function fetchLatestProductUpdate(options = {}) {
  return apiRequest('/api/product-updates/latest', {
    allowStatuses: [404],
    silent: true,
    suppressGlobalError: true,
    ...options
  });
}
