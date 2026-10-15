const test = require('node:test');
const assert = require('node:assert/strict');
const { DesktopRuntime } = require('../src/desktop-runtime');

function storeWithPairing() {
  return {
    load: () => ({ baseUrl: 'https://v78.example', executorId: 'lex-1', token: 'device-token' }),
    save: value => value
  };
}

function liveAdapter(calls) {
  return {
    async prepare() { calls.push('prepare'); },
    async submit() { calls.push('submit'); return { status: 'accepted', submissionId: 'msg-1' }; },
    async waitForCompletion() { calls.push('wait'); return { mediaId: 'media-1' }; },
    async fetchArtifact() { calls.push('download'); return { filePath: '/tmp/video.mp4' }; },
    dispose() { calls.push('dispose'); }
  };
}

function api(calls, withJob = true) {
  let claimed = false;
  return {
    async heartbeat() {},
    async claim() {
      calls.push('claim');
      if (!withJob || claimed) return null;
      claimed = true;
      return { job: { id: 'job-1', payload: {}, state: 'leased' }, leaseToken: 'lease', leaseGeneration: 1 };
    },
    async progress(_token, _id, _lease, state) { calls.push(`progress:${state}`); },
    async acceptance() { calls.push('acceptance'); },
    async release() { calls.push('release'); },
    async fail() { calls.push('fail'); },
    async renew() {},
    async uploadArtifact() { calls.push('upload'); return { artifactId: 'artifact-1' }; },
    async result() { calls.push('result'); }
  };
}

test('paired runtime with live adapter and available account can enable and run one claimed job', async () => {
  const calls = [];
  const runtime = new DesktopRuntime({
    deviceStore: storeWithPairing(),
    apiFactory: () => api(calls),
    deviceName: 'PC', platform: 'win32', version: '1.0.0',
    accounts: [{ id: 'a1', state: 'available' }],
    adapter: liveAdapter(calls),
    runnerOptions: { leaseRenewIntervalMs: 0, retryDelayMs: 0 }
  });
  runtime.setAutomationEnabled(true);
  assert.equal(runtime.getState().automation.enabled, true);
  assert.equal(runtime.getState().automation.ready, true);
  const result = await runtime.claimOnce();
  assert.deepEqual(result, { succeeded: true, artifactId: 'artifact-1' });
  assert.equal(calls.filter(x => x === 'claim').length, 1);
  assert.equal(calls.filter(x => x === 'submit').length, 1);
  assert.equal(calls.filter(x => x === 'result').length, 1);
  assert.equal(runtime.getState().currentTask, null);
});

test('automation does not claim a cloud job when no local account is available', async () => {
  const calls = [];
  const runtime = new DesktopRuntime({
    deviceStore: storeWithPairing(), apiFactory: () => api(calls),
    deviceName: 'PC', platform: 'win32', version: '1.0.0',
    accounts: [{ id: 'a1', state: 'human_verification' }],
    adapter: liveAdapter(calls)
  });
  assert.throws(() => runtime.setAutomationEnabled(true), /available Doubao account/i);
  assert.equal(calls.filter(x => x === 'claim').length, 0);
});
