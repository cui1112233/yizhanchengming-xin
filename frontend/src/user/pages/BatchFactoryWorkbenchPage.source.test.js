import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const pagePath = path.join(here, 'BatchFactoryWorkbenchPage.jsx');

test('uses only Batch Factory data and keeps novel-fetch intake handoff', () => {
  const page = fs.readFileSync(pagePath, 'utf8');
  assert.match(page, /listBatchSummaries\(\)/);
  assert.match(page, /BatchFactoryCreateModal/);
  assert.match(page, /BatchFactoryNovelList/);
  assert.match(page, /创作漫剧/);
  assert.match(page, /pendingNovelFetchIntakeId\(window\.location\.search\)/);
  assert.doesNotMatch(page, /listProjects\(/);
  assert.doesNotMatch(page, /getProductionHealth\(/);
  assert.doesNotMatch(page, /CommentaryWorkbench/);
});

test('creates a batch from an intake when the giant-material flow has already registered its placeholder book', () => {
  const page = fs.readFileSync(pagePath, 'utf8');
  assert.match(page, /input\?\.intakeId\s*\?\s*await createBatchFromIntake\(input\.intakeId,\s*\{\s*title: input\.title,\s*giantAutomation: input\.giantAutomation\s*\}\)/s);
});

test('novel-fetch handoff always opens a new batch instead of appending to the active batch', () => {
  const page = fs.readFileSync(pagePath, 'utf8');
  const handoff = page.slice(page.indexOf('const intakeId = pendingNovelFetchIntakeId'), page.indexOf('async function createBatch('));
  assert.match(handoff, /createBatchFromIntake\(intakeId/);
  assert.doesNotMatch(handoff, /appendNovelFetchIntake/);
  assert.doesNotMatch(handoff, /currentBatchId/);
  assert.match(handoff, /\/batch-factory/);
});

test('returns the created batch so giant-material OCR can bind its executor job', () => {
  const page = fs.readFileSync(pagePath, 'utf8');
  assert.match(page, /if\s*\(!input\?\.automationEnabled\)\s*return batch;/);
  assert.match(page, /onCreated=\{async input => \{\s*const batch = await createBatch\(input\);\s*setCreateOpen\(false\);\s*return batch;\s*\}\}/s);
  assert.match(page, /onBatchUpdated=\{async batchId => \{\s*const refreshed = asBatch\(await getBatch\(batchId\)\);/s);
});
