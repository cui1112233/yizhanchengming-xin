const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/v88-unified-public-image-release.yml', 'utf8');
const directReadme = fs.readFileSync('deploy/v88-direct/README.md', 'utf8');

test('old GitHub release workflow is inert and marked retired', () => {
  assert.match(workflow, /RETIRED/);
  assert.match(workflow, /jobs:\s*\{\}/);
  assert.doesNotMatch(workflow, /docker|ghcr|build-push|npm ci/i);
});

test('direct Go deployment is the documented public release path', () => {
  assert.match(directReadme, /build-direct-release\.ps1/);
  assert.match(directReadme, /send-direct-release\.ps1/);
  assert.match(directReadme, /qiantie\.service/);
  assert.match(directReadme, /runtime-build-info/);
  assert.doesNotMatch(directReadme, /GitHub Actions 编译|Docker.*部署入口/);
});
