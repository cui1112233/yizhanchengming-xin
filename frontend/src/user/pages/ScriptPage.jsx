import { Button, Dropdown, Form, Input, Modal, Popconfirm, Select, Segmented, Space, Switch, Tooltip, Typography, message } from 'antd';
import { AudioLines, Clapperboard, Copy, Download, FileText, Info, MoreHorizontal, Pencil, Plus, RefreshCw, Settings2, Star, WandSparkles } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { deleteScriptConstraintPrompt, extractCharactersAndScenes, generateScript, getConstraintPresetTexts, listScriptConstraintPrompts, listScriptPresetCatalog, saveScriptConstraintPrompt, updateScriptConstraintPrompt } from '../../shared/api/generation';
import { listHistory, saveHistory, updateHistoryVideoTasks } from '../../shared/api/history';
import { getConfig, listAvailableModels, saveConfig } from '../../shared/api/config';
import { playTaskSound } from '../../shared/notifications/taskSound';
import { textToSpeech } from '../../shared/api/tts';
import { getCurrentUsername } from '../../shared/api/auth';
import { getMemberCenter } from '../../shared/api/member';
import { apiRequest } from '../../shared/api/client';
import { useLocalExecutorStatus } from '../../shared/hooks/useLocalExecutorStatus.js';
import { IMAGE_ASPECT_RATIOS, isSupportedImageAspectRatio } from '../../shared/imageAspectRatios.js';
import { PET_APPLY_EVENT, PET_PREVIEW_EVENT, dispatchPetContext, dispatchPetState } from '../../shared/pet/stacky';
import { dispatchCmSelection } from '../../shared/pet/cmBridge';
import { getScriptDraftTabId, loadScriptDraft, saveScriptDraft, storageFailureReason } from './scriptDraftStorage';
import { DEFAULT_SCRIPT_CONSTRAINTS, constraintsForFormat, constraintsForNextGeneration, normalizeScriptConstraints } from './scriptConstraints';
import { normalizeAudioDurationSeconds, readAudioDurationFromUrl } from './scriptGenerationRules';
import { filterExtractionPresets, selectAvailableExtractionPreset } from './scriptExtractionPresets';
import { createEntity, entityData, normalizeExtractInfo, toGenerationEntities } from './scriptEntities';
import { buildReferenceAssetGenerationPayload } from './scriptEntityImages';
import { removeEntityConstraintReferences, scriptEntitySelection, useScriptCmBridge } from './scriptCmBridge';
import { applyEntityEnrichment, compactEntitySummary, entityName, normalizeEntityEnrichment } from './scriptEntityEnrichment';
import { getShotCardsWithinDuration, joinShotCards } from './scriptShotOutput';
import { getSelectedShotMatches, getShotCardStarts, replaceAllSelectedShotMatches, replaceSelectedShotMatch } from './scriptShotReplace';
import { buildFinalSegmentCard } from './scriptFinalSegment';
import { resolveShotVideoDuration } from './scriptVideoDuration';
import { buildScriptVideoPayload, collectShotReferenceImages, getEntityMedia, toggleShotReferenceState } from './scriptVideoReferences';
import { appendShotVideoTaskHistory, mergeShotVideoTaskHistoryTask, normalizeShotVideoTaskHistory } from './scriptShotVideoTasks';
import { extractJSON } from './scriptExtractionJson';
import { replaceRawShotCard } from './scriptShotCardEdit';
import { filterShotMentionCandidates, findActiveShotMention, insertActiveShotMention } from './scriptShotMentions';
import { ShotOutputCards } from '../components/ShotOutputCards';
import EntityImagePanel from '../components/EntityImagePanel';
import ScriptMentionEditor from '../components/ScriptMentionEditor';
import { createScriptVideo, getScriptVideoTask } from '../../shared/api/scriptVideo';
import { createPrivateComicTask, downloadComicPromptZip, generateBookAssetImages, getBatch, getBatchRuntimeSummary, getBookAssetImageGenerationTask, listBookAssetImages, listBookAssets, listComicPages, listSystemPresetCatalog, runBookStage, saveBatchSettings, setPrimaryBookAssetImage, updateBookAsset, updateComicPage, uploadBookAssetImage, uploadComicGlobalReference } from '../../shared/api/batchFactory';
import { defaultEnabledModelSelection, loadScriptModelSelection, reconcileScriptModelSelection, saveScriptModelSelection } from './scriptModelSelection';
import { modelOptionsForProvider, providerOptionsForModels, reconcileProviderModel } from './shuihuo/providerModelBinding';
import PrivateComicWorkspace from './PrivateComicWorkspace';
import { PrivateStudioChips } from './private-studio/PrivateStudioChips';
import { PrivateStudioAssetEditor } from './private-studio/PrivateStudioAssetEditor';
import { PrivateStudioConstraintsModal } from './private-studio/PrivateStudioConstraintsModal';
import { PrivateStudioHeaderActions } from './private-studio/PrivateStudioHeaderActions';
import { PrivateStudioNote } from './private-studio/PrivateStudioNote';
import { PrivateStudioSteps } from './private-studio/PrivateStudioSteps';
import { buildPrivateComicSettings, mergePrivateComicReference, privateComicDraftKey, privateComicTaskKey, privateComicTitle } from './privateCustomizationComic';
import { ensurePrivateComicTask, preparePrivateComicTaskSettings } from './privateComicSingleBookFlow.js';
import { comicPageErrorMessage, comicPageGenerationPayload, generateComicPageWithConflict } from './shuihuo/batchFactoryComicMode';
import { readImageAsDataURL } from './shuihuo/batchFactoryNovelListHelpers';
import './shuihuo-production.css';
import './private-customization.css';
import './private-studio/private-studio-workbench.css';
import './private-studio/private-studio-asset-editor.css';

function formatDraftSavedTime(value) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
}

function privateComicOperationId() {
  try {
    return crypto.randomUUID();
  } catch (_) {
    return `comic-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
  }
}

function videoModelKey(model) {
  return String(model?.key || model?.modelKey || model?.id || '').trim();
}

function videoModelLabel(model) {
  const key = videoModelKey(model);
  const name = String(model?.name || model?.displayName || key || '未命名视频模型').trim();
  const timing = key === 'minimax-h3-video' ? '1-15秒' : key === 'seedance-2-0-official' ? '4-15秒' : key === 'yd2-mini-video' ? '固定10秒' : '';
  const unavailable = key === 'minimax-h3-video' && model?.configured === false ? '待配置 Token' : '';
  const details = [timing, unavailable].filter(Boolean);
  return details.length ? `${name}（${details.join('，')}）` : name;
}

const DEFAULT_SCRIPT_UNIFIED_CONFIG = { textModelId: '', imageProvider: 'personal_api', imageModelId: '', videoProvider: 'personal_api', videoModelKey: 'yd2-mini-video', imageAspectRatio: '9:16', videoAspectRatio: '9:16', videoResolution: '720p' };

function normalizeScriptUnifiedConfig(value = {}) {
  return {
    ...DEFAULT_SCRIPT_UNIFIED_CONFIG,
    ...value,
    imageAspectRatio: isSupportedImageAspectRatio(value?.imageAspectRatio) ? value.imageAspectRatio : '9:16',
    videoAspectRatio: ['16:9', '9:16'].includes(value?.videoAspectRatio) ? value.videoAspectRatio : '9:16',
    videoResolution: ['480p', '720p', '1080p'].includes(value?.videoResolution) ? value.videoResolution : '720p'
  };
}

function aiText(response) {
  if (typeof response === 'string') return response;
  const candidates = [response, response?.data, response?.result, response?.response];

  for (const candidate of candidates) {
    if (!candidate) continue;
    if (typeof candidate.output_text === 'string') return candidate.output_text;
    if (candidate.人物设定 || candidate.人物 || candidate.角色设定 || candidate.characters) return candidate;

    const content = candidate?.choices?.[0]?.message?.content
      ?? candidate?.choices?.[0]?.delta?.content
      ?? candidate?.message?.content
      ?? candidate?.content;

    if (typeof content === 'string' || (content && typeof content === 'object' && !Array.isArray(content))) return content;
    if (Array.isArray(content)) {
      return content
        .map(item => typeof item === 'string' ? item : item?.text || item?.content || '')
        .join('');
    }
  }

  return '';
}

function normalizeExtraction(data) {
  const asList = value => Array.isArray(value) ? value : value ? [value] : [];
  const source = data?.data || data?.result || data || {};
  return normalizeExtractInfo({
    characters: asList(source.人物设定 || source.人物 || source.角色设定 || source.characters),
    scenes: asList(source.场景设定 || source.场景 || source.scene_options || source.scenes),
    visualStyle: source.统一风格 || source.visualStyle || source.style || source.trailer_style || ''
  });
}

function formatEntity(item) {
  const value = entityData(item);
  if (typeof value === 'string') return value;
  return value?.角色名称 || value?.场景名称 || value?.名称 || value?.name || value?.场景 || value?.scene || '未命名对象';
}

function visualFields(item) {
  const value = entityData(item);
  if (typeof value === 'string') return [{ key: '描述', label: '设定 / 描述', value }];
  const labels = { name: '名称', 名称: '名称', 人物: '名称', 场景: '名称', scene: '名称', 角色名称: '名称', 场景名称: '名称', identity: '身份', 身份: '身份', appearance: '外形', 外形: '外形', 外观描述: '外形', personality: '性格', 性格: '性格', relation: '关系', 关系: '关系', time: '时段', 时段: '时段', atmosphere: '氛围', 氛围: '氛围', 氛围概述: '氛围', description: '设定 / 描述', 描述: '设定 / 描述', 场景描述: '设定 / 描述' };
  const mediaKeys = new Set(['imageUrls', 'images', 'imageList', 'mainImageUrl']);
  return Object.entries(value || {})
    .filter(([key]) => !mediaKeys.has(key))
    .map(([key, fieldValue]) => ({ key, label: labels[key] || key, value: typeof fieldValue === 'string' ? fieldValue : JSON.stringify(fieldValue) }));
}

function ScriptMentionMenu({ anchorRect, candidates, activeIndex, onChoose, onCreateCharacter, onCreateScene }) {
  if (!anchorRect) return null;
  const menuHeight = Math.min(300, 86 + candidates.length * 50);
  const opensAbove = anchorRect.bottom + menuHeight > window.innerHeight;
  const top = opensAbove ? Math.max(8, anchorRect.top - menuHeight - 6) : anchorRect.bottom + 6;
  return <div role="listbox" aria-label="人物与场景提及候选" style={{ position: 'fixed', zIndex: 1100, left: Math.max(8, anchorRect.left), top, width: 280, maxHeight: 300, overflowY: 'auto', padding: 8, border: '1px solid #303846', borderRadius: 10, background: '#171b21', color: '#fff', boxShadow: '0 12px 28px rgb(0 0 0 / 30%)' }}>
    <Typography.Text style={{ display: 'block', padding: '2px 6px 8px', color: '#8f98a8', fontSize: 12 }}>可能@的内容</Typography.Text>
    <div style={{ display: 'flex', gap: 4, marginBottom: 4 }}><Button type="text" style={{ color: '#fff', textAlign: 'left' }} onMouseDown={event => event.preventDefault()} onClick={onCreateCharacter}>＋ 创建人物</Button><Button type="text" style={{ color: '#fff', textAlign: 'left' }} onMouseDown={event => event.preventDefault()} onClick={onCreateScene}>＋ 创建场景</Button></div>
    {candidates.length ? candidates.map((candidate, index) => {
      const candidateImageUrl = candidate.imageUrl;
      return <Button key={`${candidate.kind}-${candidate.item.id}`} type="text" block onMouseDown={event => event.preventDefault()} onClick={() => onChoose(candidate)} style={{ height: 42, display: 'flex', alignItems: 'center', gap: 8, color: '#fff', textAlign: 'left', background: index === activeIndex ? '#2c3440' : 'transparent' }}>
        {candidateImageUrl ? <img src={candidateImageUrl} alt="" style={{ width: 30, height: 30, borderRadius: 6, objectFit: 'cover' }} /> : <span style={{ width: 30, height: 30, borderRadius: 6, display: 'inline-grid', placeItems: 'center', background: '#3a4350' }}>{candidate.kind === 'scene' ? '景' : '人'}</span>}
        <span style={{ display: 'inline-flex', flexDirection: 'column', minWidth: 0 }}><b style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>@{candidate.label}</b><small style={{ color: '#aab3c2' }}>{candidate.kind === 'scene' ? '场景' : '人物'}</small></span>
      </Button>;
    }) : <Typography.Text type="secondary" style={{ display: 'block', padding: 6 }}>没有匹配项，请先创建或提取人物/场景</Typography.Text>}
  </div>;
}

export function ScriptPage() {
  const [form] = Form.useForm();
  const { status: localExecutorStatus } = useLocalExecutorStatus({ enabled: true, intervalMs: 30000 });
  const [extracting, setExtracting] = useState(false);
  const [draftSaveState, setDraftSaveState] = useState({ status: 'idle', savedAt: null, reason: '' });
  const [privateComicSaveState, setPrivateComicSaveState] = useState({ status: 'idle', savedAt: null, reason: '' });
  const [generating, setGenerating] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [soundVolume, setSoundVolume] = useState(60);
  const [regeneratingEntities, setRegeneratingEntities] = useState(false);
  const [regeneratingOutput, setRegeneratingOutput] = useState(false);
  const [generationStage, setGenerationStage] = useState('idle');
  const [extractInfo, setExtractInfo] = useState(() => normalizeExtractInfo());
  const [output, setOutput] = useState('');
  const [revisionPreview, setRevisionPreview] = useState({ open: false, summary: '', currentOutput: '', candidateOutput: '' });
  const [previousOutput, setPreviousOutput] = useState('');
  const [editingOutput, setEditingOutput] = useState(false);
  const [editingOutputSelection, setEditingOutputSelection] = useState({ start: 0, end: 0 });
  const [outputMentionAnchorRect, setOutputMentionAnchorRect] = useState(null);
  const [outputMentionMenuOpen, setOutputMentionMenuOpen] = useState(false);
  const [outputMentionActiveIndex, setOutputMentionActiveIndex] = useState(0);
  const [selectedShotIndexes, setSelectedShotIndexes] = useState(new Set());
  const [generatingShotIndexes, setGeneratingShotIndexes] = useState(() => new Set());
  const [shotVideoTasks, setShotVideoTasks] = useState({});
  const [shotVideoTaskHistory, setShotVideoTaskHistory] = useState({});
  const [shotReferenceStates, setShotReferenceStates] = useState({});
  const [scriptVideoModelKey, setScriptVideoModelKey] = useState('yd2-mini-video');
  const [scriptVideoProvider, setScriptVideoProvider] = useState('personal_api');
  const [canGenerateVideo, setCanGenerateVideo] = useState(false);
  const [scriptVideoModels, setScriptVideoModels] = useState([]);
  const [loadingScriptVideoModels, setLoadingScriptVideoModels] = useState(false);
  const [scriptModels, setScriptModels] = useState({ text: [], image: [] });
  const [scriptModelsLoaded, setScriptModelsLoaded] = useState(false);
  const [scriptModelSelection, setScriptModelSelection] = useState({ textModelId: '', imageModelId: '' });
  const [scriptImageProvider, setScriptImageProvider] = useState('personal_api');
  const [scriptImageAspectRatio, setScriptImageAspectRatio] = useState('9:16');
  const [scriptVideoAspectRatio, setScriptVideoAspectRatio] = useState('9:16');
  const [scriptVideoResolution, setScriptVideoResolution] = useState('720p');
  const [scriptConfigOpen, setScriptConfigOpen] = useState(false);
  const [scriptConfigDraft, setScriptConfigDraft] = useState(DEFAULT_SCRIPT_UNIFIED_CONFIG);
  const [savingScriptConfig, setSavingScriptConfig] = useState(false);
  const [previewVideoTask, setPreviewVideoTask] = useState(null);
  const [previewVideoHistory, setPreviewVideoHistory] = useState({ open: false, shotIndex: 0, tasks: [] });
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyEntries, setHistoryEntries] = useState([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [currentHistoryId, setCurrentHistoryId] = useState('');
  const [shotReplaceOpen, setShotReplaceOpen] = useState(false);
  const [shotFindText, setShotFindText] = useState('');
  const [shotReplaceText, setShotReplaceText] = useState('');
  const [shotMatchIndex, setShotMatchIndex] = useState(0);
  const [editingShot, setEditingShot] = useState({ index: -1, text: '' });
  const [editingShotSelection, setEditingShotSelection] = useState({ start: 0, end: 0 });
  const [shotMentionAnchorRect, setShotMentionAnchorRect] = useState(null);
  const [mentionMenuOpen, setMentionMenuOpen] = useState(false);
  const [mentionActiveIndex, setMentionActiveIndex] = useState(0);
  const [activeEntity, setActiveEntity] = useState(null);
  const entityEditorSessionRef = useRef(0);
  const [fullscreenEditor, setFullscreenEditor] = useState(false);
  const [leftPanelWidth, setLeftPanelWidth] = useState(null);
  const [narrating, setNarrating] = useState(false);

  const activeShotMention = useMemo(() => (
    mentionMenuOpen ? findActiveShotMention(editingShot.text, editingShotSelection.start) : null
  ), [editingShot.text, editingShotSelection.start, mentionMenuOpen]);
  const activeShotMentionCandidates = useMemo(() => (
    activeShotMention
      ? filterShotMentionCandidates({
        characters: extractInfo.characters,
        scenes: extractInfo.scenes,
        query: activeShotMention.query,
        label: formatEntity
      }).map(candidate => ({ ...candidate, label: formatEntity(candidate.item), imageUrl: getEntityMedia(candidate.item).mainImageUrl }))
      : []
  ), [activeShotMention, extractInfo.characters, extractInfo.scenes]);
  const activeOutputMention = useMemo(() => (
    outputMentionMenuOpen ? findActiveShotMention(output, editingOutputSelection.start) : null
  ), [editingOutputSelection.start, output, outputMentionMenuOpen]);
  const activeOutputMentionCandidates = useMemo(() => (
    activeOutputMention
      ? filterShotMentionCandidates({
        characters: extractInfo.characters,
        scenes: extractInfo.scenes,
        query: activeOutputMention.query,
        label: formatEntity
      }).map(candidate => ({ ...candidate, label: formatEntity(candidate.item), imageUrl: getEntityMedia(candidate.item).mainImageUrl }))
      : []
  ), [activeOutputMention, extractInfo.characters, extractInfo.scenes]);
  const mentionCandidates = useMemo(() => (
    filterShotMentionCandidates({
      characters: extractInfo.characters,
      scenes: extractInfo.scenes,
      query: '',
      label: formatEntity
    }).map(candidate => ({ ...candidate, label: formatEntity(candidate.item), imageUrl: getEntityMedia(candidate.item).mainImageUrl }))
  ), [extractInfo.characters, extractInfo.scenes]);

  useEffect(() => {
    let active = true;
    getMemberCenter().then(result => {
      if (!active) return;
      const member = result?.member;
      setCanGenerateVideo(
        ['dev', 'manager'].includes(member?.role)
          || member?.apiScopes?.includes('*')
          || member?.apiScopes?.includes('video')
      );
    }).catch(() => {
      if (active) setCanGenerateVideo(false);
    });
    return () => { active = false; };
  }, []);

  const [quickDirectorOpen, setQuickDirectorOpen] = useState(false);
  const [customizationWorkspace, setCustomizationWorkspace] = useState('script');
  const [privateComicCatalog, setPrivateComicCatalog] = useState([]);
  const [privateComicDraft, setPrivateComicDraft] = useState({ title: '', sourceText: '', comic: { enabled: true }, textModelId: '', imageModelId: '', imageProvider: 'personal_api' });
  const [privateComicTask, setPrivateComicTask] = useState(null);
  const [privateComicPendingReference, setPrivateComicPendingReference] = useState('');
  const [privateComicBusy, setPrivateComicBusy] = useState(false);
  const [privateComicError, setPrivateComicError] = useState('');
  const [privateComicAssets, setPrivateComicAssets] = useState([]);
  const [privateComicAssetImages, setPrivateComicAssetImages] = useState({});
  const [privateComicAssetImageFailures, setPrivateComicAssetImageFailures] = useState([]);
  const [privateComicPages, setPrivateComicPages] = useState([]);
  // 后端 listComicPages 返回的书级权威就绪度；前端只渲染，不再自行重算口径。
  const [privateComicReadiness, setPrivateComicReadiness] = useState(null);
  const [privateComicAction, setPrivateComicAction] = useState('');
  const [privateComicActionKey, setPrivateComicActionKey] = useState('');
  const privateComicPollRequestRef = useRef(0);
  const [quickDirectorOptions, setQuickDirectorOptions] = useState({ matchAudio: false });
  const [sourceAudioUrl, setSourceAudioUrl] = useState('');
  const [sourceAudioDurationSeconds, setSourceAudioDurationSeconds] = useState(null);
  const [sourceAudioError, setSourceAudioError] = useState('');
  const [instructionModalOpen, setInstructionModalOpen] = useState(false);
  const [pendingExtractionPreset, setPendingExtractionPreset] = useState('standard');
  const [constraintModalOpen, setConstraintModalOpen] = useState(false);
  const [activeConstraintCategory, setActiveConstraintCategory] = useState('prefix');
  const [constraints, setConstraints] = useState(DEFAULT_SCRIPT_CONSTRAINTS);
  const [outputConstraints, setOutputConstraints] = useState(DEFAULT_SCRIPT_CONSTRAINTS);
  const [draftConstraints, setDraftConstraints] = useState(DEFAULT_SCRIPT_CONSTRAINTS);
  const [constraintCatalog, setConstraintCatalog] = useState([]);
  const [constraintTexts, setConstraintTexts] = useState({});
  const [personalConstraintPrompts, setPersonalConstraintPrompts] = useState({});
  const [loadingPersonalConstraintPrompts, setLoadingPersonalConstraintPrompts] = useState({});
  const [savingConstraintCategory, setSavingConstraintCategory] = useState('');
  const [editingPersonalPromptId, setEditingPersonalPromptId] = useState('');
  const [editingPersonalPromptName, setEditingPersonalPromptName] = useState('');
  const [personalPromptNameModal, setPersonalPromptNameModal] = useState({ open: false, category: '', name: '' });
  const [extractionPresets, setExtractionPresets] = useState([]);
  const [loadingExtractionPresets, setLoadingExtractionPresets] = useState(true);
  const [extractionPresetError, setExtractionPresetError] = useState('');
  const selectedFormat = Form.useWatch('format', form);
  const selectedMode = Form.useWatch('mode', form);
  const selectedDuration = Form.useWatch('duration', form);
  const novelText = Form.useWatch('novelText', form) || '';
  useEffect(() => {
    const saved = loadScriptModelSelection(window.localStorage, draftUsernameRef.current);
    setScriptModelSelection(current => ({ textModelId: current.textModelId || saved.textModelId, imageModelId: current.imageModelId || saved.imageModelId }));
    setScriptVideoModelKey(current => current || saved.videoModelKey || 'yd2-mini-video');
  }, []);
  useEffect(() => {
    if (!scriptModelsLoaded) return;
    saveScriptModelSelection(window.localStorage, draftUsernameRef.current, { ...scriptModelSelection, videoModelKey: scriptVideoModelKey });
  }, [scriptModelsLoaded, scriptModelSelection, scriptVideoModelKey]);
  useEffect(() => {
    let active = true;
    setLoadingScriptVideoModels(true);
    listAvailableModels('video')
      .then(result => {
        if (!active) return;
        const models = (Array.isArray(result?.models) ? result.models : [])
          .filter(model => model?.kind === 'video' && videoModelKey(model));
        setScriptVideoModels(models);
        setScriptVideoModelKey(current => models.some(model => videoModelKey(model) === current)
          ? current
          : videoModelKey(models[0]) || current);
      })
      .catch(error => { if (active) message.warning(error?.message || '读取视频模型失败'); })
      .finally(() => { if (active) setLoadingScriptVideoModels(false); });
    return () => { active = false; };
  }, []);
  useEffect(() => {
    let active = true;
    Promise.all(['text', 'image'].map(kind => listAvailableModels(kind).then(result => [kind, Array.isArray(result?.models) ? result.models : []])))
      .then(entries => {
        if (!active) return;
        const nextModels = Object.fromEntries(entries);
        setScriptModels(nextModels);
        setScriptModelSelection(current => {
          const next = reconcileScriptModelSelection(current, nextModels);
          if (next.unavailableKinds.length) message.warning(`已选${next.unavailableKinds.map(kind => kind === 'text' ? '文本模型' : '图片模型').join('、')}已下架，请重新选择`);
          return defaultEnabledModelSelection(next, nextModels);
        });
      })
      .catch(error => { if (active) message.warning(error?.message || '读取剧本模型失败'); })
      .finally(() => { if (active) setScriptModelsLoaded(true); });
    return () => { active = false; };
  }, []);
  useEffect(() => {
    if (!scriptModelsLoaded) return;
    setScriptModelSelection(current => {
      const next = reconcileScriptModelSelection(current, scriptModels);
      if (next.unavailableKinds.length) message.warning(`已选${next.unavailableKinds.map(kind => kind === 'text' ? '文本模型' : '图片模型').join('、')}已下架，请重新选择`);
      return defaultEnabledModelSelection(next, scriptModels);
    });
  }, [scriptModels, scriptModelsLoaded]);
  const scriptImageProviderOptions = useMemo(() => providerOptionsForModels(scriptModels.image, 'image'), [scriptModels.image]);
  const scriptVideoProviderOptions = useMemo(() => providerOptionsForModels(scriptVideoModels, 'video'), [scriptVideoModels]);
  const scriptImageModelOptions = useMemo(() => modelOptionsForProvider(scriptModels.image, 'image', scriptImageProvider), [scriptModels.image, scriptImageProvider]);
  const scriptVideoModelOptions = useMemo(() => modelOptionsForProvider(scriptVideoModels, 'video', scriptVideoProvider), [scriptVideoModels, scriptVideoProvider]);
  const privateComicModelOptions = useMemo(() => ({
    text: scriptModels.text.map(model => ({ value: model.id, label: model.displayName || model.id })),
    image: scriptImageModelOptions
  }), [scriptModels.text, scriptImageModelOptions]);
  function privateComicScopedKey(baseKey) {
    const username = String(draftUsernameRef.current || getCurrentUsername() || 'anonymous').trim().toLowerCase() || 'anonymous';
    return `${baseKey}:${username}`;
  }
  useEffect(() => {
    try {
      const draftKey = privateComicScopedKey(privateComicDraftKey);
      const taskKey = privateComicScopedKey(privateComicTaskKey);
      const saved = JSON.parse(window.localStorage.getItem(draftKey) || 'null');
      if (saved && typeof saved === 'object') setPrivateComicDraft(current => ({ ...current, ...saved, comic: { ...current.comic, ...(saved.comic || {}) } }));
      const savedTask = JSON.parse(window.localStorage.getItem(taskKey) || 'null');
      if (savedTask?.batchId) setPrivateComicTask(savedTask);
      window.localStorage.removeItem(privateComicDraftKey);
      window.localStorage.removeItem(privateComicTaskKey);
    } catch {
      // A malformed local draft should never prevent the page from opening.
    }
  }, []);
  useEffect(() => {
    if (customizationWorkspace !== 'comic' || privateComicCatalog.length) return undefined;
    let active = true;
    Promise.all([listSystemPresetCatalog('script'), listSystemPresetCatalog('batch-factory')])
      .then(results => {
        if (!active) return;
        setPrivateComicCatalog(results.flatMap(result => Array.isArray(result?.catalog) ? result.catalog : []));
      })
      .catch(error => { if (active) setPrivateComicError(error?.message || '画风库和漫画预设读取失败'); });
    return () => { active = false; };
  }, [customizationWorkspace, privateComicCatalog.length]);
  useEffect(() => {
    try {
      window.localStorage.setItem(privateComicScopedKey(privateComicDraftKey), JSON.stringify({ ...privateComicDraft, comic: { ...privateComicDraft.comic } }));
      setPrivateComicSaveState({ status: 'saved', savedAt: new Date(), reason: '' });
    } catch (error) {
      // 配额满 / 隐私模式与剧本草稿同一条失败链路：不抛异常，界面给出可读原因。
      const reason = storageFailureReason(error);
      setPrivateComicSaveState({ status: 'failed', savedAt: null, reason });
      console.error('[私人定制] 漫画草稿保存失败：', reason);
    }
  }, [privateComicDraft]);
  useEffect(() => {
    try {
      if (privateComicTask?.batchId) window.localStorage.setItem(privateComicScopedKey(privateComicTaskKey), JSON.stringify(privateComicTask));
    } catch {
      // Task restoration is best effort and never blocks the current run.
    }
  }, [privateComicTask]);
  useEffect(() => { setSelectedShotIndexes(new Set()); }, [selectedFormat]);
  const rawShotCards = useMemo(() => {
    const parsed = getShotCardsWithinDuration(selectedFormat, output, selectedDuration);
    if (parsed.length) return parsed;
    // 统一 ### 分镜N 协议仍是首选边界；但模型偶发未按协议输出标题时，
    // 仍需把现有输出作为一张展示卡送入最终组装，否则基础设定和已开启约束会全部不可见。
    return output ? [output] : [];
  }, [selectedFormat, selectedDuration, output]);
  const shotCards = useMemo(() => rawShotCards.map((card, index) => buildFinalSegmentCard(card, {
    extractInfo,
    // 卡片必须跟随当前已保存的基础设定和约束开关即时重组，不必再次请求模型。
    constraints: constraintsForFormat(constraints, selectedFormat, extractInfo),
    index,
    duration: selectedDuration
  })), [rawShotCards, extractInfo, constraints, selectedFormat, selectedDuration]);
  const scriptSourceTitle = String(novelText || '').trim().split(/\r?\n/).map(line => line.trim()).find(Boolean) || '';
  const scriptHeaderTitle = scriptSourceTitle ? `《${scriptSourceTitle.replace(/^《|》$/g, '').slice(0, 24)}》` : '未命名剧本';
  const scriptHeaderShotCount = shotCards.length;
  // TODO(总时长)：后端未提供整篇总时长字段。当前优先累加分镜原文里的“总时长：Ns”，
  // 解析不到时按“实际分镜条数 × 当前单卡上限（10s/15s）”估算；两者都拿不到就不显示。
  // 后端补总时长字段后，这里改为直接读取，不再前端估算。
  const scriptCardDurations = shotCards.map(card => Number(String(card || '').match(/总时长[：:]\s*(\d+)s/)?.[1] || 0));
  const scriptHeaderDurationSeconds = shotCards.length
    ? (scriptCardDurations.every(value => value > 0)
      ? scriptCardDurations.reduce((sum, value) => sum + value, 0)
      : shotCards.length * (selectedDuration === '15s' ? 15 : 10))
    : 0;
  const scriptHeaderDurationText = scriptHeaderDurationSeconds
    ? scriptHeaderDurationSeconds >= 60
      ? `预计 ${Math.round(scriptHeaderDurationSeconds / 60)} 分钟`
      : `预计 ${scriptHeaderDurationSeconds} 秒`
    : '';
  const scriptHeaderMeta = novelText.trim()
    ? ['一段内容', `${scriptHeaderShotCount} 个镜头`, scriptHeaderDurationText].filter(Boolean).join(' · ')
    : '等待输入内容';
  const scriptHeaderStatus = generationStage === 'extracting'
    ? '提取中'
    : generationStage === 'generating'
      ? '生成中'
      : generationStage === 'extracted' || generationStage === 'complete'
        ? '准备生成'
        : '待输入';
  const scriptReadCharCount = Array.from(String(novelText || '')).length;
  const scriptEntityCount = extractInfo.characters.length + extractInfo.scenes.length;
  // 步骤态必须由真实字段派生：空正文不得显示“已读取/当前”，否则 0 字也会宣称已读取。
  const hasScriptText = Boolean(novelText.trim());
  const hasScriptOutput = Boolean(output);
  const hasScriptAssets = scriptEntityCount > 0;
  const scriptSteps = [
    { key: 'read', label: hasScriptText || hasScriptOutput ? '内容已读取' : '待读取', state: hasScriptText || hasScriptOutput ? 'done' : 'pending' },
    { key: 'entities', label: '确认人物与场景', state: hasScriptOutput || hasScriptAssets ? 'done' : hasScriptText ? 'active' : 'pending' },
    { key: 'output', label: '生成创作结果', state: hasScriptOutput ? 'done' : hasScriptAssets ? 'active' : 'pending' }
  ];
  const draftSaveLabel = draftSaveState.status === 'saving'
    ? '保存中…'
    : draftSaveState.status === 'saved'
      ? `已保存 · ${formatDraftSavedTime(draftSaveState.savedAt) || '—'}`
      : draftSaveState.status === 'failed'
        ? '保存失败'
        : '尚未保存';
  const scriptAssetChips = [
    ...extractInfo.characters.map(item => ({ key: `character-${item.id}`, label: entityName(item) || '未命名', suffix: '人物', title: `点击编辑：${entityName(item) || '未命名'}（人物）`, onClick: () => openEntityEditor('characters', item.id) })),
    ...extractInfo.scenes.map(item => ({ key: `scene-${item.id}`, label: entityName(item) || '未命名', suffix: '场景', title: `点击编辑：${entityName(item) || '未命名'}（场景）`, onClick: () => openEntityEditor('scenes', item.id) })),
    ...extractInfo.scenes
      .filter(item => String(item.data?.['氛围'] || '').trim())
      .map(item => ({ key: `atmosphere-${item.id}`, label: String(item.data['氛围']).trim(), suffix: '氛围', title: `点击编辑：${String(item.data['氛围']).trim()}（氛围，所属场景）`, onClick: () => openEntityEditor('scenes', item.id) }))
  ];
  const shotCardStarts = useMemo(() => getShotCardStarts(output, rawShotCards), [output, rawShotCards]);
  const selectedShotMatches = useMemo(
    () => getSelectedShotMatches(output, rawShotCards, selectedShotIndexes, shotFindText),
    [output, rawShotCards, selectedShotIndexes, shotFindText]
  );
  const activeShotMatch = selectedShotMatches[shotMatchIndex] || null;
  useEffect(() => {
    setShotMatchIndex(index => selectedShotMatches.length ? Math.min(index, selectedShotMatches.length - 1) : 0);
  }, [selectedShotMatches.length]);
  const isShotCardView = shotCards.length > 0 && !editingOutput;
  const workbenchRef = useRef(null);
  const draftReadyRef = useRef(false);
  const draftSaveTimerRef = useRef(null);
  const draftUsernameRef = useRef(getCurrentUsername());
  const draftTabIdRef = useRef(getScriptDraftTabId(window.sessionStorage));
  const requestGenerationRef = useRef({ workflow: 0, narrate: 0 });
  const sourceGenerationRef = useRef(0);
  const mountedRef = useRef(true);
  const sourceAudioUrlRef = useRef('');

  function beginRequest(kind) {
    return {
      kind,
      requestId: ++requestGenerationRef.current[kind],
      sourceGeneration: sourceGenerationRef.current
    };
  }

  function isCurrentRequest(request) {
    return mountedRef.current
      && requestGenerationRef.current[request.kind] === request.requestId
      && sourceGenerationRef.current === request.sourceGeneration;
  }

  function invalidateRequests() {
    sourceGenerationRef.current += 1;
  }

  function clearNovelText() {
    form.setFieldValue('novelText', '');
    persistDraft({ ...form.getFieldsValue(), novelText: '' });
  }

  function openAudioMatchSettings() {
    const source = String(form.getFieldValue('novelText') || '').trim();
    if (!source) return message.warning('请先粘贴小说原文');
    setQuickDirectorOpen(true);
  }

  function replaceSourceAudio(nextUrl) {
    if (sourceAudioUrlRef.current) URL.revokeObjectURL(sourceAudioUrlRef.current);
    sourceAudioUrlRef.current = nextUrl;
    setSourceAudioUrl(nextUrl);
    setSourceAudioDurationSeconds(null);
    setSourceAudioError('');
  }

  function snapshotDraft(values = form.getFieldsValue()) {
    return {
      values: {
        mode: values.mode || 'continuous',
        format: values.format || 'storyboard',
        duration: values.duration || '10s',
        novelText: values.novelText || '',
        extractionPreset: values.extractionPreset || 'standard'
      },
      quickDirectorOptions,
      constraints: constraintsForNextGeneration(constraints),
      outputConstraints: normalizeScriptConstraints(outputConstraints),
      extractInfo,
      output,
      editingOutput,
      shotVideoTasks,
      videoTaskHistory: shotVideoTaskHistory,
      shotReferenceStates,
      scriptModelSelection,
      generationStage: generationStage === 'extracting' || generationStage === 'generating' ? 'idle' : generationStage
    };
  }

  function persistDraft(values, overrides = {}, options = {}) {
    if (!draftReadyRef.current) return;
    const write = () => {
      const result = saveScriptDraft(window.localStorage, draftUsernameRef.current, draftTabIdRef.current, {
        ...snapshotDraft(values),
        ...overrides
      });
      if (result.ok) {
        setDraftSaveState({ status: 'saved', savedAt: result.savedAt, reason: '' });
        return;
      }
      setDraftSaveState({ status: 'failed', savedAt: null, reason: result.reason });
      console.error('[私人定制] 草稿保存失败：', result.reason);
    };
    if (draftSaveTimerRef.current) window.clearTimeout(draftSaveTimerRef.current);
    draftSaveTimerRef.current = null;
    if (options.immediate) {
      write();
      return;
    }
    setDraftSaveState(current => current.status === 'saving' ? current : { status: 'saving', savedAt: current.savedAt, reason: '' });
    draftSaveTimerRef.current = window.setTimeout(write, 300);
  }

  function updateOutputDraft(nextOutput, preserveSelectedShots = false) {
    setOutput(nextOutput);
    if (!preserveSelectedShots) setSelectedShotIndexes(new Set());
    else {
      const nextCards = getShotCardsWithinDuration(selectedFormat, nextOutput, selectedDuration);
      setSelectedShotIndexes(current => new Set([...current].filter(index => index < nextCards.length)));
    }
    persistDraft(undefined, { output: nextOutput });
  }

  function syncShotMentionSelection(text, start, end = start, anchorRect = null) {
    const selection = { start: Number.isInteger(start) ? start : String(text || '').length, end: Number.isInteger(end) ? end : start };
    setEditingShotSelection(selection);
    setMentionMenuOpen(Boolean(findActiveShotMention(text, selection.start)));
    setShotMentionAnchorRect(anchorRect);
    setMentionActiveIndex(0);
  }

  function syncOutputMentionSelection(text, start, end = start, anchorRect = null) {
    const selection = { start: Number.isInteger(start) ? start : String(text || '').length, end: Number.isInteger(end) ? end : start };
    setEditingOutputSelection(selection);
    setOutputMentionMenuOpen(Boolean(findActiveShotMention(text, selection.start)));
    setOutputMentionAnchorRect(anchorRect);
    setOutputMentionActiveIndex(0);
  }

  function insertOutputMention(candidate) {
    const mention = findActiveShotMention(output, editingOutputSelection.start);
    if (!mention) return;
    const { text: nextText, cursor } = insertActiveShotMention(output, mention, formatEntity(candidate.item));
    updateOutputDraft(nextText);
    setEditingOutputSelection({ start: cursor, end: cursor });
    setOutputMentionMenuOpen(false);
  }

  function toggleOutputEditor() {
    setSelectedShotIndexes(new Set());
    if (editingOutput) {
      setEditingOutput(false);
      setOutputMentionMenuOpen(false);
      return;
    }
    const cursor = output.length;
    setEditingOutputSelection({ start: cursor, end: cursor });
    setOutputMentionMenuOpen(false);
    setOutputMentionActiveIndex(0);
    setEditingOutput(true);
  }

  function openShotEditor(index) {
    const text = rawShotCards[index] || '';
    setEditingShot({ index, text });
    setEditingShotSelection({ start: text.length, end: text.length });
    setMentionMenuOpen(false);
    setMentionActiveIndex(0);
  }

  function closeShotEditor() {
    setEditingShot({ index: -1, text: '' });
    setMentionMenuOpen(false);
    setMentionActiveIndex(0);
  }

  function insertShotMention(candidate) {
    const name = formatEntity(candidate.item);
    const mention = findActiveShotMention(editingShot.text, editingShotSelection.start);
    if (!mention) return;
    const { text: nextText, cursor } = insertActiveShotMention(editingShot.text, mention, name);
    setEditingShot(current => ({ ...current, text: nextText }));
    setEditingShotSelection({ start: cursor, end: cursor });
    setMentionMenuOpen(false);
  }

  function replaceCurrentShotMatch() {
    if (!activeShotMatch) return;
    const nextOutput = replaceSelectedShotMatch(output, activeShotMatch, shotReplaceText);
    updateOutputDraft(nextOutput, true);
  }

  function replaceAllShotMatches() {
    if (!selectedShotMatches.length) return;
    const nextOutput = replaceAllSelectedShotMatches(output, selectedShotMatches, shotReplaceText);
    updateOutputDraft(nextOutput, true);
  }

  async function copyText(text) {
    if (!text) return message.warning('没有可复制的内容');
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
      } else {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.setAttribute('readonly', '');
        textarea.style.cssText = 'position:fixed;opacity:0;pointer-events:none;';
        document.body.appendChild(textarea);
        textarea.select();
        const copied = document.execCommand('copy');
        textarea.remove();
        if (!copied) throw new Error('copy failed');
      }
      message.success('已复制');
    } catch {
      message.error('复制失败，请检查浏览器权限');
    }
  }

  function exportOutput() {
    const content = String((isShotCardView ? shotCards.join('\n\n') : output) || '').trim();
    if (!content) return message.warning('没有可导出的内容');
    const title = String(scriptHeaderTitle || '剧本').trim() || '剧本';
    const filename = `${title.replace(/[\\/:*?"<>|]/g, '_')}.txt`;
    const blob = new Blob([`${title}\n\n${content}`], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    setTimeout(() => URL.revokeObjectURL(url), 0);
    message.success('剧本已导出为 TXT 文件');
  }

  async function generateVideoForShot(card, index, { withoutReferences = false } = {}) {
    const prompt = String(card || '').trim();
    if (!prompt) return message.warning('该分镜没有可生成的视频提示词');
    if (scriptVideoProvider === 'doubao_local_executor' && !localExecutorStatus.online) message.warning('豆包执行器未在线，任务仍会提交并显示执行结果。');
    const fallbackDuration = selectedDuration === '15s' ? 15 : 10;
    const modelUsesShotDuration = scriptVideoModelKey === 'minimax-h3-video' || scriptVideoModelKey === 'seedance-2-0-official';
    const resolvedDuration = modelUsesShotDuration
      ? resolveShotVideoDuration({
        shotText: prompt,
        fallbackDuration,
        ...(scriptVideoModelKey === 'seedance-2-0-official' ? { minDuration: 4, modelLabel: 'Seedance' } : {})
      })
      : { ok: true, duration: fallbackDuration };
    if (!resolvedDuration.ok) {
      message.error(resolvedDuration.error);
      return;
    }
    if (scriptVideoModelKey === 'yd2-mini-video') {
      const cardDuration = resolveShotVideoDuration({ shotText: prompt, fallbackDuration });
      if (!cardDuration.ok) {
        message.error(cardDuration.error);
        return;
      }
      if (cardDuration.duration !== 10) {
        message.warning('YD2.0 Mini 固定生成 10 秒；请切换 H3 或 Seedance 后再按当前分镜时长生成');
        return;
      }
    }
    setGeneratingShotIndexes(current => new Set([...current, index]));
    try {
      const historyId = await ensureCurrentHistory();
      const videoPayload = buildScriptVideoPayload({
        prompt,
        modelKey: scriptVideoProvider === 'doubao_local_executor' ? 'local-doubao-executor-video' : scriptVideoModelKey,
        model: scriptVideoProvider === 'doubao_local_executor' ? scriptVideoModelKey : '',
        duration: resolvedDuration.duration,
        resolution: scriptVideoResolution,
        aspectRatio: scriptVideoAspectRatio,
        imageUrls: withoutReferences ? [] : collectShotReferenceImages({ shotText: prompt, extractInfo, shotIndex: index, shotReferenceStates })
      });
      const result = await createScriptVideo(videoPayload);
      const previousTask = shotVideoTasks[index];
      const nextVideoTaskHistory = appendShotVideoTaskHistory(shotVideoTaskHistory, index, previousTask);
      const nextVideoTasks = { ...shotVideoTasks, [index]: { taskId: result.taskId, status: 'processing', prompt } };
      setShotVideoTasks(nextVideoTasks);
      setShotVideoTaskHistory(nextVideoTaskHistory);
      if (historyId) updateHistoryVideoTasks(historyId, nextVideoTasks, nextVideoTaskHistory).catch(() => {});
      watchShotVideoTask(index, result.taskId);
      message.success(`已提交第 ${index + 1} 条分镜的视频任务（任务 ID：${result.taskId}）`);
    } catch (error) {
      if (error.status === 409 && error.message === '当前视频模型不支持参考图，是否允许无参考图生成' && !withoutReferences) {
        Modal.confirm({
          title: '参考图不可用',
          content: '当前视频模型不支持参考图，是否允许无参考图生成',
          okText: '允许无参考图生成',
          cancelText: '取消生成',
          onOk: () => generateVideoForShot(card, index, { withoutReferences: true })
        });
      } else {
        message.error(error.message || '视频任务提交失败');
      }
    } finally {
      setGeneratingShotIndexes(current => {
        const next = new Set(current);
        next.delete(index);
        return next;
      });
    }
  }

  async function ensureCurrentHistory() {
    if (currentHistoryId) return currentHistoryId;
    if (!output) return '';
    const values = form.getFieldsValue();
    const historyId = 'react-' + Date.now().toString(36);
    await saveHistory({
      id: historyId,
      mode: values.mode || 'continuous',
      format: values.format || 'storyboard',
      formatName: { storyboard: '画布模式', shortdrama: '剧本模式', screenplay: '剧情模式', shotlist: '分镜模式', q版: 'Q版模式' }[values.format] || '剧本',
      duration: values.duration || '10s',
      output,
      novelText: values.novelText || '',
      extractInfo,
      constraints: constraintsForFormat(outputConstraints, values.format, extractInfo),
      textModelId: scriptModelSelection.textModelId,
      imageModelId: scriptModelSelection.imageModelId,
      videoTasks: shotVideoTasks,
      videoTaskHistory: shotVideoTaskHistory,
      shotReferenceStates
    });
    setCurrentHistoryId(historyId);
    return historyId;
  }

  function watchShotVideoTask(index, taskId) {
    const poll = async () => {
      try {
        const task = await getScriptVideoTask(taskId);
        if (task.status === 'succeeded') {
          setShotVideoTasks(current => current[index]?.taskId === taskId
            ? { ...current, [index]: { ...task, prompt: current[index].prompt } }
            : current);
          message.success(`第 ${index + 1} 条分镜视频生成成功`);
          return;
        }
        if (task.status === 'failed') {
          setShotVideoTasks(current => current[index]?.taskId === taskId
            ? { ...current, [index]: { ...task, prompt: current[index].prompt } }
            : current);
          message.error(task.error || `第 ${index + 1} 条分镜视频生成失败`);
          return;
        }
      } catch (error) {
        // 上游短暂不可用时继续轮询，避免误判为任务失败。
      }
      if (mountedRef.current) window.setTimeout(poll, 4000);
    };
    window.setTimeout(poll, 2000);
  }

  async function openVideoHistory(shotIndex, tasks) {
    setPreviewVideoHistory({ open: true, shotIndex, tasks });
    const refreshedTasks = await Promise.all(tasks.map(async task => {
      try {
        return mergeShotVideoTaskHistoryTask(task, await getScriptVideoTask(task.taskId));
      } catch {
        return { ...task, historyStatusReadError: true };
      }
    }));
    if (!mountedRef.current) return;
    setPreviewVideoHistory(current => current.open && current.shotIndex === shotIndex
      ? { ...current, tasks: refreshedTasks }
      : current);
    const nextHistory = { ...shotVideoTaskHistory, [shotIndex]: refreshedTasks };
    setShotVideoTaskHistory(nextHistory);
    persistDraft(undefined, { videoTaskHistory: nextHistory });
    if (currentHistoryId) updateHistoryVideoTasks(currentHistoryId, shotVideoTasks, nextHistory).catch(() => {});
  }

  async function openScriptHistory() {
    setHistoryOpen(true); setHistoryLoading(true);
    try { const data = await listHistory(); setHistoryEntries(Array.isArray(data?.entries) ? data.entries : []); }
    catch (error) { message.error(error.message || '历史记录加载失败'); }
    finally { setHistoryLoading(false); }
  }

  function restoreHistory(entry) {
    if (!entry?.output) return;
    form.setFieldsValue({ mode: entry.mode || 'continuous', format: entry.format || 'storyboard', duration: entry.duration || '10s', novelText: entry.novelText || '' });
    setExtractInfo(normalizeExtractInfo(entry.extractInfo));
    const restoredConstraints = normalizeScriptConstraints(entry.constraints || DEFAULT_SCRIPT_CONSTRAINTS);
    setConstraints(restoredConstraints);
    setOutputConstraints(restoredConstraints);
    setDraftConstraints(restoredConstraints);
    setOutput(entry.output); setEditingOutput(false); setGenerationStage('complete');
    setShotVideoTasks(entry.videoTasks || {});
    setShotVideoTaskHistory(normalizeShotVideoTaskHistory(entry.videoTaskHistory));
    setShotReferenceStates(entry.shotReferenceStates || {});
    const restoredSelection = { textModelId: entry.textModelId || '', imageModelId: entry.imageModelId || '' };
    if (!scriptModelsLoaded) {
      setScriptModelSelection(restoredSelection);
    } else {
      const restoredModels = reconcileScriptModelSelection(restoredSelection, scriptModels);
      setScriptModelSelection({ textModelId: restoredModels.textModelId, imageModelId: restoredModels.imageModelId });
      if (restoredModels.unavailableKinds.length) message.warning('历史记录中的模型已下架，请重新选择');
    }
    setCurrentHistoryId(entry.id); setHistoryOpen(false);
    Object.entries(entry.videoTasks || {}).forEach(([index, task]) => { if (task.status === 'processing') watchShotVideoTask(Number(index), task.taskId); });
  }

  useEffect(() => {
    const restoredDraft = loadScriptDraft(window.localStorage, draftUsernameRef.current, draftTabIdRef.current);
    if (restoredDraft) {
      form.setFieldsValue(restoredDraft.values);
      setExtractInfo(normalizeExtractInfo(restoredDraft.extractInfo));
      setOutput(typeof restoredDraft.output === 'string' ? restoredDraft.output : '');
      setEditingOutput(Boolean(restoredDraft.editingOutput));
      setGenerationStage(restoredDraft.generationStage || 'idle');
      if (restoredDraft.quickDirectorOptions && typeof restoredDraft.quickDirectorOptions === 'object') {
        setQuickDirectorOptions(current => ({ ...current, ...restoredDraft.quickDirectorOptions }));
      }
      setConstraints(normalizeScriptConstraints(restoredDraft.constraints));
      setDraftConstraints(normalizeScriptConstraints(restoredDraft.constraints));
      setOutputConstraints(normalizeScriptConstraints(restoredDraft.outputConstraints || restoredDraft.constraints));
      setShotVideoTasks(restoredDraft.shotVideoTasks || {});
      setShotVideoTaskHistory(normalizeShotVideoTaskHistory(restoredDraft.videoTaskHistory));
      setShotReferenceStates(restoredDraft.shotReferenceStates || {});
      setScriptModelSelection(restoredDraft.scriptModelSelection || { textModelId: '', imageModelId: '' });
      Object.entries(restoredDraft.shotVideoTasks || {}).forEach(([index, task]) => {
        if (task.status === 'processing') watchShotVideoTask(Number(index), task.taskId);
      });
    }

    // Wait for restored React state to commit before allowing auto-save.
    const readyTimer = window.setTimeout(() => {
      draftReadyRef.current = true;
      if (restoredDraft) {
        const result = saveScriptDraft(window.localStorage, draftUsernameRef.current, draftTabIdRef.current, restoredDraft);
        setDraftSaveState(result.ok
          ? { status: 'saved', savedAt: result.savedAt, reason: '' }
          : { status: 'failed', savedAt: null, reason: result.reason });
        if (!result.ok) console.error('[私人定制] 草稿保存失败：', result.reason);
      } else {
        persistDraft();
      }
    }, 0);

    return () => {
      window.clearTimeout(readyTimer);
      draftReadyRef.current = false;
    };
  }, [form]);

  useEffect(() => {
    if (!draftReadyRef.current) return;
    persistDraft();
  }, [extractInfo, output, editingOutput, generationStage, constraints, quickDirectorOptions, shotVideoTasks, shotVideoTaskHistory, shotReferenceStates, scriptModelSelection]);

  useEffect(() => {
    if (currentHistoryId) updateHistoryVideoTasks(currentHistoryId, shotVideoTasks, shotVideoTaskHistory).catch(() => {});
  }, [currentHistoryId, shotVideoTasks, shotVideoTaskHistory]);

  useEffect(() => {
    let active = true;
    getConfig()
      .then(config => {
        if (!active) return;
        setSoundEnabled(config.notifications?.soundEnabled !== false);
        setSoundVolume(Number.isFinite(config.notifications?.soundVolume) ? config.notifications.soundVolume : 60);
        const defaults = normalizeScriptUnifiedConfig(config.scriptDefaults);
        setScriptModelSelection(current => ({
          textModelId: defaults.textModelId || current.textModelId,
          imageModelId: defaults.imageModelId || current.imageModelId
        }));
        setScriptImageProvider(defaults.imageProvider || 'personal_api');
        setScriptVideoProvider(defaults.videoProvider || 'personal_api');
        setScriptVideoModelKey(defaults.videoModelKey);
        setScriptImageAspectRatio(defaults.imageAspectRatio);
        setScriptVideoAspectRatio(defaults.videoAspectRatio);
        setScriptVideoResolution(defaults.videoResolution);
      })
      .catch(() => {});
    const handleNotificationsUpdated = event => {
      setSoundEnabled(event.detail?.soundEnabled !== false);
      setSoundVolume(Number.isFinite(event.detail?.soundVolume) ? event.detail.soundVolume : 60);
    };
    window.addEventListener('qiantie:notifications-updated', handleNotificationsUpdated);
    return () => {
      active = false;
      window.removeEventListener('qiantie:notifications-updated', handleNotificationsUpdated);
    };
  }, []);

  function openScriptUnifiedConfig() {
    setScriptConfigDraft(normalizeScriptUnifiedConfig({
      textModelId: scriptModelSelection.textModelId,
      imageProvider: scriptImageProvider,
      imageModelId: scriptModelSelection.imageModelId,
      videoProvider: scriptVideoProvider,
      videoModelKey: scriptVideoModelKey,
      imageAspectRatio: scriptImageAspectRatio,
      videoAspectRatio: scriptVideoAspectRatio,
      videoResolution: scriptVideoResolution
    }));
    setScriptConfigOpen(true);
  }

  async function saveScriptUnifiedConfig() {
    const next = normalizeScriptUnifiedConfig(scriptConfigDraft);
    setSavingScriptConfig(true);
    try {
      const saved = await saveConfig({ scriptDefaults: next });
      const defaults = normalizeScriptUnifiedConfig(saved.scriptDefaults || next);
      setScriptModelSelection({ textModelId: defaults.textModelId, imageModelId: defaults.imageModelId });
      setScriptImageProvider(defaults.imageProvider || 'personal_api');
      setScriptVideoProvider(defaults.videoProvider || 'personal_api');
      setScriptVideoModelKey(defaults.videoModelKey);
      setScriptImageAspectRatio(defaults.imageAspectRatio);
      setScriptVideoAspectRatio(defaults.videoAspectRatio);
      setScriptVideoResolution(defaults.videoResolution);
      setScriptConfigOpen(false);
      message.success('剧本创作统一配置已保存');
    } catch (error) {
      message.error(error?.message || '保存统一配置失败');
    } finally { setSavingScriptConfig(false); }
  }

  useEffect(() => {
    let active = true;
    listScriptPresetCatalog()
      .then(result => {
        if (!active) return;
        const catalog = Array.isArray(result?.catalog) ? result.catalog : [];
        setConstraintCatalog(catalog);
        const presets = filterExtractionPresets(catalog);
        setExtractionPresets(presets);
        const selected = selectAvailableExtractionPreset(form.getFieldValue('extractionPreset'), presets);
        if (selected) form.setFieldValue('extractionPreset', selected);
        setExtractionPresetError(selected ? '' : '暂无已发布的提取指令');
      })
      .catch(() => {
        if (!active) return;
        setConstraintCatalog([]);
        setExtractionPresets([]);
        setExtractionPresetError('提取指令加载失败，请刷新或联系管理员');
        message.warning('系统预设加载失败。');
      })
      .finally(() => { if (active) setLoadingExtractionPresets(false); });
    return () => { active = false; };
  }, [form]);

  useEffect(() => {
    const saveBeforeLeave = () => persistDraft(undefined, {}, { immediate: true });
    window.addEventListener('pagehide', saveBeforeLeave);
    return () => {
      window.removeEventListener('pagehide', saveBeforeLeave);
      if (draftSaveTimerRef.current) window.clearTimeout(draftSaveTimerRef.current);
    };
  }, [extractInfo, output, editingOutput, generationStage]);

  function syncPetContext(novelText = form.getFieldValue('novelText')) {
    const characterCount = extractInfo.characters.length;
    const sceneCount = extractInfo.scenes.length;
    const stageLabel = {
      idle: '等待提取人物与场景',
      extracting: '正在提取人物与场景',
      extracted: '已提取，等待确认后生成剧本',
      generating: '正在生成剧本',
      complete: '剧本已生成',
      error: '上一次处理失败'
    }[generationStage] || '等待处理';
    dispatchPetContext({
      page: '剧本生成',
      pagePath: '/script',
      mode: selectedMode || 'continuous',
      summary: `当前阶段：${stageLabel}；人物 ${characterCount} 个；场景 ${sceneCount} 个；剧本结果：${output.trim() ? '已生成' : '未生成'}`,
      entities: {
        characterCount,
        sceneCount,
        generationStage,
        hasOutput: Boolean(output.trim()),
        characterNames: extractInfo.characters.map(formatEntity).filter(Boolean).slice(0, 6).join('、'),
        constraintModalOpen
      },
      actions: generationStage === 'extracted'
        ? ['检查人物与场景', '编辑人物与场景', '生成剧本']
        : ['提取人物与场景', '检查当前剧本', '生成剧本'],
      novelText: String(novelText || ''),
      extracted: extractInfo,
      scriptOutput: output
    });
  }

  useEffect(() => {
    syncPetContext();
  }, [extractInfo, output, generationStage, constraintModalOpen]);

  useEffect(() => {
    function openRevisionPreview(event) {
      const candidateOutput = String(event.detail?.candidateOutput || '').trim();
      if (!output.trim() || !candidateOutput) return;
      setRevisionPreview({
        open: true,
        summary: String(event.detail?.summary || '').trim(),
        currentOutput: output,
        candidateOutput
      });
    }
    window.addEventListener(PET_PREVIEW_EVENT, openRevisionPreview);
    return () => window.removeEventListener(PET_PREVIEW_EVENT, openRevisionPreview);
  }, [output]);

  useEffect(() => () => {
    mountedRef.current = false;
    invalidateRequests();
    if (sourceAudioUrlRef.current) URL.revokeObjectURL(sourceAudioUrlRef.current);
    sourceAudioUrlRef.current = '';
  }, []);

  function addTextFile(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      const novelText = String(reader.result || '').trim();
      form.setFieldValue('novelText', novelText);
      persistDraft({ ...form.getFieldsValue(), novelText });
      syncPetContext(novelText);
      message.success('TXT 原文已添加');
    };
    reader.onerror = () => message.error('TXT 文件读取失败');
    reader.readAsText(file);
    event.target.value = '';
  }

  async function narrateSource() {
    const input = String(form.getFieldValue('novelText') || '').trim();
    if (!input) return message.warning('请先输入或添加小说原文');
    const requestId = beginRequest('narrate');
    setNarrating(true);
    setSourceAudioError('');
    dispatchPetState('working', { title: '原文配音正在生成' });
    try {
      const config = await getConfig();
      const blob = await textToSpeech({ input, ...(config.tts || {}) });
      if (!isCurrentRequest(requestId)) return;
      const nextAudioUrl = URL.createObjectURL(blob);
      if (!isCurrentRequest(requestId)) {
        URL.revokeObjectURL(nextAudioUrl);
        return;
      }
      replaceSourceAudio(nextAudioUrl);
      try {
        const duration = await readAudioDurationFromUrl(nextAudioUrl);
        if (!isCurrentRequest(requestId)) return;
        setSourceAudioDurationSeconds(normalizeAudioDurationSeconds(duration));
        message.success(`已生成原文配音，已读取 ${duration} 秒`);
        dispatchPetState('success', { title: '原文配音已生成', detail: `当前配音时长 ${duration} 秒，可用于匹配音频。` });
      } catch (error) {
        if (!isCurrentRequest(requestId)) return;
        message.warning(error.message || '配音已生成，但未能读取音频时长；匹配音频需要重新生成配音。');
        dispatchPetState('success', { title: '原文配音已生成', detail: '音频可以播放，但暂未读取到时长。' });
      }
    } catch (error) {
      if (!isCurrentRequest(requestId)) return;
      setSourceAudioError(error.message || '原文配音失败，请检查配音配置后重试');
      message.error(error.message || '原文配音失败');
      dispatchPetState('error', { title: '原文配音生成失败', detail: error.message || '请检查配音模型配置后重试。' });
    } finally {
      if (isCurrentRequest(requestId)) setNarrating(false);
    }
  }

  function handleResizeStart(event) {
    if (event.button !== 0 || !workbenchRef.current) return;

    event.preventDefault();
    const workbench = workbenchRef.current;

    function resize(moveEvent) {
      const rect = workbench.getBoundingClientRect();
      const minLeft = 320;
      const minRight = 460;
      const dividerWidth = 8;
      const nextWidth = Math.min(
        rect.width - minRight - dividerWidth,
        Math.max(minLeft, moveEvent.clientX - rect.left)
      );
      setLeftPanelWidth(nextWidth);
    }

    function finishResize() {
      window.removeEventListener('pointermove', resize);
      window.removeEventListener('pointerup', finishResize);
      document.body.classList.remove('is-resizing');
    }

    document.body.classList.add('is-resizing');
    window.addEventListener('pointermove', resize);
    window.addEventListener('pointerup', finishResize, { once: true });
  }

  async function extractEntities(novelText) {
    const extractionPreset = selectAvailableExtractionPreset(form.getFieldValue('extractionPreset'), extractionPresets);
    if (!extractionPreset) throw new Error(extractionPresetError || '暂无已发布的提取指令');
    const extractResponse = await extractCharactersAndScenes(novelText, extractionPreset, scriptModelSelection.textModelId);
    const extraction = normalizeExtraction(extractJSON(aiText(extractResponse)));
    if (!extraction.characters.length && !extraction.scenes.length) {
      throw new Error('模型未返回人物或场景，请检查提取模板或重试');
    }
    return extraction;
  }

  async function handleExtract(values) {
    const requestId = beginRequest('workflow');
    setExtractInfo(normalizeExtractInfo());
    setOutput('');
    setShotVideoTasks({});
    setShotReferenceStates({});
    setSelectedShotIndexes(new Set());
    setEditingOutput(false);
    setExtracting(true);
    setGenerationStage('extracting');
    dispatchPetState('working', { title: '人物与场景正在提取' });
    try {
      const extraction = await extractEntities(values.novelText);
      if (!isCurrentRequest(requestId)) return;
      setExtractInfo(extraction);
      setGenerationStage('extracted');
      message.success(`已提取 ${extraction.characters.length} 个人物和 ${extraction.scenes.length} 个场景`);
      playTaskSound('success', soundEnabled, soundVolume);
      dispatchPetState('success', { title: '人物与场景提取完成', detail: `已提取 ${extraction.characters.length} 个人物和 ${extraction.scenes.length} 个场景。` });
    } catch (error) {
      if (!isCurrentRequest(requestId)) return;
      setGenerationStage('error');
      playTaskSound('warning', soundEnabled, soundVolume);
      dispatchPetState('error', { title: '人物与场景提取失败', detail: error.message || '请检查文本或提取模型后重试。' });
    } finally {
      if (isCurrentRequest(requestId)) setExtracting(false);
    }
  }

  async function regenerateEntities() {
    const novelText = String(form.getFieldValue('novelText') || '').trim();
    if (!novelText) return message.warning('请先输入或添加小说原文');
    const requestId = beginRequest('workflow');
    setRegeneratingEntities(true);
    setGenerationStage('extracting');
    dispatchPetState('working', { title: '人物与场景正在重新提取' });
    try {
      const extraction = await extractEntities(novelText);
      if (!isCurrentRequest(requestId)) return;
      setExtractInfo(extraction);
      setGenerationStage('extracted');
      message.success('人物与场景已重新提取');
      dispatchPetState('success', { title: '人物与场景已重新提取' });
    } catch (error) {
      if (!isCurrentRequest(requestId)) return;
      setGenerationStage('error');
      dispatchPetState('error', { title: '人物与场景重新提取失败', detail: error.message || '请检查模型配置后重试。' });
    } finally {
      if (isCurrentRequest(requestId)) setRegeneratingEntities(false);
    }
  }

  async function generateOutput() {
    const values = form.getFieldsValue();
    if (!extractInfo.characters.length && !extractInfo.scenes.length) return message.warning('请先提取人物与场景');
    if (quickDirectorOptions.matchAudio && !sourceAudioDurationSeconds) {
      return message.warning('请先生成当前原文配音；匹配音频需要读取实际音频时长');
    }
    const requestId = beginRequest('workflow');
    setGenerating(true);
    setGenerationStage('generating');
    dispatchPetState('working', { title: '剧本正在生成' });
    setEditingOutput(false);
    try {
      const entities = toGenerationEntities(extractInfo);
      const requestConstraints = constraintsForFormat(constraints, values.format, extractInfo);
      const scriptResponse = await generateScript({
        mode: values.mode,
        format: values.format,
        duration: values.duration,
        novelText: values.novelText,
        ...entities,
        constraints: requestConstraints,
        matchAudio: quickDirectorOptions.matchAudio,
        audioTotalSeconds: quickDirectorOptions.matchAudio ? sourceAudioDurationSeconds : null,
        textModelId: scriptModelSelection.textModelId
      });
      const nextOutput = aiText(scriptResponse);
      if (typeof nextOutput !== 'string' || !nextOutput.trim()) throw new Error('模型未返回剧本内容');

      if (!isCurrentRequest(requestId)) return;

      setShotVideoTasks({});
      setShotReferenceStates({});
      setOutputConstraints(requestConstraints);
      updateOutputDraft(nextOutput);
      setGenerationStage('complete');
      setCurrentHistoryId('');
      let historySaved = false;
      try {
        const historyId = 'react-' + Date.now().toString(36);
        await saveHistory({
          id: historyId,
          mode: values.mode,
          format: values.format,
          formatName: { storyboard: '画布模式', shortdrama: '剧本模式', screenplay: '剧情模式', shotlist: '分镜模式', q版: 'Q版模式' }[values.format] || '剧本',
          duration: values.duration,
          output: nextOutput,
          novelText: values.novelText,
          extractInfo,
          constraints: requestConstraints,
          textModelId: scriptModelSelection.textModelId,
          imageModelId: scriptModelSelection.imageModelId
        });
        if (!isCurrentRequest(requestId)) return;
        setCurrentHistoryId(historyId);
        historySaved = true;
      } catch {
        if (!isCurrentRequest(requestId)) return;
        message.warning('生成成功，但保存历史失败');
      }
      if (!isCurrentRequest(requestId)) return;
      if (!constraints.enabled) {
        const nextConstraints = constraintsForNextGeneration(constraints);
        setConstraints(nextConstraints);
        setDraftConstraints(nextConstraints);
        persistDraft(undefined, { constraints: nextConstraints, outputConstraints: requestConstraints });
      }
      message.success('生成完成');
      playTaskSound('success', soundEnabled, soundVolume);
      dispatchPetState('success', {
        title: '剧本已生成',
        detail: historySaved ? '结果已保存到当前工作台和生成历史。' : '结果已保存到当前工作台，但生成历史保存失败。'
      });
    } catch (error) {
      if (!isCurrentRequest(requestId)) return;
      setGenerationStage('error');
      message.error(error.message || '生成剧本失败');
      playTaskSound('warning', soundEnabled, soundVolume);
      dispatchPetState('error', { title: '生成剧本失败', detail: error.message || '请检查模型配置后重试。' });
    } finally {
      if (isCurrentRequest(requestId)) setGenerating(false);
    }
  }

  async function regenerateOutput() {
    setRegeneratingOutput(true);
    try {
      await generateOutput();
    } finally {
      setRegeneratingOutput(false);
    }
  }

  function closeRevisionPreview() {
    setRevisionPreview({ open: false, summary: '', currentOutput: '', candidateOutput: '' });
  }

  function applyRevisionPreview() {
    const nextOutput = revisionPreview.candidateOutput.trim();
    if (!nextOutput) return;
    if (revisionPreview.currentOutput !== output) {
      closeRevisionPreview();
      message.warning('当前剧本已变化，请重新让 CM 生成修改稿。');
      return;
    }
    setPreviousOutput(output);
    updateOutputDraft(nextOutput);
    setEditingOutput(true);
    closeRevisionPreview();
    message.success('CM 的修改稿已应用，可撤销一次。');
  }

  function undoLastRevision() {
    if (!previousOutput) return;
    updateOutputDraft(previousOutput);
    setEditingOutput(true);
    setPreviousOutput('');
    message.success('已撤销 CM 的本次修改。');
  }

  function invalidateEntityOutput(nextInfo) {
    setOutput('');
    setShotVideoTasks({});
    setShotReferenceStates({});
    setEditingOutput(false);
    setSelectedShotIndexes(new Set());
    setGenerationStage(nextInfo.characters.length || nextInfo.scenes.length ? 'extracted' : 'idle');
  }

  function openEntityEditor(type, id) {
    entityEditorSessionRef.current += 1;
    setActiveEntity({ type, id, isNew: false, editorSessionId: entityEditorSessionRef.current });
    setFullscreenEditor(false);
  }

  function addEntity(type) {
    const data = type === 'characters'
      ? { 名称: '', 身份: '', 外形: '', 性格: '' }
      : { 名称: '', 时段: '', 氛围: '', 描述: '' };
    const draft = createEntity(data);
    entityEditorSessionRef.current += 1;
    setActiveEntity({ type, id: draft.id, isNew: true, data: draft.data, editorSessionId: entityEditorSessionRef.current });
    setFullscreenEditor(false);
  }

  function updateActiveEntity(fields, media = {}) {
    if (!activeEntity) return;
    setExtractInfo(current => {
      const normalized = normalizeExtractInfo(current);
      const items = [...normalized[activeEntity.type]];
      const currentIndex = items.findIndex(item => item.id === activeEntity.id);
      let dataChanged = activeEntity.isNew;
      if (activeEntity.isNew) items.push({ ...createEntity(fields, activeEntity.id), ...media });
      else if (currentIndex !== -1) {
        dataChanged = JSON.stringify(items[currentIndex].data || {}) !== JSON.stringify(fields || {});
        items[currentIndex] = {
          ...items[currentIndex],
          data: fields,
          imageUrls: Array.isArray(media.imageUrls) ? media.imageUrls : items[currentIndex].imageUrls,
          mainImageUrl: typeof media.mainImageUrl === 'string' ? media.mainImageUrl : items[currentIndex].mainImageUrl
        };
      }
      const next = { ...normalized, [activeEntity.type]: items };
      if (dataChanged) invalidateEntityOutput(next);
      return next;
    });
  }

  function deleteActiveEntity() {
    if (!activeEntity || activeEntity.isNew) return;
    const deletedEntityId = activeEntity.id;
    setExtractInfo(current => {
      const normalized = normalizeExtractInfo(current);
      const next = {
        ...normalized,
        [activeEntity.type]: normalized[activeEntity.type].filter(item => item.id !== activeEntity.id),
        protagonistIds: activeEntity.type === 'characters'
          ? normalized.protagonistIds.filter(id => id !== activeEntity.id)
          : normalized.protagonistIds
      };
      invalidateEntityOutput(next);
      return next;
    });
    setConstraints(current => removeEntityConstraintReferences(current, deletedEntityId));
    setDraftConstraints(current => removeEntityConstraintReferences(current, deletedEntityId));
    setActiveEntity(null);
  }

  function toggleProtagonist(id) {
    setExtractInfo(current => {
      const normalized = normalizeExtractInfo(current);
      const protagonistIds = normalized.protagonistIds.includes(id)
        ? normalized.protagonistIds.filter(item => item !== id)
        : [...normalized.protagonistIds, id];
      const next = { ...normalized, protagonistIds };
      invalidateEntityOutput(next);
      return next;
    });
  }

  function constraintOptions(category) {
    return constraintCatalog
      .filter(item => item.constraintCategory === category)
      .map(item => ({ label: item.name, value: item.id }));
  }

  function constraintPresetBody(category, presetId) {
    if (!constraintCatalog.some(item => item.constraintCategory === category && item.id === presetId)) return '';
    return String(constraintTexts[presetId] || '').trim();
  }

  async function loadConstraintPresetText(category, presetId) {
    if (!presetId || !constraintCatalog.some(item => item.constraintCategory === category && item.id === presetId)) return '';
    if (constraintTexts[presetId]) return String(constraintTexts[presetId]).trim();
    try {
      const result = await getConstraintPresetTexts([presetId]);
      const body = String(result.texts?.[presetId] || '').trim();
      setConstraintTexts(current => ({ ...current, ...(result.texts || {}) }));
      return body;
    } catch {
      message.warning('系统预设提示词读取失败');
      return '';
    }
  }

  async function loadPersonalConstraintPrompts(category) {
    setLoadingPersonalConstraintPrompts(current => ({ ...current, [category]: true }));
    try {
      const result = await listScriptConstraintPrompts(category);
      setPersonalConstraintPrompts(current => ({ ...current, [category]: Array.isArray(result?.prompts) ? result.prompts : [] }));
    } catch {
      message.warning('我的提示词加载失败');
    } finally {
      setLoadingPersonalConstraintPrompts(current => ({ ...current, [category]: false }));
    }
  }

  function updateDraftConstraint(category, patch) {
    setDraftConstraints(current => ({
      ...current,
      [category]: { ...current[category], ...patch }
    }));
  }

  async function selectSystemConstraint(category, presetId) {
    const body = await loadConstraintPresetText(category, presetId);
    updateDraftConstraint(category, { source: 'system', presetId: presetId || '', personalPromptId: '', body });
  }

  function selectPersonalConstraint(category, promptId) {
    const prompt = (personalConstraintPrompts[category] || []).find(item => item.id === promptId);
    if (!prompt) {
      updateDraftConstraint(category, { source: 'draft', personalPromptId: '' });
      return;
    }
    updateDraftConstraint(category, { source: 'personal', presetId: '', personalPromptId: prompt.id, body: prompt.body });
  }

  async function savePersonalConstraint(category, name) {
    const body = String(draftConstraints[category]?.body || '').trim();
    if (!body) return message.warning('请先填写提示词内容');
    if (name !== null && !String(name || '').trim()) return message.warning('请输入提示词名称');
    setSavingConstraintCategory(category);
    try {
      const prompt = editingPersonalPromptId
        ? await updateScriptConstraintPrompt(editingPersonalPromptId, { name, body })
        : await saveScriptConstraintPrompt({ category, name, body });
      const saved = prompt?.prompt;
      if (!saved) throw new Error('保存失败');
      await loadPersonalConstraintPrompts(category);
      updateDraftConstraint(category, { source: 'personal', presetId: '', personalPromptId: saved.id, body: saved.body });
      setEditingPersonalPromptId('');
      setEditingPersonalPromptName('');
      setPersonalPromptNameModal({ open: false, category: '', name: '' });
      message.success(name === null ? '当前草稿已保存' : '已保存到我的提示词');
    } catch (error) {
      message.error(error.message || '提示词保存失败');
    } finally {
      setSavingConstraintCategory('');
    }
  }

  async function deletePersonalConstraint(category, id) {
    try {
      await deleteScriptConstraintPrompt(id);
      await loadPersonalConstraintPrompts(category);
      if (draftConstraints[category]?.personalPromptId === id) updateDraftConstraint(category, { source: 'draft', personalPromptId: '', body: '' });
      message.success('已删除我的提示词');
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  }

  function openConstraints() {
    const next = normalizeScriptConstraints(constraints);
    setDraftConstraints(next);
    ['prefix', 'quality', 'restriction', 'negative'].forEach(category => {
      loadPersonalConstraintPrompts(category);
      const value = next[category];
      if (value.enabled && value.source === 'system') loadConstraintPresetText(category, value.presetId);
    });
    setConstraintModalOpen(true);
  }

  function saveConstraints() {
    const next = normalizeScriptConstraints(draftConstraints);
    setConstraints(next);
    setOutputConstraints(next);
    setConstraintModalOpen(false);
    persistDraft(undefined, { constraints: constraintsForNextGeneration(next) });
  }

  const activeItem = activeEntity?.isNew
    ? activeEntity.data
    : activeEntity ? extractInfo[activeEntity.type].find(item => item.id === activeEntity.id) : null;
  // 草稿恢复或页面切换后阶段标记可能回到 idle，但已提取/手动添加的实体仍然有效。
  // 只要存在人物或场景，就允许继续生成剧本；请求层仍会校验实际输入。
  const canGenerateScript = generationStage === 'extracted'
    || generationStage === 'complete'
    || extractInfo.characters.length > 0
    || extractInfo.scenes.length > 0;
  const extractionPreset = selectAvailableExtractionPreset(form.getFieldValue('extractionPreset'), extractionPresets);
  const selectedExtractionPreset = extractionPresets.find(item => item.id === extractionPreset);
  const extractionPresetName = selectedExtractionPreset?.name || extractionPresetError || '正在加载提取指令';
  const extractionUnavailable = loadingExtractionPresets || Boolean(extractionPresetError) || !extractionPresets.length;

  useScriptCmBridge({
    form,
    extractInfo,
    setExtractInfo,
    output,
    updateOutputDraft,
    setPreviousOutput,
    setEditingOutput,
    constraints,
    setConstraints,
    setDraftConstraints,
    generationStage,
    invalidateEntityOutput
  });

  useEffect(() => {
    if (!activeEntity || activeEntity.isNew) {
      dispatchCmSelection(null);
      return;
    }
    const item = extractInfo[activeEntity.type]?.find(candidate => candidate.id === activeEntity.id);
    dispatchCmSelection(scriptEntitySelection(activeEntity.type, item));
  }, [activeEntity, extractInfo]);

  useEffect(() => {
    if (!constraintModalOpen) {
      if (!activeEntity || activeEntity.isNew) dispatchCmSelection(null);
      else {
        const item = extractInfo[activeEntity.type]?.find(candidate => candidate.id === activeEntity.id);
        dispatchCmSelection(scriptEntitySelection(activeEntity.type, item));
      }
      return;
    }
    const categoryLabels = {
      baseSetup: '基础设定（人物 / 场景）',
      prefix: '画面前缀词',
      quality: '画质约束',
      restriction: '画面限制',
      negative: '负面提示词'
    };
    const category = categoryLabels[activeConstraintCategory] ? activeConstraintCategory : 'prefix';
    const current = draftConstraints[category] || {};
    dispatchCmSelection({
      type: 'constraint',
      id: category,
      label: categoryLabels[category],
      meta: {
        data: {
          category,
          enabled: current.enabled === true,
          source: current.source || 'draft',
          body: String(current.body || '')
        }
      }
    });
  }, [constraintModalOpen, activeConstraintCategory, draftConstraints, activeEntity, extractInfo]);

  function privateComicBatchValue(result) {
    return result?.batch || result?.data?.batch || result?.data || result;
  }

  function privateComicRuntimeTask(batch, runtime, current = {}) {
    const automation = runtime?.automation || {};
    const stage = Object.values(runtime?.stageSummaries || {})[0] || {};
    const total = Number(stage.comicImageTotalCount || stage.comicCollectionPageCount || 0);
    const ready = Number(stage.comicImageReadyCount || stage.comicCollectionPageCount || 0);
    const progress = total > 0 ? Math.round((ready / total) * 100) : null;
    const state = String(automation.state || '').toLowerCase();
    const failed = ['failed', 'error', 'stopped'].includes(state) || Boolean(stage.lastFailed);
    const complete = stage.comicCollectionReady === true || ['succeeded', 'completed', 'success'].includes(state);
    return {
      ...current,
      batchId: current.batchId || batch?.id || '',
      bookId: current.bookId || batch?.books?.[0]?.id || '',
      status: failed ? 'failed' : complete ? 'succeeded' : (state || current.status || 'running'),
      progress,
      message: automation.message || stage.lastFailed?.message || stage.runs?.at?.(-1)?.message || '',
      error: automation.error || stage.lastFailed?.error || ''
    };
  }

  function reloadPrivateComicCatalog() {
    setPrivateComicError('');
    Promise.all([listSystemPresetCatalog('script'), listSystemPresetCatalog('batch-factory')])
      .then(results => setPrivateComicCatalog(results.flatMap(result => Array.isArray(result?.catalog) ? result.catalog : [])))
      .catch(error => setPrivateComicError(error?.message || '画风库和漫画预设读取失败'));
  }

  async function loadPrivateComicProductionData(batchId, bookId) {
    if (!batchId || !bookId) return;
    try {
      const [assetResult, pageResult] = await Promise.all([
        listBookAssets(batchId, bookId, { suppressGlobalError: true }),
        listComicPages(batchId, bookId, { suppressGlobalError: true })
      ]);
      const assets = Array.isArray(assetResult?.assets) ? assetResult.assets : [];
      const pages = Array.isArray(pageResult?.pages) ? pageResult.pages : [];
      setPrivateComicAssets(assets);
      setPrivateComicPages(pages);
      setPrivateComicReadiness(pageResult?.readiness || null);
      const imageEntries = await Promise.all(assets.map(async asset => {
        try {
          const result = await listBookAssetImages(batchId, bookId, asset.id, { suppressGlobalError: true });
          return [asset.id, Array.isArray(result?.images) ? result.images : []];
        } catch (_) {
          return [asset.id, []];
        }
      }));
      setPrivateComicAssetImages(Object.fromEntries(imageEntries));
    } catch (_) {
      // 资产与分镜的只读补充失败时不覆盖任务主状态。
    }
  }

  async function refreshPrivateComicTask(batchIdOverride = '', currentOverride = null) {
    const batchId = batchIdOverride || privateComicTask?.batchId;
    if (!batchId) return null;
    const requestId = ++privateComicPollRequestRef.current;
    try {
      const [batchResult, runtime] = await Promise.all([
        getBatch(batchId, { cache: 'no-store' }),
        getBatchRuntimeSummary(batchId, { suppressGlobalError: true })
      ]);
      if (requestId !== privateComicPollRequestRef.current) return null;
      const batch = privateComicBatchValue(batchResult);
      const next = privateComicRuntimeTask(batch, runtime, currentOverride || privateComicTask);
      setPrivateComicTask(next);
      setPrivateComicError('');
      if (next.bookId) await loadPrivateComicProductionData(batchId, next.bookId);
      return next;
    } catch (error) {
      if (requestId !== privateComicPollRequestRef.current) return null;
      setPrivateComicError(error?.message || '漫画任务状态暂时读取失败，请稍后刷新');
      return null;
    }
  }

  async function ensureCurrentPrivateComicTask() {
    const ensured = await ensurePrivateComicTask({
      task: privateComicTask,
      draft: {
        ...privateComicDraft,
        imageProvider: privateComicDraft.imageProvider || scriptImageProvider
      },
      createTask: createPrivateComicTask
    });
    if (ensured.created) setPrivateComicTask(ensured.task);
    return ensured;
  }

  async function savePrivateComicSettingsPatch(batchId, settingsPatch) {
    const before = privateComicBatchValue(await getBatch(batchId, { cache: 'no-store' }));
    const expectedRevision = Number(before?.revision || 0);
    await saveBatchSettings(batchId, { expectedRevision, patch: settingsPatch });
  }

  async function runPrivateComicAssets() {
    if (privateComicAction || privateComicBusy) return;
    setPrivateComicAction('assets');
    setPrivateComicError('');
    try {
      const ensured = await ensureCurrentPrivateComicTask();
      await preparePrivateComicTaskSettings({ ensured, saveSettings: savePrivateComicSettingsPatch });
      const { batchId, bookId } = ensured.task;
      await runBookStage(batchId, bookId, 'assets', {
        mode: privateComicAssets.length ? 'force' : 'missing',
        textModelId: privateComicDraft.textModelId || undefined,
        requestId: privateComicOperationId()
      });
      await refreshPrivateComicTask(batchId);
    } catch (error) {
      setPrivateComicError(error?.message || '漫画资产提取失败');
    } finally {
      setPrivateComicAction('');
    }
  }

  async function runPrivateComicAssetImages(assetId = '', useThreeView = true) {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    const ids = assetId ? [assetId] : privateComicAssets
      .filter(asset => !(Array.isArray(privateComicAssetImages?.[asset?.id]) && privateComicAssetImages[asset.id].length))
      .map(asset => asset.id);
    if (!batchId || !bookId || !ids.length || privateComicAction) return;
    setPrivateComicAction('images');
    setPrivateComicActionKey(assetId);
    setPrivateComicAssetImageFailures(current => current.filter(id => !ids.includes(id)));
    setPrivateComicError('');
    try {
      const result = await generateBookAssetImages(batchId, bookId, {
        assetIds: ids,
        modelId: privateComicDraft.imageModelId || undefined,
        provider: privateComicDraft.imageProvider || undefined,
        imageAspectRatio: '9:16',
        useThreeView: Boolean(useThreeView)
      });
      const images = Array.isArray(result?.images) ? result.images : [];
      if (!images.length && result?.taskId) {
        for (let attempt = 0; attempt < 40; attempt += 1) {
          await new Promise(resolve => setTimeout(resolve, 3000));
          const status = await getBookAssetImageGenerationTask(batchId, bookId, result.taskId);
          if (status?.status === 'processing' || status?.status === 'pending') continue;
          if (status?.status === 'failed') throw new Error(status.error || '资产图生成失败');
          break;
        }
      }
      if (result?.warning) {
        const finished = new Set(images.map(image => image.assetId || image.assetID));
        const failed = ids.filter(id => !finished.has(id));
        if (failed.length) setPrivateComicAssetImageFailures(failed);
      }
      await refreshPrivateComicTask(batchId);
    } catch (error) {
      setPrivateComicAssetImageFailures(ids);
      setPrivateComicError(error?.message || '资产图生成失败');
    } finally {
      setPrivateComicAction('');
      setPrivateComicActionKey('');
    }
  }

  async function savePrivateComicAsset(asset, fields) {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || !asset?.id) return false;
    setPrivateComicError('');
    try {
      await updateBookAsset(batchId, bookId, asset.id, {
        name: fields.name,
        formalName: fields.name,
        facts: fields.facts || '',
        relationships: fields.relationships || '',
        appearance: fields.appearance || fields.prompt,
        prompt: fields.prompt,
        imageModelId: fields.imageModelId || '',
        imageAspectRatio: fields.imageAspectRatio || '',
        expectedRevision: Number(asset.revision || 0)
      });
      await loadPrivateComicProductionData(batchId, bookId);
      message.success('当前小说资产已保存。');
      return true;
    } catch (error) {
      setPrivateComicError(error?.message || '保存资产失败');
      return false;
    }
  }

  async function uploadPrivateComicAssetImage(asset, file) {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || !asset?.id || !file) return false;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
      message.error('仅支持 PNG、JPG、WEBP 图片。');
      return false;
    }
    setPrivateComicError('');
    try {
      await uploadBookAssetImage(batchId, bookId, asset.id, await readImageAsDataURL(file));
      await loadPrivateComicProductionData(batchId, bookId);
      message.success('图片版本已保存到当前小说资产。');
      return true;
    } catch (error) {
      setPrivateComicError(error?.message || '上传资产图片失败');
      return false;
    }
  }

  async function setPrivateComicPrimaryAssetImage(asset, image) {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || !asset?.id || !image?.id || image.isPrimary) return false;
    setPrivateComicError('');
    try {
      await setPrimaryBookAssetImage(batchId, bookId, asset.id, image.id);
      await loadPrivateComicProductionData(batchId, bookId);
      message.success('已切换为当前资产主图。');
      return true;
    } catch (error) {
      setPrivateComicError(error?.message || '切换主图失败');
      return false;
    }
  }

  async function runPrivateComicPrompts() {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || privateComicAction) return;
    const pageHasPrompt = page => Boolean(String(page?.comicPrompt || '').trim() || String(page?.imagePrompt || '').trim());
    const promptPageCount = privateComicPages.filter(pageHasPrompt).length;
    const mode = privateComicPages.length && promptPageCount >= privateComicPages.length
      ? 'force'
      : promptPageCount > 0 ? 'missing' : 'bootstrap';
    setPrivateComicAction('prompts');
    setPrivateComicError('');
    try {
      await runBookStage(batchId, bookId, 'comic', {
        mode,
        textModelId: privateComicDraft.textModelId || undefined,
        requestId: privateComicOperationId()
      });
      await refreshPrivateComicTask(batchId);
    } catch (error) {
      setPrivateComicError(error?.message || '分镜提示词生成失败');
    } finally {
      setPrivateComicAction('');
    }
  }

  function privateComicPagePayload(page) {
    return comicPageGenerationPayload({
      imageProvider: privateComicDraft.imageProvider || '',
      imageModelId: page?.modelId || privateComicDraft.imageModelId || '',
      imageAspectRatio: String(page?.aspectRatio || '').trim() || '9:16'
    }, page?.comicPrompt || '', page?.imagePrompt || '');
  }

  async function waitForPrivateComicPageImages(batchId, bookId, videoIds, timeoutMs = 300000) {
    if (!videoIds.length) return true;
    const wanted = new Set(videoIds);
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      await new Promise(resolve => setTimeout(resolve, 3000));
      let pages = [];
      try {
        const result = await listComicPages(batchId, bookId, { suppressGlobalError: true });
        pages = Array.isArray(result?.pages) ? result.pages : [];
      } catch (error) {
        if ([401, 403].includes(Number(error?.status))) throw error;
        continue;
      }
      const withImages = pages.filter(page => wanted.has(page?.videoId) && (Array.isArray(page.images) ? page.images.length : 0));
      if (withImages.length >= wanted.size) {
        await refreshPrivateComicTask(batchId);
        return true;
      }
      const activeStates = ['generating', 'processing', 'running', 'queued'];
      const stillRunning = pages.some(page => wanted.has(page?.videoId) && [page?.status, page?.generationState]
        .some(state => activeStates.includes(String(state || '').trim().toLowerCase())));
      if (!stillRunning && withImages.length) break;
    }
    await refreshPrivateComicTask(batchId);
    return false;
  }

  async function runPrivateComicPageImages(videoId = '') {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || privateComicAction) return;
    const missingPromptPages = privateComicPages.filter(page => page?.videoId && !String(page?.imagePrompt || '').trim());
    if (videoId && missingPromptPages.some(page => page.videoId === videoId)) {
      message.warning('该页缺少图片提示词，请先生成分镜提示词');
      return;
    }
    const runnable = privateComicPages.filter(page => {
      if (!page?.videoId) return false;
      if (!String(page?.imagePrompt || '').trim()) return false;
      if (videoId) return page.videoId === videoId;
      return !(Array.isArray(page.images) ? page.images : []).length;
    });
    if (!runnable.length) {
      if (videoId) return;
      if (missingPromptPages.length) message.warning(`有 ${missingPromptPages.length} 页缺少图片提示词，请先生成分镜提示词`);
      return;
    }
    setPrivateComicAction('pages');
    setPrivateComicActionKey(videoId);
    setPrivateComicError('');
    const submitted = [];
    const conflicts = [];
    const failures = [];
    try {
      for (const page of runnable) {
        try {
          const outcome = await generateComicPageWithConflict(batchId, bookId, page.videoId, privateComicPagePayload(page));
          if (outcome.conflict) conflicts.push(page);
          else submitted.push(page.videoId);
        } catch (error) {
          failures.push({ page, error });
        }
      }
      if (conflicts.length) {
        message.info(`有 ${conflicts.length} 页正在生成中，已启动轮询，完成后自动刷新`);
      }
      const pendingIds = submitted.concat(conflicts.map(page => page.videoId));
      const completed = await waitForPrivateComicPageImages(batchId, bookId, pendingIds);
      if (pendingIds.length && !completed) message.info('分镜生图仍在进行，可稍后点刷新查看结果');
      if (failures.length) {
        setPrivateComicError(`${failures.length} 页分镜生图失败：${comicPageErrorMessage(failures[0].error, '分镜生图')}`);
      }
      if (!videoId && missingPromptPages.length) {
        message.warning(`另有 ${missingPromptPages.length} 页缺少图片提示词，未参与本次出图`);
      }
    } finally {
      setPrivateComicAction('');
      setPrivateComicActionKey('');
    }
  }

  async function savePrivateComicPagePrompts(page, prompts = {}) {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    const videoId = page?.videoId;
    const comicPrompt = String(prompts.comicPrompt || '').trim();
    const imagePrompt = String(prompts.imagePrompt || '').trim();
    if (!comicPrompt || !imagePrompt) {
      message.warning('漫画提示词与图片提示词都不能为空');
      return false;
    }
    if (!batchId || !bookId || !videoId || privateComicAction) return false;
    setPrivateComicAction('save');
    setPrivateComicActionKey(videoId);
    setPrivateComicError('');
    try {
      const result = await updateComicPage(batchId, bookId, videoId, {
        comicPrompt,
        imagePrompt,
        modelId: page.modelId || privateComicDraft.imageModelId || undefined,
        aspectRatio: page.aspectRatio || undefined
      });
      const updated = result?.page;
      if (updated) {
        setPrivateComicPages(current => current.map(item => item.videoId === videoId
          ? { ...item, ...updated, images: Array.isArray(item.images) ? item.images : [] }
          : item));
      }
      message.success('提示词已保存');
      await refreshPrivateComicTask(batchId);
      return true;
    } catch (error) {
      setPrivateComicError(error?.message || '提示词保存失败');
      return false;
    } finally {
      setPrivateComicAction('');
      setPrivateComicActionKey('');
    }
  }

  async function downloadPrivateComicPackage() {
    const batchId = privateComicTask?.batchId;
    const bookId = privateComicTask?.bookId;
    if (!batchId || !bookId || privateComicAction) return;
    setPrivateComicAction('download');
    setPrivateComicError('');
    try {
      const blob = await downloadComicPromptZip(batchId, bookId);
      const objectUrl = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = objectUrl;
      anchor.download = `${privateComicTitle(privateComicDraft.sourceText, privateComicDraft.title) || '漫画'}_全部提示词与图片.zip`;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      setTimeout(() => URL.revokeObjectURL(objectUrl), 1000);
      message.success('已开始下载全部内容');
    } catch (error) {
      setPrivateComicError(error?.message || '打包下载失败');
    } finally {
      setPrivateComicAction('');
    }
  }

  async function submitPrivateComic() {
    if (privateComicBusy) return;
    setPrivateComicError('');
    setPrivateComicBusy(true);
    try {
      const ensured = await ensureCurrentPrivateComicTask();
      const snapshot = { ...ensured.task, status: 'running', error: '' };
      let settingsForStart = ensured.settings;
      const hadPendingReference = Boolean(privateComicPendingReference);
      if (hadPendingReference) {
        const uploaded = await uploadComicGlobalReference(snapshot.batchId, privateComicPendingReference);
        const uploadedReference = uploaded?.reference || {};
        settingsForStart = mergePrivateComicReference(settingsForStart, uploadedReference);
        setPrivateComicDraft(current => ({
          ...current,
          comic: {
            ...(current.comic || {}),
            globalReferenceImage: settingsForStart.aiPromptConfig.comic.globalReferenceImage
          }
        }));
        setPrivateComicPendingReference('');
      }
      if (!ensured.created || hadPendingReference) {
        await savePrivateComicSettingsPatch(snapshot.batchId, settingsForStart);
      }
      if (ensured.created || !privateComicAssets.length) {
        await runBookStage(snapshot.batchId, snapshot.bookId, 'assets', {
          mode: 'missing',
          textModelId: privateComicDraft.textModelId || undefined,
          requestId: privateComicOperationId()
        });
      }
      const pageHasPrompt = page => Boolean(String(page?.comicPrompt || '').trim() || String(page?.imagePrompt || '').trim());
      const promptPageCount = privateComicPages.filter(pageHasPrompt).length;
      const promptMode = privateComicPages.length && promptPageCount >= privateComicPages.length
        ? 'force'
        : promptPageCount > 0 ? 'missing' : 'bootstrap';
      await runBookStage(snapshot.batchId, snapshot.bookId, 'comic', {
        mode: promptMode,
        textModelId: privateComicDraft.textModelId || undefined,
        requestId: privateComicOperationId()
      });
      setPrivateComicTask(snapshot);
      await refreshPrivateComicTask(snapshot.batchId, { ...snapshot, status: 'running' });
    } catch (error) {
      setPrivateComicError(error?.message || '漫画任务启动失败');
      setPrivateComicTask(current => current ? { ...current, status: 'failed', error: error?.message || '漫画任务启动失败' } : current);
    } finally {
      setPrivateComicBusy(false);
    }
  }

  async function retryPrivateComic() {
    await submitPrivateComic();
  }

  async function persistPrivateComicComicSettings() {
    const batchId = privateComicTask?.batchId;
    if (!batchId) return;
    let settingsPatch = buildPrivateComicSettings({
      sourceText: privateComicDraft.sourceText,
      comic: privateComicDraft.comic,
      textModelId: privateComicDraft.textModelId,
      imageModelId: privateComicDraft.imageModelId,
      imageProvider: privateComicDraft.imageProvider || scriptImageProvider
    });
    if (privateComicPendingReference) {
      const uploaded = await uploadComicGlobalReference(batchId, privateComicPendingReference);
      settingsPatch = mergePrivateComicReference(settingsPatch, uploaded?.reference || {});
      setPrivateComicDraft(current => ({
        ...current,
        comic: {
          ...(current.comic || {}),
          globalReferenceImage: settingsPatch.aiPromptConfig.comic.globalReferenceImage
        }
      }));
      setPrivateComicPendingReference('');
    }
    const before = privateComicBatchValue(await getBatch(batchId, { cache: 'no-store' }));
    const expectedRevision = Number(before?.revision || 0);
    const saved = await saveBatchSettings(batchId, { expectedRevision, patch: settingsPatch });
    const savedRevision = Number(saved?.revision || 0);
    const after = privateComicBatchValue(await getBatch(batchId, { cache: 'no-store' }));
    const afterRevision = Number(after?.revision || savedRevision || 0);
    if (!(afterRevision > expectedRevision)) throw new Error('漫画约束保存后服务端版本号未更新，请重试');
    const savedComic = after?.settingsState?.patch?.aiPromptConfig?.comic;
    if (savedComic && String(savedComic.presetId || '').trim() !== String(privateComicDraft.comic?.presetId || '').trim()) {
      throw new Error('漫画约束保存后回读不一致，请重试');
    }
    setPrivateComicTask(current => current ? { ...current, revision: afterRevision } : current);
    message.success('漫画约束已保存到当前任务');
  }

  useEffect(() => {
    if (customizationWorkspace !== 'comic' || !privateComicTask?.batchId) return undefined;
    let active = true;
    const poll = () => { if (active) void refreshPrivateComicTask(); };
    poll();
    const timer = window.setInterval(poll, 5000);
    return () => { active = false; window.clearInterval(timer); };
  }, [customizationWorkspace, privateComicTask?.batchId]);

  return (
    <Form
      className="script-workbench-form"
      form={form}
      initialValues={{ mode: 'continuous', format: 'storyboard', duration: '10s', extractionPreset: 'standard' }}
      onFinish={handleExtract}
      onValuesChange={(changed, allValues) => {
        persistDraft(allValues);
        if (Object.hasOwn(changed, 'novelText')) {
          invalidateRequests();
          replaceSourceAudio('');
          setExtractInfo(normalizeExtractInfo());
          setOutput('');
          setShotVideoTasks({});
          setShotReferenceStates({});
          setSelectedShotIndexes(new Set());
          setCurrentHistoryId('');
          setEditingOutput(false);
          setGenerationStage('idle');
          setExtracting(false);
          setGenerating(false);
          setRegeneratingEntities(false);
          setRegeneratingOutput(false);
          setNarrating(false);
          syncPetContext(changed.novelText);
        }
      }}
    >
      <div
        ref={workbenchRef}
        id="private-customization-panel-script"
        role="tabpanel"
        aria-labelledby="private-customization-tab-script"
        className="script-workbench utility-workbench private-studio-workbench private-studio-resizable"
        style={customizationWorkspace === 'script'
          ? (leftPanelWidth ? { gridTemplateColumns: `${leftPanelWidth}px 8px minmax(460px, 1fr)` } : undefined)
          : { display: 'none' }}
      >
        <div className="script-left private-studio-side">
        <div className="script-left-scroll">
        <div className="private-studio-left-fixed">
          <div className="private-customization-brand-row">
            <div className="private-customization-brand"><span className="private-customization-brand-mark">私</span><strong>私人定制</strong></div>
            <span
              className={`private-customization-save-state is-${draftSaveState.status}`}
              role="status"
              aria-live="polite"
              title={draftSaveState.status === 'failed' ? `草稿保存失败：${draftSaveState.reason}` : undefined}
            ><i />{draftSaveLabel}</span>
          </div>
          <div className="private-customization-workspaces" role="tablist" aria-label="私人定制工作区">
            <button type="button" role="tab" id="private-customization-tab-script" aria-controls="private-customization-panel-script" aria-selected={customizationWorkspace === 'script'} className={customizationWorkspace === 'script' ? 'is-active' : ''} onClick={() => setCustomizationWorkspace('script')}>剧本创作</button>
            <button type="button" role="tab" id="private-customization-tab-comic" aria-controls="private-customization-panel-comic" aria-selected={customizationWorkspace === 'comic'} className={customizationWorkspace === 'comic' ? 'is-active' : ''} onClick={() => setCustomizationWorkspace('comic')}>漫画生成</button>
          </div>
          <div className="private-customization-panel-heading">
            <div>
              <span className="private-customization-eyebrow">01 · 粘贴你的内容</span>
              <h2>输入内容</h2>
            </div>
            <span className="private-customization-save-state">支持 TXT / 长文本</span>
          </div>
          <div className="script-chat-shell">
            <div className="script-source-input">
              <Form.Item name="novelText" rules={[{ required: true, message: '请先粘贴小说原文' }]}>
                <Input.TextArea className="script-chat-textarea" autoSize={{ minRows: 4, maxRows: 10 }} placeholder="粘贴小说原文，开始构思... ✦" />
              </Form.Item>
              {novelText ? (
                <button
                  type="button"
                  className="script-source-clear"
                  aria-label="清空小说原文"
                  title="清空小说原文"
                  onClick={clearNovelText}
                >×</button>
              ) : null}
            </div>
            <div className="script-chat-options">
              <div className="script-chat-tools">
                <label className="script-file-button" title="添加 TXT 原文">
                  <FileText size={17} strokeWidth={1.8} aria-hidden="true" />
                  <input type="file" accept=".txt,text/plain" onChange={addTextFile} />
                </label>
                <button type="button" aria-label="切换人物场景提取指令" title="切换指令" onClick={() => {
                  setPendingExtractionPreset(extractionPreset);
                  setInstructionModalOpen(true);
                }} disabled={extractionUnavailable}><Plus size={17} strokeWidth={1.8} aria-hidden="true" /></button>
                <button type="button" aria-label="配音原文" title="按当前配音预设生成原文配音" onClick={narrateSource} disabled={narrating}><AudioLines size={17} strokeWidth={1.8} aria-hidden="true" /></button>
                <button type="button" aria-label="匹配音频" title="设置生成剧本时是否匹配当前配音时长" onClick={openAudioMatchSettings}><Clapperboard size={17} strokeWidth={1.8} aria-hidden="true" /></button>
                <Button
                  className="private-studio-extract"
                  type="primary"
                  htmlType="submit"
                  loading={extracting}
                  disabled={generating || extractionUnavailable || !hasScriptText}
                  aria-label="提取人物与场景"
                  title={hasScriptText ? '提取人物与场景；生成剧本请用右侧「开始生成」' : '先粘贴或输入内容'}
                  icon={<WandSparkles size={14} strokeWidth={1.8} aria-hidden="true" />}
                >{extracting ? '提取中…' : '提取'}</Button>
              </div>
            </div>
          </div>
          <div className="private-studio-count-row">
            <span>{hasScriptText ? `已读取 ${scriptReadCharCount} 字` : '尚未读取'}</span>
            <span>自动保存</span>
          </div>
          {draftSaveState.status === 'failed' ? (
            <div className="private-studio-draft-error" role="alert">
              草稿保存失败，内容仍在页面上，请勿关闭（{draftSaveState.reason}）
            </div>
          ) : null}
          <PrivateStudioSteps steps={scriptSteps} />
        </div>
        <div className="private-studio-left-scroll-body">
          <PrivateStudioChips
            title="已识别素材"
            actionLabel="可编辑"
            items={scriptAssetChips}
            onAdd={() => addEntity('characters')}
          />

      <div className="script-instruction-status">提取：{extractionPresetName}</div>
          {(sourceAudioUrl || narrating || sourceAudioError) ? (
            <div className="script-source-audio-card" role="status" aria-live="polite">
              <div className="script-source-audio-header">
                <span>当前配音</span>
                {sourceAudioDurationSeconds ? <span className="script-source-audio-duration">时长 {sourceAudioDurationSeconds} 秒</span> : null}
              </div>
              {sourceAudioUrl ? <audio className="script-source-audio" controls src={sourceAudioUrl} /> : null}
              {narrating ? <div className="script-source-audio-status">正在生成原文配音...</div> : null}
              {sourceAudioError ? <div className="script-source-audio-error">{sourceAudioError}</div> : null}
            </div>
          ) : null}
          {extracting ? <div className="script-generation-stage" role="status">正在提取人物与场景...</div> : null}
          {!extracting && generationStage === 'extracted' ? <div className="script-generation-stage">人物与场景已提取，请检查、编辑或重生后，再点击右侧“生成剧本”。</div> : null}

          <EntitySection title="人物" type="characters" count={extractInfo.characters.length} items={extractInfo.characters} protagonistIds={extractInfo.protagonistIds} onAdd={() => addEntity('characters')} onEdit={id => openEntityEditor('characters', id)} onToggleProtagonist={toggleProtagonist} onRegenerate={regenerateEntities} regenerating={regeneratingEntities} />
          <EntitySection title="场景" type="scenes" count={extractInfo.scenes.length} items={extractInfo.scenes} onAdd={() => addEntity('scenes')} onEdit={id => openEntityEditor('scenes', id)} onRegenerate={regenerateEntities} regenerating={regeneratingEntities} />
        </div>
        </div>
      </div>
        <div
          className="script-resize-handle"
          role="separator"
          aria-label="调整左右面板宽度"
          aria-orientation="vertical"
          onPointerDown={handleResizeStart}
        />
        <div className="script-right private-studio-main">
        <div className="private-customization-canvas-heading">
          <div>
            <span className="private-customization-eyebrow">创作画布</span>
            <div className="private-customization-title-line"><h2>{scriptHeaderTitle}</h2><span className="private-customization-ready-badge"><i />{scriptHeaderStatus}</span></div>
            <p className="private-customization-work-meta">{scriptHeaderMeta}</p>
          </div>
          <div className="private-studio-head-right">
            <span className="private-customization-status-dot">剧本创作</span>
            <PrivateStudioHeaderActions
              onHistory={openScriptHistory}
              historyTitle="剧本创作历史版本"
              onSettings={openScriptUnifiedConfig}
            />
          </div>
        </div>
        <div className="script-tabs">
          <Typography.Text strong>剧本创作</Typography.Text>

          <Form.Item name="mode" noStyle>
            <Segmented
              options={[
                { label: '连续开头', value: 'continuous' },
                { label: '爆款开头', value: 'hook' },
                { label: '分段开头', value: 'segmented' }
              ]}
            />
          </Form.Item>
          <div style={{ flex: 1 }} />
          <Form.Item name="duration" noStyle>
            <Segmented options={['10s', '15s']} />
          </Form.Item>
        </div>
        <div className="script-toolbar private-studio-toolbar">
          <div className="private-studio-toolbar-left">
          <div className="private-studio-tool-pill">
            <span className="private-studio-tool-label">输出</span>
          <Form.Item name="format" noStyle>
            <Select
              className="private-studio-tool-select"
              style={{ width: 140 }}
              options={[
                { label: '画布模式', value: 'storyboard' },
                { label: '剧本模式', value: 'shortdrama' },
                { label: '剧情模式', value: 'screenplay' },
                { label: '分镜模式', value: 'shotlist' },
                { label: 'Q版模式', value: 'q版' }
              ]}
            />
          </Form.Item>
          </div>
          </div>
          <div className="private-studio-toolbar-right">
          <Button
            className="script-constraint-button private-studio-tool-pill-button"
            icon={<Settings2 size={16} strokeWidth={1.8} aria-hidden="true" />}
            onClick={openConstraints}
            title="约束设置"
          >约束条件</Button>
          <Space>
            <Button className="private-studio-tool-primary" type="primary" icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={generateOutput} loading={generating} disabled={extracting || !canGenerateScript || (!extractInfo.characters.length && !extractInfo.scenes.length)}>开始生成</Button>
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  { key: 'copy', label: '复制', icon: <Copy size={14} />, disabled: !output },
                  { key: 'replace', label: '查找替换', disabled: !isShotCardView || !selectedShotIndexes.size },
                  { key: 'edit', label: editingOutput ? '完成编辑' : '编辑', icon: <Pencil size={14} />, disabled: !output },
                  { key: 'undo', label: '撤销', disabled: !previousOutput },
                  { key: 'regenerate', label: '重新生成', icon: <RefreshCw size={14} />, disabled: !output || !canGenerateScript },
                  { key: 'export', label: '导出', icon: <Download size={14} />, disabled: !output }
                ],
                onClick: ({ key }) => {
                  if (key === 'copy') copyText(isShotCardView ? shotCards.join('\n\n') : output);
                  if (key === 'replace') setShotReplaceOpen(true);
                  if (key === 'edit') toggleOutputEditor();
                  if (key === 'undo') undoLastRevision();
                  if (key === 'regenerate') regenerateOutput();
                  if (key === 'export') exportOutput();
                }
              }}
            >
              <Button aria-label="更多操作" title="更多操作" icon={<MoreHorizontal size={17} strokeWidth={1.8} />} />
            </Dropdown>
          </Space>
        </div>
          </div>
        <div className="private-studio-content">
        <div className="private-studio-center">
        <div className="script-output">
          {output ? (
            isShotCardView ? <><div className="private-studio-canvas-head"><b>分镜草稿</b><span>{shotCards.length} 个镜头</span></div><ShotOutputCards
              cards={shotCards}
              duration={form.getFieldValue('duration')}
              selectedIndexes={selectedShotIndexes}
              onToggle={index => setSelectedShotIndexes(current => {
                const next = new Set(current);
                if (next.has(index)) next.delete(index);
                else next.add(index);
                return next;
              })}
              onToggleAll={() => setSelectedShotIndexes(current => current.size === shotCards.length ? new Set() : new Set(shotCards.map((_, index) => index)))}
              onCopy={copyText}
              onCopySelected={() => copyText(joinShotCards(shotCards, selectedShotIndexes))}
              onGenerateVideo={canGenerateVideo ? generateVideoForShot : null}
              generatingIndexes={generatingShotIndexes}
              videoTasks={shotVideoTasks}
              videoTaskHistory={shotVideoTaskHistory}
              extractInfo={extractInfo}
              mentionCandidates={mentionCandidates}
              shotReferenceStates={shotReferenceStates}
              onToggleReference={(index, imageUrl) => {
                const current = shotReferenceStates?.[index] || {};
                const disabledImageUrls = new Set(current.disabledImageUrls || []);
                if (disabledImageUrls.has(imageUrl)) disabledImageUrls.delete(imageUrl);
                else disabledImageUrls.add(imageUrl);
                const next = toggleShotReferenceState(shotReferenceStates, index, { disabledImageUrls: [...disabledImageUrls] });
                setShotReferenceStates(next);
                persistDraft(undefined, { shotReferenceStates: next });
              }}
              onOpenVideo={setPreviewVideoTask}
              onOpenVideoHistory={openVideoHistory}
              onEditPrompt={openShotEditor}
              output={output}
              activeMatch={shotReplaceOpen ? activeShotMatch : null}
              cardStarts={shotCardStarts}
            /></> : <div style={{ position: 'relative' }}>
              <ScriptMentionEditor
                value={output}
                candidates={mentionCandidates}
                editable={editingOutput}
                placeholder="任意位置输入 @ 选择人物或场景"
                selectionOffset={editingOutputSelection.start}
                onChange={next => updateOutputDraft(next)}
                onQueryChange={({ text, cursor, rect }) => editingOutput && syncOutputMentionSelection(text, cursor, cursor, rect)}
                onEditorKeyDown={event => {
                  if (!editingOutput || !activeOutputMention) return;
                  if (event.key === 'Escape') { event.preventDefault(); setOutputMentionMenuOpen(false); return; }
                  if (!activeOutputMentionCandidates.length) return;
                  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); const delta = event.key === 'ArrowDown' ? 1 : -1; setOutputMentionActiveIndex(current => (current + delta + activeOutputMentionCandidates.length) % activeOutputMentionCandidates.length); }
                  if (event.key === 'Enter') { event.preventDefault(); insertOutputMention(activeOutputMentionCandidates[outputMentionActiveIndex] || activeOutputMentionCandidates[0]); }
                }}
              />
              {editingOutput && activeOutputMention ? <ScriptMentionMenu anchorRect={outputMentionAnchorRect} candidates={activeOutputMentionCandidates} activeIndex={outputMentionActiveIndex} onChoose={insertOutputMention} onCreateCharacter={() => addEntity('characters')} onCreateScene={() => addEntity('scenes')} /> : null}
            </div>
          ) : generating ? (
            <CmLoader />
          ) : (
            <div className="script-empty legacy-panel-card">
              <div className="script-empty-card">
                <span className="script-empty-dot" aria-hidden="true" />
                <span className="script-empty-ray" aria-hidden="true" />
                <span className="script-empty-line script-empty-line--top" aria-hidden="true" />
                <span className="script-empty-line script-empty-line--bottom" aria-hidden="true" />
              </div>
              <div className="script-empty-title">准备创作</div>
              <div className="script-empty-copy">{hasScriptText ? (hasScriptAssets ? '素材已就绪，点击右侧「开始生成」' : '提取素材后开始生成') : '先粘贴正文，再提取素材'}</div>
            </div>
          )}
          <Modal title="编辑分镜提示词" open={editingShot.index >= 0} onCancel={closeShotEditor} onOk={() => {
            const nextOutput = replaceRawShotCard(output, rawShotCards, editingShot.index, editingShot.text);
            updateOutputDraft(nextOutput, true);
            closeShotEditor();
          }}>
            <div style={{ position: 'relative' }}>
              <ScriptMentionEditor
                value={editingShot.text}
                candidates={mentionCandidates}
                editable
                placeholder="任意位置输入 @ 选择人物或场景"
                selectionOffset={editingShotSelection.start}
                onChange={next => setEditingShot(current => ({ ...current, text: next }))}
                onQueryChange={({ text, cursor, rect }) => syncShotMentionSelection(text, cursor, cursor, rect)}
                onEditorKeyDown={event => {
                  if (!activeShotMention) return;
                  if (event.key === 'Escape') { event.preventDefault(); setMentionMenuOpen(false); return; }
                  if (!activeShotMentionCandidates.length) return;
                  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); const delta = event.key === 'ArrowDown' ? 1 : -1; setMentionActiveIndex(current => (current + delta + activeShotMentionCandidates.length) % activeShotMentionCandidates.length); }
                  if (event.key === 'Enter') { event.preventDefault(); insertShotMention(activeShotMentionCandidates[mentionActiveIndex] || activeShotMentionCandidates[0]); }
                }}
              />
              {activeShotMention ? <ScriptMentionMenu anchorRect={shotMentionAnchorRect} candidates={activeShotMentionCandidates} activeIndex={mentionActiveIndex} onChoose={insertShotMention} onCreateCharacter={() => addEntity('characters')} onCreateScene={() => addEntity('scenes')} /> : null}
            </div>
          </Modal>
        </div>
        </div>
        <aside className="private-studio-side-preview">
          <PrivateStudioNote mode="script" />
        </aside>
        </div>
        </div>
      </div>
      {customizationWorkspace === 'comic' ? (
        <PrivateComicWorkspace
          catalog={privateComicCatalog}
          modelOptions={privateComicModelOptions}
          draft={privateComicDraft}
          onDraftChange={setPrivateComicDraft}
          draftSaveState={privateComicSaveState}
          onSubmit={submitPrivateComic}
          task={privateComicTask}
          comicAssets={privateComicAssets}
          comicAssetImages={privateComicAssetImages}
          comicAssetImageFailures={privateComicAssetImageFailures}
          comicPages={privateComicPages}
          readiness={privateComicReadiness}
          action={privateComicAction}
          actionKey={privateComicActionKey}
          onExtractAssets={() => void runPrivateComicAssets()}
          onGenerateAssetImage={(asset, useThreeView) => void runPrivateComicAssetImages(asset?.id || '', useThreeView)}
          onSaveAsset={savePrivateComicAsset}
          onUploadAssetImage={uploadPrivateComicAssetImage}
          onSetPrimaryAssetImage={setPrivateComicPrimaryAssetImage}
          onGenerateMissingAssetImages={() => void runPrivateComicAssetImages()}
          onGeneratePrompts={() => void runPrivateComicPrompts()}
          onGeneratePageImage={page => void runPrivateComicPageImages(page?.videoId || '')}
          onGenerateAllPageImages={() => void runPrivateComicPageImages()}
          onSavePagePrompt={(page, prompts) => savePrivateComicPagePrompts(page, prompts)}
          onDownloadAll={() => void downloadPrivateComicPackage()}
          busy={privateComicBusy}
          error={privateComicError}
          onRefresh={() => void refreshPrivateComicTask()}
          onRetry={() => void retryPrivateComic()}
          onPendingReferenceChange={setPrivateComicPendingReference}
          onPersistSettings={persistPrivateComicComicSettings}
          onOpenUnifiedConfig={openScriptUnifiedConfig}
          onReloadCatalog={reloadPrivateComicCatalog}
          onOpenBatchFactory={() => window.location.assign(`/batch-factory?batchId=${encodeURIComponent(privateComicTask?.batchId || '')}`)}
          customizationWorkspace={customizationWorkspace}
          onWorkspaceChange={setCustomizationWorkspace}
        />
      ) : null}
      <Modal title="生成的视频" open={Boolean(previewVideoTask?.videoUrl)} footer={null} onCancel={() => setPreviewVideoTask(null)} width={520}>
        {previewVideoTask?.videoUrl ? <video controls autoPlay style={{ width: '100%', maxHeight: '70vh' }} src={previewVideoTask.videoUrl} /> : null}
      </Modal>
      <Modal
        title={`第 ${previewVideoHistory.shotIndex + 1} 条分镜历史视频`}
        open={previewVideoHistory.open}
        footer={null}
        onCancel={() => setPreviewVideoHistory(current => ({ ...current, open: false }))}
        width={560}
      >
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          {previewVideoHistory.tasks.map((task, index) => (
            <div key={task.taskId || index}>
              <Typography.Text type="secondary">历史版本 {index + 1}{task.prompt ? ` · ${task.prompt.slice(0, 80)}` : ''}</Typography.Text>
              {task.status === 'succeeded' && task.videoUrl ? <video controls preload="metadata" style={{ display: 'block', width: '100%', marginTop: 8 }} src={task.videoUrl} /> : null}
              {task.historyStatusReadError ? <Typography.Paragraph type="secondary">暂时无法读取任务状态，请稍后再试</Typography.Paragraph> : null}
              {!task.historyStatusReadError && task.status === 'processing' ? <Typography.Paragraph type="secondary">视频生成中，暂不可播放</Typography.Paragraph> : null}
              {!task.historyStatusReadError && task.status === 'failed' ? <Typography.Paragraph type="danger">生成失败{task.error ? `：${task.error}` : ''}</Typography.Paragraph> : null}
              {!task.historyStatusReadError && task.status === 'succeeded' && !task.videoUrl ? <Typography.Paragraph type="danger">任务已完成，但未返回可播放地址</Typography.Paragraph> : null}
              {!task.historyStatusReadError && !['succeeded', 'processing', 'failed'].includes(task.status) ? <Typography.Paragraph type="secondary">暂时无法读取任务状态，请稍后再试</Typography.Paragraph> : null}
            </div>
          ))}
        </Space>
      </Modal>
      <Modal
        title="匹配音频设置"
        open={quickDirectorOpen}
        okText="保存设置"
        cancelText="取消"
        onCancel={() => setQuickDirectorOpen(false)}
        onOk={() => {
          if (quickDirectorOptions.matchAudio && !sourceAudioDurationSeconds) {
            message.warning('请先生成当前原文配音；匹配音频需要读取实际音频时长');
            return;
          }
          setQuickDirectorOpen(false);
        }}
      >
        <Typography.Paragraph type="secondary">
          这里只控制下一次“生成剧本”是否加入匹配音频规则，不会单独调用 AI。开启后，后台会把实际配音秒数与当前 10s / 15s 单分镜上限一起加入本次生成提示词。
        </Typography.Paragraph>
        <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
          <div>
            <Typography.Text strong>匹配当前配音时长</Typography.Text>
            <Typography.Paragraph type="secondary" style={{ margin: '2px 0 0' }}>
              {sourceAudioDurationSeconds ? '当前配音：' + sourceAudioDurationSeconds + ' 秒；生成剧本时总分镜时长将严格匹配。' : '请先点击配音按钮生成当前原文配音。'}
            </Typography.Paragraph>
          </div>
          <Switch
            checked={quickDirectorOptions.matchAudio}
            onChange={matchAudio => setQuickDirectorOptions(current => ({ ...current, matchAudio }))}
          />
        </Space>
      </Modal>
      <Modal title="剧本创作历史" open={historyOpen} footer={null} onCancel={() => setHistoryOpen(false)}>
        {historyLoading ? <Typography.Text type="secondary">正在加载历史记录…</Typography.Text> : historyEntries.length ? historyEntries.map(entry => {
          const archivedVideoCount = Object.values(entry.videoTaskHistory || {}).reduce((total, tasks) => total + (Array.isArray(tasks) ? tasks.length : 0), 0);
          return <Button key={entry.id} block style={{ height: 'auto', marginBottom: 8, textAlign: 'left', whiteSpace: 'normal' }} onClick={() => restoreHistory(entry)}><div>{entry.preview || '未命名剧本'}</div><Typography.Text type="secondary">{entry.duration || '-'} · 视频 {Object.keys(entry.videoTasks || {}).length + archivedVideoCount} 个</Typography.Text></Button>;
        }) : <Typography.Text type="secondary">暂无生成历史</Typography.Text>}
      </Modal>
      <Modal
        title="替换已选分镜文字"
        open={shotReplaceOpen}
        onCancel={() => setShotReplaceOpen(false)}
        footer={null}
      >
        <Typography.Paragraph>已选 {selectedShotIndexes.size} 条分镜</Typography.Paragraph>
        <Form.Item label="查找内容">
          <Input value={shotFindText} onChange={event => { setShotFindText(event.target.value); setShotMatchIndex(0); }} />
        </Form.Item>
        <Form.Item label="替换为">
          <Input value={shotReplaceText} onChange={event => setShotReplaceText(event.target.value)} />
        </Form.Item>
        {!shotFindText ? <Typography.Text type="secondary">请输入查找内容</Typography.Text> : !selectedShotMatches.length ? <Typography.Text type="secondary">未找到匹配内容</Typography.Text> : (
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
            找到内容：“{shotFindText}” · 分镜 {activeShotMatch.cardIndex + 1} · {shotMatchIndex + 1} / {selectedShotMatches.length}
          </Typography.Paragraph>
        )}
        <Space wrap style={{ marginTop: 16 }}>
          <Button disabled={!selectedShotMatches.length} onClick={() => setShotMatchIndex(index => (index - 1 + selectedShotMatches.length) % selectedShotMatches.length)}>上一个</Button>
          <Button disabled={!selectedShotMatches.length} onClick={() => setShotMatchIndex(index => (index + 1) % selectedShotMatches.length)}>下一个</Button>
          <Button disabled={!activeShotMatch} onClick={replaceCurrentShotMatch}>替换当前</Button>
          <Button type="primary" disabled={!selectedShotMatches.length} onClick={replaceAllShotMatches}>全部替换</Button>
        </Space>
      </Modal>
      <Modal
        title="CM 修改预览"
        open={revisionPreview.open}
        onCancel={closeRevisionPreview}
        onOk={applyRevisionPreview}
        okText="应用修改"
        cancelText="放弃修改"
        width={1100}
      >
        {revisionPreview.summary ? <Typography.Paragraph>{revisionPreview.summary}</Typography.Paragraph> : null}
        <div className="cm-revision-preview">
          <section>
            <Typography.Title level={5}>当前剧本</Typography.Title>
            <Input.TextArea value={revisionPreview.currentOutput} rows={24} readOnly />
          </section>
          <section>
            <Typography.Title level={5}>CM 修改稿</Typography.Title>
            <Input.TextArea value={revisionPreview.candidateOutput} rows={24} readOnly />
          </section>
        </div>
      </Modal>
      <Modal
        title="选择基础设定（人物 / 场景）提示词"
        open={instructionModalOpen}
        onCancel={() => setInstructionModalOpen(false)}
        onOk={() => {
          form.setFieldValue('extractionPreset', pendingExtractionPreset);
          setConstraints(current => ({
            ...current,
            enabled: true,
            baseSetup: {
              ...current.baseSetup,
              enabled: true,
              source: 'system',
              presetId: pendingExtractionPreset,
              personalPromptId: ''
            }
          }));
          setDraftConstraints(current => ({
            ...current,
            enabled: true,
            baseSetup: {
              ...current.baseSetup,
              enabled: true,
              source: 'system',
              presetId: pendingExtractionPreset,
              personalPromptId: ''
            }
          }));
          setInstructionModalOpen(false);
          persistDraft({
            ...form.getFieldsValue(),
            extractionPreset: pendingExtractionPreset,
            constraints: {
              ...constraintsForNextGeneration(constraints),
              enabled: true,
              baseSetup: {
                ...constraints.baseSetup,
                enabled: true,
                source: 'system',
                presetId: pendingExtractionPreset,
                personalPromptId: ''
              }
            }
          });
        }}
        okText="保存"
      >
        <Form.Item label="基础设定提示词">
          <Select
            value={pendingExtractionPreset}
            onChange={setPendingExtractionPreset}
            options={extractionPresets.map(item => ({ label: `${item.name} · v${item.version}`, value: item.id }))}
          />
        </Form.Item>
        <Typography.Paragraph type="secondary">
          {selectedExtractionPreset?.description || extractionPresetError || '管理员发布后可在此选择提取指令。'}
        </Typography.Paragraph>
      </Modal>
      <PrivateStudioConstraintsModal
        open={constraintModalOpen}
        onCancel={() => setConstraintModalOpen(false)}
        onOk={saveConstraints}
        footerNote="子开关始终作用于本次生成。打开“后续剧本输出约束”才会把当前选择保存并自动用于下次生成。"
      >
        <div className="script-constraint-section">
          <Space align="center">
            <Switch
              checked={draftConstraints.enabled}
              onChange={enabled => setDraftConstraints(current => ({ ...current, enabled }))}
            />
            <Typography.Text strong>后续剧本输出约束</Typography.Text>
          </Space>
        </div>
        <div className="script-constraint-section">
          <Space align="center">
            <Switch
              checked={draftConstraints.baseSetup?.enabled !== false}
              onChange={enabled => updateDraftConstraint('baseSetup', { enabled })}
            />
            <div onClick={() => setActiveConstraintCategory('baseSetup')}>
              <Typography.Text strong>基础设定（人物 / 场景）</Typography.Text>
              <Typography.Paragraph type="secondary" style={{ margin: '2px 0 0' }}>开启后，自动将已提取的人物设定和第一个场景加入每张分镜。</Typography.Paragraph>
            </div>
          </Space>
        </div>
        {[
          ['prefix', '画面前缀词'],
          ['quality', '画质约束'],
          ['restriction', '画面限制'],
          ['negative', '负面提示词']
        ].map(([category, label]) => (
          <ConstraintCategoryEditor
            key={category}
            category={category}
            label={label}
            value={draftConstraints[category]}
            systemOptions={constraintOptions(category)}
            personalPrompts={personalConstraintPrompts[category] || []}
            loadingPersonalPrompts={loadingPersonalConstraintPrompts[category]}
            saving={savingConstraintCategory === category}
            editingPersonalPromptId={editingPersonalPromptId}
            onFocus={() => setActiveConstraintCategory(category)}
            onChange={patch => updateDraftConstraint(category, patch)}
            onSelectSystem={presetId => selectSystemConstraint(category, presetId)}
            onSelectPersonal={promptId => selectPersonalConstraint(category, promptId)}
            onSaveDraft={() => { setEditingPersonalPromptId(''); setEditingPersonalPromptName(''); savePersonalConstraint(category, null); }}
             onSaveNamed={() => setPersonalPromptNameModal({ open: true, category, name: editingPersonalPromptName })}
             onEditPersonal={prompt => {
               setEditingPersonalPromptId(prompt.id);
               setEditingPersonalPromptName(prompt.name || '');
               updateDraftConstraint(category, { source: 'draft', presetId: '', personalPromptId: '', body: prompt.body });
             }}
            onDeletePersonal={id => deletePersonalConstraint(category, id)}
          />
        ))}
      </PrivateStudioConstraintsModal>
      <Modal
        title={editingPersonalPromptId ? '编辑我的提示词' : '保存为我的提示词'}
        open={personalPromptNameModal.open}
        onCancel={() => {
          setPersonalPromptNameModal({ open: false, category: '', name: '' });
          setEditingPersonalPromptId('');
          setEditingPersonalPromptName('');
        }}
        onOk={() => savePersonalConstraint(personalPromptNameModal.category, personalPromptNameModal.name.trim())}
        okText="保存"
        confirmLoading={savingConstraintCategory === personalPromptNameModal.category}
      >
        <Input
          autoFocus
          maxLength={80}
          placeholder="请输入提示词名称"
          value={personalPromptNameModal.name}
          onChange={event => setPersonalPromptNameModal(current => ({ ...current, name: event.target.value }))}
        />
        <Typography.Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
          将当前提示词内容保存到您的账号。保存后可在任何设备登录继续使用，最近一次用于生成会排在最前。
        </Typography.Paragraph>
      </Modal>
      <Modal
        title={customizationWorkspace === 'comic' ? '漫画生成 · 统一配置' : '剧本创作 · 统一配置'}
        open={scriptConfigOpen}
        onCancel={() => setScriptConfigOpen(false)}
        onOk={saveScriptUnifiedConfig}
        okText="保存为默认配置"
        confirmLoading={savingScriptConfig}
        destroyOnClose
      >
        <Space direction="vertical" size={14} style={{ width: '100%' }}>
          <Form.Item label="文本模型" style={{ marginBottom: 0 }}><Select value={scriptConfigDraft.textModelId || undefined} placeholder="选择文本模型" options={scriptModels.text.map(model => ({ value: model.id, label: model.displayName || model.id }))} onChange={textModelId => setScriptConfigDraft(current => ({ ...current, textModelId }))} /></Form.Item>
          <Form.Item label="图片生成通道" style={{ marginBottom: 0 }}><Select value={scriptConfigDraft.imageProvider || 'personal_api'} options={scriptImageProviderOptions} onChange={imageProvider => setScriptConfigDraft(current => ({ ...current, imageProvider, imageModelId: reconcileProviderModel({ models: scriptModels.image, kind: 'image', provider: imageProvider, modelId: current.imageModelId }).modelId }))} /></Form.Item>
          <Form.Item label="图片模型" style={{ marginBottom: 0 }}><Select value={scriptConfigDraft.imageModelId || undefined} placeholder="选择当前通道的图片模型" options={modelOptionsForProvider(scriptModels.image, 'image', scriptConfigDraft.imageProvider)} onChange={imageModelId => setScriptConfigDraft(current => ({ ...current, imageModelId }))} /></Form.Item>
          <Form.Item label="视频生成通道" style={{ marginBottom: 0 }}><Select value={scriptConfigDraft.videoProvider || 'personal_api'} options={scriptVideoProviderOptions} onChange={videoProvider => setScriptConfigDraft(current => ({ ...current, videoProvider, videoModelKey: reconcileProviderModel({ models: scriptVideoModels, kind: 'video', provider: videoProvider, modelId: current.videoModelKey }).modelId }))} /></Form.Item>
          <Form.Item label="视频模型" style={{ marginBottom: 0 }}><Select loading={loadingScriptVideoModels} value={scriptConfigDraft.videoModelKey} options={modelOptionsForProvider(scriptVideoModels, 'video', scriptConfigDraft.videoProvider)} onChange={videoModelKey => setScriptConfigDraft(current => ({ ...current, videoModelKey }))} /></Form.Item>
          <Form.Item label="图片画幅" style={{ marginBottom: 0 }}><Segmented block options={IMAGE_ASPECT_RATIOS} value={scriptConfigDraft.imageAspectRatio} onChange={imageAspectRatio => setScriptConfigDraft(current => ({ ...current, imageAspectRatio }))} /></Form.Item>
          <Form.Item label="视频画幅" style={{ marginBottom: 0 }}><Segmented block options={['16:9', '9:16']} value={scriptConfigDraft.videoAspectRatio} onChange={videoAspectRatio => setScriptConfigDraft(current => ({ ...current, videoAspectRatio }))} /></Form.Item>
          <Form.Item label="视频分辨率" style={{ marginBottom: 0 }}><Segmented block options={['480p', '720p', '1080p']} value={scriptConfigDraft.videoResolution} onChange={videoResolution => setScriptConfigDraft(current => ({ ...current, videoResolution }))} /></Form.Item>
          {scriptConfigDraft.videoModelKey === 'minimax-h3-video' ? <Typography.Text type="secondary">H3 会保持所选横竖画幅；720p 与 1080p 会按 H3 的 768p 档提交。</Typography.Text> : null}
          <Typography.Text type="secondary">保存后会成为当前账号后续剧本创作的默认值；仍可随时在这里更换。</Typography.Text>
        </Space>
      </Modal>
      <EntityEditor
        key={activeEntity?.editorSessionId || 'closed'}
        entity={activeItem}
        type={activeEntity?.type}
        assetId={activeEntity?.id || ''}
        editorSessionId={activeEntity?.editorSessionId || 0}
        isNew={Boolean(activeEntity?.isNew)}
        open={Boolean(activeEntity)}
        fullscreen={fullscreenEditor}
        onClose={() => setActiveEntity(null)}
        onToggleFullscreen={() => setFullscreenEditor(value => !value)}
        novelText={form.getFieldValue('novelText')}
        extractionPreset={extractionPreset}
        existingEntitySummary={compactEntitySummary(extractInfo, activeEntity?.isNew ? '' : activeEntity?.id)}
        textModelId={scriptModelSelection.textModelId}
        imageModelId={scriptModelSelection.imageModelId}
        imageProvider={scriptImageProvider}
        imageAspectRatio={scriptImageAspectRatio}
        onChange={updateActiveEntity}
        onDelete={deleteActiveEntity}
      />
    </Form>
  );
}

export default ScriptPage;

function ConstraintCategoryEditor({ category, label, value, systemOptions, personalPrompts, loadingPersonalPrompts, saving, editingPersonalPromptId, onFocus, onChange, onSelectSystem, onSelectPersonal, onSaveDraft, onSaveNamed, onEditPersonal, onDeletePersonal }) {
  const isSystem = value.source === 'system';
  const selectedPersonalPrompt = personalPrompts.find(item => item.id === value.personalPromptId);
  return (
    <div className="script-constraint-section">
      <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
        <Typography.Text strong>{label}</Typography.Text>
        <Switch checked={value.enabled} onChange={enabled => onChange({ enabled })} />
      </Space>
      {value.enabled ? <>
        <Segmented
          block
          options={[{ label: '系统预设', value: 'system' }, { label: '我的提示词', value: 'personal' }]}
          value={isSystem ? 'system' : 'personal'}
          onChange={source => onChange(source === 'system'
            ? { source: 'system', personalPromptId: '' }
            : { source: value.personalPromptId ? 'personal' : 'draft', presetId: '' })}
          style={{ marginTop: 10 }}
        />
        {isSystem ? <Select
          allowClear
          placeholder="选择系统预设"
          options={systemOptions}
          value={value.presetId || undefined}
          onChange={onSelectSystem}
          style={{ width: '100%', marginTop: 8 }}
        /> : <>
          <Select
            allowClear
            loading={loadingPersonalPrompts}
            placeholder="选择我的提示词，或直接编辑当前草稿"
            options={personalPrompts.map(item => ({ label: item.name || '未命名个人副本', value: item.id }))}
            value={value.personalPromptId || undefined}
            onChange={onSelectPersonal}
            style={{ width: '100%', marginTop: 8 }}
          />
          {selectedPersonalPrompt ? <Space size={8} wrap style={{ marginTop: 8 }}>
            <Button size="small" onClick={() => onEditPersonal(selectedPersonalPrompt)}>编辑所选提示词</Button>
            <Popconfirm title="确认删除该个人提示词？" onConfirm={() => onDeletePersonal(selectedPersonalPrompt.id)}><Button size="small" danger>删除</Button></Popconfirm>
          </Space> : null}
        </>}
        <Typography.Text type="secondary" style={{ display: 'block', marginTop: 12 }}>提示词内容</Typography.Text>
        <Input.TextArea
          rows={6}
          placeholder="选择系统预设后可在此编辑完整提示词内容，保存不会影响系统预设。"
          value={value.body}
          onFocus={onFocus}
          onChange={event => onChange({ source: 'draft', personalPromptId: '', body: event.target.value })}
          style={{ marginTop: 6 }}
        />
        <Space wrap style={{ marginTop: 10 }}>
          <Button loading={saving} onClick={onSaveDraft}>保存当前草稿</Button>
          <Button type="primary" loading={saving} onClick={onSaveNamed}>{editingPersonalPromptId ? '保存编辑' : '保存为我的提示词'}</Button>
        </Space>
        {editingPersonalPromptId ? <Typography.Text type="secondary" style={{ display: 'block', marginTop: 8 }}>正在编辑个人提示词，保存后会覆盖该条记录。</Typography.Text> : null}
      </> : null}
    </div>
  );
}

function CmLoader() {
  return (
    <div className="cm-loader" role="status" aria-live="polite">
      <svg className="cm-loader-defs" aria-hidden="true">
        <defs>
          <linearGradient id="cm-blue">
            <stop stopColor="#973bed" />
            <stop offset="1" stopColor="#007cff" />
          </linearGradient>
          <linearGradient id="cm-green">
            <stop stopColor="#00e0ed" />
            <stop offset="1" stopColor="#00da72" />
          </linearGradient>
        </defs>
      </svg>
      <div className="cm-loader-mark" aria-hidden="true">
        <svg viewBox="0 0 64 64" data-letter="C" aria-hidden="true">
          <path
            className="cm-loader-dash"
            pathLength="360"
            stroke="url(#cm-blue)"
            strokeWidth="8"
            strokeLinecap="round"
            d="M52 14C47.2 8.2 39.9 5 32 5C17.1 5 5 17.1 5 32s12.1 27 27 27c7.9 0 15.2-3.2 20-9"
          />
        </svg>
        <svg viewBox="0 0 64 64" data-letter="M" aria-hidden="true">
          <path
            className="cm-loader-dash"
            pathLength="360"
            stroke="url(#cm-green)"
            strokeWidth="8"
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M7 57V7l25 31L57 7v50"
          />
        </svg>
      </div>
      <strong>正在生成剧本</strong>
      <span>正在提取人物、场景并组织剧情</span>
    </div>
  );
}

function EntitySection({ title, type, count, items, protagonistIds = [], onAdd, onEdit, onToggleProtagonist, onRegenerate, regenerating }) {
  const isCharacter = type === 'characters';
  const protagonistIdSet = new Set(protagonistIds);
  const [allOpen, setAllOpen] = useState(false);
  const hint = `可手动添加${title}，或提取后显示${title}信息`;
  const recentItems = items.slice(0, 3);
  return (
    <div className={`entity-section legacy-panel-card${items.length ? '' : ' is-empty'}`}>
      <div className="entity-section-header">
        <Typography.Text strong>{title}</Typography.Text>
        <Space size={8}>
          <span className="legacy-muted">{count}</span>
          <Button size="small" type="text" className="entity-add-button" onClick={onAdd}>添加{title}</Button>
          <Button
            size="small"
            type="text"
            aria-label={`重新提取${title}`}
            title={`重新提取${title}`}
            icon={<RefreshCw size={15} strokeWidth={1.8} aria-hidden="true" />}
            loading={regenerating}
            onClick={onRegenerate}
          />
          <Tooltip title={hint} trigger={['click', 'focus']}>
            <button type="button" className="entity-section-info" aria-label={hint}>
              <Info size={14} strokeWidth={1.8} aria-hidden="true" />
            </button>
          </Tooltip>
        </Space>
      </div>
      {recentItems.length ? (
        <div className="entity-section-summary">
          {recentItems.map(item => <button type="button" key={item.id} className="entity-section-name" onClick={() => onEdit(item.id)} title={`编辑${formatEntity(item)}`}>{formatEntity(item)}</button>)}
          <Button size="small" type="text" className="entity-section-more" onClick={() => setAllOpen(true)}>查看全部</Button>
        </div>
      ) : null}
      <Modal title={`${title} · ${count}`} open={allOpen} onCancel={() => setAllOpen(false)} footer={null} width={520}>
        <div className="entity-list">
          {items.map(item => {
            const isProtagonist = protagonistIdSet.has(item.id);
            const mainImageUrl = getEntityMedia(item).mainImageUrl;
            const imageCount = getEntityMedia(item).imageUrls.length;
            return <div className="entity-card-row" key={item.id}>
              <button className="entity-card entity-card-button entity-card-main" type="button" onClick={() => { setAllOpen(false); onEdit(item.id); }}>
                <span>{formatEntity(item)}</span>
                <small className="entity-media-status">{mainImageUrl ? '主图已选择' : '未选择主图'}{imageCount ? ` · 图片 ${imageCount} 张` : ''}</small>
              </button>
              {isCharacter && <Button
                className={`entity-protagonist-toggle${isProtagonist ? ' is-protagonist' : ''}`}
                type="text"
                aria-label={isProtagonist ? '取消主角标记' : '设为主角'}
                title={isProtagonist ? '取消主角标记' : '设为主角'}
                icon={<Star size={16} fill={isProtagonist ? 'currentColor' : 'none'} />}
                onClick={() => onToggleProtagonist(item.id)}
              />}
            </div>;
          })}
        </div>
      </Modal>
    </div>
  );
}

function EntityEditor({ entity, type, assetId, editorSessionId, isNew, open, fullscreen, novelText, extractionPreset, existingEntitySummary, textModelId, imageModelId, imageProvider, imageAspectRatio, onClose, onToggleFullscreen, onChange, onDelete }) {
  const [fields, setFields] = useState({});
  const [imageUrls, setImageUrls] = useState([]);
  const [mainImageUrl, setMainImageUrl] = useState('');
  const [enriching, setEnriching] = useState(false);
  const [enrichment, setEnrichment] = useState(null);
  const [enrichmentError, setEnrichmentError] = useState('');
  const activeImageRequestKey = useRef('');

  useEffect(() => {
    if (!open) return;
    setFields(Object.fromEntries(visualFields(entity).map(field => [field.key, field.value])));
    const media = getEntityMedia(entity);
    setImageUrls(media.imageUrls);
    setMainImageUrl(media.mainImageUrl || (media.imageUrls.length === 1 ? media.imageUrls[0] : ''));
    setEnrichment(null);
    setEnrichmentError('');
  }, [entity, open]);

  const fieldsToRender = visualFields(entity);
  const title = `${isNew ? '添加' : '编辑'}${type === 'characters' ? '人物' : '场景'}`;
  const deleteLabel = type === 'characters' ? '删除人物' : '删除场景';
  const canEnrich = Boolean(String(novelText || '').trim() && entityName(fields));
  const imageRequestKey = open && assetId ? `${editorSessionId}:${assetId}` : '';
  activeImageRequestKey.current = imageRequestKey;

  function updateImages(media, sourceKey) {
    if (!imageRequestKey || sourceKey !== activeImageRequestKey.current) return;
    setImageUrls(Array.isArray(media?.imageUrls) ? media.imageUrls : []);
    setMainImageUrl(typeof media?.mainImageUrl === 'string' ? media.mainImageUrl : '');
  }

  async function enrich() {
    if (!canEnrich) return;
    setEnriching(true);
    setEnrichmentError('');
    try {
      const response = await enrichScriptEntity({
        entityType: type === 'characters' ? 'character' : 'scene',
        novelText,
        entity: fields,
        existingEntitySummary,
        extractionPreset,
        textModelId
      });
      const result = normalizeEntityEnrichment(response?.enrichment);
      setFields(current => applyEntityEnrichment(current, result));
      setEnrichment(result);
    } catch (error) {
      setEnrichmentError(error.message || '智能补全失败，请稍后重试');
    } finally {
      setEnriching(false);
    }
  }

  const basicFields = fieldsToRender.filter(field => ['名称', 'name', '角色名称', '场景名称', 'scene'].includes(field.key));
  const settingFields = fieldsToRender.filter(field => !basicFields.includes(field));
  // 剧本侧的「外形」与漫画侧的「人物外形」是两个独立字段，这里显式改名，避免被误认为同一个输入框。
  const fieldLabel = field => type === 'characters' && ['外形', 'appearance', '外观描述'].includes(field.key) ? '外形（剧本）' : field.label;
  const renderField = field => <Form.Item key={field.key} label={fieldLabel(field)}>
    <Input.TextArea value={fields[field.key] || ''} autoSize={{ minRows: 1, maxRows: 6 }} onChange={event => setFields(current => ({ ...current, [field.key]: event.target.value }))} />
  </Form.Item>;

  return (
    <PrivateStudioAssetEditor
      open={open}
      title={isNew ? title : `编辑资产 · ${entityName(fields) || formatEntity(entity) || '未命名'}`}
      onClose={onClose}
      width={920}
      fullscreen={fullscreen}
      basic={<>
        <div className="private-studio-asset-editor-kind">资产分类<b>{type === 'characters' ? '人物' : '场景'}</b></div>
        {basicFields.length ? basicFields.map(renderField) : null}
      </>}
      settings={settingFields.length ? <>
        {type === 'characters' ? <Typography.Text type="secondary">「外形（剧本）」与漫画侧资产里的「人物外形」是两个独立字段，分开保存，互不覆盖。</Typography.Text> : null}
        {settingFields.map(renderField)}
      </> : <Typography.Text type="secondary">该类型没有独立设定字段。</Typography.Text>}
      prompts={<>
        <Typography.Text type="secondary">剧本侧的图片生成提示词由「设定」与原文在生成时自动拼装，这里不单独保存提示词。</Typography.Text>
        <Space wrap>
          <Button onClick={enrich} loading={enriching} disabled={!canEnrich}>根据小说智能补全</Button>
          {!String(novelText || '').trim() ? <Typography.Text type="secondary">请先输入小说原文</Typography.Text> : null}
        </Space>
        {enrichmentError ? <Typography.Paragraph type="danger" style={{ margin: 0 }}>{enrichmentError}</Typography.Paragraph> : null}
        {enrichment ? <div className="entity-enrichment-result">
          {enrichment.evidence.length ? <EnrichmentList title="原文依据" items={enrichment.evidence} /> : null}
          {enrichment.suggestions.length ? <EnrichmentList title="AI 建议" items={enrichment.suggestions} /> : null}
          {enrichment.uncertainties.length ? <EnrichmentList title="不确定项" items={enrichment.uncertainties} /> : null}
        </div> : null}
      </>}
      aside={<EntityImagePanel
        key={imageRequestKey}
        assetType={type === 'characters' ? 'character' : 'scene'}
        assetId={assetId}
        generationPayload={{ ...buildReferenceAssetGenerationPayload({ type, entity, fields, novelText, extractionPreset }), asset_id: assetId }}
        imageModelId={imageModelId}
        imageProvider={imageProvider}
        imageAspectRatio={imageAspectRatio}
        imageUrls={imageUrls}
        mainImageUrl={mainImageUrl}
        requestKey={imageRequestKey}
        onChange={updateImages}
        disabled={!assetId}
      />}
      footer={<Space><Button onClick={onToggleFullscreen}>{fullscreen ? '退出全屏' : '全屏编辑'}</Button>{!isNew && <Popconfirm title={`确认${deleteLabel}？`} onConfirm={onDelete}><Button danger className="entity-editor-danger">{deleteLabel}</Button></Popconfirm>}<Button onClick={onClose}>取消</Button><Button type="primary" onClick={() => { onChange(fields, { imageUrls, mainImageUrl }); onClose(); }}>保存</Button></Space>}
    />
  );
}

function EnrichmentList({ title, items }) {
  return <div style={{ marginTop: 12 }}><Typography.Text strong>{title}</Typography.Text><ul style={{ margin: '6px 0 0', paddingLeft: 20 }}>{items.map(item => <li key={item}>{item}</li>)}</ul></div>;
}
