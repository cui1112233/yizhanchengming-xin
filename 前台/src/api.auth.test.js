import { beforeEach, describe, expect, it, vi } from 'vitest'
import { __resetAuthRecoveryForTests, requestJSON } from './api'

function jsonResponse(status, payload) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('auth-aware API client', () => {
  beforeEach(() => {
    __resetAuthRecoveryForTests()
    vi.restoreAllMocks()
  })

  it('uses one refresh flight for concurrent 401s and replays both requests', async () => {
    let refreshCalls = 0
    const attempts = new Map()
    global.fetch = vi.fn(async (url, options = {}) => {
      if (url === '/api/auth/refresh') {
        refreshCalls += 1
        return jsonResponse(200, { user: { id: 7, name: 'Alice' } })
      }
      const count = (attempts.get(url) || 0) + 1
      attempts.set(url, count)
      if (count === 1) return jsonResponse(401, { code: 'AUTH_UNAUTHENTICATED', message: 'expired' })
      return jsonResponse(200, { ok: true, url, credentials: options.credentials })
    })

    const [first, second] = await Promise.all([
      requestJSON('/api/v1/batch-projects'),
      requestJSON('/api/v1/intakes'),
    ])

    expect(refreshCalls).toBe(1)
    expect(first.ok).toBe(true)
    expect(second.ok).toBe(true)
    expect(first.credentials).toBe('include')
    expect(second.credentials).toBe('include')
    expect(attempts.get('/api/v1/batch-projects')).toBe(2)
    expect(attempts.get('/api/v1/intakes')).toBe(2)
  })

  it('does not treat 403 as an expired login', async () => {
    let refreshCalls = 0
    global.fetch = vi.fn(async (url) => {
      if (url === '/api/auth/refresh') {
        refreshCalls += 1
        return jsonResponse(200, {})
      }
      return jsonResponse(403, { code: 'AUTH_FORBIDDEN', message: '你没有执行此操作的权限' })
    })

    await expect(requestJSON('/api/v1/batch-projects')).rejects.toMatchObject({
      status: 403,
      code: 'AUTH_FORBIDDEN',
      message: '你没有执行此操作的权限',
    })
    expect(refreshCalls).toBe(0)
  })
})
