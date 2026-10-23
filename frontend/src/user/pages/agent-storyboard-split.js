export const STORYBOARD_SPLIT_DEFAULTS = Object.freeze({ column: 39, row: 50 });

export function storyboardSplitRatio(value, minimum = 24) {
  const safeMinimum = Math.min(45, Math.max(10, Number(minimum) || 24));
  const numeric = Number(value);
  const candidate = Number.isFinite(numeric) ? numeric : 50;
  return Math.round(Math.max(safeMinimum, Math.min(100 - safeMinimum, candidate)));
}

export function storyboardSplitRatioFromPointer(pointer, bounds, axis) {
  const start = axis === 'x' ? Number(bounds?.left) : Number(bounds?.top);
  const size = axis === 'x' ? Number(bounds?.width) : Number(bounds?.height);
  if (!Number.isFinite(start) || !Number.isFinite(size) || size <= 0) return 50;
  return storyboardSplitRatio(((Number(pointer) - start) / size) * 100);
}
