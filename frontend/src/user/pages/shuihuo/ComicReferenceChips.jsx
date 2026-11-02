import { CloseOutlined, PictureOutlined } from '@ant-design/icons';
import { Image, Popconfirm, Tooltip } from 'antd';
import { useEffect, useState } from 'react';

import { getBookAssetImageBlob, getComicReferencePreviewBlob } from '../../../shared/api/batchFactory';
import { referenceSourceLabel } from './comicReferenceMaterials.js';

function referenceBlob(previewURL) {
  if (String(previewURL || '').includes('/assets/')) return getBookAssetImageBlob(previewURL);
  return getComicReferencePreviewBlob(previewURL);
}

function ComicReferenceThumbnail({ reference }) {
  const [display, setDisplay] = useState({ url: '', failed: false });
  useEffect(() => {
    let active = true;
    let objectURL = '';
    const previewURL = String(reference?.previewUrl || '').trim();
    setDisplay({ url: '', failed: false });
    if (/^https?:\/\//i.test(previewURL) || previewURL.startsWith('data:')) {
      setDisplay({ url: previewURL, failed: false });
      return undefined;
    }
    if (!previewURL) {
      setDisplay({ url: '', failed: true });
      return undefined;
    }
    referenceBlob(previewURL).then(blob => {
      if (!active) return;
      objectURL = URL.createObjectURL(blob);
      setDisplay({ url: objectURL, failed: false });
    }).catch(() => {
      if (active) setDisplay({ url: '', failed: true });
    });
    return () => {
      active = false;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [reference?.previewUrl]);

  if (!display.url) return <span className={`batch-factory-comic-reference-fallback${display.failed ? ' is-failed' : ''}`}><PictureOutlined /></span>;
  return <Image width={40} height={40} src={display.url} alt={reference?.name || '漫画参考素材'} preview={{ mask: '查看' }} />;
}

export function ComicReferenceChips({ referenceImages = [], onRemove, removingURL = '' }) {
  const items = Array.isArray(referenceImages) ? referenceImages : [];
  return <section className="batch-factory-comic-reference-materials" aria-label="漫画参考素材">
    <div className="batch-factory-comic-reference-heading"><b>参考素材</b><span>{items.length}</span></div>
    {items.length ? <div className="batch-factory-comic-reference-strip">
      {items.map((reference, index) => <Tooltip key={`${reference?.url || reference?.id}:${index}`} title={`${reference?.name || `参考素材 ${index + 1}`} · ${referenceSourceLabel(reference?.source)}`}>
        <div className="batch-factory-comic-reference-chip">
          <ComicReferenceThumbnail reference={reference} />
          <span className="batch-factory-comic-reference-copy"><b>{reference?.name || `参考素材 ${index + 1}`}</b><small>{referenceSourceLabel(reference?.source)}</small></span>
          {reference?.removable ? <Popconfirm title="删除这张手动参考图？" description="只会从当前漫画页移除，不影响其他页面。" okText="删除" cancelText="取消" onConfirm={() => onRemove?.(reference)}>
            <button type="button" className="batch-factory-comic-reference-remove" aria-label={`删除${reference?.name || '手动参考图'}`} disabled={removingURL === reference?.url}><CloseOutlined /></button>
          </Popconfirm> : null}
        </div>
      </Tooltip>)}
    </div> : <span className="batch-factory-comic-reference-empty">当前没有接入参考素材</span>}
  </section>;
}
