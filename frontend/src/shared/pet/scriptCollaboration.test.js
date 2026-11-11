import assert from 'node:assert/strict';
import test from 'node:test';
import { classifyPetRequestError, parseScriptRevision } from './scriptCollaboration.js';

test('parses only a non-empty complete revision marked by 【修改稿】', () => {
  assert.deepEqual(parseScriptRevision('修改重点：加强第三场冲突。\n【修改稿】\n场景一：修改后的完整剧本'), {
    summary: '修改重点：加强第三场冲突。',
    candidateOutput: '场景一：修改后的完整剧本'
  });
  assert.equal(parseScriptRevision('只有建议，没有完整稿'), null);
  assert.equal(parseScriptRevision('说明\n【修改稿】\n   '), null);
});

test('classifies timeout, configuration, upstream, and unknown pet request failures', () => {
  assert.deepEqual(classifyPetRequestError({ status: 504, message: '上游模型请求超时，请稍后重试' }), {
    message: '连接模型超时，当前剧本已保留。', action: 'retry'
  });
  // 401/403 是会话失效，即使响应体里带 key 字样也不能说成模型配置问题。
  assert.deepEqual(classifyPetRequestError({ status: 401, message: 'API key invalid' }), {
    message: '登录已过期，请重新登录后再试。', action: 'login'
  });
  assert.deepEqual(classifyPetRequestError({ status: 403, message: 'API key invalid' }), {
    message: '登录已过期，请重新登录后再试。', action: 'login'
  });
  // 422 直接展示后端业务校验文案，不再统一翻译成模型配置。
  assert.deepEqual(classifyPetRequestError({ status: 422, message: '第三场缺少出场人物' }), {
    message: '第三场缺少出场人物', action: 'retry'
  });
  assert.deepEqual(classifyPetRequestError(new Error('Model is required')), {
    message: '模型配置无法使用，请检查接口、模型名或密钥。', action: 'settings'
  });
  assert.deepEqual(classifyPetRequestError({ status: 502, message: 'Upstream returned non-JSON response' }), {
    message: '模型返回内容异常，当前剧本未改动。', action: 'retry'
  });
  assert.deepEqual(classifyPetRequestError(new Error('Network request failed')), {
    message: '网络连接不稳定，当前剧本已保留。', action: 'retry'
  });
});
