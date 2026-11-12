import { useEffect, useMemo, useRef, useState } from 'react';
import { planMultiCamGroup, reviewMultiCamAlignment } from './multicam.js';
import './multicam-workspace.css';

const number = value => value === '' ? NaN : Number(value);
function CameraPreview({ asset, sourceMs, playing, label, active, onSelect }) {
  const ref = useRef(null);
  const [failure, setFailure] = useState('');
  useEffect(() => {
    const video = ref.current;
    return () => { if (video) { video.pause(); video.removeAttribute('src'); video.load(); } };
  }, []);
  useEffect(() => {
    const video = ref.current;
    if (!video || !asset) return;
    const sync = () => { if (Number.isFinite(sourceMs) && Math.abs(video.currentTime * 1000 - sourceMs) > (playing ? 200 : 20)) video.currentTime = sourceMs / 1000; };
    if (video.readyState >= 1) sync();
    video.addEventListener('loadedmetadata', sync);
    if (playing) video.play().catch(() => setFailure('此机位无法播放，请检查素材')); else video.pause();
    return () => video.removeEventListener('loadedmetadata', sync);
  }, [asset, sourceMs, playing]);
  return <button type="button" className={`aicut-multicam-tile${active ? ' is-active' : ''}`} onClick={onSelect} aria-pressed={active} aria-label={`切换到${label}`}>
    {asset ? <video ref={ref} src={asset.proxyUrl || asset.url} muted playsInline preload="metadata" onError={() => setFailure('素材加载失败')} /> : <span>请选择机位素材</span>}
    <strong>{label}{asset ? ` · ${asset.name || asset.id}` : ''}{active ? ' · 当前画面' : ''}</strong>{failure && <small role="alert">{failure}</small>}
  </button>;
}

// The parent owns project/revision checks and persists this exact reviewed plan.
export default function MultiCamWorkspace({ assets = [], onApply, onCancel, onAutoAlign }) {
  const videos = assets.filter(a => a.kind === 'video');
  const audios = assets.filter(a => a.kind === 'audio');
  const [cameras, setCameras] = useState([{assetId:'',offsetMs:0},{assetId:'',offsetMs:0}]);
  const [cuts, setCuts] = useState([{atMs:0,cameraIndex:0}]);
  const [startMs, setStartMs] = useState(0);
  const [audioId, setAudioId] = useState('');
  const [audioStartMs, setAudioStartMs] = useState(0);
  const [position, setPosition] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [layout, setLayout] = useState(4);
  const [page, setPage] = useState(0);
  const [applyError, setApplyError] = useState('');
  const [applying, setApplying] = useState(false);
  const [aligning, setAligning] = useState(false);
  const [alignmentError, setAlignmentError] = useState('');
  const [alignmentResult, setAlignmentResult] = useState(null);
  const [windowSeconds, setWindowSeconds] = useState(30);
  const [maxLagSeconds, setMaxLagSeconds] = useState(5);
  const alignmentRequest = useRef(null);
  const cameraIdsKey = JSON.stringify(cameras.map(c => c.assetId));
  const cameraIdsRef = useRef(cameraIdsKey);
  cameraIdsRef.current = cameraIdsKey;
  const cancelAlignment = () => {
    const controller = alignmentRequest.current;
    alignmentRequest.current = null;
    if (controller) controller.abort();
    setAligning(false);
  };
  useEffect(() => {
    cancelAlignment(); setAlignmentResult(null); setAlignmentError('');
    return () => { const controller = alignmentRequest.current; alignmentRequest.current = null; if (controller) controller.abort(); };
  }, [cameraIdsKey]);
  const review = useMemo(() => { try { return {plan: planMultiCamGroup(assets, {cameras,cuts,startMs,audio:audioId ? {mode:'asset',assetId:audioId,sourceStartMs:audioStartMs} : {mode:'none'}})}; } catch(error) { return {error:error.message}; } }, [assets,cameras,cuts,startMs,audioId,audioStartMs]);
  const plan = review.plan;
  const duration = plan?.summary.durationMs || 0;
  useEffect(() => { setPlaying(false); setPosition(0); setApplyError(''); }, [cameras,startMs,audioId,audioStartMs]);
  useEffect(() => {
    if (!playing || !duration) return;
    const origin = performance.now() - position; let handle;
    const tick = now => { const next = Math.min(duration - 1, Math.round(now - origin)); setPosition(next); if (next >= duration - 1) setPlaying(false); else handle = requestAnimationFrame(tick); };
    handle = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(handle);
  }, [playing, duration]); // Position is captured only when playback starts.
  const activeCamera = [...cuts].reverse().find(c => c.atMs <= position)?.cameraIndex ?? 0;
  const pages = Math.ceil(cameras.length / layout);
  const visible = cameras.slice(page * layout, (page + 1) * layout);
  const editCamera = (index, patch) => { cancelAlignment(); setCameras(old => old.map((c,i) => i === index ? {...c,...patch} : c)); };
  const autoAlign = async () => {
    if (typeof onAutoAlign !== 'function' || aligning) return;
    const assetIds = cameras.map(c => c.assetId);
    if (assetIds.some(id => !id || !videos.some(a => a.id === id)) || new Set(assetIds).size !== cameras.length) { setAlignmentError('请先为每个机位选择不同的自有视频'); return; }
    const windowMs = windowSeconds * 1000, maxLagMs = maxLagSeconds * 1000;
    if (!Number.isSafeInteger(windowMs) || windowMs < 2000 || windowMs > 60000 || !Number.isSafeInteger(maxLagMs) || maxLagMs < 20 || maxLagMs > 10000) { setAlignmentError('分析时长须为2到60秒，最大偏移须为0.02到10秒'); return; }
    cancelAlignment(); const controller = new AbortController(); alignmentRequest.current = controller;
    const inputKey = cameraIdsKey; setAligning(true); setAlignmentError(''); setAlignmentResult(null); setPlaying(false);
    try {
      const response = await onAutoAlign({assetIds, referenceAssetId:assetIds[0], windowMs, maxLagMs, signal:controller.signal});
      if (controller.signal.aborted || alignmentRequest.current !== controller || cameraIdsRef.current !== inputKey) return;
      const reviewed = reviewMultiCamAlignment(cameras, response);
      setCameras(reviewed.cameras); setAlignmentResult(reviewed);
    } catch (error) {
      if (!controller.signal.aborted && alignmentRequest.current === controller && cameraIdsRef.current === inputKey) setAlignmentError(error.message || '声音同步失败，请手动对齐');
    } finally { if (alignmentRequest.current === controller) { alignmentRequest.current = null; setAligning(false); } }
  };
  const switchCamera = cameraIndex => { if (!plan) return; setCuts(old => [...old.filter(c => c.atMs !== position), {atMs:position,cameraIndex}].sort((a,b) => a.atMs-b.atMs)); };
  const apply = async () => { if (!plan || applying) return; setPlaying(false); setApplying(true); try { await onApply(plan); } catch(error) { setApplyError(error.message || '应用失败，请重试'); } finally { setApplying(false); } };
  return <section className="aicut-multicam" aria-label="多机位剪辑">
    <header><div><h2>{onAutoAlign ? '多机位剪辑' : '多机位剪辑 · 手动对齐'}</h2><p>明确选择2到9段自有视频。偏移表示共同时间0对应的源时间，可填负数；仅使用所有机位共有的有效时段。</p><p>{onAutoAlign ? '可根据前段录音自动对齐，也可手动调整。' : '当前支持手动切换；自动声音同步尚未接入。'}所有机位原声静音，主音频请单独选择；需要某个机位原声时，请先提取该机位音频。</p></div>{onCancel && <button type="button" onClick={() => {cancelAlignment();onCancel();}}>取消</button>}</header>
    <div className="aicut-multicam-inputs">{cameras.map((camera,index) => <div key={index} className="aicut-multicam-slot"><label>机位 {index+1}<select value={camera.assetId} onChange={e => editCamera(index,{assetId:e.target.value})}><option value="">请选择自有视频</option>{videos.map(a => <option key={a.id} value={a.id} disabled={cameras.some((c,i) => i !== index && c.assetId === a.id)}>{a.name || a.id}</option>)}</select></label><label>源时间偏移（毫秒）<input type="number" step="1" value={Number.isFinite(camera.offsetMs) ? camera.offsetMs : ''} onChange={e => editCamera(index,{offsetMs:number(e.target.value)})} /></label>{cameras.length > 2 && <button type="button" onClick={() => {setCameras(old => old.filter((_,i) => i !== index));setCuts([{atMs:0,cameraIndex:0}]);setPage(0);}}>移除机位</button>}</div>)}</div>
    <button type="button" disabled={cameras.length >= 9} onClick={() => setCameras(old => [...old,{assetId:'',offsetMs:0}])}>添加机位</button>
    {onAutoAlign && <div className="aicut-multicam-inputs"><label>分析时长（秒）<input type="number" min="2" max="60" step="1" value={Number.isFinite(windowSeconds) ? windowSeconds : ''} onChange={e => {cancelAlignment();setWindowSeconds(number(e.target.value));}} /></label><label>最大偏移（秒）<input type="number" min="0.02" max="10" step="0.02" value={Number.isFinite(maxLagSeconds) ? maxLagSeconds : ''} onChange={e => {cancelAlignment();setMaxLagSeconds(number(e.target.value));}} /></label><button type="button" disabled={aligning || applying} onClick={autoAlign}>{aligning ? '正在分析真实录音…' : '自动声音对齐'}</button>{aligning && <button type="button" onClick={cancelAlignment}>取消分析</button>}<small>参考：机位1。仅分析指定长度的前段声音；没有共同录音或匹配不足时，请手动对齐。</small></div>}
    {alignmentError && <p role="alert">{alignmentError}</p>}
    {alignmentResult && <div role="status"><p>已使用声音能量互相关分析回填对齐结果，时间精度 {alignmentResult.timeResolutionMs} 毫秒。可以继续手动调整偏移，确认后再应用。</p>{alignmentResult.results.map((result,index) => <p key={result.assetId}>机位 {index+1}：偏移 {result.offsetMs} 毫秒，匹配度 {(result.confidence*100).toFixed(1)}%</p>)}</div>}
    <div className="aicut-multicam-inputs"><label>插入时间（毫秒）<input type="number" min="0" step="1" value={Number.isFinite(startMs) ? startMs : ''} onChange={e => setStartMs(number(e.target.value))} /></label><label>主音频<select value={audioId} onChange={e => setAudioId(e.target.value)}><option value="">静音，不添加音频</option>{audios.map(a => <option key={a.id} value={a.id}>{a.name || a.id}</option>)}</select></label>{audioId && <label>音频源起点（毫秒）<input type="number" min="0" step="1" value={Number.isFinite(audioStartMs) ? audioStartMs : ''} onChange={e => setAudioStartMs(number(e.target.value))} /></label>}</div>
    {review.error && <p role="alert">{review.error}</p>}
    <div className="aicut-multicam-toolbar"><button type="button" aria-pressed={layout === 4} onClick={() => {setLayout(4);setPage(0);}}>4 格</button><button type="button" aria-pressed={layout === 9} onClick={() => {setLayout(9);setPage(0);}}>9 格</button>{pages > 1 && <label>预览页<select value={page} onChange={e => setPage(Number(e.target.value))}>{Array.from({length:pages},(_,i) => <option key={i} value={i}>第 {i+1} 页</option>)}</select></label>}<span>当前输出：机位 {activeCamera+1}</span></div>
    <div className={`aicut-multicam-grid columns-${layout === 9 ? 3 : 2}`}>{visible.map((camera,i) => {const index=page*layout+i; const asset=videos.find(a => a.id === camera.assetId); return <CameraPreview key={`${index}-${camera.assetId}`} asset={asset} label={`机位 ${index+1}`} sourceMs={plan ? plan.summary.commonStartMs+position+camera.offsetMs : NaN} playing={playing && !!plan} active={activeCamera === index} onSelect={() => switchCamera(index)} />;})}</div>
    <div className="aicut-multicam-toolbar"><button type="button" disabled={!plan} onClick={() => setPlaying(v => !v)}>{playing ? '暂停预览' : '播放预览'}</button><input aria-label="多机位预览位置" type="range" min="0" max={Math.max(0,duration-1)} step="1" disabled={!plan} value={Math.min(position,Math.max(0,duration-1))} onChange={e => {setPlaying(false);setPosition(Number(e.target.value));}} /><span>{position} / {duration} 毫秒</span></div>
    <p>点击机位画面，在当前时间添加切换点。预览为静音，主音频在应用后的时间线上播放。</p>
    <ol className="aicut-multicam-cuts" aria-label="切换点">{cuts.map((cut,index) => <li key={cut.atMs}><span>{cut.atMs} 毫秒 → 机位 {cut.cameraIndex+1}</span><button type="button" onClick={() => {setPlaying(false);setPosition(cut.atMs);}}>定位</button>{index > 0 && <button type="button" onClick={() => setCuts(old => old.filter((_,i) => i !== index))}>删除切换点</button>}</li>)}</ol>
    {plan && <p role="status">共同可用时长 {duration} 毫秒；将新增 {cuts.length} 段可单独编辑的画面{audioId ? '和一条连续主音频' : '，无音频'}。插入位置 {startMs} 毫秒。</p>}
    {applyError && <p role="alert">{applyError}</p>}<footer><button type="button" disabled={!plan || applying || aligning || typeof onApply !== 'function'} onClick={apply}>{applying ? '正在应用…' : '应用已检查的多机位剪辑'}</button></footer>
  </section>;
}
