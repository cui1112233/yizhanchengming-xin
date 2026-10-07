import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, Modal, Select, Space, Spin, Tag, Typography } from 'antd'
import { createBatchProject, createIntake, executeIntake, listBatchProjects } from './api.js'
import { NOVEL_FETCH_PLATFORMS } from './NovelFetchPage.jsx'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'

const sourceOptions = NOVEL_FETCH_PLATFORMS.map((item) => ({ value: item.label, label: item.label, platformId: item.value }))

function safeError(error) {
  if (error?.status === 403) return '无权限访问批量工厂。'
  return error instanceof Error ? error.message : '请求失败，请稍后重试。'
}

export default function BatchFactoryHome({ onOpenProject }) {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [denied, setDenied] = useState(false)
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('')
  const [runStatus, setRunStatus] = useState('')
  const [creating, setCreating] = useState(false)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true); setError('')
    try {
      const payload = await listBatchProjects()
      setProjects(Array.isArray(payload?.projects) ? payload.projects : [])
      setDenied(false)
    } catch (reason) {
      if (reason?.status === 403) setDenied(true)
      else setError(safeError(reason))
    } finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [])

  const visible = useMemo(() => projects.filter((project) => {
    const text = `${project.name || ''} ${(project.sources || []).join(' ')}`.toLowerCase()
    return (!query || text.includes(query.toLowerCase()))
      && (!source || (project.sources || []).includes(source))
      && (!runStatus || project.runStatus === runStatus)
  }), [projects, query, source, runStatus])

  const submit = async (values) => {
    const selected = sourceOptions.find((item) => item.value === values.source)
    if (!selected) return
    setCreating(true); setError('')
    try {
      const created = await createIntake({
        name: values.intakeName || values.projectName,
        groups: [{ source: selected.label, platformId: selected.platformId, books: [{ bookId: String(values.bookId).trim(), title: '', gender: '', style: '' }] }],
      })
      const intake = created?.intake
      if (!intake?.id) throw new Error('创建 Intake 后未返回 ID')
      const executed = await executeIntake(intake.id, values.maxText || 4000)
      if (executed?.status !== 'completed') throw new Error('Intake 未全部完成，不能创建 BatchProject；请到小说获取处理失败书籍后重试。')
      await createBatchProject(intake.id, { name: values.projectName })
      setOpen(false); form.resetFields(); await load()
    } catch (reason) { setError(safeError(reason)) }
    finally { setCreating(false) }
  }

  if (denied) return <PageState state="forbidden" title="无权访问批量工厂" description="请联系管理员开通批量查看权限。" />
  if (loading) return <PageState state="loading" title="正在读取批量项目" />
  if (error && projects.length === 0) return <PageState state="failed" title="批量项目读取失败" description={error} onRetry={() => void load()} />
  return <main className="page-shell">
    <div className="page-heading"><div><Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text><Typography.Title level={2}>批量工厂</Typography.Title><Typography.Paragraph type="secondary">项目列表和状态来自 Go API / MySQL；刷新或重新登录后会重新读取。</Typography.Paragraph></div><Space><Button onClick={() => void load()}>刷新</Button><Button type="primary" onClick={() => setOpen(true)}>新建批量</Button></Space></div>
    {error && <Alert type="error" showIcon message={error} closable onClose={() => setError('')} className="feedback" />}
    <Card className="result-card" title="批量项目">
      <Space wrap style={{ marginBottom: 16 }}><Input aria-label="搜索批量项目" placeholder="搜索项目或来源" value={query} onChange={(event) => setQuery(event.target.value)} style={{ width: 220 }} /><Select aria-label="来源筛选" allowClear placeholder="来源筛选" value={source || undefined} onChange={(value) => setSource(value || '')} options={sourceOptions.map(({ value, label }) => ({ value, label }))} style={{ width: 150 }} /><Select aria-label="运行状态筛选" allowClear placeholder="运行状态" value={runStatus || undefined} onChange={(value) => setRunStatus(value || '')} options={[{ value: 'pending', label: '待执行' }, { value: 'running', label: '执行中' }, { value: 'completed', label: '已完成' }, { value: 'failed', label: '失败' }]} style={{ width: 150 }} /></Space>
      {visible.length === 0 ? <Empty description="暂无匹配的批量项目" /> : <div className="batch-project-card-grid">{visible.map((project) => <Card key={project.id} size="small" title={project.name || '未命名项目'} extra={<StatusTag status={project.runStatus || 'pending'} />}><Space direction="vertical" size={8} style={{ width: '100%' }}><Typography.Text>书籍数量：{project.bookCount ?? 0}</Typography.Text><div>{(project.sources || []).map((item) => <Tag key={item}>{item}</Tag>)}</div><div>{(project.genders || []).map((item) => <Tag key={item}>{item}</Tag>)}</div><Button type="primary" onClick={() => onOpenProject(project.id)}>进入项目</Button></Space></Card>)}</div>}
    </Card>
    <Modal title="新建批量" open={open} confirmLoading={creating} onCancel={() => !creating && setOpen(false)} onOk={() => form.submit()} okText="获取并创建项目">
      <Alert type="info" showIcon message="将依次创建 Intake、执行正文获取；只有全部完成后才创建 BatchProject。" style={{ marginBottom: 16 }} />
      <Form form={form} layout="vertical" onFinish={submit} initialValues={{ maxText: 4000 }}><Form.Item name="projectName" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}><Input /></Form.Item><Form.Item name="source" label="书城来源" rules={[{ required: true, message: '请选择书城来源' }]}><Select options={sourceOptions.map(({ value, label }) => ({ value, label }))} /></Form.Item><Form.Item name="bookId" label="Book ID" rules={[{ required: true, message: '请输入一个 Book ID' }]}><Input /></Form.Item><Form.Item name="maxText" label="单书正文长度"><InputNumber min={100} max={100000} style={{ width: '100%' }} /></Form.Item></Form>
    </Modal>
  </main>
}
