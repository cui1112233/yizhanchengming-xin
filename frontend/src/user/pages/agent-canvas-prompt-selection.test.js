import assert from 'node:assert/strict';
import test from 'node:test';
import { batchFactoryPromptSelectorTypes, saveBatchFactoryPromptSelection } from './agent-canvas-prompt-selection.js';

test('maps each Batch Factory card to only its applicable prompt templates', () => {
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'batch_factory', resourceType: 'book_script' }), ['hook', 'director']);
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'batch_factory', resourceType: 'book_asset' }), ['visual']);
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'batch_factory', resourceType: 'video' }), ['video']);
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'batch_factory', resourceType: 'audio' }), ['audio_match']);
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'batch_factory', resourceType: 'merge' }), ['book_merge']);
  assert.deepEqual(batchFactoryPromptSelectorTypes({ source: 'manual', resourceType: 'video' }), []);
});

test('saving a canvas prompt template preserves the other Batch Factory selections', async () => {
  let received = null;
  const batch = {
    id: 'batch-1',
    settingsState: {
      revision: 7,
      patch: { promptSelections: { visual: 'visual-old', audio_match: 'audio-old' } }
    }
  };

  const saved = await saveBatchFactoryPromptSelection({
    batch,
    type: 'video',
    promptId: 'video-new',
    save: async (batchId, input) => {
      received = { batchId, input };
      return { revision: 8 };
    }
  });

  assert.deepEqual(received, {
    batchId: 'batch-1',
    input: {
      patch: { promptSelections: { visual: 'visual-old', audio_match: 'audio-old', video: 'video-new' } },
      expectedRevision: 7
    }
  });
  assert.deepEqual(saved, { revision: 8 });
});
