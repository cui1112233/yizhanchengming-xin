import {Alert,Button,Checkbox,Empty,Modal,Select,Space,Tag} from 'antd';
import {useEffect,useRef,useState} from 'react';
import {getComicChapters} from '../../shared/api/comicScreenplayWorkflow.js';
import {extractComicAssets} from '../../shared/api/comicScreenplay.js';
import {generateStoryboardPlan,listStoryboardGenerationAttempts} from './comic-screenplay-storyboard-api.js';
import {matchStoryboardChapter,storyboardDensityOptions} from './comic-screenplay-storyboard-settings.js';
import {createEpisodeRunner,validateEpisodeBatch,storyboardAttemptRows} from './comic-screenplay-batch-tools.js';
import {useComicConfirmation} from './comic-screenplay-confirm.js';
import OriginalIcon from './ComicScreenplayIcon.jsx';

export function ComicEpisodeBatchDialog({projectId,workspace,episodes,checkOnly=false,engine,modelId,canGenerate,promptAvailable=true,onGenerate,onChanged,onSaveDensity,onBusy,onClose}){
 const [selected,setSelected]=useState([]),[chapters,setChapters]=useState([]),[bindings,setBindings]=useState({}),[loading,setLoading]=useState(true),[error,setError]=useState(''),[busy,setBusy]=useState(false),[results,setResults]=useState([]),[task,setTask]=useState('generate'),[mode,setMode]=useState('new'),[density,setDensity]=useState(workspace?.project?.creative_config?.plan_density||'standard'),[original,setOriginal]=useState(false),[stopping,setStopping]=useState(false);
 const executing=useRef(false),stopRequested=useRef(false),runner=useRef(createEpisodeRunner()),alive=useRef(true),controller=useRef(null),busyCallback=useRef(onBusy);busyCallback.current=onBusy;
 const confirm=useComicConfirmation(projectId);
 useEffect(()=>{alive.current=true;getComicChapters(projectId).then(data=>{if(!alive.current)return;if(!Array.isArray(data?.state?.chapters))throw Error('章节读取失败');setChapters(data.state.chapters);setBindings(Object.fromEntries(episodes.map(ep=>[ep.no,matchStoryboardChapter(data.state.chapters,ep)])));}).catch(e=>{if(alive.current)setError(e.message);}).finally(()=>{if(alive.current)setLoading(false);});return()=>{alive.current=false;runner.current.stop();controller.current?.abort();busyCallback.current?.(false);};},[projectId]);
 const chosen=episodes.filter(ep=>selected.includes(ep.no)),locked=loading||busy,needsChapter=!checkOnly&&(task==='plan'||mode!=='none'||original);
 let reason='';try{validateEpisodeBatch({selected:chosen,modelId,mode:checkOnly?'none':mode,task,bindings,chapters,useOriginal:!checkOnly&&original});}catch(e){reason=e.message;}
 if(!checkOnly&&task==='generate'&&!promptAvailable)reason='所选分镜词不可用，请先重新选择并保存';
 const title=checkOnly?'资产检查 · 选择剧本集':'批量分镜 · 选择剧本集';
 async function execute(){
  if(executing.current||loading||reason||!canGenerate||!alive.current)return;
  executing.current=true;stopRequested.current=false;
  setBusy(true);busyCallback.current?.(true);setError('');setResults([]);setStopping(false);
  const captured=chosen.map(ep=>({...ep,chapterId:bindings[ep.no]}));
  try{
   if(!checkOnly&&(task==='plan'||mode==='new')&&density!==(workspace?.project?.creative_config?.plan_density||'standard')){const saved=await onSaveDensity?.(density);if(!saved)throw Error('规划密度没有保存，尚未开始');if(!alive.current)return;}
   if(stopRequested.current||!alive.current)return;
   await runner.current.run(captured,async ep=>{
    if(!alive.current)return false;
    if(checkOnly){let added=0;for(const type of ['character','scene','prop']){if(!alive.current)return false;const result=await extractComicAssets(projectId,{type,degree:workspace.project.asset_degree||'master',content:'# 第'+ep.no+'集 '+ep.title+'\n'+ep.content,model:modelId,missingOnly:true,concurrency:1});if(!alive.current)return false;if(result?.requires_confirmation)throw Error('需要拥有者确认模型服务');added+=(result.assets||[]).length;await onChanged?.();if(!alive.current)return false;if(result.partial_failure)throw Error('部分资产已保存，仍有失败：'+(result.warnings||[]).join('；'));}return '新增 '+added+' 项资产';}
    let planningId=ep.chapterId;
    if(task==='plan'||mode==='new'){
     const data=await getComicChapters(projectId);if(!alive.current)return false;
     const chapter=data.state?.chapters?.find(c=>c.id===ep.chapterId);if(!chapter)throw Error('对应章节已不存在，请刷新后重新选择');
     if(chapter.planning_status==='running')throw Error('章节规划仍在执行，请核对原请求');
     controller.current=new AbortController();
     const result=await generateStoryboardPlan(projectId,{chapter_id:ep.chapterId,model:modelId,expected_revision:data.state.revision,use_original_content_directly:original,idempotencyKey:crypto.randomUUID()},controller.current.signal);
     if(!alive.current)return false;if(!result?.chapter?.planning_text?.trim())throw Error('没有返回已保存的规划');await onChanged?.();if(!alive.current)return false;
    }
    if(task==='plan')return '规划已保存';
    const value=await onGenerate(ep.no,{...(ep.chapterId?{chapterId:ep.chapterId}:{}),usePlanning:mode!=='none',...(mode!=='none'?{planningChapterId:planningId}:{}),useOriginalContentDirectly:original});
    if(value===true)return '分镜已保存';if(value?.not_started){runner.current.stop();throw Error('已取消，未提交生成');}runner.current.stop();return false;
   },rows=>{if(alive.current)setResults(rows);});
  }catch(e){if(alive.current)setError(e.message||'批量操作失败');}finally{executing.current=false;if(alive.current){setBusy(false);busyCallback.current?.(false);controller.current=null;}}
 }
 function start(){if(reason||locked||!canGenerate)return;confirm({title:checkOnly?'确认检查遗漏资产？':'确认开始批量任务？',content:'已选 '+chosen.length+' 集。将使用当前保存的文本模型，可能消耗额度。'+(checkOnly?'只新增缺失人物、场景和道具，保留已有资产。':'所选集按当前准备选项逐集处理；已有结果可能更新，未确认请求会先核对。'),okText:'确认开始',cancelText:'取消',onOk:execute});}
 return <Modal className="comic-batch-tools-modal" title={title} width={760} open onCancel={()=>{if(!busy)onClose();}} closable={!busy} maskClosable={!busy} footer={<><Button disabled={busy} onClick={onClose}>关闭</Button>{busy?<Button disabled={stopping} danger onClick={()=>{stopRequested.current=true;runner.current.stop();setStopping(true);}}>{stopping?'正在等待当前集完成':'完成当前集后停止'}</Button>:<Button type="primary" disabled={!!reason||loading||!canGenerate} onClick={start}>开始{checkOnly?'检查':task==='plan'?'规划':'分镜'}</Button>}</>}>
  <Alert type="info" showIcon message={checkOnly?'检查遗漏的人物、场景和道具；同类型同名资产保留原有内容。':'先选择集数，再选择规划方式。成功结果逐集保存；停止不会撤销已完成结果。'}/>
  {error?<Alert type="error" showIcon message={error}/>:null}{reason?<p className="comic-helper-text">{reason}</p>:null}
  {!checkOnly?<Space wrap style={{margin:'16px 0'}}><Select aria-label="批量任务" disabled={locked} value={task} options={[{value:'generate',label:'生成分镜'},{value:'plan',label:'只生成规划'}]} onChange={setTask}/><Select aria-label="批量规划方式" disabled={locked||task==='plan'} value={mode} options={[{value:'new',label:'先生成新规划'},{value:'existing',label:'使用本集已有规划'},{value:'none',label:'不使用规划'}]} onChange={setMode}/><Select aria-label="批量规划密度" disabled={locked||task!=='plan'&&mode!=='new'} value={density} options={storyboardDensityOptions} onChange={setDensity}/><Checkbox disabled={locked} checked={original} onChange={e=>setOriginal(e.target.checked)}>直接使用章节原文</Checkbox></Space>:null}
  <div style={{margin:'16px 0'}}><Checkbox disabled={locked||!episodes.length} checked={selected.length===episodes.length&&!!selected.length} indeterminate={selected.length>0&&selected.length<episodes.length} onChange={e=>setSelected(e.target.checked?episodes.map(ep=>ep.no):[])}>全选</Checkbox><span>已选 {selected.length} 集</span></div>
  <div style={{maxHeight:'40vh',overflow:'auto'}}>{episodes.map(ep=><div className="comic-batch-episode-row" key={ep.no}><Checkbox disabled={locked} checked={selected.includes(ep.no)} onChange={e=>setSelected(rows=>e.target.checked?[...rows,ep.no]:rows.filter(no=>no!==ep.no))}>第{ep.no}集 · {ep.title}</Checkbox>{needsChapter?<Select aria-label={'第'+ep.no+'集对应章节'} disabled={locked} value={bindings[ep.no]||undefined} placeholder="选择实际对应章节" options={chapters.map(c=>({value:c.id,label:c.title||c.id}))} onChange={id=>setBindings(rows=>({...rows,[ep.no]:id}))}/>:null}</div>)}</div>
  {!episodes.length?<Empty description="没有已保存的剧本集"/>:null}
  {!!results.length?<div className="comic-batch-results" aria-live="polite">{results.map(r=><p key={r.item.no}><Tag color={r.status==='succeeded'?'green':r.status==='failed'?'red':'default'}>第{r.item.no}集 · {{running:'处理中',succeeded:'已完成',failed:'失败',not_started:'未开始',unconfirmed:'未确认，请核对'}[r.status]}</Tag>{r.error||typeof r.value==='string'&&r.value}</p>)}</div>:null}
 </Modal>;
}

export function ComicStoryboardFailures({projectId,engine,episodes,onRetry,onClose}){
 const [attempts,setAttempts]=useState([]),[loading,setLoading]=useState(true),[error,setError]=useState(''),[onlyFailed,setOnlyFailed]=useState(true);const ticket=useRef(0);
 async function reload(){const n=++ticket.current;setLoading(true);setError('');try{const data=await listStoryboardGenerationAttempts(projectId,{engine});if(n!==ticket.current)return;if(!Array.isArray(data?.attempts))throw Error('失败记录读取不完整');setAttempts(data.attempts);}catch(e){if(n===ticket.current)setError(e.message);}finally{if(n===ticket.current)setLoading(false);}}
 useEffect(()=>{void reload();return()=>{ticket.current++;};},[projectId,engine]);
 const rows=storyboardAttemptRows(attempts,engine).filter(r=>!onlyFailed||r.failed);
 return <Modal title="分镜生成记录" width={760} open onCancel={onClose} footer={<Button onClick={onClose}>关闭</Button>}><Space><Checkbox checked={onlyFailed} onChange={e=>setOnlyFailed(e.target.checked)}>只看失败或取消</Checkbox><Button loading={loading} onClick={()=>void reload()} icon={<OriginalIcon name="RefreshOutline" size={14}/>}>刷新记录</Button></Space>{error?<Alert type="error" showIcon message={error}/>:null}<p className="comic-helper-text">读取最近50个请求，按每集最近状态展示。状态未知时先核对，不重复生成。</p>{rows.map(r=><article className="comic-batch-episode-row" key={r.episodeNo}><div><b>第{r.episodeNo}集 · {r.status}</b><p>{r.error||'没有错误详情'}</p><small>{r.requestKey}</small></div><Button disabled={!episodes.some(ep=>ep.no===r.episodeNo)} onClick={()=>{onClose();onRetry(episodes.find(ep=>ep.no===r.episodeNo));}}>核对与重试</Button></article>)}{!loading&&!error&&!rows.length?<Empty description={onlyFailed?'没有失败或取消的生成记录':'没有生成记录'}/>:null}</Modal>;
}
