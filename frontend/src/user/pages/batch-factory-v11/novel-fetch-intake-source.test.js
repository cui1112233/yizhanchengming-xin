import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = relative => fs.readFileSync(path.join(here, relative), 'utf8');
const readRepo = relative => fs.readFileSync(path.resolve(here, '../../../../../', relative), 'utf8');
const readRepoPublic = relative => fs.readFileSync(path.resolve(here, '../../../../../', relative.replace(/^public\//, 'frontend/public/')), 'utf8');

test('Novel Fetch transfer targets the unified V12 intake and preserves source fields', () => {
  const source = read('../../../../public/batch-rewrite/app.js');
  assert.match(source, /platformApi\(["']\/api\/batch-factory\/v12\/intakes\/novel-fetch/);
  assert.match(source, /books:\s*items\.map/);
  assert.match(source, /sourceTaskId/);
  assert.match(source, /txtText/);
  assert.match(source, /sourceContentVersion/);
  assert.match(source, /ai_texts/);
  assert.match(source, /selectedProcessVersions\(\)/);
  assert.match(source, /platform_id\s*\|\|\s*meta\.platform_name/);
  assert.match(source, /contentCaptureCharacters/);
  assert.doesNotMatch(source, /platformApi\(["']\/api\/batch-factory\/intakes\/novel-fetch/);
  assert.match(source, /redirectTo.*batch-factory\?intake/);
});

test('novel fetch persists every processing choice immediately when changed', () => {
  const source = read('../../../../public/batch-rewrite/app.js');
  assert.match(source, /platform_id:\s*\$\("platformSelect"\)\?\.value/);
  assert.match(source, /selected_versions:\s*selectedProcessVersions\(\)/);
  assert.match(source, /ai_slot_methods:\s*processAiMethods\(\)/);
  assert.match(source, /sensitive_ai_enabled:\s*sensitiveAiProcessEnabled\(\)/);
  const toggleStart = source.indexOf('const sensitiveAiToggle = $("sensitiveAiProcessEnabled");');
  const toggleEnd = source.indexOf('const persistVersionSelection', toggleStart);
  assert.ok(toggleStart >= 0 && toggleEnd > toggleStart, 'sensitive-word toggle persistence handler must remain present');
  assert.match(source.slice(toggleStart, toggleEnd), /persistWorkFormChoiceNow/);
});

test('novel fetch keeps the 121 environment probe in the single startup hydrator', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  const hotfix = read('../../../../public/batch-rewrite/121-login-hotfix.js');
  assert.match(app, /__qiantieWebLoginEnvironmentPromise/);
  assert.match(app, /qiantieEnsureWebLoginEnvironment/);
  assert.doesNotMatch(app.slice(app.indexOf('await loadConfig()')), /api\("\/api\/web-submit\/environment"\)/);
  assert.doesNotMatch(hotfix, /window\.qiantieEnsureWebLoginEnvironment/);
  assert.doesNotMatch(hotfix, /restoreLatestProcessJob/);
});

test('novel fetch gates the initial render until saved choices and login status hydrate', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  const styles = read('../../../../public/batch-rewrite/styles.css');
  assert.match(app, /document\.documentElement\.classList\.add\("qiantie-novel-fetch-hydrating"\)/);
  assert.match(app, /finally \{\s*document\.documentElement\.classList\.remove\("qiantie-novel-fetch-hydrating"\)/);
  assert.match(styles, /html\.qiantie-novel-fetch-hydrating body > \*\s*\{\s*visibility: hidden/);
  assert.match(styles, /正在读取当前配置/);
});

test('novel fetch startup tolerates controls omitted by the enhanced layout', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  const startup = app.slice(app.indexOf('window.addEventListener("DOMContentLoaded"'));

  assert.match(startup, /function bindOptionalControl\(id, property, handler\)/);
  assert.doesNotMatch(startup, /\$\("[^"]+"\)\.(?:onclick|onchange|oninput)\s*=/,
    'a layout-specific missing control must not abort the whole startup hydrator');
});

test('novel fetch reuses a short-lived 121 session validation cache', () => {
  const service = readRepo('lib/novel-fetch-workshop/121-web-submit-service.js');
  const store = readRepo('lib/novel-fetch-store.js');
  assert.match(service, /SESSION_VALIDATION_TTL_MS/);
  assert.match(service, /validatedAt/);
  assert.match(service, /最近已验证/);
  assert.match(store, /validatedAt/);
});

test('novel fetch starts login validation in the background', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  assert.match(app, /void window\.qiantieEnsureWebLoginEnvironment\(\)\.catch/);
});

test('novel fetch does not block first render on slow 121 environment validation', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  assert.match(app, /Promise\.allSettled\(\[loadConfig\(\), loadTasks\(\)\]\)/,
    'initial render should wait only for local config and task hydration');
  assert.match(app, /void window\.qiantieEnsureWebLoginEnvironment\(\)\.catch/,
    '121 environment validation should continue in the background');
});

test('novel fetch has one startup hydrator and does not replay finished work', () => {
  const app = read('../../../../public/batch-rewrite/app.js');
  const hotfix = read('../../../../public/batch-rewrite/121-login-hotfix.js');
  assert.doesNotMatch(hotfix, /void \(async \(\) => \{/,
    'the login hotfix must not start a second page bootstrap');
  const restoreStart = app.indexOf('async function restoreLatestProcessJob()');
  const restoreEnd = app.indexOf('\nasync function pollProcessJob', restoreStart);
  assert.ok(restoreStart >= 0 && restoreEnd > restoreStart, 'latest-job restore helper must remain present');
  const restoreSource = app.slice(restoreStart, restoreEnd);
  assert.match(restoreSource, /job\.status !== ["']running["']/,
    'only an active job may be restored into the processing status area');
});

test('novel fetch status messages are bridged to the centered parent header', () => {
  const layout = read('../../../../src/shared/layouts/UserLayout.jsx');
  const feedback = read('../../../../public/batch-rewrite/interaction-feedback.js');
  assert.match(layout, /qiantie:novel-fetch-status/);
  assert.match(layout, /event\.origin !== window\.location\.origin/);
  assert.match(layout, /legacy-global-status/);
  assert.match(feedback, /qiantie:novel-fetch-status/);
  assert.match(feedback, /postMessage/);
  const middleware = readRepo('lib/novel-fetch-workshop/v2-page.js');
  assert.match(middleware, /INTERACTION_FEEDBACK_PATH/);
  assert.match(middleware, /qiantie-novel-fetch-interaction-feedback/);
});

test('novel fetch clears finished header status after a short display window', () => {
  const layout = read('../../../../src/shared/layouts/UserLayout.jsx');
  assert.match(layout, /expiringHeaderStatus\(statusText, tone\)/);
  assert.match(layout, /current\.expiresAt === expiresAt/);
});

test('novel fetch scheduling uses Chinese progress states and supports deleting records', () => {
  const controls = readRepoPublic('public/batch-rewrite/v78-novel-fetch-v2-run-controls.js');
  assert.match(controls, /button\.textContent = '开始定时'/);
  assert.match(controls, /正在定时中/);
  assert.match(controls, /定时完成/);
  assert.match(controls, /method: 'DELETE'/);
  assert.match(controls, /删除记录/);
});

test('novel fetch table translates backend task states to Chinese', () => {
  const layout = readRepoPublic('public/batch-rewrite/v78-novel-fetch-v2-layout.js');
  assert.match(layout, /original_done/);
  assert.match(layout, /已完成/);
  assert.match(layout, /statusLabel/);
});
