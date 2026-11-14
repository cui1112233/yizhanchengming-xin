import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const page = readFileSync(new URL('./NovelCreationPage.jsx', import.meta.url), 'utf8');
const layout = readFileSync(new URL('../../shared/layouts/UserLayout.jsx', import.meta.url), 'utf8');
const app = readFileSync(new URL('../App.jsx', import.meta.url), 'utf8');

test('小说创作拥有独立导航与路由', () => {
  assert.match(layout, /href: '\/novel-creation'[\s\S]{0,80}label: '小说创作'/);
  assert.match(app, /'\/novel-creation': NovelCreationPage/);
});

test('小说创作按真实洛水入口展示六种模式与六个工具', () => {
  for (const label of ['长篇原创', '短篇原创', '仿写长篇', '仿写短篇', '番外创作', '续写创作']) {
    assert.match(page, new RegExp(label));
  }
  for (const label of ['写作风格', '导入素材', 'Agent 设置', '定时写作', '数据看板', '备忘录']) {
    assert.match(page, new RegExp(label));
  }
  assert.doesNotMatch(page, /同人创作|创作看板|风格库/);
  assert.match(page, /开始创作/);
  assert.match(page, /我的作品/);
  assert.match(page, /会话记录/);
  assert.match(page, /新建会话/);
});

test('小说创作保留三栏工作台和真实空白页文案', () => {
  assert.match(page, /novel-creation-sidebar/);
  assert.match(page, /novel-creation-conversation/);
  assert.match(page, /novel-creation-inspector/);
  assert.match(page, /novel-creation-active-mode/);
  assert.match(page, /novel-creation-inspector-width/);
  assert.match(page, /告诉我你想写什么——题材、世界观、主角、核心冲突。/);
  assert.match(page, /输入指令… 可在模型列表中设置思考强度 \(Enter 发送，Shift\+Enter 换行\)/);
});

test('小说创作成功后直接展示后端保存的章节正文', () => {
  assert.match(page, /snapshot\?\.chapters/);
  assert.match(page, /chapter\.content/);
  assert.match(page, /chapter\.wordCount/);
});
