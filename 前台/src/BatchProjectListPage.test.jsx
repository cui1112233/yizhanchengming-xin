import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import BatchProjectListPage from './BatchProjectListPage.jsx'

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(),
  getProjectGeneration: vi.fn(),
  runProjectGeneration: vi.fn(),
  runBookGeneration: vi.fn(),
  retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(),
}))

import * as api from './api.js'

describe('BatchProjectListPage generation workbench', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 3, name: '测试项目' }] })
    api.getProjectGeneration.mockResolvedValue({
      batchProjectId: 3,
      completed: 0,
      failed: 1,
      pending: 0,
      running: 0,
      books: [{
        bookId: 11,
        stages: {
          SCRIPT: { stage: 'SCRIPT', status: 'completed', outputText: 'script' },
          HOOK: { stage: 'HOOK', status: 'completed', outputText: 'hook' },
          DIRECTOR: { stage: 'DIRECTOR', status: 'failed', errorMessage: '导演输出格式错误' },
          FINAL_PROMPT: { stage: 'FINAL_PROMPT', status: 'pending' },
        },
      }],
    })
  })

  it('shows four real stage statuses and retry for failed stage', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))

    await screen.findByText('Script')
    expect(screen.getByText('Hook')).toBeTruthy()
    expect(screen.getByText('Director')).toBeTruthy()
    expect(screen.getByText('Final Prompt')).toBeTruthy()
    expect(screen.getByText('导演输出格式错误')).toBeTruthy()
    expect(screen.getByRole('button', { name: '重试 Director' })).toBeTruthy()
  })

  it('starts batch generation through the Go API', async () => {
    api.runProjectGeneration.mockResolvedValue({ batchProjectId: 3, completed: 1, failed: 0, books: [] })
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('批量执行')
    fireEvent.click(screen.getByRole('button', { name: '批量执行' }))
    await waitFor(() => expect(api.runProjectGeneration).toHaveBeenCalledTimes(1))
  })
})
