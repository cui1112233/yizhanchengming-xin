import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import { measureH3VideoLines } from './h3LineAudio.js';
import { downloadProductionContent } from './batchFactoryContentDownload.js';

const here = path.dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const { normalizeSavedWebSubmit } = require(path.join(here, '../../../../../lib/novel-fetch-workshop/121-web-submit-service.js'));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '../shuihuo-production.css'), 'utf8');
const engineSource = fs.readFileSync(path.join(here, 'BatchFactoryEngineSettingsDrawer.jsx'), 'utf8');
const bookSettingsSource = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
const reasoningSource = fs.readFileSync(path.join(here, 'BatchFactoryAiReasoningModal.jsx'), 'utf8');
const presetLibrarySource = fs.readFileSync(path.join(here, '../../../admin/pages/PresetLibraryPage.jsx'), 'utf8');
const configRegionSource = fs.readFileSync(path.join(here, 'batchFactoryBookConfigRegions.js'), 'utf8');

test('downloads the current production content as a UTF-8 TXT file', async () => {
  const clicks = [];
  const urls = [];
  const link = { click: () => clicks.push(link), remove: () => {} };
  const ok = downloadProductionContent({
    title: '野藤盛开', content: '可下载的生产正文',
    documentRef: { createElement: () => link, body: { appendChild: () => {} } },
    urlApi: {
      createObjectURL: blob => { urls.push(blob); return 'blob:book'; },
      revokeObjectURL: url => urls.push(url)
    }
  });
  assert.equal(ok, true);
  assert.equal(link.download, '野藤盛开-生产正文.txt');
  assert.equal(await urls[0].text(), '可下载的生产正文');
  assert.equal(clicks.length, 1);
  assert.equal(urls[1], 'blob:book');
});

test('book content exposes production-content edit and TXT download actions', () => {
  assert.match(source, /编辑正文/);
  assert.match(source, /下载 TXT/);
  assert.match(source, /downloadProductionContent\(/);
  assert.match(source, /openContentEditor\(book\)/);
});

test('TXT download uses the complete captured source instead of the production line preview', () => {
  assert.match(source, /const productionText = batchFactoryCleanSourceText\(book\.sourceText\);/);
});

test('book detail renders captured source instead of the short production draft', () => {
  assert.match(source, /const \[viewingBook, setViewingBookRaw\] = useState\(null\);/);
  assert.match(source, /setViewingBookRaw\(book \? \{ \.\.\.book, workingFrontContent: '' \} : null\);/);
});

test('keeps the workbench editor on the selected lines while reserving full source editing for the detail view', () => {
  assert.match(source, /setEditingContentValue\(runtimeResolveBookProductionText\(currentBook\)\);/);
  assert.match(source, /function openSourceEditor\(book\)/);
  assert.match(source, /openSourceEditor\(viewingBook\)/);
});

test('keeps people and scene presets inside each book row instead of the global toolbar', () => {
  const toolbar = source.match(/<div className="shuihuo-workbench-toolbar"[\s\S]*?<\/div>\n      <div className="shuihuo-workbench-export">/)?.[0] || '';
  assert.equal(toolbar.includes('人物场景预设'), false);
	assert.match(source, /管理当前书资产/);
});

test('offers existing batches explicit immediate and Beijing-time automation actions', () => {
  const scheduleDialog = source.match(/<Modal title="自动生产"[\s\S]*?<\/Modal>/)?.[0] || '';
  assert.match(scheduleDialog, /<b>自动化预设<\/b>/);
  assert.match(scheduleDialog, /<b>执行模式<\/b>/);
  assert.match(scheduleDialog, /<b>自动启动时间（北京时间 UTC\+8）<\/b>/);
  assert.match(scheduleDialog, />立即执行<\/Button>/);
  assert.match(scheduleDialog, />保存定时任务<\/Button>/);
  assert.match(scheduleDialog, /同时处理书籍/);
  assert.match(scheduleDialog, /全自动生成并提交（成片完成后上传）/);
  assert.doesNotMatch(scheduleDialog, /预设名称/);
  assert.doesNotMatch(scheduleDialog, /saveCurrentAutomationPreset/);
  assert.match(source, /请先选择自动化预设/);
  assert.match(source, /parseBeijingDatetimeLocal/);
  assert.match(source, /下次重试：北京时间/);
});

test('task logs explain that giant batches are waiting for source text before production begins', () => {
  const logs = source.match(/function BatchLogs\([\s\S]*?\n}\n\nfunction /)?.[0] || '';
  assert.match(logs, /等待正文读取/);
  assert.match(logs, /巨量素材/);
});

test('keeps H3 director cards in the workbench until final VIDEO compilation', () => {
  assert.match(source, /resolvePrecompiledVideoWorkspace/);
  assert.match(source, /resolvePrecompiledStoryboardFrame/);
  assert.match(source, /batchFactoryPrecompiledStoryboard/);
  assert.match(source, /待编译的 H3 分镜视频/);
});

test('turns a director compile click without an H3 card into first-time generation', () => {
  const action = source.match(/async function runBookStageAction\([\s\S]*?\n  async function retryLastFailedStage/)?.[0] || '';
  assert.match(action, /const requestedMode = stage === 'director' && mode === 'compile' \? 'missing' : mode;/);
  assert.match(action, /mode: requestedMode,/);
});

test('renders an uncompiled H3 card as readable storyboard text with a complete action bar', () => {
  const start = source.indexOf('if (!selectedVideo && selectedPrecompiledFrame) {');
  const end = source.indexOf('\n  }\n  return <div className="batch-factory-prompt-modal-stack">', start);
  const precompiled = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(source, /function formatH3DirectorCardPrompt/);
  assert.match(precompiled, /value=\{formatH3DirectorCardPrompt\(/);
  assert.doesNotMatch(precompiled, /JSON\.stringify\(card, null, 2\)/);
  assert.match(precompiled, />保存<\/Button>/);
  assert.match(precompiled, />编译最终提示词<\/Button>/);
  assert.match(precompiled, />编辑<\/Button>/);
  assert.match(precompiled, />生成视频<\/Button>/);
  assert.match(precompiled, />查看 H3 Trace<\/Button>/);
  assert.match(precompiled, />重试<\/Button>/);
});

test('keeps an H3 final-template compile request valid without leaking fallback prose into the template', () => {
  assert.match(source, /output_constraints: videoPromptTemplate \? '按冻结导演数据和所选最终模板生成当前 VIDEO 提示词。' :/);
});

test('makes a saved unified configuration authoritative for every book', () => {
  const toolbar = source.match(/<div className="shuihuo-workbench-toolbar"[\s\S]*?<\/div>\n      <div className="shuihuo-workbench-export">/)?.[0] || '';
  assert.match(toolbar, />统一配置<\/Button>/);
  assert.doesNotMatch(toolbar, />引擎配置<\/Button>/);
  assert.doesNotMatch(toolbar, />AI 推理<\/Button>/);
  assert.match(source, /BatchFactoryUnifiedSettingsModal/);
  assert.match(source, /const UNIFIED_CONFIGURATION_KEYS = \[/);
  assert.match(source, /UNIFIED_CONFIGURATION_KEYS = \[[\s\S]*?'openingEnabled', 'openingCount'[\s\S]*?\]/);
  assert.match(source, /async function syncUnifiedSettingsToBooks\(currentBatch\)/);
  assert.match(source, /restoreKeys: UNIFIED_CONFIGURATION_KEYS/);
  assert.match(source, /await syncUnifiedSettingsToBooks\(savedBatch\)/);
});

test('recompiles existing H3 VIDEO prompts after a unified configuration save without regenerating director cards', () => {
  assert.match(source, /async function recompileUnifiedH3Prompts\(currentBatch\)/);
  assert.match(source, /await recompileUnifiedH3Prompts\(syncedBatch\)/);
  assert.match(source, /compileBookH3Videos\(book, \{ interactive: false, allowAudioSynthesis: true \}\)/);
});

test('lets a book review a viral candidate before replacing working content', () => {
  assert.match(source, /生成爆款候选/);
  assert.match(source, /替换为当前生产内容/);
  assert.match(source, /取消候选/);
  assert.match(source, /衍生开篇/);
  assert.match(source, /derivedOpeningPresetId/);
  assert.match(source, /promptPresetId: derivedOpeningPresetId/);
  assert.match(source, /重试生成爆款候选/);
  assert.match(source, /当前不能生成爆款候选/);
  assert.match(source, /rewriteWorkingFront/);
});

test('treats a missing saved viral candidate as an empty review state', () => {
  assert.match(source, /getDraft\([\s\S]*suppressGlobalError: true/);
});

test('keeps existing video prompts usable while smart unified analysis is pending', () => {
  assert.match(source, /smartUnifiedPending/);
  assert.match(source, /智能统一待分析/);
  assert.match(source, /资产提取时与资产同次获取/);
  assert.match(source, /不影响本分镜的提示词使用/);
});

test('exposes an explicit regenerate-copy action once a book already has director output', () => {
  assert.match(source, /runBookStageAction\(book, 'assets', 'force'\)/);
  assert.match(source, /runBookStageAction\(book, 'director', 'force'\)/);
  assert.match(source, /重新执行智能统一分析/);
});

test('scopes a single-book loading state to that book instead of showing every book as executing', () => {
  assert.match(source, /function stageActionKey\(stage, bookId\)/);
  assert.match(source, /setActionBusy\(stageActionKey\(stage, book\.id\)\)/);
  assert.match(source, /actionBusy === stageActionKey\('director', book\.id\)/);
  assert.doesNotMatch(source, /actionBusy === 'stage-director'/);
});

test('reloads book and durable stage status after a failed stage', () => {
  assert.match(source, /setActionBusy\(''\);\n\s*message\.error\(error\?\.message \|\| '当前小说阶段执行失败'\);\n\s*void Promise\.all\(\[refreshBatch\(\), loadRuntimeStatus\(\{ quiet: true \}\)\]\)\.catch\(\(\) => \{\}\);/);
  assert.match(source, /message\.error\(error\?\.message \|\| '当前小说阶段执行失败'\);/);
});

test('retries a single-book settings save once against the latest revision', () => {
  assert.match(bookSettingsSource, /getBatch, listSystemPresetCatalog, saveBookOverride/);
  assert.match(bookSettingsSource, /async function saveOverrideWithLatestRevision/);
  assert.match(bookSettingsSource, /Number\(error\?\.status\) !== 409/);
  assert.match(bookSettingsSource, /await getBatch\(batch\.id\)/);
  assert.match(bookSettingsSource, /expectedRevision: Number\(latestBook\.revision \|\| 0\)/);
});

test('uses a centered card modal with a 10s default storyboard duration switch', () => {
	assert.match(engineSource, /<Modal title="引擎配置" open=\{open\} onCancel=\{onClose\} width=\{980\}/);
	assert.match(engineSource, /分镜时长/);
	assert.match(engineSource, /value: 10, label: '10s'/);
	assert.match(engineSource, /value: 15, label: '15s'/);
	assert.doesNotMatch(engineSource, /VIDEO 时长策略/);
	assert.match(stylesheet, /batch-factory-engine-grid/);
	assert.match(stylesheet, /batch-factory-audio-option/);
});

test('anchors the media bottom sheet against its visible workspace when dragging', () => {
  assert.match(source, /function sheetGeometry\(\)/);
  assert.match(source, /parent\.getBoundingClientRect\(\)\.bottom/);
  assert.match(source, /rawTop: rect\.top - styleOffset/);
  assert.match(source, /visibleHeightForSheetAnchor/);
  assert.match(source, /sheetGeometry\(\)\?\.appliedOffset/);
  assert.match(source, /drag\.lastOffset/);
});

test('keeps the engine modal render-safe when another workbench control opens', () => {
	assert.match(engineSource, /import \{ Alert, Button, Input, Modal,/);
});

test('stores 121 material reuse and horizontal flip in publish settings and shows their effective values before confirmation', () => {
  assert.match(engineSource, /materialReuse/);
  assert.match(engineSource, /horizontalFlip/);
  assert.match(engineSource, /不复用/);
  assert.match(engineSource, /不翻转/);
  assert.match(source, /本次素材使用：/);
  assert.match(source, /本次水平翻转：/);
});

test('keeps an unverified video management system session readable in the upload dialog', () => {
  assert.match(source, /className={`batch-factory-publish-session-tag \${publishSessionReady \? 'is-ready' : 'is-unverified'}`}/);
  assert.match(source, /121 后台账号尚未登录或未验证/);
  assert.match(stylesheet, /\.batch-factory-publish-session-tag\.is-unverified/);
  assert.match(stylesheet, /background:\s*#4a1718\s*!important/);
});

test('prefills upload organization from effective settings without saving a dialog override', () => {
  assert.match(source, /const defaultOrganizationID =/);
  assert.match(source, /setOrganizationID\(defaultOrganizationID\)/);
  assert.match(source, /organization: organizationID/);
  assert.doesNotMatch(source, /saveBatchSettings\(.*organizationID/);
});

test('selects Batch Factory image, text and video models from enabled Personal Center models', () => {
  assert.match(engineSource, /import \{ listAvailableModels \} from '\.\.\/\.\.\/\.\.\/shared\/api\/modelCatalog'/);
  assert.match(engineSource, /listAvailableModels\('image'\)/);
  assert.match(engineSource, /listAvailableModels\('text'\)/);
  assert.match(engineSource, /listAvailableModels\('video'\)/);
  assert.match(engineSource, /href="\/api-config"/);
  assert.match(engineSource, /前往个人中心配置模型/);
  assert.doesNotMatch(engineSource, /<Input value=\{form\.imageModelId/);
  assert.doesNotMatch(engineSource, /<Input value=\{form\.textModelId/);
  assert.doesNotMatch(engineSource, /const videoModelOptions/);
});

test('uses the Shuihuo preset modal layout with an account model backed image generation chain', () => {
  assert.match(source, /shuihuo-preset-toolbar/);
  assert.match(source, /shuihuo-preset-layout/);
  assert.match(source, /label: '人物'/);
  assert.match(source, /label: '场景'/);
  assert.match(source, /label: '道具'/);
  assert.doesNotMatch(source, /label: 'AI角色'/);
  assert.match(source, /engineSettings\?\.imageModelId/);
  assert.match(source, /generateBookAssetImages/);
  assert.match(source, /AI生成（已选/);
  assert.match(source, /shuihuo-asset-tabs/);
  assert.match(source, /shuihuo-image-library-grid/);
  assert.doesNotMatch(source, /图片模型生成未接入/);
  assert.doesNotMatch(source, /AI 生成暂不可用/);
});

test('keeps the asset-prompt selector from overlapping the asset action toolbar', () => {
  assert.match(stylesheet, /\.shuihuo-preset-toolbar \{ display:flex; flex-wrap:wrap;/);
  assert.match(stylesheet, /\.batch-factory-preset-toolbar > \.ant-select \{ flex:0 1 260px; min-width:230px;/);
  assert.doesNotMatch(stylesheet, /\.shuihuo-preset-toolbar \{ display: grid; grid-template-columns: 150px 160px 104px/);
});

test('offers only complete asset schemes in the asset prompt dropdown', () => {
  assert.match(source, /listSystemPresetCatalog\('batch-factory'\)/);
  assert.match(source, /item\?\.slot === 'script\.asset-extraction'/);
  assert.doesNotMatch(source, /item\?\.slot === 'batch\.character-meta'/);
  assert.match(source, /onAssetPromptChange=\{async selection =>/);
  assert.match(source, /extraction: selection/);
});

test('opens the asset workspace directly from the book asset setting and uses the engine text model', () => {
  assert.match(source, /region\.key === 'assets' \? setAssetBook\(book\) : setConfigTarget\(\{ book, region: region\.key \}\)/);
  assert.match(source, /engineSettings\?\.textModelId/);
  assert.match(source, /runBookStageAction\(assetBook, 'assets', 'missing'/);
  assert.doesNotMatch(source, /onTextModelChange/);
});

test('regenerates assets through the asset-only stage instead of rebuilding storyboard and VIDEO output', () => {
  assert.match(source, /onRegenerate=\{async textModelId => \{ await refreshAssetPresetSnapshot\(assetBook\); await runBookStageAction\(assetBook, 'assets', 'force', '', textModelId\); \}\}/);
  assert.doesNotMatch(source, /onRegenerate=\{textModelId => runBookStageAction\(assetBook, 'director', 'force', '', textModelId\)\}/);
});

test('uses the engine image model for asset images instead of rendering a second selector', () => {
  assert.match(source, /engineSettings\?\.imageModelId/);
  assert.match(source, /modelId: imageModelId/);
  assert.doesNotMatch(source, /placeholder="选择图片模型"/);
});

test('refreshes the open book asset modal after an override save so consecutive model selections use the latest revision', () => {
  assert.match(source, /const refreshed = books\.find\(book => book\.id === assetBook\.id\)/);
  assert.match(source, /\[batch\?\.id, assetBook\?\.id, books\]/);
});

test('uses the engine image aspect ratio for asset images independently from video settings', () => {
  assert.match(source, /engineSettings\?\.imageAspectRatio \|\| engineSettings\?\.aspectRatio/);
  assert.match(source, /imageAspectRatio: imageAspectRatio \|\| '9:16'/);
  assert.doesNotMatch(source, /aria-label="当前书画幅"/);
});

test('keeps AI reasoning constraints in the same per-book override model as other V11 modules', () => {
  assert.match(reasoningSource, /constraints:\{\.\.\.scope\(n,true\)/);
  assert.match(reasoningSource, /<Range value=\{value\}/);
});

test('defaults the unified asset rule to the script extraction preset that includes props', () => {
  assert.match(reasoningSource, /script-extract-assets/);
  assert.match(reasoningSource, /人物场景道具提取/);
  assert.match(reasoningSource, /catalog\.script\.filter\(x=>x\.slot==='script\.asset-extraction'\)/);
});

test('keeps video prompts in their own published selector, with the composed storyboard meta as one option', () => {
  for (const slot of ['script.asset-extraction', 'batch.video-meta', 'batch.visual-meta']) {
    assert.match(reasoningSource, new RegExp(slot.replace('.', '\\.')));
  }
	assert.match(reasoningSource, /<Rule required title="视频提示词"/);
	assert.match(reasoningSource, /<Field required label="视频提示词"/);
  assert.match(reasoningSource, /分镜元提示词（自动组合）.*其中一个选项/);
  assert.doesNotMatch(reasoningSource, /导演与改编/);
  assert.doesNotMatch(reasoningSource, /视频风格前缀/);
  assert.doesNotMatch(reasoningSource, /\['character','人物提示词'/);
});

test('keeps H3 visual limits user-selected instead of silently enabling them with a video preset', () => {
  assert.doesNotMatch(reasoningSource, /const h3Policy=pick\(constraints,'script-constraint-restriction-h3-visual-policy'\)/);
  assert.doesNotMatch(reasoningSource, /用户随后仍可在“画面限制”里关闭或替换它/);
});

test('uses upload success as the header completion rate and exposes every other state share', () => {
  assert.match(source, /batchFactoryBatchProgress/);
  assert.match(source, /完成率 \$\{batchProgress\.completionPercent\}%/);
  assert.match(source, /上传成功/);
  assert.match(source, /异常/);
  assert.match(source, /执行中/);
  assert.match(source, /待上传/);
  assert.doesNotMatch(source, /aria-label=\{`原文就绪 \$\{progress\}%`\}/);
  assert.match(stylesheet, /batch-factory-completion-track/);
});

test('task logs distinguish scheduled auto-publish from automation that stops for manual upload', () => {
  assert.match(source, /automationStatus\?\.autoPublish === true/);
  assert.match(source, /视频管理系统回读确认/);
  assert.match(source, /自动生产默认停在待上传/);
});

test('renders every novel row with prompt and clip-library cells without a separate storyboard asset column', () => {
  assert.doesNotMatch(source, /shuihuo-preset-cell batch-factory-book-preset-cell/);
  assert.match(source, /shuihuo-prompt-cell batch-factory-book-prompt-cell/);
  assert.match(source, /shuihuo-library-cell batch-factory-book-library-cell/);
	assert.doesNotMatch(source, /label: '分镜资产'/);
  assert.match(source, /画面提示词/);
	assert.match(source, /function InlineMediaLibrary/);
});

test('puts selected storyboard assets into the storyboard prompt as @ mentions', () => {
  assert.match(source, /function storyboardAssetMentions\(book, video\)/);
  assert.match(source, /map\(asset => `@\$\{assetName\(asset\)\}`\)/);
  assert.match(source, /function withStoryboardAssetMentions\(prompt, book, video\)/);
  assert.match(source, /withStoryboardAssetMentions\(compiledPrompt \|\| displayPrompt \|\| rawVideoPrompt, book, video\)/);
  assert.match(source, /资产：\$\{mentions\}/);
});

test('renders storyboard @asset mentions as compact visual pills', () => {
  assert.match(source, /function renderStoryboardPromptWithAssetMentions\(prompt, book, video\)/);
  assert.match(source, /batch-factory-asset-mention/);
  assert.match(source, /renderStoryboardPromptWithAssetMentions\(activePrompt, book, video\)/);
});

test('binds the current storyboard to selectable asset cards whose grey state is saved as a video override', () => {
  assert.match(source, /function InlineStoryboardAssets/);
  assert.match(source, /assetSelection/);
  assert.match(source, /excludedAssetIds/);
  assert.match(source, /当前分镜的图片和视频生成/);
  assert.match(source, /saveVideoOverride\(batchId, book\.id, selectedVideo\.id/);
	assert.match(source, /batch-factory-storyboard-asset-card/);
	assert.match(source, /'is-muted'/);
});

test('storyboard asset cards support normal, starred and excluded states', () => {
	assert.match(source, /starredCharacterNames/);
	assert.match(source, /onDoubleClick/);
	assert.match(source, /is-starred/);
	assert.match(source, /saveBookOverride\(batchId, book\.id/);
});

test('uses the prompt cell as one book-level entry without an extra storyboard map', () => {
  assert.doesNotMatch(source, /batch-factory-storyboard-video-map/);
  assert.match(source, /function InlineBookPrompts/);
  assert.match(source, /batch-factory-prompt-entry-card/);
  assert.match(source, /batchFactoryBookState\(book, \{ productionStatus, mergeStatus, stageSummary \}\)/);
  assert.match(source, /productionStatus=\{productionStatus\} mergeStatus=\{mergeStatus\}/);
});

test('moves through the same storyboard and VIDEO with visible previous and next controls', () => {
  assert.match(source, /上一分镜/);
  assert.match(source, /下一分镜/);
  assert.match(source, /function StoryboardVideoNavigator/);
});

test('keeps visual and video prompts as separate per-video saved fields', () => {
  assert.match(source, /saveVideoOverride/);
  assert.match(source, /visualPrompt/);
  assert.match(source, /videoPrompt/);
  assert.match(source, /画面提示词/);
  assert.match(source, /视频提示词/);
});

test('shows the submitted H3 prompt while keeping its editable director copy separate', () => {
	assert.match(source, /getFinalPrompt\(batchId, book\.id, video\.id, \{ silent: true, suppressGlobalError: true \}\)/);
	assert.match(source, /setDisplayPrompt\(String\(response\?\.finalPrompt\?\.displayPrompt/);
	assert.match(source, /setCompiledPrompt\(String\(response\?\.finalPrompt\?\.compiledPrompt/);
	assert.match(source, /editing \? videoPrompt : submittedVideoPrompt/);
	assert.doesNotMatch(source, /displayPrompt \|\| response\?\.finalPrompt\?\.compiledPrompt/);
	assert.match(source, /const activeLabel = promptKind === 'visual' \? '画面提示词' : '分镜视频提示词';/);
});

test('defers compiled prompt reads until the user opens the prompt editor', () => {
  assert.match(source, /function useCompiledVideoPrompt\(batchId, book, video, settingsRevision = 0, \{ enabled = true \} = \{\}\)/);
  assert.match(source, /if \(!enabled \|\| !batchId \|\| !book\?\.id \|\| !video\?\.id\)/);

  const inlineStart = source.indexOf('function InlineBookPrompts(');
  const inlineEnd = source.indexOf('\nfunction StoryboardVideoNavigator(', inlineStart);
  const inline = inlineStart >= 0 && inlineEnd > inlineStart ? source.slice(inlineStart, inlineEnd) : '';
  assert.match(inline, /useCompiledVideoPrompt\(batchId, book, video, settingsRevision, \{ enabled: false \}\)/);

  assert.match(source, /useCompiledVideoPrompt\(batchId, book, selectedVideo, settingsRevision\)/);
});

test('opens a modal from the whole stacked-card prompt entry and edits only the selected storyboard', () => {
  assert.match(source, /className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell batch-factory-prompt-entry-card"/);
  assert.match(source, /className="batch-factory-prompt-entry-content"/);
  assert.doesNotMatch(source, /点击卡片编辑、重生、生成与查看候选版本/);
  assert.match(source, /className="batch-factory-prompt-modal-stack"/);
  assert.match(source, /aria-label="上一分镜"/);
  assert.match(source, /aria-label="下一分镜"/);
  assert.match(source, /onManage\?\.\(video.id\)/);
  assert.match(source, /initialVideoId=\{promptVideoId\}/);
});

test('uses the selected video media version as the durable primary merge choice', () => {
  assert.match(source, /primaryMediaTaskId/);
  assert.match(source, /切换为分镜主版本/);
  assert.match(source, /片段库与最终合成/);
});

test('keeps 1.0x as the default final merge while allowing another speed version to be chosen', () => {
  assert.match(source, /function defaultBookMerge\(jobs\)/);
  assert.match(source, /Math\.abs\(mergeJobSpeed\(job\) - 1\) < 0\.001/);
  assert.match(source, /使用成片 \$\{mergeJobSpeed\(job\)\.toFixed\(1\)\}x/);
  assert.match(source, /mergeJobs=\{bookMergeJobs\(mergeStatus, mediaBook\.id\)\}/);
	assert.match(source, /全书上传默认使用 1\.0x 合成成片/);
});

test('shows terminal merge failures and per-video merge progress in the clip library', () => {
  assert.match(source, /mergeJob\.progressPhase === 'downloading'/);
  assert.match(source, /已处理 \$\{Number\(mergeJob\.progressCurrent \|\| 0\)\} \/ \$\{Number\(mergeJob\.progressTotal \|\| videos\.length\)\} 个分镜/);
  assert.match(source, /合成已结束，等待重试/);
});

test('shows actual merge queue position and active worker phase in task logs', () => {
  const logs = source.match(/function BatchLogs\([\s\S]*?\n}\n\nfunction /)?.[0] || '';
  assert.match(logs, /合成队列 · 前方/);
  assert.match(logs, /当前工位 · \$\{merge\.progressPhase/);
	assert.match(logs, /\+ \(activeMerge \? 1 : 0\)/);
	assert.match(logs, /const automationBookDetail = book =>/);
	assert.match(logs, /String\(book\?\.stage \|\| ''\)\.toLowerCase\(\) !== 'merge'/);
	assert.match(logs, /automationBookDetail\(book\) \|\| batchFactoryVisibleError\(book\.message\)/);
});

test('renders saved book-city and novel-fetch metadata without exposing internal platform IDs', () => {
  assert.match(source, /function bookPlatformName\(book/);
  assert.match(source, /metadata\.platformName/);
	assert.match(source, /getWorkshopPlatforms/);
	assert.match(source, /batchFactoryPlatformOptions/);
  assert.match(source, /platformNames\[String\(book\?\.platform \|\| ''\)\]/);
  assert.match(source, /<span>标签<\/span>/);
  assert.match(source, /<span>推荐理由<\/span>/);
  assert.match(source, /<span>视频<\/span>/);
  assert.match(source, /function bookVideoReady/);
  assert.match(source, /视频已准备好/);
  assert.match(source, /label="书城">\{bookPlatformName\(viewingBook, platformNames\)\}/);
});

test('keeps an optional platform-label lookup from surfacing a global API failure', () => {
  assert.match(source, /getWorkshopPlatforms\(\{ suppressGlobalError: true \}\)/);
});

test('keeps persisted production and merge status readable when new work is disabled', () => {
  assert.match(source, /getBatchRuntimeSummary\(batch\.id, \{ suppressGlobalError: true \}\)/);
  assert.match(source, /setProductionStatus\(runtime\?\.production \|\| \{ jobs: \[\] \}\)/);
  assert.match(source, /setMergeStatus\(runtime\?\.merge \|\| \{ jobs: \[\] \}\)/);
});

test('delegates retry eligibility to the persisted backend record when the runtime snapshot is incomplete', () => {
  const retry = source.match(/async function retryLastFailedStage\([\s\S]*?\n  }\n  async function/)?.[0] || '';
  assert.match(retry, /const runtime = await loadRuntimeStatus\(\{ quiet: true \}\);/);
  assert.match(retry, /runtime\?\.stageSummaries\?\.\[book\.id\] \|\| stageSummaries\[book\.id\]/);
  assert.doesNotMatch(retry, /当前书没有可读取的阶段记录，暂不能自动重试。/);
  assert.doesNotMatch(retry, /if \(!summary\?\.lastFailed && !retryAsVideo\)/);
  assert.doesNotMatch(source, /getBookStageSummary/);
  assert.doesNotMatch(source, /async function loadStageSummaries/);
});

test('retries an explicitly failed storyboard even when its asynchronous provider failure has no failed stage run', () => {
  const retry = source.match(/async function retryLastFailedStage\([\s\S]*?\n  }\n  async function/)?.[0] || '';
  assert.match(retry, /batchFactoryVideoProgress\(book, runtime\?\.production \|\| productionStatus\)/);
  assert.match(retry, /failedVideo\?\.status === 'failed'/);
  assert.match(retry, /await retryBookStage\(batch\.id, book\.id/);
  assert.match(retry, /retryAsVideo/);
});

test('list-level retry resolves a persisted failed storyboard before it requires a failed stage run', () => {
  const retry = source.match(/async function retryLastFailedStage\([\s\S]*?\n  }\n  async function/)?.[0] || '';
  assert.match(retry, /const effectiveVideoID = String\(videoId \|\| \[\.\.\.videoProgress\.byVideo\.entries\(\)\]\.find\(\(\[, item\]\) => item\?\.status === 'failed'\)\?\.\[0\] \|\| ''\);/);
  assert.match(retry, /const failedVideo = effectiveVideoID \? videoProgress\.byVideo\.get\(effectiveVideoID\) : null;/);
  assert.match(retry, /videoId: effectiveVideoID,/);
});

test('does not offer a manual retry while server automation owns the book', () => {
  const retry = source.match(/async function retryLastFailedStage\([\s\S]*?\n  }\n  async function/)?.[0] || '';
  assert.match(retry, /if \(automationActive\) \{[\s\S]*?服务端自动生产正在处理该小说；会自动重试，无需手动操作。/);
  assert.match(retry, /return;[\s\S]*?setActionBusy\(stageActionKey\('retry', book\.id\)\);/);
});

test('persists named AI reasoning presets through the V11 backend and lets users rename and load them', () => {
  assert.match(reasoningSource, /AI 推理预设/);
  assert.match(reasoningSource, /createConfigVersion/);
  assert.match(reasoningSource, /renameConfigVersion/);
  assert.match(reasoningSource, /保存为新预设/);
  assert.match(reasoningSource, /重命名所选预设/);
  assert.match(reasoningSource, /载入预设/);
  assert.match(source, /presetVersions=\{configVersions\}/);
});

test('routes every new director run through the unified structured pipeline', () => {
  assert.match(source, /function usesSelectedH3VideoPreset\(settings = \{\}\)/);
  assert.match(source, /if \(presetId\) return presetId === 'batch-video-h3-director'/);
  assert.match(source, /h3: true,/);
  assert.match(source, /if \(stage === 'director'\) await compileBookH3Videos\(book\)/);
  assert.match(source, /if \(stage === 'director' && mode === 'compile' && h3DirectorCards\(book\)\.length\)/);
});

test('uses the script-generation constraint layers instead of one generic multi-select', () => {
  assert.match(reasoningSource, /基础设定（人物 \/ 场景）/);
  assert.match(reasoningSource, /画面前缀词/);
  assert.match(reasoningSource, /画面约束提示词/);
  assert.match(reasoningSource, /画面限制/);
  assert.match(reasoningSource, /负面提示词/);
  assert.match(reasoningSource, /ConstraintLayers/);
});

test('uses five persisted V11 prompt modules in one centered AI reasoning modal', () => {
  assert.match(source, /BatchFactoryAiReasoningModal/);
  assert.match(reasoningSource, /资产设置/);
  assert.match(reasoningSource, /约束设置/);
  assert.match(reasoningSource, /衍生开篇/);
  assert.match(reasoningSource, /batch\.hook-adaptation/);
  assert.match(reasoningSource, /视频设置/);
  assert.match(reasoningSource, /画面设置/);
  assert.match(reasoningSource, /aiPromptConfig/);
  assert.doesNotMatch(reasoningSource, /fixedSingleVideo/);
  assert.match(engineSource, /开启后只生产 VIDEO01；关闭后按全部分镜执行/);
});

test('keeps AI reasoning open for configuration even when the runtime is unavailable', () => {
  const toolbar = source.match(/<div className="shuihuo-workbench-toolbar"[\s\S]*?<\/div>\n      <div className="shuihuo-workbench-export">/)?.[0] || '';
  assert.match(toolbar, /setAiOpen\(true\)/);
  assert.doesNotMatch(toolbar, /disabled=\{!runCapability\.available\}/);
});

test('selects typed AI rules from Personal Center system presets without exposing their bodies', () => {
  assert.match(reasoningSource, /listSystemPresetCatalog/);
  assert.match(reasoningSource, /listSystemPresetCatalog\('script'\)/);
  assert.match(reasoningSource, /listSystemPresetCatalog\('batch-factory'\)/);
  assert.match(reasoningSource, /人物场景道具提示词/);
  assert.doesNotMatch(reasoningSource, /\['character','人物提示词'/);
  assert.doesNotMatch(reasoningSource, /\['scene','场景提示词'/);
  assert.doesNotMatch(reasoningSource, /\['prop','道具提示词'/);
  assert.match(reasoningSource, /presetVersion/);
  assert.match(reasoningSource, /presetId/);
  assert.doesNotMatch(reasoningSource, /listPersonalConstraintPrompts/);
  assert.doesNotMatch(reasoningSource, /Input\.TextArea/);
});

test('makes Batch Factory available in the Personal Center system-preset categories', () => {
  assert.match(presetLibrarySource, /label: '批量工厂', value: 'batch-factory'/);
});

test('cancels only provider-confirmed local production tasks and shows the runtime capability reason otherwise', () => {
  assert.match(source, /cancelBatchProduction/);
  assert.match(source, /setActionBusy\('cancel'\)/);
  assert.match(source, /cancelCapability = capability\(capabilities, 'production\.cancel'\)/);
  assert.match(source, /取消可取消的本地执行器任务/);
  assert.doesNotMatch(source, /V11 当前生产服务没有取消接口/);
});

test('scopes cancellation to the production operation created in this page, rather than every batch task', () => {
  assert.match(source, /activeProductionRequestID/);
  assert.match(source, /cancelBatchProduction\(batch\.id, activeProductionRequestID\)/);
  assert.match(source, /取消只会作用于该次操作/);
});

test('places five regional single-book configuration buttons between novel content and presets', () => {
  assert.match(source, /BATCH_FACTORY_TABLE_COLUMNS/);
  assert.match(source, /label: '单书配置'/);
  assert.match(source, /label: '预设'/);
  assert.match(source, /BOOK_CONFIG_REGIONS\.map/);
  assert.match(source, /setConfigTarget\(\{ book, region: region\.key \}\)/);
  assert.match(source, /onOpenBookAssets=\{book => setAssetBook\(book\)\}/);
  assert.match(configRegionSource, /引擎配置/);
  assert.match(configRegionSource, /资产设置/);
  assert.match(configRegionSource, /约束设置/);
  assert.match(configRegionSource, /视频设置/);
  assert.match(configRegionSource, /画面设置/);
  assert.match(configRegionSource, /bookAssetSummary/);
  const actions = source.match(/<div className="shuihuo-workbench-cell batch-factory-actions">[\s\S]*?<\/div>/)?.[0] || '';
  assert.doesNotMatch(actions, /setConfigTarget/);
});

test('runs and retries stages per book through the durable V11 stage routes', () => {
  assert.match(source, /runBookStage\(batch\.id, book\.id, stage/);
  assert.match(source, /retryBookStage\(batch\.id, book\.id/);
  assert.match(source, /runBookStageAction\(book, 'assets', 'force'\)/);
  assert.match(source, /runBookStageAction\(book, 'director', 'force'\)/);
  assert.match(source, />生成图片</);
  assert.match(source, />生成视频</);
  assert.match(source, />重试失败步骤</);
});

test('does not present unavailable director or video production stages as clickable actions', () => {
  const actions = source.match(/<div className="shuihuo-workbench-cell batch-factory-actions">[\s\S]*?<\/div>\n        <\/article>/)?.[0] || '';
  assert.match(actions, /runCapability\.available/);
  assert.match(actions, /productionCapability\.available/);
  assert.match(actions, /productionCapability\.reason/);
});

test('does not present unavailable video generation as runnable from the prompt modal', () => {
  assert.match(source, /productionAvailable=\{productionCapability\.available\}/);
  assert.match(source, /productionReason=\{productionCapability\.reason\}/);
  assert.match(source, /disabled=\{!productionAvailable \|\| regenerating\}/);
  assert.match(source, /onGenerateVideo=\{videoId => runBookStageAction\(promptBook, 'video', 'missing', videoId\)\}/);
});

test('groups per-book operations into the requested acquisition and recovery controls', () => {
  const actions = source.match(/<div className="shuihuo-workbench-cell batch-factory-actions">[\s\S]*?<\/div>\n        <\/article>/)?.[0] || '';
  assert.match(actions, /batch-factory-action-group is-utility/);
  assert.match(actions, /batch-factory-action-group is-production/);
  assert.match(actions, /batch-factory-action-group is-recovery/);
  assert.match(actions, /batch-factory-action-button-grid/);
  assert.match(actions, /资产获取/);
  assert.match(actions, /画面获取/);
  assert.match(actions, /视频获取/);
  assert.match(actions, />提取</);
  assert.match(actions, /生成图片/);
  assert.match(actions, /生成视频/);
  assert.match(actions, /重试/);
});

test('keeps production status inside 查看资料 instead of the operation column', () => {
  const actions = source.match(/<div className="shuihuo-workbench-cell batch-factory-actions">[\s\S]*?<\/div>\n        <\/article>/)?.[0] || '';
  assert.match(source, /getBookStageSummary/);
  assert.match(source, /batchFactoryBookTimeline/);
  assert.doesNotMatch(actions, /batch-factory-production-timeline/);
  assert.match(actions, />查看资料</);
  assert.match(source, /流程异常与生成状态/);
  assert.match(source, /小说正文/);
  assert.match(source, /hasActiveProduction/);
  assert.match(source, /viewingBook \|\| mediaBook \|\| promptBook \|\| hasActiveProduction/);
  assert.match(source, /const tick = \(\) => \{ loadRuntimeStatus\(\{ quiet: true \}\); \}/);
});

test('keeps failed runtime status diagnostics actionable in the task log', () => {
  assert.match(source, /const endpoint = String\(error\?\.source \|\| ''\)\.trim\(\)/);
  assert.match(source, /const status = Number\.isInteger\(error\?\.status\) \? `HTTP \$\{error\.status\}` : ''/);
  assert.match(source, /状态接口：\$\{endpoint\}/);
});

test('lets every production-table column after the serial number resize from its header edge', () => {
  assert.match(source, /const \[columnWidths, setColumnWidths\] = useState\(null\)/);
  assert.match(source, /function startColumnResize\(event, index, preventDefault = true\)/);
  assert.match(source, /调整\$\{column\.label\}列宽/);
  assert.match(source, /onMouseDown=\{event => startColumnResize\(event, index\)\}/);
  assert.match(source, /onPointerDown=\{event => startColumnResize\(event, index\)\}/);
  assert.match(source, /onDragStart=\{event => startColumnResize\(event, index, false\)\}/);
  assert.match(source, /onDrag=\{moveColumnResize\}/);
  assert.match(source, /onDragEnd=\{finishColumnResize\}/);
  assert.match(source, /setPointerCapture\?\./);
  assert.match(source, /map\(Number\.parseFloat\)/);
  assert.match(source, /columnWidths\.every\(Number\.isFinite\)/);
  assert.match(source, /gridTemplateColumns: columnWidths/);
  assert.match(source, /<i aria-hidden="true" \/>/);
  assert.match(stylesheet, /\.batch-factory-workbench-table \.shuihuo-workbench-head > div \{[^}]*position: relative/);
  assert.match(stylesheet, /\.batch-factory-column-resize-handle \{[^}]*position: absolute/);
});

test('offers explicit regenerate and retry controls in assets, prompts and video versions', () => {
  assert.match(source, /重新生成资产/);
  assert.match(source, /重试资产/);
  assert.match(source, /重新生成文案/);
  assert.match(source, /重新生成画面提示词/);
  assert.match(source, />重试</);
  assert.match(source, /重新生成视频/);
  assert.match(source, /重试视频/);
  assert.match(source, /不改变当前分镜主版本/);
});

test('keeps one modal prompt mode through horizontal storyboard navigation', () => {
  assert.match(source, /const \[promptKind, setPromptKind\] = useState\('video'\)/);
  assert.match(source, /const hasVisualPrompt = Boolean\(String\(visualPrompt \|\| ''\)\.trim\(\)\)/);
  assert.match(source, /请生成视频提示词再查看/);
  assert.match(source, /切换分镜时会保持当前查看类型/);
});

test('renders stacked visual and video prompt controls with an explicit copy regeneration action', () => {
  assert.match(source, /batch-factory-prompt-modal-tabs/);
  assert.match(source, /promptKind === 'visual' \? '重新生成画面提示词' : '重新生成文案'/);
  assert.match(source, /重新执行智能统一分析/);
  assert.match(source, /生成图片/);
  assert.match(source, /生成视频/);
});

test('refreshes the open prompt modal after a save so another edit uses its latest revision', () => {
  assert.match(source, /if \(!promptBook\) return;/);
  assert.match(source, /setPromptBook\(refreshed\)/);
  assert.match(source, /\[selectedVideoId, selectedVideo\?\.revision\]/);
});

test('uses the whole prompt cell as a stacked-card entry and opens its editor only in a modal', () => {
  assert.match(source, /className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell batch-factory-prompt-entry-card"/);
  assert.match(source, /className="batch-factory-prompt-entry-content"/);
  assert.doesNotMatch(source, /点击卡片编辑、重生、生成与查看候选版本/);
  assert.match(source, /onClick=\{\(\) => onManage\?\.\(video.id\)\}/);
  assert.match(source, /className="batch-factory-prompt-modal-stack"/);
  assert.match(source, /className="batch-factory-prompt-entry-nav"/);
  assert.match(source, /className=\{`batch-factory-prompt-entry-tabs is-\$\{promptKind\}`\}/);
  assert.doesNotMatch(source, /batch-factory-prompt-expanded/);
});


test('opens the production-content editor from the whole novel-content cell', () => {
  const contentCell = source.match(/batch-factory-book-content[\s\S]*?\<\/div\>/)?.[0] || '';
  assert.match(contentCell, /onClick=\{\(\) => openContentEditor\(book\)\}/);
  assert.match(source, /async function openContentEditor\(book\)/);
});


test('shows one selected storyboard VIDEO and only expands real per-video candidates from a right rail', () => {
  assert.match(source, /function InlineMediaLibrary/);
  assert.match(source, /const candidates = versions\.filter/);
  assert.match(source, /className=\{`batch-factory-inline-media/);
  assert.match(source, /当前分镜候选版本，悬停展开/);
  assert.match(source, /PassiveMediaPoster/);
  assert.match(source, /已生成 · 点击查看/);
  assert.doesNotMatch(source, /batch-factory-inline-media-stage/);
  assert.match(stylesheet, /batch-factory-inline-media\.has-candidates:hover/);
  assert.match(stylesheet, /batch-factory-inline-media-rail/);
  assert.match(stylesheet, /batch-factory-inline-media\.is-landscape/);
  assert.match(stylesheet, /aspect-ratio:16 \/ 9/);
  assert.match(stylesheet, /grid-template-rows:minmax\(0,1fr\) 30px/);
});

test('opens the clip library safely while its media-version index is still unavailable', () => {
  assert.match(source, /const mediaVersions = versionsByVideo instanceof Map \? versionsByVideo : new Map\(\)/);
  assert.match(source, /mediaVersions\.get\(selectedVideo\.id\)/);
});

test('keeps clip-library MP4 loading to the selected player and leaves other storyboards as passive posters', () => {
  const mediaPanel = source.match(/function MediaVersionPanel\([\s\S]*?\n}\nfunction BatchLogs/)?.[0] || '';
  const inlineLibrary = source.match(/function InlineMediaLibrary\([\s\S]*?\n}\n\nfunction MediaVersionPanel/)?.[0] || '';
  const candidateRenderer = mediaPanel.match(/const candidateCard[\s\S]*?\n  };/)?.[0] || '';
  const storyboardRenderer = mediaPanel.match(/const storyboardCard[\s\S]*?\n  };/)?.[0] || '';
  assert.match(mediaPanel, /className="batch-factory-media-primary-video"[^>]*preload="none"/);
  assert.match(mediaPanel, /<PassiveMediaPoster label=\{/);
  assert.doesNotMatch(storyboardRenderer, /<video/);
  assert.doesNotMatch(candidateRenderer, /<video/);
  assert.doesNotMatch(inlineLibrary, /<video/);
  assert.match(stylesheet, /\.batch-factory-media-poster \{/);
});

test('starts the selected storyboard only after the user clicks its passive card', () => {
  const mediaPanel = source.match(/function MediaVersionPanel\([\s\S]*?\n}\nfunction BatchLogs/)?.[0] || '';
  assert.match(mediaPanel, /const \[playRequested, setPlayRequested\] = useState\(false\)/);
  assert.match(mediaPanel, /setSelectedVideoId\(video\.id\); setPlayRequested\(true\)/);
  assert.match(mediaPanel, /primaryPlayerRef\.current\.play\(\)/);
  assert.match(mediaPanel, /autoPlay=\{playRequested\}/);
});

test('keeps final merge output scoped to each book and opens it from a compact 合 control', () => {
  assert.match(source, /submitBookMerge/);
  assert.match(source, /batch-factory-final-merge-trigger/);
  assert.match(source, /片段库与最终合成/);
  assert.match(source, /按各分镜主版本合成/);
  assert.match(stylesheet, /batch-factory-inline-media-rail \{[^}]*right:-10px/);
});

test('validates the real 121 website session and synchronizes website profiles and styles in publish settings', () => {
  assert.match(engineSource, /checkWebSubmitEnvironment/);
  assert.match(engineSource, /testWebSubmitVisible/);
  assert.match(engineSource, /syncWebSubmitConfigs/);
  assert.match(engineSource, /syncWebSubmitStyles/);
  assert.doesNotMatch(engineSource, /disabled title="V11 尚未提供 121 配置档同步接口"/);
  assert.match(engineSource, /websiteProfiles/);
  assert.match(engineSource, /websiteStyleCatalog/);
});


test('shows target, requested, and actual video durations in task status', () => {
  assert.match(source, /targetDurationSeconds/);
  assert.match(source, /requestedDurationSeconds/);
  assert.match(source, /actualDurationSeconds/);
});


test('keeps audio planning and audio merge independent, with merge speed derived from actual durations', () => {
  assert.match(engineSource, /audioPlanningEnabled/);
  assert.match(engineSource, /audioMergeEnabled/);
  assert.match(engineSource, /audioDurationSeconds/);
  assert.match(source, /mergePayload/);
});


test('does not rely on blank source metadata: it exposes a durable editor for style, gender, and tags', () => {
  assert.match(source, /updateBookMetadata/);
  assert.match(source, /编辑列表信息/);
});

test('repairs legacy empty books in place using their saved platform and Book ID', () => {
  assert.match(source, /fetchBookOriginal/);
  assert.match(source, /获取正文/);
  assert.match(source, /原文尚未获取/);
  assert.match(source, /已获取并写入当前小说正文/);
});

test('shows cleaned original text and the giant-material read state in the 小说正文 cell', () => {
  assert.match(source, /const previewText = batchFactoryPreviewText\(book\.sourceText, book\);/);
  assert.match(source, /<BatchFactoryGiantMaterialPendingProgress book=\{book\} batchId=\{batch\?\.id\} visible/);
  assert.doesNotMatch(source, /BatchFactoryGiantMaterialPendingProgress book=\{book\} batchId=\{batch\?\.id\} visible=\{false\}/);
});

test('keeps Batch Factory flush with the workbench while balancing wide-screen columns', () => {
  assert.match(stylesheet, /Keep the production table flush with its container/);
  assert.match(stylesheet, /width: 100%;/);
  assert.match(stylesheet, /min-width: 1665px/);
  assert.match(stylesheet, /grid-template-columns: 56px minmax\(300px, 1\.05fr\) minmax\(176px, \.70fr\) minmax\(205px, \.75fr\) minmax\(330px, 1fr\) minmax\(330px, 1fr\) minmax\(268px, \.80fr\)/);
  assert.doesNotMatch(stylesheet, /min-width: 1920px; grid-template-columns: 56px minmax\(355px, 1\.45fr\)/);
});

test('uses the continuous screenplay-workbench row rhythm for every batch book', () => {
  assert.match(stylesheet, /Batch Factory uses the screenplay workbench rhythm/);
  assert.match(stylesheet, /\.batch-factory-book-row \{\n  margin-top: 0/);
  assert.match(stylesheet, /min-height: 304px/);
  assert.match(stylesheet, /grid-template-rows: repeat\(5, minmax\(0, 1fr\)\)/);
  assert.match(stylesheet, /\.batch-factory-action-group\.is-recovery \{\n  margin-top: auto/);
});

test('uses the full clip-library width for a landscape primary VIDEO and overlays candidates', () => {
  assert.match(stylesheet, /The horizontal main VIDEO is the whole clip-library frame/);
  assert.match(stylesheet, /\.batch-factory-inline-media\.is-landscape \.batch-factory-inline-media-primary \{\n  justify-self: stretch;\n  width: 100%;/);
  assert.match(stylesheet, /\.batch-factory-inline-media\.has-candidates:hover \{\n  padding-right: 16px;/);
});

test('makes one batch clip-library cell a single full-width primary VIDEO frame', () => {
  assert.match(stylesheet, /\.batch-factory-book-library-cell \{\n  display: block !important;/);
  assert.match(stylesheet, /\.batch-factory-book-library-cell \.batch-factory-inline-media \{\n  width: 100%;/);
});

test('validates the exact 121 tttadmin session before the publish mapping is usable', () => {
  assert.match(engineSource, /two\.121w\.com\/tttadmin\/index\.php/);
  assert.match(engineSource, /121 后台登录会话/);
});

test('accepts the shared 121 session-check name returned by the verification API', () => {
  assert.match(engineSource, /SESSION_CHECK_NAMES\s*=\s*\[\s*'视频管理系统登录会话',\s*'121 后台登录会话'\s*\]/);
  assert.match(engineSource, /SESSION_CHECK_NAMES\.includes\(check\.name\)/);
  assert.match(source, /'目标站登录会话'/);
});

test('shares the 121 account session while retaining upload mappings inside the current batch', () => {
  assert.match(engineSource, /getWebSubmitConfig/);
  assert.match(engineSource, /saveWebSubmitConfig/);
  assert.match(engineSource, /登录并验证/);
  assert.match(engineSource, /更换账号/);
  assert.match(engineSource, /publishSessionReady/);
  assert.match(source, /checkWebSubmitEnvironment/);
  assert.match(source, /testWebSubmitVisible/);
  assert.match(source, /onOpenPublish/);
});

test('renders the shared 121 account as a compact publish card', () => {
  assert.match(stylesheet, /\.batch-factory-121-account-card \{/);
});

test('imports the account-card status tag so opening the workbench cannot crash', () => {
  assert.match(engineSource, /Tabs, Tag, message/);
});

const webSubmitServiceSource = fs.readFileSync(path.join(here, '../../../../../lib/novel-fetch-workshop/121-web-submit-service.js'), 'utf8');

test('keeps Batch Factory website profile and style synchronization out of novel-fetch settings', () => {
  assert.match(engineSource, /syncWebSubmitConfigs\(\{ persist: false \}\)/);
  assert.match(engineSource, /syncWebSubmitStyles\(\{ persist: false \}\)/);
  assert.match(webSubmitServiceSource, /async function syncConfigs\(owner, \{ persist = true \} = \{\}\)/);
  assert.match(webSubmitServiceSource, /async function syncStyles\(owner, \{ persist = true \} = \{\}\)/);
});

test('does not clear existing 121 submission settings when Batch Factory only refreshes the shared login', () => {
  assert.match(webSubmitServiceSource, /enabled: incoming\.enabled === undefined \? existing\.enabled === true : incoming\.enabled === true/);
});


test('preserves existing novel-fetch submit settings while a shared 121 login refreshes', () => {
  const refreshed = normalizeSavedWebSubmit({ enabled: true, skip_submitted: false, selected_profile: '小说获取默认档' }, { username: '121-account' });
  assert.equal(refreshed.enabled, true);
  assert.equal(refreshed.selected_profile, '小说获取默认档');
  assert.equal(refreshed.skip_submitted, false);
});

test('lets each book measure temporary TTS audio for storyboard planning without persisting the audio file', () => {
  const settingsSource = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
  assert.match(settingsSource, /textToSpeech\(\{ input, \.\.\.tts \}\)/);
  assert.match(settingsSource, /const config = await getConfig\(\)/);
  assert.match(settingsSource, /readAudioDuration\(blob\)/);
  assert.match(settingsSource, /audioPlanningEnabled: true, audioDurationSeconds/);
  assert.match(settingsSource, /下载配音/);
  assert.match(settingsSource, /音频不会保存或加入最终合并/);
});

test('prepares only temporary audio duration when unified audio planning is enabled', () => {
  const start = source.indexOf('async function prepareAudioPlanningForBatch(currentBatch)');
  const end = source.indexOf('async function compileBookH3Videos(book,', start);
  const preparation = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(source, /const audioPlanningWasEnabled = batch\?\.settingsState\?\.patch\?\.audioPlanningEnabled === true;/);
  assert.match(source, /audioPlanningWasEnabled === false && normalized\.audioPlanningEnabled === true/);
  assert.match(source, /await prepareAudioPlanningForBatch\(refreshedBatch\)/);
  assert.match(preparation, /await ensureBookAudioDuration\(book, \{ force: true, quiet: true, batchSnapshot: currentBatch \}\)/);
  assert.doesNotMatch(preparation, /runBookStage|runBatchDirector|submitBatchProduction|runProduction/);
});

test('keeps a saved audio duration usable when its fingerprint is stale and refresh fails', () => {
  const start = source.indexOf('async function ensureBookAudioDuration(book');
  const end = source.indexOf('async function prepareAudioPlanningForBatch', start);
  const ensureDuration = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(ensureDuration, /catch \(error\) \{[\s\S]*?currentSeconds > 0[\s\S]*?return currentSeconds;/);
  assert.match(ensureDuration, /继续使用已保存的配音时长/);
});

test('offers image generation inside the asset edit dialog after saving the current prompt', () => {
  const editor = source.match(/function AssetEditor[\s\S]*?function storyboardAssetDefaults/)?.[0] || '';
  assert.match(editor, /async function generateEditedAssetImage/);
  assert.match(editor, /await saveAsset\(\{ close: false \}\)/);
  assert.match(editor, /generateImages\(\[savedAsset\.id\]\)/);
  assert.match(editor, /生成图片/);
});

test('refreshes a historical book asset-preset snapshot before asset generation or regeneration', () => {
  assert.match(source, /async function refreshAssetPresetSnapshot\(book\)/);
  assert.match(source, /saveBookOverride\(batch\.id, book\.id, \{[\s\S]*?aiPromptConfig/);
  const assetModal = source.match(/<Modal title=\{assetBook[\s\S]*?<\/Modal>/)?.[0] || '';
  assert.match(assetModal, /refreshAssetPresetSnapshot\(assetBook\)/);
});

test('derives each 121 platform from its book and classifies remaining metadata before submission', () => {
  assert.match(source, /publishSettings: \{ \.\.\.\(batchPatch\.publishSettings \|\| \{\}\), \.\.\.\(bookPatch\.publishSettings \|\| \{\}\) \}/);
  assert.match(source, /const effectiveSettings = effectiveBookSettings\(batch, book\)/);
  assert.match(source, /function selectedPublishProfile\(settings = \{\}\)/);
  assert.match(source, /function sourcePlatformIdForUpload\(book = \{\}\)/);
  assert.match(source, /function completePublishMapping\(_settings = \{\}, book = \{\}\)/);
  assert.match(source, /逐本识别发布信息/);
  assert.match(source, /提交前会按每本书正文自动识别男女频、121 风格和标签/);
  assert.match(source, /sourcePlatformIdForUpload\(book\)/);
  assert.match(source, /const missingPublishMapping = targetBooks\.filter/);
  assert.match(source, /单书发布覆盖优先、批量统一设置兜底/);
});

test('keeps AI publication classification progress and failures visible in the book detail modal', () => {
  assert.match(source, /function classificationDetail\(metadata = \{\}\)/);
  assert.match(source, /正在识别男女频、风格和标签/);
  assert.match(source, /AI 判断未完成/);
  assert.match(source, /classifyError/);
});

test('uses grouped cards for the unified Batch Factory engine settings', () => {
  assert.match(engineSource, /function EngineCard\(/);
  assert.match(engineSource, /className="batch-factory-engine-card/);
  assert.match(engineSource, /title="视频"/);
  assert.match(engineSource, /title="跟随音频"/);
  assert.match(engineSource, /title="图片与文本"/);
  assert.match(engineSource, /title="版本对应配置档"/);
});

test('uses the same engine cards for unified settings and single-book overrides', () => {
  const bookSettings = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
  assert.match(engineSource, /className="batch-factory-engine-card/);
  assert.match(bookSettings, /function ConfigCard\(/);
  assert.match(bookSettings, /className="batch-factory-engine-card/);
  assert.match(bookSettings, /ConfigCard title="视频"/);
  assert.match(bookSettings, /ConfigCard title="发布统一"/);
});


test('opens single-book configuration in the same centered card modal as unified engine settings', () => {
  const bookSettings = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
  assert.match(bookSettings, /import \{ Alert, Button, Divider, Input, InputNumber, Modal,/);
  assert.match(bookSettings, /<Modal title=\{book\?\.title/);
  assert.match(bookSettings, /onCancel=\{onClose\} width=\{980\}/);
  assert.doesNotMatch(bookSettings, /<Drawer title=\{book\?\.title/);
});

test('121 upload UI gates missing decompression, exposes progress, and requires an explicit reupload', () => {
  const bookSettings = fs.readFileSync(path.join(here, 'BatchFactoryBookSettingsModal.jsx'), 'utf8');
  assert.match(source, /请添加解压/);
  assert.match(source, /websiteSubmitProgress/);
  assert.match(source, /重新上传/);
  assert.doesNotMatch(bookSettings, /网站风格类型/);
  assert.doesNotMatch(engineSource, /网站风格类型/);
});

test('keeps the novel-list status synchronized with the matching single-book stage summary', () => {
  assert.match(source, /function NovelMetadata\(\{ books, createdAt, selectedBookIds, onSelectionChange, onViewBook, platformNames, productionStatus, mergeStatus, stageSummaries \}\)/);
  assert.match(source, /stageSummary: stageSummaries\?\.\[book\.id\]/);
  assert.match(source, /stageSummaries=\{stageSummaries\}/);
});

test('adjusts each book final merge speed from the black clip-library frame', () => {
  assert.match(source, /onOpenMerge=\{\(\) => setMediaBook\(book\)\}/);
  assert.match(source, /<b>合成视频<\/b>/);
  assert.match(source, /<InputNumber min=\{0\.5\} max=\{4\} step=\{0\.1\} value=\{mergeSpeed\}/);
  assert.match(source, /mergePayload\('bf11-book-merge', book, mergeSpeed\)/);
  assert.match(source, /合并跟随音频已开启，倍率按成片与音频真实时长自动计算/);
});

test('keeps the final-merge speed control compact inside the final-video dialog', () => {
  assert.match(stylesheet, /\.batch-factory-library-merge \{/);
  assert.match(stylesheet, /justify-content:space-between/);
});

test('keeps shot versions, final merge, and the selected 121 upload video in one clip-library dialog', () => {
  assert.match(source, /primaryUploadSource/);
  assert.match(source, /设为上传主视频/);
  assert.match(source, /当前上传主视频/);
  assert.match(source, /onOpenMerge=\{\(\) => setMediaBook\(book\)\}/);
  assert.match(source, /mergeJob=\{latestBookMerge\(mergeStatus, mediaBook\.id\)\}/);
  assert.match(source, /submitBookTo121\(batch\.id, book\.id, \{ organization: organizationID, category: 'NEW_BOOK', startTime: localStartTime\(\), textModelId: effectiveBookSettings\(batch, book\)\.textModelId, reupload \}\)/);
  assert.match(source, /function localStartTime\(\)/);
  assert.match(source, /请选择 121 组织归属/);
  assert.match(source, /txt: `\$\{book\.bookId\}\.txt`/);
  assert.match(source, /mp4: `\$\{book\.bookId\}\.mp4`/);
  assert.match(source, /const hasBookUploadSource = book =>/);
  assert.match(source, /targetBooks\.every\(hasBookUploadSource\)/);
  assert.doesNotMatch(source, /<Modal title=\{mergeBook \?/);
});

test('checks the newest book-stage summary before retrying, so a completed retry is not reported as a revision conflict', () => {
  assert.match(source, /const summary = resultData\(await getBookStageSummary\(batch\.id, book\.id\), 'summary'\)/);
  assert.match(source, /if \(!summary\?\.lastFailed\) \{[\s\S]*?当前小说没有可重试的失败步骤/);
  assert.match(source, /await retryBookStage\(batch\.id, book\.id/);
});

test('unified settings retries only a revision-only conflict and refreshes when settings changed', () => {
  assert.match(source, /async function saveSettings\(patch\)/);
  assert.match(source, /const latestResult = await getBatch\(batch\.id\)/);
  assert.match(source, /sameSettingsPatch\(latestBatch\?\.settingsState\?\.patch, batch\?\.settingsState\?\.patch\)/);
  assert.match(source, /配置已更新，已刷新当前统一配置；请核对后重新保存/);
  assert.match(source, /expectedRevision: Number\(latestBatch\?\.settingsState\?\.revision \|\| 0\)/);
});

test('keeps prompt editing in the prompt dialog and exposes media actions only in the clip library', () => {
  const mediaPanel = source.match(/function MediaVersionPanel\([\s\S]*?\n}\n\nfunction BatchLogs/)?.[0] || '';
  assert.match(source, /删除候选版本/);
  assert.match(source, /切换为分镜主版本/);
  assert.match(source, /当前上传主视频/);
  assert.doesNotMatch(mediaPanel, /编辑提示词/);
});

test('labels every clip-library storyboard option without relying on an undefined render helper', () => {
  assert.match(source, /function storyboardVideoLabel\(video, index\)/);
  assert.match(source, /label: storyboardVideoLabel\(video, videoIndex\)/);
});

test('keeps the inline clip-library preview safe before runtime media status hydrates', () => {
  assert.match(source, /function InlineMediaLibrary[\s\S]*?const mediaVersions = versionsByVideo instanceof Map \? versionsByVideo : new Map\(\)/);
  assert.match(source, /const versions = video \? \(mediaVersions\.get\(video\.id\) \|\| \[\]\) : \[\];/);
});

test('initializes the selected primary video before an autoplay effect reads it', () => {
  const panel = source.match(/function MediaVersionPanel\([\s\S]*?\n}\nfunction BatchLogs/)?.[0] || '';
  const primaryIndex = panel.indexOf('const primary = primaryMediaVersion(selectedVideo, versions);');
  const autoplayEffectIndex = panel.indexOf('if (!playRequested || !primary?.mediaUrl || !primaryPlayerRef.current) return;');
  assert.ok(primaryIndex >= 0, 'clip library must define its selected primary media');
  assert.ok(autoplayEffectIndex > primaryIndex, 'autoplay must not read primary before it is initialized');
});

test('compiles existing H3 director cards instead of regenerating them from video extraction', () => {
  assert.match(source, /画面获取[\s\S]{0,500}runBookStageAction\(book, 'visual', 'force'\)/);
  const start = source.indexOf('async function runBookStageAction(book, stage');
  const end = source.indexOf('async function retryLastFailedStage', start);
  const action = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(source, /视频获取[\s\S]{0,500}runBookStageAction\(book, 'director', 'compile'\)/);
  assert.match(action, /if \(stage === 'director' && mode === 'compile' && h3DirectorCards\(book\)\.length\) \{[\s\S]*?await compileBookH3Videos\(book\);[\s\S]*?return;/);
  assert.match(action, /await runBookStage\(batch\.id, book\.id, stage/);
  assert.match(source, /onRegenerateVisual=\{\(\) => runBookStageAction\(promptBook, 'visual', 'force'\)\}/);
});

test('names the prompt workbench as storyboard cards carrying video prompts', () => {
  assert.match(source, /title=\{promptBook \? `分镜卡（视频提示词） · \$\{promptBook\.title\}` : '分镜卡（视频提示词）'\}/);
});

test('lets a VIDEO card inspect the actual submitted prompt and compile trace', () => {
  assert.match(source, /getH3Trace/);
  assert.match(source, /查看 H3 Trace/);
  assert.match(source, /实际提交视频模型的 H3 最终 Prompt/);
  assert.match(source, /compiled_prompt/);
  assert.match(source, /compile_trace/);
});

test('shows persisted H3 director cards before final VIDEO compilation instead of looking empty', () => {
  assert.match(source, /function h3DirectorCards\(book\)/);
  assert.match(source, /H3 导演卡已提取/);
  assert.match(source, /等待真实配音时长编译最终 VIDEO/);
  assert.match(source, /h3Card\.source_text/);
});

test('uses the same media-library structure for an H3 director card before VIDEO compilation', () => {
  const start = source.indexOf('function MediaVersionPanel(');
  const end = source.indexOf('export function BatchFactoryNovelList(', start);
  const panel = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(panel, /batch-factory-media-pickstation-workspace/);
  assert.match(panel, /当前分镜暂不可播放视频/);
  assert.match(panel, /precompiledWorkspace\.frames\.map/);
  assert.match(panel, /打开分镜提示词/);
  assert.match(panel, /查看导演提示词/);
});

test('closes the media library before opening its director prompt', () => {
  assert.match(source, /const \[pendingMediaPrompt, setPendingMediaPrompt\] = useState\(null\);/);
  assert.match(source, /if \(!pendingMediaPrompt \|\| mediaBook\) return;/);
  assert.match(source, /setPromptVideoId\(pendingMediaPrompt\.frameKey \|\| ''\);/);
  assert.match(source, /setPromptBook\(nextBook\);/);
  assert.match(source, /setPendingMediaPrompt\(null\);/);
  assert.match(source, /onOpenDirector=\{frameKey => \{ setPendingMediaPrompt\(\{ bookId: mediaBook\.id, frameKey: frameKey \|\| '' \}\); setMediaBook\(null\); \}\}/);
});

test('keeps the awaiting-H3 media rail above its collapsed control bar', () => {
  assert.match(source, /batch-factory-media-bottom-sheet is-precompiled-collapsed/);
  assert.match(stylesheet, /\.batch-factory-media-bottom-sheet\.is-precompiled-collapsed\s*\{\s*height:56px;/);
});

test('compiles a successful H3 director run from real TTS audio before exposing editable VIDEO cards', () => {
  assert.match(source, /measureH3Audio/);
  assert.match(source, /compileH3Video/);
  assert.match(source, /async function compileBookH3Videos/);
  assert.match(source, /await compileBookH3Videos\(book\)/);
  assert.match(source, /audio_measurement/);
  assert.match(source, /director_revision_id/);
});

test('uses semantic 10 or 15 second compilation without TTS when follow-audio is disabled', () => {
  const start = source.indexOf('async function compileBookH3Videos(book,');
  const end = source.indexOf('async function refreshAfterBookSettingsSaved()', start);
  const compile = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(compile, /if \(settings\.audioPlanningEnabled === true\)\s*\{[\s\S]*?measureH3VideoLines/);
  assert.match(compile, /allow_semantic_timeline:\s*settings\.audioPlanningEnabled !== true/);
});

test('returns only a persisted line measurement identity for subsequent H3 compilation', async () => {
  const events = [];
  const measured = await measureH3VideoLines({ directorId: 'd1',
    document: { director_cards: [{ source_key: 'line_0001', source_text: '正文' }] }, tts: { voice: 'voice' },
    synthesize: async () => 'audio', encode: async audio => audio,
    measure: async payload => { events.push('persist'); assert.equal(payload.lines[0].audio_base64, 'audio'); return { audio_asset_id: 'persisted-lines' }; }
  });
  events.push('compile-ready');
  assert.equal(measured.audio_asset_id, 'persisted-lines');
  assert.deepEqual(events, ['persist', 'compile-ready']);
  await assert.rejects(measureH3VideoLines({ directorId:'d1', document:{director_cards:[{source_key:'a',source_text:'正文'}]}, tts:{}, synthesize:async()=> 'audio', encode:async x=>x, measure:async()=>{throw new Error('测量保存失败');} }), /测量保存失败/);
});

test('recompiles existing H3 director data after constraint or VIDEO preset settings change', () => {
	assert.match(source, /async function refreshAfterBookSettingsSaved/);
	assert.match(source, /\['constraints', 'video', 'media'\]\.includes\(configTarget\?\.region\)/);
	assert.match(source, /await runBookStageAction\(configTarget\.book, 'director', 'compile'\)/);
	assert.match(source, /onSaved=\{refreshAfterBookSettingsSaved\}/);
});

test('maps the selected smart-unified prefix to baseline injection', () => {
	assert.match(source, /buildBatchFactoryH3Constraints\(constraints\)/);
	assert.match(bookSettingsSource, /智能统一/);
});

test('passes only the selected V11 final template and its H3 protocol key to compilation', () => {
  assert.match(source, /function finalVideoPromptTemplate\(body\)/);
  assert.match(source, /const videoPromptTemplate = finalVideoPromptTemplate\(videoPresetBody\);/);
  assert.match(source, /key: String\(videoPreset\.presetKey \|\| videoPreset\.presetId \|\| videoPreset\.id \|\| 'h3-video-normal'\)/);
});


test('shows live provider task progress on the matching storyboard card and keeps polling while production runs', () => {
  assert.match(source, /batchFactoryVideoProgress/);
  assert.match(source, /currentProgress\?\.message/);
  assert.match(source, /productionStatus=\{productionStatus\}/);
  assert.match(source, /hasActiveProduction/);
});

test('loads protected local merge files through the authenticated media boundary', () => {
  assert.match(source, /import \{ ProductionMediaBoundary \} from '\.\.\/batch-factory-v11\/ProductionMediaBoundary';/);
  assert.match(source, /<ProductionMediaBoundary showDownload=\{false\}>/);
});

test('keeps production polling failures inside the task panel', () => {
  assert.match(source, /getBatchRuntimeSummary\(batch\.id, \{ suppressGlobalError: true \}\)/);
});

test('shows opening variant prompts from the VIDEO01 card instead of candidate versions', () => {
  assert.match(source, /openingVariants/);
  assert.match(source, /查看换开头提示词/);
  assert.match(source, /分镜一（原始）/);
  assert.match(source, /AI 未生成出该变体/);
  assert.match(source, /engineSettings\.openingEnabled === true/);
  assert.doesNotMatch(source, /查看候选版本/);
  assert.doesNotMatch(source, /onViewVideoCandidates/);
  assert.doesNotMatch(source, /onViewVisualCandidates/);
});

test('uses the VIDEO card plus button only for local MP4 upload', () => {
  const start = source.indexOf('function MediaVersionPanel(');
  const end = source.indexOf('function BatchLogs(', start);
  const mediaPanel = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(source, /uploadBookVideoMaterial/);
  assert.match(mediaPanel, /accept="video\/mp4,\.mp4"/);
  assert.match(mediaPanel, /请上传 MP4 视频/);
  assert.match(mediaPanel, /素材已上传，并已设为当前分镜主版本。/);
  assert.match(mediaPanel, /上传素材/);
  assert.doesNotMatch(mediaPanel, />重新生成视频</);
  const uploadControl = mediaPanel.match(/<span className="batch-factory-media-pickstation-add"[\s\S]*?<\/span>/)?.[0] || '';
  assert.doesNotMatch(uploadControl, /onRegenerate/);
});
