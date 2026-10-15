const test = require('node:test');
const assert = require('node:assert/strict');
const { ExecutorApiClient, ApiError } = require('../src/api-client');

function fakeFetch(responses, calls) {
  return async (url, options) => {
    calls.push({ url, options });
    const next = responses.shift();
    return {
      status: next.status,
      ok: next.status >= 200 && next.status < 300,
      async text() { return next.body === undefined ? '' : JSON.stringify(next.body); }
    };
  };
}

test('uses public device paths and bearer headers', async () => {
  const calls = [];
  const client = new ExecutorApiClient({ baseUrl: 'https://v78.example.com/', fetchImpl: fakeFetch([{ status: 200, body: { ok: true } }], calls) });
  await client.heartbeat('device-token', { accounts: { total: 0 } });
  assert.equal(calls[0].url, 'https://v78.example.com/api/local-executor/v1/heartbeat');
  assert.equal(calls[0].options.headers.Authorization, 'Bearer device-token');
});

test('claim returns null on 204', async () => {
  const calls = [];
  const client = new ExecutorApiClient({ baseUrl: 'https://v78.example.com', fetchImpl: fakeFetch([{ status: 204 }], calls) });
  assert.equal(await client.claim('token'), null);
});

test('conflict is surfaced as ApiError', async () => {
  const client = new ExecutorApiClient({ baseUrl: 'https://v78.example.com', fetchImpl: fakeFetch([{ status: 409, body: { error: 'stale lease' } }], []) });
  await assert.rejects(() => client.progress('token', 'j1', { token: 'l', generation: 1 }, 'preparing'), err => err instanceof ApiError && err.status === 409);
});
