import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryGiantMaterialPendingProgress.jsx'), 'utf8');

test('direct-first preserves a failed book-city read instead of creating an OCR job', () => {
  assert.match(source, /OCR 任务未派发/);
  assert.match(source, /原文获取/);
  assert.match(source, /enqueueDirectFirstRead/);
  assert.match(source, /originalReadStage: 'failed'/);
  assert.match(source, /originalReadStage: 'failed', originalReadError: failure, contentPending: false/);
  assert.doesNotMatch(source, /原始书城读取失败，正在自动改用滚屏 OCR/);
  assert.doesNotMatch(source, /directReadFailed \? startOcrFallback\(\) : refetchOriginal\(true\)/);
});

test('keeps direct original fetch available while OCR executor is offline', () => {
  assert.match(source, /OCR 执行器当前未在线；可直接重试书城获取正文/);
  assert.match(source, /onClick=\{\(\) => refetchOriginal\(false\)\}>获取正文/);
  assert.match(source, /listGiantMaterialExecutors\(\)/);
  assert.match(source, /await fetchBookOriginal\(batchId, latestBook\.id\)/);
});

test('automatically retries book-city reads for previously failed direct-first books', () => {
  const autoRead = source.match(/useEffect\(\(\) => \{[\s\S]*?enqueueDirectFirstRead\(\(\) => refetchOriginal\(true\)\);[\s\S]*?\n  \}, \[[^\]]+\]\);/)?.[0] || '';
  assert.match(autoRead, /originalReadStrategy === 'direct_first'/);
  assert.doesNotMatch(autoRead, /!persistedOriginalReadError/);
  assert.match(autoRead, /autoStartRef\.current === key/);
});

test('starts a saved giant automation plan after a legacy direct read succeeds', () => {
  const directRead = source.match(/async function refetchOriginal\([\s\S]*?\n  async function startOcrFallback/)?.[0] || '';
  assert.match(directRead, /const plan = latestBook\.sourceMetadata\?\.giantAutomationPlan/);
  assert.match(directRead, /await startSavedGiantAutomation\(batchId, plan\)/);
  assert.match(source, /getBatchAutomationStatus/);
  assert.match(source, /giantAutomationStartClaims/);
});
