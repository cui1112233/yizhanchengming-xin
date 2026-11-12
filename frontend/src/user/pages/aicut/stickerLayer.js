export function addStickerLayer(timeline, asset, startMs = 0, durationMs = 5000, idFactory = prefix => `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2)}`) {
  if (!asset?.id || asset.kind !== 'image') throw new Error('请选择本工程的图片贴纸');
  if (!Number.isSafeInteger(startMs) || startMs < 0 || !Number.isSafeInteger(durationMs) || durationMs < 1 || startMs + durationMs > 86400000) throw new Error('贴纸时间范围无效');
  const tracks = timeline?.tracks || [];
  const used = new Set(tracks.flatMap(track => [track.id, ...(track.clips || []).map(clip => clip.id)]));
  const unique = (prefix) => {
    for (let attempt = 0; attempt < 100; attempt++) {
      const value = idFactory(prefix);
      if (typeof value === 'string' && value && value.length <= 96 && !value.includes('::') && !used.has(value)) { used.add(value); return value; }
    }
    throw new Error('无法生成贴纸编号');
  };
  const trackId = unique('track-sticker');
  const clipId = unique('clip-sticker');
  const track = { id: trackId, name: `贴纸 · ${asset.name || '图片'}`, kind: 'video', order: Math.max(-1, ...tracks.map(track => Number(track.order) || 0)) + 1, locked: false, muted: false, clips: [{ id: clipId, assetId: asset.id, kind: 'image', name: asset.name || '图片贴纸', startMs, durationMs, sourceStartMs: 0, sourceDurationMs: durationMs, canvasTransform: { x: 0, y: 0, scaleX: .35, scaleY: .35, rotation: 0, opacity: 1 } }] };
  return { timeline: { ...timeline, tracks: [...tracks, track] }, selectionKey: `${trackId}::${clipId}` };
}
