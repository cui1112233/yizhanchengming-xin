const test = require('node:test');
const assert = require('node:assert/strict');
const { AccountPool } = require('../src/account-pool');

test('an available account can only be acquired once', () => {
  const pool = new AccountPool([{ id: 'a1', state: 'available' }]);
  assert.equal(pool.acquire().id, 'a1');
  assert.equal(pool.acquire(), null);
});

test('quota auth and verification accounts are excluded', () => {
  const pool = new AccountPool([
    { id: 'a1', state: 'quota_exhausted' },
    { id: 'a2', state: 'auth_required' },
    { id: 'a3', state: 'human_verification' },
    { id: 'a4', state: 'available' }
  ]);
  assert.equal(pool.acquire().id, 'a4');
});

test('release makes account available again', () => {
  const pool = new AccountPool([{ id: 'a1', state: 'available' }]);
  pool.acquire();
  pool.release('a1');
  assert.equal(pool.acquire().id, 'a1');
});
