import React, { useCallback, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import RouterApp from './RouterApp.jsx'
import AuthBoundary from './AuthBoundary.jsx'
import { getWorkspaceSettings } from './api.js'
import { applyThemeVariables, themeConfig } from './theme.js'

const THEME_STORAGE_KEY = 'yizhan-theme'

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
  const restoreServerTheme = useCallback(async () => {
    try {
      const payload = await getWorkspaceSettings()
      if (payload?.settings?.theme === 'dark' || payload?.settings?.theme === 'light') setTheme(payload.settings.theme)
    } catch {
      // The session has already been authenticated. A transient preferences read
      // failure keeps the permitted non-sensitive UI fallback without changing facts.
    }
  }, [])

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
        <AuthBoundary onAuthenticated={restoreServerTheme}>
          <RouterApp theme={theme} onToggleTheme={(nextTheme) => setTheme((current) => nextTheme === 'dark' || nextTheme === 'light' ? nextTheme : (current === 'dark' ? 'light' : 'dark'))} />
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
