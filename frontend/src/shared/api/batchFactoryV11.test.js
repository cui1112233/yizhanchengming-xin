import test from 'node:test';
import assert from 'node:assert/strict';
import * as batchFactoryV11 from './batchFactoryV11.js';

const { bf11Path, bf11ScopePath } = batchFactoryV11;

test('Batch Factory UI requests the V12 API namespace', () => {
	assert.equal(bf11Path('/capabilities'), '/api/batch-factory/v12/capabilities');
	assert.equal(bf11Path('batches'), '/api/batch-factory/v12/batches');
});

test('automation status always bypasses a stale browser response cache', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ automation: { state: 'idle' } }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.getBatchAutomationStatus('batch 1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/automation');
  assert.equal(calls[0].options.cache, 'no-store');
});

test('runtime summary uses one non-cached batch endpoint', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ batchId: 'batch-1', production: { jobs: [] }, merge: { jobs: [] }, stageSummaries: {} }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.getBatchRuntimeSummary('batch 1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/runtime-summary');
  assert.equal(calls[0].options.cache, 'no-store');
});

test('book deletion uses the persisted production retention window', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    if (path === '/api/config') {
      return new Response(JSON.stringify({ productionRetentionDays: 14 }), { status: 200, headers: { 'content-type': 'application/json' } });
    }
    return new Response(null, { status: 204 });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.deleteBatchFactoryBook('batch 1', 'book/1');

  assert.deepEqual(calls.map(call => call.path), [
    '/api/config',
    '/api/batch-factory/v12/batches/batch%201/books/book%2F1?retentionDays=14'
  ]);
});

test('project deletion uses the persisted production retention window', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    if (path === '/api/config') {
      return new Response(JSON.stringify({ productionRetentionDays: 30 }), { status: 200, headers: { 'content-type': 'application/json' } });
    }
    return new Response(null, { status: 204 });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.deleteBatchFactoryProject('batch 1');

  assert.deepEqual(calls.map(call => call.path), [
    '/api/config',
    '/api/batch-factory/v12/batches/batch%201?retentionDays=30'
  ]);
});

test('deletion stops when the retention setting cannot be read', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ error: '设置读取失败' }), { status: 500, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await assert.rejects(() => batchFactoryV11.deleteBatchFactoryProject('batch 1'), /设置读取失败/);
  assert.deepEqual(calls.map(call => call.path), ['/api/config']);
});

test('scope paths address batch, book and VIDEO overrides without legacy routes', () => {
	assert.equal(bf11ScopePath({ scope: 'batch', batchId: 'b 1' }), '/api/batch-factory/v12/batches/b%201/settings');
	assert.equal(bf11ScopePath({ scope: 'book', batchId: 'b1', bookId: 'k/1' }), '/api/batch-factory/v12/batches/b1/books/k%2F1/override');
	assert.equal(bf11ScopePath({ scope: 'video', batchId: 'b1', bookId: 'k1', videoId: 'v?1' }), '/api/batch-factory/v12/batches/b1/books/k1/videos/v%3F1/override');
});

test('source updates use the book-owned V11 route', () => {
  assert.equal(typeof batchFactoryV11.updateBookSource, 'function');
  assert.match(String(batchFactoryV11.updateBookSource), /books\/\$\{id\(bookId\)\}\/source/);
});

test('comic page client uses book-owned V12 routes', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ pages: [], collections: [], image: {}, collection: {} }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.listComicPages('batch 1', 'book/1');
  await batchFactoryV11.generateComicPage('batch 1', 'book/1', 'video?1', { imagePrompt: '完整漫画页' });
  await batchFactoryV11.generateComicBook('batch 1', 'book/1');
  await batchFactoryV11.setPrimaryComicPageImage('batch 1', 'book/1', 'page#1', 'image/1');
	await batchFactoryV11.setComicPageReferenceImages('batch 1', 'book/1', 'video?1', ['ref-a']);
	await batchFactoryV11.applyComicPageReferenceImages('batch 1', 'book/1', 'video?1');
  await batchFactoryV11.createComicCollection('batch 1', 'book/1');
  await batchFactoryV11.listComicCollections('batch 1', 'book/1');
	await batchFactoryV11.getComicCollectionPublishRecord('batch 1', 'book/1', 'collection#1');
	await batchFactoryV11.getN8ComicPublishStatus();
	await batchFactoryV11.searchN8ComicWorks('小妾造反后');
	await batchFactoryV11.publishComicCollection('batch 1', 'book/1', 'collection#1', { producerAdminName: '崔家铭', audioCategoryId: 3, workIds: [2001] });

  assert.deepEqual(calls.map(call => [call.path, call.options.method || 'GET']), [
    ['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-pages', 'GET'],
    ['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-pages/video%3F1/generate', 'POST'],
    ['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-collections/generate', 'POST'],
    ['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-pages/page%231/images/image%2F1/primary', 'PUT'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-pages/video%3F1/reference-images', 'PUT'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-pages/video%3F1/reference-images/apply-to-all', 'POST'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-collections', 'POST'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-collections', 'GET'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-collections/collection%231/publish', 'GET'],
		['/api/batch-factory/v12/comic-publish/status', 'GET'],
		['/api/batch-factory/v12/comic-publish/works?query=%E5%B0%8F%E5%A6%BE%E9%80%A0%E5%8F%8D%E5%90%8E', 'GET'],
		['/api/batch-factory/v12/batches/batch%201/books/book%2F1/comic-collections/collection%231/publish', 'POST']
	]);
  assert.deepEqual(JSON.parse(calls[1].options.body), { imagePrompt: '完整漫画页' });
  assert.deepEqual(JSON.parse(calls[4].options.body), { referenceImageUrls: ['ref-a'] });
	assert.deepEqual(JSON.parse(calls[11].options.body), { producerAdminName: '崔家铭', audioCategoryId: 3, workIds: [2001] });
});

test('director prompt candidate requests use the public V12 HTTP contract', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ candidate: { id: 'candidate-1' }, video: { id: 'video?1' }, cancelled: true }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.createDirectorPromptCandidate('batch 1', 'book/1', '7');
  await batchFactoryV11.getDirectorPromptCandidate('batch 1', 'book/1');
  await batchFactoryV11.acceptDirectorPromptCandidate('batch 1', 'book/1', 'video?1', 'candidate-1', '8');
  await batchFactoryV11.cancelDirectorPromptCandidate('batch 1', 'book/1', '9');

  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/director-prompt-candidate');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), { expectedRevision: 7 });
  assert.equal(calls[1].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/director-prompt-candidate');
  assert.equal(calls[1].options.method || 'GET', 'GET');
  assert.equal(calls[2].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/videos/video%3F1/director-prompt-candidate/accept');
  assert.equal(calls[2].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[2].options.body), { candidateId: 'candidate-1', expectedRevision: 8 });
  assert.equal(calls[3].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/director-prompt-candidate');
  assert.equal(calls[3].options.method, 'DELETE');
  assert.deepEqual(JSON.parse(calls[3].options.body), { expectedRevision: 9 });
});

test('per-book merge client never falls back to legacy batch merge routes', () => {
  assert.equal(typeof batchFactoryV11.submitBookMerge, 'function');
  assert.equal(typeof batchFactoryV11.getBookMergeStatus, 'function');
  assert.match(String(batchFactoryV11.submitBookMerge), /books\/\$\{id\(bookId\)\}\/merge/);
  assert.match(String(batchFactoryV11.getBookMergeStatus), /books\/\$\{id\(bookId\)\}\/merge-status/);
  assert.match(String(batchFactoryV11.getBookMergeStatus), /requestId/);
});

test('unsupported scope fails instead of falling back to a legacy API', () => {
  assert.throws(() => bf11ScopePath({ scope: 'project', batchId: 'b1' }), /Unsupported batch factory scope/);
});

test('submitBookMerge sends the complete payload to the encoded V12 book merge route', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options, headers: new Headers(options.headers || {}) });
    return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  const payload = { requestId: 'req-1', timingMode: 'speed', speed: 1.2, ttsSpeed: 1.7, audioDurationSeconds: 10 };
  await batchFactoryV11.submitBookMerge('batch 1', 'book/1', payload);

  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/merge');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), payload);
  assert.match(calls[0].headers.get('authorization') || '', /^Bearer\s+contract-test-token$/);
});

test('single-book production sends the selected video model in its V12 request', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.submitBookProduction('batch 1', 'book/1', 'request-1', 'p', 'model-1');

  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/production');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), {
    requestId: 'request-1',
    provider: 'p',
    videoModelId: 'model-1'
  });
});

test('batch production keeps its selected video model in the V12 request', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.submitBatchProduction('batch 1', 'request-1', 'p', 'model-1');

  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/production');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), {
    requestId: 'request-1',
    provider: 'p',
    videoModelId: 'model-1'
  });
});

test('batch factory direct fetch uses its native V12 endpoint', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ results: [] }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.fetchDirectOriginals({ platform: '15', bookIds: ['2084012035524801698'], maxTxt: 4000 });
  assert.match(calls[0].path, /^\/api\/batch-factory\/v12\/fetch-originals\?/);
  const requestURL = new URL(calls[0].path, 'http://qiantie.local');
  assert.equal(requestURL.searchParams.get('platform'), '15');
  assert.equal(requestURL.searchParams.get('maxTxt'), '4000');
  assert.deepEqual(requestURL.searchParams.getAll('bookId'), ['2084012035524801698']);
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), {
    platform: 15,
    bookIds: ['2084012035524801698'],
    maxTxt: 4000
  });
});
test('batch factory metadata backfill uses the bounded V12 batch endpoint', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ results: [] }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.backfillBatch121Metadata('batch 1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/backfill-121-metadata');
  assert.equal(calls[0].options.method, 'POST');
});
test('a legacy empty book can fetch and persist its own original through V12', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ book: { sourceText: '正文' } }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.fetchBookOriginal('batch 1', 'book/1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/fetch-original');
  assert.equal(calls[0].options.method, 'POST');
});

test('book audio planning uses the V12 book-owned route without creating a merge', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ audioArtifactId: 'tts-1', durationSeconds: 7.5 }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.createBookAudioPlan('batch 1', 'book/1', { requestId: 'canvas-audio-1' });

  assert.equal(calls.length, 1);
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/audio-plan');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), { requestId: 'canvas-audio-1' });
});

test('single-book fetch forwards the selected content limit and platform', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ book: { sourceText: '正文' } }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.fetchBookOriginal('batch 1', 'book/1', { maxTxt: 8000, platform: '15' });
  assert.deepEqual(JSON.parse(calls[0].options.body), { maxTxt: 8000, platform: '15' });
});

test('a frozen Novel Fetch intake joins the selected V12 batch without a duplicate override', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ batch: { id: 'batch 1', books: [] } }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.appendNovelFetchIntake('batch 1', 'intake/1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/intakes/intake%2F1/books');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[0].options.body), { allowDuplicate: false });
});

test('protected local executor media is fetched as an authenticated blob', async t => {
  assert.equal(typeof batchFactoryV11.getProductionMediaBlob, 'function');

  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = {
    getItem(key) { return key === 'auth_token' ? 'contract-test-token' : ''; },
    setItem() {},
    removeItem() {}
  };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, headers: new Headers(options.headers || {}) });
    return new Response(new Blob(['contract-video'], { type: 'video/mp4' }), {
      status: 200,
      headers: { 'content-type': 'video/mp4' }
    });
  };
  t.after(() => {
    globalThis.localStorage = originalLocalStorage;
    globalThis.fetch = originalFetch;
  });

  const blob = await batchFactoryV11.getProductionMediaBlob('/api/shuihuo-production/local-executor-artifacts/lea_1');
  assert.equal(blob.type, 'video/mp4');
  assert.equal(calls.length, 1);
  assert.equal(calls[0].path, '/api/shuihuo-production/local-executor-artifacts/lea_1');
  assert.match(calls[0].headers.get('authorization') || '', /^Bearer\s+/);
});

test('protected local merge media is fetched as an authenticated blob', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem(key) { return key === 'auth_token' ? 'contract-test-token' : ''; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, headers: new Headers(options.headers || {}) });
    return new Response(new Blob(['contract-video'], { type: 'video/mp4' }), { status: 200, headers: { 'content-type': 'video/mp4' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  const path = '/api/batch-factory/v11/batches/batch_1/merge-media/merge_local_1';
  const blob = await batchFactoryV11.getProductionMediaBlob(path);
  assert.equal(blob.type, 'video/mp4');
  assert.equal(calls[0].path, path);
  assert.match(calls[0].headers.get('authorization') || '', /^Bearer\s+/);
});

test('private merge covers are fetched as authenticated JPG blobs rather than direct image URLs', async t => {
  assert.equal(typeof batchFactoryV11.getBatchFactoryMergeCoverBlob, 'function');

  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem(key) { return key === 'auth_token' ? 'contract-test-token' : ''; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, headers: new Headers(options.headers || {}), options });
    return new Response(new Blob(['contract-jpg'], { type: 'image/jpeg' }), { status: 200, headers: { 'content-type': 'image/jpeg' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  const path = '/api/batch-factory/v11/batches/batch_1/merge-cover/merge_local_1';
  const blob = await batchFactoryV11.getBatchFactoryMergeCoverBlob(path);
  assert.equal(blob.type, 'image/jpeg');
  assert.equal(calls[0].path, path);
  assert.match(calls[0].headers.get('authorization') || '', /^Bearer\s+/);
  assert.equal(calls[0].options.silent, true);
});

test('merge-cover blob helper refuses external URLs so the app bearer token is never forwarded off-origin', async () => {
  await assert.rejects(
    () => batchFactoryV11.getBatchFactoryMergeCoverBlob('https://media.example/cover.jpg'),
    /merge cover/i
  );
});

test('private book asset images are fetched as authenticated blobs instead of native image requests', async t => {
  assert.equal(typeof batchFactoryV11.getBookAssetImageBlob, 'function');

  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem(key) { return key === 'auth_token' ? 'contract-test-token' : ''; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, headers: new Headers(options.headers || {}), options });
    return new Response(new Blob(['contract-png'], { type: 'image/png' }), { status: 200, headers: { 'content-type': 'image/png' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  const path = '/api/batch-factory/v12/batches/batch_1/books/book_1/assets/asset_1/images/image_1/content';
  const blob = await batchFactoryV11.getBookAssetImageBlob(path);
  assert.equal(blob.type, 'image/png');
  assert.equal(calls[0].path, path);
  assert.equal(calls[0].headers.get('authorization'), 'Bearer contract-test-token');
  assert.equal(calls[0].options.silent, true);
});

test('book asset image blob helper refuses external URLs so the app bearer token cannot leak', async () => {
  await assert.rejects(
    () => batchFactoryV11.getBookAssetImageBlob('https://media.example/private.png'),
    /book asset image/i
  );
});

test('production media blob helper refuses external URLs so the app bearer token is never forwarded off-origin', async () => {
  assert.equal(typeof batchFactoryV11.getProductionMediaBlob, 'function');
  await assert.rejects(
    () => batchFactoryV11.getProductionMediaBlob('https://media.example/video.mp4'),
    /local executor artifact/i
  );
});

test('H3 trace reads the native V12 route and keeps the compilation identity', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ legacy: false, compilation: { id: 'compile-1' } }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  const result = await batchFactoryV11.getH3Trace('batch 1', 'book/1', 'compile-1');
  assert.equal(result.compilation.id, 'compile-1');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/h3/trace?compilationId=compile-1');
});

test('H3 audio measurement sends bytes without accepting a client duration', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ audio_asset_id: 'h3-audio-1' }), { status: 201, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.measureH3Audio('batch 1', 'book/1', 'YXVkaW8=');
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/h3/audio-measurement');
  assert.deepEqual(JSON.parse(calls[0].options.body), { audio_base64: 'YXVkaW8=' });
  const lines = { director_revision_id:'d1', tts_fingerprint:'voice', lines:[{source_key:'line_0001',source_text:'正文',audio_base64:'YXVkaW8='}] };
  await batchFactoryV11.measureH3Audio('b','k',lines);
  assert.deepEqual(JSON.parse(calls[1].options.body), lines);
});

test('uploads a book VIDEO MP4 as authenticated multipart data through V12', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  const calls = [];
  globalThis.localStorage = { getItem() { return 'contract-test-token'; }, setItem() {}, removeItem() {} };
  globalThis.fetch = async (path, options = {}) => {
    calls.push({ path, options });
    return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  t.after(() => { globalThis.localStorage = originalLocalStorage; globalThis.fetch = originalFetch; });

  await batchFactoryV11.uploadBookVideoMaterial('batch 1', 'book/1', 'video?1', new Blob(['mp4'], { type: 'video/mp4' }));
  assert.equal(calls[0].path, '/api/batch-factory/v12/batches/batch%201/books/book%2F1/videos/video%3F1/upload');
  assert.equal(calls[0].options.method, 'POST');
  assert.ok(calls[0].options.body instanceof FormData);
  assert.equal(calls[0].options.body.get('file').type, 'video/mp4');
  const headers = new Headers(calls[0].options.headers);
  assert.match(headers.get('authorization') || '', /^Bearer\s+contract-test-token$/);
  assert.equal(headers.has('content-type'), false);
});
