import { beforeEach, describe, expect, it, vi } from 'vitest'

import { APIError, __resetAuthRecoveryForTests, requestJSON } from './api'

describe('API request correlation', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    __resetAuthRecoveryForTests()
  })

  it('captures X-Request-ID on APIError without dumping response metadata', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      headers: new Headers({ 'X-Request-ID': 'req_e2e_123' }),
      json: vi.fn().mockResolvedValue({ code: 'INTERNAL_ERROR', message: '服务暂时不可用', request_id: 'body-id', payload: 'secret-canary' }),
    })

    let caught
    try {
      await requestJSON('/api/v1/example', { skipAuthRecovery: true })
    } catch (error) {
      caught = error
    }

    expect(caught).toBeInstanceOf(APIError)
    expect(caught.requestId).toBe('req_e2e_123')
    expect(caught.message).toBe('服务暂时不可用')
    expect(caught.payload.payload).toBe('secret-canary')
    expect(String(caught)).not.toContain('secret-canary')
  })

  it('uses payload correlation only when the header is absent', async () => {
    global.fetch = vi.fn().mockResolvedValue({ ok: false, status: 503, headers: new Headers(), json: async () => ({ message: '服务暂时不可用', request_id: 'req_body_123' }) })
    await expect(requestJSON('/api/test', { skipAuthRecovery: true })).rejects.toMatchObject({ requestId: 'req_body_123' })
  })
})
