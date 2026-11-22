import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const selectedProductionSource = source.slice(
  source.indexOf('async function runSelectedProduction()'),
  source.indexOf('function mergePayload(', source.indexOf('async function runSelectedProduction()'))
);

test('video immediate selected execution starts the complete automation pipeline instead of only the video stage', () => {
  assert.match(selectedProductionSource, /setAutomationBookScope\('selected'\)/);
  assert.match(selectedProductionSource, /ensureBatchExecution\(batch\.id,/);
  assert.doesNotMatch(selectedProductionSource, /runAutomationAction\('start'/);
  assert.doesNotMatch(selectedProductionSource, /runBookStage\(batch\.id, book\.id, 'video'/);
  assert.match(source, /const batchMenuItems = batchProductionMode === 'comic'[\s\S]*?: \[[\s\S]*?key: 'selected-production'[\s\S]*?批量生成视频/);
});

test('selected execution ignores an active preset from the other production mode', () => {
  assert.match(selectedProductionSource, /const modePresets = automationPresetsForMode\(presets, batchProductionMode\)/);
  assert.match(selectedProductionSource, /modePresets\.some\(item => String\(item\.id\) === activePresetID\)/);
  assert.match(selectedProductionSource, /modePresets\.some\(item => String\(item\.id\) === selectedPresetID\)/);
});

test('selected immediate execution remains available when the video capability reports a configuration problem', () => {
  assert.match(source, /key: 'selected-production', label: automationActive \? `排队执行已选（\$\{selectedBookIds\.length\}）` : `立即执行已选（\$\{selectedBookIds\.length\}）`, disabled: !selectedBookIds\.length/);
  assert.doesNotMatch(source, /key: 'selected-production',[\s\S]*?disabled: !productionCapability\.available \|\| !selectedBookIds\.length/);
});

test('automatic production refuses to start when the URL batchId differs from the open batch', () => {
  assert.match(source, /new URLSearchParams\(window\.location\.search\)\.get\('batchId'\)/);
  assert.match(source, /页面批次与当前网址不一致/);
  assert.match(source, /if \(key === 'start' && urlBatchId && urlBatchId !== String\(batch\.id\)\) \{[\s\S]*return;/);
});
