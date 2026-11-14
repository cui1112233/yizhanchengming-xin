import test from 'node:test';
import assert from 'node:assert/strict';
import {listComicAssetChecks,getComicAssetCheck,startComicAssetCheck,cancelComicAssetCheck,applyComicAssetCheck,discardComicAssetCheck} from './comicScreenplayAssetChecks.js';

test('后台检查只提交明确来源及集号；读取、取消、确认使用同一任务',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return {job:{id:'check_1',revision:4}};};
 await listComicAssetChecks('p',request);await startComicAssetCheck('p',{requestKey:'same-key',sourceMode:'original',episodeNumbers:[3,1,3],model:'own-text',confirmed:true,content:'不得上传的旧前端稿'},request);
 await getComicAssetCheck('p','check_1',request);await cancelComicAssetCheck('p','check_1',request);await applyComicAssetCheck('p','check_1',4,request);await discardComicAssetCheck('p','check_1',4,request);
 assert.deepEqual(JSON.parse(calls[1].options.body),{requestKey:'same-key',sourceMode:'original',episodeNumbers:[1,3],model:'own-text',confirmed:true});
 assert.equal(calls[2].path,'/api/comic-screenplays/projects/p/asset-checks/check_1');assert.equal(calls[3].path,calls[2].path+'/cancel');assert.deepEqual(JSON.parse(calls[3].options.body),{});assert.deepEqual(JSON.parse(calls[4].options.body),{expectedRevision:4});
 assert.equal(calls[5].path,calls[2].path+'/discard');assert.deepEqual(JSON.parse(calls[5].options.body),{expectedRevision:4});
 assert.ok(calls.every(c=>c.options.timeoutMs===15000));assert.ok(calls.every(c=>!c.options.signal));
});

test('空输入、错来源、非法集号、版本和路径在网络前拒绝',()=>{
 let count=0;const request=()=>{count++;};const valid={requestKey:'key',sourceMode:'script',episodeNumbers:[1],model:'own',confirmed:true};
 for(const patch of [{confirmed:false},{requestKey:''},{sourceMode:'fake'},{episodeNumbers:[]},{episodeNumbers:[0]},{episodeNumbers:[1.5]},{model:''}])assert.throws(()=>startComicAssetCheck('p',{...valid,...patch},request));
 assert.throws(()=>getComicAssetCheck('p','../x',request));assert.throws(()=>listComicAssetChecks('../p',request));assert.throws(()=>applyComicAssetCheck('p','job',NaN,request));assert.equal(count,0);
});
