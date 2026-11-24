import {Button,Text,View} from '@tarojs/components';
import {useState} from 'react';
import {apiRequest} from '../../../services/api';
import {createNovelManagement} from '../../../services/novelManagement';
import {createNovelClient} from '../../../services/novelCreation';
import {novelUID,type NovelRecord} from '../../../services/novelCreationInputs';
import type {NovelJournal} from '../../../services/novelRecovery';
import NovelMaterials from './NovelMaterials';
import NovelAutomation from './NovelAutomation';
import {NovelField,NovelSelect,NovelNotice,novelConfirm,useNovelOperations,useNovelRead} from './shared';
const client=createNovelManagement(apiRequest),creation=createNovelClient(apiRequest);
export default function NovelManagement({project,session,mode,models,references,initialTab='',initialPlanId='',initialPlanKind='',journal,onBack,onOpenSession,onApply,onRun,onLibrary}:{initialTab?:string;initialPlanId?:string;initialPlanKind?:string;project:NovelRecord;session:string;mode:string;models:NovelRecord[];references:NovelRecord;journal:NovelJournal;onBack:()=>void;onOpenSession:(id:string)=>void;onApply:(value:NovelRecord)=>void;onRun:(value:NovelRecord)=>void;onLibrary:()=>void}){
 const cache=`manage-main:${project.id}`,saved=journal.read(cache),op=useNovelOperations();
 const [tab,setTab]=useState(initialTab==='automation'?'automation':'materials'),[title,setTitle]=useState(project.title||''),[sessions,setSessions]=useState<NovelRecord[]>([]),[selected,setSelected]=useState(session),[sessionTitle,setSessionTitle]=useState(''),[continuity,setContinuity]=useState<NovelRecord|null>(null),[pending,setPending]=useState<NovelRecord|null>(saved?.pending||null);
 const readonly=['archived','purging'].includes(project.status);
 const refresh=useNovelRead(async()=>{const [rows,facts]=await Promise.allSettled([client.list(project.id,'sessions'),client.call(project.id,'/continuity')]);return{rows,facts};},r=>{if(r.rows.status==='fulfilled')setSessions(r.rows.value);else op.setError(`会话读取失败：${r.rows.reason.message}`);if(r.facts.status==='fulfilled')setContinuity(r.facts.value);else op.setError(`事实与记忆读取失败：${r.facts.reason.message}`);},()=>false,e=>op.setError(e.message));
 function savePending(value:NovelRecord|null){journal.save(cache,{pending:value});setPending(value);}
 async function maintenance(kind:'continuity_refresh'|'style_analysis',style='',replay=false){await op.run('提交分析任务',async()=>{
  if(readonly)return;if(pending&&!replay)throw Error('请先核对上次分析任务');if(!session)throw Error('请先选择已有创作会话');
  if(!await novelConfirm(kind==='style_analysis'?'分析这个文风？':'刷新作品事实与记忆？','使用作品已保存的模型配置，可能消耗额度；结果会保存在对应创作会话中。')||!op.current())return;
  const intent=replay&&pending?pending:{kind,style,session,key:novelUID()};savePending(intent);
  try{const r=await client.maintenance(project.id,intent.kind,intent.session,intent.key,intent.style);if(!r?.run?.id)throw Error('分析提交结果未确认，请核对原请求');if(op.current()){savePending(null);onRun(r.run);}}
  catch(e:any){if(op.current()&&e.statusCode>=400&&e.statusCode<500&&![408,409,429].includes(e.statusCode))savePending(null);throw e;}
 });}
 async function recover(){if(!pending)return;await op.run('核对原分析任务',async()=>{const result=await creation.snapshot(pending.session);if(!op.current())return;if(result.run?.idempotencyKey===pending.key){savePending(null);onRun(result.run);return;}savePending({...pending,allowResubmit:true});op.setError('快照未找到原编号，可手动以原编号核对重发，参数保持不变。');});}
 return <><View className="nw-heading"><Text className="nw-title">作品管理</Text><Button className="nw-link" disabled={Boolean(op.busy)} onClick={onBack}>返回作品</Button></View><NovelNotice busy={op.busy} error={op.error}/>
 <View className="nw-tabs">{[['materials','参考内容'],['automation','自动写作'],['settings','作品设置']].map(([id,label])=><Button key={id} className={tab===id?'is-selected':''} disabled={Boolean(op.busy)} onClick={()=>setTab(id)}>{label}</Button>)}</View>
 {pending?<View className="nw-card"><Text className="nw-error">原分析提交结果待核对，不会自动再发请求。</Text><Button disabled={Boolean(op.busy)} onClick={()=>void recover()}>核对原分析任务</Button>{pending.allowResubmit?<Button disabled={Boolean(op.busy)} onClick={()=>void maintenance(pending.kind,pending.style,true)}>以原编号重发分析</Button>:null}</View>:null}
 {tab==='materials'?<NovelMaterials project={project} session={session} references={references} journal={journal} onApply={onApply} onMaintenance={(kind,style)=>void maintenance(kind,style)}/>:tab==='automation'?<NovelAutomation initialPlanId={initialPlanId} initialKind={initialPlanKind==='batches'?'batches':'schedules'} project={project} session={session} mode={mode} models={models} references={references} journal={journal} onOpenSession={onOpenSession}/>:<>
  <View className="nw-card"><Text className="nw-subtitle">作品资料</Text><NovelField id="novel-manage-title" label="作品名称" value={title} onChange={setTitle} disabled={Boolean(op.busy)||readonly}/><Button className="nw-wide" disabled={Boolean(op.busy)||readonly||!title.trim()} onClick={()=>void op.run('保存作品名称',async()=>{const r=await client.rename(project.id,title);if(!r?.project?.id)throw Error('改名结果未确认，请刷新作品列表');if(op.current())onLibrary();})}>保存作品名称</Button><Button className="nw-link" disabled={Boolean(op.busy)} onClick={()=>void op.run(readonly?'恢复作品':'归档作品',async()=>{if(!await novelConfirm(readonly?'恢复这部作品？':'归档这部作品？',readonly?'恢复后可继续创作。':'作品将从常用列表移到归档列表，章节保留；可随时恢复。')||!op.current())return;await client.retain(project.id,readonly);if(op.current())onLibrary();})}>{readonly?'恢复作品':'归档作品'}</Button></View>
  <View className="nw-card"><Text className="nw-subtitle">创作会话</Text><NovelSelect label="已有会话" items={sessions} value={selected} onChange={id=>{setSelected(id);setSessionTitle(sessions.find(s=>s.id===id)?.title||'');}} disabled={Boolean(op.busy)}/><NovelField label="会话名称" value={sessionTitle} onChange={setSessionTitle} disabled={Boolean(op.busy)||readonly}/><View className="nw-actions"><Button disabled={Boolean(op.busy)||readonly||!selected||!sessionTitle.trim()} onClick={()=>void op.run('重命名会话',async()=>{await client.manageSession(project.id,selected,'rename',sessionTitle);if(op.current())await refresh();})}>保存会话名称</Button><Button disabled={Boolean(op.busy)||!selected} onClick={()=>onOpenSession(selected)}>打开会话</Button></View><Button className="nw-link" disabled={Boolean(op.busy)||readonly||!selected} onClick={()=>void op.run('归档会话',async()=>{if(!await novelConfirm('归档这个会话？','请先停止正在运行的任务，历史记录和章节仍保留。')||!op.current())return;await client.manageSession(project.id,selected,'archive','');if(op.current()){setSelected('');await refresh();}})}>归档当前会话</Button></View>
  <View className="nw-card"><Text className="nw-subtitle">事实与记忆</Text><Text className="nw-caption">{continuity?.dirtyFrom?`从第 ${continuity.dirtyFrom} 章起需要刷新`:'以服务端检查结果为准'}</Text><Button className="nw-wide" disabled={Boolean(op.busy)||readonly||!session||Boolean(pending)} onClick={()=>void maintenance('continuity_refresh')}>刷新事实与记忆</Button></View>
 </>}
 </>;
}
