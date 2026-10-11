(() => {
  const API_ROOT = '/api/batch-rewrite';
  const LAYOUT_ID = 'v78ImageAlignedStyles';
  let mounted = false;

  const $ = id => document.getElementById(id);
  const list = value => Array.isArray(value) ? value : [];
  const text = value => String(value == null ? '' : value);
  const safe = value => text(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

  function headers(extra = {}) {
    const token = localStorage.getItem('auth_token') || '';
    return { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...extra };
  }

  async function api(path) {
    if (window.QiantieNovelFetchRuntime) return window.QiantieNovelFetchRuntime.api(path);
    const response = await fetch(`${API_ROOT}${path}`, { headers: headers() });
    const body = await response.text();
    let data = {};
    try { data = body ? JSON.parse(body) : {}; } catch (_) { data = {}; }
    if (!response.ok) {
      const error = new Error(data.error || data.message || `HTTP ${response.status}`);
      error.status = response.status;
      throw error;
    }
    return data;
  }

  function installStyles() {
    if ($(LAYOUT_ID)) return;
    const style = document.createElement('style');
    style.id = LAYOUT_ID;
    style.textContent = `
      /* Reference-image aligned V78 processing layout */
      body.v78-reference-shell{--v78-blue:#1677ff;--v78-blue-strong:#0f6be6;--v78-blue-soft:rgba(22,119,255,.16);--v78-blue-line:rgba(22,119,255,.5);--v78-green:#4fc06b;--v78-red:#ff4d4f;--v78-purple:#9b6cff}
      body.v78-reference-shell .topbar{position:relative;display:grid;grid-template-columns:minmax(260px,1fr) auto minmax(260px,1fr);align-items:center;min-height:58px;padding:10px 20px;background:#0d1219;border-bottom:1px solid rgba(255,255,255,.08)}
      body.v78-reference-shell .topbar>div:first-child{display:flex;align-items:center;gap:10px;min-width:0}
      body.v78-reference-shell .topbar h1{font-size:20px;white-space:nowrap}
      body.v78-reference-shell .topbar #summaryText{display:none}
      #v78BrandBadge{display:inline-flex;align-items:center;justify-content:center;height:23px;padding:0 9px;border-radius:999px;background:var(--v78-blue);color:#fff;font-size:11px;font-weight:700;letter-spacing:.2px}
      body.v78-reference-shell .tabs{grid-column:2;display:flex;justify-content:center;gap:2px;flex-wrap:nowrap}
      body.v78-reference-shell .tab{min-height:38px;padding:8px 13px;border:0;border-bottom:2px solid transparent;border-radius:0;background:transparent;color:var(--muted)}
      body.v78-reference-shell .tab:hover{color:var(--text);background:rgba(255,255,255,.03)}
      body.v78-reference-shell .tab.active{border-bottom-color:var(--v78-blue);color:var(--v78-blue);font-weight:700;background:transparent}
      body.v78-reference-shell main{padding:14px 18px 24px}
      #work{--v78-status-ok:var(--v78-green);--v78-status-running:var(--v78-blue);--v78-status-fail:var(--v78-red);--v78-status-warn:var(--v78-purple)}
      #work .workspace#v78ImageWorkGrid{display:grid;grid-template-columns:minmax(0,1.35fr) minmax(360px,1fr);gap:14px;align-items:stretch}
      #v78ImageInputPanel,#v78ImageStatusPanel{min-height:610px;border:1px solid var(--line);border-radius:8px;padding:15px;background:linear-gradient(180deg,color-mix(in srgb,var(--panel) 98%,#111827 2%),var(--panel));box-shadow:none}
      #v78ImageInputPanel .section-title,#v78ImageStatusPanel .section-title{margin-bottom:11px}
      #v78ImageInputPanel .section-title h2,#v78ImageStatusPanel .section-title h2{font-size:17px;color:var(--v78-blue)}
      #v78ImageInputPanel .control-grid{grid-template-columns:repeat(4,minmax(120px,1fr));gap:10px}
      #v78ImageInputPanel select,#v78ImageInputPanel input,#v78ImageInputPanel textarea{border-color:color-mix(in srgb,var(--line) 78%,#718096 22%)}
      #v78ImageInputPanel .input-area{min-height:158px;max-height:225px;margin-top:12px;border-radius:7px;background:color-mix(in srgb,var(--panel) 88%,#06090e 12%)}
      #v78TargetVersions{margin-top:12px!important;padding:12px 0 0!important;border:0!important;border-top:1px solid var(--line)!important;border-radius:0!important;background:transparent!important}
      #v78TargetVersions .v78-inline-head{align-items:center}
      #v78TargetVersions .v78-inline-head h3{font-size:16px;color:var(--v78-blue)}
      #v78TargetVersions .v78-inline-head .v78-muted{display:none}
      #v78TargetVersions .v78-version-grid#v78CompactTargets{display:grid;grid-template-columns:repeat(6,minmax(86px,1fr));gap:7px;margin-top:8px}
      #v78TargetVersions .v78-version-item{min-height:36px;position:relative;display:flex;align-items:center;justify-content:flex-start;gap:6px;padding:7px 9px;border:1px solid var(--line);border-radius:6px;background:color-mix(in srgb,var(--panel-soft) 92%,#06090e 8%);color:var(--text);font-size:12px}
      #v78TargetVersions .v78-version-item:has(input:checked){border-color:var(--v78-blue-line);background:linear-gradient(180deg,rgba(22,119,255,.22),rgba(22,119,255,.13));box-shadow:inset 0 0 0 1px rgba(22,119,255,.08)}
      #v78TargetVersions .v78-version-item input{accent-color:var(--v78-blue)}
      #v78TargetVersions .v78-version-item small{display:block;margin-left:auto;font-size:10px;color:var(--muted);white-space:nowrap}
      .v78-run-meta{display:grid;grid-template-columns:minmax(150px,190px) minmax(0,1fr);gap:10px;align-items:end;margin:8px 0 3px}
      .v78-run-meta label{display:flex;gap:5px}.v78-run-hint{font-size:11px;color:var(--muted);padding-bottom:8px}
      .v78-settings-cards{display:grid;grid-template-columns:1fr 1fr;gap:9px;margin-top:10px}
      .v78-setting-card{min-height:62px;padding:10px 11px;border:1px solid var(--line);border-radius:7px;background:color-mix(in srgb,var(--panel-soft) 92%,#06090e 8%)}
      .v78-setting-card-title{font-size:12px;font-weight:700;color:var(--v78-blue);margin-bottom:7px}
      .v78-setting-card .inline-check{flex-direction:row;align-items:center;color:var(--text);gap:8px;font-size:12px}
      .v78-setting-card #webLoginStatus{margin:0;min-height:27px;display:inline-flex;align-items:center;cursor:pointer}
      #v78ImageInputPanel>.actions{display:none!important}
      #v78ImageInputPanel>#webLoginStatus,#v78ImageInputPanel>#webSubmitMount{display:none!important}
      #v78ImageStatusPanel{display:flex;flex-direction:column}
      #v78ImageStatusPanel .section-title>span{display:none}
      #v78StatusMetrics{display:grid;grid-template-columns:repeat(5,minmax(72px,1fr));gap:8px;margin-bottom:12px}
      .v78-status-metric{padding:9px 9px;border:1px solid var(--line);border-radius:7px;background:color-mix(in srgb,var(--panel-soft) 92%,#06090e 8%);min-width:0}
      .v78-status-metric span{display:block;font-size:11px;color:var(--muted);white-space:nowrap}.v78-status-metric strong{display:block;margin-top:5px;font-size:18px;color:var(--text)}
      .v78-status-metric.ok strong{color:var(--v78-status-ok)}.v78-status-metric.running strong{color:var(--v78-status-running)}.v78-status-metric.fail strong{color:var(--v78-status-fail)}
      .v78-status-divider{height:1px;background:var(--line);margin:0 0 11px}
      .v78-status-subtitle{font-size:13px;font-weight:700;color:var(--v78-blue);margin:0 0 9px}
      #v78ProgressRows{display:grid;gap:10px;margin-bottom:12px;min-height:76px}
      .v78-progress-row{display:grid;grid-template-columns:88px minmax(0,1fr) 68px;gap:10px;align-items:center;font-size:12px;color:var(--text)}
      .v78-progress-track{height:7px;border:1px solid rgba(255,255,255,.12);border-radius:999px;background:rgba(255,255,255,.035);overflow:hidden}.v78-progress-fill{height:100%;border-radius:999px;background:var(--v78-blue);transition:width .25s ease}.v78-progress-fill.ok{background:var(--v78-green)}.v78-progress-fill.web{background:var(--v78-purple)}
      .v78-log-head{display:flex;justify-content:space-between;align-items:center;margin-top:4px;margin-bottom:7px}.v78-log-head b{font-size:13px;color:var(--v78-blue)}
      #v78ImageStatusPanel #processResult{flex:1;min-height:285px;max-height:360px;margin:0;padding:12px;background:#06090e;border:1px solid rgba(255,255,255,.1);border-radius:7px;overflow:auto;white-space:pre-wrap;color:#b7c4d6}
      #v78WorkActionBar{display:flex;justify-content:space-between;align-items:center;gap:12px;margin:9px 2px 9px;flex-wrap:wrap}
      #v78WorkActionBar .v78-action-group{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
      #v78WorkActionBar button{min-height:37px;padding:8px 15px;border-color:var(--line)}
      #v78WorkActionBar button.primary,#v78WorkActionBar #processBtn.primary{background:var(--v78-blue);border-color:var(--v78-blue);color:#fff}
      #v78WorkActionBar button.primary:hover,#v78WorkActionBar #processBtn.primary:hover{background:var(--v78-blue-strong);border-color:var(--v78-blue-strong)}
      #v78WorkActionBar #processBtn{min-width:150px}
      #v78CurrentBatchPanel{margin-top:0!important;padding:0!important;overflow:hidden;background:var(--panel);border:1px solid var(--line);border-radius:8px}
      #v78CurrentBatchPanel>.v78-inline-head{padding:10px 13px;border-bottom:1px solid var(--line)}
      #v78CurrentBatchPanel>.v78-inline-head h3{color:var(--text);font-size:13px}
      #v78CurrentBatchPanel #v78CurrentBatchMeta{display:none!important}
      #v78CurrentBatchPanel #v78CurrentBatchSummary,#v78CurrentBatchPanel #v78CurrentBatchBooks{display:none!important}
      #v78CurrentBatchTable{width:100%;overflow:auto}
      #v78CurrentBatchTable table{width:100%;border-collapse:collapse;min-width:960px;font-size:12px}
      #v78CurrentBatchTable th,#v78CurrentBatchTable td{padding:9px 10px;border-right:1px solid var(--line);border-bottom:1px solid var(--line);text-align:center;white-space:nowrap}
      #v78CurrentBatchTable th{background:color-mix(in srgb,var(--panel-soft) 92%,#06090e 8%);font-weight:700;color:#c7d2e0}#v78CurrentBatchTable tr:last-child td{border-bottom:0}#v78CurrentBatchTable th:last-child,#v78CurrentBatchTable td:last-child{border-right:0}
      #v78CurrentBatchTable tbody tr:hover{background:rgba(22,119,255,.045)}
      .v78-cell-ok{color:var(--v78-green);font-weight:700}.v78-cell-run{color:var(--v78-blue)}.v78-cell-fail{color:var(--v78-red);font-weight:700}.v78-cell-muted{color:var(--muted)}
      .v78-table-view{padding:4px 8px!important;min-height:auto!important;color:var(--v78-blue)}
      .v78-preview-login-note{padding:14px 16px;color:var(--muted);font-size:12px;text-align:left}
      @media(max-width:1120px){body.v78-reference-shell .topbar{grid-template-columns:1fr auto}body.v78-reference-shell .tabs{grid-column:2}#work .workspace#v78ImageWorkGrid{grid-template-columns:1fr}#v78ImageInputPanel,#v78ImageStatusPanel{min-height:auto}.v78-settings-cards{grid-template-columns:1fr}#v78TargetVersions .v78-version-grid#v78CompactTargets{grid-template-columns:repeat(3,minmax(100px,1fr))}}
      @media(max-width:720px){body.v78-reference-shell .topbar{display:flex;flex-wrap:wrap}body.v78-reference-shell .tabs{width:100%;justify-content:flex-start;overflow:auto}#v78ImageInputPanel .control-grid{grid-template-columns:1fr 1fr}#v78StatusMetrics{grid-template-columns:repeat(2,1fr)}.v78-run-meta{grid-template-columns:1fr}#v78TargetVersions .v78-version-grid#v78CompactTargets{grid-template-columns:repeat(2,1fr)}}
    `;
    document.head.appendChild(style);
  }

  function ensureReferenceTopbar() {
    document.body.classList.add('v78-reference-shell');
    const topbar = document.querySelector('.topbar');
    if (!topbar) return false;
    const title = topbar.querySelector('h1');
    if (title && !$('v78BrandBadge')) {
      const badge = document.createElement('span');
      badge.id = 'v78BrandBadge';
      badge.textContent = 'V78';
      title.insertAdjacentElement('afterend', badge);
    }
    return true;
  }

  function syncRunMaxTxt() {
    const proxy = $('v78RunMaxTxt');
    const source = $('fetchMaxTxt');
    if (!proxy || !source) return;
    if (!proxy.dataset.bound) {
      proxy.value = source.value || '4000';
      proxy.addEventListener('input', () => { source.value = proxy.value; source.dispatchEvent(new Event('input', { bubbles: true })); });
      source.addEventListener('input', () => { if (document.activeElement !== proxy) proxy.value = source.value; });
      proxy.dataset.bound = '1';
    }
  }

  function moveRunSettings() {
    const target = $('v78TargetVersions');
    if (!target) return false;
    const heading = target.querySelector('.v78-inline-head h3');
    if (heading) heading.textContent = '本次处理设置';
    const versionGrid = target.querySelector('.v78-version-grid');
    if (versionGrid) versionGrid.id = 'v78CompactTargets';

    if (!$('v78RunMeta')) {
      const meta = document.createElement('div');
      meta.id = 'v78RunMeta';
      meta.className = 'v78-run-meta';
      meta.innerHTML = `<label>截取字数<input id="v78RunMaxTxt" type="number" min="1" value="4000" /></label><div class="v78-run-hint">本次版本只选择一次；未选择的 AI 不生成、不显示、不提交网络。</div>`;
      target.insertBefore(meta, versionGrid || target.firstChild);
    }
    syncRunMaxTxt();

    if (!$('v78SettingsCards')) {
      const cards = document.createElement('div');
      cards.id = 'v78SettingsCards';
      cards.className = 'v78-settings-cards';
      cards.innerHTML = `<div id="v78SensitiveCard" class="v78-setting-card"><div class="v78-setting-card-title">敏感处理</div></div><div id="v78LoginCard" class="v78-setting-card"><div class="v78-setting-card-title">121 登录</div></div>`;
      const parsed = $('v78ParsedBooks');
      target.insertBefore(cards, parsed || null);
    }
    const sensitive = $('sensitiveAiProcessEnabled')?.closest('label');
    if (sensitive && sensitive.parentElement?.id !== 'v78SensitiveCard') $('v78SensitiveCard')?.appendChild(sensitive);
    const login = $('webLoginStatus');
    if (login && login.parentElement?.id !== 'v78LoginCard') $('v78LoginCard')?.appendChild(login);
    return true;
  }

  function progressRow(label, data, cls = '') {
    const pct = data.total > 0 ? Math.max(0, Math.min(100, Math.round(data.done * 100 / data.total))) : 0;
    return `<div class="v78-progress-row"><span>${label}</span><div class="v78-progress-track"><div class="v78-progress-fill ${cls}" style="width:${pct}%"></div></div><span>${data.done} / ${data.total}</span></div>`;
  }

  function zeroProgressHtml() {
    return [
      progressRow('原文获取', { done: 0, total: 0 }, 'ok'),
      progressRow('AI文案生成', { done: 0, total: 0 }, ''),
      progressRow('121网站提交', { done: 0, total: 0 }, 'web')
    ].join('');
  }

  function ensureStatusPanel() {
    const statusPanel = document.querySelector('#work .workspace .process-status-section');
    if (!statusPanel) return false;
    statusPanel.id = 'v78ImageStatusPanel';
    if (!$('v78StatusMetrics')) {
      const metrics = document.createElement('div');
      metrics.id = 'v78StatusMetrics';
      metrics.innerHTML = [
        ['v78MetricTotal', '总数量', ''],
        ['v78MetricParsed', '已解析', 'ok'],
        ['v78MetricRunning', '处理中', 'running'],
        ['v78MetricDone', '已完成', 'ok'],
        ['v78MetricFailed', '失败', 'fail']
      ].map(([id, label, cls]) => `<div class="v78-status-metric ${cls}"><span>${label}</span><strong id="${id}">0 本</strong></div>`).join('');
      const title = statusPanel.querySelector('.section-title');
      title?.insertAdjacentElement('afterend', metrics);
    }
    if (!$('v78ProgressRows')) {
      const wrap = document.createElement('div');
      wrap.innerHTML = `<div class="v78-status-divider"></div><div class="v78-status-subtitle">实时进度</div><div id="v78ProgressRows">${zeroProgressHtml()}</div><div class="v78-log-head"><b>实时日志</b></div>`;
      const result = $('processResult');
      statusPanel.insertBefore(wrap, result || null);
    }
    return true;
  }

  function makeShortcut(id, label, targetId, primary = false) {
    if ($(id)) return $(id);
    const button = document.createElement('button');
    button.id = id;
    button.type = 'button';
    button.textContent = label;
    if (primary) button.className = 'primary';
    button.addEventListener('click', () => {
      const target = $(targetId);
      const tasksTab = document.querySelector('.tab[data-tab="tasks"]');
      tasksTab?.click();
      // Submit directly through the legacy runtime. The old target can be
      // replaced while switching task panels, so a synthetic click may do nothing.
      if (targetId === 'openWebSubmitBtn') window.qiantieSubmitSelectedTasks?.(window.qiantieV78SelectedTaskIds?.());
      window.setTimeout(() => {
        target?.scrollIntoView({ behavior: 'smooth', block: 'center' });
        target?.focus?.();
      }, 80);
    });
    return button;
  }

  function ensureActionBar(workspace) {
    if ($('v78WorkActionBar')) return;
    const bar = document.createElement('div');
    bar.id = 'v78WorkActionBar';
    bar.innerHTML = `<div id="v78WorkPrimaryActions" class="v78-action-group"></div><div id="v78WorkTaskActions" class="v78-action-group"></div>`;
    workspace.insertAdjacentElement('afterend', bar);
    const primary = $('v78WorkPrimaryActions');
    const processBtn = $('processBtn');
    const refreshBtn = $('refreshBtn');
    if (processBtn) primary.appendChild(processBtn);
    if (refreshBtn) primary.appendChild(refreshBtn);
    const secondary = $('v78WorkTaskActions');
    secondary.appendChild(makeShortcut('v78QuickRetry', '重试失败', 'retryFailedBtn'));
    secondary.appendChild(makeShortcut('v78QuickRules', '规则处理', 'applyRulesSelectedBtn'));
    secondary.appendChild(makeShortcut('v78QuickSubmit', '提交网络', 'openWebSubmitBtn'));
    secondary.appendChild(makeShortcut('v78QuickFactory', '进入批量工厂', 'transferBatchFactoryBtn', true));
  }

  function mountImageAlignedLayout() {
    installStyles();
    ensureReferenceTopbar();
    const workspace = document.querySelector('#work .workspace');
    if (!workspace) return false;
    workspace.id = 'v78ImageWorkGrid';
    const sections = workspace.querySelectorAll(':scope > .section');
    if (sections[0]) sections[0].id = 'v78ImageInputPanel';
    if (sections[1]) sections[1].id = 'v78ImageStatusPanel';
    const title = sections[0]?.querySelector('.section-title h2');
    if (title) title.textContent = '批量输入';
    moveRunSettings();
    ensureStatusPanel();
    ensureActionBar(workspace);
    ensureCurrentBatchTable();
    mounted = Boolean($('v78TargetVersions') && $('v78CurrentBatchPanel'));
    return mounted;
  }

  function statusClass(status) {
    const value = text(status).toLowerCase();
    if (/done|complete|submitted|confirmed|已完成|成功/.test(value)) return 'v78-cell-ok';
    if (/fail|error|timeout|partial|失败|错误|超时|异常/.test(value)) return 'v78-cell-fail';
    if (/run|process|queue|pending|处理中|等待|排队/.test(value)) return 'v78-cell-run';
    return 'v78-cell-muted';
  }

  const STATUS_LABELS = {
    original_done: '原文已完成',
    original_processing: '原文处理中',
    original_failed: '原文处理失败',
    ai_done: 'AI文案已完成',
    ai_processing: 'AI文案处理中',
    ai_failed: 'AI文案生成失败',
    pending_upload: '待上传',
    uploading: '上传中',
    site_submitted: '网站已提交',
    accepted_pending: '等待网站确认',
    queued: '排队中',
    running: '正在执行中…',
    processing: '处理中…',
    pending: '等待处理',
    waiting_retry: '等待重试',
    done: '已完成',
    completed: '已完成',
    failed: '处理失败',
    error: '处理失败',
    cancelled: '已取消',
    stopped: '已停止'
  };

  function statusLabel(value, fallback = '处理中') {
    const raw = text(value).trim();
    if (!raw) return fallback;
    return STATUS_LABELS[raw.toLowerCase()] || (/^[a-z0-9_:-]+$/i.test(raw) ? fallback : raw);
  }

  function statusMark(status, ready = false) {
    const value = text(status);
    if (ready || /done|complete|submitted|confirmed|已完成|成功/i.test(value)) return '<span class="v78-cell-ok">✓</span>';
    if (/fail|error|timeout|partial|失败|错误|超时|异常/i.test(value)) return '<span class="v78-cell-fail">×</span>';
    if (/run|process|queue|pending|处理中|等待|排队/i.test(value)) return '<span class="v78-cell-run">○</span>';
    return '<span class="v78-cell-muted">-</span>';
  }

  function taskId(task) { return text(task?.book_id || task?.bookId || task?.id); }
  function targetVersions(batch) {
    const settings = batch?.settingsSnapshot || {};
    return list(settings.target_versions || settings.targetVersions).map(value => text(value).toLowerCase()).filter(Boolean);
  }
  function generatedVersions(task) {
    const values = [
      ...list(task?.ai_generated_versions),
      ...list(task?.ai_files),
      ...list(task?.generated_versions)
    ];
    return new Set(values.map(value => typeof value === 'string' ? value.toLowerCase() : text(value?.version || value?.name).toLowerCase()).filter(Boolean));
  }
  function confirmedVersions(task) {
    return new Set([
      ...list(task?.site_submit_confirmed_versions),
      ...list(task?.siteSubmitConfirmedVersions),
      ...list(task?.confirmed_versions)
    ].map(value => text(value).toLowerCase()).filter(Boolean));
  }

  function emptyBatchTable(message) {
    return `<table><thead><tr><th>ID</th><th>书名</th><th>Book ID</th><th>原文</th><th>AI文案</th><th>网站提交</th><th>状态</th><th>字数</th><th>创建时间</th><th>操作</th></tr></thead><tbody><tr><td colspan="10" class="v78-preview-login-note">${safe(message)}</td></tr></tbody></table>`;
  }

  function ensureCurrentBatchTable() {
    const panel = $('v78CurrentBatchPanel');
    if (!panel) return false;
    if (!$('v78CurrentBatchTable')) {
      const wrap = document.createElement('div');
      wrap.id = 'v78CurrentBatchTable';
      wrap.innerHTML = emptyBatchTable('正在读取当前批次...');
      panel.appendChild(wrap);
    }
    const actionBar = $('v78WorkActionBar');
    if (actionBar && panel.previousElementSibling !== actionBar) actionBar.insertAdjacentElement('afterend', panel);
    return true;
  }

  function dateLabel(value) {
    if (!value) return '-';
    const date = new Date(value);
    return Number.isFinite(date.getTime()) ? date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }) : text(value);
  }

  function wordLabel(task, versions) {
    const raw = task?.word_count || task?.wordCount || task?.original_chars || task?.originalChars || task?.chars;
    if (raw != null && raw !== '') return safe(raw);
    const pieces = [];
    if (versions.includes('original') && task?.original_chars != null) pieces.push(task.original_chars);
    for (const version of versions.filter(v => /^ai[1-5]$/.test(v))) {
      const value = task?.[`${version}_chars`] ?? task?.[`${version}Chars`];
      if (value != null) pieces.push(value);
    }
    return pieces.length ? pieces.map(safe).join('/') : '-';
  }

  function renderCurrentBatchTable(batch, tasks) {
    const wrap = $('v78CurrentBatchTable');
    if (!wrap) return;
    if (!batch) {
      wrap.innerHTML = emptyBatchTable('暂无当前批次');
      return;
    }
    const targets = targetVersions(batch);
    const showOriginal = targets.includes('original');
    const aiTargets = targets.filter(version => /^ai[1-5]$/.test(version));
    const map = new Map(list(tasks).map(task => [taskId(task), task]));
    const states = new Map(list(batch.taskStates).map(task => [taskId(task), task]));
    const ids = list(batch.taskIds).length ? list(batch.taskIds) : [...states.keys()];
    const versionHeads = `${showOriginal ? '<th>原文</th>' : ''}${aiTargets.map(version => `<th>${version.toUpperCase()}</th>`).join('')}`;
    const rows = ids.map((id, index) => {
      const live = map.get(text(id)) || {};
      const snap = states.get(text(id)) || {};
      const task = { ...snap, ...live };
      const generated = generatedVersions(task);
      const confirmed = confirmedVersions(task);
      const originalStatus = task.original_status || task.originalStatus || '';
      const aiStatus = task.ai_status || task.aiStatus || '';
      const siteStatus = task.site_submit_status || task.siteSubmitStatus || '';
      const status = task.status || snap.status || '';
      const versionCells = `${showOriginal ? `<td>${statusMark(originalStatus, /done|complete|成功|已完成/i.test(text(originalStatus)))}</td>` : ''}${aiTargets.map(version => `<td>${statusMark(aiStatus, generated.has(version))}</td>`).join('')}`;
      const targetReadyCount = targets.filter(version => version === 'original' ? /done|complete|成功|已完成/i.test(text(originalStatus)) : generated.has(version)).length;
      const submitReady = targets.length > 0 && (confirmed.size ? targets.every(version => confirmed.has(version)) : /submitted|confirmed|done|success|成功|已提交/i.test(text(siteStatus)));
      const submitCell = statusMark(siteStatus, submitReady);
      const displayStatus = statusLabel(status || siteStatus || aiStatus || originalStatus, targetReadyCount === targets.length && targets.length ? '已完成' : '处理中');
      return `<tr data-v78-current-id="${safe(id)}"><td>${index + 1}</td><td>${safe(task.book_name || task.bookName || '-')}</td><td>${safe(id)}</td>${versionCells}<td>${submitCell}</td><td class="${statusClass(status || siteStatus || aiStatus || originalStatus)}">${safe(displayStatus)}</td><td>${wordLabel(task, targets)}</td><td>${dateLabel(task.created_at || task.createdAt || batch.createdAt)}</td><td><button type="button" class="v78-table-view" data-v78-open-task="${safe(id)}">查看</button></td></tr>`;
    }).join('');
    wrap.innerHTML = `<table><thead><tr><th>ID</th><th>书名</th><th>Book ID</th>${versionHeads}<th>网站提交</th><th>状态</th><th>字数</th><th>创建时间</th><th>操作</th></tr></thead><tbody>${rows || `<tr><td colspan="${8 + aiTargets.length + (showOriginal ? 1 : 0)}" class="v78-cell-muted">本批次任务尚未生成</td></tr>`}</tbody></table>`;
    wrap.onclick = event => {
      const button = event.target.closest('[data-v78-open-task]');
      if (!button) return;
      openTask(button.dataset.v78OpenTask || '');
    };
  }

  function openTask(id) {
    document.querySelector('.tab[data-tab="tasks"]')?.click();
    window.setTimeout(() => {
      const checker = [...document.querySelectorAll('#tasksBody .task-check')].find(node => text(node.dataset.id) === text(id));
      const row = checker?.closest('tr');
      const view = [...(row?.querySelectorAll('button') || [])].find(button => /查看/.test(button.textContent || ''));
      if (view) view.click();
      else row?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }, 100);
  }

  function taskCounters(batch) {
    const states = list(batch?.taskStates);
    const total = list(batch?.taskIds).length || states.length;
    let running = 0;
    let done = 0;
    let failed = 0;
    for (const task of states) {
      const joined = [task.status, task.originalStatus, task.aiStatus, task.siteSubmitStatus, task.error].filter(Boolean).join(' ');
      if (/fail|error|timeout|partial|失败|错误|超时|异常/i.test(joined)) failed += 1;
      else if (/done|complete|已完成|成功/i.test(text(task.status))) done += 1;
      else if (/run|process|queue|pending|处理中|等待|排队/i.test(joined)) running += 1;
    }
    if (!states.length && batch?.status === 'running') running = total;
    return { total, parsed: total, running, done, failed };
  }

  function progressCounts(batch, tasks) {
    const targets = targetVersions(batch);
    const ids = list(batch?.taskIds);
    const map = new Map(list(tasks).map(task => [taskId(task), task]));
    const states = new Map(list(batch?.taskStates).map(task => [taskId(task), task]));
    const totalBooks = ids.length;
    let originalDone = 0;
    let aiDone = 0;
    let submitDone = 0;
    const aiTargets = targets.filter(version => /^ai[1-5]$/.test(version));
    const uploadTargets = targets.length;
    for (const id of ids) {
      const task = { ...(states.get(text(id)) || {}), ...(map.get(text(id)) || {}) };
      const originalStatus = task.original_status || task.originalStatus || '';
      if (/done|complete|成功|已完成/i.test(text(originalStatus))) originalDone += 1;
      const generated = generatedVersions(task);
      aiDone += aiTargets.filter(version => generated.has(version)).length;
      const confirmed = confirmedVersions(task);
      if (confirmed.size) submitDone += targets.filter(version => confirmed.has(version)).length;
      else if (/submitted|confirmed|done|success|成功|已提交/i.test(text(task.site_submit_status || task.siteSubmitStatus))) submitDone += uploadTargets;
    }
    return {
      original: { done: originalDone, total: totalBooks },
      ai: { done: aiDone, total: totalBooks * aiTargets.length },
      submit: { done: submitDone, total: totalBooks * uploadTargets }
    };
  }

  function updateStatus(batch, tasks) {
    const counters = taskCounters(batch);
    const map = {
      v78MetricTotal: counters.total,
      v78MetricParsed: counters.parsed,
      v78MetricRunning: counters.running,
      v78MetricDone: counters.done,
      v78MetricFailed: counters.failed
    };
    for (const [id, value] of Object.entries(map)) if ($(id)) $(id).textContent = `${value} 本`;
    const rows = $('v78ProgressRows');
    if (rows) {
      const progress = progressCounts(batch, tasks);
      rows.innerHTML = [
        progressRow('原文获取', progress.original, 'ok'),
        progressRow('AI文案生成', progress.ai, ''),
        progressRow('121网站提交', progress.submit, 'web')
      ].join('');
    }
  }

  function friendlyAuthMessage(error) {
    const value = text(error?.message);
    if (error?.status === 401 || /unauthorized|expired token|invalid token/i.test(value)) {
      return '登录已过期，请重新登录 V78 后刷新';
    }
    return value || '当前批次读取失败';
  }

  async function refreshImageLayoutData() {
    if (!mounted) mountImageAlignedLayout();
    if (!$('v78CurrentBatchTable') || !$('v78StatusMetrics')) return;
    try {
      const [batchData, taskData] = await Promise.all([api('/batches/current'), api('/tasks')]);
      const batch = batchData.batch || null;
      const tasks = list(taskData.tasks);
      renderCurrentBatchTable(batch, tasks);
      updateStatus(batch, tasks);
    } catch (error) {
      const wrap = $('v78CurrentBatchTable');
      const message = friendlyAuthMessage(error);
      if (wrap) wrap.innerHTML = emptyBatchTable(message === '登录已过期，请重新登录 V78 后刷新' ? '请登录后查看当前批次' : message);
      const rows = $('v78ProgressRows');
      if (rows && !rows.children.length) rows.innerHTML = zeroProgressHtml();
    }
  }

  function boot() {
    if (window.__qiantieNovelFetchLayoutBooted) return;
    window.__qiantieNovelFetchLayoutBooted = true;
    let attempts = 0;
    const waiter = window.setInterval(() => {
      mountImageAlignedLayout();
      attempts += 1;
      if (mounted || attempts >= 40) {
        window.clearInterval(waiter);
        void refreshImageLayoutData();
      }
    }, 200);
  }

  window.V78NovelFetchLayout = { mountImageAlignedLayout, refreshImageLayoutData };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot, { once: true });
  else boot();
})();
