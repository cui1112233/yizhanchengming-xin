export function createEpisodeRunner(){
 let busy=false,stopped=false;
 return {stop(){stopped=true;},get busy(){return busy;},async run(items,execute,onProgress=()=>{}){
  if(busy)throw Error('批量执行中，请等待当前操作');if(!items.length)throw Error('请先选择集数');busy=true;stopped=false;const results=[];
  try{for(const item of items){if(stopped){results.push({item,status:'not_started'});continue;}onProgress([...results,{item,status:'running'}]);try{const value=await execute(item);results.push({item,status:value===false?'unconfirmed':'succeeded',value});}catch(error){results.push({item,status:'failed',error:error.message||'操作失败'});}onProgress([...results]);}onProgress([...results]);return results;}finally{busy=false;}
 }};
}
export function validateEpisodeBatch({selected,modelId,mode='none',task='generate',bindings={},chapters=[],useOriginal=false}){
 if(!selected?.length)throw Error('请先选择集数');if(!modelId)throw Error('请先选择并保存文本模型');
 for(const ep of selected){const c=chapters.find(c=>c.id===bindings[ep.no]);if(task==='plan'||mode!=='none'||useOriginal){if(!c)throw Error('第'+ep.no+'集需要明确绑定章节');if(!String(c[useOriginal?'source_text':'script_text']||'').trim())throw Error('第'+ep.no+'集对应章节内容为空');}if(task==='generate'&&mode==='existing'&&!String(c?.planning_text||'').trim())throw Error('第'+ep.no+'集还没有已保存规划');}
 return true;
}
export function storyboardAttemptRows(attempts,engine){
 const seen=new Set(),rows=[];
 for(const a of [...attempts].filter(a=>!engine||a.engine===engine).sort((a,b)=>(Date.parse(b.created_at)||0)-(Date.parse(a.created_at)||0))){
  for(const ep of a.episodes||[]){const no=Number(ep.episode_no);if(!no||seen.has(no))continue;seen.add(no);rows.push({episodeNo:no,status:ep.status,error:ep.error||a.error||'',failed:['failed','partial','cancelled'].includes(ep.status),requestKey:a.request_key,canRetry:a.can_retry===true,attempt:a});}
 }return rows.sort((a,b)=>a.episodeNo-b.episodeNo);
}
