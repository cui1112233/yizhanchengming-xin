import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryAvailableVersions, batchFactorySourceText, batchFactoryVersionLabel } from './novel-fetch-batch-factory.js';

test('uses the sensitive-processed original by default', () => {
  assert.equal(batchFactorySourceText({ processedText: '已处理正文', originalRaw: '完整原文' }), '已处理正文');
  assert.equal(batchFactorySourceText({ original: '编辑后的原文', originalRaw: '完整原文' }), '编辑后的原文');
});

test('can select AI1 or the full original explicitly', () => {
  const detail = { processedText: '敏感处理后', originalRaw: '完整原文', versions: { ai1: 'AI1改文' } };
  assert.equal(batchFactorySourceText(detail, 'ai1'), 'AI1改文');
  assert.equal(batchFactorySourceText(detail, 'original'), '完整原文');
});

test('only exposes AI versions available for every selected book', () => {
  const details = [{ processedText: '一', versions: { ai1: 'AI1一', ai2: 'AI2一' } }, { processedText: '二', versions: { ai1: 'AI1二' } }];
  assert.deepEqual(batchFactoryAvailableVersions(details), ['sensitive', 'original', 'ai1']);
  assert.equal(batchFactoryVersionLabel('ai1'), 'AI1');
});
