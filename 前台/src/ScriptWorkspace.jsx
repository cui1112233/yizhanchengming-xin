import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Col, Drawer, Empty, Form, Input, List, Modal, Row, Select, Space, Spin, Steps, Tag, Typography } from 'antd'
import { createBatchProject, getAudioMeasurement, getGenerationStage, getScriptWorkspace, listBatchProjects, retryGenerationStage, runBookGeneration, saveProductionSettings } from './api.js'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './script-workspace.css'

const { TextArea } = Input
const STAGES = ['SCRIPT', 'HOOK', 'DIRECTOR', 'FINAL_PROMPT']
const labels = { SCRIPT: '剧本', HOOK: '黄金三秒', DIRECTOR: '导演分镜', FINAL_PROMPT: '视频提示词' }
const errorText = (error, fallback) => error instanceof Error ? error.message : fallback

function workspaceSettings(settings) {
  const value = settings?.project?.production?.scriptWorkspace
  return value && typeof value === 'object' ? value : { characters: '', scenes: '', constraints: '', mode: 'continuous', duration: 10, directorMode: 'normal', model: '' }
}

export default function ScriptWorkspace() {
  const [projects, setProjects] = useState([])
  const [projectId, setProjectId] = useState(null)
  const [workspace, setWorkspace] = useState(null)
  const [settings, setSettings] = useState(workspaceSettings(null))
  const [bookId, setBookId] = useState(null)
  const [loading, setLoading] = useState(true)
  const [working, setWorking] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm()

  const refreshProjects = useCallback(async () => {
    const response = await listBatchProjects()
    const rows = response?.projects || []
    setProjects(rows)
    setProjectId(current => rows.some(row => row.id === current) ? current : rows[0]?.id || null)
    return rows
  }, [])
  const refreshWorkspace = useCallback(async (id = projectId) => {
    if (!id) { setWorkspace(null); return }
    const value = await getScriptWorkspace(id)
    setWorkspace(value)
    setSettings(workspaceSettings(value.settings))
    const books = value.project?.books || []
    setBookId(current => books.some(book => book.id === current) ? current : books[0]?.id || null)
    setError('')
  }, [projectId])
  useEffect(() => { let live = true; setLoading(true); refreshProjects().catch(reason => live && setError(errorText(reason, '读取剧本项目失败'))).finally(() => live && setLoading(false)); return () => { live = false } }, [refreshProjects])
  useEffect(() => { if (!projectId) return; setLoading(true); refreshWorkspace(projectId).catch(reason => setError(errorText(reason, '读取工作台失败'))).finally(() => setLoading(false)) }, [projectId, refreshWorkspace])

  const books = workspace?.project?.books || []
  const selectedBook = books.find(book => book.id === bookId) || null
  const summary = useMemo(() => (workspace?.generation?.books || []).find(item => item.bookId === bookId) || { stages: {} }, [workspace, bookId])
  const videos = useMemo(() => (workspace?.video?.books || []).find(item => item.bookId === bookId), [workspace, bookId])
  const prompts = workspace?.prompts?.prompts || []
  const promptVersions = Object.fromEntries(prompts.map(item => [item.key, item.version]))

  const saveSettings = async () => {
    if (!projectId) return
    setWorking(true)
    try {
      const existing = workspace?.settings?.project?.production || {}
      await saveProductionSettings(projectId, { ...existing, scriptWorkspace: settings })
      await refreshWorkspace(projectId)
    } catch (reason) { setError(errorText(reason, '保存剧本配置失败')) } finally { setWorking(false) }
  }
  const run = async () => {
    if (!projectId || !selectedBook) return
    setWorking(true)
    try {
      let matchAudio = false
      try { await getAudioMeasurement(projectId, selectedBook.id); matchAudio = Boolean(settings.matchAudio) } catch { matchAudio = false }
      await runBookGeneration(projectId, selectedBook.id, {
        hookEnabled: settings.mode !== 'continuous', plotMode: settings.mode === 'segmented', directorMode: settings.directorMode || 'normal', matchAudio,
        shotDurationLimitSec: Number(settings.duration) || 10,
        requestId: `script-${projectId}-${selectedBook.id}-${Date.now()}`,
        processingRules: settings.constraints || '', projectConfig: JSON.stringify({ characters: settings.characters || '', scenes: settings.scenes || '', model: settings.model || '' }),
      })
      await refreshWorkspace(projectId)
    } catch (reason) { setError(errorText(reason, '生成剧本失败')) } finally { setWorking(false) }
  }
  const retry = async (stage) => { if (!projectId || !selectedBook) return; setWorking(true); try { await retryGenerationStage(projectId, selectedBook.id, stage, `script-retry-${stage}-${Date.now()}`); await refreshWorkspace(projectId) } catch (reason) { setError(errorText(reason, '重试失败')) } finally { setWorking(false) } }
  const openStage = async (stage) => { if (!projectId || !selectedBook) return; try { setResult(await getGenerationStage(projectId, selectedBook.id, stage)) } catch (reason) { setError(errorText(reason, '读取阶段结果失败')) } }
  const createProject = async () => { try { const values = await createForm.validateFields(); const response = await createBatchProject(values.intakeId, { name: values.name }); setCreateOpen(false); await refreshProjects(); setProjectId(response?.project?.id || null) } catch (reason) { if (reason?.errorFields) return; setError(errorText(reason, '创建项目失败')) } }

  if (loading && !workspace && !projects.length) return <PageState state="loading" title="正在恢复剧本工作区…" />
  return <main className="script-workspace-page">
    <section className="script-workspace-hero"><div><Typography.Text className="eyebrow">一战晟铭 · SCRIPT WORKSPACE</Typography.Text><Typography.Title level={2}>剧本生成</Typography.Title><Typography.Paragraph>复用已入库项目、Go generation 与 MySQL StageRun；刷新和重新登录后从服务端恢复。</Typography.Paragraph></div><Space><Button onClick={() => { setLoading(true); void refreshProjects().then(rows => refreshWorkspace(projectId || rows[0]?.id)).catch(reason => setError(errorText(reason, '刷新失败'))).finally(() => setLoading(false))}}>刷新</Button><Button type="primary" onClick={() => setCreateOpen(true)}>从已获取小说创建项目</Button></Space></section>
    {error && <Alert type="error" showIcon closable className="script-workspace-alert" message={error} onClose={() => setError('')} />}
    <Row gutter={[16, 16]}><Col xs={24} xl={5}><Card title="脚本项目" className="script-project-list" loading={loading}>{projects.length ? <List dataSource={projects} renderItem={item => <List.Item className={item.id === projectId ? 'is-selected' : ''} onClick={() => setProjectId(item.id)}><Space direction="vertical" size={1}><Typography.Text strong>{item.name}</Typography.Text><Typography.Text type="secondary">#{item.id} · {item.bookCount || 0} 本</Typography.Text></Space><StatusTag status={item.runStatus || 'pending'} /></List.Item>} /> : <Empty description="暂无项目。请先在小说获取中保存原文，再创建项目。" />}</Card></Col>
      <Col xs={24} xl={12}><Card title="脚本工作台" loading={loading} extra={selectedBook && <Select aria-label="选择小说" value={bookId} onChange={setBookId} options={books.map(book => ({ value: book.id, label: book.title || `Book ${book.id}` }))} style={{ minWidth: 200 }} />}>{selectedBook ? <><TextArea aria-label="小说原文" value={selectedBook.originalText || '原文由小说获取服务保存；此处只读展示，避免绕过统一来源。'} readOnly autoSize={{ minRows: 8, maxRows: 14 }} /><div className="script-stage-row">{STAGES.map(stage => { const state = summary.stages?.[stage]; return <Card size="small" key={stage} title={labels[stage]}><StatusTag status={state?.status || 'pending'} />{state?.outputText && <Button type="link" size="small" onClick={() => void openStage(stage)}>查看</Button>}{state?.status === 'failed' && <Button type="link" danger size="small" onClick={() => void retry(stage)}>重试</Button>}<Typography.Text type="secondary" className="script-prompt-version">{state?.promptKey ? `${state.promptKey} v${state.promptVersion}` : stage === 'SCRIPT' ? `script.default v${promptVersions['script.default'] || '-'}` : '等待执行'}</Typography.Text></Card> })}</div><Space wrap><Button type="primary" loading={working} onClick={() => void run()}>生成剧本</Button><Button onClick={() => setHistoryOpen(true)}>历史</Button>{videos?.attempts?.length ? <Tag color="blue">视频任务 {videos.attempts.length}</Tag> : <Tag>暂无视频任务</Tag>}</Space></> : <Empty description="请选择一个已有项目和小说" />}</Card></Col>
      <Col xs={24} xl={7}><Card title="人物、场景与约束" extra={<Button type="link" loading={working} onClick={() => void saveSettings()}>保存到服务端</Button>}><Form layout="vertical"><Form.Item label="人物"><TextArea aria-label="人物" value={settings.characters} onChange={event => setSettings(value => ({ ...value, characters: event.target.value }))} placeholder="每行一个人物" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="场景"><TextArea aria-label="场景" value={settings.scenes} onChange={event => setSettings(value => ({ ...value, scenes: event.target.value }))} placeholder="每行一个场景" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="约束设置"><TextArea aria-label="约束设置" value={settings.constraints} onChange={event => setSettings(value => ({ ...value, constraints: event.target.value }))} placeholder="从服务端 prompt 模块编译，不在前端写系统提示词" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="开头策略"><Select value={settings.mode} onChange={mode => setSettings(value => ({ ...value, mode }))} options={[{ value: 'continuous', label: '连续开头' }, { value: 'hook', label: '爆款开头' }, { value: 'segmented', label: '分段开头' }]} /></Form.Item><Form.Item label="分镜时长"><Select value={settings.duration} onChange={duration => setSettings(value => ({ ...value, duration }))} options={[{ value: 10, label: '10s' }, { value: 15, label: '15s' }]} /></Form.Item><Form.Item label="导演模型"><Select value={settings.directorMode} onChange={directorMode => setSettings(value => ({ ...value, directorMode }))} options={[{ value: 'normal', label: '标准导演' }, { value: 'h3', label: 'H3 导演' }]} /></Form.Item><Form.Item label="模型备注"><Input aria-label="模型备注" value={settings.model} onChange={event => setSettings(value => ({ ...value, model: event.target.value }))} placeholder="仅保存项目配置；模型凭据由后台管理" /></Form.Item></Form></Card></Col></Row>
    <Drawer title="阶段历史" open={historyOpen} onClose={() => setHistoryOpen(false)} width={560}><Steps direction="vertical" current={-1} items={STAGES.map(stage => ({ title: labels[stage], description: `${summary.stages?.[stage]?.status || 'pending'} · ${summary.stages?.[stage]?.updatedAt || '尚未执行'}` }))} /></Drawer>
    <Modal title={result ? `${labels[result.stage] || result.stage} · 结果` : '结果'} open={Boolean(result)} onCancel={() => setResult(null)} footer={null} width={820}><Typography.Paragraph type="secondary">服务端提示词：{result?.promptKey || '-'} v{result?.promptVersion || '-'}</Typography.Paragraph><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{result?.outputText || result?.errorMessage || '暂无输出'}</Typography.Paragraph></Modal>
    <Modal title="从已获取小说创建脚本项目" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => void createProject()} okText="创建"><Alert type="info" showIcon message="仅复用已有 Intake" description="不会创建第二套项目、任务或队列；请填写已有小说获取批次 ID。" /><Form form={createForm} layout="vertical" style={{ marginTop: 16 }}><Form.Item name="intakeId" label="小说获取批次 ID" rules={[{ required: true, message: '请输入 Intake ID' }]}><Input inputMode="numeric" /></Form.Item><Form.Item name="name" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}><Input /></Form.Item></Form></Modal>
  </main>
}
