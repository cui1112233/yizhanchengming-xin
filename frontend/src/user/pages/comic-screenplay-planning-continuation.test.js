import test from 'node:test';
import assert from 'node:assert/strict';
import {generateStoryboardPlan} from './comic-screenplay-storyboard-api.js';

test('多轮规划保留取消信号，不被单轮十分钟计时器中断',async()=>{
 const controller=new AbortController();let captured;
 await generateStoryboardPlan('own-project',{chapter_id:'c',idempotencyKey:'once'},controller.signal,async(path,options)=>{captured={path,options};return {chapter:{planning_rounds:8}};});
 assert.equal(captured.path,'/api/comic-screenplays/projects/own-project/storyboard-plan');
 assert.equal(captured.options.timeoutMs,0);
 assert.equal(captured.options.signal,controller.signal);
 assert.deepEqual(JSON.parse(captured.options.body),{chapter_id:'c',idempotencyKey:'once'});
 controller.abort();assert.equal(captured.options.signal.aborted,true);
});

test('续写规划透传 is_continue 与已有规划文本',async()=>{
 const controller=new AbortController();let captured;
 await generateStoryboardPlan('own-project',{chapter_id:'c',idempotencyKey:'k2',is_continue:true,existing_plan:'# 已保存规划\n第1章…'},controller.signal,async(path,options)=>{captured=JSON.parse(options.body);return {chapter:{planning_rounds:1}};});
 assert.equal(captured.is_continue,true);
 assert.equal(captured.existing_plan,'# 已保存规划\n第1章…');
 assert.equal(captured.idempotencyKey,'k2');
});
