import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('comic bulk menu never exposes video generation or VIDEO merge actions', () => {
  assert.match(source, /batchProductionMode === 'comic'[\s\S]*?批量分析漫画分页/);
  assert.match(source, /batchProductionMode === 'comic'[\s\S]*?生成漫画合集/);
  assert.doesNotMatch(source, /const batchMenuItems = \[[\s\S]*?批量生成视频[\s\S]*?合并当前主 VIDEO/);
});

test('selected comic execution starts automation with the selected book ids', () => {
  assert.match(source, /bookIds:\s*selectedBookIds/);
  assert.match(source, /productionMode:\s*batchProductionMode/);
});

test('automation dialog lets the user choose all books or selected books', () => {
  assert.match(source, /执行书籍/);
  assert.match(source, /automationBookScope/);
  assert.match(source, /已选小说（\$\{selectedBookIds\.length\} 本）/);
});

test('comic upload toolbar opens comic publishing instead of the 121 video uploader', () => {
  assert.match(source, /batchProductionMode === 'comic' \? '发布漫画'/);
  assert.match(source, /openComicBatchPublish/);
});

test('per-book comic actions use whole-book preparation and generation', () => {
  assert.match(source, /runBookStageAction\(book, 'comic', 'bootstrap'\)/);
  assert.match(source, /listComicPages\(batch\.id, book\.id/);
  assert.match(source, /const videos = book\.videos \|\| \[\];[\s\S]*?for \(const video of videos\)/);
  assert.match(source, /generateComicPage\(batch\.id, book\.id, video\.id/);
  assert.match(source, /await refreshBatch\(\)/);
  assert.doesNotMatch(source, /await generateComicBook\(batch\.id, book\.id\)/);
  assert.match(source, /function clearLocalExecutorSubmittingToast\(\)/);
  assert.match(source, /finally \{ clearLocalExecutorSubmittingToast\(\); setActionBusy\('\'\); \}/);
  assert.match(source, /重新生成本页/);
  assert.match(source, /generateBookComicPage/);
});
