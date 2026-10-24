import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const app = fs.readFileSync(path.join(here, 'App.jsx'), 'utf8');
const preload = fs.readFileSync(path.join(here, 'workbenchPreload.js'), 'utf8');

test('keeps Batch Factory on the single cover-library route', () => {
  assert.match(app, /const ShuihuoProductionPage = lazy\(\(\) => import\('\.\/pages\/ShuihuoProductionPage'\)\)/);
  assert.match(app, /'\/batch-factory': ShuihuoProductionPage/);
  assert.match(app, /function CanonicalBatchFactoryRedirect\(\)/);
  assert.match(app, /'\/shuihuo-production': CanonicalBatchFactoryRedirect/);
  assert.match(app, /'\/shuihuo-production\/creative': CanonicalBatchFactoryRedirect/);
  assert.doesNotMatch(app, /BatchFactoryWorkbenchPage/);
  assert.match(preload, /'\/batch-factory': \(\) => import\('\.\/pages\/ShuihuoProductionPage'\)/);
  assert.doesNotMatch(preload, /BatchFactoryWorkbenchPage/);
});
