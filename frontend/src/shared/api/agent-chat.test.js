import test from 'node:test';
import assert from 'node:assert/strict';
import { askAgent } from './agent.js';

test('Agent chat forwards the caller abort signal to the long-running request', async () => {
  const controller = new AbortController();
  let received;
  await askAgent({ taskId: 'project-1', prompt: '继续', context: { mode: '创作助手' }, signal: controller.signal }, async (path, options) => {
    received = { path, options };
    return { ok: true };
  });
  assert.equal(received.path, '/api/agent/chat');
  assert.equal(received.options.signal, controller.signal);
  assert.equal(received.options.suppressGlobalError, true);
});
