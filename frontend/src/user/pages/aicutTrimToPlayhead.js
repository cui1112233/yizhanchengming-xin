import { resizeTimelineClipEdge } from './aicutTimeline.js';
import { compoundPathForClip } from './aicutCompoundModel.js';
import { speedCurveSourceOffsetMs } from './aicutSpeedCurve.js';
import { keyframeEasingSamples, sampleKeyframeEasing } from './aicutKeyframeEasing.js';

function retainedKeyframes(clip, from, to) {
  const frames = [...(clip.keyframes || [])].sort((a, b) => Number(a.offsetMs) - Number(b.offsetMs));
  if (!frames.length) return [];
  const base = { x: 0, y: 0, scaleX: 1, scaleY: 1, rotation: 0, opacity: 1, volume: Number(clip.volume) || 1, ...(clip.canvasTransform || {}) };
  const valuesAt = time => {
    const left = [...frames].reverse().find(frame => Number(frame.offsetMs) <= time) || frames[0];
    const right = frames.find(frame => Number(frame.offsetMs) >= time) || frames.at(-1);
    const span = Number(right.offsetMs) - Number(left.offsetMs);
    const ratio = span > 0 ? sampleKeyframeEasing(left.easing, (time - Number(left.offsetMs)) / span) : 0;
    const start = { ...base, ...left.values };
    const end = { ...base, ...right.values };
    return Object.fromEntries(Object.keys(base).map(key => {
      const leftValue = Number.isFinite(Number(start[key])) ? Number(start[key]) : base[key];
      const rightValue = Number.isFinite(Number(end[key])) ? Number(end[key]) : base[key];
      return [key, leftValue + (rightValue - leftValue) * ratio];
    }));
  };
  const result = new Map(frames.filter(frame => frame.offsetMs >= from && frame.offsetMs <= to)
    .map(frame => [Number(frame.offsetMs), { ...frame, values: { ...frame.values } }]));
  const usedIds = new Set(frames.map(frame => frame.id));
  const insert = (time, easing = { type: 'linear' }) => {
    let id = result.get(time)?.id;
    if (!id) {
      id = `${clip.id}-trim-${time}`;
      while (usedIds.has(id)) id += '-';
      usedIds.add(id);
    }
    result.set(time, { ...(result.get(time) || {}), id, offsetMs: time, values: valuesAt(time), easing });
  };
  for (const boundary of [from, to]) {
    if (result.has(boundary)) continue;
    const left = [...frames].reverse().find(frame => Number(frame.offsetMs) < boundary);
    const right = frames.find(frame => Number(frame.offsetMs) > boundary);
    if (!left || !right) { insert(boundary); continue; }
    const segmentStart = Math.max(from, Number(left.offsetMs));
    const segmentEnd = Math.min(to, Number(right.offsetMs));
    const easing = left.easing?.type === 'hold' ? { type: 'hold' } : { type: 'linear' };
    // Both preview and export evaluate Bezier easing through eight linear
    // segments. Keep those bends on the two surrounding integer milliseconds
    // so the retained motion is identical at every supported timeline instant.
    const times = new Set([segmentStart, segmentEnd]);
    const samples = !left.easing || ['linear', 'hold'].includes(left.easing.type) ? [] : keyframeEasingSamples(left.easing);
    for (const sample of samples) {
      const time = Number(left.offsetMs) + sample.x * (Number(right.offsetMs) - Number(left.offsetMs));
      for (const rounded of [Math.floor(time), Math.ceil(time)]) if (rounded > segmentStart && rounded < segmentEnd) times.add(rounded);
    }
    for (const time of times) insert(time, time === Number(right.offsetMs) ? right.easing : easing);
  }
  return [...result.values()].sort((a, b) => a.offsetMs - b.offsetMs).map(frame => ({ ...frame, offsetMs: Number(frame.offsetMs) - from }));
}

function sourceSpan(clip) {
  const duration = Number(clip.durationMs);
  return Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2
    ? speedCurveSourceOffsetMs(duration, duration, clip.speedCurve)
    : duration * (Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1);
}

export function trimTimelineClipToPlayhead(timeline, { trackId, clipId, edge, timeMs }) {
  const track = timeline?.tracks?.find(item => item.id === trackId);
  const clip = track?.clips?.find(item => item.id === clipId);
  if (!track || !clip) throw new Error('片段不存在');
  if (track.locked) throw new Error('轨道已锁定');
  if (!['left', 'right'].includes(edge)) throw new Error('不支持的裁剪方向');
  if (timeMs === null || timeMs === undefined || timeMs === '' || !Number.isFinite(Number(timeMs))) throw new Error('播放头时间无效');
  const requested = Math.round(Number(timeMs));
  const start = Number(clip.startMs) || 0;
  const duration = Number(clip.durationMs);
  const end = start + duration;
  if (!Number.isFinite(duration) || duration <= 0 || requested < start || requested > end) throw new Error('播放头必须位于片段内');
  if (requested === (edge === 'left' ? start : end)) return { timeline, changed: false };
  if (compoundPathForClip(clip).length) throw new Error('请进入复合片段后裁剪组内素材');
  const hasTransitionBinding = (timeline.tracks || []).some(item => (item.clips || []).some(candidate =>
    candidate.transitionTrackId === trackId && (candidate.transitionLeftClipId === clipId || candidate.transitionRightClipId === clipId)));
  if (track.kind === 'transition' || clip.kind === 'transition' || clip.transition || hasTransitionBinding) throw new Error('请先移除关联转场，再裁剪到播放头');

  // Normalize only the selected legacy clip, then reuse the drag-trim rules for
  // keyframes, speed curves and the minimum retained duration.
  const sourceStartMs = Math.max(0, Number(clip.sourceStartMs) || 0);
  const normalized = { ...clip, startMs: start, durationMs: duration, sourceStartMs };
  const single = { ...timeline, tracks: [{ ...track, clips: [normalized] }] };
  const resized = resizeTimelineClipEdge(single, trackId, clipId, edge, requested).tracks[0].clips[0];
  resized.keyframes = retainedKeyframes(normalized, edge === 'left' ? requested - start : 0, edge === 'right' ? requested - start : duration);
  const frozen = clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null;
  if (frozen) resized.sourceStartMs = sourceStartMs;
  else if (clip.reversed) {
    resized.sourceStartMs = edge === 'left' ? sourceStartMs : sourceStartMs + sourceSpan(normalized) - sourceSpan(resized);
  }
  const media = ['video', 'audio'].includes(clip.kind || track.kind);
  const sourceDuration = Number(clip.sourceDurationMs);
  if (media && sourceDuration > 0 && (frozen
    ? Number(clip.freezeSourceMs) < 0 || Number(clip.freezeSourceMs) >= sourceDuration
    : resized.sourceStartMs + sourceSpan(resized) > sourceDuration)) throw new Error('裁剪后的范围超出素材时长');
  return {
    changed: true,
    timeline: { ...timeline, tracks: timeline.tracks.map(item => item.id !== trackId ? item : {
      ...item, clips: item.clips.map(candidate => candidate.id === clipId ? resized : candidate)
    }) }
  };
}
