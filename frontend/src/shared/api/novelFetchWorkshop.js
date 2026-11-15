import { apiRequest } from './client.js';

export function workshopPlatformsPath() {
  return '/api/novel-fetch-workshop/platforms';
}

export function getWorkshopPlatforms() {
  return apiRequest(workshopPlatformsPath());
}

// 改文工作台：后端接口封装（apiRequest 统一处理鉴权与错误上报）

export function listBatchRewriteIssues() {
  return apiRequest('/api/novel-fetch-workshop/issues');
}

// 批量处理清单：解析 →（可选）分类 → 建任务 →（可选）抓原文 →（可选）AI 改文
export function processBatch(payload) {
  return apiRequest('/api/novel-fetch-workshop/process', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

// 启动后台批处理作业
export function startWorkshopProcess(payload) {
  return apiRequest('/api/novel-fetch-workshop/process/start', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

// 查询单个后台作业进度
export function getWorkshopJob(jobId) {
  return apiRequest(`/api/novel-fetch-workshop/process/jobs/${encodeURIComponent(jobId)}`);
}

// 查询最近后台作业
export function getWorkshopJobs(limit = 10, options = {}) {
  return apiRequest(`/api/novel-fetch-workshop/process/jobs/latest?limit=${encodeURIComponent(limit)}`, options);
}

export function workshopRealtimeStatusPath() {
  return '/api/novel-fetch-workshop/realtime/status';
}

export function getWorkshopRealtimeStatus() {
  return apiRequest(workshopRealtimeStatusPath());
}

export function startWorkshopAutomation() {
  return apiRequest('/api/novel-fetch-workshop/automation/start', { method: 'POST', body: JSON.stringify({}) });
}

export function pauseWorkshopAutomation() {
  return apiRequest('/api/novel-fetch-workshop/automation/pause', { method: 'POST', body: JSON.stringify({}) });
}

export function previewWorkshopProcess(payload) {
  return apiRequest('/api/novel-fetch-workshop/process/preview', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export function getWorkshopCurrentBatch() {
  return apiRequest('/api/novel-fetch-workshop/batches/current');
}

export function getWorkshopBatches() {
  return apiRequest('/api/novel-fetch-workshop/batches');
}

export function rerunWorkshopBatch(batchId, mode = 'all') {
  return apiRequest(`/api/novel-fetch-workshop/batches/${encodeURIComponent(batchId)}/rerun`, {
    method: 'POST',
    body: JSON.stringify({ mode })
  });
}

export function stopWorkshopTasks(ids) {
  return apiRequest('/api/novel-fetch-workshop/tasks/stop-selected', {
    method: 'POST',
    body: JSON.stringify({ ids })
  });
}

export function listWorkshopSchedules() {
  return apiRequest('/api/novel-fetch-workshop/schedules');
}

export function saveWorkshopSchedule(schedule) {
  return apiRequest('/api/novel-fetch-workshop/schedules', {
    method: 'POST',
    body: JSON.stringify(schedule)
  });
}

export function updateWorkshopSchedule(id, schedule) {
  return apiRequest(`/api/novel-fetch-workshop/schedules/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(schedule)
  });
}

export function deleteWorkshopSchedule(id) {
  return apiRequest(`/api/novel-fetch-workshop/schedules/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export function cancelWorkshopSchedule(id) {
  return apiRequest(`/api/novel-fetch-workshop/schedules/${encodeURIComponent(id)}/cancel`, { method: 'POST', body: '{}' });
}

export function getWorkshopLogs() {
  return apiRequest('/api/novel-fetch-workshop/logs');
}

// 任务列表
export function listWorkshopTasks(options = {}) {
  return apiRequest('/api/novel-fetch-workshop/tasks', options);
}

// 批量删除任务（ids: bookId 数组）
export function deleteWorkshopTasks(ids) {
  return apiRequest('/api/novel-fetch-workshop/tasks', {
    method: 'DELETE',
    body: JSON.stringify({ ids })
  });
}

// 批量重试任务
export function retryWorkshopTasks(ids) {
  return apiRequest('/api/novel-fetch-workshop/tasks/batch-retry', {
    method: 'POST',
    body: JSON.stringify({ ids })
  });
}

// 任务详情（meta + 处理后原文 + 日志）
export function getWorkshopTask(bookId) {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}`);
}

export function getWorkshopTaskSummary(bookId) {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}/summary`);
}

export function workshopTaskDiagnosticPath(bookId, kind) {
  return `/api/batch-rewrite/tasks/${encodeURIComponent(bookId)}/${kind}`;
}

export function getWorkshopTaskSensitiveLog(bookId) {
  return apiRequest(workshopTaskDiagnosticPath(bookId, 'sensitive-log'));
}

export function getWorkshopTaskRulesTrace(bookId) {
  return apiRequest(workshopTaskDiagnosticPath(bookId, 'rules-trace'));
}

export function getWorkshopTaskSiteSubmitLog(bookId) {
  return apiRequest(workshopTaskDiagnosticPath(bookId, 'site-submit-log'));
}

export function saveWorkshopBody(bookId, versionId, content, state = 'ready') {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}/bodies/${encodeURIComponent(versionId)}`, {
    method: 'PUT',
    body: JSON.stringify({ content, state })
  });
}

// 抓取单本原文
export function fetchWorkshopOriginal(bookId, maxTxt) {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}/fetch`, {
    method: 'POST',
    body: JSON.stringify({ maxTxt })
  });
}

// 从原始备份恢复原文
export function restoreWorkshopOriginal(bookId) {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}/restore-original`, {
    method: 'POST'
  });
}

// 生成 AI 改文版本（数量 1~20）
export function generateWorkshopAi(bookId, count) {
  return apiRequest(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}/generate-ai`, {
    method: 'POST',
    body: JSON.stringify({ count })
  });
}

export function updateWorkshopAiCount(ids, count) {
  return apiRequest('/api/novel-fetch-workshop/tasks/batch-ai-count', {
    method: 'POST',
    body: JSON.stringify({ ids, count })
  });
}

export function applyWorkshopRules(ids, scope = 'both', sensitiveAiEnabled = false) {
  return apiRequest('/api/novel-fetch-workshop/tasks/apply-rules', {
    method: 'POST',
    body: JSON.stringify({ ids, scope, sensitive_ai_enabled: sensitiveAiEnabled })
  });
}

export function reprocessWorkshopSensitive(ids, sensitiveAiEnabled = true) {
  return apiRequest('/api/novel-fetch-workshop/tasks/reprocess-sensitive', {
    method: 'POST',
    body: JSON.stringify({ ids, sensitive_ai_enabled: sensitiveAiEnabled })
  });
}

// 读取工作台配置（appConfig / platforms / styles / aiConfig）
export function getWorkshopConfig() {
  return apiRequest('/api/novel-fetch-workshop/config');
}

// 保存工作台配置（一期仅持久化 appConfig）
export function saveWorkshopConfig(patch) {
  return apiRequest('/api/novel-fetch-workshop/config', {
    method: 'POST',
    body: JSON.stringify(patch)
  });
}

// 生成规则建议，不自动写入配置
export function suggestWorkshopRules(text, ruleType, goal) {
  return apiRequest('/api/novel-fetch-workshop/rules/suggest', {
    method: 'POST',
    body: JSON.stringify({ text, ruleType, goal })
  });
}

// 预览规则排版结果与分阶段 trace
export function previewWorkshopRules(text, scope = 'ai', options = {}) {
  return apiRequest('/api/novel-fetch-workshop/rules/preview', {
    method: 'POST',
    body: JSON.stringify({ text, scope, ...options })
  });
}

// AI 接口连通性测试
export function testWorkshopAi(purpose) {
  return apiRequest('/api/novel-fetch-workshop/ai/test', {
    method: 'POST',
    body: JSON.stringify({ purpose })
  });
}

export function getWorkshopKnowledgeSummary() {
  return apiRequest('/api/novel-fetch-workshop/knowledge/summary');
}

export function getWorkshopKnowledge(kind) {
  return apiRequest(`/api/novel-fetch-workshop/knowledge/${encodeURIComponent(kind)}`);
}

export function saveWorkshopKnowledge(kind, item, options = {}) {
  return apiRequest(`/api/novel-fetch-workshop/knowledge/${encodeURIComponent(kind)}`, {
    method: 'POST',
    body: JSON.stringify({ item, ...options })
  });
}

export function deleteWorkshopKnowledge(kind, id, options = {}) {
  const query = options.configRevision === undefined ? '' : `?configRevision=${encodeURIComponent(options.configRevision)}`;
  return apiRequest(`/api/novel-fetch-workshop/knowledge/${encodeURIComponent(kind)}/${encodeURIComponent(id)}${query}`, {
    method: 'DELETE'
  });
}

export function optimizeWorkshopKnowledge(kind, id, options = {}) {
  return apiRequest(`/api/novel-fetch-workshop/knowledge/${encodeURIComponent(kind)}/${encodeURIComponent(id)}/optimize`, {
    method: 'POST',
    body: JSON.stringify(options)
  });
}

export function analyzeWorkshopOpening(text, options = {}) {
  return apiRequest('/api/novel-fetch-workshop/opening/analyze', {
    method: 'POST',
    body: JSON.stringify({ text, style: options.style || '', extra: options.extra || '' })
  });
}

export function saveWorkshopOpening(item) {
  return apiRequest('/api/novel-fetch-workshop/opening/save', {
    method: 'POST',
    body: JSON.stringify({ item })
  });
}

export function normalizeWorkshopOpening() {
  return apiRequest('/api/novel-fetch-workshop/opening/normalize', {
    method: 'POST'
  });
}
