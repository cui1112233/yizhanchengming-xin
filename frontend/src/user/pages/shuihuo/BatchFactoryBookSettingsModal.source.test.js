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

test('single-book assets use one combined SD asset preset', () => {
  assert.match(source, /script\.asset-extraction/);
  assert.match(source, /SD 资产预设提示词/);
  assert.match(source, /updateAssetSelection\('extraction', selection\)/);
  assert.doesNotMatch(source, /updateAssetSelection\('character', selection\)/);
  assert.doesNotMatch(source, /updateAssetSelection\('scene', selection\)/);
  assert.match(source, /const moduleKeys = region === 'video' \|\| region === 'media' \? \['video'\] : \[moduleKey\]/);
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
  assert.match(source, /const moduleKeys = region === 'video' \|\| region === 'media' \? \['video'\]/);
	assert.match(videoRegion, /<RuleModule required title="视频提示词"/);
  assert.match(videoRegion, /<PromptSelect required label="视频提示词"/);
  assert.match(source, /const DEFAULT_VIDEO_PROMPT = \{ presetId: 'batch-video-sd'/);
	assert.match(source, /video: requiredVideoPrompt\(aiPromptConfig\.video\)/);
});

test('uses the comic page preset for a book comic-generation override', () => {
  assert.match(source, /const comicPageRules = useMemo\(\(\) => catalog\.batch\.filter\(item => item\.slot === 'batch\.comic-page'\)/);
  assert.match(source, /const moduleKeys = region === 'media' \? \['video', 'comic'\]/);
  assert.match(source, /<RuleModule title="漫画生成" value=\{ai\.comic\}/);
  assert.match(source, /<PromptSelect label="漫画提示词" value=\{ai\.comic\}/);
});


test('single-book settings keep publish mappings inside engine configuration without replacing inherited fields', () => {
  assert.match(source, /audioMergeEnabled/);
  assert.match(source, /label: '发布统一'/);
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
