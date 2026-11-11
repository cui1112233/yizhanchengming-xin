import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const app = fs.readFileSync(new URL('../App.jsx', import.meta.url), 'utf8');
const layout = fs.readFileSync(new URL('../../shared/layouts/UserLayout.jsx', import.meta.url), 'utf8');
const page = fs.readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const editorShell = fs.readFileSync(new URL('./aicut/EditorShell.jsx', import.meta.url), 'utf8');
const pageCSS = fs.readFileSync(new URL('./aicut-page.css', import.meta.url), 'utf8');
const agentPanel = fs.readFileSync(new URL('../components/AICutAgentPanel.jsx', import.meta.url), 'utf8');

test('AI-CUT按确认设计使用独立全屏剪辑器，不再套用全站侧栏', () => {
  assert.match(app, /['"]\/ai-cut['"]:\s*AICutPage/);
  assert.match(layout, /href:\s*['"]\/ai-cut['"][^\n]+label:\s*['"]AI-CUT['"]/);
  assert.match(app, /if \(pathname === ['"]\/ai-cut['"]\)[\s\S]{0,220}<AICutPage/);
  assert.match(page, /aicut-fullscreen-editor/);
  assert.match(page, /返回一战晟铭/);
  assert.match(page, /AI-CUT Hub · 在线版/);
  assert.match(page, /AI 创作画布/);
  assert.match(page, /专业剪辑/);
  assert.match(agentPanel, /AI-CUT Agent/);
  assert.match(agentPanel, /今天想怎么剪/);
  for (const label of ['媒体', '音频', '文本', '特效', '转场', '字幕', 'AI工具']) assert.match(page, new RegExp(label));
  for (const label of ['贴纸', '滤镜', '调节', '模板', '数字人']) assert.doesNotMatch(page, new RegExp(`\\['${label}'`));
  assert.match(agentPanel, /粗剪成片/);
  assert.match(agentPanel, /识别字幕/);
  assert.match(agentPanel, /去口误停顿/);
  assert.match(agentPanel, /配背景音乐/);
  assert.match(page, /导出/);
  assert.match(page, /时间线/);
  assert.doesNotMatch(page, /项目列表/);
  assert.doesNotMatch(page, /本地增强/);
});

test('AI-CUT颜色跟随全站主题变量且具备剪映式四区布局', () => {
  assert.match(pageCSS, /--aicut-canvas:var\(--surface-base\)/);
  assert.match(pageCSS, /--aicut-panel:var\(--surface-panel\)/);
  assert.match(pageCSS, /background:var\(--aicut-canvas\)/);
  assert.match(pageCSS, /background:var\(--aicut-panel\)/);
  assert.doesNotMatch(pageCSS, /\.aicut-page\{[^}]*background:#/);
  assert.doesNotMatch(pageCSS, /aicut-standalone-shell/);
  assert.match(pageCSS, /\.aicut-fullscreen-editor/);
  assert.match(pageCSS, /\.aicut-tool-rail/);
  assert.match(pageCSS, /\.aicut-library-panel/);
  assert.match(pageCSS, /\.aicut-inspector/);
  assert.match(pageCSS, /\.aicut-agent-drawer/);
});

test('AI-CUT专业剪辑严格使用57201确认的桌面网格和统一时间线', () => {
  assert.match(pageCSS, /grid-template-columns:74px 292px minmax\(460px,1fr\) 304px/);
  assert.match(pageCSS, /grid-template-rows:minmax\(310px,58%\) minmax\(225px,42%\)/);
  assert.match(pageCSS, /\.aicut-tool-rail\{[^}]*grid-row:1\/3/);
  assert.match(pageCSS, /\.aicut-jianying-shell>\.aicut-timeline\{[^}]*grid-column:2\/5/);
  assert.match(page, /aicut-timeline-body/);
  assert.match(page, /aicut-tracks-canvas/);
  assert.doesNotMatch(page, /if \(true\) return/);
});

test('AI-CUT全屏编辑器覆盖旧窄屏最小高度，骨架始终收进单个视口', () => {
  assert.match(pageCSS, /\.aicut-editor-shell\.aicut-reference-editor\{[^}]*min-height:0/);
  assert.match(pageCSS, /\.aicut-fullscreen-editor\{[^}]*overflow:hidden/);
});

test('AI-CUT默认媒体栏只显示已经接通的功能卡', () => {
  const capabilities = fs.readFileSync(new URL('./aicutCapabilities.js', import.meta.url), 'utf8');
  for (const text of ['网页上传', '项目素材']) assert.match(capabilities, new RegExp(text));
  for (const text of ['授权素材库', '浏览器录制', '智能搜索', 'AI生成素材']) assert.doesNotMatch(page, new RegExp(text));
  for (const tag of ['浏览器', '云服务', 'AI模型', 'Go后台']) assert.match(page, new RegExp(tag));
  assert.match(page, /搜索功能或素材/);
  assert.match(page, /全部/);
  assert.match(page, /收藏/);
  assert.match(page, /最近使用/);
});

test('AI-CUT使用57201原始顶栏和编辑器语义结构而不是AntD近似结构', () => {
  assert.match(page, /aicut-reference-top/);
  assert.match(page, /aicut-reference-brand/);
  assert.match(page, /aicut-reference-project/);
  assert.match(page, /aicut-reference-actions/);
  assert.match(editorShell, /aicut-reference-editor/);
  assert.match(page, /aicut-reference-stage/);
  assert.match(page, /aicut-reference-player/);
  assert.match(page, /aicut-reference-timeline/);
});

test('AI自动剪辑入口保留在原顶栏且草案抽屉不改变四区骨架', () => {
  assert.match(page, /aicut-reference-actions[\s\S]{0,1800}AI 自动剪辑/);
  assert.match(page, /AICutAutoEditDrawer/);
  assert.match(page, /setAutoEditOpen/);
  assert.match(editorShell, /aicut-reference-editor/);
});

test('AI-CUT播放器和真实轨道时间线保留完整可见控制', () => {
  for (const label of ['适应', '安全框', '全屏预览', '轨道菜单', '字幕 1', '特效 1', '视频 1', '音频 1']) assert.match(page, new RegExp(label));
  assert.doesNotMatch(page, /多时间线/);
  assert.match(page, /aicut-reference-labels/);
  assert.match(page, /aicut-reference-tracks/);
  assert.match(pageCSS, /height:45px/);
  assert.match(pageCSS, /height:52px/);
  assert.match(pageCSS, /aicut-reference-tracks\{min-width:100%/);
  assert.match(page, /width: `max\(100%, \$\{canvasWidth\}px\)`/);
});

test('AI-CUT时间线提供真实剃刀与联动模式并把转场显示在连接处', () => {
  assert.match(page, /setBladeMode/);
  assert.match(page, />\s*剃刀\s*</);
  assert.match(page, /setLinkedMove/);
  assert.match(page, />\s*联动\s*</);
  assert.match(page, /aicut-transition-chip/);
});

test('AI-CUT选中片段两侧提供真实裁剪手柄而不是固定按钮演示', () => {
  assert.match(page, /function beginClipTrim/);
  assert.match(page, /type: 'clip\.resize'/);
  assert.match(page, /runTimelineCommand/);
  assert.match(page, /event\.clientX - rect\.left <= 8/);
  assert.match(pageCSS, /aicut-clip\.selected:before/);
  assert.match(pageCSS, /cursor:ew-resize/);
});

test('AI-CUT专业编辑器使用57201的紧凑字号和低对比表面系统', () => {
  assert.match(pageCSS, /--acr-bg:var\(--surface-base\)/);
  assert.match(pageCSS, /--acr-panel:var\(--surface-panel\)/);
  assert.match(pageCSS, /--acr-line:var\(--border-default\)/);
  assert.match(pageCSS, /\.aicut-reference-editor\{[^}]*font-size:12px/);
  assert.match(pageCSS, /\.aicut-feature-card\{[^}]*background:var\(--acr-panel-2\)/);
  assert.match(pageCSS, /\.aicut-reference-timeline\{[^}]*background:var\(--acr-panel\)/);
  assert.match(page, /<ConfigProvider\s+theme=/);
  assert.match(pageCSS, /--acr-accent:var\(--accent-primary/);
});

test('AI-CUT页面使用真实项目和上传API', () => {
  assert.match(page, /listAICutProjects/);
  assert.match(page, /createAICutProject/);
  assert.match(page, /uploadAICutAsset/);
  assert.match(page, /saveAICutTimeline/);
  assert.doesNotMatch(page, /功能开发中/);
});

test('AI-CUT页面提供独立SRT导入并在取消选择时不改变工程', () => {
  assert.match(page, /importAICutSRT/);
  assert.match(page, /accept="\.srt"/);
  assert.match(page, /导入 SRT/);
  assert.match(page, /if \(!file\) return Upload\.LIST_IGNORE/);
  assert.match(page, /setTimeline\(result\.timeline\)/);
});

test('AI-CUT字幕纠错使用后端提示词链路并可取消', () => {
  assert.match(page, /correctAICutCaptions/);
  assert.match(page, /字幕纠错/);
  assert.match(page, /校对剧本/);
  assert.match(page, /new AbortController\(\)/);
  assert.match(page, /captionAbortRef\.current\?\.abort\(\)/);
  assert.match(page, /setTimeline\(result\.timeline\)/);
  assert.doesNotMatch(page, /你负责依据剧本校对中文视频字幕/);
});

test('AI-CUT语音转字幕使用独立ASR模型和可取消任务链路', () => {
  assert.match(page, /listAvailableModels\(['"]asr['"]\)/);
  assert.match(page, /createAICutASRJob/);
  assert.match(page, /listAICutASRJobs/);
  assert.match(page, /cancelAICutASRJob/);
  assert.match(page, /识别字幕/);
  assert.doesNotMatch(page, /Node\.js|Node 状态/);
});

test('AI-CUT时间线工具栏和BGM卡片具有真实操作处理器', () => {
  assert.match(page, /undoTimeline/);
  assert.match(page, /redoTimeline/);
  assert.match(page, /addTrack/);
  assert.match(page, /toggleTrackFlag/);
  assert.match(page, /removeTrack/);
  assert.match(page, /bladeMode/);
  assert.match(page, /setBGMOpen\(true\)/);
  assert.match(page, /加入背景音乐/);
});

test('AI-CUT保存后的撤销重做由服务端历史恢复', () => {
  assert.match(page, /getAICutTimelineHistory/);
  assert.match(page, /undoAICutTimeline/);
  assert.match(page, /redoAICutTimeline/);
  assert.match(page, /serverHistory/);
});

test('AI-CUT本地修改停止一段时间后自动保存且不弹成功提示', () => {
  assert.match(page, /saveState !== 'dirty'/);
  assert.match(page, /setTimeout\(\(\) => saveTimeline\(\{ silent: true \}\), 1500\)/);
  assert.match(page, /if \(!options\.silent\) message\.success\('时间线已保存'\)/);
});

test('AI-CUT成片导出显示持久后台队列并可取消', () => {
  assert.match(page, /createAICutExportJob/);
  assert.match(page, /listAICutExportJobs/);
  assert.match(page, /cancelAICutExportJob/);
  assert.match(page, /downloadAICutExportJob/);
  assert.match(page, /<ExportQueuePanel jobs=\{exportJobs\}/);
});

test('AI-CUT预览由时间线播放头驱动而不是仅预览选中素材', () => {
  assert.match(page, /buildAICutPreviewFrame/);
  assert.match(page, /previewTimeMs/);
  assert.match(page, /aicut-preview-seek/);
  assert.match(page, /togglePreviewPlayback/);
  assert.match(page, /PreviewMedia/);
  assert.doesNotMatch(page, /!selectedAsset \? <Empty[^\n]+选择素材进行预览/);
});

test('AI-CUT七个原版页签不再使用占位壳', () => {
  assert.match(page, /同一视频轨中相接的第二个片段/);
  assert.match(page, /导入剧本/);
  assert.match(page, /开始分析/);
  assert.match(page, /提取代表帧/);
  assert.match(page, /extractAICutVisualFrames/);
  assert.match(page, /手动字幕|添加字幕/);
  assert.match(page, /短剧AI剪辑/);
  assert.doesNotMatch(page, /该素材分类的真实链路尚未接入/);
});

test('AI-CUT去口误停顿使用真实后端操作并支持取消', () => {
  assert.match(page, /trimAICutFillers/);
  assert.match(page, /setFillerOpen\(true\)/);
  assert.match(page, /fillerAbortRef\.current\?\.abort\(\)/);
  assert.doesNotMatch(page, /title="口误停顿检测链路待接入"/);
});

test('AI-CUT四张建议卡与原版一致且点击只写入Agent输入框', () => {
  assert.match(agentPanel, /按剧本把这一集粗剪出来，删掉重复和废镜头/);
  assert.match(agentPanel, /识别全片字幕，并对照剧本纠正错字/);
  assert.match(agentPanel, /删掉口误、语气词和超过 1 秒的停顿/);
  assert.match(agentPanel, /根据情绪给高潮段落配一段合适的背景音乐/);
  assert.match(agentPanel, /chooseAgentSuggestion\(item\.text\)/);
  assert.match(agentPanel, /setPrompt\(text\)/);
  assert.match(agentPanel, /composerRef\.current\?\.focus/);
  assert.match(agentPanel, /sendAgentPrompt/);
});

test('AI-CUT Agent以浮动入口打开独立抽屉，不占用素材栏', () => {
  assert.match(page, /AICutAgentPanel/);
  assert.match(page, /onTimelineChanged/);
  assert.match(page, /aicut-agent-fab/);
  assert.match(page, /aicut-agent-drawer/);
  assert.match(page, /setAgentDrawerOpen/);
  assert.match(pageCSS, /aicut-agent-drawer/);
});

test('AI-CUT Agent不再通过五个关键词假装执行模型工具链', () => {
  assert.doesNotMatch(page, /resolveAICutAgentTool/);
  assert.doesNotMatch(page, /这条指令还没有可执行的 AI-CUT 工具/);
  assert.match(page, /AICutAgentPanel/);
});

test('AI-CUT工作台提供素材区和时间线上下拖动、收起状态', () => {
  assert.match(page, /aicut-resize-handle library-preview/);
  assert.match(page, /aicut-resize-handle upper-timeline/);
  assert.match(page, /mediaCollapsed/);
  assert.match(page, /timelineCollapsed/);
  assert.match(page, /aicut\.editorLayout\.v3/);
  assert.match(editorShell, /aicut-editor-v3/);
  assert.match(pageCSS, /--aicut-media-width/);
  assert.match(pageCSS, /--aicut-top-height/);
});

test('AI-CUT右侧属性栏和时间线只显示真实操作', () => {
  for (const label of ['基础', '速度']) assert.match(page, new RegExp(label));
  for (const label of ['轨道', '分割', '复制', '删除', '吸附']) assert.match(page, new RegExp(label));
	for (const label of ['多时间线', '需要配置对应的在线处理服务']) assert.doesNotMatch(page, new RegExp(label));
	assert.match(page, /clip\.freeze_set/);
	assert.match(page, /clip\.reverse_set/);
  assert.match(page, /clip\.keyframe_upsert/);
  assert.match(page, /aicut-inspector/);
  assert.match(page, /aicut-tool-rail/);
});

test('AI-CUT素材可拖进轨道并按落点写入时间线', () => {
  assert.match(page, /draggable/);
  assert.match(page, /application\/x-aicut-asset/);
  assert.match(page, /dropAssetOnTrack/);
  assert.match(page, /timelineDropTime/);
});

test('AI-CUT顶栏可新建项目且快捷键按钮打开真实面板', () => {
  assert.match(page, />新建项目</);
  assert.match(page, /setCreateOpen\(true\)/);
  assert.match(page, /shortcutOpen/);
  assert.match(page, /resolveAICutShortcut/);
  assert.match(page, /title="键盘快捷键"/);
});

test('播放器把混合限制在画布内并保持每个画面层透明', () => {
  assert.match(pageCSS, /\.aicut-preview-stage\{[^}]*isolation:isolate/);
  assert.match(pageCSS, /\.aicut-preview-visual\{[^}]*background:transparent/);
});
