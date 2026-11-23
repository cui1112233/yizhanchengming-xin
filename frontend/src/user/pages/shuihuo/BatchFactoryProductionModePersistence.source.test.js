import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const unifiedSource = fs.readFileSync(path.join(directory, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const novelListSource = fs.readFileSync(path.join(directory, 'BatchFactoryNovelList.jsx'), 'utf8');

test('production mode switch persists immediately without saving the rest of the draft', () => {
  assert.match(unifiedSource, /onProductionModeChange/);
  assert.match(unifiedSource, /await onProductionModeChange\?\.\(productionMode\)/);
  assert.match(novelListSource, /patch: \{ productionMode \}/);
});

test('persisting the unified production mode makes every book inherit that mode', () => {
  assert.match(novelListSource, /restoreKeys: \['productionMode'\]/);
  assert.match(novelListSource, /onProductionModeChange=\{saveProductionMode\}/);
});

test('batch mode persistence keeps all supported values and defaults unknown values to video', async () => {
  const persistSource = novelListSource.match(/^  async function saveProductionMode\(value\) \{[\s\S]*?^  \}/m)?.[0];
  assert.ok(persistSource, 'expected the batch mode persistence function');
  const writes = [];
  const batch = { id: 'batch-1', settingsState: { revision: 3 }, books: [] };
  const saveProductionMode = new Function('batch', 'saveBatchSettings', 'getBatch', 'resultData', 'refreshBatch',
    persistSource + '\nreturn saveProductionMode;')(
    batch,
    async (id, settings) => writes.push({ id, ...settings }),
    async () => ({ batch }),
    (response, key) => response[key],
    async () => {}
  );

  for (const value of ['video', 'comic', 'text', 'unknown']) {
    await saveProductionMode(value);
  }

  assert.deepEqual(writes, ['video', 'comic', 'text', 'video'].map(productionMode => ({
    id: 'batch-1',
    patch: { productionMode },
    expectedRevision: 3
  })));
});
