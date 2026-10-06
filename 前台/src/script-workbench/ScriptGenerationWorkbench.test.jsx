// @vitest-environment jsdom

import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ScriptGenerationWorkbench from './ScriptGenerationWorkbench.jsx'

function generationSummary(overrides = {}) {
  return {
    sourceText: '林夏来到医院。',
    editableOutput: '### 分镜一\n00:00-00:03 | 林夏推门',
    compiledPrompt: 'SYSTEM PRESET:\nSYS\n\nSCRIPT:\n### 分镜一\n00:00-00:03 | 林夏推门\n\nHOOK:\nHOOK\n\nDIRECTOR:\nDIRECTOR\n\nPROCESSING RULES:\nRULES',
    latest: {
      SCRIPT: { status: 'completed', outputText: '### 分镜一\n00:00-00:03 | 林夏推门' },
      HOOK: { status: 'completed', outputText: 'HOOK' },
      DIRECTOR: { status: 'completed', outputText: 'DIRECTOR' },
      FINAL_PROMPT: { status: 'completed', outputText: 'SYSTEM PRESET:\nSYS\n\nSCRIPT:\n### 分镜一\n00:00-00:03 | 林夏推门\n\nHOOK:\nHOOK' },
    },
    history: [],
    ...overrides,
  }
}

describe('ScriptGenerationWorkbench', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })

  it('uses the exact output-mode combobox and edits SCRIPT instead of FINAL_PROMPT', async () => {
    const generated = generationSummary({ history: undefined })
    const refreshed = generationSummary({ history: [{ run: { id: 9, status: 'completed' }, editableOutput: generated.editableOutput, latest: generated.latest }] })
    const runBookGeneration = vi.fn(async (_projectId, _bookId, input) => {
      if (input.action === 'extract') {
        return { extraction: {
          characters: [{ id: 'c1', name: '林夏', description: '记者', protagonist: true }],
          scenes: [{ id: 's1', name: '医院走廊', description: '夜间走廊' }],
        } }
      }
      return generated
    })
    const apiClient = {
      getBookGeneration: vi.fn()
        .mockResolvedValueOnce({ sourceText: '林夏来到医院。', latest: {}, history: [] })
        .mockResolvedValueOnce(refreshed),
      runBookGeneration,
    }

    render(<ScriptGenerationWorkbench projectId={3} bookId={11} bookTitle="测试书" apiClient={apiClient} />)
    expect(await screen.findByDisplayValue('林夏来到医院。')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '提取人物与场景' }))
    expect(await screen.findByDisplayValue('林夏')).toBeTruthy()

    fireEvent.click(screen.getByText('爆款开头'))
    fireEvent.mouseDown(screen.getByRole('combobox', { name: '输出模式' }))
    fireEvent.click(await screen.findByRole('option', { name: '分镜模式' }))
    fireEvent.change(screen.getByLabelText('画面前缀'), { target: { value: '写实电影感' } })
    fireEvent.click(screen.getByRole('button', { name: '生成剧本与分镜' }))

    await waitFor(() => expect(runBookGeneration).toHaveBeenCalledTimes(2))
    await waitFor(() => expect(apiClient.getBookGeneration).toHaveBeenCalledTimes(2))
    const generateInput = runBookGeneration.mock.calls[1][2]
    expect(generateInput).toMatchObject({
      workbench: true,
      action: 'generate',
      force: true,
      sourceText: '林夏来到医院。',
      openingMode: 'hook',
      outputMode: 'shotlist',
      hookEnabled: true,
      directorMode: 'normal',
    })
    expect(generateInput.characters[0]).toMatchObject({ name: '林夏', protagonist: true })
    expect(generateInput.constraints.visualPrefix).toBe('写实电影感')
    expect(await screen.findByDisplayValue(/林夏推门/)).toBeTruthy()
    expect(screen.queryByDisplayValue(/SYSTEM PRESET/)).toBeNull()
    expect(screen.getByText(/编辑区只使用 SCRIPT 可编辑成品/)).toBeTruthy()
  })

  it('refreshes server history after generation without closing the workbench', async () => {
    const initial = generationSummary({ editableOutput: '', compiledPrompt: '', latest: {}, history: [] })
    const generated = generationSummary({ history: undefined })
    const refreshed = generationSummary({
      history: [{ run: { id: 27, status: 'completed' }, editableOutput: '### 分镜一\n新版本', latest: generated.latest }],
    })
    const apiClient = {
      getBookGeneration: vi.fn().mockResolvedValueOnce(initial).mockResolvedValueOnce(refreshed),
      runBookGeneration: vi.fn(async () => generated),
    }
    render(<ScriptGenerationWorkbench projectId={3} bookId={11} apiClient={apiClient} />)
    await screen.findByDisplayValue('林夏来到医院。')
    fireEvent.click(screen.getByRole('button', { name: '生成剧本与分镜' }))
    await waitFor(() => expect(apiClient.getBookGeneration).toHaveBeenCalledTimes(2))
    fireEvent.click(screen.getByRole('button', { name: '历史恢复' }))
    expect(await screen.findByText('Run #27 · completed')).toBeTruthy()
  })

  it('supports edit -> undo -> local save -> reopen -> export and labels restore boundaries accurately', async () => {
    const initial = generationSummary({
      sourceText: '原文',
      editableOutput: '初始成品',
      compiledPrompt: 'SYSTEM PRESET:\nSYS\n\nSCRIPT:\n初始成品',
      latest: { SCRIPT: { status: 'completed', outputText: '初始成品' }, FINAL_PROMPT: { status: 'completed', outputText: 'SYSTEM PRESET:\nSYS\n\nSCRIPT:\n初始成品' } },
      history: [{ run: { id: 5, status: 'completed' }, editableOutput: '初始成品', latest: { SCRIPT: { status: 'completed', outputText: '初始成品' } } }],
    })
    const apiClient = { getBookGeneration: vi.fn(async () => initial), runBookGeneration: vi.fn() }
    const onExport = vi.fn()
    const first = render(<ScriptGenerationWorkbench projectId={8} bookId={22} bookTitle="测试书" apiClient={apiClient} onExport={onExport} />)
    expect(await screen.findByDisplayValue('初始成品')).toBeTruthy()

    fireEvent.change(screen.getByLabelText('分镜编辑1'), { target: { value: '编辑版' } })
    expect(screen.getByDisplayValue('编辑版')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '撤销' }))
    expect(screen.getByDisplayValue('初始成品')).toBeTruthy()

    fireEvent.change(screen.getByLabelText('分镜编辑1'), { target: { value: '最终保存版' } })
    fireEvent.click(screen.getByRole('button', { name: '保存当前版本' }))
    await screen.findByText(/完整工作台快照仅保存到本机 localStorage/)
    first.unmount()

    render(<ScriptGenerationWorkbench projectId={8} bookId={22} bookTitle="测试书" apiClient={apiClient} onExport={onExport} />)
    expect(await screen.findByDisplayValue('最终保存版')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '历史恢复' }))
    expect(await screen.findByText(/服务端生成历史当前只恢复可编辑输出/)).toBeTruthy()
    expect(screen.getByText(/手工保存/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    fireEvent.click(screen.getByRole('button', { name: '导出 TXT' }))
    expect(onExport).toHaveBeenCalledWith({ filename: '测试书-canvas.txt', text: '最终保存版' })
  })

  it('marks reference URLs as text-only and does not claim image understanding or dubbing is available', async () => {
    const apiClient = {
      getBookGeneration: vi.fn(async () => ({ sourceText: '原文', latest: {}, history: [] })),
      runBookGeneration: vi.fn(),
    }
    render(<ScriptGenerationWorkbench projectId={3} bookId={11} apiClient={apiClient} />)
    await screen.findByDisplayValue('原文')
    expect(screen.getByText(/参考图 URL 当前仅作为提示词文本元数据/)).toBeTruthy()
    expect(screen.getByText(/配音\/音频生成也未接入本工作台/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '添加人物' }))
    fireEvent.change(screen.getByLabelText('人物名称1'), { target: { value: '林夏' } })
    fireEvent.click(screen.getByRole('button', { name: '共享图片服务未注入' }))
    expect(await screen.findByText(/不会进行图片理解/)).toBeTruthy()
  })
})
