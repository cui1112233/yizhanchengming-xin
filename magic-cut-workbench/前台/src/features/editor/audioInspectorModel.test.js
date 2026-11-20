import assert from 'node:assert/strict'
import test from 'node:test'

import { audioInspectorValues, patchAudioClip } from './audioInspectorModel.js'

const clip = { id: 'audio-1', durationMs: 1000, sourceStartMs: 0, speed: 1, volume: 0.7 }

test('audio inspector exposes the original volume speed and duration units', () => {
  assert.deepEqual(audioInspectorValues(clip), {
    volumePercent: 70,
    speed: 1,
    durationSeconds: 1,
    minDurationSeconds: 0.25,
    maxDurationSeconds: 10,
  })
})

test('zero percent is real mute and does not fall back to the default volume', () => {
  assert.deepEqual(patchAudioClip(clip, { volumePercent: 0 }), { volume: 0 })
})

test('changing speed keeps the selected source span and changes timeline duration', () => {
  assert.deepEqual(patchAudioClip(clip, { speed: 2 }), { speed: 2, durationMs: 500 })
})

test('changing duration derives the matching playback speed', () => {
  assert.deepEqual(patchAudioClip(clip, { durationSeconds: 2 }), { speed: 0.5, durationMs: 2000 })
})
