// @vitest-environment jsdom

import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import IntakeWorkbench, { parseBookIds } from './App.jsx'

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
})

function jsonResponse(payload) {
  return Promise.resolve({
    ok: true,
    status: 200,
    json: async () => payload,
  })
}

afterEach(() => {
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/')
})

describe('小说获取工作台行为', () => {
  it('解析 Book ID 时支持多种分隔符并去重', () => {
    expect(parseBookIds('1001\n1002,1002，1003;1004；1005 1006')).toEqual([
      '1001',
      '1002',
      '1003',
      '1004',
      '1005',
      '1006',
    ])
  })

  it('121 部分失败时停止在结果页，不创建 BatchProject', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => jsonResponse({ intake: { id: 17 }, books: [] }))
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 17, status: 'partial_failed', fetched: 1, failed: 1 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 1,
              intakeId: 17,
              source: '阳光',
              platformId: '4',
              bookId: '1001',
              title: '成功小说',
              category: '男生生活',
              genre: '都市',
              gender: '男频',
              style: '',
              status: 'fetched',
              errorMessage: '',
            },
            {
              id: 2,
              intakeId: 17,
              source: '阳光',
              platformId: '4',
              bookId: '1002',
              title: '',
              category: '',
              genre: '',
              gender: '',
              style: '',
              status: 'retryable_failed',
              errorMessage: '121 upstream error',
            },
          ],
        }),
      )

    render(<IntakeWorkbench />)

    fireEvent.change(screen.getByPlaceholderText('例如：阳光、常读、知乎'), {
      target: { value: '阳光' },
    })
    fireEvent.change(screen.getByPlaceholderText('例如：4'), {
      target: { value: '4' },
    })
    fireEvent.change(screen.getByPlaceholderText(/每行一个/), {
      target: { value: '1001\n1002' },
    })
    fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(
      await screen.findByText(/本批次未全部成功（成功 1，失败 1），已停止创建生产任务/),
    ).toBeTruthy()

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3))
    expect(
      fetchMock.mock.calls.some(([url]) => String(url).endsWith('/batch-projects')),
    ).toBe(false)
    expect(await screen.findByText('成功小说')).toBeTruthy()
    expect(await screen.findByText('121 upstream error')).toBeTruthy()
  }, 15000)

  it('批量工厂列表从 API 展示项目名称', async () => {
    window.history.replaceState({}, '', '/batch-factory')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() =>
        jsonResponse({
          projects: [
            { id: 52, intakeId: 12, name: '点众批次' },
            { id: 51, intakeId: 11, name: '知乎批次' },
          ],
        }),
      )

    render(<IntakeWorkbench />)

    expect(await screen.findByText('知乎批次')).toBeTruthy()
    expect(screen.getByText('点众批次')).toBeTruthy()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1))
    expect(String(fetchMock.mock.calls[0][0])).toBe('/api/v1/batch-projects')
  }, 15000)
})
