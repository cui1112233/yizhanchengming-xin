import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryAutomationMenu } from './batchFactoryAutomationMenu.js';
const item = (state, key) => batchFactoryAutomationMenu({ state }, true).find(row => row.key === key);
test('pause and resume follow actual controllable states', () => {
  assert.equal(item('running', 'pause').disabled, false);
  for (const state of ['idle', 'paused', 'scheduled', 'stopped', 'completed']) assert.equal(item(state, 'pause').disabled, true);
  assert.equal(item('paused', 'resume').disabled, false);
  assert.equal(item('completed', 'resume').disabled, true);
  assert.equal(item('stopped', 'cancel').disabled, true);
  assert.equal(item('needs_attention', 'cancel').disabled, false);
});
test('no failed task explains how to start instead of implying an unavailable retry', () => {
  const retry = item('idle', 'retry');
  assert.equal(retry.disabled, true);
  assert.match(retry.title, /启动/);
});
test('real failed book enables retry even when the aggregate incorrectly says zero', () => {
  const retry = batchFactoryAutomationMenu({ state: 'failed', counts: { failed: 0 }, books: [{ status: 'failed' }] }, true).find(row => row.key === 'retry');
  assert.equal(retry.disabled, false);
  assert.match(retry.label, /1/);
});
test('manual failed stages use the existing stage retry chain without inventing an automation job', () => {
  const retry = batchFactoryAutomationMenu({ state: 'idle' }, true, 2).find(row => row.key === 'retry-stages');
  assert.equal(retry.disabled, false);
});
