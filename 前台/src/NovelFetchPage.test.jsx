// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import NovelFetchPage, { summarizeNovelFetchBooks } from './NovelFetchPage.jsx'

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

function jsonResponse(payload, status = 200) {
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status,
    headers: { get: () => '' },
    json: async () => payload,
  })
}

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/novel-fetch')
})

describe('Novel Fetch 主页面', () => {
  it('从 Go API 恢复 MySQL 中的任务、进度、书籍结果和失败原因', async () => {
    window.history.replaceState({}, '', '/novel-fetch')
    global.fetch = vi.fn((url) => {
      if (url === '/api/v1/intakes') {
        return jsonResponse({
          intakes: [{ id: 42, name: '恢复批次', status: 'partial_failed', updatedAt: '2026-10-07T01:00:00Z' }],
        })
      }
      if (url === '/api/v1/intakes/42/books') {
        return jsonResponse({
          books: [
            { id: 1, intakeId: 42, source: '番茄付费', platformId: '2', bookId: '1001', title: '成功书', category: '都市', genre: '8', gender: '男频', style: '现代', status: 'fetched' },
            { id: 2, intakeId: 42, source: '知乎付费', platformId: '15', bookId: '1002', title: '失败书', status: 'retryable_failed', errorMessage: 'upstream timeout' },
          ],
        })
      }
      return jsonResponse({})
    })

    render(<NovelFetchPage />)

    expect(await screen.findByText('恢复批次')).toBeTruthy()
    expect(screen.getByText(/成功 1\/2/)).toBeTruthy()
    expect(screen.getByText('upstream timeout')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: '查看结果' }))
    expect(await screen.findByText('1001')).toBeTruthy()
    expect(screen.getByText('成功书')).toBeTruthy()
    expect(screen.getByText('1002')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'AI 处理配置' }).getAttribute('href')).toBe('/novel-fetch-workshop')
    expect(screen.getByRole('link', { name: '处理规则' }).getAttribute('href')).toBe('/novel-fetch-workshop')
  })

  it('重试复用同一 Intake execute 链路，并在成功后恢复完成状态', async () => {
    let retried = false
    const calls = []
    global.fetch = vi.fn((url, options = {}) => {
      calls.push([url, options.method || 'GET'])
      if (url === '/api/v1/intakes') {
        return jsonResponse({
          intakes: [{ id: 9, name: '可重试批次', status: retried ? 'completed' : 'partial_failed', updatedAt: '2026-10-07T01:00:00Z' }],
        })
      }
      if (url === '/api/v1/intakes/9/books') {
        return jsonResponse({
          books: retried
            ? [
                { id: 1, intakeId: 9, source: '番茄付费', platformId: '2', bookId: '2001', title: 'A', status: 'fetched' },
                { id: 2, intakeId: 9, source: '知乎付费', platformId: '15', bookId: '2002', title: 'B', status: 'fetched' },
              ]
            : [
                { id: 1, intakeId: 9, source: '番茄付费', platformId: '2', bookId: '2001', title: 'A', status: 'fetched' },
                { id: 2, intakeId: 9, source: '知乎付费', platformId: '15', bookId: '2002', title: 'B', status: 'retryable_failed', errorMessage: 'temporary timeout' },
              ],
        })
      }
      if (url === '/api/v1/intakes/9/execute' && options.method === 'POST') {
        retried = true
        return jsonResponse({ intakeId: 9, status: 'completed', fetched: 2, failed: 0 })
      }
      if (url === '/api/v1/intakes/9/batch-projects' && options.method === 'POST') {
        return jsonResponse({
          project: { id: 77, intakeId: 9, name: '可重试批次' },
          run: { id: 88, batchProjectId: 77, status: 'pending', runAt: '2026-10-07T01:02:00Z' },
        }, 201)
      }
      return jsonResponse({})
    })

    render(<NovelFetchPage />)
    const retryButton = await screen.findByRole('button', { name: '重试失败项' })
    expect(retryButton.disabled).toBe(false)
    fireEvent.click(retryButton)

    expect(await screen.findByText('失败项重试完成，批次已恢复为完成状态。')).toBeTruthy()
    await waitFor(() => {
      expect(calls.some(([url, method]) => url === '/api/v1/intakes/9/execute' && method === 'POST')).toBe(true)
    })
    expect(screen.getByText(/成功 2\/2/)).toBeTruthy()
  })

  it('403 只显示无权限，不触发认证刷新或强制退出', async () => {
    const calls = []
    global.fetch = vi.fn((url, options = {}) => {
      calls.push([url, options.method || 'GET'])
      if (url === '/api/v1/intakes') {
        return jsonResponse({ code: 'AUTH_FORBIDDEN', message: '你没有执行此操作的权限' }, 403)
      }
      return jsonResponse({})
    })

    render(<NovelFetchPage />)

    expect(await screen.findByText('无权限查看小说获取任务。')).toBeTruthy()
    expect(calls.some(([url]) => url === '/api/auth/refresh')).toBe(false)
  })

  it('进度统计区分成功、失败和待处理', () => {
    expect(summarizeNovelFetchBooks([
      { status: 'fetched' },
      { status: 'retryable_failed' },
      { status: 'pending' },
      { status: 'fetched' },
    ])).toEqual({ total: 4, fetched: 2, failed: 1, pending: 1, percent: 50 })
  })
})
