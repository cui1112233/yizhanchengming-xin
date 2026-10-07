// @vitest-environment jsdom
import React from 'react'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
vi.mock('./api.js', () => ({ listIssues: vi.fn() }))
import { listIssues } from './api.js'
import IssuesPage from './IssuesPage.jsx'
afterEach(() => { cleanup(); vi.clearAllMocks() })
describe('IssuesPage', () => { it('renders only the server issue projection with filters', async () => { listIssues.mockResolvedValue({ page: 1, entries: [{ id:'stage_run-4', source:'stage_run', status:'failed', message:'safe failure', projectId:2, bookId:8, at:'2026-10-07T00:00:00Z' }] }); render(<IssuesPage />); expect(await screen.findByText('safe failure')).toBeTruthy(); await waitFor(()=>expect(listIssues).toHaveBeenCalledWith(expect.objectContaining({ page:1 }))); expect(screen.queryByText(/secret/i)).toBeNull() }) })
