(() => {
  const API_ROOT = '/api/batch-rewrite';
  const taskFilters = { date: '', bookId: '', status: '', scope: '' };
  const previewState = { active: false, tasks: [], selected: new Set(), originalInput: '' };
  let rerunSourceBatchId = '';
  let legacyTaskBridgeInstalled = false;
  let retryProgressBridgeInstalled = false;
  let currentBatchTimer = null;
  const processLogState = { lines: [] };
  const seenProgressKeys = new Set();

  function byId(id) { return document.getElementById(id); }
  function value(id, fallback = '') { return byId(id)?.value ?? fallback; }
  function asArray(value) { return Array.isArray(value) ? value : []; }
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
  function publishStatus(message, tone = 'info') {
    if (typeof window.qiantiePublishNovelFetchStatus === 'function') {
      window.qiantiePublishNovelFetchStatus(message, tone);
      return;
    }
    const text = String(message || '').trim();
    if (text && window.parent && window.parent !== window) window.parent.postMessage({ type: 'qiantie:novel-fetch-status', text, tone }, window.location.origin);
  }
  function setText(id, text) {
    const node = byId(id);
    if (node) node.textContent = String(text || '');
    if (id === 'v78PreviewStatus' && text) publishStatus(text, /失败|错误|异常|超时/.test(String(text)) ? 'error' : /正在|处理中|解析|排队/.test(String(text)) ? 'working' : 'info');
  }
  function resetProcessLog() {
    processLogState.lines = [];
    seenProgressKeys.clear();
    setText('processResult', '');
  }
  function appendProcessLog(message) {
    const text = String(message || '').trim();
    if (!text) return;
    const stamp = new Date().toLocaleTimeString();
    processLogState.lines.push(`[${stamp}] ${text}`);
    if (processLogState.lines.length > 120) processLogState.lines.splice(0, processLogState.lines.length - 120);
    const node = byId('processResult');
    if (node) {
      node.textContent = processLogState.lines.join('\n');
      node.scrollTop = node.scrollHeight;
    }
    publishStatus(text, /失败|错误|异常|超时/.test(text) ? 'error' : /完成|成功|已创建/.test(text) ? 'success' : 'working');
  }
  function queueStateLabel(value) {
    return ({ queued: '任务已进入队列', running: '批次正在处理', waiting_retry: '处理失败，正在等待重试', done: '处理完成', failed: '处理失败', stopped: '处理已停止' })[String(value || '')] || `任务状态：${value || '处理中'}`;
  }
  function localDateText(raw) {
    if (!raw) return '-';
    const date = new Date(raw);
    return Number.isFinite(date.getTime()) ? date.toLocaleString() : String(raw);
  }
  function activateLegacyTab(tab) {
    if (typeof activateTab === 'function') activateTab(tab);
    else document.querySelector(`.tab[data-tab="${tab}"]`)?.click();
  }
  function methodLabel(method) {
    const labels = { high_imitation: '高仿文章', opening_instruction: '开头词+指令', instruction: '指令改文' };
    return labels[String(method || '')] || '自动轮换';
  }
  function configuredAiSlotMethods() {
    const rewrite = (typeof state === 'object' && state?.config?.app_config?.rewrite) || {};
    return rewrite.ai_slot_methods && typeof rewrite.ai_slot_methods === 'object' ? { ...rewrite.ai_slot_methods } : {};
  }
  function selectedTargetVersions() {
    const versions = [];
    if (byId('v78TargetOriginal')?.checked) versions.push('original');
    for (let index = 1; index <= 5; index += 1) if (byId(`v78TargetAi${index}`)?.checked) versions.push(`ai${index}`);
    return versions;
  }
  function selectedAiVersions(task = {}) {
    const explicit = asArray(task.ai_target_versions).length
      ? asArray(task.ai_target_versions)
      : asArray(task.target_versions).filter(version => /^ai[1-5]$/i.test(String(version || '')));
    if (explicit.length) return [...new Set(explicit.map(version => String(version).toLowerCase()))];
    const files = asArray(task.ai_generated_versions).length ? task.ai_generated_versions : task.ai_files;
    if (asArray(files).length) return [...new Set(files.map(version => String(version).toLowerCase()).filter(version => /^ai[1-5]$/.test(version)))];
    const count = Math.max(0, Math.min(Number(task.ai_count) || 0, 5));
    return Array.from({ length: count }, (_, index) => `ai${index + 1}`);
  }
  function selectedParsedTasks() {
    return previewState.tasks.filter(task => previewState.selected.has(String(task.bookId || task.book_id || task.id || '')));
  }

  function syncLegacyTaskSelection() {
    const ids = [...previewState.selected].map(id => String(id || '').trim()).filter(Boolean);
    window.dispatchEvent(new CustomEvent('qiantie-v78-task-selection', { detail: { ids } }));
  }
  window.qiantieV78SelectedTaskIds = () => [...previewState.selected].map(id => String(id || '').trim()).filter(Boolean);
  function currentWorkSnapshot() {
    const targets = selectedTargetVersions();
    const selected = previewState.active ? selectedParsedTasks() : [];
    const inputText = selected.length
      ? selected.map(task => String(task.sourceLine || task.source_line || '')).filter(Boolean).join('\n')
      : value('inputText');
    return {
      platform_id: value('platformSelect', '2'),
      parse_mode: value('parseModeSelect', 'smart'),
      column_preset_id: value('columnPresetSelect'),
      column_order: value('columnOrderInput', '书籍ID,书名,推荐理由,男女频,标签,评级'),
      input_text: inputText,
      max_txt: Number(value('fetchMaxTxt', 4000)) || 4000,
      text_model_id: (byId('v78RunTextModel')?.value || byId('v78ProcessingTextModel')?.value || state?.config?.app_config?.text_model_id || ''),
      target_versions: targets,
      targetVersions: targets,
      ai_slot_methods_snapshot: configuredAiSlotMethods(),
      task_ids: selected.map(task => String(task.bookId || task.book_id || task.id || '')).filter(Boolean),
      sensitive_ai_enabled: Boolean(byId('sensitiveAiProcessEnabled')?.checked),
      ...(rerunSourceBatchId ? { source_batch_id: rerunSourceBatchId } : {})
    };
  }
  function buildTaskQuery(filters = taskFilters) {
    const params = new URLSearchParams();
    if (String(filters.date || '').trim()) params.set('date', String(filters.date).trim());
    if (String(filters.bookId || '').trim()) params.set('bookId', String(filters.bookId).trim());
    if (String(filters.status || '').trim()) params.set('status', String(filters.status).trim());
    if (String(filters.scope || '').trim()) params.set('scope', String(filters.scope).trim());
    const query = params.toString();
    return query ? `?${query}` : '';
  }
  const TASK_STATUS_LABELS = {
    original_done: '原文已完成',
    original_processing: '原文处理中',
    original_failed: '原文处理失败',
    ai_done: 'AI文案已完成',
    ai_processing: 'AI文案处理中',
    ai_failed: 'AI文案生成失败',
    site_submitted: '网站已提交',
    accepted_pending: '等待网站确认',
    input_ready: '分类信息已就绪',
    queued: '排队中',
    running: '正在执行中…',
    processing: '正在执行中…',
    classifying: 'AI判断中…',
    generating: '正在生成AI文案…',
    done: '已完成',
    completed: '已完成',
    failed: '处理失败',
    error: '处理失败',
    cancelled: '已取消',
    stopped: '已停止',
    pending: '等待处理'
  };
  function taskStatusLabel(value, fallback = '') {
    const text = String(value || '').trim();
    if (!text) return fallback;
    if (TASK_STATUS_LABELS[text]) return TASK_STATUS_LABELS[text];
    return /^[a-z0-9_:-]+$/i.test(text) ? (fallback || '处理中') : text;
  }
  function originalCountLabel(task = {}) {
    const meta = task.meta && typeof task.meta === 'object' ? task.meta : {};
    const raw = Number(task.original_raw_chars ?? task.originalRawChars ?? meta.original_raw_chars ?? meta.originalRawChars ?? 0);
    const maxTxt = Number(task.max_txt ?? task.maxTxt ?? meta.max_txt ?? meta.maxTxt ?? 0);
    const actual = Number(task.original_chars ?? task.originalChars ?? meta.original_chars ?? meta.originalChars ?? 0);
    const processed = raw > 0 && maxTxt > 0 ? Math.min(maxTxt, raw) : actual;
    return raw > 0 ? `${processed}/${raw}` : (actual > 0 ? `${actual}字` : '');
  }

  function taskFilterLabel() {
    const parts = [];
    if (taskFilters.date) parts.push(taskFilters.date);
    if (taskFilters.bookId) parts.push(`ID ${taskFilters.bookId}`);
    if (taskFilters.status) parts.push(`状态 ${taskFilters.status}`);
    return parts.length ? parts.join(' · ') : '今天 + 历史未完成';
  }

  function injectStyles() {
    if (byId('v78NovelFetchV2Styles')) return;
    const style = document.createElement('style');
    style.id = 'v78NovelFetchV2Styles';
    style.textContent = `
      .v78-inline-box{margin-top:12px;padding:12px;border:1px solid var(--border,#333);border-radius:10px;background:var(--panel,#171717)}
      .v78-inline-head{display:flex;align-items:center;justify-content:space-between;gap:10px;flex-wrap:wrap}.v78-inline-head h3{margin:0;font-size:14px}
      .v78-muted{opacity:.7;font-size:12px}.v78-actions{display:flex;gap:8px;align-items:center;flex-wrap:wrap}.v78-actions.compact{margin-top:8px}
      .v78-version-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:8px;margin-top:9px}
      .v78-version-item{display:flex;align-items:center;gap:7px;padding:8px;border:1px solid var(--border,#333);border-radius:8px}.v78-version-item small{opacity:.7;margin-left:auto}
      .v78-parsed-list,.v78-batch-list{display:grid;gap:6px;margin-top:8px;max-height:260px;overflow:auto}.v78-parsed-row,.v78-batch-row{display:flex;align-items:center;gap:9px;padding:8px;border:1px solid var(--border,#333);border-radius:8px;flex-wrap:wrap}
      .v78-parsed-row .grow,.v78-batch-row .grow{flex:1;min-width:160px}.v78-ok{color:#52c41a}.v78-warn{color:#faad14}.v78-error{color:#ff4d4f}
      #v78CurrentBatchPanel{margin:12px 0 0}.v78-batch-summary{display:flex;gap:12px;flex-wrap:wrap;margin-top:8px}.v78-batch-summary span{font-size:12px;opacity:.85}
      #v78HistoryBatchesPanel[hidden]{display:none!important}.v78-history-tabs{display:flex;gap:8px;margin-bottom:10px}.v78-history-tabs button.active{font-weight:700}
      #webSubmitMount{display:none!important}
    `;
    document.head.appendChild(style);
  }

  function bindV2ProcessButton(button) {
    if (!button || button.dataset.v78V2ProcessBound === '1') return;
    button.dataset.v78V2ProcessBound = '1';
    // The legacy app registers its own onclick after this V2 client loads.
    // Capture the gesture so one click always follows preview -> V2 queue start.
    button.addEventListener('click', event => {
      event.preventDefault();
      event.stopImmediatePropagation();
      void startSelectedProcessing();
    }, true);
  }

  function mountProcessingControls() {
    if (byId('v78TargetVersions')) return;
    const input = byId('inputText');
    if (!input) return;
    const box = document.createElement('div');
    box.id = 'v78TargetVersions';
    box.className = 'v78-inline-box';
    box.innerHTML = `
      <div class="v78-inline-head"><div><h3>本次处理</h3><div class="v78-muted">选择本次需要的文案版本；点击“开始处理”后自动解析并创建当前批次。</div></div></div>
      <div class="v78-version-grid">
        <label class="v78-version-item"><input id="v78TargetOriginal" type="checkbox"/> 原文</label>
        <label class="v78-version-item"><input id="v78TargetAi1" type="checkbox" checked/> AI1 <small data-v78-method="ai1">自动轮换</small></label>
        <label class="v78-version-item"><input id="v78TargetAi2" type="checkbox"/> AI2 <small data-v78-method="ai2">自动轮换</small></label>
        <label class="v78-version-item"><input id="v78TargetAi3" type="checkbox"/> AI3 <small data-v78-method="ai3">自动轮换</small></label>
        <label class="v78-version-item"><input id="v78TargetAi4" type="checkbox"/> AI4 <small data-v78-method="ai4">自动轮换</small></label>
        <label class="v78-version-item"><input id="v78TargetAi5" type="checkbox"/> AI5 <small data-v78-method="ai5">自动轮换</small></label>
      </div>
      <div class="v78-muted" style="margin-top:8px">AI文案处理优先方案：跟随“配置”中 AI1～AI5 的现有设置。</div>
      <div id="v78ParsedBooks" class="v78-parsed-list" hidden></div>
      <div class="v78-actions compact"><button id="v78BackToInput" type="button" hidden>返回编辑</button><span id="v78PreviewStatus" class="v78-muted"></span></div>`;
    input.insertAdjacentElement('afterend', box);
    byId('v78BackToInput').onclick = backToInput;
    box.addEventListener('change', event => {
      if (!event.target.matches('#v78TargetVersions input')) return;
      if (typeof saveWorkFormState === 'function') saveWorkFormState();
      if (typeof updateVersionConfigSummary === 'function') updateVersionConfigSummary();
    });
    if (window.__batchRewritePendingWorkFormState && typeof window.batchRewriteApplyWorkFormState === 'function') {
      window.batchRewriteApplyWorkFormState(window.__batchRewritePendingWorkFormState);
      delete window.__batchRewritePendingWorkFormState;
    }
    refreshSlotMethodLabels();
    const processButton = byId('processBtn');
    bindV2ProcessButton(processButton);
  }

  async function loadProcessingTextModels(select) {
    try {
      const data = typeof api === 'function'
        ? await api('/api/models?kind=text')
        : await fetch('/api/models?kind=text', { headers: tokenHeaders() }).then(async response => {
          const body = await response.json();
          if (!response.ok) throw new Error(body.error || `HTTP ${response.status}`);
          return body;
        });
      const models = asArray(data.models);
      select.innerHTML = '';
      if (!models.length) {
        select.innerHTML = '<option value="">暂无可用文本模型</option>';
        setText('v78ProcessingTextModelStatus', '请先在 API 配置中启用文本模型');
        return;
      }
      for (const model of models) {
        const option = document.createElement('option');
        option.value = model.id || model.modelId || '';
        option.textContent = model.displayName || model.name || model.modelId || model.id;
        select.appendChild(option);
      }
      const saved = String(state?.config?.app_config?.text_model_id || state?.config?.text_model_id || '');
      select.value = models.some(model => String(model.id || model.modelId) === saved) ? saved : select.options[0].value;
      setText('v78ProcessingTextModelStatus', '执行时使用此模型；修改后将作为后续批次默认值');
    } catch (error) {
      select.innerHTML = '<option value="">读取文本模型失败</option>';
      setText('v78ProcessingTextModelStatus', `读取模型失败：${error.message || '请稍后重试'}`);
    }
  }

  function refreshSlotMethodLabels() {
    const methods = configuredAiSlotMethods();
    for (let index = 1; index <= 5; index += 1) {
      const node = document.querySelector(`[data-v78-method="ai${index}"]`);
      if (node) node.textContent = methodLabel(methods[`ai${index}`]);
    }
  }

  async function previewInput(preselectedBookIds = null) {
    const raw = String(value('inputText') || '');
    if (!raw.trim()) { setText('v78PreviewStatus', '请先填写批量输入'); return null; }
    setText('v78PreviewStatus', '正在解析...');
    try {
      const payload = currentWorkSnapshot();
      payload.input_text = raw;
      const data = await v2Api('/process/preview', { method: 'POST', body: JSON.stringify(payload) });
      previewState.active = true;
      previewState.tasks = asArray(data.tasks);
      previewState.originalInput = raw;
      const wanted = Array.isArray(preselectedBookIds) ? new Set(preselectedBookIds.map(String)) : null;
      previewState.selected = new Set(previewState.tasks
        .map(task => String(task.bookId || task.book_id || task.id || ''))
        .filter(id => id && (!wanted || wanted.has(id))));
      renderParsedBooks();
      byId('inputText').hidden = true;
      byId('v78ParsedBooks').hidden = false;
      byId('v78BackToInput').hidden = false;
      setText('v78PreviewStatus', `已解析 ${previewState.tasks.length} 本，已选 ${previewState.selected.size} 本`);
      return data;
    } catch (error) {
      setText('v78PreviewStatus', error.message);
      return null;
    }
  }

  function renderParsedBooks() {
    const box = byId('v78ParsedBooks');
    if (!box) return;
    box.innerHTML = '';
    for (const task of previewState.tasks) {
      const id = String(task.bookId || task.book_id || task.id || '');
      const row = document.createElement('label');
      row.className = 'v78-parsed-row';
      row.innerHTML = `<input type="checkbox" data-v78-preview-id="${id}" ${previewState.selected.has(id) ? 'checked' : ''}/><span class="grow"><b>${id}</b>${task.bookName ? ` · ${task.bookName}` : ''}</span><span class="v78-muted">${task.gender || ''}${task.style ? ` · ${task.style}` : ''}</span>`;
      box.appendChild(row);
    }
    if (!previewState.tasks.length) box.textContent = '没有解析到有效书籍';
    box.onchange = event => {
      const input = event.target.closest('[data-v78-preview-id]');
      if (!input) return;
      const id = String(input.dataset.v78PreviewId || '');
      if (input.checked) previewState.selected.add(id); else previewState.selected.delete(id);
      syncLegacyTaskSelection();
      setText('v78PreviewStatus', `已解析 ${previewState.tasks.length} 本，已选 ${previewState.selected.size} 本`);
    };
  }

  function backToInput() {
    previewState.active = false;
    byId('inputText').hidden = false;
    byId('v78ParsedBooks').hidden = true;
    byId('v78BackToInput').hidden = true;
    setText('v78PreviewStatus', '');
  }

  async function startSelectedProcessing() {
    syncLegacyTaskSelection();
    const button = byId('processBtn');
    if (button?.disabled) return;
    if (!selectedTargetVersions().length) { setText('v78PreviewStatus', '请至少选择一个文案版本'); return; }
    if (button) button.disabled = true;
    const processResult = byId('processResult');
    try {
      resetProcessLog();
      appendProcessLog('正在解析批量输入...');
      if (!previewState.active) {
        await previewInput();
        if (!previewState.active) {
          appendProcessLog('输入解析失败，未创建批次');
          return;
        }
      }
      appendProcessLog(`输入解析完成：${previewState.selected.size} 本小说`);
      if (!previewState.selected.size) { setText('v78PreviewStatus', '请至少填写一本有效小说'); appendProcessLog('没有解析到可处理的小说'); return; }
      const targets = selectedTargetVersions().map(item => item === 'original' ? '原文' : item.toUpperCase());
      appendProcessLog(`正在创建当前批次：${targets.join('、')}`);
      const job = await v2Api('/process/start', { method: 'POST', body: JSON.stringify(currentWorkSnapshot()) });
      if (typeof state === 'object') state.activeProcessJobId = job.id;
      appendProcessLog(`当前批次已创建${job.batch_id ? `：${job.batch_id}` : ''}，任务已排队`);
      await Promise.allSettled([loadCurrentBatch(), typeof loadTasks === 'function' ? loadTasks() : Promise.resolve()]);
      let lastQueueState = '';
      let lastRefreshAt = Date.now();
      const deadline = Date.now() + (10 * 60 * 1000);
      for (;;) {
        await new Promise(resolve => setTimeout(resolve, 800));
        const current = await v2Api(`/process/jobs/${encodeURIComponent(job.id)}`);
        for (const entry of asArray(current.progress)) {
          const key = [entry.at, entry.book_id, entry.stage, entry.status].join(':');
          if (seenProgressKeys.has(key)) continue;
          seenProgressKeys.add(key);
          appendProcessLog(`${entry.book_id || '批次'} · ${entry.stage || 'processing'} · ${taskStatusLabel(entry.status, entry.status)}${entry.message ? `：${entry.message}` : ''}`);
        }
        if (current.queue_state && current.queue_state !== lastQueueState) {
          lastQueueState = current.queue_state;
          appendProcessLog(queueStateLabel(current.queue_state));
        }
        if (Date.now() - lastRefreshAt >= 1500 || ['done', 'failed', 'cancelled'].includes(current.status)) {
          await Promise.allSettled([loadCurrentBatch(), typeof loadTasks === 'function' ? loadTasks() : Promise.resolve()]);
          lastRefreshAt = Date.now();
        }
        if (current.status === 'done') {
          const result = current.result || {};
          appendProcessLog(`处理完成：${result.unique_tasks || previewState.selected.size} 本`);
          if (typeof renderProcessResult === 'function') {
            const summary = renderProcessResult(result);
            if (summary) appendProcessLog(summary);
          }
          if (typeof renderTasks === 'function') renderTasks(result.tasks || []);
          break;
        }
        if (current.status === 'failed' || current.status === 'cancelled') {
          appendProcessLog(current.status === 'cancelled' ? '处理已取消；已完成内容已保留。' : (current.error || '处理失败'));
          break;
        }
        if (Date.now() >= deadline) { appendProcessLog('处理轮询超时，请到当前批次或任务列表查看进度'); break; }
      }
      rerunSourceBatchId = '';
      await Promise.allSettled([loadCurrentBatch(), typeof loadTasks === 'function' ? loadTasks() : Promise.resolve()]);
    } catch (error) {
      appendProcessLog(error.message);
    } finally {
      if (button) button.disabled = false;
    }
  }

  function installLegacyTaskListBridge() {
    if (legacyTaskBridgeInstalled) return true;
    if (typeof loadTasks !== 'function' || typeof renderTasks !== 'function' || typeof taskDateKey !== 'function' || typeof state !== 'object') return false;
    const legacyRenderTasks = renderTasks;
    renderTasks = function(tasks) {
      const result = legacyRenderTasks(tasks);
      patchTaskTableForV78(state.tasks || []);
      return result;
    };
    loadTasks = async function(options = {}) {
      const data = await v2Api(`/tasks${buildTaskQuery(taskFilters)}`);
      const incomingTasks = Array.isArray(data.tasks) ? data.tasks : [];
      if (options.preserveOnEmpty && incomingTasks.length === 0 && Array.isArray(state.allTasks) && state.allTasks.length > 0) {
        const fallbackTasks = Array.isArray(state.tasks) && state.tasks.length ? state.tasks : state.allTasks;
        renderTasks(fallbackTasks, { preserveVisible: true });
      } else {
        state.allTasks = typeof mergeTasksKeepingIds === 'function'
          ? mergeTasksKeepingIds(incomingTasks, options.preserveIds || [])
          : incomingTasks;
        renderTasks(state.allTasks, options);
      }
      setText('summaryText', `${taskFilterLabel()} 显示 ${state.tasks.length} 个任务`);
      return data;
    };
    legacyTaskBridgeInstalled = true;
    return true;
  }

  function patchTaskTableForV78(tasks = []) {
    const table = byId('tasksBody')?.closest('table');
    if (!table) return;
    const headerRow = table.querySelector('thead tr');
    if (!headerRow) return;
    let headers = [...headerRow.children];
    if (!headers.some(th => th.textContent.trim() === '推送日期')) {
      const parseIndex = headers.findIndex(th => th.textContent.trim() === '解析');
      const th = document.createElement('th');
      th.textContent = '推送日期';
      headerRow.insertBefore(th, headers[parseIndex + 1] || null);
    }
    headers = [...headerRow.children];
    const pushIndex = headers.findIndex(th => th.textContent.trim() === '推送日期');
    const classifyIndex = headers.findIndex(th => th.textContent.trim() === 'AI判断');
    const originalIndex = headers.findIndex(th => th.textContent.trim() === '原文');
    const aiIndex = headers.findIndex(th => th.textContent.trim() === 'AI文案');
    const taskById = new Map(asArray(tasks).map(task => [String(task.id || task.book_id || task.bookId || ''), task]));
    for (const row of byId('tasksBody').querySelectorAll('tr')) {
      const id = String(row.querySelector('.task-check')?.dataset.id || row.querySelector('[data-id]')?.dataset.id || '');
      const task = taskById.get(id);
      if (!task) continue;
      if (row.children.length === headers.length - 1) {
        const td = document.createElement('td');
        td.textContent = localDateText(task.push_date || task.batch_created_at || task.created_at || task.createdAt);
        row.insertBefore(td, row.children[pushIndex] || null);
      } else if (row.children[pushIndex]) {
        row.children[pushIndex].textContent = localDateText(task.push_date || task.batch_created_at || task.created_at || task.createdAt);
      }
      if (classifyIndex >= 0 && row.children[classifyIndex]) {
        row.children[classifyIndex].textContent = taskStatusLabel(task.classify_status ?? task.classifyStatus, task.classifier_model || task.classifierModel ? '已完成判断' : '待判断');
      }
      if (originalIndex >= 0 && row.children[originalIndex]) {
        const label = originalCountLabel(task);
        if (label) row.children[originalIndex].textContent = label;
      }
      const selected = selectedAiVersions(task);
      const generated = new Set(asArray(task.ai_generated_versions).concat(asArray(task.ai_files)).map(version => String(version).toLowerCase()));
      const aiCell = row.children[aiIndex];
      if (aiCell && selected.length) {
        const aiBusy = state.aiProcessingIds?.has(id);
        const displayTask = aiBusy ? { ...task, ai_status: 'generating', ai_current_version: task.ai_current_version || selected[0] } : task;
        const display = typeof aiCopyStatusText === 'function'
          ? aiCopyStatusText(displayTask, selected, [...generated])
          : selected.map(version => generated.has(version) ? `${version.toUpperCase()}已生成` : `${version.toUpperCase()}待生成`).join('；');
        aiCell.textContent = display;
        aiCell.title = String(task.ai_error || task.aiError || '');
      }
    }
  }

  function mountCurrentBatch() {
    if (byId('v78CurrentBatchPanel')) return;
    const work = byId('work');
    const workspace = work?.querySelector('.workspace');
    if (!work || !workspace) return;
    const panel = document.createElement('section');
    panel.id = 'v78CurrentBatchPanel';
    panel.className = 'section full';
    panel.innerHTML = `
      <div class="v78-inline-head"><div><h3>当前批次</h3><div id="v78CurrentBatchMeta" class="v78-muted">正在读取...</div></div><div class="v78-actions"><button id="v78CurrentBatchRefresh">刷新</button><button id="v78StopBatch" class="danger">停止处理</button><button id="v78ViewAllTasks">查看全部任务</button></div></div>
      <div id="v78CurrentBatchSummary" class="v78-batch-summary"></div><div id="v78CurrentBatchBooks" class="v78-batch-list"></div>`;
    workspace.insertAdjacentElement('afterend', panel);
    byId('v78CurrentBatchRefresh').onclick = () => void loadCurrentBatch();
    byId('v78StopBatch').onclick = () => void stopCurrentBatch();
    byId('v78ViewAllTasks').onclick = () => { activateLegacyTab('tasks'); if (typeof loadTasks === 'function') void loadTasks(); };
  }

  function renderCurrentBatch(batch, taskRows = []) {
    const meta = byId('v78CurrentBatchMeta');
    const summary = byId('v78CurrentBatchSummary');
    const books = byId('v78CurrentBatchBooks');
    if (!meta || !summary || !books) return;
    if (!batch) {
      meta.textContent = '暂无当前批次'; summary.innerHTML = ''; books.innerHTML = ''; return;
    }
    const settings = batch.settingsSnapshot || {};
    const targets = asArray(settings.target_versions || settings.targetVersions).map(item => item === 'original' ? '原文' : String(item).toUpperCase());
    meta.textContent = `${localDateText(batch.createdAt)} · ${taskStatusLabel(batch.status, '处理中')} · ${batch.id}`;
    summary.innerHTML = `<span>小说 ${asArray(batch.taskIds).length} 本</span><span>本次版本 ${targets.join('、') || '-'}</span><span>完成 ${batch.resultSummary?.fetched || 0}</span><span>AI文案 ${batch.resultSummary?.generated_ai_files || 0}</span>`;
    books.innerHTML = '';
    const states = new Map(asArray(batch.taskStates).map(task => [String(task.bookId || ''), task]));
    for (const task of asArray(taskRows)) states.set(String(task.id || task.book_id || task.bookId || ''), task);
    for (const id of asArray(batch.taskOrder || batch.taskIds).slice(0, 30)) {
      const task = states.get(String(id)) || { bookId: id, status: 'queued' };
      const row = document.createElement('div'); row.className = 'v78-batch-row';
      row.innerHTML = `<span class="grow"><b>${task.bookId || task.book_id || id}</b>${task.bookName || task.book_name ? ` · ${task.bookName || task.book_name}` : ''}</span><span>${taskStatusLabel(task.status, '排队中')}</span>`;
      books.appendChild(row);
    }
    if (!books.childNodes.length && asArray(batch.taskIds).length) books.textContent = `本批次共 ${batch.taskIds.length} 本，处理完成后会显示每本状态。`;
  }

  async function loadCurrentBatch() {
    try {
      const [batchData, taskData] = await Promise.all([v2Api('/batches/current'), v2Api('/tasks')]);
      renderCurrentBatch(batchData.batch || null, taskData.tasks || []);
    }
    catch (error) { setText('v78CurrentBatchMeta', error.message); }
  }

  async function refreshAllData() {
    const results = await Promise.allSettled([
      loadCurrentBatch(),
      refreshQueueSummary(),
      typeof loadTasks === 'function' ? loadTasks() : Promise.resolve()
    ]);
    const tasks = typeof state === 'object' && Array.isArray(state.tasks) ? state.tasks : [];
    const batch = byId('v78CurrentBatchMeta')?.textContent || '';
    const active = tasks.some(task => window.QiantieNovelFetchRuntime?.isActiveTask(task))
      || /(排队|处理中|执行中|等待|running|processing|queued)/i.test(batch);
    if (window.QiantieNovelFetchRuntime) {
      if (active) window.QiantieNovelFetchRuntime.startPolling(refreshAllData);
      else window.QiantieNovelFetchRuntime.stopPolling();
    }
    return results;
  }

  function installRetryProgressBridge() {
    if (retryProgressBridgeInstalled) return;
    retryProgressBridgeInstalled = true;
    window.addEventListener('qiantie-novel-fetch-retry-queued', () => {
      const runtime = window.QiantieNovelFetchRuntime;
      runtime?.invalidate('/tasks');
      runtime?.invalidate('/batches/current');
      void refreshAllData();
      runtime?.startPolling(refreshAllData);
    });
  }

  async function stopCurrentBatch() {
    try {
      await v2Api('/process/queue/stop', { method: 'POST', body: '{}' });
      setText('v78CurrentBatchMeta', '正在停止；当前正在执行的步骤会先保存结果。');
      await loadCurrentBatch();
    } catch (error) { setText('v78CurrentBatchMeta', error.message); }
  }

  function mountTaskHistory() {
    if (byId('v78HistoryBatchesPanel')) return;
    const host = document.querySelector('#tasks .section.full');
    const listDetails = byId('taskListDetails');
    if (!host || !listDetails) return;
    const tabs = document.createElement('div');
    tabs.className = 'v78-history-tabs';
    tabs.innerHTML = `<button id="v78CurrentTasksTab" class="active">当前任务</button><button id="v78HistoryTab">历史批次</button>`;
    host.insertBefore(tabs, listDetails);
    const queueSummary = document.createElement('div');
    queueSummary.id = 'v78QueueSummary';
    queueSummary.className = 'v78-inline-box v78-muted';
    queueSummary.textContent = '任务队列正在读取...';
    host.insertBefore(queueSummary, listDetails);
    const scopes = document.createElement('div');
    scopes.className = 'v78-actions compact';
    scopes.innerHTML = '<button data-v78-task-scope="current">当前批次</button><button data-v78-task-scope="unfinished">历史未完成</button><button data-v78-task-scope="all">全部任务</button>';
    scopes.onclick = event => {
      const button = event.target.closest('[data-v78-task-scope]');
      if (!button) return;
      taskFilters.scope = button.dataset.v78TaskScope || '';
      if (typeof loadTasks === 'function') void loadTasks();
    };
    host.insertBefore(scopes, listDetails);
    const history = document.createElement('div');
    history.id = 'v78HistoryBatchesPanel';
    history.className = 'v78-inline-box';
    history.hidden = true;
    history.innerHTML = `<div class="v78-inline-head"><h3>历史批次</h3><button id="v78HistoryRefresh">刷新</button></div><div id="v78HistoryBatchList" class="v78-batch-list"></div>`;
    host.insertBefore(history, listDetails);
    byId('v78CurrentTasksTab').onclick = () => showCurrentTasksView();
    byId('v78HistoryTab').onclick = () => showHistoryView();
    byId('v78HistoryRefresh').onclick = () => void loadHistoryBatches();

    const actions = host.querySelector('.batch-actions');
    if (actions && !byId('v78StopSelected')) {
      const button = document.createElement('button');
      button.id = 'v78StopSelected';
      button.className = 'danger';
      button.textContent = '停止选中';
      const retryFailed = byId('retryFailedBtn');
      if (retryFailed?.parentElement === actions) retryFailed.insertAdjacentElement('afterend', button); else actions.appendChild(button);
      button.onclick = () => void stopSelectedTasks();
    }
  }

  async function refreshQueueSummary() {
    const box = byId('v78QueueSummary');
    if (!box) return;
    try {
      const data = await v2Api('/realtime/status');
      const counts = data.counts || {};
      box.textContent = `任务中心 · 执行中 ${counts.running || 0} · 排队 ${counts.queued || 0} · 重试等待 ${counts.waiting_retry || 0} · 失败 ${counts.failed || 0} · 已停止 ${counts.stopped || 0}`;
    } catch (error) {
      box.textContent = `任务队列读取失败：${error.message || '请稍后重试'}`;
    }
  }

  function showCurrentTasksView() {
    byId('v78CurrentTasksTab')?.classList.add('active');
    byId('v78HistoryTab')?.classList.remove('active');
    byId('v78HistoryBatchesPanel').hidden = true;
    byId('taskListDetails').hidden = false;
    document.querySelector('#tasks .task-detail-section')?.removeAttribute('hidden');
  }
  function showHistoryView() {
    byId('v78HistoryTab')?.classList.add('active');
    byId('v78CurrentTasksTab')?.classList.remove('active');
    byId('v78HistoryBatchesPanel').hidden = false;
    byId('taskListDetails').hidden = true;
    document.querySelector('#tasks .task-detail-section')?.setAttribute('hidden', '');
    void loadHistoryBatches();
  }
  function v78TaskToday() {
    taskFilters.date = '';
    taskFilters.bookId = '';
    taskFilters.status = '';
    if (typeof loadTasks === 'function') void loadTasks();
  }

  async function stopSelectedTasks() {
    const ids = typeof selectedTaskIds === 'function' ? selectedTaskIds() : [...document.querySelectorAll('.task-check:checked')].map(node => node.dataset.id).filter(Boolean);
    if (!ids.length) { if (typeof setBatchStatus === 'function') setBatchStatus('先选择任务'); return; }
    try {
      const result = await v2Api('/tasks/stop-selected', { method: 'POST', body: JSON.stringify({ ids }) });
      if (typeof setBatchStatus === 'function') setBatchStatus(`已停止 ${result.updated || 0} 个选中任务`);
      if (typeof loadTasks === 'function') await loadTasks();
    } catch (error) { if (typeof setBatchStatus === 'function') setBatchStatus(error.message); }
  }

  async function loadHistoryBatches() {
    const box = byId('v78HistoryBatchList');
    if (!box) return;
    try {
      const [all, current] = await Promise.all([v2Api('/batches'), v2Api('/batches/current')]);
      const currentId = current.batch?.id || '';
      const batches = asArray(all.batches).filter(batch => batch.id !== currentId);
      box.innerHTML = '';
      for (const batch of batches) {
        const abnormal = asArray(batch.taskStates).filter(task => /(failed|error|timeout|interrupted|incomplete|partial|失败|错误|超时|中断|未完成|121异常)/i.test([task.status, task.originalStatus, task.aiStatus, task.siteSubmitStatus, task.error].filter(Boolean).join(' ')) && !/(cancelled|已取消)/i.test(String(task.status || ''))).length;
        const targets = asArray(batch.settingsSnapshot?.target_versions || batch.settingsSnapshot?.targetVersions).map(item => item === 'original' ? '原文' : String(item).toUpperCase()).join('、');
        const row = document.createElement('div'); row.className = 'v78-batch-row';
        row.innerHTML = `<span class="grow"><b>${localDateText(batch.createdAt)}</b><br><span class="v78-muted">${batch.taskIds?.length || 0} 本 · ${targets || '-'} · ${taskStatusLabel(batch.status, '处理中')}</span></span><button data-v78-rerun="all" data-batch-id="${batch.id}">全部重跑</button><button data-v78-rerun="abnormal" data-batch-id="${batch.id}" ${abnormal ? '' : 'disabled'}>重跑异常${abnormal ? ` ${abnormal}` : ''}</button>`;
        box.appendChild(row);
      }
      if (!box.childNodes.length) box.textContent = '暂无历史批次';
      box.onclick = event => {
        const button = event.target.closest('[data-v78-rerun]');
        if (button) void loadHistoricalBatchForRerun(button.dataset.batchId, button.dataset.v78Rerun);
      };
    } catch (error) { box.textContent = error.message; }
  }

  async function loadHistoricalBatchForRerun(batchId, mode) {
    try {
      const data = await v2Api(`/batches/${encodeURIComponent(batchId)}/rerun`, { method: 'POST', body: JSON.stringify({ mode }) });
      const payload = data.payload || {};
      rerunSourceBatchId = String(data.source_batch_id || batchId || '');
      activateLegacyTab('work');
      if (byId('platformSelect') && payload.platform_id != null) byId('platformSelect').value = String(payload.platform_id);
      if (byId('parseModeSelect') && payload.parse_mode) byId('parseModeSelect').value = payload.parse_mode;
      if (byId('columnPresetSelect') && payload.column_preset_id != null) byId('columnPresetSelect').value = payload.column_preset_id;
      if (byId('columnOrderInput') && payload.column_order != null) byId('columnOrderInput').value = payload.column_order;
      if (byId('inputText')) { byId('inputText').hidden = false; byId('inputText').value = String(data.input_snapshot || payload.input_text || ''); }
      const targets = new Set(asArray(payload.target_versions || payload.targetVersions).map(String));
      byId('v78TargetOriginal').checked = targets.has('original');
      for (let index = 1; index <= 5; index += 1) byId(`v78TargetAi${index}`).checked = targets.has(`ai${index}`);
      if (byId('sensitiveAiProcessEnabled') && payload.sensitive_ai_enabled != null) byId('sensitiveAiProcessEnabled').checked = payload.sensitive_ai_enabled === true;
      await previewInput(asArray(data.preselected_book_ids));
      setText('v78PreviewStatus', `${mode === 'abnormal' ? '异常小说' : '全部小说'}已载入，请确认后点击“开始处理”`);
    } catch (error) { setText('v78PreviewStatus', error.message); }
  }

  function enforceSourceAlignedUi() {
    const toggle = byId('fetchAutoDetectPlatform');
    if (toggle) {
      toggle.checked = false; toggle.disabled = true;
      const label = toggle.closest('label'); if (label) label.hidden = true;
    }
    const mount = byId('webSubmitMount'); if (mount) mount.hidden = true;
    refreshSlotMethodLabels();
  }

  function boot() {
    if (window.__qiantieNovelFetchV2Booted) return;
    window.__qiantieNovelFetchV2Booted = true;
    injectStyles();
    installLegacyTaskListBridge();
    mountProcessingControls();
    mountCurrentBatch();
    mountTaskHistory();
    enforceSourceAlignedUi();
    installRetryProgressBridge();
    void refreshAllData();
    let attempts = 0;
    const timer = window.setInterval(() => {
      installLegacyTaskListBridge();
      mountProcessingControls();
      mountCurrentBatch();
      mountTaskHistory();
      enforceSourceAlignedUi();
      attempts += 1;
      if (attempts >= 24) window.clearInterval(timer);
    }, 250);
    currentBatchTimer = null;
  }

  installLegacyTaskListBridge();
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot, { once: true });
  else boot();
  window.addEventListener('load', boot, { once: true });
})();
