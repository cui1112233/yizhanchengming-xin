// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./api.js', () => ({
  createBatchProject: vi.fn(), getAudioMeasurement: vi.fn(), getGenerationStage: vi.fn(), getScriptWorkspace: vi.fn(), listBatchProjects: vi.fn(), retryGenerationStage: vi.fn(), runBookGeneration: vi.fn(), saveProductionSettings: vi.fn(), saveScriptOriginalText: vi.fn(),
}))
import * as api from './api.js'
import ScriptWorkspace from './ScriptWorkspace.jsx'

const project = { id: 3, name: '剧本项目', bookCount: 1, runStatus: 'completed' }
const detail = {
  project: { project, books: [{ id: 11, title: '测试小说', originalText: '第一章 原文', status: 'fetched' }] },
  generation: { books: [{ bookId: 11, stages: { SCRIPT: { status: 'completed', outputText: '剧本结果', promptKey: 'script.default', promptVersion: 2 }, HOOK: { status: 'pending' }, DIRECTOR: { status: 'failed', errorMessage: '导演失败' }, FINAL_PROMPT: { status: 'pending' } } }] },
  settings: { project: { production: { scriptWorkspace: { characters: '甲', scenes: '客厅', constraints: '禁用旁白', mode: 'continuous', duration: 10, directorMode: 'normal', model: '后台模型' } } } },
  video: { books: [] }, prompts: { prompts: [{ key: 'script.default', version: 2 }] },
}
describe('ScriptWorkspace', () => {
  beforeEach(() => { vi.clearAllMocks(); api.listBatchProjects.mockResolvedValue({ projects: [project] }); api.getScriptWorkspace.mockResolvedValue(detail); api.getAudioMeasurement.mockRejectedValue(new Error('audio_measurement_required')); api.runBookGeneration.mockResolvedValue({}); api.saveProductionSettings.mockResolvedValue({}) })
  afterEach(cleanup)
  it('reads persisted project, original text and StageRun results from server', async () => { render(<ScriptWorkspace />); expect(await screen.findByText('剧本项目')).toBeTruthy(); expect(await screen.findByDisplayValue('第一章 原文')).toBeTruthy(); expect(screen.getByText('剧本')).toBeTruthy(); expect(screen.getByText('导演分镜')).toBeTruthy() })
  it('saves characters, scenes and constraints through the existing MySQL-backed settings API', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.change(screen.getByLabelText('人物'), { target: { value: '甲（青年）' } }); fireEvent.click(screen.getByRole('button', { name: '保存到服务端' })); await waitFor(() => expect(api.saveProductionSettings).toHaveBeenCalledWith(3, expect.objectContaining({ scriptWorkspace: expect.objectContaining({ characters: '甲（青年）', scenes: '客厅' }) }))) })
  it('keeps TXT/original-text edits in the server-backed book record', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.change(screen.getByLabelText('小说原文'), { target: { value: '手工修订后的原文' } }); fireEvent.click(screen.getByRole('button', { name: '保存原文' })); await waitFor(() => expect(api.saveScriptOriginalText).toHaveBeenCalledWith(3, 11, '手工修订后的原文')) })
  it('runs the existing four-stage generation pipeline and retries only a failed stage', async () => { render(<ScriptWorkspace />); await screen.findByText('剧本项目'); fireEvent.click(screen.getByRole('button', { name: '生成剧本' })); await waitFor(() => expect(api.runBookGeneration).toHaveBeenCalledWith(3, 11, expect.objectContaining({ directorMode: 'normal', shotDurationLimitSec: 10 }))); fireEvent.click(screen.getByRole('button', { name: '重试' })); await waitFor(() => expect(api.retryGenerationStage).toHaveBeenCalledWith(3, 11, 'DIRECTOR', expect.any(String))) })
})
