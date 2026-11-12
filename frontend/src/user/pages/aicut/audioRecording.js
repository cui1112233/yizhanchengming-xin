const MIME_TYPES = ['audio/webm;codecs=opus', 'audio/ogg;codecs=opus', 'audio/mp4', 'audio/webm', 'audio/ogg'];

export function chooseAudioMimeType(Recorder) {
  return MIME_TYPES.find(type => Recorder?.isTypeSupported?.(type)) || '';
}

function recordingError(error) {
  if (['NotAllowedError', 'SecurityError'].includes(error?.name)) return '未获得麦克风权限，请在浏览器中允许后重新开始。';
  if (error?.name === 'NotFoundError') return '没有找到麦克风，请连接麦克风后重试。';
  if (['NotReadableError', 'AbortError'].includes(error?.name)) return '无法读取麦克风，可能被其他应用占用，请检查后重试。';
  return error?.message || '录音中断，请检查麦克风后重试。';
}

// All browser resources belong to this session, including late permission results.
export function createAudioRecording(options = {}) {
  const Recorder = options.MediaRecorder ?? globalThis.MediaRecorder;
  const getUserMedia = options.getUserMedia ?? (constraints => globalThis.navigator.mediaDevices.getUserMedia(constraints));
  const createURL = options.createObjectURL ?? (blob => URL.createObjectURL(blob));
  const revokeURL = options.revokeObjectURL ?? (url => URL.revokeObjectURL(url));
  const now = options.now ?? Date.now;
  const schedule = options.setInterval ?? globalThis.setInterval;
  const unschedule = options.clearInterval ?? globalThis.clearInterval;
  const maxDurationMs = options.maxDurationMs ?? 600000;
  const maxBytes = options.maxBytes ?? 30 * 1024 * 1024;
  let state = { status: 'idle', elapsedMs: 0, file: null, url: '', error: '', notice: '', canPause: false };
  let stream = null; let recorder = null; let timer = null; let chunks = []; let bytes = 0;
  let generation = 0; let disposed = false; let startedAt = 0; let segmentStart = null; let accumulated = 0; let saving = null;
  const emit = patch => { state = { ...state, ...patch }; if (!disposed) options.onChange?.(state); };
  const elapsed = () => accumulated + (segmentStart === null ? 0 : Math.max(0, now() - segmentStart));
  function release() {
    if (timer !== null) { unschedule(timer); timer = null; }
    if (recorder) { recorder.ondataavailable = null; recorder.onstop = null; recorder.onerror = null; }
    stream?.getTracks().forEach(track => { track.onended = null; track.stop(); }); stream = null;
  }
  function stop(notice = '') {
    if (!recorder || !['recording', 'paused'].includes(state.status)) return;
    accumulated = elapsed(); segmentStart = null;
    emit({ status: 'stopping', elapsedMs: accumulated, notice: notice || state.notice });
    if (timer !== null) { unschedule(timer); timer = null; }
    try { if (recorder.state !== 'inactive') recorder.stop(); }
    catch (error) { release(); chunks = []; emit({ status: 'error', error: recordingError(error) }); }
  }
  function discard() {
    generation++;
    const previous = recorder; release();
    try { if (previous && previous.state !== 'inactive') previous.stop(); } catch { /* Already ended. */ }
    recorder = null; chunks = []; bytes = 0; accumulated = 0; segmentStart = null; saving = null;
    if (state.url) revokeURL(state.url);
    emit({ status: 'idle', elapsedMs: 0, file: null, url: '', error: '', notice: '', canPause: false });
  }
  async function start() {
    if (disposed || ['requesting', 'recording', 'paused', 'stopping', 'uploading'].includes(state.status)) return;
    discard(); const attempt = generation;
    const mimeType = chooseAudioMimeType(Recorder);
    if (!Recorder || !mimeType || (!options.getUserMedia && !globalThis.navigator?.mediaDevices?.getUserMedia)) {
      emit({ status: 'error', error: '当前浏览器不支持麦克风录音，请使用新版 Chrome、Edge 或 Safari，并通过安全连接访问。' }); return;
    }
    emit({ status: 'requesting' });
    try {
      const acquired = await getUserMedia({ audio: true });
      if (disposed || attempt !== generation) { acquired.getTracks().forEach(track => track.stop()); return; }
      stream = acquired; recorder = new Recorder(stream, { mimeType, audioBitsPerSecond: 128000 });
      const actualMime = recorder.mimeType || mimeType;
      recorder.ondataavailable = event => {
        if (attempt !== generation || !event.data?.size) return;
        chunks.push(event.data); bytes += event.data.size;
        if (bytes >= maxBytes) stop('录音已达到大小上限，可以试听并保存。');
      };
      recorder.onerror = event => {
        emit({ error: recordingError(event.error) }); stop('录音已中断，已收到的内容可试听并保存。');
      };
      recorder.onstop = () => {
        if (attempt !== generation) return;
        accumulated = elapsed(); segmentStart = null; release();
        const blob = new Blob(chunks, { type: actualMime }); chunks = []; recorder = null;
        if (!blob.size) { emit({ status: 'error', error: state.error || '没有录到声音，请重新开始。' }); return; }
        const extension = actualMime.includes('mp4') ? 'm4a' : actualMime.includes('ogg') ? 'ogg' : 'webm';
        try {
          const file = new File([blob], `录音-${new Date(now()).toISOString().replace(/[:.]/g, '-')}.${extension}`, { type: actualMime });
          emit({ status: 'ready', file, url: createURL(file), elapsedMs: accumulated });
        } catch (error) { emit({ status: 'error', error: recordingError(error) }); }
      };
      stream.getTracks().forEach(track => { track.onended = () => stop('麦克风已断开，已录内容可试听并保存。'); });
      recorder.start(1000); startedAt = now(); segmentStart = startedAt;
      emit({ status: 'recording', canPause: typeof recorder.pause === 'function' && typeof recorder.resume === 'function' });
      timer = schedule(() => {
        emit({ elapsedMs: elapsed() });
        if (now() - startedAt >= maxDurationMs) stop('已达到 10 分钟录音上限，可以试听并保存。');
      }, 250);
    } catch (error) {
      if (attempt !== generation || disposed) return;
      release(); recorder = null; emit({ status: 'error', error: recordingError(error) });
    }
  }
  function pause() {
    if (state.status !== 'recording' || !state.canPause) return;
    try { recorder.pause(); accumulated = elapsed(); segmentStart = null; emit({ status: 'paused', elapsedMs: accumulated }); }
    catch (error) { emit({ error: recordingError(error) }); }
  }
  function resume() {
    if (state.status !== 'paused') return;
    try { recorder.resume(); segmentStart = now(); emit({ status: 'recording' }); }
    catch (error) { emit({ error: recordingError(error) }); }
  }
  function save(upload) {
    if (saving) return saving;
    if (state.status !== 'ready' || !state.file || typeof upload !== 'function') return Promise.resolve();
    const attempt = generation; const file = state.file;
    emit({ status: 'uploading', error: '' });
    saving = new Promise(resolve => resolve(upload(file))).then(
      result => {
        if (attempt === generation && !disposed) emit({ status: 'saved' });
        return result;
      },
      error => {
        if (attempt === generation && !disposed) emit({ status: 'ready', error: error?.message || '保存失败，请重试。' });
      },
    ).finally(() => { if (attempt === generation) saving = null; });
    return saving;
  }
  return { getState: () => state, start, pause, resume, stop, save, discard, dispose() { disposed = true; discard(); } };
}
