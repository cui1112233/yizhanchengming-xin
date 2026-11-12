import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import * as theme from './theme.js';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const schemeCSS = fs.readFileSync(path.join(dirname, 'theme-schemes.css'), 'utf8');

const expectedSchemeIds = ['coral', 'ocean', 'graphite', 'violet', 'glass'];
const expectedModes = ['light', 'dark', 'system'];

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
  'textOnAccent',
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

function cssBlock(selector) {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = schemeCSS.match(new RegExp(`${escaped}\\s*\\{([^}]*)\\}`, 's'));
  assert.ok(match, `missing scheme block ${selector}`);
  return match[1];
}

test('theme picker only exposes cinematic glass while old saved themes remain readable', () => {
  const schemes = theme.listThemeSchemes();
  assert.deepEqual(schemes.map(item => item.id), ['glass']);
  for (const scheme of schemes) {
    assert.ok(scheme.label, `${scheme.id} must have a visible label`);
    for (const mode of ['light', 'dark']) {
      const preview = scheme.preview[mode];
      assert.match(preview.page, /^#[0-9A-F]{6}$/i, `${scheme.id}.${mode}.page`);
      assert.match(preview.surface, /^#[0-9A-F]{6}$/i, `${scheme.id}.${mode}.surface`);
      assert.match(preview.primary, /^#[0-9A-F]{6}$/i, `${scheme.id}.${mode}.primary`);
    }
  }
});

test('theme modes keep light, dark and follow-system', () => {
  assert.deepEqual(theme.listThemeModes().map(mode => mode.id), expectedModes);
});

test('every scheme and mode resolves to a complete semantic token set', () => {
  for (const scheme of expectedSchemeIds) {
    for (const mode of ['light', 'dark']) {
      const definition = theme.themeDefinition(`${scheme}-${mode}`);
      assert.equal(definition.scheme, scheme);
      assert.equal(definition.colorScheme, mode);
      for (const token of requiredTokens) {
        assert.ok(definition.tokens[token], `${scheme}-${mode} must define ${token}`);
      }
    }
  }
});

test('legacy theme ids map to the coral scheme without changing depth', () => {
  const mapping = {
    light: 'coral-light',
    dark: 'coral-dark',
    system: 'coral-system',
    pureblack: 'coral-dark',
    purple: 'coral-dark',
    graywhite: 'coral-light',
    mint: 'coral-light',
    snow: 'coral-light',
    vanilla: 'coral-light',
    sakura: 'coral-light',
    willow: 'coral-light',
    ivory: 'coral-light',
    lotus: 'coral-light'
  };
  for (const [legacy, expected] of Object.entries(mapping)) {
    assert.equal(theme.normalizeThemeId(legacy), expected, `${legacy} must map to ${expected}`);
  }
  assert.equal(theme.normalizeThemeId('unknown-theme'), 'coral-system');
  assert.equal(theme.normalizeThemeId(''), 'coral-system');
  assert.equal(theme.normalizeThemeId('coral-dark'), 'coral-dark');
  assert.equal(theme.normalizeThemeId('violet-light'), 'violet-light');
  assert.equal(theme.normalizeThemeId('graphite'), 'graphite-system');
});

test('follow-system keeps the stored scheme while resolving depth', () => {
  assert.equal(theme.resolveThemeId('violet-system', 'light'), 'light');
  assert.equal(theme.resolveThemeId('violet-system', 'dark'), 'dark');
  assert.equal(theme.resolveThemeId('system', 'unexpected'), 'dark');
  assert.equal(theme.resolveThemeId('ocean-light', 'dark'), 'light');
  assert.equal(theme.resolveSchemeId('mint'), 'coral');
  assert.equal(theme.resolveSchemeId('graphite-dark'), 'graphite');
});

test('Ant Design tokens follow the selected scheme and mode', () => {
  const coralDark = theme.createAntTheme('coral-dark').token;
  assert.equal(coralDark.colorPrimary, '#F07167');
  assert.equal(coralDark.colorBgBase, '#12100E');
  assert.equal(coralDark.colorText, '#F5F0E8');
  assert.equal(coralDark.colorTextLightSolid, '#2B0F0C');

  const graphiteLight = theme.createAntTheme('graphite-light').token;
  assert.equal(graphiteLight.colorPrimary, '#2F5BEA');
  assert.equal(graphiteLight.colorBgBase, '#FAFAFA');
  assert.equal(graphiteLight.colorTextLightSolid, '#FFFFFF');

  assert.deepEqual(theme.createAntTheme('coral-system', 'light').token, theme.createAntTheme('coral-light').token);
  assert.deepEqual(theme.createAntTheme('system', 'dark').token, theme.createAntTheme('coral-dark').token);
});

test('stylesheets define a CSS block for every scheme and mode with the catalog primary', () => {
  for (const scheme of expectedSchemeIds) {
    for (const mode of ['light', 'dark']) {
      const block = cssBlock(`html[data-scheme='${scheme}'][data-theme='${mode}']`);
      const { tokens } = theme.themeDefinition(`${scheme}-${mode}`);
      assert.match(
        block,
        new RegExp(`--theme-primary:\\s*${tokens.primary}`, 'i'),
        `${scheme}-${mode} CSS primary must match the catalog`
      );
      assert.match(block, new RegExp(`--surface-base:\\s*${tokens.background}`, 'i'));
      assert.match(block, new RegExp(`--text-primary:\\s*${tokens.textPrimary}`, 'i'));
      assert.match(block, new RegExp(`--text-on-accent:\\s*${tokens.textOnAccent}`, 'i'));
    }
  }
});
