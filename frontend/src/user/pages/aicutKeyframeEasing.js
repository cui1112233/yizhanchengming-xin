const clamp01 = value => Math.min(1, Math.max(0, Number.isFinite(Number(value)) ? Number(value) : 0));

export const keyframeEasingPresets = [
  { id: 'linear', label: '线性', x1: 0, y1: 0, x2: 1, y2: 1 },
  { id: 'ease-in', label: '缓入', x1: .42, y1: 0, x2: 1, y2: 1 },
  { id: 'ease-out', label: '缓出', x1: 0, y1: 0, x2: .58, y2: 1 },
  { id: 'ease-in-out', label: '缓入缓出', x1: .42, y1: 0, x2: .58, y2: 1 },
  { id: 'hold', label: '定格', x1: 0, y1: 0, x2: 1, y2: 0 },
  { id: 'custom', label: '自定义', x1: .25, y1: .1, x2: .25, y2: 1 }
];

export function normalizeKeyframeEasing(easing) {
  const preset = keyframeEasingPresets.find(item => item.id === String(easing?.type || 'linear')) || keyframeEasingPresets[0];
  if (preset.id !== 'custom') return { type: preset.id, x1: preset.x1, y1: preset.y1, x2: preset.x2, y2: preset.y2 };
  return { type: 'custom', x1: clamp01(easing?.x1 ?? preset.x1), y1: clamp01(easing?.y1 ?? preset.y1), x2: clamp01(easing?.x2 ?? preset.x2), y2: clamp01(easing?.y2 ?? preset.y2) };
}

function cubicCoordinate(control1, control2, ratio) {
  const inverse = 1 - ratio;
  return 3 * inverse * inverse * ratio * control1 + 3 * inverse * ratio * ratio * control2 + ratio * ratio * ratio;
}

export function keyframeEasingSamples(easing, steps = 8) {
  const normalized = normalizeKeyframeEasing(easing);
  if (normalized.type === 'hold') return [{ x: 0, y: 0 }, { x: 1, y: 1 }];
  return Array.from({ length: steps + 1 }, (_, index) => {
    const ratio = index / steps;
    return { x: cubicCoordinate(normalized.x1, normalized.x2, ratio), y: cubicCoordinate(normalized.y1, normalized.y2, ratio) };
  });
}

export function sampleKeyframeEasing(easing, progress) {
  const requested = clamp01(progress);
  const normalized = normalizeKeyframeEasing(easing);
  if (normalized.type === 'hold') return requested >= 1 ? 1 : 0;
  const samples = keyframeEasingSamples(normalized);
  const rightIndex = samples.findIndex(point => point.x >= requested);
  if (rightIndex <= 0) return samples[0].y;
  const left = samples[rightIndex - 1];
  const right = samples[rightIndex];
  const span = right.x - left.x;
  return clamp01(left.y + (right.y - left.y) * (span > 0 ? (requested - left.x) / span : 0));
}
