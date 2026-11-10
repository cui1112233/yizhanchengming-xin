function finiteNumber(value, fallback = 0) {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
}

function orderedTracks(timeline) {
  return [...(timeline?.tracks || [])].sort((left, right) => finiteNumber(left.order) - finiteNumber(right.order));
}

function isActive(clip, timeMs) {
  const startMs = Math.max(0, finiteNumber(clip?.startMs));
  const durationMs = Math.max(0, finiteNumber(clip?.durationMs));
  return durationMs > 0 && timeMs >= startMs && timeMs < startMs + durationMs;
}

export function nextAICutCanvasFit(value) {
  if (value === 'fit') return 'fill';
  if (value === 'fill') return 'original';
  return 'fit';
}

function previewMedia(track, clip, asset, timeMs) {
  const startMs = Math.max(0, finiteNumber(clip.startMs));
  const rate = finiteNumber(clip.playbackRate, 1) > 0 ? finiteNumber(clip.playbackRate, 1) : 1;
  return {
    trackId: track.id,
    clipId: clip.id,
    kind: asset.kind,
    asset,
    startMs,
    durationMs: Math.max(0, finiteNumber(clip.durationMs)),
    sourceTimeMs: Math.max(0, finiteNumber(clip.sourceStartMs) + (timeMs - startMs) * rate),
    playbackRate: rate,
    volume: finiteNumber(clip.volume, 1) > 0 ? finiteNumber(clip.volume, 1) : 1,
    transition: clip.transition || '',
    transitionMs: Math.max(0, finiteNumber(clip.transitionMs)),
    visualEffect: clip.visualEffect || null,
    canvasTransform: clip.canvasTransform || null
  };
}

export function getAICutTimelineDuration(timeline) {
  return orderedTracks(timeline).flatMap(track => track.clips || []).reduce((maximum, clip) => {
    return Math.max(maximum, Math.max(0, finiteNumber(clip.startMs)) + Math.max(0, finiteNumber(clip.durationMs)));
  }, 0);
}

export function seekAICutPreviewTime(value, durationMs) {
  const maximum = Math.max(0, finiteNumber(durationMs));
  return Math.min(maximum, Math.max(0, finiteNumber(value)));
}

export function buildAICutPreviewFrame(timeline, assets, requestedTimeMs) {
  const durationMs = getAICutTimelineDuration(timeline);
  const timeMs = seekAICutPreviewTime(requestedTimeMs, durationMs);
  const assetMap = new Map((assets || []).map(asset => [asset.id, asset]));
  const frame = { timeMs, visuals: [], audios: [], subtitles: [] };

  for (const track of orderedTracks(timeline)) {
    if (track.muted) continue;
    for (const clip of track.clips || []) {
      if (!isActive(clip, timeMs)) continue;
      if (track.kind === 'subtitle' || clip.kind === 'subtitle') {
        if (track.settings?.previewVisible === false || !String(clip.text || '').trim()) continue;
        frame.subtitles.push({
          trackId: track.id,
          clipId: clip.id,
          text: String(clip.text).trim(),
          styleId: track.settings?.subtitleStyleId || '黑底白字',
          rhythm: track.settings?.subtitleRhythm || 'sentence',
          animation: track.settings?.subtitleAnimation || '无动效'
        });
        continue;
      }
      const asset = assetMap.get(clip.assetId);
      if (!asset) continue;
      const media = previewMedia(track, clip, asset, timeMs);
      if (asset.kind === 'audio') frame.audios.push(media);
      else if (asset.kind === 'video' || asset.kind === 'image') frame.visuals.push(media);
    }
  }
  return frame;
}
