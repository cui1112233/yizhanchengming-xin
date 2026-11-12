import test from 'node:test';
import assert from 'node:assert/strict';
const api=await import('./aicut.js');
test('multi-camera sync API sends bounded explicit ids, carries cancellation outside JSON and returns reviewable response',async()=>{
 assert.equal(typeof api.alignAICutMultiCamAudio,'function');const signal=new AbortController().signal;let call;const result={cameras:[{assetId:'a',offsetMs:0,confidence:1}]};
 assert.equal(await api.alignAICutMultiCamAudio('p1',{assetIds:['a','b'],referenceAssetId:'a',signal},async(path,options)=>{call={path,options};return result}),result);
 assert.equal(call.path,'/api/ai-cut/projects/p1/multicam-sync');assert.equal(call.options.signal,signal);assert.equal(call.options.timeoutMs,100000);assert.equal(call.options.method,'POST');assert.deepEqual(JSON.parse(call.options.body),{assetIds:['a','b'],referenceAssetId:'a',windowMs:30000,maxLagMs:5000});
});
test('multi-camera sync API rejects bad counts, duplicate or missing reference and bounds before network',async()=>{
 assert.equal(typeof api.alignAICutMultiCamAudio,'function');let calls=0;const request=async()=>{calls++};for(const input of [{assetIds:['a'],referenceAssetId:'a'},{assetIds:['a','a'],referenceAssetId:'a'},{assetIds:['a','b'],referenceAssetId:'c'},{assetIds:['a','b'],referenceAssetId:'a',windowMs:60001},{assetIds:['a','b'],referenceAssetId:'a',maxLagMs:NaN}]) await assert.rejects(()=>api.alignAICutMultiCamAudio('p1',input,request));assert.equal(calls,0);
});
