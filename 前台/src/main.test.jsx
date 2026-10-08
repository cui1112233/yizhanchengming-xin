// @vitest-environment jsdom

import { act, fireEvent, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const completeSettings = { theme: 'dark', notificationsEnabled: false, storagePreference: 'local_executor', petId: 'fox', soundVolume: 23, petVisible: false, companionActive: true }
const payload = (settings = completeSettings) => ({ settings, executors: [], executorStatus: { status: 'unavailable', reasonCode: 'not_configured' } })
const json = (body, status = 200, headers = {}) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json', ...headers } })
const deferred = () => { let resolve, reject; const promise = new Promise((yes, no) => { resolve = yes; reject = no }); return { promise, resolve, reject } }
let root
let reads
let writes

beforeEach(() => {
  vi.resetModules()
  document.body.innerHTML = '<div id="root"></div>'
  window.localStorage.setItem('unrelated-sentinel', 'preserved')
  reads = vi.spyOn(Storage.prototype, 'getItem')
  writes = vi.spyOn(Storage.prototype, 'setItem')
})

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

afterEach(async () => {
  if (root) await act(async () => root.unmount())
  root = null
  await settleReactAfterUnmount()
  expect(reads.mock.calls.every(([key]) => key === 'yizhan-theme')).toBe(true)
  expect(writes.mock.calls.every(([key]) => key === 'yizhan-theme')).toBe(true)
  expect([...reads.mock.instances, ...writes.mock.instances].every(storage => storage === window.localStorage)).toBe(true)
  vi.restoreAllMocks()
  expect(window.localStorage.getItem('unrelated-sentinel')).toBe('preserved')
  window.history.replaceState({}, '', '/')
  window.localStorage.clear()
  document.body.replaceChildren()
})

async function boot(settingsRequest = async () => json(payload()), path = '/', loginUserId = 7) {
  window.history.replaceState({}, '', path)
  global.fetch = vi.fn(async (url, options = {}) => {
    if (url === '/api/auth/current-user' || url === '/api/auth/login') return json({ user: { id: url === '/api/auth/login' ? loginUserId : 7, name: 'Test User', role: 'admin' } })
    if (url === '/api/v1/workspace/settings') return settingsRequest(options)
    return json({})
  })
  await act(async () => { root = (await import('./main.jsx')).appRoot })
  await screen.findByRole('button', { name: '主题', exact: true })
}
const toggle = () => fireEvent.click(screen.getByRole('button', { name: '主题', exact: true }))
const puts = () => global.fetch.mock.calls.filter(([url, options]) => url === '/api/v1/workspace/settings' && options?.method === 'PUT').map(([, options]) => JSON.parse(options.body))
const expectTheme = (theme) => {
  expect(document.documentElement.dataset.theme).toBe(theme)
  expect(window.localStorage.getItem('yizhan-theme')).toBe(theme)
}

async function restoreSession() {
  act(() => window.dispatchEvent(new CustomEvent('ycm:auth-unauthenticated')))
  fireEvent.change(await screen.findByLabelText('用户名'), { target: { value: 'test-user' } })
  fireEvent.change(screen.getByLabelText('密码'), { target: { value: 'test-password' } })
  fireEvent.click(screen.getByRole('button', { name: '登录并进入工作台' }))
  await screen.findByRole('button', { name: '主题', exact: true })
}

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
        return new Response(JSON.stringify(payload({ ...completeSettings, theme: 'light' })), { status: 200, headers: { 'Content-Type': 'application/json' } })
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

  it.each([['dark', 'light'], ['light', 'dark']])('server theme overrides %s fallback with %s', async (fallback, serverTheme) => {
    window.localStorage.setItem('yizhan-theme', fallback)
    const read = deferred()
    await boot(() => read.promise)
    expectTheme(fallback)
    await act(async () => read.resolve(json(payload({ ...completeSettings, theme: serverTheme }))))
    expectTheme(serverTheme)
  })

  it('renders optimistically and preserves all confirmed preferences in the PUT', async () => {
    const save = deferred()
    await boot(options => options.method === 'PUT' ? save.promise : json(payload()))
    toggle()
    expectTheme('light')
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }]))
    await act(async () => save.resolve(json(payload({ ...completeSettings, theme: 'light' }))))
    expectTheme('light')
  })

  it('adopts normalized responses and merges subsequent toggles from that complete response', async () => {
    const normalized = { ...completeSettings, soundVolume: 47, petId: 'default' }
    await boot(options => options.method === 'PUT' ? json(payload(normalized)) : json(payload()))
    toggle()
    await vi.waitFor(() => expectTheme('dark'))
    toggle()
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }, { ...normalized, theme: 'light' }]))
  })

  it('rolls back failed saves, shows only fixed copy plus request ID, and retains the confirmed snapshot', async () => {
    const save = deferred()
    await boot(options => options.method === 'PUT' ? save.promise : json(payload()))
    toggle()
    expectTheme('light')
    await act(async () => save.resolve(json({ code: 'SETTINGS_SAVE_FAILED', message: '<b>private token</b>', secret: 'sensitive-payload' }, 500, { 'X-Request-ID': 'req-theme-7' })))
    await screen.findByText(/保存主题失败/)
    expectTheme('dark')
    expect(screen.getByText(/req-theme-7/)).toBeTruthy()
    expect(document.body.textContent).not.toMatch(/private token|sensitive-payload/)
    toggle()
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }, { ...completeSettings, theme: 'light' }]))
  })

  it('waits for a complete pending read before saving the optimistic theme', async () => {
    const read = deferred()
    await boot(options => options.method === 'PUT' ? json(payload({ ...completeSettings, theme: 'light' })) : read.promise)
    toggle()
    expectTheme('light')
    expect(puts()).toEqual([])
    await act(async () => read.resolve(json(payload())))
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }]))
    expectTheme('light')
    expect(global.fetch.mock.calls.filter(([url, options]) => url === '/api/v1/workspace/settings' && options?.method !== 'PUT')).toHaveLength(1)
  })

  it.each(['rejected', 'incomplete'])('never writes from a %s initial snapshot', async (mode) => {
    const read = deferred()
    window.localStorage.setItem('yizhan-theme', 'light')
    await boot(() => read.promise)
    toggle()
    expectTheme('dark')
    await act(async () => read.resolve(mode === 'rejected' ? json({ message: 'sensitive-read' }, 500, { 'X-Request-ID': 'req-read-8' }) : json({ settings: { theme: 'dark' } })))
    await screen.findByText(/读取主题设置失败/)
    expectTheme('light')
    expect(puts()).toEqual([])
    expect(document.body.textContent).not.toContain('sensitive-read')
    if (mode === 'rejected') expect(screen.getByText(/req-read-8/)).toBeTruthy()
  })

  it('serializes rapid toggles and merges the second PUT from the first normalized response', async () => {
    const first = deferred(), second = deferred()
    let saves = 0
    await boot(options => options.method === 'PUT' ? (++saves === 1 ? first.promise : second.promise) : json(payload()))
    toggle()
    await vi.waitFor(() => expect(puts()).toHaveLength(1))
    toggle()
    expectTheme('dark')
    expect(puts()).toHaveLength(1)
    const normalized = { ...completeSettings, theme: 'light', soundVolume: 48 }
    await act(async () => first.resolve(json(payload(normalized))))
    expectTheme('dark')
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }, { ...normalized, theme: 'dark' }]))
    await act(async () => second.resolve(json(payload({ ...normalized, theme: 'dark' }))))
    expectTheme('dark')
  })

  it('orders full-form Save behind a shell theme PUT and adopts its normalized response exactly once', async () => {
    const first = deferred(), second = deferred()
    let saves = 0
    await boot(options => options.method === 'PUT' ? (++saves === 1 ? first.promise : second.promise) : json(payload()), '/settings')
    await screen.findByLabelText('浅色')
    fireEvent.click(screen.getByRole('switch', { name: /提醒关闭/ }))
    toggle()
    await vi.waitFor(() => expect(puts()).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: /保\s*存/ }))
    expect(puts()).toHaveLength(1)
    await act(async () => first.resolve(json(payload({ ...completeSettings, theme: 'light' }))))
    await vi.waitFor(() => expect(puts()).toHaveLength(2))
    expect(puts()[1]).toEqual({ ...completeSettings, notificationsEnabled: true })
    await act(async () => second.resolve(json(payload({ ...completeSettings, theme: 'light', notificationsEnabled: true, soundVolume: 44 }))))
    expectTheme('light')
    expect(screen.getByLabelText('浅色').checked).toBe(true)
    expect(puts()).toHaveLength(2)
  })

  it.each(['stale', 'failed'])('ignores a %s auth restore arriving after a newer normalized PUT', async mode => {
    const restore = deferred(), save = deferred()
    let gets = 0
    await boot(options => options.method === 'PUT' ? save.promise : ++gets === 1 ? json(payload()) : restore.promise)
    toggle()
    await vi.waitFor(() => expect(puts()).toHaveLength(1))
    await restoreSession()
    expect(gets).toBe(2)
    const normalized = { ...completeSettings, theme: 'light', soundVolume: 51 }
    await act(async () => save.resolve(json(payload(normalized))))
    expectTheme('light')
    await act(async () => restore.resolve(mode === 'stale' ? json(payload()) : json({ message: 'private-restore' }, 500)))
    expectTheme('light')
    expect(screen.queryByText(/读取主题设置失败/)).toBeNull()
    toggle()
    await vi.waitFor(() => expect(puts()[1]).toEqual({ ...normalized, theme: 'dark' }))
  })

  it('keeps the previously confirmed theme and preferences if a later auth restore fails', async () => {
    window.localStorage.setItem('yizhan-theme', 'light')
    let gets = 0
    await boot(options => options.method === 'PUT' ? json(payload({ ...completeSettings, theme: 'light' })) : ++gets === 1 ? json(payload()) : json({ message: 'private-restore' }, 500, { 'X-Request-ID': 'restore-failed' }))
    expectTheme('dark')
    await restoreSession()
    await screen.findByText(/读取主题设置失败/)
    expectTheme('dark')
    toggle()
    await vi.waitFor(() => expect(puts()).toEqual([{ ...completeSettings, theme: 'light' }]))
  })

  it('a prior rejected request does not undo a newer optimistic intent or stop the save queue', async () => {
    const first = deferred(), second = deferred(), third = deferred()
    let saves = 0
    await boot(options => options.method === 'PUT' ? [first, second, third][saves++].promise : json(payload()))
    toggle()
    await vi.waitFor(() => expect(puts()).toHaveLength(1))
    toggle()
    toggle()
    expectTheme('light')
    await act(async () => first.resolve(json({ message: 'private-old-save' }, 500)))
    await vi.waitFor(() => expect(puts()).toHaveLength(2))
    expectTheme('light')
    expect(screen.queryByText(/保存主题失败/)).toBeNull()
    expect(puts()[1]).toEqual(completeSettings)
    await act(async () => second.resolve(json(payload({ ...completeSettings, soundVolume: 52 }))))
    await vi.waitFor(() => expect(puts()).toHaveLength(3))
    expectTheme('light')
    expect(puts()[2]).toEqual({ ...completeSettings, soundVolume: 52, theme: 'light' })
    await act(async () => third.resolve(json(payload({ ...completeSettings, soundVolume: 52, theme: 'light' }))))
    expectTheme('light')
  })

  it('invalidates account A pending saves and queued writes when account B authenticates', async () => {
    const saveA = deferred(), readB = deferred()
    const accountB = { ...completeSettings, theme: 'light', petId: 'default', soundVolume: 73, notificationsEnabled: true }
    let gets = 0, saves = 0
    await boot(options => options.method === 'PUT' ? ++saves === 1 ? saveA.promise : json(payload(accountB)) : ++gets === 1 ? json(payload()) : readB.promise, '/', 8)
    toggle()
    await vi.waitFor(() => expect(puts()).toHaveLength(1))
    toggle()
    await restoreSession()
    await act(async () => readB.resolve(json(payload(accountB))))
    expectTheme('light')
    await act(async () => saveA.resolve(json(payload({ ...completeSettings, theme: 'dark', soundVolume: 14 }))))
    expectTheme('light')
    expect(puts()).toHaveLength(1)
    toggle()
    await vi.waitFor(() => expect(puts()[1]).toEqual({ ...accountB, theme: 'dark' }))
  })

  it('does not share account A pending restore with account B or adopt its late response', async () => {
    const readA = deferred(), readB = deferred()
    const accountB = { ...completeSettings, theme: 'light', soundVolume: 74 }
    let gets = 0
    await boot(options => options.method === 'PUT' ? json(payload(accountB)) : ++gets === 1 ? readA.promise : readB.promise, '/', 8)
    await restoreSession()
    expect(gets).toBe(2)
    await act(async () => readB.resolve(json(payload(accountB))))
    expectTheme('light')
    await act(async () => readA.resolve(json(payload())))
    expectTheme('light')
    toggle()
    await vi.waitFor(() => expect(puts()).toEqual([{ ...accountB, theme: 'dark' }]))
  })
})
