// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import AuthBoundary from './AuthBoundary.jsx'

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
})

describe('AuthBoundary', () => {
  it('keeps bootstrap in initializing until current-user resolves', async () => {
    let resolveUser
    const api = {
      getCurrentUser: vi.fn(() => new Promise((resolve) => { resolveUser = resolve })),
      login: vi.fn(),
    }

    render(<AuthBoundary api={api}><div>受保护工作台</div></AuthBoundary>)

    expect(screen.getByText('正在恢复登录状态…')).toBeTruthy()
    expect(screen.queryByText('受保护工作台')).toBeNull()

    resolveUser({ user: { id: 7, name: 'Alice' } })
    expect(await screen.findByText('受保护工作台')).toBeTruthy()
  })

  it('shows login only after recovery fails and preserves the original target URL', async () => {
    window.history.replaceState({}, '', '/batch-factory?project=123')
    const api = {
      getCurrentUser: vi.fn(async () => { throw Object.assign(new Error('expired'), { status: 401 }) }),
      login: vi.fn(async () => ({ user: { id: 7, name: 'Alice' } })),
    }

    render(<AuthBoundary api={api}><div>受保护工作台</div></AuthBoundary>)

    expect(await screen.findByRole('button', { name: /登\s*录/ })).toBeTruthy()
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=123')

    fireEvent.change(screen.getByLabelText('用户名'), { target: { value: 'alice' } })
    fireEvent.change(screen.getByLabelText('密码'), { target: { value: 'secret' } })
    fireEvent.click(screen.getByRole('button', { name: /登\s*录/ }))

    await waitFor(() => expect(api.login).toHaveBeenCalledWith({ username: 'alice', password: 'secret' }))
    expect(await screen.findByText('受保护工作台')).toBeTruthy()
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=123')
  })

  it('moves to unauthenticated once when global recovery reports session expiry', async () => {
    const api = {
      getCurrentUser: vi.fn(async () => ({ user: { id: 7, name: 'Alice' } })),
      login: vi.fn(),
    }

    render(<AuthBoundary api={api}><div>受保护工作台</div></AuthBoundary>)
    expect(await screen.findByText('受保护工作台')).toBeTruthy()

    window.dispatchEvent(new CustomEvent('ycm:auth-unauthenticated'))

    expect(await screen.findByRole('button', { name: /登\s*录/ })).toBeTruthy()
    expect(screen.queryByText('受保护工作台')).toBeNull()
  })
})
