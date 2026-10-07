import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Col, Descriptions, Drawer, Empty, Form, Input, Modal, Progress, Row, Space, Statistic, Table, Typography } from 'antd'
import { cancelVideoTask, getGenerationStage, getProjectGeneration, getProjectVideoStatus, listBatchProjects, retryGenerationStage, retryVideoTask, runBookGeneration, runProjectGeneration, startVideoTask } from './api.js'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './shuihuo-production.css'

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

  return <main className="shuihuo-production-page">
    <section className="shuihuo-hero"><div><Typography.Text className="eyebrow">一战晟铭 · 生产工作台</Typography.Text><Typography.Title level={2}>水货生产</Typography.Title><Typography.Paragraph>从已入库小说继续执行剧本、分镜、视频生产；状态与恢复均以 Go API 和 MySQL 为准。</Typography.Paragraph></div><Space wrap><Button onClick={() => { setLoading(true); void refreshProjects().then(() => refreshProject()).catch((reason) => setError(errorText(reason, '刷新失败'))).finally(() => setLoading(false)) }}>刷新项目</Button><Button type="primary" loading={working} disabled={!selected} onClick={() => void run()}>批量继续执行</Button></Space></section>
    {error && <Alert className="shuihuo-alert" type="error" showIcon closable message={error} onClose={() => setError('')} />}
    <Row gutter={[16, 16]}><Col xs={24} lg={7}><Card title="生产项目" className="shuihuo-projects" loading={loading}>{projects.length ? <Space direction="vertical" size={10} style={{ width: '100%' }}>{projects.map((project) => <button key={project.id} type="button" className={`shuihuo-project ${selected?.id === project.id ? 'is-selected' : ''}`} onClick={() => { setSelected(project); setSummary(null); setVideos(null); void refreshProject(project).catch((reason) => setError(errorText(reason, '读取项目状态失败'))) }}><span><b>{project.name || '未命名项目'}</b><small>#{project.id} · {project.bookCount || 0} 本小说</small></span><StatusTag status={project.runStatus || 'pending'} /></button>)}</Space> : <Empty description="暂无可生产项目，请先在小说获取或批量工厂创建项目。" />}</Card></Col><Col xs={24} lg={17}><Card className="shuihuo-console" title={selected ? `${selected.name} · 生产控制台` : '生产控制台'} extra={selected && <UnifiedSettingsPanel project={selected} />}>{selected ? <><Descriptions size="small" column={{ xs: 2, sm: 4 }}><Descriptions.Item label="完成"><Statistic value={completed} /></Descriptions.Item><Descriptions.Item label="执行中"><Statistic value={summary?.running || 0} /></Descriptions.Item><Descriptions.Item label="失败"><Statistic value={summary?.failed || 0} valueStyle={{ color: '#cf1322' }} /></Descriptions.Item><Descriptions.Item label="待执行"><Statistic value={summary?.pending || 0} /></Descriptions.Item></Descriptions><Progress percent={total ? Math.round(completed / total * 100) : 0} showInfo={false} status={summary?.failed ? 'exception' : 'active'} /><Table className="shuihuo-table" rowKey="bookId" columns={columns} dataSource={rows} loading={loading || working} pagination={false} scroll={{ x: 1320 }} locale={{ emptyText: '项目尚无可执行的小说记录' }} /></> : <Empty description="请选择一个项目" />}</Card></Col></Row>
    <Modal title={stageResult ? `${stageLabel[stageResult.stage] || stageResult.stage} · 执行结果` : '执行结果'} open={Boolean(stageResult)} onCancel={() => setStageResult(null)} footer={null} width={820}><Typography.Paragraph type="secondary">提示词：{stageResult?.promptKey || '-'} v{stageResult?.promptVersion || '-'}</Typography.Paragraph><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{stageResult?.outputText || '暂无输出'}</Typography.Paragraph></Modal>
    <Drawer title="提交视频任务" open={Boolean(videoTarget)} onClose={() => setVideoTarget(null)} width={420} extra={<Button type="primary" loading={working} onClick={() => void submitVideo()}>提交</Button>}><Alert type="info" showIcon message="仅提交已完成视频提示词的小说" description="提供方与模型由现有视频配置决定；服务端会校验配置、权限与最终提示词。" /><Form form={videoForm} layout="vertical" style={{ marginTop: 20 }}><Form.Item label="提供方" name="provider" rules={[{ required: true, message: '请输入已配置的视频提供方' }]}><Input placeholder="例如：yfai" /></Form.Item><Form.Item label="模型" name="model" rules={[{ required: true, message: '请输入已启用的视频模型' }]}><Input placeholder="例如：seedance-1.0-pro" /></Form.Item></Form></Drawer>
  </main>
}
