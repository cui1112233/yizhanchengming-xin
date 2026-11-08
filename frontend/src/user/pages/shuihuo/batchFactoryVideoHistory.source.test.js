import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { batchFactoryNovelCopyText } from './batchFactoryRuntimeLogic.js';
import { batchFactoryRuntimePollDelay, comicRowResultPollDelay, finishComicPageGeneration } from './batchFactoryPerformance.js';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const runtimeLogicSource = fs.readFileSync(path.join(here, 'batchFactoryRuntimeLogic.js'), 'utf8');

test('keeps generated videos from an earlier round visible as a selectable previous version', () => {
  // 历史轮次（orphaned）的成功媒体必须挂回同一分镜并标注「上一版」，不能随轮换消失。
  assert.match(source, /const mediaVersions = bookVideoVersions\(book, productionStatus\)/);
  assert.match(source, /version\.isHistory \? ' · 上一版' : ''/);
  assert.match(source, /main\.isHistory \? ' · 上一版' : ''/);
  // 历史记录不属于当前片段库，不提供“移除候选”操作。
  assert.match(source, /!isPrimary && !isUpload && !version\.isHistory/);
  // 上传来源只允许当前 active 记录下的任务，避免 video/task 错配。
  assert.match(source, /state\.main\?\.mediaUrl && !state\.main\.isHistory/);
});

test('copies every book as title / source id / platform in tab-separated lines', () => {
  const books = [
    { title: '拒绝攻略后女主她发癫了', bookId: '2099304888681359141', platform: 'zhihu' },
    { title: '绝对AA制后室友崩溃了', bookId: '1922690039302845551', platform: 'zhihu' },
    { title: '缺字段的书' },
  ];
  const text = batchFactoryNovelCopyText(books, { zhihu: '知乎' });
  const lines = text.split('\n');
  // 顺序沿用列表序号，行数与书数一致；每条始终三列，用制表符分隔。
  assert.equal(lines.length, books.length);
  assert.equal(lines[0], '拒绝攻略后女主她发癫了\t2099304888681359141\t知乎');
  assert.equal(lines[1], '绝对AA制后室友崩溃了\t1922690039302845551\t知乎');
  assert.equal(lines[2], '缺字段的书\t\t');
});

test('the copy-all-books icon lives in the novel header and never triggers row clicks', () => {
  assert.match(source, /复制全部书名 \/ 书城ID \/ 书城/);
  assert.match(source, /batchFactoryNovelCopyText\(books, platformNames\)/);
  assert.match(source, /event\?\.stopPropagation\?\.\(\)/);
  assert.match(source, /已复制 \$\{books\.length\} 本书/);
  assert.match(source, /复制失败，请检查浏览器剪贴板权限后重试/);
});

test('runtime summary polls every 8 seconds and pauses while the page is hidden', () => {
  assert.equal(batchFactoryRuntimePollDelay('running', 0), 8000);
  assert.match(source, /document\.hidden/);
  assert.match(source, /visibilitychange/);
  assert.match(source, /clearTimeout/);
});

test('an unfinished comic image is checked every second', () => {
  assert.equal(comicRowResultPollDelay([{ generationState: 'generating', images: [] }]), 1000);
  assert.equal(comicRowResultPollDelay([{ generationState: 'generating', images: [{ isPrimary: true }] }]), 0);
});

test('revealing a finished comic image does not wait for the full batch refresh', async () => {
  let releaseRefresh;
  const refreshBlocked = new Promise(resolve => { releaseRefresh = resolve; });
  const events = [];
  const result = await finishComicPageGeneration({
    waitForResult: async () => ({ id: 'page-1' }),
    revealResult: page => events.push(`visible:${page.id}`),
    refreshBatch: async () => { events.push('refresh-started'); await refreshBlocked; },
  });
  assert.deepEqual(result, { id: 'page-1' });
  assert.deepEqual(events, ['visible:page-1', 'refresh-started']);
  releaseRefresh();
});

test('videoHistory carries playable media fields as a fallback source of versions', () => {
  assert.match(runtimeLogicSource, /historyVideo\?\.productionTaskId/);
  assert.match(runtimeLogicSource, /normalizedText\(historyVideo\?\.mediaUrl\)/);
  assert.match(runtimeLogicSource, /productionTaskId: taskId/);
});
