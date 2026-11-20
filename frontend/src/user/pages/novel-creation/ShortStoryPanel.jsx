import { Button, Input, Modal } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { ArrowLeft, Download, Pencil, Save, X } from 'lucide-react';
import { novelRequest, novelProjectRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { downloadBlob } from './contracts.js';

export default function ShortStoryPanel({owner,projectId,refreshKey,onRun,onTitle,modelId,thinkingLevel}){
 const scope=useWorkspaceScope(`${owner}:${projectId}`),[story,setStory]=useState(null),[error,setError]=useState(''),[edit,setEdit]=useState(null),[preview,setPreview]=useState(null),[draft,setDraft]=useState(null),[deferred,setDeferred]=useState(false),[busy,setBusy]=useState(false),lock=useRef(null);
 async function load(){try{await scope.call(novelProjectRequest(projectId,'/short-story',undefined,{signal:scope.signal}),r=>{setStory(r.story);onTitle?.(r.story.title);setError('');});}catch(e){if(scope.isCurrent())setError(e.message||'短篇资料读取失败');}}
 useEffect(()=>{setStory(null);setPreview(null);setDraft(null);setEdit(null);setDeferred(false);setError('');setBusy(false);lock.current=null;load();},[owner,projectId]);useEffect(()=>{setDeferred(false);load();},[refreshKey]);
 async function action(fn){if(lock.current)return;const token=Symbol();lock.current=token;setBusy(true);setError('');try{await fn();}catch(e){if(scope.isCurrent())setError(e.message||'操作失败，已保存资料保留');}finally{if(lock.current===token){lock.current=null;if(scope.isCurrent())setBusy(false);}}}
 async function confirm(){await action(async()=>{const r=await scope.call(novelRequest(`/runs/${encodeURIComponent(story.runId)}/confirm-short-story`,{version:story.version,modelId:modelId||undefined,thinkingLevel:modelId?thinkingLevel:undefined},{signal:scope.signal}));if(r){onRun?.(r.run);await load();}});}
 const download=f=>downloadBlob(new Blob([f.content],{type:'text/plain;charset=utf-8'}),f.key.replaceAll('/','-'));
 // Only the server's matching artifact can grant editing, including chapter shortcuts.
 const selected=preview&&(story?.artifacts.find(f=>f.key===preview.key)||{...preview,editable:false});
 // ShortStory persists "writing" through full-draft writing, review and revision.
 const producing=['queued','running','writing'].includes(story?.status);
 function openFile(file){setPreview(file);setDraft(null);setError('');}
 function back(){setPreview(null);setDraft(null);setError('');}
 async function saveFile(){if(!draft||!selected?.editable)return;await action(async()=>{const r=await scope.call(novelProjectRequest(projectId,'/short-story/artifact',{version:draft.version,key:draft.key,content:draft.content},{method:'PUT',signal:scope.signal}));if(r){setStory(r.story);onTitle?.(r.story.title);setPreview(r.story.artifacts.find(f=>f.key===draft.key)||{...selected,content:draft.content});setDraft(null);}});}
 return <section className="novel-panel novel-short-story" aria-label="短篇资料">
  <h3>{story?.title||'短篇资料'}</h3>{error&&<p role="alert">{error}<Button onClick={load}>重新读取短篇资料</Button></p>}
  {story?.status==='awaiting_confirmation'&&!deferred&&<div role="status"><p>{story.checkpoint==='blueprint'?'创作蓝图已生成，可在右侧查看或修改后确认继续生成正文。':'大纲已保存，可修改后确认继续。'}</p>{story.checkpoint!=='blueprint'&&story.outline&&<Button disabled={busy||!!draft} onClick={()=>setEdit({version:story.version,outline:structuredClone(story.outline)})}>编辑短篇大纲</Button>}<Button type="primary" loading={busy} disabled={!!draft||!!edit} onClick={confirm}>{story.checkpoint==='blueprint'?'确认继续':'确认大纲并继续'}</Button><Button disabled={busy} onClick={()=>setDeferred(true)}>稍后</Button></div>}
  {selected?<section aria-label="文件预览">
   <div style={{display:'flex',alignItems:'center',gap:6,flexWrap:'wrap',margin:'12px 0'}}><Button size="small" icon={<ArrowLeft size={14}/>} disabled={busy} onClick={back}>返回</Button><strong style={{flex:1}}>{selected.title}</strong>{draft?<><Button size="small" icon={<Save size={14}/>} type="primary" loading={busy} onClick={saveFile}>保存</Button><Button size="small" icon={<X size={14}/>} disabled={busy} onClick={()=>{setDraft(null);setError('');}}>取消</Button></>:<>{selected.editable&&<Button size="small" icon={<Pencil size={14}/>} disabled={busy} onClick={()=>setDraft({key:selected.key,version:story.version,content:selected.content??''})}>编辑</Button>}<Button size="small" icon={<Download size={14}/>} onClick={()=>download(selected)}>下载文件</Button></>}</div>
   {draft?<Input.TextArea aria-label="文件内容" value={draft.content} disabled={busy} autoSize={{minRows:14,maxRows:30}} onChange={e=>setDraft(v=>({...v,content:e.target.value}))}/>:<pre style={{whiteSpace:'pre-wrap',overflowWrap:'anywhere',maxHeight:'65vh',overflow:'auto'}}>{selected.content??'文件不存在'}</pre>}
  </section>:story&&<><p>目标总字数 {story.input.targetWordCount.toLocaleString()} · {story.input.targetChapters} 章</p>
  {story.status==='completed'&&<p role="status">生成完成，正文与资料已保存。</p>}
  {story.fullDraft&&<section aria-label="正文版本">
   <h4>正文版本</h4>
   {story.fullDraft.revisionStatus!=='drafting'&&['v001','v002'].includes(story.fullDraft.currentVersion)&&<p>当前采用{story.fullDraft.currentVersion==='v002'?'第二稿':'第一稿'}。</p>}
   {story.fullDraft.revisionStatus==='drafting'&&<p role="status">{story.fullDraft.missingChapters?.length?`${producing?'正在补齐正文':'正文尚未补齐'}，已补写${story.fullDraft.completionAttempts}次（最多3次），还缺第${story.fullDraft.missingChapters.join('、')}章。`:producing?'正在生成整篇正文。':'整篇正文尚未完成。'}</p>}
   {story.fullDraft.revisionStatus==='pending'&&<p role="status">{producing?'第一稿已保存，正在审稿或修订。':'第一稿已保存，审稿或修订未完成。'}</p>}
   {story.fullDraft.revisionStatus==='adopted'&&<p role="status">已采用第二稿，第一稿保留供查看。</p>}
   {story.fullDraft.revisionStatus==='rejected'&&<p role="alert">第二稿未采用，继续保留第一稿。{story.fullDraft.revisionWarning}</p>}
   <div style={{display:'flex',gap:6,flexWrap:'wrap'}}>{[
    ['drafts/v001.md','查看第一稿'],['drafts/v002.md','查看第二稿'],['drafts/v002-rejected.md','查看未采用的第二稿'],['drafts/v001-incomplete.md','查看未完成稿'],['reviews/full-draft.md','查看整篇审稿'],['package/report.md','查看成品报告']
   ].map(([key,label])=>{const file=story.artifacts.find(f=>f.key===key);return file?<Button size="small" key={key} onClick={()=>openFile(file)}>{label}</Button>:null;})}</div>
  </section>}
  <details open><summary>章节（{story.chapters.length}）</summary>{story.chapters.length?story.chapters.map(c=><Button block type="text" key={c.number} onClick={()=>openFile({key:`chapters/${String(c.number).padStart(3,'0')}.txt`,title:c.title,content:c.content})}>{c.number}. {c.title}</Button>):<p>暂无已保存正文。</p>}</details>
  <details open><summary>大纲</summary>{story.outline?<><p>{story.outline.summary.slice(0,180)}{story.outline.summary.length>180?'…':''}</p><Button type="link" onClick={()=>openFile({key:'outline.txt',title:'短篇大纲',content:[story.outline.title,story.outline.summary,...story.outline.chapters.map(c=>`${c.number}. ${c.title}\n${c.summary}`)].join('\n\n')})}>查看完整大纲 →</Button></>:<p>大纲生成后显示摘要。</p>}</details>
  <details open><summary>创作资料与成品</summary>{story.artifacts.length?story.artifacts.map(f=><Button block type="text" key={f.key} onClick={()=>openFile(f)}>{f.title}</Button>):<p>暂无成品文件。</p>}</details>
  {story.imitation&&<p role="status">{story.imitation.editorial ? `编辑终审：${story.imitation.editorial.rating}。${story.qualityStale?'正文或资料修改后，原终审结论已过期。':''}审核意见和正文已保存在上方文件中。` : story.imitation.blueprint ? '仿写蓝图已保存；拆书、方向及蓝图可在上方查看。' : '仿写中间资料会在每个步骤完成后保存。'}</p>}
  {story.recommendation&&<details open><summary>推荐</summary><p>{story.recommendation.slice(0,180)}</p><Button type="link" onClick={()=>openFile({key:'recommendation.txt',title:'推荐',content:story.recommendation})}>查看完整推荐 →</Button></details>}
  </>}
  <Modal rootClassName="novel-modal" open={!!edit} title="编辑短篇大纲" okText="保存大纲" cancelText="取消" confirmLoading={busy} onCancel={()=>setEdit(null)} onOk={()=>action(async()=>{const r=await scope.call(novelProjectRequest(projectId,'/short-story/outline',{version:edit.version,outline:edit.outline},{method:'PUT',signal:scope.signal}));if(r){setStory(r.story);setEdit(null);}})}>
   <label>作品标题<Input aria-label="短篇大纲标题" value={edit?.outline.title||''} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,title:e.target.value}}))}/></label>
   <label>故事大纲<Input.TextArea aria-label="短篇大纲内容" rows={5} value={edit?.outline.summary||''} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,summary:e.target.value}}))}/></label>
   {edit?.outline.chapters.map((c,index)=><section key={c.number}><b>第{c.number}章</b><Input aria-label={`第${c.number}章标题`} value={c.title} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,chapters:v.outline.chapters.map((item,i)=>i===index?{...item,title:e.target.value}:item)}}))}/><Input.TextArea aria-label={`第${c.number}章梗概`} rows={3} value={c.summary} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,chapters:v.outline.chapters.map((item,i)=>i===index?{...item,summary:e.target.value}:item)}}))}/></section>)}
  </Modal>
 </section>;
}
