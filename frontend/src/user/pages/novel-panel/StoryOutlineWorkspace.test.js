import test from 'node:test';
import assert from 'node:assert/strict';
import { buildOutlineViewModel } from './storyOutlineModel.js';

test('keeps scenes and shots in source order for display', () => {
  const model = buildOutlineViewModel({
    outline_shots: [
      { scene_index: 2, shot_index: 1, prompt: '第二场' },
      { scene_index: 1, shot_index: 1, prompt: '第一场' }
    ]
  });
  assert.deepEqual(model.scenes.map(scene => scene.index), [1, 2]);
  assert.equal(model.scenes[0].shots[0].prompt, '第一场');
});

test('keeps the previous outline when a generation response is invalid', () => {
  const previous = { scenes: [{ index: 1, shots: [{ prompt: '旧提示词' }] }] };
  assert.deepEqual(buildOutlineViewModel(null, previous), previous);
});

test('marks unsaved edits without changing the generated prompt', () => {
  const model = buildOutlineViewModel({ outline_shots: [{ scene_index: 1, shot_index: 1, prompt: '原提示词' }] });
  assert.equal(model.scenes[0].shots[0].prompt, '原提示词');
  assert.equal(model.dirty, false);
});
