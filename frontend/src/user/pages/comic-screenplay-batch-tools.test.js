import test from 'node:test';
import assert from 'node:assert/strict';
import {createEpisodeRunner,validateEpisodeBatch,storyboardAttemptRows} from './comic-screenplay-batch-tools.js';
test('rejects empty selection, model and missing chapter before requests',()=>{
 assert.throws(()=>validateEpisodeBatch({selected:[],modelId:'m'}),/选择/);
 assert.throws(()=>validateEpisodeBatch({selected:[{no:1}],modelId:''}),/模型/);
 assert.throws(()=>validateEpisodeBatch({selected:[{no:1}],modelId:'m',mode:'new',bindings:{},chapters:[]}),/章节/);
});
test('runner serializes, blocks duplicate clicks, preserves partial results and stops after current',async()=>{
 const runner=createEpisodeRunner();let release,calls=[];
 const first=runner.run([1,2,3],async no=>{calls.push(no);await new Promise(r=>release=r);return no;});
 await assert.rejects(runner.run([9],async()=>{}),/执行中/);runner.stop();release();
 const result=await first;assert.deepEqual(calls,[1]);assert.equal(result[0].status,'succeeded');assert.equal(result[1].status,'not_started');
 const next=await runner.run([1,2,3],async n=>{if(n===2)throw Error('上游失败');return n});
 assert.deepEqual(next.map(x=>x.status),['succeeded','failed','succeeded']);assert.equal(next[1].error,'上游失败');
});
test('failure list uses newest durable episode status and never guesses failure from empty boards',()=>{
 const rows=storyboardAttemptRows([{request_key:'old',engine:'e',created_at:'2026-01-01',episodes:[{episode_no:1,status:'failed',error:'old'},{episode_no:2,status:'failed'}]},{request_key:'new',engine:'e',created_at:'2026-01-02',episodes:[{episode_no:1,status:'succeeded'},{episode_no:3,status:'unknown'}]}],'e');
 assert.deepEqual(rows.filter(x=>x.failed).map(x=>x.episodeNo),[2]);assert.equal(rows.find(x=>x.episodeNo===3).status,'unknown');assert.equal(rows.length,3);
});
