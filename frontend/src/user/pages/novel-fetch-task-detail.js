import { novelFetchStageLabel } from './novel-fetch-status.js';

const PLATFORM_LABELS = {
  1: '黑岩付费', 2: '番茄付费', 3: '七猫付费', 4: '点众付费', 7: '番茄免费',
  15: '知乎付费', 20: '掌阅付费', 26: '卓越付费', 29: '九州书城', 31: '掌文付费'
};

const present = value => value !== undefined && value !== null && value !== '';

function first(...values) {
  return values.find(present);
}

function display(value, fallback = '—') {
  if (!present(value)) return fallback;
  if (typeof value === 'boolean') return value ? '是' : '否';
  return String(value);
}

function countValue(...values) {
  const value = first(...values);
  const number = Number(value);
  return Number.isFinite(number) ? number : 0;
}

function formatTime(value) {
  if (!present(value)) return '—';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString('zh-CN', { hour12: false });
}

function bodyContent(value) {
  if (typeof value === 'string') return value;
  if (!value || typeof value !== 'object') return '';
  return String(first(value.content, value.text, value.body, value.value) || '');
}

function modeLabel(value) {
  return {
    direct: '直接替换', system_rule: '系统规则替换', per_line: 'AI 逐条修改', group: 'AI 整组修改'
  }[String(value || '').toLowerCase()] || display(value);
}

function stageLabel(value) {
  return String(value || '').toLowerCase() === 'input_ready' ? '分类信息已就绪' : novelFetchStageLabel(value);
}

function sourceLabel(value) {
  return {
    ai: 'AI 处理', ai_group: 'AI 整组处理', group: 'AI 整组处理',
    system_rule: '系统规则', direct: '直接替换', fallback: '规则兜底替换'
  }[String(value || '').toLowerCase()] || display(value);
}

export function novelFetchAiVersions(data = {}) {
  const versions = data.versions && typeof data.versions === 'object' ? data.versions : {};
  const meta = data.meta && typeof data.meta === 'object' ? data.meta : {};
  const history = first(meta.rewriteKnowledgeHistory, meta.rewrite_knowledge_history);
  const evidence = first(meta.rewriteEvidenceByVersion, meta.rewrite_evidence_by_version) || {};
  const historyByVersion = new Map((Array.isArray(history) ? history : []).map(item => [String(first(item.version, item.versionId, item.ai_index, item.aiIndex, '')).toLowerCase(), item]));
  const rows = Object.entries(versions)
    .filter(([key]) => /^ai\d+$/i.test(key))
    .map(([key, value]) => {
      const item = { key: key.toLowerCase(), label: `AI 版本 ${Number(key.replace(/\D/g, ''))}`, content: bodyContent(value) };
      const knowledge = evidence[key.toLowerCase()] || historyByVersion.get(key.toLowerCase()) || historyByVersion.get(`ai${key.replace(/\D/g, '')}`);
      if (knowledge) item.knowledge = knowledge;
      return item;
    })
    .filter(item => item.content.trim())
    .sort((left, right) => Number(left.key.slice(2)) - Number(right.key.slice(2)));
  if (rows.length) return rows;
  const legacy = bodyContent(first(data.aiText, data.ai_text));
  const legacyRow = { key: 'ai1', label: 'AI 版本 1', content: legacy };
  if (legacy.trim() && historyByVersion.get('ai1')) legacyRow.knowledge = historyByVersion.get('ai1');
  return legacy.trim() ? [legacyRow] : [];
}

export function novelFetchTaskSummaryItems(data = {}) {
  const meta = data.meta && typeof data.meta === 'object' ? data.meta : {};
  const bookInfo = data.bookInfo && typeof data.bookInfo === 'object'
    ? data.bookInfo
    : (data.book_info && typeof data.book_info === 'object' ? data.book_info : {});
  const sensitiveHits = meta.sensitiveHits && typeof meta.sensitiveHits === 'object'
    ? meta.sensitiveHits
    : (meta.sensitive_hits && typeof meta.sensitive_hits === 'object' ? meta.sensitive_hits : {});
  const sensitiveFixed = meta.sensitiveFixed && typeof meta.sensitiveFixed === 'object'
    ? meta.sensitiveFixed
    : (meta.sensitive_fixed && typeof meta.sensitive_fixed === 'object' ? meta.sensitive_fixed : {});
  const versions = novelFetchAiVersions(data);
  const currentVersionKey = String(first(meta.aiCurrentVersion, meta.ai_current_version, '')).toLowerCase();
  const latestVersion = versions.find(version => version.key === currentVersionKey) || versions.at(-1);
  const knowledge = latestVersion?.knowledge || {};
  const platformId = first(meta.platformId, meta.platform_id, data.platformId, data.platform_id);
  const platform = first(meta.platformName, meta.platform_name, data.platformName, data.platform_name, PLATFORM_LABELS[platformId], platformId);
  const processed = countValue(data.processedChars, data.processed_chars, meta.processedChars, meta.processed_chars, data.original?.length);
  const raw = countValue(data.originalRawChars, data.original_raw_chars, meta.originalRawChars, meta.original_raw_chars, data.originalRaw?.length);
  const hitCount = countValue(meta.sensitiveHitCount, meta.sensitive_hit_count, sensitiveHits.hit_count, sensitiveHits.hitCount, data.sensitiveHitCount, data.sensitive_hit_count);
  const fixedCount = countValue(meta.sensitiveFixedCount, meta.sensitive_fixed_count, sensitiveFixed.fixed_count, sensitiveFixed.fixedCount, data.sensitiveFixedCount, data.sensitive_fixed_count);
  const readyReferenceCount = Object.entries(data.bodyRefs || {}).filter(([key, ref]) => /^ai\d+$/i.test(key) && ref?.state === 'ready').length;
  const aiCount = countValue(meta.aiGeneratedCount, meta.ai_generated_count, data.aiGeneratedCount, data.ai_generated_count, readyReferenceCount || versions.length);
  // meta.error 是任务级通用错误，可能来自 AI、敏感词或提交阶段；
  // 只有明确的原文错误字段才能显示在“原文失败原因”，否则会误导排查方向。
  const originalError = first(meta.originalError, meta.original_error, meta.originalFetchLastError, meta.original_fetch_last_error, data.original_fetch_last_error);
  const upstreamCode = /(?:business error|HTTP)\s+(\d{3})\b/i.exec(String(originalError || ''))?.[1];
  return [
    ['书籍 ID', first(data.bookId, data.book_id, meta.bookId, meta.book_id, meta.id)],
    ['书名', first(data.bookName, data.book_name, meta.bookName, meta.book_name, meta.title, bookInfo.workTitle, bookInfo.work_title)],
    ['平台', platform],
    ['风格', first(data.style, meta.style, meta.styleName, meta.style_name)],
    ['男女频', first(data.gender, meta.gender, meta.channel)],
    ['判断依据', first(meta.genderReason, meta.gender_reason, meta.genderEvidence, meta.gender_evidence, meta.aiGenderReason, meta.ai_gender_reason)],
    ['分类模型', first(meta.classifierModel, meta.classifier_model)],
    ['分类失败原因', first(meta.classifyError, meta.classify_error)],
    ['原文状态', stageLabel(first(data.originalStatus, data.original_status, meta.originalStatus, meta.original_status))],
    ['原文失败原因', originalError],
    ['原文错误码', first(meta.originalErrorCode, meta.original_error_code, meta.originalUpstreamCode, meta.original_upstream_code, meta.originalFetchErrorCode, meta.original_fetch_error_code, upstreamCode)],
    ['处理字数', raw > 0 ? `${processed} / ${raw}` : (processed > 0 ? String(processed) : '—')],
    ['敏感词状态', stageLabel(first(data.sensitiveStatus, data.sensitive_status, meta.sensitiveStatus, meta.sensitive_status))],
    ['敏感词模型', first(meta.sensitiveModel, meta.sensitive_model, sensitiveFixed.model, sensitiveHits.model)],
    ['命中 / 已处理', `${hitCount} / ${fixedCount}`],
    ['AI 状态', stageLabel(first(data.aiStatus, data.ai_status, meta.aiStatus, meta.ai_status))],
    ['AI 失败原因', first(meta.aiError, meta.ai_error)],
    ['AI 当前版本', first(meta.aiCurrentVersion, meta.ai_current_version, latestVersion?.key.toUpperCase())],
    ['AI 文案版本', `${aiCount || versions.length} 个`],
    ['AI 尝试次数', first(meta.aiLastAttemptCount, meta.ai_last_attempt_count)],
    ['AI 最近失败', first(meta.aiLastAttemptError, meta.ai_last_attempt_error, '无')],
    ['网站提交失败原因', first(meta.siteSubmitError, meta.site_submit_error)],
    ['改文模型', first(knowledge.model, meta.rewriteModel, meta.rewrite_model)],
    ['改文方案', first(knowledge.strategy_name, knowledge.strategyName, knowledge.method, meta.rewriteKnowledge?.strategyName, meta.rewrite_knowledge?.strategy_name, meta.rewriteKnowledge?.strategy, meta.rewrite_knowledge?.strategy)],
    ['改文模板', first(knowledge.template_name, knowledge.rewrite_template_name, meta.rewriteKnowledge?.rewriteTemplateName, meta.rewrite_knowledge?.rewrite_template_name)],
    ['开头词', first(knowledge.opening_phrase_name, meta.rewriteKnowledge?.openingPhraseName, meta.rewrite_knowledge?.opening_phrase_name)],
    ['高仿参考', first(knowledge.high_imitation_reference_name, meta.rewriteKnowledge?.highImitationReferenceName, meta.rewrite_knowledge?.high_imitation_reference_name)],
    ['创建时间', formatTime(first(data.createdAt, data.created_at, meta.createdAt, meta.created_at))],
    ['更新时间', formatTime(first(data.updatedAt, data.updated_at, meta.updatedAt, meta.updated_at))],
    ['当前批次', first(data.batchId, data.batch_id, meta.batchId, meta.batch_id, meta.currentBatchId, meta.current_batch_id)]
  ].map(([label, value]) => ({ key: label, label, value: display(value) }));
}

export function novelFetchSensitiveComparisons(data = {}) {
  const meta = data.meta && typeof data.meta === 'object' ? data.meta : {};
  const hitsData = data.sensitive_hits || data.sensitiveHits || {};
  const fixedData = data.sensitive_fixed || data.sensitiveFixed || {};
  const hits = Array.isArray(hitsData.hits) ? hitsData.hits : [];
  const fixedItems = Array.isArray(fixedData.items) ? fixedData.items : [];
  const fixedByIndex = new Map(fixedItems.map(item => [String(first(item.hit_index, item.hitIndex, '')), item]));
  const items = hits.map((hit, rowIndex) => {
    const index = first(hit.hit_index, hit.hitIndex, rowIndex + 1);
    const fixed = fixedByIndex.get(String(index)) || {};
    return {
      index,
      keyword: display(first(hit.keyword, hit.word), '未知词'),
      status: novelFetchStageLabel(first(fixed.status, fixedData.status, hitsData.status, '')),
      position: present(hit.start) || present(hit.end) ? `${display(hit.start, '')} - ${display(hit.end, '')}` : '—',
      source: sourceLabel(first(fixed.source, fixedData.source)),
      // Deterministic replacements have no AI model. Show the actual engine
      // instead of an empty placeholder that makes a working rule look broken.
      model: display(first(fixed.model, fixed.model_id, fixed.modelId),
        String(first(fixed.source, fixedData.source)).toLowerCase() === 'system_rule' ? '规则引擎' : '—'),
      attempts: first(fixed.attempt, fixed.attempt_count, fixed.attemptCount, present(fixed.status) ? 1 : '—'),
      error: display(first(fixed.error, fixed.error_message, fixed.errorMessage)),
      recovered: ['done', 'completed', 'fixed'].includes(String(fixed.status || fixedData.status || '').toLowerCase()) && ['fallback', 'system_rule'].includes(String(fixed.source || fixedData.source || '').toLowerCase()),
      before: display(first(hit.snippet, fixed.before), '暂无命中片段'),
      after: display(fixed.after, '暂无处理后内容')
    };
  });
  return {
    summary: {
      bookId: display(first(meta.bookId, meta.book_id, meta.id)),
      bookName: display(first(meta.bookName, meta.book_name, meta.title)),
      mode: modeLabel(first(hitsData.mode, fixedData.mode, meta.sensitiveMode, meta.sensitive_mode)),
      status: novelFetchStageLabel(first(fixedData.status, hitsData.status, meta.sensitiveStatus, meta.sensitive_status)),
      hitCount: countValue(hitsData.hit_count, hitsData.hitCount, meta.sensitiveHitCount, meta.sensitive_hit_count, hits.length),
      fixedCount: countValue(fixedData.fixed_count, fixedData.fixedCount, meta.sensitiveFixedCount, meta.sensitive_fixed_count, fixedItems.length),
      hitUpdatedAt: formatTime(first(hitsData.updated_at, hitsData.updatedAt)),
      fixedUpdatedAt: formatTime(first(fixedData.updated_at, fixedData.updatedAt))
    },
    items,
    logs: Array.isArray(data.logs) ? data.logs : []
  };
}

export function novelFetchSubmitDiagnostic(data = {}) {
  const result = data.result && typeof data.result === 'object' && Object.keys(data.result).length ? data.result : null;
  const statusLabels = { submitted: '已回读确认', completed: '已完成', failed: '提交失败', error: '提交失败', waiting_login: '等待登录', waiting_configuration: '等待配置', pending_confirmation: '等待后台确认', started: '开始提交', queued: '等待提交', skipped: '已跳过' };
  const eventLabels = { site_submit: '提交结果', auto_submit_started: '开始自动提交', auto_submit_completed: '自动提交完成', auto_submit_failed: '自动提交失败', auto_submit_deferred: '等待提交' };
  const rows = (Array.isArray(data.logs) ? data.logs : []).map((row, index) => ({
    ...row,
    key: String(index),
    timeLabel: formatTime(first(row.time, row.createdAt, row.created_at)),
    statusLabel: statusLabels[row.status] || eventLabels[row.event] || '提交记录',
    versionLabel: row.version ? String(row.version).toUpperCase() : '—',
    message: String(first(row.error, row.message, row.detail, row.remote_receipt?.verified === true ? '视频管理系统已回读确认' : '', typeof row.result === 'string' ? row.result : '', eventLabels[row.event]) || '—')
  }));
  return { result, rows, verified: result?.remote_receipt?.verified === true || result?.verified === true,
    receiptId: display(first(result?.remote_receipt?.submission_id, result?.submissionId)),
    reason: String(data.reason || '尚未发起提交，请先完成正文与 AI 文案处理') };
}
