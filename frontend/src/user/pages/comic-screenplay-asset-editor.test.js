import test from 'node:test';
import assert from 'node:assert/strict';
import { assetEditorDraft, assetEditorPayload, voiceDraftTarget, voiceDesignAudio, validateVoiceDesign, comicMediaServerRequestId } from './comic-screenplay-asset-editor.js';
import { createHash } from 'node:crypto';

test('asset editor preserves extra metadata and explicit blank alias on save', () => {
  const original = { id: 'a', type: 'character', name: '阿明', revision: 4, extra: { alias: '小明', other: '保留' }, prompt: '外观' };
  const draft = assetEditorDraft(original, 'character', 'standard');
  assert.equal(draft.alias, '小明');
  const input = assetEditorPayload({ ...draft, alias: '', name: ' 阿明 ' });
  assert.equal(input.name, '阿明'); assert.equal(input.extra.alias, ''); assert.equal(input.extra.other, '保留'); assert.equal(input.expectedRevision, 4);
  assert.throws(() => assetEditorPayload({ name: '' }), /名称/);
});
test('voice previews bind to exact draft or project asset and require real voice metadata', () => {
  const draft = { id: 8, name: '阿明', asset_id: 'character-1' };
  assert.equal(voiceDraftTarget(draft), 'character-1');
  assert.equal(voiceDraftTarget({ id: 9 }), 'voice-draft:9');
  const real = { id: 'audio2', kind: 'audio', targetId: 'character-1', voiceId: 'real-id', voiceEngine: 'minimax', status: 'succeeded' };
  assert.equal(voiceDesignAudio([real, { ...real, id: 'normal-tts', voiceId: '', voiceEngine: '' }], draft), real);
  assert.equal(voiceDesignAudio([{ ...real, targetId: 'other' }], draft), undefined);
});
test('voice engine validation follows actual provider limits', () => {
  assert.match(validateVoiceDesign({ prompt: '温柔女声', preview_text: '你好' }, 'qwen-tts'), /15/);
  assert.equal(validateVoiceDesign({ prompt: '温柔女声', preview_text: '这是用于试听的完整文本，希望你能够听清我的声音。' }, 'qwen-tts'), '');
  assert.match(validateVoiceDesign({ prompt: 'a'.repeat(501), preview_text: 'a'.repeat(20) }, 'qwen-tts'), /500/);
  assert.match(validateVoiceDesign({ prompt: '温柔女声', preview_text: 'a'.repeat(501) }, 'minimax'), /500/);
  assert.match(validateVoiceDesign({ prompt: '温柔女声', preview_text: '文本' }, 'tts'), /引擎/);
});
test('retry lookup uses exact backend request hashes across analysis, voice design and images', async () => {
  const hash = value => createHash('sha256').update(value).digest('hex');
  assert.equal(await comicMediaServerRequestId('voice_analyze','key'), 'voice-analysis-'+hash('key'));
  assert.equal(await comicMediaServerRequestId('voice_design','key',{engine:'minimax',targetId:'voice-draft:8'}), 'sp-voice-'+hash('minimax:voice-draft:8:key'));
  assert.equal(await comicMediaServerRequestId('image','key'), 'sp-attempt-'+hash('image:key'));
});
