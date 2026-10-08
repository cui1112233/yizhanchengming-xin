import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Col, Descriptions, Drawer, Form, Input, Row, Select, Space, Table, Tag, Typography } from 'antd'
import { executeIntake, getNovelFetchWorkshop, listIntakes, restoreIntakeBook, saveNovelFetchWorkshop } from './api.js'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './novel-fetch.css'

const { TextArea } = Input

function messageFor(error) {
  if (error?.status === 403) return '当前账号无权限访问小说处理工作台。'
  if (error?.status === 401) return '登录状态已过期，请重新登录。'
  return error?.message || '请求失败，请稍后重试。'
}

function requestedIntakeID() {
  const value = new URLSearchParams(window.location.search).get('intakeId')
  const id = Number(value)
  return Number.isInteger(id) && id > 0 ? id : 0
}

export default function NovelFetchWorkshop() {
  const [intakes, setIntakes] = useState([])
  const [intakeId, setIntakeId] = useState(requestedIntakeID)
  const [snapshot, setSnapshot] = useState(null)
  const [settings, setSettings] = useState({ processingRules: '', knowledgeBase: '', taskNotes: '' })
  const [state, setState] = useState('loading')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [retrying, setRetrying] = useState(false)
  const [original, setOriginal] = useState(null)

  const restoreBook = async (bookId) => {
    setRetrying(true); setError('')
    try { await restoreIntakeBook(intakeId, bookId); await load(intakeId) }
    catch (cause) { setError(messageFor(cause)) } finally { setRetrying(false) }
  }

  const load = async (preferred = intakeId) => {
    setState('loading'); setError('')
    try {
      const listing = await listIntakes()
      const rows = Array.isArray(listing?.intakes) ? listing.intakes : []
      setIntakes(rows)
      const id = preferred || rows[0]?.id || 0
      if (!id) { setSnapshot(null); setState('empty'); return }
      const value = await getNovelFetchWorkshop(id)
      setIntakeId(id); setSnapshot(value)
      setSettings({ processingRules: value?.settings?.processingRules || '', knowledgeBase: value?.settings?.knowledgeBase || '', taskNotes: value?.settings?.taskNotes || '' })
      setState('ready')
    } catch (cause) { setSnapshot(null); setError(messageFor(cause)); setState(cause?.status === 403 ? 'forbidden' : 'failed') }
  }

  useEffect(() => { void load() }, [])
  const books = snapshot?.books || []
  const promptRows = snapshot?.prompts || []
  const failedCount = books.filter((book) => book.status === 'retryable_failed').length
  const sourceSummary = useMemo(() => [...new Set(books.map((book) => book.source).filter(Boolean))].join('、') || '-', [books])

  const changeIntake = (next) => {
    const id = Number(next)
    window.history.replaceState({}, '', `/novel-fetch-workshop?intakeId=${id}`)
    void load(id)
  }
  const save = async () => {
    setSaving(true); setError('')
    try { const result = await saveNovelFetchWorkshop(intakeId, settings); setSettings(result.settings || settings) }
    catch (cause) { setError(messageFor(cause)) } finally { setSaving(false) }
  }
  const retry = async () => {
    setRetrying(true); setError('')
    try { await executeIntake(intakeId, 4000); await load(intakeId) }
    catch (cause) { setError(messageFor(cause)) } finally { setRetrying(false) }
  }

  if (state === 'loading') return <PageState state="loading" title="正在从服务端恢复 Workshop…" />
  if (state === 'forbidden') return <PageState state="failed" title="无权限访问" description={error} />
  if (state === 'failed') return <PageState state="failed" title="加载 Workshop 失败" description={error} onRetry={() => void load()} />
  if (state === 'empty') return <PageState title="暂无小说获取任务" description="请先在小说获取中创建 Intake，再进入处理工作台。" />

  const columns = [
    { title: '书名 / Book ID', render: (_, row) => <Space direction="vertical" size={0}><b>{row.title || '-'}</b><Typography.Text type="secondary">{row.bookId}</Typography.Text></Space> },
    { title: '来源', dataIndex: 'source' },
    { title: '男女频 / 风格', render: (_, row) => <Space wrap><Tag>{row.gender || '-'}</Tag><Tag>{row.style || '-'}</Tag></Space> },
    { title: '状态', render: (_, row) => <StatusTag status={row.status} /> },
    { title: '原文', render: (_, row) => <Space><Button size="small" disabled={!row.originalText} onClick={() => setOriginal(row)}>查看原文</Button><Button size="small" loading={retrying} onClick={() => void restoreBook(row.id)}>恢复原文</Button></Space> },
  ]

  return <main className="novel-fetch-page novel-fetch-workshop-shell">
    <Space direction="vertical" size={16} style={{ display: 'flex' }}>
      <Card className="legacy-panel-card" title="小说处理工作台" extra={<Button href="/novel-fetch">返回小说获取</Button>}>
        {error ? <Alert type="error" showIcon message={error} closable onClose={() => setError('')} style={{ marginBottom: 16 }} /> : null}
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}><Form.Item label="Intake 任务"><Select value={intakeId} onChange={changeIntake} options={intakes.map((item) => ({ value:item.id, label:`#${item.id} · ${item.name} · ${item.status}` }))} /></Form.Item></Col>
          <Col xs={24} md={12}><Descriptions size="small" column={2}><Descriptions.Item label="任务状态"><StatusTag status={snapshot?.intake?.status} /></Descriptions.Item><Descriptions.Item label="来源">{sourceSummary}</Descriptions.Item><Descriptions.Item label="小说数">{books.length}</Descriptions.Item><Descriptions.Item label="失败">{failedCount}</Descriptions.Item></Descriptions></Col>
        </Row>
      </Card>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={14}><Card className="legacy-panel-card" title="处理与任务"><Table rowKey="id" columns={columns} dataSource={books} pagination={{ pageSize: 10 }} locale={{ emptyText:'该任务暂无书籍' }} /><Space style={{ marginTop: 14 }}><Button type="primary" loading={retrying} disabled={!failedCount} onClick={() => void retry()}>重试失败书籍</Button><Typography.Text type="secondary">重试复用既有 Intake execute 服务与 provider121，不创建新队列。</Typography.Text></Space></Card></Col>
        <Col xs={24} xl={10}><Card className="legacy-panel-card" title="服务端持久化处理设置"><Form layout="vertical"><Form.Item label="处理规则"><TextArea aria-label="处理规则" value={settings.processingRules} onChange={(event) => setSettings({ ...settings, processingRules:event.target.value })} rows={4} placeholder="保存为当前 Intake 的处理规则" /></Form.Item><Form.Item label="知识库说明"><TextArea aria-label="知识库说明" value={settings.knowledgeBase} onChange={(event) => setSettings({ ...settings, knowledgeBase:event.target.value })} rows={3} placeholder="由服务端持久化，不保存到浏览器" /></Form.Item><Form.Item label="任务备注"><Input aria-label="任务备注" value={settings.taskNotes} onChange={(event) => setSettings({ ...settings, taskNotes:event.target.value })} /></Form.Item><Button type="primary" loading={saving} onClick={() => void save()}>保存配置</Button></Form></Card><Card className="legacy-panel-card" title="后端提示词 / 知识库入口" style={{ marginTop:16 }}><Typography.Paragraph type="secondary">提示词由 Go `generation_prompts` 服务端事实源提供，前端不写入系统提示词。</Typography.Paragraph>{promptRows.map((item) => <Tag key={`${item.key}-${item.version}`}>{item.key} v{item.version}{item.enabled ? '' : '（停用）'}</Tag>)}</Card></Col>
      </Row>
    </Space>
    <Drawer title={original ? `${original.title || '小说'} · 原文` : '原文'} open={!!original} width={720} onClose={() => setOriginal(null)}>{original ? <Typography.Paragraph style={{ whiteSpace:'pre-wrap' }}>{original.originalText}</Typography.Paragraph> : null}</Drawer>
  </main>
}
