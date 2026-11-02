export const STORYBOARD_SPLIT_DEFAULTS = Object.freeze({
  director: 31,
  column: 25,
  leftRow: 66,
  rightRow: 55
});

export function storyboardSplitRatio(value, minimum = 24, fallback = 50) {
  const safeMinimum = Math.min(45, Math.max(10, Number(minimum) || 24));
  const numeric = Number(value);
  const fallbackNumeric = Number(fallback);
  const candidate = Number.isFinite(numeric) ? numeric : (Number.isFinite(fallbackNumeric) ? fallbackNumeric : 50);
  return Math.round(Math.max(safeMinimum, Math.min(100 - safeMinimum, candidate)));
}

export function storyboardSplitRatioFromPointer(pointer, bounds, axis, minimum = 24, fallback = 50) {
  const start = axis === 'x' ? Number(bounds?.left) : Number(bounds?.top);
  const size = axis === 'x' ? Number(bounds?.width) : Number(bounds?.height);
  if (!Number.isFinite(start) || !Number.isFinite(size) || size <= 0) return storyboardSplitRatio(undefined, minimum, fallback);
  return storyboardSplitRatio(((Number(pointer) - start) / size) * 100, minimum, fallback);
}
