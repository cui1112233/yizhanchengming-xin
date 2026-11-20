import assert from 'node:assert/strict'
import test from 'node:test'

import { defaultClipVolume, executeCommand, needsAutosave } from './commands.js'

const original = {
  revision: 2,
  durationMs: 8000,
  tracks: [
    { id: 'video-1', kind: 'video', clips: [{ id: 'clip-1', materialId: 'mcm-1', startMs: 1000, durationMs: 4000, sourceStartMs: 0, speed: 1 }] },
    { id: 'audio-1', kind: 'audio', clips: [] },
  ],
}

test('ordinary audio uses original 70 percent default while video original sound stays at 100 percent', () => {
  assert.equal(defaultClipVolume('audio'), 0.7)
  assert.equal(defaultClipVolume('video'), 1)
})

test('clip volume accepts mute and rejects values outside the original 0 to 200 percent range', () => {
  const muted = executeCommand(original, { type: 'insert', trackId: 'audio-1', clip: { id: 'muted', materialId: 'audio', startMs: 0, durationMs: 1000, sourceStartMs: 0, speed: 1, volume: 0 } })
  assert.equal(muted.state.tracks[1].clips[0].volume, 0)
  assert.throws(() => executeCommand(original, { type: 'insert', trackId: 'audio-1', clip: { id: 'too-loud', materialId: 'audio', startMs: 0, durationMs: 1000, sourceStartMs: 0, speed: 1, volume: 2.1 } }), /volume/)
})

for (const [name, command] of [
	['insert', { type: 'insert', trackId: 'video-1', clip: { id: 'clip-2', materialId: 'mcm-2', startMs: 5000, durationMs: 1000, sourceStartMs: 0, speed: 1 } }],
  ['delete', { type: 'delete', clipId: 'clip-1' }],
  ['duplicate', { type: 'duplicate', clipId: 'clip-1', newClipId: 'clip-2', offsetMs: 250 }],
  ['split', { type: 'split', clipId: 'clip-1', atMs: 2500, newClipId: 'clip-2' }],
  ['move', { type: 'move', clipId: 'clip-1', trackId: 'audio-1', startMs: 300 }],
  ['resize', { type: 'resize', clipId: 'clip-1', durationMs: 1500 }],
	['update-properties', { type: 'update-properties', clipId: 'clip-1', patch: { volume: 0.4, speed: 2, durationMs: 2000 } }],
]) {
  test(`${name} command returns an exact inverse`, () => {
    const result = executeCommand(original, command)
    assert.notDeepEqual(result.state, original)
    const restored = executeCommand(result.state, result.inverse)
    assert.deepEqual(restored.state, original)
  })
}

test('property update rejects fields outside the audio contract', () => {
  assert.throws(() => executeCommand(original, { type: 'update-properties', clipId: 'clip-1', patch: { startMs: 0 } }), /property/)
})

test('commands reject negative clip bounds', () => {
  assert.throws(() => executeCommand(original, { type: 'move', clipId: 'clip-1', trackId: 'video-1', startMs: -1 }), /nonnegative/)
  assert.throws(() => executeCommand(original, { type: 'resize', clipId: 'clip-1', durationMs: 0 }), /positive/)
})

test('autosave runs only for edits newer than the last persisted edit', () => {
  assert.equal(needsAutosave(0, 0), false)
  assert.equal(needsAutosave(2, 2), false)
  assert.equal(needsAutosave(3, 2), true)
})
