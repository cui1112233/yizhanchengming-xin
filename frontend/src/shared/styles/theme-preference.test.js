import test from 'node:test';
import assert from 'node:assert/strict';
import {
  THEME_STORAGE_KEY,
  loadCachedTheme,
  persistCachedTheme,
  selectThemePreference
} from './theme-preference.js';

function memoryStorage(initial = {}) {
  const values = new Map(Object.entries(initial));
  return {
    getItem(key) { return values.has(key) ? values.get(key) : null; },
    setItem(key, value) { values.set(key, String(value)); }
  };
}

test('theme cache accepts light/dark and falls back to system for retired or invalid values', () => {
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'light' })), 'light');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'dark' })), 'dark');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'system' })), 'system');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'mint' })), 'system');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'pureblack' })), 'system');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'unsupported' })), 'system');
  assert.equal(loadCachedTheme(memoryStorage()), 'system');
});

test('persistCachedTheme stores only the three supported ids', () => {
  const storage = memoryStorage();
  assert.equal(persistCachedTheme('light', storage), 'light');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'light');
  assert.equal(persistCachedTheme('sakura', storage), 'system');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'system');
  assert.equal(persistCachedTheme('unsupported', storage), 'system');
});

test('logged-in account preference overrides startup cache while guests keep the cache', () => {
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'dark', cachedThemeId: 'light' }), 'dark');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'purple', cachedThemeId: 'light' }), 'system');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'unsupported', cachedThemeId: 'light' }), 'system');
  assert.equal(selectThemePreference({ loggedIn: false, accountThemeId: 'purple', cachedThemeId: 'light' }), 'light');
  assert.equal(selectThemePreference({ loggedIn: false, cachedThemeId: 'willow' }), 'system');
  assert.equal(selectThemePreference({ loggedIn: false }), 'system');
});
