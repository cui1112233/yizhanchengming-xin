import React, { useEffect, useImperativeHandle, useRef, useState } from 'react';
import { createTextCollection, getTextCollection, listTextCollections, listTextPages, prepareTextBook } from '../../../shared/api/batchFactoryText.js';
import * as api from '../../../shared/api/batchFactoryText.js';
import { getBatch, saveBookOverride, publishTextCollection, getTextPublishStatus } from '../../../shared/api/batchFactoryV12.js';
import AuthenticatedTextPageImage from './AuthenticatedTextPageImage.jsx';
import { canCreateTextCollection, textBookCompletion, textPagePresentation, textPageDirtyImpact } from './textMediaState.js';
import { initialPageDraft,pageChanges,splitChoices,uploadLabel } from './textMediaActions.js';
import { TextPageEditor,useDraftPreview } from './TextPageEditor.jsx';
import { textPublishImageProgress, textPublishStatus } from './batchFactoryBookState.js';
import { textRetryConfirmation, textUploadControl } from './batchFactoryNovelListHelpers.js';
import { updateTextProductionSettings } from './textProductionConfig.js';
import './text-media-panel.css';

const tabs = [['cover', '封面'], ['pages', '文字页'], ['collections', '文本合集']];
const empty = { pages: [], artifacts: [], document: null };
const pageNumber = value => String(value).padStart(2, '0');

// Keep polling tied to the lifetime of this book. In-flight reads cannot replace a newer snapshot.
function useTextLibrary(batchId, bookId, refresh) {
  const [snapshot, setSnapshot] = useState(empty);
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  useEffect(() => {
    const controller = new AbortController();
    const options = { signal: controller.signal, suppressGlobalError: true };
    let timer;
    async function read(poll = false) {
      try {
        const [next, list] = await Promise.all([
          listTextPages(batchId, bookId, options),
          poll ? null : listTextCollections(batchId, bookId, options)
        ]);
        if (controller.signal.aborted) return;
        setSnapshot(next);
        if (list) setCollections(list.items || []);
        setError('');
        if (next.pages?.some(page => page.status === 'rendering')) timer = setTimeout(() => read(true), 2000);
      } catch (e) {
        if (!controller.signal.aborted) setError(e.message || '文字页读取失败');
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    setLoading(true);
    void read();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [batchId, bookId, refresh]);
  return { snapshot, collections, loading, error };
}

function CollectionChoices({ collections, current, choose }) {
  return <div className="text-media-collections">{collections.length ? collections.map((item, index) =>
    <button type="button" key={item.id} aria-pressed={item.id === current?.id} onClick={() => choose(item.id)}>
      文本合集 {pageNumber(collections.length - index)}
    </button>) : <p>暂无文本合集</p>}</div>;
}

function CandidateList({ items, selectedId, activeId, onSelect }) {
  return <section><h4>候选版本</h4>{items.length ? <div className="text-media-candidates">{items.map(item =>
    <button type="button" key={item.id} aria-pressed={selectedId === item.id} onClick={() => onSelect(item.id)}>
      <AuthenticatedTextPageImage passive contentUrl={item.contentUrl} alt={`版本 ${item.renderVersion}`} />
      <small>{item.sourceKind === 'user_upload' ? '上传' : '系统'} · v{item.renderVersion}{item.id === activeId ? ' · 当前' : ''}</small>
    </button>)}</div> : <p>暂无成品图</p>}</section>;
}

function TextBookLibrary({ batchId, book, settings = {}, initialPageId, initialTab = 'pages', onSaved, onClose, onUpload, closeGuard }) {
  const [refresh, setRefresh] = useState(0);
  const [tab, setTab] = useState(initialTab);
  const [pageId, setPageId] = useState(initialPageId);
  const [selectedCandidate, selectCandidate] = useState('');
  const [chosenCollection, chooseCollection] = useState('');
  const [collectionDetail, setCollectionDetail] = useState(null);
  const [receipt, setReceipt] = useState(null);
  const [collectionImage, setCollectionImage] = useState('');
  const [expanded, setExpanded] = useState(false);
  const [busy, setBusy] = useState('');
  const [actionError, setActionError] = useState('');
  const [session,setSession]=useState(null);
  const [start,setStart]=useState(1);
  const [end,setEnd]=useState(null);
  const [split,setSplit]=useState('');
  const actionLock=useRef(false);
  const backgroundUploading=useRef(false);
  const [savedSettings,setSavedSettings]=useState(null);
  const bookSettings=savedSettings?.revision>=book.revision?savedSettings:{revision:book.revision,patch:book.settingsState?.patch};
  const live=useDraftPreview(batchId,book.id,session);
  const { snapshot, collections, loading, error } = useTextLibrary(batchId, book.id, refresh);
  const doc = snapshot.document;
  const pages = snapshot.pages || [];
  const artifacts = snapshot.artifacts || [];
  const page = pages.find(item => item.id === pageId) || pages[0];
  const pageIndex = pages.indexOf(page);
  const state = textPagePresentation(page, artifacts);
  const completion = textBookCompletion(pages);
  const collection = collections.find(item => item.id === chosenCollection) || collections[0];
  useEffect(() => {
    if (tab !== 'collections' || !collection?.id) return;
    const controller = new AbortController();
    getTextCollection(batchId, book.id, collection.id, { signal: controller.signal, suppressGlobalError: true })
      .then(value => { if (!controller.signal.aborted) setCollectionDetail(value); })
      .catch(e => { if (!controller.signal.aborted) setActionError(e.message); });
    return () => controller.abort();
  }, [tab, collection?.id, batchId, book.id, refresh]);
  const publishing = busy === 'publish';
  useEffect(() => {
    if (!collection?.id || onUpload) return;
    const controller = new AbortController();
    let timer;
    async function read() {
      try {
        const status = await getTextPublishStatus(batchId, book.id, collection.id, {
          signal:controller.signal, suppressGlobalError:true
        });
        if (controller.signal.aborted) return;
        if (status?.publish?.requestKey) textRetryConfirmation(batchId, book.id, collection.id).finish(status.publish.requestKey);
        setReceipt({id:collection.id, status});
        const state = textPublishStatus(status?.publish);
        if (['uploading','pending'].includes(state) || (publishing && !state)) timer = setTimeout(read, 3000);
      } catch (e) {
        if (!controller.signal.aborted) {
          setReceipt(previous => ({id:collection.id, status:previous?.id === collection.id ? previous.status : null, error:e.message || '上传状态读取失败'}));
          timer = setTimeout(read, 3000);
        }
      }
    }
    void read();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [batchId, book.id, collection?.id, refresh, onUpload, publishing]);
  const selectedReceipt = receipt?.id === collection?.id ? receipt : null;
  const publishRecord = selectedReceipt?.status?.publish;
  const publishState = textPublishStatus(publishRecord);
  const upload = onUpload && textPublishStatus(collection) !== 'uploading'
    ? {label:uploadLabel(collection), force:uploadLabel(collection) === '再次上传', disabled:!collection}
    : textUploadControl(collection, onUpload ? {publish:collection} : selectedReceipt?.status);
  async function publish() {
    if (upload.disabled || actionLock.current) return;
    let requestKey = upload.requestKey;
    let confirmation;
    if (upload.force) {
      try {
        confirmation = textRetryConfirmation(batchId, book.id, collection.id);
        requestKey = await confirmation.authorize(
          () => window.confirm('这个文本合集已上传，确认再次上传？'),
          () => window.confirm('上次重传尚未收到回复，要继续同一次上传吗？')
        );
        if (!requestKey) return;
      } catch (error) {
        setActionError(error.message || '无法保存上传确认编号');
        return;
      }
    }
    void run('publish', async () => {
      const payload = {force:upload.force, ...(requestKey && {requestKey})};
      if (onUpload) await onUpload(collection, payload);
      else {
        const status = await publishTextCollection(batchId, book.id, collection.id, payload);
        setReceipt({id:collection.id, status});
      }
      confirmation?.finish(requestKey);
    });
  }
  const currentCollection = collectionDetail?.id === collection?.id ? collectionDetail : collection;
  const collectionItems = currentCollection?.items || [];
  const candidates = artifacts.filter(a => tab === 'cover' ? a.kind === 'cover' : a.kind === 'page' && a.pageId === page?.id);
  const activeId = tab === 'cover' ? doc?.coverArtifactId : state.artifactId;
  const display = candidates.find(a => a.id === selectedCandidate) || candidates.find(a => a.id === activeId);
  const preview = tab === 'collections'
    ? collectionItems.find(i => i.artifactId === collectionImage)?.contentUrl || collectionItems[0]?.contentUrl
    : display?.contentUrl;
  const title = book.title || '未命名作品';
  const selectedIds=pages.slice(start-1,end ?? pages.length).map(p=>p.id);
  const canCompose = canCreateTextCollection(pages, selectedIds, doc, artifacts);
  const background = page?.background || doc?.configSnapshot?.background || settings.textBackground;
  const reload = () => setRefresh(value => value + 1);
  const impact=session?textPageDirtyImpact(session.base,session.draft):'none';
  const discard=()=> {
    if(impact!=='none' && !window.confirm('放弃本页未保存的修改？'))return false;
    setSession(null);return true;
  };
  const move = (newId) => { if(!discard())return false;setPageId(newId); selectCandidate('');setSplit('');return true; };
  function requestClose() {
    if(!actionLock.current && !backgroundUploading.current && discard())onClose?.();
  }
  useImperativeHandle(closeGuard,()=>requestClose);
  const cover=tab==='cover';
  const locked=cover?doc?.coverLocked:page?.locked;
  const targetRevision=cover?doc?.revision:page?.revision;
  const breaks=splitChoices(page,snapshot.sentenceBreaks);
  const stale=!!session && (session.docRev!==doc?.revision || session.rev!==targetRevision || (!cover && session.id!==page?.id) || session.bookSettings.revision!==bookSettings.revision);
  const edit=()=>{
    const base=initialPageDraft(doc,page,cover);
    setSession({id:cover?'cover':page.id,cover,base,draft:base,rev:targetRevision,docRev:doc.revision,bookSettings});
    selectCandidate('');setActionError('');
  };
  async function run(action, callback) {
    if (actionLock.current) return;
    actionLock.current=true;
    setBusy(action); setActionError('');
    try { await callback(); reload(); await onSaved?.(); }
    catch (e) { setActionError(e.message || '操作失败'); reload(); }
    finally { actionLock.current=false;setBusy(''); }
  }
  async function saveBookPatch(changes) {
    const update=updateTextProductionSettings(session.bookSettings.patch,changes);
    try {
      const saved=await saveBookOverride(batchId,book.id,{...update,expectedRevision:session.bookSettings.revision});
      setSavedSettings({revision:saved.revision,patch:saved.patch});
    } catch(e) {
      if(e.status===409) {
        const response=await getBatch(batchId);
        const current=(response.batch || response).books?.find(b=>b.id===book.id);
        if(current)setSavedSettings({revision:current.revision,patch:current.settingsState?.patch});
        throw Object.assign(Error('单书配置已变化。请重新打开排版调整后再保存。'),{status:409});
      }
      throw e;
    }
    try {
      return await prepareTextBook(batchId,book.id,{expectedRevision:session.docRev});
    } catch(e) {
      setSession(null);
      await onSaved?.();
      throw Error('单书设置已保存，更新页面失败：'+e.message+'。请点击更新文字页。');
    }
  }
  async function saveEdit() {
    const changes=pageChanges(session.base,session.draft,session.rev,session.docRev);
    if(cover) {
      const config={};
      if(changes.layout)for(const key of ['fontId','fontSize','fontWeight','align','textColor']){
        if(changes.layout[key]!==session.base.layout[key])config['textCover.layout.'+key]=changes.layout[key];
      }
      const result=await saveBookPatch(config);
      setSession(null);selectCandidate('');
      if(result.document && !result.document.coverLocked) {
        try {selectCandidate((await api.renderTextCover(batchId,book.id,{expectedRevision:result.document.revision})).id);}
        catch(e){throw Error('封面设置已保存，生成失败：'+e.message);}
      }
    } else {
      const payload=impact==='reflow'?{...changes,text:session.draft.text,reflow:true}:changes;
      const result=await api.patchTextPage(batchId,book.id,session.id,payload);
      if(result.pages?.[pageIndex])setPageId(result.pages[pageIndex].id);
      setSession(null);selectCandidate('');
      const updated=Array.isArray(result.affectedPageIds)?result.pages.filter(p=>result.affectedPageIds.includes(p.id)):
        result.pages?.filter(p=>!p.locked && p.status!=='ready') || [result];
      await renderUpdated(updated);
    }
  }
  async function renderUpdated(updated) {
    const failures=[];
    for (const item of updated) {
      if(!item.revision || !item.id || item.locked || item.status==='ready')continue;
      try {await api.renderTextPage(batchId,book.id,item.id,{expectedRevision:item.revision});}
      catch(e){failures.push('第 '+item.ordinal+' 页：'+e.message);}
    }
    if(failures.length)throw Error('排版已保存，以下页面需重新生成：'+failures.join('；'));
  }
  async function applyBookBackground() {
    const result=await saveBookPatch({textBackground:session.draft.background});
    setSession(null);selectCandidate('');
    await renderUpdated((result.pages || []).filter(item=>!item.background));
  }
  const renderCurrent=()=>cover?api.renderTextCover(batchId,book.id,{expectedRevision:doc.revision}):
    api.renderTextPage(batchId,book.id,page.id,{expectedRevision:page.revision});
  function restoreCurrent(){
    if(!window.confirm(cover?'恢复系统生成封面并重新生成？':'解除本页锁定并重新生成系统图片？'))return;
    void run('restore',async()=>{
      const next=cover?await api.restoreTextCover(batchId,book.id,{expectedRevision:doc.revision}):
        await api.restoreTextPage(batchId,book.id,page.id,{expectedRevision:page.revision});
      selectCandidate('');
      if(cover)await api.renderTextCover(batchId,book.id,{expectedRevision:next.document.revision});
      else await api.renderTextPage(batchId,book.id,page.id,{expectedRevision:next.pages.find(p=>p.id===page.id).revision});
    });
  }
  function uploadCurrent(file){
    if(!file)return;
    void run('upload',async()=>{
      const result=cover ? await api.uploadTextCover(batchId,book.id,file,doc.revision) :
        await api.uploadTextPageArtifact(batchId,book.id,page.id,file,page.revision);
      selectCandidate(result.id);
    });
  }
  function changePages(merge) {
    void run('paginate',async()=>{
      const next=merge?await api.mergeTextPage(batchId,book.id,page.id,{expectedRevision:doc.revision}):
        await api.splitTextPage(batchId,book.id,page.id,{expectedRevision:doc.revision,offset:Number(split||breaks[0]?.offset)});
      setPageId(next.pages[pageIndex].id);setSplit('');selectCandidate('');
    });
  }
  async function compose() {
    const result = await createTextCollection(batchId, book.id, {
      expectedRevision:doc.revision, versionId:doc.activeVersionId, paginationVersion:doc.paginationVersion,
      pageIds:selectedIds
    });
    chooseCollection(result.id); setCollectionDetail(result); setCollectionImage(''); setTab('collections');
  }
  return <div className="text-media-panel" aria-label="文本媒体库">
    <section className="text-media-preview" aria-label="成品预览">{live.url?<img src={live.url} alt="排版预览"/>:preview
      ? <AuthenticatedTextPageImage contentUrl={preview} alt={tab === 'pages' ? `文字页 ${page?.ordinal}` : tab === 'cover' ? '文本封面' : '文本合集预览'} />
      : <div className="text-media-empty">{loading ? '正在读取文本页…' : tab === 'cover' ? '暂无封面图' : tab === 'collections' ? '暂无合集预览' : state.error || state.label}</div>}
      {(live.error||live.warning) && <div className="text-preview-note" role="status">{live.error||live.warning}</div>}</section>
    <aside className="text-media-tools">
      <header><div><small>媒体库 · 文本</small><h3>{title}</h3></div>
        <button type="button" aria-label="关闭文本媒体库" disabled={!!busy} onClick={requestClose}>×</button>
        <div className="text-media-progress"><span>完成度 {completion.ready} / {completion.total} 页</span><progress max="100" value={completion.percent}/></div>
      </header>
      <nav role="tablist" aria-label="文本媒体视图">{tabs.map(([key,label]) =>
        <button key={key} type="button" role="tab" aria-label={label} disabled={!!busy} aria-selected={tab === key} onClick={() => {if(!discard())return;setTab(key); selectCandidate('');setActionError('');}}>
          {label}{key === 'pages' ? ' '+pages.length : key === 'collections' ? ' '+collections.length : ''}</button>)}</nav>
      <div className="text-media-scroll" role="tabpanel">
        {(error || actionError) && <div role="alert">{error || actionError} <button type="button" onClick={() => {setActionError('');reload();}}>刷新</button></div>}
        {selectedReceipt?.error && <div role="alert">{selectedReceipt.error} <button type="button" onClick={reload}>刷新上传状态</button></div>}
        {publishRecord && <div role="status">{selectedReceipt.status.published ? '已上传' : ({
          uploading:publishRecord.message || '正在上传合集图片', pending:publishRecord.message || '已提交，等待上传确认',
          failed:'上传失败', upload_incomplete:'图片上传未完成'
        })[publishState]} {textPublishImageProgress(publishRecord)} {publishRecord.errorMessage || publishRecord.error}</div>}
        {tab === 'collections' ? <>
          <CollectionChoices collections={collections} current={collection} choose={id=>{chooseCollection(id);setCollectionImage('');}} />
          {collection && <section><h4>合集内容</h4><div className="text-media-page-choices">
            {collectionItems.map(i=><button type="button" key={i.artifactId} onClick={()=>setCollectionImage(i.artifactId)}>文字页 {i.ordinal}</button>)}</div></section>}
        </> : <>
          <CandidateList items={candidates} selectedId={display?.id} activeId={activeId} onSelect={id=>{if(!busy && discard())selectCandidate(id);}}/>
          {tab === 'pages' ? <>
            <section><h4>本页背景</h4>{background?.color ? <span><i className="text-media-swatch" style={{backgroundColor:background.color}}/>{background.color}</span> : background?.id ? '背景图片' : '继承整书背景'}</section>
            <section><h4>页面内容</h4><p className="text-media-copy">{page?.text || '尚未分页'}</p></section>
          </> : <section><h4>封面标题</h4><p>{doc?.configSnapshot?.cover?.title || title}</p>{doc?.coverLocked && <small>上传成品封面</small>}</section>}
          {doc && (page || cover) && !session && <section className="text-editor-actions">
            <button type="button" disabled={!!busy || locked} onClick={edit}>{cover?'调整封面':'排版调整'}</button>
            <button type="button" disabled={!!busy || locked} onClick={()=>run('render',async()=>selectCandidate((await renderCurrent()).id))}>{cover?'生成封面':'生成本页'}</button>
            <label className="text-upload-input">{cover?'上传成品封面':'上传成品页'}<input type="file" disabled={!!busy} accept="image/png,image/jpeg,image/webp" aria-label={cover?'上传成品封面':'上传成品页'} onChange={e=>{const file=e.target.files[0];e.target.value='';uploadCurrent(file);}}/></label>
            {locked && <button type="button" disabled={!!busy} onClick={restoreCurrent}>恢复系统生成</button>}
            {!cover && !locked && <>
              {breaks.length>0 && <label>句末<select aria-label="拆页句末" value={split || breaks[0].offset} onChange={e=>setSplit(e.target.value)}>{breaks.map(b=><option key={b.offset} value={b.offset}>{b.label}</option>)}</select></label>}
              <button type="button" disabled={!!busy || !breaks.length} onClick={()=>changePages(false)}>在完整句子后拆页</button>
              <button type="button" disabled={!!busy || !pages[pageIndex+1] || pages[pageIndex+1].locked} onClick={()=>changePages(true)}>与下一页合并</button>
            </>}
          </section>}
          {session && <TextPageEditor key={session.id} session={session} fonts={snapshot.fonts} batchId={batchId} bookId={book.id} busy={!!busy} stale={stale} impact={impact} onUploading={value=>{backgroundUploading.current=value;}}
            onDraft={fn=>setSession(s=>s?.id===session.id?{...s,draft:fn(s.draft)}:s)} onCancel={discard}
            onSave={()=>run('save',saveEdit)} onApplyBackground={()=>run('background',applyBookBackground)}/>}
          {!loading && !error && !session && <button type="button" disabled={!!busy} onClick={()=>run('prepare',()=>prepareTextBook(batchId,book.id,doc?{expectedRevision:doc.revision}:{}))}>{doc?'更新文字页':'准备文字页'}</button>}
          {doc && !session && <button type="button" disabled={!!busy} onClick={()=>{if(window.confirm('按当前排版重新分配页数？手动拆页位置可能变化。'))void run('reflow',()=>prepareTextBook(batchId,book.id,{expectedRevision:doc.revision,reflow:true}));}}>重新分页</button>}
        </>}
        {tab==='collections' && pages.length>0 && <section className="text-layout-fields"><label>合集起始页<input type="number" min="1" max={pages.length} value={start} onChange={e=>setStart(Math.max(1,Math.min(pages.length,Number(e.target.value))))}/></label>
          <label>合集结束页<input type="number" min="1" max={pages.length} value={end ?? pages.length} onChange={e=>setEnd(Math.max(1,Math.min(pages.length,Number(e.target.value))))}/></label></section>}
        {expanded && <section><h4>文字页目录</h4><div className="text-media-page-choices">{pages.map(item=><button type="button" key={item.id} disabled={!!busy} aria-pressed={page?.id===item.id} onClick={()=>{if(move(item.id))setTab('pages');}}>第 {item.ordinal} 页</button>)}</div></section>}
      </div>
      <footer>{tab === 'pages' && <div className="text-media-turn">
        <button type="button" aria-label="上一页" disabled={!!busy || pageIndex<=0} onClick={()=>move(pages[pageIndex-1].id)}>‹</button>
        <span>{pageNumber(pageIndex+1)} / {pageNumber(pages.length)}</span>
        <button type="button" aria-label="下一页" disabled={!!busy || pageIndex<0 || pageIndex>=pages.length-1} onClick={()=>move(pages[pageIndex+1].id)}>›</button>
      </div>}<div className="text-media-actions">
        <button type="button" aria-expanded={expanded} onClick={()=>setExpanded(v=>!v)}>{expanded ? '收起' : '展开'}</button>
        <button type="button" disabled={!!busy || !!session || !canCompose} onClick={()=>run('compose',compose)}>{collections.length ? '重新合成' : '合成当前书'}</button>
        <button type="button" disabled={upload.disabled || !!selectedReceipt?.error || !!busy || !!session} onClick={publish}>{upload.label}</button>
      </div></footer>
    </aside>
  </div>;
}

export default function TextMediaPanel(props) {
  return <TextBookLibrary key={JSON.stringify([props.batchId,props.book.id])} {...props}/>;
}
