// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ShuihuoProductionPage from './ShuihuoProductionPage.jsx'

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(), getProjectGeneration: vi.fn(), getProjectVideoStatus: vi.fn(),
  runProjectGeneration: vi.fn(), runBookGeneration: vi.fn(), retryGenerationStage: vi.fn(),
  getGenerationStage: vi.fn(), startVideoTask: vi.fn(), retryVideoTask: vi.fn(), cancelVideoTask: vi.fn(),
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
  })
  afterEach(cleanup)

  it('从真实项目、生成和视频 API 读取状态并显示失败恢复入口', async () => {
    render(<ShuihuoProductionPage />)
    expect(await screen.findByText('生产项目 · 生产控制台')).toBeTruthy()
    expect(screen.getByText('测试小说')).toBeTruthy()
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
    await waitFor(() => expect(api.runBookGeneration).toHaveBeenCalledWith(3, 7, expect.objectContaining({ hookEnabled: true, directorMode: 'normal' })))
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ }))
    await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 7, 'DIRECTOR', expect.any(String)))
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
})
