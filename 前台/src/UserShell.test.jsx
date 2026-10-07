// @vitest-environment jsdom

import React from 'react'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import UserShell from './UserShell.jsx'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('UserShell', () => {
  it('uses the authenticated session name for the account menu trigger', () => {
    render(
      <UserShell
        pathname="/"
        theme="dark"
        onToggleTheme={() => {}}
        onNavigate={() => {}}
        currentUser={{ name: '测试创作者' }}
      >
        <main>内容</main>
      </UserShell>,
    )

    expect(screen.getByRole('button', { name: '测试创作者' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: '用户入口' })).toBeNull()
  })

  it('provides a collapsible desktop workspace sidebar outside the home route', () => {
    render(
      <UserShell
        pathname="/history"
        theme="dark"
        onToggleTheme={() => {}}
        onNavigate={() => {}}
        currentUser={{ name: '测试创作者' }}
      >
        <main>内容</main>
      </UserShell>,
    )

    const sidebar = screen.getByRole('navigation', { name: '工作区侧边导航' })
    expect(within(sidebar).getByRole('link', { name: '历史' }).getAttribute('href')).toBe('/history')
    fireEvent.click(screen.getByRole('button', { name: '收起侧边导航' }))
    expect(screen.getByRole('button', { name: '展开侧边导航' })).toBeTruthy()
  })
})
