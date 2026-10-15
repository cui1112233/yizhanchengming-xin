const test = require('node:test');
const assert = require('node:assert/strict');
const { DesktopController } = require('../src/desktop-controller');

function makeRuntime() {
  const accounts = [];
  const state = () => ({ pairing: { paired: false }, accounts: accounts.map(x => ({ ...x })), automation: { enabled: false, ready: false } });
  return {
    accounts,
    async pair(input) { this.pairInput = input; return state(); },
    addAccount(account) { accounts.push({ ...account, state: 'auth_required' }); return accounts.at(-1); },
    getAccount(id) { return accounts.find(x => x.id === id) || null; },
    markAccountAvailable(id) { const item = accounts.find(x => x.id === id); item.state = 'available'; return item; },
    async heartbeat() { this.heartbeatCount = (this.heartbeatCount || 0) + 1; },
    recordError(error) { this.error = error.message; },
    getState: state
  };
}

test('adding an account persists metadata and opens its isolated browser', () => {
  const runtime = makeRuntime();
  const saved = [];
  const opened = [];
  const controller = new DesktopController({
    runtime,
    accountStore: { save(accounts) { saved.push(accounts); } },
    accountWindows: { open(id) { opened.push(id); } },
    makeAccountId: () => 'acc-1'
  });
  const account = controller.addAccount();
  assert.equal(account.id, 'acc-1');
  assert.deepEqual(opened, ['acc-1']);
  assert.equal(saved.length, 1);
  assert.equal(saved[0][0].state, 'auth_required');
});

test('marking login complete persists available state and broadcasts', () => {
  const runtime = makeRuntime();
  runtime.addAccount({ id: 'a1', name: '豆包账号 1' });
  const broadcasts = [];
  const controller = new DesktopController({
    runtime,
    accountStore: { save() {} },
    accountWindows: { open() {} },
    onState: state => broadcasts.push(state)
  });
  controller.markAccountAvailable('a1');
  assert.equal(runtime.accounts[0].state, 'available');
  assert.equal(broadcasts.length, 1);
});

test('unknown account ids cannot open arbitrary isolated browser partitions', () => {
  const runtime = makeRuntime();
  const opened = [];
  const controller = new DesktopController({
    runtime,
    accountStore: { save() {} },
    accountWindows: { open(id) { opened.push(id); } }
  });
  assert.throws(() => controller.openAccount('not-registered'), /account not found/i);
  assert.deepEqual(opened, []);
});

test('heartbeat errors are recorded but do not crash the controller', async () => {
  const runtime = makeRuntime();
  runtime.heartbeat = async () => { throw new Error('offline'); };
  const controller = new DesktopController({ runtime, accountStore: { save() {} }, accountWindows: { open() {} } });
  const ok = await controller.heartbeatOnce();
  assert.equal(ok, false);
  assert.equal(runtime.error, 'offline');
});
