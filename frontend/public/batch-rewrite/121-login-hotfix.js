(() => {
  'use strict';

  const CONFIG_PATH = '/api/config';
  const REQUEST_TIMEOUT_MS = 65_000;
  const REQUEST_TIMEOUT_SECONDS = Math.ceil(REQUEST_TIMEOUT_MS / 1000);
  const GUARDED_REQUEST = /^\/api\/(?:config|web-submit(?:\/|$))/;
  const originalApi = api;
  const originalSaveWebSubmitConfig = saveWebSubmitConfig;
  const originalOpenWebLoginDialog = openWebLoginDialog;

  function configFailureMessage(error) {
    const raw = String(error?.message || error || '未知错误').trim();
    if (/401|unauthori[sz]ed|未登录|登录.*失效/i.test(raw)) {
      return '小说获取配置加载失败：主站登录已失效，请重新登录后刷新页面。';
    }
    if (/404|not found/i.test(raw)) {
      return '小说获取配置加载失败：当前运行版本缺少 /api/batch-rewrite/config。';
    }
    return `小说获取配置加载失败：${raw || '未知错误'}`;
  }

  function showConfigFailure(error) {
    const text = configFailureMessage(error);
    state.config = null;
    state.configLoadError = text;
    state.webLoginSession = false;

    const loginBox = document.getElementById('webLoginStatus');
    if (loginBox) {
      loginBox.className = 'web-login-status error';
      loginBox.title = `${text} 已阻止发送 121 账号密码。`;
      loginBox.innerHTML = '<span class="web-login-dot" aria-hidden="true"></span><span>配置加载失败</span>';
    }
    if (typeof setSiteSubmitStatus === 'function') {
      setSiteSubmitStatus(`${text} 已阻止发送 121 账号密码。`);
    }
    const processResult = document.getElementById('processResult');
    if (processResult && !String(processResult.textContent || '').trim()) {
      processResult.textContent = `${text}\n请先恢复主站登录和配置接口，再进行 121 登录验证。`;
    }
    return text;
  }

  api = async function guardedBatchRewriteApi(path, options = {}) {
    if (!GUARDED_REQUEST.test(String(path || ''))) {
      return originalApi(path, options);
    }

    const controller = new AbortController();
    const inheritedSignal = options.signal;
    let inheritedAbortHandler = null;
    if (inheritedSignal) {
      if (inheritedSignal.aborted) controller.abort();
      else {
        inheritedAbortHandler = () => controller.abort();
        inheritedSignal.addEventListener('abort', inheritedAbortHandler, { once: true });
      }
    }
    const timer = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

    try {
      return await originalApi(path, { ...options, signal: controller.signal });
    } catch (error) {
      const timedOut = controller.signal.aborted && !inheritedSignal?.aborted;
      if (timedOut) {
        const target = String(path || '').includes('/web-submit/') ? '121 登录/验证请求' : '小说获取配置请求';
        throw new Error(`请求超时（${REQUEST_TIMEOUT_SECONDS} 秒）：${target}未返回，请检查主站登录和服务状态后重试。`);
      }
      throw error;
    } finally {
      window.clearTimeout(timer);
      if (inheritedSignal && inheritedAbortHandler) {
        inheritedSignal.removeEventListener('abort', inheritedAbortHandler);
      }
    }
  };

  async function ensureServerConfig() {
    if (state.config && !state.configLoadError) return true;
    try {
      const loaded = await api(CONFIG_PATH);
      if (!loaded || typeof loaded !== 'object' || Array.isArray(loaded)) {
        throw new Error('配置接口返回了无效数据');
      }
      state.config = loaded;
      state.configLoadError = '';
      renderConfig();
      return true;
    } catch (error) {
      showConfigFailure(error);
      return false;
    }
  }

  saveWebSubmitConfig = async function guardedSaveWebSubmitConfig(silent = false) {
    if (!(await ensureServerConfig())) {
      const error = new Error(state.configLoadError || '小说获取配置加载失败');
      if (!silent && typeof setSiteSubmitStatus === 'function') setSiteSubmitStatus(error.message);
      throw error;
    }
    return originalSaveWebSubmitConfig(silent);
  };

  openWebLoginDialog = async function safeOpenWebLoginDialog() {
    if (!(await ensureServerConfig())) {
      if (typeof setSiteSubmitStatus === 'function') {
        setSiteSubmitStatus(`${state.configLoadError || '小说获取配置加载失败'} 已阻止发送 121 账号密码。`);
      }
      return;
    }

    let dialog = document.getElementById('webLoginDialog');
    if (!dialog) {
      await originalOpenWebLoginDialog();
      return;
    }

    const usernameInput = document.getElementById('webLoginUsername');
    const passwordInput = document.getElementById('webLoginPassword');
    if (usernameInput) usernameInput.value = state.config?.web_submit?.username || '';
    if (passwordInput) passwordInput.value = '';
    if (!dialog.open) dialog.showModal();
  };

  // Startup hydration belongs to app.js. This hotfix only guards config and
  // login opening; it must not start a second config/task/job bootstrap
  // when the iframe is opened.
})();
