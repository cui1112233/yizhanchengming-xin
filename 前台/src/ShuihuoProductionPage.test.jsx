// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ShuihuoProductionPage from './ShuihuoProductionPage.jsx'

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(), getProjectGeneration: vi.fn(), getProjectVideoStatus: vi.fn(),
  runProjectGeneration: vi.fn(), runBookGeneration: vi.fn(), retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(), startVideoTask: vi.fn(), retryVideoTask: vi.fn(), cancelVideoTask: vi.fn(),
  listShuihuoSegments: vi.fn(), updateShuihuoSegment: vi.fn(), reorderShuihuoSegments: vi.fn(), listShuihuoAssets: vi.fn(), listShuihuoMediaTasks: vi.fn(), listShuihuoCandidates: vi.fn(), selectShuihuoCandidate: vi.fn(), retryShuihuoMediaTask: vi.fn(),
  getUnifiedSettings: vi.fn(), saveProductionSettings: vi.fn(), savePublishingSettings: vi.fn(),
}))
import * as api from './api.js'

Object.defineProperty(window, 'matchMedia', { writable: true, value: () => ({ matches: false, addListener() {}, removeListener() {}, addEventListener() {}, removeEventListener() {} }) })

const generation = { completed: 1, running: 0, failed: 1, pending: 0, books: [{ bookId: 7, title: '测试小说', stages: { SCRIPT: { status: 'completed', outputText: '剧本正文' }, HOOK: { status: 'completed' }, DIRECTOR: { status: 'failed', errorMessage: '导演失败' }, FINAL_PROMPT: { status: 'completed' } } }] }

describe('ShuihuoProductionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 3, name: '生产项目', bookCount: 1, runStatus: 'running' }] })
    api.getProjectGeneration.mockResolvedValue(generation)
    api.getProjectVideoStatus.mockResolvedValue({ books: [] })
    api.listShuihuoSegments.mockResolvedValue([{ id: 51, position: 0, text: '第一段分镜', version: 1, editRevision: 'r1' }, { id: 52, position: 1, text: '第二段分镜', version: 1, editRevision: 'r1' }])
    api.listShuihuoAssets.mockResolvedValue([{ id: 61, type: 'reference_image', status: 'ready', objectKey: 'reference/a.png' }])
    api.listShuihuoMediaTasks.mockResolvedValue([{ id: 71, kind: 'image', status: 'retryable_failed', errorMessage: '执行器离线' }])
  })
  afterEach(cleanup)

  it('从真实项目、生成和视频 API 读取状态并显示失败恢复入口', async () => {
    render(<ShuihuoProductionPage />)
    expect(await screen.findByText('生产项目 · 生产控制台')).toBeTruthy()
    expect(screen.getAllByText('测试小说').length).toBeGreaterThan(0)
    expect(screen.getByText('导演失败')).toBeTruthy()
    expect(screen.getByRole('button', { name: /重\s*试/ })).toBeTruthy()
    expect(screen.getByRole('button', { name: '提交视频' })).toBeTruthy()
    await waitFor(() => expect(api.getProjectGeneration).toHaveBeenCalledWith(3))
    expect(api.getProjectVideoStatus).toHaveBeenCalledWith(3)
  })

  it('单本继续执行与阶段重试复用现有 Go API', async () => {
    api.runBookGeneration.mockResolvedValue({})
    api.retryGenerationStage.mockResolvedValue({})
    render(<ShuihuoProductionPage />)
    await screen.findByText('测试小说')
    fireEvent.click(screen.getByRole('button', { name: '继续执行' }))
    await waitFor(() => expect(api.runBookGeneration).toHaveBeenCalledWith(3, 7, expect.objectContaining({ hookEnabled: true, directorMode: 'normal' })), { timeout: 1500 })
    fireEvent.click(screen.getAllByRole('button', { name: /重\s*试/ })[0])
    await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 7, 'DIRECTOR', expect.any(String)), { timeout: 1500 })
  })

  it('仅在 FINAL_PROMPT 完成后允许提交视频，并调用原有视频 API', async () => {
    api.startVideoTask.mockResolvedValue({})
    render(<ShuihuoProductionPage />)
    fireEvent.click(await screen.findByRole('button', { name: '提交视频' }))
    fireEvent.change(screen.getByLabelText('提供方'), { target: { value: 'yfai' } })
    fireEvent.change(screen.getByLabelText('模型'), { target: { value: 'seedance-1.0-pro' } })
    fireEvent.click(screen.getAllByRole('button', { name: /提\s*交/ }).at(-1))
    await waitFor(() => expect(api.startVideoTask).toHaveBeenCalledWith(3, 7, expect.objectContaining({ provider: 'yfai', model: 'seedance-1.0-pro' })))
  })

  it('通过 Task 12B API 编辑、排序、选择候选和安全重试媒体任务', async () => {
    api.updateShuihuoSegment.mockResolvedValue({})
    api.reorderShuihuoSegments.mockResolvedValue([])
    api.listShuihuoCandidates.mockResolvedValue([{ id: 81, position: 0, assetId: 61, selected: false }])
    api.selectShuihuoCandidate.mockResolvedValue({})
    api.retryShuihuoMediaTask.mockResolvedValue({})
    render(<ShuihuoProductionPage />)
    const input = await screen.findByLabelText('分镜 1 文本')
    fireEvent.change(input, { target: { value: '已修改分镜' } })
    fireEvent.click(screen.getAllByRole('button', { name: /保\s*存/ })[0])
    await waitFor(() => expect(api.updateShuihuoSegment).toHaveBeenCalledWith(3, 7, 51, expect.objectContaining({ text: '已修改分镜', version: 1 })))
    fireEvent.click(screen.getAllByRole('button', { name: /下\s*移/ })[0])
    await waitFor(() => expect(api.reorderShuihuoSegments).toHaveBeenCalledWith(3, 7, [52, 51]))
    fireEvent.click(screen.getByRole('button', { name: '候选结果' }))
    await waitFor(() => expect(api.listShuihuoCandidates).toHaveBeenCalledWith(3, 7, 71))
    fireEvent.click(await screen.findByRole('button', { name: '设为主候选' }))
    await waitFor(() => expect(api.selectShuihuoCandidate).toHaveBeenCalledWith(3, 7, 71, 81))
    fireEvent.click(screen.getByRole('button', { name: '安全重试' }))
    await waitFor(() => expect(api.retryShuihuoMediaTask).toHaveBeenCalledWith(3, 7, 71))
  }, 15000)

  it('409 编辑冲突不会静默覆盖，并提供刷新恢复', async () => {
    api.updateShuihuoSegment.mockRejectedValue({ status: 409, code: 'version_conflict' })
    render(<ShuihuoProductionPage />)
    fireEvent.change(await screen.findByLabelText('分镜 1 文本'), { target: { value: '冲突文本' } })
    fireEvent.click(screen.getAllByRole('button', { name: /保\s*存/ })[0])
    expect(await screen.findByText(/分镜已被其他操作更新/)).toBeTruthy()
    expect(screen.getByRole('button', { name: '刷新分镜' })).toBeTruthy()
  }, 15000)
})
