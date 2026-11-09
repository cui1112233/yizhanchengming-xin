import test from 'node:test';
import assert from 'node:assert/strict';
import * as theme from './theme.js';

const expectedThemeIds = [
  'system',
  'dark',
  'light',
  'graywhite',
  'mint',
  'snow',
  'vanilla',
  'sakura',
  'pureblack',
  'purple',
  'willow',
  'ivory',
  'lotus'
];

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

test('theme catalog exposes system plus the twelve approved visual themes', () => {
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

test('theme ids keep legacy choices and reject unsupported stored values', () => {
  assert.equal(theme.normalizeThemeId('light'), 'light');
  assert.equal(theme.normalizeThemeId('dark'), 'dark');
  assert.equal(theme.normalizeThemeId('system'), 'system');
  assert.equal(theme.normalizeThemeId('unknown-theme'), 'dark');
  assert.equal(theme.normalizeThemeId(''), 'dark');
});

test('system theme resolves without replacing the stored user choice', () => {
  assert.equal(theme.resolveThemeId('system', 'light'), 'light');
  assert.equal(theme.resolveThemeId('system', 'dark'), 'dark');
  assert.equal(theme.resolveThemeId('system', 'unexpected'), 'dark');
  assert.equal(theme.resolveThemeId('mint', 'dark'), 'mint');
  assert.equal(theme.themeDefinition('system', 'light').id, 'light');
});

test('Ant Design tokens are derived from the selected semantic theme', () => {
  const pureBlack = theme.createAntTheme('pureblack').token;
  assert.equal(pureBlack.colorBgBase, '#000000');
  assert.equal(pureBlack.colorText, '#f7f8fb');
  assert.equal(pureBlack.colorPrimary, '#63dfc0');

  const mint = theme.createAntTheme('mint').token;
  assert.equal(mint.colorBgBase, '#eef8f2');
  assert.equal(mint.colorText, '#17372d');
  assert.equal(mint.colorPrimary, '#16866c');
});
