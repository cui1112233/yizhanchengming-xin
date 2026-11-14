import test from 'node:test';import assert from 'node:assert/strict';
import {prepareStoryboardGenerationRequest,completeStoryboardGenerationRequest,reconcileStoryboardGenerationRequest,markStoryboardGenerationNotStarted} from './comic-screenplay-storyboard-attempts.js';
const empty=async()=>({attempts:[]});

test('公共入口未确认提交重复点击保同key，只有成功后的明确新操作才换key，false选项保留',async()=>{
 let keys=0;const context={projectId:'stable',engine:'engine',episodeNumbers:[1],options:{usePlanning:false,content:'完整正文'},loadAttempts:empty,newKey:()=>`key_${++keys}`};
 const first=await prepareStoryboardGenerationRequest({...context,explicitNewOperation:true});const repeated=await prepareStoryboardGenerationRequest({...context,explicitNewOperation:true});assert.equal(first.requestOptions.idempotencyKey,repeated.requestOptions.idempotencyKey);assert.equal(keys,1);assert.equal(first.requestOptions.usePlanning,false);
 completeStoryboardGenerationRequest(first.cacheKey);const replay=await prepareStoryboardGenerationRequest(context);assert.equal(replay.requestOptions.idempotencyKey,first.requestOptions.idempotencyKey);const next=await prepareStoryboardGenerationRequest({...context,explicitNewOperation:true});assert.notEqual(next.requestOptions.idempotencyKey,first.requestOptions.idempotencyKey);assert.equal(keys,2);
});

test('server running/unknown跨选项也禁止新key，查询失败不自动建付费请求',async()=>{
 for(const status of ['running','unknown']){const result=await prepareStoryboardGenerationRequest({projectId:`blocked_${status}`,engine:'e',episodeNumbers:[1],options:{usePlanning:false},explicitNewOperation:true,loadAttempts:async()=>({attempts:[{request_key:'old_uuid',engine:'e',episode_numbers:[1,2],status}]}),newKey:()=>assert.fail('不可新key')});assert.equal(result.allowed,false);}
 const offline=await prepareStoryboardGenerationRequest({projectId:'offline',engine:'e',episodeNumbers:[1],loadAttempts:async()=>{throw Error('断网');},newKey:()=>assert.fail('不可新key')});assert.equal(offline.allowed,false);
});

test('partial/failed/cancelled仅明确确认can_retry后创建新key及retryOf，完整集集合不缩小',async()=>{
 for(const status of ['partial','failed','cancelled']){const attempt={request_key:'original_uuid',engine:'e',episode_numbers:[1,2],status,can_retry:true};const context={projectId:`retry_${status}`,engine:'e',episodeNumbers:[2,1],options:{usePlanning:true,planningChapterId:'c'},loadAttempts:async()=>({attempts:[attempt]}),newKey:()=>`new_${status}`};const cancelled=await prepareStoryboardGenerationRequest({...context,confirmRetry:async()=>false});assert.equal(cancelled.allowed,false);const resumed=await prepareStoryboardGenerationRequest({...context,confirmRetry:async current=>{assert.equal(current.request_key,'original_uuid');return true;}});assert.equal(resumed.requestOptions.retryOf,'original_uuid');assert.equal(resumed.requestOptions.confirmedRetry,true);assert.equal(resumed.requestOptions.idempotencyKey,`new_${status}`);}
});

test('终态成功无明确重新生成重放原uuid，reconcile只GET并同步缓存状态',async()=>{
 const context={projectId:'replayed',engine:'e',episodeNumbers:[1],loadAttempts:async()=>({attempts:[{request_key:'raw_uuid',engine:'e',episode_numbers:[1],status:'succeeded'}]})};const replay=await prepareStoryboardGenerationRequest(context);assert.equal(replay.requestOptions.idempotencyKey,'raw_uuid');
 let reads=0;const result=await reconcileStoryboardGenerationRequest({projectId:'replayed',cacheKey:replay.cacheKey,idempotencyKey:'raw_uuid',loadAttempt:async(id,key)=>{reads++;assert.equal(key,'raw_uuid');return {attempt:{request_key:key,status:'succeeded'}};}});assert.equal(result.status,'succeeded');assert.equal(reads,1);
});


test('成功回包丢失且页面刷新未确认时，下次明确点击仍重放原key，complete之后才准新key',async()=>{
 let keys=0;const scope={projectId:'acknowledgement',engine:'e',episodeNumbers:[1],options:{usePlanning:false},explicitNewOperation:true,newKey:()=>`ack_${++keys}`};
 const original=await prepareStoryboardGenerationRequest({...scope,loadAttempts:empty});const attempt={request_key:original.requestOptions.idempotencyKey,engine:'e',episode_numbers:[1],status:'succeeded'};
 await reconcileStoryboardGenerationRequest({projectId:scope.projectId,cacheKey:original.cacheKey,loadAttempt:async()=>({attempt})});
 const recovered=await prepareStoryboardGenerationRequest({...scope,loadAttempts:async()=>({attempts:[attempt]})});assert.equal(recovered.requestOptions.idempotencyKey,original.requestOptions.idempotencyKey);assert.equal(recovered.replayed,true);assert.equal(keys,1);
 completeStoryboardGenerationRequest(recovered.cacheKey);const fresh=await prepareStoryboardGenerationRequest({...scope,loadAttempts:async()=>({attempts:[attempt]})});assert.notEqual(fresh.requestOptions.idempotencyKey,original.requestOptions.idempotencyKey);assert.equal(keys,2);
});


test('终态输入已改可明确fresh新建不带retryOf，unknown不能用fresh绕过；未确认授权保原key',async()=>{
 const attempt={request_key:'failed_original',engine:'e',episode_numbers:[1],status:'failed',can_retry:true};const fresh=await prepareStoryboardGenerationRequest({projectId:'changed',engine:'e',episodeNumbers:[1],options:{content:'已修改输入'},loadAttempts:async()=>({attempts:[attempt]}),confirmRetry:async()=> 'fresh',newKey:()=> 'fresh_changed'});assert.equal(fresh.allowed,true);assert.equal(fresh.requestOptions.idempotencyKey,'fresh_changed');assert.equal(fresh.requestOptions.retryOf,undefined);assert.equal(fresh.requestOptions.confirmedRetry,undefined);
 const unknown=await prepareStoryboardGenerationRequest({projectId:'changed_unknown',engine:'e',episodeNumbers:[1],loadAttempts:async()=>({attempts:[{...attempt,status:'unknown'}]}),confirmRetry:()=>assert.fail('未知不可确认fresh'),newKey:()=>assert.fail('未知不可新key')});assert.equal(unknown.allowed,false);
 const params={projectId:'permission_cancelled',engine:'e',episodeNumbers:[1],options:{content:'原输入'},loadAttempts:empty,newKey:()=> 'permission_original'};const first=await prepareStoryboardGenerationRequest(params);markStoryboardGenerationNotStarted(first.cacheKey);const repeated=await prepareStoryboardGenerationRequest(params);assert.equal(repeated.requestOptions.idempotencyKey,'permission_original');
 const changed=await prepareStoryboardGenerationRequest({...params,options:{content:'修改后输入'},newKey:()=> 'permission_changed'});assert.equal(changed.allowed,true);
});
