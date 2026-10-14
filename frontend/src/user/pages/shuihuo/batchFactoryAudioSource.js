export const AUDIO_SOURCE_TTS = 'tts';
export const AUDIO_SOURCE_ORIGINAL = 'original';

export function normalizeAudioSource(value) {
  return String(value || '').trim().toLowerCase() === AUDIO_SOURCE_ORIGINAL
    ? AUDIO_SOURCE_ORIGINAL
    : AUDIO_SOURCE_TTS;
}

export function normalizeAudioSourceSettings(settings = {}) {
  const source = normalizeAudioSource(settings.audioSource);
  if (source === AUDIO_SOURCE_ORIGINAL) {
    return {
      ...settings,
      audioSource: source,
      audioPlanningEnabled: true,
      audioMergeEnabled: false
    };
  }
  return { ...settings, audioSource: source, audioPlanningEnabled: true, audioMergeEnabled: true };
}
