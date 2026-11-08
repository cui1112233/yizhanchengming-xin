import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

import { loadBatchFactoryPresetCatalog } from './batchFactoryPresetCatalog.js';

test('preset catalog keeps the successful module when the other module fails', async () => {
  const result = await loadBatchFactoryPresetCatalog(async module => {
    if (module === 'script') throw new Error('temporary script failure');
    return { catalog: [{ id: 'style-1', slot: 'batch.comic-style' }] };
  });

  assert.deepEqual(result.catalog, [{ id: 'style-1', slot: 'batch.comic-style' }]);
  assert.deepEqual(result.failedModules, ['script']);
});

test('preset catalog reports both modules when both requests fail', async () => {
  const result = await loadBatchFactoryPresetCatalog(async () => {
    throw new Error('offline');
  });

  assert.deepEqual(result.catalog, []);
  assert.deepEqual(result.failedModules, ['script', 'batch-factory']);
});

test('preset catalog API always bypasses the browser cache', () => {
  const source = fs.readFileSync(new URL('../../../shared/api/batchFactoryV12.js', import.meta.url), 'utf8');
  assert.match(source, /function listSystemPresetCatalog[\s\S]*?cache:\s*['"]no-store['"]/);
  assert.match(source, /function listSystemPresetCatalog[\s\S]*?Date\.now\(\)[\s\S]*?[?&]_=/);
});

test('unified settings primes the small preset catalog before the heavy book media render', () => {
  const source = fs.readFileSync(new URL('./BatchFactoryUnifiedSettingsModal.jsx', import.meta.url), 'utf8');
  assert.match(source, /const initialBatchFactoryPresetCatalog = loadBatchFactoryPresetCatalog\(listSystemPresetCatalog\)/);
  assert.match(source, /catalogReloadKey === 0[\s\S]*initialBatchFactoryPresetCatalog/);
});
