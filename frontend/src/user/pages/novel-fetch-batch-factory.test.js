import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryAvailableVersions, batchFactorySourceMetadata, batchFactoryRawSourceText, batchFactorySourceText, batchFactoryVersionLabel } from './novel-fetch-batch-factory.js';

test('uses the sensitive-processed original by default', () => {
  assert.equal(batchFactorySourceText({ processedText: '已处理正文', originalRaw: '完整原文' }), '已处理正文');
  assert.equal(batchFactorySourceText({ original: '编辑后的原文', originalRaw: '完整原文' }), '编辑后的原文');
});

test('can select AI1 or the full original explicitly', () => {
  const detail = { processedText: '敏感处理后', originalRaw: '完整原文', versions: { ai1: 'AI1改文' } };
  assert.equal(batchFactorySourceText(detail, 'ai1'), 'AI1改文');
  assert.equal(batchFactorySourceText(detail, 'original'), '完整原文');
});

test('keeps complete raw text separate from the selected production version', () => {
  const detail = { processedText: '敏感处理后', originalRaw: '完整原始正文', versions: { ai1: 'AI1改文' } };
  assert.equal(batchFactorySourceText(detail, 'ai1'), 'AI1改文');
  assert.equal(batchFactoryRawSourceText(detail), '完整原始正文');
});

test('carries classification provenance and processing state into the intake', () => {
  const metadata = batchFactorySourceMetadata({
    bookId: 'task-123',
    meta: { genderSource: 'bookinfo', genderReason: '平台分类明确', genderConfidence: 1, styleSource: 'ai', styleReason: '根据正文判断', styleConfidence: 0.91, category: '女频', genre: '现代悬疑', sensitiveStatus: 'done', sensitiveReplaced: 3, sensitiveHits: { hits: [{ snippet: '不得传递的命中正文' }] }, rulesTrace: [{ stage: 'sensitive', title: '敏感词替换', changed: true, chars: 4000, lines: 50, samples: ['不得传递的命中正文'] }] },
    originalRawChars: 5200,
    processedChars: 4000
  }, 'ai1', { gender: '女频', style: '现代悬疑' });
  assert.deepEqual(metadata, {
    sourceMode: 'novel_fetch', sourceContentVersion: 'ai1',
    sourceTaskId: 'task-123', processingRuleVersion: 'novel-fetch-rules-v1',
    gender: '女频', genderSource: 'bookinfo', genderReason: '平台分类明确', genderConfidence: 1,
    style: '现代悬疑', styleSource: 'ai', styleReason: '根据正文判断', styleConfidence: 0.91,
    category: '女频', genre: '现代悬疑', sensitiveStatus: 'done', sensitiveReplaced: 3,
    originalRawChars: 5200, processedChars: 4000,
    processingSummary: { originalRawChars: 5200, processedChars: 4000, sensitiveReplaced: 3, stages: [{ stage: 'sensitive', title: '敏感词替换', changed: true, chars: 4000, lines: 50 }] }
  });
  assert.doesNotMatch(JSON.stringify(metadata), /不得传递的命中正文/);
  assert.doesNotMatch(JSON.stringify(metadata), /"(?:sensitiveHits|originalRaw|processedText)"/);
});

test('only exposes AI versions available for every selected book', () => {
  const details = [{ processedText: '一', versions: { ai1: 'AI1一', ai2: 'AI2一' } }, { processedText: '二', versions: { ai1: 'AI1二' } }];
  assert.deepEqual(batchFactoryAvailableVersions(details), ['sensitive', 'original', 'ai1']);
  assert.equal(batchFactoryVersionLabel('ai1'), 'AI1');
});
