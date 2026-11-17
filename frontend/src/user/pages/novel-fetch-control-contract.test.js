import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const page = fs.readFileSync(new URL('./NovelFetchPage.jsx', import.meta.url), 'utf8');
const api = fs.readFileSync(new URL('../../shared/api/novelFetchWorkshop.js', import.meta.url), 'utf8');
const submitApi = fs.readFileSync(new URL('../../shared/api/novelFetch.js', import.meta.url), 'utf8');

test('visible workflow actions have Go-backed handlers', () => {
  for (const name of [
    'fetchWorkshopOriginal',
    'startWorkshopProcess',
    'retryWorkshopTasks',
    'stopWorkshopTasks',
    'generateWorkshopAi',
    'applyWorkshopRules',
    'previewWebSubmit',
    'startWebSubmit'
  ]) {
    assert.match(page, new RegExp(name));
    assert.match(name.endsWith('WebSubmit') ? submitApi : api, new RegExp(name));
  }
});

test('workflow actions refresh state after mutation', () => {
  assert.match(page, /loadTasks\(\)/);
  assert.match(page, /loadCurrentBatch\(\)/);
  assert.match(page, /setWebSubmitResults/);
});

test('row and knowledge actions surface failures and refresh their lists', () => {
  assert.match(page, /runTaskRowAction/);
  assert.match(page, /runKnowledgeItemAction/);
  assert.match(page, /rerunHistoryBatch/);
  assert.match(page, /retryWebSubmitVersion/);
  assert.match(page, /loading=\{tasksLoading\}/);
  assert.match(page, /loading=\{knowledgeLoading\}/);
});

test('visible action labels do not contain the retired 121 name', () => {
  assert.doesNotMatch(page, />[^<{]*121[^<{]*</);
});

test('sensitive AI controls bind to the persisted backend fields', () => {
  for (const field of ['group_hits', 'concurrency', 'max_hits_per_task', 'context_chars', 'retries', 'temperature', 'prompt']) {
    assert.match(page, new RegExp(`sensitive_ai.*${field}|${field}.*sensitive_ai`));
  }
});

test('config save hydrates the form from the server response', () => {
  assert.match(page, /const saved = await saveWorkshopConfig/);
  assert.match(page, /saved\?\.settings/);
  assert.match(page, /configForm\.setFieldsValue\(effectiveValues\)/);
});

test('AI version picker and maxTxt limits match backend capabilities', () => {
  assert.match(page, /AI_VERSION_KEYS = Array\.from\(\{ length: 5 \}/);
  assert.match(page, /max=\{200000\}/);
  assert.match(api, /JSON\.stringify\(\{ purpose, model: model \|\| '' \}\)/);
  assert.match(page, /testWorkshopAi\('rewrite', textModel\)/);
});
