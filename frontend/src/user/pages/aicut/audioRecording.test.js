import test from 'node:test';
import assert from 'node:assert/strict';
import { createAudioRecording, chooseAudioMimeType } from './audioRecording.js';

function setup(options = {}) {
  let recorder; let tick; let now = 0; let stops = 0; const revoked = [];
  class Recorder {
    static isTypeSupported(type) { return type === 'audio/webm;codecs=opus'; }
    constructor(stream, config) { this.mimeType = config.mimeType; this.state = 'inactive'; recorder = this; }
    start() { this.state = 'recording'; }
    pause() { this.state = 'paused'; }
    resume() { this.state = 'recording'; }
    stop() { this.state = 'inactive'; }
    finish(text = 'voice') { this.ondataavailable?.({ data: new Blob([text]) }); this.onstop?.(); }
  }
  const controller = createAudioRecording({ MediaRecorder: Recorder, getUserMedia: async () => ({ getTracks: () => [{ stop() { stops++; } }] }),
    createObjectURL: () => 'blob:voice', revokeObjectURL: url => revoked.push(url), now: () => now,
    setInterval: callback => { tick = callback; return 1; }, clearInterval: () => { tick = null; }, ...options });
  return { controller, recorder: () => recorder, advance(ms) { now += ms; tick?.(); }, stops: () => stops, revoked };
}

test('recording starts only on explicit start; pause excludes elapsed time and stop creates playable file', async () => {
  const s = setup(); assert.equal(s.recorder(), undefined);
  await s.controller.start(); s.advance(1000); s.controller.pause(); s.advance(2000);
  assert.equal(s.controller.getState().elapsedMs, 1000);
  s.controller.resume(); s.advance(500); s.controller.stop(); s.recorder().finish();
  const state = s.controller.getState();
  assert.equal(state.status, 'ready'); assert.equal(state.elapsedMs, 1500);
  assert.equal(state.file.type, 'audio/webm;codecs=opus'); assert.match(state.file.name, /\.webm$/);
  assert.equal(s.stops(), 1); assert.equal(state.url, 'blob:voice');
});

test('permission denial gives a useful message without starting recording', async () => {
  const s = setup({ getUserMedia: async () => { throw Object.assign(new Error(), { name: 'NotAllowedError' }); } });
  await s.controller.start(); assert.equal(s.controller.getState().status, 'error');
  assert.match(s.controller.getState().error, /麦克风权限/); assert.equal(s.recorder(), undefined);
});

test('discard and dispose release microphone, detach late events, clear timer and revoke preview', async () => {
  const s = setup(); await s.controller.start(); const recorder = s.recorder();
  s.controller.discard(); recorder.finish(); assert.equal(s.controller.getState().status, 'idle');
  assert.equal(s.stops(), 1); assert.equal(recorder.onstop, null);
  await s.controller.start(); s.controller.stop(); s.recorder().finish(); s.controller.dispose();
  assert.deepEqual(s.revoked, ['blob:voice']); assert.equal(s.stops(), 2);
});

test('cancel while permission is pending closes the late stream', async () => {
  let resolve; let stops = 0;
  const s = setup({ getUserMedia: () => new Promise(done => { resolve = done; }) });
  const pending = s.controller.start(); s.controller.discard();
  resolve({ getTracks: () => [{ stop() { stops++; } }] }); await pending;
  assert.equal(stops, 1); assert.equal(s.recorder(), undefined);
});

test('save deduplicates clicks and keeps a failed upload available for retry', async () => {
  const s = setup(); await s.controller.start(); s.controller.stop(); s.recorder().finish();
  let finish; let calls = 0; const upload = () => { calls++; return new Promise(resolve => { finish = resolve; }); };
  const first = s.controller.save(upload); const duplicate = s.controller.save(upload);
  assert.equal(calls, 1); finish({ id: 1 }); await Promise.all([first, duplicate]);
  assert.equal(s.controller.getState().status, 'saved'); await s.controller.save(upload); assert.equal(calls, 1);
  s.controller.discard(); await s.controller.start(); s.controller.stop(); s.recorder().finish();
  await s.controller.save(async () => { throw new Error('上传失败'); });
  assert.equal(s.controller.getState().status, 'ready'); assert.match(s.controller.getState().error, /上传失败/);
  await s.controller.save(() => { throw new Error('同步保存失败'); });
  await s.controller.save(async () => ({ id: 2 }));
  assert.equal(s.controller.getState().status, 'saved');
});

test('duration and size limits stop gracefully and preserve a usable recording', async () => {
  const s = setup({ maxDurationMs: 1000, maxBytes: 4 }); await s.controller.start(); s.advance(1000);
  assert.equal(s.controller.getState().status, 'stopping'); s.recorder().finish(); assert.equal(s.controller.getState().status, 'ready');
  s.controller.discard(); await s.controller.start(); s.recorder().ondataavailable({ data: new Blob(['12345']) });
  assert.equal(s.controller.getState().status, 'stopping'); s.recorder().onstop();
  assert.equal(s.controller.getState().file.size, 5);
});

test('unsupported codecs and missing input are explained', async () => {
  assert.equal(chooseAudioMimeType({ isTypeSupported: () => false }), '');
  const s = setup({ getUserMedia: async () => { throw Object.assign(new Error(), { name: 'NotFoundError' }); } });
  await s.controller.start(); assert.match(s.controller.getState().error, /麦克风/);
});
