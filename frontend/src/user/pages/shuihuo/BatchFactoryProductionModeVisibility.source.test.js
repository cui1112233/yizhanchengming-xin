import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const unifiedSource = fs.readFileSync(path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const bookSource = fs.readFileSync(path.join(directory, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');

test('unified settings use the shared production-mode visibility model', () => {
  assert.match(unifiedSource, /import \{ batchFactoryProductionModeVisibility \} from '\.\/batchFactoryProductionMode'/);
  assert.match(unifiedSource, /const visibility = batchFactoryProductionModeVisibility\(value\.productionMode\)/);
  assert.match(unifiedSource, /visibility\.video && sections\.includes\('audio'\)/);
});

test('unified media and reasoning render only the active mode details', () => {
  assert.match(unifiedSource, /visibility\.video \? <div className="batch-factory-ai-module-body">/);
  assert.match(unifiedSource, /visibility\.comic \? <div className="batch-factory-ai-module-body">/);
  assert.match(unifiedSource, /visibility\.comic \? <BatchFactoryComicStylePicker/);
  assert.match(unifiedSource, /visibility\.video \? <section className="batch-factory-engine-card"><header><b>视频设置<\/b>/);
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
