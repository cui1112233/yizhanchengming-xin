function finite(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function angle(value) {
  const normalized = ((value + 180) % 360 + 360) % 360 - 180;
  return normalized === -180 ? 180 : normalized;
}

export function normalizeTransform(value = {}) {
  return {
    x: clamp(finite(value.x, 0), -1, 1),
    y: clamp(finite(value.y, 0), -1, 1),
    scaleX: clamp(finite(value.scaleX, 1), 0.05, 20),
    scaleY: clamp(finite(value.scaleY, 1), 0.05, 20),
    rotation: angle(finite(value.rotation, 0)),
    opacity: clamp(finite(value.opacity, 1), 0, 1),
  };
}

export function dragTransform(value, delta, canvas) {
  const current = normalizeTransform(value);
  const width = Math.max(1, finite(canvas?.width, 1));
  const height = Math.max(1, finite(canvas?.height, 1));
  return normalizeTransform({ ...current, x: current.x + finite(delta?.x, 0) / width, y: current.y + finite(delta?.y, 0) / height });
}

export function resizeTransform(value, delta, canvas, options = {}) {
  const current = normalizeTransform(value);
  const width = Math.max(1, finite(canvas?.width, 1));
  const height = Math.max(1, finite(canvas?.height, 1));
  const factorX = 1 + finite(delta?.x, 0) / width;
  const factorY = 1 + finite(delta?.y, 0) / height;
  if (options.lockRatio) {
    const factor = Math.min(factorX, factorY);
    return normalizeTransform({ ...current, scaleX: current.scaleX * factor, scaleY: current.scaleY * factor });
  }
  return normalizeTransform({ ...current, scaleX: current.scaleX * factorX, scaleY: current.scaleY * factorY });
}

export function rotateTransform(value, startPoint, currentPoint, centerPoint) {
  const current = normalizeTransform(value);
  const start = Math.atan2(startPoint.y - centerPoint.y, startPoint.x - centerPoint.x);
  const end = Math.atan2(currentPoint.y - centerPoint.y, currentPoint.x - centerPoint.x);
  return normalizeTransform({ ...current, rotation: current.rotation + (end - start) * 180 / Math.PI });
}
