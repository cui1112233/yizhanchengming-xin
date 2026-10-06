// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
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

function addStore(source, platformId, ids) {
  fireEvent.change(screen.getByPlaceholderText('例如：阳光、常读、知乎'), {
    target: { value: source },
  })
  fireEvent.change(screen.getByPlaceholderText('例如：4'), {
    target: { value: platformId },
  })
  const bookInput = screen.getByPlaceholderText(/每行一个/)
  fireEvent.change(bookInput, { target: { value: ids } })
  fireEvent.click(screen.getByRole('button', { name: '添加书城' }))
  return bookInput
}

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/')
})

describe('Task 11 水货生产 / 小说获取入口衔接', () => {
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

  it('/shuihuo-production 可以直接渲染，并且不再出现“解析输入”', () => {
    window.history.replaceState({}, '', '/shuihuo-production')
    render(<IntakeWorkbench />)

    expect(screen.getByRole('heading', { name: '小说获取工作台' })).toBeTruthy()
    expect(screen.queryByText('解析输入')).toBeNull()
  })

  it('水货生产可直接进入 Batch Factory，且列表从真实 API 契约读取', async () => {
    window.history.replaceState({}, '', '/shuihuo-production')
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementationOnce(() =>
      jsonResponse({
        projects: [{ id: 52, intakeId: 12, name: '刚创建的批量项目' }],
      }),
    )

    render(<IntakeWorkbench />)
    fireEvent.click(screen.getByRole('button', { name: '批量工厂' }))

    expect(await screen.findByRole('heading', { name: '批量工厂' })).toBeTruthy()
    expect(await screen.findByText('刚创建的批量项目')).toBeTruthy()
    expect(window.location.pathname).toBe('/batch-factory')
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1))
    expect(String(fetchMock.mock.calls[0][0])).toBe('/api/v1/batch-projects')
  }, 15000)

  it('多书城可顺序添加；同书城 Book ID 去重并合并标签；添加后输入框清空；总数正确', () => {
    window.history.replaceState({}, '', '/shuihuo-production')
    render(<IntakeWorkbench />)

    let bookInput = addStore('知乎', '11', '1001\n1002\n1002')
    expect(bookInput.value).toBe('')
    expect(screen.getByText(/知乎 2本/)).toBeTruthy()

    bookInput = addStore('知乎', '11', '1002\n1003')
    expect(bookInput.value).toBe('')
    expect(screen.getAllByText(/知乎 3本/)).toHaveLength(1)
    expect(screen.queryByText(/知乎 1本/)).toBeNull()

    // 不同书城允许业务上相同的 Book ID。
    addStore('点众', '22', '1001\n2001')
    expect(screen.getByText(/点众 2本/)).toBeTruthy()

    const stats = screen.getByText('已选小说').closest('.ant-statistic')
    expect(stats).toBeTruthy()
    expect(within(stats).getByText('5')).toBeTruthy()
  }, 15000)

  it('121 完成后展示完整真实字段，创建 BatchProject/Run，并出现进入批量工厂入口', async () => {
    window.history.replaceState({}, '', '/shuihuo-production')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => jsonResponse({ intake: { id: 17 }, books: [] }))
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 17, status: 'completed', fetched: 1, failed: 0 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 1,
              intakeId: 17,
              source: '知乎',
              platformId: '11',
              bookId: '1001',
              title: '测试小说',
              category: '悬疑',
              genre: '都市悬疑',
              gender: '男频',
              style: '爽文',
              status: 'fetched',
              errorMessage: '',
            },
          ],
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          project: { id: 88, intakeId: 17, name: '测试批次' },
          run: { id: 99, status: 'queued', runAt: '2026-10-05T12:00:00Z' },
        }),
      )

    render(<IntakeWorkbench />)
    fireEvent.change(screen.getByPlaceholderText('不填则自动生成'), {
      target: { value: '测试批次' },
    })
    addStore('知乎', '11', '1001')
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(await screen.findByText('测试小说')).toBeTruthy()
    expect(screen.getByText('知乎')).toBeTruthy()
    expect(screen.getByText('11')).toBeTruthy()
    expect(screen.getByText('1001')).toBeTruthy()
    expect(screen.getByText('悬疑')).toBeTruthy()
    expect(screen.getByText('都市悬疑')).toBeTruthy()
    expect(screen.getByText('男频')).toBeTruthy()
    expect(screen.getByText('爽文')).toBeTruthy()
    expect(screen.getByText('已获取')).toBeTruthy()
    expect(await screen.findByText(/#88 · 测试批次/)).toBeTruthy()
    expect(await screen.findByText(/#99/)).toBeTruthy()
    expect(screen.getByRole('button', { name: '进入批量工厂' })).toBeTruthy()

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(4))
    expect(
      fetchMock.mock.calls.some(([url]) => String(url).endsWith('/batch-projects')),
    ).toBe(true)
  }, 15000)

  it('缺失的小说字段明确显示 -，不由前端猜测补值', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => jsonResponse({ intake: { id: 18 }, books: [] }))
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 18, status: 'partial_failed', fetched: 0, failed: 1 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 2,
              intakeId: 18,
              source: '',
              platformId: '',
              bookId: '',
              title: '',
              category: '',
              genre: '',
              gender: '',
              style: '',
              status: 'retryable_failed',
              errorMessage: '无法识别元数据',
            },
          ],
        }),
      )

    render(<IntakeWorkbench />)
    addStore('黑岩', '33', '3001')
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(await screen.findByText('无法识别元数据')).toBeTruthy()
    expect(screen.getAllByText('-').length).toBeGreaterThanOrEqual(8)
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3))
  }, 15000)

  it('单本错误包含敏感信息时会脱敏，不直接显示 Token/密码/DSN', async () => {
    const sensitiveError = 'Authorization: Bearer secret-token password=supersecret mysql://user:pass@db/app'
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => jsonResponse({ intake: { id: 19 }, books: [] }))
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 19, status: 'partial_failed', fetched: 0, failed: 1 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 3,
              intakeId: 19,
              source: '知乎',
              platformId: '11',
              bookId: '9001',
              title: '',
              category: '',
              genre: '',
              gender: '',
              style: '',
              status: 'retryable_failed',
              errorMessage: sensitiveError,
            },
          ],
        }),
      )

    render(<IntakeWorkbench />)
    addStore('知乎', '11', '9001')
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(await screen.findByText('请求失败，请稍后重试或查看服务端日志。')).toBeTruthy()
    expect(screen.queryByText(sensitiveError)).toBeNull()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3))
  }, 15000)

  it('121 部分失败时显示真实成功/失败数量和每本错误，不创建 BatchProject', async () => {
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
    addStore('阳光', '4', '1001\n1002')
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    expect(
      await screen.findByText(/本批次未全部成功（成功 1，失败 1），已停止创建生产任务/),
    ).toBeTruthy()
    expect(await screen.findByText('成功小说')).toBeTruthy()
    expect(await screen.findByText('121 upstream error')).toBeTruthy()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3))
    expect(
      fetchMock.mock.calls.some(([url]) => String(url).endsWith('/batch-projects')),
    ).toBe(false)
  }, 15000)

  it('刷新后按 intake 查询 MySQL 状态，并可只重试失败书', async () => {
    window.history.replaceState({}, '', '/novel-fetch?intake=17')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 2,
              intakeId: 17,
              source: '阳光',
              platformId: '4',
              bookId: '1002',
              title: '失败小说',
              status: 'retryable_failed',
              errorMessage: '121 timeout',
            },
          ],
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          book: { id: 2, intakeId: 17, status: 'fetched' },
          summary: { intakeId: 17, status: 'completed', fetched: 1, failed: 0 },
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 2,
              intakeId: 17,
              source: '阳光',
              platformId: '4',
              bookId: '1002',
              title: '失败小说',
              status: 'fetched',
              errorMessage: '',
            },
          ],
        }),
      )

    render(<IntakeWorkbench />)

    expect(await screen.findByText('已从 MySQL 恢复 Intake #17 的书籍状态。')).toBeTruthy()
    expect(await screen.findByText('121 timeout')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '重试失败书' }))

    expect(await screen.findByText('失败小说 重试成功。')).toBeTruthy()
    expect(screen.queryByText('121 timeout')).toBeNull()
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3))
    expect(String(fetchMock.mock.calls[0][0])).toBe('/api/v1/intakes/17/books')
    expect(String(fetchMock.mock.calls[1][0])).toBe('/api/v1/intakes/17/books/2/retry')
    expect(fetchMock.mock.calls[1][1]?.method).toBe('POST')
  }, 15000)

  it('成功创建后点击进入批量工厂，列表会重新从 API 读取最新项目', async () => {
    window.history.replaceState({}, '', '/shuihuo-production')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => jsonResponse({ intake: { id: 20 }, books: [] }))
      .mockImplementationOnce(() =>
        jsonResponse({ intakeId: 20, status: 'completed', fetched: 1, failed: 0 }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          books: [
            {
              id: 7,
              intakeId: 20,
              source: '点众',
              platformId: '22',
              bookId: '7001',
              title: '新项目小说',
              category: '都市',
              genre: '都市',
              gender: '男频',
              style: '爽文',
              status: 'fetched',
              errorMessage: '',
            },
          ],
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({
          project: { id: 101, intakeId: 20, name: '最新项目' },
          run: { id: 102, status: 'queued', runAt: '2026-10-05T12:00:00Z' },
        }),
      )
      .mockImplementationOnce(() =>
        jsonResponse({ projects: [{ id: 101, intakeId: 20, name: '最新项目' }] }),
      )

    render(<IntakeWorkbench />)
    fireEvent.change(screen.getByPlaceholderText('不填则自动生成'), {
      target: { value: '最新项目' },
    })
    addStore('点众', '22', '7001')
    fireEvent.click(screen.getByRole('button', { name: '立即执行' }))

    const enterButton = await screen.findByRole('button', { name: '进入批量工厂' })
    fireEvent.click(enterButton)

    expect(await screen.findByRole('heading', { name: '批量工厂' })).toBeTruthy()
    expect(await screen.findByText('最新项目')).toBeTruthy()
    expect(window.location.pathname).toBe('/batch-factory')
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(5))
  }, 15000)
})
