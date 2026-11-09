import { listThemeDefinitions } from './theme-catalog.js';

export const THEME_STORAGE_KEY = 'yizhan-theme';

const supportedThemeIds = new Set(listThemeDefinitions().map(theme => theme.id));

function supportedThemeId(value) {
  const themeId = typeof value === 'string' ? value.trim().toLowerCase() : '';
  return supportedThemeIds.has(themeId) ? themeId : 'system';
}

export function loadCachedTheme(storage = globalThis.localStorage) {
  try {
    return supportedThemeId(storage?.getItem(THEME_STORAGE_KEY));
  } catch {
    return 'system';
  }
}

export function persistCachedTheme(themeId, storage = globalThis.localStorage) {
  const nextThemeId = supportedThemeId(themeId);
  try {
    storage?.setItem(THEME_STORAGE_KEY, nextThemeId);
  } catch {
    // The visible selection still works when private browsing blocks storage.
  }
  return nextThemeId;
}

export function selectThemePreference({ loggedIn, accountThemeId, cachedThemeId }) {
  return supportedThemeId(loggedIn ? accountThemeId : cachedThemeId);
}
