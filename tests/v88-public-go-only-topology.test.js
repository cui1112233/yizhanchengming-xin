const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const test = require('node:test');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const files = [
  '-f',
  'deploy/v88-public/docker-compose.yml',
  '-f',
  'deploy/v88-public/docker-compose.go-only.yml',
];

const baseFiles = ['-f', 'deploy/v88-public/docker-compose.yml'];

function environmentValue(service, key) {
  if (Array.isArray(service.environment)) {
    return service.environment
      .find((entry) => entry.startsWith(`${key}=`))
      ?.slice(key.length + 1);
  }
  return service.environment?.[key];
}

function publicGoOnlyConfig() {
  const result = spawnSync(
    'docker',
    ['compose', ...files, 'config', '--no-interpolate', '--format', 'json'],
    { cwd: root, encoding: 'utf8' },
  );
  assert.equal(result.status, 0, result.stderr || 'Go-only public compose config failed');
  return JSON.parse(result.stdout);
}

function publicBaseConfig() {
  const result = spawnSync(
    'docker',
    ['compose', ...baseFiles, 'config', '--no-interpolate', '--format', 'json'],
    { cwd: root, encoding: 'utf8' },
  );
  assert.equal(result.status, 0, result.stderr || 'Public compose config failed');
  return JSON.parse(result.stdout);
}

test('public Go-only topology routes nginx to Go and removes legacy services', () => {
  const config = publicGoOnlyConfig();
  const goApi = config.services['go-api'];
  const nginx = config.services.nginx;

  assert.ok(goApi);
  assert.equal(environmentValue(goApi, 'QIANTIE_GO_LISTEN_ADDR'), ':3000');
  assert.ok(goApi.expose.includes('3000'));
  assert.deepEqual(nginx.depends_on, {
    'go-api': { condition: 'service_started', required: true },
  });
  assert.equal(goApi.networks?.qiantie_internal?.aliases, undefined);
  assert.equal(environmentValue(goApi, 'QIANTIE_BATCH_FACTORY_V11_SLICE'), '${QIANTIE_BATCH_FACTORY_V11_SLICE:-6}');
  assert.equal(config.services['v88-node'], undefined);
  assert.equal(config.services['browser-worker'], undefined);
  assert.equal(config.services['shuihuo-compat'], undefined);
});

test('public base topology is Go-first without a second default request path', () => {
  const config = publicBaseConfig();
  const goApi = config.services['go-api'];
  const nginx = config.services.nginx;

  assert.equal(nginx.depends_on['go-api'].condition, 'service_started');
  assert.equal(nginx.depends_on['v88-node'], undefined);
  assert.ok(goApi.expose.includes('3000'));
  assert.equal(goApi.networks?.qiantie_internal?.aliases, undefined);
  assert.equal(environmentValue(goApi, 'QIANTIE_BATCH_FACTORY_V11_SLICE'), '${QIANTIE_BATCH_FACTORY_V11_SLICE:-6}');
  assert.deepEqual(config.services['v88-node'].profiles, ['node-legacy']);
  assert.deepEqual(config.services['browser-worker'].profiles, ['node-legacy']);
  assert.deepEqual(config.services['shuihuo-compat'].profiles, ['node-legacy']);
});
