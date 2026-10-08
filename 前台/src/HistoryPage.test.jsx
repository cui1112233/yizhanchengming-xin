// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HistoryPage from './HistoryPage.jsx'

vi.mock('./api.js', () => ({ listWorkspaceHistory: vi.fn() }))
import * as api from './api.js'

describe('HistoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listWorkspaceHistory.mockResolvedValue({ entries: [{ id: '3-7', projectId: 3, projectName: '历史项目', bookId: 7, title: '历史小说', status: 'failed', updatedAt: '2026-10-07T09:00:00Z' }], page: 1, total: 1 })
  })

  it('按服务端用户范围投影展示、搜索并可进入项目', async () => {
    render(<HistoryPage />)
    expect(await screen.findByText('历史小说')).toBeTruthy()
    fireEvent.change(screen.getByLabelText('搜索历史'), { target: { value: '历史小说' } })
    fireEvent.keyDown(screen.getByLabelText('搜索历史'), { key: 'Enter', code: 'Enter' })
    expect(await screen.findByRole('link', { name: '打开项目' })).toBeTruthy()
    expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ q: '历史小说' }))
  })

  it('sends status filtering to the MySQL projection instead of filtering browser history', async () => {
    render(<HistoryPage />)
    await screen.findByText('历史小说')
    fireEvent.mouseDown(screen.getByRole('combobox', { name: '筛选历史状态' }))
    fireEvent.click(await screen.findByText('failed'))
    await waitFor(() => expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'failed', page: 1 })))
  })
})
