import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldAutoCollapseAgentNavigation } from './agentNavigation.js';

test('Agent routes start with compact site navigation', () => {
  assert.equal(shouldAutoCollapseAgentNavigation('/agent'), true);
  assert.equal(shouldAutoCollapseAgentNavigation('/agent/canvas'), true);
  assert.equal(shouldAutoCollapseAgentNavigation('/settings'), false);
});
