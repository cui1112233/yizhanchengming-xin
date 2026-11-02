import test from 'node:test';
import assert from 'node:assert/strict';
import * as admin from './admin.js';

test('updates a draft by its existing preset version', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem: () => 'admin-token', setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ preset: { id: 'batch-video-sd', version: 3, revision: 2 } }), {
      status: 200, headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await admin.updatePresetDraft('batch-video-sd', 3, { body: 'V3', expectedRevision: 1 });

  assert.equal(calls[0].path, '/api/admin/presets/batch-video-sd/3/draft');
  assert.equal(calls[0].options.method, 'PUT');
  assert.deepEqual(JSON.parse(calls[0].options.body), { body: 'V3', expectedRevision: 1 });
});

test('deletes a preset version through the admin delete endpoint', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem: () => 'admin-token', setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ preset: { id: 'batch-video-sd', version: 3, status: 'archived' } }), {
      status: 200, headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await admin.deletePreset('batch-video-sd', 3);

  assert.equal(calls[0].path, '/api/admin/presets/batch-video-sd/3');
  assert.equal(calls[0].options.method, 'DELETE');
});
