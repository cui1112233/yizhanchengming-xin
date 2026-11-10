import { Button, Checkbox, ConfigProvider, Dropdown, Empty, Input, InputNumber, Modal, Progress, Select, Spin, Tag, Upload, message } from 'antd';
import { Bot, Captions, Clapperboard, CloudUpload, Copy, Film, ImageIcon, Keyboard, Magnet, Music, PanelBottomClose, PanelBottomOpen, PanelLeftClose, PanelLeftOpen, Pause, Play, Plus, Redo2, Save, Scissors, Sparkles, Star, Trash2, Type, Undo2, Upload as UploadIcon, WandSparkles, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelAICutASRJob, cancelAICutExportJob, cleanAICutOralTimeline, correctAICutCaptions, createAICutASRJob, createAICutExportJob, createAICutProject, downloadAICutExportJob, extractAICutAudio, extractAICutVisualFrames, getAICutExportJob, getAICutTimelineHistory, importAICutSRT, listAICutASRJobs, listAICutAssets, listAICutProjects, loadAICutTimeline, redoAICutTimeline, saveAICutTimeline, smartSplitAICutVideo, trimAICutFillers, undoAICutTimeline, uploadAICutAsset } from '../../shared/api/aicut.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import AICutAgentPanel from '../components/AICutAgentPanel.jsx';
import { addEffectLayer, addTimelineTrack, draggedTimelineStart, markTimelineClip, mergeTimelineClipWithNext, moveTimelineClip, normalizeSubtitleRhythm, planTimelineBGM, removeTimelineTrack, renameTimelineTrack, setClipPlaybackRate, setClipTransition, setTimelineClipText, splitTimelineClip, splitTimelineClipAt, subtitleRhythmLabel, timelinePixelsToTime, timelineTimeToPixels, toggleTimelineTrackFlag, trimTimelineClipEdge, updateTimelineTrackSettings } from './aicutTimeline.js';
import { buildAICutPreviewFrame, getAICutTimelineDuration, nextAICutCanvasFit, seekAICutPreviewTime } from './aicutPreview.js';
import { resolveAICutShortcut, timelineDropTime } from './aicutWorkbench.js';
import { createAICutCapabilities, filterAICutCapabilities, recordRecentAICutFeature, toggleFavoriteAICutFeature } from './aicutCapabilities.js';
import { normalizeVisualEffect, subtitleAnimationCatalog, subtitleStyleCatalog, textStyleCatalog, transitionCatalog, visualEffectCatalog, visualEffectPreviewStyle, visualEffectPreset } from './aicutEffects.js';
import EditorShell from './aicut/EditorShell.jsx';
import ResourceLibrary from './aicut/ResourceLibrary.jsx';
import PreviewCanvas from './aicut/PreviewCanvas.jsx';
import InspectorPanel from './aicut/InspectorPanel.jsx';
import CaptionWorkspace from './aicut/CaptionWorkspace.jsx';
import { normalizeEditorLayout } from './aicut/editorLayout.js';
import './aicut-page.css';
import './aicut/editor.css';

const kindLabels = { image: '图片', video: '视频', audio: '音频', subtitle: '字幕' };
const trackLabels = { video: '画面轨道', audio: '音频轨道', subtitle: '字幕轨道' };
const referenceTrackNames = { subtitle: '字幕 1', effect: '特效 1', video: '视频 1', audio: '音频 1' };
const asrStateLabels = { queued: '排队中', running: '识别中', cancel_requested: '正在取消', succeeded: '已写入字幕', failed: '失败', canceled: '已取消' };
const editorTools = [
  ['媒体', ImageIcon], ['音频', Music], ['文本', Type], ['特效', Sparkles], ['转场', Film],
  ['字幕', Captions], ['AI工具', WandSparkles]
];
const inspectorTabs = ['基础', '速度'];
const toolSourceLabels = { browser: '浏览器', cloud: '云服务', ai: 'AI模型', render: 'Go后台' };
const catalogPresetMap = {
  'filter-original': 'original', 'filter-bright': 'bright', 'filter-warm-sun': 'warm', 'filter-cool-blue': 'cool-cinema',
  'filter-monochrome': 'grayscale', 'filter-soft': 'soft-focus', 'effect-blur': 'soft-focus', 'effect-vignette': 'vignette',
  'effect-glow': 'glow', 'effect-shake': 'shake', 'effect-breathe-zoom': 'breathe', 'effect-flash': 'flash', 'effect-mosaic': 'mosaic',
};

function visualEffectForCatalogResource(resource) {
  const preset = catalogPresetMap[resource?.id];
  if (preset) return visualEffectPreset(preset);
  if (resource?.category === 'filter') return normalizeVisualEffect({ preset: 'custom', contrast: 1.08, saturation: 1.12, intensity: .5, speed: 1 });
  if (resource?.category === 'effect') return normalizeVisualEffect({ preset: 'custom', intensity: .5, speed: 1 });
  return null;
}
function newID(prefix) {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
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

function srtTime(ms) {
  const value = Math.max(0, Number(ms) || 0);
  const hours = Math.floor(value / 3600000);
  const minutes = Math.floor((value % 3600000) / 60000);
  const seconds = Math.floor((value % 60000) / 1000);
  const millis = Math.floor(value % 1000);
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')},${String(millis).padStart(3, '0')}`;
}

function subtitleClips(timeline) {
  return timeline.tracks.filter(track => track.kind === 'subtitle').flatMap(track => track.clips || []).filter(clip => String(clip.text || '').trim()).sort((a, b) => Number(a.startMs) - Number(b.startMs));
}

function previewClock(ms) {
  const value = Math.max(0, Number(ms) || 0);
  const minutes = Math.floor(value / 60000);
  const seconds = Math.floor((value % 60000) / 1000);
  const hundredths = Math.floor((value % 1000) / 10);
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}.${String(hundredths).padStart(2, '0')}`;
}

function PreviewMedia({ item, playing }) {
  const mediaRef = useRef(null);
  useEffect(() => {
    const media = mediaRef.current;
    if (!media || item.kind === 'image') return;
    const expected = item.sourceTimeMs / 1000;
    if (Number.isFinite(expected) && Math.abs(media.currentTime - expected) > 0.2) media.currentTime = expected;
    media.playbackRate = item.playbackRate || 1;
    if (item.kind === 'audio') media.volume = Math.min(1, Math.max(0, item.volume));
    if (playing) media.play().catch(() => {});
    else media.pause();
  }, [item.kind, item.playbackRate, item.sourceTimeMs, item.volume, playing]);
  const style = visualEffectPreviewStyle(item.visualEffect, item.sourceTimeMs);
  if (item.canvasTransform) {
    const transform = item.canvasTransform;
    const directTransform = `translate(${Number(transform.x || 0) * 100}%, ${Number(transform.y || 0) * 100}%) rotate(${Number(transform.rotation || 0)}deg) scale(${Number(transform.scaleX || 1)}, ${Number(transform.scaleY || 1)})`;
    style.transform = `${style.transform || ''} ${directTransform}`.trim();
    style.opacity = Number.isFinite(Number(transform.opacity)) ? Number(transform.opacity) : (style.opacity ?? 1);
  }
  if (item.kind === 'image') return <img src={item.asset.url} alt={item.asset.name} style={style} />;
  if (item.kind === 'audio') return <audio ref={mediaRef} src={item.asset.url} preload="auto" />;
  return <video ref={mediaRef} src={item.asset.url} preload="auto" muted playsInline style={style} />;
}

export default function AICutPage() {
  const [projects, setProjects] = useState([]);
  const [projectId, setProjectId] = useState('');
  const [assets, setAssets] = useState([]);
  const [timeline, setTimeline] = useState({ revision: 0, tracks: [] });
  const [selectedAssetId, setSelectedAssetId] = useState('');
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
  const [previewResource, setPreviewResource] = useState(null);
  const [inspectorTab, setInspectorTab] = useState('基础');
  const [inspectorDrawerOpen, setInspectorDrawerOpen] = useState(false);
  const [agentDrawerOpen, setAgentDrawerOpen] = useState(false);
  const [assetSearch, setAssetSearch] = useState('');
  const [libraryFilter, setLibraryFilter] = useState('全部');
  const [favoriteFeatures, setFavoriteFeatures] = useState(() => {
    try { return JSON.parse(window.localStorage.getItem('aicut.favoriteFeatures') || '[]'); } catch { return []; }
  });
  const [recentFeatures, setRecentFeatures] = useState(() => {
    try { return JSON.parse(window.localStorage.getItem('aicut.recentFeatures') || '[]'); } catch { return []; }
  });
  const [toolWorkbenchOpen, setToolWorkbenchOpen] = useState('');
  const [editorLayout, setEditorLayout] = useState(() => {
    try {
      const stored = JSON.parse(window.localStorage.getItem('aicut.editorLayout.v3') || '{}');
      return normalizeEditorLayout(stored, { width: window.innerWidth, height: window.innerHeight });
    } catch {
      return normalizeEditorLayout({}, { width: window.innerWidth, height: window.innerHeight });
    }
  });
  const mediaCollapsed = editorLayout.resourceCollapsed;
  const inspectorCollapsed = editorLayout.inspectorCollapsed;
  const timelineCollapsed = editorLayout.timelineCollapsed;
  const setMediaCollapsed = value => setEditorLayout(current => ({ ...current, resourceCollapsed: typeof value === 'function' ? value(current.resourceCollapsed) : Boolean(value) }));
  const setInspectorCollapsed = value => setEditorLayout(current => ({ ...current, inspectorCollapsed: typeof value === 'function' ? value(current.inspectorCollapsed) : Boolean(value) }));
  const setTimelineCollapsed = value => setEditorLayout(current => ({ ...current, timelineCollapsed: typeof value === 'function' ? value(current.timelineCollapsed) : Boolean(value) }));
  const [shortcutOpen, setShortcutOpen] = useState(false);
  const [agentModels, setAgentModels] = useState([]);
  const [agentDraft, setAgentDraft] = useState(null);
  const [activeAgentSkill, setActiveAgentSkill] = useState('');
  const [bladeMode, setBladeMode] = useState(false);
  const [magneticSnap, setMagneticSnap] = useState(true);
  const [linkedMove, setLinkedMove] = useState(true);
  const [canvasFit, setCanvasFit] = useState('fit');
  const [trackEditId, setTrackEditId] = useState('');
  const [trackNameDraft, setTrackNameDraft] = useState('');
  const [replaceOpen, setReplaceOpen] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [exportingFormat, setExportingFormat] = useState('');
  const [exportJob, setExportJob] = useState(null);
  const [mediaToolBusy, setMediaToolBusy] = useState('');
  const [replaceFrom, setReplaceFrom] = useState('');
  const [replaceTo, setReplaceTo] = useState('');
  const [bgmPreference, setBGMPreference] = useState('');
  const [bgmOptions, setBGMOptions] = useState({ align: true, fade: true, continue: false, fill: true });
  const [bgmReferenceTrackId, setBGMReferenceTrackId] = useState('');
  const [visualRequirement, setVisualRequirement] = useState('');
  const [visualFrames, setVisualFrames] = useState([]);
  const [visualBusy, setVisualBusy] = useState(false);
  const [visualEffectDraft, setVisualEffectDraft] = useState(() => normalizeVisualEffect());
  const [subtitleStyle, setSubtitleStyle] = useState('classic-outline');
  const [subtitleRhythm, setSubtitleRhythm] = useState('按句/短语');
  const [subtitleAnimation, setSubtitleAnimation] = useState('none');
  const [manualSubtitle, setManualSubtitle] = useState({ text: '', startMs: 0, durationMs: 2000 });
  const [textStyleId, setTextStyleId] = useState('plain');
  const [selectedClipKey, setSelectedClipKey] = useState('');
  const [previewTimeMs, setPreviewTimeMs] = useState(0);
  const [previewPlaying, setPreviewPlaying] = useState(false);
  const [safeGuide, setSafeGuide] = useState(false);
  const [timelineZoom, setTimelineZoom] = useState(1);
  const [dragPreview, setDragPreview] = useState(null);
  const [bgmOpen, setBGMOpen] = useState(false);
  const [fillerOpen, setFillerOpen] = useState(false);
  const [oralOpen, setOralOpen] = useState(false);
  const [fillerModelId, setFillerModelId] = useState('');
  const [trimmingFillers, setTrimmingFillers] = useState(false);
  const fillerAbortRef = useRef(null);
  const [historyVersion, setHistoryVersion] = useState(0);
  const [serverHistory, setServerHistory] = useState({ canUndo: false, canRedo: false });
  const undoRef = useRef([]);
  const redoRef = useRef([]);
  const seenASRTerminalRef = useRef(new Set());
  const clipDragRef = useRef(null);
  const layoutDragRef = useRef(null);
  const editorShellRef = useRef(null);
  const previewStageRef = useRef(null);
  const downloadedExportRef = useRef('');

  const currentProject = projects.find(item => item.id === projectId);
  const selectedAsset = assets.find(item => item.id === selectedAssetId) || assets[0];
  const visibleAssets = useMemo(() => assets.filter(asset => {
    if (materialTab === '音频' && asset.kind !== 'audio') return false;
    if (materialTab === '视觉' && !['video', 'image'].includes(asset.kind)) return false;
    if (materialTab === '字幕' && asset.kind !== 'subtitle') return false;
    return !assetSearch.trim() || asset.name.toLowerCase().includes(assetSearch.trim().toLowerCase());
  }), [assets, assetSearch, materialTab]);
  const durationMs = useMemo(() => getAICutTimelineDuration(timeline), [timeline]);
  const previewFrame = useMemo(() => {
    const frame = buildAICutPreviewFrame(timeline, assets, previewTimeMs);
    const previewEffect = visualEffectForCatalogResource(previewResource);
    const selectedPreviewClipId = selectedClipKey.split('::')[1];
    if (!previewEffect || !selectedPreviewClipId) return frame;
    return { ...frame, visuals: frame.visuals.map(item => item.clipId === selectedPreviewClipId ? { ...item, visualEffect: previewEffect } : item) };
  }, [timeline, assets, previewTimeMs, previewResource, selectedClipKey]);
  const selectedClipEntry = useMemo(() => {
    const [trackId, clipId] = selectedClipKey.split('::');
    const track = timeline.tracks.find(item => item.id === trackId);
    const clip = track?.clips?.find(item => item.id === clipId);
    return track && clip ? { track, clip } : null;
  }, [timeline, selectedClipKey]);
  const capabilities = createAICutCapabilities({
    uploadAsset: () => { document.querySelector('.aicut-library-upload input[type=file]')?.click(); return false; },
    openAssets: () => { setMaterialTab('我的素材'); setToolWorkbenchOpen('项目素材'); },
    extractAudio: () => selectedAsset?.kind === 'video' ? extractAudioAsset(selectedAsset) : (message.warning('请先在媒体中选择视频素材'), false),
    openText: () => { setMaterialTab('文字稿'); setToolWorkbenchOpen('默认文本'); },
    openBatchText: () => subtitleClips(timeline).length ? setReplaceOpen(true) : (message.warning('请先添加字幕'), false),
    openVisualEffects: () => {
      if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return (message.warning('请先选择画面片段'), false);
      if (selectedClipEntry.track.locked) return (message.warning('轨道已锁定'), false);
      setVisualEffectDraft(normalizeVisualEffect(selectedClipEntry.clip.visualEffect));
      setMaterialTab('画面特效');
      setToolWorkbenchOpen('画面特效');
    },
    openTransitions: () => { setMaterialTab('转场'); setToolWorkbenchOpen('基础转场'); },
    openTranscription: () => selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? setASROpen(true) : (message.warning('请先选择音频或视频素材'), false),
    openCaptionCorrection: agentModels.length ? () => {
      if (!subtitleClips(timeline).length) return (message.warning('请先添加或识别字幕'), false);
      setCaptionScript(timeline.script || captionScript);
      setCaptionOpen(true);
    } : undefined,
    openSubtitleUpload: () => { setMaterialTab('文字稿'); setToolWorkbenchOpen('字幕文件上传'); },
    downloadSubtitles: () => exportSRT(),
    openSpeechCleanup: agentModels.length ? () => subtitleClips(timeline).length ? setOralOpen(true) : (message.warning('请先添加或识别字幕'), false) : undefined,
    smartSplit: () => selectedAsset?.kind === 'video' ? smartSplitAsset(selectedAsset) : (message.warning('请先在媒体中选择视频素材'), false),
  });
  const visibleEditorTools = editorTools.filter(([label]) => capabilities.some(item => item.category === label));
  const displayTracks = useMemo(() => ['subtitle', 'video', 'audio'].flatMap(kind => {
    const matches = timeline.tracks.filter(track => track.kind === kind);
    return matches.length ? matches : [{ id: `empty-${kind}`, kind, name: `${kind === 'subtitle' ? '字幕' : kind === 'video' ? '视频' : '音频'}1`, clips: [], placeholder: true }];
  }), [timeline]);

  useEffect(() => {
    let active = true;
    listAICutProjects().then(items => {
      if (!active) return;
      setProjects(items);
      setProjectId(items[0]?.id || '');
    }).catch(error => message.error(error?.message || '剪辑项目读取失败')).finally(() => active && setLoading(false));
    return () => { active = false; };
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
    Promise.all([loadAICutTimeline(projectId), listAICutAssets(projectId), getAICutTimelineHistory(projectId)]).then(([nextTimeline, nextAssets, nextHistory]) => {
      if (!active) return;
      setTimeline(nextTimeline);
      setSaveState('saved');
      undoRef.current = [];
      redoRef.current = [];
      setHistoryVersion(value => value + 1);
      setAssets(nextAssets);
      setServerHistory(nextHistory);
      setSelectedAssetId(nextAssets[0]?.id || '');
    }).catch(error => message.error(error?.message || 'AI-CUT 项目加载失败')).finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [projectId]);

  useEffect(() => {
    setPreviewPlaying(false);
    setPreviewTimeMs(0);
  }, [projectId]);

  useEffect(() => {
    setPreviewTimeMs(current => seekAICutPreviewTime(current, durationMs));
  }, [durationMs]);

  useEffect(() => {
    if (!previewPlaying || durationMs <= 0) return undefined;
    const originTime = previewTimeMs;
    const originClock = performance.now();
    let frameID = 0;
    const tick = now => {
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
    if (!projectId || !exportJob?.id || !['queued', 'running', 'cancel_requested'].includes(exportJob.state)) return undefined;
    let active = true;
    const timer = window.setInterval(async () => {
      try {
        const current = await getAICutExportJob(projectId, exportJob.id);
        if (active && current) setExportJob(current);
      } catch (error) {
        if (active) {
          setExportingFormat('');
          message.error(error?.message || '导出任务状态读取失败');
        }
      }
    }, 800);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [projectId, exportJob?.id, exportJob?.state]);

  useEffect(() => {
    if (!projectId || !exportJob?.id) return;
    if (exportJob.state === 'succeeded' && downloadedExportRef.current !== exportJob.id) {
      downloadedExportRef.current = exportJob.id;
      downloadAICutExportJob(projectId, exportJob.id).then(blob => {
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = `${currentProject?.title || 'AI-CUT成片'}.${exportJob.format}`;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        URL.revokeObjectURL(url);
        message.success(`${exportJob.format.toUpperCase()} 已渲染并开始下载`);
      }).catch(error => message.error(error?.message || '导出文件下载失败')).finally(() => setExportingFormat(''));
    } else if (exportJob.state === 'failed') {
      setExportingFormat('');
      message.error(exportJob.error || '成片导出失败');
    } else if (exportJob.state === 'canceled') {
      setExportingFormat('');
      message.info('成片导出已取消');
    }
  }, [projectId, exportJob?.id, exportJob?.state]);

  useEffect(() => {
    const settings = timeline.tracks.find(track => track.kind === 'subtitle')?.settings;
    if (!settings) return;
    setSubtitleStyle(settings.subtitleStyleId || '黑底白字');
    setSubtitleRhythm(subtitleRhythmLabel(settings.subtitleRhythm));
    setSubtitleAnimation(settings.subtitleAnimation || '无动效');
  }, [timeline.revision, projectId]);

  useEffect(() => {
    if (!projectId) { setASRJobs([]); return; }
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
            setTimeline(await loadAICutTimeline(projectId));
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
    return () => { active = false; window.clearInterval(timer); };
  }, [projectId]);

  useEffect(() => {
    let active = true;
    listAvailableModels('asr').then(models => {
      if (!active) return;
      setASRModels(models);
      setASRModelId(current => current || models[0]?.id || '');
    }).catch(error => { if (active) message.error(error?.message || 'ASR 模型读取失败'); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    let active = true;
    listAvailableModels('text').then(models => {
      if (active) setAgentModels(models);
    }).catch(error => { if (active) message.error(error?.message || 'Agent 模型读取失败'); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    const handleResize = () => setEditorLayout(current => normalizeEditorLayout(current, { width: window.innerWidth, height: window.innerHeight }));
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
      const action = resolveAICutShortcut({
        key: event.key,
        code: event.code,
        ctrlKey: event.ctrlKey,
        metaKey: event.metaKey,
        shiftKey: event.shiftKey,
        targetTag: event.target?.tagName,
        targetEditable: event.target?.isContentEditable
      }, { hasProject: Boolean(projectId), hasSelection: Boolean(selectedClipEntry) });
      if (!action) return;
      event.preventDefault();
      if (action === 'close') {
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
      else if (action === 'canvas-view') setWorkspaceMode('canvas');
      else if (action === 'editor-view') setWorkspaceMode('editor');
      else if (action === 'new-project') setCreateOpen(true);
      else if (action === 'save') saveTimeline();
      else if (action === 'export') setExportOpen(true);
      else if (action === 'play') togglePreviewPlayback();
      else if (action === 'undo') undoTimeline();
      else if (action === 'redo') redoTimeline();
      else if (action === 'blade') setBladeMode(value => !value);
      else if (action === 'split' && selectedClipEntry) splitClip(selectedClipEntry.track.id, selectedClipEntry.clip);
      else if (action === 'delete' && selectedClipEntry) {
        removeClip(selectedClipEntry.track.id, selectedClipEntry.clip.id);
        setSelectedClipKey('');
      }
    }
    window.addEventListener('keydown', handleShortcut);
    return () => window.removeEventListener('keydown', handleShortcut);
  }, [projectId, selectedClipEntry, timeline, durationMs, previewPlaying, previewTimeMs, serverHistory]);

  async function createProject() {
    if (!newTitle.trim()) return message.warning('请输入项目名称');
    try {
      const project = await createAICutProject({ title: newTitle, aspectRatio: newRatio });
      setProjects(current => [project, ...current]);
      setProjectId(project.id);
      setNewTitle('');
      setCreateOpen(false);
      message.success('AI-CUT 项目已创建');
    } catch (error) {
      message.error(error?.message || '项目创建失败');
    }
  }

  async function upload(file) {
    if (!projectId) return Upload.LIST_IGNORE;
    setUploading(true);
    try {
      const asset = await uploadAICutAsset(projectId, file);
      setAssets(current => [...current, asset]);
      setSelectedAssetId(asset.id);
      message.success('素材已上传并保存');
    } catch (error) {
      message.error(error?.message || '素材上传失败');
    } finally {
      setUploading(false);
    }
    return Upload.LIST_IGNORE;
  }

  async function importSRT(file) {
    if (!file) return Upload.LIST_IGNORE;
    if (!projectId) {
      message.warning('请先创建剪辑项目');
      return Upload.LIST_IGNORE;
    }
    setImportingSRT(true);
    try {
      const result = await importAICutSRT(projectId, file, timeline.revision);
      if (!result) return Upload.LIST_IGNORE;
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
    commitTimeline(current => {
      const tracks = current.tracks.map(track => ({ ...track, clips: [...(track.clips || [])] }));
      let track = options.trackId ? tracks.find(item => item.id === options.trackId && item.kind === kind && !item.locked) : null;
      if (!track) track = tracks.find(item => item.kind === kind && !item.locked);
      if (!track) {
        track = { id: newID(`track-${kind}`), kind, name: trackLabels[kind], order: tracks.length, muted: false, locked: false, clips: [] };
        tracks.push(track);
      }
      const appendedStart = track.clips.reduce((max, clip) => Math.max(max, Number(clip.startMs) + Number(clip.durationMs)), 0);
      const startMs = Number.isFinite(Number(options.startMs)) ? Math.max(0, Number(options.startMs)) : appendedStart;
      track.clips.push({ id: newID('clip'), assetId: asset.id, kind: asset.kind, name: asset.name, startMs, durationMs: defaultDuration(asset), sourceStartMs: 0, sourceDurationMs: defaultDuration(asset), playbackRate: 1, text: '' });
      return { ...current, tracks };
    });
  }

  function dropAssetOnTrack(event, track) {
    event.preventDefault();
    const assetId = event.dataTransfer.getData('application/x-aicut-asset');
    const asset = assets.find(item => item.id === assetId);
    if (!asset) return message.warning('没有读取到要拖入的素材');
    if (track.locked) return message.warning('轨道已锁定');
    if (trackKindForAsset(asset) !== track.kind) return message.warning('素材类型与轨道不匹配');
    const scrollLeft = event.currentTarget.closest('.aicut-tracks-scroll')?.scrollLeft || 0;
    const startMs = timelineDropTime(event.clientX, event.currentTarget.getBoundingClientRect().left, scrollLeft, timelineZoom);
    addAssetToTimeline(asset, { trackId: track.placeholder ? '' : track.id, startMs });
    setSelectedAssetId(asset.id);
    message.success(`已把 ${asset.name} 放到 ${previewClock(startMs)}`);
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
    commitTimeline(current => ({ ...current, tracks: current.tracks.map(track => track.id !== trackId ? track : { ...track, clips: track.clips.map(clip => clip.id === clipId ? { ...clip, ...patch } : clip) }) }));
  }

  function applyVisualEffect() {
    if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return message.warning('请先选择画面片段');
    if (selectedClipEntry.track.locked) return message.warning('轨道已锁定');
    updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { visualEffect: normalizeVisualEffect(visualEffectDraft) });
    setToolWorkbenchOpen('');
    message.success('画面效果已应用，可保存工程或直接导出');
  }

  function removeClip(trackId, clipId) {
    commitTimeline(current => ({ ...current, tracks: current.tracks.map(track => track.id !== trackId ? track : { ...track, clips: track.clips.filter(clip => clip.id !== clipId) }) }));
  }

  function duplicateSelectedClip() {
    if (!selectedClipEntry || selectedClipEntry.track.locked) return;
    const { track, clip } = selectedClipEntry;
    const duplicate = { ...clip, id: newID('clip'), name: `${clip.name} 副本`, startMs: Number(clip.startMs) + Number(clip.durationMs || 1) };
    commitTimeline(current => ({ ...current, tracks: current.tracks.map(item => item.id === track.id ? { ...item, clips: [...(item.clips || []), duplicate] } : item) }));
    setSelectedClipKey(`${track.id}::${duplicate.id}`);
  }

  function splitClip(trackId, clip) {
    if (clip.durationMs < 2000) return message.warning('片段至少需要 2 秒才能分割');
    commitTimeline(current => splitTimelineClip(current, trackId, clip.id, () => newID('clip')));
  }

  function splitClipAtPointer(track, clip, event) {
    if (!bladeMode || track.locked) return false;
    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width)));
    const offsetMs = Math.round(Number(clip.durationMs) * ratio);
    try {
      commitTimeline(current => splitTimelineClipAt(current, track.id, clip.id, offsetMs, () => newID('clip')));
      setSelectedClipKey('');
    } catch (error) {
      message.warning(error?.message || '分割失败');
    }
    return true;
  }

  function trimClip(trackId, clipId, edge) {
    try {
      commitTimeline(current => trimTimelineClipEdge(current, trackId, clipId, edge, 100));
    } catch (error) {
      message.warning(error?.message || '裁剪失败');
    }
  }

  function changeClipSpeed(trackId, clipId, rate) {
    try {
      commitTimeline(current => setClipPlaybackRate(current, trackId, clipId, rate));
    } catch (error) {
      message.warning(error?.message || '变速失败');
    }
  }

  function moveClip(trackId, clipId, startMs) {
    try {
      commitTimeline(current => moveTimelineClip(current, trackId, clipId, startMs, { magnetic: magneticSnap, linked: linkedMove, thresholdMs: 150 }));
    } catch (error) {
      message.warning(error?.message || '移动片段失败');
    }
  }

  function togglePreviewPlayback() {
    if (durationMs <= 0) return message.warning('请先把素材加入时间线');
    if (!previewPlaying && previewTimeMs >= durationMs) setPreviewTimeMs(0);
    setPreviewPlaying(value => !value);
  }

  function seekPreview(value) {
    setPreviewPlaying(false);
    setPreviewTimeMs(seekAICutPreviewTime(value, durationMs));
  }

  function beginClipDrag(event, track, clip) {
    if (track.locked || event.button !== 0 || event.target.closest('button,input,.ant-select')) return;
    if (splitClipAtPointer(track, clip, event)) {
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    clipDragRef.current = { trackId: track.id, clipId: clip.id, originX: event.clientX, originStartMs: Number(clip.startMs) || 0 };
    setDragPreview({ key: `${track.id}::${clip.id}`, startMs: Number(clip.startMs) || 0 });
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function updateClipDrag(event) {
    const drag = clipDragRef.current;
    if (!drag) return;
    setDragPreview({ key: `${drag.trackId}::${drag.clipId}`, startMs: draggedTimelineStart(drag.originStartMs, event.clientX - drag.originX, timelineZoom) });
  }

  function finishClipDrag(event) {
    const drag = clipDragRef.current;
    if (!drag) return;
    const nextStart = draggedTimelineStart(drag.originStartMs, event.clientX - drag.originX, timelineZoom);
    clipDragRef.current = null;
    setDragPreview(null);
    if (nextStart !== drag.originStartMs) moveClip(drag.trackId, drag.clipId, nextStart);
  }

  function seekFromTrack(event) {
    if (event.target !== event.currentTarget) return;
    const rect = event.currentTarget.getBoundingClientRect();
    seekPreview(timelinePixelsToTime(event.clientX - rect.left, timelineZoom));
  }

  function markClip(trackId, clipId, status) {
    try {
      commitTimeline(current => markTimelineClip(current, trackId, clipId, status));
    } catch (error) {
      message.warning(error?.message || '标记失败');
    }
  }

  function persistSubtitleSettings(patch) {
    const subtitleTrack = timeline.tracks.find(track => track.kind === 'subtitle' && !track.locked);
    if (!subtitleTrack) return message.warning('请先新建字幕轨');
    commitTimeline(current => updateTimelineTrackSettings(current, subtitleTrack.id, patch));
  }

  function commitTimeline(update) {
    setTimeline(current => {
      const next = typeof update === 'function' ? update(current) : update;
      undoRef.current = [...undoRef.current.slice(-49), current];
      redoRef.current = [];
      setHistoryVersion(value => value + 1);
      return next;
    });
    setSaveState('dirty');
  }

  async function undoTimeline() {
    if (undoRef.current.length) {
      setTimeline(current => {
      const previous = undoRef.current.pop();
      if (!previous) return current;
      redoRef.current = [current, ...redoRef.current].slice(0, 50);
      setHistoryVersion(value => value + 1);
      return previous;
      });
	  setSaveState('dirty');
      return;
    }
    if (!projectId || !serverHistory.canUndo) return;
    try {
      const result = await undoAICutTimeline(projectId, timeline.revision);
      setTimeline(result.timeline);
      setSaveState('saved');
      setServerHistory(result.history);
      setHistoryVersion(value => value + 1);
    } catch (error) {
      message.error(error?.message || '撤销失败');
    }
  }

  async function redoTimeline() {
    if (redoRef.current.length) {
      setTimeline(current => {
      const next = redoRef.current.shift();
      if (!next) return current;
      undoRef.current = [...undoRef.current.slice(-49), current];
      setHistoryVersion(value => value + 1);
      return next;
      });
	  setSaveState('dirty');
      return;
    }
    if (!projectId || !serverHistory.canRedo) return;
    try {
      const result = await redoAICutTimeline(projectId, timeline.revision);
      setTimeline(result.timeline);
      setSaveState('saved');
      setServerHistory(result.history);
      setHistoryVersion(value => value + 1);
    } catch (error) {
      message.error(error?.message || '重做失败');
    }
  }

  function addTrack(kind) {
    commitTimeline(current => addTimelineTrack(current, kind, () => newID(`track-${kind}`)));
  }

  function toggleTrackFlag(trackId, field) {
    commitTimeline(current => toggleTimelineTrackFlag(current, trackId, field));
  }

  function removeTrack(track) {
    if (track.clips?.length) return message.warning('轨道中还有片段，不能删除');
    commitTimeline(current => removeTimelineTrack(current, track.id));
  }

  function openTrackRename(track) {
    if (!track || track.placeholder) return;
    setTrackEditId(track.id);
    setTrackNameDraft(track.name || '');
  }

  function confirmTrackRename() {
    try {
      commitTimeline(current => renameTimelineTrack(current, trackEditId, trackNameDraft));
      setTrackEditId('');
      setTrackNameDraft('');
    } catch (error) {
      message.warning(error?.message || '轨道重命名失败');
    }
  }

  function trackMenuItems(track) {
    if (!track || track.placeholder) return [];
    return [
      { key: 'rename', label: '重命名轨道', onClick: () => openTrackRename(track) },
      { key: 'locked', label: track.locked ? '解锁轨道' : '锁定轨道', onClick: () => toggleTrackFlag(track.id, 'locked') },
      { key: 'muted', label: track.muted ? '取消静音' : '静音轨道', onClick: () => toggleTrackFlag(track.id, 'muted') },
      { type: 'divider' },
      { key: 'delete', danger: true, label: '删除空轨道', onClick: () => removeTrack(track) }
    ];
  }

  async function saveTimeline() {
    if (!projectId) return;
    setSaving(true);
    setSaveState('saving');
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      setTimeline(saved);
      setSaveState('saved');
      undoRef.current = [];
      redoRef.current = [];
      setServerHistory(await getAICutTimelineHistory(projectId));
      setHistoryVersion(value => value + 1);
      setSaveState('saved');
      message.success('时间线已保存');
    } catch (error) {
      setSaveState('failed');
      message.error(error?.message || '时间线保存失败');
    } finally {
      setSaving(false);
    }
  }

  async function exportMedia(format) {
    if (!projectId || exportingFormat) return;
    setExportingFormat(format);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      setTimeline(saved);
      undoRef.current = [];
      redoRef.current = [];
      setServerHistory(await getAICutTimelineHistory(projectId));
      downloadedExportRef.current = '';
      setExportJob(await createAICutExportJob(projectId, format, saved.revision));
      message.success('导出任务已提交，可查看进度或取消');
    } catch (error) {
      message.error(error?.message || '成片导出失败');
      setExportingFormat('');
    }
  }

  async function cancelExport() {
    if (!projectId || !exportJob?.id || !['queued', 'running'].includes(exportJob.state)) return;
    try {
      setExportJob(await cancelAICutExportJob(projectId, exportJob.id));
    } catch (error) {
      message.error(error?.message || '取消导出失败');
    }
  }

  async function smartSplitAsset(asset) {
    if (!asset || asset.kind !== 'video' || mediaToolBusy) return false;
    if (!timeline.tracks.some(track => track.clips?.some(clip => clip.assetId === asset.id))) {
      message.warning('请先把这个视频加入时间线');
      return false;
    }
    setMediaToolBusy(`split:${asset.id}`);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      const result = await smartSplitAICutVideo(projectId, { assetId: asset.id, threshold: 0.4, minIntervalMs: 500, expectedRevision: saved.revision });
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
      setAssets(current => [...current.filter(item => item.id !== created.id), created]);
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

  async function correctCaptions() {
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
    if (!timeline.tracks.some(track => track.kind === 'subtitle' && !track.locked && track.clips?.some(clip => String(clip.text || '').trim()))) {
      return message.warning('时间线上没有可处理的字幕，或字幕轨已锁定');
    }
    const controller = new AbortController();
    fillerAbortRef.current = controller;
    setTrimmingFillers(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      setTimeline(saved);
      const result = await trimAICutFillers(projectId, {
        modelId: fillerModelId,
        expectedRevision: saved.revision,
        signal: controller.signal
      });
      setTimeline(result.timeline);
      undoRef.current = [];
      redoRef.current = [];
      setHistoryVersion(value => value + 1);
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
    if (!timeline.tracks.some(track => track.kind === 'subtitle' && !track.locked && track.clips?.some(clip => String(clip.text || '').trim()))) {
      return message.warning('时间线上没有可处理的字幕，或字幕轨已锁定');
    }
    const controller = new AbortController();
    fillerAbortRef.current = controller;
    setTrimmingFillers(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      setTimeline(saved);
      const result = await cleanAICutOralTimeline(projectId, { modelId: fillerModelId, maxPauseMs: 1000, expectedRevision: saved.revision, signal: controller.signal });
      setTimeline(result.timeline);
      undoRef.current = [];
      redoRef.current = [];
      setHistoryVersion(value => value + 1);
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
    if (!selectedAsset || !['audio', 'video'].includes(selectedAsset.kind)) return message.warning('请选择音频或视频素材');
    if (!asrModelId) return message.warning('请先到 API 配置启用语音识别模型');
    if (asrJobs.some(job => job.assetId === selectedAsset.id && ['queued', 'running', 'cancel_requested'].includes(job.state))) return message.warning('该素材正在识别，请等待完成或先取消');
    setASRBusy(true);
    try {
      const saved = await saveAICutTimeline(projectId, timeline);
      setTimeline(saved);
      const job = await createAICutASRJob(projectId, { assetId: selectedAsset.id, modelId: asrModelId, language: asrLanguage, expectedRevision: saved.revision });
      setASRJobs(current => [job, ...current.filter(item => item.id !== job.id)]);
      setASROpen(false);
      message.success('语音识别任务已提交');
    } catch (error) {
      message.error(error?.message || '语音识别任务提交失败');
    } finally { setASRBusy(false); }
  }

  async function cancelASR(job) {
    try {
      const updated = await cancelAICutASRJob(projectId, job.id);
      setASRJobs(current => current.map(item => item.id === updated.id ? updated : item));
      message.info(updated.state === 'canceled' ? '排队任务已取消' : '正在取消语音识别');
    } catch (error) { message.error(error?.message || '取消失败'); }
  }

  async function refreshTimelineFromAgent() {
    const [nextTimeline, nextHistory] = await Promise.all([loadAICutTimeline(projectId), getAICutTimelineHistory(projectId)]);
    setTimeline(nextTimeline);
    setServerHistory(nextHistory);
    undoRef.current = [];
    redoRef.current = [];
    setHistoryVersion(value => value + 1);
  }

  async function importScript(file) {
    if (!file) return Upload.LIST_IGNORE;
    try {
      const text = await file.text();
      if (!text.trim()) return message.warning('剧本文件是空的');
      if ([...text].length > 200000) return message.error('剧本不能超过 20 万字');
      commitTimeline(current => ({ ...current, script: text }));
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
    const url = URL.createObjectURL(new Blob([`\uFEFF${body}`], { type: 'application/x-subrip;charset=utf-8' }));
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
    commitTimeline(current => ({ ...current, tracks: current.tracks.map(track => track.kind !== 'subtitle' || track.locked ? track : { ...track, clips: (track.clips || []).map(clip => {
      const text = String(clip.text || '');
      if (!text.includes(replaceFrom)) return clip;
      count += text.split(replaceFrom).length - 1;
      return { ...clip, text: text.split(replaceFrom).join(replaceTo) };
    }) }) }));
    setReplaceOpen(false);
    message.success(count ? `已替换 ${count} 处` : '没有找到匹配文字');
  }

  function addManualSubtitle() {
    const text = manualSubtitle.text.trim();
    if (!text) return message.warning('请输入字幕文字');
    if (manualSubtitle.startMs < 0 || manualSubtitle.durationMs <= 0) return message.warning('字幕时间必须大于 0');
    commitTimeline(current => {
      const tracks = current.tracks.map(track => ({ ...track, clips: [...(track.clips || [])] }));
      let track = tracks.find(item => item.kind === 'subtitle' && !item.locked);
      if (!track) {
        track = { id: newID('track-subtitle'), kind: 'subtitle', name: '字幕1', order: tracks.length, muted: false, locked: false, clips: [] };
        tracks.push(track);
      }
      track.clips.push({ id: newID('subtitle'), kind: 'subtitle', name: text.slice(0, 16), startMs: Number(manualSubtitle.startMs), durationMs: Number(manualSubtitle.durationMs), sourceStartMs: 0, sourceDurationMs: Number(manualSubtitle.durationMs), text });
      track.clips.sort((a, b) => Number(a.startMs) - Number(b.startMs));
      return { ...current, tracks };
    });
    setManualSubtitle(current => ({ ...current, text: '', startMs: current.startMs + current.durationMs }));
    message.success('字幕已加入时间线');
  }

  function addTextOverlay() {
    const text = manualSubtitle.text.trim();
    if (!text) return message.warning('请输入文字内容');
    if (manualSubtitle.startMs < 0 || manualSubtitle.durationMs <= 0) return message.warning('文字时间必须大于 0');
    commitTimeline(current => {
      const tracks = current.tracks.map(track => ({ ...track, settings: { ...(track.settings || {}) }, clips: [...(track.clips || [])] }));
      const style = textStyleCatalog.find(item => item.id === textStyleId) || textStyleCatalog[0];
      const track = { id: newID('track-text'), kind: 'subtitle', name: style.label, order: tracks.length, muted: false, locked: false, settings: { subtitleStyleId: `text-${style.id}`, subtitleAnimation: 'none', previewVisible: true }, clips: [] };
      track.clips.push({ id: newID('text'), kind: 'subtitle', name: text.slice(0, 16), startMs: Number(manualSubtitle.startMs), durationMs: Number(manualSubtitle.durationMs), sourceStartMs: 0, sourceDurationMs: Number(manualSubtitle.durationMs), text });
      tracks.push(track);
      return { ...current, tracks };
    });
    setManualSubtitle(current => ({ ...current, text: '', startMs: current.startMs + current.durationMs }));
    message.success('文字已加入时间线');
  }

  function previewCatalogResource(resource) {
    setPreviewResource(resource);
    if (['filter', 'effect'].includes(resource.category) && !selectedClipEntry) message.info('先选择一个画面片段，再预览该资源');
  }

  function applyCatalogResource(resource) {
    setRecentFeatures(current => recordRecentAICutFeature(current, resource.id));
    if (resource.category === 'effect') {
      const result = addEffectLayer(timeline, { resourceId: resource.id, name: resource.labelZh, startMs: previewTimeMs, durationMs: 2000, parameters: { intensity: .5, speed: 1 } }, () => newID('effect'));
      if (!result.ok) return message.warning(result.reasonZh);
      commitTimeline(result.timeline);
      setPreviewResource(null);
      return message.success(`已把${resource.labelZh}加入特效轨`);
    }
    if (resource.category === 'filter') {
      if (!selectedClipEntry || !['video', 'image'].includes(selectedClipEntry.clip.kind)) return message.warning('请先选择视频或图片片段');
      if (selectedClipEntry.track.locked) return message.warning('轨道已锁定');
      const visualEffect = visualEffectForCatalogResource(resource);
      updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { visualEffect, catalogResourceId: resource.id });
      setPreviewResource(null);
      return message.success(`已应用${resource.labelZh}`);
    }
    if (resource.category === 'transition') {
      if (!selectedClipEntry || selectedClipEntry.track.kind !== 'video') return message.warning('请先选择同一视频轨中的后一个片段');
      const result = setClipTransition(timeline, selectedClipEntry.track.id, selectedClipEntry.clip.id, resource.id, 500);
      if (!result.ok) return message.warning(result.reasonZh);
      commitTimeline(result.timeline);
      return message.success(`已添加${resource.labelZh}`);
    }
    if (resource.category === 'text') {
      commitTimeline(current => {
        const tracks = current.tracks.map(track => ({ ...track, clips: [...(track.clips || [])] }));
        const track = { id: newID('track-text'), kind: 'subtitle', name: resource.labelZh, order: tracks.length, muted: false, locked: false, settings: { textResourceId: resource.id, previewVisible: true }, clips: [{ id: newID('text'), kind: 'subtitle', name: resource.labelZh, startMs: previewTimeMs, durationMs: 3000, sourceStartMs: 0, sourceDurationMs: 3000, text: '双击编辑文字', textResourceId: resource.id }] };
        tracks.push(track);
        return { ...current, tracks };
      });
      return message.success(`${resource.labelZh}已加入时间线`);
    }
    if (resource.category === 'template') {
      commitTimeline(current => ({ ...current, editorTemplateId: resource.id }));
      return message.success(`已应用${resource.labelZh}工程模板`);
    }
  }

  function applyTransition(kind) {
    const [trackId, clipId] = selectedClipKey.split('::');
    return applyTransitionTo(trackId, clipId, kind);
  }

  function applyTransitionTo(trackId, clipId, kind) {
    const track = timeline.tracks.find(item => item.id === trackId);
    const clipIndex = track?.clips?.findIndex(item => item.id === clipId) ?? -1;
    if (!transitionCatalog.some(item => item.id === kind)) return message.error('该转场尚未支持');
    if (!track || track.kind !== 'video' || clipIndex < 1) return message.warning('转场需要放在同一视频轨的两个相邻片段之间');
    if (track.locked) return message.warning('视频轨已锁定');
    const previous = track.clips[clipIndex - 1];
    const current = track.clips[clipIndex];
    if (Number(previous.startMs) + Number(previous.durationMs) !== Number(current.startMs)) return message.warning('两个片段之间有空隙，无法添加转场');
    commitTimeline(current => ({ ...current, tracks: current.tracks.map(item => item.id !== trackId ? item : { ...item, clips: item.clips.map(clip => clip.id === clipId ? { ...clip, transition: kind, transitionMs: 500 } : clip) }) }));
    message.success(`${transitionCatalog.find(item => item.id === kind)?.label || '转场'}已添加，时长 0.5 秒`);
  }

  function startBGMAnalysis() {
    const audioAssets = assets.filter(asset => asset.kind === 'audio');
    if (!audioAssets.length) return message.warning('请先上传至少一条音频素材');
    setBGMOpen(true);
  }

  function applyBGMPlan(asset) {
    try {
      commitTimeline(current => planTimelineBGM(current, asset, { ...bgmOptions, referenceTrackId: bgmReferenceTrackId }, () => newID('bgm')));
      setBGMOpen(false);
      message.success('背景音乐已按参考画面写入 BGM 轨道');
    } catch (error) {
      message.warning(error?.message || '背景音乐规划失败');
    }
  }

  async function startVisualUnderstanding() {
    const visual = selectedAsset && ['video', 'image'].includes(selectedAsset.kind) ? selectedAsset : assets.find(asset => ['video', 'image'].includes(asset.kind));
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

  function renderAssetList() {
    return <div className="aicut-asset-list">{visibleAssets.length ? visibleAssets.map(asset => <article key={asset.id} role="button" tabIndex={0} draggable title="拖到下方同类型轨道，或点加入时间线" className={selectedAsset?.id === asset.id ? 'active' : ''} onDragStart={event => { event.dataTransfer.effectAllowed = 'copy'; event.dataTransfer.setData('application/x-aicut-asset', asset.id); }} onClick={() => setSelectedAssetId(asset.id)} onKeyDown={event => { if (event.key === 'Enter' || event.key === ' ') setSelectedAssetId(asset.id); }}><span>{kindLabels[asset.kind] || asset.kind}</span><strong title={asset.name}>{asset.name}</strong><small>{Math.max(1, Math.round(asset.byteSize / 1024))} KB</small><div className="aicut-asset-actions">{['audio', 'video'].includes(asset.kind) ? <Button size="small" onClick={event => { event.stopPropagation(); setSelectedAssetId(asset.id); setASROpen(true); }}>识别字幕</Button> : null}{asset.kind === 'video' ? <><Button size="small" loading={mediaToolBusy === `split:${asset.id}`} onClick={event => { event.stopPropagation(); smartSplitAsset(asset); }}>智能切镜</Button><Button size="small" loading={mediaToolBusy === `audio:${asset.id}`} onClick={event => { event.stopPropagation(); extractAudioAsset(asset); }}>分离音频</Button></> : null}<Button size="small" icon={<Plus size={13} />} onClick={event => { event.stopPropagation(); addAssetToTimeline(asset); }}>加入时间线</Button></div></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="素材池是空的。点上传导入视频、图片或音频。" />}</div>;
  }

  function renderMaterialBody() {
    if (materialTab === '画面特效') return <div className="aicut-tool-pane aicut-visual-effect-pane"><p>下列效果会写入工程，并同时用于播放器预览和服务器导出。</p><strong>{selectedClipEntry?.clip?.name || '请先选择画面片段'}</strong><div className="aicut-choice-grid">{visualEffectCatalog.map(item => <button type="button" key={item.id} className={visualEffectDraft.preset === item.id ? 'active' : ''} onClick={() => setVisualEffectDraft(visualEffectPreset(item.id))}>{item.label}</button>)}</div><div className="aicut-effect-fields"><label>亮度<InputNumber min={-1} max={1} step={.05} value={visualEffectDraft.brightness} onChange={value => setVisualEffectDraft(current => ({ ...current, preset: 'custom', brightness: Number(value ?? 0) }))} /></label><label>对比度<InputNumber min={.5} max={2} step={.05} value={visualEffectDraft.contrast} onChange={value => setVisualEffectDraft(current => ({ ...current, preset: 'custom', contrast: Number(value ?? 1) }))} /></label><label>饱和度<InputNumber min={0} max={3} step={.05} value={visualEffectDraft.saturation} onChange={value => setVisualEffectDraft(current => ({ ...current, preset: 'custom', saturation: Number(value ?? 1) }))} /></label><label>模糊<InputNumber min={0} max={20} step={.5} value={visualEffectDraft.blur} onChange={value => setVisualEffectDraft(current => ({ ...current, preset: 'custom', blur: Number(value ?? 0) }))} addonAfter="px" /></label><label>强度<InputNumber min={0} max={1} step={.05} value={visualEffectDraft.intensity ?? .5} onChange={value => setVisualEffectDraft(current => ({ ...current, intensity: Number(value ?? .5) }))} /></label><label>速度<InputNumber min={.1} max={5} step={.1} value={visualEffectDraft.speed ?? 1} onChange={value => setVisualEffectDraft(current => ({ ...current, speed: Number(value ?? 1) }))} /></label><Checkbox checked={visualEffectDraft.grayscale} onChange={event => setVisualEffectDraft(current => ({ ...current, preset: 'custom', grayscale: event.target.checked }))}>黑白</Checkbox></div><div className="aicut-tool-actions"><Button onClick={() => setToolWorkbenchOpen('')}>取消</Button><Button type="primary" disabled={!selectedClipEntry} onClick={applyVisualEffect}>应用到片段</Button></div></div>;
    if (materialTab === '转场') return <div className="aicut-tool-pane"><p>选择同一视频轨中相接的第二个片段，再点击或拖入转场。两个片段之间有空隙时会拒绝添加。</p><div className="aicut-transition-grid">{transitionCatalog.map(item => <button type="button" key={item.id} draggable title={item.description} onDragStart={event => event.dataTransfer.setData('text/aicut-transition', item.id)} onClick={() => applyTransition(item.id)}><strong>{item.label}</strong><small>{item.description}</small></button>)}</div></div>;
    if (materialTab === '文本') return <div className="aicut-tool-pane"><p>添加独立文字轨，可在时间线中移动、复制、分割和删除。</p><div className="aicut-choice-grid">{textStyleCatalog.map(item => <button type="button" key={item.id} className={textStyleId === item.id ? 'active' : ''} onClick={() => setTextStyleId(item.id)}>{item.label}</button>)}</div><div className="aicut-manual-subtitle"><Input value={manualSubtitle.text} onChange={event => setManualSubtitle(value => ({ ...value, text: event.target.value }))} placeholder="输入文字内容" /><InputNumber min={0} value={manualSubtitle.startMs} onChange={value => setManualSubtitle(current => ({ ...current, startMs: Number(value || 0) }))} addonAfter="ms" /><InputNumber min={1} value={manualSubtitle.durationMs} onChange={value => setManualSubtitle(current => ({ ...current, durationMs: Number(value || 1) }))} addonAfter="ms" /><Button type="primary" onClick={addTextOverlay}>添加文字</Button></div></div>;
    if (materialTab === '文字稿') return <div className="aicut-tool-pane aicut-transcript-pane"><div className="aicut-tool-actions"><Upload showUploadList={false} beforeUpload={importSRT} accept=".srt"><Button loading={importingSRT}>导入 SRT</Button></Upload><Button onClick={exportSRT} disabled={!subtitleClips(timeline).length}>导出 SRT</Button><Button onClick={() => setReplaceOpen(true)} disabled={!subtitleClips(timeline).length}>替换</Button><Button onClick={() => setFillerOpen(true)} disabled={!subtitleClips(timeline).length}>去水词</Button><Button onClick={() => { setCaptionScript(timeline.script || captionScript); setCaptionOpen(true); }} disabled={!subtitleClips(timeline).length}>字幕纠错</Button></div><div className="aicut-tool-actions"><Upload showUploadList={false} beforeUpload={importScript} accept=".txt,.md,.text"><Button>导入剧本</Button></Upload><Tag>{timeline.tracks.filter(track => track.kind === 'video').length} 条视频轨</Tag><Tag>{timeline.tracks.filter(track => track.kind === 'audio').length} 条音频轨</Tag></div><Input.TextArea value={timeline.script || ''} onChange={event => commitTimeline(current => ({ ...current, script: event.target.value }))} rows={8} placeholder="导入剧本或在这里粘贴文字稿；保存工程后持久化" /><small>逐轨转写请在“我的素材”选中视频或音频后点“识别字幕”。</small></div>;
    if (materialTab === '音频') return <div className="aicut-tool-pane"><label>已上传曲库<Select value={selectedAsset?.kind === 'audio' ? selectedAsset.id : undefined} onChange={setSelectedAssetId} placeholder="选择音频素材" options={assets.filter(asset => asset.kind === 'audio').map(asset => ({ value: asset.id, label: asset.name }))} /></label><p>线上版本使用已上传的项目音频，不读取用户电脑文件夹，也不依赖剪映草稿。</p><div className="aicut-check-grid"><Checkbox checked={bgmOptions.align} onChange={event => setBGMOptions(value => ({ ...value, align: event.target.checked }))}>对齐片段</Checkbox><Checkbox checked={bgmOptions.fade} onChange={event => setBGMOptions(value => ({ ...value, fade: event.target.checked }))}>渐入渐出</Checkbox><Checkbox checked={bgmOptions.continue} onChange={event => setBGMOptions(value => ({ ...value, continue: event.target.checked }))}>续接 BGM</Checkbox><Checkbox checked={bgmOptions.fill} onChange={event => setBGMOptions(value => ({ ...value, fill: event.target.checked }))}>铺满轨道</Checkbox></div><label>参考轨道<Select value={bgmReferenceTrackId || undefined} onChange={setBGMReferenceTrackId} placeholder="默认使用第一条有画面的轨道" options={timeline.tracks.filter(track => track.kind === 'video').map(track => ({ value: track.id, label: track.name }))} /></label><label>BGM 偏好<Input.TextArea value={bgmPreference} onChange={event => setBGMPreference(event.target.value)} rows={4} placeholder="例如：虐恋用低沉钢琴，反转处换鼓点，平淡对白留白" /></label><Button type="primary" onClick={startBGMAnalysis}>开始分析</Button></div>;
    if (materialTab === '视觉') return <div className="aicut-tool-pane"><p>先从视频真实提取代表帧。当前文字模型不能读取图片，因此不会伪造“AI 已理解”；接入视觉模型后，这些帧会用于生成保留、复核、删除建议。</p><label>素材<Select value={selectedAsset && ['video', 'image'].includes(selectedAsset.kind) ? selectedAsset.id : undefined} onChange={value => { setSelectedAssetId(value); setVisualFrames([]); }} placeholder="选择视觉素材" options={assets.filter(asset => ['video', 'image'].includes(asset.kind)).map(asset => ({ value: asset.id, label: asset.name }))} /></label><label>额外要求<Input.TextArea value={visualRequirement} onChange={event => setVisualRequirement(event.target.value)} rows={4} placeholder="例如：走路空镜一律建议删除，回头和停步保留" /></label><Button type="primary" loading={visualBusy} onClick={startVisualUnderstanding}>提取代表帧</Button>{visualFrames.length ? <div className="aicut-visual-frames">{visualFrames.map(frame => <figure key={`${frame.timeMs}-${frame.url}`}><img src={frame.url} alt={`${Number(frame.timeMs) / 1000} 秒画面`} /><figcaption>{(Number(frame.timeMs) / 1000).toFixed(2)}s</figcaption></figure>)}</div> : null}</div>;
    if (materialTab === '字幕') return <div className="aicut-tool-pane"><div className="aicut-tool-actions"><Button onClick={() => addTrack('subtitle')}>新建字幕轨</Button><Button onClick={() => assets.some(asset => ['audio', 'video'].includes(asset.kind)) ? setASROpen(true) : message.warning('请先导入视频或音频')}>识别字幕</Button><Upload showUploadList={false} beforeUpload={importSRT} accept=".srt"><Button loading={importingSRT}>导入字幕</Button></Upload><Checkbox checked={timeline.tracks.find(track => track.kind === 'subtitle')?.settings?.previewVisible !== false} onChange={event => persistSubtitleSettings({ previewVisible: event.target.checked })}>预览中显示</Checkbox></div><strong>样式外观</strong><div className="aicut-choice-grid">{subtitleStyleCatalog.map(item => <button type="button" key={item.id} className={subtitleStyle === item.id ? 'active' : ''} onClick={() => { setSubtitleStyle(item.id); persistSubtitleSettings({ subtitleStyleId: item.id }); }}>{item.label}</button>)}</div><strong>显示节奏</strong><div className="aicut-tool-actions">{['按句/短语', '逐词高亮'].map(item => <Button key={item} type={subtitleRhythm === item ? 'primary' : 'default'} onClick={() => { setSubtitleRhythm(item); persistSubtitleSettings({ subtitleRhythm: normalizeSubtitleRhythm(item) }); }}>{item}</Button>)}</div><strong>字幕动效</strong><Select value={subtitleAnimationCatalog.some(item => item.id === subtitleAnimation) ? subtitleAnimation : 'none'} onChange={value => { setSubtitleAnimation(value); persistSubtitleSettings({ subtitleAnimation: value }); }} options={subtitleAnimationCatalog.map(item => ({ value: item.id, label: item.label }))} /><div className="aicut-manual-subtitle"><Input value={manualSubtitle.text} onChange={event => setManualSubtitle(value => ({ ...value, text: event.target.value }))} placeholder="输入字幕文字" /><InputNumber min={0} value={manualSubtitle.startMs} onChange={value => setManualSubtitle(current => ({ ...current, startMs: Number(value || 0) }))} addonAfter="ms" /><InputNumber min={1} value={manualSubtitle.durationMs} onChange={value => setManualSubtitle(current => ({ ...current, durationMs: Number(value || 1) }))} addonAfter="ms" /><Button type="primary" onClick={addManualSubtitle}>添加字幕</Button></div></div>;
    if (materialTab === 'SKILLS') return <div className="aicut-tool-pane aicut-skills-pane"><Input value={assetSearch} onChange={event => setAssetSearch(event.target.value)} placeholder="搜索 skills" />{[
      { title: '短剧AI剪辑', detail: '读取已导入剧本，只对缺字幕片段补转写，对照剧本纠错；智能切镜后以红/蓝/绿标记删除、复核、保留建议。保持素材原顺序，不自动换镜或删片。', prompt: '按短剧AI剪辑流程处理当前工程：读取剧本、补转写、纠错、切镜并给出红蓝绿保留建议' },
      { title: '技能创作器', detail: '把一段固定剪辑流程整理成可复用技能。', prompt: '把我接下来描述的剪辑流程制作成可复用技能' }
    ].filter(skill => !assetSearch.trim() || `${skill.title}${skill.detail}`.toLowerCase().includes(assetSearch.trim().toLowerCase())).map(skill => <button type="button" key={skill.title} onClick={() => { setActiveAgentSkill(skill.title); setAgentDraft({ text: skill.prompt, nonce: Date.now() }); }}><strong>{skill.title}</strong><span>{skill.detail}</span></button>)}</div>;
    return <><div className="aicut-media-tools"><Input value={assetSearch} onChange={event => setAssetSearch(event.target.value)} placeholder="搜索项目素材" /><Upload multiple showUploadList={false} beforeUpload={upload} accept="image/*,video/*,audio/*"><Button loading={uploading} icon={<UploadIcon size={15} />}>上传素材</Button></Upload></div>{renderAssetList()}</>;
  }

  function chooseToolCategory(category) {
    setToolCategory(category);
    setLibraryFilter('全部');
    const tab = { 媒体: '我的素材', 音频: '音频', 文本: '文本', 转场: '转场', 字幕: '字幕' }[category];
    if (tab) setMaterialTab(tab);
  }

  function openAgentWithPrompt(prompt, skill = '') {
    setActiveAgentSkill(skill);
    setAgentDraft({ text: prompt, nonce: Date.now() });
    setAgentDrawerOpen(true);
  }

  function renderToolLibrary() {
    const items = filterAICutCapabilities(capabilities, { category: toolCategory, filter: libraryFilter, favorites: favoriteFeatures, recent: recentFeatures, search: assetSearch });
    return <div className="aicut-reference-library"><div className="aicut-library-tools"><Input value={assetSearch} onChange={event => setAssetSearch(event.target.value)} placeholder="搜索功能或素材" /><Upload className="aicut-library-upload" multiple showUploadList={false} beforeUpload={upload} accept="image/*,video/*,audio/*"><Button loading={uploading} icon={<Plus size={14} />}>添加</Button></Upload></div><nav className="aicut-library-subtabs">{['全部', '收藏', '最近使用'].map(item => <button type="button" key={item} className={libraryFilter === item ? 'active' : ''} onClick={() => setLibraryFilter(item)}>{item}</button>)}</nav><div className="aicut-feature-list">{items.length ? items.map(item => {
      const Icon = editorTools.find(([label]) => label === item.category)?.[1] || Sparkles;
      const run = async () => {
        const result = await item.action();
        if (result !== false) setRecentFeatures(current => recordRecentAICutFeature(current, item.id));
      };
      return <div className="aicut-feature-card-wrap" key={item.id}><button type="button" className="aicut-feature-card active ready" onClick={() => run().catch(error => message.error(error?.message || `${item.title}执行失败`))}><span className="aicut-feature-icon"><Icon size={17} /></span><span><strong>{item.title}</strong><small>{item.detail}</small></span><em className={`source-${item.source}`}>{toolSourceLabels[item.source]}</em></button><button type="button" className={`aicut-feature-favorite${favoriteFeatures.includes(item.id) ? ' active' : ''}`} aria-label={`${favoriteFeatures.includes(item.id) ? '取消收藏' : '收藏'}${item.title}`} onClick={() => setFavoriteFeatures(current => toggleFavoriteAICutFeature(current, item.id))}><Star size={14} fill={favoriteFeatures.includes(item.id) ? 'currentColor' : 'none'} /></button></div>;
    }) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={libraryFilter === '收藏' ? '还没有收藏功能' : libraryFilter === '最近使用' ? '还没有使用记录' : '没有匹配的可用功能'} />}</div></div>;
  }

  function renderUnifiedTimeline() {
    const canvasWidth = Math.max(900, timelineTimeToPixels(durationMs + 2000, timelineZoom) + 20);
    const orderedTracks = ['subtitle', 'effect', 'video', 'audio'].map(kind => displayTracks.find(track => track.kind === kind) || { id: `${kind}-display-track`, kind, name: referenceTrackNames[kind], clips: [], placeholder: true });
    return <div className={`aicut-timeline-body${bladeMode ? ' blade-mode' : ''}`}><div className="aicut-track-labels aicut-reference-labels"><div className="aicut-ruler-spacer">00:00:00</div>{orderedTracks.map(track => <div className={`aicut-track-label kind-${track.kind}`} key={track.id}><span className="aicut-track-kind">{track.kind === 'subtitle' ? 'T' : track.kind === 'video' ? '▣' : '♫'}</span><span className="aicut-track-copy"><strong>{track.name}</strong><small>{track.kind === 'subtitle' ? `${track.locked ? '锁定' : '未锁定'} · 显示` : track.kind === 'video' ? `主轨 · ${linkedMove ? '联动' : '独立'}` : `${track.muted ? '静音' : '有声'} · 波形`}</small></span><div className="aicut-track-actions">{!track.placeholder && <><Dropdown trigger={['click']} menu={{ items: trackMenuItems(track) }}><button type="button" title="轨道菜单" onClick={event => event.stopPropagation()}>☷</button></Dropdown><button type="button" title={track.locked ? '解锁轨道' : '锁定轨道'} onClick={() => toggleTrackFlag(track.id, 'locked')}>{track.locked ? '🔒' : '♢'}</button><button type="button" title={track.muted ? '取消静音' : '静音'} onClick={() => toggleTrackFlag(track.id, 'muted')}>{track.muted ? '×' : '◉'}</button></>}</div></div>)}</div><div className="aicut-tracks-scroll"><div className="aicut-tracks-canvas aicut-reference-tracks" style={{ width: canvasWidth }}><div className="aicut-shared-ruler">00.00　　02.00　　04.00　　06.00　　08.00　　10.00　　12.00　　14.00　　16.00　　18.00　　20.00</div><i className="aicut-playhead" style={{ left: timelineTimeToPixels(previewTimeMs, timelineZoom) }} />{orderedTracks.map(track => <div className={`aicut-track-row kind-${track.kind}`} key={track.id} onPointerDown={seekFromTrack} onDragOver={event => { if (!track.locked && event.dataTransfer.types.includes('application/x-aicut-asset')) event.preventDefault(); }} onDrop={event => { if (event.dataTransfer.types.includes('application/x-aicut-asset')) dropAssetOnTrack(event, track); }}>{track.clips.map(clip => { const key = `${track.id}::${clip.id}`; const shownStart = dragPreview?.key === key ? dragPreview.startMs : Number(clip.startMs) || 0; const transitionLabel = transitionCatalog.find(item => item.id === clip.transition)?.label || clip.transition; return <article key={clip.id} style={{ left: timelineTimeToPixels(shownStart, timelineZoom), width: Math.max(80, timelineTimeToPixels(clip.durationMs, timelineZoom)) }} className={`aicut-clip ${track.kind}${track.locked ? ' locked' : ''}${clip.reviewStatus ? ` review-${clip.reviewStatus}` : ''}${selectedClipKey === key ? ' selected' : ''}${dragPreview?.key === key ? ' dragging' : ''}`} onPointerDown={event => beginClipDrag(event, track, clip)} onPointerMove={updateClipDrag} onPointerUp={finishClipDrag} onPointerCancel={() => { clipDragRef.current = null; setDragPreview(null); }} onClick={() => { if (!bladeMode) setSelectedClipKey(key); }}><strong>{clip.name}</strong>{clip.transition ? <button type="button" className="aicut-transition-chip" title={`转场：${transitionLabel}`} onClick={event => { event.stopPropagation(); setSelectedClipKey(key); }}>{transitionLabel}</button> : null}</article>; })}</div>)}</div></div></div>;
  }

  return (
    <ConfigProvider theme={{ token: { colorBgBase: 'var(--surface-base)', colorBgContainer: 'var(--surface-panel)', colorBgElevated: 'var(--surface-panel)', colorText: 'var(--text-primary)', colorTextSecondary: 'var(--text-secondary)', colorBorder: 'var(--border-default)', colorPrimary: 'var(--accent-primary)' } }}>
    <main className="aicut-page aicut-fullscreen-editor">
      {!projectId ? <div className="aicut-welcome"><Clapperboard size={42} /><h2>创建第一个 AI-CUT 项目</h2><p>创建后进入在线专业剪辑器，素材、预览、属性和时间线都保存到项目。</p><Button type="primary" onClick={() => setCreateOpen(true)}>新建项目</Button></div> : <Spin spinning={loading}>
        <header className="aicut-editor-topbar aicut-hub-topbar aicut-reference-top">
          <button type="button" className="aicut-reference-plain" onClick={() => window.location.assign('/')}>← 返回一战晟铭</button>
          <div className="aicut-reference-brand" aria-label="AI-CUT Hub · 在线版"><b>AI</b>-CUT Hub · 在线版</div>
          <select className="aicut-reference-project" aria-label="切换剪辑项目" value={projectId} onChange={event => setProjectId(event.target.value)}>{projects.map(project => <option key={project.id} value={project.id}>{project.title}</option>)}</select>
          <div className="aicut-reference-modes"><button type="button" className={workspaceMode === 'canvas' ? 'on' : ''} onClick={() => setWorkspaceMode('canvas')}>AI 创作画布</button><button type="button" className={workspaceMode === 'editor' ? 'on' : ''} onClick={() => setWorkspaceMode('editor')}>专业剪辑</button></div>
          <div className="aicut-reference-actions"><span className={`save-${saveState}`}>● {{ saved: '云端已保存', dirty: '有未保存修改', saving: '保存中', failed: '保存失败' }[saveState]}</span><button type="button" aria-label="撤销" title="撤销 Ctrl+Z" onClick={undoTimeline} disabled={!undoRef.current.length && !serverHistory.canUndo}><Undo2 size={15} /></button><button type="button" aria-label="重做" title="重做 Ctrl+Shift+Z" onClick={redoTimeline} disabled={!redoRef.current.length && !serverHistory.canRedo}><Redo2 size={15} /></button><button type="button" aria-label="快捷键" title="键盘快捷键" onClick={() => setShortcutOpen(true)}><Keyboard size={16} /></button><button type="button" title="保存 Ctrl+S" onClick={saveTimeline}><Save size={16} />{saving ? '保存中' : '保存'}</button><button type="button" className="aicut-reference-export" title="导出 Ctrl+E" onClick={() => setExportOpen(true)}><CloudUpload size={15} />云端导出</button></div>
        </header>

        {workspaceMode === 'canvas' ? <section className="aicut-creative-canvas">
          <div className="aicut-canvas-heading"><div><Sparkles size={22} /><h2>AI 创作画布</h2></div><p>把真实项目素材和时间线交给 Agent 编排；每次工具执行都会校验项目版本。</p></div>
          <div className="aicut-canvas-flow"><article><span>1</span><strong>导入素材</strong><small>{assets.length} 个项目素材</small><Button onClick={() => setWorkspaceMode('editor')}>打开媒体区</Button></article><article><span>2</span><strong>描述剪辑目标</strong><small>粗剪、字幕、去口误、配乐</small><Button type="primary" onClick={() => setAgentDrawerOpen(true)}>打开 Agent</Button></article><article><span>3</span><strong>检查时间线</strong><small>{timeline.tracks.length} 条真实轨道 · 修订 {timeline.revision}</small><Button onClick={() => setWorkspaceMode('editor')}>进入专业剪辑</Button></article><article><span>4</span><strong>云端导出</strong><small>MP4 / WebM / MP3 / SRT</small><Button onClick={() => setExportOpen(true)}>打开导出</Button></article></div>
          <div className="aicut-canvas-agent"><AICutAgentPanel embedded projectId={projectId} timelineRevision={timeline.revision} models={agentModels} activeSkill={activeAgentSkill} draft={agentDraft} onTimelineChanged={refreshTimelineFromAgent} /></div>
        </section> : <EditorShell ref={editorShellRef} layout={editorLayout} inspectorDrawerOpen={inspectorDrawerOpen}>
          {editorLayout.inspectorMode === 'drawer' || inspectorCollapsed ? <button type="button" className="aicut-open-inspector" aria-label={editorLayout.inspectorMode === 'drawer' && inspectorDrawerOpen ? '关闭属性栏' : '打开属性栏'} onClick={() => editorLayout.inspectorMode === 'drawer' ? setInspectorDrawerOpen(value => !value) : setInspectorCollapsed(false)}>{editorLayout.inspectorMode === 'drawer' && inspectorDrawerOpen ? '关闭属性' : '属性'}</button> : null}
          <ResourceLibrary activeTab={toolCategory} assets={assets} favorites={favoriteFeatures} recent={recentFeatures} uploading={uploading} collapsed={mediaCollapsed} onToggleCollapsed={() => setMediaCollapsed(value => !value)} onTabChange={setToolCategory} onUpload={upload} onPreviewAsset={(asset) => setSelectedAssetId(asset.id)} onAddAsset={addAssetToTimeline} onOpenASR={(asset) => { setSelectedAssetId(asset.id); setASROpen(true); }} onPreviewResource={previewCatalogResource} onApplyResource={applyCatalogResource} onToggleFavorite={(id) => setFavoriteFeatures(current => toggleFavoriteAICutFeature(current, id))} />
          {toolCategory === '文字' ? <CaptionWorkspace assets={assets} models={asrModels} jobs={asrJobs} assetId={selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? selectedAsset.id : ''} modelId={asrModelId} language={asrLanguage} busy={asrBusy} onAssetChange={setSelectedAssetId} onModelChange={setASRModelId} onLanguageChange={setASRLanguage} onStart={startASR} onCancel={cancelASR} onRetry={(job) => { setSelectedAssetId(job.assetId); setASROpen(true); }} onImportSRT={importSRT} onExportSRT={exportSRT} onAddManualCaption={() => { setManualSubtitle(current => ({ ...current, text: '新字幕', startMs: previewTimeMs })); setToolWorkbenchOpen('字幕'); }} /> : null}
          <PreviewCanvas enabled={Boolean(selectedClipEntry && ['video', 'image'].includes(selectedClipEntry.clip.kind))} transform={selectedClipEntry?.clip?.canvasTransform} onCommitTransform={(canvasTransform) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { canvasTransform })} />
          <InspectorPanel selection={selectedClipEntry} drawerMode={editorLayout.inspectorMode === 'drawer'} onClose={() => setInspectorDrawerOpen(false)} onCollapse={() => setInspectorCollapsed(true)} onPatchClip={(patch) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, patch)} onPatchEffect={(patch) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, patch)} onPatchTransition={(patch) => selectedClipEntry && updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, patch)} />
          <nav className="aicut-tool-rail aicut-reference-rail" aria-label="剪辑工具">{visibleEditorTools.map(([label, Icon]) => <button type="button" key={label} className={toolCategory === label ? 'active' : ''} onClick={() => chooseToolCategory(label)}><Icon size={18} /><span>{label}</span></button>)}</nav>
          <section className={`aicut-library-panel aicut-reference-library-panel${mediaCollapsed ? ' collapsed' : ''}`}>{mediaCollapsed ? <button type="button" className="aicut-collapsed-rail" aria-label="展开素材区" onClick={() => setMediaCollapsed(false)}><PanelLeftOpen size={17} /><span>工具</span></button> : <><header><div><strong>{toolCategory}</strong><small>{capabilities.filter(item => item.category === toolCategory).length} 项</small></div><Button type="text" size="small" aria-label="收起素材区" icon={<PanelLeftClose size={15} />} onClick={() => setMediaCollapsed(true)} /></header>{renderToolLibrary()}</>}</section>
          {!mediaCollapsed ? <div role="separator" aria-label="调整素材区宽度" className="aicut-resize-handle library-preview" onPointerDown={event => beginLayoutResize('media', event)} onPointerMove={updateLayoutResize} onPointerUp={finishLayoutResize} onPointerCancel={finishLayoutResize} /> : null}
          {!timelineCollapsed ? <div role="separator" aria-label="上下拖动调整时间线高度" className="aicut-resize-handle upper-timeline" onPointerDown={event => beginLayoutResize('timeline', event)} onPointerMove={updateLayoutResize} onPointerUp={finishLayoutResize} onPointerCancel={finishLayoutResize} /> : null}

          <section className="aicut-preview aicut-reference-stage"><div className="aicut-panel-title"><strong>播放器</strong><small>{previewFrame.visuals.length} 个画面层 · {previewFrame.audios.length} 个音频层</small></div><div ref={previewStageRef} className={`aicut-preview-stage fit-${canvasFit} ratio-${String(currentProject?.aspectRatio || '9:16').replace(':', '-')}${safeGuide ? ' show-safe-guide' : ''}`}>{!previewFrame.visuals.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="把视频或图片加入时间线后，从播放头预览成片" /> : previewFrame.visuals.map((item, index) => <div className="aicut-preview-visual" style={{ zIndex: index + 1 }} key={`${item.trackId}:${item.clipId}`}><PreviewMedia item={item} playing={previewPlaying} /></div>)}{previewFrame.subtitles.length ? <div className="aicut-preview-subtitles">{previewFrame.subtitles.map(item => <div className={`aicut-preview-caption style-${item.styleId} animation-${item.animation || 'none'}`} key={`${item.trackId}:${item.clipId}`}>{item.text}</div>)}</div> : null}<div className="aicut-preview-audio">{previewFrame.audios.map(item => <PreviewMedia item={item} playing={previewPlaying} key={`${item.trackId}:${item.clipId}`} />)}</div></div><div className="aicut-preview-controls aicut-reference-player"><button type="button" aria-label={previewPlaying ? '暂停预览' : '播放预览'} onClick={togglePreviewPlayback}>{previewPlaying ? <Pause size={15} /> : <Play size={15} />}</button><span>{previewClock(previewTimeMs)}</span><input className="aicut-preview-seek" aria-label="预览播放头" type="range" min="0" max={Math.max(1, durationMs)} step="10" value={previewTimeMs} onChange={event => seekPreview(Number(event.target.value))} /><span>{previewClock(durationMs)}</span><button type="button" title="依次切换完整显示、铺满画布和原始大小" onClick={() => setCanvasFit(value => nextAICutCanvasFit(value))}>{canvasFit === 'fit' ? '适应' : canvasFit === 'fill' ? '填充' : '100%'}</button><button type="button" className={safeGuide ? 'active' : ''} onClick={() => setSafeGuide(value => !value)}>安全框</button><button type="button" onClick={() => previewStageRef.current?.requestFullscreen?.()}>全屏预览</button></div></section>

          <aside className="aicut-inspector"><nav>{inspectorTabs.map(tab => <button type="button" key={tab} className={inspectorTab === tab ? 'active' : ''} onClick={() => setInspectorTab(tab)}>{tab}</button>)}</nav>{selectedClipEntry ? <div className="aicut-inspector-body"><div className="aicut-inspector-title"><Film size={16} /><strong>{selectedClipEntry.clip.name}</strong></div>{inspectorTab === '基础' && <>{selectedClipEntry.clip.kind === 'subtitle' ? <label>文字内容<Input.TextArea value={selectedClipEntry.clip.text || ''} rows={4} onChange={event => { try { commitTimeline(current => setTimelineClipText(current, selectedClipEntry.track.id, selectedClipEntry.clip.id, event.target.value)); } catch (error) { if (event.target.value.trim()) message.error(error.message); } }} /></label> : null}<label>开始时间<InputNumber min={0} value={selectedClipEntry.clip.startMs} onChange={value => moveClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, Number(value || 0))} addonAfter="ms" /></label><label>片段时长<InputNumber min={1} value={selectedClipEntry.clip.durationMs} onChange={value => updateClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, { durationMs: Number(value || 1) })} addonAfter="ms" /></label><div><Button onClick={() => trimClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, 'left')}>左裁 0.1s</Button><Button onClick={() => trimClip(selectedClipEntry.track.id, selectedClipEntry.clip.id, 'right')}>右裁 0.1s</Button>{selectedClipEntry.clip.kind === 'subtitle' ? <Button onClick={() => { try { commitTimeline(current => mergeTimelineClipWithNext(current, selectedClipEntry.track.id, selectedClipEntry.clip.id)); } catch (error) { message.warning(error.message); } }}>合并下一条</Button> : null}</div></>}{inspectorTab === '速度' && <label>播放速度<Select value={Number(selectedClipEntry.clip.playbackRate) || 1} onChange={value => changeClipSpeed(selectedClipEntry.track.id, selectedClipEntry.clip.id, value)} options={[.5, .75, 1, 1.25, 1.5, 2].map(value => ({ value, label: `${value}x` }))} /></label>}</div> : <div className="aicut-inspector-empty"><strong>选择时间线片段</strong><span>片段参数会显示在这里</span></div>}</aside>

          <section className={`aicut-timeline aicut-reference-timeline${timelineCollapsed ? ' collapsed' : ''}`}><div className="aicut-timeline-toolbar"><div><Select key={historyVersion} size="small" placeholder="＋ 轨道" value={undefined} onChange={addTrack} options={[{ value: 'subtitle', label: '新增字幕轨' }, { value: 'video', label: '新增视频轨' }, { value: 'audio', label: '新增音频轨' }]} /><Button size="small" type={bladeMode ? 'primary' : 'default'} icon={<Scissors size={13} />} onClick={() => setBladeMode(value => !value)}>剃刀</Button><Button size="small" icon={<Scissors size={13} />} disabled={!selectedClipEntry} onClick={() => selectedClipEntry && splitClip(selectedClipEntry.track.id, selectedClipEntry.clip)}>中点分割</Button><Button size="small" icon={<Copy size={13} />} disabled={!selectedClipEntry} onClick={duplicateSelectedClip}>复制</Button><Button size="small" danger icon={<Trash2 size={13} />} disabled={!selectedClipEntry} onClick={() => selectedClipEntry && removeClip(selectedClipEntry.track.id, selectedClipEntry.clip.id)}>删除</Button>{(() => { const track = selectedClipEntry?.track || timeline.tracks[0]; return <Dropdown disabled={!track} trigger={['click']} menu={{ items: trackMenuItems(track) }}><Button size="small" title={track ? `管理：${track.name}` : '当前没有轨道'}>轨道菜单</Button></Dropdown>; })()}</div><div><Button size="small" type={linkedMove ? 'primary' : 'default'} onClick={() => setLinkedMove(value => !value)}>联动</Button><Button size="small" type={magneticSnap ? 'primary' : 'default'} icon={<Magnet size={13} />} onClick={() => setMagneticSnap(value => !value)}>吸附</Button><Select size="small" aria-label="时间线缩放" value={timelineZoom} onChange={setTimelineZoom} options={[{ value: .5, label: '50%' }, { value: 1, label: '100%' }, { value: 2, label: '200%' }, { value: 4, label: '400%' }]} /><button type="button" className="aicut-timeline-collapse" aria-label={timelineCollapsed ? '展开时间线' : '收起时间线'} onClick={() => setEditorLayout(current => ({ ...current, timelineCollapsed: !current.timelineCollapsed }))}>{timelineCollapsed ? <PanelBottomOpen size={15} /> : <PanelBottomClose size={15} />}</button></div></div>{renderUnifiedTimeline()}</section>
        </EditorShell>}

        <button type="button" className="aicut-agent-fab" aria-label="打开 AI-CUT Agent" title="打开 AI-CUT Agent" onClick={() => setAgentDrawerOpen(true)}><Bot size={18} /><span>AI-CUT Agent</span></button>{agentDrawerOpen && <div className="aicut-agent-drawer"><header><strong>AI-CUT Agent</strong><Button type="text" aria-label="关闭 Agent" icon={<X size={18} />} onClick={() => setAgentDrawerOpen(false)} /></header><AICutAgentPanel embedded projectId={projectId} timelineRevision={timeline.revision} models={agentModels} activeSkill={activeAgentSkill} draft={agentDraft} onTimelineChanged={refreshTimelineFromAgent} /></div>}
        {asrJobs.length ? <section className="aicut-asr-jobs"><div className="aicut-panel-title"><strong>语音识别任务</strong></div>{asrJobs.slice(0, 5).map(job => <div className="aicut-asr-job" key={job.id}><div><strong>{assets.find(asset => asset.id === job.assetId)?.name || job.assetId}</strong><small>{asrStateLabels[job.state] || job.state} · {job.progress || 0}%{job.error ? ` · ${job.error}` : ''}</small></div><Tag>{asrStateLabels[job.state] || job.state}</Tag>{['queued', 'running', 'cancel_requested'].includes(job.state) ? <Button size="small" onClick={() => cancelASR(job)}>取消</Button> : ['failed', 'canceled'].includes(job.state) ? <Button size="small" onClick={() => { setSelectedAssetId(job.assetId); setASROpen(true); }}>重新识别</Button> : null}</div>)}</section> : null}
      </Spin>}

      <Modal title="键盘快捷键" open={shortcutOpen} onCancel={() => setShortcutOpen(false)} footer={<Button type="primary" onClick={() => setShortcutOpen(false)}>知道了</Button>} width={760}><div className="aicut-shortcut-grid"><section><h3>通用</h3><p><span>新建项目</span><kbd>Ctrl</kbd><kbd>N</kbd></p><p><span>保存工程</span><kbd>Ctrl</kbd><kbd>S</kbd></p><p><span>导出作品</span><kbd>Ctrl</kbd><kbd>E</kbd></p></section><section><h3>视图</h3><p><span>AI 创作画布</span><kbd>Ctrl</kbd><kbd>1</kbd></p><p><span>专业剪辑</span><kbd>Ctrl</kbd><kbd>2</kbd></p><p><span>快捷键</span><kbd>Ctrl</kbd><kbd>/</kbd></p></section><section><h3>播放与剪辑</h3><p><span>播放 / 暂停</span><kbd>Space</kbd></p><p><span>分割</span><kbd>Ctrl</kbd><kbd>B</kbd></p><p><span>删除</span><kbd>Delete</kbd></p></section><section><h3>历史</h3><p><span>撤销</span><kbd>Ctrl</kbd><kbd>Z</kbd></p><p><span>重做</span><kbd>Ctrl</kbd><kbd>Shift</kbd><kbd>Z</kbd></p></section></div></Modal>
      <Modal title="新建 AI-CUT 项目" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={createProject} okText="创建"><div className="aicut-create-form"><label>项目名称<Input value={newTitle} onChange={event => setNewTitle(event.target.value)} placeholder="例如：第一条竖屏口播" /></label><label>画面比例<Select value={newRatio} onChange={setNewRatio} options={[{ value: '9:16', label: '9:16 竖屏' }, { value: '16:9', label: '16:9 横屏' }, { value: '1:1', label: '1:1 方形' }]} /></label></div></Modal>
      <Modal title="重命名轨道" open={Boolean(trackEditId)} onCancel={() => { setTrackEditId(''); setTrackNameDraft(''); }} onOk={confirmTrackRename} okText="保存"><Input value={trackNameDraft} maxLength={64} autoFocus onPressEnter={confirmTrackRename} onChange={event => setTrackNameDraft(event.target.value)} placeholder="输入轨道名称" /></Modal>
      <Modal title="字幕纠错" open={captionOpen} onCancel={cancelCaptionCorrection} onOk={correctCaptions} okText="开始纠错" confirmLoading={correctingCaptions}><div className="aicut-create-form"><label>校对剧本<Input.TextArea value={captionScript} onChange={event => setCaptionScript(event.target.value)} rows={8} /></label><label>文本模型（可选）<Input value={captionModelId} onChange={event => setCaptionModelId(event.target.value)} /></label></div></Modal>
      <Modal title="识别字幕" open={asrOpen} onCancel={() => setASROpen(false)} onOk={startASR} okText="开始识别" okButtonProps={{ disabled: !asrModels.length || !assets.some(asset => ['audio', 'video'].includes(asset.kind)) }} confirmLoading={asrBusy}><div className="aicut-create-form"><label>识别素材<Select value={selectedAsset && ['audio', 'video'].includes(selectedAsset.kind) ? selectedAsset.id : undefined} onChange={setSelectedAssetId} placeholder="选择视频或音频" options={assets.filter(asset => ['audio', 'video'].includes(asset.kind)).map(asset => ({ value: asset.id, label: asset.name }))} /></label><label>语音识别模型<Select value={asrModelId || undefined} onChange={setASRModelId} placeholder="选择已启用的 ASR 模型" options={asrModels.map(model => ({ value: model.id, label: model.displayName || model.modelId || model.id }))} /></label>{!asrModels.length ? <Button onClick={() => window.location.assign('/settings')}>前往 API 配置</Button> : null}<label>语言<Select value={asrLanguage} onChange={setASRLanguage} options={[{ value: 'zh', label: '中文' }, { value: 'en', label: '英文' }, { value: '', label: '自动检测' }]} /></label><small>识别成功后会写入新的字幕轨，不会覆盖现有字幕。任务可取消，失败后可以重试。</small></div></Modal>
      <Modal title="加入背景音乐" open={bgmOpen} footer={null} onCancel={() => setBGMOpen(false)}><div className="aicut-bgm-list">{assets.filter(asset => asset.kind === 'audio').map(asset => <button type="button" key={asset.id} onClick={() => applyBGMPlan(asset)}><strong>{asset.name}</strong><span>加入时间线</span></button>)}</div></Modal>
      <Modal title="批量替换字幕" open={replaceOpen} onCancel={() => setReplaceOpen(false)} onOk={replaceSubtitleText}><div className="aicut-create-form"><label>查找<Input value={replaceFrom} onChange={event => setReplaceFrom(event.target.value)} /></label><label>替换为<Input value={replaceTo} onChange={event => setReplaceTo(event.target.value)} /></label></div></Modal>
      <Modal title="导出" open={exportOpen} footer={null} onCancel={() => setExportOpen(false)} width={720}><div className="aicut-export-panel"><div><strong>{currentProject?.title || '未命名工程'}</strong><span>{currentProject?.aspectRatio || '9:16'} · 30 fps</span></div><div className="aicut-export-types"><Button loading={exportingFormat === 'mp4'} onClick={() => exportMedia('mp4')}>成片 · MP4</Button><Button loading={exportingFormat === 'webm'} onClick={() => exportMedia('webm')}>成片 · WebM</Button><Button loading={exportingFormat === 'mp3'} onClick={() => exportMedia('mp3')}>音轨 · MP3</Button><Button type="primary" onClick={exportSRT}>字幕文件 · SRT</Button></div>{exportJob && <Progress percent={Number(exportJob.progress || 0)} />}<p>由服务器读取已保存时间线，通过 FFmpeg 后台渲染；可查看进度并取消。</p></div></Modal>
      <Modal title="识别并删除水词字幕" open={fillerOpen} onCancel={cancelFillerTrim} onOk={trimFillers} confirmLoading={trimmingFillers}><label>文本模型（可选）<Input value={fillerModelId} onChange={event => setFillerModelId(event.target.value)} /></label></Modal>
      <Modal title="去口误停顿" open={oralOpen} onCancel={cancelFillerTrim} onOk={cleanOralTimeline} confirmLoading={trimmingFillers}><label>文本模型（可选）<Input value={fillerModelId} onChange={event => setFillerModelId(event.target.value)} /></label></Modal>
      <Modal title={toolWorkbenchOpen} open={Boolean(toolWorkbenchOpen)} footer={null} width={820} onCancel={() => setToolWorkbenchOpen('')}><div className="aicut-tool-workbench">{renderMaterialBody()}</div></Modal>
    </main>
    </ConfigProvider>
  );
}

