import test from 'node:test';
import assert from 'node:assert/strict';

import {
  automationFailureNotificationsReady,
  automationFailureSummaryContent,
  collectNewAutomationFailures
} from './batchFactoryFailureNotice.js';

function resolveFrom(map) {
  return (_status, _batchId, bookId) => map[bookId] || null;
}

function book(id) {
  return { id, title: id, bookId: id };
}

test('3 本 fetch_original 失败只汇总成 1 条提示', () => {
  const books = [book('b1'), book('b2'), book('b3')];
  const resolveEntry = resolveFrom({
    b1: { status: 'failed', stage: 'fetch_original', error: '自动获取正文失败：121 upstream business error 400: 参数不完整' },
    b2: { status: 'failed', stage: 'fetch_original', error: '自动获取正文失败：121 upstream business error 400: 参数不完整' },
    b3: { status: 'failed', stage: 'fetch_original', error: '自动获取正文失败：121 upstream business error 400: 参数不完整' }
  });
  const notified = {};
  const entries = collectNewAutomationFailures({ books, automationStatus: {}, batchId: 'batch-1', notified, resolveEntry });

  assert.equal(entries.length, 3);
  const content = automationFailureSummaryContent(entries);
  assert.equal(content, '3 本书自动生产失败（获取原文等）');
  assert.deepEqual(Object.keys(notified).sort(), ['b1', 'b2', 'b3']);
});

test('同一失败签名不会重复进入汇总', () => {
  const books = [book('b1')];
  const resolveEntry = resolveFrom({
    b1: { status: 'failed', stage: 'fetch_original', error: '121 upstream business error 400: 参数不完整' }
  });
  const notified = { b1: 'fetch_original:121 upstream business error 400: 参数不完整' };
  const entries = collectNewAutomationFailures({ books, automationStatus: {}, batchId: 'batch-1', notified, resolveEntry });
  assert.equal(entries.length, 0);
});

test('单本 fetch_original 失败使用中文阶段名，不再出现裸 stage key', () => {
  const books = [book('b1')];
  const resolveEntry = resolveFrom({
    b1: { status: 'failed', stage: 'fetch_original', error: '自动获取正文失败：参数不完整' }
  });
  const entries = collectNewAutomationFailures({ books, automationStatus: {}, batchId: 'batch-1', notified: {}, resolveEntry });
  assert.equal(entries.length, 1);
  const content = automationFailureSummaryContent(entries);
  assert.equal(content, '获取原文失败');
  assert.doesNotMatch(content, /fetch_original/);
});

test('成功场景不产生任何汇总提示', () => {
  const books = [book('b1'), book('b2')];
  const resolveEntry = resolveFrom({
    b1: { status: 'completed', stage: 'video' },
    b2: { status: 'running', stage: 'video' }
  });
  const entries = collectNewAutomationFailures({ books, automationStatus: {}, batchId: 'batch-1', notified: {}, resolveEntry });
  assert.equal(entries.length, 0);
  assert.equal(automationFailureSummaryContent(entries), '');
});

test('首屏已有失败只记录不弹提示', () => {
  const books = [book('b1')];
  const resolveEntry = resolveFrom({
    b1: { status: 'failed', stage: 'fetch_original', error: '参数不完整' }
  });
  const notified = {};
  const entries = collectNewAutomationFailures({ books, automationStatus: {}, batchId: 'batch-1', notified, resolveEntry, firstPass: true });
  assert.equal(entries.length, 0);
  assert.equal(notified.b1, 'fetch_original:参数不完整');
});

test('状态尚未返回书目时，不把随后加载的历史失败当作新失败', () => {
  assert.equal(automationFailureNotificationsReady({ state: 'idle', counts: { total: 0 } }), false);
  assert.equal(automationFailureNotificationsReady({ state: 'ready', books: [] }), true);
});

test('标记为历史的失败不进入顶部汇总', () => {
  const entries = collectNewAutomationFailures({
    books: [book('b1')],
    automationStatus: {},
    batchId: 'batch-1',
    notified: {},
    resolveEntry: resolveFrom({
      b1: { status: 'failed', stage: 'merge', error: '历史合成失败', stale: true }
    })
  });

  assert.equal(entries.length, 0);
});
