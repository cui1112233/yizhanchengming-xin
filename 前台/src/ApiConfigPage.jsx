import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, Input, Skeleton, Space, Switch, Tag, Typography, message } from 'antd';
import { getVideoProviderConfig, putVideoProviderConfig } from './providerConfigApi.js';
import './ApiConfigPage.css';

const { Paragraph, Text, Title } = Typography;

const PRESETS = [
  {
    id: 'yd2-mini-video',
    provider: 'personal_api',
    model: 'yd2.0-mini',
    title: 'YD2.0 Mini（图生）',
    description: '平台已维护个人视频 API 适配器；只需填写 API Key。',
    badge: '个人 API',
    remote: true,
  },
  {
    id: 'minimax-h3-video',
    provider: 'autodl_comfyui',
    model: 'minimax-h3-video',
    title: 'MiniMax H3 多图生视频',
    description: 'AutoDL H3；有参考图和无参考图会自动选择对应工作流，最多 9 张参考图。',
    badge: 'AutoDL',
    remote: true,
    details: [
      '有参考图：minimax_h3_lightx2v_v5_15s',
      '无参考图：minimax_h3_lightx2v_no_pic',
    ],
  },
  {
    id: 'seedance-2-0-official',
    provider: 'yfai_seedance',
    model: 'seedance-2-0-official',
    title: 'Seedance 2.0 官方',
    description: 'YFAI 官方直连；支持文生视频和参考图，时长 4–15 秒。',
    badge: 'YFAI',
    remote: true,
  },
  {
    id: 'local-doubao-executor-video',
    provider: 'doubao_local_executor',
    model: 'doubao-seedance',
    title: '本地豆包执行器',
    description: '旧 v88 支持本地执行器配对；新 Go 主线正在迁移该执行链，正式结果仍会统一进入 TOS。',
    badge: '本地执行器',
    remote: false,
    migrating: true,
  },
];

function normalizeRecord(preset, record) {
  return {
    provider: preset.provider,
    model: record?.model || preset.model,
    enabled: record?.enabled === true,
    configured: record?.configured === true,
    createUrl: record?.create_url || '',
    tasksUrl: record?.tasks_url || '',
    resultUrl: record?.result_url || '',
  };
}

function ProviderCard({ preset, record, apiKey, saving, onKeyChange, onToggle, onSave }) {
  const configured = record?.configured === true;
  const enabled = record?.enabled === true;
  return (
    <div className={`api-platform-card${preset.migrating ? ' is-migrating' : ''}`}>
      <div className="api-platform-info">
        <span className="api-platform-icon">▶</span>
        <div>
          <div className="api-platform-name-row">
            <strong>{preset.title}</strong>
            <Tag color={enabled ? 'green' : configured ? 'blue' : 'default'}>{enabled ? '已启用' : configured ? '已配置' : '未配置'}</Tag>
            <Tag>{preset.badge}</Tag>
          </div>
          <small>{preset.description}</small>
          {preset.details?.length ? <div className="api-platform-details">{preset.details.map(item => <code key={item}>{item}</code>)}</div> : null}
        </div>
      </div>

      <div className="api-platform-credential">
        {preset.migrating ? (
          <Alert type="warning" showIcon message="本地执行器新链路尚未接入" description="这里不会假装已经可以执行。完成本地执行器 → TOS 迁移后再开放启用。" />
        ) : (
          <>
            <Input.Password
              value={apiKey}
              onChange={event => onKeyChange(event.target.value)}
              placeholder={configured ? '留空表示不修改已保存的 Key' : '填写 API Key'}
              autoComplete="new-password"
            />
            <div className="api-platform-security">
              <span>🔒 API Key 仅提交给 Go 服务端，并以 AES-256-GCM 加密保存。</span>
              <Button size="small" onClick={onSave} loading={saving}>{configured ? '保存修改' : '保存配置'}</Button>
            </div>
          </>
        )}
      </div>

      <div className="api-platform-state">
        <Switch
          checked={enabled}
          disabled={preset.migrating || saving || (!configured && !String(apiKey || '').trim())}
          onChange={onToggle}
        />
        <small>{preset.migrating ? '等待迁移' : enabled ? '业务可用' : '业务不可用'}</small>
      </div>
    </div>
  );
}

export default function ApiConfigPage() {
  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState({});
  const [keys, setKeys] = useState({});
  const [saving, setSaving] = useState({});
  const [loadError, setLoadError] = useState('');
  const [messageApi, contextHolder] = message.useMessage();
  const videoSectionRef = useRef(null);

  const configuredCount = useMemo(() => PRESETS.filter(preset => records[preset.provider]?.configured).length, [records]);

  async function loadConfigs() {
    setLoading(true);
    setLoadError('');
    try {
      const pairs = await Promise.all(PRESETS.filter(preset => !preset.migrating).map(async preset => {
        const value = await getVideoProviderConfig(preset.provider);
        return [preset.provider, normalizeRecord(preset, value)];
      }));
      setRecords(current => ({ ...current, ...Object.fromEntries(pairs) }));
    } catch (error) {
      setLoadError(error?.message || '视频模型配置加载失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { loadConfigs(); }, []);

  useEffect(() => {
    const section = new URLSearchParams(window.location.search).get('section');
    if (section === 'video-models') window.setTimeout(() => videoSectionRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 80);
  }, [loading]);

  async function savePreset(preset, enabled = records[preset.provider]?.enabled === true) {
    if (preset.migrating) return;
    const apiKey = String(keys[preset.provider] || '').trim();
    const existing = records[preset.provider];
    if (!existing?.configured && !apiKey) {
      messageApi.warning('首次配置请先填写 API Key');
      return;
    }
    setSaving(current => ({ ...current, [preset.provider]: true }));
    try {
      const result = await putVideoProviderConfig(preset.provider, {
        model: preset.model,
        apiKey,
        enabled,
      });
      setRecords(current => ({ ...current, [preset.provider]: normalizeRecord(preset, result) }));
      setKeys(current => ({ ...current, [preset.provider]: '' }));
      messageApi.success(enabled ? `${preset.title} 已保存并启用` : `${preset.title} 配置已保存`);
    } catch (error) {
      messageApi.error(error?.message || '保存视频模型配置失败');
    } finally {
      setSaving(current => ({ ...current, [preset.provider]: false }));
    }
  }

  async function togglePreset(preset, enabled) {
    await savePreset(preset, enabled);
  }

  if (loading) return <main className="api-config-page"><Skeleton active paragraph={{ rows: 10 }} /></main>;

  return (
    <main className="api-config-page">
      {contextHolder}
      <header className="api-config-header">
        <div>
          <span className="api-config-eyebrow">MODEL CENTER</span>
          <Title level={1}>API 配置</Title>
          <Paragraph>模型只在这里配置一次；Agent、批量工厂、创作漫剧等业务共用同一份服务端配置。</Paragraph>
        </div>
        <Space wrap>
          <Button onClick={() => window.location.assign('/agent')}>返回 Agent</Button>
          <Button onClick={loadConfigs}>刷新配置</Button>
        </Space>
      </header>

      {loadError ? <Alert className="api-config-alert" type="error" showIcon message="Provider 配置服务暂不可用" description={`${loadError}。如果服务端尚未配置 PROVIDER_CREDENTIAL_KEY，请先配置 32 字节加密密钥。`} /> : null}

      <section className="api-config-summary">
        <div><strong>{configuredCount}</strong><span>已配置视频服务</span></div>
        <div><strong>TOS</strong><span>图片 / 视频统一正式存储</span></div>
        <div><strong>AES-256-GCM</strong><span>Provider API Key 服务端加密</span></div>
      </section>

      <section className="api-config-panel" ref={videoSectionRef} id="video-models">
        <div className="api-config-panel-head">
          <div><span>PLATFORM PRESETS</span><Title level={3}>平台预设视频模型</Title></div>
          <Text type="secondary">完成配置并启用后，Agent 与其他生产链才允许调用对应 Provider。</Text>
        </div>
        <div className="api-platform-list">
          {PRESETS.map(preset => <ProviderCard
            key={preset.id}
            preset={preset}
            record={records[preset.provider] || normalizeRecord(preset, null)}
            apiKey={keys[preset.provider] || ''}
            saving={saving[preset.provider] === true}
            onKeyChange={value => setKeys(current => ({ ...current, [preset.provider]: value }))}
            onSave={() => savePreset(preset)}
            onToggle={enabled => togglePreset(preset, enabled)}
          />)}
        </div>
      </section>

      <section className="api-config-panel api-config-secondary">
        <div className="api-config-panel-head"><div><span>TEXT MODELS</span><Title level={3}>文本模型</Title></div><Tag>按 v88 链路迁移中</Tag></div>
        <Paragraph type="secondary">当前 Agent 智能问答仍通过服务端 AI_* 配置运行。下一步会把旧 v88 的文本模型目录、测试连接、启用状态迁入同一 MySQL 模型中心。</Paragraph>
      </section>

      <section className="api-config-panel api-config-secondary">
        <div className="api-config-panel-head"><div><span>IMAGE MODELS</span><Title level={3}>图片模型</Title></div><Tag>按 v88 链路迁移中</Tag></div>
        <Paragraph type="secondary">图片生成和修改 Provider 将与视频一样使用统一 Provider 配置、TOS 资产和 Agent Tool Registry；未接入前不会显示为可用。</Paragraph>
      </section>
    </main>
  );
}
