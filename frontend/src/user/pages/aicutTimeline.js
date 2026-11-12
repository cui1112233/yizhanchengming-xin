const labels = { video: '视频', audio: '音频', subtitle: '字幕', transition: '转场' };
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

export function timelinePixelDeltaToTime(pixels, zoom = 1) {
  return Math.round((Number(pixels) || 0) / (timelineBasePixelsPerSecond * timelineZoom(zoom)) * 1000);
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
  const beats = [...new Set((options.beatsMs || []).map(value => Math.round(Number(value))).filter(value => Number.isFinite(value) && value > 0 && value <= assetDuration))].sort((a, b) => a - b);
  let markers = [...(timeline.markers || [])];
  if (beats.length) {
    markers = markers.filter(marker => !(marker.source === 'audio-beat' && marker.assetId === asset.id));
    let markerIndex = 0;
    for (const clip of added) {
      for (const beat of beats) {
        if (beat > clip.durationMs) continue;
        markerIndex += 1;
        markers.push({
          id: idFactory(),
          timeMs: clip.startMs + beat,
          label: `节拍 ${markerIndex}`,
          color: '#22c55e',
          source: 'audio-beat',
          assetId: asset.id
        });
      }
    }
    markers.sort((a, b) => Number(a.timeMs) - Number(b.timeMs));
  }
  return { ...timeline, tracks, markers };
}

function copyClipAnimation(value) {
  return value ? Object.fromEntries(Object.entries(value).map(([phase, setting]) => [phase, setting ? { ...setting } : setting])) : value;
}

function copyTracks(timeline) {
  return (timeline?.tracks || []).map(track => ({ ...track, settings: { ...(track.settings || {}) }, clips: (track.clips || []).map(clip => ({ ...clip, ...(clip.animation ? { animation: copyClipAnimation(clip.animation) } : {}) })) }));
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

function clipRange(clip) {
  const start = Math.max(0, Number(clip?.startMs) || 0);
  return { start, end: start + Math.max(1, Number(clip?.durationMs) || 1) };
}

function rangesOverlap(left, right) {
  return left.start < right.end && right.start < left.end;
}

function startIsFree(track, clipId, startMs, durationMs) {
  if (!['video', 'audio'].includes(track?.kind)) return true;
  const moving = { start: startMs, end: startMs + Math.max(1, durationMs) };
  return (track.clips || []).every(clip => clip.id === clipId || !rangesOverlap(moving, clipRange(clip)));
}

function nearestFreeStart(track, clipId, requestedStartMs, durationMs) {
  const requested = Math.max(0, Math.round(Number(requestedStartMs) || 0));
  if (startIsFree(track, clipId, requested, durationMs)) return requested;
  const candidates = [0];
  for (const clip of track.clips || []) {
    if (clip.id === clipId) continue;
    const range = clipRange(clip);
    candidates.push(Math.max(0, range.start - durationMs), range.end);
  }
  const valid = [...new Set(candidates.map(value => Math.round(value)))]
    .filter(value => startIsFree(track, clipId, value, durationMs))
    .sort((left, right) => Math.abs(left - requested) - Math.abs(right - requested) || left - right);
  if (!valid.length) throw new Error('轨道中没有可放置片段的空位');
  return valid[0];
}

export function duplicateTimelineClip(timeline, trackId, clipId, duplicateId) {
  const tracks = copyTracks(timeline);
  const track = tracks.find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  const clip = track.clips.find(item => item.id === clipId);
  if (!clip) throw new Error('片段不存在');
  if (!String(duplicateId || '').trim()) throw new Error('片段编号无效');
  const durationMs = Math.max(1, Math.round(Number(clip.durationMs) || 1));
  const preferredStart = (Number(clip.startMs) || 0) + durationMs;
  const startMs = nearestFreeStart(track, '', preferredStart, durationMs);
  const duplicate = {
    ...clip,
    id: duplicateId,
    name: `${clip.name || '片段'} 副本`,
    startMs,
    visualEffect: clip.visualEffect ? { ...clip.visualEffect } : clip.visualEffect,
    ...(clip.animation ? { animation: copyClipAnimation(clip.animation) } : {}),
    canvasTransform: clip.canvasTransform ? { ...clip.canvasTransform } : clip.canvasTransform,
    keyframes: (clip.keyframes || []).map(frame => ({ ...frame, values: { ...(frame.values || {}) } }))
  };
  track.clips.push(duplicate);
  track.clips.sort((left, right) => Number(left.startMs) - Number(right.startMs));
  return { ...timeline, tracks };
}

export function removeTimelineClip(timeline, trackId, clipId) {
  const tracks = copyTracks(timeline);
  const track = tracks.find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  if (!track.clips.some(item => item.id === clipId)) throw new Error('片段不存在');
  track.clips = track.clips.filter(item => item.id !== clipId);
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
    const hasSpeedCurve = Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2;
    const splitPosition = firstDuration / durationMs;
    const sourceSplit = hasSpeedCurve
      ? speedCurveSourceOffsetMs(firstDuration, durationMs, clip.speedCurve)
      : Math.round(firstDuration * rate);
    const firstSpeedCurve = hasSpeedCurve ? sliceSpeedCurve(clip.speedCurve, 0, splitPosition) : clip.speedCurve;
    const secondSpeedCurve = hasSpeedCurve ? sliceSpeedCurve(clip.speedCurve, splitPosition, 1) : clip.speedCurve;
    const firstKeyframes = (clip.keyframes || []).filter(frame => Number(frame.offsetMs) < firstDuration).map(frame => ({ ...frame, values: { ...(frame.values || {}) } }));
    const secondKeyframes = (clip.keyframes || []).filter(frame => Number(frame.offsetMs) >= firstDuration).map(frame => ({ ...frame, offsetMs: Number(frame.offsetMs) - firstDuration, values: { ...(frame.values || {}) } }));
    const secondDuration = durationMs - firstDuration;
    const first = {
      ...clip,
      text: firstText,
      name: String(firstText || clip.name).slice(0, 16),
      durationMs: firstDuration,
      sourceDurationMs: clip.sourceDurationMs,
      fadeInMs: Math.min(Math.max(0, Number(clip.fadeInMs) || 0), firstDuration),
      fadeOutMs: Math.min(Math.max(0, Number(clip.fadeOutMs) || 0), firstDuration),
      speedCurve: firstSpeedCurve,
      playbackRate: hasSpeedCurve ? speedCurveAverageRate(firstSpeedCurve) : rate,
      keyframes: firstKeyframes
    };
    const second = {
      ...clip,
      ...(clip.animation ? { animation: copyClipAnimation(clip.animation) } : {}),
      id: idFactory(),
      name: `${clip.name}（后半段）`,
      startMs: clip.startMs + firstDuration,
      sourceStartMs: (Number(clip.sourceStartMs) || 0) + sourceSplit,
      sourceDurationMs: clip.sourceDurationMs,
      durationMs: secondDuration,
      fadeInMs: Math.min(Math.max(0, Number(clip.fadeInMs) || 0), secondDuration),
      fadeOutMs: Math.min(Math.max(0, Number(clip.fadeOutMs) || 0), secondDuration),
      speedCurve: secondSpeedCurve,
      playbackRate: hasSpeedCurve ? speedCurveAverageRate(secondSpeedCurve) : rate,
      keyframes: secondKeyframes,
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

export function splitTimelineAtTime(timeline, timeMs, idFactory) {
  const splitTime = Math.round(Number(timeMs));
  if (!Number.isFinite(splitTime) || splitTime < 0) throw new Error('播放头时间无效');
  if (typeof idFactory !== 'function') throw new Error('片段编号生成器无效');
  let next = timeline;
  let splitCount = 0;
  const targets = (timeline?.tracks || []).flatMap(track => {
    if (track.locked) return [];
    return (track.clips || []).flatMap(clip => {
      const start = Number(clip.startMs) || 0;
      const end = start + (Number(clip.durationMs) || 0);
      return splitTime >= start + 100 && splitTime <= end - 100
        ? [{ trackId: track.id, clipId: clip.id, offsetMs: splitTime - start }]
        : [];
    });
  });
  for (const target of targets) {
    next = splitTimelineClipAt(next, target.trackId, target.clipId, target.offsetMs, idFactory);
    splitCount += 1;
  }
  if (!splitCount) throw new Error('播放头没有可分割片段');
  return { timeline: next, splitCount };
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
    const duration = Number(clip.durationMs) || 1;
    const hasSpeedCurve = Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2;
    if (edge === 'right') {
      const durationMs = clip.durationMs - amount;
      const speedCurve = hasSpeedCurve ? sliceSpeedCurve(clip.speedCurve, 0, durationMs / duration) : clip.speedCurve;
      return { ...clip, durationMs, speedCurve, playbackRate: hasSpeedCurve ? speedCurveAverageRate(speedCurve) : clip.playbackRate, keyframes: (clip.keyframes || []).filter(frame => Number(frame.offsetMs) <= durationMs).map(frame => ({ ...frame, values: { ...(frame.values || {}) } })) };
    }
    const rate = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    const sourceOffset = hasSpeedCurve ? speedCurveSourceOffsetMs(amount, duration, clip.speedCurve) : Math.round(amount * rate);
    const speedCurve = hasSpeedCurve ? sliceSpeedCurve(clip.speedCurve, amount / duration, 1) : clip.speedCurve;
    return { ...clip, startMs: clip.startMs + amount, sourceStartMs: clip.sourceStartMs + sourceOffset, durationMs: clip.durationMs - amount, speedCurve, playbackRate: hasSpeedCurve ? speedCurveAverageRate(speedCurve) : rate, keyframes: (clip.keyframes || []).filter(frame => Number(frame.offsetMs) >= amount).map(frame => ({ ...frame, offsetMs: Number(frame.offsetMs) - amount, values: { ...(frame.values || {}) } })) };
  });
}

export function resizeTimelineClipEdge(timeline, trackId, clipId, edge, edgeTimeMs) {
  if (!['left', 'right'].includes(edge)) throw new Error('不支持的裁剪方向');
  const requested = Math.round(Number(edgeTimeMs));
  if (!Number.isFinite(requested) || requested < 0) throw new Error('裁剪时间无效');
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  const clip = track.clips?.find(item => item.id === clipId);
  if (!clip) throw new Error('片段不存在');
  const start = Number(clip.startMs) || 0;
  const duration = Math.max(1, Number(clip.durationMs) || 1);
  const end = start + duration;
  if (edge === 'left') {
    if (requested < start) throw new Error('当前素材没有可向左延伸的内容');
    const amount = requested - start;
    if (duration - amount < 100) throw new Error('片段至少保留 100 毫秒');
    return amount === 0 ? { ...timeline, tracks: copyTracks(timeline) } : trimTimelineClipEdge(timeline, trackId, clipId, 'left', amount);
  }
  if (requested > end) throw new Error('当前素材没有可向右延伸的内容');
  const amount = end - requested;
  if (duration - amount < 100) throw new Error('片段至少保留 100 毫秒');
  return amount === 0 ? { ...timeline, tracks: copyTracks(timeline) } : trimTimelineClipEdge(timeline, trackId, clipId, 'right', amount);
}

export function setClipPlaybackRate(timeline, trackId, clipId, playbackRate) {
  const rate = Number(playbackRate);
  if (!Number.isFinite(rate) || rate < 0.25 || rate > 4) throw new Error('变速范围为 0.25x–4x');
  return updateClip(timeline, trackId, clipId, clip => {
    const previous = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
    const sourceSpan = Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2
      ? Number(clip.durationMs) * speedCurveAverageRate(clip.speedCurve)
      : Number(clip.durationMs) * previous;
    const durationMs = Math.max(1, Math.round(sourceSpan / rate));
    const scale = durationMs / Math.max(1, Number(clip.durationMs));
    const { speedCurve: _speedCurve, ...withoutSpeedCurve } = clip;
    return { ...withoutSpeedCurve, playbackRate: rate, durationMs, keyframes: (clip.keyframes || []).map(frame => ({ ...frame, offsetMs: Math.max(0, Math.round(Number(frame.offsetMs) * scale)), values: { ...(frame.values || {}) } })) };
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

  const durationMs = Math.max(1, Number(target.durationMs) || 1);
  let nextStart = requested;
  if (options.magnetic !== false) {
    const threshold = Math.max(0, Number(options.thresholdMs) || 150);
    const boundaries = [0];
    const playheadMs = Number(options.playheadMs);
    if (Number.isFinite(playheadMs) && playheadMs >= 0) boundaries.push(playheadMs);
    boundaries.push(...tracks.flatMap(track => track.clips.flatMap(clip => {
      if (track.id === trackId && clip.id === clipId) return [];
      const clipStart = Number(clip.startMs) || 0;
      return [clipStart, clipStart + (Number(clip.durationMs) || 0)];
    })));
    const snapStarts = boundaries.flatMap(boundary => [boundary, boundary - durationMs]).filter(value => value >= 0);
    const nearest = snapStarts.reduce((best, value) => best == null || Math.abs(value - requested) < Math.abs(best - requested) ? value : best, null);
    if (nearest != null && Math.abs(nearest - requested) <= threshold) nextStart = nearest;
  }

  nextStart = nearestFreeStart(targetTrack, clipId, nextStart, durationMs);

  const originalStart = Number(target.startMs) || 0;
  const delta = nextStart - originalStart;
  const linkGroupId = String(target.linkGroupId || '');
  for (const track of tracks) {
    for (const clip of track.clips) {
      if (track.id === trackId && clip.id === clipId) {
        clip.startMs = nextStart;
      } else if (options.linked === true && !track.locked && (linkGroupId ? clip.linkGroupId === linkGroupId : Number(clip.startMs) === originalStart)) {
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

function transitionCutForRightClip(timeline, videoTrackId, rightClipId) {
  const track = (timeline?.tracks || []).find(item => item.id === videoTrackId);
  if (!track || track.kind !== 'video') return { reasonZh: '视频轨道不存在' };
  if (track.locked) return { reasonZh: '轨道已锁定' };
  const ordered = [...(track.clips || [])].filter(clip => ['video', 'image'].includes(clip.kind)).sort((a, b) => Number(a.startMs) - Number(b.startMs));
  const index = ordered.findIndex(item => item.id === rightClipId);
  if (index <= 0) return { reasonZh: '转场需要放在两个相邻片段之间' };
  const left = ordered[index - 1];
  const right = ordered[index];
  const cutMs = Number(right.startMs) || 0;
  if ((Number(left.startMs) || 0) + (Number(left.durationMs) || 0) !== cutMs) return { reasonZh: '两个片段之间有空隙' };
  return { track, left, right, cutMs, maximumMs: Math.max(100, Math.floor(Math.min(Number(left.durationMs), Number(right.durationMs)) / 2)) };
}

function mirrorTransitionOnVideoClip(timeline, binding, resourceId, durationMs) {
  return {
    ...timeline,
    tracks: (timeline.tracks || []).map(track => track.id !== binding.track.id ? track : ({
      ...track,
      clips: (track.clips || []).map(clip => clip.id !== binding.right.id ? clip : ({ ...clip, transition: resourceId || '', transitionMs: resourceId ? durationMs : 0 }))
    }))
  };
}

export function addTransitionClip(timeline, videoTrackId, rightClipId, resource, durationMs, idFactory) {
  const resourceId = String(resource?.id || resource?.resourceId || '').trim();
  if (!resourceId.startsWith('transition-')) return operationResult(timeline, '转场资源无效');
  if (typeof idFactory !== 'function') return operationResult(timeline, '转场片段编号生成器无效');
  const binding = transitionCutForRightClip(timeline, videoTrackId, rightClipId);
  if (binding.reasonZh) return operationResult(timeline, binding.reasonZh);
  const safeDuration = Math.min(binding.maximumMs, Math.max(100, Math.round(Number(durationMs) || 500)));
  const tracks = copyTracks(timeline);
  let transitionTrack = tracks.find(track => track.kind === 'transition' && !track.settings?.compoundId);
  if (transitionTrack?.locked) return operationResult(timeline, '转场轨道已锁定');
  const existing = transitionTrack?.clips?.find(clip => clip.transitionTrackId === videoTrackId && clip.transitionRightClipId === rightClipId);
  const id = existing?.id || idFactory('transition');
  const transitionClip = {
    id,
    kind: 'transition',
    resourceId,
    transition: resourceId,
    name: String(resource?.name || resource?.labelZh || resourceId),
    startMs: Math.max(0, binding.cutMs - Math.floor(safeDuration / 2)),
    durationMs: safeDuration,
    transitionMs: safeDuration,
    sourceStartMs: 0,
    sourceDurationMs: 0,
    transitionTrackId: videoTrackId,
    transitionLeftClipId: binding.left.id,
    transitionRightClipId: binding.right.id
  };
  if (!transitionTrack) {
    transitionTrack = { id: `track-${id}`, kind: 'transition', name: '转场 1', order: tracks.length, muted: false, locked: false, clips: [] };
    tracks.push(transitionTrack);
  }
  transitionTrack.clips = [...(transitionTrack.clips || []).filter(clip => clip.id !== id), transitionClip].sort((a, b) => Number(a.startMs) - Number(b.startMs));
  const next = mirrorTransitionOnVideoClip({ ...timeline, tracks }, binding, resourceId, safeDuration);
  return operationResult(next);
}

export function moveTransitionClip(timeline, transitionTrackId, transitionClipId, requestedStartMs) {
  const transitionTrack = (timeline?.tracks || []).find(track => track.id === transitionTrackId && track.kind === 'transition');
  if (!transitionTrack) return operationResult(timeline, '转场轨道不存在');
  if (transitionTrack.locked) return operationResult(timeline, '转场轨道已锁定');
  const transitionClip = (transitionTrack.clips || []).find(clip => clip.id === transitionClipId);
  if (!transitionClip) return operationResult(timeline, '转场片段不存在');
  const videoTrack = (timeline.tracks || []).find(track => track.id === transitionClip.transitionTrackId && track.kind === 'video');
  if (!videoTrack || videoTrack.locked) return operationResult(timeline, videoTrack ? '轨道已锁定' : '视频轨道不存在');
  const ordered = [...(videoTrack.clips || [])].filter(clip => ['video', 'image'].includes(clip.kind)).sort((a, b) => Number(a.startMs) - Number(b.startMs));
  const cuts = ordered.slice(1).map((right, index) => ({ left: ordered[index], right, cutMs: Number(right.startMs) || 0 }))
    .filter(item => (Number(item.left.startMs) || 0) + (Number(item.left.durationMs) || 0) === item.cutMs);
  if (!cuts.length) return operationResult(timeline, '没有可以放置转场的相邻剪辑点');
  const requestedCenter = Math.max(0, Number(requestedStartMs) || 0) + Math.max(1, Number(transitionClip.durationMs) || 1) / 2;
  const target = cuts.sort((a, b) => Math.abs(a.cutMs - requestedCenter) - Math.abs(b.cutMs - requestedCenter))[0];
  let next = timeline;
  const previous = transitionCutForRightClip(next, transitionClip.transitionTrackId, transitionClip.transitionRightClipId);
  if (!previous.reasonZh) next = mirrorTransitionOnVideoClip(next, previous, '', 0);
  const binding = transitionCutForRightClip(next, transitionClip.transitionTrackId, target.right.id);
  if (binding.reasonZh) return operationResult(timeline, binding.reasonZh);
  const safeDuration = Math.min(binding.maximumMs, Math.max(100, Number(transitionClip.durationMs) || 500));
  next = {
    ...next,
    tracks: next.tracks.map(track => track.id !== transitionTrackId ? track : ({ ...track, clips: track.clips.map(clip => clip.id !== transitionClipId ? clip : ({
      ...clip,
      startMs: Math.max(0, binding.cutMs - Math.floor(safeDuration / 2)),
      durationMs: safeDuration,
      transitionMs: safeDuration,
      transitionLeftClipId: binding.left.id,
      transitionRightClipId: binding.right.id
    })) }))
  };
  return operationResult(mirrorTransitionOnVideoClip(next, binding, transitionClip.resourceId || transitionClip.transition, safeDuration));
}

export function resizeTransitionClip(timeline, transitionTrackId, transitionClipId, durationMs) {
  const track = (timeline?.tracks || []).find(item => item.id === transitionTrackId && item.kind === 'transition');
  if (!track) return operationResult(timeline, '转场轨道不存在');
  if (track.locked) return operationResult(timeline, '转场轨道已锁定');
  const clip = (track.clips || []).find(item => item.id === transitionClipId);
  if (!clip) return operationResult(timeline, '转场片段不存在');
  const binding = transitionCutForRightClip(timeline, clip.transitionTrackId, clip.transitionRightClipId);
  if (binding.reasonZh) return operationResult(timeline, binding.reasonZh);
  const safeDuration = Math.min(binding.maximumMs, Math.max(100, Math.round(Number(durationMs) || 500)));
  const next = {
    ...timeline,
    tracks: timeline.tracks.map(item => item.id !== transitionTrackId ? item : ({ ...item, clips: item.clips.map(candidate => candidate.id !== transitionClipId ? candidate : ({
      ...candidate,
      startMs: Math.max(0, binding.cutMs - Math.floor(safeDuration / 2)),
      durationMs: safeDuration,
      transitionMs: safeDuration
    })) }))
  };
  return operationResult(mirrorTransitionOnVideoClip(next, binding, clip.resourceId || clip.transition, safeDuration));
}

export function removeTransitionClip(timeline, transitionTrackId, transitionClipId) {
  const track = (timeline?.tracks || []).find(item => item.id === transitionTrackId && item.kind === 'transition');
  if (!track) return operationResult(timeline, '转场轨道不存在');
  if (track.locked) return operationResult(timeline, '转场轨道已锁定');
  const clip = (track.clips || []).find(item => item.id === transitionClipId);
  if (!clip) return operationResult(timeline, '转场片段不存在');
  let next = { ...timeline, tracks: timeline.tracks.map(item => item.id !== transitionTrackId ? item : ({ ...item, clips: item.clips.filter(candidate => candidate.id !== transitionClipId) })) };
  const binding = transitionCutForRightClip(next, clip.transitionTrackId, clip.transitionRightClipId);
  if (!binding.reasonZh) next = mirrorTransitionOnVideoClip(next, binding, '', 0);
  return operationResult(next);
}
import { sliceSpeedCurve, speedCurveAverageRate, speedCurveSourceOffsetMs } from './aicutSpeedCurve.js';
const timelineTrackKinds = ['subtitle', 'effect', 'transition', 'video', 'audio'];
const timelineTrackNames = { subtitle: '字幕 1', effect: '特效 1', transition: '转场 1', video: '视频 1', audio: '音频 1' };

export function orderedTimelineTracks(tracks = []) {
  return timelineTrackKinds.flatMap((kind) => {
    const matches = (tracks || []).filter((track) => track.kind === kind).sort((left, right) => Number(left.order || 0) - Number(right.order || 0));
    if (kind === 'transition') return matches;
    return matches.length ? matches : [{ id: `${kind}-display-track`, kind, name: timelineTrackNames[kind], clips: [], placeholder: true }];
  });
}
