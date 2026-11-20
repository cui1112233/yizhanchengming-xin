import assert from 'node:assert/strict'
import test from 'node:test'

import { formatTimecode, materialFileURL, mediaTimeFromPlayhead, nextVideoClip, playbackInstanceKey, playheadFromMediaTime, videoClipsFromTimeline } from './playerModel.js'

const clip = { startMs: 5000, durationMs: 4000, sourceStartMs: 2000, speed: 2 }

test('material playback URL keeps project and material identifiers inside the owned route', () => {
  assert.equal(materialFileURL('project one', 'material/one'), '/api/magic-cut/projects/project%20one/materials/material%2Fone/file')
})

test('player converts source media time into the matching timeline playhead', () => {
  assert.equal(playheadFromMediaTime(clip, 3.5), 5750)
  assert.equal(playheadFromMediaTime(clip, 99), 9000)
})

test('timeline seeking converts the playhead back into source media time', () => {
  assert.equal(mediaTimeFromPlayhead(clip, 8000), 8)
  assert.equal(mediaTimeFromPlayhead(clip, 0), 2)
})

test('time display stays understandable for long projects', () => {
  assert.equal(formatTimecode(0), '00:00:00')
  assert.equal(formatTimecode(3723000), '01:02:03')
})

test('continuous playback uses only video clips in timeline order', () => {
  const timeline = { tracks: [
    { kind: 'audio', clips: [{ id: 'audio-1', startMs: 0, durationMs: 9000 }] },
    { kind: 'video', clips: [
      { id: 'video-2', materialId: 'material-2', startMs: 3000, durationMs: 1000, sourceStartMs: 0, speed: 1 },
      { id: 'video-1', materialId: 'material-1', startMs: 0, durationMs: 1000, sourceStartMs: 0, speed: 1 },
    ] },
  ] }
  assert.deepEqual(videoClipsFromTimeline(timeline).map(item => item.id), ['video-1', 'video-2'])
})

test('continuous playback advances once and stops after the last clip', () => {
  const clips = [
    { id: 'video-1', startMs: 0, durationMs: 1000 },
    { id: 'video-2', startMs: 3000, durationMs: 1000 },
  ]
  assert.equal(nextVideoClip(clips, clips[0])?.id, 'video-2')
  assert.equal(nextVideoClip(clips, clips[1]), null)
})

test('two timeline clips using the same material still get separate playback instances', () => {
  assert.notEqual(
    playbackInstanceKey({ id: 'clip-1', materialId: 'same-material' }),
    playbackInstanceKey({ id: 'clip-2', materialId: 'same-material' }),
  )
})
