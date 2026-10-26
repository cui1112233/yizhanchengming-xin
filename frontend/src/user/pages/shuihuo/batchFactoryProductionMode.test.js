import test from 'node:test';
import assert from 'node:assert/strict';
import {
  batchFactoryProductionMode,
  batchFactoryProductionModeVisibility
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
