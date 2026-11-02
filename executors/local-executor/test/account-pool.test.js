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

test('available accounts are selected in round-robin order', () => {
  const pool = new AccountPool([
    { id: 'a1', state: 'available' },
    { id: 'a2', state: 'available' },
    { id: 'a3', state: 'available' }
  ]);

  const selected = [];
  for (let index = 0; index < 6; index += 1) {
    const account = pool.acquire({ jobId: `job-${index}` });
    selected.push(account.id);
    pool.release(account.id);
  }

  assert.deepEqual(selected, ['a1', 'a2', 'a3', 'a1', 'a2', 'a3']);
});

test('busy accounts are never double-leased while another account is available', () => {
  const pool = new AccountPool([
    { id: 'a1', state: 'available' },
    { id: 'a2', state: 'available' }
  ]);

  const first = pool.acquire({ jobId: 'job-1' });
  const second = pool.acquire({ jobId: 'job-2' });

  assert.notEqual(first.id, second.id);
  assert.equal(pool.acquire({ jobId: 'job-3' }), null);
  assert.equal(first.jobId, 'job-1');
  assert.equal(second.jobId, 'job-2');
});
