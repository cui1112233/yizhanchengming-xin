import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Col, Descriptions, Drawer, Empty, Form, Input, Modal, Progress, Row, Space, Statistic, Table, Typography, Tag, Tabs } from 'antd'
import { cancelVideoTask, getGenerationStage, getProjectGeneration, getProjectVideoStatus, listBatchProjects, retryGenerationStage, retryVideoTask, runBookGeneration, runProjectGeneration, startVideoTask, listShuihuoSegments, updateShuihuoSegment, reorderShuihuoSegments, listShuihuoAssets, listShuihuoMediaTasks, listShuihuoCandidates, selectShuihuoCandidate, retryShuihuoMediaTask } from './api.js'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './shuihuo-production.css'
import './shuihuo-media.css'

const STAGES = ['SCRIPT', 'HOOK', 'DIRECTOR', 'FINAL_PROMPT']
const stageLabel = { SCRIPT: '剧本', HOOK: '黄金三秒', DIRECTOR: '导演分镜', FINAL_PROMPT: '视频提示词' }

function errorText(reason, fallback) {
  return reason instanceof Error ? reason.message : fallback
}

function newestTask(video) {
  const attempts = Array.isArray(video?.attempts) ? video.attempts : []
  return attempts.at(-1) || null
}

export default function ShuihuoProductionPage() {
  const [projects, setProjects] = useState([])
  const [selected, setSelected] = useState(null)
  const [summary, setSummary] = useState(null)
  const [videos, setVideos] = useState(null)
  const [loading, setLoading] = useState(true)
  const [working, setWorking] = useState(false)
  const [error, setError] = useState('')
  const [stageResult, setStageResult] = useState(null)
  const [videoTarget, setVideoTarget] = useState(null)
  const [videoForm] = Form.useForm()
  const [workbook, setWorkbook] = useState(null)
  const [segments, setSegments] = useState([])
  const [assets, setAssets] = useState([])
  const [mediaTasks, setMediaTasks] = useState([])
  const [candidates, setCandidates] = useState({})
  const [mediaLoading, setMediaLoading] = useState(false)
  const [conflict, setConflict] = useState('')

  const refreshProjects = useCallback(async () => {
    const payload = await listBatchProjects()
    const next = payload?.projects || []
    setProjects(next)
    setSelected((current) => next.find((item) => item.id === current?.id) || current || next[0] || null)
    return next
  }, [])

  const refreshProject = useCallback(async (project = selected) => {
    if (!project?.id) return
    const [generation, video] = await Promise.all([getProjectGeneration(project.id), getProjectVideoStatus(project.id)])
    setSummary(generation)
    setVideos(video)
    setError('')
  }, [selected])

  const refreshMedia = useCallback(async (project = selected, book = workbook) => {
    if (!project?.id || !book?.bookId) return
    setMediaLoading(true)
    try {
      const [nextSegments, nextAssets, nextTasks] = await Promise.all([
        listShuihuoSegments(project.id, book.bookId), listShuihuoAssets(project.id, book.bookId), listShuihuoMediaTasks(project.id, book.bookId),
      ])
      setSegments(nextSegments || []); setAssets(nextAssets || []); setMediaTasks(nextTasks || []); setCandidates({}); setConflict('')
    } catch (reason) { setError(errorText(reason, '读取分镜媒体状态失败')) } finally { setMediaLoading(false) }
  }, [selected, workbook])

  useEffect(() => {
    const next = (summary?.books || [])[0] || null
    setWorkbook((current) => (summary?.books || []).find((row) => row.bookId === current?.bookId) || next)
  }, [summary])
  useEffect(() => { if (workbook) void refreshMedia(selected, workbook) }, [selected, workbook?.bookId, refreshMedia])

  useEffect(() => {
    let alive = true
    setLoading(true)
    refreshProjects().then((next) => next[0] && refreshProject(next[0])).catch((reason) => alive && setError(errorText(reason, '读取生产项目失败'))).finally(() => alive && setLoading(false))
    return () => { alive = false }
  }, [refreshProjects, refreshProject])

  useEffect(() => {
    if (!selected?.id) return undefined
    const hasInFlight = [...(summary?.books || []), ...(videos?.books || [])].some((book) => {
      const task = newestTask(book)
      return ['queued', 'running'].includes(task?.status || book?.status)
    })
    if (!hasInFlight) return undefined
    const timer = window.setInterval(() => { void refreshProject(selected).catch(() => {}) }, 2500)
    return () => window.clearInterval(timer)
  }, [selected, summary, videos, refreshProject])

  const run = async (bookId = null) => {
    if (!selected) return
    setWorking(true)
    try {
      const input = { hookEnabled: true, plotMode: false, directorMode: 'normal', matchAudio: false, requestId: `shuihuo-${bookId || 'batch'}-${Date.now()}` }
      if (bookId) await runBookGeneration(selected.id, bookId, input)
      else await runProjectGeneration(selected.id, input)
      await refreshProject(selected)
    } catch (reason) { setError(errorText(reason, '提交生产任务失败')) } finally { setWorking(false) }
  }

  const retryStage = async (bookId, stage) => {
    setWorking(true)
    try { await retryGenerationStage(selected.id, bookId, stage, `shuihuo-retry-${bookId}-${stage}-${Date.now()}`); await refreshProject(selected) } catch (reason) { setError(errorText(reason, '重试失败')) } finally { setWorking(false) }
  }

  const taskAction = async (task, action) => {
    setWorking(true)
    try {
      if (action === 'retry') await retryVideoTask(task.id, `shuihuo-video-retry-${Date.now()}`)
      else await cancelVideoTask(task.id)
      await refreshProject(selected)
    } catch (reason) { setError(errorText(reason, '视频任务操作失败')) } finally { setWorking(false) }
  }

  const openVideo = (row) => { videoForm.resetFields(); setVideoTarget(row) }
  const submitVideo = async () => {
    const input = await videoForm.validateFields()
    setWorking(true)
    try {
      await startVideoTask(selected.id, videoTarget.bookId, { ...input, requestId: `shuihuo-video-${videoTarget.bookId}-${Date.now()}` })
      setVideoTarget(null)
      await refreshProject(selected)
    } catch (reason) { setError(errorText(reason, '提交视频任务失败')) } finally { setWorking(false) }
  }

  const videoByBook = useMemo(() => new Map((videos?.books || []).map((item) => [String(item.bookId), item])), [videos])
  const rows = summary?.books || []
  const completed = Number(summary?.completed || 0)
  const total = rows.length
  const columns = [
    { title: '小说', key: 'book', width: 180, render: (_, row) => <Space direction="vertical" size={0}><Typography.Text strong>{row.title || `Book ${row.bookId}`}</Typography.Text><Typography.Text type="secondary">#{row.bookId}</Typography.Text></Space> },
    ...STAGES.map((stage) => ({ title: stageLabel[stage], key: stage, width: 150, render: (_, row) => { const item = row.stages?.[stage]; return <Space direction="vertical" size={4}><StatusTag status={item?.status || 'pending'} />{item?.errorMessage && <Typography.Text type="danger" ellipsis={{ tooltip: item.errorMessage }}>{item.errorMessage}</Typography.Text>}{item?.outputText && <Button size="small" type="link" onClick={async () => { try { setStageResult(await getGenerationStage(selected.id, row.bookId, stage)) } catch (reason) { setError(errorText(reason, '读取阶段结果失败')) } }}>查看</Button>}{item?.status === 'failed' && <Button size="small" danger onClick={() => void retryStage(row.bookId, stage)}>重试</Button>}</Space> }})),
    { title: '视频', key: 'video', width: 220, render: (_, row) => { const video = videoByBook.get(String(row.bookId)); const task = newestTask(video); if (!task) return <Button size="small" disabled={row.stages?.FINAL_PROMPT?.status !== 'completed'} onClick={() => openVideo(row)}>提交视频</Button>; return <Space direction="vertical" size={4}><StatusTag status={task.status || video.status} />{task.errorMessage && <Typography.Text type="danger" ellipsis={{ tooltip: task.errorMessage }}>{task.errorMessage}</Typography.Text>}{task.outputUrl && <Button size="small" type="link" href={task.outputUrl} target="_blank" rel="noreferrer">查看结果</Button>}{['failed', 'cancelled'].includes(task.status) && <Button size="small" danger onClick={() => void taskAction(task, 'retry')}>重试视频</Button>}{['queued', 'running'].includes(task.status) && <Button size="small" onClick={() => void taskAction(task, 'cancel')}>取消</Button>}</Space> } },
    { title: '操作', key: 'action', fixed: 'right', width: 110, render: (_, row) => <Button type="primary" size="small" loading={working} onClick={() => void run(row.bookId)}>继续执行</Button> },
  ]

  const saveSegment = async (segment, text) => {
    setWorking(true); setConflict('')
    try {
      await updateShuihuoSegment(selected.id, workbook.bookId, segment.id, { text, editRevision: `ui-${Date.now()}`, version: segment.version })
      await refreshMedia()
    } catch (reason) {
      if (reason?.status === 409 || reason?.code === 'version_conflict') setConflict('分镜已被其他操作更新。请先刷新后再保存，未保存内容不会被静默覆盖。')
      else setError(errorText(reason, '保存分镜失败'))
    } finally { setWorking(false) }
  }
  const reorder = async (from, to) => {
    const next = [...segments]; const [entry] = next.splice(from, 1); next.splice(to, 0, entry)
    setWorking(true)
    try { await reorderShuihuoSegments(selected.id, workbook.bookId, next.map((item) => item.id)); await refreshMedia() } catch (reason) { setError(errorText(reason, '保存分镜顺序失败')) } finally { setWorking(false) }
  }
  const loadCandidates = async (task) => { try { const out = await listShuihuoCandidates(selected.id, workbook.bookId, task.id); setCandidates((old) => ({ ...old, [task.id]: out || [] })) } catch (reason) { setError(errorText(reason, '读取候选结果失败')) } }
  const selectCandidate = async (task, candidate) => { setWorking(true); try { await selectShuihuoCandidate(selected.id, workbook.bookId, task.id, candidate.id); await loadCandidates(task) } catch (reason) { setError(errorText(reason, '选择主候选失败')) } finally { setWorking(false) } }
  const retryMedia = async (task) => { setWorking(true); try { await retryShuihuoMediaTask(selected.id, workbook.bookId, task.id); await refreshMedia() } catch (reason) { setError(errorText(reason, '媒体任务重试失败')) } finally { setWorking(false) } }

  return <main className="shuihuo-production-page">
    <section className="shuihuo-hero"><div><Typography.Text className="eyebrow">一战晟铭 · 生产工作台</Typography.Text><Typography.Title level={2}>水货生产</Typography.Title><Typography.Paragraph>从已入库小说继续执行剧本、分镜、视频生产；状态与恢复均以 Go API 和 MySQL 为准。</Typography.Paragraph></div><Space wrap><Button onClick={() => { setLoading(true); void refreshProjects().then(() => refreshProject()).catch((reason) => setError(errorText(reason, '刷新失败'))).finally(() => setLoading(false)) }}>刷新项目</Button><Button type="primary" loading={working} disabled={!selected} onClick={() => void run()}>批量继续执行</Button></Space></section>
    {error && <Alert className="shuihuo-alert" type="error" showIcon closable message={error} onClose={() => setError('')} />}{conflict && <Alert className="shuihuo-alert" type="warning" showIcon message={conflict} action={<Button size="small" onClick={() => void refreshMedia()}>刷新分镜</Button>} />}
    <Row gutter={[16, 16]}><Col xs={24} lg={7}><Card title="生产项目" className="shuihuo-projects" loading={loading}>{projects.length ? <Space direction="vertical" size={10} style={{ width: '100%' }}>{projects.map((project) => <button key={project.id} type="button" className={`shuihuo-project ${selected?.id === project.id ? 'is-selected' : ''}`} onClick={() => { setSelected(project); setSummary(null); setVideos(null); void refreshProject(project).catch((reason) => setError(errorText(reason, '读取项目状态失败'))) }}><span><b>{project.name || '未命名项目'}</b><small>#{project.id} · {project.bookCount || 0} 本小说</small></span><StatusTag status={project.runStatus || 'pending'} /></button>)}</Space> : <Empty description="暂无可生产项目，请先在小说获取或批量工厂创建项目。" />}</Card></Col><Col xs={24} lg={17}><Card className="shuihuo-console" title={selected ? `${selected.name} · 生产控制台` : '生产控制台'} extra={selected && <UnifiedSettingsPanel project={selected} />}>{selected ? <><Descriptions size="small" column={{ xs: 2, sm: 4 }}><Descriptions.Item label="完成"><Statistic value={completed} /></Descriptions.Item><Descriptions.Item label="执行中"><Statistic value={summary?.running || 0} /></Descriptions.Item><Descriptions.Item label="失败"><Statistic value={summary?.failed || 0} valueStyle={{ color: '#cf1322' }} /></Descriptions.Item><Descriptions.Item label="待执行"><Statistic value={summary?.pending || 0} /></Descriptions.Item></Descriptions><Progress percent={total ? Math.round(completed / total * 100) : 0} showInfo={false} status={summary?.failed ? 'exception' : 'active'} /><Table className="shuihuo-table" rowKey="bookId" columns={columns} dataSource={rows} loading={loading || working} pagination={false} scroll={{ x: 1320 }} locale={{ emptyText: '项目尚无可执行的小说记录' }} /></> : <Empty description="请选择一个项目" />}</Card></Col></Row>
    {selected && <Card className="shuihuo-storyboard" title="分镜与素材工作台" extra={<Space><Button onClick={() => void refreshMedia()} loading={mediaLoading}>刷新</Button></Space>}>
      {workbook ? <Tabs activeKey={String(workbook.bookId)} items={(summary?.books || []).map((book) => ({ key: String(book.bookId), label: book.title || `Book ${book.bookId}`, children: <div className="shuihuo-media-layout">
        <section><Typography.Title level={4}>分镜编辑</Typography.Title>{segments.length ? segments.map((segment, index) => <SegmentEditor key={segment.id} segment={segment} index={index} total={segments.length} busy={working} onSave={saveSegment} onMove={reorder} />) : <Empty description="暂无已保存分镜；当前不会把生成文本伪造成分镜。" />}</section>
        <section><Typography.Title level={4}>资产与任务</Typography.Title><AssetList assets={assets} /><MediaTaskList tasks={mediaTasks} candidates={candidates} busy={working} onCandidates={loadCandidates} onSelect={selectCandidate} onRetry={retryMedia} /></section>
      </div> }))} onChange={(key) => setWorkbook((summary?.books || []).find((book) => String(book.bookId) === key) || null)} /> : <Empty description="项目没有可用小说" />}
    </Card>}
    <Modal title={stageResult ? `${stageLabel[stageResult.stage] || stageResult.stage} · 执行结果` : '执行结果'} open={Boolean(stageResult)} onCancel={() => setStageResult(null)} footer={null} width={820}><Typography.Paragraph type="secondary">提示词：{stageResult?.promptKey || '-'} v{stageResult?.promptVersion || '-'}</Typography.Paragraph><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{stageResult?.outputText || '暂无输出'}</Typography.Paragraph></Modal>
    <Drawer title="提交视频任务" open={Boolean(videoTarget)} onClose={() => setVideoTarget(null)} width={420} extra={<Button type="primary" loading={working} onClick={() => void submitVideo()}>提交</Button>}><Alert type="info" showIcon message="仅提交已完成视频提示词的小说" description="提供方与模型由现有视频配置决定；服务端会校验配置、权限与最终提示词。" /><Form form={videoForm} layout="vertical" style={{ marginTop: 20 }}><Form.Item label="提供方" name="provider" rules={[{ required: true, message: '请输入已配置的视频提供方' }]}><Input placeholder="例如：yfai" /></Form.Item><Form.Item label="模型" name="model" rules={[{ required: true, message: '请输入已启用的视频模型' }]}><Input placeholder="例如：seedance-1.0-pro" /></Form.Item></Form></Drawer>
  </main>
}

function SegmentEditor({ segment, index, total, busy, onSave, onMove }) {
  const [text, setText] = useState(segment.text)
  useEffect(() => setText(segment.text), [segment.id, segment.text, segment.version])
  return <article className="shuihuo-segment"><header><b>分镜 {index + 1}</b><Tag>v{segment.version}</Tag><Space><Button size="small" disabled={busy || index === 0} onClick={() => onMove(index, index - 1)}>上移</Button><Button size="small" disabled={busy || index + 1 === total} onClick={() => onMove(index, index + 1)}>下移</Button></Space></header><Input.TextArea aria-label={`分镜 ${index + 1} 文本`} value={text} onChange={(event) => setText(event.target.value)} autoSize={{ minRows: 3, maxRows: 10 }} /><footer><Typography.Text type="secondary">编辑版本：{segment.editRevision || '-'}</Typography.Text><Button type="primary" size="small" loading={busy} disabled={!text.trim() || text === segment.text} onClick={() => onSave(segment, text)}>保存</Button></footer></article>
}
function AssetList({ assets }) { return <div className="shuihuo-assets"><b>已登记资产</b>{assets.length ? assets.map((asset) => <div key={asset.id}><Tag color={asset.status === 'ready' ? 'green' : asset.status === 'failed' ? 'red' : 'gold'}>{asset.type}</Tag><span>{asset.objectKey || `资产 #${asset.id}`}</span></div>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无资产；TOS 上传接口尚未接入。" />}</div> }
function MediaTaskList({ tasks, candidates, busy, onCandidates, onSelect, onRetry }) { return <div className="shuihuo-media-tasks"><b>媒体任务</b>{tasks.length ? tasks.map((task) => { const status = task.status || 'unknown'; return <article key={task.id}><header><Tag color={status === 'succeeded' ? 'green' : status.includes('failed') ? 'red' : status === 'running' ? 'blue' : 'gold'}>{task.kind} · {status}</Tag><span>{task.provider || '未分配执行器'} {task.model}</span></header>{task.errorMessage && <Typography.Text type="danger">{task.errorMessage}</Typography.Text>}<Space wrap><Button size="small" onClick={() => onCandidates(task)}>候选结果</Button>{['failed', 'retryable_failed'].includes(status) && <Button danger size="small" loading={busy} onClick={() => onRetry(task)}>安全重试</Button>}</Space>{(candidates[task.id] || []).map((candidate) => <div className="shuihuo-candidate" key={candidate.id}><span>候选 {candidate.position + 1} · 资产 #{candidate.assetId}</span>{candidate.selected ? <Tag color="green">主候选</Tag> : <Button size="small" disabled={busy} onClick={() => onSelect(task, candidate)}>设为主候选</Button>}</div>)}</article> }) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无媒体任务；未配置 Provider 时不会伪造生成结果。" />}</div> }
