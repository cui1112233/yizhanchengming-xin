const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
function loadManifest(relativePath) {
  return JSON.parse(fs.readFileSync(path.join(root, relativePath), 'utf8'));
}

test('runtime topology declares every public V88 service and excludes host-local secrets', () => {
  const manifest = loadManifest('deploy/v88-public/runtime-topology.manifest.json');
  assert.deepEqual(manifest.services.map(({ id }) => id).sort(), [
    'browser-worker', 'go-api', 'mysql', 'nginx', 'node', 'shuihuo-compat',
  ]);
  const serialized = JSON.stringify(manifest);
  assert.equal(serialized.includes('/opt/qiantie/'), false);
  assert.equal(serialized.includes('.env'), false);
  assert.doesNotMatch(serialized, /\/(?:Users|home|root|etc|var|app|data|tmp)\//);

  const ids = new Set(manifest.services.map(({ id }) => id));
  for (const service of manifest.services) {
    for (const field of ['source', 'runtimeImageOrBinary', 'publicExposure']) {
      assert.ok(service[field] && typeof service[field] === 'object', `${service.id}: ${field}`);
    }
    assert.ok(Array.isArray(service.dependsOn), `${service.id}: dependsOn`);
    for (const dependency of service.dependsOn) assert.ok(ids.has(dependency), dependency);
    assert.ok(Array.isArray(service.persistentDataPurpose), `${service.id}: persistentDataPurpose`);
    assert.ok(service.retirementGate.trim(), `${service.id}: retirementGate`);
    for (const file of service.source.configurationFiles) {
      assert.equal(path.isAbsolute(file), false, file);
      assert.equal(file.split('/').includes('..'), false, file);
      assert.ok(fs.existsSync(path.join(root, file)), file);
    }
    if (service.id !== 'nginx') assert.deepEqual(service.publicExposure.publishedPorts, []);
  }
  assert.deepEqual(manifest.services.find(({ id }) => id === 'nginx').publicExposure.publishedPorts,
    [{ host: 80, target: 80 }, { host: 3000, target: 80 }]);
  const compose = fs.readFileSync(path.join(root, 'deploy/v88-public/docker-compose.yml'), 'utf8');
  assert.doesNotMatch(compose, /-\s*["']443:443["']/);
  assert.deepEqual(manifest.services.find(({ id }) => id === 'node').dependsOn,
    ['go-api', 'browser-worker', 'shuihuo-compat']);
  for (const id of ['go-api', 'shuihuo-compat']) {
    assert.ok(manifest.services.find((service) => service.id === id).dependsOn.includes('mysql'));
  }
});

test('package defaults to Go and exposes no Node application start script', () => {
  const pkg = loadManifest('package.json');
  assert.match(pkg.scripts.start, /start-local-go-review\.ps1/);
  assert.equal(Object.prototype.hasOwnProperty.call(pkg.scripts, 'start:node-legacy'), false);
  assert.equal(pkg.dependencies, undefined, 'the Go release must not install Node application dependencies');
  assert.deepEqual(Object.keys(pkg.devDependencies).sort(), [
    '@volcengine/tos-sdk', 'express', 'https-proxy-agent',
  ]);
});
