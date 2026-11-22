import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryUploadIntent } from './batchFactoryUploadIntent.js';

const uploaded = book => book.status === 'uploaded';

test('all uploaded books use the explicit reupload path', () => {
  const result = batchFactoryUploadIntent([{ id: 'a', status: 'uploaded' }, { id: 'b', status: 'uploaded' }], uploaded);
  assert.equal(result.kind, 'reupload');
});

test('uploaded and fresh books cannot be submitted together', () => {
  const result = batchFactoryUploadIntent([{ id: 'a', status: 'uploaded' }, { id: 'b', status: 'fresh' }], uploaded);
  assert.equal(result.kind, 'mixed');
  assert.deepEqual(result.uploaded.map(book => book.id), ['a']);
  assert.deepEqual(result.fresh.map(book => book.id), ['b']);
});
