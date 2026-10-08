// @vitest-environment jsdom
import React from 'react'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
vi.mock('./api.js', () => ({ getWorkspaceSettings: vi.fn(), saveWorkspaceSettings: vi.fn() }))
import { getWorkspaceSettings, saveWorkspaceSettings } from './api.js'
import SettingsPage from './SettingsPage.jsx'

const settings = { theme: 'dark', notificationsEnabled: false, storagePreference: 'local_executor', petId: 'fox', soundVolume: 23, petVisible: false, companionActive: true }
const row = { id: 'e1', name: 'Mac', providerKey: 'doubao_local_executor', model: 'doubao-seedance', online: false, lastSeenAt: '2026-10-07T00:00:00Z' }
const payload = (extra = {}) => ({ settings, executors: [row], executorStatus: { status: 'degraded', reasonCode: 'offline' }, ...extra })
const deferred = () => { let resolve, reject; const promise = new Promise((yes, no) => { resolve = yes; reject = no }); return { promise, resolve, reject } }
let reads, writes
beforeEach(() => {
  getWorkspaceSettings.mockResolvedValue(payload())
  reads = vi.spyOn(Storage.prototype, 'getItem')
  writes = vi.spyOn(Storage.prototype, 'setItem')
})
afterEach(() => { cleanup(); expect(reads).not.toHaveBeenCalled(); expect(writes).not.toHaveBeenCalled(); vi.restoreAllMocks(); vi.clearAllMocks() })

describe('SettingsPage', () => {
  it('saves Radio changes immediately and adopts normalized theme without losing unrelated form drafts', async () => {
    const transition = deferred()
    const onTheme = vi.fn(() => transition.promise)
    render(<SettingsPage onTheme={onTheme} />)
    await screen.findByText('Mac')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    fireEvent.click(screen.getByLabelText('浅色'))
    expect(onTheme).toHaveBeenCalledWith('light')
    expect(screen.getByLabelText('浅色').checked).toBe(true)
    expect(screen.getByRole('button', { name: /保\s*存/ }).disabled).toBe(true)
    await act(async () => transition.resolve(payload()))
    expect(screen.getByLabelText('深色').checked).toBe(true)
    expect(screen.getByRole('switch', { name: /提醒开启/ }).getAttribute('aria-checked')).toBe('true')
    expect(saveWorkspaceSettings).not.toHaveBeenCalled()
  })

  it('reconciles safe rollback rejection and keeps unrelated draft edits', async () => {
    const onTheme = vi.fn().mockRejectedValue({ rollbackTheme: 'dark', code: 'THEME_SAVE_FAILED', requestId: 'req-radio-1', payload: 'secret-payload', message: '<script>raw-error</script>' })
    render(<SettingsPage onTheme={onTheme} />)
    await screen.findByText('Mac')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    fireEvent.click(screen.getByLabelText('浅色'))
    await vi.waitFor(() => expect(screen.getByLabelText('深色').checked).toBe(true))
    expect(screen.getByRole('switch', { name: /提醒开启/ }).getAttribute('aria-checked')).toBe('true')
    expect(document.body.textContent).not.toMatch(/secret-payload|raw-error/)
  })

  it('sends explicit full Save through the shared queue once and adopts returned normalized settings', async () => {
    const onTheme = vi.fn().mockResolvedValue(payload({ settings: { ...settings, theme: 'light', notificationsEnabled: true } }))
    render(<SettingsPage onTheme={onTheme} />)
    await screen.findByText('Mac')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    fireEvent.click(screen.getByRole('button', { name: /保\s*存/ }))
    await vi.waitFor(() => expect(screen.getByLabelText('浅色').checked).toBe(true))
    expect(onTheme).toHaveBeenCalledExactlyOnceWith(undefined, { saveSettings: { ...settings, notificationsEnabled: true } })
    expect(saveWorkspaceSettings).not.toHaveBeenCalled()
  })

  it('restores the confirmed theme after full Save failure and keeps non-theme drafts', async () => {
    const onTheme = vi.fn().mockRejectedValue({ code: 'SETTINGS_SAVE_FAILED', rollbackTheme: 'light', requestId: 'req-full-save', message: 'private-save', payload: 'secret-save' })
    render(<SettingsPage onTheme={onTheme} />)
    await screen.findByText('Mac')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    fireEvent.click(screen.getByRole('button', { name: /保\s*存/ }))
    await screen.findByText(/req-full-save/)
    expect(screen.getByLabelText('浅色').checked).toBe(true)
    expect(screen.getByRole('switch', { name: /提醒开启/ }).getAttribute('aria-checked')).toBe('true')
    expect(document.body.textContent).not.toMatch(/private-save|secret-save/)
  })

  it('ignores a late StrictMode read after a newer read, theme save, and unrelated draft edit', async () => {
    const first = deferred(), second = deferred()
    getWorkspaceSettings.mockImplementationOnce(() => first.promise).mockImplementationOnce(() => second.promise)
    const onTheme = vi.fn().mockResolvedValue(payload({ settings: { ...settings, theme: 'light' } }))
    render(<React.StrictMode><SettingsPage onTheme={onTheme} /></React.StrictMode>)
    expect(getWorkspaceSettings).toHaveBeenCalledTimes(2)
    await act(async () => second.resolve(payload()))
    await screen.findByText('Mac')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    fireEvent.click(screen.getByLabelText('浅色'))
    await vi.waitFor(() => expect(screen.getByRole('button', { name: /保\s*存/ }).disabled).toBe(false))
    await act(async () => first.resolve(payload()))
    expect(screen.getByLabelText('浅色').checked).toBe(true)
    expect(screen.getByRole('switch', { name: /提醒开启/ }).getAttribute('aria-checked')).toBe('true')
  })

  it.each([
    [{ status: 'available', reasonCode: 'available' }, '执行器群组已就绪', '已检测到在线执行器'],
    [{ status: 'degraded', reasonCode: 'offline' }, '执行器群组就绪状态降级', '已注册执行器均离线'],
    [{ status: 'unavailable', reasonCode: 'not_configured' }, '执行器群组不可用', '未配置已注册执行器'],
    [{ status: 'unavailable', reasonCode: 'status_unavailable' }, '执行器群组不可用', '暂时无法确定执行器就绪状态'],
    [undefined, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [null, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [{ status: 'mystery', reasonCode: '<b>secret</b>' }, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [{ available: true }, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [{ status: 'available', reasonCode: 'offline' }, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [{ status: 'unavailable', reasonCode: 'available' }, '执行器群组就绪状态未知', '暂时无法确定执行器就绪状态'],
    [{ status: 'degraded', reasonCode: 'secret-reason' }, '执行器群组就绪状态降级', '暂时无法确定执行器就绪状态'],
  ])('uses explicit server status safely: %j', async (executorStatus, heading, reason) => {
    getWorkspaceSettings.mockResolvedValue(payload({ executorStatus, executors: executorStatus?.reasonCode === 'not_configured' ? [] : [row] }))
    render(<SettingsPage onTheme={vi.fn()} />)
    expect(await screen.findByText('运行时就绪状态')).toBeTruthy()
    expect(screen.getByText('已选择配置')).toBeTruthy()
    expect(screen.getByLabelText('本地执行器优先').checked).toBe(true)
    expect(screen.getByText(heading)).toBeTruthy()
    expect(screen.getByText(reason)).toBeTruthy()
    expect(document.body.textContent).not.toMatch(/secret|实际执行器状态|实际生效/)
    if (executorStatus?.status !== 'available') expect(screen.queryByText('执行器群组已就绪')).toBeNull()
    if (executorStatus?.reasonCode === 'not_configured') expect(screen.getByText('没有已注册执行器；当前不能使用本地执行器。')).toBeTruthy()
    else expect(screen.getByText('离线')).toBeTruthy()
  })

  it('uses safe read failure copy and retries', async () => {
    getWorkspaceSettings.mockRejectedValueOnce({ message: 'private message', payload: 'secret', requestId: 'req-settings-read' }).mockResolvedValueOnce(payload())
    render(<SettingsPage />)
    expect(await screen.findByText('设置读取失败')).toBeTruthy()
    expect(screen.getByText(/req-settings-read/)).toBeTruthy()
    expect(document.body.textContent).not.toMatch(/private message|secret/)
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ }))
    expect(await screen.findByText('Mac')).toBeTruthy()
    expect(getWorkspaceSettings).toHaveBeenCalledTimes(2)
  })
})
