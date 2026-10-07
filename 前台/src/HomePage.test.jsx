// @vitest-environment jsdom
import React from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
vi.mock('./api.js', () => ({ listBatchProjects: vi.fn() }))
import { listBatchProjects } from './api.js'
import HomePage from './HomePage.jsx'

afterEach(() => { cleanup(); vi.clearAllMocks() })
describe('HomePage recent projects', () => {
  it('renders only server supplied accessible projects and opens the selected project', async () => {
    listBatchProjects.mockResolvedValue({ projects: [{ id: 18, name: '已授权项目', bookCount: 2, runStatus: 'running' }] })
    render(<HomePage onNavigate={() => {}} />)
    const link = await screen.findByRole('link', { name: /已授权项目/ })
    expect(link.getAttribute('href')).toBe('/shuihuo-production?projectId=18')
    expect(screen.getByText('2 本小说 · running')).toBeTruthy()
  })
})
