import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryVisibleError } from './batchFactoryBookState.js';

test('maps upstream 401 to an API key check message', () => {
  const message = batchFactoryVisibleError('YFAI Seedance returned HTTP 401（密钥无效或无权限，请检查 API Key）');
  assert.equal(message, '视频模型的 API Key 无效或无权限，请到「API 配置」检查或重建密钥。');
});

test('maps upstream 429 to a rate limit message', () => {
  const message = batchFactoryVisibleError('YFAI Seedance returned HTTP 429（接口限流，请 7 秒后重试）');
  assert.equal(message, '视频模型接口限流，请稍后再试；系统会按上游 Retry-After 自动退避。');
});

test('keeps the existing 402 quota message', () => {
  const message = batchFactoryVisibleError('YFAI Seedance returned HTTP 402（账户余额不足，请充值）');
  assert.equal(message, '当前视频模型的账户额度或权限不足，请更换视频模型，或检查该模型账户的额度与权限。');
});

test('passes unrelated provider errors through unchanged', () => {
  assert.equal(batchFactoryVisibleError('video provider returned HTTP 500'), 'video provider returned HTTP 500');
});
