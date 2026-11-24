import Taro,{useDidHide,useDidShow} from '@tarojs/taro';
import {Button,Text,View} from '@tarojs/components';
import {useEffect,useRef,useState} from 'react';
import BrandHeader from '../../components/BrandHeader';
import {apiRequest} from '../../services/api';
import {loadBatchFactoryWorks,type BatchFactoryBatch} from '../../services/batchFactoryV11';
import {loadAuthenticatedMedia} from '../../services/mediaCache';
import {createProductionClients,waterImageModels,type Row} from '../../services/productionClients';
import {confirmReviewedSegmentation,createSegmentationReview,sourceRevisionKey} from '../../services/segmentationReview';
import {onSessionInvalidated,sessionRevision} from '../../services/sessionBoundary';
import {withThemedPage} from '../../theme';
import WaterLibrary from './WaterLibrary';
import WaterProject,{waterTaskEnded} from './WaterProject';
import WaterEditor,{WaterField,WaterGenerate} from './WaterEditor';
import './water-studio.scss';
const client=createProductionClients(apiRequest);
export type ShuihuoWorkspaceProps={projectId?:string};
async function confirm(title:string,content:string){return (await Taro.showModal({title,content,confirmText:'确认继续'})).confirm;}
function ShuihuoWorkspace({projectId=''}:ShuihuoWorkspaceProps){
 const [view,setView]=useState(projectId?'project':'library'),[selected,setSelected]=useState(projectId),[tab,setTab]=useState('source');
 const [projects,setProjects]=useState<Row[]>([]),[batches,setBatches]=useState<BatchFactoryBatch[]>([]),[models,setModels]=useState<Row[]>([]);
 const [catalogState,setCatalogState]=useState<'loading'|'ready'|'error'>('loading'),[batchState,setBatchState]=useState<'loading'|'ready'|'error'>('loading');
 const [catalogError,setCatalogError]=useState(''),[batchError,setBatchError]=useState(''),[modelError,setModelError]=useState(''),[error,setError]=useState(''),[busy,setBusy]=useState(''),[reading,setReading]=useState(false);
 const [detail,setDetail]=useState<Row|null>(null),[source,setSource]=useState(''),[name,setName]=useState(''),[newSource,setNewSource]=useState('');
 const [mode,setMode]=useState('paragraph'),[textModel,setTextModel]=useState(''),[candidates,setCandidates]=useState<Row[]>([]),[ticket,setTicket]=useState(0);
 const [editor,setEditor]=useState<Row|null>(null),[generate,setGenerate]=useState<Row|null>(null),[preview,setPreview]=useState<Row|null>(null);
 const [created,setCreated]=useState<Row|null>(null),[createUnknown,setCreateUnknown]=useState(false);
 const live=useRef({alive:true,visible:true,session:sessionRevision(),selected:projectId,read:0,list:0,lock:false,dirty:false,sourceBase:'',detail:null as Row|null,created:null as Row|null,seed:null as Row|null});
 const review=useRef(createSegmentationReview()),unknown=useRef(new Set<string>()),timer=useRef<ReturnType<typeof setTimeout>>();
 const current=()=>live.current.alive&&live.current.session===sessionRevision();
 const stop=()=>{if(timer.current)clearTimeout(timer.current);timer.current=undefined;};
 const choose=(id:string)=>{live.current.selected=id;live.current.read++;live.current.dirty=false;live.current.sourceBase='';live.current.detail=null;setReading(false);review.current.invalidate();setCandidates([]);setDetail(null);setEditor(null);setGenerate(null);setPreview(null);setError('');setTab('source');setSelected(id);setView('project');};
 const back=()=>{if(live.current.lock)return;stop();live.current.selected='';live.current.read++;live.current.dirty=false;review.current.invalidate();setSelected('');setDetail(null);setEditor(null);setGenerate(null);setCandidates([]);setError('');setView('library');void load();};
 async function load(){const version=++live.current.list;setCatalogState('loading');setBatchState('loading');setCatalogError('');setBatchError('');setModelError('');
  await Promise.all([
   client.waterProjects().then(rows=>{if(current()&&version===live.current.list){setProjects(rows);setCatalogState('ready');}}).catch(cause=>{if(current()&&version===live.current.list){setCatalogState('error');setCatalogError(cause.message);}}),
   loadBatchFactoryWorks({request:apiRequest}).then(rows=>{if(current()&&version===live.current.list){setBatches(rows);setBatchState('ready');}}).catch(cause=>{if(current()&&version===live.current.list){setBatchState('error');setBatchError(cause.message);}}),
   client.waterModels().then(rows=>{if(current()&&version===live.current.list)setModels(rows.filter(row=>row.enabled!==false));}).catch(cause=>{if(current()&&version===live.current.list){setModels([]);setModelError(cause.message);}}),
  ]);
 }
 function accept(next:Row,id:string){if(String(next.project.id)!==id)throw Error('项目编号不匹配，请重新打开作品');const old=live.current.detail;
  if(old&&sourceRevisionKey(old.project)!==sourceRevisionKey(next.project)){review.current.invalidate();setCandidates([]);}
  if(live.current.seed?.id===id){const draft=live.current.seed.source;live.current.seed=null;live.current.sourceBase=sourceRevisionKey(next.project);if(next.project.sourceText!==draft){setSource(draft);live.current.dirty=true;}else{setSource(next.project.sourceText||'');live.current.dirty=false;}}
  else if(!live.current.dirty){setSource(next.project.sourceText||'');live.current.sourceBase=sourceRevisionKey(next.project);}
  if(String(live.current.created?.id)===id&&newSource.trim()&&String(next.project.sourceText||'').trim()===newSource.trim()){live.current.created=null;setCreated(null);setName('');setNewSource('');}
  live.current.detail=next;setDetail(next);
 }
 async function read(){const id=live.current.selected;if(!id||live.current.lock||!current())return;const version=++live.current.read;setReading(true);try{const next=await client.waterDetail(id);if(current()&&live.current.visible&&version===live.current.read&&live.current.selected===id)accept(next,id);}catch(cause:any){if(current()&&version===live.current.read)setError(cause.message||'项目读取失败');}finally{if(current()&&version===live.current.read)setReading(false);}}
 async function run(label:string,work:()=>Promise<void|boolean>,key='',paid=false){if(live.current.lock||!current())return false;if(key&&unknown.current.has(key)){setError('原操作结果未知，请先刷新项目核对，勿重复提交');return false;}const id=live.current.selected;live.current.lock=true;live.current.read++;setReading(false);setBusy(label);setError('');
  try{if(await work()===false)return false;if(!current()||live.current.selected!==id)return false;const next=await client.waterDetail(id);if(!current()||live.current.selected!==id)return false;accept(next,id);return true;}
  catch(cause:any){if(current()&&live.current.selected===id){const uncertain=!(cause.statusCode>=400&&cause.statusCode<500&&![408,409,429].includes(cause.statusCode));if(key&&paid&&uncertain)unknown.current.add(key);setError((cause.message||'操作失败')+(key&&paid&&uncertain?'。结果未知，请刷新核对，勿重复提交。':''));}return false;}
  finally{live.current.lock=false;if(current())setBusy('');}
 }
 async function create(){if(live.current.lock||!current()||createUnknown||created)return;if(!name.trim()||!newSource.trim()){setError('请填写作品名称和字幕原文');return;}live.current.lock=true;setBusy('保存项目');setError('');live.current.created=null;
  try{const project=await client.createWater(name,newSource,item=>{live.current.created=item;});if(!current())return;setCreated(null);setName('');setNewSource('');choose(String(project.id));void load();}
  catch(cause:any){if(current()){const existing=cause.createdProject||live.current.created;if(existing){setCreated(existing);live.current.seed={id:String(existing.id),source:newSource};setError(cause.message);}else{const uncertain=!(cause.statusCode>=400&&cause.statusCode<500&&![408,409,429].includes(cause.statusCode));setCreateUnknown(uncertain);setError(`${cause.message||'创建失败'}${uncertain?'。创建结果尚未确认，请返回作品库核对同名作品，勿重复新建。':''}`);}}}
  finally{live.current.lock=false;if(current())setBusy('');}
 }
 async function saveSource(){if(!detail)return;const base=live.current.sourceBase||sourceRevisionKey(detail.project);const saved=await run('保存原文',async()=>{const fresh=await client.waterDetail(selected);if(sourceRevisionKey(fresh.project)!==base)throw Error('服务端原文已更新，当前草稿已保留，请复制后重新打开项目核对');await client.saveWaterSource(selected,source);live.current.dirty=false;review.current.invalidate();setCandidates([]);});if(saved&&String(created?.id)===selected){setCreated(null);setName('');setNewSource('');}}
 async function previewSegments(){if(!detail||live.current.dirty||source!==detail.project.sourceText){setError('请先保存原文，再生成候选分镜');return;}const sourceKey=sourceRevisionKey(detail.project);const candidateTicket=review.current.begin(sourceKey);await run('预览分镜',async()=>{
   if(mode==='smart'){if(!await confirm('使用AI智能分镜？','将使用系统已配置的文本模型，可能产生费用。')||!current())return false;}
   const next=mode==='smart'?await client.waterSmartCandidates(selected):await client.waterCandidates(selected);if(current()&&live.current.selected===selected&&review.current.current(candidateTicket,sourceKey)){setCandidates(next);setTicket(candidateTicket);setTab('segments');}
  },mode==='smart'?`smart:${selected}:${sourceKey}`:'',mode==='smart');}
 async function confirmCandidates(){if(!detail)return;await run('确认分镜',async()=>{if(live.current.dirty)throw Error('原文已修改，请保存后重新预览');if(detail.segments.length&&!await confirm('替换当前分镜？','确认后采用当前候选，现有正式分镜将被替换。'))return false;await confirmReviewedSegmentation({review:review.current,ticket,sourceKey:sourceRevisionKey(detail.project),read:()=>client.waterDetail(selected) as Promise<{project:Row}>,isCurrent:()=>current()&&live.current.selected===selected,submit:()=>client.confirmWaterCandidates(selected,candidates),clear:()=>setCandidates([])});});}
 async function saveEditor(){if(!editor)return;if(editor.editorKind==='asset'&&!String(editor.name||'').trim()){setError('请填写资产名称');return;}if(editor.editorKind==='candidate'){if(!String(editor.text||'').trim()){setError('请填写候选分镜正文');return;}setCandidates(rows=>rows.map((item,index)=>index===editor.candidateIndex?{text:editor.text,speaker:editor.speaker||'旁白'}:item));setEditor(null);return;}
  const value={...editor};delete value.editorKind;const saved=await run('保存编辑',async()=>{if(editor.editorKind==='asset'){const result=await client.saveWaterAsset(selected,value);const id=result?.asset?.id||result?.id;if(id&&current())setEditor(item=>item?{...item,id}:item);}else await client.saveWaterSegment(value);},editor.editorKind==='asset'&&!editor.id?'new-asset':'',editor.editorKind==='asset'&&!editor.id);if(saved)setEditor(null);
 }
 async function generateMedia(modelId:string,kind:string){const segment=generate;const chosen=waterImageModels(models).find(item=>String(item.id)===modelId);if(kind!=='image'||!segment?.confirmed||!chosen){setError('请选择已确认分镜与对应类型的可用模型');return;}
  const saved=await run('提交生成任务',async()=>{if(!await confirm('开始生成素材？','将调用所选模型，可能产生费用。生成结果以任务状态为准。')||!current())return false;await client.waterTask(selected,String(segment.id),String(chosen.modelId),kind);},`generate:${selected}:${segment.id}:${modelId}:${kind}`,true);if(saved){setGenerate(null);setTab('tasks');}
 }
 async function taskAction(task:Row,action:'retry'|'cancel'){await run(action==='retry'?'重试任务':'取消任务',async()=>{if(!await confirm(action==='retry'?'重试这条失败任务？':'取消这条任务？',action==='retry'?'可能重新调用原模型并产生费用。':'已发给上游的请求可能仍在处理，随后请刷新核对。')||!current())return false;await client.waterTaskAction(String(task.id),action);},`task:${task.id}:${action}`,true);}
 async function media(item:Row){const target=item.media||item;await run('读取素材',async()=>{const result=await loadAuthenticatedMedia(`/api/shuihuo-production/media/${encodeURIComponent(target.id)}/download`,target.kind==='video'?'mp4':target.kind==='audio'?'mp3':'png');if(current())setPreview({path:result.path,kind:target.kind});});}
 useEffect(()=>{void load();const off=onSessionInvalidated(()=>{live.current.read++;live.current.list++;live.current.detail=null;stop();review.current.invalidate();setProjects([]);setBatches([]);setModels([]);setDetail(null);setSource('');setNewSource('');setName('');setCandidates([]);setEditor(null);setGenerate(null);setPreview(null);setCreated(null);live.current.created=null;live.current.seed=null;setBusy('');setReading(false);setError('账号已切换，请返回功能页重新打开');});return()=>{off();live.current.alive=false;live.current.read++;live.current.list++;stop();};},[]);
 useEffect(()=>{if(projectId&&projectId!==live.current.selected)choose(projectId);},[projectId]);
 useEffect(()=>{if(selected)void read();return()=>{live.current.read++;stop();};},[selected]);
 useEffect(()=>{stop();if(detail&&detail.tasks.some((task:Row)=>!waterTaskEnded(task))&&!busy&&live.current.visible)timer.current=setTimeout(()=>void read(),3000);return stop;},[detail,selected,busy]);
 useDidShow(()=>{live.current.visible=true;void load();if(live.current.selected)void read();});useDidHide(()=>{live.current.visible=false;live.current.read++;live.current.list++;stop();});
 return <View className="mp-page mp-detail-page water-studio"><BrandHeader title="水货生产" showBack/><View className="water-content">{view!=='library'?<View className="water-heading"><Text className="water-title">{view==='create'?'新建漫剧':detail?.project.name||'项目工作区'}</Text><Button id="water-back" className="water-link" disabled={Boolean(busy)} onClick={back}>返回作品库</Button></View>:null}{error?<Text className="water-error">{error}</Text>:null}{busy?<Text className="water-note">正在{busy}…</Text>:null}
 {view==='library'?<><WaterLibrary projects={projects} batches={batches} waterState={catalogState} batchState={batchState} waterError={catalogError} batchError={batchError} onCreate={()=>{setError('');setView('create');}} onBatch={id=>void Taro.navigateTo({url:id?`/pages/batch/index?batchId=${encodeURIComponent(id)}`:'/pages/batch/index'})} onOpen={choose} onRefresh={()=>void load()}/>{modelError?<Text className="water-error">模型读取失败：{modelError}，可刷新作品库重试。</Text>:null}</>:view==='create'?<><Text className="water-note">粘贴字幕或小说原文，保存后预览并确认分镜。</Text><WaterField id="water-create-name" label="作品名称" value={name} onChange={setName} disabled={Boolean(busy)}/><WaterField id="water-create-source" label="字幕 / 小说原文" value={newSource} multiline onChange={setNewSource} disabled={Boolean(busy)}/><Text className="water-note">这里只保存项目，不自动调用AI生成素材。</Text><Button id="water-create-save" className="water-primary water-wide" disabled={Boolean(busy)||!name.trim()||!newSource.trim()||Boolean(created)||createUnknown} onClick={()=>void create()}>保存项目</Button>{created?<Button className="water-wide" onClick={()=>choose(String(created.id))}>打开已创建项目</Button>:null}</>:detail?<><Button className="water-link" disabled={Boolean(busy)} onClick={()=>void read()}>{reading?'正在读取项目…':'刷新项目'}</Button>{modelError?<Text className="water-error">模型读取失败：{modelError}</Text>:null}<WaterProject detail={detail} tab={tab} onTab={setTab} source={source} onSource={value=>{live.current.dirty=true;review.current.invalidate();setCandidates([]);setSource(value);}} sourceDirty={live.current.dirty} models={models} mode={mode} onMode={setMode} textModel={textModel} onTextModel={setTextModel} candidates={candidates} busy={Boolean(busy)} onSaveSource={()=>void saveSource()} onPreview={()=>void previewSegments()} onConfirm={()=>void confirmCandidates()} onEdit={value=>{setError('');setEditor(value);}} onGenerate={value=>{setError('');setGenerate(value);}} onTask={(task,action)=>void taskAction(task,action)} onMedia={item=>void media(item)} preview={preview} onRefresh={()=>void read()}/></>:<View className="water-empty"><Text>{reading?'正在读取项目…':'项目尚未读取成功'}</Text><Button className="water-link" onClick={()=>void read()}>重新读取项目</Button></View>}
 {editor?<WaterEditor edit={editor} onChange={setEditor} onSave={()=>void saveEditor()} onClose={()=>{setEditor(null);setError('');}} busy={Boolean(busy)} error={error}/>:null}{generate?<WaterGenerate segment={generate} models={models} onSave={(id,kind)=>void generateMedia(id,kind)} onClose={()=>{setGenerate(null);setError('');}} busy={Boolean(busy)} error={error}/>:null}
 </View></View>;
}
export default withThemedPage(ShuihuoWorkspace);
