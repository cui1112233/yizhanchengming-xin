import {stripEpisodeHeading} from './comic-screenplay-episode-headings.js';
export const storyboardDensityOptions=[{value:'compact',label:'更少小节',caption:'更紧凑'},{value:'standard',label:'标准',caption:'默认'},{value:'more',label:'更多小节',caption:'细拆'},{value:'sd25',label:'SD2.5专项规划',caption:'30秒一组'}];
export function normalizeStoryDensity(value){
 if(typeof value==='boolean'||value==null)return 'standard';
 if(/^[0-3]$/.test(String(value)))return storyboardDensityOptions[Number(value)].value;
 const text=String(value).trim().toLowerCase();if(storyboardDensityOptions.some(item=>item.value===text))return text;
 return ['sd2.5','sd2_5','sd2.5专项','sd2.5专项规划','sd25专项','sd25专项规划','seedance25','seedance2.5'].includes(text)?'sd25':'standard';
}
const dateValue=row=>row.updated_at||row.updatedAt||row.created_at||row.createdAt||'';
const dateInput=value=>String(value).trim().replace(' ','T');
export function storyboardPromptDate(row){const raw=String(dateValue(row)).trim();if(!raw)return '';const date=new Date(dateInput(raw));if(!Number.isFinite(date.getTime()))return raw.slice(0,16);const pad=n=>String(n).padStart(2,'0');return `${pad(date.getMonth()+1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;}
export function storyboardPromptRows(prompts,query='',{includeCompatibility=false}={}){
 const search=String(query).trim().toLowerCase();return (prompts||[]).filter(item=>(includeCompatibility||item.catalog_group!=='compatibility')&&String(item.purpose||item.type||'').replaceAll('-','_')==='storyboard'&&(String(item.name||'').toLowerCase().includes(search)||String(item.description||'').toLowerCase().includes(search))).map(item=>({...item,id:String(item.id)})).sort((a,b)=>{const time=(Date.parse(dateInput(dateValue(b)))||0)-(Date.parse(dateInput(dateValue(a)))||0);if(time)return time;return String(b.source_id||b.id).localeCompare(String(a.source_id||a.id),'zh-CN',{numeric:true});});
}
export function storyboardPromptSections(prompts,favorites=[],recent=[],query=''){const rows=storyboardPromptRows(prompts,query),byId=new Map(rows.map(item=>[String(item.id),item]));const favoriteIds=[...new Set(favorites.map(String))];return {favorites:favoriteIds.map(id=>byId.get(id)).filter(Boolean),recent:[...new Set(recent.map(String))].filter(id=>!favoriteIds.includes(id)).map(id=>byId.get(id)).filter(Boolean).slice(0,8)};}
export function validateStoryboardPrefix(preset){const name=String(preset.name||'').trim(),content=String(preset.content||'').trim();if(!name)throw Error('预设名称不能为空');if(!content)throw Error('预设内容不能为空');if(content.length>5000)throw Error('预设内容不能超过5000字');return {...(preset.id?{id:String(preset.id)}:{}),name,content};}
export function importStoryboardPrefixes(text,existing=[],newId=()=>`sb_preset_${Date.now()}_${crypto.randomUUID().slice(0,4)}`){let parsed;try{parsed=JSON.parse(text);}catch{throw Error('请选择有效的JSON文件');}if(!Array.isArray(parsed))throw Error('导入内容必须是预设数组');const items=existing.map(item=>({id:String(item.id),name:item.name,content:item.content})),seen=new Set(items.map(item=>item.id));let added=0,duplicates=0,invalid=0;for(const raw of parsed){let preset;try{preset=validateStoryboardPrefix(raw||{});}catch{invalid++;continue;}const id=preset.id||newId();if(seen.has(id)){duplicates++;continue;}seen.add(id);items.push({...preset,id});added++;}return {items,added,duplicates,invalid};}

export function matchStoryboardChapter(chapters,episode){const content=stripEpisodeHeading(episode?.content);if(!content)return '';const matches=(chapters||[]).filter(item=>stripEpisodeHeading(item.script_text)===content);return matches.length===1?matches[0].id:'';}
export function storyboardPlanningAttempt(chapter,key){if(!chapter||chapter.planning_request_key!==key)return 'unknown';return ['succeeded','running','failed','cancelled'].includes(chapter.planning_status)?chapter.planning_status:'unknown';}

export function storyboardGenerationCanRetry(attempt){return attempt?.can_retry===true&&['partial','failed','cancelled'].includes(attempt.status);}
export function storyboardRestorableGeneration(attempts,episode){const rows=(attempts||[]).filter(item=>(item.episode_numbers||item.episodes?.map(row=>row.episode_no)||[]).map(Number).includes(Number(episode))).sort((a,b)=>(Date.parse(b.created_at)||0)-(Date.parse(a.created_at)||0));const pending=rows.find(item=>['running','unknown'].includes(item.status));if(pending)return pending;const latest=rows[0];return latest&&latest.status!=='succeeded'?latest:null;}
