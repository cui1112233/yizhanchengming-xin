import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import {
  Alert,
  App,
  Button,
  Card,
  ConfigProvider,
  Descriptions,
  Drawer,
  Empty,
  Form,
  Input,
  Layout,
  Menu,
  Modal,
  Result,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import 'antd/dist/reset.css'
import './admin.css'
import { requestJSON } from './api.js'
import { applyThemeVariables, themeConfig } from './theme.js'

const { Header, Sider, Content } = Layout
const { TextArea } = Input

function errorMessage(error, fallback) {
  const message = error.message || fallback
  return error.requestId ? `${message}（请求编号：${error.requestId}）` : message
}

function CapabilityBoundary({ status, children }) {
  if (status === 'loading') return <div className="admin-state"><Spin size="large" tip="正在读取后台权限…" /></div>
  if (status === 'unauthorized') return <Result status="warning" title="登录已过期" subTitle="请重新登录后访问管理端。" />
  if (status === 'forbidden') return <Result status="403" title="没有后台访问权限" subTitle="后台入口仅对拥有 admin.* capability 的账号开放。" />
  if (status === 'error') return <Result status="error" title="后台权限加载失败" subTitle="请稍后重试。" />
  return children
}

function PromptPage({ capabilities }) {
  const { message } = App.useApp()
  const canView = capabilities.includes('admin.prompt.view')
  const canEdit = capabilities.includes('admin.prompt.edit')
  const canPublish = capabilities.includes('admin.prompt.publish')
  const [prompts, setPrompts] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [detail, setDetail] = useState(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form] = Form.useForm()

  const loadPrompts = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const payload = await requestJSON('/api/v1/admin/prompts')
      setPrompts(Array.isArray(payload?.prompts) ? payload.prompts : [])
    } catch (err) {
      setError(err)
      setPrompts([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void loadPrompts() }, [loadPrompts])

  const openDetail = async (record) => {
    if (!canView) return
    setDrawerOpen(true)
    setDetailLoading(true)
    setDetail(null)
    try {
      const payload = await requestJSON(`/api/v1/admin/prompts/${encodeURIComponent(record.key)}/versions/${record.version}`)
      setDetail(payload?.prompt || null)
    } catch (err) {
      setDetail({ error: err })
    } finally {
      setDetailLoading(false)
    }
  }

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setEditorOpen(true)
  }

  const openEdit = async () => {
    if (!detail || detail.error) return
    form.setFieldsValue({ content: detail.content || '' })
    setEditing(detail)
    setEditorOpen(true)
  }

  const saveDraft = async (values) => {
    try {
      if (editing) {
        await requestJSON(`/api/v1/admin/prompts/${encodeURIComponent(editing.key)}/drafts/${editing.version}`, { method: 'PUT', body: JSON.stringify({ content: values.content }) })
        message.success('草稿已保存')
      } else {
        await requestJSON(`/api/v1/admin/prompts/${encodeURIComponent(values.key)}/drafts`, { method: 'POST', body: JSON.stringify({ content: values.content }) })
        message.success('草稿已创建')
      }
      setEditorOpen(false)
      await loadPrompts()
    } catch (err) {
      message.error(errorMessage(err, '草稿保存失败'))
    }
  }

  const publish = (record) => {
    Modal.confirm({
      title: '发布提示词版本？',
      content: `${record.key} v${record.version} 将成为运行时版本。`,
      okText: '发布',
      cancelText: '取消',
      onOk: async () => {
        try {
          await requestJSON(`/api/v1/admin/prompts/${encodeURIComponent(record.key)}/versions/${record.version}/publish`, { method: 'POST' })
          message.success('提示词已发布')
          await loadPrompts()
        } catch (err) {
          message.error(errorMessage(err, '发布失败'))
        }
      },
    })
  }

  const restore = (record) => {
    Modal.confirm({
      title: '恢复历史版本？',
      content: `${record.key} v${record.version} 将生成一个新的运行时版本，历史不会被覆盖。`,
      okText: '恢复',
      cancelText: '取消',
      onOk: async () => {
        try {
          await requestJSON(`/api/v1/admin/prompts/${encodeURIComponent(record.key)}/versions/${record.version}/restore`, { method: 'POST' })
          message.success('历史版本已恢复为新版本')
          await loadPrompts()
        } catch (err) {
          message.error(errorMessage(err, '恢复失败'))
        }
      },
    })
  }

  const columns = useMemo(() => [
    { title: 'Prompt Key', dataIndex: 'key', key: 'key', render: (value) => <Typography.Text code>{value}</Typography.Text> },
    { title: '版本', dataIndex: 'version', key: 'version', width: 90, render: (value) => `v${value}` },
    { title: '状态', dataIndex: 'lifecycle', key: 'lifecycle', width: 110, render: (value) => <Tag color={value === 'published' ? 'green' : value === 'draft' ? 'gold' : 'default'}>{value}</Tag> },
    { title: '来源', dataIndex: 'seedSource', key: 'seedSource', width: 110 },
    { title: '启用', dataIndex: 'enabled', key: 'enabled', width: 80, render: (value) => <Tag color={value ? 'green' : 'default'}>{value ? '是' : '否'}</Tag> },
    { title: '内容 SHA-256', dataIndex: 'contentSha256', key: 'contentSha256', render: (value) => <Typography.Text ellipsis={{ tooltip: value }} className="admin-hash">{value || '-'}</Typography.Text> },
    {
      title: '操作', key: 'actions', width: 280, render: (_, record) => (
        <Space wrap>
          {canView && <Button size="small" onClick={() => void openDetail(record)}>查看正文</Button>}
          {canPublish && record.lifecycle !== 'published' && <Button size="small" type="primary" onClick={() => publish(record)}>发布</Button>}
          {canPublish && <Button size="small" onClick={() => restore(record)}>恢复</Button>}
        </Space>
      ),
    },
  ], [canPublish, canView])

  if (!canView) return <Result status="403" title="没有提示词查看权限" subTitle="当前账号没有 admin.prompt.view。" />

  return (
    <div className="admin-page">
      <div className="admin-page-heading">
        <div>
          <Typography.Title level={3}>提示词库</Typography.Title>
          <Typography.Paragraph type="secondary">版本、草稿和发布状态来自 Go / MySQL；列表不会返回正文。</Typography.Paragraph>
        </div>
        <Space>
          <Button onClick={() => void loadPrompts()} loading={loading}>刷新</Button>
          {canEdit && <Button type="primary" onClick={openCreate}>新建草稿</Button>}
        </Space>
      </div>
      {error && <Alert type="error" showIcon message="提示词加载失败" description={errorMessage(error, '请检查登录状态或稍后重试。')} action={<Button size="small" onClick={() => void loadPrompts()}>重新加载</Button>} />}
      <Card className="admin-card" bordered={false}>
        <Table rowKey={(record) => `${record.key}-${record.version}`} loading={loading} columns={columns} dataSource={prompts} locale={{ emptyText: <Empty description="暂无提示词版本" /> }} pagination={{ pageSize: 10 }} />
      </Card>

      <Drawer title={detail ? `${detail.key} · v${detail.version}` : '提示词详情'} width={620} open={drawerOpen} onClose={() => setDrawerOpen(false)} extra={detail && !detail.error && canEdit && detail.lifecycle === 'draft' ? <Button onClick={openEdit}>编辑草稿</Button> : null}>
        {detailLoading && <Spin />}
        {!detailLoading && detail?.error && <Alert type="error" message="正文加载失败" description={errorMessage(detail.error, '请关闭后重试。')} />}
        {!detailLoading && detail && !detail.error && <>
          <Descriptions bordered size="small" column={1} items={[
            { key: 'key', label: 'Prompt Key', children: detail.key },
            { key: 'version', label: '版本', children: `v${detail.version}` },
            { key: 'lifecycle', label: '状态', children: detail.lifecycle },
            { key: 'sha', label: '内容 SHA-256', children: detail.contentSha256 || '-' },
          ]} />
          <Typography.Title level={5} style={{ marginTop: 24 }}>正文</Typography.Title>
          <TextArea value={detail.content || ''} readOnly autoSize={{ minRows: 12, maxRows: 24 }} />
        </>}
      </Drawer>

      <Modal title={editing ? `编辑草稿 · ${editing.key} v${editing.version}` : '新建提示词草稿'} open={editorOpen} onCancel={() => setEditorOpen(false)} footer={null} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={saveDraft} preserve={false}>
          {!editing && <Form.Item name="key" label="Prompt Key" rules={[{ required: true, message: '请输入 Prompt Key' }]}><Input placeholder="例如 script.default" /></Form.Item>}
          <Form.Item name="content" label="提示词正文" rules={[{ required: true, message: '请输入提示词正文' }]}><TextArea autoSize={{ minRows: 12, maxRows: 24 }} /></Form.Item>
          <Space>
            <Button onClick={() => setEditorOpen(false)}>取消</Button>
            <Button type="primary" htmlType="submit">保存草稿</Button>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}

function DashboardPage() {
  return <Card className="admin-card" bordered={false}><Typography.Title level={3}>后台首页</Typography.Title><Typography.Paragraph type="secondary">请选择左侧模块开始管理。模型、技能、成员和错误日志将在后续批次接入。</Typography.Paragraph></Card>
}

export function AdminShell() {
  const [status, setStatus] = useState('loading')
  const [capabilities, setCapabilities] = useState([])
  const [selected, setSelected] = useState('prompts')

  useEffect(() => {
    let active = true
    requestJSON('/api/v1/admin/capabilities').then((payload) => {
      if (!active) return
      const values = Array.isArray(payload?.capabilities) ? payload.capabilities : []
      setCapabilities(values)
      setSelected(values.includes('admin.prompt.view') ? 'prompts' : 'dashboard')
      setStatus('ready')
    }).catch((err) => {
      if (!active) return
      setStatus(err.status === 401 ? 'unauthorized' : err.status === 403 ? 'forbidden' : 'error')
    })
    return () => { active = false }
  }, [])

  const menuItems = useMemo(() => {
    const items = []
    if (capabilities.includes('admin.dashboard.view')) items.push({ key: 'dashboard', label: '后台首页' })
    if (capabilities.includes('admin.prompt.view')) items.push({ key: 'prompts', label: '提示词库' })
    return items
  }, [capabilities])

  return (
    <ConfigProvider theme={themeConfig('light')}>
      <App>
        <CapabilityBoundary status={status}>
          {menuItems.length === 0 ? <Result status="403" title="没有可访问的后台模块" subTitle="请联系管理员授予具体 admin.* capability。" /> : <Layout className="admin-layout">
            <Sider breakpoint="lg" collapsedWidth="0" className="admin-sider">
              <div className="admin-brand"><span className="admin-brand-mark">YC</span><span>一战晟铭 · 管理端</span></div>
              <Menu theme="dark" mode="inline" selectedKeys={[selected]} items={menuItems} onClick={({ key }) => setSelected(key)} />
            </Sider>
            <Layout>
              <Header className="admin-header"><Typography.Title level={4}>治理控制台</Typography.Title><Typography.Text type="secondary">Go Admin · MySQL facts</Typography.Text></Header>
              <Content className="admin-content">{selected === 'prompts' && capabilities.includes('admin.prompt.view') ? <PromptPage capabilities={capabilities} /> : <DashboardPage />}</Content>
            </Layout>
          </Layout>}
        </CapabilityBoundary>
      </App>
    </ConfigProvider>
  )
}

const root = document.getElementById('root')
if (root) {
  root.classList.add('admin-theme-root')
  applyThemeVariables(root.ownerDocument.documentElement, 'light')
  createRoot(root).render(<React.StrictMode><AdminShell /></React.StrictMode>)
}
