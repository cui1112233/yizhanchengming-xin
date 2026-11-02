import test from 'node:test';
import assert from 'node:assert/strict';
import { novelFetchProgressLines } from './novel-fetch-progress.js';

test('formats live progress with time, book name, id and message', () => {
  const lines = novelFetchProgressLines({
    progress: [
      { time: '2026-10-18T03:04:05Z', book_id: '12345678901', book_name: '测试书名', message: '开始处理' },
      { time: '2026-10-18T03:04:06Z', book_id: '12345678901', book_name: '测试书名', message: '处理完成' }
    ]
  });
  assert.equal(lines.length, 2);
  assert.match(lines[0], /测试书名/);
  assert.match(lines[0], /12345678901/);
  assert.match(lines[0], /开始处理/);
  assert.match(lines[1], /处理完成/);
});

test('falls back to job state when no progress entries exist', () => {
  assert.deepEqual(novelFetchProgressLines({ status: 'queued' }), ['任务已进入队列，等待执行。']);
  assert.deepEqual(novelFetchProgressLines(null), ['暂无正在处理的任务。']);
});

test('aggregates live logs for every book in the current batch', () => {
  const lines = novelFetchProgressLines(
    { progress: [{ time: '2026-10-18T03:04:05Z', book_id: 'book-a', book_name: '第一本', message: '处理完成' }] },
    [
      { book_id: 'book-a', book_name: '第一本', logs: [{ time: '2026-10-18T03:04:05Z', detail: '处理完成' }] },
      { book_id: 'book-b', book_name: '第二本', logs: [{ time: '2026-10-18T03:04:06Z', detail: '开始处理' }] },
      { book_id: 'book-c', book_name: '第三本', logs: [{ time: '2026-10-18T03:04:07Z', detail: '处理完成' }] }
    ]
  );
  assert.equal(lines.some(line => line.includes('第一本')), true);
  assert.equal(lines.some(line => line.includes('第二本')), true);
  assert.equal(lines.some(line => line.includes('第三本')), true);
  assert.equal(lines.filter(line => line.includes('第一本') && line.includes('处理完成')).length, 1);
});

test('deduplicates logs that render identically across live and persisted shapes', () => {
  const lines = novelFetchProgressLines(
    { progress: [{ time: '2026-10-18T03:04:05Z', book_id: 'book-a', book_name: '第一本', message: '处理完成' }] },
    [{ book_id: 'book-a', book_name: '第一本', logs: [{ at: '2026-10-18T03:04:05Z', event: '处理完成' }] }]
  );
  assert.equal(lines.filter(line => line.includes('第一本') && line.includes('处理完成')).length, 1);
});

test('merges same-book persisted stage events with live progress in chronological order', () => {
  const lines = novelFetchProgressLines(
    { progress: [{ time: '2026-10-18T03:04:06Z', book_id: 'book-a', book_name: '第一本', message: '处理中' }] },
    [{ book_id: 'book-a', book_name: '第一本', logs: [
      { time: '2026-10-18T03:04:05Z', stage: 'fetch', status: 'started', event: 'fetch_started', detail: '开始获取原文' },
      { time: '2026-10-18T03:04:07Z', stage: 'sensitive_processing', status: 'completed', event: 'sensitive_processing_completed', detail: '敏感词处理结果已保存' }
    ] }]
  );
  assert.equal(lines.length, 3);
  assert.match(lines[0], /开始获取原文/);
  assert.match(lines[1], /处理中/);
  assert.match(lines[2], /敏感词处理结果已保存/);
});
