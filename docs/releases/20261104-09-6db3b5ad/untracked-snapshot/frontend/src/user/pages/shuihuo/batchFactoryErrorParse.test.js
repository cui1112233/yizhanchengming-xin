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

test('真正的能力门控才使用“未开通”文案', () => {
  const parts = splitBatchFactoryError('capability unavailable: CAPABILITY_NOT_ENABLED');
  assert.equal(parts.capabilityGate, true);
  assert.equal(parts.upstreamRejected, false);
  assert.match(parts.summary, /尚未开通/);
});
