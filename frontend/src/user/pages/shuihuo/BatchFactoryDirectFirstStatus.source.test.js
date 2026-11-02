import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('does not describe a failed direct-first read as an automatic OCR queue', () => {
  assert.match(source, /书城读取失败/);
  assert.match(source, /可手动重试或使用 OCR/);
  assert.doesNotMatch(source, /已转入 OCR 等待队列/);
  assert.doesNotMatch(source, /失败后自动转为滚屏 OCR/);
});

test('counts only active direct or OCR reads in the background-read summary', () => {
  assert.match(source, /function isBatchFactoryContentReadRunning\(book = \{\}\)/);
  assert.match(source, /stage === 'direct' \|\| stage === 'ocr'/);
  assert.match(source, /const pendingContentBooks = books\.filter\(isBatchFactoryContentReadRunning\)/);
  assert.match(source, /const giantWaitingBooks = books\.filter\(book => book\?\.sourceMetadata\?\.giantMaterialId && isBatchFactoryContentReadRunning\(book\)\)/);
});
