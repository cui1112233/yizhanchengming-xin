import { requestJSON } from '../api.js'

export function createNovelFetchClient(request = requestJSON, prefix = '/api/v1/novel-fetch') {
  const path = (value) => `${prefix}${value}`
  return {
    createBatch(input) {
      return request(path('/batches'), { method: 'POST', body: JSON.stringify(input) })
    },
    startRun(batchId, runAt = '') {
      return request(path(`/batches/${encodeURIComponent(batchId)}/runs`), {
        method: 'POST',
        body: JSON.stringify(runAt ? { runAt } : {}),
      })
    },
    getRun(runId) {
      return request(path(`/runs/${encodeURIComponent(runId)}`))
    },
    retryBook(runId, bookKey) {
      return request(path(`/runs/${encodeURIComponent(runId)}/books/${encodeURIComponent(bookKey)}/retry`), { method: 'POST' })
    },
    records(runId) {
      return request(path(`/runs/${encodeURIComponent(runId)}/records`))
    },
    history() {
      return request(path('/history'))
    },
    handoff(runId) {
      return request(path(`/runs/${encodeURIComponent(runId)}/handoff`), { method: 'POST' })
    },
    submitIntent(runId, bookKey, input) {
      return request(path(`/runs/${encodeURIComponent(runId)}/books/${encodeURIComponent(bookKey)}/submit-intents`), {
        method: 'POST',
        body: JSON.stringify(input),
      })
    },
    getConfig() {
      return request(path('/config'))
    },
    saveConfig(config) {
      return request(path('/config'), { method: 'PUT', body: JSON.stringify(config) })
    },
    listKnowledge(kind = '') {
      const suffix = kind ? `?kind=${encodeURIComponent(kind)}` : ''
      return request(path(`/knowledge${suffix}`))
    },
    createKnowledge(item) {
      const { id: _id, ...payload } = item || {}
      return request(path('/knowledge'), {
        method: 'POST',
        body: JSON.stringify(payload),
      })
    },
    updateKnowledge(item) {
      if (!item?.id) throw new Error('知识条目 ID 不能为空')
      return request(path(`/knowledge/${encodeURIComponent(item.id)}`), {
        method: 'PUT',
        body: JSON.stringify(item),
      })
    },
    deleteKnowledge(kind, id) {
      return request(path(`/knowledge/${encodeURIComponent(id)}?kind=${encodeURIComponent(kind || '')}`), { method: 'DELETE' })
    },
    previewRules(text, config) {
      return request(path('/rules/preview'), {
        method: 'POST',
        body: JSON.stringify({ text, config }),
      })
    },
  }
}

export const novelFetchClient = createNovelFetchClient()
