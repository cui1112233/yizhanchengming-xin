import {useEffect,useRef,useState} from 'react';
import Taro, {useDidHide,useDidShow} from '@tarojs/taro';
import {apiRequest} from '../../services/api';
import {createNovelJournal} from '../../services/novelRecovery';
import {sessionRevision} from '../../services/sessionBoundary';
import {booktestJobEnded,createBooktestJobs,type BooktestJob} from '../../services/booktestJobs';

export function useBooktestJob(enabled:boolean,onRead:()=>void,request=apiRequest,scopeKey="self"){
 const journal=createNovelJournal({get:k=>Taro.getStorageSync?.(k),set:(k,v)=>Taro.setStorageSync(k,v),remove:k=>Taro.removeStorageSync(k)},scopeKey);
 const [recovery,setRecovery]=useState<any>(()=>enabled?journal.read('process-start'):null),[recent,setRecent]=useState<BooktestJob[]>([]);
 const [job,setJob]=useState<BooktestJob|null>(null),[error,setError]=useState(''),[busy,setBusy]=useState(false);
 const state=useRef({job:null as BooktestJob|null,alive:true,visible:true,epoch:0,starting:false,unknown:Boolean(recovery),revision:sessionRevision(),timer:undefined as ReturnType<typeof setTimeout>|undefined});
 const callbacks=useRef(onRead);callbacks.current=onRead;const client=createBooktestJobs(request);
 const current=()=>enabled&&state.current.alive&&state.current.revision===sessionRevision();
 const clear=()=>{state.current.epoch++;if(state.current.timer)clearTimeout(state.current.timer);state.current.timer=undefined;};
 const accept=(value:BooktestJob)=>{if(!value?.id)throw Error('没有返回有效的测书作业');state.current.job=value;setJob(value);setError('');};
 const read=async(id:string)=>{clear();const epoch=state.current.epoch;try{const value=await client.read(id);if(!current()||!state.current.visible||epoch!==state.current.epoch)return;if(value?.id!==id)throw Error('测书作业编号不匹配');accept(value);callbacks.current();if(!booktestJobEnded(value))state.current.timer=setTimeout(()=>void read(id),2500);}catch(cause:any){if(current()&&epoch===state.current.epoch)setError(cause.message||'作业状态读取失败');}};
 const refresh=async()=>{if(!current())return;if(state.current.job?.id&&!state.current.unknown){await read(state.current.job.id);return;}clear();const epoch=state.current.epoch;try{const result=await client.latest();if(!current()||epoch!==state.current.epoch)return;if(!Array.isArray(result?.jobs))throw Error('最近作业返回格式不正确');setRecent(result.jobs);if(state.current.unknown){setError('上次启动结果尚未确认，请核对最近作业，勿重复处理');return;}const latest=result.jobs[0];if(latest?.id){accept(latest);if(!booktestJobEnded(latest))void read(latest.id);}else setError('');}catch(cause:any){if(current())setError(cause.message||'最近作业读取失败');}};
 const start=async(input:any)=>{if(!current()||state.current.starting) return false;if(state.current.unknown)throw Error('上次启动结果未知，请先核对最近作业，勿重复启动');if(state.current.job&&!booktestJobEnded(state.current.job))throw Error('已有测书作业仍在执行，请先查看任务结果');state.current.starting=true;clear();setBusy(true);setError('');let sent=false;try{const baseline=await client.latest();if(!current())return false;const pending={input,baselineIds:(baseline.jobs||[]).map((j:any)=>j.id),at:new Date().toISOString()};journal.save('process-start',pending);setRecovery(pending);sent=true;const value=await client.start(input);if(!current())return false;if(!value?.id)throw Error('启动结果尚未确认');journal.clear('process-start');setRecovery(null);accept(value);void read(value.id);return true;}catch(cause:any){if(current()){state.current.unknown=sent&&!(cause.statusCode>=400&&cause.statusCode<500&&![408,409,429].includes(cause.statusCode));if(!state.current.unknown){journal.clear('process-start');setRecovery(null);}throw Error(`${cause.message||'启动失败'}${state.current.unknown?'。结果未知，请核对最近作业，勿重复启动。':''}`);}return false;}finally{state.current.starting=false;if(current())setBusy(false);}};
 const stop=async()=>{const id=state.current.job?.id;if(!current()||!id||state.current.starting)return;state.current.starting=true;setBusy(true);try{await client.stop(id);if(current())await read(id);}finally{state.current.starting=false;if(current())setBusy(false);}};
 useEffect(()=>{if(enabled)void refresh();return()=>{state.current.alive=false;clear();};},[]);
 useDidHide(()=>{state.current.visible=false;clear();});useDidShow(()=>{state.current.visible=true;if(enabled)void refresh();});
 const adopt=async(id:string)=>{if(!recent.some(j=>j.id===id))throw Error('请先刷新最近作业');const value=await client.read(id);if(!current())return;if(value.id!==id)throw Error('作业编号不匹配');journal.clear('process-start');state.current.unknown=false;setRecovery(null);accept(value);void read(id);};
 const release=()=>{journal.clear('process-start');state.current.unknown=false;setRecovery(null);setError('');};
 return {job,error,busy,start,refresh,stop,recovery,recent,adopt,release,running:Boolean(job&&!booktestJobEnded(job))};
}
