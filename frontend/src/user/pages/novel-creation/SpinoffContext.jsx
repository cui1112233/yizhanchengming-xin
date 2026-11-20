import { useEffect, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
export default function SpinoffContext({projectId,refreshKey,onLoaded}){
 const [creation,setCreation]=useState(null),[error,setError]=useState('');
 useEffect(()=>{let current=true;const controller=new AbortController();setCreation(null);setError('');novelProjectRequest(projectId,'/spinoff',undefined,{signal:controller.signal}).then(r=>{if(current){setCreation(r.creation||null);onLoaded?.(projectId,r.creation||null);}}).catch(e=>{if(current&&!e.aborted)setError(e.message||'番外来源读取失败');});return()=>{current=false;controller.abort();};},[projectId,refreshKey]);
 if(error)return <p role="alert">{error}</p>;if(!creation)return null;
 return <section className="novel-spinoff-context"><p>正传：{creation.parent.title} · {({sequel:'后传',prequel:'前传',parallel:'平行线',if:'If 线'})[creation.input.spinoffType]}</p><p>{creation.input.storyIdea}</p>{creation.parentCanon?<><details><summary>正传正典</summary><pre style={{whiteSpace:'pre-wrap'}}>{creation.parentCanon}</pre></details><details><summary>番外设定</summary><pre style={{whiteSpace:'pre-wrap'}}>{[creation.foundation?.world, creation.foundation?.coreConflict, ...(creation.foundation?.characters||[]).map(c=>`${c.name}：${c.role}，目标：${c.goal}`), ...(creation.foundation?.volumes||[]).map(v=>`${v.title}：${v.goal}`)].filter(Boolean).join('\n\n')}</pre></details></>:<p>正典与设定尚未完成，请查看初始化任务状态。</p>}</section>;
}
