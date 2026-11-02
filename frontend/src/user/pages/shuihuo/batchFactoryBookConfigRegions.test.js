import assert from 'node:assert/strict';
import test from 'node:test';
import { bookConfigRegionStatus } from './batchFactoryBookConfigRegions.js';

test('engine status reflects a book-specific comic or video prompt', () => {
  for (const module of ['comic', 'video']) {
    const book = { settingsState: { patch: { aiPromptConfig: { [module]: { presetId: 'preset' } } } } };
    assert.equal(bookConfigRegionStatus(book, 'engine').tone, 'overridden');
  }
  assert.equal(bookConfigRegionStatus({ settingsState: { patch: {} } }, 'engine').tone, 'inherited');
});
