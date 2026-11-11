import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const source = fs.readFileSync(new URL('../components/AISegmentWorkbench.jsx', import.meta.url), 'utf8');

test('AI分段工作台使用模式合同并提供真实分析导出动作', () => {
  assert.match(source, /visibleAISegmentModes/);
  assert.match(source, /aiSegmentModeHas/);
  assert.match(source, /analyzeAISegment/);
  assert.match(source, /exportAISegment/);
  assert.match(source, /draft_content\.json/);
  assert.match(source, /请选择已导出的成片/);
  assert.match(source, /取消分析|controllerRef\.current\?\.abort/);
  assert.doesNotMatch(source, /功能开发中|尚未接入|知道了/);
});
