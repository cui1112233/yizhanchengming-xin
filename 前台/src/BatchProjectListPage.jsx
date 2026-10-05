import React, { useEffect, useState } from 'react'
import { Alert, Button, Card, Space, Table, Tag, Typography } from 'antd'
import { getBatchProject, listBatchProjects } from './api.js'

const RUN_STATUS_LABELS = {
  pending: '待执行',
  running: '执行中',
  completed: '已完成',
  failed: '失败',
}

const BOOK_STATUS_LABELS = {
  pending: '待获取',
  fetched: '已获取',
  retryable_failed: '可重试失败',
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

function BatchProjectDetail({ projectId, onBack }) {
  const [project, setProject] = useState(null)
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    setLoading(true)
    setError('')

    getBatchProject(projectId)
      .then((payload) => {
        if (!active) return
        setProject(payload?.project || null)
        setBooks(payload?.books || [])
      })
      .catch((reason) => {
        if (!active) return
        setError(reason instanceof Error ? reason.message : '读取批量项目详情失败')
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [projectId])

  const columns = [
    { title: 'Book ID', dataIndex: 'bookId', key: 'bookId', render: (value) => value || '-' },
    { title: '书名', dataIndex: 'title', key: 'title', render: (value) => value || '-' },
    { title: '书城', dataIndex: 'source', key: 'source', render: (value) => value || '-' },
    { title: 'platformId', dataIndex: 'platformId', key: 'platformId', render: (value) => value || '-' },
    { title: '男女频', dataIndex: 'gender', key: 'gender', render: (value) => value || '-' },
    { title: '风格', dataIndex: 'style', key: 'style', render: (value) => value || '-' },
    {
      title: '正文获取状态',
      dataIndex: 'status',
      key: 'status',
      render: (value) => <Tag>{BOOK_STATUS_LABELS[value] || value || '未知'}</Tag>,
    },
    { title: '错误信息', dataIndex: 'errorMessage', key: 'errorMessage', render: (value) => value || '-' },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">Batch Factory V11 工作台</Typography.Text>
          <Typography.Title level={2}>{project?.name || '批量项目'}</Typography.Title>
          <Typography.Paragraph type="secondary">
            项目与小说数据直接读取 Go API / MySQL，不使用浏览器缓存作为事实源。
          </Typography.Paragraph>
        </div>
        <Button onClick={onBack}>返回项目列表</Button>
      </div>

      {error && <Alert type="error" showIcon message={error} className="feedback" />}

      <Card title="小说列表" className="result-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={books}
          loading={loading}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '该项目暂无小说' }}
        />
      </Card>
    </main>
  )
}

export default function BatchProjectListPage() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedProjectId, setSelectedProjectId] = useState(null)

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

  if (selectedProjectId != null) {
    return <BatchProjectDetail projectId={selectedProjectId} onBack={() => setSelectedProjectId(null)} />
  }

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (value, project) => (
        <Button type="link" onClick={() => setSelectedProjectId(project.id)}>
          {value || '未命名项目'}
        </Button>
      ),
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
