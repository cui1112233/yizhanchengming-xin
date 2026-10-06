import { describe, expect, it, vi } from 'vitest'
import { createNovelFetchClient } from './client.js'

describe('novel fetch client knowledge contract', () => {
  it('连续新增走 POST /knowledge，不发送伪造 new id；编辑分别 PUT 各自 id', async () => {
    let next = 1
    const calls = []
    const request = vi.fn(async (path, options = {}) => {
      calls.push({ path, options })
      const body = options.body ? JSON.parse(options.body) : {}
      if (path.endsWith('/knowledge') && options.method === 'POST') {
        return { item: { ...body, id: `k${next++}` } }
      }
      const id = path.split('/').pop()
      return { item: { ...body, id } }
    })
    const client = createNovelFetchClient(request)

    const first = await client.createKnowledge({ id: '', kind: 'rewrite', title: 'A', content: 'one', enabled: true })
    const second = await client.createKnowledge({ id: '', kind: 'rewrite', title: 'B', content: 'two', enabled: true })
    expect(first.item.id).toBe('k1')
    expect(second.item.id).toBe('k2')
    expect(calls[0].path).toBe('/api/v1/novel-fetch/knowledge')
    expect(calls[0].options.method).toBe('POST')
    expect(JSON.parse(calls[0].options.body).id).toBeUndefined()
    expect(calls[1].path).toBe('/api/v1/novel-fetch/knowledge')

    await client.updateKnowledge({ ...first.item, content: 'one-edited' })
    await client.updateKnowledge({ ...second.item, content: 'two-edited' })
    expect(calls[2].path).toBe('/api/v1/novel-fetch/knowledge/k1')
    expect(calls[3].path).toBe('/api/v1/novel-fetch/knowledge/k2')
  })
})
