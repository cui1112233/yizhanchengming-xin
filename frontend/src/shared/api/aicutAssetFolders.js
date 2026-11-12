import { apiRequest } from './client.js';

function path(projectId, suffix = '') {
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(projectId || '')) throw new Error('工程无效');
  return `/api/ai-cut/projects/${encodeURIComponent(projectId)}${suffix}`;
}
export async function listAssetFolders(projectId, request = apiRequest) { return (await request(path(projectId, '/asset-folders'))).folders || []; }
export async function createAssetFolder(projectId, name, request = apiRequest) { return (await request(path(projectId, '/asset-folders'), { method: 'POST', body: JSON.stringify({ name }) })).folder; }
export async function renameAssetFolder(projectId, id, name, request = apiRequest) { return request(path(projectId, `/asset-folders/${encodeURIComponent(id)}`), { method: 'PATCH', body: JSON.stringify({ name }) }); }
export async function deleteAssetFolder(projectId, id, request = apiRequest) { return request(path(projectId, `/asset-folders/${encodeURIComponent(id)}`), { method: 'DELETE' }); }
export async function moveAssetToFolder(projectId, assetId, folderId, request = apiRequest) { return request(path(projectId, `/assets/${encodeURIComponent(assetId)}/folder`), { method: 'PATCH', body: JSON.stringify({ folderId }) }); }
