function providerPath(provider) {
  const value = String(provider || '').trim().toLowerCase();
  if (!/^[a-z0-9_-]{1,64}$/.test(value)) throw new Error('无效的视频服务商');
  return `/api/provider-configs/video/${encodeURIComponent(value)}`;
}

async function readJSON(response) {
  const text = await response.text();
  if (!text) return null;
  try { return JSON.parse(text); } catch { return null; }
}

export async function getVideoProviderConfig(provider) {
  const response = await fetch(providerPath(provider), { headers: { Accept: 'application/json' } });
  if (response.status === 404) return null;
  if (!response.ok) {
    const body = await readJSON(response);
    throw new Error(body?.message || body?.error || `读取视频模型配置失败（HTTP ${response.status}）`);
  }
  return readJSON(response);
}

export async function putVideoProviderConfig(provider, input = {}) {
  const payload = {
    model: String(input.model || '').trim(),
    api_key: String(input.apiKey || ''),
    enabled: input.enabled !== false,
  };
  if (input.createUrl) payload.create_url = String(input.createUrl).trim();
  if (input.tasksUrl) payload.tasks_url = String(input.tasksUrl).trim();
  if (input.resultUrl) payload.result_url = String(input.resultUrl).trim();
  if (input.settings && typeof input.settings === 'object') payload.settings = input.settings;

  const response = await fetch(providerPath(provider), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text?.trim() || `保存视频模型配置失败（HTTP ${response.status}）`);
  }
  return readJSON(response);
}
