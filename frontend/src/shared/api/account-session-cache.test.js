import test from 'node:test';
import assert from 'node:assert/strict';
import { createAccountSessionCache } from './account-session-cache.js';

function setup() {
  const values = new Map([['auth_token', 'session-a'], ['auth_username', 'alice']]);
  const storage = { getItem: key => values.get(key) || null, setItem: (key, value) => values.set(key, value), removeItem: key => values.delete(key) };
  let now = 1000, calls = 0;
  const pending = [];
  const cache = createAccountSessionCache({ storage, now: () => now, request: () => { calls++; return new Promise((resolve, reject) => pending.push({ resolve, reject })); } });
  return { cache, values, storage, pending, calls: () => calls, advance: ms => { now += ms; } };
}
test('cold start reads only non-secret presentation fields for the current signed-in account', () => {
  const s = setup();
  s.cache.savePreview({ username: 'alice', themeId: 'ocean-dark', password: 'never-store', token: 'never-store', apiKeys: ['secret'], role: 'admin' });
  assert.deepEqual(s.cache.readPreview(), { username: 'alice', themeId: 'ocean-dark' });
  assert.doesNotMatch(s.values.get('auth_account_preview'), /never-store|apiKeys|admin/);
  s.values.set('auth_username', 'bob');
  assert.equal(s.cache.readPreview(), null);
});
test('parallel and shortly repeated checks share one session request', async () => {
  const s = setup();
  const first = s.cache.load(), second = s.cache.load();
  assert.equal(s.calls(), 1);
  s.pending[0].resolve({ username: 'alice', role: 'member' });
  assert.deepEqual(await first, await second);
  await s.cache.load();
  assert.equal(s.calls(), 1);
  s.advance(11000);
  const next = s.cache.load();
  assert.equal(s.calls(), 2);
  s.pending[1].resolve({ username: 'alice' });
  await next;
});
test('an old account response cannot populate a replacement session', async () => {
  const s = setup();
  const first = s.cache.load();
  s.values.set('auth_token', 'session-b');
  s.values.set('auth_username', 'bob');
  s.pending[0].resolve({ username: 'alice' });
  await assert.rejects(first, /会话已切换/);
  assert.equal(s.cache.readPreview(), null);
});
test('missing tokens, expired snapshots and explicit invalidation do not restore a login', () => {
  const s = setup();
  s.cache.savePreview({ username: 'alice' });
  s.advance(24 * 60 * 60 * 1000 + 1);
  assert.equal(s.cache.readPreview(), null);
  s.cache.savePreview({ username: 'alice' });
  s.values.delete('auth_token');
  assert.equal(s.cache.readPreview(), null);
  s.cache.invalidate();
  assert.equal(s.values.has('auth_account_preview'), false);
});
test('a transient validation failure leaves presentation cache available', async () => {
  const s = setup();
  s.cache.savePreview({ username: 'alice' });
  const check = s.cache.load();
  s.pending[0].reject(new Error('temporary network failure'));
  await assert.rejects(check);
  assert.equal(s.cache.readPreview().username, 'alice');
});
test('profile edits force a fresh request and an older response cannot overwrite it', async () => {
  const s = setup();
  const old = s.cache.load();
  const refreshed = s.cache.load({ force: true });
  assert.equal(s.calls(), 2);
  s.pending[1].resolve({ username: 'alice', avatarUrl: '/new-avatar.png' });
  await refreshed;
  s.pending[0].resolve({ username: 'alice', avatarUrl: '/old-avatar.png' });
  await assert.rejects(old, /会话已切换/);
  assert.equal(s.cache.readPreview().avatarUrl, '/new-avatar.png');
});
