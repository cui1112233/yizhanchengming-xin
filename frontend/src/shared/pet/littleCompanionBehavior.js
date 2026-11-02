const clamp01 = (value, fallback = 0.5) => {
  const number = Number(value);
  if (!Number.isFinite(number)) return fallback;
  return Math.max(0, Math.min(1, number));
};

const ranged = (minimum, maximum, randomValue) => (
  Math.round(minimum + (maximum - minimum) * clamp01(randomValue))
);

export const littleCompanionIdleDelay = randomValue => ranged(20_000, 45_000, randomValue);

export const littleCompanionFatigueDelay = randomValue => ranged(120_000, 240_000, randomValue);

export const littleCompanionPeekDuration = randomValue => ranged(3_000, 6_000, randomValue);

export const littleCompanionInwardDistance = randomValue => ranged(120, 240, randomValue);

export function littleCompanionChooseAction({
  activeMs = 0,
  fatigueMs = 180_000,
  actionRoll = 0.5,
  durationRoll = 0.5
} = {}) {
  const elapsed = Math.max(0, Number.isFinite(Number(activeMs)) ? Number(activeMs) : 0);
  const fatigue = Math.max(0, Number.isFinite(Number(fatigueMs)) ? Number(fatigueMs) : 180_000);
  const choice = clamp01(actionRoll);

  if (elapsed >= fatigue) {
    return choice < 0.6
      ? {
          kind: 'sleep',
          pose: 'sleep',
          durationMs: ranged(20_000, 35_000, durationRoll),
          resetFatigue: true
        }
      : {
          kind: 'think',
          pose: 'think',
          durationMs: ranged(10_000, 18_000, durationRoll),
          resetFatigue: true
        };
  }

  if (choice < 0.45) {
    return {
      kind: 'idle',
      pose: 'idle',
      durationMs: littleCompanionIdleDelay(durationRoll),
      resetFatigue: false
    };
  }
  if (choice < 0.80) return { kind: 'walk', pose: 'walk', durationMs: 0, resetFatigue: false };
  if (choice < 0.88) return { kind: 'candy', pose: 'candy', durationMs: 4_800, resetFatigue: false };
  if (choice < 0.93) return { kind: 'dance', pose: 'dance', durationMs: 610, resetFatigue: false };
  return { kind: 'edge-walk', pose: 'walk', durationMs: 0, resetFatigue: false };
}
