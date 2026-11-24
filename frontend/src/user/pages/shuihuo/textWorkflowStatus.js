import {defaultTextProductionConfig} from './textProductionConfig.js';
import {layoutFields} from './textMediaActions.js';

function layout(snapshot,legacy=false){
 const result={};const defaults=defaultTextProductionConfig();
 for(const [key,path] of Object.entries(layoutFields))result[path]=snapshot?.layout?.[path] ?? defaults[key];
 if(legacy && !snapshot?.layout?.paginationMode)result.paginationMode='target';
 return result;
}
function background(value){return {kind:value?.kind || 'color',color:value?.color || '#111318',id:value?.id || ''};}
const equal=(a,b)=>JSON.stringify(a)===JSON.stringify(b);

export function collectionIsOutdated(collection,document,pages=[]){
 if(!collection || !document)return false;
 if(collection.versionId && collection.versionId!==document.activeVersionId)return true;
 if(collection.paginationVersion && collection.paginationVersion!==document.paginationVersion)return true;
 return (collection.items || []).some(item=>{const page=pages.find(page=>page.id===item.pageId);return !page || page.activeArtifactId!==item.artifactId;});
}

export function textUpdateState(snapshot={},collection){
 const doc=snapshot.document,pages=snapshot.pages || [];
 if(snapshot.settingsError)return {stage:'blocked',label:snapshot.settingsError};
 if(!doc)return {stage:'prepare',label:'尚未分页'};
 const next=snapshot.settingsSnapshot,old=doc.configSnapshot;
 let reflow=false,repaint=false;
 if(next && old){
  const a=layout(old,true),b=layout(next);
  const paintKeys=['align','textColor','highlightColor','highlightRules','highlightLines'];
  repaint=paintKeys.some(key=>a[key]!==b[key]) || !equal(background(old.background),background(next.background)) || Boolean(old.highlightReuseEnabled)!==Boolean(next.highlightReuseEnabled) || old.highlightRevision!==next.highlightRevision || old.coverFingerprint!==next.coverFingerprint;
  for(const key of paintKeys){delete a[key];delete b[key];}
  if(a.paginationMode==='auto' && b.paginationMode==='auto'){delete a.targetCharacters;delete b.targetCharacters;}
  reflow=!equal(a,b);
 }
 const pending=pages.filter(page=>page.status!=='ready' || !page.activeArtifactId);
 const locked=pages.filter(page=>page.locked).length;
 const base={locked,pending:pending.length,settingsChanged:reflow || repaint};
 if(reflow)return {...base,stage:'reflow',label:'排版已修改，待重新分页和生成'};
 if(repaint)return {...base,stage:'render',label:'样式已修改，待更新图片'};
 if(pending.length)return {...base,stage:'render',label:`${pending.length} 页待生成`};
 if(!collection || collectionIsOutdated(collection,doc,pages))return {...base,stage:'compose',label:'文字页已就绪，合集待更新'};
 return {...base,stage:'ready',label:'文字页与合集已就绪'};
}
