import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./CaptionWorkspace.jsx', import.meta.url), 'utf8');

test('字幕工作区覆盖模型配置、语言、任务状态和取消重试', () => {
  for (const text of ['识别字幕', '前往 API 配置', '识别语言', '排队中', '识别中', '已取消', '重新识别', '取消']) assert.match(source, new RegExp(text));
  assert.match(source, /onStart/);
  assert.match(source, /onCancel/);
  assert.match(source, /activeJob/);
});

test('字幕工作区保留导入导出和手动字幕入口', () => {
  for (const text of ['导入 SRT', '导出 SRT', '添加手动字幕']) assert.match(source, new RegExp(text));
});
