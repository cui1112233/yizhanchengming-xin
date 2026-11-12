import { parseThemeId } from './theme-catalog.js';

export const THEME_STORAGE_KEY = 'yizhan-theme';

// Retired palettes keep their depth preference while adopting the unified material.
const currentThemeId = value => `glass-${parseThemeId(value).mode}`;
export function loadCachedTheme(storage = globalThis.localStorage) {
  try {
    return currentThemeId(storage?.getItem(THEME_STORAGE_KEY));
  } catch {
    return currentThemeId('system');
  }
}

export function persistCachedTheme(themeId, storage = globalThis.localStorage) {
  const nextThemeId = currentThemeId(themeId);
  try {
    storage?.setItem(THEME_STORAGE_KEY, nextThemeId);
  } catch {
    // The visible selection still works when private browsing blocks storage.
  }
  return nextThemeId;
}

export function selectThemePreference({ loggedIn, accountThemeId, cachedThemeId }) {
  return currentThemeId(loggedIn ? accountThemeId : cachedThemeId);
}
