import assert from 'node:assert/strict';
import test from 'node:test';
import { DEFAULT_PET_ID, getPetDefinition, getPetOptions, petAnimationDelay, previewPetSelection } from './petCatalog.js';

test('offers Little Companion as the only selectable pet', () => {
  assert.equal(DEFAULT_PET_ID, 'little-companion');
  assert.deepEqual(getPetOptions(), [
    { label: '小家伙', value: 'little-companion' }
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

test('falls back to Little Companion for any unknown or legacy pet id', () => {
  assert.equal(getPetDefinition(undefined).id, 'little-companion');
  assert.equal(getPetDefinition(null).id, 'little-companion');
  assert.equal(getPetDefinition('').id, 'little-companion');
  assert.equal(getPetDefinition('legacy-unknown-pet').id, 'little-companion');
  assert.equal(getPetDefinition('stacky').id, 'little-companion');
  assert.equal(getPetDefinition('pixiu').id, 'little-companion');
  assert.equal(getPetDefinition({ id: 'unknown' }).id, 'little-companion');
});

test('previews the companion through the pet selection event', () => {
  const events = [];
  const previousWindow = globalThis.window;
  globalThis.window = { dispatchEvent: event => events.push(event) };

  try {
    const pet = previewPetSelection('little-companion');
    assert.equal(pet.id, 'little-companion');
    assert.equal(events.length, 1);
    assert.equal(events[0].type, 'qiantie:pet-selection');
    assert.equal(events[0].detail.pet.id, 'little-companion');
  } finally {
    globalThis.window = previousWindow;
  }
});

test('keeps the unified animation cadence for every state', () => {
  assert.equal(petAnimationDelay(getPetDefinition('little-companion'), 'idle'), 180);
  assert.equal(petAnimationDelay(getPetDefinition('little-companion'), 'working'), 120);
  assert.equal(petAnimationDelay(getPetDefinition('little-companion'), 'success'), 180);
});
