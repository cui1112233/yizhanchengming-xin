import React, { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Form,
  Input,
  InputNumber,
  Progress,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import { createBatchProject, createIntake, executeIntake, listBooks, listIntakes } from './api.js'
import StatusTag from './ui/StatusTag.jsx'
import PageState from './ui/PageState.jsx'

const { TextArea } = Input

export const NOVEL_FETCH_PLATFORMS = [
  { value: '1', label: '黑岩付费' },
  { value: '2', label: '番茄付费' },
  { value: '3', label: '七猫付费' },
  { value: '4', label: '点众付费' },
  { value: '7', label: '番茄免费' },
  { value: '15', label: '知乎付费' },
  { value: '20', label: '掌阅付费' },
  { value: '26', label: '卓越付费' },
  { value: '29', label: '九州书城' },
  { value: '31', label: '掌文付费' },
]

export function parseNovelFetchBookIds(raw) {
  return [...new Set(
    String(raw || '')
      .split(/[\s,，;；]+/)
      .map((value) => value.trim())
      .filter(Boolean),
  )]
}

export function summarizeNovelFetchBooks(books = []) {
  const total = books.length
  const fetched = books.filter((book) => book.status === 'fetched').length
  const failed = books.filter((book) => book.status === 'retryable_failed').length
  const pending = Math.max(0, total - fetched - failed)
  return {
    total,
    fetched,
    failed,
    pending,
    percent: total ? Math.round((fetched / total) * 100) : 0,
  }
}

function safeErrorMessage(error) {
  if (error?.status === 403) return '无权限访问小说获取任务。'
  if (error?.status === 401) return '登录状态已过期，请重新登录。'
  const message = error instanceof Error ? error.message : String(error || '请求失败')
  if (/(password|passwd|token|authorization|mysql:\/\/|dsn|secret|credential)/i.test(message)) {
    return '请求失败，请稍后重试或查看服务端日志。'
  }
  return message || '请求失败'
}

function buildBatchName(inputName) {
  const trimmed = String(inputName || '').trim()
  if (trimmed) return trimmed
  return '小说获取批次 ' + new Date().toLocaleString('zh-CN', { hour12: false })
}

function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', { hour12: false })
}

function firstFailure(books = []) {
  const row = books.find((book) => book.status === 'retryable_failed' && book.errorMessage)
  return row ? safeErrorMessage(row.errorMessage) : ''
}

export default function NovelFetchPage() {
  const [groups, setGroups] = useState([])
  const [platformId, setPlatformId] = useState('')
  const [bookIdText, setBookIdText] = useState('')
  const [batchName, setBatchName] = useState('')
  const [maxText, setMaxText] = useState(4000)
  const [books, setBooks] = useState([])
  const [summary, setSummary] = useState(null)
  const [project, setProject] = useState(null)
  const [run, setRun] = useState(null)
  const [history, setHistory] = useState([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [selectedIntakeId, setSelectedIntakeId] = useState(0)
  const [feedback, setFeedback] = useState(null)
  const [working, setWorking] = useState(false)
  const [retryingId, setRetryingId] = useState(0)
  const [accessDenied, setAccessDenied] = useState(false)

  const totalBooks = useMemo(
    () => groups.reduce((total, group) => total + group.books.length, 0),
    [groups],
  )

  const selectedPlatform = useMemo(
    () => NOVEL_FETCH_PLATFORMS.find((item) => item.value === platformId),
    [platformId],
  )

  const loadHistory = async (preferredId = 0) => {
    setHistoryLoading(true)
    try {
      const payload = await listIntakes()
      const intakes = Array.isArray(payload?.intakes) ? payload.intakes : []
      const rows = await Promise.all(intakes.map(async (item) => {
        try {
          const result = await listBooks(item.id)
          return { ...item, books: Array.isArray(result?.books) ? result.books : [], loadError: '' }
        } catch (error) {
          return { ...item, books: [], loadError: safeErrorMessage(error) }
        }
      }))
      setHistory(rows)
      setAccessDenied(false)
      const targetId = preferredId || selectedIntakeId || rows[0]?.id || 0
      if (targetId) {
        const target = rows.find((item) => item.id === targetId)
        if (target) {
          setSelectedIntakeId(target.id)
          setBooks(target.books)
        }
      } else {
        setBooks([])
      }
      return rows
    } catch (error) {
      if (error?.status === 403) {
        setAccessDenied(true)
        setHistory([])
        setBooks([])
      } else {
        setFeedback({ type: 'error', message: safeErrorMessage(error) })
      }
      return []
    } finally {
      setHistoryLoading(false)
    }
  }

  useEffect(() => {
    void loadHistory()
  }, [])

  const addGroup = () => {
    const ids = parseNovelFetchBookIds(bookIdText)
    if (!selectedPlatform) {
      setFeedback({ type: 'error', message: '请先选择小说来源 / 平台。' })
      return
    }
    if (ids.length === 0) {
      setFeedback({ type: 'error', message: '请至少填写一个 Book ID。' })
      return
    }

    const existing = new Set(
      groups
        .filter((group) => group.platformId === selectedPlatform.value)
        .flatMap((group) => group.books.map((book) => book.bookId)),
    )
    const uniqueIds = ids.filter((id) => !existing.has(id))
    if (uniqueIds.length === 0) {
      setFeedback({ type: 'warning', message: '这些 Book ID 已经添加到当前来源。' })
      return
    }
    if (totalBooks + uniqueIds.length > 200) {
      setFeedback({ type: 'error', message: '单个获取任务最多 200 本小说。' })
      return
    }

    setGroups((current) => {
      const index = current.findIndex((group) => group.platformId === selectedPlatform.value)
      if (index === -1) {
        return [...current, {
          key: selectedPlatform.value,
          source: selectedPlatform.label,
          platformId: selectedPlatform.value,
          books: uniqueIds.map((bookId) => ({ bookId })),
        }]
      }
      return current.map((group, itemIndex) => (
        itemIndex === index
          ? { ...group, books: [...group.books, ...uniqueIds.map((bookId) => ({ bookId }))] }
          : group
      ))
    })
    setBookIdText('')
    setFeedback({ type: 'success', message: '已添加 ' + selectedPlatform.label + ' ' + uniqueIds.length + ' 本，可继续切换来源添加。' })
  }

  const removeGroup = (key) => {
    setGroups((current) => current.filter((group) => group.key !== key))
  }

  const createProductionProject = async (intakeId, name) => {
    try {
      const batch = await createBatchProject(intakeId, { name })
      setProject(batch.project || null)
      setRun(batch.run || null)
      return true
    } catch (error) {
      setFeedback({
        type: 'warning',
        message: '小说获取已完成，但批量项目创建失败：' + safeErrorMessage(error),
      })
      return false
    }
  }

  const runWorkflow = async () => {
    if (groups.length === 0) {
      setFeedback({ type: 'error', message: '请先添加至少一个小说来源。' })
      return
    }

    setWorking(true)
    setSummary(null)
    setProject(null)
    setRun(null)
    let intakeId = 0
    try {
      const name = buildBatchName(batchName)
      setFeedback({ type: 'info', message: '正在创建小说获取任务…' })
      const created = await createIntake({
        name,
        groups: groups.map((group) => ({
          source: group.source,
          platformId: group.platformId,
          books: group.books,
        })),
      })
      intakeId = created?.intake?.id || 0
      setSelectedIntakeId(intakeId)
      setBooks(created?.books || [])

      setFeedback({ type: 'info', message: '任务已创建，正在通过 121 获取正文与元数据…' })
      const execution = await executeIntake(intakeId, maxText)
      setSummary(execution)
      const result = await listBooks(intakeId)
      setBooks(result?.books || [])

      if (execution.status === 'completed') {
        const projectCreated = await createProductionProject(intakeId, name)
        if (projectCreated) {
          setFeedback({ type: 'success', message: '小说获取完成，任务与书籍结果已持久化。' })
        }
      } else {
        setFeedback({
          type: execution.status === 'partial_failed' ? 'warning' : 'error',
          message: '本次获取成功 ' + execution.fetched + ' 本，失败 ' + execution.failed + ' 本。失败项可在任务记录中重试。',
        })
      }
      await loadHistory(intakeId)
    } catch (error) {
      if (intakeId) {
        try {
          const result = await listBooks(intakeId)
          setBooks(result?.books || [])
        } catch {
          // Preserve the primary server error.
        }
        await loadHistory(intakeId)
      }
      setFeedback({ type: error?.status === 403 ? 'warning' : 'error', message: safeErrorMessage(error) })
    } finally {
      setWorking(false)
    }
  }

  const retryIntake = async (row) => {
    setRetryingId(row.id)
    setSelectedIntakeId(row.id)
    setFeedback({ type: 'info', message: '正在重试失败项；已成功书籍不会重复获取。' })
    try {
      const execution = await executeIntake(row.id, maxText)
      setSummary(execution)
      const result = await listBooks(row.id)
      setBooks(result?.books || [])

      if (execution.status === 'completed') {
        await createProductionProject(row.id, row.name)
        setFeedback({ type: 'success', message: '失败项重试完成，批次已恢复为完成状态。' })
      } else {
        setFeedback({
          type: execution.status === 'partial_failed' ? 'warning' : 'error',
          message: '重试后仍有 ' + execution.failed + ' 本失败，请查看安全错误原因后再次重试。',
        })
      }
      await loadHistory(row.id)
    } catch (error) {
      setFeedback({ type: error?.status === 403 ? 'warning' : 'error', message: safeErrorMessage(error) })
    } finally {
      setRetryingId(0)
    }
  }

  const viewIntake = (row) => {
    setSelectedIntakeId(row.id)
    setBooks(row.books || [])
    setSummary({
      intakeId: row.id,
      status: row.status,
      fetched: summarizeNovelFetchBooks(row.books).fetched,
      failed: summarizeNovelFetchBooks(row.books).failed,
    })
  }

  const bookColumns = [
    { title: 'Book ID', dataIndex: 'bookId', key: 'bookId', width: 140, fixed: 'left', render: (value) => value || '-' },
    { title: '书名', dataIndex: 'title', key: 'title', width: 200, ellipsis: true, render: (value) => value || '-' },
    { title: '来源', dataIndex: 'source', key: 'source', width: 130, render: (value) => value || '-' },
    { title: '平台', dataIndex: 'platformId', key: 'platformId', width: 80, render: (value) => value || '-' },
    { title: '分类', dataIndex: 'category', key: 'category', width: 130, render: (value) => value || '-' },
    { title: '类型', dataIndex: 'genre', key: 'genre', width: 130, render: (value) => value || '-' },
    { title: '男女频', dataIndex: 'gender', key: 'gender', width: 90, render: (value) => value || '-' },
    { title: '风格', dataIndex: 'style', key: 'style', width: 130, render: (value) => value || '-' },
    { title: '状态', dataIndex: 'status', key: 'status', width: 130, render: (value) => <StatusTag status={value} /> },
    {
      title: '错误',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
      width: 260,
      render: (value) => value ? <Typography.Text type="danger">{safeErrorMessage(value)}</Typography.Text> : '-',
    },
  ]

  const historyColumns = [
    { title: '任务', dataIndex: 'name', key: 'name', width: 220, ellipsis: true },
    {
      title: '来源',
      key: 'sources',
      width: 220,
      render: (_, row) => {
        const values = [...new Set((row.books || []).map((book) => book.source).filter(Boolean))]
        return values.length ? values.map((value) => <Tag key={value}>{value}</Tag>) : '-'
      },
    },
    {
      title: '进度',
      key: 'progress',
      width: 230,
      render: (_, row) => {
        const stats = summarizeNovelFetchBooks(row.books)
        return (
          <div>
            <Progress percent={stats.percent} size="small" status={stats.failed ? 'exception' : stats.percent === 100 ? 'success' : 'active'} />
            <Typography.Text type="secondary">成功 {stats.fetched}/{stats.total} · 失败 {stats.failed} · 待处理 {stats.pending}</Typography.Text>
          </div>
        )
      },
    },
    { title: '状态', dataIndex: 'status', key: 'status', width: 130, render: (value) => <StatusTag status={value} /> },
    {
      title: '失败原因',
      key: 'error',
      width: 280,
      render: (_, row) => {
        const value = row.loadError || firstFailure(row.books)
        return value ? <Typography.Text type="danger">{value}</Typography.Text> : '-'
      },
    },
    { title: '更新时间', dataIndex: 'updatedAt', key: 'updatedAt', width: 180, render: formatTime },
    {
      title: '操作',
      key: 'actions',
      fixed: 'right',
      width: 210,
      render: (_, row) => {
        const stats = summarizeNovelFetchBooks(row.books)
        return (
          <Space>
            <Button size="small" onClick={() => viewIntake(row)}>查看结果</Button>
            <Button
              size="small"
              type={stats.failed ? 'primary' : 'default'}
              disabled={!stats.failed}
              loading={retryingId === row.id}
              onClick={() => void retryIntake(row)}
            >
              重试失败项
            </Button>
          </Space>
        )
      },
    },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Novel Fetch</Typography.Text>
          <Typography.Title level={2}>小说获取</Typography.Title>
          <Typography.Paragraph type="secondary">
            选择小说来源并批量添加 Book ID。任务、进度和结果以 Go API + MySQL 为事实源，刷新页面后会自动恢复。
          </Typography.Paragraph>
        </div>
        <Space size="large" className="heading-stats">
          <Statistic title="本次来源" value={groups.length} suffix="个" />
          <Statistic title="本次小说" value={totalBooks} suffix="本" />
        </Space>
      </div>

      {feedback ? (
        <Alert className="feedback" type={feedback.type} message={feedback.message} showIcon closable onClose={() => setFeedback(null)} />
      ) : null}
      {accessDenied ? (
        <Alert className="feedback" type="warning" showIcon message="无权限查看小说获取任务。" description="当前登录仍保留；请联系管理员授予 batch.view 权限。" />
      ) : null}

      <Row gutter={[20, 20]}>
        <Col xs={24} xl={10}>
          <Card title="1. 小说输入与来源" className="panel-card">
            <Form layout="vertical">
              <Form.Item label="任务名称（可选）">
                <Input value={batchName} onChange={(event) => setBatchName(event.target.value)} placeholder="不填则自动生成" disabled={working} />
              </Form.Item>
              <Form.Item label="来源 / 平台" required>
                <Select
                  showSearch
                  optionFilterProp="label"
                  value={platformId || undefined}
                  onChange={setPlatformId}
                  options={NOVEL_FETCH_PLATFORMS}
                  placeholder="选择小说来源"
                  disabled={working}
                />
              </Form.Item>
              <Form.Item label="Book ID" required>
                <TextArea
                  value={bookIdText}
                  onChange={(event) => setBookIdText(event.target.value)}
                  placeholder={'每行一个，也支持逗号 / 空格分隔\n例如：\n123456\n789012'}
                  autoSize={{ minRows: 6, maxRows: 12 }}
                  disabled={working}
                />
              </Form.Item>
              <Button type="primary" block onClick={addGroup} disabled={working}>添加到本次任务</Button>
            </Form>
            <Divider />
            <Typography.Title level={5}>已添加来源</Typography.Title>
            {groups.length ? (
              <Space wrap size={[8, 10]}>
                {groups.map((group) => (
                  <Tag key={group.key} closable={!working} onClose={() => removeGroup(group.key)} className="source-tag">
                    {group.source} {group.books.length} 本 · P{group.platformId}
                  </Tag>
                ))}
              </Space>
            ) : <PageState title="还没有添加小说来源" />}
          </Card>
        </Col>

        <Col xs={24} xl={14}>
          <Card title="2. 获取任务" className="panel-card">
            <Row gutter={12} align="bottom">
              <Col xs={24} md={9}>
                <Form.Item label="正文最大字符数" className="compact-form-item">
                  <InputNumber min={100} max={100000} value={maxText} onChange={(value) => setMaxText(value || 4000)} disabled={working} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col xs={24} md={15}>
                <Space wrap className="action-row">
                  <Button type="primary" size="large" loading={working} disabled={!groups.length} onClick={() => void runWorkflow()}>
                    开始获取
                  </Button>
                  <Button size="large" loading={historyLoading} onClick={() => void loadHistory(selectedIntakeId)}>
                    刷新任务
                  </Button>
                </Space>
              </Col>
            </Row>
            <Divider />
            {summary ? (
              <Descriptions bordered size="small" column={{ xs: 1, sm: 2, md: 4 }}>
                <Descriptions.Item label="Intake ID">{summary.intakeId}</Descriptions.Item>
                <Descriptions.Item label="任务状态"><StatusTag status={summary.status} /></Descriptions.Item>
                <Descriptions.Item label="获取成功">{summary.fetched}</Descriptions.Item>
                <Descriptions.Item label="获取失败">{summary.failed}</Descriptions.Item>
              </Descriptions>
            ) : <PageState title="创建任务后这里显示实时状态" />}
            {project && run ? (
              <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }} style={{ marginTop: 14 }}>
                <Descriptions.Item label="BatchProject">#{project.id} · {project.name}</Descriptions.Item>
                <Descriptions.Item label="Run">#{run.id} · <StatusTag status={run.status} /></Descriptions.Item>
              </Descriptions>
            ) : null}
          </Card>

          <Card title="AI 处理配置与规则" className="panel-card" style={{ marginTop: 20 }}>
            <Typography.Paragraph type="secondary">
              Workshop 本轮不迁移；这里只保留正确入口，不在 Novel Fetch 主页面复制第二套 AI 配置或规则数据。
            </Typography.Paragraph>
            <Space wrap>
              <Button href="/novel-fetch-workshop">AI 处理配置</Button>
              <Button href="/novel-fetch-workshop">处理规则</Button>
            </Space>
          </Card>
        </Col>
      </Row>

      <Card title="获取任务记录" className="result-card">
        <Table
          rowKey="id"
          loading={historyLoading}
          columns={historyColumns}
          dataSource={history}
          scroll={{ x: 1400 }}
          pagination={{ pageSize: 10, hideOnSinglePage: true }}
          locale={{ emptyText: accessDenied ? '当前账号无权限查看任务' : '暂无已创建的小说获取任务' }}
        />
      </Card>

      <Card title={selectedIntakeId ? '小说结果 · Intake #' + selectedIntakeId : '小说结果'} className="result-card">
        <Table
          rowKey={(record) => record.id || record.source + '-' + record.bookId}
          columns={bookColumns}
          dataSource={books}
          scroll={{ x: 1480 }}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '选择任务或开始获取后显示 Book ID、书名、分类、类型、男女频、风格、状态和错误' }}
        />
      </Card>
    </main>
  )
}
