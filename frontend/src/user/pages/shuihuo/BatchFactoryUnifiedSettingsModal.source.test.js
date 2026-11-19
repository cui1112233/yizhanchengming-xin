import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const sourcePath = path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx');
const source = fs.readFileSync(sourcePath, 'utf8');
const stylesheet = fs.readFileSync(path.join(directory, '..', 'shuihuo-production.css'), 'utf8');

test('unified configuration owns models, AI reasoning and publication in one editable patch', () => {
  assert.match(source, /模型配置/);
  assert.match(source, /BatchFactoryPublishSettingsForm/);
  assert.match(source, /AI 推理/);
  assert.match(source, /发布统一/);
  assert.match(source, /BatchFactoryEngineSettingsForm/);
  assert.match(source, /BatchFactoryAiReasoningForm/);
  assert.match(source, /onSaved\(draftPatch\)/);
});

test('keeps image and video media settings independent in unified configuration', () => {
  assert.match(source, /图片画幅/);
  assert.match(source, /视频画幅/);
  assert.match(source, /视频分辨率/);
  assert.match(source, /imageAspectRatio/);
  assert.match(source, /videoAspectRatio/);
  assert.match(source, /videoResolution/);
  assert.match(source, /视频不会跟随参考图画幅/);
});

test('keeps the full video management publication controls inside unified configuration', () => {
  assert.match(source, /function BatchFactoryPublishSettingsForm/);
  assert.match(source, /视频管理系统/);
  assert.match(source, /同步网站配置档/);
  assert.match(source, /同步风格目录/);
  assert.match(source, /websiteProfileId/);
  assert.match(source, /horizontalFlip/);
  assert.match(source, /连接与环境状态/);
});

test('shows simple comic publication defaults instead of video publication controls in comic mode', () => {
  assert.match(source, /function BatchFactoryComicPublishSettingsForm/);
  assert.match(source, /comicPublishSettings/);
  assert.match(source, /producerAdminName/);
  assert.match(source, /audioCategoryId/);
  assert.match(source, /workMatchMode/);
  assert.match(source, /制作者管理员账号/);
	assert.match(source, /启用中的音频分类/);
	assert.match(source, /audioCategoryName/);
	assert.match(source, /启用中的音频分类/);
	assert.match(source, /音频分类 ID（必填）/);
	assert.match(source, /audioCategorySelectionMode/);
	assert.match(source, /发布时按每本书的男女频和风格自动匹配并随机选择音频/);
	assert.doesNotMatch(source, /没有配套音频时可以不填/);
  assert.match(source, /按书名自动匹配/);
  assert.match(source, /productionMode === 'comic'\s*\? <BatchFactoryComicPublishSettingsForm/);
  assert.match(source, /: <BatchFactoryPublishSettingsForm/);
});

test('loads server-managed comic producer accounts into a searchable selector', () => {
  const start = source.indexOf('export function BatchFactoryComicPublishSettingsForm');
  const end = source.indexOf('export function BatchFactoryAiReasoningForm');
  const comicPublishForm = source.slice(start, end);
  assert.match(comicPublishForm, /getN8ComicPublishStatus/);
  assert.match(comicPublishForm, /producerAccounts/);
  assert.match(comicPublishForm, /<Select/);
  assert.match(comicPublishForm, /showSearch/);
  assert.match(comicPublishForm, /账号目录尚未配置/);
  assert.doesNotMatch(comicPublishForm, /<Input\s+[\s\S]*?producerAdminName/);
});

test('stores a unified organization selection independently from AI classification fields', () => {
  assert.match(source, /get121OrganizationOptions/);
  assert.match(source, /<b>组织归属<\/b>/);
  assert.match(source, /publishSettings: \{ \.\.\.publish, organization/);
  assert.doesNotMatch(source, /<b>男女频 \/ 风格 \/ 标签<\/b>/);
});

test('saves the unified configuration as one settings patch instead of nesting patch.patch', () => {
  assert.doesNotMatch(source, /onSaved\(\{ patch: draftPatch, expectedRevision:/);
  assert.match(source, /onSaved\(draftPatch\)/);
});

test('opens legacy nested unified settings as the original editable configuration', () => {
  assert.match(source, /const normalizeLegacyPatch = value =>/);
  assert.match(source, /setDraftPatch\(normalizeLegacyPatch\(batch\?\.settingsState\?\.patch\)\)/);
});

test('does not discard an open unified-configuration draft when the batch refreshes', () => {
  const resetEffect = source.match(/useEffect\(\(\) => \{\s*if \(open\) setDraftPatch\(normalizeLegacyPatch\(batch\?\.settingsState\?\.patch\)\);\s*\}, \[[^\]]+\]\);/)?.[0] || '';
  assert.doesNotMatch(resetEffect, /batch\?\.settingsState\?\.revision/);
  assert.match(resetEffect, /\[open, batch\?\.id\]/);
});

test('unified configuration loads every typed model without passing Array map indexes as request functions', () => {
  assert.match(source, /listAvailableModels\('text'\)/);
  assert.match(source, /listAvailableModels\('image'\)/);
  assert.match(source, /listAvailableModels\('video'\)/);
  assert.doesNotMatch(source, /\.map\(listAvailableModels\)/);
  assert.match(source, /const \[modelsLoading, setModelsLoading\] = useState\(false\)/);
  assert.match(source, /const \[modelsError, setModelsError\] = useState\(''\)/);
  assert.doesNotMatch(source, /listAvailableModels\('image'\)\.catch\(\(\) => \[\]\)/);
  assert.match(source, /模型目录暂不可用/);
});

test('automation preset manager changes the unified draft only after confirmation', () => {
  assert.match(source, /自动化预设/);
  assert.match(source, /保存为新预设/);
  assert.match(source, /重命名所选预设/);
  assert.match(source, /删除所选预设/);
  assert.match(source, /确认载入此预设/);
  assert.match(source, /onLoad\(clonePresetConfig\)/);
  assert.doesNotMatch(source, /setDraftPatch\(.*saveBatchSettings/);
});

test('automation preset manager can edit the selected preset with the current draft', () => {
  assert.match(source, /编辑所选预设/);
  assert.match(source, /把当前统一配置草稿覆盖写入该预设，名称不变/);
  // 覆盖保存会把当前批次的生产模式写进预设（withAutomationPresetMode），
  // 让预设与漫画/视频模式绑定。
  assert.match(source, /updateAutomationPreset\(selectedPreset\.id, \{ name: selectedPreset\.name, config: withAutomationPresetMode\(normalizeBatchAutoAudioConfig\(value\), productionMode\), expectedVersion: selectedPreset\.version \}\)/);
  assert.match(source, /预设刚被其他窗口修改过/);
});

test('unified configuration does not describe production settings as frozen preset metadata', () => {
  assert.doesNotMatch(source, /automationPresetSnapshot/);
  assert.doesNotMatch(source, /本批次已冻结自动化预设/);
});

test('shows the precise 121 session failure and never reports a failed verification as a completed login', () => {
  assert.doesNotMatch(source, /Boolean\(result\?\.ok && \(result\?\.checks \|\| \[\]\)\.some/);
  assert.match(source, /const hasSession = result => Boolean\(\(result\?\.checks \|\| \[\]\)\.some/);
  assert.match(source, /const sessionDetail = result =>/);
  assert.match(source, /SESSION_CHECK_NAMES\.includes\(check\.name\)/);
  assert.match(source, /sessionDetail\(environment\?\.environment\)/);
  assert.match(source, /const checked = await selfCheck\(\);/);
  assert.match(source, /if \(!checked\?\.session \|\| !checked\.visible\?\.ok\)/);
});

test('marks the video-management session stale and refreshes organizations after login', () => {
  assert.match(source, /const markPublishSessionExpired = error =>/);
  assert.match(source, /session: false, visible: \{ ok: false, error: detail \}/);
  assert.match(source, /const refreshOrganizations = async \(\) =>/);
  assert.match(source, /const organizations = await refreshOrganizations\(\);/);
});

test('treats the legacy novel-fetch browser session as the same shared 121 login', () => {
  assert.match(source, /SESSION_CHECK_NAMES\s*=\s*\[\s*'视频管理系统登录会话',\s*'121 后台登录会话',\s*'目标站登录会话'\s*\]/);
});

test('unified configuration has the same mutually exclusive comic media settings as a single book', () => {
  assert.match(source, /export function BatchFactoryMediaSettingsForm/);
  assert.match(source, /const productionMode = value\.productionMode === 'comic' \? 'comic' : 'video';/);
  assert.match(source, /title="漫画生成"/);
  assert.match(source, /batch\.comic-page/);
  assert.match(source, /batch\.comic-image-derivation/);
  assert.match(source, /\{ key: 'media', label: '画面与视频', children: <BatchFactoryMediaSettingsForm value=\{draftPatch\} onChange=\{setDraftPatch\} active=\{open\} \/> \}/);
});

test('unified configuration exposes both local-executor provider selectors', () => {
  assert.match(source, /from '.\/providerModelBinding'/);
  assert.match(source, /图片生成通道/);
  assert.match(source, /视频引擎/);
  assert.match(source, /value\.imageProvider/);
  assert.match(source, /value\.videoProvider/);
  assert.match(source, /providerOptionsForModels\(models, 'image'\)/);
  assert.match(source, /providerOptionsForModels\(models, 'video'\)/);
});

test('local Doubao video provider exposes the executor-supported model list', () => {
  for (const [model, description] of [
    ['Seedance 2.5', '旗舰视频标杆 · 5 倍消耗'],
    ['Seedance 2.0', '进阶画面表现 · 2 倍消耗'],
    ['Seedance 2.0 Fast', '快速出片选择'],
    ['Seedance 2.0 Mini', '日常生成使用']
  ]) {
    assert.match(source, /reconcileProviderModel/);
    assert.match(source, /modelOptionsForProvider/);
  }
  assert.match(source, /modelOptionsForProvider\(models, 'video', value\.videoProvider\)/);
  assert.match(source, /reconcileProviderModel\(\{ models, kind: 'video'/);
});

test('local Doubao image provider exposes the executor-supported model list', () => {
  for (const [model, description] of [
    ['Seedream 5.0 Flash', '专业快速'],
    ['Seedream 5.0 Pro', '专业出图 · 4 倍消耗'],
    ['Seedream 4.5', '日常生成'],
    ['Seedream 4.0', '基础生图']
  ]) {
    assert.match(source, /reconcileProviderModel/);
    assert.match(source, /modelOptionsForProvider/);
  }
  assert.match(source, /modelOptionsForProvider\(models, 'image', value\.imageProvider\)/);
  assert.match(source, /reconcileProviderModel\(\{ models, kind: 'image'/);
  assert.doesNotMatch(source, /豆包本地版/);
});

test('automation preset dialog uses the unified button treatment', () => {
  assert.match(source, /className="batch-factory-automation-preset-modal"/);
  assert.match(stylesheet, /BF_AUTOMATION_PRESET_BUTTONS_BEGIN[\s\S]*batch-factory-automation-preset-modal[\s\S]*box-shadow/);
});

test('opens the shared login form when rechecking finds an expired session', () => {
  const selfCheck = source.slice(source.indexOf('const selfCheck'), source.indexOf('useEffect', source.indexOf('const selfCheck')));
  assert.match(selfCheck, /openLoginForm\(\)/);
});

test('unified engine settings write the opening variants toggle and count at the patch root', () => {
  assert.match(source, /换开头/);
  assert.match(source, /openingEnabled/);
  assert.match(source, /openingCount/);
  assert.match(source, /min=\{1\}/);
  assert.match(source, /max=\{8\}/);
  assert.match(source, /disabled=\{(?:value|config)\.openingEnabled !== true\}/);
  assert.match(source, /条（含原始）/);
  assert.match(source, /(?:patch|onChange)\(\{\s*(?:\.\.\.config,\s*)?openingEnabled: enabled, openingCount: (?:value|config)\.openingCount \?\? 4 \}\)/);
  assert.match(source, /(?:patch|onChange)\(\{\s*(?:\.\.\.config,\s*)?openingCount: count \?\? 4 \}\)/);
  // 换开头仍是根级生产开关，只是展示在 AI 推理的资产设置卡中。
  assert.doesNotMatch(source, /openingConfig/);
  assert.doesNotMatch(source, /onOpeningChange/);
});

test('gives visible feedback when publish verification succeeds or fails', () => {
  assert.match(source, /message\.error\(detail\)/);
  assert.match(source, /message\.success\('重新验证完成，组织目录已更新'\)/);
  assert.match(source, /loading=\{checking\}/);
});

test('places opening variants under AI reasoning asset settings with a compact help tip', () => {
  const assetStart = source.indexOf('<section className="batch-factory-engine-card"><header><b>资产设置</b>');
  const assetEnd = source.indexOf('<section className="batch-factory-engine-card"><header><b>约束设置</b>');
  const modelStart = source.indexOf('<section className="batch-factory-engine-card"><header><b>模型配置</b>');
  const modelEnd = source.indexOf("{sections.includes('sensitive')");
  assert.match(source.slice(assetStart, assetEnd), /换开头/);
  assert.match(source.slice(assetStart, assetEnd), /batch-factory-help-button/);
  assert.doesNotMatch(source.slice(modelStart, modelEnd), /换开头/);
  assert.match(source, /batch-factory-option-title/);
  assert.match(source, /batch-factory-unified-settings-title/);
});

test('keeps configuration explanations behind question-mark tips and removes fixed opening', () => {
  assert.match(source, /batch-factory-help-button/);
  assert.doesNotMatch(source, /固定开头/);
  assert.match(source, /<HelpTip title="模型配置">/);
  assert.match(source, /<HelpTip title="跟随配音">/);
});

test('unified AI reasoning keeps one asset extraction prompt and preserves video prompt selection', () => {
  assert.match(source, /value=\{comic\.assets\?\.presetId \|\| assets\.extraction\?\.presetId \|\| undefined\}/);
  assert.match(source, /bySlots\(\['script\.asset-extraction', 'batch\.comic-assets'\]\)/);
  assert.doesNotMatch(source, /<b>人物提示词<\/b>/);
  assert.doesNotMatch(source, /<b>场景提示词<\/b>/);
  assert.match(source, /noPromptOption = \{ value: '__none__', label: '无' \}/);
  assert.match(source, /promptOptions = options => \[noPromptOption, \.\.\.options\]/);
  assert.match(source, /normalizePromptId = presetId => presetId === '__none__'/);
  assert.match(source, /placeholder="选择提示词或无"/);
  assert.match(source, /video: \{ \.\.\.video, \.\.\.presetValue\(catalog\.find\(item => item\.id === normalizedId\)\), enabled: Boolean\(normalizedId\) \}/);
});

test('unified AI reasoning keeps one selector per setting and includes comic presets in it', () => {
  assert.match(source, /bySlots\(\['script\.asset-extraction', 'batch\.comic-assets'\]\)/);
  assert.match(source, /byConstraint\(\['quality', 'comic_quality'\]\)/);
  assert.match(source, /byConstraint\(\['restriction', 'comic_restriction'\]\)/);
  assert.match(source, /byConstraint\(\['negative', 'comic_negative'\]\)/);
  assert.match(source, /bySlots\(\['batch\.visual-meta', 'batch\.comic-page'\]\)/);
  assert.doesNotMatch(source, /<b>漫画资产提示词<\/b>/);
  assert.doesNotMatch(source, /<b>漫画画质约束<\/b>/);
  assert.doesNotMatch(source, /<b>漫画画面限制<\/b>/);
  assert.doesNotMatch(source, /<b>漫画负面提示词<\/b>/);
  assert.doesNotMatch(source, /<b>漫画提示词<\/b>/);
});

test('explains that the video prompt preset controls rules while the video model controls execution', () => {
  assert.match(source, /不决定导演分镜或提示词内容/);
  assert.match(source, /决定导演分镜和最终 VIDEO Prompt 的规则与格式/);
  assert.match(source, /与上面的具体视频模型相互独立/);
});

test('keeps the unified video prompt selector independent from its help tip', () => {
  assert.match(source, /<div className="batch-factory-engine-field"><span><b>视频提示词<\/b><HelpTip/);
});

test('keeps model dropdowns outside the label that owns the help-tip button', () => {
  // label 的受控元素若先命中“？”按钮，点击下拉会激活该按钮并让弹层失焦收起（闪回）。
  // 修复模型（多选）、视频模型（单选）都必须用 div 包字段。
  assert.doesNotMatch(source, /<label className="batch-factory-engine-field"><span><b>(?:修复模型|视频模型)<\/b><HelpTip[^>]*>[^<]*<\/HelpTip><\/span><Select/);
  assert.match(source, /<div className="batch-factory-engine-field"><span><b>修复模型<\/b><HelpTip[\s\S]*?<Select mode="multiple"/);
  assert.match(source, /<div className="batch-factory-engine-field"><span><b>视频模型<\/b><HelpTip[\s\S]*?<Select/);
});

test('keeps the automation gear beside the unified configuration label', () => {
  assert.match(source, /<div className="batch-factory-unified-settings-title"><BatchFactoryAutomationPresetManager value=\{draftPatch\} productionMode=\{productionMode\} onLoad=\{setDraftPatch\} \/><span>统一配置<\/span><\/div>/);
});

test('makes selected segmented options unmistakable in the unified settings modal', () => {
  assert.match(source, /className="batch-factory-unified-settings-modal"/);
  const selectedRule = stylesheet.match(/\.batch-factory-unified-settings-modal \.ant-segmented-item-selected\{[^}]+\}/)?.[0] || '';
  assert.match(selectedRule, /color:[^;]+!important/);
  assert.match(selectedRule, /background:var\(--sh-accent\)!important/);
  assert.match(selectedRule, /border:[^;]+!important/);
  assert.match(selectedRule, /box-shadow:[^;]+!important/);
  assert.match(stylesheet, /\.batch-factory-unified-settings-modal \.ant-segmented-item-selected \.ant-segmented-item-label::before\{content:'✓'/);
});

test('uses a website profile organization as the unified organization default', () => {
  assert.match(source, /const websiteProfileOrganization = profile =>/);
  assert.match(source, /const selectedProfileOrganization = websiteProfileOrganization\(selectedWebsiteProfile\)/);
  assert.match(source, /organization: websiteProfileOrganization\(profile\)/);
  assert.doesNotMatch(source, /请选择视频管理系统组织归属/);
  assert.doesNotMatch(source, /onChange=\{organization => patch\(\{ publishSettings:/);
});

test('unified AI reasoning exposes all four constraint layers plus the base-setup switch', () => {
  assert.match(source, /<b>画面前缀词<\/b>/);
  assert.match(source, /<b>画质约束<\/b>/);
  assert.match(source, /<b>画面限制<\/b>/);
  assert.match(source, /<b>负面提示词<\/b>/);
  assert.match(source, /selectedConstraint\('prefix'\)/);
  assert.match(source, /byConstraint\('prefix'\)/);
  assert.match(source, /updateConstraint\('prefix', null, presetId\)/);
  assert.match(source, /byConstraint\(\['quality', 'comic_quality'\]\)/);
  assert.match(source, /updateConstraint\('quality', 'comic_quality', presetId\)/);
  assert.match(source, /byConstraint\(\['restriction', 'comic_restriction'\]\)/);
  assert.match(source, /updateConstraint\('restriction', 'comic_restriction', presetId\)/);
  assert.match(source, /byConstraint\(\['negative', 'comic_negative'\]\)/);
  assert.match(source, /updateConstraint\('negative', 'comic_negative', presetId\)/);
});

test('unified constraint selector fetches and persists the preset body instead of saving an empty selection', () => {
  assert.match(source, /import \{ getConstraintPresetTexts \} from '..\/..\/..\/shared\/api\/generation'/);
  assert.match(source, /getConstraintPresetTexts\(\[normalizedId\]\)/);
  assert.match(source, /apply\(String\(result\?\.texts\?\.\[normalizedId\] \|\| ''\)\)/);
  assert.match(source, /\{ \.\.\.presetValue\(selected\), constraintCategory: comicCategory, body \}/);
  assert.match(source, /\{ \.\.\.presetValue\(selected\), body \}/);
});

test('unified settings keep smart unified selectable when catalog metadata is incomplete', () => {
  assert.match(source, /SMART_UNIFIED_PRESET_ID = 'script-constraint-prefix-smart-unified'/);
  assert.match(source, /category === 'prefix' && item\.id === SMART_UNIFIED_PRESET_ID/);
  assert.match(source, /preset\.id === SMART_UNIFIED_PRESET_ID \? 'prefix'/);
});

test('persists the default SD video prompt when unified settings are saved without a video selection', () => {
  assert.match(source, /const DEFAULT_VIDEO_PROMPT =/);
  assert.match(source, /function requiredVideoPrompt\(/);
  assert.match(source, /const video = requiredVideoPrompt\(aiPromptConfig\.video\)/);
  assert.match(source, /aiPromptConfig: \{ \.\.\.aiPromptConfig, video \}/);
});

test('keeps base setup as the switch controlling asset prompt display', () => {
  assert.match(source, /<span>基础设定<\/span>/);
  assert.match(source, /constraints\.baseSetup\?\.enabled/);
});

test('shows a compact collapsible style catalog from the existing web-submit config', () => {
  assert.match(source, /const \[styleOpen, setStyleOpen\] = useState\(false\)/);
  assert.match(source, /account\?\.style_catalog/);
  assert.match(source, /styleCatalogNames\.length/);
  assert.match(source, /尚未同步风格目录/);
  assert.match(source, /batch-factory-style-catalog-list/);
  assert.match(source, /setAccount\(current => \(\{ \.\.\.\(current \|\| \{\}\), style_catalog: styles \}\)\)/);
  assert.match(stylesheet, /\.batch-factory-style-catalog\{/);
  assert.match(stylesheet, /@media\(max-width:900px\)\{\.batch-factory-style-catalog/);
});

test('统一配置面板不超视口且底部保存按钮固定', () => {
  assert.match(stylesheet, /\.batch-factory-unified-settings-modal\s*\{[^}]*max-height:\s*calc\(100vh - 48px\)/s);
  assert.match(stylesheet, /\.batch-factory-unified-settings-modal \.ant-modal-body\s*\{[^}]*overflow-y:\s*auto/s);
  assert.match(stylesheet, /\.batch-factory-unified-settings-modal \.ant-modal-footer\s*\{[^}]*position:\s*sticky;[^}]*bottom:\s*0/s);
});
