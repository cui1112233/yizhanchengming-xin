import { Button, Image, ScrollView, Text, Textarea, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import { comicPageReadiness } from './comicJourney';
import AppIcon from '../../../components/AppIcon';
import type { ComicPageItem } from '../../services/comicResult';
import { replaceComicImage, saveComicPagePrompts } from '../../services/comicEditing';

type Props = {
  onGenerateImages?: () => void; onGeneratePrompts?: () => void; imageGenerationBlockedReason?: string;
  batchId: string; bookId: string; pages: ComicPageItem[]; pageIndex: number; paths: Record<string,string>;
  generating?: boolean; busy: string; error: string; onSelect: (index:number) => void; onPreview: (id:string) => void;
  onRegenerate: () => void; onPrimary: (id:string) => void; onSave: () => void; onExport: () => void; onReload: () => Promise<unknown>;
};
export default function ComicPageWorkspace(p:Props) {
  const [tool,setTool] = useState<''|'prompts'|'versions'>('');
  const [draft,setDraft] = useState({comicPrompt:'',imagePrompt:''});
  const [pending,setPending] = useState(false);
  const [error,setError] = useState('');
  const lock = useRef(false); const alive = useRef(true);
  useEffect(() => { alive.current = true; return () => { alive.current = false; }; },[]);
  const page = p.pages[p.pageIndex];
  const selected = page?.images.find(i=>i.id===page.primaryImageId);
  const blocked = Boolean(p.busy) || pending;
  const navigationBlocked = pending || Boolean(p.busy && p.busy !== "images");
  const readiness = comicPageReadiness(p.pages);
  const perform = async (action:()=>Promise<unknown>) => {
    if(lock.current || blocked) return;
    lock.current=true;setPending(true);setError('');
    try { await action(); if(alive.current) { await p.onReload(); setTool(''); } }
    catch(cause:any) { if(alive.current && !/cancel/i.test(cause?.errMsg || '')) setError(cause?.message || cause?.errMsg || '操作失败，请重试'); }
    finally { lock.current=false;if(alive.current)setPending(false); }
  };
  if(!page) return <View className="ce-empty"><Text>{p.generating ? '正在准备漫画页图' : '还没有漫画页图'}</Text><Text>{p.generating ? '页图生成后会显示在这里，也可以稍后返回查看。' : '完成资产检查后，确认配置即可开始分页出图。'}</Text></View>;
  return <View className="ce-pages">
    <View className="ce-result-summary"><View className="ce-result-summary-head"><Text>已读取 {readiness.total} 页</Text><Button disabled={blocked || (p.onGenerateImages ? !p.pages.length : !readiness.ready)} onClick={p.onExport}>导出漫画</Button></View><Text className="ce-muted">有主图 {readiness.ready} 页 · 无图失败 {readiness.failed} 页 · 待出图 {readiness.waiting} 页</Text><Text className="ce-muted">{p.generating ? "仍在生成，可以先查看已有图片；导出只包含当前已有主图。" : "先检查页图，再保存当前页或导出已有漫画。"}</Text>{readiness.firstFailedIndex >= 0 ? <Button className="ce-failed-page-link" disabled={blocked} onClick={() => p.onSelect(readiness.firstFailedIndex)}>查看失败页</Button> : null}</View>
    {p.onGeneratePrompts ? <View className="ce-prompt-review-head"><Text>先检查本页提示词，满意后再生成图片</Text><Button disabled={blocked} onClick={p.onGeneratePrompts}>{p.pages.every(page => page.imagePrompt?.trim()) ? "重新生成提示词" : "补齐提示词"}</Button></View> : null}
    <View className="ce-page-preview" onClick={()=>{if(selected) p.onPreview(selected.id);}}>
      {selected && p.paths[selected.id] ? <Image className="ce-page-main-image" mode="aspectFit" src={p.paths[selected.id]} onClick={event=>{event.stopPropagation();p.onPreview(selected.id);}} /> : p.onGenerateImages && (page.comicPrompt || page.imagePrompt) ? <ScrollView scrollY className="ce-page-prompt-preview"><Text className="ce-prompt-preview-title">第 {page.ordinal} 页 · 提示词预览</Text><Text>漫画提示词</Text><Text>{page.comicPrompt || "尚未生成"}</Text><Text>图片描述</Text><Text>{page.imagePrompt || "尚未生成"}</Text></ScrollView> : <View className="ce-empty"><Text>{/fail|error/i.test(page.status) ? '本页生成失败，可重新生成' : page.images.length ? '图片暂未读取，点击图片版本查看' : '这一页尚未生成图片'}</Text></View>}
    </View>
    <View className="ce-page-count"><Button disabled={p.pageIndex===0 || navigationBlocked} onClick={()=>p.onSelect(p.pageIndex-1)}>‹</Button><Text>第 {page.ordinal} 页 / 共 {p.pages.length} 页</Text><Button disabled={p.pageIndex===p.pages.length-1 || navigationBlocked} onClick={()=>p.onSelect(p.pageIndex+1)}>›</Button></View>
    <ScrollView scrollX className="ce-page-strip"><View>{p.pages.map((item,index)=>{const image=item.images.find(i=>i.id===item.primaryImageId);return <Button key={item.pageId || item.videoId} className={index===p.pageIndex?'is-selected':''} disabled={navigationBlocked} onClick={()=>p.onSelect(index)}>{image && p.paths[image.id] ? <Image src={p.paths[image.id]} mode="aspectFill"/> : <View className="ce-page-placeholder"><Text>{/fail|error/i.test(item.status)?'失败':item.images.length?'未读取':'待出图'}</Text></View>}<Text>{item.ordinal}</Text></Button>;})}</View></ScrollView>
    {error || p.error ? <Text className="mp-error">{error || p.error}</Text> : null}
    <View className="ce-tools ce-page-tools"><Text className="ce-page-scope-note">仅操作第 {page.ordinal} 页 · 重生成新增候选图，原图保留</Text>
      <Button disabled={blocked} onClick={()=>{setDraft({comicPrompt:page.comicPrompt || '',imagePrompt:page.imagePrompt || ''});setTool('prompts');}}><AppIcon name="file-pen-line" size={23}/><Text>编辑提示词</Text></Button>
      <Button disabled={blocked} onClick={p.onRegenerate}><AppIcon name="refresh-ccw" size={23}/><Text>{p.onGenerateImages && !page.images.length ? "生成本页" : "重新生成"}</Text></Button>
      <Button disabled={blocked} onClick={()=>setTool('versions')}><AppIcon name="notebook-tabs" size={23}/><Text>图片版本</Text></Button>
      <Button disabled={blocked} onClick={()=>perform(()=>replaceComicImage(p.batchId,p.bookId,{videoId:page.videoId}))}><AppIcon name="camera" size={23}/><Text>替换图片</Text></Button>
    </View>
    <View className="comic-dock"><View className="comic-dock-copy"><Text className="comic-dock-line">{p.onGenerateImages && readiness.ready < readiness.total ? "检查提示词后，单独开始出图" : `第 ${page.ordinal} 页 · ${selected ? '已有主图' : page.status}`}</Text>{p.onGenerateImages && p.imageGenerationBlockedReason ? <Text className="comic-dock-caption">{p.imageGenerationBlockedReason}</Text> : null}</View>{p.onGenerateImages && readiness.ready < readiness.total ? <Button className="mp-primary" disabled={blocked || Boolean(p.imageGenerationBlockedReason) || !p.pages.some(item=>!item.images.length&&item.imagePrompt?.trim())} loading={p.busy==='images'} onClick={p.onGenerateImages}>生成漫画图片</Button> : <Button className="mp-primary" disabled={blocked || !selected} loading={p.busy==='save'} onClick={p.onSave}>保存到相册</Button>}</View>
    {tool ? <><View className="comic-sheet-mask" onClick={()=>{if(!blocked)setTool('');}}/><View className="comic-sheet ce-choice-sheet"><View className="ce-sheet-title"><Text>第 {page.ordinal} 页 · {tool==='prompts'?'编辑提示词':'图片版本'}</Text><Button disabled={blocked} onClick={()=>setTool('')}>关闭</Button></View><ScrollView scrollY className="ce-sheet-scroll">
      {tool==='prompts'?<View className="ce-edit-fields"><Text>漫画提示词</Text><Textarea maxlength={-1} disabled={blocked} value={draft.comicPrompt} onInput={e=>setDraft({...draft,comicPrompt:e.detail.value})}/><Text>图片提示词</Text><Textarea maxlength={-1} disabled={blocked} value={draft.imagePrompt} onInput={e=>setDraft({...draft,imagePrompt:e.detail.value})}/><Button className="mp-primary" disabled={blocked || !draft.comicPrompt.trim() || !draft.imagePrompt.trim()} loading={pending} onClick={()=>perform(()=>saveComicPagePrompts(p.batchId,p.bookId,page.videoId,draft.comicPrompt,draft.imagePrompt))}>保存提示词</Button></View>:<View className="ce-version-list">{page.images.map(image=><View key={image.id}>{p.paths[image.id]?<Image mode="aspectFit" src={p.paths[image.id]} onClick={()=>p.onPreview(image.id)}/>:<Text>图片暂不可用，请重新读取</Text>}<Button disabled={blocked || image.isPrimary} onClick={()=>p.onPrimary(image.id)}>{image.isPrimary?'当前主图':'设为主图'}</Button></View>)}{!page.images.length?<Text>还没有图片版本</Text>:null}</View>}
      {error || p.error ? <Text className="mp-error">{error || p.error}</Text>:null}
    </ScrollView></View></>:null}
  </View>;
}
