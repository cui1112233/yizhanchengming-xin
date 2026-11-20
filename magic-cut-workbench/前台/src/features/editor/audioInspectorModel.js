function rounded(value, digits = 2) {
  const factor = 10 ** digits
  return Math.round(value * factor) / factor
}

function sourceSpanMs(clip) {
  return Math.max(1, clip.durationMs * clip.speed)
}

export function audioInspectorValues(clip) {
  const sourceMs = sourceSpanMs(clip)
  return {
    volumePercent: Math.round((clip.volume ?? 0.7) * 100),
    speed: rounded(clip.speed, 1),
    durationSeconds: rounded(clip.durationMs / 1000),
    minDurationSeconds: rounded(sourceMs / 4 / 1000),
    maxDurationSeconds: rounded(sourceMs / 0.1 / 1000),
  }
}

export function patchAudioClip(clip, change) {
  if (change.volumePercent != null) {
    if (change.volumePercent < 0 || change.volumePercent > 200) throw new Error('volume must stay between 0 and 200')
    return { volume: rounded(change.volumePercent / 100) }
  }
  if (change.speed != null) {
    if (change.speed < 0.1 || change.speed > 4) throw new Error('speed must stay between 0.1 and 4')
    return { speed: rounded(change.speed, 1), durationMs: Math.max(1, Math.round(sourceSpanMs(clip) / change.speed)) }
  }
  if (change.durationSeconds != null) {
    const durationMs = Math.round(change.durationSeconds * 1000)
    const speed = sourceSpanMs(clip) / durationMs
    if (durationMs <= 0 || speed < 0.1 || speed > 4) throw new Error('duration exceeds the original speed range')
    return { speed: rounded(speed, 1), durationMs }
  }
  throw new Error('audio property change is required')
}
