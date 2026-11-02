import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('complete source modal saves controlled textarea text instead of the Ant modal click event', () => {
  assert.match(source, /async function saveSourceEditor\(\)\s*\{\s*return saveCompleteSource\(sourceEditorBook, sourceEditorValue\);\s*\}/);
  assert.match(source, /onOk=\{\(\) => saveSourceEditor\(\)\}/);
});
