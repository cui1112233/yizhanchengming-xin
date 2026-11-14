import { apiRequest } from './client.js';
import { comicWorkflowPath } from './comicScreenplayWorkflow.js';
export function loadVoiceSwitchLibrary(request=apiRequest) { return request('/api/screenplays/voice-library',{suppressGlobalError:true}); }
export function readVoiceSwitchMedia(projectId,request=apiRequest) { return request(comicWorkflowPath(projectId,'media'),{suppressGlobalError:true}); }
export function bindComicVoice(projectId,input,request=apiRequest) { return request(comicWorkflowPath(projectId,'media/actions'),{method:'POST',body:JSON.stringify({...input,action:'voice_bind'}),timeoutMs:0,suppressGlobalError:true}); }
export function uploadComicVoiceReference(projectId,file,targetId,expectedRevision,request=apiRequest) {
  if(!(file instanceof Blob)) return Promise.reject(new Error('请选择WAV或MP3音频'));
  const body=new FormData();body.append('file',file,file.name||'reference.wav');body.append('targetId',targetId);body.append('expectedRevision',String(expectedRevision));
  return request(comicWorkflowPath(projectId,'media/voice-upload'),{method:'POST',body,timeoutMs:0,suppressGlobalError:true});
}
