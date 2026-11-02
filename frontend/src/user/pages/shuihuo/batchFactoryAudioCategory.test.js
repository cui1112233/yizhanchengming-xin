import assert from 'node:assert/strict';
import test from 'node:test';

import { inferBatchN8AudioCategory } from './batchFactoryAudioCategory.js';

const categories = [
  { id: 9, name: '女-爽文（打脸）', status: 'ENABLE' },
  { id: 10, name: '古风-女悲伤', status: 'ENABLE' },
  { id: 11, name: '古风-男悲伤', status: 'ENABLE' },
  { id: 12, name: '古风-爽文（打脸）', status: 'ENABLE' }
];

test('maps a fully classified all-female revenge batch to the exact N8 category', () => {
  const result = inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '现代爽文' } },
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '女主打脸爽文' } }
  ], categories);
  assert.equal(result.category?.id, 9);
  assert.equal(result.reason, 'matched');
});

test('does not infer when books disagree, metadata is incomplete, or style is unsupported', () => {
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '爽文' } },
    { sourceMetadata: { classifyStatus: 'classified', gender: '男频', style: '爽文' } }
  ], categories).reason, 'mixed');
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'pending_confirmation', gender: '女频', style: '爽文' } }
  ], categories).reason, 'unclassified');
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '现代悬疑' } }
  ], categories).reason, 'unsupported');
});
