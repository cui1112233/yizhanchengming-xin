import test from 'node:test';
import assert from 'node:assert/strict';

import { createAgentVideoJob } from './agent.js';

test('createAgentVideoJob exposes a persisted failed job from a non-2xx provider response', async t => {
  const originalLocalStorage = globalThis.localStorage;
  const originalFetch = globalThis.fetch;
  globalThis.localStorage = { getItem: () => 'token', setItem() {}, removeItem() {} };
  globalThis.fetch = async () => new Response(JSON.stringify({
    error: 'provider unavailable',
    job: { id: 'job-failed', projectId: 'project-1', nodeId: 'video-1', state: 'failed', errorMessage: 'provider unavailable' }
  }), { status: 502, headers: { 'content-type': 'application/json' } });
  t.after(() => {
    globalThis.localStorage = originalLocalStorage;
    globalThis.fetch = originalFetch;
  });

  const result = await createAgentVideoJob('project-1', { nodeId: 'video-1' });
  assert.equal(result.error, 'provider unavailable');
  assert.equal(result.job.id, 'job-failed');
  assert.equal(result.job.state, 'failed');
});
