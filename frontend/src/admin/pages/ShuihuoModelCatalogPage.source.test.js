import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./ShuihuoModelCatalogPage.jsx', import.meta.url), 'utf8');

test('model catalog offers safe disable and enable actions instead of only destructive removal', () => {
  assert.match(source, /setAdminModelEnabled/);
  assert.match(source, /enabled \? '停用' : '启用'/);
  assert.match(source, /确认停用这个模型/);
});

test('generic audio model can configure a server-side voice clone adapter', () => {
  assert.match(source, /supportsVoiceClone/);
  assert.match(source, /voiceCloneRequestTemplate/);
  assert.match(source, /voiceCloneResponseMapping/);
  assert.match(source, /reference_audio_base64/);
  assert.match(source, /参考音频只由后端读取/);
});
