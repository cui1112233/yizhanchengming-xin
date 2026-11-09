import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const html = await readFile(new URL('./index.html', import.meta.url), 'utf8');

test('prototype is self-contained and exposes the three approved modes', () => {
  assert.match(html, /<style>[\s\S]*<\/style>/);
  assert.match(html, /<script>[\s\S]*<\/script>/);
  assert.doesNotMatch(html, /<script[^>]+src=/);
  assert.doesNotMatch(html, /<link[^>]+stylesheet/);
  for (const label of ['一战晟铭', 'Agent代理', 'Duet']) assert.match(html, new RegExp(label));
  assert.doesNotMatch(html, />\s*内容创作\s*</);
});

test('prototype publishes state inspection and persistence contracts', () => {
  for (const token of ['cm-agent-duet-prototype-v1', 'loadState', 'saveState', 'patchState', 'renderApp', 'dispatchAction', '__CM_PROTOTYPE__']) {
    assert.match(html, new RegExp(token));
  }
});

test('prototype includes three approved themes without external assets', () => {
  for (const theme of ['dark', 'light', 'snow']) assert.match(html, new RegExp(`data-theme=["']?${theme}|theme:\\s*["']${theme}`));
  assert.doesNotMatch(html, /https?:\/\//);
});

test('agent skill center and composer contracts are present', () => {
  for (const token of [
    'Skills', 'MCP', 'Codex', '搜索技能', '全部开启', '导入 Skill',
    'renderSkillCenter', 'renderSkillDetail', 'renderContextPicker',
    'TOGGLE_SKILL', 'ADD_CONTEXT', 'REMOVE_CONTEXT', 'SET_PERMISSION_MODE'
  ]) assert.match(html, new RegExp(token));
  for (const label of ['计划模式', '自动审批', '完全运行']) assert.match(html, new RegExp(label));
});
