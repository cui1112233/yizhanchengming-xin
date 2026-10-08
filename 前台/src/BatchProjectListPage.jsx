import React, { useEffect, useMemo, useState } from 'react'
import { generationOutcomeMessage, generationValidationSummary } from './ui/generationOutcome.js'
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
import StatusTag from './ui/StatusTag.jsx'

const STAGES = [
  ['SCRIPT', 'Script'],
  ['HOOK', 'Hook'],
  ['DIRECTOR', 'Director'],
  ['FINAL_PROMPT', 'Final Prompt'],
]

function renderTags(values) {
  const items = Array.isArray(values) ? values.filter(Boolean) : []
  if (items.length === 0) return '-'
  return (
    <Space size={[4, 4]} wrap>
      {items.map((value) => <Tag key={value}>{value}</Tag>)}
    </Space>
  )
}

function BatchProjectDetail({ projectId, onBack, onOpenProduction }) {
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
      render: (value) => <StatusTag status={value} />,
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
        <Space>
          <Button type="primary" onClick={() => onOpenProduction(project)}>进入生产工作台</Button>
          <Button onClick={onBack}>返回项目列表</Button>
        </Space>
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

export default function BatchProjectListPage({ initialProjectId = null, onClearProject }) {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedDetailProjectId, setSelectedDetailProjectId] = useState(null)
  const [entryError, setEntryError] = useState('')
  const [selected, setSelected] = useState(null)
  const [summary, setSummary] = useState(null)
  const [videoStatus, setVideoStatus] = useState(null)
  const [generationLoading, setGenerationLoading] = useState(false)
  const [resultModal, setResultModal] = useState(null)
  const [audioMeasurements, setAudioMeasurements] = useState({})
  const [matchAudioByBook, setMatchAudioByBook] = useState({})

  useEffect(() => {
    let active = true
    setLoading(true)
    setSelectedDetailProjectId(null)
    setEntryError('')
    listBatchProjects()
      .then((payload) => {
        if (!active) return
        const visibleProjects = Array.isArray(payload?.projects) ? payload.projects : []
        setProjects(visibleProjects)
        if (initialProjectId != null) {
          if (visibleProjects.some((project) => Number(project.id) === Number(initialProjectId))) setSelectedDetailProjectId(initialProjectId)
          else setEntryError('项目不可访问')
        }
      })
      .catch((reason) => active && setError(reason instanceof Error ? reason.message : '读取批量项目失败'))
      .finally(() => active && setLoading(false))
    return () => { active = false }
  }, [initialProjectId])

  const refreshMeasurements = async (project, books = []) => {
    if (!project) return
    const values = await Promise.all((books || []).map(async (book) => {
      try {
        const measurement = await getAudioMeasurement(project.id, book.bookId)
        return [book.bookId, measurement?.durationMs > 0 ? measurement : null]
      } catch {
        return [book.bookId, null]
      }
    }))
    setAudioMeasurements(Object.fromEntries(values))
    setMatchAudioByBook((current) => {
      const next = { ...current }
      for (const [bookId, measurement] of values) {
        if (!measurement) next[bookId] = false
      }
      return next
    })
  }

  const refreshGeneration = async (project = selected) => {
    if (!project) return
    setGenerationLoading(true)
    try {
      const [generationPayload, videoPayload] = await Promise.all([
        getProjectGeneration(project.id),
        getProjectVideoStatus(project.id),
      ])
      setSummary(generationPayload)
      setVideoStatus(videoPayload)
      await refreshMeasurements(project, generationPayload?.books || [])
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
    setVideoStatus(null)
    setAudioMeasurements({})
    setMatchAudioByBook({})
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
        matchAudio: false,
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
    const measurement = audioMeasurements[bookId]
    const matchAudio = Boolean(matchAudioByBook[bookId] && measurement?.durationMs > 0)
    setGenerationLoading(true)
    try {
      await runBookGeneration(selected.id, bookId, {
        hookEnabled: true,
        plotMode: false,
        directorMode: 'normal',
        matchAudio,
        audioDurationSec: matchAudio ? measurement.durationMs / 1000 : undefined,
        shotDurationLimitSec: matchAudio ? 15 : undefined,
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

  const retryVideo = async (bookId, taskId) => {
    if (!selected || !taskId) return
    setGenerationLoading(true)
    try {
      await retryVideoTask(taskId, `video-retry-${bookId}-${Date.now()}`)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'VIDEO 重试失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const cancelVideo = async (taskId) => {
    if (!selected || !taskId) return
    setGenerationLoading(true)
    try {
      await cancelVideoTask(taskId)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'VIDEO 取消失败')
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

  const videoByBook = useMemo(() => {
    const items = Array.isArray(videoStatus?.books) ? videoStatus.books : []
    return new Map(items.map((item) => [String(item.bookId), item]))
  }, [videoStatus])

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
      render: (value) => (value ? <StatusTag status={value} /> : '-'),
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
            <StatusTag status={status} />
            {stage?.errorMessage && <Typography.Text type="danger">{generationOutcomeMessage(stage)}</Typography.Text>}
            {stage?.outputText && <Button size="small" onClick={() => void showStage(row.bookId, key)}>查看结果</Button>}
            {status === 'failed' && <Button size="small" danger onClick={() => void retryStage(row.bookId, key)}>重试 {label}</Button>}
          </Space>
        )
      },
    }))
    return [
      { title: 'Book ID', dataIndex: 'bookId', key: 'bookId', width: 90 },
      { title: '书名', dataIndex: 'title', key: 'title', width: 150, render: (value) => value || '-' },
      {
        title: '匹配音频',
        key: 'matchAudio',
        width: 230,
        render: (_, row) => {
          const measurement = audioMeasurements[row.bookId]
          const enabled = Boolean(measurement?.durationMs > 0)
          const checked = Boolean(enabled && matchAudioByBook[row.bookId])
          return (
            <Space direction="vertical" size={4}>
              <Switch
                aria-label={`匹配音频 Book ${row.bookId}`}
                disabled={!enabled}
                checked={checked}
                checkedChildren="匹配音频"
                unCheckedChildren="匹配音频"
                onChange={(value) => setMatchAudioByBook((current) => ({ ...current, [row.bookId]: value }))}
              />
              {enabled
                ? <Typography.Text type="secondary">已检测音频：{(measurement.durationMs / 1000).toFixed(2)} 秒</Typography.Text>
                : <Typography.Text type="warning">请先生成或检测音频</Typography.Text>}
              {checked && <Typography.Text type="success">最终分镜总时长将严格匹配音频时长</Typography.Text>}
            </Space>
          )
        },
      },
      ...stageColumns,
      {
        title: 'VIDEO',
        key: 'video',
        width: 280,
        render: (_, row) => {
          const video = videoByBook.get(String(row.bookId))
          if (!video) return <Typography.Text type="secondary">未生成</Typography.Text>
          const attempts = Array.isArray(video.attempts) ? video.attempts : []
          const latest = attempts.length > 0 ? attempts[attempts.length - 1] : null
          const status = latest?.status || video.status || 'queued'
          const errorMessage = latest?.errorMessage || video.errorMessage || ''
          const outputUrl = latest?.outputUrl || video.outputUrl || ''
          return (
            <Space direction="vertical" size={4}>
              <Typography.Text>{video.provider || '-'}</Typography.Text>
              <Typography.Text type="secondary">{video.model || '-'}</Typography.Text>
              <StatusTag status={status} />
              <Typography.Text type="secondary">尝试 {attempts.length} 次</Typography.Text>
              {errorMessage && <Typography.Text type="danger">{errorMessage}</Typography.Text>}
              {outputUrl && <Button type="link" size="small" href={outputUrl} target="_blank" rel="noreferrer">查看视频</Button>}
              {(status === 'failed' || status === 'cancelled') && latest?.id && (
                <Button size="small" danger onClick={() => void retryVideo(row.bookId, latest.id)}>重试 VIDEO</Button>
              )}
              {(status === 'queued' || status === 'running') && latest?.id && (
                <Button size="small" onClick={() => void cancelVideo(latest.id)}>取消 VIDEO</Button>
              )}
            </Space>
          )
        },
      },
      {
        title: '单本操作',
        key: 'bookAction',
        width: 120,
        render: (_, row) => <Button type="primary" onClick={() => void runOne(row.bookId)}>单本执行</Button>,
      },
    ]
  }, [selected, audioMeasurements, matchAudioByBook, videoByBook])

  if (selectedDetailProjectId != null) {
    return <BatchProjectDetail
      projectId={selectedDetailProjectId}
      onBack={() => { setSelectedDetailProjectId(null); onClearProject?.() }}
      onOpenProduction={(project) => {
        if (!project) return
        setSelectedDetailProjectId(null)
        void openGeneration(project)
      }}
    />
  }

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            BatchProject、小说汇总、统一设置以及 Script / Hook / Director / Final Prompt / VIDEO 状态均来自 Go + MySQL 事实源。
          </Typography.Paragraph>
        </div>
      </div>

      {entryError && <Alert type="warning" showIcon message={entryError} description="该项目不在当前账号可见范围内；已保留可访问项目列表。" className="feedback" />}
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
        title={selected ? `${selected.name} · 剧本与 VIDEO 流水线` : '剧本与 VIDEO 流水线'}
        width="92vw"
        open={Boolean(selected)}
        onClose={() => { setSelected(null); setSummary(null); setVideoStatus(null); setAudioMeasurements({}); setMatchAudioByBook({}) }}
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
            <Table rowKey="bookId" columns={generationColumns} dataSource={summary.books || []} loading={generationLoading} pagination={false} scroll={{ x: 1680 }} />
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
            {generationValidationSummary(resultModal.validationResult) && <Typography.Text type="secondary">{generationValidationSummary(resultModal.validationResult)}</Typography.Text>}
            {resultModal.errorMessage && <Alert type="error" showIcon message={generationOutcomeMessage(resultModal)} />}
            <Typography.Paragraph copyable style={{ whiteSpace: 'pre-wrap' }}>{resultModal.outputText || '暂无输出'}</Typography.Paragraph>
          </Space>
        )}
      </Modal>
    </main>
  )
}
