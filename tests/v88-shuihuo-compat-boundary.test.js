const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const http = require('node:http');
const crypto = require('node:crypto');
const express = require('express');
const { createShuihuoProductionRouter } = require('../routes/shuihuo-production');
const { listShuihuoProjects } = require('../routes/platform-projects');

const manifest = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../deploy/v88-public/runtime-topology.manifest.json'), 'utf8'));

test('compatibility service is explicitly legacy-opt-in while Go owns the default chain', () => {
  const compat = manifest.services.find(({ id }) => id === 'shuihuo-compat');
  assert.equal(compat.classification, 'legacy-opt-in');
  assert.match(compat.activation, /legacy profile only/);
  assert.ok(compat.publicContracts.length >= 2);
  for (const contract of compat.publicContracts) {
    assert.ok(contract.id && contract.publicRoute && contract.owner && contract.upstreamRoute);
    assert.ok(contract.retirementGate && contract.rollbackTest);
  }
  assert.equal(compat.trafficObservation.status, 'unavailable');
  assert.equal(compat.trafficObservation.requiredWindowHours, 24);
  assert.equal(compat.trafficObservation.observedWindowHours, null);
  assert.equal(compat.trafficObservation.routeCounts, null);
  assert.match(compat.retirementGate, /Go/);
});

test('signed production forwarding and platform aggregation use the retained compatibility API', async t => {
  const secret = 'boundary-test-secret';
  const account = { username: 'boundary-user', isOwner: false };
  const upstream = http.createServer((req, res) => {
    const pathname = req.url.split('?')[0];
    const issuedAt = req.headers['x-qiantie-issued-at'];
    const signature = crypto.createHmac('sha256', secret)
      .update([account.username, issuedAt, 'false', req.method, pathname].join('')).digest('hex');
    if (req.headers['x-qiantie-username'] !== account.username || req.headers['x-qiantie-signature'] !== signature || pathname !== '/api/shuihuo-production/projects') {
      res.writeHead(401).end();
      return;
    }
    res.setHeader('Content-Type', 'application/json');
    res.end(JSON.stringify({ projects: [{ id: 7, name: 'compat-project', segmentationStatus: 'confirmed', createdAt: '2026-10-05T01:00:00Z' }] }));
  });
  await new Promise(resolve => upstream.listen(0, '127.0.0.1', resolve));
  t.after(() => new Promise(resolve => upstream.close(resolve)));
  const gatewayOptions = { targetBaseUrl: `http://127.0.0.1:${upstream.address().port}`, bridgeSecret: secret };
  const app = express();
  app.use('/api/shuihuo-production', createShuihuoProductionRouter({
    ...gatewayOptions, authenticate(req, _res, next) { req.auth = { account }; next(); }
  }));
  const gateway = http.createServer(app);
  await new Promise(resolve => gateway.listen(0, '127.0.0.1', resolve));
  t.after(() => new Promise(resolve => gateway.close(resolve)));
  const response = await fetch(`http://127.0.0.1:${gateway.address().port}/api/shuihuo-production/projects`);
  assert.equal(response.status, 200);
  assert.equal((await response.json()).projects[0].name, 'compat-project');
  const entries = await listShuihuoProjects(gatewayOptions, account);
  assert.equal(entries[0].id, 'shuihuo-production:7');
  assert.equal(entries[0].route, '/shuihuo-production?project=7');
});
