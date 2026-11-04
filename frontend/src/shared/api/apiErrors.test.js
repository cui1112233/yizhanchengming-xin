import test from 'node:test';
import assert from 'node:assert/strict';
import { apiErrorCode, apiErrorMessage, ERROR_MESSAGES } from '../apiErrors.js';
import contract from '../apiErrorCodes.json' with { type: 'json' };

const REQUIRED_CODES = [
  'IMAGE_MODEL_NOT_CONFIGURED',
  'VIDEO_PROVIDER_NOT_CONFIGURED',
  'EXECUTOR_NOT_READY',
  'GIANT_PLUGIN_NOT_CONFIGURED'
];

test('apiErrorCodes.json declares the four setup error codes', () => {
  for (const code of REQUIRED_CODES) {
    assert.ok(contract.codes.some(item => item.code === code), `missing contract entry: ${code}`);
  }
});

test('apiErrorCode prefers the explicit server code', () => {
  assert.equal(apiErrorCode({ code: 'EXECUTOR_NOT_READY', message: '执行器离线' }), 'EXECUTOR_NOT_READY');
  assert.equal(apiErrorCode({ message: 'GIANT_EXECUTOR_OFFLINE' }), 'GIANT_EXECUTOR_OFFLINE');
  assert.equal(apiErrorCode({ message: '普通中文错误' }), '');
});

test('apiErrorMessage maps the four setup codes to the shared copy', () => {
  for (const code of REQUIRED_CODES) {
    const message = apiErrorMessage({ code, message: 'legacy text' });
    assert.equal(message, ERROR_MESSAGES[code]);
    assert.ok(message.length > 10);
  }
});

test('CAPABILITY_NOT_ENABLED 显示为可读的能力未开启文案', () => {
  assert.ok(contract.codes.some(item => item.code === 'CAPABILITY_NOT_ENABLED'));
  const message = apiErrorMessage({ code: 'CAPABILITY_NOT_ENABLED', message: '404 page not found' });
  assert.equal(message, ERROR_MESSAGES.CAPABILITY_NOT_ENABLED);
  assert.match(message, /未开启/);
  assert.doesNotMatch(message, /404 page not found/);
});

test('apiErrorMessage keeps legacy text and rewrites HTML gateway pages', () => {
  assert.equal(apiErrorMessage({ message: '旧的错误文案' }), '旧的错误文案');
  assert.equal(apiErrorMessage({ message: '502 Bad Gateway' }), '服务器通讯短暂中断（正在更新或重启），请稍后重试。');
  assert.equal(apiErrorMessage(null, '兜底文案'), '兜底文案');
});
