const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/v88-unified-public-image-release.yml', 'utf8');

test('unified V88 release can be explicitly requested from v88 without using retired host-stage paths', () => {
  assert.match(workflow, /push:\s*[\s\S]*branches:\s*\[?v88\]?/);
  assert.match(workflow, /deploy\/v88-public\/UNIFIED-DEPLOY-REQUEST/);
  assert.doesNotMatch(workflow, /stage-node-host\.sh|cutover-node-host\.sh/);
  assert.doesNotMatch(workflow, /(?:curl|proxy_pass)[^\n]*18081/);
});

test('V88 release publishes only the immutable Go image', () => {
  assert.match(workflow, /qiantie-v88-go:\$\{\{ github\.sha \}\}/);
  assert.match(workflow, /QIANTIE_RELEASE_SHA=\$\{\{ github\.sha \}\}/);
  assert.doesNotMatch(workflow, /qiantie-v88-node|QIANTIE_NODE_IMAGE|v88-node/);
});

test('unified V88 release reports the Go runtime identity', () => {
  assert.match(workflow, /runtime=go-api/);
  assert.match(workflow, /node=retired/);
});

test('unified V88 release no longer contains the retired remote deploy path', () => {
  assert.doesNotMatch(workflow, /ECS_HOST|V88_ECS_SSH_PRIVATE_KEY|docker compose/);
});

test('unified V88 release keeps the Git-managed trigger path', () => {
  assert.match(workflow, /deploy\/v88-public\/UNIFIED-DEPLOY-REQUEST/);
});

test('unified V88 release builds from the repository Dockerfile', () => {
  assert.match(workflow, /context:\s*\.\s*\n\s*file:\s*Dockerfile/);
  assert.match(workflow, /docker\/build-push-action/);
});

test('unified V88 release publishes through GHCR with the short-lived GitHub token', () => {
  assert.match(workflow, /password: \$\{\{ secrets\.GITHUB_TOKEN \}\}/);
  assert.match(workflow, /push:\s*true/);
  assert.match(workflow, /qiantie-v88-go/);
});

test('unified V88 release does not claim external deployment acceptance', () => {
  assert.doesNotMatch(workflow, /external rejection|deploy_public|ECS_PUBLIC_HOST/);
});

test('unified V88 release verifies source contracts before publishing', () => {
  assert.match(workflow, /Verify Go-only release contracts/);
  assert.match(workflow, /tests\/v88-public-release-identity\.test\.js/);
});

test('unified V88 release embeds the frontend bundle into Go', () => {
  assert.match(workflow, /rm -rf backend\/web\/dist/);
  assert.match(workflow, /cp -a frontend\/dist\/\. backend\/web\/dist\//);
});

test('unified V88 release has no Node runtime smoke step', () => {
  assert.doesNotMatch(workflow, /docker exec[^\n]*node|node_image|npm ci --omit=dev/);
});
