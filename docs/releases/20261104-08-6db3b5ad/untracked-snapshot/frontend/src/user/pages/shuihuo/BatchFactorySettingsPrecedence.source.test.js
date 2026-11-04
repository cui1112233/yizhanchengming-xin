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
    'sensitiveAIEnabled', 'sensitiveAIModels', 'sensitiveAIOnFailure',
    'publishSettings', 'publishRewriteEnabled', 'comicPublishSettings',
    'maxVideoDuration', 'videoAspectRatio', 'videoResolution', 'storyboardDurationLimit',
    'openingEnabled', 'openingCount'
  ]) {
    assert.match(keys, new RegExp(`'${key}'`), `missing ${key}`);
  }
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

test('book settings show whether a value is inherited, pushed or book-specific', () => {
  assert.match(bookModal, /function FieldSource\(\{ value, inherited, options, explicit \}\)/);
  assert.ok(bookModal.includes('继承批次统一配置${inheritedText ? `（当前生效：${inheritedText}）` : \'\'}'), 'missing inherited-source label');
  assert.match(bookModal, /'统一配置推送' : '单书自定义'/);
  assert.match(bookModal, /'单书自定义'/);
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

test('the config-region badge follows the same precedence as runtime', () => {
  assert.match(regions, /export function bookConfigRegionStatus\(book, region, batchPatch = \{\}\) \{/);
  assert.match(regions, /{ label: '来自统一配置', tone: 'pushed' }/);
  assert.match(regions, /const pushedKeys = keys => keys\.every/);
  assert.match(novelList, /bookConfigRegionStatus\(book, region\.key, batch\?\.settingsState\?\.patch \|\| \{\}\)/);
});

test('the automation failure notice lives outside the editable-content cell and does not reuse the sensitive-review class', () => {
  const contentCell = novelList.match(/className="shuihuo-workbench-cell batch-factory-book-content"[\s\S]*?batch-factory-book-content-actions/)?.[0] || '';
  assert.ok(contentCell.length > 0, 'content cell not found');
  assert.doesNotMatch(contentCell, /BatchFactoryErrorNotice/);
  assert.doesNotMatch(contentCell, /自动生产失败/);
  assert.match(novelList, /<div className="batch-factory-book-row-error-strip">/);
  assert.ok(novelList.includes("failed?.stale ? '历史自动生产失败' : '自动生产失败'"), 'missing stage failure title');
  const strip = novelList.match(/<div className="batch-factory-book-row-error-strip">[\s\S]*?<\/div>/)?.[0] || '';
  assert.match(strip, /<BatchFactoryErrorNotice/);
  const sensitiveCount = (novelList.match(/batch-factory-sensitive-review/g) || []).length;
  assert.equal(sensitiveCount, 1, 'only the real sensitive-word alert may use batch-factory-sensitive-review');
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
