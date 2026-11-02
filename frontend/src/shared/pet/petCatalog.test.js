import assert from 'node:assert/strict';
import test from 'node:test';
import { getPetDefinition, getPetOptions, petAnimationDelay, previewPetSelection } from './petCatalog.js';

test('offers Little Companion as the default first selectable pet', () => {
  assert.deepEqual(getPetOptions(), [
    { label: '小家伙', value: 'little-companion' },
    { label: 'CM', value: 'stacky' },
    { label: '貔貅', value: 'pixiu' }
  ]);
  assert.deepEqual(
    (({ id, displayName, renderer, assetBasePath, previewPose }) => ({ id, displayName, renderer, assetBasePath, previewPose }))(
      getPetDefinition('little-companion')
    ),
    {
      id: 'little-companion',
      displayName: '小家伙',
      renderer: 'multi-sheet',
      assetBasePath: '/pets/little-companion',
      previewPose: 'idle'
    }
  );
});

test('previews a selected pet immediately through the companion event', () => {
  const events = [];
  const previousWindow = globalThis.window;
  globalThis.window = { dispatchEvent: event => events.push(event) };

  try {
    const pet = previewPetSelection('pixiu');
    assert.equal(pet.id, 'pixiu');
    assert.equal(events.length, 1);
    assert.equal(events[0].type, 'qiantie:pet-selection');
    assert.equal(events[0].detail.pet.id, 'pixiu');
  } finally {
    globalThis.window = previousWindow;
  }
});

test('uses a slower animation cadence for Pixiu without changing CM timing', () => {
  assert.equal(petAnimationDelay(getPetDefinition('stacky'), 'idle'), 180);
  assert.equal(petAnimationDelay(getPetDefinition('stacky'), 'working'), 120);
  assert.equal(petAnimationDelay(getPetDefinition('pixiu'), 'idle'), 320);
  assert.equal(petAnimationDelay(getPetDefinition('pixiu'), 'working'), 220);
});
