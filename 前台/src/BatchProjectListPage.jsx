import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Descriptions, Drawer, Modal, Space, Switch, Table, Tag, Typography } from 'antd'
import {
  cancelVideoTask,
  getAudioMeasurement,
  getBatchProject,
  getGenerationStage,
  getProjectGeneration,
  getProjectVideoStatus,
  listBatchProjects,
  retryGenerationStage,
  retryVideoTask,
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
  succeeded: '成功',
  partial_failed: '部分失败',
  failed: '失败',
}

const BOOK_STATUS_LABELS = {
  pending: '待获取',
  fetched: '已获取',
  retryable_failed: '可重试失败',
}

function statusColor(status) {
  if (status === 'completed' || status === 'succeeded') return 'success'
  if (status === 'failed') return 'error'
  if (status === 'running') return 'processing'
  if (status === 'skipped' || status === 'cancelled') return 'default'
  return 'warning'
}

function statusLabel(status) {
  return ({ pending: '待执行', queued: '排队中', running: '执行中', completed: '完成', succeeded: '成功', partial_failed: '部分失败', failed: '失败', cancelled: '已取消', skipped: '已跳过' })[status] || '待执行'
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
          pagination={false}
          scroll={{ x: 1050 }}
          locale={{ emptyText: loading ? '加载中…' : '该项目暂无小说' }}
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
  const [generationOpen, setGenerationOpen] = useState(false)
  const [generationProject, setGenerationProject] = useState(null)
  const [generation, setGeneration] = useState(null)
  const [generationLoading, setGenerationLoading] = useState(false)
  const [generationError, setGenerationError] = useState('')
  const [busyKey, setBusyKey] = useState('')
  const [matchAudioByBook, setMatchAudioByBook] = useState({})
  const [audioByBook, setAudioByBook] = useState({})
  const [videoByBook, setVideoByBook] = useState({})
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [settingsProjectId, setSettingsProjectId] = useState(null)

  const loadProjects = async () => {
    setLoading(true)
    setError('')
    try {
      const payload = await listBatchProjects()
      setProjects(payload?.projects || [])
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '读取批量项目列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadProjects() }, [])

  const openGeneration = async (project) => {
    setGenerationOpen(true)
    setGenerationProject(project)
    setGeneration(null)
    setGenerationError('')
    setGenerationLoading(true)
    try {
      const payload = await getProjectGeneration(project.id)
      setGeneration(payload)
      const audioResults = await Promise.all((payload?.books || []).map(async (book) => {
        try {
          const measurement = await getAudioMeasurement(project.id, book.bookId)
          return [book.bookId, measurement]
        } catch {
          return [book.bookId, null]
        }
      }))
      setAudioByBook(Object.fromEntries(audioResults))
      const video = await getProjectVideoStatus(project.id).catch(() => ({ books: [] }))
      const videoEntries = (video?.books || []).map((book) => [book.bookId, book])
      setVideoByBook(Object.fromEntries(videoEntries))
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '读取生成状态失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const refreshGeneration = async () => {
    if (!generationProject) return
    await openGeneration(generationProject)
  }

  const executeProject = async () => {
    if (!generationProject) return
    setBusyKey('batch')
    setGenerationError('')
    try {
      await runProjectGeneration(generationProject.id, { requestId: `ui-batch-${Date.now()}` })
      await refreshGeneration()
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '批量执行失败')
    } finally {
      setBusyKey('')
    }
  }

  const executeBook = async (bookId) => {
    if (!generationProject) return
    const audio = audioByBook[bookId]
    const matchAudio = Boolean(matchAudioByBook[bookId])
    setBusyKey(`book-${bookId}`)
    setGenerationError('')
    try {
      await runBookGeneration(generationProject.id, bookId, {
        requestId: `ui-book-${bookId}-${Date.now()}`,
        matchAudio,
        ...(matchAudio && audio?.durationMs ? { audioDurationSec: audio.durationMs / 1000 } : {}),
        shotDurationLimitSec: 15,
      })
      await refreshGeneration()
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '单本执行失败')
    } finally {
      setBusyKey('')
    }
  }

  const retryStage = async (bookId, stage) => {
    if (!generationProject) return
    setBusyKey(`retry-${bookId}-${stage}`)
    setGenerationError('')
    try {
      await retryGenerationStage(generationProject.id, bookId, stage, { requestId: `ui-retry-${bookId}-${stage}-${Date.now()}` })
      await refreshGeneration()
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '重试失败')
    } finally {
      setBusyKey('')
    }
  }

  const retryVideo = async (task) => {
    if (!task?.taskId) return
    setBusyKey(`video-retry-${task.taskId}`)
    setGenerationError('')
    try {
      await retryVideoTask(task.taskId, {})
      await refreshGeneration()
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '视频重试失败')
    } finally {
      setBusyKey('')
    }
  }

  const cancelVideo = async (task) => {
    if (!task?.taskId) return
    setBusyKey(`video-cancel-${task.taskId}`)
    setGenerationError('')
    try {
      await cancelVideoTask(task.taskId)
      await refreshGeneration()
    } catch (reason) {
      setGenerationError(reason instanceof Error ? reason.message : '取消视频失败')
    } finally {
      setBusyKey('')
    }
  }

  if (selectedProjectId) {
    return <BatchProjectDetail projectId={selectedProjectId} onBack={() => setSelectedProjectId(null)} />
  }

  const columns = [
    {
      title: '项目',
      dataIndex: 'name',
      key: 'name',
      render: (value, row) => <Button type="link" onClick={() => setSelectedProjectId(row.id)}>{value || `项目 #${row.id}`}</Button>,
    },
    { title: '书城来源', dataIndex: 'sources', key: 'sources', render: renderTags },
    { title: '小说数量', dataIndex: 'bookCount', key: 'bookCount' },
    { title: '男女频', dataIndex: 'genders', key: 'genders', render: renderTags },
    { title: '风格', dataIndex: 'styles', key: 'styles', render: renderTags },
    {
      title: 'Run 状态',
      dataIndex: 'runStatus',
      key: 'runStatus',
      render: (value) => <Tag color={statusColor(value)}>{RUN_STATUS_LABELS[value] || value || '-'}</Tag>,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, row) => (
        <Space wrap>
          <Button onClick={() => void openGeneration(row)}>生成状态</Button>
          <Button onClick={() => { setSettingsProjectId(row.id); setSettingsOpen(true) }}>统一设置</Button>
        </Space>
      ),
    },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Phase 1</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            这里只读取服务端 BatchProject / Run / Book 数据，不使用浏览器缓存模拟生产状态。
          </Typography.Paragraph>
        </div>
      </div>
      {error && <Alert type="error" showIcon message={error} className="feedback" />}
      <Card title="BatchProject 列表" className="result-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={projects}
          loading={loading}
          pagination={false}
          locale={{ emptyText: loading ? '加载中…' : '暂无批量项目' }}
        />
      </Card>

      <Modal
        open={generationOpen}
        title={generationProject ? `${generationProject.name || `项目 #${generationProject.id}`} · 生成状态` : '生成状态'}
        onCancel={() => setGenerationOpen(false)}
        footer={null}
        width={980}
        destroyOnHidden
      >
        {generationError && <Alert type="error" showIcon message={generationError} className="feedback" />}
        <Space wrap style={{ marginBottom: 16 }}>
          <Button type="primary" loading={busyKey === 'batch'} onClick={() => void executeProject()}>批量执行</Button>
          <Button loading={generationLoading} onClick={() => void refreshGeneration()}>刷新</Button>
        </Space>
        <Table
          rowKey="bookId"
          loading={generationLoading}
          pagination={false}
          dataSource={generation?.books || []}
          scroll={{ x: 1400 }}
          columns={[
            { title: 'Book', dataIndex: 'bookId', key: 'bookId', width: 90 },
            ...STAGES.map(([stage, label]) => ({
              title: label,
              key: stage,
              width: 220,
              render: (_, book) => {
                const value = book.stages?.[stage] || {}
                return (
                  <Space direction="vertical" size={4}>
                    <Tag color={statusColor(value.status)}>{statusLabel(value.status)}</Tag>
                    {value.errorMessage ? <Typography.Text type="danger">{value.errorMessage}</Typography.Text> : null}
                    {value.status === 'failed' ? (
                      <Button
                        size="small"
                        loading={busyKey === `retry-${book.bookId}-${stage}`}
                        onClick={() => void retryStage(book.bookId, stage)}
                      >
                        重试 {label}
                      </Button>
                    ) : null}
                  </Space>
                )
              },
            })),
            {
              title: '匹配音频',
              key: 'matchAudio',
              width: 260,
              render: (_, book) => {
                const measurement = audioByBook[book.bookId]
                const enabled = Boolean(matchAudioByBook[book.bookId])
                return (
                  <Space direction="vertical" size={4}>
                    <Switch
                      aria-label={`匹配音频 Book ${book.bookId}`}
                      checked={enabled}
                      disabled={!measurement?.durationMs}
                      onChange={(checked) => setMatchAudioByBook((current) => ({ ...current, [book.bookId]: checked }))}
                    />
                    {measurement?.durationMs ? (
                      <Typography.Text type="secondary">已检测音频：{(measurement.durationMs / 1000).toFixed(2)} 秒</Typography.Text>
                    ) : (
                      <Typography.Text type="warning">请先生成或检测音频</Typography.Text>
                    )}
                    {enabled ? <Typography.Text type="secondary">最终分镜总时长将严格匹配音频时长</Typography.Text> : null}
                  </Space>
                )
              },
            },
            {
              title: '操作',
              key: 'bookAction',
              width: 120,
              fixed: 'right',
              render: (_, book) => (
                <Button
                  loading={busyKey === `book-${book.bookId}`}
                  onClick={() => void executeBook(book.bookId)}
                >
                  单本执行
                </Button>
              ),
            },
            {
              title: 'VIDEO',
              key: 'video',
              width: 280,
              render: (_, book) => {
                const current = videoByBook[book.bookId]
                const task = current?.tasks?.[0]
                if (!task) return <Typography.Text type="secondary">暂无视频任务</Typography.Text>
                return (
                  <Space direction="vertical" size={4}>
                    <Tag color={statusColor(task.status)}>{task.status}</Tag>
                    {task.errorMessage ? <Typography.Text type="danger">{task.errorMessage}</Typography.Text> : null}
                    <Space>
                      <Button
                        size="small"
                        disabled={task.status !== 'failed'}
                        loading={busyKey === `video-retry-${task.taskId}`}
                        onClick={() => void retryVideo(task)}
                      >视频重试</Button>
                      <Button
                        size="small"
                        danger
                        disabled={!['queued', 'submitting', 'submitted', 'processing', 'cancelling'].includes(task.status)}
                        loading={busyKey === `video-cancel-${task.taskId}`}
                        onClick={() => void cancelVideo(task)}
                      >取消视频</Button>
                    </Space>
                  </Space>
                )
              },
            },
          ]}
        />
      </Modal>

      <Drawer
        open={settingsOpen}
        title="生产 / 发布统一设置"
        placement="right"
        width="min(760px, 92vw)"
        destroyOnHidden
        onClose={() => { setSettingsOpen(false); setSettingsProjectId(null) }}
      >
        {settingsProjectId ? <UnifiedSettingsPanel projectId={settingsProjectId} /> : null}
      </Drawer>
    </main>
  )
}
