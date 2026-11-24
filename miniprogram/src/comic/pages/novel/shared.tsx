import Taro,{useDidHide,useDidShow} from '@tarojs/taro';
import {Button,Input,Picker,Text,Textarea,View} from '@tarojs/components';
import {useEffect,useRef,useState} from 'react';
import {sessionRevision} from '../../../services/sessionBoundary';
import type {NovelRecord} from '../../../services/novelCreationInputs';
export function useNovelOperations(){
  const [busy,setBusy]=useState(''),[error,setError]=useState('');const alive=useRef(true),lock=useRef(false),session=useRef(sessionRevision());
  const current=()=>alive.current&&session.current===sessionRevision();
  useEffect(()=>()=>{alive.current=false;},[]);
  const run=async(label:string,work:()=>Promise<void>)=>{if(lock.current||!current())return;lock.current=true;setBusy(label);setError('');try{await work();}catch(cause:any){if(current())setError(cause.statusCode===409?'数据已变化，当前草稿保留，请刷新核对后继续。':cause.message||cause.errMsg||'操作失败，请重试');}finally{lock.current=false;if(current())setBusy('');}};
  return {busy,error,setError,run,current,lock};
}
export function useNovelRead(reader:()=>Promise<any>,accept:(value:any)=>void,watch:(value:any)=>boolean,onError:(cause:any)=>void){
  const callbacks=useRef({reader,accept,watch,onError});callbacks.current={reader,accept,watch,onError};
  const alive=useRef(true),visible=useRef(true),epoch=useRef(0),timer=useRef<ReturnType<typeof setTimeout>>();
  const stop=()=>{epoch.current++;if(timer.current)clearTimeout(timer.current);timer.current=undefined;};
  const refresh=async()=>{stop();const ticket=epoch.current;const session=sessionRevision();try{const value=await callbacks.current.reader();if(!alive.current||!visible.current||ticket!==epoch.current||session!==sessionRevision())return;callbacks.current.accept(value);if(callbacks.current.watch(value))timer.current=setTimeout(()=>void refresh(),2500);}catch(cause){if(alive.current&&ticket===epoch.current&&session===sessionRevision())callbacks.current.onError(cause);}};
  useEffect(()=>{void refresh();return()=>{alive.current=false;stop();};},[]);
  useDidHide(()=>{visible.current=false;stop();});useDidShow(()=>{visible.current=true;void refresh();});
  return refresh;
}
export function NovelField({label,value,onChange,multiline=false,number=false,disabled=false,id,hint}:{label:string;value:any;onChange:(value:string)=>void;multiline?:boolean;number?:boolean;disabled?:boolean;id?:string;hint?:string}){
  return <View className="nw-field"><Text className="nw-label">{label}</Text>{multiline?<Textarea id={id} className="nw-input nw-textarea" maxlength={-1} value={String(value??'')} disabled={disabled} onInput={event=>onChange(event.detail.value)} placeholder={hint||''}/>:<Input id={id} className="nw-input" maxlength={-1} type={number?'number':'text'} value={String(value??'')} disabled={disabled} onInput={event=>onChange(event.detail.value)} placeholder={hint||''}/>}</View>;
}
export function NovelSelect({label,items,value,onChange,disabled=false}:{label:string;items:NovelRecord[];value:string;onChange:(id:string)=>void;disabled?:boolean}){
  return <Picker disabled={disabled||!items.length} range={items.map(item=>item.name||item.title||item.id)} value={Math.max(0,items.findIndex(item=>item.id===value))} onChange={event=>{const item=items[Number(event.detail.value)];if(item)onChange(item.id);}}><View className="nw-select"><Text className="nw-label">{label}</Text><Text>{(items.find(item=>item.id===value)?.name||items.find(item=>item.id===value)?.title)||(items.length?'请选择':'暂无可用选项')}</Text></View></Picker>;
}
export function NovelNotice({busy,error}:{busy:string;error:string}){return <>{busy?<Text className="nw-caption">正在{busy}…</Text>:null}{error?<Text className="nw-error">{error}</Text>:null}</>;}
export async function novelConfirm(title:string,content:string){return (await Taro.showModal({title,content,confirmText:'确认继续',cancelText:'暂不继续'})).confirm;}
export const THINKING=[{id:'fast',name:'快速'},{id:'standard',name:'标准思考'},{id:'deep',name:'深度思考'}];
export function NovelBack({title,onBack,disabled=false}:{title:string;onBack:()=>void;disabled?:boolean}){return <View className="nw-heading"><View><Text className="nw-title">{title}</Text></View><Button className="nw-link" disabled={disabled} onClick={onBack}>返回作品库</Button></View>;}
