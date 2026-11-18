import {Button,InputNumber,Select} from 'antd';
import {useEffect,useRef,useState} from 'react';
import {novelProjectRequest} from '../../../shared/api/novelCreation.js';
import {readSaved,STATUS} from './contracts.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import './SchedulerPanel.css';
import AutomationPanel from './AutomationPanel.jsx';

export default function SchedulerPanel({owner,projectId,sessionId,mode,modelId,models=[],ensureSession,onOpenSession}) {
 const scope=useWorkspaceScope(`${owner}:${projectId}`),cache=`novel:scheduler-start:${encodeURIComponent(owner)}:${encodeURIComponent(projectId||'_')}`;
 const [items,setItems]=useState([]),[selected,setSelected]=useState(''),[ready,setReady]=useState(!projectId),[error,setError]=useState(''),[busy,setBusy]=useState(false),[pending,setPending]=useState(()=>readSaved(cache,null));
 const [legacy,setLegacy]=useState(false);
 const [minutes,setMinutes]=useState(30),[cap,setCap]=useState(10),[failures,setFailures]=useState(3);
 const lock=useRef(false),revision=useRef(0);
 const plan=items.find(v=>v.id===selected)||null,ambiguous=items.length>1&&!plan;
 const running=plan&&['waiting','running','not_started'].includes(plan.status),paused=plan&&['paused','failed'].includes(plan.status);
 const request=(path,body)=>novelProjectRequest(projectId,path,body,{signal:scope.signal});
 const matches=(p,v)=>p.creation?.idempotencyKey==='scheduler:'+v.payload.idempotencyKey;
 function savePending(v){if(v)localStorage.setItem(cache,JSON.stringify(v));else localStorage.removeItem(cache);setPending(v);}
 function accept(rows){setItems(rows);setSelected(old=>rows.some(v=>v.id===old)?old:rows.length===1?rows[0].id:'');setReady(true);}
 async function load(){if(!projectId||lock.current)return;const seq=++revision.current;try{await scope.call(request('/schedules'),r=>{if(seq!==revision.current)return;accept(r.schedules||[]);setError('');});}catch(e){if(scope.isCurrent()&&seq===revision.current){setReady(false);setError(e.message||'调度状态读取失败');}}}
 useEffect(()=>{setItems([]);setSelected('');setLegacy(false);setError('');setReady(!projectId);setPending(readSaved(cache,null));setMinutes(30);setCap(10);setFailures(3);load();const timer=setInterval(load,10000);return()=>clearInterval(timer);},[owner,projectId]);
 useEffect(()=>{if(plan){setMinutes(plan.intervalMinutes);setCap(plan.dailyChapterCap??10);setFailures(plan.maxConsecutiveFailures??3);}},[plan?.id]);
 async function action(fn){if(lock.current)return;lock.current=true;++revision.current;setBusy(true);setError('');try{await fn();}catch(e){if(scope.isCurrent())setError(e.message||'操作失败，原调度状态保留');}finally{lock.current=false;if(scope.isCurrent())setBusy(false);}}
 function acknowledge(row,intent){if(!row?.id||!matches(row,intent))throw Error('启动结果编号不匹配，请核对原请求');++revision.current;savePending(null);setItems(old=>[row,...old.filter(v=>v.id!==row.id)]);setSelected(row.id);setReady(true);}
 async function recover(){if(!pending)return;await action(async()=>{const r=await scope.call(request('/schedules'));if(!r)return;const row=(r.schedules||[]).find(v=>matches(v,pending));if(row){acknowledge(row,pending);return;}savePending({...pending,allowResubmit:true});setError('未找到原启动记录，可以明确重新提交原编号和参数。');});}
 async function start(resubmit=false){if(pending&&!resubmit)return recover();await action(async()=>{
  if(!projectId)throw Error('请先在左侧选择要续写的长篇作品');
  if(['short_original','short_imitation'].includes(mode))throw Error('短篇使用独立成品流程，请选择长篇作品会话后启动调度器');
  if(ambiguous)throw Error('此作品存在多个历史计划，请先选择要查看的调度器');
  let intent=pending;
  if(!intent){if(!Number.isInteger(minutes)||minutes<1||minutes>1440||!Number.isInteger(cap)||cap<1||cap>100||!Number.isInteger(failures)||failures<1||failures>10)throw Error('请填写有效的写作间隔、每日上限和连续失败阈值');if(modelId&&!models.some(v=>v.value===modelId))throw Error('当前文本模型不可用，请先在对话区选择授权模型');const actualSession=sessionId||(await ensureSession?.())?.id;if(!scope.isCurrent())return;if(!actualSession)throw Error('未能建立创作会话，请返回对话重试');intent={payload:{idempotencyKey:crypto.randomUUID(),sessionId:actualSession,modelId,intervalSeconds:minutes*60,dailyChapterCap:cap,maxConsecutiveFailures:failures}};savePending(intent);}
  try{const r=await scope.call(request('/scheduler/start',intent.payload));if(r)acknowledge(r.schedule,intent);}catch(e){if(scope.isCurrent()&&e.status>=400&&e.status<500&&![408,409,429].includes(e.status))savePending(null);throw e;}
 });}
 async function control(verb){await action(async()=>{const r=await scope.call(request(`/schedules/${encodeURIComponent(plan.id)}/${verb}`,{}));if(r){++revision.current;setReady(true);setItems(old=>old.map(v=>v.id===r.schedule.id?r.schedule:v));}});}
 const status=!ready?(error?'读取失败':'读取中'):ambiguous?'请选择历史计划':!plan?'已停止':({waiting:'运行中',running:'运行中',not_started:'运行中',paused:'已暂停',stopped:'已停止',failed:'执行失败',completed:'已完成'})[plan.status]||'状态待核对';
 const next=plan?.eligibleAt||plan?.nextRunAt,remaining=next?Math.max(0,Math.ceil((Date.parse(next)-Date.now())/60000)):0;
 return <section className="novel-panel novel-scheduler-page" aria-label="自动写作调度器"><h3>自动写作调度器</h3>
  {items.length>1&&<label>已有调度计划<Select aria-label="已有调度计划" value={selected||undefined} placeholder="选择已有计划，不会自动新建" options={items.map(v=>({value:v.id,label:`${v.id.slice(-8)} · ${v.status}`}))} onChange={setSelected}/></label>}
  <div className="novel-scheduler-status"><span>状态：<b role="status">{status}</b></span>{plan&&<span>今日 {plan.dailySavedChapters??0} 章 | 累计 {plan.totalSavedChapters??0} 章</span>}</div>
  {error&&<p role="alert">{error}</p>}{projectId&&!ready&&<Button onClick={load}>重新读取状态</Button>}
  {(plan?.consecutiveFailures>0||plan?.lastFailure||plan?.lastError||plan?.configurationError)&&<p role="alert">连续失败 {plan.consecutiveFailures??0} 次{plan.lastFailure||plan.lastError||plan.configurationError?` — ${plan.lastFailure||plan.lastError||plan.configurationError}`:''}</p>}
  {plan?.waitingReason&&<p role="status">{STATUS[plan.waitingReason]||plan.waitingReason}</p>}
  {plan?.sessionId&&onOpenSession&&<Button type="link" onClick={()=>onOpenSession(plan.sessionId)}>查看创作会话</Button>}
  {pending&&<p role="status">启动结果尚未核对，将查询原编号，不会自动重新启动。<Button disabled={busy} loading={busy} onClick={recover}>核对启动结果</Button></p>}
  {running||paused?<div className="novel-scheduler-controls"><Button loading={busy} onClick={()=>control(paused?'resume':'pause')}>{paused?'恢复':'暂停'}</Button><Button danger loading={busy} onClick={()=>control('stop')}>停止</Button></div>:<>
   <div className="novel-scheduler-fields"><label>写作间隔（分钟）<InputNumber controls={false} aria-label="写作间隔（分钟）" min={1} max={1440} value={minutes} onChange={setMinutes} disabled={busy||!!pending}/></label><label>每日上限<InputNumber controls={false} aria-label="每日上限" min={1} max={100} value={cap} onChange={setCap} disabled={busy||!!pending}/></label><label>连续失败暂停阈值<InputNumber controls={false} aria-label="连续失败暂停阈值" min={1} max={10} value={failures} onChange={setFailures} disabled={busy||!!pending}/></label></div>
   <Button type="primary" block loading={busy} disabled={busy||!!projectId&&!ready||ambiguous} onClick={()=>start()}>{pending?'等待核对启动结果':'启动调度器'}</Button>{pending?.allowResubmit&&<Button disabled={busy} onClick={()=>start(true)}>重新提交原编号</Button>}
  </>}
  {running&&Number.isFinite(remaining)&&<p>下次写作：约 {remaining} 分钟后</p>}
  {!!items.length&&<Button type="link" onClick={()=>{setLegacy(v=>!v);if(legacy)load();}}>{legacy?'关闭历史计划管理':'历史计划管理'}</Button>}
  {legacy&&<AutomationPanel projectId={projectId} sessionId={sessionId} mode={mode} input={{modelId}} models={models} ensureSession={ensureSession} manageOnly/>}
 </section>;
}
