import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { bookSensitiveModeUpdate } from './batchFactoryBookConfigRegions.js';

const directory = path.dirname(fileURLToPath(import.meta.url));
const unifiedSource = fs.readFileSync(path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const bookSource = fs.readFileSync(path.join(directory, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
const createSource = fs.readFileSync(path.join(directory, 'BatchFactoryCreateModal.jsx'), 'utf8');
const novelListSource = fs.readFileSync(path.join(directory, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(directory, '..', 'shuihuo-production.css'), 'utf8');

test('unified settings use the shared production-mode visibility model', () => {
  assert.match(unifiedSource, /import \{[^}]*\bbatchFactoryProductionModeVisibility\b[^}]*\} from '\.\/batchFactoryProductionMode'/);
  assert.match(unifiedSource, /const visibility = batchFactoryProductionModeVisibility\(value\.productionMode\)/);
  assert.match(unifiedSource, /visibility\.video && sections\.includes\('audio'\)/);
});

test('unified reasoning renders only the active mode details', () => {
  assert.match(unifiedSource, /visibility\.comic \? <BatchFactoryComicStylePicker/);
  assert.match(unifiedSource, /visibility\.comic \? <section className="batch-factory-engine-card"><header><b>漫画生成<\/b>/);
  assert.match(unifiedSource, /visibility\.video \? <section className="batch-factory-engine-card"><header><b>视频设置<\/b>/);
});

test('unified reasoning no longer exposes the legacy visual prompt selector', () => {
  assert.doesNotMatch(unifiedSource, /<b>画面提示词<\/b>/);
  assert.doesNotMatch(unifiedSource, /bySlot\('batch\.visual-meta'\)/);
});

test('unified settings switch production mode beside the title without a media tab', () => {
  assert.match(unifiedSource, /export function BatchFactoryProductionModeControl/);
  assert.match(unifiedSource, /<span>统一配置<\/span><BatchFactoryProductionModeControl/);
  assert.doesNotMatch(unifiedSource, /key: 'media', label: '画面与视频'/);
  assert.match(createSource, /<span>新建批量 · 统一配置<\/span><BatchFactoryProductionModeControl/);
  assert.doesNotMatch(createSource, /key: 'media', label: '画面与视频'/);
});

test('the shared production switch offers video, comic and text generation', () => {
  const optionList = unifiedSource.match(/const PRODUCTION_MODE_OPTIONS\s*=\s*\[([\s\S]*?)\]/)?.[1] ?? '';
  const options = Array.from(optionList.matchAll(/\{\s*value:\s*'([^']+)',\s*label:\s*'([^']+)'\s*,?\s*\}/g), ([, value, label]) => [value, label]);
  assert.deepEqual(Object.fromEntries(options), {
    video: '视频生成',
    comic: '漫画生成',
    text: '文本生成'
  });
  assert.match(unifiedSource, /function BatchFactoryProductionModeControl\b[\s\S]*?<Segmented[^>]*\boptions=\{PRODUCTION_MODE_OPTIONS\}/);
});

test('text mode in unified settings keeps the text model and hides the image and video models', () => {
  const modelControls = unifiedSource.match(/<b>模型配置<\/b>[\s\S]*?<\/section>/)?.[0] ?? '';
  const imageControls = modelControls.match(/\{!visibility\.text\s*\?\s*<>([\s\S]*?)<\/>\s*:\s*null\}/)?.[1] ?? '';
  const videoControls = modelControls.match(/\{visibility\.video\s*\?\s*<>([\s\S]*?)<\/>\s*:\s*null\}/)?.[1] ?? '';
  const sharedControls = modelControls.replace(imageControls, '').replace(videoControls, '');
  assert.match(sharedControls, /<b>文本模型<\/b><\/span><Select\b/);
  assert.ok(!/<b>(?:图片生成通道|图片模型|视频模型)<\/b>/.test(sharedControls), 'text mode should expose only the unified text model');
  assert.match(imageControls, /<b>图片生成通道<\/b>/, 'text mode should hide the unified image provider');
  assert.match(imageControls, /<b>图片模型<\/b>/, 'text mode should hide the unified image model');
  assert.match(imageControls, /<ImageCapacityControl\b/, 'text mode should hide the unified image capacity controls');
  assert.match(videoControls, /<b>视频模型<\/b>/, 'only video mode should show the unified video model');
  for (const label of ['视频引擎', '视频画幅', '视频分辨率', '音频来源', '配音音色', '配音风格', '语速']) {
    assert.ok(videoControls.includes(`<b>${label}</b>`), `only video mode should show ${label}`);
  }
  assert.match(unifiedSource, /sections\.includes\('sensitive'\) \? <section[^>]*><header><b>敏感词处理<\/b>/);
});

test('text mode in single-book settings keeps the text model and hides the image and video models', () => {
  const auxiliaryControls = bookSource.match(/<ConfigCard title="辅助生成"[\s\S]*?<\/ConfigCard>/)?.[0] ?? '';
  const imageControls = auxiliaryControls.match(/\{!visibility\.text\s*\?\s*<>([\s\S]*?)<\/>\s*:\s*null\}/)?.[1] ?? '';
  const videoControls = bookSource.match(/\{visibility\.video\s*\?\s*<ConfigCard title="核心视频引擎"[\s\S]*?<\/ConfigCard>/)?.[0] ?? '';
  const sharedControls = auxiliaryControls.replace(imageControls, '');
  assert.match(sharedControls, /label="文本模型" field="textModelId"/);
  assert.ok(!/field="(?:imageProvider|imageModelId|imageAspectRatio|videoModelId)"/.test(sharedControls), 'text mode should expose only the book text model');
  assert.match(imageControls, /label="图片生成通道" field="imageProvider"/, 'text mode should hide the book image provider');
  assert.match(imageControls, /label="图片模型" field="imageModelId"/, 'text mode should hide the book image model');
  assert.match(imageControls, /label="图片画幅" field="imageAspectRatio"/, 'text mode should hide the book image aspect ratio');
  assert.match(videoControls, /label="视频模型" field="videoModelId"/, 'only video mode should show the book video model');
  for (const field of ['videoProvider', 'videoAspectRatio', 'videoResolution']) {
    assert.ok(videoControls.includes(`field="${field}"`), `only video mode should show ${field}`);
  }
  assert.match(bookSource, /visibility\.video \? <ConfigCard title="配音与时长"/);
  assert.match(bookSource, /<ConfigCard title="敏感词处理"/);
});

test('a text batch hides media controls even when a book still carries an old production mode', () => {
  assert.match(bookSource, /batchFactoryProductionModeVisibility\(batchPatch\.productionMode === 'text' \? 'text' : form\.productionMode\)/);
});

test('text book settings omit the media tab and the video and comic switches', () => {
  assert.ok(/\.\.\.\(visibility\.text \? \[\] : \[\{ key: 'media', label: '画面与视频', children: mediaBody \}\]\)/.test(bookSource), 'text book settings should omit the media tab');
  assert.ok(/const engineTab = visibility\.text && \(selectedEngineTab === 'media' \|\| activeRegion === 'media'\) \? 'models' : selectedEngineTab/.test(bookSource), 'stale media tabs and direct entry should fall back to model settings');
  assert.ok(/const mediaBody = visibility\.text \? null : <div className="batch-factory-media-setting-stack">/.test(bookSource), 'direct access to the media region should hide its mode switches');
});

test('text book settings cannot save a production mode from stale media controls', () => {
  const updateSource = bookSource.match(/^export (function buildBookRegionUpdate\b[\s\S]*?^})/m)?.[1];
  const fieldsSource = bookSource.match(/^const ENGINE_MODEL_OVERRIDE_FIELDS = .+$/m)?.[0];
  const equalSource = bookSource.match(/^function equal\b.+$/m)?.[0];
  assert.ok(updateSource && fieldsSource && equalSource, 'expected the book settings update and its field definitions');
  const buildBookRegionUpdate = new Function('bookSensitiveModeUpdate',
    [fieldsSource, equalSource, updateSource, 'return buildBookRegionUpdate;'].join('\n')
  )(bookSensitiveModeUpdate);
  const inherited = { productionMode: 'text', textModelId: 'batch-text' };
  const bookPatch = { productionMode: 'video' };
  const edited = { productionMode: 'comic', textModelId: 'book-text' };

  assert.deepEqual(buildBookRegionUpdate(inherited, bookPatch, 'engine', edited), {
    patch: { textModelId: 'book-text' },
    restoreKeys: []
  });
  assert.deepEqual(buildBookRegionUpdate(inherited, bookPatch, 'engine', edited, 'media'), {
    patch: {},
    restoreKeys: []
  });
  assert.deepEqual(buildBookRegionUpdate(inherited, bookPatch, 'media', edited), {
    patch: {},
    restoreKeys: []
  });
});

test('text unified settings omit AI reasoning and return to the models from a stale reasoning tab', () => {
  assert.ok(/\.\.\.\(textMode \? \[\] : \[\{ key: 'reasoning', label: 'AI 推理', children: <BatchFactoryAiReasoningForm\b/.test(unifiedSource), 'text unified settings should omit media prompts and constraints');
  assert.ok(/const activeTab = textMode && selectedTab === 'reasoning' \? 'models' : selectedTab/.test(unifiedSource), 'a previously selected reasoning tab should fall back to model settings');
});

test('production mode switch uses distinct visible colors for video and comic', () => {
  assert.match(unifiedSource, /batch-factory-production-mode-switch is-\$\{productionMode\}/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-video[\s\S]*var\(--text-link\)/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-comic[\s\S]*var\(--text-danger\)/);
  assert.match(stylesheet, /\.ant-segmented-item-selected[\s\S]*color:\s*var\(--surface-base\)/);
});

test('comic mode owns comic generation prompts inside AI reasoning', () => {
  assert.match(unifiedSource, /visibility\.comic \? <section className="batch-factory-engine-card"><header><b>漫画生成<\/b>/);
  assert.match(unifiedSource, /<b>漫画提示词<\/b>/);
  assert.match(unifiedSource, /<b>图片描述词推导<\/b>/);
  assert.match(createSource, /productionMode=\{initialBatchSettings\.productionMode\}/);
});

test('single-book settings use the shared production-mode visibility model', () => {
  assert.match(bookSource, /import \{ batchFactoryProductionModeVisibility \} from '\.\/batchFactoryProductionMode'/);
  assert.match(bookSource, /const visibility = batchFactoryProductionModeVisibility\(/);
  assert.match(bookSource, /visibility\.video \? <ConfigCard title="核心视频引擎"/);
  assert.match(bookSource, /visibility\.video \? <ConfigCard title="配音与时长"/);
  assert.match(bookSource, /visibility\.comic \? <section className="batch-factory-engine-card"><header><b>漫画图片约束<\/b>/);
  assert.match(bookSource, /visibility\.comic \? <BatchFactoryComicStylePicker/);
  assert.match(bookSource, /visibility\.video \? <RuleModule required title="视频提示词"/);
});

test('comic rows explain the separate analysis target', () => {
  assert.match(novelListSource, /productionMode === 'comic' \? `正文总字数 \$\{batchFactoryCleanSourceText\(book\.sourceText\)\.length\} 字 · 漫画分析 \$\{comicAnalysisCharactersForBook\(book\)\} 字` : `生产前 \$\{rangeLines\} 行`/);
  assert.match(novelListSource, /batchProductionMode === 'video' \? <label className="batch-factory-engine-field"><span><b>内容范围<\/b><small>生产使用的正文行数<\/small>/);
  assert.match(novelListSource, /<b>漫画分析<\/b><small>按完整句子和段落边界取整 · 预计约 \{comicEstimatedPageRange\(comicAnalysisCharacters\)\.min\}-\{comicEstimatedPageRange\(comicAnalysisCharacters\)\.max\} 张<\/small>/);
  assert.match(novelListSource, /editingContentProductionMode === 'video' \? <label className="shuihuo-form-label"><span>内容范围<small>本次生产使用的正文行数<\/small>/);
  assert.match(novelListSource, /runtimeResolveBookEditableContent\(currentBook, productionMode\)/);
});

test('book actions show only the active production mode controls', () => {
  assert.match(novelListSource, /productionMode === 'comic' \? <div className="batch-factory-action-group is-production"><span>漫画获取<\/span>[\s\S]*? : <div className="batch-factory-action-group is-production"><span>视频获取<\/span>/);
  assert.match(novelListSource, /漫画获取[\s\S]*?runBookStageAction\(book, 'comic', 'bootstrap'\)/);
  assert.match(novelListSource, /视频获取[\s\S]*?runBookStageAction\(book, 'director', 'compile'\)/);
});

test('comic automation uses comic-specific run mode labels', () => {
  assert.match(novelListSource, /batchProductionMode === 'comic' \? \[\{ value: 'storyboard_only', label: '只生成漫画分页与提示词' \}, \{ value: 'video_no_submit', label: '生成漫画合集不上传' \}, \{ value: 'full_submit', label: '全自动生成并提交' \}\]/);
});

test('comic automation never saves or reports the video content-line range', () => {
  assert.match(novelListSource, /if \(batchProductionMode === 'video'\) metadata\.contentRangeLines = rangeLines/);
  assert.match(novelListSource, /batchProductionMode === 'comic'\s*\? `内容设置已保存：漫画分析 \$\{analysisCharacters\} 字，内容截取 \$\{captureCharacters\} 字。`/);
  assert.match(novelListSource, /batchProductionMode === 'comic' \? '保存正文截取设置' : '保存内容设置'/);
});

test('comic reasoning lets source content determine the page count', () => {
  assert.doesNotMatch(unifiedSource, /目标漫画页数/);
  assert.doesNotMatch(unifiedSource, /targetPageCount/);
});

test('comic workbench uses comic-only table and status labels', () => {
  assert.match(novelListSource, /batchProductionMode === 'comic' && index === 4 \? '漫画页' : column\.label/);
  assert.match(novelListSource, /productionMode === 'comic' \? `漫画页 \$\{String\(selectedIndex \+ 1\)\.padStart\(2, '0'\)\}` : storyboardVideoLabel\(video, selectedIndex\)/);
  assert.match(novelListSource, /batchProductionMode === 'comic' \? <span><b>\{mediaCounts\.storyboards\}<\/b> 漫画分页<\/span> : <>/);
});
