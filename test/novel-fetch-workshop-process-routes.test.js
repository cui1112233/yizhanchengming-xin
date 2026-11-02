const test = require('node:test');
const assert = require('node:assert/strict');
const express = require('express');
const { createNovelFetchWorkshopRouter } = require('../routes/novel-fetch-workshop');
const { resolveNovelFetchBridge } = require('../app');

function listen(app) {
  return new Promise(resolve => app.listen(0, '127.0.0.1', function onListen() { resolve(this); }));
}
function close(server) {
  return new Promise((resolve, reject) => server.close(error => error ? reject(error) : resolve()));
}

test('桥接配置允许本机和 Docker backend，并拒绝公网目标与默认密钥', () => {
  assert.equal(resolveNovelFetchBridge({ QIANTIE_GO_BASE_URL: 'http://backend:4000', QIANTIE_BRIDGE_SECRET: 'explicit-secret' }).targetBaseUrl, 'http://backend:4000');
  assert.throws(() => resolveNovelFetchBridge({ QIANTIE_GO_BASE_URL: 'http://8.8.8.8:4000', QIANTIE_BRIDGE_SECRET: 'explicit-secret' }), /只允许/);
  assert.equal(resolveNovelFetchBridge({}).bridgeSecret, '');
});

test('小说工作台后台作业路径转发到已有 batch-rewrite 作业实现', async () => {
  const delegated = express.Router();
  delegated.post('/process/start', (_req, res) => res.json({ id: 'job-1', status: 'running' }));
  delegated.get('/process/jobs/latest', (_req, res) => res.json({ latest: { id: 'job-1' }, jobs: [{ id: 'job-1' }] }));
  const router = createNovelFetchWorkshopRouter({
    auth: (_req, _res, next) => next(),
    processRouter: delegated,
    tasks: { async getConfig() { return {}; } },
    configStore: { getConfig() { return {}; }, getPlatforms() { return []; }, getStyles() { return []; } }
  });
  const app = express();
  app.use(express.json());
  app.use('/api/novel-fetch-workshop', router);
  const server = await listen(app);
  try {
    const base = `http://127.0.0.1:${server.address().port}/api/novel-fetch-workshop`;
    const start = await fetch(`${base}/process/start`, { method: 'POST', headers: { 'content-type': 'application/json' }, body: '{}' });
    assert.notEqual(start.status, 404);
    assert.equal((await start.json()).id, 'job-1');
    const latest = await fetch(`${base}/process/jobs/latest?limit=1`);
    assert.notEqual(latest.status, 404);
    assert.equal((await latest.json()).latest.id, 'job-1');
  } finally {
    await close(server);
  }
});
