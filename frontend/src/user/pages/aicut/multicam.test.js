import test from 'node:test';
import assert from 'node:assert/strict';
const core = await import('./multicam.js').catch(() => ({}));
const assets = [ {id:'a',kind:'video',url:'/a',durationMs:10000,name:'A'}, {id:'b',kind:'video',url:'/b',durationMs:8000,name:'B'}, {id:'sound',kind:'audio',url:'/sound',durationMs:15000} ];
const opts = {cameras:[{assetId:'a',offsetMs:1000},{assetId:'b',offsetMs:-2000}],cuts:[{atMs:0,cameraIndex:0},{atMs:2000,cameraIndex:1},{atMs:4000,cameraIndex:0}],startMs:500,audio:{mode:'asset',assetId:'sound',sourceStartMs:100}};
test('aligns signed source offsets using only common available source range',()=>{
 assert.equal(typeof core.planMultiCamGroup,'function');
 const p=core.planMultiCamGroup(assets,opts);
 assert.equal(p.summary.commonStartMs,2000); assert.equal(p.summary.durationMs,7000);
 assert.deepEqual(p.tracks[0].clips.map(c=>[c.assetId,c.startMs,c.durationMs,c.sourceStartMs,c.volume]),[['a',500,2000,3000,0],['b',2500,2000,2000,0],['a',4500,3000,7000,0]]);
 assert.equal(p.tracks[1].kind,'audio'); assert.equal(p.tracks[1].clips.length,1); assert.equal(p.tracks[1].clips[0].sourceStartMs,100);
});
test('rejects ambiguous sources, missing cuts, bad timing, unavailable audio and camera count',()=>{
 assert.equal(typeof core.planMultiCamGroup,'function');
 const bad=[{cameras:[]},{cameras:[opts.cameras[0]]},{cameras:[opts.cameras[0],opts.cameras[0]]},{cuts:[{atMs:100,cameraIndex:0}]},{cuts:[{atMs:0,cameraIndex:3}]},{cuts:[{atMs:0,cameraIndex:0},{atMs:0,cameraIndex:1}]},{cuts:[{atMs:0,cameraIndex:0},{atMs:7000,cameraIndex:1}]},{cameras:[{assetId:'a',offsetMs:10000},opts.cameras[1]]},{audio:{mode:'asset',assetId:'sound',sourceStartMs:14999}},{startMs:NaN}];
 for(const change of bad) assert.throws(()=>core.planMultiCamGroup(assets,{...opts,...change}));
 assert.throws(()=>core.planMultiCamGroup([...assets,assets[0]],opts));
});
test('applies additive real clips with unique ids without mutating inputs and roundtrips JSON',()=>{
 assert.equal(typeof core.applyMultiCamGroup,'function');
 const t={revision:4,tracks:[{id:'old',kind:'video',order:8,clips:[{id:'id1',kind:'video',assetId:'a'}]}]};
 const before=structuredClone({t,assets,opts}); const p=core.planMultiCamGroup(assets,opts); let n=0;
 const result=core.applyMultiCamGroup(t,JSON.parse(JSON.stringify(p)),()=>`id${++n}`);
 assert.deepEqual({t,assets,opts},before); assert.deepEqual(result.timeline.tracks[0],t.tracks[0]);
 assert.deepEqual(result.timeline.tracks.slice(1).map(t=>t.order),[9,10]);
 const ids=result.timeline.tracks.flatMap(t=>[t.id,...t.clips.map(c=>c.id)]); assert.equal(new Set(ids).size,ids.length);
 assert.equal(result.selectionKey,`${result.timeline.tracks[1].id}::${result.timeline.tracks[1].clips[0].id}`);
 assert.deepEqual(JSON.parse(JSON.stringify(result)),result);
 assert.throws(()=>core.applyMultiCamGroup(t,p,()=> 'old'));
 const forged=structuredClone(p); forged.tracks[0].clips[0].durationMs=200000; assert.throws(()=>core.applyMultiCamGroup(t,forged,()=>`id${++n}`));
});

test('supports nine explicit cameras, rejects ten and unsupported camera audio mode',()=>{
 const sources=Array.from({length:10},(_,i)=>({id:`v${i}`,kind:'video',url:`/v${i}`,durationMs:5000+i}));
 const cameras=sources.slice(0,9).map((a,i)=>({assetId:a.id,offsetMs:i}));
 const p=core.planMultiCamGroup(sources,{cameras,cuts:[{atMs:0,cameraIndex:8}]});
 assert.equal(p.summary.cameraCount,9);assert.equal(p.summary.durationMs,5000);assert.equal(p.tracks[0].clips[0].sourceStartMs,8);assert.equal(p.tracks[0].clips[0].assetId,'v8');assert.equal(p.tracks.length,1);
 assert.throws(()=>core.planMultiCamGroup(sources,{cameras:[...cameras,{assetId:'v9'}],cuts:[{atMs:0,cameraIndex:0}]}));
 assert.throws(()=>core.planMultiCamGroup(assets,{...opts,audio:{mode:'camera',cameraIndex:0}}));
});
test('never clips beyond any source and does not add unsupported persistence metadata',()=>{
 for(let offset=-9000;offset<=7000;offset+=1000){
  const options={cameras:[{assetId:'a',offsetMs:offset},{assetId:'b',offsetMs:0}],cuts:[{atMs:0,cameraIndex:0}],audio:{mode:'none'}};
  let p;try{p=core.planMultiCamGroup(assets,options);}catch{continue;}
  const duration=p.summary.durationMs;
  if(duration>1) options.cuts.push({atMs:Math.floor(duration/2),cameraIndex:1});
  p=core.planMultiCamGroup(assets,options);
  for(const c of p.tracks[0].clips){const a=assets.find(a=>a.id===c.assetId);assert.ok(c.sourceStartMs>=0);assert.ok(c.sourceStartMs+c.durationMs<=a.durationMs);}
  let n=0;const result=core.applyMultiCamGroup({tracks:[]},p,()=>`new${++n}`);
  assert.deepEqual(Object.keys(result.timeline),['tracks']);
 }
});

test('alignment response must match selected identities and contain real finite confident offsets',()=>{
 assert.equal(typeof core.reviewMultiCamAlignment,'function');
 const cameras=[{assetId:'a',offsetMs:90},{assetId:'b',offsetMs:80}];const response={method:'audio-energy-cross-correlation',timeResolutionMs:20,referenceAssetId:'a',cameras:[{assetId:'b',offsetMs:-340,confidence:.98},{assetId:'a',offsetMs:0,confidence:1}]};
 const result=core.reviewMultiCamAlignment(cameras,response);assert.deepEqual(result.cameras,[{assetId:'a',offsetMs:0},{assetId:'b',offsetMs:-340}]);assert.equal(cameras[0].offsetMs,90);
 for(const change of [{cameras:[response.cameras[0]]},{referenceAssetId:'b'},{cameras:[response.cameras[0],response.cameras[0]]},{cameras:[response.cameras[1],{assetId:'b',offsetMs:NaN,confidence:.9}]},{cameras:[response.cameras[1],{assetId:'b',offsetMs:20,confidence:.1}]},{timeResolutionMs:0}]) assert.throws(()=>core.reviewMultiCamAlignment(cameras,{...response,...change}));
});

test('review summary describes source offsets without claiming whether analysis has run',()=>{
 const p=core.planMultiCamGroup(assets,opts);assert.equal(p.summary.alignment,'source-offset-review');assert.ok(p.summary.warnings.every(w=>!w.includes('尚未进行自动声音同步')));
});
