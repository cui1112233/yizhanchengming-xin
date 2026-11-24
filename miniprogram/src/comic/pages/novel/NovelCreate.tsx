import {Button,Text,View} from '@tarojs/components';
import {useEffect,useState} from 'react';
import {apiRequest} from '../../../services/api';
import {createNovelClient} from '../../../services/novelCreation';
import {buildNovelCreation,NOVEL_MODES,novelUID,type NovelMode,type NovelRecord} from '../../../services/novelCreationInputs';
import type {NovelJournal} from '../../../services/novelRecovery';
import {NovelBack,NovelField,NovelSelect,NovelNotice,THINKING,novelConfirm,useNovelOperations} from './shared';

export default function NovelCreate({mode,projects,models,journal,onBack,onCreated}:{mode:NovelMode;projects:NovelRecord[];models:NovelRecord[];journal:NovelJournal;onBack:()=>void;onCreated:(result:NovelRecord)=>void}){
  const key=`create:${mode}`,saved=journal.read(key);
  const [draft,setDraft]=useState<NovelRecord>(()=>saved?.draft||{title:'',referenceText:'',modelId:models[0]?.id||'',thinkingLevel:'standard',targetWordCount:mode==='long_imitation'?5000:13000,targetChapters:mode==='long_imitation'?100:1,spinoffType:'sequel',knowledgeBookIds:[],memoReferences:[]});
  const [intent,setIntent]=useState<NovelRecord|null>(()=>saved?.intent||null),[styles,setStyles]=useState<NovelRecord[]>([]),[memos,setMemos]=useState<NovelRecord[]>([]),[references,setReferences]=useState(false),[metadataError,setMetadataError]=useState('');
  const op=useNovelOperations(),client=createNovelClient(apiRequest),frozen=Boolean(op.busy||intent),long=mode==='long_imitation',imitation=long||mode==='short_imitation';
  const persist=(nextDraft:NovelRecord,nextIntent:NovelRecord|null)=>journal.save(key,{draft:nextDraft,intent:nextIntent});
  const change=(patch:NovelRecord)=>{if(frozen)return;const next={...draft,...patch};setDraft(next);try{persist(next,null);op.setError('');}catch(cause:any){op.setError(cause.message);}};
  const metadata=async()=>{const values=await Promise.allSettled([client.styles(),client.memos()]);if(!op.current())return;if(values[0].status==='fulfilled')setStyles(values[0].value);if(values[1].status==='fulfilled')setMemos(values[1].value);if(values.some(value=>value.status==='rejected'))setMetadataError('部分引用读取失败，已保留选择，可稍后重新读取。');else setMetadataError('');};
  useEffect(()=>{void metadata();},[]);
  const accept=(result:NovelRecord)=>{
    if(!(result?.project?.id||result?.book?.id))throw Error('作品返回不完整，原提交已保留，请核对');
    if(mode!=='continuation'&&(!result.session?.id||!result.run?.id))throw Error('创作返回不完整，原提交已保留');
    persist(draft,null);setIntent(null);onCreated(result);
  };
  const submit=(replay=false)=>op.run(intent?'核对原提交':'提交创作',async()=>{
    let pending=intent;
    if(!pending){
      if(mode!=='continuation'&&!models.some(item=>item.id===draft.modelId))throw Error('请选择当前可用的文本模型');
      const payload=buildNovelCreation(mode,draft,novelUID());
      if(!await novelConfirm(mode==='continuation'?'导入这些正式章节？':'开始本次创作？',mode==='continuation'?'只保存粘贴的正式章节，导入后再选择是否续写。':'将使用所选文本模型，可能产生费用。提交后可查看真实执行状态。')||!op.current())return;
      pending={...payload};persist(draft,pending);setIntent(pending);
    }
    if(!op.current())return;
    if(intent&&!replay&&['/long-imitation','/short-stories'].includes(pending.path)){
      try{const result=await client.call(`${pending.path}/requests/${encodeURIComponent(pending.data.idempotencyKey)}`);if(op.current())accept(result);}
      catch(cause:any){if(op.current()&&cause.statusCode===404){const next={...pending,allowResubmit:true};persist(draft,next);setIntent(next);throw Error('未找到原请求，可明确点击“以原编号重试”，不会更换参数。');}throw cause;}
    }else{let result;try{result=await client.call(pending.path,pending.data);}catch(cause:any){if(op.current()&&cause.statusCode>=400&&cause.statusCode<500&&![408,409,429].includes(cause.statusCode)){persist(draft,null);setIntent(null);}throw cause;}if(op.current())accept(result);}
  });
  const choices=projects.filter(item=>!['archived','purging'].includes(item.status));
  const field=(name:string,label:string,multiline=false,number=false,hint='')=><NovelField key={name} id={`novel-create-${name}`} label={label} value={draft[name]} onChange={value=>change({[name]:value})} multiline={multiline} number={number} disabled={frozen} hint={hint}/>;
  return <><NovelBack title={NOVEL_MODES.find(item=>item.id===mode)?.name||'小说创作'} onBack={onBack} disabled={Boolean(op.busy)}/><Text className="nw-caption">{mode==='continuation'?'将已写好的章节导入作品，保存后再续写。':mode==='fanfiction'?'选择正传，再填写你希望展开的番外。':long?'参考原文风格，创作全新长篇故事。':'参考原文风格，创作完整短篇小说。'}</Text><NovelNotice busy={op.busy} error={op.error}/>
    {intent?<View className="nw-notice"><Text>上次提交尚未核对，原参数已锁定。</Text><Text className="nw-caption">返回或刷新不会自动重复调用模型，请先核对原请求。</Text></View>:null}
    {mode==='fanfiction'?<NovelSelect label="正传书籍" items={choices} value={draft.parentBookId||''} disabled={frozen} onChange={id=>change({parentBookId:id})}/>:mode==='continuation'?<NovelSelect label="导入到作品" items={[{id:'',name:'新建作品'},...choices]} value={draft.bookId||''} disabled={frozen} onChange={id=>change({bookId:id})}/>:null}
    {field('title',mode==='fanfiction'?'番外书名':mode==='continuation'?'新作品名称':'暂定标题（可选）')}
    {mode==='fanfiction'?<>{<NovelSelect label="番外类型" items={[{id:'sequel',name:'后传'},{id:'prequel',name:'前传'},{id:'parallel',name:'平行线'},{id:'if',name:'If线'}]} value={draft.spinoffType} disabled={frozen} onChange={id=>change({spinoffType:id})}/>} {field('storyIdea','番外方向',true)}</>:null}
    {field('referenceText',imitation?'参考原文（至少100字）':mode==='continuation'?'正式章节正文':'参考原文（可选）',true,false,mode==='continuation'?'可粘贴带“第1章”等标题的正文，自动按章节拆分。':'粘贴需要参考的原文')}
    <Text className="nw-caption">{String(draft.referenceText||'').replace(/\s/g,'').length.toLocaleString()} 字（不计空白）</Text>
    {imitation?<><View className="nw-two">{field('targetWordCount',long?'每章字数':'全篇目标字数',false,true)}{field('targetChapters','目标章节数',false,true)}</View>{field('readerPreferences','读者偏好')}{long?<>{field('storyIdea','原创故事想法（可选）',true)}{field('selectedDirection','仿写方向（可选）',true)}</>:null}{field('supplementaryInfo','补充信息（可选）',true)}</>:null}
    {mode!=='continuation'?<><NovelSelect label="文本模型" items={models} value={draft.modelId||''} onChange={id=>change({modelId:id})} disabled={frozen}/><NovelSelect label="思考要求" items={THINKING} value={draft.thinkingLevel||'standard'} onChange={id=>change({thinkingLevel:id})} disabled={frozen}/></>:null}
    <Button className="nw-row-button" disabled={Boolean(op.busy)} onClick={()=>setReferences(!references)}>参考资料 · {(draft.knowledgeBookIds||[]).length} 个文风 · {(draft.memoReferences||[]).length} 个备忘录</Button>
    {references?<View className="nw-card"><Text className="nw-label">文风（可多选）</Text>{!styles.length?<Text className="nw-caption">暂无可用文风</Text>:styles.map(item=><Button key={item.id} disabled={frozen} className={(draft.knowledgeBookIds||[]).includes(item.id)?'nw-chip is-selected':'nw-chip'} onClick={()=>change({knowledgeBookIds:(draft.knowledgeBookIds||[]).includes(item.id)?draft.knowledgeBookIds.filter((id:string)=>id!==item.id):[...(draft.knowledgeBookIds||[]),item.id]})}>{item.title||item.name}</Button>)}{mode!=='continuation'&&mode!=='fanfiction'?<><Text className="nw-label">关联备忘录</Text>{!memos.length?<Text className="nw-caption">暂无可用备忘录</Text>:memos.map(item=>{const chosen=(draft.memoReferences||[]).some((ref:NovelRecord)=>ref.projectId===item.projectId&&ref.noteId===item.id);return <Button key={`${item.projectId}:${item.id}`} disabled={frozen} className={chosen?'nw-chip is-selected':'nw-chip'} onClick={()=>change({memoReferences:chosen?draft.memoReferences.filter((ref:NovelRecord)=>ref.projectId!==item.projectId||ref.noteId!==item.id):[...(draft.memoReferences||[]),{projectId:item.projectId,noteId:item.id}]})}>{item.title}</Button>;})}</>:null}{metadataError?<Text className="nw-error">{metadataError}</Text>:null}<Button className="nw-link" onClick={()=>void metadata()}>重新读取引用</Button></View>:null}
    <View className="nw-dock"><Text className="nw-caption">{intent?'保留原编号与原参数':mode==='continuation'?'只导入正文，不自动续写':'提交后可在作品中跟踪进度'}</Text><Button id="novel-create-submit" className="nw-primary" disabled={Boolean(op.busy)} onClick={()=>void submit()}>{intent?'核对原提交':mode==='continuation'?'保存正式章节':'开始创作'}</Button>{intent?.allowResubmit?<Button disabled={Boolean(op.busy)} onClick={()=>void submit(true)}>以原编号重试</Button>:null}</View>
  </>;
}
