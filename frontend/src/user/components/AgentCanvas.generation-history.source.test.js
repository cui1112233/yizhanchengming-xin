import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('canvas exposes generation history and refreshes its shared server canvas after insertion', async () => {
  const source = await readFile(new URL('./AgentCanvas.jsx', import.meta.url), 'utf8');

  assert.match(source, /AgentGenerationHistoryDrawer/);
  assert.match(source, /打开生成历史/);
  assert.match(source, /data-agent-generation-history-entry/);
  assert.match(source, /onServerCanvasChange/);
});
