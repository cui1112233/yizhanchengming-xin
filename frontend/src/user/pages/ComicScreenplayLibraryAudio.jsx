import {useEffect,useRef,useState} from 'react';
import {apiRequest} from '../../shared/api/client.js';
import OriginalIcon from './ComicScreenplayIcon.jsx';

export function useLibraryAudioPlayer(){
 const [playingKey,setPlayingKey]=useState(''),[error,setError]=useState('');
 const active=useRef({version:0,key:'',audio:null,url:'',controller:null}),alive=useRef(true);
 function stop(){const current=active.current;current.version++;current.key='';current.controller?.abort();current.controller=null;if(current.audio){current.audio.onended=null;current.audio.onerror=null;current.audio.pause();current.audio.currentTime=0;current.audio=null;}if(current.url){URL.revokeObjectURL(current.url);current.url='';}if(alive.current){setPlayingKey('');setError('');}}
 useEffect(()=>{alive.current=true;return()=>{alive.current=false;stop();};},[]);
 async function toggle(key,source){
  if(active.current.key===key){stop();return;}stop();
  if(!source){setError('请先选择参考音频');return;}
  const state=active.current,version=state.version,controller=new AbortController();state.controller=controller;state.key=key;setPlayingKey(key);
  try{
   if(!(source instanceof Blob)&&(!String(source).startsWith('/api/')))throw Error('参考音频地址无效');
   const blob=source instanceof Blob?source:await apiRequest(source,{responseType:'blob',signal:controller.signal,suppressGlobalError:true});
   if(!alive.current||active.current.version!==version)return;
   state.url=URL.createObjectURL(blob);const audio=new Audio(state.url);state.audio=audio;
   audio.onended=()=>{if(active.current.version===version)stop();};
   audio.onerror=()=>{if(active.current.version===version){stop();setError('参考音频播放失败，请检查文件或重新选择');}};
   await audio.play();
   if(!alive.current||active.current.version!==version){audio.pause();audio.currentTime=0;}
  }catch(e){if(alive.current&&active.current.version===version){stop();setError(`参考音频播放失败：${e.message||'请重新选择音频'}`);}}
 }
 return {playingKey,error,toggle,stop};
}

export default function LibraryAudioField({file,saved,disabled,onChoose,player}){
 const input=useRef(null),playing=player.playingKey==='editor';
 return <div className="comic-library-reference-audio"><span>参考音频</span><div><span className="comic-library-reference-audio-name" title={file?.name||saved?.audioFileName||''}>{file?.name||saved?.audioFileName||'未选择参考音频'}</span><button type="button" disabled={disabled} onClick={()=>input.current?.click()}>选择音频</button><button type="button" disabled={disabled} aria-label={playing?'停止参考音频':'试听参考音频'} onClick={()=>void player.toggle('editor',file||saved?.audio)}><OriginalIcon name={playing?'StopOutline':'PlayOutline'} size={14}/>{playing?'停止':'试听'}</button></div><input hidden ref={input} type="file" accept=".mp3,.wav,.m4a,.ogg,.flac,audio/*" aria-label="角色参考音频" disabled={disabled} onChange={event=>{onChoose(event.target.files?.[0]);event.target.value='';}}/><small>支持 MP3、WAV、M4A、OGG、FLAC，最大32MB</small></div>;
}
