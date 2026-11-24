import Taro,{useDidHide} from '@tarojs/taro';
import {useEffect,useRef,useState} from 'react';
import {apiRequest} from '../../services/api';
import {onSessionInvalidated,sessionRevision} from '../../services/sessionBoundary';
import {authStorageKey} from '../../config/localAuth';
import {createNovelJournal,type NovelJournal} from '../../services/novelRecovery';
export default function useScriptDraft(value:any,historyId:string,restore:(value:any)=>boolean|void){
 const [recovery,setRecovery]=useState<any>(null),[status,setStatus]=useState('正在确认草稿账号'),[ready,setReady]=useState(false);
 const epoch=useRef(0);
 const context=useRef({journal:null as NovelJournal|null,alive:true,revision:sessionRevision(),recovery:null as any}),latest=useRef(value);latest.current=value;
 const save=()=>{const c=context.current;if(!c.alive||c.revision!==sessionRevision()||!c.journal||c.recovery||(!latest.current.state?.novelText?.trim()&&!latest.current.result))return;try{c.journal.save('draft',{...latest.current,at:new Date().toISOString()});setStatus('草稿已保存到本机');}catch(e:any){setStatus(e.message||'草稿保存失败，请复制保留正文');}};
 useEffect(()=>{const ticket=++epoch.current,c=context.current;c.alive=true;c.revision=sessionRevision();c.journal=null;c.recovery=null;setReady(false);setRecovery(null);void apiRequest<any>('/api/member/me').then(r=>{if(ticket!==epoch.current||!c.alive||c.revision!==sessionRevision())return;if(!r?.member?.username)throw Error('账号未确认，暂未保存草稿');c.journal=createNovelJournal({get:k=>Taro.getStorageSync(k),set:(k,v)=>Taro.setStorageSync(k,v),remove:k=>Taro.removeStorageSync(k)},authStorageKey(`script:${r.member.username}:${historyId||'new'}`));const prior=c.journal.read('draft');if(prior?.state?.novelText||prior?.result){c.recovery=prior;setRecovery(prior);setStatus('发现本机草稿，可恢复或保留当前内容');}else setStatus('修改后会自动保存到本机');setReady(true);}).catch(e=>{if(ticket===epoch.current&&c.alive)setStatus(e.message||'草稿读取失败，请复制保留正文');});const off=onSessionInvalidated(()=>{c.revision=-1;c.journal=null;c.recovery=null;setRecovery(null);setReady(false);setStatus('账号已改变，旧草稿已隔离');});return()=>{save();epoch.current++;c.alive=false;c.journal=null;off();};},[historyId]);
 useEffect(()=>{if(!ready||recovery)return;const timer=setTimeout(save,350);return()=>clearTimeout(timer);},[value.state,value.result,value.savedHistoryId,ready,recovery]);
 useDidHide(save);
 return {recovery,status,restore(){const prior=context.current.recovery;if(!prior)return;if(restore(prior)===false)return;context.current.recovery=null;setRecovery(null);setStatus('已恢复本机草稿；原文配音需重新生成');},keepCurrent(){context.current.recovery=null;setRecovery(null);save();}};
}
