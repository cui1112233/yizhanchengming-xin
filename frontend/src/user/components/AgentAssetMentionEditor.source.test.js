import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const sourcePath = new URL('./AgentAssetMentionEditor.jsx', import.meta.url);

test('renders a content-editable prompt with atomic asset chips', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /contentEditable=\{!disabled\}/);
  assert.match(source, /chip\.contentEditable = 'false'/);
  assert.match(source, /promptMentionRanges\(value, references\)/);
  assert.match(source, /chip\.dataset\.assetMentionValue/);
  assert.match(source, /agent-asset-mention-editor/);
});

test('keeps editing compatible with IME, plain-text paste, and caret restoration', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /onCompositionStart/);
  assert.match(source, /onCompositionEnd/);
  assert.match(source, /clipboardData\.getData\('text\/plain'\)/);
  assert.match(source, /restoreOffsetRef/);
  assert.match(source, /setSelectionOffset/);
});

test('serializes canonical prompt text without layout-generated chip newlines', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /plainTextFromEditor/);
  assert.match(source, /before\.cloneContents\(\)/);
  assert.doesNotMatch(source, /root\.innerText/);
});

test('preserves the live caret when reference updates redraw the same prompt', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /const activeOffset = root\.contains\(window\.getSelection\(\)\?\.anchorNode\)/);
  assert.match(source, /activeOffset/);
});

test('removes a whole referenced asset from backspace or delete boundaries', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /\['Backspace', 'Delete'\]/);
  assert.match(source, /assetMentionDeletionAtCaret/);
  assert.match(source, /assetMentionAtDOMCaret/);
  assert.match(source, /removeAssetMention\(value, boundary\.asset\)/);
  assert.match(source, /onRemoveReference\?\.\(boundary\.asset, nextValue\)/);
});

test('uses image thumbnails with a non-text type fallback', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /asset\?\.kind === 'image'/);
  assert.match(source, /agent-asset-mention-chip__fallback/);
  assert.match(source, /fallback\.dataset\.kind/);
  assert.match(source, /image\.addEventListener\('error'/);
});
