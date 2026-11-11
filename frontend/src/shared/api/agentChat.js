import { apiRequest, getToken } from './client';

// Workspaces / sessions / messages for the conversational Agent workspace.
// Read endpoints go through apiRequest so timeout and auth-failure handling
// stay identical to the rest of the app; the streaming send uses raw fetch
// because apiRequest buffers the whole body.

export function listAgentChatWorkspaces() {
  return apiRequest('/api/agent/workspaces');
}

export function createAgentChatWorkspace(name) {
  return apiRequest('/api/agent/workspaces', { method: 'POST', body: JSON.stringify({ name }) });
}

export function updateAgentChatWorkspace(id, patch) {
  return apiRequest(`/api/agent/workspaces/${encodeURIComponent(id)}`, { method: 'PATCH', body: JSON.stringify(patch) });
}

export function deleteAgentChatWorkspace(id) {
  return apiRequest(`/api/agent/workspaces/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export function listAgentChatTree(workspaceId, path = '') {
  const query = path ? `?path=${encodeURIComponent(path)}` : '';
  return apiRequest(`/api/agent/workspaces/${encodeURIComponent(workspaceId)}/tree${query}`);
}

export function readAgentChatFile(workspaceId, path) {
  return apiRequest(`/api/agent/workspaces/${encodeURIComponent(workspaceId)}/file?path=${encodeURIComponent(path)}`);
}

export function listAgentChatSessions(workspaceId = '') {
  const query = workspaceId ? `?workspaceId=${encodeURIComponent(workspaceId)}` : '';
  return apiRequest(`/api/agent/sessions${query}`);
}

export function createAgentChatSession({ title, workspaceId, model, permission }) {
  return apiRequest('/api/agent/sessions', {
    method: 'POST',
    body: JSON.stringify({ title, workspaceId, model, permission })
  });
}

export function updateAgentChatSession(id, patch) {
  return apiRequest(`/api/agent/sessions/${id}`, { method: 'PATCH', body: JSON.stringify(patch) });
}

export function deleteAgentChatSession(id) {
  return apiRequest(`/api/agent/sessions/${id}`, { method: 'DELETE' });
}

export function getAgentChatMessages(id) {
  return apiRequest(`/api/agent/sessions/${id}/messages`);
}

export function clearAgentChatMessages(id) {
  return apiRequest(`/api/agent/sessions/${id}/messages`, { method: 'DELETE' });
}

export function listAgentChatPermissions() {
  return apiRequest('/api/agent/permissions');
}

export function listAgentChatModels() {
  return apiRequest('/api/agent/models');
}

export function stopAgentChatRun(id) {
  return apiRequest(`/api/agent/sessions/${id}/stop`, { method: 'POST' });
}

function parseSSEChunk(chunk) {
  const lines = chunk.split('\n');
  let event = 'message';
  const dataLines = [];
  lines.forEach(line => {
    if (line.startsWith('event:')) event = line.slice(6).trim();
    else if (line.startsWith('data:')) dataLines.push(line.slice(5).trim());
  });
  if (!dataLines.length) return null;
  try {
    return { event, data: JSON.parse(dataLines.join('\n')) };
  } catch (_) {
    return { event, data: {} };
  }
}

// streamAgentChatMessage reads the SSE response and forwards each named event
// to handlers[event](data). Returns when the stream closes.
export async function streamAgentChatMessage(sessionId, body, handlers = {}, signal) {
  const response = await fetch(`/api/agent/sessions/${sessionId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getToken()}`
    },
    body: JSON.stringify(body),
    signal
  });
  if (!response.ok) {
    let message = '请求失败';
    try {
      const payload = await response.json();
      if (payload && typeof payload.error === 'string') message = payload.error;
    } catch (_) {
      // keep the default Chinese message
    }
    throw new Error(message);
  }
  if (!response.body) throw new Error('请求失败');
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    let boundary = buffer.indexOf('\n\n');
    while (boundary >= 0) {
      const raw = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);
      const parsed = parseSSEChunk(raw);
      if (parsed && typeof handlers[parsed.event] === 'function') handlers[parsed.event](parsed.data);
      boundary = buffer.indexOf('\n\n');
    }
  }
}
