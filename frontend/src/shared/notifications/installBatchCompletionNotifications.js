import { getCurrentUsername } from '../api/auth.js';
import { apiRequest, getToken } from '../api/client.js';
import { getConfig } from '../api/config.js';
import { dispatchGlobalTaskNotification } from './globalTaskCenter.js';
import { createBatchCompletionMonitor } from './batchCompletion.js';
import { createBatchCompletionAudio } from './batchCompletionAudio.js';
import { installNotificationPreferencesSync } from './notificationPreferencesSync.js';

let installed=false;
let sharedAudio=null;
const getAudio=()=>sharedAudio ||= createBatchCompletionAudio();
export async function previewBatchCompletionSound(preferences,options={}){
  const audio=getAudio();
  const session=getToken(),owner=getCurrentUsername();
  if(!await audio.prepare())return false;
  return audio.play(preferences,{...options,isCurrent:()=>session===getToken() && owner===getCurrentUsername()});
}
export function installBatchCompletionNotifications(){
  if(installed || typeof window==='undefined')return;
  installed=true;
  const audio=getAudio();
  const monitor=createBatchCompletionMonitor({
    getIdentity:()=>({owner:getCurrentUsername(),session:getToken()}),
    fetchJobs:()=>apiRequest('/api/batch-factory/v12/automation/notifications',{cache:'no-store',suppressGlobalError:true,timeoutMs:10000}),
    fetchPreferences:async()=>(await getConfig()).notifications || {},
    storage:window.localStorage,sessionStorage:window.sessionStorage,locks:navigator.locks,
    notify:job=>dispatchGlobalTaskNotification({
      id:`batch-complete:${job.completionKey}`,status:'success',title:'批量工厂全部完成',
      detail:`本批次 ${job.total} 本书已全部完成`,page:'批量工厂',pagePath:`/batch-factory?batchId=${encodeURIComponent(job.batchId)}`
    }),
    play:(preferences,options)=>audio.play(preferences,options)
  });
  const poll=()=>void monitor.poll();
  const prepare=()=>void audio.prepare().then(()=>{if(monitor.hasPending())poll();});
  window.addEventListener('pointerdown',prepare,{capture:true});
  window.addEventListener('keydown',prepare,{capture:true});
  installNotificationPreferencesSync({target:window,storage:window.localStorage,getOwner:getCurrentUsername,onUpdate:monitor.setPreferences});
  window.addEventListener('focus',poll);
  window.setInterval(poll,10000);
  poll();
}
