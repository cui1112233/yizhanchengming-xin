import assert from 'node:assert/strict';
import test from 'node:test';

import { getVideoProviderConfig, putVideoProviderConfig } from './providerConfigApi.js';

function mockResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    async json() { return body; },
    async text() { return typeof body === 'string' ? body : JSON.stringify(body); },
  };
}

test('GET 404 is treated as unconfigured provider', async () => {
  const original = globalThis.fetch;
  globalThis.fetch = async () => mockResponse(404, 'not found');
  try {
    const result = await getVideoProviderConfig('personal_api');
    assert.equal(result, null);
  } finally { globalThis.fetch = original; }
});

test('PUT sends provider settings without leaking old credentials', async () => {
  const original = globalThis.fetch;
  let request;
  globalThis.fetch = async (url, init) => {
    request = { url, init };
    return mockResponse(200, { provider: 'personal_api', model: 'yd2.0-mini', configured: true, enabled: true });
  };
  try {
    const result = await putVideoProviderConfig('personal_api', { model: 'yd2.0-mini', apiKey: '', enabled: true });
    assert.equal(request.url, '/api/provider-configs/video/personal_api');
    assert.equal(request.init.method, 'PUT');
    assert.deepEqual(JSON.parse(request.init.body), { model: 'yd2.0-mini', api_key: '', enabled: true });
    assert.equal(result.configured, true);
  } finally { globalThis.fetch = original; }
});
