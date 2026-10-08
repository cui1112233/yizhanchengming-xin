// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import RouterApp from './RouterApp.jsx'

beforeAll(() => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }),
  })
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/')
})

describe('Task 1 首页与用户路由基础', () => {
  it('根路由展示公网首页主视觉、六个创作入口和最近创作入口，不再展示小说获取工作台', () => {
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)

    expect(screen.getByRole('heading', { name: '让小说章节直接进入可视化剧本工作流' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '开始生成' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '进入配音' })).toBeTruthy()
    expect(screen.queryByRole('heading', { name: '小说获取工作台' })).toBeNull()

    const expectedRoutes = ['/script', '/novel-panel', '/batch-factory', '/shuihuo-production', '/agent', '/tts']
    const actionLinks = screen.getAllByRole('link').filter((node) => expectedRoutes.includes(node.getAttribute('href')))
    expect(new Set(actionLinks.map((node) => node.getAttribute('href')))).toEqual(new Set(expectedRoutes))
    expect(screen.getByRole('heading', { name: '最近创作项目' })).toBeTruthy()
    expect(screen.getByText('正在读取你的最近项目…')).toBeTruthy()
  })

  it('用户导航保留全部基础入口、主题、设置和用户入口', () => {
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)

    for (const label of ['首页', '剧本生成', '小说获取', '小说面板', '水货生产', 'Agent 工作区', '历史', '配音']) {
      expect(screen.getByRole('link', { name: label })).toBeTruthy()
    }
    expect(screen.queryByRole('link', { name: '问题日志' })).toBeNull()
    expect(screen.getByRole('button', { name: '主题' })).toBeTruthy()
    expect(screen.getByRole('link', { name: '设置' }).getAttribute('href')).toBe('/settings')
    expect(screen.getByRole('button', { name: '用户入口' })).toBeTruthy()
  })

  it('/novel-fetch 继续渲染现有小说获取页面', async () => {
    window.history.replaceState({}, '', '/novel-fetch')
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      status: 200,
      headers: { get: () => '' },
      json: async () => ({ intakes: [] }),
    })

    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)

    expect(await screen.findByRole('heading', { name: '小说获取' })).toBeTruthy()
    expect(screen.queryByRole('heading', { name: '让小说章节直接进入可视化剧本工作流' })).toBeNull()
    expect(window.location.pathname).toBe('/novel-fetch')
  })

  it('/issues renders the 404 page without requesting issue data', async () => {
    window.history.replaceState({}, '', '/issues')
    const fetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true, status: 200, headers: { get: () => '' }, json: async () => ({ issues: [] }) })
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByText('页面不存在')).toBeTruthy()
    expect(fetch).not.toHaveBeenCalled()
  })

  it.each(['/agent', '/agent/canvas?projectId=17'])('%s renders the reserved Agent page without requesting Agent data', async (path) => {
    window.history.replaceState({}, '', path)
    const fetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true, status: 200, headers: { get: () => '' }, json: async () => ({ projects: [], messages: [], executions: [], skills: [], attachments: [], versions: [], canvas: {} }) })
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByText('Agent 工作区待重新设计')).toBeTruthy()
    expect(screen.getByText('Agent 将重新设计，当前不可用')).toBeTruthy()
    expect(screen.getByText('当前版本不执行 Agent 任务，也不会创建项目或调用 Provider。')).toBeTruthy()
    expect(fetch).not.toHaveBeenCalled()
  })

  it('three production URLs mount distinct roots instead of IntakeWorkbench', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true, status: 200, headers: { get: () => '' }, json: async () => ({ intakes: [], projects: [] }) })
    window.history.replaceState({}, '', '/novel-fetch')
    const first = render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByRole('heading', { name: '小说获取' })).toBeTruthy()
    first.unmount()
    window.history.replaceState({}, '', '/batch-factory')
    const second = render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByRole('heading', { name: '批量工厂' })).toBeTruthy()
    second.unmount()
    window.history.replaceState({}, '', '/shuihuo-production')
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByRole('heading', { name: '水货生产' })).toBeTruthy()
  })

  it('首页开始生成进入 /script', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      status: 200,
      headers: { get: () => '' },
      json: async () => ({ projects: [] }),
    })
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)

    fireEvent.click(screen.getByRole('button', { name: '开始生成' }))
    expect(window.location.pathname).toBe('/script')
    expect(await screen.findByText('暂无项目。请先在小说获取中保存原文，再创建项目。')).toBeTruthy()
  })

  it('主题按钮只调用主题切换，不保存任何项目业务数据', () => {
    const onToggleTheme = vi.fn()
    render(<RouterApp theme="dark" onToggleTheme={onToggleTheme} />)
    fireEvent.click(screen.getByRole('button', { name: '主题' }))
    expect(onToggleTheme).toHaveBeenCalledTimes(1)
  })

  it('/tts 与 /history 渲染真实服务端投影页面而不是基础占位页', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true, status: 200, headers: { get: () => '' }, json: async () => ({ projects: [] }) })
    window.history.replaceState({}, '', '/tts')
    const { unmount } = render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByRole('heading', { name: '配音' })).toBeTruthy()
    expect(screen.queryByText(/路由基础已恢复/)).toBeNull()
    unmount()
    window.history.replaceState({}, '', '/history')
    render(<RouterApp theme="dark" onToggleTheme={() => {}} />)
    expect(await screen.findByRole('heading', { name: '历史记录' })).toBeTruthy()
  })
})
