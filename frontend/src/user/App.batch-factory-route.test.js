import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const app = fs.readFileSync(path.join(here, 'App.jsx'), 'utf8');

test('keeps Batch Factory on its durable workbench route', () => {
  assert.match(app, /const loadBatchFactoryWorkbenchPage = \(\) => import\('\.\/pages\/BatchFactoryWorkbenchPage'\)/);
  assert.match(app, /const BatchFactoryWorkbenchPage = lazy\(loadBatchFactoryWorkbenchPage\)/);
  assert.match(app, /'\/batch-factory': BatchFactoryWorkbenchPage/);
  assert.match(app, /function CanonicalBatchFactoryRedirect\(\)/);
  assert.match(app, /'\/shuihuo-production': CanonicalBatchFactoryRedirect/);
  assert.match(app, /'\/shuihuo-production\/creative': ShuihuoProductionPage/);
  for (const requiredFile of ['pages/BatchFactoryWorkbenchPage.jsx', 'pages/BatchFactoryWorkbenchPage.css']) {
    assert.equal(fs.existsSync(path.join(here, requiredFile)), true, `${requiredFile} must remain available`);
  }
});
