const labels = {
  scheduled_waiting: '等待定时执行',
  input_ready: '分类信息已就绪',
  queued: '排队中',
  processing: '正在执行中…',
  waiting_retry: '等待重试',
  waiting_classification: '等待确认分类',
  waiting_sensitive_ai: '等待敏感词 AI',
  pending_submit: '处理完成，待提交',
  upload_pending: '正文已完成，待重试上传',
  waiting_configuration: '等待配置',
  waiting_login: '等待登录',
  completed: '完成',
  failed: '失败',
  stopped: '已停止'
};

const colors = {
  scheduled_waiting: 'blue',
  input_ready: 'blue',
  queued: 'gold',
  processing: 'processing',
  waiting_retry: 'orange',
  waiting_classification: 'orange',
  waiting_sensitive_ai: 'orange',
  pending_submit: 'cyan',
  upload_pending: 'orange',
  waiting_configuration: 'orange',
  waiting_login: 'orange',
  completed: 'green',
  failed: 'red',
  stopped: 'default'
};

const text = value => String(value ?? '').trim().toLowerCase();
const done = value => /^(done|success|succeeded|completed|complete|ready|ai_done|submitted|uploaded)$/.test(text(value));

export function novelFetchStageLabel(value) {
  const normalized = text(value);
  return {
    done: '已完成', success: '已完成', succeeded: '已完成', completed: '已完成', complete: '已完成',
    ready: '已就绪', ai_done: '已生成', generating: '生成中…', ai_processing: '生成中…',
    queued: '排队中', processing: '正在执行中…', running: '正在执行中…', waiting_retry: '等待重试',
    waiting_classification: '等待确认分类',
    waiting_sensitive_ai: '等待敏感词 AI',
    pending_upload: '待提交', pending_submit: '待提交', submitted: '已提交', uploaded: '已上传',
    waiting_configuration: '等待配置', waiting_login: '等待登录',
    failed: '失败', error: '失败', stopped: '已停止'
  }[normalized] || String(value || '未开始');
}

export function novelFetchOriginalFetchRetry(row = {}) {
  return {
    attempts: Number(row.originalFetchRetryAttempts ?? row.original_fetch_retry_count ?? row.original_fetch_retry_attempts) || 0,
    nextRetryAt: String(row.originalFetchNextRetryAt ?? row.original_fetch_next_retry_at ?? '').trim(),
    error: String(row.originalFetchLastError ?? row.original_fetch_last_error ?? '').trim()
  };
}

export function novelFetchBatchResultSummary(result = {}) {
  if (Object.hasOwn(result, 'queued')) {
    return `已加入后台重试 ${Number(result.queued) || 0} 项，已有任务 ${Number(result.already_running) || 0} 项，跳过 ${Number(result.skipped) || 0} 项，失败 ${Number(result.failed) || 0} 项；执行结果请看任务状态`;
  }
  const requested = Number(result.requested) || 0;
  const updated = Number(result.updated ?? result.retried ?? result.processed ?? result.deleted) || 0;
  const skipped = Number(result.skipped) || 0;
  const failed = Number(result.failed) || 0;
  return `请求 ${requested} 项，已处理 ${updated} 项，跳过 ${skipped} 项，失败 ${failed} 项`;
}

const submitStatusLabels = {
  submitted: '已提交',
  success: '已提交',
  done: '已提交',
  uploaded: '已提交',
  pending_confirmation: '待远端核验',
  pending_verification: '待远端核验',
  unknown: '结果未知',
  partial: '部分接收',
  pending_upload: '待提交',
  pending: '待提交',
  pending_submit: '待提交',
  queued: '待提交',
  waiting_configuration: '等待配置',
  waiting_config: '等待配置',
  waiting_login: '等待登录',
  waiting_classification: '等待确认分类',
  waiting_classify: '等待确认分类',
  waiting_generate: '待生成',
  pending_generate: '待生成',
  executing: '执行中',
  processing: '执行中',
  running: '执行中',
  deferred: '已延后',
  skipped: '已跳过',
  failed: '失败',
  error: '失败',
  stopped: '已停止'
};

// 提交状态只保留一套展示词汇：后端历史字段（pending_upload/error 等）
// 和新字段（waiting_config/unknown 等）都从这里翻译，页面不再直接显示英文状态。
export function webSubmitStatusLabel(status) {
  const key = text(status);
  return submitStatusLabels[key] || String(status ?? '').trim() || '未开始';
}

export function novelFetchSubmissionLabel(row = {}) {
  const submitStatus = text(row.siteSubmitStatus ?? row.site_submit_status);
  const versions = key => (Array.isArray(row[key]) ? row[key] : []).map(value => String(value || '').trim()).filter(Boolean).join('、');
  if (submitStatus === 'unknown') {
    const pending = versions('site_submit_pending_versions') || versions('site_submit_missing_versions');
    return `待远端核验${pending ? ` · 待提交 ${pending.toUpperCase()}` : ''}`;
  }
  if (submitStatus === 'partial') {
    const failed = versions('site_submit_failed_versions');
    return `部分接收${failed ? ` · 失败 ${failed.toUpperCase()}` : ''}`;
  }
  if (submitStatus === 'waiting_generate' || submitStatus === 'pending_generate') {
    const waiting = versions('site_submit_waiting_generate_versions');
    return `待生成${waiting ? ` ${waiting.toUpperCase()}` : ''}`;
  }
  if (submitStatus) return webSubmitStatusLabel(submitStatus);
  const aiStatus = row.aiStatus ?? row.ai_status;
  const aiCount = Number(row.aiGeneratedCount ?? row.ai_generated_count) || 0;
  return aiCount > 0 || done(aiStatus) ? '待提交' : '—';
}

export function novelFetchVersionReady(row = {}, version = 'ai1') {
  const versionID = String(version || '').trim().toLowerCase();
  if (!versionID) return false;
  const versions = row.versions && typeof row.versions === 'object' ? row.versions : {};
  const value = versions[versionID];
  if (typeof value === 'string' && value.trim()) return true;
  if (value && typeof value === 'object') return text(value.state) === 'ready';
  if (value === false) return false;
  const files = row.aiFiles ?? row.ai_files;
  if (Array.isArray(files)) return files.some(file => String(file).toLowerCase() === versionID);
  if (versionID !== 'ai1') return false;
  return done(row.aiStatus ?? row.ai_status)
    || Number(row.aiGeneratedCount ?? row.ai_generated_count) > 0;
}

export function beijingDateKey(date = new Date()) {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit'
  }).formatToParts(date);
  const values = Object.fromEntries(parts.map(part => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}`;
}

export function novelFetchTaskDateKey(row = {}) {
  const raw = String(row.createdAt ?? row.created_at ?? row.updatedAt ?? row.updated_at ?? '').trim();
  if (!raw) return '';
  const date = new Date(raw);
  return Number.isNaN(date.getTime()) ? raw.slice(0, 10) : beijingDateKey(date);
}

export function filterNovelFetchTasks(rows = [], options = {}) {
  const source = Array.isArray(rows) ? rows : [];
  const query = String(options.query || '').trim().toLowerCase();
  const mode = options.mode === 'date' ? 'date' : 'default';
  const selectedDate = String(options.date || '').trim();
  // 参考实现的默认视图返回当前身份的全部可见任务，只有显式切到日期视图
  // 才按日期过滤；这里保持同一语义，不再用「近 3 天」截断历史任务。
  return source.filter(row => {
    const identity = `${row.bookId ?? row.book_id ?? row.id ?? ''} ${row.bookName ?? row.book_name ?? row.title ?? ''}`.toLowerCase();
    if (query && !identity.includes(query)) return false;
    if (mode === 'date') return novelFetchTaskDateKey(row) === selectedDate;
    return true;
  });
}

export function retainVisibleTaskIds(selectedIds = [], visibleRows = []) {
  const visible = new Set(visibleRows
    .map(row => String(row.bookId ?? row.book_id ?? row.id ?? '').trim())
    .filter(Boolean));
  return selectedIds.filter(id => visible.has(String(id ?? '').trim()));
}

export function currentBatchBookIds(batch = {}) {
  const direct = batch.taskIds ?? batch.task_ids;
  const rows = batch.taskStates ?? batch.task_states;
  const source = Array.isArray(direct) && direct.length
    ? direct
    : (Array.isArray(rows) ? rows.map(row => row.bookId ?? row.book_id ?? row.id) : []);
  return [...new Set(source.map(value => String(value ?? '').trim()).filter(Boolean))];
}

export function stoppableCurrentBatchBookIds(batch = {}) {
  const rows = batch.taskStates ?? batch.task_states;
  if (!Array.isArray(rows)) return [];
  return rows
    .filter(row => ['queued', 'processing', 'waiting_retry'].includes(novelFetchTaskStatus(row).status))
    .map(row => String(row.bookId ?? row.book_id ?? row.id ?? '').trim())
    .filter(Boolean);
}

export function shouldPollNovelFetchRuntime({ activeTab, processing, currentBatchCount } = {}) {
  return activeTab === 'process'
    || activeTab === 'tasks'
    || Boolean(processing)
    || Number(currentBatchCount) > 0;
}

// Keep lightweight runtime feedback responsive without repeatedly fetching the
// full task table. Values mirror the reference workflow's polling contract.
export function novelFetchPollingPlan({ activeTab, processing, documentHidden, taskFailures = 0, realtimeFailures = 0 } = {}) {
  const active = activeTab === 'process' || activeTab === 'tasks' || Boolean(processing);
  if (!active || documentHidden) return { realtimeMs: 0, queueMs: 0, tasksMs: 0 };
  const taskDelay = Number(taskFailures) >= 2 ? 30000 : Number(taskFailures) > 0 ? 20000 : 10000;
  const realtimeDelay = Number(realtimeFailures) >= 2 ? 15000 : Number(realtimeFailures) > 0 ? 5000 : 2000;
  return { realtimeMs: realtimeDelay, queueMs: 5000, tasksMs: taskDelay };
}

export function novelFetchCurrentBatchTaskDetail(row = {}) {
  const bookId = String(row.bookId ?? row.book_id ?? row.id ?? '').trim();
  return bookId ? { bookId, initialTab: 'summary' } : null;
}

export function novelFetchQueueSummary(status = {}) {
  const counts = status.counts && typeof status.counts === 'object' ? status.counts : {};
  const running = Number(status.processing_count ?? status.processingCount ?? counts.running) || 0;
  const queued = Number(status.pending_count ?? status.pendingCount ?? counts.queued) || 0;
  const waitingRetry = Number(counts.waiting_retry ?? counts.waitingRetry) || 0;
  const scope = status.scope && typeof status.scope === 'object' ? status.scope : {};
  const scopeDate = String(scope.date ?? '');
  const scopeBatchId = String(scope.batch_id ?? scope.batchId ?? '');
  return {
    label: running + queued + waitingRetry > 0 ? '队列运行中' : '队列空闲',
    enabled: Boolean(status.auto_queue_enabled ?? status.autoQueueEnabled ?? status.enabled),
    activeJobId: String(status.active_job_id ?? status.activeJobId ?? ''),
    latestJob: status.latest_job && typeof status.latest_job === 'object' ? status.latest_job : (status.latestJob || {}),
    running,
    queued,
    waitingRetry,
    scopeDate,
    scopeBatchId,
    scopeLabel: scopeDate || (scopeBatchId ? '当前批次' : '全部任务'),
    matchedDocuments: Number(scope.matched_documents ?? scope.matchedDocuments) || 0,
    updatedAt: String(status.updated_at ?? status.updatedAt ?? '')
  };
}

// A temporary polling failure must not look like a user-paused queue. Keep the
// last confirmed switch state until the backend explicitly returns a boolean.
export function mergeNovelFetchQueueStatus(previous = {}, next = {}) {
  const prior = previous && typeof previous === 'object' ? previous : {};
  const incoming = next && typeof next === 'object' ? next : {};
  const merged = { ...prior, ...incoming };
  for (const key of ['enabled', 'auto_queue_enabled', 'autoQueueEnabled']) {
    if (typeof incoming[key] !== 'boolean' && typeof prior[key] === 'boolean') merged[key] = prior[key];
  }
  return merged;
}

export function novelFetchTaskStatus(row = {}) {
  // 后端返回的 effective_status 是唯一有效状态；阶段状态只在旧接口缺字段时兜底。
  const canonical = text(row.effective_status ?? row.effectiveStatus);
  if (canonical && labels[canonical]) {
    return { status: canonical, label: labels[canonical], color: colors[canonical] };
  }
  const raw = text(row.status ?? row.taskStatus ?? row.task_status);
  if (raw === 'scheduled_waiting' || row.scheduledHold === true) {
    return { status: 'scheduled_waiting', label: labels.scheduled_waiting, color: colors.scheduled_waiting };
  }
  const original = row.originalStatus ?? row.original_status;
  const ai = row.aiStatus ?? row.ai_status;
  const submit = row.siteSubmitStatus ?? row.site_submit_status;
  const hasProcessedContent = done(original) || Number(row.originalRawChars ?? row.original_raw_chars) > 0;
  const hasAIContent = done(ai)
    || Number(row.aiGeneratedCount ?? row.ai_generated_count) > 0
    || (Array.isArray(row.aiFiles ?? row.ai_files) && (row.aiFiles ?? row.ai_files).some(version => /^ai[1-9]\d*$/i.test(String(version))));
  const submitConfirmed = done(submit);
  const pendingSubmit = hasProcessedContent && hasAIContent && !submitConfirmed;
  let status = raw;
  if (/fail|error/.test(raw)) status = 'failed';
  else if (/^(running|processing)$/.test(raw)) status = 'processing';
  else if (/^(done|success|succeeded|complete|completed)$/.test(raw)) status = pendingSubmit ? 'pending_submit' : 'completed';
  else if (/^(queue|queued|pending)$/.test(raw)) status = 'queued';
  else if (/^(waiting_retry|retrying)$/.test(raw)) status = 'waiting_retry';
  else if (/^waiting_classification$/.test(raw)) status = 'waiting_classification';
  else if (/^(waiting_sensitive_ai|waiting_ai_service)$/.test(raw)) status = 'waiting_sensitive_ai';
  else if (/^waiting_configuration$/.test(raw)) status = 'waiting_configuration';
  else if (/^waiting_login$/.test(raw)) status = 'waiting_login';

  if (!labels[status]) {
    if ([original, ai, submit].some(value => /fail|error/.test(text(value)))) status = 'failed';
    else if ([original, ai].some(value => /waiting_sensitive_ai|waiting_ai_service/.test(text(value)))) status = 'waiting_sensitive_ai';
    else if ([original, ai].some(value => /waiting_fetch_retry|waiting_retry|retrying/.test(text(value)))) status = 'waiting_retry';
    else if ([original, ai].some(value => /running|processing/.test(text(value)))) status = 'processing';
    else if (submitConfirmed) status = 'completed';
    else if (pendingSubmit) status = 'pending_submit';
    else status = 'input_ready';
  }

  return { status, label: labels[status], color: colors[status] };
}

export function novelFetchCharCount(row = {}) {
  const processed = Number(row.processedChars ?? row.processed_chars ?? row.wordCount ?? row.word_count ?? row.charCount ?? row.char_count);
  const original = Number(row.originalRawChars ?? row.original_raw_chars);
  if (processed > 0 && original > 0) return `${processed}/${original}`;
  if (original > 0) return String(original);
  if (processed > 0) return String(processed);
  return '-';
}
