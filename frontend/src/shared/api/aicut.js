import { apiRequest as baseApiRequest } from './client.js';
import { scopedAICutOptions } from './aicutTimelineScope.js';

const apiRequest = (path, options = {}) => baseApiRequest(path, scopedAICutOptions(path, options));

function projectPath(projectId, suffix = '') {
  const id = String(projectId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('AI-CUT 项目无效');
  return `/api/ai-cut/projects/${id}${suffix}`;
}

export async function ensureAICutCreativeCanvas(projectId, request = apiRequest) {
  return request(projectPath(projectId, '/creative-canvas'), { method: 'POST' });
}

export async function transferAICutCreativeAssets(projectId, assetIds, request = apiRequest) {
  const ids = [...new Set((Array.isArray(assetIds) ? assetIds : []).map(id => String(id || '').trim()).filter(Boolean))];
  if (!ids.length || ids.length > 30) throw new Error('每次请选择 1–30 个素材');
  return request(projectPath(projectId, '/creative-canvas/assets'), {
    method: 'POST', body: JSON.stringify({ assetIds: ids })
  });
}

export async function transferAICutCreativeScript(projectId, nodeIds, expectedRevision, request = apiRequest) {
  const ids = [...new Set((Array.isArray(nodeIds) ? nodeIds : []).map(id => String(id || '').trim()).filter(Boolean))];
  if (!ids.length || ids.length > 10) throw new Error('每次请选择 1–10 个故事或分镜节点');
  if (!Number.isSafeInteger(expectedRevision) || expectedRevision < 0) throw new Error('当前时间线版本无效，请刷新后重试');
  return request(projectPath(projectId, '/creative-canvas/script'), {
    method: 'POST', body: JSON.stringify({ nodeIds: ids, expectedRevision })
  });
}

export async function listAICutProjects(request = apiRequest) {
  const response = await request('/api/ai-cut/projects');
  return Array.isArray(response?.projects) ? response.projects : [];
}

export async function createAICutProject({ title, aspectRatio = '9:16' }, request = apiRequest) {
  const response = await request('/api/ai-cut/projects', {
    method: 'POST',
    body: JSON.stringify({ title: String(title || '').trim(), aspectRatio })
  });
  return response?.project;
}

export async function updateAICutProject(projectId, value = {}, request = apiRequest) {
  const body = {};
  if (Object.hasOwn(value, 'aspectRatio')) {
    if (!['9:16', '16:9', '1:1'].includes(value.aspectRatio)) throw new Error('画面比例仅支持 9:16、16:9 或 1:1');
    body.aspectRatio = value.aspectRatio;
  }
  if (Object.hasOwn(value, 'title')) {
    const title = String(value.title || '').trim();
    if (!title) throw new Error('项目名称不能为空');
    body.title = title;
  }
  if (Object.hasOwn(value, 'backgroundMode')) {
    const backgroundMode = String(value.backgroundMode || '').trim().toLowerCase();
    if (!['color', 'transparent'].includes(backgroundMode)) throw new Error('画布背景模式仅支持背景色或透明');
    body.backgroundMode = backgroundMode;
  }
  if (Object.hasOwn(value, 'backgroundColor')) {
    const backgroundColor = String(value.backgroundColor || '').trim().toUpperCase();
    if (!/^#[0-9A-F]{6}$/.test(backgroundColor)) throw new Error('画布背景颜色必须为 #RRGGBB');
    body.backgroundColor = backgroundColor;
  }
  if (!Object.keys(body).length) throw new Error('没有可更新的项目设置');
  const response = await request(projectPath(projectId), {
    method: 'PATCH',
    body: JSON.stringify(body)
  });
  return response?.project;
}

export async function loadAICutTimeline(projectId, request = apiRequest) {
  const response = await request(projectPath(projectId, '/timeline'));
  return response?.timeline || { revision: 0, tracks: [] };
}

export async function saveAICutTimeline(projectId, timeline, request = apiRequest) {
  const current = timeline || { revision: 0, tracks: [] };
  const response = await request(projectPath(projectId, '/timeline'), {
    method: 'PUT',
    body: JSON.stringify({ expectedRevision: Number(current.revision || 0), timeline: current })
  });
  return response?.timeline;
}

export async function getAICutTimelineHistory(projectId, request = apiRequest) {
  const response = await request(projectPath(projectId, '/timeline/history'));
  return response?.history || { canUndo: false, canRedo: false };
}

async function moveAICutTimelineHistory(projectId, action, expectedRevision, request) {
  return request(projectPath(projectId, `/timeline/history/${action}`), {
    method: 'POST',
    body: JSON.stringify({ expectedRevision: Number(expectedRevision || 0) })
  });
}

export async function undoAICutTimeline(projectId, expectedRevision, request = apiRequest) {
  return moveAICutTimelineHistory(projectId, 'undo', expectedRevision, request);
}

export async function redoAICutTimeline(projectId, expectedRevision, request = apiRequest) {
  return moveAICutTimelineHistory(projectId, 'redo', expectedRevision, request);
}

export async function listAICutAssets(projectId, request = apiRequest) {
  const response = await request(projectPath(projectId, '/assets'));
  return Array.isArray(response?.assets) ? response.assets : [];
}

export async function uploadAICutAsset(projectId, file, request = apiRequest) {
  if (!(file instanceof Blob)) throw new Error('请选择素材文件');
  const body = new FormData();
  body.append('file', file, file.name || 'asset.bin');
  const response = await request(projectPath(projectId, '/assets'), { method: 'POST', body, timeoutMs: 0 });
  return response?.asset;
}

export function normalizeAICutTTSInput(input = {}) {
  const text = String(input.text || '').trim();
  const speed = input.speed === undefined || input.speed === null || input.speed === '' ? 1.8 : Number(input.speed);
  const pitch = input.pitch === undefined || input.pitch === null || input.pitch === '' ? 10 : Number(input.pitch);
  if (!text) throw new Error('配音文本不能为空');
  if (Array.from(text).length > 5000) throw new Error('配音文本不能超过 5000 个字符');
  if (!Number.isFinite(speed) || speed < .5 || speed > 2) throw new Error('语速必须在 0.5 到 2 之间');
  if (!Number.isFinite(pitch) || pitch < -50 || pitch > 50) throw new Error('音调必须在 -50 到 50 之间');
  return {
    text,
    name: String(input.name || '').trim(),
    voice: String(input.voice || '').trim() || 'zh-CN-XiaoxiaoNeural',
    speed,
    pitch,
    style: String(input.style || '').trim() || 'general'
  };
}

export async function generateAICutSpeech(projectId, input, request = apiRequest) {
  const body = normalizeAICutTTSInput(input);
  const response = await request(projectPath(projectId, '/tts'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify(body)
  });
  return response?.asset;
}

export function normalizeAICutVoiceCloneInput(input = {}) {
  const assetId = String(input.assetId || '').trim();
  const text = String(input.text || '').trim();
  const speed = input.speed === undefined || input.speed === null || input.speed === '' ? 1 : Number(input.speed);
  const pitch = input.pitch === undefined || input.pitch === null || input.pitch === '' ? 0 : Number(input.pitch);
  if (!assetId) throw new Error('请选择参考音频');
  if (!text) throw new Error('克隆音色试听文本不能为空');
  if (Array.from(text).length > 1000) throw new Error('克隆音色试听文本不能超过 1000 个字符');
  if (!Number.isFinite(speed) || speed < .5 || speed > 2) throw new Error('语速必须在 0.5 到 2 之间');
  if (!Number.isFinite(pitch) || pitch < -50 || pitch > 50) throw new Error('音调必须在 -50 到 50 之间');
  return { assetId, text, name: String(input.name || '').trim(), speed, pitch, style: String(input.style || '').trim() || 'general' };
}

export async function cloneAICutVoice(projectId, input, request = apiRequest) {
  const body = normalizeAICutVoiceCloneInput(input);
  const response = await request(projectPath(projectId, '/voice-clone'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify(body)
  });
  return response?.asset;
}

export async function generateAICutProxy(projectId, assetId, request = apiRequest) {
  const id = String(assetId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('AI-CUT 素材无效');
  const response = await request(projectPath(projectId, `/assets/${id}/proxy`), { method: 'POST', timeoutMs: 0 });
  return response?.asset;
}

export function normalizeAICutConsolidateInput(input = {}) {
  const clipIds = [...new Set((Array.isArray(input.clipIds) ? input.clipIds : []).map((value) => String(value || '').trim()).filter(Boolean))];
  const expectedRevision = Number(input.expectedRevision);
  if (clipIds.length < 2 || clipIds.length > 50) throw new Error('请选择 2 到 50 个视频片段进行合并');
  if (!Number.isInteger(expectedRevision) || expectedRevision < 0) throw new Error('时间线版本无效，请刷新后重试');
  return { clipIds, expectedRevision };
}

export async function consolidateAICutClips(projectId, input = {}, request = apiRequest) {
  const body = normalizeAICutConsolidateInput(input);
  const response = await request(projectPath(projectId, '/consolidate'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify(body)
  });
  return response?.asset;
}

export async function removeAICutImageBackground(projectId, assetId, input = {}, request = apiRequest) {
  const id = String(assetId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('AI-CUT 素材无效');
  const response = await request(projectPath(projectId, `/assets/${id}/remove-background`), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({ name: String(input?.name || '').trim() })
  });
  return response?.asset;
}

export async function importAICutSRT(projectId, file, expectedRevision, request = apiRequest) {
  if (!file) return null;
  const filename = String(file.name || 'subtitles.srt');
  const srt = typeof file.text === 'function' ? await file.text() : String(file.text || '');
  if (new TextEncoder().encode(srt).byteLength > 2 * 1024 * 1024) throw new Error('字幕文件不能超过 2 MB');
  const response = await request(projectPath(projectId, '/subtitles/import'), {
    method: 'POST',
    body: JSON.stringify({ filename, srt, expectedRevision: Number(expectedRevision || 0) })
  });
  return response || null;
}

export async function correctAICutCaptions(projectId, input, request = apiRequest) {
  const script = String(input?.script || '').trim();
  if (!script) throw new Error('校对剧本不能为空');
  return request(projectPath(projectId, '/captions/correct'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({
      script,
      modelId: String(input?.modelId || '').trim(),
      expectedRevision: Number(input?.expectedRevision || 0)
    })
  });
}

export async function trimAICutFillers(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/filler-trim'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({
      modelId: String(input?.modelId || '').trim(),
      expectedRevision: Number(input?.expectedRevision || 0)
    })
  });
}

export async function analyzeAICutCaptionCorrections(projectId, input, request = apiRequest) {
  const script = String(input?.script || '').trim();
  if (!script) throw new Error('校对剧本不能为空');
  return request(projectPath(projectId, '/captions/correct'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({
      script,
      modelId: String(input?.modelId || '').trim(),
      expectedRevision: Number(input?.expectedRevision || 0),
      preview: true
    })
  });
}

export async function applyAICutCaptionCorrections(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/captions/correct'), {
    method: 'POST',
    signal: input?.signal,
    body: JSON.stringify({
      expectedRevision: Number(input?.expectedRevision || 0),
      corrections: input?.corrections ?? []
    })
  });
}

export async function analyzeAICutFillerCandidates(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/filler-trim'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({
      modelId: String(input?.modelId || '').trim(),
      expectedRevision: Number(input?.expectedRevision || 0),
      preview: true
    })
  });
}

export async function applyAICutFillerCandidates(projectId, input, request = apiRequest) {
  return request(projectPath(projectId, '/filler-trim'), {
    method: 'POST',
    signal: input?.signal,
    body: JSON.stringify({
      expectedRevision: Number(input?.expectedRevision || 0),
      fillers: input?.fillers ?? []
    })
  });
}

export async function cleanAICutOralTimeline(projectId, input, request = apiRequest) {
  const maxPauseMs = Number(input?.maxPauseMs || 1000);
  if (!Number.isFinite(maxPauseMs) || maxPauseMs < 100 || maxPauseMs > 10000) throw new Error('停顿上限必须在 0.1 到 10 秒之间');
  return request(projectPath(projectId, '/oral-clean'), {
    method: 'POST',
    signal: input?.signal,
    timeoutMs: 0,
    body: JSON.stringify({
      modelId: String(input?.modelId || '').trim(),
      maxPauseMs: Math.round(maxPauseMs),
      expectedRevision: Number(input?.expectedRevision || 0)
    })
  });
}

export async function createAICutASRJob(projectId, input, request = apiRequest) {
  const assetId = String(input?.assetId || '').trim();
  const modelId = String(input?.modelId || '').trim();
  if (!assetId || !modelId) throw new Error('请选择语音素材和 ASR 模型');
  const response = await request(projectPath(projectId, '/asr-jobs'), {
    method: 'POST',
    body: JSON.stringify({ assetId, modelId, language: String(input?.language || '').trim(), expectedRevision: Number(input?.expectedRevision || 0) })
  });
  return response?.job;
}

export async function listAICutASRJobs(projectId, request = apiRequest) {
  const response = await request(projectPath(projectId, '/asr-jobs'), { suppressGlobalError: true });
  return Array.isArray(response?.jobs) ? response.jobs : [];
}

export async function cancelAICutASRJob(projectId, jobId, request = apiRequest) {
  const id = String(jobId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('语音识别任务无效');
  const response = await request(projectPath(projectId, `/asr-jobs/${id}/cancel`), { method: 'POST', body: '{}' });
  return response?.job;
}

export async function exportAICutMedia(projectId, format, expectedRevision, request = apiRequest) {
  const normalized = String(format || '').trim().toLowerCase();
  if (!['mp4', 'webm', 'mp3'].includes(normalized)) throw new Error('仅支持 MP4、WebM 或 MP3 导出');
  return request(projectPath(projectId, `/exports/${normalized}`), {
    method: 'POST',
    responseType: 'blob',
    timeoutMs: 0,
    body: JSON.stringify({ expectedRevision: Number(expectedRevision || 0) })
  });
}

function exportJobPath(projectId, jobId, suffix = '') {
  const id = String(jobId || '').trim();
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('导出任务无效');
  return projectPath(projectId, `/export-jobs/${id}${suffix}`);
}

export function normalizeAICutExportSettings(value = {}, format = 'mp4') {
  const resolution = String(value.resolution || '1080p').toLowerCase();
  const frameRate = Number(value.frameRate || 30);
  const quality = String(value.quality || 'standard').toLowerCase();
  const normalizedFormat = String(format || 'mp4').toLowerCase();
  const codec = String(value.codec || (normalizedFormat === 'mov' ? 'prores4444' : normalizedFormat === 'webm' ? 'vp9' : 'h264')).toLowerCase();
  const bitDepth = Number(value.bitDepth || (codec === 'prores4444' ? 12 : 8));
  const bitrateMbps = Number(value.bitrateMbps || 0);
  const alpha = Boolean(value.alpha);
  const hardwareAcceleration = String(value.hardwareAcceleration || 'auto').toLowerCase();
  if (!['720p', '1080p', '4k'].includes(resolution)) throw new Error('导出分辨率仅支持 720P、1080P 或 4K');
  if (![24, 25, 30, 50, 60].includes(frameRate)) throw new Error('导出帧率仅支持 24、25、30、50 或 60 fps');
  if (!['draft', 'standard', 'high'].includes(quality)) throw new Error('导出质量仅支持快速预览、标准或高质量');
  if (normalizedFormat === 'mp4' && !['h264', 'hevc', 'av1'].includes(codec)) throw new Error('MP4 编码器仅支持 H.264、HEVC 或 AV1');
  if (normalizedFormat === 'webm' && !['vp9', 'av1'].includes(codec)) throw new Error('WebM 编码器仅支持 VP9 或 AV1');
  if (normalizedFormat === 'mov' && codec !== 'prores4444') throw new Error('MOV 透明导出仅支持 ProRes 4444');
  if (![8, 10, 12].includes(bitDepth)) throw new Error('导出色深仅支持 8-bit、10-bit 或 ProRes 4444 的 12-bit');
  if (bitDepth === 12 && codec !== 'prores4444') throw new Error('12-bit 仅用于 ProRes 4444');
  if (codec === 'h264' && bitDepth === 10) throw new Error('H.264 暂不提供 10-bit，请选择 HEVC 或 AV1');
  if (alpha && !((normalizedFormat === 'webm' && codec === 'vp9') || (normalizedFormat === 'mov' && codec === 'prores4444'))) throw new Error('透明导出仅支持 WebM VP9 或 MOV ProRes 4444');
  if (!Number.isInteger(bitrateMbps) || bitrateMbps < 0 || bitrateMbps > 100) throw new Error('目标码率必须为自动或 1–100 Mbps');
  if (!['auto', 'software'].includes(hardwareAcceleration)) throw new Error('硬件加速仅支持自动选择或软件编码');
  return { resolution, frameRate, quality, codec, bitDepth, bitrateMbps, hardwareAcceleration, ...(alpha ? { alpha: true } : {}) };
}

export async function createAICutExportJob(projectId, format, expectedRevision, settingsOrRequest = {}, request = apiRequest) {
  const normalized = String(format || '').trim().toLowerCase();
  if (!['mp4', 'webm', 'mov', 'mp3'].includes(normalized)) throw new Error('仅支持 MP4、WebM、MOV 或 MP3 导出');
  const hasLegacyRequest = typeof settingsOrRequest === 'function';
  if (hasLegacyRequest) request = settingsOrRequest;
  const settings = hasLegacyRequest || normalized === 'mp3' ? null : normalizeAICutExportSettings(settingsOrRequest, normalized);
  const response = await request(projectPath(projectId, '/export-jobs'), {
    method: 'POST',
    body: JSON.stringify({ format: normalized, expectedRevision: Number(expectedRevision || 0), ...(settings || {}) })
  });
  return response?.job;
}

export async function getAICutExportJob(projectId, jobId, request = apiRequest) {
  const response = await request(exportJobPath(projectId, jobId), { suppressGlobalError: true });
  return response?.job;
}

export async function listAICutExportJobs(projectId, request = apiRequest) {
  const response = await request(projectPath(projectId, '/export-jobs'), { suppressGlobalError: true });
  return Array.isArray(response?.jobs) ? response.jobs : [];
}

export async function cancelAICutExportJob(projectId, jobId, request = apiRequest) {
  const response = await request(exportJobPath(projectId, jobId, '/cancel'), { method: 'POST', body: '{}' });
  return response?.job;
}

export async function downloadAICutExportJob(projectId, jobId, request = apiRequest) {
  return request(exportJobPath(projectId, jobId, '/download'), { responseType: 'blob', timeoutMs: 0 });
}

export async function smartSplitAICutVideo(projectId, input, request = apiRequest) {
  const assetId = String(input?.assetId || '').trim();
  if (!assetId) throw new Error('请选择视频素材');
  return request(projectPath(projectId, '/smart-split'), {
    method: 'POST',
    timeoutMs: 0,
    signal: input?.signal,
    body: JSON.stringify({ assetId, threshold: Number(input?.threshold || 0.4), minIntervalMs: Number(input?.minIntervalMs || 500), expectedRevision: Number(input?.expectedRevision || 0) })
  });
}

export async function extractAICutAudio(projectId, assetId, request = apiRequest) {
  const id = String(assetId || '').trim();
  if (!id) throw new Error('请选择视频素材');
  const response = await request(projectPath(projectId, '/extract-audio'), {
    method: 'POST',
    timeoutMs: 0,
    body: JSON.stringify({ assetId: id })
  });
  return response?.asset;
}

export async function extractAICutVisualFrames(projectId, assetId, count = 6, request = apiRequest) {
  const id = String(assetId || '').trim();
  const frameCount = Math.round(Number(count));
  if (!id) throw new Error('请选择视频或图片素材');
  if (!Number.isFinite(frameCount) || frameCount < 1 || frameCount > 8) throw new Error('抽帧数量必须在 1 到 8 之间');
  return request(projectPath(projectId, '/visual-frames'), {
    method: 'POST',
    timeoutMs: 0,
    body: JSON.stringify({ assetId: id, count: frameCount })
  });
}

export async function extractAICutTrackingFrames(projectId, input = {}, request = apiRequest) {
  const assetId = String(input.assetId || '').trim();
  const startMs = Math.round(Number(input.startMs));
  const durationMs = Math.round(Number(input.durationMs));
  const count = Math.round(Number(input.count));
  const width = Math.round(Number(input.width || 192));
  if (!assetId) throw new Error('请选择视频素材');
  if (!Number.isFinite(startMs) || startMs < 0 || !Number.isFinite(durationMs) || durationMs < 1 || durationMs > 300000) throw new Error('跟踪区间必须在 1 毫秒到 5 分钟之间');
  if (!Number.isFinite(count) || count < 2 || count > 60) throw new Error('跟踪帧数量必须在 2 到 60 之间');
  if (!Number.isFinite(width) || width < 96 || width > 320) throw new Error('跟踪帧宽度必须在 96 到 320 之间');
  return request(projectPath(projectId, '/tracking-frames'), {
    method: 'POST',
    timeoutMs: 0,
    signal: input.signal,
    body: JSON.stringify({ assetId, startMs, durationMs, count, width })
  });
}

export async function getAICutAudioAnalysis(projectId, assetId, points = 240, request = apiRequest) {
  const id = String(assetId || '').trim();
  const pointCount = Math.round(Number(points));
  if (!id) throw new Error('请选择音频或视频素材');
  if (!Number.isFinite(pointCount) || pointCount < 16 || pointCount > 1024) throw new Error('波形采样点必须在 16 到 1024 之间');
  const query = new URLSearchParams({ assetId: id, points: String(pointCount) });
  const response = await request(projectPath(projectId, `/audio-analysis?${query}`), { timeoutMs: 0 });
  return response?.analysis;
}

export async function alignAICutMultiCamAudio(projectId, input = {}, request = apiRequest) {
  const assetIds = input.assetIds;
  const referenceAssetId = input.referenceAssetId;
  const windowMs = input.windowMs ?? 30000;
  const maxLagMs = input.maxLagMs ?? 5000;
  if (!Array.isArray(assetIds) || assetIds.length < 2 || assetIds.length > 9 || assetIds.some(id => typeof id !== 'string' || !id.trim() || id !== id.trim()) || new Set(assetIds).size !== assetIds.length) throw new Error('请选择2到9个不同的自有视频机位');
  if (!assetIds.includes(referenceAssetId)) throw new Error('参考机位须在所选机位中');
  if (!Number.isSafeInteger(windowMs) || windowMs < 2000 || windowMs > 60000 || !Number.isSafeInteger(maxLagMs) || maxLagMs < 20 || maxLagMs > 10000) throw new Error('分析时长须为2000到60000毫秒，最大偏移须为20到10000毫秒');
  return request(projectPath(projectId, '/multicam-sync'), {
    method: 'POST', timeoutMs: 100000, signal: input.signal, suppressGlobalError: true,
    body: JSON.stringify({ assetIds, referenceAssetId, windowMs, maxLagMs }),
  });
}
