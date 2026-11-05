import { SettingOutlined } from '@ant-design/icons';
import { Alert, Button, Divider, Input, InputNumber, Modal, Segmented, Select, Space, Switch, Tabs, Tooltip, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { listAvailableModels, modelSelectOptions } from '../../../shared/api/modelCatalog';
import { checkWebSubmitEnvironment, getWebSubmitConfig, saveWebSubmitConfig, syncWebSubmitConfigs, syncWebSubmitStyles, testWebSubmitVisible } from '../../../shared/api/novelFetch';
import { apiErrorMessage } from '../../../shared/apiErrors';
import { IMAGE_ASPECT_RATIOS } from '../../../shared/imageAspectRatios';
import {
  createAutomationPreset,
  deleteAutomationPreset,
  get121OrganizationOptions,
  getN8ComicPublishStatus,
  getN8ComicPublishAudioCategories,
  getN8ComicPublishAudios,
  listAutomationPresets,
  listSystemPresetCatalog,
  updateAutomationPreset
} from '../../../shared/api/batchFactory';
import { modelOptionsForProvider, normalizeModelProvider, providerForCatalogModel, providerModelSettingsPatch, providerOptionsForModels } from './providerModelBinding';
import { normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { getConstraintPresetTexts } from '../../../shared/api/generation';
import { BatchFactoryComicStylePicker } from './BatchFactoryComicStylePicker';
import { batchFactoryProductionModeVisibility } from './batchFactoryProductionMode';
import { DEFAULT_AUDIO_CATEGORIES, chooseN8Audio, producerAccountOptions, readComicPublishDirectoryCache, writeComicPublishDirectoryCache } from './batchFactoryComicPublishAccounts';
import { inferBatchN8AudioCategory } from './batchFactoryAudioCategory';

const clone = value => JSON.parse(JSON.stringify(value || {}));
const normalizeLegacyPatch = value => {
  const patch = clone(value);
  delete patch.expectedRevision;
  if (!patch.patch || typeof patch.patch !== 'object' || Array.isArray(patch.patch)) return patch;
  const nested = normalizeLegacyPatch(patch.patch);
  delete patch.patch;
  return { ...nested, ...patch };
};
const DEFAULT_TTS = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 };
const DEFAULT_VIDEO_PROMPT = { presetId: 'batch-video-sd', presetName: 'SD 视频提示词', presetSlot: 'batch.video-meta', presetVersion: 1, enabled: true };
// N8 目录读取失败必须分三类展示：会话登录失效（401/403）、服务端未配置、
// 上游业务错误。401 一旦被写成“N8 令牌失效”，就会把登录问题误报成配置问题。
function describeN8DirectoryFailure(error) {
	const status = Number(error?.status || 0);
	if (status === 401 || status === 403) return { kind: 'auth', message: '登录已失效，请重新登录' };
	if (status >= 500 || status === 0) return { kind: 'unavailable', message: 'N8 目录服务暂时不可用，请稍后重试' };
	const detail = String(error?.message || '').trim();
	return { kind: 'upstream', message: detail || 'N8 目录服务读取失败，请稍后重试' };
}
function requiredVideoPrompt(value) {
  const selected = value && typeof value === 'object' ? value : {};
  const presetId = String(selected.presetId || selected.presetKey || '').trim();
  if (!presetId || presetId === 'batch-video-meta' || presetId === '__none__') return { ...DEFAULT_VIDEO_PROMPT };
  return { ...DEFAULT_VIDEO_PROMPT, ...selected, enabled: true };
}
const AUDIO_SOURCE_OPTIONS = [{ value: 'tts', label: 'TTS（系统配音）' }, { value: 'original', label: '原声（视频自带声音）' }];
const voices = [
  { value: 'zh-CN-XiaoxiaoNeural', label: '晓晓（女声·温柔）' },
  { value: 'zh-CN-XiaochenNeural', label: '晓辰（女声·知性）' },
  { value: 'zh-CN-YunxiNeural', label: '云希（男声·清朗）' },
  { value: 'zh-CN-YunyangNeural', label: '云扬（男声·阳光）' }
];
const styles = ['general', 'cheerful', 'sad', 'friendly', 'chat'].map(value => ({ value, label: value === 'general' ? '通用' : value }));
const selectOption = preset => ({ value: preset.id, label: `${preset.name || preset.id} · v${preset.version || 1}` });
const SMART_UNIFIED_PRESET_ID = 'script-constraint-prefix-smart-unified';
const presetValue = preset => preset ? { presetId: preset.id, presetName: preset.name || preset.id, presetSlot: preset.slot || '', presetVersion: preset.version || 1, constraintCategory: preset.constraintCategory || (preset.id === SMART_UNIFIED_PRESET_ID ? 'prefix' : '') } : { presetId: '', presetName: '', presetSlot: '', presetVersion: null, constraintCategory: '' };
const TARGET_ADMIN_URL = 'http://two.121w.com/tttadmin/index.php';
const N8_MATERIAL_ALBUM_URL = 'https://n8-view.hnqingyuwen.top/#/material/album';
const websiteProfileOrganization = profile => {
  const advanced = profile?.advanced || profile?.config_data || {};
  return profile?.organization
    || profile?.organization_id
    || profile?.organizationId
    || profile?.org_id
    || advanced?.organization_id
    || advanced?.organizationId
    || advanced?.organization
    || advanced?.org_id
    || '';
};
// 小说获取旧版浏览器登录器仍返回“目标站登录会话”。它与当前直连
// 登录器的两个名称表达的是同一份账号级会话，不能因此把已验证会话
// 误判为未登录。
const SESSION_CHECK_NAMES = ['视频管理系统登录会话', '121 后台登录会话', '目标站登录会话'];

function HelpTip({ title, children }) {
  return <Tooltip title={children}><button type="button" className="batch-factory-help-button" aria-label={`查看${title}说明`}>?</button></Tooltip>;
}

export function BatchFactoryProductionModeControl({ value, onChange, disabled = false }) {
  const productionMode = batchFactoryProductionModeVisibility(value).mode;
  return <Segmented className={`batch-factory-production-mode-switch is-${productionMode}`} size="small" value={productionMode} options={[{ value: 'video', label: '视频生成' }, { value: 'comic', label: '漫画生成' }]} disabled={disabled} onChange={onChange} />;
}

export function BatchFactoryEngineSettingsForm({ value, onChange, sections = ['models', 'audio', 'publish'], active = true }) {
  const [models, setModels] = useState([]);
  const [modelsError, setModelsError] = useState('');
  const [modelsLoading, setModelsLoading] = useState(false);
  const [modelReloadKey, setModelReloadKey] = useState(0);
  const patch = next => onChange({ ...value, ...next });
  const tts = { ...DEFAULT_TTS, ...(value.tts || {}) };
  const visibility = batchFactoryProductionModeVisibility(value.productionMode);
  const modelSettingsEnabled = sections.includes('models');
  const modelOptions = useMemo(() => ({
    text: modelSelectOptions(models, 'text'),
    image: modelSelectOptions(models, 'image'),
    video: modelSelectOptions(models, 'video')
  }), [models]);
  const imageProviderOptions = useMemo(() => providerOptionsForModels(models, 'image'), [models]);
  const videoProviderOptions = useMemo(() => providerOptionsForModels(models, 'video'), [models]);
  const videoModelOptions = modelOptionsForProvider(models, 'video', value.videoProvider, value.videoModelId, { expandFanKeModels: true });
  const imageModelOptions = modelOptionsForProvider(models, 'image', value.imageProvider);
  useEffect(() => {
    if (!active || !modelSettingsEnabled) return undefined;
    let alive = true;
    setModelsLoading(true);
    setModelsError('');
    Promise.all([
      listAvailableModels('text'),
      listAvailableModels('image'),
      listAvailableModels('video')
    ]).then(groups => {
      if (alive) setModels(groups.flat());
    }).catch(error => {
      if (!alive) return;
      setModels([]);
      setModelsError(error?.message || '未能读取个人中心已启用模型');
    }).finally(() => { if (alive) setModelsLoading(false); });
    return () => { alive = false; };
  }, [active, modelSettingsEnabled, modelReloadKey]);
  const publish = value.publishSettings || {};
  return <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>当前批次统一生产参数 <HelpTip title="当前批次统一生产参数">保存后作为全部小说的默认值；单书已保存的显式覆盖不会被预设载入或本页保存清空。</HelpTip></span>} />
    {modelSettingsEnabled ? <><section className="batch-factory-engine-card"><header><b>模型配置</b><HelpTip title="模型配置">文本、图片、视频和配音。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>文本模型</b></span><Select allowClear loading={modelsLoading} value={value.textModelId || undefined} options={modelOptions.text} placeholder="选择已启用文本模型" onChange={textModelId => patch({ textModelId: textModelId || '' })} /></label>
      <label className="batch-factory-engine-field"><span><b>图片生成通道</b></span><Select value={value.imageProvider || 'personal_api'} options={imageProviderOptions} onChange={imageProvider => patch(providerModelSettingsPatch({ models, kind: 'image', provider: imageProvider, modelId: value.imageModelId }))} /></label>
      <label className="batch-factory-engine-field"><span><b>图片模型</b></span><Select allowClear loading={value.imageProvider === 'doubao_local_executor' ? false : modelsLoading} value={value.imageModelId || undefined} options={imageModelOptions} placeholder="选择当前通道的图片模型" onChange={imageModelId => patch({ imageModelId: imageModelId || '', ...(imageModelId ? { imageProvider: providerForCatalogModel(models.find(model => model.id === imageModelId), value.imageProvider) } : {}) })} /></label>
      <label className="batch-factory-engine-field"><span><b>图片画幅</b><HelpTip title="图片画幅">用于漫画成图与视频参考图；资产图默认 9:16，可单独覆盖。</HelpTip></span><Segmented value={value.imageAspectRatio || value.aspectRatio || '9:16'} options={IMAGE_ASPECT_RATIOS} onChange={imageAspectRatio => patch({ imageAspectRatio })} /></label>
      <label className="batch-factory-engine-field"><span><b>出图并发</b><HelpTip title="出图并发">同一批次同时向上游提交的图片请求数量，漫画资产图与漫画页共用。默认 2 路，上游限流时下调，网络稳定时可上调到 4 路。</HelpTip></span><Segmented value={[1, 2, 3, 4].includes(Number(value.imageConcurrency)) ? Number(value.imageConcurrency) : 2} options={[{ value: 1, label: '1 路' }, { value: 2, label: '2 路' }, { value: 3, label: '3 路' }, { value: 4, label: '4 路' }]} onChange={imageConcurrency => patch({ imageConcurrency })} /></label>
      <label className="batch-factory-engine-field"><span><b>生成资产图片</b><HelpTip title="生成资产图片">默认开启。关闭后跳过自动化的资产生成阶段（不调用生图、不消耗额度），不影响分镜、视频与漫画流程；已有资产图片仍会照常使用，没有资产图片时沿用空图兜底继续生成。</HelpTip></span><Switch checked={value.assetImagesEnabled !== false} onChange={assetImagesEnabled => patch({ assetImagesEnabled })} /></label>
      <label className="batch-factory-engine-field"><span><b>添加水印</b><HelpTip title="添加水印">开启后在每张漫画图片右下角叠加白底方框文字。</HelpTip></span><Switch checked={value.watermarkEnabled === true} onChange={watermarkEnabled => patch({ watermarkEnabled })} /></label>
      {value.watermarkEnabled === true ? <label className="batch-factory-engine-field"><span><b>水印文字</b><HelpTip title="水印文字">显示在图片右下角的文字，默认为“热门小说”。</HelpTip></span><Input value={value.watermarkText || '热门小说'} placeholder="热门小说" maxLength={20} onChange={event => patch({ watermarkText: event.target.value })} /></label> : null}
      {visibility.video ? <>
      <label className="batch-factory-engine-field"><span><b>视频引擎</b></span><Select value={normalizeModelProvider(value.videoProvider || 'personal_api', 'personal_api', value.videoModelId)} options={videoProviderOptions} onChange={videoProvider => patch(providerModelSettingsPatch({ models, kind: 'video', provider: videoProvider, modelId: value.videoModelId, expandFanKeModels: true }))} /></label>
      <div className="batch-factory-engine-field"><span><b>视频模型</b><HelpTip title="视频模型">模型列表只显示当前视频引擎的可用模型；选择模型时会同步对应的生成通道。</HelpTip></span><Select allowClear loading={value.videoProvider === 'doubao_local_executor' ? false : modelsLoading} value={value.videoModelId || undefined} options={videoModelOptions} placeholder="选择当前通道的视频模型" onChange={videoModelId => patch({ videoModelId: videoModelId || '', ...(videoModelId ? { videoProvider: providerForCatalogModel(models.find(model => model.id === videoModelId), value.videoProvider) } : {}) })} /></div>
      <label className="batch-factory-engine-field"><span><b>视频画幅</b><HelpTip title="视频画幅">视频不会跟随参考图画幅。</HelpTip></span><Segmented value={value.videoAspectRatio || (value.aspectRatio === '16:9' ? '16:9' : '9:16')} options={['16:9', '9:16']} onChange={videoAspectRatio => patch({ videoAspectRatio })} /></label>
      <label className="batch-factory-engine-field"><span><b>视频分辨率</b></span><Segmented value={value.videoResolution || '720p'} options={['480p', '720p', '1080p']} onChange={videoResolution => patch({ videoResolution })} /></label>
      <label className="batch-factory-engine-field"><span><b>分镜时长</b></span><Segmented value={Number(value.storyboardDurationLimit) === 15 ? 15 : 10} options={[{ value: 10, label: '10 秒' }, { value: 15, label: '15 秒' }]} onChange={storyboardDurationLimit => patch({ storyboardDurationLimit, maxVideoDuration: storyboardDurationLimit })} /></label>
      <label className="batch-factory-engine-field"><span><b>音频来源</b><HelpTip title="音频来源">默认使用 TTS，原声保留视频模型自带声音。</HelpTip></span><Segmented block value={normalizeAudioSource(value.audioSource)} options={AUDIO_SOURCE_OPTIONS} onChange={audioSource => onChange(normalizeAudioSourceSettings({ ...value, audioSource }))} /></label>
      <label className="batch-factory-engine-field"><span><b>配音音色</b></span><Select value={tts.voice} options={voices} onChange={voice => patch({ tts: { ...tts, voice } })} /></label>
      <label className="batch-factory-engine-field"><span><b>配音风格</b></span><Select value={tts.style} options={styles} onChange={style => patch({ tts: { ...tts, style } })} /></label>
      <label className="batch-factory-engine-field"><span><b>语速</b></span><InputNumber min={0.5} max={2} step={0.1} value={tts.speed} onChange={speed => patch({ tts: { ...tts, speed: speed ?? DEFAULT_TTS.speed } })} /></label>
      </> : null}
    </div></section>
    {modelsError ? <Alert type="warning" showIcon message="模型目录暂不可用" description={<Space direction="vertical"><span>{modelsError}</span><Button size="small" onClick={() => setModelReloadKey(value => value + 1)}>重试读取模型</Button></Space>} /> : null}
    {!modelsLoading && !modelsError && !models.length ? <Alert type="warning" showIcon message="个人中心没有已启用模型" description={<Space direction="vertical"><span>请先在个人中心按文本、图片、视频类型新增并启用模型，再返回当前批量作品选择。</span><Button type="link" href="/api-config">前往个人中心配置模型</Button></Space>} /> : null}
    </> : null}
    {sections.includes('sensitive') ? <section className="batch-factory-engine-card"><header><b>敏感词处理</b><HelpTip title="敏感词处理">批量工厂自己的开关：关闭后只走规则脱敏；AI 修复失败会自动降级到规则替换/打码，并在卡片上等待人工确认，不再把整本小说判成失败。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>AI 修复</b><HelpTip title="AI 修复">默认关闭；只有显式开启后才会调用 AI 修复，关闭时只走规则脱敏/打码，不影响后续阶段。</HelpTip></span><Switch checked={value.sensitiveAIEnabled === true} onChange={sensitiveAIEnabled => patch({ sensitiveAIEnabled })} /></label>
      {/* 这里不能用 label 包住字段：同一 label 内的“？”帮助按钮是 label 的受控元素，
          浏览器会把点击下拉当成激活帮助按钮，导致弹层打开后立即失焦收回（闪回）。 */}
      <div className="batch-factory-engine-field"><span><b>修复模型</b><HelpTip title="修复模型">可多选，按顺序逐个尝试；留空表示继承小说获取页的全局修复模型。</HelpTip></span><Select mode="multiple" allowClear loading={modelsLoading} value={Array.isArray(value.sensitiveAIModels) ? value.sensitiveAIModels : []} options={modelOptions.text} placeholder="默认继承全局修复模型" onChange={sensitiveAIModels => patch({ sensitiveAIModels })} /></div>
      <label className="batch-factory-engine-field"><span><b>失败时策略</b><HelpTip title="失败时策略">“保留原文待确认”会打码保留正文并等待人工确认；“判为失败”保持旧的严格语义。</HelpTip></span><Segmented block value={value.sensitiveAIOnFailure === 'fail' ? 'fail' : 'review'} options={[{ value: 'review', label: '保留原文待确认' }, { value: 'fail', label: '判为失败' }]} onChange={sensitiveAIOnFailure => patch({ sensitiveAIOnFailure })} /></label>
    </div></section> : null}
    {visibility.video && sections.includes('audio') ? <section className="batch-factory-engine-card"><header><b>跟随配音</b><HelpTip title="跟随配音">开启后只测量没有有效缓存的视频原文非空行。</HelpTip></header><div className="batch-factory-engine-card-body"><Space direction="vertical" style={{ width: '100%' }}>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>分镜规划跟随配音</span><Switch checked={value.audioPlanningEnabled === true} onChange={audioPlanningEnabled => patch({ audioPlanningEnabled })} /></Space>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>合并跟随配音</span><Switch checked={value.audioMergeEnabled === true} onChange={audioMergeEnabled => patch({ audioMergeEnabled })} /></Space>
    </Space></div></section> : null}
    {sections.includes('publish') ? <section className="batch-factory-engine-card"><header><b>发布统一</b><HelpTip title="发布统一">视频管理系统的批量默认上传设置。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>网站配置档</b></span><Input value={publish.versionProfile || ''} placeholder="例如：女频短剧版" onChange={event => patch({ publishSettings: { ...publish, versionProfile: event.target.value } })} /></label>
      <label className="batch-factory-engine-field"><span><b>上传视频类型</b></span><Segmented value={publish.uploadVideoType || 'merged'} options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} onChange={uploadVideoType => patch({ publishSettings: { ...publish, uploadVideoType } })} /></label>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>改文后上传</span><Switch checked={value.publishRewriteEnabled === true} onChange={publishRewriteEnabled => patch({ publishRewriteEnabled })} /></Space>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>素材复用</span><Switch checked={publish.materialReuse === true} onChange={materialReuse => patch({ publishSettings: { ...publish, materialReuse } })} /></Space>
    </div></section> : null}
  </Space>;
}

export function BatchFactoryPublishSettingsForm({ value, onChange, active }) {
  const [checking, setChecking] = useState(false);
  const [syncing, setSyncing] = useState('');
  const [styleOpen, setStyleOpen] = useState(false);
  const [account, setAccount] = useState(null);
  const [environment, setEnvironment] = useState(null);
  const [loginOpen, setLoginOpen] = useState(false);
  const [loginUsername, setLoginUsername] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [loginBusy, setLoginBusy] = useState(false);
  const [organizations, setOrganizations] = useState([]);
  const [organizationsError, setOrganizationsError] = useState('');
  const publish = value.publishSettings || {};
  const patch = next => onChange({ ...value, ...next });
  const selectedWebsiteProfile = (publish.websiteProfiles || []).find(profile => profile.id === publish.websiteProfileId);
  const selectedProfileOrganization = websiteProfileOrganization(selectedWebsiteProfile);
  useEffect(() => {
    if (!publish.websiteProfileId) return;
    const organization = selectedProfileOrganization;
    if ((publish.organization || '') === organization) return;
    // 选中配置档后组织归属必须完全以该档为准：该档没有提供组织时清空旧值，
    // 避免上一次选择残留的 organization 被误认为“由配置档自动带入”。
    patch({ publishSettings: { ...publish, organization } });
  }, [publish.websiteProfileId, publish.websiteProfiles, selectedProfileOrganization]);
  const hasSession = result => Boolean((result?.checks || []).some(check => SESSION_CHECK_NAMES.includes(check.name) && check.ok));
  const checkDetail = (result, name) => String((result?.checks || []).find(check => check.name === name)?.detail || '').trim();
  const sessionDetail = result => String((result?.checks || []).find(check => SESSION_CHECK_NAMES.includes(check.name))?.detail || (result?.checks || [])[0]?.detail || '').trim();
  // 状态与同步共用后端同一个凭证解析结果（result.credential），前端不再自行
  // 用“有没有账号”猜状态，也不再统一显示成“登录已失效”。
  const credentialState = result => String(result?.credential?.state || '').trim();
  const credentialLabel = result => {
    const state = credentialState(result);
    if (state === 'verified') return '已登录并验证';
    if (state === 'expired') return '登录已过期';
    if (state === 'incomplete') return '凭证不全';
    if (state === 'not_logged_in') return '未登录';
    if (state === 'unavailable') return '状态暂不可用';
    return result ? (hasSession(result) ? '已登录并验证' : account?.username ? '待重新验证' : '未登录') : '等待检查';
  };
  const publishSessionReady = Boolean(environment?.session && environment?.visible?.ok);
  const credentialFailure = error => /^(TARGET_SESSION_EXPIRED|TARGET_NOT_LOGGED_IN|TARGET_CREDENTIAL_INCOMPLETE)$/.test(String(error?.code || ''));
  const organizationCatalogReady = Boolean(environment?.session);
  // 风格目录与「同步风格目录」共用 GET /web-submit/config 的 settings.style_catalog，
  // 不新增接口；同步成功后原地回写，避免用户刷新。
  const styleCatalog = Array.isArray(account?.style_catalog) && account.style_catalog.length
    ? account.style_catalog
    : (Array.isArray(publish.websiteStyleCatalog) ? publish.websiteStyleCatalog : []);
  const styleCatalogNames = styleCatalog
    .map(item => String(typeof item === 'string' ? item : item?.name || item?.id || '').trim())
    .filter(Boolean);
  const loadAccount = async () => {
    const [config, result] = await Promise.all([getWebSubmitConfig(), checkWebSubmitEnvironment()]);
    setAccount(config?.settings || config?.config || {});
    setEnvironment({ environment: result, session: hasSession(result), visible: null });
  };
  const markPublishSessionExpired = error => {
    const detail = apiErrorMessage(error, '视频管理系统登录会话已失效，请重新验证登录');
    setEnvironment(current => ({ ...current, credential: { ...(current?.credential || {}), state: 'expired', code: 'TARGET_SESSION_EXPIRED', message: detail }, error: detail, session: false, visible: { ok: false, error: detail } }));
  };
  const refreshOrganizations = async () => {
    setOrganizationsError('');
    try {
      const result = await get121OrganizationOptions();
      setOrganizations(Array.isArray(result?.organizations) ? result.organizations : []);
      return { ok: true };
    } catch (error) {
      setOrganizations([]);
      const detail = error?.message || '视频管理系统组织目录读取失败';
      setOrganizationsError(detail);
      if (/(登录会话|session[_\s-]*expired|unauthorized|\b401\b)/i.test(detail)) markPublishSessionExpired(error);
      return { ok: false, error: new Error(detail) };
    }
  };
  const openLoginForm = () => {
    setLoginUsername(String(account?.username || '').trim());
    setLoginPassword('');
    setLoginOpen(true);
  };
  const selfCheck = async () => {
    setChecking(true);
    try {
      const result = await checkWebSubmitEnvironment();
      const session = hasSession(result);
      const visible = session ? await testWebSubmitVisible() : null;
      const checked = { environment: result, session, visible };
      setEnvironment(checked);
      if (!session) {
        const detail = sessionDetail(result) || '视频管理系统登录会话无效，请重新登录';
        message.error(detail);
        openLoginForm();
        return { ...checked, error: new Error(detail) };
      }
      if (!visible?.ok) {
        const detail = visible?.error || '视频管理系统后台验证未通过';
        message.error(detail);
        openLoginForm();
        return { ...checked, error: new Error(detail) };
      }
      const organizations = await refreshOrganizations();
      if (!organizations?.ok) message.error(organizations.error?.message || '视频管理系统组织目录读取失败');
      else message.success('重新验证完成，组织目录已更新');
      return { ...checked, organizations };
    } catch (error) {
      const checked = { error: error?.message || '视频管理系统环境自检失败', session: false, visible: null };
      setEnvironment(checked);
      message.error(checked.error);
      return checked;
    } finally { setChecking(false); }
  };
  useEffect(() => {
    if (!active) return undefined;
    let alive = true;
    loadAccount().catch(error => alive && setEnvironment({ error: error?.message || '无法读取视频管理系统状态', session: false, visible: null }));
    return () => { alive = false; };
  }, [active]);
  useEffect(() => {
    if (!active || !organizationCatalogReady) {
      setOrganizations([]);
      setOrganizationsError('');
      return undefined;
    }
    let alive = true;
    refreshOrganizations();
    return () => { alive = false; };
  }, [active, organizationCatalogReady]);
  const login = async () => {
    if (!loginUsername.trim() || !loginPassword) return message.warning('请输入视频管理系统账号和密码');
    setLoginBusy(true);
    try {
      const saved = await saveWebSubmitConfig({ settings: { ...(account || {}), username: loginUsername.trim(), password: loginPassword } });
      setAccount(saved?.settings || { username: loginUsername.trim(), password_masked: true });
      setLoginPassword('');
      const checked = await selfCheck();
      if (!checked?.session || !checked.visible?.ok) throw new Error(sessionDetail(checked?.environment) || checked?.error || '视频管理系统登录后验证未通过');
      if (!checked.organizations?.ok) throw (checked.organizations?.error || new Error('视频管理系统组织目录读取失败'));
      setLoginOpen(false);
      message.success('视频管理系统已登录、验证并读取组织目录');
    } catch (error) { message.error(error?.message || '视频管理系统登录失败'); }
    finally { setLoginBusy(false); }
  };
  const syncProfiles = async () => {
    if (!publishSessionReady) return message.warning('请先登录并验证视频管理系统');
    setSyncing('profiles');
    try {
      const result = await syncWebSubmitConfigs({ persist: false });
      patch({ publishSettings: { ...publish, websiteProfiles: Array.isArray(result?.groups) ? result.groups : [] } });
    } catch (error) {
      if (credentialFailure(error)) markPublishSessionExpired(error);
      message.error(apiErrorMessage(error, '同步网站配置档失败'));
    }
    finally { setSyncing(''); }
  };
  const syncStyles = async () => {
    if (!publishSessionReady) return message.warning('请先登录并验证视频管理系统');
    setSyncing('styles');
    try {
      const result = await syncWebSubmitStyles({ persist: true });
      const styles = Array.isArray(result?.settings?.style_catalog) ? result.settings.style_catalog : (Array.isArray(result?.styles) ? result.styles : []);
      setAccount(current => ({ ...(current || {}), style_catalog: styles }));
      patch({ publishSettings: { ...publish, websiteStyleCatalog: styles } });
    } catch (error) {
      if (credentialFailure(error)) markPublishSessionExpired(error);
      message.error(apiErrorMessage(error, '同步风格目录失败'));
    }
    finally { setSyncing(''); }
  };
  return <><Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>批量默认发布规则 <HelpTip title="批量默认发布规则">这里维护视频管理系统连接、网站配置映射和上传素材规则；单书配置可覆盖默认值。</HelpTip></span>} />
    <section className="batch-factory-vms-connection-card"><div className="batch-factory-vms-main"><div className="batch-factory-vms-icon" aria-hidden="true">▣</div><div className="batch-factory-vms-copy"><div className="batch-factory-vms-title-row"><b>视频管理系统</b><span>{credentialLabel(environment?.environment)}</span></div><div className="batch-factory-vms-meta"><span>当前账号</span><strong>{account?.username || '尚未登录'}</strong></div><div className="batch-factory-vms-meta"><span>后台地址</span><a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">two.121w.com/tttadmin</a></div><div className="batch-factory-vms-meta"><span>素材相册</span><a href={N8_MATERIAL_ALBUM_URL} target="_blank" rel="noreferrer">打开 N8 素材相册</a></div></div></div><div className="batch-factory-vms-actions"><Button href={TARGET_ADMIN_URL} target="_blank">打开后台</Button><Button onClick={openLoginForm}>{account?.username ? '更换账号' : '登录并验证'}</Button><Button onClick={selfCheck} loading={checking}>{account?.username ? '重新验证' : '环境自检'}</Button></div></section>
    <div className="batch-factory-publish-two-column"><section className="batch-factory-engine-card"><header><b>发布映射</b><HelpTip title="发布映射">同步真实后台配置后，绑定到当前批量作品。</HelpTip></header><div className="batch-factory-engine-card-body"><label className="batch-factory-engine-field"><span><b>网站配置档</b></span><Select allowClear showSearch value={publish.websiteProfileId || undefined} options={(publish.websiteProfiles || []).map(profile => ({ value: profile.id, label: profile.name || profile.id }))} placeholder="同步网站配置档后选择" onChange={websiteProfileId => { const profile = (publish.websiteProfiles || []).find(item => item.id === websiteProfileId) || {}; patch({ publishSettings: { ...publish, websiteProfileId: websiteProfileId || '', versionProfile: profile.name || '', organization: websiteProfileOrganization(profile) } }); }} /></label><div className="batch-factory-publish-sync-row"><Button disabled={!publishSessionReady} onClick={syncProfiles} loading={syncing === 'profiles'}>同步网站配置档</Button><Button disabled={!publishSessionReady} onClick={syncStyles} loading={syncing === 'styles'}>同步风格目录</Button></div><div className="batch-factory-style-catalog">{styleCatalogNames.length ? <button type="button" className="batch-factory-style-catalog-toggle" aria-expanded={styleOpen} onClick={() => setStyleOpen(open => !open)}><span className="batch-factory-style-catalog-label">风格目录</span><span className="batch-factory-style-catalog-count">{styleCatalogNames.length} 条</span><span className={`batch-factory-style-catalog-chevron${styleOpen ? ' is-open' : ''}`} aria-hidden="true" /></button> : <div className="batch-factory-style-catalog-toggle is-empty"><span className="batch-factory-style-catalog-label">风格目录</span><span className="batch-factory-style-catalog-empty">尚未同步风格目录</span></div>}{styleOpen && styleCatalogNames.length ? <div className="batch-factory-style-catalog-list">{styleCatalogNames.map((name, index) => <span className="batch-factory-style-catalog-name" key={`${name}-${index}`}>{name}</span>)}</div> : null}</div><div className="batch-factory-publish-ai-note"><b>组织归属</b><span>{selectedProfileOrganization ? `由网站配置档自动带入：${selectedProfileOrganization}` : publish.websiteProfileId ? '当前网站配置档未提供组织归属，提交时需在确认单选择。' : '尚未选择网站配置档。'}</span></div></div></section>
      <section className="batch-factory-engine-card"><header><b>批量默认上传规则</b><HelpTip title="批量默认上传规则">这些规则可被单书发布设置覆盖。</HelpTip></header><div className="batch-factory-engine-card-body"><label className="batch-factory-engine-field"><span><b>上传视频类型</b></span><Segmented value={publish.uploadVideoType || 'merged'} options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} onChange={uploadVideoType => patch({ publishSettings: { ...publish, uploadVideoType } })} /></label><div className="batch-factory-publish-rule-list"><section><div><b>素材复用</b></div><Switch checked={publish.materialReuse === true} onChange={materialReuse => patch({ publishSettings: { ...publish, materialReuse } })} /></section><section><div><b>水平翻转</b></div><Switch checked={publish.horizontalFlip === true} onChange={horizontalFlip => patch({ publishSettings: { ...publish, horizontalFlip } })} /></section><section><div><b>改文后上传</b></div><Switch checked={value.publishRewriteEnabled === true} onChange={publishRewriteEnabled => patch({ publishRewriteEnabled })} /></section></div></div></section></div>
    <section className="batch-factory-engine-card"><header><b>连接与环境状态</b><HelpTip title="连接与环境状态">登录、后台页面和直接接口均需可用。</HelpTip></header><div className="batch-factory-engine-card-body"><div className="batch-factory-env-status-row"><div><b>登录会话</b><span>{environment?.error ? '检查失败' : credentialLabel(environment?.environment)}</span><small>{sessionDetail(environment?.environment)}</small></div><div><b>后台页面</b><span>{environment?.visible ? (environment.visible.ok ? '可访问' : '验证未通过') : '等待检查'}</span></div><div><b>直接接口</b><span>{checkDetail(environment?.environment, '121 直接接口') || '等待检查'}</span></div><Button onClick={selfCheck} loading={checking}>重新检查</Button></div>{environment?.error ? <Alert type="error" showIcon message="视频管理系统环境自检失败" description={environment.error} /> : null}</div></section>
  </Space>
  <Modal title={account?.username ? '更换视频管理系统账号' : '登录视频管理系统'} open={loginOpen} onCancel={() => { setLoginPassword(''); setLoginOpen(false); }} onOk={login} confirmLoading={loginBusy} okText="登录并验证" destroyOnClose><Space direction="vertical" size={12} style={{ width: '100%' }}><Input value={loginUsername} autoComplete="username" placeholder="视频管理系统账号" onChange={event => setLoginUsername(event.target.value)} /><Input.Password value={loginPassword} autoComplete="current-password" placeholder="视频管理系统密码" onChange={event => setLoginPassword(event.target.value)} /><a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">打开视频管理系统后台</a></Space></Modal>
  </>;
}

export function BatchFactoryComicPublishSettingsForm({ value, onChange, singleBook = false, books = [] }) {
  const comicPublishSettings = value.comicPublishSettings || {};
	const cachedDirectory = readComicPublishDirectoryCache();
	const [producerAccounts, setProducerAccounts] = useState(cachedDirectory?.producerAccounts || []);
	const [organizations, setOrganizations] = useState(cachedDirectory?.organizations || []);
	const [audioCategories, setAudioCategories] = useState(cachedDirectory?.audioCategories?.length ? cachedDirectory.audioCategories : DEFAULT_AUDIO_CATEGORIES);
	const [producerAccountOrganizations, setProducerAccountOrganizations] = useState(cachedDirectory?.producerAccountOrganizations || {});
	const [audios, setAudios] = useState([]);
	const [audiosError, setAudiosError] = useState('');
	const [audiosReloadKey, setAudiosReloadKey] = useState(0);
	const [producerAccountsLoading, setProducerAccountsLoading] = useState(!cachedDirectory);
	const [producerAccountsWarning, setProducerAccountsWarning] = useState('');
	const [organizationsWarning, setOrganizationsWarning] = useState('');
	const [audioCategoriesWarning, setAudioCategoriesWarning] = useState('');
	const [directoryFailureKind, setDirectoryFailureKind] = useState('');
	const [directoryFailureMessage, setDirectoryFailureMessage] = useState('');
	const [directoryAvailable, setDirectoryAvailable] = useState(null);
	const [publishKeyConfigured, setPublishKeyConfigured] = useState(null);
	useEffect(() => {
		let active = true;
		if (!cachedDirectory) setProducerAccountsLoading(true);
		getN8ComicPublishStatus({ silent: true, suppressGlobalError: true }).then(response => {
			if (!active) return;
			setProducerAccounts(Array.isArray(response?.producerAccounts) ? response.producerAccounts : []);
			setOrganizations(Array.isArray(response?.organizations) ? response.organizations : []);
			const categories = Array.isArray(response?.audioCategories) ? response.audioCategories : [];
			if (categories.length) setAudioCategories(categories);
			setProducerAccountOrganizations(response?.producerAccountOrganizations && typeof response.producerAccountOrganizations === 'object' ? response.producerAccountOrganizations : {});
			const latestDirectory = readComicPublishDirectoryCache();
			writeComicPublishDirectoryCache({ ...response, audioCategories: categories.length ? categories : (latestDirectory?.audioCategories || cachedDirectory?.audioCategories || []) });
			setProducerAccountsWarning(String(response?.producerAccountsWarning || ''));
			setOrganizationsWarning(String(response?.organizationsWarning || ''));
			setAudioCategoriesWarning(String(response?.audioCategoriesWarning || ''));
			// 目录是否可读用 directoryAvailable，发布密钥用 publishKeyConfigured。
			// 二者不是同一个开关，旧字段 configured 不再参与目录告警判定。
			setDirectoryAvailable(typeof response?.directoryAvailable === 'boolean' ? response.directoryAvailable : null);
			setPublishKeyConfigured(typeof response?.publishKeyConfigured === 'boolean' ? response.publishKeyConfigured : null);
			setDirectoryFailureKind('');
			setDirectoryFailureMessage('');
		}).catch(error => {
			if (!active) return;
			const failure = describeN8DirectoryFailure(error);
			setProducerAccounts([]);
			setProducerAccountsWarning(failure.message);
			// 读取失败时不知道服务端是否配置，不能擅自断言“未配置/令牌失效”。
			setDirectoryAvailable(null);
			setDirectoryFailureKind(failure.kind);
			setDirectoryFailureMessage(failure.message);
		}).finally(() => { if (active) setProducerAccountsLoading(false); });
		getN8ComicPublishAudioCategories({ silent: true, suppressGlobalError: true }).then(response => {
			if (!active) return;
			const categories = Array.isArray(response?.audioCategories) ? response.audioCategories : [];
			if (!categories.length) return;
			setAudioCategories(categories);
			const latestDirectory = readComicPublishDirectoryCache();
			writeComicPublishDirectoryCache({ ...(latestDirectory || cachedDirectory || {}), audioCategories: categories });
		}).catch(() => {});
		return () => { active = false; };
	}, []);
  const patchComicPublishSettings = next => onChange({
    ...value,
    comicPublishSettings: { ...comicPublishSettings, ...next }
  });
  const inferredAudioCategory = useMemo(() => inferBatchN8AudioCategory(books, audioCategories), [books, audioCategories]);
  useEffect(() => {
    if (comicPublishSettings.audioCategorySelectionMode === 'manual') return;
    const selected = inferredAudioCategory.category;
    if (selected) {
      if (Number(comicPublishSettings.audioCategoryId) !== Number(selected.id) || comicPublishSettings.audioCategoryName !== selected.name || comicPublishSettings.audioCategorySelectionMode !== 'auto') {
        patchComicPublishSettings({ audioCategoryId: selected.id, audioCategoryName: selected.name, audioCategorySelectionMode: 'auto', specifiedAudioId: '', specifiedAudioName: '' });
      }
      return;
    }
    if (comicPublishSettings.audioCategorySelectionMode === 'auto' && comicPublishSettings.audioCategoryId) {
      patchComicPublishSettings({ audioCategoryId: null, audioCategoryName: '', audioCategorySelectionMode: '', specifiedAudioId: '', specifiedAudioName: '' });
    }
  }, [inferredAudioCategory.category?.id, inferredAudioCategory.category?.name, inferredAudioCategory.reason, comicPublishSettings.audioCategorySelectionMode, books]);
  useEffect(() => {
    const organizationId = String(comicPublishSettings.organizationId || '').trim();
    const audioCategoryId = String(comicPublishSettings.audioCategoryId || '').trim();
    if (!organizationId || !audioCategoryId) { setAudios([]); setAudiosError(''); return undefined; }
    let active = true;
    setAudiosError('');
    getN8ComicPublishAudios({ organizationId, audioCategoryId }, { silent: true, suppressGlobalError: true }).then(result => {
      if (!active) return;
      const candidates = Array.isArray(result?.audios) ? result.audios : [];
      setAudios(candidates);
      setAudiosError('');
      const selected = chooseN8Audio(candidates, comicPublishSettings.specifiedAudioId);
      patchComicPublishSettings(selected ? { specifiedAudioId: selected.id, specifiedAudioName: selected.name } : { specifiedAudioId: '', specifiedAudioName: '' });
    }).catch(error => { if (active) { const failure = describeN8DirectoryFailure(error); setAudios([]); setAudiosError(failure.message); patchComicPublishSettings({ specifiedAudioId: '', specifiedAudioName: '' }); } });
    return () => { active = false; };
  }, [comicPublishSettings.organizationId, comicPublishSettings.audioCategoryId, audiosReloadKey]);
  // 目录读取成功但上游返回业务错误时，三路 warning 里保存的就是上游 code/message。
  const directoryWarning = organizationsWarning || audioCategoriesWarning || producerAccountsWarning;
  return <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert
      type="info"
      showIcon
      message={singleBook ? '当前书漫画发布配置' : '漫画发布默认设置'}
      description={singleBook ? '当前书未单独修改的项目继续继承批次默认值；这里只保存本书的差异。图片顺序、宽高、格式、签名和上传重试由系统自动处理。' : '这里只设置一次，当前批次的每本漫画都会自动继承。图片顺序、宽高、格式、签名和上传重试由系统自动处理。'}
    />
    {directoryFailureKind === 'auth'
      ? <Alert type="error" showIcon message="登录已失效，请重新登录" />
      : null}
    {directoryFailureKind !== 'auth' && (directoryFailureMessage || directoryWarning)
      ? <Alert type="error" showIcon message={directoryFailureMessage || directoryWarning} />
      : null}
    {directoryFailureKind !== 'auth' && !(directoryFailureMessage || directoryWarning) && directoryAvailable === false
      ? <Alert type="error" showIcon message="N8 目录服务未配置（服务器缺少 N8 后台登录令牌），请联系管理员配置后重新打开本弹窗。" />
      : null}
    {publishKeyConfigured === false
      ? <Alert type="info" showIcon message="发布密钥未配置：目录可以读取，但发布/上传不可用。请配置服务器环境变量 QIANTIE_N8_COMIC_OPEN_API_KEY 后重启服务。" />
      : null}
    <section className="batch-factory-engine-card">
      <header><b>素材字段映射</b><HelpTip title="素材字段映射">这里和 N8 素材库的字段保持一致；单书配置未覆盖时，自动继承发布统一设置。</HelpTip></header>
      <div className="batch-factory-engine-card-body batch-factory-engine-grid">
        <label className="batch-factory-engine-field">
          <span><b>文案来源</b></span>
          <Segmented block value={comicPublishSettings.copySource || 'original'} options={[{ value: 'original', label: '原创' }, { value: 'bookstore', label: '书城' }, { value: 'third_party', label: '第三方' }]} onChange={copySource => patchComicPublishSettings({ copySource })} />
        </label>
        <label className="batch-factory-engine-field">
          <span><b>素材来源</b></span>
          <Select value={comicPublishSettings.materialSource || 'ORIGINAL'} options={[{ value: 'ORIGINAL', label: '原创制作' }, { value: 'ORIGINAL_TEXT', label: '原文制作' }, { value: 'DERIVATIZATION', label: '衍生制作' }, { value: 'AI_PRE', label: 'AI前帖' }, { value: 'AUTOMATED_CLIPPING', label: '自动化剪辑' }, { value: 'AUTOMATED_CLIPPING_A', label: '自动化剪辑AI文案' }, { value: 'AUTOMATED_CLIPPING_AI_PRE', label: '自动化剪辑AI前贴' }, { value: 'MANGA_VIDEO_AUTOMATED_CLIPPING', label: '漫剧自动化剪辑' }, { value: 'THIRD_PARTY', label: '第三方' }, { value: 'THIRD_PARTY_REMOVE_WATERMARKS', label: '第三方去水印' }, { value: 'TOOL', label: '素材工具' }, { value: 'GROUP_AI', label: '自动化AI' }, { value: 'AUTOMATED_CLIPPING_BLAZING_FAST', label: '自动化3分钟' }, { value: 'AUTOMATED_CLIPPING_MANGA', label: '自动化漫改' }]} onChange={materialSource => patchComicPublishSettings({ materialSource })} />
        </label>
        <div className="batch-factory-engine-field">
          <span><b>制作者管理员账号（创意者）</b><HelpTip title="制作者管理员账号（创意者）">这就是 N8 素材库里的创意者，不再单独保存第二个账号。</HelpTip></span>
          <Select showSearch optionFilterProp="label" placement="bottomLeft" loading={producerAccountsLoading} value={comicPublishSettings.producerAdminName || undefined} options={producerAccountOptions(producerAccounts, comicPublishSettings.producerAdminName, comicPublishSettings.organizationId, producerAccountOrganizations)} placeholder={producerAccountsLoading ? '正在同步账号目录' : '请选择制作者管理员账号'} disabled={!producerAccountsLoading && producerAccountOptions(producerAccounts, comicPublishSettings.producerAdminName, comicPublishSettings.organizationId, producerAccountOrganizations).length === 0} getPopupContainer={() => document.body} onChange={producerAdminName => patchComicPublishSettings({ producerAdminName, creativeAdminName: producerAdminName })} />
          {!producerAccountsLoading && !producerAccountsWarning && producerAccountOptions(producerAccounts, comicPublishSettings.producerAdminName, comicPublishSettings.organizationId, producerAccountOrganizations).length === 0 ? <small>当前组织没有可用的制作者账号，请先选择正确的组织。</small> : null}
          {producerAccountsWarning ? <small>{producerAccountsWarning}</small> : null}
        </div>
        <div className="batch-factory-engine-field">
          <span><b>组织</b></span>
          <Select placement="bottomLeft" value={String(comicPublishSettings.organizationId || '') || undefined} options={organizations.map(item => ({ value: String(item.id), label: `${item.name}（组织 ${item.id}）` }))} placeholder="请选择 N8 组织" getPopupContainer={() => document.body} onChange={organizationId => { const selected = organizations.find(item => String(item.id) === String(organizationId)); patchComicPublishSettings({ organizationId: String(organizationId || ''), organizationName: selected?.name || '', specifiedAudioId: '', specifiedAudioName: '' }); }} />
          {organizationsWarning ? <small>{organizationsWarning}</small> : null}
        </div>
      </div>
    </section>
    <section className="batch-factory-engine-card">
      <header><b>发布账号与作品匹配</b><HelpTip title="漫画发布统一">账号必须与漫画后台中已启用的管理员账号完全一致。找不到同名作品时，发布前会让你选择一次。</HelpTip></header>
      <div className="batch-factory-engine-card-body batch-factory-engine-grid">
		<div className="batch-factory-engine-field">
          <span><b>启用中的音频分类</b><HelpTip title="启用中的音频分类">请照着漫画后台“音频分类”页面中启用状态的分类名称填写，发布前会同时显示名称和 ID 供你核对。</HelpTip></span>
          <Select
            value={comicPublishSettings.audioCategoryName || ''}
            options={audioCategories.map(item => ({ value: item.name, label: `${item.name}（ID ${item.id}）`, id: item.id }))}
            placeholder="请选择 N8 启用中的音频分类"
            showSearch
            optionFilterProp="label"
            placement="bottomLeft"
            getPopupContainer={() => document.body}
            onChange={audioCategoryName => {
              const selected = audioCategories.find(item => item.name === audioCategoryName);
              patchComicPublishSettings({ audioCategoryName, audioCategoryId: selected?.id || null, audioCategorySelectionMode: 'manual', specifiedAudioId: '', specifiedAudioName: '' });
            }}
          />
          {audioCategoriesWarning ? <small>{audioCategoriesWarning}</small> : null}
		</div>
		<div className="batch-factory-engine-field">
          <span><b>音频分类 ID（必填）</b><HelpTip title="音频分类 ID">选择分类名称后自动带出对应 ID，不再手填数字。</HelpTip></span>
          <Select
            showSearch
            optionFilterProp="label"
            value={String(comicPublishSettings.audioCategoryId || '') || undefined}
            options={audioCategories.map(item => ({ value: String(item.id), label: `${item.name}（ID ${item.id}）` }))}
            placeholder="请选择 N8 启用中的音频分类"
            placement="bottomLeft"
            getPopupContainer={() => document.body}
            onChange={audioCategoryId => {
              const selected = audioCategories.find(item => String(item.id) === String(audioCategoryId));
              patchComicPublishSettings({ audioCategoryId: selected?.id || null, audioCategoryName: selected?.name || '', audioCategorySelectionMode: 'manual', specifiedAudioId: '', specifiedAudioName: '' });
            }}
          />
		</div>
        <div className="batch-factory-engine-field"><span><b>指定音频</b></span><Select showSearch optionFilterProp="label" value={String(comicPublishSettings.specifiedAudioId || '') || undefined} options={audios.map(item => ({ value: String(item.id), label: `${item.name}（ID ${item.id}）` }))} placeholder="请选择指定音频（可选）" onChange={specifiedAudioId => { const selected = audios.find(item => String(item.id) === String(specifiedAudioId)); patchComicPublishSettings({ specifiedAudioId: selected?.id || '', specifiedAudioName: selected?.name || '' }); }} />
          {audiosError ? <small>{audiosError}<Button type="link" size="small" onClick={() => setAudiosReloadKey(key => key + 1)}>重试</Button></small> : null}
          {!audiosError && !audios.length && comicPublishSettings.organizationId && comicPublishSettings.audioCategoryId ? <small>当前分类和组织下没有可用指定音频</small> : null}
          <small>{comicPublishSettings.audioCategorySelectionMode === 'auto' ? '已根据批次 AI 分类自动匹配，可更改' : inferredAudioCategory.reason === 'mixed' ? '批次书籍分类不一致，需手动选择音频分类' : inferredAudioCategory.reason === 'unclassified' ? '书籍 AI 分类未完成，需手动选择音频分类' : '已自动匹配，可更改'}</small></div>
        <label className="batch-factory-engine-field" style={{ gridColumn: '1 / -1' }}>
          <span><b>作品匹配方式</b></span>
          <Segmented
            block
            value={comicPublishSettings.workMatchMode || 'title_auto'}
            options={[
              { value: 'title_auto', label: '按书名自动匹配' },
              { value: 'manual', label: '发布时手动选择' }
            ]}
            onChange={workMatchMode => patchComicPublishSettings({ workMatchMode })}
          />
        </label>
      </div>
    </section>
    <Alert type="success" showIcon message={singleBook ? '当前书发布时，只需要确认漫画合集和匹配到的作品。' : '以后发布时，你只需要确认漫画合集和匹配到的作品。'} />
  </Space>;
}

export function BatchFactoryAiReasoningForm({ value, onChange, batchId, productionMode = 'video' }) {
  const [catalog, setCatalog] = useState([]);
  const visibility = batchFactoryProductionModeVisibility(productionMode);
  const config = value && typeof value === 'object' ? value : {};
  const assets = config.assets || {};
  const constraints = config.constraints || {};
  const derivedOpening = config.derivedOpening || {};
  const video = config.video?.presetId && config.video.presetId !== 'batch-video-meta'
    ? config.video
    : { presetId: 'batch-video-sd', presetName: 'SD 视频提示词', presetSlot: 'batch.video-meta', presetVersion: null, constraintCategory: '' };
  const comic = config.comic || {};
  useEffect(() => {
    let active = true;
    Promise.all([listSystemPresetCatalog('script'), listSystemPresetCatalog('batch-factory')]).then(results => {
      if (active) setCatalog(results.flatMap(result => Array.isArray(result?.catalog) ? result.catalog : []));
    }).catch(() => active && setCatalog([]));
    return () => { active = false; };
  }, []);
  const noPromptOption = { value: '__none__', label: '无' };
  const normalizePromptId = presetId => presetId === '__none__' ? '' : (presetId || '');
  const promptOptions = options => [noPromptOption, ...options];
  const catalogOptionsBySlot = slot => catalog.filter(item => item.slot === slot).map(selectOption);
  const bySlot = slot => promptOptions(catalogOptionsBySlot(slot));
  // CM 资产提取提示词属于 batch.character-meta，但视频资产阶段也要能选择它。
  // 这里只接入完整资产提取预设，不把人物单独编译预设误放进整套资产选择器。
  const videoAssetOptions = () => promptOptions(catalog.filter(item => item.slot === 'script.asset-extraction' || item.id === 'batch-character-meta').map(selectOption));
  const presetFor = presetId => presetValue(catalog.find(item => item.id === presetId));
  const bySlots = slots => promptOptions(catalog.filter(item => slots.includes(item.slot)).map(selectOption));
  const selectedConstraint = category => (constraints.selections || []).find(item => item.constraintCategory === category)?.presetId || constraints[category]?.presetId || '';
  const selectedComicConstraint = category => (comic.constraints || []).find(item => item.constraintCategory === category)?.presetId || '';
  const byConstraint = categories => {
    const values = Array.isArray(categories) ? categories : [categories];
    return promptOptions(catalog.filter(item => item.kind === 'addon' && values.some(category => item.constraintCategory === category || (category === 'prefix' && item.id === SMART_UNIFIED_PRESET_ID))).map(selectOption));
  };
  const comicEnabled = nextComic => Boolean(nextComic.assets?.presetId || nextComic.presetId || (nextComic.constraints || []).some(item => item.presetId));
  const updateComic = next => {
    const nextComic = { ...comic, ...next };
    onChange({ ...config, comic: { ...nextComic, enabled: comicEnabled(nextComic) } });
  };
  const updateConstraint = (category, comicCategory, presetId) => {
    const normalizedId = normalizePromptId(presetId);
    const normalRest = (constraints.selections || []).filter(item => item.constraintCategory !== category);
    const comicRest = (comic.constraints || []).filter(item => item.constraintCategory !== comicCategory);
    const selected = catalog.find(item => item.id === normalizedId);
    const isComic = selected?.constraintCategory === comicCategory;
    const apply = body => {
      const nextComic = {
        ...comic,
        constraints: isComic && normalizedId
          ? [...comicRest, { ...presetValue(selected), constraintCategory: comicCategory, body }]
          : comicRest
      };
      onChange({
        ...config,
        constraints: {
          ...constraints,
          enabled: true,
          enabledCategories: [...new Set([...(constraints.enabledCategories || []), category])],
          selections: !normalizedId || isComic ? normalRest : [...normalRest, { ...presetValue(selected), body }]
        },
        comic: { ...nextComic, enabled: comicEnabled(nextComic) }
      });
    };
    if (!normalizedId) return apply('');
    // 预设目录为了提速不带正文，这里按预设号取一次权威正文一并保存，
    // 否则导演/视频生产拿不到约束内容（历史版本正是漏了这一步）。
    getConstraintPresetTexts([normalizedId])
      .then(result => apply(String(result?.texts?.[normalizedId] || '')))
      .catch(() => message.error('系统预设提示词读取失败，请重试'));
  };
  return <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>AI 推理 <HelpTip title="AI 推理">这里选的是「提示词规则模板」：决定用哪一套已发布的规则模板、以及模板正文（画面前缀、画质、限制、负面词等），不是每本书已经生成出来的分镜和提示词。各书单独改过的规则仍可在单书配置里覆盖。</HelpTip></span>} />
    <section className="batch-factory-engine-card"><header><b>资产设置</b><HelpTip title="资产设置">人物、场景和道具统一使用提取提示词；选择“无”表示不使用提示词。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      {visibility.comic ? <label className="batch-factory-engine-field"><span><b>漫画资产提示词</b></span><Select allowClear value={comic.assets?.presetId || undefined} options={bySlot('batch.comic-assets')} placeholder="选择漫画资产提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); const nextComic = { ...comic, assets: presetFor(normalizedId) }; onChange({ ...config, comic: { ...nextComic, enabled: comicEnabled(nextComic) } }); }} /></label> : <label className="batch-factory-engine-field"><span><b>人物场景道具提取提示词</b></span><Select allowClear value={assets.extraction?.presetId || undefined} options={videoAssetOptions()} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, assets: { ...assets, enabled: Boolean(normalizedId), extraction: presetFor(normalizedId) } }); }} /></label>}
      {visibility.video ? <label className="batch-factory-engine-field"><span className="batch-factory-option-title"><b>换开头</b><Tooltip title="开启后，2 个及以上分镜的书会为分镜一生成多个开场变体并分别合成上传；数量包含原始开头。"><span className="batch-factory-help-button" aria-label="换开头说明">?</span></Tooltip></span><Space><Switch checked={config.openingEnabled === true} onChange={enabled => onChange({ ...config, openingEnabled: enabled, openingCount: config.openingCount ?? 4 })} /><InputNumber min={1} max={8} precision={0} disabled={config.openingEnabled !== true} value={config.openingCount ?? 4} onChange={count => onChange({ ...config, openingCount: count ?? 4 })} addonAfter="条（含原始）" /></Space></label> : null}
    </div></section>
    <section className="batch-factory-engine-card"><header><b>约束设置</b><HelpTip title="约束设置">画面前缀、画质、画面限制和负面提示词在这里统一选择；不选的类别不会注入，可在单书配置里单独覆盖。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      {visibility.video ? <label className="batch-factory-engine-field"><span><b>画面前缀词</b></span><Select allowClear value={selectedConstraint('prefix')} options={byConstraint('prefix')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('prefix', null, presetId)} /></label> : null}
      <label className="batch-factory-engine-field"><span><b>画质约束</b></span><Select allowClear value={visibility.comic ? selectedComicConstraint('comic_quality') : selectedConstraint('quality')} options={byConstraint(visibility.comic ? 'comic_quality' : 'quality')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('quality', visibility.comic ? 'comic_quality' : null, presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>画面限制</b></span><Select allowClear value={visibility.comic ? selectedComicConstraint('comic_restriction') : selectedConstraint('restriction')} options={byConstraint(visibility.comic ? 'comic_restriction' : 'restriction')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('restriction', visibility.comic ? 'comic_restriction' : null, presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>负面提示词</b></span><Select allowClear value={visibility.comic ? selectedComicConstraint('comic_negative') : selectedConstraint('negative')} options={byConstraint(visibility.comic ? 'comic_negative' : 'negative')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('negative', visibility.comic ? 'comic_negative' : null, presetId)} /></label>
      {visibility.video ? <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>基础设定</span><Switch checked={constraints.baseSetup?.enabled !== false} onChange={enabled => onChange({ ...config, constraints: { ...constraints, baseSetup: { enabled } } })} /></Space> : null}
    </div>{visibility.comic ? <BatchFactoryComicStylePicker batchId={batchId} catalog={catalog} value={comic} onChange={nextComic => onChange({ ...config, comic: nextComic })} /> : null}</section>
    {visibility.comic ? <section className="batch-factory-engine-card"><header><b>漫画生成</b><HelpTip title="漫画生成">漫画提示词决定整页条漫的结构，图片描述词推导用于生成每页可直接出图的描述。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>漫画提示词</b></span><Select value={comic.presetId || undefined} options={catalogOptionsBySlot('batch.comic-page')} placeholder="请选择已发布漫画提示词" onChange={presetId => updateComic(presetFor(normalizePromptId(presetId)))} /></label>
      <label className="batch-factory-engine-field"><span><b>图片描述词推导</b></span><Select allowClear value={comic.imageDerivation?.presetId || undefined} options={catalogOptionsBySlot('batch.comic-image-derivation')} placeholder="未选择" onChange={presetId => updateComic({ imageDerivation: presetFor(normalizePromptId(presetId)) })} /></label>
    </div></section> : null}
    {visibility.video ? <section className="batch-factory-engine-card"><header><b>视频设置</b><HelpTip title="视频设置">视频提示词决定结构化导演分镜的输出规则与最终模板。</HelpTip></header><div className="batch-factory-engine-card-body">
      <label className="batch-factory-engine-field"><span><b>衍生开篇</b></span><Select allowClear value={derivedOpening.presetId || undefined} options={bySlots(['batch.hook-adaptation', 'batch.original-director', 'batch.viral-director'])} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, derivedOpening: { ...derivedOpening, ...presetValue(catalog.find(item => item.id === normalizedId)), enabled: Boolean(normalizedId) } }); }} /></label>
      <div className="batch-factory-engine-field"><span><b>视频提示词</b><HelpTip title="视频提示词">决定导演分镜和最终 VIDEO Prompt 的规则与格式；与上面的具体视频模型相互独立。</HelpTip></span><Select allowClear value={video.presetId || undefined} options={bySlot('batch.video-meta')} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, video: { ...video, ...presetValue(catalog.find(item => item.id === normalizedId)), enabled: Boolean(normalizedId) } }); }} /></div>
    </div></section> : null}
  </Space>;
}

export function BatchFactoryAutomationPresetManager({ value, onLoad }) {
  const [presetOpen, setPresetOpen] = useState(false);
  const [presets, setPresets] = useState([]);
  const [selectedPresetId, setSelectedPresetId] = useState('');
  const [presetName, setPresetName] = useState('');
  const [presetBusy, setPresetBusy] = useState(false);
  const selectedPreset = presets.find(item => item.id === selectedPresetId);
  const loadPresets = async () => {
    const result = await listAutomationPresets();
    const values = Array.isArray(result?.presets) ? result.presets : [];
    setPresets(values);
    return values;
  };
  const openPresetManager = async () => {
    try { await loadPresets(); setPresetOpen(true); }
    catch (error) { message.error(error?.message || '读取自动化预设失败'); }
  };
  const clonePresetConfig = clone(selectedPreset?.config);
  const loadPreset = () => {
    if (!selectedPreset) return message.warning('请先选择自动化预设');
    Modal.confirm({ title: '确认载入此预设', content: '它只替换当前统一配置草稿；确认当前批量配置后才会应用。', okText: '载入草稿', onOk: () => onLoad(clonePresetConfig) });
  };
  const savePreset = async () => {
    const name = presetName.trim();
    if (!name) return message.warning('请填写预设名称');
    setPresetBusy(true);
    try { const result = await createAutomationPreset({ name, config: value }); const saved = result?.preset || result; await loadPresets(); setSelectedPresetId(saved?.id || ''); setPresetName(''); message.success('已保存自动化预设。'); }
    catch (error) { message.error(error?.message || '保存自动化预设失败'); }
    finally { setPresetBusy(false); }
  };
  const renamePreset = async () => {
    if (!selectedPreset) return message.warning('请先选择自动化预设');
    const name = presetName.trim();
    if (!name) return message.warning('请填写新的预设名称');
    setPresetBusy(true);
    try { await updateAutomationPreset(selectedPreset.id, { name, config: selectedPreset.config, expectedVersion: selectedPreset.version }); await loadPresets(); setPresetName(''); message.success('自动化预设已重命名。'); }
    catch (error) { message.error(error?.message || '重命名自动化预设失败'); }
    finally { setPresetBusy(false); }
  };
  const removePreset = () => {
    if (!selectedPreset) return message.warning('请先选择自动化预设');
    Modal.confirm({ title: '删除所选预设', content: `删除“${selectedPreset.name}”不会影响已经排期或执行中的任务。`, okText: '删除', okButtonProps: { danger: true }, onOk: async () => { await deleteAutomationPreset(selectedPreset.id); setSelectedPresetId(''); await loadPresets(); } });
  };
  return <><Tooltip title="自动化预设"><Button type="text" icon={<SettingOutlined />} aria-label="自动化预设" onClick={openPresetManager} /></Tooltip>
  <Modal className="batch-factory-automation-preset-modal" title="自动化预设" open={presetOpen} onCancel={() => setPresetOpen(false)} footer={null} width={620} destroyOnClose><Space direction="vertical" size={14} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>预设只保存统一配置 <HelpTip title="预设只保存统一配置">不保存正文、资产、导演结果或视频结果；载入后仍须确认当前统一配置才会应用到批量。</HelpTip></span>} />
    <label className="batch-factory-engine-field"><span><b>已保存预设</b></span><Select allowClear value={selectedPresetId || undefined} onChange={setSelectedPresetId} placeholder="选择自动化预设" options={presets.map(item => ({ value: item.id, label: `${item.name} · v${item.version}` }))} /></label>
    <Space wrap><Button disabled={!selectedPreset} onClick={loadPreset}>载入预设</Button></Space>
    <Divider />
    <label className="batch-factory-engine-field"><span><b>{selectedPreset ? '预设新名称' : '新预设名称'}</b></span><Input value={presetName} maxLength={255} placeholder={selectedPreset?.name || '例如：女频 H3 全自动'} onChange={event => setPresetName(event.target.value)} /></label>
    <Space wrap><Button type="primary" loading={presetBusy} onClick={savePreset}>保存为新预设</Button><Button disabled={!selectedPreset} loading={presetBusy} onClick={renamePreset}>重命名所选预设</Button><Button danger disabled={!selectedPreset} onClick={removePreset}>删除所选预设</Button></Space>
  </Space></Modal></>;
}

export function BatchFactoryUnifiedSettingsModal({ open, batch, onClose, onSaved, onProductionModeChange }) {
  const [draftPatch, setDraftPatch] = useState({});
  const [saving, setSaving] = useState(false);
  const [modeSaving, setModeSaving] = useState(false);
  const [modeNotice, setModeNotice] = useState('');
  useEffect(() => {
    if (open) setDraftPatch(normalizeLegacyPatch(batch?.settingsState?.patch));
  }, [open, batch?.id]);
  const save = async () => {
    setSaving(true);
    const aiPromptConfig = draftPatch.aiPromptConfig && typeof draftPatch.aiPromptConfig === 'object' ? draftPatch.aiPromptConfig : {};
    const video = requiredVideoPrompt(aiPromptConfig.video);
    const nextPatch = {
      ...draftPatch,
      aiPromptConfig: { ...aiPromptConfig, video },
      ...(video.presetId === 'batch-video-sd' && !draftPatch.storyTimelineVersion
        ? { storyTimelineVersion: 'sd.original-story.v1' }
        : {})
    };
    try { const result = await onSaved(nextPatch); if (result !== false) onClose(); }
    finally { setSaving(false); }
  };
  const changeProductionMode = async productionMode => {
    const previousMode = batchFactoryProductionModeVisibility(draftPatch.productionMode).mode;
    const nextLabel = productionMode === 'comic' ? '漫画生成' : '视频生成';
    if (!onProductionModeChange || productionMode === previousMode) {
      setDraftPatch(current => ({ ...current, productionMode }));
      return;
    }
    const batchBooks = Array.isArray(batch?.books) ? batch.books : [];
    const overrideCount = batchBooks.filter(book => Object.hasOwn(book?.settingsState?.patch || {}, 'productionMode')).length;
    if (batchBooks.length) {
      const confirmed = await new Promise(resolve => {
        Modal.confirm({
          title: `确认把整批切换为${nextLabel}？`,
          content: `整批 ${batchBooks.length} 本将切换为${nextLabel}；${overrideCount ? `${overrideCount} 本单书的模式覆盖将被清除；` : ''}在途${previousMode === 'video' ? '视频' : '漫画'}任务不会迁移。`,
          okText: '确认切换',
          cancelText: '保持当前模式',
          okButtonProps: { danger: true },
          onOk: () => resolve(true),
          onCancel: () => resolve(false)
        });
      });
      if (!confirmed) return;
    }
    setDraftPatch(current => ({ ...current, productionMode }));
    setModeNotice('');
    setModeSaving(true);
    try {
      const result = await onProductionModeChange?.(productionMode);
      const failedBookTitles = Array.isArray(result?.failedBookTitles) ? result.failedBookTitles : [];
      if (failedBookTitles.length) {
        // 半成功：保留可重试的明确状态，并列出未清除覆盖的具体书名。
        setModeNotice(`生产模式已保存，但 ${failedBookTitles.length} 本小说的旧模式覆盖未清除：${failedBookTitles.join('、')}。请重试切换。`);
      } else {
        message.success(`已切换生产模式（${batchBooks.length ? `整批 ${batchBooks.length} 本` : '当前批次'}）`);
      }
    } catch (error) {
      setDraftPatch(current => ({ ...current, productionMode: previousMode }));
      setModeNotice('');
      message.error(`切换生产模式失败：${apiErrorMessage(error, '请稍后重试')}`);
    } finally { setModeSaving(false); }
  };
  const productionMode = batchFactoryProductionModeVisibility(draftPatch.productionMode).mode;
  return <Modal title={<div className="batch-factory-unified-settings-title"><BatchFactoryAutomationPresetManager value={draftPatch} onLoad={setDraftPatch} /><span>统一配置</span><BatchFactoryProductionModeControl value={draftPatch.productionMode} disabled={modeSaving} onChange={changeProductionMode} /></div>} open={open} onCancel={onClose} width={980} destroyOnClose={false} className="batch-factory-unified-settings-modal" footer={<Space><span className="batch-factory-unified-settings-scope">保存将覆盖本批次 {(Array.isArray(batch?.books) ? batch.books.length : 0) || 0} 本书的对应配置（包括单书已改过的值）</span><Button onClick={onClose}>取消</Button><Button type="primary" loading={saving} onClick={save}>保存统一配置</Button></Space>}>
    {modeNotice ? <Alert type="warning" showIcon closable message={modeNotice} style={{ marginBottom: 12 }} onClose={() => setModeNotice('')} /> : null}
    <Tabs items={[
      { key: 'models', label: '模型配置', children: <BatchFactoryEngineSettingsForm value={draftPatch} onChange={setDraftPatch} sections={['models', 'audio', 'sensitive']} active={open} /> },
      { key: 'reasoning', label: 'AI 推理', children: <BatchFactoryAiReasoningForm batchId={batch?.id} productionMode={draftPatch.productionMode} value={draftPatch.aiPromptConfig} onChange={aiPromptConfig => setDraftPatch(current => ({ ...current, aiPromptConfig }))} /> },
      { key: 'publish', label: '发布统一', forceRender: true, children: productionMode === 'comic' ? <BatchFactoryComicPublishSettingsForm value={draftPatch} onChange={setDraftPatch} books={batch?.books || []} /> : <BatchFactoryPublishSettingsForm value={draftPatch} onChange={setDraftPatch} active={open} /> }
    ]} />
  </Modal>;
}
