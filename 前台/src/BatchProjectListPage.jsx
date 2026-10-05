import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Descriptions, Drawer, Modal, Space, Switch, Table, Tag, Typography } from 'antd'
import {
  getAudioMeasurement,
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

export default function BatchProjectListPage() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState(null)
  const [summary, setSummary] = useState(null)
  const [generationLoading, setGenerationLoading] = useState(false)
  const [resultModal, setResultModal] = useState(null)
  const [audioMeasurements, setAudioMeasurements] = useState({})
  const [matchAudioByBook, setMatchAudioByBook] = useState({})

  useEffect(() => {
    let active = true
    listBatchProjects()
      .then((payload) => active && setProjects(payload?.projects || []))
      .catch((reason) => active && setError(reason instanceof Error ? reason.message : '读取批量项目失败'))
      .finally(() => active && setLoading(false))
    return () => { active = false }
  }, [])

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
      const nextSummary = await getProjectGeneration(project.id)
      setSummary(nextSummary)
      await refreshMeasurements(project, nextSummary?.books || [])
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
    setAudioMeasurements({})
    setMatchAudioByBook({})
    await refreshGeneration(project)
  }

  const runBatch = async () => {
    if (!selected) return
    setGenerationLoading(true)
    try {
      // Batch execution keeps Task 12 behavior unless every book is explicitly
      // executed with its own authoritative measurement. Per-book matchAudio is
      // therefore never inferred from client state.
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
    { title: '项目名称', dataIndex: 'name', key: 'name' },
    {
      title: '统一设置', key: 'settings', width: 420,
      render: (_, row) => <UnifiedSettingsPanel project={row} />,
    },
    {
      title: '生成', key: 'actions', width: 140,
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
        title: '单本操作', key: 'bookAction', width: 120,
        render: (_, row) => <Button type="primary" onClick={() => void runOne(row.bookId)}>单本执行</Button>,
      },
    ]
  }, [selected, audioMeasurements, matchAudioByBook])

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text>
          <Typography.Title level={2}>批量工厂</Typography.Title>
          <Typography.Paragraph type="secondary">
            BatchProject、统一设置与 Script / Hook / Director / Final Prompt 状态均来自 Go + MySQL 事实源。
          </Typography.Paragraph>
        </div>
      </div>

      {error && <Alert type="error" showIcon message={error} className="feedback" closable onClose={() => setError('')} />}

      <Card title="批量项目" className="result-card">
        <Table rowKey="id" columns={projectColumns} dataSource={projects} loading={loading} pagination={{ pageSize: 20, hideOnSinglePage: true }} locale={{ emptyText: '暂无批量项目' }} />
      </Card>

      <Drawer
        title={selected ? `${selected.name} · 剧本生成流水线` : '剧本生成流水线'}
        width="92vw"
        open={Boolean(selected)}
        onClose={() => { setSelected(null); setSummary(null); setAudioMeasurements({}); setMatchAudioByBook({}) }}
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
            <Table rowKey="bookId" columns={generationColumns} dataSource={summary.books || []} loading={generationLoading} pagination={false} scroll={{ x: 1280 }} />
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
            {resultModal.validationResult && <Typography.Text type="secondary">Validation: {resultModal.validationResult}</Typography.Text>}
            {resultModal.errorMessage && <Alert type="error" showIcon message={resultModal.errorMessage} />}
            <Typography.Paragraph copyable style={{ whiteSpace: 'pre-wrap' }}>{resultModal.outputText || '暂无输出'}</Typography.Paragraph>
          </Space>
        )}
      </Modal>
    </main>
  )
}
