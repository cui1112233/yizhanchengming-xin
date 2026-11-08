import assert from 'node:assert/strict';
import test from 'node:test';

import { listComicPages } from './batchFactoryV12.js';

test('comic page status reads bypass the browser cache after image generation', async () => {
  const previousFetch = globalThis.fetch;
  const previousStorage = globalThis.localStorage;
  let requestCache;
  globalThis.localStorage = { getItem() { return ''; } };
  globalThis.fetch = async (_path, init) => {
    requestCache = init.cache;
    return new Response(JSON.stringify({ pages: [] }), {
      status: 200,
      headers: { 'content-type': 'application/json' }
    });
  };
  try {
    await listComicPages('batch-1', 'book-1');
    assert.equal(requestCache, 'no-store');
  } finally {
    globalThis.fetch = previousFetch;
    globalThis.localStorage = previousStorage;
  }
});
