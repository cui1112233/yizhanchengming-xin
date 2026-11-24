import {Button,Input,Text,View} from '@tarojs/components';
import {useState} from 'react';
import AppIcon from '../../components/AppIcon';
import BatchCover from '../batch/BatchCover';
import type {Row} from '../../services/productionClients';
type ReadState='loading'|'ready'|'error';
export default function WaterLibrary({projects,batches,waterState,batchState,waterError,batchError,onCreate,onBatch,onOpen,onRefresh}:{projects:Row[];batches:Row[];waterState:ReadState;batchState:ReadState;waterError:string;batchError:string;onCreate:()=>void;onBatch:(id?:string)=>void;onOpen:(id:string)=>void;onRefresh:()=>void}){
 const [query,setQuery]=useState('');
 const count=projects.length+batches.length,ready=waterState==='ready'&&batchState==='ready';
 const matches=(item:Row)=>String(item.name||item.title||item.id).toLowerCase().includes(query.trim().toLowerCase());
 const water=projects.filter(matches),batch=batches.filter(matches);
 return <><View className="water-heading"><View><Text className="water-title">我的生产作品</Text><Text className="water-note">从字幕到分镜，再到你的下一部作品</Text></View><Button className="water-link" onClick={onRefresh}>刷新作品库</Button></View>
 <View className="water-library-actions"><Button id="water-create" className="water-primary" onClick={onCreate}><AppIcon name="plus" size={18}/>新建漫剧</Button><Button id="water-batch" onClick={()=>onBatch()}>批量工厂</Button></View>
 <Input id="water-search" className="water-input" value={query} onInput={event=>setQuery(event.detail.value)} placeholder="搜索作品名称"/>
 <View className="water-section"><Text>个人作品</Text><Text className="water-note">{ready?`${count} 部`:`已读取 ${count} 部`}</Text></View>
 {waterState==='loading'?<Text className="water-note">正在读取作品库…</Text>:waterState==='error'?<Text className="water-error">作品库读取失败：{waterError}</Text>:null}
 {batchState==='loading'?<Text className="water-note">正在读取批量工厂作品…</Text>:batchState==='error'?<Text className="water-error">批量工厂作品读取失败：{batchError}</Text>:null}
 {water.map(item=><Button key={`water:${item.id}`} id={`water-project-${item.id}`} aria-label={item.name||String(item.id)} className="water-work" onClick={()=>onOpen(String(item.id))}><View className="water-work-icon"><AppIcon name="clapperboard" size={28}/></View><View className="water-work-copy"><Text>{item.name||'未命名漫剧'}</Text><Text className="water-note">漫剧作品{item.updatedAt?` · ${new Date(item.updatedAt).toLocaleDateString()}`:''}</Text></View><AppIcon name="chevron-right" size={16}/></Button>)}
 {batch.map(item=><Button key={`batch:${item.id}`} className="water-work" aria-label={item.title||item.name||item.id} onClick={()=>onBatch(item.id)}><View className="water-work-poster"><BatchCover url={String(item.coverUrl||item.cover?.url||'')} batchId={item.id}/></View><View className="water-work-copy"><Text>{item.title||item.name||'未命名作品'}</Text><Text className="water-note">批量作品 · {({comic:'漫画',text:'文字页',video:'视频'} as Row)[item.productionMode]||'内容制作'}</Text><Text className="water-note">{typeof item.bookCount==='number'&&Number.isFinite(item.bookCount)?`${item.bookCount} 本小说`:'打开查看制作进度'}</Text></View><AppIcon name="chevron-right" size={16}/></Button>)}
 {ready&&!count?<View className="water-empty"><AppIcon name="clapperboard" size={30}/><Text>暂无水货项目或批量作品</Text><Text className="water-note">粘贴字幕创建漫剧，或进入批量工厂导入书单。</Text></View>:ready&&!water.length&&!batch.length?<Text className="water-note">没有匹配作品，请调整名称。</Text>:null}
 </>;
}
