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

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <UserApp />
  </React.StrictMode>,
)
