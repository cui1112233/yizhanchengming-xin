import test from 'node:test';
import assert from 'node:assert/strict';
import {
  batchFactoryProductionMode,
  batchFactoryProductionModeVisibility,
  batchFactoryShouldPrepareAudio,
  batchFactoryShouldRecompileVideoPrompts
} from './batchFactoryProductionMode.js';

test('defaults historical settings to video mode', () => {
  assert.equal(batchFactoryProductionMode(undefined), 'video');
  assert.deepEqual(batchFactoryProductionModeVisibility('unknown'), {
    mode: 'video',
    video: true,
    comic: false
  });
});

test('comic mode exposes comic settings and hides video settings', () => {
  assert.deepEqual(batchFactoryProductionModeVisibility('comic'), {
    mode: 'comic',
    video: false,
    comic: true
  });
});

test('comic unified settings never trigger video audio measurement or video prompt recompilation', () => {
  const comicSettings = { productionMode: 'comic', audioPlanningEnabled: true };
  assert.equal(batchFactoryShouldPrepareAudio(comicSettings), false);
  assert.equal(batchFactoryShouldRecompileVideoPrompts(comicSettings), false);
  assert.equal(batchFactoryShouldPrepareAudio({ productionMode: 'video', audioPlanningEnabled: true }), true);
  assert.equal(batchFactoryShouldRecompileVideoPrompts({ productionMode: 'video' }), true);
});
