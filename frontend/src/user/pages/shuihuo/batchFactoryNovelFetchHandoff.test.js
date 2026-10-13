import test from 'node:test';
import assert from 'node:assert/strict';
import {
  batchFactoryNovelFetchHandoffPath,
  novelFetchIntakeBooks,
  pendingNovelFetchIntakeId
} from './batchFactoryNovelFetchHandoff.js';

test('Novel Fetch always hands off to the dedicated Batch Factory entry', () => {
  assert.equal(
    batchFactoryNovelFetchHandoffPath('intake/a b'),
    '/shuihuo-production?intake=intake%2Fa%20b'
  );
});

test('only a valid intake query activates a Novel Fetch handoff', () => {
  assert.equal(pendingNovelFetchIntakeId('?intake=intake-1'), 'intake-1');
  assert.equal(pendingNovelFetchIntakeId('?other=intake-1'), '');
  assert.equal(pendingNovelFetchIntakeId('?intake='), '');
});

test('reads the frozen books from the V12 intake payload without re-fetching source text', () => {
  assert.deepEqual(
    novelFetchIntakeBooks({ payload: JSON.stringify({ books: [{ bookId: 'source-1', title: 'AI1·原书', sourceText: '已确认正文' }] }) }),
    [{ bookId: 'source-1', title: 'AI1·原书', sourceText: '已确认正文' }]
  );
});
