// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import AccountCenterPage from './AccountCenterPage.jsx'
afterEach(() => { cleanup(); vi.restoreAllMocks() })
describe('AccountCenterPage', () => {
  it('renders only the authenticated user’s safe server profile', () => {
    render(<AccountCenterPage user={{ username: 'alice', name: 'Alice', role: 'member', teamId: 3, capabilities: ['batch.view'] }} onLogout={vi.fn()} />)
    expect(screen.getByRole('heading', { name: '个人资料' })).toBeTruthy(); expect(screen.getByText('Alice')).toBeTruthy(); expect(screen.getByText('batch.view')).toBeTruthy(); expect(screen.queryByText('ycm_access')).toBeNull()
  })
  it('delegates logout to the existing session boundary', () => { const onLogout = vi.fn(); render(<AccountCenterPage user={{ username: 'alice', role: 'member' }} onLogout={onLogout} />); fireEvent.click(screen.getByRole('button', { name: '退出登录' })); expect(onLogout).toHaveBeenCalledTimes(1) })
})
