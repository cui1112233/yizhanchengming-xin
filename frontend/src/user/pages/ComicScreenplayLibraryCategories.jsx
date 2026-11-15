import {useEffect,useRef,useState} from 'react';
import {Select} from 'antd';
import {listComicAssetLibraryCategories,changeComicAssetLibraryCategory} from '../../shared/api/comicScreenplayAssetLibrary.js';

// This component belongs to one editor session. Closing it never undoes confirmed category writes.
export default function ComicScreenplayLibraryCategories({value,onChange,disabled,onBusyChange}) {
  const [state,setState]=useState(null),[search,setSearch]=useState(''),[busy,setBusy]=useState(false),[error,setError]=useState(''),[pending,setPending]=useState(null);
  const live=useRef(false),lock=useRef(false),epoch=useRef(0),controller=useRef(null),selected=useRef(value),busyCallback=useRef(onBusyChange);
  selected.current=value;busyCallback.current=onBusyChange;
  function reportBusy(value){setBusy(value);busyCallback.current?.(value);}
  useEffect(()=>{live.current=true;lock.current=false;void load();return()=>{live.current=false;epoch.current++;controller.current?.abort();busyCallback.current?.(false);};},[]);
  async function load(){
    if(lock.current)return;lock.current=true;reportBusy(true);const version=++epoch.current,abort=new AbortController();controller.current=abort;
    try{const next=await listComicAssetLibraryCategories({signal:abort.signal});if(live.current&&version===epoch.current){setState(next);setError('');setPending(null);}}
    catch(e){if(live.current&&version===epoch.current)setError(`分类读取失败：${e.message||'请重试'}`);}
    finally{if(live.current&&version===epoch.current){lock.current=false;reportBusy(false);}}
  }
  async function change(action,name){
    if(lock.current||!state||disabled)return;lock.current=true;reportBusy(true);setError('');const version=++epoch.current,abort=new AbortController();controller.current=abort;
    const current=()=>live.current&&version===epoch.current;
    try{
      let next;
      try{next=await changeComicAssetLibraryCategory({action,name,expectedRevision:state.revision},{signal:abort.signal});}
      catch(e){if(e.status!==409)throw e;const latest=await listComicAssetLibraryCategories({signal:abort.signal});if(!current())return;setState(latest);next=await changeComicAssetLibraryCategory({action,name,expectedRevision:latest.revision},{signal:abort.signal});}
      if(current()){setState(next);setPending(null);setSearch('');if(action==='add')onChange(name);else if(selected.current===name)onChange('');}
    }catch(e){if(current()){setError(`分类保存失败：${e.message||'请重试'}`);setPending({action,name});}}
    finally{if(current()){lock.current=false;reportBusy(false);}}
  }
  const names=state?[...new Set([...state.builtin_categories,...state.custom_categories])]:[];
  const options=names.map(name=>({value:name,label:name}));
  function createOnEnter(event){
    if(event.key!=='Enter')return;
    if(event.nativeEvent?.isComposing||event.isComposing||event.keyCode===229){event.stopPropagation();return;}
    if(!search.trim()||names.includes(search))return;
    event.preventDefault();event.stopPropagation();
    if(!lock.current&&!disabled&&state)void change('add',search);
  }
  return <div className="comic-library-category-field" style={{display:'flex',flexDirection:'column',gap:6,fontSize:12}}><span>分类</span><Select aria-label="档案分类" style={{width:'100%'}} showSearch allowClear virtual={false} onInputKeyDown={createOnEnter} disabled={disabled||busy} loading={busy} value={value||undefined} searchValue={search} onSearch={text=>{if(!lock.current)setSearch(text);}} onChange={name=>{setSearch('');onChange(name||'');}} options={options} optionRender={option=><div style={{display:'flex',alignItems:'center',justifyContent:'space-between',gap:16}}><span>{option.value}</span>{state.custom_categories.includes(option.value)&&!state.builtin_categories.includes(option.value)?<button type="button" aria-label={`删除分类${option.value}`} style={{border:0,background:'transparent',color:'#f87171',fontSize:12,cursor:'pointer',padding:'0 4px'}} disabled={disabled||busy} onMouseDown={event=>{event.preventDefault();event.stopPropagation();}} onClick={event=>{event.preventDefault();event.stopPropagation();void change('delete',option.value);}}>删除</button>:null}</div>} popupRender={menu=><>{menu}{search.trim()&&!names.includes(search)?<button type="button" disabled={disabled||busy||!state} onMouseDown={event=>event.preventDefault()} onClick={()=>void change('add',search)}>{`创建分类“${search}”`}</button>:null}</>}/><small>分类独立保存，取消档案编辑不会撤销分类修改。</small>{error?<div role="alert">{error}<button type="button" disabled={busy||disabled} onClick={()=>pending?void change(pending.action,pending.name):void load()}>{pending?'重试分类操作':'重新读取分类'}</button></div>:null}</div>;
}
