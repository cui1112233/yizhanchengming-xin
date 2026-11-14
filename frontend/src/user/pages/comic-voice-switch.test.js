import test from 'node:test';
import assert from 'node:assert/strict';
import { filterVoiceSwitchLibrary, projectVoiceLibrary, voiceBindingPreview, voiceEngineLabel, validateVoiceReferenceFile } from './comic-voice-switch.js';
test('voice library search is trimmed and case-insensitive, matching names/tags before 40 limit', () => {
  const rows=Array.from({length:55},(_,i)=>({name:`Voice ${i}`,tags:['MiniMax']}));
  assert.equal(filterVoiceSwitchLibrary(rows,' MINIMAX ').length,40);
  assert.equal(filterVoiceSwitchLibrary(rows,'voice 54')[0].name,'Voice 54');
});
test('shared project library contains only actual designed voices from that project', () => {
  const rows=projectVoiceLibrary([{id:'ref',kind:'audio',status:'succeeded',url:'/api/ref'},{id:'v',kind:'audio',status:'succeeded',voiceId:'real',voiceEngine:'qwen-tts',url:'/api/v',name:'真实音色'}],'p','工程');
  assert.equal(rows.length,1);assert.equal(rows[0].media_item_id,'v');assert.equal(rows[0].voice_id,'real');assert.equal(rows[0].project_id,'p');
});
test('reference binding is displayed as reference audio without a manufactured voice ID', () => {
  const preview=voiceBindingPreview({settings:{voiceBindings:{'voice-draft:1':{media_item_id:'ref',engine:'custom',reference_only:true,voice_id:''}}},items:[{id:'ref',kind:'audio',url:'/api/ref'}]},'voice-draft:1');
  assert.equal(preview.item.url,'/api/ref');assert.equal(preview.label,'参考音频');assert.equal(preview.voiceId,'');assert.equal(voiceEngineLabel('minimax'),'MiniMax');assert.equal(voiceEngineLabel('other'),'other');
});
test('reference uploads reject other formats, empty files and more than 32MB', () => {
  assert.equal(validateVoiceReferenceFile({name:'sample.WAV',size:44}), '');
  assert.match(validateVoiceReferenceFile({name:'sample.m4a',size:44}),/WAV/);
  assert.match(validateVoiceReferenceFile({name:'sample.mp3',size:0}),/为空/);
  assert.match(validateVoiceReferenceFile({name:'sample.mp3',size:32*1024*1024+1}),/32MB/);
});
