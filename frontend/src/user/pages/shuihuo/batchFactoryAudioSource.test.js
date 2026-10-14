import test from 'node:test';
import assert from 'node:assert/strict';

import { normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource.js';

test('defaults missing audio source to TTS', () => {
  assert.equal(normalizeAudioSource(undefined), 'tts');
  assert.equal(normalizeAudioSource('unexpected'), 'tts');
});

test('original audio keeps TTS planning but disables only final TTS merge', () => {
  assert.deepEqual(normalizeAudioSourceSettings({ audioSource: 'original', audioPlanningEnabled: true, audioMergeEnabled: true, audioDurationSeconds: 42, audioDurationFingerprint: 'fp', audioDurationManual: true }), {
    audioSource: 'original',
    audioPlanningEnabled: true,
    audioMergeEnabled: false,
    audioDurationSeconds: 42,
    audioDurationFingerprint: 'fp',
    audioDurationManual: true
  });
});

test('TTS audio enables final audio merge and always keeps the shared timing flow', () => {
  assert.deepEqual(normalizeAudioSourceSettings({ audioSource: 'tts', audioPlanningEnabled: false, audioMergeEnabled: false }), {
    audioSource: 'tts',
    audioPlanningEnabled: true,
    audioMergeEnabled: true
  });
});
