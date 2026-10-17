import test from 'node:test';
import assert from 'node:assert/strict';
import { cylinderVisualState, nextCylinderIndex } from './novel-fetch-cylinder.js';

test('centered cylinder item keeps full size and brightness', () => {
  assert.deepEqual(cylinderVisualState(0), {
    rotateX: 0,
    translateZ: 28,
    scale: 1,
    opacity: 1,
    brightness: 1
  });
});

test('distant cylinder item is clamped to readable edge values', () => {
  assert.deepEqual(cylinderVisualState(10), {
    rotateX: -48,
    translateZ: -74.4,
    scale: 0.78,
    opacity: 0.18,
    brightness: 0.52
  });
  assert.equal(cylinderVisualState(-10).rotateX, 48);
});

test('keyboard selection moves one item and stops at list boundaries', () => {
  assert.equal(nextCylinderIndex(5, 2, 1), 3);
  assert.equal(nextCylinderIndex(5, 2, -1), 1);
  assert.equal(nextCylinderIndex(5, 0, -1), 0);
  assert.equal(nextCylinderIndex(5, 4, 1), 4);
  assert.equal(nextCylinderIndex(0, 0, 1), -1);
});
