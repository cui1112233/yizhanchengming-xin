const assert = require('node:assert/strict');
const test = require('node:test');

const { attachV78NovelFetchV2, resolveBrowserClientFactory } = require('./v2-compose');

test('does not use the browser worker unless explicitly opted in', () => {
  const workerFactory = () => ({ kind: 'worker' });
  const directFactory = () => ({ kind: 'direct' });

  const selected = resolveBrowserClientFactory({
    QIANTIE_121_BROWSER_WORKER_URL: 'http://browser-worker:8787',
    QIANTIE_121_WORKER_SECRET: 'internal-secret'
  }, { workerFactory, directFactory });

  assert.equal(selected, directFactory);
  assert.equal(selected().kind, 'direct');
});

test('allows the browser worker only with an explicit emergency opt-in', () => {
  const workerFactory = () => ({ kind: 'worker' });
  const directFactory = () => ({ kind: 'direct' });

  const selected = resolveBrowserClientFactory({
    QIANTIE_121_USE_BROWSER_WORKER: '1',
    QIANTIE_121_BROWSER_WORKER_URL: 'http://browser-worker:8787',
    QIANTIE_121_WORKER_SECRET: 'internal-secret'
  }, { workerFactory, directFactory });

  assert.equal(selected, workerFactory);
  assert.equal(selected().kind, 'worker');
});

test('wires the direct 121 client into the web-submit service', () => {
  let receivedDirectClient = null;
  let schedulerStarted = false;
  const directClient = { login() {}, verify() {}, action() {} };
  const shellApp = { use() {}, locals: {} };
  const coreApp = {
    locals: {
      authRuntime: { accountStore: { files: { audit: '/tmp/system/audit.json' }, getAccount() { return null; } } },
      novelFetchStore: { getBrowserSession() {}, setBrowserSession() {} }
    }
  };

  attachV78NovelFetchV2({
    shellApp,
    coreApp,
    bodyParser: {},
    createBrowserClient: () => directClient,
    createCredentialStore: () => ({}),
    createWebSubmit: options => {
      receivedDirectClient = options.browserClient;
      return { ensureSession: async () => {}, syncStyles: async () => {}, submit: async () => {} };
    },
    createTargetWebSubmit: options => options.service,
    createBatchExecutor: () => async () => ({}),
    createRuntime: () => ({ queue: {}, scheduler: {}, startScheduler() { schedulerStarted = true; } }),
    createRouter: () => ({}),
    createTombstones: () => ({}),
    createTaskOps: () => ({ processConflicts() { return []; } }),
    createBatches: () => ({ complete() {}, updateTaskState() {} })
  });

  assert.equal(receivedDirectClient, directClient);
  assert.equal(schedulerStarted, true);
  assert.equal(coreApp.locals.novelFetchV2WebSubmit, shellApp.locals.novelFetchV2WebSubmit);
});
