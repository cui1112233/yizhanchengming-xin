// @vitest-environment jsdom

import React from 'react'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
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
  getGenerationRun: vi.fn(),
  runProjectGeneration: vi.fn(),
  runBookGeneration: vi.fn(),
  retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(),
  getAudioMeasurement: vi.fn(),
  getProjectVideoStatus: vi.fn(),
  retryVideoTask: vi.fn(),
  cancelVideoTask: vi.fn(),
  restoreBatchProject: vi.fn(),
}))

import * as api from './api.js'

describe('BatchProjectListPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
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
      project: { id: 3, intakeId: 9, name: '测试项目', bookCount: 2, runStatus: 'running', sources: ['知乎', '点众'], genders: ['女频', '男频'], styles: ['情感', '悬疑'] },
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
          DIRECTOR: { stage: 'DIRECTOR', status: 'failed', bookRunId: 72, errorMessage: '导演输出格式错误' },
          FINAL_PROMPT: { stage: 'FINAL_PROMPT', status: 'pending' },
        },
      }],
    })
    api.getAudioMeasurement.mockRejectedValue(new Error('audio_measurement_required'))
    api.getProjectVideoStatus.mockResolvedValue({ batchProjectId: 3, books: [] })
    api.getGenerationRun.mockResolvedValue({ runId: 44, batchProjectId: 3, status: 'completed', terminal: true, counts: { total: 1, completed: 1 }, tasks: [] })
  })

  afterEach(() => cleanup())

  it('展示直接读取详情的真实 BatchProject 汇总字段', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
    await screen.findByText('测试项目')
    expect(screen.getAllByText('知乎').length).toBeGreaterThan(0)
    expect(screen.getAllByText('点众').length).toBeGreaterThan(0)
    expect(screen.getByText('小说数量：2')).toBeTruthy()
    expect(screen.getAllByText('女频').length).toBeGreaterThan(0)
    expect(screen.getAllByText('男频').length).toBeGreaterThan(0)
    expect(screen.getAllByText('情感').length).toBeGreaterThan(0)
    expect(screen.getAllByText('悬疑').length).toBeGreaterThan(0)
    expect(screen.getByText('执行中')).toBeTruthy()
  })

  it('reads an owned deep link directly without loading the catalog', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
    expect(await screen.findByText('Batch Factory V11 工作台')).toBeTruthy()
    await waitFor(() => expect(api.getBatchProject).toHaveBeenCalledWith(3))
    expect(api.listBatchProjects).not.toHaveBeenCalled()
  })

  it('shows safe denied detail and correlation without loading a fallback catalog', async () => {
    api.getBatchProject.mockRejectedValueOnce(Object.assign(new Error('owner-secret-canary'), { status: 403, requestId: 'req-denied' }))
    render(<BatchProjectListPage initialProjectId={999} />)
    expect(await screen.findByText('项目不可访问')).toBeTruthy()
    expect(screen.getByText(/req-denied/)).toBeTruthy()
    expect(document.body.textContent).not.toContain('owner-secret-canary')
    expect(api.listBatchProjects).not.toHaveBeenCalled()
    expect(screen.queryByRole('button', { name: '进入生产工作台' })).toBeNull()
  })

  it('keeps the seven semantic book columns and hides raw errors', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)

    expect(await screen.findByText('Batch Factory V11 工作台')).toBeTruthy()
    expect(screen.getByText('1001')).toBeTruthy()
    expect(screen.getByText('成功小说')).toBeTruthy()
    expect(screen.getByText('已获取')).toBeTruthy()
    expect(screen.getByText('1002')).toBeTruthy()
    expect(screen.getByText('失败小说')).toBeTruthy()
    expect(screen.getByText('可重试失败')).toBeTruthy()
    expect(document.body.textContent).not.toContain('121 upstream error')
    expect(screen.getAllByRole('columnheader').map((node) => node.textContent)).toEqual(['序号', '小说', '内容', '风格', '男女频', '状态', '操作'])
    await waitFor(() => expect(api.getBatchProject).toHaveBeenCalledWith(3))
  })

  it('from project books opens the persisted production workbench', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '进入生产工作台' }))
    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getAllByText('Script').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Final Prompt').length).toBeGreaterThan(0)
    expect(screen.getByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
  }, 15000)

  it('shows four real stage statuses and retry for failed stage', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
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
    api.runProjectGeneration.mockResolvedValue({ runId: 44, status: 'queued', dispatch: 'queued', taskIds: [81] })
    render(<BatchProjectListPage initialProjectId={3} />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('批量执行')
    fireEvent.click(screen.getByRole('button', { name: '批量执行' }))
    await waitFor(() => expect(api.runProjectGeneration).toHaveBeenCalledTimes(1))
    expect(api.runProjectGeneration.mock.calls[0][2]).toEqual(expect.objectContaining({ idempotencyKey: expect.any(String) }))
    await waitFor(() => expect(api.getGenerationRun).toHaveBeenCalledWith(3, 44, expect.objectContaining({ signal: expect.any(AbortSignal) })))
  }, 15000)

  it.each([
    ['{"valid":false,"repaired":true,"durationMs":28250,"error":"validation-canary","nested":{"error":"nested-canary"}}', true],
    ['{"error":"validation-canary"', false],
    ['{"valid":"validation-canary","repaired":[],"durationMs":"validation-canary"}', false],
  ])('renders only safe stage copy and recognized validation facts (%s)', async (validationResult, hasFacts) => {
    api.getProjectGeneration.mockResolvedValue({ batchProjectId: 3, books: [{ bookId: 11, stages: { DIRECTOR: { stage: 'DIRECTOR', status: 'failed', bookRunId: 72, errorMessage: 'provider-canary-short', errorCode: 'unknown-code', outputText: '已存输出' } } }] })
    api.getGenerationStage.mockResolvedValue({ stage: 'DIRECTOR', status: 'failed', errorMessage: 'provider-canary-short', errorCode: 'GENERATION_TIMELINE_INVALID', validationResult, inputSnapshot: 'snapshot-canary', outputText: '保留正文 provider-canary-short' })
    api.retryGenerationStage.mockResolvedValue({ runId: 44, status: 'queued', dispatch: 'queued', taskIds: [81] })
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '生成状态' }))
    expect(await screen.findByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
    expect(document.body.textContent).not.toContain('provider-canary-short')
    fireEvent.click(screen.getByRole('button', { name: '查看结果' }))
    expect(await screen.findByText('导演分镜时长校验失败，请重试导演阶段')).toBeTruthy()
    expect(screen.getByText('保留正文 provider-canary-short')).toBeTruthy()
    for (const canary of ['validation-canary', 'nested-canary', 'snapshot-canary']) expect(document.body.textContent).not.toContain(canary)
    if (hasFacts) expect(screen.getByText(/校验未通过.*已修复.*28.25 秒/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '重试 Director' }))
    await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 11, 'DIRECTOR', expect.objectContaining({ sourceBookRunId: 72 }), expect.objectContaining({ idempotencyKey: expect.any(String) })))
  }, 15000)

  it('disables match audio when no authoritative measurement exists', async () => {
    render(<BatchProjectListPage initialProjectId={3} />)
    await screen.findByText('测试项目')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await screen.findByText('请先生成或检测音频')
    const matchSwitch = screen.getByRole('switch', { name: '匹配音频 Book 11' })
    expect(matchSwitch.disabled).toBe(true)
  }, 15000)

  it('shows measured duration and sends matchAudio only after measurement exists', async () => {
    api.getAudioMeasurement.mockResolvedValue({ bookId: 11, durationMs: 28000, measuredAt: '2026-10-05T08:00:00Z' })
    api.runBookGeneration.mockResolvedValue({ runId: 44, status: 'queued', dispatch: 'queued', taskIds: [81] })
    render(<BatchProjectListPage initialProjectId={3} />)
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
  it('keeps archived detail readonly and reloads authoritative detail after restore', async () => {
    api.getBatchProject.mockResolvedValueOnce({ project: { id: 3, name: '归档项目', archivedAt: '2026-10-09T04:05:06Z' }, books: [] })
    api.restoreBatchProject.mockResolvedValue({})
    render(<BatchProjectListPage initialProjectId={3} />)
    expect(await screen.findByText('项目已归档，当前为只读查看。恢复后才能修改或执行。')).toBeTruthy()
    for (const name of ['进入生产工作台', '生成状态', '统一设置', '批量执行', '单本执行']) expect(screen.queryByRole('button', { name })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: '恢复项目' }))
    await waitFor(() => expect(api.restoreBatchProject).toHaveBeenCalledWith(3))
    expect(await screen.findByRole('button', { name: '进入生产工作台' })).toBeTruthy()
    expect(screen.queryByText('项目已归档，当前为只读查看。恢复后才能修改或执行。')).toBeNull()
    expect(api.getBatchProject).toHaveBeenCalledTimes(2)
  })
  it('retries failed detail safely without first reading a catalog', async () => {
    api.getBatchProject.mockRejectedValueOnce(Object.assign(new Error('sql-secret-canary'), { requestId: 'req-detail' }))
    render(<BatchProjectListPage initialProjectId={3} />)
    expect(await screen.findByText(/req-detail/)).toBeTruthy()
    expect(document.body.textContent).not.toContain('sql-secret-canary')
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ }))
    await screen.findByText('成功小说')
    fireEvent.click(screen.getByRole('button', { name: '刷新项目' }))
    await waitFor(() => expect(api.getBatchProject).toHaveBeenCalledTimes(3))
    expect(api.listBatchProjects).not.toHaveBeenCalled()
  })
  it('reads archived stage history while hiding all pipeline mutations', async () => {
    api.getBatchProject.mockResolvedValue({ project: { id: 3, name: '归档项目', archivedAt: '2026-10-09T04:05:06Z' }, books: [] })
    api.getAudioMeasurement.mockResolvedValue({ bookId: 11, durationMs: 28000 })
    api.getGenerationStage.mockResolvedValue({ stage: 'SCRIPT', status: 'completed', outputText: '归档历史脚本' })
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '查看生成状态' }))
    expect(await screen.findByText('生成阶段执行失败，请稍后重试')).toBeTruthy()
    for (const name of ['批量执行', '单本执行', '重试 Director', '重试 VIDEO', '取消 VIDEO']) expect(screen.queryByRole('button', { name })).toBeNull()
    expect(screen.getByRole('switch', { name: '匹配音频 Book 11' }).disabled).toBe(true)
    fireEvent.click(screen.getAllByRole('button', { name: '查看结果' })[0])
    expect(await screen.findByText('归档历史脚本')).toBeTruthy()
    expect(api.getGenerationStage).toHaveBeenCalledWith(3, 11, 'SCRIPT')
  }, 15000)

  it.each([
    ['batch', 'runProjectGeneration', '批量执行', 'failed'],
    ['single', 'runBookGeneration', '单本执行', 'failed'],
    ['stage retry', 'retryGenerationStage', '重试 Director', 'failed'],
    ['video retry', 'retryVideoTask', '重试 VIDEO', 'failed'],
    ['video cancel', 'cancelVideoTask', '取消 VIDEO', 'running'],
    ['result read', 'getGenerationStage', '查看结果', 'failed'],
  ])('keeps safe %s errors and correlation visible inside the drawer across status refresh', async (_, method, button, status) => {
    api.getProjectVideoStatus.mockResolvedValue({ batchProjectId: 3, books: [{ bookId: 11, provider: 'test-provider', model: 'test-model', attempts: [{ id: 91, status, errorMessage: status === 'failed' ? 'video-secret' : '' }] }] })
    api[method].mockRejectedValueOnce(Object.assign(new Error('mutation-secret'), { status: 500, requestId: `req-${method}` }))
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '生成状态' }))
    const drawer = await screen.findByRole('dialog')
    await screen.findByText('test-provider')
    fireEvent.click(within(drawer).getAllByRole('button', { name: button })[0])
    expect(await within(drawer).findByText(`请求编号：req-${method}`)).toBeTruthy()
    expect(document.body.textContent).not.toContain('mutation-secret')
    expect(document.body.textContent).not.toContain('video-secret')
    fireEvent.click(screen.getByRole('button', { name: '生成状态' }))
    await waitFor(() => expect(api.getProjectGeneration).toHaveBeenCalledTimes(method === 'runProjectGeneration' || method === 'runBookGeneration' ? 3 : 2))
    expect(within(drawer).getByText(`请求编号：req-${method}`)).toBeTruthy()
  }, 15000)

  it('re-reads archived detail after a write conflict without losing its correlation', async () => {
    api.runProjectGeneration.mockRejectedValueOnce(Object.assign(new Error('archive-secret'), { status: 409, code: 'BATCH_PROJECT_ARCHIVED', requestId: 'req-archive-write' }))
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '生成状态' }))
    await screen.findByRole('button', { name: '单本执行' })
    api.getBatchProject.mockResolvedValue({ project: { id: 3, name: '归档项目', archivedAt: '2026-10-09T04:05:06Z' }, books: [] })
    fireEvent.click(screen.getByRole('button', { name: '批量执行' }))
    expect(await screen.findByText('项目已归档，当前为只读查看。恢复后才能修改或执行。')).toBeTruthy()
    expect(screen.getAllByText(/req-archive-write/).length).toBeGreaterThan(0)
    expect(document.body.textContent).not.toContain('archive-secret')
    fireEvent.click(screen.getByRole('button', { name: '查看生成状态' }))
    expect(await within(screen.getByRole('dialog')).findByText(/req-archive-write/)).toBeTruthy()
    expect(screen.queryByRole('button', { name: '批量执行' })).toBeNull()
  }, 15000)

  it.each(['refresh denied', 'close', 'project switch'])('discards a delayed stage result after %s', async (boundary) => {
    let resolveStage
    api.getGenerationStage.mockReturnValueOnce(new Promise((resolve) => { resolveStage = resolve }))
    const view = render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '生成状态' }))
    fireEvent.click((await screen.findAllByRole('button', { name: '查看结果' }))[0])
    if (boundary === 'refresh denied') {
      api.getBatchProject.mockRejectedValueOnce(Object.assign(new Error('denied-secret'), { status: 403 }))
      fireEvent.click(screen.getByRole('button', { name: '刷新项目' }))
      await screen.findByText('项目不可访问')
    } else if (boundary === 'close') {
      fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Close' }))
    } else {
      api.getBatchProject.mockResolvedValueOnce({ project: { id: 4, name: '新项目' }, books: [] })
      view.rerender(<BatchProjectListPage initialProjectId={4} />)
      await screen.findByText('新项目')
    }
    await act(async () => resolveStage({ stage: 'SCRIPT', outputText: '迟到结果-canary' }))
    expect(document.body.textContent).not.toContain('迟到结果-canary')
    expect(screen.queryByText('SCRIPT · 执行结果')).toBeNull()
  }, 15000)

  it.each(['failed', 'running'])('hides archived video writes with actual %s attempts', async (status) => {
    api.getBatchProject.mockResolvedValue({ project: { id: 3, name: '归档项目', archivedAt: '2026-10-09T04:05:06Z' }, books: [] })
    api.getProjectVideoStatus.mockResolvedValue({ batchProjectId: 3, books: [{ bookId: 11, provider: 'archived-provider', model: 'archived-model', attempts: [{ id: 91, status, errorMessage: status === 'failed' ? 'video-secret' : '' }] }] })
    render(<BatchProjectListPage initialProjectId={3} />)
    fireEvent.click(await screen.findByRole('button', { name: '查看生成状态' }))
    expect(await screen.findByText('archived-provider')).toBeTruthy()
    expect(screen.getByText('尝试 1 次')).toBeTruthy()
    for (const name of ['批量执行', '单本执行', '重试 Director', '重试 VIDEO', '取消 VIDEO']) expect(screen.queryByRole('button', { name })).toBeNull()
    expect(document.body.textContent).not.toContain('video-secret')
  }, 15000)
})
