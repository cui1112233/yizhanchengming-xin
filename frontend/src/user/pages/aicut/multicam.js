const MAX_MS = 86400000;
function ms(value, label, min = 0, max = MAX_MS) {
  if (!Number.isSafeInteger(value) || value < min || value > max) throw new Error(`${label}须为${min}到${max}之间的整数毫秒`);
  return value;
}
function assetFor(assets, id, kind) {
  const matches = assets.filter(a => a?.id === id);
  if (typeof id !== 'string' || !id.trim() || matches.length !== 1 || matches[0].kind !== kind) throw new Error('请明确选择不重复的自有素材');
  const a = matches[0];
  let validURL = typeof a.url === 'string' && /^\/(?!\/)/.test(a.url);
  try { const u = new URL(a.url); validURL ||= ['http:', 'https:', 'blob:'].includes(u.protocol) && !u.username && !u.password; } catch { /* relative route checked above */ }
  if (!validURL || /[\s\\]/.test(a.url)) throw new Error('素材缺少可用地址');
  ms(a.durationMs, '素材时长', 1);
  return { id: a.id, kind: a.kind, name: a.name || '自有素材', url: a.url, durationMs: a.durationMs };
}
function clip(asset, startMs, durationMs, sourceStartMs, kind = asset.kind) {
  return { kind, assetId: asset.id, name: asset.name, startMs, durationMs, sourceStartMs, sourceDurationMs: asset.durationMs, playbackRate: 1, volume: kind === 'video' ? 0 : 1, fadeInMs: 0, fadeOutMs: 0 };
}
export function planMultiCamGroup(assets, options = {}) {
  if (!Array.isArray(assets) || !options || typeof options !== 'object' || Array.isArray(options)) throw new Error('多机位参数无效');
  if (!Array.isArray(options.cameras) || options.cameras.length < 2 || options.cameras.length > 9) throw new Error('请选择2到9个机位');
  const cameras = options.cameras.map(c => ({ asset: assetFor(assets, c?.assetId, 'video'), offsetMs: ms(c?.offsetMs ?? 0, '源时间偏移', -MAX_MS) }));
  if (new Set(cameras.map(c => c.asset.id)).size !== cameras.length) throw new Error('每个机位须选择不同素材');
  const commonStartMs = Math.max(0, ...cameras.map(c => -c.offsetMs));
  const commonEndMs = Math.min(...cameras.map(c => c.asset.durationMs - c.offsetMs));
  const durationMs = commonEndMs - commonStartMs;
  if (durationMs < 1) throw new Error('所选机位没有共同可用的时间范围，请调整源时间偏移');
  const startMs = ms(options.startMs ?? 0, '插入时间');
  ms(startMs + durationMs, '结束时间');
  if (!Array.isArray(options.cuts) || !options.cuts.length || options.cuts[0]?.atMs !== 0) throw new Error('第一个机位切换点须从0毫秒开始');
  const cuts = options.cuts.map((c, i) => {
    const atMs = ms(c?.atMs, '切换时间', 0, durationMs - 1);
    const cameraIndex = ms(c?.cameraIndex, '机位序号', 0, cameras.length - 1);
    if (i > 0 && atMs <= options.cuts[i - 1].atMs) throw new Error('切换时间须严格递增且不能重叠');
    return { atMs, cameraIndex };
  });
  const tracks = [{ kind: 'video', name: '多机位 · 画面', muted: false, locked: false, settings: { previewVisible: true }, clips: cuts.map((cut, i) => {
    const camera = cameras[cut.cameraIndex];
    return clip(camera.asset, startMs + cut.atMs, (cuts[i + 1]?.atMs ?? durationMs) - cut.atMs, commonStartMs + cut.atMs + camera.offsetMs);
  }) }];
  const audio = options.audio ?? { mode: 'none' };
  if (!audio || !['none', 'asset'].includes(audio.mode)) throw new Error('请选择自有音频或静音；机位原声须先提取为音频素材');
  let audioAsset;
  let audioSourceStartMs = 0;
  if (audio.mode === 'asset') {
    audioAsset = assetFor(assets, audio.assetId, 'audio');
    audioSourceStartMs = ms(audio.sourceStartMs ?? 0, '音频源起点');
    if (audioSourceStartMs + durationMs > audioAsset.durationMs) throw new Error('主音频不足以覆盖共同画面时长，请调整音频源起点或更换音频');
    tracks.push({ kind: 'audio', name: '多机位 · 主音频', muted: false, locked: false, settings: {}, clips: [clip(audioAsset, startMs, durationMs, audioSourceStartMs)] });
  }
  const normalizedOptions = { cameras: cameras.map(c => ({assetId: c.asset.id, offsetMs: c.offsetMs})), cuts, startMs, audio: audioAsset ? {mode:'asset',assetId:audioAsset.id,sourceStartMs:audioSourceStartMs} : {mode:'none'} };
  return { type: 'manual-multicam', assets: [...cameras.map(c => c.asset), ...(audioAsset ? [audioAsset] : [])], options: normalizedOptions, tracks, summary: { cameraCount: cameras.length, clipCount: cuts.length, durationMs, startMs, endMs: startMs + durationMs, commonStartMs, commonEndMs, audioMode: audio.mode, audioAssetId: audioAsset?.id ?? '', alignment: 'source-offset-review', warnings: ['请复核源时间偏移和共同可用范围后应用', ...(audio.mode === 'none' ? ['全部机位原声静音，未添加主音频'] : [])] } };
}
function collectIds(value, used) {
  if (!value || typeof value !== 'object') return;
  if (typeof value.id === 'string') used.add(value.id);
  Object.values(value).forEach(child => { if (child && typeof child === 'object') collectIds(child, used); });
}
export function applyMultiCamGroup(timeline, plan, idFactory = () => crypto.randomUUID()) {
  if (!timeline || !Array.isArray(timeline.tracks) || typeof idFactory !== 'function') throw new Error('工程时间线或编号生成器无效');
  const rebuilt = planMultiCamGroup(plan?.assets, plan?.options);
  if (JSON.stringify(rebuilt) !== JSON.stringify(plan)) throw new Error('多机位方案已变化，请重新检查后应用');
  const used = new Set(); collectIds(timeline, used); collectIds(plan.assets, used);
  function nextId() {
    for (let i = 0; i < 100; i += 1) { const id = idFactory(); if (typeof id === 'string' && id.trim() && id.length <= 96 && !id.includes('::') && !used.has(id)) { used.add(id); return id; } }
    throw new Error('无法生成不重复的多机位编号');
  }
  const nextOrder = Math.max(timeline.tracks.length - 1, -1, ...timeline.tracks.map(t => Number.isSafeInteger(t.order) ? t.order : -1)) + 1;
  if (!Number.isSafeInteger(nextOrder + rebuilt.tracks.length)) throw new Error('轨道顺序超出有效范围');
  const tracks = rebuilt.tracks.map((t, index) => ({ ...t, order: nextOrder + index, id: nextId(), clips: t.clips.map(c => ({...c, id: nextId()})) }));
  return { timeline: {...timeline, tracks: [...timeline.tracks, ...tracks]}, selectionKey: `${tracks[0].id}::${tracks[0].clips[0].id}`, summary: rebuilt.summary };
}

export function reviewMultiCamAlignment(cameras, response) {
  if (!response || response.method !== 'audio-energy-cross-correlation' || response.timeResolutionMs !== 20 || !Array.isArray(cameras) || !Array.isArray(response.cameras) || response.cameras.length !== cameras.length || response.referenceAssetId !== cameras[0]?.assetId) throw new Error('声音同步结果与当前机位不符，请重新分析');
  const results = cameras.map(camera => {
    const matches = response.cameras.filter(result => result.assetId === camera.assetId);
    const result = matches[0];
    if (matches.length !== 1 || !Number.isSafeInteger(result.offsetMs) || Math.abs(result.offsetMs) > 10000 || result.offsetMs % 20 !== 0 || typeof result.confidence !== 'number' || !Number.isFinite(result.confidence) || result.confidence < .65 || result.confidence > 1 || (camera.assetId === response.referenceAssetId && result.offsetMs !== 0)) throw new Error('声音同步结果不完整或匹配不足，请改用手动对齐');
    return { assetId: camera.assetId, offsetMs: result.offsetMs, confidence: result.confidence };
  });
  return { cameras: results.map(({assetId,offsetMs}) => ({assetId,offsetMs})), results, timeResolutionMs: response.timeResolutionMs };
}
