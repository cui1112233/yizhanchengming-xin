import { apiRequest } from '../../shared/api/client.js';
function projectID(id){if(!id||!/^[a-zA-Z0-9_-]+$/.test(String(id)))throw Error('作品编号无效');return encodeURIComponent(id);}
export async function loadStoryboardPreferences(id,request=apiRequest){const result=await request(`/api/screenplays/${projectID(id)}/storyboard-preferences`,{suppressGlobalError:true});if(!result?.state||!Array.isArray(result.state.prefix_presets))throw Error('分镜偏好读取失败，请重试');return result;}
export function changeStoryboardPreferences(input,request=apiRequest){return request('/api/screenplays/storyboard-preferences/actions',{method:'POST',body:JSON.stringify(input),suppressGlobalError:true});}

// Eight sequential model rounds can exceed the shared ten-minute limit. The
// provider has its own timeout; the caller's explicit cancel signal stays active.
export function generateStoryboardPlan(id,input,signal,request=apiRequest){return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-plan`,{method:'POST',body:JSON.stringify(input),signal,timeoutMs:0,suppressGlobalError:true});}
export function loadStoryboardPlanProgress(id,chapterId,key,signal,request=apiRequest){if(!chapterId||!String(key||'').trim()||String(key).length>256)throw Error('规划请求无效');const query=new URLSearchParams({chapter_id:chapterId,request_key:key});return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-plan-progress?${query}`,{signal,suppressGlobalError:true,timeoutMs:8000});}
export function loadStoryboardAssistantDraft(id,input,request=apiRequest){return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-assistant-draft`,{method:'POST',body:JSON.stringify(input),suppressGlobalError:true});}

export function loadStoryboardGenerationAttempt(id,key,request=apiRequest){if(!key||!String(key).match(/^[a-zA-Z0-9_-]+$/))throw Error('生成请求编号无效');return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-generation-attempts/${encodeURIComponent(key)}`,{suppressGlobalError:true});}
export function listStoryboardGenerationAttempts(id,{engine,episode}={},request=apiRequest){const params=new URLSearchParams();if(engine)params.set('engine',engine);if(Number(episode)>0)params.set('episode',String(episode));return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-generation-attempts${params.size?'?'+params:''}`,{suppressGlobalError:true});}

// 分镜合并是本地整理，不调用模型；dryRun 只算不影响已保存分镜。
export function mergeStoryboardShots(id,episode,engine,{maxSeconds,dryRun}={},request=apiRequest){
  if(!(Number(episode)>0))throw Error('剧集编号无效');
  if(!engine||!/^[a-zA-Z0-9_]+$/.test(String(engine)))throw Error('分镜引擎无效');
  const body={max_seconds:Number(maxSeconds)||30};
  if(dryRun)body.dry_run=true;
  return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboards/${encodeURIComponent(engine)}/${Number(episode)}/shots/merge`,{method:'POST',body:JSON.stringify(body),suppressGlobalError:true});
}

// 超长视频：按推荐时长分组，逐组生成后合成成片。
function longVideoBody({episode,engine,targetSeconds,capSeconds,indexes,sections}={}){
  const body={episode:Number(episode)||0,target_seconds:Number(targetSeconds)||300};
  if(engine)body.engine=engine;
  if(capSeconds)body.cap_seconds=Number(capSeconds);
  if(Array.isArray(indexes)&&indexes.length)body.indexes=indexes.map(Number).filter(value=>value>0);
  if(Array.isArray(sections)&&sections.length)body.sections=sections.map(section=>({index:Number(section.index)||0,prompt:String(section.prompt||''),shot_ids:section.shot_ids||[],total_seconds:Number(section.total_seconds)||0,label:section.label||''}));
  return body;
}
export function planLongVideo(id,input={},request=apiRequest){
  return request(`/api/comic-screenplays/projects/${projectID(id)}/long-video/plan`,{method:'POST',body:JSON.stringify(longVideoBody(input)),suppressGlobalError:true});
}
export function startLongVideo(id,input={},request=apiRequest){
  return request(`/api/comic-screenplays/projects/${projectID(id)}/long-video/generate`,{method:'POST',body:JSON.stringify(longVideoBody(input)),suppressGlobalError:true});
}
export function loadLongVideoStatus(id,{episode}={},request=apiRequest){
  const query=new URLSearchParams({episode:String(Number(episode)||0)});
  return request(`/api/comic-screenplays/projects/${projectID(id)}/long-video/status?${query}`,{suppressGlobalError:true});
}
export function combineLongVideo(id,{episode}={},request=apiRequest){
  return request(`/api/comic-screenplays/projects/${projectID(id)}/long-video/combine`,{method:'POST',body:JSON.stringify({episode:Number(episode)||0}),suppressGlobalError:true});
}

// 片段重拍：20 帧时间轴框选 + 首尾帧参考 + 拼回原视频。
function reshootBody({shotId,episode,engine,start,end,duration,prompt,mode}={}){
  const body={shot_id:String(shotId||'')};
  if(Number(episode)>0)body.episode=Number(episode);
  if(engine)body.engine=engine;
  if(start!==undefined)body.start=Number(start)||0;
  if(end!==undefined)body.end=Number(end)||0;
  if(duration!==undefined)body.duration=Number(duration)||0;
  if(prompt!==undefined)body.prompt=String(prompt||'');
  if(mode)body.mode=mode;
  return body;
}
export function listReshootCandidates(id,{episode,engine}={},request=apiRequest){
  const query=new URLSearchParams();
  if(Number(episode)>0)query.set('episode',String(Number(episode)));
  if(engine)query.set('engine',engine);
  return request(`/api/comic-screenplays/projects/${projectID(id)}/reshoot/candidates${query.size?'?'+query:''}`,{suppressGlobalError:true});
}
export function extractReshootFrames(id,{shotId,start,end}={},request=apiRequest){
  if(!String(shotId||'').trim())throw Error('请选择分镜');
  return request(`/api/comic-screenplays/projects/${projectID(id)}/reshoot/frames`,{method:'POST',body:JSON.stringify(reshootBody({shotId,start:start||0,end:end||0})),timeoutMs:120000,suppressGlobalError:true});
}
export function startReshoot(id,input={},request=apiRequest){
  if(!String(input.shotId||'').trim())throw Error('请选择分镜');
  return request(`/api/comic-screenplays/projects/${projectID(id)}/reshoot/generate`,{method:'POST',body:JSON.stringify(reshootBody(input)),timeoutMs:120000,suppressGlobalError:true});
}
export function loadReshootStatus(id,{shotId}={},request=apiRequest){
  const query=new URLSearchParams({shot_id:String(shotId||'')});
  return request(`/api/comic-screenplays/projects/${projectID(id)}/reshoot/status?${query}`,{suppressGlobalError:true});
}
export function applyReshoot(id,{shotId}={},request=apiRequest){
  return request(`/api/comic-screenplays/projects/${projectID(id)}/reshoot/apply`,{method:'POST',body:JSON.stringify(reshootBody({shotId})),timeoutMs:300000,suppressGlobalError:true});
}

// 换装：先圈分镜范围、再按场景筛选，把绑定的人物 A 批量换成目标人物（保留人物顺序）。
function costumePath(id,episode,engine,action){
  if(!(Number(episode)>0))throw Error('剧集编号无效');
  if(!engine||!/^[a-zA-Z0-9_]+$/.test(String(engine)))throw Error('分镜引擎无效');
  return `/api/comic-screenplays/projects/${projectID(id)}/storyboards/${encodeURIComponent(engine)}/${Number(episode)}/costume/${action}`;
}
function costumeBody({shotStart,shotEnd,scene,from,to}={}){
  const body={shot_start:Math.max(1,Number(shotStart)||1),shot_end:Math.max(0,Number(shotEnd)||0)};
  body.scene=String(scene||'').trim();
  body.from=String(from||'').trim();
  body.to=String(to||'').trim();
  return body;
}
export function loadCostumeOptions(id,{episode,engine}={},request=apiRequest){
  return request(costumePath(id,episode,engine,'options'),{suppressGlobalError:true});
}
export function previewCostumeChange(id,{episode,engine,...input}={},request=apiRequest){
  return request(costumePath(id,episode,engine,'preview'),{method:'POST',body:JSON.stringify(costumeBody(input)),suppressGlobalError:true});
}
export function applyCostumeChange(id,{episode,engine,...input}={},request=apiRequest){
  return request(costumePath(id,episode,engine,'apply'),{method:'POST',body:JSON.stringify(costumeBody(input)),suppressGlobalError:true});
}

// 批量翻译：按类型（画面/视频提示词）与方向（中→英 / 英→中）逐条翻译分镜提示词，空提示词回退台词。
function translatePath(id,episode,engine,action){
  if(!(Number(episode)>0))throw Error('剧集编号无效');
  if(!engine||!/^[a-zA-Z0-9_]+$/.test(String(engine)))throw Error('分镜引擎无效');
  return `/api/comic-screenplays/projects/${projectID(id)}/storyboards/${encodeURIComponent(engine)}/${Number(episode)}/translate/${action}`;
}
function translateBody({field,direction,scope,shotStart,shotEnd}={}){
  return {field:String(field||'image'),direction:String(direction||'zh2en'),scope:String(scope||'all'),shot_start:Math.max(1,Number(shotStart)||1),shot_end:Math.max(0,Number(shotEnd)||0)};
}
export function prepareBatchTranslate(id,{episode,engine,...input}={},request=apiRequest){
  return request(translatePath(id,episode,engine,'prepare'),{method:'POST',body:JSON.stringify(translateBody(input)),suppressGlobalError:true});
}
export function translateStoryboardItem(id,{episode,engine,shotId,model,...input}={},request=apiRequest){
  return request(translatePath(id,episode,engine,'item'),{method:'POST',body:JSON.stringify({...translateBody(input),shot_id:String(shotId||''),model:String(model||'')}),timeoutMs:180000,suppressGlobalError:true});
}

// 分镜检查（洛水 storyboard_check）：检查序号缺失/重复/异常并出报告，缺失分镜可由 AI 逐条补齐。
function checkPath(id,episode,engine,action){
  if(!(Number(episode)>0))throw Error('剧集编号无效');
  if(!engine||!/^[a-zA-Z0-9_]+$/.test(String(engine)))throw Error('分镜引擎无效');
  return `/api/comic-screenplays/projects/${projectID(id)}/storyboards/${encodeURIComponent(engine)}/${Number(episode)}/check${action}`;
}
export function checkStoryboardShots(id,{episode,engine}={},request=apiRequest){
  return request(checkPath(id,episode,engine,''),{method:'POST',body:JSON.stringify({}),suppressGlobalError:true});
}
export function fillStoryboardCheckShot(id,{episode,engine,seq,model}={},request=apiRequest){
  if(!(Number(seq)>0))throw Error('请选择要补全的分镜序号');
  return request(checkPath(id,episode,engine,'/fill'),{method:'POST',body:JSON.stringify({seq:Number(seq),model:String(model||'')}),timeoutMs:180000,suppressGlobalError:true});
}

// 医生审核（洛水 luoshui_doctors_v1 review 模式）：剧本医生审剧本全文，分镜医生对照剧本审分镜；只出报告不改稿。
function doctorPath(id,episode,engine){
  if(!(Number(episode)>0))throw Error('剧集编号无效');
  if(!engine||!/^[a-zA-Z0-9_]+$/.test(String(engine)))throw Error('分镜引擎无效');
  return `/api/comic-screenplays/projects/${projectID(id)}/storyboards/${encodeURIComponent(engine)}/${Number(episode)}/doctor/storyboard`;
}
export function reviewScreenplayDoctor(id,{model}={},request=apiRequest){
  if(!String(model||'').trim())throw Error('请先选择文本模型');
  return request(`/api/comic-screenplays/projects/${projectID(id)}/doctor/screenplay`,{method:'POST',body:JSON.stringify({model:String(model||'')}),timeoutMs:300000,suppressGlobalError:true});
}
export function reviewStoryboardDoctor(id,{episode,engine,model}={},request=apiRequest){
  if(!String(model||'').trim())throw Error('请先选择文本模型');
  return request(doctorPath(id,episode,engine),{method:'POST',body:JSON.stringify({model:String(model||'')}),timeoutMs:300000,suppressGlobalError:true});
}
