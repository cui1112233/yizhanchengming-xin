import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./ApiConfigPage.jsx', import.meta.url), 'utf8');

test('API configuration exposes a dedicated audio/TTS model form', () => {
  assert.match(source, /key: 'audio'/);
  assert.match(source, /requestTemplate/);
  assert.match(source, /responseMapping/);
  assert.match(source, /音频模型/);
});
