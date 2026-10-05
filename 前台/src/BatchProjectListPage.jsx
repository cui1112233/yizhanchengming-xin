import React, { useEffect, useState } from 'react'
import { Alert, Card, Space, Table, Tag, Typography } from 'antd'
import { listBatchProjects } from './api.js'

const RUN_STATUS_LABELS = {
  pending: '待执行',
  running: '执行中',
  completed: '已完成',
  failed: '失败',
}

function renderTags(values) {
  const items = Array.isArray(values) ? values.filter(Boolean) : []
  if (items.length === 0) return '-'
  return (
    <Space size={[4, 4]} wrap>
      {items.map((value) => (
        <Tag key={value}>{value}</Tag>
      ))}
    </Space>
  )
}

export default function BatchProjectListPage() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true

    listBatchProjects()
      .then((payload) => {
        if (!active) return
        setProjects(payload?.projects || [])
      })
      .catch((reason) => {
        if (!active) return
        setError(reason instanceof Error ? reason.message : '读取批量项目失败')
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [])

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '书城来源',
      key: 'sources',
      render: (_, project) => renderTags(project.sources),
    },
    {
      title: '小说数量',
      dataIndex: 'bookCount',
      key: 'bookCount',
      render: (value) => value ?? 0,
    },
    {
      title: '男女频',
      key: 'genders',
      render: (_, project) => renderTags(project.genders),
    },
    {
      title: '风格',
      key: 'styles',
      render: (_, project) => renderTags(project.styles),
    },
    {
      title: '运行状态',
      dataIndex: 'runStatus',
      key: 'runStatus',
      render: (value) => (value ? <Tag>{RUN_STATUS_LABELS[value] || value}</Tag> : '-'),
    },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            这里展示由小说获取工作台创建的真实 BatchProject，以及书城来源、小说数量、男女频、风格和最新运行状态。
          </Typography.Paragraph>
        </div>
      </div>

      {error && <Alert type="error" showIcon message={error} className="feedback" />}

      <Card title="批量项目" className="result-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={projects}
          loading={loading}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '暂无批量项目' }}
        />
      </Card>
    </main>
  )
}
