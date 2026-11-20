import { Alert, Button, Empty, Progress, Table, Tag } from 'antd'
import { RefreshCw, RotateCcw, XCircle } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { TASK_COLUMNS, taskAction, taskStatus, taskTypeLabel } from './taskCenterModel.js'

export function TaskCenter() {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback((quiet = false) => {
    if (!quiet) setLoading(true)
    return apiRequest('/api/magic-cut/tasks').then(payload => { setTasks(payload.tasks || []); setError('') }).catch(caught => setError(caught.message || '任务列表读取失败')).finally(() => setLoading(false))
  }, [])

  useEffect(() => { reload(); const timer = window.setInterval(() => reload(true), 3000); return () => window.clearInterval(timer) }, [reload])

  async function act(task, action) {
    try { await apiRequest(`/api/magic-cut/tasks/${task.id}/${action}`, { method: 'POST' }); await reload(true) }
    catch (caught) { setError(caught.message || '任务操作失败') }
  }

  const columns = [
    { title: TASK_COLUMNS[0], dataIndex: 'type', render: value => taskTypeLabel(value) },
    { title: TASK_COLUMNS[1], dataIndex: 'projectId', render: value => value || '—' },
    { title: TASK_COLUMNS[2], dataIndex: 'status', render: value => { const status = taskStatus(value); return <Tag color={status.color}>{status.label}</Tag> } },
    { title: TASK_COLUMNS[3], dataIndex: 'progress', render: value => <Progress percent={Math.round((Number(value) || 0) * 100)} size="small" /> },
    { title: TASK_COLUMNS[4], dataIndex: 'updatedAt', render: value => value ? new Date(value).toLocaleString('zh-CN') : '—' },
    { title: TASK_COLUMNS[5], key: 'action', render: (_, task) => { const action = taskAction(task); return action === 'cancel' ? <Button size="small" danger icon={<XCircle size={14} />} onClick={() => act(task, action)}>取消</Button> : action === 'retry' ? <Button size="small" icon={<RotateCcw size={14} />} onClick={() => act(task, action)}>重试</Button> : '—' } },
  ]

  return <main className="task-center"><header><div><p>MAGIC CUT TASKS</p><h1>任务中心</h1></div><Button icon={<RefreshCw size={16} />} onClick={() => reload()}>刷新</Button></header>{error ? <Alert type="error" showIcon message={error} /> : null}<Table rowKey="id" loading={loading} columns={columns} dataSource={tasks} locale={{ emptyText: <Empty description="暂无任务" /> }} expandable={{ expandedRowRender: task => <div className="task-detail">{task.errorMessage ? <Alert type="error" message={task.errorMessage} description={task.errorCode} /> : <span>任务编号：{task.id}</span>}</div> }} /></main>
}
