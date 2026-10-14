import { h3PromptEditRequest } from './h3PromptEditing.js';
import { measureH3VideoLines } from './h3LineAudio.js';
import { buildBatchFactoryH3Constraints } from './batchFactoryH3Constraints.js';
import { prependSmartUnifiedStyle, smartUnifiedAnalysisForBook, smartUnifiedDisplayEnabled } from './batchFactorySmartUnified.js';
import { publicationMetadataValue } from './batchFactoryPublicationMetadata.js';
import { resolvePrecompiledStoryboardAssets, resolvePrecompiledStoryboardFrame, resolvePrecompiledVideoWorkspace } from './batchFactoryPrecompiledStoryboard.js';
import {
  ArrowLeftOutlined,
  BarsOutlined,
  CloudUploadOutlined,
  FileTextOutlined,
  LoadingOutlined,
  ReloadOutlined,
	LeftOutlined,
  PictureOutlined,
	RightOutlined,
  SettingOutlined,
  UploadOutlined
  ,DeleteOutlined
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Descriptions,
  Dropdown,
  Input,
  InputNumber,
  Image,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Tag,
  Tooltip,
  message
} from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  createBookAsset,
	createNovelFetchIntake,
  appendNovelFetchIntake,
	backfillBatch121Metadata,
	  cancelBatchProduction,
  fetchBookOriginal,
  getCapabilities,
  getBatch,
  generateBookAssetImages,
  listBookAssets,
  listBookAssetImages,
  getConfigVersions,
  getFinalPrompt,
  getH3Trace,
	measureH3Audio,
	compileH3Video,
  getDraft,
  getBatchRuntimeSummary,
  listAutomationPresets,
  startBatchAutomation,
  pauseBatchAutomation,
  resumeBatchAutomation,
  retryBatchAutomation,
  cancelBatchAutomation,
  getChangeImpact,
  rewriteWorkingFront,
  listSystemPresetCatalog,
  runBatchDirector,
  runBookStage,
  retryBookStage,
  refreshBookSmartUnified,
  saveBatchSettings,
  saveBookOverride,
  updateBookAsset,
  updateBookMetadata,
  updateBookSource,
  uploadBookAssetImage,
  uploadBookVideoMaterial,
  uploadBookMergeMaterial,
  setPrimaryBookAssetImage,
  saveDraft,
	  saveVideoOverride,
  deleteProductionTask,
	deleteBatchFactoryBook,
  submitBatchMerge,
  submitBookMerge,
  submitBatchProduction,
  submitBookTo121,
  classifyBookPublishMetadata,
  get121OrganizationOptions
} from '../../../shared/api/batchFactoryV11';
import { getConfig } from '../../../shared/api/config';
import { textToSpeech } from '../../../shared/api/tts';
import { BatchFactoryUnifiedSettingsModal } from './BatchFactoryUnifiedSettingsModal';
import { BatchFactoryBookSettingsModal } from './BatchFactoryBookSettingsModal';
import { audioPlanningRequired, normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { BatchFactoryGiantMaterialPendingProgress } from './BatchFactoryGiantMaterialPendingProgress';
import { batchFactoryBatchProgress, batchFactoryBookState, batchFactoryBookTimeline, batchFactoryNovelTableRow, batchFactoryVideoProgress, batchFactoryVisibleError } from './batchFactoryBookState';
import { batchFactoryCleanSourceText, batchFactoryPreviewText, batchFactoryProductionText, contentRangeLinesForBook, sourceTextAfterViralAdoption } from './batchFactoryContentRange';
import { batchMediaCounts as runtimeBatchMediaCounts, resolveBookProductionText as runtimeResolveBookProductionText } from './batchFactoryRuntimeLogic';
import { downloadProductionContent } from './batchFactoryContentDownload.js';
import { realVideoDurationSeconds } from './batchFactoryVideoMeta';
import { getWorkshopPlatforms } from '../../../shared/api/novelFetchWorkshop';
import { ProductionMediaBoundary } from '../batch-factory-v11/ProductionMediaBoundary';
import { checkWebSubmitEnvironment, getWebSubmitConfig, testWebSubmitVisible } from '../../../shared/api/novelFetch';
import { batchFactoryPlatformOptions } from './batchFactoryPlatformOptions';
import { BOOK_CONFIG_REGIONS, bookAssetSummary, bookConfigRegionStatus } from './batchFactoryBookConfigRegions';
import { videoProviderForModel } from './videoProviderBinding';
import { stripBaseSetupSection } from '../scriptFinalSegment';
import { formatBeijingDatetimeLocal, normalizeAutomationConcurrency, parseBeijingDatetimeLocal } from './batchFactoryAutomationSchedule';


const DEFAULT_TTS = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 };
const UNIFIED_CONFIGURATION_KEYS = [
  'textModelId', 'imageModelId', 'videoModelId', 'videoProvider', 'aspectRatio', 'imageAspectRatio', 'videoAspectRatio', 'videoResolution', 'productionMode',
  'storyboardDurationLimit', 'maxVideoDuration', 'audioPlanningEnabled',
  'audioSource', 'audioMergeEnabled', 'tts', 'publishRewriteEnabled', 'publishSettings', 'aiPromptConfig', 'openingEnabled', 'openingCount'
];
function readBatchFactoryAudioDuration(blob) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(blob);
    const audio = new Audio();
    const clear = () => { audio.removeAttribute('src'); URL.revokeObjectURL(url); };
    audio.preload = 'metadata';
    audio.onloadedmetadata = () => {
      const duration = Number(audio.duration || 0);
      clear();
      duration > 0 && Number.isFinite(duration) ? resolve(duration) : reject(new Error('无法读取配音真实时长'));
    };
    audio.onerror = () => { clear(); reject(new Error('配音音频无法读取')); };
    audio.src = url;
  });
}
function batchFactoryAudioFingerprint(input, tts = {}) {
  const payload = JSON.stringify([String(input || ''), String(tts.voice || ''), String(tts.style || ''), Number(tts.speed || 0), Number(tts.pitch || 0)]);
  let hash = 2166136261;
  for (let index = 0; index < payload.length; index += 1) {
    hash ^= payload.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return `a1-${(hash >>> 0).toString(16)}-${payload.length}`;
}
function batchFactoryBookAudioInput(book = {}) {
  return batchFactoryProductionText(book?.sourceText || book?.contentPreview, book?.workingFrontContent, book);
}
async function batchFactoryBookTts(settings = {}) {
  const config = await getConfig().catch(() => ({}));
  return { ...DEFAULT_TTS, ...(config?.tts || {}), ...(settings?.tts || {}) };
}
async function generateBatchFactoryBookAudioMeasurement(book, settings = {}) {
  const input = batchFactoryBookAudioInput(book);
  if (!input) throw new Error('当前书没有可用于配音的生产内容');
  const tts = await batchFactoryBookTts(settings);
  const blob = await textToSpeech({ input, ...tts });
  const duration = await readBatchFactoryAudioDuration(blob);
  const lines = String(input).replace(/\r\n/g, '\n').split('\n').filter(line => line.trim());
  const lineDurations = [];
  for (let index = 0; index < lines.length; index += 1) {
    const lineBlob = await textToSpeech({ input: lines[index], ...tts });
    const lineDuration = await readBatchFactoryAudioDuration(lineBlob);
    lineDurations.push({ sourceIndex: index + 1, sourceText: lines[index], durationMs: Math.max(1, Math.round(lineDuration * 1000)) });
  }
  return {
    durationSeconds: Number(duration.toFixed(2)),
    fingerprint: batchFactoryAudioFingerprint(input, tts),
    lineDurations,
    tts,
    blob
  };
}
function blobToBase64(blob) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onerror = () => reject(new Error('无法读取 H3 配音数据'));
		reader.onload = () => resolve(String(reader.result || '').split(',').at(-1) || '');
		reader.readAsDataURL(blob);
	});
}
async function saveBookOverrideWithRetry(batchId, bookId, expectedRevision, input) {
  try {
    return await saveBookOverride(batchId, bookId, { ...input, expectedRevision: Number(expectedRevision || 0) });
  } catch (error) {
    if (Number(error?.status) !== 409) throw error;
    const latest = await getBatch(batchId);
    const latestBatch = latest?.batch || latest;
    const latestBook = (latestBatch?.books || []).find(item => item?.id === bookId);
    if (!latestBook) throw error;
    return saveBookOverride(batchId, bookId, { ...input, expectedRevision: Number(latestBook.revision || 0) });
  }
}
function batchFactoryContentMode(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const stored = String(metadata.productionContentMode || '').trim().toLowerCase();
  if (stored === 'viral' || stored === 'custom' || stored === 'original') return stored;
  const working = String(book?.workingFrontContent || '').trim();
  const source = String(book?.sourceText || '').trim();
  return !working || (source && working === source) ? 'original' : 'custom';
}
function batchFactoryContentModeLabel(book = {}) {
  const mode = batchFactoryContentMode(book);
  return mode === 'viral' ? '爆款' : mode === 'custom' ? '自创' : '原文';
}
function batchFactoryWorkbenchConfigRegions(book) {
  const base = BOOK_CONFIG_REGIONS.filter(region => !['video', 'visual'].includes(region.key));
  const video = bookConfigRegionStatus(book, 'video');
  const visual = bookConfigRegionStatus(book, 'visual');
  const labels = [...new Set([video?.label, visual?.label].filter(Boolean))];
  const tone = video?.tone === visual?.tone ? (video?.tone || 'muted') : (video?.tone || visual?.tone || 'muted');
  return [...base, { key: 'media', label: '画面与视频', combinedStatus: { tone, label: labels.join(' / ') || '已配置' } }];
}

function value(metadata, key) { return String(metadata?.[key] || '').trim() || '—'; }
function bookPlatformName(book, platformNames = {}) {
  const metadata = book?.sourceMetadata || {};
  return String(metadata.platformName || metadata.platformLabel || platformNames[String(book?.platform || '')] || '').trim() || '未命名书城';
}
function resultData(result, key) { return result?.[key] || result || {}; }
function sameSettingsPatch(left, right) {
  return JSON.stringify(left || {}) === JSON.stringify(right || {});
}
function h3DirectorCards(book) {
  const document = h3DirectorDocument(book);
  return Array.isArray(document.director_cards) ? document.director_cards : Array.isArray(document.directorCards) ? document.directorCards : [];
}
function h3DirectorDocument(book) {
  const output = book?.directorRevision?.output || {};
  return output.h3_director || output.h3Director || {};
}
function h3Text(value) { return String(value || '').trim(); }
function h3NumberText(value) {
  const number = Number(value);
  return Number.isFinite(number) ? String(number) : '';
}
function h3Value(object, snakeKey, camelKey = '') {
  return object?.[snakeKey] ?? (camelKey ? object?.[camelKey] : undefined);
}
function h3CardCharacterNames(document, slotIDs = []) {
  const roster = h3Value(document, 'character_roster', 'characterRoster');
  const bySlot = new Map((Array.isArray(roster) ? roster : []).map(character => [h3Text(h3Value(character, 'slot_id', 'slotId')), character]));
  return (Array.isArray(slotIDs) ? slotIDs : []).map(slotID => {
    const character = bySlot.get(h3Text(slotID));
    const name = h3Text(h3Value(character, 'canonical_name', 'canonicalName'));
    return name ? `${slotID}(${name})` : h3Text(slotID);
  }).filter(Boolean).join('、');
}
function h3AudioText(audio = {}, document = {}) {
  const speaker = h3Text(h3Value(audio, 'speaker_slot_id', 'speakerSlotId'));
  const speakerName = h3CardCharacterNames(document, speaker ? [speaker] : []);
  const dialogue = h3Text(audio.dialogue);
  const voiceOver = h3Text(h3Value(audio, 'voice_over', 'voiceOver'));
  const effects = h3Value(audio, 'sound_effects', 'soundEffects');
  const ambience = audio.ambience;
  const parts = [
    speakerName ? `说话人：${speakerName}` : '',
    dialogue ? `对白：${dialogue}` : '',
    voiceOver ? `旁白：${voiceOver}` : '',
    Array.isArray(effects) && effects.length ? `音效：${effects.filter(Boolean).join('、')}` : '',
    Array.isArray(ambience) && ambience.length ? `环境：${ambience.filter(Boolean).join('、')}` : h3Text(audio.ambient) ? `环境：${h3Text(audio.ambient)}` : ''
  ].filter(Boolean);
  return parts.join('；');
}
function formatH3DirectorCardPrompt(document = {}, card = {}) {
  const roster = h3Value(document, 'character_roster', 'characterRoster');
  const subjects = (Array.isArray(roster) ? roster : []).map((character, index) => {
    const name = h3Text(h3Value(character, 'canonical_name', 'canonicalName')) || h3Text(h3Value(character, 'slot_id', 'slotId'));
    const appearance = h3Text(character.appearance);
    return name ? `<Subject ${index + 1}> ${name}${appearance ? `：${appearance}` : ''}` : '';
  }).filter(Boolean);
  const sourceIndex = Number(card.source_index ?? card.sourceIndex) || 1;
  const scene = h3Text(h3Value(card, 'scene_description', 'sceneDescription')) || h3Text(card.visual_context ?? card.visualContext);
  const sourceText = h3Text(card.source_text ?? card.sourceText);
  const visual = h3Text(card.visual_context ?? card.visualContext);
  const action = h3Text(card.action);
  const characters = h3CardCharacterNames(document, h3Value(card, 'character_slot_ids', 'characterSlotIds'));
  const camera = card.camera || {};
  const movement = card.movement || {};
  const cameraParts = [h3Text(h3Value(camera, 'shot_size', 'shotSize')), h3Text(h3Value(camera, 'shot_angle', 'shotAngle')), h3Text(camera.framing)].filter(Boolean);
  const movementParts = [h3Text(h3Value(movement, 'camera_movement', 'cameraMovement')), h3Text(h3Value(movement, 'subject_movement', 'subjectMovement')), h3Text(movement.transition)].filter(Boolean);
  const duration = h3NumberText(card.preferred_duration ?? card.preferredDuration ?? card.duration);
  const lines = [];
  if (subjects.length) lines.push('subject_definitions:', ...subjects, '');
  lines.push(`[Scene ${sourceIndex}] ${scene || '未命名场景'}`);
  if (sourceText) lines.push(`Event: ${sourceText}`);
  if (visual) lines.push(`导演调度：${visual}`);
  if (duration) lines.push(`Total duration: ${duration} seconds.`);
  lines.push('[Shot 1]');
  if (characters) lines.push(`人物：${characters}`);
  if (visual) lines.push(`画面：${visual}`);
  if (action) lines.push(`动作：${action}`);
  if (cameraParts.length) lines.push(`机位：${cameraParts.join('，')}`);
  if (movementParts.length) lines.push(`运镜：${movementParts.join('，')}`);
  if (h3Text(card.rhythm)) lines.push(`节奏：${h3Text(card.rhythm)}`);
  const audio = h3AudioText(card.audio || {}, document);
  if (audio) lines.push(`Audio: ${audio}`);
  return lines.join('\n');
}
function finalVideoPromptTemplate(body) {
  const value = String(body || '');
  const marker = '【批量工厂最终 Prompt 模板】';
  const index = value.indexOf(marker);
  return index >= 0 ? value.slice(index + marker.length).trim() : '';
}
function requestID(prefix) { return `${prefix}-${globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`}`; }
function stageActionKey(stage, bookId) { return `stage-${stage}:${String(bookId || '')}`; }
function capability(caps, key) { return caps?.[key] || { available: false, reason: '正在读取 V12 服务能力' }; }
function assetName(asset) { return String(asset?.name || asset?.label || asset?.id || '未命名预设'); }
function assetPrompt(asset) { return String(asset?.prompt || asset?.visualPrompt || asset?.description || ''); }
function savedDraftKey(type, asset) { return `asset:${type}:${asset?.id || assetName(asset)}`; }
function effectiveBookSettings(batch, book) {
  const batchPatch = batch?.settingsState?.patch || {};
  const bookPatch = book?.settingsState?.patch || {};
  return normalizeAudioSourceSettings({ ...batchPatch, ...bookPatch, publishSettings: { ...(batchPatch.publishSettings || {}), ...(bookPatch.publishSettings || {}) } });
}
function baseSetupEnabled(settings = {}) {
  return settings?.aiPromptConfig?.constraints?.baseSetup?.enabled !== false;
}
function h3CompilationStatus(settings = {}) {
  if (audioPlanningRequired(settings)) return '等待真实配音时长编译最终 VIDEO';
  const seconds = Number(settings.storyboardDurationLimit) === 15 ? 15 : 10;
  return `将按 ${seconds} 秒确定性切段并编译最终 VIDEO`;
}
function h3ConstraintPreview(settings = {}) {
  const constraints = settings?.aiPromptConfig?.constraints || {};
  const compiled = buildBatchFactoryH3Constraints(constraints);
  const rows = [];
  if (compiled.switches.smart_unified) rows.push('智能统一：由智能统一结果注入');
  if (compiled.switches.base_setup) rows.push('基础设定：人物 / 场景资产');
  if (compiled.switches.prefix && compiled.prefix_text) rows.push(`画面前缀词：${compiled.prefix_text}`);
  if (compiled.switches.quality && compiled.quality_text) rows.push(`画质约束：${compiled.quality_text}`);
  if (compiled.switches.visual_restriction && compiled.visual_restriction_text) rows.push(`画面限制：${compiled.visual_restriction_text}`);
  if (compiled.switches.negative && compiled.negative_text) rows.push(`负面提示词：${compiled.negative_text}`);
  return rows;
}
function storyboardAssetPromptText(book = {}) {
  return (Array.isArray(book?.assetRecords) ? book.assetRecords : [])
    .map(asset => {
      const structured = [
        asset?.formalName ? `正式名称：${asset.formalName}` : '',
        asset?.facts ? `人物事实：${asset.facts}` : '',
        asset?.appearance ? `人物外形：${asset.appearance}` : '',
        asset?.relationships ? `人物关系：${asset.relationships}` : ''
      ].filter(Boolean);
      const prompt = assetPrompt(asset);
      if (prompt && !structured.some(value => value.endsWith(prompt))) structured.push(`资产提示词：${prompt}`);
      if (!structured.length) return '';
      // Keep every asset as a separate visual block.  The rich prompt renderer
      // preserves these newlines, so the next character/scene/prop cannot be
      // glued to the previous asset's description.
      return `${assetKindLabel(asset?.kind)}「@${assetName(asset)}」\n${structured.join('\n')}`;
    })
    .filter(Boolean)
    .join('\n\n');
}

function composeStoryboardDisplayPrompt(prompt, assetText) {
  const body = String(prompt || '').trim();
  const assets = String(assetText || '').trim();
  if (!assets || body.includes('【基础设定】') || body.includes('基础设定：')) return body;
  return `【基础设定】\n${assets}${body ? `\n\n${body}` : ''}`;
}
function assetKindLabel(kind) {
  return kind === 'character' ? '人物' : kind === 'scene' ? '场景' : kind === 'prop' ? '道具' : '资产';
}
function promptPreview(value, limit = 180) {
  const text = String(value || '').replace(/\s+/g, ' ').trim();
  return text.length > limit ? `${text.slice(0, limit)}…` : text;
}
function selectedVideoPromptLabel(settings = {}) {
  const video = settings?.aiPromptConfig?.video || {};
  const id = String(video.presetId || video.presetKey || '').trim();
  if (!id || id === 'batch-video-meta') return 'SD 视频提示词';
  return String(video.presetName || id).trim() || 'SD 视频提示词';
}
function usesH3VideoPrompt(settings = {}) {
  const video = settings?.aiPromptConfig?.video || {};
  const id = String(video.presetId || video.presetKey || '').trim();
  return id === 'batch-video-h3-director' || id === 'h3-video-normal';
}
function effectiveBookAssetRules(batch, book) {
  return {
    ...(batch?.settingsState?.patch?.aiPromptConfig?.assets || {}),
    ...(book?.settingsState?.patch?.aiPromptConfig?.assets || {})
  };
}
function selectedPublishProfile(settings = {}) {
  const wanted = String(settings.websiteProfileId || '').trim();
  return (Array.isArray(settings.websiteProfiles) ? settings.websiteProfiles : [])
    .find(profile => String(profile?.id || '').trim() === wanted) || {};
}
function decompressionCountForUpload(batch, book) {
  const settings = effectiveBookSettings(batch, book).publishSettings || {};
  const profile = selectedPublishProfile(settings);
  const candidate = settings?.advanced?.jieyaNum ?? profile?.advanced?.jieyaNum;
  const numeric = Number(candidate);
  return Number.isFinite(numeric) ? Math.max(0, Math.round(numeric)) : 4;
}
function isBookUploadedTo121(book) {
  return String(book?.sourceMetadata?.websiteSubmitStatus || '').trim().toLowerCase() === 'uploaded';
}
function uploadProgressForBook(book) {
  const value = book?.sourceMetadata?.websiteSubmitProgress;
  return value && typeof value === 'object' ? value : null;
}
function sourcePlatformIdForUpload(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const candidates = [book?.platform, metadata.platformId, metadata.platform_id, metadata.platformName, metadata.platformLabel];
  const labels = { '黑岩付费': '1', '番茄付费': '2', '七猫付费': '3', '点众付费': '4', '阅文付费': '6', '番茄免费': '7', '知乎付费': '15', '掌阅付费': '20', '卓越付费': '26', '九州书城': '29', '掌文付费': '31' };
  return candidates.map(value => String(value || '').trim()).find(value => /^(1|2|3|4|6|7|15|20|26|29|31)$/.test(value))
    || candidates.map(value => labels[String(value || '').trim()]).find(Boolean)
    || '';
}
function completePublishMapping(_settings = {}, book = {}) {
  // 121 platform derives from the source book. Gender/style/tags are per-book
  // classification data and the server fills missing values immediately before
  // submit; a batch profile may only supply presentation/advanced options.
  return Boolean(sourcePlatformIdForUpload(book));
}
function classificationLabel(metadata = {}) {
  const status = String(metadata?.classifyStatus || '').trim();
  if (status === 'classified') return '已识别';
  if (status === 'manual') return '已手调';
  if (status === 'failed') return '识别失败';
  return '待识别';
}
function classificationDetail(metadata = {}) {
  if (String(metadata?.classifyStatus || '').trim() === 'failed') return String(metadata?.classifyError || '男女频和风格识别未完成，请重试。').trim();
  if (String(metadata?.classifyStatus || '').trim() === 'classified') return String(metadata?.classifyReason || '已根据当前书正文保存识别结果。').trim();
  return '尚未发起识别。';
}
const BATCH_FACTORY_TABLE_COLUMNS = [
  { label: '序号', minWidth: 56 },
  { label: '小说正文', minWidth: 300 },
  { label: '单书配置', minWidth: 176 },
  { label: '分镜提示词', minWidth: 330 },
  { label: '分镜视频', minWidth: 300 },
  { label: '操作', minWidth: 220 }
];
function completedMediaVersions(status) {
	const entries = new Map();
	for (const job of status?.jobs || []) for (const task of job?.tasks || []) {
		if (task?.status !== 'succeeded' || !String(task?.mediaUrl || '').trim()) continue;
		const versions = entries.get(task.videoId) || [];
		versions.push({ ...task, productionJobId: job.id });
		entries.set(task.videoId, versions);
	}
	return entries;
}

function primaryMediaVersion(video, versions) {
	const primaryID = video?.settingsState?.patch?.primaryMediaTaskId;
	return versions.find(item => item.id === primaryID) || versions.at(-1) || null;
}

function storyboardVideoLabel(video, index) {
  const ordinal = Number(index) + 1;
  const label = String(video?.label || video?.title || '').trim();
  return label ? `分镜 ${ordinal} · ${label}` : `分镜 ${ordinal}`;
}

function bookMergeJobs(status, bookId) {
  return (status?.jobs || []).filter(job => job?.bookId === bookId).sort((left, right) => String(left?.updatedAt || left?.createdAt || '').localeCompare(String(right?.updatedAt || right?.createdAt || '')));
}

function mergeJobSpeed(job) {
  const speed = Number(job?.speed || 0);
  return speed > 0 ? speed : 1;
}

function defaultBookMerge(jobs) {
  const completed = jobs.filter(job => job?.status === 'succeeded' && String(job?.outputUrl || '').trim());
  return [...completed].reverse().find(job => Math.abs(mergeJobSpeed(job) - 1) < 0.001) || completed.at(-1) || null;
}

function latestBookMerge(status, bookId) {
  const jobs = bookMergeJobs(status, bookId);
  return [...jobs].reverse().find(job => job?.status === 'running' || job?.status === 'queued') || defaultBookMerge(jobs) || jobs.at(-1) || null;
}

function bookVideoReady(book, productionStatus) {
  const videoIds = (book?.videos || []).map(video => String(video?.id || '')).filter(Boolean);
  if (!videoIds.length) return false;
  const completed = new Set((productionStatus?.jobs || [])
    .filter(job => job?.bookId === book?.id)
    .flatMap(job => job?.tasks || [])
    .filter(task => task?.status === 'succeeded' && String(task?.mediaUrl || '').trim())
    .map(task => String(task?.videoId || '')));
  return videoIds.every(videoId => completed.has(videoId));
}

function NovelMetadata({ books, createdAt, selectedBookIds, onSelectionChange, onViewBook, platformNames, platformOptions, productionStatus, mergeStatus, stageSummaries, batchId, onContentReady, onBackfill121Metadata, backfilling121Metadata, onAddBooks }) {
  const selected = new Set(selectedBookIds);
  const allSelected = books.length > 0 && books.every(book => selected.has(book.id));
  const selectedCount = selectedBookIds.length;
  const toggleAll = checked => onSelectionChange(checked ? books.map(book => book.id) : []);
  const pendingContentBooks = books.filter(book => book?.sourceMetadata?.contentPending);
  return <div className="batch-factory-novel-list batch-factory-novel-fetch-list batch-factory-novel-list-v13" role="table" aria-label="小说列表">
    {pendingContentBooks.length ? <div className="batch-factory-novel-pending-strip" role="status">
      <span className="batch-factory-novel-pending-count">正在读取 {pendingContentBooks.length} 本书正文</span>
      {pendingContentBooks.map(book => <BatchFactoryGiantMaterialPendingProgress key={book.id} book={book} batchId={batchId} variant="banner" onContentReady={onContentReady} />)}
    </div> : null}
    <div className="batch-factory-novel-list-toolbar">
      <Space size={8}>
        <Checkbox checked={allSelected} indeterminate={!allSelected && selected.size > 0} onChange={event => toggleAll(event.target.checked)} aria-label="全选小说"><b>序号</b></Checkbox>
        {selectedCount > 0 && <span className="batch-factory-novel-selected-count">已选 {selectedCount} 本</span>}
		<Button size="small" loading={backfilling121Metadata} disabled={!batchId || backfilling121Metadata} onClick={onBackfill121Metadata}>补全 121 男女频</Button>
      </Space>
    </div>
    <div className="batch-factory-novel-list-head" role="row">
      <span className="is-order">#</span>
      <span>小说</span><span>内容</span><span>风格</span><span>男女频</span><span>状态</span><span>操作</span>
    </div>
    {books.map((book, index) => {
      const row = batchFactoryNovelTableRow(book, index, createdAt, { productionStatus, mergeStatus, stageSummary: stageSummaries?.[book.id] });
      const metadata = book.sourceMetadata || {};
      const mode = batchFactoryContentMode(book);
      return <div className="batch-factory-novel-list-row" key={book.id} role="row">
        <span className="is-order"><Checkbox checked={selected.has(book.id)} onChange={event => onSelectionChange(event.target.checked ? [...selected, book.id] : [...selected].filter(id => id !== book.id))} aria-label={`选择 ${row.title}`} /><b>{String(index + 1).padStart(2, '0')}</b></span>
        <span className="is-book"><strong title={row.title}>{row.title}</strong><small>ID {row.bookId} · {bookPlatformName(book, platformNames)}</small></span>
        <span><i className={`batch-factory-content-mode is-${mode}`}>{batchFactoryContentModeLabel(book)}</i></span>
        <span>{publicationMetadataValue(metadata, 'style')}</span>
        <span>{publicationMetadataValue(metadata, 'gender')}</span>
        <span className={row.status === '处理中' ? 'is-scheduled' : ''} title={row.state?.detail}>{row.status}{row.state?.manual ? ' · 已手调' : ''}</span>
        <span><Button size="small" onClick={() => onViewBook(book)}>查看</Button></span>
      </div>;
    })}
    <AddBooksRow books={books} platformOptions={platformOptions} onAddBooks={onAddBooks} />
  </div>;
}

function AddBooksRow({ books, platformOptions = [], onAddBooks }) {
  const [open, setOpen] = useState(false);
  const [inputText, setInputText] = useState('');
  const [platform, setPlatform] = useState('');
  const [allowDuplicate, setAllowDuplicate] = useState(false);
  const [busy, setBusy] = useState(false);
  const options = Array.isArray(platformOptions) && platformOptions.length
    ? platformOptions
    : batchFactoryPlatformOptions([], { fallback: true });
  useEffect(() => {
    if (!platform && options.length) setPlatform(String(options[0].value));
  }, [options, platform]);
  const parsed = useMemo(() => String(inputText || '').split(/\r?\n/).map(line => line.trim()).filter(Boolean).map(line => {
    const match = line.match(/^(\d{1,25})(?:\s+|[，,、|])?(.*)$/);
    return match ? { bookId: match[1], title: String(match[2] || '').trim() || `小说 ${match[1]}` } : null;
  }).filter(Boolean), [inputText]);
  const existing = useMemo(() => new Set(books.map(book => String(book?.bookId || '').trim()).filter(Boolean)), [books]);
  const duplicates = parsed.filter(item => existing.has(item.bookId));
  async function submit() {
    if (!parsed.length) { message.warning('请按每行“书 ID”或“书 ID + 书名”输入'); return; }
    if (!platform) { message.warning('请选择书城'); return; }
    if (duplicates.length && !allowDuplicate) { message.warning(`发现 ${duplicates.length} 本已存在的书，勾选“重复 ID 仍然添加”后再提交`); return; }
    setBusy(true);
    try {
      await onAddBooks(parsed, platform, allowDuplicate);
      setInputText('');
      setAllowDuplicate(false);
      setOpen(false);
    } catch (error) {
      message.error(error?.message || '添加新书失败');
    } finally { setBusy(false); }
  }
  return <>
    <div className="batch-factory-novel-list-row batch-factory-novel-add-row" role="row">
      <span className="is-order"><b>＋</b></span>
      <span className="is-book"><Button type="link" size="small" onClick={() => setOpen(true)}>添加新书</Button><small>输入多个书 ID 或“书 ID + 书名”</small></span>
      <span>—</span><span>—</span><span>—</span><span>—</span><span>选择书城后添加</span>
    </div>
    <Modal className="batch-factory-add-books-modal" title={<span className="batch-factory-add-books-title"><b>添加新书</b><small>写入当前批次小说列表</small></span>} open={open} onCancel={() => setOpen(false)} onOk={submit} confirmLoading={busy} okText="添加到当前批次" width={620} destroyOnClose>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Alert type="info" showIcon message="添加后自动获取正文与分类信息" description="新书会继承当前批次配置，自动获取正文并识别风格、男女频；不会加入正在运行的生产队列。" />
        <label className="batch-factory-engine-field"><span><b>书城</b></span><Select value={platform || undefined} onChange={value => setPlatform(String(value || ''))} options={options.map(item => ({ value: String(item.value), label: item.label }))} placeholder="请选择书城" style={{ width: '100%' }} /></label>
        <label className="batch-factory-engine-field"><span><b>书 ID / 书名</b><small>每行一本，例如：2091926811672429633\t书名</small></span><Input.TextArea rows={8} value={inputText} onChange={event => setInputText(event.target.value)} placeholder={'2091926811672429633\n209208048803840415\t第二本小说'} /></label>
        {duplicates.length ? <Alert type="warning" showIcon message={`发现 ${duplicates.length} 本已存在的书`} description={duplicates.map(item => item.bookId).join('、')} /> : null}
        <Checkbox checked={allowDuplicate} onChange={event => setAllowDuplicate(event.target.checked)}>重复 ID 仍然添加</Checkbox>
      </Space>
    </Modal>
  </>;
}

function readImageAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('读取图片失败'));
    reader.onload = () => resolve(String(reader.result || ''));
    reader.readAsDataURL(file);
  });
}

function AssetEditor({ book, batchId, onSaved, onGenerate, onRegenerate, onRetry, onAssetPromptChange, assetRules, canGenerate, generateReason, generating, engineSettings }) {
  const [assets, setAssets] = useState([]);
  const [loading, setLoading] = useState(false);
  const [savingId, setSavingId] = useState('');
  const [kind, setKind] = useState('character');
  const [libraryKind, setLibraryKind] = useState('character');
  const [search, setSearch] = useState('');
  const [assetPromptCatalog, setAssetPromptCatalog] = useState([]);
  const [assetPromptLoading, setAssetPromptLoading] = useState(false);
  const [selectedAssetIds, setSelectedAssetIds] = useState([]);
  const [imageGenerating, setImageGenerating] = useState(false);
  const [assetImages, setAssetImages] = useState({});
  const [editing, setEditing] = useState(null);
	const [starredCharacterNames, setStarredCharacterNames] = useState([]);
  const kindOptions = [{ value: 'character', label: '人物' }, { value: 'scene', label: '场景' }, { value: 'prop', label: '道具' }];
  const categoryLabel = value => kindOptions.find(item => item.value === value)?.label || '资产';
  async function loadImages(nextAssets, { quiet = false } = {}) {
    const entries = await Promise.all((nextAssets || []).map(async asset => {
      try { return [asset.id, resultData(await listBookAssetImages(batchId, book.id, asset.id), 'images')]; }
      catch (error) { if (!quiet) message.error(error?.message || '读取资产图片版本失败'); return [asset.id, []]; }
    }));
    setAssetImages(Object.fromEntries(entries.map(([assetId, images]) => [assetId, Array.isArray(images) ? images : []])));
  }
  const loadAssets = async ({ quiet = false } = {}) => {
    if (!batchId || !book?.id) return;
    setLoading(true);
    try {
      const response = await listBookAssets(batchId, book.id);
      const next = resultData(response, 'assets');
      const saved = Array.isArray(next) ? next : [];
      setAssets(saved);
      await loadImages(saved, { quiet: true });
    } catch (error) {
      if (!quiet) message.error(error?.message || '读取当前小说资产失败');
    } finally { setLoading(false); }
  };
  useEffect(() => {
    setAssets(Array.isArray(book?.assetRecords) ? book.assetRecords : []);
    setSearch('');
    setSelectedAssetIds([]);
    loadAssets({ quiet: true });
  }, [batchId, book?.id]);
	useEffect(() => { setStarredCharacterNames(Array.isArray(book?.settingsState?.patch?.starredCharacterNames) ? book.settingsState.patch.starredCharacterNames : []); }, [book?.id, book?.revision, book?.settingsState?.revision]);
  useEffect(() => {
    let active = true;
    setAssetPromptLoading(true);
    listSystemPresetCatalog('script').then(script => {
      if (active) setAssetPromptCatalog(
        (Array.isArray(script?.catalog) ? script.catalog : []).filter(item => item?.slot === 'script.asset-extraction')
      );
    }).catch(error => { if (active) message.error(error?.message || '读取资产提示词失败'); }).finally(() => { if (active) setAssetPromptLoading(false); });
    return () => { active = false; };
  }, [book?.id]);
  const textModelId = String(engineSettings?.textModelId || '').trim();
  const imageModelId = String(engineSettings?.imageModelId || '').trim();
  const imageAspectRatio = String(engineSettings?.imageAspectRatio || engineSettings?.aspectRatio || '9:16');
  const currentAssetPromptId = String(assetRules?.extraction?.presetId || '').trim();
  const assetPromptOptions = (currentAssetPromptId && !assetPromptCatalog.some(item => item.id === currentAssetPromptId)
    ? [{ id: currentAssetPromptId, name: assetRules?.extraction?.presetName || '预设已失效', version: assetRules?.extraction?.presetVersion || 1, disabled: true }, ...assetPromptCatalog]
    : assetPromptCatalog).map(item => ({ value: item.id, label: `${item.name || item.id} · v${item.version || 1}`, disabled: item.disabled === true }));
  const visibleAssets = assets.filter(asset => asset.kind === kind).filter(asset => {
    const query = search.trim().toLowerCase();
    return !query || `${asset.name} ${asset.prompt}`.toLowerCase().includes(query);
  });
  const libraryImages = assets.filter(asset => asset.kind === libraryKind).flatMap(asset => (assetImages[asset.id] || []).map(image => ({ image, asset }))).filter(({ asset }) => {
    const query = search.trim().toLowerCase();
    return !query || `${asset.name} ${asset.prompt}`.toLowerCase().includes(query);
  });
  async function saveAsset({ close = true } = {}) {
    if (!editing?.name?.trim() || !editing?.prompt?.trim()) { message.warning('请填写资产名称和提示词。'); return null; }
    const existing = assets.find(asset => asset.id === editing.id);
    setSavingId(existing?.id || 'new');
    try {
      const assetDetails = {
        name: editing.name,
        formalName: editing.formalName,
        facts: editing.facts,
        appearance: editing.appearance || editing.prompt,
        relationships: editing.relationships,
        prompt: editing.prompt
      };
      const result = existing
        ? await updateBookAsset(batchId, book.id, existing.id, { ...assetDetails, expectedRevision: Number(existing.revision || 0) })
        : await createBookAsset(batchId, book.id, { kind: editing.kind, ...assetDetails });
      const savedAsset = resultData(result, 'asset');
      if (close) setEditing(null);
      await loadAssets({ quiet: true });
      await onSaved?.();
      message.success('当前小说资产已保存。');
      return savedAsset?.id ? savedAsset : null;
    } catch (error) { message.error(error?.message || '保存资产失败'); return null; } finally { setSavingId(''); }
  }
  async function runPreset() {
    try {
      if (!textModelId) { message.warning('请先选择个人中心已启用的文本模型。'); return; }
      await onGenerate?.(textModelId);
      await loadAssets({ quiet: true });
    } catch (_) { /* runAi already reports its business error */ }
  }
  async function chooseAssetPrompt(presetId) {
    const preset = assetPromptCatalog.find(item => item.id === presetId);
    try {
      await onAssetPromptChange?.(preset ? { presetId: preset.id, presetName: preset.name, presetSlot: preset.slot, presetVersion: preset.version, constraintCategory: preset.constraintCategory || '' } : { presetId: '' });
      message.success(preset?.id === 'batch-assets-h3' ? '人物场景道具方案已选中，点击“重新生成资产”应用到当前书。' : '资产提示词已保存。');
    } catch (error) { message.error(error?.message || '保存资产提示词失败'); }
  }
  function toggleAssetSelection(assetId, checked) {
    setSelectedAssetIds(current => checked ? [...new Set([...current, assetId])] : current.filter(id => id !== assetId));
  }
	async function toggleCharacterStar(asset) {
		if (asset?.kind !== 'character') return;
		const name = assetName(asset);
		const starred = starredCharacterNames.includes(name);
		const next = starred ? starredCharacterNames.filter(item => item !== name) : [...new Set([...starredCharacterNames, name])];
		setSavingId(`star-${asset.id}`);
		try {
			await saveBookOverride(batchId, book.id, { patch: { starredCharacterNames: next }, expectedRevision: Number(book?.revision || 0) });
			setStarredCharacterNames(next);
			await onSaved?.();
			message.success(starred ? `已取消“${name}”的星标聚焦。` : `已将“${name}”设为星标聚焦人物。`);
		} catch (error) { message.error(error?.message || '保存星标人物失败'); } finally { setSavingId(''); }
	}
  async function generateImages(assetIds = selectedAssetIds) {
    if (!assetIds.length) { message.warning('请先选择要生成图片的资产。'); return false; }
    if (!imageModelId) { message.warning('请先在引擎配置选择图片模型。'); return false; }
    setImageGenerating(true);
    try {
      const result = await generateBookAssetImages(batchId, book.id, {
        assetIds,
        modelId: imageModelId,
        imageAspectRatio: imageAspectRatio || '9:16'
      });
      await loadAssets({ quiet: true });
      await onSaved?.();
      message.success(`已保存 ${resultData(result, 'images')?.length || assetIds.length} 个当前书资产图片版本。`);
      return true;
    } catch (error) { message.error(error?.message || '图片生成失败'); return false; } finally { setImageGenerating(false); }
  }
  async function generateEditedAssetImage() {
    const savedAsset = await saveAsset({ close: false });
    if (!savedAsset) return;
    if (await generateImages([savedAsset.id])) setEditing(null);
  }
  async function uploadImage(asset, event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) { message.error('仅支持 PNG、JPG、WEBP 图片。'); return; }
    setSavingId(`upload-${asset.id}`);
    try {
      await uploadBookAssetImage(batchId, book.id, asset.id, await readImageAsDataURL(file));
      await loadImages(assets, { quiet: true });
      await onSaved?.();
      message.success('图片版本已保存到当前小说资产。');
    } catch (error) { message.error(error?.message || '上传资产图片失败'); } finally { setSavingId(''); }
  }
  async function makePrimary(asset, image) {
    if (image.isPrimary) return;
    setSavingId(`primary-${image.id}`);
    try {
      await setPrimaryBookAssetImage(batchId, book.id, asset.id, image.id);
      await loadImages(assets, { quiet: true });
      await onSaved?.();
      message.success('已切换为当前资产主图。');
    } catch (error) { message.error(error?.message || '切换主图失败'); } finally { setSavingId(''); }
  }
  return <section className="batch-factory-preset-shell">
    <Alert type="info" showIcon message="当前小说的人物、场景与道具" description="每一项都归属当前小说并持久化保存。智能预设成功后会写入导演提取的资产；手动修改会保留为该书的人工资产。" />
    <div className="shuihuo-preset-toolbar batch-factory-preset-toolbar">
      <Tooltip title="选择当前书的人物、场景与道具提取提示词；模型统一在引擎配置设置。"><Select value={currentAssetPromptId || undefined} loading={assetPromptLoading} placeholder="选择资产提示词" style={{ minWidth: 230 }} options={assetPromptOptions} onChange={chooseAssetPrompt} /></Tooltip>
      <Tooltip title={canGenerate ? '仅在当前书没有文案时提取人物、场景、道具；已有手动资产不会被覆盖。' : generateReason}><Button type="primary" loading={generating} disabled={!canGenerate || generating} onClick={runPreset}>智能预设</Button></Tooltip>
      <Tooltip title="基于当前书内容重新生成 AI 文案和自动资产；手动编辑的资产提示词保留。"><Button loading={generating} disabled={!canGenerate || generating || !textModelId} onClick={() => onRegenerate?.(textModelId)}>重新生成资产</Button></Tooltip>
      <Tooltip title="只重跑当前小说最后失败的阶段。"><Button disabled={!canGenerate || generating || !textModelId} onClick={() => onRetry?.(textModelId)}>重试资产</Button></Tooltip>
      <Button loading={loading} onClick={() => loadAssets()}>刷新资产</Button>
      <Tooltip title="使用引擎配置中的图片模型和画幅，为选中的当前书资产生成新版本；不会覆盖已有主图。"><Button type="primary" loading={imageGenerating} disabled={imageGenerating || !selectedAssetIds.length || !imageModelId} onClick={() => generateImages()}>AI生成（已选 {selectedAssetIds.length}）</Button></Tooltip>
      <Button onClick={() => setEditing({ kind, name: '', prompt: '' })}>手动添加</Button>
    </div>
    {!imageModelId ? <Alert type="warning" showIcon message="尚未配置图片模型" description="请在该书的引擎配置中选择图片模型后再生成资产图片。" /> : null}
    <div className="shuihuo-preset-layout batch-factory-preset-layout">
      <aside className="shuihuo-preset-list">
        <div className="shuihuo-preset-list-head"><strong>资产列表 ({visibleAssets.length})</strong><Select size="small" value={kind} onChange={setKind} options={kindOptions} /><Input.Search value={search} onChange={event => setSearch(event.target.value)} placeholder="搜索名称或提示词" allowClear /></div>
        <div className="shuihuo-prompt-list">
          {visibleAssets.map(asset => <article className={`shuihuo-prompt-row${selectedAssetIds.includes(asset.id) ? ' is-selected' : ''}${asset.kind === 'character' && starredCharacterNames.includes(assetName(asset)) ? ' is-starred' : ''}`} key={asset.id}><div className="shuihuo-prompt-row-head"><Checkbox checked={selectedAssetIds.includes(asset.id)} onChange={event => toggleAssetSelection(asset.id, event.target.checked)}>生成</Checkbox><Tag>{categoryLabel(asset.kind)}</Tag><strong>{assetName(asset)}</strong>{asset.kind === 'character' ? <Tooltip title={starredCharacterNames.includes(assetName(asset)) ? '取消星标人物聚焦' : '设为星标人物聚焦'}><Button type="text" size="small" className="batch-factory-character-star" loading={savingId === `star-${asset.id}`} onClick={() => toggleCharacterStar(asset)} aria-label={`${assetName(asset)} 星标人物聚焦`}>★</Button></Tooltip> : null}<Button type="text" size="small" onClick={() => setEditing({ ...asset })}>编辑</Button></div><p>{assetPrompt(asset) || '尚未填写可视化提示词'}</p></article>)}
          {!loading && !visibleAssets.length ? <div className="shuihuo-prompt-empty"><b>＋</b><p>暂无提示词，请点击“智能预设”或手动添加。</p><Button onClick={() => setEditing({ kind, name: '', prompt: '' })}>添加提示词</Button></div> : null}
        </div>
      </aside>
      <main className="shuihuo-preset-library">
        <div className="shuihuo-asset-tabs">{kindOptions.map(option => <button type="button" className={libraryKind === option.value ? 'active' : ''} onClick={() => setLibraryKind(option.value)} key={option.value}>{option.label}</button>)}</div>
        <Input.Search value={search} onChange={event => setSearch(event.target.value)} placeholder={`搜索${categoryLabel(libraryKind)}图片...`} className="shuihuo-preset-search" allowClear />
        <div className="shuihuo-image-library-grid">{libraryImages.map(({ image, asset }) => <article className="shuihuo-image-card" key={image.id}><div className="shuihuo-image-card-visual"><img src={image.url} alt={`${assetName(asset)} 图片版本`} /></div><div className="shuihuo-image-card-title"><strong>{assetName(asset)}</strong><Tag>{image.isPrimary ? '主图' : `版本 ${image.revision || ''}`}</Tag></div><p>{assetPrompt(asset) || '生成时未保存提示词快照'}</p><div className="shuihuo-image-card-actions">{!image.isPrimary ? <Button type="text" size="small" loading={savingId === `primary-${image.id}`} onClick={() => makePrimary(asset, image)}>设主图</Button> : <span>当前主图</span>}<label className="batch-factory-image-upload"><input type="file" accept="image/png,image/jpeg,image/webp" onChange={event => uploadImage(asset, event)} disabled={savingId !== ''} />上传版本</label></div></article>)}{!libraryImages.length ? <div className="shuihuo-image-empty"><b>图片库</b><p>暂无已生成图片</p><span>在左侧勾选人物、场景或道具后，点击顶部“AI生成”。</span></div> : null}</div>
      </main>
    </div>
    <Modal title={editing?.id ? `编辑资产 · ${assetName(editing)}` : `添加${categoryLabel(editing?.kind || kind)}`} open={Boolean(editing)} onCancel={() => setEditing(null)} footer={<Space><Button onClick={() => setEditing(null)}>取消</Button><Button type="primary" loading={savingId !== ''} onClick={saveAsset}>保存</Button></Space>} width={920} destroyOnClose>
	  <div className="batch-factory-asset-editor-body">
	    <div className="batch-factory-asset-editor-fields">
	      <label className="shuihuo-form-label">资产分类<Select value={editing?.kind} options={kindOptions} onChange={value => setEditing(current => ({ ...current, kind: value }))} /></label>
	      <label className="shuihuo-form-label">名称<Input value={editing?.name || ''} onChange={event => setEditing(current => ({ ...current, name: event.target.value }))} /></label>
	      {editing?.kind === 'character' ? <>
	        <label className="shuihuo-form-label">正式人物名单<Input value={editing?.formalName || ''} onChange={event => setEditing(current => ({ ...current, formalName: event.target.value }))} placeholder="例如：赵姨娘（正式人物名）" /></label>
	        <label className="shuihuo-form-label">人物事实<Input.TextArea rows={3} value={editing?.facts || ''} onChange={event => setEditing(current => ({ ...current, facts: event.target.value }))} placeholder="身份、性别、年龄阶段、别名等原文事实" /></label>
	        <label className="shuihuo-form-label">人物外形<Input.TextArea rows={5} value={editing?.appearance || editing?.prompt || ''} onChange={event => setEditing(current => ({ ...current, appearance: event.target.value }))} placeholder="脸型、发型、服装、体态等跨镜头一致性锚点" /></label>
	        <label className="shuihuo-form-label">人物关系<Input.TextArea rows={3} value={editing?.relationships || ''} onChange={event => setEditing(current => ({ ...current, relationships: event.target.value }))} placeholder="与其他人物的关系及原文依据" /></label>
	      </> : null}
	      <label className="shuihuo-form-label">合并资产提示词 / AI 生图提示词<Input.TextArea rows={editing?.kind === 'character' ? 6 : 12} value={editing?.prompt || ''} onChange={event => setEditing(current => ({ ...current, prompt: event.target.value }))} placeholder="输入最终合并后的可视化提示词" /></label>
	      <p className="batch-factory-asset-editor-hint">规则：这个资产有图片时，生成视频走“图片生视频”，不发文字设定；没图片时，按上面的文字走“文字生视频”。</p>
	    </div>
	    <div className="batch-factory-asset-editor-panel">
	      {editing?.id ? (() => {
	        const imageList = assetImages[editing.id] || [];
	        const mainImage = imageList.find(item => item.isPrimary) || imageList[0] || null;
	        const panelBusy = savingId !== '' || imageGenerating;
	        return <>
	          <div className="batch-factory-asset-editor-panel-head"><b>主图预览</b></div>
	          {mainImage ? <div className="batch-factory-asset-editor-main"><Image src={mainImage.url} alt={`${assetName(editing)} 主图`} preview={{ mask: '点击放大' }} /></div>
	            : <div className="batch-factory-asset-editor-main is-empty"><PictureOutlined /><span>暂无主图，请上传图片或用 AI 生成</span></div>}
	          <div className="batch-factory-asset-editor-panel-actions">
	            <label className="batch-factory-image-upload"><input type="file" accept="image/png,image/jpeg,image/webp" onChange={event => uploadImage(editing, event)} disabled={panelBusy} />上传图片版本</label>
	            <Tooltip title={imageModelId ? '先保存当前提示词，再用引擎配置的图片模型为这个资产出新图。' : '请先在引擎配置选择图片模型。'}><Button type="primary" loading={imageGenerating} disabled={panelBusy || !imageModelId} onClick={generateEditedAssetImage}>AI生成</Button></Tooltip>
	          </div>
	          <div className="batch-factory-asset-editor-thumbs">{imageList.map(item => <button type="button" key={item.id} className={`batch-factory-asset-editor-thumb${item.id === mainImage?.id ? ' is-main' : ''}`} title={item.isPrimary ? '当前主图' : '切换为主图'} disabled={panelBusy} onClick={() => makePrimary(editing, item)}><img src={item.url} alt={`${assetName(editing)} 版本 ${item.revision || ''}`} />{item.isPrimary ? <span className="batch-factory-asset-editor-thumb-tag">主图</span> : null}</button>)}{!imageList.length ? <span className="batch-factory-asset-editor-thumbs-empty">还没有图片版本</span> : null}</div>
	        </>;
	      })() : <div className="batch-factory-asset-editor-main is-empty"><PictureOutlined /><span>先点右下角“保存”建立这个资产，<br />然后才能上传或生成图片</span></div>}
	    </div>
	  </div>
	</Modal>
  </section>;
}

function storyboardAssetNamesFromPrompt(prompt, assets) {
  const text = String(prompt || '').trim();
  if (!text) return [];
  return (assets || []).map(asset => String(asset?.name || '').trim()).filter(name => name && text.includes(name));
}

function storyboardAssetDefaults(book, videoId) {
  const videos = book?.videos || [];
  const ordinal = videos.findIndex(video => video.id === videoId);
  const draft = book?.directorRevision?.output?.storyboard?.[ordinal] || {};
  const structuredReferences = [...(draft.characters || []), ...(draft.scene ? [draft.scene] : []), ...(draft.props || [])].map(value => String(value || '').trim()).filter(Boolean);
  // Older SD cards retained a complete final_prompt but no structured asset
  // references. Match their durable asset names against that prompt so their
  // current cards remain readable without regenerating the book.
  const promptReferences = structuredReferences.length ? [] : storyboardAssetNamesFromPrompt(draft.final_prompt || draft.finalPrompt, book?.assetRecords);
  const wanted = new Set([...structuredReferences, ...promptReferences]);
  return (book?.assetRecords || []).filter(asset => wanted.has(String(asset?.name || '').trim())).map(asset => String(asset.id || '')).filter(Boolean);
}

function readStoryboardAssetSelection(book, video) {
  const defaults = storyboardAssetDefaults(book, video?.id);
  const raw = video?.settingsState?.patch?.assetSelection;
  const autoAssetIds = Array.isArray(raw?.autoAssetIds) ? raw.autoAssetIds.map(String).filter(Boolean) : defaults;
  const addedAssetIds = Array.isArray(raw?.addedAssetIds) ? raw.addedAssetIds.map(String).filter(Boolean) : [];
  const excludedAssetIds = Array.isArray(raw?.excludedAssetIds) ? raw.excludedAssetIds.map(String).filter(Boolean) : [];
  const excluded = new Set(excludedAssetIds);
  return { autoAssetIds: [...new Set(autoAssetIds)], addedAssetIds: [...new Set(addedAssetIds)], excludedAssetIds: [...new Set(excludedAssetIds)], candidateIds: [...new Set([...autoAssetIds, ...addedAssetIds, ...excludedAssetIds])], selectedIds: new Set([...autoAssetIds, ...addedAssetIds].filter(id => !excluded.has(id))) };
}

function storyboardAssetMentions(book, video) {
  const selectedIds = readStoryboardAssetSelection(book, video).selectedIds;
  return (book?.assetRecords || [])
    .filter(asset => selectedIds.has(String(asset?.id || '')))
    .map(asset => `@${assetName(asset)}`)
    .filter(Boolean)
    .join(' ');
}

function stripStandaloneAssetMentionLine(prompt, book) {
  const text = String(prompt || '');
  const names = (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length);
  if (!text || !names.length) return text;
  return text.split('\n').filter(line => {
    const match = line.match(/^\s*资产\s*[：:]\s*(.*)$/);
    if (!match) return true;
    let remainder = match[1];
    names.forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      remainder = remainder.replace(new RegExp(`@${escaped}`, 'g'), '');
    });
    return remainder.replace(/[\s、,，|]+/g, '') !== '';
  }).join('\n').replace(/\n{3,}/g, '\n\n');
}

function withStoryboardAssetMentions(prompt, book, video) {
  const text = String(prompt || '').trim();
  // 展示层只负责把正文中已有的资产名称渲染成标签；不能额外追加一行
  // “资产：@...” ，否则卡片和编辑器会与最终 VIDEO Prompt 重复显示。
  return stripStandaloneAssetMentionLine(text, book);
}

function compactUnifiedAssetDescriptions(prompt, book) {
  let text = String(prompt || '');
  const records = (book?.assetRecords || []).filter(asset => asset?.kind === 'character' || asset?.kind === 'scene' || asset?.kind === 'prop');
  if (!text || !records.length) return text;
  const start = text.search(/统一人物\s*[：:]/);
  if (start < 0) return text;
  const endCandidates = [text.indexOf('段内执行约束', start), text.indexOf('[场景', start), text.indexOf('【场景', start)].filter(index => index >= 0);
  const end = endCandidates.length ? Math.min(...endCandidates) : text.length;
  let block = text.slice(start, end);
  records
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      block = block.replace(new RegExp(`${escaped}\\s*[（(][^\\n）)]*[）)]`, 'g'), `@${name}`);
    });
  return `${text.slice(0, start)}${block}${text.slice(end)}`;
}

function separateStoryboardAssetMentions(text, book) {
  let value = String(text || '');
  const start = value.search(/基础设定\s*[：:]/);
  if (start < 0) return value;
  const endCandidates = [
    value.indexOf('视频提示词', start),
    value.indexOf('分镜视频提示词', start),
    value.indexOf('段内执行约束', start),
    value.indexOf('[场景', start),
    value.indexOf('【场景', start)
  ].filter(index => index >= 0);
  const end = endCandidates.length ? Math.min(...endCandidates) : value.length;
  let assetBlock = value.slice(start, end);
  (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      // 展示层把每个资产当作一个独立条目：同一段描述中的下一个人物、场景
      // 或道具另起一行，但不改写实际保存/提交的 VIDEO Prompt。
      assetBlock = assetBlock.replace(new RegExp(`(^|[^\\n])(@${escaped})`, 'g'), '$1\n$2');
    });
  return `${value.slice(0, start)}${assetBlock}${value.slice(end)}`;
}

function renderStoryboardPromptTextPart(part, keyPrefix) {
  return String(part || '').split('\n').flatMap((line, lineIndex) => [
    lineIndex ? <br key={`${keyPrefix}-br-${lineIndex}`} /> : null,
    <span key={`${keyPrefix}-text-${lineIndex}`}>{line}</span>
  ].filter(Boolean));
}

function renderStoryboardPromptWithAssetMentions(prompt, book, video, assetImages = {}) {
  let text = separateStoryboardAssetMentions(compactUnifiedAssetDescriptions(withStoryboardAssetMentions(prompt, book, video), book), book);
  // Director/SD cards from older runs often contain the asset name followed by
  // its full appearance description instead of an @ reference. Normalize only
  // the durable asset name for display; the description remains readable after
  // the thumbnail chip and no regeneration is needed.
  (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      text = text.replace(new RegExp(`(^|[^@])${escaped}`, 'g'), `$1@${name}`);
    });
  text = separateStoryboardAssetMentions(text, book);
  const selectedIds = readStoryboardAssetSelection(book, video).selectedIds;
  const assets = (book?.assetRecords || [])
    .filter(asset => selectedIds.has(String(asset?.id || '')) || text.includes(`@${assetName(asset)}`))
    .map(asset => ({ ...asset, mention: `@${assetName(asset)}`, imageUrl: assetImages[String(asset.id || '')] || asset?.imageUrl || asset?.image_url || asset?.thumbnailUrl || '' }))
    .filter(asset => asset.mention.length > 1)
    .sort((left, right) => right.mention.length - left.mention.length);
  if (!text || !assets.length) return text;
  const pattern = new RegExp(`(${assets.map(asset => asset.mention.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g');
  const assetByMention = new Map(assets.map(asset => [asset.mention, asset]));
  return text.split(pattern).map((part, index) => {
    const asset = assetByMention.get(part);
    return asset
      ? <span key={`${asset.id || asset.mention}-${index}`} contentEditable={false} className={`batch-factory-asset-mention is-${asset.kind || 'asset'}`} title={`${asset.kind || '资产'}：${assetName(asset)}`}><span className="batch-factory-asset-mention-avatar">{asset.imageUrl ? <img src={asset.imageUrl} alt="" /> : asset.kind === 'character' ? '人' : asset.kind === 'scene' ? '景' : '物'}</span>{assetName(asset)}</span>
      : renderStoryboardPromptTextPart(part, `prompt-text-${index}`);
  });
}

function escapePromptHTML(value) {
  return String(value ?? '').replace(/[&<>"']/g, character => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[character]));
}

function renderStoryboardPromptHTML(prompt, book, video, assetImages = {}) {
  let text = separateStoryboardAssetMentions(compactUnifiedAssetDescriptions(withStoryboardAssetMentions(prompt, book, video), book), book);
  (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      text = text.replace(new RegExp(`(^|[^@])${escaped}`, 'g'), `$1@${name}`);
    });
  text = separateStoryboardAssetMentions(text, book);
  const selectedIds = readStoryboardAssetSelection(book, video).selectedIds;
  const assets = (book?.assetRecords || [])
    .filter(asset => selectedIds.has(String(asset?.id || '')) || text.includes(`@${assetName(asset)}`))
    .map(asset => ({ ...asset, mention: `@${assetName(asset)}`, imageUrl: assetImages[String(asset.id || '')] || asset?.imageUrl || asset?.image_url || asset?.thumbnailUrl || '' }))
    .filter(asset => asset.mention.length > 1)
    .sort((left, right) => right.mention.length - left.mention.length);
  if (!text || !assets.length) return escapePromptHTML(text);
  const pattern = new RegExp(`(${assets.map(asset => asset.mention.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g');
  const assetByMention = new Map(assets.map(asset => [asset.mention, asset]));
  return text.split(pattern).map((part, index) => {
    const asset = assetByMention.get(part);
    if (!asset) return escapePromptHTML(part).replace(/\n/g, '<br>');
    const kind = asset.kind || 'asset';
    const fallback = kind === 'character' ? '人' : kind === 'scene' ? '景' : '物';
    const avatar = asset.imageUrl
      ? `<img src="${escapePromptHTML(asset.imageUrl)}" alt="" />`
      : `<span class="batch-factory-asset-mention-fallback" data-fallback="${fallback}" aria-hidden="true"></span>`;
    return `<span class="batch-factory-asset-mention is-${escapePromptHTML(kind)}" contenteditable="false" data-asset-mention="${escapePromptHTML(asset.mention)}"><span class="batch-factory-asset-mention-avatar">${avatar}</span>${escapePromptHTML(assetName(asset))}</span>`;
  }).join('');
}

function contentEditableTextOffset(root) {
  if (!root) return 0;
  const selection = typeof window !== 'undefined' ? window.getSelection() : null;
  if (!selection?.rangeCount || !root.contains(selection.anchorNode)) return contentEditablePromptText(root).length;
  let offset = 0;
  let found = false;
  const walk = node => {
    if (found) return;
    if (node === selection.anchorNode) {
      offset += Number(selection.anchorOffset || 0);
      found = true;
      return;
    }
    if (node.nodeType === Node.ELEMENT_NODE && node.dataset?.assetMention) {
      offset += node.dataset.assetMention.length;
      if (node.contains(selection.anchorNode)) found = true;
      return;
    }
    if (node.nodeType === Node.TEXT_NODE) {
      offset += String(node.nodeValue || '').length;
      return;
    }
    Array.from(node.childNodes || []).forEach(walk);
  };
  walk(root);
  return offset;
}

function contentEditablePromptText(root) {
  const walk = node => {
    if (node.nodeType === Node.TEXT_NODE) return String(node.nodeValue || '');
    if (node.nodeType === Node.ELEMENT_NODE && node.dataset?.assetMention) return String(node.dataset.assetMention);
    return Array.from(node.childNodes || []).map(walk).join('');
  };
  return walk(root).replace(/\u00a0/g, ' ');
}

function setContentEditableTextOffset(root, offset) {
  if (!root) return;
  const selection = window.getSelection();
  const range = document.createRange();
  let remaining = Math.max(0, Number(offset) || 0);
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  let node;
  while ((node = walker.nextNode())) {
    if (remaining <= node.nodeValue.length) {
      range.setStart(node, remaining);
      range.collapse(true);
      selection.removeAllRanges();
      selection.addRange(range);
      root.focus();
      return;
    }
    remaining -= node.nodeValue.length;
  }
  range.selectNodeContents(root);
  range.collapse(false);
  selection.removeAllRanges();
  selection.addRange(range);
  root.focus();
}

function InlineStoryboardAssets({ book, batchId, selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage, onSaved }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const [savingAssetId, setSavingAssetId] = useState('');
  const [starredCharacterNames, setStarredCharacterNames] = useState([]);
  const clickTimers = useRef(new Map());
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  useEffect(() => { setStarredCharacterNames(Array.isArray(book?.settingsState?.patch?.starredCharacterNames) ? book.settingsState.patch.starredCharacterNames : []); }, [book?.id, book?.revision, book?.settingsState?.revision]);
  useEffect(() => () => { for (const timer of clickTimers.current.values()) clearTimeout(timer); clickTimers.current.clear(); }, []);
  useEffect(() => {
    const h3Frame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
    const next = h3Frame?.key || (videos.some(video => video.id === selectedVideoId) ? selectedVideoId : (videos[0]?.id || ''));
    if (next !== selectedVideoId) selectVideo(next);
  }, [book?.id, videos.map(video => video.id).join('|')]);
  const selectedVideo = videos.find(video => video.id === selectedVideoId) || videos[0];
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
  const selection = readStoryboardAssetSelection(book, selectedVideo);
  const precompiledAssets = resolvePrecompiledStoryboardAssets(book);
  const showingPrecompiledAssets = !selectedVideo && precompiledAssets.length > 0;
  const assets = selectedVideo
    ? (book?.assetRecords || []).filter(asset => selection.candidateIds.includes(String(asset.id || '')))
    : precompiledAssets;
  const currentIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideo?.id));
  const switchVideo = offset => selectVideo(videos[(currentIndex + offset + videos.length) % videos.length]?.id || '');
  const switchPrecompiledFrame = offset => {
    const next = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + offset];
    if (next) selectVideo(next.key);
  };
  const assetGroups = [
    ['character', '人物'],
    ['scene', '场景'],
    ['prop', '道具']
  ].map(([kind, label]) => ({ kind, label, items: assets.filter(asset => asset.kind === kind) }));

  async function saveStar(asset, starred) {
    if (asset?.kind !== 'character') return false;
    const name = assetName(asset);
    const next = starred ? [...new Set([...starredCharacterNames, name])] : starredCharacterNames.filter(item => item !== name);
    try {
      await saveBookOverride(batchId, book.id, { patch: { starredCharacterNames: next }, expectedRevision: Number(book?.revision || 0) });
      setStarredCharacterNames(next);
      await onSaved?.();
      return true;
    } catch (error) { message.error(error?.message || '保存星标人物失败'); return false; }
  }

  async function toggleAsset(asset, { clearStar = false } = {}) {
    if (!selectedVideo || savingAssetId) return;
    const assetId = String(asset.id || '');
    if (!assetId) return;
    const isSelected = selection.selectedIds.has(assetId);
    const autoAssetIds = selection.autoAssetIds;
    const addedAssetIds = isSelected ? selection.addedAssetIds.filter(id => id !== assetId) : [...new Set([...selection.addedAssetIds, ...(autoAssetIds.includes(assetId) ? [] : [assetId])])];
    const excludedAssetIds = isSelected ? [...new Set([...selection.excludedAssetIds, assetId])] : selection.excludedAssetIds.filter(id => id !== assetId);
    setSavingAssetId(assetId);
    try {
      if (clearStar && !(await saveStar(asset, false))) return;
      await saveVideoOverride(batchId, book.id, selectedVideo.id, { patch: { assetSelection: { autoAssetIds, addedAssetIds, excludedAssetIds } }, expectedRevision: Number(selectedVideo.revision || 0) });
      await onSaved?.();
      message.success(isSelected ? `已停用“${assetName(asset)}”，不会进入当前分镜的画面和视频生成。` : `已启用“${assetName(asset)}”。`);
    } catch (error) { message.error(error?.message || '保存分镜资产选择失败'); } finally { setSavingAssetId(''); }
  }

  function onAssetClick(asset) {
    const assetId = String(asset?.id || '');
    if (!assetId) return;
    if (asset.kind === 'character' && starredCharacterNames.includes(assetName(asset))) {
      toggleAsset(asset, { clearStar: true });
      return;
    }
    clickTimers.current.set(assetId, setTimeout(() => { clickTimers.current.delete(assetId); toggleAsset(asset); }, 220));
  }

  async function onAssetDoubleClick(asset) {
    const assetId = String(asset?.id || '');
    const timer = clickTimers.current.get(assetId);
    if (timer) { clearTimeout(timer); clickTimers.current.delete(assetId); }
    if (asset.kind !== 'character') return;
    setSavingAssetId(assetId);
    try {
      const starred = !starredCharacterNames.includes(assetName(asset));
      if (await saveStar(asset, starred)) message.success(starred ? `“${assetName(asset)}”已设为星标聚焦人物。` : `已取消“${assetName(asset)}”的星标聚焦。`);
    } finally { setSavingAssetId(''); }
  }

  if (!selectedVideo && !showingPrecompiledAssets) return <div className="batch-factory-cell-shell batch-factory-inline-assets is-empty"><div className="batch-factory-cell-empty">AI 推理后会在这里显示当前书的分镜资产。</div></div>;

  const selectedCount = showingPrecompiledAssets ? assets.length : assets.filter(asset => selection.selectedIds.has(String(asset.id))).length;
  return <div className="batch-factory-cell-shell batch-factory-inline-assets">
    <header className="batch-factory-cell-head">
      <div><b>{showingPrecompiledAssets ? `导演分镜卡 ${(selectedPrecompiledFrame?.index || 0) + 1} · 资产` : storyboardVideoLabel(selectedVideo, currentIndex)}</b><small>{showingPrecompiledAssets ? '导演分镜卡已提取，等待 VIDEO 编译' : `${selectedCount}/${assets.length || 0} 已启用`}</small></div>
      <span className="batch-factory-cell-state">{assets.length ? '资产就绪' : '待提取'}</span>
    </header>
    <div className="batch-factory-cell-content batch-factory-asset-content">
      {assets.length ? assetGroups.map(group => group.items.length ? <section className="batch-factory-asset-group-inline" key={group.kind}>
        <div className="batch-factory-asset-group-label"><b>{group.label}</b><span>{group.items.length}</span></div>
        <div className="batch-factory-storyboard-asset-cards">{group.items.map(asset => {
          const selected = showingPrecompiledAssets || selection.selectedIds.has(String(asset.id));
          const starred = !showingPrecompiledAssets && selected && asset.kind === 'character' && starredCharacterNames.includes(assetName(asset));
          const onClick = showingPrecompiledAssets ? () => onManage?.() : () => onAssetClick(asset);
          const onDoubleClick = showingPrecompiledAssets ? undefined : () => onAssetDoubleClick(asset);
          const title = showingPrecompiledAssets ? '导演分镜卡已提取；最终 VIDEO 编译后可按分镜调整资产。点击管理全部资产。' : starred ? '星标人物聚焦；点击后取消星标并停用，双击可取消星标。' : selected ? '点击停用；双击人物卡设为星标聚焦。' : '点击启用：用于当前分镜的图片和视频生成';
          return <button type="button" key={asset.id} disabled={savingAssetId === asset.id} onClick={onClick} onDoubleClick={onDoubleClick} className={`batch-factory-storyboard-asset-card ${starred ? 'is-starred' : selected ? 'is-selected' : 'is-muted'}`} title={title}>
            <span>{group.label}</span><b>{assetName(asset)}</b>{starred ? <i aria-label="星标人物">★</i> : null}
          </button>;
        })}</div>
      </section> : null) : <div className="batch-factory-cell-empty">当前分镜尚未识别到需要调用的资产。</div>}
      <div className="batch-factory-asset-summary">
        <span>{assetGroups[0].items.length} 人物</span><span>{assetGroups[1].items.length} 场景</span><span>{assetGroups[2].items.length} 道具</span>
        <button type="button" onClick={() => onManage?.(selectedVideo?.id || selectedPrecompiledFrame?.key)}>管理全部资产</button>
      </div>
    </div>
    {showingPrecompiledAssets ? <div className="batch-factory-cell-pager"><button type="button" aria-label="上一张导演分镜卡资产" disabled={!selectedPrecompiledFrame || selectedPrecompiledFrame.index === 0} onClick={() => switchPrecompiledFrame(-1)}><LeftOutlined /></button><span>{(selectedPrecompiledFrame?.index || 0) + 1} / {precompiledWorkspace.frames.length}</span><button type="button" aria-label="下一张导演分镜卡资产" disabled={!selectedPrecompiledFrame || selectedPrecompiledFrame.index >= precompiledWorkspace.frames.length - 1} onClick={() => switchPrecompiledFrame(1)}><RightOutlined /></button></div> : <div className="batch-factory-cell-pager">
      <button type="button" aria-label="上一分镜资产" disabled={currentIndex === 0} onClick={() => switchVideo(-1)}><LeftOutlined /></button>
      <span>{currentIndex + 1} / {videos.length}</span>
      <button type="button" aria-label="下一分镜资产" disabled={currentIndex >= videos.length - 1} onClick={() => switchVideo(1)}><RightOutlined /></button>
    </div>}
  </div>;
}

function useCompiledVideoPrompt(batchId, book, video, settingsRevision = 0, { enabled = true } = {}) {
  const [displayPrompt, setDisplayPrompt] = useState('');
  const [compiledPrompt, setCompiledPrompt] = useState('');
  const [baseSetupPrompt, setBaseSetupPrompt] = useState('');
  const [smartUnifiedPending, setSmartUnifiedPending] = useState(false);
  const [loading, setLoading] = useState(false);
  const bookRevision = Number(book?.revision || 0);
  const videoRevision = Number(video?.revision || 0);
  useEffect(() => {
    let active = true;
    if (!enabled || !batchId || !book?.id || !video?.id) {
      setDisplayPrompt('');
      setCompiledPrompt('');
      setBaseSetupPrompt('');
      setSmartUnifiedPending(false);
      setLoading(false);
      return () => { active = false; };
    }
    setDisplayPrompt('');
    setCompiledPrompt('');
    setBaseSetupPrompt('');
    setSmartUnifiedPending(false);
    setLoading(true);
    getFinalPrompt(batchId, book.id, video.id, { silent: true, suppressGlobalError: true }).then(response => {
      if (!active) return;
	  setDisplayPrompt(String(response?.finalPrompt?.displayPrompt || ''));
	  setCompiledPrompt(String(response?.finalPrompt?.compiledPrompt || ''));
      setBaseSetupPrompt(String(response?.finalPrompt?.baseSetupPrompt || ''));
      setSmartUnifiedPending(response?.finalPrompt?.smartUnifiedPending === true);
    }).catch(() => {
      // The raw director output remains readable when the compiler capability
      // is temporarily unavailable. API errors are intentionally quiet here:
      // this read runs once for every visible book row.
      if (active) { setDisplayPrompt(''); setCompiledPrompt(''); setBaseSetupPrompt(''); }
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [enabled, batchId, book?.id, bookRevision, video?.id, videoRevision, settingsRevision]);
  return { displayPrompt, compiledPrompt, baseSetupPrompt, smartUnifiedPending, loading };
}

function InlineBookPrompts({ batch, book, batchId, settingsRevision, selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage }) {
  const videos = book?.videos || [];
	const h3Cards = h3DirectorCards(book);
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const [assetThumbs, setAssetThumbs] = useState({});
  const smartUnifiedAnalysis = smartUnifiedAnalysisForBook(book);
  const showSmartUnified = smartUnifiedDisplayEnabled(batch, book) && smartUnifiedAnalysis;
  const h3CompilationMessage = h3CompilationStatus(effectiveBookSettings(batch, book));
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const [promptKind, setPromptKind] = useState('video');
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const video = videos[selectedIndex] || null;
  useEffect(() => {
    let active = true;
    const records = (book?.assetRecords || []).filter(asset => asset?.id);
    if (!batchId || !book?.id || !records.length) { setAssetThumbs({}); return () => { active = false; }; }
    Promise.all(records.map(async asset => {
      try {
        const response = await listBookAssetImages(batchId, book.id, asset.id);
        const images = Array.isArray(resultData(response, 'images')) ? resultData(response, 'images') : [];
        const primary = images.find(image => image?.isPrimary) || images[0];
        return [String(asset.id), String(primary?.url || primary?.imageUrl || '')];
      } catch (_) { return [String(asset.id), '']; }
    })).then(entries => {
      if (active) setAssetThumbs(Object.fromEntries(entries));
    });
    return () => { active = false; };
  }, [batchId, book?.id, book?.revision, book?.assetRecords?.length]);
  const rawVideoPrompt = String(video?.videoPrompt || '');
  const visualPrompt = String(video?.visualPrompt || '');
  const hasVisualPrompt = Boolean(visualPrompt.trim());
  // 分镜卡直接读取 Go 编译后的最终 Prompt，让已开启的智能统一、基础设定、
  // 画质、限制和负面规则在卡片中可见；弹窗与卡片使用同一份编译结果。
  const currentSettings = effectiveBookSettings(batch, book);
  const baseSetupEnabledForBook = baseSetupEnabled(currentSettings);
  const { displayPrompt, compiledPrompt, baseSetupPrompt, smartUnifiedPending, loading } = useCompiledVideoPrompt(batchId, book, video, settingsRevision, { enabled: true });
  // 智能统一是增强层。视觉基线尚未取得时，仍须展示已经编译的 VIDEO
  // 提示词，并明确标出基线状态，不能把一条可执行分镜伪装成“什么都没有”。
  const rawCompiledPrompt = compiledPrompt || displayPrompt || rawVideoPrompt;
  const videoPrompt = withStoryboardAssetMentions(baseSetupEnabledForBook ? rawCompiledPrompt : stripBaseSetupSection(rawCompiledPrompt), book, video);
  const assetPromptText = baseSetupEnabledForBook ? String(baseSetupPrompt || '').trim() : '';
  const unifiedStyleText = showSmartUnified ? String(smartUnifiedAnalysis.prompt || '').trim() : '';
  const videoPromptWithUnifiedStyle = prependSmartUnifiedStyle(videoPrompt, unifiedStyleText);
  const displayVideoPrompt = baseSetupEnabledForBook
    ? composeStoryboardDisplayPrompt(videoPromptWithUnifiedStyle, assetPromptText || storyboardAssetPromptText(book))
    : videoPromptWithUnifiedStyle;
  const activePrompt = promptKind === 'visual' ? visualPrompt : displayVideoPrompt;

  useEffect(() => {
    const first = book?.videos?.[0]?.id || '';
    if (!controlledSelectedVideoId) setLocalSelectedVideoId(first);
    setPromptKind('video');
  }, [book?.id]);

  useEffect(() => {
    if (!videos.some(item => item.id === selectedVideoId) && videos[0]?.id) selectVideo(videos[0].id);
  }, [videos.map(item => item.id).join('|')]);

  function move(direction) {
    const next = videos[selectedIndex + direction];
    if (next) selectVideo(next.id);
  }

  if (!video && h3Cards.length) {
		const h3Frame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
		const h3Card = h3Frame?.card || {};
		const h3CardIndex = h3Frame?.index || 0;
		const h3ConstraintRows = h3ConstraintPreview(effectiveBookSettings(batch, book));
		const h3ConstraintLabels = h3ConstraintRows.map(row => row.split('：', 1)[0]).join(' · ');
    const h3AssetPromptText = baseSetupEnabled(currentSettings) ? storyboardAssetPromptText(book) : '';
		const h3VideoPromptLabel = selectedVideoPromptLabel(effectiveBookSettings(batch, book));
		const selectH3Frame = index => {
			const next = precompiledWorkspace.frames?.[index];
			if (next) selectVideo(next.key);
		};
		const camera = h3Card.camera || {};
		return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell">
			<div className="batch-factory-cell-shell batch-factory-prompt-entry-card is-h3-director-card">
				<header className="batch-factory-cell-head"><div><b>导演分镜卡 {h3CardIndex + 1}</b><small>视频提示词：{h3VideoPromptLabel} · 已提取，尚未编译 VIDEO</small></div></header>
				<button type="button" className="batch-factory-cell-content batch-factory-prompt-entry-content" aria-label="打开分镜提示词编辑" onClick={() => onManage?.(h3Frame?.key || '')}>
					<span className="batch-factory-prompt-status">导演分镜卡已提取</span>
					<p>{String(h3Card.source_text || '')}{h3Card.action ? `\n\n动作：${h3Card.action}` : ''}{camera.shot_size ? `\n机位：${camera.shot_size}${camera.shot_angle ? ` · ${camera.shot_angle}` : ''}` : ''}{h3AssetPromptText ? `\n\n【资产提示词】\n${h3AssetPromptText}` : ''}{h3ConstraintRows.length ? `\n\n【当前已启用约束】\n${h3ConstraintRows.join('\n')}` : ''}</p>
					<small title={[h3AssetPromptText, ...h3ConstraintRows].filter(Boolean).join('\n')}>{h3AssetPromptText ? `资产提示词：${promptPreview(h3AssetPromptText)}` : '暂无资产提示词'} · {h3ConstraintLabels ? `约束：${h3ConstraintLabels}` : '未启用约束'} · {h3CompilationMessage}</small>
				</button>
				<div className="batch-factory-cell-pager">
					<button type="button" aria-label="上一张导演分镜卡" disabled={h3CardIndex === 0} onClick={() => selectH3Frame(h3CardIndex - 1)}><LeftOutlined /></button>
					<span>{h3CardIndex + 1} / {h3Cards.length}</span>
					<button type="button" aria-label="下一张导演分镜卡" disabled={h3CardIndex >= h3Cards.length - 1} onClick={() => selectH3Frame(h3CardIndex + 1)}><RightOutlined /></button>
				</div>
			</div>
		</div>;
	}

  if (!video) return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell"><div className="batch-factory-cell-shell"><div className="batch-factory-cell-empty">AI 推理后会在这里显示当前书的分镜提示词。</div></div></div>;

  const stateLabel = promptKind === 'visual' ? (hasVisualPrompt ? '画面已生成' : '待生成') : loading ? '编译中' : videoPrompt.trim() ? (smartUnifiedPending ? '基线未获取' : '最终已编译') : '待生成';
  const videoPromptLabel = selectedVideoPromptLabel(effectiveBookSettings(batch, book));

  return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell">
    <div className="batch-factory-cell-shell batch-factory-prompt-entry-card">
      <header className="batch-factory-cell-head">
        <div><b>{storyboardVideoLabel(video, selectedIndex)}</b><small>{promptKind === 'visual' ? '画面提示词' : '最终视频提示词'}</small></div>
        <div className="batch-factory-cell-segmented" aria-label="提示词类型">
          <button type="button" className={promptKind === 'video' ? 'is-active' : ''} onClick={() => setPromptKind('video')}>视频</button>
          <button type="button" className={promptKind === 'visual' ? 'is-active' : ''} disabled={!hasVisualPrompt} onClick={() => setPromptKind('visual')}>画面</button>
        </div>
      </header>
      <button type="button" className="batch-factory-cell-content batch-factory-prompt-entry-content" aria-label="打开分镜提示词编辑" onClick={() => onManage?.(video.id)}>
        <span className="batch-factory-prompt-status">{stateLabel}</span>
        <span className="batch-factory-prompt-entry-rich-view">
          {activePrompt ? renderStoryboardPromptWithAssetMentions(activePrompt, book, video, assetThumbs) : (smartUnifiedPending && promptKind === 'video' ? '智能统一视觉基线尚未取得；不影响本分镜的提示词使用' : loading && promptKind === 'video' ? '正在编译最终视频提示词…' : promptKind === 'visual' ? '请生成画面提示词再查看' : '请生成视频提示词再查看')}
        </span>
      </button>
      <div className="batch-factory-cell-pager">
        <button type="button" aria-label="上一分镜" disabled={selectedIndex === 0} onClick={() => move(-1)}><LeftOutlined /></button>
        <span>{selectedIndex + 1} / {videos.length}</span>
        <button type="button" aria-label="下一分镜" disabled={selectedIndex >= videos.length - 1} onClick={() => move(1)}><RightOutlined /></button>
      </div>
    </div>
  </div>;
}

function StoryboardVideoNavigator({ videos, selectedVideoId, onSelect }) {
	const index = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
	const hasVideos = videos.length > 0;
	return <Space wrap className="batch-factory-storyboard-navigator"><Tooltip title="上一分镜"><Button aria-label="上一分镜" icon={<LeftOutlined />} disabled={!hasVideos || index === 0} onClick={() => onSelect(videos[index - 1]?.id)} /></Tooltip><Select value={selectedVideoId || undefined} onChange={onSelect} style={{ minWidth: 260 }} placeholder="选择分镜 / VIDEO" options={videos.map((video, videoIndex) => ({ value: video.id, label: storyboardVideoLabel(video, videoIndex) }))} /><Tooltip title="下一分镜"><Button aria-label="下一分镜" icon={<RightOutlined />} disabled={!hasVideos || index >= videos.length - 1} onClick={() => onSelect(videos[index + 1]?.id)} /></Tooltip></Space>;
}

function PromptPanel({ book, batch, batchId, settingsRevision, initialVideoId = '', onSaved, onCompile, onRegenerate, onRegenerateVisual, onRetry, onGenerateVideo, onGenerateVisual, regenerating, productionAvailable = false, productionReason = '' }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const [selectedVideoId, setSelectedVideoId] = useState(initialVideoId || videos[0]?.id || '');
  const [promptKind, setPromptKind] = useState('video');
  const [editing, setEditing] = useState(false);
  const [editingSnapshot, setEditingSnapshot] = useState(null);
  const [saving, setSaving] = useState(false);
  const [videoPrompt, setVideoPrompt] = useState('');
  const [visualPrompt, setVisualPrompt] = useState('');
  const [assetMentionMenu, setAssetMentionMenu] = useState(null);
  const promptInputRef = useRef(null);
  const promptEditableRef = useRef(null);
  const [h3TraceOpen, setH3TraceOpen] = useState(false);
  const [openingVariantsOpen, setOpeningVariantsOpen] = useState(false);
  const [h3TraceBusy, setH3TraceBusy] = useState(false);
  const [h3Trace, setH3Trace] = useState(null);
  const [h3TraceError, setH3TraceError] = useState('');
  const [promptEditTrace, setPromptEditTrace] = useState(null);
  const [assetThumbs, setAssetThumbs] = useState({});
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const selectedVideo = videos[selectedIndex] || null;
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId || initialVideoId);
  const h3Document = h3DirectorDocument(book);
  const engineSettings = effectiveBookSettings(batch, book);
  const { displayPrompt, compiledPrompt, smartUnifiedPending, loading: compilingPrompt } = useCompiledVideoPrompt(batchId, book, selectedVideo, settingsRevision);
  const smartUnifiedAnalysis = smartUnifiedAnalysisForBook(book);
  const showSmartUnified = smartUnifiedDisplayEnabled(batch, book) && smartUnifiedAnalysis;
  const hasVisualPrompt = Boolean(String(visualPrompt || '').trim());
  const hasVideoPrompt = Boolean(String(videoPrompt || '').trim());
  useEffect(() => {
    const first = book?.videos?.find(video => video.id === initialVideoId) || book?.videos?.[0] || null;
    const firstH3Frame = resolvePrecompiledStoryboardFrame(book, initialVideoId);
    setSelectedVideoId(first?.id || firstH3Frame?.key || '');
    setVideoPrompt(first?.videoPrompt || '');
    setVisualPrompt(first?.visualPrompt || '');
    setPromptKind('video');
    setEditing(false);
  }, [book?.id, initialVideoId]);
  useEffect(() => {
    setVideoPrompt(selectedVideo?.videoPrompt || '');
    setVisualPrompt(selectedVideo?.visualPrompt || '');
    setEditing(false);
    setAssetMentionMenu(null);
  }, [selectedVideoId, selectedVideo?.revision]);
  useEffect(() => {
    let active = true;
    const records = (book?.assetRecords || []).filter(asset => asset?.id);
    if (!batchId || !book?.id || !records.length) { setAssetThumbs({}); return () => { active = false; }; }
    Promise.all(records.map(async asset => {
      try {
        const response = await listBookAssetImages(batchId, book.id, asset.id);
        const images = resultData(response, 'images');
        const primary = Array.isArray(images) ? (images.find(image => image?.isPrimary) || images[0]) : null;
        return [String(asset.id), String(primary?.url || primary?.imageUrl || primary?.image_url || primary?.thumbnailUrl || primary?.path || '')];
      } catch (_) { return [String(asset.id), '']; }
    })).then(entries => { if (active) setAssetThumbs(Object.fromEntries(entries)); });
    return () => { active = false; };
  }, [batchId, book?.id, book?.revision, book?.assetRecords?.length]);
  useEffect(() => {
    if (!editing || promptKind !== 'video' || !promptEditableRef.current) return;
    promptEditableRef.current.innerHTML = renderStoryboardPromptHTML(videoPrompt, book, selectedVideo, assetThumbs);
  }, [editing, promptKind, selectedVideoId, selectedVideo?.revision, promptEditTrace, assetThumbs]);
  async function savePrompt() {
    if (!selectedVideo) return;
    setSaving(true);
    try {
      const patch = promptKind === 'visual' ? { visualPrompt } : { videoPrompt };
      if (promptKind === 'video' && (book?.directorRevision?.output?.h3_director || book?.directorRevision?.output?.h3Director)) {
        if (!promptEditTrace) throw new Error('编辑版本缺失，请重新打开编辑');
        await compileH3Video(batchId, book.id, h3PromptEditRequest(promptEditTrace, selectedIndex, videoPrompt));
      } else {
        await saveVideoOverride(batchId, book.id, selectedVideo.id, { patch, expectedRevision: Number(selectedVideo.revision || 0) });
      }
      await onSaved?.();
      setEditing(false);
      setEditingSnapshot(null);
      message.success(promptKind === 'visual' ? '当前分镜画面提示词已保存。' : '当前分镜视频提示词已保存。');
    } catch (error) { message.error(error?.message || '保存分镜提示词失败'); } finally { setSaving(false); }
  }
  async function openH3Trace() {
    setH3TraceOpen(true);
    setH3TraceBusy(true);
    setH3TraceError('');
    try {
      setH3Trace(await getH3Trace(batchId, book.id));
    } catch (error) {
      setH3Trace(null);
      setH3TraceError(error?.message || '读取 H3 Trace 失败');
    } finally {
      setH3TraceBusy(false);
    }
  }
  function move(direction) {
    const next = videos[selectedIndex + direction];
    if (next) setSelectedVideoId(next.id);
  }

  function movePrecompiledFrame(direction) {
    const frame = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + direction];
    if (frame) setSelectedVideoId(frame.key);
  }

  function updatePromptValue(event) {
    const nextValue = event.target.value;
    if (promptKind === 'visual') { setVisualPrompt(nextValue); return; }
    setVideoPrompt(nextValue);
    const cursor = Number(event.target.selectionStart || nextValue.length);
    const match = nextValue.slice(0, cursor).match(/@([^\s@]*)$/);
    if (!match) { setAssetMentionMenu(null); return; }
    const query = match[1].toLowerCase();
    const candidates = (book?.assetRecords || []).filter(asset => asset?.id && String(assetName(asset) || '').toLowerCase().includes(query));
    setAssetMentionMenu({ start: cursor - match[0].length, end: cursor, candidates });
  }

  function updateEditablePromptValue(event) {
    const root = event.currentTarget;
    const nextValue = contentEditablePromptText(root);
    setVideoPrompt(nextValue);
    updateAssetMentionMenu(root, nextValue);
  }

  function updateAssetMentionMenu(root, value = String(root?.innerText || '').replace(/\u00a0/g, ' ')) {
    let cursor = contentEditableTextOffset(root);
    let match = value.slice(0, cursor).match(/@([^\s@]*)$/);
    // Some browsers keep the old selection when a contenteditable is
    // rehydrated with rich mention spans.  If the actual input ends in an @
    // query, prefer that real text position instead of silently closing the
    // picker because the stale selection points into the previous span.
    const trailingMatch = value.match(/@([^\s@]*)$/);
    if (!match && trailingMatch) {
      cursor = value.length;
      match = trailingMatch;
    }
    if (!match) { setAssetMentionMenu(null); return; }
    const query = match[1].toLowerCase();
    const candidates = (book?.assetRecords || []).filter(asset => asset?.id && String(assetName(asset) || '').toLowerCase().includes(query));
    const selection = typeof window !== 'undefined' ? window.getSelection() : null;
    const range = selection?.rangeCount ? selection.getRangeAt(0) : null;
    const parent = root?.parentElement;
    const caretRect = range?.getBoundingClientRect?.();
    const parentRect = parent?.getBoundingClientRect?.();
    const left = caretRect && parentRect ? Math.max(8, caretRect.left - parentRect.left) : 8;
    const top = caretRect && parentRect ? Math.max(8, caretRect.bottom - parentRect.top + 6) : 8;
    setAssetMentionMenu({ start: cursor - match[0].length, end: cursor, candidates, left, top });
  }

  function handleEditablePromptKeyUp(event) {
    if (event.key === '@' || event.key.length === 1) updateAssetMentionMenu(event.currentTarget);
  }

  function handleEditablePromptKeyDown(event) {
    if (event.key === '@') requestAnimationFrame(() => updateAssetMentionMenu(event.currentTarget));
  }

  function cancelEditing() {
    if (editingSnapshot) {
      setVideoPrompt(editingSnapshot.videoPrompt);
      setVisualPrompt(editingSnapshot.visualPrompt);
    }
    setEditingSnapshot(null);
    setAssetMentionMenu(null);
    setEditing(false);
  }

  async function toggleEditing() {
    if (editing) {
      setEditing(false);
      setAssetMentionMenu(null);
      return;
    }
    setEditingSnapshot({ videoPrompt, visualPrompt });
    if (promptKind === 'video' && (book?.directorRevision?.output?.h3_director || book?.directorRevision?.output?.h3Director)) {
      try {
        const trace = await getH3Trace(batchId, book.id);
        const segment = trace?.compilation?.compilation?.segments?.[selectedIndex];
        if (!segment) throw new Error('分镜编译记录缺失');
        setPromptEditTrace(trace);
        setVideoPrompt(segment.compiled_prompt);
      } catch (error) { setEditingSnapshot(null); message.error(error?.message || '读取分镜失败'); return; }
    } else if (promptKind === 'video') setVideoPrompt(submittedVideoPrompt);
    setEditing(true);
  }

  function insertAssetMention(asset) {
    const menu = assetMentionMenu;
    if (!menu || !asset) return;
    const value = videoPrompt;
    const mention = `@${assetName(asset)} `;
    const next = `${value.slice(0, menu.start)}${mention}${value.slice(menu.end)}`;
    setVideoPrompt(next);
    setAssetMentionMenu(null);
    requestAnimationFrame(() => {
      if (promptEditableRef.current) {
        promptEditableRef.current.innerHTML = renderStoryboardPromptHTML(next, book, selectedVideo, assetThumbs);
        setContentEditableTextOffset(promptEditableRef.current, menu.start + mention.length);
        return;
      }
      const input = promptInputRef.current?.resizableTextArea?.textArea || promptInputRef.current?.textArea;
      if (input) { input.focus(); const position = menu.start + mention.length; input.setSelectionRange(position, position); }
    });
  }
	  const activeLabel = promptKind === 'visual' ? '画面提示词' : '分镜视频提示词';
  const rawSubmittedVideoPrompt = compiledPrompt || displayPrompt || videoPrompt;
  const rawPromptWithUnifiedStyle = showSmartUnified
    ? prependSmartUnifiedStyle(rawSubmittedVideoPrompt, smartUnifiedAnalysis.prompt)
    : rawSubmittedVideoPrompt;
  const submittedVideoPrompt = baseSetupEnabled(engineSettings) ? rawPromptWithUnifiedStyle : stripBaseSetupSection(rawPromptWithUnifiedStyle);
  const displaySubmittedVideoPrompt = baseSetupEnabled(engineSettings)
    ? composeStoryboardDisplayPrompt(submittedVideoPrompt, storyboardAssetPromptText(book))
    : submittedVideoPrompt;
  const activeValue = promptKind === 'visual' ? visualPrompt : editing ? videoPrompt : displaySubmittedVideoPrompt;
  const activeReady = promptKind === 'visual' ? hasVisualPrompt : Boolean(displaySubmittedVideoPrompt.trim());
  const regenerateCurrentPrompt = promptKind === 'visual' ? onRegenerateVisual : onRegenerate;
  const h3Segments = h3Trace?.compilation?.compilation?.segments || [];
  const h3Segment = h3Segments[selectedIndex] || null;
	  const h3ConstraintRows = h3ConstraintPreview(engineSettings);
  const h3AssetPromptText = baseSetupEnabled(engineSettings) ? storyboardAssetPromptText(book) : '';
  const apiDurationSeconds = Number(selectedVideo?.durationSeconds || 0);
  const realDurationSeconds = realVideoDurationSeconds(submittedVideoPrompt);
  const durationLabel = realDurationSeconds !== null
    ? `${realDurationSeconds.toFixed(3)}s / ${apiDurationSeconds}s`
    : apiDurationSeconds ? `${apiDurationSeconds}s` : '—';
  const textModelLabel = String(engineSettings.textModelId || '—');
  const videoModelLabel = String(engineSettings.videoModelId || '—');
  if (!selectedVideo && selectedPrecompiledFrame) {
    const card = selectedPrecompiledFrame.card || {};
    return <div className="batch-factory-prompt-modal-stack">
      <div className="batch-factory-prompt-modal-head"><Space><Tooltip title="上一张导演分镜卡"><Button aria-label="上一张导演分镜卡" icon={<LeftOutlined />} disabled={selectedPrecompiledFrame.index === 0} onClick={() => movePrecompiledFrame(-1)} /></Tooltip><span className="batch-factory-prompt-modal-index">{selectedPrecompiledFrame.index + 1}/{precompiledWorkspace.frames.length}</span><Tooltip title="下一张导演分镜卡"><Button aria-label="下一张导演分镜卡" icon={<RightOutlined />} disabled={selectedPrecompiledFrame.index >= precompiledWorkspace.frames.length - 1} onClick={() => movePrecompiledFrame(1)} /></Tooltip></Space><span className="batch-factory-prompt-status">导演分镜卡 · 待编译 VIDEO</span></div>
      <Alert type="info" showIcon message="当前是导演分镜结构，不是最终 VIDEO 提示词" description="提示词预设决定内容和格式；视频模型只负责最终生成。开启跟随配音后会先读取每行真实时长，再按 10/15 秒确定性切段并填充最终 VIDEO Prompt。" />
			<label className="shuihuo-form-label batch-factory-prompt-editor-label">分镜视频提示词<Input.TextArea rows={18} readOnly value={`${formatH3DirectorCardPrompt(h3Document, card)}${h3AssetPromptText ? `\n\n【资产提示词】\n${h3AssetPromptText}` : ''}${h3ConstraintRows.length ? `\n\n【当前已启用约束】\n${h3ConstraintRows.join('\n')}` : ''}`} /></label>
      <div className="batch-factory-prompt-modal-actions">
        <Tooltip title="等待最终 VIDEO 编译后才能保存"><Button type="primary" disabled>保存</Button></Tooltip>
        <Tooltip title="复用当前导演分镜卡、真实配音时长和当前约束设置，编译最终 VIDEO 提示词；不会重新调用导演模型。"><Button type="primary" loading={regenerating} disabled={regenerating || !onCompile} onClick={onCompile}>编译最终提示词</Button></Tooltip>
        <Button loading={regenerating} disabled={regenerating || !onRegenerate} onClick={onRegenerate}>重新生成导演分镜</Button>
        <Tooltip title="等待最终 VIDEO 编译后才能编辑"><Button disabled>编辑</Button></Tooltip>
        <Tooltip title="等待最终 VIDEO 编译"><Button disabled>生成视频</Button></Tooltip>
        <Button onClick={openH3Trace}>查看编译 Trace</Button>
        <Button type="text" disabled={regenerating} onClick={() => onRetry?.()}>重试</Button>
      </div>
      <p className="shuihuo-modal-note">此处与当前行的资产、分镜视频共用同一张导演分镜卡；最终 VIDEO 生成前，不会伪造可播放视频。</p>
      <Modal title="分镜编译与提交 Trace" open={h3TraceOpen} onCancel={() => setH3TraceOpen(false)} footer={null} width={980}>{h3TraceBusy ? <Alert type="info" showIcon message="正在读取编译 Trace…" /> : h3TraceError ? <Alert type="error" showIcon message="编译 Trace 读取失败" description={h3TraceError} /> : <Alert type="info" showIcon message="当前导演卡尚未产生最终 VIDEO 编译记录" />}</Modal>
    </div>;
  }
  return <div className="batch-factory-prompt-modal-stack">
    <div className="batch-factory-prompt-modal-head">
      <Space>
        <Tooltip title="上一分镜"><Button aria-label="上一分镜" icon={<LeftOutlined />} disabled={selectedIndex === 0} onClick={() => move(-1)} /></Tooltip>
        <span className="batch-factory-prompt-modal-index">{videos.length ? `${selectedIndex + 1}/${videos.length}` : '0/0'}</span>
        <Tooltip title="下一分镜"><Button aria-label="下一分镜" icon={<RightOutlined />} disabled={!videos.length || selectedIndex >= videos.length - 1} onClick={() => move(1)} /></Tooltip>
      </Space>
      <div className="batch-factory-prompt-modal-tabs" aria-label="提示词类型">
        <button type="button" className={promptKind === 'visual' ? 'is-visual active' : 'is-visual'} disabled={!hasVisualPrompt} onClick={() => setPromptKind('visual')}>画面提示词</button>
        <button type="button" className={promptKind === 'video' ? 'is-video active' : 'is-video'} disabled={!hasVideoPrompt} onClick={() => setPromptKind('video')}>视频提示词</button>
      </div>
    </div>
    <div className="batch-factory-prompt-meta-bar">
      <span className="batch-factory-prompt-meta-item"><b>真实/API时长：</b>{durationLabel}</span>
      <span className="batch-factory-prompt-meta-item"><b>文本模型：</b>{textModelLabel}</span>
      <span className="batch-factory-prompt-meta-item"><b>视频模型：</b>{videoModelLabel}</span>
    </div>
    {!activeReady ? <Alert type="info" showIcon message={smartUnifiedPending && promptKind === 'video' ? '智能统一待分析' : promptKind === 'visual' ? '请生成画面提示词再查看' : '请生成视频提示词再查看'} description={smartUnifiedPending && promptKind === 'video' ? '智能统一与资产提示词分开获取；未取得基线不会阻断资产、分镜或视频。可在资产区单独刷新。' : '当前分镜尚未有这一类提示词；切换分镜时会保持当前查看类型。'} /> : <>
      <label className="shuihuo-form-label batch-factory-prompt-editor-label">
        {activeLabel}
        {!editing && promptKind === 'video' ? <div className="batch-factory-prompt-rich-view">{renderStoryboardPromptWithAssetMentions(activeValue, book, selectedVideo, assetThumbs)}</div> : editing && promptKind === 'video' ? <div className="batch-factory-prompt-editor-rich-wrap">
          <div ref={promptEditableRef} className="batch-factory-prompt-rich-view is-editing" contentEditable suppressContentEditableWarning role="textbox" aria-label="可编辑分镜视频提示词" onInput={updateEditablePromptValue} onKeyDown={handleEditablePromptKeyDown} onKeyUp={handleEditablePromptKeyUp} onBlur={() => setTimeout(() => setAssetMentionMenu(null), 120)} />
          {assetMentionMenu ? <div className="batch-factory-asset-mention-menu" style={{ left: assetMentionMenu.left, top: assetMentionMenu.top }}>{assetMentionMenu.candidates?.length ? assetMentionMenu.candidates.map(asset => { const imageUrl = assetThumbs[String(asset.id)] || asset.imageUrl || asset.image_url || asset.thumbnailUrl; const kindLabel = asset.kind === 'character' ? '角色' : asset.kind === 'scene' ? '场景' : '道具'; return <button type="button" key={asset.id} onMouseDown={event => event.preventDefault()} onClick={() => insertAssetMention(asset)}>{imageUrl ? <img src={imageUrl} alt="" /> : <span className="batch-factory-asset-mention-fallback" aria-hidden="true">{asset.kind === 'character' ? '人' : asset.kind === 'scene' ? '景' : '物'}</span>}<span>{assetName(asset)}</span><small>{kindLabel}</small></button>; }) : <div className="batch-factory-asset-mention-empty">暂无可用资产，请先在资产设置添加人物、场景或道具</div>}</div> : null}
        </div> : <div className="batch-factory-prompt-editor-wrap"><Input.TextArea ref={promptInputRef} rows={14} value={activeValue} onChange={event => setVisualPrompt(event.target.value)} placeholder="仅用于生成当前分镜的画面图片" /></div>}
      </label>
      <div className="batch-factory-prompt-modal-actions">
        <Button type="primary" loading={saving} disabled={!editing || !selectedVideo} onClick={savePrompt}>保存</Button>
        <Tooltip title={promptKind === 'visual' ? '按当前小说原文重新生成画面提示词。' : '按当前小说原文重新生成导演分镜，并重编最终 VIDEO 提示词；不会提交视频生成任务。'}><Button icon={promptKind === 'visual' ? <ReloadOutlined /> : null} loading={regenerating} disabled={regenerating || !regenerateCurrentPrompt} onClick={regenerateCurrentPrompt}>{promptKind === 'visual' ? '重新生成画面提示词' : '重新生成导演分镜'}</Button></Tooltip>
        {editing ? <Button onClick={cancelEditing}>取消编辑</Button> : null}
        <Button onClick={toggleEditing}>{editing ? '完成编辑' : '编辑'}</Button>
        {promptKind === 'visual' ? <Button onClick={() => onGenerateVisual?.(selectedVideo?.id)}>生成图片</Button> : <Tooltip title={productionAvailable ? '为当前分镜生成视频候选版本。' : productionReason}><Button disabled={!productionAvailable || regenerating} loading={regenerating} onClick={() => onGenerateVideo?.(selectedVideo?.id)}>生成视频</Button></Tooltip>}
        {promptKind === 'video' && selectedIndex === 0 && engineSettings.openingEnabled === true ? <Button onClick={() => setOpeningVariantsOpen(true)}>查看换开头提示词</Button> : null}
        {promptKind === 'video' ? <Button onClick={openH3Trace}>查看编译 Trace</Button> : null}
        <Button type="text" disabled={regenerating} onClick={() => onRetry?.(selectedVideo?.id)}>重试</Button>
      </div>
    </>}
    <p className="shuihuo-modal-note">{promptKind === 'visual' ? '画面版本从当前书的人物场景预设中管理和生成。' : editing ? '编辑的是当前分镜的视频描述；保存后会按已启用资产与约束重新编译最终视频提示词。' : compilingPrompt ? '正在读取已启用资产与约束编译的最终视频提示词。' : '视频版本只在当前分镜内生成和切换，不会覆盖其他分镜。'}</p>
    <Modal title="分镜编译与提交 Trace" open={h3TraceOpen} onCancel={() => setH3TraceOpen(false)} footer={null} width={980}>
      {h3TraceBusy ? <Alert type="info" showIcon message="正在读取编译 Trace…" /> : h3TraceError ? <Alert type="error" showIcon message="编译 Trace 读取失败" description={h3TraceError} /> : h3Trace?.legacy ? <Alert type="warning" showIcon message="旧数据只读" description={h3Trace?.notice || '旧版导演数据，无完整结构化 Trace'} /> : h3Segment ? <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Descriptions size="small" bordered column={2} items={[
          { key: 'director', label: '导演修订', children: h3Trace?.director_revision_id || '—' },
          { key: 'compilation', label: '编译修订', children: h3Trace?.compilation?.id || '—' },
          { key: 'segment', label: '分段', children: h3Segment.segment_key || `VIDEO ${selectedIndex + 1}` },
          { key: 'duration', label: '确定性时长', children: `${h3Segment.canonical_duration_ms || 0}ms → ${h3Segment.request_duration_ms || 0}ms` },
          { key: 'hash', label: 'Prompt Hash', children: h3Segment.compiled_prompt_hash || '—', span: 2 }
        ]} />
        <label className="shuihuo-form-label">实际提交视频模型的最终 VIDEO Prompt<Input.TextArea rows={16} readOnly value={h3Segment.compiled_prompt || ''} /></label>
        <label className="shuihuo-form-label">Compile Trace<Input.TextArea rows={12} readOnly value={JSON.stringify(h3Segment.compile_trace || {}, null, 2)} /></label>
      </Space> : <Alert type="info" showIcon message="当前分镜尚无编译记录" />}
    </Modal>
    <Modal title="分镜一 | 换开头提示词" open={openingVariantsOpen} onCancel={() => setOpeningVariantsOpen(false)} footer={null} width={860}>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Card size="small" title="分镜一（原始）"><pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{submittedVideoPrompt || '—'}</pre></Card>
        {(selectedVideo?.settingsState?.patch?.openingVariants || []).map(variant => (
          <Card key={variant.index} size="small" title={variant.label || `分镜一 | 换开头${variant.index}`} extra={<Tag color={variant.status === 'success' ? 'success' : 'error'}>{variant.status === 'success' ? '成功' : '失败'}</Tag>}>
            {variant.status === 'success' ? <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{variant.prompt}</pre> : <Alert type="error" showIcon message="AI 未生成出该变体，已阻断后续生产" description={variant.failureReason || '历史记录未保留失败原因；请在任务/日志中查看并重试失败步骤。'} />}
          </Card>
        ))}
        {!(selectedVideo?.settingsState?.patch?.openingVariants || []).length ? <Alert type="info" showIcon message="换开头变体尚未生成" /> : null}
      </Space>
    </Modal>
  </div>;
}
function PassiveMediaPoster({ className = 'batch-factory-media-poster', label, detail = '已生成 · 点击查看' }) {
  return <span className={className} role="img" aria-label={label}><PictureOutlined /><b>{label}</b><small>{detail}</small></span>;
}

function batchFactoryPassiveMediaUrl(media) {
  return String(media?.mediaUrl || media?.outputUrl || '').trim();
}

function batchFactoryPassivePosterUrl(media) {
  return String(
    media?.posterUrl
    || media?.thumbnailUrl
    || media?.coverUrl
    || media?.previewImageUrl
    || media?.firstFrameUrl
    || ''
  ).trim();
}

function PassiveMediaPreview({
  media,
  className = 'batch-factory-media-poster',
  label = '视频预览',
  detail = '已生成',
  onLoadedMetadata
}) {
  const mediaUrl = batchFactoryPassiveMediaUrl(media);
  const posterUrl = batchFactoryPassivePosterUrl(media);

  if (!mediaUrl && !posterUrl) {
    return <PassiveMediaPoster className={className} label={label} detail={detail} />;
  }

  const primeVideoFrame = event => {
    onLoadedMetadata?.(event);
    const element = event.currentTarget;
    const duration = Number(element?.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.08, duration / 2) : 0.05;
    try {
      if (element && Number(element.currentTime || 0) < 0.001 && target > 0) element.currentTime = target;
    } catch (_) {}
  };

  const ensureDecodedFrame = event => {
    const element = event.currentTarget;
    const duration = Number(element?.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.08, duration / 2) : 0.05;
    try {
      if (element && Number(element.currentTime || 0) < 0.001 && target > 0) element.currentTime = target;
    } catch (_) {}
  };

  return <span className={`${className} has-media`} role="img" aria-label={label}>
    {posterUrl
      ? <img src={posterUrl} alt="" loading="lazy" />
      : <ProductionMediaBoundary showDownload={false}><video
          src={mediaUrl}
          muted
          playsInline
          preload="auto"
          disablePictureInPicture
          tabIndex={-1}
          aria-hidden="true"
          onLoadedMetadata={primeVideoFrame}
          onLoadedData={ensureDecodedFrame}
        /></ProductionMediaBoundary>}
  </span>;
}


function batchFactoryParseMediaRatio(value, fallback = 9 / 16) {
  const parts = String(value || '').trim().split(/[:/x×]/).map(Number).filter(Number.isFinite);
  if (parts.length >= 2 && parts[0] > 0 && parts[1] > 0) return parts[0] / parts[1];
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric > 0 ? numeric : fallback;
}

function batchFactoryMediaOrientation(ratio) {
  const safe = Number.isFinite(Number(ratio)) && Number(ratio) > 0 ? Number(ratio) : 9 / 16;
  if (Math.abs(safe - 1) < 0.035) return 'square';
  return safe > 1 ? 'landscape' : 'portrait';
}

function batchFactoryRatioText(width, height, fallbackRatio = 9 / 16) {
  const w = Math.round(Number(width || 0));
  const h = Math.round(Number(height || 0));
  if (w > 0 && h > 0) {
    function gcd(a, b) { return b ? gcd(b, a % b) : a; }
    const divisor = gcd(w, h) || 1;
    return `${Math.round(w / divisor)}:${Math.round(h / divisor)}`;
  }
  const ratio = Number(fallbackRatio) > 0 ? Number(fallbackRatio) : 9 / 16;
  if (Math.abs(ratio - 16 / 9) < 0.04) return '16:9';
  if (Math.abs(ratio - 9 / 16) < 0.04) return '9:16';
  if (Math.abs(ratio - 1) < 0.04) return '1:1';
  return ratio > 1 ? `${Math.round(ratio * 100)}:100` : `100:${Math.round(100 / ratio)}`;
}

function batchFactoryStableMediaLabel(prefix, id) {
  const compact = String(id || '').replace(/[^a-z0-9]/gi, '').slice(-6).toUpperCase();
  return compact ? `${prefix}-${compact}` : prefix;
}
function InlineMediaLibrary({ book, versionsByVideo, productionStatus, mergeJob, aspectRatio = '16:9', h3CompilationMessage = '将按 10 秒确定性切段并编译最终 VIDEO', selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage, onOpenMerge }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const mediaVersions = versionsByVideo instanceof Map ? versionsByVideo : new Map();
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const video = videos[selectedIndex] || null;
  const versions = video ? (mediaVersions.get(video.id) || []) : [];
  const primary = primaryMediaVersion(video, versions);
  const currentProgress = video ? videoProgress.byVideo.get(String(video.id || '')) : null;
  const mediaFrameLayout = batchFactoryMediaOrientation(batchFactoryParseMediaRatio(video?.settingsState?.patch?.aspectRatio || video?.aspectRatio || aspectRatio, 16 / 9));
  const candidates = versions.filter(version => version.id !== primary?.id).slice(-4).reverse();
  useEffect(() => { if (!controlledSelectedVideoId) setLocalSelectedVideoId(book?.videos?.[0]?.id || ''); }, [book?.id]);
  useEffect(() => { if (!videos.some(item => item.id === selectedVideoId) && videos[0]?.id) selectVideo(videos[0].id); }, [videos.map(item => item.id).join('|')]);
  function move(direction) { const next = videos[selectedIndex + direction]; if (next) selectVideo(next.id); }
  if (!video && precompiledWorkspace.status === 'awaiting_compilation') return <button type="button" className="batch-factory-cell-shell batch-factory-inline-media is-empty" title="打开待编译的分镜视频" onClick={() => onManage?.(selectedVideoId)}>
    <header className="batch-factory-cell-head"><div><b>分镜视频</b><small>导演分镜卡已提取，尚未编译 VIDEO</small></div><span className="batch-factory-cell-state">待编译</span></header>
    <div className="batch-factory-cell-content batch-factory-video-content"><div className="batch-factory-cell-empty"><CloudUploadOutlined /><span>点击查看待编译的分镜视频；{h3CompilationMessage}。</span></div></div>
    <div className="batch-factory-cell-pager"><span>{precompiledWorkspace.cards.length} 张导演分镜卡待编译 · 点击进入</span></div>
  </button>;
  if (!video) return <div className="batch-factory-cell-shell batch-factory-inline-media is-empty"><div className="batch-factory-cell-empty">AI 推理后显示该书的分镜视频。</div></div>;
  const running = currentProgress?.status === 'running' || currentProgress?.status === 'queued';
  const mergeStatus = String(mergeJob?.status || '').toLowerCase();
  const mergeReady = mergeStatus === 'succeeded' && String(mergeJob?.outputUrl || '').trim();
  const mergeText = mergeReady ? `最终成片 ${batchFactoryStableMediaLabel('M', mergeJob.id)} · ${mergeJobSpeed(mergeJob).toFixed(1)}x`
    : mergeStatus === 'running' || mergeStatus === 'queued' ? '最终成片 · 合成中'
    : mergeStatus === 'failed' ? '最终成片 · 合成失败' : '最终成片 · 尚未合成';
  return <div className={`batch-factory-cell-shell batch-factory-inline-media is-${mediaFrameLayout}`}>
    <header className="batch-factory-cell-head">
      <div><b>{storyboardVideoLabel(video, selectedIndex)}</b><small>{running ? '正在生成' : primary?.mediaUrl ? '当前主版本' : '尚无可用视频'}</small></div>
      <span className={`batch-factory-cell-state${primary?.mediaUrl ? ' is-ready' : running ? ' is-running' : ''}`}>{primary?.mediaUrl ? batchFactoryStableMediaLabel('V', primary.id) : running ? '生成中' : '待生成'}</span>
    </header>
    <div className="batch-factory-cell-content batch-factory-video-content batch-factory-video-content-v13">
      <button type="button" className={`batch-factory-inline-media-primary${running ? ' is-producing' : ''}`} onClick={() => onManage?.(video.id)}>
        {primary?.mediaUrl ? <PassiveMediaPreview className="batch-factory-inline-media-poster" media={primary} label={`分镜 ${selectedIndex + 1} 已生成`} /> : <>{running ? <LoadingOutlined spin /> : <CloudUploadOutlined />}<span>{currentProgress?.message || '当前分镜暂无视频'}</span></>}
      </button>
      <div className="batch-factory-inline-media-versions"><span>候选</span>{candidates.map((candidate, index) => <button type="button" key={candidate.id} className="batch-factory-inline-media-version" onClick={() => onManage?.(video.id)}>{candidate.mediaUrl ? <PassiveMediaPreview className="batch-factory-inline-media-poster is-candidate" media={candidate} label={`候选 ${index + 1}`} detail="已生成" /> : <PictureOutlined />}</button>)}<button type="button" className="batch-factory-inline-media-version is-more" onClick={() => onManage?.(video.id)} title="打开完整片段库">•••</button></div>
      <button type="button" className={`batch-factory-inline-merge-status is-${mergeStatus || 'idle'}`} onClick={() => onOpenMerge?.()}><span>{mergeText}</span><b>查看</b></button>
    </div>
    <div className="batch-factory-cell-pager"><button type="button" disabled={selectedIndex === 0} onClick={() => move(-1)}><LeftOutlined /></button><span>{selectedIndex + 1} / {videos.length}</span><button type="button" disabled={selectedIndex >= videos.length - 1} onClick={() => move(1)}><RightOutlined /></button></div>
  </div>;
}

function MediaVersionPanel({ book, batchId, versionsByVideo, productionStatus, mergeJob, mergeJobs = [], mergeSettings = {}, mergeSpeed, onMergeSpeedChange, onMerge, onUpload, mergeAvailable = false, mergeReason = '', merging = false, onSaved, onDeleted, onRegenerate, onRetry, regenerating, productionAvailable = false, productionReason = '', initialVideoId = '', initialTab = 'clips', onOpenDirector }) {
  const videos = book?.videos || [];
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  const mediaVersions = versionsByVideo instanceof Map ? versionsByVideo : new Map();
  const [viewerKind, setViewerKind] = useState(initialTab === 'merges' ? 'merge' : 'video');
  const [selectedVideoId, setSelectedVideoId] = useState(initialVideoId || videos[0]?.id || '');
  const [selectedVersionId, setSelectedVersionId] = useState('');
  const [selectedMergePreviewId, setSelectedMergePreviewId] = useState('');
  const [savingTaskId, setSavingTaskId] = useState('');
  const [playRequested, setPlayRequested] = useState(false);
  const [mediaMetrics, setMediaMetrics] = useState({});
  const [primaryOverrides, setPrimaryOverrides] = useState({});
  const [followAudio, setFollowAudio] = useState(normalizeAudioSource(mergeSettings.audioSource) !== 'original' && mergeSettings.audioMergeEnabled === true);
  const [audioDurationSeconds, setAudioDurationSeconds] = useState(Number(mergeSettings.audioDurationSeconds || 0));
  const [audioBusy, setAudioBusy] = useState(false);
  const [pendingUploadKey, setPendingUploadKey] = useState('');
  const [sheetAnchor, setSheetAnchor] = useState('collapsed');
  const [sheetOffset, setSheetOffset] = useState(null);
  const [sheetDragging, setSheetDragging] = useState(false);
  const sheetRef = useRef(null);
  const sheetDragRef = useRef(null);
  const uploadInputRefs = useRef({});
  const mergeUploadInputRef = useRef(null);

  const completedMerges = mergeJobs.filter(job => job?.status === 'succeeded' && String(job?.outputUrl || '').trim());
  const mergeJobsNewest = mergeJobs
    .filter(job => String(job?.status || '').toLowerCase() !== 'failed')
    .sort((left, right) => String(right?.updatedAt || right?.createdAt || '').localeCompare(String(left?.updatedAt || left?.createdAt || '')));
  const selectedUpload = book?.settingsState?.patch?.primaryUploadSource || {};
  const selectedMerge = selectedUpload?.kind === 'merged'
    ? completedMerges.find(job => job.id === selectedUpload.mergeJobId) || null
    : defaultBookMerge(mergeJobs);
  const mergedReady = selectedMerge?.status === 'succeeded' && String(selectedMerge?.outputUrl || '').trim();
  const uploadVideo = selectedUpload?.kind === 'video' ? videos.find(video => video.id === selectedUpload.videoId) : null;
  const uploadVersion = uploadVideo ? (mediaVersions.get(uploadVideo.id) || []).find(version => version.id === selectedUpload.taskId) || null : null;
  const defaultRatio = batchFactoryParseMediaRatio(mergeSettings.aspectRatio, 9 / 16);

  function label(video) {
    return video?.label || video?.title || `分镜 ${videos.findIndex(item => item.id === video?.id) + 1}`;
  }

  function productionProgressFor(video) {
    return videoProgress.byVideo.get(String(video?.id || '')) || null;
  }

  function fallbackRatioFor(video) {
    return batchFactoryParseMediaRatio(
      video?.settingsState?.patch?.aspectRatio || video?.aspectRatio || mergeSettings.aspectRatio,
      defaultRatio
    );
  }

  function mediaInfo(version, video) {
    const cached = mediaMetrics[String(version?.id || '')] || {};
    const width = Number(cached.width || version?.width || version?.videoWidth || 0);
    const height = Number(cached.height || version?.height || version?.videoHeight || 0);
    const ratio = width > 0 && height > 0 ? width / height : fallbackRatioFor(video);
    return {
      width,
      height,
      ratio,
      orientation: batchFactoryMediaOrientation(ratio),
      ratioText: batchFactoryRatioText(width, height, ratio)
    };
  }

  function mergeInfo(job) {
    const key = `merge:${String(job?.id || '')}`;
    const cached = mediaMetrics[key] || {};
    const width = Number(cached.width || job?.width || job?.videoWidth || 0);
    const height = Number(cached.height || job?.height || job?.videoHeight || 0);
    const ratio = width > 0 && height > 0 ? width / height : defaultRatio;
    return {
      key,
      width,
      height,
      ratio,
      orientation: batchFactoryMediaOrientation(ratio),
      ratioText: batchFactoryRatioText(width, height, ratio)
    };
  }

  function rememberMediaMetrics(key, event) {
    const element = event.currentTarget;
    if (!(element?.videoWidth > 0 && element?.videoHeight > 0)) return;
    const next = {
      width: element.videoWidth,
      height: element.videoHeight,
      duration: Number.isFinite(element.duration) ? element.duration : 0
    };
    setMediaMetrics(current => {
      const previous = current[key];
      if (previous?.width === next.width && previous?.height === next.height && previous?.duration === next.duration) return current;
      return { ...current, [key]: next };
    });
  }

  function primeViewerFrame(key, event) {
    rememberMediaMetrics(key, event);
    const element = event.currentTarget;
    if (!element || !element.paused || Number(element.currentTime || 0) >= 0.001) return;
    const duration = Number(element.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.06, duration / 2) : 0.04;
    try {
      if (target > 0) element.currentTime = target;
    } catch (_) {}
  }

  function primaryFor(video, list) {
    const overrideId = primaryOverrides[String(video?.id || '')];
    return (overrideId ? list.find(version => version.id === overrideId) : null) || primaryMediaVersion(video, list);
  }

  function videoState(video) {
    const list = mediaVersions.get(video.id) || [];
    const main = primaryFor(video, list);
    const progress = productionProgressFor(video);
    const active = ['running', 'queued'].includes(String(progress?.status || '').toLowerCase());
    const failed = String(progress?.status || '').toLowerCase() === 'failed';
    return { list, main, progress, active, failed, ready: Boolean(main?.mediaUrl) };
  }

  const selectedVideo = videos.find(video => video.id === selectedVideoId) || videos[0] || null;
  const selectedVideoState = selectedVideo ? videoState(selectedVideo) : { list: [], main: null };
  const selectedVersion = selectedVideoState.list.find(version => version.id === selectedVersionId)
    || selectedVideoState.main
    || selectedVideoState.list.at(-1)
    || null;
  const mergePreview = mergeJobs.find(job => job.id === selectedMergePreviewId)
    || selectedMerge
    || mergeJobsNewest[0]
    || null;

  const readyVideos = videos.filter(video => videoState(video).ready);
  const missingVideos = videos.filter(video => !videoState(video).ready);
  const mergeReady = videos.length > 0 && readyVideos.length === videos.length;

  useEffect(() => {
    const next = videos.find(video => video.id === initialVideoId) || videos[0] || null;
    const precompiledFrame = resolvePrecompiledStoryboardFrame(book, initialVideoId);
    setViewerKind(initialTab === 'merges' ? 'merge' : 'video');
    setSelectedVideoId(next?.id || precompiledFrame?.key || '');
    setSelectedVersionId('');
    setSelectedMergePreviewId('');
    setPlayRequested(false);
    setMediaMetrics({});
    setPrimaryOverrides({});
    setSheetAnchor('collapsed');
    setSheetOffset(null);
  }, [book?.id, initialVideoId, initialTab]);

  useEffect(() => {
    setFollowAudio(normalizeAudioSource(mergeSettings.audioSource) === 'original' ? false : mergeSettings.audioMergeEnabled === true);
    setAudioDurationSeconds(Number(mergeSettings.audioDurationSeconds || 0));
  }, [book?.id, mergeSettings.audioSource, mergeSettings.audioMergeEnabled, mergeSettings.audioDurationSeconds]);

  useEffect(() => {
    if (!selectedVideo) {
      if (selectedVersionId) setSelectedVersionId('');
      return;
    }
    if (selectedVersionId && selectedVideoState.list.some(version => version.id === selectedVersionId)) return;
    setSelectedVersionId(selectedVideoState.main?.id || selectedVideoState.list.at(-1)?.id || '');
  }, [selectedVideoId, selectedVersionId, selectedVideoState.main?.id, selectedVideoState.list.map(version => version.id).join('|')]);

  useEffect(() => {
    if (selectedMergePreviewId && mergeJobs.some(job => job.id === selectedMergePreviewId)) return;
    setSelectedMergePreviewId(selectedMerge?.id || mergeJobsNewest[0]?.id || '');
  }, [selectedMergePreviewId, selectedMerge?.id, mergeJobsNewest.map(job => job.id).join('|')]);

  function uploadSourceKey(source = selectedUpload) {
    if (source?.kind === 'video' && source.videoId && source.taskId) return `video:${source.videoId}:${source.taskId}`;
    if (source?.kind === 'merged' && source.mergeJobId) return `merged:${source.mergeJobId}`;
    if (selectedMerge?.id) return `merged:${selectedMerge.id}`;
    return '';
  }

  useEffect(() => {
    setPendingUploadKey(uploadSourceKey());
  }, [book?.revision, selectedUpload?.kind, selectedUpload?.videoId, selectedUpload?.taskId, selectedUpload?.mergeJobId, selectedMerge?.id]);

  function selectVideoPreview(video, version = null, play = true) {
    if (!video) return;
    const state = videoState(video);
    const nextVersion = version || state.main || state.list.at(-1) || null;
    setViewerKind('video');
    setSelectedVideoId(video.id);
    setSelectedVersionId(nextVersion?.id || '');
    setPlayRequested(Boolean(play && nextVersion?.mediaUrl));
  }

  function selectMergePreview(job, play = true) {
    if (!job) return;
    setViewerKind('merge');
    setSelectedMergePreviewId(job.id);
    setPlayRequested(Boolean(play && job.status === 'succeeded' && String(job.outputUrl || '').trim()));
  }

  async function choosePrimary(video, version) {
    if (!video || !version || savingTaskId) return;
    const key = String(video.id);
    const previous = primaryOverrides[key];
    setPrimaryOverrides(current => ({ ...current, [key]: version.id }));
    setSavingTaskId(version.id);
    setSelectedVideoId(video.id);
    setSelectedVersionId(version.id);
    setViewerKind('video');
    setPlayRequested(Boolean(version.mediaUrl));
    try {
      await saveVideoOverride(batchId, book.id, video.id, { patch: { primaryMediaTaskId: version.id }, expectedRevision: Number(video.revision || 0) });
      await onSaved?.();
      message.success(`${label(video)} 已切换到 ${batchFactoryStableMediaLabel('V', version.id)}；下一次合成将读取它。`);
    } catch (error) {
      setPrimaryOverrides(current => {
        const next = { ...current };
        if (previous) next[key] = previous;
        else delete next[key];
        return next;
      });
      message.error(error?.message || '切换分镜主版本失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function uploadVideoMaterial(video, event) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (file.type !== 'video/mp4' && !/\.mp4$/i.test(file.name || '')) {
      message.error('请上传 MP4 视频');
      return;
    }
    const uploadKey = `upload-material-${video.id}`;
    setSavingTaskId(uploadKey);
    try {
      await uploadBookVideoMaterial(batchId, book.id, video.id, file);
      await onSaved?.();
      message.success('素材已上传，并已设为当前分镜主版本。');
    } catch (error) {
      message.error(error?.message || '上传视频素材失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function uploadMergeMaterial(event) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (file.type !== 'video/mp4' && !/\.mp4$/i.test(file.name || '')) {
      message.error('请上传 MP4 视频');
      return;
    }
    setSavingTaskId('upload-merge');
    try {
      await uploadBookMergeMaterial(batchId, book.id, file);
      await onSaved?.();
      message.success('成片已上传，可在合成成片中查看。');
    } catch (error) {
      message.error(error?.message || '上传成片失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function removeCandidate(video, version) {
    if (!video || !version || savingTaskId) return;
    setSavingTaskId(`delete-${version.id}`);
    try {
      await deleteProductionTask(batchId, book.id, video.id, version.id);
      await onDeleted?.();
      message.success('候选视频已从当前书片段库移除。');
    } catch (error) {
      message.error(error?.message || '删除候选视频失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function chooseUploadSource(source) {
    if (!book?.id || savingTaskId) return;
    setSavingTaskId(source.kind === 'merged' ? `upload-merged-${source.mergeJobId}` : `upload-video-${source.videoId}`);
    try {
      await saveBookOverride(batchId, book.id, { patch: { primaryUploadSource: source }, expectedRevision: Number(book.revision || 0) });
      await onSaved?.();
      setPendingUploadKey(uploadSourceKey(source));
      message.success(source.kind === 'merged'
        ? '合成成片已设为 121 上传主视频。'
        : `${source.videoId === videos[0]?.id ? '首分镜已设为上传主视频和作品封面。' : '分镜主版本已设为 121 上传主视频。'}`);
    } catch (error) {
      message.error(error?.message || '设置上传主视频失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function applyPendingUploadSource() {
    if (!pendingUploadKey) return;
    const [kind, first, second] = pendingUploadKey.split(':');
    if (kind === 'merged' && first) {
      await chooseUploadSource({ kind: 'merged', mergeJobId: first });
      return;
    }
    if (kind === 'video' && first && second) {
      await chooseUploadSource({ kind: 'video', videoId: first, taskId: second });
    }
  }

  const uploadOptions = [
    ...completedMerges.map(job => ({
      value: `merged:${job.id}`,
      label: `成片 ${batchFactoryStableMediaLabel('M', job.id)} · ${mergeJobSpeed(job).toFixed(1)}x`
    })),
    ...videos.flatMap(video => {
      const state = videoState(video);
      return state.main?.mediaUrl ? [{
        value: `video:${video.id}:${state.main.id}`,
        label: `${label(video)} · 主 ${batchFactoryStableMediaLabel('V', state.main.id)}`
      }] : [];
    })
  ];

  const uploadSourceText = selectedUpload?.kind === 'video'
    ? (uploadVideo && uploadVersion
      ? `${label(uploadVideo)} · ${batchFactoryStableMediaLabel('V', uploadVersion.id)}`
      : '所选分镜来源暂不可用')
    : mergedReady
      ? `合成成片 · ${batchFactoryStableMediaLabel('M', selectedMerge?.id)} · ${mergeJobSpeed(selectedMerge).toFixed(1)}x`
      : '尚未选择可上传视频';

  function visibleHeightForSheetAnchor(anchor, height) {
    if (anchor === 'full') return height;
    if (anchor === 'half') return Math.min(320, height - 74);
    return 74;
  }

  // The media library is rendered in an Ant Modal portal.  On some viewport
  // sizes its visual origin differs from the CSS offset parent, so a simple
  // `height - visible` transform leaves the sheet floating midway down the
  // modal.  Derive each snap point from the visible workspace instead.
  function sheetGeometry() {
    const sheet = sheetRef.current;
    const parent = sheet?.parentElement;
    if (!sheet || !parent) return null;
    const styleOffset = Number(String(sheet.style.transform || '').match(/translateY\(([-\d.]+)px\)/)?.[1] || 0);
    const rect = sheet.getBoundingClientRect();
    return {
      height: rect.height || 520,
      rawTop: rect.top - styleOffset,
      parentBottom: parent.getBoundingClientRect().bottom,
      appliedOffset: styleOffset
    };
  }

  function sheetAnchors(height) {
    const geometry = sheetGeometry();
    if (geometry) {
      const anchorOffset = anchor => geometry.parentBottom - visibleHeightForSheetAnchor(anchor, geometry.height) - geometry.rawTop;
      return { full: anchorOffset('full'), half: anchorOffset('half'), collapsed: anchorOffset('collapsed') };
    }
    const full = 0;
    const collapsed = Math.max(0, height - 74);
    const half = Math.max(0, height - Math.min(320, height - 74));
    return { full, half, collapsed };
  }

  function sheetAnchorOffset(anchor = sheetAnchor) {
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    return sheetAnchors(height)[anchor] ?? 0;
  }

  function rubberBand(value, min, max) {
    if (value < min) {
      const distance = min - value;
      return min - (distance * 0.28) / (1 + distance / 260);
    }
    if (value > max) {
      const distance = value - max;
      return max + (distance * 0.28) / (1 + distance / 260);
    }
    return value;
  }

  function startSheetDrag(event) {
    if (event.button !== undefined && event.button !== 0) return;
    const startOffset = sheetOffset == null
      ? (sheetGeometry()?.appliedOffset ?? sheetAnchorOffset())
      : sheetOffset;
    sheetDragRef.current = {
      pointerId: event.pointerId,
      startY: event.clientY,
      startOffset,
      lastOffset: startOffset,
      lastY: event.clientY,
      lastTime: performance.now(),
      velocity: 0
    };
    setSheetDragging(true);
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function moveSheetDrag(event) {
    const drag = sheetDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const now = performance.now();
    const deltaTime = Math.max(1, now - drag.lastTime);
    const instantVelocity = (event.clientY - drag.lastY) / deltaTime;
    drag.velocity = drag.velocity * 0.72 + instantVelocity * 0.28;
    drag.lastY = event.clientY;
    drag.lastTime = now;
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    const anchors = sheetAnchors(height);
    const nextOffset = rubberBand(drag.startOffset + event.clientY - drag.startY, anchors.full, anchors.collapsed);
    drag.lastOffset = nextOffset;
    setSheetOffset(nextOffset);
  }

  function endSheetDrag(event) {
    const drag = sheetDragRef.current;
    if (!drag || (event?.pointerId != null && drag.pointerId !== event.pointerId)) return;
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    const anchors = sheetAnchors(height);
    const current = Number.isFinite(drag.lastOffset)
      ? drag.lastOffset
      : (sheetOffset == null ? sheetAnchorOffset() : sheetOffset);
    const velocity = drag.velocity;
    const order = ['full', 'half', 'collapsed'];
    let target;
    if (Math.abs(velocity) > 0.42) {
      const currentIndex = order.indexOf(sheetAnchor);
      const nextIndex = velocity < 0 ? Math.max(0, currentIndex - 1) : Math.min(order.length - 1, currentIndex + 1);
      target = order[nextIndex];
    } else {
      target = order.reduce((best, name) => Math.abs(current - anchors[name]) < Math.abs(current - anchors[best]) ? name : best, 'collapsed');
    }
    sheetDragRef.current = null;
    setSheetDragging(false);
    setSheetAnchor(target);
    setSheetOffset(null);
  }

  function toggleSheet() {
    setSheetOffset(null);
    setSheetAnchor(current => current === 'collapsed' ? 'half' : current === 'half' ? 'full' : 'collapsed');
  }

  async function persistMediaAudioPatch(patch) {
    const result = await saveBookOverrideWithRetry(batchId, book.id, Number(book?.revision || 0), { patch });
    await onSaved?.();
    return result;
  }

  async function measureMergeAudioDuration({ force = false } = {}) {
    if (audioBusy) return Number(audioDurationSeconds || 0);
    const input = batchFactoryBookAudioInput(book);
    const tts = await batchFactoryBookTts(mergeSettings);
    const fingerprint = batchFactoryAudioFingerprint(input, tts);
    const savedFingerprint = String(mergeSettings.audioDurationFingerprint || '');
    const savedSeconds = Number(mergeSettings.audioDurationSeconds || 0);
    if (!force && savedSeconds > 0 && (mergeSettings.audioDurationManual === true || savedFingerprint === fingerprint)) {
      setAudioDurationSeconds(savedSeconds);
      return savedSeconds;
    }
    setAudioBusy(true);
    try {
      const measured = await generateBatchFactoryBookAudioMeasurement(book, mergeSettings);
      setAudioDurationSeconds(measured.durationSeconds);
      await persistMediaAudioPatch({ audioDurationSeconds: measured.durationSeconds, audioDurationFingerprint: measured.fingerprint, audioDurationManual: false });
      message.success(`已读取当前书配音时长：${measured.durationSeconds.toFixed(2)} 秒`);
      return measured.durationSeconds;
    } catch (error) {
      message.error(error?.message || '读取当前书配音时长失败；仍可手动填写秒数');
      return 0;
    } finally { setAudioBusy(false); }
  }

  async function toggleFollowAudio(checked) {
    if (normalizeAudioSource(mergeSettings.audioSource) === 'original') { message.info('当前为原声模式，不会合入 TTS。'); return; }
    setFollowAudio(checked);
    try {
      await persistMediaAudioPatch({ audioMergeEnabled: checked });
      if (checked && Number(audioDurationSeconds || 0) <= 0) await measureMergeAudioDuration();
    } catch (error) { message.error(error?.message || '保存合并跟随配音设置失败'); }
  }

  async function persistManualAudioDuration() {
    const seconds = Number(audioDurationSeconds || 0);
    if (!(seconds > 0)) return;
    try {
      await persistMediaAudioPatch({ audioDurationSeconds: seconds, audioDurationManual: true });
    } catch (error) { message.error(error?.message || '保存配音时长失败'); }
  }

  async function runCurrentMerge() {
    if (!mergeReady || !mergeAvailable || merging) return;
    let seconds = Number(audioDurationSeconds || 0);
    if (followAudio && seconds <= 0) seconds = await measureMergeAudioDuration();
    if (followAudio && seconds <= 0) return;
    onMerge?.({ speed: Number(mergeSpeed || 1), followAudio, audioDurationSeconds: seconds });
  }

  function renderThumb(media, info, labelText, className = 'batch-factory-media-pickstation-thumb') {
    return <PassiveMediaPreview
      media={media}
      className={`${className} is-${info.orientation}`}
      label={labelText}
      onLoadedMetadata={event => rememberMediaMetrics(String(media?.id || ''), event)}
    />;
  }

  function renderMergeRail() {
    const current = mergePreview || selectedMerge || mergeJobsNewest[0] || null;
    const currentInfo = mergeInfo(current);
    return <section className={`batch-factory-media-pickstation-rail-card is-merge${viewerKind === 'merge' ? ' is-active' : ''}`}>
      <header>
        <div><span className="batch-factory-media-pickstation-index">成</span><b>合成成片</b></div>
        <small>{current ? `当前 ${batchFactoryStableMediaLabel('M', current.id)}` : '尚无成片'}</small>
      </header>
      {current ? <button type="button" className="batch-factory-media-pickstation-main" onClick={() => selectMergePreview(current)}>
        {current.status === 'succeeded' && String(current.outputUrl || '').trim()
          ? renderThumb(current, currentInfo, `合成成片 ${batchFactoryStableMediaLabel('M', current.id)}`)
          : <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label="等待成片" detail={current.status === 'failed' ? '合成失败' : '处理中'} />}
        <span><b>{batchFactoryStableMediaLabel('M', current.id)}</b><small>{mergeJobSpeed(current).toFixed(1)}x · {currentInfo.ratioText}</small><em>{selectedMerge?.id === current.id ? '当前上传来源' : '历史成片'}</em></span>
      </button> : <div className="batch-factory-media-pickstation-empty">全部分镜选定主版本后，可在底部发起合成。</div>}
      <div className="batch-factory-media-pickstation-versions">
        {mergeJobsNewest.map(job => {
          const info = mergeInfo(job);
          const selected = job.id === current?.id;
          return <button
            type="button"
            key={job.id}
            className={`batch-factory-media-pickstation-version${selected ? ' is-selected' : ''}`}
            style={{ '--bf-media-ratio': info.ratio }}
            onClick={() => { setSelectedMergePreviewId(job.id); selectMergePreview(job); }}
            title={`查看 ${batchFactoryStableMediaLabel('M', job.id)}`}
          >
            <PassiveMediaPreview media={job} className="batch-factory-media-pickstation-version-media" label={batchFactoryStableMediaLabel('M', job.id)} onLoadedMetadata={event => rememberMediaMetrics(info.key, event)} />
            <span>{batchFactoryStableMediaLabel('M', job.id)}</span>
          </button>;
        })}
        <Tooltip title="上传成片">
          <span className="batch-factory-media-pickstation-add" title="上传成片">
            <input ref={el => { mergeUploadInputRef.current = el; }} type="file" accept="video/mp4,.mp4" hidden onChange={uploadMergeMaterial} />
            <Button size="small" loading={savingTaskId === 'upload-merge'} disabled={savingTaskId === 'upload-merge'} onClick={() => mergeUploadInputRef.current?.click()}>＋</Button>
          </span>
        </Tooltip>
      </div>
    </section>;
  }

  function renderVideoRail(video, index) {
    const state = videoState(video);
    const main = state.main;
    const info = mediaInfo(main, video);
    const active = viewerKind === 'video' && video.id === selectedVideoId;
    const uploadMain = selectedUpload?.kind === 'video' && selectedUpload.videoId === video.id && selectedUpload.taskId === main?.id;
    return <section key={video.id} className={`batch-factory-media-pickstation-rail-card${active ? ' is-active' : ''}`}>
      <header>
        <div><span className="batch-factory-media-pickstation-index">{String(index + 1).padStart(2, '0')}</span><b>{label(video)}</b></div>
        <small>{main ? `主 ${batchFactoryStableMediaLabel('V', main.id)}` : state.active ? '生成中' : state.failed ? '失败' : '待生成'}</small>
      </header>
      {main?.mediaUrl ? <button type="button" className="batch-factory-media-pickstation-main" onClick={() => selectVideoPreview(video, main)}>
        {renderThumb(main, info, `${label(video)} 主版本`)}
        <span><b>{batchFactoryStableMediaLabel('V', main.id)} · {info.ratioText}</b><small>{main.actualDurationSeconds ? `${Number(main.actualDurationSeconds).toFixed(1)}s · ` : ''}{state.list.length} 个版本</small><em>{uploadMain ? '上传来源' : '合成主版本'}</em></span>
      </button> : <button type="button" className="batch-factory-media-pickstation-main is-empty" onClick={() => selectVideoPreview(video, null, false)}>
        <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label={`分镜 ${index + 1}`} detail={state.active ? '生成中' : state.failed ? '生成失败' : '暂无视频'} />
        <span><b>{state.progress?.message || '暂无可用主版本'}</b><small>{state.active ? '任务会自动刷新' : '生成成功后显示版本'}</small></span>
      </button>}
      <div className="batch-factory-media-pickstation-versions">
        {state.list.map(version => {
          const versionInfo = mediaInfo(version, video);
          const isPrimary = version.id === main?.id;
          const isUpload = selectedUpload?.kind === 'video' && selectedUpload.videoId === video.id && selectedUpload.taskId === version.id;
          return <span className="batch-factory-media-pickstation-version-wrap" key={version.id}>
            <button
              type="button"
              className={`batch-factory-media-pickstation-version${isPrimary ? ' is-selected' : ''}`}
              style={{ '--bf-media-ratio': versionInfo.ratio }}
              onClick={() => choosePrimary(video, version)}
              disabled={savingTaskId === version.id}
              title={isPrimary ? '当前合成主版本' : `切换到 ${batchFactoryStableMediaLabel('V', version.id)} 并作为合成主版本`}
            >
              <PassiveMediaPreview media={version} className="batch-factory-media-pickstation-version-media" label={batchFactoryStableMediaLabel('V', version.id)} onLoadedMetadata={event => rememberMediaMetrics(String(version.id), event)} />
              <span>{batchFactoryStableMediaLabel('V', version.id)}</span>
              {isUpload ? <i>传</i> : null}
            </button>
            {!isPrimary && !isUpload ? <Popconfirm title="移除候选版本？" description="只移除当前书片段库引用，不删除远端媒体或历史合成记录。" okText="移除" cancelText="取消" onConfirm={() => removeCandidate(video, version)}>
              <button type="button" className="batch-factory-media-pickstation-version-remove" aria-label={`移除 ${batchFactoryStableMediaLabel('V', version.id)}`}>×</button>
            </Popconfirm> : null}
          </span>;
        })}
        <Tooltip title="上传素材">
          <span className="batch-factory-media-pickstation-add" title="上传素材">
            <input ref={element => { uploadInputRefs.current[video.id] = element; }} type="file" accept="video/mp4,.mp4" hidden disabled={savingTaskId === `upload-material-${video.id}`} onChange={event => uploadVideoMaterial(video, event)} />
            <Button size="small" loading={savingTaskId === `upload-material-${video.id}`} disabled={savingTaskId === `upload-material-${video.id}`} onClick={() => uploadInputRefs.current[video.id]?.click()}>＋</Button>
          </span>
        </Tooltip>
        {state.failed ? <Tooltip title={productionAvailable ? '重试当前分镜最后失败的视频任务。' : productionReason}><Button size="small" danger disabled={!productionAvailable || regenerating} onClick={() => onRetry?.(video.id)}>重试</Button></Tooltip> : null}
      </div>
    </section>;
  }

  const viewerMerge = viewerKind === 'merge' ? mergePreview : null;
  const viewerVideo = viewerKind === 'video' ? selectedVideo : null;
  const viewerVersion = viewerKind === 'video' ? selectedVersion : null;
  const viewerInfo = viewerKind === 'merge' ? mergeInfo(viewerMerge) : mediaInfo(viewerVersion, viewerVideo);
  const viewerUrl = viewerKind === 'merge'
    ? String(viewerMerge?.outputUrl || '').trim()
    : String(viewerVersion?.mediaUrl || '').trim();
  const viewerTitle = viewerKind === 'merge'
    ? (viewerMerge ? `合成成片 ${batchFactoryStableMediaLabel('M', viewerMerge.id)}` : '合成成片')
    : (viewerVideo ? label(viewerVideo) : '分镜');
  const viewerMeta = viewerKind === 'merge'
    ? (viewerMerge ? `${mergeJobSpeed(viewerMerge).toFixed(1)}x · ${viewerInfo.ratioText}` : '—')
    : (viewerVersion ? `${batchFactoryStableMediaLabel('V', viewerVersion.id)} · ${viewerInfo.ratioText}` : '—');

  const primaryVideoDurationRows = videos.map(video => {
    const main = videoState(video).main;
    if (!main) return { seconds: 0, exact: false };
    const actual = Number(main.actualDurationSeconds || mediaMetrics[String(main.id || '')]?.duration || 0);
    if (actual > 0) return { seconds: actual, exact: true };
    return { seconds: Number(main.requestedDurationSeconds || video.durationSeconds || 0), exact: false };
  });
  const videoTotalSeconds = primaryVideoDurationRows.reduce((sum, item) => sum + (Number.isFinite(item.seconds) ? item.seconds : 0), 0);
  const videoTotalIsExact = primaryVideoDurationRows.length > 0 && primaryVideoDurationRows.every(item => item.exact && item.seconds > 0);
  const autoMergeRateRaw = Number(audioDurationSeconds || 0) > 0 && videoTotalSeconds > 0 ? videoTotalSeconds / Number(audioDurationSeconds) : 0;
  const autoMergeRate = autoMergeRateRaw > 0 ? Math.min(4, Math.max(0.5, autoMergeRateRaw)) : 0;
  const estimatedFinalSeconds = autoMergeRate > 0 ? videoTotalSeconds / autoMergeRate : 0;

  const mergeTooltip = !mergeAvailable
    ? mergeReason
    : !mergeReady
      ? `还有 ${missingVideos.length} 个分镜没有可用主版本`
      : followAudio && Number(audioDurationSeconds || 0) <= 0
          ? '点击合成时会先自动生成配音并读取真实时长'
          : '按各分镜当前主版本合成；不会覆盖历史成片。';

  const currentSheetOffset = sheetOffset == null ? sheetAnchorOffset() : sheetOffset;

  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
  const movePrecompiledFrame = direction => {
    const frame = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + direction];
    if (frame) setSelectedVideoId(frame.key);
  };
  if (precompiledWorkspace.status === 'awaiting_compilation') {
    const card = selectedPrecompiledFrame?.card || {};
    const camera = card?.camera || {};
    const currentIndex = selectedPrecompiledFrame?.index || 0;
    return <div className="batch-factory-media-pickstation batch-factory-media-precompiled-workspace">
      <div className="batch-factory-media-pickstation-topline">
        <span>左侧当前分镜预览 · 右侧选择导演卡；编译完成后自动切换为 VIDEO 版本库</span>
        <b>0/{precompiledWorkspace.frames.length} 分镜已就绪</b>
      </div>
      <div className="batch-factory-media-pickstation-workspace">
        <section className="batch-factory-media-pickstation-viewer">
          <header>
            <div><b>VIDEO {String(currentIndex + 1).padStart(2, '0')}</b><small>导演分镜卡 · 待编译</small></div>
            <Tag color="default">待生成</Tag>
          </header>
          <div className="batch-factory-media-pickstation-stage">
            <div className="batch-factory-media-pickstation-frame is-landscape" style={{ '--bf-media-ratio': 16 / 9 }}>
              <div className="batch-factory-media-primary-empty"><PictureOutlined /><span>当前分镜暂不可播放视频</span></div>
            </div>
          </div>
          <footer><span>当前为导演分镜卡 {(currentIndex || 0) + 1}；{h3CompilationStatus(mergeSettings)} Prompt。</span><small>待编译</small></footer>
        </section>
        <aside className="batch-factory-media-pickstation-rail">
          <section className="batch-factory-media-pickstation-rail-card">
            <header><div><span className="batch-factory-media-pickstation-index">成</span><b>合成成片</b></div><small>尚无成片</small></header>
            <div className="batch-factory-media-pickstation-main is-empty"><PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label="合成成片" detail="全部 VIDEO 编译并生成后可合成" /><span><b>等待 VIDEO 生成</b><small>不会在导演阶段伪造成片</small></span></div>
          </section>
          {precompiledWorkspace.frames.map((frame, index) => {
            const frameCard = frame.card || {};
            const active = frame.key === selectedPrecompiledFrame?.key;
            return <section key={frame.key} className={`batch-factory-media-pickstation-rail-card${active ? ' is-active' : ''}`}>
              <header><div><span className="batch-factory-media-pickstation-index">{String(index + 1).padStart(2, '0')}</span><b>VIDEO {String(index + 1).padStart(2, '0')}</b></div><small>待生成</small></header>
              <button type="button" className="batch-factory-media-pickstation-main is-empty" onClick={() => setSelectedVideoId(frame.key)}>
                <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label={`分镜 ${index + 1}`} detail="暂无视频" />
                <span><b>{String(frameCard.source_text || '导演分镜卡')}</b><small>{frameCard.action ? `动作：${frameCard.action}` : '等待最终 VIDEO 编译'}</small></span>
              </button>
              <div className="batch-factory-media-pickstation-versions"><Button size="small" type={active ? 'primary' : 'default'} onClick={() => onOpenDirector?.(frame.key)}>打开分镜提示词</Button></div>
            </section>;
          })}
        </aside>
      </div>
      <section className="batch-factory-media-bottom-sheet is-precompiled-collapsed">
        <div className="batch-factory-media-bottom-sheet-bar">
          <div className="batch-factory-media-bottom-sheet-ready"><span><PictureOutlined /></span><div><b>0 / {precompiledWorkspace.frames.length} 分镜已选好</b><small>先完成最终 VIDEO 编译，再生成、选择与合成视频。</small></div></div>
          <div className="batch-factory-media-bottom-sheet-quick"><Button size="small" onClick={() => onOpenDirector?.(selectedPrecompiledFrame?.key || '')}>查看导演提示词</Button><Tooltip title="等待最终 VIDEO 编译"><Button type="primary" disabled>合成当前书</Button></Tooltip><Tooltip title="等待可上传视频"><Button size="small" className="batch-factory-upload-network-button" disabled>上传网络</Button></Tooltip></div>
        </div>
      </section>
    </div>;
  }

  return <div className="batch-factory-media-pickstation">
    <div className="batch-factory-media-pickstation-topline">
      <span>左侧看片 · 右侧选片 · 小版本点击后立即成为该分镜的合成主版本</span>
      <b>{readyVideos.length}/{videos.length} 分镜已就绪</b>
    </div>

    <div className="batch-factory-media-pickstation-workspace">
      <section className="batch-factory-media-pickstation-viewer">
        <header>
          <div><b>{viewerTitle}</b><small>{viewerMeta}</small></div>
          <Tag color={viewerKind === 'merge' ? 'gold' : 'success'}>{viewerKind === 'merge' ? '成片预览' : '合成主版本'}</Tag>
        </header>
        <div className="batch-factory-media-pickstation-stage">
          <div className={`batch-factory-media-pickstation-frame is-${viewerInfo.orientation}`} style={{ '--bf-media-ratio': viewerInfo.ratio }}>
            {viewerUrl ? <ProductionMediaBoundary showDownload={false}><video
              key={`${viewerKind}:${viewerMerge?.id || viewerVersion?.id || ''}`}
              src={viewerUrl}
              controls
              preload="auto"
              playsInline
              autoPlay={playRequested}
              onLoadedMetadata={event => primeViewerFrame(viewerKind === 'merge' ? viewerInfo.key : String(viewerVersion?.id || ''), event)}
              onLoadedData={event => primeViewerFrame(viewerKind === 'merge' ? viewerInfo.key : String(viewerVersion?.id || ''), event)}
            /></ProductionMediaBoundary> : <div className="batch-factory-media-primary-empty">
              {viewerVideo && videoState(viewerVideo).active ? <LoadingOutlined spin /> : <PictureOutlined />}
              <span>{viewerKind === 'merge' ? '当前没有可播放的合成成片' : '当前分镜暂无可播放视频'}</span>
            </div>}
          </div>
        </div>
        <footer>
          {viewerKind === 'video' && viewerVideo ? <span>正在看 <b>{viewerVersion ? batchFactoryStableMediaLabel('V', viewerVersion.id) : '—'}</b>　｜　合成用 <b>{videoState(viewerVideo).main ? batchFactoryStableMediaLabel('V', videoState(viewerVideo).main.id) : '未选择'}</b></span>
            : <span>历史成片只用于预览和上传选择，不会自动改写分镜。</span>}
          <small>{viewerInfo.width > 0 && viewerInfo.height > 0 ? `${viewerInfo.width} × ${viewerInfo.height} · ` : ''}{viewerInfo.ratioText}</small>
        </footer>
      </section>

      <aside className="batch-factory-media-pickstation-rail">
        {renderMergeRail()}
        {videos.map(renderVideoRail)}
      </aside>
    </div>

    <section
      ref={sheetRef}
      className={`batch-factory-media-bottom-sheet is-${sheetAnchor}${sheetDragging ? ' is-dragging' : ''}`}
      style={{ transform: `translateY(${currentSheetOffset}px)` }}
    >
      <div
        className="batch-factory-media-bottom-sheet-handle"
        onPointerDown={startSheetDrag}
        onPointerMove={moveSheetDrag}
        onPointerUp={endSheetDrag}
        onPointerCancel={endSheetDrag}
      ><i /></div>

      <div className="batch-factory-media-bottom-sheet-bar">
        <div className="batch-factory-media-bottom-sheet-ready">
          <span className={mergeReady ? 'is-ready' : ''}><PictureOutlined /></span>
          <div><b>{readyVideos.length} / {videos.length} 分镜已选好</b><small>{viewerKind === 'video' && viewerVideo ? `当前：${label(viewerVideo)} · ${videoState(viewerVideo).main ? batchFactoryStableMediaLabel('V', videoState(viewerVideo).main.id) : '未选主版本'}` : uploadSourceText}</small></div>
        </div>
        <div className="batch-factory-media-bottom-sheet-quick">
          {!followAudio ? <InputNumber min={0.5} max={4} step={0.1} value={mergeSpeed} onChange={value => onMergeSpeedChange?.(Number(value || 1))} addonAfter="x" /> : <span className="batch-factory-audio-quick-state">{audioBusy ? '读取配音…' : `配音 ${Number(audioDurationSeconds || 0).toFixed(2)}s · 自动 ${autoMergeRate > 0 ? autoMergeRate.toFixed(3) : '—'}x`}</span>}
          <Button size="small" onClick={toggleSheet}>{sheetAnchor === 'collapsed' ? '展开' : sheetAnchor === 'half' ? '全展开' : '收起'}</Button>
          <Tooltip title={mergeTooltip}><Button type="primary" loading={merging} disabled={!mergeAvailable || merging || !mergeReady} onClick={runCurrentMerge}>合成当前书</Button></Tooltip>
          <Tooltip title="上传当前小说"><Button size="small" className="batch-factory-upload-network-button" onClick={onUpload} disabled={!mergedReady && !uploadVersion?.mediaUrl}>上传网络</Button></Tooltip>
        </div>
      </div>

      <div className="batch-factory-media-bottom-sheet-content">
        <div className="batch-factory-media-bottom-sheet-grid">
          <section>
            <h3>合成控制</h3>
            <div className="batch-factory-media-bottom-sheet-row"><span>本次倍率</span>{followAudio ? <b>{autoMergeRate > 0 ? `自动 · ${autoMergeRate.toFixed(3)}x` : '自动计算'}</b> : <InputNumber min={0.5} max={4} step={0.1} value={mergeSpeed} onChange={value => onMergeSpeedChange?.(Number(value || 1))} addonAfter="x" />}</div>
            <div className="batch-factory-media-bottom-sheet-row"><span>合并跟随配音</span><div className="batch-factory-follow-audio-control"><Switch checked={followAudio} loading={audioBusy} onChange={toggleFollowAudio} /><small>可事后开启并自动读取当前书配音时长</small></div></div>
            {followAudio ? <><div className="batch-factory-media-bottom-sheet-row"><span>配音目标</span><div className="batch-factory-audio-duration-edit"><InputNumber min={0.1} step={0.01} value={audioDurationSeconds || undefined} onChange={value => setAudioDurationSeconds(Number(value || 0))} onBlur={persistManualAudioDuration} addonAfter="s" /><Button size="small" loading={audioBusy} onClick={() => measureMergeAudioDuration({ force: true })}>重新读取</Button></div></div><div className="batch-factory-media-bottom-sheet-row"><span>主视频总时长</span><b>{videoTotalSeconds > 0 ? `${videoTotalIsExact ? '' : '≈ '}${videoTotalSeconds.toFixed(2)}s` : '等待读取'}</b></div><div className="batch-factory-media-bottom-sheet-row"><span>自动倍率</span><b>{autoMergeRate > 0 ? `${autoMergeRate.toFixed(3)}x` : '—'}</b></div><div className="batch-factory-media-bottom-sheet-row"><span>预计成片</span><b>{estimatedFinalSeconds > 0 ? `${estimatedFinalSeconds.toFixed(2)}s` : '—'}</b></div></> : null}
            <div className="batch-factory-media-bottom-sheet-actions"><Tooltip title={mergeTooltip}><Button type="primary" loading={merging || audioBusy} disabled={!mergeAvailable || merging || !mergeReady} onClick={runCurrentMerge}>{mergeJob?.status === 'succeeded' ? '重新合成' : '合成当前书'}</Button></Tooltip></div>
          </section>

          <section>
            <h3>上传与成片</h3>
            <div className="batch-factory-media-bottom-sheet-row"><span>当前上传来源</span><b>{uploadSourceText}</b></div>
            <Select value={pendingUploadKey || undefined} onChange={setPendingUploadKey} options={uploadOptions} placeholder="选择成片或分镜主版本" style={{ width: '100%' }} />
            <div className="batch-factory-media-bottom-sheet-actions">
              <Button onClick={applyPendingUploadSource} disabled={!pendingUploadKey || pendingUploadKey === uploadSourceKey() || Boolean(savingTaskId)} loading={Boolean(savingTaskId)}>设为上传来源</Button>
              <Tooltip title="上传当前小说"><Button className="batch-factory-upload-network-button" onClick={onUpload} disabled={!mergedReady && !uploadVersion?.mediaUrl}>上传网络</Button></Tooltip>
            </div>
          </section>

          <section className="is-wide">
            <h3>本次合成将使用</h3>
            <div className="batch-factory-media-bottom-sheet-snapshot">
              {videos.map((video, index) => {
                const main = videoState(video).main;
                return <button type="button" key={video.id} onClick={() => selectVideoPreview(video, main)}><b>分镜 {String(index + 1).padStart(2, '0')}</b><span>{main ? batchFactoryStableMediaLabel('V', main.id) : '未就绪'}</span></button>;
              })}
            </div>
          </section>
        </div>
        <p className="batch-factory-media-bottom-sheet-note">拖拽柄支持收起 / 半展开 / 全展开。超出边界时使用阻尼橡皮筋，松手根据手势速度选择吸附锚点。上传网络继续进入现有视频管理系统登录、组织、来源书城、解压数量与重复上传检查。</p>
      </div>
    </section>
  </div>;
}


function BatchLogs({ automationStatus, productionStatus, mergeStatus, error, books = [] }) {
  const jobs = productionStatus?.jobs || [];
  const merges = mergeStatus?.jobs || [];
	const activeMerge = merges.find(merge => ['running', 'queued'].includes(String(merge?.status || '').toLowerCase()) && String(merge?.progressPhase || '').toLowerCase() !== 'queued');
	const queuedMerges = merges.filter(merge => String(merge?.status || '').toLowerCase() === 'queued').slice().sort((left, right) => String(left?.createdAt || '').localeCompare(String(right?.createdAt || '')));
	const mergeQueueDetail = merge => {
	  const status = String(merge?.status || '').toLowerCase();
	  if (status === 'running') return `当前工位 · ${merge.progressPhase || '处理中'} · 已处理 ${Number(merge.progressCurrent || 0)} / ${Number(merge.progressTotal || 0)} 个分镜`;
	  if (status === 'queued') {
	    const ahead = Math.max(0, queuedMerges.findIndex(item => item?.id === merge?.id)) + (activeMerge ? 1 : 0);
	    return `合成队列 · 前方 ${ahead} 项${activeMerge ? ` · 当前工位 · ${activeMerge.progressPhase || '处理中'}` : ' · 等待工位'}`;
	  }
	  return merge.outputUrl || batchFactoryVisibleError(merge.errorMessage) || '等待合并结果';
	};
	const automationBookDetail = book => {
	  if (String(book?.stage || '').toLowerCase() !== 'merge') return '';
	  const bookID = String(book?.bookId || '');
	  const bookMerges = merges.filter(merge => String(merge?.bookId || '') === bookID).slice().sort((left, right) => String(right?.createdAt || '').localeCompare(String(left?.createdAt || '')));
	  const activeBookMerge = bookMerges.find(merge => ['running', 'queued'].includes(String(merge?.status || '').toLowerCase())) || bookMerges[0];
	  return activeBookMerge ? mergeQueueDetail(activeBookMerge) : '';
	};
  const automationBooks = Array.isArray(automationStatus?.books) ? automationStatus.books : [];
  const automationCounts = automationStatus?.counts || {};
  const runMode = String(automationStatus?.runMode || 'full_submit');
  const modeDescription = runMode === 'storyboard_only'
    ? '当前任务仅生成分镜与最终 VIDEO Prompt，停在“待生成视频”；不会提交视频、合成或上传。'
    : runMode === 'video_no_submit'
      ? '当前任务会生成视频与合成成片，停在“待上传”；刷新只回读状态，不会额外提交新任务。'
      : '自动生产会在成片后提交视频管理系统，并等待视频管理系统回读确认；未确认前不会显示上传成功。';
  const giantWaitingBooks = books.filter(book => book?.sourceMetadata?.giantMaterialId && book?.sourceMetadata?.contentPending === true);
  return <><Alert type={error ? 'warning' : 'info'} showIcon message={error || '实时读取 V12 自动生产、视频与合并状态'} description={modeDescription} />
    {automationStatus?.state && automationStatus.state !== 'idle' ? <Alert type={['paused', 'needs_attention', 'unavailable'].includes(automationStatus.state) ? 'warning' : automationStatus.state === 'completed' ? 'success' : 'info'} showIcon message={automationStatus.state === 'unavailable' ? '自动生产状态暂不可读' : automationStatus.state === 'paused' ? '自动生产 · 已暂停' : `自动生产 · ${automationStatus.state}`} description={automationStatus.state === 'unavailable' ? '不影响当前书的手动提取、配音、VIDEO 或上传操作。' : automationStatus.state === 'paused' ? `${automationStatus.lastError || '失败项最终自动补跑后仍未成功，已暂停等待处理。'}；并行 ${Number(automationStatus.concurrency || 2)} 本；就绪 ${Number(automationCounts.ready || 0)} / ${Number(automationCounts.total || automationBooks.length)}；失败 ${Number(automationCounts.failed || 0)}；阻塞 ${Number(automationCounts.blocked || 0)}` : `并行 ${Number(automationStatus.concurrency || 2)} 本；就绪 ${Number(automationCounts.ready || 0)} / ${Number(automationCounts.total || automationBooks.length)}；执行中 ${Number(automationCounts.running || 0)}；失败 ${Number(automationCounts.failed || 0)}；阻塞 ${Number(automationCounts.blocked || 0)}`} /> : null}
    <div className="batch-factory-log-list">{giantWaitingBooks.map(book => <section key={`giant-source-${book.id}`}><strong>巨量素材 · {book.title || book.bookId}</strong><span>等待正文读取</span><p>{book.sourceMetadata?.originalReadStage === 'ocr' ? '正在等待滚屏 OCR 回填正文；正文到位后自动转为滚屏 OCR。' : '正在优先通过书城读取正文；失败后自动转为滚屏 OCR。'}</p></section>)}{automationBooks.map(book => <section key={`automation-${book.bookId}`}><strong>自动生产 · {book.title || book.bookId}</strong><span>{book.stage || 'pending'} · {book.retryRequested === true ? '重试排队中' : book.status || 'pending'} · {book.updatedAt || '—'}</span><p>{book.retryRequested === true ? '已接受手动重试，正在等待生产工位；无需再次点击重试。' : automationBookDetail(book) || batchFactoryVisibleError(book.message) || '等待自动生产'}{Number(book.retryCount || 0) > 0 ? ` · 自动重试 ${Math.min(3, Number(book.retryCount || 0))}/3` : ''}{book.retryAt ? ` · 下次重试：北京时间 ${formatBeijingDatetimeLocal(book.retryAt)}` : ''}{book.error ? ` · ${batchFactoryVisibleError(book.error)}` : ''}</p></section>)}{jobs.map(job => <section key={job.id}><strong>生成任务 · {job.status}</strong><span>{job.bookId || '批量任务'} · {job.updatedAt || job.createdAt || '—'}</span>{(job.tasks || []).map(task => { const durations = [['目标', task.targetDurationSeconds], ['请求', task.requestedDurationSeconds], ['实际', task.actualDurationSeconds]].filter(([, value]) => Number(value) > 0).map(([label, value]) => `${label} ${Number(value).toFixed(2).replace(/\.00$/, '')}s`).join(' · '); return <p key={task.id}>{task.videoId} · {task.status}{durations ? ` · ${durations}` : ''}{task.errorMessage ? ` · ${batchFactoryVisibleError(task.errorMessage)}` : ''}</p>; })}</section>)}{merges.map(merge => <section key={merge.id}><strong>合并任务 · {merge.status}</strong><span>{merge.updatedAt || merge.createdAt || '—'}</span><p>{mergeQueueDetail(merge)}</p></section>)}{!giantWaitingBooks.length && !automationBooks.length && !jobs.length && !merges.length ? <p>当前没有自动生产、视频或合并任务。</p> : null}</div>
  </>;
}

function localStartTime() {
  const now = new Date();
  return new Date(now.getTime() - now.getTimezoneOffset() * 60_000).toISOString().slice(0, 19);
}

function UploadNetwork({ batch, books, selectedBookIds, productionStatus, mergeStatus, mode, onClose, onOpenPublish, onRefresh }) {
  const [account, setAccount] = useState(null);
  const [publishSession, setPublishSession] = useState(null);
  const [manifest, setManifest] = useState(null);
  const [results, setResults] = useState([]);
  const [busy, setBusy] = useState(false);
  const [uploadingBookId, setUploadingBookId] = useState('');
  const [organizations, setOrganizations] = useState([]);
  const [organizationID, setOrganizationID] = useState('');
  const [organizationsError, setOrganizationsError] = useState('');
  const reupload = mode === 'reupload';
  const targetBooks = mode === 'all' ? books : books.filter(book => selectedBookIds.includes(book.id));
  const completedTasks = (productionStatus?.jobs || []).flatMap(job => job?.tasks || []);
  const hasBookUploadSource = book => {
    const source = book?.settingsState?.patch?.primaryUploadSource || {};
    if (source.kind === 'video') return completedTasks.some(task => task?.id === source.taskId && task?.videoId === source.videoId && task?.status === 'succeeded' && String(task?.mediaUrl || '').trim());
    return (mergeStatus?.jobs || []).some(job => job?.bookId === book?.id && job?.status === 'succeeded' && String(job?.outputUrl || '').trim());
  };
  const uploadMediaReady = targetBooks.length > 0 && targetBooks.every(hasBookUploadSource);
  const publishSettings = batch?.settingsState?.patch?.publishSettings || {};
  const defaultOrganizationIDs = [...new Set(targetBooks
    .map(book => String(effectiveBookSettings(batch, book).publishSettings?.organization || '').trim())
    .filter(Boolean))];
  const defaultOrganizationID = defaultOrganizationIDs.length === 1 ? defaultOrganizationIDs[0] : '';
  const organizationSelectionMixed = defaultOrganizationIDs.length > 1;
  const hasSingleBookPublishOverride = targetBooks.some(book => Object.hasOwn(book?.settingsState?.patch || {}, 'publishSettings') || Object.hasOwn(book?.settingsState?.patch || {}, 'publishRewriteEnabled'));
  const materialReuseLabel = publishSettings.materialReuse === true ? '复用' : '不复用';
  const horizontalFlipLabel = publishSettings.horizontalFlip === true ? '翻转' : '不翻转';
  const missingPublishMapping = targetBooks.filter(book => {
    const settings = effectiveBookSettings(batch, book).publishSettings || {};
    return !completePublishMapping(settings, book);
  });
  const missingDecompression = targetBooks.filter(book => decompressionCountForUpload(batch, book) < 1);
  const alreadyUploaded = targetBooks.filter(book => isBookUploadedTo121(book));
  const requiresReupload = alreadyUploaded.length > 0 && !reupload;
  async function verify121Session() {
    const [config, environment] = await Promise.all([getWebSubmitConfig(), checkWebSubmitEnvironment()]);
    const settings = config?.settings || config?.config || {};
    setAccount(settings);
    const session = Boolean((environment?.checks || []).some(check => ['视频管理系统登录会话', '121 后台登录会话', '目标站登录会话'].includes(check.name) && check.ok));
    if (!session) {
      setPublishSession({ environment, visible: null });
      return false;
    }
    const visible = await testWebSubmitVisible();
    setPublishSession({ environment, visible });
    return visible?.ok === true;
  }
  useEffect(() => {
    if (!mode) return undefined;
    setManifest(null);
    setResults([]);
    setOrganizationID(defaultOrganizationID);
    let active = true;
    Promise.all([verify121Session(), get121OrganizationOptions()]).then(([, response]) => {
      if (!active) return;
      setOrganizations(Array.isArray(response?.organizations) ? response.organizations : []);
    }).catch(error => { if (active) { setOrganizationsError(error?.message || '121 组织目录读取失败'); setPublishSession({ error: error?.message || '121 后台会话验证失败' }); } });
    return () => { active = false; };
  }, [mode, defaultOrganizationID]);
  useEffect(() => {
    if (!uploadingBookId || !onRefresh) return undefined;
    const timer = setInterval(() => {
      if (document.hidden) return; // 切到后台时不刷，上传在服务端照常进行
      onRefresh().catch(() => {});
    }, 5000);
    return () => clearInterval(timer);
  }, [uploadingBookId, onRefresh]);
  const publishSessionReady = Boolean(publishSession?.environment?.ok && publishSession?.visible?.ok);
  async function prepareUpload() {
    setBusy(true);
    try {
      if (!(await verify121Session())) throw new Error('请先在发布统一登录并验证 121 后台账号');
      if (missingPublishMapping.length) throw new Error(`当前书缺少来源书城，无法上传：${missingPublishMapping.map(book => book.title || book.bookId).join('、')}`);
      if (missingDecompression.length) throw new Error(`请添加解压，才能上传 AI 前贴视频：${missingDecompression.map(book => book.title || book.bookId).join('、')}`);
      if (requiresReupload) throw new Error(`当前书已上传，请使用重新上传：${alreadyUploaded.map(book => book.title || book.bookId).join('、')}`);
      setManifest(targetBooks.map(book => {
        const effectiveSettings = effectiveBookSettings(batch, book);
        return {
          id: book.id,
          title: book.title || book.bookId,
          txt: `${book.bookId}.txt`,
          mp4: `${book.bookId}.mp4`,
          videoType: effectiveSettings.publishSettings?.uploadVideoType === 'individual' ? '独立 VIDEO' : '最终合成视频',
          jieyaNum: decompressionCountForUpload(batch, book)
        };
      }));
    } catch (error) { message.error(error?.message || '生成上传清单失败'); } finally { setBusy(false); }
  }
  async function submitTo121() {
    setBusy(true);
    const submitted = [];
    try {
      for (const book of targetBooks) {
        try {
          setUploadingBookId(book.id);
          setResults(current => [...current.filter(item => item.id !== book.id), { id: book.id, title: book.title || book.bookId, pending: true, message: reupload ? '再次上传中：正在验证 121 会话、上传 AI 前贴与提交 TXT' : '正在上传：正在验证 121 会话、上传 AI 前贴与提交 TXT' }]);
          const result = await submitBookTo121(batch.id, book.id, { organization: organizationID, category: 'NEW_BOOK', startTime: localStartTime(), textModelId: effectiveBookSettings(batch, book).textModelId, reupload });
          submitted.push({ id: book.id, title: book.title || book.bookId, ok: true, result: resultData(result) });
          await onRefresh?.();
        } catch (error) {
          submitted.push({ id: book.id, title: book.title || book.bookId, ok: false, error: error?.message || '121 提交失败' });
          // Each book has its own confirmation and readback. One remote
          // rejection must not hold the remaining ready books in this queue.
        }
      }
      setResults(submitted);
      const failed = submitted.filter(item => !item.ok);
      if (failed.length) message.warning(`${failed.length} 本提交失败，其余 ${submitted.length - failed.length} 本已继续提交；失败原因已逐本保留。`);
      else message.success(submitted.every(item => item.result?.status === 'confirmed') ? '已提交到 121，并已完成后台列表回读。' : '已提交到 121，正在等待后台列表回读。');
    } catch (error) { message.error(error?.message || '121 提交失败'); } finally { setUploadingBookId(''); setBusy(false); }
  }
  const liveUploadBook = books.find(book => book.id === uploadingBookId);
  const liveUploadProgress = uploadProgressForBook(liveUploadBook);
  return <Modal title={`上传网络 · ${mode === 'all' ? '提交全部' : reupload ? '重新上传' : '提交选中'}`} open={Boolean(mode)} onCancel={onClose} footer={null} width={700} destroyOnClose>
    <Space direction="vertical" size={14} style={{ width: '100%' }}>
      <Alert type="info" showIcon message={`已选择 ${targetBooks.length} 本小说`} description="每本书会作为一个独立的 121 包提交：`书号.txt` 为正文，`书号.mp4` 为自定义 AI 头部视频。提交后以 121 后台列表回读为准。" />
      <Alert type="info" showIcon message="本次上传素材设置" description={hasSingleBookPublishOverride ? '上传时按每本书的单书发布覆盖优先、批量统一设置兜底；生效值会随确认单冻结，不会修改 121 后台配置方案。' : `本次素材使用：${materialReuseLabel}；本次水平翻转：${horizontalFlipLabel}。这两个值会随确认单冻结，不会修改 121 后台配置方案。`} />
      {missingPublishMapping.length ? <Alert type="warning" showIcon message="存在缺少来源书城的书" description={missingPublishMapping.map(book => book.title || book.bookId).join('、')} /> : null}
      <Alert type="info" showIcon message="逐本识别发布信息" description="提交前会按每本书正文自动识别男女频、121 风格和标签；已有手动或已识别结果会直接复用。识别失败会终止当前书，不会改用批量默认值。" />
      {missingDecompression.length ? <Alert type="warning" showIcon message="请添加解压" description={`以下小说的解压数量为 0，不能附带 AI 前贴视频上传：${missingDecompression.map(book => book.title || book.bookId).join('、')}`} /> : null}
      {requiresReupload ? <Alert type="warning" showIcon message="存在已上传小说" description="普通上传不会重复提交已成功回读的小说。请从该书“查看资料”中选择重新上传。" /> : null}
      {!uploadMediaReady ? <Alert type="warning" showIcon message="存在尚未准备好的上传主视频" description="每本书需要先在片段库选择可用分镜主版本，或完成该书的最终合成成片；确认单不会提交空视频或未完成视频。" /> : null}
      <Tag
        className={`batch-factory-publish-session-tag ${publishSessionReady ? 'is-ready' : 'is-unverified'}`}
        color={publishSessionReady ? 'green' : 'error'}
      >
        {publishSessionReady ? `121 后台已登录：${account?.username || '当前账号'}` : '121 后台账号尚未登录或未验证'}
      </Tag>
      {publishSession?.error ? <Alert type="warning" showIcon message="121 后台会话验证失败" description={publishSession.error} /> : null}
      {!publishSessionReady ? <Button onClick={onOpenPublish}>前往发布统一登录并验证</Button> : null}
      {organizationsError ? <Alert type="warning" showIcon message="121 组织目录读取失败" description={organizationsError} /> : null}
      {organizationSelectionMixed ? <Alert type="warning" showIcon message="所选小说存在不同的单书组织覆盖" description="请为本次上传统一选择一个组织归属；此处改选只作用于本次确认单，不会回写任何统一或单书配置。" /> : null}
      <Select value={organizationID || undefined} placeholder="请选择 121 组织归属" style={{ width: '100%' }} onChange={setOrganizationID} options={organizations.map(item => ({ value: item.id, label: item.level ? `${item.name}（${item.level}）` : item.name }))} disabled={!publishSessionReady || busy} />
      {!manifest ? (() => {
        const disabledReasons = [];
        if (!targetBooks.length) disabledReasons.push('本次没有可上传的小说');
        if (!publishSessionReady) disabledReasons.push('121 后台账号尚未登录或未验证');
        if (!uploadMediaReady) disabledReasons.push('存在尚未准备好的上传主视频');
        if (missingPublishMapping.length > 0) disabledReasons.push('存在发布信息没填完整的小说');
        if (missingDecompression.length > 0) disabledReasons.push('存在解压码数量不足的小说');
        if (requiresReupload) disabledReasons.push('所选小说已上传过，请从该书“查看资料”里选择重新上传');
        if (!organizationID) disabledReasons.push('请先选择 121 组织归属');
        return (
          <Tooltip title={disabledReasons.length ? `暂时不能生成：${disabledReasons.join('；')}` : ''}>
            <span>
              <Button type="primary" onClick={prepareUpload} loading={busy} disabled={disabledReasons.length > 0}>生成上传清单</Button>
            </span>
          </Tooltip>
        );
      })() : null}
      {manifest ? <><Descriptions size="small" column={1} bordered items={manifest.map(item => ({ key: item.id, label: item.title, children: <span>{item.txt} + {item.mp4}（{item.videoType}，解压 {item.jieyaNum}）</span> }))} /><Alert type="warning" showIcon message={reupload ? '确认后将重新提交到 121' : '确认后将直接提交到 121'} description="提交到接口只表示对方已接收；本页会显示等待 121 后台列表回读，不能替代后台完成状态。" /><Button danger type="primary" onClick={submitTo121} loading={busy}>{reupload ? '确认重新上传到 121' : '确认并上传到 121'}</Button></> : null}
      {liveUploadBook ? <Alert type="info" showIcon message={`${liveUploadBook.title || liveUploadBook.bookId}：${liveUploadProgress?.message || '正在创建上传任务'}`} description={liveUploadProgress?.phase ? `当前阶段：${liveUploadProgress.phase}` : '正在等待服务端返回当前阶段。'} /> : null}
      {results.map(item => <Alert key={item.id} type={item.pending ? 'info' : item.ok ? 'info' : 'error'} showIcon message={`${item.title}：${item.pending ? item.message : item.ok ? (item.result?.status === 'confirmed' ? '已提交，121 已回读' : '已提交，等待 121 回读') : '提交失败'}`} description={item.pending ? '上传过程中的实际阶段会写入当前书状态；完成或失败后自动回读。' : item.ok ? `${item.result?.sourceTextFile || ''} + ${item.result?.aiHeadVideoFile || ''}${item.result?.receipt?.remote_record?.detail ? `；${item.result.receipt.remote_record.detail}` : ''}` : item.error} />)}
    </Space>
  </Modal>;
}

export function BatchFactoryNovelList({ batch, onBack, onBatchChanged, openAutomationOnLoad = false }) {
  const [novelListOpen, setNovelListOpen] = useState(false);
  const [viewingBook, setViewingBookRaw] = useState(null);
  const [sourceEditorBook, setSourceEditorBook] = useState(null);
  const [sourceEditorValue, setSourceEditorValue] = useState('');
  const [sourceSaving, setSourceSaving] = useState(false);
  // 小说详情只展示抓取后保存的完整原文；生产草稿只能在“编辑生产内容”中查看和修改。
  const setViewingBook = book => setViewingBookRaw(book ? { ...book, workingFrontContent: '' } : null);
  const [stageSummaries, setStageSummaries] = useState({});
  const [metadataBook, setMetadataBook] = useState(null);
  const [metadataValue, setMetadataValue] = useState({});
  const [metadataSaving, setMetadataSaving] = useState(false);
  const [assetBook, setAssetBook] = useState(null);
  const [configTarget, setConfigTarget] = useState(null);
  const [editingContentBook, setEditingContentBook] = useState(null);
  const [editingContentValue, setEditingContentValue] = useState('');
  const [editingContentMode, setEditingContentMode] = useState('original');
  const [viralCandidate, setViralCandidate] = useState('');
  const [derivedOpeningOptions, setDerivedOpeningOptions] = useState([]);
  const [derivedOpeningPresetId, setDerivedOpeningPresetId] = useState('');
  const [contentSaving, setContentSaving] = useState(false);
  const [rewritingFront, setRewritingFront] = useState(false);
	const [promptBook, setPromptBook] = useState(null);
	const [promptVideoId, setPromptVideoId] = useState('');
	const [mediaBook, setMediaBook] = useState(null);
	const [mediaVideoId, setMediaVideoId] = useState('');
	const [mediaStartTab, setMediaStartTab] = useState('clips');
	const [pendingMediaPrompt, setPendingMediaPrompt] = useState(null);
	const [rowStoryboardSelection, setRowStoryboardSelection] = useState({});
  const [unifiedSettingsOpen, setUnifiedSettingsOpen] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsError, setLogsError] = useState('');
  const [automationStatus, setAutomationStatus] = useState({ state: 'idle', counts: { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 } });
  const [automationBusy, setAutomationBusy] = useState('');
  const [automationStartOpen, setAutomationStartOpen] = useState(false);
  const automationAutoOpenedRef = useRef(false);
  const [automationPresets, setAutomationPresets] = useState([]);
  const [automationPresetID, setAutomationPresetID] = useState('');
  const [automationRunMode, setAutomationRunMode] = useState('full_submit');
  const [automationScheduledAt, setAutomationScheduledAt] = useState('');
  const [automationConcurrency, setAutomationConcurrency] = useState(2);
  const [contentRangeLines, setContentRangeLines] = useState(5);
  const [contentCaptureCharacters, setContentCaptureCharacters] = useState(4000);
  const [contentSettingsSaving, setContentSettingsSaving] = useState(false);
  const [productionStatus, setProductionStatus] = useState(null);
	const [activeProductionRequestID, setActiveProductionRequestID] = useState('');
  const [mergeStatus, setMergeStatus] = useState(null);
  const [mergeSpeed, setMergeSpeed] = useState(1);
  const [capabilities, setCapabilities] = useState({});
  const [configVersions, setConfigVersions] = useState([]);
  const [configVersionsError, setConfigVersionsError] = useState(null);
  const [actionBusy, setActionBusy] = useState('');
  const [uploadMode, setUploadMode] = useState('');
  const [selectedBookIds, setSelectedBookIds] = useState([]);
  const [platformNames, setPlatformNames] = useState({});
  const [platformOptions, setPlatformOptions] = useState([]);
  const [columnWidths, setColumnWidths] = useState(null);
  const resizeActiveRef = useRef(false);
  const resizeStateRef = useRef(null);
  const runtimeStatusRequestRef = useRef(null);
  const books = Array.isArray(batch?.books) ? batch.books : [];
  useEffect(() => {
    if (!mediaBook) return;
    const refreshed = books.find(book => book.id === mediaBook.id);
    if (refreshed && refreshed.revision !== mediaBook.revision) setMediaBook(refreshed);
  }, [books, mediaBook]);
	useEffect(() => {
		if (!pendingMediaPrompt || mediaBook) return;
		const nextBook = books.find(book => book.id === pendingMediaPrompt.bookId);
		if (nextBook) {
			setPromptVideoId(pendingMediaPrompt.frameKey || '');
			setPromptBook(nextBook);
		}
		setPendingMediaPrompt(null);
	}, [books, mediaBook, pendingMediaPrompt]);
	const batchProgress = batchFactoryBatchProgress(books, { productionStatus, mergeStatus, stageSummaries });
	const progressSegments = [
		{ key: 'uploaded', label: '上传成功', count: batchProgress.uploaded },
		{ key: 'failed', label: '异常', count: batchProgress.failed },
		{ key: 'running', label: '执行中', count: batchProgress.running },
		{ key: 'awaiting-upload', label: '待上传', count: batchProgress.awaitingUpload },
		{ key: 'pending', label: '待开始', count: batchProgress.pending }
	].filter(segment => segment.count > 0).map(segment => ({ ...segment, percent: batchProgress.total ? (segment.count / batchProgress.total) * 100 : 0 }));
	const mediaVersionsByVideo = useMemo(() => completedMediaVersions(productionStatus), [productionStatus]);
  const mediaCounts = runtimeBatchMediaCounts(books, productionStatus, mergeStatus);
  const runCapability = capability(capabilities, 'director.run');
  const workingFrontCapability = capability(capabilities, 'working-front.viral');
  const compilerCapability = capability(capabilities, 'compiler.preview');
  const productionCapability = capability(capabilities, 'production.submit');
	const cancelCapability = capability(capabilities, 'production.cancel');
  const mergeCapability = capability(capabilities, 'merge.run');
  const automationState = String(automationStatus?.state || 'idle');
  const automationCounts = automationStatus?.counts || { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 };
  const automationActive = automationState === 'running' || automationState === 'scheduled';

  async function refreshBatch() {
    await onBatchChanged?.();
  }
  async function addBooksToBatch(items, platform, allowDuplicate) {
    if (!batch?.id) throw new Error('当前批次尚未加载完成，请刷新后重试');
    if (!items?.length) throw new Error('没有可添加的书籍');
    const option = platformOptions.find(item => String(item.value) === String(platform));
    const booksInput = items.map(item => ({
      id: item.bookId,
      bookId: item.bookId,
      title: item.title,
      platform: String(platform),
      sourceMetadata: { platformId: String(platform), platformName: option?.label || String(platform), sourceMode: 'manual_original', contentRangeLines: Number(contentRangeLines || 5), contentCaptureCharacters: Number(contentCaptureCharacters || 4000), queueStatus: 'manual_pending' }
    }));
    const beforeBookIDs = new Set(books.map(book => String(book?.id || '')));
    const intakeResult = await createNovelFetchIntake({ books: booksInput, metadata: { sourceMode: 'manual_original', platformId: String(platform), platformName: option?.label || String(platform), contentRangeLines: Number(contentRangeLines || 5), contentCaptureCharacters: Number(contentCaptureCharacters || 4000) } });
    const intake = resultData(intakeResult, 'intake');
    if (!intake?.id) throw new Error('新增书籍暂存失败');
    const appendedResult = await appendNovelFetchIntake(batch.id, intake.id, { allowDuplicate: allowDuplicate === true });
    const appendedBatch = resultData(appendedResult, 'batch') || resultData(await getBatch(batch.id), 'batch');
    const newlyAddedBooks = (appendedBatch?.books || []).filter(book => !beforeBookIDs.has(String(book?.id || '')) && booksInput.some(item => String(item.bookId) === String(book?.bookId)));
    const fetchFailures = [];
    for (const book of newlyAddedBooks) {
      try {
        await fetchBookOriginal(batch.id, book.id, { maxTxt: Number(contentCaptureCharacters || 4000), platform: String(platform) });
      } catch (error) {
        fetchFailures.push(`${book.title || book.bookId}: ${error?.message || '获取失败'}`);
      }
    }
    await refreshBatch();
    if (fetchFailures.length) {
      message.warning(`已添加 ${newlyAddedBooks.length} 本书；${fetchFailures.length} 本正文或分类获取失败，可在书籍详情中重试。`);
    } else {
      message.success(`已添加并获取 ${newlyAddedBooks.length} 本书的正文与分类信息`);
    }
  }
  async function deleteBook(book) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`delete-${book.id}`);
    try {
      await deleteBatchFactoryBook(batch.id, book.id);
      setViewingBook(null);
      setSelectedBookIds(current => current.filter(id => id !== book.id));
      await refreshBatch();
      message.success('已删除当前小说及其本地生产记录；121 内容未受影响');
    } catch (error) { message.error(error?.message || '删除小说失败'); }
    finally { setActionBusy(''); }
  }
  async function fetchMissingBookSource(book) {
    if (!batch?.id || !book?.id || actionBusy || String(book?.sourceText || '').trim()) return;
    setActionBusy(`source-${book.id}`);
    try {
      await fetchBookOriginal(batch.id, book.id);
      await refreshBatch();
      message.success('已获取并写入当前小说正文');
    } catch (error) {
      message.error(error?.message || '获取正文失败');
    } finally {
      setActionBusy('');
    }
  }
  async function ensureBookAudioDuration(book, { force = false, quiet = false, batchSnapshot = batch } = {}) {
    const currentBatch = batchSnapshot || batch;
    const settings = effectiveBookSettings(currentBatch, book);
    if (!audioPlanningRequired(settings)) return Number(settings.audioDurationSeconds || 0);
    const input = batchFactoryBookAudioInput(book);
    const tts = await batchFactoryBookTts(settings);
    const fingerprint = batchFactoryAudioFingerprint(input, tts);
    const currentSeconds = Number(settings.audioDurationSeconds || 0);
    if (!force && currentSeconds > 0 && (settings.audioDurationManual === true || String(settings.audioDurationFingerprint || '') === fingerprint)) return currentSeconds;
    if (!quiet) message.info(`正在为《${book.title || book.bookId || '当前小说'}》读取真实配音时长…`);
    try {
      const measured = await generateBatchFactoryBookAudioMeasurement(book, settings);
      await saveBookOverrideWithRetry(currentBatch.id, book.id, Number(book.revision || 0), { patch: { audioDurationSeconds: measured.durationSeconds, audioDurationFingerprint: measured.fingerprint, audioLineDurations: measured.lineDurations, audioDurationManual: false } });
      return measured.durationSeconds;
    } catch (error) {
      if (currentSeconds > 0) {
        if (!quiet) message.warning(`配音时长重新读取失败，继续使用已保存的配音时长 ${currentSeconds.toFixed(2)} 秒。`);
        return currentSeconds;
      }
      throw error;
    }
  }

  async function prepareAudioPlanningForBatch(currentBatch) {
    const candidates = (currentBatch?.books || []).filter(book => {
      const settings = effectiveBookSettings(currentBatch, book);
      return audioPlanningRequired(settings);
    });
    if (!candidates.length) return;
    message.info(`正在逐本生成配音并读取 ${candidates.length} 本小说的真实时长…`);
    const failed = [];
    for (const book of candidates) {
      try {
        await ensureBookAudioDuration(book, { force: true, quiet: true, batchSnapshot: currentBatch });
      } catch (error) {
        failed.push(`${book.title || book.bookId || '当前小说'}：${error?.message || '配音时长读取失败'}`);
      }
    }
    if (failed.length) {
      message.warning(`已完成 ${candidates.length - failed.length}/${candidates.length} 本配音时长读取；${failed.join('；')}`);
      return;
    }
    message.success(`已生成并读取 ${candidates.length} 本小说的真实配音时长。`);
  }

  async function syncUnifiedSettingsToBooks(currentBatch) {
    const candidates = currentBatch?.books || [];
    const results = await Promise.allSettled(candidates.map(async book => {
      const bookPatch = book?.settingsState?.patch || {};
      const restoreKeys = UNIFIED_CONFIGURATION_KEYS.filter(key => Object.hasOwn(bookPatch, key));
      if (!restoreKeys.length) return;
      await saveBookOverrideWithRetry(currentBatch.id, book.id, Number(book.revision || 0), { patch: {}, restoreKeys: UNIFIED_CONFIGURATION_KEYS });
    }));
    const failed = results.flatMap((result, index) => result.status === 'rejected'
      ? [`${candidates[index]?.title || candidates[index]?.bookId || '当前小说'}：${result.reason?.message || '同步失败'}`]
      : []);
    if (failed.length) throw new Error(`统一配置已保存，但 ${failed.length} 本小说未完成同步：${failed.join('；')}`);
  }
	async function recompileUnifiedH3Prompts(currentBatch) {
		const candidates = (currentBatch?.books || []).filter(book => h3DirectorCards(book).length > 0);
		const result = { compiled: 0, manual: 0, waitingForAudio: 0, failed: [] };
		for (const book of candidates) {
			try {
				const outcome = await compileBookH3Videos(book, { interactive: false, allowAudioSynthesis: true });
				if (outcome?.skipped === 'manual') result.manual += 1;
				else if (outcome?.skipped === 'audio') result.waitingForAudio += 1;
				else if (outcome?.compiled) result.compiled += 1;
			} catch (error) {
				result.failed.push(`${book.title || book.bookId || '当前小说'}：${error?.message || '重新编译失败'}`);
			}
		}
		return result;
	}
	async function compileBookH3Videos(book, { interactive = true, allowAudioSynthesis = true } = {}) {
		const latestResponse = await getBatch(batch.id);
		const latestBatch = latestResponse?.batch || latestResponse;
		const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
		const director = latestBook?.directorRevision;
		const h3Document = director?.output?.h3_director || director?.output?.h3Director;
		if (!director?.id || !h3Document) return;
		const settings = effectiveBookSettings(latestBatch, latestBook);
		let priorTrace = null;
		try { priorTrace = await getH3Trace(latestBatch.id, latestBook.id); }
		catch (error) { if (Number(error?.status) !== 404 && !String(error?.message || '').includes('404')) throw error; }
		const hasManualPrompts = (priorTrace?.compilation?.compilation?.segments || []).some(segment => segment.compile_trace?.editable_copy_source === 'user_final_prompt');
		if (hasManualPrompts) {
			if (!interactive) return { skipped: 'manual' };
			const confirmed = await new Promise(resolve => Modal.confirm({ title: '重新编译会替换手动编辑的分镜提示词', content: '历史提交记录保留。是否按当前配置重新生成完整提示词？', okText: '确认重新编译', cancelText: '保留手动提示词', onOk: () => resolve(true), onCancel: () => resolve(false) }));
			if (!confirmed) throw new Error('已保留手动提示词，当前配置尚未应用到最终分镜');
		}
		let audioResult = null;
    if (audioPlanningRequired(settings)) {
			const tts = await batchFactoryBookTts(settings);
			const previousMeasurement = priorTrace?.timeline?.timeline?.audio_measurement;
			if (!allowAudioSynthesis && !previousMeasurement?.asset_id) return { skipped: 'audio' };
			audioResult = await measureH3VideoLines({ directorId: director.id, document: h3Document, tts,
				previous: previousMeasurement,
				synthesize: textToSpeech, encode: blobToBase64,
				measure: payload => measureH3Audio(latestBatch.id, latestBook.id, payload) });
		}
		const promptConfig = settings?.aiPromptConfig || {};
		const videoPreset = promptConfig.video || {};
		const videoPresetBody = String(videoPreset.body || '');
		const videoPromptTemplate = finalVideoPromptTemplate(videoPresetBody);
		const constraints = promptConfig.constraints || {};
		const h3Constraints = buildBatchFactoryH3Constraints(constraints);
		const maxSegmentSeconds = Number(settings.storyboardDurationLimit) === 15 ? 15 : 10;
		await compileH3Video(latestBatch.id, latestBook.id, {
			director_revision_id: director.id,
			audio_asset_id: audioResult?.audio_asset_id || audioResult?.audioMeasurement?.measurement?.asset_id || audioResult?.audio_measurement?.measurement?.asset_id,
			allow_semantic_timeline: !audioPlanningRequired(settings),
			preset: {
				key: String(videoPreset.presetKey || videoPreset.presetId || videoPreset.id || 'h3-video-normal'),
				revision: Math.max(1, Number(videoPreset.presetVersion || videoPreset.version || 1)),
				format: 'h3-structured-v1',
				max_segment_ms: maxSegmentSeconds * 1000,
				request_duration_mode: 'ceil-second',
				prompt_template: videoPromptTemplate,
				output_constraints: videoPromptTemplate ? '按冻结导演数据和所选最终模板生成当前 VIDEO 提示词。' : (videoPresetBody || '按结构化时间线输出当前 VIDEO，保持人物、动作、机位和场景连续性。')
			},
			...h3Constraints,
			editable_copy_overrides: {},
			expected_compilation_id: priorTrace?.compilation?.id || '',
			allow_replace_manual_prompts: hasManualPrompts
		});
		return { compiled: true };
	}
	async function refreshAfterBookSettingsSaved() {
		const shouldRecompileH3 = ['constraints', 'video', 'media'].includes(configTarget?.region)
			&& Boolean(configTarget?.book?.directorRevision?.output?.h3_director || configTarget?.book?.directorRevision?.output?.h3Director);
		if (shouldRecompileH3) {
			try {
				message.info('设置已保存，正在复用现有导演数据重新编译最终 VIDEO Prompt…');
				await runBookStageAction(configTarget.book, 'director', 'compile');
				message.success('设置已生效，最终 VIDEO Prompt 已重新编译。');
			} catch (error) {
				message.error(error?.message ? `设置已保存，但 H3 重编译失败：${error.message}` : '设置已保存，但 H3 重编译失败');
			}
		}
		await refreshBatch();
	}
  async function resetBookAudioMeasurement(book, expectedRevision = book?.revision) {
    return saveBookOverrideWithRetry(batch.id, book.id, Number(expectedRevision || 0), { patch: { audioDurationSeconds: 0, audioDurationFingerprint: '', audioDurationManual: false } });
  }
  async function loadRuntimeStatus({ quiet = false, runtimeCapabilities = capabilities } = {}) {
    if (!batch?.id) return;
    // 自动化轮询、生产抽屉和首次加载可能在同一时刻触发；复用同一个请求，
    // 让浏览器也始终只有一条批量运行状态读取在飞行。
    if (runtimeStatusRequestRef.current) return runtimeStatusRequestRef.current;
    setLogsLoading(true);
    const request = (async () => {
      try {
        // 一个批量摘要替代“状态 + 合并 + 每本阶段”的浏览器 N+2 次请求。
        // 服务端合并读取并短缓存，页面只消费同一时刻的快照。
        const runtime = await getBatchRuntimeSummary(batch.id, { suppressGlobalError: true });
        setProductionStatus(runtime?.production || { jobs: [] });
        setMergeStatus(runtime?.merge || { jobs: [] });
        setStageSummaries(runtime?.stageSummaries || {});
        const nextAutomation = runtime?.automation || { state: 'idle', counts: { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 } };
        setAutomationStatus(nextAutomation);
        setLogsError('');
        return runtime;
      } catch (error) {
        const endpoint = String(error?.source || '').trim();
        const status = Number.isInteger(error?.status) ? `HTTP ${error.status}` : '';
        const text = [error?.message || '读取任务状态失败', status, endpoint ? `状态接口：${endpoint}` : ''].filter(Boolean).join(' · ');
        setLogsError(text);
        if (!quiet) message.error(text);
        return undefined;
      } finally {
        setLogsLoading(false);
      }
    })();
    runtimeStatusRequestRef.current = request;
    try { return await request; }
    finally {
      if (runtimeStatusRequestRef.current === request) runtimeStatusRequestRef.current = null;
    }
  }
  useEffect(() => {
    if (!batch?.id) return;
    let active = true;
    Promise.all([getCapabilities(), getConfigVersions()]).then(async ([nextCapabilities, nextVersions]) => {
      if (!active) return;
      setCapabilities(nextCapabilities || {});
      setConfigVersions(resultData(nextVersions, 'configVersions') || []);
      setConfigVersionsError(null);
      await loadRuntimeStatus({ quiet: true, runtimeCapabilities: nextCapabilities || {} });
    }).catch(error => { if (active) setConfigVersionsError(error); });
    return () => { active = false; };
  }, [batch?.id]);
  useEffect(() => {
    if (!batch?.id) return undefined;
    let active = true;
    let polling = false;
    let wasRunning = false;
    const poll = async () => {
      if (polling) return;
      polling = true;
      try {
        const runtime = await loadRuntimeStatus({ quiet: true });
        if (!active) return;
        const next = runtime?.automation;
        setAutomationStatus(next || { state: 'idle', counts: { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 } });
        if (next?.state === 'running') {
          wasRunning = true;
          await onBatchChanged?.();
        } else if (wasRunning) {
          // 任务刚从“执行中”结束，补刷最后一次，拿到最终书籍状态，
          // 避免页面停在结束前的旧画面（如 1 成功 5 异常）。
          wasRunning = false;
          await onBatchChanged?.();
        }
      } catch (error) {
        if (active && Number(error?.status || 0) !== 404) console.error('[batch-factory-automation] status read failed', error);
      } finally { polling = false; }
    };
    void poll();
    // 页面切到后台（最小化/切标签）时暂停轮询，回前台立刻补刷一次再继续，
    // 避免多人多标签把小服务器刷爆。页面一打开就在后台则等回前台再启动。
    let timer = document.hidden ? null : setInterval(poll, 15000);
    const onVisibilityChange = () => {
      if (document.hidden) {
        if (timer) { clearInterval(timer); timer = null; }
      } else if (timer === null) {
        void poll();
        timer = setInterval(poll, 15000);
      }
    };
    document.addEventListener('visibilitychange', onVisibilityChange);
    return () => {
      active = false;
      if (timer) clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisibilityChange);
    };
  }, [batch?.id]);
  useEffect(() => {
    let active = true;
    getWorkshopPlatforms({ suppressGlobalError: true }).then(result => {
      if (!active) return;
      const options = batchFactoryPlatformOptions(result?.platforms);
      setPlatformOptions(options);
      setPlatformNames(Object.fromEntries(options.map(option => [String(option.value), option.label])));
    }).catch(() => { if (active) setPlatformNames({}); });
    return () => { active = false; };
  }, []);
  useEffect(() => { setSelectedBookIds(current => current.filter(id => books.some(book => book.id === id))); }, [batch?.id, books.length]);
  useEffect(() => {
    if (!assetBook) return;
    const refreshed = books.find(book => book.id === assetBook.id);
    if (refreshed && refreshed !== assetBook) setAssetBook(refreshed);
  }, [batch?.id, assetBook?.id, books]);
  useEffect(() => {
    if (!promptBook) return;
    const refreshed = books.find(book => book.id === promptBook.id);
    if (refreshed && refreshed !== promptBook) setPromptBook(refreshed);
  }, [batch?.id, promptBook?.id, books]);
  useEffect(() => {
    if (!configTarget?.book) return;
    const refreshed = books.find(book => book.id === configTarget.book.id);
    if (refreshed && refreshed !== configTarget.book) setConfigTarget(current => current ? { ...current, book: refreshed } : current);
  }, [batch?.id, configTarget?.book?.id, books]);
  useEffect(() => {
    if (!editingContentBook) return;
    const refreshed = books.find(book => book.id === editingContentBook.id);
    if (!refreshed || refreshed === editingContentBook) return;
    setEditingContentBook(refreshed);
    if (!String(editingContentValue || '').trim()) setEditingContentValue(runtimeResolveBookProductionText(refreshed));
  }, [batch?.id, books, editingContentBook?.id]);
  const hasActiveProduction = (productionStatus?.jobs || []).some(job => (job?.tasks || []).some(task => ['queued', 'running'].includes(String(task?.status || '').toLowerCase())));
  useEffect(() => {
    if (!batch?.id || !(viewingBook || mediaBook || promptBook || hasActiveProduction)) return undefined;
    const delay = hasActiveProduction ? 15000 : 30000;
    const tick = () => { loadRuntimeStatus({ quiet: true }); };
    // 同自动化轮询：切后台暂停，回前台补刷。
    let timer = document.hidden ? null : setInterval(tick, delay);
    const onVisibilityChange = () => {
      if (document.hidden) {
        if (timer) { clearInterval(timer); timer = null; }
      } else if (timer === null) {
        tick();
        timer = setInterval(tick, delay);
      }
    };
    document.addEventListener('visibilitychange', onVisibilityChange);
    return () => {
      if (timer) clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisibilityChange);
    };
  }, [batch?.id, viewingBook?.id, mediaBook?.id, promptBook?.id, hasActiveProduction]);

  async function openContentEditor(book) {
    const currentBook = books.find(item => item?.id === book?.id) || book;
    setEditingContentBook(currentBook);
    setEditingContentValue(runtimeResolveBookProductionText(currentBook));
    setEditingContentMode(batchFactoryContentMode(currentBook));
    setViralCandidate('');
    setDerivedOpeningOptions([]);
    setDerivedOpeningPresetId('');
    const [draftRequest, catalogRequest] = await Promise.allSettled([
      getDraft({ key: `working-front-candidate:${currentBook.id}`, kind: 'working-front-viral-candidate', scope: batch.id }, { suppressGlobalError: true }),
      listSystemPresetCatalog('batch-factory')
    ]);
    if (draftRequest.status === 'fulfilled') {
      setViralCandidate(String(resultData(draftRequest.value, 'draft')?.content || ''));
    } else if (draftRequest.reason?.status && draftRequest.reason.status !== 404) {
      message.error(draftRequest.reason?.message || '读取爆款候选失败');
    }
    if (catalogRequest.status !== 'fulfilled') {
      message.error(catalogRequest.reason?.message || '读取衍生开篇提示词失败');
      return;
    }
    const allowedSlots = new Set(['batch.hook-adaptation', 'batch.original-director', 'batch.viral-director']);
    const options = (Array.isArray(catalogRequest.value?.catalog) ? catalogRequest.value.catalog : [])
      .filter(item => allowedSlots.has(String(item?.slot || '')))
      .map(item => ({ value: item.id, label: `${item.name || item.id} · v${item.version || 1}` }));
    setDerivedOpeningOptions(options);
    const inheritedId = String(batch?.settingsState?.patch?.aiPromptConfig?.derivedOpening?.presetId || '');
    setDerivedOpeningPresetId(options.some(option => option.value === inheritedId) ? inheritedId : (options[0]?.value || ''));
  }

  function openSourceEditor(book) {
    const currentBook = books.find(item => item?.id === book?.id) || book;
    setSourceEditorBook(currentBook);
    setSourceEditorValue(batchFactoryCleanSourceText(currentBook.sourceText));
  }

  async function saveCompleteSource(book, nextSourceText) {
    if (!book || !batch?.id) return;
    const sourceText = sourceTextAfterViralAdoption(nextSourceText);
    if (!sourceText) { message.error('完整原文不能为空'); return; }
    setSourceSaving(true);
    try {
      const save = revision => updateBookSource(batch.id, book.id, { sourceText, replaceSource: true, expectedRevision: Number(revision || 0) });
      try {
        await save(book.revision);
      } catch (error) {
        if (Number(error?.status) !== 409) throw error;
        const latestResult = await getBatch(batch.id);
        const latestBatch = resultData(latestResult, 'batch');
        const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
        if (!latestBook) throw error;
        await save(latestBook.revision);
      }
      // The accepted complete source is authoritative. Remove a previous
      // first-lines production draft so it cannot override this book's new
      // source during subsequent production or a rewrite-enabled upload.
      await saveDraft({ key: `working-front:${book.id}`, kind: 'working-front-content', scope: batch.id, content: '' });
      await refreshBatch();
      setViewingBook({ ...book, sourceText });
      setSourceEditorBook(null);
      setEditingContentBook(null);
      setViralCandidate('');
      message.success('完整原文已保存；下游内容将按新原文重新生成');
    } catch (error) {
      message.error(error?.message || '保存完整原文失败');
    } finally {
      setSourceSaving(false);
    }
  }

  async function saveSourceEditor(nextSourceText = sourceEditorValue) {
    return saveCompleteSource(sourceEditorBook, nextSourceText);
  }

  async function adoptViralCandidate() {
    return saveCompleteSource(editingContentBook || sourceEditorBook, viralCandidate);
  }
  async function saveWorkingContent() {
    if (!editingContentBook) return;
    setContentSaving(true);
    try {
      await saveDraft({ key: `working-front:${editingContentBook.id}`, kind: 'working-front-content', scope: batch.id, content: editingContentValue.trim() });
      const metadataResult = await updateBookMetadata(batch.id, editingContentBook.id, { metadata: { ...(editingContentBook.sourceMetadata || {}), productionContentMode: editingContentMode }, expectedRevision: Number(editingContentBook.revision || 0) });
      const updatedBook = resultData(metadataResult, 'book') || metadataResult;
      await resetBookAudioMeasurement(editingContentBook, Number(updatedBook?.revision || editingContentBook.revision || 0));
      await refreshBatch();
      setEditingContentBook(null);
      message.success('当前小说的生产内容已保存；原文不会被覆盖。');
    } catch (error) { message.error(error?.message || '保存生产内容失败'); } finally { setContentSaving(false); }
  }
  async function createViralCandidate() {
    const targetBook = editingContentBook || sourceEditorBook;
    const sourceText = editingContentBook ? editingContentValue : sourceEditorValue;
    if (!targetBook || rewritingFront) return;
    setRewritingFront(true);
    try {
      const result = await rewriteWorkingFront(batch.id, targetBook.id, sourceText, {
        textModelId: effectiveBookSettings(batch, targetBook).textModelId,
        promptPresetId: derivedOpeningPresetId
      });
      const candidate = String(resultData(result, 'candidate') || '').trim();
      setViralCandidate(candidate);
      message.success('爆款候选已生成；录用后会替换完整正文，工作区仍只显示所选行数。');
    } catch (error) { message.error(error?.message || '生成爆款候选失败'); } finally { setRewritingFront(false); }
  }
  async function cancelViralCandidate() {
    const targetBook = editingContentBook || sourceEditorBook;
    if (!targetBook) return;
    try {
      await saveDraft({ key: `working-front-candidate:${targetBook.id}`, kind: 'working-front-viral-candidate', scope: batch.id, content: '' });
      setViralCandidate('');
      message.info('已取消本次爆款候选，完整正文未改动。');
    } catch (error) { message.error(error?.message || '取消爆款候选失败'); }
  }
  async function refreshAssetPresetSnapshot(book) {
    const aiPromptConfig = book?.settingsState?.patch?.aiPromptConfig;
    if (!aiPromptConfig || typeof aiPromptConfig !== 'object') return;
    // Older local builds persisted only preset metadata. Re-save through the
    // current Node bridge so the selected published body is frozen before the
    // asset-only stage asks the text provider to extract anything.
    await saveBookOverride(batch.id, book.id, {
      patch: { aiPromptConfig },
      expectedRevision: Number(book.revision || 0)
    });
    await refreshBatch();
  }
  async function runBookStageAction(book, stage, mode = 'missing', videoId = '', textModelId = '') {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(stageActionKey(stage, book.id));
    try {
      const settings = effectiveBookSettings(batch, book);
      const h3VideoPrompt = usesH3VideoPrompt(settings);
      if (stage === 'director' && mode === 'compile' && h3VideoPrompt && h3DirectorCards(book).length) {
        await compileBookH3Videos(book);
        await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
        message.success('已复用统一导演卡编译最终 VIDEO Prompt。');
        return;
      }
      const requestedMode = stage === 'director' && mode === 'compile' ? 'missing' : mode;
      if (stage === 'director' && audioPlanningRequired(settings)) {
        await ensureBookAudioDuration(book);
      }
      const stageResult = await runBookStage(batch.id, book.id, stage, {
        mode: requestedMode,
        // The selected video prompt preset chooses the director contract:
        // H3 uses the structured H3 document, while SD uses the SD text
        // contract. The video model remains an execution-only setting.
        h3: h3VideoPrompt,
        videoId,
        textModelId: textModelId || settings.textModelId,
        ...(stage === 'video' ? {
          provider: videoProviderForModel(settings.videoModelId, settings.videoProvider),
          videoModelId: settings.videoModelId
        } : {}),
        requestId: requestID(`bf11-${stage}-${mode}`)
      });
      if (stage === 'director' && h3VideoPrompt) await compileBookH3Videos(book);
      // The modal keeps its own book snapshot while it is open.  Refresh the
      // exact book immediately after a director run so a newly-created
      // director revision (and its rebuilt H3 cards) is visible without
      // closing/reopening the modal or relying on a parent render race.
      if (stage === 'director') {
        const latestResponse = await getBatch(batch.id);
        const latestBatch = latestResponse?.batch || latestResponse;
        const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
        if (latestBook) setPromptBook(latestBook);
      }
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      const labels = { assets: mode === 'force' ? '已重新生成资产；分镜和 VIDEO 未改动' : '已生成资产', director: mode === 'force' ? '已重新生成文案' : '已生成文案', image: mode === 'force' ? '已重新生成图片' : '已生成图片', video: mode === 'force' ? '已创建新的视频候选版本' : '已提交视频生成' };
      const baseLabel = labels[stage] || '当前小说阶段已提交。';
      if (stage === 'director' && stageResult?.openingVariants?.triggered) {
        const ov = stageResult.openingVariants;
        if (ov.succeeded) message.success(`${baseLabel}；已自动生成 ${ov.generated} 个换开头变体，可在分镜一点击“查看换开头提示词”。`);
        else message.warning(`${baseLabel}；但换开头变体未生成：${ov.reason || '可重试失败步骤'}`);
      } else {
        message.success(baseLabel);
      }
    } catch (error) {
      // A provider can reject a request synchronously (quota, moderation, or
      // invalid input).  Reload the persisted stage run as well as the book so
      // the error is visible in the card and task timeline immediately.
      // End the visible action before any best-effort readback.  A slow status
      // endpoint must never leave a rejected upstream request spinning.
      setActionBusy('');
      message.error(error?.message || '当前小说阶段执行失败');
      void Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]).catch(() => {});
    } finally { setActionBusy(''); }
  }
  async function retryLastFailedStage(book, videoId = '', textModelId = '') {
    if (!batch?.id || !book?.id || actionBusy) return;
    // A running server job is the single retry owner.  Letting this modal
    // create a second request races that job and, worse, makes a provider
    // failure look like the client simply found no retryable stage.
    if (automationActive) {
      message.info('服务端自动生产正在处理该小说；会自动重试，无需手动操作。');
      return;
    }
    setActionBusy(stageActionKey('retry', book.id));
    try {
      // The runtime summary is a display cache. Historical provider failures
      // can exist before it has indexed the corresponding stage row, so the
      // retry endpoint remains the source of truth for eligibility.
      const runtime = await loadRuntimeStatus({ quiet: true });
      const summary = runtime?.stageSummaries?.[book.id] || stageSummaries[book.id];
      const videoProgress = batchFactoryVideoProgress(book, runtime?.production || productionStatus);
      const effectiveVideoID = String(videoId || [...videoProgress.byVideo.entries()].find(([, item]) => item?.status === 'failed')?.[0] || '');
      const failedVideo = effectiveVideoID ? videoProgress.byVideo.get(effectiveVideoID) : null;
      // A video can fail only after its initial stage submission has succeeded.
      // In that case there is no failed stage-run row, but the persisted task is
      // still an explicit retry target for this storyboard.
      const retryAsVideo = failedVideo?.status === 'failed';
      const settings = effectiveBookSettings(batch, book);
      if (summary?.lastFailed?.stage === 'director' && audioPlanningRequired(settings)) await ensureBookAudioDuration(book);
      await retryBookStage(batch.id, book.id, {
        h3: usesH3VideoPrompt(settings),
        videoId: effectiveVideoID,
        textModelId: textModelId || settings.textModelId,
        ...(summary?.lastFailed?.stage === 'video' || retryAsVideo || !summary?.lastFailed ? {
          provider: videoProviderForModel(settings.videoModelId, settings.videoProvider),
          videoModelId: settings.videoModelId
        } : {}),
        requestId: requestID('bf11-book-retry')
      });
		if (summary?.lastFailed?.stage === 'director' && usesH3VideoPrompt(settings)) await compileBookH3Videos(book);
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      message.success('已识别并重跑该书最后失败的步骤。');
    } catch (error) {
      setActionBusy('');
      message.error(error?.message || '当前小说没有可重试的失败步骤');
      void Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]).catch(() => {});
    } finally { setActionBusy(''); }
  }
  async function refreshBookSmartUnifiedAction(book, { quiet = false } = {}) {
    if (!batch?.id || !book?.id || actionBusy) return null;
    setActionBusy(`smart-unified-${book.id}`);
    try {
      const settings = effectiveBookSettings(batch, book);
      const response = await refreshBookSmartUnified(batch.id, book.id, { textModelId: settings.textModelId });
      const smartUnified = resultData(response, 'smartUnified') || response?.smartUnified || {};
      await refreshBatch();
      if (smartUnified.available) {
        const latest = resultData(await getBatch(batch.id), 'batch');
        const refreshedBook = (latest?.books || []).find(item => item?.id === book.id);
        if (refreshedBook && h3DirectorCards(refreshedBook).length) await compileBookH3Videos(refreshedBook, { interactive: false });
        await refreshBatch();
      }
      if (!quiet) {
        if (smartUnified.available) message.success('智能统一视觉基线已刷新，最终 VIDEO Prompt 已按当前导演卡重新编译。');
        else message.info(smartUnified.reason || '智能统一未获取，已保留原流程和既有导演卡。');
      }
      return smartUnified;
    } catch (error) {
      if (!quiet) message.error(error?.message || '刷新智能统一失败');
      return null;
    } finally { setActionBusy(''); }
  }
  async function refreshAllSmartUnified() {
    const targets = books.filter(book => Boolean(book?.id));
    if (!targets.length || actionBusy) return;
    let succeeded = 0;
    let unavailable = 0;
    for (const book of targets) {
      const result = await refreshBookSmartUnifiedAction(book, { quiet: true });
      if (result?.available) succeeded += 1;
      else unavailable += 1;
    }
    message.info(`已逐本刷新智能统一：${succeeded} 本获取成功，${unavailable} 本未获取但未阻断任何生产步骤。`);
  }
  async function runAi(scope, textModelId = '') {
    if (!batch?.id || actionBusy) return;
    if (scope?.id) { await runBookStageAction(scope, 'director', 'missing', '', textModelId); return; }
    setActionBusy('director');
    try {
      const audioBooks = books.filter(book => { const settings = effectiveBookSettings(batch, book); return audioPlanningRequired(settings); });
      if (audioBooks.length) message.info(`正在逐本读取 ${audioBooks.length} 本小说的真实配音时长…`);
      for (const book of audioBooks) await ensureBookAudioDuration(book, { quiet: true });
      await runBatchDirector(batch.id, { textModelId: textModelId || batch?.settingsState?.patch?.textModelId });
      await refreshBatch();
      message.success('批量 AI 推理已提交并回读当前作品。');
      setAiOpen(false);
    } catch (error) { message.error(error?.message || 'AI 推理失败'); } finally { setActionBusy(''); }
  }
  async function runProduction(book = null) {
    if (!batch?.id || actionBusy) return;
    setActionBusy('production');
    try {
      const operationRequestID = requestID(book ? 'bf11-book-production' : 'bf11-batch-production');
      const settings = book ? effectiveBookSettings(batch, book) : (batch?.settingsState?.patch || {});
      const provider = videoProviderForModel(settings.videoModelId, settings.videoProvider);
      if (book) await runBookStage(batch.id, book.id, 'video', { mode: 'missing', provider, videoModelId: settings.videoModelId, requestId: operationRequestID });
      else await submitBatchProduction(batch.id, operationRequestID, provider, settings.videoModelId);
      setActiveProductionRequestID(operationRequestID);
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      message.success(book ? '当前小说的视频任务已提交。' : '批量视频任务已提交。');
    } catch (error) { message.error(error?.message || '提交视频任务失败'); } finally { setActionBusy(''); }
  }
  function mergePayload(prefix, book = null, speed = 1, overrides = {}) {
    const settings = normalizeAudioSourceSettings(book ? effectiveBookSettings(batch, book) : (batch?.settingsState?.patch || {}));
    const followAudio = normalizeAudioSource(settings.audioSource) === 'original' ? false : (Object.hasOwn(overrides, 'followAudio') ? overrides.followAudio === true : settings.audioMergeEnabled === true);
    const audioDurationSeconds = Object.hasOwn(overrides, 'audioDurationSeconds')
      ? Number(overrides.audioDurationSeconds || 0)
      : Number(settings.audioDurationSeconds || 0);
    return {
      requestId: requestID(prefix),
      timingMode: followAudio ? 'audio' : 'speed',
      speed: followAudio ? 0 : Number(overrides.speed ?? speed ?? 1),
      audioDurationSeconds: followAudio ? audioDurationSeconds : 0
    };
  }
  async function runMerge() {
    if (!batch?.id || actionBusy) return;
    setActionBusy('merge');
    try {
      await submitBatchMerge(batch.id, mergePayload('bf11-merge'));
      await loadRuntimeStatus({ quiet: true });
      message.success('批量合并任务已提交。');
    } catch (error) { message.error(error?.message || '提交批量合并失败'); } finally { setActionBusy(''); }
  }
  async function runBookMerge(book, overrides = {}) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`merge-${book.id}`);
    try {
      await submitBookMerge(batch.id, book.id, mergePayload('bf11-book-merge', book, Number(overrides.speed ?? mergeSpeed ?? 1), overrides));
      await loadRuntimeStatus({ quiet: true });
      message.success('当前小说最终合成任务已提交。');
    } catch (error) { message.error(error?.message || '提交当前小说合成失败'); } finally { setActionBusy(''); }
  }
	async function cancelProduction() {
		if (!batch?.id || actionBusy || !cancelCapability.available) return;
		if (!activeProductionRequestID) {
			message.info('请在本次页面中先发起一个批量视频操作；取消只会作用于该次操作。');
			return;
		}
		setActionBusy('cancel');
		try {
			const result = await cancelBatchProduction(batch.id, activeProductionRequestID);
			await loadRuntimeStatus({ quiet: true });
			const cancelled = resultData(result, 'cancelledTaskIds') || [];
			const skipped = resultData(result, 'skipped') || [];
			if (cancelled.length) message.success(`已取消 ${cancelled.length} 个本地执行器任务。`);
			else if (skipped.length) message.warning(`没有可取消的本地任务：${skipped.map(item => item.reason).filter(Boolean).join('；')}`);
			else message.info('当前没有运行中的视频任务。');
		} catch (error) { message.error(error?.message || '取消视频任务失败'); } finally { setActionBusy(''); }
	}
  async function loadAutomationPresets() {
    const result = await listAutomationPresets();
    const values = resultData(result, 'presets') || [];
    setAutomationPresets(values);
    return values;
  }
  async function openAutomationStart() {
    try {
      const firstBook = books[0];
      setContentRangeLines(Number(firstBook?.sourceMetadata?.contentRangeLines || 5));
      setContentCaptureCharacters(Number(firstBook?.sourceMetadata?.contentCaptureCharacters || 4000));
      const values = await loadAutomationPresets();
      setAutomationPresetID(current => current || values[0]?.id || '');
      setAutomationScheduledAt(current => current || formatBeijingDatetimeLocal(new Date(Date.now() + 10 * 60 * 1000).toISOString()));
      setAutomationStartOpen(true);
    } catch (error) { message.error(error?.message || '读取自动化预设失败'); }
  }
  async function saveContentSettings() {
    if (!batch?.id || !books.length || contentSettingsSaving) return false;
    const rangeLines = Math.min(500, Math.max(1, Number(contentRangeLines) || 5));
    const captureCharacters = Math.min(100000, Math.max(1000, Number(contentCaptureCharacters) || 4000));
    setContentSettingsSaving(true);
    try {
      for (const book of books) {
        await updateBookMetadata(batch.id, book.id, {
          metadata: { ...(book.sourceMetadata || {}), contentRangeLines: rangeLines, contentCaptureCharacters: captureCharacters },
          expectedRevision: Number(book.revision || 0)
        });
      }
      await refreshBatch();
      setContentRangeLines(rangeLines);
      setContentCaptureCharacters(captureCharacters);
      message.success(`内容设置已保存：范围 ${rangeLines} 行，截取 ${captureCharacters} 字。`);
      return true;
    } catch (error) {
      message.error(error?.message || '内容范围和截取设置保存失败');
      return false;
    } finally { setContentSettingsSaving(false); }
  }
  useEffect(() => {
    if (openAutomationOnLoad && !automationAutoOpenedRef.current) {
      automationAutoOpenedRef.current = true;
      void openAutomationStart();
    }
    // The parent only sets this flag for a newly created novel-fetch batch.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [openAutomationOnLoad]);
  async function runAutomationAction(key, { immediate = false } = {}) {
    if (!batch?.id || automationBusy) return;
    if (key === 'start' && !automationPresetID) { message.warning('请先选择自动化预设'); return; }
    const scheduledAt = immediate ? '' : parseBeijingDatetimeLocal(automationScheduledAt);
    if (key === 'start' && !immediate && (!scheduledAt || new Date(scheduledAt).getTime() <= Date.now())) {
      message.warning('请选择晚于现在的北京时间自动启动时间');
      return;
    }
    setAutomationBusy(key);
    try {
      if (key === 'start' && !(await saveContentSettings())) return;
      let result;
      if (key === 'start') result = await startBatchAutomation(batch.id, { presetId: automationPresetID, runMode: automationRunMode, scheduledAt: scheduledAt || undefined, concurrency: normalizeAutomationConcurrency(automationConcurrency) });
      else if (key === 'pause') result = await pauseBatchAutomation(batch.id);
      else if (key === 'resume') result = await resumeBatchAutomation(batch.id);
      else if (key === 'retry') result = await retryBatchAutomation(batch.id);
      else if (key === 'cancel') result = await cancelBatchAutomation(batch.id);
      const next = resultData(result, 'automation');
      setAutomationStatus(next || automationStatus);
      if (key === 'start') { setAutomationStartOpen(false); message.success(immediate ? '已立即启动自动化任务。' : '已创建北京时间定时自动化任务。'); }
      else if (key === 'pause') message.info('已暂停自动化；已提交给模型或合并器的在途任务不会被强制删除。');
      else if (key === 'resume') message.success('已继续自动化；已在途任务将继续回读，失败小说请使用“重试失败小说”。');
      else if (key === 'retry') message.success('已重置失败或阻塞小说，并从缺失阶段继续。');
      else if (key === 'cancel') message.info('已停止自动化编排；已在途的供应商任务仍会保留。');
      await Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
    } catch (error) {
      message.error(error?.message || '批量工厂自动化操作失败');
    } finally { setAutomationBusy(''); }
  }

  function openMetadataEditor(book) {
    setMetadataBook(book);
    setMetadataValue({ ...(book?.sourceMetadata || {}) });
  }
  async function classifyBookMetadata(book, force = true) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`classify-${book.id}`);
    try {
      const result = await classifyBookPublishMetadata(batch.id, book.id, {
        textModelId: effectiveBookSettings(batch, book).textModelId,
        force
      });
      const nextBook = resultData(result, 'book');
      await refreshBatch();
      if (nextBook?.id === viewingBook?.id) setViewingBook(nextBook);
      message.success(resultData(result, 'reused') ? '已复用当前书的发布分类信息。' : '已识别当前书的男女频、风格和标签。');
    } catch (error) {
      // Classification failures are persisted by the service.  Reload the
      // current book so the modal keeps the actionable reason instead of
      // leaving the user with a disappearing toast only.
      try {
        const refreshed = resultData(await getBatch(batch.id), 'batch');
        const latest = (refreshed?.books || []).find(item => String(item?.id) === String(book.id));
        if (latest) setViewingBook(latest);
      } catch (_) { /* retain the original request error */ }
      message.error(error?.message || 'AI 判断发布信息失败');
    } finally { setActionBusy(''); }
  }
	async function backfill121Metadata() {
		if (!batch?.id || actionBusy) return;
		setActionBusy('backfill-121-metadata');
		try {
			const result = await backfillBatch121Metadata(batch.id);
			const rows = Array.isArray(resultData(result, 'results')) ? resultData(result, 'results') : [];
			const backfilled = rows.filter(row => row?.status === 'backfilled').length;
			const aiFallback = rows.filter(row => row?.classification?.status === 'classified').length;
			const failed = rows.filter(row => row?.status === 'failed' || row?.classification?.status === 'failed').length;
			await refreshBatch();
			if (failed) message.warning(`121 已补全 ${backfilled} 本，AI 已兜底 ${aiFallback} 本；${failed} 本未完成，请在书籍资料中查看原因。`);
			else if (backfilled || aiFallback) message.success(`121 已补全 ${backfilled} 本，AI 已兜底 ${aiFallback} 本。`);
			else message.info('没有需要补全的 121 分类信息，也没有需要 AI 判断的书籍。');
		} catch (error) {
			message.error(error?.message || '121 分类信息回填失败');
		} finally { setActionBusy(''); }
	}
  async function saveBookMetadata() {
    if (!batch?.id || !metadataBook || metadataSaving) return;
    setMetadataSaving(true);
    try {
      await updateBookMetadata(batch.id, metadataBook.id, { metadata: metadataValue, expectedRevision: Number(metadataBook.revision || 0) });
      await refreshBatch();
      setViewingBook(null);
      setMetadataBook(null);
      message.success('小说列表元数据已保存。');
    } catch (error) { message.error(error?.message || '保存小说元数据失败'); } finally { setMetadataSaving(false); }
  }
  async function saveSettings(patch) {
    const normalized = normalizeAudioSourceSettings({ ...patch, audioDurationSeconds: 0 });
    // TTS and line-duration measurement are always part of the production
    // timeline. The audio source switch only changes the final mix, so every
    // save that keeps the shared audio flow enabled must refresh the measured
    // TTS before recompiling the director cards.
    const shouldPrepareAudio = normalized.audioPlanningEnabled === true;
    try {
      await saveBatchSettings(batch.id, { patch: normalized, expectedRevision: Number(batch?.settingsState?.revision || 0) }, { suppressGlobalError: true });
    } catch (error) {
      if (Number(error?.status) !== 409) {
        message.error(error?.message || '应用统一引擎配置失败');
        return false;
      }
      try {
        const latestResult = await getBatch(batch.id);
        const latestBatch = resultData(latestResult, 'batch');
        if (!sameSettingsPatch(latestBatch?.settingsState?.patch, batch?.settingsState?.patch)) {
          await refreshBatch();
          message.warning('配置已更新，已刷新当前统一配置；请核对后重新保存。');
          return false;
        }
        await saveBatchSettings(batch.id, {
          patch: normalized,
          expectedRevision: Number(latestBatch?.settingsState?.revision || 0)
        }, { suppressGlobalError: true });
      } catch (retryError) {
        message.error(retryError?.message || '应用统一引擎配置失败');
        return false;
      }
    }
    try {
      await refreshBatch();
      const savedResult = await getBatch(batch.id);
      const savedBatch = resultData(savedResult, 'batch');
      await syncUnifiedSettingsToBooks(savedBatch);
			const syncedResult = await getBatch(batch.id);
			const syncedBatch = resultData(syncedResult, 'batch');
      if (shouldPrepareAudio) {
        await prepareAudioPlanningForBatch(syncedBatch);
      }
      // Re-read after measurements so compilation uses the just-saved audio
      // asset, line durations and fingerprint instead of the pre-measurement
      // snapshot.
      const compiledResult = await getBatch(batch.id);
      const compiledBatch = resultData(compiledResult, 'batch');
			const recompileResult = await recompileUnifiedH3Prompts(compiledBatch);
      await refreshBatch();
			if (recompileResult.failed.length) message.warning(`统一配置已同步；${recompileResult.failed.join('；')}`);
			else if (recompileResult.waitingForAudio || recompileResult.manual) {
				message.info(`统一配置已同步；${recompileResult.compiled} 本最终 VIDEO Prompt 已更新，${recompileResult.manual ? `${recompileResult.manual} 本保留手动提示词，` : ''}${recompileResult.waitingForAudio ? `${recompileResult.waitingForAudio} 本等待已有逐行配音后再编译。` : ''}`);
			} else message.success(`统一配置已同步到当前批量全部小说；${recompileResult.compiled} 本最终 VIDEO Prompt 已更新。`);
      return true;
    } catch (error) { message.error(error?.message || '统一配置已保存，但刷新工作台失败'); return false; }
  }
  async function previewChangeImpact(patch) {
    try { const impact = await getChangeImpact(batch.id, { patch, expectedRevision: Number(batch?.settingsState?.revision || 0) }); return { ok: true, impact: resultData(impact, 'impact') }; } catch (error) { return { ok: false, message: error?.message || '无法读取配置影响' }; }
  }
  const batchMenuItems = [
    { key: 'production', label: '批量生成视频', disabled: !productionCapability.available, title: productionCapability.reason },
    { key: 'merge', label: '合并当前主 VIDEO', disabled: !mergeCapability.available, title: mergeCapability.reason }
  ];
  const automationStartBlocked = !books.length;
  const automationRetryQueued = (automationStatus?.books || []).filter(book => book?.retryRequested === true).length;
  const automationMenuItems = [
    { key: 'start', label: '开始定时', disabled: automationActive || automationStartBlocked },
    { key: 'pause', label: '暂停自动化', disabled: !automationActive },
    { key: 'resume', label: '继续已暂停自动化', disabled: !['paused', 'needs_attention', 'completed'].includes(automationState) },
    { key: 'retry', label: automationRetryQueued ? `重试队列中（${automationRetryQueued}）` : `重试失败小说（${Number(automationCounts.failed || 0) + Number(automationCounts.blocked || 0)}）`, disabled: automationRetryQueued > 0 || !(Number(automationCounts.failed || 0) + Number(automationCounts.blocked || 0)) },
    { key: 'cancel', label: '停止自动化编排', danger: true, disabled: ['idle', 'completed', 'cancelled'].includes(automationState) }
  ];
  const automationLabel = automationState === 'running'
    ? `自动生产 ${automationCounts.ready || 0}/${automationCounts.total || books.length}`
    : automationState === 'scheduled'
      ? '自动生产 · 已定时'
      : automationState === 'paused'
        ? '自动生产 · 已暂停'
        : automationState === 'needs_attention'
          ? '自动生产 · 待处理'
          : automationState === 'unavailable'
            ? '自动生产 · 状态暂不可读'
          : automationState === 'completed'
            ? '自动生产 · 已就绪'
            : '自动生产';
  const uploadMenuItems = [{ key: 'selected', label: `提交选中（${selectedBookIds.length}）`, disabled: !selectedBookIds.length }, { key: 'all', label: '提交全部' }];

  function moveColumnResize(event) {
    const resize = resizeStateRef.current;
    if (!resize || !Number.isFinite(event.clientX)) return;
    const nextWidths = resize.startingWidths.map((width, columnIndex) => columnIndex === resize.index ? Math.max(BATCH_FACTORY_TABLE_COLUMNS[resize.index].minWidth, Math.min(780, resize.startWidth + event.clientX - resize.startX)) : width);
    setColumnWidths(nextWidths);
  }

  function finishColumnResize() {
    if (!resizeActiveRef.current && !resizeStateRef.current) return;
    resizeActiveRef.current = false;
    resizeStateRef.current = null;
    document.body.classList.remove('batch-factory-column-resizing');
    document.removeEventListener('mousemove', moveColumnResize);
    document.removeEventListener('mouseup', finishColumnResize);
    document.removeEventListener('pointermove', moveColumnResize);
    document.removeEventListener('pointerup', finishColumnResize);
  }

  function startColumnResize(event, index, preventDefault = true) {
    if (resizeActiveRef.current) return;
    const header = event.currentTarget.closest('.shuihuo-workbench-head');
    const measured = header ? (getComputedStyle(header).gridTemplateColumns.match(/[\d.]+px/g)?.map(Number.parseFloat) || []) : [];
    const savedWidths = Array.isArray(columnWidths) && columnWidths.length === BATCH_FACTORY_TABLE_COLUMNS.length && columnWidths.every(Number.isFinite) ? columnWidths : null;
    const startingWidths = savedWidths || measured;
    if (startingWidths.length !== BATCH_FACTORY_TABLE_COLUMNS.length || !startingWidths.every(Number.isFinite)) return;
    if (preventDefault) {
      event.preventDefault();
      event.stopPropagation();
    }
    resizeActiveRef.current = true;
    resizeStateRef.current = { index, startX: event.clientX, startWidth: startingWidths[index], startingWidths };
    if (Number.isInteger(event.pointerId)) event.currentTarget.setPointerCapture?.(event.pointerId);
    document.body.classList.add('batch-factory-column-resizing');
    document.addEventListener('mousemove', moveColumnResize);
    document.addEventListener('mouseup', finishColumnResize, { once: true });
    document.addEventListener('pointermove', moveColumnResize);
    document.addEventListener('pointerup', finishColumnResize, { once: true });
  }

  return <section className="shuihuo-workbench batch-factory-workbench">
    <header className="shuihuo-workbench-header">
      <div className="shuihuo-workbench-heading"><button className="shuihuo-back-link" type="button" title="返回个人作品" aria-label="返回个人作品" onClick={onBack}><ArrowLeftOutlined /></button><strong>批量工厂 V12 · {batch?.title || '未命名批量'}</strong><div className="shuihuo-workbench-progress" aria-label={`完成率 ${batchProgress.completionPercent}%`}><div className="shuihuo-workbench-progress-track batch-factory-completion-track">{progressSegments.map(segment => <i key={segment.key} className={`is-${segment.key}`} title={`${segment.label} ${segment.count} 本`} style={{ width: `${segment.percent}%` }} />)}</div><span>{batchProgress.completionPercent}%</span></div></div>
      <div className="shuihuo-workbench-toolbar" role="toolbar" aria-label="批量工厂工具栏">
        <Button type="text" icon={<BarsOutlined />} onClick={() => setNovelListOpen(true)}>小说列表</Button>
        <Button type="text" icon={<SettingOutlined />} onClick={() => setUnifiedSettingsOpen(true)}>统一配置</Button>
        <Tooltip title="只重新获取每本书的智能统一视觉基线，并按现有导演卡重新编译最终 VIDEO Prompt；不重做资产、分镜、视频或成片。"><Button type="text" loading={String(actionBusy).startsWith('smart-unified-')} disabled={Boolean(actionBusy)} onClick={refreshAllSmartUnified}>刷新智能统一</Button></Tooltip>
        <Dropdown menu={{ items: automationMenuItems, onClick: ({ key }) => key === 'start' ? openAutomationStart() : runAutomationAction(key) }}><Button type="text" loading={Boolean(automationBusy)} disabled={automationState === 'idle' && automationStartBlocked}>{automationLabel}</Button></Dropdown>
        <Dropdown menu={{ items: batchMenuItems, onClick: ({ key }) => key === 'production' ? runProduction() : runMerge() }}><Button className="shuihuo-batch-button" type="text" icon={<PictureOutlined />} loading={actionBusy === 'production' || actionBusy === 'merge'}>批量操作</Button></Dropdown>
		<Tooltip title={cancelCapability.available ? '取消可取消的本地执行器任务；其它供应商保持在途状态。' : cancelCapability.reason}><Button className="shuihuo-cancel-button" type="text" loading={actionBusy === 'cancel'} disabled={!cancelCapability.available || Boolean(actionBusy)} onClick={cancelProduction}>取消操作</Button></Tooltip>
      </div>
      <div className="shuihuo-workbench-export"><Button type="text" icon={<BarsOutlined />} onClick={() => { setLogsOpen(true); loadRuntimeStatus(); }} loading={logsLoading}>任务/日志</Button><Dropdown menu={{ items: uploadMenuItems, onClick: ({ key }) => setUploadMode(key) }}><Button icon={<UploadOutlined />}>上传网络</Button></Dropdown></div>
      <div className="shuihuo-project-stats"><span><b>{books.length}</b> 本小说</span><span className="is-uploaded"><b>{batchProgress.uploaded}</b> 上传成功</span>{batchProgress.failed ? <span className="is-failed"><b>{batchProgress.failed}</b> 异常</span> : null}{batchProgress.running ? <span className="is-running"><b>{batchProgress.running}</b> 执行中</span> : null}{batchProgress.awaitingUpload ? <span className="is-awaiting-upload"><b>{batchProgress.awaitingUpload}</b> 待上传</span> : null}{batchProgress.pending ? <span><b>{batchProgress.pending}</b> 待开始</span> : null}{automationState !== 'idle' ? <span className={automationState === 'needs_attention' ? 'is-failed' : automationActive ? 'is-running' : 'is-awaiting-upload'}><b>{automationCounts.ready || 0}/{automationCounts.total || books.length}</b> 自动就绪</span> : null}<span><b>{mediaCounts.storyboards}</b> 分镜</span>{productionCapability.available ? <span className="is-running"><b>{mediaCounts.playable}</b> 可播放</span> : <span className="is-disabled">视频模块未启用</span>}{mergeCapability.available ? <span className="is-uploaded"><b>{mediaCounts.merges}</b> 成片</span> : <span className="is-disabled">合并模块未启用</span>}</div>
    </header>
    <div className="shuihuo-workbench-table batch-factory-workbench-table" role="table" aria-label="批量工厂小说生产表">
      <div className="shuihuo-workbench-head" role="row" style={columnWidths ? { gridTemplateColumns: columnWidths.map(width => `${width}px`).join(' ') } : undefined}>{BATCH_FACTORY_TABLE_COLUMNS.map((column, index) => <div role="columnheader" key={column.label}>{index === 0 ? <><Checkbox checked={books.length > 0 && selectedBookIds.length === books.length} indeterminate={selectedBookIds.length > 0 && selectedBookIds.length < books.length} onChange={event => setSelectedBookIds(event.target.checked ? books.map(book => book.id) : [])} aria-label="全选小说" /><span>序号</span></> : column.label}{index > 0 ? <button type="button" draggable className="batch-factory-column-resize-handle" onPointerDown={event => startColumnResize(event, index)} onMouseDown={event => startColumnResize(event, index)} onDragStart={event => startColumnResize(event, index, false)} onDrag={moveColumnResize} onDragEnd={finishColumnResize} aria-label={`调整${column.label}列宽`} title="拖动分隔线调整列宽"><i aria-hidden="true" /></button> : null}</div>)}</div>
      {books.map((book, index) => {
        const stageSummary = stageSummaries[book.id];
        const state = batchFactoryBookState(book, { productionStatus, mergeStatus, stageSummary });
        const rangeLines = contentRangeLinesForBook(book);
        // “小说正文”只展示从书城/巨量回填的原文摘录；改文草稿仍只用于生产，
        // 不能遮住已成功抓到的正文，避免误判为“没有获取内容”。
        const previewText = batchFactoryPreviewText(book.sourceText, book);
        // 下载必须使用抓取后保存的完整正文（当前最多 4000 字）；
        // “生产前 N 行”只用于分镜生产，不得截断用户下载的原文。
        const productionText = batchFactoryCleanSourceText(book.sourceText);
        const videos = book.videos || [];
        return <article className="shuihuo-workbench-row batch-factory-book-row" key={book.id} role="row" style={columnWidths ? { gridTemplateColumns: columnWidths.map(width => `${width}px`).join(' ') } : undefined}>
          <div className="shuihuo-workbench-cell shuihuo-order-cell"><Checkbox checked={selectedBookIds.includes(book.id)} onChange={event => setSelectedBookIds(current => event.target.checked ? [...new Set([...current, book.id])] : current.filter(id => id !== book.id))} aria-label={`选择 ${book.title || `小说 ${index + 1}`}`} /><strong>{index + 1}</strong></div>
          <div className="shuihuo-workbench-cell batch-factory-book-content" role="button" tabIndex={0} title="点击编辑当前小说所选行数的生产内容" onClick={() => openContentEditor(book)} onKeyDown={event => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); openContentEditor(book); } }}><strong>{book.title || `小说 ${index + 1}`}</strong><span>bookId：{book.bookId || '—'} · 书城：{bookPlatformName(book, platformNames)} · 生产前 {rangeLines} 行</span><p>{previewText || '原文尚未获取'}</p>{!previewText ? <BatchFactoryGiantMaterialPendingProgress book={book} batchId={batch?.id} visible onContentReady={refreshBatch} /> : null}<Space wrap><Button size="small" onClick={event => { event.stopPropagation(); openContentEditor(book); }}>编辑正文</Button><Button size="small" disabled={!productionText} onClick={event => { event.stopPropagation(); downloadProductionContent({ title: book.title, content: productionText }); }}>下载 TXT</Button><Popconfirm title="删除本书？" description="将删除本书及本地生产记录，不影响已提交到 121 的内容。" okText="删除" cancelText="取消" onConfirm={event => { event?.stopPropagation?.(); return deleteBook(book); }}><Button danger size="small" icon={<DeleteOutlined />} loading={actionBusy === `delete-${book.id}`} onClick={event => event.stopPropagation()}>删除本书</Button></Popconfirm></Space></div>
          <div className="shuihuo-workbench-cell batch-factory-book-config-cell"><div className="batch-factory-book-config-regions">{batchFactoryWorkbenchConfigRegions(book).map(region => { const status = region.combinedStatus || bookConfigRegionStatus(book, region.key); const detail = region.key === 'assets' ? bookAssetSummary(book) : status.label; return <button type="button" key={region.key} className={`batch-factory-book-config-region is-${status.tone}`} onClick={() => region.key === 'assets' ? setAssetBook(book) : setConfigTarget({ book, region: region.key })}><b>{region.label}</b><small>{detail}</small></button>; })}</div></div>
		  <InlineBookPrompts batch={batch} book={book} batchId={batch?.id} settingsRevision={batch?.settingsState?.revision} selectedVideoId={rowStoryboardSelection[book.id] || videos[0]?.id || ''} onSelectedVideoChange={videoId => setRowStoryboardSelection(current => ({ ...current, [book.id]: videoId }))} onManage={videoId => { setPromptVideoId(videoId || ''); setPromptBook(book); }} />
		  <div className="shuihuo-workbench-cell shuihuo-library-cell batch-factory-book-library-cell"><InlineMediaLibrary book={book} versionsByVideo={mediaVersionsByVideo} productionStatus={productionStatus} mergeJob={latestBookMerge(mergeStatus, book.id)} aspectRatio={effectiveBookSettings(batch, book).aspectRatio} h3CompilationMessage={h3CompilationStatus(effectiveBookSettings(batch, book))} selectedVideoId={rowStoryboardSelection[book.id] || videos[0]?.id || ''} onSelectedVideoChange={videoId => setRowStoryboardSelection(current => ({ ...current, [book.id]: videoId }))} onManage={videoId => { setMediaVideoId(videoId || ''); setMediaStartTab('clips'); setMediaBook(book); }} onOpenMerge={() => { setMediaVideoId(''); setMediaStartTab('merges'); setMediaBook(book); }} /></div>
          <div className="shuihuo-workbench-cell batch-factory-actions">
            <div className="batch-factory-action-group is-utility"><span>资料与流程</span><div className="batch-factory-action-button-grid"><Button size="small" onClick={() => setViewingBook(book)}>查看资料</Button>{previewText ? <Button size="small" onClick={() => refreshBatch()} disabled={Boolean(actionBusy)}>刷新状态</Button> : <Button size="small" type="primary" onClick={() => fetchMissingBookSource(book)} loading={actionBusy === `source-${book.id}`} disabled={Boolean(actionBusy)}>获取正文</Button>}</div></div>
            <div className="batch-factory-action-group is-production"><span>资产获取</span><div className="batch-factory-action-button-grid"><Tooltip title={runCapability.available ? '提取当前书的人物、场景、道具提示词；手动资产保留。' : runCapability.reason}><Button size="small" onClick={() => runBookStageAction(book, 'assets', 'force')} loading={actionBusy === stageActionKey('assets', book.id)} disabled={!runCapability.available || Boolean(actionBusy)}>提取</Button></Tooltip><Tooltip title="只刷新智能统一视觉基线，不重做资产、分镜、视频或成片。"><Button size="small" loading={actionBusy === `smart-unified-${book.id}`} disabled={Boolean(actionBusy)} onClick={() => refreshBookSmartUnifiedAction(book)}>刷新智能统一</Button></Tooltip><Tooltip title="打开资产图选择与生成面板；选择资产和图片模型后生成。"><Button size="small" onClick={() => setAssetBook(book)} disabled={Boolean(actionBusy)}>生成图片</Button></Tooltip></div></div>
            <div className="batch-factory-action-group is-production"><span>视频获取</span><div className="batch-factory-action-button-grid"><Tooltip title={runCapability.available ? (h3DirectorCards(book).length && usesH3VideoPrompt(effectiveBookSettings(batch, book)) ? '复用已有统一导演卡，按当前规则编译最终 VIDEO Prompt。' : '按统一流程生成导演分镜：选中的视频提示词决定 SD 或 H3 的输出协议。') : runCapability.reason}><Button size="small" onClick={() => runBookStageAction(book, 'director', 'compile')} loading={actionBusy === stageActionKey('director', book.id)} disabled={!runCapability.available || Boolean(actionBusy)}>提取</Button></Tooltip><Tooltip title={productionCapability.available ? '按当前书最终编译视频提示词、可用参考图和画幅创建 VIDEO 任务。' : productionCapability.reason}><Button size="small" className="batch-factory-action-video" onClick={() => runBookStageAction(book, 'video')} loading={actionBusy === stageActionKey('video', book.id)} disabled={!productionCapability.available || Boolean(actionBusy)}>生成视频</Button></Tooltip></div></div>
            <div className="batch-factory-action-group is-production"><span>画面获取</span><div className="batch-factory-action-button-grid"><Tooltip title={runCapability.available ? '根据已生成的分镜卡（视频提示词）和画面提示词预设，生成每张分镜的画面提示词。' : runCapability.reason}><Button size="small" onClick={() => runBookStageAction(book, 'visual', 'force')} loading={actionBusy === stageActionKey('visual', book.id)} disabled={!runCapability.available || Boolean(actionBusy)}>提取</Button></Tooltip><Tooltip title="画面首帧图片生成服务尚未配置；可先在分镜提示词中查看或编辑画面提示词。"><Button size="small" disabled>生成图片</Button></Tooltip></div></div>
            <div className="batch-factory-action-group is-recovery"><span>异常处理</span><div className="batch-factory-action-button-grid"><Button size="small" onClick={() => refreshBatch()} disabled={Boolean(actionBusy)}>刷新</Button><Button size="small" danger onClick={() => retryLastFailedStage(book)} loading={actionBusy === stageActionKey('retry', book.id)} disabled={Boolean(actionBusy) && actionBusy !== stageActionKey('retry', book.id)}>重试失败步骤</Button></div></div>
          </div>
        </article>;
      })}
      {!books.length ? <div className="shuihuo-workbench-empty">当前批量还没有小说。返回个人作品后，从“批量工厂”新建书单。</div> : null}
    </div>

    <Modal title={`小说列表 · ${books.length} 本`} open={novelListOpen} onCancel={() => setNovelListOpen(false)} footer={null} width="min(1480px, calc(100vw - 48px))" className="batch-factory-novel-modal"><NovelMetadata books={books} createdAt={batch?.createdAt} selectedBookIds={selectedBookIds} onSelectionChange={setSelectedBookIds} onViewBook={setViewingBook} platformNames={platformNames} platformOptions={platformOptions} productionStatus={productionStatus} mergeStatus={mergeStatus} stageSummaries={stageSummaries} batchId={batch?.id} onContentReady={refreshBatch} onBackfill121Metadata={backfill121Metadata} backfilling121Metadata={actionBusy === 'backfill-121-metadata'} onAddBooks={addBooksToBatch} /></Modal>
    <Modal title={viewingBook?.title || '小说详情'} open={Boolean(viewingBook)} onCancel={() => setViewingBook(null)} footer={viewingBook ? <Space><Button loading={actionBusy === `classify-${viewingBook.id}`} onClick={() => classifyBookMetadata(viewingBook)}>AI 判断发布信息</Button>{isBookUploadedTo121(viewingBook) ? <Button danger onClick={() => { setSelectedBookIds([viewingBook.id]); setUploadMode('reupload'); }}>重新上传</Button> : null}<Button onClick={() => openSourceEditor(viewingBook)}>编辑完整正文</Button><Button onClick={() => openMetadataEditor(viewingBook)}>编辑列表信息</Button></Space> : null} width={860} className="batch-factory-book-detail-modal">{viewingBook ? (() => { const summary = stageSummaries[viewingBook.id]; const state = batchFactoryBookState(viewingBook, { productionStatus, mergeStatus, stageSummary: summary }); const timeline = batchFactoryBookTimeline(viewingBook, { productionStatus, mergeStatus, stageSummaries: stageSummaries?.[viewingBook.id] }); const uploadProgress = uploadProgressForBook(viewingBook); const uploadHistory = Array.isArray(viewingBook?.sourceMetadata?.websiteSubmitHistory) ? viewingBook.sourceMetadata.websiteSubmitHistory : []; const classifyBusy = actionBusy === `classify-${viewingBook.id}`; const classifyFailed = String(viewingBook?.sourceMetadata?.classifyStatus || '').trim() === 'failed'; return <div className="batch-factory-book-detail"><section className="batch-factory-book-detail-status"><h3>流程异常与生成状态</h3><div><Tag color={state.tone}>{state.label}</Tag><strong>{state.detail}</strong></div>{classifyBusy ? <Alert type="info" showIcon message="正在识别男女频、风格和标签…" description="识别完成后会自动写入当前书；不会阻塞生产流程。" /> : null}{classifyFailed ? <Alert type="warning" showIcon message="AI 判断未完成" description={classificationDetail(viewingBook.sourceMetadata)} /> : null}{uploadProgress ? <Alert type={uploadProgress.status === 'failed' ? 'error' : 'info'} showIcon message={`121 上传：${uploadProgress.message || uploadProgress.phase}`} description={`阶段：${uploadProgress.phase || '—'} · ${uploadProgress.updatedAt || '—'}`} /> : null}{state.label === '异常' ? <Button danger size="small" loading={actionBusy === stageActionKey('retry', viewingBook.id)} onClick={() => retryLastFailedStage(viewingBook)}>重试失败步骤</Button> : null}<div className="batch-factory-production-timeline is-detail">{timeline.map(item => <div className={`is-${item.status}`} key={item.key}><i aria-hidden="true" /><span>{item.label}</span><small>{item.detail}</small></div>)}</div></section><section className="batch-factory-book-detail-content"><h3>小说正文</h3><pre>{viewingBook.workingFrontContent || batchFactoryCleanSourceText(viewingBook.sourceText) || '尚未获取原文。'}</pre></section><Descriptions bordered size="small" column={2}><Descriptions.Item label="Book ID">{viewingBook.bookId || '—'}</Descriptions.Item><Descriptions.Item label="书城">{bookPlatformName(viewingBook, platformNames)}</Descriptions.Item><Descriptions.Item label="风格">{value(viewingBook.sourceMetadata, 'style')}</Descriptions.Item><Descriptions.Item label="男女频">{value(viewingBook.sourceMetadata, 'gender')}</Descriptions.Item><Descriptions.Item label="标签">{value(viewingBook.sourceMetadata, 'tags')}</Descriptions.Item><Descriptions.Item label="AI 判断">{classificationLabel(viewingBook.sourceMetadata)}</Descriptions.Item><Descriptions.Item label="来源">{value(viewingBook.sourceMetadata, 'sourceMode') === 'manual_original' ? '手动书单' : '小说获取'}</Descriptions.Item></Descriptions>{!classifyBusy ? <Alert type={classifyFailed ? 'warning' : 'info'} showIcon message={`AI 判断：${classificationLabel(viewingBook.sourceMetadata)}`} description={classificationDetail(viewingBook.sourceMetadata)} /> : null}{uploadHistory.length ? <section className="batch-factory-book-upload-history"><h3>121 上传历史</h3>{uploadHistory.slice().reverse().map((item, index) => <p key={`${item.submittedAt || 'history'}-${index}`}>{item.submittedAt || '—'} · {item.status || '—'} · {item.detail || '—'}</p>)}</section> : null}</div>; })() : null}</Modal>
    <Modal title={metadataBook ? `编辑列表信息 · ${metadataBook.title}` : '编辑列表信息'} open={Boolean(metadataBook)} onCancel={() => setMetadataBook(null)} onOk={saveBookMetadata} confirmLoading={metadataSaving} okText="保存" width={620}>{metadataBook ? <Space direction="vertical" size={12} style={{ width: '100%' }}><Alert type="info" showIcon message="此处保存风格、男女频、标签、推荐理由和评级" description="不会改动小说正文或书城来源。" />{[['style', '风格'], ['gender', '男女频'], ['tags', '标签'], ['reason', '推荐理由'], ['rating', '评级']].map(([key, label]) => <label key={key} className="batch-factory-engine-field"><span><b>{label}</b></span><Input value={metadataValue[key] || ''} onChange={event => setMetadataValue(current => ({ ...current, [key]: event.target.value }))} /></label>)}</Space> : null}</Modal>
    <Modal title={editingContentBook ? `编辑生产内容 · ${editingContentBook.title}` : '编辑生产内容'} open={Boolean(editingContentBook)} onCancel={() => setEditingContentBook(null)} onOk={saveWorkingContent} confirmLoading={contentSaving} okText="保存生产内容" width={820} destroyOnClose><Space direction="vertical" size={14} style={{ width: '100%' }}><Alert type="info" showIcon message="这里只展示和编辑当前书所选行数的生产内容" description="完整正文只在“查看资料”中展示；录用爆款候选时才会替换完整正文。" />{editingContentBook ? <BatchFactoryGiantMaterialPendingProgress book={editingContentBook} batchId={batch?.id} onContentReady={refreshBatch} /> : null}<Input.TextArea rows={16} value={editingContentValue} onChange={event => { setEditingContentValue(event.target.value); setEditingContentMode('custom'); }} placeholder="输入当前小说的生产内容" /><label className="shuihuo-form-label"><span>衍生开篇</span><Select value={derivedOpeningPresetId || undefined} onChange={setDerivedOpeningPresetId} options={derivedOpeningOptions} loading={!derivedOpeningOptions.length} placeholder="选择已发布的衍生开篇提示词" style={{ width: 320 }} /></label>{!workingFrontCapability.available ? <Alert type="warning" showIcon message="当前不能生成爆款候选" description={workingFrontCapability.reason || '请先完成当前书的可执行配置。'} /> : null}<Space wrap><Tooltip title={workingFrontCapability.available ? '按选中的衍生开篇提示词生成候选；不会立即改动完整正文。' : workingFrontCapability.reason}><Button type="primary" onClick={createViralCandidate} loading={rewritingFront} disabled={!workingFrontCapability.available || !String(editingContentValue || '').trim() || !derivedOpeningPresetId}>生成爆款候选</Button></Tooltip><Tooltip title={workingFrontCapability.available ? '按当前选择的同一提示词重新生成候选；不会立即改动完整正文。' : workingFrontCapability.reason}><Button onClick={createViralCandidate} loading={rewritingFront} disabled={!workingFrontCapability.available || !String(editingContentValue || '').trim() || !derivedOpeningPresetId}>重试生成爆款候选</Button></Tooltip></Space>{viralCandidate ? <Alert type="warning" showIcon message="爆款候选尚未录用" description={<Space direction="vertical" size={8} style={{ width: '100%' }}><pre className="batch-factory-viral-candidate">{viralCandidate}</pre><Space><Button type="primary" loading={sourceSaving} onClick={adoptViralCandidate}>录用并替换完整正文</Button><Button onClick={cancelViralCandidate}>取消候选</Button></Space></Space>} /> : null}</Space></Modal>
    <Modal title={sourceEditorBook ? `编辑完整正文 · ${sourceEditorBook.title}` : '编辑完整正文'} open={Boolean(sourceEditorBook)} onCancel={() => setSourceEditorBook(null)} onOk={saveSourceEditor} confirmLoading={sourceSaving} okText="保存完整正文" width={860} destroyOnClose>
      <Space direction="vertical" size={14} style={{ width: '100%' }}>
        <Alert type="info" showIcon message="编辑的是完整获取正文" description="保存后会替换当前书原文（最多 4000 字）；工作区卡片和工作区编辑始终只显示你选择的前几行。" />
        <Button onClick={() => downloadProductionContent({ title: sourceEditorBook?.title, content: sourceEditorValue })} disabled={!String(sourceEditorValue || '').trim()}>下载完整正文 TXT</Button>
        <Input.TextArea rows={20} value={sourceEditorValue} onChange={event => setSourceEditorValue(event.target.value)} placeholder="完整获取正文" />
      </Space>
    </Modal>
    <BatchFactoryBookSettingsModal open={Boolean(configTarget)} batch={batch} book={configTarget?.book} activeRegion={configTarget?.region} onClose={() => setConfigTarget(null)} onSaved={refreshAfterBookSettingsSaved} onOpenBookAssets={book => setAssetBook(book)} />
    <Modal title={assetBook ? `人物场景预设 · ${assetBook.title}` : '人物场景预设'} open={Boolean(assetBook)} onCancel={() => setAssetBook(null)} footer={null} width="min(1440px, calc(100vw - 48px))" className="batch-factory-assets-modal">{assetBook ? <AssetEditor book={assetBook} batchId={batch?.id} onSaved={refreshBatch} onGenerate={async textModelId => { await refreshAssetPresetSnapshot(assetBook); await runBookStageAction(assetBook, 'assets', 'missing', '', textModelId); }} onRegenerate={async textModelId => { await refreshAssetPresetSnapshot(assetBook); await runBookStageAction(assetBook, 'assets', 'force', '', textModelId); }} onRetry={textModelId => retryLastFailedStage(assetBook, '', textModelId)} assetRules={effectiveBookAssetRules(batch, assetBook)} onAssetPromptChange={async selection => { const bookPromptConfig = assetBook?.settingsState?.patch?.aiPromptConfig || {}; const assetRules = effectiveBookAssetRules(batch, assetBook); await saveBookOverrideWithRetry(batch.id, assetBook.id, assetBook.revision, { patch: { aiPromptConfig: { ...bookPromptConfig, assets: { ...assetRules, extraction: selection, scope: 'custom', bookIds: [assetBook.id] } } } }); await refreshBatch(); }} canGenerate={runCapability.available} generateReason={runCapability.reason} generating={actionBusy === stageActionKey('assets', assetBook.id)} engineSettings={effectiveBookSettings(batch, assetBook)} /> : null}</Modal>
	<Modal title={promptBook ? `分镜卡（视频提示词） · ${promptBook.title}` : '分镜卡（视频提示词）'} open={Boolean(promptBook)} onCancel={() => { setPromptBook(null); setPromptVideoId(''); }} footer={null} width={900} className="batch-factory-prompt-modal">{promptBook ? <PromptPanel key={`${promptBook.id}:${promptBook.revision || 0}:${promptBook.directorRevision?.id || ''}`} book={promptBook} batch={batch} batchId={batch?.id} settingsRevision={batch?.settingsState?.revision} initialVideoId={promptVideoId} onSaved={refreshBatch} onCompile={() => runBookStageAction(promptBook, 'director', 'compile')} onRegenerate={() => runBookStageAction(promptBook, 'director', 'force')} onRegenerateVisual={() => runBookStageAction(promptBook, 'visual', 'force')} onRetry={videoId => retryLastFailedStage(promptBook, videoId)} onGenerateVideo={videoId => runBookStageAction(promptBook, 'video', 'missing', videoId)} onGenerateVisual={() => setAssetBook(promptBook)} regenerating={actionBusy === stageActionKey('director', promptBook.id) || actionBusy === stageActionKey('video', promptBook.id) || actionBusy === stageActionKey('retry', promptBook.id)} productionAvailable={productionCapability.available} productionReason={productionCapability.reason} /> : null}</Modal>
	<Modal
      title={mediaBook ? `片段库 · ${mediaBook.title}` : '片段库'}
      open={Boolean(mediaBook)}
      onCancel={() => { setMediaBook(null); setMediaVideoId(''); setMediaStartTab('clips'); }}
      footer={null}
      width="min(1180px, calc(100vw - 40px))"
      className="batch-factory-media-library-modal"
      destroyOnClose
    >
      {mediaBook ? <MediaVersionPanel
        book={mediaBook}
        batchId={batch?.id}
        versionsByVideo={mediaVersionsByVideo}
        productionStatus={productionStatus}
        mergeJob={latestBookMerge(mergeStatus, mediaBook.id)}
        mergeJobs={bookMergeJobs(mergeStatus, mediaBook.id)}
        mergeSettings={effectiveBookSettings(batch, mediaBook)}
        mergeSpeed={mergeSpeed}
        onMergeSpeedChange={setMergeSpeed}
        onMerge={options => runBookMerge(mediaBook, options)}
        mergeAvailable={mergeCapability.available}
        mergeReason={mergeCapability.reason}
        merging={actionBusy === `merge-${mediaBook.id}`}
        onSaved={async () => { await refreshBatch(); await loadRuntimeStatus({ quiet: true }); }}
        onDeleted={async () => { await refreshBatch(); await loadRuntimeStatus({ quiet: true }); }}
        onUpload={() => { setSelectedBookIds([mediaBook.id]); setMediaBook(null); setMediaVideoId(''); setMediaStartTab('clips'); setUploadMode('selected'); }}
        onRegenerate={videoId => runBookStageAction(mediaBook, 'video', 'force', videoId)}
        onRetry={videoId => retryLastFailedStage(mediaBook, videoId)}
        regenerating={actionBusy === stageActionKey('video', mediaBook.id) || actionBusy === stageActionKey('retry', mediaBook.id)}
        productionAvailable={productionCapability.available}
        productionReason={productionCapability.reason}
        initialVideoId={mediaVideoId}
        initialTab={mediaStartTab}
        onOpenDirector={frameKey => { setPendingMediaPrompt({ bookId: mediaBook.id, frameKey: frameKey || '' }); setMediaBook(null); }}
      /> : null}
    </Modal>

    <BatchFactoryUnifiedSettingsModal open={unifiedSettingsOpen} batch={batch} onClose={() => setUnifiedSettingsOpen(false)} onSaved={saveSettings} />
    <Modal title="任务 / 日志" open={logsOpen} onCancel={() => setLogsOpen(false)} footer={<Button onClick={() => loadRuntimeStatus()}>刷新状态</Button>} width={860}><BatchLogs books={books} automationStatus={automationStatus} productionStatus={productionStatus} mergeStatus={mergeStatus} error={logsError} /></Modal>
    <Modal className="batch-factory-automation-modal" title={<div className="batch-factory-automation-title"><strong>自动生产</strong><span>批量执行设置</span></div>} open={automationStartOpen} onCancel={() => setAutomationStartOpen(false)} width={680} destroyOnClose footer={<div className="batch-factory-automation-footer"><Button onClick={() => setAutomationStartOpen(false)}>取消</Button><Button onClick={() => { setAutomationStartOpen(false); setUnifiedSettingsOpen(true); }}>统一配置</Button><span /><Button loading={contentSettingsSaving} onClick={saveContentSettings}>保存内容设置</Button><Button loading={automationBusy === 'start'} onClick={() => runAutomationAction('start', { immediate: true })}>立即执行</Button><Button type="primary" loading={automationBusy === 'start'} onClick={() => runAutomationAction('start')}>保存定时任务</Button></div>}>
      <div className="batch-factory-automation-form">
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">01</span><div><b>内容设置</b><small>应用到当前批量的全部小说</small></div></div><div className="batch-factory-automation-grid"><label className="batch-factory-engine-field"><span><b>内容范围</b><small>生产使用的正文行数</small></span><InputNumber min={1} max={500} value={contentRangeLines} onChange={value => setContentRangeLines(value || 5)} addonAfter="行" style={{ width: '100%' }} /></label><label className="batch-factory-engine-field"><span><b>内容截取</b><small>正文最多保留字数</small></span><Select value={contentCaptureCharacters} onChange={setContentCaptureCharacters} style={{ width: '100%' }} options={[1000, 2000, 4000, 8000, 12000, 20000, 50000, 100000].map(value => ({ value, label: `${value} 字` }))} /></label></div></section>
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">02</span><div><b>生产策略</b><small>选择预设和生成范围</small></div></div><div className="batch-factory-automation-grid"><label className="batch-factory-engine-field"><span><b>自动化预设</b></span><Select value={automationPresetID || undefined} onChange={setAutomationPresetID} placeholder="选择已保存预设" options={automationPresets.map(item => ({ value: item.id, label: `${item.name} · v${item.version}` }))} style={{ width: '100%' }} /></label><label className="batch-factory-engine-field"><span><b>执行模式</b></span><Select value={automationRunMode} onChange={setAutomationRunMode} style={{ width: '100%' }} options={[{ value: 'storyboard_only', label: '只生成分镜' }, { value: 'video_no_submit', label: '生成视频不提交' }, { value: 'full_submit', label: '全自动生成并提交' }]} /></label></div></section>
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">03</span><div><b>启动方式</b><small>立即开始或按北京时间启动</small></div></div><div className="batch-factory-automation-grid"><label className="batch-factory-engine-field"><span><b>自动启动时间</b></span><Input type="datetime-local" value={automationScheduledAt} onChange={event => setAutomationScheduledAt(event.target.value)} /></label><label className="batch-factory-engine-field"><span><b>同时处理书籍</b></span><Select value={automationConcurrency} onChange={value => setAutomationConcurrency(normalizeAutomationConcurrency(value))} style={{ width: '100%' }} options={[1, 2, 4].map(value => ({ value, label: `${value} 本并行` }))} /></label></div></section>
      </div>
    </Modal>
    <UploadNetwork batch={batch} books={books} selectedBookIds={selectedBookIds} productionStatus={productionStatus} mergeStatus={mergeStatus} mode={uploadMode} onClose={() => setUploadMode('')} onOpenPublish={() => { setUploadMode(''); setUnifiedSettingsOpen(true); }} onRefresh={refreshBatch} />
  </section>;
}
