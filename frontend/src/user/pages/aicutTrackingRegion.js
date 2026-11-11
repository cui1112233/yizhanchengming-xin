const MIN_TRACKING_REGION_SIZE = .05;

function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, Number(value) || 0));
}

function clean(value) {
  return Number(Number(value).toFixed(6));
}

export function normalizeTrackingRegion(value = {}) {
  const width = clean(clamp(value.width ?? .4, MIN_TRACKING_REGION_SIZE, 1));
  const height = clean(clamp(value.height ?? .4, MIN_TRACKING_REGION_SIZE, 1));
  return {
    centerX: clean(clamp(value.centerX ?? .5, width / 2, 1 - width / 2)),
    centerY: clean(clamp(value.centerY ?? .5, height / 2, 1 - height / 2)),
    width,
    height
  };
}

export function moveTrackingRegion(value, deltaX = 0, deltaY = 0) {
  const region = normalizeTrackingRegion(value);
  return normalizeTrackingRegion({
    ...region,
    centerX: region.centerX + Number(deltaX || 0),
    centerY: region.centerY + Number(deltaY || 0)
  });
}

export function resizeTrackingRegion(value, handle = 'se', deltaX = 0, deltaY = 0) {
  const region = normalizeTrackingRegion(value);
  let left = region.centerX - region.width / 2;
  let right = region.centerX + region.width / 2;
  let top = region.centerY - region.height / 2;
  let bottom = region.centerY + region.height / 2;
  const dx = Number(deltaX || 0);
  const dy = Number(deltaY || 0);

  if (handle.includes('w')) left = clamp(left + dx, 0, right - MIN_TRACKING_REGION_SIZE);
  if (handle.includes('e')) right = clamp(right + dx, left + MIN_TRACKING_REGION_SIZE, 1);
  if (handle.includes('n')) top = clamp(top + dy, 0, bottom - MIN_TRACKING_REGION_SIZE);
  if (handle.includes('s')) bottom = clamp(bottom + dy, top + MIN_TRACKING_REGION_SIZE, 1);

  return normalizeTrackingRegion({
    centerX: (left + right) / 2,
    centerY: (top + bottom) / 2,
    width: right - left,
    height: bottom - top
  });
}

export function drawTrackingRegion(start = {}, end = {}) {
  const startX = clamp(start.x, 0, 1);
  const startY = clamp(start.y, 0, 1);
  const endX = clamp(end.x, 0, 1);
  const endY = clamp(end.y, 0, 1);
  const left = Math.min(startX, endX);
  const right = Math.max(startX, endX);
  const top = Math.min(startY, endY);
  const bottom = Math.max(startY, endY);
  return normalizeTrackingRegion({
    centerX: (left + right) / 2,
    centerY: (top + bottom) / 2,
    width: Math.max(MIN_TRACKING_REGION_SIZE, right - left),
    height: Math.max(MIN_TRACKING_REGION_SIZE, bottom - top)
  });
}
