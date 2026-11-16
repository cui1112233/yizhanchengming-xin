import { apiRequest } from '../../shared/api/client.js';
function projectID(id){if(!id||!/^[a-zA-Z0-9_-]+$/.test(String(id)))throw Error('作品编号无效');return encodeURIComponent(id);}
export async function loadStoryboardPreferences(id,request=apiRequest){const result=await request(`/api/screenplays/${projectID(id)}/storyboard-preferences`,{suppressGlobalError:true});if(!result?.state||!Array.isArray(result.state.prefix_presets))throw Error('分镜偏好读取失败，请重试');return result;}
export function changeStoryboardPreferences(input,request=apiRequest){return request('/api/screenplays/storyboard-preferences/actions',{method:'POST',body:JSON.stringify(input),suppressGlobalError:true});}

// Eight sequential model rounds can exceed the shared ten-minute limit. The
// provider has its own timeout; the caller's explicit cancel signal stays active.
export function generateStoryboardPlan(id,input,signal,request=apiRequest){return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-plan`,{method:'POST',body:JSON.stringify(input),signal,timeoutMs:0,suppressGlobalError:true});}
export function loadStoryboardAssistantDraft(id,input,request=apiRequest){return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-assistant-draft`,{method:'POST',body:JSON.stringify(input),suppressGlobalError:true});}

export function loadStoryboardGenerationAttempt(id,key,request=apiRequest){if(!key||!String(key).match(/^[a-zA-Z0-9_-]+$/))throw Error('生成请求编号无效');return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-generation-attempts/${encodeURIComponent(key)}`,{suppressGlobalError:true});}
export function listStoryboardGenerationAttempts(id,{engine,episode}={},request=apiRequest){const params=new URLSearchParams();if(engine)params.set('engine',engine);if(Number(episode)>0)params.set('episode',String(episode));return request(`/api/comic-screenplays/projects/${projectID(id)}/storyboard-generation-attempts${params.size?'?'+params:''}`,{suppressGlobalError:true});}
