import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryCreateModal.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');

test('uses a full novel-fetch metadata preset by default', () => {
  // 输入格式/列顺序固定为默认值，不再暴露控件（AI 自动分析男女频/风格）
  assert.match(source, /const DEFAULT_COLUMN_PRESET_ID = 'full_metadata'/);
  assert.match(source, /const DEFAULT_COLUMN_ORDER = '书籍ID,书名,男女频,风格,标签,推荐理由,评级'/);
  assert.doesNotMatch(source, /自定义列顺序/);
  assert.doesNotMatch(source, /列顺序预设/);
});

test('platform group tags use theme-aware css classes, not antd preset colors', () => {
  // 暗色主题下 antd color="blue" 对比度太低看不清，改用 CSS 类控制配色
  assert.match(source, /className=\{editingPlatformId === group\.platformId \? 'platform-group-tag is-editing' : 'platform-group-tag'\}/);
  assert.doesNotMatch(source, /color=\{editingPlatformId/);
});

test('offers immediate and Beijing-time automation with a frozen concurrency limit', () => {
  assert.match(source, /开始定时/);
  assert.match(source, /立即执行/);
  assert.match(source, /listAutomationPresets/);
  assert.match(source, /automationPresetID/);
  assert.match(source, /automationRunMode/);
  assert.match(source, /选择已保存预设/);
  assert.match(source, /只生成分镜/);
  assert.match(source, /生成视频不提交/);
  assert.match(source, /自动启动时间（北京时间 UTC\+8）/);
  assert.match(source, /到设定时间才开始生产；生成、合成与上传按后续流程继续。/);
  assert.match(source, /全自动生成并提交（成片完成后上传）/);
  assert.match(source, /presetId: automationPresetID/);
  assert.match(source, /runMode: automationRunMode/);
  assert.match(source, /scheduleMode: scheduledRun \? automationScheduleMode/);
  assert.match(source, /定时执行/);
  assert.match(source, /定时提交/);
  assert.match(source, /automationConcurrency/);
  assert.match(source, /同时处理书籍/);
  assert.match(source, /parseBeijingDatetimeLocal/);
  assert.match(source, /定时任务/);
  assert.doesNotMatch(source, /<strong>自动补抓正文<\/strong>/);
  assert.doesNotMatch(source, /<strong>自动生产<\/strong>/);
  assert.doesNotMatch(source, /自动上传视频管理系统/);
  assert.doesNotMatch(source, /<strong>定时自动执行<\/strong>/);
  assert.doesNotMatch(source, /开启后，点击创建会先真实抓取/);
  assert.match(source, /batch-factory-create-toolbar/);
  assert.doesNotMatch(source, /内容范围决定本次配音、H3 导演和 VIDEO 编译/);
});

test('immediate execution defaults to the end-to-end upload automation mode', () => {
  assert.match(source, /const \[automationRunMode, setAutomationRunMode\] = useState\('full_submit'\)/);
	assert.match(source, /if \(!scheduled\) setAutomationRunMode\('full_submit'\);/);
});

test('lets a first-time batch owner save unified settings before creating the batch', () => {
  assert.match(source, /BatchFactoryEngineSettingsForm/);
  assert.match(source, /BatchFactoryAiReasoningForm/);
  assert.match(source, /BatchFactoryPublishSettingsForm/);
  assert.match(source, /BatchFactoryAutomationPresetManager/);
  assert.match(source, /const \[initialBatchSettings, setInitialBatchSettings\] = useState\(\{\}\)/);
  assert.match(source, />统一配置<\/Button>/);
  assert.match(source, /新建批量 · 统一配置/);
  assert.match(source, /<BatchFactoryAutomationPresetManager value=\{initialBatchSettings\} productionMode=\{isComicProduction \? 'comic' : 'video'\} onLoad=\{setInitialBatchSettings\} \/>/);
  assert.match(source, /initialBatchSettings: clone\(initialBatchSettings\)/);
});

test('lets a first-time batch owner switch production mode beside the unified title', () => {
  assert.match(source, /BatchFactoryProductionModeControl/);
  assert.match(source, /<span>新建批量 · 统一配置<\/span><BatchFactoryProductionModeControl/);
  assert.match(source, /productionMode=\{initialBatchSettings\.productionMode\}/);
  assert.doesNotMatch(source, /key: 'media', label: '画面与视频'/);
});

test('shows comic analysis characters separately from the 4000-character capture limit', () => {
  assert.match(source, /initialBatchSettings\.productionMode === 'comic'/);
  assert.match(source, /生产模式<Select value=\{initialBatchSettings\.productionMode \|\| 'video'\}/);
  assert.match(source, /漫画分析/);
  assert.match(source, /comicAnalysisCharacters/);
  assert.match(source, /contentCaptureCharacters/);
  assert.match(source, /comicAnalysisCharacters,/);
});

test('uses the unified button treatment for the new batch toolbar', () => {
  assert.match(source, /className="shuihuo-create-project-modal"/);
  assert.match(stylesheet, /BF_CREATE_BATCH_BUTTONS_BEGIN[\s\S]*shuihuo-create-project-modal[\s\S]*box-shadow/);
});

test('grouped submit surfaces failed book cities and counts only fetched pending books', () => {
  // I3：失败书城可见——catch 里记录书城名并 console.warn，最终 warning 带出具体书城
  assert.match(source, /const failedPlatforms = \[\]/);
  assert.match(source, /failedPlatforms\.push\(platformLabel\)/);
  assert.match(source, /console\.warn\('分组抓取失败', group\.platformId, error\)/);
  assert.match(source, /以下书城抓取失败：\$\{failedPlatforms\.join\('、'\)\}/);
  assert.match(source, /失败的书已在列表中标红，可单独重试/);
  // 全部成功但有书没正文时，沿用原来的 X/Y 文案
  assert.match(source, /\} else if \(fetchedCount < totalCount\) \{/);
  // #6：fetchedCount 只对本组 pending id 中真正拿到非空正文的计数
  assert.match(source, /let fetchedCount = 0/);
  assert.match(source, /for \(const bookId of pending\) \{\s*if \(String\(groupSources\[bookId\] \|\| ''\)\.trim\(\)\) fetchedCount \+= 1;/);
});

test('offers giant material intake through the resident executor', () => {
  assert.match(source, /GIANT_MATERIAL_PLATFORM_OPTION/);
  assert.match(source, /resolveGiantMaterialForBatch/);
  assert.match(source, /createGiantMaterialJob/);
  // 多 ID：解析全部 + 一次占位 intake + 工作区进度条（不在弹窗内等 OCR）
  assert.match(source, /解析全部/);
  assert.match(source, /buildGiantMaterialPlaceholderIntakes/);
  assert.match(source, /每行一个巨量素材 ID，数量不限/);
  assert.match(source, /parseGiantMaterialIds/);
  assert.match(source, /BatchFactoryGiantMaterialExecutorStatus/);
  assert.match(source, /onBatchUpdated\?\.\(batchId\)/);
  assert.doesNotMatch(source, /extractGiantMaterial\(/);
  // 旧单 ID 等待式流程已移除
  assert.doesNotMatch(source, /waitForGiantMaterialJob/);
});

test('highlights later giant material rows that resolve to a duplicate Book ID', () => {
  assert.match(source, /duplicateGiantMaterialBookIndexes/);
  assert.match(source, /is-duplicate/);
  assert.match(source, /重复，将跳过/);
  assert.match(stylesheet, /batch-factory-giant-item\.is-duplicate/);
  assert.match(stylesheet, /batch-factory-giant-duplicate-tag/);
});

test('giant immediate execution opens the same automation-plan dialog as other book cities', () => {
  const toolbarStart = source.search(/\{isGiantMaterial\r?\n        \? <>/);
  const toolbarEnd = source.indexOf('\n      <Button onClick={() => openAutomationDialog(\'scheduled\')}>开始定时</Button>', toolbarStart);
  const giantToolbar = toolbarStart >= 0 && toolbarEnd > toolbarStart ? source.slice(toolbarStart, toolbarEnd) : '';
  assert.match(giantToolbar, /onClick=\{\(\) => openAutomationDialog\('immediate'\)\}/);
  assert.doesNotMatch(giantToolbar, /onClick=\{\(\) => submitGiantMaterial\(\)\}/);
});

test('giant creation selects a versioned preset without requiring frozen metadata before dispatching reads', () => {
	assert.match(source, /giantAutomationPlan:/);
	assert.match(source, /expectedPresetVersion: selectedPreset\.version/);
	assert.match(source, /automationPresetSnapshot\(automationPresets, automationPresetID\)/);
	assert.doesNotMatch(source, /saveBatchSettings/);
	assert.doesNotMatch(source, /settingsState\?\.patch\?\.automationPresetSnapshot/);
	assert.doesNotMatch(source, /巨量素材自动化预设未能随创建冻结/);
});

test('lets giant batches choose a mutually exclusive original acquisition priority', () => {
  assert.match(source, /优先直接获取原文/);
  assert.match(source, /const \[giantOriginalReadStrategy, setGiantOriginalReadStrategy\] = useState\('ocr_first'\)/);
  assert.match(source, /originalReadStrategy: giantOriginalReadStrategy/);
  assert.match(source, /giantOriginalReadStrategy === 'direct_first'/);
  assert.match(source, /await fetchBookOriginal\(batchId, book\.id\)/);
  assert.match(source, /await createGiantMaterialJob\(/);
});

test('direct-first keeps a book-city failure visible and makes OCR a manual fallback', () => {
  assert.match(source, /失败后保留原因，可在工作区手动使用滚屏 OCR。/);
  assert.doesNotMatch(source, /失败后自动改用滚屏 OCR。/);
});
