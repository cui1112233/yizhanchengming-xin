(function expose(root, factory) {
  const api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  if (root) root.YiZhanUiState = api;
})(typeof window !== 'undefined' ? window : globalThis, function createUiState() {
  const accountLabels = {
    available: '可用',
    busy: '使用中',
    quota_exhausted: '今日额度用完',
    auth_required: '需要登录',
    human_verification: '需要人工验证',
    cooldown: '冷却中',
    disabled: '已停用'
  };

  function accountStatusText(state) {
    return accountLabels[state] || '状态异常';
  }

  function pairingStatusText(pairing) {
    return pairing?.paired ? '已绑定' : '未绑定';
  }

  return { accountStatusText, pairingStatusText };
});
