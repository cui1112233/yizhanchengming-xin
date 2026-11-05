import { Alert, Button, Divider, Input, InputNumber, Modal, Popconfirm, Popover, Segmented, Select, Space, Switch, Tabs, Typography, message} from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { listAvailableModels, modelSelectOptions } from '../../../shared/api/modelCatalog';
import { get121OrganizationOptions, getBatch, listSystemPresetCatalog, saveBookOverride } from '../../../shared/api/batchFactory';
import { getConfig } from '../../../shared/api/config';
import { textToSpeech } from '../../../shared/api/tts';
import { deleteScriptConstraintPrompt, getConstraintPresetTexts, listScriptConstraintPrompts, saveScriptConstraintPrompt, updateScriptConstraintPrompt } from '../../../shared/api/generation';
import { modelOptionsForProvider, normalizeModelProvider, providerForCatalogModel, providerModelSettingsPatch, providerOptionsForModels } from './providerModelBinding';
import { normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { audioDurationFingerprint as runtimeAudioDurationFingerprint, resolveBookProductionText as runtimeResolveBookProductionText } from './batchFactoryRuntimeLogic';
import { batchFactoryPreviewText } from './batchFactoryContentRange';
import { BatchFactoryComicStylePicker } from './BatchFactoryComicStylePicker';
import { BatchFactoryComicPublishSettingsForm } from './BatchFactoryUnifiedSettingsModal';
import { batchFactoryProductionModeVisibility } from './batchFactoryProductionMode';
import { IMAGE_ASPECT_RATIO_OPTIONS } from '../../../shared/imageAspectRatios';

const ENGINE_MODEL_OVERRIDE_FIELDS = ['textModelId', 'imageModelId', 'imageProvider', 'videoModelId', 'videoProvider', 'aspectRatio', 'imageAspectRatio', 'videoAspectRatio', 'videoResolution', 'productionMode', 'assetImagesEnabled', 'storyboardDurationLimit', 'maxVideoDuration', 'audioSource', 'audioPlanningEnabled', 'audioMergeEnabled', 'audioDurationSeconds', 'audioDurationFingerprint', 'tts', 'sensitiveAIEnabled', 'sensitiveAIModels', 'sensitiveAIOnFailure'];
const ENGINE_PUBLISH_OVERRIDE_FIELDS = ['publishRewriteEnabled', 'publishSettings', 'comicPublishSettings'];
const ENGINE_OVERRIDE_FIELDS = [...ENGINE_MODEL_OVERRIDE_FIELDS, ...ENGINE_PUBLISH_OVERRIDE_FIELDS];
const BOOK_OVERRIDE_FIELDS = [...ENGINE_OVERRIDE_FIELDS, 'starredCharacterNames', 'openingEnabled', 'openingCount', 'aiPromptConfig'];
const PUBLISH_CATALOG_FIELDS = new Set(['websiteProfiles', 'websiteStyleCatalog', 'organizations', 'organizationOptions']);
export const AI_REGION_KEYS = new Map([
  ['assets', 'assets'],
  ['constraints', 'constraints'],
  ['video', 'video'],
  ['visual', 'visual']
]);
const REGION_LABELS = {
  engine: '引擎配置',
  assets: '资产设置',
  constraints: '约束设置',
  video: '视频设置',
  visual: '画面设置',
  media: '画面与视频'
};
const CONSTRAINT_LAYERS = [
  ['baseSetup', '基础设定（人物 / 场景）'],
  ['prefix', '画面前缀词'],
  ['quality', '画质约束'],
  ['restriction', '画面限制'],
  ['negative', '负面提示词']
];
const EDITABLE_CONSTRAINT_LAYERS = CONSTRAINT_LAYERS.filter(([category]) => category !== 'baseSetup');
const DEFAULT_ENABLED_CONSTRAINT_CATEGORIES = ['prefix', 'quality', 'restriction', 'negative'];
const DEFAULT_VIDEO_PROMPT = { presetId: 'batch-video-sd', presetName: 'SD 视频提示词', presetSlot: 'batch.video-meta', presetVersion: null, constraintCategory: '' };
const DEFAULT_TTS = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 };
const SMART_UNIFIED_PRESET_ID = 'script-constraint-prefix-smart-unified';
const TTS_VOICE_OPTIONS = [
  { label: '晓晓（女声·温柔）', value: 'zh-CN-XiaoxiaoNeural' },
  { label: '晓辰（女声·知性）', value: 'zh-CN-XiaochenNeural' },
  { label: '云希（男声·清朗）', value: 'zh-CN-YunxiNeural' },
  { label: '云扬（男声·阳光）', value: 'zh-CN-YunyangNeural' },
  { label: '晓伊（女声·甜美）', value: 'zh-CN-XiaoyiNeural' },
  { label: '云健（男声·稳重）', value: 'zh-CN-YunjianNeural' }
];
const TTS_STYLE_OPTIONS = [
  { label: '通用', value: 'general' }, { label: '开心', value: 'cheerful' }, { label: '悲伤', value: 'sad' }, { label: '友好', value: 'friendly' }, { label: '聊天', value: 'chat' }
];

function readAudioDuration(blob) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(blob);
    const audio = new Audio();
    const clear = () => { audio.removeAttribute('src'); URL.revokeObjectURL(url); };
    audio.preload = 'metadata';
    audio.onloadedmetadata = () => { const duration = Number(audio.duration || 0); clear(); duration > 0 && Number.isFinite(duration) ? resolve(duration) : reject(new Error('无法读取配音真实时长')); };
    audio.onerror = () => { clear(); reject(new Error('配音音频无法读取')); };
    audio.src = url;
  });
}
function downloadAudio(blob, filename) {
  const url = URL.createObjectURL(blob); const link = document.createElement('a');
  link.href = url; link.download = filename; link.click(); URL.revokeObjectURL(url);
}
function audioDurationFingerprint(input, tts = {}) {
  const payload = JSON.stringify([String(input || ''), String(tts.voice || ''), String(tts.style || ''), Number(tts.speed || 0), Number(tts.pitch || 0)]);
  let hash = 2166136261;
  for (let index = 0; index < payload.length; index += 1) {
    hash ^= payload.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return `a1-${(hash >>> 0).toString(16)}-${payload.length}`;
}
function HelpTip({ title, children }) {
  return <Popover trigger="click" placement="top" title={title} content={<div className="batch-factory-help-popover">{children}</div>}><button type="button" className="batch-factory-help-button" aria-label={`查看${title}说明`}>?</button></Popover>;
}
function equal(left, right) { return JSON.stringify(left ?? null) === JSON.stringify(right ?? null); }
function mergePromptConfig(base, override) {
  const merged = { ...(base || {}), ...(override || {}) };
  for (const key of ['assets', 'constraints', 'hook', 'originalDirector', 'viralDirector', 'video', 'visual', 'comic']) {
    merged[key] = { ...(base?.[key] || {}), ...(override?.[key] || {}) };
  }
  return merged;
}
function requiredVideoPrompt(value) {
  const selected = value && typeof value === 'object' ? value : {};
  if (!selected.presetId || selected.presetId === 'batch-video-meta') return { ...DEFAULT_VIDEO_PROMPT, enabled: true };
  return { ...DEFAULT_VIDEO_PROMPT, ...selected, enabled: true };
}
function mergePublishSettings(base, override) { return { ...(base || {}), ...(override || {}) }; }
function effectiveValues(batchPatch, bookPatch) {
  const aiPromptConfig = mergePromptConfig(batchPatch?.aiPromptConfig, bookPatch?.aiPromptConfig);
  const values = { ...(batchPatch || {}), ...(bookPatch || {}), aiPromptConfig: { ...aiPromptConfig, video: requiredVideoPrompt(aiPromptConfig.video) }, publishSettings: mergePublishSettings(batchPatch?.publishSettings, bookPatch?.publishSettings), comicPublishSettings: mergePublishSettings(batchPatch?.comicPublishSettings, bookPatch?.comicPublishSettings) };
  return normalizeAudioSourceSettings(values);
}
function selectedMeta(catalog, id) {
  const preset = catalog.find(item => item.id === id);
  return preset ? { presetId: preset.id, presetName: preset.name, presetSlot: preset.slot, presetVersion: preset.version, constraintCategory: preset.constraintCategory || (preset.id === SMART_UNIFIED_PRESET_ID ? 'prefix' : '') } : { presetId: '' };
}
function isConstraintPreset(item, category) {
  return item?.constraintCategory === category || (category === 'prefix' && item?.id === SMART_UNIFIED_PRESET_ID);
}
function presetLabel(preset) { return String(preset?.name || preset?.presetName || preset?.label || preset?.id || '').trim(); }
function optionFor(preset) { return { value: preset.id, label: <span className="batch-factory-preset-option"><span>{presetLabel(preset)}</span><small>v{preset.version || 1}</small></span> }; }
// Preserve the frozen display name for an existing override if its published
// catalog entry is unavailable during a refresh.
function optionsWithCurrent(records, value, fallbackName = '系统预设词') {
  const list = Array.isArray(records) ? records : [];
  const id = String(value?.presetId || '').trim();
  return id && !list.some(item => item.id === id)
    ? [{ id, name: '预设已失效', version: value?.presetVersion || 1, disabled: true }, ...list]
    : list;
}

function normalizeConstraintLayer(value, fallback = {}) {
  const source = ['system', 'personal', 'draft'].includes(value?.source) ? value.source : (fallback?.personalPromptId ? 'personal' : 'system');
  return {
    enabled: value?.enabled === true,
    source,
    presetId: String(value?.presetId || fallback?.presetId || ''),
    presetName: String(value?.presetName || fallback?.presetName || ''),
    presetSlot: String(value?.presetSlot || fallback?.presetSlot || ''),
    presetVersion: Number(value?.presetVersion || fallback?.presetVersion || 0),
    personalPromptId: String(value?.personalPromptId || fallback?.personalPromptId || ''),
    body: String(value?.body || fallback?.body || '')
  };
}

function normalizeConstraintRules(value) {
  const raw = value && typeof value === 'object' ? value : {};
  const { wrapper: _retiredWrapper, ...current } = raw;
  const selections = Array.isArray(current.selections) ? current.selections : [];
  const enabledCategories = Array.isArray(current.enabledCategories) ? current.enabledCategories : DEFAULT_ENABLED_CONSTRAINT_CATEGORIES;
  const layers = Object.fromEntries(EDITABLE_CONSTRAINT_LAYERS.map(([category]) => {
    const fallback = selections.find(item => item?.constraintCategory === category) || {};
    const layer = normalizeConstraintLayer(raw[category], fallback);
    if (enabledCategories.includes(category) && raw[category]?.enabled === undefined) layer.enabled = true;
    return [category, layer];
  }));
  return {
    ...current,
    // 约束模块没有总开关；历史配置里的 enabled 只保留兼容字段，不能再关闭整组约束。
    enabled: true,
    baseSetup: { enabled: current.baseSetup?.enabled !== false },
    ...layers
  };
}

function withDirectorConstraintSelections(value) {
  const rules = normalizeConstraintRules(value);
  const enabledCategories = EDITABLE_CONSTRAINT_LAYERS.filter(([category]) => rules[category].enabled).map(([category]) => category);
  const selections = EDITABLE_CONSTRAINT_LAYERS.flatMap(([category]) => {
    const layer = rules[category];
    if (!layer.enabled || !layer.body.trim()) return [];
    return [{
      presetId: layer.presetId,
      presetName: layer.presetName,
      presetSlot: layer.presetSlot,
      presetVersion: layer.presetVersion,
      personalPromptId: layer.personalPromptId,
      body: layer.body.trim(),
      constraintCategory: category
    }];
  });
  return { ...rules, enabled: true, enabledCategories, selections };
}

export function buildBookOverridePatch(inherited, edited) {
  return Object.fromEntries(BOOK_OVERRIDE_FIELDS
    .filter(key => !equal(inherited?.[key], edited?.[key]))
    .map(key => [key, edited[key]]));
}

function buildRestoreKeys(inherited, edited, currentPatch) {
  return BOOK_OVERRIDE_FIELDS.filter(key => Object.hasOwn(currentPatch || {}, key) && equal(inherited?.[key], edited?.[key]));
}

export function buildBookRegionUpdate(inherited, bookPatch, region, edited, engineTab = 'models') {
  if (region === 'engine' && engineTab === 'media') region = 'media';
  const patch = {};
  const restoreKeys = [];
  if (region === 'engine') {
    if (engineTab === 'publish') {
      if (!equal(inherited?.publishRewriteEnabled, edited?.publishRewriteEnabled)) patch.publishRewriteEnabled = edited?.publishRewriteEnabled;
      else if (Object.hasOwn(bookPatch || {}, 'publishRewriteEnabled')) restoreKeys.push('publishRewriteEnabled');

      const inheritedPublish = inherited?.publishSettings || {};
      const editedPublish = edited?.publishSettings || {};
      const publishOverride = Object.fromEntries(Object.entries(editedPublish)
        .filter(([key]) => !PUBLISH_CATALOG_FIELDS.has(key))
        .filter(([key, value]) => !equal(inheritedPublish[key], value)));
      if (Object.keys(publishOverride).length) patch.publishSettings = publishOverride;
      else if (Object.hasOwn(bookPatch || {}, 'publishSettings')) restoreKeys.push('publishSettings');
      const inheritedComicPublish = inherited?.comicPublishSettings || {};
      const editedComicPublish = edited?.comicPublishSettings || {};
      const comicPublishOverride = Object.fromEntries(Object.entries(editedComicPublish)
        .filter(([key, value]) => !equal(inheritedComicPublish[key], value)));
      if (Object.keys(comicPublishOverride).length) patch.comicPublishSettings = comicPublishOverride;
      else if (Object.hasOwn(bookPatch || {}, 'comicPublishSettings')) restoreKeys.push('comicPublishSettings');
      return { patch, restoreKeys };
    }

    for (const key of ENGINE_MODEL_OVERRIDE_FIELDS) {
      if (!equal(inherited?.[key], edited?.[key])) patch[key] = edited?.[key];
      else if (Object.hasOwn(bookPatch || {}, key)) restoreKeys.push(key);
    }
    // 换开头是统一配置根级的生产开关；单书只存差异，拨回相同值即恢复继承。
    for (const key of ['openingEnabled', 'openingCount']) {
      if (!equal(inherited?.[key], edited?.[key])) patch[key] = edited?.[key];
      else if (Object.hasOwn(bookPatch || {}, key)) restoreKeys.push(key);
    }
    return { patch, restoreKeys };
  }
  const moduleKey = region === 'media' ? 'video' : AI_REGION_KEYS.get(region);
  if (!moduleKey) return { patch, restoreKeys };
  if (region === 'media') {
    if (!equal(inherited?.productionMode, edited?.productionMode)) patch.productionMode = edited?.productionMode;
    else if (Object.hasOwn(bookPatch || {}, 'productionMode')) restoreKeys.push('productionMode');
  }
  if (region === 'assets') {
    if (!equal(inherited?.starredCharacterNames, edited?.starredCharacterNames)) patch.starredCharacterNames = edited?.starredCharacterNames || [];
    else if (Object.hasOwn(bookPatch || {}, 'starredCharacterNames')) restoreKeys.push('starredCharacterNames');
  }
  const rawBookAI = bookPatch?.aiPromptConfig && typeof bookPatch.aiPromptConfig === 'object' ? { ...bookPatch.aiPromptConfig } : {};
  const moduleKeys = region === 'media' ? ['video', 'comic'] : region === 'constraints' ? ['constraints', 'comic'] : region === 'video' ? ['video'] : [moduleKey];
  for (const key of moduleKeys) {
    const inheritedModule = inherited?.aiPromptConfig?.[key] || {};
    const editedModule = edited?.aiPromptConfig?.[key] || {};
    if (equal(inheritedModule, editedModule)) delete rawBookAI[key];
    else rawBookAI[key] = editedModule;
  }
  if (Object.keys(rawBookAI).length) patch.aiPromptConfig = rawBookAI;
  else if (Object.hasOwn(bookPatch || {}, 'aiPromptConfig')) restoreKeys.push('aiPromptConfig');
  return { patch, restoreKeys };
}

// 生效来源只标偏差：继承批次统一配置、或值等于批次统一配置（最近一次保存
// 统一配置时写入本书）时完全不渲染标注，避免每个字段都挂一段“当前生效：…”
// 把真正需要注意的单书偏差淹没。整页的继承说明只在弹窗顶部出现一次。
// explicit 必须由调用方传入“单书 patch 是否真的写了这个键”。表单里的值来自
// effectiveValues(批次 + 单书)，带默认值的字段即使单书没覆盖也会有值，不能靠
// value 是否为空来判断来源。
function FieldSource({ value, inherited, explicit }) {
  if (!explicit) return null;
  const same = JSON.stringify(value ?? null) === JSON.stringify(inherited ?? null);
  if (same) return null;
  return <Popover trigger="click" placement="top" title="单书自定义" content={<div className="batch-factory-help-popover">该项已在本单书单独设置，不会跟随批次统一配置。</div>}><em className="batch-factory-field-source is-custom">自定义</em></Popover>;
}

function InheritedField({ label, field, value, inherited, options, onChange, loading, explicit }) {
  return <div className="batch-factory-engine-field"><span><b>{label}</b><FieldSource value={value} inherited={inherited} explicit={explicit} /></span><Select allowClear loading={loading} value={value || inherited || undefined} options={options} placeholder="未配置" onChange={next => onChange(next || '')} /></div>;
}

function InheritedMultiField({ label, value, inherited, options, onChange, loading, explicit }) {
  const current = Array.isArray(value) ? value : undefined;
  const inheritedValue = Array.isArray(inherited) ? inherited : [];
  return <div className="batch-factory-engine-field"><span><b>{label}</b><FieldSource value={current} inherited={inherited} explicit={explicit} /></span><Select mode="multiple" allowClear loading={loading} value={current ?? inheritedValue} options={options} placeholder="默认继承全局修复模型" onChange={next => onChange(next || [])} /></div>;
}

function InheritedNumberField({ label, value, inherited, min, max, onChange, explicit, precision }) {
  return <label className="batch-factory-engine-field"><span><b>{label}</b><FieldSource value={value} inherited={inherited} explicit={explicit} /></span><InputNumber min={min} max={max} precision={precision} value={value ?? inherited ?? undefined} placeholder="未配置" onChange={next => onChange(next ?? '')} /></label>;
}

function InheritedStoryboardDurationField({ value, inherited, onChange, explicit }) {
  const duration = Number(value ?? inherited) === 15 ? 15 : 10;
  return <div className="batch-factory-engine-field"><span><b>分镜时长</b><HelpTip title="分镜时长">视频提示词按此上限切分。</HelpTip><FieldSource value={value} inherited={inherited} explicit={explicit} /></span><Segmented value={duration} options={[{ value: 10, label: '10s' }, { value: 15, label: '15s' }]} onChange={onChange} /></div>;
}

function InheritedSwitchField({ label, value, inherited, onChange, explicit }) {
  return <label className="batch-factory-engine-field"><span><b>{label}</b><FieldSource value={value} inherited={inherited} explicit={explicit} /></span><Switch checked={value ?? inherited ?? false} onChange={onChange} /></label>;
}
function AudioOptionCard({ title, description, help, value, inherited, onChange, disabled = false }) {
  return <section className={`batch-factory-audio-option${disabled ? ' is-disabled' : ''}`}><div><div className="batch-factory-option-title"><b>{title}</b><HelpTip title={title}>{help || description}</HelpTip></div></div><Switch disabled={disabled} checked={value ?? inherited ?? false} onChange={onChange} /></section>;
}
function ConfigCard({ title, description, children }) {
  return <section className="batch-factory-engine-card"><header><b>{title}</b>{description ? <HelpTip title={title}>{description}</HelpTip> : null}</header><div className="batch-factory-engine-card-body">{children}</div></section>;
}

function PromptSelect({ label, value, options, onChange, disabled, required = false }) {
  const catalog = optionsWithCurrent(options.catalog || [], value, label);
  // 这里不能用 label 包住整个字段：帮助按钮也是 label 的可点击子元素，
  // 浏览器会把点击 Select 当成激活“？”按钮，导致下拉框被 Popover 抢走。
  return <div className="batch-factory-engine-field"><span><b>{label}</b><HelpTip title={label}>从个人中心已发布预设词选择。</HelpTip></span><Select allowClear={!required} disabled={disabled} value={value?.presetId || undefined} options={catalog.map(optionFor)} placeholder={required ? `请选择已发布${label}` : '未选择'} onChange={id => onChange(selectedMeta(catalog, id))} /></div>;
}

function RuleModule({ title, value, onChange, children, required = false, showWhenDisabled = false }) {
  const module = value || {};
  const enabled = required || module.enabled === true;
  return <section className="batch-factory-ai-module">
    <div className="batch-factory-ai-module-head"><div><b>{title}</b><HelpTip title={title}>保存后只覆盖当前小说；所选预设及当前编辑内容会随本书配置保存。</HelpTip></div>{required ? null : <Switch checked={enabled} onChange={enabled => onChange({ ...module, enabled })} />}</div>
    {enabled || showWhenDisabled ? <div className="batch-factory-ai-module-body">{children}</div> : null}
  </section>;
}

function ConstraintCategoryEditor({ category, label, value, systemOptions, personalPrompts, loading, saving, editingPromptId, onChange, onSelectSystem, onSelectPersonal, onSaveDraft, onSaveNamed, onEditPersonal, onDeletePersonal }) {
  const isSystem = value.source === 'system';
  const isSmartUnified = value.presetId === SMART_UNIFIED_PRESET_ID;
  const selectedPersonalPrompt = personalPrompts.find(item => item.id === value.personalPromptId);
  return <div className="batch-factory-constraint-layer">
    <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}><b>{label}</b><Switch checked={value.enabled} onChange={enabled => onChange({ enabled })} /></Space>
    {value.enabled ? <>
      <Segmented block options={[{ label: '系统预设', value: 'system' }, { label: '我的提示词', value: 'personal' }]} value={isSystem ? 'system' : 'personal'} onChange={source => onChange(source === 'system' ? { source: 'system', personalPromptId: '' } : { source: value.personalPromptId ? 'personal' : 'draft', presetId: '' })} style={{ marginTop: 10 }} />
      {isSystem ? <Select allowClear showSearch loading={loading} value={value.presetId || undefined} options={systemOptions.map(optionFor)} placeholder="选择系统预设" onChange={onSelectSystem} style={{ width: '100%', marginTop: 8 }} /> : <>
        <Select allowClear loading={loading} value={value.personalPromptId || undefined} options={personalPrompts.map(item => ({ label: item.name || '未命名个人副本', value: item.id }))} placeholder="选择我的提示词，或直接编辑当前草稿" onChange={onSelectPersonal} style={{ width: '100%', marginTop: 8 }} />
        {selectedPersonalPrompt ? <Space size={8} wrap style={{ marginTop: 8 }}><Button size="small" onClick={() => onEditPersonal(selectedPersonalPrompt)}>编辑所选提示词</Button><Popconfirm title="确认删除该个人提示词？" onConfirm={() => onDeletePersonal(selectedPersonalPrompt.id)}><Button size="small" danger>删除</Button></Popconfirm></Space> : null}
      </>}
      {isSmartUnified ? <Typography.Text type="secondary" style={{ display: 'block', marginTop: 10 }}>智能统一由系统自动分析，原始系统提示词不在这里展示。</Typography.Text> : <>
        <span className="batch-factory-constraint-label">提示词内容</span>
        <Input.TextArea rows={5} value={value.body} placeholder="选择预设后可编辑完整提示词；保存不会修改系统预设。" onChange={event => onChange({ source: 'draft', personalPromptId: '', body: event.target.value })} style={{ marginTop: 6 }} />
        <Space wrap style={{ marginTop: 10 }}><Button loading={saving} onClick={onSaveDraft}>保存当前草稿</Button><Button type="primary" loading={saving} onClick={onSaveNamed}>{editingPromptId ? '保存编辑' : '保存为我的提示词'}</Button></Space>
      </>}
    </> : null}
  </div>;
}

function ConstraintLayers({ value, records, personalPrompts, loading, saving, editingPromptId, onChange, onSelectSystem, onSelectPersonal, onSaveDraft, onSaveNamed, onEditPersonal, onDeletePersonal }) {
  const rules = normalizeConstraintRules(value);
  return <div className="batch-factory-constraint-layers">
    <div className="batch-factory-constraint-layer"><Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}><div><b>基础设定（人物 / 场景）</b><HelpTip title="基础设定（人物 / 场景）">开启后，当前书每个分镜自动带入已提取的人物与场景设定；关闭后不显示资产提示词。</HelpTip></div><Switch checked={rules.baseSetup.enabled} onChange={enabled => onChange({ ...rules, baseSetup: { enabled } })} /></Space></div>
    {EDITABLE_CONSTRAINT_LAYERS.map(([category, label]) => <ConstraintCategoryEditor key={category} category={category} label={label} value={rules[category]} systemOptions={records.filter(item => item.constraintCategory === category)} personalPrompts={personalPrompts[category] || []} loading={loading} saving={saving === category} editingPromptId={editingPromptId === category} onChange={patch => onChange({ ...rules, [category]: { ...rules[category], ...patch } })} onSelectSystem={presetId => onSelectSystem(category, presetId)} onSelectPersonal={promptId => onSelectPersonal(category, promptId)} onSaveDraft={() => onSaveDraft(category)} onSaveNamed={() => onSaveNamed(category)} onEditPersonal={prompt => onEditPersonal(category, prompt)} onDeletePersonal={promptId => onDeletePersonal(category, promptId)} />)}
  </div>;
}

export function BatchFactoryBookSettingsModal({ open, batch, book, activeRegion = 'engine', onClose, onSaved, onOpenBookAssets }) {
  const batchPatch = batch?.settingsState?.patch || {};
  const bookPatch = book?.settingsState?.patch || {};
	const frozenAutomationPreset = batchPatch?.automationPresetSnapshot && typeof batchPatch.automationPresetSnapshot === 'object'
	  ? batchPatch.automationPresetSnapshot
	  : null;
  const region = REGION_LABELS[activeRegion] ? activeRegion : 'engine';
  const inherited = useMemo(() => ({ ...batchPatch, aiPromptConfig: mergePromptConfig(batchPatch.aiPromptConfig, {}), publishSettings: mergePublishSettings(batchPatch.publishSettings, {}), comicPublishSettings: mergePublishSettings(batchPatch.comicPublishSettings, {}) }), [batch?.id, batch?.settingsState?.revision]);
  const [form, setForm] = useState(() => effectiveValues(batchPatch, bookPatch));
  const [models, setModels] = useState([]);
  const [organizations, setOrganizations] = useState([]);
  const [catalog, setCatalog] = useState({ script: [], batch: [] });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [audioBusy, setAudioBusy] = useState(false);
  const [audioBlob, setAudioBlob] = useState(null);
  const [accountTts, setAccountTts] = useState(DEFAULT_TTS);
  const [loadError, setLoadError] = useState('');
  const [personalPrompts, setPersonalPrompts] = useState({});
  const [savingPersonalCategory, setSavingPersonalCategory] = useState('');
  const [editingPersonalPrompt, setEditingPersonalPrompt] = useState({ category: '', id: '', name: '' });
  const [personalPromptNameModal, setPersonalPromptNameModal] = useState({ open: false, category: '', name: '' });
  const [engineTab, setEngineTab] = useState('models');

  useEffect(() => {
    if (open) setForm(effectiveValues(batchPatch, bookPatch));
  }, [open, region, batch?.id, batch?.settingsState?.revision, book?.id, book?.settingsState?.revision]);
  useEffect(() => {
    if (!open) return undefined;
    let active = true;
    setLoading(true); setLoadError('');
    Promise.all([
      listAvailableModels('text'), listAvailableModels('image'), listAvailableModels('video'),
      listSystemPresetCatalog('script'), listSystemPresetCatalog('batch-factory'), getConfig().catch(() => ({ tts: DEFAULT_TTS })), get121OrganizationOptions().catch(() => ({ organizations: [] }))
    ]).then(([text, image, video, scriptResult, batchResult, config, organizationResult]) => {
      if (!active) return;
      setModels([...text, ...image, ...video]);
      setCatalog({ script: Array.isArray(scriptResult?.catalog) ? scriptResult.catalog : [], batch: Array.isArray(batchResult?.catalog) ? batchResult.catalog : [] });
      setAccountTts({ ...DEFAULT_TTS, ...(config?.tts || {}) });
      setOrganizations(Array.isArray(organizationResult?.organizations) ? organizationResult.organizations : []);
    }).catch(error => { if (active) setLoadError(error?.message || '读取模型或已发布预设词失败'); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [open]);
  useEffect(() => {
    if (!open || region !== 'constraints') return;
    EDITABLE_CONSTRAINT_LAYERS.forEach(([category]) => { loadPersonalConstraintPrompts(category); });
  }, [open, region]);

  // 存量批量/单书配置可能只保存了系统约束的 presetId 而没有正文（旧版统一配置选择器漏抓 body）。
  // 打开约束页时按预设号静默补拉一次正文用于回显，避免“下拉有名字、内容框空白”；
  // 只补“系统预设 + 正文为空”的层，个人提示词/草稿不碰，后端生产另有权威兜底。
  useEffect(() => {
    if (!open || region !== 'constraints' || loading) return undefined;
    const initialRules = normalizeConstraintRules((form.aiPromptConfig || {}).constraints);
    const missing = EDITABLE_CONSTRAINT_LAYERS
      .map(([category]) => ({ category, layer: initialRules[category] }))
      .filter(({ layer }) => layer && layer.enabled && layer.source === 'system' && layer.presetId && !String(layer.body || '').trim());
    if (!missing.length) return undefined;
    let alive = true;
    getConstraintPresetTexts(missing.map(({ layer }) => layer.presetId))
      .then(result => {
        if (!alive) return;
        const texts = result?.texts || {};
        setForm(current => {
          const rules = normalizeConstraintRules((current.aiPromptConfig || {}).constraints);
          let changed = false;
          for (const { category, layer } of missing) {
            const body = String(texts[layer.presetId] || '').trim();
            if (body && !String(rules[category]?.body || '').trim()) { rules[category] = { ...rules[category], body }; changed = true; }
          }
          if (!changed) return current;
          return { ...current, aiPromptConfig: { ...(current.aiPromptConfig || {}), constraints: withDirectorConstraintSelections(rules) } };
        });
      })
      .catch(() => {});
    return () => { alive = false; };
    // 只在弹窗打开/切到约束页/目录加载完成/换书时执行，不跟随每次输入。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, region, loading, catalog, book?.id, book?.settingsState?.revision]);

  const modelOptions = useMemo(() => ({
    text: modelSelectOptions(models, 'text'),
    image: modelSelectOptions(models, 'image'),
    video: modelSelectOptions(models, 'video')
  }), [models]);
  const effectiveVideoModelId = form.videoModelId || inherited.videoModelId;
  const effectiveVideoProvider = normalizeModelProvider(form.videoProvider || inherited.videoProvider || 'personal_api', 'personal_api', effectiveVideoModelId);
  const videoProviderOptions = useMemo(() => providerOptionsForModels(models, 'video'), [models]);
  const videoModelOptions = modelOptionsForProvider(models, 'video', effectiveVideoProvider, effectiveVideoModelId, { expandFanKeModels: true });
  const effectiveImageProvider = form.imageProvider || inherited.imageProvider || 'personal_api';
  const imageProviderOptions = useMemo(() => providerOptionsForModels(models, 'image'), [models]);
  const imageModelOptions = modelOptionsForProvider(models, 'image', effectiveImageProvider);
  const scriptExtraction = useMemo(() => catalog.script.filter(item => item.slot === 'script.asset-extraction'), [catalog.script]);

  const comicAssetRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.comic-assets'), [catalog.batch]);
  const characterPromptRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.character-meta'), [catalog.batch]);
  const videoAssetRules = useMemo(() => [
    ...scriptExtraction,
    ...characterPromptRules.filter(item => item.id === 'batch-character-meta')
  ], [scriptExtraction, characterPromptRules]);
  const scenePromptRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.scene-meta'), [catalog.batch]);
  const scriptConstraints = useMemo(() => catalog.script.filter(item => item.kind === 'addon' && EDITABLE_CONSTRAINT_LAYERS.some(([category]) => isConstraintPreset(item, category))), [catalog.script]);
  const videoRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.video-meta'), [catalog.batch]);
  const visualRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.visual-meta'), [catalog.batch]);
  const comicPageRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.comic-page'), [catalog.batch]);
  const comicImageDerivationRules = useMemo(() => catalog.batch.filter(item => item.slot === 'batch.comic-image-derivation'), [catalog.batch]);
  const ai = form.aiPromptConfig || {};
  const visibility = batchFactoryProductionModeVisibility(form.productionMode);
  const productionMode = visibility.mode;
  const assetRules = ai.assets || {};
  const starredCharacterNames = Array.isArray(form.starredCharacterNames) ? form.starredCharacterNames : [];
  const characterOptions = [...new Map([
    ...(book?.assetRecords || []).filter(asset => asset?.kind === 'character').map(asset => [String(asset.name || '').trim(), asset]),
    ...(book?.assets?.characters || []).map(asset => [String(asset.name || '').trim(), asset])
  ].filter(([name]) => name)).values()].map(asset => ({ value: String(asset.name || '').trim(), label: String(asset.name || '').trim() }));
  const publish = form.publishSettings || {};
  const inheritedPublish = inherited.publishSettings || {};
  const tts = { ...DEFAULT_TTS, ...accountTts, ...(form.tts || {}) };
  const constraintRules = normalizeConstraintRules(ai.constraints);
  const moduleKey = AI_REGION_KEYS.get(region);
  const hasBookOverride = region === 'engine'
    ? engineTab === 'media'
      ? Object.hasOwn(bookPatch, 'productionMode') || Object.hasOwn(bookPatch.aiPromptConfig || {}, 'video') || Object.hasOwn(bookPatch.aiPromptConfig || {}, 'comic')
    : engineTab === 'publish'
      ? Object.hasOwn(bookPatch, 'publishRewriteEnabled') || Object.hasOwn(bookPatch, 'publishSettings')
      : ENGINE_MODEL_OVERRIDE_FIELDS.some(key => Object.hasOwn(bookPatch, key)) || Object.hasOwn(bookPatch, 'openingEnabled') || Object.hasOwn(bookPatch, 'openingCount')
    : region === 'assets'
      ? Object.hasOwn(bookPatch, 'starredCharacterNames') || Object.hasOwn(bookPatch.aiPromptConfig || {}, 'assets')
    : region === 'media'
      ? Object.hasOwn(bookPatch.aiPromptConfig || {}, 'video') || Object.hasOwn(bookPatch.aiPromptConfig || {}, 'comic')
      : region === 'video'
      ? Object.hasOwn(bookPatch.aiPromptConfig || {}, 'video')
      : Boolean(moduleKey && Object.hasOwn(bookPatch.aiPromptConfig || {}, moduleKey));

  function patch(next) { setForm(current => ({ ...current, ...next })); }
  function patchTts(next) { patch({ tts: { ...tts, ...next }, audioDurationSeconds: 0, audioDurationFingerprint: '', audioDurationManual: false }); }
  function patchAI(key, next) { patch({ aiPromptConfig: { ...ai, [key]: next } }); }
  function scoped(module) { return { ...module, scope: 'custom', bookIds: book?.id ? [book.id] : [] }; }
  function updateAssetSelection(key, selection) { patchAI('assets', scoped({ ...assetRules, [key]: selection })); }
  function setProductionMode(mode) {
    const nextMode = mode === 'comic' ? 'comic' : 'video';
    if (nextMode === 'comic' && !ai.comic?.presetId) {
      const defaultComicPage = comicPageRules[0];
      if (defaultComicPage) patchAI('comic', scoped({ ...ai.comic, ...selectedMeta(comicPageRules, defaultComicPage.id), enabled: true }));
    }
    patch({ productionMode: nextMode });
  }
  function updateConstraintRules(next) { patchAI('constraints', scoped(withDirectorConstraintSelections(next))); }
  function updateConstraintLayer(category, patch) { updateConstraintRules({ ...constraintRules, [category]: { ...constraintRules[category], ...patch } }); }
  function selectedComicConstraint(category) {
    return (ai.comic?.constraints || []).find(item => item.constraintCategory === category)?.presetId || undefined;
  }
  function comicConstraintOptions(category) {
    return catalog.script.filter(item => item.kind === 'addon' && item.constraintCategory === category).map(optionFor);
  }
  function updateComicConstraint(category, presetId) {
    const comic = ai.comic || {};
    const rest = (comic.constraints || []).filter(item => item.constraintCategory !== category);
    if (!presetId) {
      patchAI('comic', scoped({ ...comic, constraints: rest, enabled: true }));
      return;
    }
    getConstraintPresetTexts([presetId]).then(result => {
      setForm(current => {
        const currentAI = current.aiPromptConfig || {};
        const currentComic = currentAI.comic || {};
        const currentRest = (currentComic.constraints || []).filter(item => item.constraintCategory !== category);
        return {
          ...current,
          aiPromptConfig: {
            ...currentAI,
            comic: scoped({
              ...currentComic,
              constraints: [...currentRest, { ...selectedMeta(catalog.batch, presetId), constraintCategory: category, body: String(result?.texts?.[presetId] || '') }],
              enabled: true
            })
          }
        };
      });
    }).catch(() => message.error('系统预设提示词读取失败，请重试'));
  }

  async function loadPersonalConstraintPrompts(category) {
    try {
      const result = await listScriptConstraintPrompts(category);
      setPersonalPrompts(current => ({ ...current, [category]: Array.isArray(result?.prompts) ? result.prompts : [] }));
    } catch (error) {
      message.error(error?.message || '我的提示词加载失败');
    }
  }

  async function selectSystemConstraint(category, presetId) {
    if (!presetId) {
      updateConstraintLayer(category, { source: 'system', presetId: '', presetName: '', presetSlot: '', presetVersion: 0, body: '' });
      return;
    }
    try {
      const result = await getConstraintPresetTexts([presetId]);
      const preset = scriptConstraints.find(item => item.id === presetId);
      updateConstraintLayer(category, { source: 'system', personalPromptId: '', ...selectedMeta(scriptConstraints, presetId), body: String(result?.texts?.[presetId] || ''), presetName: preset?.name || '' });
    } catch (error) {
      message.error(error?.message || '系统预设提示词读取失败');
    }
  }

  function selectPersonalConstraint(category, promptId) {
    const prompt = (personalPrompts[category] || []).find(item => item.id === promptId);
    if (!prompt) {
      updateConstraintLayer(category, { source: 'draft', personalPromptId: '', presetId: '' });
      return;
    }
    updateConstraintLayer(category, { source: 'personal', presetId: '', presetName: '', presetSlot: '', presetVersion: 0, personalPromptId: prompt.id, body: prompt.body || '' });
  }

  async function savePersonalConstraint(category, name) {
    const body = String(constraintRules[category]?.body || '').trim();
    if (!body) return message.warning('请先填写提示词内容');
    if (name !== null && !String(name || '').trim()) return message.warning('请输入提示词名称');
    setSavingPersonalCategory(category);
    try {
      const result = editingPersonalPrompt.category === category && editingPersonalPrompt.id
        ? await updateScriptConstraintPrompt(editingPersonalPrompt.id, { name, body })
        : await saveScriptConstraintPrompt({ category, name, body });
      const saved = result?.prompt;
      if (!saved) throw new Error('提示词保存失败');
      await loadPersonalConstraintPrompts(category);
      updateConstraintLayer(category, { source: 'personal', presetId: '', presetName: '', presetSlot: '', presetVersion: 0, personalPromptId: saved.id, body: saved.body || body });
      setEditingPersonalPrompt({ category: '', id: '', name: '' });
      setPersonalPromptNameModal({ open: false, category: '', name: '' });
      message.success('已保存我的提示词');
    } catch (error) {
      message.error(error?.message || '提示词保存失败');
    } finally {
      setSavingPersonalCategory('');
    }
  }

  async function deletePersonalConstraint(category, promptId) {
    try {
      await deleteScriptConstraintPrompt(promptId);
      await loadPersonalConstraintPrompts(category);
      if (constraintRules[category]?.personalPromptId === promptId) updateConstraintLayer(category, { source: 'draft', personalPromptId: '', body: '' });
      message.success('已删除我的提示词');
    } catch (error) {
      message.error(error?.message || '删除失败');
    }
  }
  async function measureBookAudio() {
    const input = runtimeResolveBookProductionText(book);
    if (!input) { message.warning('当前选择的小说正文没有可用于配音的内容'); return; }
    setAudioBusy(true);
    try {
      const config = await getConfig();
      const nextTts = { ...DEFAULT_TTS, ...(config?.tts || {}), ...(form.tts || {}) };
      const blob = await textToSpeech({ input, ...nextTts });
      const duration = await readAudioDuration(blob);
      const edited = {
        ...form,
        audioDurationSeconds: Number(duration.toFixed(2)),
        audioDurationFingerprint: runtimeAudioDurationFingerprint(book, nextTts)
      };
      setForm(edited);
      setAudioBlob(blob);
      await persist(edited, false);
      message.success(`已取得当前正文配音真实时长：${duration.toFixed(2)} 秒`);
    } catch (error) {
      message.error(error?.message || '当前正文配音生成失败');
    } finally {
      setAudioBusy(false);
    }
  }
  async function saveOverrideWithLatestRevision(patchValue, restoreKeys) {
    const input = { patch: patchValue, restoreKeys, expectedRevision: Number(book?.revision || 0) };
    try {
      return await saveBookOverride(batch.id, book.id, input);
    } catch (error) {
      if (Number(error?.status) !== 409) throw error;
      const latest = await getBatch(batch.id);
      const latestBatch = latest?.batch || latest;
      const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
      if (!latestBook) throw error;
      return saveBookOverride(batch.id, book.id, { ...input, expectedRevision: Number(latestBook.revision || 0) });
    }
  }
  async function persist(edited, closeWhenSaved = true) {
    if (!batch?.id || !book?.id) return;
    const { patch: patchValue, restoreKeys } = buildBookRegionUpdate(inherited, bookPatch, region, edited, engineTab);
    if (!Object.keys(patchValue).length && !restoreKeys.length) { if (closeWhenSaved) onClose?.(); return; }
    setSaving(true);
    try {
      await saveOverrideWithLatestRevision(patchValue, restoreKeys);
      await onSaved?.();
      if (closeWhenSaved) onClose?.();
    } finally { setSaving(false); }
  }
  function save() { return persist(form); }
  function restoreCurrentRegion() { return persist(inherited); }

  const currentAudioFingerprint = runtimeAudioDurationFingerprint(book, tts);
  const audioDurationStale = Number(form.audioDurationSeconds || 0) > 0
    && String(form.audioDurationFingerprint || '') !== currentAudioFingerprint;

  const mediaBody = <div className="batch-factory-media-setting-stack">
    <RuleModule title="视频生成" value={{ enabled: productionMode === 'video' }} onChange={next => { if (next.enabled) setProductionMode('video'); }}>
      <PromptSelect required label="视频提示词" value={ai.video} options={Object.assign(videoRules.map(optionFor), { catalog: videoRules })} disabled={loading} onChange={selection => patchAI('video', scoped({ ...ai.video, ...selection, enabled: true }))} />
    </RuleModule>
    <RuleModule title="漫画生成" value={{ ...ai.comic, enabled: productionMode === 'comic' }} onChange={next => { if (next.enabled) setProductionMode('comic'); }}>
      <PromptSelect required label="漫画提示词" value={ai.comic} options={Object.assign(comicPageRules.map(optionFor), { catalog: comicPageRules })} disabled={loading} onChange={selection => patchAI('comic', scoped({ ...ai.comic, ...selection, enabled: true }))} />
      <PromptSelect label="图片描述词推导" value={ai.comic?.imageDerivation} options={Object.assign(comicImageDerivationRules.map(optionFor), { catalog: comicImageDerivationRules })} disabled={loading} onChange={selection => patchAI('comic', scoped({ ...ai.comic, imageDerivation: selection, enabled: true }))} />
    </RuleModule>
    {visibility.comic ? <BatchFactoryComicStylePicker batchId={batch?.id} catalog={catalog.batch} value={ai.comic} inherited={!Object.hasOwn(bookPatch.aiPromptConfig || {}, 'comic')} onChange={nextComic => patchAI('comic', scoped(nextComic))} /> : null}
  </div>;
  let body = null;
  if (region === 'engine') {
    const audioEnabled = form.audioPlanningEnabled === true || form.audioMergeEnabled === true;
    const profiles = Array.isArray(publish.websiteProfiles)
      ? publish.websiteProfiles
      : (Array.isArray(inheritedPublish.websiteProfiles) ? inheritedPublish.websiteProfiles : []);
    const organizationOptions = organizations.map(item => ({ value: item.id, label: item.level ? `${item.name}（${item.level}）` : item.name }));
    const hasOrganizationOverride = Object.hasOwn(bookPatch.publishSettings || {}, 'organization');

    body = <Tabs activeKey={engineTab} onChange={setEngineTab} className="batch-factory-book-engine-tabs" items={[
      {
        key: 'models',
        label: '模型配置',
        children: <div className="batch-factory-engine-drawer batch-factory-book-engine-content">
          <Alert type="info" showIcon message={<span>当前书独立模型配置 <HelpTip title="当前书独立模型配置">这里只保存当前小说与批量默认值不同的参数。</HelpTip></span>} />
          {visibility.video ? <ConfigCard title="核心视频引擎" description="视频生成、分镜时长和画幅。"><div className="batch-factory-engine-grid">
            <InheritedField label="视频引擎" field="videoProvider" value={form.videoProvider ? normalizeModelProvider(form.videoProvider, 'personal_api', effectiveVideoModelId) : ''} inherited={effectiveVideoProvider} options={videoProviderOptions} loading={false} explicit={Object.hasOwn(bookPatch, 'videoProvider')} onChange={videoProvider => patch(providerModelSettingsPatch({ models, kind: 'video', provider: videoProvider, modelId: form.videoModelId || inherited.videoModelId, expandFanKeModels: true }))} />
            <InheritedField label="视频模型" field="videoModelId" value={form.videoModelId} inherited={inherited.videoModelId} options={videoModelOptions} loading={effectiveVideoProvider === 'doubao_local_executor' ? false : loading} explicit={Object.hasOwn(bookPatch, 'videoModelId')} onChange={videoModelId => patch({ videoModelId, ...(videoModelId ? { videoProvider: providerForCatalogModel(models.find(model => model.id === videoModelId), effectiveVideoProvider) } : {}) })} />
            <InheritedStoryboardDurationField value={form.storyboardDurationLimit} inherited={inherited.storyboardDurationLimit} explicit={Object.hasOwn(bookPatch, 'storyboardDurationLimit')} onChange={storyboardDurationLimit => patch({ storyboardDurationLimit, maxVideoDuration: storyboardDurationLimit })} />
            <InheritedField label="视频画幅" field="videoAspectRatio" value={form.videoAspectRatio || (form.aspectRatio === '16:9' ? '16:9' : '9:16')} inherited={inherited.videoAspectRatio || (inherited.aspectRatio === '16:9' ? '16:9' : '9:16')} options={[{ value: '16:9', label: '16:9' }, { value: '9:16', label: '9:16' }]} loading={false} explicit={Object.hasOwn(bookPatch, 'videoAspectRatio')} onChange={videoAspectRatio => patch({ videoAspectRatio })} />
            <InheritedField label="视频分辨率" field="videoResolution" value={form.videoResolution || '720p'} inherited={inherited.videoResolution || '720p'} options={[{ value: '480p', label: '480p' }, { value: '720p', label: '720p' }, { value: '1080p', label: '1080p' }]} loading={false} explicit={Object.hasOwn(bookPatch, 'videoResolution')} onChange={videoResolution => patch({ videoResolution })} />
          </div>
          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
            <span><b>换开头</b><HelpTip title="换开头">2 个及以上分镜的书为分镜一生成多个开场变体并分别合成上传；不改这里就跟随批量（批量当前：{inherited.openingEnabled === true ? `开启 · ${inherited.openingCount ?? 4} 条（含原始）` : '关闭'}）。</HelpTip><FieldSource value={form.openingEnabled === true ? `开启 · ${form.openingCount ?? inherited.openingCount ?? 4} 条（含原始）` : '关闭'} inherited={inherited.openingEnabled === true ? `开启 · ${inherited.openingCount ?? 4} 条（含原始）` : '关闭'} explicit={Object.hasOwn(bookPatch, 'openingEnabled') || Object.hasOwn(bookPatch, 'openingCount')} /></span>
            <Space>
              <Switch checked={form.openingEnabled === true} onChange={enabled => patch({ openingEnabled: enabled, openingCount: form.openingCount ?? inherited.openingCount ?? 4 })} />
              <InputNumber min={1} max={8} precision={0} disabled={form.openingEnabled !== true} value={form.openingCount ?? inherited.openingCount ?? 4} onChange={count => patch({ openingCount: count ?? 4 })} addonAfter="条（含原始）" />
            </Space>
          </Space>
          </ConfigCard> : null}

          {visibility.video ? <ConfigCard title="配音与时长" description="完整配音和逐行时长固定参与时间轴；音频来源只决定最终合成使用 TTS 还是视频原声。"><section className="batch-factory-book-audio-content">
            <div className="batch-factory-book-audio-head"><div><b>当前正文配音</b><HelpTip title="当前正文配音">只读取小说正文当前选择范围或已保存的生产正文，不读取整本全文。</HelpTip></div><Space wrap><Button size="small" loading={audioBusy} onClick={measureBookAudio}>生成配音并读取时长</Button>{audioBlob ? <Button size="small" onClick={() => downloadAudio(audioBlob, `${book?.bookId || book?.id || 'book'}-配音.mp3`)}>下载配音</Button> : null}</Space></div>
            <div className="batch-factory-engine-grid">
              <label className="batch-factory-engine-field"><span><b>音频来源</b><HelpTip title="音频来源">默认 TTS；原声保留视频模型自带声音。</HelpTip></span><Segmented block value={normalizeAudioSource(form.audioSource)} options={[{ value: 'tts', label: 'TTS（系统配音）' }, { value: 'original', label: '原声（视频自带声音）' }]} onChange={audioSource => patch(normalizeAudioSourceSettings({ ...form, audioSource }))} /></label>
            </div>
            <div className="batch-factory-audio-options">
              <AudioOptionCard title="完整配音与逐行时长" description="系统固定执行，用于建立原文时间计划。" value={true} inherited={false} onChange={() => {}} />
              <AudioOptionCard disabled={normalizeAudioSource(form.audioSource) === 'original'} title="合并跟随配音" description={normalizeAudioSource(form.audioSource) === 'original' ? '仅最终合成不合入 TTS；配音与逐行时长仍固定执行。' : '成片按真实配音总时长校正倍率。'} value={form.audioMergeEnabled} inherited={inherited.audioMergeEnabled} onChange={audioMergeEnabled => patch({ audioMergeEnabled })} />
            </div>
            <div className="batch-factory-engine-grid">
              <label className="batch-factory-engine-field"><span><b>配音音色</b><HelpTip title="配音音色">当前书 TTS。</HelpTip></span><Select value={tts.voice} options={TTS_VOICE_OPTIONS} onChange={voice => patchTts({ voice })} /></label>
              <label className="batch-factory-engine-field"><span><b>配音风格</b><HelpTip title="配音风格">当前书 TTS。</HelpTip></span><Select value={tts.style} options={TTS_STYLE_OPTIONS} onChange={style => patchTts({ style })} /></label>
              <label className="batch-factory-engine-field"><span><b>语速</b><HelpTip title="语速">改变后需重新读取时长。</HelpTip></span><InputNumber min={0.5} max={2} step={0.1} value={tts.speed} onChange={speed => patchTts({ speed: speed ?? DEFAULT_TTS.speed })} /></label>
              <label className="batch-factory-engine-field"><span><b>音调</b><HelpTip title="音调">改变后需重新读取时长。</HelpTip></span><InputNumber min={-50} max={50} step={1} value={tts.pitch} onChange={pitch => patchTts({ pitch: pitch ?? DEFAULT_TTS.pitch })} /></label>
            </div>
            {audioDurationStale ? <Alert type="warning" showIcon message={<span>配音时长已过期 <HelpTip title="配音时长已过期">当前正文或 TTS 参数已变化，请重新生成配音并读取时长。</HelpTip></span>} /> : null}
            {audioEnabled ? <div className="batch-factory-book-audio-duration"><InheritedNumberField label="配音目标（秒）" value={form.audioDurationSeconds ?? 0} inherited={inherited.audioDurationSeconds ?? 0} explicit={Object.hasOwn(bookPatch, 'audioDurationSeconds')} min={0} max={86400} precision={2} onChange={audioDurationSeconds => patch({ audioDurationSeconds })} /></div> : null}
          </section></ConfigCard> : null}

          <ConfigCard title="辅助生成" description="图片与文本模型用于资产、AI 推理和画面生成。"><div className="batch-factory-engine-grid">
            <InheritedField label="图片生成通道" field="imageProvider" value={form.imageProvider || 'personal_api'} inherited={inherited.imageProvider || 'personal_api'} options={imageProviderOptions} loading={false} explicit={Object.hasOwn(bookPatch, 'imageProvider')} onChange={imageProvider => patch(providerModelSettingsPatch({ models, kind: 'image', provider: imageProvider, modelId: form.imageModelId || inherited.imageModelId }))} />
            <InheritedField label="图片模型" field="imageModelId" value={form.imageModelId} inherited={inherited.imageModelId} options={imageModelOptions} loading={effectiveImageProvider === 'doubao_local_executor' ? false : loading} explicit={Object.hasOwn(bookPatch, 'imageModelId')} onChange={imageModelId => patch({ imageModelId, ...(imageModelId ? { imageProvider: providerForCatalogModel(models.find(model => model.id === imageModelId), effectiveImageProvider) } : {}) })} />
            <InheritedField label="文本模型" field="textModelId" value={form.textModelId} inherited={inherited.textModelId} options={modelOptions.text} loading={loading} explicit={Object.hasOwn(bookPatch, 'textModelId')} onChange={textModelId => patch({ textModelId })} />
            <InheritedField label="图片画幅" field="imageAspectRatio" value={form.imageAspectRatio || form.aspectRatio || '9:16'} inherited={inherited.imageAspectRatio || inherited.aspectRatio || '9:16'} options={IMAGE_ASPECT_RATIO_OPTIONS} loading={false} explicit={Object.hasOwn(bookPatch, 'imageAspectRatio')} onChange={imageAspectRatio => patch({ imageAspectRatio })} />
            <InheritedSwitchField label="生成资产图片" value={form.assetImagesEnabled} inherited={inherited.assetImagesEnabled !== false} explicit={Object.hasOwn(bookPatch, 'assetImagesEnabled')} onChange={assetImagesEnabled => patch({ assetImagesEnabled })} />
          </div></ConfigCard>
          <ConfigCard title="敏感词处理" description="默认继承统一配置；这里可单独覆盖当前书。AI 修复失败会降级到规则/打码并等待人工确认。">
            <InheritedSwitchField label="AI 修复" value={form.sensitiveAIEnabled} inherited={inherited.sensitiveAIEnabled !== false} explicit={Object.hasOwn(bookPatch, 'sensitiveAIEnabled')} onChange={sensitiveAIEnabled => patch({ sensitiveAIEnabled })} />
            <InheritedMultiField label="修复模型" value={form.sensitiveAIModels} inherited={inherited.sensitiveAIModels} options={modelOptions.text} loading={loading} explicit={Object.hasOwn(bookPatch, 'sensitiveAIModels')} onChange={sensitiveAIModels => patch({ sensitiveAIModels })} />
            <label className="batch-factory-engine-field"><span><b>失败时策略</b></span><Segmented block value={form.sensitiveAIOnFailure || inherited.sensitiveAIOnFailure || 'review'} options={[{ value: 'review', label: '保留原文待确认' }, { value: 'fail', label: '判为失败' }]} onChange={sensitiveAIOnFailure => patch({ sensitiveAIOnFailure })} /></label>
          </ConfigCard>
        </div>
      },
      { key: 'media', label: '画面与视频', children: mediaBody },
      {
        key: 'publish',
        label: '发布统一',
        children: productionMode === 'comic' ? <div className="batch-factory-engine-drawer batch-factory-book-engine-content">
          <BatchFactoryComicPublishSettingsForm value={form} onChange={patch} singleBook />
        </div> : <div className="batch-factory-engine-drawer batch-factory-book-engine-content">
          <Alert type="info" showIcon message="当前书独立发布配置" description="发布目录仍来自账号和批量作品；这里只保存当前书的差异值。" />
          <ConfigCard title="发布映射" description="网站风格继续读取本书 AI 判断结果。">
            <InheritedField label="网站配置档" value={publish.websiteProfileId} inherited={inheritedPublish.websiteProfileId} options={profiles.map(profile => ({ value: profile.id, label: profile.name || profile.id }))} loading={false} explicit={Object.hasOwn(bookPatch.publishSettings || {}, 'websiteProfileId')} onChange={websiteProfileId => { const profile = profiles.find(item => item.id === websiteProfileId) || {}; patch({ publishSettings: { ...publish, websiteProfileId, versionProfile: profile.name || '' } }); }} />
            <InheritedField label={hasOrganizationOverride ? '组织归属（当前书覆盖）' : '组织归属（继承批量组织归属）'} value={publish.organization} inherited={inheritedPublish.organization} options={organizationOptions} loading={loading} explicit={Object.hasOwn(bookPatch.publishSettings || {}, 'organization')} onChange={organization => patch({ publishSettings: { ...publish, organization } })} />
            <Alert type="info" showIcon message="男女频、风格和标签属于当前书" description="这里不会覆盖书级 AI 判断结果。" />
          </ConfigCard>
          <ConfigCard title="上传规则" description="只影响当前小说。">
            <label className="batch-factory-engine-field"><span><b>上传视频类型</b><HelpTip title="上传视频类型">当前书发布策略。</HelpTip></span><Segmented value={publish.uploadVideoType || 'merged'} options={[{ value: 'merged', label: '合并成品' }, { value: 'individual', label: '独立 VIDEO' }]} onChange={uploadVideoType => patch({ publishSettings: { ...publish, uploadVideoType } })} /></label>
            <div className="batch-factory-engine-grid">
              <label className="batch-factory-engine-field"><span><b>素材使用</b><HelpTip title="素材使用">当前书发布策略。</HelpTip></span><Select value={publish.materialReuse === true ? 'reuse' : 'no_reuse'} options={[{ value: 'no_reuse', label: '不复用' }, { value: 'reuse', label: '复用' }]} onChange={value => patch({ publishSettings: { ...publish, materialReuse: value === 'reuse' } })} /></label>
              <label className="batch-factory-engine-field"><span><b>水平翻转</b><HelpTip title="水平翻转">当前书发布策略。</HelpTip></span><Select value={publish.horizontalFlip === true ? 'flip' : 'no_flip'} options={[{ value: 'no_flip', label: '不翻转' }, { value: 'flip', label: '翻转' }]} onChange={value => patch({ publishSettings: { ...publish, horizontalFlip: value === 'flip' } })} /></label>
            </div>
            <InheritedSwitchField label="改文后上传" value={form.publishRewriteEnabled} inherited={inherited.publishRewriteEnabled} explicit={Object.hasOwn(bookPatch, 'publishRewriteEnabled')} onChange={publishRewriteEnabled => patch({ publishRewriteEnabled })} />
          </ConfigCard>
        </div>
      }
    ]} />;
  } else if (region === 'assets') {
    body = <><Alert type="info" showIcon message="当前书的人物场景预设" description="人物、场景、道具的实际 Prompt 与图片在本书这一行的“添加角色 / 添加场景 / 添加道具”中维护；这里选择这本书调用的提取与推理预设。" />
      <RuleModule title="资产设置" value={assetRules} onChange={next => patchAI('assets', scoped(next))}>
        <PromptSelect label={visibility.comic ? '漫画资产提示词' : '人物场景道具提取提示词'} value={visibility.comic ? ai.comic?.assets : assetRules.extraction} options={Object.assign((visibility.comic ? comicAssetRules : videoAssetRules).map(optionFor), { catalog: visibility.comic ? comicAssetRules : videoAssetRules })} disabled={loading} onChange={selection => visibility.comic ? patchAI('comic', scoped({ ...ai.comic, assets: selection, enabled: true })) : updateAssetSelection('extraction', selection)} />
        <label className="batch-factory-engine-field"><span><b>星标人物聚焦</b><HelpTip title="星标人物聚焦">与公网私人定制一致：只作为导演分镜的剧情与镜头聚焦变量，不等同于分镜资产灰显。</HelpTip></span><Select mode="multiple" allowClear value={starredCharacterNames} options={characterOptions} placeholder={characterOptions.length ? '选择已提取的人物' : '请先提取人物资产'} onChange={starredCharacterNames => patch({ starredCharacterNames })} /></label>
        <Button onClick={() => { onClose?.(); onOpenBookAssets?.(book); }}>维护当前书人物场景预设</Button>
      </RuleModule>
    </>;
  } else if (region === 'constraints') {
    body = <>{visibility.video ? <RuleModule required title="视觉约束" value={constraintRules} onChange={updateConstraintRules}>
      <ConstraintLayers value={constraintRules} records={scriptConstraints} personalPrompts={personalPrompts} loading={loading} saving={savingPersonalCategory} editingPromptId={editingPersonalPrompt.category} onChange={updateConstraintRules} onSelectSystem={selectSystemConstraint} onSelectPersonal={selectPersonalConstraint} onSaveDraft={category => savePersonalConstraint(category, null)} onSaveNamed={category => setPersonalPromptNameModal({ open: true, category, name: editingPersonalPrompt.category === category ? editingPersonalPrompt.name : '' })} onEditPersonal={(category, prompt) => { setEditingPersonalPrompt({ category, id: prompt.id, name: prompt.name || '' }); updateConstraintLayer(category, { source: 'draft', presetId: '', personalPromptId: '', body: prompt.body || '' }); }} onDeletePersonal={deletePersonalConstraint} />
    </RuleModule> : null}
    {visibility.comic ? <section className="batch-factory-engine-card"><header><b>漫画图片约束</b><HelpTip title="漫画图片约束">统一配置会自动继承到当前书；在这里改动后只覆盖当前书。</HelpTip></header><div className="batch-factory-engine-card-body batch-factory-engine-grid">
      <label className="batch-factory-engine-field"><span><b>画质约束</b></span><Select allowClear value={selectedComicConstraint('comic_quality')} options={comicConstraintOptions('comic_quality')} placeholder="未选择" onChange={presetId => updateComicConstraint('comic_quality', presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>画面限制</b></span><Select allowClear value={selectedComicConstraint('comic_restriction')} options={comicConstraintOptions('comic_restriction')} placeholder="未选择" onChange={presetId => updateComicConstraint('comic_restriction', presetId)} /></label>
      <label className="batch-factory-engine-field"><span><b>负面提示词</b></span><Select allowClear value={selectedComicConstraint('comic_negative')} options={comicConstraintOptions('comic_negative')} placeholder="未选择" onChange={presetId => updateComicConstraint('comic_negative', presetId)} /></label>
    </div></section> : null}</>;
  } else if (region === 'media') {
    body = mediaBody;
  } else if (region === 'video') {
    body = visibility.video ? <RuleModule required title="视频提示词" value={ai.video} onChange={next => patchAI('video', scoped({ ...next, enabled: true }))}>
      <PromptSelect required label="视频提示词" value={ai.video} options={Object.assign(videoRules.map(optionFor), { catalog: videoRules })} disabled={loading} onChange={selection => patchAI('video', scoped({ ...ai.video, ...selection, enabled: true }))} />
    </RuleModule> : null;
  } else {
    body = visibility.video ? <RuleModule title="画面设置" value={ai.visual} onChange={next => patchAI('visual', scoped(next))}>
      <PromptSelect label="画面提示词" value={ai.visual} options={Object.assign(visualRules.map(optionFor), { catalog: visualRules })} disabled={loading} onChange={selection => patchAI('visual', scoped({ ...ai.visual, ...selection }))} />
    </RuleModule> : null;
  }

  return <><Modal
    title={book?.title ? `单书配置 · ${book.title} · ${REGION_LABELS[region]}` : `单书配置 · ${REGION_LABELS[region]}`}
    open={open}
    onCancel={onClose}
    width={region === 'engine' ? 860 : 920}
    className={`shuihuo-engine-modal batch-factory-engine-modal${region === 'engine' ? ' batch-factory-book-engine-modal' : ''}`}
    footer={<Space>{hasBookOverride ? <Button danger disabled={saving} onClick={restoreCurrentRegion}>恢复{region === 'engine' ? (engineTab === 'publish' ? '批量发布配置' : engineTab === 'media' ? '批量画面与视频配置' : '批量模型配置') : '作品配置'}</Button> : null}<Button onClick={onClose}>取消</Button><Button type="primary" loading={saving} onClick={save}>保存{region === 'engine' ? (engineTab === 'publish' ? '当前书发布配置' : engineTab === 'media' ? '当前书画面与视频配置' : '当前书模型配置') : '当前书覆盖'}</Button></Space>}
  >
		{frozenAutomationPreset?.id ? <Alert type="info" showIcon message={`继承统一配置：${frozenAutomationPreset.name || frozenAutomationPreset.id} · v${frozenAutomationPreset.version || 1}`} description="当前书未保存覆盖时，始终使用该批次冻结的统一配置。" /> : null}
    <div className="batch-factory-inherit-hint">整页默认继承批次统一配置；改动哪个字段，哪个字段就变成“自定义”，保存后不会跟随批次。</div>
    {loadError ? <Alert type="warning" showIcon message="配置目录读取失败" description={loadError} /> : null}
    {region !== 'engine' ? <Divider orientation="left">{REGION_LABELS[region]}</Divider> : null}
    {body}
  </Modal><Modal title={editingPersonalPrompt.id ? '编辑我的提示词' : '保存为我的提示词'} open={personalPromptNameModal.open} onCancel={() => { setPersonalPromptNameModal({ open: false, category: '', name: '' }); setEditingPersonalPrompt({ category: '', id: '', name: '' }); }} onOk={() => savePersonalConstraint(personalPromptNameModal.category, personalPromptNameModal.name.trim())} okText="保存" confirmLoading={savingPersonalCategory === personalPromptNameModal.category}><Input autoFocus value={personalPromptNameModal.name} placeholder="提示词名称" onChange={event => setPersonalPromptNameModal(current => ({ ...current, name: event.target.value }))} /></Modal></>;
}
