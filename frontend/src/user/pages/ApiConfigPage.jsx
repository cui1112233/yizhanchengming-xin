import { Alert, Button, Input, Select, Skeleton, Space, Tag, message } from 'antd';
import { Cpu, KeyRound, RefreshCw, ShieldCheck, Video } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { canManageModelCatalog, getConfig } from '../../shared/api/config';
import { createManagedModel, getManagedModelQuotas, invalidateAvailableModelCatalog, listManagedModels, updateManagedModel } from '../../shared/api/modelCatalog';
import { getMemberCenter } from '../../shared/api/member';
import { FANKE_VIDEO_MODELS } from './shuihuo/providerModelBinding';
import { PageHeader, Panel, RoleBadge } from './accountCenterShared';
import { getPresetCardState } from './api-config-preset-state';
import { isCatalogModelAvailable, modelNeedsNoCredential } from './apiConfigCatalogMerge';
import ApiIntegrationSettings from './ApiIntegrationSettings';

const DEFAULT_H3_WORKFLOW_ID = 'minimax_h3_image_audio_to_video_v2_15s';

// 输入类字段停止输入后自动提交；失焦会立即提交。
const AUTOSAVE_DELAY = 800;

const PLATFORM_PRESETS = [
  { id: 'yd2-mini-video', displayName: 'YD2.0 Mini（图生）', description: '平台已维护视频适配器；只需填写 API Key。', credentialMode: 'apiKey' },
  { id: 'minimax-h3-video', displayName: 'MiniMax H3 多图生视频', description: '支持切换 AutoDL 工作流；填写 API Key 和工作流 ID。', credentialMode: 'apiKey' },
  { id: 'seedance-2-0-official', displayName: 'Seedance 2.0 官方', description: 'YFAI 官方直连；支持文生视频和参考图；只需填写 API Key。', credentialMode: 'apiKey' },
  { id: 'fanke-video-api', displayName: '梵客视频 API', description: '支持多个视频模型；填写 API Key 并选择一个模型。', credentialMode: 'apiKey' }
];

const PLATFORM_PRESET_IDS = PLATFORM_PRESETS.map(preset => preset.id);

const SAVE_STATUS_TEXT = { idle: '修改后自动保存', saving: '保存中…', saved: '已保存', error: '保存失败' };

function quotaText(value) {
  return Number.isFinite(Number(value)) ? Number(value).toLocaleString('zh-CN', { maximumFractionDigits: 2 }) : '—';
}

function ModelQuotaStatus({ quota, error }) {
  if (error && !quota) return <div className="ac-model-quota error"><span>额度读取失败</span></div>;
  if (!quota) return <div className="ac-model-quota unknown"><span>额度尚未查询</span></div>;
  const labels = { available: '额度正常', depleted: '额度已耗尽', unauthorized: '密钥无效或无权限', timeout: '额度查询超时', error: '额度查询失败', unknown: '额度未知', not_configured: '未配置 API Key', not_applicable: '无需 API 额度' };
  const label = labels[quota.status] || '额度未知';
  const hasBalance = Number.isFinite(Number(quota.available)) && quota.percent !== null;
  return <div className={`ac-model-quota ${quota.status || 'unknown'}`}>
    <div className="ac-model-quota-head"><span>{label}</span>{hasBalance ? <b>可用额度 {quotaText(quota.available)} / {quotaText(quota.total)} {quota.unit || ''}</b> : null}</div>
    {hasBalance ? <div className="ac-model-quota-track"><i style={{ width: `${Math.max(0, Math.min(100, Number(quota.percent) || 0))}%` }} /></div> : null}
    <small>{quota.checkedAt ? `更新于 ${new Date(quota.checkedAt).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}` : '等待刷新'}</small>
  </div>;
}

export default function ApiConfigPage() {
  const [loading, setLoading] = useState(true);
  const [config, setConfig] = useState(null);
  const [center, setCenter] = useState(null);
  const [models, setModels] = useState([]);
  const [presetKeys, setPresetKeys] = useState({});
  const [presetWorkflowIds, setPresetWorkflowIds] = useState({});
  const [presetModelIds, setPresetModelIds] = useState({});
  const [presetBaseURLs, setPresetBaseURLs] = useState({});
  const [presetSavingId, setPresetSavingId] = useState('');
  const [saveStatus, setSaveStatus] = useState('idle');
  const [quotas, setQuotas] = useState([]);
  const [quotaError, setQuotaError] = useState('');
  const [quotaLoading, setQuotaLoading] = useState(false);
  const [catalogModelSavingId, setCatalogModelSavingId] = useState('');
  const canManageApi = canManageModelCatalog(config);
  const member = center?.member;

  // 自动保存需要的“最新值”引用：定时器里不能依赖闭包里的旧 state。
  const queueRef = useRef(Promise.resolve());
  const timersRef = useRef({});
  const dirtyRef = useRef({});
  const versionsRef = useRef({});
  const modelsRef = useRef([]);
  const keysRef = useRef({});
  const workflowRef = useRef({});
  const modelIdsRef = useRef({});
  const baseURLsRef = useRef({});
  // 记录用户是否真的动过 Key 输入框：只有“输入过又清空”才算清除，没动过则保持已存 Key。
  const touchedKeysRef = useRef(new Set());
  // 正在输入的 Key 字段：保存回包先不清输入框，避免把用户没输完的 Key 冲掉。
  const focusedKeyRef = useRef(new Set());

  useEffect(() => { modelsRef.current = models; }, [models]);
  useEffect(() => { keysRef.current = presetKeys; }, [presetKeys]);
  useEffect(() => { workflowRef.current = presetWorkflowIds; }, [presetWorkflowIds]);
  useEffect(() => { modelIdsRef.current = presetModelIds; }, [presetModelIds]);
  useEffect(() => { baseURLsRef.current = presetBaseURLs; }, [presetBaseURLs]);

  const refreshQuotaState = useCallback(async () => {
    if (!canManageApi) return;
    setQuotaLoading(true);
    try {
      setQuotas(await getManagedModelQuotas());
      setQuotaError('');
    } catch (error) {
      // 读取失败不能退化成“额度尚未查询”，否则用户会把服务异常当成没填 Key。
      setQuotaError(error.message || '额度查询失败');
      message.error(error.message || '额度查询失败');
    }
    finally { setQuotaLoading(false); }
  }, [canManageApi]);

  const refreshModels = useCallback(async () => {
    if (!canManageApi) return;
    const [nextModels, nextQuotas] = await Promise.all([
      listManagedModels(),
      getManagedModelQuotas().then(data => ({ data })).catch(error => ({ error }))
    ]);
    setModels(nextModels);
    const h3 = nextModels.find(model => model.id === 'minimax-h3-video');
    if (h3 && !dirtyRef.current['minimax-h3-video']) setPresetWorkflowIds(current => ({ ...current, 'minimax-h3-video': h3.workflowId || DEFAULT_H3_WORKFLOW_ID }));
    const fanke = nextModels.find(model => model.id === 'fanke-video-api');
    if (fanke && !dirtyRef.current['fanke-video-api']) setPresetModelIds(current => ({ ...current, 'fanke-video-api': fanke.modelId || FANKE_VIDEO_MODELS[0].value }));
    if (nextQuotas.error) setQuotaError(nextQuotas.error.message || '额度查询失败');
    else { setQuotas(nextQuotas.data || []); setQuotaError(''); }
  }, [canManageApi]);

  useEffect(() => {
    let alive = true;
    Promise.all([getConfig(), getMemberCenter()]).then(async ([nextConfig, nextCenter]) => {
      if (!alive) return;
      setConfig(nextConfig);
      setCenter(nextCenter);
      if (canManageModelCatalog(nextConfig)) {
        const [nextModels, nextQuotas] = await Promise.all([
          listManagedModels(),
          getManagedModelQuotas().then(data => ({ data })).catch(error => ({ error }))
        ]);
        setModels(nextModels);
        const h3 = nextModels.find(model => model.id === 'minimax-h3-video');
        if (h3 && !dirtyRef.current['minimax-h3-video']) setPresetWorkflowIds(current => ({ ...current, 'minimax-h3-video': h3.workflowId || DEFAULT_H3_WORKFLOW_ID }));
        const fanke = nextModels.find(model => model.id === 'fanke-video-api');
        if (fanke && !dirtyRef.current['fanke-video-api']) setPresetModelIds(current => ({ ...current, 'fanke-video-api': fanke.modelId || FANKE_VIDEO_MODELS[0].value }));
        if (nextQuotas.error) setQuotaError(nextQuotas.error.message || '额度查询失败');
        else { setQuotas(nextQuotas.data || []); setQuotaError(''); }
      }
    }).catch(error => message.error(error.message || 'API 配置加载失败'))
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  }, []);

  const modelById = useMemo(() => new Map(models.map(model => [model.id, model])), [models]);
  const quotaById = useMemo(() => new Map(quotas.map(quota => [quota.modelId, quota])), [quotas]);
  // 不需要凭证的本地执行器：三组里按“平台托管”过滤掉，但不等于从下拉消失。
  // 在这里只读列出，保证「页面上看得到的」与「下拉里选得到的」一致。
  const noCredentialModels = useMemo(
    () => models.filter(model => modelNeedsNoCredential(model) && isCatalogModelAvailable(model)),
    [models]
  );

  // 串行队列：同一时刻只有一个请求在飞，后提交的一定后落库，
  // 避免连打时“先发的旧值覆盖后发的新值”。
  function enqueuePresetSave(task) {
    queueRef.current = queueRef.current.then(task).catch(error => {
      setSaveStatus('error');
      message.error(error.message || '保存失败');
    });
    return queueRef.current;
  }

  function presetPayload(preset) {
    const existing = new Map(modelsRef.current.map(model => [model.id, model])).get(preset.id);
    const key = String(keysRef.current[preset.id] || '').trim();
    // 输入过 Key 又清空 = 停用：显式清除凭证，页面与下拉同时消失。
    const clearing = !key && touchedKeysRef.current.has(preset.id) && existing?.hasCredential === true;
    const payload = {
      id: preset.id,
      kind: preset.kind || 'video',
      displayName: preset.displayName,
      enabled: !clearing
    };
    if (key) payload.credential = key;
    else if (clearing) payload.clearCredential = true;
    if (preset.id === 'minimax-h3-video') payload.workflowId = String(workflowRef.current[preset.id] || '').trim() || DEFAULT_H3_WORKFLOW_ID;
    if (preset.id === 'fanke-video-api') payload.modelId = modelIdsRef.current[preset.id] || FANKE_VIDEO_MODELS[0].value;
    if (preset.provider) {
      payload.provider = preset.provider;
      payload.baseUrl = String(baseURLsRef.current[preset.id] || existing?.baseUrl || 'http://127.0.0.1:8787').trim();
      payload.modelId = preset.modelId;
    }
    return { existing, key, payload };
  }

  async function persistPreset(preset) {
    const version = versionsRef.current[preset.id] || 0;
    const { existing, key, payload } = presetPayload(preset);
    // 新预设至少要有一个 Key 才建条目，避免只改工作流 ID 就生成空配置。
    if (!existing && !key) return;
    setPresetSavingId(preset.id);
    setSaveStatus('saving');
    try {
      if (existing) await updateManagedModel(preset.id, payload);
      else await createManagedModel(payload);
      // 只清掉刚刚提交的那份输入；保存期间又输入的新内容保留，等待下一轮保存。
      if (key && !focusedKeyRef.current.has(preset.id)) {
        setPresetKeys(current => (String(current[preset.id] || '').trim() === key ? { ...current, [preset.id]: '' } : current));
      }
      touchedKeysRef.current.delete(preset.id);
      await refreshModels();
      if ((versionsRef.current[preset.id] || 0) === version) dirtyRef.current[preset.id] = false;
      setSaveStatus('saved');
    } finally {
      setPresetSavingId('');
    }
  }

  // Key 输入只登记“有改动”，不启防抖定时器；真正提交发生在失焦或显式保存。
  function markPresetDirty(preset) {
    versionsRef.current[preset.id] = (versionsRef.current[preset.id] || 0) + 1;
    dirtyRef.current[preset.id] = true;
    if (timersRef.current[preset.id]) {
      clearTimeout(timersRef.current[preset.id]);
      delete timersRef.current[preset.id];
    }
  }

  function schedulePresetSave(preset) {
    versionsRef.current[preset.id] = (versionsRef.current[preset.id] || 0) + 1;
    dirtyRef.current[preset.id] = true;
    if (timersRef.current[preset.id]) clearTimeout(timersRef.current[preset.id]);
    timersRef.current[preset.id] = setTimeout(() => {
      delete timersRef.current[preset.id];
      if (!dirtyRef.current[preset.id]) return;
      enqueuePresetSave(() => persistPreset(preset));
    }, AUTOSAVE_DELAY);
  }

  function flushPreset(preset) {
    if (timersRef.current[preset.id]) {
      clearTimeout(timersRef.current[preset.id]);
      delete timersRef.current[preset.id];
    }
    if (!dirtyRef.current[preset.id]) return;
    enqueuePresetSave(() => persistPreset(preset));
  }

  // 目录模型凭证的唯一编辑入口（管理员）：写入 modelCatalog.credential，
  // 服务端据此重新计算 available，页面与 /api/models 下拉同时生效/消失。
  async function saveCatalogCredential(model, key, clear) {
    if (!canManageApi || !model?.id) throw new Error('无权修改目录模型凭证');
    const trimmed = String(key || '').trim();
    if (!clear && !trimmed) throw new Error('请输入 API Key');
    setCatalogModelSavingId(model.id);
    try {
      await updateManagedModel(model.id, clear ? { clearCredential: true } : { credential: trimmed, enabled: true });
      invalidateAvailableModelCatalog();
      await refreshModels();
      message.success(clear ? '已清除目录 Key，该模型不再出现在下拉里' : '目录 Key 已保存，立即生效');
    } catch (error) {
      message.error(error.message || '目录模型凭证保存失败');
      throw error;
    } finally {
      setCatalogModelSavingId('');
    }
  }

  // 用户自建目录模型的唯一编辑入口：与自定义渠道共用同一张表单，
  // 保存写回 modelCatalog 的这一条（displayName / modelId / baseUrl / credential）。
  async function saveCatalogModel(model, values) {
    if (!canManageApi || !model?.id) throw new Error('无权修改目录模型');
    setCatalogModelSavingId(model.id);
    try {
      const payload = {
        displayName: values.name,
        name: values.name,
        modelId: values.model,
        baseUrl: values.baseUrl
      };
      if (values.keyTouched && values.apiKey) payload.credential = values.apiKey;
      else if (values.keyTouched && !values.apiKey) payload.clearCredential = true;
      await updateManagedModel(model.id, payload);
      invalidateAvailableModelCatalog();
      await refreshModels();
      message.success('目录模型已保存，立即生效');
    } catch (error) {
      message.error(error.message || '目录模型保存失败');
      throw error;
    } finally {
      setCatalogModelSavingId('');
    }
  }

  if (loading) return <div className="account-center-page"><Skeleton active paragraph={{ rows: 9 }} /></div>;

  return <div className="account-center-page api-config-page">
    <PageHeader title="API 配置" subtitle="统一管理平台预设模型与各渠道 API 接口；业务页面按凭证自动读取可用模型。" />
    {!canManageApi ? <div className="ac-managed-api-card">
      <span><ShieldCheck size={28} /></span>
      <div><div><h2>模型服务由团队托管</h2>{member ? <RoleBadge role={member.role} /> : null}</div><p>你可在获授权的业务下拉框中选择管理员已配置凭证的模型；不会显示 API Key。</p></div>
      <Tag color={config?.managedBy ? 'green' : 'gold'}>{config?.managedBy ? `托管账号 @${config.managedBy}` : '等待绑定 MANAGER'}</Tag>
    </div> : <Panel title="平台预设模型" eyebrow="PLATFORM PRESETS" className="ac-form-panel" action={<Space size={12}>
        <span className="ac-muted-copy" role="status" aria-live="polite">{SAVE_STATUS_TEXT[saveStatus]}</span>
        <Button icon={<RefreshCw size={15} />} loading={quotaLoading} onClick={refreshQuotaState}>刷新额度</Button>
      </Space>}>
        <p className="ac-muted-copy">填好 API Key 后立即生效并出现在业务的视频模型下拉框；清空 Key 即从下拉移除，不需要额外开关。输入框失焦或停止输入约 0.8 秒自动保存；额度状态由服务商接口返回，不支持查询的服务商会显示“额度未知”。</p>
        {quotaError ? <Alert
          type="error"
          showIcon
          message="额度信息读取失败"
          description={quotaError}
          action={<Button size="small" onClick={refreshQuotaState} loading={quotaLoading}>重试</Button>}
          style={{ marginBottom: 12 }}
        /> : null}
        {PLATFORM_PRESETS.map(preset => {
          const model = modelById.get(preset.id);
          const cardState = getPresetCardState(preset, model, presetSavingId === preset.id);
          return <div className="ac-platform-preset-grid" key={preset.id}>
            <div className="ac-platform-preset-info"><span className="ac-security-card-icon violet"><Video size={20} /></span><div><strong>{preset.displayName}</strong><small>{preset.description}</small></div></div>
            <div className="ac-platform-preset-credential">
              <Space direction="vertical" style={{ width: '100%' }} size={6}>
                {preset.provider ? <Input disabled={!cardState.editable} value={presetBaseURLs[preset.id] ?? model?.baseUrl ?? 'http://127.0.0.1:8787'} onChange={event => { setPresetBaseURLs(current => ({ ...current, [preset.id]: event.target.value })); schedulePresetSave(preset); }} onBlur={() => flushPreset(preset)} addonBefore="Relay 地址" placeholder="http://127.0.0.1:8787" /> : null}
                <Input.Password disabled={!cardState.editable} value={presetKeys[preset.id] || ''} onChange={event => { touchedKeysRef.current.add(preset.id); setPresetKeys(current => ({ ...current, [preset.id]: event.target.value })); markPresetDirty(preset); }} onFocus={() => focusedKeyRef.current.add(preset.id)} onBlur={() => { focusedKeyRef.current.delete(preset.id); flushPreset(preset); }} prefix={<KeyRound size={15} />} placeholder={model?.hasCredential ? '已保存；输入新 Key 覆盖，输入后清空并离开即停用' : '填写 API Key 后生效'} />
                {preset.id === 'minimax-h3-video' ? <Input disabled={!cardState.editable} value={presetWorkflowIds[preset.id] || DEFAULT_H3_WORKFLOW_ID} onChange={event => { setPresetWorkflowIds(current => ({ ...current, [preset.id]: event.target.value })); schedulePresetSave(preset); }} onBlur={() => flushPreset(preset)} addonBefore="AutoDL 工作流 ID" placeholder={DEFAULT_H3_WORKFLOW_ID} /> : null}
                {preset.id === 'fanke-video-api' ? <Select disabled={!cardState.editable} value={presetModelIds[preset.id] || model?.modelId || FANKE_VIDEO_MODELS[0].value} onChange={value => { setPresetModelIds(current => ({ ...current, [preset.id]: value })); schedulePresetSave(preset); flushPreset(preset); }} options={FANKE_VIDEO_MODELS} placeholder="选择视频模型" /> : null}
              </Space>
              <ModelQuotaStatus quota={quotaById.get(preset.id)} error={quotaError} />
            </div>
            <div className="ac-platform-preset-state">
              <small>{presetSavingId === preset.id ? '保存中' : cardState.configured ? 'Key 已填写，已生效' : '填写 Key 后生效'}</small>
            </div>
          </div>;
        })}
        {noCredentialModels.map(model => (
          <div className="ac-platform-preset-grid" key={model.id}>
            <div className="ac-platform-preset-info"><span className="ac-security-card-icon violet"><Cpu size={20} /></span><div><strong>{model.displayName || model.id}</strong><small>本地执行器：不需要 API Key，始终出现在业务下拉框里。</small></div></div>
            <div className="ac-platform-preset-credential"><span className="ac-muted-copy">无需填写凭证</span></div>
            <div className="ac-platform-preset-state"><small>始终可用</small></div>
          </div>
        ))}
      </Panel>}
    <ApiIntegrationSettings
      models={models}
      quotas={quotas}
      platformPresetIds={PLATFORM_PRESET_IDS}
      onSaveCatalogCredential={canManageApi ? saveCatalogCredential : null}
      onUpdateCatalogModel={canManageApi ? saveCatalogModel : null}
      catalogModelSavingId={catalogModelSavingId}
    />
  </div>;
}
