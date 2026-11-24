import {Button,Text,View} from '@tarojs/components';
import {useEffect,useRef,useState} from 'react';
import {apiRequest} from '../../../services/api';
import {createNovelClient} from '../../../services/novelCreation';
import {NOVEL_MODES,NOVEL_STATUS,novelUID,type NovelMode,type NovelRecord} from '../../../services/novelCreationInputs';
import type {NovelJournal} from '../../../services/novelRecovery';
import {NovelBack,NovelField,NovelSelect,NovelNotice,THINKING,novelConfirm,useNovelOperations,useNovelRead} from './shared';
const ended=new Set(['completed','rejected','invalid_output']);
export default function NovelStudio({mode,sessionId,models,journal,onBack,onTransition}:{mode:NovelMode;sessionId:string;models:NovelRecord[];journal:NovelJournal;onBack:()=>void;onTransition:(result:NovelRecord)=>void}){
  const key=`studio:${mode}:${sessionId||'new'}`,saved=journal.read(key);
  const [id,setId]=useState(sessionId||saved?.intent?.sessionId||saved?.sessionId||'');
  const [draft,setDraft]=useState<NovelRecord>(()=>saved?.draft||{content:'',modelId:models[0]?.id||'',thinkingLevel:'standard',contextBudget:120000});
  const [intent,setIntent]=useState<NovelRecord|null>(()=>saved?.intent||null),[snapshot,setSnapshot]=useState<NovelRecord|null>(null),[settings,setSettings]=useState(false);
  const [checkedAction,setCheckedAction]=useState('');
  const op=useNovelOperations(),client=createNovelClient(apiRequest),latest=useRef({id,draft,intent});latest.current={id,draft,intent};
  const save=(pending:NovelRecord|null,session=id,nextDraft=draft)=>{const value={sessionId:session,draft:nextDraft,intent:pending};journal.save(key,value);if(session)journal.save(`studio:${mode}:${session}`,value);};
  const accept=(value:NovelRecord,pending=latest.current.intent)=>{
    if(!value?.session?.id)throw Error('未返回有效讨论会话，请核对原请求');
    if(latest.current.id&&value.session.id!==latest.current.id)throw Error('讨论会话不匹配，未替换当前内容');
    setSnapshot(value);setId(value.session.id);
    if(pending&&value.turn?.idempotencyKey===pending.messageKey&&ended.has(value.turn.state)){
      const next=value.turn.state==='completed'?{...latest.current.draft,content:''}:latest.current.draft;
      save(null,value.session.id,next);setIntent(null);setDraft(next);
    }
  };
  const refresh=useNovelRead(async()=>latest.current.id?client.studio(latest.current.id):null,value=>{if(value)accept(value);},value=>['prepared','may_have_sent','received'].includes(value?.turn?.state),cause=>op.setError(cause.message||'讨论状态读取失败'));
  useEffect(()=>{if(id)void refresh();},[id]);
  const change=(patch:NovelRecord)=>{if(intent||op.busy)return;const next={...draft,...patch};setDraft(next);try{save(null,id,next);op.setError('');}catch(cause:any){op.setError(cause.message);}};
  const prepareSession=async(pending:NovelRecord)=>{
    if(pending.sessionId)return pending;
    const created=await client.createStudio(mode,pending.creationKey);if(!op.current())return null;if(!created?.session?.id)throw Error('讨论会话未确认，原编号已保存');
    const next={...pending,sessionId:created.session.id};save(next,created.session.id);setId(created.session.id);setIntent(next);return next;
  };
  const send=(resume=false)=>op.run(intent?'核对原消息':'发送创作要求',async()=>{
    let pending=intent;
    if(!pending){if(!draft.content.trim())throw Error('请输入创作要求');if(!models.some(item=>item.id===draft.modelId))throw Error('请选择当前可用的文本模型');
      pending={mode,creationKey:novelUID(),messageKey:novelUID(),sessionId:id,input:{content:draft.content.trim(),modelId:draft.modelId,thinkingLevel:draft.thinkingLevel,contextBudget:Number(draft.contextBudget)||120000}};save(pending);setIntent(pending);
    }else if(!resume){
      pending=await prepareSession(pending);if(!pending||!op.current())return;
      const value=await client.studio(pending.sessionId);if(!op.current())return;accept(value,pending);
      if(value.turn?.idempotencyKey!==pending.messageKey)throw Error('会话已找回，此消息尚未确认发送。可点击“发送原消息”，沿用原编号。');
      return;
    }
    pending=await prepareSession(pending);if(!pending||!op.current())return;
    const value=await client.sendStudio(pending.sessionId,{...pending.input,idempotencyKey:pending.messageKey});if(op.current())accept(value,pending);
  });
  const action=snapshot?.pendingAction,actionPending=action&&['pending','proposed'].includes(action.status),actionKey=`studio-action:${id}:${action?.id||'none'}`;
  const transition=(value:NovelRecord)=>{if(value.assistedAction||value.project?.id)onTransition(value);else throw Error('确认结果尚未完整，请刷新讨论核对，勿重复创建');};
  const act=(cancel=false)=>op.run(cancel?'取消方案':'确认方案',async()=>{
    if(!action?.id)return;
    const receipt=journal.read(actionKey);
    if(receipt&&!cancel&&checkedAction!==action.id){const value=await client.studio(id);if(!op.current())return;accept(value);const session=value.session;if(session?.projectId&&session.productionSessionId){journal.clear(actionKey);onTransition({project:{id:session.projectId,title:action.book?.title||action.title||'小说作品'},productionSession:{id:session.productionSessionId,mode},run:{id:session.productionRunId}});return;}if(value.pendingAction?.id===action.id&&['pending','proposed'].includes(value.pendingAction.status)){setCheckedAction(action.id);throw Error('原方案仍待确认，可再次点击确认，沿用相同方案与模型。');}throw Error('方案状态已变化，请刷新讨论核对。');}
    if(!cancel&&(!models.some(item=>item.id===draft.modelId)||!await novelConfirm('确认这个创作方案？','将按当前方案建立作品或继续生成，可能调用文本模型。')||!op.current()))return;
    const params=receipt||{actionId:action.id,modelId:draft.modelId,thinkingLevel:draft.thinkingLevel};
    journal.save(actionKey,params);
    const result=await client.studioAction(id,action.id,params.modelId,params.thinkingLevel,cancel);if(!op.current())return;
    if(cancel){accept(result);journal.clear(actionKey);}else{transition(result);journal.clear(actionKey);}
  });
  return <><NovelBack title={NOVEL_MODES.find(item=>item.id===mode)?.name||'原创讨论'} onBack={onBack} disabled={Boolean(op.busy)}/><Text className="nw-caption">先讨论题材、人物和冲突，确认方案后才建立作品。</Text><NovelNotice busy={op.busy} error={op.error}/>
    <View className="nw-section"><Text>创作对话</Text><Button className="nw-link" disabled={!id||Boolean(op.busy)} onClick={()=>void refresh()}>刷新</Button></View>
    {!snapshot?.messages?.length?<View className="nw-empty"><Text>从一个想法开始</Text><Text className="nw-caption">写下题材、主角、冲突，或希望的故事结局。</Text></View>:(snapshot.messages as NovelRecord[]).map(item=><View className={`nw-message role-${item.role}`} key={item.id}><Text className="nw-label">{item.role==='user'?'我':'创作助手'}</Text><Text selectable className="nw-prose">{item.content}</Text></View>)}
    {snapshot?.turn?<Text className="nw-caption">{NOVEL_STATUS[snapshot.turn.state]||snapshot.turn.state}{snapshot.turn.errorMessage?` · ${snapshot.turn.errorMessage}`:''}</Text>:null}
    {actionPending?<View className="nw-card"><Text className="nw-subtitle">待确认的创作方案</Text>{Object.entries({...action.book,title:action.title||action.book?.title,direction:action.direction,summary:action.summary,targetWordCount:action.targetWordCount}).filter(([key,value])=>Boolean(value)&&!['id','status','type'].includes(key)).map(([key,value])=><View key={key}><Text className="nw-label">{({title:'作品名称',genre:'题材',world:'世界设定',protagonist:'主角',coreConflict:'核心冲突',direction:'创作方向',summary:'故事梗概',targetWordCount:'目标字数',targetChapterWords:'每章字数'} as NovelRecord)[key]||'方案内容'}</Text><Text className="nw-prose">{String(value)}</Text></View>)}<View className="nw-actions"><Button disabled={Boolean(op.busy)} onClick={()=>void act(true)}>取消方案</Button><Button className="nw-primary" disabled={Boolean(op.busy)} onClick={()=>void act()}>确认方案</Button></View></View>:null}
    <NovelField id="novel-studio-content" label="创作要求" value={draft.content} onChange={value=>change({content:value})} multiline disabled={Boolean(op.busy||intent)} hint="告诉我你想写怎样的故事"/>
    <NovelSelect label="文本模型" items={models} value={draft.modelId} onChange={value=>change({modelId:value})} disabled={Boolean(op.busy||intent)}/>
    <Button className="nw-link" onClick={()=>setSettings(!settings)}>思考与上下文设置</Button>{settings?<><NovelSelect label="思考要求" items={THINKING} value={draft.thinkingLevel} onChange={value=>change({thinkingLevel:value})} disabled={Boolean(op.busy||intent)}/><NovelField label="上下文字数预算" value={draft.contextBudget} onChange={value=>change({contextBudget:value})} number disabled={Boolean(op.busy||intent)}/></>:null}
    <View className="nw-dock"><Text className="nw-caption">{intent?'原请求已保留，不会自动重发':'发送会调用所选文本模型'}</Text><Button id="novel-studio-send" className="nw-primary" disabled={Boolean(op.busy)||(!intent&&!draft.content.trim())} onClick={()=>void send()}>{intent?'核对原消息':'发送'}</Button>{intent?.sessionId&&snapshot?.turn?.idempotencyKey!==intent.messageKey?<Button disabled={Boolean(op.busy)} onClick={()=>void send(true)}>发送原消息</Button>:null}</View>
  </>;
}
