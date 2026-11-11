export function compoundPathForClip(clip) {
  const path = Array.isArray(clip?.compoundPath)
    ? clip.compoundPath.map(item => ({ id: String(item?.id || '').trim(), name: String(item?.name || '').trim() })).filter(item => item.id)
    : [];
  if (path.length) return path;
  const id = String(clip?.compoundId || '').trim();
  return id ? [{ id, name: String(clip?.compoundName || '复合片段').trim() || '复合片段' }] : [];
}

export function clipHasCompound(clip, compoundId) {
  const id = String(compoundId || '').trim();
  return Boolean(id) && compoundPathForClip(clip).some(item => item.id === id);
}

export function outerCompoundRef(clip) {
  return compoundPathForClip(clip).at(-1) || null;
}

export function compoundRefForClip(clip, compoundId) {
  const id = String(compoundId || '').trim();
  return compoundPathForClip(clip).find(item => item.id === id) || null;
}

export function applyCompoundPath(clip, path) {
  const next = { ...clip };
  delete next.compoundId;
  delete next.compoundName;
  delete next.compoundPath;
  const normalized = (path || []).map(item => ({ id: String(item?.id || '').trim(), name: String(item?.name || '').trim() || '复合片段' })).filter(item => item.id);
  if (!normalized.length) return next;
  const outer = normalized.at(-1);
  return { ...next, compoundId: outer.id, compoundName: outer.name, compoundPath: normalized };
}

export function replaceCompoundRef(clip, compoundId, nextRef) {
  const id = String(compoundId || '').trim();
  return applyCompoundPath(clip, compoundPathForClip(clip).map(item => item.id === id ? { ...nextRef } : item));
}

export function removeCompoundRefs(clip, compoundIds) {
  const ids = compoundIds instanceof Set ? compoundIds : new Set(compoundIds || []);
  return applyCompoundPath(clip, compoundPathForClip(clip).filter(item => !ids.has(item.id)));
}

export function compoundEntries(timeline, compoundId) {
  const id = String(compoundId || '').trim();
  if (!id) return [];
  return (timeline?.tracks || []).flatMap(track => (track.clips || []).flatMap(clip =>
    clipHasCompound(clip, id) ? [{ track, clip, key: `${track.id}::${clip.id}` }] : []
  ));
}

export function trackBelongsToCompound(track, compoundId) {
  return String(track?.settings?.compoundId || '').trim() === String(compoundId || '').trim();
}
