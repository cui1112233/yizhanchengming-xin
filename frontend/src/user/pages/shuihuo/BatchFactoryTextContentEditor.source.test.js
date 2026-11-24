import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('text batches keep text mode in rows and the content editor even with stale book overrides', () => {
  assert.match(source, /const productionMode = batchProductionMode === 'text' \? 'text' : \(settings\.productionMode \|\| 'video'\)/);
  assert.match(source, /const editingContentProductionMode = editingContentBook \? \(batchProductionMode === 'text' \? 'text' : effectiveBookSettings/);
  assert.match(source, /const productionMode = batchProductionMode === 'text' \? 'text' : effectiveBookSettings/);
});

test('text content editor uses the captured full text and text-specific copy', () => {
  assert.match(source, /editingContentProductionMode === 'text' \? '这里编辑当前书用于文本分页的完整正文'/);
  assert.match(source, /editingContentProductionMode === 'text' \? <label className="shuihuo-form-label"><span>内容截取<small>文本分页使用的正文上限<\/small><\/span>/);
  assert.match(source, /当前小说的文本分页正文已保存；内容截取 \$\{captureCharacters\} 字/);
  assert.match(source, /productionMode === "text"\s*\? `正文总字数 \$\{batchFactoryCleanSourceText\(book\.sourceText\)\.length\} 字 · 单页约 \$\{Number\(settings\.textPageTarget \|\| 150\)\} 字`/);
});

test('text rows delegate assigned copy and rendered page navigation to the shared workbench', () => {
  assert.match(source, /import InlineTextBookPages from '\.\/InlineTextBookPages\.jsx'/);
  assert.match(source, /<InlineTextBookPages/);
  assert.match(source, /onOpen=\{pageId => openTextBook\(book, 'pages', pageId\)\}/);
  assert.doesNotMatch(source, /function DeferredInlineTextPages\(/);
});

test('saving pasted text promotes it to the formal source instead of leaving fetch-original failed', () => {
  assert.match(source, /if \(editingContentProductionMode !== 'text'\) await saveDraft\(\{ key: `working-front:\$\{editingContentBook\.id\}`/);
  assert.match(source, /editingContentProductionMode === 'text'[\s\S]*await updateBookSource\(batch\.id, editingContentBook\.id, \{ sourceText, rawSourceText: sourceText, replaceSource: true, expectedRevision:/);
  assert.match(source, /await saveDraft\(\{ key: `working-front:\$\{editingContentBook\.id\}`, kind: 'working-front-content', scope: batch\.id, content: '' \}\)/);
});
