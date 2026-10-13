import { Alert, Button, Input, InputNumber, Modal, Popover, Segmented, Select, Space, Switch, Tabs, Tag, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { listAvailableModels } from '../../../shared/api/modelCatalog';
import { getConfig } from '../../../shared/api/config';
import { checkWebSubmitEnvironment, getWebSubmitConfig, saveWebSubmitConfig, syncWebSubmitConfigs, syncWebSubmitStyles, testWebSubmitVisible } from '../../../shared/api/novelFetch';
import { selectedModelState } from './modelSelectionState';
import { videoProviderForModel } from './videoProviderBinding';
import { normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';

const TARGET_ADMIN_URL = 'http://two.121w.com/tttadmin/index.php';
// The shared novel-fetch verification service predates the VMS UI label.  It
// still returns the original check name, so accept both names at this boundary.
const SESSION_CHECK_NAMES = ['视频管理系统登录会话', '121 后台登录会话'];

const providerOptions = [
  { value: 'personal_api', label: '个人中心 API' },
  { value: 'doubao_local_executor', label: '豆包本地执行器' },
  { value: 'autodl_comfyui', label: 'AutoDL ComfyUI' }
];
const DEFAULT_TTS = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 };
const AUDIO_SOURCE_OPTIONS = [{ value: 'tts', label: 'TTS（系统配音）' }, { value: 'original', label: '原声（视频自带声音）' }];
const TTS_VOICE_OPTIONS = [
  { label: '晓晓（女声·温柔）', value: 'zh-CN-XiaoxiaoNeural' },
  { label: '晓辰（女声·知性）', value: 'zh-CN-XiaochenNeural' },
  { label: '云希（男声·清朗）', value: 'zh-CN-YunxiNeural' },
  { label: '云扬（男声·阳光）', value: 'zh-CN-YunyangNeural' },
  { label: '晓伊（女声·甜美）', value: 'zh-CN-XiaoyiNeural' },
  { label: '云健（男声·稳重）', value: 'zh-CN-YunjianNeural' }
];
const TTS_STYLE_OPTIONS = [
  { label: '通用', value: 'general' },
  { label: '开心', value: 'cheerful' },
  { label: '悲伤', value: 'sad' },
  { label: '友好', value: 'friendly' },
  { label: '聊天', value: 'chat' }
];
function HelpTip({ title, children }) {
  return <Popover trigger="click" placement="top" title={title} content={<div className="batch-factory-help-popover">{children}</div>}><button type="button" className="batch-factory-help-button" aria-label={`查看${title}说明`}>?</button></Popover>;
}
function Field({ label, children, note }) {
  return <label className="batch-factory-engine-field"><span><b>{label}</b>{note ? <small>{note}</small> : null}</span>{children}</label>;
}
function EngineCard({ title, description, children }) {
  return <section className="batch-factory-engine-card"><header><b>{title}</b>{description ? <small>{description}</small> : null}</header><div className="batch-factory-engine-card-body">{children}</div></section>;
}
function storyboardDuration(value) { return Number(value) === 15 ? 15 : 10; }

export function BatchFactoryEngineSettingsDrawer({ open, batch, onClose, onSave }) {
  const [form, setForm] = useState({});
  const [saving, setSaving] = useState(false);
  const [checking, setChecking] = useState(false);
  const [environment, setEnvironment] = useState(null);
  const [models, setModels] = useState([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [modelsError, setModelsError] = useState('');
  const [syncing, setSyncing] = useState('');
  const [account, setAccount] = useState(null);
  const [publishSession, setPublishSession] = useState(null);
  const [loginOpen, setLoginOpen] = useState(false);
  const [loginUsername, setLoginUsername] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [loginBusy, setLoginBusy] = useState(false);
  const [accountTts, setAccountTts] = useState(DEFAULT_TTS);
  useEffect(() => {
    if (!open) return;
    const current = batch?.settingsState?.patch || {};
    const duration = storyboardDuration(current.storyboardDurationLimit);
    const fixedSingleVideo = current.fixedSingleVideo === true;
    setForm({
      ...current,
      storyboardDurationLimit: duration,
      maxVideoDuration: duration,
      audioDurationSeconds: 0,
      ...(fixedSingleVideo ? { audioPlanningEnabled: false, audioMergeEnabled: false } : {})
    });
  }, [open, batch?.id, batch?.settingsState?.revision]);
  useEffect(() => {
    if (!open) return undefined;
    let active = true;
    setModelsLoading(true);
    setModelsError('');
    Promise.all([
      listAvailableModels('image'),
      listAvailableModels('text'),
      listAvailableModels('video')
    ]).then(([imageModels, textModels, videoModels]) => {
      if (active) setModels([...imageModels, ...textModels, ...videoModels]);
    }).catch(error => {
      if (active) {
        setModels([]);
        setModelsError(error?.message || '未能读取个人中心已启用模型');
      }
    }).finally(() => { if (active) setModelsLoading(false); });
    return () => { active = false; };
  }, [open]);
  useEffect(() => {
    if (!open) return undefined;
    let active = true;
    getConfig().then(config => {
      if (!active) return;
      setAccountTts({ ...DEFAULT_TTS, ...(config?.tts || {}) });
    }).catch(() => { if (active) setAccountTts(DEFAULT_TTS); });
    return () => { active = false; };
  }, [open]);
  const modelOptions = useMemo(() => ({
    image: models.filter(model => model.kind === 'image').map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id })),
    text: models.filter(model => model.kind === 'text').map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id })),
    video: models.filter(model => model.kind === 'video').map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id }))
  }), [models]);
  const selectedTextModel = selectedModelState(models.filter(model => model.kind === 'text'), form.textModelId);
  function modelPlaceholder(kind, label) {
    if (modelsLoading) return '正在读取个人中心模型…';
    return modelOptions[kind].length ? `选择已启用${label}模型` : `个人中心没有已启用${label}模型`;
  }
  const tts = { ...DEFAULT_TTS, ...accountTts, ...(form.tts || {}) };
  function patch(next) { setForm(current => ({ ...current, ...next })); }
  function patchTts(next) { patch({ tts: { ...tts, ...next } }); }
  function patchFixedSingleVideo(fixedSingleVideo) {
    patch({ fixedSingleVideo, ...(fixedSingleVideo ? { audioPlanningEnabled: false, audioMergeEnabled: false } : {}) });
    if (fixedSingleVideo) message.info('固定开头只生产 VIDEO01，两个跟随配音功能已关闭。');
  }
  async function save() {
    const normalized = normalizeAudioSourceSettings({
      ...form,
      tts: { ...tts },
      audioDurationSeconds: 0,
      ...(form.fixedSingleVideo === true ? { audioPlanningEnabled: false, audioMergeEnabled: false } : {})
    });
    setSaving(true);
    try { const saved = await onSave(normalized); if (saved) onClose(); } finally { setSaving(false); }
  }
  function has121Session(environment) {
    return Boolean(environment?.ok && (environment?.checks || []).find(check => SESSION_CHECK_NAMES.includes(check.name))?.ok);
  }
  function check121Detail(environment, name) {
    return String((environment?.checks || []).find(check => check.name === name)?.detail || '').trim();
  }
  async function load121Account() {
    const [config, environment] = await Promise.all([getWebSubmitConfig(), checkWebSubmitEnvironment()]);
    const settings = config?.settings || config?.config || {};
    setAccount(settings);
    setPublishSession({ environment, visible: null });
    return { settings, environment };
  }
  async function selfCheck() {
    setChecking(true);
    try {
      const environment = await checkWebSubmitEnvironment();
      const session = has121Session(environment);
      if (!session) {
        setEnvironment({ environment, session: false, visible: null });
        setPublishSession({ environment, visible: null });
        return;
      }
      const visible = await testWebSubmitVisible();
      setEnvironment({ environment, session: true, visible });
      setPublishSession({ environment, visible });
    } catch (error) {
      const failure = { error: error?.message || '视频管理系统环境自检失败' };
      setEnvironment(failure);
      setPublishSession(failure);
    } finally { setChecking(false); }
  }
  function openLogin() {
    setLoginUsername(String(account?.username || '').trim());
    setLoginPassword('');
    setLoginOpen(true);
  }
  async function login121() {
    if (!loginUsername.trim() || !loginPassword) {
      message.warning('请输入视频管理系统账号和密码');
      return;
    }
    setLoginBusy(true);
    try {
      const saved = await saveWebSubmitConfig({ settings: { ...(account || {}), username: loginUsername.trim(), password: loginPassword } });
      setAccount(saved?.settings || { username: loginUsername.trim(), password_masked: true });
      setLoginPassword('');
      await selfCheck();
      setLoginOpen(false);
      message.success('视频管理系统已登录并完成验证');
    } catch (error) { message.error(error?.message || '视频管理系统登录失败'); } finally { setLoginBusy(false); }
  }
  async function syncProfiles() {
    if (!publishSessionReady) { message.warning('请先登录并验证视频管理系统'); return; }
    setSyncing('profiles');
    try {
      const result = await syncWebSubmitConfigs({ persist: false });
      const profiles = Array.isArray(result?.groups) ? result.groups : [];
      patch({ publishSettings: { ...(form.publishSettings || {}), websiteProfiles: profiles } });
    } finally { setSyncing(''); }
  }
  async function syncStyles() {
    if (!publishSessionReady) { message.warning('请先登录并验证视频管理系统'); return; }
    setSyncing('styles');
    try {
      const result = await syncWebSubmitStyles({ persist: false });
      const styles = Array.isArray(result?.settings?.style_catalog) ? result.settings.style_catalog : (Array.isArray(result?.styles) ? result.styles : []);
      patch({ publishSettings: { ...(form.publishSettings || {}), websiteStyleCatalog: styles } });
    } finally { setSyncing(''); }
  }
  const publish = form.publishSettings || {};
  const publishSessionReady = Boolean(has121Session(publishSession?.environment) && publishSession?.visible?.ok);
  useEffect(() => {
    if (!open) return undefined;
    let active = true;
    load121Account().catch(error => { if (active) setPublishSession({ error: error?.message || '无法读取 视频管理系统状态' }); });
    return () => { active = false; };
  }, [open]);
  return <><Modal
    title="引擎配置"
    open={open}
    onCancel={onClose}
    width={880}
    destroyOnClose={false}
    className="shuihuo-engine-modal batch-factory-engine-modal batch-factory-engine-modal-final"
    footer={<div className="batch-factory-engine-footer"><span className="batch-factory-engine-footer-note">这里是当前批次的统一设置；应用后所有小说先统一为这一套。</span><Space><Button onClick={onClose}>取消</Button><Button type="primary" loading={saving} onClick={save}>应用到当前全部小说</Button></Space></div>}
  >
    <Tabs items={[
      {
        key: 'models',
        label: '模型配置',
        children: <div className="batch-factory-engine-drawer batch-factory-engine-models-final">
          <Alert type="info" showIcon message="当前批次统一生产参数" description="这里统一设置当前批次所有小说；之后仍可进入任意单书单独修改，不会影响其它小说。" />
          <EngineCard title="核心视频引擎" description="当前设置应用到本批次全部小说；之后可单独修改任意一本。">
            <div className="batch-factory-engine-grid">
              <Field label="视频引擎" note="生成通道"><Select value={form.videoProvider || 'personal_api'} options={providerOptions} onChange={videoProvider => patch({ videoProvider })} /></Field>
              <Field label="视频模型" note="选择后自动匹配生成通道"><Select allowClear loading={modelsLoading} value={form.videoModelId || undefined} options={modelOptions.video} placeholder={modelPlaceholder('video', '视频')} onChange={videoModelId => patch({ videoModelId: videoModelId || '', ...(videoModelId ? { videoProvider: videoProviderForModel(videoModelId, form.videoProvider) } : {}) })} /></Field>
              <Field label="分镜时长" note="单个 VIDEO 的目标上限"><Segmented value={storyboardDuration(form.storyboardDurationLimit)} options={[{ value: 10, label: '10s' }, { value: 15, label: '15s' }]} onChange={storyboardDurationLimit => patch({ storyboardDurationLimit, maxVideoDuration: storyboardDurationLimit })} /></Field>
              <Field label="图片画幅" note="仅用于参考图与资产生图"><Segmented value={form.imageAspectRatio || form.aspectRatio || '9:16'} options={['16:9', '9:16', '1:1']} onChange={imageAspectRatio => patch({ imageAspectRatio })} /></Field>
              <Field label="视频画幅" note="只影响视频输出"><Segmented value={form.videoAspectRatio || (form.aspectRatio === '16:9' ? '16:9' : '9:16')} options={['16:9', '9:16']} onChange={videoAspectRatio => patch({ videoAspectRatio })} /></Field>
              <Field label="视频分辨率" note="随视频任务提交"><Segmented value={form.videoResolution || '720p'} options={['480p', '720p', '1080p']} onChange={videoResolution => patch({ videoResolution })} /></Field>
            </div>
            <section className="batch-factory-audio-option batch-factory-fixed-option"><div><b>固定开头</b><small>开启后只生产 VIDEO01，同时禁用两个跟随配音功能。</small></div><Switch checked={form.fixedSingleVideo === true} onChange={patchFixedSingleVideo} /></section>
          </EngineCard>

          <EngineCard title="配音与时长" description="统一的是配音参数与跟随策略；每本小说的真实秒数会在需要时独立生成和读取。">
            <div className="batch-factory-engine-grid batch-factory-tts-grid">
              <Field label="音频来源"><Segmented block value={normalizeAudioSource(form.audioSource)} options={AUDIO_SOURCE_OPTIONS} onChange={audioSource => patch(normalizeAudioSourceSettings({ ...form, audioSource }))} /></Field>
            </div>
            <div className="batch-factory-engine-grid batch-factory-tts-grid">
              <Field label="配音音色"><Select value={tts.voice} options={TTS_VOICE_OPTIONS} onChange={voice => patchTts({ voice })} /></Field>
              <Field label="配音风格"><Select value={tts.style} options={TTS_STYLE_OPTIONS} onChange={style => patchTts({ style })} /></Field>
              <Field label="语速"><InputNumber min={0.5} max={2} step={0.1} value={tts.speed} onChange={speed => patchTts({ speed: speed ?? DEFAULT_TTS.speed })} /></Field>
              <Field label="音调"><InputNumber min={-50} max={50} step={1} value={tts.pitch} onChange={pitch => patchTts({ pitch: pitch ?? DEFAULT_TTS.pitch })} /></Field>
            </div>
            <div className="batch-factory-audio-options batch-factory-audio-options-v13">
              <section className={`batch-factory-audio-option${form.fixedSingleVideo === true ? ' is-disabled' : ''}`}>
                <div><div className="batch-factory-option-title"><b>分镜规划跟随配音</b><HelpTip title="分镜规划跟随配音">先为每本小说按当前正文与配音参数生成临时配音并读取真实时长，再据此规划 VIDEO 数量和每段目标时长。剧情节点、对白/旁白密度优先，不会机械平均切分。</HelpTip></div><small>{form.fixedSingleVideo === true ? '固定开头只生产 VIDEO01，当前不可使用。' : '生成分镜前先按每本书真实配音时长规划 VIDEO。'}</small></div>
                <Switch disabled={form.fixedSingleVideo === true} checked={form.audioPlanningEnabled === true} onChange={audioPlanningEnabled => patch({ audioPlanningEnabled })} />
              </section>
              <section className={`batch-factory-audio-option${form.fixedSingleVideo === true ? ' is-disabled' : ''}`}>
                <div><div className="batch-factory-option-title"><b>合并跟随配音</b><HelpTip title="合并跟随配音">可在视频生成前或生成后开启。系统读取当前书配音时长，并根据各分镜主视频的实际总时长自动计算最终合并倍率，用来修正生成视频产生的时长误差。</HelpTip></div><small>{form.fixedSingleVideo === true ? '固定开头只生产 VIDEO01，当前不可使用。' : '成片阶段按实际视频总时长二次校正倍率。'}</small></div>
                <Switch disabled={form.fixedSingleVideo === true} checked={form.audioMergeEnabled === true} onChange={audioMergeEnabled => patch({ audioMergeEnabled })} />
              </section>
            </div>
          </EngineCard>

          <EngineCard title="辅助生成" description="图片与文本模型用于资产、AI 推理与画面生成。">
            <div className="batch-factory-engine-grid">
              <Field label="图片模型" note="可留空，继承账号默认设置"><Select allowClear loading={modelsLoading} value={form.imageModelId || undefined} options={modelOptions.image} placeholder={modelPlaceholder('image', '图片')} onChange={imageModelId => patch({ imageModelId: imageModelId || '' })} /></Field>
              <Field label="文本模型" note="AI 推理、分类与分镜生成"><Space direction="vertical" size={6} style={{ width: '100%' }}><Select allowClear loading={modelsLoading} value={form.textModelId || undefined} options={modelOptions.text} placeholder={modelPlaceholder('text', '文本')} onChange={textModelId => patch({ textModelId: textModelId || '' })} />{selectedTextModel.state === 'available' ? null : <Tag color={selectedTextModel.state === 'unavailable' ? 'red' : 'default'}>{selectedTextModel.label}</Tag>}</Space></Field>
            </div>
            {modelsError ? <Alert type="warning" showIcon message="模型目录暂不可用" description={modelsError} /> : null}
            {!modelsLoading && !modelsError && !models.length ? <Alert type="warning" showIcon message="个人中心没有已启用模型" description={<Space direction="vertical"><span>请先在个人中心按文本、图片、视频类型新增并启用模型，再返回当前批量作品选择。</span><Button type="link" href="/api-config">前往个人中心配置模型</Button></Space>} /> : null}
          </EngineCard>
        </div>
      },
      {
        key: 'publish',
        label: '发布统一',
        children: <div className="batch-factory-engine-drawer batch-factory-engine-publish-final">
          <Alert type="info" showIcon message="批量默认发布规则" description="这里维护视频管理系统连接、网站配置映射和上传素材规则；不会直接提交小说。真正上传仍从顶部“上传网络”发起，单书配置可以覆盖本页默认值。" />

          <section className="batch-factory-vms-connection-card">
            <div className="batch-factory-vms-main">
              <div className="batch-factory-vms-icon" aria-hidden="true">▣</div>
              <div className="batch-factory-vms-copy">
                <div className="batch-factory-vms-title-row"><b>视频管理系统</b><Tag color={publishSessionReady ? 'green' : 'default'}>{publishSessionReady ? '已登录并验证' : account?.username ? '待重新验证' : '未登录'}</Tag></div>
                <div className="batch-factory-vms-meta"><span>当前账号</span><strong>{account?.username || '尚未登录'}</strong></div>
                <div className="batch-factory-vms-meta"><span>后台地址</span><a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">two.121w.com/tttadmin</a></div>
              </div>
            </div>
            <div className="batch-factory-vms-actions">
              <Button href={TARGET_ADMIN_URL} target="_blank">打开后台</Button>
              <Button onClick={openLogin}>{account?.username ? '更换账号' : '登录并验证'}</Button>
              <Button onClick={selfCheck} loading={checking}>{account?.username ? '重新验证' : '环境自检'}</Button>
            </div>
          </section>

          <div className="batch-factory-publish-two-column">
            <EngineCard title="发布映射" description="同步真实后台配置后，绑定到当前批量作品。">
              <Field label="网站配置档" note="先同步，再选择绑定"><Select allowClear showSearch value={publish.websiteProfileId || undefined} options={(publish.websiteProfiles || []).map(profile => ({ value: profile.id, label: profile.name || profile.id }))} placeholder="同步网站配置档后选择" onChange={websiteProfileId => { const profile = (publish.websiteProfiles || []).find(item => item.id === websiteProfileId) || {}; patch({ publishSettings: { ...publish, websiteProfileId: websiteProfileId || '', versionProfile: profile.name || '' } }); }} /></Field>
              <div className="batch-factory-publish-sync-row">
                <Button disabled={!publishSessionReady} onClick={syncProfiles} loading={syncing === 'profiles'}>同步网站配置档</Button>
                <Button disabled={!publishSessionReady} onClick={syncStyles} loading={syncing === 'styles'}>同步风格目录</Button>
              </div>
              <div className="batch-factory-publish-ai-note"><b>男女频 / 风格 / 标签</b><span>提交时按每本书已保存结果或 AI 判断结果处理，不由批量配置档覆盖。</span></div>
            </EngineCard>

            <EngineCard title="批量默认上传规则" description="这些规则会被单书发布设置覆盖。">
              <Field label="上传视频类型"><Segmented value={publish.uploadVideoType || 'merged'} options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} onChange={uploadVideoType => patch({ publishSettings: { ...publish, uploadVideoType } })} /></Field>
              <div className="batch-factory-publish-rule-list">
                <section><div><b>素材复用</b><small>上传 AI 前贴时复用已选素材</small></div><Switch checked={publish.materialReuse === true} onChange={materialReuse => patch({ publishSettings: { ...publish, materialReuse } })} /></section>
                <section><div><b>水平翻转</b><small>提交时按规则翻转视频素材</small></div><Switch checked={publish.horizontalFlip === true} onChange={horizontalFlip => patch({ publishSettings: { ...publish, horizontalFlip } })} /></section>
                <section><div><b>改文后上传</b><small>使用当前生产内容覆盖配置范围内原文</small></div><Switch checked={form.publishRewriteEnabled === true} onChange={publishRewriteEnabled => patch({ publishRewriteEnabled })} /></section>
              </div>
            </EngineCard>
          </div>

          <EngineCard title="连接与环境状态" description="登录、直接接口和后台页面必须都可用，发布链路才算就绪。">
            <div className="batch-factory-env-status-row">
              <div className={environment?.session ? 'is-ok' : 'is-warn'}><b>登录会话</b><span>{environment ? (environment.session ? '已登录' : '未登录 / 已失效') : publishSessionReady ? '已验证' : '等待检查'}</span></div>
              <div className={environment?.visible?.ok ? 'is-ok' : 'is-warn'}><b>后台页面</b><span>{environment ? (environment.visible?.ok ? '可访问' : environment.session ? '验证未通过' : '未执行') : '等待检查'}</span></div>
              <div className={check121Detail(environment?.environment, '121 直接接口') ? 'is-ok' : 'is-neutral'}><b>直接接口</b><span>{check121Detail(environment?.environment, '121 直接接口') || '等待检查'}</span></div>
              <Button onClick={selfCheck} loading={checking}>重新检查</Button>
            </div>
            {environment?.error ? <Alert type="error" showIcon message="视频管理系统环境自检失败" description={environment.error} /> : null}
            <p className="batch-factory-engine-note">组织归属会在真正“上传网络”时从视频管理系统实时读取并选择；这里不伪装成发布目录设置。</p>
          </EngineCard>
        </div>
      }
    ]} />
  </Modal>

  <Modal
    title={account?.username ? '更换视频管理系统账号' : '登录视频管理系统'}
    open={loginOpen}
    onCancel={() => { setLoginPassword(''); setLoginOpen(false); }}
    onOk={login121}
    confirmLoading={loginBusy}
    okText="登录并验证"
    destroyOnClose
    className="batch-factory-vms-login-modal"
  >
    <Space direction="vertical" size={12} style={{ width: '100%' }}>
      <Alert type="info" showIcon message="账号与小说获取共用" description="登录会话保存在当前用户的共享网站配置中；密码不会写入批量作品设置。" />
      <Input value={loginUsername} autoComplete="username" placeholder="视频管理系统账号" onChange={event => setLoginUsername(event.target.value)} />
      <Input.Password value={loginPassword} autoComplete="current-password" placeholder="视频管理系统密码" onChange={event => setLoginPassword(event.target.value)} />
      <a href={TARGET_ADMIN_URL} target="_blank" rel="noreferrer">打开视频管理系统后台</a>
    </Space>
  </Modal></>;
}
