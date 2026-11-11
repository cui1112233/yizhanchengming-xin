import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const file = new URL('./AICutAgentPanel.jsx', import.meta.url);
const source = fs.existsSync(file) ? fs.readFileSync(file, 'utf8') : '';

test('AI-CUT Agent面板连接真实运行、事件、消息和取消接口', () => {
  assert.match(source, /createAICutAgentRun/);
  assert.match(source, /getAICutAgentRun/);
  assert.match(source, /listAICutAgentEvents/);
  assert.match(source, /listAICutAgentMessages/);
  assert.match(source, /cancelAICutAgentRun/);
  assert.match(source, /onTimelineChanged/);
});

test('AI-CUT Agent面板具备独立收起、模式、模型、建议卡和停止操作', () => {
  assert.match(source, /aicut\.agentCollapsed/);
  assert.match(source, /Agent/);
  assert.match(source, /问答/);
  assert.match(source, /粗剪成片/);
  assert.match(source, /识别字幕/);
  assert.match(source, /去口误停顿/);
  assert.match(source, /配背景音乐/);
  assert.match(source, /停止/);
  assert.match(source, /展开 Agent/);
  assert.match(source, /收起 Agent/);
});

test('建议卡只填入输入框，发送必须经过后端Agent运行', () => {
  assert.match(source, /chooseAgentSuggestion/);
  assert.match(source, /setPrompt\(text\)/);
  assert.match(source, /sendAgentPrompt/);
  assert.doesNotMatch(source, /resolveAICutAgentTool/);
});
