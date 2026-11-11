import test from 'node:test';
import assert from 'node:assert/strict';
import * as theme from './theme.js';

const expectedThemeIds = ['system', 'dark', 'light'];

const requiredTokens = [
  'background',
  'surface',
  'surfaceElevated',
  'textPrimary',
  'textSecondary',
  'textMuted',
  'border',
  'borderStrong',
  'primary',
  'primaryHover',
  'primaryActive',
  'success',
  'warning',
  'danger',
  'info',
  'canvasBackground',
  'canvasGrid',
  'nodeSurface',
  'edgeColor',
  'viewportBackground',
  'timelineBackground'
];

test('theme catalog exposes system plus the two approved visual themes', () => {
  const definitions = theme.listThemeDefinitions();
  assert.deepEqual(definitions.map(item => item.id), expectedThemeIds);
  for (const definition of definitions) {
    assert.ok(definition.label, `${definition.id} must have a visible label`);
    assert.ok(['light', 'dark', 'system'].includes(definition.colorScheme));
    for (const token of requiredTokens) {
      assert.ok(definition.tokens[token], `${definition.id} must define ${token}`);
    }
  }
});

test('retired palette ids fall back to system instead of leaking an unknown theme', () => {
  for (const retired of ['graywhite', 'mint', 'snow', 'vanilla', 'sakura', 'pureblack', 'purple', 'willow', 'ivory', 'lotus']) {
    assert.equal(theme.normalizeThemeId(retired), 'system', `${retired} must fall back to system`);
    assert.equal(theme.resolveThemeId(retired, 'light'), 'light', `${retired} must resolve through system`);
    assert.equal(theme.resolveThemeId(retired, 'dark'), 'dark', `${retired} must resolve through system`);
  }
  assert.equal(theme.themeDefinition('pureblack', 'light').id, 'light');
});

test('theme ids keep legacy choices and reject unsupported stored values', () => {
  assert.equal(theme.normalizeThemeId('light'), 'light');
  assert.equal(theme.normalizeThemeId('dark'), 'dark');
  assert.equal(theme.normalizeThemeId('system'), 'system');
  assert.equal(theme.normalizeThemeId('unknown-theme'), 'system');
  assert.equal(theme.normalizeThemeId(''), 'system');
});

test('system theme resolves without replacing the stored user choice', () => {
  assert.equal(theme.resolveThemeId('system', 'light'), 'light');
  assert.equal(theme.resolveThemeId('system', 'dark'), 'dark');
  assert.equal(theme.resolveThemeId('system', 'unexpected'), 'dark');
  assert.equal(theme.themeDefinition('system', 'light').id, 'light');
});

test('Ant Design tokens are derived from the selected semantic theme', () => {
  const dark = theme.createAntTheme('dark').token;
  assert.equal(dark.colorBgBase, '#070b11');
  assert.equal(dark.colorText, '#f5f7fb');
  assert.equal(dark.colorPrimary, '#f07167');

  const light = theme.createAntTheme('light').token;
  assert.equal(light.colorBgBase, '#f4f6fa');
  assert.equal(light.colorText, '#172033');
  assert.equal(light.colorPrimary, '#f07167');

  assert.deepEqual(theme.createAntTheme('system', 'light').token, light);
  assert.deepEqual(theme.createAntTheme('system', 'dark').token, dark);
});
