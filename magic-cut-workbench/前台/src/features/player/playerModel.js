function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value))
}

export function materialFileURL(projectID, materialID) {
  if (!projectID || !materialID) return ''
  return `/api/magic-cut/projects/${encodeURIComponent(projectID)}/materials/${encodeURIComponent(materialID)}/file`
}

export function playheadFromMediaTime(clip, mediaTimeSeconds) {
  if (!clip) return 0
  const speed = Number(clip.speed) > 0 ? Number(clip.speed) : 1
  const start = Number(clip.startMs) || 0
  const end = start + (Number(clip.durationMs) || 0)
  const sourceElapsed = (Number(mediaTimeSeconds) * 1000) - (Number(clip.sourceStartMs) || 0)
  return Math.round(clamp(start + (sourceElapsed / speed), start, end))
}

export function mediaTimeFromPlayhead(clip, playheadMs) {
  if (!clip) return 0
  const speed = Number(clip.speed) > 0 ? Number(clip.speed) : 1
  const start = Number(clip.startMs) || 0
  const end = start + (Number(clip.durationMs) || 0)
  const timelineTime = clamp(Number(playheadMs) || 0, start, end) - start
  return ((Number(clip.sourceStartMs) || 0) + (timelineTime * speed)) / 1000
}

export function formatTimecode(valueMs) {
  const totalSeconds = Math.max(0, Math.floor((Number(valueMs) || 0) / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  return [hours, minutes, seconds].map(value => String(value).padStart(2, '0')).join(':')
}

export function videoClipsFromTimeline(timeline) {
  return (timeline?.tracks || [])
    .filter(track => track.kind === 'video')
    .flatMap(track => track.clips || [])
    .filter(clip => clip?.id && clip?.materialId && Number(clip.durationMs) > 0)
    .sort((left, right) => (Number(left.startMs) || 0) - (Number(right.startMs) || 0) || String(left.id).localeCompare(String(right.id)))
}

export function nextVideoClip(clips, currentClip) {
  if (!currentClip) return null
  const ordered = [...(clips || [])].sort((left, right) => (Number(left.startMs) || 0) - (Number(right.startMs) || 0) || String(left.id).localeCompare(String(right.id)))
  const index = ordered.findIndex(clip => clip.id === currentClip.id)
  return index >= 0 && index + 1 < ordered.length ? ordered[index + 1] : null
}

export function playbackInstanceKey(clip) {
  return clip ? `${clip.id || 'clip'}:${clip.materialId || 'material'}` : 'empty-player'
}
