import {
  Button,
  Divider,
  Drawer,
  Segmented,
  Select,
  Space,
  Tag,
  Typography
} from 'antd';
import { CloudDownload, Layers3, Save, Settings2, TimerReset } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { BatchFactoryV11ConstraintEditor } from './BatchFactoryV11ConstraintEditor';
import { ChangeImpactNotice } from './ChangeImpactNotice';
import { FixedSingleVideoControl } from './FixedSingleVideoControl';
import { buildConfigVersionSyncPatch, configVersionOptions } from './bf11UiAdapter.js';
import { runSaveFlow } from './saveFlow.js';
import './batch-factory-v11-settings.css';
import providerCapabilities from './videoProviderCapabilities.json';

const full = { width: '100%' };
const CHANGE_IMPACT_DEBOUNCE_MS = 250;

// 通道清单来自共享契约文件，管理员模型目录使用同一份数据。
const VIDEO_PROVIDERS = providerCapabilities.providers.map(({ id, label }) => ({ value: id, label }));
const VIDEO_PROVIDER_DEFAULT_MODELS = providerCapabilities.providers.reduce((result, item) => {
  result[item.id] = item.defaultModel;
  return result;
}, {});

function configuredModelOptions(models = []) {
  return (Array.isArray(models) ? models : [])
    .filter(model => model?.enabled !== false)
    .map(model => ({
      value: model.id || model.modelId,
      label: model.displayName || model.name || model.modelId || model.id
    }))
    .filter(option => option.value);
}


function SettingField({ label, description, children }) {
  return <div className="bf11-setting-field">
    <div className="bf11-setting-field-copy">
      <Typography.Text strong>{label}</Typography.Text>
      {description ? <Typography.Text type="secondary">{description}</Typography.Text> : null}
    </div>
    <div className="bf11-setting-field-control">{children}</div>
  </div>;
}

export function ProductionSettingsDrawer({
  open,
  batch,
  initialValue = {},
  configVersions = [],
  configVersionsError = null,
  personalPrompts = {},
  personalPromptsError = null,
  onClose,
  onSave,
  onSyncConfigVersion,
  onPreviewChangeImpact,
  onSaveDraft,
  onSavePersonalPrompt,
  videoProviders = {},
  apiModels = {},
  promptCatalog = [],
  localExecutors = [],
  onCreateLocalExecutorPairing
}) {
  const [form, setForm] = useState(initialValue);
  const [selectedConfigVersionId, setSelectedConfigVersionId] = useState(initialValue.versionConfigId || '');
  const [saving, setSaving] = useState(false);
  const [syncingConfigVersion, setSyncingConfigVersion] = useState(false);
  const [impactResult, setImpactResult] = useState(null);
  const [impactLoading, setImpactLoading] = useState(false);
  const impactTimerRef = useRef(null);
  const impactRequestRef = useRef(0);
  const [pairingBusy, setPairingBusy] = useState(false);
  const [pairingSecret, setPairingSecret] = useState(null);
  const configuredTextModels = useMemo(() => configuredModelOptions(apiModels.text), [apiModels.text]);
  const configuredImageModels = useMemo(() => configuredModelOptions(apiModels.image), [apiModels.image]);
  const configuredVideoModels = useMemo(() => configuredModelOptions(apiModels.video), [apiModels.video]);

  useEffect(() => {
    if (impactTimerRef.current) {
      clearTimeout(impactTimerRef.current);
      impactTimerRef.current = null;
    }
    impactRequestRef.current += 1;
    setImpactLoading(false);
    setImpactResult(null);
    if (!open) return undefined;
    setForm({ videoProvider: 'personal_api', ...initialValue });
    setSelectedConfigVersionId(initialValue.versionConfigId || '');
    setPairingSecret(null);
    return () => {
      if (impactTimerRef.current) clearTimeout(impactTimerRef.current);
      impactTimerRef.current = null;
      impactRequestRef.current += 1;
    };
  }, [open, batch?.id, initialValue]);

  const enabledConstraintCount = useMemo(() => [
    form.injectBaseSettings === true,
    form.injectCharacterPrompt === true,
    form.injectScenePrompt === true,
    form.injectPropPrompt === true,
    form.prefixEnabled === true,
    form.qualityEnabled === true,
    form.restrictionEnabled === true,
    form.negativeEnabled === true
  ].filter(Boolean).length, [form]);

  const configOptions = useMemo(() => configVersionOptions(configVersions), [configVersions]);
  const canSyncConfigVersion = useMemo(
    () => !configVersionsError
      && selectedConfigVersionId !== form.versionConfigId
      && Boolean(buildConfigVersionSyncPatch(configVersions, selectedConfigVersionId)),
    [configVersions, configVersionsError, selectedConfigVersionId, form.versionConfigId]
  );

  function scheduleImpactPreview(proposedPatch) {
    if (!onPreviewChangeImpact) return;
    if (impactTimerRef.current) clearTimeout(impactTimerRef.current);
    const requestId = impactRequestRef.current + 1;
    impactRequestRef.current = requestId;
    setImpactResult(null);
    impactTimerRef.current = setTimeout(async () => {
      impactTimerRef.current = null;
      setImpactLoading(true);
      try {
        const result = await onPreviewChangeImpact(proposedPatch);
        if (impactRequestRef.current === requestId) {
          setImpactResult(result || { ok: false, status: 0, message: '无法读取影响' });
        }
      } catch (error) {
        if (impactRequestRef.current === requestId) {
          setImpactResult({
            ok: false,
            status: Number(error?.status || 0),
            message: error?.message || '读取变更影响失败，请稍后重试。'
          });
        }
      } finally {
        if (impactRequestRef.current === requestId) setImpactLoading(false);
      }
    }, CHANGE_IMPACT_DEBOUNCE_MS);
  }

  function patch(next) {
    const proposedPatch = { ...form, ...next };
    setForm(proposedPatch);
    scheduleImpactPreview(proposedPatch);
  }

  async function syncConfigVersion() {
    const syncPatch = buildConfigVersionSyncPatch(configVersions, selectedConfigVersionId);
    if (!syncPatch || !onSyncConfigVersion) return false;
    setImpactLoading(true);
    let impact;
    try {
      impact = onPreviewChangeImpact
        ? await onPreviewChangeImpact(syncPatch)
        : { ok: true, impact: null };
      setImpactResult(impact);
    } finally {
      setImpactLoading(false);
    }
    if (!impact?.ok) return false;
    setSyncingConfigVersion(true);
    try {
      const saved = await onSyncConfigVersion(syncPatch);
      if (saved) setForm(current => ({ ...current, ...syncPatch }));
      return saved;
    } finally {
      setSyncingConfigVersion(false);
    }
  }

  async function createPairing() {
    if (!onCreateLocalExecutorPairing || pairingBusy) return;
    setPairingBusy(true);
    try {
      const result = await onCreateLocalExecutorPairing('doubao');
      if (!result?.ok) {
        setPairingSecret({ error: result?.message || '配对码生成失败' });
        return;
      }
      setPairingSecret(result.raw || {});
    } finally {
      setPairingBusy(false);
    }
  }

  async function save() {
    setSaving(true);
    try {
      return await runSaveFlow({ payload: form, onSave, onClose });
    } finally {
      setSaving(false);
    }
  }

  return <Drawer
    title={<Space><Settings2 size={18} /><span>生产统一设置</span></Space>}
    width={820}
    open={open}
    onClose={onClose}
    destroyOnClose={false}
    extra={<Button type="primary" loading={saving} icon={<Save size={15} />} onClick={save}>保存生产统一设置</Button>}
  >
    <div className="bf11-settings-drawer">
      <div className="bf11-settings-scope">
        <div>
          <Typography.Text strong>应用于当前批次</Typography.Text>
          <Typography.Text type="secondary">{batch?.title || '当前批次'} · {batch?.count || 0} 本小说</Typography.Text>
        </div>
        <Space wrap>
          {form.productionMode ? <Tag color="blue">{form.productionMode === 'viral' ? '爆款开头' : '原文直转'}</Tag> : <Tag>生产方式继承系统</Tag>}
          {form.aspectRatio ? <Tag>{form.aspectRatio}</Tag> : <Tag>画幅继承系统</Tag>}
          <Tag>{enabledConstraintCount} 项显式启用</Tag>
        </Space>
      </div>

      <Divider orientation="left">配置版本</Divider>
      <section className="bf11-setting-section bf11-config-version-card">
        <div className="bf11-config-version-head">
          <div>
            <Typography.Text strong>当前生产配置</Typography.Text>
            <Typography.Text type="secondary">这里只保存当前批次自己的 sparse patch；未设置字段继续继承系统层。</Typography.Text>
          </div>
          <Tag>{form.versionConfigId || '继承系统'}</Tag>
        </div>
        <div className="bf11-config-version-actions">
          <Select
            placeholder={configOptions.length ? '选择服务端配置版本' : '服务端暂无配置版本'}
            value={selectedConfigVersionId || undefined}
            onChange={setSelectedConfigVersionId}
            options={configOptions}
            disabled={Boolean(configVersionsError) || !configOptions.length}
            style={full}
          />
          <Button
            icon={<CloudDownload size={15} />}
            loading={syncingConfigVersion}
            disabled={!canSyncConfigVersion || !onSyncConfigVersion}
            title={canSyncConfigVersion ? '把所选服务端版本写入当前 Batch sparse patch' : '请先选择服务端返回的配置版本'}
            onClick={syncConfigVersion}
          >同步批量后台配置</Button>
        </div>
        {configVersionsError
          ? <Typography.Text type="danger">配置版本读取失败：{configVersionsError.message || '服务端 catalog 不可用'}。同步保持关闭，不使用本地 fallback。</Typography.Text>
          : <Typography.Text type="secondary">版本列表和 version ID 只来自 V11 服务端；选择后点击同步只保存当前 Batch 的 versionConfigId。前端不会推断最新版本，也不会修改 Book / VIDEO override。</Typography.Text>}
      </section>

      <ChangeImpactNotice result={impactResult} loading={impactLoading} />

      <Divider orientation="left">提示词配置</Divider>
      <section className="bf11-setting-section">
        <Typography.Text type="secondary">提示词来自管理后台“系统预设词 → 批量工厂”。下拉框只显示已发布提示词名称和版本。</Typography.Text>
        {[
          ['hook', 'Hook 提示词'],
          ['director', 'Director 提示词'],
          ['visual', '画面提示词'],
          ['video', '视频提示词'],
          ['audio_match', '音频匹配提示词'],
          ['shot_merge', 'Shot 合成提示词'],
          ['book_merge', '成片合成提示词']
        ].map(([type, label]) => {
          const options = promptCatalog.filter(item => item?.type === type && item?.enabled !== false)
            .map(item => ({ value: item.id, label: item.version ? label + ' · ' + item.name + ' · v' + item.version : label + ' · ' + item.name }));
          return <SettingField key={type} label={label} description="由后台预设词库动态提供。">
            <Select allowClear placeholder="继承系统默认提示词" value={form.promptSelections?.[type]} options={options}
              onChange={value => patch({ promptSelections: { ...(form.promptSelections || {}), [type]: value || '' } })} />
          </SettingField>;
        })}
      </section>

      <Divider orientation="left">基础生产设置</Divider>
      <section className="bf11-setting-section">
        <SettingField label="生产方式" description="未选择时继承系统层；选择后写入当前批次 patch。">
          <Segmented
            value={form.productionMode}
            onChange={productionMode => patch({ productionMode })}
            options={[
              { value: 'original', label: '原文直转' },
              { value: 'viral', label: '爆款开头' }
            ]}
          />
        </SettingField>

        <SettingField label="剧本提示词" description="普通生产页只选择名称；提示词正文由 V11 Prompt 库管理。">
          <Select
            allowClear
            placeholder="继承系统提示词"
            value={form.scriptPromptPresetId}
            onChange={scriptPromptPresetId => patch({ scriptPromptPresetId })}
            options={[
              { value: 'standard-short-drama', label: '标准短剧分镜' },
              { value: 'commercial-dynamic', label: '商业动态分镜' },
              { value: 'spatial-continuity', label: '空间连续分镜' }
            ]}
          />
        </SettingField>

        <SettingField label="人物场景提示词" description="用于人物 / 场景 / 道具基础资产提取。">
          <Select
            allowClear
            placeholder="继承系统资产提示词"
            value={form.assetPromptPresetId}
            onChange={assetPromptPresetId => patch({ assetPromptPresetId })}
            options={[{ value: 'standard-asset-extraction', label: '标准资产提取' }]}
          />
        </SettingField>

        <SettingField label="文本模型" description="Director / Hook 只使用 API 配置中已启用的文本模型。">
          <Select
            allowClear
            placeholder="继承系统文本模型"
            value={form.textModelId}
            onChange={textModelId => patch({ textModelId })}
            options={configuredTextModels}
            style={full}
          />
        </SettingField>

        <SettingField label="图片模型" description="人物、场景、道具和画面图生成使用 API 配置中已启用的图片模型。">
          <Select
            allowClear
            placeholder="继承系统图片模型"
            value={form.imageModelId}
            onChange={imageModelId => patch({ imageModelId })}
            options={configuredImageModels}
            style={full}
          />
        </SettingField>

        <SettingField label="视频生成通道" description="默认使用个人中心 API；选择豆包本地执行器后，任务会交给你已配对且在线的本地电脑执行。">
          <Space direction="vertical" style={full} size={8}>
            <Select
              value={form.videoProvider || 'personal_api'}
              onChange={videoProvider => patch({
                videoProvider,
                videoModelId: VIDEO_PROVIDER_DEFAULT_MODELS[videoProvider] || 'yd2.0-mini'
              })}
              options={VIDEO_PROVIDERS}
              style={full}
            />
            <Space wrap>
              <Tag color={videoProviders.personalAPI?.configured ? 'green' : 'default'}>
                个人 API {videoProviders.personalAPI?.configured ? '已配置' : '未配置'}
              </Tag>
              <Tag color={localExecutors.some(item => item.online) ? 'green' : 'default'}>
                豆包执行器 {localExecutors.some(item => item.online) ? '在线' : '未在线'}
              </Tag>
              <Tag color={videoProviders.h3?.configured ? 'green' : 'default'}>
                AutoDL H3 {videoProviders.h3?.configured ? '已配置' : '未配置'}
              </Tag>
              <Tag color={videoProviders.yfai?.configured ? 'green' : 'default'}>
                梵客 Seedance {videoProviders.yfai?.configured ? '已配置' : '未配置'}
              </Tag>
            </Space>
            {(form.videoProvider || 'personal_api') === 'doubao_local_executor' ? <div className="bf11-provider-pairing">
              <Button size="small" loading={pairingBusy} disabled={!onCreateLocalExecutorPairing} onClick={createPairing}>生成豆包配对码</Button>
              {pairingSecret?.code ? <Typography.Text copyable={{ text: pairingSecret.code }}>配对码：{pairingSecret.code}（10 分钟内有效）</Typography.Text> : null}
              {pairingSecret?.error ? <Typography.Text type="danger">{pairingSecret.error}</Typography.Text> : null}
              <Typography.Text type="secondary">在你的 Mac 执行器中输入配对码并保持豆包账号已登录；本页面不会保存或读取豆包密码。</Typography.Text>
            </div> : null}
          </Space>
        </SettingField>

        <SettingField label="视频模型" description="导演前绑定模型；最大时长表示单个 VIDEO 的能力上限。">
          <Select
            allowClear
            placeholder="继承系统模型"
            value={form.videoModelId}
            onChange={videoModelId => {
              const selected = (apiModels.video || []).find(model => (model.id || model.modelId) === videoModelId);
              const videoProvider = selected?.adapterKind === 'local_executor_video'
                ? 'doubao_local_executor'
                : selected?.adapterKind === 'autodl_comfyui_video' ? 'autodl_comfyui' : 'personal_api';
              patch({ videoModelId, videoProvider });
            }}
            options={configuredVideoModels}
          />
        </SettingField>

        <SettingField label="视频画幅" description="未覆盖时继续继承系统层。">
          <Segmented
            value={form.aspectRatio}
            onChange={aspectRatio => patch({ aspectRatio })}
            options={['9:16', '16:9']}
          />
        </SettingField>

        <SettingField label="VIDEO 时长策略" description="AI 自动按剧情密度自然分配时长，不要求凑满模型最大值。">
          <Segmented
            value={form.durationMode}
            onChange={durationMode => patch({ durationMode })}
            options={[
              { value: 'auto', label: 'AI 自动' },
              { value: 'model-max', label: '按模型上限规划' }
            ]}
          />
        </SettingField>

        <SettingField label="固定单 VIDEO" description="只有用户操作开关后才写入 true / false；显式 false 也会被保留。">
          <FixedSingleVideoControl
            checked={form.fixedSingleVideo === true}
            hasOverride={Object.prototype.hasOwnProperty.call(form, 'fixedSingleVideo')}
            maxDurationSeconds={Number(batch?.modelMaxDuration || batch?.modelCapability?.maxDurationSeconds || 0)}
            onChange={fixedSingleVideo => patch({ fixedSingleVideo })}
          />
        </SettingField>

        <SettingField label="前缀模式" description="自动模式由内容判断前缀；手动模式使用统一指定前缀。">
          <Segmented
            value={form.prefixMode}
            onChange={prefixMode => patch({ prefixMode })}
            options={[
              { value: 'auto', label: 'AI 自动' },
              { value: 'manual', label: '手动统一' }
            ]}
          />
        </SettingField>

        <SettingField label="字幕策略" description="“禁止自动对白字幕”不等于禁止人物说话或 Lip-sync。">
          <Select
            allowClear
            placeholder="继承系统字幕策略"
            value={form.subtitlePolicy}
            onChange={subtitlePolicy => patch({ subtitlePolicy })}
            options={[
              { value: 'forbid-auto-dialogue-subtitle', label: '禁止自动对白字幕' },
              { value: 'allow', label: '允许自动字幕' }
            ]}
          />
        </SettingField>
      </section>

      <Divider orientation="left">约束设置</Divider>
      <section className="bf11-setting-section bf11-constraint-section">
        <div className="bf11-section-intro">
          <Space><Layers3 size={17} /><Typography.Text strong>生成约束</Typography.Text></Space>
          <Typography.Text type="secondary">所有开关都在当前位置直接编辑；不再进入二级设置。</Typography.Text>
        </div>
        <BatchFactoryV11ConstraintEditor
          value={form}
          onChange={patch}
          scopeLabel={`应用于当前批次 · ${batch?.count || 0} 本小说`}
          scope={batch?.id || ''}
          personalPrompts={personalPrompts}
          personalPromptsError={personalPromptsError}
          onSaveDraft={onSaveDraft}
          onSavePersonalPrompt={onSavePersonalPrompt}
        />
      </section>

      <div className="bf11-settings-footer-note">
        <TimerReset size={16} />
        <Typography.Text type="secondary">保存统一写入批量工厂 V12 Go 服务的 SettingsState；Book / VIDEO sparse override 不会被清空。</Typography.Text>
      </div>
    </div>
  </Drawer>;
}

export default ProductionSettingsDrawer;
