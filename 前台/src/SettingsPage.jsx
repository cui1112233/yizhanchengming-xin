import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Card, Radio, Select, Slider, Space, Switch, Table, Typography } from 'antd'
import { getWorkspaceSettings } from './api.js'
import PageState from './ui/PageState.jsx'
import './settings-page.css'

const validTheme = value => value === 'dark' || value === 'light'

function safeError(copy, error) {
  const requestId = typeof error?.requestId === 'string' && /^[a-zA-Z0-9._:-]{1,128}$/.test(error.requestId) ? error.requestId : ''
  return requestId ? `${copy} 请求 ID：${requestId}` : copy
}

function readiness(status) {
  const labels = { available: '执行器群组已就绪', degraded: '执行器群组就绪状态降级', unavailable: '执行器群组不可用' }
  const reasons = { available: '已检测到在线执行器', offline: '已注册执行器均离线', not_configured: '未配置已注册执行器', status_unavailable: '暂时无法确定执行器就绪状态' }
  const contradictory = status?.status === 'available'
    ? Boolean(status.reasonCode && status.reasonCode !== 'available')
    : status?.reasonCode === 'available'
  const known = Object.hasOwn(labels, status?.status) && !contradictory
  const reason = known && Object.hasOwn(reasons, status?.reasonCode) ? status.reasonCode : 'status_unavailable'
  return { label: known ? labels[status.status] : '执行器群组就绪状态未知', reason: reasons[reason], ready: known && status.status === 'available' }
}

export default function SettingsPage({ onTheme, theme }) {
  const [settings, setSettings] = useState(null)
  const [devices, setDevices] = useState([])
  const [executorStatus, setExecutorStatus] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [changingTheme, setChangingTheme] = useState(false)
  const themeIntent = useRef(0)
  const loadIntent = useRef(0)
  const globalTheme = useRef(theme)
  globalTheme.current = theme

  const reconcileTheme = useCallback(value => validTheme(globalTheme.current) ? globalTheme.current : value, [])
  useEffect(() => {
    if (validTheme(theme)) setSettings(current => current ? { ...current, theme } : current)
  }, [theme])

  const load = useCallback(async () => {
    const intent = ++loadIntent.current
    setLoading(true)
    setError('')
    try {
      const result = await getWorkspaceSettings()
      if (intent !== loadIntent.current) return
      setSettings({ ...result.settings, theme: reconcileTheme(result.settings.theme) })
      setDevices(result.executors || [])
      setExecutorStatus(result.executorStatus || null)
    } catch (readError) {
      if (intent === loadIntent.current) setError(safeError('读取设置失败，请重试。', readError))
    } finally {
      if (intent === loadIntent.current) setLoading(false)
    }
  }, [reconcileTheme])

  useEffect(() => {
    void load()
    return () => { loadIntent.current += 1 }
  }, [load])

  const save = async () => {
    setSaving(true)
    setError('')
    try {
      const result = await onTheme(undefined, { saveSettings: settings })
      setSettings({ ...result.settings, theme: reconcileTheme(result.settings.theme) })
      setDevices(result.executors || [])
      setExecutorStatus(result.executorStatus || null)
    } catch (saveError) {
      if (validTheme(saveError?.rollbackTheme)) setSettings(current => ({ ...current, theme: reconcileTheme(saveError.rollbackTheme) }))
      setError(safeError('保存设置失败，请重试。', saveError))
    } finally {
      setSaving(false)
    }
  }

  const changeTheme = async nextTheme => {
    const intent = ++themeIntent.current
    const previousTheme = settings.theme
    setSettings(current => ({ ...current, theme: nextTheme }))
    setChangingTheme(true)
    setError('')
    try {
      const result = await onTheme(nextTheme)
      if (intent === themeIntent.current) {
        setSettings(current => ({ ...current, theme: reconcileTheme(result.settings.theme) }))
        setDevices(result.executors || [])
        setExecutorStatus(result.executorStatus || null)
      }
    } catch (failure) {
      if (intent === themeIntent.current) {
        const rollback = ['dark', 'light'].includes(failure?.rollbackTheme) ? failure.rollbackTheme : previousTheme
        setSettings(current => ({ ...current, theme: reconcileTheme(rollback) }))
        setError(safeError('保存主题失败，请重试。', failure))
      }
    } finally {
      if (intent === themeIntent.current) setChangingTheme(false)
    }
  }

  if (loading && !settings) return <PageState state="loading" title="正在读取设置…" className="tts-history-state" />
  if (error && !settings) return <PageState state="failed" title="设置读取失败" description={error} onRetry={() => void load()} className="tts-history-state" />

  const runtime = readiness(executorStatus)
  return <main className="task6-settings">
    <section className="tts-history-heading">
      <div>
        <Typography.Text type="secondary">一战晟铭 · 服务端偏好</Typography.Text>
        <Typography.Title level={2}>设置</Typography.Title>
        <Typography.Paragraph type="secondary">已选择配置由 MySQL 持久化；设备在线状态直接读取已注册本地执行器，不会伪造可用状态。</Typography.Paragraph>
      </div>
      <Button type="primary" onClick={() => void save()} loading={saving} disabled={changingTheme}>保存</Button>
    </section>
    {error && <Alert type="error" showIcon message={error} />}
    <Card className="settings-card" title="工作台与提醒">
      <Space direction="vertical">
        <Radio.Group name="workspace-theme" value={reconcileTheme(settings.theme)} disabled={saving} onChange={event => void changeTheme(event.target.value)}>
          <Radio.Button value="dark">深色</Radio.Button><Radio.Button value="light">浅色</Radio.Button>
        </Radio.Group>
        <Switch checked={settings.notificationsEnabled} onChange={value => setSettings({ ...settings, notificationsEnabled: value })} checkedChildren="提醒开启" unCheckedChildren="提醒关闭" />
        <label>桌面宠物 <Select aria-label="桌面宠物" value={settings.petId || 'default'} onChange={petId => setSettings({ ...settings, petId })} options={[{ value: 'default', label: '默认宠物' }, { value: 'fox', label: '小狐' }]} /></label>
        <label>提示音音量 <Slider value={settings.soundVolume ?? 60} onChange={soundVolume => setSettings({ ...settings, soundVolume })} min={0} max={100} /></label>
        <Switch checked={settings.petVisible !== false} onChange={petVisible => setSettings({ ...settings, petVisible })} checkedChildren="显示宠物" unCheckedChildren="隐藏宠物" />
        <Switch checked={Boolean(settings.companionActive)} onChange={companionActive => setSettings({ ...settings, companionActive })} checkedChildren="主动说话" unCheckedChildren="静默" />
      </Space>
    </Card>
    <Card className="settings-card" title="已选择配置">
      <Typography.Paragraph type="secondary">存储偏好表示你的选择，具体任务的执行后端以任务运行结果为准。</Typography.Paragraph>
      <Radio.Group name="workspace-storage" value={settings.storagePreference} onChange={event => setSettings({ ...settings, storagePreference: event.target.value })}>
        <Radio value="tos">TOS 持久化媒体</Radio><Radio value="local_executor">本地执行器优先</Radio>
      </Radio.Group>
    </Card>
    <Card className="settings-card" title="运行时就绪状态">
      <Alert type={runtime.ready ? 'success' : 'warning'} showIcon message={runtime.label} description={runtime.reason} />
      <Typography.Paragraph type="secondary">此状态仅反映执行器群组就绪情况，不代表提供方或模型兼容性，也不证明具体任务的执行后端。</Typography.Paragraph>
      <Table rowKey="id" dataSource={devices} pagination={false} locale={{ emptyText: '没有已注册执行器；当前不能使用本地执行器。' }} columns={[
        { title: '设备', dataIndex: 'name' },
        { title: '提供方 / 模型', render: (_, device) => `${device.providerKey} / ${device.model}` },
        { title: '状态', render: (_, device) => device.online ? '在线' : '离线' },
        { title: '最近心跳', dataIndex: 'lastSeenAt', render: value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' },
      ]} />
    </Card>
  </main>
}
