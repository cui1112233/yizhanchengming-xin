import assert from 'node:assert/strict';
import test from 'node:test';

import {
  detectDirectorMode,
  normalizeDirectorMode,
  summarizeDirectorPlan
} from './agent-director-mode.js';

test('director mode stays supervised until the user explicitly requests automatic execution', () => {
  assert.equal(normalizeDirectorMode(''), 'supervised');
  assert.equal(normalizeDirectorMode('unexpected'), 'supervised');
  assert.equal(detectDirectorMode('帮我做这本书', 'supervised'), 'supervised');
  assert.equal(detectDirectorMode('先给我看看方案再做', 'automatic'), 'supervised');
  assert.equal(detectDirectorMode('全自动化完成这本书', 'supervised'), 'automatic');
});

test('director plan summary exposes review and execution counts', () => {
  assert.deepEqual(summarizeDirectorPlan({
    commands: [
      { type: 'create_node' },
      { type: 'remove_nodes', nodeIds: ['node-1'] },
      { type: 'start_generation', nodeId: 'image-1' }
    ]
  }), {
    commandCount: 3,
    destructiveCount: 1,
    executionCount: 1
  });
});
