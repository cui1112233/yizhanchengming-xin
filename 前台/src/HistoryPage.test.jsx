// @vitest-environment jsdom
import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HistoryPage from './HistoryPage.jsx'

vi.mock('./api.js', () => ({ listWorkspaceHistory: vi.fn() }))
import * as api from './api.js'
const batch = { id: 'batch:project:3', kind: 'batch', origin: 'project', sourceId: '3', intakeId: 2, projectId: 3, bookId: 0, bookRunId: 0, attempt: 0, revision: 0, title: '历史项目', projectName: '历史项目', status: 'failed', sourceStatus: 'failed', errorCode: 'HISTORY_SOURCE_FAILED', updatedAt: '2026-10-07T09:00:00Z', archivedAt: null, href: '/batch-factory?projectId=3', canPreview: false }
const page = (entries = [batch], input = {}) => ({ entries, page: 1, limit: 20, total: entries.length, ...input })
async function select(label, text) { fireEvent.mouseDown(screen.getByRole('combobox', { name: label })); fireEvent.click(await screen.findByText(text, { selector: '.ant-select-item-option-content' })) }

describe('HistoryPage', () => {
  beforeEach(() => { api.listWorkspaceHistory.mockReset(); api.listWorkspaceHistory.mockResolvedValue(page()) })

  it('uses an independent history table with all five columns reachable by scrolling', async () => {
    const { container } = render(<HistoryPage />)
    expect(await screen.findByRole('link', { name: '查看项目' })).toHaveProperty('href', expect.stringContaining('/batch-factory?projectId=3'))
    expect(container.querySelector('main.history-page')).toBeTruthy()
    expect(container.querySelector('.ant-card')).toBeNull()
    expect(screen.getAllByRole('columnheader')).toHaveLength(5)
    expect(container.querySelector('.ant-table-content').style.overflowX).toBe('auto')
    expect(container.querySelector('table').style.minWidth).toBe('100%')
    expect(screen.queryByText('清空历史')).toBeNull(); expect(screen.queryByText('删除')).toBeNull()
  })

  it('submits search, kind, status and archive filters to the server and resets the page', async () => {
    render(<HistoryPage />); await screen.findByRole('link', { name: '查看项目' })
    fireEvent.change(screen.getByLabelText('搜索历史'), { target: { value: '历史小说' } })
    expect(api.listWorkspaceHistory).toHaveBeenCalledTimes(1)
    fireEvent.keyDown(screen.getByLabelText('搜索历史'), { key: 'Enter', code: 'Enter' })
    await waitFor(() => expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, limit: 20, q: '历史小说', archived: 'all' })))
    await select('筛选历史类型', '剧本'); await select('筛选历史状态', '失败'); await select('筛选历史归档', '已归档')
    await waitFor(() => expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'script', status: 'failed', archived: 'archived', q: '历史小说', page: 1 })))
  })

  it('paginates on the server and refreshes the currently selected page', async () => {
    api.listWorkspaceHistory.mockResolvedValueOnce(page([batch], { total: 41 })).mockResolvedValue(page([batch], { total: 41, page: 2 }))
    render(<HistoryPage />); await screen.findByRole('link', { name: '查看项目' })
    fireEvent.click(screen.getByTitle('Next Page'))
    await waitFor(() => expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 })))
    fireEvent.click(screen.getByRole('button', { name: '刷新' }))
    await waitFor(() => expect(api.listWorkspaceHistory).toHaveBeenCalledTimes(3))
    expect(api.listWorkspaceHistory).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
  })

  it('does not imply restore or silently enter unsupported sources or unsafe hrefs', async () => {
    api.listWorkspaceHistory.mockResolvedValue(page([{ ...batch, id: 'novel_panel:revision:9', kind: 'novel_panel', title: '小说分镜修订', href: '', revision: 2, archivedAt: '2026-10-08T00:00:00Z' }, { ...batch, id: 'bad', title: '坏链接', href: '//evil.invalid' }]))
    render(<HistoryPage />); await screen.findByText('小说分镜修订')
    expect(screen.queryByRole('link')).toBeNull()
    expect(screen.getAllByText('暂不支持精确进入')).toHaveLength(2)
    expect(screen.getByText('已归档 · 只读')).toBeTruthy()
    expect(screen.queryByRole('button', { name: /恢复|预览|删除|清空/ })).toBeNull()
  })

  it.each([[401, '登录已过期，请重新登录'], [403, '没有权限查看历史记录'], [503, '历史服务暂不可用']])('shows first-load HTTP %s and request correlation with retry', async (status, message) => {
    api.listWorkspaceHistory.mockRejectedValueOnce({ status, message: '内部错误', requestId: 'history-request-7' }).mockResolvedValue(page())
    render(<HistoryPage />); expect(await screen.findByText(message)).toBeTruthy(); expect(screen.getByText('请求编号：history-request-7')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ })); expect(await screen.findByRole('link', { name: '查看项目' })).toBeTruthy()
  })

  it('keeps prior successful rows visible after a failed refresh and identifies their age', async () => {
    api.listWorkspaceHistory.mockResolvedValueOnce(page()).mockRejectedValueOnce({ status: 503, requestId: 'refresh-request' })
    render(<HistoryPage />); await screen.findByRole('link', { name: '查看项目' }); fireEvent.click(screen.getByRole('button', { name: '刷新' }))
    expect(await screen.findByText('历史服务暂不可用')).toBeTruthy(); expect(screen.getByRole('link', { name: '查看项目' })).toBeTruthy()
    expect(screen.getByText(/保留上次成功读取的记录/)).toBeTruthy(); expect(screen.getByText(/refresh-request/)).toBeTruthy()
  })

  it('ignores late successful and failed requests after a newer filter result', async () => {
    let resolveOld; let rejectOld
    api.listWorkspaceHistory.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve })).mockResolvedValueOnce(page([{ ...batch, title: '最新结果' }]))
    render(<HistoryPage />)
    await select('筛选历史类型', '剧本'); expect(await screen.findByText('最新结果')).toBeTruthy()
    await act(async () => { resolveOld(page([{ ...batch, title: '过时结果' }])) })
    expect(screen.queryByText('过时结果')).toBeNull()
    api.listWorkspaceHistory.mockImplementationOnce(() => new Promise((_, reject) => { rejectOld = reject })).mockResolvedValueOnce(page([{ ...batch, title: '更新结果' }]))
    fireEvent.click(screen.getByRole('button', { name: '刷新' })); await select('筛选历史状态', '失败'); expect(await screen.findByText('更新结果')).toBeTruthy()
    await act(async () => { rejectOld({ status: 503, requestId: 'late-failure' }) }); expect(screen.queryByText(/late-failure/)).toBeNull()
  })

  it('shows true empty state and unknown status without claiming completion', async () => {
    api.listWorkspaceHistory.mockResolvedValueOnce(page([])).mockResolvedValueOnce(page([{ ...batch, status: 'unknown' }]))
    render(<HistoryPage />); expect(await screen.findByText('暂无可查看的历史记录')).toBeTruthy(); fireEvent.click(screen.getByRole('button', { name: '刷新' })); expect(await screen.findByText('未知状态')).toBeTruthy()
  })
})
