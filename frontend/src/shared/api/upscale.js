import { apiRequest } from './client.js';

// 火山 AI MediaKit 超分：后端代理上传、提交与轮询，密钥不进入前端。
export async function getUpscaleConfig(request = apiRequest) {
  const response = await request('/api/ai-upscale/config', { timeoutMs: 8000, suppressGlobalError: true });
  return {
    configured: Boolean(response?.configured),
    resolutions: Array.isArray(response?.resolutions) ? response.resolutions : []
  };
}

export async function submitUpscaleJob({ file, resolution }, request = apiRequest) {
  if (!(file instanceof Blob)) throw new Error('请选择要超分的视频文件');
  const body = new FormData();
  body.append('file', file, file.name || 'input.mp4');
  body.append('resolution', String(resolution || '1080p'));
  const response = await request('/api/ai-upscale/jobs', { method: 'POST', body, timeoutMs: 0 });
  if (!response?.job) throw new Error('提交超分任务失败');
  return response.job;
}

export async function getUpscaleJob(jobId, request = apiRequest) {
  const response = await request(`/api/ai-upscale/jobs/${encodeURIComponent(jobId)}`, { timeoutMs: 15000, suppressGlobalError: true });
  return { job: response?.job || null, warning: String(response?.warning || '') };
}

export async function cancelUpscaleJob(jobId, request = apiRequest) {
  const response = await request(`/api/ai-upscale/jobs/${encodeURIComponent(jobId)}/cancel`, { method: 'POST', timeoutMs: 20000 });
  if (!response?.job) throw new Error('取消超分任务失败');
  return response.job;
}
