// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
vi.mock('./api.js', () => ({ listBatchProjects: vi.fn() }))
import { listBatchProjects } from './api.js'
import HomePage from './HomePage.jsx'

afterEach(() => { cleanup(); vi.clearAllMocks() })
describe('HomePage recent projects', () => {
  it('keeps the Agent product entrance marked for redesign and opens only the reserved route', async () => {
    listBatchProjects.mockResolvedValue({ projects: [] })
    const onNavigate = vi.fn()
    render(<HomePage onNavigate={onNavigate} />)
    const agent = screen.getByRole('link', { name: /AI 智能 Agent/ })
    expect(within(agent).getByText('待重新设计')).toBeTruthy()
    expect(agent.getAttribute('href')).toBe('/agent')
    fireEvent.click(agent)
    expect(onNavigate).toHaveBeenCalledWith('/agent')
    await screen.findByText('暂无可访问的最近项目')
  })
  it('renders only server supplied accessible projects and opens the selected project', async () => {
    listBatchProjects.mockResolvedValue({ projects: [{ id: 18, name: '已授权项目', bookCount: 2, runStatus: 'running' }] })
    render(<HomePage onNavigate={() => {}} />)
    const link = await screen.findByRole('link', { name: /已授权项目/ })
    expect(link.getAttribute('href')).toBe('/shuihuo-production?projectId=18')
    expect(screen.getByText('2 本小说 · running')).toBeTruthy()
  })

  it('keeps server supplied recent projects visible after refresh', async () => {
    listBatchProjects.mockResolvedValueOnce({ projects: [{ id: 18, name: '刷新前项目', bookCount: 1, runStatus: 'running' }] }).mockResolvedValueOnce({ projects: [{ id: 19, name: '刷新后项目', bookCount: 3, runStatus: 'completed' }] })
    render(<HomePage onNavigate={() => {}} />)
    await screen.findByText('刷新前项目')
    fireEvent.click(screen.getByRole('button', { name: '刷新最近项目' }))
    expect(await screen.findByText('刷新后项目')).toBeTruthy()
  })

  it('renders controlled local Hero media with a readable fallback', async () => {
    listBatchProjects.mockResolvedValue({ projects: [] })
    render(<HomePage onNavigate={() => {}} />)
    const hero = await screen.findByLabelText('首页视觉背景')
    expect(hero.getAttribute('src')).toBe('/assets/home-hero.mp4')
    expect(hero.getAttribute('poster')).toBe('/assets/brand-logo-white.png')
  })

  it('shows the Go API empty state without inventing a project', async () => {
    listBatchProjects.mockResolvedValue({ projects: [] })
    render(<HomePage onNavigate={() => {}} />)
    expect(await screen.findByText('暂无可访问的最近项目')).toBeTruthy()
    expect(document.querySelector('[href*="projectId="]')).toBeNull()
  })

  it('retries a failed recent-project read and restores the returned project', async () => {
    listBatchProjects.mockRejectedValueOnce(new Error('暂时无法读取项目')).mockResolvedValueOnce({ projects: [{ id: 25, name: '重试恢复项目', bookCount: 1, runStatus: 'pending' }] })
    render(<HomePage onNavigate={() => {}} />)
    expect(await screen.findByText('暂时无法读取项目')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /^重\s*试$/ }))
    expect(await screen.findByText('重试恢复项目')).toBeTruthy()
  })
})
