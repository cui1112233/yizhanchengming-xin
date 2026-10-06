import React, { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
} from 'antd'
import { novelFetchClient } from './client.js'

const { TextArea } = Input

export function parseNovelIds(raw) {
  return [...new Set(String(raw || '').split(/[\s,，;；]+/).map((value) => value.trim()).filter(Boolean))]
}

const defaultConfig = {
  textModelId: '',
  maxText: 4000,
  targetVersions: ['original', 'ai1'],
  rewriteProfiles: {},
  sensitiveReplacements: {},
  chapterRemovePrefixes: [],
  trimLines: true,
  dropBlankLines: true,
}

function errorText(error) {
  const message = error instanceof Error ? error.message : String(error || '操作失败')
  return /(password|token|authorization|cookie|secret|dsn|mysql:\/\/)/i.test(message)
    ? '请求失败，敏感错误详情已隐藏。'
    : message
}

function statusTag(status) {
  const value = String(status || 'pending')
  const color = value === 'succeeded' ? 'success'
    : value.includes('failed') ? 'error'
      : value === 'scheduled' ? 'processing'
        : value === 'blocked' ? 'warning'
          : 'default'
  return <Tag color={color}>{value}</Tag>
}

export default function NovelFetchModule({ client = novelFetchClient }) {
  const [tab, setTab] = useState('process')
  const [batchName, setBatchName] = useState('')
  const [source, setSource] = useState('')
  const [platformId, setPlatformId] = useState('')
  const [bookIds, setBookIds] = useState('')
  const [groups, setGroups] = useState([])
  const [batch, setBatch] = useState(null)
  const [run, setRun] = useState(null)
  const [books, setBooks] = useState([])
  const [scheduleAt, setScheduleAt] = useState('')
  const [working, setWorking] = useState(false)
  const [feedback, setFeedback] = useState(null)
  const [config, setConfig] = useState(defaultConfig)
  const [knowledge, setKnowledge] = useState([])
  const [knowledgeDraft, setKnowledgeDraft] = useState({ kind: 'rewrite', title: '', content: '', enabled: true })
  const [ruleInput, setRuleInput] = useState('')
  const [ruleOutput, setRuleOutput] = useState('')
  const [records, setRecords] = useState([])
  const [publishingAccountId, setPublishingAccountId] = useState('')
  const [submitVersion, setSubmitVersion] = useState('original')

  const totalBooks = useMemo(() => groups.reduce((sum, group) => sum + group.books.length, 0), [groups])

  useEffect(() => {
    let active = true
    Promise.all([client.getConfig(), client.listKnowledge()])
      .then(([configPayload, knowledgePayload]) => {
        if (!active) return
        setConfig({ ...defaultConfig, ...(configPayload?.config || {}) })
        setKnowledge(knowledgePayload?.items || [])
      })
      .catch((error) => {
        if (active) setFeedback({ type: 'warning', message: `共享配置尚未接通：${errorText(error)}` })
      })
    return () => { active = false }
  }, [client])

  const addGroup = () => {
    const cleanSource = source.trim()
    const cleanPlatform = platformId.trim()
    const ids = parseNovelIds(bookIds)
    if (!cleanSource || !cleanPlatform || ids.length === 0) {
      setFeedback({ type: 'error', message: '请填写书城、platformId 和至少一个 Book ID。' })
      return
    }
    setGroups((current) => {
      const index = current.findIndex((item) => item.source === cleanSource && item.platformId === cleanPlatform)
      if (index < 0) {
        return [...current, { source: cleanSource, platformId: cleanPlatform, books: ids.map((bookId) => ({ bookId })) }]
      }
      const existing = new Set(current[index].books.map((item) => item.bookId))
      const next = ids.filter((id) => !existing.has(id)).map((bookId) => ({ bookId }))
      return current.map((item, itemIndex) => itemIndex === index ? { ...item, books: [...item.books, ...next] } : item)
    })
    setBookIds('')
    setFeedback({ type: 'success', message: `已加入 ${cleanSource} ${ids.length} 个 ID，可继续添加其他书城。` })
  }

  const start = async (scheduled) => {
    if (groups.length === 0) {
      setFeedback({ type: 'error', message: '请先添加至少一个书城。' })
      return
    }
    let iso = ''
    if (scheduled) {
      const date = new Date(scheduleAt)
      if (!scheduleAt || Number.isNaN(date.getTime()) || date.getTime() <= Date.now()) {
        setFeedback({ type: 'error', message: '定时时间必须是未来时间。' })
        return
      }
      iso = date.toISOString()
    }
    setWorking(true)
    try {
      const created = await client.createBatch({
        name: batchName.trim() || '小说获取批次',
        groups,
      })
      setBatch(created.batch)
      setBooks(created.books || [])
      const started = await client.startRun(created.batch.id, iso)
      setRun(started.run)
      setFeedback({
        type: 'success',
        message: scheduled ? '定时 Run 已创建并等待统一 Runtime 派发。' : 'Run 已进入统一 Runtime 队列。',
      })
      setTab('tasks')
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    } finally {
      setWorking(false)
    }
  }

  const refreshTasks = async () => {
    if (!run?.id) return
    try {
      const payload = await client.runBooks(run.id)
      setBooks(payload.books || [])
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const retryBook = async (book) => {
    if (!run?.id) return
    setWorking(true)
    try {
      await client.retryBook(run.id, book.key)
      await refreshTasks()
      setFeedback({ type: 'success', message: `${book.title || book.bookId} 已只重试失败阶段。` })
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    } finally {
      setWorking(false)
    }
  }

  const handoff = async () => {
    if (!run?.id) return
    try {
      const result = await client.handoff(run.id)
      setFeedback({ type: 'success', message: `已创建批量工厂项目 #${result.batchProjectId}。` })
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const requestSubmit = async (book) => {
    const account = Number(publishingAccountId)
    if (!Number.isInteger(account) || account <= 0) {
      setFeedback({ type: 'error', message: '请填写统一发布服务中的发布账号 ID。' })
      return
    }
    try {
      const result = await client.submitIntent(run.id, book.key, { version: submitVersion, publishingAccountId: account })
      setFeedback({ type: 'success', message: `发布意图 #${result.intentId} 已创建；真正发布由统一发布服务执行。` })
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const saveConfig = async () => {
    try {
      const payload = await client.saveConfig(config)
      setConfig({ ...defaultConfig, ...(payload.config || {}) })
      setFeedback({ type: 'success', message: '版本与处理配置已保存。' })
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const saveKnowledge = async () => {
    try {
      const payload = await client.upsertKnowledge({ ...knowledgeDraft, id: '' })
      setKnowledge((current) => [...current.filter((item) => item.id !== payload.item.id), payload.item])
      setKnowledgeDraft({ kind: 'rewrite', title: '', content: '', enabled: true })
      setFeedback({ type: 'success', message: '知识条目已保存。' })
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const previewRules = async () => {
    try {
      const payload = await client.previewRules(ruleInput, config)
      setRuleOutput(payload.text || '')
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const loadRecords = async () => {
    if (!run?.id) {
      setRecords([])
      return
    }
    try {
      const payload = await client.records(run.id)
      setRecords(payload.records || [])
    } catch (error) {
      setFeedback({ type: 'error', message: errorText(error) })
    }
  }

  const taskColumns = [
    { title: '书城', dataIndex: 'source', width: 100 },
    { title: '平台', dataIndex: 'platformId', width: 80 },
    { title: 'Book ID', dataIndex: 'bookId', width: 120 },
    { title: '书名', dataIndex: 'title', width: 180, render: (value) => value || '-' },
    { title: '原文/处理字数', key: 'chars', width: 130, render: (_, item) => `${item.originalChars || 0}/${item.processedChars || 0}` },
    { title: '阶段', dataIndex: 'currentStage', width: 120, render: (value) => value || '-' },
    { title: '状态', dataIndex: 'status', width: 130, render: statusTag },
    { title: '错误', dataIndex: 'error', width: 220, render: (value) => value || '-' },
    {
      title: '操作', key: 'actions', width: 220, fixed: 'right',
      render: (_, item) => (
        <Space wrap>
          <Button size="small" disabled={!String(item.status).includes('failed')} onClick={() => retryBook(item)}>重试失败阶段</Button>
          <Button size="small" onClick={() => requestSubmit(item)}>提交网络</Button>
        </Space>
      ),
    },
  ]

  const processView = (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={10}>
        <Card title="多平台批量 ID">
          <Form layout="vertical">
            <Form.Item label="批次名称"><Input value={batchName} onChange={(e) => setBatchName(e.target.value)} /></Form.Item>
            <Row gutter={8}>
              <Col span={14}><Form.Item label="书城"><Input value={source} onChange={(e) => setSource(e.target.value)} placeholder="例如：阳光" /></Form.Item></Col>
              <Col span={10}><Form.Item label="platformId"><Input value={platformId} onChange={(e) => setPlatformId(e.target.value)} placeholder="例如：4" /></Form.Item></Col>
            </Row>
            <Form.Item label="Book ID"><TextArea value={bookIds} onChange={(e) => setBookIds(e.target.value)} rows={7} placeholder="每行一个，也支持逗号和空格" /></Form.Item>
            <Button type="primary" onClick={addGroup}>添加书城</Button>
          </Form>
          <Space wrap style={{ marginTop: 16 }}>
            {groups.map((group) => <Tag key={`${group.platformId}-${group.source}`}>{group.source} {group.books.length}本 · P{group.platformId}</Tag>)}
          </Space>
        </Card>
      </Col>
      <Col xs={24} lg={14}>
        <Card title="执行">
          <Descriptions size="small" bordered column={2}>
            <Descriptions.Item label="书城">{groups.length}</Descriptions.Item>
            <Descriptions.Item label="小说">{totalBooks}</Descriptions.Item>
            <Descriptions.Item label="文本模型">{config.textModelId || '待配置'}</Descriptions.Item>
            <Descriptions.Item label="目标版本">{(config.targetVersions || []).join(' / ')}</Descriptions.Item>
          </Descriptions>
          <Space wrap style={{ marginTop: 16 }}>
            <Button type="primary" loading={working} onClick={() => start(false)}>立即执行</Button>
            <Input type="datetime-local" value={scheduleAt} onChange={(e) => setScheduleAt(e.target.value)} style={{ width: 230 }} />
            <Button loading={working} onClick={() => start(true)}>开始定时</Button>
          </Space>
          <Alert style={{ marginTop: 16 }} type="info" showIcon message="立即与定时都只进入统一 Runtime；本页面不会在浏览器里直接跑后台任务。" />
        </Card>
      </Col>
    </Row>
  )

  const tasksView = (
    <Card
      title="任务中心"
      extra={<Space><Button onClick={refreshTasks} disabled={!run?.id}>刷新</Button><Button type="primary" onClick={handoff} disabled={!run?.id}>转入批量工厂</Button></Space>}
    >
      <Space wrap style={{ marginBottom: 12 }}>
        <span>Run：{run?.id || '-'}</span>
        {run && statusTag(run.status)}
        <Input value={publishingAccountId} onChange={(e) => setPublishingAccountId(e.target.value)} placeholder="发布账号 ID" style={{ width: 150 }} />
        <Select value={submitVersion} onChange={setSubmitVersion} style={{ width: 130 }} options={(config.targetVersions || ['original']).map((value) => ({ value, label: value }))} />
      </Space>
      <Table rowKey="key" columns={taskColumns} dataSource={books} scroll={{ x: 1300 }} pagination={false} />
    </Card>
  )

  const configView = (
    <Card title="版本配置">
      <Form layout="vertical">
        <Row gutter={16}>
          <Col xs={24} md={8}><Form.Item label="默认文本模型 ID"><Input value={config.textModelId} onChange={(e) => setConfig((c) => ({ ...c, textModelId: e.target.value }))} /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item label="处理最大字数"><InputNumber min={100} max={100000} value={config.maxText} onChange={(value) => setConfig((c) => ({ ...c, maxText: value || 4000 }))} style={{ width: '100%' }} /></Form.Item></Col>
          <Col xs={24} md={8}>
            <Form.Item label="目标版本">
              <Checkbox.Group
                value={config.targetVersions}
                onChange={(value) => setConfig((c) => ({ ...c, targetVersions: value }))}
                options={['original', 'ai1', 'ai2', 'ai3', 'ai4', 'ai5']}
              />
            </Form.Item>
          </Col>
        </Row>
        <Row gutter={16}>
          {['ai1', 'ai2', 'ai3'].map((version) => (
            <Col xs={24} md={8} key={version}>
              <Form.Item label={`${version} 改文配置档`}>
                <Input
                  value={config.rewriteProfiles?.[version] || ''}
                  onChange={(e) => setConfig((c) => ({ ...c, rewriteProfiles: { ...(c.rewriteProfiles || {}), [version]: e.target.value } }))}
                />
              </Form.Item>
            </Col>
          ))}
        </Row>
        <Button type="primary" onClick={saveConfig}>保存配置</Button>
      </Form>
    </Card>
  )

  const knowledgeView = (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={10}>
        <Card title="新增知识">
          <Form layout="vertical">
            <Form.Item label="类型"><Input value={knowledgeDraft.kind} onChange={(e) => setKnowledgeDraft((v) => ({ ...v, kind: e.target.value }))} /></Form.Item>
            <Form.Item label="标题"><Input value={knowledgeDraft.title} onChange={(e) => setKnowledgeDraft((v) => ({ ...v, title: e.target.value }))} /></Form.Item>
            <Form.Item label="内容"><TextArea rows={8} value={knowledgeDraft.content} onChange={(e) => setKnowledgeDraft((v) => ({ ...v, content: e.target.value }))} /></Form.Item>
            <Checkbox checked={knowledgeDraft.enabled} onChange={(e) => setKnowledgeDraft((v) => ({ ...v, enabled: e.target.checked }))}>启用</Checkbox>
            <Button type="primary" onClick={saveKnowledge} style={{ marginLeft: 12 }}>保存知识</Button>
          </Form>
        </Card>
      </Col>
      <Col xs={24} lg={14}>
        <Card title="知识库">
          <Table
            rowKey="id"
            dataSource={knowledge}
            pagination={false}
            columns={[
              { title: '类型', dataIndex: 'kind', width: 120 },
              { title: '标题', dataIndex: 'title', width: 180 },
              { title: '内容', dataIndex: 'content', ellipsis: true },
              { title: '状态', dataIndex: 'enabled', width: 90, render: (value) => value ? <Tag color="success">启用</Tag> : <Tag>停用</Tag> },
            ]}
          />
        </Card>
      </Col>
    </Row>
  )

  const rulesView = (
    <Row gutter={[16, 16]}>
      <Col span={12}><Card title="规则预览输入"><TextArea rows={16} value={ruleInput} onChange={(e) => setRuleInput(e.target.value)} /><Button type="primary" onClick={previewRules} style={{ marginTop: 12 }}>预览处理</Button></Card></Col>
      <Col span={12}><Card title="处理结果"><TextArea rows={16} readOnly value={ruleOutput} /></Card></Col>
    </Row>
  )

  const recordsView = (
    <Card title="处理记录" extra={<Button onClick={loadRecords} disabled={!run?.id}>刷新记录</Button>}>
      <Table
        rowKey={(item) => `${item.bookKey}-${item.stage}-${item.attempt}`}
        dataSource={records}
        pagination={false}
        columns={[
          { title: 'Book', dataIndex: 'bookKey', width: 130 },
          { title: '阶段', dataIndex: 'stage', width: 150 },
          { title: '尝试', dataIndex: 'attempt', width: 70 },
          { title: '状态', dataIndex: 'status', width: 120, render: statusTag },
          { title: '错误', dataIndex: 'error' },
        ]}
      />
    </Card>
  )

  const items = [
    { key: 'process', label: '处理', children: processView },
    { key: 'tasks', label: '任务', children: tasksView },
    { key: 'config', label: '配置', children: configView },
    { key: 'knowledge', label: '知识库', children: knowledgeView },
    { key: 'rules', label: '处理规则', children: rulesView },
    { key: 'records', label: '记录', children: recordsView },
  ]

  return (
    <main style={{ padding: 20 }}>
      <Typography.Title level={2}>小说获取</Typography.Title>
      <Typography.Paragraph type="secondary">
        多平台正文获取、多版本改文、知识库与规则、任务重试、批量工厂交接和统一发布意图。
      </Typography.Paragraph>
      {feedback && <Alert showIcon closable type={feedback.type} message={feedback.message} onClose={() => setFeedback(null)} style={{ marginBottom: 16 }} />}
      {batch && <Typography.Text type="secondary">当前批次：{batch.name} · {batch.id}</Typography.Text>}
      <Tabs activeKey={tab} onChange={setTab} items={items} />
    </main>
  )
}
