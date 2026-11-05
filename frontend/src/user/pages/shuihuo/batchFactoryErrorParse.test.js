import assert from 'node:assert/strict';
import test from 'node:test';
import { splitBatchFactoryError } from './batchFactoryErrorParse.js';

const H3_REJECTION = 'capability unavailable: 视频模型提交失败（video_85472c44367867d248c865aacf AutoDL H3 rejected submission: code=RequestParametersWrong message=缺少必填参数: ref_image_0 request_id=req-123';

test('上游业务拒绝不再显示 capability unavailable 前缀', () => {
  const parts = splitBatchFactoryError(H3_REJECTION);
  assert.equal(parts.capabilityGate, false);
  assert.equal(parts.upstreamRejected, true);
  assert.equal(parts.provider, 'AutoDL H3');
  assert.equal(parts.summary, '上游要求参考图 ref_image_0，本次提交未包含，已被拒绝。');
  assert.doesNotMatch(parts.summary, /capability unavailable/i);
  assert.equal(parts.code, 'RequestParametersWrong');
  assert.equal(parts.message, '缺少必填参数: ref_image_0');
  assert.equal(parts.requestId, 'req-123');
});

test('缺配置不再被渲染成 revision conflict', () => {
  const parts = splitBatchFactoryError('revision conflict: 请在资产设置中选择漫画资产提示词');
  assert.equal(parts.summary, '请在资产设置中选择漫画资产提示词');
  assert.doesNotMatch(parts.summary, /revision conflict/i);
  // 详情里仍保留原始记录，便于排查，但折叠文案不串味。
  assert.match(parts.raw, /revision conflict/);
});

test('真正的版本冲突使用重试文案', () => {
  assert.equal(splitBatchFactoryError('revision conflict').summary, '数据版本冲突，请重试');
  assert.equal(splitBatchFactoryError('revision conflict: 漫画页版本已变化').summary, '数据版本冲突，请重试');
});

test('真正的能力门控才使用“未开通”文案', () => {
  const parts = splitBatchFactoryError('capability unavailable: CAPABILITY_NOT_ENABLED');
  assert.equal(parts.capabilityGate, true);
  assert.equal(parts.upstreamRejected, false);
  assert.match(parts.summary, /尚未开通/);
});

test('capability unavailable 里的缺配置原因不再显示为“未开通”', () => {
  const parts = splitBatchFactoryError(
    'capability unavailable: 文本模型请求失败：请在资产设置中选择漫画资产提示词'
  );
  assert.equal(parts.capabilityGate, true);
  assert.equal(parts.summary, '漫画资产提示词未选择，请在资产设置中选择。');
  assert.doesNotMatch(parts.summary, /尚未开通|Slice 门控/);
});

test('其它 capability unavailable 业务原因优先展示原因本身', () => {
  const parts = splitBatchFactoryError('capability unavailable: 文本模型请求失败：未配置可用文本模型');
  assert.equal(parts.summary, '文本模型请求失败：未配置可用文本模型');
});
