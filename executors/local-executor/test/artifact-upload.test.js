const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const fsp = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const { ExecutorApiClient } = require('../src/api-client');

test('artifact upload sends raw MP4 bytes with device and lease credentials', async () => {
  const seen = [];
  const client = new ExecutorApiClient({
    baseUrl: 'https://v78.example',
    fetchImpl: async (url, init) => {
      seen.push([url, init]);
      return { ok: true, status: 200, text: async () => JSON.stringify({ artifactId: 'artifact-1' }) };
    }
  });
  const bytes = Buffer.from('00000018667479706d703432', 'hex');
  const result = await client.uploadArtifact('device-token', 'job/a', { leaseToken: 'lease-secret', leaseGeneration: 7 }, bytes);
  assert.equal(result.artifactId, 'artifact-1');
  const [url, init] = seen[0];
  assert.equal(url, 'https://v78.example/api/local-executor/v1/jobs/job%2Fa/artifact');
  assert.equal(init.headers.Authorization, 'Bearer device-token');
  assert.equal(init.headers['Content-Type'], 'video/mp4');
  assert.equal(init.headers['X-Lease-Token'], 'lease-secret');
  assert.equal(init.headers['X-Lease-Generation'], '7');
  assert.equal(Buffer.compare(Buffer.from(init.body), bytes), 0);
});

test('artifact upload streams a local MP4 file instead of buffering it into JSON', async () => {
  const dir = await fsp.mkdtemp(path.join(os.tmpdir(), 'executor-artifact-'));
  const filePath = path.join(dir, 'video.mp4');
  const bytes = Buffer.from('000000186674797069736f6d0000000069736f6d6d703432766964656f', 'hex');
  await fsp.writeFile(filePath, bytes);
  const seen = [];
  const client = new ExecutorApiClient({
    baseUrl: 'https://v78.example',
    fetchImpl: async (url, init) => {
      const chunks = [];
      for await (const chunk of init.body) chunks.push(Buffer.from(chunk));
      seen.push([url, init, Buffer.concat(chunks)]);
      return { ok: true, status: 201, text: async () => JSON.stringify({ artifactId: 'artifact-file' }) };
    }
  });
  try {
    const result = await client.uploadArtifact('device-token', 'job-file', { leaseToken: 'lease', leaseGeneration: 2 }, filePath);
    assert.equal(result.artifactId, 'artifact-file');
    const [, init, uploaded] = seen[0];
    assert.equal(init.body instanceof fs.ReadStream, true);
    assert.equal(init.duplex, 'half');
    assert.equal(Buffer.compare(uploaded, bytes), 0);
  } finally {
    await fsp.rm(dir, { recursive: true, force: true });
  }
});
