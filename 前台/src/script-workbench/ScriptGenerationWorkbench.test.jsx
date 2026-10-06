import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ScriptGenerationWorkbench from './ScriptGenerationWorkbench.jsx'

describe('ScriptGenerationWorkbench', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })

  it('loads source, extracts editable entities and composes the existing generation pipeline request', async () => {
    const runBookGeneration = vi.fn(async (_projectId, _bookId, input) => {
      if (input.action === 'extract') {
        return {
          extraction: {
            characters: [{ id: 'c1', name: '林夏', description: '记者', protagonist: true }],
            scenes: [{ id: 's1', name: '医院走廊', description: '夜间走廊' }],
          },
        }
      }
      return {
        run: { id: 9, status: 'completed' },
        latest: {
          SCRIPT: { status: 'completed', outputText: 'SCRIPT' },
          HOOK: { status: 'completed', outputText: 'HOOK' },
          DIRECTOR: { status: 'completed', outputText: 'DIRECTOR' },
          FINAL_PROMPT: { status: 'completed', outputText: '### 分镜一\n00:00-00:03 | 林夏推门' },
        },
        history: [],
      }
    })
    const apiClient = {
      getBookGeneration: vi.fn(async () => ({ sourceText: '林夏来到医院。', latest: {}, history: [] })),
      runBookGeneration,
    }

    render(<ScriptGenerationWorkbench projectId={3} bookId={11} bookTitle="测试书" apiClient={apiClient} />)

    expect(await screen.findByDisplayValue('林夏来到医院。')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '提取人物与场景' }))
    expect(await screen.findByDisplayValue('林夏')).toBeInTheDocument()
    expect(screen.getByText('设为主角')).toBeInTheDocument()

    fireEvent.click(screen.getByText('爆款开头'))
    fireEvent.mouseDown(screen.getByLabelText('输出模式'))
    fireEvent.click(await screen.findByText('分镜模式'))
    fireEvent.change(screen.getByLabelText('画面前缀'), { target: { value: '写实电影感' } })
    fireEvent.click(screen.getByRole('button', { name: '生成剧本与分镜' }))

    await waitFor(() => expect(runBookGeneration).toHaveBeenCalledTimes(2))
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
    expect(await screen.findByDisplayValue(/林夏推门/)).toBeInTheDocument()
  })

  it('does not invent a provider configuration when shared image service is not injected', async () => {
    const apiClient = {
      getBookGeneration: vi.fn(async () => ({ sourceText: '原文', latest: {}, history: [] })),
      runBookGeneration: vi.fn(),
    }
    render(<ScriptGenerationWorkbench projectId={3} bookId={11} apiClient={apiClient} />)
    await screen.findByDisplayValue('原文')
    fireEvent.click(screen.getByRole('button', { name: '添加人物' }))
    fireEvent.change(screen.getByLabelText('人物名称1'), { target: { value: '林夏' } })
    fireEvent.click(screen.getByRole('button', { name: '通过共享图片服务补参考图' }))
    expect(await screen.findByText(/共享图片服务接入/)).toBeInTheDocument()
  })
})
