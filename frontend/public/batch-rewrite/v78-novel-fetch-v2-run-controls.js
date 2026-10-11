(() => {
  const API_ROOT = '/api/batch-rewrite';
  const STYLE_ID = 'v78RunControlsStyles';
  const METHOD_OPTIONS = [
    ['', '自动轮换'],
    ['instruction', '指令改文'],
    ['opening_instruction', '开头词+指令'],
    ['high_imitation', '高仿文章']
  ];
  let mounted = false;
  let fetchPatched = false;

  const $ = id => document.getElementById(id);
  const text = value => String(value == null ? '' : value);

  function headers(extra = {}) {
    const token = localStorage.getItem('auth_token') || '';
    return { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...extra };
  }

  async function api(path, options = {}) {
    const response = await fetch(`${API_ROOT}${path}`, { ...options, headers: headers(options.headers || {}) });
    const raw = await response.text();
    let data = {};
    try { data = raw ? JSON.parse(raw) : {}; } catch (_) { data = { raw }; }
    if (!response.ok) throw new Error(data.error || data.message || data.raw || `HTTP ${response.status}`);
    return data;
  }

  function installStyles() {
    if ($(STYLE_ID)) return;
    const style = document.createElement('style');
    style.id = STYLE_ID;
    style.textContent = `
      #v78TargetVersions .v78-version-item small{display:none!important}
      .v78-run-ai-label{margin:8px 0 4px;font-size:12px;color:var(--muted)}
      #v78RunMethodGrid{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:8px;margin:10px 0 2px}
      #v78RunMethodGrid label{display:grid;gap:6px;min-width:0;font-size:11px;color:var(--muted)}
      #v78RunMethodGrid select{width:100%;min-height:34px;padding:6px 8px;border-radius:6px}
      #v78RunMethodGrid select:disabled{opacity:.48;cursor:not-allowed}
      #v78BatchInputControls .v78-run-text-model{display:grid;gap:4px;min-width:0;font-size:12px;color:var(--muted)}
      #v78BatchInputControls .v78-run-text-model select{width:100%;min-height:34px;min-width:0;padding:6px 8px;border-radius:6px}
      #v78BatchInputControls .v78-run-text-model small{min-height:14px;font-size:11px;line-height:14px}
      #v78ScheduleBtn,#v78ScheduleManagerBtn,#v78SensitiveRepairBtn{min-height:37px}
      #v78ScheduleDialog,#v78ScheduleListDialog{border:1px solid var(--line);border-radius:10px;background:var(--panel);color:var(--text);padding:0;box-shadow:0 24px 80px rgba(0,0,0,.5)}
      #v78ScheduleDialog{width:min(420px,calc(100vw - 32px))}
      #v78ScheduleListDialog{width:min(760px,calc(100vw - 32px));max-height:min(76vh,720px)}
      #v78ScheduleDialog::backdrop,#v78ScheduleListDialog::backdrop{background:rgba(0,0,0,.58)}
      #v78ScheduleDialog .v78-schedule-form,#v78ScheduleListDialog .v78-schedule-manager{display:grid;gap:12px;padding:18px}
      #v78ScheduleDialog h3,#v78ScheduleListDialog h3{margin:0;color:var(--v78-blue,#1677ff);font-size:17px}
      #v78ScheduleDialog label{display:grid;gap:7px;font-size:12px;color:var(--muted)}
      #v78ScheduleDialog input{min-height:38px;padding:7px 9px;border:1px solid var(--line);border-radius:6px;background:var(--panel-soft);color:var(--text)}
      #v78ScheduleDialog .v78-schedule-actions,#v78ScheduleListDialog .v78-schedule-manager-actions{display:flex;justify-content:flex-end;gap:8px;flex-wrap:wrap}
      #v78ScheduleResult,#v78ScheduleListResult{min-height:18px;font-size:12px;color:var(--muted)}
      #v78ScheduleList{display:grid;gap:8px;max-height:52vh;overflow:auto;padding-right:2px}
      .v78-schedule-row{display:grid;grid-template-columns:minmax(150px,1.2fr) minmax(120px,1fr) minmax(86px,.65fr) auto;align-items:center;gap:10px;padding:10px 12px;border:1px solid var(--line);border-radius:8px;background:var(--panel-soft)}
      .v78-schedule-row strong{font-size:12px;color:var(--text);font-weight:600}
      .v78-schedule-row small{font-size:11px;color:var(--muted)}
      .v78-schedule-row .v78-schedule-row-actions{display:flex;gap:6px;justify-content:flex-end;flex-wrap:wrap}
      .v78-schedule-row button{min-height:30px;padding:4px 9px;font-size:11px}
      .v78-schedule-empty{padding:26px 12px;text-align:center;border:1px dashed var(--line);border-radius:8px;color:var(--muted);font-size:12px}
      @media(max-width:900px){#v78RunMethodGrid{grid-template-columns:repeat(2,minmax(0,1fr))}.v78-schedule-row{grid-template-columns:1fr}.v78-schedule-row .v78-schedule-row-actions{justify-content:flex-start}}
    `;
    document.head.appendChild(style);
  }

  function configuredMethods() {
    try {
      if (typeof state === 'object') {
        const methods = state?.config?.app_config?.rewrite?.ai_slot_methods;
        if (methods && typeof methods === 'object') return { ...methods };
      }
    } catch (_) {}
    return {};
  }

  function collectRunMethods() {
    const result = {};
    for (let index = 1; index <= 5; index += 1) {
      const value = text($(`v78RunAiMethod${index}`)?.value).trim();
      if (value) result[`ai${index}`] = value;
    }
    return result;
  }

  function syncMethodsIntoV78() {
    try {
      if (typeof state !== 'object' || !state) return;
      state.config = state.config || {};
      state.config.app_config = state.config.app_config || {};
      state.config.app_config.rewrite = state.config.app_config.rewrite || {};
      state.config.app_config.rewrite.ai_slot_methods = collectRunMethods();
    } catch (_) {}
  }

  function applyMethodSnapshot(methods = {}) {
    for (let index = 1; index <= 5; index += 1) {
      const select = $(`v78RunAiMethod${index}`);
      if (!select) continue;
      const value = text(methods[`ai${index}`]).trim();
      select.value = [...select.options].some(option => option.value === value) ? value : '';
    }
    syncMethodsIntoV78();
    syncMethodAvailability();
  }

  function selectedTargets() {
    const versions = [];
    if ($('v78TargetOriginal')?.checked) versions.push('original');
    for (let index = 1; index <= 5; index += 1) {
      if ($(`v78TargetAi${index}`)?.checked) versions.push(`ai${index}`);
    }
    return versions;
  }

  function selectedPreviewIds() {
    const inputs = [...document.querySelectorAll('[data-v78-preview-id]')];
    if (!inputs.length) return [];
    return inputs.filter(input => input.checked).map(input => text(input.dataset.v78PreviewId).trim()).filter(Boolean);
  }

  function buildCurrentRunSnapshot() {
    const targets = selectedTargets();
    return {
      platform_id: $('platformSelect')?.value || '2',
      parse_mode: $('parseModeSelect')?.value || 'smart',
      column_preset_id: $('columnPresetSelect')?.value || '',
      column_order: $('columnOrderInput')?.value || '书籍ID,书名,推荐理由,男女频,标签,评级',
      input_text: $('inputText')?.value || '',
      max_txt: Number($('v78RunMaxTxt')?.value || $('fetchMaxTxt')?.value || 4000) || 4000,
      text_model_id: $('v78RunTextModel')?.value || '',
      target_versions: targets,
      targetVersions: targets,
      ai_slot_methods_snapshot: collectRunMethods(),
      task_ids: selectedPreviewIds(),
      sensitive_ai_enabled: $('sensitiveAiProcessEnabled')?.checked === true
    };
  }

  async function loadRunTextModels(select, status) {
    try {
      const response = await fetch('/api/models?kind=text', { headers: headers() });
      const body = await response.json();
      if (!response.ok) throw new Error(body.error || `HTTP ${response.status}`);
      const models = Array.isArray(body.models) ? body.models : [];
      select.innerHTML = '';
      if (!models.length) {
        select.innerHTML = '<option value="">暂无可用文本模型</option>';
        status.textContent = '请先在 API 配置中启用文本模型';
        return;
      }
      for (const model of models) {
        const option = document.createElement('option');
        option.value = model.id || model.modelId || '';
        option.textContent = model.displayName || model.name || model.modelId || model.id;
        select.appendChild(option);
      }
      const saved = text(state?.config?.app_config?.text_model_id || state?.config?.text_model_id || '');
      select.value = models.some(model => text(model.id || model.modelId) === saved) ? saved : select.options[0].value;
      status.textContent = '执行时使用此模型；修改后将作为后续批次默认值';
    } catch (error) {
      select.innerHTML = '<option value="">读取文本模型失败</option>';
      status.textContent = `读取文本模型失败：${error.message || '请稍后重试'}`;
    }
  }

  function mountTextModelSelector() {
    const controls = $('v78BatchInputControls');
    if (!controls) return false;
    if ($('v78RunTextModel')) return true;
    const label = document.createElement('label');
    label.className = 'v78-run-text-model';
    label.textContent = '文本模型';
    const select = document.createElement('select');
    select.id = 'v78RunTextModel';
    select.innerHTML = '<option value="">正在读取模型...</option>';
    const status = document.createElement('small');
    status.id = 'v78RunTextModelStatus';
    label.append(select, status);
    controls.appendChild(label);
    select.addEventListener('change', async () => {
      try {
        if (typeof window.persistSelectedTextModel !== 'function') throw new Error('文本模型保存功能尚未就绪');
        await window.persistSelectedTextModel(select.value);
        status.textContent = '已保存为后续批次默认文本模型';
      } catch (error) {
        status.textContent = `保存文本模型失败：${error.message || '请稍后重试'}`;
      }
    });
    void loadRunTextModels(select, status);
    return true;
  }

  function syncMethodAvailability() {
    for (let index = 1; index <= 5; index += 1) {
      const checkbox = $(`v78TargetAi${index}`);
      const select = $(`v78RunAiMethod${index}`);
      if (!select) continue;
      select.disabled = checkbox ? !checkbox.checked : false;
      if (checkbox && !checkbox.dataset.v78RunMethodBound) {
        checkbox.addEventListener('change', syncMethodAvailability);
        checkbox.dataset.v78RunMethodBound = '1';
      }
    }
  }

  function mountMethodSelectors() {
    const target = $('v78TargetVersions');
    const versionGrid = $('v78CompactTargets') || target?.querySelector('.v78-version-grid');
    if (!target || !versionGrid) return false;
    if (!$('v78RunAiLabel')) {
      const label = document.createElement('div');
      label.id = 'v78RunAiLabel';
      label.className = 'v78-run-ai-label';
      label.textContent = 'AI文案（选择本次需要生成的版本）';
      versionGrid.insertAdjacentElement('beforebegin', label);
    }
    if (!$('v78RunMethodGrid')) {
      const grid = document.createElement('div');
      grid.id = 'v78RunMethodGrid';
      for (let index = 1; index <= 5; index += 1) {
        const label = document.createElement('label');
        label.textContent = `AI${index} 改文方式`;
        const select = document.createElement('select');
        select.id = `v78RunAiMethod${index}`;
        for (const [value, name] of METHOD_OPTIONS) {
          const option = document.createElement('option');
          option.value = value;
          option.textContent = name;
          select.appendChild(option);
        }
        select.addEventListener('change', syncMethodsIntoV78);
        label.appendChild(select);
        grid.appendChild(label);
      }
      versionGrid.insertAdjacentElement('afterend', grid);
      applyMethodSnapshot(configuredMethods());
    }
    syncMethodAvailability();
    return true;
  }

  function setWorkMessage(message) {
    const preview = $('v78PreviewStatus');
    if (preview) preview.textContent = message;
    const result = $('processResult');
    if (result && !result.textContent.trim()) result.textContent = message;
    const tone = /失败|错误|异常|超时/.test(text(message)) ? 'error' : /正在|处理中|验证|读取|保存/.test(text(message)) ? 'working' : /完成|成功|已创建|已删除|已取消/.test(text(message)) ? 'success' : 'info';
    if (typeof window.qiantiePublishNovelFetchStatus === 'function') window.qiantiePublishNovelFetchStatus(message, tone);
    else if (window.parent && window.parent !== window) window.parent.postMessage({ type: 'qiantie:novel-fetch-status', text: text(message), tone }, window.location.origin);
  }

  function ensureScheduleDialog() {
    if ($('v78ScheduleDialog')) return $('v78ScheduleDialog');
    const dialog = document.createElement('dialog');
    dialog.id = 'v78ScheduleDialog';
    dialog.innerHTML = `
      <form method="dialog" class="v78-schedule-form">
        <h3>定时处理</h3>
        <label>开始时间<input id="v78ScheduleRunAt" type="datetime-local" required></label>
        <div id="v78ScheduleSummary" class="v78-run-ai-label"></div>
        <div id="v78ScheduleResult"></div>
        <div class="v78-schedule-actions"><button value="cancel">取消</button><button id="v78ScheduleConfirm" value="default" class="primary">确认定时</button></div>
      </form>`;
    document.body.appendChild(dialog);
    dialog.addEventListener('submit', async event => {
      if (event.submitter?.id !== 'v78ScheduleConfirm') return;
      event.preventDefault();
      const result = $('v78ScheduleResult');
      const runAtValue = $('v78ScheduleRunAt')?.value || '';
      const runAt = new Date(runAtValue);
      if (!Number.isFinite(runAt.getTime())) { result.textContent = '请选择处理时间'; return; }
      if (runAt.getTime() <= Date.now()) { result.textContent = '定时时间需要晚于现在'; return; }
      const snapshot = buildCurrentRunSnapshot();
      if (!text(snapshot.input_text).trim()) { result.textContent = '请先填写批量输入'; return; }
      if (!snapshot.target_versions.length) { result.textContent = '请至少选择一个文案版本'; return; }
      const previewInputs = [...document.querySelectorAll('[data-v78-preview-id]')];
      if (previewInputs.length && !snapshot.task_ids.length) { result.textContent = '请至少选择一本小说'; return; }
      const button = $('v78ScheduleConfirm');
      button.disabled = true;
      result.textContent = '正在保存定时任务...';
      try {
        await api('/schedules', {
          method: 'POST',
          body: JSON.stringify({ runAt: runAt.toISOString(), inputSnapshot: snapshot })
        });
        result.textContent = `已定时：${runAt.toLocaleString()}`;
        setWorkMessage('定时完成');
        window.setTimeout(() => dialog.close(), 650);
      } catch (error) {
        result.textContent = error.message || '定时处理创建失败';
        setWorkMessage(`定时失败：${error.message || '定时处理创建失败'}`);
      } finally {
        button.disabled = false;
      }
    });
    return dialog;
  }

  function openScheduleDialog() {
    const snapshot = buildCurrentRunSnapshot();
    if (!text(snapshot.input_text).trim()) { setWorkMessage('请先填写批量输入'); return; }
    if (!snapshot.target_versions.length) { setWorkMessage('请至少选择一个文案版本'); return; }
    setWorkMessage('正在定时中....');
    const dialog = ensureScheduleDialog();
    const date = new Date(Date.now() + 10 * 60 * 1000);
    const pad = value => String(value).padStart(2, '0');
    $('v78ScheduleRunAt').value = `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
    $('v78ScheduleSummary').textContent = `本次版本：${snapshot.target_versions.map(version => version === 'original' ? '原文' : version.toUpperCase()).join('、')}`;
    $('v78ScheduleResult').textContent = '';
    dialog.showModal();
  }

  function scheduleStatusLabel(item = {}) {
    if (item.enabled === false || item.status === 'cancelled') return '已取消';
    const labels = { scheduled: '等待执行', running: '处理中', done: '已完成', failed: '失败' };
    return labels[item.status] || (/^[a-z0-9_:-]+$/i.test(text(item.status)) ? '等待处理' : text(item.status || '等待执行'));
  }

  function scheduleTargetLabel(item = {}) {
    const versions = Array.isArray(item?.inputSnapshot?.target_versions)
      ? item.inputSnapshot.target_versions
      : (Array.isArray(item?.inputSnapshot?.targetVersions) ? item.inputSnapshot.targetVersions : []);
    if (!versions.length) return '按保存配置';
    return versions.map(version => version === 'original' ? '原文' : text(version).toUpperCase()).join('、');
  }

  function ensureScheduleListDialog() {
    if ($('v78ScheduleListDialog')) return $('v78ScheduleListDialog');
    const dialog = document.createElement('dialog');
    dialog.id = 'v78ScheduleListDialog';
    dialog.innerHTML = `
      <div class="v78-schedule-manager">
        <h3>定时任务管理</h3>
        <div class="v78-run-ai-label">查看已保存的定时处理；等待中的任务可以取消，历史记录可以删除。</div>
        <div class="v78-schedule-manager-actions"><button id="v78ScheduleListRefresh" type="button">刷新定时任务</button><button id="v78ScheduleListClose" type="button">关闭</button></div>
        <div id="v78ScheduleListResult"></div>
        <div id="v78ScheduleList"></div>
      </div>`;
    document.body.appendChild(dialog);
    $('v78ScheduleListRefresh').addEventListener('click', () => void loadSchedules());
    $('v78ScheduleListClose').addEventListener('click', () => dialog.close());
    return dialog;
  }

  function renderScheduleList(schedules = []) {
    const list = $('v78ScheduleList');
    if (!list) return;
    list.innerHTML = '';
    if (!schedules.length) {
      const empty = document.createElement('div');
      empty.className = 'v78-schedule-empty';
      empty.textContent = '暂无定时任务';
      list.appendChild(empty);
      return;
    }
    for (const item of schedules) {
      const row = document.createElement('div');
      row.className = 'v78-schedule-row';

      const when = document.createElement('strong');
      const date = new Date(item.runAt);
      when.textContent = Number.isFinite(date.getTime()) ? date.toLocaleString() : text(item.runAt || '-');

      const versions = document.createElement('small');
      versions.textContent = `版本：${scheduleTargetLabel(item)}`;

      const status = document.createElement('small');
      status.textContent = scheduleStatusLabel(item);

      const upload = document.createElement('small');
      upload.textContent = item.autoSubmit === false ? '上传：需人工提交' : '上传：自动上传';

      const actions = document.createElement('div');
      actions.className = 'v78-schedule-row-actions';
      if (item.enabled !== false && item.status === 'scheduled') {
        const cancel = document.createElement('button');
        cancel.type = 'button';
        cancel.textContent = '取消定时';
        cancel.addEventListener('click', () => void cancelSchedule(item.id));
        actions.appendChild(cancel);
      }
      const remove = document.createElement('button');
      remove.type = 'button';
      remove.textContent = '删除记录';
      remove.addEventListener('click', () => void deleteSchedule(item.id));
      actions.appendChild(remove);

      row.append(when, versions, upload, status, actions);
      list.appendChild(row);
    }
  }

  async function loadSchedules() {
    const result = $('v78ScheduleListResult');
    if (result) result.textContent = '正在读取定时任务...';
    try {
      const data = await api('/schedules');
      const schedules = Array.isArray(data?.schedules) ? data.schedules : [];
      renderScheduleList(schedules);
      if (result) result.textContent = `共 ${schedules.length} 条定时任务`;
      setWorkMessage(`已读取定时任务：${schedules.length} 条`);
      return schedules;
    } catch (error) {
      renderScheduleList([]);
      if (result) result.textContent = error.message || '定时任务读取失败';
      setWorkMessage(`定时任务读取失败：${error.message || '请稍后重试'}`);
      return [];
    }
  }

  async function cancelSchedule(id) {
    const result = $('v78ScheduleListResult');
    if (result) result.textContent = '正在取消定时任务...';
    try {
      await api(`/schedules/${encodeURIComponent(text(id))}`, {
        method: 'PATCH',
        body: JSON.stringify({ enabled: false, status: 'cancelled' })
      });
      if (result) result.textContent = '定时任务已取消';
      setWorkMessage('定时任务已取消');
      await loadSchedules();
    } catch (error) {
      if (result) result.textContent = error.message || '取消定时任务失败';
      setWorkMessage(`取消定时任务失败：${error.message || '请稍后重试'}`);
    }
  }

  async function deleteSchedule(id) {
    if (!window.confirm('确定删除这条定时记录吗？')) return;
    const result = $('v78ScheduleListResult');
    if (result) result.textContent = '正在删除定时记录...';
    try {
      await api(`/schedules/${encodeURIComponent(text(id))}`, { method: 'DELETE' });
      if (result) result.textContent = '定时记录已删除';
      setWorkMessage('定时记录已删除');
      await loadSchedules();
    } catch (error) {
      if (result) result.textContent = error.message || '删除定时记录失败';
      setWorkMessage(`删除定时记录失败：${error.message || '请稍后重试'}`);
    }
  }

  function openScheduleManager() {
    const dialog = ensureScheduleListDialog();
    dialog.showModal();
    void loadSchedules();
  }

  async function repairCurrentBatchSensitive() {
    const button = $('v78SensitiveRepairBtn');
    if (button) button.disabled = true;
    setWorkMessage('正在进行敏感词 AI 修复...');
    try {
      const current = await api('/batches/current');
      const ids = Array.isArray(current?.batch?.taskIds) ? current.batch.taskIds.map(text).filter(Boolean) : [];
      if (!ids.length) throw new Error('当前批次还没有可处理的小说');
      const result = await api('/tasks/reprocess-sensitive', {
        method: 'POST',
        body: JSON.stringify({ mode: 'selected', ids, restore_from_backup: false, sensitive_ai_enabled: true })
      });
      setWorkMessage(`敏感词 AI 修复完成：已处理 ${Number(result.processed || 0)} 本，失败 ${Number(result.failed || 0)} 本`);
      if (typeof loadTasks === 'function') await loadTasks();
      $('v78CurrentBatchRefresh')?.click();
    } catch (error) {
      setWorkMessage(error.message || '敏感词 AI 修复失败');
    } finally {
      if (button) button.disabled = false;
    }
  }

  function mountActionButtons() {
    const primary = $('v78WorkPrimaryActions');
    const processBtn = $('processBtn');
    const refreshBtn = $('refreshBtn');
    if (!primary || !processBtn) return false;
    if (!$('v78ScheduleBtn')) {
      const button = document.createElement('button');
      button.id = 'v78ScheduleBtn';
      button.type = 'button';
      button.textContent = '开始定时';
      button.addEventListener('click', openScheduleDialog);
      processBtn.insertAdjacentElement('afterend', button);
    }
    if (!$('v78ScheduleManagerBtn')) {
      const button = document.createElement('button');
      button.id = 'v78ScheduleManagerBtn';
      button.type = 'button';
      button.textContent = '定时任务';
      button.addEventListener('click', openScheduleManager);
      $('v78ScheduleBtn').insertAdjacentElement('afterend', button);
    }
    if (!$('v78SensitiveRepairBtn')) {
      const button = document.createElement('button');
      button.id = 'v78SensitiveRepairBtn';
      button.type = 'button';
      button.textContent = '敏感词 AI 修复';
      button.addEventListener('click', () => void repairCurrentBatchSensitive());
      $('v78ScheduleManagerBtn').insertAdjacentElement('afterend', button);
    }
    if (refreshBtn && refreshBtn.parentElement === primary) primary.appendChild(refreshBtn);
    return true;
  }

  function patchProcessRequests() {
    if (fetchPatched) return;
    const nativeFetch = window.fetch.bind(window);
    window.fetch = async function(input, init = {}) {
      const url = typeof input === 'string' ? input : text(input?.url);
      let nextInit = init;
      if (/\/api\/batch-rewrite\/process\/start(?:\?|$)/.test(url) && text(init.method || 'GET').toUpperCase() === 'POST' && typeof init.body === 'string') {
        try {
          const body = JSON.parse(init.body);
          const targets = selectedTargets();
          if (targets.length) {
            body.target_versions = targets;
            body.targetVersions = targets;
          }
          body.ai_slot_methods_snapshot = collectRunMethods();
          nextInit = { ...init, body: JSON.stringify(body) };
        } catch (_) {}
      }
      const response = await nativeFetch(input, nextInit);
      if (/\/api\/batch-rewrite\/batches\/[^/]+\/rerun(?:\?|$)/.test(url) && response.ok) {
        response.clone().json().then(data => {
          const methods = data?.payload?.ai_slot_methods_snapshot;
          if (methods && typeof methods === 'object') window.setTimeout(() => applyMethodSnapshot(methods), 0);
        }).catch(() => {});
      }
      return response;
    };
    fetchPatched = true;
  }

  function mount() {
    installStyles();
    patchProcessRequests();
    const textModelReady = mountTextModelSelector();
    const methodsReady = mountMethodSelectors();
    const actionsReady = mountActionButtons();
    mounted = textModelReady && methodsReady && actionsReady;
    return mounted;
  }

  function boot() {
    mount();
    let attempts = 0;
    const timer = window.setInterval(() => {
      mount();
      attempts += 1;
      if (mounted || attempts >= 40) window.clearInterval(timer);
    }, 250);
  }

  window.v78NovelFetchRunControls = {
    buildCurrentRunSnapshot,
    collectRunMethods,
    repairCurrentBatchSensitive,
    openScheduleDialog,
    openScheduleManager,
    loadSchedules,
    cancelSchedule,
    deleteSchedule
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot, { once: true });
  else boot();
  window.addEventListener('load', mount, { once: true });
})();
