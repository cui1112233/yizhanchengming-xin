import assert from 'node:assert/strict'
import test from 'node:test'

import { formatTimecode, materialFileURL, mediaTimeFromPlayhead, playheadFromMediaTime } from './playerModel.js'

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
