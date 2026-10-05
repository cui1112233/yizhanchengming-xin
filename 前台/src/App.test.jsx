// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
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
  cleanup()
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

  it('批量工厂列表展示真实项目汇总信息', async () => {
    window.history.replaceState({}, '', '/batch-factory')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() =>
        jsonResponse({
          projects: [
            {
              id: 52,
              intakeId: 12,
              name: '跨书城批次',
              sources: ['点众', '知乎'],
              bookCount: 3,
              genders: ['女频', '男频'],
              styles: ['情感', '悬疑'],
              runStatus: 'running',
            },
            {
              id: 51,
              intakeId: 11,
              name: '知乎批次',
              sources: ['知乎'],
              bookCount: 1,
              genders: ['男频'],
              styles: ['都市'],
              runStatus: 'pending',
            },
          ],
        }),
      )

    render(<IntakeWorkbench />)

    expect(await screen.findByText('跨书城批次')).toBeTruthy()
    expect(screen.getByText('知乎批次')).toBeTruthy()
    expect(screen.getByText('点众')).toBeTruthy()
    expect(screen.getAllByText('知乎').length).toBeGreaterThan(0)
    expect(screen.getByText('3')).toBeTruthy()
    expect(screen.getByText('女频')).toBeTruthy()
    expect(screen.getAllByText('男频').length).toBeGreaterThan(0)
    expect(screen.getByText('情感')).toBeTruthy()
    expect(screen.getByText('悬疑')).toBeTruthy()
    expect(screen.getByText('执行中')).toBeTruthy()
    expect(screen.getByText('待执行')).toBeTruthy()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1))
    expect(String(fetchMock.mock.calls[0][0])).toBe('/api/v1/batch-projects')
  }, 15000)

  it('Task 8 创建项目后下一次进入批量工厂会重新从列表 API 读取新项目', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() =>
        jsonResponse({ intake: { id: 18 }, books: [] }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 18, status: 'completed', fetched: 1, failed: 0 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 3,
              intakeId: 18,
              source: '知乎',
              platformId: '15',
              bookId: '2001',
              title: '刚创建的小说',
              gender: '女频',
              style: '情感',
              status: 'fetched',
              errorMessage: '',
            },
          ],
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          project: { id: 88, intakeId: 18, name: '刚创建的批次' },
          run: { id: 99, batchProjectId: 88, runAt: '2026-10-05T08:00:00Z', status: 'pending' },
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          projects: [
            {
              id: 88,
              intakeId: 18,
              name: '刚创建的批次',
              sources: ['知乎'],
              bookCount: 1,
              genders: ['女频'],
              styles: ['情感'],
              runStatus: 'pending',
            },
          ],
        }),
      )

    const intakeView = render(<IntakeWorkbench />)

    fireEvent.change(screen.getByPlaceholderText('不填则自动生成'), {
      target: { value: '刚创建的批次' },
    })
    fireEvent.change(screen.getByPlaceholderText('例如：阳光、常读、知乎'), {
      target: { value: '知乎' },
    })
    fireEvent.change(screen.getByPlaceholderText('例如：4'), {
      target: { value: '15' },
    })
    fireEvent.change(screen.getByPlaceholderText(/每行一个/), {
      target: { value: '2001' },
    })
    fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(await screen.findByText('立即执行任务已创建。')).toBeTruthy()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(4))

    intakeView.unmount()
    cleanup()
    window.history.replaceState({}, '', '/batch-factory')

    render(<IntakeWorkbench />)

    expect(await screen.findByText('刚创建的批次')).toBeTruthy()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(5))
    expect(String(fetchMock.mock.calls[4][0])).toBe('/api/v1/batch-projects')
  }, 15000)
})
