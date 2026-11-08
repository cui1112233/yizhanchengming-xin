import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const pagePath = path.join(here, 'ShuihuoProductionPage.jsx');

test('shared Shuihuo library loads self-contained batch summaries instead of polling every historical project', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /Promise\.allSettled\(\[\s*listProjects\(\{ silent: true \}\),\s*listBatchSummaries\(\)/s);
  assert.match(page, /projectLibraryRefreshResult\(\{[\s\S]*batchProjectsFrom: batchFactoryProjectsFrom/);
  assert.doesNotMatch(page, /getProductionStatus\(project\.batchId/);
  assert.doesNotMatch(page, /getMergeStatus\(project\.batchId/);
});

test('shared project library renders directly without background cover fan-out', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /projectsRef\.current = refreshResult\.projects;\s*setProjects\(refreshResult\.projects\);\s*setProjectsLoadError\(''\);\s*setLoading\(false\)/s);
  assert.doesNotMatch(page, /enrichBatchProjectCovers/);
  assert.doesNotMatch(page, /projectCoverCacheRef/);
  assert.match(page, /if \(requestId === refreshRequestRef\.current\) \{\s*console\.error\('\[共享作品库\] 批量工厂读取失败'/s);
});

test('shared project library preserves displayed projects when either source is unavailable', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /previousProjects: projectsRef\.current/);
  assert.match(page, /if \(refreshResult\.error\) \{\s*setProjectsLoadError\(refreshResult\.error\);\s*setLoading\(false\);\s*return;/s);
  assert.match(page, /loadError=\{projectsLoadError\}/);
});

test('manual batch creation in the public Shuihuo page starts detached publish-metadata classification', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /classifyFetchedBatchMetadata/);
  assert.match(page, /await classifyFetchedBatchMetadata\(batch\.id\)/);
});

test('Novel Fetch handoff never offers a same-batch duplicate override', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.doesNotMatch(page, /允许加入/);
  assert.doesNotMatch(page, /allowDuplicate/);
  assert.match(page, /同书城同 Book ID 的 \$\{duplicates\.length\} 本内容已跳过/);
});

test('batchId query opens that exact batch and cannot be overwritten by an older request', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /new URLSearchParams\(window\.location\.search\)\.get\('batchId'\)/);
  assert.match(page, /getBatch\(requestedBatchId, \{ suppressGlobalError: true \}\)/);
  assert.match(page, /requestId !== projectRequestRef\.current/);
  assert.match(page, /setActiveBatchProject\(batch\);\s*setView\('batch-novels'\)/s);
});

test('giant-material creation forwards the saved immediate-execution plan under the backend field name', () => {
  const page = fs.readFileSync(pagePath, 'utf8');

  assert.match(page, /createBatchFromIntake\(input\.intakeId, \{ title: input\.title, initialBatchSettings: input\.initialBatchSettings, giantAutomationPlan: input\.giantAutomationPlan \}\)/);
});
