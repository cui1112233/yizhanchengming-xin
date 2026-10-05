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
      json: vi.fn().mockResolvedValue({ code: 'INTERNAL_ERROR', message: '服务暂时不可用' }),
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
  })
})
