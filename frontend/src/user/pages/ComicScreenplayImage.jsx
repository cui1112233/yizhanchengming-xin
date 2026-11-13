import { useEffect, useState } from 'react';
import { ImageOff } from 'lucide-react';
import { apiRequest } from '../../shared/api/client.js';

export default function AssetImage({ src, name }) {
  const [url, setUrl] = useState(''), [failed, setFailed] = useState(false);
  useEffect(() => {
    let alive = true, objectURL = '';
    setFailed(false); setUrl('');
    if (!src) return;
    if (/^https?:\/\//.test(src)) { setUrl(src); return; }
    if (!src.startsWith('/api/')) { setFailed(true); return; }
    apiRequest(src, { responseType: 'blob', suppressGlobalError: true }).then(blob => { if (alive) { objectURL = URL.createObjectURL(blob); setUrl(objectURL); } }).catch(() => { if (alive) setFailed(true); });
    return () => { alive = false; if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [src]);
  return url && !failed ? <img src={url} alt={name} onError={() => setFailed(true)} /> : <div className="comic-asset-image-placeholder"><ImageOff size={28} /><span>{failed ? '图片读取失败' : '尚无图片'}</span></div>;
}
