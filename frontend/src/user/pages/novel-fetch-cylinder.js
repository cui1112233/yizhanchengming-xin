function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

export function cylinderVisualState(distance) {
  const normalizedDistance = Number.isFinite(distance) ? distance : 0;
  const amount = Math.min(Math.abs(normalizedDistance), 3.2);

  return {
    rotateX: normalizedDistance === 0 ? 0 : clamp(normalizedDistance * -10, -30, 30),
    translateZ: 28 - amount * 18,
    scale: Math.max(0.9, 1 - amount * 0.035),
    opacity: Math.max(0.64, 1 - amount * 0.11),
    brightness: Math.max(0.84, 1 - amount * 0.05)
  };
}

export function nextCylinderIndex(length, currentIndex, direction) {
  if (!Number.isInteger(length) || length <= 0) return -1;
  const safeIndex = clamp(Number.isInteger(currentIndex) ? currentIndex : 0, 0, length - 1);
  const step = direction < 0 ? -1 : 1;
  return clamp(safeIndex + step, 0, length - 1);
}
