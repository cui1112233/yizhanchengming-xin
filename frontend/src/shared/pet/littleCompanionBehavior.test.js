import assert from 'node:assert/strict';
import test from 'node:test';
import {
  littleCompanionChooseAction,
  littleCompanionFatigueDelay,
  littleCompanionIdleDelay,
  littleCompanionInwardDistance,
  littleCompanionPeekDuration
} from './littleCompanionBehavior.js';

test('keeps idle and fatigue delays inside the approved natural ranges', () => {
  assert.equal(littleCompanionIdleDelay(0), 20_000);
  assert.equal(littleCompanionIdleDelay(0.5), 32_500);
  assert.equal(littleCompanionIdleDelay(1), 45_000);
  assert.equal(littleCompanionFatigueDelay(0), 120_000);
  assert.equal(littleCompanionFatigueDelay(0.5), 180_000);
  assert.equal(littleCompanionFatigueDelay(1), 240_000);
});

test('uses the approved lightweight action weights before fatigue', () => {
  const input = { activeMs: 60_000, fatigueMs: 180_000, durationRoll: 0 };
  assert.deepEqual(littleCompanionChooseAction({ ...input, actionRoll: 0.20 }), {
    kind: 'idle', pose: 'idle', durationMs: 20_000, resetFatigue: false
  });
  assert.deepEqual(littleCompanionChooseAction({ ...input, actionRoll: 0.60 }), {
    kind: 'walk', pose: 'walk', durationMs: 0, resetFatigue: false
  });
  assert.deepEqual(littleCompanionChooseAction({ ...input, actionRoll: 0.84 }), {
    kind: 'candy', pose: 'candy', durationMs: 4_800, resetFatigue: false
  });
  assert.deepEqual(littleCompanionChooseAction({ ...input, actionRoll: 0.90 }), {
    kind: 'dance', pose: 'dance', durationMs: 610, resetFatigue: false
  });
  assert.deepEqual(littleCompanionChooseAction({ ...input, actionRoll: 0.97 }), {
    kind: 'edge-walk', pose: 'walk', durationMs: 0, resetFatigue: false
  });
});

test('does not sleep or think before the fatigue threshold', () => {
  const action = littleCompanionChooseAction({
    activeMs: 179_999,
    fatigueMs: 180_000,
    actionRoll: 0.2,
    durationRoll: 0
  });
  assert.equal(action.kind, 'idle');
  assert.equal(action.resetFatigue, false);
});

test('rests only after fatigue and resets the fatigue clock', () => {
  assert.deepEqual(littleCompanionChooseAction({
    activeMs: 180_000,
    fatigueMs: 180_000,
    actionRoll: 0.2,
    durationRoll: 0
  }), {
    kind: 'sleep', pose: 'sleep', durationMs: 20_000, resetFatigue: true
  });
  assert.deepEqual(littleCompanionChooseAction({
    activeMs: 180_000,
    fatigueMs: 180_000,
    actionRoll: 0.8,
    durationRoll: 1
  }), {
    kind: 'think', pose: 'think', durationMs: 18_000, resetFatigue: true
  });
});

test('clamps invalid random input for peek and inward movement', () => {
  assert.equal(littleCompanionPeekDuration(-1), 3_000);
  assert.equal(littleCompanionPeekDuration(2), 6_000);
  assert.equal(littleCompanionInwardDistance(Number.NaN), 180);
  assert.equal(littleCompanionInwardDistance(1), 240);
});
