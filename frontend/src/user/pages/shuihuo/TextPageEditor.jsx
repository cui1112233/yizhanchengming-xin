import React,{useEffect,useState} from 'react';
import * as api from '../../../shared/api/batchFactoryText.js';
import {layoutFields,pageControls} from './textMediaActions.js';

export function useDraftPreview(batchId,bookId,session) {
  const [view,setView]=useState({});
  const key=session?.id;
  useEffect(()=>{setView({});},[key]);
  useEffect(()=>{
    if(!session) return;
    const controller=new AbortController();
    const timer=setTimeout(async()=>{
      try {
        const {draft,cover,id,rev,docRev}=session;
        const p={expectedRevision:rev,expectedDocumentRevision:docRev,layout:draft.layout,background:draft.background};
        const options={signal:controller.signal,withMetadata:true};
        const {blob,warning}=cover ? await api.previewTextCover(batchId,bookId,p,options)
          : await api.previewTextPage(batchId,bookId,id,{...p,text:draft.text},options);
        if(controller.signal.aborted)return;
        if(!blob?.size || blob.type!=='image/png')throw Error('未收到有效的预览图片');
        setView({key:id,url:URL.createObjectURL(blob),warning});
      } catch(e) {
        if(!controller.signal.aborted)setView(last=>({...last,key:session.id,error:e.message}));
      }
    },300);
    return ()=>{clearTimeout(timer);controller.abort();};
  },[batchId,bookId,session]);
  useEffect(()=>()=>{if(view.url)URL.revokeObjectURL(view.url);},[view.url]);
  return key && view.key===key ? view : {};
}

export function TextPageEditor({session,onDraft,onSave,onCancel,onApplyBackground,onUploading,busy,stale,impact,fonts,batchId,bookId}) {
  const {draft,cover}=session;
  const [items,setItems]=useState([]);
  const [bgError,setBgError]=useState('');
  const [bgLoading,setBgLoading]=useState(false);
  const [fileBusy,setFileBusy]=useState(false);
  useEffect(()=>{
    let active=true;
    setBgLoading(true);
    api.listTextBackgrounds(batchId,bookId).then(data=>{if(active)setItems(data?.items || []);})
      .catch(e=>{if(active)setBgError(e.message);}).finally(()=>{if(active)setBgLoading(false);});
    return ()=>{active=false;};
  },[batchId,bookId]);
  const setLayout=(key,value)=>onDraft(old=>({...old,layout:{...old.layout,[key]:value}}));
  const setBackground=background=>onDraft(old=>({...old,background}));
  const upload=async(file)=>{
    if(!file)return;
    onUploading(true);setFileBusy(true);setBgError('');
    try {
      const bg=await api.uploadTextBackground(batchId,bookId,file);
      setItems(list=>[bg,...list.filter(v=>v.id!==bg.id)]);
      setBackground({kind:'library',id:bg.id});
    } catch(e){setBgError(e.message);} finally{onUploading(false);setFileBusy(false);}
  };
  return <fieldset className="text-page-editor" disabled={busy || fileBusy}>
    <legend>{cover?'封面排版':'排版调整'}</legend>
    {!cover && <label>页面正文<textarea aria-label="页面正文" rows={6} value={draft.text} onChange={e=>{const text=e.target.value;onDraft(old=>({...old,text}));}}/></label>}
    <div className="text-layout-fields">{pageControls.filter(f=>!cover || ['fontId','fontSize','fontWeight','align','textColor'].includes(layoutFields[f.key])).map(f=>{
      const key=layoutFields[f.key], value=draft.layout[key];
      const opts=key==='fontId' && fonts?.length ? fonts.filter(v=>v.available!==false).map(v=>({value:v.id,label:v.name})) : f.options;
      return <label key={key}>{f.label}{opts ? <select value={value} onChange={e=>setLayout(key,typeof f.initial==='number'?Number(e.target.value):e.target.value)}>
        {opts.map(o=><option key={o.value} value={o.value}>{o.label}</option>)}</select>
        : <input type={f.type==='color'?'color':'number'} min={f.min} max={f.max} step={f.step || 1} value={f.type==='percent'?Math.round(value*100):value}
            onChange={e=>setLayout(key,f.type==='color'?e.target.value:Number(e.target.value)/(f.type==='percent'?100:1))}/>}</label>;
    })}</div>
    <div className="text-layout-fields">
      <label>背景色<input type="color" value={draft.background.color || '#111318'} onChange={e=>setBackground({kind:'color',color:e.target.value})}/></label>
      <label>背景图库<select aria-label="背景图库" value={draft.background.id || ''} onChange={e=>setBackground(e.target.value ? {kind:'library',id:e.target.value}:{kind:'color',color:'#111318'})}>
        <option value="">{bgLoading?'正在读取…':'纯色背景'}</option>{items.map(i=><option key={i.id} value={i.id}>{i.name}</option>)}</select></label>
    </div>
    <label className="text-upload-input">上传背景图<input type="file" accept="image/png,image/jpeg,image/webp" onChange={e=>{const file=e.target.files[0];e.target.value='';void upload(file);}}/></label>
    {bgError && <p role="alert">{bgError}</p>}
    {stale && <p role="alert">此页版本已更新。请复制需要保留的修改，再重新打开排版调整。</p>}
    <div className="text-editor-actions">
      <button type="button" disabled={stale||impact==='none'||(!cover&&!draft.text.trim())} onClick={onSave}>{cover?'保存封面设置':'保存本页'}</button>
      {!cover && <button type="button" title="更新使用整书背景的文字页，保留单页背景和已锁定的图片。" disabled={stale || draft.text!==session.base.text || Object.keys(draft.layout).some(k=>draft.layout[k]!==session.base.layout[k])} onClick={onApplyBackground}>背景应用到整本书</button>}
      <button type="button" onClick={onCancel}>关闭调整</button>
    </div>
  </fieldset>;
}
