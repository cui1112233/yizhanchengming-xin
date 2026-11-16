import {useEffect,useRef,useState} from 'react';
import {loadStoryboardPlanProgress} from './comic-screenplay-storyboard-api.js';
import {applyPlanningProgress} from './comic-screenplay-planning-progress.js';

export default function useComicPlanningProgress({projectId,chapterId,requestKey,enabled,onTerminal}){
 const [state,setState]=useState({scope:'',progress:null,error:''});
 const terminal=useRef(onTerminal);terminal.current=onTerminal;
 const scope=JSON.stringify([projectId,chapterId,requestKey]);
 useEffect(()=>{
  if(!enabled||!projectId||!chapterId||!requestKey)return;
  let stopped=false,timer=null,controller=null,previous=null;
  setState({scope,progress:null,error:''});
  async function poll(){
   controller=new AbortController();
   try{
    const result=await loadStoryboardPlanProgress(projectId,chapterId,requestKey,controller.signal);if(stopped)return;
    previous=applyPlanningProgress(previous,result,chapterId,requestKey);
    setState({scope,progress:previous,error:result.available===false&&result.status==='running'?'暂时无法读取规划进度；生成请求不会重复提交。':''});
    if(result.status!=='running'){stopped=true;terminal.current?.(result);return;}
   }catch(err){if(stopped)return;setState({scope,progress:previous,error:err.message||'暂时无法读取规划进度；生成请求不会重复提交。'});}
   if(!stopped)timer=setTimeout(poll,1000);
  }
  void poll();
  return()=>{stopped=true;clearTimeout(timer);controller?.abort();};
 },[scope,enabled,projectId,chapterId,requestKey]);
 return enabled&&state.scope===scope?state:{progress:null,error:''};
}
