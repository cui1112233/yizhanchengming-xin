export const AUDIO_SOURCE_TTS = 'tts';
export const AUDIO_SOURCE_ORIGINAL = 'original';
export const SD_ORIGINAL_STORY_VERSION = 'sd.original-story.v1';

export function normalizeAudioSource(value) {
  return String(value || '').trim().toLowerCase() === AUDIO_SOURCE_ORIGINAL
    ? AUDIO_SOURCE_ORIGINAL
    : AUDIO_SOURCE_TTS;
}

// Audio source controls only the final mix. New production flows always build
// the shared TTS/line-duration timeline; only a historical snapshot that
// explicitly stored false keeps the legacy semantic-timeline behavior.
export function audioPlanningRequired(settings = {}) {
  if (String(settings?.storyTimelineVersion || '').trim() === SD_ORIGINAL_STORY_VERSION) return true;
  if (settings?.audioSource === AUDIO_SOURCE_ORIGINAL || settings?.audioSource === AUDIO_SOURCE_TTS) return true;
  return settings?.audioPlanningEnabled !== false;
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
