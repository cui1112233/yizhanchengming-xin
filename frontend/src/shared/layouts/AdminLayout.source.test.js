import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'AdminLayout.jsx'), 'utf8');

test('admin layout reuses the account-backed multi-theme picker', () => {
  assert.match(source, /import \{ ThemePicker \}/);
  assert.match(source, /<ThemePicker value=\{theme\} onChange=\{changeTheme\}/);
  assert.match(source, /updateMemberProfile\(\{ themeId: nextThemeId \}\)/);
  assert.match(source, /selectThemePreference\(\{ loggedIn: true/);
});

test('admin layout resolves system theme before applying CSS and Ant tokens', () => {
  assert.match(source, /resolveThemeId\(theme, systemMode\)/);
  assert.match(source, /document\.documentElement\.dataset\.theme = resolvedTheme/);
  assert.match(source, /createAntTheme\(theme, systemMode\)/);
});
