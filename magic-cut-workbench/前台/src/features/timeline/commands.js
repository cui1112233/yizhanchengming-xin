function clone(value) {
  return structuredClone(value)
}

export function needsAutosave(dirtyVersion, savedDirtyVersion) {
  return dirtyVersion > savedDirtyVersion
}

function locate(state, clipId) {
  for (let trackIndex = 0; trackIndex < state.tracks.length; trackIndex += 1) {
    const clipIndex = state.tracks[trackIndex].clips.findIndex(clip => clip.id === clipId)
    if (clipIndex >= 0) return { trackIndex, clipIndex, clip: state.tracks[trackIndex].clips[clipIndex] }
  }
  throw new Error(`clip not found: ${clipId}`)
}

function validateClip(clip) {
  if (clip.startMs < 0 || clip.sourceStartMs < 0) throw new Error('clip bounds must be nonnegative')
  if (clip.durationMs <= 0 || clip.speed <= 0) throw new Error('clip duration and speed must be positive')
  if (clip.volume != null && (clip.volume < 0 || clip.volume > 2)) throw new Error('clip volume must stay between 0 and 2')
}

export function defaultClipVolume(kind) {
  return kind === 'audio' ? 0.7 : 1
}

export function executeCommand(current, command) {
  const before = clone(current)
  if (command.type === 'replace-state') return { state: clone(command.state), inverse: { type: 'replace-state', state: before } }
  const state = clone(current)
  if (command.type === 'insert') {
    const destination = state.tracks.find(track => track.id === command.trackId)
    if (!destination) throw new Error(`track not found: ${command.trackId}`)
    const inserted = clone(command.clip)
    validateClip(inserted)
    destination.clips.push(inserted)
    state.durationMs = Math.max(state.durationMs, inserted.startMs + inserted.durationMs)
    return { state, inverse: { type: 'replace-state', state: before } }
  }
  const found = locate(state, command.clipId)
  if (command.type === 'delete') {
    state.tracks[found.trackIndex].clips.splice(found.clipIndex, 1)
  } else if (command.type === 'duplicate') {
    const duplicate = { ...found.clip, id: command.newClipId, startMs: found.clip.startMs + (command.offsetMs || 0) }
    validateClip(duplicate)
    state.tracks[found.trackIndex].clips.splice(found.clipIndex + 1, 0, duplicate)
  } else if (command.type === 'split') {
    const relative = command.atMs - found.clip.startMs
    if (relative <= 0 || relative >= found.clip.durationMs) throw new Error('split point must stay inside clip')
    const trailing = {
      ...found.clip,
      id: command.newClipId,
      startMs: command.atMs,
      durationMs: found.clip.durationMs - relative,
      sourceStartMs: found.clip.sourceStartMs + relative * found.clip.speed,
    }
    found.clip.durationMs = relative
    state.tracks[found.trackIndex].clips.splice(found.clipIndex + 1, 0, trailing)
  } else if (command.type === 'move') {
    if (command.startMs < 0) throw new Error('clip bounds must be nonnegative')
    const destination = state.tracks.find(track => track.id === command.trackId)
    if (!destination) throw new Error(`track not found: ${command.trackId}`)
    const [moving] = state.tracks[found.trackIndex].clips.splice(found.clipIndex, 1)
    moving.startMs = command.startMs
    validateClip(moving)
    destination.clips.push(moving)
  } else if (command.type === 'resize') {
    found.clip.durationMs = command.durationMs
    validateClip(found.clip)
  } else {
    throw new Error(`unknown command: ${command.type}`)
  }
  return { state, inverse: { type: 'replace-state', state: before } }
}
