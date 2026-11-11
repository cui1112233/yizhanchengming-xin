import { apiRequest } from './client.js';

export async function getCoverPromptConfig(request = apiRequest) {
  const response = await request('/api/plugins/cover-generation/config', { timeoutMs: 8000, suppressGlobalError: true });
  return Array.isArray(response?.prompts) ? response.prompts : [];
}

export async function buildCoverPrompt(payload, request = apiRequest) {
  const response = await request('/api/plugins/cover-generation/prompt', {
    method: 'POST',
    body: JSON.stringify(payload || {}),
    timeoutMs: 15000,
    suppressGlobalError: true
  });
  return String(response?.prompt || '');
}

// 用文本模型执行「一键封面指令」，返回最终图片提示词。
export async function designCoverPrompt(payload, request = apiRequest) {
  const response = await request('/api/plugins/cover-generation/design', {
    method: 'POST',
    body: JSON.stringify(payload || {}),
    timeoutMs: 5 * 60 * 1000,
    suppressGlobalError: true
  });
  return String(response?.prompt || '');
}
