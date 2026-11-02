import { Children, cloneElement, isValidElement, useEffect, useMemo, useState } from 'react';
import { getProductionMediaBlob, isProtectedProductionMediaURL } from '../../../shared/api/batchFactoryV11.js';

function firstVideoSource(node) {
  if (Array.isArray(node)) {
    for (const child of node) {
      const value = firstVideoSource(child);
      if (value) return value;
    }
    return '';
  }
  if (!isValidElement(node)) return '';
  if (node.type === 'video' && typeof node.props?.src === 'string') return node.props.src;
  return firstVideoSource(node.props?.children);
}

function replaceVideoSource(node, source, replacement, onPlaybackError) {
  if (Array.isArray(node)) return node.map(child => replaceVideoSource(child, source, replacement, onPlaybackError));
  if (!isValidElement(node)) return node;
  if (node.type === 'video' && node.props?.src === source) {
    const originalOnError = node.props.onError;
    return cloneElement(node, {
      src: replacement || undefined,
      onError: event => {
        originalOnError?.(event);
        onPlaybackError?.();
      }
    });
  }
  if (node.props?.children === undefined) return node;
  return cloneElement(node, undefined, Children.map(
    node.props.children,
    child => replaceVideoSource(child, source, replacement, onPlaybackError)
  ));
}

function downloadName(source) {
  const tail = String(source || '').split('?')[0].split('/').filter(Boolean).pop() || 'video';
  return /\.mp4$/i.test(tail) ? tail : `${tail}.mp4`;
}

export function ProductionMediaBoundary({ children, showDownload = true }) {
  const source = useMemo(() => firstVideoSource(children), [children]);
  const protectedArtifact = isProtectedProductionMediaURL(source);
  const [resolvedSource, setResolvedSource] = useState(protectedArtifact ? '' : source);
  const [error, setError] = useState('');
  const [playbackError, setPlaybackError] = useState('');
  const [loading, setLoading] = useState(protectedArtifact && Boolean(source));

  useEffect(() => {
    let cancelled = false;
    let objectURL = '';
    setError('');
    setPlaybackError('');

    if (!source || !isProtectedProductionMediaURL(source)) {
      setResolvedSource(source || '');
      setLoading(false);
      return undefined;
    }

    setResolvedSource('');
    setLoading(true);
    getProductionMediaBlob(source)
      .then(blob => {
        if (cancelled) return;
        if (typeof URL?.createObjectURL !== 'function') throw new Error('当前浏览器无法创建视频预览地址');
        objectURL = URL.createObjectURL(blob);
        setResolvedSource(objectURL);
      })
      .catch(fetchError => {
        if (!cancelled) setError(fetchError?.message || '视频文件读取失败');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
      if (objectURL && typeof URL?.revokeObjectURL === 'function') URL.revokeObjectURL(objectURL);
    };
  }, [source]);

  const renderedChildren = useMemo(() => {
    if (!source) return children;
    const replacement = playbackError ? '' : (protectedArtifact ? resolvedSource : source);
    return replaceVideoSource(children, source, replacement, () => setPlaybackError('视频源地址已失效或无法访问'));
  }, [children, playbackError, protectedArtifact, resolvedSource, source]);

  return <>
    {renderedChildren}
    {loading ? <small data-bf-media-state="loading">正在读取已认证的视频文件…</small> : null}
    {error ? <small data-bf-media-state="error">视频文件读取失败：{error}</small> : null}
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
