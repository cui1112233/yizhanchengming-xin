import assert from 'node:assert/strict';
import test from 'node:test';
import { singleVersionRetryPayload, submissionPreviewMatches } from './novel-fetch-submit.js';

test('submission preview is valid only for the same selected tasks', () => {
  const plan = { groups: [{}] };
  assert.equal(submissionPreviewMatches(['book-2', 'book-1'], ['book-1', 'book-2'], plan), true);
  assert.equal(submissionPreviewMatches(['book-1', 'book-3'], ['book-1', 'book-2'], plan), false);
  assert.equal(submissionPreviewMatches([], [], plan), false);
  assert.equal(submissionPreviewMatches(['book-1'], ['book-1'], null), false);
});

test('single-version retry keeps the original operation audit id', () => {
  assert.deepEqual(singleVersionRetryPayload({ book_id: 'book-9', version: 'ai3' }), {
    taskIds: ['book-9'],
    versions: ['ai3']
  });
  assert.deepEqual(singleVersionRetryPayload({ book_id: 'book-9', version: 'ai3', operation_id: 'op-1' }), {
    taskIds: ['book-9'],
    versions: ['ai3'],
    operationId: 'op-1'
  });
  assert.equal(singleVersionRetryPayload({ bookId: 'book-9' }), null);
});
