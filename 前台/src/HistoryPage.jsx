import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Empty, Input, Select, Table, Tag, Typography } from 'antd'
import { listWorkspaceHistory } from './api.js'
import PageState from './ui/PageState.jsx'
import './history.css'

const kinds = { intake: '小说获取', batch: '批量项目', book_run: '小说运行', script: '剧本', novel_panel: '小说分镜', tts: '音频', shuihuo_image: '水货图片', shuihuo_video: '水货视频', video: '视频生产', merge: '视频合并' }
const statuses = { completed: '已完成', pending: '待执行', queued: '排队中', scheduled: '待调度', running: '执行中', partial_failed: '部分失败', failed: '失败', retryable_failed: '可重试失败', cancelled: '已取消', skipped: '已跳过', saved: '已保存', measured: '已测量', pending_executor: '等待执行器', unknown: '未知状态' }
const options = (values, label) => [{ value: '', label }, ...Object.entries(values).map(([value, text]) => ({ value, label: text }))]
function safeHref(row) {
  return row.kind === 'batch' && /^\/batch-factory\?projectId=[1-9]\d*$/.test(row.href || '') && row.href === `/batch-factory?projectId=${row.projectId}` ? row.href : ''
}
function historyError(error) {
  const message = { 401: '登录已过期，请重新登录', 403: '没有权限查看历史记录', 503: '历史服务暂不可用' }[error?.status] || error?.message || '读取历史记录失败'
  return { message, requestId: error?.requestId || '' }
}
function HistoryStatus({ value }) {
  const color = value === 'completed' ? 'success' : ['failed', 'retryable_failed', 'partial_failed'].includes(value) ? 'error' : ['running', 'queued'].includes(value) ? 'processing' : 'default'
  return <Tag color={color}>{statuses[value] || statuses.unknown}</Tag>
}

export default function HistoryPage() {
  const [query, setQuery] = useState({ page: 1, limit: 20, q: '', kind: '', status: '', archived: 'all' })
  const [keyword, setKeyword] = useState('')
  const [state, setState] = useState({ loading: true, error: null, entries: [], page: 1, limit: 20, total: 0 })
  const sequence = useRef(0)
  const load = useCallback(async () => {
    const request = ++sequence.current
    setState(previous => ({ ...previous, loading: true, error: null }))
    try {
      const data = await listWorkspaceHistory(query)
      if (request !== sequence.current) return
      setState({ loading: false, error: null, entries: data.entries || [], page: data.page || query.page, limit: data.limit || query.limit, total: data.total || 0 })
    } catch (error) {
      if (request !== sequence.current) return
      setState(previous => ({ ...previous, loading: false, error: historyError(error) }))
    }
  }, [query])
  useEffect(() => { void load(); return () => { sequence.current += 1 } }, [load])
  const filter = (field, value) => setQuery(previous => ({ ...previous, [field]: value, page: 1 }))
  const columns = [
    { title: '类型', dataIndex: 'kind', width: 118, render: value => kinds[value] || '未知类型' },
    { title: '项目 / 作品', width: 262, render: (_, row) => <div className="history-title"><Typography.Text strong>{row.title}</Typography.Text>{row.projectName && row.projectName !== row.title ? <Typography.Text type="secondary">{row.projectName}</Typography.Text> : null}{row.attempt > 0 || row.revision > 0 ? <Typography.Text type="secondary">{row.attempt > 0 ? `尝试 ${row.attempt}` : `修订 ${row.revision}`}</Typography.Text> : null}{row.archivedAt ? <Typography.Text type="secondary">已归档 · 只读</Typography.Text> : null}</div> },
    { title: '状态', dataIndex: 'status', width: 262, render: value => <HistoryStatus value={value} /> },
    { title: '最后更新', dataIndex: 'updatedAt', width: 190, render: value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '暂无时间记录' },
    { title: '操作', width: 180, render: (_, row) => safeHref(row) ? <Button size="small" href={safeHref(row)}>查看项目</Button> : <Typography.Text type="secondary">暂不支持精确进入</Typography.Text> },
  ]
  return <main className="history-page">
    <header className="history-heading"><div><Typography.Title level={2}>历史记录</Typography.Title><Typography.Paragraph type="secondary">查看小说获取、项目、执行尝试与修订记录。归档记录只读可见。</Typography.Paragraph></div><Button aria-label="刷新" onClick={() => void load()} loading={state.loading}>刷新</Button></header>
    <section className="history-controls" aria-label="历史筛选">
      <Input.Search aria-label="搜索历史" placeholder="搜索项目、作品或记录 ID" maxLength={200} value={keyword} onChange={event => setKeyword(event.target.value)} onSearch={value => filter('q', value.trim())} allowClear />
      <Select aria-label="筛选历史类型" value={query.kind} onChange={value => filter('kind', value)} options={options(kinds, '全部类型')} />
      <Select aria-label="筛选历史状态" value={query.status} onChange={value => filter('status', value)} options={options(statuses, '全部状态')} />
      <Select aria-label="筛选历史归档" value={query.archived} onChange={value => filter('archived', value)} options={[{ value: 'all', label: '全部记录' }, { value: 'active', label: '未归档' }, { value: 'archived', label: '已归档' }]} />
    </section>
    {state.error && !state.entries.length ? <PageState state="failed" title="历史记录读取失败" description={state.error.message} requestId={state.error.requestId} onRetry={() => void load()} className="history-state" /> : <>
      {state.error ? <Alert type="error" showIcon message={state.error.message} description={<><div>保留上次成功读取的记录，可能与当前筛选不一致。</div>{state.error.requestId ? <div>请求编号：{state.error.requestId}</div> : null}</>} action={<Button onClick={() => void load()}>重试</Button>} /> : null}
      <Table className="history-table" rowKey="id" loading={state.loading} dataSource={state.entries} columns={columns} scroll={{ x: 1012 }} pagination={{ current: state.page, pageSize: state.limit, total: state.total, hideOnSinglePage: true, showSizeChanger: false, onChange: page => setQuery(previous => ({ ...previous, page })), showTotal: total => `共 ${total} 条` }} locale={{ emptyText: state.loading ? <span role="status">正在读取历史记录…</span> : <Empty description="暂无可查看的历史记录" /> }} />
    </>}
  </main>
}
