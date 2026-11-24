// Shared framework-free metadata and local transforms are the same code used by the website.
import {PLUGIN_CATALOG} from '../../../frontend/src/user/pages/pluginCatalog.js';
import {formatText,mergeDocuments,mergeSubtitles} from '../../../frontend/src/user/components/localToolLogic.js';
import {parseStoryboardFiles,buildStoryboardJson} from '../../../frontend/src/user/components/storyboardOrganizer.js';
import type {CmRow} from './cmCanvas';
type Request=(path:string,options?:any)=>Promise<any>;
export const PLUGINS=PLUGIN_CATALOG as Array<{id:string;name:string;summary:string;color:string;isNew?:boolean;launch:CmRow}>;
const textTools=new Set(['novel-to-drama','ai-rewrite','storyboard-master','viral-opening','character-edit']);
export function mobilePluginKind(id:string){const item=PLUGINS.find(p=>p.id===id);if(!item)return'unknown';if(textTools.has(id))return'text';const kind=item.launch.kind;if(kind==='local-tool')return'local';if(id==='ai-cut-plugin')return'aicut';if(kind==='desktop-tool'||kind==='unavailable')return'desktop';if(['asset-reference','cover-generation','character-turnaround'].includes(id))return'image';if(id==='ai-segment')return'segment';if(kind==='upscale-tool')return'upscale';if(kind==='draft-tool')return'draft';return'unknown';}
const rows=(r:CmRow,key:string)=>{if(!Array.isArray(r?.[key]))throw Error('服务端目录格式不完整');return r[key] as CmRow[];};
const tool=(id:string)=>{if(!textTools.has(id)&&id!=='cover-generation')throw Error('请选择有效的文字插件');return encodeURIComponent(id);};
export function createPluginClient(request:Request){return{
 prompts:async(id:string)=>rows(await request(`/api/plugins/${tool(id)}/config`,{timeout:15000}),'prompts'),
 run:async(id:string,input:string,modelId:string,promptId:string)=>{const safe=tool(id);if(!input.trim()||!modelId||!promptId)throw Error('请填写正文，并选择模型和提示词');const r=await request(`/api/plugins/${safe}/run`,{method:'POST',data:{input:input.trim(),modelId,promptId,raw:false},timeout:600000});if(typeof r?.result!=='string'||!r.result.trim())throw Error('生成结果未确认，请核对后再决定是否重新生成');return r.result as string;},
 models:async(kind:string)=>rows(await request(`/api/models?kind=${encodeURIComponent(kind)}`),'models').filter(m=>m.enabled!==false),
 templates:async()=>rows(await request('/api/presets?module=shuihuo-production'),'catalog').filter(p=>p.slot==='shuihuo.asset.character-sheet'||['shuihuo-character-three-view','shuihuo-character-accessory-sheet','shuihuo-character-color-sheet'].includes(p.id)),
 template:async(id:string,description:string)=>{let r:CmRow;try{r=await request('/api/plugins/character-turnaround/prompt',{method:'POST',data:{presetId:id,description}});}catch(e:any){if(e.statusCode===404)throw Error('当前服务尚未更新三视图模板接口；可手动填写最终生成要求，或待服务端更新后预览模板');throw e;}if(typeof r?.prompt!=='string'||!r.prompt.trim())throw Error('服务端未返回三视图生成要求');return r.prompt;},
 cover:async(data:CmRow,design=false)=>{const r=await request(`/api/plugins/cover-generation/${design?'design':'prompt'}`,{method:'POST',data,timeout:design?300000:15000});if(typeof r?.prompt!=='string'||!r.prompt.trim())throw Error('未取得封面要求');return r.prompt as string;},
 segment:(data:CmRow)=>request('/api/plugins/ai-segment/analyze',{method:'POST',data,timeout:600000}),
 upscaleConfig:()=>request('/api/ai-upscale/config'),
 upscale:(id:string)=>request(`/api/ai-upscale/jobs/${encodeURIComponent(id)}`),
 cancelUpscale:(id:string)=>request(`/api/ai-upscale/jobs/${encodeURIComponent(id)}/cancel`,{method:'POST'}),
};}
export function runLocalPlugin(id:string,sources:Array<{name:string;content:string}>,options:CmRow){
 if(!sources.some(s=>s.content.trim()))throw Error('请先粘贴需要处理的内容');
 if(id==='text-format')return formatText(sources.map(s=>s.content).join('\n\n'),options);
 if(id==='document-merge')return mergeDocuments(sources,options);
 if(id==='subtitle-merge'){const r=mergeSubtitles(sources,options);if(!r.cues.length)throw Error('未识别到有效字幕，请检查SRT、ASS或VTT格式');return r;}
 if(id==='storyboard-organizer'){const shots=parseStoryboardFiles(sources,{defaultShotSeconds:Number(options.defaultShotSeconds)||3});if(!shots.length)throw Error('没有识别到分镜，请检查镜头编号和画面描述');const json=buildStoryboardJson(shots,{targetSeconds:Number(options.targetSeconds)||15});return{text:JSON.stringify(json,null,2),stats:{shots:shots.length,groups:json.groupCount},json};}
 throw Error('此插件不是本地文字处理工具');
}
