import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('canvas refreshes only mapped Batch Factory books that are still running', async () => {
  const source = await readFile(new URL('./AgentCanvas.jsx', import.meta.url), 'utf8');

  assert.match(source, /mappedBatchFactoryBooksNeedingRefresh/);
  assert.match(source, /const pendingMappedBooks = useMemo/);
  assert.match(source, /await Promise\.all\(pendingMappedBooks\.map\(\(\{ bookId \}\) => syncAgentBatchFactoryBook/);
  assert.match(source, /window\.setTimeout/);
});
