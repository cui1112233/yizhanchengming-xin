import { afterEach, describe, expect, it, vi } from 'vitest'
import { archiveBatchProject, listBatchProjects, restoreBatchProject } from './api.js'

describe('BatchProject API boundary', () => {
  afterEach(() => vi.unstubAllGlobals())
  it('serializes catalog filters and pagination through the shared request boundary', async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ projects: [], total: 0 }) })
    vi.stubGlobal('fetch', fetch)
    await listBatchProjects({ q: '项目 & 来源', source: '知乎付费', status: 'failed', archived: 'archived', page: 2, limit: 12, sort: 'name_asc' })
    const [url, options] = fetch.mock.calls[0]
    expect(Object.fromEntries(new URL(url, 'http://localhost').searchParams)).toEqual({ q: '项目 & 来源', source: '知乎付费', status: 'failed', archived: 'archived', page: '2', limit: '12', sort: 'name_asc' })
    expect(options.credentials).toBe('include')
  })
  it('archives and restores with POST through the Cookie request boundary', async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) })
    vi.stubGlobal('fetch', fetch)
    await archiveBatchProject(6)
    await restoreBatchProject(6)
    expect(fetch.mock.calls.map(([url, options]) => [url, options.method, options.credentials])).toEqual([
      ['/api/v1/batch-projects/6/archive', 'POST', 'include'],
      ['/api/v1/batch-projects/6/restore', 'POST', 'include'],
    ])
  })
})
