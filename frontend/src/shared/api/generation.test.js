import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

test('quick director uses the Go chat pipeline instead of the retired Node route', async () => {
  const source = await readFile(fileURLToPath(new URL('./generation.js', import.meta.url)), 'utf8');
  assert.match(source, /apiRequest\('\/api\/chat'/);
  assert.match(source, /promptType: 'quick_director'/);
  assert.doesNotMatch(source, /apiRequest\('\/api\/script\/director-pipeline'/);
});
