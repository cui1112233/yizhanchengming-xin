import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
import { createAICutCapabilities } from './aicutCapabilities.js';

const page = fs.readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const inspector = fs.readFileSync(new URL('./aicut/InspectorPanel.jsx', import.meta.url), 'utf8');
const audioProcessing = fs.readFileSync(new URL('./aicutAudioProcessing.js', import.meta.url), 'utf8');
const compoundWorkspace = fs.readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('AI-CUT does not render generic placeholder controls', () => {
  for (const text of ['需要先配置对应的在线处理服务', '语义搜索还没接上', '该轨道尚未接入', '尚未配置关键帧', '尚未配置定格帧']) {
    assert.doesNotMatch(page, new RegExp(text));
  }
});

test('every catalog capability requires an executable handler', () => {
  const handlers = new Proxy({}, { get: () => () => true });
  const items = createAICutCapabilities(handlers);
  assert.ok(items.length > 0);
  for (const item of items) {
    assert.equal(typeof item.action, 'function', item.id);
    assert.ok(item.id && item.category && item.title && item.detail && item.source, item.id);
  }
});

test('unwired external features are not present in the page catalog', () => {
  for (const label of ['授权素材库', '人声分离', 'AI贴纸', '人物特效', 'AI特效', '数字人形象', '口型驱动', 'AI补帧']) {
    assert.doesNotMatch(page, new RegExp(label));
  }
});

test('字幕识别入口始终可见，缺少模型时在弹窗引导配置', () => {
  assert.match(page, /openTranscription:\s*\(\)\s*=>/);
  assert.doesNotMatch(page, /openTranscription:\s*asrModels\.length\s*\?/);
  assert.match(page, /openCaptionCorrection:\s*agentModels\.length\s*\?/);
  assert.match(page, /openSpeechCleanup:\s*agentModels\.length\s*\?/);
  assert.match(page, /\['audio', 'video'\]\.includes\(asset\.kind\)[\s\S]{0,500}>识别字幕<\/Button>/);
  assert.match(page, /!asrModels\.length[\s\S]{0,500}>\s*前往 API 配置\s*<\/Button>/);
});

test('AI自动剪辑从顶栏打开草案审核且应用后刷新持久时间线', () => {
  assert.match(page, /AICutAutoEditDrawer/);
  assert.match(page, />\s*AI 自动剪辑\s*</);
  assert.match(page, /revision=\{timeline\.revision\}/);
  assert.match(page, /assets=\{assets\}/);
  assert.match(page, /models=\{agentModels\}/);
  assert.match(page, /onApplied=\{handleAutoEditApplied\}/);
  assert.match(page, /getAICutTimelineHistory/);
  assert.doesNotMatch(page, /你是一战晟铭 AI-CUT 的剪辑方案规划器/);
});

test('核心剪辑按钮统一经过命令层且保存不会清空本地撤销历史', () => {
  assert.match(page, /import \{[\s\S]{0,220}executeAICutCommand[\s\S]{0,220}reconcileTimelineSave[\s\S]{0,220}restoreTimelineSnapshot[\s\S]{0,80}from '\.\/aicutCommands\.js'/);
  for (const type of ['clip.delete', 'clip.duplicate', 'clip.split_at_playhead', 'clip.split_at', 'clip.move', 'clip.resize', 'clip.link', 'clip.unlink']) {
    assert.match(page, new RegExp(`type: '${type.replace('.', '\\.')}[^']*'|type: '${type}'`), type);
  }
  assert.match(page, /function runTimelineCommand\(command\)/);
  assert.match(page, /reconcileTimelineSave\(timelineRef\.current, saved, capturedEditSequence, editSequenceRef\.current\)/);
  assert.doesNotMatch(page, /async function saveTimeline[\s\S]{0,1800}undoRef\.current = \[\]/);
});

test('选中视频可真实渲染合并为项目素材并支持取消', () => {
  assert.match(page, /consolidateAICutClips/);
  assert.match(page, /async function consolidateSelectedClips\(\)/);
  assert.match(page, /saveTimeline\(\{ silent: true, requireLatest: true \}\)/);
  assert.match(page, /expectedRevision: saved\.revision/);
  assert.match(page, /consolidateAbortRef\.current\?\.abort\(\)/);
  assert.match(page, />\s*\{consolidateBusy \? '取消合并' : '合并为素材'\}\s*<\/Button>/);
  assert.match(page, /原片段保持不变/);
});

test('拖动裁剪预览会真实改变时间线片段的显示位置和宽度', () => {
  assert.match(page, /const shownStart = trimPreview\?\.key === key \? trimPreview\.startMs/);
  assert.match(page, /const shownDuration = trimPreview\?\.key === key \? trimPreview\.durationMs/);
  assert.match(page, /timelineTimeToPixels\(shownDuration, timelineZoom\)/);
});

test('音频片段提供真实淡入淡出拖动手柄和键盘精调', () => {
  for (const token of ['function beginAudioFadeDrag(', 'function updateAudioFadeDrag(', 'function finishAudioFadeDrag(', 'function nudgeAudioFade(']) assert.match(page, new RegExp(token.replace(/[()]/g, '\\$&')));
  for (const label of ['音频淡入手柄', '音频淡出手柄']) assert.match(page, new RegExp(label));
  assert.match(page, /clip\.audio_fade_set/);
});

test('透明工程提供持久背景设置、棋盘格预览和Alpha导出', () => {
  assert.match(page, /updateAICutProject/);
  assert.match(page, /画布背景/);
  assert.match(page, /backgroundMode/);
  assert.match(page, /backgroundColor/);
  assert.match(page, /is-transparent/);
  assert.match(page, /透明通道/);
  assert.match(page, /ProRes 4444/);
});

test('时间线标记点和关键帧经过统一命令层并提供真实页面入口', () => {
  for (const type of ['timeline.marker_add', 'timeline.marker_remove', 'clip.keyframe_upsert', 'clip.keyframe_remove']) assert.match(page, new RegExp(type.replace('.', '\\.')));
  for (const label of ['＋关键帧', '−关键帧', '＋标记', '−标记', '时间线标记点']) assert.match(page, new RegExp(label));
  assert.match(page, /aicut-timeline-marker/);
  assert.match(page, /timelineTimeToPixels\(marker\.timeMs, timelineZoom\)/);
  assert.match(page, /aicut-keyframe-dot/);
  assert.match(page, /seekPreview\(Number\(clip\.startMs\) \+ Number\(frame\.offsetMs\)\)/);
});

test('转场时长经过统一命令层并提供可拖动手柄', () => {
  assert.match(page, /type: 'clip\.transition_set'/);
  assert.match(page, /function beginTransitionDrag\(/);
  assert.match(page, /aicut-transition-handle/);
  assert.match(page, /draggedTransitionDuration/);
  assert.match(page, /function updateClipDrag\(event\) \{\s*if \(transitionDragRef\.current\)/);
  assert.match(page, /function finishClipDrag\(event\) \{\s*if \(transitionDragRef\.current\)/);
});

test('固定变速和曲线变速都经过统一命令层并有中文可视化面板', () => {
  assert.match(page, /type: 'clip\.playback_rate_set'/);
  assert.match(page, /type: 'clip\.speed_curve_set'/);
  assert.match(page, /speedCurvePresets/);
  for (const label of ['曲线变速', '英雄时刻', '蒙太奇', '子弹时间', '渐快', '渐慢']) assert.match(page, new RegExp(label));
  assert.match(page, /<SpeedCurveEditor/);
  assert.doesNotMatch(page, /commitTimeline\(current => setClipPlaybackRate/);
  assert.match(inspector, /onSetPlaybackRate/);
  assert.match(inspector, /onSetSpeedCurve/);
  assert.match(inspector, /speedCurvePresets/);
  assert.match(inspector, /<SpeedCurveEditor/);
});

test('实际属性栏的音量包络可新增、拖动和删除并经过统一命令层', () => {
  for (const type of ['clip.volume_point_upsert', 'clip.volume_point_move', 'clip.volume_point_remove']) assert.match(page, new RegExp(type.replace('.', '\\.')));
  assert.match(inspector, /VolumeEnvelopeEditor/);
  assert.match(inspector, /aicut-volume-envelope/);
  assert.match(inspector, /onPointerDown/);
  assert.match(inspector, /onDoubleClick/);
  assert.match(inspector, /删除选中音量点/);
  assert.match(inspector, /audioEnvelopePoints/);
  assert.match(inspector, /key=\{point\.id \|\| `base-\$\{index\}`\}/);
  assert.match(inspector, /<svg[^>]+onPointerMove=\{moveDraft\}[^>]+onPointerUp=\{finishDrag\}/);
});

test('复合片段使用统一命令层并能在时间线展开成员和拆散', () => {
  assert.match(page, /type: ['"]clip\.compound_create['"]/);
  assert.match(page, /type: ['"]clip\.compound_dissolve['"]/);
  assert.match(page, /创建复合片段/);
  assert.match(page, /展开复合成员/);
  assert.match(page, /拆散复合片段/);
  assert.match(page, /aicut-compound-chip/);
});

test('复合片段可以打开独立内部时间线并用统一命令编辑成员', () => {
  const workspace = fs.readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');
  assert.match(page, /打开复合时间线/);
  assert.match(page, /activeCompoundId/);
  assert.match(page, /CompoundTimelineWorkspace/);
  assert.match(workspace, /返回主时间线/);
  assert.match(workspace, /compoundMemberMoveCommand/);
  assert.match(workspace, /compoundMemberResizeCommand/);
  assert.match(workspace, /onCommand/);
  assert.match(workspace, /startOffsetRef\.current/);
  assert.match(workspace, /durationRef\.current/);
});

test('多层复合与组内轨道提供真实导航、增删和素材拖入入口', () => {
  assert.match(page, /compound\.track_add/);
  assert.match(page, /compound\.track_remove/);
  assert.match(page, /onOpenCompound=\{\(compoundId\) => setActiveCompoundId\(compoundId\)\}/);
  assert.match(page, /onAddTrack=\{addCompoundTrack\}/);
  assert.match(page, /onRemoveTrack=\{removeCompoundTrack\}/);
  assert.match(page, /onDropAsset=\{dropAssetOnCompoundTrack\}/);
  assert.match(compoundWorkspace, /新增画面轨/);
  assert.match(compoundWorkspace, /新增音频轨/);
  assert.match(compoundWorkspace, /新增字幕轨/);
  assert.match(compoundWorkspace, /进入子复合片段/);
  assert.match(compoundWorkspace, /删除空轨道/);
});

test('主时间线把复合成员显示为一个可进入和可拆散的真实片段块', () => {
  const workspace = fs.readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');
  assert.match(page, /collapseCompoundTimelineView/);
  assert.match(page, /collapsedDisplayTracks/);
  assert.match(page, /clip\.isCompoundContainer/);
  assert.match(page, /aicut-compound-container/);
  assert.match(page, /type: 'clip\.compound_move'/);
  assert.match(page, /type: 'clip\.compound_trim'/);
  assert.match(page, /function beginCompoundDrag/);
  assert.match(page, /function beginCompoundTrim/);
  assert.match(page, /if \(drag\.compoundId\) trimCompoundClip\(drag\.compoundId, drag\.edge, edgeTimeMs\)/);
  assert.match(page, /复合片段左裁剪手柄/);
  assert.match(page, /复合片段右裁剪手柄/);
  assert.match(page, /function nudgeCompoundTrim/);
  assert.match(page, /type: 'clip\.compound_split'/);
  assert.match(page, /function splitCompoundClip/);
  assert.match(page, /splitCompoundAtPointer/);
  assert.match(workspace, /在播放头分割/);
  assert.match(workspace, /onSplit/);
  assert.match(workspace, /分割位置/);
  assert.match(workspace, /splitTimeRef/);
  assert.match(page, /compoundDragMovedRef/);
  assert.match(page, /setActiveCompoundId\(clip\.compoundId\)/);
  assert.match(page, /个成员 · 单击进入/);
  assert.doesNotMatch(page, /个成员 · 双击进入/);
  assert.match(page, /onDissolve=\{\(compoundId\) => dissolveCompound\(compoundId\)\}/);
  assert.match(workspace, /onDissolve/);
  assert.match(workspace, />拆散复合片段</);
});

test('时间线片段提供显式多选按钮且复合块一次选择全部成员', () => {
  assert.match(page, /className="aicut-multiselect-toggle"/);
  assert.match(page, /aria-label=\{selectionActive \? '取消多选' : '加入多选'\}/);
  assert.match(page, /const keys = isCompoundContainer \? \(clip\.memberKeys \|\| compoundMemberKeys\(clip\.compoundId\)\) : \[key\]/);
  assert.match(page, /toggleTimelineSelection\(keys\)/);
  assert.match(page, /function toggleTimelineSelection\(keys\)/);
});

test('背景音乐面板调用真实波形节拍接口并把节拍写入时间线标记', () => {
  assert.match(page, /getAICutAudioAnalysis/);
  assert.match(page, /async function analyzeBGMAudio\(asset\)/);
  assert.match(page, /const analysis = await getAICutAudioAnalysis\(projectId, asset\.id, 240\)/);
  assert.match(page, /setBGMAudioAnalysis\(\{ \.\.\.analysis, assetId: asset\.id \}\)/);
  assert.match(page, /className="aicut-bgm-waveform"/);
  assert.match(page, /bgmAudioAnalysis\.peaks\.map/);
  assert.match(page, /检测到 \{bgmAudioAnalysis\.beatsMs\.length\} 个节拍点/);
  assert.match(page, /beatsMs: bgmAudioAnalysis\?\.assetId === asset\.id \? bgmAudioAnalysis\.beatsMs : \[\]/);
});

test('文字面板提供字体排版和逐字动画并写入轨道设置', () => {
  assert.match(page, /fontFamily/);
  assert.match(page, /fontSize/);
  assert.match(page, /textColor/);
  assert.match(page, /strokeWidth/);
  assert.match(page, /letterSpacing/);
  assert.match(page, /lineHeight/);
  assert.match(page, /textAlign/);
  assert.match(page, /逐字高亮/);
});

test('标题阴影与背景板进入统一文字样式、预览和导出数据', () => {
  for (const label of ['背景板', '背景颜色', '阴影颜色', '阴影横向', '阴影纵向', '阴影模糊']) assert.match(page, new RegExp(label));
  assert.match(page, /shadowOffsetX/);
  assert.match(page, /shadowOffsetY/);
  assert.match(page, /shadowBlur/);
  assert.match(page, /normalizeAICutTextStyle\(textStyleDraft\)/);
});

test('曲线文字提供真实弧度控件并使用逐字预览布局', () => {
  assert.match(page, /文字弧度/);
  assert.match(page, /textCurve/);
  assert.match(page, /aicutCurvedTextLayout/);
  assert.match(page, /aicut-curved-glyph/);
});

test('混排文字把输入框选区写成可保存的逐字样式', () => {
  assert.match(page, /混排选区/);
  assert.match(page, /selectionStart/);
  assert.match(page, /textRuns/);
  assert.match(page, /aicutTextRunStyleAt/);
});

test('自定义文字入口同步切换旧工作台状态后再打开弹窗', () => {
  assert.match(page, /onOpenTextEditor=\{\(\) => \{\s*setMaterialTab\('文本'\);\s*setToolWorkbenchOpen\('文本'\);/);
});

test('文字模板与识别字幕使用不同主页签避免互相覆盖', () => {
  assert.match(page, /toolCategory === '字幕' \? \(/);
  assert.doesNotMatch(page, /toolCategory === '文字' \? \(\s*<CaptionWorkspace/);
});

test('资源库文字模板应用真实预设而不是只保存模板名字', () => {
  assert.match(page, /aicutTextResourcePreset/);
  assert.match(page, /textResourceId: resource\.id/);
  assert.match(page, /\.\.\.templateStyle/);
});

test('音频处理从属性栏进入统一命令层而不是只改本地展示', () => {
  assert.match(page, /type: 'clip\.audio_processing_set'/);
  assert.match(page, /onSetAudioProcessing/);
  assert.match(inspector, /降噪/);
  assert.match(inspector, /声道混音/);
  assert.match(inspector, /audioEffectOptions/);
  assert.match(audioProcessing, /人声增强/);
  assert.match(audioProcessing, /花栗鼠声/);
  assert.match(audioProcessing, /教堂混响/);
  assert.match(audioProcessing, /黑胶唱片/);
});

test('文字转语音生成持久项目素材并可加入播放头位置或取消', () => {
  assert.match(page, /generateAICutSpeech/);
  assert.match(page, /aria-label="文字转语音"/);
  assert.match(page, /仅生成素材/);
  assert.match(page, /生成并加入时间线/);
  assert.match(page, /addAssetToTimeline\(asset, \{ startMs: previewTimeMs \}\)/);
  assert.match(page, /ttsAbortRef\.current\.abort\(\)/);
});

test('图片背景移除生成透明项目素材并支持加入播放头和取消', () => {
  assert.match(page, /removeAICutImageBackground/);
  assert.match(page, /aria-label="图片背景移除"/);
  assert.match(page, /仅生成透明素材/);
  assert.match(page, /去背并加入时间线/);
  assert.match(page, /addAssetToTimeline\(created, \{ startMs: previewTimeMs \}\)/);
  assert.match(page, /backgroundRemovalAbortRef\.current\.abort\(\)/);
});

test('声音克隆生成持久音频素材并支持加入播放头和取消', () => {
  assert.match(page, /cloneAICutVoice/);
  assert.match(page, /aria-label="声音克隆"/);
  assert.match(page, /仅生成克隆音频/);
  assert.match(page, /克隆并加入时间线/);
  assert.match(page, /addAssetToTimeline\(created, \{ startMs: previewTimeMs \}\)/);
  assert.match(page, /voiceCloneAbortRef\.current\.abort\(\)/);
});

test('高级画面从属性栏进入统一命令层并在浏览器使用真实处理器', () => {
  assert.match(page, /type: 'clip\.visual_processing_set'/);
  assert.match(page, /onSetVisualProcessing/);
  assert.match(page, /startVisualProcessingRenderer/);
  assert.match(inspector, /高级调色/);
  assert.match(inspector, /色键抠像/);
  assert.match(inspector, /反转蒙版/);
});

test('吸管取色时停用画布变换操作层，真实点击能落到播放器', () => {
  assert.match(page, /<PreviewCanvas enabled=\{Boolean\(selectedClipEntry && \['video', 'image'\]\.includes\(selectedClipEntry\.clip\.kind\)\) && !chromaPickerActive && !trackingSelectionActive && !objectRemovalSelectionActive\}/);
  assert.match(page, /onPointerDown=\{pickChromaFromPreview\}/);
});

test('时间线播放头跟随真实滚动容器并允许用户关闭', () => {
  assert.match(page, /ref=\{timelineScrollRef\} className="aicut-tracks-scroll"/);
  assert.match(page, /nextTimelineFollowScrollLeft\(\{/);
  assert.match(page, /跟随播放头/);
  assert.match(page, /playheadFollow: !current\.playheadFollow/);
});

test('时间线框选靠近边缘时持续滚动并在结束时停止', () => {
  assert.match(page, /nextLassoAutoScrollLeft\(\{/);
  assert.match(page, /drag\.autoScrollFrame = window\.requestAnimationFrame/);
  assert.match(page, /stopTimelineLassoAutoScroll\(drag\)/);
  assert.match(page, /stopTimelineLassoAutoScroll\(\)/);
});

test('专业导出参数从弹窗进入异步渲染请求', () => {
  for (const label of ['导出设置', '封装格式', '分辨率', '720P', '1080P', '4K', '帧率', '24 fps', '60 fps', '编码器', 'H.264', 'HEVC', 'AV1', '色深', '8-bit', '10-bit', '目标码率', '质量', '快速预览', '标准', '高质量']) {
    assert.match(page, new RegExp(label));
  }
  assert.match(page, /createAICutExportJob\(projectId, format, saved\.revision, settings\)/);
  assert.match(page, /aria-label="硬件加速"/);
  assert.match(page, /自动选择（失败回退软件）/);
  assert.match(page, /listAICutExportJobs\(projectId\)/);
});

test('代理媒体只替换网页预览且素材卡提供真实生成入口', () => {
  assert.match(page, /生成代理/);
  assert.match(page, /代理已启用/);
  assert.match(page, /generateAICutProxy/);
  assert.match(page, /asset\.proxyUrl \|\| item\.asset\.url/);
  assert.match(page, /setAssets\(\(current\) => current\.map/);
});
