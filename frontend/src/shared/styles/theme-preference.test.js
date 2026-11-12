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

test('theme cache migrates retired palettes to glass without changing depth', () => {
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'ocean-light' })), 'glass-light');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'graphite-dark' })), 'glass-dark');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'light' })), 'glass-light');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'dark' })), 'glass-dark');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'system' })), 'glass-system');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'mint' })), 'glass-light');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'pureblack' })), 'glass-dark');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'unsupported' })), 'glass-system');
  assert.equal(loadCachedTheme(memoryStorage()), 'glass-system');
});

test('persistCachedTheme normalises before writing', () => {
  const storage = memoryStorage();
  assert.equal(persistCachedTheme('violet-system', storage), 'glass-system');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'glass-system');
  assert.equal(persistCachedTheme('light', storage), 'glass-light');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'glass-light');
  assert.equal(persistCachedTheme('sakura', storage), 'glass-light');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'glass-light');
  assert.equal(persistCachedTheme('unsupported', storage), 'glass-system');
});

test('logged-in account preference overrides startup cache while guests keep the cache', () => {
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'graphite-dark', cachedThemeId: 'coral-light' }), 'glass-dark');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'dark', cachedThemeId: 'ocean-light' }), 'glass-dark');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'mint', cachedThemeId: 'ocean-light' }), 'glass-light');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'unsupported', cachedThemeId: 'ocean-light' }), 'glass-system');
  assert.equal(selectThemePreference({ loggedIn: false, accountThemeId: 'purple', cachedThemeId: 'violet-light' }), 'glass-light');
  assert.equal(selectThemePreference({ loggedIn: false, cachedThemeId: 'willow' }), 'glass-light');
  assert.equal(selectThemePreference({ loggedIn: false }), 'glass-system');
});
