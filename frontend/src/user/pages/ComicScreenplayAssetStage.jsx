import { Alert, Button, Checkbox, Empty, Input, Select } from 'antd';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import { useEffect, useRef, useState } from 'react';
import { listComicAssetLibrary, stageComicAssetLibraryEntry } from '../../shared/api/comicScreenplayAssetLibrary.js';
import { actComicAssetMedia } from '../../shared/api/comicScreenplayAssetActions.js';
import ComicScreenplayAssetLibraryPicker from './ComicScreenplayAssetLibraryPicker.jsx';
import { getComicMedia } from '../../shared/api/comicScreenplayWorkflow.js';
import { comicAssetMediaImage } from './comic-screenplay-stages.js';
import AssetImage from './ComicScreenplayImage.jsx';
import ComicScreenplayPositioningStage from './ComicScreenplayPositioningStage.jsx';
import ComicScreenplayAssetDialog from './ComicScreenplayAssetDialog.jsx';
import ComicScreenplayAssetBatchDialog from './ComicScreenplayAssetBatchDialog.jsx';
import ComicScreenplayVoiceDesignStage from './ComicScreenplayVoiceDesignStage.jsx';

// Shared by the displayed source label and the request, so enabling analysis
// always supplies the same real saved source rather than a fabricated script.
export function comicAssetAnalysisSource(workspace) {
  const content = String(workspace?.document?.content || '').trim();
  if (content) return {content, label:'已保存剧本'};
  const chapters = workspace?.authoringChapters || [];
  if (chapters.length) {
    const inputs = chapters.map(chapter => String(chapter.script_text || '').trim() || String(chapter.source_text || '').trim());
    const empty = inputs.findIndex(text => !text);
    if (empty >= 0) return {content:'', label:'章节原文', error:`第${empty+1}章没有剧本或原文，请先补全并保存`};
    const scriptCount = chapters.filter(chapter => String(chapter.script_text || '').trim()).length;
    const label = scriptCount === chapters.length ? '章节剧本' : scriptCount ? '章节剧本与原文' : '章节原文';
    return {content:inputs.join('\n\n'), label};
  }
  return {content:String(workspace?.project?.source_text || '').trim(), label:'已保存原文'};
}

const assetTabs = [{ key: 'character', label: '人物', icon: 'PeopleOutline' }, { key: 'scene', label: '场景', icon: 'LocationOutline' }, { key: 'prop', label: '道具', icon: 'CubeOutline' }, { key: 'positioning', label: '站位', icon: 'PeopleCircleOutline' }, { key: 'voice', label: 'AI音色', icon: 'MusicalNotesOutline' }];
export default function ComicScreenplayAssetStage({ projectId, workspace, readOnly, canGenerate, working, degree, onDegree, onSaveDegree, degreeSaving, onAdd, onEdit, onDelete, onExtract, onReparse, onMedia, onChanged }) {
  const [type, setType] = useState('character'), [query, setQuery] = useState(''), [compact, setCompact] = useState(false), [selected, setSelected] = useState([]), [visited,setVisited] = useState({character:true});
  const [positioningToolbar,setPositioningToolbar]=useState(null), [voiceToolbar,setVoiceToolbar]=useState(null), [assetDialog,setAssetDialog]=useState(null), [batchDialog,setBatchDialog]=useState(false);
  const [media, setMedia] = useState(null), [mediaError, setMediaError] = useState('');
  const [replaceAsset,setReplaceAsset]=useState(null), [replacePending,setReplacePending]=useState(null), [replaceError,setReplaceError]=useState(''), [replacing,setReplacing]=useState(false);
  const replaceLock=useRef(false), replaceKeys=useRef(new Map()), alive=useRef(true);
  const replaceContext=useRef({projectId,epoch:0});
  // Invalidate during the project render, before any old promise can resume.
  if(replaceContext.current.projectId!==projectId)replaceContext.current={projectId,epoch:replaceContext.current.epoch+1};
  useEffect(()=>{alive.current=true;return()=>{alive.current=false;};},[]);
  useEffect(()=>{
    replaceLock.current=false;replaceKeys.current.clear();setReplaceAsset(null);setReplacePending(null);setReplaceError('');setReplacing(false);setMedia(null);setAssetDialog(null);setBatchDialog(false);setSelected([]);
  },[projectId]);
  async function replaceImage(asset,entry) {
    if(readOnly||replaceLock.current)return;
    const context=replaceContext.current;
    const current=()=>alive.current&&replaceContext.current===context;
    replaceLock.current=true;setReplacing(true);setReplaceError('');setReplacePending({asset,entry});
    const fingerprint=JSON.stringify([projectId,asset.id,entry.id,entry.revision]);
    if(!replaceKeys.current.has(fingerprint))replaceKeys.current.set(fingerprint,crypto.randomUUID());
    try {
      const item=await stageComicAssetLibraryEntry(projectId,entry.id,{requestKey:replaceKeys.current.get(fingerprint),targetId:asset.id,expectedEntryRevision:entry.revision});
      if(!current())return;
      await actComicAssetMedia(projectId,{action:'select',targetId:asset.id,itemId:item.id});
      if(!current())return;
      const data=await getComicMedia(projectId);if(current()){setMedia(data);setReplacePending(null);void onChanged?.();}
    }catch(e){if(current())setReplaceError(e.message||'更换图片失败');}
    finally{if(current()){replaceLock.current=false;setReplacing(false);}}
  }
  useEffect(() => { let alive = true; const refresh = () => getComicMedia(projectId).then(data => { if (alive) { setMedia(data); setMediaError(''); } }).catch(error => { if (alive) setMediaError(error.message || '资产图片读取失败'); }); void refresh(); const timer = setInterval(() => { if (!document.hidden) void refresh(); }, 10000); return () => { alive = false; clearInterval(timer); }; }, [projectId, workspace?.project?.revision]);
  const [localAssets,setLocalAssets] = useState(workspace?.assets || []);
  useEffect(()=>{setLocalAssets(workspace?.assets||[]);},[workspace?.assets]);
  const assets = localAssets, typed = assets.filter(item => item.type === type), filtered = typed.filter(item => `${item.name} ${item.description || ''}`.toLowerCase().includes(query.toLowerCase()));
  const analysisSource = comicAssetAnalysisSource(workspace);
  const hasSource = !!analysisSource.content;
  function imageFor(asset) { return comicAssetMediaImage(media,asset.id)?.url || ''; }
  return <div className="comic-assets-stage">{type!=='voice'?<Alert type="info" showIcon icon={<OriginalIcon name="InformationCircleOutline" size={16}/>} closable message={type==='positioning'?'站位用来锁死左右、前后、座位关系。车内要写清前排左侧-驾驶位、前排右侧-副驾、后排左/中/右（默认中国左舵）。出图复用模型配置，并带上当前这一集完整剧本。':'资产图片在视频生成过程中可以继续生成与更新，无需在当前阶段全部完成，可随时回到此处补充或重新生成。'} />:null}<div className="comic-assets-heading-row"><div className="comic-asset-tabs" role="tablist" aria-label="资产类型">{assetTabs.map(tab => <button type="button" role="tab" aria-selected={type === tab.key} className={type === tab.key ? 'is-active' : ''} key={tab.key} onClick={() => { setType(tab.key); setVisited(current=>({...current,[tab.key]:true})); setQuery(''); setSelected([]); }}><OriginalIcon name={tab.icon} size={16}/><span>{tab.label}</span>{!['voice','positioning'].includes(tab.key) ? <small>{assets.filter(item => item.type === tab.key).length}</small> : null}</button>)}</div><div hidden={type!=='positioning'} ref={setPositioningToolbar}/><div hidden={type!=='voice'} ref={setVoiceToolbar}/></div>
  <div hidden={type!=='positioning'}>{visited.positioning?<ComicScreenplayPositioningStage key={projectId} projectId={projectId} workspace={workspace} media={media} toolbarHost={positioningToolbar} readOnly={readOnly} canGenerate={canGenerate} />:null}</div>
  <div hidden={type!=='voice'}>{visited.voice?<ComicScreenplayVoiceDesignStage key={projectId} projectId={projectId} workspace={workspace} toolbarHost={voiceToolbar} readOnly={readOnly} canGenerate={canGenerate} />:null}</div>
  {!['positioning','voice'].includes(type)?<>
    {replaceError?<Alert type="error" showIcon message={replaceError} action={<Button disabled={replacing||readOnly} onClick={()=>replacePending&&replaceImage(replacePending.asset,replacePending.entry)}>重试更换图片</Button>}/>:null}
    <div className="comic-asset-toolbar"><b>共 {typed.length} 个{assetTabs.find(tab => tab.key === type)?.label}</b><Input style={{ width: 180 }} prefix={<OriginalIcon name="SearchOutline" size={14} style={{opacity:.5}}/>} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索资产" allowClear /><Button aria-label="切换紧凑卡片" icon={<OriginalIcon name={compact ? "GridOutline" : "AppsOutline"} size={16}/>} onClick={() => setCompact(value => !value)}>{compact ? '标准卡片' : '紧凑卡片'}</Button><Button disabled={readOnly || !!working} icon={<OriginalIcon name="AddOutline" size={18}/>} onClick={() => setAssetDialog({type})}>新增{type==='character'?'角色':assetTabs.find(tab => tab.key === type)?.label}</Button><Select aria-label="资产解析程度" value={degree} onChange={onDegree} disabled={readOnly || !!working} options={[{value:'master',label:'大师 · 3D'},{value:'master_real',label:'大师 · 写实'},{value:'more_diff',label:'差异化'},{value:'standard',label:'标准'}]} />{onSaveDegree&&degree!==workspace?.project?.asset_degree?<Button size="small" disabled={readOnly||!!working} loading={degreeSaving} onClick={onSaveDegree}>保存解析程度</Button>:null}<Button icon={<OriginalIcon name="SparklesOutline" size={18}/>} disabled={!canGenerate || !!working || !hasSource} title={!hasSource ? (analysisSource.error || '请先保存原文或剧本') : undefined} loading={working === type} onClick={() => onExtract(type)}>资产分析</Button>{onReparse?<Button icon={<OriginalIcon name="RefreshOutline" size={18}/>} disabled={readOnly || !canGenerate || !!working || !hasSource} title="使用作品已保存的文字模型重新解析；不使用解析程度或出图画风，保留已有图片" onClick={()=>onReparse(type)}>重新解析</Button>:null}<Button icon={<OriginalIcon name="ListOutline" size={18}/>} disabled={readOnly || !canGenerate || !typed.length} onClick={() => setBatchDialog(true)}>批量制作图片 {selected.length || ''}</Button></div>
    <p className="comic-asset-analysis-source">分析来源：{analysisSource.label}{analysisSource.error ? ` · ${analysisSource.error}` : ''}</p>{!hasSource ? <Alert type="info" showIcon message="可以先手动添加资产；资产分析前请先保存原文或剧本。" /> : null}{mediaError ? <Alert type="warning" showIcon message={mediaError} /> : null}
    <div className={`comic-picture-assets${compact ? ' is-compact' : ''}`}>{filtered.map(asset => <article key={asset.id} className="comic-picture-asset" role="button" tabIndex={readOnly?-1:0} aria-label={`编辑资产${asset.name}`} onClick={()=>{if(!readOnly&&!working)setAssetDialog({type:asset.type,asset});}} onKeyDown={event=>{if(event.key==='Enter'&&event.target===event.currentTarget&&!readOnly&&!working)setAssetDialog({type:asset.type,asset});}}><div className="comic-asset-thumb"><AssetImage placeholderIcon={asset.type==='character'?'PeopleOutline':asset.type==='scene'?'LocationOutline':'CubeOutline'} placeholderSize={22} src={imageFor(asset)} name={asset.name} /><Checkbox onClick={event=>event.stopPropagation()} aria-label={`选择${asset.name}`} checked={selected.includes(asset.id)} onChange={event => setSelected(rows => event.target.checked ? [...new Set([...rows, asset.id])] : rows.filter(id => id !== asset.id))} /><div className="comic-asset-original-actions" onClick={event=>event.stopPropagation()}><Button size="small" danger aria-label={`删除资产${asset.name}`} title="删除" disabled={readOnly||!!working} icon={<OriginalIcon name="TrashOutline" size={12}/>} onClick={()=>onDelete(asset.id)}/><Button size="small" aria-label={`更换${asset.name}图片`} title="更换图片" disabled={readOnly||!!working||replacing} icon={<OriginalIcon name="FolderOpenOutline" size={12}/>} onClick={()=>setReplaceAsset(asset)}/><Button size="small" aria-label={`制作${asset.name}图片`} title="AI生图" disabled={readOnly||!canGenerate} icon={<OriginalIcon name="SparklesOutline" size={12}/>} onClick={()=>setAssetDialog({type:asset.type,asset,imageOnly:true})}/></div></div><div className="comic-picture-asset-copy"><b>{asset.name}</b><p><span className="comic-asset-setting-label">设定</span>{asset.description || asset.identity || '暂无设定'}</p></div></article>)}</div>{!filtered.length ? <Empty description={query ? '没有匹配的资产' : '还没有此类资产，可以手动新增，或明确执行资产分析'} /> : null}
  </>:null}
  {replaceAsset?<ComicScreenplayAssetLibraryPicker type={replaceAsset.type} variant="replace" loadEntries={listComicAssetLibrary} onSelect={entry=>void replaceImage(replaceAsset,entry)} onClose={()=>setReplaceAsset(null)}/>:null}
  {batchDialog?<ComicScreenplayAssetBatchDialog projectId={projectId} workspace={workspace} assets={typed} initialSelected={selected} media={media} canGenerate={canGenerate&&!readOnly} onClose={()=>setBatchDialog(false)} onMediaChanged={setMedia}/>:null}
  {assetDialog?<ComicScreenplayAssetDialog key={`${projectId}:${assetDialog.asset?.id||'new'}:${assetDialog.imageOnly?'image':'edit'}`} projectId={projectId} workspace={workspace} asset={assetDialog.asset} type={assetDialog.type} imageOnly={assetDialog.imageOnly} degree={degree} media={media} readOnly={readOnly} canGenerate={canGenerate} onMediaChanged={setMedia} onClose={()=>setAssetDialog(null)} onSaved={saved=>{setLocalAssets(rows=>rows.some(item=>item.id===saved.id)?rows.map(item=>item.id===saved.id?saved:item):[...rows,saved]);void onChanged?.();}}/>:null}
  </div>;
}
