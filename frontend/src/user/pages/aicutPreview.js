import { speedCurveRateAt, speedCurveSourceOffsetMs } from './aicutSpeedCurve.js';
import { normalizeAICutTextStyle, textAnimationFrame } from './aicutText.js';
import { sampleKeyframeEasing } from './aicutKeyframeEasing.js';
import { isPersonSegmentationCurrent, normalizePersonSegmentation } from './aicutPersonSegmentation.js';

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

function interpolatedKeyframeValues(clip, offsetMs) {
  const frames = [...(clip.keyframes || [])].filter(frame => Number.isFinite(Number(frame.offsetMs))).sort((a, b) => Number(a.offsetMs) - Number(b.offsetMs));
  if (!frames.length) return null;
  const base = { x: 0, y: 0, scaleX: 1, scaleY: 1, rotation: 0, opacity: 1, volume: finiteNumber(clip.volume, 1) || 1, ...(clip.canvasTransform || {}) };
  const left = [...frames].reverse().find(frame => Number(frame.offsetMs) <= offsetMs) || frames[0];
  const right = frames.find(frame => Number(frame.offsetMs) >= offsetMs) || frames.at(-1);
  const leftValues = { ...base, ...(left.values || {}) };
  const rightValues = { ...base, ...(right.values || {}) };
  const span = Number(right.offsetMs) - Number(left.offsetMs);
  const linearRatio = span > 0 ? Math.min(1, Math.max(0, (offsetMs - Number(left.offsetMs)) / span)) : 0;
  const ratio = span > 0 ? sampleKeyframeEasing(left.easing, linearRatio) : 0;
  return Object.fromEntries(Object.keys(base).map(key => [key, finiteNumber(leftValues[key], base[key]) + (finiteNumber(rightValues[key], base[key]) - finiteNumber(leftValues[key], base[key])) * ratio]));
}

export function nextAICutCanvasFit(value) {
  if (value === 'fit') return 'fill';
  if (value === 'fill') return 'original';
  return 'fit';
}

export function audioFadeGainAtOffset(clip, offsetMs) {
  const durationMs = Math.max(1, finiteNumber(clip?.durationMs, 1));
  const offset = Math.min(durationMs, Math.max(0, finiteNumber(offsetMs)));
  const fadeInMs = Math.min(durationMs, Math.max(0, finiteNumber(clip?.fadeInMs)));
  const fadeOutMs = Math.min(durationMs, Math.max(0, finiteNumber(clip?.fadeOutMs)));
  const fadeInGain = fadeInMs > 0 ? Math.min(1, offset / fadeInMs) : 1;
  const fadeOutGain = fadeOutMs > 0 ? Math.min(1, (durationMs - offset) / fadeOutMs) : 1;
  return Math.max(0, Math.min(1, fadeInGain * fadeOutGain));
}

function previewMedia(track, clip, asset, timeMs) {
  const startMs = Math.max(0, finiteNumber(clip.startMs));
  const outputOffsetMs = timeMs - startMs;
  const hasSpeedCurve = Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2;
  const rate = hasSpeedCurve
    ? speedCurveRateAt(clip.speedCurve, outputOffsetMs / Math.max(1, finiteNumber(clip.durationMs, 1)))
    : (finiteNumber(clip.playbackRate, 1) > 0 ? finiteNumber(clip.playbackRate, 1) : 1);
  const automation = interpolatedKeyframeValues(clip, outputOffsetMs);
	const totalSourceMs = hasSpeedCurve
	  ? speedCurveSourceOffsetMs(Math.max(1, finiteNumber(clip.durationMs, 1)), clip.durationMs, clip.speedCurve)
	  : Math.max(1, finiteNumber(clip.durationMs, 1)) * rate;
	const consumedSourceMs = hasSpeedCurve
	  ? speedCurveSourceOffsetMs(outputOffsetMs, clip.durationMs, clip.speedCurve)
	  : outputOffsetMs * rate;
	const frozen = clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null;
  const sourceTimeMs = frozen
	  ? finiteNumber(clip.freezeSourceMs)
	  : finiteNumber(clip.sourceStartMs) + (clip.reversed ? Math.max(0, totalSourceMs - consumedSourceMs - 1) : consumedSourceMs);
  const baseVolume = Math.max(0, automation ? finiteNumber(automation.volume, 1) : finiteNumber(clip.volume, 1));
  const previewVolume = asset.kind === 'audio' ? baseVolume * audioFadeGainAtOffset(clip, outputOffsetMs) : baseVolume;
  return {
    trackId: track.id,
    clipId: clip.id,
    kind: asset.kind,
    asset,
    startMs,
    durationMs: Math.max(0, finiteNumber(clip.durationMs)),
	sourceTimeMs: Math.max(0, sourceTimeMs),
    playbackRate: rate,
	reversed: Boolean(clip.reversed),
	frozen,
    volume: previewVolume,
    audioProcessing: clip.audioProcessing || null,
    transition: clip.transition || '',
    transitionMs: Math.max(0, finiteNumber(clip.transitionMs)),
    visualEffect: clip.visualEffect || null,
	visualProcessing: clip.visualProcessing || null,
	outputOffsetMs,
	blendMode: clip.blendMode || 'normal',
    canvasTransform: automation ? { x: automation.x, y: automation.y, scaleX: automation.scaleX, scaleY: automation.scaleY, rotation: automation.rotation, opacity: automation.opacity } : (clip.canvasTransform || null)
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
        const elapsedMs = timeMs - Math.max(0, finiteNumber(clip.startMs));
        const animation = track.settings?.subtitleAnimation || 'none';
        frame.subtitles.push({
          trackId: track.id,
          clipId: clip.id,
          text: String(clip.text).trim(),
          styleId: track.settings?.textResourceId || track.settings?.subtitleStyleId || '黑底白字',
          rhythm: track.settings?.subtitleRhythm || 'sentence',
          animation,
          textStyle: normalizeAICutTextStyle(track.settings),
          ...textAnimationFrame(String(clip.text).trim(), animation, elapsedMs, Math.max(1, finiteNumber(clip.durationMs, 1)))
        });
        continue;
      }
      const asset = assetMap.get(clip.assetId);
      if (!asset) continue;
      const media = previewMedia(track, clip, asset, timeMs);
      if (asset.kind === 'video') {
        const personSegmentation = normalizePersonSegmentation(clip.visualProcessing?.personSegmentation);
        media.personSegmentationCurrent = isPersonSegmentationCurrent(clip, personSegmentation);
        media.matteAsset = media.personSegmentationCurrent ? (assetMap.get(personSegmentation.matteAssetId) || null) : null;
        if (!media.matteAsset) media.personSegmentationCurrent = false;
      }
      if (asset.kind === 'audio') frame.audios.push(media);
      else if (asset.kind === 'video' || asset.kind === 'image') frame.visuals.push(media);
    }
  }
  return frame;
}
