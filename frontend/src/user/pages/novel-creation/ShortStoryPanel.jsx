import { Button, Input, Modal } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { novelRequest, novelProjectRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { downloadBlob } from './contracts.js';

export default function ShortStoryPanel({owner,projectId,refreshKey,onRun,onTitle}){
 const scope=useWorkspaceScope(`${owner}:${projectId}`),[story,setStory]=useState(null),[error,setError]=useState(''),[edit,setEdit]=useState(null),[preview,setPreview]=useState(null),[busy,setBusy]=useState(false),lock=useRef(false);
 async function load(){try{await scope.call(novelProjectRequest(projectId,'/short-story',undefined,{signal:scope.signal}),r=>{setStory(r.story);onTitle?.(r.story.title);setError('');});}catch(e){if(scope.isCurrent())setError(e.message||'短篇资料读取失败');}}
 useEffect(()=>{setStory(null);load();},[owner,projectId]);useEffect(()=>{load();},[refreshKey]);
 async function action(fn){if(lock.current)return;lock.current=true;setBusy(true);setError('');try{await fn();}catch(e){if(scope.isCurrent())setError(e.message||'操作失败，已保存资料保留');}finally{lock.current=false;if(scope.isCurrent())setBusy(false);}}
 async function confirm(){await action(async()=>{const r=await scope.call(novelRequest(`/runs/${encodeURIComponent(story.runId)}/confirm-short-story`,{version:story.version},{signal:scope.signal}));if(r){onRun?.(r.run);await load();}});}
 const download=f=>downloadBlob(new Blob([f.content],{type:'text/plain;charset=utf-8'}),f.key.replaceAll('/','-'));
 return <section className="novel-panel novel-short-story" aria-label="短篇资料">
  <h3>{story?.title||'短篇资料'}</h3>{error&&<p role="alert">{error}<Button onClick={load}>重新读取短篇资料</Button></p>}
  {story&&<><p>目标总字数 {story.input.targetWordCount.toLocaleString()} · {story.input.targetChapters} 章</p>
  {story.status==='awaiting_confirmation'&&<div role="status"><p>大纲已保存，可修改后确认继续。</p><Button disabled={busy} onClick={()=>setEdit({version:story.version,outline:structuredClone(story.outline)})}>编辑短篇大纲</Button><Button type="primary" loading={busy} onClick={confirm}>确认大纲并继续</Button></div>}
  <details open><summary>章节（{story.chapters.length}）</summary>{story.chapters.length?story.chapters.map(c=><Button block type="text" key={c.number} onClick={()=>setPreview({key:`chapters/${c.number}.txt`,title:c.title,content:c.content})}>{c.number}. {c.title}</Button>):<p>暂无已保存正文。</p>}</details>
  <details open><summary>大纲</summary>{story.outline?<><p>{story.outline.summary.slice(0,180)}{story.outline.summary.length>180?'…':''}</p><Button type="link" onClick={()=>setPreview({key:'outline.txt',title:'短篇大纲',content:[story.outline.title,story.outline.summary,...story.outline.chapters.map(c=>`${c.number}. ${c.title}\n${c.summary}`)].join('\n\n')})}>查看完整大纲 →</Button></>:<p>大纲生成后显示摘要。</p>}</details>
  <details open><summary>成品文件</summary>{story.artifacts.length?story.artifacts.map(f=><Button block type="text" key={f.key} onClick={()=>setPreview(f)}>{f.title}</Button>):<p>暂无成品文件。</p>}</details>
  {story.recommendation&&<details open><summary>推荐</summary><p>{story.recommendation.slice(0,180)}</p><Button type="link" onClick={()=>setPreview({key:'recommendation.txt',title:'推荐',content:story.recommendation})}>查看完整推荐 →</Button></details>}
  </>}
  <Modal rootClassName="novel-modal" open={!!edit} title="编辑短篇大纲" okText="保存大纲" cancelText="取消" confirmLoading={busy} onCancel={()=>setEdit(null)} onOk={()=>action(async()=>{const r=await scope.call(novelProjectRequest(projectId,'/short-story/outline',{version:edit.version,outline:edit.outline},{method:'PUT',signal:scope.signal}));if(r){setStory(r.story);setEdit(null);}})}>
   <label>作品标题<Input aria-label="短篇大纲标题" value={edit?.outline.title||''} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,title:e.target.value}}))}/></label>
   <label>故事大纲<Input.TextArea aria-label="短篇大纲内容" rows={5} value={edit?.outline.summary||''} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,summary:e.target.value}}))}/></label>
   {edit?.outline.chapters.map((c,index)=><section key={c.number}><b>第{c.number}章</b><Input aria-label={`第${c.number}章标题`} value={c.title} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,chapters:v.outline.chapters.map((item,i)=>i===index?{...item,title:e.target.value}:item)}}))}/><Input.TextArea aria-label={`第${c.number}章梗概`} rows={3} value={c.summary} onChange={e=>setEdit(v=>({...v,outline:{...v.outline,chapters:v.outline.chapters.map((item,i)=>i===index?{...item,summary:e.target.value}:item)}}))}/></section>)}
  </Modal>
  <Modal rootClassName="novel-modal" open={!!preview} title={preview?.title} onCancel={()=>setPreview(null)} footer={<><Button onClick={()=>setPreview(null)}>关闭</Button><Button onClick={()=>download(preview)}>下载文件</Button></>}><pre style={{whiteSpace:'pre-wrap',maxHeight:'65vh',overflow:'auto'}}>{preview?.content}</pre></Modal>
 </section>;
}
