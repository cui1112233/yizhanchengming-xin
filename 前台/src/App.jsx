import React, { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Result,
  Row,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import { createBatchProject, createIntake, executeIntake, listBooks } from './api.js'
import BatchProjectListPage from './BatchProjectListPage.jsx'
import './app.css'

const { TextArea } = Input

export function parseBookIds(raw) {
  return [...new Set(
    String(raw || '')
      .split(/[\s,，;；]+/)
      .map((value) => value.trim())
      .filter(Boolean),
  )]
}

function statusLabel(status) {
  const labels = {
    pending: '待处理',
    running: '处理中',
    completed: '已完成',
    partial_failed: '部分失败',
    failed: '失败',
    fetched: '已获取',
    retryable_failed: '可重试失败',
    queued: '排队中',
  }
  return labels[status] || status || '-'
}

function statusColor(status) {
  if (status === 'completed' || status === 'fetched') return 'success'
  if (status === 'partial_failed' || status === 'retryable_failed') return 'warning'
  if (status === 'failed') return 'error'
  if (status === 'running' || status === 'queued') return 'processing'
  return 'default'
}

function buildBatchName(inputName) {
  const trimmed = inputName.trim()
  if (trimmed) return trimmed
  return `小说获取批次 ${new Date().toLocaleString('zh-CN', { hour12: false })}`
}

function safeErrorMessage(error) {
  const message = error instanceof Error ? error.message : String(error || '执行失败')
  if (/(password|passwd|token|authorization|mysql:\/\/|dsn)/i.test(message)) {
    return '请求失败，请稍后重试或查看服务端日志。'
  }
  return message || '执行失败'
}

class WorkbenchErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { error: null }
  }

  static getDerivedStateFromError(error) {
    return { error }
  }

  componentDidCatch(error) {
    console.error('Task 11 page render failed', error)
  }

  render() {
    if (this.state.error) {
      return (
        <Result
          status="error"
          title="页面加载失败"
          subTitle="水货生产页面遇到异常，请刷新后重试。若问题持续，请查看服务端日志。"
        />
      )
    }
    return this.props.children
  }
}

export default function IntakeWorkbench() {
  const [pathname, setPathname] = useState(() => (
    typeof window === 'undefined' ? '/shuihuo-production' : window.location.pathname
  ))

  useEffect(() => {
    const handlePopState = () => setPathname(window.location.pathname)
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const navigate = (path) => {
    if (window.location.pathname !== path) {
      window.history.pushState({}, '', path)
    }
    setPathname(path)
  }

  let page = <NovelIntakeWorkbench onNavigate={navigate} />
  if (pathname === '/batch-factory') {
    page = <BatchProjectListPage />
  }

  return <WorkbenchErrorBoundary>{page}</WorkbenchErrorBoundary>
}

function NovelIntakeWorkbench({ onNavigate }) {
  const [groups, setGroups] = useState([])
  const [source, setSource] = useState('')
  const [platformId, setPlatformId] = useState('')
  const [bookIdText, setBookIdText] = useState('')
  const [batchName, setBatchName] = useState('')
  const [maxText, setMaxText] = useState(4000)
  const [books, setBooks] = useState([])
  const [summary, setSummary] = useState(null)
  const [project, setProject] = useState(null)
  const [run, setRun] = useState(null)
  const [feedback, setFeedback] = useState(null)
  const [working, setWorking] = useState(false)
  const [scheduleOpen, setScheduleOpen] = useState(false)
  const [scheduleValue, setScheduleValue] = useState('')

  const totalBooks = useMemo(
    () => groups.reduce((total, group) => total + group.books.length, 0),
    [groups],
  )

  const addGroup = () => {
    const cleanSource = source.trim()
    const cleanPlatformId = platformId.trim()
    const ids = parseBookIds(bookIdText)

    if (!cleanSource) {
      setFeedback({ type: 'error', message: '请先填写书城名称。' })
      return
    }
    if (!cleanPlatformId) {
      setFeedback({ type: 'error', message: '请填写该书城对应的 121 platformId。' })
      return
    }
    if (ids.length === 0) {
      setFeedback({ type: 'error', message: '请至少填写一个 Book ID。' })
      return
    }

    const existingGroup = groups.find((group) => group.source === cleanSource)
    if (existingGroup && existingGroup.platformId !== cleanPlatformId) {
      setFeedback({
        type: 'error',
        message: `${cleanSource} 已使用 platformId ${existingGroup.platformId}，同一书城不能混用不同 platformId。`,
      })
      return
    }

    const existing = new Set(
      groups
        .filter((group) => group.source === cleanSource)
        .flatMap((group) => group.books.map((book) => book.bookId)),
    )
    const uniqueIds = ids.filter((id) => !existing.has(id))
    if (uniqueIds.length === 0) {
      setFeedback({ type: 'warning', message: '这些 Book ID 已经添加到同一书城，无需重复添加。' })
      return
    }
    if (totalBooks + uniqueIds.length > 200) {
      setFeedback({ type: 'error', message: '单个批次最多 200 本小说。' })
      return
    }

    setGroups((current) => {
      const groupIndex = current.findIndex((group) => group.source === cleanSource)
      if (groupIndex === -1) {
        return [
          ...current,
          {
            key: `${Date.now()}-${current.length}`,
            source: cleanSource,
            platformId: cleanPlatformId,
            books: uniqueIds.map((bookId) => ({ bookId })),
          },
        ]
      }

      return current.map((group, index) => (
        index === groupIndex
          ? { ...group, books: [...group.books, ...uniqueIds.map((bookId) => ({ bookId }))] }
          : group
      ))
    })
    setBookIdText('')
    setFeedback({
      type: 'success',
      message: `已添加 ${cleanSource} ${uniqueIds.length} 本，可继续切换书城添加下一组。`,
    })
  }

  const removeGroup = (key) => {
    setGroups((current) => current.filter((group) => group.key !== key))
  }

  const runWorkflow = async (runAt = '') => {
    if (groups.length === 0) {
      setFeedback({ type: 'error', message: '请先添加至少一个书城。' })
      return
    }

    setWorking(true)
    setBooks([])
    setSummary(null)
    setProject(null)
    setRun(null)
    setFeedback({ type: 'info', message: '正在创建小说获取批次…' })

    let intakeId = null
    try {
      const name = buildBatchName(batchName)
      const created = await createIntake({
        name,
        groups: groups.map((group) => ({
          source: group.source,
          platformId: group.platformId,
          books: group.books,
        })),
      })
      intakeId = created.intake.id
      setBooks(created.books || [])

      setFeedback({ type: 'info', message: '批次已创建，正在通过 121 获取正文并分析信息…' })
      const execution = await executeIntake(intakeId, maxText)
      setSummary(execution)

      const bookPayload = await listBooks(intakeId)
      setBooks(bookPayload.books || [])

      if (execution.status !== 'completed') {
        setFeedback({
          type: execution.status === 'partial_failed' ? 'warning' : 'error',
          message: `本批次未全部成功（成功 ${execution.fetched}，失败 ${execution.failed}），已停止创建生产任务。请先处理失败书籍。`,
        })
        return
      }

      setFeedback({ type: 'info', message: runAt ? '小说全部获取成功，正在创建自动化任务…' : '小说全部获取成功，正在创建立即执行任务…' })
      const batch = await createBatchProject(intakeId, {
        name,
        ...(runAt ? { runAt } : {}),
      })
      setProject(batch.project)
      setRun(batch.run)
      setFeedback({
        type: 'success',
        message: runAt ? '自动化任务已创建。' : '立即执行任务已创建。',
      })
    } catch (error) {
      if (intakeId) {
        try {
          const bookPayload = await listBooks(intakeId)
          setBooks(bookPayload.books || [])
        } catch {
          // 保留主错误信息；结果读取失败不覆盖原始错误。
        }
      }
      setFeedback({ type: 'error', message: safeErrorMessage(error) })
    } finally {
      setWorking(false)
    }
  }

  const confirmSchedule = () => {
    if (!scheduleValue) {
      setFeedback({ type: 'error', message: '请选择自动化执行时间。' })
      return
    }
    const date = new Date(scheduleValue)
    if (Number.isNaN(date.getTime()) || date.getTime() <= Date.now()) {
      setFeedback({ type: 'error', message: '自动化执行时间必须晚于当前时间。' })
      return
    }
    setScheduleOpen(false)
    void runWorkflow(date.toISOString())
  }

  const textColumn = (title, dataIndex, width, options = {}) => ({
    title,
    dataIndex,
    key: dataIndex,
    width,
    ...options,
    render: (value) => value || '-',
  })

  const columns = [
    textColumn('书城', 'source', 110, { fixed: 'left' }),
    textColumn('121平台', 'platformId', 90),
    textColumn('Book ID', 'bookId', 130),
    textColumn('书名', 'title', 180, { ellipsis: true }),
    textColumn('分类', 'category', 120),
    textColumn('类型', 'genre', 120),
    textColumn('男女频', 'gender', 90),
    textColumn('风格', 'style', 120),
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value) => <Tag color={statusColor(value)}>{statusLabel(value)}</Tag>,
    },
    {
      title: '错误',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
      width: 240,
      render: (value) => value ? <Typography.Text type="danger">{safeErrorMessage(value)}</Typography.Text> : '-',
    },
  ]

  return (
    <main className="page-shell">
      <div className="page-heading">
        <div>
          <Typography.Text type="secondary">一战晟铭 · Phase 1</Typography.Text>
          <Typography.Title level={2}>小说获取工作台</Typography.Title>
          <Typography.Paragraph type="secondary">
            按书城分组添加 Book ID，一次执行 121 正文获取、男女频/风格分析，并在全部成功后创建批量任务。
          </Typography.Paragraph>
          <Button onClick={() => onNavigate('/batch-factory')}>批量工厂</Button>
        </div>
        <Space size="large" className="heading-stats">
          <Statistic title="已选书城" value={groups.length} suffix="组" />
          <Statistic title="已选小说" value={totalBooks} suffix="本" />
        </Space>
      </div>

      {feedback && (
        <Alert
          className="feedback"
          type={feedback.type}
          message={feedback.message}
          showIcon
          closable
          onClose={() => setFeedback(null)}
        />
      )}

      <Row gutter={[20, 20]}>
        <Col xs={24} xl={10}>
          <Card title="1. 添加书城" className="panel-card">
            <Form layout="vertical">
              <Form.Item label="批次名称（可选）">
                <Input
                  value={batchName}
                  onChange={(event) => setBatchName(event.target.value)}
                  placeholder="不填则自动生成"
                  disabled={working}
                />
              </Form.Item>
              <Row gutter={12}>
                <Col span={14}>
                  <Form.Item label="书城名称" required>
                    <Input
                      value={source}
                      onChange={(event) => setSource(event.target.value)}
                      placeholder="例如：阳光、常读、知乎"
                      disabled={working}
                    />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item label="121 platformId" required>
                    <Input
                      value={platformId}
                      onChange={(event) => setPlatformId(event.target.value)}
                      placeholder="例如：4"
                      disabled={working}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Alert
                className="inline-note"
                type="info"
                showIcon
                message="已验证映射：阳光/YG = 4，常读/CD = 2。其他书城请填写其实际 121 platformId。"
              />
              <Form.Item label="Book ID" required>
                <TextArea
                  value={bookIdText}
                  onChange={(event) => setBookIdText(event.target.value)}
                  placeholder={'每行一个，也支持逗号/空格分隔\n例如：\n123456\n789012'}
                  autoSize={{ minRows: 6, maxRows: 12 }}
                  disabled={working}
                />
              </Form.Item>
              <Button type="primary" block onClick={addGroup} disabled={working}>
                添加书城
              </Button>
            </Form>

            <Divider />
            <Typography.Title level={5}>所选书城</Typography.Title>
            {groups.length === 0 ? (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有添加书城" />
            ) : (
              <Space wrap size={[8, 10]}>
                {groups.map((group) => (
                  <Tag
                    key={group.key}
                    closable={!working}
                    onClose={() => removeGroup(group.key)}
                    className="source-tag"
                  >
                    {group.source} {group.books.length}本 · P{group.platformId}
                  </Tag>
                ))}
              </Space>
            )}
          </Card>
        </Col>

        <Col xs={24} xl={14}>
          <Card title="2. 执行方式" className="panel-card">
            <Row gutter={12} align="bottom">
              <Col xs={24} md={8}>
                <Form.Item label="正文最大字符数" className="compact-form-item">
                  <InputNumber
                    min={100}
                    max={100000}
                    value={maxText}
                    onChange={(value) => setMaxText(value || 4000)}
                    disabled={working}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} md={16}>
                <Space wrap className="action-row">
                  <Button
                    type="primary"
                    size="large"
                    loading={working}
                    disabled={groups.length === 0}
                    onClick={() => void runWorkflow('')}
                  >
                    立即执行
                  </Button>
                  <Button
                    size="large"
                    disabled={working || groups.length === 0}
                    onClick={() => setScheduleOpen(true)}
                  >
                    自动化
                  </Button>
                </Space>
              </Col>
            </Row>

            <Divider />
            <Typography.Title level={5}>执行结果</Typography.Title>
            {!summary && !project ? (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="执行后这里会显示批次与 Run 状态" />
            ) : (
              <Space direction="vertical" size={14} style={{ width: '100%' }}>
                {summary && (
                  <Descriptions bordered size="small" column={{ xs: 1, sm: 2, md: 4 }}>
                    <Descriptions.Item label="Intake ID">{summary.intakeId}</Descriptions.Item>
                    <Descriptions.Item label="批次状态">
                      <Tag color={statusColor(summary.status)}>{statusLabel(summary.status)}</Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="获取成功">{summary.fetched}</Descriptions.Item>
                    <Descriptions.Item label="获取失败">{summary.failed}</Descriptions.Item>
                  </Descriptions>
                )}
                {project && run && (
                  <>
                    <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
                      <Descriptions.Item label="BatchProject">#{project.id} · {project.name}</Descriptions.Item>
                      <Descriptions.Item label="Run">
                        #{run.id} · {statusLabel(run.status)} · {new Date(run.runAt).toLocaleString('zh-CN', { hour12: false })}
                      </Descriptions.Item>
                    </Descriptions>
                    <Button type="primary" onClick={() => onNavigate('/batch-factory')}>
                      进入批量工厂
                    </Button>
                  </>
                )}
              </Space>
            )}
          </Card>
        </Col>
      </Row>

      <Card title="书籍结果" className="result-card">
        <Table
          rowKey={(record) => record.id || `${record.source}-${record.bookId}`}
          columns={columns}
          dataSource={books}
          scroll={{ x: 1320 }}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: '执行后显示 ID、书名、书城、男女频、风格和状态' }}
        />
      </Card>

      <Modal
        title="自动化执行"
        open={scheduleOpen}
        onCancel={() => setScheduleOpen(false)}
        onOk={confirmSchedule}
        okText="确认创建"
        cancelText="取消"
        confirmLoading={working}
      >
        <Typography.Paragraph type="secondary">
          选择未来时间。小说会先完成 121 获取和分析，全部成功后创建带 runAt 的自动化 Run。
        </Typography.Paragraph>
        <Input
          type="datetime-local"
          value={scheduleValue}
          onChange={(event) => setScheduleValue(event.target.value)}
        />
      </Modal>
    </main>
  )
}
