import {listStoryboardGenerationAttempts,loadStoryboardGenerationAttempt} from './comic-screenplay-storyboard-api.js';
const requests=new Map();
const retryable=attempt=>attempt?.can_retry===true&&['failed','partial','cancelled'].includes(attempt.status);
function stable(value){if(Array.isArray(value))return value.map(stable);if(value&&typeof value==='object')return Object.fromEntries(Object.keys(value).sort().filter(key=>!['signal','idempotencyKey','retryOf','confirmedRetry','confirmed'].includes(key)&&value[key]!==undefined).map(key=>[key,stable(value[key])]));return value;}
function context(input){const episodes=[...new Set((input.episodeNumbers||[]).map(Number))].sort((a,b)=>a-b);if(!input.projectId||!episodes.length||episodes.some(n=>!Number.isInteger(n)||n<1))throw Error('生成目标集无效');return {projectId:input.projectId,engine:input.engine||'',episodeNumbers:episodes,options:stable(input.options||{})};}
function cacheKey(input){return JSON.stringify(context(input));}
function sameEpisodes(a,b){const left=[...new Set(a||[])].map(Number).sort((x,y)=>x-y),right=[...new Set(b||[])].map(Number).sort((x,y)=>x-y);return JSON.stringify(left)===JSON.stringify(right);}
function episodesOf(attempt){return attempt.episode_numbers||attempt.episodes?.map(item=>item.episode_no)||[];}
function adopt(attempt){for(const record of requests.values())if(record.key===attempt.request_key){record.attempt=attempt;record.status=attempt.status;}}
export function createStoryboardGenerationRecord(input,{retryOf,confirmedRetry=false,forceNew=false,newKey=()=>crypto.randomUUID()}={}){const scope=context(input),key=cacheKey(scope),old=requests.get(key);if(old&&!forceNew&&!retryOf)return old;const record={context:scope,cacheKey:key,key:newKey(),options:scope.options,status:'prepared',acknowledged:false,attempt:null,...(retryOf&&confirmedRetry?{retryOf,confirmedRetry:true}:{})};requests.set(key,record);return record;}
export function isStoryboardGenerationAcknowledged(requestKey){return [...requests.values()].some(record=>record.key===requestKey&&record.acknowledged===true);}
export function rememberStoryboardGenerationAttempt(input,record){const acknowledged=record.acknowledged===true||isStoryboardGenerationAcknowledged(record.key);const scope=context(input),key=cacheKey(scope),next={...record,acknowledged,context:scope,cacheKey:key,options:scope.options,status:record.attempt?.status||record.status||'prepared'};requests.set(key,next);return next;}
export function markStoryboardGenerationNotStarted(key){const record=requests.get(key);if(record){record.status='not_started';record.acknowledged=false;record.attempt=null;}}
export function completeStoryboardGenerationRequest(key){const record=requests.get(key);if(record){record.status='succeeded';record.acknowledged=true;}}
export async function reconcileStoryboardGenerationRequest({projectId,cacheKey:key,idempotencyKey,loadAttempt=loadStoryboardGenerationAttempt}){const record=requests.get(key),requestKey=idempotencyKey||record?.key;if(!requestKey)throw Error('缺少原生成请求编号');const result=await loadAttempt(projectId,requestKey);if(!result?.attempt)throw Error('生成状态返回不完整');adopt(result.attempt);return result.attempt;}
export async function prepareStoryboardGenerationRequest({projectId,engine,episodeNumbers,options={},explicitNewOperation=false,confirmRetry,loadAttempts=listStoryboardGenerationAttempts,newKey=()=>crypto.randomUUID()}){
 const scope=context({projectId,engine,episodeNumbers,options}),key=cacheKey(scope),old=requests.get(key);let rows;
 try{const result=await loadAttempts(projectId,{engine,episode:scope.episodeNumbers.length===1?scope.episodeNumbers[0]:undefined});if(!Array.isArray(result?.attempts))throw Error('生成请求列表返回不完整');rows=result.attempts;rows.forEach(adopt);}catch(error){return {allowed:false,cacheKey:key,attempt:old?.attempt,message:`生成状态未确认：${error.message||'请重试核对'}。不会新建付费请求。`};}
 const ordered=rows.filter(item=>!engine||item.engine===engine).sort((a,b)=>(Date.parse(b.created_at)||0)-(Date.parse(a.created_at)||0));
 const running=ordered.find(item=>['running','unknown'].includes(item.status)&&episodesOf(item).some(n=>scope.episodeNumbers.includes(Number(n))));if(running)return {allowed:false,cacheKey:key,attempt:running,message:'原分镜请求仍在运行或状态未知，请先核对，不能新建请求。'};
 const latest=ordered.find(item=>sameEpisodes(episodesOf(item),scope.episodeNumbers));let record;
 if(latest&&['failed','partial','cancelled'].includes(latest.status)){
  if(!retryable(latest))return {allowed:false,cacheKey:key,attempt:latest,message:'原批次暂不能续跑，请核对已保存结果。'};
  const choice=confirmRetry?await confirmRetry(latest):false;if(choice!=='fresh'&&choice!==true)return {allowed:false,cacheKey:key,attempt:latest,message:'已保留原批次；未确认续跑，不会调用模型。'};
  record=choice==='fresh'?createStoryboardGenerationRecord(scope,{forceNew:true,newKey}):createStoryboardGenerationRecord(scope,{retryOf:latest.request_key,confirmedRetry:true,forceNew:true,newKey});
 }else if(latest?.status==='succeeded'&&(!explicitNewOperation||![...requests.values()].some(item=>item.key===latest.request_key&&item.acknowledged===true))){record=rememberStoryboardGenerationAttempt(scope,{key:latest.request_key,options:scope.options,attempt:latest,status:'succeeded',acknowledged:[...requests.values()].some(item=>item.key===latest.request_key&&item.acknowledged===true)});
 }else{
  const pendingLocal=[...requests.values()].find(item=>item.cacheKey!==key&&item.context.projectId===projectId&&item.context.engine===scope.engine&&['prepared','running','unknown'].includes(item.status)&&item.context.episodeNumbers.some(n=>scope.episodeNumbers.includes(n)));
  if(pendingLocal)return {allowed:false,cacheKey:key,attempt:pendingLocal.attempt,message:'同目标仍有未核对的原请求，请先在原入口核对。'};
  record=createStoryboardGenerationRecord(scope,{forceNew:latest?.status==='succeeded'&&explicitNewOperation||old?.status==='succeeded'&&old.acknowledged===true&&explicitNewOperation,newKey});
 }
 return {allowed:true,cacheKey:record.cacheKey,attempt:record.attempt,replayed:record.status==='succeeded',message:record.status==='succeeded'?'沿用原请求核对已保存结果，不会重复计费。':'',requestOptions:{...scope.options,idempotencyKey:record.key,...(record.retryOf?{retryOf:record.retryOf,confirmedRetry:true}:{})}};
}
