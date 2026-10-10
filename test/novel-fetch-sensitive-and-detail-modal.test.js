const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { toV78Task } = require('../lib/novel-fetch-workshop/task-ops');

const root = path.join(__dirname, '..');
const read = relative => fs.readFileSync(path.join(root, relative), 'utf8');

test('敏感词重跑将本次开关和统一文本模型解析器传入敏感词处理', () => {
  const route = read('routes/batch-rewrite.js');
  assert.match(route, /function withSensitiveAiEnabled\(/);
  assert.match(route, /const config = withSensitiveAiEnabled\(object\(store\.getConfig\(\)\), req\.body\?\.sensitive_ai_enabled\);/);
  assert.match(route, /applySavedRulesToOriginal\(tasks, req\.username, id, config, store\)/);
  assert.match(route, /applySavedRulesToOriginal\(tasks, req\.username, req\.params\.id, config, store\)/);
});

test('所有任务详情浮层都有可见关闭按钮并支持点击遮罩关闭', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /function detailCloseControl\(/);
  assert.match(app, /renderSensitiveLog[\s\S]*detailCloseControl\(\)/);
  assert.match(app, /renderSiteSubmitLog[\s\S]*detailCloseControl\(\)/);
  assert.match(app, /event\.target\.closest\("#detail"\)/);
});

test('小说获取全部弹框支持遮罩退出，并在提交中锁定关闭', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /function setModalBusy\(/);
  assert.match(app, /dialog\.addEventListener\("pointerdown", event => \{/);
  assert.match(app, /event\.target === dialog && !isModalBusy\("webLoginDialog"\)/);
  assert.match(app, /setModalBusy\("webLoginDialog", true\)/);
  assert.match(app, /setModalBusy\("webLoginDialog", false\)/);
  assert.match(app, /loginButtons\.forEach\(button => \{ button\.disabled = true; \}\)/);
  assert.match(app, /loginButtons\.forEach\(button => \{ button\.disabled = false; \}\)/);
  assert.match(app, /event\.target === \$\("versionConfigCard"\) && !isModalBusy\("versionConfigCard"\)/);
  assert.match(app, /listenOptionalControl\("versionConfigCard", "pointerdown", event => \{/);
  assert.match(app, /setModalBusy\("versionConfigCard", true\)/);
  assert.match(app, /setModalBusy\("versionConfigCard", false\)/);
  assert.match(app, /if \(isModalBusy\("versionConfigCard"\)\) return;/);
});

test('任务列表详情遮罩按内容卡片判定，生成AI文案按钮直接执行', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /data-action="ai"[^>]*\$\{aiBusy \? "disabled" : ""\}[^>]*>\$\{aiBusy \? "生成中…" : "生成AI文案"\}<\/button>/);
  assert.match(app, /if \(event\.target\.closest\("#detail"\)\) return;/);
  const generateAi = app.match(/async function generateAi\(id\) \{[\s\S]*?\r?\n\}\r?\n\r?\nfunction selectedTaskIds/);
  assert.ok(generateAi, 'generateAi implementation should remain present');
  assert.match(generateAi[0], /\/generate-ai/);
  assert.doesNotMatch(generateAi[0], /showTask\(id\)/);
});

test('生成AI文案点击后显示处理中并把失败原因反馈到任务区', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /aiProcessingIds: new Set\(\)/);
  assert.match(app, /data-action="ai"[^>]*\$\{aiBusy \? "disabled" : ""\}/);
  const generateAi = app.match(/async function generateAi\(id\) \{[\s\S]*?\r?\n\}\r?\n\r?\nfunction selectedTaskIds/);
  assert.ok(generateAi, 'generateAi implementation should remain present');
  assert.match(generateAi[0], /state\.aiProcessingIds\.add\(id\)/);
  assert.match(generateAi[0], /正在生成AI文案/);
  assert.match(generateAi[0], /catch \(error\)/);
  assert.match(generateAi[0], /生成AI文案失败/);
  assert.match(generateAi[0], /state\.aiProcessingIds\.delete\(id\)/);
});

test('任务详情规范化保留阶段错误、当前版本和更新时间', () => {
  const normalized = toV78Task({
    bookId: 'book-1',
    originalErrorMessage: '上游返回 429：请求过于频繁',
    originalErrorCode: 'UPSTREAM_FETCH_ERROR',
    originalUpstreamCode: 429,
    classifyError: '分类模型超时',
    aiError: 'AI接口返回 403：额度不足',
    aiCurrentVersion: 'ai2',
    siteSubmitError: '121提交失败：会话已失效',
    createdAt: '2026-09-21T00:00:00.000Z',
    updatedAt: '2026-09-21T00:01:00.000Z'
  });
  assert.equal(normalized.original_error, '上游返回 429：请求过于频繁');
  assert.equal(normalized.original_error_code, 'UPSTREAM_FETCH_ERROR');
  assert.equal(normalized.original_upstream_code, 429);
  assert.equal(normalized.classify_error, '分类模型超时');
  assert.equal(normalized.ai_error, 'AI接口返回 403：额度不足');
  assert.equal(normalized.ai_current_version, 'ai2');
  assert.equal(normalized.site_submit_error, '121提交失败：会话已失效');
  assert.equal(normalized.created_at, '2026-09-21T00:00:00.000Z');
  assert.equal(normalized.updated_at, '2026-09-21T00:01:00.000Z');
});

test('任务详情展示阶段错误并在打开时持续刷新', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /metaItem\("AI失败原因", meta\.ai_error/);
  assert.match(app, /metaItem\("原文失败原因", meta\.original_error/);
  assert.match(app, /metaItem\("网站提交失败原因", meta\.site_submit_error/);
  assert.match(app, /metaItem\("最后更新时间", meta\.updated_at/);
  assert.match(app, /function startTaskDetailRefresh\(/);
  assert.match(app, /setInterval\(\(\) => { void refreshTaskDetail/);
  assert.match(app, /function stopTaskDetailRefresh\(/);
});

test('AI生成接口不会把失败结果包装成成功响应', () => {
  const route = read('routes/batch-rewrite.js');
  assert.match(route, /const result = await rewrite\.generateAiVersions\(/);
  assert.match(route, /const current = await tasks\.getTask\(req\.username, req\.params\.id\)/);
  assert.match(route, /result\.status === 'failed' \|\| result\.status === 'partial'/);
  assert.match(route, /res\.status\(422\)\.json\(/);
});

test('生成AI文案前端识别后端非完成状态', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  const generateAi = app.match(/async function generateAi\(id\) \{[\s\S]*?\r?\n\}\r?\n\r?\nfunction selectedTaskIds/);
  assert.ok(generateAi, 'generateAi implementation should remain present');
  assert.match(generateAi[0], /const response = await api\(/);
  assert.match(generateAi[0], /response\.status !== "done"/);
});

test('网关 HTML 错误转换为中文提示且生成失败不清空已有任务', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /function normalizeBatchErrorMessage\(/);
  assert.match(app, /网关暂时不可用（502）/);
  assert.match(app, /loadTasks\(\{ preserveOnEmpty: true/);
  assert.match(app, /incomingTasks\.length === 0/);
});

test('生成成功提示前会回读任务确认 AI 版本确实生成', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  const generateAi = app.match(/async function generateAi\(id\) \{[\s\S]*?\r?\n\}\r?\n\r?\nfunction selectedTaskIds/);
  assert.ok(generateAi, 'generateAi implementation should remain present');
  assert.match(generateAi[0], /const verification = await api\(`\/api\/tasks\/\$\{encodeURIComponent\(id\)\}`\)/);
  assert.match(generateAi[0], /AI文案生成未完成，请查看任务详情和失败原因/);
  assert.match(generateAi[0], /generatedAi\.includes\(version\)/);
});

test('AI文案列独立展示生成阶段、完成数量和失败原因', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /function aiCopyStatusText\(/);
  assert.match(app, /生成中/);
  assert.match(app, /失败：/);
  assert.match(app, /ai_error/);
  assert.match(app, /aiCopyStatusText\(aiTask, selectedAi, generatedAi\)/);
  assert.match(app, /function taskOverallStatusText\(/);
});

test('生成AI文案期间保留当前任务行并合并最新状态', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /function mergeTasksKeepingIds\(/);
  assert.match(app, /merged\.splice\(/);
  assert.match(app, /loadTasks\(\{ preserveOnEmpty: true, preserveIds: \[id\] \}\)/);
  assert.match(app, /preserveIds/);
});

test('手动生成AI接口同步写回任务总状态', () => {
  const route = read('routes/batch-rewrite.js');
  assert.match(route, /status: 'ai_processing'/);
  assert.match(route, /const status = result\.status === 'done'[\s\S]*'ai_done'/);
  assert.match(route, /result\.status === 'partial'[\s\S]*'ai_partial'/);
  assert.match(route, /rewrite\.reconcileAiTaskStatus/);
  assert.match(route, /reconciled\.status === 'done' \? 'ai_done'/);
});

test('任务列表不向用户显示英文原始状态和提交版本', () => {
  const app = read('frontend/public/batch-rewrite/app.js');
  assert.match(app, /ai_processing:\s*"AI文案处理中"/);
  assert.match(app, /original_processing:\s*"原文处理中"/);
  assert.match(app, /function displayVersionLabel\(version\)/);
  assert.match(app, /done\.map\(displayVersionLabel\)/);
  assert.match(app, /version === "original" \? "原文"/);
});

test('V78 任务桥接不会覆盖独立 AI 文案状态，也不会在刷新空响应时移除任务行', () => {
  const bridge = read('public/batch-rewrite/v78-novel-fetch-v2.js');
  assert.match(bridge, /const display = typeof aiCopyStatusText === 'function'/);
  assert.match(bridge, /aiCell\.title = String\(task\.ai_error/);
  assert.match(bridge, /options\.preserveOnEmpty && incomingTasks\.length === 0/);
  assert.match(bridge, /mergeTasksKeepingIds\(incomingTasks, options\.preserveIds \|\| \[\]\)/);
  assert.doesNotMatch(bridge, /if \(\/failed\|失败\/i\.test\(String\(task\.ai_status \|\| ''\)\)\) return `\$\{version\.toUpperCase\(\)\}失败`/);
});

test('V78 日期刷新桥接透传任务保留选项', () => {
  const date = read('public/batch-rewrite/v78-novel-fetch-v2-date.js');
  assert.match(date, /async function loadDateFilteredTasks\(originalLoadTasks, options = \{\}\)/);
  assert.match(date, /return originalLoadTasks\(options\)/);
  assert.match(date, /const originalLoadTasks = loadTasks/);
});

test('V78 日期控件使用表格实际渲染日期，并支持前后一天', () => {
  const date = read('public/batch-rewrite/v78-novel-fetch-v2-date.js');
  assert.match(date, /state\.viewMode = 'date'/);
  assert.match(date, /state\.taskDate = date/);
  assert.match(date, /taskPrevDayBtn/);
  assert.match(date, /taskNextDayBtn/);
  assert.match(date, /shiftTaskDateKey\(state\.taskDate \|\| todayDateKey\(\), -1\)/);
  assert.match(date, /shiftTaskDateKey\(state\.taskDate \|\| todayDateKey\(\), 1\)/);
});

test('V78 日期修复使用新的脚本缓存键', () => {
  const page = read('lib/novel-fetch-workshop/v2-page.js');
  assert.match(page, /DATE_CLIENT_PATH\}\?v=20260922-rendered-date-r1/);
});

test('任务详情区分历史分类模型和本次敏感词执行模型', () => {
  const route = read('routes/batch-rewrite.js');
  const app = read('frontend/dist/batch-rewrite/app.js');
  assert.match(route, /sensitive_model: task\.sensitiveModel \|\| task\.sensitive_model \|\| ''/);
  assert.match(route, /sensitive_status: task\.sensitiveStatus \|\| task\.sensitive_status \|\| ''/);
  assert.match(app, /分类模型（历史分类）/);
  assert.match(app, /敏感词模型（本次执行）/);
  assert.match(app, /meta\.sensitive_model/);
});

test('生产镜像将 V78 桥接脚本复制到前端静态目录', () => {
  const dockerfile = read('Dockerfile');
  assert.match(dockerfile, /COPY public\/batch-rewrite\/v78-novel-fetch-v2\.js \.\/frontend\/dist\/batch-rewrite\/v78-novel-fetch-v2\.js/);
  assert.match(dockerfile, /COPY public\/batch-rewrite\/v78-novel-fetch-v2-date\.js \.\/frontend\/dist\/batch-rewrite\/v78-novel-fetch-v2-date\.js/);
});
