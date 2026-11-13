import assert from 'node:assert/strict';
import test from 'node:test';
import { comicProductionView,comicScheduleRequest,comicProductionShotKey, buildComicTimeline, safeComicMediaUrl } from './comic-screenplay-workflow.js';
test('镜头生成结束与可交付成片明确区分',()=>{
  assert.equal(comicProductionView({status:'clips_ready'}).deliveryComplete,false);
  assert.equal(comicProductionView({status:'clips_ready'}).label,'镜头生成完成，待编辑与合成');
  assert.equal(comicProductionView({status:'running'}).active,true);
});
test('媒体地址拒绝脚本协议',()=>{
  assert.equal(safeComicMediaUrl('javascript:alert(1)'), '');
  assert.equal(safeComicMediaUrl('/api/ai-cut/projects/p/assets/a/content'),'/api/ai-cut/projects/p/assets/a/content');
});
test('时间线按顺序编排视频，音频保留自己的起始时间',()=>{
  const items=[{id:'v',kind:'video'},{id:'a',kind:'audio'}];
  assert.deepEqual(buildComicTimeline([{itemId:'a',durationMs:3000,startMs:2000,order:1},{itemId:'v',durationMs:4000,order:0}],items).map(item=>[item.itemId,item.startMs]),[['v',0],['a',2000]]);
  assert.throws(()=>buildComicTimeline([{itemId:'x',durationMs:1000}],items),/素材/);
  assert.throws(()=>buildComicTimeline([{itemId:'v',durationMs:0}],items),/时长/);
});
test('成片状态必须有第八步真实交付地址',()=>{
  assert.equal(comicProductionView({status:'completed'}).deliveryComplete,false);
  assert.equal(comicProductionView({status:'completed',batch:{steps:{'8_finalize':{status:'done',artifacts:['project://p/completed']}}}}).deliveryComplete,false);
  const complete=comicProductionView({status:'completed',batch:{steps:{'8_finalize':{status:'done',artifacts:['/api/ai-cut/export-jobs/export-1/output']}}}});
  assert.equal(complete.deliveryComplete,true);assert.deepEqual(complete.deliverables,['/api/ai-cut/export-jobs/export-1/output']);
  assert.equal(comicProductionView({status:'clips_ready',batch:{steps:{'8_finalize':{status:'done',artifacts:['https://media.example/clip.mp4']}}}}).deliveryComplete,false);
});
test('任务历史按请求版本区分同一镜头',()=>{
  const shot={project_id:'p',storyboard_id:'b',episode:1,seq:2,shot_id:'s'};
  assert.notEqual(comicProductionShotKey({...shot,request_hash:'old'}),comicProductionShotKey({...shot,request_hash:'new'}));
});
test('定时制作显示等待时间，未来时间使用本地输入转成明确时区',()=>{
 const now=Date.parse('2026-11-14T00:00:00Z');assert.throws(()=>comicScheduleRequest('2026-11-13T23:00:00Z',now),/未来/);assert.throws(()=>comicScheduleRequest('',now),/未来/);
 assert.deepEqual(comicScheduleRequest('2026-11-14T01:00:00Z',now),{run_at:'2026-11-14T01:00:00.000Z'});
 const view=comicProductionView({status:'scheduled',run_at:'2026-11-14T01:00:00Z'});assert.equal(view.scheduled,true);assert.equal(view.active,false);assert.equal(view.label,'等待定时启动');
});
