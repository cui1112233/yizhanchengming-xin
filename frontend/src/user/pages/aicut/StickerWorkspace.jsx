import { Button, Empty, Input, Upload } from 'antd';
import { Plus, Search, Upload as UploadIcon } from 'lucide-react';
import { useMemo, useState } from 'react';
import MediaArtwork from './MediaArtwork.jsx';
import './workspace-panels.css';

export default function StickerWorkspace({ assets = [], projectId, uploading = false, onUpload, onPreview, onAdd }) {
  const [query, setQuery] = useState('');
  const [previewId, setPreviewId] = useState('');
  const images = useMemo(() => assets.filter(asset => asset.kind === 'image'), [assets]);
  const keyword = query.trim().toLocaleLowerCase();
  const visible = images.filter(asset => !keyword || String(asset.name || '').toLocaleLowerCase().includes(keyword));
  const preview = asset => { setPreviewId(asset.id); onPreview?.(asset); };

  return <section className="aicut-workspace-panel aicut-sticker-workspace" aria-label="贴纸素材">
    <header className="aicut-workspace-heading"><strong>贴纸</strong><span>{images.length} 张图片</span></header>
    <Upload multiple accept="image/*" showUploadList={false} disabled={uploading || !onUpload} beforeUpload={async (file, files) => {
      if (!file.type?.startsWith('image/') && !/\.(png|jpe?g|gif|webp|avif|bmp|svg|heic|heif|tiff?)$/i.test(file.name)) return Upload.LIST_IGNORE;
      await onUpload?.(file, files);
      return false;
    }}><Button block loading={uploading} disabled={!onUpload} icon={<UploadIcon size={14} />}>上传图片贴纸</Button></Upload>
    <Input allowClear prefix={<Search size={14} />} aria-label="搜索贴纸" placeholder="搜索图片名称" value={query} onChange={event => setQuery(event.target.value)} />
    <p className="aicut-workspace-hint">上传自己的图片，预览后添加到独立贴纸轨道。</p>
    <div className="aicut-sticker-grid">
      {visible.map(asset => <article key={asset.id} className={`aicut-sticker-card${previewId === asset.id ? ' is-selected' : ''}`}>
        <button type="button" className="aicut-sticker-preview" aria-label={`预览贴纸 ${asset.name || '未命名图片'}`} onClick={() => preview(asset)}>
          <MediaArtwork asset={asset} projectId={projectId} />
        </button>
        <strong title={asset.name}>{asset.name || '未命名图片'}</strong>
        <Button size="small" block icon={<Plus size={13} />} disabled={!onAdd} onClick={() => onAdd?.(asset)}>添加贴纸</Button>
      </article>)}
    </div>
    {!visible.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={images.length ? '没有找到匹配的图片' : '上传 PNG、JPG 等图片作为贴纸；透明 PNG 可以保留透明背景。'} /> : null}
  </section>;
}
