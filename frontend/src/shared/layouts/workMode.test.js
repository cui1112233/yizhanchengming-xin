import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { WORK_MODES, WORK_MODE_AGENT, WORK_MODE_DUET, WORK_MODE_STUDIO, workModeForPath } from './workMode.js';

const layoutSource = readFileSync(fileURLToPath(new URL('./UserLayout.jsx', import.meta.url)), 'utf8');
const menuSource = readFileSync(fileURLToPath(new URL('./WorkModeMenu.jsx', import.meta.url)), 'utf8');

test('工作模式按路由推导，聊天页是 Agent 代理，其余都是一战晟铭', () => {
  assert.equal(workModeForPath('/agent/chat'), WORK_MODE_AGENT);
  assert.equal(workModeForPath('/agent/chat/'), WORK_MODE_AGENT);
  assert.equal(workModeForPath('/agent'), WORK_MODE_STUDIO);
  assert.equal(workModeForPath('/script'), WORK_MODE_STUDIO);
  assert.equal(workModeForPath('/'), WORK_MODE_STUDIO);
});

test('Duet 永远不会成为当前模式', () => {
  ['/', '/agent', '/agent/chat', '/plugins', '/duet'].forEach(pathname => {
    assert.notEqual(workModeForPath(pathname), WORK_MODE_DUET);
  });
  assert.equal(WORK_MODES.filter(mode => mode.unavailable).length, 1);
  assert.equal(WORK_MODES[2].id, WORK_MODE_DUET);
});

test('左侧菜单移除 Agent 对话，Agent 工作区改名 CM 画布', () => {
  assert.doesNotMatch(layoutSource, /label: 'Agent 对话'/);
  assert.doesNotMatch(layoutSource, /'\/agent\/chat', icon:/);
  assert.match(layoutSource, /href: '\/agent', icon: Bot, label: 'CM 画布'/);
});

test('模式下拉只接受三项且不写 localStorage', () => {
  assert.equal(WORK_MODES.length, 3);
  assert.doesNotMatch(menuSource, /localStorage/);
  assert.match(menuSource, /Duet 模式尚未实现/);
});
