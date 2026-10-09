import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Button, Card, Col, Drawer, Empty, Form, Input, List, Modal, Row, Select, Space, Spin, Steps, Tag, Typography } from 'antd'
import { createBatchProject, deleteScriptStoryboardCard, getAudioMeasurement, getGenerationStage, getScriptStoryboard, getScriptWorkspace, listBatchProjects, reorderScriptStoryboard, retryGenerationStage, runBookGeneration, saveProductionSettings, saveScriptOriginalText, saveScriptStoryboardCard } from './api.js'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import { generationOutcomeMessage } from './ui/generationOutcome.js'
import { activeGenerationRunId, useGenerationRunPolling } from './generationRuntime.js'
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
  const [originalText, setOriginalText] = useState('')
  const [storyboard, setStoryboard] = useState(null)
  const workspaceRequest = useRef(0)
  const projectIdRef = useRef(projectId)
  projectIdRef.current = projectId

  const refreshProjects = useCallback(async () => {
    const response = await listBatchProjects()
    const rows = response?.projects || []
    setProjects(rows)
    setProjectId(current => rows.some(row => row.id === current) ? current : rows[0]?.id || null)
    return rows
  }, [])
  const refreshWorkspace = useCallback(async (id = projectIdRef.current) => {
    const scope = Number(id)
    const requestId = ++workspaceRequest.current
    if (!Number.isSafeInteger(scope) || scope <= 0) { setWorkspace(null); return null }
    const value = await getScriptWorkspace(id)
    if (requestId !== workspaceRequest.current || Number(projectIdRef.current) !== scope) return null
    if (Number(value?.project?.project?.id) !== scope) throw new Error('工作台项目范围不匹配')
    setWorkspace(value)
    setSettings(workspaceSettings(value.settings))
    const books = value.project?.books || []
    setBookId(current => books.some(book => book.id === current) ? current : books[0]?.id || null)
    setError('')
    return value
  }, [])
  useEffect(() => { let live = true; setLoading(true); refreshProjects().catch(reason => live && setError(errorText(reason, '读取剧本项目失败'))).finally(() => live && setLoading(false)); return () => { live = false } }, [refreshProjects])
  useEffect(() => { if (!projectId) return; const scope = projectId; setLoading(true); refreshWorkspace(scope).catch(reason => { if (Number(projectIdRef.current) === Number(scope)) setError(errorText(reason, '读取工作台失败')) }).finally(() => { if (Number(projectIdRef.current) === Number(scope)) setLoading(false) }) }, [projectId, refreshWorkspace])

  const books = workspace?.project?.books || []
  const selectedBook = books.find(book => book.id === bookId) || null
  useEffect(() => { setOriginalText(selectedBook?.originalText || '') }, [selectedBook?.id, selectedBook?.originalText])
  const refreshStoryboard = useCallback(async () => { if (!projectId || !bookId) { setStoryboard(null); return }; setStoryboard(await getScriptStoryboard(projectId, bookId)) }, [projectId, bookId])
  useEffect(() => { void refreshStoryboard().catch(reason => { if (reason?.status !== 409 && reason?.status !== 404) setError(errorText(reason, '读取分镜失败')) }) }, [refreshStoryboard])
  const summary = useMemo(() => (workspace?.generation?.books || []).find(item => item.bookId === bookId) || { stages: {} }, [workspace, bookId])
  const videos = useMemo(() => (workspace?.video?.books || []).find(item => item.bookId === bookId), [workspace, bookId])
  const prompts = workspace?.prompts?.prompts || []
  const promptVersions = Object.fromEntries(prompts.map(item => [item.key, item.version]))
  const generationRun = useGenerationRunPolling({
    projectId,
    onTerminal: () => { if (projectId) void Promise.all([refreshWorkspace(projectId), refreshStoryboard()]) },
    onError: (reason) => setError(errorText(reason, '生成状态轮询失败，正在继续重试')),
  })
  const pipelineBusy = working || generationRun.submitting || generationRun.active
  useEffect(() => {
    const runId = activeGenerationRunId(workspace?.generation, bookId)
    if (runId) generationRun.resume(runId, workspace?.generation?.batchProjectId || workspace?.project?.project?.id)
  }, [workspace?.generation, projectId, bookId])

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
    if (!projectId || !selectedBook || pipelineBusy) return
    try {
      let matchAudio = false
      try { await getAudioMeasurement(projectId, selectedBook.id); matchAudio = Boolean(settings.matchAudio) } catch { matchAudio = false }
      await generationRun.admit(`script:${projectId}:${selectedBook.id}`, (idempotencyKey) => runBookGeneration(projectId, selectedBook.id, {
        hookEnabled: settings.mode !== 'continuous', plotMode: settings.mode === 'segmented', directorMode: settings.directorMode || 'normal', matchAudio,
        shotDurationLimitSec: Number(settings.duration) || 10,
        processingRules: settings.constraints || '', projectConfig: JSON.stringify({ characters: settings.characters || '', scenes: settings.scenes || '', model: settings.model || '' }),
      }, { idempotencyKey }))
    } catch (reason) { setError(errorText(reason, '生成剧本失败')) }
  }
  const retry = async (stage) => {
    if (!projectId || !selectedBook || pipelineBusy) return
    const sourceBookRunId = summary.stages?.[stage]?.bookRunId
    if (!Number.isSafeInteger(Number(sourceBookRunId)) || Number(sourceBookRunId) <= 0) { setError('缺少可验证的失败运行记录，请先刷新。'); return }
    try {
      await generationRun.admit(`script-stage:${projectId}:${selectedBook.id}:${stage}:${sourceBookRunId}`, (idempotencyKey) => retryGenerationStage(projectId, selectedBook.id, stage, { sourceBookRunId }, { idempotencyKey }))
    } catch (reason) { setError(errorText(reason, '重试失败')) }
  }
  const saveOriginal = async () => { if (!projectId || !selectedBook) return; setWorking(true); try { await saveScriptOriginalText(projectId, selectedBook.id, originalText); await refreshWorkspace(projectId) } catch (reason) { setError(errorText(reason, '保存原文失败')) } finally { setWorking(false) } }
  const uploadTXT = async (file) => { try { setOriginalText(await file.text()) } catch (reason) { setError(errorText(reason, '读取 TXT 失败')) } }
  const openStage = async (stage) => { if (!projectId || !selectedBook) return; try { setResult(await getGenerationStage(projectId, selectedBook.id, stage)) } catch (reason) { setError(errorText(reason, '读取阶段结果失败')) } }
  const createProject = async () => { try { const values = await createForm.validateFields(); const response = await createBatchProject(values.intakeId, { name: values.name }); setCreateOpen(false); await refreshProjects(); setProjectId(response?.project?.id || null) } catch (reason) { if (reason?.errorFields) return; setError(errorText(reason, '创建项目失败')) } }
  const saveCard = async (card) => { if (!projectId || !bookId || !storyboard) return; setWorking(true); try { setStoryboard(await saveScriptStoryboardCard(projectId, bookId, { card, expectedVersion: storyboard.version })) } catch (reason) { setError(reason?.status === 409 ? '分镜已被其他编辑更新，请刷新后处理冲突。' : errorText(reason, '保存分镜失败')) } finally { setWorking(false) } }
  const removeCard = async (card) => { if (!projectId || !bookId || !storyboard) return; setWorking(true); try { setStoryboard(await deleteScriptStoryboardCard(projectId, bookId, card.id, storyboard.version)) } catch (reason) { setError(reason?.status === 409 ? '分镜已被其他编辑更新，请刷新后处理冲突。' : errorText(reason, '删除分镜失败')) } finally { setWorking(false) } }
  const reorderCards = async (from, to) => { if (!projectId || !bookId || !storyboard || to < 0 || to >= storyboard.cards.length) return; const cards = [...storyboard.cards]; const [card] = cards.splice(from, 1); cards.splice(to, 0, card); setWorking(true); try { setStoryboard(await reorderScriptStoryboard(projectId, bookId, cards.map(item => item.id), storyboard.version)) } catch (reason) { setError(reason?.status === 409 ? '分镜顺序已被其他编辑修改，请刷新。' : errorText(reason, '排序分镜失败')) } finally { setWorking(false) } }

  if (loading && !workspace && !projects.length) return <PageState state="loading" title="正在恢复剧本工作区…" />
  return <main className="script-workbench-form">
    <section className="script-workspace-page script-workbench utility-workbench">
      <div className="script-left">
        <div className="script-left-scroll">
          <div className="script-chat-shell">
    <section className="script-workspace-hero"><div><Typography.Text className="eyebrow">一战晟铭 · SCRIPT WORKSPACE</Typography.Text><Typography.Title level={2}>剧本生成</Typography.Title><Typography.Paragraph>复用已入库项目、Go generation 与 MySQL StageRun；刷新和重新登录后从服务端恢复。</Typography.Paragraph></div><Space><Button onClick={() => { setLoading(true); void refreshProjects().then(rows => refreshWorkspace(projectId || rows[0]?.id)).catch(reason => setError(errorText(reason, '刷新失败'))).finally(() => setLoading(false))}}>刷新</Button><Button type="primary" onClick={() => setCreateOpen(true)}>从已获取小说创建项目</Button></Space></section>
    {error && <Alert type="error" showIcon closable className="script-workspace-alert" message={error} onClose={() => setError('')} />}
    <Row gutter={[16, 16]}><Col xs={24} xl={5}><Card title="脚本项目" className="script-project-list" loading={loading}>{projects.length ? <List dataSource={projects} renderItem={item => <List.Item className={item.id === projectId ? 'is-selected' : ''} onClick={() => setProjectId(item.id)}><Space direction="vertical" size={1}><Typography.Text strong>{item.name}</Typography.Text><Typography.Text type="secondary">#{item.id} · {item.bookCount || 0} 本</Typography.Text></Space><StatusTag status={item.runStatus || 'pending'} /></List.Item>} /> : <Empty description="暂无项目。请先在小说获取中保存原文，再创建项目。" />}</Card></Col>
      <Col xs={24} xl={12}><Card title="脚本工作台" loading={loading} extra={selectedBook && <Select aria-label="选择小说" value={bookId} onChange={setBookId} options={books.map(book => ({ value: book.id, label: book.title || `Book ${book.id}` }))} style={{ minWidth: 200}} />}>{selectedBook ? <><TextArea aria-label="小说原文" value={originalText} onChange={event => setOriginalText(event.target.value)} autoSize={{ minRows: 8, maxRows: 14 }} /><Space wrap style={{ marginTop: 10 }}><Button loading={working} onClick={() => void saveOriginal()}>保存原文</Button><input aria-label="上传 TXT" type="file" accept=".txt,text/plain" onChange={event => { const file = event.target.files?.[0]; if (file) void uploadTXT(file); event.target.value = '' }} /><Typography.Text type="secondary">TXT 仅载入编辑器；保存后写入 MySQL。</Typography.Text></Space><div className="script-stage-row">{STAGES.map(stage => { const state = summary.stages?.[stage]; return <Card size="small" key={stage} title={labels[stage]}><StatusTag status={state?.status || 'pending'} />{state?.outputText && <Button type="link" size="small" onClick={() => void openStage(stage)}>查看</Button>}{state?.status === 'failed' && <Button type="link" danger size="small" disabled={pipelineBusy} onClick={() => void retry(stage)}>重试</Button>}<Typography.Text type="secondary" className="script-prompt-version">{state?.promptKey ? `${state.promptKey} v${state.promptVersion}` : stage === 'SCRIPT' ? `script.default v${promptVersions['script.default'] || '-'}` : '等待执行'}</Typography.Text></Card> })}</div><Space wrap><Button type="primary" loading={pipelineBusy} onClick={() => void run()}>按已保存原文生成剧本</Button><Button onClick={() => setHistoryOpen(true)}>历史</Button>{videos?.attempts?.length ? <Tag color="blue">视频任务 {videos.attempts.length}</Tag> : <Tag>暂无视频任务</Tag>}</Space><Card className="script-storyboard" title="画布分镜" extra={<Space><Tag>版本 {storyboard?.version || '-'}</Tag><Button loading={working} onClick={() => void refreshStoryboard()}>刷新</Button><Button type="primary" disabled title="等待统一异步重编译动作接入">重新编译待接入</Button></Space>}><Typography.Paragraph type="secondary">保存卡片后，以已保存顺序重建 Director 与 Final Prompt；异步重编译接入前不会回退旧同步链路。</Typography.Paragraph>{storyboard?.cards?.length ? storyboard.cards.map((card,index) => <Card key={card.id || index} size="small" className="script-storyboard-card" title={`#${card.position} ${card.title || '分镜'}`} extra={<Space><Button size="small" disabled={working || index===0} onClick={() => void reorderCards(index,index-1)}>上移</Button><Button size="small" disabled={working || index===storyboard.cards.length-1} onClick={() => void reorderCards(index,index+1)}>下移</Button><Button size="small" danger disabled={working} onClick={() => void removeCard(card)}>删除</Button></Space>}><Input aria-label={`分镜标题 ${index+1}`} value={card.title} onChange={e => setStoryboard(value => ({...value,cards:value.cards.map(item=>item.id===card.id?{...item,title:e.target.value}:item)}))}/><TextArea aria-label={`分镜内容 ${index+1}`} value={card.content} onChange={e => setStoryboard(value => ({...value,cards:value.cards.map(item=>item.id===card.id?{...item,content:e.target.value}:item)}))} autoSize={{minRows:4,maxRows:10}} style={{marginTop:8}}/><Button size="small" type="primary" loading={working} style={{marginTop:8}} onClick={() => void saveCard(card)}>保存此分镜</Button></Card>) : <Empty description="先完成导演分镜生成，画布会从服务端 Director 结果初始化。"/>}<Button style={{marginTop:10}} disabled={!storyboard || working} onClick={() => void saveCard({title:`分镜${(storyboard?.cards?.length||0)+1}`,content:'请填写分镜内容。'})}>新增分镜</Button></Card></> : <Empty description="请选择一个已有项目和小说" />}</Card></Col>
      <Col xs={24} xl={7}><Card title="人物、场景与约束" extra={<Button type="link" loading={working} onClick={() => void saveSettings()}>保存到服务端</Button>}><Form layout="vertical"><Form.Item label="人物"><TextArea aria-label="人物" value={settings.characters} onChange={event => setSettings(value => ({ ...value, characters: event.target.value }))} placeholder="每行一个人物" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="场景"><TextArea aria-label="场景" value={settings.scenes} onChange={event => setSettings(value => ({ ...value, scenes: event.target.value }))} placeholder="每行一个场景" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="约束设置"><TextArea aria-label="约束设置" value={settings.constraints} onChange={event => setSettings(value => ({ ...value, constraints: event.target.value }))} placeholder="从服务端 prompt 模块编译，不在前端写系统提示词" autoSize={{ minRows: 2, maxRows: 5 }} /></Form.Item><Form.Item label="开头策略"><Select value={settings.mode} onChange={mode => setSettings(value => ({ ...value, mode }))} options={[{ value: 'continuous', label: '连续开头' }, { value: 'hook', label: '爆款开头' }, { value: 'segmented', label: '分段开头' }]} /></Form.Item><Form.Item label="分镜时长"><Select value={settings.duration} onChange={duration => setSettings(value => ({ ...value, duration }))} options={[{ value: 10, label: '10s' }, { value: 15, label: '15s' }]} /></Form.Item><Form.Item label="导演模型"><Select value={settings.directorMode} onChange={directorMode => setSettings(value => ({ ...value, directorMode }))} options={[{ value: 'normal', label: '标准导演' }, { value: 'h3', label: 'H3 导演' }]} /></Form.Item><Form.Item label="模型备注"><Input aria-label="模型备注" value={settings.model} onChange={event => setSettings(value => ({ ...value, model: event.target.value }))} placeholder="仅保存项目配置；模型凭据由后台管理" /></Form.Item></Form></Card></Col></Row>
    <Drawer title="阶段历史" open={historyOpen} onClose={() => setHistoryOpen(false)} width={560}><Steps direction="vertical" current={-1} items={STAGES.map(stage => ({ title: labels[stage], description: `${summary.stages?.[stage]?.status || 'pending'} · ${summary.stages?.[stage]?.updatedAt || '尚未执行'}` }))} /></Drawer>
    <Modal title={result ? `${labels[result.stage] || result.stage} · 结果` : '结果'} open={Boolean(result)} onCancel={() => setResult(null)} footer={null} width={820}><Typography.Paragraph type="secondary">服务端提示词：{result?.promptKey || '-'} v{result?.promptVersion || '-'}</Typography.Paragraph><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{result?.outputText || (result?.status === 'failed' || result?.errorMessage || result?.errorCode ? generationOutcomeMessage(result) : '暂无输出')}</Typography.Paragraph></Modal>
    <Modal title="从已获取小说创建脚本项目" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => void createProject()} okText="创建"><Alert type="info" showIcon message="仅复用已有 Intake" description="不会创建第二套项目、任务或队列；请填写已有小说获取批次 ID。" /><Form form={createForm} layout="vertical" style={{ marginTop: 16 }}><Form.Item name="intakeId" label="小说获取批次 ID" rules={[{ required: true, message: '请输入 Intake ID' }]}><Input inputMode="numeric" /></Form.Item><Form.Item name="name" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}><Input /></Form.Item></Form></Modal>
          </div>
        </div>
      </div>
    </section>
  </main>
}
