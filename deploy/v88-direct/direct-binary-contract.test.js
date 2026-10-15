const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const here = __dirname;
const read = (name) => fs.readFileSync(path.join(here, name), 'utf8');

test('direct release scripts use a Go binary and systemd, never Docker or GitHub Actions', () => {
  const build = read('build-direct-release.ps1');
  const send = read('send-direct-release.ps1');
  const activate = read('activate-direct-release.sh');
  const service = read('qiantie.service');

  assert.match(build, /GOOS.*linux/);
  assert.match(build, /qiantie\.service/);
  assert.doesNotMatch(build, /(?:^|[\s'\"])(?:docker|docker-compose|ghcr)(?:$|[\s'\"])/im);
  assert.doesNotMatch(send, /(?:docker-compose|docker compose|GHCR)/i);
  assert.doesNotMatch(activate, /docker compose|\bdocker\s/i);
  assert.match(activate, /systemctl/);
  assert.match(activate, /runtime-build-info/);
  assert.match(service, /ExecStart=.*qiantie/);
  assert.match(service, /EnvironmentFile=/);
});

test('nginx routes directly to the host Go listener', () => {
  const nginx = fs.readFileSync(path.join(here, '..', 'v88-public', 'nginx.conf'), 'utf8');
  assert.match(nginx, /proxy_pass\s+http:\/\/127\.0\.0\.1:3000/);
  assert.doesNotMatch(nginx, /proxy_pass\s+http:\/\/go-api:3000/);
});
