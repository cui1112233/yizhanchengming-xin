import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Empty, Input, Select, Space, Table, Tag, Typography } from 'antd'
import { getProjectGeneration, listBatchProjects } from './api.js'
import PageState from './ui/PageState.jsx'
import StatusTag from './ui/StatusTag.jsx'
import './tts-history.css'

function messageOf(error, fallback) { return error instanceof Error ? error.message : fallback }
function latestUpdate(book) { return Object.values(book?.stages || {}).map((stage) => stage?.updatedAt).filter(Boolean).sort().at(-1) || '' }
function formatDate(value) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '暂无阶段执行记录' }

export default function HistoryPage() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('all')
  const load = useCallback(async (showRefresh = false) => {
    if (showRefresh) setRefreshing(true); else setLoading(true)
    setError('')
    try {
      const payload = await listBatchProjects()
      const projects = payload?.projects || []
      const generations = await Promise.all(projects.map(async (project) => ({ project, generation: await getProjectGeneration(project.id) })))
      const next = generations.flatMap(({ project, generation }) => (generation?.books || []).map((book) => ({
        key: `${project.id}-${book.bookId}`, project, book, status: book.status || project.runStatus || 'pending', updatedAt: latestUpdate(book),
      }))).sort((a, b) => String(b.updatedAt).localeCompare(String(a.updatedAt)))
      setRows(next)
    } catch (reason) { setError(messageOf(reason, '读取项目历史失败')) } finally { setLoading(false); setRefreshing(false) }
  }, [])
  useEffect(() => { void load() }, [load])
  const visible = useMemo(() => rows.filter((item) => {
    const text = `${item.project.name || ''} ${item.book.title || ''} ${item.book.bookId || ''}`.toLowerCase()
    return (!keyword || text.includes(keyword.toLowerCase())) && (status === 'all' || item.status === status)
  }), [rows, keyword, status])
  if (loading) return <PageState state="loading" title="正在读取项目历史…" className="tts-history-state" />
  if (error && rows.length === 0) return <PageState state="failed" title="项目历史读取失败" description={error} onRetry={() => void load()} className="tts-history-state" />
  return <main className="tts-history-page"><section className="tts-history-heading"><div><Typography.Text type="secondary">一战晟铭 · 项目投影</Typography.Text><Typography.Title level={2}>历史记录</Typography.Title><Typography.Paragraph type="secondary">按服务端批量项目与生成阶段构建；服务端权限校验失败时会明确显示错误，不使用浏览器缓存回退。</Typography.Paragraph></div><Button onClick={() => void load(true)} loading={refreshing}>刷新</Button></section>{error && <Alert type="error" showIcon closable message={error} onClose={() => setError('')} />}<Card className="history-table-card"><Space wrap className="tts-history-controls"><Input.Search aria-label="搜索历史" placeholder="搜索项目、小说或 Book ID" value={keyword} onChange={(event) => setKeyword(event.target.value)} allowClear /><Select aria-label="筛选历史状态" value={status} onChange={setStatus} options={[{ value: 'all', label: '全部状态' }, ...['pending', 'running', 'completed', 'failed', 'retryable_failed'].map((value) => ({ value, label: value }))]} /></Space><Table rowKey="key" loading={refreshing} dataSource={visible} pagination={{ pageSize: 20, hideOnSinglePage: true }} locale={{ emptyText: <Empty description="暂无可查看的项目历史" /> }} columns={[{ title: '项目 / 小说', key: 'name', render: (_, row) => <Space direction="vertical" size={0}><Typography.Text strong>{row.book.title || `小说 #${row.book.bookId}`}</Typography.Text><Typography.Text type="secondary">{row.project.name || `项目 #${row.project.id}`} · Book #{row.book.bookId}</Typography.Text></Space> }, { title: '状态', key: 'status', width: 150, render: (_, row) => <StatusTag status={row.status} /> }, { title: '阶段', key: 'stages', render: (_, row) => <Space wrap>{Object.entries(row.book.stages || {}).map(([stage, value]) => <Tag key={stage} color={value?.status === 'completed' ? 'green' : value?.status === 'failed' ? 'red' : 'default'}>{stage}: {value?.status || 'pending'}</Tag>)}</Space> }, { title: '最后更新', key: 'updated', width: 210, render: (_, row) => formatDate(row.updatedAt) }, { title: '操作', key: 'open', width: 120, render: (_, row) => <Button size="small" href={`/shuihuo-production?projectId=${row.project.id}`}>打开项目</Button> }]} /></Card></main>
}
