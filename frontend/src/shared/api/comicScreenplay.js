import { apiRequest } from './client.js';

function projectPath(projectId, suffix = '') {
  const id = String(projectId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('漫剧项目无效');
  return `/api/comic-screenplays/projects/${id}${suffix}`;
}

function shotPath(projectId, shotId, suffix = '') {
  const id = String(shotId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('漫剧镜头无效');
  return projectPath(projectId, `/shots/${id}${suffix}`);
}

function assetPath(projectId, assetId) {
  const id = String(assetId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('漫剧资产无效');
  return projectPath(projectId, `/assets/${id}`);
}

function documentPath(projectId, documentId, suffix = '') {
  const id = String(documentId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('剧本版本无效');
  return projectPath(projectId, `/documents/${id}${suffix}`);
}

function storyboardShotPath(projectId, engine, episodeNo, shotId = '', suffix = '') {
  const engineId = String(engine || '').trim();
  if (!/^[A-Za-z0-9_][A-Za-z0-9_-]{0,63}$/.test(engineId)) throw new Error('分镜引擎无效');
  const episode = Number(episodeNo);
  if (!Number.isInteger(episode) || episode <= 0) throw new Error('分镜集数无效');
  let path = projectPath(projectId, `/storyboards/${engineId}/${episode}/shots`);
  if (shotId !== '') {
    const id = String(shotId || '').trim();
    if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('分镜镜头无效');
    path += `/${id}`;
  }
  return path + suffix;
}

function validateStoryboardShotInput(input) {
  if (!(Number(input?.durationSec) > 0)) throw new Error('镜头时长必须大于 0 秒');
  if (![input?.characters, input?.scene, input?.props, input?.videoDesc].some(value => String(value || '').trim())) throw new Error('分镜内容不能为空');
}

function validateAssetInput(input) {
  if (!['character', 'scene', 'prop'].includes(input?.type)) throw new Error('资产类型无效');
  if (!String(input?.name || '').trim()) throw new Error('资产名称不能为空');
}

export async function listComicScreenplays(request = apiRequest) {
  const response = await request('/api/comic-screenplays/projects');
  return Array.isArray(response?.projects) ? response.projects : [];
}

export async function listComicScreenplayPrompts(request = apiRequest) {
  const response = await request('/api/comic-screenplays/prompts');
  return Array.isArray(response?.prompts) ? response.prompts : [];
}

export async function listComicScreenplayStyles(request = apiRequest) {
  const response = await request('/api/comic-screenplays/styles');
  return Array.isArray(response?.styles) ? response.styles : [];
}

export async function parseComicScreenplayDraftSource(file, request = apiRequest) {
  if (!(file instanceof Blob) || !/\.(txt|docx)$/i.test(file.name || '')) throw new Error('请选择 TXT 或 DOCX 文件');
  const body = new FormData();
  body.append('file', file, file.name);
  const response = await request('/api/comic-screenplays/source/parse', { method: 'POST', body });
  return String(response?.content || '');
}

export async function createComicScreenplay(input, request = apiRequest) {
  const title = String(input?.title || '').trim();
  if (!title) throw new Error('作品名称不能为空');
  const response = await request('/api/comic-screenplays/projects', { method: 'POST', body: JSON.stringify({ ...input, title }) });
  return response?.project;
}

export async function deleteComicScreenplay(projectId, request = apiRequest) {
  return request(projectPath(projectId), { method: 'DELETE' });
}

export async function updateComicScreenplay(projectId, input, request = apiRequest) {
  const response = await request(projectPath(projectId), { method: 'PATCH', body: JSON.stringify(input) });
  return response?.project;
}

export async function saveComicScreenplaySource(projectId, content, request = apiRequest) {
  const response = await request(projectPath(projectId, '/source'), { method: 'PUT', body: JSON.stringify({ content: String(content ?? '') }) });
  return response?.project;
}

export async function saveComicScreenplayDocument(projectId, content, request = apiRequest) {
  const normalized = String(content ?? '').trim();
  if (!normalized) throw new Error('剧本内容不能为空');
  const response = await request(projectPath(projectId, '/document'), { method: 'PUT', body: JSON.stringify({ content: normalized }) });
  return response?.document;
}

export async function restoreComicScreenplayDocument(projectId, documentId, request = apiRequest) {
  const response = await request(documentPath(projectId, documentId, '/restore'), { method: 'POST' });
  return response?.document;
}

export async function loadComicScreenplayWorkspace(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/workspace'));
}

export async function createComicScreenplayShot(projectId, input, request = apiRequest) {
  if (!String(input?.heading || '').trim() && !String(input?.scriptText || '').trim()) throw new Error('场次标题和内容不能同时为空');
  const response = await request(projectPath(projectId, '/shots'), { method: 'POST', body: JSON.stringify(input) });
  return response?.shot;
}

export async function updateComicScreenplayShot(projectId, shotId, input, request = apiRequest) {
  if (!String(input?.heading || '').trim() && !String(input?.scriptText || '').trim()) throw new Error('场次标题和内容不能同时为空');
  const response = await request(shotPath(projectId, shotId), { method: 'PATCH', body: JSON.stringify(input) });
  return response?.shot;
}

export async function moveComicScreenplayShot(projectId, shotId, direction, request = apiRequest) {
  if (direction !== 'up' && direction !== 'down') throw new Error('移动方向必须是 up 或 down');
  const response = await request(shotPath(projectId, shotId, '/move'), { method: 'POST', body: JSON.stringify({ direction }) });
  return Array.isArray(response?.shots) ? response.shots : [];
}

export async function deleteComicScreenplayShot(projectId, shotId, request = apiRequest) {
  return request(shotPath(projectId, shotId), { method: 'DELETE' });
}

export async function parseComicScreenplaySource(projectId, file, request = apiRequest) {
  if (!(file instanceof Blob)) throw new Error('请选择 TXT 或 DOCX 文件');
  const body = new FormData();
  body.append('file', file, file.name || 'source.txt');
  const response = await request(projectPath(projectId, '/source/parse'), { method: 'POST', body, timeoutMs: 0 });
  return String(response?.content || '');
}

export async function generateComicScreenplay(projectId, input, request = apiRequest) {
  const content = String(input?.content || input?.novel || '').trim();
  if (!content) throw new Error('请先上传或输入原文');
  return request(projectPath(projectId, '/generate'), { method: 'POST', timeoutMs: 0, signal: input?.signal, body: JSON.stringify({ ...input, content, confirmed: true }) });
}

export async function extractComicAssets(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/assets/extract'), { method: 'POST', timeoutMs: 0, signal: input?.signal, body: JSON.stringify({ ...input, confirmed: true }) });
}

export async function createComicScreenplayAsset(projectId, input, request = apiRequest) {
  validateAssetInput(input);
  const response = await request(projectPath(projectId, '/assets'), { method: 'POST', body: JSON.stringify({ ...input, name: String(input.name).trim() }) });
  return response?.asset;
}

export async function updateComicScreenplayAsset(projectId, assetId, input, request = apiRequest) {
  validateAssetInput(input);
  const response = await request(assetPath(projectId, assetId), { method: 'PATCH', body: JSON.stringify({ ...input, name: String(input.name).trim() }) });
  return response?.asset;
}

export async function deleteComicScreenplayAsset(projectId, assetId, request = apiRequest) {
  return request(assetPath(projectId, assetId), { method: 'DELETE' });
}

export async function generateComicStoryboard(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/storyboards/generate'), { method: 'POST', timeoutMs: 0, signal: input?.signal, body: JSON.stringify({ ...input, confirmed: true }) });
}

export async function createComicStoryboardShot(projectId, engine, episodeNo, input, request = apiRequest) {
  validateStoryboardShotInput(input);
  const response = await request(storyboardShotPath(projectId, engine, episodeNo), { method: 'POST', body: JSON.stringify(input) });
  return response?.storyboard;
}

export async function updateComicStoryboardShot(projectId, engine, episodeNo, shotId, input, request = apiRequest) {
  validateStoryboardShotInput(input);
  const response = await request(storyboardShotPath(projectId, engine, episodeNo, shotId), { method: 'PATCH', body: JSON.stringify(input) });
  return response?.storyboard;
}

export async function splitComicStoryboardShot(projectId, engine, episodeNo, shotId, request = apiRequest) {
  const response = await request(storyboardShotPath(projectId, engine, episodeNo, shotId, '/split'), { method: 'POST', body: '{}' });
  return response?.storyboard;
}

export async function mergeComicStoryboardShot(projectId, engine, episodeNo, shotId, request = apiRequest) {
  const response = await request(storyboardShotPath(projectId, engine, episodeNo, shotId, '/merge-next'), { method: 'POST', body: '{}' });
  return response?.storyboard;
}

export async function deleteComicStoryboardShot(projectId, engine, episodeNo, shotId, request = apiRequest) {
  const response = await request(storyboardShotPath(projectId, engine, episodeNo, shotId), { method: 'DELETE' });
  return response?.storyboard;
}

export async function startComicScreenplayBatch(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/batch'), { method: 'POST', body: '{}' });
}

export async function getComicScreenplayBatch(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/batch'));
}

export async function runComicScreenplayBatchNext(projectId, signal, request = apiRequest) {
  return request(projectPath(projectId, '/batch/run-next'), { method: 'POST', timeoutMs: 0, signal, body: '{}' });
}

export async function cancelComicScreenplayBatch(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/batch/cancel'), { method: 'POST', body: '{}' });
}

export async function exportComicScreenplay(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/export'));
}


function creativeProjectPath(projectId) {
  const id = String(projectId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('漫剧项目无效');
  return `/api/screenplays/${id}/creative-config`;
}

export async function loadComicCreativeLibrary(request = apiRequest) {
  const response = await request('/api/screenplays/creative-library');
  return { styles: Array.isArray(response?.styles) ? response.styles : [], prompts: Array.isArray(response?.prompts) ? response.prompts : [] };
}

export async function loadComicCreativeConfig(projectId, request = apiRequest) {
  return request(creativeProjectPath(projectId));
}

export async function saveComicCreativeConfig(projectId, input, request = apiRequest) {
  return request(creativeProjectPath(projectId), { method: 'PUT', body: JSON.stringify(input) });
}

export async function getComicStylePreview(src, request = apiRequest) {
  const path = String(src || '');
  if (!/^\/api\/reference-assets\/file\/[A-Za-z0-9_-]+\/[A-Za-z0-9_-]+\/[A-Za-z0-9_-]+$/.test(path)
    && !/^\/api\/agent\/tasks\/[A-Za-z0-9_-]+\/assets\/[A-Za-z0-9_-]+\/content$/.test(path)
    && !/^\/api\/ai-cut\/projects\/[A-Za-z0-9_-]+\/assets\/[A-Za-z0-9_-]+\/content$/.test(path)) throw new Error('画风预览地址无效');
  return request(path, { responseType: 'blob', suppressGlobalError: true });
}

export async function createComicCreativeStyle(input, request = apiRequest) {
  const name = String(input?.name || '').trim(), prompt = String(input?.prompt || '').trim();
  if (!name) throw new Error('请输入画风名称');
  if (!prompt) throw new Error('请输入画风文字');
  const response = await request('/api/screenplays/styles', { method: 'POST', body: JSON.stringify({ name, prompt, images: [] }) });
  if (!response?.style?.id) throw new Error('画风保存结果缺少编号，请刷新画风库核对');
  return response.style;
}
