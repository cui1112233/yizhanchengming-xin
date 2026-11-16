import test from 'node:test';
import assert from 'node:assert/strict';
import {applyPlanningProgress} from './comic-screenplay-planning-progress.js';
import {loadStoryboardPlanProgress} from './comic-screenplay-storyboard-api.js';
const frame={type:'plan',chapter_id:'c',request_key:'once',sequence:3,round:2,total:8,chars:35,missing:['二'],status:'running'};
test('规划进度按章节与请求核对，拒绝迟到旧序号和错误快照',()=>{
 const result={chapter_id:'c',request_key:'once',status:'running',progress:frame};
 const current=applyPlanningProgress(null,result,'c','once');assert.equal(current.round,2);
 assert.equal(applyPlanningProgress(current,{...result,progress:{...frame,sequence:2,chars:1}},'c','once'),current);
 for(const wrong of [{...result,request_key:'other'},{...result,progress:{...frame,chapter_id:'other'}},{...result,progress:{...frame,round:99}}])assert.throws(()=>applyPlanningProgress(current,wrong,'c','once'));
 assert.equal(applyPlanningProgress(current,{...result,progress:null},'c','once'),current);
});
test('读取规划进度只发GET并保留取消信号',async()=>{
 const c=new AbortController();let call;
 await loadStoryboardPlanProgress('own-project','c','one key',c.signal,async(path,options)=>{call={path,options};return {};});
 assert.equal(call.path,'/api/comic-screenplays/projects/own-project/storyboard-plan-progress?chapter_id=c&request_key=one+key');
 assert.equal(call.options.method,undefined);assert.equal(call.options.body,undefined);assert.equal(call.options.signal,c.signal);
});
