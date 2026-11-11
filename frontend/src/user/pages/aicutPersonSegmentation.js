const WASM_ROOT = '/aicut/mediapipe/wasm';
const MODEL_URL = '/aicut/mediapipe/models/selfie_segmenter.tflite';
const DEFAULTS = Object.freeze({ enabled: false, matteAssetId: '', threshold: .5, feather: .1, signature: '' });

let segmenterPromise;
let lastPersonSegmentationTimestamp = -1;

function clamp(value, minimum, maximum, fallback) {
  const number = Number(value);
  return Number.isFinite(number) ? Math.min(maximum, Math.max(minimum, number)) : fallback;
}

function normalizedRate(value) {
  const rate = Number(value);
  return Number.isFinite(rate) && rate > 0 ? rate : 1;
}

export function nextPersonSegmentationTimestamp(previous, candidate) {
  const last = Number.isFinite(Number(previous)) ? Math.round(Number(previous)) : -1;
  const requested = Number.isFinite(Number(candidate)) ? Math.max(0, Math.round(Number(candidate))) : 0;
  return Math.max(requested, last + 1);
}

export function normalizePersonSegmentation(value = {}) {
  return {
    enabled: Boolean(value?.enabled),
    matteAssetId: String(value?.matteAssetId || '').trim(),
    threshold: clamp(value?.threshold, 0, 1, DEFAULTS.threshold),
    feather: clamp(value?.feather, 0, .5, DEFAULTS.feather),
    signature: String(value?.signature || '').trim()
  };
}

export function buildPersonSegmentationSignature(clip = {}) {
  const freeze = clip.freezeSourceMs === undefined || clip.freezeSourceMs === null ? 'none' : String(Math.round(Number(clip.freezeSourceMs) || 0));
  return [
    String(clip.assetId || '').trim(),
    String(Math.max(0, Math.round(Number(clip.sourceStartMs) || 0))),
    String(Math.max(1, Math.round(Number(clip.durationMs) || 1))),
    String(normalizedRate(clip.playbackRate)),
    clip.reversed ? '1' : '0',
    freeze
  ].join('|');
}

export function isPersonSegmentationCurrent(clip, settings) {
  const normalized = normalizePersonSegmentation(settings);
  return normalized.enabled && Boolean(normalized.matteAssetId) && normalized.signature === buildPersonSegmentationSignature(clip);
}

export function isAICutPersonMatteAsset(asset) {
  return String(asset?.name || '').startsWith('AI-CUT人物蒙版_');
}

export function personMaskToRGBA(confidence, width, height, settings = {}) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  if (!(confidence instanceof Float32Array) || columns < 1 || rows < 1 || confidence.length !== columns * rows) throw new Error('人物蒙版尺寸无效');
  const { threshold, feather } = normalizePersonSegmentation(settings);
  const lower = threshold - feather / 2;
  const upper = threshold + feather / 2;
  const result = new Uint8ClampedArray(confidence.length * 4);
  for (let index = 0; index < confidence.length; index += 1) {
    const value = Math.min(1, Math.max(0, Number(confidence[index]) || 0));
    const alpha = feather <= 0
      ? (value >= threshold ? 255 : 0)
      : (value <= lower ? 0 : value >= upper ? 255 : Math.round((value - lower) / Math.max(.000001, upper - lower) * 255));
    const offset = index * 4;
    result[offset] = 255;
    result[offset + 1] = 255;
    result[offset + 2] = 255;
    result[offset + 3] = alpha;
  }
  return result;
}

function abortError() {
  return new DOMException('人物蒙版生成已取消', 'AbortError');
}

function throwIfAborted(signal) {
  if (signal?.aborted) throw abortError();
}

function waitForEvent(target, success, failure, signal) {
  return new Promise((resolve, reject) => {
    const cleanup = () => {
      target.removeEventListener(success, onSuccess);
      if (failure) target.removeEventListener(failure, onFailure);
      signal?.removeEventListener('abort', onAbort);
    };
    const onSuccess = () => { cleanup(); resolve(); };
    const onFailure = () => { cleanup(); reject(new Error('视频素材读取失败')); };
    const onAbort = () => { cleanup(); reject(abortError()); };
    target.addEventListener(success, onSuccess, { once: true });
    if (failure) target.addEventListener(failure, onFailure, { once: true });
    signal?.addEventListener('abort', onAbort, { once: true });
  });
}

async function seekVideo(video, seconds, signal) {
  throwIfAborted(signal);
  if (Math.abs(Number(video.currentTime) - seconds) < .01) return;
  const waiting = waitForEvent(video, 'seeked', 'error', signal);
  video.currentTime = seconds;
  await waiting;
}

async function loadPersonSegmenter() {
  if (!segmenterPromise) {
    segmenterPromise = import('@mediapipe/tasks-vision').then(async ({ FilesetResolver, ImageSegmenter }) => {
      const files = await FilesetResolver.forVisionTasks(WASM_ROOT);
      return ImageSegmenter.createFromOptions(files, {
        baseOptions: { modelAssetPath: MODEL_URL, delegate: 'CPU' },
        runningMode: 'VIDEO',
        outputConfidenceMasks: true,
        outputCategoryMask: false
      });
    }).catch((error) => {
      segmenterPromise = null;
      throw error;
    });
  }
  return segmenterPromise;
}

function supportedRecorderType() {
  if (typeof MediaRecorder === 'undefined') throw new Error('当前浏览器不支持本地人物蒙版录制');
  for (const value of ['video/webm;codecs=vp9', 'video/webm;codecs=vp8', 'video/webm']) {
    if (!MediaRecorder.isTypeSupported || MediaRecorder.isTypeSupported(value)) return value;
  }
  throw new Error('当前浏览器没有可用的 WebM 编码器');
}

function grayscaleMaskImage(mask) {
  const confidence = mask.getAsFloat32Array();
  const pixels = new Uint8ClampedArray(confidence.length * 4);
  for (let index = 0; index < confidence.length; index += 1) {
    const value = Math.round(Math.min(1, Math.max(0, Number(confidence[index]) || 0)) * 255);
    const offset = index * 4;
    pixels[offset] = value;
    pixels[offset + 1] = value;
    pixels[offset + 2] = value;
    pixels[offset + 3] = 255;
  }
  return new ImageData(pixels, mask.width, mask.height);
}

export async function generatePersonMatteVideo({ sourceURL, startMs = 0, durationMs, playbackRate = 1, signal, onProgress } = {}) {
  const outputDuration = Math.round(Number(durationMs));
  if (!String(sourceURL || '').trim()) throw new Error('视频素材地址为空');
  if (!Number.isFinite(outputDuration) || outputDuration < 100) throw new Error('人物抠像片段至少需要 0.1 秒');
  if (outputDuration > 300000) throw new Error('单次人物抠像最长支持 5 分钟，请先分割片段');
  throwIfAborted(signal);
  const segmenter = await loadPersonSegmenter();
  throwIfAborted(signal);

  const video = document.createElement('video');
  video.crossOrigin = 'anonymous';
  video.preload = 'auto';
  video.muted = true;
  video.playsInline = true;
  video.src = sourceURL;
  if (video.readyState < 1) await waitForEvent(video, 'loadedmetadata', 'error', signal);
  const rate = normalizedRate(playbackRate);
  const sourceStart = Math.max(0, Number(startMs) || 0) / 1000;
  const sourceEnd = Math.min(Number(video.duration) || Infinity, sourceStart + outputDuration / 1000 * rate);
  if (!Number.isFinite(sourceEnd) || sourceEnd <= sourceStart) throw new Error('视频裁剪区间无效');
  await seekVideo(video, sourceStart, signal);
  video.playbackRate = rate;

  const width = Math.max(64, Math.round(Math.min(256, Number(video.videoWidth) || 256)));
  const height = Math.max(64, Math.round(width * (Number(video.videoHeight) || width) / Math.max(1, Number(video.videoWidth) || width)));
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const context = canvas.getContext('2d', { alpha: false });
  if (!context || typeof canvas.captureStream !== 'function') throw new Error('当前浏览器无法创建人物蒙版画布');
  context.fillStyle = '#000';
  context.fillRect(0, 0, width, height);

  const stream = canvas.captureStream(30);
  const mimeType = supportedRecorderType();
  const recorder = new MediaRecorder(stream, { mimeType, videoBitsPerSecond: 900000 });
  const chunks = [];
  recorder.addEventListener('dataavailable', event => { if (event.data?.size) chunks.push(event.data); });
  // 取消时会主动 stop；这里不能让 abort 提前拒绝，否则流和录像轨道来不及清理。
  const stopped = waitForEvent(recorder, 'stop', 'error');
  let frameHandle = 0;
  let lastInferenceMs = -Infinity;
  let running = true;
  const stop = () => {
    if (!running) return;
    running = false;
    cancelAnimationFrame(frameHandle);
    video.pause();
    if (recorder.state !== 'inactive') recorder.stop();
  };
  const onAbort = () => stop();
  signal?.addEventListener('abort', onAbort, { once: true });
  try {
    recorder.start(250);
    await video.play();
    const tick = () => {
      if (!running) return;
      try {
        throwIfAborted(signal);
        const elapsedMs = Math.max(0, (video.currentTime - sourceStart) / rate * 1000);
        if (video.ended || video.currentTime >= sourceEnd - .002 || elapsedMs >= outputDuration) {
          onProgress?.(1);
          stop();
          return;
        }
        if (elapsedMs - lastInferenceMs >= 60 && video.readyState >= 2) {
          lastInferenceMs = elapsedMs;
          // MediaPipe 的 VIDEO 模式要求同一个模型实例收到的时间戳始终严格递增。
          // 素材二次运行或换到更早裁剪点时，媒体时间会倒退，必须改用会话级单调时间戳。
          lastPersonSegmentationTimestamp = nextPersonSegmentationTimestamp(
            lastPersonSegmentationTimestamp,
            video.currentTime * 1000
          );
          const result = segmenter.segmentForVideo(video, lastPersonSegmentationTimestamp);
          const mask = result.confidenceMasks?.at(-1);
          if (!mask) throw new Error('人物模型没有返回置信度蒙版');
          const image = grayscaleMaskImage(mask);
          const maskCanvas = document.createElement('canvas');
          maskCanvas.width = mask.width;
          maskCanvas.height = mask.height;
          maskCanvas.getContext('2d').putImageData(image, 0, 0);
          context.drawImage(maskCanvas, 0, 0, width, height);
          result.close?.();
          onProgress?.(Math.min(.99, elapsedMs / outputDuration));
        }
        frameHandle = requestAnimationFrame(tick);
      } catch (error) {
        recorder.__aicutError = error;
        stop();
      }
    };
    frameHandle = requestAnimationFrame(tick);
    await stopped;
    throwIfAborted(signal);
    if (recorder.__aicutError) throw recorder.__aicutError;
    const blob = new Blob(chunks, { type: mimeType });
    if (!blob.size) throw new Error('人物模型没有生成有效蒙版视频');
    return { blob, mimeType, width, height, durationMs: outputDuration };
  } finally {
    stop();
    signal?.removeEventListener('abort', onAbort);
    stream.getTracks().forEach(track => track.stop());
    video.removeAttribute('src');
    video.load();
  }
}

export function startPersonMattePreviewRenderer(canvas, source, maskVideo, settings) {
  const context = canvas?.getContext?.('2d');
  if (!context || !source || !maskVideo) return null;
  const maskCanvas = document.createElement('canvas');
  const maskContext = maskCanvas.getContext('2d', { willReadFrequently: true });
  let stopped = false;
  let frame = 0;
  const render = () => {
    if (stopped) return;
    const width = Number(source.width || source.videoWidth || source.naturalWidth);
    const height = Number(source.height || source.videoHeight || source.naturalHeight);
    if (width > 0 && height > 0 && maskVideo.readyState >= 2) {
      if (canvas.width !== width || canvas.height !== height) { canvas.width = width; canvas.height = height; }
      maskCanvas.width = Math.max(1, Number(maskVideo.videoWidth) || 256);
      maskCanvas.height = Math.max(1, Number(maskVideo.videoHeight) || 256);
      maskContext.drawImage(maskVideo, 0, 0, maskCanvas.width, maskCanvas.height);
      const image = maskContext.getImageData(0, 0, maskCanvas.width, maskCanvas.height);
      const confidence = new Float32Array(maskCanvas.width * maskCanvas.height);
      for (let index = 0; index < confidence.length; index += 1) confidence[index] = image.data[index * 4] / 255;
      const alphaPixels = personMaskToRGBA(confidence, maskCanvas.width, maskCanvas.height, settings);
      maskContext.putImageData(new ImageData(alphaPixels, maskCanvas.width, maskCanvas.height), 0, 0);
      context.clearRect(0, 0, width, height);
      context.globalCompositeOperation = 'source-over';
      context.drawImage(source, 0, 0, width, height);
      context.globalCompositeOperation = 'destination-in';
      context.drawImage(maskCanvas, 0, 0, width, height);
      context.globalCompositeOperation = 'source-over';
    }
    frame = requestAnimationFrame(render);
  };
  frame = requestAnimationFrame(render);
  return { stop() { stopped = true; cancelAnimationFrame(frame); } };
}
