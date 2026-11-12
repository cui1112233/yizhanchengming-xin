import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

test('recorder exposes explicit microphone action, native preview, upload and resource cleanup', () => {
  const source = readFileSync(new URL('./AudioRecorder.jsx', import.meta.url), 'utf8');
  assert.match(source, /开始录音/);
  assert.match(source, /<audio[^>]*controls/);
  assert.match(source, /\.save\(onUpload\)/);
  assert.match(source, /\.dispose\(\)/);
  assert.match(source, /停止录音/);
});
