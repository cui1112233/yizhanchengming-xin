import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./TextProductionSettingsForm.jsx', import.meta.url), 'utf8');

test('text settings form stays quiet for inherited fields and marks only real book overrides', () => {
  assert.match(source, /TEXT_PRODUCTION_SECTIONS\.map/);
  assert.match(source, /scope==='book'/);
  assert.doesNotMatch(source, /' · 继承'/);
  assert.match(source, /own && <small> · 本书覆盖<\/small>/);
  assert.match(source, /onRestore\(field\.key\)/);
});

test('text settings form renders all sections from the shared backend-facing field catalog', () => {
  assert.match(source, /section\.fields\.map/);
  assert.match(source, /\{\[field\.key\]:next\}/);
  assert.match(source, /items=\{TEXT_PRODUCTION_SECTIONS\.map/);
});
