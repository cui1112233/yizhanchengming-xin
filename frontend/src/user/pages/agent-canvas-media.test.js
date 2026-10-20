import test from 'node:test';
import assert from 'node:assert/strict';

import { getAgentAssetBlob } from '../../shared/api/agent.js';
import { createAuthenticatedMediaSession, createAuthenticatedMediaStore } from './agent-canvas-media.js';

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

test('Agent asset video is fetched through the authenticated API client', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = {
    getItem(key) { return key === 'auth_token' ? 'agent-media-token' : ''; },
    setItem() {},
    removeItem() {}
  };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, headers: new Headers(options.headers || {}) });
    return new Response(new Blob(['video'], { type: 'video/mp4' }), {
      status: 200,
      headers: { 'content-type': 'video/mp4' }
    });
  };
  t.after(() => {
    globalThis.localStorage = originalLocalStorage;
    globalThis.fetch = originalFetch;
  });

  const path = '/api/agent/tasks/project-1/assets/asset-1/content';
  const blob = await getAgentAssetBlob(path);

  assert.equal(blob.type, 'video/mp4');
  assert.equal(calls.length, 1);
  assert.equal(calls[0].path, path);
  assert.equal(calls[0].headers.get('authorization'), 'Bearer agent-media-token');
});

test('Agent asset blob helper refuses external URLs before sending authentication', async () => {
  await assert.rejects(() => getAgentAssetBlob('https://media.example/private.mp4'), /Agent asset/i);
});

test('authenticated media store shares one download and revokes after the last release', async () => {
  const loads = [];
  const revoked = [];
  const store = createAuthenticatedMediaStore({
    loadBlob: async source => { loads.push(source); return new Blob(['video']); },
    createObjectUrl: () => 'blob:shared-video',
    revokeObjectUrl: url => revoked.push(url)
  });

  const first = store.acquire('/private/video');
  const second = store.acquire('/private/video');
  assert.equal(await first.promise, 'blob:shared-video');
  assert.equal(await second.promise, 'blob:shared-video');
  assert.deepEqual(loads, ['/private/video']);

  first.release();
  assert.deepEqual(revoked, []);
  second.release();
  assert.deepEqual(revoked, ['blob:shared-video']);
});

test('authenticated media store revokes a late object URL after every consumer unmounts', async () => {
  const response = deferred();
  const revoked = [];
  const store = createAuthenticatedMediaStore({
    loadBlob: () => response.promise,
    createObjectUrl: () => 'blob:late-video',
    revokeObjectUrl: url => revoked.push(url)
  });

  const lease = store.acquire('/private/late-video');
  lease.release();
  response.resolve(new Blob(['video']));

  assert.equal(await lease.promise, 'blob:late-video');
  assert.deepEqual(revoked, ['blob:late-video']);
});

test('authenticated media store releases the old URL when the media source changes', async () => {
  const revoked = [];
  const store = createAuthenticatedMediaStore({
    loadBlob: async source => new Blob([source]),
    createObjectUrl: blob => `blob:${blob.size}`,
    revokeObjectUrl: url => revoked.push(url)
  });

  const first = store.acquire('/private/a');
  await first.promise;
  first.release();
  const second = store.acquire('/private/longer-source');
  await second.promise;
  second.release();

  assert.equal(revoked.length, 2);
});

test('authenticated media session reports loading, success, and releases on unmount', async () => {
  const response = deferred();
  const states = [];
  let releases = 0;
  const session = createAuthenticatedMediaSession('/private/video', {
    store: { acquire: () => ({ promise: response.promise, release: () => { releases += 1; } }) },
    onState: state => states.push(state)
  });

  assert.deepEqual(states, [{ url: '', loading: true, error: '' }]);
  response.resolve('blob:video');
  await response.promise;
  await Promise.resolve();
  assert.deepEqual(states.at(-1), { url: 'blob:video', loading: false, error: '' });

  session.dispose();
  assert.equal(releases, 1);
  session.dispose();
  assert.equal(releases, 1);
});

test('authenticated media session reports errors and ignores late results after unmount', async () => {
  const failure = deferred();
  const failureStates = [];
  createAuthenticatedMediaSession('/private/failure', {
    store: { acquire: () => ({ promise: failure.promise, release() {} }) },
    onState: state => failureStates.push(state)
  });
  failure.reject(new Error('登录已失效'));
  await assert.rejects(failure.promise, /登录已失效/);
  await Promise.resolve();
  assert.deepEqual(failureStates.at(-1), { url: '', loading: false, error: '登录已失效' });

  const late = deferred();
  const lateStates = [];
  let releases = 0;
  const session = createAuthenticatedMediaSession('/private/late', {
    store: { acquire: () => ({ promise: late.promise, release: () => { releases += 1; } }) },
    onState: state => lateStates.push(state)
  });
  session.dispose();
  late.resolve('blob:late');
  await late.promise;
  await Promise.resolve();
  assert.equal(releases, 1);
  assert.deepEqual(lateStates, [{ url: '', loading: true, error: '' }]);
});
