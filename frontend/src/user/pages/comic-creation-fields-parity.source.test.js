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

test('新建漫剧主表单按洛水顺序：模型提示词配置徽标 → 风格/比例并排 → 资产解析横排 → 上传 → 高级折叠', () => {
  const modal = page.slice(page.indexOf('title="新建漫剧"'));
  const order = [
    '作品名称 <i>*</i>',
    'comic-model-config-badge',
    'comic-create-row',
    'comic-field-title">资产解析程度',
    'ComicDegreeSegments',
    'ComicSourceInput',
    'comic-advanced-toggle',
    'ComicCreativeFields hideStyle'
  ];
  let cursor = -1;
  for (const token of order) {
    const at = modal.indexOf(token, cursor + 1);
    assert.ok(at > cursor, `字段顺序错位或未找到：${token}`);
    cursor = at;
  }
  // 洛水主表单不含平铺的 4 个提示词：必须收进折叠区
  const beforeAdvanced = modal.slice(0, modal.indexOf('comic-advanced-toggle'));
  assert.doesNotMatch(beforeAdvanced, /剧本提示词|分镜提示词|角色设计提示词|场景设计提示词|图片核心要求/);
});

test('资产解析程度用洛水横排 4 段组件，默认提示词按用途自动补齐', () => {
  assert.match(fields, /ComicDegreeSegments/);
  assert.match(fields, /degree-segments/);
  assert.match(page, /defaultCreativePrompts\(library, current\)/);
  const creative = readFileSync(new URL('./comic-screenplay-creative.js', import.meta.url), 'utf8');
  assert.match(creative, /export function defaultCreativePrompts/);
  assert.match(creative, /for \(const \{ field, purpose \} of creativePromptFields\)/);
});
