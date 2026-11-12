import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('complete source modal saves controlled textarea text instead of the Ant modal click event', () => {
  assert.match(source, /async function saveSourceEditor\(\)\s*\{\s*return saveCompleteSource\(sourceEditorBook, sourceEditorValue\);\s*\}/);
  assert.match(source, /onOk=\{\(\) => saveSourceEditor\(\)\}/);
});

test('production content editor lets the user rename the book before saving', () => {
  // 发布按书名匹配漫画作品；占位书名或抓取书名有误时，用户必须在「编辑内容」
  // 弹窗里直接改书名，并随元数据一起提交。
  assert.match(source, /const \[editingContentTitle, setEditingContentTitle\] = useState\(''\)/);
  assert.match(source, /setEditingContentTitle\(String\(currentBook\?\.title \|\| ''\)\)/);
  assert.match(source, /书名不能为空/);
  assert.match(source, /updateBookMetadata\(batch\.id, editingContentBook\.id, \{ title: nextTitle/);
  assert.match(source, /value=\{editingContentTitle\}/);
});

test('production content editor lets the user correct the source book id', () => {
  // 书号填错会让正文抓取确定性失败。用户必须能在这里改书号，保存后服务端
  // 会重置抓取状态，前端只提交纯数字书号。
  assert.match(source, /const \[editingContentSourceBookId, setEditingContentSourceBookId\] = useState\(''\)/);
  assert.match(source, /setEditingContentSourceBookId\(String\(currentBook\?\.bookId \|\| ''\)\)/);
  assert.match(source, /书号只能填写 3-20 位数字/);
  assert.match(source, /sourceBookId: nextSourceBookId \|\| undefined/);
  assert.match(source, /value=\{editingContentSourceBookId\}/);
  assert.match(source, /书号填错会导致正文抓取失败，修改后保存并重试即可重新抓取。/);
});
