import React from 'react'
import { Button, Result } from 'antd'

export default function AgentReservedPage({ onNavigate }) {
  return (
    <main className="route-foundation-page">
      <p>Agent 将重新设计，当前不可用</p>
      <Result status="info" title="Agent 工作区待重新设计" subTitle="当前版本不执行 Agent 任务，也不会创建项目或调用 Provider。" extra={<Button onClick={() => onNavigate('/')}>返回首页</Button>} />
    </main>
  )
}
