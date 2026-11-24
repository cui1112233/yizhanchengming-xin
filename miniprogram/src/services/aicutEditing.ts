import type { CutRecord, CutTimeline } from './aicut';
const clone = (value: CutTimeline): CutTimeline => JSON.parse(JSON.stringify(value));
const locate = (value: CutTimeline, trackId: string, clipId: string) => {
  const track = value.tracks.find(item => item.id === trackId);
  if (!track) throw new Error('轨道不存在，请重新读取');
  if (track.locked) throw new Error('轨道已锁定，请在网页解除锁定');
  const clip = track.clips.find((item: CutRecord) => item.id === clipId);
  if (!clip) throw new Error('片段不存在');
  return { track, clip };
};
export function basicEditReason(clip: CutRecord) {
  return clip.speedCurve?.length || clip.keyframes?.length || clip.linkGroupId || clip.compoundId || clip.compoundPath?.length || clip.reversed || clip.freezeSourceMs != null || clip.transition || clip.visualProcessing || clip.audioProcessing || clip.visualEffect || clip.animation ? '该片段含复杂效果，请在网页编辑裁剪；手机保留原设置' : '';
}
export function editClip(value: CutTimeline, trackId: string, clipId: string, patch: CutRecord) {
  const next = clone(value); const { clip } = locate(next, trackId, clipId);
  if (basicEditReason(clip)) throw new Error(basicEditReason(clip));
  const fields = Object.fromEntries(Object.entries(patch).filter(([key]) => ['name', 'text', 'startMs', 'sourceStartMs', 'durationMs'].includes(key)));
  Object.assign(clip, fields);
  if (![clip.startMs, clip.sourceStartMs, clip.durationMs].every(Number.isSafeInteger) || clip.startMs < 0 || clip.sourceStartMs < 0 || clip.durationMs <= 0) throw new Error('请输入有效的开始时间与时长');
  if (['video', 'audio'].includes(clip.kind) && clip.sourceDurationMs > 0 && clip.sourceStartMs + clip.durationMs * (clip.playbackRate || 1) > clip.sourceDurationMs) throw new Error('裁剪范围超过素材时长');
  if ((clip.fadeInMs || 0) > clip.durationMs || (clip.fadeOutMs || 0) > clip.durationMs) throw new Error('时长不能短于原有淡入淡出，请先在网页调整');
  return next;
}
export function removeClip(value: CutTimeline, trackId: string, clipId: string) {
  const next = clone(value); const { track, clip } = locate(next, trackId, clipId);
  if (basicEditReason(clip)) throw new Error('该片段含复杂联动，请在网页删除');
  track.clips = track.clips.filter((item: CutRecord) => item.id !== clipId); return next;
}
export function addAsset(value: CutTimeline, asset: CutRecord | null, id: string, text = '') {
  const next = clone(value); const kind = asset?.kind || 'subtitle';
  if (!['video', 'image', 'audio', 'subtitle'].includes(kind)) throw new Error('该素材类型暂不支持放入手机时间线');
  const trackKind = kind === 'image' ? 'video' : kind;
  let track = next.tracks.find(item => item.kind === trackKind && !item.locked);
  if (!track) { track = { id: `track-${id}`, kind: trackKind, name: ({ video: '画面', audio: '音频', subtitle: '字幕' } as CutRecord)[trackKind], order: next.tracks.length, muted: false, locked: false, clips: [] }; next.tracks.push(track); }
  if (asset && kind !== 'image' && !(asset.durationMs > 0)) throw new Error('素材时长尚未读取，请刷新后再添加');
  const startMs = track.clips.reduce((end: number, clip: CutRecord) => Math.max(end, clip.startMs + clip.durationMs), 0);
  track.clips.push({ id, kind, name: asset?.name || '字幕', ...(asset ? { assetId: asset.id } : { text }), startMs, durationMs: kind === 'image' || !asset ? 3000 : asset.durationMs, sourceStartMs: 0, sourceDurationMs: asset?.durationMs || 0, volume: 1 });
  return next;
}
