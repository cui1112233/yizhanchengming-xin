import test from 'node:test';
import assert from 'node:assert/strict';
import { createAgentImageTool } from './agent.js';

test('submits image tools through the project generation API', async () => {
  const calls = [];
  const request = async (url, options) => {
    calls.push({ url, options });
    return { task: { id: 'generation-tool' } };
  };
  const payload = { nodeId: 'image-1', nodeVersion: 2, operation: 'mirror', sourceAssetId: 'asset-1', settings: { direction: 'horizontal' }, idempotencyKey: 'tool-1' };
  const result = await createAgentImageTool('project-1', payload, request);
  assert.deepEqual(result, { task: { id: 'generation-tool' } });
  assert.deepEqual(calls, [{
    url: '/api/agent/tasks/project-1/image-tools',
    options: { method: 'POST', body: JSON.stringify(payload), suppressGlobalError: true }
  }]);
});
