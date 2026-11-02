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

test('distant cylinder item keeps a readable presence at the edge', () => {
  const edge = cylinderVisualState(10);
  assert.equal(edge.rotateX, -30);
  assert.equal(edge.translateZ, -29.6);
  assert.equal(edge.scale, 0.9);
  assert.ok(edge.opacity >= 0.64);
  assert.ok(edge.brightness >= 0.84);
  assert.equal(cylinderVisualState(-10).rotateX, 30);
});

test('keyboard selection moves one item and stops at list boundaries', () => {
  assert.equal(nextCylinderIndex(5, 2, 1), 3);
  assert.equal(nextCylinderIndex(5, 2, -1), 1);
  assert.equal(nextCylinderIndex(5, 0, -1), 0);
  assert.equal(nextCylinderIndex(5, 4, 1), 4);
  assert.equal(nextCylinderIndex(0, 0, 1), -1);
});
