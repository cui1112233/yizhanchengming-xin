// 版本对应配置档：单一权威保存链路。
// app.js 保留通用配置/网站提交保存入口；本文件只接管“保存版本配置”按钮，
// 将版本选择、AI 槽位、提示词与 121 配置档绑定合并为一次 /api/config 持久化。

async function saveVersionConfigAuthority() {
  clearTimeout(workFormSaveTimer);
  workFormSaveTimer = null;

  syncVersionPromptConfigToForm();
  const formState = collectWorkFormState();
  const appConfig = syncFormToAppConfig();
  appConfig.work_form = formState;
  appConfig.web_submit = syncFormToWebSubmitConfig();
  appConfig.knowledge = state.config?.knowledge || {};

  const result = await api('/api/config', {
    method: 'POST',
    body: JSON.stringify({ app_config: appConfig }),
  });
  if (!result?.config) throw new Error('保存成功但服务端未返回最新版本配置');

  // 服务端回读结果是保存后的唯一权威状态，禁止继续使用旧 state.config。
  state.config = result.config;
  try {
    localStorage.setItem(WORK_FORM_STORAGE_KEY, JSON.stringify(formState));
  } catch {
    // localStorage 不可用时仍以服务端配置为准。
  }
  return result.config;
}

async function confirmWebSubmitSelectionAuthority() {
  const status = $('webSubmitSelectionStatus');
  const versions = selectedProcessVersions();
  if (!versions.length) {
    if (status) status.textContent = '请至少选择一个文案版本';
    return;
  }

  if (status) status.textContent = '正在保存版本配置...';
  const button = $('confirmWebSubmitSelectionBtn');
  if (button) button.disabled = true;
  try {
    await saveVersionConfigAuthority();
    updateVersionConfigSummary();
    if (status) status.textContent = '版本配置已保存';
    window.setTimeout(closeVersionConfigCard, 120);
  } catch (error) {
    if (status) status.textContent = error.message || '版本配置保存失败';
  } finally {
    if (button) button.disabled = false;
  }
}

function bindVersionConfigAuthority() {
  const button = $('confirmWebSubmitSelectionBtn');
  if (button) button.onclick = confirmWebSubmitSelectionAuthority;
}

// app.js 初始化后可能绑定旧处理器；当前脚本由 V2 页面层最后注入，重新覆盖一次。
bindVersionConfigAuthority();
window.setTimeout(bindVersionConfigAuthority, 0);
