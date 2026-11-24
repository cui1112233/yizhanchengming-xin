import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./TextProductionSettingsForm.jsx', import.meta.url), 'utf8');

test('text settings form keeps inherited and book override markers in every section', () => {
  assert.match(source, /TEXT_PRODUCTION_SECTIONS\.map/);
  assert.match(source, /scope==='book'/);
  assert.match(source, /own\?' · 本书覆盖':' · 继承'/);
  assert.match(source, /onRestore\(field\.key\)/);
});

test('text settings form renders all sections from the shared backend-facing field catalog', () => {
  assert.match(source, /section\.fields\.map/);
  assert.match(source, /\{\[field\.key\]:next\}/);
  assert.match(source, /items=\{TEXT_PRODUCTION_SECTIONS\.map/);
});
