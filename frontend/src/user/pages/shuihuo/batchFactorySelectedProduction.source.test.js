import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('video bulk actions can immediately run only the books selected in the novel list', () => {
  assert.match(source, /async function runSelectedProduction\(\)/);
  assert.match(source, /const selectedBooks = books\.filter\(book => selectedBookIds\.includes\(book\.id\)\);/);
  assert.match(source, /for \(const book of selectedBooks\)[\s\S]*?await runBookStage\(batch\.id, book\.id, 'video'/);
  assert.match(source, /const batchMenuItems = batchProductionMode === 'comic'[\s\S]*?: \[[\s\S]*?key: 'selected-production'[\s\S]*?批量生成视频/);
});

test('selected immediate execution remains available when the video capability reports a configuration problem', () => {
  assert.match(source, /key: 'selected-production', label: `立即执行已选（\$\{selectedBookIds\.length\}）`, disabled: !selectedBookIds\.length/);
  assert.doesNotMatch(source, /key: 'selected-production',[\s\S]*?disabled: !productionCapability\.available \|\| !selectedBookIds\.length/);
});

test('automatic production refuses to start when the URL batchId differs from the open batch', () => {
  assert.match(source, /new URLSearchParams\(window\.location\.search\)\.get\('batchId'\)/);
  assert.match(source, /页面批次与当前网址不一致/);
  assert.match(source, /if \(key === 'start' && urlBatchId && urlBatchId !== String\(batch\.id\)\) \{[\s\S]*return;/);
});
