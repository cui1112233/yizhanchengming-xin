import Taro from '@tarojs/taro';
import {Button,ScrollView,Text,View} from '@tarojs/components';
import {useEffect,useRef,useState} from 'react';
import {apiRequest} from '../../../services/api';
import {loadAuthenticatedMedia} from '../../../services/mediaCache';
import {createNovelClient,novelID} from '../../../services/novelCreation';
import {continuationMessage,NOVEL_MODES,NOVEL_STATUS,NOVEL_TERMINAL,novelUID,type NovelMode,type NovelRecord} from '../../../services/novelCreationInputs';
import type {NovelJournal} from '../../../services/novelRecovery';
import {NovelBack,NovelField,NovelNotice,NovelSelect,THINKING,novelConfirm,useNovelOperations,useNovelRead} from './shared';
const object=(value:any):NovelRecord=>{if(value&&typeof value==='object')return value;try{return JSON.parse(value||'{}');}catch{return {};}};
export default function NovelProject({project,initialSession,mode,models,journal,onBack,onCreate}:{project:NovelRecord;initialSession?:NovelRecord;mode?:NovelMode;models:NovelRecord[];journal:NovelJournal;onBack:()=>void;onCreate:(mode:NovelMode)=>void}){
  const [sessions,setSessions]=useState<NovelRecord[]>([]),[session,setSession]=useState(initialSession?.id||''),[snapshot,setSnapshot]=useState<NovelRecord|null>(null),[chapters,setChapters]=useState<NovelRecord[]>([]);
  const [tab,setTab]=useState('create'),[chapterId,setChapterId]=useState(''),[edit,setEdit]=useState<NovelRecord|null>(null),[versions,setVersions]=useState<NovelRecord[]|null>(null),[blueprint,setBlueprint]=useState<NovelRecord|null>(null),[short,setShort]=useState<NovelRecord|null>(null);
  const [sessionMode,setSessionMode]=useState<NovelMode>(mode||initialSession?.mode||'continuation');
  const cache=`project:${project.id}:${session||'new'}`,saved=journal.read(cache);
  const [draft,setDraft]=useState<NovelRecord>(()=>saved?.draft||{content:'',modelId:project.modelId||models[0]?.id||'',thinkingLevel:'standard',referenceText:'',existingSetting:''});
  const [intent,setIntent]=useState<NovelRecord|null>(()=>saved?.intent||null),[loaded,setLoaded]=useState(false);
  const op=useNovelOperations(),client=createNovelClient(apiRequest),latest=useRef({session,sessions,chapters,draft,intent});latest.current={session,sessions,chapters,draft,intent};
  const readonly=['archived','purging'].includes(project.status);
  const sessionValue=sessions.find(item=>item.id===session)||initialSession;
  const activeMode=sessionValue?.mode||sessionMode;
  const run=snapshot?.run?.id?snapshot.run:null,rt=object(run?.runtimeSnapshot),running=Boolean(run&&!NOVEL_TERMINAL.has(run.status));
  const shortMode=['short_original','short_imitation'].includes(activeMode);
  const persist=(pending:NovelRecord|null,next=draft)=>journal.save(cache,{draft:next,intent:pending});
  const load=async()=>{
    const currentSession=latest.current.session;
    const [list,content,value]=await Promise.all([client.sessions(project.id),client.chapters(project.id),currentSession?client.snapshot(currentSession):Promise.resolve(null)]);
    if(value?.session?.id&&value.session.id!==currentSession)throw Error('创作会话不匹配，未更新页面');
    let story=null;
    if(value?.run?.runtimeSnapshot&&object(value.run.runtimeSnapshot).shortStoryId||['short_original','short_imitation'].includes(value?.session?.mode||'')){
      try{story=(await client.call(`/projects/${novelID(project.id)}/short-story`)).story;}catch(cause:any){if(cause.statusCode!==404)throw cause;}
    }
    return {list,content,value,story};
  };
  const refresh=useNovelRead(load,result=>{
    setSessions(result.list);setChapters(result.content);setLoaded(true);setShort(result.story||null);if(result.value)setSnapshot(result.value);
    if(!latest.current.session){const selected=result.list.find((item:NovelRecord)=>!item.archivedAt);if(selected)setSession(selected.id);}
    const pending=latest.current.intent;
    if(pending&&result.value?.run?.idempotencyKey===pending.data.idempotencyKey){persist(null,{...latest.current.draft,content:''});setIntent(null);setDraft(value=>({...value,content:''}));}
  },result=>{const status=result.value?.run?.status;return Boolean(status&&!NOVEL_TERMINAL.has(status)&&!['awaiting_confirmation','paused'].includes(status));},cause=>op.setError(cause.message||'作品读取失败'));
  useEffect(()=>{const stored=journal.read(cache);setDraft(stored?.draft||{content:'',modelId:project.modelId||models[0]?.id||'',thinkingLevel:'standard',referenceText:'',existingSetting:''});setIntent(stored?.intent||null);setSnapshot(null);if(session)void refresh();},[session]);
  const change=(patch:NovelRecord)=>{if(op.busy||intent)return;const next={...draft,...patch};setDraft(next);try{persist(null,next);}catch(cause:any){op.setError(cause.message);}};
  const createSession=()=>op.run('创建创作会话',async()=>{if(readonly)return;const result=await client.createSession(project.id,sessionMode);if(!op.current())return;if(!result.session?.id)throw Error('会话创建尚未确认，请刷新已有会话核对');setSessions(value=>[result.session,...value]);setSession(result.session.id);});
  const send=(replay=false)=>op.run(intent?'核对原请求':'提交创作要求',async()=>{
    if(readonly)return;
    if(intent&&!replay){const value=await client.snapshot(intent.sessionId);if(!op.current())return;if(value.run?.idempotencyKey===intent.data.idempotencyKey){persist(null);setIntent(null);setSnapshot(value);await refresh();return;}if(value.run&&!NOVEL_TERMINAL.has(value.run.status))throw Error('当前会话仍有任务执行，请先核对它的结果');const checked={...intent,allowResubmit:true};persist(checked);setIntent(checked);throw Error('当前快照未找到原请求。可手动以原编号重试，不会更改参数。');}
    if(!session)throw Error('请先选择或创建创作会话');if(!draft.content.trim())throw Error('请输入创作要求');if(!models.some(item=>item.id===draft.modelId))throw Error('请选择当前可用文本模型');
    const input=activeMode==='continuation'?continuationMessage(chapters,draft.content,draft.modelId,novelUID()):{content:draft.content.trim(),modelId:draft.modelId,idempotencyKey:novelUID(),modeInputs:activeMode.includes('imitation')?{referenceText:draft.referenceText||''}:activeMode==='fanfiction'?{existingSetting:draft.existingSetting||''}:{}};
    const pending=replay&&intent?intent:{sessionId:session,data:{...input,projectId:project.id,thinkingLevel:draft.thinkingLevel,knowledgeIds:[],noteIds:[]}};
    persist(pending);setIntent(pending);const result=await client.send(session,pending.data);if(!op.current())return;
    if(!result.run?.id)throw Error('未确认创作任务，原请求已保存');persist(null,{...draft,content:''});setIntent(null);setDraft(value=>({...value,content:''}));setSnapshot(value=>({...value,run:result.run}));await refresh();
  });
  const runAction=(action:'confirm'|'retry'|'cancel')=>op.run('处理创作任务',async()=>{
    if(!run||readonly)return;
    if(!await novelConfirm(action==='cancel'?'停止后续步骤？':action==='retry'?'从失败处重试？':'确认后继续生成？',action==='cancel'?'已保存章节保留，已经发送给上游的请求可能仍在处理。':'将恢复本次创作，可能继续调用所选文本模型。')||!op.current())return;
    if(action==='confirm'&&rt.confirmationKind==='short_outline'){
      if(!short?.version)throw Error('短篇大纲尚未读取，请刷新核对');await client.call(`/runs/${novelID(run.id)}/confirm-short-story`,{version:short.version,modelId:draft.modelId,thinkingLevel:draft.thinkingLevel});
    }else await client.runAction(run.id,action,action==='confirm'&&blueprint?{blueprint}:{});
    if(op.current()){setBlueprint(null);await refresh();}
  });
  const chapter=chapters.find(item=>item.id===chapterId)||chapters[0];
  const saveChapter=()=>op.run('保存章节',async()=>{if(!edit||readonly)return;const result=await client.saveChapter(project.id,edit);if(!op.current())return;if(!result.chapter?.id)throw Error('保存结果未确认，请保留草稿再刷新');setEdit(null);await refresh();});
  const exportBook=()=>op.run('导出已保存正文',async()=>{if(!chapters.length)throw Error('还没有已保存的正文');let index:number;try{index=(await Taro.showActionSheet({itemList:['导出TXT','导出Markdown']})).tapIndex;}catch{return;}const format=index===0?'txt':'md';const file=await loadAuthenticatedMedia(`/api/novel-creation/projects/${novelID(project.id)}/export?format=${format}`,format);if(!op.current())return;if(typeof Taro.shareFileMessage!=='function')throw Error('当前微信版本不支持发送文件，可先复制章节正文');try{await Taro.shareFileMessage({filePath:file.path,fileName:`${String(project.title||'小说').replace(/[\\/:*?"<>|]/g,'_')}.${format}`});}catch(cause:any){if(!/cancel/i.test(cause.errMsg||''))throw cause;}});
  return <><NovelBack title={project.title||'小说作品'} onBack={onBack} disabled={Boolean(op.busy)}/><NovelNotice busy={op.busy} error={op.error}/>{readonly?<Text className="nw-notice">归档作品仅查看，不提交创作或修改。</Text>:null}
    <View className="nw-tabs"><Button className={tab==='create'?'is-selected':''} onClick={()=>setTab('create')}>创作与进度</Button><Button className={tab==='chapters'?'is-selected':''} onClick={()=>setTab('chapters')}>已保存章节 · {loaded?chapters.length:'—'}</Button></View>
    <Button className="nw-link" disabled={Boolean(op.busy)} onClick={()=>void refresh()}>刷新作品与结果</Button>
    {tab==='create'?<>
      <NovelSelect label="创作会话" items={sessions.map(item=>({...item,name:`${NOVEL_MODES.find(mode=>mode.id===item.mode)?.name||'创作'} · ${item.title||item.id}`}))} value={session} onChange={id=>{if(!op.busy){setSession(id);setEdit(null);}}} disabled={Boolean(op.busy||intent)}/>
      {!session&&!readonly?<View className="nw-card"><NovelSelect label="新会话模式" items={NOVEL_MODES.filter(item=>!item.id.startsWith('short_'))} value={sessionMode} onChange={id=>setSessionMode(id as NovelMode)} disabled={Boolean(op.busy)}/><Button disabled={Boolean(op.busy)} onClick={()=>void createSession()}>创建创作会话</Button></View>:null}
      {run?<View className="nw-card"><Text className="nw-subtitle">{NOVEL_STATUS[run.status]||'状态待确认'}</Text>{run.errorMessage?<Text className="nw-error">{run.errorMessage}</Text>:null}<Text className="nw-caption">只显示服务端实际阶段，不估算百分比。</Text>{rt.foundationPlan?<Text className="nw-prose">{rt.foundationPlan.title||rt.foundationPlan.coreConflict||''}</Text>:null}
        {run.status==='awaiting_confirmation'?<>{short?.outline?<><Text className="nw-label">短篇大纲</Text><Text className="nw-prose">{short.outline.summary||short.outline.title||short.title}</Text>{(short.outline.chapters||[]).map((item:NovelRecord,index:number)=><Text key={index} className="nw-prose">{item.title} {item.summary}</Text>)}</>:null}{rt.foundationPlan&&rt.confirmationKind!=='next_model_call'?<Button disabled={Boolean(op.busy)||readonly} onClick={()=>setBlueprint(JSON.parse(JSON.stringify(rt.foundationPlan)))}>查看与修改创作蓝图</Button>:null}<Button className="nw-primary" disabled={Boolean(op.busy)||readonly} onClick={()=>void runAction('confirm')}>确认后继续生成</Button></>:null}
        {run.errorCode!=='novel_upstream_result_unknown'&&(run.status==='failed'||['ready','save_pending'].includes(snapshot?.executionState))?<Button disabled={Boolean(op.busy)||readonly} onClick={()=>void runAction('retry')}>{snapshot?.executionState==='save_pending'?'恢复保存结果':'从失败处重试'}</Button>:null}
        {running?<Button disabled={Boolean(op.busy)||readonly} onClick={()=>void runAction('cancel')}>停止后续步骤</Button>:null}
      </View>:loaded?<Text className="nw-caption">当前会话还没有创作任务</Text>:null}
      {(snapshot?.messages||[]).map((item:NovelRecord)=><View key={item.id} className={`nw-message role-${item.role}`}><Text className="nw-label">{item.role==='user'?'我':'创作助手'}</Text><Text selectable className="nw-prose">{item.content}</Text></View>)}
      {shortMode?<><Text className="nw-caption">本篇的写作与审核由对应短篇任务执行；完成正文在“已保存章节”中查看。</Text><Button className="nw-row-button" onClick={()=>onCreate(activeMode as NovelMode)}>开始另一篇{activeMode==='short_imitation'?'短篇仿写':'短篇原创'}</Button></>:<>
      <NovelField id="novel-project-content" label="本次创作要求" value={draft.content} onChange={value=>change({content:value})} multiline disabled={Boolean(op.busy||intent||running)||readonly}/>
      {activeMode.includes('imitation')?<NovelField label="参考文本（已有创作资料可不填）" value={draft.referenceText} onChange={value=>change({referenceText:value})} multiline disabled={Boolean(op.busy||intent||running)||readonly}/>:activeMode==='fanfiction'?<NovelField label="既有人物与世界规则（已有番外设定可不填）" value={draft.existingSetting} onChange={value=>change({existingSetting:value})} multiline disabled={Boolean(op.busy||intent||running)||readonly}/>:null}
      <NovelSelect label="文本模型" items={models} value={draft.modelId} onChange={value=>change({modelId:value})} disabled={Boolean(op.busy||intent||running)||readonly}/><NovelSelect label="思考要求" items={THINKING} value={draft.thinkingLevel} onChange={value=>change({thinkingLevel:value})} disabled={Boolean(op.busy||intent||running)||readonly}/>
      <Button id="novel-project-send" className="nw-primary nw-wide" disabled={Boolean(op.busy)||readonly||(!intent&&(!session||running||!draft.content.trim()))} onClick={()=>void send()}>{intent?'核对原请求':'开始本次创作'}</Button>{intent?.allowResubmit?<Button disabled={Boolean(op.busy)||readonly} onClick={()=>void send(true)}>以原编号重试</Button>:null}<Text className="nw-caption">{activeMode==='continuation'?'前文从最新已保存章节读取。提交会调用文本模型。':'提交后按服务端流程执行，可能产生模型费用。'}</Text></>}
      {(snapshot?.events||[]).length?<View className="nw-card"><Text className="nw-label">最近执行记录</Text>{snapshot!.events.slice(-6).map((event:NovelRecord)=><Text className="nw-caption" key={event.eventId}>{event.message||event.type}</Text>)}</View>:null}
    </>:<>
      <View className="nw-heading"><Text className="nw-subtitle">正式章节</Text><Button className="nw-link" disabled={Boolean(op.busy)||!chapters.length} onClick={()=>void exportBook()}>导出整本</Button></View>
      {!chapters.length?<View className="nw-empty"><Text>还没有已保存的章节</Text><Text className="nw-caption">生成完成或导入正式章节后，在这里阅读。</Text></View>:<><NovelSelect label="选择章节" items={chapters} value={chapter?.id||''} onChange={id=>{setChapterId(id);setEdit(null);setVersions(null);}} disabled={Boolean(op.busy)}/><Text className="nw-caption">版本 {chapter?.version} · {chapter?.wordCount??chapter?.content?.length??0} 字</Text><Text className="nw-prose" selectable>{chapter?.content}</Text><View className="nw-actions"><Button onClick={()=>void Taro.setClipboardData({data:chapter.content})}>复制正文</Button><Button disabled={readonly||Boolean(op.busy)||running} onClick={()=>setEdit({...chapter})}>编辑正文</Button><Button disabled={Boolean(op.busy)} onClick={()=>void op.run('读取版本',async()=>{const result=await client.versions(project.id,chapter.id);if(op.current()){if(!Array.isArray(result.entries))throw Error('版本历史格式不正确');setVersions(result.entries);}})}>版本历史</Button></View></>}
      {versions?<View className="nw-card"><Text className="nw-subtitle">版本历史</Text>{versions.map(item=><View className="nw-version" key={item.version}><Text>版本 {item.version} · {item.title}</Text><Text className="nw-prose">{item.content}</Text><Button disabled={readonly||running||Boolean(op.busy)||item.version===chapter.version} onClick={()=>void op.run('恢复版本',async()=>{if(!await novelConfirm('恢复这个版本？','恢复会创建新版本，其他历史会保留。')||!op.current())return;await client.restoreChapter(project.id,chapter.id,chapter.version,item.version);if(op.current()){setVersions(null);await refresh();}})}>恢复为新版本</Button></View>)}</View>:null}
    </>}
    {edit||blueprint?<View className="nw-overlay"><View className="nw-mask" onClick={()=>{if(!op.busy){setEdit(null);setBlueprint(null);}}}/><View className="nw-sheet"><View className="nw-heading"><Text className="nw-title">{edit?'编辑已保存正文':'核对创作蓝图'}</Text><Button className="nw-link" disabled={Boolean(op.busy)} onClick={()=>{setEdit(null);setBlueprint(null);}}>取消</Button></View><ScrollView scrollY className="nw-sheet-body">{edit?<><NovelField label="章节标题" value={edit.title} onChange={value=>setEdit({...edit,title:value})}/><NovelField label="章节正文" value={edit.content} onChange={value=>setEdit({...edit,content:value})} multiline/><Text className="nw-caption">版本冲突时保留此草稿；保存会创建新版本。</Text><Button className="nw-primary" disabled={Boolean(op.busy)} onClick={()=>void saveChapter()}>保存新版本</Button></>:<>{[['title','作品名称'],['genre','题材'],['world','世界设定'],['coreConflict','核心冲突']].map(([key,label])=><NovelField key={key} label={label} value={blueprint![key]} onChange={value=>setBlueprint({...blueprint,[key]:value})} multiline={key==='world'||key==='coreConflict'}/>)}<Text className="nw-caption">人物与卷章安排保留服务端原方案。</Text><Button className="nw-primary" disabled={Boolean(op.busy)} onClick={()=>void runAction('confirm')}>确认蓝图并继续</Button></>}{op.error?<Text className="nw-error">{op.error}</Text>:null}</ScrollView></View></View>:null}
  </>;
}
