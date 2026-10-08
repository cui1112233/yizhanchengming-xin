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

  it('uses the safe generation envelope while retaining structured results and current header correlation', async () => {
    const payload = { code: 'GENERATION_FAILED', message: '生成阶段执行失败，请稍后重试', error: '生成阶段执行失败，请稍后重试', request_id: 'body-id', run: { id: 17, requestId: 'execution-old' }, stages: [{ attempt: 2 }] }
    global.fetch = vi.fn().mockResolvedValue({ ok: false, status: 500, headers: new Headers({ 'X-Request-ID': 'http-current' }), json: async () => payload })
    await expect(requestJSON('/api/v1/batch-projects/3/books/11/generation', { skipAuthRecovery: true })).rejects.toMatchObject({ message: payload.message, requestId: 'http-current', payload })
  })

  it('retains 207 mixed generation outcomes as a successful structured response', async () => {
    const payload = { batchProjectId: 3, completed: 1, failed: 1, books: [{ bookId: 11, errorCode: 'GENERATION_FAILED', error: '生成阶段执行失败，请稍后重试' }, { bookId: 12, run: { status: 'completed' } }] }
    global.fetch = vi.fn().mockResolvedValue({ ok: true, status: 207, headers: new Headers({ 'X-Request-ID': 'http-current' }), json: async () => payload })
    expect(await requestJSON('/api/v1/batch-projects/3/generation', { skipAuthRecovery: true })).toEqual(payload)
  })
})
