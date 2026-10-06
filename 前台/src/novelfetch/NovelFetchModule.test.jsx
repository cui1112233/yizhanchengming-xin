// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import NovelFetchModule, { parseNovelIds } from './NovelFetchModule.jsx'

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query) => ({
    matches: false, media: query, onchange: null,
    addListener: () => {}, removeListener: () => {},
    addEventListener: () => {}, removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
})

function makeClient() {
  return {
    getConfig: vi.fn().mockResolvedValue({ config: { textModelId: 'model-1', maxText: 4000, targetVersions: ['original', 'ai1'], rewriteProfiles: {} } }),
    saveConfig: vi.fn(async (config) => ({ config })),
    listKnowledge: vi.fn().mockResolvedValue({ items: [] }),
    upsertKnowledge: vi.fn(async (item) => ({ item: { ...item, id: 'k1' } })),
    deleteKnowledge: vi.fn(),
    previewRules: vi.fn().mockResolvedValue({ text: '处理后' }),
    createBatch: vi.fn().mockResolvedValue({
      batch: { id: 'b1', name: '批次' },
      books: [
        { key: '4_1001', source: '阳光', platformId: '4', bookId: '1001', status: 'pending' },
        { key: '2_1001', source: '常读', platformId: '2', bookId: '1001', status: 'pending' },
      ],
    }),
    startRun: vi.fn().mockResolvedValue({ run: { id: 'r1', status: 'queued' } }),
    runBooks: vi.fn().mockResolvedValue({ books: [] }),
    retryBook: vi.fn().mockResolvedValue({ book: {} }),
    records: vi.fn().mockResolvedValue({ records: [] }),
    handoff: vi.fn().mockResolvedValue({ batchProjectId: 9 }),
    submitIntent: vi.fn().mockResolvedValue({ intentId: 7, status: 'pending' }),
  }
}

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('NovelFetchModule', () => {
  it('解析批量 ID 时支持多分隔符并去重', () => {
    expect(parseNovelIds('1001\n1002,1002；1003 1004')).toEqual(['1001', '1002', '1003', '1004'])
  })

  it('保留六个公网功能区', async () => {
    render(<NovelFetchModule client={makeClient()} />)
    for (const label of ['处理', '任务', '配置', '知识库', '处理规则', '记录']) {
      expect(await screen.findByRole('tab', { name: label })).toBeTruthy()
    }
  })

  it('支持多平台分组，相同 Book ID 跨平台共存，并进入统一 Run 队列', async () => {
    const client = makeClient()
    render(<NovelFetchModule client={client} />)
    await screen.findByText('model-1')

    fireEvent.change(screen.getByPlaceholderText('例如：阳光'), { target: { value: '阳光' } })
    fireEvent.change(screen.getByPlaceholderText('例如：4'), { target: { value: '4' } })
    fireEvent.change(screen.getByPlaceholderText('每行一个，也支持逗号和空格'), { target: { value: '1001\n1001' } })
    fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
    expect(screen.getByText(/阳光 1本/)).toBeTruthy()

    fireEvent.change(screen.getByPlaceholderText('例如：阳光'), { target: { value: '常读' } })
    fireEvent.change(screen.getByPlaceholderText('例如：4'), { target: { value: '2' } })
    fireEvent.change(screen.getByPlaceholderText('每行一个，也支持逗号和空格'), { target: { value: '1001' } })
    fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
    expect(screen.getByText(/常读 1本/)).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))
    await waitFor(() => expect(client.createBatch).toHaveBeenCalledTimes(1))
    expect(client.createBatch.mock.calls[0][0].groups).toHaveLength(2)
    await waitFor(() => expect(client.startRun).toHaveBeenCalledWith('b1', ''))
    expect(await screen.findByText(/Run：r1/)).toBeTruthy()
  })

  it('网络提交只创建发布意图，转入批量工厂走独立边界', async () => {
    const client = makeClient()
    render(<NovelFetchModule client={client} />)
    await screen.findByText('model-1')

    fireEvent.change(screen.getByPlaceholderText('例如：阳光'), { target: { value: '阳光' } })
    fireEvent.change(screen.getByPlaceholderText('例如：4'), { target: { value: '4' } })
    fireEvent.change(screen.getByPlaceholderText('每行一个，也支持逗号和空格'), { target: { value: '1001' } })
    fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))
    await screen.findByText(/Run：r1/)

    fireEvent.change(screen.getByPlaceholderText('发布账号 ID'), { target: { value: '5' } })
    fireEvent.click(screen.getByRole('button', { name: '提交网络' }))
    await waitFor(() => expect(client.submitIntent).toHaveBeenCalledWith('r1', '4_1001', { version: 'original', publishingAccountId: 5 }))

    fireEvent.click(screen.getByRole('button', { name: '转入批量工厂' }))
    await waitFor(() => expect(client.handoff).toHaveBeenCalledWith('r1'))
    expect(await screen.findByText(/批量工厂项目 #9/)).toBeTruthy()
  })
})
