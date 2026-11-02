import assert from 'node:assert/strict';
import test from 'node:test';
import {
  canLittleCompanionRoam,
  clampLittleCompanionToViewport,
  interpolateLittleCompanionPosition,
  littleCompanionActivityBounds,
  littleCompanionEdgeSide,
  littleCompanionEdgeSpriteOffset,
  littleCompanionEdgeTarget,
  littleCompanionInwardTarget,
  littleCompanionPeekTarget,
  littleCompanionWalkDuration,
  littleCompanionWalkTarget
} from './littleCompanionRoaming.js';

test('uses the bottom 220 pixels as the desktop activity area', () => {
  assert.deepEqual(littleCompanionActivityBounds({ width: 1440, height: 900 }), {
    minLeft: 8,
    maxLeft: 1312,
    minTop: 680,
    maxTop: 762
  });
});

test('walks a bounded local distance instead of crossing the whole viewport', () => {
  assert.deepEqual(littleCompanionWalkTarget(
    { left: 500, top: 650 },
    { width: 1200, height: 800 },
    0,
    1
  ), {
    left: 660,
    top: 662,
    direction: 1
  });
  assert.equal(littleCompanionWalkDuration(
    { left: 500, top: 662 },
    { left: 660, top: 662 },
    1
  ), 1143);
});

test('detects only the correct nearby edge', () => {
  assert.equal(littleCompanionEdgeSide({ left: 20, top: 600 }, { width: 1000, height: 800 }), 'left');
  assert.equal(littleCompanionEdgeSide({ left: 850, top: 600 }, { width: 1000, height: 800 }), 'right');
  assert.equal(littleCompanionEdgeSide({ left: 440, top: 600 }, { width: 1000, height: 800 }), null);
  assert.equal(littleCompanionEdgeSide({ left: 0, top: 0 }, { width: 180, height: 300 }), null);
});

test('keeps the full shell visible while peeking and exits inward', () => {
  assert.deepEqual(littleCompanionEdgeTarget(
    { left: 20, top: 650 },
    { width: 1000, height: 800 },
    'left'
  ), { left: 8, top: 650, side: 'left' });
  assert.deepEqual(littleCompanionEdgeTarget(
    { left: 850, top: 650 },
    { width: 1000, height: 800 },
    'right'
  ), { left: 872, top: 650, side: 'right' });
  assert.deepEqual(littleCompanionInwardTarget(
    { left: 8, top: 650 },
    { width: 1000, height: 800 },
    'left',
    160
  ), { left: 168, top: 650, direction: 1 });
  assert.deepEqual(littleCompanionPeekTarget(
    { left: 850, top: 650 },
    { width: 1000, height: 800 }
  ), { left: 872, top: 650, side: 'right' });
});

test('compensates the transparent sprite margin when the companion reaches a wall', () => {
  assert.equal(littleCompanionEdgeSpriteOffset('left'), -22);
  assert.equal(littleCompanionEdgeSpriteOffset('right'), 22);
  assert.equal(littleCompanionEdgeSpriteOffset(null), 0);
});

test('clamps resize changes without pulling a manual upper position into the bottom band', () => {
  assert.deepEqual(clampLittleCompanionToViewport(
    { left: 900, top: 120 },
    { width: 700, height: 600 }
  ), { left: 572, top: 120 });
});

test('centers safely when the viewport cannot fit the normal pet bounds', () => {
  assert.deepEqual(littleCompanionActivityBounds({ width: 100, height: 120 }), {
    minLeft: 0,
    maxLeft: 0,
    minTop: 0,
    maxTop: 0
  });
});

test('clamps interpolation progress and ignores invalid coordinates', () => {
  assert.deepEqual(interpolateLittleCompanionPosition(
    { left: Number.NaN, top: 10 },
    { left: 110, top: 210 },
    2
  ), { left: 110, top: 210 });
  assert.deepEqual(interpolateLittleCompanionPosition(
    { left: 10, top: 10 },
    { left: 110, top: 210 },
    -1
  ), { left: 10, top: 10 });
});

test('allows roaming only for a visible idle Little Companion', () => {
  const ready = {
    petId: 'little-companion',
    state: 'idle',
    chatOpen: false,
    dragging: false,
    pageVisible: true,
    reducedMotion: false,
    spriteFailed: false
  };
  assert.equal(canLittleCompanionRoam(ready), true);
  for (const patch of [
    { petId: 'stacky' },
    { state: 'working' },
    { chatOpen: true },
    { dragging: true },
    { pageVisible: false },
    { reducedMotion: true },
    { spriteFailed: true }
  ]) {
    assert.equal(canLittleCompanionRoam({ ...ready, ...patch }), false);
  }
});
