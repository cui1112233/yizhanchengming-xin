import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  WORK_MODES,
  WORK_MODE_AGENT,
  WORK_MODE_DUET,
  WORK_MODE_MAGIC_CUT,
  WORK_MODE_STUDIO,
  workModeForLocation,
  workModeForPath,
  workModeHref
} from './workMode.js';

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
  assert.equal(WORK_MODES[3].id, WORK_MODE_DUET);
});

test('4040 端口会被识别成魔剪，不会误显示一战晟铭', () => {
  assert.equal(workModeForLocation({ pathname: '/', port: '4040' }), WORK_MODE_MAGIC_CUT);
  assert.equal(workModeForLocation({ pathname: '/agent/chat', port: '4001' }), WORK_MODE_AGENT);
  assert.equal(workModeForLocation({ pathname: '/', port: '4001' }), WORK_MODE_STUDIO);
});

test('本地模式在 4001 和 4040 之间只替换端口', () => {
  const studio = { protocol: 'http:', hostname: '127.0.0.1', port: '4001', pathname: '/' };
  const magicCut = { protocol: 'http:', hostname: '127.0.0.1', port: '4040', pathname: '/' };

  assert.equal(workModeHref(WORK_MODE_MAGIC_CUT, studio), 'http://127.0.0.1:4040/');
  assert.equal(workModeHref(WORK_MODE_STUDIO, magicCut), 'http://127.0.0.1:4001/');
  assert.equal(workModeHref(WORK_MODE_AGENT, magicCut), 'http://127.0.0.1:4001/agent/chat');
});

test('左侧菜单移除 Agent 对话，Agent 工作区改名 CM 画布', () => {
  assert.doesNotMatch(layoutSource, /label: 'Agent 对话'/);
  assert.doesNotMatch(layoutSource, /'\/agent\/chat', icon:/);
  assert.match(layoutSource, /href: '\/agent', icon: Bot, label: 'CM 画布'/);
});

test('模式下拉按固定顺序提供四项且不写 localStorage', () => {
  assert.deepEqual(WORK_MODES.map(mode => mode.label), ['一战晟铭', 'Agent 代理', '魔剪', 'Duet']);
  assert.doesNotMatch(menuSource, /localStorage/);
  assert.match(menuSource, /Duet 模式尚未实现/);
});
