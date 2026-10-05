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
})

describe('小说获取工作台', () => {
  it('提供书城分组、执行入口和书籍结果区', async () => {
    document.body.innerHTML = '<div id="root"></div>'
    global.fetch = vi.fn(async (url) => {
      if (url === '/api/auth/current-user') {
        return new Response(JSON.stringify({ user: { id: 7, name: 'Test User', role: 'admin' } }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      return new Response(JSON.stringify({}), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })

    const { appRoot } = await import('./main.jsx')
    try {
      expect(await screen.findByText('小说获取工作台')).toBeTruthy()
      expect(screen.getByRole('button', { name: '添加书城' })).toBeTruthy()
      expect(screen.getByRole('button', { name: '立即执行' })).toBeTruthy()
      expect(screen.getByRole('button', { name: '自动化' })).toBeTruthy()
      expect(screen.getByText('书籍结果')).toBeTruthy()
    } finally {
      await act(async () => {
        appRoot.unmount()
      })
    }
  }, 15000)
})
