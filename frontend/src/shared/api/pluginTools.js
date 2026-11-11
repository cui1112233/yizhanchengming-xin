import { apiRequest } from './client.js';

const toolIdPattern = /^[a-z0-9-]+$/;

export async function getPluginToolConfig(toolId, request = apiRequest) {
  if (!toolIdPattern.test(String(toolId || ''))) throw new Error('插件工具无效');
  const response = await request(`/api/plugins/${toolId}/config`, { timeoutMs: 8000, suppressGlobalError: true });
  return Array.isArray(response?.prompts) ? response.prompts : [];
}

export async function runPluginTool({ toolId, input, modelId = '', promptId = '', raw = false }, request = apiRequest) {
  if (!toolIdPattern.test(String(toolId || ''))) throw new Error('插件工具无效');
  const response = await request(`/api/plugins/${toolId}/run`, {
    method: 'POST',
    body: JSON.stringify({ input: String(input || ''), modelId: String(modelId || ''), promptId: String(promptId || ''), raw: raw === true }),
    timeoutMs: 10 * 60 * 1000
  });
  return String(response?.result || '');
}

export async function extractDocumentText(file, request = apiRequest) {
  if (!(file instanceof Blob)) throw new Error('请选择文档文件');
  const body = new FormData();
  body.append('file', file, file.name || 'document.txt');
  const response = await request('/api/plugins/extract-text', {
    method: 'POST',
    body,
    timeoutMs: 2 * 60 * 1000,
    suppressGlobalError: true
  });
  return String(response?.text || '');
}

export async function analyzeAISegment(input, request = apiRequest, signal) {
  return request('/api/plugins/ai-segment/analyze', {
    method: 'POST',
    body: JSON.stringify(input || {}),
    timeoutMs: 10 * 60 * 1000,
    signal
  });
}

export async function exportAISegment({ video, episodes, resolution = '', fps = 0 }, request = apiRequest, signal) {
  if (!(video instanceof Blob)) throw new Error('请选择视频文件');
  if (!Array.isArray(episodes) || episodes.length === 0) throw new Error('请先完成分段分析');
  const body = new FormData();
  body.append('video', video, video.name || 'input.mp4');
  body.append('episodes', JSON.stringify(episodes));
  body.append('resolution', String(resolution || ''));
  body.append('fps', String(fps || 0));
  return request('/api/plugins/ai-segment/export', {
    method: 'POST',
    body,
    responseType: 'blob',
    timeoutMs: 0,
    signal
  });
}
