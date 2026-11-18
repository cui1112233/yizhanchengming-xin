import { Button, Input, Spin } from 'antd';
import { BookOpen, Sparkles } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { novelRequest } from '../../../shared/api/novelCreation.js';
import { availableModel, readSaved } from './contracts.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import './StudioChat.css';
import ComposerToolbar from './ComposerToolbar.jsx';

const waiting = new Set(['prepared', 'may_have_sent', 'received']);
const terminal = new Set(['completed', 'rejected', 'invalid_output']);
const keyFor = (owner,id) => `novel:studio:${encodeURIComponent(owner)}:${encodeURIComponent(id||'new')}`;
export default function StudioChat({owner,sessionId='',newSessionToken=0,models=[],onSessionSelected,onSessionCreated,onTransition}) {
 const scope=useWorkspaceScope(`${owner}:${sessionId}:${newSessionToken}`),key=keyFor(owner,sessionId);
 const [snapshot,setSnapshot]=useState(null),[localID,setLocalID]=useState(sessionId),[error,setError]=useState(''),[busy,setBusy]=useState(false),[loading,setLoading]=useState(false);
 const [draftState,setDraftState]=useState(null),[revision,setRevision]=useState(0),[actionBusy,setActionBusy]=useState(false);
 const flight=useRef(null),composition=useRef(false);
 const initial={content:'',modelId:models[0]?.value||'',thinkingLevel:'standard',contextBudget:64000};
 const draft=draftState?.key===key?draftState.value:{...initial,...readSaved(key+':draft',{})};
 const intent=readSaved(key+':intent',null);
 function change(patch){if(intent||busy)return;const value={...draft,...patch};setDraftState({key,value});try{localStorage.setItem(key+':draft',JSON.stringify(value));}catch{setError('草稿暂时无法保存，请保留正文后重试');}}
 function bindSession(pending,id){
  const bound={...pending,sessionId:id,aliases:[...new Set([...(pending.aliases||[key]),keyFor(owner,id)])]};
  const savedDraft={...draft,...pending.input};
  for(const alias of bound.aliases){localStorage.setItem(alias+':intent',JSON.stringify(bound));localStorage.setItem(alias+':draft',JSON.stringify(savedDraft));}
  return bound;
 }
 function rememberError(pending,text){for(const alias of pending?.aliases||[key])localStorage.setItem(alias+':error',JSON.stringify(text));}
 function clearIntent(pending){for(const alias of pending?.aliases||[key]){const old=readSaved(alias+':intent',null);if(old?.messageKey===pending?.messageKey)localStorage.removeItem(alias+':intent');}setRevision(v=>v+1);}
 function accept(next,pending=readSaved(key+':intent',null)) {
  if(!next?.session?.id)throw Error('会话返回内容无效，请核对原请求');setSnapshot(next);setLocalID(next.session.id);
  if(pending&&next.turn?.idempotencyKey===pending.messageKey&&terminal.has(next.turn.state)) {clearIntent(pending);if(next.turn.state==='completed'){for(const alias of pending.aliases||[key])localStorage.removeItem(alias+':error');setError('');const cleared={...draft,content:''};setDraftState({key,value:cleared});for(const alias of pending.aliases||[key])localStorage.setItem(alias+':draft',JSON.stringify(cleared));}}
 }
 async function check(){const id=localID||intent?.sessionId||sessionId;if(!id)return;try{await scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(id)}`,undefined,{signal:scope.signal}),next=>{accept(next);setError('');localStorage.setItem('novel:studio:selected:'+owner,id);if(sessionId!==id)onSessionSelected?.(id);});}catch(e){if(scope.isCurrent())setError(e.message);}}
 useEffect(()=>{setSnapshot(null);setLocalID(sessionId);setError(readSaved(key+':error',''));setBusy(false);setActionBusy(false);flight.current=null;if(!sessionId)return;setLoading(true);scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(sessionId)}`,undefined,{signal:scope.signal}),next=>accept(next)).catch(e=>setError(e.message)).finally(()=>{if(scope.isCurrent())setLoading(false);});},[owner,sessionId,newSessionToken]);
 useEffect(()=>{if(!localID||(!intent&&!waiting.has(snapshot?.turn?.state)))return;let timer,disposed=false;
  async function poll(){try{const next=await scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(localID)}`,undefined,{signal:scope.signal}));if(!next||disposed)return;accept(next);if(!waiting.has(next.turn?.state)&&next.turn?.state)return;}catch(e){if(scope.isCurrent())setError(e.message);}if(scope.isCurrent()&&!disposed)timer=setTimeout(poll,2000);}
  timer=setTimeout(poll,1200);return()=>{disposed=true;clearTimeout(timer);};
 },[localID,snapshot?.turn?.state,!!intent,revision]);
 async function send() {
  if(flight.current||!scope.isCurrent())return;
  if(intent){await check();return;}
  if(!draft.content.trim()){setError('请输入创作要求');return;}
  if(!availableModel(models,draft.modelId)){setError('请选择当前可用的文本模型');return;}
  const token={};flight.current=token;setBusy(true);setError('');localStorage.removeItem(key+':error');
  let pending={creationKey:crypto.randomUUID(),messageKey:crypto.randomUUID(),sessionId:sessionId||localID,input:{content:draft.content.trim(),modelId:draft.modelId,thinkingLevel:draft.thinkingLevel,contextBudget:draft.contextBudget},aliases:[key]};
  try{
   localStorage.setItem(key+':intent',JSON.stringify(pending));setRevision(v=>v+1);
   if(!pending.sessionId){const created=await scope.call(novelRequest('/studio-sessions',{idempotencyKey:pending.creationKey,mode:'long_original'},{signal:scope.signal}));if(!created)return;if(!created.session?.id)throw Error('创建会话返回无效，原编号已保留');pending=bindSession(pending,created.session.id);setLocalID(created.session.id);onSessionCreated?.(created.session);}
   const next=await scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(pending.sessionId)}/messages`,{...pending.input,idempotencyKey:pending.messageKey},{signal:scope.signal}));if(!next)return;accept(next,pending);
  }catch(e){if(scope.isCurrent()){setError(e.message||'回复暂时无法读取，原请求已保留');rememberError(pending,e.message||'回复暂时无法读取，原请求已保留');if(e.status>=400&&e.status<500&&![408,409,429].includes(e.status))clearIntent(pending);}}
  finally{if(flight.current===token)flight.current=null;if(scope.isCurrent()){setBusy(false);if(pending.sessionId){localStorage.setItem('novel:studio:selected:'+owner,pending.sessionId);onSessionSelected?.(pending.sessionId);}}}
 }
 async function recoverCreation(){
  if(!intent||intent.sessionId||flight.current)return;
  const token={};flight.current=token;setBusy(true);setError('');
  try{const created=await scope.call(novelRequest('/studio-sessions',{idempotencyKey:intent.creationKey,mode:'long_original'},{signal:scope.signal}));if(!created?.session?.id)return;const pending=bindSession(intent,created.session.id);onSessionCreated?.(created.session);setLocalID(created.session.id);onSessionSelected?.(created.session.id);setError('原会话已核对。该消息尚未发给模型，请明确选择发送原消息。');}
  catch(e){if(scope.isCurrent())setError(e.message);}finally{if(flight.current===token)flight.current=null;if(scope.isCurrent())setBusy(false);}
 }
 async function sendPrepared(){
  if(!intent?.sessionId||flight.current||snapshot?.turn?.idempotencyKey===intent.messageKey)return;
  const token={};flight.current=token;setBusy(true);setError('');
  try{const next=await scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(intent.sessionId)}/messages`,{...intent.input,idempotencyKey:intent.messageKey},{signal:scope.signal}));if(next)accept(next,intent);}catch(e){if(scope.isCurrent()){setError(e.message);rememberError(intent,e.message);if(e.status>=400&&e.status<500&&![408,409,429].includes(e.status))clearIntent(intent);}}finally{if(flight.current===token)flight.current=null;if(scope.isCurrent())setBusy(false);}
 }
 const action=snapshot?.pendingAction,actionPending=['pending','proposed'].includes(action?.status),canConfirm=action?.type==='create_book'&&actionPending;
 async function act(cancel=false){if(flight.current||actionBusy||!action?.id)return;const token={};flight.current=token;setActionBusy(true);setError('');try{const next=await scope.call(novelRequest(`/studio-sessions/${encodeURIComponent(localID)}/actions/${cancel?'cancel':'confirm'}`,cancel?{actionId:action.id}:{actionId:action.id,modelId:draft.modelId,thinkingLevel:draft.thinkingLevel},{signal:scope.signal}));if(!next)return;if(cancel)accept(next);else if(next.project?.id&&next.productionSession?.id&&next.run?.id)onTransition?.(next);else throw Error('创建结果尚不完整，请核对同一动作');}catch(e){if(scope.isCurrent())setError(e.message);}finally{if(flight.current===token)flight.current=null;if(scope.isCurrent())setActionBusy(false);}}
 const unconfirmed=intent&&snapshot?.turn?.idempotencyKey!==intent.messageKey;
 const linked=!!snapshot?.session?.projectId||action?.status==='confirmed';
 return <section className="novel-creation-conversation novel-studio-chat" aria-label="建书前会话">
  <header className="novel-chat-heading"><b>{snapshot?.session?.title||'长篇原创'}</b><span>创作对话</span></header>
  <div className="novel-creation-chat-scroll">
   {loading&&<Spin aria-label="读取创作会话"/>}
   {!snapshot?.messages?.length&&<div className="novel-creation-welcome"><BookOpen size={36}/><h3>开始你的故事</h3><p>告诉我你想写什么——题材、世界观、主角、核心冲突。</p><small>先讨论创作方向，确认后再创建作品。</small></div>}
   {(snapshot?.messages||[]).map(m=><article key={m.id} className={`novel-creation-message ${m.role}`}><b>{m.role==='user'?'我':'小Y'}</b><p>{m.content}</p></article>)}
   {action&&<section className="novel-studio-action" aria-label="创建作品建议"><h3><Sparkles size={17}/> {action.status==='confirmed'?'作品已创建':action.status==='cancelled'?'已取消':action.status==='superseded'?'已更新创作建议':'确认创建作品'}</h3><dl>{[['作品名称',action.book?.title],['题材',action.book?.genre],['世界设定',action.book?.world],['主角',action.book?.protagonist],['核心冲突',action.book?.coreConflict],['每章目标字数',action.book?.targetChapterWords]].map(([label,value])=>value!==undefined&&value!==''&&<div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
    {canConfirm&&<div className="novel-actions"><Button type="primary" loading={actionBusy} disabled={busy||!!intent||!availableModel(models,draft.modelId)} onClick={()=>act(false)}>确认创建作品</Button><Button disabled={busy||actionBusy} onClick={()=>act(true)}>取消创建</Button></div>}
    {action.status==='confirmed'&&<Button loading={actionBusy} onClick={()=>act(false)}>打开已创建作品</Button>}
   </section>}
   {(busy||waiting.has(snapshot?.turn?.state))&&<p role="status">小Y正在整理创作建议，尚未创建作品…</p>}
   {snapshot?.turn?.state==='unknown'&&<p role="alert">上游结果待确认，已暂停自动重试，不会再次发送同一消息。</p>}
   {snapshot?.turn?.errorMessage&&<p role="alert">{snapshot.turn.errorMessage}</p>}
   {error&&<div className="novel-studio-error" role="alert"><p>{error}</p>{localID&&!intent&&<Button onClick={check}>重新读取会话</Button>}</div>}
   {intent&&<div className="novel-studio-recovery" role="status"><p>原请求已保留，先核对结果，避免重复调用模型。</p><Button disabled={busy} onClick={intent.sessionId?check:recoverCreation}>核对原请求</Button>{unconfirmed&&snapshot?.session?.id===intent.sessionId&&<Button disabled={busy} onClick={sendPrepared}>发送原消息</Button>}</div>}
  </div>
  <div className="novel-creation-composer novel-creation-composer--compact">
   <Input.TextArea aria-label="创作要求" value={draft.content} disabled={linked||busy||!!intent||waiting.has(snapshot?.turn?.state)||loading} rows={3} placeholder="输入指令… Enter 发送，Shift+Enter 换行" onChange={e=>change({content:e.target.value})} onCompositionStart={()=>{composition.current=true;}} onCompositionEnd={()=>{composition.current=false;}} onKeyDown={e=>{if(e.key==='Enter'&&!e.shiftKey&&!e.nativeEvent.isComposing&&!composition.current){e.preventDefault();send();}}}/>
   <ComposerToolbar models={models} modelId={draft.modelId} onModelChange={modelId=>change({modelId})} thinkingLevel={draft.thinkingLevel} thinkingLabel="创作思考要求" onThinkingChange={thinkingLevel=>change({thinkingLevel})} contextBudget={draft.contextBudget} onBudgetChange={contextBudget=>change({contextBudget})} locked={busy||!!intent} sending={busy} disabled={linked||busy||!!intent||loading||waiting.has(snapshot?.turn?.state)} hasContent={!!draft.content.trim()} onSend={send}/>
  </div>
 </section>;
}
