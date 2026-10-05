// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import BatchProjectListPage from './BatchProjectListPage.jsx'

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
})

it('点击真实项目进入 V11 工作台并展示全部小说状态与错误', async () => {
  const fetchMock = vi
    .spyOn(globalThis, 'fetch')
    .mockImplementationOnce(() =>
      jsonResponse({
        projects: [
          {
            id: 51,
            intakeId: 11,
            name: '真实批次',
            sources: ['知乎', '点众'],
            bookCount: 2,
            genders: ['女频', '男频'],
            styles: ['情感', '悬疑'],
            runStatus: 'running',
          },
        ],
      }),
    )
    .mockImplementationOnce(() =>
      jsonResponse({
        project: { id: 51, intakeId: 11, name: '真实批次' },
        books: [
          {
            id: 31,
            intakeId: 11,
            bookId: '1001',
            title: '成功小说',
            source: '知乎',
            platformId: '15',
            gender: '女频',
            style: '情感',
            status: 'fetched',
            errorMessage: '',
          },
          {
            id: 32,
            intakeId: 11,
            bookId: '1002',
            title: '失败小说',
            source: '点众',
            platformId: '7',
            gender: '男频',
            style: '悬疑',
            status: 'retryable_failed',
            errorMessage: '121 upstream error',
          },
        ],
      }),
    )

  render(<BatchProjectListPage />)

  const projectName = await screen.findByText('真实批次')
  fireEvent.click(projectName)

  expect(await screen.findByText('Batch Factory V11 工作台')).toBeTruthy()
  expect(screen.getByText('1001')).toBeTruthy()
  expect(screen.getByText('成功小说')).toBeTruthy()
  expect(screen.getByText('知乎')).toBeTruthy()
  expect(screen.getByText('15')).toBeTruthy()
  expect(screen.getByText('女频')).toBeTruthy()
  expect(screen.getByText('情感')).toBeTruthy()
  expect(screen.getByText('已获取')).toBeTruthy()
  expect(screen.getByText('1002')).toBeTruthy()
  expect(screen.getByText('失败小说')).toBeTruthy()
  expect(screen.getByText('点众')).toBeTruthy()
  expect(screen.getByText('7')).toBeTruthy()
  expect(screen.getByText('男频')).toBeTruthy()
  expect(screen.getByText('悬疑')).toBeTruthy()
  expect(screen.getByText('可重试失败')).toBeTruthy()
  expect(screen.getByText('121 upstream error')).toBeTruthy()

  await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2))
  expect(String(fetchMock.mock.calls[1][0])).toBe('/api/v1/batch-projects/51')
})
