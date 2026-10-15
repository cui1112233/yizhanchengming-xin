const test = require('node:test');
const assert = require('node:assert/strict');
const { DesktopRuntime } = require('../src/desktop-runtime');

function fakeDeviceStore(initial = null) {
  let saved = initial;
  return {
    load() { return saved; },
    save(value) { saved = { ...value }; return { baseUrl: value.baseUrl, executorId: value.executorId }; },
    snapshot() { return saved; }
  };
}

function fakeApiFactory(log) {
  return baseUrl => ({
    async pair(input) {
      log.push(['pair', baseUrl, input]);
      return { executorId: 'lex-1', token: 'device-token', heartbeatIntervalSeconds: 15 };
    },
    async heartbeat(token, input) {
      log.push(['heartbeat', token, input]);
      return { ok: true, heartbeatIntervalSeconds: 15 };
    }
  });
}

test('pair stores the returned executor token and exposes paired state without the token', async () => {
  const log = [];
  const store = fakeDeviceStore();
  const runtime = new DesktopRuntime({
    deviceStore: store,
    apiFactory: fakeApiFactory(log),
    deviceName: 'DESKTOP-1',
    platform: 'win32',
    version: '1.0.0-dev.1'
  });

  await runtime.pair({ baseUrl: 'https://example.test/', code: 'ABCD-EFGH' });
  assert.equal(store.snapshot().token, 'device-token');
  assert.deepEqual(runtime.getState().pairing, {
    paired: true,
    baseUrl: 'https://example.test',
    executorId: 'lex-1'
  });
  assert.equal(JSON.stringify(runtime.getState()).includes('device-token'), false);
});

test('restored pairing can heartbeat with non-sensitive account counters', async () => {
  const log = [];
  const runtime = new DesktopRuntime({
    deviceStore: fakeDeviceStore({ baseUrl: 'https://example.test', executorId: 'lex-1', token: 'secret' }),
    apiFactory: fakeApiFactory(log),
    deviceName: 'DESKTOP-1',
    platform: 'darwin',
    version: '1.0.0-dev.1'
  });
  runtime.addAccount({ id: 'a1', name: '账号 1' });
  runtime.markAccountAvailable('a1');
  runtime.addAccount({ id: 'a2', name: '账号 2' });

  await runtime.heartbeat();
  const call = log.find(entry => entry[0] === 'heartbeat');
  assert.equal(call[1], 'secret');
  assert.deepEqual(call[2].accounts, {
    total: 2,
    available: 1,
    busy: 0,
    quotaExhausted: 0,
    loginError: 1,
    humanVerification: 0
  });
});

test('successful heartbeat does not erase the last local job error', async () => {
  const log = [];
  const runtime = new DesktopRuntime({
    deviceStore: fakeDeviceStore({ baseUrl: 'https://example.test', executorId: 'lex-1', token: 'secret' }),
    apiFactory: fakeApiFactory(log),
    deviceName: 'DESKTOP-1',
    platform: 'win32',
    version: '1.0.1'
  });
  runtime.recordError(new Error('Doubao control is ambiguous: 视频'));
  assert.equal(runtime.getState().lastError, 'Doubao control is ambiguous: 视频');

  await runtime.heartbeat();

  assert.equal(runtime.getState().lastError, 'Doubao control is ambiguous: 视频');
});

test('new accounts require login and may be marked available after manual login', () => {
  const runtime = new DesktopRuntime({
    deviceStore: fakeDeviceStore(),
    apiFactory: fakeApiFactory([]),
    deviceName: 'DESKTOP-1',
    platform: 'win32',
    version: '1.0.0-dev.1'
  });
  runtime.addAccount({ id: 'a1', name: '豆包账号 1' });
  assert.equal(runtime.getState().accounts[0].state, 'auth_required');
  runtime.markAccountAvailable('a1');
  assert.equal(runtime.getState().accounts[0].state, 'available');
});

test('automation remains fail-closed until a live Doubao adapter is installed', () => {
  const runtime = new DesktopRuntime({
    deviceStore: fakeDeviceStore(),
    apiFactory: fakeApiFactory([]),
    deviceName: 'DESKTOP-1',
    platform: 'win32',
    version: '1.0.0-dev.1'
  });
  assert.throws(() => runtime.setAutomationEnabled(true), /live Doubao adapter/i);
  assert.equal(runtime.getState().automation.enabled, false);
});
