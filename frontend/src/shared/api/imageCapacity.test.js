import test from 'node:test';
import assert from 'node:assert/strict';
import { getImageCapacity, saveImageCapacity } from './imageCapacity.js';
import { readFileSync } from 'node:fs';

test('capacity reads saved model profile and saves administrator policy with authentication', async t => {
  const originalFetch = globalThis.fetch;
  const originalStorage = globalThis.localStorage;
  globalThis.localStorage = { getItem: () => 'session', removeItem() {} };
  const calls = [];
  globalThis.fetch = async (path, options) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ policy: { modelLimit: 20 }, running: 3 }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.fetch = originalFetch; globalThis.localStorage = originalStorage; });
  const result = await getImageCapacity('custom:a &b', { aspectRatio: '16:9' });
  assert.equal(result.running, 3);
  assert.equal(calls[0].path, '/api/config/image-capacity?modelId=custom%3Aa+%26b&aspectRatio=16%3A9');
  assert.equal(calls[0].options.headers.get('Authorization'), 'Bearer session');
  await saveImageCapacity('custom:a', { aspectRatio: '16:9', accountLimit: 20, modelLimit: 12, initialLimit: 8, groupKey: 'shared' });
  assert.equal(calls[1].options.method, 'PUT');
  assert.deepEqual(JSON.parse(calls[1].options.body), { modelId: 'custom:a', aspectRatio: '16:9', accountLimit: 20, modelLimit: 12, initialLimit: 8, groupKey: 'shared' });
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'Redis 暂不可用' }), { status: 503 });
  await assert.rejects(getImageCapacity('custom:a'), /Redis 暂不可用/);
});

test('unified settings uses shared automatic capacity instead of per-book fixed choices', () => {
  const modal = readFileSync(new URL('../../user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx', import.meta.url), 'utf8');
  assert.match(modal, /<ImageCapacityControl modelId=\{value.imageModelId\} aspectRatio=\{value.imageAspectRatio \|\| value.aspectRatio \|\| \'9:16\'\} active=\{active\}/);
  assert.doesNotMatch(modal, /onChange=\{imageConcurrency =>/);
});


test('capacity panel explains reserved uncertain outcomes and shortens account identity', () => {
  const source = readFileSync(new URL('../../user/pages/shuihuo/ImageCapacityControl.jsx', import.meta.url), 'utf8');
  assert.match(source, /state && !error &&/);
  assert.match(source, /待核对占用 \{state.uncertain \|\| 0\}/);
  assert.match(source, /请求已发出但未确认完成，暂时保留名额以防超额；不会自动再生成扣费/);
  assert.match(source, /title=\{state.policy.accountKey\}/);
  assert.match(source, /slice\(0, 16\)/);
});
