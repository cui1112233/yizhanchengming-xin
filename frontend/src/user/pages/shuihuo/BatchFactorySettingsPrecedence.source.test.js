import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const read = name => readFileSync(join(here, name), 'utf8');
const novelList = read('BatchFactoryNovelList.jsx');
const unifiedModal = read('BatchFactoryUnifiedSettingsModal.jsx');
const bookModal = read('BatchFactoryBookSettingsModal.jsx');
const engineDrawer = read('BatchFactoryEngineSettingsDrawer.jsx');
const helpers = read('batchFactoryNovelListHelpers.js');
const regions = read('batchFactoryBookConfigRegions.js');

test('runtime merge chain keeps the book patch above the batch unified configuration', () => {
  const fn = helpers.match(/export function effectiveBookSettings\(batch, book\) \{[\s\S]*?\n\}/)?.[0] || '';
  assert.match(fn, /\.\.\.batchPatch,\s*\.\.\.bookPatch/);
  assert.match(fn, /publishSettings: \{ \.\.\.\(batchPatch\.publishSettings \|\| \{\}\), \.\.\.\(bookPatch\.publishSettings \|\| \{\}\) \}/);
});

test('batch productionMode stays authoritative over a stale single-book value', () => {
  const fn = helpers.match(/export function effectiveBookSettings\(batch, book\) \{[\s\S]*?\n\}/)?.[0] || '';
  assert.match(fn, /if \(\['comic', 'video'\]\.includes\(batchPatch\.productionMode\)\) effective\.productionMode = batchPatch\.productionMode;/);
  // 视频批次不允许因为单书遗留的 comic 配置进入漫画栏目。
  assert.match(fn, /else if \(!\['comic', 'video'\]\.includes\(effective\.productionMode\)/);
});

test('a saved unified configuration is pushed through to every book in the batch', () => {
  assert.match(novelList, /async function syncUnifiedSettingsToBooks\(currentBatch\) \{/);
  assert.match(novelList, /const pushKeys = UNIFIED_CONFIGURATION_KEYS[\s\S]{0,200}filter\(key => Object\.hasOwn\(unifiedPatch, key\)\)/);
  assert.match(novelList, /const candidates = currentBatch\?\.books \|\| \[\]/);
  assert.match(novelList, /saveBookOverrideWithRetry\([\s\S]{0,160}\{ patch: pushPatch \}/);
  // 旧的 restoreKeys 写法会删掉单书覆盖值，等价于“批次盖单书”，必须消失。
  assert.doesNotMatch(novelList, /restoreKeys: UNIFIED_CONFIGURATION_KEYS/);
  assert.match(novelList, /统一配置已覆盖本批次 /);
  // 写透后按实际落库的批次重新取书，再用取回的书数提示覆盖范围。
  assert.match(novelList, /const syncedBooks = Array\.isArray\(syncedBatch\?\.books\) \? syncedBatch\.books : \[\]/);
  assert.match(novelList, /统一配置已覆盖本批次 \$\{syncedBooks\.length\}/);
});

test('unified configuration covers engine, model, media, opening and publish fields', () => {
  const keys = novelList.match(/const UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  for (const key of [
    'videoProvider', 'videoModelId', 'textModelId', 'imageModelId', 'imageProvider',
    'assetImagesEnabled',
    'sensitiveAIEnabled', 'sensitiveAIModels', 'sensitiveAIOnFailure',
    'publishSettings', 'publishRewriteEnabled', 'comicPublishSettings',
    'maxVideoDuration', 'videoAspectRatio', 'videoResolution', 'storyboardDurationLimit',
    'openingEnabled', 'openingCount'
  ]) {
    assert.match(keys, new RegExp(`'${key}'`), `missing ${key}`);
  }
});

test('the asset image switch defaults to off, is pushed batch-wide and only badges single-book deviations', () => {
  // 统一配置弹窗里的开关：默认关（只有显式 true 才开启），文案说明跳过资产不影响后续流程。
  assert.match(unifiedModal, /<b>生成资产图片<\/b>/);
  assert.match(unifiedModal, /checked=\{value\.assetImagesEnabled === true\}/);
  assert.match(unifiedModal, /默认关闭/);
  assert.match(unifiedModal, /关闭时跳过资产生成阶段/);
  assert.match(unifiedModal, /不影响分镜、视频与漫画流程/);
  assert.match(unifiedModal, /已有资产图片仍会照常使用/);
  // 保存统一配置时纳入写透集合，逐本下发，单书可覆盖。
  const keys = novelList.match(/const UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  assert.match(keys, /'assetImagesEnabled'/, 'assetImagesEnabled must be part of UNIFIED_CONFIGURATION_KEYS');
  assert.match(bookModal, /ENGINE_MODEL_OVERRIDE_FIELDS = \[[\s\S]*?'assetImagesEnabled'[\s\S]*?\];/);
  // 单书沿用统一 FieldSource 规则：批次未写该键时按关闭继承，只有偏离批次并显式保存时才出现“自定义”。
  assert.match(bookModal, /<InheritedSwitchField label="生成资产图片"[\s\S]{0,220}inherited=\{inherited\.assetImagesEnabled === true\}[\s\S]{0,220}explicit=\{Object\.hasOwn\(bookPatch, 'assetImagesEnabled'\)\}/);
});

test('the batch AI repair switch defaults to off and only an explicit true turns it on', () => {
  // 统一配置弹窗：未设置时不再默认开启，只有显式 true 才打开。
  assert.match(unifiedModal, /<b>AI 修复<\/b>/);
  assert.match(unifiedModal, /checked=\{value\.sensitiveAIEnabled === true\}/);
  assert.match(unifiedModal, /默认关闭；只有显式开启后才会调用 AI 修复/);
  // 单书继承值同样默认关闭，偏离批次时才出现“自定义”。
  assert.match(bookModal, /<InheritedSwitchField label="AI 修复"[\s\S]{0,260}inherited=\{inherited\.sensitiveAIEnabled === true\}/);
  assert.match(bookModal, /<InheritedSwitchField label="AI 修复"[\s\S]{0,260}explicit=\{Object\.hasOwn\(bookPatch, 'sensitiveAIEnabled'\)\}/);
  // 仍纳入统一配置写透集合，保证批次级开关能按现有继承链下发。
  const keys = novelList.match(/const UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  assert.match(keys, /'sensitiveAIEnabled'/, 'sensitiveAIEnabled must stay in UNIFIED_CONFIGURATION_KEYS');
});

test('unified configuration only writes configuration and never pushes generated content', () => {
  const configBlock = novelList.match(/const UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  const contentBlock = novelList.match(/const GENERATED_CONTENT_SETTINGS_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  const listed = block => [...block.matchAll(/'([A-Za-z0-9_]+)'/g)].map(match => match[1]);
  const configKeys = listed(configBlock);
  const contentKeys = listed(contentBlock);
  assert.ok(contentKeys.length > 0, 'generated content key list is missing');
  for (const contentKey of ['audioLineDurations', 'audioDurationFingerprint', 'ttsAudioArtifactId', 'h3StyleAnalysis', 'storyTimelineVersion']) {
    assert.ok(contentKeys.includes(contentKey), `generated content key ${contentKey} must be declared`);
    assert.ok(!configKeys.includes(contentKey), `unified configuration must not push generated content key ${contentKey}`);
  }
  // 已生成的配音测量值、配音指纹、H3 风格分析都不允许出现在推送键里。
  assert.ok(!configKeys.includes('audioDurationSeconds'), 'audioDurationSeconds is a measured artifact');
  // 推送链路必须显式排除内容键，防止以后有人往配置键里加内容字段。
  assert.match(novelList, /pushKeys = UNIFIED_CONFIGURATION_KEYS[\s\S]{0,200}!GENERATED_CONTENT_SETTINGS_KEYS\.includes\(key\)/);
});

test('aiPromptConfig is documented as configuration only, not per-book generated prompts', () => {
  assert.ok(novelList.includes('aiPromptConfig 存的是提示词预设引用'), 'missing aiPromptConfig configuration-only note');
  assert.ok(novelList.includes('不含任何一本书已生成的分镜/提示词/图片/视频文本'), 'missing generated-content exclusion note');
  const configBlock = novelList.match(/const UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?\n\];/)?.[0] || '';
  assert.match(configBlock, /'aiPromptConfig'/);
});

test('the unified settings footer states the batch-wide overwrite scope', () => {
  assert.ok(unifiedModal.includes('保存将覆盖本批次 '), 'missing batch-wide scope hint');
  assert.ok(unifiedModal.includes('本书的对应配置（包括单书已改过的值）'), 'missing overwrite wording');
});

test('book settings only badge the fields that deviate from the batch configuration', () => {
  const fieldSource = bookModal.match(/function FieldSource\(\{[\s\S]*?\n\}/)?.[0] || '';
  assert.ok(fieldSource.length, 'FieldSource not found');
  assert.match(bookModal, /function FieldSource\(\{ value, inherited, explicit \}\)/);
  // 继承态（含统一配置推送且值与批次一致）不渲染任何来源文字。
  assert.match(fieldSource, /if \(!explicit\) return null;/);
  assert.match(fieldSource, /if \(same\) return null;/);
  assert.doesNotMatch(fieldSource, /当前生效：/);
  assert.doesNotMatch(fieldSource, /heredText/);
  // 只有真正偏离批次时才出现唯一的短徽标“自定义”，解释放在 HelpTip 里。
  assert.ok(fieldSource.includes('batch-factory-field-source is-custom">自定义</em>'), 'missing single custom badge');
  assert.match(fieldSource, /该项已在本单书单独设置，不会跟随批次统一配置。/);
  assert.doesNotMatch(bookModal, /统一配置推送/);
  assert.equal((bookModal.match(/整页默认继承批次统一配置/g) || []).length, 1, 'the inheritance hint must appear exactly once');
  assert.ok(bookModal.includes('<div className="batch-factory-inherit-hint">整页默认继承批次统一配置；改动哪个字段，哪个字段就变成“自定义”，保存后不会跟随批次。</div>'), 'missing single inheritance hint');
  for (const [label, key] of [
    ['视频引擎', 'videoProvider'], ['视频模型', 'videoModelId'], ['文本模型', 'textModelId'],
    ['图片模型', 'imageModelId'], ['修复模型', 'sensitiveAIModels'],
    ['视频分辨率', 'videoResolution'], ['视频画幅', 'videoAspectRatio'], ['改文后上传', 'publishRewriteEnabled']
  ]) {
    assert.match(bookModal, new RegExp(`label="${label}"[\\s\\S]{0,900}explicit=\\{Object\\.hasOwn\\(bookPatch, '${key}'\\)\\}`), `missing explicit source for ${label}`);
  }
  assert.match(bookModal, /<InheritedStoryboardDurationField[\s\S]{0,160}explicit=\{Object\.hasOwn\(bookPatch, 'storyboardDurationLimit'\)\}/);
  assert.match(bookModal, /explicit=\{Object\.hasOwn\(bookPatch\.publishSettings \|\| \{\}, 'websiteProfileId'\)\}/);
});

test('the audio target displays two decimals without touching the stored value', () => {
  const field = bookModal.match(/<InheritedNumberField label="配音目标（秒）"[^\n]*/)?.[0] || '';
  assert.ok(field.length, 'audio target field not found');
  assert.match(field, /precision=\{2\}/);
  // 底层仍然写入原始数值，只做展示精度。
  assert.match(field, /patch\(\{ audioDurationSeconds \}\)\} \/>/);
});

test('the config-region badge follows the same precedence as runtime', () => {
  assert.match(regions, /export function bookConfigRegionStatus\(book, region, batchPatch = \{\}\) \{/);
  assert.match(regions, /{ label: '来自统一配置', tone: 'pushed' }/);
  assert.match(regions, /const pushedKeys = keys => keys\.every/);
  assert.match(novelList, /bookConfigRegionStatus\(book, region\.key, batch\?\.settingsState\?\.patch \|\| \{\}\)/);
});

test('the automation failure notice lives outside the editable-content cell and the removed sensitive-review card stays hidden', () => {
  const contentCell = novelList.match(/className="shuihuo-workbench-cell batch-factory-book-content"[\s\S]*?batch-factory-book-content-actions/)?.[0] || '';
  assert.ok(contentCell.length > 0, 'content cell not found');
  assert.doesNotMatch(contentCell, /BatchFactoryErrorNotice/);
  assert.doesNotMatch(contentCell, /自动生产失败/);
  assert.match(novelList, /<div className="batch-factory-book-row-error-strip">/);
  assert.ok(novelList.includes("historical ? '历史自动生产失败' : '自动生产失败'"), 'missing stage failure title');
  const strip = novelList.match(/<div className="batch-factory-book-row-error-strip">[\s\S]*?<\/div>/)?.[0] || '';
  assert.match(strip, /<BatchFactoryErrorNotice/);
  const sensitiveCount = (novelList.match(/batch-factory-sensitive-review/g) || []).length;
  assert.equal(sensitiveCount, 0, 'the sensitive-word review card must stay removed');
});

test('unified settings save closes the dialog before audio measurement and prompt recompilation', () => {
  const save = novelList.match(/async function saveSettings\(patch\) \{[\s\S]*?\n  \}/)?.[0] || '';
  assert.ok(save.length, 'saveSettings not found');
  assert.doesNotMatch(save, /await prepareAudioPlanningForBatch/);
  assert.doesNotMatch(save, /await recompileUnifiedH3Prompts/);
  assert.match(save, /message\.success\(`统一配置已覆盖本批次 /);
  assert.match(save, /void runSettingsBackgroundSync\(normalized, syncedBatch\)/);

  const background = novelList.match(/async function runSettingsBackgroundSync\(patch, initialBatch\) \{[\s\S]*?\n  \}/)?.[0] || '';
  assert.ok(background.length, 'background sync not found');
  assert.match(background, /await ensureBookAudioDuration\(book, \{ force: true, quiet: true, batchSnapshot: initialBatch \}\)/, 'background must measure per book using the per-book effective settings snapshot');
  assert.match(background, /await recompileUnifiedH3Prompts\(compiledBatch\)/);
  assert.match(novelList, /后台处理中 · 配音测量 \$\{backgroundSettingsTask\.done\}\/\$\{backgroundSettingsTask\.total\}/);
  assert.match(novelList, /重试后台处理/);
});

test('batch factory video dropdowns opt into the FanKe model expansion', () => {
  assert.match(unifiedModal, /modelOptionsForProvider\(models, 'video', value\.videoProvider, value\.videoModelId, \{ expandFanKeModels: true \}\)/);
  assert.match(bookModal, /modelOptionsForProvider\(models, 'video', effectiveVideoProvider, effectiveVideoModelId, \{ expandFanKeModels: true \}\)/);
  assert.match(engineDrawer, /modelOptionsForProvider\(models, 'video', form\.videoProvider, form\.videoModelId, \{ expandFanKeModels: true \}\)/);
  for (const file of [unifiedModal, bookModal, engineDrawer]) {
    assert.match(file, /providerModelSettingsPatch\(\{ models, kind: 'video', provider: [^,]+, modelId: [^,]+, expandFanKeModels: true \}/);
  }
});

test('prompt template settings stay human-readable and never expose the raw key', () => {
  assert.match(unifiedModal, /提示词规则模板/);
  assert.match(unifiedModal, /不是每本书已经生成出来的分镜和提示词/);
  // 键名在代码里被合法读取（draftPatch.aiPromptConfig），但不得作为可见文本/标签渲染给用户。
  assert.doesNotMatch(unifiedModal, />\s*aiPromptConfig\s*</);
  assert.doesNotMatch(unifiedModal, /label="aiPromptConfig"/);
  assert.doesNotMatch(unifiedModal, /提示词\s*aiPromptConfig/);
  assert.doesNotMatch(bookModal, />\s*aiPromptConfig\s*</);
  assert.doesNotMatch(bookModal, /label="aiPromptConfig"/);
});
