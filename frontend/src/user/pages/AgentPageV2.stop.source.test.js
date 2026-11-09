import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const source = fs.readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');

test('Agent conversation can stop the current request without deleting its history', () => {
  assert.match(source, /const askControllerRef = useRef\(null\)/);
  assert.match(source, /new AbortController\(\)/);
  assert.match(source, /askAgent\(\{[\s\S]*signal: controller\.signal/);
  assert.match(source, /function stopQuestion\(\)[\s\S]*askControllerRef\.current\?\.abort\(\)/);
  assert.match(source, /aria-label=\{asking \? '停止生成' : '发送消息'\}/);
  assert.match(source, /asking \? <Square/);
  assert.doesNotMatch(source, /stopQuestion[\s\S]{0,240}clearAgentTask/);
});
