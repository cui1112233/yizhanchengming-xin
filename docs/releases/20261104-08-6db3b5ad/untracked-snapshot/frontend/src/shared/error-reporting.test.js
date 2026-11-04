import assert from 'node:assert/strict';
import test from 'node:test';

test('同一面板的同一错误只上报一次 client-errors', async () => {
  const requests = [];
  globalThis.window = { location: { pathname: '/novel-fetch' } };
  globalThis.localStorage = { getItem: () => 'token' };
  globalThis.fetch = async (path, options) => {
    requests.push({ path, body: JSON.parse(options.body) });
    return { ok: true };
  };
  const { reportClientError, resetClientErrorDedupe } = await import('./error-reporting.js');
  resetClientErrorDedupe();

  const payload = { kind: 'api-error', message: '状态服务暂不可用', source: '/api/batches/x/runtime-summary', method: 'GET', status: 503 };
  reportClientError(payload);
  reportClientError(payload);
  reportClientError(payload);
  assert.equal(requests.length, 1, '重复的 503 不应重复上报');

  reportClientError({ ...payload, message: '另一种错误' });
  assert.equal(requests.length, 2, '不同错误仍应上报');

  resetClientErrorDedupe();
  reportClientError(payload);
  assert.equal(requests.length, 3, '去重窗口内被重置后可以再次上报');
});
