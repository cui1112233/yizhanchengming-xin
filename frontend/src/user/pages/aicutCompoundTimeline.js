import { clipHasCompound, compoundEntries, compoundPathForClip, outerCompoundRef, trackBelongsToCompound } from './aicutCompoundModel.js';

function finiteNonNegative(value, message) {
  const number = Number(value);
  if (!Number.isFinite(number) || number < 0) throw new Error(message);
  return Math.round(number);
}

export function collapseCompoundTimelineView(timeline) {
  const groups = new Map();
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      const outer = outerCompoundRef(clip);
      if (!outer) continue;
      const compoundId = outer.id;
      const entry = { track, clip, key: `${track.id}::${clip.id}` };
      if (!groups.has(compoundId)) groups.set(compoundId, { ref: outer, entries: [] });
      groups.get(compoundId).entries.push(entry);
    }
  }

  const containers = new Map();
  for (const [compoundId, group] of groups) {
    const { entries, ref } = group;
    if (entries.length < 2) continue;
    const startMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
    const endMs = Math.max(...entries.map(({ clip }) => (Number(clip.startMs) || 0) + Math.max(0, Number(clip.durationMs) || 0)));
    const representative = entries.reduce((best, entry) => (Number(entry.clip.startMs) || 0) < (Number(best.clip.startMs) || 0) ? entry : best);
    containers.set(compoundId, {
      representativeKey: representative.key,
      clip: {
        ...representative.clip,
        id: `compound-container:${compoundId}`,
        kind: 'compound',
        compoundId,
        name: String(ref.name || '复合片段'),
        compoundName: String(ref.name || '复合片段'),
        startMs,
        durationMs: Math.max(1, endMs - startMs),
        memberCount: entries.length,
        memberKeys: entries.map(entry => entry.key),
        isCompoundContainer: true
      }
    });
  }

  return {
    ...timeline,
    tracks: (timeline?.tracks || []).map(track => ({
      ...track,
      clips: (track.clips || []).flatMap(clip => {
        const compoundId = outerCompoundRef(clip)?.id || '';
        const container = containers.get(compoundId);
        if (!container) return [clip];
        return container.representativeKey === `${track.id}::${clip.id}` ? [container.clip] : [];
      })
    }))
  };
}

export function buildCompoundTimelineView(timeline, compoundId) {
  const id = String(compoundId || '').trim();
  const entries = compoundEntries(timeline, id);
  if (!id || entries.length < 2) throw new Error('复合片段不存在或成员不足');
  const originMs = Math.min(...entries.map(({ clip }) => Number(clip.startMs) || 0));
  const endMs = Math.max(...entries.map(({ clip }) => (Number(clip.startMs) || 0) + Math.max(0, Number(clip.durationMs) || 0)));
  const memberKeys = entries.map(entry => entry.key);
  const memberSet = new Set(memberKeys);
  const firstPath = compoundPathForClip(entries[0].clip);
  const firstIndex = firstPath.findIndex(item => item.id === id);
  const parentCompoundId = firstIndex >= 0 ? firstPath[firstIndex + 1]?.id || '' : '';
  const targetRef = firstIndex >= 0 ? firstPath[firstIndex] : { id, name: entries[0].clip.compoundName || '复合片段' };
  const childGroups = new Map();
  for (const entry of entries) {
    const path = compoundPathForClip(entry.clip);
    const index = path.findIndex(item => item.id === id);
    const child = index > 0 ? path[index - 1] : null;
    if (!child) continue;
    if (!childGroups.has(child.id)) childGroups.set(child.id, { ref: child, entries: [] });
    childGroups.get(child.id).entries.push(entry);
  }
  const childContainers = new Map();
  for (const [childId, group] of childGroups) {
    const childStart = Math.min(...group.entries.map(({ clip }) => Number(clip.startMs) || 0));
    const childEnd = Math.max(...group.entries.map(({ clip }) => (Number(clip.startMs) || 0) + Math.max(1, Number(clip.durationMs) || 1)));
    const representative = group.entries.reduce((best, entry) => (Number(entry.clip.startMs) || 0) < (Number(best.clip.startMs) || 0) ? entry : best);
    childContainers.set(childId, {
      representativeKey: representative.key,
      clip: {
        ...representative.clip,
        id: `compound-container:${childId}`,
        kind: 'compound',
        name: group.ref.name || '复合片段',
        compoundId: childId,
        compoundName: group.ref.name || '复合片段',
        startMs: childStart - originMs,
        absoluteStartMs: childStart,
        durationMs: Math.max(1, childEnd - childStart),
        memberCount: group.entries.length,
        memberKeys: group.entries.map(entry => entry.key),
        isCompoundContainer: true,
        isNestedCompound: true
      }
    });
  }
  const tracks = (timeline.tracks || []).map(track => ({
    ...track,
    clips: (track.clips || []).flatMap(clip => {
      const key = `${track.id}::${clip.id}`;
      if (!memberSet.has(key)) return [];
      const path = compoundPathForClip(clip);
      const index = path.findIndex(item => item.id === id);
      const childId = index > 0 ? path[index - 1].id : '';
      const container = childContainers.get(childId);
      if (container) return container.representativeKey === key ? [container.clip] : [];
      return [{ ...clip, startMs: (Number(clip.startMs) || 0) - originMs, absoluteStartMs: Number(clip.startMs) || 0 }];
    })
  })).filter(track => track.clips.length || trackBelongsToCompound(track, id));
  return {
    id,
    name: String(targetRef.name || '复合片段'),
    parentCompoundId,
    originMs,
    durationMs: Math.max(1, endMs - originMs),
    memberKeys,
    tracks
  };
}

function requireMember(view, memberKey) {
  const key = String(memberKey || '');
  if (!view?.memberKeys?.includes(key)) throw new Error('复合成员不存在');
  const [trackId, clipId] = key.split('::');
  return { trackId, clipId };
}

export function compoundMemberMoveCommand(view, memberKey, startOffsetMs) {
  const { trackId, clipId } = requireMember(view, memberKey);
  const offset = finiteNonNegative(startOffsetMs, '组内开始时间不能小于零');
  return { type: 'clip.move', trackId, clipId, startMs: view.originMs + offset, magnetic: false, linked: false };
}

export function compoundMemberResizeCommand(view, memberKey, startOffsetMs, durationMs) {
  const { trackId, clipId } = requireMember(view, memberKey);
  const offset = finiteNonNegative(startOffsetMs, '组内开始时间不能小于零');
  const duration = Number(durationMs);
  if (!Number.isFinite(duration) || duration <= 0) throw new Error('片段时长必须大于零');
  return { type: 'clip.resize', trackId, clipId, edge: 'right', edgeTimeMs: view.originMs + offset + Math.round(duration), ripple: false };
}

export function compoundMemberTrackMoveCommand(view, memberKey, targetTrackId) {
  const { trackId, clipId } = requireMember(view, memberKey);
  const targetId = String(targetTrackId || '');
  requireTrackIndex(view, targetId);
  const sourceTrack = (view.tracks || []).find(track => track.id === trackId);
  const clip = sourceTrack?.clips?.find(item => item.id === clipId);
  if (!clip || clip.isCompoundContainer) throw new Error('复合成员不存在');
  return {
    type: 'compound.member_track_move',
    compoundId: view.id,
    memberKey,
    targetTrackId: targetId,
    startMs: Number.isFinite(Number(clip.absoluteStartMs)) ? Number(clip.absoluteStartMs) : view.originMs + (Number(clip.startMs) || 0)
  };
}

export function compoundMemberTransferCommand(view, memberKey, targetCompoundId, targetTrackId) {
  const { trackId, clipId } = requireMember(view, memberKey);
  const sourceTrack = (view.tracks || []).find(track => track.id === trackId);
  const clip = sourceTrack?.clips?.find(item => item.id === clipId);
  if (!clip || clip.isCompoundContainer) throw new Error('复合成员不存在');
  const targetId = String(targetCompoundId || '').trim();
  const targetTrack = String(targetTrackId || '').trim();
  if (!targetId || !targetTrack) throw new Error('目标复合片段和轨道不能为空');
  return {
    type: 'compound.member_transfer',
    sourceCompoundId: view.id,
    targetCompoundId: targetId,
    memberKey,
    targetTrackId: targetTrack,
    startMs: Number.isFinite(Number(clip.absoluteStartMs)) ? Number(clip.absoluteStartMs) : view.originMs + (Number(clip.startMs) || 0)
  };
}

function compatibleTrackKind(trackKind, clipKind) {
  return trackKind === 'video' ? ['video', 'image'].includes(clipKind) : trackKind === clipKind;
}

export function compoundTransferTargets(timeline, sourceCompoundId, memberKey) {
  const source = compoundEntries(timeline, sourceCompoundId).find(entry => entry.key === String(memberKey || ''));
  if (!source) throw new Error('复合成员不存在');
  const ids = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      for (const ref of compoundPathForClip(clip)) {
        if (ref.id !== sourceCompoundId && !ids.includes(ref.id)) ids.push(ref.id);
      }
    }
  }
  return ids.flatMap(compoundId => {
    const entries = compoundEntries(timeline, compoundId);
    if (entries.length < 2) return [];
    const view = buildCompoundTimelineView(timeline, compoundId);
    const tracks = view.tracks.filter(track => !track.locked && compatibleTrackKind(track.kind, source.clip.kind));
    if (!tracks.length) return [];
    return [{ compoundId, name: view.name, tracks: tracks.map(track => ({ id: track.id, name: track.name || track.id, kind: track.kind })) }];
  });
}

export function compoundBlockTransferCommand(sourceCompoundId, targetCompoundId) {
  const sourceId = String(sourceCompoundId || '').trim();
  const targetId = String(targetCompoundId || '').trim();
  if (!sourceId || !targetId) throw new Error('来源和目标复合片段不能为空');
  return { type: 'compound.block_transfer', sourceCompoundId: sourceId, targetCompoundId: targetId };
}

export function compoundBlockTransferTargets(timeline, sourceCompoundId) {
  const sourceId = String(sourceCompoundId || '').trim();
  const entries = compoundEntries(timeline, sourceId);
  if (entries.length < 2) throw new Error('来源复合片段不存在或成员不足');
  const sourcePath = compoundPathForClip(entries[0].clip);
  const sourceIndex = sourcePath.findIndex(ref => ref.id === sourceId);
  const currentParentId = sourceIndex >= 0 ? sourcePath[sourceIndex + 1]?.id || '' : '';
  const ids = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      for (const ref of compoundPathForClip(clip)) {
        if (ref.id !== sourceId && ref.id !== currentParentId && !ids.includes(ref.id)) ids.push(ref.id);
      }
    }
  }
  return ids.flatMap(compoundId => {
    const targetEntries = compoundEntries(timeline, compoundId);
    if (targetEntries.length < 2) return [];
    const targetPath = compoundPathForClip(targetEntries[0].clip);
    const targetIndex = targetPath.findIndex(ref => ref.id === compoundId);
    if (targetIndex < 0 || targetPath.slice(targetIndex).some(ref => ref.id === sourceId)) return [];
    const ref = targetPath[targetIndex];
    return [{ compoundId, name: ref.name || '复合片段' }];
  });
}

export function compoundMembersTransferCommand(view, memberKeys, targetCompoundId, targetTrackByMember = {}) {
  const keys = (memberKeys || []).map(key => String(key || '').trim()).filter(Boolean);
  if (keys.length < 2) throw new Error('批量迁移至少选择两个成员');
  keys.forEach(key => requireMember(view, key));
  const targetId = String(targetCompoundId || '').trim();
  if (!targetId) throw new Error('目标复合片段不能为空');
  const overrides = Object.fromEntries(Object.entries(targetTrackByMember || {})
    .map(([key, value]) => [String(key || '').trim(), String(value || '').trim()])
    .filter(([key, value]) => keys.includes(key) && value));
  return {
    type: 'compound.members_transfer', sourceCompoundId: view.id, targetCompoundId: targetId, memberKeys: keys, autoMapTracks: true,
    ...(Object.keys(overrides).length ? { targetTrackByMember: overrides } : {})
  };
}

export function compoundBatchTransferTargets(timeline, sourceCompoundId, memberKeys = []) {
  const sourceId = String(sourceCompoundId || '').trim();
  const entries = compoundEntries(timeline, sourceId);
  if (entries.length < 2) throw new Error('来源复合片段不存在或成员不足');
  const ids = [];
  for (const track of timeline?.tracks || []) {
    for (const clip of track.clips || []) {
      for (const ref of compoundPathForClip(clip)) {
        if (ref.id !== sourceId && !ids.includes(ref.id)) ids.push(ref.id);
      }
    }
  }
  const selectedKinds = [...new Set((memberKeys || []).map(key => {
    const entry = entries.find(item => item.key === String(key || ''));
    if (!entry) throw new Error('复合成员不存在');
    return entry.track.kind === 'video' || entry.clip.kind === 'image' ? 'video' : entry.track.kind;
  }))];
  return ids.flatMap(compoundId => {
    const targetEntries = compoundEntries(timeline, compoundId);
    if (targetEntries.length < 2) return [];
    const path = compoundPathForClip(targetEntries[0].clip);
    const index = path.findIndex(ref => ref.id === compoundId);
    if (index < 0 || path.slice(index).some(ref => ref.id === sourceId)) return [];
    return [{ compoundId, name: path[index].name || '复合片段', trackKinds: selectedKinds }];
  });
}

function requireTrackIndex(view, trackId) {
  const index = (view?.tracks || []).findIndex(track => track.id === String(trackId || ''));
  if (index < 0) throw new Error('复合片段内部轨道不存在');
  return index;
}

export function reorderCompoundTrackIds(view, sourceTrackId, targetTrackId) {
  const sourceIndex = requireTrackIndex(view, sourceTrackId);
  const targetIndex = requireTrackIndex(view, targetTrackId);
  const ids = view.tracks.map(track => track.id);
  if (sourceIndex === targetIndex) return ids;
  const [sourceId] = ids.splice(sourceIndex, 1);
  ids.splice(targetIndex, 0, sourceId);
  return ids;
}

export function moveCompoundTrackByOffset(view, trackId, offset) {
  const sourceIndex = requireTrackIndex(view, trackId);
  const targetIndex = Math.max(0, Math.min(view.tracks.length - 1, sourceIndex + Math.sign(Number(offset) || 0)));
  return reorderCompoundTrackIds(view, trackId, view.tracks[targetIndex].id);
}
