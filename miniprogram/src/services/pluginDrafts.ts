import {onSessionInvalidated,sessionRevision} from './sessionBoundary';
type Data=Record<string,any>;
const drafts=new Map<string,Data>(),listeners=new Map<string,Set<(draft:Data,patch:Data)=>void>>();
onSessionInvalidated(()=>{drafts.clear();for(const group of listeners.values())for(const listener of group)listener({},{});});
export function pluginDraft(id:string){return drafts.get(id);}
export function rememberPluginDraft(id:string,patch:Data,session:number){if(session!==sessionRevision())return;const next={...drafts.get(id),...patch};drafts.set(id,next);for(const listener of listeners.get(id)||[])listener(next,patch);}
export function subscribePluginDraft(id:string,listener:(draft:Data,patch:Data)=>void){let group=listeners.get(id);if(!group){group=new Set();listeners.set(id,group);}group.add(listener);return()=>{group!.delete(listener);if(!group!.size)listeners.delete(id);};}
export function beginPluginOperation(id:string,label:string,session:number){if(session!==sessionRevision()||drafts.get(id)?.running)return false;rememberPluginDraft(id,{running:label},session);return true;}
export function endPluginOperation(id:string,session:number){rememberPluginDraft(id,{running:''},session);}
