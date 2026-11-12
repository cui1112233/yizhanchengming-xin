const MAX_END = 86400000;
const copy = value => JSON.parse(JSON.stringify(value));
const clipFields = new Set('kind name startMs durationMs sourceStartMs sourceDurationMs text playbackRate reversed freezeSourceMs blendMode volume fadeInMs fadeOutMs visualEffect catalogResourceId textResourceId resourceId parameters canvasTransform keyframes speedCurve audioProcessing visualProcessing animation'.split(' '));
const settingFields = new Set('subtitleStyleId subtitleRhythm subtitleAnimation textResourceId previewVisible fontFamily fontSize textCurve textColor strokeColor strokeWidth backgroundColor shadowColor shadowOffsetX shadowOffsetY shadowBlur letterSpacing lineHeight textAlign textRuns'.split(' '));
function integer(value, min, max, label) {
  if (!Number.isSafeInteger(value) || value < min || value > max) throw new Error(`${label}无效`);
  return value;
}
function rejectReferences(value) {
  if (!value || typeof value !== 'object') return;
  for (const [key, child] of Object.entries(value)) {
    if (/url|assetId|signature|compound|linkGroup|transition.*Id/i.test(key) && child !== '' && child != null && child !== false && !(Array.isArray(child) && !child.length)) throw new Error('模板不能保存私有素材引用、人物蒙版、复合片段或联动');
    rejectReferences(child);
  }
}
function assertLayout(layout) {
  if (!layout || layout.version !== 1 || !Array.isArray(layout.tracks) || !layout.tracks.length || layout.tracks.length > 100 || !Array.isArray(layout.slots) || layout.slots.length > 500) throw new Error('个人模板格式无效');
  const slots = new Map(); const used = new Set(); let count = 0;
  layout.slots.forEach((slot, i) => {
    if (slot.id !== `slot-${i + 1}` || slot.label !== `素材 ${i + 1}` || !['video', 'image', 'audio'].includes(slot.kind)) throw new Error('模板素材位置无效');
    slots.set(slot.id, slot);
  });
  layout.tracks.forEach((track, i) => {
    if (track.id !== `track-${i + 1}` || track.order !== i || !['video', 'audio', 'subtitle'].includes(track.kind) || !Array.isArray(track.clips) || !track.clips.length) throw new Error('模板轨道无效');
    rejectReferences(track.settings);
    for (const clip of track.clips) {
      count += 1;
      if (count > 500 || clip.id !== `clip-${count}` || !['video', 'image', 'audio', 'subtitle'].includes(clip.kind)) throw new Error('模板片段无效');
      integer(clip.startMs, 0, MAX_END - 1, '片段起点'); integer(clip.durationMs, 1, MAX_END - clip.startMs, '片段时长');
      integer(clip.sourceStartMs ?? 0, 0, MAX_END, '素材起点');
      const safe = { ...clip }; delete safe.assetId; rejectReferences(safe);
      if (clip.transition || clip.transitionMs || clip.visualProcessing?.personSegmentation?.enabled) throw new Error('暂不支持转场和人物蒙版');
      if (clip.kind === 'subtitle') { if (clip.assetId) throw new Error('字幕不能引用源素材'); }
      else { if (slots.get(clip.assetId)?.kind !== clip.kind) throw new Error('模板素材位置不匹配'); used.add(clip.assetId); }
    }
  });
  if (!count || used.size !== slots.size) throw new Error('模板素材位置无效');
  return layout;
}
export function extractPersonalTemplate(timeline) {
  if (!timeline || !Array.isArray(timeline.tracks)) throw new Error('当前时间线无效');
  const slots = []; const sourceSlots = new Map(); let count = 0;
  const tracks = timeline.tracks.filter(track => track.clips?.length).map((track, i) => {
    if (!['video', 'audio', 'subtitle'].includes(track.kind)) throw new Error('目前只支持画面、音频和字幕轨道');
    const settings = track.settings || {}; rejectReferences(settings);
    for (const key of Object.keys(settings)) if (!settingFields.has(key) && settings[key]) throw new Error(`暂不支持轨道设置：${key}`);
    return { id: `track-${i + 1}`, kind: track.kind, name: track.name || '模板轨道', order: i, muted: !!track.muted, locked: false,
      settings: copy(Object.fromEntries(Object.entries(settings).filter(([key]) => settingFields.has(key)))),
      clips: track.clips.map(clip => {
        const { id, assetId, reviewStatus, reviewReason, ...rest } = clip;
        rejectReferences(rest);
        if (clip.transition || clip.transitionMs || clip.kind === 'transition') throw new Error('请先移除转场再保存模板');
        if (clip.speedCurve?.length) throw new Error('模板暂不支持变速曲线，请先改为固定速度');
        for (const [key, value] of Object.entries(rest)) if (!clipFields.has(key) && value != null && value !== '' && value !== false && value !== 0 && !(Array.isArray(value) && !value.length)) throw new Error(`暂不支持片段设置：${key}`);
        let slotId;
        if (clip.kind !== 'subtitle') {
          if (!assetId) throw new Error('画面或音频缺少源素材');
          const existing = sourceSlots.get(assetId);
          if (existing && existing.kind !== clip.kind) throw new Error('同一源素材类型不一致');
          if (!existing) { const slot = { id: `slot-${slots.length + 1}`, kind: clip.kind, label: `素材 ${slots.length + 1}` }; slots.push(slot); sourceSlots.set(assetId, slot); }
          slotId = sourceSlots.get(assetId).id;
        }
        count += 1;
        return { ...copy(Object.fromEntries(Object.entries(rest).filter(([key]) => clipFields.has(key)))), id: `clip-${count}`, ...(slotId ? { assetId: slotId } : {}), sourceStartMs: clip.sourceStartMs ?? 0, sourceDurationMs: clip.sourceDurationMs ?? clip.durationMs };
      }) };
  });
  return assertLayout({ version: 1, slots, tracks });
}
function usableURL(value) {
  if (typeof value !== 'string' || /[\s\\]/.test(value)) return false;
  if (/^\/(?!\/)/.test(value)) return true;
  try { const u = new URL(value); return ['https:', 'http:', 'blob:'].includes(u.protocol) && !u.username && !u.password; } catch { return false; }
}
export function planPersonalTemplate(timeline, template, assets, options = {}) {
  const layout = assertLayout(copy(template?.layout || template));
  if (!Array.isArray(timeline?.tracks) || !Array.isArray(assets)) throw new Error('当前工程无效');
  const startMs = integer(options.startMs ?? 0, 0, MAX_END - 1, '插入时间');
  const selected = new Map();
  for (const slot of layout.slots) {
    const id = options.bindings?.[slot.id]; const matches = assets.filter(asset => asset.id === id);
    if (!id || matches.length !== 1 || matches[0].kind !== slot.kind || !usableURL(matches[0].url)) throw new Error(`请为${slot.label}选择本工程可用的${slot.kind === 'audio' ? '音频' : slot.kind === 'image' ? '图片' : '视频'}`);
    selected.set(slot.id, matches[0]);
  }
  const maxOrder = Math.max(-1, timeline.tracks.length - 1, ...timeline.tracks.map(track => track.order).filter(Number.isSafeInteger));
  integer(maxOrder + layout.tracks.length, 0, Number.MAX_SAFE_INTEGER, '轨道顺序');
  const tracks = layout.tracks.map((track, index) => ({ ...track, order: maxOrder + index + 1, clips: track.clips.map(clip => {
    if (startMs + clip.startMs + clip.durationMs > MAX_END) throw new Error('插入后超出一天时间范围');
    const asset = selected.get(clip.assetId);
    const result = { ...clip, startMs: startMs + clip.startMs };
    if (asset) {
      result.assetId = asset.id;
      if (asset.kind !== 'image') {
        integer(asset.durationMs, 1, MAX_END, '替换素材时长');
        const rate = clip.playbackRate ?? 1;
        if (!Number.isFinite(rate) || rate <= 0 || clip.speedCurve?.length) throw new Error('模板暂不支持变速曲线，请先改为固定速度');
        const need = clip.freezeSourceMs == null ? Math.ceil(clip.durationMs * rate) : 1;
        const sourceStart = clip.freezeSourceMs ?? clip.sourceStartMs;
        if (sourceStart + need > asset.durationMs) throw new Error(`${asset.name || '替换素材'}时长不足；请换用更长素材，模板不会截短或拉伸`);
      }
      result.sourceDurationMs = asset.kind === 'image' ? clip.durationMs : asset.durationMs;
    }
    return result;
  }) }));
  return { tracks, summary: { title: template.title || '个人模板', trackCount: tracks.length, clipCount: tracks.reduce((n,t)=>n+t.clips.length,0), startMs, durationMs: Math.max(...tracks.flatMap(t=>t.clips.map(c=>c.startMs+c.durationMs))) - startMs, assets: [...selected.values()].map(({id,name,kind})=>({id,name,kind})), warnings: [] } };
}
export function applyPersonalTemplate(timeline, template, assets, options = {}, idFactory = () => crypto.randomUUID()) {
  const plan = planPersonalTemplate(timeline, template, assets, options);
  const used = new Set();
  function collect(v) { if (!v || typeof v !== 'object') return; if (typeof v.id === 'string') used.add(v.id); Object.values(v).forEach(collect); }
  collect(timeline); collect(assets); collect(template);
  function next() { for (let i=0;i<100;i++) { const id=idFactory(); if (typeof id==='string' && /^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id) && !used.has(id)) { used.add(id);return id; } } throw new Error('无法生成不重复的片段编号'); }
  const tracks = plan.tracks.map(track => ({ ...track, id: next(), clips: track.clips.map(clip => ({ ...clip, id: next() })) }));
  return { timeline: { ...timeline, tracks: [...timeline.tracks, ...tracks] }, summary: plan.summary, selectionKey: `${tracks[0].id}::${tracks[0].clips[0].id}` };
}
