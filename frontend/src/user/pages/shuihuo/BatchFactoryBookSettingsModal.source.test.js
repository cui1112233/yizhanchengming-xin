import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');

test('single-book configuration compares inherited settings and only saves changed fields', () => {
  assert.match(source, /export function buildBookOverridePatch/);
  assert.match(source, /buildBookRegionUpdate/);
  assert.match(source, /saveBookOverride/);
  assert.match(source, /expectedRevision/);
  assert.match(source, /当前书覆盖/);
  assert.match(source, /当前批量作品配置/);
});

test('single-book configuration selects enabled text, image and video models', () => {
  assert.match(source, /listAvailableModels\('text'\)/);
  assert.match(source, /listAvailableModels\('image'\)/);
  assert.match(source, /listAvailableModels\('video'\)/);
  assert.match(source, /modelSelectOptions\(models, 'text'\)/);
  assert.match(source, /modelSelectOptions\(models, 'image'\)/);
  assert.match(source, /modelSelectOptions\(models, 'video'\)/);
  assert.match(source, /文本模型/);
  assert.match(source, /图片模型/);
  assert.match(source, /视频模型/);
  assert.match(source, /import \{ listAvailableModels, modelSelectOptions \} from '\.\.\/\.\.\/\.\.\/shared\/api\/modelCatalog'/);
  for (const kind of ['text', 'image', 'video']) {
    assert.ok(source.includes(`modelSelectOptions(models, '${kind}')`));
  }
  assert.doesNotMatch(source, /item\.name \|\| item\.id/);
});

test('single-book configuration has the same layered script constraints and video overrides as the confirmed V11 plan', () => {
  assert.match(source, /基础设定（人物 \/ 场景）/);
  assert.match(source, /画面前缀词/);
  assert.match(source, /画质约束/);
  assert.match(source, /画面限制/);
  assert.match(source, /负面提示词/);
  assert.match(source, /enabledCategories/);
	assert.match(source, /分镜时长/);
	assert.match(source, /value: 10, label: '10s'/);
	assert.match(source, /value: 15, label: '15s'/);
  assert.doesNotMatch(source, /固定开头/);
  assert.match(source, /视频引擎/);
  assert.doesNotMatch(source, /约束设置总规则/);
});

test('single-book constraints support the same system, personal and editable prompt sources as script generation', () => {
  assert.match(source, /getConstraintPresetTexts/);
  assert.match(source, /listScriptConstraintPrompts/);
  assert.match(source, /saveScriptConstraintPrompt/);
  assert.match(source, /updateScriptConstraintPrompt/);
  assert.match(source, /deleteScriptConstraintPrompt/);
  assert.match(source, /我的提示词/);
  assert.match(source, /提示词内容/);
  assert.match(source, /保存为我的提示词/);
  assert.match(source, /开启后，当前书每个分镜自动带入已提取的人物与场景设定/);
});

test('single-book media configuration saves video and comic modules together instead of dropping the combined region', () => {
  assert.match(source, /const moduleKey = region === 'media' \? 'video' : AI_REGION_KEYS\.get\(region\)/);
  assert.match(source, /region === 'media' \? \['video', 'comic'\] : region === 'video' \? \['video'\] : \[moduleKey\]/);
  assert.match(source, /if \(region === 'media'\) \{[\s\S]*?patch\.productionMode = edited\?\.productionMode;[\s\S]*?\}/);
});

test('single-book comic style library lives in engine media settings instead of constraints', () => {
  const mediaStart = source.indexOf('const mediaBody =');
  const mediaEnd = source.indexOf('let body = null;', mediaStart);
  const mediaBody = source.slice(mediaStart, mediaEnd);
  const constraintsStart = source.indexOf("} else if (region === 'constraints') {");
  const constraintsEnd = source.indexOf("} else if (region === 'media') {", constraintsStart);
  const constraintsBody = source.slice(constraintsStart, constraintsEnd);

  assert.match(mediaBody, /<BatchFactoryComicStylePicker/);
  assert.doesNotMatch(constraintsBody, /<BatchFactoryComicStylePicker/);
});

test('single-book configuration can choose a separate image generation provider', () => {
  assert.match(source, /图片生成通道/);
  assert.match(source, /field="imageProvider"/);
  assert.match(source, /from '.\/providerModelBinding'/);
  assert.match(source, /'imageProvider'/);
  assert.match(source, /providerOptionsForModels\(models, 'image'\)/);
  assert.match(source, /providerOptionsForModels\(models, 'video'\)/);
});

test('single-book local Doubao provider exposes its supported video models', () => {
  for (const [model, description] of [
    ['Seedance 2.5', '旗舰视频标杆 · 5 倍消耗'],
    ['Seedance 2.0', '进阶画面表现 · 2 倍消耗'],
    ['Seedance 2.0 Fast', '快速出片选择'],
    ['Seedance 2.0 Mini', '日常生成使用']
  ]) {
    assert.match(source, /providerModelSettingsPatch/);
    assert.match(source, /modelOptionsForProvider/);
  }
  assert.match(source, /modelOptionsForProvider\(models, 'video', effectiveVideoProvider, effectiveVideoModelId\)/);
  assert.match(source, /providerModelSettingsPatch\(\{ models, kind: 'video'/);
});

test('single-book local Doubao provider exposes its supported image models', () => {
  for (const [model, description] of [
    ['Seedream 5.0 Flash', '专业快速'],
    ['Seedream 5.0 Pro', '专业出图 · 4 倍消耗'],
    ['Seedream 4.5', '日常生成'],
    ['Seedream 4.0', '基础生图']
  ]) {
    assert.match(source, /providerModelSettingsPatch/);
    assert.match(source, /modelOptionsForProvider/);
  }
  assert.match(source, /modelOptionsForProvider\(models, 'image', effectiveImageProvider\)/);
  assert.match(source, /providerModelSettingsPatch\(\{ models, kind: 'image'/);
  assert.doesNotMatch(source, /豆包本地版/);
});

test('switching the single-book image channel also selects a model from that channel', () => {
  assert.match(source, /label="图片生成通道"[\s\S]*?onChange=\{imageProvider => patch\(providerModelSettingsPatch\(\{ models, kind: 'image'/);
});

test('inherited engine selects are not wrapped by a label that re-triggers option clicks', () => {
  const start = source.indexOf('function InheritedField(');
  const end = source.indexOf('function InheritedNumberField(', start);
  const field = source.slice(start, end);
  assert.match(field, /return <div className="batch-factory-engine-field"/);
  assert.doesNotMatch(field, /return <label className="batch-factory-engine-field"/);
});

test('single-book configuration opens and saves one explicit region at a time', () => {
  assert.match(source, /activeRegion/);
  assert.match(source, /buildBookRegionUpdate/);
  assert.match(source, /AI_REGION_KEYS/);
  assert.match(source, /onOpenBookAssets/);
  assert.match(source, /维护当前书人物场景预设/);
  assert.match(source, /restoreCurrentRegion/);
});

test('single-book assets expose script-compatible starred-character focus without treating it as VIDEO exclusion', () => {
  assert.match(source, /starredCharacterNames/);
  assert.match(source, /星标人物聚焦/);
  assert.match(source, /mode="multiple"/);
});

test('single-book assets reuse one selector for video and comic asset presets', () => {
  assert.match(source, /script\.asset-extraction/);
  assert.match(source, /batch\.comic-assets/);
  assert.match(source, /选择资产提示词/);
  assert.match(source, /updateAssetSelection\('extraction', selection\)/);
  assert.doesNotMatch(source, /updateAssetSelection\('character', selection\)/);
  assert.doesNotMatch(source, /updateAssetSelection\('scene', selection\)/);
  assert.match(source, /const moduleKeys = region === 'media' \? \['video', 'comic'\]/);
});

test('single-book engine settings can override opening variants and restore batch inheritance', () => {
  assert.match(source, /'openingEnabled', 'openingCount'/);
  assert.match(source, /for \(const key of \['openingEnabled', 'openingCount'\]\)/);
  assert.match(source, /form\.openingEnabled === true/);
  assert.match(source, /patch\(\{ openingEnabled: enabled, openingCount: form\.openingCount \?\? inherited\.openingCount \?\? 4 \}\)/);
  assert.match(source, /patch\(\{ openingCount: count \?\? 4 \}\)/);
  assert.match(source, /Object\.hasOwn\(bookPatch, 'openingEnabled'\)/);
  // 换开头属于引擎配置，不再出现在资产设置卡
  assert.doesNotMatch(source, /换开头（当前书覆盖）/);
});

test('single-book configuration keeps explanations behind question-mark tips and removes fixed opening', () => {
  assert.match(source, /batch-factory-help-button/);
  assert.doesNotMatch(source, /固定开头/);
  assert.match(source, /function ConfigCard\(\{ title, description, children \}\)/);
  assert.match(source, /<ConfigCard title="核心视频引擎" description=/);
  assert.match(source, /<ConfigCard title="配音与时长" description=/);
});

test('single-book constraints view backfills empty system-constraint bodies on open for display', () => {
  // 打开约束页时，对“系统预设 + 有 presetId + 正文为空”的层静默拉取正文回显
  assert.match(source, /if \(!open \|\| region !== 'constraints' \|\| loading\) return undefined/);
  assert.match(source, /layer\.source === 'system' && layer\.presetId && !String\(layer\.body \|\| ''\)\.trim\(\)/);
  assert.match(source, /getConstraintPresetTexts\(missing\.map\(\(\{ layer \}\) => layer\.presetId\)\)/);
  assert.match(source, /const body = String\(texts\[layer\.presetId\] \|\| ''\)\.trim\(\)/);
  assert.match(source, /withDirectorConstraintSelections\(rules\)/);
});

test('keeps storyboard duration in engine configuration instead of video settings', () => {
  const videoRegion = source.match(/else if \(region === 'video'\) \{([\s\S]*?)\n  \} else \{/ )?.[1] || '';
  assert.doesNotMatch(videoRegion, /InheritedStoryboardDurationField/);
	assert.doesNotMatch(videoRegion, /fixedSingleVideo/);
	assert.doesNotMatch(videoRegion, /固定开头/);
});

test('does not expose a separate fixed VIDEO duration control', () => {
  assert.doesNotMatch(source, /固定 VIDEO 时长/);
  assert.doesNotMatch(source, /fixedVideoDuration/);
});

test('keeps derived-opening presets out of single-book video settings', () => {
  const videoRegion = source.match(/else if \(region === 'video'\) \{([\s\S]*?)\n  \} else \{/ )?.[1] || '';
  assert.doesNotMatch(videoRegion, /原文直转导演/);
  assert.doesNotMatch(videoRegion, /爆款开头导演/);
  assert.doesNotMatch(videoRegion, /爆款开头改编/);

	assert.match(videoRegion, /<RuleModule required title="视频提示词"/);
  assert.match(videoRegion, /<PromptSelect required label="视频提示词"/);
	assert.match(source, /const DEFAULT_VIDEO_PROMPT = \{ presetId: 'batch-video-sd'/);
	assert.match(source, /video: requiredVideoPrompt\(aiPromptConfig\.video\)/);
});

test('single-book constraints keep the smart-unified preset visible when old catalogs omit its category', () => {
  assert.match(source, /SMART_UNIFIED_PRESET_ID = 'script-constraint-prefix-smart-unified'/);
  assert.match(source, /category === 'prefix' && item\?\.id === SMART_UNIFIED_PRESET_ID/);
  assert.match(source, /preset\.id === SMART_UNIFIED_PRESET_ID \? 'prefix'/);
});

test('single-book constraints hide the raw smart-unified system prompt from the editor', () => {
  assert.match(source, /const isSmartUnified = value\.presetId === SMART_UNIFIED_PRESET_ID/);
  assert.match(source, /智能统一由系统自动分析，原始系统提示词不在这里展示/);
  assert.match(source, /isSmartUnified \? <Typography\.Text/);
});

test('uses the comic page and image-derivation presets for a book comic-generation override', () => {
  assert.match(source, /const comicPageRules = useMemo\(\(\) => catalog\.batch\.filter\(item => item\.slot === 'batch\.comic-page'\)/);
  assert.match(source, /const moduleKeys = region === 'media' \? \['video', 'comic'\]/);
  assert.match(source, /<RuleModule title="漫画生成" value=\{\{ \.\.\.ai\.comic, enabled: productionMode === 'comic' \}\}/);
  assert.match(source, /<PromptSelect required label="漫画提示词" value=\{ai\.comic\}/);
  assert.match(source, /<PromptSelect label="图片描述词推导"/);
});

test('single-book media configuration uses one mutually exclusive production mode', () => {
  assert.match(source, /productionMode/);
  assert.match(source, /图片描述词推导/);
  assert.match(source, /batch\.comic-image-derivation/);
  assert.match(source, /选择资产提示词/);
  assert.doesNotMatch(source, /<PromptSelect label="漫画资产提示词"/);
});

test('enabling comic production selects the default comic preset and uses comic-specific empty text', () => {
  assert.match(source, /const defaultComicPage = comicPageRules\[0\]/);
  assert.match(source, /patchAI\('comic', scoped\(\{ \.\.\.ai\.comic, \.\.\.selectedMeta\(comicPageRules, defaultComicPage\.id\), enabled: true \}\)\)/);
  assert.match(source, /placeholder=\{required \? `请选择已发布\$\{label\}` : '未选择'\}/);
});


test('single-book settings keep publish mappings inside engine configuration without replacing inherited fields', () => {
  assert.match(source, /audioMergeEnabled/);
  assert.match(source, /label: '发布统一'/);
  assert.match(source, /BatchFactoryComicPublishSettingsForm/);
  assert.match(source, /productionMode === 'comic' \? <div[\s\S]*?BatchFactoryComicPublishSettingsForm/);
  assert.match(source, /comicPublishSettings: mergePublishSettings/);
  assert.match(source, /mergePublishSettings/);
  assert.match(source, /网站配置档/);
  assert.match(source, /素材使用/);
  assert.match(source, /水平翻转/);
  assert.doesNotMatch(source, /region === 'publish'/);
});

test('lets a book override only the inherited publish organization', () => {
  assert.match(source, /get121OrganizationOptions/);
  assert.match(source, /组织归属（当前书覆盖）/);
  assert.match(source, /继承批量组织归属/);
  assert.match(source, /publishSettings: \{ \.\.\.publish, organization/);
});

test('does not expose the retired fixed-opening switch in the current book engine configuration', () => {
  assert.doesNotMatch(source, /function InheritedFixedVideoSwitch/);
  assert.doesNotMatch(source, /固定开头/);
  assert.match(source, /batch-factory-book-audio/);
  assert.match(source, /batch-factory-audio-option/);
});

test('keeps help buttons from hijacking prompt-select clicks', () => {
  const start = source.indexOf('function PromptSelect');
  const end = source.indexOf('function RuleModule', start);
  const promptSelect = source.slice(start, end);
  assert.match(promptSelect, /return <div className="batch-factory-engine-field">/);
  assert.doesNotMatch(promptSelect, /return <label className="batch-factory-engine-field">/);
});

test('uses the current Personal Center TTS defaults for each book and persists only a book-level TTS override', () => {
  assert.match(source, /import \{ getConfig \} from '\.\.\/\.\.\/\.\.\/shared\/api\/config'/);
  assert.match(source, /'tts'/);
  assert.match(source, /const config = await getConfig\(\)/);
  assert.match(source, /textToSpeech\(\{ input, \.\.\.nextTts \}\)/);
  assert.match(source, /const nextTts = \{ \.\.\.DEFAULT_TTS, \.\.\.\(config\?\.tts \|\| \{\}\), \.\.\.\(form\.tts \|\| \{\}\) \}/);
  assert.match(source, /当前书 TTS/);
});
