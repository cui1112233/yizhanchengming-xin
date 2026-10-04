import React from 'react'
import { createRoot } from 'react-dom/client'
import { App, Card, ConfigProvider, Typography } from 'antd'
import 'antd/dist/reset.css'

function AdminShell() {
  return (
    <ConfigProvider>
      <App>
        <main style={{ maxWidth: 960, margin: '48px auto', padding: '0 24px' }}>
          <Card>
            <Typography.Title level={2}>一战晟铭 · 管理端</Typography.Title>
            <Typography.Paragraph>
              新主线管理端已建立。业务功能将按迁移计划逐步接入 Go API。
            </Typography.Paragraph>
          </Card>
        </main>
      </App>
    </ConfigProvider>
  )
}

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <AdminShell />
  </React.StrictMode>,
)
