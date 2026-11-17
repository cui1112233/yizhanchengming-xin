import test from 'node:test';
import assert from 'node:assert/strict';

const helper = await import('./continuationImport.js').catch(() => ({}));
test('续写拆章保持源序、分离标题、去空白按UTF16计数', () => {
  assert.equal(typeof helper.splitContinuationChapters, 'function', 'missing local chapter splitter');
  assert.deepEqual(helper.splitContinuationChapters('  序言 😀\n\n第十章 渡口\n 雨 停了。\r\n第2章 归来\n他\t回来了。  '), [
    { title: '第1章', content: '序言 😀', wordCount: 4 },
    { title: '第十章 渡口', content: '雨 停了。', wordCount: 4 },
    { title: '第2章 归来', content: '他\t回来了。', wordCount: 5 }
  ]);
});
test('无标题、空白、仅标题及不支持的标题按已恢复拆章合同处理', () => {
  assert.equal(typeof helper.splitContinuationChapters, 'function');
  assert.deepEqual(helper.splitContinuationChapters(' \n\t '), []);
  assert.deepEqual(helper.splitContinuationChapters(' 第1章 只有标题 '), [{ title: '第1章', content: '第1章 只有标题', wordCount: 7 }]);
  assert.deepEqual(helper.splitContinuationChapters('第1章 空\n第3章 有文\n正文'), [{ title: '第3章 有文', content: '正文', wordCount: 2 }]);
  assert.deepEqual(helper.splitContinuationChapters('第一节\n 正文 '), [{ title: '第1章', content: '第一节\n 正文', wordCount: 5 }]);
});
test('同一行多个章节标记仍按每个源标记拆分', () => {
  assert.deepEqual(helper.splitContinuationChapters('第1章 空 第2章 后续\n正文'), [{ title: '第2章 后续', content: '正文', wordCount: 2 }]);
});
