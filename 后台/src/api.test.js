import { afterEach, expect, test, vi } from 'vitest'
import { APIError, requestJSON } from './api.js'

function response(status, payload, headers = {}) {
  return { ok: status >= 200 && status < 300, status, headers: new Headers(headers), json: async () => payload }
}

afterEach(() => vi.restoreAllMocks())

test('includes cookies and preserves caller headers on successful reads', async () => {
  const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(response(200, { prompts: [] }))
  await expect(requestJSON('/api/v1/admin/prompts', { credentials: 'omit', headers: { Accept: 'application/json' } })).resolves.toEqual({ prompts: [] })
  const [path, options] = fetchMock.mock.calls[0]
  expect(path).toBe('/api/v1/admin/prompts')
  expect(options.credentials).toBe('include')
  expect(options.headers.get('Accept')).toBe('application/json')
  expect(options.headers.has('Content-Type')).toBe(false)
})

test.each(['POST', 'PUT'])('uses JSON and cookies for %s without client tokens or storage', async (method) => {
  const storageSpies = ['getItem', 'setItem', 'removeItem', 'clear', 'key'].map((methodName) => vi.spyOn(Storage.prototype, methodName))
  const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(response(200, { saved: true }))
  await requestJSON('/api/v1/admin/prompts/script.default/drafts', { method, body: '{"content":"draft"}', headers: { 'X-Custom': 'keep' } })
  const options = fetchMock.mock.calls[0][1]
  expect(options).toMatchObject({ method, body: '{"content":"draft"}', credentials: 'include' })
  expect(options.headers.get('Content-Type')).toBe('application/json')
  expect(options.headers.get('X-Custom')).toBe('keep')
  expect(options.headers.has('Authorization')).toBe(false)
  expect(options.headers.has('X-CSRF-Token')).toBe(false)
  storageSpies.forEach((spy) => expect(spy).not.toHaveBeenCalled())
})

test.each([401, 403])('retains structured HTTP %s errors without refresh or payload exposure', async (status) => {
  const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(response(status, {
    message: 'permission denied', code: 'ADMIN_DENIED', request_id: 'body-id', payload: { secret: 'private' },
  }, { 'X-Request-ID': 'header-id' }))
  const error = await requestJSON('/api/v1/admin/capabilities').catch((err) => err)
  expect(error).toBeInstanceOf(APIError)
  expect(error).toMatchObject({ name: 'APIError', message: 'permission denied', status, code: 'ADMIN_DENIED', requestId: 'header-id' })
  expect(error).not.toHaveProperty('payload')
  expect(Object.keys(error).sort()).toEqual(['code', 'name', 'requestId', 'status'])
  expect(fetchMock).toHaveBeenCalledTimes(1)
})

test.each([
  [response(500, { request_id: 'body-id' }), 'body-id'],
  [{ ok: false, status: 500, json: async () => ({ request_id: 'missing-header-id' }) }, 'missing-header-id'],
  [response(500, {}), ''],
])('falls back to body request ID and a stable code when headers or code are absent', async (fixture, requestId) => {
  vi.spyOn(global, 'fetch').mockResolvedValue(fixture)
  await expect(requestJSON('/api/v1/admin/prompts')).rejects.toMatchObject({
    message: '请求失败（HTTP 500）', status: 500, code: 'API_ERROR', requestId,
  })
})

test('preserves HTTP failure and header request ID when the response is not JSON', async () => {
  vi.spyOn(global, 'fetch').mockResolvedValue({
    ok: false, status: 502, headers: new Headers({ 'X-Request-ID': 'gateway-id' }), json: async () => { throw new SyntaxError('HTML') },
  })
  await expect(requestJSON('/api/v1/admin/prompts')).rejects.toMatchObject({
    message: '请求失败（HTTP 502）', status: 502, code: 'API_ERROR', requestId: 'gateway-id',
  })
})

test('returns null for successful empty or malformed JSON', async () => {
  vi.spyOn(global, 'fetch').mockResolvedValue({ ok: true, status: 204, json: async () => { throw new SyntaxError('empty') } })
  await expect(requestJSON('/api/v1/admin/prompts')).resolves.toBeNull()
})

test('propagates network errors without manufacturing an HTTP error', async () => {
  const error = new TypeError('Failed to fetch')
  vi.spyOn(global, 'fetch').mockRejectedValue(error)
  await expect(requestJSON('/api/v1/admin/prompts')).rejects.toBe(error)
})
