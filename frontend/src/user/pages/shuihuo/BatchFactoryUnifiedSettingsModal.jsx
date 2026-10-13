import { SettingOutlined } from '@ant-design/icons';
import { Alert, Button, Divider, Input, InputNumber, Modal, Segmented, Select, Space, Switch, Tabs, Tooltip, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { listAvailableModels, modelSelectOptions } from '../../../shared/api/modelCatalog';
import { checkWebSubmitEnvironment, getWebSubmitConfig, saveWebSubmitConfig, syncWebSubmitConfigs, syncWebSubmitStyles, testWebSubmitVisible } from '../../../shared/api/novelFetch';
import {
  createAutomationPreset,
  deleteAutomationPreset,
  get121OrganizationOptions,
  listAutomationPresets,
  listSystemPresetCatalog,
  updateAutomationPreset
} from '../../../shared/api/batchFactoryV11';
import { videoProviderForModel } from './videoProviderBinding';
import { normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { getConstraintPresetTexts } from '../../../shared/api/generation';

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
const AUDIO_SOURCE_OPTIONS = [{ value: 'tts', label: 'TTS（系统配音）' }, { value: 'original', label: '原声（视频自带声音）' }];
const voices = [
  { value: 'zh-CN-XiaoxiaoNeural', label: '晓晓（女声·温柔）' },
  { value: 'zh-CN-XiaochenNeural', label: '晓辰（女声·知性）' },
  { value: 'zh-CN-YunxiNeural', label: '云希（男声·清朗）' },
  { value: 'zh-CN-YunyangNeural', label: '云扬（男声·阳光）' }
];
const styles = ['general', 'cheerful', 'sad', 'friendly', 'chat'].map(value => ({ value, label: value === 'general' ? '通用' : value }));
const selectOption = preset => ({ value: preset.id, label: `${preset.name || preset.id} · v${preset.version || 1}` });
const presetValue = preset => preset ? { presetId: preset.id, presetName: preset.name || preset.id, presetSlot: preset.slot || '', presetVersion: preset.version || 1, constraintCategory: preset.constraintCategory || '' } : { presetId: '', presetName: '', presetSlot: '', presetVersion: null, constraintCategory: '' };
const TARGET_ADMIN_URL = 'http://two.121w.com/tttadmin/index.php';
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

export function BatchFactoryEngineSettingsForm({ value, onChange, sections = ['models', 'audio', 'publish'], active = true }) {
  const [models, setModels] = useState([]);
  const [modelsError, setModelsError] = useState('');
  const [modelsLoading, setModelsLoading] = useState(false);
  const [modelReloadKey, setModelReloadKey] = useState(0);
  const patch = next => onChange({ ...value, ...next });
  const tts = { ...DEFAULT_TTS, ...(value.tts || {}) };
  const modelSettingsEnabled = sections.includes('models');
  const modelOptions = useMemo(() => ({
    text: modelSelectOptions(models, 'text'),
    image: modelSelectOptions(models, 'image'),
    video: modelSelectOptions(models, 'video')
  }), [models]);
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
      <label className="batch-factory-engine-field"><span><b>图片模型</b></span><Select allowClear loading={modelsLoading} value={value.imageModelId || undefined} options={modelOptions.image} placeholder="选择已启用图片模型" onChange={imageModelId => patch({ imageModelId: imageModelId || '' })} /></label>
      <label className="batch-factory-engine-field"><span><b>视频模型</b><HelpTip title="视频模型">只决定最终使用哪个视频引擎及其画幅、分辨率等生成参数，不决定导演分镜或提示词内容。</HelpTip></span><Select allowClear loading={modelsLoading} value={value.videoModelId || undefined} options={modelOptions.video} placeholder="选择已启用视频模型" onChange={videoModelId => patch({ videoModelId: videoModelId || '', ...(videoModelId ? { videoProvider: videoProviderForModel(videoModelId, value.videoProvider) } : {}) })} /></label>
      <label className="batch-factory-engine-field"><span><b>图片画幅</b><HelpTip title="图片画幅">仅用于参考图与资产生图。</HelpTip></span><Segmented value={value.imageAspectRatio || value.aspectRatio || '9:16'} options={['16:9', '9:16', '1:1']} onChange={imageAspectRatio => patch({ imageAspectRatio })} /></label>
      <label className="batch-factory-engine-field"><span><b>视频画幅</b><HelpTip title="视频画幅">视频不会跟随参考图画幅。</HelpTip></span><Segmented value={value.videoAspectRatio || (value.aspectRatio === '16:9' ? '16:9' : '9:16')} options={['16:9', '9:16']} onChange={videoAspectRatio => patch({ videoAspectRatio })} /></label>
      <label className="batch-factory-engine-field"><span><b>视频分辨率</b></span><Segmented value={value.videoResolution || '720p'} options={['480p', '720p', '1080p']} onChange={videoResolution => patch({ videoResolution })} /></label>
      <label className="batch-factory-engine-field"><span><b>分镜时长</b></span><Segmented value={Number(value.storyboardDurationLimit) === 15 ? 15 : 10} options={[{ value: 10, label: '10 秒' }, { value: 15, label: '15 秒' }]} onChange={storyboardDurationLimit => patch({ storyboardDurationLimit, maxVideoDuration: storyboardDurationLimit })} /></label>
      <label className="batch-factory-engine-field"><span><b>音频来源</b><HelpTip title="音频来源">默认使用 TTS，原声保留视频模型自带声音。</HelpTip></span><Segmented block value={normalizeAudioSource(value.audioSource)} options={AUDIO_SOURCE_OPTIONS} onChange={audioSource => onChange(normalizeAudioSourceSettings({ ...value, audioSource }))} /></label>
      <label className="batch-factory-engine-field"><span><b>配音音色</b></span><Select value={tts.voice} options={voices} onChange={voice => patch({ tts: { ...tts, voice } })} /></label>
      <label className="batch-factory-engine-field"><span><b>配音风格</b></span><Select value={tts.style} options={styles} onChange={style => patch({ tts: { ...tts, style } })} /></label>
      <label className="batch-factory-engine-field"><span><b>语速</b></span><InputNumber min={0.5} max={2} step={0.1} value={tts.speed} onChange={speed => patch({ tts: { ...tts, speed: speed ?? DEFAULT_TTS.speed } })} /></label>
    </div></section>
    {modelsError ? <Alert type="warning" showIcon message="模型目录暂不可用" description={<Space direction="vertical"><span>{modelsError}</span><Button size="small" onClick={() => setModelReloadKey(value => value + 1)}>重试读取模型</Button></Space>} /> : null}
    {!modelsLoading && !modelsError && !models.length ? <Alert type="warning" showIcon message="个人中心没有已启用模型" description={<Space direction="vertical"><span>请先在个人中心按文本、图片、视频类型新增并启用模型，再返回当前批量作品选择。</span><Button type="link" href="/api-config">前往个人中心配置模型</Button></Space>} /> : null}
    </> : null}
    {sections.includes('audio') ? <section className="batch-factory-engine-card"><header><b>跟随配音</b><HelpTip title="跟随配音">开启后只测量没有有效缓存的视频原文非空行。</HelpTip></header><div className="batch-factory-engine-card-body"><Space direction="vertical" style={{ width: '100%' }}>
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
    const organization = selectedProfileOrganization;
    if (organization && organization !== publish.organization) {
      patch({ publishSettings: { ...publish, organization } });
    }
  }, [publish.websiteProfileId, publish.websiteProfiles, selectedProfileOrganization]);
  const hasSession = result => Boolean((result?.checks || []).some(check => SESSION_CHECK_NAMES.includes(check.name) && check.ok));
  const checkDetail = (result, name) => String((result?.checks || []).find(check => check.name === name)?.detail || '').trim();
  const sessionDetail = result => String((result?.checks || []).find(check => SESSION_CHECK_NAMES.includes(check.name))?.detail || (result?.checks || [])[0]?.detail || '').trim();
  const publishSessionReady = Boolean(environment?.session && environment?.visible?.ok);
  const organizationCatalogReady = Boolean(environment?.session);
  const loadAccount = async () => {
    const [config, result] = await Promise.all([getWebSubmitConfig(), checkWebSubmitEnvironment()]);
    setAccount(config?.settings || config?.config || {});
    setEnvironment({ environment: result, session: hasSession(result), visible: null });
  };
  const markPublishSessionExpired = error => {
    const detail = error?.message || '视频管理系统登录会话已失效，请重新验证登录';
    setEnvironment(current => ({ ...current, error: detail, session: false, visible: { ok: false, error: detail } }));
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
    } catch (error) { message.error(error?.message || '同步网站配置档失败'); }
    finally { setSyncing(''); }
  };
  const syncStyles = async () => {
    if (!publishSessionReady) return message.warning('请先登录并验证视频管理系统');
    setSyncing('styles');
    try {
      const result = await syncWebSubmitStyles({ persist: false });
      const styles = Array.isArray(result?.settings?.style_catalog) ? result.settings.style_catalog : (Array.isArray(result?.styles) ? result.styles : []);
      patch({ publishSettings: { ...publish, websiteStyleCatalog: styles } });
    } catch (error) { message.error(error?.message || '同步风格目录失败'); }
    finally { setSyncing(''); }
  };
  return <><Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>批量默认发布规则 <HelpTip title="批量默认发布规则">这里维护视频管理系统连接、网站配置映射和上传素材规则；单书配置可覆盖默认值。</HelpTip></span>} />
    <section className="batch-factory-vms-connection-card"><div className="batch-factory-vms-main"><div className="batch-factory-vms-icon" aria-hidden="true">▣</div><div className="batch-factory-vms-copy"><div className="batch-factory-vms-title-row"><b>视频管理系统</b><span>{publishSessionReady ? '已登录并验证' : account?.username ? '待重新验证' : '未登录'}</span></div><div className="batch-factory-vms-meta"><span>当前账号</span><strong>{account?.username || '尚未登录'}</strong></div><div className="batch-factory-vms-meta"><span>后台地址</span><a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">two.121w.com/tttadmin</a></div></div></div><div className="batch-factory-vms-actions"><Button href={TARGET_ADMIN_URL} target="_blank">打开后台</Button><Button onClick={openLoginForm}>{account?.username ? '更换账号' : '登录并验证'}</Button><Button onClick={selfCheck} loading={checking}>{account?.username ? '重新验证' : '环境自检'}</Button></div></section>
    <div className="batch-factory-publish-two-column"><section className="batch-factory-engine-card"><header><b>发布映射</b><HelpTip title="发布映射">同步真实后台配置后，绑定到当前批量作品。</HelpTip></header><div className="batch-factory-engine-card-body"><label className="batch-factory-engine-field"><span><b>网站配置档</b></span><Select allowClear showSearch value={publish.websiteProfileId || undefined} options={(publish.websiteProfiles || []).map(profile => ({ value: profile.id, label: profile.name || profile.id }))} placeholder="同步网站配置档后选择" onChange={websiteProfileId => { const profile = (publish.websiteProfiles || []).find(item => item.id === websiteProfileId) || {}; patch({ publishSettings: { ...publish, websiteProfileId: websiteProfileId || '', versionProfile: profile.name || '', organization: websiteProfileOrganization(profile) } }); }} /></label><div className="batch-factory-publish-sync-row"><Button disabled={!publishSessionReady} onClick={syncProfiles} loading={syncing === 'profiles'}>同步网站配置档</Button><Button disabled={!publishSessionReady} onClick={syncStyles} loading={syncing === 'styles'}>同步风格目录</Button></div><div className="batch-factory-publish-ai-note"><b>组织归属</b><span>由网站配置档自动带入，不在这里重复选择。</span></div></div></section>
      <section className="batch-factory-engine-card"><header><b>批量默认上传规则</b><HelpTip title="批量默认上传规则">这些规则可被单书发布设置覆盖。</HelpTip></header><div className="batch-factory-engine-card-body"><label className="batch-factory-engine-field"><span><b>上传视频类型</b></span><Segmented value={publish.uploadVideoType || 'merged'} options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} onChange={uploadVideoType => patch({ publishSettings: { ...publish, uploadVideoType } })} /></label><div className="batch-factory-publish-rule-list"><section><div><b>素材复用</b></div><Switch checked={publish.materialReuse === true} onChange={materialReuse => patch({ publishSettings: { ...publish, materialReuse } })} /></section><section><div><b>水平翻转</b></div><Switch checked={publish.horizontalFlip === true} onChange={horizontalFlip => patch({ publishSettings: { ...publish, horizontalFlip } })} /></section><section><div><b>改文后上传</b></div><Switch checked={value.publishRewriteEnabled === true} onChange={publishRewriteEnabled => patch({ publishRewriteEnabled })} /></section></div></div></section></div>
    <section className="batch-factory-engine-card"><header><b>连接与环境状态</b><HelpTip title="连接与环境状态">登录、后台页面和直接接口均需可用。</HelpTip></header><div className="batch-factory-engine-card-body"><div className="batch-factory-env-status-row"><div><b>登录会话</b><span>{environment ? (environment.session ? '已登录' : '未登录 / 已失效') : '等待检查'}</span><small>{sessionDetail(environment?.environment)}</small></div><div><b>后台页面</b><span>{environment?.visible ? (environment.visible.ok ? '可访问' : '验证未通过') : '等待检查'}</span></div><div><b>直接接口</b><span>{checkDetail(environment?.environment, '121 直接接口') || '等待检查'}</span></div><Button onClick={selfCheck} loading={checking}>重新检查</Button></div>{environment?.error ? <Alert type="error" showIcon message="视频管理系统环境自检失败" description={environment.error} /> : null}</div></section>
  </Space>
  <Modal title={account?.username ? '更换视频管理系统账号' : '登录视频管理系统'} open={loginOpen} onCancel={() => { setLoginPassword(''); setLoginOpen(false); }} onOk={login} confirmLoading={loginBusy} okText="登录并验证" destroyOnClose><Space direction="vertical" size={12} style={{ width: '100%' }}><Input value={loginUsername} autoComplete="username" placeholder="视频管理系统账号" onChange={event => setLoginUsername(event.target.value)} /><Input.Password value={loginPassword} autoComplete="current-password" placeholder="视频管理系统密码" onChange={event => setLoginPassword(event.target.value)} /><a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">打开视频管理系统后台</a></Space></Modal>
  </>;
}

export function BatchFactoryAiReasoningForm({ value, onChange }) {
  const [catalog, setCatalog] = useState([]);
  const config = value && typeof value === 'object' ? value : {};
  const assets = config.assets || {};
  const constraints = config.constraints || {};
  const derivedOpening = config.derivedOpening || {};
  const video = config.video?.presetId && config.video.presetId !== 'batch-video-meta'
    ? config.video
    : { presetId: 'batch-video-sd', presetName: 'SD 视频提示词', presetSlot: 'batch.video-meta', presetVersion: null, constraintCategory: '' };
  const visual = config.visual || {};
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
  const bySlot = slot => promptOptions(catalog.filter(item => item.slot === slot).map(selectOption));
  const presetFor = presetId => presetValue(catalog.find(item => item.id === presetId));
  const bySlots = slots => promptOptions(catalog.filter(item => slots.includes(item.slot)).map(selectOption));
  const selectedConstraint = category => (constraints.selections || []).find(item => item.constraintCategory === category)?.presetId || constraints[category]?.presetId || '';
  const byConstraint = category => promptOptions(catalog.filter(item => item.kind === 'addon' && item.constraintCategory === category).map(selectOption));
  const updateConstraint = (category, presetId) => {
    presetId = normalizePromptId(presetId);
    const rest = (constraints.selections || []).filter(item => item.constraintCategory !== category);
    if (!presetId) {
      onChange({ ...config, constraints: { ...constraints, enabled: true, enabledCategories: [...new Set([...(constraints.enabledCategories || []), category])], selections: rest } });
      return;
    }
    const selected = catalog.find(item => item.id === presetId);
    // 预设目录为了提速不带正文，这里按预设号取一次权威正文一并保存，
    // 否则导演/视频生产拿不到约束内容（历史版本正是漏了这一步）。
    getConstraintPresetTexts([presetId])
      .then(result => {
        const body = String(result?.texts?.[presetId] || '');
        onChange({ ...config, constraints: { ...constraints, enabled: true, enabledCategories: [...new Set([...(constraints.enabledCategories || []), category])], selections: [...rest, { ...presetValue(selected), body }] } });
      })
      .catch(() => message.error('系统预设提示词读取失败，请重试'));
  };
  return <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>AI 推理 <HelpTip title="AI 推理">选择规则后，资产提取、智能统一、导演分镜和最终 Prompt 都共用同一份批量草稿。</HelpTip></span>} />
    <section className="batch-factory-engine-card"><header><b>资产设置</b><HelpTip title="资产设置">人物、场景和道具统一使用提取提示词；选择“无”表示不使用提示词。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>人物场景道具提取提示词</b></span><Select allowClear value={assets.extraction?.presetId || undefined} options={bySlot('script.asset-extraction')} placeholder="选择提示词或无" onChange={presetId => onChange({ ...config, assets: { ...assets, enabled: Boolean(normalizePromptId(presetId)), extraction: presetFor(normalizePromptId(presetId)) } })} /></label>
      <label className="batch-factory-engine-field"><span className="batch-factory-option-title"><b>换开头</b><Tooltip title="开启后，2 个及以上分镜的书会为分镜一生成多个开场变体并分别合成上传；数量包含原始开头。"><span className="batch-factory-help-button" aria-label="换开头说明">?</span></Tooltip></span><Space><Switch checked={config.openingEnabled === true} onChange={enabled => onChange({ ...config, openingEnabled: enabled, openingCount: config.openingCount ?? 4 })} /><InputNumber min={1} max={8} precision={0} disabled={config.openingEnabled !== true} value={config.openingCount ?? 4} onChange={count => onChange({ ...config, openingCount: count ?? 4 })} addonAfter="条（含原始）" /></Space></label>
    </div></section>
    <section className="batch-factory-engine-card"><header><b>约束设置</b><HelpTip title="约束设置">画面前缀、画质、画面限制和负面提示词在这里统一选择；不选的类别不会注入，可在单书配置里单独覆盖。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>画面前缀词（智能统一）</b></span><Select allowClear value={selectedConstraint('prefix')} options={byConstraint('prefix')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('prefix', presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>画质约束</b></span><Select allowClear value={selectedConstraint('quality')} options={byConstraint('quality')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('quality', presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>画面限制</b></span><Select allowClear value={selectedConstraint('restriction')} options={byConstraint('restriction')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('restriction', presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>负面提示词</b></span><Select allowClear value={selectedConstraint('negative')} options={byConstraint('negative')} placeholder="选择提示词或无" onChange={presetId => updateConstraint('negative', presetId)} /></label>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}><span>基础设定</span><Switch checked={constraints.baseSetup?.enabled !== false} onChange={enabled => onChange({ ...config, constraints: { ...constraints, baseSetup: { enabled } } })} /></Space>
    </div></section>
    <section className="batch-factory-engine-card"><header><b>视频设置</b><HelpTip title="视频设置">视频提示词决定结构化导演分镜的输出规则与最终模板。</HelpTip></header><div className="batch-factory-engine-card-body">
      <label className="batch-factory-engine-field"><span><b>衍生开篇</b></span><Select allowClear value={derivedOpening.presetId || undefined} options={bySlots(['batch.hook-adaptation', 'batch.original-director', 'batch.viral-director'])} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, derivedOpening: { ...derivedOpening, ...presetValue(catalog.find(item => item.id === normalizedId)), enabled: Boolean(normalizedId) } }); }} /></label>
      <label className="batch-factory-engine-field"><span><b>视频提示词</b><HelpTip title="视频提示词">决定导演分镜和最终 VIDEO Prompt 的规则与格式；与上面的具体视频模型相互独立。</HelpTip></span><Select allowClear value={video.presetId || undefined} options={bySlot('batch.video-meta')} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, video: { ...video, ...presetValue(catalog.find(item => item.id === normalizedId)), enabled: Boolean(normalizedId) } }); }} /></label>
      <label className="batch-factory-engine-field"><span><b>画面提示词</b></span><Select allowClear value={visual.presetId || undefined} options={bySlot('batch.visual-meta')} placeholder="选择提示词或无" onChange={presetId => { const normalizedId = normalizePromptId(presetId); onChange({ ...config, visual: { ...visual, ...presetValue(catalog.find(item => item.id === normalizedId)) } }); }} /></label>
    </div></section>
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
  <Modal title="自动化预设" open={presetOpen} onCancel={() => setPresetOpen(false)} footer={null} width={620} destroyOnClose><Space direction="vertical" size={14} style={{ width: '100%' }}>
    <Alert type="info" showIcon message={<span>预设只保存统一配置 <HelpTip title="预设只保存统一配置">不保存正文、资产、导演结果或视频结果；载入后仍须确认当前统一配置才会应用到批量。</HelpTip></span>} />
    <label className="batch-factory-engine-field"><span><b>已保存预设</b></span><Select allowClear value={selectedPresetId || undefined} onChange={setSelectedPresetId} placeholder="选择自动化预设" options={presets.map(item => ({ value: item.id, label: `${item.name} · v${item.version}` }))} /></label>
    <Space wrap><Button disabled={!selectedPreset} onClick={loadPreset}>载入预设</Button></Space>
    <Divider />
    <label className="batch-factory-engine-field"><span><b>{selectedPreset ? '预设新名称' : '新预设名称'}</b></span><Input value={presetName} maxLength={255} placeholder={selectedPreset?.name || '例如：女频 H3 全自动'} onChange={event => setPresetName(event.target.value)} /></label>
    <Space wrap><Button type="primary" loading={presetBusy} onClick={savePreset}>保存为新预设</Button><Button disabled={!selectedPreset} loading={presetBusy} onClick={renamePreset}>重命名所选预设</Button><Button danger disabled={!selectedPreset} onClick={removePreset}>删除所选预设</Button></Space>
  </Space></Modal></>;
}

export function BatchFactoryUnifiedSettingsModal({ open, batch, onClose, onSaved }) {
  const [draftPatch, setDraftPatch] = useState({});
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    if (open) setDraftPatch(normalizeLegacyPatch(batch?.settingsState?.patch));
  }, [open, batch?.id]);
  const save = async () => {
    setSaving(true);
    const videoPresetId = draftPatch.aiPromptConfig?.video?.presetId;
    const nextPatch = videoPresetId === 'batch-video-sd' && !draftPatch.storyTimelineVersion
      ? { ...draftPatch, storyTimelineVersion: 'sd.original-story.v1' }
      : draftPatch;
    try { const result = await onSaved(nextPatch); if (result !== false) onClose(); }
    finally { setSaving(false); }
  };
  return <Modal title={<div className="batch-factory-unified-settings-title"><BatchFactoryAutomationPresetManager value={draftPatch} onLoad={setDraftPatch} /><span>统一配置</span></div>} open={open} onCancel={onClose} width={980} destroyOnClose={false} className="batch-factory-unified-settings-modal" footer={<Space><Button onClick={onClose}>取消</Button><Button type="primary" loading={saving} onClick={save}>保存统一配置</Button></Space>}>
    <Tabs items={[
      { key: 'models', label: '模型配置', children: <BatchFactoryEngineSettingsForm value={draftPatch} onChange={setDraftPatch} sections={['models', 'audio']} active={open} /> },
      { key: 'reasoning', label: 'AI 推理', children: <BatchFactoryAiReasoningForm value={draftPatch.aiPromptConfig} onChange={aiPromptConfig => setDraftPatch(current => ({ ...current, aiPromptConfig }))} /> },
      { key: 'publish', label: '发布统一', children: <BatchFactoryPublishSettingsForm value={draftPatch} onChange={setDraftPatch} active={open} /> }
    ]} />
  </Modal>;
}
