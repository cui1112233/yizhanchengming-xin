import test from 'node:test';
import assert from 'node:assert/strict';
import { loadVoiceSwitchLibrary, readVoiceSwitchMedia, bindComicVoice, uploadComicVoiceReference } from './comicVoiceSwitch.js';
test('voice switch reads actual library and binds exact item/source with the frozen revision', async () => {
  const calls=[],request=async(path,options={})=>{calls.push({path,options});return {items:[]};};
  await loadVoiceSwitchLibrary(request);await readVoiceSwitchMedia('p',request);
  await bindComicVoice('p',{targetId:'voice-draft:1',itemId:'audio',sourceProjectId:'source',expectedRevision:7},request);
  assert.equal(calls[0].path,'/api/screenplays/voice-library');assert.equal(calls[1].path,'/api/comic-screenplays/projects/p/media');
  assert.deepEqual(JSON.parse(calls[2].options.body),{action:'voice_bind',targetId:'voice-draft:1',itemId:'audio',sourceProjectId:'source',expectedRevision:7});
});
test('reference file is sent once only by explicit upload, without generating voice ID', async () => {
  const file=new Blob(['RIFF'],{type:'audio/wav'});Object.defineProperty(file,'name',{value:'reference.wav'});let count=0;
  await uploadComicVoiceReference('p',file,'voice-draft:2',9,async(path,options)=>{count++;assert.equal(path,'/api/comic-screenplays/projects/p/media/voice-upload');assert.equal(options.body.get('expectedRevision'),'9');assert.equal(options.body.get('targetId'),'voice-draft:2');assert.equal(options.body.get('file').name,'reference.wav');assert.equal(options.body.get('voiceId'),null);return {};});
  assert.equal(count,1);
});
