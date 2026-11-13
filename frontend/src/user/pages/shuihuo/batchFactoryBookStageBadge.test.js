import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryBookStageBadge } from './batchFactoryBookStageBadge.js';

const book = { id: 'book-1' };
const missing = { comicImageTotalCount: 2, comicImageReadyCount: 0, runs: [{ stage: 'comic', status: 'succeeded' }] };
test('planned comic pages without a running task are pending, not generating', () => {
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: missing }).label, '待生成 0/2');
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, comicImageReadyCount: 1 } }).label, '待继续生成 1/2');
});
test('only actual current image tasks show comic generation', () => {
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, comicImageGeneratingCount: 1 } }).label, '漫画生成中 0/2');
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: missing, automationState: 'running', automationBook: { status: 'running', stage: 'assets' } }).label, '资产生成中');
});
test('paused and stopped orchestration do not present an unfinished comic as running', () => {
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: missing, automationState: 'paused', automationBook: { status: 'running', stage: 'comic_images' } }).label, '已暂停');
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: missing, automationState: 'stopped', automationBook: { status: 'running', stage: 'comic_images' } }).label, '本轮已结束');
});
test('page failures and expired tasks expose recovery status', () => {
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, comicImageFailedCount: 1 } }).tone, 'failed');
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, comicImageInterruptedCount: 1 } }).label, '生成已中断 0/2');
});
test('video image stages do not invent a comic page count', () => {
  assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { runs: [{ stage: 'image', status: 'running' }] } }).label, '图片生成中');
});
test('existing primary images or collections do not hide a new generation or its failure', () => {
  for (const completed of [{ comicImageReady: true }, { comicCollectionReady: true }]) {
    assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, ...completed, comicImageReadyCount: 2, comicImageGeneratingCount: 1 } }).tone, 'running');
    assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, ...completed, comicImageFailedCount: 1 } }).tone, 'failed');
    assert.equal(batchFactoryBookStageBadge(book, { stageSummary: { ...missing, ...completed, comicImageInterruptedCount: 1 } }).tone, 'failed');
  }
});
