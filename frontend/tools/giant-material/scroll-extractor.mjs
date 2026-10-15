import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { fileURLToPath } from 'node:url';
import { appendFrame } from './scroll-merge.mjs';

const failure = code => Object.assign(new Error(code), { code });

export function validateMaterial(material) {
  let url;
  try { url = new URL(material?.videoUrl); } catch { throw failure('OCR_VIDEO_NOT_ALLOWED'); }
  if (url.protocol !== 'https:' || url.hostname !== 'material.hnqingyuwen.top' || url.username || url.password || url.port) throw failure('OCR_VIDEO_NOT_ALLOWED');
  const durationSeconds = Number(material?.durationSeconds);
  if (!Number.isFinite(durationSeconds) || durationSeconds <= 0 || durationSeconds > 1800) throw failure('OCR_DURATION_NOT_SUPPORTED');
  return { videoUrl: url.href, durationSeconds };
}

export function cleanExtractedText(value = '') {
  return String(value ?? '').replace(/\r\n?/g, '\n').split('\n').map(line => line.trim()).filter(line => !line || /[\p{L}\p{N}\p{Script=Han}]/u.test(line)).join('\n');
}

export function filterFrameLines(frame) {
  return (frame.lines || []).filter(line => line.y >= .12 && line.y + line.height <= (frame.seconds === 0 ? .94 : .90) && !/^热门小说$|^本故事纯属虚构|^无不良引导|^危[险脸]动作请勿模仿|^点击下方链接$|^免费获取全文$/.test(line.text.trim())).map(line => line.text).filter(line => /[\p{L}\p{N}\p{Script=Han}]/u.test(String(line))).join('\n');
}

export async function collectScrollText(frames, { durationSeconds, signal, onProgress = () => {} }) {
  signal?.throwIfAborted();
  let body = '', previous = '', count = 0, duplicates = 0;
  const issues = [], frameReadFailures = [], emptyBodyFrames = [];
  const report = seconds => onProgress({ seconds, durationSeconds, frames: count, characters: body.replace(/\s/g, '').length, unaligned: issues.length, frameReadFailures: frameReadFailures.length, emptyBodyFrames: emptyBodyFrames.length });
  report(0);
  for await (const frame of frames) {
    signal?.throwIfAborted(); count++;
    if (frame.error) frameReadFailures.push(frame.seconds);
    else {
      const text = filterFrameLines(frame);
      if (text.trim()) {
        const result = appendFrame(body, previous, text);
        if (body && !result.boundary) issues.push({ seconds: frame.seconds, reason: 'NO_CONFIRMED_OVERLAP' });
        if (result.skipped) duplicates++;
        body = result.body; previous = text;
        if (Buffer.byteLength(body) > 2 * 1024 * 1024) throw failure('OCR_OUTPUT_TOO_LARGE');
      } else emptyBodyFrames.push(frame.seconds);
    }
    report(frame.seconds);
  }
  signal?.throwIfAborted(); body = cleanExtractedText(body);
  if (!body.trim()) throw failure('OCR_NO_TEXT');
  return { text: body, characters: body.replace(/\s/g, '').length, frames: count, duplicates, durationSeconds, sampleIntervalSeconds: 1, issues, frameReadFailures, emptyBodyFrames, requiresProofreading: true, sourceCompleteness: 'video_excerpt' };
}

async function* readNativeFrames(material, signal) {
  signal?.throwIfAborted();
  if (process.platform !== 'darwin') throw failure('OCR_PLATFORM_NOT_SUPPORTED');
  const env = Object.fromEntries(['PATH', 'TMPDIR', 'LANG', 'DEVELOPER_DIR'].filter(key => process.env[key]).map(key => [key, process.env[key]]));
  const child = spawn('/usr/bin/xcrun', ['swift', fileURLToPath(new URL('./scroll-video-ocr.swift', import.meta.url)), material.videoUrl, String(material.durationSeconds), '1'], { env, stdio: ['ignore', 'pipe', 'ignore'] });
  let spawnError;
  const ended = new Promise(resolve => { child.once('error', error => { spawnError = error; }); child.once('close', resolve); });
  let forceKill;
  const stop = () => { if (child.exitCode !== null || child.signalCode !== null) return; child.kill('SIGTERM'); forceKill ||= setTimeout(() => child.kill('SIGKILL'), 2000); };
  signal?.addEventListener('abort', stop, { once: true });
  const lines = createInterface({ input: child.stdout, crlfDelay: Infinity });
  try {
    for await (const raw of lines) { signal?.throwIfAborted(); if (raw.length > 200000) throw failure('OCR_OUTPUT_TOO_LARGE'); try { yield JSON.parse(raw); } catch (error) { if (error instanceof SyntaxError) throw failure('OCR_EXECUTION_FAILED'); throw error; } }
    const exitCode = await ended; signal?.throwIfAborted(); if (spawnError || exitCode !== 0) throw failure('OCR_EXECUTION_FAILED');
  } finally { lines.close(); stop(); await ended; clearTimeout(forceKill); signal?.removeEventListener('abort', stop); }
}

export async function extractScrollText(material, options = {}) {
  const checked = validateMaterial(material);
  return collectScrollText(readNativeFrames(checked, options.signal), { ...options, durationSeconds: checked.durationSeconds });
}
