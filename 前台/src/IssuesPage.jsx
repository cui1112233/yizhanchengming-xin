import React, { useCallback, useEffect, useState } from 'react'
import { Alert, Button, Empty, Input, Select, Space, Table, Tag, Typography } from 'antd'
import { listIssues } from './api.js'
import PageState from './ui/PageState.jsx'

export default function IssuesPage() {
  const [state, setState] = useState({ loading: true, refreshing: false, error: '', entries: [], page: 1, total: 0 })
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('')
  const [status, setStatus] = useState('')

  const load = useCallback(async (page = 1, refreshing = false) => {
    setState(value => ({ ...value, loading: !refreshing, refreshing, error: '' }))
    try {
      const result = await listIssues({ page, q: query, source, status })
      setState({ loading: false, refreshing: false, error: '', entries: result.entries || [], page: result.page || page, total: result.total || 0 })
    } catch (error) {
      setState(value => ({ ...value, loading: false, refreshing: false, error: error.message || '读取问题记录失败' }))
    }
  }, [query, source, status])

  useEffect(() => { void load(1) }, [load])

  if (state.loading && !state.entries.length) return <PageState state="loading" title="正在读取问题记录…" className="tts-history-state" />
  if (state.error && !state.entries.length) return <PageState state="failed" title="问题记录读取失败" description={state.error} onRetry={() => void load(1)} className="tts-history-state" />

  return <main className="tts-history-page">
    <section className="tts-history-heading">
      <div>
        <Typography.Text type="secondary">一战晟铭 · MySQL 执行事实投影</Typography.Text>
        <Typography.Title level={2}>问题日志</Typography.Title>
        <Typography.Paragraph type="secondary">仅显示当前账号可访问项目的失败执行事实；Provider 凭据、Cookie、Token 与内部错误细节已脱敏。</Typography.Paragraph>
      </div>
      <Button onClick={() => void load(1, true)} loading={state.refreshing}>刷新</Button>
    </section>
    {state.error && <Alert type="error" showIcon message={state.error} />}
    <Space wrap className="tts-history-controls">
      <Input.Search placeholder="搜索项目、小说或错误说明" value={query} onChange={event => setQuery(event.target.value)} onSearch={() => void load(1)} />
      <Select value={source} onChange={setSource} options={[{ value: '', label: '全部来源' }, { value: 'book_run', label: '执行任务' }, { value: 'stage_run', label: '生成阶段' }, { value: 'video_task', label: '视频任务' }, { value: 'media_task', label: '媒体任务' }]} />
      <Select aria-label="筛选问题状态" value={status} onChange={setStatus} options={[{ value: '', label: '全部状态' }, { value: 'failed', label: 'failed' }, { value: 'retryable_failed', label: 'retryable_failed' }]} />
    </Space>
    <Table
      rowKey="id"
      loading={state.refreshing}
      dataSource={state.entries}
      locale={{ emptyText: <Empty description="暂无可查看的问题记录" /> }}
      pagination={{ current: state.page, pageSize: 20, total: state.total, showTotal: total => `共 ${total} 条`, onChange: page => void load(page) }}
      columns={[
        { title: '来源', dataIndex: 'source', render: value => <Tag>{value}</Tag> },
        { title: '状态', dataIndex: 'status', render: value => <Tag color="red">{value}</Tag> },
        { title: '问题', dataIndex: 'message' },
        { title: '项目 / 小说', render: (_, row) => `#${row.projectId} / #${row.bookId}` },
        { title: '时间', dataIndex: 'at', render: value => new Date(value).toLocaleString('zh-CN', { hour12: false }) },
        { title: '操作', render: (_, row) => <Button size="small" href={`/shuihuo-production?projectId=${row.projectId}`}>打开项目</Button> },
      ]}
    />
  </main>
}
