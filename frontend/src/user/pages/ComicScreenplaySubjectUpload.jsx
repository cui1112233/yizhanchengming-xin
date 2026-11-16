import {useEffect,useRef,useState} from 'react';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import {getToken} from '../../shared/api/client.js';
import {getComicSubjectUploadCapability,createComicSubjectUpload,getComicSubjectUpload} from '../../shared/api/comicScreenplayAssetLibrary.js';

// Page-lifetime recovery only. Files and private draft text never enter browser storage.
// Each account and library entry has one unresolved operation, including a new draft slot.
const pendingByOwner=new Map();
function ownerAttempts(owner){if(!pendingByOwner.has(owner))pendingByOwner.set(owner,new Map());return pendingByOwner.get(owner);}
export default function SubjectUpload({editor,disabled,onBusyChange,onSuccess,onRestore}) {
 const owner=useRef(getToken()),slot=useRef(editor.entry?.id||'new-character'),attempt=useRef(ownerAttempts(owner.current).get(slot.current)||null);
 const [capability,setCapability]=useState(null),[message,setMessage]=useState(attempt.current?'上次上传尚未核对，请核对本次上传；不会重新创建素材。':'正在检查上传配置…'),[working,setWorking]=useState(false),[notFound,setNotFound]=useState(false),[safeNew,setSafeNew]=useState(false),[verified,setVerified]=useState(false),[operation,setOperation]=useState(attempt.current?.operation||null);
 const alive=useRef(false),controller=useRef(null),epoch=useRef(0),busy=useRef(false);
 const current=version=>alive.current&&epoch.current===version&&owner.current===getToken();
 const matchesDraft=()=>{
  const saved=attempt.current;if(!saved)return false;
  return editor.draft.name===saved.input.name&&editor.draft.description===saved.input.description&&
   (saved.input.file?editor.file===saved.input.file:!editor.file&&editor.entry?.id===saved.input.entryId&&editor.entry?.revision===saved.input.expectedRevision);
 };
 const canRestore=!!attempt.current&&(!!attempt.current.input.file||editor.entry?.revision===attempt.current.input.expectedRevision);
 useEffect(()=>{alive.current=true;const abort=new AbortController();getComicSubjectUploadCapability({signal:abort.signal}).then(value=>{if(alive.current&&owner.current===getToken()){setCapability(value);if(!attempt.current)setMessage(value.available?'':value.reason||'此上传渠道尚不可用');}}).catch(e=>{if(alive.current&&!abort.signal.aborted)setMessage(e.message||'无法读取上传配置');});return()=>{alive.current=false;epoch.current++;abort.abort();controller.current?.abort();};},[]);
 function finish(op,checked){
  setNotFound(false);setSafeNew(checked&&op.status==='failed'&&op.safeToStartNew===true);
  attempt.current.operation=op;setOperation(op);setVerified(true);
  if(op.status==='succeeded'){
   if(!matchesDraft()){setMessage('已核对上次上传成功，但当前草稿的图片或名称已变化；未写入当前草稿。请恢复上次上传草稿后再核对。');return;}
   ownerAttempts(owner.current).delete(slot.current);
   onSuccess({sora2Info:op.result.assetId,subjectUploadKey:op.requestKey});setMessage('上传成功，已填入当前草稿；点击应用后才保存档案。');
  }else setMessage((op.message||'上传尚未完成')+(op.status==='needs_review'?'；创建结果不明，需要人工核对，请勿重新创建素材。':''));
 }
 async function run(check=false){
  if(busy.current||disabled)return;
  if(owner.current!==getToken()){setMessage('账号已变化，请关闭并重新打开档案。');return;}
  if(!check&&!attempt.current){
   if(!capability?.available)return;
   if(!editor.draft.name.trim()){setMessage('请输入角色名称');return;}
   if(!editor.file&&!editor.entry?.image?.url){setMessage('请选择角色图片');return;}
   const input={requestKey:crypto.randomUUID(),name:editor.draft.name,description:editor.draft.description,groupId:editor.entry?.subjectUpload?.groupId||'',...(editor.file?{file:editor.file}:{entryId:editor.entry.id,expectedRevision:editor.entry.revision})};
   attempt.current={input};ownerAttempts(owner.current).set(slot.current,attempt.current);
  }else if(!check&&(!notFound&&(!verified||!matchesDraft()||attempt.current.operation?.status!=='paused'))){setMessage('请先核对上次上传，不能重复创建素材。');return;}
  if(!attempt.current)return;
  const version=++epoch.current,abort=new AbortController();controller.current=abort;busy.current=true;setWorking(true);setVerified(false);setNotFound(false);setSafeNew(false);onBusyChange(true);setMessage(check?'正在核对本次上传…':'正在上传真人素材…');
  try{const input=attempt.current.input,op=await(check?getComicSubjectUpload(input.requestKey,{signal:abort.signal}):createComicSubjectUpload(input,{signal:abort.signal}));if(current(version)){if(op.requestKey!==input.requestKey)throw Error('上传请求编号不一致，请人工核对');finish(op,check);}}
  catch(e){if(current(version)){if(check&&e.status===404){setNotFound(true);setMessage('未找到本次操作记录，请求可能仍在途；不能确认远端是否已创建。可按原请求编号和原输入重试本次请求。');}else setMessage((e.message||'上传结果尚未确认')+'；请核对本次上传，不要重复创建素材。');}}
  finally{if(current(version)){busy.current=false;setWorking(false);onBusyChange(false);}}
 }
 function stop(){setVerified(false);epoch.current++;controller.current?.abort();busy.current=false;setWorking(false);onBusyChange(false);setMessage('已停止等待。远端已经创建的素材无法撤销，请核对本次上传；旧存档仍保留。');}
 function modifyAfterRejection(){
  if(owner.current!==getToken()||!safeNew||!verified||operation?.status!=='failed'||operation.safeToStartNew!==true)return;
  ownerAttempts(owner.current).delete(slot.current);attempt.current=null;setOperation(null);setVerified(false);setSafeNew(false);setNotFound(false);setMessage('已确认上次请求被拒绝，旧请求不能再创建素材。请修改图片或名称后重新上传。');
 }
 function restore(){if(owner.current!==getToken()||!canRestore)return;const saved=attempt.current;onRestore?.({file:saved.input.file||null,draft:{...editor.draft,name:saved.input.name,description:saved.input.description}});setMessage('已恢复上次上传的名称、描述与图片；主体存档保留当前内容，请核对本次上传。');}
 return <span data-subject-upload style={{display:'inline-flex',gap:8,flexWrap:'wrap',alignItems:'center'}}>
  {!attempt.current?<button type="button" disabled={disabled||!capability?.available} onClick={()=>void run()}><OriginalIcon name="CloudUploadOutline" size={14}/>{capability?.label||'上传真人素材'}</button>:null}
  {working?<button type="button" onClick={stop}>停止等待</button>:attempt.current?<><button type="button" disabled={disabled} onClick={()=>void run(true)}>核对本次上传</button>{!matchesDraft()&&onRestore?<button type="button" disabled={disabled||!canRestore} onClick={restore}>恢复上次上传草稿</button>:null}{notFound?<button type="button" disabled={disabled} onClick={()=>void run(false)}>重试本次请求</button>:null}{safeNew?<button type="button" disabled={disabled} onClick={modifyAfterRejection}>修改后重新上传</button>:null}{verified&&operation?.status==='paused'?<button type="button" disabled={disabled||!matchesDraft()} onClick={()=>void run(false)}>继续本次上传</button>:null}</>:null}
  {attempt.current&&!matchesDraft()&&!canRestore?<span role="status">源档案图片版本已变化，只可核对上次上传，不能写入当前草稿。</span>:null}
  {message?<span role="status">{message}</span>:null}
 </span>;
}
