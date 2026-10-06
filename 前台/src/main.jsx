import React from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import IntakeWorkbench from './App.jsx'
import AuthBoundary from './AuthBoundary.jsx'

export function UserApp() {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: '#5b62d9', borderRadius: 10, controlHeight: 40, fontSize: 14 }, components: { Card: { headerHeight: 52 }, Table: { cellPaddingBlock: 14 }, Drawer: { footerPaddingBlock: 16, footerPaddingInline: 24 } } }}>
      <AntApp>
        <AuthBoundary>
          <IntakeWorkbench />
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
