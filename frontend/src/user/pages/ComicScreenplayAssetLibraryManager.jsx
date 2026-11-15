import { useEffect, useRef, useState } from 'react';
import { Modal } from 'antd';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import AssetImage from './ComicScreenplayImage.jsx';
import { listComicAssetLibrary, uploadComicAssetLibraryFile, updateComicAssetLibraryEntry, deleteComicAssetLibraryEntry, replaceComicAssetLibraryImage } from '../../shared/api/comicScreenplayAssetLibrary.js';
import { buildCharacterDirectoryImport } from './comic-screenplay-character-directory.js';
import './comic-screenplay-asset-library-manager.css';

const types=['character','scene','prop'], labels={character:'角色',scene:'场景',prop:'道具'};
const defaultCategories=['古言','现言','男生','女生','男主','女主','男配','女配','小孩','中老年','老人','其他职业','动物'];
// CM uses hash IDs. Real creation dates take precedence; legacy records retain the original numeric-ID fallback.
const sortValue=entry=>{const created=Date.parse(entry.created_at||'');return Number.isFinite(created)?created:Number(String(entry.id||'').match(/\d+/)?.[0]||0);};
const draftFor=entry=>({name:'',description:'',category:'',prompt:'',voice:'',...entry});
function IconButton({icon,children,...props}) { return <button type="button" {...props}>{icon?<OriginalIcon name={icon} size={14}/>:null}{children}</button>; }

export default function ComicScreenplayAssetLibraryManager({onBack}) {
  const [section,setSection]=useState('characters'), [tab,setTab]=useState('personal'), [resourceType,setResourceType]=useState(null);
  const [rows,setRows]=useState({character:[],scene:[],prop:[]}), [loading,setLoading]=useState({character:true,scene:true,prop:true}), [errors,setErrors]=useState({});
  const [query,setQuery]=useState(''), [category,setCategory]=useState(''), [size,setSize]=useState('standard'), [sort,setSort]=useState('desc'), [page,setPage]=useState(1);
  const [previewEntry,setPreviewEntry]=useState(null);
  const characterImageInput=useRef(null);
  const [editor,setEditor]=useState(null), [filePreview,setFilePreview]=useState(''), [busy,setBusy]=useState(''), [error,setError]=useState(''), [notice,setNotice]=useState(''), [queue,setQueue]=useState([]);
  const [directoryPlan,setDirectoryPlan]=useState(null), directoryInput=useRef(null), directoryEpoch=useRef(0);
  const alive=useRef(true), lock=useRef(false), requests=useRef({character:0,scene:0,prop:0}), controller=useRef(null), uploadInput=useRef(null);
  async function reload(type) {
    const version=++requests.current[type];setLoading(current=>({...current,[type]:true}));
    try {const entries=await listComicAssetLibrary({type});if(alive.current&&requests.current[type]===version){setRows(current=>({...current,[type]:entries}));setErrors(current=>({...current,[type]:''}));}}
    catch(e){if(alive.current&&requests.current[type]===version)setErrors(current=>({...current,[type]:e.message||'读取失败'}));}
    finally{if(alive.current&&requests.current[type]===version)setLoading(current=>({...current,[type]:false}));}
  }
  useEffect(()=>{alive.current=true;void Promise.allSettled(types.map(reload));return()=>{alive.current=false;controller.current?.abort();};},[]);
  useEffect(()=>{if(!editor?.file){setFilePreview('');return;}const url=URL.createObjectURL(editor.file);setFilePreview(url);return()=>URL.revokeObjectURL(url);},[editor?.file]);
  function upsert(type,entry) {++requests.current[type];setRows(current=>({...current,[type]:[entry,...current[type].filter(item=>item.id!==entry.id)]}));setLoading(current=>({...current,[type]:false}));}
  function switchSection(value){setSection(value);setResourceType(null);setQuery('');setCategory('');setPage(1);setQueue([]);setError('');setNotice('');}
  function openEditor(type,entry=null){setEditor({type,entry,draft:draftFor(entry),file:null,requestKey:crypto.randomUUID(),error:'',uncertain:false});setError('');}
  function changeDraft(values){setEditor(current=>({...current,draft:{...current.draft,...values},error:''}));}
  function chooseCharacterImage(file){if(!file)return;if(!file.type.startsWith('image/')||file.size>40*1024*1024){setEditor(current=>({...current,error:'请选择不超过40MB的图片'}));return;}setEditor(current=>({...current,file,requestKey:crypto.randomUUID(),draft:{...current.draft,name:current.draft.name.trim()?current.draft.name:file.name.replace(/\.[^.]+$/,'').trim()},error:''}));}
  async function saveEditor(){
    if(lock.current||!editor)return;
    if(!editor.draft.name.trim()){setEditor(current=>({...current,error:'请输入名称'}));return;}
    if(!editor.entry&&!editor.file){setEditor(current=>({...current,error:'请选择角色图片'}));return;}
    lock.current=true;setBusy('editor');const snapshot=editor,abort=new AbortController();controller.current=abort;
    try {
      const entry=snapshot.entry?(snapshot.file?await replaceComicAssetLibraryImage(snapshot.entry.id,{...snapshot.draft,type:snapshot.type,file:snapshot.file,requestKey:snapshot.requestKey,expectedRevision:snapshot.entry.revision},{signal:abort.signal}):await updateComicAssetLibraryEntry(snapshot.entry.id,{...snapshot.draft,expectedRevision:snapshot.entry.revision},{signal:abort.signal})):await uploadComicAssetLibraryFile({...snapshot.draft,type:snapshot.type,file:snapshot.file,requestKey:snapshot.requestKey},{signal:abort.signal});
      if(alive.current){upsert(snapshot.type,entry);setEditor(null);setNotice(snapshot.entry?'修改成功':'添加成功');}
    }catch(e){if(alive.current)setEditor(current=>({...current,error:e.message||'保存失败',uncertain:(!snapshot.entry||!!snapshot.file)&&(!e.status||e.status>=500)}));}
    finally{lock.current=false;if(alive.current)setBusy('');}
  }
  async function remove(type,entry){
    if(lock.current)return;lock.current=true;setBusy(`delete:${entry.id}`);setError('');
    try{await deleteComicAssetLibraryEntry(entry.id,entry.revision);if(alive.current){++requests.current[type];setRows(current=>({...current,[type]:current[type].filter(item=>item.id!==entry.id)}));setNotice('删除成功');}}
    catch(e){if(alive.current)setError(e.message||'删除失败');}
    finally{lock.current=false;if(alive.current)setBusy('');}
  }
  async function runUploads(items){
    if(lock.current)return;lock.current=true;setBusy('uploads');setError('');const abort=new AbortController();controller.current=abort;
    const results=items.map(item=>({...item}));
    for(const item of results){
      if(!alive.current)break;if(item.status==='succeeded')continue;
      item.status='uploading';setQueue(results.map(row=>({...row})));
      try{const entry=await uploadComicAssetLibraryFile({file:item.file,type:item.type,name:item.file.name,requestKey:item.requestKey},{signal:abort.signal});if(!alive.current)break;upsert(item.type,entry);item.status='succeeded';item.error='';}
      catch(e){if(!alive.current)break;item.status='failed';item.error=e.message||'上传失败';}
      if(alive.current)setQueue(results.map(row=>({...row})));
    }
    lock.current=false;if(alive.current){setBusy('');setNotice(`已上传 ${results.filter(item=>item.status==='succeeded').length}/${results.length} 张图片`);}
  }
  function chooseResourceFiles(files){if(lock.current||!resourceType||!files.length)return;const items=files.map(file=>({file,type:resourceType,requestKey:crypto.randomUUID(),status:'queued',error:''}));setQueue(items);void runUploads(items);}
  function chooseDirectory(files){
    if(lock.current||!files.length)return;
    try{const plan=buildCharacterDirectoryImport(files);setDirectoryPlan(plan);setError('');setNotice('');}
    catch(e){setDirectoryPlan(null);setError(e.message||'文件夹读取失败');}
  }
  function closeDirectory(){
    directoryEpoch.current++;
    if(busy==='directory'){controller.current?.abort();lock.current=false;setBusy('');setNotice('已停止后续导入。已发送的图片可能已保存，请重新读取个人档案核对。');}
    setDirectoryPlan(null);
  }
  async function importDirectory(){
    if(lock.current||!directoryPlan?.items.length)return;
    lock.current=true;setBusy('directory');const epoch=++directoryEpoch.current,plan=directoryPlan,results=plan.items.map(item=>({...item})),abort=new AbortController();controller.current=abort;
    const current=()=>alive.current&&directoryEpoch.current===epoch;
    try{
      for(const item of results){
        if(!current())return;if(item.status==='succeeded')continue;
        item.status='uploading';setDirectoryPlan({...plan,items:results.map(row=>({...row}))});
        try{
          const entry=await uploadComicAssetLibraryFile({file:item.file,type:'character',name:item.name,category:item.category,requestKey:item.requestKey},{signal:abort.signal});
          if(!current())return;upsert('character',entry);item.status='succeeded';item.error='';
        }catch(e){if(!current())return;item.status='failed';item.error=e.message||'导入失败';}
        if(current())setDirectoryPlan({...plan,items:results.map(row=>({...row}))});
      }
      if(!current())return;
      const succeeded=results.filter(item=>item.status==='succeeded').length;
      setNotice(`已确认导入 ${succeeded}/${results.length} 个角色${plan.skipped.length?`，已跳过${plan.skipped.length}个文件`:''}`);
      if(succeeded){setTab('personal');setQuery('');setCategory('');setPage(1);await reload('character');}
      if(current()&&succeeded===results.length)setDirectoryPlan(null);
    }finally{if(current()){lock.current=false;setBusy('');}}
  }
  const characterSearch=query.trim().toLowerCase();
  const characters=rows.character.filter(entry=>(!category||entry.category===category)&&(!characterSearch||[entry.name,entry.description,...Array.isArray(entry.audioTags)?entry.audioTags:[]].some(value=>String(value||'').toLowerCase().includes(characterSearch)))).slice().sort((a,b)=>sort==='asc'?sortValue(a)-sortValue(b):sortValue(b)-sortValue(a));
  const categories=[...new Set(rows.character.map(entry=>entry.category).filter(Boolean))].sort();
  const pageSize=size==='small'?200:100, totalPages=Math.max(1,Math.ceil(characters.length/pageSize)),activePage=Math.min(page,totalPages);
  const resources=resourceType?rows[resourceType].filter(entry=>String(entry.name||'').toLowerCase().includes(query.toLowerCase())):[];
  const activeType=section==='characters'?'character':resourceType;
  const editorDisabled=!!busy||editor?.uncertain;
  function listState(type){return loading[type]?<p role="status">加载中...</p>:errors[type]?<div role="alert">{errors[type]}<IconButton icon="RefreshOutline" onClick={()=>void reload(type)}>重新读取</IconButton></div>:null;}
  return <div className="comic-library-manager">
    <header className="comic-library-manager-heading"><IconButton icon="ArrowBackOutline" disabled={!!busy} onClick={onBack}>返回作品库</IconButton><nav aria-label="资产库管理分类"><IconButton icon="PeopleOutline" className={section==='characters'?'active':''} disabled={!!busy} onClick={()=>switchSection('characters')}>角色库</IconButton><IconButton icon="FolderOutline" className={section==='assets'?'active':''} disabled={!!busy} onClick={()=>switchSection('assets')}>我的资产</IconButton></nav></header>
    <h2>{section==='characters'?'角色库':'我的资产'}</h2>{section==='assets'?<p className="comic-library-manager-note">管理您的场景和道具资源</p>:null}
    {error?<div role="alert">{error}{activeType?<IconButton icon="RefreshOutline" disabled={!!busy} onClick={()=>{setError('');void reload(activeType);}}>重新读取</IconButton>:null}</div>:null}{notice?<p role="status" className="comic-library-manager-notice">{notice}</p>:null}
    {section==='characters'?<>
      <div className="comic-library-manager-tabs"><IconButton aria-label="切换角色库说明" className={tab==='public'?'active':''} disabled={!!busy} onClick={()=>setTab('public')}>角色库</IconButton><IconButton className={tab==='personal'?'active':''} disabled={!!busy} onClick={()=>setTab('personal')}>个人档案</IconButton></div>
      {tab==='public'?<div className="comic-library-manager-public"><OriginalIcon name="FolderOutline" size={40}/><h3>本地角色库导入说明</h3><p>选择一个本地文件夹（如古言），文件夹名称作为分类，自动命名为古言1、古言2、古言3…。</p><IconButton icon="CloudDownloadOutline" disabled={!!busy} onClick={()=>directoryInput.current?.click()}>批量导入角色库</IconButton><input ref={directoryInput} hidden type="file" multiple webkitdirectory="" directory="" aria-label="选择角色图片文件夹" disabled={!!busy} onChange={event=>{const files=[...event.target.files];event.target.value='';chooseDirectory(files);}}/><p>CM导入方式：保留浏览器提供的文件顺序，子文件夹的图片统一归入根文件夹分类；只支持 PNG、JPEG、WEBP、GIF，单张不超过40MB。确认前可预览命名及跳过文件。</p></div>:<>
        <div className="comic-library-manager-toolbar"><span>共 {characters.length} 个档案</span><select aria-label="角色卡片大小" value={size} onChange={event=>{setSize(event.target.value);setPage(1);}}><option value="standard">标准</option><option value="large">放大</option><option value="small">缩小</option></select><select aria-label="角色分类" value={category} onChange={event=>{setCategory(event.target.value);setPage(1);}}><option value="">全部分类</option>{categories.map(value=><option key={value}>{value}</option>)}</select><select aria-label="角色排序" value={sort} onChange={event=>setSort(event.target.value)}><option value="desc">最新优先</option><option value="asc">最早优先</option></select><label className="comic-library-manager-search"><OriginalIcon name="SearchOutline" size={14}/><input aria-label="搜索角色档案" placeholder="搜索名称、描述" value={query} onChange={event=>{setQuery(event.target.value);setPage(1);}}/></label><IconButton icon="RefreshOutline" disabled={!!busy||loading.character} onClick={()=>void reload('character')}>重新读取</IconButton><IconButton icon="AddOutline" disabled={!!busy} onClick={()=>openEditor('character')}>新增档案</IconButton></div>
        {listState('character')}{!loading.character&&!errors.character?<><div className={`comic-library-manager-character-grid card-size-${size}`}>{characters.slice((activePage-1)*pageSize,activePage*pageSize).map(entry=><article key={entry.id} className="comic-library-manager-character-card"><button type="button" className="comic-library-manager-character-image" aria-label={`查看${entry.name}`} onClick={()=>{if(entry.image?.url)setPreviewEntry(entry);}}><AssetImage src={entry.image?.url||entry.thumbnail||''} name={entry.name} placeholderIcon="PeopleOutline" placeholderSize={40}/></button><div className="comic-library-manager-character-copy"><strong title={entry.name}>{entry.name}</strong><div><IconButton icon="CreateOutline" aria-label={`编辑${entry.name}`} disabled={!!busy} onClick={()=>openEditor('character',entry)}/><IconButton icon="TrashOutline" aria-label={`删除${entry.name}`} disabled={!!busy} onClick={()=>void remove('character',entry)}/></div></div>{entry.category?<small className="comic-library-manager-category">{entry.category}</small>:null}</article>)}</div>{!characters.length?<p className="comic-library-manager-empty">{query||category?'没有匹配的档案':'暂无个人档案'}</p>:null}{characters.length>pageSize?<div className="comic-library-manager-pagination"><IconButton disabled={activePage<=1} onClick={()=>setPage(activePage-1)}>上一页</IconButton><span>{activePage} / {totalPages}</span><IconButton disabled={activePage>=totalPages} onClick={()=>setPage(activePage+1)}>下一页</IconButton></div>:null}</>:null}
        <p className="comic-library-manager-note">当前支持图片与档案资料管理。参考音频、外部主体同步、整库导入导出及图片替换尚未接入。</p>
      </>}
    </>:!resourceType?<div className="comic-library-manager-folders">{['scene','prop'].map(type=><button type="button" key={type} aria-label={labels[type]} className="comic-library-manager-folder" onClick={()=>{setResourceType(type);setQuery('');setQueue([]);}}><span><OriginalIcon name="FolderOutline" size={40}/>{rows[type].length?<small>{rows[type].length}</small>:null}</span><b>{labels[type]}</b>{loading[type]?<small>加载中...</small>:errors[type]?<small>读取失败</small>:null}</button>)}</div>:<>
      <div className="comic-library-manager-toolbar"><IconButton icon="ArrowBackOutline" disabled={!!busy} onClick={()=>{setResourceType(null);setQuery('');setQueue([]);}}>返回分类</IconButton><strong><OriginalIcon name={resourceType==='scene'?'ImageOutline':'CubeOutline'} size={18}/>{labels[resourceType]}</strong><label className="comic-library-manager-search"><OriginalIcon name="SearchOutline" size={14}/><input aria-label={`搜索${labels[resourceType]}`} placeholder="搜索资产" value={query} onChange={event=>setQuery(event.target.value)}/></label><IconButton icon="RefreshOutline" disabled={!!busy||loading[resourceType]} onClick={()=>void reload(resourceType)}>重新读取</IconButton><IconButton icon="AddOutline" disabled={!!busy} onClick={()=>uploadInput.current?.click()}>上传</IconButton><input ref={uploadInput} hidden type="file" multiple accept="image/*" aria-label={`上传${labels[resourceType]}图片`} disabled={!!busy} onChange={event=>{const files=[...event.target.files];event.target.value='';chooseResourceFiles(files);}}/></div>
      {queue.length?<div className="comic-library-manager-upload-results" role="status">{queue.map(item=><p key={item.requestKey}>{item.file.name} · {item.status==='succeeded'?'上传成功':item.status==='failed'?item.error:item.status==='uploading'?'正在上传':'等待上传'}</p>)}{queue.some(item=>item.status==='failed')?<IconButton disabled={!!busy} icon="RefreshOutline" onClick={()=>void runUploads(queue)}>重试失败上传</IconButton>:null}</div>:null}
      {listState(resourceType)}{!loading[resourceType]&&!errors[resourceType]?<><div className="comic-library-manager-resources">{resources.map(entry=><article key={entry.id} className="comic-library-manager-resource-card"><button type="button" className="comic-library-manager-resource-open" aria-label={`编辑${entry.name}`} disabled={!!busy} onClick={()=>openEditor(resourceType,entry)}><div className="comic-library-manager-resource-thumb"><AssetImage src={entry.image?.url||entry.thumbnail_url||''} name={entry.name} placeholderIcon={resourceType==='scene'?'ImageOutline':'CubeOutline'} placeholderSize={24}/></div><span title={entry.name}>{entry.name}</span></button><IconButton className="comic-library-manager-resource-delete" icon="TrashOutline" aria-label={`删除${entry.name}`} disabled={!!busy} onClick={()=>void remove(resourceType,entry)}/></article>)}</div>{!resources.length?<p className="comic-library-manager-empty">{query?'没有匹配的资产':`暂无${labels[resourceType]}，可上传图片`}</p>:null}</>:null}
    </>}
    {previewEntry?<Modal open width="auto" style={{maxWidth:'90vw'}} className="comic-library-manager-preview" title={previewEntry.name} footer={null} onCancel={()=>setPreviewEntry(null)}><AssetImage src={previewEntry.image?.url} name={previewEntry.name}/></Modal>:null}
    {directoryPlan?<Modal open width={760} className="comic-library-manager-editor comic-library-manager-directory" title="批量导入角色库" maskClosable={false} onCancel={closeDirectory} footer={<div className="comic-library-manager-editor-footer"><IconButton onClick={closeDirectory}>{busy==='directory'?'停止后续导入':directoryPlan.items.some(item=>item.status!=='queued')?'关闭导入结果':'取消导入'}</IconButton>{busy!=='directory'?<IconButton disabled={!directoryPlan.items.length} onClick={()=>void importDirectory()}>{directoryPlan.items.some(item=>item.status==='failed')?'重试失败导入':'确认导入'}</IconButton>:null}</div>}>
      <p>分类：<strong>{directoryPlan.category}</strong> · 可导入 {directoryPlan.items.length} 张图片 · 已跳过{directoryPlan.skipped.length}个文件</p><p className="comic-library-manager-note">按下表顺序上传，失败重试保留原名称与编号；成功项不会重复上传。</p>
      <div className="comic-library-manager-directory-table"><table><thead><tr><th>原文件</th><th>导入名称</th><th>状态</th></tr></thead><tbody>{directoryPlan.items.map(item=><tr key={item.requestKey}><td>{item.relativePath}</td><td>{item.name}</td><td>{item.status==='succeeded'?'导入成功':item.status==='failed'?item.error:item.status==='uploading'?'正在导入':'待导入'}</td></tr>)}</tbody></table></div>
      {!directoryPlan.items.length?<p role="alert">没有可导入的图片，请选择支持格式的图片文件夹。</p>:null}{directoryPlan.skipped.length?<details className="comic-library-manager-directory-skipped"><summary>已跳过{directoryPlan.skipped.length}个文件</summary><ul>{directoryPlan.skipped.map(item=><li key={item.relativePath}>{item.relativePath}：{item.reason}</li>)}</ul></details>:null}
    </Modal>:null}
    {editor?<Modal open width={editor.type==='character'?900:560} className="comic-library-manager-editor" title={editor.type==='character'?(editor.entry?'编辑档案':'新建档案'):`编辑${labels[editor.type]}`} maskClosable={false} keyboard={!busy} onCancel={()=>{if(!busy)setEditor(null);}} footer={<div className="comic-library-manager-editor-footer"><IconButton disabled={!!busy} onClick={()=>setEditor(null)}>取消</IconButton><IconButton disabled={!!busy} onClick={()=>void saveEditor()}>{busy==='editor'?'正在保存…':editor.error&&editor.uncertain?'重试应用':editor.type==='character'?'应用':'保存'}</IconButton></div>}>
      {editor.error?<div role="alert">{editor.error}{editor.uncertain?<p>保存结果尚未核对，请保持本次内容重试。</p>:null}</div>:null}<div className={editor.type==='character'?'comic-library-manager-character-editor':'comic-library-manager-resource-editor'}><div className="comic-library-manager-editor-preview" role={editor.type==='character'?'button':undefined} tabIndex={editor.type==='character'?0:undefined} aria-label={editor.type==='character'?'选择角色图片':undefined} aria-disabled={editorDisabled} onClick={()=>{if(editor.type==='character'&&!editorDisabled)characterImageInput.current?.click();}} onKeyDown={event=>{if(editor.type==='character'&&!editorDisabled&&(event.key==='Enter'||event.key===' ')){event.preventDefault();characterImageInput.current?.click();}}}>{filePreview?<img src={filePreview} alt="待上传角色"/>:<AssetImage src={editor.entry?.image?.url} name={editor.draft.name||labels[editor.type]} placeholderIcon={editor.type==='character'?'PeopleOutline':'ImageOutline'} placeholderSize={40}/>}</div><div className="comic-library-manager-editor-fields"><label>名称<input aria-label={editor.type==='character'?'档案名称':'素材名称'} disabled={editorDisabled} value={editor.draft.name} onChange={event=>changeDraft({name:event.target.value})}/></label>{editor.type==='character'?<><label>分类<input aria-label="档案分类" list="comic-library-manager-categories" disabled={editorDisabled} value={editor.draft.category} onChange={event=>changeDraft({category:event.target.value})}/><datalist id="comic-library-manager-categories">{[...new Set([...defaultCategories,...categories])].map(value=><option key={value} value={value}/>)}</datalist></label><label>{editor.entry?'更换图片':'上传图片'}<input ref={characterImageInput} type="file" accept="image/*" aria-label="角色图片" disabled={editorDisabled} onChange={event=>{chooseCharacterImage(event.target.files?.[0]);event.target.value='';}}/></label><label>描述<textarea rows={8} aria-label="档案描述" disabled={editorDisabled} value={editor.draft.description} onChange={event=>changeDraft({description:event.target.value})}/></label></>:<label>提示词/描述（可选）<textarea rows={6} aria-label="素材提示词" disabled={editorDisabled} value={editor.draft.prompt} onChange={event=>changeDraft({prompt:event.target.value})}/></label>}</div></div>
    </Modal>:null}
  </div>;
}
