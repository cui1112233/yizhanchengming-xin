import assert from 'node:assert/strict'
import test from 'node:test'

import { executeCommand, needsAutosave } from './commands.js'

const original = {
  revision: 2,
  durationMs: 8000,
  tracks: [
    { id: 'video-1', kind: 'video', clips: [{ id: 'clip-1', materialId: 'mcm-1', startMs: 1000, durationMs: 4000, sourceStartMs: 0, speed: 1 }] },
    { id: 'audio-1', kind: 'audio', clips: [] },
  ],
}

for (const [name, command] of [
	['insert', { type: 'insert', trackId: 'video-1', clip: { id: 'clip-2', materialId: 'mcm-2', startMs: 5000, durationMs: 1000, sourceStartMs: 0, speed: 1 } }],
  ['delete', { type: 'delete', clipId: 'clip-1' }],
  ['duplicate', { type: 'duplicate', clipId: 'clip-1', newClipId: 'clip-2', offsetMs: 250 }],
  ['split', { type: 'split', clipId: 'clip-1', atMs: 2500, newClipId: 'clip-2' }],
  ['move', { type: 'move', clipId: 'clip-1', trackId: 'audio-1', startMs: 300 }],
  ['resize', { type: 'resize', clipId: 'clip-1', durationMs: 1500 }],
]) {
  test(`${name} command returns an exact inverse`, () => {
    const result = executeCommand(original, command)
    assert.notDeepEqual(result.state, original)
    const restored = executeCommand(result.state, result.inverse)
    assert.deepEqual(restored.state, original)
  })
}

test('commands reject negative clip bounds', () => {
  assert.throws(() => executeCommand(original, { type: 'move', clipId: 'clip-1', trackId: 'video-1', startMs: -1 }), /nonnegative/)
  assert.throws(() => executeCommand(original, { type: 'resize', clipId: 'clip-1', durationMs: 0 }), /positive/)
})

test('autosave runs only for edits newer than the last persisted edit', () => {
  assert.equal(needsAutosave(0, 0), false)
  assert.equal(needsAutosave(2, 2), false)
  assert.equal(needsAutosave(3, 2), true)
})
