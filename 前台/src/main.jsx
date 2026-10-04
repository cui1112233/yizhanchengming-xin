import React from 'react'
import { createRoot } from 'react-dom/client'
import { App, Card, ConfigProvider, Typography } from 'antd'
import 'antd/dist/reset.css'

function UserShell() {
  return (
    <ConfigProvider>
      <App>
        <main style={{ maxWidth: 960, margin: '48px auto', padding: '0 24px' }}>
          <Card>
            <Typography.Title level={2}>一战晟铭 · 用户端</Typography.Title>
            <Typography.Paragraph>
              新主线用户端已建立。第一阶段将接入小说获取与批量项目流程。
            </Typography.Paragraph>
          </Card>
        </main>
      </App>
    </ConfigProvider>
  )
}

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <UserShell />
  </React.StrictMode>,
)
