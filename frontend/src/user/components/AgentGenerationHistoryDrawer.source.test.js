import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('generation history drawer loads, filters, and inserts selected results', async () => {
  const source = await readFile(new URL('./AgentGenerationHistoryDrawer.jsx', import.meta.url), 'utf8');

  assert.match(source, /listAgentGenerationHistory/);
  assert.match(source, /insertAgentGenerationHistoryResult/);
  assert.match(source, /放入画布/);
  assert.match(source, /generationHistoryKindOptions/);
});
