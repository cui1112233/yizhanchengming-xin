const test = require('node:test');
const assert = require('node:assert/strict');
const express = require('express');
const { createSignedBridgeHeaders } = require('../lib/batch-factory-v11/go-proxy');

const {
  enrichBatchFactorySystemPresetConfig,
  resolveDerivedOpeningPrompt,
  refreshBatchFactoryPresetSnapshot,
  presetDrivenExecutionPath,
  smartUnifiedSelected,
  directorVisualBaselineRequired,
  styleSystemBookPath,
  analyzeBatchFactorySmartUnifiedStyle,
  acquireBatchFactorySmartUnifiedBaseline,
  needsPersonalConfigSync,
  needsH3ConfigSync,
  redactBatchFactorySystemPromptBodies,
  upstreamErrorMessage,
  batchFactory121PublishPath,
  batchFactory121OrganizationsPath,
  organizationOptions,
  resolveBatchFactory121Session,
  mpegAudioDurationSeconds,
  automationPublishSettings,
  submitBatchFactoryBookTo121,
  classifyBatchFactoryBookFor121,
  resolveBatchFactoryBookClassificationTextProvider,
  classificationFailureMetadata,
  parseBatchBookClassification,
  ensureBatchFactory121ResubmissionAllowed,
  persisted121PublicationMetadata,
  v11JSONRequest,
  safeAutomationStatus,
  liveGiantAutomationRecovery,
  createBatchFactoryV11Router,
  validateBatchFactoryModelPatch,
  resolveBatchFactoryRuntimeSettings,
  syncPersonalProviderConfig,
  syncH3ProviderConfig,
  preflightBatchFactoryAutomationVideo,
  batchFactoryProductionText,
  splitVideoPresetBody,
  singleBookDirectorStageTarget,
  generateOpeningVariantsAfterSingleDirector,
  fetchBatchFactoryAutomationDirectSource
} = require('./batch-factory-v11');

test('automation direct source refill uses the same book-store source write as manual retrieval', async () => {
  const writes = [];
  const result = await fetchBatchFactoryAutomationDirectSource({
    owner: 'alice',
    isOwner: false,
    batch: { id: 'batch-1' },
    book: { id: 'book-1', bookId: '101', platform: '3', sourceText: '', revision: 2, sourceMetadata: { sourceMode: 'giant_material', contentPending: true } },
    options: {
      workshopStoreFactory: () => ({
        getPlatforms: () => [{ id: '3', name: '七猫付费' }],
        fetchDirectOriginal: async () => ({ text: '书城正文', attempts: 1 })
      })
    },
    upstreamOptions: {
      goBaseUrl: 'http://go.local',
      bridgeSecret: 'secret',
      fetchImpl: async (url, init) => {
        writes.push({ pathname: new URL(url).pathname, payload: JSON.parse(init.body) });
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
    }
  });

  assert.deepEqual(result, { state: 'succeeded', characters: 4 });
  assert.deepEqual(writes, [{
    pathname: '/api/batch-factory/v11/batches/batch-1/books/book-1/source',
    payload: {
      sourceText: '书城正文',
      expectedRevision: 2,
      sourceMetadata: {
        sourceMode: 'giant_material', contentPending: false,
        sourceFetchedAt: writes[0]?.payload?.sourceMetadata?.sourceFetchedAt,
        sourceFetchAttempts: 1, sourceCaptureCharacters: 4000, sourceBookId: '101', sourceOriginalRaw: '书城正文',
        platformId: '3',
        originalReadStage: 'completed', originalReadVia: 'bookstore', originalReadError: ''
      }
    }
  }]);
});

test('runtime settings inherit the current enabled account default using its catalogue ID', () => {
  const batch = { settingsState: { patch: { aiPromptConfig: { assets: { enabled: true } } } } };
  const book = {
    settingsState: {
      patch: { aiPromptConfig: { visual: { enabled: false } } }
    }
  };
  const configReader = username => {
    assert.equal(username, 'alice');
    return { model: 'provider-default', modelCatalog: [
      { id: 'account-default', kind: 'text', modelId: 'provider-default', enabled: true }
    ] };
  };
  assert.deepEqual(resolveBatchFactoryRuntimeSettings({ username: 'alice', batch, book, configReader }), {
    textModelId: 'account-default',
    aiPromptConfig: { assets: { enabled: true }, visual: { enabled: false } },
    publishSettings: {}
  });
});

test('runtime account inheritance respects explicit book selections and blank clears', () => {
  const batch = { settingsState: { patch: { textModelId: 'batch-text' } } };
  const configReader = () => { throw new Error('explicit selection must not read a fallback'); };
  for (const textModelId of ['book-text', '']) {
    const book = { settingsState: { patch: { textModelId } } };
    assert.equal(resolveBatchFactoryRuntimeSettings({ username: 'alice', batch, book, configReader }).textModelId, textModelId);
  }
});

test('runtime settings do not inherit a disabled account default', () => {
  const settings = resolveBatchFactoryRuntimeSettings({ username: 'alice', batch: {}, book: {},
    configReader: () => ({ model: 'account-disabled', modelCatalog: [{ id: 'account-disabled', modelId: 'provider-disabled', kind: 'text', enabled: false }] })
  });
  assert.equal(Object.hasOwn(settings, 'textModelId'), false);
});

function modelValidationOptions() {
  return {
    memberStore: { getMember: username => ({ username, active: true, role: 'manager' }) },
    configReader: () => ({ modelCatalogVersion: 1, modelCatalog: [
      { id: 'text-preset', kind: 'text', enabled: true, baseUrl: 'https://text.example/v1', modelId: 'text-provider', credential: 'test-key' },
      { id: 'image-current', kind: 'image', enabled: true, baseUrl: 'https://image.example/v1', modelId: 'image-provider', credential: 'test-key' },
      { id: 'video-current', kind: 'video', enabled: true, baseUrl: 'https://video.example/v1', modelId: 'video-provider', credential: 'test-key' },
      { id: 'text-disabled', kind: 'text', enabled: false, credential: 'test-key' }
    ] })
  };
}

for (const patch of [{ textModelId: 'old-text' }, { textModelId: 'text-disabled' }, { textModelId: 'image-current' }]) {
  test(`rejects unavailable or wrong-kind text selection ${patch.textModelId}`, () => {
    assert.throws(() => validateBatchFactoryModelPatch({ username: 'alice', patch, ...modelValidationOptions() }),
      error => error.status === 422 && /文本模型不可用、未配置或尚未启用/.test(error.message));
  });
}

test('validates all three model kinds without rejecting missing inherited book fields or mutating the patch', () => {
  const patch = { textModelId: 'text-preset', imageModelId: 'image-current', videoModelId: 'video-current' };
  assert.doesNotThrow(() => validateBatchFactoryModelPatch({ username: 'alice', patch, ...modelValidationOptions() }));
  assert.deepEqual(patch, { textModelId: 'text-preset', imageModelId: 'image-current', videoModelId: 'video-current' });
  assert.doesNotThrow(() => validateBatchFactoryModelPatch({ username: 'alice', patch: { openingEnabled: false, textModelId: '' }, configReader: () => ({}) }));
  for (const key of ['imageModelId', 'videoModelId']) {
    assert.throws(() => validateBatchFactoryModelPatch({ username: 'alice', patch: { [key]: 'text-preset' }, ...modelValidationOptions() }), error => error.status === 422);
  }
});

for (const target of ['settings', 'books/book-1/override']) {
  test(`V11 ${target} rejects an invalid selected model before bridge persistence`, async t => {
    const writes = [];
    const app = express();
    app.use(express.json());
    app.use((req, res, next) => { req.username = 'alice'; next(); });
    app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
      ...modelValidationOptions(), goBaseUrl: 'http://go.local', bridgeSecret: 'secret', automationController: {}, presetStore: presetStore(),
      fetchImpl: async (url, init) => { writes.push({ url, init }); return new Response('{}', { status: 200 }); }
    }));
    const server = app.listen(0, '127.0.0.1');
    t.after(() => new Promise(resolve => server.close(resolve)));
    await new Promise(resolve => server.once('listening', resolve));
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/batches/batch-1/${target}`, {
      method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ patch: { textModelId: 'old-text' }, expectedRevision: 7 })
    });
    assert.equal(response.status, 422);
    assert.match((await response.json()).error, /文本模型不可用/);
    assert.deepEqual(writes, []);
  });
}

for (const target of ['intakes/manual', 'intakes/intake-1/batches']) {
  test(`V11 ${target} rejects an invalid initial unified model before batch creation`, async t => {
    const writes = [];
    const app = express();
    app.use(express.json());
    app.use((req, res, next) => { req.username = 'alice'; next(); });
    app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
      ...modelValidationOptions(), goBaseUrl: 'http://go.local', bridgeSecret: 'secret', automationController: {}, presetStore: presetStore(),
      fetchImpl: async (url, init) => { writes.push({ url, init }); return new Response('{}', { status: 201 }); }
    }));
    const server = app.listen(0, '127.0.0.1');
    t.after(() => new Promise(resolve => server.close(resolve)));
    await new Promise(resolve => server.once('listening', resolve));
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/${target}`, {
      method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ title: '首次配置', initialBatchSettings: { textModelId: 'old-text' } })
    });
    assert.equal(response.status, 422);
    assert.match((await response.json()).error, /文本模型不可用/);
    assert.deepEqual(writes, []);
  });
}

for (const invalidTarget of ['preset', 'book']) {
  test(`automation start rejects an unavailable ${invalidTarget} model before batch persistence or queueing`, async t => {
    const calls = [];
    let starts = 0;
    const app = express();
    app.use(express.json());
    app.use((req, res, next) => { req.username = 'alice'; next(); });
    app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
      ...modelValidationOptions(), goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
      automationPresetStore: { get: async () => ({ id: 'preset-1', config: { textModelId: invalidTarget === 'preset' ? 'text-disabled' : 'text-preset' } }) },
      automationController: { start: async () => { starts++; return {}; } },
      fetchImpl: async (url, init) => {
        calls.push(init.method);
        return new Response(JSON.stringify({ batch: { id: 'batch-1', revision: 7, settingsState: { patch: {} },
          books: [{ id: 'book-1', settingsState: { patch: invalidTarget === 'book' ? { imageModelId: 'old-image' } : {} } }] } }), { status: 200 });
      }
    }));
    const server = app.listen(0, '127.0.0.1');
    t.after(() => new Promise(resolve => server.close(resolve)));
    await new Promise(resolve => server.once('listening', resolve));
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/batches/batch-1/automation/start`, {
      method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ presetId: 'preset-1' })
    });
    assert.equal(response.status, 422);
    assert.match((await response.json()).error, /模型不可用/);
    assert.deepEqual(calls, ['GET']);
    assert.equal(starts, 0);
  });
}

for (const stage of ['opening', 'visual']) {
  test(`manual ${stage} dispatch rejects wrong-kind selected models before any bridge request`, async t => {
    const calls = [];
    const app = express();
    app.use(express.json());
    app.use((req, res, next) => { req.username = 'alice'; next(); });
    app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
      ...modelValidationOptions(), goBaseUrl: 'http://go.local', bridgeSecret: 'secret', automationController: {},
      fetchImpl: async (url, init) => { calls.push({ url, init }); return new Response('{}', { status: 201 }); }
    }));
    const server = app.listen(0, '127.0.0.1');
    t.after(() => new Promise(resolve => server.close(resolve)));
    await new Promise(resolve => server.once('listening', resolve));
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/batches/batch-1/books/book-1/stages/${stage}`, {
      method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ imageModelId: 'text-preset' })
    });
    assert.equal(response.status, 422);
    assert.match((await response.json()).error, /图片模型不可用/);
    assert.deepEqual(calls, []);
  });
}

test('automation revalidates the live catalogue before stage dispatch after a selected image model is disabled', async t => {
  const fs = require('node:fs/promises');
  const os = require('node:os');
  const path = require('node:path');
  const stateDir = await fs.mkdtemp(path.join(os.tmpdir(), 'bf-model-validation-'));
  const options = modelValidationOptions();
  const config = options.configReader();
  const calls = [];
  const batch = { id: 'batch-1', settingsState: { patch: { textModelId: 'text-preset', imageModelId: 'image-current' } },
    books: [{ id: 'book-1', revision: 1, sourceText: '测试小说正文', settingsState: { patch: {} } }] };
  const router = createBatchFactoryV11Router({
    ...options, configReader: () => config, goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    automationStatePath: path.join(stateDir, 'automation.json'), automationRecoveryEnabled: false, automationPollMs: 60000,
    fetchImpl: async (url, init) => {
      calls.push({ method: init.method, pathname: new URL(url).pathname });
      if (url.endsWith('/stages')) {
        config.modelCatalog.find(model => model.id === 'image-current').enabled = false;
        return new Response(JSON.stringify({ summary: { runs: [] } }), { status: 200 });
      }
      if (url.endsWith('/batches/batch-1')) return new Response(JSON.stringify({ batch }), { status: 200 });
      return new Response(JSON.stringify({ jobs: [] }), { status: 200 });
    }
  });
  const controller = router.automationController;
  t.after(async () => {
    await controller.cancel({ owner: 'alice', batchId: 'batch-1' });
    await fs.rm(stateDir, { recursive: true, force: true });
  });
  await controller.start({ owner: 'alice', batchId: 'batch-1', runMode: 'storyboard_only' });
  let status;
  for (let attempt = 0; attempt < 100; attempt++) {
    await new Promise(resolve => setTimeout(resolve, 5));
    status = controller.status({ owner: 'alice', batchId: 'batch-1' });
    const saved = JSON.parse(await fs.readFile(path.join(stateDir, 'automation.json'), 'utf8'));
    if (Object.values(saved.jobs).some(job => job.state === 'needs_attention')) {
      await new Promise(resolve => setImmediate(resolve));
      break;
    }
  }
  assert.match(status?.books?.[0]?.error || '', /图片模型不可用/);
  assert.equal(calls.some(call => call.method !== 'GET'), false, 'no stage/provider request or settings write is allowed');
});

test('automation start persists the selected preset as unified settings before queueing metadata', async t => {
  const calls = [];
  let batch = {
    id: 'batch-1', revision: 7,
    settingsState: { patch: {
      textModelId: 'text-current', imageModelId: 'image-current',
      automationPresetSnapshot: { id: 'legacy' },
      aiPromptConfig: { assets: { enabled: true }, video: { presetId: 'current-video' } },
      publishSettings: { organization: 'current-org', category: 'NEW_BOOK' }
    } }
  };
  const app = express();
  app.use(express.json());
  app.use((req, res, next) => { req.username = 'alice'; next(); });
  app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
    ...modelValidationOptions(),
    goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    automationPresetStore: {
      get: async () => ({ id: 'preset-1', name: '预设', version: 2, config: {
        textModelId: 'text-preset', openingEnabled: false, openingCount: 0,
        automationPresetSnapshot: { id: 'preset-legacy' },
        aiPromptConfig: { video: { presetId: 'preset-video' } },
        publishSettings: { category: 'FINISHED_BOOK', startTime: '' }
      } })
    },
    automationController: {
      start: async input => { calls.push({ start: input }); return { state: 'scheduled' }; }
    },
    fetchImpl: async (url, options) => {
      const payload = options.body ? JSON.parse(options.body) : undefined;
      calls.push({ method: options.method, pathname: new URL(url).pathname, payload });
      if (options.method === 'PUT') {
        assert.equal(payload.expectedRevision, batch.revision);
        const patch = { ...batch.settingsState.patch };
        for (const key of payload.restoreKeys || []) delete patch[key];
        batch = { ...batch, revision: 8, settingsState: { patch: { ...patch, ...payload.patch } } };
      }
      return { ok: true, status: 200, text: async () => JSON.stringify({ batch }) };
    }
  }));
  const server = app.listen(0, '127.0.0.1');
  t.after(() => new Promise(resolve => server.close(resolve)));
  await new Promise(resolve => server.once('listening', resolve));
  const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/batches/batch-1/automation/start`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ presetId: 'preset-1', runMode: 'storyboard_only', concurrency: 4, scheduledAt: '2026-10-02T10:01:00.000Z' })
  });
  assert.equal(response.status, 201);
  // The background classification sidecar refreshes the batch once after the
  // durable preset write; the user-facing automation response still returns
  // only after the four primary calls below have completed.
  assert.ok(calls.length >= 4);
  assert.deepEqual(calls.slice(0, 3), [
    { method: 'GET', pathname: '/api/batch-factory/v11/batches/batch-1', payload: undefined },
    { method: 'PUT', pathname: '/api/batch-factory/v11/batches/batch-1/settings', payload: {
      expectedRevision: 7, restoreKeys: ['automationPresetSnapshot'], patch: {
        textModelId: 'text-preset', imageModelId: 'image-current', openingEnabled: false, openingCount: 0,
        aiPromptConfig: { assets: { enabled: true }, video: { presetId: 'preset-video' } },
        publishSettings: { organization: 'current-org', category: 'FINISHED_BOOK', startTime: '' }
      }
    } },
    { method: 'GET', pathname: '/api/batch-factory/v11/batches/batch-1', payload: undefined }
  ]);
  assert.equal(batch.settingsState.patch.textModelId, 'text-preset');
  assert.equal(Object.hasOwn(batch.settingsState.patch, 'automationPresetSnapshot'), false);
  assert.deepEqual(calls[3].start, {
    owner: 'alice', isOwner: false, batchId: 'batch-1',
    scheduledAt: '2026-10-02T10:01:00.000Z', runMode: 'storyboard_only', concurrency: 4,
    preset: { id: 'preset-1', name: '预设', version: 2 }
  });
});

for (const failedRequest of [1, 2, 3]) {
  test(`automation does not queue when preset application bridge request ${failedRequest} fails`, async t => {
    let requests = 0;
    let starts = 0;
    const app = express();
    app.use(express.json());
    app.use((req, res, next) => { req.username = 'alice'; next(); });
    app.use('/api/batch-factory/v11', createBatchFactoryV11Router({
      ...modelValidationOptions(),
      goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
      automationPresetStore: { get: async () => ({ id: 'preset-1', config: { textModelId: 'text-preset' } }) },
      automationController: { start: async () => { starts++; return { state: 'running' }; } },
      fetchImpl: async () => {
        requests++;
        const failed = requests === failedRequest;
        return {
          ok: !failed, status: failed ? 409 : 200,
          text: async () => JSON.stringify(failed ? { error: 'revision conflict' } : {
            batch: { id: 'batch-1', revision: 7, settingsState: { patch: {} } }
          })
        };
      }
    }));
    const server = app.listen(0, '127.0.0.1');
    t.after(() => new Promise(resolve => server.close(resolve)));
    await new Promise(resolve => server.once('listening', resolve));
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/batch-factory/v11/batches/batch-1/automation/start`, {
      method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ presetId: 'preset-1' })
    });
    assert.equal(response.status, 409);
    assert.equal(starts, 0);
    assert.equal(requests, failedRequest);
  });
}

test('recovers giant automation using its saved plan without requiring a frozen preset', () => {
  const recovered = liveGiantAutomationRecovery({
    settingsState: { patch: { textModelId: 'text-current' } }
  }, { presetId: 'preset-1', runMode: 'full_submit', concurrency: 2, scheduledAt: '2026-10-02T10:01:00.000Z' });

  assert.deepEqual(recovered, {
    preset: { id: 'preset-1' },
    runMode: 'full_submit', concurrency: 2, scheduledAt: '2026-10-02T10:01:00.000Z'
  });
});

test('giant recovery does not require historical preset metadata to match its saved plan', () => {
  const recovered = liveGiantAutomationRecovery({
    settingsState: { patch: { automationPresetSnapshot: { id: 'old-preset' }, textModelId: 'text-current' } }
  }, { presetId: 'saved-preset', runMode: 'storyboard_only', concurrency: 4 });
  assert.deepEqual(recovered, { preset: { id: 'saved-preset' }, runMode: 'storyboard_only', concurrency: 4, scheduledAt: '' });
});

test('giant recovery validates its saved plan identity and run timing', () => {
  assert.throws(() => liveGiantAutomationRecovery({}, {}), error => error.code === 'GIANT_AUTOMATION_PLAN_MISSING');
  assert.throws(() => liveGiantAutomationRecovery({
    settingsState: { patch: { automationPresetSnapshot: { id: 'preset-1' } } }
  }, { presetId: 'preset-1', scheduledAt: 'invalid-date' }), /定时执行时间无效/);
});

test('keeps book-city loading placeholders out of automated production text', () => {
  assert.equal(
    batchFactoryProductionText({
      sourceText: '修改中&nbsp;\n。\n第一段 <b>正文</b>\n第二段正文',
      sourceMetadata: { contentRangeLines: 2 }
    }),
    '第一段 正文\n第二段正文'
  );
});

test('keeps a V12 book submission on the Node-owned 121 publisher', () => {
  assert.deepEqual(
    batchFactory121PublishPath('/api/batch-factory/v12/batches/batch-1/books/book-1/publish-121'),
    { batchId: 'batch-1', bookId: 'book-1' }
  );
});

test('uses the first enabled text model to classify a newly imported book without batch settings', () => {
  const provider = resolveBatchFactoryBookClassificationTextProvider(
    { username: 'alice', body: {} },
    { settingsState: { patch: {} } },
    { settingsState: { patch: {} } },
    {
      memberStore: { getMember: username => ({ username, active: true, role: 'manager' }) },
      configReader: () => ({ modelCatalog: [{ id: 'text-auto', kind: 'text', enabled: true, baseUrl: 'https://text.example/v1', modelId: 'gpt-classifier', credential: 'classify-key', displayName: '自动分类模型' }] })
    }
  );
  assert.deepEqual(provider, {
    endpoint: 'https://text.example/v1/chat/completions',
    apiKey: 'classify-key',
    model: 'gpt-classifier',
    displayName: '自动分类模型'
  });
});

test('keeps existing publication fields while persisting a visible classification failure', () => {
  const metadata = classificationFailureMetadata({
    gender: '女频', style: '现代甜文', tags: '追妻', classifyStatus: 'classified'
  }, new Error('请先在单书配置或引擎配置中选择已启用的文本模型'), () => new Date('2026-09-25T10:00:00.000Z'));

  assert.deepEqual(metadata, {
    gender: '女频',
    style: '现代甜文',
    tags: '追妻',
    classifyStatus: 'failed',
    classifyError: '请先在单书配置或引擎配置中选择已启用的文本模型',
    classifyAt: '2026-09-25T10:00:00.000Z'
  });
});

test('resolves the selected Seedance model instead of reusing the YD credential', () => {
  const { resolveBatchVideoProviderConfig } = require('./batch-factory-v11');
  const config = resolveBatchVideoProviderConfig('owner', 'seedance-2-0-official', {
    accountStore: { getInternalAccount: () => ({ isOwner: true }) },
    configReader: () => ({ modelCatalog: [
      { id: 'yd2-mini-video', kind: 'video', enabled: true, adapterKind: 'openai_video', credential: 'yd-key' },
      { id: 'seedance-2-0-official', kind: 'video', enabled: true, adapterKind: 'yfai_seedance', baseUrl: 'https://yf.token6688.com', modelId: 'seedance-2-0-official', credential: 'seedance-key' }
    ] })
  });
  assert.deepEqual(config, { provider: 'yfai_seedance', model: 'seedance-2-0-official', apiKey: 'seedance-key', baseUrl: 'https://yf.token6688.com' });
});

test('synchronizes the selected Seedance provider when a video-stage request only carries its provider', async () => {
  const writes = [];
  await syncPersonalProviderConfig({
    username: 'owner',
    auth: { account: { isOwner: true } },
    body: { provider: 'yfai_seedance' }
  }, {
    goBaseUrl: 'http://go.local',
    bridgeSecret: 'test-secret',
    now: () => 0,
    accountStore: { getInternalAccount: () => ({ isOwner: true }) },
    configReader: () => ({ modelCatalog: [
      { id: 'seedance-2-0-official', kind: 'video', enabled: true, adapterKind: 'yfai_seedance', baseUrl: 'https://yf.token6688.com', modelId: 'seedance-2-0-official', credential: 'seedance-key' }
    ] }),
    fetchImpl: async (url, init) => {
      writes.push({ url, payload: JSON.parse(init.body) });
      return new Response('{}', { status: 200 });
    }
  });

  assert.deepEqual(writes, [{
    url: 'http://go.local/api/batch-factory/v11/video-provider/config',
    payload: {
      provider: 'yfai_seedance',
      model: 'seedance-2-0-official',
      apiKey: 'seedance-key',
      createUrl: 'https://yf.token6688.com'
    }
  }]);
});

test('synchronizes the configured AutoDL H3 workflow instead of the hard-coded fallback', async () => {
  const writes = [];
  await syncH3ProviderConfig({
    username: 'owner',
    auth: { account: { isOwner: true } },
    body: { provider: 'autodl_comfyui', videoModelId: 'minimax-h3-video' }
  }, {
    goBaseUrl: 'http://go.local',
    bridgeSecret: 'test-secret',
    now: () => 0,
    accountStore: { getInternalAccount: () => ({ isOwner: true }) },
    configReader: () => ({
      modelCatalog: [{
        id: 'minimax-h3-video', kind: 'video', enabled: true,
        adapterKind: 'autodl_comfyui_video', credential: 'h3-key',
        workflowId: 'minimax_h3_image_audio_to_video_v1'
      }]
    }),
    fetchImpl: async (url, init) => {
      writes.push({ url, payload: JSON.parse(init.body) });
      return new Response('{}', { status: 200 });
    }
  });

  assert.deepEqual(writes, [{
    url: 'http://go.local/api/batch-factory/v11/video-provider/config',
    payload: {
      provider: 'autodl_comfyui',
      model: 'minimax-h3-video',
      apiKey: 'h3-key',
      createUrl: 'https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}',
      tasksUrl: 'https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}',
      workflowId: 'minimax_h3_image_audio_to_video_v1'
    }
  }]);
});

test('synchronizes an authorized member H3 request with the bound manager credential', async () => {
  const writes = [];
  await syncH3ProviderConfig({
    username: 'member',
    auth: { account: { isOwner: false } },
    body: { provider: 'autodl_comfyui', videoModelId: 'minimax-h3-video' }
  }, {
    goBaseUrl: 'http://go.local',
    bridgeSecret: 'test-secret',
    now: () => 0,
    memberStore: {
      getMember(username) {
        return username === 'member'
          ? { username, active: true, role: 'member', boundTo: 'manager' }
          : { username, active: true, role: 'manager' };
      },
      canUseApi: () => true
    },
    accountStore: { getInternalAccount: () => ({ isOwner: false }) },
    configReader: username => username === 'manager'
      ? { modelCatalog: [{
        id: 'minimax-h3-video', kind: 'video', enabled: true,
        adapterKind: 'autodl_comfyui_video', credential: 'manager-h3-key',
        workflowId: 'manager-h3-workflow'
      }] }
      : { modelCatalog: [] },
    fetchImpl: async (url, init) => {
      writes.push({ url, payload: JSON.parse(init.body) });
      return new Response('{}', { status: 200 });
    }
  });

  assert.deepEqual(writes, [{
    url: 'http://go.local/api/batch-factory/v11/video-provider/config',
    payload: {
      provider: 'autodl_comfyui',
      model: 'minimax-h3-video',
      apiKey: 'manager-h3-key',
      createUrl: 'https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}',
      tasksUrl: 'https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}',
      workflowId: 'manager-h3-workflow'
    }
  }]);
});

test('automation H3 preflight accepts the authorized manager model without a member key', () => {
  const result = preflightBatchFactoryAutomationVideo({
    username: 'member',
    isOwner: false,
    batch: {
      settingsState: { patch: { videoModelId: 'minimax-h3-video' } },
      books: [{ id: 'book-1', settingsState: { patch: {} } }]
    },
    options: {
      memberStore: {
        getMember(username) {
          return username === 'member'
            ? { username, active: true, role: 'member', boundTo: 'manager' }
            : { username, active: true, role: 'manager' };
        },
        canUseApi: () => true
      },
      accountStore: { getInternalAccount: () => ({ isOwner: false }) },
      configReader: username => username === 'manager'
        ? { modelCatalog: [{
          id: 'minimax-h3-video', kind: 'video', enabled: true,
          adapterKind: 'autodl_comfyui_video', credential: 'manager-h3-key'
        }] }
        : { modelCatalog: [] }
    }
  });
  assert.equal(result, true);
});

test('uses the bound manager video credential for an authorized member on the legacy personal provider', () => {
  const { resolveBatchVideoProviderConfig } = require('./batch-factory-v11');
  const config = resolveBatchVideoProviderConfig('member', 'yd2-mini-video', {
    memberStore: {
      getMember(username) {
        return username === 'member'
          ? { username, active: true, role: 'member', boundTo: 'manager' }
          : { username, active: true, role: 'manager' };
      },
      canUseApi: () => true
    },
    accountStore: { getInternalAccount: () => ({ isOwner: false }) },
    configReader: username => username === 'manager'
      ? { modelCatalog: [{ id: 'yd2-mini-video', kind: 'video', enabled: true, adapterKind: 'openai_video', credential: 'manager-video-key' }] }
      : { modelCatalog: [] }
  });
  assert.deepEqual(config, { provider: 'personal_api', model: 'yd2.0-mini', apiKey: 'manager-video-key' });
});

test('reads the real duration from MPEG audio frames used by automatic planning', () => {
  const frameLength = Math.floor((144000 * 128) / 44100);
  const frame = Buffer.alloc(frameLength);
  frame.writeUInt32BE(0xfffb9000, 0);
  const seconds = mpegAudioDurationSeconds(Buffer.concat(Array.from({ length: 10 }, () => frame)));
  assert.ok(Math.abs(seconds - ((10 * 1152) / 44100)) < 0.002);
});

test('Batch Factory 121 session auto-recovers through the shared durable login service', async () => {
  const calls = [];
  const session = await resolveBatchFactory121Session('alice', {
    sessionStore: { getSession: () => null, getBrowserSession: () => null },
    webSubmit: { async ensureSession(owner) { calls.push(owner); return { request: { sessionKey: 'PHPSESSID=renewed' } }; } }
  });
  assert.deepEqual(calls, ['alice']);
  assert.equal(session.cookie, 'PHPSESSID=renewed');
});

test('style.system runs during asset extraction but never when a director stage is requested', () => {
  const assetPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/assets';
  const directorPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/director';
  assert.deepEqual(styleSystemBookPath(assetPath), { batchId: 'batch-1', bookId: 'book-1' });
  assert.equal(styleSystemBookPath(directorPath), null);
});

test('automation status degrades to a readable idle state when its controller throws', () => {
  const result = safeAutomationStatus({
    status() { throw new Error('legacy automation state is unreadable'); }
  }, { owner: 'alice', batchId: 'batch-1' });

  assert.equal(result.state, 'unavailable');
  assert.equal(result.diagnosticCode, 'AUTOMATION_STATUS_UNAVAILABLE');
  assert.deepEqual(result.counts, { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 });
});

test('splits one H3 video preset into director rules and its final prompt template', () => {
  const result = splitVideoPresetBody('原版导演规则\ncharacter_slot_ids\n【批量工厂最终 Prompt 模板】\n{{storyboard}}');
  assert.equal(result.directorRules, '原版导演规则\ncharacter_slot_ids');
  assert.equal(result.finalTemplate, '{{storyboard}}');
});

test('identifies the explicit per-book 121 publish action without matching other V11 routes', () => {
  assert.deepEqual(
    batchFactory121PublishPath('/api/batch-factory/v11/batches/batch-1/books/book-1/publish-121'),
    { batchId: 'batch-1', bookId: 'book-1' }
  );
  assert.equal(batchFactory121PublishPath('/api/batch-factory/v11/batches/batch-1/status'), null);
  assert.equal(batchFactory121OrganizationsPath('/api/batch-factory/v11/publish-121/organizations'), true);
  assert.equal(batchFactory121OrganizationsPath('/publish-121/organizations'), true);
});

test('normalizes only usable 121 organization options', () => {
  assert.deepEqual(organizationOptions({ success: true, data: [
    { id: 1, name: '博量', level: 1 },
    { organization_id: 3, organization_name: '璇奕组织', level_name: '层2' },
    { id: '', name: '无效' }
  ] }), [
    { id: '1', name: '博量', level: '1' },
    { id: '3', name: '璇奕组织', level: '层2' }
  ]);
});

test('automation publish settings keep an explicit frozen blank instead of falling back to live batch settings', () => {
  const batch = { settingsState: { patch: { publishSettings: { organization: 'live-org', category: 'LIVE' } } } };
  const book = { sourceText: '完整视频原文', settingsState: { patch: {} } };
  assert.deepEqual(automationPublishSettings(batch, book, { publishSettings: { organization: '', category: 'FROZEN' } }), { organization: '', category: 'FROZEN' });
});

test('blocks an already uploaded book until the caller explicitly requests a reupload', () => {
  const book = { sourceMetadata: { websiteSubmitStatus: 'uploaded' } };
  assert.throws(() => ensureBatchFactory121ResubmissionAllowed(book, {}), /已上传，请使用重新上传/);
  assert.doesNotThrow(() => ensureBatchFactory121ResubmissionAllowed(book, { reupload: true }));
});

test('keeps append-only 121 submission history and the latest readback progress', () => {
  const metadata = persisted121PublicationMetadata({
    websiteSubmitHistory: [{ submittedAt: '2026-09-16T00:00:00.000Z', status: 'confirmed' }]
  }, {
    status: 'accepted_pending',
    uploadedAt: '2026-09-17T00:00:00.000Z',
    sourceTextFile: '208.txt',
    aiHeadVideoFile: '208.mp4',
    receipt: { verified: false, remote_record: { found: false, detail: '等待 121 回读' } }
  });
  assert.equal(metadata.websiteSubmitHistory.length, 2);
  assert.deepEqual(metadata.websiteSubmitProgress, {
    phase: 'readback', status: 'waiting', message: '等待 121 回读', updatedAt: '2026-09-17T00:00:00.000Z'
  });
});

test('submits only the requested book to the direct 121 publisher', async () => {
  const requests = [];
  const actions = [];
  const batch = {
    id: 'batch-1',
    settingsState: { patch: { publishSettings: { websiteProfileId: 'profile-1', platformId: '15', organization: '1', gender: '女', styleType: '现代虐文' } } },
    books: [{ id: 'book-1', bookId: '2080310440299710279', platform: '15', sourceText: '正文', sourceMetadata: { sourceMode: 'manual_original', gender: '女频', style: '现代虐文', classifyStatus: 'classified' }, revision: 1, settingsState: { patch: {} } }]
  };
  const req = {
    username: 'alice',
    auth: { account: { isOwner: false } },
    app: { locals: { novelFetchStore: { getSession: () => ({ cookie: 'PHPSESSID=ready' }) } } }
  };
  const responseFor = value => ({
    ok: true,
    status: 200,
    text: async () => JSON.stringify(value),
    arrayBuffer: async () => Buffer.from([0, 1, 2, 3]).buffer
  });
  const result = await submitBatchFactoryBookTo121(req, { batchId: 'batch-1', bookId: 'book-1' }, {
    goBaseUrl: 'http://v11.test',
    bridgeSecret: 'bridge',
    directClient: {
      verify: async () => ({ ok: true }),
      action: async request => {
        if (request.path === '/tttadmin/api/music_put_url.php') {
          return { body: JSON.stringify({ success: true, data: { bucket_url: 'https://assets.121.test', files: [{ name: 'head.mp4', upload_url: 'https://upload.121.test/head.mp4', object_key: 'heads/head.mp4', headers: {} }] } }) };
        }
        actions.push(request);
        return { body: JSON.stringify(request.method === 'GET' ? { success: true, data: [] } : { success: true, result: { success: { count: 1 }, failed: { count: 0 } } }) };
      },
      uploadPresigned: async () => ({ status: 200 })
    },
    fetchImpl: async (url, options) => {
      requests.push({ url, options });
      if (url.endsWith('/batches/batch-1')) return responseFor({ batch });
      if (url.endsWith('/batches/batch-1/books/book-1/metadata') && options.method === 'PUT') return responseFor({ book: { ...batch.books[0], revision: 2 } });
      if (url.endsWith('/batches/batch-1/merge-status')) return responseFor({ jobs: [{ bookId: 'book-1', status: 'succeeded', outputUrl: '/api/batch-factory/v11/media/merge-1.mp4' }] });
      if (url.endsWith('/api/batch-factory/v11/media/merge-1.mp4')) return { ok: true, status: 200, arrayBuffer: async () => Uint8Array.from([0, 1, 2, 3]).buffer };
      throw new Error(`unexpected URL ${url}`);
    }
  });

  assert.equal(result.status, 'accepted_pending');
  assert.equal(actions.length, 2);
  assert.equal(actions[0].method, 'POST');
  assert.match(actions[0].body.toString('latin1'), /filename="2080310440299710279\.txt"/);
  assert.match(actions[0].body.toString('utf8'), /name="jieya_ai_head_video"\r\n\r\n\[{"name":"head\.mp4"/);
  assert.equal(requests.some(item => item.url.endsWith('/batches/batch-1/status')), false);
  const metadataWrite = requests.filter(item => item.url.endsWith('/batches/batch-1/books/book-1/metadata')).at(-1);
  assert.ok(metadataWrite, 'the 121 receipt must be persisted to the current book');
  const metadataPayload = JSON.parse(metadataWrite.options.body);
  assert.equal(metadataPayload.expectedRevision, 1);
  assert.equal(metadataPayload.metadata.websiteSubmitStatus, 'submitted');
  assert.equal(metadataPayload.metadata.websiteSubmitProgress.phase, 'readback');
  assert.equal(metadataPayload.metadata.websiteSubmitProgress.status, 'waiting');
  assert.equal(metadataPayload.metadata.websiteSubmitHistory.length, 1);
  const progressPhases = requests
    .filter(item => item.url.endsWith('/batches/batch-1/books/book-1/metadata') && item.options.method === 'PUT')
    .map(item => JSON.parse(item.options.body).metadata.websiteSubmitProgress?.phase)
    .filter(Boolean);
  assert.deepEqual(progressPhases, ['classification', 'validation', 'session', 'ai_head', 'txt_submit', 'readback', 'readback']);
});

test('classifies missing per-book publish metadata and persists it before any 121 submission', async () => {
  const calls = [];
  const book = { id: 'book-1', title: '女主重生复仇', platform: '15', sourceText: '沈薇重生回到离婚前，决定查清真相。', sourceMetadata: { tags: '' }, revision: 7 };
  const result = await classifyBatchFactoryBookFor121({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'bridge', now: () => Date.parse('2026-09-17T01:02:03.000Z'),
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'gpt-5.4', displayName: 'GPT-5.4' },
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', books: [book] } }), { status: 200 });
      if (url === 'http://text.local/v1/chat/completions') {
        return new Response(JSON.stringify({ choices: [{ message: { content: '```json\n{"gender":"女频","style":"现代虐文","tags":["重生","复仇"],"reason":"现代女性复仇线"}\n```' } }] }), { status: 200 });
      }
      if (init.method === 'PUT') return new Response(JSON.stringify({ book: { ...book, revision: 8 } }), { status: 200 });
      throw new Error(`unexpected request: ${init.method} ${url}`);
    }
  });
  assert.equal(result.reused, false);
  assert.deepEqual(result.classification, { gender: '女频', style: '现代虐文', tags: '重生、复仇', reason: '现代女性复仇线' });
  const save = calls.find(call => call.init.method === 'PUT');
  assert.ok(save, 'the classification must persist before upload');
  const payload = JSON.parse(save.init.body);
  assert.deepEqual(payload, {
    metadata: { gender: '女频', style: '现代虐文', tags: '重生、复仇', classifyStatus: 'classified', classifyReason: '现代女性复仇线', classifyModel: 'GPT-5.4', classifyAt: '2026-09-17T01:02:03.000Z' },
    expectedRevision: 7
  });
  assert.equal(calls.filter(call => call.url === 'http://text.local/v1/chat/completions').length, 1);
});

test('uses AI only for missing style and never overwrites 121-derived gender', async () => {
  const calls = [];
  const book = {
    id: 'book-121', title: '都市男频书', platform: '2', sourceText: '男主在都市创业。', revision: 3,
    sourceMetadata: { category: '男频-都市', genre: '都市', gender: '男频', genderSource: '121_category' }
  };
  await classifyBatchFactoryBookFor121({
    username: 'alice', batchId: 'batch-1', bookId: 'book-121', goBaseUrl: 'http://go.local', bridgeSecret: 'bridge',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'gpt-5.4' },
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', books: [book] } }), { status: 200 });
      if (url === 'http://text.local/v1/chat/completions') return new Response(JSON.stringify({ choices: [{ message: { content: '{"gender":"女频","style":"男频都市","tags":["创业"],"reason":"都市创业"}' } }] }), { status: 200 });
      if (init.method === 'PUT') return new Response(JSON.stringify({ book }), { status: 200 });
      throw new Error(`unexpected request: ${init.method} ${url}`);
    }
  });
  const saved = JSON.parse(calls.find(call => call.init.method === 'PUT').init.body).metadata;
  assert.equal(saved.gender, '男频');
  assert.equal(saved.genderSource, '121_category');
  assert.equal(saved.style, '男频都市');
});

test('rejects a model classification that is not a valid 121 gender and style', () => {
  assert.throws(() => parseBatchBookClassification('{"gender":"未知","style":"仙侠"}'), /必须返回男女频/);
});

test('accepts common model JSON drift with Chinese punctuation and single quotes', () => {
  const result = parseBatchBookClassification(`判断结果如下：
\`\`\`json
{'gender'：'女频'，'style'：'现代虐文'，'tags'：['重生'，'复仇']，'reason'：'现代女性复仇线'}
\`\`\``);
  assert.deepEqual(result, {
    gender: '女频',
    style: '现代虐文',
    tags: '重生、复仇',
    reason: '现代女性复仇线'
  });
});

test('asks the selected model to repair one invalid classification response before failing', async () => {
  const calls = [];
  const book = { id: 'book-1', title: '女主重生复仇', platform: '15', sourceText: '沈薇重生回到离婚前，决定查清真相。', sourceMetadata: {}, revision: 7 };
  const result = await classifyBatchFactoryBookFor121({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'bridge', now: () => Date.parse('2026-09-17T01:02:03.000Z'),
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'gpt-5.4', displayName: 'GPT-5.4' },
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', books: [book] } }), { status: 200 });
      if (url === 'http://text.local/v1/chat/completions') {
        const modelCalls = calls.filter(call => call.url === url).length;
        const content = modelCalls === 1
          ? '女频，现代虐文，标签是重生和复仇。'
          : '{"gender":"女频","style":"现代虐文","tags":["重生","复仇"],"reason":"现代女性复仇线"}';
        return new Response(JSON.stringify({ choices: [{ message: { content } }] }), { status: 200 });
      }
      if (init.method === 'PUT') return new Response(JSON.stringify({ book: { ...book, revision: 8 } }), { status: 200 });
      throw new Error(`unexpected request: ${init.method} ${url}`);
    }
  });
  assert.equal(result.reused, false);
  assert.equal(calls.filter(call => call.url === 'http://text.local/v1/chat/completions').length, 2);
  const repairPayload = JSON.parse(calls.filter(call => call.url === 'http://text.local/v1/chat/completions')[1].init.body);
  assert.match(repairPayload.messages.at(-1).content, /只返回合法 JSON/);
  assert.deepEqual(result.classification, { gender: '女频', style: '现代虐文', tags: '重生、复仇', reason: '现代女性复仇线' });
});

test('uses the runtime fetch when V11 helper receives no injected fetch', async () => {
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.fetch = async (url, options) => {
    calls.push({ url, options });
    return new Response(JSON.stringify({ ok: true }), { status: 200 });
  };
  try {
    const result = await v11JSONRequest({
      username: 'alice',
      isOwner: false,
      method: 'GET',
      pathname: '/api/batch-factory/v11/batches/batch-1',
      goBaseUrl: 'http://v11.test',
      bridgeSecret: 'bridge'
    });
    assert.deepEqual(result, { ok: true });
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, 'http://v11.test/api/batch-factory/v11/batches/batch-1');
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test('preserves a deletion retention query but signs only the Go request path', async () => {
  const calls = [];
  const now = Date.parse('2026-10-04T03:04:05.000Z');
  await v11JSONRequest({
    username: 'alice',
    isOwner: true,
    method: 'DELETE',
    pathname: '/api/batch-factory/v11/batches/batch-1?retentionDays=3',
    goBaseUrl: 'http://v11.test',
    bridgeSecret: 'bridge',
    now: () => now,
    fetchImpl: async (url, options) => {
      calls.push({ url: String(url), headers: options.headers });
      return new Response(null, { status: 204 });
    }
  });

  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, 'http://v11.test/api/batch-factory/v11/batches/batch-1?retentionDays=3');
  const expected = createSignedBridgeHeaders({
    username: 'alice', isOwner: true, method: 'DELETE',
    pathname: '/api/batch-factory/v11/batches/batch-1', secret: 'bridge', now
  });
  const incorrect = createSignedBridgeHeaders({
    username: 'alice', isOwner: true, method: 'DELETE',
    pathname: '/api/batch-factory/v11/batches/batch-1?retentionDays=3', secret: 'bridge', now
  });
  assert.equal(calls[0].headers['X-Qiantie-Signature'], expected['X-Qiantie-Signature']);
  assert.notEqual(calls[0].headers['X-Qiantie-Signature'], incorrect['X-Qiantie-Signature']);
});

test('asset preparation skips style.system when smart-unified is disabled', async () => {
  const calls = [];
  const fields = {
    imageMedium: '真人数字电影短剧', captureProcess: '数字电影摄影', grainTexture: '细腻胶片颗粒',
    filterColorSystem: '低饱和冷暖对比', lensLanguage: '克制叙事镜头语言', opticalCharacter: '柔和高光',
    contrast: '中等对比', saturation: '低饱和', lightingHierarchy: '层次化侧逆光',
    narrativeComposition: '人物关系优先', atmosphere: '克制悬疑'
  };
  const style = await acquireBatchFactorySmartUnifiedBaseline({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProviders: [{ id: 'text-model', endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' }],
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') {
        return new Response(JSON.stringify({
          batch: {
            id: 'batch-1',
            settingsState: { patch: { aiPromptConfig: { constraints: { selections: [] } } } },
            books: [{
              id: 'book-1',
              sourceText: '完整原文',
              settingsState: { patch: { aiPromptConfig: { constraints: { enabled: false } } } },
              assetRecords: []
            }]
          }
        }), { status: 200 });
      }
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
    }
  });
  assert.equal(calls.length, 1);
  assert.equal(style.style, '');
  assert.equal(style.skipped, true);
});

test('smart-unified falls back to another enabled text model after an upstream failure', async () => {
  const fields = { final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市', negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定' };
  const calls = [];
  const result = await acquireBatchFactorySmartUnifiedBaseline({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', force: true,
    textProviders: [
      { id: 'primary', endpoint: 'http://primary.local/v1/chat/completions', apiKey: 'one', model: 'primary', displayName: '主模型' },
      { id: 'backup', endpoint: 'http://backup.local/v1/chat/completions', apiKey: 'two', model: 'backup', displayName: '备用模型' }
    ],
    fetchImpl: async (url, init = {}) => {
      calls.push(url);
      if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { constraints: { selections: [{ presetId: 'script-constraint-prefix-smart-unified', constraintCategory: 'prefix' }] } } } }, books: [{ id: 'book-1', sourceText: '完整原文', settingsState: { patch: {} }, assetRecords: [] }] } }), { status: 200 });
      if (url.startsWith('http://primary.local')) return new Response(JSON.stringify({ error: { message: 'temporary unavailable' } }), { status: 502 });
      if (url.startsWith('http://backup.local')) return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
      if (init.method === 'PUT') return new Response(JSON.stringify({ book: { id: 'book-1', revision: 2 } }), { status: 200 });
      throw new Error(`unexpected ${url}`);
    }
  });
  assert.equal(result.provider.id, 'backup');
  assert.match(result.style, /现代都市短剧/);
  assert.deepEqual(result.attempts.map(item => item.status), ['failed', 'succeeded']);
});

test('smart-unified failure is non-blocking after every enabled model fails', async () => {
  const result = await acquireBatchFactorySmartUnifiedBaseline({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', force: true,
    textProviders: [{ id: 'primary', endpoint: 'http://primary.local/v1/chat/completions', apiKey: 'one', model: 'primary' }],
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { constraints: { selections: [{ presetId: 'script-constraint-prefix-smart-unified', constraintCategory: 'prefix' }] } } } }, books: [{ id: 'book-1', sourceText: '完整原文', settingsState: { patch: {} }, assetRecords: [] }] } }), { status: 200 });
      return new Response(JSON.stringify({ error: { message: 'temporary unavailable' } }), { status: 502 });
    }
  });
  assert.equal(result.style, '');
  assert.equal(result.nonBlocking, true);
  assert.equal(result.attempts.length, 1);
  assert.match(result.attempts[0].message, /temporary unavailable/);
});

test('asset preparation accepts style.system JSON wrapped in a model explanation', async () => {
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  const style = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: `分析如下：\n\n\`\`\`json\n${JSON.stringify(fields)}\n\`\`\`` } }] }), { status: 200 });
    }
  });
  assert.match(style, /现代都市短剧；高级电影感；当代都市/);
});

test('asset preparation accepts a bare style.system object surrounded by model prose', async () => {
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  const style = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: `以下是分析结果：\n${JSON.stringify(fields)}\n请按上述字段使用。` } }] }), { status: 200 });
    }
  });
  assert.match(style, /现代都市短剧；高级电影感；当代都市/);
});

test('asset preparation reads style.system text from OpenAI-compatible content parts', async () => {
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  const style = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: [{ type: 'text', text: JSON.stringify(fields) }] } }] }), { status: 200 });
    }
  });
  assert.match(style, /现代都市短剧；高级电影感；当代都市/);
});

test('asset preparation requests a JSON object from compatible style.system providers', async () => {
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  let modelRequest;
  await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    presetStore: { getPublished: id => ({ id, name: '智能统一', version: 1, body: '仅按风格分析。' }), listAll: () => [] },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      modelRequest = JSON.parse(init.body);
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
    }
  });
  assert.deepEqual(modelRequest.response_format, { type: 'json_object' });
  assert.equal(modelRequest.messages.some(message => /json/i.test(String(message.content || ''))), true);
});

test('style.system analyzes only the configured production lines', async () => {
  const requests = [];
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{
          id: 'book-1', sourceText: '一\n\n二\n三\n四\n五\n六', sourceMetadata: { contentRangeLines: 5 }, settingsState: { patch: {} }, assetRecords: []
        }] }
      }), { status: 200 });
      requests.push(JSON.parse(init.body));
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
    }
  });
  const prompt = requests[0].messages.map(message => String(message.content || '')).join('\n');
  assert.match(prompt, /一\n二\n三\n四\n五/);
  assert.doesNotMatch(prompt, /六/);
});

test('asset preparation freezes the style.system result on the book for later director use', async () => {
  const calls = [];
  const fields = {
    final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
    negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
  };
  await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', persist: true,
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: {} }, books: [{ id: 'book-1', revision: 7, sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      if (url === 'http://text.local/v1/chat/completions') return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }
  });
  assert.equal(calls.length, 3);
  assert.equal(calls[2].init.method, 'PUT');
  const frozen = JSON.parse(calls[2].init.body);
  assert.equal(frozen.expectedRevision, 7);
  assert.match(frozen.patch.h3StyleAnalysis, /现代都市短剧/);
  assert.match(frozen.patch.h3StyleSourceHash, /^[a-f0-9]{64}$/);
});

test('H3 obtains the hidden visual baseline even when smart-unified display is off', async () => {
  const calls = [];
  const fields = {
    imageMedium: '真人数字电影短剧', captureProcess: '数字电影摄影', grainTexture: '细腻胶片颗粒',
    filterColorSystem: '低饱和冷暖对比', lensLanguage: '克制叙事镜头语言', opticalCharacter: '柔和高光',
    contrast: '中等对比', saturation: '低饱和', lightingHierarchy: '层次化侧逆光',
    narrativeComposition: '人物关系优先', atmosphere: '克制悬疑'
  };
  const style = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (url, init = {}) => {
      calls.push({ url, init });
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { video: { enabled: true, presetKey: 'h3-video-normal' } } } }, books: [{ id: 'book-1', sourceText: '完整原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify(fields) } }] }), { status: 200 });
    }
  });
  assert.match(style, /影像媒介：真人数字电影短剧/);
  assert.equal(calls.length, 2);
  assert.equal(calls[0].init.method, 'GET');
  assert.equal(calls[1].init.method, 'POST');
});

test('H3 style.system uses the frozen smart-unified preset and returns a traceable analysis', async () => {
  const calls = [];
  const styleSystem = 'CUSTOM STYLE.SYSTEM RULE: return final_genre, trailer_style and story_era as JSON.';
  const result = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    presetStore: { getPublished: id => id === 'script-constraint-prefix-smart-unified' ? { id, name: '智能统一', version: 7, body: styleSystem } : null, listAll: () => [] },
    fetchImpl: async (_url, init = {}) => {
      calls.push(init);
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { video: { enabled: true, presetId: 'batch-video-h3-director' } } } }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify({
        final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
        negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
      }) } }] }), { status: 200 });
    }
  });

  assert.equal(JSON.parse(calls[1].body).messages[0].content, styleSystem);
  assert.deepEqual(JSON.parse(result), {
    schema_version: 'h3-style-system/v1',
    prompt: '现代都市短剧；高级电影感；当代都市。',
    fields: {
      final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
      negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
    },
    preset: { id: 'script-constraint-prefix-smart-unified', name: '智能统一', version: 7 }
  });
});

test('H3 style.system accepts a provider envelope that wraps its fields', async () => {
  const result = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { video: { enabled: true, presetId: 'batch-video-h3-director' } } } }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify({
        schema_version: 'h3-style-system/v1',
        prompt: '模型展示文案不作为权威来源',
        fields: {
          final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
          negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
        }
      }) } }] }), { status: 200 });
    }
  });

  const parsed = JSON.parse(result);
  assert.equal(parsed.fields.final_genre, '现代都市短剧');
  assert.equal(parsed.prompt, '现代都市短剧；高级电影感；当代都市。');
});

test('H3 style.system accepts an H3 result nested below a provider result envelope', async () => {
  const result = await analyzeBatchFactorySmartUnifiedStyle({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
    textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'text-model' },
    fetchImpl: async (_url, init = {}) => {
      if (init.method === 'GET') return new Response(JSON.stringify({
        batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { video: { enabled: true, presetId: 'batch-video-h3-director' } } } }, books: [{ id: 'book-1', sourceText: '完整视频原文', settingsState: { patch: {} }, assetRecords: [] }] }
      }), { status: 200 });
      return new Response(JSON.stringify({ choices: [{ message: { content: JSON.stringify({ result: { fields: {
        final_genre: '现代都市短剧', genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市',
        negative_prompt: '无畸形', picture_limit_prompt: '无字幕', quality_constraint_prompt: '画面稳定'
      } } }) } }] }), { status: 200 });
    }
  });

  assert.equal(JSON.parse(result).fields.final_genre, '现代都市短剧');
});

test('returns a smart-unified provider credential failure instead of calling the Go service unavailable', () => {
  const error = Object.assign(new Error('智能统一视觉分析模型“gemini-3.5-flash-maxthinking”请求失败：当前无可用凭证'), {
    code: 'SMART_UNIFIED_PROVIDER_FAILED'
  });
  assert.equal(upstreamErrorMessage(error, 502), '智能统一视觉分析模型“gemini-3.5-flash-maxthinking”请求失败：当前无可用凭证');
});

test('identifies the selected text model when smart-unified analysis rejects its credential', async () => {
  await assert.rejects(
    analyzeBatchFactorySmartUnifiedStyle({
      username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret',
      textProvider: { endpoint: 'http://text.local/v1/chat/completions', apiKey: 'key', model: 'gpt-5.4', displayName: 'GPT-5.4' },
      fetchImpl: async (_url, init = {}) => {
        if (init.method === 'GET') return new Response(JSON.stringify({ batch: { id: 'batch-1', settingsState: { patch: { aiPromptConfig: { constraints: { selections: [{ presetId: 'script-constraint-prefix-smart-unified', constraintCategory: 'prefix' }] } } } }, books: [{ id: 'book-1', sourceText: '完整原文', assetRecords: [] }] } }), { status: 200 });
        return new Response(JSON.stringify({ error: { message: '当前无可用凭证' } }), { status: 401 });
      }
    }),
    /智能统一视觉分析模型“GPT-5\.4”请求失败：当前无可用凭证/
  );
});

function makePreset(id, module, kind, protocolLock, body) {
  return { id, module, kind, name: id, version: 3, protocolLock, body };
}

function presetStore() {
  const records = new Map([
    ['script-extract', makePreset('script-extract', 'script', 'base', { format: 'extract', slot: 'script.extract' }, '提取人物与场景')],
    ['script-extract-assets', makePreset('script-extract-assets', 'script', 'base', { format: 'extract', slot: 'script.asset-extraction' }, '提取人物、场景与关键道具')],
    ['script-extract-assets-female', makePreset('script-extract-assets-female', 'script', 'base', { format: 'extract', slot: 'script.asset-extraction' }, '女频人物、场景与关键道具')],
    ['batch-assets-sd-character-scene', makePreset('batch-assets-sd-character-scene', 'script', 'base', { format: 'extract', slot: 'script.asset-extraction' }, 'SD 一次资产调用')],
    ['script-constraint-wrapper', makePreset('script-constraint-wrapper', 'script', 'base', { slot: 'script.constraint.wrapper' }, '约束总规则')],
    ['batch-hook-adaptation', makePreset('batch-hook-adaptation', 'batch-factory', 'base', { slot: 'batch.hook-adaptation' }, '爆款开头规则')],
    ['batch-original-director', makePreset('batch-original-director', 'batch-factory', 'base', { slot: 'batch.original-director' }, '原文导演规则')],
    ['batch-viral-director', makePreset('batch-viral-director', 'batch-factory', 'base', { slot: 'batch.viral-director' }, '爆款导演规则')],
    ['batch-video-meta', makePreset('batch-video-meta', 'batch-factory', 'base', { slot: 'batch.video-meta' }, '完整分镜元提示词')],
    ['batch-video-custom', makePreset('batch-video-custom', 'batch-factory', 'base', { slot: 'batch.video-meta' }, '自定义视频提示词')],
    ['script-segmented', makePreset('script-segmented', 'script', 'base', { slot: 'script.segmented' }, '分段开头规则')],
    ['script-general', makePreset('script-general', 'script', 'base', { slot: 'script.general' }, '通用规则')],
    ['script-character-focus', makePreset('script-character-focus', 'script', 'base', { slot: 'script.character-focus' }, '星标人物聚焦规则')],
    ['script-audio-match', makePreset('script-audio-match', 'script', 'base', { slot: 'script.audio-match' }, '匹配音频规则')],
    ['script-card-protocol', makePreset('script-card-protocol', 'script', 'base', { slot: 'script.card.protocol' }, '统一外层分镜卡片协议')],
    ['script-format-shotlist', makePreset('script-format-shotlist', 'script', 'base', { slot: 'script.format.shotlist' }, '分镜模式规则')],
    ['batch-prefix-modern-conflict', makePreset('batch-prefix-modern-conflict', 'batch-factory', 'base', { slot: 'batch.prefix', key: 'modern_conflict', format: 'video-prefix' }, '现代冲突前缀')],
    ['batch-character-meta', makePreset('batch-character-meta', 'batch-factory', 'base', { slot: 'batch.character-meta' }, '只生成人物提示词')],
    ['batch-scene-meta', makePreset('batch-scene-meta', 'batch-factory', 'base', { slot: 'batch.scene-meta' }, '只生成场景提示词')],
    ['script-constraint-quality-4k', makePreset('script-constraint-quality-4k', 'script', 'addon', { format: 'constraint', slot: 'script.constraint.quality' }, '4K 约束')],
    ['script-constraint-prefix-smart-unified', makePreset('script-constraint-prefix-smart-unified', 'script', 'addon', { format: 'constraint', slot: 'script.constraint.prefix' }, '智能统一元提示词')]
  ]);
  return { getPublished: id => records.get(id) || null, listAll: () => [] };
}

test('enrichment rejects a non-script extraction preset for the unified asset rule', () => {
  const input = { patch: { aiPromptConfig: { assets: { extraction: { presetId: 'batch-scene-meta' } } } } };
  assert.throws(
    () => enrichBatchFactorySystemPresetConfig(input, presetStore()),
    /SD 资产预设提示词/
  );
});

test('derivative-opening selection accepts only its three published prompt slots', () => {
  const selected = resolveDerivedOpeningPrompt('batch-viral-director', presetStore());
  assert.equal(selected.presetId, 'batch-viral-director');
  assert.equal(selected.body, '爆款导演规则');
  assert.throws(() => resolveDerivedOpeningPrompt('batch-video-meta', presetStore()), /衍生开篇/);
});

test('final-prompt previews never refresh or write preset snapshots', () => {
  const result = presetDrivenExecutionPath(
    { method: 'GET' },
    '/api/batch-factory/v11/batches/batch-1/books/book-1/videos/video-1/final-prompt'
  );
  assert.equal(result, null);
});

test('native V12 H3 compilation refreshes the selected prompt preset snapshot', () => {
  const result = presetDrivenExecutionPath(
    { method: 'POST' },
    '/api/batch-factory/v12/batches/batch-1/books/book-1/h3/compile'
  );
  assert.deepEqual(result, { batchId: 'batch-1', bookId: 'book-1' });
});

test('batch status reads never require a personal video provider sync', () => {
  const request = { method: 'GET' };
  const statusPath = '/api/batch-factory/v11/batches/batch-1/status';
  assert.equal(needsPersonalConfigSync(request, statusPath), false);
  assert.equal(needsH3ConfigSync(request, statusPath), false);
});

test('single-book VIDEO stages synchronize the selected provider before submitting work', () => {
  const request = { method: 'POST' };
  const stagePath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/video';
  assert.equal(needsPersonalConfigSync(request, stagePath), true);
  assert.equal(needsH3ConfigSync(request, stagePath), true);
});

test('retrying a non-video stage never requires a video provider sync', () => {
  const request = { method: 'POST' };
  const retryPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/retry';
  assert.equal(needsPersonalConfigSync(request, retryPath), false);
  assert.equal(needsH3ConfigSync(request, retryPath), false);
});

test('retrying an asynchronously failed Seedance storyboard synchronizes its provider', () => {
  const request = { method: 'POST', body: { provider: 'yfai_seedance', videoModelId: 'seedance-2-0-official' } };
  const retryPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/retry';
  assert.equal(needsPersonalConfigSync(request, retryPath), true);
  assert.equal(needsH3ConfigSync(request, retryPath), false);
});

test('retrying an asynchronously failed H3 storyboard synchronizes its provider', () => {
  const request = { method: 'POST', body: { provider: 'autodl_comfyui', videoModelId: 'minimax-h3-video' } };
  const retryPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/retry';
  assert.equal(needsPersonalConfigSync(request, retryPath), false);
  assert.equal(needsH3ConfigSync(request, retryPath), true);
});

test('automatic video retry uses the VIDEO synchronization path', () => {
  const request = { method: 'POST' };
  const videoPath = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/video';
  assert.equal(needsPersonalConfigSync(request, videoPath), true);
});

test('enrichment snapshots the only combined script extraction preset and strips browser supplied bodies', () => {
  const enriched = enrichBatchFactorySystemPresetConfig({
    patch: {
      aiPromptConfig: {
        assets: {
          extraction: { presetId: 'script-extract-assets', body: '浏览器伪造正文', prompt: '旧正文' }
        }
      }
    }
  }, presetStore());
  const { extraction, character } = enriched.patch.aiPromptConfig.assets;
  assert.equal(extraction.body, '提取人物、场景与关键道具');
  assert.equal(extraction.prompt, undefined);
  assert.equal(character, undefined);
  assert.equal(redactBatchFactorySystemPromptBodies(enriched).patch.aiPromptConfig.assets.extraction.body, undefined);
});

test('enrichment rejects the character-and-scene-only script extractor for Batch Factory assets', () => {
  assert.throws(
    () => enrichBatchFactorySystemPresetConfig({ patch: { aiPromptConfig: { assets: { extraction: { presetId: 'script-extract' } } } } }, presetStore()),
    /SD 资产预设提示词/
  );
});

test('enrichment defaults old batches to the one-call SD asset preset', () => {
  const enriched = enrichBatchFactorySystemPresetConfig({
    patch: { aiPromptConfig: { assets: { enabled: true } } }
  }, presetStore());
  assert.deepEqual(enriched.patch.aiPromptConfig.assets.extraction, {
    presetId: 'batch-assets-sd-character-scene',
    presetName: 'batch-assets-sd-character-scene',
    presetSlot: 'script.asset-extraction',
    presetVersion: 3,
    body: 'SD 一次资产调用'
  });
});

test('migrates a legacy single-book smart-unified prefix into the typed constraint selection', () => {
  const legacyBook = { settingsState: { patch: { aiPromptConfig: { constraints: { enabled: true, prefix: { enabled: true, presetId: 'script-constraint-prefix-smart-unified' } } } } } };
  assert.equal(smartUnifiedSelected({ settingsState: { patch: {} } }, legacyBook), true);
  const enriched = enrichBatchFactorySystemPresetConfig({ patch: legacyBook.settingsState.patch }, presetStore());
  const constraints = enriched.patch.aiPromptConfig.constraints;
  assert.equal(constraints.prefix, undefined);
  assert.deepEqual(constraints.selections.map(item => [item.presetId, item.constraintCategory]), [['script-constraint-prefix-smart-unified', 'prefix']]);
  assert.equal(constraints.selections[0].body, '智能统一元提示词');
});

test('H3 runs visual-baseline analysis without enabling smart-unified display', () => {
  const batch = { settingsState: { patch: { aiPromptConfig: {
    video: { enabled: true, presetId: 'batch-video-h3-director', presetKey: 'h3-video-normal' }
  } } } };
  const book = { sourceText: '完整视频原文', settingsState: { patch: {} } };
  assert.equal(smartUnifiedSelected(batch, book), false);
	assert.equal(directorVisualBaselineRequired(batch, book), true);
});

test('enrichment snapshots the ordinary public script composition for every V11 book', () => {
  const records = new Map([
    ['script-segmented', makePreset('script-segmented', 'script', 'base', { slot: 'script.segmented' }, 'SEGMENTED {duration}')],
    ['script-format-shotlist', makePreset('script-format-shotlist', 'script', 'base', { slot: 'script.format.shotlist' }, 'SHOTLIST {duration}')],
    ['script-general', makePreset('script-general', 'script', 'base', { slot: 'script.general' }, 'GENERAL {duration}')],
    ['script-character-focus', makePreset('script-character-focus', 'script', 'base', { slot: 'script.character-focus' }, 'FOCUS {focusCharacters}/{focusCount}')],
    ['script-audio-match', makePreset('script-audio-match', 'script', 'base', { slot: 'script.audio-match' }, 'AUDIO {audioDurationSec}/{unitMaxSec}')],
    ['script-card-protocol', makePreset('script-card-protocol', 'script', 'base', { slot: 'script.card.protocol' }, 'CARD {duration}')],
    ['script-constraint-wrapper', makePreset('script-constraint-wrapper', 'script', 'base', { slot: 'script.constraint.wrapper' }, 'WRAPPER {duration}')]
  ]);
  const store = { getPublished: id => records.get(id) || null, listAll: () => [...records.values()] };
  const enriched = enrichBatchFactorySystemPresetConfig({ patch: { aiPromptConfig: {} } }, store);
  const composition = enriched.patch.aiPromptConfig.scriptComposition;
  assert.deepEqual(
    Object.fromEntries(Object.entries(composition).map(([key, value]) => [key, value.presetId])),
    {
      segmented: 'script-segmented', shotlist: 'script-format-shotlist', general: 'script-general',
      characterFocus: 'script-character-focus', audioMatch: 'script-audio-match',
      cardProtocol: 'script-card-protocol', constraintWrapper: 'script-constraint-wrapper'
    }
  );
  assert.equal(composition.segmented.body, 'SEGMENTED {duration}');
  assert.equal(composition.cardProtocol.body, 'CARD {duration}');
});

test('enrichment resolves script constraints as typed selections', () => {
  const enriched = enrichBatchFactorySystemPresetConfig({
    patch: {
      aiPromptConfig: {
        constraints: { enabled: true, selections: [{ presetId: 'script-constraint-quality-4k' }] }
      }
    }
  }, presetStore());
  assert.deepEqual(enriched.patch.aiPromptConfig.constraints.selections[0], {
    presetId: 'script-constraint-quality-4k',
    presetName: 'script-constraint-quality-4k',
    presetSlot: 'script.constraint.quality',
    presetVersion: 3,
    constraintCategory: 'quality',
    body: '4K 约束'
  });
});

test('enrichment keeps the selected published video prompt as one option in the Batch Factory video prompt selector', () => {
  const enriched = enrichBatchFactorySystemPresetConfig({
    patch: {
      aiPromptConfig: {
        video: { enabled: true, presetId: 'batch-video-custom', body: '浏览器伪造的旧视频规则' }
      }
    }
  }, presetStore());
  const video = enriched.patch.aiPromptConfig.video;
  assert.equal(video.presetId, 'batch-video-custom');
  assert.equal(video.presetName, 'batch-video-custom');
  assert.equal(video.body, '自定义视频提示词');
  assert.equal(video.sourcePresets, undefined);
});

test('enrichment drops retired character and scene renderer selections', () => {
  const enriched = enrichBatchFactorySystemPresetConfig({
    patch: {
      aiPromptConfig: {
        assets: { extraction: { presetId: 'script-extract-assets-female' }, character: { presetId: 'batch-character-meta' } },
        constraints: { wrapper: { presetId: 'script-constraint-wrapper' } },
        hook: { presetId: 'batch-hook-adaptation' },
        originalDirector: { presetId: 'batch-original-director' },
        viralDirector: { presetId: 'batch-viral-director' },
        prefix: { presetId: 'batch-prefix-modern-conflict' }
      }
    }
  }, presetStore());
  const config = enriched.patch.aiPromptConfig;
  assert.equal(config.assets.extraction.body, '女频人物、场景与关键道具');
	assert.equal(config.assets.character, undefined);
	assert.equal(config.assets.scene, undefined);
  assert.equal(config.constraints.wrapper, undefined);
  assert.equal(config.hook.body, '爆款开头规则');
  assert.equal(config.originalDirector.body, '原文导演规则');
  assert.equal(config.viralDirector.body, '爆款导演规则');
  assert.equal(config.prefix, undefined);
});

test('single-book director target matches only POST stages/director for one book', () => {
  const path = '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/director';
  assert.deepEqual(singleBookDirectorStageTarget({ method: 'POST' }, path), { batchId: 'batch-1', bookId: 'book-1' });
  assert.equal(singleBookDirectorStageTarget({ method: 'GET' }, path), null);
  assert.equal(singleBookDirectorStageTarget({ method: 'POST' }, '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/opening'), null);
  assert.equal(singleBookDirectorStageTarget({ method: 'POST' }, '/api/batch-factory/v11/batches/batch-1/books/book-1/stages/director/extra'), null);
});

function singleDirectorOptions({ openingEnabled, openingCount = 3, metaBody, fetchImpl }) {
  const batch = {
    id: 'batch-1',
    settingsState: { patch: { openingEnabled, openingCount, textModelId: 'text-1' } },
    books: [{
      id: 'book-1',
      settingsState: { patch: {} },
      videos: [
        { id: 'v0', settingsState: { patch: {} } },
        { id: 'v1', settingsState: { patch: {} } }
      ]
    }]
  };
  return {
    req: { username: 'alice', auth: { account: { isOwner: true } } },
    target: { batchId: 'batch-1', bookId: 'book-1' },
    options: {
      goBaseUrl: 'http://go.local',
      bridgeSecret: 'secret',
      presetStore: { getPublished: id => (id === 'batch-opening-meta' && metaBody ? { id, body: metaBody } : null), listAll: () => [] },
      memberStore: { getMember: username => ({ username, active: true, role: 'manager' }), canUseApi: () => true },
      configReader: () => ({ modelCatalog: [{ id: 'text-1', kind: 'text', enabled: true, baseUrl: 'https://text.example/v1', modelId: 'gpt-x', credential: 'key', displayName: '文本X' }] }),
      fetchImpl
    },
    batch
  };
}

test('after single-book director, opening variants are generated with the meta preset and text provider', async () => {
  const posts = [];
  let batchGetCount = 0;
  const ctx = singleDirectorOptions({
    openingEnabled: true,
    metaBody: '换开头元规则正文',
    fetchImpl: async (url, init = {}) => {
      if (init.method === 'POST' && url.endsWith('/stages/opening')) {
        posts.push({ url, body: JSON.parse(init.body) });
        return new Response(JSON.stringify({ runs: [{ stage: 'opening', status: 'succeeded' }] }), { status: 201 });
      }
      if (url.endsWith('/batches/batch-1') && (!init.method || init.method === 'GET')) {
        batchGetCount += 1;
        if (batchGetCount === 1) return new Response(JSON.stringify({ batch: ctx.batch }), { status: 200 });
        const withVariants = JSON.parse(JSON.stringify(ctx.batch));
        withVariants.books[0].videos[0].settingsState.patch.openingVariants = [
          { index: 1, status: 'success', prompt: '变体一：茶盏碎裂。' },
          { index: 2, status: 'success', prompt: '变体二：雨夜推门。' }
        ];
        return new Response(JSON.stringify({ batch: withVariants }), { status: 200 });
      }
      throw new Error(`unexpected ${init.method || 'GET'} ${url}`);
    }
  });
  const result = await generateOpeningVariantsAfterSingleDirector(ctx.req, ctx.target, ctx.options);
  assert.deepEqual(result, { triggered: true, generated: 2, succeeded: true, reason: '' });
  assert.equal(posts.length, 1);
  assert.match(posts[0].url, /books\/book-1\/stages\/opening$/);
  assert.equal(posts[0].body.mode, 'force');
  assert.equal(posts[0].body.openingMeta.presetId, 'batch-opening-meta');
  assert.equal(posts[0].body.openingMeta.body, '换开头元规则正文');
  assert.equal(posts[0].body.textProvider.model, 'gpt-x');
});

test('single-book director reports opening failure when a required variant is missing', async () => {
  let batchGetCount = 0;
  const ctx = singleDirectorOptions({
    openingEnabled: true,
    openingCount: 3,
    metaBody: '换开头元规则正文',
    fetchImpl: async (url, init = {}) => {
      if (init.method === 'POST' && url.endsWith('/stages/opening')) {
        return new Response(JSON.stringify({ runs: [{ stage: 'opening', status: 'failed' }] }), { status: 422 });
      }
      if (url.endsWith('/batches/batch-1') && (!init.method || init.method === 'GET')) {
        batchGetCount += 1;
        if (batchGetCount === 1) return new Response(JSON.stringify({ batch: ctx.batch }), { status: 200 });
        const withFailure = JSON.parse(JSON.stringify(ctx.batch));
        withFailure.books[0].videos[0].settingsState.patch.openingVariants = [
          { index: 1, status: 'success', prompt: '变体一：茶盏碎裂。' },
          { index: 2, status: 'failed', prompt: '', failureReason: '模型未输出该变体分段' }
        ];
        return new Response(JSON.stringify({ batch: withFailure }), { status: 200 });
      }
      throw new Error(`unexpected ${init.method || 'GET'} ${url}`);
    }
  });
  const result = await generateOpeningVariantsAfterSingleDirector(ctx.req, ctx.target, ctx.options);
  assert.equal(result.triggered, true);
  assert.equal(result.generated, 1);
  assert.equal(result.succeeded, false);
  assert.match(result.reason, /换开头/);
});

test('single-book director does not call opening when the switch is off', async () => {
  const posts = [];
  const ctx = singleDirectorOptions({
    openingEnabled: false,
    metaBody: '元规则',
    fetchImpl: async (url, init = {}) => {
      if (init.method === 'POST' && url.endsWith('/stages/opening')) {
        posts.push(url);
        return new Response(JSON.stringify({ runs: [] }), { status: 201 });
      }
      return new Response(JSON.stringify({ batch: ctx.batch }), { status: 200 });
    }
  });
  const result = await generateOpeningVariantsAfterSingleDirector(ctx.req, ctx.target, ctx.options);
  assert.equal(posts.length, 0);
  assert.equal(result.triggered, false);
});

test('execution refreshes the selected preset snapshot to the latest published name, version and body', async () => {
  const calls = [];
  const store = presetStore();
  const latest = store.getPublished('script-extract-assets');
  latest.name = '统一资产提取（最新）';
  latest.version = 4;
  latest.body = '最新资产提取规则';
  const result = await refreshBatchFactoryPresetSnapshot({
    username: 'alice', isOwner: false, batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', presetStore: store,
    fetchImpl: async (url, init) => {
      calls.push({ url, init });
      const batch = { id: 'batch-1', revision: 7, settingsState: { patch: { aiPromptConfig: { assets: { extraction: { presetId: 'script-extract-assets', presetName: '旧名称', presetVersion: 3, body: '旧正文' } } } } }, books: [{ id: 'book-1', revision: 5, settingsState: { patch: {} } }] };
      return new Response(JSON.stringify({ batch }), { status: 200, headers: { 'content-type': 'application/json' } });
    }
  });
  assert.equal(result.batchRefreshed, true);
  assert.equal(calls.length, 2);
  const saved = JSON.parse(calls[1].init.body);
  assert.equal(saved.patch.aiPromptConfig.assets.extraction.presetName, '统一资产提取（最新）');
  assert.equal(saved.patch.aiPromptConfig.assets.extraction.presetVersion, 4);
  assert.equal(saved.patch.aiPromptConfig.assets.extraction.body, '最新资产提取规则');
});

test('execution refreshes a single-book prompt override without touching another book', async () => {
  const calls = [];
  const store = presetStore();
  const latest = store.getPublished('script-extract-assets');
  latest.name = '单书资产提取（最新）'; latest.version = 4; latest.body = '单书最新规则';
  const result = await refreshBatchFactoryPresetSnapshot({
    username: 'alice', batchId: 'batch-1', bookId: 'book-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', presetStore: store,
    fetchImpl: async (url, init) => {
      calls.push({ url, init });
      const batch = { id: 'batch-1', revision: 7, settingsState: { patch: {} }, books: [
        { id: 'book-1', revision: 5, settingsState: { patch: { aiPromptConfig: { assets: { extraction: { presetId: 'script-extract-assets', body: '旧正文' } } } } } },
        { id: 'book-2', revision: 6, settingsState: { patch: { aiPromptConfig: { assets: { extraction: { presetId: 'script-extract-assets', body: '另一书旧正文' } } } } } }
      ] };
      return new Response(JSON.stringify({ batch }), { status: 200, headers: { 'content-type': 'application/json' } });
    }
  });
  assert.deepEqual(result, { batchRefreshed: true, bookRefreshed: true });
  assert.equal(calls.length, 3);
  assert.match(calls[2].url, /books\/book-1\/override$/);
  assert.equal(JSON.parse(calls[2].init.body).patch.aiPromptConfig.assets.extraction.body, '单书最新规则');
});

test('execution initializes public script composition for a legacy batch without an AI prompt config', async () => {
  const calls = [];
  const result = await refreshBatchFactoryPresetSnapshot({
    username: 'alice', batchId: 'batch-1', goBaseUrl: 'http://go.local', bridgeSecret: 'secret', presetStore: presetStore(),
    fetchImpl: async (url, init) => {
      calls.push({ url, init });
      const batch = { id: 'batch-1', revision: 7, settingsState: { patch: {} }, books: [] };
      return new Response(JSON.stringify({ batch }), { status: 200, headers: { 'content-type': 'application/json' } });
    }
  });
  assert.deepEqual(result, { batchRefreshed: true, bookRefreshed: false });
  assert.equal(calls.length, 2);
  const saved = JSON.parse(calls[1].init.body);
  assert.equal(saved.patch.aiPromptConfig.scriptComposition.segmented.presetId, 'script-segmented');
  assert.equal(saved.patch.aiPromptConfig.scriptComposition.shotlist.presetId, 'script-format-shotlist');
});
