import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('../components/AgentModelPicker.jsx', import.meta.url), 'utf8');

test('模型选择器支持搜索并展示渠道、能力和不可用原因', () => {
  assert.match(source, /showSearch/);
  assert.match(source, /providerLabel/);
  assert.match(source, /capabilityDescription/);
  assert.match(source, /当前模型不可用/);
  assert.match(source, /expectedDuration/);
});

test('模型选择器只在用户选择后通知节点保存', () => {
  assert.match(source, /onChange=\{modelId =>/);
  assert.doesNotMatch(source, /useEffect[\s\S]*onChange/);
});
