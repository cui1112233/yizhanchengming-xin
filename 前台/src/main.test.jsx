// @vitest-environment jsdom

import { act, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

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

afterEach(() => {
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/')
  window.localStorage.clear()
})

async function settleReactAfterUnmount() {
  await act(async () => {
    await Promise.resolve()
    await new Promise((resolve) => setTimeout(resolve, 32))
  })
}

describe('用户前台入口', () => {
  it('认证恢复后根路由展示首页而不是小说获取工作台', async () => {
    document.body.innerHTML = '<div id="root"></div>'
    window.history.replaceState({}, '', '/')
    global.fetch = vi.fn(async (url) => {
      if (url === '/api/auth/current-user') {
        return new Response(JSON.stringify({ user: { id: 7, name: 'Test User', role: 'admin' } }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      if (url === '/api/v1/workspace/settings') {
        return new Response(JSON.stringify({ settings: { theme: 'light' }, executors: [] }), { status: 200, headers: { 'Content-Type': 'application/json' } })
      }
      return new Response(JSON.stringify({}), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })

    const { appRoot } = await import('./main.jsx')
    try {
      expect(await screen.findByRole('heading', { name: '让小说章节直接进入可视化剧本工作流' })).toBeTruthy()
      expect(screen.queryByRole('heading', { name: '小说获取工作台' })).toBeNull()
      expect(screen.getByRole('button', { name: '开始生成' })).toBeTruthy()
      expect(screen.getByRole('link', { name: '小说获取' }).getAttribute('href')).toBe('/novel-fetch')
      await act(async () => { await Promise.resolve() })
      expect(document.documentElement.dataset.theme).toBe('light')
    } finally {
      await act(async () => {
        appRoot.unmount()
      })
      await settleReactAfterUnmount()
      document.body.replaceChildren()
    }
  }, 15000)
})
