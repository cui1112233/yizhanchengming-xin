import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const fields = readFileSync(new URL('./ComicScreenplayCreationFields.jsx', import.meta.url), 'utf8');
const page = readFileSync(new URL('./ComicScreenplayPage.jsx', import.meta.url), 'utf8');

test('漫剧创作标题由洛水剧本创作改名而来', () => {
  assert.match(page, /<h1>漫剧创作<\/h1><p>管理和创建您的漫剧作品<\/p>/);
  assert.match(page, /title="新建漫剧"/);
});

test('资产解析程度文案对齐洛水：大模型-3D / 大模型-写实 / 差异化 / 标准', () => {
  assert.match(fields, /value: 'master', label: '大模型-3D'/);
  assert.match(fields, /value: 'master_real', label: '大模型-写实'/);
  assert.match(fields, /value: 'more_diff', label: '差异化'/);
  assert.match(fields, /value: 'standard', label: '标准'/);
  assert.doesNotMatch(fields, /大师模式 · 3D|大师模式 · 写实/);
});

test('上传剧本格式与拖拽提示对齐洛水', () => {
  assert.match(fields, /上传剧本格式 <i>\*<\/i>/);
  assert.match(fields, /点击或拖动文件 \(txt \/ docx\)/);
});

test('选择风格占位与标签对齐洛水', () => {
  assert.match(fields, /<span>选择风格<\/span><ComicStylePicker/);
  assert.match(fields, /value\.style_name \|\| '点击选择画风'/);
});

test('改名不改数据值：degree 的 value 仍是 master/master_real', () => {
  assert.match(fields, /value: 'master'/);
  assert.match(fields, /value: 'master_real'/);
});
