export type NovelMode='long_original'|'short_original'|'long_imitation'|'short_imitation'|'fanfiction'|'continuation';
export type NovelRecord=Record<string,any>;
export const NOVEL_MODES:Array<{id:NovelMode;name:string;description:string}>=[
  {id:'long_original',name:'长篇原创',description:'先讨论设定，再开始章节写作'},
  {id:'short_original',name:'短篇原创',description:'围绕核心冲突，构思完整短篇'},
  {id:'long_imitation',name:'仿写长篇',description:'参考文风，构建全新的长篇故事'},
  {id:'short_imitation',name:'仿写短篇',description:'参考原文，创作完整短篇'},
  {id:'fanfiction',name:'番外创作',description:'关联正传，延伸人物与世界'},
  {id:'continuation',name:'续写创作',description:'导入正式章节，承接已有故事'},
];
export const isNovelStudio=(mode:string)=>mode==='long_original'||mode==='short_original';
export const novelUID=()=>`mini-${Date.now().toString(36)}-${Math.random().toString(36).slice(2,12)}`;
export function splitNovelChapters(value:string){
  const text=String(value||'').trim();if(!text)return [];
  const starts=[...text.matchAll(/第[0-9一二三四五六七八九十百千]+章/g)].map(m=>m.index!);
  const edges=[...new Set([0,...starts,text.length])], chapters:NovelRecord[]=[];
  for(let i=0;i<edges.length-1;i++){const segment=text.slice(edges[i],edges[i+1]).trim();if(!segment)continue;const title=segment.match(/^第[0-9一二三四五六七八九十百千]+章[^\n]*/)?.[0];const content=(title?segment.slice(title.length):segment).trim();if(content)chapters.push({title:title||`第${chapters.length+1}章`,content,wordCount:content.replace(/\s/g,'').length});}
  return chapters.length?chapters:[{title:'第1章',content:text,wordCount:text.replace(/\s/g,'').length}];
}
export function buildNovelCreation(mode:NovelMode,draft:NovelRecord,key:string){
  if(!key)throw Error('缺少提交编号');
  const referenceText=String(draft.referenceText||''),title=String(draft.title||'').trim();
  const shared={idempotencyKey:key,modelId:draft.modelId,thinkingLevel:draft.thinkingLevel||'standard',title,knowledgeBookIds:draft.knowledgeBookIds||[]};
  if(mode==='continuation'){
    if(!draft.bookId&&!title)throw Error('请填写新作品名称，或选择已有作品');
    const chapters=splitNovelChapters(referenceText);if(!chapters.length)throw Error('请粘贴正式章节正文');
    return {path:'/chapter-imports',data:{idempotencyKey:key,...(draft.bookId?{bookId:draft.bookId}:{title}),chapters,knowledgeBookIds:shared.knowledgeBookIds}};
  }
  if(!draft.modelId)throw Error('请选择可用的文本模型');
  if(mode==='fanfiction'){
    if(!draft.parentBookId)throw Error('请选择正传书籍');if(!title)throw Error('请填写番外书名');
    return {path:'/spinoffs',data:{...shared,parentBookId:draft.parentBookId,spinoffType:draft.spinoffType||'sequel',storyIdea:draft.storyIdea||'',referenceText}};
  }
  if(!['long_imitation','short_imitation'].includes(mode))throw Error('原创作品请先进入创作讨论');
  if(referenceText.replace(/\s/g,'').length<100)throw Error('参考原文至少需要100字（不计空白）');
  if(referenceText.length>2_000_000)throw Error('参考原文超过200万字元，请缩小范围');
  const words=Number(draft.targetWordCount),chapters=Number(draft.targetChapters),long=mode==='long_imitation';
  if(!Number.isInteger(words)||words<1000||words>(long?50000:500000))throw Error(long?'每章字数应为1000至50000的整数':'目标字数应为1000至500000的整数');
  if(!Number.isInteger(chapters)||chapters<1||chapters>(long?500:50))throw Error(long?'目标章节应为1至500的整数':'目标章节应为1至50的整数');
  const extra={referenceText,targetWordCount:words,targetChapters:chapters,readerPreferences:draft.readerPreferences||'',supplementaryInfo:draft.supplementaryInfo||'',memoReferences:draft.memoReferences||[]};
  return long?{path:'/long-imitation',data:{...shared,...extra,storyIdea:draft.storyIdea||'',selectedDirection:draft.selectedDirection||''}}:{path:'/short-stories',data:{...shared,...extra,mode,requirements:draft.requirements||'',inspirationTags:draft.inspirationTags||[],linkedSaveId:draft.linkedSaveId||'',pauseAfterOutline:Boolean(draft.pauseAfterOutline)}};
}
export function continuationMessage(chapters:NovelRecord[],content:string,modelId:string,key:string){
  const last=[...chapters].sort((a,b)=>Number(b.number)-Number(a.number))[0];if(!last?.content?.trim())throw Error('请先导入并保存正式章节');
  return {content:content.trim(),modelId,idempotencyKey:key,modeInputs:{previousText:last.content},thinkingLevel:'standard',knowledgeIds:[],noteIds:[]};
}
export const NOVEL_TERMINAL=new Set(['succeeded','succeeded_with_warnings','failed','cancelled']);
export const NOVEL_STATUS:Record<string,string>={queued:'排队中',planning:'生成创作蓝图',analyzing:'分析资料',writing:'写作正文',normalizing:'整理正文',auditing:'审核章节',revising:'修订章节',settling:'整理人物与伏笔',indexing:'建立记忆索引',awaiting_confirmation:'等待确认',paused:'已暂停',succeeded:'创作完成',succeeded_with_warnings:'完成，有审核提醒',failed:'创作失败',cancelled:'已停止',running:'正在处理',completed:'已完成',pending:'等待处理',provider_unknown:'上游结果待核对'};
