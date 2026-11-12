import { speedCurveSourceOffsetMs } from '../aicutSpeedCurve.js';

const number = (value, fallback = 0) => Number.isFinite(Number(value)) ? Number(value) : fallback;

function sourceAt(clip, fraction, durationMs) {
  if (!clip) return durationMs * fraction;
  if (clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null) return Math.max(0, number(clip.freezeSourceMs));
  const duration = Math.max(0, number(clip.durationMs));
  const offset = position => Array.isArray(clip.speedCurve) && clip.speedCurve.length >= 2
    ? speedCurveSourceOffsetMs(duration * position, duration, clip.speedCurve)
    : duration * position * Math.max(.01, number(clip.playbackRate, 1));
  return Math.max(0, number(clip.sourceStartMs)) + (clip.reversed ? Math.max(0, offset(1) - offset(fraction) - 1) : offset(fraction));
}

export function artworkFrameTimes(asset = {}, clip, strip = false) {
  const duration = Math.max(0, number(asset.durationMs));
  return (strip ? [0, .5, 1] : [0]).map(position => {
    const atExclusiveEnd = position === 1 && clip && !clip.reversed && clip.freezeSourceMs == null && number(clip.durationMs) > 0;
    const sampledTime = sourceAt(clip, position, duration);
    const time = atExclusiveEnd ? Math.max(number(clip.sourceStartMs), sampledTime - 1) : sampledTime;
    return Math.round(duration > 0 ? Math.min(Math.max(0, duration - 1), time) : time);
  });
}

// Peaks represent source audio only; processing effects are deliberately not invented.
export function artworkPeaks(analysis, clip, count = 120) {
  const peaks = analysis?.peaks;
  const duration = number(analysis?.durationMs);
  if (!Array.isArray(peaks) || !peaks.length || duration <= 0) return [];
  const size = Math.max(1, Math.min(240, Math.round(number(count, 120))));
  return Array.from({ length: size }, (_, index) => {
    const time = sourceAt(clip, (index + .5) / size, duration);
    const position = Math.max(0, Math.min(peaks.length - 1, Math.floor(time / duration * peaks.length)));
    return Math.min(1, Math.max(0, Math.abs(number(peaks[position]))));
  });
}

const aborted = () => Object.assign(new Error('素材预览已取消'), { name: 'AbortError' });

// Shared subscribers keep a single decode/request alive; offscreen-only work is cancelled.
export function createArtworkQueue({ concurrency = 2, maxCache = 100, maxPending = 64 } = {}) {
  const cache = new Map();
  const entries = new Map();
  const pending = [];
  let active = 0;
  function pump() {
    while (active < concurrency && pending.length) {
      const entry = pending.shift();
      if (!entry.refs) continue;
      entry.started = true;
      active++;
      Promise.resolve().then(() => entry.task(entry.controller.signal)).then(value => {
        if (!entry.controller.signal.aborted) {
          cache.set(entry.key, { value });
          while (cache.size > maxCache) cache.delete(cache.keys().next().value);
        }
        entry.resolve(value);
      }, error => {
        // Failed decoding is also remembered, avoiding repeated failing media loads.
        if (!entry.controller.signal.aborted && error?.name !== 'AbortError') {
          cache.set(entry.key, { error });
          while (cache.size > maxCache) cache.delete(cache.keys().next().value);
        }
        entry.reject(error);
      }).finally(() => {
        if (entries.get(entry.key) === entry) entries.delete(entry.key);
        active--;
        pump();
      });
    }
  }
  return {
    acquire(key, task) {
      if (cache.has(key)) {
        const result = cache.get(key);
        cache.delete(key);
        cache.set(key, result);
        return { promise: result.error ? Promise.reject(result.error) : Promise.resolve(result.value), release() {} };
      }
      let entry = entries.get(key);
      if (!entry) {
        if (pending.length >= maxPending) return { promise: Promise.reject(new Error('素材预览队列已满')), release() {} };
        entry = { key, task, refs: 0, started: false, controller: new AbortController() };
        entry.promise = new Promise((resolve, reject) => { entry.resolve = resolve; entry.reject = reject; });
        entries.set(key, entry);
        pending.push(entry);
      }
      entry.refs++;
      pump();
      let released = false;
      return { promise: entry.promise, release() {
        if (released) return;
        released = true;
        if (--entry.refs || !entries.has(key)) return;
        entry.controller.abort();
        if (entries.get(key) === entry) entries.delete(key);
        const index = pending.indexOf(entry);
        if (index >= 0) pending.splice(index, 1);
        entry.reject(aborted());
      } };
    },
  };
}

export const videoArtworkQueue = createArtworkQueue();
export const audioArtworkQueue = createArtworkQueue();

export function decodeArtworkFrame(url, timeMs, signal) {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) { reject(aborted()); return; }
    const video = document.createElement('video');
    let finished = false;
    let timer;
    const finish = (error, value) => {
      if (finished) return;
      finished = true;
      clearTimeout(timer);
      signal?.removeEventListener('abort', cancel);
      video.onloadedmetadata = video.onloadeddata = video.onseeked = video.onerror = null;
      video.pause();
      video.removeAttribute('src');
      video.load();
      if (error) reject(error); else resolve(value);
    };
    const cancel = () => finish(aborted());
    const capture = () => {
      if (!video.videoWidth || video.readyState < 2) return;
      try {
        const canvas = document.createElement('canvas');
        const scale = Math.min(240 / video.videoWidth, 240 / video.videoHeight);
        canvas.width = Math.max(1, Math.round(video.videoWidth * scale));
        canvas.height = Math.max(1, Math.round(video.videoHeight * scale));
        canvas.getContext('2d').drawImage(video, 0, 0, canvas.width, canvas.height);
        finish(null, canvas.toDataURL('image/jpeg', .72));
      } catch (error) { finish(error); }
    };
    signal?.addEventListener('abort', cancel, { once: true });
    timer = setTimeout(() => finish(new Error('视频预览读取超时')), 12000);
    video.crossOrigin = 'anonymous';
    video.muted = true;
    video.preload = 'auto';
    video.playsInline = true;
    let target = 0;
    video.onloadedmetadata = () => {
      target = Math.max(0, timeMs / 1000);
      if (Number.isFinite(video.duration)) target = Math.min(target, Math.max(0, video.duration - .04));
      if (target > 0) video.currentTime = target;
      else capture();
    };
    video.onloadeddata = () => { if (target === 0) capture(); };
    video.onseeked = capture;
    video.onerror = () => finish(new Error('视频预览暂不可用'));
    video.src = url;
    video.load();
  });
}
