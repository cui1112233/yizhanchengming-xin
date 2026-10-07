// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HistoryPage from './HistoryPage.jsx'

vi.mock('./api.js', () => ({ listBatchProjects: vi.fn(), getProjectGeneration: vi.fn() }))
import * as api from './api.js'

describe('HistoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 3, name: '历史项目', runStatus: 'running' }] })
    api.getProjectGeneration.mockResolvedValue({ books: [{ bookId: 7, title: '历史小说', status: 'failed', stages: { SCRIPT: { status: 'completed', updatedAt: '2026-10-07T08:00:00Z' }, DIRECTOR: { status: 'failed', updatedAt: '2026-10-07T09:00:00Z' } } }] })
  })

  it('按服务端项目和生成阶段展示、搜索并可进入项目', async () => {
    render(<HistoryPage />)
    expect(await screen.findByText('历史小说')).toBeTruthy()
    expect(screen.getByText('DIRECTOR: failed')).toBeTruthy()
    fireEvent.change(screen.getByLabelText('搜索历史'), { target: { value: '不存在' } })
    expect(await screen.findByText('暂无可查看的项目历史')).toBeTruthy()
    fireEvent.change(screen.getByLabelText('搜索历史'), { target: { value: '历史小说' } })
    expect(await screen.findByRole('link', { name: '打开项目' })).toBeTruthy()
    expect(api.getProjectGeneration).toHaveBeenCalledWith(3)
  })
})
