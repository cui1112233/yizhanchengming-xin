import test from 'node:test';
import assert from 'node:assert/strict';
import { createBatchCompletionObserver, createBatchCompletionMonitor } from './batchCompletion.js';

const active={id:'j1',batchId:'b1',state:'running',success:false};
const done={...active,state:'completed',success:true,total:3,completionKey:'j1:finished-1'};
test('history is silent; only all-success after observed work notifies once',()=>{
 const observer=createBatchCompletionObserver();
 assert.deepEqual(observer.observe([done]),[]);
 assert.deepEqual(observer.observe([active]),[]);
 assert.deepEqual(observer.observe([{...done,success:false}]),[]);
 assert.deepEqual(observer.observe([done]),[done]);
 assert.deepEqual(observer.observe([done]),[]);
 observer.observe([active]);assert.equal(observer.observe([{...done,completionKey:'j1:finished-2'}]).length,1);
});
test('refresh preserves known active jobs and failed/stopped work never announces success',()=>{
 let armed=[];const observer=createBatchCompletionObserver({save:value=>{armed=value}});
 observer.observe([active]);const resumed=createBatchCompletionObserver({initial:armed});
 assert.equal(resumed.observe([done]).length,1);
 for(const state of ['failed','stopped','paused','needs_attention']){
  const next=createBatchCompletionObserver();next.observe([active]);assert.deepEqual(next.observe([{...done,state,success:false}]),[]);
 }
});

function memoryStorage(){const data=new Map();return{getItem:key=>data.get(key)??null,setItem:(key,value)=>data.set(key,value)}}
test('global polling continues across routes, honors preferences and deduplicates two monitors',async()=>{
 const storage=memoryStorage(),sessionStorage=memoryStorage();let jobs=[active];const notices=[],plays=[];
 const options={getIdentity:()=>({owner:'alice',session:'session-a'}),fetchJobs:async()=>jobs,fetchPreferences:async()=>({soundEnabled:true,soundVolume:25}),storage,sessionStorage,notify:n=>notices.push(n),play:p=>{plays.push(p);return true;}};
 const first=createBatchCompletionMonitor(options),second=createBatchCompletionMonitor(options);
 await first.poll();await second.poll();jobs=[done];await first.poll();await second.poll();await first.poll();
 assert.equal(notices.length,1);assert.deepEqual(plays,[{soundEnabled:true,soundVolume:25}]);
 const refreshed=createBatchCompletionMonitor(options);await refreshed.poll();assert.equal(plays.length,1);
});
test('mute still records completion without losing volume preference',async()=>{
 let jobs=[active];const played=[];
 const m=createBatchCompletionMonitor({getIdentity:()=>({owner:'alice',session:'a'}),fetchJobs:async()=>jobs,fetchPreferences:async()=>({soundEnabled:false,soundVolume:40}),storage:memoryStorage(),sessionStorage:memoryStorage(),notify:()=>{},play:p=>{played.push(p);return true;}});
 await m.poll();jobs=[done];await m.poll();assert.deepEqual(played,[]);
});
test('a response from an old login cannot announce or play on the new account',async()=>{
 let identity={owner:'alice',session:'a'},jobs=[active],resolve;
 const played=[];const m=createBatchCompletionMonitor({getIdentity:()=>identity,fetchJobs:()=>new Promise(r=>{resolve=r}),fetchPreferences:async()=>({}),storage:memoryStorage(),sessionStorage:memoryStorage(),notify:()=>played.push('notice'),play:()=>played.push('sound')});
 const poll=m.poll();identity={owner:'bob',session:'b'};resolve([done]);await poll;assert.deepEqual(played,[]);
});
test('temporary status errors retain the armed job and do not invent completion',async()=>{
 let jobs=[active],error=false;const played=[];
 const m=createBatchCompletionMonitor({getIdentity:()=>({owner:'alice',session:'a'}),fetchJobs:async()=>{if(error)throw Error('offline');return jobs},fetchPreferences:async()=>({soundEnabled:true,soundVolume:60}),storage:memoryStorage(),sessionStorage:memoryStorage(),notify:()=>{},play:p=>{played.push(p);return true;}});
 await m.poll();error=true;await m.poll();assert.equal(played.length,0);error=false;jobs=[done];await m.poll();assert.equal(played.length,1);
});
test('new mute settings win over an older in-flight preferences request',async()=>{
 let resolve;const played=[];let jobs=[active];
 const m=createBatchCompletionMonitor({getIdentity:()=>({owner:'alice',session:'a'}),fetchJobs:async()=>jobs,fetchPreferences:()=>new Promise(r=>{resolve=r}),storage:memoryStorage(),sessionStorage:memoryStorage(),notify:()=>{},play:p=>{played.push(p);return true;}});
 const polling=m.poll();await Promise.resolve();m.setPreferences({soundEnabled:false,soundVolume:15});resolve({soundEnabled:true,soundVolume:90});await polling;
 jobs=[done];await m.poll();assert.deepEqual(played,[]);
});
test('an audio-blocked tab cannot silence a second unlocked tab',async()=>{
 const storage=memoryStorage();let jobs=[active],sounded=0,notices=0;
 const options={getIdentity:()=>({owner:'alice',session:'a'}),fetchJobs:async()=>jobs,fetchPreferences:async()=>({soundEnabled:true,soundVolume:60}),storage,notify:()=>notices++};
 const blocked=createBatchCompletionMonitor({...options,sessionStorage:memoryStorage(),play:async()=>false});
 const ready=createBatchCompletionMonitor({...options,sessionStorage:memoryStorage(),play:async()=>{sounded++;return true;}});
 await blocked.poll();await ready.poll();jobs=[done];await blocked.poll();await ready.poll();await blocked.poll();
 assert.equal(sounded,1);assert.equal(notices,1);
});
test('completion pending during a config error survives full-page navigation',async()=>{
 const storage=memoryStorage(),sessionStorage=memoryStorage(),played=[];let jobs=[active],fail=true;
 const options={getIdentity:()=>({owner:'alice',session:'a'}),fetchJobs:async()=>jobs,fetchPreferences:async()=>{if(fail)throw Error('config offline');return{soundEnabled:true,soundVolume:60}},storage,sessionStorage,notify:()=>{},play:p=>{played.push(p);return true;}};
 const first=createBatchCompletionMonitor(options);await first.poll();jobs=[done];await first.poll();assert.equal(played.length,0);first.dispose();fail=false;
 const next=createBatchCompletionMonitor(options);await next.poll();assert.equal(played.length,1);await next.poll();assert.equal(played.length,1);
});
