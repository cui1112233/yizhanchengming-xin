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
