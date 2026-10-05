import React, { useEffect, useState } from 'react'
import { Alert, Card, Table, Typography } from 'antd'
import { listBatchProjects } from './api.js'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'

export default function BatchProjectListPage() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    listBatchProjects()
      .then((payload) => { if (active) setProjects(payload?.projects || []) })
      .catch((reason) => { if (active) setError(reason instanceof Error ? reason.message : '读取批量项目失败') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])

  const columns = [
    { title: '项目名称', dataIndex: 'name', key: 'name' },
    {
      title: '统一设置',
      key: 'settings',
      width: 420,
      render: (_, project) => <UnifiedSettingsPanel project={project} />,
    },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            当前页保留 BatchProject 上下文；生产统一设置、发布统一设置和版本对应配置档均在右侧 Drawer 内完成。
          </Typography.Paragraph>
        </div>
      </div>
      {error && <Alert type="error" showIcon message={error} className="feedback" />}
      <Card title="批量项目" className="result-card">
        <Table rowKey="id" columns={columns} dataSource={projects} loading={loading}
          pagination={{ pageSize: 20, hideOnSinglePage: true }} locale={{ emptyText: '暂无批量项目' }} />
      </Card>
    </main>
  )
}
