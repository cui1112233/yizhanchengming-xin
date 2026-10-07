import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Drawer, Empty, Select, Space, Tag, Typography } from 'antd'
import { createShuihuoMediaTask, getBatchProject, listBatchProjects, listShuihuoAssets, listShuihuoCandidates, listShuihuoMediaTasks, listShuihuoSegments, retryShuihuoMediaTask, shuihuoAssetContentURL } from './api.js'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './tts-history.css'

const ACTIVE = new Set(['pending_executor', 'queued', 'running'])

function messageOf(error, fallback) { return error instanceof Error ? error.message : fallback }

export default function TtsPage() {
  const [projects, setProjects] = useState([])
  const [catalog, setCatalog] = useState([])
  const [selectedProject, setSelectedProject] = useState('')
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [status, setStatus] = useState('all')
  const [retrying, setRetrying] = useState(0)
  const [createOpen, setCreateOpen] = useState(false)
  const [createBook, setCreateBook] = useState('')
  const [createSegment, setCreateSegment] = useState('')
  const [segments, setSegments] = useState([])
  const [segmentLoading, setSegmentLoading] = useState(false)
  const [creating, setCreating] = useState(false)

  const load = useCallback(async (showRefresh = false) => {
    if (showRefresh) setRefreshing(true); else setLoading(true)
    setError('')
    try {
      const projectPayload = await listBatchProjects()
      const nextProjects = projectPayload?.projects || []
      setProjects(nextProjects)
      setSelectedProject((current) => current && nextProjects.some((item) => String(item.id) === current) ? current : (nextProjects[0] ? String(nextProjects[0].id) : ''))
      const details = await Promise.all(nextProjects.map(async (project) => {
        const detail = await getBatchProject(project.id)
        const books = detail?.books || []
        const media = await Promise.all(books.map(async (book) => {
          const [tasks, assets] = await Promise.all([listShuihuoMediaTasks(project.id, book.bookId), listShuihuoAssets(project.id, book.bookId)])
          const audioTasks = (tasks || []).filter((task) => task.kind === 'audio')
          const candidates = await Promise.all(audioTasks.map((task) => listShuihuoCandidates(project.id, book.bookId, task.id).catch(() => [])))
          const selectedAsset = new Map()
          audioTasks.forEach((task, index) => {
            const candidate = (candidates[index] || []).find((item) => item.selected) || (candidates[index] || [])[0]
            if (candidate) selectedAsset.set(task.id, candidate.assetId)
          })
          return audioTasks.map((task) => ({ ...task, project, book, assetId: selectedAsset.get(task.id) || 0, hasAsset: (assets || []).some((asset) => asset.id === selectedAsset.get(task.id)) }))
        }))
        return { project, books, tasks: media.flat() }
      }))
      setCatalog(details.map(({ project, books }) => ({ project, books })))
      setRows(details.flatMap((item) => item.tasks).sort((a, b) => String(b.updatedAt || b.createdAt).localeCompare(String(a.updatedAt || a.createdAt))))
    } catch (reason) { setError(messageOf(reason, '读取配音任务失败')) } finally { setLoading(false); setRefreshing(false) }
  }, [])

  useEffect(() => { void load() }, [load])
  useEffect(() => {
    if (!rows.some((task) => ACTIVE.has(task.status))) return undefined
    const timer = window.setInterval(() => { void load(true) }, 3000)
    return () => window.clearInterval(timer)
  }, [rows, load])

  const visible = useMemo(() => rows.filter((task) => (!selectedProject || String(task.project.id) === selectedProject) && (status === 'all' || task.status === status)), [rows, selectedProject, status])
  const retry = async (task) => {
    setRetrying(task.id); setError('')
    try { await retryShuihuoMediaTask(task.project.id, task.book.bookId, task.id); await load(true) } catch (reason) { setError(messageOf(reason, '重试配音失败')) } finally { setRetrying(0) }
  }
  const selectedCatalog = catalog.find((item) => String(item.project.id) === selectedProject)
  const selectedBook = selectedCatalog?.books.find((book) => String(book.bookId) === createBook)
  const openCreate = () => { setCreateBook(''); setCreateSegment(''); setSegments([]); setCreateOpen(true) }
  const loadSegments = async (bookId) => {
    setCreateBook(bookId); setCreateSegment(''); setSegments([])
    if (!selectedCatalog || !bookId) return
    setSegmentLoading(true)
    try { setSegments(await listShuihuoSegments(selectedCatalog.project.id, Number(bookId)) || []) } catch (reason) { setError(messageOf(reason, '读取可配音分镜失败')) } finally { setSegmentLoading(false) }
  }
  const create = async () => {
    if (!selectedCatalog || !selectedBook || !createSegment) return
    setCreating(true); setError('')
    try {
      await createShuihuoMediaTask(selectedCatalog.project.id, selectedBook.bookId, { kind: 'audio', segmentId: Number(createSegment), requestId: `tts-${selectedCatalog.project.id}-${selectedBook.bookId}-${Date.now()}` })
      setCreateOpen(false); await load(true)
    } catch (reason) { setError(messageOf(reason, '创建配音任务失败')) } finally { setCreating(false) }
  }

  if (loading) return <PageState state="loading" title="正在读取配音任务…" className="tts-history-state" />
  if (error && rows.length === 0) return <PageState state="failed" title="配音任务读取失败" description={error} onRetry={() => void load()} className="tts-history-state" />
  return <main className="tts-history-page">
    <section className="tts-history-heading"><div><Typography.Text type="secondary">一战晟铭 · 媒体任务</Typography.Text><Typography.Title level={2}>配音</Typography.Title><Typography.Paragraph type="secondary">从已保存分镜创建配音任务；刷新后状态、试听和下载均从 Go API / MySQL / TOS 恢复。</Typography.Paragraph></div><Space><Button onClick={() => void load(true)} loading={refreshing}>刷新</Button><Button type="primary" disabled={!selectedCatalog?.books.length} onClick={openCreate}>创建配音任务</Button></Space></section>
    {error && <Alert type="error" showIcon closable message={error} onClose={() => setError('')} />}
    <Card className="tts-history-filters"><Space wrap><Select aria-label="筛选项目" value={selectedProject} style={{ minWidth: 220 }} onChange={setSelectedProject} options={[{ value: '', label: '全部项目' }, ...projects.map((project) => ({ value: String(project.id), label: project.name || `项目 #${project.id}` }))]} /><Select aria-label="筛选配音状态" value={status} style={{ minWidth: 160 }} onChange={setStatus} options={[{ value: 'all', label: '全部状态' }, ...['pending_executor', 'queued', 'running', 'succeeded', 'failed', 'retryable_failed'].map((value) => ({ value, label: value }))]} /></Space></Card>
    {visible.length ? <div className="tts-task-grid">{visible.map((task) => <TtsTaskCard key={task.id} task={task} retrying={retrying === task.id} onRetry={() => void retry(task)} />)}</div> : <Empty className="tts-history-empty" description="暂无已持久化的配音任务。未配置 Provider 时系统不会伪造音频结果。" />}
    <Drawer title="从分镜创建配音" width={440} open={createOpen} onClose={() => setCreateOpen(false)} extra={<Button type="primary" loading={creating} disabled={!createSegment} onClick={() => void create()}>提交</Button>}><Alert type="info" showIcon message="Provider 与模型仅由服务端统一配置" description="此处只选择已保存分镜；分镜文本会作为 TTS 输入，密钥不会发送到浏览器。" /><Typography.Paragraph type="secondary" style={{ marginTop: 18 }}>项目：{selectedCatalog?.project.name || '-'}</Typography.Paragraph><Typography.Text>小说</Typography.Text><Select aria-label="选择配音小说" value={createBook || undefined} placeholder="选择小说" style={{ width: '100%', margin: '8px 0 16px' }} onChange={(value) => void loadSegments(value)} options={(selectedCatalog?.books || []).map((book) => ({ value: String(book.bookId), label: book.title || `小说 #${book.bookId}` }))} /><Typography.Text>分镜</Typography.Text><Select aria-label="选择配音分镜" value={createSegment || undefined} loading={segmentLoading} disabled={!createBook} placeholder="选择已保存分镜" style={{ width: '100%', marginTop: 8 }} onChange={setCreateSegment} options={segments.map((segment, index) => ({ value: String(segment.id), label: `分镜 ${index + 1} · ${String(segment.text || '').slice(0, 36)}` }))} />{createBook && !segmentLoading && !segments.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该小说暂无已保存分镜" />}</Drawer>
  </main>
}

function TtsTaskCard({ task, retrying, onRetry }) {
  const playable = task.status === 'succeeded' && task.assetId && task.hasAsset
  const source = playable ? shuihuoAssetContentURL(task.project.id, task.book.bookId, task.assetId) : ''
  return <Card className="tts-task-card" title={<Space><span>{task.book.title || `小说 #${task.book.bookId}`}</span><StatusTag status={task.status} /></Space>} extra={<Typography.Text type="secondary">#{task.id}</Typography.Text>}>
    <Typography.Paragraph type="secondary" className="tts-task-meta">{task.project.name || `项目 #${task.project.id}`} · Provider：{task.provider || '等待服务端分配'} · 模型：{task.model || '-'}</Typography.Paragraph>
    {task.errorMessage && <Alert type="error" showIcon message={task.errorMessage} />}
    {playable ? <audio controls preload="metadata" src={source} className="tts-audio-player" /> : <Typography.Paragraph type="secondary" className="tts-audio-unavailable">{task.status === 'succeeded' ? '音频候选资产不可用，请刷新后重试。' : '音频生成完成后可在此试听和下载。'}</Typography.Paragraph>}
    <Space wrap>{playable && <Button type="primary" href={source} download={`tts-task-${task.id}.audio`}>下载音频</Button>}{['failed', 'retryable_failed'].includes(task.status) && <Button danger loading={retrying} onClick={onRetry}>重试</Button>}{task.status === 'pending_executor' && <Tag color="gold">等待 TTS Provider 执行器</Tag>}</Space>
  </Card>
}
