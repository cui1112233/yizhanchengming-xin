import assert from 'node:assert/strict';
import test from 'node:test';
import { defaultEnabledModelSelection, loadScriptModelSelection, reconcileScriptModelSelection, saveScriptModelSelection } from './scriptModelSelection.js';

test('keeps each script selection while its configured model remains available', () => {
  assert.deepEqual(
    reconcileScriptModelSelection(
      { textModelId: 'text-qwen', imageModelId: 'image-flux' },
      {
        text: [{ id: 'text-qwen', displayName: 'Qwen', enabled: true }],
        image: [{ id: 'image-flux', displayName: 'Flux', enabled: true }]
      }
    ),
    { textModelId: 'text-qwen', imageModelId: 'image-flux', unavailableKinds: [] }
  );
});

test('clears only the selection whose configured model was removed or disabled', () => {
  assert.deepEqual(
    reconcileScriptModelSelection(
      { textModelId: 'text-qwen', imageModelId: 'image-flux' },
      {
        text: [{ id: 'text-qwen', displayName: 'Qwen', enabled: true }],
        image: [{ id: 'image-flux', displayName: 'Flux', enabled: false }]
      }
    ),
    { textModelId: 'text-qwen', imageModelId: '', unavailableKinds: ['image'] }
  );
});

test('persists model selection per user and restores it', () => {
  const values = new Map();
  const storage = { setItem: (key, value) => values.set(key, value), getItem: key => values.get(key) || null };
  saveScriptModelSelection(storage, 'alice', { textModelId: 't1', imageModelId: 'i1', videoModelKey: 'v1' });
  assert.deepEqual(loadScriptModelSelection(storage, 'alice'), { textModelId: 't1', imageModelId: 'i1', videoModelKey: 'v1' });
  assert.deepEqual(loadScriptModelSelection(storage, 'bob'), { textModelId: '', imageModelId: '', videoModelKey: '' });
});

test('defaults empty kinds to the first enabled configured model', () => {
  const models = {
    text: [{ id: 'text-disabled', enabled: false }, { id: 'text-configured', enabled: true }],
    image: [{ id: 'image-configured', enabled: true }]
  };
  assert.deepEqual(defaultEnabledModelSelection({ textModelId: '', imageModelId: '' }, models), {
    textModelId: 'text-configured',
    imageModelId: 'image-configured'
  });
  assert.deepEqual(defaultEnabledModelSelection({ textModelId: 'text-configured', imageModelId: '' }, models), {
    textModelId: 'text-configured',
    imageModelId: 'image-configured'
  });
});
