import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'ThemePicker.jsx'), 'utf8');

test('theme picker renders scheme swatch cards plus a depth segmented control', () => {
  assert.match(source, /listThemeSchemes/);
  assert.match(source, /listThemeModes/);
  assert.match(source, /className="theme-scheme-grid" role="radiogroup"/);
  assert.match(source, /className=\{`theme-scheme-card/);
  assert.match(source, /className="theme-scheme-swatch"/);
  assert.match(source, /--swatch-page/);
  assert.match(source, /--swatch-primary/);
  assert.match(source, /className="theme-segmented" role="radiogroup"/);
  assert.match(source, /className=\{`theme-segmented-option/);
  assert.match(source, /role="radio"/);
  assert.match(source, /aria-checked=\{selected\}/);
  assert.match(source, /\{scheme\.label\}/);
  assert.match(source, /\{mode\.label\}/);
});

test('theme picker reports a combined scheme-depth id', () => {
  assert.match(source, /<Popover/);
  assert.match(source, /getPopupContainer=\{\(\) => document\.body\}/);
  assert.match(source, /composeThemeId\(scheme\.id, activeMode\)/);
  assert.match(source, /composeThemeId\(activeScheme, mode\.id\)/);
  assert.match(source, /parseThemeId\(value\)/);
});
