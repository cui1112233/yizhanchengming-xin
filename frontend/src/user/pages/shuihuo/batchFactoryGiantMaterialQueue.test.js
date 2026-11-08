import assert from 'node:assert/strict';
import test from 'node:test';
import {
  createGiantMaterialPlatformBookClaims,
  createGiantMaterialQueue,
  duplicateGiantMaterialBookIndexes,
  parseGiantMaterialIds,
  queueSummary,
  runSequentialGiantMaterialQueue
} from './batchFactoryGiantMaterialQueue.js';

test('marks only later rows with the same resolved Book ID as duplicates', () => {
  const duplicates = duplicateGiantMaterialBookIndexes([
    { platformBookId: '7703405385419328574', title: '同名书' },
    { platformBookId: '7703405385419328574', title: '同名书' },
    { platformBookId: 'other', title: '同名书' },
    { platformBookId: '', title: '同名书' }
  ]);

  assert.deepEqual([...duplicates], [1]);
});

test('claims a resolved platform book only once per submission regardless of title', () => {
  const claims = createGiantMaterialPlatformBookClaims();
  assert.deepEqual(claims.claim({ platformBookId: '748725', title: '旧标题', giantMaterialId: '7689285397448523826' }), {
    accepted: true, platformBookId: '748725', duplicateOfMaterialId: ''
  });
  assert.deepEqual(claims.claim({ bookId: '748725', title: '新标题', giantMaterialId: '7613606077155459091' }), {
    accepted: false, platformBookId: '748725', duplicateOfMaterialId: '7689285397448523826'
  });
});

test('parses newline, comma and Chinese punctuation while deduplicating IDs', () => {
  const parsed = parseGiantMaterialIds(' 7689285397448523826\n7613606077155459091, 7689285397448523826；bad ');
  assert.deepEqual(parsed.valid, ['7689285397448523826', '7613606077155459091']);
  assert.deepEqual(parsed.invalid, ['bad']);
});

test('creates visible pending items with stable IDs', () => {
  const queue = createGiantMaterialQueue('7689285397448523826\n7613606077155459091');
  assert.deepEqual(queue.map(item => ({ id: item.id, status: item.status, stage: item.stage })), [
    { id: '7689285397448523826', status: 'pending', stage: 'idle' },
    { id: '7613606077155459091', status: 'pending', stage: 'idle' }
  ]);
});

test('runs one item at a time and keeps a failed item retryable', async () => {
  const queue = createGiantMaterialQueue('7689285397448523826\n7613606077155459091');
  const events = [];
  let active = 0;
  let maxActive = 0;
  const result = await runSequentialGiantMaterialQueue(queue, async item => {
    active += 1;
    maxActive = Math.max(maxActive, active);
    events.push(`start:${item.id}`);
    await Promise.resolve();
    active -= 1;
    if (item.id.endsWith('091')) throw new Error('OCR_NO_TEXT');
    return { title: 'ok' };
  }, { onState: (item, state) => events.push(`${state.status}:${item.id}`) });

  assert.equal(maxActive, 1);
  assert.deepEqual(events, [
    'running:7689285397448523826',
    'start:7689285397448523826',
    'success:7689285397448523826',
    'running:7613606077155459091',
    'start:7613606077155459091',
    'error:7613606077155459091'
  ]);
  assert.equal(result[0].status, 'success');
  assert.equal(result[1].status, 'error');
  assert.equal(result[1].error, 'OCR_NO_TEXT');
  assert.deepEqual(queueSummary(result), { total: 2, pending: 0, running: 0, success: 1, skipped: 0, error: 1 });
});
