import React, { useCallback, useEffect, useState } from 'react'
import { Alert, Button, Card, Radio, Space, Switch, Table, Typography } from 'antd'
import { getWorkspaceSettings, saveWorkspaceSettings } from './api.js'
import PageState from './ui/PageState.jsx'

export default function SettingsPage({ onTheme }) {
  const [settings, setSettings] = useState(null)
  const [devices, setDevices] = useState([])
  const [executorStatus, setExecutorStatus] = useState({ available: true })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const result = await getWorkspaceSettings()
      setSettings(result.settings)
      setDevices(result.executors || [])
      setExecutorStatus(result.executorStatus || { available: true })
    } catch (readError) {
      setError(readError.message || '读取设置失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const save = async () => {
    setSaving(true)
    setError('')
    try {
      const result = await saveWorkspaceSettings(settings)
      setSettings(result.settings)
      setDevices(result.executors || [])
      setExecutorStatus(result.executorStatus || { available: true })
      onTheme?.(result.settings.theme)
    } catch (saveError) {
      setError(saveError.message || '保存设置失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading && !settings) return <PageState state="loading" title="正在读取设置…" className="tts-history-state" />
  if (error && !settings) return <PageState state="failed" title="设置读取失败" description={error} onRetry={() => void load()} className="tts-history-state" />

  return <main className="tts-history-page">
    <section className="tts-history-heading">
      <div>
        <Typography.Text type="secondary">一战晟铭 · 服务端偏好</Typography.Text>
        <Typography.Title level={2}>设置</Typography.Title>
        <Typography.Paragraph type="secondary">已选择配置由 MySQL 持久化；设备在线状态直接读取已注册本地执行器，不会伪造可用状态。</Typography.Paragraph>
      </div>
      <Button type="primary" onClick={() => void save()} loading={saving}>保存</Button>
    </section>
    {error && <Alert type="error" showIcon message={error} />}
    <Card title="界面与提醒">
      <Space direction="vertical">
        <Radio.Group value={settings.theme} onChange={event => setSettings({ ...settings, theme: event.target.value })}>
          <Radio.Button value="dark">深色</Radio.Button><Radio.Button value="light">浅色</Radio.Button>
        </Radio.Group>
        <Switch checked={settings.notificationsEnabled} onChange={value => setSettings({ ...settings, notificationsEnabled: value })} checkedChildren="提醒开启" unCheckedChildren="提醒关闭" />
      </Space>
    </Card>
    <Card title="存储偏好" style={{ marginTop: 16 }}>
      <Radio.Group value={settings.storagePreference} onChange={event => setSettings({ ...settings, storagePreference: event.target.value })}>
        <Radio value="tos">TOS 持久化媒体</Radio><Radio value="local_executor">本地执行器优先</Radio>
      </Radio.Group>
    </Card>
    <Card title="实际执行器状态" style={{ marginTop: 16 }}>
      {!executorStatus.available && <Alert type="warning" showIcon message={executorStatus.reason || '执行器状态暂不可用'} />}
      <Table rowKey="id" dataSource={devices} pagination={false} locale={{ emptyText: executorStatus.available ? '没有已注册执行器；当前不能使用本地执行器。' : executorStatus.reason || '执行器状态暂不可用' }} columns={[
        { title: '设备', dataIndex: 'name' },
        { title: '提供方 / 模型', render: (_, device) => `${device.providerKey} / ${device.model}` },
        { title: '状态', render: (_, device) => device.online ? '在线' : '离线' },
        { title: '最近心跳', dataIndex: 'lastSeenAt', render: value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' },
      ]} />
    </Card>
  </main>
}
