import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('../components/AgentModelPicker.jsx', import.meta.url), 'utf8');
const styles = readFileSync(new URL('../../shared/styles/agent-workspace.css', import.meta.url), 'utf8');

test('模型选择器支持搜索并展示渠道、能力和不可用原因', () => {
  assert.match(source, /showSearch/);
  assert.match(source, /providerLabel/);
  assert.match(source, /capabilityDescription/);
  assert.match(source, /当前模型不可用/);
  assert.match(source, /expectedDuration/);
  assert.match(source, /notFoundContent="当前通道没有可用模型，请在设置中切换生成通道或到 API 配置添加模型"/);
  assert.match(source, /popupMatchSelectWidth=\{false\}/);
  assert.match(source, /popupClassName="agent-model-picker__dropdown"/);
  assert.match(styles, /\.agent-model-picker__dropdown\s*\{[^}]*min-width:\s*284px\s*!important/);
});

test('模型选择器只在用户选择后通知节点保存', () => {
  assert.match(source, /onChange=\{modelId =>/);
  assert.doesNotMatch(source, /useEffect[\s\S]*onChange/);
});
