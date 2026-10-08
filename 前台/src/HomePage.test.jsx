// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
vi.mock('./api.js', () => ({ listWorkspaceRecent: vi.fn() }))
import { listWorkspaceRecent } from './api.js'
import HomePage from './HomePage.jsx'

afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.clearAllMocks() })
describe('HomePage recent projects', () => {
  it('keeps the Agent product entrance marked for redesign and opens only the reserved route', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [] })
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
    listWorkspaceRecent.mockResolvedValue({ items: [{ kind: 'script', id: 'script:18', title: '已授权项目', status: 'running', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=18' }] })
    const onNavigate = vi.fn()
    render(<HomePage onNavigate={onNavigate} />)
    const link = await screen.findByRole('link', { name: /已授权项目/ })
    expect(link.getAttribute('href')).toBe('/batch-factory?projectId=18')
    expect(screen.getByText(/剧本生成 · 执行中/)).toBeTruthy()
    fireEvent.click(link)
    expect(onNavigate).toHaveBeenCalledWith('/batch-factory?projectId=18')
  })

  it('uses source-namespaced identities as distinct React keys for colliding TTS row ids', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    listWorkspaceRecent.mockResolvedValue({ items: [
      { kind: 'tts', id: 'tts:measurement:44', title: '测量配音', status: 'measured', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=12' },
      { kind: 'tts', id: 'tts:media:44', title: '水货音频', status: 'completed', updatedAt: '2026-10-09T04:04:06Z', href: '/batch-factory?projectId=12' },
    ] })
    render(<HomePage onNavigate={() => {}} />)
    await screen.findByText('测量配音')
    expect(screen.getByText('水货音频')).toBeTruthy()
    expect(document.querySelector('[data-recent-id="tts:measurement:44"]')).toBeTruthy()
    expect(document.querySelector('[data-recent-id="tts:media:44"]')).toBeTruthy()
    expect(consoleError.mock.calls.flat().join(' ')).not.toContain('same key')
  })

  it('keeps server supplied recent work visible while refresh is pending', async () => {
    let resolveRefresh
    listWorkspaceRecent.mockResolvedValueOnce({ items: [{ kind: 'batch', id: 'batch:18', title: '刷新前项目', status: 'running', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=18' }] }).mockImplementationOnce(() => new Promise((resolve) => { resolveRefresh = resolve }))
    render(<HomePage onNavigate={() => {}} />)
    await screen.findByText('刷新前项目')
    fireEvent.click(screen.getByRole('button', { name: '刷新最近项目' }))
    expect(screen.getByText('刷新前项目')).toBeTruthy()
    resolveRefresh({ items: [{ kind: 'video', id: 'video:19', title: '刷新后项目', status: 'completed', updatedAt: '2026-10-09T05:05:06Z', href: '/batch-factory?projectId=19' }] })
    expect(await screen.findByText('刷新后项目')).toBeTruthy()
  })

  it('keeps old work visible and shows a correlated error when refresh fails', async () => {
    listWorkspaceRecent.mockResolvedValueOnce({ items: [{ kind: 'batch', id: 'batch:18', title: '仍可进入的项目', status: 'running', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=18' }] })
      .mockRejectedValueOnce(Object.assign(new Error('读取最近创作失败'), { status: 500, requestId: 'req-refresh-9' }))
    render(<HomePage onNavigate={() => {}} />)
    await screen.findByText('仍可进入的项目')
    fireEvent.click(screen.getByRole('button', { name: '刷新最近项目' }))
    expect(await screen.findByText('请求编号：req-refresh-9')).toBeTruthy()
    expect(screen.getByRole('link', { name: /仍可进入的项目/ })).toBeTruthy()
  })

  it('does not expose an unclassified client exception', async () => {
    listWorkspaceRecent.mockRejectedValueOnce(new Error('mysql://root:secret@prod/private'))
    render(<HomePage onNavigate={() => {}} />)
    expect(await screen.findByText('最近创作暂时无法读取，请稍后重试。')).toBeTruthy()
    expect(document.body.textContent).not.toContain('secret')
  })

  it('does not let an older slow response overwrite a newer refresh', async () => {
    let resolveInitial
    listWorkspaceRecent.mockImplementationOnce(() => new Promise((resolve) => { resolveInitial = resolve }))
      .mockResolvedValueOnce({ items: [{ kind: 'batch', id: 'batch:2', title: '较新的结果', status: 'completed', updatedAt: '2026-10-09T05:05:06Z', href: '/batch-factory?projectId=2' }] })
    render(<HomePage onNavigate={() => {}} />)
    fireEvent.click(screen.getByRole('button', { name: '刷新最近项目' }))
    expect(await screen.findByText('较新的结果')).toBeTruthy()
    resolveInitial({ items: [{ kind: 'batch', id: 'batch:1', title: '过期慢响应', status: 'running', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=1' }] })
    await waitFor(() => expect(screen.queryByText('过期慢响应')).toBeNull())
  })

  it('renders controlled local Hero media with a readable fallback', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [] })
    render(<HomePage onNavigate={() => {}} />)
    const hero = await screen.findByLabelText('首页视觉背景')
    expect(hero.getAttribute('src')).toBe('/assets/home-hero.mp4')
    expect(hero.getAttribute('poster')).toBe('/assets/brand-logo-white.png')
  })

  it('shows the Go API empty state without inventing a project', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [] })
    render(<HomePage onNavigate={() => {}} />)
    expect(await screen.findByText('暂无可访问的最近项目')).toBeTruthy()
    expect(document.querySelector('[href*="projectId="]')).toBeNull()
  })

  it('retries a failed recent-project read and restores the returned project', async () => {
    listWorkspaceRecent.mockRejectedValueOnce(Object.assign(new Error('读取最近创作失败'), { requestId: 'req-home-123', status: 500 })).mockResolvedValueOnce({ items: [{ kind: 'batch', id: 'batch:25', title: '重试恢复项目', status: 'pending', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=25' }] })
    render(<HomePage onNavigate={() => {}} />)
    expect(await screen.findByText('读取最近创作失败')).toBeTruthy()
    expect(screen.getByText('请求编号：req-home-123')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /^重\s*试$/ }))
    expect(await screen.findByText('重试恢复项目')).toBeTruthy()
  })

  it('keeps invalid or external server hrefs visible but not clickable', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [
      { kind: 'batch', id: 'batch:1', title: '外部入口', status: 'completed', updatedAt: '2026-10-09T04:05:06Z', href: 'https://evil.example/batch-factory?projectId=1' },
      { kind: 'batch', id: 'batch:2', title: '额外参数', status: 'completed', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=2&next=/admin' },
      { kind: 'batch', id: 'batch:3', title: '非法编号', status: 'completed', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=0' },
    ] })
    render(<HomePage onNavigate={() => {}} />)
    await screen.findByText('外部入口')
    for (const title of ['外部入口', '额外参数', '非法编号']) {
      expect(screen.queryByRole('link', { name: new RegExp(title) })).toBeNull()
    }
    expect(screen.getAllByText('入口不可用')).toHaveLength(3)
  })

  it('preserves browser behavior for modified clicks', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [{ kind: 'batch', id: 'batch:18', title: '可新标签打开', status: 'running', updatedAt: '2026-10-09T04:05:06Z', href: '/batch-factory?projectId=18' }] })
    const onNavigate = vi.fn()
    render(<HomePage onNavigate={onNavigate} />)
    window.addEventListener('click', (event) => event.preventDefault(), { once: true })
    fireEvent.click(await screen.findByRole('link', { name: /可新标签打开/ }), { ctrlKey: true })
    expect(onNavigate).not.toHaveBeenCalled()
  })

  it('renders the public contact callout and opens server settings', async () => {
    listWorkspaceRecent.mockResolvedValue({ items: [] })
    const onNavigate = vi.fn()
    render(<HomePage onNavigate={onNavigate} />)
    const callout = document.getElementById('contact')
    expect(callout).toBeTruthy()
    expect(within(callout).getByText('需要配置账号或部署环境？')).toBeTruthy()
    fireEvent.click(within(callout).getByRole('button', { name: '打开设置' }))
    expect(onNavigate).toHaveBeenCalledWith('/settings')
  })
})
