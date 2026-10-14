'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');

function read(path) { return fs.readFileSync(path, 'utf8'); }

const workflow = read('.github/workflows/v88-unified-public-image-release.yml');
const compose = read('deploy/v88-public/docker-compose.yml');

test('unified V88 release publishes one immutable Go image from the checked-out revision', () => {
  assert.match(workflow, /packages: write/);
  assert.match(workflow, /context: \./);
  assert.match(workflow, /file: Dockerfile/);
  assert.match(workflow, /QIANTIE_RELEASE_SHA=\$\{\{ github\.sha \}\}/);
  assert.match(workflow, /qiantie-v88-go/);
  assert.doesNotMatch(workflow, /qiantie-v88-node|qiantie-go-api/);
});

test('browser worker remains on the shared Docker DNS network without host IP discovery', () => {
  assert.match(compose, /browser-worker:/);
  assert.match(compose, /QIANTIE_121_BROWSER_WORKER_URL: http:\/\/browser-worker:8787/);
  assert.match(compose, /name: v88-public_qiantie_internal/);
  assert.doesNotMatch(compose, /NetworkSettings|container_ip|172\.19\./);
});

test('unified public compose uses the aligned 121 login timeout budget', () => {
  assert.match(compose, /QIANTIE_121_LOGIN_TIMEOUT_MS: \$\{QIANTIE_121_LOGIN_TIMEOUT_MS:-55000\}/);
});
