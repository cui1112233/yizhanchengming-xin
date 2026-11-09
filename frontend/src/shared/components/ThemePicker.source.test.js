import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'ThemePicker.jsx'), 'utf8');

test('theme picker renders the shared catalog with named four-color swatches', () => {
  assert.match(source, /listThemeDefinitions/);
  assert.match(source, /theme\.label/);
  assert.match(source, /tokens\.background/);
  assert.match(source, /tokens\.surface/);
  assert.match(source, /tokens\.primary/);
  assert.match(source, /tokens\.textPrimary/);
});

test('theme picker portals its panel and reports the selected theme id', () => {
  assert.match(source, /<Popover/);
  assert.match(source, /getPopupContainer=\{\(\) => document\.body\}/);
  assert.match(source, /onChange\?\.\(theme\.id\)/);
  assert.match(source, /aria-pressed=\{theme\.id === value\}/);
});
