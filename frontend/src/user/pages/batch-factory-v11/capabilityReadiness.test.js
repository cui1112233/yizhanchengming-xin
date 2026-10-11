import assert from 'node:assert/strict';
import test from 'node:test';
import { unavailableCapabilitySummary } from './capabilityReadiness.js';

test('能力提示只列出服务端明确关闭的生产能力并翻译原因', () => {
  const result = unavailableCapabilitySummary({
    'production.submit': { available: false, reason: 'Production slice not released' },
    'merge.run': { available: false, reason: 'Merge slice not released' },
    'publish.121': { available: true },
    'publish.yadi': { available: false, reason: 'Yadi is not enabled' }
  });
  assert.deepEqual(result.map(item => item.label), ['正式生产', '批量合成', 'Yadi 发布']);
  assert.deepEqual(result.map(item => item.reason), ['正式生产模块尚未发布', '合成模块尚未发布', 'Yadi 发布尚未启用']);
});

test('没有服务端能力记录时不误报生产阻塞', () => {
  assert.deepEqual(unavailableCapabilitySummary({}), []);
});
