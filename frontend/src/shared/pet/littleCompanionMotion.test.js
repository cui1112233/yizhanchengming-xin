import assert from 'node:assert/strict';
import test from 'node:test';
import {
  littleCompanionFrame,
  littleCompanionIdleFrameAfterWalk,
  littleCompanionPoseStep,
  littleCompanionTimeline,
  littleCompanionWalkFrame,
  nextLittleCompanionStep
} from './littleCompanionMotion.js';

test('keeps idle stable until the parent behavior scheduler chooses an action', () => {
  assert.deepEqual(littleCompanionTimeline('idle'), [
    { pose: 'idle', durationMs: 3600 }
  ]);
});

test('task states interrupt the idle timeline', () => {
  assert.deepEqual(littleCompanionTimeline('working').map(item => item.pose), ['think']);
  assert.deepEqual(littleCompanionTimeline('success').map(item => item.pose), ['peace', 'dance']);
  assert.deepEqual(littleCompanionTimeline('error').map(item => item.pose), ['idle']);
});

test('calculates a wrapped frame from hand-checked sprite metadata', () => {
  assert.deepEqual(littleCompanionFrame('walk', 13), {
    pose: 'walk',
    src: '/pets/little-companion/walk.png',
    column: 1,
    row: 0,
    columns: 8,
    rows: 2,
    frameCount: 12,
    frameDelayMs: 60
  });
});

test('walks with the two source frames that keep the head fixed', () => {
  assert.deepEqual(
    (({ column, row, frameCount, frameDelayMs }) => ({ column, row, frameCount, frameDelayMs }))(littleCompanionWalkFrame(0)),
    { column: 6, row: 0, frameCount: 2, frameDelayMs: 120 }
  );
  assert.deepEqual(
    (({ column, row }) => ({ column, row }))(littleCompanionWalkFrame(3)),
    { column: 2, row: 1 }
  );
  assert.deepEqual(
    (({ column, row }) => ({ column, row }))(littleCompanionWalkFrame(11)),
    { column: 2, row: 1 }
  );
  assert.deepEqual(
    (({ column, row }) => ({ column, row }))(littleCompanionWalkFrame(12)),
    { column: 6, row: 0 }
  );
});

test('hands each walk foot to its closest idle frame instead of restarting idle at frame zero', () => {
  assert.equal(littleCompanionIdleFrameAfterWalk(0), 4);
  assert.equal(littleCompanionIdleFrameAfterWalk(1), 5);
  assert.equal(littleCompanionIdleFrameAfterWalk(2), 4);
});

test('falls back to idle for an unknown pose', () => {
  assert.equal(littleCompanionFrame('missing', 0).pose, 'idle');
});

test('wraps a negative frame to the final valid frame', () => {
  assert.deepEqual(
    (({ column, row }) => ({ column, row }))(littleCompanionFrame('idle', -1)),
    { column: 7, row: 2 }
  );
});

test('advances within a timeline and wraps without leaking the previous task state', () => {
  assert.equal(nextLittleCompanionStep('idle', 5), 0);
  assert.equal(nextLittleCompanionStep('success', 0), 1);
  assert.equal(nextLittleCompanionStep('working', 4), 0);
});

test('exposes the current pose step for the roaming controller', () => {
  assert.deepEqual(littleCompanionPoseStep('idle', 1, false), { pose: 'idle', durationMs: 3600 });
  assert.deepEqual(littleCompanionPoseStep('idle', 4, true), { pose: 'idle', durationMs: 3600 });
  assert.deepEqual(littleCompanionPoseStep('success', 1, false), { pose: 'dance', durationMs: 610 });
});
