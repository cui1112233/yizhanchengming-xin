// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('./api.js', () => ({ listIssues: vi.fn() }))

import { listIssues } from './api.js'
import IssuesPage from './IssuesPage.jsx'

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('IssuesPage', () => {
  it('renders only the server issue projection with source, status and search filters', async () => {
    listIssues.mockResolvedValue({ page: 1, entries: [{ id: 'stage_run-4', source: 'stage_run', status: 'failed', message: 'safe failure', projectId: 2, bookId: 8, at: '2026-10-07T00:00:00Z' }] })
    render(<IssuesPage />)

    expect(await screen.findByText('safe failure')).toBeTruthy()
    expect(screen.getByRole('link', { name: '打开项目' }).getAttribute('href')).toBe('/shuihuo-production?projectId=2')
    await waitFor(() => expect(listIssues).toHaveBeenCalledWith(expect.objectContaining({ page: 1, status: '' })))
    fireEvent.mouseDown(screen.getByRole('combobox', { name: '筛选问题状态' }))
    fireEvent.click(await screen.findByText('retryable_failed'))
    await waitFor(() => expect(listIssues).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'retryable_failed' })))
    expect(screen.queryByText(/secret/i)).toBeNull()
  })

  it('shows a specific empty state when the scoped projection has no failures', async () => {
    listIssues.mockResolvedValue({ page: 1, entries: [], total: 0 })
    render(<IssuesPage />)

    expect(await screen.findByText('暂无可查看的问题记录')).toBeTruthy()
  })
})
