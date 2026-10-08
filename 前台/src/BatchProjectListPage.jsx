import React, { useEffect, useMemo, useRef, useState } from 'react'
import { generationOutcomeMessage, generationValidationSummary } from './ui/generationOutcome.js'
import { Alert, Button, Card, Descriptions, Drawer, Modal, Space, Switch, Table, Tag, Typography } from 'antd'
import { cancelVideoTask, getAudioMeasurement, getBatchProject, getGenerationStage, getProjectGeneration, getProjectVideoStatus, restoreBatchProject, retryGenerationStage, retryVideoTask, runBookGeneration, runProjectGeneration } from './api.js'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'
import StatusTag from './ui/StatusTag.jsx'
import PageState from './ui/PageState.jsx'
import { batchError, batchUpdatedAt } from './batchFactoryPresentation.js'
import './batch-factory.css'

const STAGES = [['SCRIPT', 'Script'], ['HOOK', 'Hook'], ['DIRECTOR', 'Director'], ['FINAL_PROMPT', 'Final Prompt']]
function renderTags(values) {
  const items = Array.isArray(values) ? values.filter(Boolean) : []
  return items.length ? <Space size={[4, 4]} wrap>{items.map((value) => <Tag key={value}>{value}</Tag>)}</Space> : '-'
}

export default function BatchProjectListPage({ initialProjectId = null, onClearProject }) {
  const [project, setProject] = useState(null)
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [detailError, setDetailError] = useState(null)
  const [restoreError, setRestoreError] = useState(null)
  const [restoring, setRestoring] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [contentBook, setContentBook] = useState(null)
  const request = useRef(0)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState(null)
  const [summary, setSummary] = useState(null)
  const [videoStatus, setVideoStatus] = useState(null)
  const [generationLoading, setGenerationLoading] = useState(false)
  const [resultModal, setResultModal] = useState(null)
  const [audioMeasurements, setAudioMeasurements] = useState({})
  const [matchAudioByBook, setMatchAudioByBook] = useState({})
  const readonly = Boolean(project?.archivedAt)
  const refreshDetail = () => setRefreshKey((value) => value + 1)

  useEffect(() => {
    const id = ++request.current
    setLoading(true); setDetailError(null); setProject(null); setBooks([])
    setSelected(null); setSummary(null); setVideoStatus(null); setContentBook(null); setResultModal(null)
    if (!Number.isSafeInteger(Number(initialProjectId)) || Number(initialProjectId) <= 0) {
      setDetailError({ message: '项目入口无效' }); setLoading(false)
      return () => { request.current++ }
    }
    getBatchProject(initialProjectId).then((payload) => {
      if (id !== request.current) return
      if (!payload?.project?.id || Number(payload.project.id) !== Number(initialProjectId)) throw new Error('invalid project detail')
      setProject(payload.project)
      setBooks(Array.isArray(payload.books) ? payload.books : [])
    }).catch((reason) => {
      if (id === request.current) setDetailError(batchError(reason, '项目详情读取失败，请稍后重试。'))
    }).finally(() => { if (id === request.current) setLoading(false) })
    return () => { request.current++ }
  }, [initialProjectId, refreshKey])

  const restoreProject = async () => {
    if (!project || !readonly) return
    setRestoring(true); setRestoreError(null)
    try { await restoreBatchProject(project.id); refreshDetail() }
    catch (reason) { setRestoreError(batchError(reason, '恢复项目失败，请稍后重试。')) }
    finally { setRestoring(false) }
  }

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
      setError(batchError(reason).message || '读取生成状态失败')
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
    if (!selected || selected.archivedAt) return
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
      setError(batchError(reason).message || '批量执行失败')
      await refreshGeneration(selected)
    } finally {
      setGenerationLoading(false)
    }
  }

  const runOne = async (bookId) => {
    if (!selected || selected.archivedAt) return
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
      setError(batchError(reason).message || '单本执行失败')
      await refreshGeneration(selected)
    } finally {
      setGenerationLoading(false)
    }
  }

  const retryStage = async (bookId, stage) => {
    if (!selected || selected.archivedAt) return
    setGenerationLoading(true)
    try {
      await retryGenerationStage(selected.id, bookId, stage, `retry-${bookId}-${stage}-${Date.now()}`)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(batchError(reason).message || '重试失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const retryVideo = async (bookId, taskId) => {
    if (!selected || selected.archivedAt || !taskId) return
    setGenerationLoading(true)
    try {
      await retryVideoTask(taskId, `video-retry-${bookId}-${Date.now()}`)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(batchError(reason).message || 'VIDEO 重试失败')
    } finally {
      setGenerationLoading(false)
    }
  }

  const cancelVideo = async (taskId) => {
    if (!selected || selected.archivedAt || !taskId) return
    setGenerationLoading(true)
    try {
      await cancelVideoTask(taskId)
      await refreshGeneration(selected)
    } catch (reason) {
      setError(batchError(reason).message || 'VIDEO 取消失败')
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
      setError(batchError(reason).message || '读取 Stage 结果失败')
    }
  }

  const videoByBook = useMemo(() => {
    const items = Array.isArray(videoStatus?.books) ? videoStatus.books : []
    return new Map(items.map((item) => [String(item.bookId), item]))
  }, [videoStatus])

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
            {!readonly && status === 'failed' && <Button size="small" danger onClick={() => void retryStage(row.bookId, key)}>重试 {label}</Button>}
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
                disabled={readonly || !enabled}
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
              {errorMessage && <Typography.Text type="danger">视频生成失败</Typography.Text>}
              {outputUrl && <Button type="link" size="small" href={outputUrl} target="_blank" rel="noreferrer">查看视频</Button>}
              {!readonly && (status === 'failed' || status === 'cancelled') && latest?.id && (
                <Button size="small" danger onClick={() => void retryVideo(row.bookId, latest.id)}>重试 VIDEO</Button>
              )}
              {!readonly && (status === 'queued' || status === 'running') && latest?.id && (
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
        render: (_, row) => readonly ? null : <Button type="primary" onClick={() => void runOne(row.bookId)}>单本执行</Button>,
      },
    ]
  }, [selected, readonly, audioMeasurements, matchAudioByBook, videoByBook])


  const bookColumns = [
    { title: '序号', key: 'index', width: 65, render: (_, row, index) => index + 1 },
    { title: '小说', key: 'novel', width: 220, render: (_, row) => <Space direction="vertical" size={2}><Typography.Text>{row.title || '未命名小说'}</Typography.Text><Typography.Text type="secondary">{row.bookId || '-'}</Typography.Text><Typography.Text type="secondary">{row.source || '-'}</Typography.Text></Space> },
    { title: '内容', key: 'content', width: 260, render: (_, row) => <Typography.Paragraph ellipsis={{ rows: 3 }}>{row.originalText || '暂无正文'}</Typography.Paragraph> },
    { title: '风格', dataIndex: 'style', key: 'style', width: 120, render: (value) => value || '-' },
    { title: '男女频', dataIndex: 'gender', key: 'gender', width: 100, render: (value) => value || '-' },
    { title: '状态', dataIndex: 'status', key: 'status', width: 130, render: (value) => <StatusTag status={value} /> },
    { title: '操作', key: 'actions', width: 160, render: (_, row) => <Space direction="vertical"><Button disabled={!row.originalText} onClick={() => setContentBook(row)}>查看正文</Button>{!readonly && <Button onClick={() => void openGeneration(project)}>生产详情</Button>}</Space> },
  ]

  return <main className="page-shell batch-factory-page">
    <div className="page-heading"><div><Typography.Text type="secondary">Batch Factory V11 工作台</Typography.Text><Typography.Title level={2}>{project?.name || '批量项目'}</Typography.Title><Typography.Paragraph type="secondary">查看小说内容、生成阶段与生产进度。</Typography.Paragraph></div><Space wrap><Button onClick={refreshDetail}>刷新项目</Button><Button onClick={onClearProject}>返回项目列表</Button></Space></div>
    {detailError && <PageState state="failed" title={detailError.message} requestId={detailError.requestId} onRetry={refreshDetail} />}
    {loading && <div role="status">正在读取项目详情</div>}
    {restoreError && <Alert type="error" showIcon message={restoreError.message} description={restoreError.requestId ? `请求编号：${restoreError.requestId}` : undefined} className="feedback" />}
    {project && <>
      {readonly && <Alert type="warning" showIcon message="项目已归档，当前为只读查看。恢复后才能修改或执行。" action={<Button loading={restoring} onClick={() => void restoreProject()}>恢复项目</Button>} className="feedback" />}
      <Space wrap className="batch-project-detail-summary"><Typography.Text>小说数量：{books.length}</Typography.Text>{renderTags(project.sources || [...new Set(books.map((book) => book.source))])}{renderTags(project.genders || [...new Set(books.map((book) => book.gender))])}{renderTags(project.styles || [...new Set(books.map((book) => book.style))])}{project.runStatus && <StatusTag status={project.runStatus} />}<Typography.Text type="secondary">最近更新：{batchUpdatedAt(project.updatedAt || project.createdAt)}</Typography.Text></Space>
      {readonly && <Button className="batch-project-detail-actions" onClick={() => void openGeneration(project)}>查看生成状态</Button>}
      {!readonly && <Space wrap className="batch-project-detail-actions"><UnifiedSettingsPanel project={project} /><Button type="primary" onClick={() => void openGeneration(project)}>进入生产工作台</Button><Button onClick={() => void openGeneration(project)}>生成状态</Button></Space>}
      {error && <Alert type="error" showIcon message={error} className="feedback" />}
      <Card title="小说列表" className="result-card"><Table rowKey="id" columns={bookColumns} dataSource={books} pagination={false} locale={{ emptyText: '该项目暂无小说' }} scroll={{ x: 1055 }} /></Card>
    </>}
    <Modal title={contentBook?.title || '小说正文'} open={Boolean(contentBook)} onCancel={() => setContentBook(null)} footer={null}><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{contentBook?.originalText || '暂无正文'}</Typography.Paragraph></Modal>
      <Drawer
        title={selected ? `${selected.name} · 剧本与 VIDEO 流水线` : '剧本与 VIDEO 流水线'}
        width="92vw"
        open={Boolean(selected)}
        onClose={() => { setSelected(null); setSummary(null); setVideoStatus(null); setAudioMeasurements({}); setMatchAudioByBook({}) }}
        extra={readonly ? null : <Button type="primary" loading={generationLoading} onClick={() => void runBatch()}>批量执行</Button>}
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
}
