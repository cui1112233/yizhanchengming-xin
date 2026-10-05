// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
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

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(),
  getProjectGeneration: vi.fn(),
  runProjectGeneration: vi.fn(),
  runBookGeneration: vi.fn(),
  retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(),
  getAudioMeasurement: vi.fn(),
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
    api.getAudioMeasurement.mockRejectedValue(new Error('audio_measurement_required'))
  })

  afterEach(() => cleanup())

  it('shows four real stage statuses and retry for failed stage', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))

    expect((await screen.findAllByText('Script')).length).toBeGreaterThan(0)
    expect(screen.getAllByText('Hook').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Director').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Final Prompt').length).toBeGreaterThan(0)
    expect(screen.getByText('导演输出格式错误')).toBeTruthy()
    expect(screen.getByRole('button', { name: '重试 Director' })).toBeTruthy()
  }, 15000)

  it('starts batch generation through the Go API', async () => {
    api.runProjectGeneration.mockResolvedValue({ batchProjectId: 3, completed: 1, failed: 0, books: [] })
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('批量执行')
    fireEvent.click(screen.getByRole('button', { name: '批量执行' }))
    await waitFor(() => expect(api.runProjectGeneration).toHaveBeenCalledTimes(1))
  }, 15000)

  it('disables match audio when no authoritative measurement exists', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('请先生成或检测音频')
    const matchSwitch = screen.getByRole('switch', { name: '匹配音频 Book 11' })
    expect(matchSwitch.disabled).toBe(true)
  }, 15000)

  it('shows measured duration and sends matchAudio only after measurement exists', async () => {
    api.getAudioMeasurement.mockResolvedValue({ bookId: 11, durationMs: 28000, measuredAt: '2026-10-05T08:00:00Z' })
    api.runBookGeneration.mockResolvedValue({ run: { status: 'completed' }, stages: [] })
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('已检测音频：28.00 秒')
    const matchSwitch = screen.getByRole('switch', { name: '匹配音频 Book 11' })
    expect(matchSwitch.disabled).toBe(false)
    fireEvent.click(matchSwitch)
    expect(await screen.findByText('最终分镜总时长将严格匹配音频时长')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '单本执行' }))
    await waitFor(() => expect(api.runBookGeneration).toHaveBeenCalledTimes(1))
    expect(api.runBookGeneration.mock.calls[0][2]).toMatchObject({
      matchAudio: true,
      audioDurationSec: 28,
      shotDurationLimitSec: 15,
    })
  }, 15000)
})