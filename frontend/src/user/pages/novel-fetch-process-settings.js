export const NOVEL_FETCH_PROCESS_SETTINGS_KEY = 'novel-fetch-v78-process-settings';

const DEFAULT_SETTINGS = {
  platform: 2,
  maxTxt: 4000,
  bookIds: '',
  textModel: '',
  selectedVersions: ['original', 'ai1'],
  rewriteMethods: { ai1: 'auto', ai2: 'auto', ai3: 'auto', ai4: 'auto', ai5: 'auto' },
  sensitiveAiProcessEnabled: true
};

function readStorageValue(storage, key) {
  if (!storage) return null;
  if (typeof storage.getItem === 'function') return storage.getItem(key);
  if (typeof storage.get === 'function') return storage.get(key);
  return null;
}

function writeStorageValue(storage, key, value) {
  if (!storage) return;
  if (typeof storage.setItem === 'function') {
    storage.setItem(key, value);
    return;
  }
  if (typeof storage.set === 'function') storage.set(key, value);
}

function getDefaultStorage() {
  return typeof window !== 'undefined' ? window.localStorage : null;
}

export function readNovelFetchProcessSettings(storage = getDefaultStorage()) {
  const raw = readStorageValue(storage, NOVEL_FETCH_PROCESS_SETTINGS_KEY);
  if (!raw) return { ...DEFAULT_SETTINGS, rewriteMethods: { ...DEFAULT_SETTINGS.rewriteMethods } };

  try {
    const saved = JSON.parse(raw);
    return {
      ...DEFAULT_SETTINGS,
      ...saved,
      bookIds: typeof saved.bookIds === 'string' ? saved.bookIds : '',
      selectedVersions: Array.isArray(saved.selectedVersions) ? saved.selectedVersions : [...DEFAULT_SETTINGS.selectedVersions],
      rewriteMethods: { ...DEFAULT_SETTINGS.rewriteMethods, ...(saved.rewriteMethods || {}) }
    };
  } catch {
    return { ...DEFAULT_SETTINGS, rewriteMethods: { ...DEFAULT_SETTINGS.rewriteMethods } };
  }
}

export function writeNovelFetchProcessSettings(settings, storage = getDefaultStorage()) {
  const normalized = { ...settings };
  writeStorageValue(storage, NOVEL_FETCH_PROCESS_SETTINGS_KEY, JSON.stringify(normalized));
  return normalized;
}
