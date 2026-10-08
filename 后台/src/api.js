export class APIError extends Error {
  constructor(message, { status = 0, code = 'API_ERROR', requestId = '' } = {}) {
    super(message)
    Object.assign(this, { name: 'APIError', status, code: code || 'API_ERROR', requestId })
  }
}

export async function requestJSON(path, options = {}) {
  const headers = new Headers(options.headers || {})
  if (options.body !== undefined) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...options, headers, credentials: 'include' })
  const payload = await response.json().catch(() => null)
  if (response.ok) return payload
  throw new APIError(payload?.message || `请求失败（HTTP ${response.status}）`, {
    status: response.status,
    code: payload?.code,
    requestId: response.headers?.get?.('X-Request-ID') || payload?.request_id || '',
  })
}
