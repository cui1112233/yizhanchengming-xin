(() => {
  'use strict';

  const FEEDBACK_ROOT_ID = 'novelFetchInteractionFeedback';
  const STYLE_ID = 'novelFetchInteractionFeedbackStyle';
  const BUSY_TIMEOUT_MS = 60000;
  const watchedStatusIds = new Set([
    'processResult',
    'siteSubmitStatus',
    'batchStatus',
    'webSubmitSelectionStatus',
  ]);
  const actionLabels = {
    processBtn: '正在处理，请稍候…',
    openWebSubmitBtn: '正在准备提交网络…',
    openWebSubmitFromTasks: '正在准备提交网络…',
    submitWebBtn: '正在提交网络，请稍候…',
  };
  const errorPattern = /(失败|错误|异常|超时|未登录|登录失效|不可用|请求失败|网络错误|请先|不能为空|请选择|unauthorized|forbidden|HTTP\s*[45]\d\d|network\s*error|failed|error|timeout)/i;
  const successPattern = /(成功|完成|已提交|已启动|已加入|已创建|已保存|已恢复|确认完成)/i;

  let busyTimer = null;
  let lastToastKey = '';
  let lastToastAt = 0;

  function ensureStyles() {
    if (document.getElementById(STYLE_ID)) return;
    const style = document.createElement('style');
    style.id = STYLE_ID;
    style.textContent = `
      #${FEEDBACK_ROOT_ID} {
        position: fixed;
        top: 18px;
        right: 18px;
        z-index: 2147483647;
        width: min(420px, calc(100vw - 36px));
        display: flex;
        flex-direction: column;
        gap: 10px;
        pointer-events: none;
      }
      .novel-fetch-feedback-toast {
        display: none !important;
        box-sizing: border-box;
        border-radius: 10px;
        padding: 12px 14px;
        background: rgba(17, 24, 39, 0.96);
        color: #fff;
        box-shadow: 0 12px 34px rgba(0, 0, 0, 0.28);
        font-size: 14px;
        line-height: 1.5;
        word-break: break-word;
        pointer-events: auto;
        border: 1px solid rgba(255, 255, 255, 0.14);
      }
      .novel-fetch-feedback-toast[data-type="error"] {
        background: rgba(127, 29, 29, 0.97);
        border-color: rgba(254, 202, 202, 0.65);
      }
      .novel-fetch-feedback-toast[data-type="success"] {
        background: rgba(20, 83, 45, 0.97);
        border-color: rgba(187, 247, 208, 0.55);
      }
      [aria-busy="true"].novel-fetch-feedback-busy {
        cursor: progress !important;
        position: relative;
      }
    `;
    document.head.appendChild(style);
  }

  function ensureRoot() {
    ensureStyles();
    let root = document.getElementById(FEEDBACK_ROOT_ID);
    if (root) return root;
    root = document.createElement('div');
    root.id = FEEDBACK_ROOT_ID;
    root.setAttribute('aria-live', 'polite');
    root.setAttribute('aria-atomic', 'true');
    document.body.appendChild(root);
    return root;
  }

  function normalizeMessage(value) {
    return String(value || '').replace(/\s+/g, ' ').trim().slice(0, 600);
  }

  function headerMessage(value) {
    const lines = String(value || '').split(/\r?\n/).map(line => line.trim()).filter(Boolean);
    const latest = lines[lines.length - 1] || '';
    return normalizeMessage(latest.replace(/^\[[^\]]+\]\s*/, '')).slice(0, 140);
  }

  function publishParentStatus(message, type = 'info') {
    const text = headerMessage(message);
    if (!text || !window.parent || window.parent === window) return;
    const allowed = new Set(['info', 'working', 'success', 'warning', 'error']);
    window.parent.postMessage({
      type: 'qiantie:novel-fetch-status',
      text,
      tone: allowed.has(type) ? type : 'info'
    }, window.location.origin);
  }

  window.qiantiePublishNovelFetchStatus = publishParentStatus;

  function showToast(message, type = 'info', durationMs) {
    const text = normalizeMessage(message);
    if (!text) return;
    publishParentStatus(text, type);

    const now = Date.now();
    const key = `${type}:${text}`;
    if (key === lastToastKey && now - lastToastAt < 1200) return;
    lastToastKey = key;
    lastToastAt = now;

    const toast = document.createElement('div');
    toast.className = 'novel-fetch-feedback-toast';
    toast.dataset.type = type;
    if (type === 'error') {
      toast.setAttribute('role', 'alert');
      toast.setAttribute('aria-live', 'assertive');
    } else {
      toast.setAttribute('role', 'status');
    }
    toast.textContent = text;
    ensureRoot().appendChild(toast);

    const ttl = Number(durationMs) || (type === 'error' ? 9000 : 4200);
    window.setTimeout(() => toast.remove(), ttl);
  }

  function clearBusy() {
    if (busyTimer) {
      window.clearTimeout(busyTimer);
      busyTimer = null;
    }
    document.querySelectorAll('.novel-fetch-feedback-busy[aria-busy="true"]').forEach((button) => {
      button.setAttribute('aria-busy', 'false');
      button.classList.remove('novel-fetch-feedback-busy');
    });
  }

  function markBusy(button, label) {
    clearBusy();
    if (button instanceof HTMLElement) {
      button.setAttribute('aria-busy', 'true');
      button.classList.add('novel-fetch-feedback-busy');
    }
    showToast(label, 'info');
    busyTimer = window.setTimeout(() => {
      clearBusy();
      showToast('操作等待时间较长，请检查当前任务状态或网络连接。', 'error');
    }, BUSY_TIMEOUT_MS);
  }

  function classifyStatus(text) {
    if (!text) return null;
    if (/处理完成/.test(text) && !/(失败|错误|异常|超时)\s*[：:]?\s*[1-9]\d*/.test(text)) return 'success';
    if (errorPattern.test(text)) return 'error';
    if (successPattern.test(text)) return 'success';
    return 'info';
  }

  function surfaceStatusElement(element) {
    if (!(element instanceof HTMLElement) || !watchedStatusIds.has(element.id)) return;
    const text = normalizeMessage(element.textContent);
    if (!text || text === '-' || text === '—') return;
    const type = classifyStatus(text);
    publishParentStatus(text, type);
    if (!type) return;
    if (type === 'error') {
      clearBusy();
      showToast(text, 'error');
      return;
    }
    if (document.querySelector('.novel-fetch-feedback-busy[aria-busy="true"]')) {
      clearBusy();
      showToast(text, type);
    }
  }

  function surfaceMutationTarget(target) {
    if (!(target instanceof Node)) return;
    const element = target.nodeType === Node.ELEMENT_NODE ? target : target.parentElement;
    if (!(element instanceof HTMLElement)) return;
    if (watchedStatusIds.has(element.id)) {
      surfaceStatusElement(element);
      return;
    }
    const watched = element.closest && element.closest('#processResult, #siteSubmitStatus, #batchStatus, #webSubmitSelectionStatus');
    if (watched) surfaceStatusElement(watched);
  }

  function installStatusObserver() {
    const observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        surfaceMutationTarget(mutation.target);
        mutation.addedNodes.forEach((node) => {
          if (!(node instanceof HTMLElement)) return;
          if (watchedStatusIds.has(node.id)) surfaceStatusElement(node);
          node.querySelectorAll?.('#processResult, #siteSubmitStatus, #batchStatus, #webSubmitSelectionStatus').forEach(surfaceStatusElement);
        });
      }
    });
    observer.observe(document.body, { childList: true, subtree: true, characterData: true });
  }

  function installActionFeedback() {
    document.addEventListener('click', (event) => {
      const button = event.target instanceof Element ? event.target.closest('button') : null;
      if (!(button instanceof HTMLElement)) return;
      const label = actionLabels[button.id];
      if (!label) return;
      markBusy(button, label);
    }, true);
  }

  function installGlobalErrorFeedback() {
    window.addEventListener('unhandledrejection', (event) => {
      const reason = event?.reason;
      const message = reason?.message || reason || '操作失败：发生未处理的异步异常';
      clearBusy();
      showToast(`操作失败：${normalizeMessage(message)}`, 'error');
    });

    window.addEventListener('error', (event) => {
      if (!event || (!event.error && !event.message)) return;
      const message = event.error?.message || event.message || '页面执行异常';
      clearBusy();
      showToast(`页面执行异常：${normalizeMessage(message)}`, 'error');
    });
  }

  function init() {
    ensureRoot();
    installActionFeedback();
    installStatusObserver();
    installGlobalErrorFeedback();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init, { once: true });
  } else {
    init();
  }
})();
