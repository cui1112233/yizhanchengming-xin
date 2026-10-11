const path = require('node:path');

function resolveBrowserClientFactory(env = process.env, {
  workerFactory = require('./121-browser-worker-client').create121BrowserWorkerClient,
  directFactory = require('./121-browser-client').create121BrowserClient
} = {}) {
  const workerConfigured = String(env.QIANTIE_121_BROWSER_WORKER_URL || '').trim()
    && String(env.QIANTIE_121_WORKER_SECRET || '').trim();
  const workerDisabled = String(env.QIANTIE_121_USE_BROWSER_WORKER || '').trim() === '0';
  return workerConfigured && !workerDisabled ? workerFactory : directFactory;
}

function attachV78NovelFetchV2({
  shellApp,
  coreApp,
  bodyParser,
  createBatchExecutor,
  createRuntime,
  createRouter,
  createTombstones,
  createTaskOps,
  createBatches,
  createBrowserClient,
  createCredentialStore,
  createWebSubmit,
  createTargetWebSubmit
} = {}) {
  if (!shellApp || typeof shellApp.use !== 'function') throw new Error('shellApp is required');
  const authRuntime = coreApp?.locals?.authRuntime;
  if (!authRuntime?.accountStore) throw new Error('core auth runtime is required');
  // 兼容壳层/旧启动方式没有把模型解析器挂到 locals 的情况。
  // 敏感词重处理和批量改文必须复用同一个中央文本模型目录，不能因此退化成“服务不可用”。
  const runtimeModelResolver = coreApp.locals?.resolveRuntimeModel || ((username, kind, modelId) => {
    const { resolveRuntimeModel } = require('../model-catalog-runtime');
    const { readConfig } = require('../shared');
    return resolveRuntimeModel({
      username,
      kind,
      modelId,
      memberStore: coreApp.locals?.memberStore,
      accountStore: authRuntime.accountStore,
      configReader: readConfig
    });
  });
  const auditPath = authRuntime.accountStore.files?.audit;
  if (!auditPath) throw new Error('core audit path is required');

  const batchExecutorFactory = createBatchExecutor || require('./v2-batch-executor').createV78NovelFetchBatchExecutor;
  const runtimeFactory = createRuntime || require('./v2-runtime').createNovelFetchV2Runtime;
  const routerFactory = createRouter || require('../../routes/batch-rewrite-v2').createBatchRewriteV2Router;
  const tombstoneFactory = createTombstones || require('./tombstones').createNovelFetchTombstones;
  const taskOpsFactory = createTaskOps || require('./task-ops').createNovelFetchTaskOps;
  const batchesFactory = createBatches || require('./batches').createNovelFetchBatches;
  const browserClientFactory = createBrowserClient || resolveBrowserClientFactory();
  const credentialStoreFactory = createCredentialStore || require('./121-credential-store').create121CredentialStore;
  const webSubmitFactory = createWebSubmit || require('./121-web-submit-service').create121WebSubmitService;
  const targetWebSubmitFactory = createTargetWebSubmit || require('./target-web-submit').createTargetAwareWebSubmit;
  const parser = bodyParser || require('express').json({ limit: '50mb' });
  const errorLogStore = coreApp.locals?.errorLogStore;

  const accountResolver = username => {
    try { return authRuntime.accountStore.getAccount(username) || null; }
    catch (_) { return null; }
  };
  const usersDir = path.join(path.dirname(auditPath), '..', 'users');
  const batches = batchesFactory({ usersDir });
  const browserClient = browserClientFactory();
  const credentialStore = credentialStoreFactory({
    usersDir,
    secret: process.env.QIANTIE_121_CREDENTIAL_SECRET || process.env.QIANTIE_BRIDGE_SECRET || ''
  });
  const sessionStore = coreApp.locals?.novelFetchStore;
  if (!sessionStore?.getBrowserSession || !sessionStore?.setBrowserSession) throw new Error('core novel fetch browser session store is required');
  const createStore = ({ account }) => require('./mysql-store').createMySQLWorkshopStore({ account });
  const baseWebSubmit = webSubmitFactory({ accountResolver, createStore, browserClient, sessionStore, credentialStore });
  // Explicit test/integration injections preserve their exact object contract. Production V78 adds only
  // the target-aware wrapper, so the existing “提交网络” button remains the single execution entry.
  const webSubmit = createWebSubmit
    ? baseWebSubmit
    : targetWebSubmitFactory({
        service: baseWebSubmit,
        accountResolver,
        createStore,
        onTaskUpdated(owner, meta) {
          const batchId = String(meta?.batchId || meta?.batch_id || '').trim();
          if (batchId && typeof batches.updateTaskState === 'function') batches.updateTaskState(owner, batchId, meta);
        }
      });
  const executeBatch = batchExecutorFactory({
    accountResolver,
    resolveRuntimeModel: runtimeModelResolver,
    retentionConfigReader: username => require('../shared').readConfig(username),
    webSessionReady: async owner => {
      try { await webSubmit.ensureSession(owner); return true; }
      catch (_) { return false; }
    },
    syncSiteStyles: owner => webSubmit.syncStyles(owner),
    submit: (owner, request) => webSubmit.submit(owner, request)
  });
  const tombstones = tombstoneFactory({ usersDir });
  const taskOpsOptions = {
    accountResolver,
    tombstones,
    resolveRuntimeModel: runtimeModelResolver
  };
  if (!createTaskOps) {
    taskOpsOptions.createStore = options => require('./mysql-store').createMySQLWorkshopStore(options);
    taskOpsOptions.parseBooks = require('./parse').parseBooks;
  }
  const taskOps = taskOpsFactory(taskOpsOptions);
  const guardedExecuteBatch = async (owner, payload, executionContext = {}) => {
    const conflicts = taskOps.processConflicts(owner, payload || {});
    if (conflicts.length) {
      const error = new Error(`书籍已永久删除：${conflicts.join('、')}`);
      error.status = 410;
      error.code = 'permanently_deleted';
      error.bookIds = conflicts;
      error.recoverable = false;
      throw error;
    }
    const result = await executeBatch(owner, payload, executionContext);
    if (payload?.batch_id && typeof batches.complete === 'function') {
      batches.complete(owner, payload.batch_id, result || {});
    }
    return result;
  };
  const runtime = runtimeFactory({
    usersDir,
    executeBatch: guardedExecuteBatch,
    onSchedulerError(error) {
      if (errorLogStore && typeof errorLogStore.record === 'function') {
        errorLogStore.record({ kind: 'novel-fetch-v2.scheduler', message: error?.message || String(error), status: 500 });
      }
    }
  });

  shellApp.locals.authRuntime = authRuntime;
  if (coreApp.locals?.memberStore) shellApp.locals.memberStore = coreApp.locals.memberStore;
  shellApp.locals.novelFetchV2Runtime = runtime;
  shellApp.locals.novelFetchV2TaskOps = taskOps;
  shellApp.locals.novelFetchV2Batches = batches;
  shellApp.locals.novelFetchV2Tombstones = tombstones;
  shellApp.locals.novelFetchV2WebSubmit = webSubmit;
  // Batch Factory routes live on the mounted core app. Share the same durable
  // 121 session service so login, synchronization and upload all use one
  // account-scoped session and the encrypted-credential auto-login path.
  coreApp.locals.novelFetchV2WebSubmit = webSubmit;
  shellApp.use('/api/batch-rewrite', parser, routerFactory({ queue: runtime.queue, scheduler: runtime.scheduler, taskOps, batches, webSubmit, throughput: runtime.throughput }));
  runtime.startScheduler();
  return runtime;
}

module.exports = { attachV78NovelFetchV2, resolveBrowserClientFactory };
