const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { downloadReferenceImages } = require('../src/reference-images');

test('downloads protected reference URLs as local image paths', async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'executor-reference-'));
  try {
    const images = await downloadReferenceImages({
      imageUrls: ['/api/local-executor/v1/jobs/job-1/agent-reference-assets/asset-1'],
      jobId: 'job-1',
      baseUrl: 'http://server.example',
      token: 'executor-token',
      tempDir,
      fetchImpl: async (url, init) => {
        assert.equal(url, 'http://server.example/api/local-executor/v1/jobs/job-1/agent-reference-assets/asset-1');
        assert.equal(init.headers.Authorization, 'Bearer executor-token');
        return new Response(new Uint8Array([137, 80, 78, 71]), { headers: { 'content-type': 'image/png' } });
      }
    });
    assert.equal(images.length, 1);
    assert.match(images[0], /\.png$/);
    assert.deepEqual([...fs.readFileSync(images[0])], [137, 80, 78, 71]);
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
});
