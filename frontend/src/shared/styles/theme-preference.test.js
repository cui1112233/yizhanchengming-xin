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

test('theme cache accepts old dark/light values and defaults missing or invalid values to system', () => {
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'light' })), 'light');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'dark' })), 'dark');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'mint' })), 'mint');
  assert.equal(loadCachedTheme(memoryStorage({ [THEME_STORAGE_KEY]: 'unsupported' })), 'system');
  assert.equal(loadCachedTheme(memoryStorage()), 'system');
});

test('persistCachedTheme stores only supported ids', () => {
  const storage = memoryStorage();
  assert.equal(persistCachedTheme('sakura', storage), 'sakura');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'sakura');
  assert.equal(persistCachedTheme('unsupported', storage), 'system');
  assert.equal(storage.getItem(THEME_STORAGE_KEY), 'system');
});

test('logged-in account preference overrides startup cache while guests keep the cache', () => {
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'purple', cachedThemeId: 'mint' }), 'purple');
  assert.equal(selectThemePreference({ loggedIn: true, accountThemeId: 'unsupported', cachedThemeId: 'mint' }), 'system');
  assert.equal(selectThemePreference({ loggedIn: false, accountThemeId: 'purple', cachedThemeId: 'mint' }), 'mint');
  assert.equal(selectThemePreference({ loggedIn: false, cachedThemeId: 'unsupported' }), 'system');
});
