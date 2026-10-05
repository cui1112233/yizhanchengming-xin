const SAFE_REQUEST_ID = /^[A-Za-z0-9._:-]+$/
const MAX_REQUEST_ID_LENGTH = 128

export function safeRequestIdFromHeaders(headers) {
  if (!headers || typeof headers !== 'object') return ''
  const raw = headers['x-request-id'] ?? headers['X-Request-ID']
  if (Array.isArray(raw)) return ''
  if (typeof raw !== 'string') return ''
  const value = raw.trim()
  if (!value || value.length > MAX_REQUEST_ID_LENGTH) return ''
  if (!SAFE_REQUEST_ID.test(value)) return ''
  return value
}
