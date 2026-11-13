import assert from 'node:assert/strict';
import test from 'node:test';
import { getComicProduction, scheduleComicProduction,cancelComicProductionSchedule,rebuildComicProduction,startComicProduction, pauseComicProduction, getComicChapters, saveComicChapters, sendComicAssistant, applyComicProposal, batchEditComicShots, comicWorkflowPath, downloadComicMediaPackage, generateComicChapter } from './comicScreenplayWorkflow.js';

test('后台生产开始、查询与暂停使用持久任务接口，页面不推进run-next', async () => {
  const calls=[]; const request=async (path,options={})=>{calls.push([path,options.method||'GET']);return {batch:{status:'running'}};};
  await startComicProduction('comic-1',request); await getComicProduction('comic-1',request); await pauseComicProduction('comic-1',request);
  assert.deepEqual(calls,[['/api/comic-screenplays/projects/comic-1/production/start','POST'],['/api/comic-screenplays/projects/comic-1/production/status','GET'],['/api/comic-screenplays/projects/comic-1/production/pause','POST']]);
});
test('章节保存携带修订号并保留原文字段',async()=>{
  let body; const request=async(path,options)=>{if(!options)return {state:{revision:2,chapters:[]}};body=JSON.parse(options.body);return {state:{revision:3}};};
  assert.equal((await getComicChapters('comic-1',request)).state.revision,2);
  await saveComicChapters('comic-1',{expected_revision:2,chapters:[{id:'c1',title:'第一章',source_text:'\n完整原文。\n',script_text:''}]},request);
  assert.equal(body.expected_revision,2); assert.equal(body.chapters[0].source_text,'\n完整原文。\n');
});
test('助手先对话，应用提案必须显式确认并携带修订号',async()=>{
  const calls=[];const request=async(path,options)=>{calls.push([path,JSON.parse(options.body)]);return {};};
  await sendComicAssistant('comic-1',{message:'整理第一章',expected_revision:4},request);
  await applyComicProposal('comic-1',{expected_revision:5,proposal_id:'p1',confirmed:true},request);
  assert.equal(calls[0][0],'/api/comic-screenplays/projects/comic-1/assistant/chat');assert.equal(calls[1][1].confirmed,true);
  await assert.rejects(()=>applyComicProposal('comic-1',{proposal_id:'p1'},request),/确认/);
});
test('批量分镜操作限定具体引擎与集数，拒绝无效路径',async()=>{
  let path;await batchEditComicShots('comic-1','xiaoluo_v7_15s',2,{operation:'duration',duration:4,shot_ids:['s1'],expected_revision:1},async value=>{path=value;return {};});
  assert.equal(path,'/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/2/shots/batch');
  assert.throws(()=>comicWorkflowPath('../owner','chapters'),/项目/);
});
test('重建变更请求必须明确确认并向服务端传递rebuild',async()=>{
  const calls=[];const request=async(path,options)=>{calls.push([path,JSON.parse(options.body)]);return {};};
  await assert.rejects(()=>rebuildComicProduction('comic-1',{},request),/确认/);assert.equal(calls.length,0);
  await rebuildComicProduction('comic-1',{confirmed:true},request);assert.deepEqual(calls,[['/api/comic-screenplays/projects/comic-1/production/start',{rebuild:true}]]);
});

test('资料包通过登录请求取Blob，章节允许已持久保存的中文编号',async()=>{const calls=[];const request=async(path,options)=>{calls.push([path,options]);return {};};await downloadComicMediaPackage('comic-1','jianying',request);assert.deepEqual(calls[0],['/api/comic-screenplays/projects/comic-1/media/package?format=jianying',{responseType:'blob'}]);await generateComicChapter('comic-1','第一章',{expected_revision:3},request);assert.equal(calls[1][0],'/api/comic-screenplays/projects/comic-1/chapters/%E7%AC%AC%E4%B8%80%E7%AB%A0/script');assert.equal(calls[1][1].timeoutMs,600000);});
test('定时制作必须选择未来时间，并明确发出定时或取消请求',async()=>{
 const calls=[];const request=async(path,options)=>{calls.push([path,JSON.parse(options.body)]);return {};};const at=new Date(Date.now()+3600000).toISOString();
 await assert.rejects(()=>scheduleComicProduction('comic-1',{run_at:at},request),/确认/);
 await assert.rejects(()=>scheduleComicProduction('comic-1',{run_at:'2000-01-01T00:00:00Z',confirmed:true},request),/未来/);
 await scheduleComicProduction('comic-1',{run_at:at,confirmed:true},request);await cancelComicProductionSchedule('comic-1',request);
 assert.deepEqual(calls,[['/api/comic-screenplays/projects/comic-1/production/schedule',{run_at:at}],['/api/comic-screenplays/projects/comic-1/production/cancel-schedule',{}]]);
});

test('媒体工程包导入带时间线版本，成片通过登录请求下载',async()=>{const {importComicMediaPackage,listComicMediaVoices,downloadComicMediaResult}=await import('./comicScreenplayWorkflow.js');const calls=[];const request=async(path,options)=>{calls.push([path,options]);return {};};await importComicMediaPackage('comic-1',new Blob(['zip']),7,request);assert.equal(calls[0][0],'/api/comic-screenplays/projects/comic-1/media/import');assert.equal(calls[0][1].body.get('expectedRevision'),'7');await listComicMediaVoices('comic-1',request);assert.equal(calls[1][0],'/api/comic-screenplays/projects/comic-1/media/voices');await downloadComicMediaResult('comic-1','export-1',request);assert.deepEqual(calls[2],['/api/comic-screenplays/projects/comic-1/media/exports/export-1',{responseType:'blob'}]);});
