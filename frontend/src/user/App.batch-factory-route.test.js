import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const app = fs.readFileSync(path.join(here, 'App.jsx'), 'utf8');

test('routes the Shuihuo entry to the shared work library while retaining the Batch Factory entry', () => {
  assert.match(app, /const loadBatchFactoryWorkbenchPage = \(\) => import\('\.\/pages\/BatchFactoryWorkbenchPage'\)/);
  assert.match(app, /const BatchFactoryWorkbenchPage = lazy\(loadBatchFactoryWorkbenchPage\)/);
  assert.match(app, /'\/batch-factory': BatchFactoryWorkbenchPage/);
  assert.match(app, /'\/shuihuo-production': ShuihuoProductionPage/);
  assert.doesNotMatch(app, /BatchFactoryFromShuihuoPage/);
});
