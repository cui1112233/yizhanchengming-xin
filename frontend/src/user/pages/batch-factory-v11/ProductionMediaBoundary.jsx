import { Children, cloneElement, isValidElement, useEffect, useMemo, useState } from 'react';
import { LoadingOutlined } from '@ant-design/icons';
import {
  getMergeMediaPlaybackURL,
  getProductionMediaBlob,
  isMergeMediaURL,
  isProtectedProductionMediaURL
} from '../../../shared/api/batchFactoryV11.js';

// 合成成片走短期签名直链后由浏览器原生流式播放；等待首帧超过该时长才提示
// 超时，避免 0:00 静默黑屏。
export const MEDIA_FIRST_FRAME_TIMEOUT_MS = 15000;

function firstVideoNode(node) {
  if (Array.isArray(node)) {
    for (const child of node) {
      const found = firstVideoNode(child);
      if (found) return found;
    }
    return null;
  }
  if (!isValidElement(node)) return null;
  if (node.type === 'video') return node;
  return firstVideoNode(node.props?.children);
}

function firstVideoSource(node) {
  const video = firstVideoNode(node);
  return typeof video?.props?.src === 'string' ? video.props.src : '';
}

function replaceVideoSource(node, source, replacement, onPlaybackError, onVideoReady) {
  if (Array.isArray(node)) return node.map(child => replaceVideoSource(child, source, replacement, onPlaybackError, onVideoReady));
  if (!isValidElement(node)) return node;
  if (node.type === 'video' && node.props?.src === source) {
    const originalOnError = node.props.onError;
    const originalOnLoadedData = node.props.onLoadedData;
    return cloneElement(node, {
      src: replacement || undefined,
      onError: event => {
        originalOnError?.(event);
        onPlaybackError?.();
      },
      onLoadedData: event => {
        originalOnLoadedData?.(event);
        onVideoReady?.();
      }
    });
  }
  if (node.props?.children === undefined) return node;
  return cloneElement(node, undefined, Children.map(
    node.props.children,
    child => replaceVideoSource(child, source, replacement, onPlaybackError, onVideoReady)
  ));
}

function downloadName(source) {
  const tail = String(source || '').split('?')[0].split('/').filter(Boolean).pop() || 'video';
  return /\.mp4$/i.test(tail) ? tail : `${tail}.mp4`;
}

export function ProductionMediaBoundary({ children, showDownload = true }) {
  const videoNode = useMemo(() => firstVideoNode(children), [children]);
  const source = useMemo(() => firstVideoSource(children), [children]);
  const mergeMedia = isMergeMediaURL(source);
  const protectedArtifact = isProtectedProductionMediaURL(source);
  const managedSource = Boolean(source) && (mergeMedia || protectedArtifact);
  const autoPlay = Boolean(videoNode?.props?.autoPlay);
  const [resolvedSource, setResolvedSource] = useState(managedSource ? '' : source);
  const [error, setError] = useState('');
  const [playbackError, setPlaybackError] = useState('');
  const [loading, setLoading] = useState(managedSource);
  const [videoReady, setVideoReady] = useState(false);
  const [timedOut, setTimedOut] = useState(false);

  useEffect(() => {
    let cancelled = false;
    let objectURL = '';
    setError('');
    setPlaybackError('');
    setVideoReady(false);
    setTimedOut(false);

    if (!source || (!mergeMedia && !protectedArtifact)) {
      setResolvedSource(source || '');
      setLoading(false);
      return undefined;
    }

    setResolvedSource('');
    setLoading(true);

    // 合成成片：只换取短期签名地址，交给 <video> 原生 Range 流式播放，
    // 不再整片下载成 Blob（整片下载期间 src 为空就是 0:00 黑屏的根因）。
    const load = mergeMedia
      ? getMergeMediaPlaybackURL(source).then(url => {
        if (!cancelled) setResolvedSource(url);
      })
      : getProductionMediaBlob(source).then(blob => {
        if (cancelled) return;
        if (typeof URL?.createObjectURL !== 'function') throw new Error('当前浏览器无法创建视频预览地址');
        objectURL = URL.createObjectURL(blob);
        setResolvedSource(objectURL);
      });

    load
      .catch(fetchError => {
        if (!cancelled) setError(fetchError?.message || (mergeMedia ? '获取视频播放地址失败' : '视频文件读取失败'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
      if (objectURL && typeof URL?.revokeObjectURL === 'function') URL.revokeObjectURL(objectURL);
    };
  }, [mergeMedia, protectedArtifact, source]);

  useEffect(() => {
    if (!managedSource || !resolvedSource || videoReady || !autoPlay) return undefined;
    const timer = setTimeout(() => setTimedOut(true), MEDIA_FIRST_FRAME_TIMEOUT_MS);
    return () => clearTimeout(timer);
  }, [autoPlay, managedSource, resolvedSource, videoReady]);

  const renderedChildren = useMemo(() => {
    if (!source) return children;
    const replacement = playbackError ? '' : resolvedSource;
    return replaceVideoSource(
      children,
      source,
      replacement,
      () => setPlaybackError('视频源地址已失效或无法访问'),
      () => setVideoReady(true)
    );
  }, [children, playbackError, resolvedSource, source]);

  const showOverlay = managedSource && !videoReady && !error && !playbackError && (loading || autoPlay);
  const overlayText = timedOut ? '视频首帧加载超时，请重试' : (loading ? '正在加载视频…' : '正在加载首帧…');

  return <>
    {renderedChildren}
    {showOverlay ? <span className="batch-factory-media-overlay" data-bf-media-state={timedOut ? 'timeout' : 'loading'}>
      <LoadingOutlined spin />
      <small>{overlayText}</small>
    </span> : null}
    {error ? <small data-bf-media-state="error">视频加载失败：{error}</small> : null}
    {playbackError ? <small data-bf-media-state="error">{playbackError}</small> : null}
    {showDownload && source && resolvedSource ? <a
      data-bf-media-download="video"
      href={resolvedSource}
      download={downloadName(source)}
      target={protectedArtifact ? undefined : '_blank'}
      rel={protectedArtifact ? undefined : 'noreferrer'}
    >下载视频</a> : null}
  </>;
}

export { firstVideoSource, replaceVideoSource };
