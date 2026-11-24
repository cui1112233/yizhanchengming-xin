import Taro from '@tarojs/taro';
import {Button,Input,Text,View} from '@tarojs/components';
import {useEffect,useRef,useState} from 'react';
import BrandHeader from '../../../components/BrandHeader';
import AppIcon from '../../../components/AppIcon';
import {ThemedPage} from '../../../theme';
import {apiRequest} from '../../../services/api';
import {apiUrl} from '../../../config/runtime';
import {readAccountSession} from '../../../services/session';
import {onSessionInvalidated} from '../../../services/sessionBoundary';
import {createNovelClient} from '../../../services/novelCreation';
import {createNovelJournal,type NovelJournal} from '../../../services/novelRecovery';
import {NOVEL_MODES,isNovelStudio,type NovelMode,type NovelRecord} from '../../../services/novelCreationInputs';
import {useNovelOperations,NovelNotice} from './shared';
import NovelStudio from './NovelStudio';
import NovelCreate from './NovelCreate';
import NovelProject from './NovelProject';
import './style.scss';

export default function NovelPage(){
  const [projects,setProjects]=useState<NovelRecord[]>([]),[studios,setStudios]=useState<NovelRecord[]>([]),[models,setModels]=useState<NovelRecord[]>([]);
  const [journal,setJournal]=useState<NovelJournal|null>(null),[view,setView]=useState<NovelRecord>({kind:'library'}),[loaded,setLoaded]=useState(false),[query,setQuery]=useState('');
  const [archived,setArchived]=useState(false);
  const op=useNovelOperations(),epoch=useRef(0),client=createNovelClient(apiRequest);
  const reload=()=>op.run('读取作品',async()=>{
    const ticket=epoch.current;
    const results=await Promise.allSettled([client.projects(),client.studios(),apiRequest<any>('/api/models?kind=text'),readAccountSession()]);
    if(!op.current()||ticket!==epoch.current)return;
    if(results[0].status==='rejected')throw results[0].reason;
    setProjects(results[0].value);setLoaded(true);
    if(results[1].status==='fulfilled')setStudios(results[1].value);else op.setError('讨论列表读取失败，可刷新重试；作品列表已保留。');
    if(results[2].status==='fulfilled'){const data=results[2].value;const list=Array.isArray(data)?data:data?.models;if(!Array.isArray(list))throw Error('模型目录格式不正确');setModels(list.filter(item=>item.enabled!==false).map(item=>({...item,id:String(item.id||item.key),name:String(item.name||item.label||item.id)})));}
    else throw results[2].reason;
    if(results[3].status==='rejected')throw results[3].reason;
    const username=results[3].value.account?.username;if(!username)throw Error('尚未读到账号，创作提交暂不可用');
    setJournal(createNovelJournal({get:key=>Taro.getStorageSync(key),set:(key,value)=>Taro.setStorageSync(key,value),remove:key=>Taro.removeStorageSync(key)},`${apiUrl('/')}::${username}`));
  });
  useEffect(()=>{void reload();const off=onSessionInvalidated(()=>{epoch.current++;setView({kind:'library'});setJournal(null);setProjects([]);setStudios([]);setModels([]);setLoaded(false);op.setError('账号已切换，请返回功能页重新打开');});return off;},[]);
  const back=()=>{setView({kind:'library'});void reload();};
  const openMode=(mode:NovelMode)=>setView({kind:isNovelStudio(mode)?'studio':'create',mode});
  const accept=(result:NovelRecord)=>{const project=result.project||result.book;if(!project?.id)throw Error('作品创建结果尚未确认，请核对原提交');setProjects(value=>[project,...value.filter(item=>item.id!==project.id)]);setView({kind:'project',project,session:result.productionSession||result.session,run:result.run,mode:result.session?.mode||result.productionSession?.mode||'continuation'});};
  const active=projects.filter(item=>item.status!=='purging'&&(archived?item.status==='archived':item.status!=='archived')&&String(item.title||'').includes(query));
  return <ThemedPage className="mp-page nw-page" dynamicGlass={false}><BrandHeader title="小说创作" showBack/><View className="nw-content">
    {view.kind==='library'?<>
      <View className="nw-heading"><View><Text className="nw-title">写下你的下一个故事</Text><Text className="nw-caption">原创、仿写、番外与续写</Text></View><Button className="nw-link" disabled={Boolean(op.busy)} onClick={()=>void reload()}>刷新</Button></View>
      <NovelNotice busy={op.busy} error={op.error}/>
      <View className="nw-modes">{NOVEL_MODES.map((item,index)=><Button id={`novel-mode-${item.id}`} key={item.id} disabled={!journal||Boolean(op.busy)} className="nw-mode" onClick={()=>openMode(item.id)}><View className={`nw-mode-icon tone-${index%3}`}><AppIcon name={index<2?'file-pen-line':index<4?'book-open':index===4?'users-round':'notebook-tabs'} size={23}/></View><Text className="nw-mode-name">{item.name}</Text><Text className="nw-caption">{item.description}</Text></Button>)}</View>
      <View className="nw-actions"><Button className="nw-link" disabled={Boolean(op.busy)} onClick={()=>setArchived(value=>!value)}>{archived?'返回常用作品':'查看归档作品'}</Button></View><View className="nw-section"><Text>{archived?'归档作品':'我的作品'} · {loaded?active.length:'—'}</Text></View><Input id="novel-search" className="nw-input" value={query} onInput={event=>setQuery(event.detail.value)} placeholder="搜索作品名称"/>
      {!op.busy&&loaded&&!active.length?<View className="nw-empty"><AppIcon name="book-open" size={32}/><Text>{query?'没有找到匹配的作品':'还没有作品'}</Text><Text className="nw-caption">{query?'试试其他名称':'选择上方的创作方式开始构思'}</Text></View>:null}
      {active.map(item=><Button className="nw-library-row" key={item.id} onClick={()=>setView({kind:'project',project:item})}><AppIcon name="book-open" size={23}/><View><Text>{item.title||'未命名作品'}</Text><Text className="nw-caption">{item.genre||'小说作品'} · {item.updatedAt?new Date(item.updatedAt).toLocaleDateString():'时间未返回'}</Text></View><Text>打开</Text></Button>)}
      {studios.length?<><View className="nw-section"><Text>创作讨论</Text><Text className="nw-caption">讨论确认后才建立作品</Text></View>{studios.map(item=><Button className="nw-library-row" key={item.id} onClick={()=>setView({kind:'studio',mode:item.mode,session:item.id})}><AppIcon name="file-pen-line" size={22}/><View><Text>{item.title||'创作讨论'}</Text><Text className="nw-caption">{NOVEL_MODES.find(mode=>mode.id===item.mode)?.name||'原创讨论'}</Text></View><Text>继续</Text></Button>)}</>:null}
    </>:journal&&view.kind==='studio'?<NovelStudio key={`${view.mode}:${view.session||'new'}`} mode={view.mode} sessionId={view.session||''} models={models} journal={journal} onBack={back} onTransition={result=>{if(result.assistedAction)setView({kind:'create',mode:result.targetRoute==='imitation'?'long_imitation':'continuation'});else accept(result);}}/>:journal&&view.kind==='create'?<NovelCreate key={view.mode} mode={view.mode} projects={projects} models={models} journal={journal} onBack={back} onCreated={accept}/>:journal&&view.kind==='project'?<NovelProject key={view.project.id} project={view.project} initialSession={view.session} mode={view.mode} models={models} journal={journal} onBack={back} onCreate={openMode}/>:null}
  </View></ThemedPage>;
}
