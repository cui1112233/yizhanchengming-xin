import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Descriptions, Drawer, Modal, Space, Table, Tag, Typography } from 'antd'
import {
  getBatchProject,
  getGenerationStage,
  getProjectGeneration,
  listBatchProjects,
  retryGenerationStage,
  runBookGeneration,
  runProjectGeneration,
} from './api.js'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'

const STAGES = [
  ['SCRIPT', 'Script'],
  ['HOOK', 'Hook'],
  ['DIRECTOR', 'Director'],
  ['FINAL_PROMPT', 'Final Prompt'],
]

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

function statusColor(status) {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'error'
  if (status === 'running') return 'processing'
  if (status === 'skipped') return 'default'
  return 'warning'
}

function statusLabel(status) {
  return ({ pending: '待执行', running: '执行中', completed: '完成', failed: '失败', skipped: '已跳过' })[status] || '待执行'
}

function renderTags(values) {
  const items = Array.isArray(values) ? values.filter(Boolean) : []
  if (items.length === 0) return '-'
  return (
    <Space size={[4, 4]} wrap>
      {items.map((value) => <Tag key={value}>{value}</Tag>)}
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
    return () => { active = false }
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
          scroll={{ x: 1100 }}
        />
      </Card>
    </main>
  )
}

export default function BatchProjectListPage() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedDetailProjectId, setSelectedDetailProjectId] = useState(null)
  const [selected, setSelected] = useState(null)
  const [summary, setSummary] = useState(null)
  const [generationLoading, setGenerationLoading] = useState(false)
  const [resultModal, setResultModal] = useState(null)

  useEffect(() => {
    let active = true
    listBatchProjects()
      .then((payload) => active && setProjects(payload?.projects || []))
      .catch((reason) => active && setError(reason instanceof Error ? reason.message : '读取批量项目失败'))
      .finally(() => active && setLoading(false))
    return () => { active = false }
  }, [])

  const refreshGeneration = async (project = selected) => {
    if (!project) return
    setGenerationLoading(true)
    try {
      setSummary(await getProjectGeneration(project.id))
      setError('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '读取生成状态失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const openGeneration = async (project) => {
    setSelected(project)
    setSummary(null)
    await refreshGeneration(project)
  }

  const runBatch = async () => {
    if (!selected) return
    setGenerationLoading(true)
    try {
      await runProjectGeneration(selected.id, {
        hookEnabled: true,
        plotMode: false,
        directorMode: 'normal',
        requestId: `batch-${Date.now()}`,
      })
      await refreshGeneration(selected)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '批量执行失败')
      await refreshGeneration(selected)
    } finally {
      setGenerationLoading(false)
    }
  }

  const runOne = async (bookId) => {
    if (!selected) return
    setGenerationLoading(true)
    try {
      await runBookGeneration(selected.id, bookId, {
        hookEnabled: true,
        plotMode: false,
        directorMode: 'normal',
        requestId: `book-${bookId}-${Date.now()}`,
      })
      await refreshGeneration(selected)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '单本执行失败')
      await refreshGeneration(selected)
    } finally {
      setGenerationLoading(false)
    }
  }

  const retryStage = async (bookId, stage) => {
    if (!selected) return
    setGenerationLoading(true)
    try {
      await retryGenerationStage(selected.id, bookId, stage, `retry-${bookId}-${stage}-${Date.now()}`)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '重试失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const showStage = async (bookId, stage) => {
    if (!selected) return
    try {
      const result = await getGenerationStage(selected.id, bookId, stage)
      setResultModal(result)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '读取 Stage 结果失败')
    }
  }

  const projectColumns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (value, project) => (
        <Button type="link" onClick={() => setSelectedDetailProjectId(project.id)}>
          {value || '未命名项目'}
        </Button>
      ),
    },
    { title: '书城来源', key: 'sources', render: (_, project) => renderTags(project.sources) },
    { title: '小说数量', dataIndex: 'bookCount', key: 'bookCount', render: (value) => value ?? 0 },
    { title: '男女频', key: 'genders', render: (_, project) => renderTags(project.genders) },
    { title: '风格', key: 'styles', render: (_, project) => renderTags(project.styles) },
    {
      title: '运行状态',
      dataIndex: 'runStatus',
      key: 'runStatus',
      render: (value) => (value ? <Tag>{RUN_STATUS_LABELS[value] || value}</Tag> : '-'),
    },
    {
      title: '统一设置',
      key: 'settings',
      width: 420,
      render: (_, row) => <UnifiedSettingsPanel project={row} />,
    },
    {
      title: '生成',
      key: 'actions',
      width: 140,
      render: (_, row) => <Button onClick={() => void openGeneration(row)}>生成状态</Button>,
    },
  ]

  const generationColumns = useMemo(() => {
    const stageColumns = STAGES.map(([key, label]) => ({
      title: label,
      key,
      width: 150,
      render: (_, row) => {
        const stage = row.stages?.[key]
        const status = stage?.status || 'pending'
        return (
          <Space direction="vertical" size={4}>
            <Tag color={statusColor(status)}>{statusLabel(status)}</Tag>
            {stage?.errorMessage && <Typography.Text type="danger">{stage.errorMessage}</Typography.Text>}
            {stage?.outputText && <Button size="small" onClick={() => void showStage(row.bookId, key)}>查看结果</Button>}
            {status === 'failed' && <Button size="small" danger onClick={() => void retryStage(row.bookId, key)}>重试 {label}</Button>}
          </Space>
        )
      },
    }))
    return [
      { title: 'Book ID', dataIndex: 'bookId', key: 'bookId', width: 90 },
      { title: '书名', dataIndex: 'title', key: 'title', width: 150, render: (value) => value || '-' },
      ...stageColumns,
      {
        title: '单本操作',
        key: 'bookAction',
        width: 120,
        render: (_, row) => <Button type="primary" onClick={() => void runOne(row.bookId)}>单本执行</Button>,
      },
    ]
  }, [selected])

  if (selectedDetailProjectId != null) {
    return <BatchProjectDetail projectId={selectedDetailProjectId} onBack={() => setSelectedDetailProjectId(null)} />
  }

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            BatchProject、小说汇总、统一设置与 Script / Hook / Director / Final Prompt 状态均来自 Go + MySQL 事实源。
          </Typography.Paragraph>
        </div>
      </div>

      {error && <Alert type="error" showIcon message={error} className="feedback" closable onClose={() => setError('')} />}

      <Card title="批量项目" className="result-card">
        <Table
          rowKey="id"
          columns={projectColumns}
          dataSource={projects}
          loading={loading}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '暂无批量项目' }}
          scroll={{ x: 1800 }}
        />
      </Card>

      <Drawer
        title={selected ? `${selected.name} · 剧本生成流水线` : '剧本生成流水线'}
        width="92vw"
        open={Boolean(selected)}
        onClose={() => { setSelected(null); setSummary(null) }}
        extra={<Button type="primary" loading={generationLoading} onClick={() => void runBatch()}>批量执行</Button>}
      >
        {summary && (
          <>
            <Descriptions bordered size="small" column={4} style={{ marginBottom: 16 }}>
              <Descriptions.Item label="完成">{summary.completed || 0}</Descriptions.Item>
              <Descriptions.Item label="运行中">{summary.running || 0}</Descriptions.Item>
              <Descriptions.Item label="失败">{summary.failed || 0}</Descriptions.Item>
              <Descriptions.Item label="待执行">{summary.pending || 0}</Descriptions.Item>
            </Descriptions>
            <Table rowKey="bookId" columns={generationColumns} dataSource={summary.books || []} loading={generationLoading} pagination={false} scroll={{ x: 1050 }} />
          </>
        )}
      </Drawer>

      <Modal
        title={resultModal ? `${resultModal.stage} · 执行结果` : '执行结果'}
        open={Boolean(resultModal)}
        onCancel={() => setResultModal(null)}
        footer={null}
        width={820}
      >
        {resultModal && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Typography.Text type="secondary">Prompt: {resultModal.promptKey || '-'} v{resultModal.promptVersion || '-'}</Typography.Text>
            {resultModal.errorMessage && <Alert type="error" showIcon message={resultModal.errorMessage} />}
            <Typography.Paragraph copyable style={{ whiteSpace: 'pre-wrap' }}>{resultModal.outputText || '暂无输出'}</Typography.Paragraph>
          </Space>
        )}
      </Modal>
    </main>
  )
}
