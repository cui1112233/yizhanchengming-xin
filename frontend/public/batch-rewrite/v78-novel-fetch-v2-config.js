(() => {
  const API_ROOT = '/api/batch-rewrite';
  let sensitiveFixGuardInstalled = false;

  function byId(id) { return document.getElementById(id); }
  function object(value) { return value && typeof value === 'object' && !Array.isArray(value) ? value : {}; }
  function tokenHeaders(extra = {}) {
    const token = localStorage.getItem('auth_token') || '';
    return { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...extra };
  }
  async function v2Api(path, options = {}) {
    if (window.QiantieNovelFetchRuntime) return window.QiantieNovelFetchRuntime.api(path, options);
    const response = await fetch(`${API_ROOT}${path}`, { ...options, headers: tokenHeaders(options.headers || {}) });
    const text = await response.text();
    let data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (_) { data = { raw: text }; }
    if (!response.ok) throw new Error(data.error || data.message || data.raw || `HTTP ${response.status}`);
    return data;
  }
  function deepMergeConfig(base, patch) {
    const out = { ...object(base) };
    for (const [key, value] of Object.entries(object(patch))) {
      out[key] = object(value) && object(out[key])
        ? deepMergeConfig(out[key], value)
        : (Array.isArray(value) ? value.slice() : value);
    }
    return out;
  }
  function setValue(id, value) { const node = byId(id); if (node) node.value = value ?? ''; }
  function setChecked(id, value) { const node = byId(id); if (node) node.checked = value === true; }
  function numberValue(id, fallback, min, max) {
    const number = Number(byId(id)?.value);
    if (!Number.isFinite(number)) return fallback;
    return Math.max(min, Math.min(max, Math.floor(number)));
  }
  function setStatus(text, error = false) {
    const node = byId('v78AdvancedConfigStatus');
    if (!node) return;
    node.textContent = String(text || '');
    node.classList.toggle('v78-v2-error', error);
    node.classList.toggle('v78-v2-ok', !error && Boolean(text));
  }

  function ensureSensitiveFixSelector() {
    if (byId('sensitiveFixSelect')) return byId('sensitiveFixSelect');
    const rewriteSelect = byId('rewriteSelect');
    const grid = rewriteSelect?.closest('.control-grid') || rewriteSelect?.parentElement?.parentElement;
    if (!grid) return null;
    const label = document.createElement('label');
    label.innerHTML = '敏感词修复模型<select id="sensitiveFixSelect"></select>';
    grid.appendChild(label);
    return byId('sensitiveFixSelect');
  }

  function installSensitiveFixLegacyGuard() {
    if (sensitiveFixGuardInstalled) return true;
    const legacy = window.syncCurrentAiPreset;
    if (typeof legacy !== 'function') return false;
    window.syncCurrentAiPreset = function(cfg) {
      const assignments = object(cfg?.ai_assignments);
      const explicit = String(byId('sensitiveFixSelect')?.value || assignments.sensitive_fix || '__current__');
      const result = legacy.apply(this, arguments);
      cfg.ai_assignments = object(cfg.ai_assignments);
      if (explicit && explicit !== '__current__') cfg.ai_assignments.sensitive_fix = explicit;
      return result;
    };
    sensitiveFixGuardInstalled = true;
    return true;
  }

  function mountAdvancedConfigPanel() {
    if (byId('v78AdvancedConfigPanel')) return;
    const grid = document.querySelector('#config .config-grid');
    if (!grid) return;
    ensureSensitiveFixSelector();
    const section = document.createElement('section');
    section.id = 'v78AdvancedConfigPanel';
    section.className = 'section';
    section.innerHTML = `
      <div class="section-title"><h2>自动处理</h2><span>补充设置会保存到当前小说获取配置。</span></div>
      <div class="control-grid two">
        <label>原文最少字数<input id="v78MinOriginalChars" type="number" min="0" max="1000000" value="0" /></label>
        <label class="checkbox-label"><input id="v78SkipShortOriginal" type="checkbox" /> 低于最少字数时保留原文并跳过改文</label>
        <label class="checkbox-label"><input id="v78AutoReclassifyStyle" type="checkbox" /> 风格不在当前目录时重新 AI 判断</label>
        <label class="checkbox-label"><input id="v78AutoSyncStyles" type="checkbox" /> 处理前自动同步 121 风格</label>
        <label class="checkbox-label"><input id="v78CleanupEnabled" type="checkbox" /> 自动清理过期已完成任务</label>
        <label>任务保留天数<input id="v78RetentionDays" type="number" min="1" max="3650" value="30" /></label>
        <label>自动提交每批任务数<input id="v78SubmitBatchSize" type="number" min="1" max="500" value="20" /></label>
        <label>自动提交等待秒数<input id="v78SubmitFlushSeconds" type="number" min="0" max="300" value="0" /></label>
        <label class="checkbox-label"><input id="v78ForceSerialBatch" type="checkbox" /> AI 改文批量按顺序处理</label>
        <label>敏感词处理模式
          <select id="v78SensitiveMode">
            <option value="replace">直接替换</option>
            <option value="ai_each">AI逐条修改</option>
            <option value="ai_group">AI整组修改</option>
          </select>
        </label>
      </div>
      <div class="notice">没有有效的 121 登录时会跳过风格同步，不影响原文和 AI 文案处理。</div>
      <div class="actions"><button id="v78SaveAdvancedConfig" class="primary">保存配置</button><button id="v78ReloadAdvancedConfig">重新读取</button><span id="v78AdvancedConfigStatus"></span></div>`;
    grid.appendChild(section);
    byId('v78SaveAdvancedConfig').onclick = saveAdvancedConfig;
    byId('v78ReloadAdvancedConfig').onclick = loadAdvancedConfig;
  }

  function renderPresetOptions(appConfig) {
    const select = ensureSensitiveFixSelector();
    if (!select) return;
    const presets = Array.isArray(appConfig.ai_presets) ? appConfig.ai_presets : [];
    const current = String(object(appConfig.ai_assignments).sensitive_fix || '__current__');
    select.innerHTML = '';
    const currentOption = document.createElement('option');
    currentOption.value = '__current__';
    currentOption.textContent = '跟随当前 AI 配置';
    select.appendChild(currentOption);
    for (const preset of presets) {
      const id = String(preset?.id || '').trim();
      if (!id) continue;
      const option = document.createElement('option');
      option.value = id;
      option.textContent = preset?.name || id;
      select.appendChild(option);
    }
    select.value = [...select.options].some(option => option.value === current) ? current : '__current__';
  }

  function renderAdvancedConfig(appConfig) {
    const app = object(appConfig);
    const fetchCfg = object(app.fetch);
    const workflow = object(app.workflow);
    const storage = object(app.storage);
    const webSubmit = object(app.web_submit);
    const ai = object(app.ai);
    const sensitive = object(app.sensitive_ai);
    setValue('v78MinOriginalChars', Number(fetchCfg.min_original_chars) || 0);
    setChecked('v78SkipShortOriginal', fetchCfg.skip_short_original === true);
    setChecked('v78AutoReclassifyStyle', workflow.auto_reclassify_invalid_style === true);
    setChecked('v78AutoSyncStyles', workflow.auto_sync_site_styles === true);
    setChecked('v78CleanupEnabled', storage.cleanup_enabled === true);
    setValue('v78RetentionDays', Number(storage.retention_days) || 30);
    setValue('v78SubmitBatchSize', Number(webSubmit.batch_size) || 20);
    setValue('v78SubmitFlushSeconds', Number(webSubmit.flush_seconds) || 0);
    setChecked('v78ForceSerialBatch', ai.force_serial_batch === true);
    setValue('v78SensitiveMode', ['replace', 'ai_each', 'ai_group'].includes(String(sensitive.mode || '')) ? sensitive.mode : (sensitive.enabled === false ? 'replace' : 'ai_each'));
    renderPresetOptions(app);
  }

  async function loadAdvancedConfig() {
    try {
      setStatus('正在读取...');
      const data = await v2Api('/config');
      if (typeof state === 'object' && data && data.app_config) {
        state.config = { ...(state.config || {}), ...data, knowledge_loaded: true };
      }
      renderAdvancedConfig(object(data.app_config));
      setStatus('已读取服务端配置');
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function saveAdvancedConfig() {
    try {
      setStatus('正在保存...');
      const data = await v2Api('/config');
      const current = object(data.app_config);
      const selectedSensitiveModel = String(byId('sensitiveFixSelect')?.value || '__current__');
      const patch = {
        fetch: {
          min_original_chars: numberValue('v78MinOriginalChars', 0, 0, 1000000),
          skip_short_original: byId('v78SkipShortOriginal')?.checked === true
        },
        workflow: {
          auto_reclassify_invalid_style: byId('v78AutoReclassifyStyle')?.checked === true,
          auto_sync_site_styles: byId('v78AutoSyncStyles')?.checked === true
        },
        storage: {
          cleanup_enabled: byId('v78CleanupEnabled')?.checked === true,
          retention_days: numberValue('v78RetentionDays', 30, 1, 3650)
        },
        web_submit: {
          batch_size: numberValue('v78SubmitBatchSize', 20, 1, 500),
          flush_seconds: numberValue('v78SubmitFlushSeconds', 0, 0, 300)
        },
        ai: {
          force_serial_batch: byId('v78ForceSerialBatch')?.checked === true
        },
        sensitive_ai: {
          mode: String(byId('v78SensitiveMode')?.value || 'ai_each')
        },
        ai_assignments: {
          sensitive_fix: selectedSensitiveModel
        }
      };
      const merged = deepMergeConfig(current, patch);
      await v2Api('/config', { method: 'POST', body: JSON.stringify({ app_config: merged }) });
      renderAdvancedConfig(merged);
      setStatus('配置已保存');
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  function boot() {
    if (window.__qiantieNovelFetchConfigBooted) return;
    window.__qiantieNovelFetchConfigBooted = true;
    mountAdvancedConfigPanel();
    ensureSensitiveFixSelector();
    installSensitiveFixLegacyGuard();
    document.addEventListener('click', event => {
      const tab = event.target.closest?.('.tab[data-tab="config"]');
      if (tab) void loadAdvancedConfig();
    }, true);
    let attempts = 0;
    const timer = window.setInterval(() => {
      mountAdvancedConfigPanel();
      ensureSensitiveFixSelector();
      installSensitiveFixLegacyGuard();
      attempts += 1;
      if (attempts >= 20) window.clearInterval(timer);
    }, 250);
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot, { once: true });
  else boot();
})();
