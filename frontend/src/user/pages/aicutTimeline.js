const labels = { video: '视频', audio: '音频', subtitle: '字幕' };
const subtitleRhythms = { sentence: '按句/短语', word: '逐词高亮' };
const timelineBasePixelsPerSecond = 72;

function timelineZoom(value) {
  const zoom = Number(value);
  return Number.isFinite(zoom) ? Math.min(4, Math.max(0.5, zoom)) : 1;
}

export function timelineTimeToPixels(timeMs, zoom = 1) {
  return Math.max(0, Number(timeMs) || 0) / 1000 * timelineBasePixelsPerSecond * timelineZoom(zoom);
}

export function timelinePixelsToTime(pixels, zoom = 1) {
  return Math.round(Math.max(0, Number(pixels) || 0) / (timelineBasePixelsPerSecond * timelineZoom(zoom)) * 1000);
}

export function draggedTimelineStart(originalStartMs, deltaPixels, zoom = 1) {
  const original = Math.max(0, Number(originalStartMs) || 0);
  const delta = Math.round((Number(deltaPixels) || 0) / (timelineBasePixelsPerSecond * timelineZoom(zoom)) * 1000);
  return Math.max(0, original + delta);
}

export function normalizeSubtitleRhythm(value) {
  if (value === 'word' || value === subtitleRhythms.word) return 'word';
  return 'sentence';
}

export function subtitleRhythmLabel(value) {
  return subtitleRhythms[normalizeSubtitleRhythm(value)];
}

export function resolveAICutAgentTool(instruction) {
  const value = String(instruction || '').trim();
  if (!value) return null;
  if (/理解.*(?:画面|镜头)|(?:画面|镜头).*(?:理解|判断|保留)/i.test(value)) return { tool: 'visual_analyze' };
  if (/识别.*字幕|字幕.*识别/i.test(value)) return { tool: 'caption_asr' };
  if (/口误|语气词|停顿|水词/i.test(value)) return { tool: 'oral_clean' };
  if (/背景音乐|配乐|BGM/i.test(value)) return { tool: 'bgm_plan' };
  if (/粗剪|切镜|废镜头|重复镜头/i.test(value)) return { tool: 'smart_split' };
  return null;
}

export function planTimelineBGM(timeline, asset, options = {}, idFactory) {
  if (!asset || asset.kind !== 'audio') throw new Error('请选择音频素材');
  const assetDuration = Math.round(Number(asset.durationMs));
  if (!Number.isFinite(assetDuration) || assetDuration <= 0) throw new Error('音频素材时长无效');
  if (typeof idFactory !== 'function') throw new Error('片段编号生成器无效');

  const tracks = copyTracks(timeline);
  const reference = tracks.find(track => track.kind === 'video' && track.id === options.referenceTrackId)
    || tracks.find(track => track.kind === 'video' && track.clips.length);
  if (!reference || !reference.clips.length) throw new Error('请选择有片段的参考画面轨道');
  const referenceStart = Math.min(...reference.clips.map(clip => Number(clip.startMs) || 0));
  const referenceEnd = Math.max(...reference.clips.map(clip => (Number(clip.startMs) || 0) + (Number(clip.durationMs) || 0)));
  if (referenceEnd <= referenceStart) throw new Error('参考画面轨道时长无效');

  let bgmTrack = tracks.find(track => track.kind === 'audio' && track.name === 'BGM');
  if (bgmTrack?.locked) throw new Error('BGM 轨道已锁定');
  if (!bgmTrack) {
    bgmTrack = { id: idFactory(), kind: 'audio', name: 'BGM', order: tracks.length, muted: false, locked: false, clips: [] };
    tracks.push(bgmTrack);
  }
  const keepExisting = options.continue === true;
  if (!keepExisting) bgmTrack.clips = [];
  let startMs = options.align === false ? 0 : referenceStart;
  if (keepExisting && bgmTrack.clips.length) {
    startMs = Math.max(startMs, ...bgmTrack.clips.map(clip => Number(clip.startMs) + Number(clip.durationMs)));
  }
  const fill = options.fill !== false;
  const targetEnd = fill ? referenceEnd : startMs + assetDuration;
  if (targetEnd <= startMs) throw new Error('参考画面范围已经铺满');
  const added = [];
  while (startMs < targetEnd) {
    const durationMs = Math.min(assetDuration, targetEnd - startMs);
    added.push({
      id: idFactory(), assetId: asset.id, kind: 'audio', name: asset.name || '背景音乐',
      startMs, durationMs, sourceStartMs: 0, sourceDurationMs: assetDuration,
      playbackRate: 1, volume: 0.25, fadeInMs: 0, fadeOutMs: 0, text: ''
    });
    startMs += durationMs;
    if (!fill) break;
  }
  if (options.fade !== false && added.length) {
    added[0].fadeInMs = Math.min(500, Math.floor(added[0].durationMs / 2));
    const last = added[added.length - 1];
    last.fadeOutMs = Math.min(500, Math.floor(last.durationMs / 2));
  }
  bgmTrack.clips.push(...added);
  bgmTrack.clips.sort((a, b) => Number(a.startMs) - Number(b.startMs));
  return { ...timeline, tracks };
}

function copyTracks(timeline) {
  return (timeline?.tracks || []).map(track => ({ ...track, settings: { ...(track.settings || {}) }, clips: (track.clips || []).map(clip => ({ ...clip })) }));
}

function updateClip(timeline, trackId, clipId, updater) {
  let found = false;
  const tracks = copyTracks(timeline).map(track => {
    if (track.id !== trackId) return track;
    const clips = track.clips.map(clip => {
      if (clip.id !== clipId) return clip;
      found = true;
      return updater(clip);
    });
    return { ...track, clips };
  });
  if (!found) throw new Error('片段不存在');
  return { ...timeline, tracks };
}

export function addTimelineTrack(timeline, kind, idFactory) {
  if (!labels[kind]) throw new Error('不支持的轨道类型');
  const tracks = copyTracks(timeline);
  const number = tracks.filter(track => track.kind === kind).length + 1;
  tracks.push({ id: idFactory(), kind, name: `${labels[kind]}${number}`, order: tracks.length, muted: false, locked: false, clips: [] });
  return { ...timeline, tracks };
}

export function toggleTimelineTrackFlag(timeline, trackId, field) {
  if (!['muted', 'locked'].includes(field)) throw new Error('不支持的轨道状态');
  let found = false;
  const tracks = copyTracks(timeline).map(track => {
    if (track.id !== trackId) return track;
    found = true;
    return { ...track, [field]: !track[field] };
  });
  if (!found) throw new Error('轨道不存在');
  return { ...timeline, tracks };
}

export function renameTimelineTrack(timeline, trackId, name) {
  const nextName = String(name || '').trim();
  if (!nextName) throw new Error('轨道名称不能为空');
  if ([...nextName].length > 64) throw new Error('轨道名称不能超过 64 个字');
  let found = false;
  const tracks = copyTracks(timeline).map(track => {
    if (track.id !== trackId) return track;
    found = true;
    return { ...track, name: nextName };
  });
  if (!found) throw new Error('轨道不存在');
  return { ...timeline, tracks };
}

export function removeTimelineTrack(timeline, trackId) {
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.clips?.length) throw new Error('轨道中还有片段，不能删除');
  return { ...timeline, tracks: copyTracks(timeline).filter(item => item.id !== trackId).map((item, order) => ({ ...item, order })) };
}

export function splitTimelineClipAt(timeline, trackId, clipId, offsetMs, idFactory) {
  let found = false;
  const tracks = copyTracks(timeline).map(track => {
    if (track.id !== trackId) return track;
    if (track.locked) throw new Error('轨道已锁定');
    const index = track.clips.findIndex(clip => clip.id === clipId);
    if (index < 0) return track;
    const clip = track.clips[index];
    const durationMs = Math.round(Number(clip.durationMs) || 0);
    const splitAt = Math.round(Number(offsetMs));
    if (durationMs < 200) throw new Error('片段太短，无法分割');
    if (!Number.isFinite(splitAt) || splitAt < 100 || splitAt > durationMs - 100) throw new Error('分割点距离片段边缘太近');
    found = true;
    const firstDuration = splitAt;
    let firstText = clip.text;
    let secondText = clip.text;
    if (track.kind === 'subtitle' || clip.kind === 'subtitle') {
      const sourceText = String(clip.text || '').trim();
      const words = sourceText.split(/\s+/).filter(Boolean);
      if (words.length > 1) {
        const middle = Math.max(1, Math.min(words.length - 1, Math.round(words.length * firstDuration / durationMs)));
        firstText = words.slice(0, middle).join(' ');
        secondText = words.slice(middle).join(' ');
      } else {
        const runes = [...sourceText];
        const middle = Math.max(1, Math.min(runes.length - 1, Math.round(runes.length * firstDuration / durationMs)));
        firstText = runes.slice(0, middle).join('');
        secondText = runes.slice(middle).join('');
      }
    }
    const rate = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    const sourceSplit = Math.round(firstDuration * rate);
    const first = { ...clip, text: firstText, name: String(firstText || clip.name).slice(0, 16), durationMs: firstDuration, sourceDurationMs: sourceSplit };
    const second = {
      ...clip,
      id: idFactory(),
      name: `${clip.name}（后半段）`,
      startMs: clip.startMs + firstDuration,
      sourceStartMs: (Number(clip.sourceStartMs) || 0) + sourceSplit,
      sourceDurationMs: Math.max(1, (Number(clip.sourceDurationMs) || Math.round(durationMs * rate)) - sourceSplit),
      durationMs: durationMs - firstDuration,
      text: secondText,
      name: String(secondText || `${clip.name}（后半段）`).slice(0, 16)
    };
    return { ...track, clips: [...track.clips.slice(0, index), first, second, ...track.clips.slice(index + 1)] };
  });
  if (!found) throw new Error('片段不存在');
  return { ...timeline, tracks };
}

export function splitTimelineClip(timeline, trackId, clipId, idFactory) {
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  const clip = track?.clips?.find(item => item.id === clipId);
  if (!clip) throw new Error('片段不存在');
  return splitTimelineClipAt(timeline, trackId, clipId, Math.floor(Number(clip.durationMs) / 2), idFactory);
}

export function setTimelineClipText(timeline, trackId, clipId, text) {
  const value = String(text || '').trim();
  if (!value) throw new Error('文字不能为空');
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  return updateClip(timeline, trackId, clipId, clip => ({ ...clip, text: value, name: value.slice(0, 16) }));
}

export function mergeTimelineClipWithNext(timeline, trackId, clipId) {
  const tracks = copyTracks(timeline);
  const track = tracks.find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  const ordered = [...track.clips].sort((left, right) => Number(left.startMs) - Number(right.startMs));
  const index = ordered.findIndex(item => item.id === clipId);
  if (index < 0) throw new Error('片段不存在');
  if (index >= ordered.length - 1) throw new Error('后面没有可合并的字幕');
  const current = ordered[index];
  const next = ordered[index + 1];
  if (track.kind !== 'subtitle' && current.kind !== 'subtitle') throw new Error('只有字幕或文字可以合并');
  const end = Math.max(Number(current.startMs) + Number(current.durationMs), Number(next.startMs) + Number(next.durationMs));
  const text = `${String(current.text || '').trim()} ${String(next.text || '').trim()}`.trim();
  const merged = { ...current, text, name: text.slice(0, 16), durationMs: end - Number(current.startMs), sourceDurationMs: end - Number(current.startMs) };
  track.clips = ordered.filter(item => item.id !== current.id && item.id !== next.id);
  track.clips.push(merged);
  track.clips.sort((left, right) => Number(left.startMs) - Number(right.startMs));
  return { ...timeline, tracks };
}

export function trimTimelineClipEdge(timeline, trackId, clipId, edge, amountMs) {
  const amount = Math.round(Number(amountMs));
  if (!['left', 'right'].includes(edge)) throw new Error('不支持的裁剪方向');
  if (!Number.isFinite(amount) || amount <= 0) throw new Error('裁剪时长无效');
  return updateClip(timeline, trackId, clipId, clip => {
    if (clip.durationMs - amount < 100) throw new Error('片段至少保留 100 毫秒');
    if (edge === 'right') return { ...clip, durationMs: clip.durationMs - amount };
    const rate = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    return { ...clip, startMs: clip.startMs + amount, sourceStartMs: clip.sourceStartMs + Math.round(amount * rate), durationMs: clip.durationMs - amount };
  });
}

export function setClipPlaybackRate(timeline, trackId, clipId, playbackRate) {
  const rate = Number(playbackRate);
  if (!Number.isFinite(rate) || rate < 0.25 || rate > 4) throw new Error('变速范围为 0.25x–4x');
  return updateClip(timeline, trackId, clipId, clip => {
    const previous = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    return { ...clip, playbackRate: rate, durationMs: Math.max(1, Math.round(clip.durationMs * previous / rate)) };
  });
}

export function moveTimelineClip(timeline, trackId, clipId, startMs, options = {}) {
  const requested = Math.round(Number(startMs));
  if (!Number.isFinite(requested) || requested < 0) throw new Error('开始时间不能小于 0');
  const tracks = copyTracks(timeline);
  const targetTrack = tracks.find(track => track.id === trackId);
  if (!targetTrack) throw new Error('轨道不存在');
  if (targetTrack.locked) throw new Error('轨道已锁定');
  const target = targetTrack.clips.find(clip => clip.id === clipId);
  if (!target) throw new Error('片段不存在');

  let nextStart = requested;
  if (options.magnetic !== false) {
    const threshold = Math.max(0, Number(options.thresholdMs) || 150);
    const boundaries = tracks.flatMap(track => track.clips.flatMap(clip => {
      if (track.id === trackId && clip.id === clipId) return [];
      const clipStart = Number(clip.startMs) || 0;
      return [clipStart, clipStart + (Number(clip.durationMs) || 0)];
    }));
    const nearest = boundaries.reduce((best, boundary) => best == null || Math.abs(boundary - requested) < Math.abs(best - requested) ? boundary : best, null);
    if (nearest != null && Math.abs(nearest - requested) <= threshold) nextStart = nearest;
  }

  const originalStart = Number(target.startMs) || 0;
  const delta = nextStart - originalStart;
  for (const track of tracks) {
    for (const clip of track.clips) {
      if (track.id === trackId && clip.id === clipId) {
        clip.startMs = nextStart;
      } else if (options.linked === true && !track.locked && Number(clip.startMs) === originalStart) {
        clip.startMs = Math.max(0, Number(clip.startMs) + delta);
      }
    }
  }
  return { ...timeline, tracks };
}

export function markTimelineClip(timeline, trackId, clipId, reviewStatus, reviewReason = '') {
  if (!['', 'delete', 'review', 'keep'].includes(reviewStatus)) throw new Error('不支持的建议状态');
  return updateClip(timeline, trackId, clipId, clip => ({ ...clip, reviewStatus, reviewReason: String(reviewReason || '').trim() }));
}

export function updateTimelineTrackSettings(timeline, trackId, patch) {
  let found = false;
  const tracks = copyTracks(timeline).map(track => {
    if (track.id !== trackId) return track;
    found = true;
    return { ...track, settings: { ...(track.settings || {}), ...(patch || {}) } };
  });
  if (!found) throw new Error('轨道不存在');
  return { ...timeline, tracks };
}

function operationResult(timeline, reasonZh = '') {
  return { ok: !reasonZh, timeline, reasonZh };
}

export function addEffectLayer(timeline, effect, idFactory) {
  const startMs = Math.round(Number(effect?.startMs));
  const durationMs = Math.round(Number(effect?.durationMs));
  if (!String(effect?.resourceId || '').startsWith('effect-')) return operationResult(timeline, '特效资源无效');
  if (!Number.isFinite(startMs) || startMs < 0) return operationResult(timeline, '开始时间不能小于 0');
  if (!Number.isFinite(durationMs) || durationMs <= 0) return operationResult(timeline, '特效时长必须大于 0');
  const tracks = copyTracks(timeline);
  let track = tracks.find(item => item.kind === 'effect');
  if (track?.locked) return operationResult(timeline, '轨道已锁定');
  const id = idFactory();
  if (!track) {
    track = { id: `track-${id}`, kind: 'effect', name: '特效 1', order: tracks.length, muted: false, locked: false, clips: [] };
    tracks.push(track);
  }
  track.clips.push({ id, kind: 'effect', resourceId: effect.resourceId, name: effect.name || effect.resourceId, startMs, durationMs, parameters: { ...(effect.parameters || {}) } });
  track.clips.sort((a, b) => Number(a.startMs) - Number(b.startMs));
  return operationResult({ ...timeline, tracks });
}

export function moveEffectLayer(timeline, trackId, effectId, startMs) {
  const requested = Math.round(Number(startMs));
  if (!Number.isFinite(requested) || requested < 0) return operationResult(timeline, '开始时间不能小于 0');
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) return operationResult(timeline, '轨道不存在');
  if (track.locked) return operationResult(timeline, '轨道已锁定');
  if (!track.clips?.some(item => item.id === effectId)) return operationResult(timeline, '特效不存在');
  return operationResult(updateClip(timeline, trackId, effectId, effect => ({ ...effect, startMs: requested })));
}

export function resizeEffectLayer(timeline, trackId, effectId, durationMs) {
  const duration = Math.round(Number(durationMs));
  if (!Number.isFinite(duration) || duration <= 0) return operationResult(timeline, '特效时长必须大于 0');
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) return operationResult(timeline, '轨道不存在');
  if (track.locked) return operationResult(timeline, '轨道已锁定');
  if (!track.clips?.some(item => item.id === effectId)) return operationResult(timeline, '特效不存在');
  return operationResult(updateClip(timeline, trackId, effectId, effect => ({ ...effect, durationMs: duration })));
}

export function removeTimelineObject(timeline, trackId, objectId) {
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) return operationResult(timeline, '轨道不存在');
  if (track.locked) return operationResult(timeline, '轨道已锁定');
  if (!track.clips?.some(item => item.id === objectId)) return operationResult(timeline, '对象不存在');
  const tracks = copyTracks(timeline).map(item => item.id === trackId ? { ...item, clips: item.clips.filter(clip => clip.id !== objectId) } : item);
  return operationResult({ ...timeline, tracks });
}

export function setClipTransition(timeline, trackId, clipId, resourceId, durationMs) {
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track || track.kind !== 'video') return operationResult(timeline, '视频轨道不存在');
  if (track.locked) return operationResult(timeline, '轨道已锁定');
  const ordered = [...(track.clips || [])].sort((a, b) => Number(a.startMs) - Number(b.startMs));
  const index = ordered.findIndex(item => item.id === clipId);
  if (index <= 0) return operationResult(timeline, '转场需要放在两个相邻片段之间');
  const previous = ordered[index - 1];
  const current = ordered[index];
  if (Number(previous.startMs) + Number(previous.durationMs) !== Number(current.startMs)) return operationResult(timeline, '两个片段之间有空隙');
  const maximum = Math.max(100, Math.floor(Math.min(Number(previous.durationMs), Number(current.durationMs)) / 2));
  const safeDuration = Math.min(maximum, Math.max(100, Math.round(Number(durationMs) || 500)));
  const next = updateClip(timeline, trackId, clipId, clip => ({ ...clip, transition: resourceId, transitionMs: safeDuration }));
  return operationResult(next);
}
