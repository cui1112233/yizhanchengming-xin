import test from 'node:test';
import assert from 'node:assert/strict';
import { responsiveCompanionPosition } from './companion-position.js';

test('mobile viewport drops a stale dragged position and uses the safe corner', () => {
  assert.deepEqual(responsiveCompanionPosition({ left: 8, top: 300 }, 390, 844), { left: null, top: null });
});

test('desktop viewport clamps a dragged position inside the viewport', () => {
  assert.deepEqual(responsiveCompanionPosition({ left: 9999, top: -20 }, 1280, 720), { left: 1152, top: 58 });
});
