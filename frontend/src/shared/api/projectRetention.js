import { apiRequest } from './client.js';

export const RETENTION_MODULES = {
  batch_factory: '批量工厂',
  comic_screenplay: '漫剧创作',
  cm_canvas: 'CM 画布',
  novel_creation: '小说创作'
};

export function retentionModuleLabel(module) {
  return RETENTION_MODULES[module] || module || '作品';
}

export async function getProjectRetention(request = apiRequest) {
  const response = await request('/api/project-retention');
  return Array.isArray(response?.items) ? response.items : [];
}

export async function getArchivedProjects(request = apiRequest) {
  const response = await request('/api/project-retention/archived');
  return Array.isArray(response?.items) ? response.items : [];
}

export function restoreArchivedProject(module, projectId, request = apiRequest) {
  return request(`/api/project-retention/archived/${encodeURIComponent(module)}/${encodeURIComponent(projectId)}/restore`, { method: 'POST', body: '{}' });
}

export function purgeArchivedProject(module, projectId, request = apiRequest) {
  return request(`/api/project-retention/archived/${encodeURIComponent(module)}/${encodeURIComponent(projectId)}`, { method: 'DELETE' });
}

// 列表页只需要一个按 module+projectId 索引的映射，避免每张卡片线性查找。
export function indexRetention(items) {
  const map = {};
  (Array.isArray(items) ? items : []).forEach(item => {
    if (item?.module && item?.projectId) map[`${item.module}:${item.projectId}`] = item;
  });
  return map;
}
