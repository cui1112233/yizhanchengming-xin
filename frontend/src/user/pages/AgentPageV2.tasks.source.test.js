import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');

test('task workspace exposes status, recovery actions, and canvas handoff', () => {
  assert.match(source, /generationPresentation\(entry\)/);
  assert.match(source, /generationTaskActions\(entry\)/);
  assert.match(source, /onCancelGeneration\(entry\)/);
  assert.match(source, /onRetryGeneration\(entry\)/);
  assert.match(source, /onOpenGeneration\(entry\)/);
  assert.match(source, />取消</);
  assert.match(source, />重新生成</);
  assert.match(source, />回到画布</);
});

test('task recovery refreshes the canonical generation list', () => {
  assert.match(source, /cancelAgentGeneration\(requestedTaskId, generation\.id\)/);
  assert.match(source, /retryAgentGeneration\(requestedTaskId, generation\.id,/);
  assert.match(source, /listAgentGenerations\(requestedTaskId\)/);
});
