const { normalizeTargetVersions, hasExplicitTargetVersions } = require('./target-versions');
const { normalizeAiSlotMethods } = require('./version-selection');

function object(value) { return value && typeof value === 'object' && !Array.isArray(value) ? value : {}; }
function taskBookId(task) { return String(task?.bookId || task?.book_id || task?.id || ''); }
function taskTimestamp(task) {
  return String(task?.batchCreatedAt || task?.batch_created_at || task?.createdAt || task?.created_at || task?.updatedAt || task?.updated_at || '');
}
const taskDateFormatter = new Intl.DateTimeFormat('en-CA', {
  timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit'
});
function dateKey(value) {
  const parsed = new Date(value);
  if (!Number.isFinite(parsed.getTime())) return '';
  const parts = Object.fromEntries(taskDateFormatter.formatToParts(parsed)
    .filter(part => part.type !== 'literal')
    .map(part => [part.type, part.value]));
  return `${parts.year}-${parts.month}-${parts.day}`;
}
function taskStatusText(task) {
  return [task?.status, task?.originalStatus, task?.original_status, task?.aiStatus, task?.ai_status, task?.classifyStatus, task?.classify_status, task?.sensitiveStatus, task?.sensitive_status, task?.siteSubmitStatus, task?.site_submit_status, task?.error, task?.classifyError, task?.originalErrorCode, task?.aiError, task?.siteSubmitError]
    .filter(Boolean).join(' ').toLowerCase();
}
function effectiveTaskStatus(task = {}) {
  const current = String(task.status || '').trim();
  const site = String(task.siteSubmitStatus || task.site_submit_status || '').trim();
  const ai = String(task.aiStatus || task.ai_status || '').trim();
  const original = String(task.originalStatus || task.original_status || '').trim();
  if (/^(submitted|accepted_pending)$/.test(site)
    && /^(ai_processing|generating|processing|running)$/.test(current)) return site;
  if (ai === 'done' && original === 'done' && /^(ai_processing|generating|processing|running)$/.test(current)) return 'ai_done';
  if (!ai && original === 'done' && /^(ai_processing|generating|processing|running)$/.test(current)) return 'original_done';
  return current;
}
function isCancelledTask(task) { return /(cancelled|canceled|已取消)/i.test(taskStatusText(task)); }
function isTaskUnfinished(task) {
  const text = taskStatusText(task);
  if (!text) return true;
  if (isCancelledTask(task)) return false;
  if (/(failed|error|running|queued|waiting|created|paused|stopping|interrupted|incomplete|失败|错误|处理中|等待|中断|未完成|121异常)/i.test(text)) return true;
  if (/(done|completed|submitted|success|完成|成功)/i.test(text)) return false;
  return true;
}
function isAbnormalTask(task) {
  if (isCancelledTask(task)) return false;
  if (Array.isArray(task?.siteSubmitFailedVersions || task?.site_submit_failed_versions) && (task.siteSubmitFailedVersions || task.site_submit_failed_versions).length) return true;
  if (Number(task?.sensitiveFailedCount ?? task?.sensitive_failed_count) > 0) return true;
  return /(failed|error|timeout|interrupted|incomplete|失败|错误|超时|中断|未完成|121异常|partial|部分)/i.test(taskStatusText(task));
}
function filterTaskList(tasks, query = {}, now = new Date(), { currentBatchIds = [] } = {}) {
  const source = Array.isArray(tasks) ? tasks : [];
  const current = new Set((Array.isArray(currentBatchIds) ? currentBatchIds : []).map(value => String(value || '').trim()).filter(Boolean));
  const wantedDate = String(query.date || '').trim();
  const wantedBook = String(query.bookId || query.book_id || '').trim().toLowerCase();
  const wantedStatus = String(query.status || '').trim().toLowerCase();
  const wantedBatch = String(query.batchId || query.batch_id || '').trim();
  const scope = String(query.scope || '').trim().toLowerCase();
  const explicit = Boolean(wantedDate || wantedBook || wantedStatus || wantedBatch || scope);
  const today = dateKey(now);
  return source.filter(task => {
    const day = dateKey(taskTimestamp(task));
    const id = taskBookId(task).toLowerCase();
    const status = taskStatusText(task);
    if (wantedDate && day !== wantedDate) return false;
    if (wantedBatch && String(task?.batchId || task?.batch_id || '').trim() !== wantedBatch) return false;
    if (wantedBook && !id.includes(wantedBook)) return false;
    if (wantedStatus && !status.includes(wantedStatus)) return false;
    if (scope === 'current') return current.has(taskBookId(task));
    if (scope === 'unfinished') return isTaskUnfinished(task);
    if (scope === 'all') return true;
    if (explicit) return true;
    if (current.has(taskBookId(task))) return true;
    return day === today || (day && day < today && isTaskUnfinished(task));
  }).sort((left, right) => {
    if (current.size) {
      const leftOrder = current.has(taskBookId(left)) ? [...current].indexOf(taskBookId(left)) : Number.MAX_SAFE_INTEGER;
      const rightOrder = current.has(taskBookId(right)) ? [...current].indexOf(taskBookId(right)) : Number.MAX_SAFE_INTEGER;
      if (leftOrder !== rightOrder) return leftOrder - rightOrder;
    }
    return taskTimestamp(right).localeCompare(taskTimestamp(left));
  });
}

function explicitAiTargets(task) {
  if (!hasExplicitTargetVersions(task)) return null;
  return normalizeTargetVersions(task.targetVersions).filter(version => /^ai[1-5]$/.test(version));
}

function toV78Task(task = {}) {
  const explicitTargets = hasExplicitTargetVersions(task) ? normalizeTargetVersions(task.targetVersions) : null;
  const targetAi = explicitAiTargets(task);
  const legacyAiCount = Number(task.aiCount || task.ai_count) || 1;
  const selectedVersions = explicitTargets || normalizeTargetVersions(undefined, legacyAiCount);
  const slotMethods = normalizeAiSlotMethods(
    task.aiSlotMethodsSnapshot || task.ai_slot_methods_snapshot || task.aiSlotMethods || task.ai_slot_methods
  );
  const legacyGenerated = Number(task.aiGeneratedCount || task.ai_generated_count) || 0;
  const explicitGenerated = Array.isArray(task.aiGeneratedVersions)
    ? normalizeTargetVersions(task.aiGeneratedVersions).filter(version => /^ai[1-5]$/.test(version))
    : null;
  const aiFiles = explicitGenerated || (Array.isArray(task.ai_files)
    ? task.ai_files
    : Array.from({ length: legacyGenerated }, (_, index) => `ai${index + 1}`));
  return {
    ...task,
    // Older runs could finish a stage without replacing the transient
    // top-level status. Derive the visible status from the terminal stage so
    // completed work is not shown as permanently "AI processing".
    status: effectiveTaskStatus(task),
    id: task.bookId || task.book_id || task.id || '',
    book_id: task.bookId || task.book_id || task.id || '',
    book_name: task.bookName || task.book_name || '',
    platform_id: task.platformId || task.platform_id || '',
    platform_name: task.platformName || task.platform_name || '',
    parse_mode: task.parseMode || task.parse_mode || '',
    push_date: task.pushDate || task.push_date || task.pushedAt || task.pushed_at || task.batchCreatedAt || task.batch_created_at || task.createdAt || task.created_at || '',
    batch_id: task.batchId || task.batch_id || '',
    batch_created_at: task.batchCreatedAt || task.batch_created_at || '',
    batch_order: Number(task.batchOrder ?? task.batch_order) || 0,
    original_status: task.originalStatus || task.original_status || '',
    original_chars: Number(task.originalChars ?? task.original_chars) || 0,
    original_raw_chars: Number(task.originalRawChars ?? task.original_raw_chars) || 0,
    original_error: task.originalErrorMessage || task.original_error || task.error || '',
    original_error_code: task.originalErrorCode || task.original_error_code || '',
    original_upstream_code: task.originalUpstreamCode ?? task.original_upstream_code ?? null,
    ai_status: task.aiStatus || task.ai_status || '',
    ai_error: task.aiError || task.ai_error || '',
    ai_current_version: task.aiCurrentVersion || task.ai_current_version || '',
    ai_generated_count: Number(task.aiGeneratedCount ?? task.ai_generated_count) || 0,
    ai_count: targetAi ? targetAi.length : legacyAiCount,
    selected_versions: selectedVersions,
    ai_slot_methods: slotMethods,
    ai_files: aiFiles,
    target_versions: explicitTargets,
    ai_target_versions: targetAi,
    ai_generated_versions: aiFiles,
    classify_status: task.classifyStatus || task.classify_status || '',
    classify_error: task.classifyError || task.classify_error || '',
    classifier_model: task.classifierModel || task.classifier_model || '',
    retry_stage: task.retryStage || task.retry_stage || '',
    sensitive_hit_count: Number(task.sensitiveHitCount ?? task.sensitive_hit_count) || 0,
    sensitive_fixed_count: Number(task.sensitiveFixedCount ?? task.sensitive_fixed_count) || 0,
    sensitive_failed_count: Number(task.sensitiveFailedCount ?? task.sensitive_failed_count) || 0,
    site_submit_status: task.siteSubmitStatus || task.site_submit_status || '',
    site_submit_error: task.siteSubmitError || task.site_submit_error || '',
    site_submit_done_versions: task.siteSubmitDoneVersions || task.site_submit_done_versions || [],
    site_submit_accepted_versions: task.siteSubmitAcceptedVersions || task.site_submit_accepted_versions || [],
    site_submit_failed_versions: task.siteSubmitFailedVersions || task.site_submit_failed_versions || [],
    created_at: task.createdAt || task.created_at || '',
    updated_at: task.updatedAt || task.updated_at || '',
    error: task.error || '',
    cancelled: isCancelledTask(task)
  };
}

function normalizeIds(ids) {
  return [...new Set((Array.isArray(ids) ? ids : [ids]).map(value => String(value || '').trim()).filter(Boolean))];
}

function retryInputText(meta = {}) {
  const headers = ['书籍ID', '书名', '男女频', '风格', '标签', '推荐理由', '评级'];
  const values = [meta.bookId, meta.bookName, meta.gender, meta.style, meta.tags, meta.reason, meta.rating]
    .map(value => String(value || '').replace(/[\t\r\n]+/g, ' ').trim());
  return `${headers.join('\t')}\n${values.join('\t')}`;
}

const RETRY_STAGES = ['classify', 'original', 'sensitive', 'rewrite', 'submit'];

function hasFailedValue(...values) {
  return values.some(value => /(failed|error|timeout|interrupted|incomplete|失败|错误|超时|中断|未完成|121异常|partial|部分)/i.test(String(value || '')));
}

function retryStageForTask(task = {}) {
  const classifyFailed = hasFailedValue(task.classifyStatus, task.classify_status, task.classifyError, task.classify_error);
  if (classifyFailed || /waiting_ai_config|等待.*ai.*配置/i.test(String(task.classifyStatus || task.classify_status || ''))) return 'classify';
  const originalFailed = hasFailedValue(task.originalStatus, task.original_status, task.originalErrorCode, task.original_error_code);
  if (originalFailed) return 'original';
  const sensitiveFailed = hasFailedValue(task.sensitiveStatus, task.sensitive_status) || Number(task.sensitiveFailedCount ?? task.sensitive_failed_count) > 0;
  if (sensitiveFailed) return 'sensitive';
  const submitFailedVersions = task.siteSubmitFailedVersions || task.site_submit_failed_versions;
  const rewriteFailed = hasFailedValue(task.aiStatus, task.ai_status, task.aiError, task.ai_error);
  if (rewriteFailed) return 'rewrite';
  if (hasFailedValue(task.siteSubmitStatus, task.site_submit_status, task.siteSubmitError, task.site_submit_error)
    || (Array.isArray(submitFailedVersions) && submitFailedVersions.length)) return 'submit';
  return null;
}

function retryStageLabel(stage) {
  return ({ classify: 'AI 判断', original: '原文抓取', sensitive: '敏感词处理', rewrite: 'AI 文案生成', submit: '网站提交' })[stage] || '失败步骤';
}

function createNovelFetchTaskOps({ accountResolver, createStore, tombstones, parseBooks, clock = () => new Date(), applySavedRules, createConfigSnapshot } = {}) {
  if (typeof accountResolver !== 'function') throw new Error('accountResolver is required');
  if (typeof createStore !== 'function') throw new Error('createStore is required');
  if (!tombstones) throw new Error('tombstones is required');
  if (typeof parseBooks !== 'function') throw new Error('parseBooks is required');
  const batchExecutor = (!applySavedRules || !createConfigSnapshot) ? require('./v2-batch-executor') : null;
  const applySavedRulesToOriginal = applySavedRules || batchExecutor.createApplySavedRulesToOriginal();
  const makeConfigSnapshot = createConfigSnapshot || batchExecutor.createConfigStoreSnapshot;

  function accountFor(owner) {
    const account = accountResolver(owner);
    if (!account?.username || account.username !== owner) {
      const error = new Error(`账号 ${owner} 不存在或不可用`);
      error.recoverable = false;
      throw error;
    }
    return account;
  }
  function storeFor(owner) { return createStore({ account: accountFor(owner) }); }

  function processConflicts(owner, payload = {}) {
    const parsed = parseBooks({
      inputText: String(payload.input_text || ''),
      parseMode: payload.parse_mode || 'smart',
      columnPresetId: payload.column_preset_id || '',
      columnOrder: payload.column_order || '',
      styles: []
    });
    const ids = normalizeIds((parsed?.tasks || []).map(task => task.bookId || task.book_id || task.id));
    return ids.filter(id => tombstones.has(owner, id));
  }

  async function previewInput(owner, payload = {}) {
    const store = storeFor(owner);
    const config = typeof store.getConfig === 'function' ? object(await store.getConfig()) : {};
    const styles = typeof store.getStyles === 'function' ? (await store.getStyles() || []) : (Array.isArray(config.styles) ? config.styles : []);
    const parsed = parseBooks({
      inputText: String(payload.input_text || ''),
      parseMode: payload.parse_mode || 'smart',
      columnPresetId: payload.column_preset_id || '',
      columnOrder: payload.column_order || '',
      styles
    });
    return {
      parsed: Number(parsed?.parsed) || 0,
      unique_tasks: Number(parsed?.uniqueTasks) || 0,
      duplicate_count: Number(parsed?.duplicateCount) || 0,
      empty_id_count: Number(parsed?.emptyIdCount) || 0,
      tasks: (Array.isArray(parsed?.tasks) ? parsed.tasks : []).map(task => ({
        ...toV78Task(task),
        source_line: task.sourceLine || task.source_line || '',
        selected: true
      }))
    };
  }

  async function permanentDelete(owner, ids) {
    const normalized = normalizeIds(ids);
    const store = storeFor(owner);
    const result = await store.deleteTasks(owner, normalized);
    tombstones.add(owner, normalized);
    return { deleted: Number(result?.deleted) || normalized.length, tombstoned: normalized };
  }

  function restoreTombstone(owner, bookId) { accountFor(owner); return tombstones.restore(owner, String(bookId || '').trim()); }

  async function list(owner, query = {}, options = {}) {
    const store = storeFor(owner);
    return filterTaskList(await store.listTasks(owner), query, clock(), options).map(toV78Task);
  }

  async function setSelectedVersions(owner, ids, selectedVersions, aiSlotMethods = {}) {
    const versions = normalizeTargetVersions(Array.isArray(selectedVersions) ? selectedVersions : []);
    if (!versions.length) throw new Error('请至少选择一个文案版本');
    const methods = normalizeAiSlotMethods(aiSlotMethods);
    const normalized = normalizeIds(ids);
    const store = storeFor(owner);
    const missing = [];
    const records = [];
    for (const id of normalized) {
      const record = await store.getTask(owner, id);
      if (!record) { missing.push(id); continue; }
      records.push([id, record]);
    }
    if (missing.length) return { ok: false, missing, updated: 0 };
    for (const [id] of records) {
      await store.updateTaskMeta(owner, id, {
        targetVersions: versions,
        selectedVersions: versions,
        aiSlotMethodsSnapshot: methods,
        aiSlotMethods: methods,
        aiCount: versions.filter(version => /^ai[1-5]$/.test(version)).length
      });
    }
    return {
      ok: true,
      missing: [],
      updated: records.length,
      selectedVersions: versions,
      targetVersions: versions,
      aiSlotMethods: methods
    };
  }

  async function setAiCount(owner, ids, aiCount) {
    const count = Number(aiCount);
    if (!Number.isInteger(count) || count < 1 || count > 20) throw new Error('AI 数量必须在 1 到 20 之间');
    const normalized = normalizeIds(ids);
    const store = storeFor(owner);
    const conflicts = [];
    const missing = [];
    const records = [];
    for (const id of normalized) {
      const record = await store.getTask(owner, id);
      if (!record) { missing.push(id); continue; }
      records.push([id, record]);
      const versions = object(record.document?.versions);
      if (Object.keys(versions).some(key => /^ai\d+$/.test(key))) conflicts.push(id);
    }
    if (conflicts.length || missing.length) return { ok: false, conflicts, missing, updated: 0 };
    for (const [id] of records) await store.updateTaskMeta(owner, id, { aiCount: count });
    return { ok: true, conflicts: [], missing: [], updated: records.length, aiCount: count };
  }

  async function cancelSelected(owner, ids) {
    const normalized = normalizeIds(ids);
    const store = storeFor(owner);
    const cancelledAt = clock().toISOString();
    let updated = 0;
    const missing = [];
    for (const id of normalized) {
      const record = await store.getTask(owner, id);
      if (!record?.meta) { missing.push(id); continue; }
      await store.updateTaskMeta(owner, id, {
        status: 'cancelled',
        cancelRequestedAt: cancelledAt,
        cancelledAt,
        error: ''
      });
      if (typeof store.appendLog === 'function') {
        await store.appendLog(owner, id, 'task_cancelled', { source: 'manual_selected', cancelledAt });
      }
      updated += 1;
    }
    return { updated, missing, status: 'cancelled' };
  }

  async function abnormalIds(owner, query = {}) {
    const store = storeFor(owner);
    return filterTaskList(await store.listTasks(owner), query, clock()).filter(isAbnormalTask).map(taskBookId).filter(Boolean);
  }

  async function prepareRetryPayloads(owner, ids) {
    const normalized = normalizeIds(ids);
    const store = storeFor(owner);
    const payloads = [];
    for (const id of normalized) {
      const record = await store.getTask(owner, id);
      if (!record?.meta) continue;
      const meta = record.meta;
      const retryStage = retryStageForTask(meta);
      // 仅允许失败阶段进入重试；已完成、运行中或排队中的任务不重复加入队列。
      if (!retryStage) continue;
      const failedSubmitVersions = (meta.siteSubmitFailedVersions || meta.site_submit_failed_versions || [])
        .map(version => String(version || '').trim()).filter(Boolean);
      const reset = {
        status: 'queued',
        retryStage,
        cancelRequestedAt: '',
        cancelledAt: '',
        error: ''
      };
      if (retryStage === 'classify') Object.assign(reset, {
        classifyStatus: '', classifyError: '', classifyConfidence: null, classifyReason: '', classifierModel: ''
      });
      if (retryStage === 'original') Object.assign(reset, { originalStatus: '', originalErrorCode: '' });
      if (retryStage === 'sensitive') Object.assign(reset, { sensitiveStatus: '', sensitiveFailedCount: 0 });
      if (retryStage === 'rewrite') Object.assign(reset, { aiStatus: '', aiError: '' });
      if (retryStage === 'submit') Object.assign(reset, {
        siteSubmitStatus: '', siteSubmitError: '', siteSubmitFailedVersions: []
      });
      await store.updateTaskMeta(owner, id, reset);
      const payload = {
        platform_id: String(meta.platformId || meta.platform_id || '2'),
        input_text: retryInputText(meta),
        parse_mode: 'header',
        max_txt: Number(meta.maxTxt || meta.max_txt) || 4000,
        retry_existing_task: true,
        retry_stage: retryStage
      };
      if (failedSubmitVersions.length) payload.retry_submit_versions = failedSubmitVersions;
      if (hasExplicitTargetVersions(meta)) {
        payload.target_versions = normalizeTargetVersions(meta.targetVersions);
        payload.selected_versions = normalizeTargetVersions(meta.targetVersions);
        payload.ai_slot_methods = { ...object(meta.aiSlotMethodsSnapshot || meta.ai_slot_methods_snapshot || meta.aiSlotMethods || meta.ai_slot_methods) };
        payload.ai_slot_methods_snapshot = { ...payload.ai_slot_methods };
      } else {
        payload.ai_count = Math.max(1, Number(meta.aiCount || meta.ai_count) || 1);
      }
      const retryVersions = (payload.retry_submit_versions || payload.target_versions || []).map(value => String(value || '').trim()).filter(Boolean).sort();
      payload.retry_idempotency_key = `${owner}:${id}:${retryStage}:${retryVersions.join(',')}`;
      payloads.push(payload);
    }
    return payloads;
  }

  async function detail(owner, id) {
    const store = storeFor(owner);
    const record = await store.getTask(owner, id);
    if (!record?.meta) return null;
    const meta = record.meta;
    let aiVersions;
    if (hasExplicitTargetVersions(meta)) {
      aiVersions = normalizeTargetVersions(meta.targetVersions).filter(version => /^ai[1-5]$/.test(version));
    } else {
      const versions = object(record.document?.versions);
      aiVersions = Object.keys(versions).filter(version => /^ai\d+$/.test(version)).sort((a, b) => Number(a.slice(2)) - Number(b.slice(2)));
      if (!aiVersions.length) aiVersions = Array.from({ length: Number(meta.aiGeneratedCount) || 0 }, (_, index) => `ai${index + 1}`);
    }
    const aiTexts = [];
    for (const version of aiVersions) {
      const text = typeof store.readVersionText === 'function' ? await store.readVersionText(owner, id, version) : String(record.document?.versions?.[version] || '');
      if (String(text || '').trim()) aiTexts.push({ name: version.toUpperCase(), version, text });
    }
    const original = typeof store.readOriginal === 'function' ? await store.readOriginal(owner, id) : String(record.document?.original || '');
    const logs = typeof store.readLogs === 'function' ? await store.readLogs(owner, id) : (Array.isArray(record.document?.logs) ? record.document.logs : []);
    return {
      meta: toV78Task(meta),
      original,
      ai_texts: aiTexts,
      target_versions: hasExplicitTargetVersions(meta) ? normalizeTargetVersions(meta.targetVersions) : null,
      selected_versions: toV78Task(meta).selected_versions,
      ai_slot_methods: toV78Task(meta).ai_slot_methods,
      has_original_raw: record.hasOriginalRaw === true || Boolean(record.document?.originalRaw),
      logs
    };
  }

  async function reprocessSensitive(owner, body = {}) {
    const store = storeFor(owner);
    const config = object(await store.getConfig());
    const configStore = makeConfigSnapshot(store, config);
    const all = await store.listTasks(owner);
    let ids;
    if (body.mode === 'selected') ids = normalizeIds(body.ids || []);
    else if (body.mode === 'failed') ids = all.filter(task => /failed|error|失败|错误/i.test(taskStatusText(task)) && !isCancelledTask(task)).map(taskBookId);
    else ids = all.map(taskBookId).filter(Boolean);
    const restoreFromBackup = body.restore_from_backup === true;
    const sensitiveConfig = object(config.sensitive_ai);
    const configuredMode = String(sensitiveConfig.mode || '').trim();
    const mode = ['replace', 'ai_each', 'ai_group'].includes(configuredMode)
      ? configuredMode
      : (sensitiveConfig.enabled === false ? 'replace' : 'ai_each');
    let processed = 0;
    let restored = 0;
    let failed = 0;
    for (const id of normalizeIds(ids)) {
      try {
        const task = await store.getTask(owner, id);
        if (!task?.meta) throw new Error('任务不存在');
        if (restoreFromBackup) {
          if (typeof store.restoreOriginal !== 'function') throw new Error('当前存储不支持恢复原文备份');
          await store.restoreOriginal(owner, id);
          restored += 1;
        }
        await applySavedRulesToOriginal(store, owner, id, config, configStore);
        if (typeof store.appendLog === 'function') {
          await store.appendLog(owner, id, 'sensitive_reprocessed', {
            source: restoreFromBackup ? 'original_backup' : 'current_text',
            mode,
            sensitive_ai_enabled: mode !== 'replace'
          });
        }
        processed += 1;
      } catch (_) {
        failed += 1;
      }
    }
    return { processed, restored, failed, tasks: (await store.listTasks(owner)).map(toV78Task) };
  }

  return {
    processConflicts,
    previewInput,
    permanentDelete,
    restoreTombstone,
    list,
    setSelectedVersions,
    setAiCount,
    cancelSelected,
    abnormalIds,
    prepareRetryPayloads,
    retryStageForTask,
    detail,
    reprocessSensitive
  };
}

module.exports = {
  isCancelledTask,
  isAbnormalTask,
  isTaskUnfinished,
  filterTaskList,
  toV78Task,
  retryInputText,
  retryStageForTask,
  retryStageLabel,
  createNovelFetchTaskOps
};
