const test = require('node:test');
const assert = require('node:assert/strict');
const { JobRunner } = require('../src/job-runner');
const { AccountPool } = require('../src/account-pool');

function claim() {
  return { job: { id: 'job-hold', payload: {}, state: 'leased' }, leaseToken: 'lease', leaseGeneration: 1 };
}

function fakeApi() {
  const calls = [];
  return {
    calls,
    async progress(_token, _id, _lease, state) { calls.push(['progress', state]); },
    async release(_token, _id, _lease, reason) { calls.push(['release', reason]); },
    async fail(_token, _id, _lease, input) { calls.push(['fail', input]); },
    async renew() {},
    async acceptance() {},
    async uploadArtifact() { return { artifactId: 'artifact' }; },
    async result() {}
  };
}

for (const accountState of ['human_verification', 'quota_exhausted', 'auth_required']) {
  test(`adapter ${accountState} error keeps account unavailable after job cleanup`, async () => {
    const api = fakeApi();
    const pool = new AccountPool([{ id: 'a1', state: 'available' }]);
    const error = Object.assign(new Error(`Doubao ${accountState}`), { accountState });
    let disposed = 0;
    const adapter = {
      async prepare() { throw error; },
      dispose(id) { assert.equal(id, 'job-hold'); disposed++; }
    };
    const runner = new JobRunner({ api, token: 'device', accountPool: pool, adapter, leaseRenewIntervalMs: 0, retryDelayMs: 0 });
    await assert.rejects(() => runner.runClaim(claim()), error);
    assert.equal(pool.list()[0].state, accountState);
    assert.equal(pool.list()[0].jobId, null);
    assert.equal(disposed, 1);
  });
}

test('ordinary pre-acceptance error releases account back to available and disposes adapter state', async () => {
  const api = fakeApi();
  const pool = new AccountPool([{ id: 'a1', state: 'available' }]);
  let disposed = 0;
  const adapter = {
    async prepare() { throw new Error('temporary page problem'); },
    dispose() { disposed++; }
  };
  const runner = new JobRunner({ api, token: 'device', accountPool: pool, adapter, leaseRenewIntervalMs: 0, retryDelayMs: 0 });
  await assert.rejects(() => runner.runClaim(claim()), /temporary page problem/);
  assert.equal(pool.list()[0].state, 'available');
  assert.equal(disposed, 1);
});
