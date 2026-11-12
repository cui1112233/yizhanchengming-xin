import { aicutTextResourcePreset } from '../aicutText.js';
import { normalizeVisualEffect, visualEffectPreset } from '../aicutEffects.js';
import { normalizeAudioProcessing } from '../aicutAudioProcessing.js';
import { normalizeTransform } from './canvasTransform.js';

const MAX_DURATION_MS = 600000;
const MAX_END_MS = 86400000;
const layouts = {
  'template-short-drama': { labelZh: '短剧模板', description: '新增短剧画面与醒目标题，保留原工程内容。', text: 'text-outline-caption', animation: 'pop', effect: 'cool-cinema', scale: 1.04, denoise: .25, volume: .2 },
  'template-commentary': { labelZh: '解说模板', description: '新增原色画面与带底板的解说文字，可选自有配乐。', text: 'text-panel-caption', animation: 'none', effect: 'original', scale: 1, denoise: .35, volume: .18 },
  'template-interview': { labelZh: '访谈模板', description: '新增原色画面与人物信息条，可选自有音频。', text: 'text-lower-third', animation: 'fade', effect: 'original', scale: 1, denoise: .4, volume: .15 },
  'template-product': { labelZh: '商品模板', description: '新增明亮的商品画面与粗体标题，可选自有配乐。', text: 'text-bold-title', animation: 'float-up', effect: 'bright', scale: 1.08, denoise: 0, volume: .3 },
};

export const projectTemplateCatalog = Object.freeze(Object.entries(layouts).map(([id, layout]) => Object.freeze({
  id, labelZh: layout.labelZh, category: 'template', description: layout.description,
  parameters: [
    { key: 'title', labelZh: '标题', type: 'text', defaultValue: layout.labelZh },
    { key: 'startMs', labelZh: '插入时间', type: 'number', defaultValue: 0, min: 0, max: MAX_END_MS - 1, step: 1 },
    { key: 'durationMs', labelZh: '持续时间', type: 'number', defaultValue: 5000, min: 1, max: MAX_DURATION_MS, step: 1 },
  ],
  requiredSlots: [{ key: 'visualAssetId', labelZh: '自有画面素材', kinds: ['video', 'image'] }],
  optionalSlots: [{ key: 'audioAssetId', labelZh: '自有音频素材', kinds: ['audio'] }],
})));

function milliseconds(value, fallback, label, maximum, minimum = 0) {
  const number = value === undefined ? fallback : value;
  if (typeof number !== 'number' || !Number.isSafeInteger(number) || number < minimum || number > maximum) {
    throw new Error(`${label}无效，请输入${minimum}到${maximum}之间的整数毫秒`);
  }
  return number;
}

function usableURL(value) {
  if (typeof value !== 'string' || value !== value.trim() || /[\s\\]/.test(value)) return false;
  if (/^\/(?!\/)/.test(value)) return true;
  try {
    const url = new URL(value);
    return ['https:', 'http:', 'blob:'].includes(url.protocol) && !url.username && !url.password;
  } catch {
    return false;
  }
}

function selectedAsset(assets, id, kinds, label) {
  if (typeof id !== 'string' || !id.trim()) throw new Error(`请明确选择${label}`);
  const matches = assets.filter(asset => asset?.id === id);
  if (matches.length !== 1 || !kinds.includes(matches[0].kind)) throw new Error(`${label}不存在、编号重复或类型不符`);
  const asset = matches[0];
  if (!usableURL(asset.url)) throw new Error(`${label}缺少可用的素材地址`);
  if (asset.kind !== 'image') milliseconds(asset.durationMs, undefined, `${label}源时长`, MAX_END_MS, 1);
  return asset;
}

function mediaClip(asset, durationMs, startMs, sourceStartMs) {
  return {
    kind: asset.kind, assetId: asset.id, name: asset.name || '自有素材', startMs, durationMs,
    sourceStartMs, sourceDurationMs: asset.kind === 'image' ? durationMs : asset.durationMs,
    playbackRate: 1, volume: 1, fadeInMs: 0, fadeOutMs: 0,
  };
}

// Review and execution use the same builder, so a reviewed plan has the same
// source range, real styles and track additions as the applied command.
export function planProjectTemplate(timeline, templateId, assets, options = {}) {
  if (!Object.hasOwn(layouts, templateId)) throw new Error('请选择有效的工程模板');
  if (!timeline || !Array.isArray(timeline.tracks)) throw new Error('工程时间线无效');
  if (!Array.isArray(assets)) throw new Error('工程素材列表无效');
  if (!options || typeof options !== 'object' || Array.isArray(options)) throw new Error('模板参数无效');
  const layout = layouts[templateId];
  const visual = selectedAsset(assets, options.visualAssetId, ['video', 'image'], '自有视频或图片');
  const audio = options.audioAssetId === undefined || options.audioAssetId === '' ? null
    : selectedAsset(assets, options.audioAssetId, ['audio'], '自有音频');
  const startMs = milliseconds(options.startMs, 0, '插入时间', MAX_END_MS - 1);
  const requestedDurationMs = milliseconds(options.durationMs, 5000, '模板时长', MAX_DURATION_MS, 1);
  if (startMs + requestedDurationMs > MAX_END_MS) throw new Error('模板结束时间超出一天范围');
  const sourceStartMs = milliseconds(options.sourceStartMs, 0, '画面源起点', visual.kind === 'image' ? 0 : visual.durationMs - 1);
  const audioSourceStartMs = milliseconds(options.audioSourceStartMs, 0, '音频源起点', audio ? audio.durationMs - 1 : 0);
  const durationMs = visual.kind === 'image' ? requestedDurationMs : Math.min(requestedDurationMs, visual.durationMs - sourceStartMs);
  const title = options.title === undefined ? layout.labelZh : options.title;
  if (typeof title !== 'string' || !title.trim() || title.trim().length > 512) throw new Error('请输入1到512字的标题');
  const warnings = [];
  if (durationMs < requestedDurationMs) warnings.push('画面源素材较短，模板已缩短至可用范围');
  const existingOrders = timeline.tracks.map(track => track.order).filter(order => Number.isSafeInteger(order) && order >= 0);
  const nextOrder = Math.max(-1, timeline.tracks.length - 1, ...existingOrders) + 1;
  if (!Number.isSafeInteger(nextOrder + 2)) throw new Error('现有轨道顺序超出有效范围');
  const effect = visualEffectPreset(layout.effect);
  if (effect.preset !== layout.effect) throw new Error('工程模板的画面预设不受支持');
  const visualClip = {
    ...mediaClip(visual, durationMs, startMs, sourceStartMs),
    visualEffect: normalizeVisualEffect(effect),
    canvasTransform: normalizeTransform({ scaleX: layout.scale, scaleY: layout.scale }),
  };
  const tracks = [{
    kind: 'video', name: `${layout.labelZh} · 画面`, order: nextOrder,
    muted: false, locked: false, settings: { previewVisible: true }, clips: [visualClip],
  }];
  if (audio) {
    const audioDurationMs = Math.min(durationMs, audio.durationMs - audioSourceStartMs);
    if (audioDurationMs < durationMs) warnings.push('音频源素材较短，仅使用可用音频，不循环或拉伸');
    const fadeMs = Math.min(300, Math.floor(audioDurationMs / 2));
    tracks.push({
      kind: 'audio', name: `${layout.labelZh} · 配乐`, order: nextOrder + tracks.length,
      muted: false, locked: false, settings: {}, clips: [{
        ...mediaClip(audio, audioDurationMs, startMs, audioSourceStartMs),
        volume: layout.volume, fadeInMs: fadeMs, fadeOutMs: fadeMs,
        audioProcessing: normalizeAudioProcessing({ denoise: layout.denoise, loudnessEnabled: true, targetLUFS: -18 }),
      }],
    });
  }
  tracks.push({
    kind: 'subtitle', name: `${layout.labelZh} · 标题`, order: nextOrder + tracks.length,
    muted: false, locked: false,
    settings: { previewVisible: true, subtitleStyleId: layout.text, subtitleAnimation: layout.animation, ...aicutTextResourcePreset(layout.text) },
    clips: [{ kind: 'subtitle', name: title.trim().slice(0, 16), text: title.trim(), startMs, durationMs, sourceStartMs: 0, sourceDurationMs: durationMs }],
  });
  const selectedAssets = [visual, ...(audio ? [audio] : [])].map(asset => ({ id: asset.id, name: asset.name || '自有素材', kind: asset.kind }));
  const summary = {
    templateId, labelZh: layout.labelZh, startMs, durationMs, endMs: startMs + durationMs,
    trackNames: tracks.map(track => track.name), trackCount: tracks.length,
    clipCount: tracks.reduce((count, track) => count + track.clips.length, 0), assets: selectedAssets, warnings,
  };
  return {
    templateId, labelZh: layout.labelZh,
    options: { visualAssetId: visual.id, audioAssetId: audio?.id || '', title: title.trim(), startMs, durationMs: requestedDurationMs, sourceStartMs, audioSourceStartMs },
    inputReferences: { revision: timeline.revision ?? 0, visualAssetId: visual.id, audioAssetId: audio?.id || '' },
    validation: { valid: true, warnings: [...warnings] }, summary, tracks,
  };
}

function collectIds(value, ids) {
  if (!value || typeof value !== 'object') return;
  if (typeof value.id === 'string') ids.add(value.id);
  for (const child of Object.values(value)) if (child && typeof child === 'object') collectIds(child, ids);
}

export function applyProjectTemplate(timeline, templateId, assets, options = {}, idFactory = () => crypto.randomUUID()) {
  const plan = planProjectTemplate(timeline, templateId, assets, options);
  if (typeof idFactory !== 'function') throw new Error('模板编号生成器无效');
  const used = new Set();
  collectIds(timeline, used);
  collectIds(assets, used);
  const newId = () => {
    for (let attempt = 0; attempt < 100; attempt += 1) {
      const id = idFactory();
      if (typeof id !== 'string' || !id.trim() || id.length > 96 || id.includes('::') || used.has(id)) continue;
      used.add(id);
      return id;
    }
    throw new Error('无法生成不重复的模板编号');
  };
  const addedTracks = plan.tracks.map(track => ({ ...track, id: newId(), clips: track.clips.map(clip => ({ ...clip, id: newId() })) }));
  return {
    timeline: { ...timeline, tracks: [...timeline.tracks, ...addedTracks] },
    summary: plan.summary,
    selectionKey: `${addedTracks[0].id}::${addedTracks[0].clips[0].id}`,
  };
}
