import test from 'node:test';
import assert from 'node:assert/strict';
import { createBatchCompletionAudio } from './batchCompletionAudio.js';

test('completion sample is unlocked by interaction, cached and played at the selected volume',async()=>{
 const calls=[];let loads=0,contexts=0;
 class Context {
  constructor(){contexts++;this.state='suspended';this.currentTime=3;this.destination={};}
  resume(){this.state='running';calls.push('resume');return Promise.resolve();}
  decodeAudioData(){return Promise.resolve({sample:true});}
  createBufferSource(){return {connect:()=>{},disconnect:()=>{},start:()=>calls.push('start')};}
  createGain(){return {gain:{setValueAtTime:value=>calls.push(value)},connect:()=>{},disconnect:()=>{}};}
  close(){this.state='closed';return Promise.resolve();}
 }
 const audio=createBatchCompletionAudio({AudioContextClass:Context,load:async()=>{loads++;return{ok:true,arrayBuffer:async()=>new ArrayBuffer(4)}}});
 assert.equal(await audio.play({soundEnabled:true,soundVolume:60}),false);
 await audio.prepare();await audio.prepare();
 assert.equal(await audio.play({soundEnabled:false,soundVolume:60}),false);
 assert.equal(await audio.play({soundEnabled:true,soundVolume:0}),false);
 assert.equal(await audio.play({soundEnabled:true,soundVolume:25}),true);
 assert.equal(loads,1);assert.equal(contexts,1);assert.deepEqual(calls,['resume',.25,'start']);
 audio.dispose();
});
test('muting or changing account during sample decode prevents the old sound from starting',async()=>{
 for(const action of ['mute','account']){
  let decode,started=0,current=true,preferences={soundEnabled:true,soundVolume:60},ready;
  const decoding=new Promise(resolve=>{ready=resolve});
  class Context{
   constructor(){this.state='running';this.destination={};this.currentTime=0;}
   decodeAudioData(){return new Promise(r=>{decode=r;ready()});}
   createBufferSource(){return{connect(){},disconnect(){},start(){started++}};}
   createGain(){return{gain:{setValueAtTime(){}},connect(){},disconnect(){}};}
  }
  const audio=createBatchCompletionAudio({AudioContextClass:Context,load:async()=>({ok:true,arrayBuffer:async()=>new ArrayBuffer(1)})});
  const preparing=audio.prepare();await decoding;
  const playing=audio.play(preferences,{isCurrent:()=>current,getPreferences:()=>preferences});
  if(action==='mute')preferences={soundEnabled:false,soundVolume:60};else current=false;
  decode({sample:true});await preparing;assert.equal(await playing,false);assert.equal(started,0);
 }
});
