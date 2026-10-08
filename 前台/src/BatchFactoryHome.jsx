import React, { useEffect, useRef, useState } from 'react'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, Modal, Pagination, Select, Space, Spin, Tag, Typography } from 'antd'
import { archiveBatchProject, createBatchProject, createIntake, executeIntake, listBatchProjects, restoreBatchProject } from './api.js'
import { NOVEL_FETCH_PLATFORMS } from './NovelFetchPage.jsx'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import { batchError, batchUpdatedAt } from './batchFactoryPresentation.js'
import './batch-factory.css'

const sourceOptions = NOVEL_FETCH_PLATFORMS.map((item) => ({ value: item.label, label: item.label, platformId: item.value }))

function parseBookIDs(value) {
  return [...new Set(String(value || '').split(/[\s,，;；\n]+/).map((item) => item.trim()).filter(Boolean))]
}

export default function BatchFactoryHome({ onOpenProject, initialError = '', initialArchived = 'active', onArchivedChange }) {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [mutationError, setMutationError] = useState(null)
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState({ q: '', source: '', status: '', archived: initialArchived, page: 1, limit: 12, sort: 'updated_desc' })
  const [total, setTotal] = useState(0)
  const [refreshKey, setRefreshKey] = useState(0)
  const [view, setView] = useState('grid')
  const [creating, setCreating] = useState(false)
  const [open, setOpen] = useState(false)
  const [archiveTarget, setArchiveTarget] = useState(null)
  const [changingId, setChangingId] = useState(null)
  const request = useRef(0)
  const [form] = Form.useForm()
  const refresh = () => setRefreshKey((value) => value + 1)
  const changeFilter = (key, value) => setFilters((current) => ({ ...current, [key]: value || '', page: 1 }))

  useEffect(() => {
    const timer = setTimeout(() => setFilters((current) => current.q === query.trim() ? current : { ...current, q: query.trim(), page: 1 }), 300)
    return () => clearTimeout(timer)
  }, [query])
  useEffect(() => { setFilters((current) => current.archived === initialArchived ? current : { ...current, archived: initialArchived, page: 1 }) }, [initialArchived])
  useEffect(() => {
    const id = ++request.current
    setLoading(true); setError(null)
    listBatchProjects(filters).then((payload) => {
      if (id !== request.current) return
      const count = Number(payload?.total) || 0
      const lastPage = Math.max(1, Math.ceil(count / 12))
      if (filters.page > lastPage) { setFilters((current) => ({ ...current, page: lastPage })); return }
      setProjects(Array.isArray(payload?.projects) ? payload.projects : [])
      setTotal(count)
    }).catch((reason) => {
      if (id !== request.current) return
      if (reason?.status === 403) { setProjects([]); setTotal(0) }
      setError(batchError(reason, '批量项目读取失败，请稍后重试。'))
    }).finally(() => { if (id === request.current) setLoading(false) })
    return () => { request.current++ }
  }, [filters, refreshKey])

  const changeArchived = (archived) => {
    changeFilter('archived', archived)
    onArchivedChange?.(archived)
  }
  const changeArchive = async (project, restore = false) => {
    setChangingId(project.id); setMutationError(null)
    try {
      await (restore ? restoreBatchProject(project.id) : archiveBatchProject(project.id))
      setArchiveTarget(null)
      refresh()
    } catch (reason) {
      setMutationError(batchError(reason, restore ? '恢复项目失败，请稍后重试。' : '归档项目失败，请稍后重试。'))
      setArchiveTarget(null)
    } finally { setChangingId(null) }
  }

  const submit = async (values) => {
    const groups = (values.groups || []).map((group) => {
      const selected = sourceOptions.find((item) => item.value === group.source)
      return {
        source: selected?.label || '',
        platformId: selected?.platformId || '',
        books: parseBookIDs(group.bookIds).map((bookId) => ({ bookId, title: '', gender: '', style: '' })),
      }
    }).filter((group) => group.source && group.platformId && group.books.length > 0)
    if (groups.length === 0) {
      setError({ message: '请至少添加一个书城分组及一本 Book ID。' })
      return
    }
    setCreating(true); setError(null)
    try {
      const created = await createIntake({
        name: values.intakeName || values.projectName,
        groups,
      })
      const intake = created?.intake
      if (!intake?.id) throw new Error('创建 Intake 后未返回 ID')
      const executed = await executeIntake(intake.id, values.maxText || 4000)
      if (executed?.status !== 'completed') throw new Error('Intake 未全部完成，不能创建 BatchProject；请到小说获取处理失败书籍后重试。')
      await createBatchProject(intake.id, { name: values.projectName })
      setOpen(false); form.resetFields(); refresh()
    } catch (reason) { setError(batchError(reason)) }
    finally { setCreating(false) }
  }

  return <main className="page-shell batch-factory-page">
    <div className="page-heading"><div><Typography.Text type="secondary">一战晟铭 · Batch Factory</Typography.Text><Typography.Title level={2}>批量工厂</Typography.Title><Typography.Paragraph type="secondary">管理批量项目，查看小说与生产进度。</Typography.Paragraph></div><Space wrap><Button onClick={refresh}>刷新</Button><Button type="primary" onClick={() => setOpen(true)}>新建批量</Button></Space></div>
    {initialError && <Alert type="warning" showIcon message={initialError} className="feedback" />}
    {mutationError && <Alert type="error" showIcon message={mutationError.message} description={mutationError.requestId ? `请求编号：${mutationError.requestId}` : undefined} className="feedback" />}
    <Card className="result-card" title="批量项目">
      <Space wrap className="batch-project-filters">
        <Button type={filters.archived === 'active' ? 'primary' : 'default'} onClick={() => changeArchived('active')}>活跃项目</Button>
        <Button type={filters.archived === 'archived' ? 'primary' : 'default'} onClick={() => changeArchived('archived')}>已归档</Button>
        <Input aria-label="搜索批量项目" placeholder="搜索项目或来源" value={query} onChange={(event) => setQuery(event.target.value)} style={{ width: 220 }} />
        <Select aria-label="来源筛选" allowClear placeholder="来源筛选" value={filters.source || undefined} onChange={(value) => changeFilter('source', value)} options={sourceOptions.map(({ value, label }) => ({ value, label }))} style={{ width: 150 }} />
        <Select aria-label="运行状态筛选" allowClear placeholder="运行状态" value={filters.status || undefined} onChange={(value) => changeFilter('status', value)} options={[{ value: 'pending', label: '待执行' }, { value: 'running', label: '执行中' }, { value: 'completed', label: '已完成' }, { value: 'failed', label: '失败' }]} style={{ width: 150 }} />
        <Select aria-label="排序方式" value={filters.sort} onChange={(value) => changeFilter('sort', value)} options={[{ value: 'updated_desc', label: '最近更新' }, { value: 'name_asc', label: '名称排序' }]} style={{ width: 130 }} />
        <Button onClick={() => setView(view === 'grid' ? 'list' : 'grid')}>{view === 'grid' ? '列表视图' : '网格视图'}</Button>
      </Space>
      {error && <PageState state="failed" title={error.message} requestId={error.requestId} onRetry={refresh} />}
      <Spin spinning={loading}>
        {projects.length === 0 ? (loading ? <div role="status">正在读取批量项目</div> : !error && <Empty description="暂无匹配的批量项目" />) : <div className={`batch-project-card-grid ${view === 'list' ? 'is-list' : ''}`}>{projects.map((project) => <article key={project.id} className="batch-project-card">
          <div className="batch-project-card-heading"><Typography.Title level={5}>{project.name || '未命名项目'}</Typography.Title><StatusTag status={project.runStatus || 'pending'} /></div>
          <Typography.Text>书籍数量：{project.bookCount ?? 0}</Typography.Text>
          <div>{(project.sources || []).map((item) => <Tag key={item}>{item}</Tag>)}{(project.genders || []).map((item) => <Tag key={item}>{item}</Tag>)}{(project.styles || []).map((item) => <Tag key={item}>{item}</Tag>)}</div>
          <Typography.Text type={project.failureCount ? 'danger' : 'secondary'}>失败：{project.failureCount ?? 0}</Typography.Text>
          <Typography.Text type="secondary">最近更新：{batchUpdatedAt(project.updatedAt || project.createdAt)}</Typography.Text>
          <Space wrap className="batch-project-card-actions"><Button type="primary" onClick={() => onOpenProject(project.id)}>{project.archivedAt ? '只读查看' : '进入项目'}</Button><Button disabled={changingId !== null} onClick={() => project.archivedAt ? void changeArchive(project, true) : setArchiveTarget(project)}>{project.archivedAt ? '恢复' : '归档'}</Button></Space>
        </article>)}</div>}
      </Spin>
      <Pagination className="batch-project-pagination" current={filters.page} total={total} pageSize={12} showSizeChanger={false} hideOnSinglePage onChange={(page) => setFilters((current) => ({ ...current, page }))} />
    </Card>
    <Modal title="归档项目" open={Boolean(archiveTarget)} okText="确认归档" cancelText="取消" confirmLoading={changingId !== null} onCancel={() => changingId === null && setArchiveTarget(null)} onOk={() => archiveTarget && void changeArchive(archiveTarget)}>
      <Typography.Paragraph>确认归档“{archiveTarget?.name || '未命名项目'}”？小说与生产记录会保留，恢复前项目将只读。</Typography.Paragraph>
    </Modal>
    <Modal title="新建批量" open={open} confirmLoading={creating} onCancel={() => !creating && setOpen(false)} onOk={() => form.submit()} okText="获取并创建项目">
      <Alert type="info" showIcon message="将依次创建 Intake、执行正文获取；只有全部完成后才创建 BatchProject。" style={{ marginBottom: 16 }} />
      <Form form={form} layout="vertical" onFinish={submit} initialValues={{ maxText: 4000, groups: [{ source: undefined, bookIds: '' }] }}>
        <Form.Item name="projectName" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}><Input /></Form.Item>
        <Form.List name="groups">
          {(fields, { add, remove }) => <Space direction="vertical" size={12} style={{ width: '100%' }}>
            {fields.map(({ key, ...field }, index) => <Card key={key} size="small" title={`书城分组 ${index + 1}`} extra={fields.length > 1 ? <Button type="link" danger onClick={() => remove(field.name)}>删除分组</Button> : null}>
              <Form.Item {...field} name={[field.name, 'source']} label="书城来源" rules={[{ required: true, message: '请选择书城来源' }]}><Select options={sourceOptions.map(({ value, label }) => ({ value, label }))} /></Form.Item>
              <Form.Item {...field} name={[field.name, 'bookIds']} label="Book ID" rules={[{ required: true, message: '请输入至少一个 Book ID' }]} extra="可用空格、逗号或换行批量录入；同组重复 ID 会自动去重。"><Input.TextArea aria-label={`Book ID 分组 ${index + 1}`} rows={3} placeholder="例如：1001, 1002\n1003" /></Form.Item>
            </Card>)}
            <Button onClick={() => add({ source: undefined, bookIds: '' })}>添加书城分组</Button>
          </Space>}
        </Form.List>
        <Form.Item name="maxText" label="单书正文长度"><InputNumber min={100} max={100000} style={{ width: '100%' }} /></Form.Item>
      </Form>
    </Modal>
  </main>
}
