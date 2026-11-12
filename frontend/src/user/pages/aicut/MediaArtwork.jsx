import React, { useEffect, useRef, useState } from 'react';
import { getAICutAudioAnalysis } from '../../../shared/api/aicut.js';
import { apiRequest } from '../../../shared/api/client.js';
import { assetThumbnailUrl } from './resourceLibraryModel.js';
import { artworkFrameTimes, artworkPeaks, audioArtworkQueue, decodeArtworkFrame, videoArtworkQueue } from './mediaArtwork.js';
import './media-artwork.css';

function Fallback({ kind, pending }) {
  return <span className="aicut-media-artwork-fallback">
    <svg viewBox="0 0 24 24" aria-hidden="true">{kind === 'audio'
      ? <path d="M9 17V5l11-2v12M9 8l11-2M9 17c0 2-5 4-6 1s5-5 6-1Zm11-2c0 2-5 4-6 1s5-5 6-1Z" />
      : <><rect x="3" y="4" width="18" height="16" rx="2" /><path d="m10 8 6 4-6 4ZM3 8h3M3 16h3M18 8h3M18 16h3" /></>}</svg>
    <span>{pending ? '正在读取' : kind === 'audio' ? '波形暂不可用' : '预览暂不可用'}</span>
  </span>;
}

export default function MediaArtwork({ asset = {}, projectId, clip, strip = false }) {
  const root = useRef(null);
  const [visible, setVisible] = useState(false);
  const [result, setResult] = useState({ key: '', status: 'pending', frames: [], analysis: null });
  const kind = asset.kind || clip?.kind || 'video';
  const sourceUrl = asset.proxyUrl || asset.url || '';
  // An asset poster is a reliable library cover, but cannot represent an arbitrary trimmed clip.
  const directUrl = kind === 'image' ? assetThumbnailUrl(asset) : kind === 'video' && !clip && !strip ? assetThumbnailUrl(asset) : '';
  const times = artworkFrameTimes(asset, clip, strip);
  const timesKey = times.join(',');
  const key = `${kind}:${projectId || ''}:${asset.id || ''}:${sourceUrl}:${directUrl}:${timesKey}`;

  useEffect(() => {
    const element = root.current;
    if (!element) return undefined;
    if (typeof IntersectionObserver === 'undefined') { setVisible(true); return undefined; }
    const observer = new IntersectionObserver(entries => setVisible(entries.some(entry => entry.isIntersecting)), { threshold: .01 });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!visible) return undefined;
    let live = true;
    const handles = [];
    setResult({ key, status: 'pending', frames: [], analysis: null });
    if (directUrl) return () => { live = false; };
    if (kind === 'audio' && projectId && asset.id) {
      const handle = audioArtworkQueue.acquire(`${projectId}:${asset.id}:${sourceUrl}:240`, signal =>
        getAICutAudioAnalysis(projectId, asset.id, 240, (path, options) => apiRequest(path, { ...options, signal, timeoutMs: 20000, silent: true })));
      handles.push(handle);
      handle.promise.then(analysis => {
        if (live) setResult({ key, status: analysis?.peaks?.length ? 'ready' : 'failed', frames: [], analysis });
      }, () => { if (live) setResult({ key, status: 'failed', frames: [], analysis: null }); });
    } else if (kind === 'video' && sourceUrl) {
      const requests = times.map(time => {
        const handle = videoArtworkQueue.acquire(`${sourceUrl}:${time}`, signal => decodeArtworkFrame(sourceUrl, time, signal));
        handles.push(handle);
        return handle.promise.catch(() => '');
      });
      Promise.all(requests).then(frames => {
        if (live) setResult({ key, status: frames.some(Boolean) ? 'ready' : 'failed', frames, analysis: null });
      });
    } else setResult({ key, status: 'failed', frames: [], analysis: null });
    return () => { live = false; handles.forEach(handle => handle.release()); };
  // The sampling key includes every relevant source-time input without depending on clip object identity.
  }, [visible, key]);

  const current = result.key === key ? result : { status: 'pending', frames: [] };
  const peaks = kind === 'audio' && current.analysis ? artworkPeaks(current.analysis, clip, strip ? 160 : 80) : [];
  const frames = directUrl ? Array(strip ? 3 : 1).fill(directUrl) : current.frames;
  const showImages = visible && frames?.some(Boolean) && current.status !== 'failed';
  return <span ref={root} className={`aicut-media-artwork${strip ? ' is-strip' : ''} is-${kind}`} aria-hidden="true">
    {peaks.length ? <svg className="aicut-media-waveform" viewBox={`0 0 ${peaks.length * 2} 64`} preserveAspectRatio="none">
      <path d={peaks.map((peak, index) => `M${index * 2 + 1} ${32 - peak * 28}V${32 + peak * 28}`).join(' ')} />
      <path className="aicut-media-waveform-baseline" d={`M0 32H${peaks.length * 2}`} />
    </svg> : showImages ? <span className="aicut-media-frames">{frames.map((frame, index) => frame
      ? <img key={`${frame}:${index}`} src={frame} alt="" draggable={false} loading="lazy" decoding="async" onLoad={() => {
        if (directUrl) setResult(previous => previous.key === key ? { ...previous, status: 'ready' } : previous);
      }} onError={() => { if (directUrl) setResult({ key, status: 'failed', frames: [], analysis: null }); }} />
      : <Fallback key={index} kind={kind} pending={false} />)}</span>
      : <Fallback kind={kind} pending={current.status === 'pending'} />}
  </span>;
}
