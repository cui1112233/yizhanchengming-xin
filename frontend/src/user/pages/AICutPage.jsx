import { Button, Checkbox, ConfigProvider, Dropdown, Empty, Input, InputNumber, Modal, Select, Spin, Tag, Upload, message } from 'antd';
import { Bot, ClipboardPaste, MoreHorizontal, LockKeyhole, UnlockKeyhole, Volume2, VolumeX, SkipBack, SkipForward, Maximize2, Captions, Clapperboard, CloudUpload, Copy, Film, ImageIcon, Keyboard, Magnet, Music, Palette, PanelBottomClose, PanelBottomOpen, PanelLeftClose, PanelLeftOpen, Pause, Play, Plus, Redo2, Save, Scissors, Sparkles, Star, Trash2, Type, Undo2, Upload as UploadIcon, WandSparkles, X } from 'lucide-react';
import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import PlayerViewControls from './aicut/PlayerViewControls.jsx';
import PreviewViewport from './aicut/PreviewViewport.jsx';
import { previewSource } from './aicut/playerView.js';
import { listAssetFolders, createAssetFolder, renameAssetFolder, deleteAssetFolder, moveAssetToFolder } from '../../shared/api/aicutAssetFolders.js';
import { clipAnimationPreviewStyle } from './aicutClipAnimation.js';
import { catalogTextSettings, gradientGlyphColor, startCatalogPreviewRenderer } from './aicut/catalogPreview.js';
import PersonalTemplateWorkspace from './aicut/PersonalTemplateWorkspace.jsx';
import TimelineZoomControls from './aicut/TimelineZoomControls.jsx';
import { cancelAICutASRJob, cancelAICutExportJob, cleanAICutOralTimeline, cloneAICutVoice, consolidateAICutClips, correctAICutCaptions, createAICutASRJob, createAICutExportJob, createAICutProject, downloadAICutExportJob, extractAICutAudio, extractAICutTrackingFrames, extractAICutVisualFrames, generateAICutProxy, generateAICutSpeech, getAICutAudioAnalysis, getAICutTimelineHistory as getDefaultAICutTimelineHistory, importAICutSRT, listAICutASRJobs, listAICutAssets, listAICutExportJobs, listAICutProjects, loadAICutTimeline as loadDefaultAICutTimeline, redoAICutTimeline as redoDefaultAICutTimeline, removeAICutImageBackground, saveAICutTimeline as saveDefaultAICutTimeline, smartSplitAICutVideo, trimAICutFillers, undoAICutTimeline as undoDefaultAICutTimeline, updateAICutProject, uploadAICutAsset } from '../../shared/api/aicut.js';
import { listProjectTimelines, createProjectTimeline, renameProjectTimeline, duplicateProjectTimeline, deleteProjectTimeline, reorderProjectTimelines, loadProjectTimeline, saveProjectTimeline, projectTimelineHistory, moveProjectTimelineHistory } from '../../shared/api/aicutTimelines.js';
import { setAICutTimelineRequestScope } from '../../shared/api/aicutTimelineScope.js';
import { saveTimelineDraftCopy } from '../../shared/api/aicutDraftRecovery.js';
import TimelineTabs from './aicut/TimelineTabs.jsx';
import AudioRecorder from './aicut/AudioRecorder.jsx';
import MultiCamWorkspace from './aicut/MultiCamWorkspace.jsx';
import { getConfig } from '../../shared/api/config.js';
import { alignAICutMultiCamAudio } from '../../shared/api/aicut.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import AICutAgentPanel from '../components/AICutAgentPanel.jsx';
import AICutPetHost from '../components/AICutPetHost.jsx';
import { getCurrentUsername } from '../../shared/api/auth.js';
import { registerPetPageService, openPetPageService } from '../../shared/pet/pageService.js';
const AICutCreativeCanvas = lazy(() => import('../components/AICutCreativeCanvas.jsx'));
import { addEffectLayer, addTimelineTrack, draggedTimelineStart, markTimelineClip, normalizeSubtitleRhythm, orderedTimelineTracks, removeTimelineTrack, renameTimelineTrack, setTimelineClipText, subtitleRhythmLabel, timelinePixelDeltaToTime, timelinePixelsToTime, timelineTimeToPixels, toggleTimelineTrackFlag, updateTimelineTrackSettings } from './aicutTimeline.js';
import { createAICutClipboard, executeAICutCommand, findTimelineEditPoint, reconcileTimelineSave, restoreTimelineSnapshot } from './aicutCommands.js';
import { buildAICutPreviewFrame, getAICutTimelineDuration, nextAICutCanvasFit, seekAICutPreviewTime } from './aicutPreview.js';
import { speedCurvePresets } from './aicutSpeedCurve.js';
import { lassoGestureMoved, nextLassoAutoScrollLeft, normalizeLassoRect, selectLassoClipKeys } from './aicutLasso.js';
import { draggedTransitionDuration, nextTimelineFollowScrollLeft, resolveAICutShortcut, timelineDropTime } from './aicutWorkbench.js';
import { createAICutCapabilities, filterAICutCapabilities, recordRecentAICutFeature, toggleFavoriteAICutFeature } from './aicutCapabilities.js';
import { normalizeVisualEffect, subtitleAnimationCatalog, subtitleStyleCatalog, textStyleCatalog, transitionCatalog, visualEffectCatalog, visualEffectPreviewStyle, visualEffectPreset } from './aicutEffects.js';
import EditorShell from './aicut/EditorShell.jsx';
import TimelineToolbar from './aicut/TimelineToolbar.jsx';
import ProjectOverview from './aicut/ProjectOverview.jsx';
import SourcePreviewPanel from './aicut/SourcePreviewPanel.jsx';
import ProjectTemplateDialog from './aicut/ProjectTemplateDialog.jsx';
import StickerWorkspace from './aicut/StickerWorkspace.jsx';
import AdjustmentWorkspace from './aicut/AdjustmentWorkspace.jsx';
import PackagingWorkspace from './aicut/PackagingWorkspace.jsx';
import EditorPluginWorkspace from './aicut/EditorPluginWorkspace.jsx';
import { hasColorGradeAdjustments } from './aicut/colorGrade.js';
import { projectTemplateCatalog } from './aicut/projectTemplates.js';
import MediaArtwork from './aicut/MediaArtwork.jsx';
import { timelineRulerTicks } from './aicut/timelineRuler.js';
import { buildTimelineAdvancedItems, buildTimelinePrimaryActions } from './aicut/timelineTools.js';
import ResourceLibrary from './aicut/ResourceLibrary.jsx';
import PreviewCanvas from './aicut/PreviewCanvas.jsx';
import TrackingRegionOverlay from './aicut/TrackingRegionOverlay.jsx';
import InspectorPanel from './aicut/InspectorPanel.jsx';
import CaptionWorkspace from './aicut/CaptionWorkspace.jsx';
import CompoundTimelineWorkspace from './aicut/CompoundTimelineWorkspace.jsx';
import ExportQueuePanel from './aicut/ExportQueuePanel.jsx';
import SpeedCurveEditor from './aicut/SpeedCurveEditor.jsx';
import { collapseCompoundTimelineView } from './aicutCompoundTimeline.js';
import { applyCompoundPath, compoundEntries, compoundPathForClip } from './aicutCompoundModel.js';
import { normalizeEditorLayout } from './aicut/editorLayout.js';
import { catalogById, catalogForTab, normalizeCatalogParameters } from './aicut/editorCatalog.js';
import { aicutCurvedTextLayout, aicutFontFaceCSS, aicutFontOptionsForAssets, aicutTextPreviewStyle, aicutTextResourcePreset, aicutTextRunStyleAt, normalizeAICutTextStyle } from './aicutText.js';
import { createAudioProcessingGraph } from './aicutAudioProcessing.js';
import { detectChromaScreenColor, normalizeVisualProcessing, sampleChromaColorFromPixels, startVisualProcessingRenderer, trackRegionSequence, visualProcessingCSS } from './aicutVisualProcessing.js';
import { buildPersonSegmentationSignature, generatePersonMatteVideo, isAICutPersonMatteAsset, normalizePersonSegmentation, startPersonMattePreviewRenderer } from './aicutPersonSegmentation.js';
import { normalizeTrackingRegion } from './aicutTrackingRegion.js';
import './aicut-page.css';
import './aicut/editor.css';
import './aicut/editor-redesign.css';

const kindLabels = {
  image: '图片',
  video: '视频',
  audio: '音频',
  subtitle: '字幕'
};
const trackLabels = {
  video: '画面轨道',
  audio: '音频轨道',
  subtitle: '字幕轨道'
};
const referenceTrackNames = {
  subtitle: '字幕 1',
  effect: '特效 1',
  video: '视频 1',
  audio: '音频 1'
};
const asrStateLabels = {
  queued: '排队中',
  running: '识别中',
  cancel_requested: '正在取消',
  succeeded: '已写入字幕',
  failed: '失败',
  canceled: '已取消'
};
const editorTools = [
  ['媒体', ImageIcon],
  ['音频', Music],
  ['文本', Type],
  ['特效', Sparkles],
  ['转场', Film],
  ['字幕', Captions],
  ['AI工具', WandSparkles]
];
const ttsVoiceOptions = [
  { label: '晓晓（女声·温柔）', value: 'zh-CN-XiaoxiaoNeural' },
  { label: '晓辰（女声·知性）', value: 'zh-CN-XiaochenNeural' },
  { label: '云希（男声·清朗）', value: 'zh-CN-YunxiNeural' },
  { label: '云扬（男声·阳光）', value: 'zh-CN-YunyangNeural' },
  { label: '晓伊（女声·甜美）', value: 'zh-CN-XiaoyiNeural' },
  { label: '云健（男声·稳重）', value: 'zh-CN-YunjianNeural' }
];
const ttsStyleOptions = [
  { label: '通用', value: 'general' },
  { label: '开心', value: 'cheerful' },
  { label: '悲伤', value: 'sad' },
  { label: '友好', value: 'friendly' },
  { label: '聊天', value: 'chat' }
];
const inspectorTabs = ['基础', '速度'];
const toolSourceLabels = {
  browser: '浏览器',
  cloud: '云服务',
  ai: 'AI模型',
  render: 'Go后台'
};
const catalogPresetMap = {
  'filter-original': 'original',
  'filter-bright': 'bright',
  'filter-warm-sun': 'warm',
  'filter-cool-blue': 'cool-cinema',
  'filter-monochrome': 'grayscale',
  'filter-soft': 'soft-focus',
  'filter-golden-hour': 'golden-hour',
  'filter-noir': 'noir',
  'filter-vintage': 'vintage',
  'filter-cyberpunk': 'cyberpunk',
  'filter-vaporwave': 'vaporwave',
  'filter-mint-haze': 'mint-haze',
  'filter-bleach-bypass': 'bleach-bypass',
  'filter-winter-cool': 'winter-cool',
  'filter-rosy': 'rosy',
  'filter-film': 'film',
  'filter-super-8': 'super-8',
  'filter-instant-print': 'instant-print',
  'effect-blur': 'soft-focus',
  'effect-vignette': 'vignette',
  'effect-glow': 'glow',
  'effect-shake': 'shake',
  'effect-breathe-zoom': 'breathe',
  'effect-flash': 'flash',
  'effect-mosaic': 'mosaic'
};

function visualEffectForCatalogResource(resource, parameters = resource?.previewParameters || {}) {
  const preset = catalogPresetMap[resource?.id];
  if (preset) return normalizeVisualEffect({ ...visualEffectPreset(preset), ...parameters });
  if (resource?.category === 'filter')
    return normalizeVisualEffect({
      preset: 'custom',
      contrast: 1.08,
      saturation: 1.12,
      intensity: Number(parameters.intensity ?? 0.5),
      speed: Number(parameters.speed ?? 1)
    });
  if (resource?.category === 'effect')
    return normalizeVisualEffect({
      preset: 'custom',
      intensity: Number(parameters.intensity ?? 0.5),
      speed: Number(parameters.speed ?? 1),
      blur: Number(parameters.radius ?? 0)
    });
  return null;
}
function newID(prefix) {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

function objectRemovalToTrackingRegion(value = {}) {
  return normalizeTrackingRegion({ centerX: value.x, centerY: value.y, width: value.width, height: value.height });
}

function trackKindForAsset(asset) {
  if (asset.kind === 'audio') return 'audio';
  if (asset.kind === 'subtitle') return 'subtitle';
  return 'video';
}

function defaultDuration(asset) {
  if (Number(asset.durationMs) > 0) return Number(asset.durationMs);
  if (asset.kind === 'image') return 5000;
  return 10000;
}

function aicutAbortError() {
  return typeof DOMException === 'function' ? new DOMException('操作已取消', 'AbortError') : Object.assign(new Error('操作已取消'), { name: 'AbortError' });
}

function loadTrackingImage(url, signal) {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) return reject(aicutAbortError());
    const image = new Image();
    const cleanup = () => signal?.removeEventListener('abort', abort);
    const abort = () => {
      image.src = '';
      cleanup();
      reject(aicutAbortError());
    };
    image.onload = () => { cleanup(); resolve(image); };
    image.onerror = () => { cleanup(); reject(new Error('跟踪帧解码失败')); };
    signal?.addEventListener('abort', abort, { once: true });
    image.src = String(url || '');
  });
}

async function decodeTrackingFrames(frames, signal) {
  const decoded = [];
  let width = 0;
  let height = 0;
  const canvas = document.createElement('canvas');
  const context = canvas.getContext('2d', { willReadFrequently: true });
  if (!context) throw new Error('浏览器无法读取跟踪画面');
  for (const frame of frames || []) {
    if (signal?.aborted) throw aicutAbortError();
    const image = await loadTrackingImage(frame.url, signal);
    if (!width || !height) {
      width = Math.max(1, image.naturalWidth || image.width || 0);
      height = Math.max(1, image.naturalHeight || image.height || 0);
      canvas.width = width;
      canvas.height = height;
    }
    context.clearRect(0, 0, width, height);
    context.drawImage(image, 0, 0, width, height);
    const rgba = context.getImageData(0, 0, width, height).data;
    const pixels = new Uint8Array(width * height);
    for (let index = 0; index < pixels.length; index += 1) {
      const offset = index * 4;
      pixels[index] = Math.round(rgba[offset] * .2126 + rgba[offset + 1] * .7152 + rgba[offset + 2] * .0722);
    }
    decoded.push({ timeMs: Number(frame.timeMs), pixels });
  }
  return { frames: decoded, width, height };
}

function srtTime(ms) {
  const value = Math.max(0, Number(ms) || 0);
  const hours = Math.floor(value / 3600000);
  const minutes = Math.floor((value % 3600000) / 60000);
  const seconds = Math.floor((value % 60000) / 1000);
  const millis = Math.floor(value % 1000);
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')},${String(millis).padStart(3, '0')}`;
}

function subtitleClips(timeline) {
  return timeline.tracks
    .filter((track) => track.kind === 'subtitle')
    .flatMap((track) => track.clips || [])
    .filter((clip) => String(clip.text || '').trim())
    .sort((a, b) => Number(a.startMs) - Number(b.startMs));
}

function previewClock(ms) {
  const value = Math.max(0, Number(ms) || 0);
  const minutes = Math.floor(value / 60000);
  const seconds = Math.floor((value % 60000) / 1000);
  const hundredths = Math.floor((value % 1000) / 10);
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}.${String(hundredths).padStart(2, '0')}`;
}

function applyCanvasTransform(style, canvasTransform) {
  if (!canvasTransform) return style;
  const transform = canvasTransform;
  const directTransform = `translate(${Number(transform.x || 0) * 100}%, ${Number(transform.y || 0) * 100}%) rotate(${Number(transform.rotation || 0)}deg) scale(${Number(transform.scaleX || 1)}, ${Number(transform.scaleY || 1)})`;
  return {
    ...style,
    transform: `${style.transform || ''} ${directTransform}`.trim(),
    opacity: Number.isFinite(Number(transform.opacity)) ? Number(transform.opacity) : (style.opacity ?? 1)
  };
}

function BasicPreviewMedia({ item, playing, quality = 'original' }) {
  const mediaRef = useRef(null);
  const audioGraphRef = useRef(null);
	const mediaURL = previewSource(item.asset, quality);
  useEffect(() => {
    if (item.kind !== 'audio' || !mediaRef.current) return undefined;
    audioGraphRef.current = createAudioProcessingGraph(mediaRef.current, item.audioProcessing);
    return () => {
      audioGraphRef.current?.close().catch(() => {});
      audioGraphRef.current = null;
    };
  }, [item.kind, mediaURL]);
  useEffect(() => {
    audioGraphRef.current?.update(item.audioProcessing);
  }, [item.audioProcessing]);
  useEffect(() => {
    const media = mediaRef.current;
    if (!media || item.kind === 'image') return;
    const expected = item.sourceTimeMs / 1000;
	const manualSeek = item.frozen || item.reversed;
	if (Number.isFinite(expected) && Math.abs(media.currentTime - expected) > (manualSeek ? .035 : .2)) media.currentTime = expected;
    media.playbackRate = item.playbackRate || 1;
    if (item.kind === 'audio') media.volume = Math.min(1, Math.max(0, item.volume));
	if (manualSeek) media.pause();
	else if (playing) {
      audioGraphRef.current?.resume().catch(() => {});
      media.play().catch(() => {});
    }
    else media.pause();
  }, [item.frozen, item.kind, item.playbackRate, item.reversed, item.sourceTimeMs, item.volume, playing]);
  const effectStyle = visualEffectPreviewStyle(item.visualEffect, item.sourceTimeMs);
  const processingStyle = visualProcessingCSS(item.visualProcessing);
  const style = applyCanvasTransform({
    ...effectStyle,
    ...processingStyle,
    filter: [effectStyle.filter, processingStyle.filter].filter(Boolean).join(' ') || undefined
  }, item.canvasTransform);
  if (item.kind === 'image') return <img src={item.asset.url} alt={item.asset.name} style={style} />;
  if (item.kind === 'audio') return <audio ref={mediaRef} src={mediaURL} preload="auto" />;
  return <video ref={mediaRef} src={mediaURL} preload="auto" muted playsInline style={style} />;
}

function hasAdvancedVisualProcessing(value) {
  const processing = normalizeVisualProcessing(value);
  const color = processing.color;
	const geometry = processing.geometry;
  return processing.chromaKey.enabled
    || processing.objectRemoval.enabled
    || processing.personSegmentation.enabled
    || processing.stabilization.enabled
    || processing.chromaKey.removeBlack
    || processing.mask.type !== 'none'
	|| geometry.cropLeft !== 0
	|| geometry.cropTop !== 0
	|| geometry.cropRight !== 0
	|| geometry.cropBottom !== 0
	|| geometry.flipHorizontal
	|| geometry.flipVertical
    || processing.details.sharpen !== 0
    || processing.details.vignette !== 0
    || processing.details.grain !== 0
    || color.curvePoints.length !== 5
    || color.curvePoints.some(point => Math.abs(point.y - point.x) > .000001)
    || color.exposure !== 0
    || color.temperature !== 0
    || color.tint !== 0
    || color.highlights !== 0
    || color.shadows !== 0
    || color.gamma !== 1 || hasColorGradeAdjustments(color);
}

function ProcessedVisualMedia({ item, playing, quality = 'original' }) {
  const mediaRef = useRef(null);
  const canvasRef = useRef(null);
  const rendererRef = useRef(null);
  const catalogCanvasRef = useRef(null);
  const catalogRendererRef = useRef(null);
	const matteVideoRef = useRef(null);
	const matteCanvasRef = useRef(null);
	const matteRendererRef = useRef(null);
	const mediaURL = previewSource(item.asset, quality);
  const personSegmentation = normalizePersonSegmentation(item.visualProcessing?.personSegmentation);
  const matteURL = item.matteAsset?.proxyUrl || item.matteAsset?.url || '';
  const hasPersonMatte = item.kind === 'video' && item.personSegmentationCurrent && Boolean(matteURL);
  const [rendererReady, setRendererReady] = useState(false);
  const processingKey = JSON.stringify(normalizeVisualProcessing(item.visualProcessing));
  const catalogLayers = [...(item.catalogResourceId ? [{ resourceId:item.catalogResourceId,parameters:item.catalogParameters || {} }] : []), ...(item.layerEffects || [])];
  const catalogKey = JSON.stringify(catalogLayers);
  const hasCatalog = catalogLayers.length > 0;
  const [catalogReady,setCatalogReady] = useState(false);

  useEffect(() => {
    const media = mediaRef.current;
    if (!media || item.kind === 'image') return;
    const expected = item.sourceTimeMs / 1000;
	const manualSeek = item.frozen || item.reversed;
	if (Number.isFinite(expected) && Math.abs(media.currentTime - expected) > (manualSeek ? .035 : .2)) media.currentTime = expected;
    media.playbackRate = item.playbackRate || 1;
	if (manualSeek) media.pause();
	else if (playing) media.play().catch(() => {});
    else media.pause();
    const matte = matteVideoRef.current;
    if (matte) {
      const matteTime = Math.max(0, Number(item.outputOffsetMs) || 0) / 1000;
      if (Math.abs(matte.currentTime - matteTime) > .08) matte.currentTime = matteTime;
      matte.playbackRate = 1;
      if (manualSeek) matte.pause();
      else if (playing) matte.play().catch(() => {});
      else matte.pause();
    }
  }, [hasPersonMatte, item.frozen, item.kind, item.outputOffsetMs, item.playbackRate, item.reversed, item.sourceTimeMs, playing]);

  useEffect(() => {
    rendererRef.current?.stop();
    rendererRef.current = startVisualProcessingRenderer(canvasRef.current, mediaRef.current, item.visualProcessing);
    setRendererReady(Boolean(rendererRef.current));
    return () => {
      rendererRef.current?.stop();
      rendererRef.current = null;
    };
  }, [mediaURL, item.kind, processingKey]);

  useEffect(() => {
    matteRendererRef.current?.stop();
    matteRendererRef.current = hasPersonMatte
      ? startPersonMattePreviewRenderer(matteCanvasRef.current, canvasRef.current, matteVideoRef.current, personSegmentation)
      : null;
    return () => {
      matteRendererRef.current?.stop();
      matteRendererRef.current = null;
    };
  }, [hasPersonMatte, matteURL, personSegmentation.threshold, personSegmentation.feather, processingKey]);

  useEffect(() => {
    catalogRendererRef.current?.stop();
    catalogRendererRef.current = hasCatalog ? startCatalogPreviewRenderer(catalogCanvasRef.current, hasPersonMatte ? matteCanvasRef.current : canvasRef.current,catalogLayers,item.outputOffsetMs) : null;
    setCatalogReady(Boolean(catalogRendererRef.current));
    return () => { catalogRendererRef.current?.stop();catalogRendererRef.current=null; };
  },[hasCatalog,catalogKey,hasPersonMatte,mediaURL,processingKey]);
  useEffect(() => { catalogRendererRef.current?.update(catalogLayers,item.outputOffsetMs); },[catalogKey,item.outputOffsetMs]);

  const canvasStyle = applyCanvasTransform(visualEffectPreviewStyle(item.catalogResourceId ? null : item.visualEffect, item.sourceTimeMs), item.canvasTransform);
  const source = item.kind === 'image'
    ? <img ref={mediaRef} src={mediaURL} alt="" className="aicut-processed-source" />
    : <video ref={mediaRef} src={mediaURL} preload="auto" muted playsInline className="aicut-processed-source" />;
  return (
    <div className="aicut-processed-wrap">
      {source}
      <canvas ref={canvasRef} className={`aicut-processed-visual${hasPersonMatte ? ' aicut-person-matte-processor' : ''}`} style={{...(hasPersonMatte ? {} : canvasStyle),...(hasCatalog?{position:'absolute',visibility:'hidden'}:{})}} aria-label="GPU高级画面预览" />
      {hasPersonMatte ? <><video ref={matteVideoRef} src={matteURL} preload="auto" muted playsInline className="aicut-person-matte-source" /><canvas ref={matteCanvasRef} className="aicut-processed-visual aicut-person-matte-output" style={{...canvasStyle,...(hasCatalog?{position:'absolute',visibility:'hidden'}:{})}} aria-label="AI人物抠像预览" /></> : null}
      {hasCatalog ? <canvas ref={catalogCanvasRef} className="aicut-processed-visual" style={canvasStyle} aria-label="时间线滤镜与特效预览" /> : null}
      {!rendererReady ? <span className="aicut-processed-fallback">正在准备高级画面预览…</span> : null}
      {hasCatalog && !catalogReady ? <span className="aicut-processed-fallback">当前浏览器无法初始化特效预览，请检查 WebGL 支持。</span> : null}
    </div>
  );
}

function PreviewMedia(props) {
  const { item } = props;
  if (['video', 'image'].includes(item.kind) && (hasAdvancedVisualProcessing(item.visualProcessing) || item.catalogResourceId || item.layerEffects?.length)) {
    return <ProcessedVisualMedia {...props} />;
  }
  return <BasicPreviewMedia {...props} />;
}

function AICutCaptionText({ item }) {
  const curve = Number(item?.textStyle?.textCurve || 0);
  const hasTextRuns = Array.isArray(item?.textStyle?.textRuns) && item.textStyle.textRuns.length > 0;
  const gradient = item.styleId === 'text-gradient-title' || item.styleId === 'text-gradient';
  if (!curve && !hasTextRuns && !gradient) {
    return item.highlightedCount > 0 ? <><span className="aicut-text-highlighted">{Array.from(item.visibleText).slice(0, item.highlightedCount).join('')}</span>{Array.from(item.visibleText).slice(item.highlightedCount).join('')}</> : item.visibleText;
  }
  const glyphs = aicutCurvedTextLayout(item.visibleText, curve, item.textStyle?.fontSize, item.textStyle?.letterSpacing);
  return (
    <span className="aicut-curved-text" aria-label={item.visibleText}>
      {glyphs.map((glyph, index) => (
        <span
          aria-hidden="true"
          className={`aicut-curved-glyph${index < item.highlightedCount ? ' aicut-text-highlighted' : ''}`}
          key={`${index}:${glyph.character}`}
          style={{ ...(gradient ? {color:gradientGlyphColor(index,Array.from(item.text).length),WebkitTextFillColor:gradientGlyphColor(index,Array.from(item.text).length)} : {}),...aicutTextRunStyleAt(item.textStyle, index, 'var(--aicut-preview-px, 1px)'), transform: curve ? `translateY(calc(${glyph.translateY.toFixed(2)} * var(--aicut-preview-px, 1px))) rotate(${glyph.rotate.toFixed(2)}deg)` : undefined }}
        >
          {glyph.character}
        </span>
      ))}
    </span>
  );
}

export default function AICutPage() {
  const [projects, setProjects] = useState([]);
  const [projectId, setProjectId] = useState('');
  const [activeTimelineId, setActiveTimelineId] = useState(() => new URLSearchParams(window.location.search).get('timeline') || 'default');
  const [projectTimelines, setProjectTimelines] = useState([]);
  const [timelineBusy, setTimelineBusy] = useState(false);
  const activeTimelineScopeRef = useRef('');
  activeTimelineScopeRef.current = `${projectId}:${activeTimelineId}`;
  const [loadedTimelineProjectId, setLoadedTimelineProjectId] = useState('');
  const [assets, setAssets] = useState([]);
  const [timeline, setTimeline] = useState({ revision: 0, tracks: [] });
  const [selectedAssetId, setSelectedAssetId] = useState('');
  const [sourcePreviewAsset, setSourcePreviewAsset] = useState(null);
  const [templateApplication, setTemplateApplication] = useState(null);
  const [multiCamView, setMultiCamView] = useState(null);
  const [conflictRecovery, setConflictRecovery] = useState(null);
  const [recoveringDraft, setRecoveringDraft] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saveState, setSaveState] = useState('saved');
  const [uploading, setUploading] = useState(false);
  const [importingSRT, setImportingSRT] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newRatio, setNewRatio] = useState('9:16');
  const [captionOpen, setCaptionOpen] = useState(false);
  const [captionScript, setCaptionScript] = useState('');
  const [captionModelId, setCaptionModelId] = useState('');
  const [correctingCaptions, setCorrectingCaptions] = useState(false);
  const captionAbortRef = useRef(null);
  const [asrOpen, setASROpen] = useState(false);
  const [asrModels, setASRModels] = useState([]);
  const [asrModelId, setASRModelId] = useState('');
  const [asrLanguage, setASRLanguage] = useState('zh');
  const [asrJobs, setASRJobs] = useState([]);
  const [asrBusy, setASRBusy] = useState(false);
  const [materialTab, setMaterialTab] = useState('我的素材');
  const [workspaceMode, setWorkspaceMode] = useState('editor');
  const [toolCategory, setToolCategory] = useState('媒体');
  const [assetFolders, setAssetFolders] = useState([]);
  const [selectedAssetFolderId, setSelectedAssetFolderId] = useState('all');
  const [previewResource, setPreviewResource] = useState(null);
  const [inspectorTab, setInspectorTab] = useState('基础');
  const [inspectorDrawerOpen, setInspectorDrawerOpen] = useState(false);
  const [agentDrawerOpen, setAgentDrawerOpen] = useState(false);
  const petPageServiceCleanup = useRef(null);
  const petPageRuntime = useRef(null);
  const [autoEditOpen, setAutoEditOpen] = useState(false);
  const autoEditScopeRef = useRef('');
  const [assetSearch, setAssetSearch] = useState('');
  const [libraryFilter, setLibraryFilter] = useState('全部');
  const [favoriteFeatures, setFavoriteFeatures] = useState(() => {
    try {
      return JSON.parse(window.localStorage.getItem('aicut.favoriteFeatures') || '[]');
    } catch {
      return [];
    }
  });
  const [recentFeatures, setRecentFeatures] = useState(() => {
    try {
      return JSON.parse(window.localStorage.getItem('aicut.recentFeatures') || '[]');
    } catch {
      return [];
    }
  });
  const [toolWorkbenchOpen, setToolWorkbenchOpen] = useState('');
  const [editorLayout, setEditorLayout] = useState(() => {
    try {
      const stored = JSON.parse(window.localStorage.getItem('aicut.editorLayout.v3') || '{}');
      return normalizeEditorLayout(stored, {
        width: window.innerWidth,
        height: window.innerHeight
      });
    } catch {
      return normalizeEditorLayout({}, { width: window.innerWidth, height: window.innerHeight });
    }
  });
  const mediaCollapsed = editorLayout.resourceCollapsed;
  const inspectorCollapsed = editorLayout.inspectorCollapsed;
  const timelineCollapsed = editorLayout.timelineCollapsed;
  const setMediaCollapsed = (value) =>
    setEditorLayout((current) => ({
      ...current,
      resourceCollapsed: typeof value === 'function' ? value(current.resourceCollapsed) : Boolean(value)
    }));
  const setInspectorCollapsed = (value) =>
    setEditorLayout((current) => ({
      ...current,
      inspectorCollapsed: typeof value === 'function' ? value(current.inspectorCollapsed) : Boolean(value)
    }));
  const setTimelineCollapsed = (value) =>
    setEditorLayout((current) => ({
      ...current,
      timelineCollapsed: typeof value === 'function' ? value(current.timelineCollapsed) : Boolean(value)
    }));
  const [shortcutOpen, setShortcutOpen] = useState(false);
  const [agentModels, setAgentModels] = useState([]);
  const [agentDraft, setAgentDraft] = useState(null);
  const [activeAgentSkill, setActiveAgentSkill] = useState('');
  const [bladeMode, setBladeMode] = useState(false);
  const [magneticSnap, setMagneticSnap] = useState(true);
  const [linkedMove, setLinkedMove] = useState(true);
  const [canvasFit, setCanvasFit] = useState('fit');
  const [previewQuality, setPreviewQuality] = useState('original');
  const [previewFullscreen, setPreviewFullscreen] = useState(false);
  const [previewExpanded, setPreviewExpanded] = useState(false);
  const [trackEditId, setTrackEditId] = useState('');
  const [trackNameDraft, setTrackNameDraft] = useState('');
  const [replaceOpen, setReplaceOpen] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [exportingFormat, setExportingFormat] = useState('');
  const [exportJobs, setExportJobs] = useState([]);
  const [exportQueueError, setExportQueueError] = useState('');
  const [exportActionBusy, setExportActionBusy] = useState('');
  const [exportSettings, setExportSettings] = useState({ format: 'mp4', resolution: '1080p', frameRate: 30, quality: 'standard', codec: 'h264', bitDepth: 8, bitrateMbps: 0, hardwareAcceleration: 'auto' });
  const [canvasSettingsOpen, setCanvasSettingsOpen] = useState(false);
  const [canvasSettingsSaving, setCanvasSettingsSaving] = useState(false);
  const [canvasDraft, setCanvasDraft] = useState({ backgroundMode: 'color', backgroundColor: '#000000' });
  const [mediaToolBusy, setMediaToolBusy] = useState('');
  const [replaceFrom, setReplaceFrom] = useState('');
  const [replaceTo, setReplaceTo] = useState('');
  const [bgmPreference, setBGMPreference] = useState('');
  const [bgmOptions, setBGMOptions] = useState({
    align: true,
    fade: true,
    continue: false,
    fill: true
  });
  const [bgmReferenceTrackId, setBGMReferenceTrackId] = useState('');
  const [bgmAudioAnalysis, setBGMAudioAnalysis] = useState(null);
  const [bgmAnalysisBusy, setBGMAnalysisBusy] = useState(false);
  const [ttsDraft, setTTSDraft] = useState({ text: '', name: '', voice: 'zh-CN-XiaoxiaoNeural', speed: 1.8, pitch: 10, style: 'general' });
  const [ttsBusy, setTTSBusy] = useState(false);
  const ttsAbortRef = useRef(null);
  const [backgroundRemovalOpen, setBackgroundRemovalOpen] = useState(false);
  const [backgroundRemovalAsset, setBackgroundRemovalAsset] = useState(null);
  const [backgroundRemovalName, setBackgroundRemovalName] = useState('');
  const [backgroundRemovalBusyId, setBackgroundRemovalBusyId] = useState('');
  const backgroundRemovalAbortRef = useRef(null);
  const [voiceCloneOpen, setVoiceCloneOpen] = useState(false);
  const [voiceCloneDraft, setVoiceCloneDraft] = useState({ assetId: '', text: '', name: '', speed: 1, pitch: 0, style: 'general' });
  const [voiceCloneBusy, setVoiceCloneBusy] = useState(false);
  const voiceCloneAbortRef = useRef(null);
  const [visualRequirement, setVisualRequirement] = useState('');
  const [visualFrames, setVisualFrames] = useState([]);
  const [visualBusy, setVisualBusy] = useState(false);
  const [visualEffectDraft, setVisualEffectDraft] = useState(() => normalizeVisualEffect());
  const [subtitleStyle, setSubtitleStyle] = useState('classic-outline');
  const [subtitleRhythm, setSubtitleRhythm] = useState('按句/短语');
  const [subtitleAnimation, setSubtitleAnimation] = useState('none');
  const [manualSubtitle, setManualSubtitle] = useState({
    text: '',
    startMs: 0,
    durationMs: 2000
  });
  const [textStyleId, setTextStyleId] = useState('plain');
  const [textStyleDraft, setTextStyleDraft] = useState(() => normalizeAICutTextStyle());
  const textInputRef = useRef(null);
  const [textSelection, setTextSelection] = useState({ start: 0, end: 0 });
  const [textRunDraft, setTextRunDraft] = useState({ textColor: '#ffcc00', fontSize: 72, bold: true, italic: false });
  const [selectedClipKey, setSelectedClipKey] = useState('');
  const [selectedClipKeys, setSelectedClipKeys] = useState([]);
  const [consolidateBusy, setConsolidateBusy] = useState(false);
  const consolidateAbortRef = useRef(null);
  const [chromaPickerActive, setChromaPickerActive] = useState(false);
  const [activeCompoundId, setActiveCompoundId] = useState('');
  const [clipboardCount, setClipboardCount] = useState(0);
  const [keyframeClipboardCount, setKeyframeClipboardCount] = useState(0);
  const [regionTrackingBusy, setRegionTrackingBusy] = useState(false);
  const [trackingRegion, setTrackingRegion] = useState(() => normalizeTrackingRegion());
  const [trackingSelectionActive, setTrackingSelectionActive] = useState(false);
  const [objectRemovalRegion, setObjectRemovalRegion] = useState(() => normalizeTrackingRegion({ width: .2, height: .2 }));
  const [objectRemovalSelectionActive, setObjectRemovalSelectionActive] = useState(false);
  const trackingAbortRef = useRef(null);
  const [personSegmentationBusy, setPersonSegmentationBusy] = useState(false);
  const [personSegmentationProgress, setPersonSegmentationProgress] = useState(0);
  const personSegmentationAbortRef = useRef(null);
  const [rippleMode, setRippleMode] = useState(false);
  const [previewTimeMs, setPreviewTimeMs] = useState(0);
  const [previewPlaying, setPreviewPlaying] = useState(false);
  const [safeGuide, setSafeGuide] = useState(false);
  const [timelineZoom, setTimelineZoom] = useState(1);
  const [dragPreview, setDragPreview] = useState(null);
  const [trimPreview, setTrimPreview] = useState(null);
  const [transitionPreview, setTransitionPreview] = useState(null);
  const [audioFadePreview, setAudioFadePreview] = useState(null);
  const [timelineLasso, setTimelineLasso] = useState(null);
  const [bgmOpen, setBGMOpen] = useState(false);
  const [fillerOpen, setFillerOpen] = useState(false);
  const [oralOpen, setOralOpen] = useState(false);
  const [fillerModelId, setFillerModelId] = useState('');
  const [trimmingFillers, setTrimmingFillers] = useState(false);
  const fillerAbortRef = useRef(null);
  const [historyVersion, setHistoryVersion] = useState(0);
  const [serverHistory, setServerHistory] = useState({
    canUndo: false,
    canRedo: false
  });
  const undoRef = useRef([]);
  const redoRef = useRef([]);
  const timelineRef = useRef(timeline);
  const editSequenceRef = useRef(0);
  const lastPersistedEditSequenceRef = useRef(0);
  const saveInFlightRef = useRef(null);
  const seenASRTerminalRef = useRef(new Set());
  const clipDragRef = useRef(null);
  const audioFadeDragRef = useRef(null);
  const compoundDragMovedRef = useRef(false);
  const trimDragRef = useRef(null);
  const transitionDragRef = useRef(null);
  const timelineLassoRef = useRef(null);
  const timelineCanvasRef = useRef(null);
  const openedTimelineProjectRef = useRef('');
  const timelineScrollRef = useRef(null);
  const timelineRulerScrollRef = useRef(null);
  const layoutDragRef = useRef(null);
  const editorShellRef = useRef(null);
  const previewStageRef = useRef(null);
  const previewPanelRef = useRef(null);
  useEffect(() => {
    const update = () => setPreviewFullscreen(document.fullscreenElement === previewPanelRef.current);
    const escape = event => { if (event.key === 'Escape') setPreviewExpanded(false); };
    document.addEventListener('fullscreenchange', update);
    document.addEventListener('keydown', escape);
    return () => { document.removeEventListener('fullscreenchange', update); document.removeEventListener('keydown', escape); };
  }, []);
  const clipboardRef = useRef(null);
  const keyframeClipboardRef = useRef(null);

  useEffect(() => () => {
    trackingAbortRef.current?.abort();
    personSegmentationAbortRef.current?.abort();
    if (timelineLassoRef.current?.autoScrollFrame) window.cancelAnimationFrame(timelineLassoRef.current.autoScrollFrame);
  }, []);

  useEffect(() => {
    if (!chromaPickerActive) return undefined;
    const cancelPicker = (event) => {
      if (event.key === 'Escape') setChromaPickerActive(false);
    };
    window.addEventListener('keydown', cancelPicker);
    return () => window.removeEventListener('keydown', cancelPicker);
  }, [chromaPickerActive]);

  useEffect(() => {
    timelineRef.current = timeline;
  }, [timeline]);

  const currentProject = projects.find((item) => item.id === projectId);
	const fontAssets = useMemo(() => assets.filter((asset) => asset.kind === 'font' && asset.fontFamily), [assets]);
	const fontOptions = useMemo(() => aicutFontOptionsForAssets(fontAssets), [fontAssets]);
	const projectFontCSS = useMemo(() => aicutFontFaceCSS(fontAssets), [fontAssets]);
  const selectedAsset = assets.find((item) => item.id === selectedAssetId) || assets[0];
  const visibleAssets = useMemo(
    () =>
      assets.filter((asset) => {
        if (isAICutPersonMatteAsset(asset)) return false;
        if (materialTab === '音频' && asset.kind !== 'audio') return false;
        if (materialTab === '视觉' && !['video', 'image'].includes(asset.kind)) return false;
        if (materialTab === '字幕' && asset.kind !== 'subtitle') return false;
        return !assetSearch.trim() || asset.name.toLowerCase().includes(assetSearch.trim().toLowerCase());
      }),
    [assets, assetSearch, materialTab]
  );
  const durationMs = useMemo(() => getAICutTimelineDuration(timeline), [timeline]);
  const previewFrame = useMemo(() => {
    const frame = buildAICutPreviewFrame(timeline, assets, previewTimeMs);
    const selectedPreviewClipId = selectedClipKey.split('::')[1];
    if (!previewResource || !['filter','effect'].includes(previewResource.category) || !selectedPreviewClipId) return frame;
    return {
      ...frame,
      visuals: frame.visuals.map((item) => item.clipId !== selectedPreviewClipId ? item : previewResource.category === 'filter' ? { ...item,visualEffect:null,catalogResourceId:previewResource.id,catalogParameters:previewResource.previewParameters || {} } : { ...item,layerEffects:[...(item.layerEffects||[]),{resourceId:previewResource.id,parameters:previewResource.previewParameters||{}}] })
    };
  }, [timeline, assets, previewTimeMs, previewResource, selectedClipKey]);
  const selectedClipEntry = useMemo(() => {
    const [trackId, clipId] = selectedClipKey.split('::');
    const track = timeline.tracks.find((item) => item.id === trackId);
    const clip = track?.clips?.find((item) => item.id === clipId);
    return track && clip ? { track, clip } : null;
  }, [timeline, selectedClipKey]);

  useEffect(() => {
    setTrackingSelectionActive(false);
    setObjectRemovalSelectionActive(false);
    const removal = normalizeVisualProcessing(selectedClipEntry?.clip?.visualProcessing).objectRemoval;
    setObjectRemovalRegion(objectRemovalToTrackingRegion(removal));
  }, [selectedClipKey]);

  const commitObjectRemovalRegion = (region, enabled = true) => {
    if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return;
    const normalized = normalizeTrackingRegion(region);
    const processing = normalizeVisualProcessing(selectedClipEntry.clip.visualProcessing);
    setObjectRemovalRegion(normalized);
    runTimelineCommand({
      type: 'clip.visual_processing_set',
      trackId: selectedClipEntry.track.id,
      clipId: selectedClipEntry.clip.id,
      processing: {
        ...processing,
        objectRemoval: { enabled: Boolean(enabled), x: normalized.centerX, y: normalized.centerY, width: normalized.width, height: normalized.height }
      }
    });
  };

  const captureSelectedPreviewFrame = () => {
    const stage = previewStageRef.current;
    const visual = [...(stage?.querySelectorAll?.('.aicut-preview-visual') || [])].find(item => item.dataset.clipKey === selectedClipKey);
    if (!visual) throw new Error('播放头处没有当前片段画面，请先把播放头移到片段内');
    const source = visual.querySelector('.aicut-processed-source') || visual.querySelector('video, img') || visual.querySelector('canvas');
    const width = Number(source?.videoWidth || source?.naturalWidth || source?.width);
    const height = Number(source?.videoHeight || source?.naturalHeight || source?.height);
    if (!source || width < 1 || height < 1) throw new Error('当前画面还没有准备好，请稍后再试');
    const scale = Math.min(1, 320 / Math.max(width, height));
    const canvas = document.createElement('canvas');
    canvas.width = Math.max(1, Math.round(width * scale));
    canvas.height = Math.max(1, Math.round(height * scale));
    const context = canvas.getContext('2d', { willReadFrequently: true });
    context.drawImage(source, 0, 0, canvas.width, canvas.height);
    return { image: context.getImageData(0, 0, canvas.width, canvas.height), rect: visual.getBoundingClientRect() };
  };

  const setSelectedChromaColor = (color) => {
    if (!selectedClipEntry || !color) return;
    const processing = normalizeVisualProcessing(selectedClipEntry.clip.visualProcessing);
    runTimelineCommand({
      type: 'clip.visual_processing_set',
      trackId: selectedClipEntry.track.id,
      clipId: selectedClipEntry.clip.id,
      processing: { ...processing, chromaKey: { ...processing.chromaKey, enabled: true, removeBlack: false, color } }
    });
  };

  const detectSelectedChroma = () => {
    try {
      const { image } = captureSelectedPreviewFrame();
      const color = detectChromaScreenColor(image.data, image.width, image.height);
      if (!color) throw new Error('没有识别到可用幕布颜色');
      setSelectedChromaColor(color);
      message.success(`已识别幕布颜色 ${color}`);
    } catch (error) {
      message.warning(error?.message || '自动识别幕布失败');
    }
  };

  const pickChromaFromPreview = (event) => {
    if (!chromaPickerActive) return;
    event.preventDefault();
    try {
      const { image, rect } = captureSelectedPreviewFrame();
      const x = Math.floor((event.clientX - rect.left) / Math.max(1, rect.width) * image.width);
      const y = Math.floor((event.clientY - rect.top) / Math.max(1, rect.height) * image.height);
      const color = sampleChromaColorFromPixels(image.data, image.width, image.height, x, y);
      if (!color) throw new Error('点击位置不在有效画面内');
      setSelectedChromaColor(color);
      setChromaPickerActive(false);
      message.success(`已取色 ${color}`);
    } catch (error) {
      message.warning(error?.message || '吸管取色失败');
    }
  };
  const capabilities = createAICutCapabilities({
    uploadAsset: () => {
      document.querySelector('.aicut-library-upload input[type=file]')?.click();
      return false;
    },
    openAssets: () => {
      setMaterialTab('我的素材');
      setToolWorkbenchOpen('项目素材');
    },
    extractAudio: () => (selectedAsset?.kind === 'video' ? extractAudioAsset(selectedAsset) : (message.warning('请先在媒体中选择视频素材'), false)),
    openText: () => {
      setMaterialTab('文字稿');
      setToolWorkbenchOpen('默认文本');
    },
    openBatchText: () => (subtitleClips(timeline).length ? setReplaceOpen(true) : (message.warning('请先添加字幕'), false)),
    openVisualEffects: () => {
      if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return (message.warning('请先选择画面片段'), false);
      if (selectedClipEntry.track.locked) return (message.warning('轨道已锁定'), false);
      setVisualEffectDraft(normalizeVisualEffect(selectedClipEntry.clip.visualEffect));
      setMaterialTab('画面特效');
      setToolWorkbenchOpen('画面特效');
    },
    openTransitions: () => {
      setMaterialTab('转场');
      setToolWorkbenchOpen('基础转场');
    },
    openTranscription: () => (selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? setASROpen(true) : (message.warning('请先选择音频或视频素材'), false)),
    openCaptionCorrection: agentModels.length
      ? () => {
          if (!subtitleClips(timeline).length) return (message.warning('请先添加或识别字幕'), false);
          setCaptionScript(timeline.script || captionScript);
          setCaptionOpen(true);
        }
      : undefined,
    openSubtitleUpload: () => {
      setMaterialTab('文字稿');
      setToolWorkbenchOpen('字幕文件上传');
    },
    downloadSubtitles: () => exportSRT(),
    openSpeechCleanup: agentModels.length ? () => (subtitleClips(timeline).length ? setOralOpen(true) : (message.warning('请先添加或识别字幕'), false)) : undefined,
    smartSplit: () => (selectedAsset?.kind === 'video' ? smartSplitAsset(selectedAsset) : (message.warning('请先在媒体中选择视频素材'), false))
  });
  const visibleEditorTools = editorTools.filter(([label]) => capabilities.some((item) => item.category === label));
  const displayTracks = useMemo(
    () =>
      ['subtitle', 'video', 'audio'].flatMap((kind) => {
        const matches = timeline.tracks.filter((track) => track.kind === kind);
        return matches.length
          ? matches
          : [
              {
                id: `empty-${kind}`,
                kind,
                name: `${kind === 'subtitle' ? '字幕' : kind === 'video' ? '视频' : '音频'}1`,
                clips: [],
                placeholder: true
              }
            ];
      }),
    [timeline]
  );

  useEffect(() => {
    let active = true;
    listAICutProjects()
      .then((items) => {
        if (!active) return;
        setProjects(items);
        const requestedId = new URLSearchParams(window.location.search).get('project');
        const requested = items.find(item => item.id === requestedId);
        if (requestedId && !requested) {
          message.warning('这个工程不存在或暂时无法访问，请从作品库重新打开');
          window.location.replace('/ai-cut');
          return;
        }
        setProjectId(requested?.id || items[0]?.id || '');
      })
      .catch((error) => message.error(error?.message || '剪辑项目读取失败'))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!projectId) {
      setAssets([]);
      setTimeline({ revision: 0, tracks: [] });
      setServerHistory({ canUndo: false, canRedo: false });
      return;
    }
    let active = true;
    setLoading(true);
    setSelectedClipKey('');
    setSelectedClipKeys([]);
    setLoadedTimelineProjectId('');
    setTimeline({ revision: 0, tracks: [] });
    timelineRef.current = { revision: 0, tracks: [] };
    setSaveState('saved');
    setAICutTimelineRequestScope(projectId, activeTimelineId);
    setAutoEditOpen(false); setAgentDrawerOpen(false);
    Promise.all([loadAICutTimeline(projectId), listAICutAssets(projectId), getAICutTimelineHistory(projectId)])
      .then(([nextTimeline, nextAssets, nextHistory]) => {
        if (!active) return;
        setTimeline(nextTimeline);
        setSourcePreviewAsset(null);
        setLoadedTimelineProjectId(`${projectId}:${activeTimelineId}`);
        editSequenceRef.current = 0; lastPersistedEditSequenceRef.current = 0;
        setSaveState('saved');
        undoRef.current = [];
        redoRef.current = [];
        setHistoryVersion((value) => value + 1);
        setAssets(nextAssets);
        setServerHistory(nextHistory);
        setSelectedAssetId(nextAssets[0]?.id || '');
      })
      .catch((error) => message.error(error?.message || 'AI-CUT 项目加载失败'))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [projectId, activeTimelineId]);

  useEffect(() => {
    setPreviewPlaying(false);
    setPreviewTimeMs(0);
    setSelectedClipKey('');
    setSelectedClipKeys([]);
    clipboardRef.current = null;
    setClipboardCount(0);
    keyframeClipboardRef.current = null;
    setKeyframeClipboardCount(0);
  }, [projectId]);

  useEffect(() => {
    setPreviewTimeMs((current) => seekAICutPreviewTime(current, durationMs));
  }, [durationMs]);

  useEffect(() => {
    if (!previewPlaying || durationMs <= 0) return undefined;
    const originTime = previewTimeMs;
    const originClock = performance.now();
    let frameID = 0;
    const tick = (now) => {
      const next = seekAICutPreviewTime(originTime + now - originClock, durationMs);
      setPreviewTimeMs(next);
      if (next >= durationMs) {
        setPreviewPlaying(false);
        return;
      }
      frameID = window.requestAnimationFrame(tick);
    };
    frameID = window.requestAnimationFrame(tick);
    return () => window.cancelAnimationFrame(frameID);
  }, [previewPlaying, durationMs]);

  useEffect(() => {
    const scroller = timelineScrollRef.current;
    if (!scroller || timelineCollapsed) return;
    const next = nextTimelineFollowScrollLeft({
      enabled: editorLayout.playheadFollow,
      scrollLeft: scroller.scrollLeft,
      viewportWidth: scroller.clientWidth,
      contentWidth: scroller.scrollWidth,
      playheadX: timelineTimeToPixels(previewTimeMs, timelineZoom),
    });
    if (Math.abs(next - scroller.scrollLeft) >= 1) scroller.scrollLeft = next;
  }, [previewTimeMs, timelineZoom, timelineCollapsed, editorLayout.playheadFollow]);

  useEffect(() => {
    if (!projectId) {
      setExportJobs([]);
      setExportQueueError('');
      return undefined;
    }
    let active = true;
    const refresh = async (reportError = false) => {
      try {
        const jobs = await listAICutExportJobs(projectId);
        if (active) {
          setExportJobs(jobs);
          setExportQueueError('');
        }
      } catch (error) {
        if (!active) return;
        // 读取失败不能保持空数组，否则面板会显示成「还没有导出任务」。
        setExportQueueError(error?.message || '导出任务队列读取失败');
        if (reportError) message.error(error?.message || '导出任务队列读取失败');
      }
    };
    refresh(true);
    const timer = window.setInterval(() => refresh(false), 1000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [projectId, activeTimelineId]);

  useEffect(() => {
    const settings = timeline.tracks.find((track) => track.kind === 'subtitle')?.settings;
    if (!settings) return;
    setSubtitleStyle(settings.subtitleStyleId || '黑底白字');
    setSubtitleRhythm(subtitleRhythmLabel(settings.subtitleRhythm));
    setSubtitleAnimation(settings.subtitleAnimation || '无动效');
  }, [timeline.revision, projectId]);

  useEffect(() => {
    if (!projectId) {
      setASRJobs([]);
      return;
    }
    let active = true;
    const refresh = async () => {
      try {
        const jobs = await listAICutASRJobs(projectId);
        if (!active) return;
        setASRJobs(jobs);
        for (const job of jobs) {
          if (!['succeeded', 'failed', 'canceled'].includes(job.state) || seenASRTerminalRef.current.has(job.id)) continue;
          seenASRTerminalRef.current.add(job.id);
          if (job.state === 'succeeded') {
            const refreshedTimeline = await loadAICutTimeline(projectId);
            if (!active) return;
            if (lastPersistedEditSequenceRef.current < editSequenceRef.current) {
              message.warning('字幕结果已保存，保留当前尚未保存的编辑；保存时会检查版本冲突');
              continue;
            }
            setTimeline(refreshedTimeline);
            undoRef.current = []; redoRef.current = [];
            message.success(`语音识别完成，已写入 ${job.imported || 0} 条字幕`);
          } else if (job.state === 'failed') message.error(job.error || '语音识别失败');
          else message.info('语音识别已取消');
        }
      } catch (error) {
        if (active) message.error(error?.message || '语音识别任务读取失败');
      }
    };
    seenASRTerminalRef.current = new Set();
    refresh();
    const timer = window.setInterval(refresh, 2000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [projectId, activeTimelineId]);

  useEffect(() => {
    let active = true;
    listAvailableModels('asr')
      .then((models) => {
        if (!active) return;
        setASRModels(models);
        setASRModelId((current) => current || models[0]?.id || '');
      })
      .catch((error) => {
        if (active) message.error(error?.message || 'ASR 模型读取失败');
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    listAvailableModels('text')
      .then((models) => {
        if (active) setAgentModels(models);
      })
      .catch((error) => {
        if (active) message.error(error?.message || 'Agent 模型读取失败');
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    getConfig()
      .then((config) => {
        if (!active || !config?.tts) return;
        setTTSDraft((current) => ({
          ...current,
          voice: String(config.tts.voice || current.voice),
          style: String(config.tts.style || current.style),
          speed: Number(config.tts.speed ?? current.speed),
          pitch: Number(config.tts.pitch ?? current.pitch)
        }));
      })
      .catch(() => {});
    return () => {
      active = false;
      ttsAbortRef.current?.abort();
      backgroundRemovalAbortRef.current?.abort();
      voiceCloneAbortRef.current?.abort();
      consolidateAbortRef.current?.abort();
    };
  }, []);

  useEffect(() => {
    const handleResize = () =>
      setEditorLayout((current) =>
        normalizeEditorLayout(current, {
          width: window.innerWidth,
          height: window.innerHeight
        })
      );
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    window.localStorage.setItem('aicut.editorLayout.v3', JSON.stringify({ version: 3, ...editorLayout }));
  }, [editorLayout]);

  useEffect(() => {
    window.localStorage.setItem('aicut.favoriteFeatures', JSON.stringify(favoriteFeatures));
  }, [favoriteFeatures]);

  useEffect(() => {
    window.localStorage.setItem('aicut.recentFeatures', JSON.stringify(recentFeatures));
  }, [recentFeatures]);

  useEffect(() => {
    function handleShortcut(event) {
      if (event.defaultPrevented || event.target?.closest?.('.ant-modal')) return;
      const action = resolveAICutShortcut(
        {
          key: event.key,
          code: event.code,
          ctrlKey: event.ctrlKey,
          metaKey: event.metaKey,
          shiftKey: event.shiftKey,
          altKey: event.altKey,
          targetTag: event.target?.tagName,
          targetEditable: event.target?.isContentEditable
        },
        {
          hasProject: Boolean(projectId),
          hasSelection: selectedClipKeys.length > 0,
          hasClipboard: Boolean(clipboardRef.current?.items?.length)
        }
      );
      if (!action || (sourcePreviewAsset && action === 'play')) return;
      event.preventDefault();
      if (action === 'close') {
        setSourcePreviewAsset(null);
        setShortcutOpen(false);
        setCreateOpen(false);
        setExportOpen(false);
        setReplaceOpen(false);
        setCaptionOpen(false);
        setASROpen(false);
        setBGMOpen(false);
        setFillerOpen(false);
        setOralOpen(false);
        setAgentDrawerOpen(false);
      } else if (action === 'shortcuts') setShortcutOpen(true);
      else if (action === 'canvas-view') void openCreativeCanvas();
      else if (action === 'editor-view') setWorkspaceMode('editor');
      else if (action === 'new-project') setCreateOpen(true);
      else if (action === 'save') saveTimeline();
      else if (action === 'export') openExportPanel();
      else if (action === 'play') togglePreviewPlayback();
      else if (action === 'undo') undoTimeline();
      else if (action === 'redo') redoTimeline();
      else if (action === 'blade') setBladeMode((value) => !value);
      else if (action === 'split') splitAtPlayhead();
      else if (action === 'copy') copySelectedClips();
      else if (action === 'cut') cutSelectedClips();
      else if (action === 'paste') pasteClips();
      else if (action === 'delete') deleteSelectedClips();
      else if (action === 'nudge-previous') nudgeSelectedClip(-1);
      else if (action === 'nudge-next') nudgeSelectedClip(1);
      else if (action === 'edit-previous') jumpEditPoint('previous');
      else if (action === 'edit-next') jumpEditPoint('next');
    }
    window.addEventListener('keydown', handleShortcut);
    return () => window.removeEventListener('keydown', handleShortcut);
  }, [projectId, selectedClipEntry, selectedClipKeys, timeline, durationMs, previewPlaying, previewTimeMs, serverHistory, rippleMode, sourcePreviewAsset]);

  useEffect(() => {
    if (!projectId || saveState !== 'dirty' || saving) return undefined;
    const timer = window.setTimeout(() => saveTimeline({ silent: true }), 1500);
    return () => window.clearTimeout(timer);
  }, [projectId, timeline, saveState, saving]);

  const loadAICutTimeline = id => activeTimelineId === 'default' ? loadDefaultAICutTimeline(id) : loadProjectTimeline(id, activeTimelineId);
  const saveAICutTimeline = (id, value) => activeTimelineId === 'default' ? saveDefaultAICutTimeline(id, value) : saveProjectTimeline(id, activeTimelineId, value);
  const getAICutTimelineHistory = id => activeTimelineId === 'default' ? getDefaultAICutTimelineHistory(id) : projectTimelineHistory(id, activeTimelineId);
  const undoAICutTimeline = (id, revision) => activeTimelineId === 'default' ? undoDefaultAICutTimeline(id, revision) : moveProjectTimelineHistory(id, activeTimelineId, 'undo', revision);
  const redoAICutTimeline = (id, revision) => activeTimelineId === 'default' ? redoDefaultAICutTimeline(id, revision) : moveProjectTimelineHistory(id, activeTimelineId, 'redo', revision);

  useEffect(() => {
    setAICutTimelineRequestScope(projectId, activeTimelineId);
    return () => setAICutTimelineRequestScope();
  }, [projectId, activeTimelineId]);

  useEffect(() => {
    if (!projectId) { setProjectTimelines([]); return; }
    let live = true;
    listProjectTimelines(projectId).then(items => live && setProjectTimelines(items)).catch(error => live && message.error(error.message));
    return () => { live = false; };
  }, [projectId]);

  async function refreshTimelineList() { const items = await listProjectTimelines(projectId); setProjectTimelines(items); return items; }
  useEffect(() => {
    if (!projectId) return;
    const params = new URLSearchParams({ project: projectId });
    if (activeTimelineId !== 'default') params.set('timeline', activeTimelineId);
    window.history.replaceState(null, '', `/ai-cut?${params}`);
  }, [projectId, activeTimelineId]);
  useEffect(() => {
    let live = true;
    setAssetFolders([]); setSelectedAssetFolderId('all');
    if (projectId) listAssetFolders(projectId).then(items => { if (live) setAssetFolders(items); }).catch(error => { if (live) message.error(error.message || '素材文件夹读取失败'); });
    return () => { live = false; };
  }, [projectId]);
  async function manageAssetFolder(action) {
    const scope = activeTimelineScopeRef.current;
    try {
      await action();
      const [folders, nextAssets] = await Promise.all([listAssetFolders(projectId), listAICutAssets(projectId)]);
      if (activeTimelineScopeRef.current !== scope) return;
      setAssetFolders(folders); setAssets(nextAssets);
      setSelectedAssetFolderId(id => id === 'all' || !id || folders.some(folder => folder.id === id) ? id : '');
    } catch (error) { message.error(error.message || '文件夹操作失败'); throw error; }
  }
  async function persistBeforeTimelineChange() {
    if (saveInFlightRef.current || saveState === 'dirty' || saveState === 'failed') {
      const saved = await saveTimeline({ silent: true, requireLatest: true });
      if (!saved) throw new Error('先保存当前时间线，再切换');
    }
  }
  async function changeActiveTimeline(id) {
    if (timelineBusy || id === activeTimelineId) return;
    setTimelineBusy(true);
    try { await persistBeforeTimelineChange(); setLoading(true); setActiveTimelineId(id); }
    catch (error) { message.error(error.message); }
    finally { setTimelineBusy(false); }
  }
  async function timelineOperation(action) {
    setTimelineBusy(true);
    try { await persistBeforeTimelineChange(); const result = await action(); await refreshTimelineList(); return result; }
    catch (error) { message.error(error.message || '时间线操作失败'); throw error; }
    finally { setTimelineBusy(false); }
  }

  async function createProject() {
    if (timelineNetworkBusy) return message.info('请等待当前处理结束');
    if (!newTitle.trim()) return message.warning('请输入项目名称');
    try {
      if (projectId) await persistBeforeTimelineChange();
      const project = await createAICutProject({
        title: newTitle,
        aspectRatio: newRatio
      });
      setProjects((current) => [project, ...current]);
      setActiveTimelineId('default');
      setProjectId(project.id);
      window.history.replaceState(null, '', `/ai-cut?project=${encodeURIComponent(project.id)}`);
      setNewTitle('');
      setCreateOpen(false);
      message.success('AI-CUT 项目已创建');
    } catch (error) {
      message.error(error?.message || '项目创建失败');
    }
  }

  function openCanvasSettings() {
    setCanvasDraft({
      backgroundMode: currentProject?.backgroundMode || 'color',
      backgroundColor: currentProject?.backgroundColor || '#000000'
    });
    setCanvasSettingsOpen(true);
  }

  async function saveCanvasSettings() {
    if (!projectId) return;
    setCanvasSettingsSaving(true);
    try {
      const project = await updateAICutProject(projectId, canvasDraft);
      setProjects((current) => current.map((item) => (item.id === project.id ? project : item)));
      setCanvasSettingsOpen(false);
      message.success('画布背景已保存');
    } catch (error) {
      message.error(error?.message || '画布背景保存失败');
    } finally {
      setCanvasSettingsSaving(false);
    }
  }

  async function changeProjectRatio(aspectRatio) {
    if (!projectId || aspectRatio === currentProject?.aspectRatio) return;
    const scope = activeTimelineScopeRef.current;
    try {
      await persistBeforeTimelineChange();
      const updated = await updateAICutProject(projectId, { aspectRatio });
      if (activeTimelineScopeRef.current !== scope) return;
      setProjects(items => items.map(item => item.id === updated.id ? updated : item));
      setCanvasFit('fit');
      message.success(`画面比例已保存为 ${aspectRatio}`);
    } catch (error) { message.error(error.message || '画面比例保存失败'); }
  }

  async function togglePreviewFullscreen() {
    try {
      if (previewExpanded) { setPreviewExpanded(false); return; }
      if (document.fullscreenElement) await document.exitFullscreen();
      else if (previewPanelRef.current?.requestFullscreen) await previewPanelRef.current.requestFullscreen();
      else setPreviewExpanded(true);
    } catch { setPreviewExpanded(true); }
  }

  async function uploadRecording(file) {
    if (!projectId) throw new Error('先创建工程再保存录音');
    const scope = `${projectId}:${activeTimelineId}`;
    setUploading(true);
    try {
      const asset = await uploadAICutAsset(projectId, file);
      if (!asset?.id) throw new Error('录音没有保存成功，请重试');
      if (scope === activeTimelineScopeRef.current) {
        setAssets(current => [...current.filter(item => item.id !== asset.id), asset]);
        setSelectedAssetId(asset.id);
        setToolCategory('音频');
      }
      return asset;
    } finally { setUploading(false); }
  }
  async function openMultiCam() {
    try {
      await persistBeforeTimelineChange();
      setMultiCamView({ scope: `${projectId}:${activeTimelineId}`, snapshot: timelineRef.current });
    } catch (error) { message.warning(error.message); }
  }
  async function upload(file) {
    if (!projectId) return Upload.LIST_IGNORE;
    const uploadScope = activeTimelineScopeRef.current;
    const targetFolder = selectedAssetFolderId && selectedAssetFolderId !== 'all' ? selectedAssetFolderId : '';
    setUploading(true);
    try {
      const asset = await uploadAICutAsset(projectId, file);
      if (targetFolder) {
        try { await moveAssetToFolder(projectId, asset.id, targetFolder); asset.folderId = targetFolder; }
        catch { message.warning('素材已上传到未分类，目标文件夹已变化，请手动移动'); }
      }
      if (activeTimelineScopeRef.current !== uploadScope) return Upload.LIST_IGNORE;
      setAssets((current) => [...current, asset]);
      setSelectedAssetId(asset.id);
	  if (asset.kind === 'font') {
		setTextStyleDraft((current) => ({ ...current, fontFamily: asset.fontFamily, fontAssetId: asset.id }));
		setMaterialTab('文本');
		setToolWorkbenchOpen('文本');
	  }
      message.success('素材已上传并保存');
    } catch (error) {
      message.error(error?.message || '素材上传失败');
    } finally {
      setUploading(false);
    }
    return Upload.LIST_IGNORE;
  }

  async function generateSpeechAsset(addToTimeline) {
    if (!projectId) return message.warning('请先创建剪辑项目');
    if (ttsBusy) return message.info('配音正在生成，请等待或先取消');
    const controller = new AbortController();
    ttsAbortRef.current = controller;
    setTTSBusy(true);
    try {
      const asset = await generateAICutSpeech(projectId, { ...ttsDraft, signal: controller.signal });
      if (controller.signal.aborted || !asset) return;
      setAssets((current) => current.some((item) => item.id === asset.id) ? current : [...current, asset]);
      setSelectedAssetId(asset.id);
      if (addToTimeline) addAssetToTimeline(asset, { startMs: previewTimeMs });
      setTTSDraft((current) => ({ ...current, name: '' }));
      message.success(addToTimeline ? '配音已生成并加入播放头位置' : '配音已生成并保存到项目素材');
    } catch (error) {
      if (controller.signal.aborted || error?.name === 'AbortError') message.info('已取消本次配音生成');
      else message.error(error?.message || '配音生成失败');
    } finally {
      if (ttsAbortRef.current === controller) {
        ttsAbortRef.current = null;
        setTTSBusy(false);
      }
    }
  }

  function cancelSpeechGeneration() {
    if (!ttsAbortRef.current) return;
    ttsAbortRef.current.abort();
  }

  function openBackgroundRemoval(asset) {
    if (!asset || asset.kind !== 'image') return message.warning('请选择图片素材');
    if (backgroundRemovalBusyId === asset.id && backgroundRemovalAbortRef.current) {
      backgroundRemovalAbortRef.current.abort();
      return;
    }
    if (backgroundRemovalBusyId) return message.info('已有图片正在去背，请等待或先取消');
    setBackgroundRemovalAsset(asset);
    setBackgroundRemovalName(`${String(asset.name || '').replace(/\.[^.]+$/, '')}-透明背景`);
    setBackgroundRemovalOpen(true);
  }

  async function runBackgroundRemoval(addToTimeline) {
    if (!backgroundRemovalAsset || backgroundRemovalBusyId) return;
    const controller = new AbortController();
    backgroundRemovalAbortRef.current = controller;
    setBackgroundRemovalBusyId(backgroundRemovalAsset.id);
    try {
      const created = await removeAICutImageBackground(projectId, backgroundRemovalAsset.id, { name: backgroundRemovalName, signal: controller.signal });
      if (controller.signal.aborted || !created) return;
      setAssets((current) => current.some((item) => item.id === created.id) ? current : [...current, created]);
      setSelectedAssetId(created.id);
      if (addToTimeline) addAssetToTimeline(created, { startMs: previewTimeMs });
      setBackgroundRemovalOpen(false);
      message.success(addToTimeline ? '透明素材已生成并加入播放头位置' : '透明素材已生成并保存到项目素材');
    } catch (error) {
      if (controller.signal.aborted || error?.name === 'AbortError') message.info('已取消本次背景移除');
      else message.error(error?.message || '背景移除失败');
    } finally {
      if (backgroundRemovalAbortRef.current === controller) {
        backgroundRemovalAbortRef.current = null;
        setBackgroundRemovalBusyId('');
      }
    }
  }

  function cancelBackgroundRemoval() {
    if (backgroundRemovalAbortRef.current) {
      backgroundRemovalAbortRef.current.abort();
      return;
    }
    setBackgroundRemovalOpen(false);
  }

  function openVoiceClone(asset) {
    if (voiceCloneBusy) return message.info('声音克隆正在生成，请等待或先取消');
    const reference = asset?.kind === 'audio'
      ? asset
      : selectedAsset?.kind === 'audio'
        ? selectedAsset
        : assets.find((item) => item.kind === 'audio');
    if (!reference) return message.warning('请先上传或生成一段参考音频');
    setVoiceCloneDraft((current) => ({
      ...current,
      assetId: reference.id,
      name: current.name || `${String(reference.name || '').replace(/\.[^.]+$/, '')}-克隆音色`
    }));
    setVoiceCloneOpen(true);
  }

  async function runVoiceClone(addToTimeline) {
    if (!projectId || voiceCloneBusy) return;
    const controller = new AbortController();
    voiceCloneAbortRef.current = controller;
    setVoiceCloneBusy(true);
    try {
      const created = await cloneAICutVoice(projectId, { ...voiceCloneDraft, signal: controller.signal });
      if (controller.signal.aborted || !created) return;
      setAssets((current) => current.some((item) => item.id === created.id) ? current : [...current, created]);
      setSelectedAssetId(created.id);
      if (addToTimeline) addAssetToTimeline(created, { startMs: previewTimeMs });
      setVoiceCloneOpen(false);
      setVoiceCloneDraft((current) => ({ ...current, name: '', text: '' }));
      message.success(addToTimeline ? '克隆音频已生成并加入播放头位置' : '克隆音频已生成并保存到项目素材');
    } catch (error) {
      if (controller.signal.aborted || error?.name === 'AbortError') message.info('已取消本次声音克隆');
      else message.error(error?.message || '声音克隆失败');
    } finally {
      if (voiceCloneAbortRef.current === controller) {
        voiceCloneAbortRef.current = null;
        setVoiceCloneBusy(false);
      }
    }
  }

  function cancelVoiceClone() {
    if (voiceCloneAbortRef.current) {
      voiceCloneAbortRef.current.abort();
      return;
    }
    setVoiceCloneOpen(false);
  }

  async function importSRT(file) {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!file) return Upload.LIST_IGNORE;
    if (!projectId) {
      message.warning('请先创建剪辑项目');
      return Upload.LIST_IGNORE;
    }
    setImportingSRT(true);
    try {
      const result = await importAICutSRT(projectId, file, timeline.revision);
      if (!result) return Upload.LIST_IGNORE;
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(result.timeline);
      message.success(`已导入 ${result.imported} 条字幕`);
    } catch (error) {
      message.error(error?.message || 'SRT 字幕导入失败');
    } finally {
      setImportingSRT(false);
    }
    return Upload.LIST_IGNORE;
  }

  function addAssetToTimeline(asset, options = {}) {
    const kind = trackKindForAsset(asset);
    commitTimeline((current) => {
      const tracks = current.tracks.map((track) => ({
        ...track,
        clips: [...(track.clips || [])]
      }));
      let track = options.trackId ? tracks.find((item) => item.id === options.trackId && item.kind === kind && !item.locked) : null;
      if (!track) track = tracks.find((item) => item.kind === kind && !item.locked);
      if (!track) {
        const compoundEntry = options.compoundId ? compoundEntries(current, options.compoundId)[0] : null;
        track = {
          id: newID(`track-${kind}`),
          kind,
          name: trackLabels[kind],
          order: tracks.length,
          muted: false,
          locked: false,
          settings: compoundEntry ? { compoundId: options.compoundId, compoundName: compoundPathForClip(compoundEntry.clip).find(item => item.id === options.compoundId)?.name || '复合片段' } : {},
          clips: []
        };
        tracks.push(track);
      }
      const appendedStart = track.clips.reduce((max, clip) => Math.max(max, Number(clip.startMs) + Number(clip.durationMs)), 0);
      const startMs = Number.isFinite(Number(options.startMs)) ? Math.max(0, Number(options.startMs)) : appendedStart;
      let addedClip = {
        id: newID('clip'),
        assetId: asset.id,
        kind: asset.kind,
        name: asset.name,
        startMs,
        durationMs: defaultDuration(asset),
        sourceStartMs: 0,
        sourceDurationMs: defaultDuration(asset),
        playbackRate: 1,
        text: ''
      };
      if (options.compoundId) {
        const compoundEntry = compoundEntries(current, options.compoundId)[0];
        if (!compoundEntry) throw new Error('复合片段不存在');
        const path = compoundPathForClip(compoundEntry.clip);
        const index = path.findIndex(item => item.id === options.compoundId);
        addedClip = applyCompoundPath(addedClip, path.slice(index));
      }
      track.clips.push(addedClip);
      return { ...current, tracks };
    });
  }

  function dropAssetOnTrack(event, track) {
    event.preventDefault();
    const assetId = event.dataTransfer.getData('application/x-aicut-asset');
    const asset = assets.find((item) => item.id === assetId);
    if (!asset) return message.warning('没有读取到要拖入的素材');
    if (track.locked) return message.warning('轨道已锁定');
    if (trackKindForAsset(asset) !== track.kind) return message.warning('素材类型与轨道不匹配');
    const scrollLeft = event.currentTarget.closest('.aicut-tracks-scroll')?.scrollLeft || 0;
    const startMs = timelineDropTime(event.clientX, event.currentTarget.getBoundingClientRect().left, scrollLeft, timelineZoom);
    addAssetToTimeline(asset, {
      trackId: track.placeholder ? '' : track.id,
      startMs
    });
    setSelectedAssetId(asset.id);
    message.success(`已把 ${asset.name} 放到 ${previewClock(startMs)}`);
  }

  function dropCatalogResourceOnTrack(event, track) {
    event.preventDefault();
    let resource;
    try {
      resource = JSON.parse(event.dataTransfer.getData('application/x-aicut-resource'));
    } catch {
      return message.warning('没有读取到要拖入的资源');
    }
    const catalogResource = catalogById(resource?.id);
    if (!catalogResource) return message.warning('该资源不在当前资源库中');
    if (track.locked) return message.warning('轨道已锁定');
    const scrollLeft = event.currentTarget.closest('.aicut-tracks-scroll')?.scrollLeft || 0;
    const startMs = timelineDropTime(event.clientX, event.currentTarget.getBoundingClientRect().left, scrollLeft, timelineZoom);
    const parameters = normalizeCatalogParameters(catalogResource.id, {});
    if (catalogResource.category === 'effect' || catalogResource.category === 'text') {
      return applyCatalogResource(catalogResource, parameters, { startMs });
    }
    if (!['filter', 'transition'].includes(catalogResource.category) || track.kind !== 'video') return message.warning('该资源只能拖到对应的画面轨道');
    const clips = [...(track.clips || [])].filter((clip) => ['video', 'image'].includes(clip.kind)).sort((left, right) => Number(left.startMs) - Number(right.startMs));
    const clip = catalogResource.category === 'transition'
      ? clips.slice(1).sort((left, right) => Math.abs(Number(left.startMs) - startMs) - Math.abs(Number(right.startMs) - startMs))[0]
      : clips.find((item) => startMs >= Number(item.startMs) && startMs <= Number(item.startMs) + Number(item.durationMs));
    if (!clip) return message.warning(catalogResource.category === 'transition' ? '转场需要拖到两个相邻片段的连接处' : '请把滤镜拖到画面片段上');
    return applyCatalogResource(catalogResource, parameters, { trackId: track.id, clipId: clip.id, startMs });
  }

  function dropAssetOnCompoundTrack(event, track, view) {
    event.preventDefault();
    const assetId = event.dataTransfer.getData('application/x-aicut-asset');
    const asset = assets.find((item) => item.id === assetId);
    if (!asset) return message.warning('没有读取到要拖入的素材');
    if (track.locked) return message.warning('轨道已锁定');
    if (trackKindForAsset(asset) !== track.kind) return message.warning('素材类型与轨道不匹配');
    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width)));
    const startMs = view.originMs + Math.round(view.durationMs * ratio);
    addAssetToTimeline(asset, { trackId: track.id, startMs, compoundId: view.id });
    setSelectedAssetId(asset.id);
    message.success(`已把 ${asset.name} 加入 ${view.name}`);
  }

  function beginLayoutResize(kind, event) {
    const shell = editorShellRef.current;
    if (!shell || event.button !== 0) return;
    const rect = shell.getBoundingClientRect();
    layoutDragRef.current = {
      kind,
      originX: event.clientX,
      originY: event.clientY,
      layout: editorLayout,
      viewport: { width: rect.width, height: rect.height }
    };
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.preventDefault();
  }

  function updateLayoutResize(event) {
    const drag = layoutDragRef.current;
    if (!drag) return;
    const next = { ...drag.layout };
    if (drag.kind === 'media') next.resourceWidth += event.clientX - drag.originX;
    if (drag.kind === 'timeline') next.timelineHeight -= event.clientY - drag.originY;
    setEditorLayout(normalizeEditorLayout(next, drag.viewport));
  }

  function finishLayoutResize(event) {
    if (!layoutDragRef.current) return;
    updateLayoutResize(event);
    layoutDragRef.current = null;
  }

  function updateClip(trackId, clipId, patch) {
    commitTimeline((current) => ({
      ...current,
      tracks: current.tracks.map((track) =>
        track.id !== trackId
          ? track
          : {
              ...track,
              clips: track.clips.map((clip) => (clip.id === clipId ? { ...clip, ...patch } : clip))
            }
      )
    }));
  }

  function applyVisualEffect() {
    if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return message.warning('请先选择画面片段');
    if (selectedClipEntry.track.locked) return message.warning('轨道已锁定');
    updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, {
      visualEffect: normalizeVisualEffect(visualEffectDraft)
    });
    setToolWorkbenchOpen('');
    message.success('画面效果已应用，可保存工程或直接导出');
  }

  function removeClip(trackId, clipId) {
    try {
      runTimelineCommand({ type: 'clip.delete', trackId, clipId });
      setSelectedClipKey('');
      setSelectedClipKeys([]);
    } catch (error) {
      message.warning(error?.message || '删除失败');
    }
  }

  function duplicateSelectedClip() {
    if (!selectedClipEntry) return;
    const { track, clip } = selectedClipEntry;
    const duplicateId = newID('clip');
    try {
      runTimelineCommand({
        type: 'clip.duplicate',
        trackId: track.id,
        clipId: clip.id,
        newId: duplicateId
      });
      setSelectedClipKey(`${track.id}::${duplicateId}`);
      setSelectedClipKeys([`${track.id}::${duplicateId}`]);
    } catch (error) {
      message.warning(error?.message || '复制失败');
    }
  }

  function selectClip(key, additive = false) {
    if (!additive) {
      setSelectedClipKey(key);
      setSelectedClipKeys([key]);
      return;
    }
    setSelectedClipKeys((current) => {
      const exists = current.includes(key);
      const next = exists ? current.filter((item) => item !== key) : [...current, key];
      setSelectedClipKey(exists ? next.at(-1) || '' : key);
      return next;
    });
  }

  function toggleTimelineSelection(keys) {
    const selectionKeys = [...new Set((keys || []).filter(Boolean))];
    if (!selectionKeys.length) return;
    setSelectedClipKeys((current) => {
      const removesSelection = selectionKeys.every((key) => current.includes(key));
      const next = removesSelection
        ? current.filter((key) => !selectionKeys.includes(key))
        : [...new Set([...current, ...selectionKeys])];
      setSelectedClipKey(removesSelection ? next.at(-1) || '' : selectionKeys.at(-1));
      return next;
    });
  }

  function copySelectedClips() {
    try {
      clipboardRef.current = createAICutClipboard(timelineRef.current, selectedClipKeys);
      setClipboardCount(clipboardRef.current.items.length);
      message.success(`已复制 ${clipboardRef.current.items.length} 个片段`);
    } catch (error) {
      message.warning(error?.message || '复制失败');
    }
  }

  function cutSelectedClips() {
    try {
      clipboardRef.current = createAICutClipboard(timelineRef.current, selectedClipKeys);
      setClipboardCount(clipboardRef.current.items.length);
      runTimelineCommand({
        type: 'clip.delete_many',
        clipKeys: selectedClipKeys,
        ripple: rippleMode
      });
      setSelectedClipKey('');
      setSelectedClipKeys([]);
      message.success(`已剪切 ${clipboardRef.current.items.length} 个片段`);
    } catch (error) {
      message.warning(error?.message || '剪切失败');
    }
  }

  function pasteClips() {
    try {
      const result = runTimelineCommand({
        type: 'clip.paste',
        clipboard: clipboardRef.current,
        atMs: previewTimeMs
      });
      const pastedKeys = result.metadata?.pastedKeys || [];
      setSelectedClipKeys(pastedKeys);
      setSelectedClipKey(pastedKeys[0] || '');
      message.success(`已在播放头粘贴 ${pastedKeys.length} 个片段`);
    } catch (error) {
      message.warning(error?.message || '粘贴失败');
    }
  }

  function copySelectedKeyframes(keyframeIds) {
    if (!selectedClipEntry) return message.warning('请先选择源片段');
    try {
      const requested = new Set((keyframeIds || []).map(String));
      const frames = (selectedClipEntry.clip.keyframes || [])
        .filter(frame => requested.has(String(frame.id)))
        .sort((left, right) => Number(left.offsetMs) - Number(right.offsetMs))
        .map(frame => ({
          offsetMs: Number(frame.offsetMs),
          values: { ...(frame.values || {}) },
          easing: frame.easing ? { ...frame.easing } : undefined
        }));
      if (!frames.length || frames.length !== requested.size) throw new Error('请选择有效的关键帧');
      keyframeClipboardRef.current = { frames, sourceClipId: selectedClipEntry.clip.id };
      setKeyframeClipboardCount(frames.length);
      message.success(`已复制 ${frames.length} 个关键帧`);
    } catch (error) {
      message.warning(error?.message || '复制关键帧失败');
    }
  }

  function pasteKeyframesAtPlayhead() {
    if (!selectedClipEntry) return message.warning('请先选择目标片段');
    try {
      const anchorMs = Math.round(previewTimeMs - Number(selectedClipEntry.clip.startMs || 0));
      const result = runTimelineCommand({
        type: 'clip.keyframes_paste',
        trackId: selectedClipEntry.track.id,
        clipId: selectedClipEntry.clip.id,
        anchorMs,
        frames: keyframeClipboardRef.current?.frames || []
      });
      message.success(`已在播放头粘贴 ${result.metadata?.keyframeIds?.length || 0} 个关键帧`);
    } catch (error) {
      message.warning(error?.message || '粘贴关键帧失败');
    }
  }

  function deleteSelectedClips() {
    try {
      const result = runTimelineCommand({
        type: 'clip.delete_many',
        clipKeys: selectedClipKeys,
        ripple: rippleMode
      });
      setSelectedClipKey('');
      setSelectedClipKeys([]);
      message.success(`${rippleMode ? '波纹' : ''}删除了 ${result.metadata?.deletedCount || 0} 个片段`);
    } catch (error) {
      message.warning(error?.message || '删除失败');
    }
  }

  function nudgeSelectedClip(frames) {
    if (!selectedClipEntry) return message.warning('请先选择片段');
    try {
      runTimelineCommand({
        type: 'clip.nudge',
        trackId: selectedClipEntry.track.id,
        clipId: selectedClipEntry.clip.id,
        frames,
        fps: 30,
        linked: linkedMove
      });
    } catch (error) {
      message.warning(error?.message || '逐帧移动失败');
    }
  }

  function jumpEditPoint(direction) {
    const point = findTimelineEditPoint(timelineRef.current, previewTimeMs, direction);
    seekPreview(point);
  }

  function addMarkerAtPlayhead() {
    try {
      const number = (timelineRef.current.markers?.length || 0) + 1;
      runTimelineCommand({
        type: 'timeline.marker_add',
        timeMs: previewTimeMs,
        label: `标记 ${number}`,
        color: '#f59e0b'
      });
      message.success('已在播放头添加标记点');
    } catch (error) {
      message.warning(error?.message || '添加标记点失败');
    }
  }

  function removeMarkerAtPlayhead() {
    const marker = (timelineRef.current.markers || []).find((item) => Number(item.timeMs) === Math.round(previewTimeMs));
    if (!marker) return message.warning('播放头位置没有标记点');
    try {
      runTimelineCommand({
        type: 'timeline.marker_remove',
        markerId: marker.id
      });
      message.success('已删除标记点');
    } catch (error) {
      message.warning(error?.message || '删除标记点失败');
    }
  }

  function addKeyframeAtPlayhead() {
    if (!selectedClipEntry) return message.warning('请先选择片段');
    const clip = selectedClipEntry.clip;
    const offsetMs = Math.round(previewTimeMs - Number(clip.startMs || 0));
    const transform = clip.canvasTransform || {
      x: 0,
      y: 0,
      scaleX: 1,
      scaleY: 1,
      rotation: 0,
      opacity: 1
    };
    try {
      runTimelineCommand({
        type: 'clip.keyframe_upsert',
        trackId: selectedClipEntry.track.id,
        clipId: clip.id,
        offsetMs,
        values: { ...transform, volume: Number(clip.volume ?? 1) }
      });
      message.success('已在播放头写入关键帧');
    } catch (error) {
      message.warning(error?.message || '添加关键帧失败');
    }
  }

  function removeKeyframeAtPlayhead() {
    if (!selectedClipEntry) return message.warning('请先选择片段');
    const offsetMs = Math.round(previewTimeMs - Number(selectedClipEntry.clip.startMs || 0));
    const frame = (selectedClipEntry.clip.keyframes || []).find((item) => Number(item.offsetMs) === offsetMs);
    if (!frame) return message.warning('播放头位置没有关键帧');
    try {
      runTimelineCommand({
        type: 'clip.keyframe_remove',
        trackId: selectedClipEntry.track.id,
        clipId: selectedClipEntry.clip.id,
        keyframeId: frame.id
      });
      message.success('已删除关键帧');
    } catch (error) {
      message.warning(error?.message || '删除关键帧失败');
    }
  }

  function splitClip(trackId, clip) {
    try {
      runTimelineCommand({ type: 'clip.split', trackId, clipId: clip.id });
    } catch (error) {
      message.warning(error?.message || '分割失败');
    }
  }

  function splitAtPlayhead() {
    try {
      runTimelineCommand({
        type: 'clip.split_at_playhead',
        timeMs: previewTimeMs
      });
      setSelectedClipKey('');
      setSelectedClipKeys([]);
    } catch (error) {
      message.warning(error?.message || '播放头分割失败');
    }
  }

  function splitClipAtPointer(track, clip, event) {
    if (!bladeMode || track.locked) return false;
    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width)));
    const offsetMs = Math.round(Number(clip.durationMs) * ratio);
    try {
      runTimelineCommand({
        type: 'clip.split_at',
        trackId: track.id,
        clipId: clip.id,
        offsetMs
      });
      setSelectedClipKey('');
      setSelectedClipKeys([]);
    } catch (error) {
      message.warning(error?.message || '分割失败');
    }
    return true;
  }

  function splitCompoundClip(compoundId, timeMs) {
    try {
      const result = runTimelineCommand({
        type: 'clip.compound_split',
        compoundId,
        timeMs,
        rightCompoundId: newID('compound')
      });
      setActiveCompoundId('');
      setSelectedClipKey('');
      setSelectedClipKeys([]);
      message.success(`已分成左右两个复合片段，各含 ${result.metadata.leftMemberKeys.length} / ${result.metadata.rightMemberKeys.length} 个成员`);
      return result;
    } catch (error) {
      message.warning(error?.message || '复合片段分割失败');
      return null;
    }
  }

  function splitCompoundAtPointer(track, clip, event) {
    if (!bladeMode || track.locked) return false;
    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width)));
    const timeMs = Math.round((Number(clip.startMs) || 0) + Number(clip.durationMs) * ratio);
    splitCompoundClip(clip.compoundId, timeMs);
    return true;
  }

  function trimClip(trackId, clipId, edge) {
    try {
      const track = timelineRef.current.tracks.find((item) => item.id === trackId);
      const clip = track?.clips?.find((item) => item.id === clipId);
      if (!clip) throw new Error('片段不存在');
      const startMs = Number(clip.startMs) || 0;
      const durationMs = Math.max(100, Number(clip.durationMs) || 100);
      const edgeTimeMs = edge === 'left' ? startMs + Math.min(100, durationMs - 100) : startMs + Math.max(100, durationMs - 100);
      runTimelineCommand({
        type: 'clip.resize',
        trackId,
        clipId,
        edge,
        edgeTimeMs,
        ripple: rippleMode
      });
    } catch (error) {
      message.warning(error?.message || '裁剪失败');
    }
  }

  function changeClipSpeed(trackId, clipId, rate) {
    try {
      runTimelineCommand({
        type: 'clip.playback_rate_set',
        trackId,
        clipId,
        playbackRate: rate
      });
    } catch (error) {
      message.warning(error?.message || '变速失败');
    }
  }

  function applyClipSpeedCurve(trackId, clipId, presetId) {
    try {
      const preset = speedCurvePresets.find((item) => item.id === presetId);
      if (!preset) throw new Error('速度曲线预设不存在');
      runTimelineCommand({
        type: 'clip.speed_curve_set',
        trackId,
        clipId,
        points: preset.points
      });
      message.success(`已应用${preset.label}曲线`);
    } catch (error) {
      message.warning(error?.message || '曲线变速失败');
    }
  }

  function runVolumeEnvelopeCommand(type, payload) {
    if (!selectedClipEntry) return;
    try {
      runTimelineCommand({
        type,
        trackId: selectedClipEntry.track.id,
        clipId: selectedClipEntry.clip.id,
        ...payload
      });
    } catch (error) {
      message.warning(error?.message || '音量包络调整失败');
    }
  }

  function setAudioFades(trackId, clipId, fadeInMs, fadeOutMs) {
    try {
      return runTimelineCommand({ type: 'clip.audio_fade_set', trackId, clipId, fadeInMs, fadeOutMs });
    } catch (error) {
      message.warning(error?.message || '音频淡入淡出调整失败');
      return null;
    }
  }

  function moveClip(trackId, clipId, startMs) {
    try {
      runTimelineCommand({
        type: 'clip.move',
        trackId,
        clipId,
        startMs,
        magnetic: magneticSnap,
        linked: linkedMove,
        thresholdMs: 150,
        playheadMs: previewTimeMs
      });
    } catch (error) {
      message.warning(error?.message || '移动片段失败');
    }
  }

  function moveCompoundClip(compoundId, startMs) {
    try {
      runTimelineCommand({
        type: 'clip.compound_move',
        compoundId,
        startMs
      });
    } catch (error) {
      message.warning(error?.message || '移动复合片段失败');
    }
  }

  function trimCompoundClip(compoundId, edge, edgeTimeMs) {
    try {
      return runTimelineCommand({
        type: 'clip.compound_trim',
        compoundId,
        edge,
        edgeTimeMs
      });
    } catch (error) {
      message.warning(error?.message || '裁剪复合片段失败');
      return null;
    }
  }

  function linkSelectedClips() {
    if (!selectedClipEntry) return message.warning('请先选择要联动的片段');
    const startMs = Number(selectedClipEntry.clip.startMs) || 0;
    const clipKeys = selectedClipKeys.length >= 2 ? selectedClipKeys : timelineRef.current.tracks.flatMap((track) => (track.locked ? [] : (track.clips || []).filter((clip) => Number(clip.startMs) === startMs).map((clip) => `${track.id}::${clip.id}`)));
    if (clipKeys.length < 2) return message.warning('同一开始时间至少要有两个未锁定片段才能建立联动');
    try {
      const result = runTimelineCommand({
        type: 'clip.link',
        clipKeys,
        groupId: newID('link')
      });
      if (result.changed) message.success(`已联动 ${clipKeys.length} 个片段`);
    } catch (error) {
      message.warning(error?.message || '建立联动失败');
    }
  }

  function unlinkSelectedClips() {
    if (!selectedClipEntry) return message.warning('请先选择已联动片段');
    try {
      const result = runTimelineCommand({
        type: 'clip.unlink',
        clipKeys: selectedClipKeys.length ? selectedClipKeys : [selectedClipKey]
      });
      if (result.changed) message.success('已取消该组联动');
      else message.info('当前片段没有联动组');
    } catch (error) {
      message.warning(error?.message || '取消联动失败');
    }
  }

  function createCompoundClip() {
    if (selectedClipKeys.length < 2) return message.warning('请至少选择两个片段再创建复合片段');
    try {
      const result = runTimelineCommand({
        type: 'clip.compound_create',
        clipKeys: selectedClipKeys,
        compoundId: newID('compound'),
        name: `复合片段 ${new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
      });
      if (result.changed) message.success(`已创建包含 ${selectedClipKeys.length} 个成员的复合片段`);
    } catch (error) {
      message.warning(error?.message || '创建复合片段失败');
    }
  }

  function compoundMemberKeys(compoundId) {
    return compoundEntries(timelineRef.current, compoundId).map(entry => entry.key);
  }

  function compoundParentId(compoundId) {
    const entry = compoundEntries(timelineRef.current, compoundId)[0];
    if (!entry) return '';
    const path = compoundPathForClip(entry.clip);
    const index = path.findIndex(item => item.id === compoundId);
    return index >= 0 ? path[index + 1]?.id || '' : '';
  }

  function addCompoundTrack(compoundId, kind) {
    try {
      const result = runTimelineCommand({
        type: 'compound.track_add',
        compoundId,
        kind,
        trackId: newID(`track-${kind}`)
      });
      if (result.changed) message.success('已新增组内轨道，可把左侧素材拖入该轨道');
    } catch (error) {
      message.warning(error?.message || '新增组内轨道失败');
    }
  }

  function removeCompoundTrack(compoundId, trackId) {
    try {
      const result = runTimelineCommand({ type: 'compound.track_remove', compoundId, trackId });
      if (result.changed) message.success('已删除空的组内轨道');
    } catch (error) {
      message.warning(error?.message || '删除组内轨道失败');
    }
  }

  function reorderCompoundTracks(compoundId, orderedTrackIds) {
    try {
      const result = runTimelineCommand({ type: 'compound.track_reorder', compoundId, orderedTrackIds });
      if (result.changed) message.success('组内轨道顺序已更新');
    } catch (error) {
      message.warning(error?.message || '调整组内轨道顺序失败');
    }
  }

  function expandSelectedCompound() {
    const compoundId = selectedClipEntry?.clip?.compoundId;
    if (!compoundId) return message.info('当前片段不属于复合片段');
    const keys = compoundMemberKeys(compoundId);
    setSelectedClipKeys(keys);
    message.success(`已展开选择 ${keys.length} 个复合成员`);
  }

  function dissolveCompound(compoundId) {
    const keys = compoundMemberKeys(compoundId);
    if (!keys.length) return message.info('复合片段不存在或已经拆散');
    const parentId = compoundParentId(compoundId);
    try {
      const result = runTimelineCommand({
        type: 'clip.compound_dissolve',
        compoundId,
        clipKeys: [keys[0]]
      });
      setSelectedClipKeys(result.metadata.memberKeys || keys);
      setActiveCompoundId(parentId);
      if (result.changed) message.success('已拆散复合片段，原成员和时间位置保持不变');
    } catch (error) {
      message.warning(error?.message || '拆散复合片段失败');
    }
  }

  function dissolveSelectedCompound() {
    const compoundId = selectedClipEntry?.clip?.compoundId;
    if (!compoundId) return message.info('当前片段不属于复合片段');
    dissolveCompound(compoundId);
  }

  function togglePreviewPlayback() {
    if (durationMs <= 0) return message.warning('请先把素材加入时间线');
    if (!previewPlaying && previewTimeMs >= durationMs) setPreviewTimeMs(0);
    setPreviewPlaying((value) => !value);
  }

  function seekPreview(value) {
    setPreviewPlaying(false);
    setPreviewTimeMs(seekAICutPreviewTime(value, durationMs));
  }

  function audioFadeValuesForPointer(drag, clientX) {
    const deltaMs = Math.round(timelinePixelDeltaToTime(clientX - drag.originX, timelineZoom) / 10) * 10;
    const clampFade = (value) => Math.max(0, Math.min(drag.durationMs, Math.round(value)));
    return drag.edge === 'in'
      ? { fadeInMs: clampFade(drag.fadeInMs + deltaMs), fadeOutMs: drag.fadeOutMs }
      : { fadeInMs: drag.fadeInMs, fadeOutMs: clampFade(drag.fadeOutMs - deltaMs) };
  }

  function beginAudioFadeDrag(event, track, clip, edge) {
    if (track.locked || event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    const key = `${track.id}::${clip.id}`;
    const drag = {
      key,
      trackId: track.id,
      clipId: clip.id,
      edge,
      originX: event.clientX,
      durationMs: Math.max(1, Number(clip.durationMs) || 1),
      fadeInMs: Math.max(0, Number(clip.fadeInMs) || 0),
      fadeOutMs: Math.max(0, Number(clip.fadeOutMs) || 0)
    };
    audioFadeDragRef.current = drag;
    setAudioFadePreview({ key, fadeInMs: drag.fadeInMs, fadeOutMs: drag.fadeOutMs });
    selectClip(key);
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function updateAudioFadeDrag(event) {
    const drag = audioFadeDragRef.current;
    if (!drag) return;
    setAudioFadePreview({ key: drag.key, ...audioFadeValuesForPointer(drag, event.clientX) });
  }

  function finishAudioFadeDrag(event) {
    const drag = audioFadeDragRef.current;
    if (!drag) return;
    const values = audioFadeValuesForPointer(drag, event.clientX);
    audioFadeDragRef.current = null;
    setAudioFadePreview(null);
    if (values.fadeInMs !== drag.fadeInMs || values.fadeOutMs !== drag.fadeOutMs) {
      setAudioFades(drag.trackId, drag.clipId, values.fadeInMs, values.fadeOutMs);
    }
  }

  function nudgeAudioFade(event, track, clip, edge) {
    if (!['ArrowLeft', 'ArrowRight'].includes(event.key) || track.locked) return;
    event.preventDefault();
    event.stopPropagation();
    const durationMs = Math.max(1, Number(clip.durationMs) || 1);
    const fadeInMs = Math.max(0, Number(clip.fadeInMs) || 0);
    const fadeOutMs = Math.max(0, Number(clip.fadeOutMs) || 0);
    const direction = event.key === 'ArrowRight' ? 1 : -1;
    const nextFadeIn = edge === 'in' ? Math.max(0, Math.min(durationMs, fadeInMs + direction * 100)) : fadeInMs;
    const nextFadeOut = edge === 'out' ? Math.max(0, Math.min(durationMs, fadeOutMs - direction * 100)) : fadeOutMs;
    setAudioFades(track.id, clip.id, nextFadeIn, nextFadeOut);
  }

  function beginClipDrag(event, track, clip) {
    if (track.locked || event.button !== 0) return;
    if (event.target.closest('.aicut-transition-chip')) {
      beginTransitionDrag(event, track, clip);
      return;
    }
    if (event.target.closest('button,input,.ant-select')) return;
    if (splitClipAtPointer(track, clip, event)) {
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    const rect = event.currentTarget.getBoundingClientRect();
    if (selectedClipKey === `${track.id}::${clip.id}` && event.clientX - rect.left <= 8) {
      beginClipTrim(event, track, clip, 'left');
      return;
    }
    if (selectedClipKey === `${track.id}::${clip.id}` && rect.right - event.clientX <= 8) {
      beginClipTrim(event, track, clip, 'right');
      return;
    }
    clipDragRef.current = {
      trackId: track.id,
      clipId: clip.id,
      originX: event.clientX,
      originStartMs: Number(clip.startMs) || 0
    };
    setDragPreview({
      key: `${track.id}::${clip.id}`,
      startMs: Number(clip.startMs) || 0
    });
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function beginCompoundDrag(event, track, clip) {
    if (event.button !== 0) return;
    if (splitCompoundAtPointer(track, clip, event)) {
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    const rect = event.currentTarget.getBoundingClientRect();
    if (event.clientX - rect.left <= 8) {
      beginCompoundTrim(event, track, clip, 'left');
      return;
    }
    if (rect.right - event.clientX <= 8) {
      beginCompoundTrim(event, track, clip, 'right');
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    compoundDragMovedRef.current = false;
    const key = `${track.id}::${clip.id}`;
    clipDragRef.current = {
      compoundId: clip.compoundId,
      trackId: track.id,
      clipId: clip.id,
      key,
      originX: event.clientX,
      originStartMs: Number(clip.startMs) || 0
    };
    setDragPreview({ key, startMs: Number(clip.startMs) || 0 });
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function beginCompoundTrim(event, track, clip, edge) {
    if (event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    compoundDragMovedRef.current = false;
    const key = `${track.id}::${clip.id}`;
    trimDragRef.current = {
      compoundId: clip.compoundId,
      trackId: track.id,
      clipId: clip.id,
      key,
      edge,
      originX: event.clientX,
      startMs: Number(clip.startMs) || 0,
      durationMs: Math.max(100, Number(clip.durationMs) || 100)
    };
    setTrimPreview({
      key,
      startMs: Number(clip.startMs) || 0,
      durationMs: Math.max(100, Number(clip.durationMs) || 100)
    });
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function nudgeCompoundTrim(event, clip, edge) {
    if (!['ArrowLeft', 'ArrowRight'].includes(event.key)) return;
    event.preventDefault();
    event.stopPropagation();
    const startMs = Number(clip.startMs) || 0;
    const endMs = startMs + Math.max(100, Number(clip.durationMs) || 100);
    const inward = (edge === 'left' && event.key === 'ArrowRight') || (edge === 'right' && event.key === 'ArrowLeft');
    if (!inward) {
      message.info('复合裁剪只能收紧边界；需要恢复请使用撤销');
      return;
    }
    compoundDragMovedRef.current = true;
    trimCompoundClip(clip.compoundId, edge, edge === 'left' ? startMs + 100 : endMs - 100);
  }

  function updateClipDrag(event) {
    if (transitionDragRef.current) {
      updateTransitionDrag(event);
      return;
    }
    if (audioFadeDragRef.current) {
      updateAudioFadeDrag(event);
      return;
    }
    if (trimDragRef.current) {
      updateClipTrim(event);
      return;
    }
    const drag = clipDragRef.current;
    if (!drag) return;
    if (drag.compoundId && Math.abs(event.clientX - drag.originX) > 3) compoundDragMovedRef.current = true;
    setDragPreview({
      key: drag.key || `${drag.trackId}::${drag.clipId}`,
      startMs: draggedTimelineStart(drag.originStartMs, event.clientX - drag.originX, timelineZoom)
    });
  }

  function finishClipDrag(event) {
    if (transitionDragRef.current) {
      finishTransitionDrag(event);
      return;
    }
    if (audioFadeDragRef.current) {
      finishAudioFadeDrag(event);
      return;
    }
    if (trimDragRef.current) {
      finishClipTrim(event);
      return;
    }
    const drag = clipDragRef.current;
    if (!drag) return;
    const nextStart = draggedTimelineStart(drag.originStartMs, event.clientX - drag.originX, timelineZoom);
    clipDragRef.current = null;
    setDragPreview(null);
    if (nextStart !== drag.originStartMs) {
      if (drag.compoundId) moveCompoundClip(drag.compoundId, nextStart);
      else moveClip(drag.trackId, drag.clipId, nextStart);
    }
  }

  function beginClipTrim(event, track, clip, edge) {
    if (track.locked || event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    trimDragRef.current = {
      trackId: track.id,
      clipId: clip.id,
      edge,
      originX: event.clientX,
      startMs: Number(clip.startMs) || 0,
      durationMs: Math.max(100, Number(clip.durationMs) || 100)
    };
    setSelectedClipKey(`${track.id}::${clip.id}`);
    setSelectedClipKeys([`${track.id}::${clip.id}`]);
    setTrimPreview({
      key: `${track.id}::${clip.id}`,
      startMs: Number(clip.startMs) || 0,
      durationMs: Math.max(100, Number(clip.durationMs) || 100)
    });
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.currentTarget.addEventListener(
      'lostpointercapture',
      () => {
        trimDragRef.current = null;
        setTrimPreview(null);
      },
      { once: true }
    );
  }

  function updateClipTrim(event) {
    const drag = trimDragRef.current;
    if (!drag) return;
    const deltaMs = timelinePixelDeltaToTime(event.clientX - drag.originX, timelineZoom);
    if (drag.compoundId && Math.abs(event.clientX - drag.originX) > 3) compoundDragMovedRef.current = true;
    const key = drag.key || `${drag.trackId}::${drag.clipId}`;
    if (drag.edge === 'left') {
      const amount = Math.max(0, Math.min(drag.durationMs - 100, deltaMs));
      setTrimPreview({
        key,
        startMs: drag.startMs + amount,
        durationMs: drag.durationMs - amount
      });
    } else {
      const durationMs = Math.max(100, Math.min(drag.durationMs, drag.durationMs + deltaMs));
      setTrimPreview({
        key,
        startMs: drag.startMs,
        durationMs
      });
    }
  }

  function finishClipTrim(event) {
    const drag = trimDragRef.current;
    if (!drag) return;
    updateClipTrim(event);
    const deltaMs = timelinePixelDeltaToTime(event.clientX - drag.originX, timelineZoom);
    const edgeTimeMs = drag.edge === 'left' ? drag.startMs + Math.max(0, Math.min(drag.durationMs - 100, deltaMs)) : drag.startMs + Math.max(100, Math.min(drag.durationMs, drag.durationMs + deltaMs));
    trimDragRef.current = null;
    setTrimPreview(null);
    const unchanged = drag.edge === 'left' ? edgeTimeMs === drag.startMs : edgeTimeMs === drag.startMs + drag.durationMs;
    if (unchanged) return;
    try {
      if (drag.compoundId) trimCompoundClip(drag.compoundId, drag.edge, edgeTimeMs);
      else {
        runTimelineCommand({
          type: 'clip.resize',
          trackId: drag.trackId,
          clipId: drag.clipId,
          edge: drag.edge,
          edgeTimeMs,
          ripple: rippleMode
        });
      }
    } catch (error) {
      message.warning(error?.message || '裁剪失败');
    }
  }

  function beginTransitionDrag(event, track, clip) {
    if (track.locked || event.button !== 0) return;
    const ordered = [...(track.clips || [])].sort((a, b) => Number(a.startMs) - Number(b.startMs));
    const index = ordered.findIndex((item) => item.id === clip.id);
    if (index <= 0) return;
    const previous = ordered[index - 1];
    const maximumMs = Math.max(100, Math.floor(Math.min(Number(previous.durationMs), Number(clip.durationMs)) / 2));
    const durationMs = Math.min(maximumMs, Math.max(100, Math.round(Number(clip.transitionMs) || 500)));
    event.preventDefault();
    event.stopPropagation();
    const handle = event.target.closest('.aicut-transition-handle, .aicut-transition-chip');
    handle?.classList.add('aicut-transition-handle');
    transitionDragRef.current = {
      trackId: track.id,
      clipId: clip.id,
      originX: event.clientX,
      durationMs,
      maximumMs,
      handle
    };
    setSelectedClipKey(`${track.id}::${clip.id}`);
    setSelectedClipKeys([`${track.id}::${clip.id}`]);
    setTransitionPreview({ key: `${track.id}::${clip.id}`, durationMs });
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.currentTarget.addEventListener(
      'lostpointercapture',
      () => {
        transitionDragRef.current = null;
        setTransitionPreview(null);
      },
      { once: true }
    );
  }

  function updateTransitionDrag(event) {
    const drag = transitionDragRef.current;
    if (!drag) return;
    event.preventDefault();
    event.stopPropagation();
    const durationMs = draggedTransitionDuration(drag.durationMs, event.clientX - drag.originX, timelineZoom, drag.maximumMs);
    if (drag.handle) {
      drag.handle.style.width = `${Math.max(18, timelineTimeToPixels(durationMs, timelineZoom))}px`;
      drag.handle.title = `拖动调整转场时长：${(durationMs / 1000).toFixed(2)} 秒`;
    }
    setTransitionPreview({
      key: `${drag.trackId}::${drag.clipId}`,
      durationMs
    });
  }

  function finishTransitionDrag(event) {
    const drag = transitionDragRef.current;
    if (!drag) return;
    const durationMs = draggedTransitionDuration(drag.durationMs, event.clientX - drag.originX, timelineZoom, drag.maximumMs);
    const entry = timelineRef.current.tracks.find((track) => track.id === drag.trackId)?.clips?.find((clip) => clip.id === drag.clipId);
    transitionDragRef.current = null;
    setTransitionPreview(null);
    event.preventDefault();
    event.stopPropagation();
    if (!entry?.transition || durationMs === Number(entry.transitionMs)) return;
    try {
      runTimelineCommand({
        type: 'clip.transition_set',
        trackId: drag.trackId,
        clipId: drag.clipId,
        transition: entry.transition,
        transitionMs: durationMs
      });
    } catch (error) {
      message.warning(error?.message || '转场时长调整失败');
    }
  }

  function cancelTransitionDrag(event) {
    transitionDragRef.current = null;
    setTransitionPreview(null);
    event?.stopPropagation?.();
  }

  function seekFromTrack(event) {
    if (event.target !== event.currentTarget) return;
    const rect = event.currentTarget.getBoundingClientRect();
    seekPreview(timelinePixelsToTime(event.clientX - rect.left, timelineZoom));
  }

  function stopTimelineLassoAutoScroll(drag = timelineLassoRef.current) {
    if (!drag?.autoScrollFrame) return;
    window.cancelAnimationFrame(drag.autoScrollFrame);
    drag.autoScrollFrame = 0;
  }

  function refreshTimelineLasso(drag, scheduleNext = true) {
    const canvas = timelineCanvasRef.current;
    const scroller = timelineScrollRef.current;
    if (!drag || timelineLassoRef.current !== drag || !canvas || !scroller) return;
    const viewport = scroller.getBoundingClientRect();
    const nextScrollLeft = nextLassoAutoScrollLeft({
      clientX: drag.clientX,
      viewportLeft: viewport.left,
      viewportRight: viewport.right,
      scrollLeft: scroller.scrollLeft,
      scrollWidth: scroller.scrollWidth,
      viewportWidth: scroller.clientWidth,
    });
    const moved = Math.abs(nextScrollLeft - scroller.scrollLeft) >= 1;
    if (moved) scroller.scrollLeft = nextScrollLeft;
    const canvasRect = canvas.getBoundingClientRect();
    drag.current = { x: drag.clientX - canvasRect.left, y: drag.clientY - canvasRect.top };
    setTimelineLasso(normalizeLassoRect(drag.start, drag.current));
    if (scheduleNext && moved && !drag.autoScrollFrame) {
      drag.autoScrollFrame = window.requestAnimationFrame(() => {
        drag.autoScrollFrame = 0;
        refreshTimelineLasso(drag, true);
      });
    }
  }

  function beginTimelineLasso(event) {
    if (event.button !== 0 || event.target !== event.currentTarget || !timelineCanvasRef.current) return;
    const canvasRect = timelineCanvasRef.current.getBoundingClientRect();
    const point = { x: event.clientX - canvasRect.left, y: event.clientY - canvasRect.top };
    timelineLassoRef.current = {
      start: point,
      current: point,
      clientX: event.clientX,
      clientY: event.clientY,
      additive: event.shiftKey || event.ctrlKey || event.metaKey,
      pointerId: event.pointerId,
      autoScrollFrame: 0,
    };
    setTimelineLasso(normalizeLassoRect(point, point));
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function updateTimelineLasso(event) {
    const drag = timelineLassoRef.current;
    if (!drag || drag.pointerId !== event.pointerId || !timelineCanvasRef.current) return;
    drag.clientX = event.clientX;
    drag.clientY = event.clientY;
    refreshTimelineLasso(drag, true);
  }

  function finishTimelineLasso(event) {
    const drag = timelineLassoRef.current;
    if (!drag || drag.pointerId !== event.pointerId || !timelineCanvasRef.current) return;
    stopTimelineLassoAutoScroll(drag);
    drag.clientX = event.clientX;
    drag.clientY = event.clientY;
    refreshTimelineLasso(drag, false);
    timelineLassoRef.current = null;
    setTimelineLasso(null);
    if (!lassoGestureMoved(drag.start, drag.current)) {
      const rowRect = event.currentTarget.getBoundingClientRect();
      seekPreview(timelinePixelsToTime(event.clientX - rowRect.left, timelineZoom));
      return;
    }
    const canvasRect = timelineCanvasRef.current.getBoundingClientRect();
    const selection = normalizeLassoRect(drag.start, drag.current);
    const items = [...timelineCanvasRef.current.querySelectorAll('.aicut-clip[data-lasso-keys]')].map(element => {
      const rect = element.getBoundingClientRect();
      let keys = [];
      try { keys = JSON.parse(element.dataset.lassoKeys || '[]'); } catch { keys = []; }
      return {
        keys,
        locked: element.dataset.lassoLocked === 'true',
        rect: { left: rect.left - canvasRect.left, top: rect.top - canvasRect.top, right: rect.right - canvasRect.left, bottom: rect.bottom - canvasRect.top }
      };
    });
    const keys = selectLassoClipKeys(items, selection);
    const nextKeys = drag.additive ? [...new Set([...selectedClipKeys, ...keys])] : keys;
    setSelectedClipKeys(nextKeys);
    setSelectedClipKey(nextKeys[0] || '');
    if (keys.length) message.info(`已框选 ${keys.length} 个片段`);
  }

  function cancelTimelineLasso() {
    stopTimelineLassoAutoScroll();
    timelineLassoRef.current = null;
    setTimelineLasso(null);
  }

  function markClip(trackId, clipId, status) {
    try {
      commitTimeline((current) => markTimelineClip(current, trackId, clipId, status));
    } catch (error) {
      message.warning(error?.message || '标记失败');
    }
  }

  function persistSubtitleSettings(patch) {
    const subtitleTrack = timeline.tracks.find((track) => track.kind === 'subtitle' && !track.locked);
    if (!subtitleTrack) return message.warning('请先新建字幕轨');
    commitTimeline((current) => updateTimelineTrackSettings(current, subtitleTrack.id, patch));
  }

  function commitTimeline(update) {
    if (activeTimelineScopeRef.current !== `${projectId}:${activeTimelineId}` || loadedTimelineProjectId !== `${projectId}:${activeTimelineId}`) { message.warning('请等待当前时间线读取完成'); return; }
    setTimeline((current) => {
      const next = typeof update === 'function' ? update(current) : update;
      if (!next || next === current) return current;
      undoRef.current = [...undoRef.current.slice(-49), current];
      redoRef.current = [];
      editSequenceRef.current += 1;
      timelineRef.current = next;
      setHistoryVersion((value) => value + 1);
      return next;
    });
    setSaveState('dirty');
  }

  function runTimelineCommand(command) {
    if (activeTimelineScopeRef.current !== `${projectId}:${activeTimelineId}`) throw new Error('当前工程或时间线已切换，请重新操作');
    const result = executeAICutCommand(timelineRef.current, command, {
      idFactory: (prefix) => newID(prefix || 'clip')
    });
    if (result.changed) commitTimeline(result.timeline);
    return result;
  }

  async function undoTimeline() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    const capturedEditSequence = editSequenceRef.current;
    if (undoRef.current.length) {
      if (activeTimelineScopeRef.current !== operationScope) return;
      if (capturedEditSequence !== editSequenceRef.current) { setConflictRecovery({ scope: operationScope, copyId: '' }); return; }
      setTimeline((current) => {
        const previous = undoRef.current.pop();
        if (!previous) return current;
        redoRef.current = [current, ...redoRef.current].slice(0, 50);
        editSequenceRef.current += 1;
        setHistoryVersion((value) => value + 1);
        const restored = restoreTimelineSnapshot(previous, current.revision);
        timelineRef.current = restored;
        return restored;
      });
      setSaveState('dirty');
      return;
    }
    if (!projectId || !serverHistory.canUndo) return;
    try {
      const result = await undoAICutTimeline(projectId, timeline.revision);
      if (activeTimelineScopeRef.current !== operationScope) return;
      if (capturedEditSequence !== editSequenceRef.current) { setConflictRecovery({ scope: operationScope, copyId: '' }); return; }
      setTimeline(result.timeline);
      setSaveState('saved');
      setServerHistory(result.history);
      setHistoryVersion((value) => value + 1);
    } catch (error) {
      message.error(error?.message || '撤销失败');
    }
  }

  async function redoTimeline() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    const capturedEditSequence = editSequenceRef.current;
    if (redoRef.current.length) {
      if (activeTimelineScopeRef.current !== operationScope) return;
      if (capturedEditSequence !== editSequenceRef.current) { setConflictRecovery({ scope: operationScope, copyId: '' }); return; }
      setTimeline((current) => {
        const next = redoRef.current.shift();
        if (!next) return current;
        undoRef.current = [...undoRef.current.slice(-49), current];
        editSequenceRef.current += 1;
        setHistoryVersion((value) => value + 1);
        const restored = restoreTimelineSnapshot(next, current.revision);
        timelineRef.current = restored;
        return restored;
      });
      setSaveState('dirty');
      return;
    }
    if (!projectId || !serverHistory.canRedo) return;
    try {
      const result = await redoAICutTimeline(projectId, timeline.revision);
      if (activeTimelineScopeRef.current !== operationScope) return;
      if (capturedEditSequence !== editSequenceRef.current) { setConflictRecovery({ scope: operationScope, copyId: '' }); return; }
      setTimeline(result.timeline);
      setSaveState('saved');
      setServerHistory(result.history);
      setHistoryVersion((value) => value + 1);
    } catch (error) {
      message.error(error?.message || '重做失败');
    }
  }

  function addTrack(kind) {
    commitTimeline((current) => addTimelineTrack(current, kind, () => newID(`track-${kind}`)));
  }

  function toggleTrackFlag(trackId, field) {
    commitTimeline((current) => toggleTimelineTrackFlag(current, trackId, field));
  }

  function removeTrack(track) {
    if (track.clips?.length) return message.warning('轨道中还有片段，不能删除');
    commitTimeline((current) => removeTimelineTrack(current, track.id));
  }

  function openTrackRename(track) {
    if (!track || track.placeholder) return;
    setTrackEditId(track.id);
    setTrackNameDraft(track.name || '');
  }

  function confirmTrackRename() {
    try {
      commitTimeline((current) => renameTimelineTrack(current, trackEditId, trackNameDraft));
      setTrackEditId('');
      setTrackNameDraft('');
    } catch (error) {
      message.warning(error?.message || '轨道重命名失败');
    }
  }

  function trackMenuItems(track) {
    if (!track || track.placeholder) return [];
    return [
      {
        key: 'rename',
        label: '重命名轨道',
        onClick: () => openTrackRename(track)
      },
      {
        key: 'locked',
        label: track.locked ? '解锁轨道' : '锁定轨道',
        onClick: () => toggleTrackFlag(track.id, 'locked')
      },
      {
        key: 'muted',
        label: track.muted ? '取消静音' : '静音轨道',
        onClick: () => toggleTrackFlag(track.id, 'muted')
      },
      { type: 'divider' },
      {
        key: 'delete',
        danger: true,
        label: '删除空轨道',
        onClick: () => removeTrack(track)
      }
    ];
  }

  async function saveTimeline(options = {}) {
    if (!projectId) return;
    if (loadedTimelineProjectId !== `${projectId}:${activeTimelineId}`) return null;
    const capturedScope = `${projectId}:${activeTimelineId}`;
    if (saveInFlightRef.current) {
      const saved = await saveInFlightRef.current;
      if (activeTimelineScopeRef.current !== capturedScope) return null;
      if (options.requireLatest && lastPersistedEditSequenceRef.current < editSequenceRef.current) {
        saveInFlightRef.current = null;
        return saveTimeline(options);
      }
      return saved;
    }
    setSaving(true);
    setSaveState('saving');
    const operation = (async () => {
      let saved = null;
      let reconciliation = null;
      do {
        const capturedTimeline = timelineRef.current;
        const capturedEditSequence = editSequenceRef.current;
        saved = await saveAICutTimeline(projectId, capturedTimeline);
        if (activeTimelineScopeRef.current !== capturedScope) return saved;
        reconciliation = reconcileTimelineSave(timelineRef.current, saved, capturedEditSequence, editSequenceRef.current);
        timelineRef.current = reconciliation.timeline;
        setTimeline(reconciliation.timeline);
        setSaveState(reconciliation.dirty ? 'dirty' : 'saved');
        if (!reconciliation.dirty) lastPersistedEditSequenceRef.current = editSequenceRef.current;
      } while (options.requireLatest && reconciliation?.dirty);
      const savedHistory = await getAICutTimelineHistory(projectId);
      if (activeTimelineScopeRef.current !== capturedScope) return saved;
      setServerHistory(savedHistory);
      setHistoryVersion((value) => value + 1);
      if (!options.silent) message.success('时间线已保存');
      return saved;
    })();
    saveInFlightRef.current = operation;
    try {
      return await operation;
    } catch (error) {
      if (activeTimelineScopeRef.current !== capturedScope) return null;
      setSaveState('failed');
      if (error.status === 409) setConflictRecovery({ scope: capturedScope, copyId: '' });
      message.error(error?.message || '时间线保存失败');
      return null;
    } finally {
      if (saveInFlightRef.current === operation) saveInFlightRef.current = null;
      if (activeTimelineScopeRef.current === capturedScope) setSaving(false);
    }
  }

  async function consolidateSelectedClips() {
    if (consolidateBusy) {
      consolidateAbortRef.current?.abort();
      return;
    }
    const entries = selectedClipKeys.map((key) => {
      const [trackId, clipId] = key.split('::');
      const track = timelineRef.current.tracks.find((item) => item.id === trackId);
      return { track, clip: track?.clips?.find((item) => item.id === clipId) };
    });
    if (entries.length < 2 || entries.length > 50 || entries.some((entry) => !entry.track || !entry.clip)) {
      return message.warning('请选择 2 到 50 个视频片段进行合并');
    }
    if (entries.some((entry) => entry.track.locked || entry.track.kind !== 'video' || entry.clip.kind !== 'video')) {
      return message.warning('渲染合并只支持未锁定画面轨道中的视频片段');
    }
    if (new Set(entries.map((entry) => entry.track.id)).size !== 1) {
      return message.warning('请选择同一画面轨道中的视频片段');
    }
    const controller = new AbortController();
    consolidateAbortRef.current = controller;
    setConsolidateBusy(true);
    try {
      const saved = await saveTimeline({ silent: true, requireLatest: true });
      if (!saved) throw new Error('时间线尚未保存');
      const created = await consolidateAICutClips(projectId, {
        clipIds: entries.map((entry) => entry.clip.id),
        expectedRevision: saved.revision,
        signal: controller.signal
      });
      setAssets((current) => current.some((item) => item.id === created.id) ? current : [...current, created]);
      setToolCategory('媒体');
      message.success('所选视频已渲染为新的项目素材，原片段保持不变');
    } catch (error) {
      if (error?.name === 'AbortError') message.info('已取消片段渲染合并，时间线没有改变');
      else message.error(error?.message || '片段渲染合并失败');
    } finally {
      if (consolidateAbortRef.current === controller) consolidateAbortRef.current = null;
      setConsolidateBusy(false);
    }
  }

  function mergeExportJob(job) {
    if (!job?.id) return;
    setExportJobs((current) => [job, ...current.filter((item) => item.id !== job.id)]);
  }

  async function submitExportJob(format, saved, settings = exportSettings) {
    const job = await createAICutExportJob(projectId, format, saved.revision, settings);
    mergeExportJob(job);
    return job;
  }

  async function exportMedia(format, settings = exportSettings) {
    if (!projectId || exportingFormat) return;
    setExportingFormat(format);
    try {
      const saved = await saveTimeline({ silent: true, requireLatest: true });
      if (!saved) throw new Error('时间线尚未保存');
      await submitExportJob(format, saved, settings);
      message.success('导出任务已加入队列');
    } catch (error) {
      message.error(error?.message || '成片导出失败');
    } finally {
      setExportingFormat('');
    }
  }

  async function exportMediaBatch() {
    if (!projectId || exportingFormat) return;
    setExportingFormat('batch');
    try {
      const saved = await saveTimeline({ silent: true, requireLatest: true });
      if (!saved) throw new Error('时间线尚未保存');
      const created = [];
      const failures = [];
      for (const request of [{ format: exportSettings.format, settings: exportSettings }, { format: 'mp3', settings: {} }]) {
        try {
          created.push(await submitExportJob(request.format, saved, request.settings));
        } catch (error) {
          failures.push(error?.message || `${request.format.toUpperCase()} 提交失败`);
        }
      }
      if (created.length) message.success(`已加入 ${created.length} 个导出任务`);
      if (failures.length) message.warning(failures.join('；'));
      if (!created.length) throw new Error('批量导出没有成功加入任何任务');
    } catch (error) {
      message.error(error?.message || '批量导出失败');
    } finally {
      setExportingFormat('');
    }
  }

  async function cancelExport(job) {
    if (!projectId || !job?.id || !['queued', 'running'].includes(job.state)) return;
    setExportActionBusy(job.id);
    try {
      mergeExportJob(await cancelAICutExportJob(projectId, job.id));
      message.info('已请求取消导出任务');
    } catch (error) {
      message.error(error?.message || '取消导出失败');
    } finally {
      setExportActionBusy('');
    }
  }

  async function downloadExport(job) {
    if (!projectId || !job?.id || job.state !== 'succeeded') return;
    setExportActionBusy(job.id);
    try {
      const blob = await downloadAICutExportJob(projectId, job.id);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = `${currentProject?.title || 'AI-CUT成片'}-${job.id}.${job.format}`;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
      message.success(`${String(job.format || '').toUpperCase()} 已开始下载`);
    } catch (error) {
      message.error(error?.message || '导出文件下载失败');
    } finally {
      setExportActionBusy('');
    }
  }

  async function retryExport(job) {
    if (!job?.id) return;
    await exportMedia(job.format, job.options || {});
  }

  async function reloadExportQueue() {
    if (!projectId) return;
    try {
      const jobs = await listAICutExportJobs(projectId);
      setExportJobs(jobs);
      setExportQueueError('');
    } catch (error) {
      setExportQueueError(error?.message || '导出任务队列读取失败');
      message.error(error?.message || '导出任务队列读取失败');
    }
  }

  function openExportPanel() {
    setExportOpen(true);
    void reloadExportQueue();
  }

  async function smartSplitAsset(asset) {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!asset || asset.kind !== 'video' || mediaToolBusy) return false;
    if (!timeline.tracks.some((track) => track.clips?.some((clip) => clip.assetId === asset.id))) {
      message.warning('请先把这个视频加入时间线');
      return false;
    }
    setMediaToolBusy(`split:${asset.id}`);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      const result = await smartSplitAICutVideo(projectId, {
        assetId: asset.id,
        threshold: 0.4,
        minIntervalMs: 500,
        expectedRevision: saved.revision
      });
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(result.timeline || saved);
      setSaveState('saved');
      message.success(result.message || `已切出 ${result.splitCount || 0} 个新片段`);
      return true;
    } catch (error) {
      message.error(error?.message || '智能切镜失败');
      return false;
    } finally {
      setMediaToolBusy('');
    }
  }

  async function extractAudioAsset(asset) {
    if (!asset || asset.kind !== 'video' || mediaToolBusy) return false;
    setMediaToolBusy(`audio:${asset.id}`);
    try {
      const created = await extractAICutAudio(projectId, asset.id);
      setAssets((current) => [...current.filter((item) => item.id !== created.id), created]);
      setSelectedAssetId(created.id);
      message.success('音频已分离并加入项目素材池');
      return true;
    } catch (error) {
      message.error(error?.message || '分离音频失败');
      return false;
    } finally {
      setMediaToolBusy('');
    }
  }

  async function generateProxyAsset(asset) {
    if (!asset || asset.kind !== 'video' || mediaToolBusy) return false;
    setMediaToolBusy(`proxy:${asset.id}`);
    try {
      const updated = await generateAICutProxy(projectId, asset.id);
      setAssets((current) => current.map((item) => (item.id === updated.id ? { ...item, ...updated } : item)));
      message.success(asset.proxyUrl ? '代理媒体已重新生成' : '代理媒体已生成，网页预览将减少内存占用');
      return true;
    } catch (error) {
      message.error(error?.message || '代理媒体生成失败');
      return false;
    } finally {
      setMediaToolBusy('');
    }
  }

  async function correctCaptions() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!captionScript.trim()) return message.warning('请输入校对剧本');
    const controller = new AbortController();
    captionAbortRef.current = controller;
    setCorrectingCaptions(true);
    try {
      const result = await correctAICutCaptions(projectId, {
        script: captionScript,
        modelId: captionModelId,
        expectedRevision: timeline.revision,
        signal: controller.signal
      });
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(result.timeline);
      setCaptionOpen(false);
      message.success(result.message || `已修正 ${result.applied || 0} 条字幕`);
    } catch (error) {
      if (controller.signal.aborted) message.info('字幕纠错已取消');
      else message.error(error?.message || '字幕纠错失败');
    } finally {
      if (captionAbortRef.current === controller) captionAbortRef.current = null;
      setCorrectingCaptions(false);
    }
  }

  async function trimFillers() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!timeline.tracks.some((track) => track.kind === 'subtitle' && !track.locked && track.clips?.some((clip) => String(clip.text || '').trim()))) {
      return message.warning('时间线上没有可处理的字幕，或字幕轨已锁定');
    }
    const controller = new AbortController();
    fillerAbortRef.current = controller;
    setTrimmingFillers(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(saved);
      const result = await trimAICutFillers(projectId, {
        modelId: fillerModelId,
        expectedRevision: saved.revision,
        signal: controller.signal
      });
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(result.timeline);
      undoRef.current = [];
      redoRef.current = [];
      setHistoryVersion((value) => value + 1);
      setFillerOpen(false);
      const summary = result.summary || {};
      message.success(result.message || `已删除 ${summary.fillerCount || 0} 条水词字幕`);
    } catch (error) {
      if (controller.signal.aborted) message.info('已取消，没有删除字幕');
      else message.error(error?.message || '水词字幕删除失败');
    } finally {
      if (fillerAbortRef.current === controller) fillerAbortRef.current = null;
      setTrimmingFillers(false);
    }
  }

  function cancelFillerTrim() {
    fillerAbortRef.current?.abort();
    setFillerOpen(false);
    setOralOpen(false);
  }

  async function cleanOralTimeline() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!timeline.tracks.some((track) => track.kind === 'subtitle' && !track.locked && track.clips?.some((clip) => String(clip.text || '').trim()))) {
      return message.warning('时间线上没有可处理的字幕，或字幕轨已锁定');
    }
    const controller = new AbortController();
    fillerAbortRef.current = controller;
    setTrimmingFillers(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(saved);
      const result = await cleanAICutOralTimeline(projectId, {
        modelId: fillerModelId,
        maxPauseMs: 1000,
        expectedRevision: saved.revision,
        signal: controller.signal
      });
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(result.timeline);
      undoRef.current = [];
      redoRef.current = [];
      setHistoryVersion((value) => value + 1);
      setOralOpen(false);
      const summary = result.summary || {};
      message.success(result.message || `已删除 ${summary.fillerCount || 0} 处口误，压缩 ${summary.pauseCount || 0} 处停顿`);
    } catch (error) {
      if (controller.signal.aborted) message.info('去口误停顿已取消，时间线没有改变');
      else message.error(error?.message || '去口误停顿失败');
    } finally {
      if (fillerAbortRef.current === controller) fillerAbortRef.current = null;
      setTrimmingFillers(false);
    }
  }

  function cancelCaptionCorrection() {
    captionAbortRef.current?.abort();
    setCaptionOpen(false);
  }

  async function startASR() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    if (!selectedAsset || !['audio', 'video'].includes(selectedAsset.kind)) return message.warning('请选择音频或视频素材');
    if (!asrModelId) return message.warning('请先到 API 配置启用语音识别模型');
    if (asrJobs.some((job) => job.assetId === selectedAsset.id && ['queued', 'running', 'cancel_requested'].includes(job.state))) return message.warning('该素材正在识别，请等待完成或先取消');
    setASRBusy(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      if (activeTimelineScopeRef.current !== operationScope) return;
      setTimeline(saved);
      const job = await createAICutASRJob(projectId, {
        assetId: selectedAsset.id,
        modelId: asrModelId,
        language: asrLanguage,
        expectedRevision: saved.revision
      });
      setASRJobs((current) => [job, ...current.filter((item) => item.id !== job.id)]);
      setASROpen(false);
      message.success('语音识别任务已提交');
    } catch (error) {
      message.error(error?.message || '语音识别任务提交失败');
    } finally {
      setASRBusy(false);
    }
  }

  async function cancelASR(job) {
    try {
      const updated = await cancelAICutASRJob(projectId, job.id);
      setASRJobs((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      message.info(updated.state === 'canceled' ? '排队任务已取消' : '正在取消语音识别');
    } catch (error) {
      message.error(error?.message || '取消失败');
    }
  }

  async function openCreativeCanvas() {
    const scope = activeTimelineScopeRef.current;
    try {
      await persistBeforeTimelineChange();
      if (activeTimelineScopeRef.current === scope) setWorkspaceMode('canvas');
    } catch (error) {
      message.error(error?.message || '请先保存当前剪辑后再打开创作画布');
    }
  }

  async function refreshTimelineFromAgent() {
    const operationScope = `${projectId}:${activeTimelineId}`;
    const capturedEditSequence = editSequenceRef.current;
    const [nextTimeline, nextHistory] = await Promise.all([loadAICutTimeline(projectId), getAICutTimelineHistory(projectId)]);
    if (activeTimelineScopeRef.current !== operationScope) return;
    if (capturedEditSequence !== editSequenceRef.current || lastPersistedEditSequenceRef.current < editSequenceRef.current) { setConflictRecovery({ scope: operationScope, copyId: '' }); return; }
    if (activeTimelineScopeRef.current !== operationScope) return;
    setTimeline(nextTimeline);
        setSourcePreviewAsset(null);
        setLoadedTimelineProjectId(`${projectId}:${activeTimelineId}`);
        editSequenceRef.current = 0; lastPersistedEditSequenceRef.current = 0;
    setServerHistory(nextHistory);
    undoRef.current = [];
    redoRef.current = [];
    setHistoryVersion((value) => value + 1);
  }

  function openAutoEditor() {
    autoEditScopeRef.current = `${projectId}:${activeTimelineId}`;
    setAutoEditOpen(true);
    setAgentDrawerOpen(true);
  }
  async function handleAutoEditApplied(result) {
    const operationScope = `${projectId}:${activeTimelineId}`;
    const capturedEditSequence = editSequenceRef.current;
    if (operationScope !== activeTimelineScopeRef.current) return;
    if (!result?.timeline) return;
    let history;
    try { history = await getAICutTimelineHistory(projectId); }
    catch (error) { if (operationScope === activeTimelineScopeRef.current) message.warning(error?.message || '剪辑已应用，撤销记录读取失败，请刷新工程'); return; }
    if (operationScope !== activeTimelineScopeRef.current) return;
    if (capturedEditSequence !== editSequenceRef.current || lastPersistedEditSequenceRef.current < editSequenceRef.current) {
      setConflictRecovery({ scope: operationScope, copyId: '' });
      return;
    }
    setTimeline(result.timeline);
    setSaveState('saved');
    setServerHistory(history);
    undoRef.current = [];
    redoRef.current = [];
    setHistoryVersion((value) => value + 1);
    setAutoEditOpen(false);
    message.success(`已应用 ${Number(result.summary?.applied || 0)} 项 AI 剪辑操作，可使用撤销恢复`);
  }

  async function importScript(file) {
    if (!file) return Upload.LIST_IGNORE;
    try {
      const text = await file.text();
      if (!text.trim()) return message.warning('剧本文件是空的');
      if ([...text].length > 200000) return message.error('剧本不能超过 20 万字');
      commitTimeline((current) => ({ ...current, script: text }));
      setCaptionScript(text);
      message.success('剧本已导入，保存工程后会一起持久化');
    } catch (error) {
      message.error(error?.message || '剧本读取失败');
    }
    return Upload.LIST_IGNORE;
  }

  function exportSRT() {
    const clips = subtitleClips(timeline);
    if (!clips.length) return message.warning('字幕轨还没有内容');
    const body = clips.map((clip, index) => `${index + 1}\n${srtTime(clip.startMs)} --> ${srtTime(Number(clip.startMs) + Number(clip.durationMs))}\n${String(clip.text || '').trim()}\n`).join('\n');
    const url = URL.createObjectURL(
      new Blob([`\uFEFF${body}`], {
        type: 'application/x-subrip;charset=utf-8'
      })
    );
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `${currentProject?.title || 'AI-CUT'}-字幕.srt`;
    anchor.click();
    URL.revokeObjectURL(url);
    message.success(`已导出 ${clips.length} 条字幕`);
  }

  function replaceSubtitleText() {
    if (!replaceFrom) return message.warning('请输入要替换的文字');
    let count = 0;
    commitTimeline((current) => ({
      ...current,
      tracks: current.tracks.map((track) =>
        track.kind !== 'subtitle' || track.locked
          ? track
          : {
              ...track,
              clips: (track.clips || []).map((clip) => {
                const text = String(clip.text || '');
                if (!text.includes(replaceFrom)) return clip;
                count += text.split(replaceFrom).length - 1;
                return {
                  ...clip,
                  text: text.split(replaceFrom).join(replaceTo)
                };
              })
            }
      )
    }));
    setReplaceOpen(false);
    message.success(count ? `已替换 ${count} 处` : '没有找到匹配文字');
  }

  function addManualSubtitle() {
    const text = manualSubtitle.text.trim();
    if (!text) return message.warning('请输入字幕文字');
    if (manualSubtitle.startMs < 0 || manualSubtitle.durationMs <= 0) return message.warning('字幕时间必须大于 0');
    commitTimeline((current) => {
      const tracks = current.tracks.map((track) => ({
        ...track,
        clips: [...(track.clips || [])]
      }));
      let track = tracks.find((item) => item.kind === 'subtitle' && !item.locked);
      if (!track) {
        track = {
          id: newID('track-subtitle'),
          kind: 'subtitle',
          name: '字幕1',
          order: tracks.length,
          muted: false,
          locked: false,
          clips: []
        };
        tracks.push(track);
      }
      track.clips.push({
        id: newID('subtitle'),
        kind: 'subtitle',
        name: text.slice(0, 16),
        startMs: Number(manualSubtitle.startMs),
        durationMs: Number(manualSubtitle.durationMs),
        sourceStartMs: 0,
        sourceDurationMs: Number(manualSubtitle.durationMs),
        text
      });
      track.clips.sort((a, b) => Number(a.startMs) - Number(b.startMs));
      return { ...current, tracks };
    });
    setManualSubtitle((current) => ({
      ...current,
      text: '',
      startMs: current.startMs + current.durationMs
    }));
    message.success('字幕已加入时间线');
  }

  function addTextOverlay() {
    const text = manualSubtitle.text.trim();
    if (!text) return message.warning('请输入文字内容');
    if (manualSubtitle.startMs < 0 || manualSubtitle.durationMs <= 0) return message.warning('文字时间必须大于 0');
    commitTimeline((current) => {
      const tracks = current.tracks.map((track) => ({
        ...track,
        settings: { ...(track.settings || {}) },
        clips: [...(track.clips || [])]
      }));
      const style = textStyleCatalog.find((item) => item.id === textStyleId) || textStyleCatalog[0];
      const track = {
        id: newID('track-text'),
        kind: 'subtitle',
        name: style.label,
        order: tracks.length,
        muted: false,
        locked: false,
        settings: {
          subtitleStyleId: `text-${style.id}`,
          subtitleAnimation,
          previewVisible: true,
          ...normalizeAICutTextStyle(textStyleDraft)
        },
        clips: []
      };
      track.clips.push({
        id: newID('text'),
        kind: 'subtitle',
        name: text.slice(0, 16),
        startMs: Number(manualSubtitle.startMs),
        durationMs: Number(manualSubtitle.durationMs),
        sourceStartMs: 0,
        sourceDurationMs: Number(manualSubtitle.durationMs),
        text
      });
      tracks.push(track);
      return { ...current, tracks };
    });
    setManualSubtitle((current) => ({
      ...current,
      text: '',
      startMs: current.startMs + current.durationMs
    }));
    message.success('文字已加入时间线');
  }

  function applyMixedTextRun() {
	const input = textInputRef.current?.input;
	const start = Number(textSelection.start ?? input?.selectionStart ?? 0);
	const end = Number(textSelection.end ?? input?.selectionEnd ?? 0);
	if (end <= start) return message.warning('请先在文字输入框中选中要混排的字符');
	setTextStyleDraft((current) => normalizeAICutTextStyle({
	  ...current,
	  textRuns: [...(current.textRuns || []), { start, end, ...textRunDraft }]
	}));
	message.success(`已给第 ${start + 1}–${end} 个字符设置混排样式`);
  }

  function captureMixedTextSelection(event) {
	const value = String(event.currentTarget.value || '');
	setTextSelection({
	  start: Array.from(value.slice(0, Number(event.currentTarget.selectionStart || 0))).length,
	  end: Array.from(value.slice(0, Number(event.currentTarget.selectionEnd || 0))).length
	});
  }

  function previewCatalogResource(resource, suppliedParameters = {}) {
    if (!resource) return setPreviewResource(null);
    const parameters = normalizeCatalogParameters(resource.id, suppliedParameters);
    setPreviewResource({ ...resource, previewParameters: parameters });
    if (['filter', 'effect'].includes(resource.category) && !selectedClipEntry) message.info('先选择一个画面片段，再预览该资源');
  }

  function applyCatalogResource(resource, suppliedParameters = {}, options = {}) {
    const parameters = normalizeCatalogParameters(resource.id, suppliedParameters);
    const targetStartMs = Number.isFinite(Number(options.startMs)) ? Math.max(0, Number(options.startMs)) : previewTimeMs;
    const targetEntry = options.trackId && options.clipId
      ? (() => {
        const track = timeline.tracks.find((item) => item.id === options.trackId);
        const clip = track?.clips?.find((item) => item.id === options.clipId);
        return track && clip ? { track, clip } : null;
      })()
      : selectedClipEntry;
    setRecentFeatures((current) => recordRecentAICutFeature(current, resource.id));
    if (resource.category === 'effect') {
      const result = addEffectLayer(
        timeline,
        {
          resourceId: resource.id,
          name: resource.labelZh,
          startMs: targetStartMs,
          durationMs: 2000,
          parameters,
        },
        () => newID('effect')
      );
      if (!result.ok) return message.warning(result.reasonZh);
      commitTimeline(result.timeline);
      setPreviewResource(null);
      return message.success(`已把${resource.labelZh}加入特效轨`);
    }
    if (resource.category === 'filter') {
      if (!targetEntry || !['video', 'image'].includes(targetEntry.clip.kind)) return message.warning('请先选择视频或图片片段');
      if (targetEntry.track.locked) return message.warning('轨道已锁定');
      updateClip(targetEntry.track.id, targetEntry.clip.id, {
        visualEffect: null,
        catalogResourceId: resource.id,
        catalogParameters: parameters
      });
      setPreviewResource(null);
      return message.success(`已应用${resource.labelZh}`);
    }
    if (resource.category === 'transition') {
      if (!targetEntry || targetEntry.track.kind !== 'video') return message.warning('请先选择同一视频轨中的后一个片段');
      try {
        const result = runTimelineCommand({
          type: 'clip.transition_set',
          trackId: targetEntry.track.id,
          clipId: targetEntry.clip.id,
          transition: resource.id,
          transitionName: resource.labelZh,
          transitionMs: Number(parameters.durationMs ?? 500)
        });
        if (result.selectionKey) {
          setSelectedClipKey(result.selectionKey);
          setSelectedClipKeys([result.selectionKey]);
        }
        return message.success(`已添加${resource.labelZh}`);
      } catch (error) {
        return message.warning(error?.message || '转场添加失败');
      }
    }
    if (resource.category === 'text') {
	  const templateStyle = catalogTextSettings(resource.id,parameters);
	  if (!templateStyle) return message.warning('这个文字模板还没有可执行样式');
      commitTimeline((current) => {
        const tracks = current.tracks.map((track) => ({
          ...track,
          clips: [...(track.clips || [])]
        }));
        const track = {
          id: newID('track-text'),
          kind: 'subtitle',
          name: resource.labelZh,
          order: tracks.length,
          muted: false,
          locked: false,
          settings: { textResourceId: resource.id, previewVisible: true, subtitleAnimation, ...templateStyle },
          clips: [
            {
              id: newID('text'),
              kind: 'subtitle',
              name: resource.labelZh,
              startMs: targetStartMs,
              durationMs: 3000,
              sourceStartMs: 0,
              sourceDurationMs: 3000,
              text: '双击编辑文字',
              textResourceId: resource.id
            }
          ]
        };
        tracks.push(track);
        return { ...current, tracks };
      });
      return message.success(`${resource.labelZh}已加入时间线`);
    }
    if (resource.category === 'template') {
      if (!projectTemplateCatalog.some(template => template.id === resource.id)) return message.warning('工程模板不存在');
      setTemplateApplication({ id: resource.id, startMs: Math.round(targetStartMs) });
      return;
    }
  }

  function applyTransition(kind) {
    const [trackId, clipId] = selectedClipKey.split('::');
    return applyTransitionTo(trackId, clipId, kind);
  }

  function applyTransitionTo(trackId, clipId, kind) {
    const transitionResource = catalogById(kind);
    if (!transitionResource || transitionResource.category !== 'transition') return message.error('该转场尚未支持');
    try {
      const result = runTimelineCommand({
        type: 'clip.transition_set',
        trackId,
        clipId,
        transition: kind,
        transitionName: transitionResource.labelZh,
        transitionMs: 500
      });
      if (result.selectionKey) {
        setSelectedClipKey(result.selectionKey);
        setSelectedClipKeys([result.selectionKey]);
      }
      const transitionClip = result.timeline.tracks.find((item) => item.kind === 'transition')?.clips?.find((item) => item.transitionTrackId === trackId && item.transitionRightClipId === clipId);
      message.success(`${transitionResource.labelZh}已作为独立转场片段加入时间线，时长 ${(Number(transitionClip?.durationMs || 500) / 1000).toFixed(2)} 秒`);
    } catch (error) {
      message.warning(error?.message || '转场添加失败');
    }
  }

  async function startBGMAnalysis() {
    const audioAssets = assets.filter((asset) => asset.kind === 'audio');
    if (!audioAssets.length) return message.warning('请先上传至少一条音频素材');
    const asset = selectedAsset?.kind === 'audio' ? selectedAsset : audioAssets[0];
    await analyzeBGMAudio(asset);
  }

  async function analyzeBGMAudio(asset) {
    if (!asset || asset.kind !== 'audio') return message.warning('请选择音频素材');
    setSelectedAssetId(asset.id);
    setBGMOpen(true);
    setBGMAnalysisBusy(true);
    try {
      const analysis = await getAICutAudioAnalysis(projectId, asset.id, 240);
      setBGMAudioAnalysis({ ...analysis, assetId: asset.id });
      message.success(`波形分析完成，检测到 ${analysis?.beatsMs?.length || 0} 个节拍点`);
    } catch (error) {
      setBGMAudioAnalysis(null);
      message.error(error?.message || '波形和节拍分析失败');
    } finally {
      setBGMAnalysisBusy(false);
    }
  }

  function applyBGMPlan(asset) {
    try {
      runTimelineCommand({
        type: 'timeline.bgm_plan',
        asset,
        options: {
          ...bgmOptions,
          referenceTrackId: bgmReferenceTrackId,
          beatsMs: bgmAudioAnalysis?.assetId === asset.id ? bgmAudioAnalysis.beatsMs : []
        }
      });
      setBGMOpen(false);
      message.success('背景音乐已按参考画面写入 BGM 轨道');
    } catch (error) {
      message.warning(error?.message || '背景音乐规划失败');
    }
  }

  async function startVisualUnderstanding() {
    const visual = selectedAsset && ['video', 'image'].includes(selectedAsset.kind) ? selectedAsset : assets.find((asset) => ['video', 'image'].includes(asset.kind));
    if (!visual) return message.warning('请先上传并选择视频或图片素材');
    setSelectedAssetId(visual.id);
    setVisualBusy(true);
    try {
      const result = await extractAICutVisualFrames(projectId, visual.id, 6);
      setVisualFrames(Array.isArray(result?.frames) ? result.frames : []);
      message.success(result?.message || '代表帧提取完成');
    } catch (error) {
      message.error(error?.message || '视觉抽帧失败');
    } finally {
      setVisualBusy(false);
    }
  }

  function cancelRegionTracking() {
    const controller = trackingAbortRef.current;
    if (controller) controller.abort();
  }

  async function runRegionTracking(region) {
    if (!projectId) return message.warning('请先创建剪辑项目');
    const entry = selectedClipEntry;
    if (!entry || entry.clip.kind !== 'video') return message.warning('请先选择视频片段');
    if (entry.track.locked) return message.warning('轨道已锁定');
    const clip = entry.clip;
    if (clip.reversed) return message.warning('倒放片段暂不支持区域跟踪');
    if (clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null) return message.warning('定格片段不需要区域跟踪');
    if (Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2) return message.warning('曲线变速片段暂不支持区域跟踪');
    const asset = assets.find(item => item.id === clip.assetId);
    if (!asset || asset.kind !== 'video') return message.warning('片段对应的视频素材不存在');
    const playbackRate = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    const sourceSpanMs = Math.max(1, Math.round(Number(clip.durationMs) * playbackRate));
    if (sourceSpanMs > 300000) return message.warning('单次区域跟踪最长支持 5 分钟，请先分割片段');
    const sourceStartMs = Math.max(0, Math.round(Number(clip.sourceStartMs) || 0));
    const count = Math.min(60, Math.max(2, Math.ceil(sourceSpanMs / 500) + 1));
    const controller = new AbortController();
    trackingAbortRef.current?.abort();
    trackingAbortRef.current = controller;
    setTrackingSelectionActive(false);
    setRegionTrackingBusy(true);
    try {
      const response = await extractAICutTrackingFrames(projectId, {
        assetId: asset.id,
        startMs: sourceStartMs,
        durationMs: sourceSpanMs,
        count,
        width: 192,
        signal: controller.signal
      });
      const decoded = await decodeTrackingFrames(response?.frames, controller.signal);
      if (decoded.frames.length < 2) throw new Error('视频区间内没有足够的可跟踪画面');
      const targetWidth = Math.min(1, Math.max(.05, Number(region?.width) || .4));
      const targetHeight = Math.min(1, Math.max(.05, Number(region?.height) || .4));
      const target = {
        x: Math.min(1 - targetWidth, Math.max(0, (Number(region?.centerX) || .5) - targetWidth / 2)),
        y: Math.min(1 - targetHeight, Math.max(0, (Number(region?.centerY) || .5) - targetHeight / 2)),
        width: targetWidth,
        height: targetHeight
      };
      const tracked = trackRegionSequence(decoded.frames, decoded.width, decoded.height, target, {
        maxShift: Math.min(24, Math.max(4, Math.round(decoded.width * .08))),
        minimumConfidence: .005
      });
      if (controller.signal.aborted) throw aicutAbortError();
      if (tracked.length < 2) throw new Error('目标区域纹理不足，无法生成跟踪结果');
      const baseX = Number(clip.canvasTransform?.x) || 0;
      const baseY = Number(clip.canvasTransform?.y) || 0;
      const frames = tracked.map(frame => ({
        offsetMs: Math.min(Number(clip.durationMs), Math.round(frame.offsetMs / sourceSpanMs * Number(clip.durationMs))),
        x: baseX + frame.x,
        y: baseY + frame.y,
        confidence: frame.confidence
      }));
      const result = runTimelineCommand({
        type: 'clip.tracking_keyframes_apply',
        trackId: entry.track.id,
        clipId: clip.id,
        frames
      });
      message.success(`跟踪已写入 ${result.metadata.trackingKeyframeCount} 个位置关键帧，可撤销、保存并随成片导出`);
    } catch (error) {
      if (controller.signal.aborted || error?.name === 'AbortError') message.info('已取消区域跟踪，时间线未修改');
      else message.error(error?.message || '区域跟踪失败');
    } finally {
      if (trackingAbortRef.current === controller) trackingAbortRef.current = null;
      setRegionTrackingBusy(false);
    }
  }

  function cancelPersonSegmentation() {
    personSegmentationAbortRef.current?.abort();
  }

  async function runPersonSegmentation() {
    if (!projectId) return message.warning('请先创建剪辑项目');
    const entry = selectedClipEntry;
    if (!entry || entry.clip.kind !== 'video') return message.warning('请先选择视频片段');
    if (entry.track.locked) return message.warning('轨道已锁定');
    const clip = entry.clip;
    if (clip.reversed) return message.warning('倒放片段请先恢复正放后生成人物蒙版');
    if (clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null) return message.warning('定格片段不支持生成人物蒙版');
    if (Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2) return message.warning('曲线变速片段请先应用普通倍速后生成人物蒙版');
    if (Number(clip.durationMs) > 300000) return message.warning('单次人物抠像最长支持 5 分钟，请先分割片段');
    const asset = assets.find(item => item.id === clip.assetId);
    if (!asset || asset.kind !== 'video') return message.warning('片段对应的视频素材不存在');
    const controller = new AbortController();
    personSegmentationAbortRef.current?.abort();
    personSegmentationAbortRef.current = controller;
    setPersonSegmentationBusy(true);
    setPersonSegmentationProgress(0);
    try {
      const result = await generatePersonMatteVideo({
        sourceURL: asset.proxyUrl || asset.url,
        startMs: Math.max(0, Number(clip.sourceStartMs) || 0),
        durationMs: Math.max(1, Number(clip.durationMs) || 1),
        playbackRate: Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1,
        signal: controller.signal,
        onProgress: setPersonSegmentationProgress
      });
      if (controller.signal.aborted) throw aicutAbortError();
      const safeClipID = String(clip.id || 'clip').replace(/[^a-zA-Z0-9_-]+/g, '-');
      const file = new File([result.blob], `AI-CUT人物蒙版_${safeClipID}_${Date.now()}.webm`, { type: result.mimeType });
      const matteAsset = await uploadAICutAsset(projectId, file);
      if (controller.signal.aborted) throw aicutAbortError();
      setAssets(current => current.some(item => item.id === matteAsset.id) ? current : [...current, matteAsset]);
      const processing = normalizeVisualProcessing(clip.visualProcessing);
      runTimelineCommand({
        type: 'clip.visual_processing_set',
        trackId: entry.track.id,
        clipId: clip.id,
        processing: {
          ...processing,
          personSegmentation: {
            ...processing.personSegmentation,
            enabled: true,
            matteAssetId: matteAsset.id,
            signature: buildPersonSegmentationSignature(clip)
          }
        }
      });
      message.success('人物蒙版已生成并接入预览、保存和云端导出');
    } catch (error) {
      if (controller.signal.aborted || error?.name === 'AbortError') message.info('已取消人物蒙版生成，时间线未修改');
      else message.error(error?.message || '人物蒙版生成失败');
    } finally {
      if (personSegmentationAbortRef.current === controller) personSegmentationAbortRef.current = null;
      setPersonSegmentationBusy(false);
      setPersonSegmentationProgress(0);
    }
  }

  function renderAssetList() {
    return (
      <div className="aicut-asset-list">
        {visibleAssets.length ? (
          visibleAssets.map((asset) => (
            <article
              key={asset.id}
              role="button"
              tabIndex={0}
              draggable
              title="拖到下方同类型轨道，或点加入时间线"
              className={selectedAsset?.id === asset.id ? 'active' : ''}
              onDragStart={(event) => {
                event.dataTransfer.effectAllowed = 'copy';
                event.dataTransfer.setData('application/x-aicut-asset', asset.id);
              }}
              onClick={() => setSelectedAssetId(asset.id)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') setSelectedAssetId(asset.id);
              }}
            >
              <span>{kindLabels[asset.kind] || asset.kind}</span>
              <strong title={asset.name}>{asset.name}</strong>
              <small>{Math.max(1, Math.round(asset.byteSize / 1024))} KB</small>
              {asset.proxyUrl ? <Tag className="aicut-proxy-badge">代理已启用</Tag> : null}
              <div className="aicut-asset-actions">
                {['audio', 'video'].includes(asset.kind) ? (
                  <Button
                    size="small"
                    onClick={(event) => {
                      event.stopPropagation();
                      setSelectedAssetId(asset.id);
                      setASROpen(true);
                    }}
                  >
                    识别字幕
                  </Button>
                ) : null}
                {asset.kind === 'video' ? (
                  <>
                    <Button
                      size="small"
                      loading={mediaToolBusy === `split:${asset.id}`}
                      onClick={(event) => {
                        event.stopPropagation();
                        smartSplitAsset(asset);
                      }}
                    >
                      智能切镜
                    </Button>
                    <Button
                      size="small"
                      loading={mediaToolBusy === `audio:${asset.id}`}
                      onClick={(event) => {
                        event.stopPropagation();
                        extractAudioAsset(asset);
                      }}
                    >
                      分离音频
                    </Button>
                    <Button
                      size="small"
                      loading={mediaToolBusy === `proxy:${asset.id}`}
                      onClick={(event) => {
                        event.stopPropagation();
                        generateProxyAsset(asset);
                      }}
                    >
                      {asset.proxyUrl ? '重新生成代理' : '生成代理'}
                    </Button>
                  </>
                ) : null}
                <Button
                  size="small"
                  icon={<Plus size={13} />}
                  onClick={(event) => {
                    event.stopPropagation();
                    addAssetToTimeline(asset);
                  }}
                >
                  加入时间线
                </Button>
              </div>
            </article>
          ))
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="素材池是空的。点上传导入视频、图片或音频。" />
        )}
      </div>
    );
  }

  function renderMaterialBody() {
    if (materialTab === '画面特效')
      return (
        <div className="aicut-tool-pane aicut-visual-effect-pane">
          <p>下列效果会写入工程，并同时用于播放器预览和服务器导出。</p>
          <strong>{selectedClipEntry?.clip?.name || '请先选择画面片段'}</strong>
          <div className="aicut-choice-grid">
            {visualEffectCatalog.map((item) => (
              <button type="button" key={item.id} className={visualEffectDraft.preset === item.id ? 'active' : ''} onClick={() => setVisualEffectDraft(visualEffectPreset(item.id))}>
                {item.label}
              </button>
            ))}
          </div>
          <div className="aicut-effect-fields">
            <label>
              亮度
              <InputNumber
                min={-1}
                max={1}
                step={0.05}
                value={visualEffectDraft.brightness}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    preset: 'custom',
                    brightness: Number(value ?? 0)
                  }))
                }
              />
            </label>
            <label>
              对比度
              <InputNumber
                min={0.5}
                max={2}
                step={0.05}
                value={visualEffectDraft.contrast}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    preset: 'custom',
                    contrast: Number(value ?? 1)
                  }))
                }
              />
            </label>
            <label>
              饱和度
              <InputNumber
                min={0}
                max={3}
                step={0.05}
                value={visualEffectDraft.saturation}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    preset: 'custom',
                    saturation: Number(value ?? 1)
                  }))
                }
              />
            </label>
            <label>
              模糊
              <InputNumber
                min={0}
                max={20}
                step={0.5}
                value={visualEffectDraft.blur}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    preset: 'custom',
                    blur: Number(value ?? 0)
                  }))
                }
                addonAfter="px"
              />
            </label>
            <label>
              强度
              <InputNumber
                min={0}
                max={1}
                step={0.05}
                value={visualEffectDraft.intensity ?? 0.5}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    intensity: Number(value ?? 0.5)
                  }))
                }
              />
            </label>
            <label>
              速度
              <InputNumber
                min={0.1}
                max={5}
                step={0.1}
                value={visualEffectDraft.speed ?? 1}
                onChange={(value) =>
                  setVisualEffectDraft((current) => ({
                    ...current,
                    speed: Number(value ?? 1)
                  }))
                }
              />
            </label>
            <Checkbox
              checked={visualEffectDraft.grayscale}
              onChange={(event) =>
                setVisualEffectDraft((current) => ({
                  ...current,
                  preset: 'custom',
                  grayscale: event.target.checked
                }))
              }
            >
              黑白
            </Checkbox>
          </div>
          <div className="aicut-tool-actions">
            <Button onClick={() => setToolWorkbenchOpen('')}>取消</Button>
            <Button type="primary" disabled={!selectedClipEntry} onClick={applyVisualEffect}>
              应用到片段
            </Button>
          </div>
        </div>
      );
    if (materialTab === '转场')
      return (
        <div className="aicut-tool-pane">
          <p>选择同一视频轨中相接的第二个片段，再点击或拖入转场。两个片段之间有空隙时会拒绝添加。</p>
          <div className="aicut-transition-grid">
            {catalogForTab('transition').map((item) => (
              <button type="button" key={item.id} draggable title={item.description} onDragStart={(event) => event.dataTransfer.setData('application/x-aicut-resource', JSON.stringify(item))} onClick={() => applyTransition(item.id)}>
                <strong>{item.labelZh}</strong>
                <small>{item.description}</small>
              </button>
            ))}
          </div>
        </div>
      );
    if (materialTab === '文本')
      return (
        <div className="aicut-tool-pane">
          <p>添加独立文字轨，可在时间线中移动、复制、分割和删除。</p>
          <div className="aicut-choice-grid">
            {textStyleCatalog.map((item) => (
              <button type="button" key={item.id} className={textStyleId === item.id ? 'active' : ''} onClick={() => setTextStyleId(item.id)}>
                {item.label}
              </button>
            ))}
          </div>
          <div className="aicut-manual-subtitle">
            <Input
			  ref={textInputRef}
              value={manualSubtitle.text}
			  onSelect={captureMixedTextSelection}
			  onMouseUp={captureMixedTextSelection}
			  onKeyUp={captureMixedTextSelection}
              onChange={(event) =>
                setManualSubtitle((value) => ({
                  ...value,
                  text: event.target.value
                }))
              }
              placeholder="输入文字内容"
            />
            <InputNumber
              min={0}
              value={manualSubtitle.startMs}
              onChange={(value) =>
                setManualSubtitle((current) => ({
                  ...current,
                  startMs: Number(value || 0)
                }))
              }
              addonAfter="ms"
            />
            <InputNumber
              min={1}
              value={manualSubtitle.durationMs}
              onChange={(value) =>
                setManualSubtitle((current) => ({
                  ...current,
                  durationMs: Number(value || 1)
                }))
              }
              addonAfter="ms"
            />
            <Button type="primary" onClick={addTextOverlay}>
              添加文字
            </Button>
          </div>
          <div className="aicut-text-style-editor">
			<label>字体<Select value={textStyleDraft.fontAssetId ? `asset:${textStyleDraft.fontAssetId}` : textStyleDraft.fontFamily} options={fontOptions} onChange={(value) => {
			  if (String(value).startsWith('asset:')) {
				const asset = fontAssets.find((item) => item.id === String(value).slice(6));
				if (asset) setTextStyleDraft((current) => ({ ...current, fontFamily: asset.fontFamily, fontAssetId: asset.id }));
				return;
			  }
			  setTextStyleDraft((current) => ({ ...current, fontFamily: value, fontAssetId: '' }));
			}} /></label>
            <label>字号<InputNumber min={12} max={240} value={textStyleDraft.fontSize} onChange={(fontSize) => setTextStyleDraft((current) => ({ ...current, fontSize: Number(fontSize || 54) }))} /></label>
            <label>文字颜色<input type="color" value={textStyleDraft.textColor.slice(0, 7)} onChange={(event) => setTextStyleDraft((current) => ({ ...current, textColor: event.target.value }))} /></label>
            <label>描边颜色<input type="color" value={textStyleDraft.strokeColor.slice(0, 7)} onChange={(event) => setTextStyleDraft((current) => ({ ...current, strokeColor: event.target.value }))} /></label>
            <label>描边<InputNumber min={0} max={12} value={textStyleDraft.strokeWidth} onChange={(strokeWidth) => setTextStyleDraft((current) => ({ ...current, strokeWidth: Number(strokeWidth || 0) }))} /></label>
            <label>背景板<Checkbox checked={textStyleDraft.backgroundColor.length === 7 || textStyleDraft.backgroundColor.slice(7) !== '00'} onChange={(event) => setTextStyleDraft((current) => ({ ...current, backgroundColor: `${current.backgroundColor.slice(0, 7)}${event.target.checked ? 'cc' : '00'}` }))}>启用</Checkbox></label>
            <label>背景颜色<input type="color" value={textStyleDraft.backgroundColor.slice(0, 7)} onChange={(event) => setTextStyleDraft((current) => ({ ...current, backgroundColor: `${event.target.value}${current.backgroundColor.length === 9 ? current.backgroundColor.slice(7) : 'cc'}` }))} /></label>
            <label>阴影颜色<input type="color" value={textStyleDraft.shadowColor.slice(0, 7)} onChange={(event) => setTextStyleDraft((current) => ({ ...current, shadowColor: `${event.target.value}${current.shadowColor.length === 9 ? current.shadowColor.slice(7) : 'cc'}` }))} /></label>
            <label>阴影横向<InputNumber min={-40} max={40} value={textStyleDraft.shadowOffsetX} onChange={(shadowOffsetX) => setTextStyleDraft((current) => ({ ...current, shadowOffsetX: Number(shadowOffsetX || 0) }))} addonAfter="px" /></label>
            <label>阴影纵向<InputNumber min={-40} max={40} value={textStyleDraft.shadowOffsetY} onChange={(shadowOffsetY) => setTextStyleDraft((current) => ({ ...current, shadowOffsetY: Number(shadowOffsetY || 0) }))} addonAfter="px" /></label>
            <label>阴影模糊<InputNumber min={0} max={20} value={textStyleDraft.shadowBlur} onChange={(shadowBlur) => setTextStyleDraft((current) => ({ ...current, shadowBlur: Number(shadowBlur || 0) }))} addonAfter="px" /></label>
            <label>字距<InputNumber min={-10} max={40} value={textStyleDraft.letterSpacing} onChange={(letterSpacing) => setTextStyleDraft((current) => ({ ...current, letterSpacing: Number(letterSpacing || 0) }))} /></label>
            <label>行高<InputNumber min={0.8} max={3} step={0.1} value={textStyleDraft.lineHeight} onChange={(lineHeight) => setTextStyleDraft((current) => ({ ...current, lineHeight: Number(lineHeight || 1.3) }))} /></label>
            <label>文字弧度<InputNumber min={-180} max={180} step={5} value={textStyleDraft.textCurve} onChange={(textCurve) => setTextStyleDraft((current) => ({ ...current, textCurve: Number(textCurve || 0) }))} addonAfter="°" /></label>
            <label>对齐<Select value={textStyleDraft.textAlign} options={[{ value: 'left', label: '左对齐' }, { value: 'center', label: '居中' }, { value: 'right', label: '右对齐' }]} onChange={(textAlign) => setTextStyleDraft((current) => ({ ...current, textAlign }))} /></label>
            <label>动画（含逐字高亮）<Select value={subtitleAnimation} options={subtitleAnimationCatalog.map((item) => ({ value: item.id, label: item.label }))} onChange={setSubtitleAnimation} /></label>
			<div className="aicut-mixed-text-editor">
			  <strong>混排选区</strong>
			  <small>先在上方输入框中拖选字符，再设置该段的颜色、字号、粗体或斜体。</small>
			  <small>{textSelection.end > textSelection.start ? `当前选中第 ${textSelection.start + 1}–${textSelection.end} 个字符` : '当前没有选中文字'}</small>
			  <label>选区起点<InputNumber min={1} max={Math.max(1, Array.from(manualSubtitle.text).length)} value={textSelection.start + 1} onChange={(value) => setTextSelection((current) => ({ ...current, start: Math.max(0, Number(value || 1) - 1) }))} /></label>
			  <label>选区终点<InputNumber min={1} max={Math.max(1, Array.from(manualSubtitle.text).length)} value={Math.max(1, textSelection.end)} onChange={(value) => setTextSelection((current) => ({ ...current, end: Math.max(0, Number(value || 1)) }))} /></label>
			  <label>选区颜色<input type="color" value={textRunDraft.textColor} onChange={(event) => setTextRunDraft((current) => ({ ...current, textColor: event.target.value }))} /></label>
			  <label>选区字号<InputNumber min={12} max={240} value={textRunDraft.fontSize} onChange={(fontSize) => setTextRunDraft((current) => ({ ...current, fontSize: Number(fontSize || 54) }))} /></label>
			  <Checkbox checked={textRunDraft.bold} onChange={(event) => setTextRunDraft((current) => ({ ...current, bold: event.target.checked }))}>粗体</Checkbox>
			  <Checkbox checked={textRunDraft.italic} onChange={(event) => setTextRunDraft((current) => ({ ...current, italic: event.target.checked }))}>斜体</Checkbox>
			  <Button onClick={applyMixedTextRun}>应用到选中文字</Button>
			  <Button disabled={!textStyleDraft.textRuns?.length} onClick={() => setTextStyleDraft((current) => normalizeAICutTextStyle({ ...current, textRuns: [] }))}>清空混排</Button>
			  <div className="aicut-mixed-text-runs">
				{(textStyleDraft.textRuns || []).map((run, index) => <Tag key={`${run.start}:${run.end}:${index}`} color={run.textColor} closable onClose={() => setTextStyleDraft((current) => normalizeAICutTextStyle({ ...current, textRuns: current.textRuns.filter((_, itemIndex) => itemIndex !== index) }))}>{run.start + 1}–{run.end} · {run.fontSize}px{run.bold ? ' · 粗体' : ''}{run.italic ? ' · 斜体' : ''}</Tag>)}
			  </div>
			</div>
          </div>
        </div>
      );
    if (materialTab === '文字稿')
      return (
        <div className="aicut-tool-pane aicut-transcript-pane">
          <div className="aicut-tool-actions">
            <Upload showUploadList={false} beforeUpload={importSRT} accept=".srt">
              <Button loading={importingSRT}>导入 SRT</Button>
            </Upload>
            <Button onClick={exportSRT} disabled={!subtitleClips(timeline).length}>
              导出 SRT
            </Button>
            <Button onClick={() => setReplaceOpen(true)} disabled={!subtitleClips(timeline).length}>
              替换
            </Button>
            <Button onClick={() => setFillerOpen(true)} disabled={!subtitleClips(timeline).length}>
              去水词
            </Button>
            <Button
              onClick={() => {
                setCaptionScript(timeline.script || captionScript);
                setCaptionOpen(true);
              }}
              disabled={!subtitleClips(timeline).length}
            >
              字幕纠错
            </Button>
          </div>
          <div className="aicut-tool-actions">
            <Upload showUploadList={false} beforeUpload={importScript} accept=".txt,.md,.text">
              <Button>导入剧本</Button>
            </Upload>
            <Tag>{timeline.tracks.filter((track) => track.kind === 'video').length} 条视频轨</Tag>
            <Tag>{timeline.tracks.filter((track) => track.kind === 'audio').length} 条音频轨</Tag>
          </div>
          <Input.TextArea
            value={timeline.script || ''}
            onChange={(event) =>
              commitTimeline((current) => ({
                ...current,
                script: event.target.value
              }))
            }
            rows={8}
            placeholder="导入剧本或在这里粘贴文字稿；保存工程后持久化"
          />
          <small>逐轨转写请在“我的素材”选中视频或音频后点“识别字幕”。</small>
        </div>
      );
    if (materialTab === '音频')
      return (
        <div className="aicut-tool-pane">
          <section className="aicut-tts-panel" aria-label="文字转语音">
            <div className="aicut-tts-heading">
              <div>
                <strong>文字转语音</strong>
                <small>生成结果直接保存为当前工程的音频素材，可放到播放头位置。</small>
              </div>
              <Tag color="green">云端 TTS</Tag>
            </div>
            <Input.TextArea
              aria-label="配音文本"
              value={ttsDraft.text}
              onChange={(event) => setTTSDraft((current) => ({ ...current, text: event.target.value }))}
              rows={5}
              maxLength={5000}
              showCount
              placeholder="输入旁白或对白，也可以从当前字幕、文字稿填入"
            />
            <div className="aicut-tool-actions">
              <Button size="small" onClick={() => {
                const text = String(selectedClipEntry?.clip?.text || '').trim();
                if (!text) return message.warning('请先选择一条有文字的字幕片段');
                setTTSDraft((current) => ({ ...current, text, name: current.name || selectedClipEntry.clip.name || '字幕配音' }));
              }}>使用当前字幕</Button>
              <Button size="small" onClick={() => {
                const text = String(timeline.script || '').trim();
                if (!text) return message.warning('当前工程还没有文字稿');
                setTTSDraft((current) => ({ ...current, text }));
              }}>使用文字稿</Button>
            </div>
            <div className="aicut-tts-fields">
              <label>素材名称<Input aria-label="配音素材名称" value={ttsDraft.name} onChange={(event) => setTTSDraft((current) => ({ ...current, name: event.target.value }))} maxLength={64} placeholder="例如：开场旁白" /></label>
              <label>音色<Select aria-label="配音音色" value={ttsDraft.voice} options={ttsVoiceOptions} onChange={(voice) => setTTSDraft((current) => ({ ...current, voice }))} /></label>
              <label>情绪<Select aria-label="配音情绪" value={ttsDraft.style} options={ttsStyleOptions} onChange={(style) => setTTSDraft((current) => ({ ...current, style }))} /></label>
              <label>语速<InputNumber aria-label="配音语速" min={0.5} max={2} step={0.1} value={ttsDraft.speed} onChange={(speed) => setTTSDraft((current) => ({ ...current, speed: Number(speed ?? 1.8) }))} /></label>
              <label>音调<InputNumber aria-label="配音音调" min={-50} max={50} value={ttsDraft.pitch} onChange={(pitch) => setTTSDraft((current) => ({ ...current, pitch: Number(pitch ?? 10) }))} /></label>
            </div>
            <div className="aicut-tool-actions">
              <Button disabled={ttsBusy} onClick={() => generateSpeechAsset(false)}>仅生成素材</Button>
              <Button type="primary" loading={ttsBusy} onClick={() => generateSpeechAsset(true)}>生成并加入时间线</Button>
              {ttsBusy ? <Button danger onClick={cancelSpeechGeneration}>取消生成</Button> : null}
            </div>
          </section>
          {toolWorkbenchOpen !== '文字转语音' ? <>
          <label>
            已上传曲库
            <Select value={selectedAsset?.kind === 'audio' ? selectedAsset.id : undefined} onChange={setSelectedAssetId} placeholder="选择音频素材" options={assets.filter((asset) => asset.kind === 'audio').map((asset) => ({ value: asset.id, label: asset.name }))} />
          </label>
          <p>线上版本使用已上传的项目音频，不读取用户电脑文件夹，也不依赖剪映草稿。</p>
          <div className="aicut-check-grid">
            <Checkbox
              checked={bgmOptions.align}
              onChange={(event) =>
                setBGMOptions((value) => ({
                  ...value,
                  align: event.target.checked
                }))
              }
            >
              对齐片段
            </Checkbox>
            <Checkbox
              checked={bgmOptions.fade}
              onChange={(event) =>
                setBGMOptions((value) => ({
                  ...value,
                  fade: event.target.checked
                }))
              }
            >
              渐入渐出
            </Checkbox>
            <Checkbox
              checked={bgmOptions.continue}
              onChange={(event) =>
                setBGMOptions((value) => ({
                  ...value,
                  continue: event.target.checked
                }))
              }
            >
              续接 BGM
            </Checkbox>
            <Checkbox
              checked={bgmOptions.fill}
              onChange={(event) =>
                setBGMOptions((value) => ({
                  ...value,
                  fill: event.target.checked
                }))
              }
            >
              铺满轨道
            </Checkbox>
          </div>
          <label>
            参考轨道
            <Select value={bgmReferenceTrackId || undefined} onChange={setBGMReferenceTrackId} placeholder="默认使用第一条有画面的轨道" options={timeline.tracks.filter((track) => track.kind === 'video').map((track) => ({ value: track.id, label: track.name }))} />
          </label>
          <label>
            BGM 偏好
            <Input.TextArea value={bgmPreference} onChange={(event) => setBGMPreference(event.target.value)} rows={4} placeholder="例如：虐恋用低沉钢琴，反转处换鼓点，平淡对白留白" />
          </label>
          <Button type="primary" loading={bgmAnalysisBusy} onClick={startBGMAnalysis}>
            开始分析波形与节拍
          </Button>
          </> : null}
        </div>
      );
    if (materialTab === '视觉')
      return (
        <div className="aicut-tool-pane">
          <p>先从视频真实提取代表帧。当前文字模型不能读取图片，因此不会伪造“AI 已理解”；接入视觉模型后，这些帧会用于生成保留、复核、删除建议。</p>
          <label>
            素材
            <Select
              value={selectedAsset && ['video', 'image'].includes(selectedAsset.kind) ? selectedAsset.id : undefined}
              onChange={(value) => {
                setSelectedAssetId(value);
                setVisualFrames([]);
              }}
              placeholder="选择视觉素材"
              options={assets.filter((asset) => ['video', 'image'].includes(asset.kind)).map((asset) => ({ value: asset.id, label: asset.name }))}
            />
          </label>
          <label>
            额外要求
            <Input.TextArea value={visualRequirement} onChange={(event) => setVisualRequirement(event.target.value)} rows={4} placeholder="例如：走路空镜一律建议删除，回头和停步保留" />
          </label>
          <Button type="primary" loading={visualBusy} onClick={startVisualUnderstanding}>
            提取代表帧
          </Button>
          {visualFrames.length ? (
            <div className="aicut-visual-frames">
              {visualFrames.map((frame) => (
                <figure key={`${frame.timeMs}-${frame.url}`}>
                  <img src={frame.url} alt={`${Number(frame.timeMs) / 1000} 秒画面`} />
                  <figcaption>{(Number(frame.timeMs) / 1000).toFixed(2)}s</figcaption>
                </figure>
              ))}
            </div>
          ) : null}
        </div>
      );
    if (materialTab === '字幕')
      return (
        <div className="aicut-tool-pane">
          <div className="aicut-tool-actions">
            <Button onClick={() => addTrack('subtitle')}>新建字幕轨</Button>
            <Button onClick={() => (assets.some((asset) => ['audio', 'video'].includes(asset.kind)) ? setASROpen(true) : message.warning('请先导入视频或音频'))}>识别字幕</Button>
            <Upload showUploadList={false} beforeUpload={importSRT} accept=".srt">
              <Button loading={importingSRT}>导入字幕</Button>
            </Upload>
            <Checkbox
              checked={timeline.tracks.find((track) => track.kind === 'subtitle')?.settings?.previewVisible !== false}
              onChange={(event) =>
                persistSubtitleSettings({
                  previewVisible: event.target.checked
                })
              }
            >
              预览中显示
            </Checkbox>
          </div>
          <strong>样式外观</strong>
          <div className="aicut-choice-grid">
            {subtitleStyleCatalog.map((item) => (
              <button
                type="button"
                key={item.id}
                className={subtitleStyle === item.id ? 'active' : ''}
                onClick={() => {
                  setSubtitleStyle(item.id);
                  persistSubtitleSettings({ subtitleStyleId: item.id });
                }}
              >
                {item.label}
              </button>
            ))}
          </div>
          <strong>显示节奏</strong>
          <div className="aicut-tool-actions">
            {['按句/短语', '逐词高亮'].map((item) => (
              <Button
                key={item}
                type={subtitleRhythm === item ? 'primary' : 'default'}
                onClick={() => {
                  setSubtitleRhythm(item);
                  persistSubtitleSettings({
                    subtitleRhythm: normalizeSubtitleRhythm(item)
                  });
                }}
              >
                {item}
              </Button>
            ))}
          </div>
          <strong>字幕动效</strong>
          <Select
            value={subtitleAnimationCatalog.some((item) => item.id === subtitleAnimation) ? subtitleAnimation : 'none'}
            onChange={(value) => {
              setSubtitleAnimation(value);
              persistSubtitleSettings({ subtitleAnimation: value });
            }}
            options={subtitleAnimationCatalog.map((item) => ({
              value: item.id,
              label: item.label
            }))}
          />
          <div className="aicut-manual-subtitle">
            <Input
              value={manualSubtitle.text}
              onChange={(event) =>
                setManualSubtitle((value) => ({
                  ...value,
                  text: event.target.value
                }))
              }
              placeholder="输入字幕文字"
            />
            <InputNumber
              min={0}
              value={manualSubtitle.startMs}
              onChange={(value) =>
                setManualSubtitle((current) => ({
                  ...current,
                  startMs: Number(value || 0)
                }))
              }
              addonAfter="ms"
            />
            <InputNumber
              min={1}
              value={manualSubtitle.durationMs}
              onChange={(value) =>
                setManualSubtitle((current) => ({
                  ...current,
                  durationMs: Number(value || 1)
                }))
              }
              addonAfter="ms"
            />
            <Button type="primary" onClick={addManualSubtitle}>
              添加字幕
            </Button>
          </div>
        </div>
      );
    if (materialTab === 'SKILLS')
      return (
        <div className="aicut-tool-pane aicut-skills-pane">
          <Input value={assetSearch} onChange={(event) => setAssetSearch(event.target.value)} placeholder="搜索 skills" />
          {[
            {
              title: '短剧AI剪辑',
              detail: '读取已导入剧本和字幕，对照文字纠错，按需补转写、智能切镜。保持素材原顺序，范围不清楚时先追问。',
              prompt: '按短剧AI剪辑流程检查当前工程：读取剧本和字幕，先说明缺少的信息，再对照剧本纠错'
            },
            {
              title: '整理剪辑流程',
              detail: '把一段固定剪辑流程整理成可复制的操作说明。',
              prompt: '把我接下来描述的剪辑流程整理成可复制的操作说明'
            }
          ]
            .filter((skill) => !assetSearch.trim() || `${skill.title}${skill.detail}`.toLowerCase().includes(assetSearch.trim().toLowerCase()))
            .map((skill) => (
              <button
                type="button"
                key={skill.title}
                onClick={() => {
                  setActiveAgentSkill(skill.title);
                  setAgentDraft({ text: skill.prompt, nonce: Date.now() });
                  setAutoEditOpen(false);
                  setAgentDrawerOpen(true);
                }}
              >
                <strong>{skill.title}</strong>
                <span>{skill.detail}</span>
              </button>
            ))}
        </div>
      );
    return (
      <>
        <div className="aicut-media-tools">
          <Input value={assetSearch} onChange={(event) => setAssetSearch(event.target.value)} placeholder="搜索项目素材" />
          <Upload multiple showUploadList={false} beforeUpload={upload} accept="image/*,video/*,audio/*">
            <Button loading={uploading} icon={<UploadIcon size={15} />}>
              上传素材
            </Button>
          </Upload>
        </div>
        {renderAssetList()}
      </>
    );
  }

  function chooseToolCategory(category) {
    setToolCategory(category);
    setLibraryFilter('全部');
    const tab = {
      媒体: '我的素材',
      音频: '音频',
      文本: '文本',
      转场: '转场',
      字幕: '字幕'
    }[category];
    if (tab) setMaterialTab(tab);
  }

  function openAgentWithPrompt(prompt, skill = '') {
    setAutoEditOpen(false);
    setActiveAgentSkill(skill);
    setAgentDraft({ text: prompt, nonce: Date.now() });
    setAgentDrawerOpen(true);
  }

  function renderToolLibrary() {
    const items = filterAICutCapabilities(capabilities, {
      category: toolCategory,
      filter: libraryFilter,
      favorites: favoriteFeatures,
      recent: recentFeatures,
      search: assetSearch
    });
    return (
      <div className="aicut-reference-library">
        <div className="aicut-library-tools">
          <Input value={assetSearch} onChange={(event) => setAssetSearch(event.target.value)} placeholder="搜索功能或素材" />
          <Upload className="aicut-library-upload" multiple showUploadList={false} beforeUpload={upload} accept="image/*,video/*,audio/*">
            <Button loading={uploading} icon={<Plus size={14} />}>
              添加
            </Button>
          </Upload>
        </div>
        <nav className="aicut-library-subtabs">
          {['全部', '收藏', '最近使用'].map((item) => (
            <button type="button" key={item} className={libraryFilter === item ? 'active' : ''} onClick={() => setLibraryFilter(item)}>
              {item}
            </button>
          ))}
        </nav>
        <div className="aicut-feature-list">
          {items.length ? (
            items.map((item) => {
              const Icon = editorTools.find(([label]) => label === item.category)?.[1] || Sparkles;
              const run = async () => {
                const result = await item.action();
                if (result !== false) setRecentFeatures((current) => recordRecentAICutFeature(current, item.id));
              };
              return (
                <div className="aicut-feature-card-wrap" key={item.id}>
                  <button type="button" className="aicut-feature-card active ready" onClick={() => run().catch((error) => message.error(error?.message || `${item.title}执行失败`))}>
                    <span className="aicut-feature-icon">
                      <Icon size={17} />
                    </span>
                    <span>
                      <strong>{item.title}</strong>
                      <small>{item.detail}</small>
                    </span>
                    <em className={`source-${item.source}`}>{toolSourceLabels[item.source]}</em>
                  </button>
                  <button type="button" className={`aicut-feature-favorite${favoriteFeatures.includes(item.id) ? ' active' : ''}`} aria-label={`${favoriteFeatures.includes(item.id) ? '取消收藏' : '收藏'}${item.title}`} onClick={() => setFavoriteFeatures((current) => toggleFavoriteAICutFeature(current, item.id))}>
                    <Star size={14} fill={favoriteFeatures.includes(item.id) ? 'currentColor' : 'none'} />
                  </button>
                </div>
              );
            })
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={libraryFilter === '收藏' ? '还没有收藏功能' : libraryFilter === '最近使用' ? '还没有使用记录' : '没有匹配的可用功能'} />
          )}
        </div>
      </div>
    );
  }

  useEffect(() => {
    if (!projectId || loadedTimelineProjectId !== `${projectId}:${activeTimelineId}` || !timeline.tracks.length || openedTimelineProjectRef.current === `${projectId}:${activeTimelineId}`) return;
    const body = editorShellRef.current?.querySelector('.aicut-timeline-body');
    const videoTrack = body?.querySelector('.aicut-track-row.kind-video');
    if (!body || !videoTrack) return;
    openedTimelineProjectRef.current = `${projectId}:${activeTimelineId}`;
    body.scrollTop = Math.max(0, body.scrollTop + videoTrack.getBoundingClientRect().top - body.getBoundingClientRect().top - 50);
  }, [projectId, activeTimelineId, loadedTimelineProjectId, timeline.tracks.length, workspaceMode]);
  function renderUnifiedTimeline() {
    const canvasWidth = Math.max(900, (editorShellRef.current?.clientWidth || 0) - 130, timelineTimeToPixels(durationMs + 2000, timelineZoom) + 20);
    const orderedTracks = orderedTimelineTracks(displayTracks);
    const collapsedDisplayTracks = collapseCompoundTimelineView({ tracks: orderedTracks }).tracks;
    return (
      <div className={`aicut-timeline-body${bladeMode ? ' blade-mode' : ''}`}>
        <div className="aicut-track-labels aicut-reference-labels">
          {orderedTracks.map((track) => (
            <div className={`aicut-track-label kind-${track.kind}`} key={track.id}>
              <span className="aicut-track-kind">{{ subtitle: 'T', video: '▣', audio: '♫', effect: '✦', transition: '⇄' }[track.kind] || '•'}</span>
              <span className="aicut-track-copy">
                <strong>{track.name}</strong>
                <small>{track.kind === 'subtitle' ? `${track.locked ? '锁定' : '未锁定'} · 显示` : track.kind === 'video' ? `主轨 · ${linkedMove ? '联动' : '独立'}` : track.kind === 'audio' ? `${track.muted ? '静音' : '有声'} · 波形` : track.kind === 'transition' ? '可移动 · 可裁剪' : '叠加 · 显示'}</small>
              </span>
              <div className="aicut-track-actions">
                {!track.placeholder && (
                  <>
                    <Dropdown trigger={['click']} menu={{ items: trackMenuItems(track) }}>
                      <button type="button" title="轨道菜单" onClick={(event) => event.stopPropagation()}>
                        <MoreHorizontal size={13} />
                      </button>
                    </Dropdown>
                    <button type="button" title={track.locked ? '解锁轨道' : '锁定轨道'} onClick={() => toggleTrackFlag(track.id, 'locked')}>
                      {track.locked ? <LockKeyhole size={12} /> : <UnlockKeyhole size={12} />}
                    </button>
                    <button type="button" title={track.muted ? '取消静音' : '静音'} onClick={() => toggleTrackFlag(track.id, 'muted')}>
                      {track.muted ? <VolumeX size={12} /> : <Volume2 size={12} />}
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
        <div ref={timelineScrollRef} className="aicut-tracks-scroll" onScroll={event => { if (timelineRulerScrollRef.current) timelineRulerScrollRef.current.scrollLeft = event.currentTarget.scrollLeft; }} aria-label="时间线横向滚动区域">
          <div ref={timelineCanvasRef} className="aicut-tracks-canvas aicut-reference-tracks" title="在轨道空白处拖动框选；按 Shift/Ctrl 追加选择" style={{ width: `max(100%, ${canvasWidth}px)` }}>
            {(timeline.markers || []).map((marker) => (
              <button
                type="button"
                key={marker.id}
                className="aicut-timeline-marker"
                style={{
                  left: timelineTimeToPixels(marker.timeMs, timelineZoom),
                  '--marker-color': marker.color || '#f59e0b'
                }}
                title={`${previewClock(marker.timeMs)} ${marker.label}`}
                onClick={(event) => {
                  event.stopPropagation();
                  seekPreview(Number(marker.timeMs));
                  setSelectedMarkerId(marker.id);
                }}
              >
                <span>{marker.label}</span>
              </button>
            ))}
            <i
              className="aicut-playhead"
              style={{
                left: timelineTimeToPixels(previewTimeMs, timelineZoom)
              }}
            />
            {timelineLasso ? <i className="aicut-timeline-lasso" style={{ left: timelineLasso.left, top: timelineLasso.top, width: timelineLasso.width, height: timelineLasso.height }} /> : null}
            {collapsedDisplayTracks.map((track) => (
              <div
                className={`aicut-track-row kind-${track.kind}`}
                key={track.id}
                onPointerDown={beginTimelineLasso}
                onPointerMove={updateTimelineLasso}
                onPointerUp={finishTimelineLasso}
                onPointerCancel={cancelTimelineLasso}
                onDragOver={(event) => {
                  if (!track.locked && (event.dataTransfer.types.includes('application/x-aicut-asset') || event.dataTransfer.types.includes('application/x-aicut-resource'))) event.preventDefault();
                }}
                onDrop={(event) => {
                  if (event.dataTransfer.types.includes('application/x-aicut-asset')) dropAssetOnTrack(event, track);
                  else if (event.dataTransfer.types.includes('application/x-aicut-resource')) dropCatalogResourceOnTrack(event, track);
                }}
              >
                {track.clips.map((clip) => {
                  const key = `${track.id}::${clip.id}`;
                  const isCompoundContainer = Boolean(clip.isCompoundContainer);
                  const keys = isCompoundContainer ? (clip.memberKeys || compoundMemberKeys(clip.compoundId)) : [key];
                  const selectionActive = keys.length > 0 && keys.every((selectionKey) => selectedClipKeys.includes(selectionKey));
                  const shownStart = trimPreview?.key === key ? trimPreview.startMs : dragPreview?.key === key ? dragPreview.startMs : Number(clip.startMs) || 0;
                  const shownDuration = trimPreview?.key === key ? trimPreview.durationMs : Number(clip.durationMs) || 0;
                  const transitionLabel = catalogById(clip.transition)?.labelZh || transitionCatalog.find((item) => item.id === clip.transition)?.label || clip.transition;
                  return (
                    <article
                      key={clip.id}
                      data-lasso-keys={JSON.stringify(keys)}
                      data-lasso-locked={track.locked ? 'true' : 'false'}
                      style={{
                        left: timelineTimeToPixels(shownStart, timelineZoom),
                        width: Math.max(6, timelineTimeToPixels(shownDuration, timelineZoom))
                      }}
                      className={`aicut-clip ${track.kind}${isCompoundContainer ? ' aicut-compound-container' : ''}${track.locked ? ' locked' : ''}${clip.reviewStatus ? ` review-${clip.reviewStatus}` : ''}${selectionActive ? ' selected' : ''}${dragPreview?.key === key ? ' dragging' : ''}${trimPreview?.key === key ? ' trimming' : ''}`}
                      onPointerDown={(event) => {
                        if (isCompoundContainer) {
                          beginCompoundDrag(event, track, clip);
                          return;
                        }
                        beginClipDrag(event, track, clip);
                      }}
                      onPointerMove={updateClipDrag}
                      onPointerUp={finishClipDrag}
                      onPointerCancel={() => {
                        audioFadeDragRef.current = null;
                        clipDragRef.current = null;
                        compoundDragMovedRef.current = false;
                        trimDragRef.current = null;
                        setDragPreview(null);
                        setTrimPreview(null);
                        setAudioFadePreview(null);
                      }}
                      onClick={(event) => {
                        if (isCompoundContainer) {
                          event.stopPropagation();
                          if (compoundDragMovedRef.current) {
                            compoundDragMovedRef.current = false;
                            return;
                          }
                          if (event.shiftKey || event.ctrlKey || event.metaKey) {
                            const memberKeys = clip.memberKeys || compoundMemberKeys(clip.compoundId);
                            setSelectedClipKeys((current) => [...new Set([...current, ...memberKeys])]);
                            setSelectedClipKey(memberKeys[0] || '');
                            message.info('已选择完整复合片段，可继续多选其他片段后创建上层复合');
                            return;
                          }
                          setActiveCompoundId(clip.compoundId);
                          return;
                        }
                        setSourcePreviewAsset(null);
                        if (!bladeMode) selectClip(key, event.shiftKey || event.ctrlKey || event.metaKey);
                      }}
                    >
                      {isCompoundContainer ? (
                        <button
                          type="button"
                          className="aicut-compound-trim-handle left"
                          aria-label="复合片段左裁剪手柄"
                          title="向右拖动裁剪开头；方向键右移 100ms"
                          onPointerDown={(event) => beginCompoundTrim(event, track, clip, 'left')}
                          onKeyDown={(event) => nudgeCompoundTrim(event, clip, 'left')}
                          onClick={(event) => event.stopPropagation()}
                        />
                      ) : null}
                      <button
                        type="button"
                        className="aicut-multiselect-toggle"
                        aria-label={selectionActive ? '取消多选' : '加入多选'}
                        title={isCompoundContainer ? '选择或取消选择整个复合片段' : '加入或移出多选'}
                        onPointerDown={(event) => event.stopPropagation()}
                        onClick={(event) => {
                          event.stopPropagation();
                          toggleTimelineSelection(keys);
                        }}
                      >
                        {selectionActive ? '✓' : '+'}
                      </button>
                      <strong>{clip.name}</strong>{!isCompoundContainer && ['video', 'audio'].includes(track.kind) ? <MediaArtwork asset={assets.find(asset => asset.id === clip.assetId)} projectId={projectId} clip={clip} strip /> : null}
                      {isCompoundContainer ? <span className="aicut-compound-count">{clip.memberCount} 个成员 · 单击进入</span> : null}
                      {isCompoundContainer ? (
                        <button
                          type="button"
                          className="aicut-compound-trim-handle right"
                          aria-label="复合片段右裁剪手柄"
                          title="向左拖动裁剪结尾；方向键左移 100ms"
                          onPointerDown={(event) => beginCompoundTrim(event, track, clip, 'right')}
                          onKeyDown={(event) => nudgeCompoundTrim(event, clip, 'right')}
                          onClick={(event) => event.stopPropagation()}
                        />
                      ) : null}
                      {!isCompoundContainer && (clip.keyframes || []).map((frame) => (
                        <button
                          type="button"
                          key={frame.id}
                          className="aicut-keyframe-dot"
                          style={{
                            left: `${Math.min(100, Math.max(0, (Number(frame.offsetMs) / Math.max(1, Number(clip.durationMs))) * 100))}%`
                          }}
                          title={`关键帧 ${(Number(frame.offsetMs) / 1000).toFixed(2)} 秒`}
                          onClick={(event) => {
                            event.stopPropagation();
                            selectClip(key);
                            seekPreview(Number(clip.startMs) + Number(frame.offsetMs));
                          }}
                        >
                          ◆
                        </button>
                      ))}
                      {!isCompoundContainer && track.kind === 'audio' && clip.kind === 'audio' && selectionActive ? (() => {
                        const shownFades = audioFadePreview?.key === key ? audioFadePreview : clip;
                        const clipDuration = Math.max(1, Number(clip.durationMs) || 1);
                        return <>
                          <button type="button" className="aicut-audio-fade-handle in" aria-label="音频淡入手柄" title="拖动设置淡入；左右方向键每次调整 100ms" style={{ left: `${Math.min(100, Math.max(0, Number(shownFades.fadeInMs || 0) / clipDuration * 100))}%` }} onPointerDown={(event) => beginAudioFadeDrag(event, track, clip, 'in')} onKeyDown={(event) => nudgeAudioFade(event, track, clip, 'in')} onClick={(event) => event.stopPropagation()} />
                          <button type="button" className="aicut-audio-fade-handle out" aria-label="音频淡出手柄" title="拖动设置淡出；左右方向键每次调整 100ms" style={{ left: `${Math.min(100, Math.max(0, (1 - Number(shownFades.fadeOutMs || 0) / clipDuration) * 100))}%` }} onPointerDown={(event) => beginAudioFadeDrag(event, track, clip, 'out')} onKeyDown={(event) => nudgeAudioFade(event, track, clip, 'out')} onClick={(event) => event.stopPropagation()} />
                        </>;
                      })() : null}
                      {clip.compoundId && !isCompoundContainer ? (
                        <button
                          type="button"
                          className="aicut-compound-chip"
                          title={`${clip.compoundName || '复合片段'}：点击展开全部成员`}
                          onClick={(event) => {
                            event.stopPropagation();
                            selectClip(key);
                            setSelectedClipKeys(compoundMemberKeys(clip.compoundId));
                          }}
                        >
                          复合
                        </button>
                      ) : null}
                      {clip.transition ? (
                        <button
                          type="button"
                          className="aicut-transition-chip"
                          title={`转场：${transitionLabel}`}
                          onClick={(event) => {
                            event.stopPropagation();
                            selectClip(key);
                          }}
                        >
                          {transitionLabel}
                        </button>
                      ) : null}
                    </article>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  const timelineNetworkBusy = Boolean(importingSRT || correctingCaptions || trimmingFillers || mediaToolBusy || consolidateBusy || asrBusy || ttsBusy || voiceCloneBusy || backgroundRemovalBusyId || personSegmentationBusy || regionTrackingBusy || uploading);
  petPageRuntime.current = { loading, busy: timelineBusy || timelineNetworkBusy, saveTimeline };
  const petServiceScope = `${getCurrentUsername()}:${projectId}:${activeTimelineId}`;
  const refreshPetTimeline = useCallback(() => refreshTimelineFromAgent(), [projectId, activeTimelineId]);
  const preparePetAgentRun = useCallback(async () => {
    const scope = `${projectId}:${activeTimelineId}`;
    const runtime = petPageRuntime.current;
    if (activeTimelineScopeRef.current !== scope || runtime.loading || runtime.busy) throw new Error('当前工程正在处理，请稍后再发送剪辑要求');
    if (saveInFlightRef.current || lastPersistedEditSequenceRef.current < editSequenceRef.current) {
      const saved = await runtime.saveTimeline({ silent: true, requireLatest: true });
      if (!saved) throw new Error('当前时间线保存失败，请重试后再发送');
    }
    if (activeTimelineScopeRef.current !== scope) throw new Error('工程或时间线已切换，请重新发送');
    return { revision: timelineRef.current.revision };
  }, [projectId, activeTimelineId]);
  useEffect(() => () => {
    petPageServiceCleanup.current?.();
    petPageServiceCleanup.current = null;
  }, [petServiceScope]);
  useEffect(() => {
    if (!projectId) return;
    petPageServiceCleanup.current = registerPetPageService({
      id: 'aicut-agent', scope: petServiceScope,
      label: `AI-CUT 剪辑 · ${currentProject?.title || '当前工程'}`,
      render: () => <AICutAgentPanel
        key={petServiceScope} embedded projectId={projectId} timelineRevision={timeline.revision}
        models={agentModels} assets={assets} activeSkill={activeAgentSkill} draft={agentDraft}
        uiContext={{ selectedClipKeys, playheadMs: Math.round(previewTimeMs) }}
        autoEditOpen={autoEditOpen}
        onBeforeRun={preparePetAgentRun}
        onSectionChange={value => { autoEditScopeRef.current = `${projectId}:${activeTimelineId}`; setAutoEditOpen(value); }}
        onAutoApplied={handleAutoEditApplied} onTimelineChanged={refreshPetTimeline}
      />
    });
  }, [petServiceScope, projectId, activeTimelineId, timeline.revision, currentProject?.title, agentModels, assets, activeAgentSkill, agentDraft, autoEditOpen, selectedClipKeys, previewTimeMs]);
  useEffect(() => {
    if (agentDrawerOpen && projectId) {
      openPetPageService('aicut-agent');
      setAgentDrawerOpen(false);
    }
  }, [agentDrawerOpen, projectId, petServiceScope]);
  const selectedTimelineEntries = timeline.tracks.flatMap(track => (track.clips || []).filter(clip => selectedClipKeys.includes(`${track.id}::${clip.id}`)).map(clip => ({ track, clip })));
  const playheadInsideSelection = Boolean(selectedClipEntry && previewTimeMs > selectedClipEntry.clip.startMs && previewTimeMs < selectedClipEntry.clip.startMs + selectedClipEntry.clip.durationMs);
  const toolbarState = {
    hasSelection: Boolean(selectedClipEntry), selectionCount: selectedTimelineEntries.length,
    selectionLocked: selectedTimelineEntries.some(entry => entry.track.locked),
    busy: loading || saving || timelineBusy || timelineNetworkBusy,
    canUndo: Boolean(undoRef.current.length || serverHistory.canUndo),
    canRedo: Boolean(redoRef.current.length || serverHistory.canRedo),
    canSplit: timeline.tracks.some(track => !track.locked && (track.clips || []).some(clip => previewTimeMs > clip.startMs && previewTimeMs < clip.startMs + clip.durationMs)),
    canTrimLeft: playheadInsideSelection, canTrimRight: playheadInsideSelection,
    canAddKeyframe: Boolean(selectedClipEntry && previewTimeMs >= selectedClipEntry.clip.startMs && previewTimeMs <= selectedClipEntry.clip.startMs + selectedClipEntry.clip.durationMs),
    hasLink: Boolean(selectedClipEntry?.clip?.linkGroupId), hasCompound: Boolean(selectedClipEntry?.clip?.compoundId), consolidateBusy
  };
  const toolbarPrimaryActions = buildTimelinePrimaryActions(toolbarState, {
    undo: undoTimeline, redo: redoTimeline, split: splitAtPlayhead, delete: deleteSelectedClips,
    trimLeft: () => trimSelectedAtPlayhead('left'), trimRight: () => trimSelectedAtPlayhead('right'),
    addKeyframe: addKeyframeAtPlayhead, addMarker: addMarkerAtPlayhead
  });
  function trimSelectedAtPlayhead(edge) {
    try {
      if (selectedClipEntry) runTimelineCommand({ type: 'clip.trim_to_playhead', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, timeMs: previewTimeMs, edge });
    } catch (error) { message.warning(error.message); }
  }
  return (
    <ConfigProvider
      theme={{
        token: {
          colorBgBase: 'var(--surface-base)',
          colorBgContainer: 'var(--surface-panel)',
          colorBgElevated: 'var(--surface-panel)',
          colorText: 'var(--text-primary)',
          colorTextSecondary: 'var(--text-secondary)',
          colorBorder: 'var(--border-default)',
          colorPrimary: 'var(--accent-primary)'
        }
      }}
    >
	  <main className="aicut-page aicut-fullscreen-editor">
		{projectFontCSS ? <style>{projectFontCSS}</style> : null}
        {!projectId ? (
          <div className="aicut-welcome">
            <Clapperboard size={42} />
            <h2>创建第一个 AI-CUT 项目</h2>
            <p>创建后进入在线专业剪辑器，素材、预览、属性和时间线都保存到项目。</p>
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              新建项目
            </Button>
          </div>
        ) : (
          <Spin spinning={loading}>
            <header className="aicut-editor-topbar aicut-hub-topbar aicut-reference-top">
              <button type="button" className="aicut-reference-plain" disabled={loading || saving || timelineBusy || timelineNetworkBusy} onClick={async () => { try { await persistBeforeTimelineChange(); window.location.assign('/ai-cut'); } catch (error) { message.error(error.message || '保存失败，请保存后再返回作品库'); } }}>
                ← 返回作品库
              </button>
              <div className="aicut-reference-brand" aria-label="AI-CUT Hub · 在线版">
                <b>AI</b>-CUT Hub · 在线版
              </div>
              <Dropdown trigger={['click']} menu={{ items: [
                { key: 'new', label: '新建工程', onClick: () => setCreateOpen(true) },
                { key: 'multicam', label: '多机位剪辑', onClick: openMultiCam },
                { key: 'settings', label: '工程设置', onClick: () => setCanvasSettingsOpen(true) },
                { key: 'shortcuts', label: '键盘快捷键', onClick: () => setShortcutOpen(true) },
              ] }}><Button size="small" className="aicut-project-menu">菜单</Button></Dropdown>
              <select className="aicut-reference-project" aria-label="切换剪辑项目" value={projectId} disabled={loading || saving || timelineBusy || timelineNetworkBusy} onChange={async event => { const id = event.target.value; try { await persistBeforeTimelineChange(); setActiveTimelineId('default'); setProjectId(id); window.history.replaceState(null, '', `/ai-cut?project=${encodeURIComponent(id)}`); } catch (error) { message.error(error.message); } }}>
                {projects.map((project) => (
                  <option key={project.id} value={project.id}>
                    工程 · {project.title}
                  </option>
                ))}
              </select>
              <div className="aicut-reference-modes">
                <button type="button" className={workspaceMode === 'canvas' ? 'on' : ''} onClick={() => void openCreativeCanvas()}>
                  AI 创作画布
                </button>
                <button type="button" className={workspaceMode === 'editor' ? 'on' : ''} onClick={() => setWorkspaceMode('editor')}>
                  专业剪辑
                </button>
              </div>
              <div className="aicut-reference-actions">
                <span className={`save-${saveState}`}>
                  ●{' '}
                  {
                    {
                      saved: '云端已保存',
                      dirty: '有未保存修改',
                      saving: '保存中',
                      failed: '保存失败'
                    }[saveState]
                  }
                </span>
                <button type="button" aria-label="撤销" title="撤销 Ctrl+Z" onClick={undoTimeline} disabled={!undoRef.current.length && !serverHistory.canUndo}>
                  <Undo2 size={15} />
                </button>
                <button type="button" aria-label="重做" title="重做 Ctrl+Shift+Z" onClick={redoTimeline} disabled={!redoRef.current.length && !serverHistory.canRedo}>
                  <Redo2 size={15} />
                </button>
                <button type="button" aria-label="快捷键" title="键盘快捷键" onClick={() => setShortcutOpen(true)}>
                  <Keyboard size={16} />
                </button>
                <button type="button" title="保存 Ctrl+S" onClick={saveTimeline}>
                  <Save size={16} />
                  {saving ? '保存中' : '保存'}
                </button>
                <button type="button" className="aicut-reference-export" title="导出 Ctrl+E" onClick={openExportPanel}>
                  <CloudUpload size={15} />
                  云端导出
                </button>
              </div>
            </header>

            {workspaceMode === 'canvas' ? (
              <Suspense fallback={<div className="aicut-canvas-loading"><Spin tip="正在打开创作画布…" /></div>}>
                <AICutCreativeCanvas
                  key={`${projectId}:${activeTimelineId}`}
                  projectId={projectId}
                  projectTitle={currentProject?.title || '当前工程'}
                  timelineRevision={timeline.revision}
                  onOpenEditor={() => setWorkspaceMode('editor')}
                  onTransferred={async (result) => {
                    const operationScope = `${projectId}:${activeTimelineId}`;
                    const requestedProjectId = projectId;
                    const refreshed = await listAICutAssets(requestedProjectId);
                    if (activeTimelineScopeRef.current !== operationScope) return;
                    setAssets(refreshed);
                    if (result?.timeline) await refreshTimelineFromAgent();
                    if (activeTimelineScopeRef.current !== operationScope) return;
                    setWorkspaceMode('editor');
                  }}
                />
              </Suspense>
            ) : (
              <EditorShell ref={editorShellRef} layout={editorLayout} inspectorDrawerOpen={inspectorDrawerOpen} hasSelection={Boolean(selectedClipEntry)}>
                {editorLayout.inspectorMode === 'drawer' || inspectorCollapsed ? (
                  <button type="button" className="aicut-open-inspector" aria-label={(editorLayout.inspectorMode === 'drawer') && inspectorDrawerOpen ? '关闭属性栏' : '打开属性栏'} onClick={() => (editorLayout.inspectorMode === 'drawer' ? setInspectorDrawerOpen((value) => !value) : setInspectorCollapsed(false))}>
                    {(editorLayout.inspectorMode === 'drawer') && inspectorDrawerOpen ? '关闭属性' : '属性'}
                  </button>
                ) : null}
                <ResourceLibrary
                  projectId={projectId}
                  usedAssetIds={timeline.tracks.flatMap(track => (track.clips || []).map(clip => clip.assetId))}
                  workspaces={{
                    sticker: <StickerWorkspace assets={assets} projectId={projectId} uploading={uploading} onUpload={upload} onPreview={asset => { setSelectedAssetId(asset.id); setPreviewPlaying(false); setSourcePreviewAsset(asset); }} onAdd={asset => {
                      try { const result = runTimelineCommand({ type: 'timeline.sticker_add', asset, startMs: Math.round(previewTimeMs), durationMs: 5000 }); setSelectedClipKey(result.selectionKey); setSelectedClipKeys([result.selectionKey]); setSourcePreviewAsset(null); message.success('贴纸已加入独立轨道'); } catch (error) { message.warning(error.message); }
                    }} />,
                    adjustment: <AdjustmentWorkspace selection={selectedClipEntry} onChange={processing => { try { if (selectedClipEntry) runTimelineCommand({ type: 'clip.visual_processing_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, processing }); } catch (error) { message.warning(error.message); } }} onPickVisual={() => {
                      const track = timeline.tracks.find(track => !track.locked && track.clips?.some(clip => ['video', 'image'].includes(clip.kind))); const clip = track?.clips.find(clip => ['video', 'image'].includes(clip.kind)); if (clip) { setSelectedClipKey(`${track.id}::${clip.id}`); setSelectedClipKeys([`${track.id}::${clip.id}`]); } else message.info('先把画面素材加入时间线');
                    }} />,
                    packaging: <PackagingWorkspace assets={assets} onAutoEdit={openAutoEditor} onTemplate={template => setTemplateApplication({ id: template.id, startMs: Math.round(previewTimeMs) })} onCaptions={() => setASROpen(true)} onSpeech={() => { setMaterialTab('音频'); setToolWorkbenchOpen('文字转语音'); }} />,
                    plugin: <EditorPluginWorkspace tools={capabilities} />,
                  }}
                  audioRecorder={<AudioRecorder onUpload={uploadRecording} />}
                  personalTemplateWorkspace={<PersonalTemplateWorkspace timeline={timeline} assets={assets} disabled={loading || saving || timelineBusy || timelineNetworkBusy} onApply={command => {
                    try {
                      const result = runTimelineCommand({ ...command.payload, type: command.type });
                      if (result?.selectionKey) { setSelectedClipKey(result.selectionKey); setSelectedClipKeys([result.selectionKey]); }
                      return result;
                    } catch (error) { message.error(error.message || '模板应用失败'); throw error; }
                  }} />}
                  activeTab={toolCategory}
                  assets={assets}
                  assetFolders={assetFolders}
                  selectedFolderId={selectedAssetFolderId}
                  onSelectFolder={setSelectedAssetFolderId}
                  onCreateFolder={name => manageAssetFolder(() => createAssetFolder(projectId, name))}
                  onRenameFolder={(id, name) => manageAssetFolder(() => renameAssetFolder(projectId, id, name))}
                  onDeleteFolder={id => manageAssetFolder(() => deleteAssetFolder(projectId, id))}
                  onMoveAssetFolder={(id, folderId) => manageAssetFolder(() => moveAssetToFolder(projectId, id, folderId))}
                  favorites={favoriteFeatures}
                  recent={recentFeatures}
                  uploading={uploading}
                  proxyBusyId={mediaToolBusy.startsWith('proxy:') ? mediaToolBusy.slice(6) : ''}
                  backgroundBusyId={backgroundRemovalBusyId}
                  collapsed={mediaCollapsed}
                  onToggleCollapsed={() => setMediaCollapsed((value) => !value)}
                  onTabChange={setToolCategory}
                  onUpload={upload}
				  onUseFont={(asset) => {
					setTextStyleDraft((current) => ({ ...current, fontFamily: asset.fontFamily, fontAssetId: asset.id }));
					setMaterialTab('文本');
					setToolWorkbenchOpen('文本');
				  }}
                  onPreviewAsset={(asset) => { setSelectedAssetId(asset.id); setPreviewPlaying(false); setSourcePreviewAsset(asset); }}
                  onAddAsset={addAssetToTimeline}
                  onOpenASR={(asset) => {
                    setSelectedAssetId(asset.id);
                    setASROpen(true);
                  }}
                  onAnalyzeAudio={analyzeBGMAudio}
                  onOpenTTS={() => {
                    setMaterialTab('音频');
                    setToolWorkbenchOpen('文字转语音');
                  }}
                  onOpenVoiceClone={openVoiceClone}
                  onGenerateProxy={generateProxyAsset}
                  onRemoveBackground={openBackgroundRemoval}
                  onOpenTextEditor={() => {
                    setMaterialTab('文本');
                    setToolWorkbenchOpen('文本');
                  }}
                  onPreviewResource={previewCatalogResource}
                  onApplyResource={applyCatalogResource}
                  onToggleFavorite={(id) => setFavoriteFeatures((current) => toggleFavoriteAICutFeature(current, id))}
                />
                {toolCategory === '字幕' ? (
                  <CaptionWorkspace
                    assets={assets}
                    models={asrModels}
                    jobs={asrJobs}
                    assetId={selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? selectedAsset.id : ''}
                    modelId={asrModelId}
                    language={asrLanguage}
                    busy={asrBusy}
                    onAssetChange={setSelectedAssetId}
                    onModelChange={setASRModelId}
                    onLanguageChange={setASRLanguage}
                    onStart={startASR}
                    onCancel={cancelASR}
                    onRetry={(job) => {
                      setSelectedAssetId(job.assetId);
                      setASROpen(true);
                    }}
                    onImportSRT={importSRT}
                    onExportSRT={exportSRT}
                    onAddManualCaption={() => {
                      setManualSubtitle((current) => ({
                        ...current,
                        text: '新字幕',
                        startMs: previewTimeMs
                      }));
                      setToolWorkbenchOpen('字幕');
                    }}
                  />
                ) : null}
                <InspectorPanel
                  projectSummary={<ProjectOverview project={currentProject} assetCount={assets.length} trackCount={timeline.tracks.length} durationLabel={previewClock(durationMs)} />}
                  selection={selectedClipEntry}
                  selectionCount={selectedClipKeys.length}
                  drawerMode={editorLayout.inspectorMode === 'drawer'}
                  onClose={() => setInspectorDrawerOpen(false)}
                  onCollapse={() => setInspectorCollapsed(true)}
                  onPatchClip={(patch) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, patch)}
                  onSetAnimation={animation => selectedClipEntry && runTimelineCommand({ type: 'clip.animation_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, animation })}
                  onPatchTrack={(patch) => selectedClipEntry && commitTimeline((current) => updateTimelineTrackSettings(current, selectedClipEntry.track.id, patch))}
                  onPatchEffect={(patch) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, patch)}
                  onPatchTransition={(patch) =>
                    selectedClipEntry &&
                    (() => {
                      try {
                        if (selectedClipEntry.track.kind === 'transition' || selectedClipEntry.clip.kind === 'transition') {
                          runTimelineCommand({
                            type: 'clip.resize',
                            trackId: selectedClipEntry.track.id,
                            clipId: selectedClipEntry.clip.id,
                            edge: 'right',
                            edgeTimeMs: Number(selectedClipEntry.clip.startMs) + Number(patch.transitionMs),
                            ripple: false
                          });
                        } else {
                          runTimelineCommand({
                            type: 'clip.transition_set',
                            trackId: selectedClipEntry.track.id,
                            clipId: selectedClipEntry.clip.id,
                            transition: selectedClipEntry.clip.transition,
                            transitionMs: patch.transitionMs
                          });
                        }
                      } catch (error) {
                        message.warning(error?.message || '转场时长调整失败');
                      }
                    })()
                  }
                  onSetPlaybackRate={(playbackRate) => selectedClipEntry && changeClipSpeed(selectedClipEntry.track.id, selectedClipEntry.clip.id, playbackRate)}
                  onSetSpeedCurve={(points) =>
                    selectedClipEntry &&
                    (() => {
                      try {
                        runTimelineCommand({
                          type: 'clip.speed_curve_set',
                          trackId: selectedClipEntry.track.id,
                          clipId: selectedClipEntry.clip.id,
                          points
                        });
                      } catch (error) {
                        message.warning(error?.message || '曲线变速失败');
                      }
                    })()
                  }
                  onSetAudioProcessing={(processing) =>
                    selectedClipEntry &&
                    (() => {
                      try {
                        runTimelineCommand({
                          type: 'clip.audio_processing_set',
                          trackId: selectedClipEntry.track.id,
                          clipId: selectedClipEntry.clip.id,
                          processing
                        });
                      } catch (error) {
                        message.warning(error?.message || '音频处理设置失败');
                      }
                    })()
                  }
                  onSetAudioFades={({ fadeInMs, fadeOutMs }) =>
                    selectedClipEntry && setAudioFades(selectedClipEntry.track.id, selectedClipEntry.clip.id, fadeInMs, fadeOutMs)
                  }
                  onSetVisualProcessing={(processing) =>
                    selectedClipEntry &&
                    (() => {
                      try {
                        runTimelineCommand({
                          type: 'clip.visual_processing_set',
                          trackId: selectedClipEntry.track.id,
                          clipId: selectedClipEntry.clip.id,
                          processing
                        });
                      } catch (error) {
                        message.warning(error?.message || '高级画面设置失败');
                      }
                    })()
                  }
                  onBatchApplyVisualProcessing={(patch) => {
                    try {
                      runTimelineCommand({ type: 'clip.visual_processing_batch_patch', clipKeys: selectedClipKeys, patch });
                      message.success(`已批量应用到 ${selectedClipKeys.length} 个片段`);
                    } catch (error) {
                      message.warning(error?.message || '批量画面设置失败');
                    }
                  }}
                  onDetectChroma={detectSelectedChroma}
                  chromaPickerActive={chromaPickerActive}
                  onStartChromaPicker={() => setChromaPickerActive(value => !value)}
                  regionTrackingBusy={regionTrackingBusy}
                  trackingRegion={trackingRegion}
                  trackingSelectionActive={trackingSelectionActive}
                  onTrackingRegionChange={(next) => setTrackingRegion(normalizeTrackingRegion(next))}
                  onToggleTrackingSelection={() => setTrackingSelectionActive((value) => !value)}
                  onRunRegionTracking={runRegionTracking}
                  onCancelRegionTracking={cancelRegionTracking}
                  objectRemovalRegion={objectRemovalRegion}
                  objectRemovalSelectionActive={objectRemovalSelectionActive}
                  onObjectRemovalRegionChange={(next) => setObjectRemovalRegion(normalizeTrackingRegion(next))}
                  onToggleObjectRemovalSelection={() => setObjectRemovalSelectionActive((value) => {
                    const next = !value;
                    if (next) {
                      setTrackingSelectionActive(false);
                      setChromaPickerActive(false);
                    }
                    return next;
                  })}
                  onCommitObjectRemovalRegion={commitObjectRemovalRegion}
                  onSetObjectRemovalEnabled={(enabled) => commitObjectRemovalRegion(objectRemovalRegion, enabled)}
                  personSegmentationBusy={personSegmentationBusy}
                  personSegmentationProgress={personSegmentationProgress}
                  onRunPersonSegmentation={runPersonSegmentation}
                  onCancelPersonSegmentation={cancelPersonSegmentation}
				  onSetReverse={(reversed) =>
					selectedClipEntry &&
					(() => {
					  try {
						runTimelineCommand({ type: 'clip.reverse_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, reversed });
					  } catch (error) {
						message.warning(error?.message || '倒放设置失败');
					  }
					})()
				  }
                  onSetFreezeFrame={(clear) =>
					selectedClipEntry &&
					(() => {
					  try {
						runTimelineCommand({ type: 'clip.freeze_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, timeMs: previewTimeMs, clear });
					  } catch (error) {
						message.warning(error?.message || '定格设置失败');
					  }
					})()
				  }
				  onSetBlendMode={(blendMode) =>
					selectedClipEntry &&
					(() => {
					  try {
						runTimelineCommand({ type: 'clip.blend_mode_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, blendMode });
					  } catch (error) {
						message.warning(error?.message || '混合模式设置失败');
					  }
					})()
				  }
				  onSetKeyframeEasing={(keyframeId, easing) =>
					selectedClipEntry &&
					(() => {
					  try {
						runTimelineCommand({ type: 'clip.keyframe_easing_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, keyframeId, easing });
					  } catch (error) {
						message.warning(error?.message || '关键帧缓动设置失败');
					  }
					})()
				  }
				  onBatchEditKeyframes={(keyframeIds, action, payload = {}) =>
					selectedClipEntry &&
					(() => {
					  try {
						runTimelineCommand({ type: 'clip.keyframes_batch_edit', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, keyframeIds, action, ...payload });
						message.success(action === 'remove' ? '已批量删除关键帧' : action === 'shift' ? '已批量偏移关键帧' : '已统一关键帧缓动');
					  } catch (error) {
						message.warning(error?.message || '批量编辑关键帧失败');
					  }
					})()
				  }
				  onCopyKeyframes={copySelectedKeyframes}
				  onPasteKeyframes={pasteKeyframesAtPlayhead}
				  keyframeClipboardCount={keyframeClipboardCount}
				  onVolumePointUpsert={(offsetMs, volume) =>
                    runVolumeEnvelopeCommand('clip.volume_point_upsert', {
                      offsetMs,
                      volume
                    })
                  }
                  onVolumePointMove={(keyframeId, offsetMs, volume) =>
                    runVolumeEnvelopeCommand('clip.volume_point_move', {
                      keyframeId,
                      offsetMs,
                      volume
                    })
                  }
                  onVolumePointRemove={(keyframeId) =>
                    runVolumeEnvelopeCommand('clip.volume_point_remove', {
                      keyframeId
                    })
                  }
                />
                <nav className="aicut-tool-rail aicut-reference-rail" aria-label="剪辑工具">
                  {visibleEditorTools.map(([label, Icon]) => (
                    <button type="button" key={label} className={toolCategory === label ? 'active' : ''} onClick={() => chooseToolCategory(label)}>
                      <Icon size={18} />
                      <span>{label}</span>
                    </button>
                  ))}
                </nav>
                <section className={`aicut-library-panel aicut-reference-library-panel${mediaCollapsed ? ' collapsed' : ''}`}>
                  {mediaCollapsed ? (
                    <button type="button" className="aicut-collapsed-rail" aria-label="展开素材区" onClick={() => setMediaCollapsed(false)}>
                      <PanelLeftOpen size={17} />
                      <span>工具</span>
                    </button>
                  ) : (
                    <>
                      <header>
                        <div>
                          <strong>{toolCategory}</strong>
                          <small>{capabilities.filter((item) => item.category === toolCategory).length} 项</small>
                        </div>
                        <Button type="text" size="small" aria-label="收起素材区" icon={<PanelLeftClose size={15} />} onClick={() => setMediaCollapsed(true)} />
                      </header>
                      {renderToolLibrary()}
                    </>
                  )}
                </section>
                {!mediaCollapsed ? <div role="separator" aria-label="调整素材区宽度" className="aicut-resize-handle library-preview" onPointerDown={(event) => beginLayoutResize('media', event)} onPointerMove={updateLayoutResize} onPointerUp={finishLayoutResize} onPointerCancel={finishLayoutResize} /> : null}
                {!timelineCollapsed ? <div role="separator" aria-label="上下拖动调整时间线高度" className="aicut-resize-handle upper-timeline" onPointerDown={(event) => beginLayoutResize('timeline', event)} onPointerMove={updateLayoutResize} onPointerUp={finishLayoutResize} onPointerCancel={finishLayoutResize} /> : null}

                {sourcePreviewAsset ? <SourcePreviewPanel asset={sourcePreviewAsset} onClose={() => setSourcePreviewAsset(null)} /> : <section ref={previewPanelRef} className={`aicut-preview aicut-reference-stage${previewExpanded ? ' is-expanded' : ''}`}>
                  <div className="aicut-panel-title">
                    <strong>播放器</strong>
                    <div>
                      <small>{previewFrame.visuals.length} 个画面层 · {previewFrame.audios.length} 个音频层</small>
                      <Button type="text" size="small" icon={<Palette size={14} />} onClick={openCanvasSettings}>画布背景</Button>
                    </div>
                  </div>
                  <PreviewViewport ratio={currentProject?.aspectRatio || '9:16'} zoom={canvasFit}>
                  <div
                    ref={previewStageRef}
                    onPointerDown={pickChromaFromPreview}
                    className={`aicut-preview-stage fit-fit ratio-${String(currentProject?.aspectRatio || '9:16').replace(':', '-')}${safeGuide ? ' show-safe-guide' : ''}${currentProject?.backgroundMode === 'transparent' ? ' is-transparent' : ''}${chromaPickerActive ? ' chroma-picker-active' : ''}`}
                    style={currentProject?.backgroundMode === 'transparent' ? undefined : { backgroundColor: currentProject?.backgroundColor || '#000000' }}
                  >
                    {!previewFrame.visuals.length && !previewFrame.subtitles.length ? (
                      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="把视频或图片加入时间线后，从播放头预览成片" />
                    ) : (
                      previewFrame.visuals.map((item, index) => (
                        <div className="aicut-preview-visual" data-clip-key={`${item.trackId}::${item.clipId}`} style={{ zIndex: index + 1, mixBlendMode: item.blendMode === 'add' ? 'plus-lighter' : item.blendMode }} key={`${item.trackId}:${item.clipId}`}>
                          <div className="aicut-animation-layer" style={clipAnimationPreviewStyle(item.animation, item.durationMs, item.outputOffsetMs, { position: 'absolute', inset: 0, display: 'grid', placeItems: 'center', overflow: 'hidden' })}><PreviewMedia key={previewQuality} item={item} playing={previewPlaying} quality={previewQuality} /></div>
                        </div>
                      ))
                    )}
                    {previewFrame.subtitles.length ? (
                      <div className="aicut-preview-subtitles">
                        {previewFrame.subtitles.map((item) => (
                          <div className="aicut-caption-opacity" style={{opacity:item.textStyle.textOpacity ?? 1,width:'100%',display:'flex',flexDirection:'column',alignItems:'center'}} key={`${item.trackId}:${item.clipId}`}>
                            <div className={`aicut-preview-caption style-${item.styleId} animation-${item.animation || 'none'}`} style={aicutTextPreviewStyle({...item.textStyle,textOpacity:1}, 'var(--aicut-preview-px, 1px)')}><AICutCaptionText item={item} /></div>
                          </div>
                        ))}
                      </div>
                    ) : null}
                    <div className="aicut-preview-audio">
                      {previewFrame.audios.map((item) => (
                        <PreviewMedia item={item} playing={previewPlaying} key={`${item.trackId}:${item.clipId}`} />
                      ))}
                    </div>
                    <PreviewCanvas enabled={Boolean(selectedClipEntry && ['video', 'image'].includes(selectedClipEntry.clip.kind)) && !chromaPickerActive && !trackingSelectionActive && !objectRemovalSelectionActive} transform={selectedClipEntry?.clip?.canvasTransform} onCommitTransform={(canvasTransform) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { canvasTransform })} />
                    <TrackingRegionOverlay
                      active={trackingSelectionActive && selectedClipEntry?.clip?.kind === 'video'}
                      region={trackingRegion}
                      onChange={(next) => setTrackingRegion(normalizeTrackingRegion(next))}
                      onCommit={(next) => setTrackingRegion(normalizeTrackingRegion(next))}
                      onCancel={() => setTrackingSelectionActive(false)}
                    />
                    <TrackingRegionOverlay
                      active={objectRemovalSelectionActive && ['video', 'image'].includes(selectedClipEntry?.clip?.kind)}
                      region={objectRemovalRegion}
                      label="消除区域"
                      ariaLabel="物体消除框"
                      help="框住固定水印或静态物体 · 拖动及四角缩放 · Esc 退出"
                      variant="removal"
                      onChange={(next) => setObjectRemovalRegion(normalizeTrackingRegion(next))}
                      onCommit={(next) => commitObjectRemovalRegion(next, true)}
                      onCancel={() => setObjectRemovalSelectionActive(false)}
                    />
                  </div>
                  </PreviewViewport>
                  <div className="aicut-preview-controls aicut-reference-player">
                    <div className="aicut-player-time"><span>{previewClock(previewTimeMs)}</span><span>{previewClock(durationMs)}</span></div>
                    <div className="aicut-player-transport">
                      <button type="button" aria-label="预览上一剪辑点" title="上一剪辑点" onClick={() => jumpEditPoint('previous')}><SkipBack size={14} /></button>
                      <button type="button" aria-label={previewPlaying ? '暂停预览' : '播放预览'} onClick={togglePreviewPlayback}>{previewPlaying ? <Pause size={17} /> : <Play size={17} />}</button>
                      <button type="button" aria-label="预览下一剪辑点" title="下一剪辑点" onClick={() => jumpEditPoint('next')}><SkipForward size={14} /></button>
                    </div>
                    <input className="aicut-preview-seek" aria-label="预览播放头" type="range" min="0" max={Math.max(1, durationMs)} step="10" value={previewTimeMs} onChange={(event) => seekPreview(Number(event.target.value))} />
                    <PlayerViewControls quality={previewQuality} onQuality={value => { setPreviewPlaying(false); setPreviewQuality(value); }} zoom={canvasFit} onZoom={setCanvasFit} ratio={currentProject?.aspectRatio || '9:16'} onRatio={changeProjectRatio} fullscreen={previewFullscreen || previewExpanded} onFullscreen={togglePreviewFullscreen} safeGuide={safeGuide} onSafeGuide={() => setSafeGuide(value => !value)} />
                  </div>
                </section>}

                <aside className="aicut-inspector">
                  <nav>
                    {inspectorTabs.map((tab) => (
                      <button type="button" key={tab} className={inspectorTab === tab ? 'active' : ''} onClick={() => setInspectorTab(tab)}>
                        {tab}
                      </button>
                    ))}
                  </nav>
                  {selectedClipEntry ? (
                    <div className="aicut-inspector-body">
                      <div className="aicut-inspector-title">
                        <Film size={16} />
                        <strong>{selectedClipEntry.clip.name}</strong>
                      </div>
                      {inspectorTab === '基础' && (
                        <>
                          {selectedClipEntry.clip.kind === 'subtitle' ? (
                            <label>
                              文字内容
                              <Input.TextArea
                                value={selectedClipEntry.clip.text || ''}
                                rows={4}
                                onChange={(event) => {
                                  try {
                                    commitTimeline((current) => setTimelineClipText(current, selectedClipEntry.track.id, selectedClipEntry.clip.id, event.target.value));
                                  } catch (error) {
                                    if (event.target.value.trim()) message.error(error.message);
                                  }
                                }}
                              />
                            </label>
                          ) : null}
                          <label>
                            开始时间
                            <InputNumber min={0} value={selectedClipEntry.clip.startMs} onChange={(value) => moveClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, Number(value || 0))} addonAfter="ms" />
                          </label>
                          <label>
                            片段时长
                            <InputNumber min={1} value={selectedClipEntry.clip.durationMs} onChange={(value) => updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { durationMs: Number(value || 1) })} addonAfter="ms" />
                          </label>
                          <div>
                            <Button onClick={() => trimClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, 'left')}>左裁 0.1s</Button>
                            <Button onClick={() => trimClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, 'right')}>右裁 0.1s</Button>
                            {selectedClipEntry.clip.kind === 'subtitle' ? (
                              <Button
                                onClick={() => {
                                  try {
                                    runTimelineCommand({ type: 'clip.merge_next', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id });
                                  } catch (error) {
                                    message.warning(error.message);
                                  }
                                }}
                              >
                                合并下一条
                              </Button>
                            ) : null}
                          </div>
                        </>
                      )}
                      {inspectorTab === '速度' && (
                        <div className="aicut-speed-panel">
                          <label>
                            固定倍速
                            <Select value={Number(selectedClipEntry.clip.playbackRate) || 1} onChange={(value) => changeClipSpeed(selectedClipEntry.track.id, selectedClipEntry.clip.id, value)} options={[0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 3, 4].map((value) => ({ value, label: `${value}x` }))} />
                          </label>
                          {selectedClipEntry.clip.kind === 'video' ? (
                            <>
                              <div className="aicut-speed-heading">
                                <strong>曲线变速</strong>
                                <span>英雄时刻、蒙太奇、子弹时间、渐快、渐慢</span>
                              </div>
                              <SpeedCurveEditor points={selectedClipEntry.clip.speedCurve} rate={Number(selectedClipEntry.clip.playbackRate) || 1} onChange={(points) => {
                                try {
                                  runTimelineCommand({ type: 'clip.speed_curve_set', trackId: selectedClipEntry.track.id, clipId: selectedClipEntry.clip.id, points });
                                } catch (error) {
                                  message.warning(error?.message || '曲线变速失败');
                                }
                              }} />
                              <div className="aicut-speed-presets">
                                {speedCurvePresets.map((preset) => (
                                  <Button key={preset.id} size="small" type={JSON.stringify(selectedClipEntry.clip.speedCurve || []) === JSON.stringify(preset.points) ? 'primary' : 'default'} onClick={() => applyClipSpeedCurve(selectedClipEntry.track.id, selectedClipEntry.clip.id, preset.id)}>
                                    {preset.label}
                                  </Button>
                                ))}
                              </div>
                              <div className="aicut-speed-summary">
                                <span>平均倍率 {(Number(selectedClipEntry.clip.playbackRate) || 1).toFixed(2)}x</span>
                                <span>输出 {(Number(selectedClipEntry.clip.durationMs) / 1000).toFixed(2)} 秒</span>
                              </div>
                            </>
                          ) : (
                            <span className="aicut-speed-hint">音频片段使用固定倍速；曲线变速仅用于视频。</span>
                          )}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="aicut-inspector-empty">
                      <strong>选择时间线片段</strong>
                      <span>片段参数会显示在这里</span>
                    </div>
                  )}
                </aside>

                <section className={`aicut-timeline aicut-reference-timeline${timelineCollapsed ? ' collapsed' : ''}`}>
                  <TimelineTabs items={projectTimelines} activeId={activeTimelineId} busy={loading || saving || timelineBusy || timelineNetworkBusy}
                    onSelect={changeActiveTimeline}
                    onCreate={title => timelineOperation(async () => { const item = await createProjectTimeline(projectId, title); setLoading(true); setActiveTimelineId(item.id); })}
                    onRename={(id, title) => timelineOperation(() => renameProjectTimeline(projectId, id, title))}
                    onDuplicate={(id, title) => timelineOperation(async () => { const item = await duplicateProjectTimeline(projectId, id, title); setLoading(true); setActiveTimelineId(item.id); })}
                    onDelete={id => timelineOperation(async () => { await deleteProjectTimeline(projectId, id); if (id === activeTimelineId) { setLoading(true); setActiveTimelineId('default'); } })}
                    onReorder={ids => timelineOperation(() => reorderProjectTimelines(projectId, ids)).catch(() => {})}
                  />
                  <TimelineToolbar
                    primary={<>
                      <Dropdown disabled={toolbarState.busy} trigger={['click']} menu={{ items: [{key:'video',label:'新增视频轨'},{key:'audio',label:'新增音频轨'},{key:'subtitle',label:'新增字幕轨'}], onClick:({key})=>addTrack(key) }}><Button disabled={toolbarState.busy} size="small" aria-label="新增轨道" title="新增轨道" icon={<Plus size={15}/>} /></Dropdown>
                      <Button size="small" aria-label="选择模式" title="选择模式" aria-pressed={!bladeMode} type={!bladeMode ? 'primary' : 'default'} onClick={() => setBladeMode(false)} icon={<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6"><path d="m5 3 14 11-7 1-3 6z"/></svg>} />
                      <span className="aicut-toolbar-divider" aria-hidden="true" />
                      <Button size="small" type={bladeMode ? 'primary' : 'default'} aria-label="剃刀模式" aria-pressed={bladeMode} title="剃刀模式" disabled={toolbarState.busy} icon={<Scissors size={15} />} onClick={() => setBladeMode(value => !value)} />
                    </>}
                    primaryActions={toolbarPrimaryActions}
                    advancedItems={[{ key:'clipboard-actions',type:'group',label:'剪贴板',children:[{key:'copy',label:'复制',disabled:!selectedClipKeys.length || toolbarState.busy,onClick:copySelectedClips},{key:'paste',label:'粘贴',disabled:!clipboardCount || toolbarState.busy,onClick:pasteClips}] }, ...buildTimelineAdvancedItems(toolbarState, {
                      cut: cutSelectedClips, duplicate: duplicateSelectedClip, consolidate: consolidateSelectedClips,
                      previousEdit: () => jumpEditPoint('previous'), nextEdit: () => jumpEditPoint('next'),
                      nudgeBack: () => nudgeSelectedClip(-1), nudgeForward: () => nudgeSelectedClip(1),
                      addKeyframe: addKeyframeAtPlayhead, removeKeyframe: removeKeyframeAtPlayhead,
                      addMarker: addMarkerAtPlayhead, removeMarker: removeMarkerAtPlayhead,
                      link: linkSelectedClips, unlink: unlinkSelectedClips, createCompound: createCompoundClip,
                      expandCompound: expandSelectedCompound,
                      openCompound: () => selectedClipEntry?.clip?.compoundId && setActiveCompoundId(selectedClipEntry.clip.compoundId),
                      dissolveCompound: dissolveSelectedCompound
                    }, (selectedClipEntry?.track || timeline.tracks[0]) ? trackMenuItems(selectedClipEntry?.track || timeline.tracks[0]) : [], toolbarPrimaryActions.map(action => action.key))]}
                    secondary={<>
                      <AudioRecorder key={`timeline-recorder:${projectId}:${activeTimelineId}`} compact onUpload={uploadRecording} />
                      <span className="aicut-toolbar-divider" aria-hidden="true" />
                      <Button size="small" type={linkedMove ? 'primary' : 'default'} aria-pressed={linkedMove} title="联动：已建立联动的片段整组移动" onClick={() => setLinkedMove(value => !value)}>联动</Button>
                      <Button size="small" type={rippleMode ? 'primary' : 'default'} aria-pressed={rippleMode} title="波纹：删减片段后收紧空隙" onClick={() => setRippleMode(value => !value)}>波纹</Button>
                      <Button size="small" type={magneticSnap ? 'primary' : 'default'} aria-pressed={magneticSnap} title="吸附" aria-label="吸附" icon={<Magnet size={15} />} onClick={() => setMagneticSnap(value => !value)} />
                      <Button size="small" type={editorLayout.playheadFollow ? 'primary' : 'default'} aria-pressed={editorLayout.playheadFollow} title="跟随播放头" onClick={() => setEditorLayout(current => ({ ...current, playheadFollow: !current.playheadFollow }))}>跟随</Button>
                      <Select size="small" aria-label="时间线标记点" placeholder="标记" value={undefined} style={{ width: 76 }}
                        onChange={markerId => { const marker = (timeline.markers || []).find(item => item.id === markerId); if (marker) seekPreview(marker.timeMs); }}
                        options={(timeline.markers || []).map(marker => ({ value: marker.id, label: `${previewClock(marker.timeMs)} ${marker.label}` }))} />
                      <TimelineZoomControls zoom={timelineZoom} onChange={setTimelineZoom} durationMs={durationMs} viewportWidth={timelineScrollRef.current?.clientWidth || 900} />
                      <button type="button" className="aicut-timeline-collapse" aria-label={timelineCollapsed ? '展开时间线' : '收起时间线'}
                        onClick={() => setEditorLayout(current => ({ ...current, timelineCollapsed: !current.timelineCollapsed }))}>
                        {timelineCollapsed ? <PanelBottomOpen size={15} /> : <PanelBottomClose size={15} />}
                        <span>{timelineCollapsed ? '展开时间线' : '收起时间线'}</span>
                      </button>
                    </>}
                  />
                  <div className="aicut-timeline-ruler-row">
                    <div className="aicut-ruler-clock">{previewClock(previewTimeMs)}</div>
                    <div className="aicut-ruler-viewport" ref={timelineRulerScrollRef}>
                      <div className="aicut-shared-ruler" style={{ width: `max(100%, ${Math.max(900, (editorShellRef.current?.clientWidth || 0) - 130, timelineTimeToPixels(durationMs + 2000, timelineZoom) + 20)}px)` }} onPointerDown={event => seekPreview(timelinePixelsToTime(event.clientX - event.currentTarget.getBoundingClientRect().left, timelineZoom))}>
                        {timelineRulerTicks(Math.max(900, (editorShellRef.current?.clientWidth || 0) - 130, timelineTimeToPixels(durationMs + 2000, timelineZoom) + 20), timelineZoom).map(tick => <span key={tick.timeMs} style={{ left: tick.left }}>{tick.label}</span>)}
                      </div>
                    </div>
                  </div>
                  {renderUnifiedTimeline()}
                </section>
              </EditorShell>
            )}

            {activeCompoundId ? (
              <CompoundTimelineWorkspace
                timeline={timeline}
                compoundId={activeCompoundId}
                playheadMs={previewTimeMs}
                selectedClipKeys={selectedClipKeys}
                onToggleSelection={(key) => toggleTimelineSelection([key])}
                onClose={(compoundId) => setActiveCompoundId(compoundId || '')}
                onOpenCompound={(compoundId) => setActiveCompoundId(compoundId)}
                onDissolve={(compoundId) => dissolveCompound(compoundId)}
                onSplit={(compoundId, timeMs) => splitCompoundClip(compoundId, timeMs)}
                onAddTrack={addCompoundTrack}
                onRemoveTrack={removeCompoundTrack}
                onReorderTrack={reorderCompoundTracks}
                onDropAsset={dropAssetOnCompoundTrack}
                onCommand={(command) => {
                  try {
                    runTimelineCommand(command);
                  } catch (error) {
                    message.warning(error?.message || '复合片段内部编辑失败');
                  }
                }}
              />
            ) : null}

            {asrJobs.length ? (
              <section className="aicut-asr-jobs">
                <div className="aicut-panel-title">
                  <strong>语音识别任务</strong>
                </div>
                {asrJobs.slice(0, 5).map((job) => (
                  <div className="aicut-asr-job" key={job.id}>
                    <div>
                      <strong>{assets.find((asset) => asset.id === job.assetId)?.name || job.assetId}</strong>
                      <small>
                        {asrStateLabels[job.state] || job.state} · {job.progress || 0}%{job.error ? ` · ${job.error}` : ''}
                      </small>
                    </div>
                    <Tag>{asrStateLabels[job.state] || job.state}</Tag>
                    {['queued', 'running', 'cancel_requested'].includes(job.state) ? (
                      <Button size="small" onClick={() => cancelASR(job)}>
                        取消
                      </Button>
                    ) : ['failed', 'canceled'].includes(job.state) ? (
                      <Button
                        size="small"
                        onClick={() => {
                          setSelectedAssetId(job.assetId);
                          setASROpen(true);
                        }}
                      >
                        重新识别
                      </Button>
                    ) : null}
                  </div>
                ))}
              </section>
            ) : null}
          </Spin>
        )}

        <Modal
          title="键盘快捷键"
          open={shortcutOpen}
          onCancel={() => setShortcutOpen(false)}
          footer={
            <Button type="primary" onClick={() => setShortcutOpen(false)}>
              知道了
            </Button>
          }
          width={760}
        >
          <div className="aicut-shortcut-grid">
            <section>
              <h3>通用</h3>
              <p>
                <span>新建项目</span>
                <kbd>Ctrl</kbd>
                <kbd>N</kbd>
              </p>
              <p>
                <span>保存工程</span>
                <kbd>Ctrl</kbd>
                <kbd>S</kbd>
              </p>
              <p>
                <span>导出作品</span>
                <kbd>Ctrl</kbd>
                <kbd>E</kbd>
              </p>
            </section>
            <section>
              <h3>视图</h3>
              <p>
                <span>AI 创作画布</span>
                <kbd>Ctrl</kbd>
                <kbd>1</kbd>
              </p>
              <p>
                <span>专业剪辑</span>
                <kbd>Ctrl</kbd>
                <kbd>2</kbd>
              </p>
              <p>
                <span>快捷键</span>
                <kbd>Ctrl</kbd>
                <kbd>/</kbd>
              </p>
            </section>
            <section>
              <h3>播放与剪辑</h3>
              <p>
                <span>播放 / 暂停</span>
                <kbd>Space</kbd>
              </p>
              <p>
                <span>分割</span>
                <kbd>Ctrl</kbd>
                <kbd>B</kbd>
              </p>
              <p>
                <span>删除</span>
                <kbd>Delete</kbd>
              </p>
              <p>
                <span>剪切 / 复制 / 粘贴</span>
                <kbd>Ctrl</kbd>
                <kbd>X / C / V</kbd>
              </p>
              <p>
                <span>逐帧左移 / 右移</span>
                <kbd>Alt</kbd>
                <kbd>← / →</kbd>
              </p>
              <p>
                <span>上一 / 下一剪辑点</span>
                <kbd>Ctrl</kbd>
                <kbd>← / →</kbd>
              </p>
            </section>
            <section>
              <h3>历史</h3>
              <p>
                <span>撤销</span>
                <kbd>Ctrl</kbd>
                <kbd>Z</kbd>
              </p>
              <p>
                <span>重做</span>
                <kbd>Ctrl</kbd>
                <kbd>Shift</kbd>
                <kbd>Z</kbd>
              </p>
            </section>
          </div>
        </Modal>
        <Modal title="新建 AI-CUT 项目" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={createProject} okText="创建">
          <div className="aicut-create-form">
            <label>
              项目名称
              <Input value={newTitle} onChange={(event) => setNewTitle(event.target.value)} placeholder="例如：第一条竖屏口播" />
            </label>
            <label>
              画面比例
              <Select
                value={newRatio}
                onChange={setNewRatio}
                options={[
                  { value: '9:16', label: '9:16 竖屏' },
                  { value: '16:9', label: '16:9 横屏' },
                  { value: '1:1', label: '1:1 方形' }
                ]}
              />
            </label>
          </div>
        </Modal>
        <Modal
          title="重命名轨道"
          open={Boolean(trackEditId)}
          onCancel={() => {
            setTrackEditId('');
            setTrackNameDraft('');
          }}
          onOk={confirmTrackRename}
          okText="保存"
        >
          <Input value={trackNameDraft} maxLength={64} autoFocus onPressEnter={confirmTrackRename} onChange={(event) => setTrackNameDraft(event.target.value)} placeholder="输入轨道名称" />
        </Modal>
        <Modal title="字幕纠错" open={captionOpen} onCancel={cancelCaptionCorrection} onOk={correctCaptions} okText="开始纠错" confirmLoading={correctingCaptions}>
          <div className="aicut-create-form">
            <label>
              校对剧本
              <Input.TextArea value={captionScript} onChange={(event) => setCaptionScript(event.target.value)} rows={8} />
            </label>
            <label>
              文本模型（可选）
              <Input value={captionModelId} onChange={(event) => setCaptionModelId(event.target.value)} />
            </label>
          </div>
        </Modal>
        <Modal
          title="识别字幕"
          open={asrOpen}
          onCancel={() => setASROpen(false)}
          onOk={startASR}
          okText="开始识别"
          okButtonProps={{
            disabled: !asrModels.length || !assets.some((asset) => ['audio', 'video'].includes(asset.kind))
          }}
          confirmLoading={asrBusy}
        >
          <div className="aicut-create-form">
            <label>
              识别素材
              <Select value={selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? selectedAsset.id : undefined} onChange={setSelectedAssetId} placeholder="选择视频或音频" options={assets.filter((asset) => ['audio', 'video'].includes(asset.kind)).map((asset) => ({ value: asset.id, label: asset.name }))} />
            </label>
            <label>
              语音识别模型
              <Select
                value={asrModelId || undefined}
                onChange={setASRModelId}
                placeholder="选择已启用的 ASR 模型"
                options={asrModels.map((model) => ({
                  value: model.id,
                  label: model.displayName || model.modelId || model.id
                }))}
              />
            </label>
            {!asrModels.length ? <Button onClick={() => window.location.assign('/settings')}>前往 API 配置</Button> : null}
            <label>
              语言
              <Select
                value={asrLanguage}
                onChange={setASRLanguage}
                options={[
                  { value: 'zh', label: '中文' },
                  { value: 'en', label: '英文' },
                  { value: '', label: '自动检测' }
                ]}
              />
            </label>
            <small>识别成功后会写入新的字幕轨，不会覆盖现有字幕。任务可取消，失败后可以重试。</small>
          </div>
        </Modal>
        <Modal title="背景音乐波形与节拍" open={bgmOpen} footer={null} onCancel={() => setBGMOpen(false)} width={680}>
          {bgmAnalysisBusy ? <Spin tip="正在低内存分析音频…" /> : null}
          {bgmAudioAnalysis ? (
            <section className="aicut-bgm-analysis">
              <div className="aicut-bgm-waveform" aria-label="音频波形">
                {bgmAudioAnalysis.peaks.map((peak, index) => <i key={index} style={{ height: `${Math.max(2, Number(peak) * 100)}%` }} />)}
              </div>
              <p>检测到 {bgmAudioAnalysis.beatsMs.length} 个节拍点；加入时间线后会生成可吸附的绿色节拍标记。</p>
              <Button type="primary" onClick={() => applyBGMPlan(assets.find((asset) => asset.id === bgmAudioAnalysis.assetId))}>按节拍加入背景音乐</Button>
            </section>
          ) : null}
          <div className="aicut-bgm-list">
            {assets
              .filter((asset) => asset.kind === 'audio')
              .map((asset) => (
                <button type="button" key={asset.id} className={bgmAudioAnalysis?.assetId === asset.id ? 'active' : ''} onClick={() => analyzeBGMAudio(asset)}>
                  <strong>{asset.name}</strong>
                  <span>{bgmAudioAnalysis?.assetId === asset.id ? '已分析' : '分析波形与节拍'}</span>
                </button>
              ))}
          </div>
        </Modal>
        <Modal title="批量替换字幕" open={replaceOpen} onCancel={() => setReplaceOpen(false)} onOk={replaceSubtitleText}>
          <div className="aicut-create-form">
            <label>
              查找
              <Input value={replaceFrom} onChange={(event) => setReplaceFrom(event.target.value)} />
            </label>
            <label>
              替换为
              <Input value={replaceTo} onChange={(event) => setReplaceTo(event.target.value)} />
            </label>
          </div>
        </Modal>
        <Modal title="画布背景" open={canvasSettingsOpen} onCancel={() => setCanvasSettingsOpen(false)} onOk={saveCanvasSettings} okText="保存设置" confirmLoading={canvasSettingsSaving}>
          <div className="aicut-canvas-settings">
            <label>
              背景模式
              <Select aria-label="画布背景模式" value={canvasDraft.backgroundMode} options={[{ value: 'color', label: '背景色' }, { value: 'transparent', label: '透明' }]} onChange={(backgroundMode) => setCanvasDraft((current) => ({ ...current, backgroundMode }))} />
            </label>
            <label>
              背景颜色
              <input aria-label="画布背景颜色" type="color" value={canvasDraft.backgroundColor} disabled={canvasDraft.backgroundMode === 'transparent'} onChange={(event) => setCanvasDraft((current) => ({ ...current, backgroundColor: event.target.value.toUpperCase() }))} />
            </label>
            <p>{canvasDraft.backgroundMode === 'transparent' ? '透明区域会显示棋盘格；导出时可选择 WebM VP9 或 MOV ProRes 4444 保留透明通道。' : '背景颜色会立即用于播放器和普通成片导出。'}</p>
          </div>
        </Modal>
        <Modal title="导出" open={exportOpen} footer={null} onCancel={() => setExportOpen(false)} width={720}>
          <div className="aicut-export-panel">
            <div>
              <strong>{currentProject?.title || '未命名工程'}</strong>
              <span>{currentProject?.aspectRatio || '9:16'} · {exportSettings.resolution.toUpperCase()} · {exportSettings.frameRate} fps · {exportSettings.codec.toUpperCase()} · {exportSettings.bitDepth}-bit</span>
            </div>
            <div className="aicut-export-settings">
              <strong>导出设置</strong>
              <label>封装格式<Select aria-label="导出封装格式" value={exportSettings.format} options={[{ value: 'mp4', label: 'MP4' }, { value: 'webm', label: 'WebM' }, { value: 'mov', label: 'MOV' }]} onChange={(format) => setExportSettings((current) => ({ ...current, format, codec: format === 'mov' ? 'prores4444' : format === 'webm' ? 'vp9' : 'h264', bitDepth: format === 'mov' ? 12 : 8, alpha: format === 'mp4' ? false : current.alpha }))} /></label>
              <label>分辨率<Select aria-label="导出分辨率" value={exportSettings.resolution} options={[{ value: '720p', label: '720P' }, { value: '1080p', label: '1080P' }, { value: '4k', label: '4K' }]} onChange={(resolution) => setExportSettings((current) => ({ ...current, resolution }))} /></label>
              <label>帧率<Select aria-label="导出帧率" value={exportSettings.frameRate} options={[{ value: 24, label: '24 fps' }, { value: 25, label: '25 fps' }, { value: 30, label: '30 fps' }, { value: 50, label: '50 fps' }, { value: 60, label: '60 fps' }]} onChange={(frameRate) => setExportSettings((current) => ({ ...current, frameRate }))} /></label>
              <label>质量<Select aria-label="导出质量" value={exportSettings.quality} options={[{ value: 'draft', label: '快速预览' }, { value: 'standard', label: '标准' }, { value: 'high', label: '高质量' }]} onChange={(quality) => setExportSettings((current) => ({ ...current, quality }))} /></label>
              <label>编码器<Select aria-label="导出编码器" value={exportSettings.codec} options={(exportSettings.format === 'mov' ? [{ value: 'prores4444', label: 'ProRes 4444' }] : exportSettings.format === 'webm' ? [{ value: 'vp9', label: 'VP9' }, { value: 'av1', label: 'AV1' }] : [{ value: 'h264', label: 'H.264' }, { value: 'hevc', label: 'HEVC / H.265' }, { value: 'av1', label: 'AV1' }])} onChange={(codec) => setExportSettings((current) => ({ ...current, codec, bitDepth: codec === 'h264' ? 8 : current.bitDepth, alpha: codec === 'av1' ? false : current.alpha }))} /></label>
              <label>硬件加速<Select aria-label="硬件加速" value={exportSettings.hardwareAcceleration} options={[{ value: 'auto', label: '自动选择（失败回退软件）' }, { value: 'software', label: '强制软件编码' }]} onChange={(hardwareAcceleration) => setExportSettings((current) => ({ ...current, hardwareAcceleration }))} /><small>仅服务端显卡编码器通过实际试跑后才启用</small></label>
              <label>色深<Select aria-label="导出色深" value={exportSettings.bitDepth} options={[{ value: 8, label: '8-bit', disabled: exportSettings.codec === 'prores4444' }, { value: 10, label: '10-bit', disabled: exportSettings.codec === 'h264' || exportSettings.codec === 'prores4444' }, { value: 12, label: '12-bit', disabled: exportSettings.codec !== 'prores4444' }]} onChange={(bitDepth) => setExportSettings((current) => ({ ...current, bitDepth }))} /></label>
              <label>目标码率<InputNumber aria-label="导出目标码率" min={0} max={100} precision={0} value={exportSettings.bitrateMbps} addonAfter="Mbps" onChange={(bitrateMbps) => setExportSettings((current) => ({ ...current, bitrateMbps: Number(bitrateMbps || 0) }))} /><small>0 表示按质量自动分配</small></label>
              <label className="aicut-export-alpha"><Checkbox checked={Boolean(exportSettings.alpha)} disabled={currentProject?.backgroundMode !== 'transparent' || exportSettings.format === 'mp4' || exportSettings.codec === 'av1'} onChange={(event) => setExportSettings((current) => ({ ...current, alpha: event.target.checked }))}>保留透明通道</Checkbox><small>仅透明工程的 VP9 / ProRes 4444 可用</small></label>
            </div>
            <p className="aicut-export-depth-note" role="note">色深是输出编码位深。当前效果合成仍以 8-bit 处理，10/12-bit 输出不代表 HDR 或端到端高位深。</p>
            <div className="aicut-export-types">
              <Button type="primary" loading={exportingFormat === exportSettings.format} onClick={() => exportMedia(exportSettings.format)}>
                成片 · {exportSettings.format.toUpperCase()} 加入队列
              </Button>
              <Button loading={exportingFormat === 'mp3'} onClick={() => exportMedia('mp3')}>
                音轨 · MP3 加入队列
              </Button>
              <Button loading={exportingFormat === 'batch'} onClick={exportMediaBatch}>
                成片 + MP3 加入队列
              </Button>
              <Button type="primary" onClick={exportSRT}>
                字幕文件 · SRT
              </Button>
            </div>
            <ExportQueuePanel jobs={exportJobs} error={exportQueueError} onReload={reloadExportQueue} busyJobId={exportActionBusy} onCancel={cancelExport} onDownload={downloadExport} onRetry={retryExport} />
            <p>任务按提交时保存的时间线快照渲染；关闭弹窗或刷新页面不会丢失队列。</p>
          </div>
        </Modal>
        <Modal title="识别并删除水词字幕" open={fillerOpen} onCancel={cancelFillerTrim} onOk={trimFillers} confirmLoading={trimmingFillers}>
          <label>
            文本模型（可选）
            <Input value={fillerModelId} onChange={(event) => setFillerModelId(event.target.value)} />
          </label>
        </Modal>
        <Modal title="去口误停顿" open={oralOpen} onCancel={cancelFillerTrim} onOk={cleanOralTimeline} confirmLoading={trimmingFillers}>
          <label>
            文本模型（可选）
            <Input value={fillerModelId} onChange={(event) => setFillerModelId(event.target.value)} />
          </label>
        </Modal>
        <Modal title="声音克隆" open={voiceCloneOpen} footer={null} onCancel={cancelVoiceClone} closable={!voiceCloneBusy} maskClosable={!voiceCloneBusy} width={620}>
          <section className="aicut-voice-clone" aria-label="声音克隆">
            <p>从当前工程音频提取音色特征，生成新的试听音频。原参考音频不会被修改。</p>
            <label>参考音频<Select aria-label="声音克隆参考音频" value={voiceCloneDraft.assetId || undefined} disabled={voiceCloneBusy} options={assets.filter((asset) => asset.kind === 'audio').map((asset) => ({ value: asset.id, label: asset.name }))} onChange={(assetId) => setVoiceCloneDraft((current) => ({ ...current, assetId }))} /></label>
            {assets.find((asset) => asset.id === voiceCloneDraft.assetId)?.url ? <audio controls preload="metadata" src={assets.find((asset) => asset.id === voiceCloneDraft.assetId)?.url} /> : null}
            <label>试听文本<Input.TextArea aria-label="声音克隆试听文本" value={voiceCloneDraft.text} disabled={voiceCloneBusy} rows={4} maxLength={1000} showCount placeholder="输入要用克隆音色朗读的内容" onChange={(event) => setVoiceCloneDraft((current) => ({ ...current, text: event.target.value }))} /></label>
            <div className="aicut-tts-fields">
              <label>素材名称<Input aria-label="克隆音频素材名称" value={voiceCloneDraft.name} disabled={voiceCloneBusy} maxLength={64} onChange={(event) => setVoiceCloneDraft((current) => ({ ...current, name: event.target.value }))} /></label>
              <label>情绪<Select aria-label="克隆音色情绪" value={voiceCloneDraft.style} disabled={voiceCloneBusy} options={ttsStyleOptions} onChange={(style) => setVoiceCloneDraft((current) => ({ ...current, style }))} /></label>
              <label>语速<InputNumber aria-label="克隆音色语速" min={0.5} max={2} step={0.1} value={voiceCloneDraft.speed} disabled={voiceCloneBusy} onChange={(speed) => setVoiceCloneDraft((current) => ({ ...current, speed: Number(speed ?? 1) }))} /></label>
              <label>音调<InputNumber aria-label="克隆音色音调" min={-50} max={50} value={voiceCloneDraft.pitch} disabled={voiceCloneBusy} onChange={(pitch) => setVoiceCloneDraft((current) => ({ ...current, pitch: Number(pitch ?? 0) }))} /></label>
            </div>
            <small>需要管理员配置支持声音克隆的音频模型；服务密钥和请求模板只保存在后端。</small>
            <div className="aicut-tool-actions">
              <Button disabled={voiceCloneBusy} onClick={() => runVoiceClone(false)}>仅生成克隆音频</Button>
              <Button type="primary" loading={voiceCloneBusy} onClick={() => runVoiceClone(true)}>克隆并加入时间线</Button>
              {voiceCloneBusy ? <Button danger onClick={cancelVoiceClone}>取消生成</Button> : <Button onClick={cancelVoiceClone}>取消</Button>}
            </div>
          </section>
        </Modal>
        <Modal title="图片背景移除" open={backgroundRemovalOpen} footer={null} onCancel={cancelBackgroundRemoval} closable={!backgroundRemovalBusyId} maskClosable={!backgroundRemovalBusyId}>
          <section className="aicut-background-removal" aria-label="图片背景移除">
            <div className="aicut-background-removal-preview">
              {backgroundRemovalAsset ? <img src={backgroundRemovalAsset.url} alt={backgroundRemovalAsset.name || '待去背图片'} /> : null}
              <div><strong>{backgroundRemovalAsset?.name}</strong><small>保留原图，生成一张新的透明 PNG 素材。</small></div>
            </div>
            <label>素材名称<Input aria-label="透明素材名称" value={backgroundRemovalName} maxLength={64} disabled={Boolean(backgroundRemovalBusyId)} onChange={(event) => setBackgroundRemovalName(event.target.value)} /></label>
            <p>使用账号中已配置且支持图片编辑的模型。服务必须返回真正的透明像素，否则不会创建素材。</p>
            <div className="aicut-tool-actions">
              <Button disabled={Boolean(backgroundRemovalBusyId)} onClick={() => runBackgroundRemoval(false)}>仅生成透明素材</Button>
              <Button type="primary" loading={Boolean(backgroundRemovalBusyId)} onClick={() => runBackgroundRemoval(true)}>去背并加入时间线</Button>
              {backgroundRemovalBusyId ? <Button danger onClick={cancelBackgroundRemoval}>取消生成</Button> : <Button onClick={cancelBackgroundRemoval}>取消</Button>}
            </div>
          </section>
        </Modal>
        {conflictRecovery ? <Modal open title="保留未保存的编辑" okText="另存为新时间线" cancelText="继续检查" confirmLoading={recoveringDraft} onCancel={() => !recoveringDraft && setConflictRecovery(null)} onOk={async () => {
          if (conflictRecovery.scope !== activeTimelineScopeRef.current) return setConflictRecovery(null);
          setRecoveringDraft(true);
          try {
            const result = await saveTimelineDraftCopy(projectId, '冲突恢复副本', timelineRef.current, conflictRecovery.copyId);
            setConflictRecovery(null); setLoading(true); setActiveTimelineId(result.id); await refreshTimelineList();
          } catch (error) { setConflictRecovery(current => current && ({ ...current, copyId: error.createdTimelineId || current.copyId })); message.error(error.message); }
          finally { setRecoveringDraft(false); }
        }}><p>当前时间线已有其他保存结果。你的编辑仍保留在页面中，可以另存为独立时间线，同时保留原时间线。</p></Modal> : null}
        {multiCamView ? <Modal open title="多机位剪辑" footer={null} width={1040} onCancel={() => setMultiCamView(null)}>
          <MultiCamWorkspace assets={assets} onAutoAlign={input => alignAICutMultiCamAudio(projectId, input)} onCancel={() => setMultiCamView(null)} onApply={plan => {
            if (multiCamView.scope !== activeTimelineScopeRef.current || multiCamView.snapshot !== timelineRef.current) throw new Error('工程已变化，请重新打开多机位方案');
            const result = runTimelineCommand({ type: 'timeline.multicam_apply', plan, assets });
            setSelectedClipKey(result.selectionKey); setSelectedClipKeys([result.selectionKey]); setSourcePreviewAsset(null);
            seekPreview(plan.summary.startMs); setMultiCamView(null); message.success('机位切换已生成可编辑片段，可整次撤销');
          }} />
        </Modal> : null}
        {templateApplication ? <ProjectTemplateDialog
          key={`${projectId}:${templateApplication.id}:${templateApplication.startMs}`}
          template={projectTemplateCatalog.find(template => template.id === templateApplication.id)} timeline={timeline} assets={assets} startMs={templateApplication.startMs}
          onClose={() => setTemplateApplication(null)}
          onApply={({ templateId, options, reviewedTimeline }) => {
            if (reviewedTimeline !== timelineRef.current) return message.warning('工程刚刚发生变化，请重新检查模板预览');
            try {
              const result = runTimelineCommand({ type: 'timeline.template_apply', templateId, assets, options });
              setSelectedClipKey(result.selectionKey);
              setSelectedClipKeys([result.selectionKey]);
              setSourcePreviewAsset(null);
              seekPreview(options.startMs);
              setTemplateApplication(null);
              message.success(`已新增 ${result.metadata.template.trackCount} 条模板轨道，可撤销整次操作`);
            } catch (error) { message.warning(error.message || '模板应用失败'); }
          }}
        /> : null}
        <Modal title={toolWorkbenchOpen} open={Boolean(toolWorkbenchOpen)} footer={null} width={820} onCancel={() => setToolWorkbenchOpen('')}>
          <div className="aicut-tool-workbench">{renderMaterialBody()}</div>
        </Modal>
        <AICutPetHost scope={petServiceScope} />
      </main>
    </ConfigProvider>
  );
}
