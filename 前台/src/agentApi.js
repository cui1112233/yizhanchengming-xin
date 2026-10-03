async function readFailure(response, fallback) {
  try {
    const text = await response.text();
    if (text?.trim()) return text.trim();
  } catch {
    // Ignore secondary body parsing failures.
  }
  return fallback;
}

async function requestJSON(url, options = {}, fetchImpl = globalThis.fetch) {
  if (typeof fetchImpl !== 'function') throw new Error('fetch is required');
  const response = await fetchImpl(url, {
    ...options,
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(options.headers || {}),
    },
  });
  if (!response?.ok) {
    throw new Error(await readFailure(response, `${url} 请求失败`));
  }
  if (response.status === 204) return null;
  return response.json();
}

export async function listAgentThreads({ fetchImpl = globalThis.fetch, limit = 50 } = {}) {
  const result = await requestJSON(`/api/agent/threads?limit=${encodeURIComponent(limit)}`, {}, fetchImpl);
  return Array.isArray(result?.threads) ? result.threads : [];
}

export async function createAgentThread({ fetchImpl = globalThis.fetch, title = '' } = {}) {
  const result = await requestJSON('/api/agent/threads', {
    method: 'POST',
    body: JSON.stringify({ title }),
  }, fetchImpl);
  return result?.thread || result;
}

export async function getAgentThread(threadId, { fetchImpl = globalThis.fetch } = {}) {
  const id = String(threadId || '').trim();
  if (!id) throw new Error('thread id is required');
  return requestJSON(`/api/agent/threads/${encodeURIComponent(id)}`, {}, fetchImpl);
}

export async function sendAgentMessage(threadId, {
  fetchImpl = globalThis.fetch,
  content = '',
  mediaAssetIds = [],
} = {}) {
  const id = String(threadId || '').trim();
  if (!id) throw new Error('thread id is required');
  return requestJSON(`/api/agent/threads/${encodeURIComponent(id)}/messages`, {
    method: 'POST',
    body: JSON.stringify({
      content: String(content || ''),
      media_asset_ids: Array.isArray(mediaAssetIds) ? mediaAssetIds : [],
    }),
  }, fetchImpl);
}

export async function listAgentTasks({ fetchImpl = globalThis.fetch, threadId = '' } = {}) {
  const query = threadId ? `?thread_id=${encodeURIComponent(threadId)}` : '';
  const result = await requestJSON(`/api/agent/tasks${query}`, {}, fetchImpl);
  return Array.isArray(result?.tasks) ? result.tasks : [];
}

export async function updateAgentTask(taskId, input, { fetchImpl = globalThis.fetch } = {}) {
  const id = String(taskId || '').trim();
  if (!id) throw new Error('task id is required');
  const result = await requestJSON(`/api/agent/tasks/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(input || {}),
  }, fetchImpl);
  return result?.task || result;
}

export async function updateAgentToolCall(toolId, input, { fetchImpl = globalThis.fetch } = {}) {
  const id = String(toolId || '').trim();
  if (!id) throw new Error('tool id is required');
  const result = await requestJSON(`/api/agent/tool-calls/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(input || {}),
  }, fetchImpl);
  return result?.tool_call || result;
}
