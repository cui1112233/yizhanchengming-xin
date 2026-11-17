const KEY = 'v88-window-bg-appearance-v2';
const DEFAULTS = Object.freeze({ blur: 42, brightness: 50 });

export function normalizeWindowBackgroundAppearance(value = {}) {
  const bounded = (key, min, max) => {
    const raw = value?.[key];
    const number = raw === null || raw === '' || raw === undefined ? NaN : Number(raw);
    return Number.isFinite(number) ? Math.round(Math.max(min, Math.min(max, number))) : DEFAULTS[key];
  };
  return { blur: bounded('blur', 0, 42), brightness: bounded('brightness', 18, 85) };
}

export function readWindowBackgroundAppearance() {
  try {
    const saved = window.localStorage.getItem(KEY);
    if (saved) return normalizeWindowBackgroundAppearance(JSON.parse(saved));
    const legacy = JSON.parse(window.localStorage.getItem('v88-window-bg-appearance'));
    return normalizeWindowBackgroundAppearance({ blur: legacy?.blur, brightness: 50 });
  }
  catch { return { ...DEFAULTS }; }
}

export function applyWindowBackgroundAppearance(value) {
  const settings = normalizeWindowBackgroundAppearance(value);
  if (typeof document !== 'undefined') {
    document.documentElement.style.setProperty('--window-video-blur', `${settings.blur}px`);
    document.documentElement.style.setProperty('--window-video-brightness', `${settings.brightness / 100}`);
  }
  return settings;
}

export function saveWindowBackgroundAppearance(value) {
  const settings = applyWindowBackgroundAppearance(value);
  try { window.localStorage.setItem(KEY, JSON.stringify(settings)); } catch { /* Still usable without storage. */ }
  return settings;
}

saveWindowBackgroundAppearance(readWindowBackgroundAppearance());
if (typeof window !== 'undefined') {
  window.addEventListener('storage', event => {
    if (event.key === KEY) applyWindowBackgroundAppearance(readWindowBackgroundAppearance());
  });
}
