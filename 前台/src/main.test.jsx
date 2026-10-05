// @vitest-environment jsdom

import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

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

describe('小说获取工作台', () => {
  it('提供书城分组、执行入口和书籍结果区', async () => {
    document.body.innerHTML = '<div id="root"></div>'

    await import('./main.jsx')

    expect(await screen.findByText('小说获取工作台')).toBeTruthy()
    expect(screen.getByRole('button', { name: '添加书城' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '立即执行' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '自动化' })).toBeTruthy()
    expect(screen.getByText('书籍结果')).toBeTruthy()
  })
})
