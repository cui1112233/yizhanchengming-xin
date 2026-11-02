import { getBookAssetImageBlob, getBookAssetImageGenerationTask, getBatch, listBookAssetImages, listComicPages, saveBookOverride } from '../../../shared/api/batchFactory';
import { getConfig } from '../../../shared/api/config';
import { textToSpeech } from '../../../shared/api/tts';
import { audioPlanningRequired, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { BOOK_CONFIG_REGIONS } from './batchFactoryBookConfigRegions';
import { buildBatchFactoryH3Constraints } from './batchFactoryH3Constraints';
import { batchFactoryCleanSourceText, batchFactoryPreviewText, batchFactoryProductionText, batchFactorySourceText, comicAnalysisCharactersForBook } from './batchFactoryContentRange';
import { readStoryboardPromptEditableText } from './storyboardPromptEditor';

const DEFAULT_TTS = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 };
const COMIC_PAGE_RESULT_POLL_INTERVAL = 2000;
const COMIC_PAGE_RESULT_TIMEOUT = 15 * 60 * 1000;
const ASSET_IMAGE_RESULT_POLL_INTERVAL = 2000;
const ASSET_IMAGE_RESULT_TIMEOUT = 15 * 60 * 1000;

export function readBatchFactoryAudioDuration(blob) {
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

export function batchFactoryAudioFingerprint(input, tts = {}) {
  const payload = JSON.stringify([String(input || ''), String(tts.voice || ''), String(tts.style || ''), Number(tts.speed || 0), Number(tts.pitch || 0)]);
  let hash = 2166136261;
  for (let index = 0; index < payload.length; index += 1) {
    hash ^= payload.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return `a1-${(hash >>> 0).toString(16)}-${payload.length}`;
}

export function batchFactoryBookAudioInput(book = {}) {
  return batchFactoryProductionText(book?.sourceText || book?.contentPreview, book?.workingFrontContent, book);
}

export async function batchFactoryBookTts(settings = {}) {
  const config = await getConfig().catch(() => ({}));
  return { ...DEFAULT_TTS, ...(config?.tts || {}), ...(settings?.tts || {}) };
}

export async function generateBatchFactoryBookAudioMeasurement(book, settings = {}) {
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

export function blobToBase64(blob) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onerror = () => reject(new Error('无法读取 H3 配音数据'));
		reader.onload = () => resolve(String(reader.result || '').split(',').at(-1) || '');
		reader.readAsDataURL(blob);
	});
}

export async function saveBookOverrideWithRetry(batchId, bookId, expectedRevision, input) {
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

export function batchFactoryContentMode(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const stored = String(metadata.productionContentMode || '').trim().toLowerCase();
  if (stored === 'viral' || stored === 'custom' || stored === 'original') return stored;
  const working = String(book?.workingFrontContent || '').trim();
  const source = batchFactorySourceText(book?.sourceText);
  return !working || (source && working === source) ? 'original' : 'custom';
}

export function batchFactoryContentModeLabel(book = {}) {
  const mode = batchFactoryContentMode(book);
  return mode === 'viral' ? '爆款' : mode === 'custom' ? '自创' : '原文';
}

export function batchFactorySourceVersionLabel(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const version = String(metadata.sourceContentVersion || '').trim().toLowerCase();
  const sourceMode = String(metadata.sourceMode || '').trim().toLowerCase();
  if (sourceMode === 'giant_material') return '巨量素材';
  if (/^ai[1-5]$/.test(version)) return `小说获取 ${version.toUpperCase()}`;
  if (version === 'sensitive') return '小说获取（敏感）';
  if (sourceMode === 'novel_fetch') return '小说获取';
  return '原文';
}

export function batchFactorySourceVersionClass(book = {}) {
  const version = String(book?.sourceMetadata?.sourceContentVersion || '').trim().toLowerCase();
  return ['sensitive', 'original', 'ai1', 'ai2', 'ai3', 'ai4', 'ai5'].includes(version) ? version : 'original';
}

export function batchFactorySourceSummary(book = {}, productionMode = 'video') {
  const source = batchFactoryCleanSourceText(book?.sourceText || '');
  if (!source) return { ready: false, text: '暂无可用正文' };
  if (productionMode === 'comic') return { ready: true, text: `已带入 · ${source.length} 字 · 漫画分析 ${comicAnalysisCharactersForBook(book)} 字` };
  const productionLines = batchFactoryPreviewText(book?.sourceText || '', book).split(/\r?\n/).filter(Boolean).length;
  return { ready: true, text: `已带入 · ${source.length} 字 · 生产前 ${productionLines} 行` };
}

export function batchFactoryWorkbenchConfigRegions() {
  return BOOK_CONFIG_REGIONS.filter(region => !['video', 'visual'].includes(region.key));
}

export function value(metadata, key) {
  const current = String(metadata?.[key] || '').trim();
  return current || (['style', 'gender', 'tags'].includes(key) ? '待 AI 判断' : '—');
}

export function bookPlatformName(book, platformNames = {}) {
  const metadata = book?.sourceMetadata || {};
  const platformCode = String(metadata.platformCode || '').trim().toUpperCase();
  if (platformCode === 'YG') return '点众付费';
  const platform = String(book?.platform || '').trim().toUpperCase();
  if (platform === 'YG' || platform === '阳光') return '点众付费';
  return String(metadata.platformName || metadata.platformLabel || platformNames[String(book?.platform || '')] || '').trim() || '未命名书城';
}

export function sensitiveProcessingDetail(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const hitData = metadata.sensitiveHits && typeof metadata.sensitiveHits === 'object' ? metadata.sensitiveHits : {};
  const fixedData = metadata.sensitiveFixed && typeof metadata.sensitiveFixed === 'object' ? metadata.sensitiveFixed : {};
  const items = Array.isArray(fixedData.items) ? fixedData.items : [];
  const hitCount = Number(metadata.sensitiveHitCount ?? hitData.hit_count ?? hitData.hitCount ?? items.length ?? 0);
  const replacedCount = Number(metadata.sensitiveReplaced ?? fixedData.fixed_count ?? 0);
  const originalChars = Number(metadata.originalRawChars || 0);
  const processedChars = Number(metadata.processedChars || String(book?.sourceText || '').length || 0);
  const stages = Array.isArray(metadata.processingStages) ? metadata.processingStages : [];
  return {
    status: String(metadata.sensitiveStatus || fixedData.status || (hitCount ? 'done' : '')).trim(),
    hitCount: Number.isFinite(hitCount) ? hitCount : 0,
    replacedCount: Number.isFinite(replacedCount) ? replacedCount : 0,
    originalChars: Number.isFinite(originalChars) ? originalChars : 0,
    processedChars: Number.isFinite(processedChars) ? processedChars : 0,
    stages,
    items
  };
}

export function batchFactorySourceKind(book = {}) {
  const mode = String(book?.sourceMetadata?.sourceMode || '').trim().toLowerCase();
  if (mode === 'giant_material') return '巨量素材';
  if (mode === 'novel_fetch') return '小说获取';
  if (mode === 'manual_original') return '书城原文';
  return '原文';
}

export function isBatchFactoryContentReadRunning(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const stage = String(metadata.originalReadStage || '').trim().toLowerCase();
  return metadata.contentPending === true && (stage === 'direct' || stage === 'ocr');
}

export function batchFactorySourceStatus(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const sourceReady = Boolean(batchFactorySourceText(book?.sourceText));
  if (sourceReady) return { label: '原文已获取', detail: `${batchFactorySourceKind(book)} · ${batchFactoryCleanSourceText(book.sourceText).length} 字`, tone: 'ready' };
  const error = String(metadata.originalReadError || '').trim();
  if (String(metadata.originalReadStage || '').trim().toLowerCase() === 'failed' && error) return { label: '书城读取失败', detail: `${batchFactorySourceKind(book)} · ${error}；可手动重试或使用 OCR`, tone: 'error' };
  if (metadata.contentPending) {
    const stage = String(metadata.originalReadStage || '').trim().toLowerCase();
    if (stage === 'ocr') return { label: 'OCR 等待执行器', detail: '巨量素材 · 已进入 OCR 队列，等待执行器领取', tone: 'running' };
    if (stage === 'direct') return { label: '书城获取中', detail: `${batchFactorySourceKind(book)} · 正在读取书城原文`, tone: 'running' };
    return { label: '正文读取中', detail: `${batchFactorySourceKind(book)} · 等待读取任务`, tone: 'running' };
  }
  if (error) return { label: '读取失败', detail: `${batchFactorySourceKind(book)} · ${error}`, tone: 'error' };
  return { label: '原文待获取', detail: `${batchFactorySourceKind(book)} · 尚未获取正文`, tone: 'pending' };
}

export function resultData(result, key) { return result?.[key] || result || {}; }

export async function waitForComicPageResult(batchId, bookId, videoId) {
  const deadline = Date.now() + COMIC_PAGE_RESULT_TIMEOUT;
  let lastReadError = null;
  while (Date.now() < deadline) {
    try {
      const response = await listComicPages(batchId, bookId, { silent: true, suppressGlobalError: true });
      const pages = resultData(response, 'pages') || [];
      const page = pages.find(item => item?.videoId === videoId);
      const primaryImage = page?.images?.find(image => image?.isPrimary) || page?.images?.[0];
      if (primaryImage) return page;
      if (['generation_failed', 'failed'].includes(String(page?.status || '').toLowerCase())) {
        throw new Error('当前漫画页生成失败，请查看日志后重试。');
      }
      lastReadError = null;
    } catch (error) {
      // A brief read failure should not lose a provider result that is still
      // being written back. Authentication/permission errors are not
      // recoverable by polling and should still surface immediately.
      if ([401, 403].includes(Number(error?.status))) throw error;
      lastReadError = error;
    }
    await new Promise(resolve => window.setTimeout(resolve, COMIC_PAGE_RESULT_POLL_INTERVAL));
  }
  throw lastReadError || new Error('漫画页生成超过 15 分钟，已停止等待；请刷新状态后重试。');
}

export async function waitForBookAssetImages(batchId, bookId, assetIds, taskId = '') {
  const wanted = [...new Set((assetIds || []).map(id => String(id || '').trim()).filter(Boolean))];
  const deadline = Date.now() + ASSET_IMAGE_RESULT_TIMEOUT;
  let lastReadError = null;
  while (Date.now() < deadline) {
    try {
      if (taskId) {
        const task = resultData(await getBookAssetImageGenerationTask(batchId, bookId, taskId)) || {};
        if (task.status === 'failed') {
          const terminalError = new Error(task.error || '资产图片生成失败');
          terminalError.assetImageTaskTerminal = true;
          throw terminalError;
        }
        // A batch can contain several assets.  If the provider returns some
        // images but one returned URL cannot be downloaded, the backend marks
        // the task as succeeded with an error and persists the images that did
        // arrive.  Do not keep hiding those successful images while waiting
        // forever for the failed asset.
        if (task.status === 'succeeded' && task.error) {
          const partial = await listBookAssetImagesForAssets(batchId, bookId, wanted);
          if (partial.some(item => item.images.length > 0)) return partial;
          const terminalError = new Error(task.error || '资产图片生成失败');
          terminalError.assetImageTaskTerminal = true;
          throw terminalError;
        }
      }
      const response = await listBookAssetImagesForAssets(batchId, bookId, wanted);
      const ready = response.filter(item => item.images.length > 0);
      if (ready.length === wanted.length) return response;
      lastReadError = null;
    } catch (error) {
      if (error?.assetImageTaskTerminal) throw error;
      if ([401, 403].includes(Number(error?.status))) throw error;
      lastReadError = error;
    }
    await new Promise(resolve => window.setTimeout(resolve, ASSET_IMAGE_RESULT_POLL_INTERVAL));
  }
  throw lastReadError || new Error('资产图生成超过 15 分钟，已停止等待；请刷新资产设置后重试。');
}

export async function listBookAssetImagesForAssets(batchId, bookId, assetIds) {
  return Promise.all(assetIds.map(async assetId => ({
    assetId,
    images: resultData(await listBookAssetImages(batchId, bookId, assetId), 'images') || []
  })));
}

export function sameSettingsPatch(left, right) {
  return JSON.stringify(left || {}) === JSON.stringify(right || {});
}

export function h3DirectorCards(book) {
  const document = h3DirectorDocument(book);
  return Array.isArray(document.director_cards) ? document.director_cards : Array.isArray(document.directorCards) ? document.directorCards : [];
}

export function h3DirectorDocument(book) {
  const output = book?.directorRevision?.output || {};
  return output.h3_director || output.h3Director || {};
}

export function h3Text(value) { return String(value || '').trim(); }

export function h3NumberText(value) {
  const number = Number(value);
  return Number.isFinite(number) ? String(number) : '';
}

export function h3Value(object, snakeKey, camelKey = '') {
  return object?.[snakeKey] ?? (camelKey ? object?.[camelKey] : undefined);
}

export function h3CardCharacterNames(document, slotIDs = []) {
  const roster = h3Value(document, 'character_roster', 'characterRoster');
  const bySlot = new Map((Array.isArray(roster) ? roster : []).map(character => [h3Text(h3Value(character, 'slot_id', 'slotId')), character]));
  return (Array.isArray(slotIDs) ? slotIDs : []).map(slotID => {
    const character = bySlot.get(h3Text(slotID));
    const name = h3Text(h3Value(character, 'canonical_name', 'canonicalName'));
    return name ? `${slotID}(${name})` : h3Text(slotID);
  }).filter(Boolean).join('、');
}

export function h3AudioText(audio = {}, document = {}) {
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

export function formatH3DirectorCardPrompt(document = {}, card = {}) {
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

export function finalVideoPromptTemplate(body) {
  const value = String(body || '');
  const marker = '【批量工厂最终 Prompt 模板】';
  const index = value.indexOf(marker);
  return index >= 0 ? value.slice(index + marker.length).trim() : '';
}

export function requestID(prefix) { return `${prefix}-${globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`}`; }

export function stageActionKey(stage, bookId) { return `stage-${stage}:${String(bookId || '')}`; }

export function capability(caps, key) { return caps?.[key] || { available: false, reason: '正在读取 V12 服务能力' }; }

export function assetName(asset) { return String(asset?.name || asset?.label || asset?.id || '未命名预设'); }

export function assetFormalName(asset) { return String(asset?.formalName || asset?.formal_name || '').trim(); }

export function assetDisplayLabel(asset) {
  const name = assetName(asset);
  const formalName = assetFormalName(asset);
  return formalName && formalName !== name ? `${name}（${formalName}）` : name;
}

export function assetIdentitySummary(asset) {
  return [asset?.facts, asset?.relationships].map(value => String(value || '').trim()).filter(Boolean).join(' · ');
}

export function assetPrompt(asset) { return String(asset?.prompt || asset?.visualPrompt || asset?.description || ''); }

export function visualFactsWithoutRelationships(asset) {
  const facts = String(asset?.facts || '').trim();
  if (!facts) return '';
  const relationships = String(asset?.relationships || '').trim();
  const relationshipText = relationships.replace(/[；;。.,，]/g, ' ').trim();
  const markers = ['妻子', '丈夫', '老公', '老婆', '配偶', '夫妻', '妻婿', '孙媳', '孙子', '孙女', '祖父', '祖母', '爷爷', '奶奶', '父亲', '母亲', '爸爸', '妈妈', '儿子', '女儿', '兄弟', '姐妹', '长辈', '晚辈', '夫家', '主仆', '旧爱', '旧相识', '恋人', '爱人', '情感牵扯', '女主角的', '男主角的'];
  return facts
    .split(/[，,；;。]/)
    .map(value => value.trim())
    .filter(Boolean)
    .filter(value => {
      if (relationshipText && (relationshipText.includes(value) || value.includes(relationshipText))) return false;
      return !markers.some(marker => value.includes(marker));
    })
    .join('，');
}

export function mergedAssetVisualPrompt(asset) {
  const structured = [visualFactsWithoutRelationships(asset), asset?.appearance]
    .map(value => String(value || '').trim())
    .filter(Boolean);
  return (structured.length ? structured.join('，') : assetPrompt(asset)).trim();
}

export function savedDraftKey(type, asset) { return `asset:${type}:${asset?.id || assetName(asset)}`; }

export function effectiveBookSettings(batch, book) {
  const batchPatch = batch?.settingsState?.patch || {};
  const bookPatch = book?.settingsState?.patch || {};
  const effective = {
    ...batchPatch,
    ...bookPatch,
    publishSettings: { ...(batchPatch.publishSettings || {}), ...(bookPatch.publishSettings || {}) },
    comicPublishSettings: { ...(batchPatch.comicPublishSettings || {}), ...(bookPatch.comicPublishSettings || {}) }
  };
  const comic = effective?.aiPromptConfig?.comic || {};
  if (!['comic', 'video'].includes(effective.productionMode) && comic.enabled === true && String(comic.body || '').trim()) effective.productionMode = 'comic';
  return normalizeAudioSourceSettings(effective);
}

export function baseSetupEnabled(settings = {}) {
  return settings?.aiPromptConfig?.constraints?.baseSetup?.enabled !== false;
}

export function h3CompilationStatus(settings = {}) {
  if (audioPlanningRequired(settings)) return '等待真实配音时长编译最终 VIDEO';
  const seconds = Number(settings.storyboardDurationLimit) === 15 ? 15 : 10;
  return `将按 ${seconds} 秒确定性切段并编译最终 VIDEO`;
}

export function h3ConstraintPreview(settings = {}) {
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

export function storyboardAssetPromptText(book = {}, video = null) {
  const selectedIds = video
    ? readStoryboardAssetSelection(book, video).selectedIds
    : null;
  return (Array.isArray(book?.assetRecords) ? book.assetRecords : [])
    .filter(asset => !selectedIds || selectedIds.has(String(asset?.id || '')))
    .map(asset => {
      // The merged visual prompt is facts + appearance. Relationships remain
      // structured story metadata and must never enter the visual prompt;
      // older records may still have a polluted prompt, so rebuild it from
      // the structured visual fields when they are available.
      const prompt = mergedAssetVisualPrompt(asset);
      if (!prompt) return '';
      return `${assetKindLabel(asset?.kind)}「@${assetName(asset)}」\n资产提示词：${prompt}`;
    })
    .filter(Boolean)
    .join('\n\n');
}

export function composeStoryboardDisplayPrompt(prompt, assetText) {
  const body = String(prompt || '').trim();
  const assets = String(assetText || '').trim();
  if (!assets || body.includes('【基础设定】') || body.includes('基础设定：')) return body;
  return `【基础设定】\n${assets}${body ? `\n\n${body}` : ''}`;
}

export function assetKindLabel(kind) {
  return kind === 'character' ? '人物' : kind === 'scene' ? '场景' : kind === 'prop' ? '道具' : '资产';
}

export function promptPreview(value, limit = 180) {
  const text = String(value || '').replace(/\s+/g, ' ').trim();
  return text.length > limit ? `${text.slice(0, limit)}…` : text;
}

export function selectedVideoPromptLabel(settings = {}) {
  const video = settings?.aiPromptConfig?.video || {};
  const id = String(video.presetId || video.presetKey || '').trim();
  if (!id || id === 'batch-video-meta') return 'SD 视频提示词';
  return String(video.presetName || id).trim() || 'SD 视频提示词';
}

export function usesH3VideoPrompt(settings = {}) {
  const video = settings?.aiPromptConfig?.video || {};
  const id = String(video.presetId || video.presetKey || '').trim();
  return id === 'batch-video-h3-director' || id === 'h3-video-normal';
}

export function effectiveBookAssetRules(batch, book) {
  return {
    ...(batch?.settingsState?.patch?.aiPromptConfig?.assets || {}),
    ...(book?.settingsState?.patch?.aiPromptConfig?.assets || {})
  };
}

export function effectiveBookComicRules(batch, book) {
  return {
    ...(batch?.settingsState?.patch?.aiPromptConfig?.comic || {}),
    ...(book?.settingsState?.patch?.aiPromptConfig?.comic || {})
  };
}

export function selectedPublishProfile(settings = {}) {
  const wanted = String(settings.websiteProfileId || '').trim();
  return (Array.isArray(settings.websiteProfiles) ? settings.websiteProfiles : [])
    .find(profile => String(profile?.id || '').trim() === wanted) || {};
}

export function decompressionCountForUpload(batch, book) {
  const settings = effectiveBookSettings(batch, book).publishSettings || {};
  const profile = selectedPublishProfile(settings);
  const candidate = settings?.advanced?.jieyaNum ?? profile?.advanced?.jieyaNum;
  const numeric = Number(candidate);
  return Number.isFinite(numeric) ? Math.max(0, Math.round(numeric)) : 4;
}

export function isBookUploadedTo121(book) {
  return String(book?.sourceMetadata?.websiteSubmitStatus || '').trim().toLowerCase() === 'uploaded';
}

export function uploadProgressForBook(book) {
  const value = book?.sourceMetadata?.websiteSubmitProgress;
  return value && typeof value === 'object' ? value : null;
}

export function sourcePlatformIdForUpload(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const candidates = [book?.platform, metadata.platformId, metadata.platform_id, metadata.platformName, metadata.platformLabel];
  const labels = { '黑岩付费': '1', '番茄付费': '2', '七猫付费': '3', '点众付费': '4', '阅文付费': '6', '番茄免费': '7', '知乎付费': '15', '掌阅付费': '20', '卓越付费': '26', '九州书城': '29', '掌文付费': '31' };
  return candidates.map(value => String(value || '').trim()).find(value => /^(1|2|3|4|6|7|15|20|26|29|31)$/.test(value))
    || candidates.map(value => labels[String(value || '').trim()]).find(Boolean)
    || '';
}

export function completePublishMapping(_settings = {}, book = {}) {
  // 121 platform derives from the source book. Gender/style/tags are per-book
  // classification data and the server fills missing values immediately before
  // submit; a batch profile may only supply presentation/advanced options.
  return Boolean(sourcePlatformIdForUpload(book));
}

export function hasClassificationValues(metadata = {}) {
  const gender = String(metadata?.gender || metadata?.gender_name || '').trim();
  const style = String(metadata?.style || metadata?.styleName || metadata?.style_name || '').trim();
  return Boolean(gender && gender !== '待确认' && style && !['待判定', '未判定'].includes(style));
}

export function classificationLabel(metadata = {}) {
  const status = String(metadata?.classifyStatus || '').trim();
  if (status === 'manual') return '已手调';
  if (status === 'classified') return metadata?.classifyNeedsReview ? '已自动分类' : '已识别';
  if (status === 'failed' || status === 'pending_confirmation') return hasClassificationValues(metadata) ? '已自动分类' : '待识别';
  return '待识别';
}

export function classificationDetail(metadata = {}) {
  const status = String(metadata?.classifyStatus || '').trim();
  if (status === 'classified' && metadata?.classifyNeedsReview) return String(metadata?.classifyReason || '原文已获取，系统已自动确定男女频和风格，可在单书配置中修改。').trim();
  if (status === 'classified') return String(metadata?.classifyReason || '已根据当前书正文保存识别结果。').trim();
  if (status === 'failed' || status === 'pending_confirmation') {
    if (hasClassificationValues(metadata)) return '原文已获取，系统已自动确定男女频和风格，可在单书配置中修改。';
    return `原文已获取，正在自动重试分类。${String(metadata?.classifyNote || metadata?.classifyError || '').trim()}`.trim();
  }
  return '尚未发起识别。';
}

export function completedMediaVersions(status) {
	const entries = new Map();
	for (const job of status?.jobs || []) for (const task of job?.tasks || []) {
		if (task?.status !== 'succeeded' || !String(task?.mediaUrl || '').trim()) continue;
		const versions = entries.get(task.videoId) || [];
		versions.push({ ...task, productionJobId: job.id });
		entries.set(task.videoId, versions);
	}
	return entries;
}

export function primaryMediaVersion(video, versions) {
	const primaryID = video?.settingsState?.patch?.primaryMediaTaskId;
	return versions.find(item => item.id === primaryID) || versions.at(-1) || null;
}

export function storyboardVideoLabel(video, index) {
  const ordinal = Number(index) + 1;
  const label = String(video?.label || video?.title || '').trim();
  return label ? `分镜 ${ordinal} · ${label}` : `分镜 ${ordinal}`;
}

export function bookMergeJobs(status, bookId) {
  return (status?.jobs || []).filter(job => job?.bookId === bookId).sort((left, right) => String(left?.updatedAt || left?.createdAt || '').localeCompare(String(right?.updatedAt || right?.createdAt || '')));
}

export function mergeJobSpeed(job) {
  const speed = Number(job?.speed || 0);
  return speed > 0 ? speed : 1;
}

export function defaultBookMerge(jobs) {
  const completed = jobs.filter(job => job?.status === 'succeeded' && String(job?.outputUrl || '').trim());
  // The newest successful output is the only default upload source. Older
  // speed variants remain history and must never silently become current.
  return completed.at(-1) || null;
}

export function latestBookMerge(status, bookId) {
  const jobs = bookMergeJobs(status, bookId);
  return [...jobs].reverse().find(job => job?.status === 'running' || job?.status === 'queued') || defaultBookMerge(jobs) || jobs.at(-1) || null;
}

export function bookVideoReady(book, productionStatus) {
  const videoIds = (book?.videos || []).map(video => String(video?.id || '')).filter(Boolean);
  if (!videoIds.length) return false;
  const completed = new Set((productionStatus?.jobs || [])
    .filter(job => job?.bookId === book?.id)
    .flatMap(job => job?.tasks || [])
    .filter(task => task?.status === 'succeeded' && String(task?.mediaUrl || '').trim())
    .map(task => String(task?.videoId || '')));
  return videoIds.every(videoId => completed.has(videoId));
}

export function readImageAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('读取图片失败'));
    reader.onload = () => resolve(String(reader.result || ''));
    reader.readAsDataURL(file);
  });
}

export function storyboardAssetNamesFromPrompt(prompt, assets) {
  const text = String(prompt || '').trim();
  if (!text) return [];
  return (assets || []).map(asset => String(asset?.name || '').trim()).filter(name => name && text.includes(name));
}

export function storyboardAssetDefaults(book, videoId) {
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

export function readStoryboardAssetSelection(book, video) {
  const defaults = storyboardAssetDefaults(book, video?.id);
  const raw = video?.settingsState?.patch?.assetSelection;
  const autoAssetIds = Array.isArray(raw?.autoAssetIds) ? raw.autoAssetIds.map(String).filter(Boolean) : defaults;
  const addedAssetIds = Array.isArray(raw?.addedAssetIds) ? raw.addedAssetIds.map(String).filter(Boolean) : [];
  const excludedAssetIds = Array.isArray(raw?.excludedAssetIds) ? raw.excludedAssetIds.map(String).filter(Boolean) : [];
  const excluded = new Set(excludedAssetIds);
  return { autoAssetIds: [...new Set(autoAssetIds)], addedAssetIds: [...new Set(addedAssetIds)], excludedAssetIds: [...new Set(excludedAssetIds)], candidateIds: [...new Set([...autoAssetIds, ...addedAssetIds, ...excludedAssetIds])], selectedIds: new Set([...autoAssetIds, ...addedAssetIds].filter(id => !excluded.has(id))) };
}

export function storyboardAssetMentions(book, video) {
  const selectedIds = readStoryboardAssetSelection(book, video).selectedIds;
  return (book?.assetRecords || [])
    .filter(asset => selectedIds.has(String(asset?.id || '')))
    .map(asset => `@${assetName(asset)}`)
    .filter(Boolean)
    .join(' ');
}

export async function loadBookAssetDisplayURL(imageUrl) {
  const source = String(imageUrl || '').trim();
  if (!source) throw new Error('asset image URL is empty');
  if (source.startsWith('blob:')) return { url: source, owned: false };
  try {
    const blob = await getBookAssetImageBlob(source);
    return { url: URL.createObjectURL(blob), owned: true };
  } catch (error) {
    if (/^https?:\/\//i.test(source)) return { url: source, owned: false };
    throw error;
  }
}

export function comicPageReferencePreviewURL(referenceURL) {
  const matched = String(referenceURL || '').match(/^\/api\/local-executor\/v1\/batch-comic-references\/([^/]+)\/([^/]+)$/);
  return matched ? `/api/batch-factory/v12/batches/${matched[1]}/comic-references/${matched[2]}/content` : '';
}

export function fileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(new Error('参考图读取失败'));
    reader.readAsDataURL(file);
  });
}

export function stripStandaloneAssetMentionLine(prompt, book) {
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

export function withStoryboardAssetMentions(prompt, book, video) {
  const text = String(prompt || '').trim();
  // 展示层只负责把正文中已有的资产名称渲染成标签；不能额外追加一行
  // “资产：@...” ，否则卡片和编辑器会与最终 VIDEO Prompt 重复显示。
  return stripStandaloneAssetMentionLine(text, book);
}

export function compactUnifiedAssetDescriptions(prompt, book) {
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

export function separateStoryboardAssetMentions(text, book) {
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

export function collapseRepeatedStoryboardAssetMentions(text, book) {
  const value = String(text || '');
  const names = (book?.assetRecords || [])
    .filter(asset => asset?.kind === 'scene' || asset?.kind === 'prop')
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length);
  if (!value || !names.length) return value;
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
  const pattern = new RegExp(`(${names.map(name => `@${name.replace(/[.*+?^${}()|[\\]\\]/g, '\\$&')}`).join('|')})`, 'g');
  let previousMention = '';
  assetBlock = assetBlock.replace(pattern, mention => {
    if (mention === previousMention) return '';
    previousMention = mention;
    return mention;
  });
  return `${value.slice(0, start)}${assetBlock.replace(/([：:])[\s、，,；;]+/g, '$1')}${value.slice(end)}`;
}

export function escapePromptHTML(value) {
  return String(value ?? '').replace(/[&<>"']/g, character => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[character]));
}

export function renderStoryboardPromptHTML(prompt, book, video, assetImages = {}) {
  let text = compactUnifiedAssetDescriptions(withStoryboardAssetMentions(prompt, book, video), book);
  (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      text = text.replace(new RegExp(`(^|[^@])${escaped}`, 'g'), `$1@${name}`);
    });
  text = collapseRepeatedStoryboardAssetMentions(text, book);
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
    return `<span class="batch-factory-asset-mention is-${escapePromptHTML(kind)}" contenteditable="false" data-asset-mention="${escapePromptHTML(asset.mention)}" data-asset-name="${escapePromptHTML(assetName(asset))}" title="${escapePromptHTML(`${kind}：${assetDisplayLabel(asset)}${assetIdentitySummary(asset) ? ` · ${assetIdentitySummary(asset)}` : ''}`)}"><span class="batch-factory-asset-mention-avatar">${avatar}</span>${escapePromptHTML(assetDisplayLabel(asset))}</span>`;
  }).join('');
}

export function contentEditableTextOffset(root) {
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

export function contentEditablePromptText(root) {
  return readStoryboardPromptEditableText(root);
}

export function setContentEditableTextOffset(root, offset) {
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

export function batchFactoryPassiveMediaUrl(media) {
  const value = String(media?.mediaUrl || media?.outputUrl || '').trim();
  // Merge jobs created by the local Go worker may still contain the
  // historical v11 path. The public workbench is mounted on the v12
  // contract; normalize the relative URL before the browser requests it.
  return value.replace(/^\/api\/batch-factory\/v11\//, '/api/batch-factory/v12/');
}

export function batchFactoryPassivePosterUrl(media) {
  return String(
    media?.posterUrl
    || media?.thumbnailUrl
    || media?.coverUrl
    || media?.previewImageUrl
    || media?.firstFrameUrl
    || ''
  ).trim();
}

export function batchFactoryParseMediaRatio(value, fallback = 9 / 16) {
  const parts = String(value || '').trim().split(/[:/x×]/).map(Number).filter(Number.isFinite);
  if (parts.length >= 2 && parts[0] > 0 && parts[1] > 0) return parts[0] / parts[1];
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric > 0 ? numeric : fallback;
}

export function batchFactoryMediaOrientation(ratio) {
  const safe = Number.isFinite(Number(ratio)) && Number(ratio) > 0 ? Number(ratio) : 9 / 16;
  if (Math.abs(safe - 1) < 0.035) return 'square';
  return safe > 1 ? 'landscape' : 'portrait';
}

export function batchFactoryRatioText(width, height, fallbackRatio = 9 / 16) {
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

export function batchFactoryStableMediaLabel(prefix, id) {
  const compact = String(id || '').replace(/[^a-z0-9]/gi, '').slice(-6).toUpperCase();
  return compact ? `${prefix}-${compact}` : prefix;
}

export function localStartTime() {
  const now = new Date();
  return new Date(now.getTime() - now.getTimezoneOffset() * 60_000).toISOString().slice(0, 19);
}
