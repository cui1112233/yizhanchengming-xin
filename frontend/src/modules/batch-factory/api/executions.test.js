import assert from 'node:assert/strict';
import test from 'node:test';

import { ensureBatchExecution, getBatchExecution, getBatchExecutionTimeline, reconcileBatchExecution, cancelBatchExecution } from './executions.js';

function recorder() {
  const calls = [];
  return { calls, request: async (...args) => { calls.push(args); return { ok: true }; } };
}

test('execution api encodes ids and sends stable methods and bodies', async () => {
  const one = recorder();
  await ensureBatchExecution('batch/一', { bookIds: ['book-1'] }, one.request);
  assert.deepEqual(one.calls[0], ['/api/batch-factory/v12/batches/batch%2F%E4%B8%80/executions/ensure', { method: 'POST', body: JSON.stringify({ bookIds: ['book-1'] }) }]);
  await getBatchExecution('exec/一', one.request);
  assert.equal(one.calls[1][0], '/api/batch-factory/v12/executions/exec%2F%E4%B8%80');
  await getBatchExecutionTimeline('exec/一', 'book/二', one.request);
  assert.equal(one.calls[2][0], '/api/batch-factory/v12/executions/exec%2F%E4%B8%80/books/book%2F%E4%BA%8C/timeline');
  await reconcileBatchExecution('exec/一', one.request);
  assert.equal(one.calls[3][1].method, 'POST');
  await cancelBatchExecution('exec/一', one.request);
  assert.match(one.calls[4][0], /\/cancel$/);
});
