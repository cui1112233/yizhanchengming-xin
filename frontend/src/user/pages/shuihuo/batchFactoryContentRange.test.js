import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryCleanSourceText, batchFactoryPreviewText, batchFactoryProductionText, comicAnalysisCharactersForBook, contentCaptureCharactersForBook, contentRangeLinesForBook, publishContentWithWorkingFront, sourceTextAfterViralAdoption } from './batchFactoryContentRange.js';

test('uses the saved per-book content range as the work-text boundary', () => {
  const book = { sourceMetadata: { contentRangeLines: 5 } };
  assert.equal(contentRangeLinesForBook(book), 5);
  assert.equal(batchFactoryPreviewText('第一段\n\n第二段\n第三段\n第四段\n第五段\n第六段', book), '第一段\n第二段\n第三段\n第四段\n第五段');
});

test('keeps the full original separate from the limited work text', () => {
  const source = '一\n二\n三';
  assert.equal(batchFactoryPreviewText(source, { sourceMetadata: { contentRangeLines: 2 } }), '一\n二');
  assert.equal(source, '一\n二\n三');
});

test('treats an object-coercion placeholder as missing source text', () => {
  assert.equal(batchFactoryPreviewText('[object Object]', { sourceMetadata: { contentRangeLines: 5 } }), '');
  assert.equal(batchFactoryProductionText('[object Object]', '', {}), '');
  assert.equal(sourceTextAfterViralAdoption({ type: 'click' }), '');
});

test('skips repeated page-state remnants and punctuation-only lines before applying the content range', () => {
  const source = '修改中&nbsp;\n修改中&nbsp;\n。\n\n陆雪凝推门进来。\n她把伞放在门边。\n窗外下着雨。';
  const book = { sourceMetadata: { contentRangeLines: 2 } };

  assert.equal(batchFactoryPreviewText(source, book), '陆雪凝推门进来。\n她把伞放在门边。');
  assert.equal(source.startsWith('修改中&nbsp;'), true);
});

test('cleans the stored reading view with the same rules used for production', () => {
  assert.equal(
    batchFactoryCleanSourceText('修改中&nbsp;\n。\n第一段 <b>正文</b>\n第二段正文'),
    '第一段 正文\n第二段正文'
  );
});

test('uses only the configured non-empty lines as production text unless a working front exists', () => {
  const book = { sourceMetadata: { contentRangeLines: 5 } };
  const source = '一\n\n二\n三\n四\n五\n六';
  assert.equal(batchFactoryProductionText(source, '', book), '一\n二\n三\n四\n五');
  assert.equal(batchFactoryProductionText(source, '改写甲\n改写乙', book), '改写甲\n改写乙');
  assert.equal(source, '一\n\n二\n三\n四\n五\n六');
});

test('defaults to a 4000-character fetch and upload body limit', () => {
  assert.equal(contentCaptureCharactersForBook({}), 4000);
  assert.equal(contentCaptureCharactersForBook({ sourceMetadata: { contentCaptureCharacters: 8000 } }), 8000);
});

test('reports the actual comic analysis characters at the nearest safe boundary', () => {
  const source = `${'甲'.repeat(99)}。${'乙'.repeat(20)}。`;
  assert.equal(comicAnalysisCharactersForBook({ sourceText: source, sourceMetadata: { comicAnalysisCharacters: 100 } }), 100);
});

test('keeps comic analysis target separate from the 4000-character capture limit', () => {
  const source = `${'甲'.repeat(1999)}。${'乙'.repeat(2500)}。`;
  const book = { sourceText: source, sourceMetadata: { comicAnalysisCharacters: 2000, contentCaptureCharacters: 4000 } };
  assert.equal(comicAnalysisCharactersForBook(book), 2000);
  assert.equal(contentCaptureCharactersForBook(book), 4000);
});


test('only replaces the configured preview lines when publishing an approved rewrite', () => {
  const book = { sourceMetadata: { contentRangeLines: 2 } };
  assert.equal(publishContentWithWorkingFront('一\n二\n三\n四', '甲\n乙', book), '甲\n乙\n三\n四');
});

test('adopting a viral candidate replaces the complete source while the workbench preview keeps the selected line count', () => {
  const adopted = sourceTextAfterViralAdoption('爆款第一行\n爆款第二行\n爆款第三行\n爆款第四行\n爆款第五行\n爆款第六行');

  assert.equal(adopted, '爆款第一行\n爆款第二行\n爆款第三行\n爆款第四行\n爆款第五行\n爆款第六行');
  assert.equal(batchFactoryPreviewText(adopted, { sourceMetadata: { contentRangeLines: 5 } }), '爆款第一行\n爆款第二行\n爆款第三行\n爆款第四行\n爆款第五行');
});
