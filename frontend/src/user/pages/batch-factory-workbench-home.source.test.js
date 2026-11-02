import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const css = fs.readFileSync(path.join(here, 'BatchFactoryWorkbenchPage.css'), 'utf8');
const page = fs.readFileSync(path.join(here, 'BatchFactoryWorkbenchPage.jsx'), 'utf8');

test('styles the canonical batch home without rendering the legacy readiness strip', () => {
  assert.match(css, /\.batch-factory-workbench-home/);
  assert.match(css, /\.batch-factory-workbench-grid/);
  assert.doesNotMatch(page, /shuihuo-readiness-strip/);
});

test('refreshes visible batch home periodically and preserves the last list on request errors', () => {
  assert.match(page, /setInterval\([^\n]*15000/);
  assert.match(page, /visibilitychange/);
  assert.match(page, /document\.hidden/);
  const catchStart = page.indexOf('} catch (requestError) {');
  const catchEnd = page.indexOf('} finally {', catchStart);
  const catchBlock = page.slice(catchStart, catchEnd);
  assert.match(catchBlock, /setError\(reason\)/);
  assert.doesNotMatch(catchBlock, /setBatches\(\[\]\)/);
});
