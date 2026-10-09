// @vitest-environment jsdom
import React from 'react'
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./api.js', () => ({
  createBatchProject: vi.fn(), getAudioMeasurement: vi.fn(), getGenerationRun: vi.fn(), getGenerationStage: vi.fn(), getScriptWorkspace: vi.fn(), listBatchProjects: vi.fn(), retryGenerationStage: vi.fn(), runBookGeneration: vi.fn(), saveProductionSettings: vi.fn(), saveScriptOriginalText: vi.fn(), getScriptStoryboard: vi.fn(), deleteScriptStoryboardCard: vi.fn(), reorderScriptStoryboard: vi.fn(), saveScriptStoryboardCard: vi.fn(),
}))
import * as api from './api.js'
import ScriptWorkspace from './ScriptWorkspace.jsx'

const project = { id: 3, name: '剧本项目', bookCount: 1, runStatus: 'completed' }
const detail = {
  project: { project, books: [{ id: 11, title: '测试小说', originalText: '第一章 原文', status: 'fetched' }] },
  generation: { batchProjectId: 3, books: [{ bookId: 11, stages: { SCRIPT: { status: 'completed', outputText: '剧本结果', promptKey: 'script.default', promptVersion: 2 }, HOOK: { status: 'pending' }, DIRECTOR: { status: 'failed', bookRunId: 72, errorMessage: '导演失败' }, FINAL_PROMPT: { status: 'pending' } } }] },
  settings: { project: { production: { scriptWorkspace: { characters: '甲', scenes: '客厅', constraints: '禁用旁白', mode: 'continuous', duration: 10, directorMode: 'normal', model: '后台模型' } } } },
  video: { books: [] }, prompts: { prompts: [{ key: 'script.default', version: 2 }] },
}
describe('ScriptWorkspace', () => {
  beforeEach(() => { vi.clearAllMocks(); api.getScriptStoryboard.mockResolvedValue({ cards: [] }); api.listBatchProjects.mockResolvedValue({ projects: [project] }); api.getScriptWorkspace.mockResolvedValue(detail); api.getAudioMeasurement.mockRejectedValue(new Error('audio_measurement_required')); api.getGenerationRun.mockResolvedValue({ runId: 44, batchProjectId: 3, status: 'completed', terminal: true }); api.runBookGeneration.mockResolvedValue({ runId: 44, status: 'queued' }); api.retryGenerationStage.mockResolvedValue({ runId: 45, status: 'queued' }); api.saveProductionSettings.mockResolvedValue({}) })
  afterEach(cleanup)
  it('reads persisted project, original text and StageRun results in the legacy workbench structure', async () => { render(<ScriptWorkspace />); expect(await screen.findByText('剧本项目')).toBeTruthy(); expect(await screen.findByDisplayValue('第一章 原文')).toBeTruthy(); expect(screen.getByText('剧本')).toBeTruthy(); expect(screen.getByText('导演分镜')).toBeTruthy(); expect(document.querySelector('.script-workbench-form .utility-workbench .script-chat-shell')).not.toBeNull() })
  it('works with the public prompt projection containing only key and version', async () => { detail.prompts = { prompts: [{ key: 'script.default', version: 2, enabled: true }] }; render(<ScriptWorkspace />); expect(await screen.findByText('剧本项目')).toBeTruthy(); expect(screen.queryByText('published prompt body')).toBeNull() })
  it('saves characters, scenes and constraints through the existing MySQL-backed settings API', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.change(screen.getByLabelText('人物'), { target: { value: '甲（青年）' } }); fireEvent.click(screen.getByRole('button', { name: '保存到服务端' })); await waitFor(() => expect(api.saveProductionSettings).toHaveBeenCalledWith(3, expect.objectContaining({ scriptWorkspace: expect.objectContaining({ characters: '甲（青年）', scenes: '客厅' }) }))) })
  it('keeps TXT/original-text edits in the server-backed book record', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.change(screen.getByLabelText('小说原文'), { target: { value: '手工修订后的原文' } }); fireEvent.click(screen.getByRole('button', { name: '保存原文' })); await waitFor(() => expect(api.saveScriptOriginalText).toHaveBeenCalledWith(3, 11, '手工修订后的原文')) })
  it('runs the existing four-stage generation pipeline and retries only a failed stage', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.click(screen.getByRole('button', { name: '按已保存原文生成剧本' })); await waitFor(() => expect(api.runBookGeneration).toHaveBeenCalledWith(3, 11, expect.objectContaining({ directorMode: 'normal', shotDurationLimitSec: 10 }), expect.objectContaining({ idempotencyKey: expect.any(String) }))); await waitFor(() => expect(api.getGenerationRun).toHaveBeenCalledWith(3, 44, expect.objectContaining({ signal: expect.any(AbortSignal) }))); fireEvent.click(screen.getByRole('button', { name: '重试' })); await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 11, 'DIRECTOR', expect.objectContaining({ sourceBookRunId: 72 }), expect.objectContaining({ idempotencyKey: expect.any(String) }))) })
  it('keeps the newly selected project when an older workspace refresh resolves late', async () => {
    let resolveOldWorkspace
    const projectB = { id: 4, name: '剧本项目 B', bookCount: 1, runStatus: 'running' }
    const detailB = {
      ...detail,
      project: { project: projectB, books: [{ id: 12, title: 'B 小说', originalText: 'B 原文', status: 'fetched' }] },
      generation: { batchProjectId: 4, books: [{ bookId: 12, stages: {} }] },
      video: { batchProjectId: 4, books: [] },
    }
    api.listBatchProjects.mockResolvedValue({ projects: [project, projectB] })
    api.getScriptWorkspace
      .mockResolvedValueOnce(detail)
      .mockImplementationOnce(() => new Promise(resolve => { resolveOldWorkspace = resolve }))
      .mockResolvedValueOnce(detailB)

    render(<ScriptWorkspace />)
    await screen.findByDisplayValue('第一章 原文')
    fireEvent.click(screen.getByRole('button', { name: '按已保存原文生成剧本' }))
    await waitFor(() => expect(api.getScriptWorkspace).toHaveBeenCalledTimes(2))
    fireEvent.click(screen.getByText('剧本项目 B'))
    expect(await screen.findByDisplayValue('B 原文')).toBeTruthy()
    await act(async () => {
      resolveOldWorkspace(detail)
      await Promise.resolve()
    })
    expect(screen.queryByDisplayValue('第一章 原文')).toBeNull()
    expect(screen.getByDisplayValue('B 原文')).toBeTruthy()
  }, 15000)
  it.each([
    [{ errorCode: 'GENERATION_UNAVAILABLE', errorMessage: 'provider-canary-short' }, '生成服务暂不可用，请稍后重试'],
    [{ errorCode: 'legacy-untrusted', errorMessage: 'provider-canary-short password=pass-canary' }, '生成阶段执行失败，请稍后重试'],
    [{ errorCode: { toString: null }, errorMessage: 'provider-canary-short' }, '生成阶段执行失败，请稍后重试'],
    [{ errorMessage: '生成参数无效，请检查后重试' }, '生成参数无效，请检查后重试'],
    [{ status: 'failed' }, '生成阶段执行失败，请稍后重试'],
    [{ errorMessage: 'provider-canary-short', outputText: '保留正文 provider-canary-short' }, '保留正文 provider-canary-short'],
  ])('projects safe modal feedback and preserves output priority (%j)', async (outcome, message) => {
    api.getGenerationStage.mockResolvedValue({ stage: 'SCRIPT', requestId: 'execution-old', ...outcome })
    render(<ScriptWorkspace />)
    fireEvent.click(await screen.findByRole('button', { name: '查看' }))
    expect(await screen.findByText(message)).toBeTruthy()
    const modal = screen.getByRole('dialog')
    if (!outcome.outputText) expect(modal.textContent).not.toContain('provider-canary-short')
    expect(modal.textContent).not.toContain('pass-canary')
    expect(screen.getByRole('button', { name: '重试' })).toBeTruthy()
  })
})
