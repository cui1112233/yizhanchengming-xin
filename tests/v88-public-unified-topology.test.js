const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
const composeFile = path.join(root, 'deploy/v88-public/docker-compose.yml');
const envFile = path.join(root, 'deploy/v88-public/.env.example');
const nginxFile = path.join(root, 'deploy/v88-public/nginx.conf');

function composeConfig() {
  return JSON.parse(execFileSync('docker', [
    'compose', '--env-file', envFile, '-f', composeFile, 'config', '--format', 'json',
  ], { cwd: root, encoding: 'utf8' }));
}

test('public Compose renders the Go ingress by default and keeps legacy services opt-in', () => {
  const config = composeConfig();
  const services = config.services;

  assert.deepEqual(Object.keys(services).sort(), ['go-api', 'nginx']);
  assert.equal(services['go-api'].environment.QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER, '1');
  assert.equal(services['go-api'].environment.QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED, '1');
  assert.match(services['go-api'].environment.QIANTIE_MYSQL_DSN, /@tcp\(mysql:3306\)\//);
  assert.deepEqual(services.nginx.ports, [
    { mode: 'ingress', target: 80, published: '80', protocol: 'tcp' },
    { mode: 'ingress', target: 80, published: '3000', protocol: 'tcp' },
  ]);
  assert.ok(services.nginx.volumes.some(volume => volume.target === '/etc/letsencrypt' && volume.read_only));
  assert.equal(services['go-api'].build, undefined);
  assert.equal(config.networks.qiantie_internal.external, true);
});

test('public Nginx proxies application requests only through the Go service', () => {
  const nginx = fs.readFileSync(nginxFile, 'utf8');

  assert.match(nginx, /proxy_pass http:\/\/go-api:3000;/);
  assert.match(nginx, /server_name yzcm\.site;/);
  assert.match(nginx, /ssl_session_cache shared:SSL:10m;/);
  assert.match(nginx, /ssl_session_timeout 1d;/);
  assert.match(nginx, /keepalive_requests 1000;/);
  assert.doesNotMatch(nginx, /proxy_pass http:\/\/v88-node:3000;/);
  assert.doesNotMatch(nginx, /18081|172\.19\.|host\.docker\.internal/);
});
