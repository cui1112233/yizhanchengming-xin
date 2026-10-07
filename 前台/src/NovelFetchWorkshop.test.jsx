// @vitest-environment jsdom

import React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import NovelFetchWorkshop from './NovelFetchWorkshop.jsx'

function response(value, status = 200) { return { ok:status < 300, status, json:async()=>value } }

describe('Novel Fetch Workshop', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/novel-fetch-workshop?intakeId=7')
    global.fetch = vi.fn(async (url, options = {}) => {
      if (url === '/api/v1/intakes') return response({ intakes:[{ id:7, name:'恢复任务', status:'partial_failed' }] })
      if (url === '/api/v1/intakes/7/workshop' && (!options.method || options.method === 'GET')) return response({ intake:{ id:7, name:'恢复任务', status:'partial_failed' }, books:[{ id:11, bookId:'1001', title:'已恢复', source:'番茄', gender:'男频', style:'都市', status:'fetched', originalText:'服务端原文' },{ id:12, bookId:'1002', title:'待重试', source:'知乎', status:'retryable_failed' }], settings:{ processingRules:'规则' }, prompts:[{ key:'script.default', version:2, enabled:true }] })
      if (url === '/api/v1/intakes/7/workshop' && options.method === 'PUT') return response({ intakeId:7, settings:JSON.parse(options.body).settings })
      if (url === '/api/v1/intakes/7/execute' && options.method === 'POST') return response({ intakeId:7, status:'completed' })
      if (url === '/api/v1/intakes/7/books/11/restore' && options.method === 'POST') return response({ book:{ id:11, status:'fetched', originalText:'服务端原文' } })
      return response({})
    })
  })
  it('restores intake facts, server prompts and persisted settings without browser storage', async () => {
    render(<NovelFetchWorkshop />)
    expect(await screen.findByTitle('#7 · 恢复任务 · partial_failed')).toBeTruthy()
    expect(screen.getByText('script.default v2')).toBeTruthy()
    expect(screen.getByLabelText('处理规则').value).toBe('规则')
  })

  it('opens the restored original text in its own drawer', async () => {
    render(<NovelFetchWorkshop />)
    await screen.findByTitle('#7 · 恢复任务 · partial_failed')
    fireEvent.click(screen.getAllByRole('button', { name:'查看原文' })[0])
    expect(await screen.findByText('服务端原文')).toBeTruthy()
  })

  it('saves edited settings through the workshop API', async () => {
    render(<NovelFetchWorkshop />)
    await screen.findByTitle('#7 · 恢复任务 · partial_failed')
    fireEvent.change(screen.getByLabelText('处理规则'), { target:{ value:'新规则' } })
    fireEvent.click(screen.getByRole('button', { name:'保存配置' }))
    await waitFor(() => expect(global.fetch.mock.calls.some(([url,options]) => url === '/api/v1/intakes/7/workshop' && options.method === 'PUT')).toBe(true))
  })
})
