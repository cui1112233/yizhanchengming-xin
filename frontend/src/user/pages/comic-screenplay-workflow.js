const labels={scheduled:'等待定时启动',idle:'尚未启动',running:'后台执行中',paused:'已暂停后续提交',failed:'执行失败，可检查后重试',blocked:'等待处理质检问题',clips_ready:'镜头生成完成，待编辑与合成',completed:'生产步骤完成',pending:'等待执行',queued:'排队中',submitting:'正在提交',submitted:'已提交，查询结果中',succeeded:'已完成',uncertain:'提交结果待核对'};
export function comicProductionView(data){
  const status=data?.status||data?.batch?.status||'idle';
  const finalize=data?.batch?.steps?.['8_finalize'];
  const deliverables=finalize?.status==='done'?[...new Set((finalize.artifacts||[]).map(safeComicMediaUrl).filter(Boolean))]:[];
  const deliveryComplete=status==='completed'&&deliverables.length>0;
  const label=status==='completed'?(deliveryComplete?'成片已生成，可查看交付结果':'生产步骤完成，尚无交付结果'):(labels[status]||status);
  return {status,label,scheduled:status==='scheduled',active:['running','queued','submitting','submitted'].includes(status),deliveryComplete,deliverables};
}
export function comicProductionShotKey(row){return JSON.stringify([row.project_id||'',row.storyboard_id||'',row.episode,row.seq??row.shot_id??row.shot??row.id,row.request_hash||'']);}
export function safeComicMediaUrl(value){
  const text=String(value||'');return /^https?:\/\//i.test(text)||/^\/api\//.test(text)?text:'';
}
export function buildComicTimeline(clips,items){
  let videoCursor=0;
  return [...clips].sort((a,b)=>(a.order||0)-(b.order||0)).map((clip,index)=>{
    const item=items.find(row=>row.id===clip.itemId);if(!item)throw new Error('时间线包含不存在的素材');
    const durationMs=Math.round(Number(clip.durationMs));if(!(durationMs>0))throw new Error('请填写每个片段的有效时长');
    const sourceStartMs=Math.round(Number(clip.sourceStartMs)||0),scale=Number(clip.scale)||1;
    if(sourceStartMs<0||scale<.1||scale>4)throw new Error('片段裁切或缩放设置无效');
    const startMs=item.kind==='audio'?Math.round(Number(clip.startMs)||0):videoCursor;
    if(startMs<0)throw new Error('片段开始时间不能为负数');
    if(item.kind!=='audio')videoCursor=startMs+durationMs;
    return {...clip,startMs,sourceStartMs,durationMs,scale,order:index};
  });
}

export function comicScheduleRequest(value,now=Date.now()){
 const at=new Date(value||'');if(!Number.isFinite(at.getTime())||at.getTime()<=now)throw new Error('请选择未来的定时制作时间');return {run_at:at.toISOString()};
}
export function comicLocalDateTime(value){const at=new Date(value||'');if(!Number.isFinite(at.getTime()))return '';return new Date(at.getTime()-at.getTimezoneOffset()*60000).toISOString().slice(0,16);}
