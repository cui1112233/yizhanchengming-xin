import React, { useEffect, useState } from 'react';
import { createTextCollection, getTextCollection, listTextCollections, listTextPages, prepareTextBook } from '../../../shared/api/batchFactoryText.js';
import AuthenticatedTextPageImage from './AuthenticatedTextPageImage.jsx';
import { canCreateTextCollection, textBookCompletion, textPagePresentation } from './textMediaState.js';
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

function TextBookLibrary({ batchId, book, settings = {}, initialPageId, onSaved, onClose, onUpload }) {
  const [refresh, setRefresh] = useState(0);
  const [tab, setTab] = useState('pages');
  const [pageId, setPageId] = useState(initialPageId);
  const [selectedCandidate, selectCandidate] = useState('');
  const [chosenCollection, chooseCollection] = useState('');
  const [collectionDetail, setCollectionDetail] = useState(null);
  const [collectionImage, setCollectionImage] = useState('');
  const [expanded, setExpanded] = useState(false);
  const [busy, setBusy] = useState('');
  const [actionError, setActionError] = useState('');
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
  const currentCollection = collectionDetail?.id === collection?.id ? collectionDetail : collection;
  const collectionItems = currentCollection?.items || [];
  const candidates = artifacts.filter(a => tab === 'cover' ? a.kind === 'cover' : a.kind === 'page' && a.pageId === page?.id);
  const activeId = tab === 'cover' ? doc?.coverArtifactId : state.artifactId;
  const display = candidates.find(a => a.id === selectedCandidate) || candidates.find(a => a.id === activeId);
  const preview = tab === 'collections'
    ? collectionItems.find(i => i.artifactId === collectionImage)?.contentUrl || currentCollection?.coverContentUrl
    : display?.contentUrl;
  const title = book.title || '未命名作品';
  const canCompose = canCreateTextCollection(pages, undefined, doc, artifacts);
  const background = page?.background || doc?.configSnapshot?.background || settings.textBackground;
  const reload = () => setRefresh(value => value + 1);
  const move = (newId) => { setPageId(newId); selectCandidate(''); };
  async function run(action, callback) {
    if (busy) return;
    setBusy(action); setActionError('');
    try { await callback(); reload(); onSaved?.(); } catch (e) { setActionError(e.message || '操作失败'); }
    finally { setBusy(''); }
  }
  async function compose() {
    const result = await createTextCollection(batchId, book.id, {
      expectedRevision:doc.revision, versionId:doc.activeVersionId, paginationVersion:doc.paginationVersion,
      pageIds:pages.map(p => p.id)
    });
    chooseCollection(result.id); setCollectionDetail(result); setCollectionImage(''); setTab('collections');
  }
  return <div className="text-media-panel" aria-label="文本媒体库">
    <section className="text-media-preview" aria-label="成品预览">{preview
      ? <AuthenticatedTextPageImage contentUrl={preview} alt={tab === 'pages' ? `文字页 ${page?.ordinal}` : tab === 'cover' ? '文本封面' : '文本合集预览'} />
      : <div className="text-media-empty">{loading ? '正在读取文本页…' : tab === 'cover' ? '暂无封面图' : tab === 'collections' ? '暂无合集预览' : state.error || state.label}</div>}</section>
    <aside className="text-media-tools">
      <header><div><small>媒体库 · 文本</small><h3>{title}</h3></div>
        <button type="button" aria-label="关闭文本媒体库" onClick={onClose}>×</button>
        <div className="text-media-progress"><span>完成度 {completion.ready} / {completion.total} 页</span><progress max="100" value={completion.percent}/></div>
      </header>
      <nav role="tablist" aria-label="文本媒体视图">{tabs.map(([key,label]) =>
        <button key={key} type="button" role="tab" aria-label={label} aria-selected={tab === key} onClick={() => {setTab(key); selectCandidate('');setActionError('');}}>
          {label}{key === 'pages' ? ' '+pages.length : key === 'collections' ? ' '+collections.length : ''}</button>)}</nav>
      <div className="text-media-scroll" role="tabpanel">
        {(error || actionError) && <div role="alert">{error || actionError} <button type="button" onClick={() => {setActionError('');reload();}}>刷新</button></div>}
        {tab === 'collections' ? <>
          <CollectionChoices collections={collections} current={collection} choose={id=>{chooseCollection(id);setCollectionImage('');}} />
          {collection && <section><h4>合集内容</h4><div className="text-media-page-choices"><button type="button" onClick={()=>setCollectionImage('')}>封面</button>
            {collectionItems.map(i=><button type="button" key={i.artifactId} onClick={()=>setCollectionImage(i.artifactId)}>文字页 {i.ordinal}</button>)}</div></section>}
        </> : <>
          <CandidateList items={candidates} selectedId={display?.id} activeId={activeId} onSelect={selectCandidate}/>
          {tab === 'pages' ? <>
            <section><h4>本页背景</h4>{background?.color ? <span><i className="text-media-swatch" style={{backgroundColor:background.color}}/>{background.color}</span> : background?.id ? '背景图片' : '继承整书背景'}</section>
            <section><h4>页面内容</h4><p className="text-media-copy">{page?.text || '尚未分页'}</p></section>
          </> : <section><h4>封面标题</h4><p>{doc?.configSnapshot?.cover?.title || title}</p>{doc?.coverLocked && <small>上传成品封面</small>}</section>}
          {!doc && !loading && !error && <button type="button" disabled={!!busy} onClick={()=>run('prepare',()=>prepareTextBook(batchId,book.id,{}))}>准备文字页</button>}
        </>}
        {expanded && <section><h4>文字页目录</h4><div className="text-media-page-choices">{pages.map(item=><button type="button" key={item.id} aria-pressed={page?.id===item.id} onClick={()=>{move(item.id);setTab('pages');}}>第 {item.ordinal} 页</button>)}</div></section>}
      </div>
      <footer>{tab === 'pages' && <div className="text-media-turn">
        <button type="button" aria-label="上一页" disabled={pageIndex<=0} onClick={()=>move(pages[pageIndex-1].id)}>‹</button>
        <span>{pageNumber(pageIndex+1)} / {pageNumber(pages.length)}</span>
        <button type="button" aria-label="下一页" disabled={pageIndex<0 || pageIndex>=pages.length-1} onClick={()=>move(pages[pageIndex+1].id)}>›</button>
      </div>}<div className="text-media-actions">
        <button type="button" aria-expanded={expanded} onClick={()=>setExpanded(v=>!v)}>{expanded ? '收起' : '展开'}</button>
        <button type="button" disabled={!!busy || !canCompose} onClick={()=>run('compose',compose)}>{collections.length ? '重新合成' : '合成当前书'}</button>
        <button type="button" disabled={!onUpload || !collection || !!busy} onClick={()=>onUpload(collection)}>上传网络</button>
      </div></footer>
    </aside>
  </div>;
}

export default function TextMediaPanel(props) {
  return <TextBookLibrary key={JSON.stringify([props.batchId,props.book.id])} {...props}/>;
}
