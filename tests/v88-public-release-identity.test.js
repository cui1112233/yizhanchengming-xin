const assert = require('node:assert/strict');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');

async function withServer(handler, callback) {
  const server = http.createServer(handler);
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  try {
    return await callback(server.address().port);
  } finally {
    await new Promise(resolve => server.close(resolve));
  }
}

test('Go runtime identity endpoint reports its injected immutable release SHA', async () => {
  const { createRuntimeBuildInfoHandler } = require('../lib/runtime-build-info');
  await withServer(createRuntimeBuildInfoHandler({ service: 'go-api', releaseSha: 'abc123' }), async port => {
    const response = await fetch(`http://127.0.0.1:${port}/api/runtime-build-info`);
    assert.equal(response.status, 200);
    assert.deepEqual(await response.json(), { service: 'go-api', git_sha: 'abc123' });
  });
});

test('public release builds and publishes only the immutable Go image', () => {
  const goDockerfile = fs.readFileSync(path.join(root, 'Dockerfile'), 'utf8');
  const workflow = fs.readFileSync(path.join(root, '.github/workflows/v88-unified-public-image-release.yml'), 'utf8');

  assert.match(goDockerfile, /FROM golang:/);
  assert.match(goDockerfile, /backend\/go\.mod/);
  assert.match(goDockerfile, /ENTRYPOINT \["\/qiantie"\]/);
  assert.doesNotMatch(goDockerfile, /CMD \["node", "server\.js"\]/);
  assert.match(workflow, /platforms:\s*linux\/amd64/);
  assert.match(workflow, /npm --prefix frontend ci/);
  assert.match(workflow, /npm --prefix frontend run build/);
  assert.match(workflow, /QIANTIE_RELEASE_SHA=\$\{\{ github\.sha \}\}/);
  assert.match(workflow, /qiantie-v88-go:\$\{\{ github\.sha \}\}/);
  assert.match(workflow, /context:\s*\.\s*\n\s*file:\s*Dockerfile/);
  assert.doesNotMatch(workflow, /qiantie-v88-node:/);
  assert.doesNotMatch(workflow, /go=unchanged/);
  assert.doesNotMatch(workflow, /node_image/);
  assert.doesNotMatch(workflow, /latest/);
});
