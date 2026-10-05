import React from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import IntakeWorkbench from './App.jsx'

export function UserApp() {
  return (
    <ConfigProvider>
      <AntApp>
        <IntakeWorkbench />
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
