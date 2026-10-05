import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Drawer, Form, Input, InputNumber, Select, Space, Switch, Tabs, Tag, message } from 'antd'
import * as defaultApi from './api.js'

const drawerWidth = 760
const asObject = (value) => value && typeof value === 'object' && !Array.isArray(value) ? value : {}

export default function UnifiedSettingsPanel({ project, api = defaultApi }) {
  const [current, setCurrent] = useState(null)
  const [open, setOpen] = useState('')
  const [loading, setLoading] = useState(false)
  const [productionForm] = Form.useForm()
  const [publishingForm] = Form.useForm()
  const [profileForm] = Form.useForm()

  const projectId = project?.id
  const reload = async () => {
    if (!projectId) return
    const value = await api.getUnifiedSettings(projectId)
    setCurrent(value)
    productionForm.setFieldsValue(asObject(value?.project?.production))
    publishingForm.setFieldsValue(asObject(value?.project?.publishing))
    profileForm.setFieldsValue({
      name: value?.profile?.name || '默认版本配置档',
      version: value?.profile?.version || 'v1',
      processingRulePromptRef: value?.profile?.settings?.processingRulePromptRef || '',
      knowledgePromptRef: value?.profile?.settings?.knowledgePromptRef || '',
    })
  }

  useEffect(() => { void reload().catch(() => {}) }, [projectId])

  const save = async (kind) => {
    setLoading(true)
    try {
      if (kind === 'production') {
        const saved = await api.saveProductionSettings(projectId, await productionForm.validateFields())
        setCurrent(saved); message.success('生产统一设置已保存')
      } else if (kind === 'publishing') {
        const saved = await api.savePublishingSettings(projectId, await publishingForm.validateFields())
        setCurrent(saved); message.success('发布统一设置已保存')
      } else {
        const values = await profileForm.validateFields()
        const saved = await api.saveVersionProfile(projectId, {
          name: values.name,
          version: values.version,
          settings: {
            ...(current?.profile?.settings || {}),
            processingRulePromptRef: values.processingRulePromptRef || '',
            knowledgePromptRef: values.knowledgePromptRef || '',
          },
        })
        setCurrent(saved); message.success('版本对应配置档已保存')
      }
      setOpen('')
    } finally { setLoading(false) }
  }

  const sync = async (type) => {
    setLoading(true)
    try {
      const saved = type === '121' ? await api.sync121Config(projectId) : await api.syncStyleTypes(projectId)
      setCurrent(saved)
      message.success(type === '121' ? '121 网站配置已同步' : '批量风格类型已同步')
    } finally { setLoading(false) }
  }

  const styleTags = useMemo(() => {
    const snapshot = current?.profile?.settings?.styleTypes || {}
    return [...(snapshot.styles || []), ...(snapshot.genres || []), ...(snapshot.genders || [])]
  }, [current])

  return <>
    <Space wrap>
      <Button onClick={() => setOpen('production')}>生产统一设置</Button>
      <Button onClick={() => setOpen('publishing')}>发布统一设置</Button>
      <Button type="primary" onClick={() => setOpen('profile')}>版本对应配置档</Button>
    </Space>

    <Drawer title="生产统一设置" width={drawerWidth} open={open === 'production'} onClose={() => setOpen('')} destroyOnClose={false}
      extra={<Button type="primary" loading={loading} onClick={() => void save('production')}>保存生产统一设置</Button>}>
      <Alert showIcon type="info" message={`应用于当前 BatchProject：${project?.name || projectId || '-'}`} description="项目显式配置优先；未设置字段继续继承版本配置档，再继承系统默认。保存后仍留在当前工作台。" />
      <Form form={productionForm} layout="vertical" style={{ marginTop: 20 }}>
        <Card size="small" title="生产参数">
          <Form.Item name="productionMode" label="生产方式"><Select allowClear placeholder="继承版本/系统" options={[{ value: 'original', label: '原文直转' }, { value: 'viral', label: '爆款开头' }]} /></Form.Item>
          <Form.Item name="aiCopyEnabled" label="AI 文案处理" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item noStyle shouldUpdate={(a,b) => a.aiCopyEnabled !== b.aiCopyEnabled}>{({ getFieldValue }) => getFieldValue('aiCopyEnabled') ? <Form.Item name="aiCopyCount" label="AI 文案数量"><InputNumber min={1} max={20} precision={0} style={{ width: '100%' }} /></Form.Item> : null}</Form.Item>
          <Alert type="warning" showIcon message="AI 文案唯一控制位置" description="数量只在生产统一设置中配置；不再提供重复的‘默认数量/处理优先方案’入口。" />
        </Card>
      </Form>
    </Drawer>

    <Drawer title="发布统一设置" width={drawerWidth} open={open === 'publishing'} onClose={() => setOpen('')} destroyOnClose={false}
      extra={<Button type="primary" loading={loading} onClick={() => void save('publishing')}>保存发布统一设置</Button>}>
      <Alert showIcon type="info" message="当前项目发布默认值" description="只保存当前 BatchProject 的显式覆盖，不跳转独立页面。" />
      <Form form={publishingForm} layout="vertical" style={{ marginTop: 20 }}>
        <Form.Item name="uploadVideoType" label="上传视频类型"><Select allowClear placeholder="继承版本/系统" options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} /></Form.Item>
        <Form.Item name="materialReuse" label="素材复用" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="versionProfile" label="发布网站配置档"><Input placeholder="例如：女频短剧版" /></Form.Item>
      </Form>
    </Drawer>

    <Drawer title="版本对应配置档" width={drawerWidth} open={open === 'profile'} onClose={() => setOpen('')} destroyOnClose={false}
      extra={<Button type="primary" loading={loading} onClick={() => void save('profile')}>保存配置档</Button>}>
      <Tabs items={[
        { key: 'profile', label: '配置档', children: <Form form={profileForm} layout="vertical">
          <Form.Item name="name" label="配置档名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="version" label="版本" rules={[{ required: true }]}><Input /></Form.Item>
          <Card size="small" title="Prompt 引用">
            <Form.Item name="processingRulePromptRef" label="处理规则提示词" extra="规则类 Prompt 的后端引用；不是正文副本。"><Input placeholder="例如 processing-rule-v3" /></Form.Item>
            <Form.Item name="knowledgePromptRef" label="知识库提示词" extra="知识库 Prompt 的后端引用，与处理规则提示词分开。"><Input placeholder="例如 knowledge-v7" /></Form.Item>
          </Card>
        </Form> },
        { key: 'sync', label: '同步', children: <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Alert type="info" showIcon message="同步只更新当前版本配置档快照" description="同步来源是服务端当前项目数据；前端不会自行推断 121 或风格配置。" />
          <Card size="small" title="121 网站配置"><Button loading={loading} onClick={() => void sync('121')}>同步 121 网站配置</Button></Card>
          <Card size="small" title="批量风格 / 类型"><Space direction="vertical"><Button loading={loading} onClick={() => void sync('style')}>同步批量风格类型</Button><Space wrap>{styleTags.map(tag => <Tag key={tag}>{tag}</Tag>)}</Space></Space></Card>
        </Space> },
      ]} />
    </Drawer>
  </>
}
