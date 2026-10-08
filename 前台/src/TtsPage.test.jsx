// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TtsPage from './TtsPage.jsx'

vi.mock('./api.js', () => ({
  listBatchProjects: vi.fn(), getBatchProject: vi.fn(), listShuihuoMediaTasks: vi.fn(), listShuihuoAssets: vi.fn(), listShuihuoCandidates: vi.fn(), listShuihuoSegments: vi.fn(), createShuihuoMediaTask: vi.fn(), retryShuihuoMediaTask: vi.fn(), shuihuoAssetContentURL: vi.fn(() => '/api/v1/audio/content'),
}))
import * as api from './api.js'

describe('TtsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 3, name: '声音项目' }] })
    api.getBatchProject.mockResolvedValue({ books: [{ id: 7, bookId: 'external-7', title: '配音小说' }] })
    api.listShuihuoMediaTasks.mockResolvedValue([{ id: 18, kind: 'audio', status: 'failed', errorMessage: 'provider timeout', provider: 'tts', model: 'voice-1' }])
    api.listShuihuoAssets.mockResolvedValue([])
    api.listShuihuoCandidates.mockResolvedValue([])
    api.listShuihuoSegments.mockResolvedValue([{ id: 51, text: '分镜配音文本' }])
  })

  it('只从已保存分镜创建 TTS 任务，不把 Provider 密钥或模型输入交给浏览器', async () => {
    api.createShuihuoMediaTask.mockResolvedValue({ id: 19, status: 'queued' })
    render(<TtsPage />)
    fireEvent.click(await screen.findByRole('button', { name: '创建配音任务' }))
    fireEvent.mouseDown(screen.getByRole('combobox', { name: '选择配音小说' }))
    fireEvent.click((await screen.findAllByText('配音小说')).at(-1))
    fireEvent.mouseDown(screen.getByRole('combobox', { name: '选择配音分镜' }))
    fireEvent.click(await screen.findByText(/分镜 1/))
    fireEvent.click(screen.getByRole('button', { name: /提\s*交/ }))
    await waitFor(() => expect(api.createShuihuoMediaTask).toHaveBeenCalledWith(3, 7, expect.objectContaining({ kind: 'audio', segmentId: 51, requestId: expect.any(String) })))
    expect(api.createShuihuoMediaTask.mock.calls[0][2]).not.toHaveProperty('apiKey')
    expect(api.createShuihuoMediaTask.mock.calls[0][2]).not.toHaveProperty('model')
  })

  it('从已有 Shuihuo 音频任务读取失败状态并安全重试', async () => {
    api.retryShuihuoMediaTask.mockResolvedValue({ id: 18, status: 'queued' })
    render(<TtsPage />)
    expect(await screen.findByText('配音小说')).toBeTruthy()
    expect(screen.getByText('provider timeout')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ }))
    await waitFor(() => expect(api.retryShuihuoMediaTask).toHaveBeenCalledWith(3, 7, 18))
  })

  it('已选候选资产使用受权限保护的内容地址试听和下载', async () => {
    api.listShuihuoMediaTasks.mockResolvedValue([{ id: 18, kind: 'audio', status: 'succeeded', provider: 'tts' }])
    api.listShuihuoAssets.mockResolvedValue([{ id: 90, type: 'audio', status: 'ready' }])
    api.listShuihuoCandidates.mockResolvedValue([{ id: 2, assetId: 90, selected: true }])
    render(<TtsPage />)
    expect(await screen.findByRole('link', { name: /下\s*载\s*音\s*频/ })).toBeTruthy()
    expect(api.shuihuoAssetContentURL).toHaveBeenCalledWith(3, 7, 90)
  })

  it('executor_unavailable never presents an audio player or download as a completed result', async () => {
    api.listShuihuoMediaTasks.mockResolvedValue([{ id: 21, kind: 'audio', status: 'executor_unavailable', errorCode: 'executor_unavailable', errorMessage: 'TTS Provider 未配置' }])
    render(<TtsPage />)
    expect(await screen.findByText('TTS Provider 未配置')).toBeTruthy()
    expect(screen.getByText('TTS 执行器不可用，尚未生成音频。')).toBeTruthy()
    expect(screen.queryByRole('audio')).toBeNull()
    expect(screen.queryByRole('link', { name: /下载音频/ })).toBeNull()
  })
})
