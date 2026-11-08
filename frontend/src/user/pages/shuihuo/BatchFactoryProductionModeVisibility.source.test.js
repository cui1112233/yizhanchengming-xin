import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const unifiedSource = fs.readFileSync(path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const bookSource = fs.readFileSync(path.join(directory, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
const createSource = fs.readFileSync(path.join(directory, 'BatchFactoryCreateModal.jsx'), 'utf8');
const novelListSource = fs.readFileSync(path.join(directory, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(directory, '..', 'shuihuo-production.css'), 'utf8');

test('unified settings use the shared production-mode visibility model', () => {
  assert.match(unifiedSource, /import \{ batchFactoryProductionModeVisibility \} from '\.\/batchFactoryProductionMode'/);
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
  assert.match(unifiedSource, /options=\{\[\{ value: 'video', label: '视频生成' \}, \{ value: 'comic', label: '漫画生成' \}\]\}/);
  assert.match(unifiedSource, /<span>统一配置<\/span><BatchFactoryProductionModeControl/);
  assert.doesNotMatch(unifiedSource, /key: 'media', label: '画面与视频'/);
  assert.match(createSource, /<span>新建批量 · 统一配置<\/span><BatchFactoryProductionModeControl/);
  assert.doesNotMatch(createSource, /key: 'media', label: '画面与视频'/);
});

test('production mode switch uses distinct visible colors for video and comic', () => {
  assert.match(unifiedSource, /batch-factory-production-mode-switch is-\$\{productionMode\}/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-video[\s\S]*#1677ff/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-comic[\s\S]*#e85d75/);
  assert.match(stylesheet, /\.ant-segmented-item-selected[\s\S]*color:\s*#fff/);
});

test('comic mode owns comic generation prompts inside AI reasoning', () => {
  assert.match(unifiedSource, /visibility\.comic \? <section className="batch-factory-engine-card"><header><b>漫画生成<\/b>/);
  assert.match(unifiedSource, /<b>漫画提示词<\/b>/);
  assert.match(unifiedSource, /<b>图片描述词推导<\/b>/);
  assert.match(createSource, /productionMode=\{initialBatchSettings\.productionMode\}/);
});

test('single-book settings use the shared production-mode visibility model', () => {
  assert.match(bookSource, /import \{ batchFactoryProductionModeVisibility \} from '\.\/batchFactoryProductionMode'/);
  assert.match(bookSource, /const visibility = batchFactoryProductionModeVisibility\(form\.productionMode\)/);
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
