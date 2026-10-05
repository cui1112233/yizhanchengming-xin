import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Modal, Select, Space, Switch, Table, Tag, Typography } from 'antd'
import {
  getGenerationStage,
  getProjectGeneration,
  retryGenerationStage,
  runBookGeneration,
  runProjectGeneration,
} from './api.js'

const STAGES = [
  ['SCRIPT', '剧本'],
  ['HOOK', 'Hook'],
  ['DIRECTOR', 'Director'],
  ['FINAL_PROMPT', '最终提示词'],
]

function statusLabel(status) {
  const labels = {
    pending: '待执行',
    running: '执行中',
    completed: '已完成',
    failed: '失败',
    skipped: '已跳过',
  }
  return labels[status] || status || '待执行'
}

function statusColor(status) {
  if (status === 'completed') return 'success'
  if (status === 'running') return 'processing'
  if (status === 'failed') return 'error'
  if (status === 'skipped') return 'default'
  return 'warning'
}

function requestID(prefix) {
  if (globalThis.crypto?.randomUUID) return `${prefix}-${globalThis.crypto.randomUUID()}`
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export default function GenerationWorkbenchPage({ projectId }) {
  const [summary, setSummary] = useState(null)
  const [loading, setLoading] = useState(true)
  const [working, setWorking] = useState(false)
  const [error, setError] = useState('')
  const [hookEnabled, setHookEnabled] = useState(true)
  const [plotMode, setPlotMode] = useState(false)
  const [directorMode, setDirectorMode] = useState('normal')
  const [viewer, setViewer] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setSummary(await getProjectGeneration(projectId))
      setError('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '读取生成状态失败')
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    void load()
  }, [load])

  const executionOptions = useMemo(() => ({
    hookEnabled,
    plotMode,
    directorMode,
    // Task 13 只预留字段；本页不执行 matchAudio 时长重排。
    matchAudio: false,
    audioDurationSec: 0,
  }), [hookEnabled, plotMode, directorMode])

  const runBatch = async () => {
    setWorking(true)
    setError('')
    try {
      await runProjectGeneration(projectId, {
        ...executionOptions,
        requestId: requestID(`project-${projectId}`),
      })
      await load()
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '批量执行失败')
      await load()
    } finally {
      setWorking(false)
    }
  }

  const runBook = async (bookId) => {
    setWorking(true)
    setError('')
    try {
      await runBookGeneration(projectId, bookId, {
        ...executionOptions,
        requestId: requestID(`book-${bookId}`),
      })
      await load()
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '单本执行失败')
      await load()
    } finally {
      setWorking(false)
    }
  }

  const retry = async (bookId, stage) => {
    setWorking(true)
    setError('')
    try {
      await retryGenerationStage(projectId, bookId, stage, requestID(`retry-${bookId}-${stage}`))
      await load()
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Stage 重试失败')
      await load()
    } finally {
      setWorking(false)
    }
  }

  const viewStage = async (bookId, stage) => {
    try {
      const value = await getGenerationStage(projectId, bookId, stage)
      setViewer({ stage, value })
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '读取 Stage 结果失败')
    }
  }

  const stageColumn = ([stage, title]) => ({
    title,
    key: stage,
    width: stage === 'FINAL_PROMPT' ? 180 : 150,
    render: (_, row) => {
      const value = row.stages?.[stage]
      const status = value?.status || 'pending'
      return (
        <Space size={6} wrap>
          <Tag color={statusColor(status)}>{statusLabel(status)}</Tag>
          {(value?.outputText || value?.errorMessage) && (
            <Button type="link" size="small" onClick={() => void viewStage(row.bookId, stage)}>
              查看
            </Button>
          )}
          {status === 'failed' && (
            <Button danger type="link" size="small" disabled={working} onClick={() => void retry(row.bookId, stage)}>
              重试
            </Button>
          )}
        </Space>
      )
    },
  })

  const columns = [
    { title: 'Book ID', dataIndex: 'bookId', key: 'bookId', width: 100, fixed: 'left' },
    ...STAGES.map(stageColumn),
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, row) => (
        <Button size="small" type="primary" disabled={working || row.run?.status === 'running'} onClick={() => void runBook(row.bookId)}>
          单本执行
        </Button>
      ),
    },
  ]

  const viewerText = viewer?.value?.outputText || viewer?.value?.errorMessage || '暂无结果'

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Batch Factory · Task 12</Typography.Text>
          <Typography.Title level={2}>剧本 / Hook / Director / 最终提示词</Typography.Title>
          <Typography.Paragraph type="secondary">
            项目 #{projectId} 的真实 Go/MySQL Stage 状态。失败只重试目标 Stage，不重新执行小说获取。
          </Typography.Paragraph>
        </div>
      </div>

      {error && <Alert className="feedback" type="error" showIcon message={error} />}

      <Card title="生成设置" className="panel-card">
        <Space wrap size="large">
          <Space>Hook <Switch checked={hookEnabled} onChange={setHookEnabled} /></Space>
          <Space>剧情模式 <Switch checked={plotMode} onChange={setPlotMode} /></Space>
          <Space>
            Director
            <Select
              value={directorMode}
              onChange={setDirectorMode}
              style={{ width: 150 }}
              options={[
                { value: 'normal', label: '普通 Director' },
                { value: 'h3', label: 'H3 Director' },
              ]}
            />
          </Space>
          <Button type="primary" loading={working} onClick={() => void runBatch()}>
            批量执行
          </Button>
          <Button onClick={() => void load()} disabled={working}>刷新状态</Button>
        </Space>
      </Card>

      <Card title="每本小说生成状态" className="result-card" style={{ marginTop: 20 }}>
        <Table
          rowKey="bookId"
          columns={columns}
          dataSource={summary?.books || []}
          loading={loading}
          scroll={{ x: 920 }}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '暂无小说或尚未创建 BatchProject 数据' }}
        />
      </Card>

      <Modal
        title={viewer ? `${viewer.stage} 结果` : 'Stage 结果'}
        open={Boolean(viewer)}
        footer={<Button onClick={() => setViewer(null)}>关闭</Button>}
        onCancel={() => setViewer(null)}
        width={800}
      >
        <Typography.Paragraph style={{ whiteSpace: 'pre-wrap', maxHeight: 520, overflow: 'auto' }}>
          {viewerText}
        </Typography.Paragraph>
      </Modal>
    </main>
  )
}
