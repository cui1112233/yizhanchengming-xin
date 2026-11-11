function finite(value, fallback = 0) {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
}

function baseVolume(clip) {
  const value = finite(clip?.volume, 1);
  return value >= 0 && value <= 2 ? value : 1;
}

export function audioEnvelopePoints(clip) {
  const durationMs = Math.max(1, Math.round(finite(clip?.durationMs, 1)));
  const byOffset = new Map();
  for (const frame of clip?.keyframes || []) {
    const offsetMs = Math.round(finite(frame?.offsetMs, NaN));
    const volume = finite(frame?.values?.volume, NaN);
    if (!Number.isFinite(offsetMs) || offsetMs < 0 || offsetMs > durationMs || !Number.isFinite(volume) || volume < 0 || volume > 2) continue;
    byOffset.set(offsetMs, { id: String(frame.id || ''), offsetMs, position: offsetMs / durationMs, volume });
  }
  if (!byOffset.has(0)) byOffset.set(0, { id: '', offsetMs: 0, position: 0, volume: baseVolume(clip) });
  if (!byOffset.has(durationMs)) byOffset.set(durationMs, { id: '', offsetMs: durationMs, position: 1, volume: baseVolume(clip) });
  return [...byOffset.values()].sort((left, right) => left.offsetMs - right.offsetMs);
}

export function audioEnvelopeVolumeAt(clip, offsetMs) {
  const points = audioEnvelopePoints(clip);
  const at = Math.min(Math.max(0, finite(offsetMs)), Math.max(1, finite(clip?.durationMs, 1)));
  const rightIndex = points.findIndex(point => point.offsetMs >= at);
  if (rightIndex <= 0) return points[0].volume;
  const left = points[rightIndex - 1];
  const right = points[rightIndex];
  const span = right.offsetMs - left.offsetMs;
  if (span <= 0) return right.volume;
  return left.volume + (right.volume - left.volume) * ((at - left.offsetMs) / span);
}
