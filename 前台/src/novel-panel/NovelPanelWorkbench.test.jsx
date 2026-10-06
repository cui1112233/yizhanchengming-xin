// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import NovelPanelWorkbench from './NovelPanelWorkbench.jsx'

beforeAll(() => Object.defineProperty(window, 'matchMedia', { writable: true, value: vi.fn(() => ({ matches: false, addListener: vi.fn(), removeListener: vi.fn(), addEventListener: vi.fn(), removeEventListener: vi.fn() })) }))
afterEach(cleanup)

const workspace = () => ({
  projectId: 9, revision: 2, mode: 'normal', originalText: '第一段原文。\n第二段原文。',
  forcedRoster: '沈清月（青年）,顾川（青年）',
  characters: [{ id: 'a', displayName: '沈清月（青年）', baseName: '沈清月', gender: '女', appearance: '长发' }, { id: 'b', displayName: '顾川（青年）', baseName: '顾川', gender: '男', appearance: '短发' }],
  relationships: [], contentType: '小说推文', unifiedStyle: '现代都市写实', density: 'standard',
  caseLearning: { enabled: true, material: '案例学习' }, instructions: { generationRules: '不丢剧情', mustCoverDetails: '关键冲突', shotRhythmRequirements: '前三秒钩子', negativeInstructions: '' },
  shots: [{ id: 's1', sourceIndex: 1, sourceBasis: '第一段原文。', visual: '宴会中景' }, { id: 's2', sourceIndex: 2, sourceBasis: '第二段原文。', visual: '人物近景' }],
})

function api() {
  return {
    saveWorkspace: vi.fn(async (_id, payload) => ({ workspace: { ...payload.workspace, revision: payload.expectedRevision + 1 } })),
    listHistory: vi.fn(async () => [{ id: 'h1', revision: 1, note: '初版', summary: { sourcePreview: '历史原文' } }]),
    restoreHistory: vi.fn(async (_id, payload) => ({ workspace: { ...workspace(), revision: payload.expectedRevision + 1, originalText: '历史恢复原文。' } })),
    requestStoryboard: vi.fn(async () => ({ shots: [{ id: 'shared', sourceIndex: 1, sourceBasis: '第一段原文。', visual: '共享 Director 画面' }] })),
  }
}

describe('NovelPanelWorkbench', () => {
  it('exposes requested panel scope without provider secrets', () => {
    render(<NovelPanelWorkbench projectId={9} initialWorkspace={workspace()} />)
    expect(screen.getByText('小说面板')).toBeTruthy()
    expect(screen.getByRole('radio', { name: '普通模式' }).checked).toBe(true)
    expect(screen.getByRole('radio', { name: '精品带图模式' })).toBeTruthy()
    expect(screen.getByRole('textbox', { name: '人物强制名单' }).value).toContain('沈清月')
    expect(screen.getByText('人物关系')).toBeTruthy()
    expect(screen.getByText('密度、案例学习与指令设置')).toBeTruthy()
    expect(screen.queryByText(/API Key/i)).toBeNull()
  })

  it('saves revisioned workspace and uses shared Director and asset callbacks', async () => {
    const client = api(); const openAsset = vi.fn()
    render(<NovelPanelWorkbench projectId={9} initialWorkspace={workspace()} api={client} sharedServices={{ openAsset }} />)
    fireEvent.click(screen.getByRole('radio', { name: '精品带图模式' }))
    fireEvent.change(screen.getByRole('textbox', { name: '统一风格' }), { target: { value: '高质感都市电影' } })
    fireEvent.click(screen.getByRole('button', { name: '保存小说面板' }))
    await waitFor(() => expect(client.saveWorkspace).toHaveBeenCalledWith(9, expect.objectContaining({ expectedRevision: 2, workspace: expect.objectContaining({ mode: 'premium_illustrated', unifiedStyle: '高质感都市电影' }) })))
    expect(await screen.findByText(/修订 3/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '请求共享 Director 分镜' }))
    await waitFor(() => expect(client.requestStoryboard).toHaveBeenCalled())
    expect(await screen.findByDisplayValue('共享 Director 画面')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '人物资产' }))
    expect(openAsset).toHaveBeenCalledWith('characters')
  }, 15000)

  it('restores history and keeps unsaved editor state on revision conflict', async () => {
    const client = api()
    render(<NovelPanelWorkbench projectId={9} initialWorkspace={workspace()} api={client} />)
    fireEvent.click(screen.getByRole('button', { name: '保存记录与恢复' }))
    expect(await screen.findByText(/修订 1 · 初版/)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /^恢\s*复$/ }))
    await waitFor(() => expect(screen.getByRole('textbox', { name: '整段小说原文' }).value).toBe('历史恢复原文。'))
    client.saveWorkspace.mockRejectedValueOnce({ status: 409 })
    const input = screen.getByRole('textbox', { name: '整段小说原文' })
    fireEvent.change(input, { target: { value: '未保存但不能丢失的编辑。' } })
    fireEvent.click(screen.getByRole('button', { name: '保存小说面板' }))
    expect(await screen.findByText(/保存冲突/)).toBeTruthy()
    expect(screen.getByRole('textbox', { name: '整段小说原文' }).value).toBe('未保存但不能丢失的编辑。')
  }, 15000)
})
