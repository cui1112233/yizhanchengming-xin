const REVISION_MARKER = '【修改稿】';

export function parseScriptRevision(content) {
  const text = String(content || '').trim();
  const markerIndex = text.indexOf(REVISION_MARKER);
  if (markerIndex < 0) return null;
  const summary = text.slice(0, markerIndex).trim();
  const candidateOutput = text.slice(markerIndex + REVISION_MARKER.length).trim();
  return candidateOutput ? { summary, candidateOutput } : null;
}

export function classifyPetRequestError(error) {
  const status = Number(error?.status || error?.response?.status || 0);
  const source = String(error?.source || '');
  const detail = String(error?.message || error || '');
  // 登录态失效按状态码判定：此时响应体里也可能出现 key/鉴权字样，
  // 不能因此把「会话过期」误报成模型配置问题。
  if (status === 401 || status === 403) {
    return { message: '登录已过期，请重新登录后再试。', action: 'login' };
  }
  // 422 是后端业务校验，直接把校验文案交还给用户，不做二次翻译。
  if (status === 422) {
    return { message: detail || '提交内容未通过校验，请按提示修改后重试。', action: 'retry' };
  }
  if (/(?:base url|model|api key) is required|api.?key|unauthori[sz]ed|forbidden|密钥|鉴权/i.test(detail)) {
    return { message: '模型配置无法使用，请检查接口、模型名或密钥。', action: 'settings' };
  }
  if (status === 504 || /timeout|超时/i.test(detail)) {
    return { message: '连接模型超时，当前剧本已保留。', action: 'retry' };
  }
  if (status === 502 || /non-json|upstream|返回.*异常/i.test(detail)) {
    return { message: '模型返回内容异常，当前剧本未改动。', action: 'retry' };
  }
  if (source === '/api/agent/chat' || /network|failed to fetch|网络/i.test(detail)) {
    return { message: '网络连接不稳定，当前剧本已保留。', action: 'retry' };
  }
  return { message: '这次没有完成处理，当前剧本未改动。', action: 'retry' };
}
