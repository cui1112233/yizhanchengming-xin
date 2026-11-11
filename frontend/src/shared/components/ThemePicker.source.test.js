import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'ThemePicker.jsx'), 'utf8');

test('theme picker renders a segmented control ordered light, dark, system', () => {
  assert.match(source, /listThemeDefinitions/);
  assert.match(source, /const segmentOrder = \['light', 'dark', 'system'\]/);
  assert.match(source, /className="theme-segmented" role="radiogroup"/);
  assert.match(source, /className=\{`theme-segmented-option/);
  assert.match(source, /role="radio"/);
  assert.match(source, /aria-checked=\{selected\}/);
  assert.match(source, /\{theme\.label\}/);
  // 卡片式预览（四色块 / 色带 / 勾选）必须彻底移除。
  assert.doesNotMatch(source, /theme-picker-option/);
  assert.doesNotMatch(source, /theme-picker-preview/);
  assert.doesNotMatch(source, /theme-picker-name/);
  assert.doesNotMatch(source, /<Check/);
});

test('theme picker portals its panel and reports the selected theme id', () => {
  assert.match(source, /<Popover/);
  assert.match(source, /getPopupContainer=\{\(\) => document\.body\}/);
  assert.match(source, /onChange\?\.\(theme\.id\)/);
});
