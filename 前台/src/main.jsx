import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, ConfigProvider, theme as antdTheme } from 'antd'
import 'antd/dist/reset.css'
import RouterApp from './RouterApp.jsx'
import AuthBoundary from './AuthBoundary.jsx'

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

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, theme)
    } catch {
      // 主题持久化失败不影响业务；禁止在这里保存项目、任务、脚本或历史数据。
    }
  }, [theme])

  return (
    <ConfigProvider
      theme={{
        algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: { colorPrimary: '#5b62d9', borderRadius: 10, controlHeight: 40, fontSize: 14 },
        components: { Card: { headerHeight: 52 }, Table: { cellPaddingBlock: 14 }, Drawer: { footerPaddingBlock: 16, footerPaddingInline: 24 } },
      }}
    >
      <AntApp>
        <AuthBoundary>
          <RouterApp theme={theme} onToggleTheme={() => setTheme((current) => current === 'dark' ? 'light' : 'dark')} />
        </AuthBoundary>
      </AntApp>
    </ConfigProvider>
  )
}

export function mountUserApp(container) {
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
