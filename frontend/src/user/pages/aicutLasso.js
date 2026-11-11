function finite(value) {
  const number = Number(value);
  return Number.isFinite(number) ? number : 0;
}

export function normalizeLassoRect(start, current) {
  const x1 = finite(start?.x);
  const y1 = finite(start?.y);
  const x2 = finite(current?.x);
  const y2 = finite(current?.y);
  const left = Math.min(x1, x2);
  const top = Math.min(y1, y2);
  const right = Math.max(x1, x2);
  const bottom = Math.max(y1, y2);
  return { left, top, width: right - left, height: bottom - top, right, bottom };
}

export function lassoGestureMoved(start, current, threshold = 5) {
  const rect = normalizeLassoRect(start, current);
  return Math.max(rect.width, rect.height) >= Math.max(1, finite(threshold));
}

export function nextLassoAutoScrollLeft(options = {}) {
  const current = Math.max(0, finite(options.scrollLeft));
  const viewportLeft = finite(options.viewportLeft);
  const viewportRight = finite(options.viewportRight);
  const viewportWidth = finite(options.viewportWidth);
  const scrollWidth = finite(options.scrollWidth);
  if (viewportRight <= viewportLeft || viewportWidth <= 0 || scrollWidth <= viewportWidth) return current;
  const edgeSize = Math.max(24, finite(options.edgeSize) || 80);
  const maximumStep = Math.max(1, finite(options.maximumStep) || 24);
  const maximum = Math.max(0, scrollWidth - viewportWidth);
  const clientX = finite(options.clientX);
  let delta = 0;
  if (clientX < viewportLeft + edgeSize) {
    const intensity = Math.min(1, Math.max(0, (viewportLeft + edgeSize - clientX) / edgeSize));
    delta = -Math.round(maximumStep * intensity);
  } else if (clientX > viewportRight - edgeSize) {
    const intensity = Math.min(1, Math.max(0, (clientX - (viewportRight - edgeSize)) / edgeSize));
    delta = Math.round(maximumStep * intensity);
  }
  return Math.min(maximum, Math.max(0, Math.round(current + delta)));
}

function overlaps(left, right) {
  return left.left < right.right && left.right > right.left && left.top < right.bottom && left.bottom > right.top;
}

export function selectLassoClipKeys(items, selection) {
  const keys = [];
  const seen = new Set();
  for (const item of items || []) {
    if (item?.locked || !item?.rect || !overlaps(item.rect, selection)) continue;
    for (const key of item.keys || []) {
      const value = String(key || '').trim();
      if (!value || seen.has(value)) continue;
      seen.add(value);
      keys.push(value);
    }
  }
  return keys;
}
