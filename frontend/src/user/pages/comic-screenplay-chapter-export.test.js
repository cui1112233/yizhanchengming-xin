import test from 'node:test';
import assert from 'node:assert/strict';
import { chapterExportFileName, exportChapterText } from './comic-screenplay-chapter-export.js';

test('导出本章文件名与内容对齐洛水规则', () => {
  // 洛水 WorksApi.export_chapter：默认名章节标题，非法路径字符替换，utf-8 txt。
  assert.equal(chapterExportFileName({ title: '第一章 相遇' }, 0), '第一章 相遇.txt');
  assert.equal(chapterExportFileName({ title: '  ' }, 2), '第3章.txt');
  assert.equal(chapterExportFileName({ title: '第1/章:重逢' }, 0), '第1_章_重逢.txt');
  assert.equal(chapterExportFileName({}, 4), '第5章.txt');

  const saved = [];
  const name = exportChapterText({ title: '第一章', source_text: '林川走进巷口。' }, 0, (fileName, content) => saved.push({ fileName, content }));
  assert.equal(name, '第一章.txt');
  assert.deepEqual(saved, [{ fileName: '第一章.txt', content: '林川走进巷口。' }]);
});

test('导出本章空原文拒绝并提示', () => {
  assert.throws(() => exportChapterText({ title: '空章' }, 0, () => {}), /还没有原文/);
  assert.throws(() => exportChapterText({ title: '空白', source_text: '   ' }, 0, () => {}), /还没有原文/);
});
