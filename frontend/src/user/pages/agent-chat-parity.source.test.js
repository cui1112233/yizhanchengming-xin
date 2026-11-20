import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const page = readFileSync(new URL('./AgentChatPage.jsx', import.meta.url), 'utf8');
const api = readFileSync(new URL('../../shared/api/agentChat.js', import.meta.url), 'utf8');
const css = readFileSync(new URL('../../shared/styles/agent-chat.css', import.meta.url), 'utf8');
const routes = readFileSync(new URL('../../../../backend/internal/httpapi/agent_chat_routes.go', import.meta.url), 'utf8');
const skillsCompat = readFileSync(new URL('../../../../backend/internal/httpapi/agent_skills_compat.go', import.meta.url), 'utf8');

test('新会话表单包含目标、工作区、主模型、工作模型、思考强度与权限', () => {
  assert.match(page, /创作目标/);
  assert.match(page, /项目工作区/);
  assert.match(page, /主模型/);
  assert.match(page, /工作模型/);
  assert.match(page, /思考强度/);
  assert.match(page, /权限模式/);
  assert.match(page, /开始创作/);
});

test('输入框提供 / 技能与 @ 引用选择器及可移除上下文胶囊', () => {
  assert.match(page, /选择技能/);
  assert.match(page, /引用技能或项目文件/);
  assert.match(page, /\/skill:\$\{clean\.id\}/);
  assert.match(page, /@\$\{clean\.type\}:\$\{clean\.id\}/);
  assert.match(page, /agent-chat-refs/);
  assert.match(css, /\.agent-chat-picker/);
  assert.match(css, /\.agent-chat-ref/);
});

test('左栏包含自动任务计数与 Agent 技能计数，工作区支持置顶/重命名/文件页打开', () => {
  assert.match(page, /自动任务/);
  assert.match(page, /Agent 技能/);
  assert.match(page, /skills\.length/);
  assert.match(page, /置顶/);
  assert.match(page, /重命名/);
  assert.match(page, /在文件页打开/);
});

test('API 暴露技能列表并传递工作模型与上下文引用', () => {
  assert.match(api, /\/api\/agent\/skills/);
  assert.match(api, /workModel/);
  assert.match(page, /contextRefs: refs/);
  assert.match(skillsCompat, /GET \/api\/agent\/skills/);
  assert.match(routes, /workModel/);
  assert.match(routes, /contextRefs/);
});
