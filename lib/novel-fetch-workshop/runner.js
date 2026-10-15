const {
  normalizeWorkflowPolicy,
  shortOriginalDecision,
  shouldReclassifyStyle,
  effectiveFetchConcurrency,
  effectiveRewriteConcurrency,
  splitSubmitBatches
} = require('./workflow-policy');
const { normalizeTargetVersions } = require('./target-versions');
const { reconcileAiTaskStatus } = require('./rewrite');

async function runWithConcurrency(items, limit, worker) {
  const source = Array.isArray(items) ? items : [];
  if (!source.length) return [];
  const results = new Array(source.length);
  let cursor = 0;
  const count = Math.max(1, Math.floor(Number(limit) || 1));
  const runners = Array.from({ length: Math.min(count, source.length) }, async () => {
    for (;;) {
      const index = cursor++;
      if (index >= source.length) return;
      try { results[index] = { ok: true, value: await worker(source[index], index) }; }
      catch (error) { results[index] = { ok: false, error }; }
    }
  });
  await Promise.all(runners);
  return results;
}

function object(value) { return value && typeof value === 'object' && !Array.isArray(value) ? value : {}; }

async function readSavedOriginal(tasks, username, bookId) {
  if (typeof tasks.readOriginal === 'function') return String(await tasks.readOriginal(username, bookId) || '');
  if (typeof tasks.getTask === 'function') {
    const record = await tasks.getTask(username, bookId);
    return String(record?.document?.original || '');
  }
  return '';
}

async function taskStopRequested(tasks, username, bookId, shouldStop) {
  if (typeof shouldStop === 'function' && shouldStop(bookId)) return true;
  if (typeof tasks?.getTask !== 'function') return false;
  const record = await tasks.getTask(username, bookId);
  return Boolean(record?.meta?.cancelRequestedAt || record?.meta?.status === 'cancelled');
}

async function markTaskCancelled(tasks, username, bookId, source = 'stop_requested') {
  if (typeof tasks?.updateTaskMeta !== 'function' || !bookId) return;
  const timestamp = new Date().toISOString();
  await tasks.updateTaskMeta(username, bookId, {
    status: 'cancelled',
    cancelRequestedAt: timestamp,
    cancelledAt: timestamp,
    error: ''
  });
  if (typeof tasks.appendLog === 'function') {
    await tasks.appendLog(username, bookId, 'task_cancelled', { source, cancelledAt: timestamp });
  }
}

async function existingOriginalReady(tasks, username, bookId) {
  if (typeof tasks?.getTask !== 'function') return false;
  const record = await tasks.getTask(username, bookId);
  if (!record?.meta) return false;
  const original = (await readSavedOriginal(tasks, username, bookId)).trim();
  if (!original) return false;
  if (record.meta.originalStatus !== 'done' && typeof tasks.updateTaskMeta === 'function') {
    await tasks.updateTaskMeta(username, bookId, {
      originalStatus: 'done', originalErrorCode: '', status: 'original_done', error: ''
    });
  }
  return true;
}

const RETRY_STAGE_ORDER = ['classify', 'original', 'sensitive', 'rewrite', 'submit'];
function retryStageRank(stage) {
  const index = RETRY_STAGE_ORDER.indexOf(String(stage || '').trim().toLowerCase());
  return index < 0 ? -1 : index;
}

async function runNovelFetchBatch({
  username,
  payload,
  configStore,
  tasks,
  parseBooks,
  classifyMissingRows,
  applyRules,
  generateAiVersions,
  submit,
  listTasks,
  webSessionReady,
  syncSiteStyles,
  throughput,
  shouldStop = () => false,
  sleep = ms => new Promise(resolve => setTimeout(resolve, ms)),
  report = () => {}
} = {}) {
  if (!username) throw new Error('username is required');
  if (!tasks || typeof tasks.saveTasks !== 'function') throw new Error('tasks store is required');
  if (!configStore || typeof configStore.getConfig !== 'function') throw new Error('configStore is required');
  if (typeof parseBooks !== 'function') throw new Error('parseBooks is required');

  const inputText = String(payload?.input_text || '');
  if (!inputText.trim()) throw new Error('请输入书籍信息');

  const config = object(await Promise.resolve(configStore.getConfig()));
  const retryStage = String(payload?.retry_stage || payload?.retryStage || '').trim().toLowerCase() || '';
  const retryRank = retryStageRank(retryStage);
  const policy = normalizeWorkflowPolicy(config);
  const workflow = object(config.workflow);
  const runStage = (stage, work) => (throughput && typeof throughput.run === 'function'
    ? throughput.run(stage, work, { shouldStop })
    : work());

  if (policy.workflow.auto_sync_site_styles) {
    try {
      const ready = typeof webSessionReady === 'function' ? await webSessionReady() : false;
      if (ready && typeof syncSiteStyles === 'function') {
        await syncSiteStyles();
        report({ type: 'style_sync', status: 'done', message: '121 风格目录已在本批处理开始前同步。' });
      } else {
        report({ type: 'style_sync', status: 'warning', message: '未检测到可用 121 登录会话，已跳过自动风格同步。' });
      }
    } catch (error) {
      report({ type: 'style_sync', status: 'warning', message: `121 风格同步失败：${error?.message || String(error)}` });
    }
  }

  const styles = typeof configStore.getStyles === 'function' ? (await Promise.resolve(configStore.getStyles()) || []) : [];
  const parsed = parseBooks({
    inputText,
    parseMode: payload?.parse_mode || 'smart',
    columnPresetId: payload?.column_preset_id || '',
    columnOrder: payload?.column_order || '',
    styles
  });
  const selectedTaskIds = new Set((Array.isArray(payload?.task_ids) ? payload.task_ids : []).map(value => String(value || '').trim()).filter(Boolean));
  const parsedTasks = (parsed?.tasks || []).filter(item => !selectedTaskIds.size || selectedTaskIds.has(String(item?.bookId || item?.book_id || item?.id || '').trim()));
  const platforms = typeof configStore.getPlatforms === 'function' ? (await Promise.resolve(configStore.getPlatforms()) || []) : [];
  const platformId = String(payload?.platform_id || '2');
  const platform = platforms.find(item => String(item?.id) === platformId) || {};
  const maxAiCount = Math.max(1, Math.min(Number(config.rewrite?.max_ai_count) || 5, 20));
  const defaultAiCount = Math.max(1, Math.min(Number(config.rewrite?.default_ai_count) || 1, maxAiCount));
  const aiCount = Math.max(1, Math.min(Number(payload?.ai_count) || defaultAiCount, maxAiCount));
  const explicitTargetVersions = Array.isArray(payload?.targetVersions)
    ? normalizeTargetVersions(payload.targetVersions)
    : (Array.isArray(payload?.target_versions) ? normalizeTargetVersions(payload.target_versions) : null);
  const autoSubmitVersions = explicitTargetVersions || Array.from({ length: aiCount }, (_, index) => `ai${index + 1}`);
  const batchId = String(payload?.batch_id || payload?.batchId || '').trim();
  const batchCreatedAt = String(payload?.batch_created_at || payload?.batchCreatedAt || new Date().toISOString());
  const prepared = parsedTasks.map(item => ({
    ...item,
    platformId,
    platformName: platform.name || platformId,
    maxTxt: Number(payload?.max_txt) || Number(config.fetch?.default_max_txt) || 4000,
    aiCount,
    ...(batchId ? { batchId, batchCreatedAt } : {}),
    batchOrder: parsedTasks.indexOf(item),
    status: 'queued'
  }));

  report({ type: 'parse', status: 'done', message: `解析完成：有效 ${prepared.length} 个任务。` });

  // Persist the lightweight queue rows before classification or upstream work so
  // the current batch is visible immediately and can survive a restart.
  await tasks.saveTasks(username, prepared);

  let classified = prepared;
  const classifyErrors = [];
  const shouldClassify = retryStage === 'classify'
    || (!retryStage && workflow.auto_classify_missing !== false);
  if (shouldClassify && prepared.length && typeof classifyMissingRows === 'function') {
    const result = await classifyMissingRows({ configStore, tasks: prepared });
    classified = result?.tasks || prepared;
    classifyErrors.push(...(result?.errors || []));
  }
  if (shouldClassify && policy.workflow.auto_reclassify_invalid_style && classified.length && typeof classifyMissingRows === 'function') {
    const indexes = [];
    const candidates = [];
    classified.forEach((task, index) => {
      if (String(task?.style || '').trim() && shouldReclassifyStyle(task.style, styles, true)) {
        indexes.push(index);
        candidates.push({ ...task, style: '', styleSource: '', classifyStatus: '', classifyError: '' });
      }
    });
    if (candidates.length) {
      const result = await classifyMissingRows({ configStore, tasks: candidates });
      const repaired = result?.tasks || candidates;
      indexes.forEach((targetIndex, position) => { classified[targetIndex] = repaired[position] || classified[targetIndex]; });
      classifyErrors.push(...(result?.errors || []));
    }
  }
  await tasks.saveTasks(username, classified);
  report({ type: 'classify', status: classifyErrors.length ? 'warning' : 'done', message: classifyErrors.length ? classifyErrors.join('；') : '风格和男女频补齐完成。' });

  let rawFetched = 0;
  let fetched = 0;
  let fetchFailed = 0;
  let cancelled = 0;
  let skippedShortOriginal = 0;
  const fetchedIds = [];
  if (retryStage || workflow.auto_fetch_original !== false) {
    const fetchedResults = await runWithConcurrency(classified, effectiveFetchConcurrency(config.fetch), async task => {
      if (await taskStopRequested(tasks, username, task.bookId, shouldStop)) {
        await markTaskCancelled(tasks, username, task.bookId, 'before_fetch');
        report({ type: 'fetch', status: 'cancelled', book_id: task.bookId, message: `${task.bookId} 已取消；当前已完成步骤已保留。` });
        return { task, stopped: true, done: false };
      }

      const originalReady = payload?.force_refetch === true
        ? false
        : await existingOriginalReady(tasks, username, task.bookId);
      if (retryRank > retryStageRank('original') && retryStage !== 'original') {
        if (!originalReady) return { task, done: false };
        if (retryStage === 'sensitive' && typeof applyRules === 'function') {
          await applyRules(tasks, username, task.bookId, config, configStore);
        }
        return { task, done: true, reused: true, skippedShort: false };
      }
      if (originalReady) {
        report({ type: 'fetch', status: 'done', book_id: task.bookId, message: `${task.bookId} 已使用现有原文继续处理。` });
        return { task, done: true, reused: true, skippedShort: false };
      }

      if (typeof tasks.updateTaskMeta === 'function') await tasks.updateTaskMeta(username, task.bookId, { status: 'original_processing', originalStatus: 'processing' });
      let result;
      try {
        result = await runStage('fetch', () => tasks.fetchOriginal(username, task.bookId, task.maxTxt));
      } catch (error) {
        const message = error?.message || String(error || '未知错误');
        if (typeof tasks.updateTaskMeta === 'function') {
          await tasks.updateTaskMeta(username, task.bookId, {
            status: 'original_failed',
            originalStatus: 'failed',
            originalErrorCode: 'FETCH_EXCEPTION',
            error: message
          });
        }
        report({ type: 'fetch', status: 'failed', book_id: task.bookId, message: `${task.bookId} 原文抓取异常：${message}` });
        return { task, done: false };
      }
      if (result?.status !== 'done') {
        report({ type: 'fetch', status: 'failed', book_id: task.bookId, message: `${task.bookId} 原文抓取失败。` });
        return { task, done: false };
      }
      rawFetched += 1;

      if (await taskStopRequested(tasks, username, task.bookId, shouldStop)) {
        await markTaskCancelled(tasks, username, task.bookId, 'after_fetch');
        return { task, stopped: true, done: true, fetchedOnly: true };
      }

      const original = (policy.fetch.skip_short_original && policy.fetch.min_original_chars > 0)
        ? await readSavedOriginal(tasks, username, task.bookId) : '';
      const short = shortOriginalDecision(original, policy.fetch);
      if (short.skipped) {
        if (typeof tasks.updateTaskMeta === 'function') {
          await tasks.updateTaskMeta(username, task.bookId, { status: 'skipped_short_original', shortOriginalChars: short.chars, minOriginalChars: policy.fetch.min_original_chars });
        }
        report({ type: 'fetch', status: 'warning', book_id: task.bookId, message: `${task.bookId} 原文过短，已保留原文并跳过后续改文。` });
        return { task, done: true, skippedShort: true, chars: short.chars };
      }
      if (typeof applyRules === 'function') await applyRules(tasks, username, task.bookId, config, configStore);

      if (await taskStopRequested(tasks, username, task.bookId, shouldStop)) {
        await markTaskCancelled(tasks, username, task.bookId, 'after_rules');
        report({ type: 'fetch', status: 'cancelled', book_id: task.bookId, message: `${task.bookId} 已取消；当前已完成步骤已保留。` });
        return { task, stopped: true, done: true };
      }
      report({ type: 'fetch', status: 'done', book_id: task.bookId, message: `${task.bookId} 原文已抓取并套用处理规则。` });
      return { task, done: true, skippedShort: false };
    });
    for (const [index, item] of fetchedResults.entries()) {
      const task = item?.value?.task || classified[index];
      if (item?.ok && item.value?.stopped) {
        cancelled += 1;
        continue;
      }
      if (item?.ok && item.value?.done) {
        if (!item.value.reused && !item.value.fetchedOnly) rawFetched += 0;
        if (item.value.skippedShort) {
          skippedShortOriginal += 1;
          continue;
        }
        fetched += 1;
        fetchedIds.push(task.bookId);
      } else {
        fetchFailed += 1;
      }
    }
  } else {
    report({ type: 'fetch', status: 'skipped', message: '配置未开启自动抓取原文。' });
  }

  let generated = 0;
  let rewriteFailed = 0;
  const completedVersionsByBook = new Map();
  const shouldRewrite = payload?.skip_rewrite !== true && (retryStage === 'rewrite'
    || (!retryStage && workflow.auto_rewrite_after_fetch)
    || (retryStage && retryRank < retryStageRank('rewrite') && workflow.auto_rewrite_after_fetch));
  if (shouldRewrite) {
    const rewriteConcurrency = effectiveRewriteConcurrency({ ...object(config.ai), force_serial_batch: policy.ai.force_serial_batch });
    const rewriteResults = await runWithConcurrency(fetchedIds, rewriteConcurrency, async bookId => {
      if (typeof generateAiVersions !== 'function') throw new Error('generateAiVersions is required');
      const task = await tasks.getTask(username, bookId);
      if (await taskStopRequested(tasks, username, bookId, shouldStop)) {
        await markTaskCancelled(tasks, username, bookId, 'before_rewrite');
        report({ type: 'rewrite', status: 'cancelled', book_id: bookId, message: `${bookId} 已取消；已生成文案已保留。` });
        return { status: 'cancelled', generated: [], generatedVersions: [] };
      }
      if (typeof tasks.updateTaskMeta === 'function') await tasks.updateTaskMeta(username, bookId, { status: 'ai_processing', aiStatus: 'generating' });
      try {
        const result = await runStage('rewrite', () => generateAiVersions({ configStore, tasks, username, task: task?.meta, count: task?.meta?.aiCount || 1, shouldStop }));
        if ((result?.status === 'failed' || result?.status === 'partial') && typeof tasks.updateTaskMeta === 'function') {
          const reconciled = await reconcileAiTaskStatus({
            tasks,
            username,
            bookId,
            task: task?.meta,
            lastAttemptError: result.error || 'AI文案生成失败'
          });
          await tasks.updateTaskMeta(username, bookId, {
            status: reconciled.status === 'done' ? 'ai_done' : (reconciled.status === 'partial' ? 'ai_partial' : 'ai_failed')
          });
        }
        report({ type: 'rewrite', status: result?.status === 'failed' ? 'failed' : 'done', book_id: bookId, message: result?.status === 'failed' ? `${bookId} AI 文案生成失败。` : `${bookId} AI 文案生成完成。` });
        return result;
      } catch (error) {
        if (typeof tasks.updateTaskMeta === 'function') {
          const reconciled = await reconcileAiTaskStatus({
            tasks,
            username,
            bookId,
            task: task?.meta,
            lastAttemptError: error?.message || 'AI文案生成失败'
          });
          await tasks.updateTaskMeta(username, bookId, {
            status: reconciled.status === 'done' ? 'ai_done' : (reconciled.status === 'partial' ? 'ai_partial' : 'ai_failed')
          });
        }
        report({ type: 'rewrite', status: 'failed', book_id: bookId, message: `${bookId} AI 文案生成失败：${error?.message || '未知错误'}` });
        throw error;
      }
    });
    for (const [index, item] of rewriteResults.entries()) {
      if (!item?.ok) {
        rewriteFailed += 1;
        continue;
      }
      if (item.value?.status === 'cancelled') {
        cancelled += 1;
        continue;
      }
      const completedVersions = (item.value?.generated || [])
        .filter(entry => entry.status === 'done' && entry.skipped !== true)
        .map(entry => String(entry.version || '').trim())
        .filter(Boolean);
      generated += completedVersions.length;
      if (completedVersions.length) completedVersionsByBook.set(fetchedIds[index], completedVersions);
      if (item.value?.status === 'failed') rewriteFailed += 1;
    }
    report({ type: 'rewrite', status: rewriteFailed ? 'warning' : 'done', message: `AI 文案生成 ${generated} 个版本${rewriteFailed ? `，失败 ${rewriteFailed} 本` : ''}。` });
  } else {
    report({ type: 'rewrite', status: 'skipped', message: '配置未开启抓取后自动改文。' });
  }

  const retryingSubmit = retryStage === 'submit';
  const scheduledAutoSubmit = payload?.scheduled === true && payload?.auto_submit !== false;
  const shouldSubmit = retryingSubmit
    || scheduledAutoSubmit
    || (workflow.auto_submit_after_rewrite === true && workflow.auto_submit_confirmed === true);
  if (typeof tasks.updateTaskMeta === 'function') {
    for (const [bookId, versions] of completedVersionsByBook.entries()) {
      await tasks.updateTaskMeta(username, bookId, {
        siteSubmitStatus: 'pending_upload',
        siteSubmitPendingVersions: versions,
        siteSubmitError: ''
      });
    }
  }
  if (!shouldStop() && shouldSubmit && config.web_submit?.enabled === true && fetchedIds.length && typeof submit === 'function') {
    if (policy.web_submit.flush_seconds > 0) await sleep(policy.web_submit.flush_seconds * 1000);
    let successGroups = 0;
    let failedGroups = 0;
    for (const ids of splitSubmitBatches(fetchedIds, policy.web_submit.batch_size)) {
      if (shouldStop()) break;
      const retrySubmitVersions = retryStage === 'submit' && Array.isArray(payload?.retry_submit_versions)
        ? normalizeTargetVersions(payload.retry_submit_versions)
        : null;
      const submitVersions = retrySubmitVersions?.length ? retrySubmitVersions : autoSubmitVersions;
      const submitted = await runStage('submit', () => submit({ mode: 'selected', ids, versions: submitVersions }));
      successGroups += Number(submitted?.success_groups) || 0;
      failedGroups += Number(submitted?.failed_groups) || 0;
    }
    report({ type: 'submit', status: failedGroups ? 'warning' : 'done', message: `网站提交：成功 ${successGroups} 组，失败 ${failedGroups} 组。` });
  } else {
    report({ type: 'submit', status: 'skipped', message: shouldStop() ? '已停止处理；网站提交未继续执行。' : '未开启自动网站提交；文案已保留，可在任务页点击“提交网络”。' });
  }

  const currentBatchIds = prepared.map(task => task.bookId);
  const allListedTasks = typeof listTasks === 'function' ? await listTasks() : [];
  const listedByBookId = new Map((Array.isArray(allListedTasks) ? allListedTasks : []).map(task => [String(task?.bookId || task?.book_id || task?.id || '').trim(), task]));
  const currentBatchTasks = currentBatchIds.map(bookId => listedByBookId.get(String(bookId || '').trim())).filter(Boolean);

  return {
    parsed: prepared.length,
    unique_tasks: prepared.length,
    duplicate_count: parsed?.duplicateCount || 0,
    empty_id_count: parsed?.emptyIdCount || 0,
    classify_errors: classifyErrors,
    raw_fetched: rawFetched,
    fetched,
    skipped_short_original: skippedShortOriginal,
    cancelled,
    process_failed: 0,
    rewrite_failed: rewriteFailed,
    fetch_failed: fetchFailed,
    generated_ai_files: generated,
    current_batch_ids: currentBatchIds,
    tasks: currentBatchTasks
  };
}

module.exports = { runWithConcurrency, readSavedOriginal, taskStopRequested, markTaskCancelled, existingOriginalReady, runNovelFetchBatch };
