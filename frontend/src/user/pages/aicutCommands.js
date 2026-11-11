import {
  addTransitionClip,
  duplicateTimelineClip,
  mergeTimelineClipWithNext,
  moveTimelineClip,
  moveTransitionClip,
  planTimelineBGM,
  removeTimelineClip,
  removeTransitionClip,
  resizeTimelineClipEdge,
  resizeTransitionClip,
  setClipPlaybackRate,
  setClipTransition,
  splitTimelineAtTime,
  splitTimelineClip,
  splitTimelineClipAt
} from './aicutTimeline.js';
import { normalizeSpeedCurve, speedCurveAverageRate, speedCurveDurationMs, speedCurveSourceOffsetMs } from './aicutSpeedCurve.js';
import { applyCompoundPath, clipHasCompound, compoundEntries, compoundPathForClip, compoundRefForClip, outerCompoundRef, removeCompoundRefs, replaceCompoundRef, trackBelongsToCompound } from './aicutCompoundModel.js';
import { normalizeVisualProcessing, polygonArea, sampleBezierMaskPoints, sampleColorCurve } from './aicutVisualProcessing.js';
import { normalizeKeyframeEasing } from './aicutKeyframeEasing.js';
import { audioEffectValues } from './aicutAudioProcessing.js';
import { normalizePersonSegmentation } from './aicutPersonSegmentation.js';

function commandResult(timeline, extras = {}) {
  return { timeline, changed: true, selectionKey: '', metadata: {}, ...extras };
}

function findClip(timeline, key) {
  const [trackId, clipId] = String(key || '').split('::');
  const track = timeline?.tracks?.find(item => item.id === trackId);
  const clip = track?.clips?.find(item => item.id === clipId);
  if (!track || !clip) throw new Error('片段不存在');
  if (track.locked) throw new Error('轨道已锁定');
  return { track, clip, trackId, clipId };
}

function setLinkGroup(timeline, clipKeys, groupId) {
  const keys = [...new Set((clipKeys || []).filter(Boolean))];
  if (keys.length < 2) throw new Error('联动至少选择两个片段');
  const entries = keys.map(key => findClip(timeline, key));
  if (entries.every(({ clip }) => String(clip.linkGroupId || '') === groupId)) {
    return commandResult(timeline, { changed: false });
  }
  const targetKeys = new Set(keys);
  const tracks = timeline.tracks.map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => targetKeys.has(`${track.id}::${clip.id}`) ? { ...clip, linkGroupId: groupId } : { ...clip })
  }));
  return commandResult({ ...timeline, tracks }, { metadata: { linkGroupId: groupId, clipCount: entries.length } });
}

function clearLinkGroup(timeline, clipKeys) {
  const entries = [...new Set((clipKeys || []).filter(Boolean))].map(key => findClip(timeline, key));
  if (!entries.length) throw new Error('请选择要取消联动的片段');
  const groups = new Set(entries.map(({ clip }) => String(clip.linkGroupId || '')).filter(Boolean));
  if (!groups.size) return commandResult(timeline, { changed: false });
  const tracks = timeline.tracks.map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => groups.has(String(clip.linkGroupId || ''))
      ? Object.fromEntries(Object.entries({ ...clip }).filter(([key]) => key !== 'linkGroupId'))
      : { ...clip })
  }));
  return commandResult({ ...timeline, tracks }, { metadata: { unlinkedGroupCount: groups.size } });
}

function createCompound(timeline, clipKeys, compoundId, name) {
  const keys = [...new Set((clipKeys || []).filter(Boolean))];
  if (keys.length < 2) throw new Error('复合片段至少选择两个片段');
  const entries = keys.map(key => findClip(timeline, key));
  const id = String(compoundId || '').trim();
  if (!id) throw new Error('复合片段编号不能为空');
  const compoundName = String(name || '复合片段').trim() || '复合片段';
  if (entries.every(({ clip }) => outerCompoundRef(clip)?.id === id && outerCompoundRef(clip)?.name === compoundName)) {
    return commandResult(timeline, { changed: false, metadata: { compoundId: id, memberKeys: keys } });
  }
  const selectedKeys = new Set(keys);
  const existingIds = new Set(entries.flatMap(({ clip }) => compoundPathForClip(clip).map(item => item.id)));
  for (const existingId of existingIds) {
    const allMembers = compoundEntries(timeline, existingId).map(entry => entry.key);
    if (allMembers.some(key => !selectedKeys.has(key))) throw new Error('创建上层复合片段时必须完整选择子复合片段');
  }
  if ((timeline?.tracks || []).some(track => (track.clips || []).some(clip => clipHasCompound(clip, id)))) throw new Error('复合片段编号已存在');
  const targetKeys = new Set(keys);
  const tracks = timeline.tracks.map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => targetKeys.has(`${track.id}::${clip.id}`)
      ? applyCompoundPath(clip, [...compoundPathForClip(clip), { id, name: compoundName }])
      : { ...clip })
  }));
  return commandResult({ ...timeline, tracks }, { metadata: { compoundId: id, compoundName, memberKeys: keys } });
}

function moveCompound(timeline, compoundId, startMs) {
  const id = String(compoundId || '').trim();
  if (!id) throw new Error('复合片段编号不能为空');
  const entries = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      if (clipHasCompound(clip, id)) entries.push({ track, clip, key: `${track.id}::${clip.id}` });
    }
  }
  if (entries.length < 2) throw new Error('复合片段不存在或成员不足');
  if (entries.some(({ track }) => track.locked)) throw new Error('轨道已锁定');
  const targetStartMs = Number(startMs);
  if (!Number.isFinite(targetStartMs) || targetStartMs < 0) throw new Error('复合片段不能早于时间线零点');
  const originMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
  const deltaMs = Math.round(targetStartMs) - originMs;
  const memberKeys = entries.map(entry => entry.key);
  if (!deltaMs) return commandResult(timeline, { changed: false, metadata: { compoundId: id, memberKeys } });
  const memberSet = new Set(memberKeys);
  for (const { track, clip } of entries) {
    const nextStart = (Number(clip.startMs) || 0) + deltaMs;
    const nextEnd = nextStart + Math.max(0, Number(clip.durationMs) || 0);
    const overlaps = (track.clips || []).some(other => {
      if (memberSet.has(`${track.id}::${other.id}`)) return false;
      const otherStart = Number(other.startMs) || 0;
      const otherEnd = otherStart + Math.max(0, Number(other.durationMs) || 0);
      return nextStart < otherEnd && nextEnd > otherStart;
    });
    if (overlaps) throw new Error(`复合片段移动后与现有片段重叠（轨道“${track.name || track.id}”）`);
  }
  const tracks = timeline.tracks.map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => clipHasCompound(clip, id)
      ? { ...clip, startMs: (Number(clip.startMs) || 0) + deltaMs }
      : { ...clip })
  }));
  return commandResult({ ...timeline, tracks }, { metadata: { compoundId: id, memberKeys, deltaMs } });
}

function trimCompound(timeline, compoundId, edge, edgeTimeMs) {
  const id = String(compoundId || '').trim();
  if (!id) throw new Error('复合片段编号不能为空');
  if (!['left', 'right'].includes(edge)) throw new Error('不支持的裁剪方向');
  const requested = Math.round(Number(edgeTimeMs));
  if (!Number.isFinite(requested) || requested < 0) throw new Error('裁剪时间无效');
  const entries = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      if (clipHasCompound(clip, id)) entries.push({ track, clip, key: `${track.id}::${clip.id}` });
    }
  }
  if (entries.length < 2) throw new Error('复合片段不存在或成员不足');
  if (entries.some(({ track }) => track.locked)) throw new Error('轨道已锁定');
  const originMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
  const endMs = Math.max(...entries.map(({ clip }) => (Number(clip.startMs) || 0) + Math.max(1, Number(clip.durationMs) || 1)));
  if (requested < originMs || requested > endMs) throw new Error('裁剪时间无效');
  if ((edge === 'left' && requested === originMs) || (edge === 'right' && requested === endMs)) {
    return commandResult(timeline, { changed: false, metadata: { compoundId: id, memberKeys: entries.map(entry => entry.key) } });
  }
  const windowStartMs = edge === 'left' ? requested : originMs;
  const windowEndMs = edge === 'right' ? requested : endMs;
  if (windowEndMs - windowStartMs < 100) throw new Error('复合片段至少保留 100 毫秒');
  const survivors = entries.filter(({ clip }) => {
    const start = Number(clip.startMs) || 0;
    const end = start + Math.max(1, Number(clip.durationMs) || 1);
    return start < windowEndMs && end > windowStartMs;
  });
  if (survivors.length < 2) throw new Error('复合片段裁剪后至少保留两个成员');

  const survivorKeys = new Set(survivors.map(entry => entry.key));
  const removedMemberKeys = entries.filter(entry => !survivorKeys.has(entry.key)).map(entry => entry.key);
  const trimmedMemberKeys = [];
  let next = timeline;
  for (const { track, clip, key } of survivors) {
    const start = Number(clip.startMs) || 0;
    const end = start + Math.max(1, Number(clip.durationMs) || 1);
    if (start < windowStartMs) {
      next = resizeTimelineClipEdge(next, track.id, clip.id, 'left', windowStartMs);
      trimmedMemberKeys.push(key);
    }
    if (end > windowEndMs) {
      next = resizeTimelineClipEdge(next, track.id, clip.id, 'right', windowEndMs);
      if (!trimmedMemberKeys.includes(key)) trimmedMemberKeys.push(key);
    }
  }
  if (removedMemberKeys.length) {
    const removedSet = new Set(removedMemberKeys);
    next = {
      ...next,
      tracks: (next.tracks || []).map(track => ({
        ...track,
        clips: (track.clips || []).filter(clip => !removedSet.has(`${track.id}::${clip.id}`)).map(clip => ({ ...clip }))
      }))
    };
  }
  return commandResult(next, {
    metadata: {
      compoundId: id,
      memberKeys: survivors.map(entry => entry.key),
      removedMemberKeys,
      trimmedMemberKeys,
      edge,
      edgeTimeMs: requested
    }
  });
}

function splitCompound(timeline, compoundId, timeMs, rightCompoundId, idFactory) {
  const id = String(compoundId || '').trim();
  if (!id) throw new Error('复合片段编号不能为空');
  const splitTimeMs = Math.round(Number(timeMs));
  if (!Number.isFinite(splitTimeMs) || splitTimeMs < 0) throw new Error('分割时间无效');
  const entries = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      if (clipHasCompound(clip, id)) entries.push({ track, clip, key: `${track.id}::${clip.id}` });
    }
  }
  if (entries.length < 2) throw new Error('复合片段不存在或成员不足');
  if (entries.some(({ track }) => track.locked)) throw new Error('轨道已锁定');
  const originMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
  const endMs = Math.max(...entries.map(({ clip }) => (Number(clip.startMs) || 0) + Math.max(1, Number(clip.durationMs) || 1)));
  if (splitTimeMs <= originMs || splitTimeMs >= endMs) throw new Error('分割点必须在复合片段内部');
  const crosses = [];
  let leftCount = 0;
  let rightCount = 0;
  for (const entry of entries) {
    const start = Number(entry.clip.startMs) || 0;
    const end = start + Math.max(1, Number(entry.clip.durationMs) || 1);
    if (start < splitTimeMs && end > splitTimeMs) {
      crosses.push(entry);
      leftCount += 1;
      rightCount += 1;
    } else if (end <= splitTimeMs) leftCount += 1;
    else if (start >= splitTimeMs) rightCount += 1;
  }
  if (leftCount < 2 || rightCount < 2) throw new Error('复合片段分割后左右两侧都至少保留两个成员');
  const nextId = String(rightCompoundId || idFactory('compound')).trim();
  if (!nextId || nextId === id) throw new Error('右侧复合片段编号无效');
  const collides = (timeline?.tracks || []).some(track => (track.clips || []).some(clip => clipHasCompound(clip, nextId)));
  if (collides) throw new Error('右侧复合片段编号已存在');

  let next = timeline;
  for (const { track, clip } of crosses) {
    next = splitTimelineClipAt(next, track.id, clip.id, splitTimeMs - (Number(clip.startMs) || 0), () => idFactory('clip'));
  }
  const compoundName = String(compoundRefForClip(entries[0].clip, id)?.name || '复合片段');
  next = {
    ...next,
    tracks: (next.tracks || []).map(track => ({
      ...track,
      clips: (track.clips || []).map(clip => {
        if (!clipHasCompound(clip, id)) return { ...clip };
        if ((Number(clip.startMs) || 0) >= splitTimeMs) return replaceCompoundRef(clip, id, { id: nextId, name: `${compoundName}（续）` });
        return replaceCompoundRef(clip, id, { id, name: compoundName });
      })
    }))
  };
  const leftMemberKeys = [];
  const rightMemberKeys = [];
  for (const track of next.tracks || []) {
    for (const clip of track.clips || []) {
      if (clipHasCompound(clip, id)) leftMemberKeys.push(`${track.id}::${clip.id}`);
      if (clipHasCompound(clip, nextId)) rightMemberKeys.push(`${track.id}::${clip.id}`);
    }
  }
  return commandResult(next, {
    metadata: {
      splitTimeMs,
      splitMemberCount: crosses.length,
      leftCompoundId: id,
      rightCompoundId: nextId,
      leftMemberKeys,
      rightMemberKeys
    }
  });
}

function dissolveCompound(timeline, clipKeys, compoundId) {
  const entries = [...new Set((clipKeys || []).filter(Boolean))].map(key => findClip(timeline, key));
  if (!entries.length) throw new Error('请选择复合片段成员');
  const requestedId = String(compoundId || '').trim();
  if (requestedId && !entries.some(({ clip }) => clipHasCompound(clip, requestedId))) throw new Error('复合片段不存在');
  const ids = requestedId
    ? new Set([requestedId])
    : new Set(entries.map(({ clip }) => outerCompoundRef(clip)?.id).filter(Boolean));
  if (!ids.size) return commandResult(timeline, { changed: false });
  const memberKeys = [];
  const tracks = timeline.tracks.flatMap(track => {
    if (ids.has(String(track.settings?.compoundId || '')) && !(track.clips || []).length) return [];
    return [{
      ...track,
      settings: ids.has(String(track.settings?.compoundId || ''))
        ? { ...(track.settings || {}), compoundId: undefined, compoundName: undefined }
        : { ...(track.settings || {}) },
      clips: (track.clips || []).map(clip => {
        if (![...ids].some(id => clipHasCompound(clip, id))) return { ...clip };
        memberKeys.push(`${track.id}::${clip.id}`);
        return removeCompoundRefs(clip, ids);
      })
    }];
  }).map((track, order) => ({ ...track, order }));
  return commandResult({ ...timeline, tracks }, { metadata: { compoundIds: [...ids], memberKeys } });
}

function addCompoundTrack(timeline, command) {
  const compoundId = String(command.compoundId || '').trim();
  const entries = compoundEntries(timeline, compoundId);
  if (entries.length < 2) throw new Error('复合片段不存在或成员不足');
  const kind = String(command.kind || '');
  const labels = { video: '画面', audio: '声音', subtitle: '字幕', effect: '特效' };
  if (!labels[kind]) throw new Error('不支持的轨道类型');
  const trackId = String(command.trackId || '').trim();
  if (!trackId) throw new Error('轨道编号不能为空');
  if ((timeline.tracks || []).some(track => track.id === trackId)) throw new Error('轨道编号已存在');
  const compoundName = compoundRefForClip(entries[0].clip, compoundId)?.name || '复合片段';
  const number = (timeline.tracks || []).filter(track => track.kind === kind && trackBelongsToCompound(track, compoundId)).length + 1;
  const track = {
    id: trackId,
    kind,
    name: String(command.name || `组内${labels[kind]}${number}`).trim() || `组内${labels[kind]}${number}`,
    order: timeline.tracks.length,
    muted: false,
    locked: false,
    settings: { compoundId, compoundName },
    clips: []
  };
  return commandResult({ ...timeline, tracks: [...timeline.tracks.map(item => ({ ...item })), track] }, { metadata: { compoundId, trackId } });
}

function removeCompoundTrack(timeline, command) {
  const compoundId = String(command.compoundId || '').trim();
  const trackId = String(command.trackId || '').trim();
  const track = (timeline.tracks || []).find(item => item.id === trackId);
  if (!track || !trackBelongsToCompound(track, compoundId)) throw new Error('组内轨道不存在');
  if (track.locked) throw new Error('轨道已锁定');
  if ((track.clips || []).length) throw new Error('轨道中还有片段，不能删除');
  const tracks = timeline.tracks.filter(item => item.id !== trackId).map((item, order) => ({ ...item, order }));
  return commandResult({ ...timeline, tracks }, { metadata: { compoundId, trackId } });
}

function reorderCompoundTracks(timeline, command) {
  const compoundId = String(command.compoundId || '').trim();
  const entries = compoundEntries(timeline, compoundId);
  if (entries.length < 2) throw new Error('复合片段不存在或成员不足');
  const compoundTracks = (timeline.tracks || []).filter(track =>
    trackBelongsToCompound(track, compoundId) || (track.clips || []).some(clip => clipHasCompound(clip, compoundId))
  );
  const availableIds = compoundTracks.map(track => track.id);
  const availableSet = new Set(availableIds);
  const orderedTrackIds = (command.orderedTrackIds || []).map(id => String(id || '').trim()).filter(Boolean);
  if (new Set(orderedTrackIds).size !== orderedTrackIds.length) throw new Error('组内轨道顺序不能重复');
  const foreignId = orderedTrackIds.find(id => !availableSet.has(id));
  if (foreignId) throw new Error(`轨道“${foreignId}”不属于当前复合片段`);
  if (orderedTrackIds.length !== availableIds.length || availableIds.some(id => !orderedTrackIds.includes(id))) {
    throw new Error('轨道顺序必须包含全部组内轨道');
  }
  if (orderedTrackIds.every((id, index) => id === availableIds[index])) {
    return commandResult(timeline, { changed: false, metadata: { compoundId, orderedTrackIds } });
  }
  if (compoundTracks.some(track => track.locked)) throw new Error('轨道已锁定，不能调整顺序');
  const tracksById = new Map(compoundTracks.map(track => [track.id, track]));
  let compoundIndex = 0;
  const tracks = (timeline.tracks || []).map(track => availableSet.has(track.id)
    ? { ...tracksById.get(orderedTrackIds[compoundIndex++]) }
    : { ...track }
  ).map((track, order) => ({ ...track, order }));
  return commandResult({ ...timeline, tracks }, { metadata: { compoundId, orderedTrackIds } });
}

function compoundMemberTrackMove(timeline, command) {
  const compoundId = String(command.compoundId || '').trim();
  const source = findClip(timeline, command.memberKey);
  if (!clipHasCompound(source.clip, compoundId)) throw new Error('片段不属于当前复合片段');
  const targetTrackId = String(command.targetTrackId || '').trim();
  const targetTrack = (timeline.tracks || []).find(track => track.id === targetTrackId);
  if (!targetTrack) throw new Error('目标轨道不存在');
  if (targetTrack.locked) throw new Error('轨道已锁定');
  const targetBelongs = trackBelongsToCompound(targetTrack, compoundId)
    || (targetTrack.clips || []).some(clip => clipHasCompound(clip, compoundId));
  if (!targetBelongs) throw new Error('目标轨道不属于当前复合片段');
  const compatible = targetTrack.kind === 'video'
    ? ['video', 'image'].includes(source.clip.kind)
    : targetTrack.kind === source.clip.kind;
  if (!compatible) throw new Error('片段与目标轨道类型不匹配');
  const startMs = Math.round(Number(command.startMs));
  if (!Number.isFinite(startMs) || startMs < 0) throw new Error('片段开始时间无效');
  const durationMs = Math.max(1, Number(source.clip.durationMs) || 1);
  if (source.trackId === targetTrackId) {
    if ((Number(source.clip.startMs) || 0) === startMs) {
      return commandResult(timeline, { changed: false, selectionKey: `${targetTrackId}::${source.clipId}`, metadata: { compoundId, sourceTrackId: source.trackId, targetTrackId } });
    }
    const moved = moveTimelineClip(timeline, source.trackId, source.clipId, startMs, { magnetic: false, linked: false });
    return commandResult(moved, { selectionKey: `${targetTrackId}::${source.clipId}`, metadata: { compoundId, sourceTrackId: source.trackId, targetTrackId } });
  }
  if ((targetTrack.clips || []).some(clip => clip.id === source.clipId)) throw new Error('目标轨道中片段编号已存在');
  const endMs = startMs + durationMs;
  const overlaps = (targetTrack.clips || []).some(clip => {
    const clipStart = Number(clip.startMs) || 0;
    const clipEnd = clipStart + Math.max(1, Number(clip.durationMs) || 1);
    return startMs < clipEnd && endMs > clipStart;
  });
  if (overlaps) throw new Error('目标轨道已有片段占用该时间段');
  const movedClip = { ...source.clip, startMs };
  const tracks = (timeline.tracks || []).map(track => {
    if (track.id === source.trackId) return { ...track, clips: (track.clips || []).filter(clip => clip.id !== source.clipId).map(clip => ({ ...clip })) };
    if (track.id === targetTrackId) return { ...track, clips: [...(track.clips || []).map(clip => ({ ...clip })), movedClip].sort((a, b) => (Number(a.startMs) || 0) - (Number(b.startMs) || 0)) };
    return { ...track, clips: (track.clips || []).map(clip => ({ ...clip })) };
  });
  return commandResult({ ...timeline, tracks }, {
    selectionKey: `${targetTrackId}::${source.clipId}`,
    metadata: { compoundId, sourceTrackId: source.trackId, targetTrackId, clipId: source.clipId }
  });
}

function compoundMemberTransfer(timeline, command) {
  const sourceCompoundId = String(command.sourceCompoundId || '').trim();
  const targetCompoundId = String(command.targetCompoundId || '').trim();
  if (!sourceCompoundId || !targetCompoundId) throw new Error('来源和目标复合片段不能为空');
  if (sourceCompoundId === targetCompoundId) throw new Error('来源和目标复合片段不能相同');
  const source = findClip(timeline, command.memberKey);
  if (!clipHasCompound(source.clip, sourceCompoundId)) throw new Error('片段不属于来源复合片段');
  const sourceEntries = compoundEntries(timeline, sourceCompoundId);
  const targetEntries = compoundEntries(timeline, targetCompoundId);
  if (targetEntries.length < 2) throw new Error('目标复合片段不存在或成员不足');
  const targetPath = compoundPathForClip(targetEntries[0].clip);
  const targetIndex = targetPath.findIndex(ref => ref.id === targetCompoundId);
  if (targetIndex < 0) throw new Error('目标复合路径无效');
  const nextPath = targetPath.slice(targetIndex);
  const remainsInSource = nextPath.some(ref => ref.id === sourceCompoundId);
  if (!remainsInSource && sourceEntries.length <= 2) throw new Error('来源复合片段至少保留两个成员');
  const targetTrackId = String(command.targetTrackId || '').trim();
  const targetTrack = (timeline.tracks || []).find(track => track.id === targetTrackId);
  if (!targetTrack) throw new Error('目标轨道不存在');
  if (targetTrack.locked) throw new Error('轨道已锁定');
  const targetBelongs = trackBelongsToCompound(targetTrack, targetCompoundId)
    || (targetTrack.clips || []).some(clip => clipHasCompound(clip, targetCompoundId));
  if (!targetBelongs) throw new Error('目标轨道不属于目标复合片段');
  const compatible = targetTrack.kind === 'video'
    ? ['video', 'image'].includes(source.clip.kind)
    : targetTrack.kind === source.clip.kind;
  if (!compatible) throw new Error('片段与目标轨道类型不匹配');
  const startMs = Math.round(Number(command.startMs));
  if (!Number.isFinite(startMs) || startMs < 0) throw new Error('片段开始时间无效');
  if (source.trackId !== targetTrackId && (targetTrack.clips || []).some(clip => clip.id === source.clipId)) {
    throw new Error('目标轨道中片段编号已存在');
  }
  const durationMs = Math.max(1, Number(source.clip.durationMs) || 1);
  const endMs = startMs + durationMs;
  const overlaps = (targetTrack.clips || []).some(clip => {
    if (targetTrackId === source.trackId && clip.id === source.clipId) return false;
    const clipStart = Number(clip.startMs) || 0;
    const clipEnd = clipStart + Math.max(1, Number(clip.durationMs) || 1);
    return startMs < clipEnd && endMs > clipStart;
  });
  if (overlaps) throw new Error('目标轨道已有片段占用该时间段');
  const movedClip = applyCompoundPath({ ...source.clip, startMs }, nextPath);
  const tracks = (timeline.tracks || []).map(track => {
    if (source.trackId === targetTrackId && track.id === source.trackId) {
      return { ...track, clips: (track.clips || []).map(clip => clip.id === source.clipId ? movedClip : { ...clip }).sort((a, b) => (Number(a.startMs) || 0) - (Number(b.startMs) || 0)) };
    }
    if (track.id === source.trackId) return { ...track, clips: (track.clips || []).filter(clip => clip.id !== source.clipId).map(clip => ({ ...clip })) };
    if (track.id === targetTrackId) return { ...track, clips: [...(track.clips || []).map(clip => ({ ...clip })), movedClip].sort((a, b) => (Number(a.startMs) || 0) - (Number(b.startMs) || 0)) };
    return { ...track, clips: (track.clips || []).map(clip => ({ ...clip })) };
  });
  return commandResult({ ...timeline, tracks }, {
    selectionKey: `${targetTrackId}::${source.clipId}`,
    metadata: { sourceCompoundId, targetCompoundId, sourceTrackId: source.trackId, targetTrackId, clipId: source.clipId }
  });
}

function compoundBlockTransfer(timeline, command) {
  const sourceCompoundId = String(command.sourceCompoundId || '').trim();
  const targetCompoundId = String(command.targetCompoundId || '').trim();
  if (!sourceCompoundId || !targetCompoundId) throw new Error('来源和目标复合片段不能为空');
  if (sourceCompoundId === targetCompoundId) throw new Error('来源和目标复合片段不能相同');
  const sourceEntries = compoundEntries(timeline, sourceCompoundId);
  const targetEntries = compoundEntries(timeline, targetCompoundId);
  if (sourceEntries.length < 2) throw new Error('来源复合片段不存在或成员不足');
  if (!targetEntries.length) throw new Error('目标复合片段不存在或成员不足');
  if (sourceEntries.some(({ track }) => track.locked)) throw new Error('轨道已锁定');
  const sourcePath = compoundPathForClip(sourceEntries[0].clip);
  const sourceIndex = sourcePath.findIndex(ref => ref.id === sourceCompoundId);
  const currentParentId = sourceIndex >= 0 ? sourcePath[sourceIndex + 1]?.id || '' : '';
  if (currentParentId === targetCompoundId) {
    return commandResult(timeline, { changed: false, metadata: { sourceCompoundId, targetCompoundId, memberKeys: sourceEntries.map(entry => entry.key) } });
  }
  const targetPath = compoundPathForClip(targetEntries[0].clip);
  const targetIndex = targetPath.findIndex(ref => ref.id === targetCompoundId);
  if (targetIndex < 0) throw new Error('目标复合路径无效');
  const targetSuffix = targetPath.slice(targetIndex);
  if (targetSuffix.some(ref => ref.id === sourceCompoundId)) throw new Error('复合块不能拖入自己的后代');
  if (targetEntries.length < 2) throw new Error('目标复合片段不存在或成员不足');
  const memberKeys = new Set(sourceEntries.map(entry => entry.key));
  const tracks = (timeline.tracks || []).map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => {
      const key = `${track.id}::${clip.id}`;
      if (!memberKeys.has(key)) return { ...clip };
      const path = compoundPathForClip(clip);
      const index = path.findIndex(ref => ref.id === sourceCompoundId);
      if (index < 0) return { ...clip };
      return applyCompoundPath(clip, [...path.slice(0, index + 1), ...targetSuffix]);
    })
  }));
  return commandResult({ ...timeline, tracks }, {
    metadata: { sourceCompoundId, targetCompoundId, memberKeys: [...memberKeys] }
  });
}

function compoundMembersTransfer(timeline, command) {
  const sourceCompoundId = String(command.sourceCompoundId || '').trim();
  const targetCompoundId = String(command.targetCompoundId || '').trim();
  if (!sourceCompoundId || !targetCompoundId) throw new Error('来源和目标复合片段不能为空');
  if (sourceCompoundId === targetCompoundId) throw new Error('来源和目标复合片段不能相同');
  const memberKeys = (command.memberKeys || []).map(key => String(key || '').trim()).filter(Boolean);
  if (memberKeys.length < 2) throw new Error('批量迁移至少选择两个成员');
  if (new Set(memberKeys).size !== memberKeys.length) throw new Error('批量迁移成员不能重复');
  const selected = new Set(memberKeys);
  const entries = memberKeys.map(key => findClip(timeline, key));
  if (entries.some(({ clip }) => !clipHasCompound(clip, sourceCompoundId))) throw new Error('所选片段不属于来源复合片段');
  const sourceEntries = compoundEntries(timeline, sourceCompoundId);
  const targetEntries = compoundEntries(timeline, targetCompoundId);
  if (targetEntries.length < 2) throw new Error('目标复合片段不存在或成员不足');
  const targetPath = compoundPathForClip(targetEntries[0].clip);
  const targetIndex = targetPath.findIndex(ref => ref.id === targetCompoundId);
  if (targetIndex < 0) throw new Error('目标复合路径无效');
  const targetSuffix = targetPath.slice(targetIndex);
  if (targetSuffix.some(ref => ref.id === sourceCompoundId)) throw new Error('批量成员不能迁入来源复合片段的后代');
  const nestedIds = new Set();
  for (const { clip } of entries) {
    const path = compoundPathForClip(clip);
    const sourceIndex = path.findIndex(ref => ref.id === sourceCompoundId);
    path.slice(0, sourceIndex).forEach(ref => nestedIds.add(ref.id));
  }
  for (const nestedId of nestedIds) {
    const nestedKeys = compoundEntries(timeline, nestedId).map(entry => entry.key);
    if (nestedKeys.some(key => !selected.has(key))) throw new Error('批量迁移时必须完整选择子复合片段');
  }
  const remainingCount = sourceEntries.filter(entry => !selected.has(entry.key)).length;
  if (remainingCount === 1) throw new Error('来源复合片段不能只剩一个成员');
  if (command.autoMapTracks) {
    const targetName = targetSuffix[0]?.name || '复合片段';
    const labels = { video: '画面', audio: '声音', subtitle: '字幕', effect: '特效' };
    const compatible = (track, clip) => track.kind === 'video'
      ? ['video', 'image'].includes(clip.kind)
      : track.kind === clip.kind;
    const canPlace = (track, clip) => {
      if (track.locked || !compatible(track, clip)) return false;
      if ((track.clips || []).some(item => item.id === clip.id)) return false;
      const startMs = Number(clip.startMs) || 0;
      const endMs = startMs + Math.max(1, Number(clip.durationMs) || 1);
      return !(track.clips || []).some(item => {
        const itemStart = Number(item.startMs) || 0;
        const itemEnd = itemStart + Math.max(1, Number(item.durationMs) || 1);
        return startMs < itemEnd && endMs > itemStart;
      });
    };
    let tracks = (timeline.tracks || []).map(track => ({
      ...track,
      settings: { ...(track.settings || {}) },
      clips: (track.clips || []).filter(clip => !selected.has(`${track.id}::${clip.id}`)).map(clip => ({ ...clip }))
    }));
    const createdTrackIds = [];
    const targetTrackByMember = {};
    const nextTrackId = (kind) => {
      const base = `compound-${targetCompoundId}-${kind}-auto-`;
      let index = 1;
      while (tracks.some(track => track.id === `${base}${index}`)) index += 1;
      return `${base}${index}`;
    };
    for (const [entryIndex, entry] of entries.entries()) {
      const memberKey = memberKeys[entryIndex];
      const path = compoundPathForClip(entry.clip);
      const sourceIndex = path.findIndex(ref => ref.id === sourceCompoundId);
      const movedClip = applyCompoundPath(entry.clip, [...path.slice(0, sourceIndex), ...targetSuffix]);
      const candidates = tracks
        .filter(track => trackBelongsToCompound(track, targetCompoundId)
          || (track.clips || []).some(clip => clipHasCompound(clip, targetCompoundId)))
        .sort((left, right) => Number(trackBelongsToCompound(right, targetCompoundId)) - Number(trackBelongsToCompound(left, targetCompoundId)));
      const requestedTarget = String(command.targetTrackByMember?.[memberKey] || '').trim();
      let targetTrack = null;
      if (requestedTarget && requestedTarget !== '__auto__' && requestedTarget !== '__new__') {
        targetTrack = tracks.find(track => track.id === requestedTarget);
        if (!targetTrack) throw new Error('指定的目标轨道不存在');
        const belongs = trackBelongsToCompound(targetTrack, targetCompoundId)
          || (targetTrack.clips || []).some(clip => clipHasCompound(clip, targetCompoundId));
        if (!belongs) throw new Error('指定的目标轨道不属于目标复合片段');
        if (targetTrack.locked) throw new Error('指定的目标轨道已锁定');
        if (!compatible(targetTrack, movedClip)) throw new Error('指定的目标轨道类型不匹配');
        if (!canPlace(targetTrack, movedClip)) throw new Error('指定的目标轨道在该时间段已有片段');
      } else if (requestedTarget !== '__new__') {
        targetTrack = candidates.find(track => canPlace(track, movedClip));
      }
      if (!targetTrack) {
        const kind = entry.track.kind === 'video' || entry.track.kind === 'audio' || entry.track.kind === 'subtitle' || entry.track.kind === 'effect'
          ? entry.track.kind
          : movedClip.kind === 'image' ? 'video' : movedClip.kind;
        if (!labels[kind]) throw new Error(`无法为“${entry.clip.name || entry.clipId}”创建兼容的目标轨道`);
        const sameKindCount = tracks.filter(track => track.kind === kind && trackBelongsToCompound(track, targetCompoundId)).length;
        targetTrack = {
          id: nextTrackId(kind),
          kind,
          name: `组内${labels[kind]}${sameKindCount + 1}`,
          order: tracks.length,
          muted: false,
          locked: false,
          settings: { compoundId: targetCompoundId, compoundName: targetName },
          clips: []
        };
        tracks.push(targetTrack);
        createdTrackIds.push(targetTrack.id);
      }
      targetTrack.clips.push(movedClip);
      targetTrack.clips.sort((left, right) => (Number(left.startMs) || 0) - (Number(right.startMs) || 0));
      targetTrackByMember[memberKey] = targetTrack.id;
    }
    if (remainingCount === 0) {
      tracks = tracks.flatMap(track => {
        if (!trackBelongsToCompound(track, sourceCompoundId)) return [track];
        if (!(track.clips || []).length) return [];
        const hasTargetMember = track.clips.some(clip => clipHasCompound(clip, targetCompoundId));
        return [{
          ...track,
          settings: hasTargetMember
            ? { ...(track.settings || {}), compoundId: targetCompoundId, compoundName: targetName }
            : { ...(track.settings || {}), compoundId: undefined, compoundName: undefined }
        }];
      });
    }
    tracks = tracks.map((track, order) => ({ ...track, order }));
    const firstTargetTrackId = targetTrackByMember[memberKeys[0]];
    const firstClipId = entries[0].clipId;
    return commandResult({ ...timeline, tracks }, {
      selectionKey: `${firstTargetTrackId}::${firstClipId}`,
      metadata: { sourceCompoundId, targetCompoundId, memberKeys, targetTrackByMember, createdTrackIds }
    });
  }
  let tracks = (timeline.tracks || []).map(track => ({
    ...track,
    clips: (track.clips || []).map(clip => {
      const key = `${track.id}::${clip.id}`;
      if (!selected.has(key)) return { ...clip };
      const path = compoundPathForClip(clip);
      const sourceIndex = path.findIndex(ref => ref.id === sourceCompoundId);
      return applyCompoundPath(clip, [...path.slice(0, sourceIndex), ...targetSuffix]);
    })
  }));
  if (remainingCount === 0) {
    const targetName = targetSuffix[0]?.name || '复合片段';
    tracks = tracks.flatMap(track => {
      if (!trackBelongsToCompound(track, sourceCompoundId)) return [track];
      if (!(track.clips || []).length) return [];
      const hasTargetMember = track.clips.some(clip => clipHasCompound(clip, targetCompoundId));
      return [{
        ...track,
        settings: hasTargetMember
          ? { ...(track.settings || {}), compoundId: targetCompoundId, compoundName: targetName }
          : { ...(track.settings || {}), compoundId: undefined, compoundName: undefined }
      }];
    }).map((track, order) => ({ ...track, order }));
  }
  return commandResult({ ...timeline, tracks }, {
    selectionKey: memberKeys[0],
    metadata: { sourceCompoundId, targetCompoundId, memberKeys }
  });
}

export function createAICutClipboard(timeline, clipKeys) {
  const entries = [...new Set((clipKeys || []).filter(Boolean))].map(key => findClip(timeline, key));
  if (!entries.length) throw new Error('请先选择要复制的片段');
  const originMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
  return {
    version: 1,
    items: entries.map(({ track, clip }) => ({
      trackId: track.id,
      trackKind: track.kind,
      relativeStartMs: (Number(clip.startMs) || 0) - originMs,
      clip: { ...clip }
    }))
  };
}

function deleteMany(timeline, clipKeys, ripple) {
  const entries = [...new Set((clipKeys || []).filter(Boolean))].map(key => findClip(timeline, key));
  if (!entries.length) throw new Error('请先选择要删除的片段');
  let working = timeline;
  const transitionEntries = entries.filter(({ track, clip }) => track.kind === 'transition' || clip.kind === 'transition');
  for (const entry of transitionEntries) {
    const result = removeTransitionClip(working, entry.trackId, entry.clipId);
    if (!result.ok) throw new Error(result.reasonZh || '转场删除失败');
    working = result.timeline;
  }
  const regularEntries = entries.filter(({ track, clip }) => track.kind !== 'transition' && clip.kind !== 'transition');
  if (!regularEntries.length) return commandResult(working, { metadata: { deletedCount: transitionEntries.length, ripple: false } });
  const selected = new Set(regularEntries.map(({ trackId, clipId }) => `${trackId}::${clipId}`));
  const intervals = new Map();
  for (const { trackId, clip } of regularEntries) {
    const list = intervals.get(trackId) || [];
    list.push({ start: Number(clip.startMs) || 0, end: (Number(clip.startMs) || 0) + Math.max(0, Number(clip.durationMs) || 0) });
    intervals.set(trackId, list);
  }
  const tracks = working.tracks.map(track => {
    const removed = (intervals.get(track.id) || []).sort((a, b) => a.start - b.start);
    const clips = (track.clips || []).filter(clip => !selected.has(`${track.id}::${clip.id}`)).map(clip => {
      if (!ripple || !removed.length) return { ...clip };
      const startMs = Number(clip.startMs) || 0;
      const shift = removed.filter(interval => interval.end <= startMs).reduce((sum, interval) => sum + interval.end - interval.start, 0);
      return { ...clip, startMs: Math.max(0, startMs - shift) };
    });
    return { ...track, clips };
  });
  return commandResult({ ...working, tracks }, { metadata: { deletedCount: entries.length, ripple: Boolean(ripple) } });
}

function pasteClipboard(timeline, clipboard, atMs, idFactory) {
  if (!Array.isArray(clipboard?.items) || !clipboard.items.length) throw new Error('剪贴板为空');
  const baseStart = Math.max(0, Math.round(Number(atMs) || 0));
  const tracksById = new Map(timeline.tracks.map(track => [track.id, track]));
  for (const item of clipboard.items) {
    const track = tracksById.get(item.trackId);
    if (!track) throw new Error(`粘贴目标轨道不存在：${item.trackId}`);
    if (track.locked) throw new Error(`轨道已锁定：${track.name || track.id}`);
  }
  const groupMap = new Map();
  const compoundMap = new Map();
  const pastedKeys = [];
  const plannedItems = clipboard.items.map(item => ({ ...item, nextId: idFactory('clip') }));
  for (const item of plannedItems) {
    const originalGroup = String(item.clip?.linkGroupId || '');
    if (originalGroup && !groupMap.has(originalGroup)) groupMap.set(originalGroup, idFactory('link'));
    for (const ref of compoundPathForClip(item.clip)) {
      if (!compoundMap.has(ref.id)) compoundMap.set(ref.id, idFactory('compound'));
    }
  }
  const tracks = timeline.tracks.map(track => ({ ...track, clips: (track.clips || []).map(clip => ({ ...clip })) }));
  const mutableTracks = new Map(tracks.map(track => [track.id, track]));
  for (const item of plannedItems) {
    const id = item.nextId;
    const originalGroup = String(item.clip?.linkGroupId || '');
    const originalPath = compoundPathForClip(item.clip);
    let clip = {
      ...item.clip,
      id,
      name: `${String(item.clip?.name || '片段').replace(/ 副本$/, '')} 副本`,
      startMs: baseStart + Math.max(0, Number(item.relativeStartMs) || 0),
      ...(originalGroup ? { linkGroupId: groupMap.get(originalGroup) } : {})
    };
    if (originalPath.length) clip = applyCompoundPath(clip, originalPath.map(ref => ({ ...ref, id: compoundMap.get(ref.id) })));
    const track = mutableTracks.get(item.trackId);
    const overlaps = track.clips.some(existing => {
      const start = Number(existing.startMs) || 0;
      const end = start + Math.max(0, Number(existing.durationMs) || 0);
      return clip.startMs < end && start < clip.startMs + Math.max(0, Number(clip.durationMs) || 0);
    });
    if (overlaps) throw new Error(`粘贴位置与“${track.name || track.id}”现有片段重叠`);
    track.clips.push(clip);
    track.clips.sort((a, b) => Number(a.startMs) - Number(b.startMs));
    pastedKeys.push(`${track.id}::${id}`);
  }
  return commandResult({ ...timeline, tracks }, { selectionKey: pastedKeys[0] || '', metadata: { pastedKeys } });
}

function rippleResize(timeline, trackId, clipId, edge, edgeTimeMs) {
  const entry = findClip(timeline, `${trackId}::${clipId}`);
  const oldStart = Number(entry.clip.startMs) || 0;
  const oldEnd = oldStart + Math.max(0, Number(entry.clip.durationMs) || 0);
  const resized = resizeTimelineClipEdge(timeline, trackId, clipId, edge, edgeTimeMs);
  const resizedClip = resized.tracks.find(track => track.id === trackId).clips.find(clip => clip.id === clipId);
  const newStart = Number(resizedClip.startMs) || 0;
  const newEnd = newStart + Math.max(0, Number(resizedClip.durationMs) || 0);
  const shift = edge === 'left' ? newStart - oldStart : oldEnd - newEnd;
  if (shift <= 0) return resized;
  return {
    ...resized,
    tracks: resized.tracks.map(track => track.id !== trackId ? track : {
      ...track,
      clips: track.clips.map(clip => {
        if (clip.id === clipId) return edge === 'left' ? { ...clip, startMs: oldStart } : clip;
        if ((Number(clip.startMs) || 0) >= oldEnd) return { ...clip, startMs: Math.max(0, Number(clip.startMs) - shift) };
        return clip;
      })
    })
  };
}

function addTimelineMarker(timeline, command, idFactory) {
  const timeMs = Math.max(0, Math.round(Number(command.timeMs) || 0));
  const label = String(command.label || '标记').trim().slice(0, 80) || '标记';
  const color = /^#[0-9a-f]{6}$/i.test(String(command.color || '')) ? String(command.color) : '#f59e0b';
  const markers = [...(timeline.markers || [])].map(marker => ({ ...marker }));
  const existing = markers.find(marker => Number(marker.timeMs) === timeMs);
  if (existing) {
    existing.label = label;
    existing.color = color;
    return commandResult({ ...timeline, markers: markers.sort((a, b) => a.timeMs - b.timeMs) }, { metadata: { markerId: existing.id } });
  }
  const marker = { id: idFactory('marker'), timeMs, label, color };
  markers.push(marker);
  markers.sort((a, b) => a.timeMs - b.timeMs);
  return commandResult({ ...timeline, markers }, { metadata: { markerId: marker.id } });
}

function removeTimelineMarker(timeline, markerId) {
  const markers = timeline.markers || [];
  if (!markers.some(marker => marker.id === markerId)) throw new Error('标记点不存在');
  return commandResult({ ...timeline, markers: markers.filter(marker => marker.id !== markerId).map(marker => ({ ...marker })) });
}

function updateClipKeyframes(timeline, command, idFactory, remove = false) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  const current = (entry.clip.keyframes || []).map(frame => ({ ...frame, values: { ...(frame.values || {}) } }));
  let keyframes;
  let keyframeId = command.keyframeId;
  if (remove) {
    if (!current.some(frame => frame.id === keyframeId)) throw new Error('关键帧不存在');
    keyframes = current.filter(frame => frame.id !== keyframeId);
  } else {
    const offsetMs = Math.round(Number(command.offsetMs));
    if (!Number.isFinite(offsetMs) || offsetMs < 0 || offsetMs > Number(entry.clip.durationMs)) throw new Error('关键帧必须位于片段范围内');
    const allowed = ['x', 'y', 'scaleX', 'scaleY', 'rotation', 'opacity', 'volume'];
    const values = Object.fromEntries(allowed.filter(key => Number.isFinite(Number(command.values?.[key]))).map(key => [key, Number(command.values[key])]));
    if (!Object.keys(values).length) throw new Error('关键帧缺少可用参数');
    const ranges = {
      x: [-1, 1, '水平位置必须在 -1 到 1 之间'],
      y: [-1, 1, '垂直位置必须在 -1 到 1 之间'],
      scaleX: [.05, 20, '水平缩放必须在 0.05 到 20 之间'],
      scaleY: [.05, 20, '垂直缩放必须在 0.05 到 20 之间'],
      rotation: [-180, 180, '旋转角度必须在 -180 到 180 之间'],
      opacity: [0, 1, '透明度必须在 0 到 1 之间'],
      volume: [0, 2, '音量必须在 0 到 2 之间']
    };
    for (const [key, value] of Object.entries(values)) {
      const [minimum, maximum, message] = ranges[key];
      if (value < minimum || value > maximum) throw new Error(message);
    }
    const existing = current.find(frame => Number(frame.offsetMs) === offsetMs);
    if (existing) {
      existing.values = values;
      keyframeId = existing.id;
    } else {
      keyframeId = idFactory('keyframe');
      current.push({ id: keyframeId, offsetMs, values });
    }
    keyframes = current.sort((a, b) => Number(a.offsetMs) - Number(b.offsetMs));
  }
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => clip.id === entry.clipId ? { ...clip, keyframes } : clip)
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}`, metadata: { keyframeId } });
}

function applyTrackingKeyframes(timeline, command, idFactory) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.clip.kind !== 'video') throw new Error('区域跟踪只支持视频片段');
  if (!Array.isArray(command.frames) || !command.frames.length) throw new Error('没有可写入的跟踪结果');
  const durationMs = Number(entry.clip.durationMs) || 0;
  const normalized = command.frames.map((frame, index) => {
    const offsetMs = Math.round(Number(frame?.offsetMs));
    const x = Number(frame?.x);
    const y = Number(frame?.y);
    if (!Number.isFinite(offsetMs) || offsetMs < 0 || offsetMs > durationMs) throw new Error('跟踪关键帧必须位于片段范围内');
    if (!Number.isFinite(x) || x < -1 || x > 1) throw new Error('水平位置必须在 -1 到 1 之间');
    if (!Number.isFinite(y) || y < -1 || y > 1) throw new Error('垂直位置必须在 -1 到 1 之间');
    return { offsetMs, x, y, confidence: Math.max(0, Math.min(1, Number(frame?.confidence) || 0)), index };
  }).sort((a, b) => a.offsetMs - b.offsetMs || a.index - b.index);
  if (normalized.some((frame, index) => index > 0 && frame.offsetMs === normalized[index - 1].offsetMs)) throw new Error('跟踪结果存在重复时间位置');
  const manual = (entry.clip.keyframes || [])
    .filter(frame => !String(frame.id || '').startsWith('tracking-keyframe-'))
    .map(frame => ({ ...frame, values: { ...(frame.values || {}) } }));
  const manualPositionTimes = new Set(manual.filter(frame => Number.isFinite(Number(frame.values?.x)) || Number.isFinite(Number(frame.values?.y))).map(frame => Number(frame.offsetMs)));
  if (normalized.some(frame => manualPositionTimes.has(frame.offsetMs))) throw new Error('跟踪结果与人工位置关键帧冲突，请先移动或删除冲突关键帧');
  const generated = normalized.map(frame => ({
    id: idFactory('tracking-keyframe'),
    offsetMs: frame.offsetMs,
    values: { x: frame.x, y: frame.y },
    easing: { type: 'linear' }
  }));
  const keyframes = [...manual, ...generated].sort((a, b) => Number(a.offsetMs) - Number(b.offsetMs));
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => clip.id === entry.clipId ? { ...clip, keyframes } : clip)
  }));
  return commandResult({ ...timeline, tracks }, {
    selectionKey: `${entry.trackId}::${entry.clipId}`,
    metadata: { trackingKeyframeCount: generated.length, keyframeIds: generated.map(frame => frame.id) }
  });
}

function updateClipKeyframeEasing(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  const keyframes = (entry.clip.keyframes || []).map(frame => ({ ...frame, values: { ...(frame.values || {}) } }));
  const index = keyframes.findIndex(frame => frame.id === command.keyframeId);
  if (index < 0) throw new Error('关键帧不存在');
  keyframes[index] = { ...keyframes[index], easing: normalizeKeyframeEasing(command.easing) };
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({ ...track, clips: track.clips.map(clip => clip.id === entry.clipId ? { ...clip, keyframes } : clip) }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}`, metadata: { keyframeId: command.keyframeId } });
}

function batchEditClipKeyframes(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  const requestedIDs = [...new Set((command.keyframeIds || []).map(value => String(value || '').trim()).filter(Boolean))];
  if (!requestedIDs.length) throw new Error('请至少选择一个关键帧');
  const selectedIDs = new Set(requestedIDs);
  const current = (entry.clip.keyframes || []).map(frame => ({
    ...frame,
    values: { ...(frame.values || {}) },
    easing: frame.easing ? { ...frame.easing } : frame.easing
  }));
  if (requestedIDs.some(id => !current.some(frame => frame.id === id))) throw new Error('关键帧不存在');
  let keyframes;
  switch (command.action) {
    case 'easing': {
      const easing = normalizeKeyframeEasing(command.easing);
      keyframes = current.map(frame => selectedIDs.has(frame.id) ? { ...frame, easing: { ...easing } } : frame);
      break;
    }
    case 'shift': {
      const deltaMs = Math.round(Number(command.deltaMs));
      if (!Number.isFinite(deltaMs)) throw new Error('时间偏移必须是有效毫秒数');
      keyframes = current.map(frame => selectedIDs.has(frame.id) ? { ...frame, offsetMs: Number(frame.offsetMs) + deltaMs } : frame);
      if (keyframes.some(frame => frame.offsetMs < 0 || frame.offsetMs > Number(entry.clip.durationMs))) throw new Error('偏移后的关键帧必须位于片段范围内');
      const occupied = new Set();
      for (const frame of keyframes) {
        const offset = Number(frame.offsetMs);
        if (occupied.has(offset)) throw new Error('关键帧时间位置冲突');
        occupied.add(offset);
      }
      keyframes.sort((left, right) => Number(left.offsetMs) - Number(right.offsetMs));
      break;
    }
    case 'remove':
      keyframes = current.filter(frame => !selectedIDs.has(frame.id));
      break;
    default:
      throw new Error('批量编辑操作无效');
  }
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => clip.id === entry.clipId ? { ...clip, keyframes } : clip)
  }));
  return commandResult({ ...timeline, tracks }, {
    selectionKey: `${entry.trackId}::${entry.clipId}`,
    metadata: { keyframeIds: requestedIDs }
  });
}

function pasteClipKeyframes(timeline, command, idFactory) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  const sourceFrames = (command.frames || []).map(frame => ({
    offsetMs: Number(frame?.offsetMs),
    values: { ...(frame?.values || {}) },
    easing: frame?.easing ? normalizeKeyframeEasing(frame.easing) : null
  }));
  if (!sourceFrames.length) throw new Error('关键帧剪贴板为空');
  if (sourceFrames.some(frame => !Number.isFinite(frame.offsetMs) || frame.offsetMs < 0)) throw new Error('关键帧剪贴板包含无效时间');
  const anchorMs = Math.round(Number(command.anchorMs));
  if (!Number.isFinite(anchorMs) || anchorMs < 0 || anchorMs > Number(entry.clip.durationMs)) throw new Error('关键帧粘贴位置必须位于片段范围内');
  const sourceStart = Math.min(...sourceFrames.map(frame => frame.offsetMs));
  const targetOffsets = sourceFrames.map(frame => anchorMs + Math.round(frame.offsetMs - sourceStart));
  if (targetOffsets.some(offset => offset < 0 || offset > Number(entry.clip.durationMs))) throw new Error('粘贴后的关键帧必须位于片段范围内');
  const occupied = new Set((entry.clip.keyframes || []).map(frame => Number(frame.offsetMs)));
  for (const offset of targetOffsets) {
    if (occupied.has(offset)) throw new Error('关键帧时间位置冲突');
    occupied.add(offset);
  }
  let working = timeline;
  const keyframeIds = [];
  sourceFrames.forEach((frame, index) => {
    const inserted = updateClipKeyframes(working, {
      trackId: command.trackId,
      clipId: command.clipId,
      offsetMs: targetOffsets[index],
      values: frame.values
    }, idFactory, false);
    working = inserted.timeline;
    const keyframeId = inserted.metadata.keyframeId;
    if (frame.easing) working = updateClipKeyframeEasing(working, {
      trackId: command.trackId,
      clipId: command.clipId,
      keyframeId,
      easing: frame.easing
    }).timeline;
    keyframeIds.push(keyframeId);
  });
  return commandResult(working, {
    selectionKey: `${entry.trackId}::${entry.clipId}`,
    metadata: { keyframeIds, anchorMs }
  });
}

function updateVolumeEnvelopePoint(timeline, command, idFactory, mode) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'audio' || entry.clip.kind !== 'audio') throw new Error('音量包络只支持音频片段');
  const keyframes = (entry.clip.keyframes || []).map(frame => ({ ...frame, values: { ...(frame.values || {}) } }));
  if (mode === 'remove') {
    const index = keyframes.findIndex(frame => frame.id === command.keyframeId && Number.isFinite(Number(frame.values?.volume)));
    if (index < 0) throw new Error('音量点不存在');
    const values = { ...keyframes[index].values };
    delete values.volume;
    if (Object.keys(values).length) keyframes[index] = { ...keyframes[index], values };
    else keyframes.splice(index, 1);
  } else {
    const offsetMs = Math.round(Number(command.offsetMs));
    const volume = Number(command.volume);
    if (!Number.isFinite(offsetMs) || offsetMs < 0 || offsetMs > Number(entry.clip.durationMs)) throw new Error('音量点必须位于片段范围内');
    if (!Number.isFinite(volume) || volume < 0 || volume > 2) throw new Error('音量必须在 0 到 2 之间');
    let index = mode === 'move' ? keyframes.findIndex(frame => frame.id === command.keyframeId) : keyframes.findIndex(frame => Number(frame.offsetMs) === offsetMs);
    if (mode === 'move' && index < 0) throw new Error('音量点不存在');
    if (keyframes.some((frame, frameIndex) => frameIndex !== index && Number(frame.offsetMs) === offsetMs && Number.isFinite(Number(frame.values?.volume)))) throw new Error('该时间已有音量点');
    if (index >= 0) keyframes[index] = { ...keyframes[index], offsetMs, values: { ...keyframes[index].values, volume } };
    else keyframes.push({ id: idFactory('volume-point'), offsetMs, values: { volume } });
    keyframes.sort((left, right) => Number(left.offsetMs) - Number(right.offsetMs));
  }
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({ ...track, clips: track.clips.map(clip => clip.id !== entry.clipId ? clip : ({ ...clip, keyframes })) }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

function normalizedAudioProcessing(value = {}) {
  const denoise = Number(value.denoise ?? 0);
  const eqLowDB = Number(value.eqLowDB ?? 0);
  const eqMidDB = Number(value.eqMidDB ?? 0);
  const eqHighDB = Number(value.eqHighDB ?? 0);
  const channelMode = String(value.channelMode || 'stereo');
  const loudnessEnabled = Boolean(value.loudnessEnabled);
  const targetLUFS = Number(value.targetLUFS ?? -16);
  const voiceEffect = String(value.voiceEffect || 'natural');
  const pitchSemitones = Number(value.pitchSemitones ?? 0);
  if (!Number.isFinite(denoise) || denoise < 0 || denoise > 1) throw new Error('降噪强度必须在 0 到 1 之间');
  if (![eqLowDB, eqMidDB, eqHighDB].every(item => Number.isFinite(item) && item >= -12 && item <= 12)) throw new Error('均衡器增益必须在 -12 到 12 dB 之间');
  if (!['stereo', 'mono', 'left', 'right', 'swap'].includes(channelMode)) throw new Error('不支持的声道模式');
  if (!Number.isFinite(targetLUFS) || targetLUFS < -24 || targetLUFS > -8) throw new Error('目标响度必须在 -24 到 -8 LUFS 之间');
  if (!audioEffectValues.has(voiceEffect)) throw new Error('不支持的音频效果');
  if (!Number.isFinite(pitchSemitones) || pitchSemitones < -12 || pitchSemitones > 12) throw new Error('音调必须在 -12 到 12 个半音之间');
  return { denoise, eqLowDB, eqMidDB, eqHighDB, channelMode, loudnessEnabled, targetLUFS, voiceEffect, pitchSemitones };
}

function setAudioProcessing(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'audio' || entry.clip.kind !== 'audio') throw new Error('音频处理只支持音频片段');
  if (entry.track.locked) throw new Error('轨道已锁定');
  const audioProcessing = normalizedAudioProcessing(command.processing);
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => clip.id !== entry.clipId ? clip : ({ ...clip, audioProcessing }))
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

function setVisualProcessing(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (!['video', 'image'].includes(entry.track.kind) || !['video', 'image'].includes(entry.clip.kind)) throw new Error('高级画面只支持画面片段');
  if (entry.track.locked) throw new Error('轨道已锁定');
  const processing = command.processing || {};
  const frameInterpolation = String(processing.frameInterpolation || 'off');
  if (!['off', 'blend', 'optical-flow'].includes(frameInterpolation)) throw new Error('不支持的补帧模式');
  if (entry.clip.kind !== 'video' && frameInterpolation !== 'off') throw new Error('视频补帧只支持视频片段');
  const stabilizationSource = processing.stabilization || {};
  const stabilization = {
    enabled: Boolean(stabilizationSource.enabled),
    smoothness: Number(stabilizationSource.smoothness ?? 15),
    zoom: Number(stabilizationSource.zoom ?? 0),
    tripod: Boolean(stabilizationSource.tripod)
  };
  if (!Number.isFinite(stabilization.smoothness) || stabilization.smoothness < 1 || stabilization.smoothness > 100) throw new Error('防抖平滑度必须在 1 到 100 之间');
  if (!Number.isFinite(stabilization.zoom) || stabilization.zoom < 0 || stabilization.zoom > 30) throw new Error('防抖边缘缩放必须在 0% 到 30% 之间');
  stabilization.smoothness = Math.round(stabilization.smoothness);
  if (entry.clip.kind !== 'video' && stabilization.enabled) throw new Error('视频防抖只支持视频片段');
  const personSegmentationSource = processing.personSegmentation || {};
  const personSegmentation = normalizePersonSegmentation(personSegmentationSource);
  if (personSegmentation.enabled && entry.clip.kind !== 'video') throw new Error('AI人物抠像只支持视频片段');
  if (personSegmentation.enabled && !personSegmentation.matteAssetId) throw new Error('启用AI人物抠像前必须生成人物蒙版素材');
  if (personSegmentation.enabled && !personSegmentation.signature) throw new Error('人物蒙版缺少片段签名，请重新生成');
  if (personSegmentationSource.threshold !== undefined && (!Number.isFinite(Number(personSegmentationSource.threshold)) || Number(personSegmentationSource.threshold) < 0 || Number(personSegmentationSource.threshold) > 1)) throw new Error('人物阈值参数无效');
  if (personSegmentationSource.feather !== undefined && (!Number.isFinite(Number(personSegmentationSource.feather)) || Number(personSegmentationSource.feather) < 0 || Number(personSegmentationSource.feather) > .5)) throw new Error('人物边缘柔化参数无效');
  const color = processing.color || {};
  const numeric = (value, minimum, maximum, label, fallback) => {
    const number = value === undefined ? fallback : Number(value);
    if (!Number.isFinite(number) || number < minimum || number > maximum) throw new Error(`${label}参数无效`);
    return number;
  };
  const defaultCurve = [0, .25, .5, .75, 1];
  if (color.curve !== undefined && (!Array.isArray(color.curve) || color.curve.length !== 5)) throw new Error('调色曲线必须包含5个控制点');
  const legacyCurve = (color.curve || defaultCurve).map((value, index) => numeric(value, 0, 1, `调色曲线第${index + 1}点`, defaultCurve[index]));
  let curvePoints = legacyCurve.map((y, index) => ({ x: index / 4, y }));
  if (color.curvePoints !== undefined) {
    if (!Array.isArray(color.curvePoints) || color.curvePoints.length < 2 || color.curvePoints.length > 16) throw new Error('自由调色曲线必须包含2到16个控制点');
    curvePoints = color.curvePoints.map((point, index) => ({
      x: numeric(point?.x, 0, 1, `调色曲线第${index + 1}点横坐标`, 0),
      y: numeric(point?.y, 0, 1, `调色曲线第${index + 1}点纵坐标`, 0)
    }));
    if (curvePoints[0].x !== 0 || curvePoints.at(-1).x !== 1) throw new Error('调色曲线首尾控制点必须位于0和1');
    if (curvePoints.some((point, index) => index > 0 && point.x <= curvePoints[index - 1].x)) throw new Error('调色曲线横坐标必须严格递增');
  }
  const curve = defaultCurve.map(x => sampleColorCurve(curvePoints, x));
  const geometrySource = processing.geometry || {};
  const geometry = {
    cropLeft: numeric(geometrySource.cropLeft, 0, .9, '左裁剪', 0),
    cropTop: numeric(geometrySource.cropTop, 0, .9, '上裁剪', 0),
    cropRight: numeric(geometrySource.cropRight, 0, .9, '右裁剪', 0),
    cropBottom: numeric(geometrySource.cropBottom, 0, .9, '下裁剪', 0),
    flipHorizontal: Boolean(geometrySource.flipHorizontal),
    flipVertical: Boolean(geometrySource.flipVertical)
  };
  const detailsSource = processing.details || {};
  const details = {
    sharpen: numeric(detailsSource.sharpen, 0, 1, '锐化', 0),
    vignette: numeric(detailsSource.vignette, 0, 1, '暗角', 0),
    grain: numeric(detailsSource.grain, 0, 1, '颗粒', 0)
  };
  const objectRemovalSource = processing.objectRemoval || {};
  const objectRemoval = {
    enabled: Boolean(objectRemovalSource.enabled),
    x: numeric(objectRemovalSource.x, 0, 1, '消除区域中心X', .5),
    y: numeric(objectRemovalSource.y, 0, 1, '消除区域中心Y', .5),
    width: numeric(objectRemovalSource.width, .05, 1, '消除区域宽度', .2),
    height: numeric(objectRemovalSource.height, .05, 1, '消除区域高度', .2)
  };
  if (objectRemoval.x - objectRemoval.width / 2 < 0 || objectRemoval.x + objectRemoval.width / 2 > 1 || objectRemoval.y - objectRemoval.height / 2 < 0 || objectRemoval.y + objectRemoval.height / 2 > 1) {
    throw new Error('消除区域必须位于画面内');
  }
  if (geometry.cropLeft + geometry.cropRight >= .95) throw new Error('水平裁剪后必须保留至少5%画面');
  if (geometry.cropTop + geometry.cropBottom >= .95) throw new Error('垂直裁剪后必须保留至少5%画面');
  const maskSource = processing.mask || {};
  let maskPoints = [];
  let bezierPoints = [];
  if (String(maskSource.type || 'none') === 'polygon') {
    if (!Array.isArray(maskSource.points) || maskSource.points.length < 3 || maskSource.points.length > 12) throw new Error('手绘多边形蒙版必须包含3到12个控制点');
    maskPoints = maskSource.points.map((point, index) => ({
      x: numeric(point?.x, 0, 1, `蒙版第${index + 1}控制点横坐标`, 0),
      y: numeric(point?.y, 0, 1, `蒙版第${index + 1}控制点纵坐标`, 0)
    }));
    if (polygonArea(maskPoints) < .0001) throw new Error('手绘多边形蒙版必须具有有效面积');
  }
  if (String(maskSource.type || 'none') === 'bezier') {
    if (!Array.isArray(maskSource.bezierPoints) || maskSource.bezierPoints.length < 3 || maskSource.bezierPoints.length > 8) throw new Error('贝塞尔曲线蒙版必须包含3到8个锚点');
    bezierPoints = maskSource.bezierPoints.map((point, index) => ({
      x: numeric(point?.x, 0, 1, `贝塞尔第${index + 1}锚点横坐标`, 0),
      y: numeric(point?.y, 0, 1, `贝塞尔第${index + 1}锚点纵坐标`, 0),
      inX: numeric(point?.inX, 0, 1, `贝塞尔第${index + 1}入控制柄横坐标`, 0),
      inY: numeric(point?.inY, 0, 1, `贝塞尔第${index + 1}入控制柄纵坐标`, 0),
      outX: numeric(point?.outX, 0, 1, `贝塞尔第${index + 1}出控制柄横坐标`, 0),
      outY: numeric(point?.outY, 0, 1, `贝塞尔第${index + 1}出控制柄纵坐标`, 0)
    }));
    maskPoints = sampleBezierMaskPoints(bezierPoints);
    if (polygonArea(maskPoints) < .0001) throw new Error('贝塞尔曲线蒙版必须具有有效面积');
  }
  const visualProcessing = {
    frameInterpolation,
    stabilization,
    personSegmentation,
    color: {
      exposure: numeric(color.exposure, -2, 2, '曝光', 0),
      temperature: numeric(color.temperature, -1, 1, '色温', 0),
      tint: numeric(color.tint, -1, 1, '色调', 0),
      highlights: numeric(color.highlights, -1, 1, '高光', 0),
      shadows: numeric(color.shadows, -1, 1, '阴影', 0),
      gamma: numeric(color.gamma, .5, 2, '伽马', 1),
      curve,
      curvePoints
    },
    chromaKey: {
      enabled: Boolean(processing.chromaKey?.enabled),
      removeBlack: Boolean(processing.chromaKey?.removeBlack),
      color: String(processing.chromaKey?.color || '#00ff00').toLowerCase(),
      similarity: numeric(processing.chromaKey?.similarity, .01, 1, '色键相似度', .2),
      blend: numeric(processing.chromaKey?.blend, 0, 1, '色键柔化', .08)
    },
    mask: {
      type: String(maskSource.type || 'none'),
      x: numeric(maskSource.x, 0, 1, '蒙版位置', .5),
      y: numeric(maskSource.y, 0, 1, '蒙版位置', .5),
      width: numeric(maskSource.width, .05, 1, '蒙版宽度', 1),
      height: numeric(maskSource.height, .05, 1, '蒙版高度', 1),
      feather: numeric(maskSource.feather, 0, .5, '蒙版羽化', 0),
      invert: Boolean(maskSource.invert),
      points: maskPoints,
      bezierPoints
    },
    geometry,
    details,
    objectRemoval
  };
  if (!/^#[0-9a-f]{6}$/i.test(visualProcessing.chromaKey.color)) throw new Error('色键颜色必须是十六进制颜色');
  if (!['none', 'rectangle', 'ellipse', 'polygon', 'bezier'].includes(visualProcessing.mask.type)) throw new Error('不支持的蒙版类型');
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({ ...track, clips: track.clips.map(clip => clip.id !== entry.clipId ? clip : ({ ...clip, visualProcessing })) }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

function batchPatchVisualProcessing(timeline, command) {
  const clipKeys = [...new Set((command.clipKeys || []).filter(Boolean))];
  if (!clipKeys.length) throw new Error('批量画面设置至少选择一个片段');
  const entries = clipKeys.map(key => findClip(timeline, key));
  if (entries.some(entry => !['video', 'image'].includes(entry.track.kind) || !['video', 'image'].includes(entry.clip.kind))) throw new Error('高级画面只支持画面片段');
  const patch = command.patch || {};
  let next = timeline;
  entries.forEach(entry => {
    const current = normalizeVisualProcessing(entry.clip.visualProcessing);
    const processing = {
      ...current,
      ...patch,
      ...Object.fromEntries(Object.entries(current)
        .filter(([, value]) => value && typeof value === 'object' && !Array.isArray(value))
        .map(([section, value]) => [section, { ...value, ...(patch[section] || {}) }]))
    };
    next = setVisualProcessing(next, { trackId: entry.trackId, clipId: entry.clipId, processing }).timeline;
  });
  return commandResult(next, { selectionKey: clipKeys[0], metadata: { clipCount: clipKeys.length } });
}

function clipSourceTimeAtTimelineTime(clip, timeMs) {
  const startMs = Number(clip.startMs) || 0;
  const durationMs = Math.max(1, Number(clip.durationMs) || 1);
  const requested = Number(timeMs);
  if (!Number.isFinite(requested) || requested < startMs || requested >= startMs + durationMs) throw new Error('播放头必须位于片段内');
  const outputOffsetMs = requested - startMs;
  const hasCurve = Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2;
  const rate = Number(clip.playbackRate) > 0 ? Number(clip.playbackRate) : 1;
  const consumedMs = hasCurve ? speedCurveSourceOffsetMs(outputOffsetMs, durationMs, clip.speedCurve) : outputOffsetMs * rate;
  const totalSourceMs = hasCurve ? speedCurveSourceOffsetMs(durationMs, durationMs, clip.speedCurve) : durationMs * rate;
  const sourceStartMs = Math.max(0, Number(clip.sourceStartMs) || 0);
  return Math.max(0, Math.round(clip.reversed ? sourceStartMs + totalSourceMs - consumedMs - 1 : sourceStartMs + consumedMs));
}

function setClipReverse(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'video' || entry.clip.kind !== 'video') throw new Error('倒放只支持视频片段');
  const reversed = Boolean(command.reversed);
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => {
      if (clip.id !== entry.clipId) return clip;
      const next = { ...clip };
      delete next.freezeSourceMs;
      if (reversed) next.reversed = true;
      else delete next.reversed;
      return next;
    })
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

function setClipFreeze(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'video' || entry.clip.kind !== 'video') throw new Error('定格只支持视频片段');
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => {
      if (clip.id !== entry.clipId) return clip;
      const next = { ...clip };
      if (command.clear) delete next.freezeSourceMs;
      else {
        next.freezeSourceMs = clipSourceTimeAtTimelineTime(clip, command.timeMs);
        delete next.reversed;
      }
      return next;
    })
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

const supportedBlendModes = new Set(['normal', 'multiply', 'screen', 'overlay', 'soft-light', 'difference', 'add']);

function setClipBlendMode(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'video' || !['video', 'image'].includes(entry.clip.kind)) throw new Error('混合模式只支持画面片段');
  const blendMode = String(command.blendMode || 'normal');
  if (!supportedBlendModes.has(blendMode)) throw new Error(`不支持的混合模式：${blendMode}`);
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => {
      if (clip.id !== entry.clipId) return clip;
      const next = { ...clip };
      if (blendMode === 'normal') delete next.blendMode;
      else next.blendMode = blendMode;
      return next;
    })
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

function setAudioFades(timeline, command) {
  const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
  if (entry.track.kind !== 'audio' || entry.clip.kind !== 'audio') throw new Error('淡入淡出只支持音频片段');
  const durationMs = Math.max(1, Number(entry.clip.durationMs) || 1);
  const parseDuration = (value, label) => {
    const number = Number(value ?? 0);
    if (!Number.isFinite(number) || number < 0 || number > durationMs) throw new Error(`${label}时长必须在 0 到片段时长之间`);
    return Math.round(number);
  };
  const fadeInMs = parseDuration(command.fadeInMs, '淡入');
  const fadeOutMs = parseDuration(command.fadeOutMs, '淡出');
  const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
    ...track,
    clips: track.clips.map(clip => clip.id !== entry.clipId ? clip : { ...clip, fadeInMs, fadeOutMs })
  }));
  return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
}

export function findTimelineEditPoint(timeline, currentMs, direction = 'next') {
  const current = Math.max(0, Number(currentMs) || 0);
  const points = [...new Set((timeline?.tracks || []).flatMap(track => (track.clips || []).flatMap(clip => {
    const start = Math.max(0, Number(clip.startMs) || 0);
    return [start, start + Math.max(0, Number(clip.durationMs) || 0)];
  })))].sort((a, b) => a - b);
  if (!points.length) return 0;
  if (direction === 'previous') return [...points].reverse().find(point => point < current) ?? points[0];
  return points.find(point => point > current) ?? points.at(-1);
}

export function executeAICutCommand(timeline, command, dependencies = {}) {
  if (!command?.type) throw new Error('缺少剪辑命令');
  const idFactory = dependencies.idFactory || (() => {
    throw new Error('该命令缺少片段编号生成器');
  });
  switch (command.type) {
    case 'edit.cancel':
      return commandResult(timeline, { changed: false });
    case 'clip.duplicate': {
      const newId = command.newId || idFactory('clip');
      const next = duplicateTimelineClip(timeline, command.trackId, command.clipId, newId);
      return commandResult(next, { selectionKey: `${command.trackId}::${newId}` });
    }
    case 'clip.delete': {
      const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
      if (entry.track.kind === 'transition' || entry.clip.kind === 'transition') {
        const result = removeTransitionClip(timeline, command.trackId, command.clipId);
        if (!result.ok) throw new Error(result.reasonZh || '转场删除失败');
        return commandResult(result.timeline);
      }
      return commandResult(removeTimelineClip(timeline, command.trackId, command.clipId));
    }
    case 'clip.delete_many':
      return deleteMany(timeline, command.clipKeys, command.ripple);
    case 'clip.merge_next':
      return commandResult(mergeTimelineClipWithNext(timeline, command.trackId, command.clipId), {
        selectionKey: `${command.trackId}::${command.clipId}`
      });
    case 'clip.split':
      return commandResult(splitTimelineClip(timeline, command.trackId, command.clipId, idFactory));
    case 'clip.split_at':
      return commandResult(splitTimelineClipAt(timeline, command.trackId, command.clipId, command.offsetMs, idFactory));
    case 'clip.split_at_playhead': {
      const result = splitTimelineAtTime(timeline, command.timeMs, idFactory);
      return commandResult(result.timeline, { metadata: { splitCount: result.splitCount } });
    }
    case 'clip.move': {
      const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
      if (entry.track.kind === 'transition' || entry.clip.kind === 'transition') {
        const result = moveTransitionClip(timeline, command.trackId, command.clipId, command.startMs);
        if (!result.ok) throw new Error(result.reasonZh || '转场移动失败');
        return commandResult(result.timeline, { selectionKey: `${command.trackId}::${command.clipId}` });
      }
      return commandResult(moveTimelineClip(timeline, command.trackId, command.clipId, command.startMs, {
        magnetic: command.magnetic,
        linked: command.linked,
        thresholdMs: command.thresholdMs,
        playheadMs: command.playheadMs
      }));
    }
    case 'clip.resize': {
      const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
      if (entry.track.kind === 'transition' || entry.clip.kind === 'transition') {
        const startMs = Number(entry.clip.startMs) || 0;
        const endMs = startMs + (Number(entry.clip.durationMs) || 0);
        const durationMs = command.edge === 'left' ? endMs - Number(command.edgeTimeMs) : Number(command.edgeTimeMs) - startMs;
        const result = resizeTransitionClip(timeline, command.trackId, command.clipId, durationMs);
        if (!result.ok) throw new Error(result.reasonZh || '转场裁剪失败');
        return commandResult(result.timeline, { selectionKey: `${command.trackId}::${command.clipId}` });
      }
      return commandResult(command.ripple
        ? rippleResize(timeline, command.trackId, command.clipId, command.edge, command.edgeTimeMs)
        : resizeTimelineClipEdge(timeline, command.trackId, command.clipId, command.edge, command.edgeTimeMs));
    }
    case 'clip.nudge': {
      const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
      const fps = Number(command.fps) > 0 ? Number(command.fps) : 30;
      const deltaMs = Math.round((Number(command.frames) || 0) * 1000 / fps);
      const startMs = (Number(entry.clip.startMs) || 0) + deltaMs;
      if (startMs < 0) throw new Error('片段不能早于时间线零点');
      return commandResult(moveTimelineClip(timeline, command.trackId, command.clipId, startMs, { magnetic: false, linked: command.linked }));
    }
    case 'clip.paste':
      return pasteClipboard(timeline, command.clipboard, command.atMs, idFactory);
    case 'clip.transition_set': {
      const result = addTransitionClip(timeline, command.trackId, command.clipId, { id: command.transition, name: command.transitionName || command.transition }, command.transitionMs, idFactory);
      if (!result.ok) throw new Error(result.reasonZh || '转场设置失败');
      const transitionTrack = result.timeline.tracks.find(track => track.kind === 'transition' && track.clips?.some(clip => clip.transitionTrackId === command.trackId && clip.transitionRightClipId === command.clipId));
      const transitionClip = transitionTrack?.clips?.find(clip => clip.transitionTrackId === command.trackId && clip.transitionRightClipId === command.clipId);
      return commandResult(result.timeline, { selectionKey: transitionTrack && transitionClip ? `${transitionTrack.id}::${transitionClip.id}` : `${command.trackId}::${command.clipId}` });
    }
    case 'clip.playback_rate_set':
      return commandResult(setClipPlaybackRate(timeline, command.trackId, command.clipId, command.playbackRate), { selectionKey: `${command.trackId}::${command.clipId}` });
    case 'clip.speed_curve_set': {
      const entry = findClip(timeline, `${command.trackId}::${command.clipId}`);
      if (entry.track.kind !== 'video' || entry.clip.kind !== 'video') throw new Error('速度曲线只支持视频片段');
      const speedCurve = normalizeSpeedCurve(command.points);
      const previousRate = Number(entry.clip.playbackRate) > 0 ? Number(entry.clip.playbackRate) : 1;
      const sourceSpanMs = Math.max(1, Math.round((Number(entry.clip.durationMs) || 1) * previousRate));
      const durationMs = speedCurveDurationMs(sourceSpanMs, speedCurve);
      const scale = durationMs / Math.max(1, Number(entry.clip.durationMs) || 1);
      const playbackRate = speedCurveAverageRate(speedCurve);
      const tracks = timeline.tracks.map(track => track.id !== entry.trackId ? track : ({
        ...track,
        clips: track.clips.map(clip => clip.id !== entry.clipId ? clip : ({
          ...clip,
          durationMs,
          playbackRate,
          speedCurve,
          keyframes: (clip.keyframes || []).map(frame => ({ ...frame, offsetMs: Math.max(0, Math.round(Number(frame.offsetMs) * scale)), values: { ...(frame.values || {}) } }))
        }))
      }));
      return commandResult({ ...timeline, tracks }, { selectionKey: `${entry.trackId}::${entry.clipId}` });
    }
    case 'clip.volume_point_upsert':
      return updateVolumeEnvelopePoint(timeline, command, idFactory, 'upsert');
    case 'clip.volume_point_move':
      return updateVolumeEnvelopePoint(timeline, command, idFactory, 'move');
    case 'clip.volume_point_remove':
      return updateVolumeEnvelopePoint(timeline, command, idFactory, 'remove');
    case 'clip.audio_processing_set':
      return setAudioProcessing(timeline, command);
    case 'clip.audio_fade_set':
      return setAudioFades(timeline, command);
    case 'clip.visual_processing_set':
      return setVisualProcessing(timeline, command);
    case 'clip.visual_processing_batch_patch':
      return batchPatchVisualProcessing(timeline, command);
	case 'clip.reverse_set':
	  return setClipReverse(timeline, command);
	case 'clip.freeze_set':
	  return setClipFreeze(timeline, command);
	case 'clip.blend_mode_set':
	  return setClipBlendMode(timeline, command);
    case 'timeline.marker_add':
      return addTimelineMarker(timeline, command, idFactory);
    case 'timeline.marker_remove':
      return removeTimelineMarker(timeline, command.markerId);
    case 'timeline.bgm_plan':
      return commandResult(planTimelineBGM(timeline, command.asset, command.options, idFactory));
    case 'clip.keyframe_upsert':
      return updateClipKeyframes(timeline, command, idFactory, false);
    case 'clip.keyframe_remove':
      return updateClipKeyframes(timeline, command, idFactory, true);
    case 'clip.keyframe_easing_set':
      return updateClipKeyframeEasing(timeline, command);
    case 'clip.keyframes_batch_edit':
      return batchEditClipKeyframes(timeline, command);
    case 'clip.keyframes_paste':
      return pasteClipKeyframes(timeline, command, idFactory);
    case 'clip.tracking_keyframes_apply':
      return applyTrackingKeyframes(timeline, command, idFactory);
    case 'clip.link':
      return setLinkGroup(timeline, command.clipKeys, String(command.groupId || ((command.clipKeys || []).length >= 2 ? idFactory('link') : '')));
    case 'clip.unlink':
      return clearLinkGroup(timeline, command.clipKeys);
    case 'clip.compound_create':
      return createCompound(timeline, command.clipKeys, String(command.compoundId || idFactory('compound')), command.name);
    case 'clip.compound_move':
      return moveCompound(timeline, command.compoundId, command.startMs);
    case 'clip.compound_trim':
      return trimCompound(timeline, command.compoundId, command.edge, command.edgeTimeMs);
    case 'clip.compound_split':
      return splitCompound(timeline, command.compoundId, command.timeMs, command.rightCompoundId, idFactory);
    case 'clip.compound_dissolve':
      return dissolveCompound(timeline, command.clipKeys, command.compoundId);
    case 'compound.track_add':
      return addCompoundTrack(timeline, command);
    case 'compound.track_remove':
      return removeCompoundTrack(timeline, command);
    case 'compound.track_reorder':
      return reorderCompoundTracks(timeline, command);
    case 'compound.member_track_move':
      return compoundMemberTrackMove(timeline, command);
    case 'compound.member_transfer':
      return compoundMemberTransfer(timeline, command);
    case 'compound.block_transfer':
      return compoundBlockTransfer(timeline, command);
    case 'compound.members_transfer':
      return compoundMembersTransfer(timeline, command);
    default:
      throw new Error(`不支持的剪辑命令：${command.type}`);
  }
}

export function reconcileTimelineSave(currentTimeline, savedTimeline, capturedEditSequence, currentEditSequence) {
  if (capturedEditSequence === currentEditSequence) return { timeline: savedTimeline, dirty: false };
  return {
    timeline: { ...currentTimeline, revision: savedTimeline.revision },
    dirty: true
  };
}

export function restoreTimelineSnapshot(snapshot, currentRevision) {
  return { ...snapshot, revision: currentRevision };
}
