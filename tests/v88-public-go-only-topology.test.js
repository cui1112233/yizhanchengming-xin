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

function publicGoOnlyConfig() {
  const result = spawnSync(
    'docker',
    ['compose', ...files, 'config', '--no-interpolate', '--format', 'json'],
    { cwd: root, encoding: 'utf8' },
  );
  assert.equal(result.status, 0, result.stderr || 'Go-only public compose config failed');
  return JSON.parse(result.stdout);
}

test('public Go-only topology routes nginx to Go and isolates legacy services', () => {
  const config = publicGoOnlyConfig();
  const goApi = config.services['go-api'];
  const nginx = config.services.nginx;
  const environment = Object.fromEntries(
    goApi.environment.map((entry) => entry.split('=', 2)),
  );

  assert.ok(goApi);
  assert.equal(environment.QIANTIE_GO_LISTEN_ADDR, ':3000');
  assert.ok(goApi.expose.includes('3000'));
  assert.deepEqual(nginx.depends_on, {
    'go-api': { condition: 'service_started', required: true },
  });
  assert.ok(goApi.networks.qiantie_internal.aliases.includes('v88-node'));
  assert.deepEqual(config.services['v88-node'].profiles, ['node-legacy']);
  assert.deepEqual(config.services['browser-worker'].profiles, ['node-legacy']);
  assert.deepEqual(config.services['shuihuo-compat'].profiles, ['node-legacy']);
});
