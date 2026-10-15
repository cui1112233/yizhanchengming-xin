import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

test('Vite giant-material helper stays inside the frontend tree', () => {
  const helper = path.resolve('frontend/tools/giant-material/scroll-extractor.mjs');
  assert.equal(fs.existsSync(helper), true, `missing frontend helper: ${helper}`);
});
