// Original batch-rewrite UI compatibility layer.  The imported browser assets
// keep their original request contract while data stays in this application's
// authenticated workshop and upload stores.
const crypto = require('node:crypto');
const express = require('express');
const { apiAuth } = require('../middleware/auth');
const { createMySQLWorkshopStore } = require('../lib/novel-fetch-workshop/mysql-store');
const { getKnowledgeStore } = require('../lib/novel-fetch-workshop/knowledge');
const { createOpeningStore } = require('../lib/novel-fetch-workshop/opening');
const ai = require('../lib/novel-fetch-workshop/ai');
const classifier = require('../lib/novel-fetch-workshop/classifier');
const rewrite = require('../lib/novel-fetch-workshop/rewrite');
const versionSelection = require('../lib/novel-fetch-workshop/version-selection');
const parse = require('../lib/novel-fetch-workshop/parse');
const rules = require('../lib/novel-fetch-workshop/rules');
const sensitive = require('../lib/novel-fetch-workshop/sensitive');
const target = require('../lib/target-upload');
const { PLATFORMS, STYLE_NAMES } = require('./novel-fetch');
const { mergeKnowledgeSources } = require('../lib/novel-fetch-workshop/knowledge-recovery');
const { retryStageForTask } = require('../lib/novel-fetch-workshop/task-ops');

const jobs = new Map();
const LEGACY_KINDS = ['high_imitation', 'opening_phrases', 'rewrite_templates', 'layout_rules', 'symbol_rules', 'chapter_rules'];
const PROCESS_JOB_LIMIT = 20;
const CONFIG_CACHE_TTL_MS = 5000;
const HISTORICAL_AI_REPAIR_CONCURRENCY = 2;
const historicalAiStatusRepairJobs = new Map();
const historicalAiStatusRepairCompleted = new Set();

async function runWithConcurrency(items, limit, worker) {
  const results = new Array(items.length);
  let cursor = 0;
  const count = Math.max(1, Math.floor(Number(limit) || 1));
  const runners = Array.from({ length: Math.min(count, items.length) }, async () => {
    for (;;) {
      const index = cursor++;
      if (index >= items.length) return;
      try { results[index] = { ok: true, value: await worker(items[index]) }; }
      catch (error) { results[index] = { ok: false, error }; }
    }
  });
  await Promise.all(runners);
  return results;
}

function needsHistoricalAiStatusRepair(task = {}) {
  if (task.hasAi !== true) return false;
  const aiStatus = String(task.aiStatus || task.ai_status || '').toLowerCase();
  const taskStatus = String(task.status || '').toLowerCase();
  return aiStatus === 'failed' || taskStatus === 'ai_failed';
}

// 历史校正是迁移任务，不得堵塞任务表、重试或提交接口。每个账号在当前
// Node 进程内只运行一次，且以低并发读取已保存文案；生成中的任务不会进入。
function scheduleHistoricalAiStatusRepair({ tasks, username, records } = {}) {
  const owner = String(username || '').trim();
  if (!owner || historicalAiStatusRepairCompleted.has(owner) || historicalAiStatusRepairJobs.has(owner)) return;
  const candidates = (Array.isArray(records) ? records : []).filter(needsHistoricalAiStatusRepair);
  if (!candidates.length) {
    historicalAiStatusRepairCompleted.add(owner);
    return;
  }
  const job = Promise.resolve().then(() => runWithConcurrency(candidates, HISTORICAL_AI_REPAIR_CONCURRENCY, async task => {
    const bookId = task.bookId || task.book_id || task.id;
    if (!bookId) return;
    await rewrite.reconcileAiTaskStatus({ tasks, username: owner, bookId, task, recordLog: true });
  })).catch(() => undefined).finally(() => {
    historicalAiStatusRepairJobs.delete(owner);
    historicalAiStatusRepairCompleted.add(owner);
  });
  historicalAiStatusRepairJobs.set(owner, job);
}

function snakeTask(task = {}) {
  const bookId = task.bookId || task.book_id || task.id || '';
  const bookName = task.bookName || task.book_name || '';
  const platformId = task.platformId || task.platform_id || '';
  const platformName = task.platformName || task.platform_name || '';
  const aiCount = Number(task.aiCount ?? task.ai_count) || 1;
  const existingAiFiles = Array.isArray(task.ai_files) ? task.ai_files : [];
  const aiGenerated = Number(task.aiGeneratedCount ?? task.ai_generated_count ?? existingAiFiles.length) || 0;
  return {
    ...task,
    id: bookId,
    book_id: bookId,
    book_name: bookName,
    platform_id: platformId,
    platform_name: platformName,
    parse_mode: task.parseMode || task.parse_mode || '',
    original_status: task.originalStatus || task.original_status || '',
    original_chars: Number(task.originalChars ?? task.original_chars) || 0,
    original_raw_chars: Number(task.originalRawChars ?? task.original_raw_chars) || 0,
    original_error: task.originalErrorMessage || task.original_error || task.error || '',
    original_error_code: task.originalErrorCode || task.original_error_code || '',
    original_upstream_code: task.originalUpstreamCode ?? task.original_upstream_code ?? null,
    original_refresh_error: task.originalRefreshErrorMessage || task.original_refresh_error || '',
    original_refresh_error_code: task.originalRefreshErrorCode || task.original_refresh_error_code || '',
    original_refresh_upstream_code: task.originalRefreshUpstreamCode ?? task.original_refresh_upstream_code ?? null,
    original_refresh_at: task.originalRefreshAt || task.original_refresh_at || '',
    classify_error: task.classifyError || task.classify_error || '',
    ai_status: task.aiStatus || task.ai_status || '',
    ai_error: task.aiError || task.ai_error || '',
    ai_current_version: task.aiCurrentVersion || task.ai_current_version || '',
    ai_count: aiCount,
    ai_files: existingAiFiles.length ? existingAiFiles : Array.from({ length: aiGenerated }, (_, index) => `ai${index + 1}`),
    classify_status: task.classifyStatus || task.classify_status || '',
    classifier_model: task.classifierModel || task.classifier_model || '',
    rewrite_model: task.rewriteModel || task.rewrite_model || '',
    ai_last_attempt_count: Number(task.aiLastAttemptCount ?? task.ai_last_attempt_count) || 0,
    ai_last_attempt_at: task.aiLastAttemptAt || task.ai_last_attempt_at || '',
    ai_last_attempt_error: task.aiLastAttemptError || task.ai_last_attempt_error || '',
    sensitive_hit_count: Number(task.sensitiveHitCount ?? task.sensitive_hit_count) || 0,
    sensitive_fixed_count: Number(task.sensitiveFixedCount ?? task.sensitive_fixed_count) || 0,
    sensitive_failed_count: Number(task.sensitiveFailedCount ?? task.sensitive_failed_count) || 0,
    sensitive_mode: task.sensitiveMode || task.sensitive_mode || '',
    sensitive_status: task.sensitiveStatus || task.sensitive_status || '',
    sensitive_model: task.sensitiveModel || task.sensitive_model || '',
    site_submit_status: task.siteSubmitStatus || task.site_submit_status || '',
    site_submit_error: task.siteSubmitError || task.site_submit_error || '',
    site_submit_done_versions: Array.isArray(task.siteSubmitDoneVersions) ? task.siteSubmitDoneVersions : (Array.isArray(task.site_submit_done_versions) ? task.site_submit_done_versions : []),
    site_submit_accepted_versions: Array.isArray(task.siteSubmitAcceptedVersions) ? task.siteSubmitAcceptedVersions : (Array.isArray(task.site_submit_accepted_versions) ? task.site_submit_accepted_versions : []),
    site_submit_failed_versions: Array.isArray(task.siteSubmitFailedVersions) ? task.siteSubmitFailedVersions : (Array.isArray(task.site_submit_failed_versions) ? task.site_submit_failed_versions : []),
    created_at: task.createdAt || task.created_at || '',
    updated_at: task.updatedAt || task.updated_at || '',
    error: task.error || ''
  };
}

function legacyMeta(meta = {}) { return snakeTask(meta); }
function object(value) { return value && typeof value === 'object' && !Array.isArray(value) ? value : {}; }
function withSensitiveAiEnabled(config, enabled) {
  if (typeof enabled !== 'boolean') return config;
  return { ...config, sensitive_ai: { ...object(config?.sensitive_ai), enabled } };
}
function idFrom(value) { return String(value || '').trim(); }
function countEntries(value) { return Array.isArray(value) ? value.length : 0; }

// The legacy workbench still calls this router directly.  Keep its retry
// semantics aligned with the stage-aware runner: a failed AI stage must not
// fetch the already-successful original again.
function retryPlanForTask(meta = {}) {
  const stage = retryStageForTask(meta);
  return {
    stage,
    fetchOriginal: stage === 'original',
    applyOriginalRules: stage === 'original' || stage === 'sensitive',
    generateAi: stage === 'original' || stage === 'rewrite'
  };
}

function normalizeUploadProfiles(value) {
  return (Array.isArray(value) ? value : []).map((item, index) => {
    const source = object(item);
    const name = String(source.name || `配置档 ${index + 1}`).trim();
    return {
      id: String(source.id || `profile_${index + 1}`).trim(), name, enabled: source.enabled !== false,
      platform_id: String(source.platform_id || source.platformId || '').trim(),
      gender: String(source.gender || '').replace('频', '').trim(), style: String(source.style || '').trim(),
      is_default: source.is_default === true,
      config_id: String(source.config_id || '').trim(), source: String(source.source || '').trim(),
      // 配置档只保存用户明确填写的覆盖项；不能在这里补默认值，
      // 否则“仅改速度”的配置档会意外覆盖全局的字体、关键词等设置。
      advanced: { ...object(source.advanced || source.config_data) }
    };
  }).filter(item => item.id && item.name);
}

function normalizeProfileBindings(value) {
  return versionSelection.normalizeProfileBindings(value);
}

function selectBoundUploadProfile(cfg, version) {
  const profiles = normalizeUploadProfiles(cfg.upload_profiles).filter(item => item.enabled);
  const findProfile = (identity) => {
    const value = String(identity || '').trim();
    return profiles.find(item => item.id === value || item.name === value) || null;
  };
  const bound = findProfile(normalizeProfileBindings(cfg.profile_bindings)[version]);
  if (bound) return bound;
  const selected = findProfile(cfg.selected_profile);
  if (selected) return selected;
  return null;
}

function selectVersionUploadProfile(cfg, meta, version) {
  const profiles = normalizeUploadProfiles(cfg.upload_profiles).filter(item => item.enabled);
  const chosen = selectBoundUploadProfile(cfg, version);
  if (chosen) return chosen;
  const gender = String(meta.gender || '').replace('频', '').trim();
  const platformId = String(meta.platformId || '').trim();
  const style = String(meta.style || '').trim();
  const exact = profiles.find(item => item.platform_id === platformId && item.gender === gender && item.style === style);
  if (exact) return exact;
  const platformGender = profiles.find(item => item.platform_id === platformId && item.gender === gender);
  if (platformGender) return platformGender;
  return profiles.find(item => item.is_default) || profiles[0] || null;
}

function normalizeStyleCatalog(value) {
  const seen = new Set();
  return (Array.isArray(value) ? value : []).map(item => ({
    id: String(object(item).id ?? object(item).value ?? '').trim(),
    name: String(object(item).name ?? object(item).label ?? '').trim()
  })).filter(item => item.id && item.name && !seen.has(`${item.id}:${item.name}`) && (seen.add(`${item.id}:${item.name}`) || true));
}

function styleMapFromCatalog(value) {
  return Object.fromEntries(normalizeStyleCatalog(value).map(item => [item.name, Number(item.id)]).filter(([, id]) => Number.isFinite(id)));
}

function htmlText(value) {
  return String(value || '').replace(/<[^>]+>/g, '').replace(/&nbsp;/gi, ' ').replace(/&amp;/gi, '&').trim();
}

function parseTargetStyleCatalog(html) {
  const select = String(html || '').match(/<select\b(?=[^>]*\bid=["']style["'])[^>]*>([\s\S]*?)<\/select>/i);
  if (!select) throw new Error('121 自定义文案页未找到风格下拉框');
  const catalog = [];
  const pattern = /<option\b[^>]*\bvalue=["']([^"']*)["'][^>]*>([\s\S]*?)<\/option>/gi;
  for (let match; (match = pattern.exec(select[1]));) {
    const id = String(match[1] || '').trim();
    const name = htmlText(match[2]);
    if (id && name && !/请选择|全部/.test(name)) catalog.push({ id, name });
  }
  return normalizeStyleCatalog(catalog);
}

function parseTargetConfigData(value) {
  if (value && typeof value === 'object') return value;
  try { return JSON.parse(String(value || '{}')); } catch (_) { return {}; }
}

function importedAdvancedFromTarget(raw) {
  const source = object(raw);
  const jieya = object(source.jieya);
  const gunping = object(source.gunping);
  return target.normalizeAdvanced({
    tl5: Number(source.tl5), ziti: source.ziti ?? gunping.ziti, zitidx: source.zitidx ?? gunping.zitidx,
    biaohong: source.biaohong ?? gunping.biaohong, biaohongReuse: source.biaohong_reuse ?? gunping.biaohong_reuse,
    keywords: source.keywords, jieyaNum: source.jieya_num ?? jieya.jieya_num,
    jieyaAiHead: source.jieya_ai_head ?? jieya.jieya_ai_head ?? jieya.ai_head,
    jieyaSpeed: source.jieya_speed ?? jieya.jieya_speed ?? jieya.speed,
    jieyaPitch: source.jieya_pitch ?? jieya.jieya_pitch ?? jieya.pitch,
    gunpingNum: source.gunping_num ?? gunping.gunping_num,
    gunpingSpeed: source.gunping_speed ?? gunping.gunping_speed ?? gunping.speed,
    fontColorStyles: source.font_color_styles ?? jieya.font_color_styles
  });
}

function readTargetField(data, names) {
  const sources = [object(data), object(data?.data), object(data?.result)];
  for (const source of sources) {
    for (const name of names) {
      const value = source[name];
      if (value !== undefined && value !== null && String(value).trim()) return String(value).trim();
    }
  }
  return '';
}

function summarizeTargetReceipt(data) {
  const remoteId = readTargetField(data, ['task_id', 'taskId', 'job_id', 'jobId', 'queue_id', 'queueId', 'execution_id', 'executionId']);
  const remoteStatus = readTargetField(data, ['task_status', 'taskStatus', 'job_status', 'jobStatus', 'queue_status', 'queueStatus', 'status']).toLowerCase();
  const uploadResult = object(data?.result);
  const successResult = object(uploadResult.success);
  const failedResult = object(uploadResult.failed);
  const successFiles = Array.isArray(successResult.files) ? successResult.files : [];
  const failedFiles = Array.isArray(failedResult.files) ? failedResult.files : [];
  const successCount = Math.max(0, Number(successResult.count) || successFiles.length);
  const failedCount = Math.max(0, Number(failedResult.count) || failedFiles.length);
  const failedReasons = failedFiles.map(item => String(object(item).reason || object(item).message || '')).filter(Boolean);
  const uploadError = failedCount > 0
    ? `121 上传失败：${failedReasons.join('；') || `有 ${failedCount} 个文件被拒绝`}`
    : (Object.keys(uploadResult).length > 0 && successCount < 1 ? '121 未确认接收到上传文件' : '');
  const verified = !uploadError && (successCount > 0 || Boolean(remoteId) || ['queued', 'queueing', 'running', 'executing', 'created'].includes(remoteStatus));
  let raw = '';
  try { raw = JSON.stringify(data); } catch (_) { raw = String(data || ''); }
  return {
    remote_id: remoteId,
    remote_status: remoteStatus,
    verified,
    upload_success_count: successCount,
    upload_failed_count: failedCount,
    upload_error: uploadError,
    response: raw.slice(0, 4000),
    message: readTargetField(data, ['message', 'msg'])
  };
}

function appendSubmitTrace(trace, step, status, detail, extra = {}) {
  trace.push({ step, status, detail, time: new Date().toISOString(), ...extra });
}

function parseTargetJson(value) {
  if (!value || typeof value !== 'string') return {};
  try { return JSON.parse(value); } catch (_) { return {}; }
}

function summarizeRemoteBookRecord(data, bookId, fields, advanced) {
  const rows = Array.isArray(data?.data) ? data.data : [];
  const record = rows.find(item => String(object(item).bookid || '') === String(bookId)) || null;
  if (!record) return { status: '待确认', found: false, detail: '121 上传接口已确认文件，但列表暂未返回该书号记录' };
  const row = object(record);
  const jianData = parseTargetJson(row.jian_data);
  const actual = {
    platform_id: String(row.book_platform ?? ''), gender: String(row.gender ?? ''), style: String(row.style ?? ''),
    jieya_num: Number(jianData?.jieya?.jieya_num ?? -1), gunping_num: Number(jianData?.gunping?.gunping_num ?? -1)
  };
  const expected = {
    platform_id: String(fields.platform_id), gender: String(fields.gender), style: String(fields.style),
    jieya_num: Number(advanced.jieyaNum), gunping_num: Number(advanced.gunpingNum)
  };
  const mismatches = Object.keys(expected).filter(key => expected[key] !== actual[key]);
  return {
    status: mismatches.length ? '待确认' : '完成', found: true,
    detail: mismatches.length ? `121 已找到记录，但参数待核对：${mismatches.join('、')}` : '121 后台已找到对应书号，平台、风格与素材数量一致',
    remote_id: String(row.id || ''), remote_time: String(row.addtime || ''), expected, actual, mismatches
  };
}

async function verifyTargetBookRecord(httpClient, sessionCookie, bookId, fields, advanced) {
  const response = await httpClient({ method: 'GET', url: target.buildTargetBookListUrl(bookId), headers: { Cookie: sessionCookie } });
  if (target.isLoginPage(response.body)) throw new Error('121 登录会话已失效，无法核验后台记录');
  let data = {};
  try { data = JSON.parse(response.body); } catch (_) { throw new Error('121 列表核验返回了非 JSON 内容'); }
  if (data.success !== true) throw new Error(data.message || data.msg || '121 列表核验失败');
  return summarizeRemoteBookRecord(data, bookId, fields, advanced);
}

function summarizeKnowledge(knowledge) {
  const source = object(knowledge);
  const highImitation = object(source.high_imitation);
  const openingPhrases = object(source.opening_phrases);
  const rewriteTemplates = object(source.rewrite_templates);
  const layoutRules = object(source.layout_rules);
  const symbolRules = object(source.symbol_rules);
  const chapterRules = object(source.chapter_rules);
  const openingStyles = countEntries(openingPhrases.styles)
    || new Set((openingPhrases.items || []).map(item => String(item?.style || item?.category || '').trim()).filter(Boolean)).size;

  return {
    opening_phrases: countEntries(openingPhrases.items),
    opening_styles: openingStyles,
    rewrite_profiles: countEntries(rewriteTemplates.profiles) || countEntries(rewriteTemplates.items),
    temporary_instructions: countEntries(rewriteTemplates.temporary_instructions),
    high_imitation_prompts: countEntries(highImitation.prompts),
    high_imitation_references: countEntries(highImitation.references) || countEntries(highImitation.items),
    sensitive_rules: countEntries(layoutRules.sensitive_rules) || countEntries(layoutRules.sensitiveKeywords),
    symbol_rules: countEntries(symbolRules.symbol_rules),
    pair_fill_rules: countEntries(symbolRules.pair_fill_rules),
    chapter_exact_rules: countEntries(chapterRules.chapter_exact_rules),
    chapter_inline_rules: countEntries(chapterRules.chapter_inline_rules)
  };
}

function isFailed(task) {
  return [task.status, task.originalStatus, task.aiStatus, task.classifyStatus, task.error]
    .join(' ').toLowerCase().includes('failed');
}

function createBatchRewriteRouter({
  auth = apiAuth,
  targetBaseUrl,
  bridgeSecret,
  systemDir,
  novelFetchStore,
  resolveRuntimeModel,
  tasksFactory,
  configStore,
  knowledgeStore,
  openingStore,
  httpClient = target.requestHttp
} = {}) {
  const router = express.Router();
  router.use(auth);
  const knowledge = knowledgeStore || getKnowledgeStore(systemDir);
  const opening = openingStore || createOpeningStore({ systemDir, styles: [] });
  const configCache = new Map();

  function configCacheKey(req) { return String(req.username || req.auth?.username || 'anonymous'); }
  function invalidateConfig(req) { configCache.delete(configCacheKey(req)); }
  async function getCachedConfig(req, tasks) {
    const key = configCacheKey(req);
    const now = Date.now();
    const cached = configCache.get(key);
    if (cached?.value && cached.expiresAt > now) return cached.value;
    if (cached?.promise) return cached.promise;
    const promise = Promise.resolve(tasks.getConfig()).then(value => {
      configCache.set(key, { value, expiresAt: Date.now() + CONFIG_CACHE_TTL_MS });
      return value;
    }).finally(() => {
      const current = configCache.get(key);
      if (current?.promise === promise) configCache.delete(key);
    });
    configCache.set(key, { promise });
    return promise;
  }

  async function resources(req) {
    if (tasksFactory) return tasksFactory(req);
    const tasks = createMySQLWorkshopStore({ targetBaseUrl, bridgeSecret, account: req.auth?.account });
    const savedConfig = await getCachedConfig(req, tasks);
    const savedKnowledge = object(savedConfig.knowledge);
    const fallbackKnowledge = {};
    if (!Object.keys(savedKnowledge).length) {
      for (const kind of LEGACY_KINDS) fallbackKnowledge[kind] = knowledge.list(kind);
    }
    const config = Object.keys(savedKnowledge).length ? savedConfig : { ...savedConfig, knowledge: fallbackKnowledge };
    return {
      tasks,
      config,
      configStore: configStore || {
        getConfig: () => config,
        getAiConfig: () => ({
          ai: config.ai || {},
          ai_presets: config.ai_presets || [],
          ai_assignments: config.ai_assignments || {},
          text_model_id: config.text_model_id || config.textModelId
            || config.app_config?.text_model_id || config.app_config?.textModelId || ''
        }),
        resolveRuntimeModel: modelId => {
          if (typeof resolveRuntimeModel !== 'function') return null;
          return resolveRuntimeModel(req.username, 'text', modelId);
        },
        getPlatforms: () => config.platforms || PLATFORMS,
        getStyles: () => config.styles || STYLE_NAMES
      }
    };
  }

  async function readConfig(req, { lightweight = false } = {}) {
    const { tasks, config, configStore: store } = await resources(req);
    const current = object(config || store.getConfig());
    const platforms = Array.isArray(current.platforms) ? current.platforms : (store.getPlatforms?.() || PLATFORMS);
    const styles = Array.isArray(current.styles) ? current.styles : (store.getStyles?.() || STYLE_NAMES);
    const savedKnowledge = object(current.knowledge);
    let knowledgeData = lightweight ? {} : savedKnowledge;
    if (!lightweight) {
      const legacyKnowledge = {};
      for (const kind of LEGACY_KINDS) legacyKnowledge[kind] = knowledge.list(kind);
      knowledgeData = mergeKnowledgeSources(knowledgeData, legacyKnowledge);
      if (!Object.keys(knowledgeData).length) knowledgeData = legacyKnowledge;
      const high = object(knowledgeData.high_imitation);
      if (!countEntries(high.references) && countEntries(high.items)) {
        knowledgeData = {
          ...knowledgeData,
          high_imitation: {
            ...high,
            references: high.items.map((item, index) => {
              const source = object(item);
              return {
                ...source,
                id: source.id || `high_${String(index + 1).padStart(3, '0')}`,
                name: source.name || source.title || '',
                reference_text: source.reference_text || source.content || '',
                enabled: source.enabled !== false
              };
            })
          }
        };
      }
    }
    const appConfig = {
      ...current,
      workflow: {
        auto_classify_missing: current.workflow?.auto_classify_missing !== false,
        auto_fetch_original: current.workflow?.auto_fetch_original !== false,
        auto_rewrite_after_fetch: current.workflow?.auto_rewrite_after_fetch !== false,
        auto_submit_after_rewrite: current.workflow?.auto_submit_after_rewrite === true,
        ...object(current.workflow)
      }
    };
    const { process_jobs: _processJobs, site_submit_groups: _siteSubmitGroups, ...publicAppConfig } = appConfig;
    return {
      tasks,
      store,
      current: appConfig,
      response: {
        app_config: publicAppConfig,
        platforms,
        styles,
        sensitive: current.sensitive || { groups: [] },
        web_submit: publicWebSubmit(current.web_submit),
        knowledge: knowledgeData,
        knowledge_loaded: !lightweight,
        knowledge_summary: lightweight ? {} : summarizeKnowledge(knowledgeData),
        parse_modes: parse.PARSE_MODES.map(id => ({ id, name: id })),
        column_presets: parse.COLUMN_PRESETS,
        work_form: current.work_form || {}
      }
    };
  }

  function normalizeProcessJobs(value) {
    return (Array.isArray(value) ? value : []).filter(item => object(item).id).slice(-PROCESS_JOB_LIMIT);
  }

  async function persistProcessJob(req, job) {
    const { tasks, current } = await readConfig(req);
    const entries = normalizeProcessJobs(current.process_jobs);
    const next = { ...job, updated_at: new Date().toISOString() };
    const index = entries.findIndex(item => item.id === next.id);
    if (index >= 0) entries[index] = next;
    else entries.push(next);
    if (typeof tasks.saveConfig === 'function') await tasks.saveConfig({ ...current, process_jobs: entries.slice(-PROCESS_JOB_LIMIT) });
    return next;
  }

  async function findProcessJob(req, id) {
    const key = `${req.username}:${id}`;
    const active = jobs.get(key);
    if (active) return active;
    const { current, tasks } = await readConfig(req);
    const entries = normalizeProcessJobs(current.process_jobs);
    const stored = entries.find(item => item.id === id);
    if (!stored) return null;
    // 内存中不存在仍标记 running 的任务，说明服务已重启，不能继续对用户谎称它仍在执行。
    if (stored.status === 'running') {
      const interrupted = { ...stored, status: 'interrupted', error: '服务已重启，未完成的处理任务已中断；请重新执行该批。', completed_at: new Date().toISOString() };
      const index = entries.findIndex(item => item.id === id);
      entries[index] = interrupted;
      if (typeof tasks.saveConfig === 'function') await tasks.saveConfig({ ...current, process_jobs: entries });
      return interrupted;
    }
    return stored;
  }

  async function latestProcessJobs(req, limit = 1) {
    const { current, tasks } = await readConfig(req);
    const entries = normalizeProcessJobs(current.process_jobs);
    let changed = false;
    const repaired = entries.map(item => {
      if (item.status !== 'running' || jobs.has(`${req.username}:${item.id}`)) return item;
      changed = true;
      return { ...item, status: 'interrupted', error: '服务已重启，未完成的处理任务已中断；请重新执行该批。', completed_at: new Date().toISOString(), updated_at: new Date().toISOString() };
    });
    if (changed && typeof tasks.saveConfig === 'function') await tasks.saveConfig({ ...current, process_jobs: repaired });
    return repaired.slice().sort((left, right) => String(right.updated_at || right.started_at || '').localeCompare(String(left.updated_at || left.started_at || ''))).slice(0, Math.max(1, Math.min(Number(limit) || 1, PROCESS_JOB_LIMIT)));
  }

  async function persistSiteSubmitGroup(req, group) {
    const { tasks, current } = await readConfig(req);
    const entries = (Array.isArray(current.site_submit_groups) ? current.site_submit_groups : []).filter(item => object(item).group_id).slice(-199);
    const next = { ...group, updated_at: new Date().toISOString() };
    const index = entries.findIndex(item => item.group_id === next.group_id);
    if (index >= 0) entries[index] = next;
    else entries.push(next);
    if (typeof tasks.saveConfig === 'function') await tasks.saveConfig({ ...current, site_submit_groups: entries.slice(-200) });
    return next;
  }

  async function listPersistedSiteSubmitGroups(req) {
    const { current } = await readConfig(req);
    return (Array.isArray(current.site_submit_groups) ? current.site_submit_groups : [])
      .filter(item => object(item).group_id)
      .sort((left, right) => String(right.updated_at || right.started_at || '').localeCompare(String(left.updated_at || left.started_at || '')))
      .slice(0, 50);
  }

  async function requestKnowledge(req) {
    return (await readConfig(req)).response.knowledge;
  }

  function ruleKnowledge(knowledgeData) {
    return {
      layout_rules: object(knowledgeData.layout_rules),
      symbol_rules: object(knowledgeData.symbol_rules),
      chapter_rules: object(knowledgeData.chapter_rules)
    };
  }

  function processSavedRules(text, scope, config) {
    return rules.processConfiguredDocumentText(text, scope, config);
  }

  async function applySavedRulesToVersion(tasks, username, bookId, version, config) {
    // 规则入口与真实存储分开，以便兼容只负责抓取的测试/桥接实现。
    const isOriginal = version === 'original';
    const read = isOriginal ? tasks.readOriginal : tasks.readVersionText;
    const save = isOriginal ? tasks.saveOriginalText : tasks.saveVersionText;
    if (typeof read !== 'function' || typeof save !== 'function') return false;
    const text = isOriginal ? await read.call(tasks, username, bookId) : await read.call(tasks, username, bookId, version);
    if (!text) return false;
    const processed = processSavedRules(text, isOriginal ? 'original' : 'ai', config);
    if (processed === text) return false;
    if (isOriginal) await save.call(tasks, username, bookId, processed);
    else await save.call(tasks, username, bookId, version, processed);
    return true;
  }

  function sensitiveLogEntry(result, settings, aiSettings) {
    const fixedItems = Array.isArray(result?.fixedItems) ? result.fixedItems : [];
    const failedCount = fixedItems.filter(item => item?.status === 'failed').length;
    return {
      mode: settings.enabled ? 'ai_fix' : 'replace',
      status: failedCount ? 'partial' : (result?.hits?.length ? 'done' : 'skipped'),
      hit_count: Array.isArray(result?.hits) ? result.hits.length : 0,
      fixed_count: Number(result?.fixedCount) || 0,
      failed_count: failedCount,
      model: settings.enabled ? (aiSettings.model || '') : '',
      updated_at: new Date().toISOString(),
      hits: Array.isArray(result?.hits) ? result.hits : [],
      items: fixedItems
    };
  }

  async function applySavedRulesToOriginal(tasks, username, bookId, config, configStore) {
    if (typeof tasks.readOriginal !== 'function' || typeof tasks.saveOriginalText !== 'function') return false;
    const text = await tasks.readOriginal(username, bookId);
    if (!text) return false;
    const layoutRules = object(object(config.knowledge).layout_rules);
    const runSensitive = config.layout?.apply_sensitive !== false && layoutRules.apply_sensitive_replace !== false;
    // 原文敏感词只能由这一处统一处理，避免规则引擎先静态替换后 AI 已经失去命中上下文。
    const ruleConfig = runSensitive ? {
      ...config,
      layout: { ...object(config.layout), sensitive_ai_enabled: true }
    } : config;
    let processed = processSavedRules(text, 'original', ruleConfig);
    let entry = null;
    if (runSensitive) {
      const task = typeof tasks.getTask === 'function' ? await tasks.getTask(username, bookId) : null;
      const sensitiveConfig = object(config.sensitive_ai);
      const aiSource = configStore && typeof configStore.getAiConfig === 'function' ? configStore : {
        getAiConfig: () => ({ ai: object(config.ai), ai_presets: Array.isArray(config.ai_presets) ? config.ai_presets : [], ai_assignments: object(config.ai_assignments) })
      };
      const aiSettings = sensitiveConfig.enabled === true ? ai.resolveAiSettings(aiSource, 'sensitive_fix') : {};
      const result = await sensitive.processSensitiveText({
        text: processed,
        settings: {
          ...sensitiveConfig,
          enabled: sensitiveConfig.enabled === true,
          config,
          scope: 'original',
          bookId,
          bookName: task?.meta?.bookName || '',
          style: task?.meta?.style || '',
          gender: task?.meta?.gender || '',
          aiSettings
        },
        ai
      });
      processed = result.text;
      entry = sensitiveLogEntry(result, { enabled: sensitiveConfig.enabled === true }, aiSettings);
      if (typeof tasks.updateTaskMeta === 'function') await tasks.updateTaskMeta(username, bookId, {
        sensitiveMode: entry.mode,
        sensitiveStatus: entry.status,
        sensitiveHitCount: entry.hit_count,
        sensitiveFixedCount: entry.fixed_count,
        sensitiveFailedCount: entry.failed_count,
        sensitiveModel: entry.model
      });
      if (typeof tasks.appendLog === 'function') await tasks.appendLog(username, bookId, 'sensitive_processed', entry);
    }
    if (processed === text) return Boolean(entry);
    await tasks.saveOriginalText(username, bookId, processed);
    return true;
  }

  async function saveConfig(req, payload) {
    const { tasks, current } = await readConfig(req);
    const body = object(payload);
    const next = {
      ...current,
      ...(object(body.app_config)),
      ...(Array.isArray(body.platforms) ? { platforms: body.platforms } : {}),
      ...(Array.isArray(body.styles) ? { styles: body.styles } : {}),
      ...(body.sensitive ? { sensitive: body.sensitive } : {}),
      ...(body.knowledge ? { knowledge: body.knowledge } : {})
    };
    await tasks.saveConfig(next);
    invalidateConfig(req);
    return readConfig(req);
  }

  async function listTasks(req) {
    const { tasks } = await resources(req);
    const records = await tasks.listTasks(req.username);
    void scheduleHistoricalAiStatusRepair({ tasks, username: req.username, records });
    return records.map(snakeTask);
  }

  async function listSiteSubmitHistory(req) {
    const { tasks } = await resources(req);
    const records = await listTasks(req);
    const entries = [];
    for (const record of records) {
      if (typeof tasks.readSiteSubmitLog !== 'function') continue;
      const logs = await tasks.readSiteSubmitLog(req.username, record.id);
      for (const item of logs) {
        const receipt = object(item?.remote_receipt);
        const remoteRecord = object(receipt.remote_record);
        const result = item?.error
          || receipt.upload_error
          || (item?.status === 'submitted' && receipt.verified
            ? `121 已确认接收 ${Number(receipt.upload_success_count) || 0} 个文件${remoteRecord.detail ? `；${remoteRecord.detail}` : ''}`
            : item?.status === 'accepted_pending'
              ? (remoteRecord.detail || '121 已接收，等待后台确认')
              : '');
        entries.push({
          book_id: record.id,
          book_name: record.book_name || '',
          platform_name: record.platform_name || '',
          version: item?.version || '',
          status: item?.status || '',
          group_id: item?.group_id || '',
          material_allocation: item?.material_allocation || null,
          result,
          error: item?.error || receipt.upload_error || '',
          time: item?.time || item?.updated_at || ''
        });
      }
    }
    return entries.sort((left, right) => String(right.time).localeCompare(String(left.time))).slice(0, 200);
  }

  function taskIssueEntry(task, item, index) {
    const data = object(item?.data);
    const event = String(item?.event || 'task_error');
    const status = String(data.status || '');
    const failedItem = (Array.isArray(data.items) ? data.items : []).find(entry => entry?.status === 'failed');
    const detail = String(data.error || data.upload_error || failedItem?.error || task.error || task.ai_error || '').trim();
    const failed = /(?:failed|error|失败|错误)/i.test(`${event} ${status}`)
      || Number(data.failed_count) > 0
      || status === 'partial';
    if (!failed) return null;
    const kind = event === 'site_submit'
      ? 'batch-rewrite.submit'
      : event.includes('sensitive')
        ? 'batch-rewrite.sensitive'
        : event.includes('ai')
          ? 'batch-rewrite.ai'
          : 'batch-rewrite.processing';
    const message = detail || `书籍 ${task.id} 的${event === 'site_submit' ? '网站提交' : event.includes('sensitive') ? '敏感词处理' : event.includes('ai') ? 'AI 文案处理' : '原文处理'}失败`;
    return {
      id: `batch-rewrite:${task.id}:${event}:${item?.time || index}`,
      kind,
      message: `书籍 ${task.id}${task.book_name ? `《${task.book_name}》` : ''}：${message}`,
      at: item?.time || task.updated_at || task.updatedAt || new Date(0).toISOString(),
      path: '/novel-fetch',
      method: event,
      status: /(?:未登录|会话|认证)/.test(message) ? 401 : undefined
    };
  }

  async function listWorkshopIssues(req) {
    const { tasks } = await resources(req);
    const entries = [];
    for (const task of await listTasks(req)) {
      if (task.error || task.ai_error || task.aiError) {
        entries.push(taskIssueEntry(task, { event: task.ai_error || task.aiError ? 'ai_generate_failed' : 'task_failed', time: task.updated_at || task.updatedAt, data: { error: task.ai_error || task.aiError || task.error, status: 'failed' } }, 0));
      }
      for (const [index, item] of (await tasks.readLogs(req.username, task.id)).entries()) {
        const entry = taskIssueEntry(task, item, index);
        if (entry) entries.push(entry);
      }
    }
    return entries.filter(Boolean).sort((left, right) => String(right.at).localeCompare(String(left.at))).slice(0, 200);
  }

  async function readSensitiveLog(tasks, username, bookId) {
    const logs = typeof tasks.readLogs === 'function' ? await tasks.readLogs(username, bookId) : [];
    const record = [...logs].reverse().find(item => item?.event === 'sensitive_processed')?.data || {};
    const relatedLogs = logs.filter(item => ['sensitive_processed', 'original_fetched', 'original_rules_applied', 'original_restored'].includes(item?.event));
    return {
      sensitive_hits: {
        mode: record.mode || '', status: record.status || '', hit_count: Number(record.hit_count) || 0,
        updated_at: record.updated_at || '', hits: Array.isArray(record.hits) ? record.hits : []
      },
      sensitive_fixed: {
        mode: record.mode || '', status: record.status || '', fixed_count: Number(record.fixed_count) || 0,
        failed_count: Number(record.failed_count) || 0, model: record.model || '', updated_at: record.updated_at || '',
        items: Array.isArray(record.items) ? record.items : []
      },
      logs: relatedLogs
    };
  }

  async function selectTaskIds(req, body) {
    const requested = Array.isArray(body?.ids) ? body.ids.map(idFrom).filter(Boolean) : [];
    if (body?.mode === 'selected') return requested;
    const all = await listTasks(req);
    if (body?.mode === 'failed') return all.filter(isFailed).map(task => task.id);
    return all.map(task => task.id);
  }

  async function processPayload(req, payload, report = () => {}) {
    const { tasks, configStore: store } = await resources(req);
    const inputText = String(payload?.input_text || '');
    if (!inputText.trim()) throw new Error('请输入书籍信息');
    const config = withSensitiveAiEnabled(object(store.getConfig()), payload?.sensitive_ai_enabled);
    const parsed = parse.parseBooks({
      inputText,
      parseMode: 'smart',
      columnPresetId: '',
      columnOrder: '',
      styles: store.getStyles?.() || []
    });
    const platforms = store.getPlatforms?.() || [];
    const platform = platforms.find(item => String(item.id) === String(payload.platform_id || '2')) || {};
    const workflow = object(config.workflow);
    const hasExplicitVersions = Array.isArray(payload?.selected_versions);
    const selectedVersions = versionSelection.normalizeSelectedVersions(
      payload?.selected_versions,
      hasExplicitVersions ? [] : undefined
    );
    if (!selectedVersions.length) throw new Error('请至少选择一个文案版本');
    const aiSlotMethods = versionSelection.normalizeAiSlotMethods(payload.ai_slot_methods);
    const prepared = (parsed.tasks || []).map(item => ({
      ...item,
      platformId: String(payload.platform_id || '2'),
      platformName: platform.name || String(payload.platform_id || '2'),
      maxTxt: Number(payload.max_txt) || Number(config.fetch?.default_max_txt) || 4000,
      selectedVersions,
      aiSlotMethods
    }));
    // Keep the complete task store for history, but identify this input batch
    // explicitly so the workbench can show only the books just submitted.
    const currentBatchIds = prepared.map(item => String(item.bookId || '')).filter(Boolean);
    report({ type: 'version_config', status: 'done', message: `版本配置已读取：${selectedVersions.map(version => version.toUpperCase()).join('、')}；有效 ${prepared.length} 个任务。` });
    // 先把轻量任务落盘，再抓原文。抓原文返回的 bookinfo 里可能已经有
    // 男女频/分类信息；在正文还没拿到之前就调用 AI，会把本来能直接判断的
    // 任务误报成“AI 判断失败”。
    let classified = prepared;
    let classifyErrors = [];
    await tasks.saveTasks(req.username, prepared);
    let fetched = 0;
    let fetchFailed = 0;
    const fetchedIds = [];
    if (workflow.auto_fetch_original !== false) {
      const fetchedResults = await runWithConcurrency(classified, config.fetch?.concurrency || 1, async task => {
        const result = await tasks.fetchOriginal(req.username, task.bookId, task.maxTxt);
        if (result?.status !== 'done') return { task, done: false };
        await applySavedRulesToOriginal(tasks, req.username, task.bookId, config, store);
        return { task, done: true };
      });
      for (const [index, item] of fetchedResults.entries()) {
        const task = item?.value?.task || classified[index];
        if (item?.ok && item.value?.done) {
          fetched++;
          fetchedIds.push(task.bookId);
          report({ type: 'fetch', status: 'done', book_id: task.bookId, message: `${task.bookId} 原文已抓取并套用处理规则。` });
        } else {
          fetchFailed++;
          report({ type: 'fetch', status: 'failed', book_id: task?.bookId || '', message: `${task?.bookId || '任务'} 原文抓取失败。` });
        }
      }
    }

    // 原文抓取完成后再做分类：优先使用抓取接口返回的 bookinfo，只有仍缺
    // 男女频或风格时，classifier 才会调用文本模型补齐。
    if (workflow.auto_classify_missing !== false && prepared.length) {
      const refreshed = [];
      for (const task of prepared) {
        const current = await tasks.getTask(req.username, task.bookId);
        const meta = current?.meta || current || task;
        const info = meta?.bookinfo && typeof meta.bookinfo === 'object' ? meta.bookinfo : {};
        const category = String(info.category || info.category_name || '').trim();
        const gender = !meta.gender && /女频|女生|女向/i.test(category)
          ? '女'
          : (!meta.gender && /男频|男生|男向/i.test(category) ? '男' : '');
        const styleCandidates = [info.style, info.style_name, info.genre, info.genre_name, category]
          .map(value => String(value || '').replace(/^(男频|女频)[\s:_-]*/i, '').trim())
          .filter(Boolean);
        const styles = Array.isArray(store.getStyles?.()) ? store.getStyles() : [];
        const style = !meta.style
          ? styles.find(candidate => styleCandidates.some(value => value === candidate || value.includes(candidate) || candidate.includes(value))) || ''
          : '';
        refreshed.push({
          ...task,
          ...meta,
          ...(gender ? { gender, genderSource: 'bookinfo' } : {}),
          ...(style ? { style, styleSource: 'bookinfo' } : {}),
          bookinfo: info
        });
      }
      const result = await classifier.classifyMissingRows({ configStore: store, tasks: refreshed });
      classified = result.tasks || refreshed;
      classifyErrors = result.errors || [];
      await tasks.saveTasks(req.username, classified);
      report({ type: 'classify', status: classifyErrors.length ? 'warning' : 'done', message: classifyErrors.length ? classifyErrors.join('；') : '原文已获取，男女频和风格补齐完成。' });
    } else {
      report({ type: 'classify', status: 'skipped', message: '未开启自动分类，保留原文后再手动判断。' });
    }
    let generated = 0;
    let rewriteFailed = 0;
    if (workflow.auto_rewrite_after_fetch) {
      const rewriteResults = await runWithConcurrency(fetchedIds, config.ai?.max_concurrency || 1, async bookId => {
        const task = await tasks.getTask(req.username, bookId);
        return rewrite.generateAiVersions({ configStore: store, tasks, username: req.username, task: task?.meta, versions: task?.meta?.selectedVersions, slotMethods: task?.meta?.aiSlotMethods });
      });
      for (const [index, item] of rewriteResults.entries()) {
        if (!item?.ok) {
          rewriteFailed++;
          report({ type: 'rewrite', status: 'failed', book_id: fetchedIds[index], message: `${fetchedIds[index]} AI 文案生成失败：${item.error?.message || '未知错误'}` });
          continue;
        }
        generated += (item.value?.generated || []).filter(entry => entry.status === 'done').length;
      }
      report({ type: 'rewrite', status: rewriteFailed ? 'warning' : 'done', message: `AI 文案生成 ${generated} 个版本${rewriteFailed ? `，失败 ${rewriteFailed} 本` : ''}。` });
    } else {
      report({ type: 'rewrite', status: 'skipped', message: '配置未开启抓取后自动改文。' });
    }
    if (workflow.auto_submit_after_rewrite === true && workflow.auto_submit_confirmed === true && config.web_submit?.enabled === true && fetchedIds.length) {
      const submitted = await submitTasks(req, { mode: 'selected', ids: fetchedIds });
      report({ type: 'submit', status: submitted.failed_groups ? 'warning' : 'done', message: `网站提交：成功 ${submitted.success_groups || 0} 组，失败 ${submitted.failed_groups || 0} 组。` });
    } else {
      report({ type: 'submit', status: 'skipped', message: '未开启自动网站提交；文案已保留，需在网站提交页确认后再提交。' });
    }
    return {
      parsed: parsed.parsed || prepared.length,
      unique_tasks: parsed.uniqueTasks || prepared.length,
      duplicate_count: parsed.duplicateCount || 0,
      empty_id_count: parsed.emptyIdCount || 0,
      classify_errors: classifyErrors,
      raw_fetched: fetched,
      fetched,
      process_failed: 0,
      rewrite_failed: rewriteFailed,
      fetch_failed: fetchFailed,
      generated_ai_files: generated,
      current_batch_ids: currentBatchIds,
      tasks: await listTasks(req)
    };
  }

  function publicWebSubmit(value = {}) {
    const cfg = object(value);
    return {
      enabled: cfg.enabled === true,
      username: cfg.username || '',
      password: '',
      password_masked: Boolean(cfg.password_masked),
      skip_submitted: cfg.skip_submitted !== false,
      min_text_chars: Math.max(0, Math.min(Number(cfg.min_text_chars) || 0, 100000)),
      retry_times: Math.max(0, Math.min(Number(cfg.retry_times) || 1, 5)),
      submit_mode: 'version',
      selected_profile: String(cfg.selected_profile || ''),
      advanced: target.normalizeBookAdvanced(cfg.advanced),
      upload_profiles: normalizeUploadProfiles(cfg.upload_profiles),
      profile_bindings: normalizeProfileBindings(cfg.profile_bindings),
      style_catalog: normalizeStyleCatalog(cfg.style_catalog),
      target: target.TARGET_HOST
    };
  }

  async function currentWebConfig(req) {
    const { current } = await readConfig(req);
    const cfg = object(current.web_submit);
    return {
      ...cfg,
      min_text_chars: Math.max(0, Math.min(Number(cfg.min_text_chars) || 0, 100000)),
      retry_times: Math.max(0, Math.min(Number(cfg.retry_times) || 1, 5)),
      submit_mode: 'version',
      selected_profile: String(cfg.selected_profile || ''),
      advanced: target.normalizeBookAdvanced(cfg.advanced),
      upload_profiles: normalizeUploadProfiles(cfg.upload_profiles),
      profile_bindings: normalizeProfileBindings(cfg.profile_bindings),
      style_catalog: normalizeStyleCatalog(cfg.style_catalog)
    };
  }

  async function saveWebConfig(req, incoming) {
    const { tasks, current } = await readConfig(req);
    const existing = object(current.web_submit);
    const received = object(incoming);
    const clean = {
      enabled: received.enabled === true,
      username: String(received.username || existing.username || '').trim(),
      password_masked: Boolean(existing.password_masked),
      skip_submitted: received.skip_submitted !== false,
      min_text_chars: Math.max(0, Math.min(Number(received.min_text_chars) || 0, 100000)),
      retry_times: Math.max(0, Math.min(Number(received.retry_times) || 1, 5)),
      submit_mode: 'version',
      selected_profile: String(received.selected_profile || ''),
      advanced: target.normalizeBookAdvanced(received.advanced),
      upload_profiles: normalizeUploadProfiles(received.upload_profiles),
      profile_bindings: normalizeProfileBindings(received.profile_bindings),
      style_catalog: normalizeStyleCatalog(received.style_catalog || existing.style_catalog)
    };
    if (typeof received.password === 'string') clean.password = received.password;
    const password = String(clean.password || '');
    delete clean.password;
    if (password && novelFetchStore) {
      // 目标站登录页先建立 PHP 会话，再向当前的 JSON 登录接口提交凭据。
      const loginPage = await httpClient({ method: 'GET', url: target.buildLoginPageUrl() });
      const initialCookie = target.cookieHeaderFromSetCookie(loginPage.headers);
      const login = await httpClient(target.buildLoginRequest(String(clean.username || ''), password, initialCookie));
      let data = {};
      try { data = JSON.parse(login.body); } catch (_) {}
      if (data.success !== true) throw new Error(data.message || '目标站登录失败');
      const cookie = target.mergeCookieHeaders(initialCookie, target.cookieHeaderFromSetCookie(login.headers));
      if (!cookie) throw new Error('目标站登录未返回会话');
      novelFetchStore.setSession(req.username, cookie);
      clean.password_masked = true;
    }
    await tasks.saveConfig({ ...current, web_submit: clean });
    return publicWebSubmit(clean);
  }

  async function request121(req, path) {
    const session = novelFetchStore?.getSession(req.username);
    if (!session?.cookie) throw new Error('请先保存账号密码并验证 121 登录会话');
    const response = await httpClient({ method: 'GET', url: `http://${target.TARGET_HOST}${path}`, headers: { Cookie: session.cookie } });
    if (target.isLoginPage(response.body)) throw new Error('121 登录会话已失效，请重新验证登录');
    return response.body;
  }

  async function sync121Profiles(req) {
    const body = await request121(req, target.TARGET_CONFIG_LIST_PATH);
    let payload = {};
    try { payload = JSON.parse(body); } catch (_) { throw new Error('121 配置档接口返回了非 JSON 数据'); }
    if (payload.success !== true || !Array.isArray(payload.data)) throw new Error(payload.message || payload.msg || '121 配置档同步失败');
    const styleCatalog = normalizeStyleCatalog((await currentWebConfig(req)).style_catalog);
    const styleById = Object.fromEntries([...styleCatalog, ...Object.entries(target.STYLE_ID).map(([name, id]) => ({ id: String(id), name }))].map(item => [String(item.id), item.name]));
    const profiles = payload.data.map((item, index) => {
      const source = object(item);
      const data = parseTargetConfigData(source.config_data);
      const gender = String(data.gender ?? source.gender ?? '');
      return {
        id: `121-${String(source.id || index + 1)}`,
        name: String(source.config_name || `121 配置档 ${index + 1}`).trim(),
        enabled: true, source: '121', config_id: String(source.id || ''), is_default: Number(source.is_default) === 1,
        platform_id: String(data.platform_id ?? source.platform_id ?? '').trim(),
        gender: gender === '1' ? '男' : gender === '2' ? '女' : gender.replace('频', ''),
        style: styleById[String(data.style ?? source.style ?? '')] || String(data.style_name || source.style_name || '').trim(),
        advanced: importedAdvancedFromTarget(data)
      };
    }).filter(profile => profile.name && profile.config_id);
    const { tasks, current } = await readConfig(req);
    const web = object(current.web_submit);
    const local = normalizeUploadProfiles(web.upload_profiles).filter(profile => profile.source !== '121');
    const nextWeb = { ...web, upload_profiles: [...local, ...profiles] };
    if (req.body?.persist !== false) await tasks.saveConfig({ ...current, web_submit: nextWeb });
    return { ok: true, settings: publicWebSubmit(nextWeb), groups: profiles, output: [`已从 121 同步 ${profiles.length} 个配置档`] };
  }

  async function sync121Styles(req) {
    const body = await request121(req, target.TARGET_CHECK_PATH);
    const catalog = parseTargetStyleCatalog(body);
    if (!catalog.length) throw new Error('121 未返回可用风格类型');
    const { tasks, current } = await readConfig(req);
    const old = Array.isArray(current.styles) ? current.styles : [];
    const styles = catalog.map(item => item.name);
    const added = styles.filter(name => !old.includes(name));
    const removed = old.filter(name => !styles.includes(name));
    const nextWeb = { ...object(current.web_submit), style_catalog: catalog };
    if (req.body?.persist !== false) await tasks.saveConfig({ ...current, styles, web_submit: nextWeb });
    return { ok: true, settings: publicWebSubmit(nextWeb), styles, style_sync: { old_count: old.length, new_count: styles.length, added, removed }, output: [`已从 121 同步 ${styles.length} 个风格类型`] };
  }

  async function planSubmission(req, body) {
    const ids = await selectTaskIds(req, body);
    if (body?.mode === 'selected' && !ids.length) throw new Error('没有收到选中的任务。请在任务列表勾选复选框后重新提交。');
    const webConfig = await currentWebConfig(req);
    const submitMode = 'version';
    const { tasks } = await resources(req);
    const groupsById = new Map();
    const skipped = [];
    const candidatesByBook = new Map();
    for (const id of ids) {
      const task = await tasks.getTask(req.username, id);
      if (!task?.meta) { skipped.push({ id, status: 'failed', error: '任务不存在' }); continue; }
      const versions = versionSelection.taskSelectedVersions(task.meta);
      for (const version of versions) {
        const content = await tasks.readVersionText(req.username, id, version);
        if (!content) { skipped.push({ id, version, status: 'skipped', error: '未找到正文版本' }); continue; }
        const size = Buffer.byteLength(content);
        if (webConfig.min_text_chars > 0 && String(content).length < webConfig.min_text_chars) { skipped.push({ id, version, status: 'file_too_small', error: `文案少于 ${webConfig.min_text_chars} 字`, size }); continue; }
        const submitted = task.meta.siteSubmitDoneVersions || [];
        const acceptedPending = task.meta.siteSubmitAcceptedVersions || [];
        if (webConfig.skip_submitted !== false && submitted.includes(version)) {
          skipped.push({ id, version, status: 'skipped', error: '该版本已确认提交；如需重复提交请开启“允许二次提交已成功版本”' });
          continue;
        }
        if (webConfig.skip_submitted !== false && acceptedPending.includes(version)) {
          skipped.push({ id, version, status: 'awaiting_confirmation', error: '121 已接收文件，但尚未确认生成任务；如需重新上传请开启“允许二次提交已成功版本”' });
          continue;
        }
        const profile = submitMode === 'version' ? selectVersionUploadProfile(webConfig, task.meta, version) : null;
        if (submitMode === 'version' && !profile) {
          skipped.push({ id, version, status: 'skipped', error: '版本配置模式：未找到可用的 121 配置档，请先同步配置档' });
          continue;
        }
        const candidates = candidatesByBook.get(id) || [];
        candidates.push({ id, version, size, task, profile });
        candidatesByBook.set(id, candidates);
      }
    }
    for (const candidates of candidatesByBook.values()) {
      // 同一本书的素材额度只使用一次；按本书本次选中的文案版本均分给每次上传。
      const allocations = target.distributeBookMaterials(webConfig.advanced, candidates.length);
      candidates.forEach((candidate, index) => {
        const { id, version, size, task, profile } = candidate;
        const profileAdvanced = submitMode === 'version'
          ? target.normalizeAdvanced(profile?.advanced || {})
          : target.normalizeAdvanced({ ...webConfig.advanced, jieyaNum: allocations[index].jieyaNum, gunpingNum: allocations[index].gunpingNum });
        const profileId = profile.id;
        const groupId = `${task.meta.platformId}-${task.meta.gender}-${task.meta.style}-${version}-${profileId}`;
        const profileName = profile.name;
        const group = groupsById.get(groupId) || { group_id: groupId, status: 'ready', version, summary: { ...snakeTask(task.meta), profile_id: profileId, profile_name: profileName }, advanced: profileAdvanced, items: [] };
        group.items.push({ id, version, size, advanced: profileAdvanced, profile_id: profileId, profile_name: profileName });
        groupsById.set(groupId, group);
      });
    }
    return { material_limit: target.PER_BOOK_MATERIAL_LIMIT, groups: [...groupsById.values()], skipped };
  }

  async function submitTasks(req, body) {
    if (!novelFetchStore) throw new Error('目标站会话存储未启用');
    const webConfig = await currentWebConfig(req);
    if (webConfig.enabled !== true) throw new Error('网站提交尚未启用；请先保存账号并勾选“启用网站提交”');
    const session = novelFetchStore.getSession(req.username);
    if (!session?.cookie) throw new Error('请先在网站提交页保存账号密码并登录目标站');
    const plan = await planSubmission(req, body);
    if (!plan.groups.length) {
      const reasons = [...new Set(plan.skipped.map(item => String(item.error || '').trim()).filter(Boolean))];
      throw new Error(reasons.length ? `没有可提交文案：${reasons.join('；')}` : '没有可提交的任务。请先勾选任务，并确认正文和上传配置档已准备好。');
    }
    const { tasks } = await resources(req);
    const submissionId = crypto.randomUUID().slice(0, 8);
    let successGroups = 0;
    let acceptedGroups = 0;
    let failedGroups = 0;
    for (const group of plan.groups) {
      group.group_key = group.group_id;
      group.group_id = `${group.group_id}-${submissionId}`;
      group.status = 'submitting';
      group.started_at = new Date().toISOString();
      await persistSiteSubmitGroup(req, group);
      let groupFailed = false;
      let groupAwaitingConfirmation = false;
      for (const item of group.items) {
        const task = await tasks.getTask(req.username, item.id);
        const content = await tasks.readVersionText(req.username, item.id, item.version);
        const trace = [];
        try {
          const fields = target.buildUploadFields({ platformId: task.meta.platformId, gender: task.meta.gender === '男频' ? '男' : task.meta.gender === '女频' ? '女' : task.meta.gender, style: task.meta.style, advanced: item.advanced || webConfig.advanced, styleMap: styleMapFromCatalog(webConfig.style_catalog) });
          appendSubmitTrace(trace, '本地参数校验', '完成', `121 参数：平台 ${fields.platform_id} / 性别 ${fields.gender} / 风格 ${fields.style} / 解压 ${fields.jieya_num} / 滚屏 ${fields.gunping_num}`, { fields: { platform_id: fields.platform_id, gender: fields.gender, style: fields.style, jieya_num: fields.jieya_num, gunping_num: fields.gunping_num } });
          const filename = target.buildTargetUploadFilename(item.id);
          const upload = target.buildMultipart(fields, { filename, content });
          appendSubmitTrace(trace, '准备上传文件', '完成', `已生成 ${filename}，${Buffer.byteLength(content)} 字节`, { file: filename, bytes: Buffer.byteLength(content), version: item.version });
          let data = {};
          let attempts = 0;
          let lastError = null;
          for (let attempt = 0; attempt <= webConfig.retry_times; attempt += 1) {
            try {
              attempts += 1;
              appendSubmitTrace(trace, '调用 121 上传接口', '提交中', `第 ${attempts} 次请求 ${target.TARGET_UPLOAD_PATH}`, { attempt: attempts });
              const response = await httpClient({ method: 'POST', url: `http://${target.TARGET_HOST}${target.TARGET_UPLOAD_PATH}`, headers: { 'Content-Type': `multipart/form-data; boundary=${upload.boundary}`, Cookie: session.cookie }, body: upload.body });
              try { data = JSON.parse(response.body); } catch (_) { data = {}; }
              if (data.success === true) { lastError = null; break; }
              lastError = new Error(data.message || data.msg || '上传失败');
              appendSubmitTrace(trace, '121 上传接口响应', '失败', lastError.message, { attempt: attempts });
            } catch (error) { lastError = error; }
          }
          if (lastError) throw lastError;
          const receipt = summarizeTargetReceipt(data);
          if (receipt.upload_error) {
            appendSubmitTrace(trace, '121 文件处理结果', '失败', receipt.upload_error, { upload_success_count: receipt.upload_success_count, upload_failed_count: receipt.upload_failed_count });
            throw new Error(receipt.upload_error);
          }
          appendSubmitTrace(trace, '121 文件处理结果', receipt.verified ? '完成' : '待确认', receipt.verified ? `121 已确认接收 ${receipt.upload_success_count} 个文件` : '121 未返回逐文件结果，等待人工核验', { upload_success_count: receipt.upload_success_count, upload_failed_count: receipt.upload_failed_count });
          try {
            receipt.remote_record = await verifyTargetBookRecord(httpClient, session.cookie, item.id, fields, item.advanced || webConfig.advanced);
            appendSubmitTrace(trace, '121 后台记录核验', receipt.remote_record.status, receipt.remote_record.detail, receipt.remote_record);
          } catch (error) {
            receipt.remote_record = { status: '待确认', found: false, detail: error.message || '121 后台记录核验失败' };
            appendSubmitTrace(trace, '121 后台记录核验', '待确认', receipt.remote_record.detail);
          }
          if (receipt.verified) {
            await tasks.updateTaskMeta(req.username, item.id, {
              siteSubmitStatus: 'submitted',
              siteSubmitDoneVersions: [...new Set([...(task.meta.siteSubmitDoneVersions || []), item.version])],
              siteSubmitAcceptedVersions: (task.meta.siteSubmitAcceptedVersions || []).filter(version => version !== item.version),
              siteSubmitFailedVersions: (task.meta.siteSubmitFailedVersions || []).filter(version => version !== item.version)
            });
            if (typeof tasks.appendSiteSubmitLog === 'function') await tasks.appendSiteSubmitLog(req.username, item.id, { status: 'submitted', version: item.version, group_id: group.group_id, material_allocation: item.advanced, retries: Math.max(0, attempts - 1), remote_receipt: receipt, execution_trace: trace, time: new Date().toISOString() });
          } else {
            groupAwaitingConfirmation = true;
            item.status = 'accepted_pending';
            item.remote_receipt = receipt;
            await tasks.updateTaskMeta(req.username, item.id, {
              siteSubmitStatus: 'accepted_pending',
              siteSubmitAcceptedVersions: [...new Set([...(task.meta.siteSubmitAcceptedVersions || []), item.version])],
              siteSubmitFailedVersions: (task.meta.siteSubmitFailedVersions || []).filter(version => version !== item.version)
            });
            if (typeof tasks.appendSiteSubmitLog === 'function') await tasks.appendSiteSubmitLog(req.username, item.id, { status: 'accepted_pending', version: item.version, group_id: group.group_id, material_allocation: item.advanced, retries: Math.max(0, attempts - 1), remote_receipt: receipt, execution_trace: trace, time: new Date().toISOString() });
          }
        } catch (error) {
          groupFailed = true;
          item.error = error.message || '上传失败';
          appendSubmitTrace(trace, '提交结束', '失败', item.error);
          await tasks.updateTaskMeta(req.username, item.id, { siteSubmitStatus: 'failed', siteSubmitFailedVersions: [...new Set([...(task.meta.siteSubmitFailedVersions || []), item.version])] });
          if (typeof tasks.appendSiteSubmitLog === 'function') await tasks.appendSiteSubmitLog(req.username, item.id, { status: 'failed', version: item.version, group_id: group.group_id, material_allocation: item.advanced, error: item.error, execution_trace: trace, time: new Date().toISOString() });
        }
        await persistSiteSubmitGroup(req, group);
      }
      group.status = groupFailed ? 'failed' : groupAwaitingConfirmation ? 'accepted_pending' : 'submitted';
      group.finished_at = new Date().toISOString();
      await persistSiteSubmitGroup(req, group);
      if (groupFailed) failedGroups++;
      else if (groupAwaitingConfirmation) acceptedGroups++;
      else successGroups++;
    }
    return { ...plan, success_groups: successGroups, accepted_groups: acceptedGroups, failed_groups: failedGroups, tasks: await listTasks(req) };
  }

  function sendConfigResponse(req, res, payload) {
    const body = JSON.stringify(payload);
    const etag = `"${crypto.createHash('sha1').update(body).digest('hex')}"`;
    res.setHeader('Cache-Control', 'private, max-age=5, must-revalidate');
    res.setHeader('ETag', etag);
    if (req.headers['if-none-match'] === etag) return res.status(304).end();
    return res.type('json').send(body);
  }

  router.get('/bootstrap', async (req, res) => {
    try { return sendConfigResponse(req, res, (await readConfig(req, { lightweight: true })).response); }
    catch (error) { return res.status(500).json({ error: error.message }); }
  });
  router.get('/config', async (req, res) => {
    try { return sendConfigResponse(req, res, (await readConfig(req)).response); }
    catch (error) { return res.status(500).json({ error: error.message }); }
  });
  router.post('/config', async (req, res) => { try { res.json({ ok: true, config: (await saveConfig(req, req.body)).response }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/work-form', async (req, res) => { try { const state = object(req.body?.state); const { tasks, current } = await readConfig(req); await tasks.saveConfig({ ...current, work_form: state }); res.json({ ok: true, state }); } catch (error) { res.status(400).json({ error: error.message }); } });

  router.post('/process/start', async (req, res) => {
    const id = crypto.randomUUID();
    const job = { id, status: 'running', started_at: new Date().toISOString(), updated_at: new Date().toISOString(), steps: [], result: null, error: '' };
    jobs.set(`${req.username}:${id}`, job);
    await persistProcessJob(req, job);
    res.json(job);
    setImmediate(async () => {
      let persistence = Promise.resolve();
      const report = step => {
        job.steps.push({ ...step, time: step?.time || new Date().toISOString() });
        persistence = persistence.then(() => persistProcessJob(req, job)).catch(() => {});
      };
      try { job.result = await processPayload(req, req.body, report); job.status = 'done'; }
      catch (error) { job.status = 'failed'; job.error = error.message || '处理失败'; }
      finally {
        job.completed_at = new Date().toISOString();
        await persistence;
        try { await persistProcessJob(req, job); } catch (_) {}
      }
    });
  });
  router.post('/process', async (req, res) => { try { res.json(await processPayload(req, req.body)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/process/jobs/latest', async (req, res) => { try { const list = await latestProcessJobs(req, req.query?.limit); res.json({ latest: list[0] || {}, jobs: list }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/process/jobs/:id', async (req, res) => { try { const job = await findProcessJob(req, req.params.id); if (!job) return res.status(404).json({ error: '处理任务不存在' }); return res.json(job); } catch (error) { res.status(400).json({ error: error.message }); } });

  router.get('/tasks', async (req, res) => { try { res.json({ tasks: await listTasks(req) }); } catch (error) { res.status(500).json({ error: error.message }); } });
  router.post('/tasks/batch-delete', async (req, res) => { try { const ids = await selectTaskIds(req, req.body); const { tasks } = await resources(req); const result = await tasks.deleteTasks(req.username, ids); res.json({ ...result, failed: 0, tasks: await listTasks(req) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/batch-retry', async (req, res) => { try {
    const ids = await selectTaskIds(req, req.body);
    const { tasks, configStore: store } = await resources(req);
    const config = object(store.getConfig());
    let retried = 0;
    let failed = 0;
    let skipped = 0;
    const retry_stage_counts = {};
    for (const id of ids) {
      try {
        const task = await tasks.getTask(req.username, id);
        if (!task) throw new Error('任务不存在');
        const plan = retryPlanForTask(task.meta || task);
        if (!plan.stage) {
          skipped++;
          retry_stage_counts.skipped = (retry_stage_counts.skipped || 0) + 1;
          continue;
        }
        retry_stage_counts[plan.stage] = (retry_stage_counts[plan.stage] || 0) + 1;
        if (plan.stage === 'classify') {
          const result = await classifier.classifyMissingRows({ configStore: store, tasks: [task.meta || task] });
          const classified = Array.isArray(result?.tasks) ? result.tasks : [];
          if (!classified.length || result?.errors?.length) throw new Error(result?.errors?.join('；') || 'AI 判断失败');
          if (typeof tasks.saveTasks !== 'function') throw new Error('当前存储不支持保存 AI 判断结果');
          await tasks.saveTasks(req.username, classified);
        } else {
          if (plan.fetchOriginal) {
            const result = await tasks.fetchOriginal(req.username, id, task.meta?.maxTxt || task.meta?.max_txt || 4000);
            if (result?.status !== 'done') throw new Error('原文抓取失败');
          }
          if (plan.applyOriginalRules) await applySavedRulesToOriginal(tasks, req.username, id, config, store);
        }
        if (plan.generateAi && (plan.stage === 'rewrite' || config.workflow?.auto_rewrite_after_fetch)) {
          const refreshed = await tasks.getTask(req.username, id);
          await rewrite.generateAiVersions({ configStore: store, tasks, username: req.username, task: refreshed.meta, versions: refreshed.meta.selectedVersions, slotMethods: refreshed.meta.aiSlotMethods });
        }
        retried++;
      } catch (error) {
        failed++;
        if (typeof tasks.appendLog === 'function') await tasks.appendLog(req.username, id, 'retry_failed', { error: error.message || '重试失败', status: 'failed' });
      }
    }
    res.json({ retried, failed, skipped, retry_stage_counts, tasks: await listTasks(req) });
  } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/apply-rules', async (req, res) => { try {
    const ids = await selectTaskIds(req, req.body);
    const { tasks, configStore: store } = await resources(req);
    const config = withSensitiveAiEnabled(object(store.getConfig()), req.body?.sensitive_ai_enabled);
    const scope = ['original', 'ai', 'both'].includes(req.body?.scope) ? req.body.scope : 'both';
    let applied = 0;
    for (const id of ids) {
      if (scope === 'original' || scope === 'both') if (await applySavedRulesToOriginal(tasks, req.username, id, config, store)) applied++;
      if (scope === 'ai' || scope === 'both') {
        const task = await tasks.getTask(req.username, id);
        for (const version of versionSelection.generatedVersions(task?.meta || {})) {
          if (await applySavedRulesToVersion(tasks, req.username, id, version, config)) applied++;
        }
      }
    }
    res.json({ applied, failed: 0, tasks: await listTasks(req) });
  } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/reprocess-sensitive', async (req, res) => { try {
    const ids = await selectTaskIds(req, req.body);
    const { tasks, configStore: store } = await resources(req);
    const config = withSensitiveAiEnabled(object(store.getConfig()), req.body?.sensitive_ai_enabled);
    const restoreFromBackup = req.body?.restore_from_backup === true;
    let processed = 0;
    let restored = 0;
    let failed = 0;
    for (const id of ids) {
      try {
        const task = await tasks.getTask(req.username, id);
        if (!task?.meta) throw new Error('任务不存在');
        if (restoreFromBackup) {
          if (typeof tasks.restoreOriginal !== 'function') throw new Error('当前存储不支持恢复原文备份');
          await tasks.restoreOriginal(req.username, id);
          restored++;
        }
        await applySavedRulesToOriginal(tasks, req.username, id, config, store);
        if (typeof tasks.appendLog === 'function') {
          await tasks.appendLog(req.username, id, 'sensitive_reprocessed', {
            source: restoreFromBackup ? 'original_backup' : 'current_text',
            sensitive_ai_enabled: object(config.sensitive_ai).enabled === true
          });
        }
        processed++;
      } catch (_) {
        failed++;
      }
    }
    res.json({ processed, restored, failed, tasks: await listTasks(req) });
  } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/tasks/:id', async (req, res) => { try { const { tasks } = await resources(req); const task = await tasks.getTask(req.username, req.params.id); if (!task?.meta) return res.status(404).json({ error: '任务不存在' }); const aiTexts = []; for (const version of versionSelection.generatedVersions(task.meta)) aiTexts.push({ name: version.toUpperCase(), version, text: await tasks.readVersionText(req.username, req.params.id, version) }); const sensitiveLog = await readSensitiveLog(tasks, req.username, req.params.id); res.json({ meta: legacyMeta(task.meta), original: await tasks.readOriginal(req.username, req.params.id), ai_texts: aiTexts, has_original_raw: task.hasOriginalRaw === true, ...sensitiveLog, logs: await tasks.readLogs(req.username, req.params.id) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/:id/fetch', async (req, res) => { try { const { tasks, configStore: store } = await resources(req); const task = await tasks.getTask(req.username, req.params.id); if (!task) throw new Error('任务不存在'); const result = await tasks.fetchOriginal(req.username, req.params.id, task.meta.maxTxt || 4000); const current = await tasks.getTask(req.username, req.params.id); if (result.status === 'refresh_failed') return res.json({ ok: false, status: 'refresh_failed', preserved: true, message: current?.meta?.originalRefreshErrorMessage || '本次刷新原文失败，已保留原文', task: current }); if (result.status !== 'done') throw new Error(current?.meta?.originalErrorMessage || current?.meta?.error || '原文抓取失败'); const config = withSensitiveAiEnabled(object(store.getConfig()), req.body?.sensitive_ai_enabled); await applySavedRulesToOriginal(tasks, req.username, req.params.id, config, store); res.json({ ok: true, status: 'done', task: await tasks.getTask(req.username, req.params.id) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/:id/restore-original', async (req, res) => { try {
    const { tasks, configStore: store } = await resources(req);
    if (typeof tasks.restoreOriginal !== 'function') throw new Error('当前存储不支持恢复原文');
    await tasks.restoreOriginal(req.username, req.params.id);
    const config = withSensitiveAiEnabled(object(store.getConfig()), req.body?.sensitive_ai_enabled);
    await applySavedRulesToOriginal(tasks, req.username, req.params.id, config, store);
    res.json({ task: await tasks.getTask(req.username, req.params.id) });
  } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/tasks/:id/generate-ai', async (req, res) => { try {
    const { tasks, configStore: store } = await resources(req);
    const task = await tasks.getTask(req.username, req.params.id);
    if (!task?.meta) throw new Error('任务不存在');
    const hasExplicitVersions = Array.isArray(req.body?.selected_versions);
    const versions = versionSelection.normalizeSelectedVersions(hasExplicitVersions ? req.body.selected_versions : task.meta.selectedVersions, hasExplicitVersions ? [] : undefined);
    if (!versions.length) throw new Error('请至少选择一个文案版本');
    const slotMethods = versionSelection.normalizeAiSlotMethods(req.body?.ai_slot_methods || task.meta.aiSlotMethods);
    await tasks.updateTaskMeta(req.username, req.params.id, { selectedVersions: versions, aiSlotMethods: slotMethods, status: 'ai_processing', aiError: '' });
    const result = await rewrite.generateAiVersions({ configStore: store, tasks, username: req.username, task: { ...task.meta, selectedVersions: versions, aiSlotMethods: slotMethods }, versions, slotMethods });
    const status = result.status === 'done'
      ? 'ai_done'
      : result.status === 'partial'
        ? 'ai_partial'
        : result.status === 'failed'
          ? 'ai_failed'
          : result.status;
    await tasks.updateTaskMeta(req.username, req.params.id, { status, ...(result.status === 'done' ? { aiError: '' } : {}) });
    const current = await tasks.getTask(req.username, req.params.id);
    const response = { ok: result.status === 'done', status: result.status, result, task: current };
    if (result.status === 'failed' || result.status === 'partial') {
      return res.status(422).json({ ...response, error: result.error || (result.status === 'partial' ? '部分 AI 文案生成失败' : 'AI 文案生成失败') });
    }
    if (result.status === 'waiting_original' || result.status === 'waiting_ai_config') {
      return res.status(422).json({ ...response, error: result.error || (result.status === 'waiting_original' ? '原文尚未就绪' : 'AI 配置未完成') });
    }
    return res.json(response);
  } catch (error) {
    try {
      const { tasks } = await resources(req);
      const task = await tasks.getTask(req.username, req.params.id);
      const reconciled = await rewrite.reconcileAiTaskStatus({
        tasks,
        username: req.username,
        bookId: req.params.id,
        task: task?.meta,
        lastAttemptError: error.message || 'AI文案生成失败'
      });
      await tasks.updateTaskMeta(req.username, req.params.id, {
        status: reconciled.status === 'done' ? 'ai_done' : (reconciled.status === 'partial' ? 'ai_partial' : 'ai_failed')
      });
    } catch (_) { /* preserve the original route error */ }
    res.status(400).json({ error: error.message });
  } });
  router.get('/tasks/:id/sensitive-log', async (req, res) => { try { const { tasks } = await resources(req); const task = await tasks.getTask(req.username, req.params.id); if (!task?.meta) return res.status(404).json({ error: '任务不存在' }); res.json({ meta: legacyMeta(task.meta), ...(await readSensitiveLog(tasks, req.username, req.params.id)) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/tasks/:id/rules-trace', async (req, res) => { try { const { tasks, configStore: store } = await resources(req); const task = await tasks.getTask(req.username, req.params.id); const text = task ? await tasks.readOriginal(req.username, req.params.id) : ''; res.json({ meta: legacyMeta(task?.meta || {}), stages: rules.processConfiguredDocumentTrace(text, 'original', object(store.getConfig())) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/tasks/:id/site-submit-log', async (req, res) => { try { const { tasks } = await resources(req); const task = await tasks.getTask(req.username, req.params.id); const logs = typeof tasks.readSiteSubmitLog === 'function' ? await tasks.readSiteSubmitLog(req.username, req.params.id) : []; res.json({ meta: legacyMeta(task?.meta || {}), result: logs.at(-1) || {}, logs }); } catch (error) { res.status(400).json({ error: error.message }); } });

  router.get('/knowledge', async (req, res) => { try { res.json({ summary: summarizeKnowledge(await requestKnowledge(req)) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/knowledge/optimize', async (req, res) => { try { const kind = req.body?.library_type; const item = object(req.body?.item); const saved = knowledge.save(kind, item); const { configStore: store } = await resources(req); const settings = ai.resolveAiSettings(store, 'rewrite'); res.json(await knowledge.optimizeItem({ kind, id: saved.item.id }, ai, settings)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/opening/analyze', async (req, res) => { try { const { configStore: store } = await resources(req); const settings = ai.resolveAiSettings(store, 'rewrite'); res.json(await opening.analyze(ai, settings, req.body?.source_text)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/opening/save', (req, res) => { try { res.json(opening.save(req.body?.item)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/opening/normalize', (req, res) => { try { res.json(opening.normalize()); } catch (error) { res.status(400).json({ error: error.message }); } });
  // 连通性测试和正式执行共享模型目录解析；不接受请求体传入的 base_url / api_key / model 覆盖。
  router.post('/ai/test', async (req, res) => { try { const { configStore: store } = await resources(req); const settings = ai.resolveAiSettings(store, req.body?.purpose || 'classifier'); const result = await ai.chatCompletion(settings, [{ role: 'user', content: '请只回复：ok' }], { temperature: 0 }); res.json({ ok: true, content: result.text || '' }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/rules/preview', async (req, res) => { try { const { configStore: store } = await resources(req); const text = String(req.body?.text || ''); res.json({ processed: processSavedRules(text, req.body?.scope || 'original', object(store.getConfig())) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/rules/ai-suggest', async (req, res) => {
    try {
      const { configStore: store } = await resources(req);
      const settings = ai.resolveAiSettings(store, 'rewrite');
      const ruleType = String(req.body?.rule_type || 'sensitive');
      const response = await ai.chatCompletion(settings, [
        { role: 'system', content: '你是批量小说改文系统规则助手。只输出可直接保存的 JSON 规则对象，不要解释。' },
        { role: 'user', content: JSON.stringify({ rule_type: ruleType, sample_text: req.body?.sample_text || '', goal: req.body?.goal || '生成可直接使用的规则建议。' }) }
      ], { temperature: 0.2 });
      res.json({ ok: true, suggestions: ai.parseAiJsonContent(response.text || '') });
    } catch (error) { res.status(400).json({ error: error.message || '规则建议生成失败' }); }
  });

  router.get('/web-submit/config', async (req, res) => res.json({ settings: publicWebSubmit(await currentWebConfig(req)) }));
  router.get('/web-submit/history', async (req, res) => { try { res.json({ records: await listSiteSubmitHistory(req), groups: await listPersistedSiteSubmitGroups(req) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/issues', async (req, res) => { try { res.json({ entries: await listWorkshopIssues(req) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/web-submit/config', async (req, res) => { try { res.json({ ok: true, settings: await saveWebConfig(req, req.body?.settings), tasks: await listTasks(req) }); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/web-submit/environment', (req, res) => { const session = novelFetchStore?.getSession(req.username); res.json({ ok: Boolean(session), checks: [{ name: '目标站登录会话', ok: Boolean(session), detail: session ? '已登录' : '未登录' }, { name: '上传接口', ok: true, detail: target.TARGET_UPLOAD_PATH }] }); });
  router.post('/web-submit/sync-configs', async (req, res) => { try { res.json(await sync121Profiles(req)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/web-submit/sync-styles', async (req, res) => { try { res.json(await sync121Styles(req)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/web-submit/test-visible', async (req, res) => {
    try {
      const session = novelFetchStore?.getSession(req.username);
      if (!session?.cookie) throw new Error('请先保存账号密码并登录目标站');
      const check = await httpClient({ method: 'GET', url: `http://${target.TARGET_HOST}${target.TARGET_CHECK_PATH}`, headers: { Cookie: session.cookie }, timeoutMs: 15000 });
      const ok = target.isDashboard(check.body);
      res.json({ ok, output: [], result: { checks: { login_session: ok, upload_endpoint: target.TARGET_UPLOAD_PATH } } });
    } catch (error) { res.status(400).json({ ok: false, error: error.message || '验证失败' }); }
  });
  router.post('/web-submit/preview', async (req, res) => { try { res.json(await planSubmission(req, req.body)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.post('/web-submit/submit', async (req, res) => { try { res.json(await submitTasks(req, req.body)); } catch (error) { res.status(400).json({ error: error.message }); } });
  router.get('/logs', async (req, res) => { try { const { tasks } = await resources(req); const records = await listTasks(req); const logs = []; for (const task of records.slice(0, 100)) for (const item of await tasks.readLogs(req.username, task.id)) logs.push({ ...item, book_id: task.id }); res.json({ logs: logs.slice(-500).reverse() }); } catch (error) { res.status(400).json({ error: error.message }); } });

  return router;
}

module.exports = { createBatchRewriteRouter, snakeTask, retryPlanForTask };
