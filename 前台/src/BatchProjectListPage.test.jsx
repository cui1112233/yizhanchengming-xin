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
  getBatchProject: vi.fn(),
  getProjectGeneration: vi.fn(),
  runProjectGeneration: vi.fn(),
  runBookGeneration: vi.fn(),
  retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(),
  getAudioMeasurement: vi.fn(),
  getProjectVideoStatus: vi.fn(),
  retryVideoTask: vi.fn(),
  cancelVideoTask: vi.fn(),
}))

import * as api from './api.js'

describe('BatchProjectListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({
      projects: [{
        id: 3,
        name: '测试项目',
        sources: ['知乎', '点众'],
        bookCount: 2,
        genders: ['女频', '男频'],
        styles: ['情感', '悬疑'],
        runStatus: 'running',
      }],
    })
    api.getBatchProject.mockResolvedValue({
      project: { id: 3, intakeId: 9, name: '测试项目' },
      books: [
        { id: 31, bookId: '1001', title: '成功小说', source: '知乎', platformId: '15', gender: '女频', style: '情感', status: 'fetched', errorMessage: '' },
        { id: 32, bookId: '1002', title: '失败小说', source: '点众', platformId: '7', gender: '男频', style: '悬疑', status: 'retryable_failed', errorMessage: '121 upstream error' },
      ],
    })
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
    api.getProjectVideoStatus.mockResolvedValue({ batchProjectId: 3, books: [] })
  })

  afterEach(() => cleanup())

  it('展示真实 BatchProject 汇总字段', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    expect(screen.getByText('知乎')).toBeTruthy()
    expect(screen.getByText('点众')).toBeTruthy()
    expect(screen.getByText('2')).toBeTruthy()
    expect(screen.getByText('女频')).toBeTruthy()
    expect(screen.getByText('男频')).toBeTruthy()
    expect(screen.getByText('情感')).toBeTruthy()
    expect(screen.getByText('悬疑')).toBeTruthy()
    expect(screen.getByText('执行中')).toBeTruthy()
  })

  it('opens an initial project only after it appears in the server-visible list', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
    expect(await screen.findByText('Batch Factory V11 工作台')).toBeTruthy()
    await waitFor(() => expect(api.getBatchProject).toHaveBeenCalledWith(3))
  })

  it('keeps the visible list and does not read an unowned initial project', async () => {
    render(<BatchProjectListPage initialProjectId={999} />)
    expect(await screen.findByText('项目不可访问')).toBeTruthy()
    expect(screen.getByText('测试项目')).toBeTruthy()
    expect(api.getBatchProject).not.toHaveBeenCalled()
  })

  it('点击真实项目进入 V11 工作台并展示全部小说状态与错误', async () => {
    render(<BatchProjectListPage />)
    fireEvent.click(await screen.findByRole('button', { name: '测试项目' }))

    expect(await screen.findByText('Batch Factory V11 工作台')).toBeTruthy()
    expect(screen.getByText('1001')).toBeTruthy()
    expect(screen.getByText('成功小说')).toBeTruthy()
    expect(screen.getByText('15')).toBeTruthy()
    expect(screen.getByText('已获取')).toBeTruthy()
    expect(screen.getByText('1002')).toBeTruthy()
    expect(screen.getByText('失败小说')).toBeTruthy()
    expect(screen.getByText('7')).toBeTruthy()
    expect(screen.getByText('可重试失败')).toBeTruthy()
    expect(screen.getByText('121 upstream error')).toBeTruthy()
    await waitFor(() => expect(api.getBatchProject).toHaveBeenCalledWith(3))
  })

  it('from project books opens the persisted production workbench', async () => {
    render(<BatchProjectListPage />)
    fireEvent.click(await screen.findByRole('button', { name: '测试项目' }))
    fireEvent.click(await screen.findByRole('button', { name: '进入生产工作台' }))
    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getAllByText('Script').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Final Prompt').length).toBeGreaterThan(0)
    expect(screen.getByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
  }, 15000)

  it('shows four real stage statuses and retry for failed stage', async () => {
    render(<BatchProjectListPage />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))

    expect((await screen.findAllByText('Script')).length).toBeGreaterThan(0)
    expect(screen.getAllByText('Hook').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Director').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Final Prompt').length).toBeGreaterThan(0)
    expect(screen.getByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
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

  it.each([
    ['{"valid":false,"repaired":true,"durationMs":28250,"error":"validation-canary","nested":{"error":"nested-canary"}}', true],
    ['{"error":"validation-canary"', false],
    ['{"valid":"validation-canary","repaired":[],"durationMs":"validation-canary"}', false],
  ])('renders only safe stage copy and recognized validation facts (%s)', async (validationResult, hasFacts) => {
    api.getProjectGeneration.mockResolvedValue({ batchProjectId: 3, books: [{ bookId: 11, stages: { DIRECTOR: { stage: 'DIRECTOR', status: 'failed', errorMessage: 'provider-canary-short', errorCode: 'unknown-code', outputText: '已存输出' } } }] })
    api.getGenerationStage.mockResolvedValue({ stage: 'DIRECTOR', status: 'failed', errorMessage: 'provider-canary-short', errorCode: 'GENERATION_TIMELINE_INVALID', validationResult, inputSnapshot: 'snapshot-canary', outputText: '保留正文 provider-canary-short' })
    api.retryGenerationStage.mockResolvedValue({})
    render(<BatchProjectListPage />)
    fireEvent.click(await screen.findByRole('button', { name: '生成状态' }))
    expect(await screen.findByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
    expect(document.body.textContent).not.toContain('provider-canary-short')
    fireEvent.click(screen.getByRole('button', { name: '查看结果' }))
    expect(await screen.findByText('导演分镜时长校验失败，请重试导演阶段')).toBeTruthy()
    expect(screen.getByText('保留正文 provider-canary-short')).toBeTruthy()
    for (const canary of ['validation-canary', 'nested-canary', 'snapshot-canary']) expect(document.body.textContent).not.toContain(canary)
    if (hasFacts) expect(screen.getByText(/校验未通过.*已修复.*28.25 秒/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '重试 Director' }))
    await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 11, 'DIRECTOR', expect.any(String)))
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
