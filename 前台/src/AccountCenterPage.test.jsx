// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import AccountCenterPage from './AccountCenterPage.jsx'
import { __resetAuthRecoveryForTests } from './api.js'

const response = (status, body) => ({ ok: status >= 200 && status < 300, status, headers: new Headers(), json: async () => body })
const alice = { profile: { id: 7, username: 'alice', displayName: 'Alice', avatar: { availability: 'unavailable' } }, membership: { role: 'member', teamId: 3, capabilities: ['batch.view'] }, security: { emailVerification: 'unavailable', mfa: 'unavailable' } }
afterEach(() => { cleanup(); vi.restoreAllMocks(); __resetAuthRecoveryForTests() })

describe('AccountCenterPage', () => {
  it('renders Alice only from the new safe profile API', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(response(200, alice))
    render(<AccountCenterPage user={{ username: 'stale-user' }} onLogout={vi.fn()} />)
    expect(await screen.findByText('Alice')).toBeTruthy()
    expect(screen.getByText('batch.view')).toBeTruthy()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/account/profile', expect.objectContaining({ credentials: 'include' }))
    expect(screen.queryByText('stale-user')).toBeNull()
    expect(screen.queryByText('ycm_access')).toBeNull()
  })
  it('shows loading then server-backed member unavailable state', async () => {
    let resolve
    vi.spyOn(globalThis, 'fetch').mockImplementation(() => new Promise(done => { resolve = done }))
    render(<AccountCenterPage user={{}} onLogout={vi.fn()} mode="member" />)
    expect(document.querySelector('[aria-busy="true"]')).toBeTruthy()
    resolve(response(200, { ...alice, usage: { availability: 'unavailable' }, activity: [], notifications: [] }))
    expect(await screen.findByText('用量、活动与通知')).toBeTruthy()
    expect(screen.getByText(/MySQL 投影/)).toBeTruthy()
  })
  it('shows an API failure without falling back to browser profile data', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(response(500, { message: '服务暂不可用' }))
    render(<AccountCenterPage user={{ username: 'Alice from browser' }} onLogout={vi.fn()} />)
    expect(await screen.findByText('服务暂不可用')).toBeTruthy()
    expect(screen.queryByText('Alice from browser')).toBeNull()
  })
  it('handles 401 and 403 from the new API', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch')
    fetchMock.mockResolvedValueOnce(response(401, { code: 'AUTH_UNAUTHENTICATED', message: '登录状态无效' })).mockResolvedValueOnce(response(401, { code: 'AUTH_UNAUTHENTICATED', message: '登录状态无效' }))
    const { unmount } = render(<AccountCenterPage user={{}} onLogout={vi.fn()} />)
    expect(await screen.findByText('登录状态无效')).toBeTruthy()
    unmount(); fetchMock.mockReset().mockResolvedValue(response(403, { code: 'AUTH_FORBIDDEN', message: '没有权限' }))
    render(<AccountCenterPage user={{}} onLogout={vi.fn()} mode="member" />)
    expect(await screen.findByText('没有权限')).toBeTruthy()
  })
  it('saves display name through the CSRF-protected profile API', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(response(200, alice)).mockResolvedValueOnce(response(200, { ...alice, profile: { ...alice.profile, displayName: 'Alice New' } }))
    render(<AccountCenterPage user={{}} onLogout={vi.fn()} />)
    const input = await screen.findByRole('textbox', { name: '显示名称' })
    fireEvent.change(input, { target: { value: 'Alice New' } }); fireEvent.click(screen.getByRole('button', { name: '保存资料' }))
    await waitFor(() => expect(screen.getByText('Alice New')).toBeTruthy())
    expect(fetchMock).toHaveBeenLastCalledWith('/api/v1/account/profile', expect.objectContaining({ method: 'PUT', body: JSON.stringify({ displayName: 'Alice New' }) }))
  })
  it('delegates logout to the existing session boundary', async () => { vi.spyOn(globalThis, 'fetch').mockResolvedValue(response(200, alice)); const onLogout = vi.fn(); render(<AccountCenterPage user={{}} onLogout={onLogout} />); await screen.findByText('Alice'); fireEvent.click(screen.getByRole('button', { name: '退出登录' })); expect(onLogout).toHaveBeenCalledTimes(1) })
})
