const test = require('node:test');
const assert = require('node:assert/strict');
const { accountStatusText, pairingStatusText } = require('../ui/ui-state');

test('account states render explicit Chinese labels', () => {
  assert.equal(accountStatusText('available'), '可用');
  assert.equal(accountStatusText('busy'), '使用中');
  assert.equal(accountStatusText('quota_exhausted'), '今日额度用完');
  assert.equal(accountStatusText('auth_required'), '需要登录');
  assert.equal(accountStatusText('human_verification'), '需要人工验证');
  assert.equal(accountStatusText('cooldown'), '冷却中');
});

test('pairing state is understandable without relying on color', () => {
  assert.equal(pairingStatusText({ paired: true }), '已绑定');
  assert.equal(pairingStatusText({ paired: false }), '未绑定');
});
