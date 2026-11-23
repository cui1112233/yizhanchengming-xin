import React, { useEffect, useState } from 'react';
import { apiRequest } from '../../../shared/api/client.js';

const bookImage = /^\/api\/batch-factory\/v12\/batches\/([^/?#\\]+)\/books\/([^/?#\\]+)\/text-(?:artifacts|backgrounds)\/([^/?#\\]+)\/content$/;
const personalBackground = /^\/api\/batch-factory\/v12\/text-backgrounds\/([^/?#\\]+)\/content$/;
function isTextMediaURL(url) {
  const match = bookImage.exec(url) || personalBackground.exec(url);
  if (!match) return false;
  try {
    return match.slice(1).every(segment => {
      const id = decodeURIComponent(segment);
      return id !== '.' && id !== '..' && !/[/\\\x00-\x1f]/.test(id);
    });
  } catch {
    return false;
  }
}

function ImageRequest({ contentUrl, alt, className, passive }) {
  const [image, setImage] = useState('');
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    let ownedUrl = '';
    async function load() {
      try {
        const blob = await apiRequest(contentUrl, { responseType: 'blob', signal: controller.signal, suppressGlobalError: true });
        if (controller.signal.aborted) return;
        if (!blob.size || !['image/png','image/jpeg','image/webp'].includes(blob.type)) throw new Error('图片格式错误');
        ownedUrl = URL.createObjectURL(blob);
        setImage(ownedUrl);
      } catch {
        if (!controller.signal.aborted) setFailed(true);
      }
    }
    void load();
    return () => {
      controller.abort();
      if (ownedUrl) URL.revokeObjectURL(ownedUrl);
    };
  }, [contentUrl, attempt]);
  if (failed && passive) return <span>图片读取失败</span>;
  if (failed) return <button type="button" onClick={() => {
    setImage('');
    setFailed(false);
    setAttempt(value => value + 1);
  }}>图片读取失败，重试</button>;
  return image
    ? <img className={className} src={image} alt={alt} onError={() => setFailed(true)} />
    : <span role="status">正在读取图片…</span>;
}

// Only send the session token to the authenticated text-media content routes.
export default function AuthenticatedTextPageImage({ contentUrl, alt = '文字页', className, passive = false }) {
  if (!contentUrl) return <span>暂无图片</span>;
  if (!isTextMediaURL(contentUrl)) return <span>图片地址无效</span>;
  return <ImageRequest key={contentUrl} contentUrl={contentUrl} alt={alt} className={className} passive={passive} />;
}
