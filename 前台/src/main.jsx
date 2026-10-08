import React, { useCallback, useEffect, useRef, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { Alert, App as AntApp, ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import RouterApp from './RouterApp.jsx'
import AuthBoundary from './AuthBoundary.jsx'
import { getWorkspaceSettings, saveWorkspaceSettings } from './api.js'
import { applyThemeVariables, themeConfig } from './theme.js'

const THEME_STORAGE_KEY = 'yizhan-theme'

const validTheme = value => value === 'dark' || value === 'light'
function completeSettings(value) {
  return value && validTheme(value.theme)
    && ['tos', 'local_executor'].includes(value.storagePreference)
    && typeof value.petId === 'string' && /^[a-z0-9_-]{1,64}$/.test(value.petId)
    && Number.isInteger(value.soundVolume) && value.soundVolume >= 0 && value.soundVolume <= 100
    && ['notificationsEnabled', 'petVisible', 'companionActive'].every(key => typeof value[key] === 'boolean')
}

function safeFailure(error, code, rollbackTheme) {
  const requestId = typeof error?.requestId === 'string' && /^[a-zA-Z0-9._:-]{1,128}$/.test(error.requestId) ? error.requestId : ''
  return { code, requestId, rollbackTheme }
}

const failureCopy = {
  THEME_READ_FAILED: '读取主题设置失败，请重试。',
  THEME_SAVE_FAILED: '保存主题失败，已恢复上次确认的主题。',
  SETTINGS_SAVE_FAILED: '保存设置失败，请重试。',
}

function readInitialTheme() {
  if (typeof window === 'undefined') return 'dark'
  try {
    return window.localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

export function UserApp() {
  const [theme, setTheme] = useState(readInitialTheme)
  const [themeError, setThemeError] = useState(null)
  const initialTheme = useRef(theme)
  const currentTheme = useRef(theme)
  const confirmedSettings = useRef(null)
  const readFlight = useRef(null)
  const saveQueue = useRef(Promise.resolve())
  const revision = useRef(0)
  const latestIntent = useRef(0)
  const pending = useRef(0)
  const accountIdentity = useRef(null)
  const authEpoch = useRef(0)

  const renderTheme = useCallback(value => {
    currentTheme.current = value
    setTheme(value)
  }, [])

  // GETs share one flight. A response started before a successful PUT cannot
  // replace its confirmed snapshot, including during auth restoration.
  const readConfirmedSettings = useCallback(() => {
    if (readFlight.current) return readFlight.current
    const startedRevision = revision.current
    const startedEpoch = authEpoch.current
    const request = getWorkspaceSettings().then(result => {
      if (startedEpoch !== authEpoch.current) throw new Error('Session changed')
      if (startedRevision !== revision.current) return { ...result, settings: confirmedSettings.current }
      if (!completeSettings(result?.settings)) throw new Error('Incomplete settings response')
      confirmedSettings.current = result.settings
      if (!pending.current) renderTheme(result.settings.theme)
      return result
    }).catch(error => {
      throw safeFailure(error, 'THEME_READ_FAILED', confirmedSettings.current?.theme || initialTheme.current)
    }).finally(() => { if (readFlight.current === request) readFlight.current = null })
    readFlight.current = request
    return request
  }, [renderTheme])

  const restoreServerTheme = useCallback(async user => {
    const identity = ['string', 'number'].includes(typeof user?.id) ? String(user.id) : null
    if (identity === null || identity !== accountIdentity.current) {
      accountIdentity.current = identity
      authEpoch.current += 1
      revision.current += 1
      latestIntent.current += 1
      confirmedSettings.current = null
      readFlight.current = null
      saveQueue.current = Promise.resolve()
      pending.current = 0
      setThemeError(null)
      renderTheme(initialTheme.current)
    }
    const startedRevision = revision.current
    const startedEpoch = authEpoch.current
    try {
      await readConfirmedSettings()
    } catch (error) {
      if (startedEpoch === authEpoch.current && startedRevision === revision.current && !pending.current) {
        renderTheme(confirmedSettings.current?.theme || initialTheme.current)
        setThemeError(error)
      }
    }
  }, [readConfirmedSettings, renderTheme])

  // Both shell/Radio theme changes and explicit full-form saves use the same
  // full-object PUT queue. Only complete server responses enter this cache.
  const changeTheme = useCallback((nextTheme, options = {}) => {
    const intent = ++latestIntent.current
    const startedEpoch = authEpoch.current
    const draft = options.saveSettings ? { ...options.saveSettings } : null
    const requestedTheme = draft?.theme || (validTheme(nextTheme) ? nextTheme : currentTheme.current === 'dark' ? 'light' : 'dark')
    if (!options.confirmedSettings && !draft) renderTheme(requestedTheme)
    setThemeError(null)
    pending.current += 1
    const transition = saveQueue.current.then(async () => {
      if (startedEpoch !== authEpoch.current) throw new Error('Session changed')
      let result
      if (options.confirmedSettings) {
        result = { settings: options.confirmedSettings }
      } else {
        if (!confirmedSettings.current) await readConfirmedSettings()
        if (startedEpoch !== authEpoch.current) throw new Error('Session changed')
        // The form owns non-theme drafts; queued theme transitions own theme.
        // A page draft captured before those transitions must not restore it.
        const settings = draft ? { ...confirmedSettings.current, ...draft, theme: confirmedSettings.current.theme } : { ...confirmedSettings.current, theme: requestedTheme }
        if (!completeSettings(settings)) throw new Error('Incomplete settings request')
        // A 401 replay could send this account's full body under a newly logged
        // in account. Epoch guards can reject responses, but cannot undo writes.
        result = await saveWorkspaceSettings(settings, { skipAuthRecovery: true })
      }
      if (startedEpoch !== authEpoch.current) throw new Error('Session changed')
      if (!completeSettings(result?.settings)) throw new Error('Incomplete settings response')
      confirmedSettings.current = result.settings
      revision.current += 1
      if (intent === latestIntent.current) renderTheme(result.settings.theme)
      return result
    }).catch(error => {
      const failure = safeFailure(error, error?.code === 'THEME_READ_FAILED' ? 'THEME_READ_FAILED' : draft ? 'SETTINGS_SAVE_FAILED' : 'THEME_SAVE_FAILED', confirmedSettings.current?.theme || initialTheme.current)
      if (startedEpoch === authEpoch.current && intent === latestIntent.current) {
        renderTheme(failure.rollbackTheme)
        setThemeError(failure)
      }
      throw failure
    }).finally(() => { if (startedEpoch === authEpoch.current) pending.current -= 1 })
    // Shell buttons fire and forget; Settings still receives this original
    // rejecting promise and can restore its Radio selection safely.
    saveQueue.current = transition.catch(() => {})
    return transition
  }, [readConfirmedSettings, renderTheme])

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    applyThemeVariables(document.documentElement, theme)
    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, theme)
    } catch {
      // 主题持久化失败不影响业务；禁止在这里保存项目、任务、脚本或历史数据。
    }
  }, [theme])

  return (
    <ConfigProvider
      theme={themeConfig(theme)}
    >
      <AntApp>
        {themeError && <Alert type="error" showIcon message={failureCopy[themeError.code]} description={themeError.requestId ? `请求 ID：${themeError.requestId}` : undefined} />}
        <AuthBoundary onAuthenticated={restoreServerTheme}>
          <RouterApp theme={theme} onToggleTheme={changeTheme} />
        </AuthBoundary>
      </AntApp>
    </ConfigProvider>
  )
}

export function mountUserApp(container) {
  container.classList.add('user-theme-root')
  applyThemeVariables(container.ownerDocument.documentElement, readInitialTheme())
  const root = createRoot(container)
  root.render(
    <React.StrictMode>
      <UserApp />
    </React.StrictMode>,
  )
  return root
}

export const appRoot = mountUserApp(document.getElementById('root'))

export { THEME_STORAGE_KEY }
