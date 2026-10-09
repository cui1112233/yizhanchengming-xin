// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'
import { getGenerationRun, retryGenerationStage, runBookGeneration, runProjectGeneration } from './api.js'

const ok = (payload, status = 202) => ({ ok: true, status, headers: new Headers(), json: async () => payload })

describe('generation async API contract', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('sends the same explicit idempotency key in header and compatibility body', async () => {
    const fetch = vi.fn().mockResolvedValue(ok({ runId: 44 }))
    vi.stubGlobal('fetch', fetch)
    await runProjectGeneration(3, { bookIds: [7] }, { idempotencyKey: 'stable-key' })
    await runBookGeneration(3, 7, { directorMode: 'normal' }, { idempotencyKey: 'stable-key-2' })
    await retryGenerationStage(3, 7, 'DIRECTOR', { sourceBookRunId: 81 }, { idempotencyKey: 'stable-key-3' })

    for (const [index, key] of ['stable-key', 'stable-key-2', 'stable-key-3'].entries()) {
      const options = fetch.mock.calls[index][1]
      expect(options.headers['Idempotency-Key']).toBe(key)
      expect(JSON.parse(options.body).requestId).toBe(key)
    }
    expect(JSON.parse(fetch.mock.calls[2][1].body).sourceBookRunId).toBe(81)
  })

  it('reads the exact accepted run and forwards AbortSignal', async () => {
    const fetch = vi.fn().mockResolvedValue(ok({ runId: 44, terminal: false }, 200))
    vi.stubGlobal('fetch', fetch)
    const controller = new AbortController()
    await getGenerationRun(3, 44, { signal: controller.signal })
    expect(fetch).toHaveBeenCalledWith('/api/v1/batch-projects/3/generation/runs/44', expect.objectContaining({ signal: controller.signal, credentials: 'include' }))
  })
})
