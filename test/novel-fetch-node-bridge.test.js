const test = require('node:test');
const assert = require('node:assert/strict');

const { resolveNovelFetchBridge } = require('../app');

test('小说工作台默认通过本机 Go 桥接且不回退到公网项目服务器', () => {
  const previous = process.env.QIANTIE_GO_BASE_URL;
  delete process.env.QIANTIE_GO_BASE_URL;
  try {
    const bridge = resolveNovelFetchBridge();
    assert.equal(bridge.targetBaseUrl, 'http://127.0.0.1:4000');
    assert.equal(/121w|qiantie\.com|公网/i.test(bridge.targetBaseUrl), false);
  } finally {
    if (previous === undefined) delete process.env.QIANTIE_GO_BASE_URL;
    else process.env.QIANTIE_GO_BASE_URL = previous;
  }
});

test('小说工作台使用 QIANTIE_GO_BASE_URL，并携带现有桥接密钥配置', () => {
  const previousBase = process.env.QIANTIE_GO_BASE_URL;
  const previousSecret = process.env.QIANTIE_BRIDGE_SECRET;
  process.env.QIANTIE_GO_BASE_URL = 'http://127.0.0.1:43123';
  process.env.QIANTIE_BRIDGE_SECRET = 'test-bridge-secret';
  try {
    assert.deepEqual(resolveNovelFetchBridge(), {
      targetBaseUrl: 'http://127.0.0.1:43123',
      bridgeSecret: 'test-bridge-secret'
    });
  } finally {
    if (previousBase === undefined) delete process.env.QIANTIE_GO_BASE_URL;
    else process.env.QIANTIE_GO_BASE_URL = previousBase;
    if (previousSecret === undefined) delete process.env.QIANTIE_BRIDGE_SECRET;
    else process.env.QIANTIE_BRIDGE_SECRET = previousSecret;
  }
});
