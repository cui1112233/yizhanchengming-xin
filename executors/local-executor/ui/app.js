const bridge = window.yiZhanExecutor;
const ui = window.YiZhanUiState;
let lastState = null;
let lastUpdaterState = null;

const elements = {
  pairForm: document.getElementById('pairForm'),
  baseUrl: document.getElementById('baseUrl'),
  pairCode: document.getElementById('pairCode'),
  pairButton: document.getElementById('pairButton'),
  pairBadge: document.getElementById('pairBadge'),
  pairDetail: document.getElementById('pairDetail'),
  addAccount: document.getElementById('addAccount'),
  accountList: document.getElementById('accountList'),
  accountEmpty: document.getElementById('accountEmpty'),
  automationText: document.getElementById('automationText'),
  automationButton: document.getElementById('automationButton'),
  currentTask: document.getElementById('currentTask'),
  updateCurrentVersion: document.getElementById('updateCurrentVersion'),
  updateChannel: document.getElementById('updateChannel'),
  updateStatus: document.getElementById('updateStatus'),
  updateCheckButton: document.getElementById('updateCheckButton'),
  updateInstallButton: document.getElementById('updateInstallButton'),
  globalError: document.getElementById('globalError')
};

function showError(message) {
  elements.globalError.hidden = !message;
  elements.globalError.textContent = message || '';
}

function render(state) {
  lastState = state;
  const pairing = state?.pairing || { paired: false };
  elements.pairBadge.textContent = ui.pairingStatusText(pairing);
  elements.pairBadge.dataset.state = pairing.paired ? 'ok' : 'idle';
  elements.pairDetail.textContent = pairing.paired
    ? `已绑定：${pairing.baseUrl} · ${pairing.executorId}`
    : '当前未绑定。';
  if (pairing.baseUrl && !elements.baseUrl.value) elements.baseUrl.value = pairing.baseUrl;

  const accounts = Array.isArray(state?.accounts) ? state.accounts : [];
  elements.accountEmpty.hidden = accounts.length > 0;
  elements.accountList.replaceChildren(...accounts.map(renderAccount));

  const automation = state?.automation || { enabled: false, ready: false, reason: '' };
  elements.automationText.textContent = automation.reason || '请先完成绑定和豆包账号登录。';
  elements.automationButton.textContent = automation.enabled ? '关闭自动任务' : '开启自动任务';
  elements.automationButton.disabled = !automation.enabled && !automation.ready;
  elements.automationButton.setAttribute('aria-pressed', automation.enabled ? 'true' : 'false');
  elements.currentTask.textContent = state?.currentTask?.id || '暂无';
  showError(state?.lastError || '');
  if (lastUpdaterState) renderUpdater(lastUpdaterState);
}

function renderUpdater(state) {
  lastUpdaterState = state || {};
  elements.updateCurrentVersion.textContent = state?.currentVersion || '未知';
  if (state?.channel && elements.updateChannel.value !== state.channel) {
    elements.updateChannel.value = state.channel;
  }
  elements.updateStatus.textContent = state?.message || '尚未检查更新';
  elements.updateStatus.dataset.state = state?.status || 'idle';
  elements.updateCheckButton.disabled = ['checking', 'downloading', 'installing'].includes(state?.status);
  elements.updateCheckButton.textContent = state?.status === 'checking'
    ? '检查中…'
    : state?.status === 'downloading'
      ? '下载中…'
      : '检查更新';

  const downloaded = state?.status === 'downloaded' || state?.status === 'install_deferred';
  elements.updateInstallButton.hidden = !downloaded;
  elements.updateInstallButton.disabled = state?.status !== 'downloaded' || Boolean(lastState?.currentTask);
  elements.updateInstallButton.textContent = state?.status === 'install_deferred' || lastState?.currentTask
    ? '任务结束后更新'
    : '立即更新并重启';
}

function renderAccount(account) {
  const row = document.createElement('article');
  row.className = 'account-row';

  const info = document.createElement('div');
  info.className = 'account-info';
  const name = document.createElement('strong');
  name.textContent = account.name || account.id;
  const status = document.createElement('span');
  status.className = 'account-status';
  status.dataset.state = account.state || 'unknown';
  status.textContent = ui.accountStatusText(account.state);
  info.append(name, status);

  const actions = document.createElement('div');
  actions.className = 'account-actions';
  const openButton = document.createElement('button');
  openButton.className = 'secondary';
  openButton.type = 'button';
  openButton.textContent = account.state === 'human_verification' ? '打开人工验证' : '打开豆包';
  openButton.setAttribute('aria-label', `打开 ${account.name || account.id}`);
  openButton.addEventListener('click', () => invoke(() => bridge.openAccount(account.id)));
  actions.append(openButton);

  if (account.state === 'auth_required' || account.state === 'human_verification') {
    const readyButton = document.createElement('button');
    readyButton.className = 'primary';
    readyButton.type = 'button';
    readyButton.textContent = account.state === 'human_verification' ? '验证完成' : '我已登录';
    readyButton.setAttribute('aria-label', `标记 ${account.name || account.id} 可用`);
    readyButton.addEventListener('click', () => invoke(() => bridge.markAccountAvailable(account.id)));
    actions.append(readyButton);
  }

  row.append(info, actions);
  return row;
}

async function invoke(fn) {
  try {
    showError('');
    const result = await fn();
    if (result?.pairing || result?.accounts) render(result);
    else render(await bridge.getState());
    return result;
  } catch (error) {
    showError(error?.message || String(error));
    return null;
  }
}

async function invokeUpdater(fn) {
  try {
    const result = await fn();
    if (result?.status) renderUpdater(result);
    else renderUpdater(await bridge.updater.getState());
    return result;
  } catch (error) {
    renderUpdater({
      ...(lastUpdaterState || {}),
      status: 'error',
      message: error?.message || String(error)
    });
    return null;
  }
}

elements.pairForm.addEventListener('submit', async event => {
  event.preventDefault();
  elements.pairButton.disabled = true;
  elements.pairButton.textContent = '绑定中…';
  await invoke(() => bridge.pair({
    baseUrl: elements.baseUrl.value.trim(),
    code: elements.pairCode.value.trim()
  }));
  elements.pairButton.disabled = false;
  elements.pairButton.textContent = '绑定';
});

elements.addAccount.addEventListener('click', () => invoke(() => bridge.addAccount()));
elements.automationButton.addEventListener('click', () => {
  const enabled = Boolean(lastState?.automation?.enabled);
  return invoke(() => bridge.setAutomationEnabled(!enabled));
});
elements.updateCheckButton.addEventListener('click', () => invokeUpdater(() => bridge.updater.check()));
elements.updateChannel.addEventListener('change', async () => {
  await invokeUpdater(() => bridge.updater.setChannel(elements.updateChannel.value));
  return invokeUpdater(() => bridge.updater.check());
});
elements.updateInstallButton.addEventListener('click', () => invokeUpdater(() => bridge.updater.install()));

bridge.onState(render);
bridge.updater.onState(renderUpdater);
invoke(() => bridge.getState());
invokeUpdater(() => bridge.updater.getState());
