import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const unifiedSource = fs.readFileSync(path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const bookSource = fs.readFileSync(path.join(directory, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
const createSource = fs.readFileSync(path.join(directory, 'BatchFactoryCreateModal.jsx'), 'utf8');

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
