const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const { createNovelFetchBatches } = require('../lib/novel-fetch-workshop/batches');
const { createWorkshopTasks } = require('../lib/novel-fetch-workshop/tasks');
const { attachBatch } = require('../lib/novel-fetch-workshop/v2-api-contract');
const { runNovelFetchBatch } = require('../lib/novel-fetch-workshop/runner');

test('completed batch keeps only its submitted task ids and input order', () => {
  const usersDir = fs.mkdtempSync(path.join(os.tmpdir(), 'novel-fetch-batches-'));
  const batches = createNovelFetchBatches({
    usersDir,
    clock: () => new Date('2026-09-20T04:00:00.000Z')
  });
  const submitted = ['book-8', 'book-2', 'book-5', 'book-1', 'book-7', 'book-3', 'book-6', 'book-4'];
  const batch = batches.create('tester', { taskIds: submitted });

  batches.complete('tester', batch.id, {
    tasks: [...submitted, 'old-9', 'old-10', 'old-11'].map(bookId => ({
      bookId,
      status: 'done',
      originalStatus: 'done'
    }))
  });

  const current = batches.current('tester');
  assert.deepEqual(current.taskIds, submitted);
  assert.deepEqual(current.taskStates.map(task => task.bookId), submitted);
});

test('legacy empty batch adopts only the ids parsed for that run', () => {
  const usersDir = fs.mkdtempSync(path.join(os.tmpdir(), 'novel-fetch-batches-empty-'));
  const batches = createNovelFetchBatches({ usersDir, clock: () => new Date('2026-09-20T04:00:00.000Z') });
  const batch = batches.create('tester', { taskIds: [] });

  batches.complete('tester', batch.id, {
    current_batch_ids: ['new-1', 'new-2'],
    tasks: [
      { bookId: 'new-1', status: 'done' },
      { bookId: 'new-2', status: 'queued' },
      { bookId: 'old-1', status: 'done' }
    ]
  });

  const current = batches.current('tester');
  assert.deepEqual(current.taskIds, ['new-1', 'new-2']);
  assert.deepEqual(current.taskStates.map(task => task.bookId), ['new-1', 'new-2']);
});

test('queue payload carries the persisted batch creation time', () => {
  const usersDir = fs.mkdtempSync(path.join(os.tmpdir(), 'novel-fetch-batch-payload-'));
  const batches = createNovelFetchBatches({
    usersDir,
    clock: () => new Date('2026-09-20T04:00:00.000Z')
  });
  const payload = attachBatch('tester', { input_text: '1001\t测试书', task_ids: ['1001'] }, batches);

  assert.equal(payload.batch_created_at, '2026-09-20T04:00:00.000Z');
});

test('task listing recovers batch date from metadata when the index is stale', async () => {
  const usersDir = fs.mkdtempSync(path.join(os.tmpdir(), 'novel-fetch-task-index-'));
  const tasks = createWorkshopTasks({ usersDir, fetchUpstream: async () => ({ code: 200, text: '原文' }) });
  await tasks.saveTasks('tester', [{ bookId: 'old-1', batchId: 'batch-old', batchCreatedAt: '2026-09-18T08:00:00.000Z', batchOrder: 0 }]);
  const indexPath = path.join(usersDir, 'tester', 'novel-fetch-workshop', 'index.json');
  const index = JSON.parse(fs.readFileSync(indexPath, 'utf8'));
  delete index[0].batchId;
  delete index[0].batchCreatedAt;
  fs.writeFileSync(indexPath, JSON.stringify(index), 'utf8');

  const listed = tasks.listTasks('tester');
  assert.equal(listed[0].batchId, 'batch-old');
  assert.equal(listed[0].batchCreatedAt, '2026-09-18T08:00:00.000Z');
});

test('runner result contains only the books parsed for this batch, never the account history', async () => {
  const parsed = [{ bookId: 'new-1', bookName: '本次第一本' }, { bookId: 'new-2', bookName: '本次第二本' }];
  const result = await runNovelFetchBatch({
    username: 'tester',
    payload: { input_text: 'new-1\nnew-2', batch_id: 'batch-current', batch_created_at: '2026-09-24T00:00:00.000Z' },
    configStore: {
      getConfig: () => ({ workflow: { auto_classify_missing: false, auto_fetch_original: false, auto_rewrite_after_fetch: false, auto_submit_after_rewrite: false }, fetch: {}, rewrite: {} }),
      getStyles: () => [], getPlatforms: () => []
    },
    tasks: { saveTasks: async () => {} },
    parseBooks: () => ({ tasks: parsed, duplicateCount: 0, emptyIdCount: 0 }),
    listTasks: async () => [
      { bookId: 'old-1', batchId: 'batch-old' },
      { bookId: 'new-2', batchId: 'batch-current' },
      { bookId: 'new-1', batchId: 'batch-current' }
    ]
  });
  assert.deepEqual(result.tasks.map(task => task.bookId), ['new-1', 'new-2']);
});

test('manual refetch ignores an existing original and fetches a fresh copy', async () => {
  let fetches = 0;
  await runNovelFetchBatch({
    username: 'tester',
    payload: { input_text: 'new-1', force_refetch: true },
    configStore: {
      getConfig: () => ({ workflow: { auto_classify_missing: false, auto_fetch_original: true, auto_rewrite_after_fetch: false, auto_submit_after_rewrite: false }, fetch: {}, rewrite: {} }),
      getStyles: () => [], getPlatforms: () => []
    },
    tasks: {
      saveTasks: async () => {},
      getTask: async () => ({ meta: { bookId: 'new-1', originalStatus: 'done' }, document: { original: '旧正文' } }),
      fetchOriginal: async () => { fetches += 1; return { status: 'done' }; }
    },
    parseBooks: () => ({ tasks: [{ bookId: 'new-1' }], duplicateCount: 0, emptyIdCount: 0 })
  });
  assert.equal(fetches, 1);
});

test('manual refetch does not generate AI copy', async () => {
  let generated = 0;
  await runNovelFetchBatch({
    username: 'tester',
    payload: { input_text: 'new-1', force_refetch: true, skip_rewrite: true },
    configStore: {
      getConfig: () => ({ workflow: { auto_classify_missing: false, auto_fetch_original: false, auto_rewrite_after_fetch: true, auto_submit_after_rewrite: false }, fetch: {}, rewrite: {} }),
      getStyles: () => [], getPlatforms: () => []
    },
    tasks: {
      saveTasks: async () => {},
      getTask: async () => ({ meta: { bookId: 'new-1' } }),
      fetchOriginal: async () => ({ status: 'done' })
    },
    parseBooks: () => ({ tasks: [{ bookId: 'new-1' }], duplicateCount: 0, emptyIdCount: 0 }),
    generateAiVersions: async () => { generated += 1; return { status: 'done', generated: [] }; }
  });
  assert.equal(generated, 0);
});
