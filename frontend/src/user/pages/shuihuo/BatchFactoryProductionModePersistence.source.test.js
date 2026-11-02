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
