import React,{useEffect,useState} from 'react';
import { ConfigProvider, Select } from 'antd';
import * as api from '../../../shared/api/batchFactoryText.js';
import {layoutFields,pageControls} from './textMediaActions.js';
import TextBackgroundGallery from './TextBackgroundGallery.jsx';
import {TEXT_LAYOUT_TEMPLATES} from './textLayoutTemplates.js';

export function useDraftPreview(batchId,bookId,session) {
  const [view,setView]=useState({});
  const key=session?.id;
  useEffect(()=>{setView({});},[key]);
  useEffect(()=>{
    if(!session) return;
    setView({key:session.id,pending:true});
    const controller=new AbortController();
    const timer=setTimeout(async()=>{
      try {
        const {draft,cover,id,rev,docRev}=session;
        const p={expectedRevision:rev,expectedDocumentRevision:docRev,layout:draft.layout,background:draft.background};
        const options={signal:controller.signal,withMetadata:true};
        const {blob,warning,pageCount=1}=cover ? await api.previewTextCover(batchId,bookId,p,options)
          : await api.previewTextPage(batchId,bookId,id,{...p,text:draft.text},options);
        if(controller.signal.aborted)return;
        if(!blob?.size || blob.type!=='image/png')throw Error('未收到有效的预览图片');
        const reflowNote=pageCount>1?`本页内容将分为 ${pageCount} 页，当前预览第 1 页；保存后其余内容会完整顺延。`:'';
        setView({key:id,url:URL.createObjectURL(blob),warning:[reflowNote,warning].filter(Boolean).join(' ')});
      } catch(e) {
        if(!controller.signal.aborted)setView({key:session.id,error:e.message});
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
  return <ConfigProvider theme={{token:{colorBgElevated:'var(--sh-panel)',colorBgContainer:'var(--sh-panel-strong)',colorText:'var(--sh-text)',colorTextPlaceholder:'var(--sh-muted)',colorBorder:'var(--sh-line)',controlItemBgActive:'var(--sh-panel-strong)',controlItemBgHover:'var(--sh-panel-strong)',colorPrimary:'var(--sh-accent)'}}}><fieldset className="text-page-editor" disabled={busy || fileBusy}>
    <legend>{cover?'封面排版':'排版调整 · 仅本页'}</legend>
    {!cover && <small className="text-editor-scope">字号和段距只调整本页内容，放不下的句子会顺延；不会从下一页补字。整书重排请在统一配置或单书配置中调整后重新分页。</small>}
    {!cover && <label>页面正文<textarea aria-label="页面正文" rows={6} value={draft.text} onChange={e=>{const text=e.target.value;onDraft(old=>({...old,text}));}}/></label>}
    {!cover && <div className="text-layout-presets">{TEXT_LAYOUT_TEMPLATES.map(preset=><button key={preset.id} type="button" onClick={()=>onDraft(old=>({...old,layout:{...old.layout,...Object.fromEntries(Object.entries(preset.values).filter(([key])=>layoutFields[key]).map(([key,value])=>[layoutFields[key],value]))},background:{...preset.values.textBackground}}))}>{preset.label}</button>)}<button type="button" onClick={()=>setLayout('textColor','#222222')}>深色文字</button><button type="button" onClick={()=>setLayout('textColor','#FFFFFF')}>浅色文字</button></div>}
    <div className="text-layout-fields">{pageControls.filter(f=>!cover || ['fontId','fontSize','fontWeight','align','textColor'].includes(layoutFields[f.key])).map(f=>{
      const key=layoutFields[f.key], value=draft.layout[key];
      const opts=key==='fontId' && fonts?.length ? fonts.filter(v=>v.available!==false).map(v=>({value:v.id,label:f.options?.find(option=>option.value===v.id)?.label || v.name})) : f.options;
      return <label key={key}>{f.label}{opts ? <Select aria-label={f.label} disabled={busy || fileBusy} value={value} options={opts} getPopupContainer={trigger=>trigger.parentElement} onChange={next=>setLayout(key,typeof f.initial==='number'?Number(next):next)}/>
        : <input disabled={key==='targetCharacters' && draft.layout.paginationMode!=='target'} type={f.type==='color'?'color':'number'} min={f.min} max={f.max} step={f.step || 1} value={f.type==='percent'?Math.round(value*100):value}
            onChange={e=>setLayout(key,f.type==='color'?e.target.value:Number(e.target.value)/(f.type==='percent'?100:1))}/>}</label>;
    })}</div>
    {!cover && <><div className="text-layout-fields">
      <label>背景色<input type="color" value={draft.background.color || '#111318'} onChange={e=>setBackground({kind:'color',color:e.target.value})}/></label>
      <div><span>背景图库</span><TextBackgroundGallery value={draft.background} items={items} loading={bgLoading} disabled={busy || fileBusy} onChange={setBackground}/></div>
    </div>
    <label className="text-upload-input">上传背景图<input type="file" accept="image/png,image/jpeg,image/webp" onChange={e=>{const file=e.target.files[0];e.target.value='';void upload(file);}}/></label></>}
    {bgError && <p role="alert">{bgError}</p>}
    {stale && <p role="alert">此页版本已更新。请复制需要保留的修改，再重新打开排版调整。</p>}
    <div className="text-editor-actions">
      <button type="button" disabled={stale||impact==='none'||(!cover&&!draft.text.trim())} onClick={onSave}>{cover?'保存封面设置':'保存本页'}</button>
      {!cover && <button type="button" title="更新使用整书背景的文字页，保留单页背景和已锁定的图片。" disabled={stale || draft.text!==session.base.text || Object.keys(draft.layout).some(k=>draft.layout[k]!==session.base.layout[k])} onClick={onApplyBackground}>背景应用到整本书</button>}
      <button type="button" onClick={onCancel}>关闭调整</button>
    </div>
  </fieldset></ConfigProvider>;
}
