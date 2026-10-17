function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

export function cylinderVisualState(distance) {
  const normalizedDistance = Number.isFinite(distance) ? distance : 0;
  const amount = Math.min(Math.abs(normalizedDistance), 3.2);

  return {
    rotateX: normalizedDistance === 0 ? 0 : clamp(normalizedDistance * -16, -48, 48),
    translateZ: 28 - amount * 32,
    scale: Math.max(0.78, 1 - amount * 0.075),
    opacity: Math.max(0.18, 1 - amount * 0.26),
    brightness: Math.max(0.52, 1 - amount * 0.15)
  };
}

export function nextCylinderIndex(length, currentIndex, direction) {
  if (!Number.isInteger(length) || length <= 0) return -1;
  const safeIndex = clamp(Number.isInteger(currentIndex) ? currentIndex : 0, 0, length - 1);
  const step = direction < 0 ? -1 : 1;
  return clamp(safeIndex + step, 0, length - 1);
}
