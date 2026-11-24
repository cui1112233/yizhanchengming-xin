import type {CmRow} from './cmCanvas';
import {onSessionInvalidated} from './sessionBoundary';
export type CmIntent={project:string;node?:string;version?:number;kind:'split'|'batch'|'other';requestId?:string;knownTaskIds?:string[];shotIds?:string[]};
const operations=new Map<string,CmIntent>();
onSessionInvalidated(()=>operations.clear());
export function recordCmOperation(key:string,intent:CmIntent){operations.set(key,intent);}
export const cmOperationPending=(key:string)=>operations.has(key);
export const finishCmOperation=(key:string)=>operations.delete(key);
export function reconcileCmOperations(project:string,tasks:CmRow[]){for(const [key,intent] of operations){if(intent.project!==project)continue;const found=intent.kind==='batch'?Boolean(intent.shotIds?.length)&&intent.shotIds!.every(shot=>tasks.some(task=>task.settings?.batchRequestId===intent.requestId&&task.settings?.source?.shotId===shot)):intent.kind==='split'&&tasks.some(task=>task.nodeId===intent.node&&Number(task.nodeVersion)===intent.version&&task.mode==='script_split'&&!(intent.knownTaskIds||[]).includes(task.id));if(found)operations.delete(key);}}
