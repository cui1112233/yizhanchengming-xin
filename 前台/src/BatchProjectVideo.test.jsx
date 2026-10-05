// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import BatchProjectListPage from './BatchProjectListPage.jsx'

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: () => ({ matches: false, addListener: () => {}, removeListener: () => {}, addEventListener: () => {}, removeEventListener: () => {}, dispatchEvent: () => false }),
})

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(),
  getBatchProject: vi.fn(),
  getProjectGeneration: vi.fn(),
  runProjectGeneration: vi.fn(),
  runBookGeneration: vi.fn(),
  retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(),
  getProjectVideoStatus: vi.fn(),
  retryVideoTask: vi.fn(),
  cancelVideoTask: vi.fn(),
}))

import * as api from './api.js'

describe('BatchProject VIDEO status UI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 7, name: '视频项目', bookCount: 2, sources: [], genders: [], styles: [], runStatus: 'running' }] })
    api.getProjectGeneration.mockResolvedValue({
      batchProjectId: 7, completed: 0, failed: 0, pending: 2, running: 0,
      books: [
        { bookId: 11, title: '失败视频小说', stages: {} },
        { bookId: 12, title: '执行中视频小说', stages: {} },
      ],
    })
    api.getProjectVideoStatus.mockResolvedValue({
      batchProjectId: 7,
      books: [
        { bookId: 11, provider: 'yfai_seedance', model: 'seedance-2-0-official', status: 'failed', errorMessage: '视频生成失败', attempts: [{ id: 101, attempt: 1, status: 'failed', errorMessage: '视频生成失败' }] },
        { bookId: 12, provider: 'doubao_local_executor', model: 'doubao-seedance', status: 'running', attempts: [{ id: 202, attempt: 2, status: 'running' }] },
      ],
    })
    api.retryVideoTask.mockResolvedValue({})
    api.cancelVideoTask.mockResolvedValue({})
  })

  afterEach(() => cleanup())

  it('shows provider model status attempts errors and VIDEO actions without secrets', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('视频项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))

    expect(await screen.findByText('yfai_seedance')).toBeTruthy()
    expect(screen.getByText('seedance-2-0-official')).toBeTruthy()
    expect(screen.getByText('doubao_local_executor')).toBeTruthy()
    expect(screen.getByText('doubao-seedance')).toBeTruthy()
    expect(screen.getByText('视频生成失败')).toBeTruthy()
    expect(screen.getByText('尝试 1 次')).toBeTruthy()
    expect(screen.getByText('尝试 2 次')).toBeTruthy()
    expect(screen.queryByText(/secret/i)).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: '重试 VIDEO' }))
    await waitFor(() => expect(api.retryVideoTask).toHaveBeenCalledWith(101, expect.any(String)))

    fireEvent.click(screen.getByRole('button', { name: '取消 VIDEO' }))
    await waitFor(() => expect(api.cancelVideoTask).toHaveBeenCalledWith(202))
    await waitFor(() => expect(api.getProjectVideoStatus.mock.calls.length).toBeGreaterThanOrEqual(3))
  }, 15000)
})
