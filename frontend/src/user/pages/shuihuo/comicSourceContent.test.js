import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { comicSourceBoundaryOptions, comicSourceSplitPayload, comicSourceStatusLabel } from './comicSourceContent.js';

test('comic source boundaries become readable safe split options', () => {
  const options = comicSourceBoundaryOptions({
    startOffset: 10,
    sourceText: '第一句完整。第二句也完整。',
    safeBoundaries: [{ offset: 28, type: 'sentence', preview: '第一句完整。' }]
  });
  assert.equal(options.length, 1);
  assert.equal(options[0].value, 28);
  assert.match(options[0].label, /第一句完整/);
});

test('comic source boundaries tolerate a book without a saved segment', () => {
  assert.deepEqual(comicSourceBoundaryOptions(null), []);
});

test('comic source split payload carries direction boundary and revision', () => {
  assert.deepEqual(comicSourceSplitPayload('right', 88, 3), {
    direction: 'right', boundaryOffset: 88, expectedRevision: 3
  });
});

test('comic source labels expose prompt and media staleness', () => {
  assert.equal(comicSourceStatusLabel({ promptStale: true, mediaStale: true }), '原文分段已变化');
  assert.equal(comicSourceStatusLabel({ promptStale: true }), '等待生成提示词');
  assert.equal(comicSourceStatusLabel({}), '原文已绑定');
});

test('comic source content renders only inside the comic media panel', () => {
  const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
  const inlineStart = source.indexOf('function InlineComicLibrary');
  const mediaStart = source.indexOf('function ComicMediaPanel');
  assert.ok(inlineStart >= 0 && mediaStart > inlineStart);
  assert.doesNotMatch(source.slice(inlineStart, mediaStart), /<ComicSourceContent/);
  assert.match(source.slice(mediaStart), /<ComicSourceContent[\s\S]*embedded/);
});

test('comic source text and split controls can render directly in the media modal', () => {
  const source = readFileSync(new URL('./ComicSourceContent.jsx', import.meta.url), 'utf8');
  assert.match(source, /embedded[\s\S]*batch-factory-comic-source-text/);
  assert.match(source, /embedded[\s\S]*batch-factory-comic-source-split/);
});
