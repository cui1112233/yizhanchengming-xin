import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import { importCharacterArchive, clearCharacterArchive, exportCharacterArchive, validateCharacterArchiveFiles } from '../../shared/api/comicScreenplayCharacterArchive.js';
import { getToken } from '../../shared/api/client.js';

// Files and request keys remain in this browser session's memory, scoped to
// the login session. They never enter localStorage or another account's UI.
const pendingBySession=new Map();
const recoveryMessage='上次档案操作结果尚未核对，请先使用原请求核对；重新选择文件不会重复导入。';
export default function CharacterArchive({entries,busy,unavailable,lock,setBusy,reload,onNotice,onError}) {
  const folder=useRef(null),single=useRef(null),alive=useRef(true),controller=useRef(null);
  const [session]=useState(getToken);
  const [pending,setPending]=useState(()=>pendingBySession.get(session)||null),[failure,setFailure]=useState(()=>pendingBySession.has(session)?recoveryMessage:''),[chooser,setChooser]=useState(false);
  const current=()=>alive.current&&getToken()===session;
  function remember(operation){if(operation)pendingBySession.set(session,operation);else pendingBySession.delete(session);if(current())setPending(operation);}
  function complete(operation){if(pendingBySession.get(session)===operation)remember(null);}
  function unresolved(){if(pendingBySession.has(session)){setFailure(recoveryMessage);return true;}return false;}
  useEffect(()=>{alive.current=true;return()=>{alive.current=false;controller.current?.abort();};},[]);
  const disabled=!!busy||unavailable||getToken()!==session;
  async function run(operation) {
    if(lock.current||unavailable||!current())return;
    if(operation.kind!=='export'&&pendingBySession.get(session)!==operation){setPending(pendingBySession.get(session)||null);setFailure('本次操作已更新，请先核对当前档案操作。');return;}
    lock.current=true;setBusy('archive');setFailure('');onError('');onNotice('');
    const abort=new AbortController();controller.current=abort;
    try {
      if(operation.kind==='import') {
        const data=await importCharacterArchive(operation.input,{signal:abort.signal});
        complete(operation);if(!current())return;onNotice(`已导入 ${data.imported_count} 个档案`);
      } else if(operation.kind==='clear') {
        const data=await clearCharacterArchive(operation.input,{signal:abort.signal});
        complete(operation);if(!current())return;onNotice(`已清空 ${data.cleared_count} 个角色`);
      } else {
        const blob=await exportCharacterArchive({signal:abort.signal});if(!current())return;
        const url=URL.createObjectURL(blob),link=document.createElement('a');link.href=url;link.download='characters.lschar';document.body.appendChild(link);
        try{link.click();}finally{link.remove();setTimeout(()=>URL.revokeObjectURL(url),1000);}
        onNotice('档案已下载');
      }
      if(current())await reload('character');
    } catch(error) {
      if(!current())return;
      if(operation.kind==='clear'&&error.status===409){complete(operation);setFailure('角色列表已变化，请重新确认完整数量后清空。');await reload('character');}
      else {setFailure(error.message||'档案操作失败');}
    } finally {lock.current=false;if(alive.current)setBusy('');}
  }
  function choose(files) {
    if(lock.current||!files.length||!current()||unresolved())return;
    try{validateCharacterArchiveFiles(files);}catch(error){setFailure(error.message);return;}
    const operation={kind:'import',input:{files,requestKey:crypto.randomUUID()}};setChooser(false);remember(operation);void run(operation);
  }
  function openChooser(input){if(!current()||lock.current||unresolved())return;setChooser(true);input.current?.click();}
  function abandon(){
    if(lock.current||!current()||!window.confirm('上次操作可能已经保存。放弃核对后再次导入会追加新档案，可能产生重复。确定放弃核对并重新导入？'))return;
    remember(null);setFailure('');setChooser(true);folder.current?.click();
  }
  function clear() {
    if(lock.current||!current())return;
    if(pending?.kind==='clear'){void run(pending);return;}
    if(unresolved())return;
    if(!entries.length){setFailure('角色库已经是空的');return;}
    if(!window.confirm(`确定要清空所有 ${entries.length} 个角色吗？\n\n此操作不可恢复！`))return;
    const operation={kind:'clear',input:{requestKey:crypto.randomUUID(),expected:entries.map(({id,revision})=>({id,revision}))}};remember(operation);void run(operation);
  }
  const button=(name,icon,action)=><button type="button" disabled={disabled} onClick={action}><OriginalIcon name={icon} size={14}/>{name}</button>;
  return <>
    {button('导入档案','CloudDownloadOutline',()=>openChooser(folder))}
    {button('清空角色','TrashOutline',clear)}
    {button('导出档案','CloudUploadOutline',()=>{if(!entries.length){setFailure('暂无可导出档案');return;}void run({kind:'export'});})}
    <input hidden ref={folder} type="file" multiple webkitdirectory="" directory="" aria-label="导入角色档案文件夹" onCancel={()=>setChooser(false)} onChange={event=>{choose(Array.from(event.target.files||[]));event.target.value='';}}/>
    <input hidden ref={single} type="file" accept=".lschar,.json" aria-label="导入单个角色归档" onChange={event=>{choose(Array.from(event.target.files||[]));event.target.value='';}}/>
    {(failure||chooser||pending)?createPortal(<div className="comic-library-archive-feedback" role={failure?'alert':'status'} style={{position:'fixed',bottom:24,right:24,zIndex:1200,background:'var(--bg-secondary, #fff)',padding:16,borderRadius:8,maxWidth:440,boxShadow:'0 2px 16px #0003'}}>
      {failure?<p>{failure}</p>:<p>请选择包含 characters.lschar 或 characters.json 的文件夹；内联归档也可选择单个文件。</p>}
      {pending?<button type="button" disabled={disabled} onClick={()=>void run(pending)}>{pending.kind==='import'?'重试核对导入':'重试核对清空'}</button>:null}
      {pending?.kind==='import'?<button type="button" disabled={disabled} onClick={abandon}>放弃核对并重新导入</button>:null}
      <button type="button" disabled={disabled} onClick={()=>openChooser(single)}>选择单个归档</button>
      <button type="button" disabled={!!busy} onClick={()=>{setFailure('');setChooser(false);}}>关闭提示</button>
    </div>,document.body):null}
  </>;
}
