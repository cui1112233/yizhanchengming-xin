import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const sourcePath = path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx');
const source = fs.readFileSync(sourcePath, 'utf8');

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

test('unified configuration makes an unavailable model directory actionable instead of rendering empty selectors', () => {
  assert.match(source, /const \[modelsError, setModelsError\] = useState\(''\)/);
  assert.match(source, /setModelsError\(error\?\.message \|\| '未能读取个人中心已启用模型'\)/);
  assert.match(source, /message="模型目录暂不可用"/);
  assert.match(source, /前往个人中心配置模型/);
});

test('unified configuration loads every typed model without passing Array map indexes as request functions', () => {
  assert.match(source, /listAvailableModels\('text'\)/);
  assert.match(source, /listAvailableModels\('image'\)/);
  assert.match(source, /listAvailableModels\('video'\)/);
  assert.doesNotMatch(source, /\.map\(listAvailableModels\)/);
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
  const modelEnd = source.indexOf('{modelsError ?');
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
  assert.match(source, /value=\{assets\.extraction\?\.presetId \|\| undefined\}/);
  assert.match(source, /bySlot\('script\.asset-extraction'\)/);
  assert.doesNotMatch(source, /<b>人物提示词<\/b>/);
  assert.doesNotMatch(source, /<b>场景提示词<\/b>/);
  assert.match(source, /noPromptOption = \{ value: '__none__', label: '无' \}/);
  assert.match(source, /promptOptions = options => \[noPromptOption, \.\.\.options\]/);
  assert.match(source, /normalizePromptId = presetId => presetId === '__none__'/);
  assert.match(source, /placeholder="选择提示词或无"/);
  assert.match(source, /video: \{ \.\.\.video, \.\.\.presetValue\(catalog\.find\(item => item\.id === normalizedId\)\), enabled: Boolean\(normalizedId\) \}/);
});

test('keeps the automation gear beside the unified configuration label', () => {
  assert.match(source, /<div className="batch-factory-unified-settings-title"><BatchFactoryAutomationPresetManager value=\{draftPatch\} onLoad=\{setDraftPatch\} \/><span>统一配置<\/span><\/div>/);
});

test('uses a website profile organization as the unified organization default', () => {
  assert.match(source, /const websiteProfileOrganization = profile =>/);
  assert.match(source, /const selectedProfileOrganization = websiteProfileOrganization\(selectedWebsiteProfile\)/);
  assert.match(source, /organization: websiteProfileOrganization\(profile\)/);
  assert.doesNotMatch(source, /请选择视频管理系统组织归属/);
  assert.doesNotMatch(source, /onChange=\{organization => patch\(\{ publishSettings:/);
});

test('unified AI reasoning exposes all four constraint layers plus the base-setup switch', () => {
  assert.match(source, /<b>画面前缀词（智能统一）<\/b>/);
  assert.match(source, /<b>画质约束<\/b>/);
  assert.match(source, /<b>画面限制<\/b>/);
  assert.match(source, /<b>负面提示词<\/b>/);
  for (const category of ['prefix', 'quality', 'restriction', 'negative']) {
    assert.match(source, new RegExp(`selectedConstraint\\('${category}'\\)`));
    assert.match(source, new RegExp(`byConstraint\\('${category}'\\)`));
    assert.match(source, new RegExp(`updateConstraint\\('${category}', presetId\\)`));
  }
});

test('unified constraint selector fetches and persists the preset body instead of saving an empty selection', () => {
  assert.match(source, /import \{ getConstraintPresetTexts \} from '..\/..\/..\/shared\/api\/generation'/);
  assert.match(source, /getConstraintPresetTexts\(\[presetId\]\)/);
  assert.match(source, /const body = String\(result\?\.texts\?\.\[presetId\] \|\| ''\)/);
  assert.match(source, /selections: \[\.\.\.rest, \{ \.\.\.presetValue\(selected\), body \}\]/);
});
