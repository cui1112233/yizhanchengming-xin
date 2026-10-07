// @vitest-environment jsdom

import React from 'react'
import { render, screen } from '@testing-library/react'
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
})
