import assert from 'node:assert/strict';
import test from 'node:test';
import { insertAgentGenerationHistoryResult, listAgentGenerationHistory } from './agent.js';

test('generation history sends only non-empty filters to the owner-scoped route', async () => {
  const calls = [];
  const request = async (path, options) => {
    calls.push({ path, options });
    return { tasks: [] };
  };

  await listAgentGenerationHistory({ kind: 'image', modelId: 'doubao-image', state: '', query: '  夜景  ' }, request);

  assert.deepEqual(calls, [{
    path: '/api/agent/generation-history?kind=image&modelId=doubao-image&q=%E5%A4%9C%E6%99%AF',
    options: { suppressGlobalError: true }
  }]);
});

test('generation history inserts one selected result into the current project canvas', async () => {
  const calls = [];
  const request = async (path, options) => {
    calls.push({ path, options });
    return { node: { id: 'history-node-1' } };
  };

  await insertAgentGenerationHistoryResult('current project', 'source task', 'result/1', request);

  assert.deepEqual(calls, [{
    path: '/api/agent/tasks/current%20project/generation-history/source%20task/results/result%2F1/insert',
    options: { method: 'POST', suppressGlobalError: true }
  }]);
});
